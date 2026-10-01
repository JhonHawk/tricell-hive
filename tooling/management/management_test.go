package management

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

func put(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0640); err != nil {
		t.Fatal(err)
	}
}
func get(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
func setup(t *testing.T) Options {
	t.Helper()
	piCalls = nil
	old := applyPiCmd
	applyPiCmd = stubApplyPi
	t.Cleanup(func() { applyPiCmd = old })
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	o := Options{Scope: "user", Home: filepath.Join(base, "home"), StateDir: filepath.Join(base, "state"), Source: filepath.Join(base, "source"), Hosts: []string{"codex", "claude"}}
	if err = os.MkdirAll(o.Home, 0700); err != nil {
		t.Fatal(err)
	}
	put(t, filepath.Join(o.Source, GlobalSource), "# Rules\nKeep user content.\n")
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nPreserve evidence.\n")
	return o
}
func plan(t *testing.T, a string, o Options) Plan {
	t.Helper()
	p, err := BuildPlan(a, o)
	if err != nil {
		t.Fatal(err)
	}
	return p
}
func apply(t *testing.T, p Plan) string {
	t.Helper()
	s, err := (Engine{}).Apply(p)
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func absent(t *testing.T, p string) {
	t.Helper()
	if _, err := os.Lstat(p); !os.IsNotExist(err) {
		t.Fatalf("expected absent %s: %v", p, err)
	}
}
func TestLifecyclePreservesUserBytesModesAndExtraFiles(t *testing.T) {
	for _, ending := range []string{"\n", "\r\n", ""} {
		t.Run(strings.ReplaceAll(ending, "\n", "LF"), func(t *testing.T) {
			o := setup(t)
			cp := filepath.Join(o.Home, ".codex", "AGENTS.md")
			cl := filepath.Join(o.Home, ".claude", "CLAUDE.md")
			original := "user instructions" + ending
			put(t, cp, original)
			p := plan(t, "install", o)
			apply(t, p)
			if !strings.HasPrefix(get(t, cp), original) {
				t.Fatal("lost prefix")
			}
			if !strings.HasSuffix(get(t, cl), End+"\n") {
				t.Fatal("no block")
			}
			if got := apply(t, plan(t, "install", o)); got != "unchanged" {
				t.Fatal(got)
			}
			current := get(t, cp)
			put(t, cp, current+"later user text\n")
			put(t, filepath.Join(o.Source, GlobalSource), "# New rules\nKeep newer content.\n")
			apply(t, plan(t, "install", o))
			if !strings.HasSuffix(get(t, cp), "later user text\n") {
				t.Fatal("lost suffix")
			}
			info, _ := os.Stat(cp)
			if info.Mode().Perm() != 0640 {
				t.Fatal("mode changed")
			}
			skillExtra := filepath.Join(o.Home, ".agents", "skills", "workspace-conventions", "user.txt")
			put(t, skillExtra, "keep")
			old := o
			old.ReleaseID = p.Release.ID
			apply(t, plan(t, "install", old))
			if !strings.Contains(get(t, cp), "Keep user content.") {
				t.Fatal("old release not restored")
			}
			apply(t, plan(t, "remove", o))
			if get(t, cp) != original+"later user text\n" {
				t.Fatalf("outside bytes changed: %q", get(t, cp))
			}
			absent(t, cl)
			if get(t, skillExtra) != "keep" {
				t.Fatal("extra lost")
			}
			if got := apply(t, plan(t, "remove", o)); got != "unchanged" {
				t.Fatal(got)
			}
		})
	}
}
func TestPlanningDoesNotWriteStateOrTargets(t *testing.T) {
	o := setup(t)
	plan(t, "install", o)
	absent(t, o.StateDir)
	absent(t, filepath.Join(o.Home, ".codex"))
}
func TestSourceAllowlistAndFrozenPlan(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "_support", "private.txt"), "do not distribute")
	p := plan(t, "install", o)
	if len(p.Release.Files) != 2 {
		t.Fatal("allowlist")
	}
	put(t, filepath.Join(o.Source, GlobalSource), "changed after planning\n")
	apply(t, p)
	if strings.Contains(get(t, p.Changes[0].Target.Path), "changed after planning") {
		t.Fatal("plan not frozen")
	}
}
func TestConflictsPreserveAllTargets(t *testing.T) {
	for _, bad := range []string{Begin + "\n", End + "\n" + Begin + "\n", Begin + "\nx\n" + End + "\n", Begin + "\nx\n" + Begin + "\n" + End + "\n", "inline " + Begin + "\nx\n" + End + "\n"} {
		t.Run(bad[:min(len(bad), 20)], func(t *testing.T) {
			o := setup(t)
			p := filepath.Join(o.Home, ".codex", "AGENTS.md")
			put(t, p, bad)
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("accepted markers")
			}
			if get(t, p) != bad {
				t.Fatal("changed conflict")
			}
			absent(t, filepath.Join(o.Home, ".claude", "CLAUDE.md"))
		})
	}
}
func TestManagedDriftAndStalePlans(t *testing.T) {
	t.Run("stale target", func(t *testing.T) {
		o := setup(t)
		p := plan(t, "install", o)
		put(t, p.Changes[0].Target.Path, "late edit")
		if _, err := (Engine{}).Apply(p); err == nil {
			t.Fatal("stale accepted")
		}
		if get(t, p.Changes[0].Target.Path) != "late edit" {
			t.Fatal("overwritten")
		}
	})
	t.Run("stale state", func(t *testing.T) {
		o := setup(t)
		p := plan(t, "install", o)
		apply(t, p)
		if _, err := (Engine{}).Apply(p); err == nil {
			t.Fatal("stale state accepted")
		}
	})
	for _, kind := range []string{"block", "skill"} {
		t.Run(kind, func(t *testing.T) {
			o := setup(t)
			p := plan(t, "install", o)
			apply(t, p)
			for _, ch := range p.Changes {
				if ch.Target.Kind == kind {
					// A block with its markers intact but a changed body is an edit;
					// a file with no markers at all reads as a deleted block.
					put(t, ch.Target.Path, "user changed it")
					if kind == "block" {
						put(t, ch.Target.Path, Begin+"\nuser changed it\n"+End+"\n")
					}
					break
				}
			}
			for _, action := range []string{"install", "remove"} {
				if _, err := BuildPlan(action, o); err == nil {
					t.Fatal("drift accepted")
				}
			}
			s, err := Status(o)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, en := range s {
				found = found || en.Status == "drift"
			}
			if !found {
				t.Fatal("drift not reported")
			}
		})
	}
}
func TestLinksAndUnownedSkill(t *testing.T) {
	for _, mode := range []string{"symlink", "hardlink", "directory"} {
		t.Run(mode, func(t *testing.T) {
			o := setup(t)
			path := filepath.Join(o.Home, ".agents", "skills", "workspace-conventions", "SKILL.md")
			put(t, filepath.Join(o.Home, "owned"), "user")
			os.MkdirAll(filepath.Dir(path), 0700)
			if mode == "symlink" {
				os.Symlink(filepath.Join(o.Home, "owned"), path)
			} else if mode == "hardlink" {
				os.Link(filepath.Join(o.Home, "owned"), path)
			} else {
				put(t, filepath.Join(filepath.Dir(path), "notes.md"), "user")
			}
			if _, err := BuildPlan("install", o); err == nil {
				t.Fatal("collision accepted")
			}
			if get(t, filepath.Join(o.Home, "owned")) != "user" {
				t.Fatal("source altered")
			}
		})
	}
}
func TestHostResolutionAndSharedImports(t *testing.T) {
	o := setup(t)
	t.Setenv("CODEX_HOME", "/unrelated")
	t.Setenv("CLAUDE_CONFIG_DIR", "/unrelated")
	p := plan(t, "install", o)
	for _, ch := range p.Changes {
		if !strings.HasPrefix(ch.Target.Path, o.Home+string(os.PathSeparator)) {
			t.Fatal("synthetic home leaked environment")
		}
	}
	override := filepath.Join(o.Home, ".codex", "AGENTS.override.md")
	put(t, override, "")
	p = plan(t, "install", o)
	for _, ch := range p.Changes {
		if strings.Contains(ch.Target.Path, "override") {
			t.Fatal("empty override selected")
		}
	}
	put(t, override, "override")
	p = plan(t, "install", o)
	found := false
	for _, ch := range p.Changes {
		found = found || ch.Target.Path == override
	}
	if !found {
		t.Fatal("override not selected")
	}
	apply(t, p)
	os.Remove(override)
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("shadowed install accepted")
	}
	o = setup(t)
	o.Scope = "project"
	o.Root = filepath.Join(o.Home, "project")
	os.MkdirAll(o.Root, 0700)
	put(t, filepath.Join(o.Root, "CLAUDE.md"), "@AGENTS.md\n")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("duplicate import accepted")
	}
}
func TestProjectLifecycle(t *testing.T) {
	o := setup(t)
	o.Scope = "project"
	o.Root = filepath.Join(o.Home, "repo")
	os.MkdirAll(o.Root, 0700)
	p := plan(t, "install", o)
	apply(t, p)
	for _, ch := range p.Changes {
		if !strings.HasPrefix(ch.Target.Path, o.Root+"/") {
			t.Fatal("escaped project")
		}
	}
	apply(t, plan(t, "remove", o))
	absent(t, filepath.Join(o.Home, ".codex"))
}
func TestRecoveryAtEveryBoundary(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "write:2", "write:3", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			cp := filepath.Join(o.Home, ".codex", "AGENTS.md")
			put(t, cp, "original\n")
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
			if get(t, cp) != "original\n" {
				t.Fatal("recovery changed bytes")
			}
			absent(t, filepath.Join(o.Home, ".claude", "CLAUDE.md"))
			apply(t, plan(t, "install", o))
		})
	}
}
func TestRecoveryPreservesOutsideEditsAndRejectsInsideEdits(t *testing.T) {
	for _, inside := range []bool{false, true} {
		t.Run(fmtBool(inside), func(t *testing.T) {
			o := setup(t)
			p := plan(t, "install", o)
			en := p.Changes[0]
			e := Engine{failpoint: func(s string) error {
				if s == "write:0" {
					return errors.New("injected")
				}
				return nil
			}}
			e.Apply(p)
			original := get(t, en.Target.Path)
			edited := original + "later user text\n"
			if inside {
				edited = strings.Replace(original, "Keep user content.", "new user rule", 1)
			}
			put(t, en.Target.Path, edited)
			os.Chmod(en.Target.Path, 0600)
			_, err := (Engine{}).Recover(o.StateDir)
			if inside {
				if err == nil {
					t.Fatal("inside edit overwritten")
				}
				if get(t, en.Target.Path) != edited {
					t.Fatal("conflict changed")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if get(t, en.Target.Path) != "later user text\n" {
					t.Fatal("outside edit lost")
				}
			}
		})
	}
}
func fmtBool(b bool) string {
	if b {
		return "inside"
	}
	return "outside"
}
func TestPlanSerializationAndLock(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	out := filepath.Join(o.Home, "plan.json")
	if err := SavePlan(out, p); err != nil {
		t.Fatal(err)
	}
	q, err := LoadPlan(out)
	if err != nil || q.ID != p.ID {
		t.Fatal(err)
	}
	if bytes.Contains([]byte(get(t, out)), []byte("user secret")) {
		t.Fatal("unexpected content")
	}
	if err = SavePlan(out, p); err == nil {
		t.Fatal("overwrote plan")
	}
	b := encode(p)
	var tampered map[string]any
	json.Unmarshal(b, &tampered)
	tampered["action"] = "remove"
	put(t, out, string(encode(tampered)))
	if _, err = LoadPlan(out); err == nil {
		t.Fatal("tamper accepted")
	}
	unlock, err := lock(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if _, err = (Engine{}).Apply(p); err == nil {
		t.Fatal("lock ignored")
	}
}
func TestRehashedPlanMustStillMatchRelease(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.Changes[0].After.Managed = []byte("unexpected payload")
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("inconsistent payload accepted")
	}
	absent(t, p.Changes[0].Target.Path)
}
func TestRemoveRecoveryAndChangedMode(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	p := plan(t, "remove", o)
	e := Engine{failpoint: func(s string) error {
		if s == "write:1" {
			return errors.New("stop")
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("injection")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	for _, ch := range p.Changes {
		if _, err := os.Stat(ch.Target.Path); err != nil {
			t.Fatal(err)
		}
	}
	p = plan(t, "remove", o)
	for _, ch := range p.Changes {
		if ch.Target.Kind != "symlink" {
			os.Chmod(ch.Target.Path, 0644)
			break
		}
	}
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("mode change did not invalidate plan")
	}
}
func TestRecoveryDoesNotOwnConcurrentDirectories(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	external := filepath.Dir(p.Changes[0].Target.Path)
	e := Engine{failpoint: func(s string) error {
		if s == "prepared" {
			if err := os.MkdirAll(external, 0700); err != nil {
				return err
			}
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("expected directory creation conflict")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(external); err != nil || !info.IsDir() {
		t.Fatal("removed concurrently created directory")
	}
}
func TestCorruptRecoveryRecordIsPreserved(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	e := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return errors.New("stop")
		}
		return nil
	}}
	e.Apply(p)
	var pendingRecord pending
	if err := decodeFile(filepath.Join(o.StateDir, "pending.json"), &pendingRecord); err != nil {
		t.Fatal(err)
	}
	jp := filepath.Join(o.StateDir, "transactions", pendingRecord.ID+".json")
	var j journal
	if err := decodeFile(jp, &j); err != nil {
		t.Fatal(err)
	}
	j.Entries[0].Before.Data = []byte("corrupted backup")
	if err := writeJSON(jp, j); err != nil {
		t.Fatal(err)
	}
	before := get(t, p.Changes[0].Target.Path)
	if _, err := (Engine{}).Recover(o.StateDir); err == nil {
		t.Fatal("corrupt journal accepted")
	}
	if get(t, p.Changes[0].Target.Path) != before {
		t.Fatal("corrupt recovery changed target")
	}
}
