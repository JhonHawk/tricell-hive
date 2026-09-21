package management

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const researchSource = "content/skills/flow-research/SKILL.md"

func TestCatalogueInstallRollbackAndSharedRetirement(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "pi", "opencode"}
	old := plan(t, "install", o)
	apply(t, old)
	put(t, filepath.Join(o.Source, researchSource), "---\nname: flow-research\ndescription: Research\n---\nResearch evidence.\n")
	put(t, filepath.Join(o.Source, "content/skills/flow-research/references/evidence.md"), "Evidence guide.\n")
	current := plan(t, "install", o)
	if len(current.Release.Files) != 4 {
		t.Fatal("catalogue not packaged")
	}
	apply(t, current)
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("not idempotent")
	}
	research := filepath.Join(o.Home, ".agents/skills/flow-research/SKILL.md")
	alias := filepath.Join(o.Home, ".claude/skills/flow-research")
	if len(stateFor(t, o).Records[research].Consumers) != 5 {
		t.Fatal("missing consumers")
	}
	if _, err := os.Readlink(alias); err != nil {
		t.Fatal(err)
	}
	partial := o
	partial.Hosts = []string{"claude"}
	partial.ReleaseID = old.Release.ID
	apply(t, plan(t, "install", partial))
	absent(t, alias)
	if len(stateFor(t, o).Records[research].Consumers) != 4 {
		t.Fatal("retired other consumers")
	}
	rollback := o
	rollback.ReleaseID = old.Release.ID
	apply(t, plan(t, "install", rollback))
	absent(t, research)
	absent(t, filepath.Join(filepath.Dir(research), "references/evidence.md"))
	apply(t, plan(t, "install", o))
	// Remove and status consume installed state, even without a source checkout.
	o.Source = filepath.Join(t.TempDir(), "absent")
	entries, err := Status(o)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, en := range entries {
		if en.Path == research && en.Status == "installed" {
			count++
		}
	}
	if count != 5 {
		t.Fatalf("status consumers: %d", count)
	}
	apply(t, plan(t, "remove", o))
	absent(t, research)
	absent(t, alias)
}

func TestCatalogueSourceIdentityAndConflict(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, researchSource), "research payload\n")
	p := plan(t, "install", o)
	for i := range p.Changes {
		if p.Changes[i].Target.Source == researchSource && p.Changes[i].Target.Kind == "skill" {
			p.Changes[i].After.Managed = []byte("unrelated payload\n")
		}
	}
	p.ID = planID(p)
	if _, err := (Engine{}).Apply(p); err == nil {
		t.Fatal("forged resource payload accepted")
	}
	dir := filepath.Join(o.Home, ".agents/skills/flow-research")
	put(t, filepath.Join(dir, "user.md"), "user data")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("unowned directory adopted")
	}
	if get(t, filepath.Join(dir, "user.md")) != "user data" {
		t.Fatal("user file changed")
	}
}

func TestV2StateMigrationAndLegacyJournalRecovery(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	s := stateFor(t, o)
	s.Version = 2
	for path, r := range s.Records {
		r.Target.Source = ""
		s.Records[path] = r
	}
	if err := writeJSON(filepath.Join(o.StateDir, "state.json"), s); err != nil {
		t.Fatal(err)
	}
	apply(t, plan(t, "install", o))
	if stateFor(t, o).Version != 3 {
		t.Fatal("state not migrated")
	}
	// Create an interrupted transaction and encode it using the v2 shape: the
	// optional Source field must remain absent so historical integrity hashes work.
	p := plan(t, "remove", o)
	id, err := (Engine{failpoint: func(stage string) error {
		if stage == "write:0" {
			return errors.New("stop")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("missing failure")
	}
	path := filepath.Join(o.StateDir, "transactions", id+".json")
	var j journal
	if err := decodeFile(path, &j); err != nil {
		t.Fatal(err)
	}
	j.Version = 2
	j.Plan.Version = 2
	clear := func(ch *Change) {
		ch.Target.Source = ""
		if ch.Before != nil {
			ch.Before.Target.Source = ""
		}
		if ch.After != nil {
			ch.After.Target.Source = ""
		}
		if ch.Replaces != nil {
			ch.Replaces.Target.Source = ""
		}
	}
	for i := range j.Plan.Changes {
		clear(&j.Plan.Changes[i])
	}
	for i := range j.Entries {
		clear(&j.Entries[i].Change)
	}
	j.Plan.ID = planID(j.Plan)
	if err := saveJournal(path, j); err != nil {
		t.Fatal(err)
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("recovery did not restore installation")
	}
}

func TestCatalogueInterruptedRollbackRecovery(t *testing.T) {
	o := setup(t)
	old := plan(t, "install", o)
	apply(t, old)
	put(t, filepath.Join(o.Source, researchSource), "research\n")
	apply(t, plan(t, "install", o))
	rollback := o
	rollback.ReleaseID = old.Release.ID
	p := plan(t, "install", rollback)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "state" {
			return errors.New("stop")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("missing failure")
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("rollback recovery lost catalogue")
	}
}

func TestReferenceBundleUpgradeRollbackAndOwnership(t *testing.T) {
	o := setup(t)
	o.Hosts = []string{"claude", "codex", "grok", "pi", "opencode"}
	old := plan(t, "install", o)
	apply(t, old)
	source := "content/skills/workspace-conventions/references/nested/plan-format.md"
	put(t, filepath.Join(o.Source, source), "# Format\nPlan format.\n")
	p := plan(t, "install", o)
	aliases := 0
	for _, ch := range p.Changes {
		if ch.Target.Kind == "symlink" {
			aliases++
		}
	}
	if aliases != 1 {
		t.Fatalf("duplicate bundle aliases: %d", aliases)
	}
	apply(t, p)
	reference := filepath.Join(o.Home, ".agents/skills/workspace-conventions/references/nested/plan-format.md")
	if get(t, reference) != "# Format\nPlan format.\n" {
		t.Fatal("reference payload")
	}
	if apply(t, plan(t, "install", o)) != "unchanged" {
		t.Fatal("reference reinstall")
	}
	partial := o
	partial.Hosts = []string{"claude"}
	partial.ReleaseID = old.Release.ID
	apply(t, plan(t, "install", partial))
	if len(stateFor(t, o).Records[reference].Consumers) != 4 {
		t.Fatal("removed other reference consumers")
	}
	if _, err := os.Readlink(aliasPath(o)); err != nil {
		t.Fatal("retired still-required alias", err)
	}
	// A file outside managed targets survives both update and rollback.
	user := filepath.Join(filepath.Dir(reference), "user-notes.md")
	put(t, user, "user")
	rollback := o
	rollback.ReleaseID = old.Release.ID
	apply(t, plan(t, "install", rollback))
	absent(t, reference)
	if get(t, user) != "user" {
		t.Fatal("user file removed")
	}
	put(t, reference, "user reference")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("unowned reference overwritten")
	}
	if get(t, reference) != "user reference" {
		t.Fatal("conflict changed")
	}
}

func TestReferenceCatalogueRejectsUnsafeSources(t *testing.T) {
	for _, source := range []string{"content/skills/x/references/../x.md", "content/skills/x/references/.hidden.md", "content/skills/x/references/x.txt", "content/skills/x/references//x.md"} {
		if validSource(source) {
			t.Fatalf("unsafe source accepted: %s", source)
		}
	}
	r := Release{Files: []Payload{{GlobalSource, []byte("global")}, {"content/skills/x/references/x.md", []byte("reference")}}}
	r.ID = releaseID(r)
	if validateRelease(r) == nil {
		t.Fatal("orphan reference accepted")
	}
	o := setup(t)
	path := filepath.Join(o.Source, "content/skills/workspace-conventions/references")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(o.Source, GlobalSource), filepath.Join(path, "linked.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("symlink source accepted")
	}
}
