//go:build unix

package distribution

// This file runs get-hive.sh through real shell processes (/bin/sh and dash)
// against an httptest server on loopback. The shipped script reads its origin,
// protocol restriction, allowed hosts, and terminal device from fixed single
// lines. Like the retired bootstrap.sh tests, these tests run a test-only copy
// of the script in which exactly those lines are replaced; TestGetHiveShippedLines
// asserts the exact shipped values so the substitution can never hide a wrong one.
// No test reaches a real network origin or touches the real HOME.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

const (
	ghVersion      = "9.9.9"
	ghMaxArchive   = 64 << 20
	ghTTYAnswer    = "answer-from-the-terminal"
	ghScriptName   = "get-hive.sh"
	ghInstallStub  = "#!/bin/sh\nrec=${HIVE_TEST_RECORD:?}\n{\n  printf 'args:'\n  for a in \"$@\"; do printf ' [%s]' \"$a\"; done\n  printf '\\n'\n  if IFS= read -r line; then printf 'stdin:%s\\n' \"$line\"; else printf 'stdin:EOF\\n'; fi\n  printf 'umask:%s\\n' \"$(umask)\"\n  printf 'tmp:%s\\n' \"$(ls -A \"$TMPDIR\" | wc -l | tr -d ' ')\"\n  printf 'packages:'\n  for e in $(ls -A \"${0%/*}/..\"); do printf ' %s' \"$e\"; done\n  printf '\\n'\n} > \"$rec\"\nexit \"${HIVE_TEST_EXIT:-0}\"\n"
	ghShippedRepo  = "repo_url='https://github.com/JhonHawk/tricell-hive'"
	ghShippedProto = "proto_restriction='=https'"
	ghShippedHosts = "allowed_hosts='github.com .githubusercontent.com'"
	ghShippedTTY   = "tty_device=/dev/tty"
	ghFinalLine    = `{ main "$@" || exit 1; }`
)

// ---- server ----

type ghServer struct {
	*httptest.Server
	mu          sync.Mutex
	requests    []string
	archiveName string
	archive     []byte
	sum         string
	latest      http.HandlerFunc
	archiveH    http.HandlerFunc
	sumH        http.HandlerFunc
}

func (s *ghServer) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r.Method+" "+r.URL.Path)
	archive, sum, latest, archiveH, sumH := s.archive, s.sum, s.latest, s.archiveH, s.sumH
	s.mu.Unlock()
	p := r.URL.Path
	isDownload := strings.HasPrefix(p, "/releases/download/")
	switch {
	case p == "/releases/latest":
		if latest != nil {
			latest(w, r)
			return
		}
		http.Redirect(w, r, "/releases/tag/v"+ghVersion, http.StatusFound)
	case strings.Contains(p, "/releases/tag/"):
		w.WriteHeader(http.StatusOK)
	case strings.HasSuffix(p, ".tar.gz.sha256"):
		if isDownload && sumH != nil {
			sumH(w, r)
			return
		}
		_, _ = io.WriteString(w, sum)
	case strings.HasSuffix(p, ".tar.gz"):
		if isDownload && archiveH != nil {
			archiveH(w, r)
			return
		}
		_, _ = w.Write(archive)
	default:
		http.NotFound(w, r)
	}
}

func (s *ghServer) reqs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]string(nil), s.requests...)
	sort.Strings(out)
	return out
}

func (s *ghServer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.requests)
}

func (s *ghServer) port() string {
	return s.URL[strings.LastIndex(s.URL, ":")+1:]
}

func ghStreamZeros(w http.ResponseWriter, n int64, declared bool) {
	if declared {
		w.Header().Set("Content-Length", strconv.FormatInt(n, 10))
	}
	chunk := make([]byte, 1<<20)
	for n > 0 {
		size := int64(len(chunk))
		if n < size {
			size = n
		}
		if _, err := w.Write(chunk[:size]); err != nil {
			return
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		n -= size
	}
}

// ---- archive fixtures ----

type ghEntry struct {
	name string
	typ  byte
	body string
	link string
	mode int64
}

func ghGoodEntries(label string) []ghEntry {
	return []ghEntry{
		{name: label + "/", typ: tar.TypeDir, mode: 0o755},
		{name: label + "/install.sh", typ: tar.TypeReg, body: ghInstallStub, mode: 0o755},
		{name: label + "/bin/", typ: tar.TypeDir, mode: 0o755},
		{name: label + "/bin/hive", typ: tar.TypeReg, body: "binary\n", mode: 0o755},
		{name: label + "/VERSION", typ: tar.TypeReg, body: ghVersion + "\n", mode: 0o644},
	}
}

func ghTar(t *testing.T, entries []ghEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := tar.Header{Name: e.name, Typeflag: e.typ, Mode: e.mode, Linkname: e.link, ModTime: time.Unix(0, 0)}
		if e.typ == tar.TypeReg {
			h.Size = int64(len(e.body))
		}
		if e.typ == tar.TypeChar {
			h.Devmajor, h.Devminor = 1, 3
		}
		if err := tw.WriteHeader(&h); err != nil {
			t.Fatal(err)
		}
		if e.typ == tar.TypeReg {
			if _, err := io.WriteString(tw, e.body); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func ghSum(archive []byte, name string) string {
	return fmt.Sprintf("%x  %s\n", sha256.Sum256(archive), name)
}

// ---- environment ----

type ghEnv struct {
	t      *testing.T
	root   string
	home   string
	tmp    string
	work   string
	out    string
	bin    string
	tty    string
	record string
	label  string
	osName string
	arch   string
	srv    *ghServer
}

func newGhEnv(t *testing.T) *ghEnv {
	t.Helper()
	osName, arch := scriptPlatformParts(t)
	root := t.TempDir()
	e := &ghEnv{t: t, root: root, osName: osName, arch: arch}
	e.home = filepath.Join(root, "home")
	e.tmp = filepath.Join(root, "tmp")
	e.work = filepath.Join(root, "work")
	e.out = filepath.Join(root, "out")
	e.bin = filepath.Join(root, "bin")
	for _, d := range []string{e.home, e.tmp, e.work, e.out, e.bin} {
		if err := os.MkdirAll(d, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	e.tty = filepath.Join(root, "tty-device")
	mustWriteFile(t, e.tty, []byte(ghTTYAnswer+"\nsecond line\n"), 0o600)
	e.record = filepath.Join(e.out, "record")
	e.label = fmt.Sprintf("hive-%s-%s-%s", ghVersion, osName, arch)
	srv := &ghServer{archiveName: e.label + ".tar.gz"}
	srv.Server = httptest.NewServer(http.HandlerFunc(srv.handle))
	t.Cleanup(srv.Close)
	e.srv = srv
	e.setArchive(ghGoodEntries(e.label))
	return e
}

func (e *ghEnv) setArchive(entries []ghEntry) {
	e.t.Helper()
	e.srv.mu.Lock()
	defer e.srv.mu.Unlock()
	e.srv.archive = ghTar(e.t, entries)
	e.srv.sum = ghSum(e.srv.archive, e.srv.archiveName)
}

func (e *ghEnv) setSum(sum string) {
	e.srv.mu.Lock()
	defer e.srv.mu.Unlock()
	e.srv.sum = sum
}

func (e *ghEnv) setHandlers(latest, archive, sum http.HandlerFunc) {
	e.srv.mu.Lock()
	defer e.srv.mu.Unlock()
	e.srv.latest, e.srv.archiveH, e.srv.sumH = latest, archive, sum
}

func (e *ghEnv) packagesDir() string {
	return filepath.Join(e.home, ".local", "share", "hive", "packages")
}

func (e *ghEnv) missingTTY() string {
	return filepath.Join(e.root, "no-such-dir", "tty")
}

// script returns a test-only copy of get-hive.sh with the four fixed lines
// replaced. Each must exist exactly once in the shipped script.
func (e *ghEnv) script(tty string) []byte {
	e.t.Helper()
	return ghSubstitute(e.t, readRepoFile(e.t, ghScriptName), map[string]string{
		ghShippedRepo:  "repo_url='" + e.srv.URL + "'",
		ghShippedProto: "proto_restriction='=http'",
		ghShippedHosts: "allowed_hosts='127.0.0.1'",
		ghShippedTTY:   "tty_device='" + tty + "'",
	})
}

func ghSubstitute(t *testing.T, script []byte, replacements map[string]string) []byte {
	t.Helper()
	lines := strings.Split(string(script), "\n")
	for shipped, replacement := range replacements {
		found := 0
		for i, line := range lines {
			if line == shipped {
				lines[i] = replacement
				found++
			}
		}
		if found != 1 {
			t.Fatalf("shipped line %q appears %d times, want exactly 1", shipped, found)
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

func ghRealTool(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not available: %v", name, err)
	}
	return path
}

func (e *ghEnv) fakeBin(name, body string) {
	e.t.Helper()
	mustWriteFile(e.t, filepath.Join(e.bin, name), []byte("#!/bin/sh\n"+body), 0o755)
}

type ghRun struct {
	shell        string
	script       []byte
	args         []string
	pipe         bool
	stdin        string
	env          []string
	pathOnly     string
	signal       syscall.Signal
	signalMarker string
}

type ghResult struct {
	code   int
	stdout string
	stderr string
}

func (e *ghEnv) run(r ghRun) ghResult {
	e.t.Helper()
	scriptPath := filepath.Join(e.out, ghScriptName)
	mustWriteFile(e.t, scriptPath, r.script, 0o644)
	var argv []string
	argv = append(argv, "-c", `umask 027; exec "$@"`, "sh", r.shell)
	if r.pipe {
		argv = append(argv, "-s", "--")
	} else {
		argv = append(argv, scriptPath)
	}
	argv = append(argv, r.args...)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/sh", argv...)
	cmd.Dir = e.work
	path := e.bin + ":" + os.Getenv("PATH")
	if r.pathOnly != "" {
		path = r.pathOnly
	}
	cmd.Env = append([]string{
		"HOME=" + e.home,
		"TMPDIR=" + e.tmp,
		"PATH=" + path,
		"LC_ALL=C",
		"HIVE_TEST_RECORD=" + e.record,
	}, r.env...)
	if r.pipe {
		cmd.Stdin = bytes.NewReader(r.script)
	} else {
		cmd.Stdin = strings.NewReader(r.stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Start(); err != nil {
		e.t.Fatal(err)
	}
	if r.signal != 0 {
		deadline := time.Now().Add(20 * time.Second)
		for {
			if _, err := os.Stat(r.signalMarker); err == nil {
				break
			}
			if time.Now().After(deadline) {
				_ = cmd.Process.Kill()
				_ = cmd.Wait()
				e.t.Fatalf("marker %s never appeared; stderr:\n%s", r.signalMarker, stderr.String())
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err := cmd.Process.Signal(r.signal); err != nil {
			e.t.Fatal(err)
		}
	}
	err := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		e.t.Fatalf("get-hive.sh hung; stderr:\n%s", stderr.String())
	}
	return ghResult{code: exitCodeOf(e.t, err), stdout: stdout.String(), stderr: stderr.String()}
}

func ghShellPath(name string) string {
	if name == "sh" {
		return "/bin/sh"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

func ghEachShell(t *testing.T, fn func(t *testing.T, shell string)) {
	t.Helper()
	for _, name := range []string{"sh", "dash"} {
		path := ghShellPath(name)
		t.Run(name, func(t *testing.T) {
			if path == "" {
				t.Skipf("%s is not installed", name)
			}
			t.Parallel()
			fn(t, path)
		})
	}
}

// ---- assertions ----

func ghSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	if _, err := os.Lstat(root); err != nil {
		return out
	}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, _ := filepath.Rel(root, p)
		info, err := os.Lstat(p)
		if err != nil {
			return err
		}
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			target, _ := os.Readlink(p)
			out[rel] = "l:" + target
		case info.IsDir():
			out[rel] = fmt.Sprintf("d:%o", info.Mode().Perm())
		default:
			data, _ := os.ReadFile(p)
			out[rel] = fmt.Sprintf("f:%o:%x", info.Mode().Perm(), sha256.Sum256(data))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func ghEqualTrees(t *testing.T, got, want map[string]string, what string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s changed:\n got %v\nwant %v", what, got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s changed at %q:\n got %v\nwant %v", what, k, got, want)
		}
	}
}

func ghNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names
}

func (e *ghEnv) recordText() (string, bool) {
	data, err := os.ReadFile(e.record)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func (e *ghEnv) recordLine(prefix string) string {
	e.t.Helper()
	text, ok := e.recordText()
	if !ok {
		e.t.Fatal("install.sh did not run (no record)")
	}
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	e.t.Fatalf("record has no %q line:\n%s", prefix, text)
	return ""
}

// seedPrevious creates an existing package folder for the same label and a
// sibling folder of another version, and returns a snapshot of the packages dir.
func (e *ghEnv) seedPrevious() map[string]string {
	e.t.Helper()
	dir := e.packagesDir()
	mustWriteFile(e.t, filepath.Join(dir, e.label, "install.sh"), []byte("#!/bin/sh\nexit 99\n"), 0o755)
	mustWriteFile(e.t, filepath.Join(dir, e.label, "marker"), []byte("previous package\n"), 0o644)
	mustWriteFile(e.t, filepath.Join(dir, e.label, "bin", "hive"), []byte("old binary\n"), 0o755)
	mustWriteFile(e.t, filepath.Join(dir, "hive-1.0.0-"+e.osName+"-"+e.arch, "keep"), []byte("other version\n"), 0o644)
	return ghSnapshot(e.t, dir)
}

// assertRejected checks the AC4 outcome: a failure, install.sh never ran, the
// packages folder is exactly as before, and no temporary files remain.
func (e *ghEnv) assertRejected(res ghResult, before map[string]string) {
	e.t.Helper()
	if res.code == 0 {
		e.t.Fatalf("exit 0, want a failure; stdout:\n%s\nstderr:\n%s", res.stdout, res.stderr)
	}
	if strings.TrimSpace(res.stderr) == "" {
		e.t.Fatal("failed without an error message on stderr")
	}
	if _, ran := e.recordText(); ran {
		e.t.Fatal("install.sh ran after a rejected download")
	}
	ghEqualTrees(e.t, ghSnapshot(e.t, e.packagesDir()), before, "packages folder")
	if names := ghNames(e.t, e.tmp); len(names) != 0 {
		e.t.Fatalf("temporary files left in TMPDIR: %v", names)
	}
}

func (e *ghEnv) assertNothingHappened(res ghResult) {
	e.t.Helper()
	if res.code == 0 {
		e.t.Fatalf("exit 0, want a failure; stderr:\n%s", res.stderr)
	}
	if n := e.srv.count(); n != 0 {
		e.t.Fatalf("server received %d requests, want 0: %v", n, e.srv.reqs())
	}
	if names := ghNames(e.t, e.home); len(names) != 0 {
		e.t.Fatalf("HOME was written: %v", names)
	}
	if names := ghNames(e.t, e.tmp); len(names) != 0 {
		e.t.Fatalf("TMPDIR was written: %v", names)
	}
	if _, ran := e.recordText(); ran {
		e.t.Fatal("install.sh ran")
	}
}

// ---- tests ----

func TestGetHiveShippedLines(t *testing.T) {
	data := readRepoFile(t, ghScriptName)
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for _, c := range []struct{ prefix, line string }{
		{"repo_url=", ghShippedRepo},
		{"proto_restriction=", ghShippedProto},
		{"allowed_hosts=", ghShippedHosts},
		{"tty_device=", ghShippedTTY},
	} {
		count := 0
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), c.prefix) {
				count++
				if line != c.line {
					t.Errorf("line %q, want exactly %q", line, c.line)
				}
			}
		}
		if count != 1 {
			t.Errorf("%s is defined %d times, want 1", c.prefix, count)
		}
	}
	if lines[0] != "#!/bin/sh" {
		t.Errorf("first line %q, want #!/bin/sh", lines[0])
	}
	if last := lines[len(lines)-1]; last != ghFinalLine {
		t.Errorf("last line %q, want %q", last, ghFinalLine)
	}
	if !strings.HasSuffix(string(data), ghFinalLine+"\n") {
		t.Error("the script must end with the final line and a newline")
	}
	info, err := os.Stat(filepath.Join("..", "..", ghScriptName))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("mode %v, want 0755", info.Mode().Perm())
	}
	// The only top-level command is the final call: everything before it is an
	// assignment, a comment, or a function body.
	depth := 0
	for _, line := range lines[:len(lines)-1] {
		trimmed := strings.TrimSpace(line)
		switch {
		case depth == 0 && (trimmed == "" || strings.HasPrefix(trimmed, "#")):
		case depth == 0 && strings.HasSuffix(trimmed, "() {"):
			depth = 1
		case depth == 0 && strings.Contains(trimmed, "="):
		case depth == 1 && line == "}":
			depth = 0
		case depth == 1:
		default:
			t.Errorf("unexpected top-level line %q", line)
		}
	}
	if depth != 0 {
		t.Error("a function is not closed before the final line")
	}
}

func ghExpectedRequests(e *ghEnv, withLatest bool) []string {
	reqs := []string{
		"GET /releases/download/v" + ghVersion + "/" + e.srv.archiveName,
		"GET /releases/download/v" + ghVersion + "/" + e.srv.archiveName + ".sha256",
	}
	if withLatest {
		reqs = append(reqs, "HEAD /releases/latest", "HEAD /releases/tag/v"+ghVersion)
	}
	sort.Strings(reqs)
	return reqs
}

func ghAssertRequests(t *testing.T, e *ghEnv, want []string) {
	t.Helper()
	got := e.srv.reqs()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("server requests:\n got %v\nwant %v", got, want)
	}
}

func TestGetHiveFullPath(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		for _, pipe := range []bool{false, true} {
			name := "file"
			if pipe {
				name = "stdin-pipe"
			}
			t.Run(name, func(t *testing.T) {
				e := newGhEnv(t)
				// A previous download folder in TMPDIR must not matter, but this run's must go.
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--hosts", "claude,codex", "--home", "/x y"}, pipe: pipe})
				if res.code != 0 {
					t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", res.code, res.stdout, res.stderr)
				}
				if got, want := e.recordLine("args:"), "args: [--hosts] [claude,codex] [--home] [/x y]"; got != want {
					t.Errorf("%q, want %q", got, want)
				}
				if got, want := e.recordLine("stdin:"), "stdin:"+ghTTYAnswer; got != want {
					t.Errorf("install.sh stdin %q, want the terminal device (%q)", got, want)
				}
				umask, err := strconv.ParseInt(strings.TrimPrefix(e.recordLine("umask:"), "umask:"), 8, 32)
				if err != nil || umask != 0o027 {
					t.Errorf("install.sh umask %q, want the caller's 027", e.recordLine("umask:"))
				}
				if got := e.recordLine("tmp:"); got != "tmp:0" {
					t.Errorf("downloads were not removed before install.sh: %s", got)
				}
				if got, want := e.recordLine("packages:"), "packages: "+e.label; got != want {
					t.Errorf("%q, want %q", got, want)
				}
				if names := ghNames(t, e.packagesDir()); strings.Join(names, ",") != e.label {
					t.Errorf("packages folder holds %v, want only %s", names, e.label)
				}
				if names := ghNames(t, e.tmp); len(names) != 0 {
					t.Errorf("TMPDIR holds %v", names)
				}
				installed := filepath.Join(e.packagesDir(), e.label)
				if data, err := os.ReadFile(filepath.Join(installed, "VERSION")); err != nil || string(data) != ghVersion+"\n" {
					t.Errorf("VERSION = %q, %v", data, err)
				}
				if info, err := os.Stat(filepath.Join(installed, "install.sh")); err != nil || info.Mode().Perm()&0o100 == 0 {
					t.Errorf("install.sh not executable: %v, %v", info, err)
				}
				if !strings.Contains(res.stdout, filepath.Join(installed, "bin", "hive")) {
					t.Errorf("stdout does not name bin/hive:\n%s", res.stdout)
				}
				ghAssertRequests(t, e, ghExpectedRequests(e, true))
			})
		}
	})
}

func TestGetHivePropagatesInstallerExitCode(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: nil, env: []string{"HIVE_TEST_EXIT=7"}})
		if res.code != 7 {
			t.Fatalf("exit %d, want 7\nstderr:\n%s", res.code, res.stderr)
		}
		if _, ran := e.recordText(); !ran {
			t.Fatal("install.sh did not run")
		}
	})
}

func TestGetHiveDryRunNeedsNoTerminal(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		res := e.run(ghRun{shell: shell, script: e.script(e.missingTTY()), args: []string{"--dry-run"}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		if got := e.recordLine("args:"); got != "args: [--dry-run]" {
			t.Errorf("%q", got)
		}
		if got := e.recordLine("stdin:"); got != "stdin:EOF" {
			t.Errorf("with --dry-run stdin must not be redirected to the terminal: %q", got)
		}
		if _, err := os.Stat(e.missingTTY()); !os.IsNotExist(err) {
			t.Errorf("the terminal path was created: %v", err)
		}
		ghAssertRequests(t, e, ghExpectedRequests(e, true))
	})
}

func TestGetHiveWithoutTerminalFailsBeforeAnyRequest(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		res := e.run(ghRun{shell: shell, script: e.script(e.missingTTY()), args: []string{"--hosts", "claude"}})
		e.assertNothingHappened(res)
		if !strings.Contains(res.stderr, "terminal") || !strings.Contains(res.stderr, "--dry-run") {
			t.Errorf("message should name the terminal and --dry-run:\n%s", res.stderr)
		}
		if _, err := os.Stat(e.missingTTY()); !os.IsNotExist(err) {
			t.Errorf("the terminal path was created: %v", err)
		}
	})
}

func TestGetHiveVersionFlagSkipsLatestLookup(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--hosts", "a", "--version", ghVersion, "--dry-run", "--yes"}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		if got, want := e.recordLine("args:"), "args: [--hosts] [a] [--dry-run] [--yes]"; got != want {
			t.Errorf("%q, want %q (--version is consumed, order preserved)", got, want)
		}
		ghAssertRequests(t, e, ghExpectedRequests(e, false))
	})
}

func TestGetHiveHelp(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		res := e.run(ghRun{shell: shell, script: e.script(e.missingTTY()), args: []string{"--help"}})
		if res.code != 0 || !strings.Contains(res.stdout, "Usage:") {
			t.Fatalf("exit %d, stdout %q", res.code, res.stdout)
		}
		if e.srv.count() != 0 || len(ghNames(t, e.home)) != 0 {
			t.Fatal("--help made requests or wrote files")
		}
	})
}

func TestGetHiveRejectsInvalidVersionFlag(t *testing.T) {
	cases := [][]string{
		{"--version"},
		{"--version", ""},
		{"--version", "1.2"},
		{"--version", "v1.2.3"},
		{"--version", "01.2.3"},
		{"--version", "1.2.3-"},
		{"--version", "1.2.3/../x"},
		{"--version", "1.2.3 4"},
		{"--version", "1.2.3-rc..1"},
		{"--version", "1.2.3-01"},
		{"--version", strings.Repeat("1", 130) + ".2.3"},
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for i, args := range cases {
			e := newGhEnv(t)
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: args})
			if res.code == 0 || e.srv.count() != 0 || len(ghNames(t, e.home)) != 0 {
				t.Errorf("case %d %q: exit %d, %d requests, home %v", i, args, res.code, e.srv.count(), ghNames(t, e.home))
			}
			if _, ran := e.recordText(); ran {
				t.Errorf("case %d: install.sh ran", i)
			}
		}
	})
}

func TestGetHiveAcceptsPrereleaseVersion(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		label := fmt.Sprintf("hive-1.2.3-rc.1-%s-%s", e.osName, e.arch)
		e.srv.archiveName = label + ".tar.gz"
		e.setArchive(ghGoodEntries(label))
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", "1.2.3-rc.1"}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		if _, err := os.Stat(filepath.Join(e.packagesDir(), label, "install.sh")); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGetHiveRejectsBadLatestRedirect(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"invalid version":      func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/releases/tag/vbad", 302) },
		"no v prefix":          func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/releases/tag/9.9.9", 302) },
		"extra path":           func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/releases/tag/v9.9.9/extra", 302) },
		"other repository":     func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/other/releases/tag/v9.9.9", 302) },
		"not a tag page":       func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/releases", 302) },
		"no redirect":          func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) },
		"not found":            http.NotFound,
		"version with a query": func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/releases/tag/v9.9.9?x=1", 302) },
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for name, handler := range cases {
			t.Run(name, func(t *testing.T) {
				e := newGhEnv(t)
				e.setHandlers(handler, nil, nil)
				before := e.seedPrevious()
				res := e.run(ghRun{shell: shell, script: e.script(e.tty)})
				e.assertRejected(res, before)
				for _, r := range e.srv.reqs() {
					if strings.Contains(r, "/releases/download/") {
						t.Fatalf("downloaded after an invalid latest version: %v", e.srv.reqs())
					}
				}
			})
		}
	})
}

func TestGetHiveRejectsBadChecksumFiles(t *testing.T) {
	name := func(e *ghEnv) string { return e.srv.archiveName }
	good := func(e *ghEnv) string { return fmt.Sprintf("%x", sha256.Sum256(e.srv.archive)) }
	cases := map[string]func(e *ghEnv) string{
		"digest mismatch":       func(e *ghEnv) string { return strings.Repeat("0", 64) + "  " + name(e) + "\n" },
		"two lines":             func(e *ghEnv) string { return good(e) + "  " + name(e) + "\n" + good(e) + "  " + name(e) + "\n" },
		"trailing blank line":   func(e *ghEnv) string { return good(e) + "  " + name(e) + "\n\n" },
		"another archive name":  func(e *ghEnv) string { return good(e) + "  hive-0.0.1-" + e.osName + "-" + e.arch + ".tar.gz\n" },
		"path in the name":      func(e *ghEnv) string { return good(e) + "  ./" + name(e) + "\n" },
		"uppercase digest":      func(e *ghEnv) string { return strings.ToUpper(good(e)) + "  " + name(e) + "\n" },
		"non hexadecimal":       func(e *ghEnv) string { return strings.Repeat("g", 64) + "  " + name(e) + "\n" },
		"short digest":          func(e *ghEnv) string { return good(e)[:63] + "  " + name(e) + "\n" },
		"long digest":           func(e *ghEnv) string { return good(e) + "0  " + name(e) + "\n" },
		"single space":          func(e *ghEnv) string { return good(e) + " " + name(e) + "\n" },
		"no archive name":       func(e *ghEnv) string { return good(e) + "\n" },
		"empty":                 func(e *ghEnv) string { return "" },
		"no trailing newline":   func(e *ghEnv) string { return good(e) + "  " + name(e) },
		"carriage return":       func(e *ghEnv) string { return good(e) + "  " + name(e) + "\r\n" },
		"hash of other content": func(e *ghEnv) string { return fmt.Sprintf("%x  %s\n", sha256.Sum256([]byte("other")), name(e)) },
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for label, build := range cases {
			t.Run(label, func(t *testing.T) {
				e := newGhEnv(t)
				before := e.seedPrevious()
				e.setSum(build(e))
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
				e.assertRejected(res, before)
			})
		}
	})
}

func TestGetHiveRejectsDisallowedRedirectHosts(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		for _, kind := range []string{"archive", "checksum", "latest"} {
			for _, evil := range []string{"host in path", "userinfo before other host", "userinfo before allowed host", "plain other host"} {
				t.Run(kind+"/"+evil, func(t *testing.T) {
					e := newGhEnv(t)
					before := e.seedPrevious()
					port := e.srv.port()
					target := func(path string) string {
						switch evil {
						case "host in path":
							return "http://localhost:" + port + "/127.0.0.1" + path
						case "userinfo before other host":
							return "http://127.0.0.1:" + port + "@localhost:" + port + path
						case "userinfo before allowed host":
							return "http://localhost:" + port + "@127.0.0.1:" + port + path
						default:
							return "http://localhost:" + port + path
						}
					}
					redirect := func(path string) http.HandlerFunc {
						return func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target(path), http.StatusFound) }
					}
					args := []string{"--version", ghVersion}
					switch kind {
					case "archive":
						e.setHandlers(nil, redirect("/"+e.srv.archiveName), nil)
					case "checksum":
						e.setHandlers(nil, nil, redirect("/"+e.srv.archiveName+".sha256"))
					case "latest":
						e.setHandlers(redirect("/releases/tag/v"+ghVersion), nil, nil)
						args = nil
					}
					res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: args})
					e.assertRejected(res, before)
					if !strings.Contains(res.stderr, "not allowed") && kind != "latest" {
						t.Errorf("message should say the host is not allowed:\n%s", res.stderr)
					}
				})
			}
		}
	})
}

func TestGetHiveFollowsRedirectToAllowedHost(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		to := func(path string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, path, http.StatusFound) }
		}
		e.setHandlers(nil, to("/assets/"+e.srv.archiveName), to("/assets/"+e.srv.archiveName+".sha256"))
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		want := append(ghExpectedRequests(e, false),
			"GET /assets/"+e.srv.archiveName, "GET /assets/"+e.srv.archiveName+".sha256")
		sort.Strings(want)
		ghAssertRequests(t, e, want)
	})
}

func TestGetHiveRejectsOversizeDownloads(t *testing.T) {
	cases := map[string]func(e *ghEnv){
		"checksum with a declared length": func(e *ghEnv) {
			e.setHandlers(nil, nil, func(w http.ResponseWriter, r *http.Request) { ghStreamZeros(w, 300, true) })
		},
		"checksum without a declared length": func(e *ghEnv) {
			e.setHandlers(nil, nil, func(w http.ResponseWriter, r *http.Request) { ghStreamZeros(w, 300, false) })
		},
		"archive with a declared length": func(e *ghEnv) {
			e.setHandlers(nil, func(w http.ResponseWriter, r *http.Request) { ghStreamZeros(w, ghMaxArchive+1, true) }, nil)
		},
		"archive without a declared length": func(e *ghEnv) {
			e.setHandlers(nil, func(w http.ResponseWriter, r *http.Request) { ghStreamZeros(w, ghMaxArchive+1, false) }, nil)
		},
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for name, setup := range cases {
			t.Run(name, func(t *testing.T) {
				e := newGhEnv(t)
				before := e.seedPrevious()
				setup(e)
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
				e.assertRejected(res, before)
			})
		}
	})
}

func TestGetHiveRejectsUnsafeArchives(t *testing.T) {
	good := func(e *ghEnv) []ghEntry { return ghGoodEntries(e.label) }
	with := func(extra ...ghEntry) func(e *ghEnv) []ghEntry {
		return func(e *ghEnv) []ghEntry { return append(good(e), extra...) }
	}
	cases := map[string]func(e *ghEnv) []ghEntry{
		"parent traversal inside the root": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/../x", typ: tar.TypeReg, body: "x", mode: 0o644})(e)
		},
		"parent traversal from the start": with(ghEntry{name: "../x", typ: tar.TypeReg, body: "x", mode: 0o644}),
		"traversal that ends the name": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/sub/..", typ: tar.TypeDir, mode: 0o755})(e)
		},
		"absolute path":       with(ghEntry{name: "/tmp/hive-get-test-x", typ: tar.TypeReg, body: "x", mode: 0o644}),
		"another root folder": with(ghEntry{name: "other/file", typ: tar.TypeReg, body: "x", mode: 0o644}),
		"a sibling that shares the prefix": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "-x/file", typ: tar.TypeReg, body: "x", mode: 0o644})(e)
		},
		"only another root folder": func(e *ghEnv) []ghEntry {
			return []ghEntry{{name: "other/", typ: tar.TypeDir, mode: 0o755}, {name: "other/install.sh", typ: tar.TypeReg, body: ghInstallStub, mode: 0o755}}
		},
		"symbolic link entry": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/link", typ: tar.TypeSymlink, link: "/etc"})(e)
		},
		"symbolic link to a relative target": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/link", typ: tar.TypeSymlink, link: "../.."})(e)
		},
		"install.sh as a symbolic link": func(e *ghEnv) []ghEntry {
			entries := good(e)
			entries[1] = ghEntry{name: e.label + "/install.sh", typ: tar.TypeSymlink, link: "/bin/true"}
			return entries
		},
		"hard link entry": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/hl", typ: tar.TypeLink, link: "/etc/hosts"})(e)
		},
		"hard link in the root": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/hl", typ: tar.TypeLink, link: e.label + "/VERSION"})(e)
		},
		"character device": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/null", typ: tar.TypeChar, mode: 0o666})(e)
		},
		"fifo": func(e *ghEnv) []ghEntry {
			return with(ghEntry{name: e.label + "/fifo", typ: tar.TypeFifo, mode: 0o644})(e)
		},
		"no install.sh": func(e *ghEnv) []ghEntry {
			return []ghEntry{{name: e.label + "/", typ: tar.TypeDir, mode: 0o755}, {name: e.label + "/VERSION", typ: tar.TypeReg, body: "x", mode: 0o644}}
		},
		"install.sh is a folder": func(e *ghEnv) []ghEntry {
			return []ghEntry{{name: e.label + "/", typ: tar.TypeDir, mode: 0o755}, {name: e.label + "/install.sh/", typ: tar.TypeDir, mode: 0o755}}
		},
		"a file named like the root": func(e *ghEnv) []ghEntry {
			return []ghEntry{{name: e.label, typ: tar.TypeReg, body: "x", mode: 0o644}}
		},
		"not a gzip archive": nil,
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for name, build := range cases {
			t.Run(name, func(t *testing.T) {
				e := newGhEnv(t)
				before := e.seedPrevious()
				if build == nil {
					e.srv.mu.Lock()
					e.srv.archive = []byte("this is not a gzip archive\n")
					e.srv.sum = ghSum(e.srv.archive, e.srv.archiveName)
					e.srv.mu.Unlock()
				} else {
					e.setArchive(build(e))
				}
				outside := filepath.Join(e.work, "..", "x")
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
				e.assertRejected(res, before)
				for _, p := range []string{outside, "/tmp/hive-get-test-x", filepath.Join(e.packagesDir(), "x"), filepath.Join(e.home, "x")} {
					if _, err := os.Lstat(p); err == nil {
						t.Errorf("archive content escaped to %s", p)
					}
				}
			})
		}
	})
}

func TestGetHiveReplacesAnExistingPackage(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		e.seedPrevious()
		sibling := ghSnapshot(t, filepath.Join(e.packagesDir(), "hive-1.0.0-"+e.osName+"-"+e.arch))
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		want := []string{"hive-1.0.0-" + e.osName + "-" + e.arch, e.label}
		sort.Strings(want)
		if names := ghNames(t, e.packagesDir()); strings.Join(names, ",") != strings.Join(want, ",") {
			t.Fatalf("packages folder holds %v, want %v (no staging or hidden folders)", names, want)
		}
		if _, err := os.Stat(filepath.Join(e.packagesDir(), e.label, "marker")); !os.IsNotExist(err) {
			t.Errorf("the previous package content is still there: %v", err)
		}
		if data, err := os.ReadFile(filepath.Join(e.packagesDir(), e.label, "VERSION")); err != nil || string(data) != ghVersion+"\n" {
			t.Errorf("VERSION = %q, %v", data, err)
		}
		ghEqualTrees(t, ghSnapshot(t, filepath.Join(e.packagesDir(), "hive-1.0.0-"+e.osName+"-"+e.arch)), sibling, "another version's folder")
		if got, want := e.recordLine("packages:"), "packages: "+strings.Join(want, " "); got != want {
			t.Errorf("%q, want %q", got, want)
		}
	})
}

func ghFakeMv(t *testing.T, e *ghEnv) {
	t.Helper()
	e.fakeBin("mv", `n=$(cat "$HIVE_TEST_MV_COUNT" 2>/dev/null || echo 0)
n=$((n + 1))
echo "$n" > "$HIVE_TEST_MV_COUNT"
if [ "$n" -eq "${HIVE_TEST_MV_FAIL_AT:-0}" ]; then
  case ${HIVE_TEST_MV_MODE:-fail} in
    fail) echo "fake mv: failing call $n" >&2; exit 1 ;;
    hang) trap '' INT TERM HUP; : > "$HIVE_TEST_MV_MARKER"; sleep 1; echo "fake mv: failing call $n" >&2; exit 1 ;;
  esac
fi
if [ "$n" -eq "${HIVE_TEST_MV_SQUAT_AT:-0}" ]; then
  "$HIVE_TEST_REAL_MV" "$@" || exit 1
  case ${HIVE_TEST_SQUAT_KIND:-dir} in
    dir) mkdir "$HIVE_TEST_SQUAT" && : > "$HIVE_TEST_SQUAT/squatter" ;;
    symlink) ln -s "$HIVE_TEST_SQUAT_TARGET" "$HIVE_TEST_SQUAT" ;;
  esac
  exit 0
fi
exec "$HIVE_TEST_REAL_MV" "$@"
`)
}

func ghMvEnv(t *testing.T, e *ghEnv, extra ...string) []string {
	return append([]string{
		"HIVE_TEST_MV_COUNT=" + filepath.Join(e.out, "mv-count"),
		"HIVE_TEST_MV_MARKER=" + filepath.Join(e.out, "mv-marker"),
		"HIVE_TEST_REAL_MV=" + ghRealTool(t, "mv"),
	}, extra...)
}

func TestGetHiveRestoresThePreviousPackageWhenTheSecondMoveFails(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		ghFakeMv(t, e)
		before := e.seedPrevious()
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, env: ghMvEnv(t, e, "HIVE_TEST_MV_FAIL_AT=2")})
		e.assertRejected(res, before)
		if !strings.Contains(res.stderr, "fake mv: failing call 2") {
			t.Errorf("the fake mv did not fail the second call:\n%s", res.stderr)
		}
	})
}

func TestGetHiveRestoresThePreviousPackageOnSignals(t *testing.T) {
	signals := map[string]syscall.Signal{"INT": syscall.SIGINT, "TERM": syscall.SIGTERM, "HUP": syscall.SIGHUP}
	ghEachShell(t, func(t *testing.T, shell string) {
		for name, sig := range signals {
			t.Run(name, func(t *testing.T) {
				e := newGhEnv(t)
				ghFakeMv(t, e)
				before := e.seedPrevious()
				marker := filepath.Join(e.out, "mv-marker")
				res := e.run(ghRun{
					shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion},
					env:    ghMvEnv(t, e, "HIVE_TEST_MV_FAIL_AT=2", "HIVE_TEST_MV_MODE=hang"),
					signal: sig, signalMarker: marker,
				})
				e.assertRejected(res, before)
				if want := 128 + int(sig); res.code != want {
					t.Errorf("exit %d, want %d: the trap must exit with the signal's status", res.code, want)
				}
			})
		}
	})
}

func TestGetHiveNeverMovesIntoAPackageThatAppearedMidway(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		for _, kind := range []string{"dir", "symlink"} {
			t.Run(kind, func(t *testing.T) {
				e := newGhEnv(t)
				ghFakeMv(t, e)
				e.seedPrevious()
				squat := filepath.Join(e.packagesDir(), e.label)
				target := filepath.Join(e.root, "squat-target")
				if err := os.MkdirAll(target, 0o700); err != nil {
					t.Fatal(err)
				}
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion},
					env: ghMvEnv(t, e, "HIVE_TEST_MV_SQUAT_AT=1", "HIVE_TEST_SQUAT="+squat, "HIVE_TEST_SQUAT_KIND="+kind, "HIVE_TEST_SQUAT_TARGET="+target)})
				if res.code == 0 {
					t.Fatalf("exit 0, want a failure\nstderr:\n%s", res.stderr)
				}
				if _, ran := e.recordText(); ran {
					t.Fatal("install.sh ran")
				}
				if names := ghNames(t, target); len(names) != 0 {
					t.Errorf("a package was moved through the link into %v", names)
				}
				if kind == "dir" {
					if names := ghNames(t, squat); strings.Join(names, ",") != "squatter" {
						t.Errorf("the folder that appeared was modified: %v", names)
					}
				}
				if !strings.Contains(res.stderr, "kept at") {
					t.Errorf("message should say where the previous package was kept:\n%s", res.stderr)
				}
				found := false
				for _, n := range ghNames(t, e.packagesDir()) {
					if strings.HasPrefix(n, ".old.") {
						if data, err := os.ReadFile(filepath.Join(e.packagesDir(), n, e.label, "marker")); err == nil && string(data) == "previous package\n" {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("the previous package was lost: %v", ghNames(t, e.packagesDir()))
				}
			})
		}
	})
}

func TestGetHiveRejectsUnsupportedPlatforms(t *testing.T) {
	cases := []struct{ name, flag, value, want string }{
		{"operating system", "-s", "FreeBSD", "operating system"},
		{"architecture", "-m", "riscv64", "architecture"},
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				e := newGhEnv(t)
				real := ghRealTool(t, "uname")
				e.fakeBin("uname", fmt.Sprintf("if [ \"$1\" = %s ]; then echo %s; else exec %s \"$@\"; fi\n", c.flag, c.value, real))
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--dry-run"}})
				e.assertNothingHappened(res)
				if !strings.Contains(res.stderr, c.want) {
					t.Errorf("message should mention the %s:\n%s", c.want, res.stderr)
				}
			})
		}
	})
}

func TestGetHiveMapsPlatformNames(t *testing.T) {
	cases := []struct{ s, m, label string }{
		{"Darwin", "arm64", "darwin-arm64"},
		{"Linux", "aarch64", "linux-arm64"},
		{"Linux", "x86_64", "linux-amd64"},
		{"Linux", "amd64", "linux-amd64"},
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		for _, c := range cases {
			t.Run(c.label+"/"+c.m, func(t *testing.T) {
				e := newGhEnv(t)
				label := "hive-" + ghVersion + "-" + c.label
				e.srv.archiveName = label + ".tar.gz"
				e.setArchive(ghGoodEntries(label))
				e.fakeBin("uname", fmt.Sprintf("case \"$1\" in -s) echo %s ;; -m) echo %s ;; esac\n", c.s, c.m))
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}})
				if res.code != 0 {
					t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
				}
				if _, err := os.Stat(filepath.Join(e.packagesDir(), label, "install.sh")); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}

func TestGetHiveNamesAMissingRequirementBeforeAnyRequest(t *testing.T) {
	needed := []string{"curl", "tar", "mktemp", "uname", "find", "wc", "mv", "rm", "rmdir", "mkdir", "shasum", "sha256sum"}
	support := []string{"ls", "tr", "cat", "dirname", "sleep", "sh"} // used by the fixtures, not by the script
	linkTools := func(t *testing.T, dir string, names []string, skip map[string]bool) {
		for _, name := range names {
			if skip[name] {
				continue
			}
			if path, err := exec.LookPath(name); err == nil {
				if err := os.Symlink(path, filepath.Join(dir, name)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	ghEachShell(t, func(t *testing.T, shell string) {
		t.Run("baseline with every tool", func(t *testing.T) {
			e := newGhEnv(t)
			tools := t.TempDir()
			linkTools(t, tools, append(append([]string{}, needed...), support...), nil)
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, pathOnly: tools})
			if res.code != 0 {
				t.Fatalf("the restricted PATH is not sufficient: exit %d\nstderr:\n%s", res.code, res.stderr)
			}
		})
		for _, missing := range []string{"curl", "tar", "mktemp", "uname", "find", "wc"} {
			t.Run("missing "+missing, func(t *testing.T) {
				e := newGhEnv(t)
				tools := t.TempDir()
				linkTools(t, tools, append(append([]string{}, needed...), support...), map[string]bool{missing: true})
				res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, pathOnly: tools})
				e.assertNothingHappened(res)
				if !strings.Contains(res.stderr, missing) {
					t.Errorf("message should name %s:\n%s", missing, res.stderr)
				}
			})
		}
		t.Run("missing every sha256 tool", func(t *testing.T) {
			e := newGhEnv(t)
			tools := t.TempDir()
			linkTools(t, tools, append(append([]string{}, needed...), support...), map[string]bool{"shasum": true, "sha256sum": true})
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, pathOnly: tools})
			e.assertNothingHappened(res)
			if !strings.Contains(res.stderr, "shasum") {
				t.Errorf("message should name shasum:\n%s", res.stderr)
			}
		})
		t.Run("HOME unset", func(t *testing.T) {
			e := newGhEnv(t)
			// The runner always sets HOME, so an empty value stands for "unset".
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, env: []string{"HOME="}})
			if res.code == 0 || e.srv.count() != 0 || !strings.Contains(res.stderr, "HOME") {
				t.Fatalf("exit %d, %d requests, stderr:\n%s", res.code, e.srv.count(), res.stderr)
			}
		})
	})
}

func TestGetHiveIgnoresCurlConfigAndTarOptions(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		// A user's ~/.curlrc would send every request through a dead proxy unless
		// curl is started with -q.
		mustWriteFile(t, filepath.Join(e.home, ".curlrc"), []byte("proxy = \"http://127.0.0.1:1\"\n"), 0o600)
		tarLog := filepath.Join(e.out, "tar-log")
		e.fakeBin("tar", fmt.Sprintf("printf '%%s\\n' \"${TAR_OPTIONS-unset}\" >> %q\nexec %s \"$@\"\n", tarLog, ghRealTool(t, "tar")))
		res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, env: []string{"TAR_OPTIONS=--absolute-names --no-overwrite-dir"}})
		if res.code != 0 {
			t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
		}
		data, err := os.ReadFile(tarLog)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		if len(lines) < 3 {
			t.Fatalf("tar ran %d times, want the two listings and the extraction", len(lines))
		}
		for _, line := range lines {
			if line != "" && line != "unset" {
				t.Errorf("tar saw TAR_OPTIONS=%q, want it emptied", line)
			}
		}
	})
}

func TestGetHiveUsesXDGDataHomeOnlyWhenAbsolute(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		t.Run("absolute", func(t *testing.T) {
			e := newGhEnv(t)
			xdg := filepath.Join(e.root, "xdg data")
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, env: []string{"XDG_DATA_HOME=" + xdg}})
			if res.code != 0 {
				t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
			}
			if _, err := os.Stat(filepath.Join(xdg, "hive", "packages", e.label, "install.sh")); err != nil {
				t.Fatal(err)
			}
			if names := ghNames(t, e.home); len(names) != 0 {
				t.Errorf("HOME was written although XDG_DATA_HOME is absolute: %v", names)
			}
		})
		t.Run("relative is ignored", func(t *testing.T) {
			e := newGhEnv(t)
			res := e.run(ghRun{shell: shell, script: e.script(e.tty), args: []string{"--version", ghVersion}, env: []string{"XDG_DATA_HOME=relative/data"}})
			if res.code != 0 {
				t.Fatalf("exit %d\nstderr:\n%s", res.code, res.stderr)
			}
			if _, err := os.Stat(filepath.Join(e.packagesDir(), e.label, "install.sh")); err != nil {
				t.Fatal(err)
			}
			if names := ghNames(t, e.work); len(names) != 0 {
				t.Errorf("a relative XDG_DATA_HOME was honored: %v", names)
			}
		})
	})
}

// TestGetHiveTruncatedCopiesDoNothing covers AC5: a copy cut at any point, as an
// interrupted download leaves, makes no request and writes no file.
func TestGetHiveTruncatedCopiesDoNothing(t *testing.T) {
	ghEachShell(t, func(t *testing.T, shell string) {
		e := newGhEnv(t)
		script := e.script(e.tty)
		lastStart := bytes.LastIndex(script[:len(script)-1], []byte("\n")) + 1
		prevStart := bytes.LastIndex(script[:lastStart-1], []byte("\n")) + 1
		if string(script[lastStart:]) != ghFinalLine+"\n" {
			t.Fatalf("unexpected last line %q", script[lastStart:])
		}
		cuts := map[int]bool{}
		// Every byte of the last two lines, except the cut that only drops the
		// final newline and leaves a complete script.
		for k := prevStart; k <= len(script)-2; k++ {
			cuts[k] = true
		}
		// Every line start, plus a stride through the body.
		for i, b := range script {
			if b == '\n' && i+1 < prevStart {
				cuts[i+1] = true
			}
		}
		for k := 1; k < prevStart; k += 53 {
			cuts[k] = true
		}
		check := func(t *testing.T, k int, pipe bool) {
			res := e.run(ghRun{shell: shell, script: script[:k], args: []string{"--dry-run"}, pipe: pipe})
			if n := e.srv.count(); n != 0 {
				t.Fatalf("cut at %d (pipe=%v) made %d requests: %v\n%q", k, pipe, n, e.srv.reqs(), script[k-min(k, 30):k])
			}
			if names := ghNames(t, e.home); len(names) != 0 {
				t.Fatalf("cut at %d wrote into HOME: %v", k, names)
			}
			if names := ghNames(t, e.tmp); len(names) != 0 {
				t.Fatalf("cut at %d wrote into TMPDIR: %v", k, names)
			}
			if _, ran := e.recordText(); ran {
				t.Fatalf("cut at %d ran install.sh", k)
			}
			if k > lastStart && res.code == 0 {
				t.Fatalf("cut at %d inside the final line exited 0, want a syntax error", k)
			}
		}
		keys := make([]int, 0, len(cuts))
		for k := range cuts {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			check(t, k, false)
			if k >= prevStart {
				check(t, k, true)
			}
		}
	})
}
