package management

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecoverRemovesTemporaryWritesLeftByAKilledProcess(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	crash := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return os.ErrDeadlineExceeded
		}
		return nil
	}}
	if _, err := crash.Apply(p); err == nil {
		t.Fatal("missing crash")
	}
	// A SIGKILL skips deferred cleanup, leaving write temporaries behind.
	stateOrphan := filepath.Join(o.StateDir, "transactions", ".hive-write-111")
	targetOrphan := filepath.Join(filepath.Dir(p.Changes[0].Target.Path), ".hive-write-222")
	unrelated := filepath.Join(filepath.Dir(p.Changes[0].Target.Path), "user-notes.txt")
	for _, path := range []string{stateOrphan, targetOrphan, unrelated} {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{stateOrphan, targetOrphan} {
		if _, err := os.Lstat(path); !os.IsNotExist(err) {
			t.Fatalf("recovery left a temporary write: %s", path)
		}
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("recovery removed an unrelated file: %v", err)
	}
}
