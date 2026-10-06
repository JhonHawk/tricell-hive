package management

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The retired hive bootstrap wrote an "installer" key into plans and onboarding
// journals. These fixtures rebuild that layout (the key sits right after
// "product" in a plan and right after "version" in an onboarding journal) so
// state written by v0.1.0 keeps decoding.
const legacyInstallerJSON = `{"ArtifactID":"a","ManagerDigest":"m","PackageDigest":"p","Directory":"d","Manager":"hive","Package":"pkg"}`

type legacyPlan struct {
	Product   *ProductIdentity `json:"product,omitempty"`
	Installer json.RawMessage  `json:"installer,omitempty"`
	Plan
}

type legacyJournal struct {
	Version                 int
	ID, Phase               string
	Plan                    legacyPlan
	Entries                 []entry
	BeforeState, AfterState snapshot
	CreatedDirs             []string
	Integrity               string
	PackageDone             bool `json:",omitempty"`
	PackagePending          bool `json:",omitempty"`
}

type legacyOnboardingJournal struct {
	Version   int             `json:"version"`
	Installer json.RawMessage `json:"installer,omitempty"`
	OnboardingResult
	StateDir  string `json:"state_dir"`
	CoreID    string `json:"core_id"`
	Integrity string `json:"integrity"`
}

func TestRecoverAcceptsPlanWithRetiredInstallerField(t *testing.T) {
	o := setup(t)
	cp := filepath.Join(o.Home, ".codex", "AGENTS.md")
	put(t, cp, "original\n")
	p := plan(t, "install", o)
	e := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return errors.New("injected")
		}
		return nil
	}}
	if _, err := e.Apply(p); err == nil {
		t.Fatal("did not fail")
	}
	var pend pending
	if err := decodeFile(filepath.Join(o.StateDir, "pending.json"), &pend); err != nil {
		t.Fatal(err)
	}
	jp := filepath.Join(o.StateDir, "transactions", pend.ID+".json")
	var j journal
	if err := decodeFile(jp, &j); err != nil {
		t.Fatal(err)
	}
	lp := legacyPlan{Product: j.Plan.Product, Installer: json.RawMessage(legacyInstallerJSON), Plan: j.Plan}
	lp.Plan.ID = ""
	lp.Plan.ID = hash(encode(lp))
	lj := legacyJournal{Version: j.Version, ID: j.ID, Phase: j.Phase, Plan: lp, Entries: j.Entries, BeforeState: j.BeforeState, AfterState: j.AfterState, CreatedDirs: j.CreatedDirs, PackageDone: j.PackageDone, PackagePending: j.PackagePending}
	lj.Integrity = hash(encode(lj))
	if err := os.WriteFile(jp, encode(lj), 0600); err != nil {
		t.Fatal(err)
	}
	// The pending marker names the transaction by journal ID, which is unchanged.
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatalf("recover rejected a plan carrying the retired installer field: %v", err)
	}
	if get(t, cp) != "original\n" {
		t.Fatal("recovery changed bytes")
	}
}

func TestLoadOnboardingAcceptsRetiredInstallerField(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{fail: true, unknown: true}
	if _, err := (Engine{}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner); err == nil {
		t.Fatal("missing unknown outcome")
	}
	j, err := loadOnboarding(o.StateDir)
	if err != nil {
		t.Fatal(err)
	}
	lj := legacyOnboardingJournal{Version: j.Version, Installer: json.RawMessage(legacyInstallerJSON), OnboardingResult: j.OnboardingResult, StateDir: j.StateDir, CoreID: j.CoreID}
	lj.Integrity = hash(encode(lj))
	if err := os.WriteFile(onboardingPath(o.StateDir), encode(lj), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadOnboarding(o.StateDir); err != nil {
		t.Fatalf("loadOnboarding rejected a journal carrying the retired installer field: %v", err)
	}
}
