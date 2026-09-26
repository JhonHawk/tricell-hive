package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestRegisterCleanupRunsInLIFOOrderAndClearsAfterRunning(t *testing.T) {
	exitCleanups = nil
	defer func() { exitCleanups = nil }()
	var order []int
	registerCleanup(func() { order = append(order, 1) })
	registerCleanup(func() { order = append(order, 2) })
	registerCleanup(func() { order = append(order, 3) })
	runCleanups()
	if got := order; len(got) != 3 || got[0] != 3 || got[1] != 2 || got[2] != 1 {
		t.Fatalf("expected LIFO order [3 2 1], got %v", got)
	}
	order = nil
	runCleanups()
	if len(order) != 0 {
		t.Fatalf("a second run after the list is cleared must be a no-op, got %v", order)
	}
}

// TestInstallSignalCleanupRunsCleanupsAndExitsOnSignal proves the SIGINT/
// SIGTERM wiring in isolation, without sending this test process a real
// signal: installSignalCleanup takes its channel and exit function as
// parameters precisely so this is observable.
func TestInstallSignalCleanupRunsCleanupsAndExitsOnSignal(t *testing.T) {
	sig := make(chan os.Signal, 1)
	var mu sync.Mutex
	ran := false
	exited := make(chan int, 1)
	installSignalCleanup(sig, func() { mu.Lock(); ran = true; mu.Unlock() }, func(code int) { exited <- code })
	sig <- syscall.SIGTERM
	select {
	case code := <-exited:
		if want := 128 + int(syscall.SIGTERM); code != want {
			t.Fatalf("exit code = %d, want %d", code, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for the signal handler to run and exit")
	}
	mu.Lock()
	defer mu.Unlock()
	if !ran {
		t.Fatal("expected registered cleanups to run before exit")
	}
}

// TestMustRunsRegisteredCleanupsBeforeExit proves must() itself (not just the
// signal path) runs every registered cleanup before os.Exit(1). must() calls
// os.Exit directly, so this re-executes the test binary as a subprocess (the
// standard Go pattern for asserting behavior around os.Exit) rather than
// trying to observe it in-process.
func TestMustRunsRegisteredCleanupsBeforeExit(t *testing.T) {
	if os.Getenv("PILOT_MUST_CRASH_MARKER_TEST") == "1" {
		marker := os.Getenv("PILOT_MUST_CRASH_MARKER")
		registerCleanup(func() { _ = os.WriteFile(marker, []byte("cleaned"), 0600) })
		must(errors.New("boom"))
		return
	}
	dir := t.TempDir()
	marker := filepath.Join(dir, "marker")
	cmd := exec.Command(os.Args[0], "-test.run=^TestMustRunsRegisteredCleanupsBeforeExit$")
	cmd.Env = append(os.Environ(), "PILOT_MUST_CRASH_MARKER_TEST=1", "PILOT_MUST_CRASH_MARKER="+marker)
	out, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected the crasher subprocess to exit 1, got err=%v output=%s", err, out)
	}
	b, readErr := os.ReadFile(marker)
	if readErr != nil || string(b) != "cleaned" {
		t.Fatalf("expected the registered cleanup to have run before must() exited: err=%v content=%q", readErr, b)
	}
}

// TestMustCleansUpGuidanceAuthOnFailure is finding 1's real scenario
// end-to-end: main() registers guidance.cleanup right after
// setupGuidanceVariant succeeds (see main.go), so a must() failure anywhere
// between that point and the happy-path guidance.cleanup() call must still
// remove the shadow CODEX_HOME/auth.json — whether it is still the symlink
// linkCodexAuth created, or a plain file Codex renewed the token into. Each
// subtest re-executes the test binary (must() calls os.Exit) so the crash is
// observed as a real process exit, not simulated in-process.
func TestMustCleansUpGuidanceAuthOnFailure(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup func(t *testing.T, authPath string)
	}{
		{
			name: "still-a-symlink",
			setup: func(t *testing.T, authPath string) {
				real := filepath.Join(t.TempDir(), "auth.json")
				if err := os.WriteFile(real, []byte(`{"token":"real"}`), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(real, authPath); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "renewed-plain-file",
			setup: func(t *testing.T, authPath string) {
				if err := os.WriteFile(authPath, []byte(`{"token":"renewed-secret"}`), 0600); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if os.Getenv("PILOT_MUST_CRASH_AUTH_TEST") == "1" {
				g := &guidanceVariant{authPath: os.Getenv("PILOT_MUST_CRASH_AUTH_PATH")}
				registerCleanup(g.cleanup)
				must(errors.New("boom"))
				return
			}
			dir := t.TempDir()
			authPath := filepath.Join(dir, "auth.json")
			tc.setup(t, authPath)
			cmd := exec.Command(os.Args[0], "-test.run=^TestMustCleansUpGuidanceAuthOnFailure$/^"+tc.name+"$")
			cmd.Env = append(os.Environ(), "PILOT_MUST_CRASH_AUTH_TEST=1", "PILOT_MUST_CRASH_AUTH_PATH="+authPath)
			out, err := cmd.CombinedOutput()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected the crasher subprocess to exit 1, got err=%v output=%s", err, out)
			}
			if _, statErr := os.Lstat(authPath); !os.IsNotExist(statErr) {
				t.Fatalf("expected a must() failure after setup to remove the shadow auth file, got err=%v", statErr)
			}
		})
	}
}
