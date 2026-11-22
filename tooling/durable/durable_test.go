package durable

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// syncDefault records the production default before any test changes it:
// package-level initialization runs before the tests.
var syncDefault = !skipSync.Load()

func TestSyncSwitch(t *testing.T) {
	if !syncDefault {
		t.Fatal("disk sync must default to enabled")
	}
	f, err := os.CreateTemp(t.TempDir(), "sync-*")
	if err != nil {
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { skipSync.Store(false) })
	skipSync.Store(false)
	if err = Sync(f); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("enabled Sync on a closed file = %v, want os.ErrClosed", err)
	}
	skipSync.Store(true)
	if err = Sync(f); err != nil {
		t.Fatalf("disabled Sync = %v, want nil", err)
	}
}

// TestDisablePanicsOutsideTestBinary swaps the package predicate: the running
// binary is a test binary, so only the swap can reach the guard.
func TestDisablePanicsOutsideTestBinary(t *testing.T) {
	prevPredicate := isTestBinary
	prevSkip := skipSync.Load()
	t.Cleanup(func() {
		isTestBinary = prevPredicate
		skipSync.Store(prevSkip)
	})
	isTestBinary = func() bool { return false }
	skipSync.Store(false)
	const want = "DisableForTests called outside a test binary"
	var got any
	func() {
		defer func() { got = recover() }()
		DisableForTests()
	}()
	if got == nil {
		t.Fatal("DisableForTests did not panic outside a test binary")
	}
	if got != want {
		t.Fatalf("panic = %v, want %q", got, want)
	}
	if skipSync.Load() {
		t.Fatal("disk sync was disabled despite the panic")
	}
}

// TestDisableOnlyCalledFromTests keeps the hive binary durable: no non-test Go
// file in the repository may call the test-only switch.
func TestDisableOnlyCalledFromTests(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"integrations", "tests", "tooling"} {
		err = filepath.WalkDir(filepath.Join(root, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if strings.Contains(string(data), "durable.DisableForTests()") {
				t.Errorf("%s calls durable.DisableForTests outside a test file", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
