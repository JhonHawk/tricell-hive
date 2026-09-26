package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/legacy"
)

func TestMigrationReceiptAndNoop(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	if p.Version != stateVersion {
		t.Fatalf("version = %d, want 5", p.Version)
	}
	apply(t, p)
	s, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Migrations) != 1 || s.Migrations[0].Transaction == "" {
		t.Fatal("missing committed migration receipt")
	}
	before := get(t, filepath.Join(o.StateDir, "state.json"))
	p = plan(t, "install", o)
	unchanged, err := PlanUnchanged(p)
	if err != nil || !unchanged {
		t.Fatalf("unchanged %v: %v", unchanged, err)
	}
	if apply(t, p) != "unchanged" {
		t.Fatal("repeated install changed")
	}
	if get(t, filepath.Join(o.StateDir, "state.json")) != before {
		t.Fatal("noop wrote state")
	}
	apply(t, plan(t, "remove", o))
	s, _, err = readState(o.StateDir)
	if err != nil || len(s.Migrations) != 1 {
		t.Fatal("remove lost receipt", err)
	}
}

func TestLegacyPlanCannotBeSavedOutsideProtectedState(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
	seedLegacy(t, o)
	p := plan(t, "install", o)
	if len(p.Legacy) == 0 {
		t.Fatal("expected legacy edits")
	}
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := SavePlan(path, p); err == nil {
		t.Fatal("saved plan containing private legacy bytes")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("unsafe plan file exists: %v", err)
	}
}
func TestLinuxStateLocation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "external"))
	got, err := defaultStateDir(home, true, "linux")
	want := filepath.Join(home, ".local/state/tricell-hive")
	if err != nil || got != want {
		t.Fatalf("%s %v", got, err)
	}
	old := filepath.Join(home, "Library/Application Support/tricell-hive")
	if err = os.MkdirAll(old, 0700); err != nil {
		t.Fatal(err)
	}
	got, err = defaultStateDir(home, true, "linux")
	if err != nil || got != old {
		t.Fatalf("%s %v", got, err)
	}
	if err = os.MkdirAll(want, 0700); err != nil {
		t.Fatal(err)
	}
	if _, err = defaultStateDir(home, true, "linux"); err == nil {
		t.Fatal("ambiguous state accepted")
	}
}

func seedLegacy(t *testing.T, o Options) {
	t.Helper()
	b, err := os.ReadFile("../legacy/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Files []struct {
			Root, Path string
			Data       []byte
		}
	}
	if err = json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	for _, f := range c.Files {
		root := ""
		switch f.Root {
		case "codex":
			if f.Path == "AGENTS.md" {
				root = filepath.Join(o.Home, ".codex")
			}
		case "claude":
			if f.Path == "CLAUDE.md" || strings.HasPrefix(f.Path, "skills/workspace-conventions/") {
				root = filepath.Join(o.Home, ".claude")
			}
		case "shared":
			if strings.HasPrefix(f.Path, "skills/workspace-conventions/") {
				root = filepath.Join(o.Home, ".agents")
			}
		}
		if root != "" {
			put(t, filepath.Join(root, f.Path), string(f.Data))
		}
	}
}
func TestCatalogMigrationAndRecovery(t *testing.T) {
	for _, stage := range []string{"success", "prepared", "write:0", "write:1", "legacy:remove:SKILL.md", "legacy:directory", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
			seedLegacy(t, o)
			original := get(t, filepath.Join(o.Home, ".codex/AGENTS.md"))
			p := plan(t, "install", o)
			if len(p.Legacy) == 0 {
				t.Fatal("no legacy edits")
			}
			engine := Engine{}
			if stage != "success" {
				engine.failpoint = func(s string) error {
					if s == stage {
						return errors.New("interrupted")
					}
					return nil
				}
			}
			_, err := engine.Apply(p)
			if stage == "success" {
				if err != nil {
					t.Fatal(err)
				}
				if get(t, filepath.Join(o.Home, ".codex/AGENTS.md")) == original {
					t.Fatal("legacy still active")
				}
				if apply(t, plan(t, "install", o)) != "unchanged" {
					t.Fatal("not idempotent")
				}
				return
			}
			if err == nil {
				t.Fatal("failpoint not reached")
			}
			if _, err = (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if got := get(t, filepath.Join(o.Home, ".codex/AGENTS.md")); got != original {
				t.Fatal("legacy not restored")
			}
			p = plan(t, "install", o)
			apply(t, p)
		})
	}
}
func TestForgedLegacyEditRejected(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	path := filepath.Join(o.Home, "personal.txt")
	put(t, path, "personal")
	p.Legacy = []legacy.Edit{{Path: path, Before: []byte("personal"), Delete: true, Mode: 0640}}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("forged deletion accepted")
	}
	if get(t, path) != "personal" {
		t.Fatal("personal file changed")
	}
}

func TestCorruptAndLostStateAreNotFreshInstall(t *testing.T) {
	for _, data := range []string{"{}", "null", `{"version":5}`, `{"version":5,"records":null}`} {
		t.Run(data, func(t *testing.T) {
			o := setup(t)
			put(t, filepath.Join(o.StateDir, "state.json"), data)
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("corrupt state accepted")
			}
		})
	}
	o := setup(t)
	apply(t, plan(t, "install", o))
	if err := os.Remove(filepath.Join(o.StateDir, "state.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("existing rebuild without ownership accepted")
	}
}

func TestMigratedBlockRecoveryPreservesConcurrentEdits(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
	seedLegacy(t, o)
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "state" {
			return errors.New("stop")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("expected interruption")
	}
	path := filepath.Join(o.Home, ".codex/AGENTS.md")
	latest := get(t, path) + "\nnew user decision\n"
	put(t, path, latest)
	if _, err = (Engine{}).Recover(o.StateDir); err == nil {
		t.Fatal("overwrote concurrent migration edit")
	}
	if get(t, path) != latest {
		t.Fatal("concurrent edit was changed")
	}
}
func TestSavedMigrationPlanRevalidatesCatalogInventory(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
	p := plan(t, "install", o)
	seedLegacy(t, o)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("stale inventory accepted")
	}
	absent(t, filepath.Join(o.StateDir, "state.json"))
}
func TestLegacyReintroductionIsDetected(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "opencode", "pi"}
	seedLegacy(t, o)
	path := filepath.Join(o.Home, ".codex/AGENTS.md")
	old := get(t, path)
	apply(t, plan(t, "install", o))
	put(t, path, old)
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("reintroduced legacy accepted as current")
	}
}

func TestMigrationReceiptCoversSelectedSubset(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	p := plan(t, "install", o)
	unchanged, err := PlanUnchanged(p)
	if err != nil || !unchanged || p.Migration != nil {
		t.Fatalf("existing receipt did not cover subset: %v %v", unchanged, err)
	}
}
