package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/management"
)

// TestRunReleasesPrintsSnapshotsNewestFirstWithCommitsAndConsumers covers
// T4: `hive releases` prints management.Releases as indented JSON, newest
// snapshot first, with an empty (not null) Commits array for a release with
// no recorded commit.
func TestRunReleasesPrintsSnapshotsNewestFirstWithCommitsAndConsumers(t *testing.T) {
	f := newCharacterizationFixture(t)
	o := management.Options{Scope: "user", Home: f.home, StateDir: f.stateDir, Source: f.source, Hosts: []string{"codex"}}
	p1, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p1); err != nil {
		t.Fatal(err)
	}

	// A distinct second release, with a recorded source commit, so the two
	// snapshots differ and can be told apart by ID.
	putCharacterization(t, filepath.Join(f.source, management.GlobalSource), "# Rules\nUpdated.\n")
	p2, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	if p2, err = management.BindSourceCommit(p2, strings.Repeat("b", 40)); err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Apply(p2); err != nil {
		t.Fatal(err)
	}
	if p2.Release.ID == p1.Release.ID {
		t.Fatal("test setup: expected a distinct second release")
	}

	// Force a deterministic mtime order: the first snapshot is older.
	older := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(f.stateDir, "releases", p1.Release.ID+".json"), older, older); err != nil {
		t.Fatal(err)
	}

	out := captureStdout(t, func() {
		if err := run([]string{"releases", "--home", f.home, "--state-dir", f.stateDir}); err != nil {
			t.Fatal(err)
		}
	})

	var entries []management.ReleaseEntry
	if err := json.Unmarshal([]byte(out), &entries); err != nil {
		t.Fatalf("releases output is not the expected JSON: %v (%q)", err, out)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 releases, got %d: %+v", len(entries), entries)
	}
	if entries[0].ID != p2.Release.ID || entries[1].ID != p1.Release.ID {
		t.Fatalf("expected newest-first order [%s, %s], got %+v", p2.Release.ID, p1.Release.ID, entries)
	}
	if len(entries[0].Commits) != 1 || entries[0].Commits[0] != p2.SourceCommit {
		t.Fatalf("expected the newest release's recorded commit, got %+v", entries[0])
	}
	if len(entries[0].Consumers) != 1 || len(entries[1].Consumers) != 0 {
		t.Fatalf("expected only the current release to keep its consumer, got %+v", entries)
	}
	// The release with no commits record must serialize as an empty array,
	// never null: entries[1].Commits is nil here only if the raw JSON said
	// null, since json.Unmarshal into a []string preserves that distinction.
	if entries[1].Commits == nil {
		t.Fatalf("expected an empty (not null) Commits array for the release with no record, got %+v", entries[1])
	}
	if len(entries[1].Commits) != 0 {
		t.Fatalf("expected no commits for the release with no record, got %+v", entries[1])
	}
	if !strings.Contains(out, "\"commits\": []") {
		t.Fatalf("expected a literal empty Commits array in the JSON output, got:\n%s", out)
	}
}

// TestRunReleasesPrintsEmptyArrayWhenNoReleasesDirectory covers the other
// half of the empty-vs-null requirement: a state directory that was never
// onboarded (no releases/ folder at all) must print "[]", not "null".
func TestRunReleasesPrintsEmptyArrayWhenNoReleasesDirectory(t *testing.T) {
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	stateDir := filepath.Join(home, "state")
	out := captureStdout(t, func() {
		if err := run([]string{"releases", "--home", home, "--state-dir", stateDir}); err != nil {
			t.Fatal(err)
		}
	})
	if out != "[]\n" {
		t.Fatalf("expected an empty JSON array, got %q", out)
	}
}
