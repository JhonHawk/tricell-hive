package viewer

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestScanRejectsInvalidManifestsListsThemAndKeepsGoing(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	good := e.place(repo, "good", nil)

	// A symlink under images/ (AC11).
	linked := filepath.Join(repo, "_support", "workspace", "linked")
	copyTree(t, filepath.Join(fixturesRoot, "invalid", "symlink"), linked)
	if err := os.Symlink("real.png", filepath.Join(linked, "images", "link.png")); err != nil {
		t.Fatal(err)
	}
	// A path that escapes the folder.
	escape := filepath.Join(repo, "_support", "workspace", "escape")
	copyTree(t, filepath.Join(fixturesRoot, "invalid", "dotdot-path"), escape)
	// Not even JSON.
	broken := filepath.Join(repo, "_support", "workspace", "broken")
	os.MkdirAll(broken, 0o755)
	os.WriteFile(filepath.Join(broken, "walkthrough.json"), []byte("{nope"), 0o644)

	e.register(repo)
	e.start()
	res, err := e.store.Sync(context.Background())
	if err != nil {
		t.Fatalf("Sync stopped on invalid input: %v", err)
	}
	if res.Added != 1 || res.Errors != 3 {
		t.Fatalf("result = %+v, want 1 added and 3 errors", res)
	}
	if s := e.onlySession(); s.Source != good {
		t.Fatalf("ingested %s, want %s", s.Source, good)
	}
	view := e.store.View()
	byPath := map[string]string{}
	for _, ie := range view.Errors {
		byPath[ie.Path] = ie.Reason
	}
	for _, folder := range []string{linked, escape, broken} {
		reason, ok := byPath[filepath.Join(folder, "walkthrough.json")]
		if !ok || reason == "" {
			t.Fatalf("no error listed for %s: %v", folder, byPath)
		}
	}
	if !strings.Contains(byPath[filepath.Join(linked, "walkthrough.json")], "flows[0].captures[0].path") {
		t.Fatalf("symlink error does not name the field: %q", byPath[filepath.Join(linked, "walkthrough.json")])
	}
	// The listing is persisted in index.json.
	var persisted struct {
		Errors []IngestError `json:"errors"`
	}
	if err := json.Unmarshal(readFile(t, filepath.Join(e.state, "viewer", "index.json")), &persisted); err != nil || len(persisted.Errors) != 3 {
		t.Fatalf("index.json errors = %v, %v", persisted.Errors, err)
	}
	// Errors clear once the source is fixed.
	os.RemoveAll(broken)
	if res := e.sync(); res.Errors != 2 || len(e.store.View().Errors) != 2 {
		t.Fatalf("stale error kept: %+v", res)
	}
}

func TestScanUnchangedSourcesAreNotCopiedAgain(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	first := e.onlySession()
	e.advance(time.Hour)
	if res := e.sync(); res.Added != 0 || res.Replaced != 0 {
		t.Fatalf("second pass changed things: %+v", res)
	}
	if second := e.onlySession(); !second.IngestedAt.Equal(first.IngestedAt) {
		t.Fatal("ingested_at moved without a content change")
	}
}

func TestScanKeepsOneCopyPerSourceFolderAndSession(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.place(repo, "run2", func(m map[string]any) { m["session"] = map[string]any{"id": "sess-run1"} }) // same id, other folder
	e.register(repo)
	e.start()
	e.sync()
	if len(e.sessions()) != 2 {
		t.Fatalf("same session id in two folders must be two copies, got %d", len(e.sessions()))
	}
	// A new session id in the same folder is a new copy; the old one stays.
	m := readManifestMap(t, folder)
	m["session"] = map[string]any{"id": "sess-continued"}
	writeManifestMap(t, folder, m)
	e.touch(folder)
	e.sync()
	if len(e.sessions()) != 3 {
		t.Fatalf("sessions = %d, want 3", len(e.sessions()))
	}
}

func TestScanDoesNotFollowLinkedFoldersOrManifests(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	outside := t.TempDir()
	copyTree(t, filepath.Join(fixturesRoot, "valid", "final"), outside)
	ws := filepath.Join(repo, "_support", "workspace")
	os.MkdirAll(ws, 0o755)
	// A linked folder is not a work folder.
	os.Symlink(outside, filepath.Join(ws, "linkdir"))
	// A real folder whose manifest is a link.
	linkedManifest := filepath.Join(ws, "linkman")
	os.MkdirAll(linkedManifest, 0o755)
	os.Symlink(filepath.Join(outside, "walkthrough.json"), filepath.Join(linkedManifest, "walkthrough.json"))

	e.register(repo)
	e.start()
	res := e.sync()
	if res.Added != 0 || len(e.sessions()) != 0 {
		t.Fatalf("followed a link: %+v", res)
	}
	if res.Errors != 1 {
		t.Fatalf("the linked manifest must be reported, result %+v", res)
	}
}

func TestScanRefusesALinkedSupportDirectory(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	outside := t.TempDir()
	copyTree(t, filepath.Join(fixturesRoot, "valid", "final"), filepath.Join(outside, "workspace", "run1"))
	if err := os.Symlink(outside, filepath.Join(repo, "_support")); err != nil {
		t.Fatal(err)
	}
	e.register(repo)
	e.start()
	res := e.sync()
	if res.Added != 0 || res.Errors != 1 || len(e.sessions()) != 0 {
		t.Fatalf("linked _support followed: %+v", res)
	}
}

func TestScanMarksAnAbsentRepositoryAndKeepsItsCopies(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	moved := repo + "-moved"
	if err := os.Rename(repo, moved); err != nil {
		t.Fatal(err)
	}
	if res := e.sync(); res.Errors != 0 {
		t.Fatalf("an absent repository is a state, not an error: %+v", res)
	}
	view := e.store.View()
	if len(view.Repositories) != 1 || !view.Repositories[0].Absent || view.Repositories[0].Root != repo {
		t.Fatalf("repositories = %+v", view.Repositories)
	}
	if len(e.sessions()) != 1 || !view.Projects[0].Absent {
		t.Fatalf("copies of an absent repository must be kept and flagged: %+v", view.Projects)
	}
	os.Rename(moved, repo)
	e.sync()
	if v := e.store.View(); v.Repositories[0].Absent || v.Projects[0].Absent {
		t.Fatal("repository still absent after it came back")
	}
}

func TestScanSkipsSourcesOlderThanTheTTL(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	old := e.clock().Add(-15 * 24 * time.Hour)
	os.Chtimes(filepath.Join(folder, "walkthrough.json"), old, old)
	e.register(repo)
	e.start()
	res := e.sync()
	if res.Added != 0 || res.Errors != 0 || len(e.sessions()) != 0 {
		t.Fatalf("stale source ingested: %+v", res)
	}
}

func TestScanTombstoneKeepsADeletedSessionOutUntilItsContentChanges(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	if err := e.store.Delete(key); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		e.advance(5 * time.Second)
		if res := e.sync(); res.Added != 0 || len(e.sessions()) != 0 {
			t.Fatalf("deleted session came back on poll %d: %+v", i, res)
		}
	}
	if len(e.store.View().Tombstones) != 1 {
		t.Fatalf("tombstones = %+v", e.store.View().Tombstones)
	}
	os.WriteFile(filepath.Join(folder, "images", "settings-1.png"), []byte("edited"), 0o644)
	if res := e.sync(); res.Added != 1 {
		t.Fatalf("changed source did not re-enter: %+v", res)
	}
	if len(e.store.View().Tombstones) != 0 {
		t.Fatal("tombstone kept after the session re-entered")
	}
}

func TestScanDropsTombstonesWhoseSourceIsGone(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.store.Delete(e.onlySession().Key)
	e.sync()
	if len(e.store.View().Tombstones) != 1 {
		t.Fatal("expected a tombstone")
	}
	os.RemoveAll(folder)
	e.sync()
	if got := e.store.View().Tombstones; len(got) != 0 {
		t.Fatalf("tombstone for a vanished source kept: %+v", got)
	}
}

func TestScanKeepsTombstonesWhileTheRepositoryIsAbsent(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.store.Delete(e.onlySession().Key)
	moved := repo + "-moved"
	os.Rename(repo, moved)
	e.sync()
	if len(e.store.View().Tombstones) != 1 {
		t.Fatal("tombstone dropped just because the repository is unavailable")
	}
}

func TestScanRegistryReadFailureIsReturnedNotSwallowed(t *testing.T) {
	e := newEnv(t)
	e.start()
	os.WriteFile(filepath.Join(e.state, "viewer", "registry.json"), []byte("{broken"), 0o600)
	if _, err := e.store.Sync(context.Background()); err == nil {
		t.Fatal("corrupt registry ignored")
	}
}

func TestScanCancelledContextStopsBeforeCopying(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.store.Sync(ctx); err == nil {
		t.Fatal("cancelled Sync returned no error")
	}
	if len(e.sessions()) != 0 {
		t.Fatal("copied after cancellation")
	}
}

func TestScanRunIngestsWithinTwoPollIntervals(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.register(repo)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		e.store.Run(ctx, RunOptions{PollInterval: 20 * time.Millisecond, PurgeInterval: time.Hour})
	}()
	// Stop Run before the temporary directories are removed.
	defer func() { cancel(); <-done }()
	time.Sleep(60 * time.Millisecond)
	e.place(repo, "late", nil)
	waitFor(t, "the late walkthrough to appear", func() bool { return len(e.sessions()) == 1 })
}

func TestScanFailedCopyLeavesNoTemporaryDirectory(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	// Make a referenced file unreadable so the copy fails after it began.
	if os.Geteuid() == 0 {
		t.Skip("permissions are not enforced for root")
	}
	if err := os.Chmod(filepath.Join(folder, "images", "nav-3.png"), 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(filepath.Join(folder, "images", "nav-3.png"), 0o644)
	res := e.sync()
	if res.Added != 0 || res.Errors != 1 {
		t.Fatalf("result = %+v", res)
	}
	entries, _ := os.ReadDir(filepath.Join(e.state, "viewer", "sessions"))
	if len(entries) != 0 {
		t.Fatalf("leftovers after a failed copy: %v", entries)
	}
}

// H1: a folder swapped for a link after it was validated must not make the
// hash or the copy read outside it. The swap happens deterministically in a
// hook between validation and copy.
func TestScanFolderSwappedForALinkAfterValidationCopiesNothingFromOutside(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()

	outside := filepath.Join(canon(t, t.TempDir()), "outside")
	copyTree(t, folder, outside)
	secret := []byte("OUTSIDE SECRET BYTES")
	for _, rel := range []string{"images/nav-1.png", "images/nav-2.png", "images/nav-3.png", "images/settings-1.png"} {
		if err := os.WriteFile(filepath.Join(outside, rel), secret, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	original := readFile(t, filepath.Join(folder, "images", "nav-1.png"))
	aside := filepath.Join(canon(t, t.TempDir()), "aside")
	e.store.afterValidate = func(f string) {
		if err := os.Rename(f, aside); err != nil {
			t.Error(err)
			return
		}
		if err := os.Symlink(outside, f); err != nil {
			t.Error(err)
		}
	}
	e.sync()

	err := filepath.WalkDir(filepath.Join(e.state, "viewer", "sessions"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if bytes.Contains(readFile(t, path), secret) {
			t.Errorf("%s holds bytes read from outside the validated folder", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if sessions := e.sessions(); len(sessions) == 1 {
		if got := readFile(t, filepath.Join(e.sessionDir(sessions[0].Key), "images", "nav-1.png")); !bytes.Equal(got, original) {
			t.Error("the copy must come from the folder that was validated")
		}
	}
}

// H4: an unknown field name is echoed into the ingest errors; it must be
// bounded and free of control characters, and the errors of one repository
// are capped.
func TestScanBoundsIngestErrorTextAndCountPerRepository(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	hostile := "\x1b[31m\nINJECT" + strings.Repeat("x", 5000)
	for i := 0; i < 30; i++ {
		e.place(repo, "bad"+strings.Repeat("a", i%3)+string(rune('a'+i/3)), func(m map[string]any) { m[hostile] = true })
	}
	e.register(repo)
	e.start()
	res := e.sync()
	if res.Errors != 30 {
		t.Fatalf("Errors = %d, want all 30 counted", res.Errors)
	}
	errs := e.store.View().Errors
	if len(errs) != maxIngestErrorsPerRepository+1 {
		t.Fatalf("stored errors = %d, want %d plus one summary", len(errs), maxIngestErrorsPerRepository)
	}
	omitted := 0
	for _, ie := range errs {
		if n := len([]rune(ie.Reason)); n > maxErrorText+1 {
			t.Errorf("reason has %d runes", n)
		}
		if strings.ContainsAny(ie.Reason+ie.Path, "\x1b\n\r") {
			t.Errorf("control characters in %q", ie.Reason)
		}
		if strings.Contains(ie.Reason, "more ingest errors omitted") {
			omitted++
		}
	}
	if omitted != 1 {
		t.Fatalf("want one summary of omitted errors, got %d", omitted)
	}
}
