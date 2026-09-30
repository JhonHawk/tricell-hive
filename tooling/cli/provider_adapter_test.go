package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

// readOnboardingReceipt reads the single receipt a finished (or partial)
// onboarding leaves under <state-dir>/onboarding/. Unknown outer fields
// (version, installer, core_id, integrity) are intentionally ignored; only
// the exported OnboardingResult shape is asserted here.
func readOnboardingReceipt(t *testing.T, stateDir string) management.OnboardingResult {
	t.Helper()
	dir := filepath.Join(stateDir, "onboarding")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one onboarding receipt in %s, got %d", dir, len(entries))
	}
	data, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var result management.OnboardingResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}
func stepStatus(t *testing.T, result management.OnboardingResult, id string) string {
	t.Helper()
	for _, s := range result.Steps {
		if s.Step.ID == id {
			return s.Status
		}
	}
	t.Fatalf("missing step %q in %+v", id, result)
	return ""
}

// TestNativeAdapterManualOfferRunsNoProcess covers "manual offer runs no
// process": no capability in providers.Catalog has passed its own native
// validation gate in this build (M2), so selecting one must never invoke any
// external command — there is no providers.System left to inject one
// through — and must instead leave a single Manual step, disclosed in the
// preview before confirmation and in the onboarding receipt afterward.
func TestNativeAdapterManualOfferRunsNoProcess(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(nativeProviderAdapterFactory)
	args := []string{"--home", home, "--hosts", "codex", "--source", source, "--state-dir", stateDir}
	var out bytes.Buffer
	// "2" selects Context7 (offers are Engram, then Context7 for host codex);
	// no version line: every capability is unconditionally manual (U6).
	err := installWithDependencies(args, strings.NewReader("2\ny\n"), &out, true, dependencies)
	if err == nil {
		t.Fatal("expected a partial outcome; a manual step never verifies itself")
	}
	if !strings.Contains(out.String(), "Native install/host-integration validation is pending") {
		t.Fatalf("preview did not disclose the manual reason: %s", out.String())
	}
	if strings.Contains(out.String(), "Exact version for") {
		t.Fatalf("prompted for a version although every capability is unconditionally manual: %s", out.String())
	}
	if _, statErr := os.Stat(filepath.Join(home, ".local", "bin", "ctx7")); !os.IsNotExist(statErr) {
		t.Fatalf("a manual-only capability ran an installer: %v", statErr)
	}
	result := readOnboardingReceipt(t, stateDir)
	if status := stepStatus(t, result, "context7-manual"); status != "manual" {
		t.Fatalf("manual step status = %q, want manual", status)
	}
}

// TestInstallPartialOnboardingListsPerStepDetail covers "partial result
// lists each capability with reason and next action and exits non-zero".
func TestInstallPartialOnboardingListsPerStepDetail(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(nativeProviderAdapterFactory)
	args := []string{"--home", home, "--hosts", "codex", "--source", source, "--state-dir", stateDir}
	var out bytes.Buffer
	// "1" selects Engram alone, so a single manual step drives the outcome;
	// no version line, since Engram is unconditionally manual here (U6).
	err := installWithDependencies(args, strings.NewReader("1\ny\n"), &out, true, dependencies)
	if err == nil {
		t.Fatal("expected a non-nil error (non-zero exit) for a partial outcome")
	}
	if !strings.Contains(out.String(), "Partial installation") {
		t.Fatalf("missing partial summary: %s", out.String())
	}
	if !strings.Contains(out.String(), "engram-manual: manual") {
		t.Fatalf("missing per-step detail: %s", out.String())
	}
	// U4: the final partial result must repeat the human reason and a next
	// action per step, not only the raw "id: status" pair.
	if !strings.Contains(out.String(), "Reason:") || !strings.Contains(out.String(), "Next action:") {
		t.Fatalf("missing human reason or next action for the pending step: %s", out.String())
	}
	partial := out.String()[strings.Index(out.String(), "Partial installation"):]
	for _, line := range strings.Split(partial, "\n") {
		if len(line) > 80 {
			t.Fatalf("partial result line exceeds 80 columns (%d): %q", len(line), line)
		}
	}
}

// TestManualOnlyCapabilitySkipsVersionPromptAndLabelsItself covers U6: every
// provider capability is unconditionally manual in this build, so the
// wizard must label it "manual instructions only" in the capability list
// and never demand an exact version for it.
func TestManualOnlyCapabilitySkipsVersionPromptAndLabelsItself(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	source := minimalTestSource(t)
	dependencies := defaultInstallDependencies(nativeProviderAdapterFactory)
	args := []string{"--home", home, "--hosts", "codex", "--source", source, "--state-dir", stateDir}
	var out bytes.Buffer
	if err := installWithDependencies(args, strings.NewReader("1\ny\n"), &out, true, dependencies); err == nil {
		t.Fatal("expected a partial outcome; a manual step never verifies itself")
	}
	if !strings.Contains(out.String(), "manual instructions only") {
		t.Fatalf("capability list did not mark the manual-only capability: %s", out.String())
	}
}

// countingRunner wraps another ExternalRunner and counts Execute calls, so a
// recovery test can assert an interrupted step is reconciled, never replayed.
type countingRunner struct {
	inner    management.ExternalRunner
	executes int
}

func (r *countingRunner) Validate(s management.ExternalStep) error { return r.inner.Validate(s) }
func (r *countingRunner) Execute(s management.ExternalStep) (json.RawMessage, error) {
	r.executes++
	return r.inner.Execute(s)
}
func (r *countingRunner) Reconcile(s management.ExternalStep, result json.RawMessage) (string, error) {
	return r.inner.Reconcile(s, result)
}

// executeThenCrashRunner performs the real external effect through its inner
// runner and then ends its own goroutine with runtime.Goexit before
// returning. Goexit runs every deferred call up the stack (including
// Engine.Onboard's own lock release) without letting Onboard's caller
// continue past the Execute call, which is exactly what a real process crash
// between "running" and the next persisted journal write looks like from the
// journal's perspective: the step is left exactly as "running" on disk.
type executeThenCrashRunner struct {
	inner    management.ExternalRunner
	executed bool
}

func (r *executeThenCrashRunner) Validate(s management.ExternalStep) error {
	return r.inner.Validate(s)
}
func (r *executeThenCrashRunner) Execute(s management.ExternalStep) (json.RawMessage, error) {
	_, _ = r.inner.Execute(s)
	r.executed = true
	runtime.Goexit()
	return nil, nil // unreachable; Goexit never returns to its caller.
}
func (r *executeThenCrashRunner) Reconcile(s management.ExternalStep, result json.RawMessage) (string, error) {
	return r.inner.Reconcile(s, result)
}

// TestProviderRunnerRecoveryReconcilesRunningStepWithoutReexecuting covers
// "onboarding recovery through the real path": after a step is interrupted
// while "running", recovery must reconcile it back to its terminal Manual
// status without calling Execute again. This drives manualProviderRunner
// (this package's ExternalRunner) directly against the real
// management.Engine onboarding/recovery contract.
func TestProviderRunnerRecoveryReconcilesRunningStepWithoutReexecuting(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	source := minimalTestSource(t)
	o := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: []string{"codex"}}
	p, err := management.BuildPlan("install", o)
	if err != nil {
		t.Fatal(err)
	}
	step := providers.Step{ID: "context7-manual", Provider: providers.Context7, Status: providers.Manual, ManualReason: "fixture reason"}
	payload, err := json.Marshal(step)
	if err != nil {
		t.Fatal(err)
	}
	extStep := management.ExternalStep{ID: step.ID, Payload: payload}
	counting := &countingRunner{inner: manualProviderRunner{}}
	crashRunner := &executeThenCrashRunner{inner: counting}
	done := make(chan struct{})
	go func() {
		defer close(done)
		management.Engine{}.Onboard(p, []management.ExternalStep{extStep}, crashRunner)
	}()
	<-done
	if !crashRunner.executed {
		t.Fatal("crash simulation never reached Execute")
	}
	executesBeforeRecovery := counting.executes
	result, err := management.Engine{}.RecoverOnboarding(stateDir, counting)
	if err != nil {
		t.Fatal(err)
	}
	if counting.executes != executesBeforeRecovery {
		t.Fatalf("recovery re-executed the provider: %d execute(s) before, %d after", executesBeforeRecovery, counting.executes)
	}
	if len(result.Steps) != 1 || result.Steps[0].Status != string(providers.Manual) {
		t.Fatalf("recovery did not reconcile via Reconcile: %+v", result.Steps)
	}
	if _, statErr := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); statErr != nil {
		t.Fatalf("core installation did not complete before providers: %v", statErr)
	}
}
