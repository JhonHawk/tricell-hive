package distribution

// This file runs install.sh through a real shell process against a synthetic
// offline package, proving that the offline installation path performs no
// network requests. No test here reaches a real network origin: curl and wget
// are shadowed by recording stand-ins earlier on PATH.

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

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

// scriptPlatformParts mirrors install.sh's own uname
// mapping so fixture paths and the "platform" file match what the script
// under test will actually compute on this machine.
func scriptPlatformParts(t *testing.T) (osName, arch string) {
	t.Helper()
	switch runtime.GOOS {
	case "darwin", "linux":
		osName = runtime.GOOS
	default:
		t.Skipf("install.sh only supports darwin and linux; GOOS=%s", runtime.GOOS)
	}
	switch runtime.GOARCH {
	case "arm64", "amd64":
		arch = runtime.GOARCH
	default:
		t.Skipf("install.sh only supports arm64 and amd64; GOARCH=%s", runtime.GOARCH)
	}
	return osName, arch
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
