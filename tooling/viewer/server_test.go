package viewer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testToken = "test-token-0123456789abcdef"
	testPort  = 41234
	testHost  = "127.0.0.1:41234"
)

type srvEnv struct {
	*env
	repo   string
	folder string
	srv    *Server
	key    string
}

// newSrvEnv ingests valid/final from a fake repository and builds a server.
func newSrvEnv(t *testing.T) *srvEnv {
	t.Helper()
	e := newEnv(t)
	repo := newRepo(t)
	e.hive[repo] = "Sample Project"
	e.remotes[repo] = "git@github.com:Acme/Sample.git"
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	srv, err := NewServer(ServerOptions{Store: e.store, Token: testToken, Port: testPort, Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	return &srvEnv{env: e, repo: repo, folder: folder, srv: srv, key: e.onlySession().Key}
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
			{http.MethodDelete, "/api/sessions/" + s.key},
			{http.MethodPost, "/api/sessions/" + s.key + "/pin"},
			{http.MethodDelete, "/api/sessions/" + s.key + "/pin"},
			{http.MethodPost, "/api/sessions/" + s.key + "/archive"},
			{http.MethodDelete, "/api/sessions/" + s.key + "/archive"},
			{http.MethodPut, "/api/sessions/" + s.key + "/title"},
		} {
			if w := s.do(req[0], req[1], o); w.Code != http.StatusForbidden {
				t.Errorf("%s: %s %s = %d, want 403", name, req[0], req[1], w.Code)
			}
		}
	}
	if len(s.sessions()) != 1 || s.sessions()[0].Pinned || s.sessions()[0].Archived {
		t.Fatal("a rejected write changed the store")
	}
}

// S2-A: reading the index or a manifest needs the token; reading a file
// needs the per-session capability derived from it.
func TestServerReadsNeedTheTokenOrTheCapability(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{"/api/index", "/api/sessions/" + s.key} {
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
	img := "images/nav-1.png"
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
	// The index hands the authenticated page its file base.
	var v struct {
		Projects []struct {
			Tickets []struct {
				Sessions []struct {
					Key      string `json:"key"`
					FileBase string `json:"fileBase"`
				} `json:"sessions"`
			} `json:"tickets"`
		} `json:"projects"`
	}
	w := s.do(http.MethodGet, "/api/index", reqOpts{token: testToken})
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if got := v.Projects[0].Tickets[0].Sessions[0].FileBase; got != s.files() {
		t.Fatalf("fileBase = %q, want %q", got, s.files())
	}
}

func capFor(key string) string {
	mac := hmac.New(sha256.New, []byte(testToken))
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestServerChecksTheHostOnEveryRoute(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{"/", "/assets/walkthrough.mjs", "/api/index", "/api/sessions/" + s.key, s.files() + "images/nav-1.png", "/api/health"} {
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
	if h.PID != os.Getpid() || h.StateDir != s.state || h.Version != "test" || h.TTL != (14*24*time.Hour).String() {
		t.Fatalf("health = %+v", h)
	}
	if strings.Contains(w.Body.String(), testToken) {
		t.Fatal("health must not echo the token")
	}
}

func TestServerDeleteRemovesOnlyTheCopyAndItStaysGone(t *testing.T) {
	s := newSrvEnv(t)
	source := filepath.Join(s.folder, "walkthrough.json")
	image := filepath.Join(s.folder, "images", "nav-1.png")
	before := [2]string{sha(string(readFile(t, source))), sha(string(readFile(t, image)))}

	if w := s.write(http.MethodDelete, "/api/sessions/"+s.key); w.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d: %s", w.Code, w.Body)
	}
	if _, err := os.Stat(s.sessionDir(s.key)); !os.IsNotExist(err) {
		t.Fatalf("the copy is still on disk: %v", err)
	}
	if w := s.do(http.MethodGet, "/api/sessions/"+s.key, reqOpts{token: testToken}); w.Code != http.StatusNotFound {
		t.Fatalf("GET after delete = %d", w.Code)
	}
	s.advance(time.Minute)
	s.sync() // the poll: the tombstone keeps the same content out
	s.sync()
	if _, err := os.Stat(s.sessionDir(s.key)); !os.IsNotExist(err) {
		t.Fatal("the deleted copy came back after a poll")
	}
	after := [2]string{sha(string(readFile(t, source))), sha(string(readFile(t, image)))}
	if before != after {
		t.Fatal("the source files in the repository changed")
	}
	if w := s.write(http.MethodDelete, "/api/sessions/"+s.key); w.Code != http.StatusNotFound {
		t.Fatalf("second DELETE = %d, want 404", w.Code)
	}
}

func TestServerPinAndUnpin(t *testing.T) {
	s := newSrvEnv(t)
	if w := s.write(http.MethodPost, "/api/sessions/"+s.key+"/pin"); w.Code != http.StatusNoContent {
		t.Fatalf("POST pin = %d", w.Code)
	}
	if !s.onlySession().Pinned {
		t.Fatal("not pinned")
	}
	if w := s.write(http.MethodDelete, "/api/sessions/"+s.key+"/pin"); w.Code != http.StatusNoContent {
		t.Fatalf("DELETE pin = %d", w.Code)
	}
	if s.onlySession().Pinned {
		t.Fatal("still pinned")
	}
}

func TestServerKeyAndMethodErrors(t *testing.T) {
	s := newSrvEnv(t)
	missing := strings.Repeat("a", 64)
	for _, tc := range []struct {
		method, target string
		want           int
	}{
		{http.MethodGet, "/api/sessions/not-a-key", http.StatusBadRequest},
		{http.MethodDelete, "/api/sessions/..", http.StatusBadRequest},
		{http.MethodGet, "/api/sessions/" + missing, http.StatusNotFound},
		{http.MethodPost, "/api/sessions/" + missing + "/pin", http.StatusNotFound},
		{http.MethodPost, "/api/sessions/" + missing + "/archive", http.StatusNotFound},
		{http.MethodDelete, "/api/sessions/" + missing + "/archive", http.StatusNotFound},
		{http.MethodPost, "/api/sessions/not-a-key/archive", http.StatusBadRequest},
		{http.MethodDelete, "/api/sessions/not-a-key/archive", http.StatusBadRequest},
		{http.MethodGet, "/api/sessions/" + s.key + "/archive", http.StatusMethodNotAllowed},
		{http.MethodPost, "/api/sessions/" + s.key, http.StatusMethodNotAllowed},
		{http.MethodPut, "/api/index", http.StatusMethodNotAllowed},
		{http.MethodGet, "/nope", http.StatusNotFound},
		{http.MethodGet, "/assets/meta.json", http.StatusNotFound},
		{http.MethodGet, "/assets/../server.go", http.StatusNotFound},
	} {
		w := s.write(tc.method, tc.target)
		if w.Code != tc.want {
			t.Errorf("%s %s = %d, want %d", tc.method, tc.target, w.Code, tc.want)
		}
	}
}

func TestServerFilesRejectTraversalAndServeMediaByExtension(t *testing.T) {
	s := newSrvEnv(t)
	for _, target := range []string{
		s.files() + "../x",
		s.files() + "images/../../meta.json",
		s.files() + "%2e%2e/meta.json",
		s.files() + "meta.json",
		s.files() + "walkthrough.json",
		s.files() + "images/nav-1.png/",
	} {
		w := s.do(http.MethodGet, target, reqOpts{token: testToken})
		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 400 or 404", target, w.Code)
		}
		if bytes.Contains(w.Body.Bytes(), []byte("ingested_at")) {
			t.Errorf("GET %s leaked meta.json", target)
		}
	}
	w := s.do(http.MethodGet, s.files()+"images/nav-1.png", reqOpts{})
	if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("png: %d %q", w.Code, w.Header().Get("Content-Type"))
	}
	if !bytes.Equal(w.Body.Bytes(), readFile(t, filepath.Join(s.folder, "images", "nav-1.png"))) {
		t.Fatal("served bytes differ from the copy")
	}
	if w := s.do(http.MethodGet, s.files()+"images/missing.png", reqOpts{}); w.Code != http.StatusNotFound {
		t.Fatalf("missing file = %d", w.Code)
	}

	// Content-Type per extension, on files placed in the copy (the copy's own
	// validation is the Store's job; here only the header mapping is checked).
	dir := s.sessionDir(s.key)
	for name, want := range map[string]string{
		"a.jpg": "image/jpeg", "b.jpeg": "image/jpeg", "c.webp": "image/webp", "d.mp4": "video/mp4",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		w := s.do(http.MethodGet, s.files()+name, reqOpts{})
		if got := w.Header().Get("Content-Type"); w.Code != http.StatusOK || got != want {
			t.Errorf("%s: %d, Content-Type = %q, want %q", name, w.Code, got, want)
		}
	}
	// A video range request works (seek in the browser).
	r := httptest.NewRequest(http.MethodGet, s.files()+"d.mp4", nil)
	r.Host = testHost
	r.Header.Set("Range", "bytes=0-0")
	rec := httptest.NewRecorder()
	s.srv.ServeHTTP(rec, r)
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("range request = %d", rec.Code)
	}
}

func TestServerSendsSecurityHeadersAndTheRightContentTypes(t *testing.T) {
	s := newSrvEnv(t)
	wantCSP := "default-src 'self'; img-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'"
	for _, target := range []string{"/", "/assets/app/index.html", "/api/index", "/api/sessions/" + s.key, s.files() + "images/nav-1.png", "/nope"} {
		w := s.do(http.MethodGet, target, reqOpts{token: testToken})
		if got := w.Header().Get("Content-Security-Policy"); got != wantCSP {
			t.Errorf("%s: CSP = %q", target, got)
		}
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("%s: nosniff = %q", target, got)
		}
	}
	if w := s.write(http.MethodPost, "/api/sessions/"+s.key); w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("error responses need nosniff too")
	}
	for target, prefix := range map[string]string{
		"/":                       "text/html",
		"/assets/walkthrough.mjs": "text/javascript",
		"/assets/walkthrough.css": "text/css",
		"/api/index":              "application/json",
		"/api/sessions/" + s.key:  "application/json",
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
	for _, name := range []string{"walkthrough.mjs", "walkthrough.css"} {
		if w := s.do(http.MethodGet, "/assets/"+name, reqOpts{}); w.Code != http.StatusOK || w.Body.Len() == 0 {
			t.Errorf("/assets/%s = %d (%d bytes)", name, w.Code, w.Body.Len())
		}
	}
	for _, name := range []string{"viewer.mjs", "viewer.css", "index.html"} {
		if w := s.do(http.MethodGet, "/assets/"+name, reqOpts{}); w.Code != http.StatusNotFound {
			t.Errorf("retired /assets/%s = %d", name, w.Code)
		}
	}
}

func TestServerIndexGroupsByProjectTicketAndSession(t *testing.T) {
	s := newSrvEnv(t)
	// Second session of the same ticket, a session of another ticket, one
	// without a ticket, and a repository that has its own project.
	s.place(s.repo, "run2", nil)
	s.place(s.repo, "other", func(m map[string]any) { m["ticket"] = "gh-99-other" })
	s.place(s.repo, "free", func(m map[string]any) { delete(m, "ticket") })
	repo2 := newRepo(t)
	s.hive[repo2] = "Second Project"
	s.remotes[repo2] = "https://github.com/Acme/Second.git"
	s.place(repo2, "run1", nil)
	s.register(repo2)
	s.sync()

	w := s.do(http.MethodGet, "/api/index", reqOpts{token: testToken})
	if w.Code != http.StatusOK {
		t.Fatalf("index = %d", w.Code)
	}
	var v struct {
		Projects []struct {
			Key     string `json:"key"`
			Name    string `json:"name"`
			Tickets []struct {
				Ticket   string `json:"ticket"`
				Sessions []struct {
					Key       string `json:"key"`
					SessionID string `json:"session_id"`
					Pinned    bool   `json:"pinned"`
				} `json:"sessions"`
			} `json:"tickets"`
		} `json:"projects"`
		Errors []any `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if len(v.Projects) != 2 {
		t.Fatalf("projects = %d, want 2: %s", len(v.Projects), w.Body)
	}
	p := v.Projects[0]
	if p.Name != "Sample Project" || p.Key != "github.com/Acme/Sample" {
		t.Fatalf("first project = %+v", p)
	}
	byTicket := map[string][]string{}
	for _, tk := range p.Tickets {
		for _, se := range tk.Sessions {
			byTicket[tk.Ticket] = append(byTicket[tk.Ticket], se.SessionID)
		}
	}
	if len(byTicket["gh-67-flow-walkthrough-skill"]) != 2 || len(byTicket["gh-99-other"]) != 1 || len(byTicket[""]) != 1 || len(p.Tickets) != 3 {
		t.Fatalf("tickets = %v", byTicket)
	}
	if last := p.Tickets[len(p.Tickets)-1]; last.Ticket != "" {
		t.Fatalf("the no-ticket group must be last, got %q", last.Ticket)
	}
	if v.Projects[1].Name != "Second Project" || len(v.Projects[1].Tickets) != 1 || len(v.Projects[1].Tickets[0].Sessions) != 1 {
		t.Fatalf("second project = %+v", v.Projects[1])
	}
}

func TestServerServesTheStoredManifest(t *testing.T) {
	s := newSrvEnv(t)
	w := s.do(http.MethodGet, "/api/sessions/"+s.key, reqOpts{token: testToken})
	if w.Code != http.StatusOK || !bytes.Equal(w.Body.Bytes(), readFile(t, filepath.Join(s.sessionDir(s.key), "walkthrough.json"))) {
		t.Fatalf("manifest: %d", w.Code)
	}
}

func TestServerConstructorValidatesItsOptions(t *testing.T) {
	e := newEnv(t)
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
