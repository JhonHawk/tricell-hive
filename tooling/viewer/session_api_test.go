package viewer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionGroupLegacyClient(t *testing.T) {
	s := newSrvEnv(t)
	for _, req := range [][2]string{{"DELETE", ""}, {"POST", "/pin"}, {"DELETE", "/pin"}, {"POST", "/archive"}, {"DELETE", "/archive"}, {"PUT", "/title"}} {
		w := s.do(req[0], "/api/sessions/"+s.key+req[1], reqOpts{token: testToken, origin: "http://" + testHost, body: strings.NewReader(`{"title":"Old client"}`)})
		if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), `"code":"reload_required"`) {
			t.Fatalf("legacy write = %d %s", w.Code, w.Body.String())
		}
	}
	if len(s.sessions()) != 1 || s.onlySession().Pinned || s.onlySession().Archived {
		t.Fatal("legacy write changed data")
	}
}

func TestSessionGroupTitle(t *testing.T) {
	s := newSrvEnv(t)
	target := "/api/sessions/" + s.key + "/title"
	for _, body := range []string{`{"title":"  Saved 界  "}`, `{"title":""}`} {
		w := s.do("PUT", target, reqOpts{token: testToken, origin: "http://" + testHost, scope: "session", body: strings.NewReader(body)})
		if w.Code != 204 {
			t.Fatalf("title = %d %s", w.Code, w.Body.String())
		}
		var sent struct {
			Title string `json:"title"`
		}
		_ = json.Unmarshal([]byte(body), &sent)
		s.store = s.open()
		s.start()
		s.srv.store = s.store
		if s.onlySession().SessionTitle != strings.TrimSpace(sent.Title) {
			t.Fatal("HTTP title lost on restart")
		}
	}
	for _, body := range []string{`{}`, `null`, `{"title":null}`, `{"title":2}`, `{"title":"a","extra":1}`, `{"title":"a"}{}`, `{"title":"line\nbreak"}`, `{"title":"` + strings.Repeat("界", 201) + `"}`, `{"title":"` + strings.Repeat(" ", 2050) + `"}`} {
		w := s.do("PUT", target, reqOpts{token: testToken, origin: "http://" + testHost, scope: "session", body: strings.NewReader(body)})
		if w.Code != 400 {
			t.Fatalf("invalid body accepted: %d", w.Code)
		}
	}
	if s.onlySession().SessionTitle != "" {
		t.Fatal("invalid title changed preferences")
	}
}

func TestSessionGroupAPIAllMembersAndExcludedReads(t *testing.T) {
	e, repo := groupFixture(t)
	members := e.sessions()
	excluded := members[0]
	visible := members[1]
	e.place(repo, "other", nil)
	e.sync()
	e.store.tracking = func(_ context.Context, paths []string) map[string]SourceClassification {
		out := map[string]SourceClassification{}
		for _, path := range paths {
			status := "local"
			if path == filepath.Join(excluded.Source, manifestFile) {
				status = "tracked"
			}
			out[path] = SourceClassification{Status: status}
		}
		return out
	}
	e.sync()
	srv, err := NewServer(ServerOptions{Store: e.store, Token: testToken, Port: testPort})
	if err != nil {
		t.Fatal(err)
	}
	s := &srvEnv{env: e, srv: srv, key: visible.Key}
	for _, suffix := range []string{"/archive", "/pin"} {
		for _, method := range []string{"POST", "POST", "DELETE", "DELETE", "POST"} {
			w := s.write(method, "/api/sessions/"+visible.Key+suffix)
			if w.Code != 204 {
				t.Fatalf("group mutation = %d", w.Code)
			}
		}
	}
	for _, sess := range e.store.current().Sessions {
		want := sess.SessionID == visible.SessionID
		if sess.Pinned != want || sess.Archived != want {
			t.Fatalf("group scope failed: %+v", sess)
		}
	}
	if w := s.do("GET", "/api/sessions/"+excluded.Key, reqOpts{token: testToken}); w.Code != 409 {
		t.Fatal("group mutation exposed excluded document")
	}
	if w := s.write("DELETE", "/api/sessions/"+visible.Key); w.Code != 204 {
		t.Fatalf("delete = %d", w.Code)
	}
	if len(e.store.current().Sessions) != 1 || e.store.current().Sessions[0].SessionID == visible.SessionID {
		t.Fatal("group deletion left excluded sibling or removed other identity")
	}
}

func TestSessionGroupAPIPartialDeleteResponse(t *testing.T) {
	e, _ := groupFixture(t)
	members := e.sessions()
	srv, err := NewServer(ServerOptions{Store: e.store, Token: testToken, Port: testPort})
	if err != nil {
		t.Fatal(err)
	}
	s := &srvEnv{env: e, srv: srv, key: members[0].Key}
	realRename := e.store.renamePath
	e.store.renamePath = func(from, to string) error {
		if from == e.sessionDir(s.key) {
			return errors.New("private source diagnostic")
		}
		return realRename(from, to)
	}
	w := s.write("DELETE", "/api/sessions/"+s.key)
	var body struct {
		Code      string `json:"code"`
		Deleted   int    `json:"deleted"`
		Remaining int    `json:"remaining"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != 500 || body.Code != "partial_delete" || body.Deleted != 1 || body.Remaining != 1 {
		t.Fatalf("bad error response: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "private") {
		t.Fatal("internal diagnostic exposed")
	}
}

func TestSessionGroupAPICleanupReadDoesNotRepair(t *testing.T) {
	s := newSrvEnv(t)
	calls := 0
	s.store.removeFile = func(path string) error { calls++; return errors.New("unavailable") }
	w := s.write("DELETE", "/api/sessions/"+s.key)
	if w.Code != 500 || !strings.Contains(w.Body.String(), `"cleanup_pending":true`) {
		t.Fatalf("cleanup response: %d %s", w.Code, w.Body.String())
	}
	before := calls
	w = s.do("GET", "/api/index", reqOpts{token: testToken})
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"code":"delete_cleanup_pending"`) {
		t.Fatal("cleanup state invisible in index")
	}
	if calls != before {
		t.Fatal("GET attempted a repair")
	}
}

func TestSessionGroupAPITitleStatusesAndBodyLimit(t *testing.T) {
	s := newSrvEnv(t)
	target := "/api/sessions/" + s.key + "/title"
	valid := `{"title":""}`
	exact := valid + strings.Repeat(" ", 2048-len(valid))
	for _, tc := range []struct {
		method, target, body, scope string
		status                      int
	}{
		{"PUT", target, exact, "session", 204},
		{"PUT", target, exact + " ", "session", 400},
		{"POST", target, valid, "session", 405},
		{"PUT", "/api/sessions/invalid/title", valid, "session", 400},
		{"PUT", "/api/sessions/" + strings.Repeat("f", 64) + "/title", valid, "session", 404},
		{"PUT", target, valid, "copy", 409},
	} {
		w := s.do(tc.method, tc.target, reqOpts{token: testToken, origin: "http://" + testHost, scope: tc.scope, body: strings.NewReader(tc.body)})
		if w.Code != tc.status {
			t.Fatalf("%s title = %d, want %d", tc.method, w.Code, tc.status)
		}
	}
}
