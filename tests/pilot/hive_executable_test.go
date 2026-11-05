package main

import (
	"path/filepath"
	"runtime"

	"tricell-hive/tooling/management"
)

// Every user-scope install registers `hive hook` with the path of the running
// binary, and the manager refuses the temporary path of a test binary. The pilot
// tests install into scratch homes (also in the subprocess that re-runs the test
// binary), so they point the hook at the checked-in stand-in the management
// tests use.
func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot locate the test source")
	}
	exe, err := filepath.EvalSymlinks(filepath.Join(filepath.Dir(file), "..", "..", "tooling", "management", "testdata", "hive-test-executable"))
	if err != nil {
		panic(err)
	}
	management.HiveExecutable = func() (string, error) { return exe, nil }
}
