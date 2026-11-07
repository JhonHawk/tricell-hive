package management

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func antigravityHome(o Options) string { return filepath.Join(o.Home, ".gemini", "config") }
func antigravityLink(o Options, skill string) string {
	return filepath.Join(antigravityHome(o), "skills", skill)
}

func antigravitySetup(t *testing.T) Options {
	t.Helper()
	o := setup(t)
	o.Hosts = []string{"antigravity"}
	put(t, filepath.Join(o.Source, "content/skills/flow-build/SKILL.md"), "---\nname: flow-build\ndescription: Build.\n---\nBuild.\n")
	return o
}

func linkTarget(t *testing.T, path string) string {
	t.Helper()
	dest, err := os.Readlink(path)
	if err != nil {
		t.Fatal(err)
	}
	return dest
}

func TestAntigravityPlanProposesRelativeSkillLinksAndNoGuidanceBlock(t *testing.T) {
	o := antigravitySetup(t)
	p := plan(t, "install", o)
	links := 0
	for _, ch := range p.Changes {
		if ch.Target.Kind == "block" {
			t.Fatalf("T1 must not write a guidance block: %s", ch.Target.Path)
		}
		if ch.Target.Kind == "symlink" && strings.HasPrefix(ch.Target.Path, antigravityHome(o)+string(filepath.Separator)) {
			links++
		}
	}
	if links != 2 {
		t.Fatalf("want one link per skill, got %d", links)
	}
	apply(t, p)
	for _, skill := range []string{"workspace-conventions", "flow-build"} {
		link := antigravityLink(o, skill)
		if dest := linkTarget(t, link); dest != "../../../.agents/skills/"+skill {
			t.Fatalf("link target %q", dest)
		}
		resolved, err := filepath.EvalSymlinks(link)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := filepath.EvalSymlinks(filepath.Join(o.Home, ".agents", "skills", skill))
		if resolved != want {
			t.Fatalf("%s resolves to %s, want %s", skill, resolved, want)
		}
	}
	absent(t, filepath.Join(o.Home, ".gemini", "GEMINI.md"))
}

func TestAntigravityLifecycleInstallUpdateRemoveWithSharedConsumer(t *testing.T) {
	o := antigravitySetup(t)
	o.Hosts = []string{"codex", "antigravity"}
	apply(t, plan(t, "install", o))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 2 {
		t.Fatal("Antigravity did not register as a shared-skill consumer")
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("reinstall was not idempotent")
	}
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nUpdated.\n")
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, sharedPath(o)), "Updated.") {
		t.Fatal("shared skill was not updated")
	}

	partial := o
	partial.Hosts = []string{"antigravity"}
	apply(t, plan(t, "remove", partial))
	absent(t, antigravityLink(o, "workspace-conventions"))
	absent(t, antigravityLink(o, "flow-build"))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
		t.Fatal("removing Antigravity dropped Codex's shared-skill consumer")
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")), Begin) {
		t.Fatal("Codex block was removed")
	}
}

func TestAntigravityForeignSkillLinkKeepsItsTarget(t *testing.T) {
	o := antigravitySetup(t)
	skills := filepath.Join(antigravityHome(o), "skills")
	if err := os.MkdirAll(skills, 0700); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(skills, "firecrawl")
	const dest = "../../../.agents/skills/firecrawl"
	if err := os.Symlink(dest, foreign); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if got := linkTarget(t, foreign); got != dest {
		t.Fatalf("install changed the foreign link to %q", got)
	}
	apply(t, plan(t, "install", o))
	apply(t, plan(t, "remove", o))
	if got := linkTarget(t, foreign); got != dest {
		t.Fatalf("remove changed the foreign link to %q", got)
	}
	absent(t, antigravityLink(o, "workspace-conventions"))
}

func TestAntigravityRejectsProjectScope(t *testing.T) {
	o := antigravitySetup(t)
	o.Scope = "project"
	o.Root = filepath.Join(o.Home, "project")
	os.MkdirAll(o.Root, 0700)
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("project scope should fail with the unsupported-scope error, got %v", err)
	}
	absent(t, antigravityHome(o))
}

func TestAntigravityStatusInstalledRetainedAndNotInstalled(t *testing.T) {
	o := antigravitySetup(t)
	o.Hosts = []string{"codex", "antigravity"}
	apply(t, plan(t, "install", o))
	only := o
	only.Hosts = []string{"antigravity"}
	statusOf := func() map[string]string {
		entries, err := Status(only)
		if err != nil {
			t.Fatal(err)
		}
		m := map[string]string{}
		for _, en := range entries {
			m[en.Path] = en.Status
		}
		return m
	}
	m := statusOf()
	if m[antigravityLink(o, "workspace-conventions")] != "installed" || m[sharedPath(o)] != "installed" {
		t.Fatalf("status after install: %+v", m)
	}
	apply(t, plan(t, "remove", only))
	m = statusOf()
	if m[antigravityLink(o, "workspace-conventions")] != "not_installed" || m[sharedPath(o)] != "retained_shared" {
		t.Fatalf("status after removal: %+v", m)
	}
}

func TestAntigravityHomeAbsentIsCreatedAndCleanedUp(t *testing.T) {
	o := antigravitySetup(t)
	absent(t, antigravityHome(o))
	apply(t, plan(t, "install", o))
	if info, err := os.Stat(antigravityHome(o)); err != nil || !info.IsDir() {
		t.Fatal("Antigravity home was not created")
	}
	apply(t, plan(t, "remove", o))
	absent(t, antigravityLink(o, "workspace-conventions"))
}

func TestAntigravityRecoveryAtWriteAndStateBoundaries(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := antigravitySetup(t)
			p := plan(t, "install", o)
			e := Engine{failpoint: func(s string) error {
				if s == stage {
					return errors.New("injected")
				}
				return nil
			}}
			if _, err := e.Apply(p); err == nil {
				t.Fatal("did not fail")
			}
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("pending operation ignored")
			}
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			absent(t, sharedPath(o))
			absent(t, antigravityLink(o, "workspace-conventions"))
			apply(t, plan(t, "install", o))
			if _, err := os.Lstat(antigravityLink(o, "workspace-conventions")); err != nil {
				t.Fatal("install after recovery did not write the link")
			}
		})
	}
}

func TestPlanSavedBeforeAntigravitySupportAsksToRegenerate(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.Config.AntigravityHome = ""
	p.ID = planID(p)
	if err := validatePlan(p, emptyState()); err == nil || !strings.Contains(err.Error(), "regenerate") {
		t.Fatalf("legacy plan accepted or wrong error: %v", err)
	}
}
