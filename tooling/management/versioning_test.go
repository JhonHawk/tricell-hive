package management

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductVersionCreatesConsumerReceiptWithoutChangingPayloadIdentity(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.0\n")
	first := plan(t, "install", o)
	apply(t, first)
	var state map[string]json.RawMessage
	if err := json.Unmarshal([]byte(get(t, filepath.Join(o.StateDir, "state.json"))), &state); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(state["installations"]), `"1.2.0"`) {
		t.Fatal("missing product version receipt")
	}
	put(t, filepath.Join(o.Source, "VERSION"), "1.2.1\n")
	second := plan(t, "install", o)
	if second.Release.ID != first.Release.ID {
		t.Fatal("product version altered payload identity")
	}
	if unchanged, err := PlanUnchanged(second); err != nil || unchanged {
		t.Fatalf("metadata update omitted: %v %v", unchanged, err)
	}
	apply(t, second)
	if !strings.Contains(get(t, filepath.Join(o.StateDir, "state.json")), `"1.2.1"`) {
		t.Fatal("new label not recorded")
	}
}

func TestPublishedVersionCannotBeReboundToDifferentSource(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "2.0.0\n")
	apply(t, plan(t, "install", o))
	put(t, filepath.Join(o.Source, GlobalSource), "Different content.\n")
	if _, err := BuildPlan("install", o); err == nil {
		t.Fatal("version rebound to different source")
	}
}

func TestVersionReceiptRecoveryPreservesCommittedPredecessor(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "3.0.0\n")
	apply(t, plan(t, "install", o))
	before := get(t, filepath.Join(o.StateDir, "state.json"))
	put(t, filepath.Join(o.Source, "VERSION"), "3.0.1\n")
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(stage string) error {
		if stage == "state" {
			return &versionTestError{}
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("metadata-only operation did not reach journal")
	}
	if _, err = (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if got := get(t, filepath.Join(o.StateDir, "state.json")); got != before {
		t.Fatal("recovery changed previous receipt")
	}
}

type versionTestError struct{}

func (*versionTestError) Error() string { return "test interruption" }

func TestRequiredHostsExplainsSharedChangesWithoutSelectingThem(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	put(t, filepath.Join(o.Source, SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nUpdated shared content.\n")
	hosts, err := RequiredHosts(o)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(hosts, ",") != "claude,codex" {
		t.Fatalf("wrong cohort %v", hosts)
	}
	if len(o.Hosts) != 1 {
		t.Fatal("selection mutated")
	}
	if _, err = BuildPlan("install", o); err == nil {
		t.Fatal("unapproved cohort accepted")
	}
}

func TestConsumerVersionIsIndependentOfSharedRecordProvenance(t *testing.T) {
	o := setup(t)
	put(t, filepath.Join(o.Source, "VERSION"), "4.0.0\n")
	apply(t, plan(t, "install", o))
	o.Hosts = []string{"codex"}
	put(t, filepath.Join(o.Source, "VERSION"), "4.0.1\n")
	apply(t, plan(t, "install", o))
	s := stateFor(t, o)
	for _, item := range []struct{ host, version string }{{"codex", "4.0.1"}, {"claude", "4.0.0"}} {
		got, status := installationVersion(s, Consumer{item.host, "user", o.Home})
		if got != item.version || status != "verified" {
			t.Fatalf("%s: %s %s", item.host, got, status)
		}
	}
}

func TestV5StateIsNotMigratedDuringPlanning(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	apply(t, p)
	old := stateFor(t, o)
	old.Version = 5
	put(t, filepath.Join(o.StateDir, "state.json"), string(encode(old)))
	before := get(t, filepath.Join(o.StateDir, "state.json"))
	next := plan(t, "install", o)
	if get(t, filepath.Join(o.StateDir, "state.json")) != before {
		t.Fatal("planning migrated state")
	}
	if next.Release.ID != p.Release.ID {
		t.Fatal("migration changed payload identity")
	}
	apply(t, next)
	if stateFor(t, o).Version != stateVersion {
		t.Fatal("state not migrated transactionally")
	}
}
