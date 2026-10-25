package management

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// fakeHangingPi puts a `pi` on PATH that starts a long-lived descendant,
// records its PID and then blocks, so a test can check both the returned
// error and that the descendant does not outlive the run.
func fakeHangingPi(t *testing.T) (pidFile string) {
	t.Helper()
	dir := t.TempDir()
	pidFile = filepath.Join(dir, "child.pid")
	script := "#!/bin/sh\nsleep 300 &\necho $! > \"$HIVE_TEST_PID_FILE\"\nwait\n"
	if err := os.WriteFile(filepath.Join(dir, "pi"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HIVE_TEST_PID_FILE", pidFile)
	old := piCommandTimeout
	piCommandTimeout = 300 * time.Millisecond
	t.Cleanup(func() { piCommandTimeout = old })
	return pidFile
}

func waitGone(t *testing.T, pid int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	t.Fatalf("descendant %d of pi outlived the run", pid)
}

func readPID(t *testing.T, path string) int {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fake pi did not start its descendant: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		t.Fatal(err)
	}
	return pid
}

func TestExecApplyPiTimesOutAndKillsDescendants(t *testing.T) {
	pidFile := fakeHangingPi(t)
	done := make(chan error, 1)
	go func() { done <- execApplyPi(t.TempDir(), "install", "npm:pi-subagents") }()
	var err error
	select {
	case err = <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("execApplyPi did not return after its timeout")
	}
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("want a timeout error, got %v", err)
	}
	waitGone(t, readPID(t, pidFile))
}

func TestHungPiInstallLeavesRecoverableJournal(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"pi"}
	pidFile := fakeHangingPi(t)
	applyPiCmd = execApplyPi
	_, err := (Engine{}).Apply(plan(t, "install", o))
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("want a timeout error, got %v", err)
	}
	waitGone(t, readPID(t, pidFile))
	applyPiCmd = stubApplyPi
	piCalls = nil
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatalf("recover after a hung pi: %v", err)
	}
	if len(piCalls) != 0 {
		t.Fatalf("recover invoked pi for a package that was never declared: %#v", piCalls)
	}
}
