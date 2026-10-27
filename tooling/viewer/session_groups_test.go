package viewer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func groupFixture(t *testing.T) (*env, string) {
	t.Helper()
	e := newEnv(t)
	repo := newRepo(t)
	for _, name := range []string{"one", "two"} {
		e.place(repo, name, func(m map[string]any) {
			m["session"] = map[string]any{"id": "shared", "host": "codex"}
			m["ticket"] = name
		})
	}
	e.register(repo)
	e.start()
	e.sync()
	return e, repo
}

func TestSessionGroupMigration(t *testing.T) {
	e, _ := groupFixture(t)
	sessions := e.sessions()
	// Recreate a library written by the old binary.
	os.RemoveAll(filepath.Join(e.store.Dir(), "session-groups"))
	for i, sess := range sessions {
		m, err := readMeta(e.sessionDir(sess.Key))
		if err != nil {
			t.Fatal(err)
		}
		m.Archived = i == 0
		m.Pinned = i == 1
		if err := writeJSONAtomic(filepath.Join(e.sessionDir(sess.Key), "meta.json"), m); err != nil {
			t.Fatal(err)
		}
	}
	e.store = e.open()
	e.start()
	for _, sess := range e.sessions() {
		if sess.Archived || !sess.Pinned {
			t.Fatalf("mixed migration = %+v", sess)
		}
	}
}

func TestSessionGroupState(t *testing.T) {
	e, _ := groupFixture(t)
	original := e.sessions()
	key := original[0].Key
	if err := e.store.SetTitle(key, "  Session 👩🏽‍💻  "); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Archive(key, true); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(e.store.indexPath()); err != nil {
		t.Fatal(err)
	}
	e.store = e.open()
	e.start()
	for _, sess := range e.sessions() {
		if sess.SessionTitle != "Session 👩🏽‍💻" || !sess.Archived || !sess.Pinned {
			t.Fatalf("lost preferences: %+v", sess)
		}
		for _, old := range original {
			if old.Key == sess.Key && (old.Title != sess.Title || old.ContentHash != sess.ContentHash || !old.IngestedAt.Equal(sess.IngestedAt)) {
				t.Fatal("changed copy identity or content")
			}
		}
	}
	if err := e.store.SetTitle(key, ""); err != nil {
		t.Fatal(err)
	}
	e.store = e.open()
	e.start()
	if e.sessions()[0].SessionTitle != "" {
		t.Fatal("reset did not survive restart")
	}
	for _, title := range []string{"line\nbreak", string(make([]byte, 201))} {
		if err := e.store.SetTitle(key, title); err == nil {
			t.Fatal("invalid title accepted")
		}
	}
}

func TestSessionGroupIngestion(t *testing.T) {
	e, repo := groupFixture(t)
	key := e.sessions()[0].Key
	if err := e.store.Archive(key, true); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}
	e.place(repo, "later", func(m map[string]any) { m["session"] = map[string]any{"id": "shared", "host": "codex"} })
	e.place(repo, "other", func(m map[string]any) { m["session"] = map[string]any{"id": "shared", "host": "claude-code"} })
	e.sync()
	for _, sess := range e.sessions() {
		want := sess.Host == "codex"
		if sess.Pinned != want || sess.Archived != want {
			t.Fatalf("wrong inheritance: %+v", sess)
		}
	}
}

func TestSessionGroupRecoveryCorruptBlocksIngestion(t *testing.T) {
	e, repo := groupFixture(t)
	sess := e.sessions()[0]
	path := e.store.groupPath(identityOf(sess))
	if err := os.WriteFile(path, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	e.place(repo, "later", func(m map[string]any) { m["session"] = map[string]any{"id": "shared", "host": "codex"} })
	res, _ := e.store.Sync(context.Background())
	if res.Added != 0 {
		t.Fatalf("installed a copy with corrupt preferences: %+v", res)
	}
	e.advance(defaultTTL + time.Second)
	if keys, err := e.store.Purge(); err == nil || len(keys) != 0 {
		t.Fatalf("unsafe purge: %v %v", keys, err)
	}
	if len(e.store.View().Errors) == 0 {
		t.Fatal("corruption is not visible")
	}
}

func TestSessionGroupRetention(t *testing.T) {
	e, repo := groupFixture(t)
	key := e.sessions()[0].Key
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}
	e.place(repo, "continued", func(m map[string]any) { m["session"] = map[string]any{"id": "shared", "host": "codex"} })
	e.sync()
	e.advance(defaultTTL + time.Second)
	if keys, err := e.store.Purge(); err != nil || len(keys) != 0 {
		t.Fatalf("pinned members expired: %v %v", keys, err)
	}
	if err := e.store.Pin(key, false); err != nil {
		t.Fatal(err)
	}
	if keys, err := e.store.Purge(); err != nil || len(keys) != 3 {
		t.Fatalf("unpin did not restore retention: %v %v", keys, err)
	}
	if _, err := os.Stat(e.store.groupPath(identityOf(Session{ProjectKey: repo, SessionID: "shared", Host: "codex"}))); err != nil {
		t.Fatal("TTL removed logical preferences", err)
	}
}

func TestSessionGroupMigrationInterrupted(t *testing.T) {
	e, _ := groupFixture(t)
	original := e.sessions()
	if err := e.store.Archive(original[0].Key, true); err != nil {
		t.Fatal(err)
	}
	for _, sess := range original {
		m, err := readMeta(e.sessionDir(sess.Key))
		if err != nil {
			t.Fatal(err)
		}
		m.Archived = false
		m.Pinned = true
		if err := writeJSONAtomic(filepath.Join(e.sessionDir(sess.Key), "meta.json"), m); err != nil {
			t.Fatal(err)
		}
	}
	for n := 0; n < 2; n++ {
		e.store = e.open()
		e.start()
		for _, sess := range e.sessions() {
			if !sess.Archived || sess.Pinned {
				t.Fatalf("reimport overwrote authority: %+v", sess)
			}
			data := readFile(t, filepath.Join(e.sessionDir(sess.Key), "meta.json"))
			var fields map[string]any
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if _, ok := fields["pinned"]; ok {
				t.Fatal("legacy pin retained")
			}
			if _, ok := fields["archived"]; ok {
				t.Fatal("legacy archive retained")
			}
			root, err := os.OpenRoot(e.sessionDir(sess.Key))
			if err != nil {
				t.Fatal(err)
			}
			manifest, err := validateRoot(root)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := hashFiles(root, manifest)
			root.Close()
			if err != nil || hash != sess.ContentHash {
				t.Fatalf("migration changed resources: %s %v", hash, err)
			}
		}
	}
}

func TestSessionGroupStateWriteFailure(t *testing.T) {
	e, _ := groupFixture(t)
	sess := e.sessions()[0]
	path := e.store.groupPath(identityOf(sess))
	original := readFile(t, path)
	// Atomic rename cannot replace a directory, including when running as root.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := e.store.SetTitle(sess.Key, "Unsaved"); err == nil {
		t.Fatal("false success")
	}
	if e.sessions()[0].SessionTitle != "" {
		t.Fatal("failed write changed view")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	// Failure publishing the reconstructible index must preserve committed prefs.
	if err := os.Remove(e.store.indexPath()); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(e.store.indexPath(), 0700); err != nil {
		t.Fatal(err)
	}
	if err := e.store.SetTitle(sess.Key, "Committed"); err == nil {
		t.Fatal("index failure returned success")
	}
	e.store = e.open()
	if e.sessions()[0].SessionTitle != "Committed" {
		t.Fatal("durable preference lost after index failure")
	}
}

func TestSessionGroupStateIdentityAndValidation(t *testing.T) {
	var absent sessionIdentity
	absent.ProjectKey = "project"
	absent.SessionID = "session"
	if got := absent.key(); got != sha(`["project",null,"session"]`) {
		t.Fatalf("absent host tuple mismatch: %s", got)
	}
	host := "codex"
	absent.Host = &host
	if absent.key() != sha(`["project","codex","session"]`) {
		t.Fatal("declared host tuple mismatch")
	}
	if _, err := validateSessionTitle(strings.Repeat("界", 200)); err != nil {
		t.Fatal(err)
	}
	if _, err := validateSessionTitle(strings.Repeat("界", 201)); err == nil {
		t.Fatal("oversized Unicode title accepted")
	}
}

func TestSessionGroupMigrationFullyArchived(t *testing.T) {
	e, _ := groupFixture(t)
	original := e.sessions()
	if err := os.RemoveAll(e.store.groupsDir()); err != nil {
		t.Fatal(err)
	}
	for _, sess := range original {
		m, err := readMeta(e.sessionDir(sess.Key))
		if err != nil {
			t.Fatal(err)
		}
		m.Archived = true
		if err := writeJSONAtomic(filepath.Join(e.sessionDir(sess.Key), "meta.json"), m); err != nil {
			t.Fatal(err)
		}
	}
	e.store = e.open()
	e.start()
	for _, sess := range e.sessions() {
		if !sess.Archived || sess.Pinned {
			t.Fatalf("fully archived migration failed: %+v", sess)
		}
	}
}

func TestSessionGroupRecoveryRejectsIncompleteAuthority(t *testing.T) {
	e, _ := groupFixture(t)
	sess := e.sessions()[0]
	path := e.store.groupPath(identityOf(sess))
	var fields map[string]any
	if err := json.Unmarshal(readFile(t, path), &fields); err != nil {
		t.Fatal(err)
	}
	delete(fields, "pinned")
	if err := writeJSONAtomic(path, fields); err != nil {
		t.Fatal(err)
	}
	e.advance(defaultTTL + time.Second)
	if keys, err := e.store.Purge(); err == nil || len(keys) > 0 {
		t.Fatalf("missing pin accepted: %v %v", keys, err)
	}
}

func TestSessionGroupStateMissingHostUsesNullTuple(t *testing.T) {
	raw, err := json.Marshal(Session{ProjectKey: "project", SessionID: "session"})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	tuple, err := json.Marshal([]any{fields["project_key"], fields["host"], fields["session_id"]})
	if err != nil {
		t.Fatal(err)
	}
	if sha(string(tuple)) != identityOf(Session{ProjectKey: "project", SessionID: "session"}).key() {
		t.Fatalf("frontend and store identity differ: %s", tuple)
	}
}
