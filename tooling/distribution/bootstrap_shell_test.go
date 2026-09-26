package distribution

// This file runs bootstrap.sh and install.sh through a real shell process,
// complementing the unit-level tests in bootstrap_test.go which exercise the
// Go helpers (download, Extract, RetainVerifiedInstaller, ...) directly but
// never invoke the shell entrypoints themselves. It closes the gap named by
// tasks.md T5's verification line: "Probar flujo canalizado con una shell
// real y stdin ocupado por el script" and "Offline no realiza solicitudes de
// red". No test here reaches a real network origin: every origin is a local
// httptest server, reached only by running a private test copy of
// bootstrap.sh whose fixed origin and protocol-restriction lines have been
// textually substituted for that local fixture (see testBootstrapScript).
// The shipped bootstrap.sh itself carries no runtime override for this.

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// --- shared fixtures and helpers -------------------------------------------------

func readRepoFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", name))
	if err != nil {
		t.Fatalf("reading %s: %v", name, err)
	}
	return data
}

func mustWriteFile(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
}

func assertEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	if len(entries) != 0 {
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		t.Fatalf("expected %s to stay empty (no partial install/scratch left behind), found %v", dir, names)
	}
}

// scriptPlatformParts mirrors bootstrap.sh's and install.sh's own uname
// mapping so fixture paths and the "platform" file match what the script
// under test will actually compute on this machine.
func scriptPlatformParts(t *testing.T) (osName, arch string) {
	t.Helper()
	switch runtime.GOOS {
	case "darwin", "linux":
		osName = runtime.GOOS
	default:
		t.Skipf("bootstrap.sh/install.sh only support darwin and linux; GOOS=%s", runtime.GOOS)
	}
	switch runtime.GOARCH {
	case "arm64", "amd64":
		arch = runtime.GOARCH
	default:
		t.Skipf("bootstrap.sh/install.sh only support arm64 and amd64; GOARCH=%s", runtime.GOARCH)
	}
	return osName, arch
}

// scriptPlatform returns bootstrap.sh's own "$platform_os-$platform_arch" form.
func scriptPlatform(t *testing.T) string {
	osName, arch := scriptPlatformParts(t)
	return osName + "-" + arch
}

type requestLog struct {
	mu    sync.Mutex
	paths []string
}

func (l *requestLog) record(p string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.paths = append(l.paths, p)
}

func (l *requestLog) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.paths)
}

func (l *requestLog) snapshot() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.paths...)
}

func (l *requestLog) has(path string) bool {
	for _, p := range l.snapshot() {
		if p == path {
			return true
		}
	}
	return false
}

func forEachShell(t *testing.T, f func(t *testing.T, shell string)) {
	// dash exercises H8: a failed redirection on the special builtin "exec"
	// exits the shell running it (exit 2, no message) rather than just
	// returning nonzero, unlike bash. Skipped where dash is not installed.
	for _, shell := range []string{"sh", "bash", "dash"} {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			if _, err := exec.LookPath(shell); err != nil {
				t.Skipf("required shell %q is not available: %v", shell, err)
			}
			f(t, shell)
		})
	}
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func exitCodeOf(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	t.Fatalf("running command: %v", err)
	return -1
}

// runUnderControllingTTY runs shellCommand inside a shell that has a genuine
// controlling terminal (via the "script" pty-allocating utility), while the
// command's own standard input may still be a plain pipe. This reproduces
// the real "curl ... | sh" situation in an interactive session: the process
// has no controlling terminal of its own beyond what "script" grants it, and
// bootstrap.sh must read prompts from /dev/tty, never from its piped stdin.
func runUnderControllingTTY(t *testing.T, env []string, shellCommand string) (combinedOutput string, exitCode int) {
	t.Helper()
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("the \"script\" utility needed to allocate a controlling terminal for this test is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		// util-linux script(1): -q quiet, -e propagate exit status, -c command.
		cmd = exec.CommandContext(ctx, "script", "-q", "-e", "-c", shellCommand, "/dev/null")
	case "darwin":
		// BSD script(1): the wrapped command's exit status is always script's own.
		cmd = exec.CommandContext(ctx, "script", "-q", "/dev/null", "sh", "-c", shellCommand)
	default:
		t.Skipf("no known \"script\" invocation for GOOS=%s", runtime.GOOS)
	}
	cmd.Env = env
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("command timed out; output so far:\n%s", out.String())
	}
	return out.String(), exitCodeOf(t, err)
}

// bootstrapEnv builds the environment for a bootstrap.sh run, isolated to a
// synthetic HOME and TMPDIR (t.TempDir()). The script under test carries no
// origin-related environment variable any more (see testBootstrapScript):
// its target origin is baked into the script text itself.
func bootstrapEnv(home, tmp string, extra ...string) []string {
	env := []string{
		"HOME=" + home,
		"TMPDIR=" + tmp,
		"PATH=" + os.Getenv("PATH"),
	}
	return append(env, extra...)
}

// testBootstrapScript returns the path to a private copy of bootstrap.sh
// with its fixed production origin placeholder, and (for a plain-HTTP
// originURL such as httptest.NewServer's) its HTTPS-only protocol
// restriction, textually substituted for a local fixture — the same kind of
// substitution the real deployment process performs on the origin
// placeholder alone (see bootstrap.sh's top comment). This is the only
// test-side seam: no environment variable reaches bootstrap.sh, and the
// shipped script is never altered.
func testBootstrapScript(t *testing.T, dir, originURL string) string {
	t.Helper()
	script := string(readRepoFile(t, "bootstrap.sh"))

	const originPlaceholder = "origin='https://HIVE_BOOTSTRAP_ORIGIN.invalid'"
	if !strings.Contains(script, originPlaceholder) {
		t.Fatal("bootstrap.sh's origin placeholder line changed; update testBootstrapScript to match it")
	}
	script = strings.Replace(script, originPlaceholder, "origin="+shellQuote(originURL), 1)

	if strings.HasPrefix(originURL, "http://") {
		const curlHTTPSOnly = "--proto '=https'"
		const wgetHTTPSOnly = "--https-only "
		if !strings.Contains(script, curlHTTPSOnly) || !strings.Contains(script, wgetHTTPSOnly) {
			t.Fatal("bootstrap.sh's HTTPS-only protocol restriction changed; update testBootstrapScript to match it")
		}
		script = strings.Replace(script, curlHTTPSOnly, "--proto '=http'", 1)
		script = strings.Replace(script, wgetHTTPSOnly, "", 1)
	}

	path := filepath.Join(dir, "bootstrap.sh")
	mustWriteFile(t, path, []byte(script), 0755)
	return path
}

// runBootstrapPiped writes a test copy of bootstrap.sh (see
// testBootstrapScript) to a temp file and pipes it into a shell's stdin
// ("cat bootstrap.sh | sh -s -- ..."), exactly like "curl ... | sh", so the
// shell's own fd 0 is a pipe throughout the run.
func runBootstrapPiped(t *testing.T, shell, originURL, version string, extraArgs string, home, tmp string, extraEnv ...string) (out string, exitCode int) {
	t.Helper()
	scriptPath := testBootstrapScript(t, t.TempDir(), originURL)
	env := bootstrapEnv(home, tmp, extraEnv...)
	args := "--version " + version
	if extraArgs != "" {
		args = extraArgs
	}
	shellCommand := fmt.Sprintf("cat %s | %s -s -- %s", shellQuote(scriptPath), shell, args)
	return runUnderControllingTTY(t, env, shellCommand)
}

// --- Task 1: no controlling TTY must fail closed before any network use --------

func TestBootstrapFailsClosedWithoutControllingTTY(t *testing.T) {
	forEachShell(t, func(t *testing.T, shell string) {
		log := &requestLog{}
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.record(r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		home := t.TempDir()
		tmp := t.TempDir()
		env := bootstrapEnv(home, tmp)
		scriptPath := testBootstrapScript(t, t.TempDir(), srv.URL)
		script, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, shell, "-s", "--", "--version", "1.2.3")
		cmd.Stdin = bytes.NewReader(script)
		cmd.Env = env
		// Detach from any controlling terminal this test process might have
		// inherited, so /dev/tty is genuinely unavailable to the child,
		// regardless of how the test runner itself was launched.
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		if ctx.Err() == context.DeadlineExceeded {
			t.Fatalf("bootstrap.sh hung instead of failing closed without a TTY; stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
		if err == nil {
			t.Fatalf("expected bootstrap.sh to fail without a controlling TTY; stdout=%q", stdout.String())
		}
		const wantMsg = "A readable and writable terminal is required; use the offline package without a TTY."
		if !strings.Contains(stderr.String(), wantMsg) {
			t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), wantMsg)
		}
		if got := log.count(); got != 0 {
			t.Fatalf("bootstrap.sh made %d network request(s) without a TTY, want 0: %v", got, log.snapshot())
		}
		// Nothing is created before the TTY gate (scratch is made afterwards),
		// so both sandboxes must still be completely empty.
		assertEmptyDir(t, tmp)
		assertEmptyDir(t, home)
	})
}

// TestBootstrapShippedScriptHasNoTestOverride guards M8's own goal directly:
// the bootstrap.sh actually shipped (read straight off disk, never a
// testBootstrapScript copy) must carry no runtime test-origin override or
// environment-variable escape hatch at all. The only test seam left is the
// textual substitution testBootstrapScript performs on a private copy.
func TestBootstrapShippedScriptHasNoTestOverride(t *testing.T) {
	script := string(readRepoFile(t, "bootstrap.sh"))
	for _, forbidden := range []string{"HIVE_BOOTSTRAP_TEST_ORIGIN", "TEST_ORIGIN", "getenv", "GETENV"} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("bootstrap.sh contains %q; the shipped script must carry no runtime test override", forbidden)
		}
	}
	const originPlaceholder = "origin='https://HIVE_BOOTSTRAP_ORIGIN.invalid'"
	if !strings.Contains(script, originPlaceholder) {
		t.Fatal("bootstrap.sh no longer carries its fixed origin placeholder")
	}
	if !strings.Contains(script, "--proto '=https'") || !strings.Contains(script, "--https-only") {
		t.Fatal("bootstrap.sh no longer restricts downloads to HTTPS")
	}
}

// --- Task 1 (bonus): delegation reads the terminal, never the piped script -----

// TestBootstrapDelegatesControllingTerminalNotPipedScript proves the core
// property the audit gap named: when bootstrap.sh does proceed, it hands the
// packaged manager the controlling terminal (via "< /dev/tty"), never the
// pipe still carrying the rest of the piped script's bytes.
func TestBootstrapDelegatesControllingTerminalNotPipedScript(t *testing.T) {
	version := "9.9.9"
	platform := scriptPlatform(t)

	resultDir := t.TempDir()
	resultFile := filepath.Join(resultDir, "result.txt")

	stub := []byte("#!/bin/sh\n" +
		"if [ -t 0 ]; then state=tty; else state=pipe; fi\n" +
		"printf '%s|%s\\n' \"$state\" \"$*\" > \"$HIVE_TEST_RESULT_FILE\"\n")
	stubDigest := Digest(stub)

	hivePath := "/versions/" + version + "/" + platform + "/hive"
	shaPath := hivePath + ".sha256"

	log := &requestLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		switch r.URL.Path {
		case hivePath:
			_, _ = w.Write(stub)
		case shaPath:
			fmt.Fprintf(w, "%s\n", stubDigest)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	home := t.TempDir()
	tmp := t.TempDir()
	out, exitCode := runBootstrapPiped(t, "sh", srv.URL, version, "", home, tmp, "HIVE_TEST_RESULT_FILE="+resultFile)
	if exitCode != 0 {
		t.Fatalf("bootstrap.sh exited %d, want 0; output:\n%s", exitCode, out)
	}

	result, err := os.ReadFile(resultFile)
	if err != nil {
		t.Fatalf("reading delegated manager's result file: %v; bootstrap output:\n%s", err, out)
	}
	if !strings.HasPrefix(string(result), "tty|") {
		t.Fatalf("delegated manager's stdin was not the controlling terminal (piped script bytes may have leaked through): %q", string(result))
	}
	wantArgs := fmt.Sprintf("bootstrap --origin %s --version %s --manager", srv.URL, version)
	if !strings.Contains(string(result), wantArgs) {
		t.Fatalf("delegated manager did not receive the expected arguments: %q", string(result))
	}
	if !log.has(hivePath) || !log.has(shaPath) {
		t.Fatalf("expected requests for %s and %s, got %v", hivePath, shaPath, log.snapshot())
	}
	assertEmptyDir(t, tmp)
}

// --- Task 3: download failures through the real shell, no partial install ------

func TestBootstrapRejectsNonexistentVersionThroughRealShell(t *testing.T) {
	version := "9.9.9"
	platform := scriptPlatform(t)
	hivePath := "/versions/" + version + "/" + platform + "/hive"

	log := &requestLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		// No release was ever published for this version/platform.
		http.NotFound(w, r)
	}))
	defer srv.Close()

	home := t.TempDir()
	tmp := t.TempDir()
	out, exitCode := runBootstrapPiped(t, "sh", srv.URL, version, "", home, tmp)
	if exitCode == 0 {
		t.Fatalf("expected bootstrap.sh to fail for a nonexistent version; output:\n%s", out)
	}
	if !strings.Contains(out, "Could not download the Hive manager.") {
		t.Fatalf("output = %q, want the manager-download failure message", out)
	}
	if !log.has(hivePath) {
		t.Fatalf("expected a request for %s, got %v", hivePath, log.snapshot())
	}
	assertEmptyDir(t, tmp)
}

func TestBootstrapRejectsChecksumMismatchThroughRealShell(t *testing.T) {
	version := "8.8.8"
	platform := scriptPlatform(t)
	hivePath := "/versions/" + version + "/" + platform + "/hive"
	shaPath := hivePath + ".sha256"

	managerBytes := []byte("#!/bin/sh\nexit 0\n")
	correct := Digest(managerBytes)
	wrong := []byte(correct)
	if wrong[0] == 'a' {
		wrong[0] = 'b'
	} else {
		wrong[0] = 'a'
	}

	log := &requestLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		switch r.URL.Path {
		case hivePath:
			_, _ = w.Write(managerBytes)
		case shaPath:
			fmt.Fprintf(w, "%s\n", string(wrong))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	home := t.TempDir()
	tmp := t.TempDir()
	out, exitCode := runBootstrapPiped(t, "sh", srv.URL, version, "", home, tmp)
	if exitCode == 0 {
		t.Fatalf("expected bootstrap.sh to fail on a checksum mismatch; output:\n%s", out)
	}
	if !strings.Contains(out, "The manager checksum did not match.") {
		t.Fatalf("output = %q, want the checksum-mismatch message", out)
	}
	if !log.has(hivePath) || !log.has(shaPath) {
		t.Fatalf("expected requests for %s and %s, got %v", hivePath, shaPath, log.snapshot())
	}
	assertEmptyDir(t, tmp)
}

func TestBootstrapRejectsTruncatedManagerDownloadThroughRealShell(t *testing.T) {
	version := "7.7.7"
	platform := scriptPlatform(t)
	hivePath := "/versions/" + version + "/" + platform + "/hive"
	shaPath := hivePath + ".sha256"

	log := &requestLog{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		switch r.URL.Path {
		case hivePath:
			// Declare far more bytes than are actually sent, then hang up:
			// a truncated transfer, not a bad status or a bad checksum.
			w.Header().Set("Content-Length", "1000000")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("only-a-few-bytes"))
			if hj, ok := w.(http.Hijacker); ok {
				if conn, _, err := hj.Hijack(); err == nil {
					_ = conn.Close()
				}
			}
		case shaPath:
			fmt.Fprintf(w, "%s\n", Digest([]byte("irrelevant, never reached")))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	home := t.TempDir()
	tmp := t.TempDir()
	out, exitCode := runBootstrapPiped(t, "sh", srv.URL, version, "", home, tmp)
	if exitCode == 0 {
		t.Fatalf("expected bootstrap.sh to fail on a truncated manager download; output:\n%s", out)
	}
	if !strings.Contains(out, "Could not download the Hive manager.") {
		t.Fatalf("output = %q, want the manager-download failure message", out)
	}
	if !log.has(hivePath) {
		t.Fatalf("expected a request for %s, got %v", hivePath, log.snapshot())
	}
	if log.has(shaPath) {
		t.Fatalf("bootstrap.sh requested the checksum after a truncated manager download, want it to stop at the failed download: %v", log.snapshot())
	}
	assertEmptyDir(t, tmp)
}

// --- Task 2: the offline path makes zero network requests ----------------------

// runInstallOffline runs installScript (a candidate install.sh) directly
// (never piped, so "$0" resolves like a real "./install.sh" invocation)
// against a synthetic offline package, with curl and wget shadowed by
// recording stand-ins earlier on PATH so any network-tool invocation by the
// script or by its exec'd manager is observed, even though the real
// packaged manager (tooling/cli) is out of this test's scope and is replaced
// here by a minimal, non-networking stub.
func runInstallOffline(t *testing.T, installScript []byte) (exitCode int, networkHits int, out string) {
	t.Helper()
	goos, arch := scriptPlatformParts(t)

	pkgRoot := t.TempDir()
	mustWriteFile(t, filepath.Join(pkgRoot, "platform"), []byte(goos+"/"+arch+"\n"), 0644)
	mustWriteFile(t, filepath.Join(pkgRoot, "release.json"), []byte("{}\n"), 0644)
	hiveStub := []byte("#!/bin/sh\nexit 0\n")
	mustWriteFile(t, filepath.Join(pkgRoot, "bin", "hive"), hiveStub, 0755)
	mustWriteFile(t, filepath.Join(pkgRoot, "bin", "hive.sha256"), []byte(Digest(hiveStub)+"\n"), 0644)
	installPath := filepath.Join(pkgRoot, "install.sh")
	mustWriteFile(t, installPath, installScript, 0755)

	binDir := t.TempDir()
	logFile := filepath.Join(t.TempDir(), "network.log")
	fakeTool := []byte("#!/bin/sh\nprintf '%s %s\\n' \"$0\" \"$*\" >> \"$HIVE_TEST_NETWORK_LOG\"\nexit 1\n")
	for _, name := range []string{"curl", "wget"} {
		mustWriteFile(t, filepath.Join(binDir, name), fakeTool, 0755)
	}

	home := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", installPath)
	cmd.Env = []string{
		"HOME=" + home,
		"TMPDIR=" + home,
		"PATH=" + binDir + ":" + os.Getenv("PATH"),
		"HIVE_TEST_NETWORK_LOG=" + logFile,
	}
	var combined bytes.Buffer
	cmd.Stdout = &combined
	cmd.Stderr = &combined
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("install.sh hung; output so far:\n%s", combined.String())
	}
	exitCode = exitCodeOf(t, err)

	data, readErr := os.ReadFile(logFile)
	switch {
	case readErr == nil:
		trimmed := strings.TrimRight(string(data), "\n")
		if trimmed == "" {
			networkHits = 0
		} else {
			networkHits = len(strings.Split(trimmed, "\n"))
		}
	case os.IsNotExist(readErr):
		networkHits = 0
	default:
		t.Fatalf("reading network log: %v", readErr)
	}
	return exitCode, networkHits, combined.String()
}

// --- H1: the real manager, exec'd by a real bootstrap.sh, reaches its own
// interactive install summary and consent prompt through the online path ----

// repoRoot resolves the checkout root from this test file's own package
// directory, independent of the process's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// buildRealManager compiles the actual tooling/cli package (this change's
// `hive bootstrap` subcommand included) exactly as tooling/package builds a
// release, with -ldflags binding its embedded version.Current to
// productVersion so ValidateBootstrapPackage's running-version check passes,
// and setting tooling/distribution's allowLoopbackHTTPValue test seam so
// this test-only binary's own index/package downloads can reach a local
// plain-HTTP httptest fixture instead of a real HTTPS origin. A production
// binary built by tooling/package never passes that second flag.
func buildRealManager(t *testing.T, productVersion string) []byte {
	t.Helper()
	out := filepath.Join(t.TempDir(), "hive")
	ldflags := "-X tricell-hive/tooling/version.Current=" + productVersion +
		" -X tricell-hive/tooling/distribution.allowLoopbackHTTPValue=true"
	cmd := exec.Command("go", "build", "-o", out, "-ldflags", ldflags, "./tooling/cli")
	cmd.Dir = repoRoot(t)
	combined, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("building the real manager: %v\n%s", err, combined)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// buildRealPackageFixture packages this checkout's own content/ and
// integrations/agent-profiles.json — the same way tooling/distribution's own
// packageFixture builds a package directory (plain bytes for bin/hive rather
// than invoking tooling/package, which is out of this change's scope) — into
// a real gzip-tar archive Extract can accept, under one top-level directory.
func buildRealPackageFixture(t *testing.T, productVersion, platform string, managerBytes []byte) (archive []byte) {
	t.Helper()
	dir := t.TempDir()
	root := repoRoot(t)
	for _, rel := range []string{"content", "integrations/agent-profiles.json"} {
		mustCopyTree(t, filepath.Join(root, rel), filepath.Join(dir, rel))
	}
	// The manifest's own Platform (checked by verifyManifest against
	// runtime.GOOS+"/"+runtime.GOARCH) uses the slash form, distinct from the
	// hyphenated "platform" naming bootstrap.sh's own URLs and the
	// package/bin/platform marker file use.
	manifestPlatform := runtime.GOOS + "/" + runtime.GOARCH
	managerSHA256 := Digest(managerBytes)
	mustWriteFile(t, filepath.Join(dir, "bin", "hive"), managerBytes, 0755)
	mustWriteFile(t, filepath.Join(dir, "bin", "hive.sha256"), []byte(managerSHA256+"\n"), 0644)
	mustWriteFile(t, filepath.Join(dir, "install.sh"), []byte("#!/bin/sh\nexit 1\n"), 0755)
	mustWriteFile(t, filepath.Join(dir, "platform"), []byte(manifestPlatform+"\n"), 0644)
	files, err := Files(dir)
	if err != nil {
		t.Fatal(err)
	}
	manifest := Manifest{Version: 1, Platform: manifestPlatform, SourceID: strings.Repeat("a", 64), ProductVersion: productVersion, Files: files}
	metadata, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(dir, ManifestName), append(metadata, '\n'), 0644)
	return tarGzDirectoryForTest(t, dir)
}

func mustCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, info.Mode().Perm())
	}); err != nil {
		t.Fatal(err)
	}
}

// tarGzDirectoryForTest archives dir's own contents under one top-level
// directory (Extract's single-top-level-directory requirement), with no
// trailing slash on directory entry names, matching how tooling/package's
// own builder names them.
func tarGzDirectoryForTest(t *testing.T, dir string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	base := filepath.Dir(dir)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestBootstrapRealManagerReachesInteractiveInstallerThenLeavesNoChange is
// H1's shell-level end-to-end case: a REAL manager binary (compiled from
// ./tooling/cli, this change's `hive bootstrap` included), exec'd by the
// REAL bootstrap.sh through a real shell with a genuine controlling
// terminal, against a real fixture origin serving a real gzip-tar package
// (built the way tooling/distribution's own fixtures build one, per
// buildRealPackageFixture's doc comment). It proves the full online chain —
// download the index, download and verify the package, extract it safely,
// validate it, and hand off to the real interactive installer — actually
// works end to end, not just through the Go-level tests in bootstrap_test.go
// that call bootstrap() directly. The interactive wizard's own consent
// prompt is exercised thoroughly at the Go level already (see
// TestBootstrapReachesInstallSummaryAndConsentRetainsInstaller and
// TestBootstrapCancelLeavesNoPersistentChangeAndCleansItsScratch in
// tooling/cli); this test does not attempt to drive a real pty's raw-mode
// input timing past the first prompt, since bootstrap.sh's own controlling
// terminal is deliberately a real TTY, not a scriptable pipe (see
// runBootstrapPiped's doc comment on why). Reaching that first real prompt
// with zero persistent change is itself the meaningful, reliable assertion.
func TestBootstrapRealManagerReachesInteractiveInstallerThenLeavesNoChange(t *testing.T) {
	if _, err := exec.LookPath("script"); err != nil {
		t.Skip("the \"script\" utility needed to allocate a controlling terminal for this test is not available")
	}
	platform := scriptPlatform(t)
	version := "9.5.1"

	managerBytes := buildRealManager(t, version)
	archive := buildRealPackageFixture(t, version, platform, managerBytes)
	managerSHA256 := Digest(managerBytes)
	packageSHA256 := Digest(archive)

	hivePath := "/versions/" + version + "/" + platform + "/hive"
	shaPath := hivePath + ".sha256"
	indexPath := "/versions/" + version + "/index.json"
	pkgPath := "/versions/" + version + "/pkg.tar.gz"

	log := &requestLog{}
	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc(hivePath, func(w http.ResponseWriter, r *http.Request) { log.record(r.URL.Path); _, _ = w.Write(managerBytes) })
	mux.HandleFunc(shaPath, func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		fmt.Fprintf(w, "%s\n", managerSHA256)
	})
	mux.HandleFunc(indexPath, func(w http.ResponseWriter, r *http.Request) {
		log.record(r.URL.Path)
		index := DownloadIndex{Version: 1, ProductVersion: version, Releases: []DownloadRelease{{
			Platform: strings.Replace(platform, "-", "/", 1), SourceID: strings.Repeat("a", 64),
			Package: strings.TrimPrefix(pkgPath, "/"), PackageSHA256: packageSHA256,
			RawBinary: strings.TrimPrefix(hivePath, "/"), RawBinarySHA256: managerSHA256,
		}}}
		_ = json.NewEncoder(w).Encode(index)
	})
	mux.HandleFunc(pkgPath, func(w http.ResponseWriter, r *http.Request) { log.record(r.URL.Path); _, _ = w.Write(archive) })
	srv = httptest.NewServer(mux)
	defer srv.Close()

	home := t.TempDir()
	tmp := t.TempDir()
	env := bootstrapEnv(home, tmp)
	scriptPath := testBootstrapScript(t, t.TempDir(), srv.URL)
	shellCommand := fmt.Sprintf("cat %s | sh -s -- --version %s", shellQuote(scriptPath), version)

	// script's own stdin is left at its default (empty/closed), so the real
	// manager's first read of the controlling terminal (choosing hosts) sees
	// an immediate, genuine EOF and cancels on its own — no answer is ever
	// sent, exactly like TestInstallConfirmationBoundaries' "eof" case.
	out, exitCode := runUnderControllingTTY(t, env, shellCommand)
	if exitCode != 0 {
		t.Fatalf("expected a clean EOF-driven cancellation (exit 0), got %d; output:\n%s", exitCode, out)
	}
	if !strings.Contains(out, "Select CLI hosts") {
		t.Fatalf("the online hand-off never reached the real interactive installer; output:\n%s", out)
	}
	if !strings.Contains(out, "Cancelled. No changes applied.") {
		t.Fatalf("expected the trailing EOF to cancel rather than apply; output:\n%s", out)
	}
	if !log.has(indexPath) || !log.has(pkgPath) {
		t.Fatalf("expected the real manager to download the index and package itself, got requests: %v", log.snapshot())
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cancelling before consent must leave HOME untouched: %v %v", entries, err)
	}
}

func TestInstallOfflinePathMakesNoNetworkRequests(t *testing.T) {
	real := readRepoFile(t, "install.sh")

	// RED demonstration: install.sh currently has no network call at all, so
	// there is no existing behavior to turn RED by testing the shipped
	// script. Instead this proves the interceptor itself would catch a
	// regression, by injecting a synthetic network call into a scratch
	// mutant and confirming the harness observes it.
	t.Run("red_demonstration_interceptor_catches_an_injected_network_call", func(t *testing.T) {
		marker := []byte("curl --silent http://127.0.0.1:1 >/dev/null 2>&1 || true\n")
		anchor := []byte("exec \"$package_dir/bin/hive\"")
		idx := bytes.Index(real, anchor)
		if idx < 0 {
			t.Fatal("could not locate install.sh's final exec line to build the RED mutant; install.sh's shape changed")
		}
		mutant := append([]byte(nil), real[:idx]...)
		mutant = append(mutant, marker...)
		mutant = append(mutant, real[idx:]...)

		_, hits, out := runInstallOffline(t, mutant)
		if hits == 0 {
			t.Fatalf("RED demonstration failed: the network interceptor did not observe the injected curl call; output:\n%s", out)
		}
	})

	t.Run("green_real_install_sh_makes_no_requests", func(t *testing.T) {
		exitCode, hits, out := runInstallOffline(t, real)
		if exitCode != 0 {
			t.Fatalf("install.sh exited %d, want 0; output:\n%s", exitCode, out)
		}
		if hits != 0 {
			t.Fatalf("install.sh (or its delegated manager stub) made %d network-tool invocation(s), want 0; output:\n%s", hits, out)
		}
	})
}
