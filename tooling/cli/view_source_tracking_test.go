package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

func trackingTestFile(t *testing.T, root, rel string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestViewerSourceTrackingGitOwnersAndBatching(t *testing.T) {
	repo := newProjectRepo(t)
	if err := os.WriteFile(filepath.Join(repo, ".gitignore"), []byte("_support/\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tracked := trackingTestFile(t, repo, "_support/workspace/one/walkthrough.json")
	local := trackingTestFile(t, repo, "_support/workspace/two/walkthrough.json")
	ancestor := trackingTestFile(t, repo, "nested/_support/workspace/three/walkthrough.json")
	gitIn(t, repo, "add", "-f", "--", tracked, ancestor)
	calls := map[string]int{}
	reader := viewSourceTracking(func(ctx context.Context, dir string, args ...string) ([]byte, error) {
		calls[dir]++
		if args[0] != "ls-files" {
			t.Fatalf("unexpected query: %v", args)
		}
		return runViewerTrackingGit(ctx, dir, args...)
	})
	got := reader(context.Background(), []string{tracked, local, ancestor})
	if got[tracked].Status != "tracked" || got[ancestor].Status != "tracked" || got[local].Status != "local" || calls[repo] != 1 {
		t.Fatalf("classification=%v queries=%v", got, calls)
	}
	if err := os.Remove(tracked); err != nil {
		t.Fatal(err)
	}
	if got := reader(context.Background(), []string{tracked}); got[tracked].Status != "tracked" {
		t.Fatal(got)
	}
	worktree := filepath.Join(t.TempDir(), "tree")
	gitIn(t, repo, "worktree", "add", "-q", "-b", "fixture", worktree)
	worktree, err := filepath.EvalSymlinks(worktree)
	if err != nil {
		t.Fatal(err)
	}
	wp := trackingTestFile(t, worktree, "_support/workspace/worktree/walkthrough.json")
	gitIn(t, worktree, "add", "--", wp)
	if got := reader(context.Background(), []string{wp}); got[wp].Status != "tracked" {
		t.Fatal(got)
	}
	noRepo := trackingTestFile(t, t.TempDir(), "_support/workspace/local/walkthrough.json")
	if got := reader(context.Background(), []string{noRepo}); got[noRepo].Status != "local" {
		t.Fatal(got)
	}
}

func TestViewerSourceTrackingFailureAndCancellation(t *testing.T) {
	repo := newProjectRepo(t)
	path := trackingTestFile(t, repo, "_support/workspace/run/walkthrough.json")
	for _, failure := range []error{errors.New("sensitive stderr must not escape"), context.DeadlineExceeded} {
		reader := viewSourceTracking(func(context.Context, string, ...string) ([]byte, error) { return nil, failure })
		got := reader(context.Background(), []string{path})[path]
		if got.Status != "unknown" || got.Reason != "git_query_failed" {
			t.Fatal(got)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reader := viewSourceTracking(func(ctx context.Context, _ string, _ ...string) ([]byte, error) { return nil, ctx.Err() })
	if got := reader(ctx, []string{path})[path]; got.Status != "unknown" {
		t.Fatal(got)
	}
	// A Git marker that cannot be queried is not a verified no-repository case.
	broken := t.TempDir()
	if err := os.WriteFile(filepath.Join(broken, ".git"), []byte("bad marker"), 0600); err != nil {
		t.Fatal(err)
	}
	bp := trackingTestFile(t, broken, "_support/workspace/run/walkthrough.json")
	if got := viewSourceTracking(runViewerTrackingGit)(context.Background(), []string{bp})[bp]; got.Status != "unknown" {
		t.Fatal(got)
	}
}

func TestViewerSourceTrackingLargeOutputAndTruncation(t *testing.T) {
	repo := newProjectRepo(t)
	var output bytes.Buffer
	for i := 0; i < 2200; i++ {
		fmt.Fprintf(&output, "_support/workspace/prefix-%06d/file.png%c", i, 0)
	}
	rel := "_support/workspace/zzzz/walkthrough.json"
	path := trackingTestFile(t, repo, rel)
	output.WriteString(rel)
	output.WriteByte(0)
	if output.Len() <= 64<<10 {
		t.Fatal("fixture is not beyond old runner limit")
	}
	reader := viewSourceTracking(func(context.Context, string, ...string) ([]byte, error) { return output.Bytes(), nil })
	if got := reader(context.Background(), []string{path})[path]; got.Status != "tracked" {
		t.Fatal(got)
	}
	// Force overflow in a record and directly after NUL, including when an
	// earlier write filled the cap exactly. The writer must signal either case.
	firstNUL := bytes.IndexByte(output.Bytes(), 0) + 1
	for _, limit := range []int{firstNUL - 2, firstNUL, 64 << 10} {
		b := &trackingOutput{limit: limit}
		_, firstErr := b.Write(output.Bytes()[:limit])
		_, overflowErr := b.Write(output.Bytes()[limit:])
		if firstErr != nil || !b.overflow || !errors.Is(overflowErr, errViewerTrackingIncomplete) {
			t.Fatalf("overflow lost at %d", limit)
		}
		truncated := viewSourceTracking(func(context.Context, string, ...string) ([]byte, error) { return b.Bytes(), overflowErr })
		if got := truncated(context.Background(), []string{path})[path]; got.Status != "unknown" || got.Reason != "git_output_incomplete" {
			t.Fatal(got)
		}
	}
	incomplete := viewSourceTracking(func(context.Context, string, ...string) ([]byte, error) {
		return []byte(strings.TrimSuffix(rel, ".json")), nil
	})
	if got := incomplete(context.Background(), []string{path})[path]; got.Status != "unknown" {
		t.Fatal(got)
	}
}

func TestViewerSourceTrackingForcedManifestIsNotIngested(t *testing.T) {
	repo := newProjectRepo(t)
	path := trackingTestFile(t, repo, "_support/workspace/forced/walkthrough.json")
	gitIn(t, repo, "add", "-f", "--", path)
	state := t.TempDir()
	store, err := viewer.Open(viewer.Options{StateDir: state, SourceTracking: viewSourceTracking(runViewerTrackingGit)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := viewer.RegisterRepository(state, repo); err != nil {
		t.Fatal(err)
	}
	res, err := store.Sync(context.Background())
	if err != nil || res.Added != 0 || res.Errors != 0 || len(store.View().Projects) != 0 {
		t.Fatalf("tracked invalid manifest was read: result=%+v err=%v", res, err)
	}
}

func TestViewerSourceTrackingLargeRealIndex(t *testing.T) {
	repo := newProjectRepo(t)
	for i := 0; i < 1800; i++ {
		trackingTestFile(t, repo, fmt.Sprintf("_support/workspace/fixture-%06d/long-filename-for-tracking.png", i))
	}
	target := trackingTestFile(t, repo, "_support/workspace/zzzz/walkthrough.json")
	gitIn(t, repo, "add", "--", "_support/workspace")
	output, err := runViewerTrackingGit(context.Background(), repo, "ls-files", "--cached", "-z", "--", ":(top,literal)_support/workspace/")
	if err != nil {
		t.Fatal(err)
	}
	if len(output) <= 64<<10 || bytes.Index(output, []byte("_support/workspace/zzzz/walkthrough.json")) <= 64<<10 {
		t.Fatal("late tracked path did not exceed legacy output cap")
	}
	got := viewSourceTracking(runViewerTrackingGit)(context.Background(), []string{target})[target]
	if got.Status != "tracked" {
		t.Fatal(got)
	}
}
