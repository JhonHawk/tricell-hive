package viewer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func assertArchived(t *testing.T, sess Session, want bool) {
	t.Helper()
	data, err := json.Marshal(sess)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	var got bool
	raw, ok := fields["archived"]
	if !ok {
		t.Fatal("session must expose archived, including false")
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("archived = %v, want %v", got, want)
	}
}

func setArchive(t *testing.T, s *srvEnv, archived bool) {
	t.Helper()
	method := http.MethodDelete
	if archived {
		method = http.MethodPost
	}
	w := s.write(method, "/api/sessions/"+s.key+"/archive")
	if w.Code != http.StatusNoContent {
		t.Fatalf("%s archive = %d, want 204", method, w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatal("archive response must be empty")
	}
}

func TestArchivePersistsThroughSyncReplacementAndIndexLoss(t *testing.T) {
	s := newSrvEnv(t)
	original := s.onlySession()
	source := readFile(t, filepath.Join(s.folder, manifestFile))
	copied := readFile(t, filepath.Join(s.sessionDir(s.key), manifestFile))
	assertArchived(t, original, false)
	setArchive(t, s, true)
	setArchive(t, s, true)
	s.sync()
	assertArchived(t, s.onlySession(), true)
	if err := s.store.Pin(s.key, true); err != nil {
		t.Fatal(err)
	}
	assertArchived(t, s.onlySession(), true)
	if err := os.WriteFile(filepath.Join(s.folder, "images", "nav-2.png"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.advance(time.Minute)
	s.sync()
	assertArchived(t, s.onlySession(), true)
	if !s.onlySession().Pinned {
		t.Fatal("archive changed pin")
	}
	if err := os.Remove(s.store.indexPath()); err != nil {
		t.Fatal(err)
	}
	s.store = s.open()
	s.start()
	assertArchived(t, s.onlySession(), true)
	if s.onlySession().SessionStatus != original.SessionStatus {
		t.Fatal("archive changed lifecycle status")
	}
	if !bytes.Equal(source, readFile(t, filepath.Join(s.folder, manifestFile))) || !bytes.Equal(copied, readFile(t, filepath.Join(s.sessionDir(s.key), manifestFile))) {
		t.Fatal("archive changed manifest")
	}
	// The restarted server must use the rebuilt store.
	s.srv.store = s.store
	setArchive(t, s, false)
	setArchive(t, s, false)
	s.store = s.open()
	s.start()
	assertArchived(t, s.onlySession(), false)
	if !s.onlySession().Pinned {
		t.Fatal("unarchive changed pin")
	}
}

func TestArchiveDoesNotChangeRetentionOrCopyDeletion(t *testing.T) {
	s := newSrvEnv(t)
	original := s.onlySession()
	source := readFile(t, filepath.Join(s.folder, manifestFile))
	setArchive(t, s, true)
	got := s.onlySession()
	if got.Pinned || !got.IngestedAt.Equal(original.IngestedAt) {
		t.Fatal("archive changed retention")
	}
	s.advance(defaultTTL + time.Second)
	purged, err := s.store.Purge()
	if err != nil || len(purged) != 1 || purged[0] != s.key {
		t.Fatalf("archived unpinned copy not purged: %v %v", purged, err)
	}
	if !bytes.Equal(source, readFile(t, filepath.Join(s.folder, manifestFile))) {
		t.Fatal("purge changed source")
	}
	s = newSrvEnv(t)
	source = readFile(t, filepath.Join(s.folder, manifestFile))
	setArchive(t, s, true)
	if w := s.write(http.MethodDelete, "/api/sessions/"+s.key); w.Code != http.StatusNoContent {
		t.Fatalf("delete archived = %d", w.Code)
	}
	s.sync()
	if len(s.sessions()) != 0 || !bytes.Equal(source, readFile(t, filepath.Join(s.folder, manifestFile))) {
		t.Fatal("delete did not remain copy-only")
	}
}

func TestArchiveExcludedCopiesRemainManageable(t *testing.T) {
	s := newSrvEnv(t)
	setExcludedSourceStatus(s.env, "tracked")
	setArchive(t, s, true)
	excluded := s.store.View().ExcludedSessions
	if len(excluded) != 1 {
		t.Fatalf("excluded copies = %d", len(excluded))
	}
	assertArchived(t, excluded[0], true)
	setArchive(t, s, false)
	assertArchived(t, s.store.View().ExcludedSessions[0], false)
}

func TestArchiveIndexIncludesArchivedCopies(t *testing.T) {
	s := newSrvEnv(t)
	setArchive(t, s, true)
	w := s.do(http.MethodGet, "/api/index", reqOpts{token: testToken})
	var v View
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || len(v.Projects) != 1 || len(v.Projects[0].Tickets[0].Sessions) != 1 {
		t.Fatalf("archived copy missing from index: %d %s", w.Code, w.Body.String())
	}
	assertArchived(t, v.Projects[0].Tickets[0].Sessions[0], true)
}
