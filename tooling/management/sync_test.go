package management

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

// syncDefault records the production default before TestMain disables syncing:
// package-level initialization runs before TestMain.
var syncDefault = !skipDiskSync.Load()

func TestMain(m *testing.M) {
	DisableDiskSyncForTests()
	// The tests call target.Safe, which Lstats every ancestor of a path, some
	// 700k times in temp homes. cmd/go records each call in a test log and
	// re-verifies it line by line when it looks up a cached result, which cost
	// minutes and gigabytes (#91). With no log file there is nothing to verify,
	// so the package is simply run each time. A missing or renamed internal flag
	// is ignored.
	disableTestLog()
	os.Exit(m.Run())
}

func TestDiskSyncSwitch(t *testing.T) {
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
	skipDiskSync.Store(false)
	t.Cleanup(func() { skipDiskSync.Store(true) })
	if err = syncFile(f); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("enabled syncFile on a closed file = %v, want os.ErrClosed", err)
	}
	skipDiskSync.Store(true)
	if err = syncFile(f); err != nil {
		t.Fatalf("disabled syncFile = %v, want nil", err)
	}

	// With syncing enabled, exercise the real write paths once.
	skipDiskSync.Store(false)
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(base, "file.txt")
	if err = write(file, snapshot{Exists: true, Data: []byte("data"), Mode: 0600}); err != nil {
		t.Fatal(err)
	}
	if got := get(t, file); got != "data" {
		t.Fatalf("file content = %q", got)
	}
	link := filepath.Join(base, "link")
	tg := target.Target{Path: link, Kind: "symlink", LinkTarget: file}
	if err = writeResource(tg, snapshot{}, snapshot{Exists: true, Kind: "symlink", LinkTarget: file}, false, nil); err != nil {
		t.Fatal(err)
	}
	if got, err := os.Readlink(link); err != nil || got != file {
		t.Fatalf("Readlink = %q, %v; want %q", got, err, file)
	}
}

// TestDisableDiskSyncPanicsOutsideTestBinary swaps the package predicate:
// the running binary is a test binary, so only the swap can reach the guard.
func TestDisableDiskSyncPanicsOutsideTestBinary(t *testing.T) {
	prevPredicate := isTestBinary
	prevSkip := skipDiskSync.Load()
	t.Cleanup(func() {
		isTestBinary = prevPredicate
		skipDiskSync.Store(prevSkip)
	})
	isTestBinary = func() bool { return false }
	skipDiskSync.Store(false)
	const want = "DisableDiskSyncForTests called outside a test binary"
	var got any
	func() {
		defer func() { got = recover() }()
		DisableDiskSyncForTests()
	}()
	if got == nil {
		t.Fatal("DisableDiskSyncForTests did not panic outside a test binary")
	}
	if got != want {
		t.Fatalf("panic = %v, want %q", got, want)
	}
	if skipDiskSync.Load() {
		t.Fatal("disk sync was disabled despite the panic")
	}
}

// TestDisableDiskSyncOnlyCalledFromTests keeps the hive binary durable: no
// non-test Go file in the repository may call the test-only switch.
func TestDisableDiskSyncOnlyCalledFromTests(t *testing.T) {
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
			if strings.Contains(string(data), "DisableDiskSyncForTests()") && path != filepath.Join(root, "tooling", "management", "files.go") {
				t.Errorf("%s calls DisableDiskSyncForTests outside a test file", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// disableTestLog stops cmd/go from logging file-system calls of this test
// binary, which also keeps its result out of the test cache.
func disableTestLog() {
	flag.Parse()
	if f := flag.Lookup("test.testlogfile"); f != nil {
		_ = f.Value.Set("")
	}
}
