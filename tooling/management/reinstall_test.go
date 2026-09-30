package management

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/integrations/target"
)

// These tests cover managed content that the user deleted by hand: install
// and update write it again as a fresh installation would, remove completes
// without writing, and content that was edited rather than deleted is still
// refused (see the existing TestOwned* and TestManagedDriftAndStalePlans).

func noDrift(t *testing.T, o Options) {
	t.Helper()
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, en := range entries {
		if en.Status == "drift" {
			t.Fatalf("unexpected drift: %+v", en)
		}
	}
}

func removeHiveBlock(t *testing.T, path string) {
	t.Helper()
	data := []byte(get(t, path))
	a, b, err := blockRange(data, hiveMarkers)
	if err != nil || a < 0 {
		t.Fatalf("no Hive block in %s: %v", path, err)
	}
	put(t, path, string(append(append([]byte{}, data[:a]...), data[b:]...)))
}

func TestReinstallRestoresDeletedSkillAndLink(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	wantBytes := get(t, sharedPath(o))
	info, err := os.Stat(sharedPath(o))
	if err != nil {
		t.Fatal(err)
	}
	wantLink, err := os.Readlink(aliasPath(o))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{sharedPath(o), aliasPath(o)} {
		if err = os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	p := plan(t, "install", o)
	gone := 0
	for _, ch := range p.Changes {
		if ch.Gone {
			gone++
		}
	}
	if gone != 2 {
		t.Fatalf("want 2 gone changes, got %d", gone)
	}
	if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
		t.Fatalf("a reinstall must not count as unchanged: %v %v", unchanged, err)
	}
	if apply(t, p) == "unchanged" {
		t.Fatal("reinstall reported unchanged")
	}
	if get(t, sharedPath(o)) != wantBytes {
		t.Fatal("skill bytes differ")
	}
	if got, err := os.Stat(sharedPath(o)); err != nil || got.Mode().Perm() != info.Mode().Perm() {
		t.Fatalf("skill mode differs: %v %v", got, err)
	}
	if link, err := os.Readlink(aliasPath(o)); err != nil || link != wantLink {
		t.Fatalf("link: %q %v", link, err)
	}
	noDrift(t, o)
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("second install must be unchanged")
	}
}

func TestReinstallRestoresDeletedBlockKeepingUserText(t *testing.T) {
	o := setup(t)
	put(t, codexPath(o), "user before\n")
	apply(t, plan(t, "install", o))
	removeHiveBlock(t, codexPath(o))
	if get(t, codexPath(o)) != "user before\n" {
		t.Fatalf("setup: %q", get(t, codexPath(o)))
	}
	p := plan(t, "install", o)
	apply(t, p)
	got := get(t, codexPath(o))
	if !strings.HasPrefix(got, "user before\n") || !strings.Contains(got, Begin) {
		t.Fatalf("block not restored with user text kept: %q", got)
	}
	noDrift(t, o)
}

func TestReinstallRestoresBlockWhenUserTextLacksTrailingNewline(t *testing.T) {
	o := setup(t)
	put(t, codexPath(o), "no newline")
	apply(t, plan(t, "install", o))
	removeHiveBlock(t, codexPath(o))
	put(t, codexPath(o), "no newline")
	apply(t, plan(t, "install", o))
	if got := get(t, codexPath(o)); !strings.HasPrefix(got, "no newline\n"+Begin) {
		t.Fatalf("leading separator missing: %q", got)
	}
	noDrift(t, o)
	remove := o
	apply(t, plan(t, "remove", remove))
	if get(t, codexPath(o)) != "no newline" {
		t.Fatalf("remove did not restore the user text: %q", get(t, codexPath(o)))
	}
}

func TestReinstallCreatesDeletedFileThatHeldTheBlock(t *testing.T) {
	o := setup(t)
	put(t, codexPath(o), "user text\n")
	apply(t, plan(t, "install", o))
	for _, p := range []string{codexPath(o), claudePath(o)} {
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	apply(t, plan(t, "install", o))
	for _, p := range []string{codexPath(o), claudePath(o)} {
		if !strings.Contains(get(t, p), Begin) {
			t.Fatalf("%s not recreated with the block", p)
		}
	}
	st := stateFor(t, o)
	if !st.Records[codexPath(o)].CreatedFile || !st.Records[claudePath(o)].CreatedFile {
		t.Fatal("a recreated file must be recorded as created by Hive")
	}
	noDrift(t, o)
	// Removing now deletes the files Hive recreated.
	apply(t, plan(t, "remove", o))
	absent(t, codexPath(o))
	absent(t, claudePath(o))
}

func TestReinstallKeepsSharedConsumersAndCreatedDirsWithoutDuplicates(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	before := stateFor(t, o)
	if err := os.Remove(sharedPath(o)); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	after := stateFor(t, o)
	if len(after.Records[sharedPath(o)].Consumers) != len(before.Records[sharedPath(o)].Consumers) {
		t.Fatal("consumers lost on reinstall")
	}
	seen := map[string]bool{}
	for _, d := range after.CreatedDirs {
		if seen[d] {
			t.Fatalf("duplicate created dir %s", d)
		}
		seen[d] = true
	}
}

func TestRemoveOfDeletedContentCompletesWithoutWriting(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	for _, p := range []string{sharedPath(o), aliasPath(o), codexPath(o)} {
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	removeHiveBlock(t, claudePath(o))
	put(t, claudePath(o), "kept by the user\n")
	apply(t, plan(t, "remove", o))
	if len(stateFor(t, o).Records) != 0 {
		t.Fatal("records not released")
	}
	absent(t, sharedPath(o))
	absent(t, codexPath(o))
	if get(t, claudePath(o)) != "kept by the user\n" {
		t.Fatal("user text changed")
	}
}

func TestRemoveOfOneHostDoesNotRestoreAMissingSharedSkill(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	if err := os.Remove(sharedPath(o)); err != nil {
		t.Fatal(err)
	}
	partial := o
	partial.Hosts = []string{"claude"}
	p := plan(t, "remove", partial)
	apply(t, p)
	absent(t, sharedPath(o))
	if len(stateFor(t, o).Records[sharedPath(o)].Consumers) != 1 {
		t.Fatalf("want the codex consumer kept: %+v", stateFor(t, o).Records[sharedPath(o)])
	}
	// The next install brings it back.
	apply(t, plan(t, "install", o))
	if get(t, sharedPath(o)) == "" {
		t.Fatal("not restored")
	}
	noDrift(t, o)
}

func TestInterruptedReinstallRecoversToTheDeletedState(t *testing.T) {
	for _, stage := range []string{"prepared", "write:0", "write:1", "write:2", "write:3", "state"} {
		t.Run(stage, func(t *testing.T) {
			o := setup(t)
			put(t, codexPath(o), "user text\n")
			apply(t, plan(t, "install", o))
			for _, p := range []string{sharedPath(o), aliasPath(o), claudePath(o)} {
				if err := os.Remove(p); err != nil {
					t.Fatal(err)
				}
			}
			removeHiveBlock(t, codexPath(o))
			stateBefore := get(t, filepath.Join(o.StateDir, "state.json"))
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
			if _, err := (Engine{}).Recover(o.StateDir); err != nil {
				t.Fatal(err)
			}
			if get(t, filepath.Join(o.StateDir, "state.json")) != stateBefore {
				t.Fatal("state not restored")
			}
			absent(t, sharedPath(o))
			absent(t, aliasPath(o))
			absent(t, claudePath(o))
			if get(t, codexPath(o)) != "user text\n" {
				t.Fatalf("user file not restored: %q", get(t, codexPath(o)))
			}
			apply(t, plan(t, "install", o))
			noDrift(t, o)
		})
	}
}

func TestForgedGoneFailsClosedWhenTheContentIsStillThere(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	p := plan(t, "remove", o)
	for i := range p.Changes {
		p.Changes[i].Gone = true
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("a forged Gone was accepted")
	}
	for _, path := range []string{sharedPath(o), codexPath(o), claudePath(o)} {
		if _, err := os.Lstat(path); err != nil {
			t.Fatalf("%s was touched: %v", path, err)
		}
	}
	if len(stateFor(t, o).Records) == 0 {
		t.Fatal("records dropped by a forged Gone")
	}
	if ok, err := PlanUnchanged(p); ok || err == nil {
		t.Fatalf("PlanUnchanged accepted a forged Gone: %v %v", ok, err)
	}
}

func TestTransformChangeGone(t *testing.T) {
	const path = "/synthetic/home/.agents/skills/demo/SKILL.md"
	rec := skillRecord(path, "skill", []byte("managed"), 0600)
	other := Record{Target: target.Target{Path: path, Kind: "skill"}, Managed: []byte("other"), Mode: 0600, Consumers: []Consumer{{Host: "codex"}}}
	present := snapshot{Exists: true, Data: []byte("managed"), Mode: 0600}
	t.Run("remove never writes, whatever After says", func(t *testing.T) {
		got, err := transformChange(snapshot{}, Change{Target: rec.Target, Before: &rec, After: &other, Gone: true}, "remove", hiveMarkers)
		if err != nil || got.Exists {
			t.Fatalf("got %+v %v", got, err)
		}
	})
	t.Run("install writes as a fresh install", func(t *testing.T) {
		got, err := transformChange(snapshot{}, Change{Target: rec.Target, Before: &rec, After: &rec, Gone: true}, "install", hiveMarkers)
		if err != nil || !got.Exists || !bytes.Equal(got.Data, []byte("managed")) || got.Mode != 0600 {
			t.Fatalf("got %+v %v", got, err)
		}
	})
	t.Run("a retired resource writes nothing", func(t *testing.T) {
		got, err := transformChange(snapshot{}, Change{Target: rec.Target, Before: &rec, Gone: true}, "install", hiveMarkers)
		if err != nil || got.Exists {
			t.Fatalf("got %+v %v", got, err)
		}
	})
	t.Run("present content fails closed", func(t *testing.T) {
		for _, action := range []string{"install", "remove"} {
			if _, err := transformChange(present, Change{Target: rec.Target, Before: &rec, After: &rec, Gone: true}, action, hiveMarkers); err == nil {
				t.Fatalf("%s accepted Gone over present content", action)
			}
		}
	})
	t.Run("edited content fails closed with the ownership error", func(t *testing.T) {
		edited := snapshot{Exists: true, Data: []byte("edited"), Mode: 0600}
		_, err := transformChange(edited, Change{Target: rec.Target, Before: &rec, After: &rec, Gone: true}, "install", hiveMarkers)
		var changed *ManagedFileChangedError
		if !errors.As(err, &changed) || changed.Kind != ManagedFileChanged {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("no record fails closed", func(t *testing.T) {
		if _, err := transformChange(snapshot{}, Change{Target: rec.Target, After: &rec, Gone: true}, "install", hiveMarkers); err == nil {
			t.Fatal("Gone without a Before record was accepted")
		}
	})
	t.Run("without Gone behaves like transform", func(t *testing.T) {
		if _, err := transformChange(snapshot{}, Change{Target: rec.Target, Before: &rec, After: &rec}, "install", hiveMarkers); err == nil {
			t.Fatal("a missing file without Gone must still be refused")
		}
	})
}

func TestEditedContentIsStillRefusedNextToDeletedContent(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	if err := os.Remove(sharedPath(o)); err != nil {
		t.Fatal(err)
	}
	put(t, claudePath(o), Begin+"\nedited by the user\n"+End+"\n")
	for _, action := range []string{"install", "remove"} {
		if _, err := BuildPlan(action, o); err == nil {
			t.Fatalf("%s accepted an edited file", action)
		}
	}
}

func TestReinstallRecreatesAWholeDeletedSkillDirectory(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	want := get(t, sharedPath(o))
	agentsDir := filepath.Join(o.Home, ".agents")
	before := stateFor(t, o).CreatedDirs
	recorded := false
	for _, d := range before {
		recorded = recorded || strings.HasPrefix(d, agentsDir)
	}
	if !recorded {
		t.Fatalf("setup: Hive should have recorded creating directories under %s: %v", agentsDir, before)
	}
	if err := os.RemoveAll(agentsDir); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if get(t, sharedPath(o)) != want {
		t.Fatal("skill not recreated with its bytes")
	}
	seen := map[string]bool{}
	for _, d := range stateFor(t, o).CreatedDirs {
		if seen[d] {
			t.Fatalf("duplicate created dir %s in %v", d, stateFor(t, o).CreatedDirs)
		}
		seen[d] = true
	}
	noDrift(t, o)
}

func TestReinstallRestoresADeletedAgentAndStillRefusesAnEditedOne(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude"}
	modelsSource(t, o)
	apply(t, plan(t, "install", o))
	var rec Record
	for _, r := range stateFor(t, o).Records {
		if r.Target.Kind == "agent" {
			rec = r
			break
		}
	}
	if rec.Target.Path == "" {
		t.Fatal("setup: no agent installed")
	}
	want := get(t, rec.Target.Path)
	info, err := os.Stat(rec.Target.Path)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(rec.Target.Path); err != nil {
		t.Fatal(err)
	}
	p := plan(t, "install", o)
	if unchanged, err := PlanUnchanged(p); err != nil || unchanged {
		t.Fatalf("a deleted agent must not count as unchanged: %v %v", unchanged, err)
	}
	apply(t, p)
	if get(t, rec.Target.Path) != want {
		t.Fatal("agent bytes differ")
	}
	if got, err := os.Stat(rec.Target.Path); err != nil || got.Mode().Perm() != info.Mode().Perm() {
		t.Fatalf("agent mode differs: %v %v", got, err)
	}
	noDrift(t, o)

	put(t, rec.Target.Path, want+"edited\n")
	for _, action := range []string{"install", "remove"} {
		if _, err := BuildPlan(action, o); err == nil {
			t.Fatalf("%s accepted an edited agent", action)
		}
	}
	if get(t, rec.Target.Path) != want+"edited\n" {
		t.Fatal("the edit was overwritten")
	}
}

func TestRemoveOfOneHostKeepsWorkingWhenASharedBlockWithASeparatorIsGone(t *testing.T) {
	cases := map[string]func(t *testing.T, o Options){
		"block deleted": func(t *testing.T, o Options) { removeHiveBlock(t, claudePath(o)) },
		"file deleted": func(t *testing.T, o Options) {
			if err := os.Remove(claudePath(o)); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, damage := range cases {
		t.Run(name, func(t *testing.T) {
			o := setup(t)
			o.Hosts = []string{"claude", "grok"}
			put(t, claudePath(o), "no trailing newline")
			apply(t, plan(t, "install", o))
			if stateFor(t, o).Records[claudePath(o)].Leading != "\n" {
				t.Fatal("setup: the record should carry a separator")
			}
			damage(t, o)
			partial := o
			partial.Hosts = []string{"claude"}
			p := plan(t, "remove", partial)
			if _, err := PlanUnchanged(p); err != nil {
				t.Fatalf("PlanUnchanged: %v", err)
			}
			apply(t, p)
			if len(stateFor(t, o).Records[claudePath(o)].Consumers) != 1 {
				t.Fatal("grok must keep the shared record")
			}
		})
	}
}

func TestMarkersDeletedButBodyKeptIsRefusedAsAnEdit(t *testing.T) {
	o := setup(t)
	put(t, codexPath(o), "user text\n")
	apply(t, plan(t, "install", o))
	var kept []string
	for _, line := range strings.Split(get(t, codexPath(o)), "\n") {
		if line != Begin && line != End {
			kept = append(kept, line)
		}
	}
	edited := strings.Join(kept, "\n")
	put(t, codexPath(o), edited)
	for _, action := range []string{"install", "remove"} {
		if _, err := BuildPlan(action, o); err == nil {
			t.Fatalf("%s appended a second copy of the rules", action)
		}
	}
	if get(t, codexPath(o)) != edited {
		t.Fatal("the file changed")
	}
}

func TestRemoveDeletesTheEmptyFileHiveCreatedWhenOnlyVoiceRemained(t *testing.T) {
	o := voiceInstalled(t)
	removeHiveBlock(t, codexPath(o))
	apply(t, plan(t, "remove", o))
	absent(t, codexPath(o))
}

func TestReinstallPutsTheHiveBlockBeforeTheVoiceBlock(t *testing.T) {
	for _, original := range []string{"", "user text\n", "no trailing newline"} {
		o := setup(t)
		voiceSource(t, o)
		if original != "" {
			put(t, codexPath(o), original)
		}
		apply(t, plan(t, "install", o))
		apply(t, voicePlan(t, "set", o, jarvisSirSubtle))
		want := get(t, codexPath(o))
		removeHiveBlock(t, codexPath(o))
		apply(t, plan(t, "install", o))
		if got := get(t, codexPath(o)); got != want {
			t.Fatalf("original %q: not identical to a fresh install:\n%q\nwant\n%q", original, got, want)
		}
		noDrift(t, o)
	}
}
