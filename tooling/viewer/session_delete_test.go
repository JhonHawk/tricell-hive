package viewer

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionGroupDelete(t *testing.T) {
	e, repo := groupFixture(t)
	original := e.sessions()
	if err := e.store.Delete(original[0].Key); err != nil {
		t.Fatal(err)
	}
	if len(e.sessions()) != 0 {
		t.Fatalf("session siblings survived: %d", len(e.sessions()))
	}
	e.sync()
	if len(e.sessions()) != 0 {
		t.Fatal("identical content returned")
	}
	for _, sess := range original {
		if _, err := os.Stat(filepath.Join(sess.Source, manifestFile)); err != nil {
			t.Fatal("source modified", err)
		}
	}
	e.place(repo, "later", func(m map[string]any) { m["session"] = map[string]any{"id": "shared", "host": "codex"} })
	e.sync()
	if len(e.sessions()) != 1 {
		t.Fatal("new delivery blocked")
	}
}

func TestSessionGroupDeletePrecommitFailures(t *testing.T) {
	for _, stage := range []string{"tombstones", "marker"} {
		t.Run(stage, func(t *testing.T) {
			e, _ := groupFixture(t)
			sess := e.sessions()[0]
			groupPath := e.store.groupPath(identityOf(sess))
			realWrite := e.store.writeAtomic
			e.store.writeAtomic = func(path string, data []byte) error {
				if stage == "tombstones" && path == e.store.indexPath() {
					return errors.New("index unavailable")
				}
				if stage == "marker" && path == groupPath {
					var g sessionGroup
					_ = json.Unmarshal(data, &g)
					if g.DeletePending {
						return errors.New("group unavailable")
					}
				}
				return realWrite(path, data)
			}
			if err := e.store.Delete(sess.Key); err == nil {
				t.Fatal("failed persistence returned success")
			}
			if len(e.sessions()) != 2 {
				t.Fatal("precommit failure removed copies")
			}
			for _, member := range e.sessions() {
				if _, err := os.Stat(e.sessionDir(member.Key)); err != nil {
					t.Fatal("copy removed before persistence", err)
				}
			}
			if stage == "marker" {
				d, ok := e.store.readIndexFile()
				if !ok || len(d.Tombstones) != 2 {
					t.Fatal("tombstones not committed before marker")
				}
			}
		})
	}
}

func TestSessionGroupDeletePartialSurvivesRestart(t *testing.T) {
	e, _ := groupFixture(t)
	members := e.sessions()
	survivor := members[0]
	realRename := e.store.renamePath
	e.store.renamePath = func(from, to string) error {
		if from == e.sessionDir(survivor.Key) {
			return errors.New("copy locked")
		}
		return realRename(from, to)
	}
	var partial *PartialDeleteError
	if err := e.store.Delete(survivor.Key); !errors.As(err, &partial) || partial.Deleted != 1 || partial.Remaining != 1 {
		t.Fatalf("partial result: %#v %v", partial, err)
	}
	if len(e.sessions()) != 1 || e.sessions()[0].Key != survivor.Key {
		t.Fatal("published stale survivors")
	}
	e.store = e.open()
	e.start()
	e.sync()
	if len(e.sessions()) != 1 || e.sessions()[0].Key != survivor.Key {
		t.Fatal("restart deleted survivor or resurrected removed member")
	}
	if err := e.store.Delete(survivor.Key); err != nil {
		t.Fatal(err)
	}
	if len(e.sessions()) != 0 {
		t.Fatal("retry did not delete survivor")
	}
}

func TestSessionGroupDeleteFinalIndexFailure(t *testing.T) {
	e, _ := groupFixture(t)
	key := e.sessions()[0].Key
	realWrite := e.store.writeAtomic
	writes := 0
	e.store.writeAtomic = func(path string, data []byte) error {
		if path == e.store.indexPath() {
			writes++
			if writes == 2 {
				return errors.New("index unavailable after deletion")
			}
		}
		return realWrite(path, data)
	}
	var partial *PartialDeleteError
	if err := e.store.Delete(key); !errors.As(err, &partial) || partial.Deleted != 2 || partial.Remaining != 0 {
		t.Fatalf("final commit failure: %#v %v", partial, err)
	}
	if len(e.sessions()) != 0 {
		t.Fatal("failed final commit republished deleted copies")
	}
	e.store = e.open()
	e.start()
	e.sync()
	if len(e.sessions()) != 0 {
		t.Fatal("deleted content resurrected after restart")
	}
}

func TestSessionGroupDeletePendingRecoveryWithoutRestart(t *testing.T) {
	e, repo := groupFixture(t)
	sess := e.sessions()[0]
	if err := e.store.SetTitle(sess.Key, "Old title"); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Archive(sess.Key, true); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Pin(sess.Key, true); err != nil {
		t.Fatal(err)
	}
	path := e.store.groupPath(identityOf(sess))
	realRemove := e.store.removeFile
	e.store.removeFile = func(p string) error {
		if p == path {
			return errors.New("cleanup unavailable")
		}
		return realRemove(p)
	}
	var partial *PartialDeleteError
	if err := e.store.Delete(sess.Key); !errors.As(err, &partial) || !partial.CleanupPending || partial.Remaining != 0 {
		t.Fatalf("cleanup result: %#v %v", partial, err)
	}
	if len(e.store.View().Errors) == 0 {
		t.Fatal("pending cleanup invisible without member")
	}
	manifest := readManifestMap(t, sess.Source)
	manifest["title"] = "Modified delivery"
	writeManifestMap(t, sess.Source, manifest)
	e.touch(sess.Source)
	e.place(repo, "unrelated", nil)
	res, err := e.store.Sync(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Added != 1 || res.Errors != 1 || len(e.sessions()) != 1 {
		t.Fatalf("cleanup blocked unrelated library or ingested stale preferences: %+v", res)
	}
	if e.sessions()[0].SessionID == sess.SessionID {
		t.Fatal("pending identity was ingested")
	}
	e.store.removeFile = realRemove
	e.sync()
	if len(e.sessions()) != 2 {
		t.Fatal("recovered delivery did not arrive")
	}
	for _, member := range e.sessions() {
		if member.SessionID == sess.SessionID && (member.Pinned || member.Archived || member.SessionTitle != "") {
			t.Fatalf("obsolete preferences inherited: %+v", member)
		}
	}
	for _, problem := range e.store.View().Errors {
		if strings.Contains(problem.Code, "cleanup") {
			t.Fatal("recovered diagnostic retained")
		}
	}
}

func TestSessionGroupDeleteInterruptedBeforeRemoval(t *testing.T) {
	for _, stage := range []string{"tombstones", "marker"} {
		t.Run(stage, func(t *testing.T) {
			e, _ := groupFixture(t)
			sess := e.sessions()[0]
			path := e.store.groupPath(identityOf(sess))
			realWrite := e.store.writeAtomic
			e.store.writeAtomic = func(p string, data []byte) error {
				if err := realWrite(p, data); err != nil {
					return err
				}
				if stage == "tombstones" && p == e.store.indexPath() {
					return errors.New("interrupted after index rename")
				}
				if stage == "marker" && p == path {
					var g sessionGroup
					_ = json.Unmarshal(data, &g)
					if g.DeletePending {
						return errors.New("interrupted after marker rename")
					}
				}
				return nil
			}
			if err := e.store.Delete(sess.Key); err == nil {
				t.Fatal("interruption not reported")
			}
			e.store = e.open()
			e.start()
			e.sync()
			if len(e.sessions()) != 2 {
				t.Fatal("restart automatically removed surviving copies")
			}
			if err := e.store.Delete(sess.Key); err != nil {
				t.Fatal(err)
			}
			e.sync()
			if len(e.sessions()) != 0 {
				t.Fatal("confirmed retry did not finish deletion")
			}
		})
	}
}

func TestSessionGroupDeletePreingestRecovery(t *testing.T) {
	e, _ := groupFixture(t)
	sess := e.sessions()[0]
	if err := e.store.SetTitle(sess.Key, "Deleted name"); err != nil {
		t.Fatal(err)
	}
	realRemove := e.store.removeFile
	e.store.removeFile = func(string) error { return errors.New("unavailable") }
	if err := e.store.Delete(sess.Key); err == nil {
		t.Fatal("cleanup failure not reported")
	}
	m := readManifestMap(t, sess.Source)
	m["title"] = "New delivery"
	writeManifestMap(t, sess.Source, m)
	e.touch(sess.Source)
	calls := 0
	e.store.removeFile = func(path string) error {
		calls++
		if calls == 1 {
			return errors.New("transient failure")
		}
		return realRemove(path)
	}
	e.sync()
	if calls < 2 || len(e.sessions()) != 1 || e.onlySession().SessionTitle != "" {
		t.Fatalf("preingest did not recover safely: calls=%d sessions=%+v", calls, e.sessions())
	}
}

func TestSessionGroupDeleteFolderCleanupRetries(t *testing.T) {
	e, _ := groupFixture(t)
	sess := e.sessions()[0]
	realRemove := e.store.removeTree
	e.store.removeTree = func(path string) error {
		if strings.HasPrefix(filepath.Base(path), tmpPrefix+"del-") {
			if _, err := os.Stat(path); err == nil {
				return errors.New("folder cleanup unavailable")
			}
		}
		return realRemove(path)
	}
	if err := e.store.Delete(sess.Key); err == nil {
		t.Fatal("folder cleanup failure not reported")
	}
	e.sync()
	if _, err := os.Stat(e.store.groupPath(identityOf(sess))); err != nil {
		t.Fatal("cleanup marked complete while deleted folders remain")
	}
	e.store.removeTree = realRemove
	e.sync()
	entries, err := os.ReadDir(e.store.sessionsDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("automatic recovery left deleted folders")
	}
}
