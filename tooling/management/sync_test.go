package management

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
	"tricell-hive/tooling/durable"
)

func TestMain(m *testing.M) {
	durable.DisableForTests()
	// The tests call target.Safe, which Lstats every ancestor of a path, some
	// 700k times in temp homes. cmd/go records each call in a test log and
	// re-verifies it line by line when it looks up a cached result, which cost
	// minutes and gigabytes (#91). With no log file there is nothing to verify,
	// so the package is simply run each time. A missing or renamed internal flag
	// is ignored.
	disableTestLog()
	// The test binary and every t.TempDir() live under the temporary directory,
	// which the Grok hook refuses, so the executable is a checked-in file.
	exe, err := filepath.Abs(filepath.Join("testdata", "hive-test-executable"))
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		panic(err)
	}
	HiveExecutable = func() (string, error) { return exe, nil }
	os.Exit(m.Run())
}

// disableTestLog stops cmd/go from logging file-system calls of this test
// binary, which also keeps its result out of the test cache.
func disableTestLog() {
	flag.Parse()
	if f := flag.Lookup("test.testlogfile"); f != nil {
		_ = f.Value.Set("")
	}
}

// TestTestLogFlagStillExists guards the test-cache opt-out in TestMain: it
// relies on cmd/go's internal -test.testlogfile flag. If Go renames or removes
// it, this fails; revisit the opt-out (see deployment-manager.md, Verification)
// before the packages' cached runs slow down again.
func TestTestLogFlagStillExists(t *testing.T) {
	f := flag.Lookup("test.testlogfile")
	if f == nil {
		t.Fatal("flag test.testlogfile is gone: Go changed its internal flag, so the test-cache opt-out in TestMain no longer works and must be revisited")
	}
	if got := f.Value.String(); got != "" {
		t.Fatalf("test.testlogfile = %q after TestMain, want it cleared by the opt-out", got)
	}
}
