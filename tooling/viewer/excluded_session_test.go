package viewer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"
)

func setExcludedSourceStatus(e *env, status string) {
	e.store.tracking = func(_ context.Context, paths []string) map[string]SourceClassification {
		out := map[string]SourceClassification{}
		for _, path := range paths {
			out[path] = SourceClassification{Status: status}
		}
		return out
	}
	e.sync()
}

func excludedSessionsJSON(t *testing.T, v any) []Session {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Excluded []Session `json:"excluded_sessions"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Excluded == nil {
		t.Fatal("excluded_sessions must be an array")
	}
	return out.Excluded
}

func TestExcludedSessionProjectionAndReintegration(t *testing.T) {
	s := newSrvEnv(t)
	original := s.onlySession()
	if err := s.store.Pin(s.key, true); err != nil {
		t.Fatal(err)
	}
	setExcludedSourceStatus(s.env, "tracked")
	view := s.store.View()
	excluded := excludedSessionsJSON(t, view)
	if len(view.Projects) != 0 || len(s.sessions()) != 0 || len(excluded) != 1 || excluded[0].Key != s.key || !excluded[0].Pinned {
		t.Fatalf("incorrect exclusion: %+v excluded=%+v", view, excluded)
	}
	w := s.do(http.MethodGet, "/api/index", reqOpts{token: testToken})
	var api struct {
		Projects []indexProject `json:"projects"`
		Excluded []Session      `json:"excluded_sessions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &api); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(api.Projects) != 0 || len(api.Excluded) != 1 || api.Excluded[0].SourceStatus != "tracked" {
		t.Fatalf("API projection wrong: status=%d %+v", w.Code, api)
	}
	setExcludedSourceStatus(s.env, "local")
	got := s.onlySession()
	if len(excludedSessionsJSON(t, s.store.View())) != 0 || got.Key != original.Key || !got.Pinned || got.ContentHash != original.ContentHash || !got.IngestedAt.Equal(original.IngestedAt) {
		t.Fatalf("reintegration changed copy: %+v", got)
	}
}

func TestExcludedSessionDeniesDocumentAndFilesAfterAuthentication(t *testing.T) {
	s := newSrvEnv(t)
	document, err := s.store.Manifest(s.key)
	if err != nil {
		t.Fatal(err)
	}
	setExcludedSourceStatus(s.env, "tracked")
	for _, target := range []string{"/api/sessions/" + s.key, s.files() + "images/nav-1.png"} {
		w := s.do(http.MethodGet, target, reqOpts{token: testToken})
		if w.Code != http.StatusConflict {
			t.Errorf("GET %s returned %d, want 409", target, w.Code)
		}
		if w.Body.String() != "session source is tracked\n" {
			t.Fatalf("unexpected exclusion response: %q", w.Body.String())
		}
		if bytes.Contains(w.Body.Bytes(), document) || bytes.Contains(w.Body.Bytes(), []byte("action_required")) || bytes.Contains(w.Body.Bytes(), []byte("PNG")) {
			t.Fatal("excluded response leaked document bytes")
		}
	}
	if w := s.do(http.MethodHead, s.files()+"images/nav-1.png", reqOpts{}); w.Code != http.StatusConflict {
		t.Fatalf("HEAD file = %d", w.Code)
	}
	for _, target := range []string{"/api/sessions/" + s.key, "/files/" + s.key + "/invalid/images/nav-1.png"} {
		if w := s.do(http.MethodGet, target, reqOpts{}); w.Code != http.StatusForbidden {
			t.Fatalf("unauthorized GET = %d", w.Code)
		}
	}
	if data, err := s.store.Manifest(s.key); !errors.Is(err, ErrExcluded) || len(data) != 0 {
		t.Fatal("Store.Manifest allowed excluded copy")
	}
	if f, err := s.store.OpenFile(s.key, "images/nav-1.png"); !errors.Is(err, ErrExcluded) || f != nil {
		if f != nil {
			f.Close()
		}
		t.Fatal("Store.OpenFile allowed excluded copy")
	}
	for _, status := range []string{"local", "unknown"} {
		setExcludedSourceStatus(s.env, status)
		if w := s.do(http.MethodGet, "/api/sessions/"+s.key, reqOpts{token: testToken}); w.Code != http.StatusOK {
			t.Fatalf("%s document = %d", status, w.Code)
		}
		if w := s.do(http.MethodGet, s.files()+"images/nav-1.png", reqOpts{}); w.Code != http.StatusOK {
			t.Fatalf("%s file = %d", status, w.Code)
		}
	}
}

func TestExcludedSessionPinDeleteAndRetention(t *testing.T) {
	s := newSrvEnv(t)
	setExcludedSourceStatus(s.env, "tracked")
	if w := s.write(http.MethodPost, "/api/sessions/"+s.key+"/pin"); w.Code != http.StatusNoContent {
		t.Fatalf("pin = %d", w.Code)
	}
	s.advance(s.store.TTL() + time.Second)
	if purged, err := s.store.Purge(); err != nil || len(purged) != 0 {
		t.Fatalf("pinned excluded copy purged: %v %v", purged, err)
	}
	if err := os.Remove(s.store.indexPath()); err != nil {
		t.Fatal(err)
	}
	rebuilt, err := Open(Options{StateDir: s.state, Now: s.clock, SourceTracking: s.store.tracking})
	if err != nil {
		t.Fatal(err)
	}
	excluded := excludedSessionsJSON(t, rebuilt.View())
	if len(rebuilt.View().Projects) != 0 || len(excluded) != 1 || !excluded[0].Pinned || excluded[0].Key != s.key {
		t.Fatalf("reconstruction lost exclusion/pin: %+v", excluded)
	}
	if w := s.write(http.MethodDelete, "/api/sessions/"+s.key+"/pin"); w.Code != http.StatusNoContent {
		t.Fatalf("unpin = %d", w.Code)
	}
	if purged, err := s.store.Purge(); err != nil || len(purged) != 1 || purged[0] != s.key {
		t.Fatalf("expired excluded copy not purged: %v %v", purged, err)
	}
	if len(s.store.current().Tombstones) != 1 {
		t.Fatal("purge lost tombstone")
	}
	other := newSrvEnv(t)
	setExcludedSourceStatus(other.env, "tracked")
	if w := other.write(http.MethodDelete, "/api/sessions/"+other.key); w.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", w.Code)
	}
	if len(excludedSessionsJSON(t, other.store.View())) != 0 || len(other.store.current().Sessions) != 0 || len(other.store.current().Tombstones) != 1 {
		t.Fatal("explicit deletion did not remove excluded copy")
	}
}

func TestExcludedSessionMixedProjectionAndWarningUpdates(t *testing.T) {
	s := newSrvEnv(t)
	original := s.onlySession()
	s.place(s.repo, "other", nil)
	s.sync()
	s.store.tracking = func(_ context.Context, paths []string) map[string]SourceClassification {
		out := map[string]SourceClassification{}
		for _, path := range paths {
			out[path] = SourceClassification{Status: "unknown", Reason: "git_query_failed"}
			if path == s.folder+string(os.PathSeparator)+manifestFile {
				out[path] = SourceClassification{Status: "tracked"}
			}
		}
		return out
	}
	s.sync()
	if len(s.sessions()) != 1 || len(excludedSessionsJSON(t, s.store.View())) != 1 {
		t.Fatal("mixed projection duplicated a session")
	}
	for _, status := range []string{"local", "unknown", "local"} {
		setExcludedSourceStatus(s.env, status)
		w := s.do(http.MethodGet, "/api/index", reqOpts{token: testToken})
		var response struct {
			Projects []indexProject `json:"projects"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		count, found := 0, false
		for _, project := range response.Projects {
			for _, ticket := range project.Tickets {
				for _, sess := range ticket.Sessions {
					count++
					if sess.Key != s.key {
						continue
					}
					found = true
					if sess.SourceStatus != status || sess.ContentHash != original.ContentHash || !sess.IngestedAt.Equal(original.IngestedAt) {
						t.Fatalf("stale source projection: %+v", sess)
					}
					if status == "unknown" && sess.SourceReason == "" {
						t.Fatal("unknown lost warning reason")
					}
					if status == "local" && sess.SourceReason != "" {
						t.Fatal("local retained stale warning reason")
					}
				}
			}
		}
		if !found || count != 2 || len(excludedSessionsJSON(t, s.store.View())) != 0 {
			t.Fatal("reintegration count is incorrect")
		}
	}
}
