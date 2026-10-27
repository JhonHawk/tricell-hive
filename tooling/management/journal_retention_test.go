package management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// journalIDs lists the transaction journals in o's state directory.
func journalIDs(t *testing.T, o Options) map[string]bool {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(o.StateDir, "transactions", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, n := range names {
		ids[strings.TrimSuffix(filepath.Base(n), ".json")] = true
	}
	return ids
}

// cycleApplies alternates install and remove n times, so each apply commits
// a change, and returns the last transaction ID.
func cycleApplies(t *testing.T, o Options, n int) string {
	t.Helper()
	var id string
	for i := 0; i < n; i++ {
		action := "install"
		if i%2 == 1 {
			action = "remove"
		}
		id, _, _ = strings.Cut(apply(t, plan(t, action, o)), ";")
	}
	return id
}

// seedJournals writes n journals in phase into stateDir, oldest first, one
// minute apart, and returns their IDs in that order.
func seedJournals(t *testing.T, stateDir, phase string, n int) []string {
	t.Helper()
	var ids []string
	start := time.Now().Add(-time.Hour)
	for i := 0; i < n; i++ {
		id := fmt.Sprintf("%032x", len(phase)*1000+i)
		path := filepath.Join(stateDir, "transactions", id+".json")
		put(t, path, fmt.Sprintf("{\n  \"Version\": %d,\n  \"ID\": %q,\n  \"Phase\": %q,\n  \"Plan\": {}\n}", stateVersion, id, phase))
		when := start.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, when, when); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func TestPruneJournalsKeepsThePendingJournal(t *testing.T) {
	o := setup(t)
	ids := seedJournals(t, o.StateDir, "committed", retainedJournals+2)
	put(t, filepath.Join(o.StateDir, "pending.json"), fmt.Sprintf("{\"ID\": %q}", ids[0]))
	if w := pruneJournals(o.StateDir, ""); w != "" {
		t.Fatal(w)
	}
	kept := journalIDs(t, o)
	if !kept[ids[0]] || len(kept) != retainedJournals+1 {
		t.Fatalf("pending journal kept: %v, journals left: %d", kept[ids[0]], len(kept))
	}
}

// failedExternal reports a known, final failure, so Onboard ends partial.
type failedExternal struct{ fakeExternal }

func (f *failedExternal) Reconcile(_ ExternalStep, _ json.RawMessage) (string, error) {
	return StepFailed, nil
}

func TestJournalRetentionAfterOnboardingWithSteps(t *testing.T) {
	o := setup(t)
	steps := []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}
	var last OnboardingResult
	for i := 0; i < retainedJournals+2; i++ {
		action, runner := "install", ExternalRunner(&fakeExternal{})
		if i%2 == 1 {
			action = "remove"
		}
		if i == retainedJournals+1 {
			runner = &failedExternal{}
		}
		result, err := (Engine{}).Onboard(plan(t, action, o), steps, runner)
		if i < retainedJournals+1 && err != nil {
			t.Fatal(err)
		}
		last = result
	}
	if last.Phase != "partial" {
		t.Fatalf("last onboarding phase %q, want partial", last.Phase)
	}
	if ids := journalIDs(t, o); len(ids) != retainedJournals {
		t.Fatalf("kept %d journals after onboarding, want %d", len(ids), retainedJournals)
	}
}

// lockingExternal makes the transactions directory read-only when its step
// runs, after the core committed and before Onboard prunes.
type lockingExternal struct {
	fakeExternal
	dir string
}

func (f *lockingExternal) Execute(s ExternalStep) (json.RawMessage, error) {
	if err := os.Chmod(f.dir, 0500); err != nil {
		return nil, err
	}
	return f.fakeExternal.Execute(s)
}

func TestJournalPruneFailureKeepsTheOperationSuccessful(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	o := setup(t)
	cycleApplies(t, o, retainedJournals+1)
	dir := filepath.Join(o.StateDir, "transactions")
	t.Cleanup(func() { os.Chmod(dir, 0700) })
	steps := []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}
	result, err := (Engine{}).Onboard(plan(t, "remove", o), steps, &lockingExternal{dir: dir})
	if err != nil || result.Phase != "completed" {
		t.Fatalf("%+v %v", result, err)
	}
	if !strings.Contains(result.ID, "could not be pruned") {
		t.Fatalf("result %q carries no pruning warning", result.ID)
	}
}

func TestPruneJournalsSkipsWhenPendingIsUnreadable(t *testing.T) {
	o := setup(t)
	seedJournals(t, o.StateDir, "committed", retainedJournals+2)
	put(t, filepath.Join(o.StateDir, "pending.json"), "{")
	if w := pruneJournals(o.StateDir, ""); !strings.Contains(w, "skipped") {
		t.Fatalf("warning %q", w)
	}
	if ids := journalIDs(t, o); len(ids) != retainedJournals+2 {
		t.Fatalf("deleted journals with an unreadable pending.json: %d left", len(ids))
	}
}

func TestPruneJournalsKeepsThePendingOnboardingCore(t *testing.T) {
	o := setup(t)
	ids := seedJournals(t, o.StateDir, "committed", retainedJournals+2)
	put(t, onboardingPath(o.StateDir), fmt.Sprintf("{\"version\": 1, \"id\": %q, \"phase\": \"core_pending\", \"core_id\": %q}", ids[0], ids[0]))
	if w := pruneJournals(o.StateDir, ""); w != "" {
		t.Fatal(w)
	}
	kept := journalIDs(t, o)
	if !kept[ids[0]] || len(kept) != retainedJournals+1 {
		t.Fatalf("onboarding core kept: %v, journals left: %d", kept[ids[0]], len(kept))
	}
}

func TestPruneJournalsKeepsUnfinishedAndUnrecognizedJournals(t *testing.T) {
	o := setup(t)
	prepared := seedJournals(t, o.StateDir, "prepared", 1)[0]
	odd := fmt.Sprintf("%032x", 7)
	oddPath := filepath.Join(o.StateDir, "transactions", odd+".json")
	put(t, oddPath, fmt.Sprintf("{\"ID\": %q, \"Version\": 6, \"Phase\": \"committed\"}", odd))
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(oddPath, old, old); err != nil {
		t.Fatal(err)
	}
	seedJournals(t, o.StateDir, "committed", retainedJournals+2)
	if w := pruneJournals(o.StateDir, ""); w != "" {
		t.Fatal(w)
	}
	kept := journalIDs(t, o)
	if !kept[prepared] || !kept[odd] || len(kept) != retainedJournals+2 {
		t.Fatalf("prepared kept: %v, unrecognized kept: %v, journals left: %d", kept[prepared], kept[odd], len(kept))
	}
}

func TestJournalRetentionKeepsNewestCommittedJournals(t *testing.T) {
	o := setup(t)
	last := cycleApplies(t, o, retainedJournals+2)
	ids := journalIDs(t, o)
	if len(ids) != retainedJournals || !ids[last] {
		t.Fatalf("kept %d journals (last kept: %v), want %d including the last", len(ids), ids[last], retainedJournals)
	}
}
