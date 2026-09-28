package management

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/integrations/target"
)

func readCommits(t *testing.T, stateDir, releaseID string) []commitLogEntry {
	t.Helper()
	var rec commitRecord
	if err := decodeFile(filepath.Join(stateDir, "releases", releaseID+".commits.json"), &rec); err != nil {
		t.Fatal(err)
	}
	return rec.Commits
}

func TestSourceCommitRecordedOnInstall(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.SourceCommit = strings.Repeat("a", 40)
	p.ID = planID(p)
	apply(t, p)
	commits := readCommits(t, o.StateDir, p.Release.ID)
	if len(commits) != 1 || commits[0].Commit != p.SourceCommit {
		t.Fatalf("expected one recorded commit, got %+v", commits)
	}
	if _, err := time.Parse(time.RFC3339, commits[0].AppliedAt); err != nil {
		t.Fatalf("applied_at is not RFC 3339: %v", err)
	}
}

func TestSourceCommitRecordedWhenApplyReportsUnchanged(t *testing.T) {
	o := setup(t)
	commitA := strings.Repeat("a", 40)
	commitB := strings.Repeat("b", 40)
	p1 := plan(t, "install", o)
	p1.SourceCommit = commitA
	p1.ID = planID(p1)
	apply(t, p1)
	// A second commit whose catalogue content is byte-identical produces the
	// same release ID and no target changes, so Apply reports "unchanged".
	p2 := plan(t, "install", o)
	p2.SourceCommit = commitB
	p2.ID = planID(p2)
	if p2.Release.ID != p1.Release.ID {
		t.Fatal("test setup: expected identical release content")
	}
	result, err := (Engine{}).Apply(p2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(result, "unchanged") {
		t.Fatalf("expected unchanged, got %q", result)
	}
	commits := readCommits(t, o.StateDir, p1.Release.ID)
	if len(commits) != 2 || commits[0].Commit != commitA || commits[1].Commit != commitB {
		t.Fatalf("expected both commits listed once each, got %+v", commits)
	}
}

func TestSourceCommitNotDuplicatedOnReapply(t *testing.T) {
	o := setup(t)
	commitA := strings.Repeat("a", 40)
	p1 := plan(t, "install", o)
	p1.SourceCommit = commitA
	p1.ID = planID(p1)
	apply(t, p1)
	p2 := plan(t, "install", o)
	p2.SourceCommit = commitA
	p2.ID = planID(p2)
	apply(t, p2)
	commits := readCommits(t, o.StateDir, p1.Release.ID)
	if len(commits) != 1 {
		t.Fatalf("expected the same commit recorded once, got %+v", commits)
	}
}

func TestPlanWithoutSourceCommitCreatesNoRecord(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	apply(t, p)
	path := filepath.Join(o.StateDir, "releases", p.Release.ID+".commits.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no commits record, stat error: %v", err)
	}
}

func TestSourceCommitNotRecordedWhenTransactionFailsBeforeCommit(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.SourceCommit = strings.Repeat("c", 40)
	p.ID = planID(p)
	e := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("expected the injected failure")
	}
	path := filepath.Join(o.StateDir, "releases", p.Release.ID+".commits.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("commit recorded despite a transaction that failed before commit: %v", err)
	}
}

// TestPlanIDStableWithoutSourceCommit pins the plan ID of a fixed, minimal
// plan computed before SourceCommit existed. Because the field is
// json:",omitempty", a plan that leaves it empty must still hash to the same
// ID: existing saved plans stay loadable across this change.
func TestPlanIDStableWithoutSourceCommit(t *testing.T) {
	p := Plan{
		Version: stateVersion,
		Action:  "install",
		Config: target.Config{
			Scope:        "user",
			Home:         "/synthetic/home",
			CodexHome:    "/synthetic/home/.codex",
			ClaudeHome:   "/synthetic/home/.claude",
			PiHome:       "/synthetic/home/.pi/agent",
			GrokHome:     "/synthetic/home/.grok",
			OpenCodeHome: "/synthetic/home/.config/opencode",
			CursorHome:   "/synthetic/home/.cursor",
			Synthetic:    true,
		},
		Hosts:     []string{"codex"},
		StateDir:  "/synthetic/state",
		StateHash: hash(nil),
	}
	const want = "9f34b952da4ce1ee475cb3fff3ede27a680d07c50786972b1c518bb7c2d975e0"
	if got := planID(p); got != want {
		t.Fatalf("plan ID changed for a plan without SourceCommit: got %s want %s", got, want)
	}
}

func TestValidatePlanRejectsMalformedSourceCommit(t *testing.T) {
	o := setup(t)
	base := plan(t, "install", o)
	state, _, err := readState(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"not-hex", strings.Repeat("a", 39), strings.Repeat("a", 41), strings.Repeat("A", 40), strings.Repeat("g", 40), strings.Repeat("a", 63)} {
		t.Run(bad, func(t *testing.T) {
			q := base
			q.SourceCommit = bad
			q.ID = planID(q)
			if err := validatePlan(q, state); err == nil {
				t.Fatalf("accepted invalid source commit %q", bad)
			}
		})
	}
	for _, good := range []string{"", strings.Repeat("a", 40), strings.Repeat("a", 64)} {
		t.Run("valid-"+good, func(t *testing.T) {
			q := base
			q.SourceCommit = good
			q.ID = planID(q)
			if err := validatePlan(q, state); err != nil {
				t.Fatalf("rejected valid source commit %q: %v", good, err)
			}
		})
	}
}

func TestReleasesEmptyWhenNoStateDirectory(t *testing.T) {
	o := setup(t)
	entries, err := Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no releases, got %+v", entries)
	}
}

func TestReleasesReportsConsumersAndEmptyCommits(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	entries, err := Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one release, got %+v", entries)
	}
	if len(entries[0].Commits) != 0 {
		t.Fatalf("expected empty commits without a recorded commit, got %+v", entries[0].Commits)
	}
	if len(entries[0].Consumers) != 2 {
		t.Fatalf("expected two consumers (codex, claude), got %+v", entries[0].Consumers)
	}
	if _, err := time.Parse(time.RFC3339, entries[0].LastWrittenAt); err != nil {
		t.Fatalf("last_written_at is not RFC 3339: %v", err)
	}
}

func TestReleasesOrderedAndSurviveOlderReleaseReinstall(t *testing.T) {
	o := setup(t)
	commitA := strings.Repeat("a", 40)
	commitB := strings.Repeat("b", 40)
	p1 := plan(t, "install", o)
	p1.SourceCommit = commitA
	p1.ID = planID(p1)
	apply(t, p1)
	firstID := p1.Release.ID

	put(t, filepath.Join(o.Source, GlobalSource), "# New\nDifferent content.\n")
	p2 := plan(t, "install", o)
	p2.SourceCommit = commitB
	p2.ID = planID(p2)
	apply(t, p2)
	secondID := p2.Release.ID
	if secondID == firstID {
		t.Fatal("test setup: expected a distinct second release")
	}

	// Force a deterministic mtime order: the first snapshot is older.
	older := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(o.StateDir, "releases", firstID+".json"), older, older); err != nil {
		t.Fatal(err)
	}

	entries, err := Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].ID != secondID || entries[1].ID != firstID {
		t.Fatalf("expected newest-first order [%s, %s], got %+v", secondID, firstID, entries)
	}
	if len(entries[0].Consumers) != 2 {
		t.Fatalf("expected the current release to keep both consumers, got %+v", entries[0].Consumers)
	}
	if len(entries[1].Consumers) != 0 {
		t.Fatalf("expected the superseded release to have no consumers, got %+v", entries[1].Consumers)
	}
	if len(entries[1].Commits) != 1 || entries[1].Commits[0] != commitA {
		t.Fatalf("expected the superseded release to keep its commit history, got %+v", entries[1].Commits)
	}

	// Roll back to the older release: its consumers must reappear.
	old := o
	old.ReleaseID = firstID
	apply(t, plan(t, "install", old))
	entries, err = Releases(o)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]ReleaseEntry{}
	for _, e := range entries {
		byID[e.ID] = e
	}
	if len(byID[firstID].Consumers) != 2 {
		t.Fatalf("expected the reinstalled release to show its consumers again, got %+v", byID[firstID])
	}
	if len(byID[secondID].Consumers) != 0 {
		t.Fatalf("expected the superseded release to lose its consumers, got %+v", byID[secondID])
	}
}

func TestReleasesFailsNamingInvalidSnapshot(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	apply(t, p)
	path := filepath.Join(o.StateDir, "releases", p.Release.ID+".json")
	put(t, path, "{not json")
	_, err := Releases(o)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("expected an error naming %s, got %v", path, err)
	}
}

func TestReleasesFailsNamingUnreadableCommitsRecord(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	p.SourceCommit = strings.Repeat("a", 40)
	p.ID = planID(p)
	apply(t, p)
	path := filepath.Join(o.StateDir, "releases", p.Release.ID+".commits.json")
	put(t, path, "{not json")
	_, err := Releases(o)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("expected an error naming %s, got %v", path, err)
	}
}

func TestBindSourceCommitRecomputesAcceptedPlanID(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	commit := strings.Repeat("c", 40)
	bound, err := BindSourceCommit(p, commit)
	if err != nil {
		t.Fatal(err)
	}
	if bound.SourceCommit != commit || bound.ID == p.ID || bound.ID != planID(bound) {
		t.Fatalf("source commit not bound with a recomputed ID: %s", bound.ID)
	}
	if _, err := BindSourceCommit(p, "NOT-A-COMMIT"); err == nil {
		t.Fatal("malformed source commit accepted")
	}
}
