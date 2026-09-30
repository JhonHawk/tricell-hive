package management

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cursorHome(o Options) string      { return filepath.Join(o.Home, ".cursor") }
func cursorBlockPath(o Options) string { return filepath.Join(cursorHome(o), "AGENTS.md") }
func cursorAgentsDir(o Options) string { return filepath.Join(cursorHome(o), "agents") }
func cursorAgentPath(o Options, role string) string {
	return filepath.Join(cursorAgentsDir(o), role+".md")
}

func TestCursorLifecycleInstallUpdateRemoveWithSharedConsumer(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "cursor"}
	p := plan(t, "install", o)
	apply(t, p)

	if !strings.Contains(get(t, cursorBlockPath(o)), Begin) {
		t.Fatal("Cursor block not written")
	}
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 2 {
		t.Fatal("Cursor did not register as a shared-skill consumer")
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("reinstall was not idempotent")
	}

	put(t, filepath.Join(o.Source, GlobalSource), "# New rules\nUpdated content.\n")
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, cursorBlockPath(o)), "Updated content.") {
		t.Fatal("Cursor block was not updated")
	}

	partial := o
	partial.Hosts = []string{"cursor"}
	apply(t, plan(t, "remove", partial))
	absent(t, cursorBlockPath(o))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
		t.Fatal("removing Cursor also removed Codex's shared-skill consumer registration")
	}
	if !strings.Contains(get(t, filepath.Join(o.Home, ".codex", "AGENTS.md")), Begin) {
		t.Fatal("Codex block was removed by Cursor's own removal")
	}
}

func TestCursorSharedSkillUpdateRequiresAllConsumers(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "cursor"}
	apply(t, plan(t, "install", o))
	before := get(t, sharedPath(o))

	put(t, filepath.Join(o.Source, SkillSource), before+"New shared instruction.\n")
	solo := o
	solo.Hosts = []string{"cursor"}
	if _, err := BuildPlan("install", solo); err == nil || !strings.Contains(err.Error(), "all consumers") {
		t.Fatalf("partial shared-skill update accepted: %v", err)
	}
	if get(t, sharedPath(o)) != before {
		t.Fatal("conflicting plan wrote payload")
	}

	o.Hosts = []string{"codex", "cursor"}
	apply(t, plan(t, "install", o))
	if !strings.Contains(get(t, sharedPath(o)), "New shared instruction") {
		t.Fatal("complete update failed")
	}
}

func TestCursorHomeAbsentIsCreatedAndCleanedUp(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	absent(t, cursorHome(o))
	apply(t, plan(t, "install", o))
	if info, err := os.Stat(cursorHome(o)); err != nil || !info.IsDir() {
		t.Fatal("Cursor home was not created")
	}
	apply(t, plan(t, "remove", o))
	absent(t, cursorHome(o))
}

func TestCursorHomePreexistingForeignFilesArePreserved(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	foreign := filepath.Join(cursorHome(o), "cli-config.json")
	put(t, foreign, "{\"version\":1}")
	put(t, filepath.Join(cursorHome(o), "rules", "team.mdc"), "---\nalwaysApply: true\n---\nTeam rule.\n")

	apply(t, plan(t, "install", o))
	if get(t, foreign) != "{\"version\":1}" {
		t.Fatal("foreign Cursor file was changed")
	}
	if !strings.Contains(get(t, cursorBlockPath(o)), Begin) {
		t.Fatal("Cursor block was not written alongside foreign files")
	}

	apply(t, plan(t, "remove", o))
	absent(t, cursorBlockPath(o))
	if get(t, foreign) != "{\"version\":1}" {
		t.Fatal("foreign Cursor file lost on removal")
	}
	if info, err := os.Stat(cursorHome(o)); err != nil || !info.IsDir() {
		t.Fatal("foreign-owned Cursor home was removed")
	}
}

func TestCursorForeignAGENTSTextIsPreserved(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	original := "User-written Cursor notes.\n"
	put(t, cursorBlockPath(o), original)
	apply(t, plan(t, "install", o))
	if !strings.HasPrefix(get(t, cursorBlockPath(o)), original) {
		t.Fatal("foreign AGENTS.md text was not preserved")
	}
	if !strings.Contains(get(t, cursorBlockPath(o)), Begin) {
		t.Fatal("Hive block missing after preserving foreign text")
	}
}

func cursorProfilesAndRole(t *testing.T, o Options) (source string) {
	t.Helper()
	repoProfiles, err := os.ReadFile(syntheticProfiles)
	if err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, "integrations", "agent-profiles.json"), string(repoProfiles))
	source = "content/agents/design/test-agent.md"
	put(t, filepath.Join(o.Source, source), "---\nname: test-agent\ndescription: Test role\nmodel_profile: execution\naccess_profile: observe\n---\nUse evidence.\n")
	return source
}

func TestCursorCatalogueRendersAgentFile(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	source := cursorProfilesAndRole(t, o)
	p := plan(t, "install", o)
	apply(t, p)

	rendered := get(t, cursorAgentPath(o, "test-agent"))
	if !strings.Contains(rendered, `"test-agent"`) || !strings.Contains(rendered, "readonly: true") {
		t.Fatalf("Cursor role not rendered as expected:\n%s", rendered)
	}
	var found *Record
	for _, ch := range p.Changes {
		if ch.Target.Source == source && ch.Target.Kind == "agent" {
			found = ch.After
		}
	}
	if found == nil || string(found.Managed) != rendered {
		t.Fatal("frozen plan payload does not match written bytes")
	}
}

func TestCursorForeignAgentFileConflicts(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	cursorProfilesAndRole(t, o)
	put(t, cursorAgentPath(o, "test-agent"), "foreign role content")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("foreign agent file at a role path was silently adopted")
	}
	if get(t, cursorAgentPath(o, "test-agent")) != "foreign role content" {
		t.Fatal("foreign agent file was changed despite the conflict")
	}
}

func TestCursorStatusInstalledRetainedAndNotInstalled(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"codex", "cursor"}
	apply(t, plan(t, "install", o))

	only := o
	only.Hosts = []string{"cursor"}
	entries, err := Status(only)
	if err != nil {
		t.Fatal(err)
	}
	statusFor := func(entries []StatusEntry, path string) string {
		for _, en := range entries {
			if en.Path == path {
				return en.Status
			}
		}
		return ""
	}
	if statusFor(entries, cursorBlockPath(o)) != "installed" {
		t.Fatalf("Cursor block status: %+v", entries)
	}
	if statusFor(entries, sharedPath(o)) != "installed" {
		t.Fatalf("Cursor shared-skill status: %+v", entries)
	}

	apply(t, plan(t, "remove", only))
	entries, err = Status(only)
	if err != nil {
		t.Fatal(err)
	}
	if statusFor(entries, cursorBlockPath(o)) != "not_installed" {
		t.Fatalf("Cursor block status after removal: %+v", entries)
	}
	if statusFor(entries, sharedPath(o)) != "retained_shared" {
		t.Fatalf("shared skill status after partial removal: %+v", entries)
	}
}

func TestCursorRejectsProjectScope(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"cursor"}
	o.Scope = "project"
	o.Root = filepath.Join(o.Home, "project")
	os.MkdirAll(o.Root, 0700)
	if _, err := BuildPlan("install", o); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("Cursor project scope should be rejected with a clear adapter error, got %v", err)
	}
}

func TestCursorRecoveryAtWriteAndStateBoundaries(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"cursor"}
			put(t, cursorBlockPath(o), "original\n")
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
			if get(t, cursorBlockPath(o)) != "original\n" {
				t.Fatal("recovery changed the Cursor instruction file")
			}
			absent(t, sharedPath(o))
			apply(t, plan(t, "install", o))
			if !strings.Contains(get(t, cursorBlockPath(o)), Begin) {
				t.Fatal("install after recovery did not write the Cursor block")
			}
		})
	}
}

func TestPlanSavedBeforeCursorSupportAsksToRegenerate(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.Config.CursorHome = ""
	p.ID = planID(p)
	if err := validatePlan(p, emptyState()); err == nil || !strings.Contains(err.Error(), "regenerate") {
		t.Fatalf("legacy plan error = %v", err)
	}
}
