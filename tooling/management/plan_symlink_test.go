package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSavePlanAcceptsSymlinkedAncestor is the macOS mktemp case: /var is a
// symlink to /private/var, so a plan path under /var/folders has a symlink
// ancestor. The file must land in the real directory and stay readable
// through the path the caller passed.
func TestSavePlanAcceptsSymlinkedAncestor(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	real := filepath.Join(o.Home, "real-out")
	if err := os.Mkdir(real, 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(o.Home, "linked-out")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(link, "plan.json")
	if err := SavePlan(out, p); err != nil {
		t.Fatal(err)
	}
	saved := filepath.Join(real, "plan.json")
	q, err := LoadPlan(saved)
	if err != nil {
		t.Fatal(err)
	}
	if q.ID != p.ID {
		t.Fatalf("saved plan id %s, want %s", q.ID, p.ID)
	}
	if _, err := LoadPlan(out); err != nil {
		t.Fatalf("plan not readable through the caller path: %v", err)
	}
}

// TestSavePlanRefusesExistingFileThroughSymlink keeps the no-overwrite rule
// when the caller names a symlink to a plan that already exists.
func TestSavePlanRefusesExistingFileThroughSymlink(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	real := filepath.Join(o.Home, "plan.json")
	if err := SavePlan(real, p); err != nil {
		t.Fatal(err)
	}
	before := get(t, real)
	link := filepath.Join(o.Home, "plan-link.json")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	err := SavePlan(link, p)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected an existing-file error, got %v", err)
	}
	if get(t, real) != before {
		t.Fatal("existing plan changed")
	}
}
