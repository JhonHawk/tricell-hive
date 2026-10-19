package viewer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSourceClassificationMissingReader(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "local", nil)
	e.register(repo)
	e.start()
	e.sync()
	raw, err := json.Marshal(e.onlySession())
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got["source_status"] != "unknown" || got["source_reason"] == "" || got["source_reason"] == nil {
		t.Fatalf("missing reader must warn unknown, got status=%v reason=%v", got["source_status"], got["source_reason"])
	}
}

func TestSourceTrackingTransitionsAndReconstruction(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run", nil)
	path := filepath.Join(folder, manifestFile)
	status := "local"
	calls := 0
	reader := func(ctx context.Context, paths []string) map[string]SourceClassification {
		calls++
		got := map[string]SourceClassification{}
		for _, p := range paths {
			got[p] = SourceClassification{Status: status, Reason: "git_query_failed"}
		}
		return got
	}
	e.store.tracking = reader
	e.register(repo)
	e.start()
	e.sync()
	original := e.onlySession()
	if original.SourceStatus != "local" {
		t.Fatal(original)
	}
	if err := e.store.Pin(original.Key, true); err != nil {
		t.Fatal(err)
	}
	hashes := 0
	e.store.onHash = func(string) { hashes++ }
	status = "tracked"
	e.sync()
	got := e.store.current().Sessions[0]
	if got.SourceStatus != "tracked" || !got.Pinned || got.ContentHash != original.ContentHash || !got.IngestedAt.Equal(original.IngestedAt) || got.Key != original.Key {
		t.Fatalf("classification changed copy metadata: %+v", got)
	}
	status = "local"
	e.sync()
	if got := e.store.current().Sessions[0]; got.SourceStatus != "local" {
		t.Fatal(got)
	}
	if hashes != 0 {
		t.Fatalf("index-only changes rehashed source %d times", hashes)
	}
	status = "unknown"
	e.sync()
	if got := e.store.current().Sessions[0]; got.SourceStatus != "unknown" || got.SourceReason != "git_query_failed" {
		t.Fatal(got)
	}
	// A deleted source remains tracked when Git still lists its path.
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	status = "tracked"
	e.sync()
	if got := e.store.current().Sessions[0]; got.SourceStatus != "tracked" {
		t.Fatal(got)
	}
	// Without positive tracking evidence, absence must never be called local.
	status = "local"
	e.sync()
	if got := e.store.current().Sessions[0]; got.SourceStatus != "unknown" || got.SourceReason != "source_unavailable" {
		t.Fatal(got)
	}
	if err := os.Remove(e.store.indexPath()); err != nil {
		t.Fatal(err)
	}
	status = "tracked"
	reopened, err := Open(Options{StateDir: e.state, Now: e.clock, SourceTracking: reader})
	if err != nil {
		t.Fatal(err)
	}
	got = reopened.current().Sessions[0]
	if got.SourceStatus != "tracked" || !got.Pinned || got.ContentHash != original.ContentHash || !got.IngestedAt.Equal(original.IngestedAt) {
		t.Fatalf("reconstruction lost classification or metadata: %+v", got)
	}
	if calls < 7 {
		t.Fatalf("classification reused stale cache: %d calls", calls)
	}
}

func TestSourceTrackingSkipsBeforeReadingAndBatches(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	tracked := e.place(repo, "tracked", nil)
	local := e.place(repo, "local", nil)
	// Invalid bytes prove the tracked source is not even validated.
	if err := os.WriteFile(filepath.Join(tracked, manifestFile), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	e.register(repo)
	calls := 0
	e.store.tracking = func(ctx context.Context, paths []string) map[string]SourceClassification {
		calls++
		if len(paths) != 2 {
			t.Fatalf("reader did not receive all candidates: %v", paths)
		}
		return map[string]SourceClassification{
			filepath.Join(tracked, manifestFile): {Status: "tracked"},
			filepath.Join(local, manifestFile):   {Status: "local"},
		}
	}
	res := e.sync()
	if res.Added != 1 || res.Errors != 0 || calls != 1 || e.onlySession().Source != local {
		t.Fatalf("unexpected ingestion: %+v, calls=%d", res, calls)
	}
	if len(e.store.current().Tombstones) != 0 {
		t.Fatal("classification created tombstone")
	}
}

func TestSourceClassificationCancellationWithoutCandidates(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := e.store.Sync(ctx); err != context.Canceled {
		t.Fatalf("cancelled empty pass returned %v", err)
	}
}
