package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// viewFixture is one isolated state directory plus recording dependencies.
// The server is the test binary itself (HIVE_VIEW_TEST_SERVE, see TestMain),
// started through the real launcher; every launched process group is killed
// when the test ends.
type viewFixture struct {
	t        *testing.T
	stateDir string
	launched []int
	opened   []string
}

func newViewFixture(t *testing.T) *viewFixture {
	t.Helper()
	t.Setenv("HIVE_VIEW_TEST_SERVE", "1")
	t.Setenv("HIVE_VIEW_POLL_INTERVAL", "1s")
	// The command canonicalizes the state directory (/var -> /private/var on
	// macOS), and server.json records that form.
	state, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &viewFixture{t: t, stateDir: state}
	t.Cleanup(func() {
		for _, pid := range f.launched {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
		// A server started by a nested command that the fixture did not see.
		if info, err := readServerInfo(filepath.Join(f.stateDir, "viewer", "server.json")); err == nil {
			_ = syscall.Kill(-info.PID, syscall.SIGKILL)
		}
	})
	return f
}

func (f *viewFixture) deps() viewDeps {
	return viewDeps{
		launch: func(args []string, logPath string) (int, error) {
			pid, err := defaultViewLauncher(args, logPath)
			if pid > 0 {
				f.launched = append(f.launched, pid)
			}
			return pid, err
		},
		open: func(url string) error {
			f.opened = append(f.opened, url)
			return nil
		},
	}
}

// view runs `hive view <args> --state-dir <dir>` and returns its output.
func (f *viewFixture) view(args ...string) (string, error) {
	f.t.Helper()
	var buf bytes.Buffer
	err := runView(append(args, "--state-dir", f.stateDir), &buf, f.deps())
	return buf.String(), err
}

func (f *viewFixture) mustView(args ...string) string {
	f.t.Helper()
	out, err := f.view(args...)
	if err != nil {
		f.t.Fatalf("view %v: %v\n%s", args, err, out)
	}
	return out
}

func (f *viewFixture) serverJSON() string {
	return filepath.Join(f.stateDir, "viewer", "server.json")
}

func (f *viewFixture) info() serverInfo {
	f.t.Helper()
	info, err := readServerInfo(f.serverJSON())
	if err != nil {
		f.t.Fatal(err)
	}
	return info
}

// get fetches a path from the running server with the Host header the
// server requires.
func (f *viewFixture) get(path string, token bool) (int, []byte) {
	f.t.Helper()
	info := f.info()
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d%s", info.Port, path), nil)
	if err != nil {
		f.t.Fatal(err)
	}
	if token {
		req.Header.Set("X-Hive-Token", info.Token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		f.t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body
}

// newViewRepo creates a Git working tree (a .git directory is enough for
// FindGitRoot) and returns its canonical path.
func newViewRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	return dir
}

func walkthroughFixture(name string) string {
	return filepath.Join("..", "..", "tests", "fixtures", "walkthrough", name)
}

func TestViewRequiresKnownSubcommand(t *testing.T) {
	var buf bytes.Buffer
	if err := runView(nil, &buf, viewDeps{}); err == nil {
		t.Fatal("view without a subcommand succeeded")
	}
	if err := run([]string{"view", "frobnicate"}); err == nil || !strings.Contains(err.Error(), "frobnicate") {
		t.Fatalf("unknown subcommand error = %v", err)
	}
}

func TestViewRegisterIsIdempotentAndRejectsNonGit(t *testing.T) {
	f := newViewFixture(t)
	repo := newViewRepo(t)
	out := f.mustView("register", "--dir", repo)
	if !strings.Contains(out, repo) {
		t.Fatalf("register output %q does not name the root", out)
	}
	reg := filepath.Join(f.stateDir, "viewer", "registry.json")
	old := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(reg, old, old); err != nil {
		t.Fatal(err)
	}
	// A subdirectory resolves to the same root.
	sub := filepath.Join(repo, "pkg")
	if err := os.Mkdir(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	out = f.mustView("register", "--dir", sub)
	if !strings.Contains(out, "Already registered") {
		t.Fatalf("second register output = %q", out)
	}
	info, err := os.Stat(reg)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(old) {
		t.Fatalf("second register rewrote registry.json (mtime %v)", info.ModTime())
	}
	r, err := viewer.ReadRegistry(f.stateDir)
	if err != nil || len(r.Repositories) != 1 || r.Repositories[0].Root != repo {
		t.Fatalf("registry = %+v, %v", r, err)
	}

	plain, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.view("register", "--dir", plain); err == nil || !strings.Contains(err.Error(), "Git") {
		t.Fatalf("register outside Git error = %v", err)
	}
}

func TestViewStartStatusStopLifecycle(t *testing.T) {
	f := newViewFixture(t)
	launches := 0
	deps := f.deps()
	counting := deps.launch
	deps.launch = func(args []string, logPath string) (int, error) {
		launches++
		return counting(args, logPath)
	}
	run := func(args ...string) (string, error) {
		var buf bytes.Buffer
		err := runView(append(args, "--state-dir", f.stateDir), &buf, deps)
		return buf.String(), err
	}

	out, err := run("start")
	if err != nil {
		t.Fatalf("start: %v\n%s", err, out)
	}
	info := f.info()
	if info.PID <= 0 || info.Port <= 0 || info.Version == "" || info.Token == "" {
		t.Fatalf("server.json incomplete: %+v", info)
	}
	if info.StateDir != f.stateDir {
		t.Fatalf("server.json state_dir = %q, want %q", info.StateDir, f.stateDir)
	}
	st, err := os.Stat(f.serverJSON())
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("server.json mode = %v, %v", st.Mode(), err)
	}
	if strings.Contains(out, info.Token) {
		t.Fatal("start printed the token")
	}

	out, err = run("status")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, info.Token) {
		t.Fatal("status printed the token")
	}
	var status struct {
		Running      bool   `json:"running"`
		URL          string `json:"url"`
		PID          int    `json:"pid"`
		Repositories int    `json:"repositories"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("status is not JSON: %v\n%s", err, out)
	}
	if !status.Running || status.PID != info.PID || status.URL != fmt.Sprintf("http://127.0.0.1:%d", info.Port) {
		t.Fatalf("status = %+v, server.json = %+v", status, info)
	}

	out, err = run("start")
	if err != nil || !strings.Contains(out, "already running") {
		t.Fatalf("second start = %q, %v", out, err)
	}
	if launches != 1 || f.info().PID != info.PID {
		t.Fatalf("second start launched again: launches=%d pid %d -> %d", launches, info.PID, f.info().PID)
	}

	// A direct serve on the same state directory must refuse.
	if err := runView([]string{"serve", "--state-dir", f.stateDir}, io.Discard, deps); err == nil {
		t.Fatal("a second serve on the same state directory succeeded")
	}
	if f.info().PID != info.PID {
		t.Fatal("the refused serve disturbed server.json")
	}

	if out, err = run("stop"); err != nil {
		t.Fatalf("stop: %v\n%s", err, out)
	}
	if _, err := os.Stat(f.serverJSON()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("server.json after stop: %v", err)
	}
	out, err = run("status")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil || status.Running {
		t.Fatalf("status after stop = %s (%v)", out, err)
	}
	// The lock is released a moment before the process is gone.
	gone := false
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		if syscall.Kill(info.PID, 0) == syscall.ESRCH {
			gone = true
			break
		}
	}
	if !gone {
		t.Fatal("server process still alive after stop")
	}
	if out, err = run("stop"); err != nil || !strings.Contains(out, "not running") {
		t.Fatalf("stop when stopped = %q, %v", out, err)
	}

	// An orphan server.json with a free lock does not block a new start.
	if err := os.WriteFile(f.serverJSON(), []byte(`{"pid":1,"port":1,"token":"stale-stale-stale-stale"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err = run("start"); err != nil {
		t.Fatalf("start over an orphan server.json: %v\n%s", err, out)
	}
	if f.info().PID == 1 || f.info().PID == info.PID {
		t.Fatalf("orphan server.json was not replaced: %+v", f.info())
	}
	if _, err = run("stop"); err != nil {
		t.Fatal(err)
	}
}

func TestViewStartTTLReachesHealthAndStore(t *testing.T) {
	f := newViewFixture(t)
	f.mustView("start", "--ttl", "1h")
	code, body := f.get("/api/health", true)
	if code != http.StatusOK {
		t.Fatalf("health = %d", code)
	}
	var h map[string]any
	if err := json.Unmarshal(body, &h); err != nil {
		t.Fatal(err)
	}
	if h["ttl"] != "1h0m0s" {
		t.Fatalf("health ttl = %v", h["ttl"])
	}
	if f.info().TTL != "1h0m0s" {
		t.Fatalf("server.json ttl = %q", f.info().TTL)
	}
	_, body = f.get("/api/index", true)
	var v struct {
		TTLSeconds int64 `json:"ttl_seconds"`
	}
	if err := json.Unmarshal(body, &v); err != nil || v.TTLSeconds != 3600 {
		t.Fatalf("index ttl_seconds = %d (%v)", v.TTLSeconds, err)
	}
	if _, err := f.view("start", "--ttl", "soon"); err == nil {
		t.Fatal("an invalid --ttl was accepted")
	}
	f.mustView("stop")
}

func TestViewServesRegisteredRepositoryEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	f := newViewFixture(t)
	repo := newViewRepo(t)
	if err := os.RemoveAll(filepath.Join(repo, ".git")); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", "https://github.com/acme/demo.git"}} {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	agents := "# Repo\n\n## Hive\n\n- Project: Demo Project\n- Base branch: main\n"
	if err := os.WriteFile(filepath.Join(repo, "AGENTS.md"), []byte(agents), 0o600); err != nil {
		t.Fatal(err)
	}
	ws := filepath.Join(repo, "_support", "workspace")
	good := filepath.Join(ws, "2026-10-18-x")
	bad := filepath.Join(ws, "2026-10-18-bad")
	for dst, src := range map[string]string{good: "valid/final", bad: "invalid/missing-required"} {
		if err := os.MkdirAll(dst, 0o700); err != nil {
			t.Fatal(err)
		}
		copyTree(t, walkthroughFixture(src), dst)
	}
	f.mustView("register", "--dir", repo)
	f.mustView("start")

	type indexView struct {
		Projects []struct {
			Key, Name string
			Tickets   []struct {
				Sessions []struct {
					Source string `json:"source"`
				}
			}
		}
		Errors []struct{ Path, Message string }
	}
	var v indexView
	deadline := time.Now().Add(6 * time.Second) // two 1 s intervals plus slack
	for {
		_, body := f.get("/api/index", true)
		v = indexView{}
		if err := json.Unmarshal(body, &v); err != nil {
			t.Fatal(err)
		}
		if len(v.Projects) > 0 && len(v.Errors) > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("index never showed the walkthrough and the error: %s", body)
		}
		time.Sleep(100 * time.Millisecond)
	}
	p := v.Projects[0]
	if p.Name != "Demo Project" || p.Key != "github.com/acme/demo" {
		t.Fatalf("grouping = key %q name %q, want the Hive project and the normalized remote", p.Key, p.Name)
	}
	found := false
	for _, tk := range p.Tickets {
		for _, s := range tk.Sessions {
			found = found || s.Source == good
		}
	}
	if !found {
		t.Fatalf("valid/final (%s) is not listed: %+v", good, p)
	}
	if code, _ := f.get("/api/health", true); code != http.StatusOK {
		t.Fatalf("server stopped answering after an invalid fixture: %d", code)
	}
	f.mustView("stop")
}

func TestViewOpenNeedsARunningServer(t *testing.T) {
	f := newViewFixture(t)
	out := f.mustView("open")
	if len(f.opened) != 0 || len(f.launched) != 0 {
		t.Fatalf("open without a server opened %v / launched %v", f.opened, f.launched)
	}
	if !strings.Contains(out, "not running") {
		t.Fatalf("open output = %q", out)
	}
	f.mustView("start")
	f.mustView("open")
	info := f.info()
	want := fmt.Sprintf("http://127.0.0.1:%d/#token=%s", info.Port, info.Token)
	if len(f.opened) != 1 || f.opened[0] != want {
		t.Fatalf("opener received %v, want %q", f.opened, want)
	}
	f.mustView("stop")
}

func TestViewStopNeverSignalsWithoutAMatchingHealth(t *testing.T) {
	f := newViewFixture(t)
	vdir := filepath.Join(f.stateDir, "viewer")
	if err := os.MkdirAll(vdir, 0o700); err != nil {
		t.Fatal(err)
	}
	lock, err := viewer.AcquireLock(filepath.Join(vdir, "server.lock"), false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	bystander := exec.Command("sleep", "60")
	if err := bystander.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = bystander.Process.Kill(); _ = bystander.Wait() })
	data, _ := json.Marshal(serverInfo{PID: bystander.Process.Pid, Port: 1, Token: "token-token-token-token", StateDir: f.stateDir})
	if err := os.WriteFile(f.serverJSON(), data, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = f.view("stop")
	if err == nil || !strings.Contains(err.Error(), fmt.Sprint(bystander.Process.Pid)) {
		t.Fatalf("stop error = %v, want one naming the pid", err)
	}
	if syscall.Kill(bystander.Process.Pid, 0) != nil {
		t.Fatal("stop signalled a process whose health did not answer")
	}
	if _, err := os.Stat(f.serverJSON()); err != nil {
		t.Fatalf("stop removed server.json: %v", err)
	}
}

func TestViewHiveProjectReader(t *testing.T) {
	root := t.TempDir()
	if _, ok := viewHiveProject(root); ok {
		t.Fatal("a repository without AGENTS.md has a project")
	}
	write := func(s string) {
		if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(s), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("# x\n\n## Hive\n\n- Base branch: main\n")
	if _, ok := viewHiveProject(root); ok {
		t.Fatal("a section without Project has a project")
	}
	write("# x\n\n## Hive\n\n- Project: demo\n")
	if name, ok := viewHiveProject(root); !ok || name != "demo" {
		t.Fatalf("project = %q, %v", name, ok)
	}
}

func TestViewGitRemoteOnlyAnswersForAGitToplevel(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	outer, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", "https://github.com/acme/outer.git"}} {
		if out, err := exec.Command("git", append([]string{"-C", outer}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	workspace := filepath.Join(outer, "ws") // not a repository, inside one
	if err := os.MkdirAll(workspace, 0o700); err != nil {
		t.Fatal(err)
	}
	elsewhere, _ := filepath.EvalSymlinks(t.TempDir()) // not inside any repository
	read := viewGitRemote(newGitRunner())
	if got, err := read(outer); err != nil || got != "https://github.com/acme/outer.git" {
		t.Fatalf("toplevel remote = %q, %v", got, err)
	}
	if got, err := read(workspace); err == nil {
		t.Fatalf("a root inside another repository borrowed its remote: %q", got)
	}
	if got, err := read(elsewhere); err == nil {
		t.Fatalf("a root outside Git has a remote: %q", got)
	}
}

func TestViewHiveProjectReadsAWorkspaceAgentsFileWithoutGit(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("## Hive\n\n- Project: cam\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if name, ok := viewHiveProject(root); !ok || name != "cam" {
		t.Fatalf("Project = %q, %v", name, ok)
	}
}

func TestViewRegisterAcceptsAWorkspaceRootWithoutGit(t *testing.T) {
	f := newViewFixture(t)
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	work := filepath.Join(root, "_support", "workspace", "2026-10-19-x")
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatal(err)
	}
	if out := f.mustView("register", "--dir", work); !strings.Contains(out, root) {
		t.Fatalf("register output %q does not name the workspace root %s", out, root)
	}
	if out := f.mustView("register", "--dir", work); !strings.Contains(out, "Already registered") {
		t.Fatalf("second register output = %q", out)
	}
	if _, err := f.view("register", "--dir", t.TempDir()); err == nil || !strings.Contains(err.Error(), "_support/workspace") {
		t.Fatalf("a folder outside both must fail naming both cases, got %v", err)
	}
}
