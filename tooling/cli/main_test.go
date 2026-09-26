package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// stagingFailureRunner deliberately fails Execute, so an onboarding built
// with it is left pending (status "unknown") without ever reaching
// Reconcile, exactly as production onboarding does when a real provider
// step is interrupted. It exists only to stage that pending state; the
// actual recover path under test always uses the real native adapter.
type stagingFailureRunner struct{}

func (stagingFailureRunner) Validate(management.ExternalStep) error { return nil }
func (stagingFailureRunner) Execute(management.ExternalStep) (json.RawMessage, error) {
	return nil, fmt.Errorf("fixture: simulate an interrupted provider step")
}
func (stagingFailureRunner) Reconcile(management.ExternalStep, json.RawMessage) (string, error) {
	return "", fmt.Errorf("staging runner must not be asked to reconcile")
}

// TestMainRecoverReconcilesPendingOnboardingFromAnotherTerminal covers H2:
// `hive recover` used to build its adapter from an Options value with an
// empty Scope, so NormalizeOptions always failed and no pending onboarding
// could ever be recovered from a second terminal. HOME is overridden so the
// command's real-home defaulting (no --home flag exists for recover) stays
// inside this test's sandbox.
func TestMainRecoverReconcilesPendingOnboardingFromAnotherTerminal(t *testing.T) {
	// Resolved once so every path this test compares (HOME, the state dir it
	// builds onboarding under, and the one it later queries) names the same
	// canonical directory that management itself resolves internally; on
	// macOS t.TempDir() sits under /var, a symlink to /private/var.
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	stateDir := filepath.Join(home, "state")
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(providers.Step{ID: "context7-manual", Provider: providers.Context7, Status: providers.Manual, ManualReason: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := (management.Engine{}).Onboard(p, []management.ExternalStep{{ID: "context7-manual", Payload: payload}}, stagingFailureRunner{}); err == nil {
		t.Fatal("expected the staged provider step to fail, leaving onboarding pending")
	}
	if pending, err := management.OnboardingPending(stateDir); err != nil || !pending {
		t.Fatalf("expected a pending onboarding to recover: pending=%v err=%v", pending, err)
	}
	if err := run([]string{"recover", "--state-dir", stateDir}); err != nil {
		t.Fatalf("recover from another terminal failed: %v", err)
	}
	pendingAfter, err := management.OnboardingPending(stateDir)
	if err != nil {
		t.Fatal(err)
	}
	if pendingAfter {
		t.Fatal("onboarding still pending after recover")
	}
}

func TestRejectIgnoredDestinationOptionsOnApply(t *testing.T) {
	if err := run([]string{"apply", "--plan", "anything", "--home", "/temporary"}); err == nil {
		t.Fatal("apply silently ignored alternate home")
	}
}
func TestHelpAndRequiredScope(t *testing.T) {
	if err := run([]string{"--help"}); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"plan", "install", "--hosts", "codex"}); err == nil {
		t.Fatal("missing scope accepted")
	}
}
