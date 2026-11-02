package viewer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testToken = "test-token-0123456789abcdef"
	testPort  = 41234
	testHost  = "127.0.0.1:41234"
)

// env is the state directory and store a test runs against.
type env struct {
	t     *testing.T
	state string
	store *Store
}

type srvEnv struct {
	*env
	repo   string
	folder string
	srv    *Server
	image  string
	key    string
}

// newSrvEnv uses only v3 API state; shared auth/static tests do not scan manifests.
func newSrvEnv(t *testing.T) *srvEnv {
	t.Helper()
	store, v, _ := newV3Fixture(t)
	// Seed an already registered candidate for auth/static tests. Catalog writer
	// transactions are tested independently; this helper makes no writer claim.
	store.timeline[v.ViewerSessionID].Session.Candidates = append(store.timeline[v.ViewerSessionID].Session.Candidates, TimelineCandidateRecord{CandidateID: "candidate-test", TimelineCandidate: TimelineCandidate{Branch: "test", Commit: "abc", Base: "development"}})
	e := &env{t: t, state: filepath.Dir(store.Dir()), store: store}
	srv, err := NewServer(ServerOptions{Store: store, Token: testToken, Port: testPort, Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	var imageBytes bytes.Buffer
	if err = png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	artifact, err := store.TimelineAddArtifact(v.ViewerSessionID, TimelineArtifactMetadata{SchemaVersion: 3, OperationID: "security-image", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, CapturedAt: v.CreatedAt, CandidateID: "candidate-test", Viewport: "desktop", Theme: "dark", MediaType: "image/png"}, imageBytes.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return &srvEnv{env: e, repo: v.Source.ProjectRoot, srv: srv, key: v.ViewerSessionID, image: "artifacts/" + artifact.ArtifactID + ".png"}
}

type reqOpts struct {
	scope  string
	host   string
	token  string
	origin string
	body   io.Reader
}

func (s *srvEnv) do(method, target string, o reqOpts) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, o.body)
	r.Host = testHost
	if o.host != "" {
		r.Host = o.host
	}
	if o.token != "" {
		r.Header.Set("X-Hive-Token", o.token)
	}
	if o.origin != "" {
		r.Header.Set("Origin", o.origin)
	}
	if o.scope != "" {
		r.Header.Set("X-Hive-Action-Scope", o.scope)
	}
	w := httptest.NewRecorder()
	s.srv.ServeHTTP(w, r)
	return w
}

// cap is the per-session capability the server derives from the token:
// hex HMAC-SHA256(token, key).
func (s *srvEnv) cap() string {
	mac := hmac.New(sha256.New, []byte(testToken))
	mac.Write([]byte(s.key))
	return hex.EncodeToString(mac.Sum(nil))
}

// files is the base of the session's file URLs.
func (s *srvEnv) files() string { return "/files/" + s.key + "/" + s.cap() + "/" }

func (s *srvEnv) write(method, target string) *httptest.ResponseRecorder {
	return s.do(method, target, reqOpts{token: testToken, origin: "http://" + testHost, scope: "session"})
}

func TestServerRejectsWritesWithoutTokenOrWithAForeignTokenOriginOrHost(t *testing.T) {
	s := newSrvEnv(t)
	cases := map[string]reqOpts{
		"no token":       {origin: "http://" + testHost},
		"foreign token":  {token: "other-token", origin: "http://" + testHost},
		"foreign origin": {token: testToken, origin: "http://evil.example"},
		"no origin":      {token: testToken},
		"foreign host":   {token: testToken, origin: "http://evil.example", host: "evil.example"},
		"rebinding host": {token: testToken, origin: "http://" + testHost, host: "evil.example:41234"},
		"wrong port":     {token: testToken, origin: "http://" + testHost, host: "127.0.0.1:1"},
	}
	for name, o := range cases {
		for _, req := range [][2]string{
			{http.MethodDelete, "/api/timeline/sessions/" + s.key},
			{http.MethodPatch, "/api/timeline/sessions/" + s.key},
			{http.MethodPost, "/api/timeline/sessions/" + s.key + "/events"},
			{http.MethodPost, "/api/timeline/sessions/" + s.key + "/artifacts"},
			{http.MethodPost, "/api/timeline/sessions/" + s.key + "/documents"},
		} {
			if w := s.do(req[0], req[1], o); w.Code != http.StatusForbidden {
				t.Errorf("%s: %s %s = %d, want 403", name, req[0], req[1], w.Code)
			}
		}
	}
	current, err := s.store.TimelineGet(s.key)
	if err != nil || current.Version != 2 {
		t.Fatal("rejected write changed store", err)
	}
}

// S2-A: reading the index or a manifest needs the token; reading a file
// needs the per-session capability derived from it.
func TestServerReadsNeedTheTokenOrTheCapability(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{"/api/index", "/api/timeline/sessions/" + s.key} {
		for name, o := range map[string]reqOpts{
			"no token":       {},
			"foreign token":  {token: "other-token-0123456789"},
			"foreign origin": {token: testToken, origin: "http://evil.example"},
		} {
			if w := s.do(http.MethodGet, target, o); w.Code != http.StatusForbidden {
				t.Errorf("GET %s (%s) = %d, want 403", target, name, w.Code)
			} else if strings.Contains(w.Body.String(), "sess-run1") {
				t.Errorf("GET %s (%s) leaked data", target, name)
			}
		}
		if w := s.do(http.MethodGet, target, reqOpts{token: testToken}); w.Code != http.StatusOK {
			t.Errorf("GET %s with the token = %d", target, w.Code)
		}
	}
	img := s.image
	other := strings.Repeat("a", 64)
	for name, target := range map[string]string{
		"no capability":       "/files/" + s.key + "/" + img,
		"no path":             "/files/" + s.key + "/" + s.cap(),
		"wrong capability":    "/files/" + s.key + "/" + strings.Repeat("0", 64) + "/" + img,
		"capability of other": "/files/" + s.key + "/" + capFor(other) + "/" + img,
		"upper-case cap":      "/files/" + s.key + "/" + strings.ToUpper(s.cap()) + "/" + img,
		"short key":           "/files/short/" + s.cap() + "/" + img,
	} {
		if w := s.do(http.MethodGet, target, reqOpts{token: testToken}); w.Code != http.StatusForbidden && w.Code != http.StatusNotFound {
			t.Errorf("%s: GET %s = %d, want 403 or 404", name, target, w.Code)
		} else if w.Code == http.StatusOK || bytes.HasPrefix(w.Body.Bytes(), []byte("\x89PNG")) {
			t.Errorf("%s: served a file", name)
		}
	}
	if w := s.do(http.MethodGet, "/files/"+s.key+"/"+strings.Repeat("0", 64)+"/"+img, reqOpts{token: testToken}); w.Code != http.StatusForbidden {
		t.Errorf("wrong capability = %d, want 403 even with the token", w.Code)
	}
	if w := s.do(http.MethodGet, s.files()+img, reqOpts{}); w.Code != http.StatusOK {
		t.Errorf("file with the capability and no token = %d, want 200 (an <img> cannot send headers)", w.Code)
	}
	// The authenticated work response, rather than the summary, carries file capabilities.
	var response TimelineSessionResponse
	w := s.do(http.MethodGet, "/api/timeline/sessions/"+s.key, reqOpts{token: testToken})
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	for _, url := range response.ArtifactURLs {
		if url != s.files()+s.image {
			t.Fatal("wrong artifact capability")
		}
	}
	for _, suffix := range []string{"../storage.json", "artifacts/../storage.json", "documents/file.md", "artifacts/unknown.png"} {
		if w := s.do(http.MethodGet, s.files()+suffix, reqOpts{}); w.Code == 200 {
			t.Fatal("served unrecorded path", suffix)
		}
	}
}

func capFor(key string) string {
	mac := hmac.New(sha256.New, []byte(testToken))
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestServerChecksTheHostOnEveryRoute(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{"/", "/assets/hive-mark.svg", "/api/index", "/api/timeline/sessions/" + s.key, s.files() + s.image, "/api/health"} {
		if w := s.do(http.MethodGet, target, reqOpts{host: "evil.example:41234", token: testToken}); w.Code != http.StatusForbidden {
			t.Errorf("GET %s with a foreign Host = %d, want 403", target, w.Code)
		}
		if w := s.do(http.MethodGet, target, reqOpts{host: "localhost:41234", token: testToken}); w.Code != http.StatusOK {
			t.Errorf("GET %s with Host localhost = %d, want 200", target, w.Code)
		}
	}
}

func TestServerHealthNeedsTheTokenAndReportsPidStateDirVersionAndTTL(t *testing.T) {
	s := newSrvEnv(t)
	if w := s.do(http.MethodGet, "/api/health", reqOpts{}); w.Code != http.StatusForbidden {
		t.Fatalf("health without token = %d", w.Code)
	}
	if w := s.do(http.MethodGet, "/api/health", reqOpts{token: "nope"}); w.Code != http.StatusForbidden {
		t.Fatalf("health with a foreign token = %d", w.Code)
	}
	if w := s.do(http.MethodGet, "/api/health", reqOpts{token: testToken, origin: "http://evil.example"}); w.Code != http.StatusForbidden {
		t.Fatalf("health with a foreign Origin = %d", w.Code)
	}
	w := s.do(http.MethodGet, "/api/health", reqOpts{token: testToken})
	if w.Code != http.StatusOK {
		t.Fatalf("health = %d", w.Code)
	}
	var h struct {
		PID      int    `json:"pid"`
		StateDir string `json:"state_dir"`
		Version  string `json:"version"`
		TTL      string `json:"ttl"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	if h.PID != os.Getpid() || h.StateDir != s.state || h.Version != "test" || h.TTL != "0s" {
		t.Fatalf("health = %+v", h)
	}
	if strings.Contains(w.Body.String(), testToken) {
		t.Fatal("health must not echo the token")
	}
}

func TestServerSendsSecurityHeadersAndTheRightContentTypes(t *testing.T) {
	s := newSrvEnv(t)
	wantCSP := "default-src 'self'; img-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'"
	for _, target := range []string{"/", "/assets/app/index.html", "/api/index", "/api/timeline/sessions/" + s.key, s.files() + s.image, "/nope"} {
		w := s.do(http.MethodGet, target, reqOpts{token: testToken})
		if got := w.Header().Get("Content-Security-Policy"); got != wantCSP {
			t.Errorf("%s: CSP = %q", target, got)
		}
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: nosniff = %q", target, got)
		}
	}
	if w := s.write(http.MethodPost, "/api/timeline/sessions/"+s.key); w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("error responses need nosniff too")
	}
	for target, prefix := range map[string]string{
		"/":                               "text/html",
		"/assets/hive-mark.svg":           "image/svg+xml",
		"/api/index":                      "application/json",
		"/api/timeline/sessions/" + s.key: "application/json",
	} {
		w := s.do(http.MethodGet, target, reqOpts{token: testToken})
		if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), prefix) {
			t.Errorf("%s: %d %q, want %s", target, w.Code, w.Header().Get("Content-Type"), prefix)
		}
	}
}

func TestServerStaticPagesHaveNoInlineScriptOrStyle(t *testing.T) {
	s := newSrvEnv(t)
	page := s.do(http.MethodGet, "/", reqOpts{}).Body.String()
	lower := strings.ToLower(page)
	if strings.Contains(lower, " style=") || strings.Contains(lower, "<style") {
		t.Error("index.html has inline style")
	}
	for _, handler := range []string{" onclick=", " onload=", " onerror="} {
		if strings.Contains(lower, handler) {
			t.Errorf("index.html has an inline handler %q", handler)
		}
	}
	for _, chunk := range strings.Split(lower, "<script")[1:] {
		tag := chunk[:strings.Index(chunk, ">")]
		if !strings.Contains(tag, "src=") {
			t.Errorf("inline <script%s>", tag)
		}
	}
	if w := s.do(http.MethodGet, "/assets/hive-mark.svg", reqOpts{}); w.Code != http.StatusOK || w.Body.Len() == 0 {
		t.Errorf("/assets/hive-mark.svg = %d (%d bytes)", w.Code, w.Body.Len())
	}
	for _, name := range []string{"viewer.mjs", "viewer.css", "index.html"} {
		if w := s.do(http.MethodGet, "/assets/"+name, reqOpts{}); w.Code != http.StatusNotFound {
			t.Errorf("retired /assets/%s = %d", name, w.Code)
		}
	}
}

// The retired /api/sessions/* routes are not special-cased any more: they fall
// to the default route and answer 404, with or without the token.
func TestServerAnswers404ForTheRetiredSessionsAPI(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{"/api/sessions/x", "/api/sessions/" + s.key, "/api/sessions/"} {
		for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
			if w := s.do(method, target, reqOpts{token: testToken, origin: "http://" + testHost, scope: "session"}); w.Code != http.StatusNotFound {
				t.Errorf("%s %s = %d, want 404", method, target, w.Code)
			}
		}
	}
}

func TestServerConstructorValidatesItsOptions(t *testing.T) {
	store, _, _ := newV3Fixture(t)
	e := &env{t: t, state: filepath.Dir(store.Dir()), store: store}
	for name, o := range map[string]ServerOptions{
		"no store":    {Token: testToken, Port: 1},
		"no token":    {Store: e.store, Port: 1},
		"short token": {Store: e.store, Token: "abc", Port: 1},
		"no port":     {Store: e.store, Token: testToken},
	} {
		if _, err := NewServer(o); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
