package management

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestOnboardRejectsChangedRetainedInstallerBeforeAnyJournal(t *testing.T) {
	o := setup(t)
	bound := boundInstallPlan(t, o, strings.Repeat("9", 64))
	if err := os.WriteFile(bound.Installer.Package, []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	runner := &fakeExternal{}
	if _, err := (Engine{}).Onboard(bound, []ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}}, runner); err == nil {
		t.Fatal("onboarded a plan whose retained installer changed")
	}
	if pending, err := OnboardingPending(o.StateDir); err != nil || pending {
		t.Fatalf("parent journal written before installer validation (%v, %v)", pending, err)
	}
	absent(t, bound.Changes[0].Target.Path)
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
