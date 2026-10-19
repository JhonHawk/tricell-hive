package viewer

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// canon resolves symlinks the way the package does (macOS temp directories
// live behind /var -> /private/var).
func canon(t *testing.T, path string) string {
	t.Helper()
	p, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// newRepo creates a directory that looks like a Git repository root.
func newRepo(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return canon(t, root)
}

func readRegistryFile(t *testing.T, state string) Registry {
	t.Helper()
	r, err := ReadRegistry(state)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestRegistryAddIsIdempotentAndWritesNothingTheSecondTime(t *testing.T) {
	state, repo := t.TempDir(), newRepo(t)
	res, err := RegisterRepository(state, repo)
	if err != nil || !res.Added || res.Root != repo {
		t.Fatalf("first add = %+v, %v", res, err)
	}
	file := filepath.Join(state, "viewer", "registry.json")
	old := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(file, old, old); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(file)

	res, err = RegisterRepository(state, repo)
	if err != nil || res.Added || res.Root != repo {
		t.Fatalf("second add = %+v, %v", res, err)
	}
	info, _ := os.Stat(file)
	if !info.ModTime().Equal(old) {
		t.Fatalf("mtime changed on an idempotent add: %v != %v", info.ModTime(), old)
	}
	after, _ := os.ReadFile(file)
	if string(before) != string(after) {
		t.Fatal("registry content changed on an idempotent add")
	}
	if got := readRegistryFile(t, state); len(got.Repositories) != 1 || got.Version != 1 {
		t.Fatalf("registry = %+v", got)
	}
}

func TestRegistryResolvesTheGitRootOfASubdirectory(t *testing.T) {
	state, repo := t.TempDir(), newRepo(t)
	sub := filepath.Join(repo, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := RegisterRepository(state, sub)
	if err != nil || res.Root != repo {
		t.Fatalf("root = %+v, %v; want %s", res, err, repo)
	}
}

func TestRegistryAcceptsAWorktreeGitFile(t *testing.T) {
	state := t.TempDir()
	wt := canon(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(wt, ".git"), []byte("gitdir: /elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := RegisterRepository(state, wt)
	if err != nil || res.Root != wt || !res.Added {
		t.Fatalf("worktree = %+v, %v", res, err)
	}
}

func TestRegistryRejectsADirectoryOutsideGit(t *testing.T) {
	state, dir := t.TempDir(), t.TempDir()
	_, err := RegisterRepository(state, dir)
	if !errors.Is(err, ErrNotGitRepository) {
		t.Fatalf("err = %v, want ErrNotGitRepository", err)
	}
	if _, statErr := os.Stat(filepath.Join(state, "viewer", "registry.json")); statErr == nil {
		t.Fatal("registry written for a rejected directory")
	}
	if _, err := RegisterRepository(state, filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing directory accepted")
	}
}

func TestRegistryFilesAreOwnerOnly(t *testing.T) {
	state := t.TempDir()
	if _, err := RegisterRepository(state, newRepo(t)); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]os.FileMode{
		filepath.Join(state, "viewer"):                  0o700,
		filepath.Join(state, "viewer", "registry.json"): 0o600,
	} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != want {
			t.Fatalf("%s mode = %v, %v; want %v", path, info.Mode().Perm(), err, want)
		}
	}
}

func TestRegistryCorruptFileIsReportedAndLeftAlone(t *testing.T) {
	state := t.TempDir()
	dir := filepath.Join(state, "viewer")
	os.MkdirAll(dir, 0o700)
	file := filepath.Join(dir, "registry.json")
	os.WriteFile(file, []byte("{broken"), 0o600)
	if _, err := RegisterRepository(state, newRepo(t)); err == nil {
		t.Fatal("corrupt registry accepted")
	}
	if data, _ := os.ReadFile(file); string(data) != "{broken" {
		t.Fatalf("corrupt registry overwritten: %q", data)
	}
	os.WriteFile(file, []byte(`{"version":2,"repositories":[]}`), 0o600)
	if _, err := ReadRegistry(state); err == nil {
		t.Fatal("unknown registry version accepted")
	}
}

func TestRegistryMissingFileReadsAsEmpty(t *testing.T) {
	r, err := ReadRegistry(t.TempDir())
	if err != nil || len(r.Repositories) != 0 {
		t.Fatalf("empty read = %+v, %v", r, err)
	}
}

func TestRegistryWaitsForTheCrossProcessLock(t *testing.T) {
	state, repo := t.TempDir(), newRepo(t)
	if err := os.MkdirAll(filepath.Join(state, "viewer"), 0o700); err != nil {
		t.Fatal(err)
	}
	lock, err := AcquireLock(filepath.Join(state, "viewer", "lock"), true)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := RegisterRepository(state, repo); done <- err }()
	select {
	case err := <-done:
		t.Fatalf("register finished while the lock was held: %v", err)
	case <-time.After(150 * time.Millisecond):
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("register did not resume after the lock was released")
	}
}

func TestRegistryNonBlockingLockReportsErrLocked(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.lock")
	first, err := AcquireLock(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcquireLock(path, false); !errors.Is(err, ErrLocked) {
		t.Fatalf("second lock err = %v, want ErrLocked", err)
	}
	first.Release()
	again, err := AcquireLock(path, false)
	if err != nil {
		t.Fatalf("lock not reusable after release: %v", err)
	}
	again.Release()
}

// TestRegistryHelperProcess is not a test: the parent re-runs the test
// binary with HIVE_VIEWER_HELPER set so several real processes register at
// once.
func TestRegistryHelperProcess(t *testing.T) {
	args := os.Getenv("HIVE_VIEWER_HELPER")
	if args == "" {
		t.Skip("helper process")
	}
	state, repo, _ := strings.Cut(args, "|")
	if _, err := RegisterRepository(state, repo); err != nil {
		t.Fatal(err)
	}
}

func TestRegistryConcurrentProcessesLoseNoEntry(t *testing.T) {
	state := t.TempDir()
	const n = 8
	var cmds []*exec.Cmd
	var want []string
	for i := 0; i < n; i++ {
		repo := newRepo(t)
		want = append(want, repo)
		cmd := exec.Command(os.Args[0], "-test.run=^TestRegistryHelperProcess$")
		cmd.Env = append(os.Environ(), "HIVE_VIEWER_HELPER="+state+"|"+repo)
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		cmds = append(cmds, cmd)
	}
	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Fatalf("helper failed: %v", err)
		}
	}
	got := map[string]bool{}
	for _, r := range readRegistryFile(t, state).Repositories {
		got[r.Root] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Fatalf("entry %s lost; registry has %v", w, got)
		}
	}
}

// newWorkspace creates a multi-repository workspace root that is not a Git
// repository: an AGENTS.md with a `## Hive` section and one work folder.
func newWorkspace(t *testing.T) (root, work string) {
	t.Helper()
	root = canon(t, t.TempDir())
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("## Hive\n\n- Project: demo-ws\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	work = filepath.Join(root, "_support", "workspace", "2026-10-19-x")
	copyTree(t, filepath.Join(fixturesRoot, "valid", "final"), work)
	return root, work
}

func TestRegistryWorkspaceRootNeedsNoGitAndIsIdempotent(t *testing.T) {
	state := t.TempDir()
	root, work := newWorkspace(t)
	res, err := RegisterRepository(state, work)
	if err != nil || !res.Added || res.Root != root {
		t.Fatalf("register from the work folder = %+v, %v; want root %s", res, err, root)
	}
	file := filepath.Join(state, "viewer", "registry.json")
	old := time.Now().Add(-48 * time.Hour).Truncate(time.Second)
	os.Chtimes(file, old, old)
	// Any depth inside _support/workspace resolves to the same root.
	for _, dir := range []string{work, filepath.Join(work, "images"), filepath.Join(root, "_support", "workspace")} {
		res, err := RegisterRepository(state, dir)
		if err != nil || res.Added || res.Root != root {
			t.Fatalf("register from %s = %+v, %v", dir, res, err)
		}
	}
	if info, _ := os.Stat(file); !info.ModTime().Equal(old) {
		t.Fatal("registry rewritten by an idempotent register")
	}
}

func TestRegistryWorkspaceRuleWinsOverTheEnclosingGitRepository(t *testing.T) {
	state := t.TempDir()
	// A workspace root nested inside an unrelated Git repository.
	outer := newRepo(t)
	root := filepath.Join(outer, "ws")
	work := filepath.Join(root, "_support", "workspace", "w")
	os.MkdirAll(work, 0o755)
	res, err := RegisterRepository(state, work)
	if err != nil || res.Root != root {
		t.Fatalf("root = %+v, %v; want the workspace root %s, not the Git toplevel", res, err, root)
	}
}

func TestRegistryNestedGitRepositoryInAWorkspaceKeepsItsOwnToplevel(t *testing.T) {
	state := t.TempDir()
	root, _ := newWorkspace(t)
	nested := filepath.Join(root, "cam-monorepo")
	os.MkdirAll(filepath.Join(nested, ".git"), 0o755)
	sub := filepath.Join(nested, "src")
	os.MkdirAll(sub, 0o755)
	res, err := RegisterRepository(state, sub)
	if err != nil || res.Root != nested {
		t.Fatalf("root = %+v, %v; want %s", res, err, nested)
	}
}

func TestRegistryWorkspaceRuleResolvesLinksFirst(t *testing.T) {
	state := t.TempDir()
	root, work := newWorkspace(t)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(work, link); err != nil {
		t.Fatal(err)
	}
	res, err := RegisterRepository(state, link)
	if err != nil || res.Root != root {
		t.Fatalf("root via a link = %+v, %v; want %s", res, err, root)
	}
}

func TestRegistryOutsideGitAndWorkspaceNamesBothCases(t *testing.T) {
	_, err := RegisterRepository(t.TempDir(), t.TempDir())
	if !errors.Is(err, ErrNotGitRepository) {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "Git repository") || !strings.Contains(err.Error(), "_support/workspace") {
		t.Fatalf("message must mention both cases: %v", err)
	}
}
