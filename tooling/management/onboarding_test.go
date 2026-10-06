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

type fakeExternal struct {
	calls   int
	fail    bool
	unknown bool
}

func (f *fakeExternal) Validate(_ ExternalStep) error { return nil }
func (f *fakeExternal) Execute(_ ExternalStep) (json.RawMessage, error) {
	f.calls++
	if f.fail {
		return nil, fmt.Errorf("fixture failed")
	}
	return json.RawMessage(`{"installed":true}`), nil
}
func (f *fakeExternal) Reconcile(_ ExternalStep, _ json.RawMessage) (string, error) {
	if f.unknown {
		return "unknown", nil
	}
	return "verified", nil
}
func (f *fakeExternal) Revert(_ ExternalStep, _ json.RawMessage) error { return nil }

func TestOnboardingKeepsCoreAfterProviderFailure(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{fail: true}
	_, err := (Engine{}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err == nil {
		t.Fatal("partial install reported success")
	}
	if len(stateFor(t, o).Records) == 0 {
		t.Fatal("provider failure rolled back core")
	}
	if _, err := BuildPlan("remove", o); err == nil {
		t.Fatal("unknown external result failed to block mutation")
	}
	if _, err := (Engine{}).RecoverOnboarding(o.StateDir, runner); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildPlan("remove", o); err != nil {
		t.Fatal(err)
	}
}
func TestOnboardingRunsNewCapabilityWithUnchangedCore(t *testing.T) {
	o := setup(t)
	apply(t, plan(t, "install", o))
	runner := &fakeExternal{}
	result, err := (Engine{}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err != nil || result.Phase != "completed" || runner.calls != 1 {
		t.Fatalf("%+v %v calls %d", result, err, runner.calls)
	}
}
func TestOnboardingRecoverReconcilesInterruptedProviderWithoutRerun(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{}
	engine := Engine{failpoint: func(s string) error {
		if s == "provider_result" {
			return fmt.Errorf("crash")
		}
		return nil
	}}
	_, err := engine.Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err == nil {
		t.Fatal("missing crash")
	}
	if _, err = (Engine{}).RecoverOnboarding(o.StateDir, runner); err != nil {
		t.Fatal(err)
	}
	if runner.calls != 1 {
		t.Fatal("recovery reran installer")
	}
	absent(t, filepath.Join(o.StateDir, "onboarding-pending.json"))
}

func TestOnboardingDoesNotAdoptEarlierPendingCore(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	_, err := (Engine{failpoint: func(s string) error {
		if s == "prepared" {
			return fmt.Errorf("stop")
		}
		return nil
	}}).Apply(p)
	if err == nil {
		t.Fatal("missing pending core")
	}
	if _, err = (Engine{}).Onboard(p, []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, &fakeExternal{}); err == nil {
		t.Fatal("adopted unrelated pending transaction")
	}
	absent(t, filepath.Join(o.StateDir, "onboarding-pending.json"))
}
func TestOnboardingUnknownOutcomeStaysBlocked(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{fail: true, unknown: true}
	_, err := (Engine{}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err == nil {
		t.Fatal("missing unknown outcome")
	}
	_, err = (Engine{}).RecoverOnboarding(o.StateDir, runner)
	if err == nil {
		t.Fatal("unknown outcome finalized")
	}
	if _, err = BuildPlan("install", o); err == nil {
		t.Fatal("pending did not block installation")
	}
	if runner.calls != 1 {
		t.Fatal("recovery executed installer")
	}
}
func TestOnboardingStaleTargetLeavesNoPendingParent(t *testing.T) {
	o := setup(t)
	p := plan(t, "install", o)
	put(t, p.Changes[0].Target.Path, "late edit")
	runner := &fakeExternal{}
	if _, err := (Engine{}).Onboard(p, []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner); err == nil {
		t.Fatal("stale target accepted")
	}
	if pending, err := OnboardingPending(o.StateDir); err != nil || pending {
		t.Fatalf("core never started, yet onboarding is pending (%v, %v); install stays blocked", pending, err)
	}
	if get(t, p.Changes[0].Target.Path) != "late edit" || runner.calls != 0 {
		t.Fatal("stale onboarding changed the target or ran a provider")
	}
	// A fresh plan must be possible without recovery.
	if _, err := (Engine{}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner); err != nil {
		t.Fatalf("fresh plan after a stale one: %v", err)
	}
}

func TestPendingReportsNoneCoreAndOnboarding(t *testing.T) {
	o := setup(t)
	if kind, err := Pending(o.StateDir); err != nil || kind != PendingNone {
		t.Fatalf("fresh state = %v, %v; want none", kind, err)
	}
	crash := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return fmt.Errorf("stop")
		}
		return nil
	}}
	if _, err := crash.Apply(plan(t, "install", o)); err == nil {
		t.Fatal("missing core crash")
	}
	if kind, err := Pending(o.StateDir); err != nil || kind != PendingCore {
		t.Fatalf("interrupted core = %v, %v; want core", kind, err)
	}
	if _, err := (Engine{}).Recover(o.StateDir); err != nil {
		t.Fatal(err)
	}
	if _, err := crash.Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, &fakeExternal{}); err == nil {
		t.Fatal("missing onboarding crash")
	}
	// The onboarding journal takes precedence: its recovery also finishes the core.
	if kind, err := Pending(o.StateDir); err != nil || kind != PendingOnboarding {
		t.Fatalf("interrupted onboarding = %v, %v; want onboarding", kind, err)
	}
}

func TestOnboardingKeepsParentWhenCoreFailedAfterStarting(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{}
	crash := Engine{failpoint: func(s string) error {
		if s == "write:0" {
			return fmt.Errorf("stop")
		}
		return nil
	}}
	if _, err := crash.Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner); err == nil {
		t.Fatal("missing crash inside the core transaction")
	}
	if pending, err := OnboardingPending(o.StateDir); err != nil || !pending {
		t.Fatalf("core started, so the parent must stay for recovery (%v, %v)", pending, err)
	}
	result, err := (Engine{}).RecoverOnboarding(o.StateDir, runner)
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "partial" || runner.calls != 0 {
		t.Fatalf("recovery phase=%s calls=%d", result.Phase, runner.calls)
	}
}

func TestOnboardingCrashBetweenCoreCommitAndParentReceiptKeepsCore(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{}
	_, err := (Engine{failpoint: func(s string) error {
		if s == "core_applied" {
			return fmt.Errorf("stop")
		}
		return nil
	}}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err == nil {
		t.Fatal("missing crash between child commit and parent receipt")
	}
	if j, err := loadOnboarding(o.StateDir); err != nil || j.Phase != "core_pending" {
		t.Fatalf("parent journal = %q, %v; want core_pending", j.Phase, err)
	}
	result, err := (Engine{}).RecoverOnboarding(o.StateDir, runner)
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "partial" || runner.calls != 0 || len(stateFor(t, o).Records) == 0 {
		t.Fatalf("recovery reverted the committed core or replayed a provider: phase=%s calls=%d", result.Phase, runner.calls)
	}
}

func TestOnboardingCrashAfterCoreCommitSkipsUnstartedSteps(t *testing.T) {
	o := setup(t)
	runner := &fakeExternal{}
	_, err := (Engine{failpoint: func(s string) error {
		if s == "core_committed" {
			return fmt.Errorf("stop")
		}
		return nil
	}}).Onboard(plan(t, "install", o), []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner)
	if err == nil {
		t.Fatal("missing crash")
	}
	result, err := (Engine{}).RecoverOnboarding(o.StateDir, runner)
	if err != nil {
		t.Fatal(err)
	}
	if result.Phase != "partial" || runner.calls != 0 || len(stateFor(t, o).Records) == 0 {
		t.Fatal("unsafe recovery")
	}
}

// onboardOnce runs one onboarding with the given step IDs and returns its result.
func onboardOnce(t *testing.T, o Options, ids ...string) OnboardingResult {
	t.Helper()
	var steps []ExternalStep
	for _, id := range ids {
		steps = append(steps, ExternalStep{ID: id, Payload: json.RawMessage(`{}`)})
	}
	result, err := (Engine{}).Onboard(plan(t, "install", o), steps, &fakeExternal{})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestLastOnboardingWithoutRecordsReportsNone(t *testing.T) {
	o := setup(t)
	for name, prepare := range map[string]func(){
		"no state directory": func() {},
		"empty onboarding directory": func() {
			if err := os.MkdirAll(filepath.Join(o.StateDir, "onboarding"), 0700); err != nil {
				t.Fatal(err)
			}
		},
	} {
		prepare()
		if _, _, ok, err := LastOnboarding(o.StateDir); ok || err != nil {
			t.Fatalf("%s: ok=%v err=%v; want no record and no error", name, ok, err)
		}
	}
	if _, err := os.Lstat(o.StateDir + "/onboarding-pending.json"); !os.IsNotExist(err) {
		t.Fatal("reading created state")
	}
}

func TestLastOnboardingPicksTheNewestRecord(t *testing.T) {
	o := setup(t)
	first := onboardOnce(t, o, "first-manual")
	second := onboardOnce(t, o, "second-manual")
	dir := filepath.Join(o.StateDir, "onboarding")
	old := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	fresh := old.Add(48 * time.Hour)
	for id, at := range map[string]time.Time{first.ID: fresh, second.ID: old} {
		if err := os.Chtimes(filepath.Join(dir, id+".json"), at, at); err != nil {
			t.Fatal(err)
		}
	}
	got, when, ok, err := LastOnboarding(o.StateDir)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.ID != first.ID || got.Phase != "completed" || len(got.Steps) != 1 || got.Steps[0].Step.ID != "first-manual" {
		t.Fatalf("picked %+v; want the record with the newest modification time (%s)", got, first.ID)
	}
	if !when.Equal(fresh) {
		t.Fatalf("time = %s; want %s", when, fresh)
	}
}

func TestLastOnboardingReadsThroughASymlinkedStateDirectory(t *testing.T) {
	o := setup(t)
	want := onboardOnce(t, o, "engram-manual")
	link := filepath.Join(t.TempDir(), "state-link")
	if err := os.Symlink(o.StateDir, link); err != nil {
		t.Fatal(err)
	}
	got, _, ok, err := LastOnboarding(link)
	if err != nil || !ok || got.ID != want.ID {
		t.Fatalf("got %+v ok=%v err=%v; a state directory reached through a symlink must read as valid", got, ok, err)
	}
}

func TestLastOnboardingIgnoresFilesWithOtherNames(t *testing.T) {
	o := setup(t)
	want := onboardOnce(t, o, "engram-manual")
	dir := filepath.Join(o.StateDir, "onboarding")
	orig, err := os.ReadFile(filepath.Join(dir, want.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	for _, name := range []string{"notes.json", want.ID + ".json.bak", "ABCDEF0123456789ABCDEF0123456789.json", want.ID[:31] + ".json", ".hidden"} {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("not a record"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(p, future, future); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, strings.Repeat("a", 32)+".json"), 0700); err != nil {
		t.Fatal(err)
	}
	got, _, ok, err := LastOnboarding(o.StateDir)
	if err != nil || !ok || got.ID != want.ID {
		t.Fatalf("got %+v ok=%v err=%v; other names must be ignored", got, ok, err)
	}
	if after, _ := os.ReadFile(filepath.Join(dir, want.ID+".json")); string(after) != string(orig) {
		t.Fatal("reading changed the record")
	}
}

func TestLastOnboardingRejectsTamperedRecords(t *testing.T) {
	tamper := map[string]func(string) string{
		"changed step status": func(s string) string { return strings.Replace(s, `"verified"`, `"failed"`, 1) },
		"changed phase":       func(s string) string { return strings.Replace(s, `"completed"`, `"partial"`, 1) },
		"trailing data":       func(s string) string { return s + "{}" },
		"not json":            func(string) string { return "{" },
	}
	for name, edit := range tamper {
		t.Run(name, func(t *testing.T) {
			o := setup(t)
			result := onboardOnce(t, o, "engram-manual")
			p := filepath.Join(o.StateDir, "onboarding", result.ID+".json")
			raw, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			changed := edit(string(raw))
			if changed == string(raw) {
				t.Fatal("the edit changed nothing")
			}
			if err := os.WriteFile(p, []byte(changed), 0600); err != nil {
				t.Fatal(err)
			}
			if _, _, ok, err := LastOnboarding(o.StateDir); err == nil || ok {
				t.Fatalf("ok=%v err=%v; a tampered record must be an error", ok, err)
			}
		})
	}
}

func TestLastOnboardingRejectsARecordWhoseNameIsNotItsID(t *testing.T) {
	o := setup(t)
	result := onboardOnce(t, o, "engram-manual")
	dir := filepath.Join(o.StateDir, "onboarding")
	other := strings.Repeat("b", 32)
	if err := os.Rename(filepath.Join(dir, result.ID+".json"), filepath.Join(dir, other+".json")); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := LastOnboarding(o.StateDir); err == nil || ok {
		t.Fatalf("ok=%v err=%v; a file whose name differs from the record ID must be an error", ok, err)
	}
}

func TestLastOnboardingRejectsARecordFromAnotherStateDirectory(t *testing.T) {
	a, b := setup(t), setup(t)
	result := onboardOnce(t, a, "engram-manual")
	dst := filepath.Join(b.StateDir, "onboarding")
	if err := os.MkdirAll(dst, 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(a.StateDir, "onboarding", result.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dst, result.ID+".json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := LastOnboarding(b.StateDir); err == nil || ok {
		t.Fatalf("ok=%v err=%v; a record copied from another state directory must be an error", ok, err)
	}
}

func TestLastOnboardingRejectsAnUnfinishedRecord(t *testing.T) {
	o := setup(t)
	result := onboardOnce(t, o, "engram-manual")
	p := filepath.Join(o.StateDir, "onboarding", result.ID+".json")
	var j onboardingJournal
	if err := decodeFile(p, &j); err != nil {
		t.Fatal(err)
	}
	j.Phase = "providers_pending"
	j.Integrity = onboardingHash(j) // consistent, but never finished
	if err := writeJSON(p, j); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := LastOnboarding(o.StateDir); err == nil || ok {
		t.Fatalf("ok=%v err=%v; a record that never finished must be an error", ok, err)
	}
}
