package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/management"
)

// copyTree copies a directory tree verbatim, used to build a minimal --source
// checkout (only the subtrees BuildPlan actually reads) without touching the
// real repository checkout that other tests point --source at.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// coreOnlyAdapterFactory and coreOnlyAdapter are test-only doubles (M7):
// nothing in production code builds an onboardingAdapter that offers no
// capability at all, but many wizard tests here want the host-selection and
// plan/apply flow exercised without any optional-capability noise.
func coreOnlyAdapterFactory(onboardingInput) (onboardingAdapter, error) {
	return coreOnlyAdapter{}, nil
}

type coreOnlyAdapter struct{}

func (coreOnlyAdapter) Detect(management.Options) ([]providerOffer, error) { return nil, nil }
func (coreOnlyAdapter) Plan(management.Plan, []providerRequest) (onboardingPreview, error) {
	return onboardingPreview{}, nil
}
func (coreOnlyAdapter) Runner() management.ExternalRunner { return nil }

type testExternalRunner struct{}

func (testExternalRunner) Validate(management.ExternalStep) error { return nil }
func (testExternalRunner) Execute(management.ExternalStep) (json.RawMessage, error) {
	return json.RawMessage(`{}`), nil
}
func (testExternalRunner) Reconcile(management.ExternalStep, json.RawMessage) (string, error) {
	return "verified", nil
}

type testOnboardingAdapter struct {
	offers   []providerOffer
	requests []providerRequest
	runner   management.ExternalRunner
}

func (a *testOnboardingAdapter) Detect(management.Options) ([]providerOffer, error) {
	return a.offers, nil
}
func (a *testOnboardingAdapter) Plan(_ management.Plan, requests []providerRequest) (onboardingPreview, error) {
	a.requests = append([]providerRequest(nil), requests...)
	if len(requests) == 0 {
		return onboardingPreview{}, nil
	}
	return onboardingPreview{
		Steps:   []management.ExternalStep{{ID: "fixture", Payload: json.RawMessage(`{}`)}},
		Details: []providerDetail{{ID: requests[0].ID, Version: requests[0].Version, Source: "official fixture", Effects: []string{"writes fixture state"}}},
	}, nil
}
func (a *testOnboardingAdapter) Runner() management.ExternalRunner { return a.runner }

func TestInstallDryRunDoesNotCreateStateOrDestinations(t *testing.T) {
	home := t.TempDir()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"install", "--home", home, "--hosts", "codex", "--source", source, "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("dry run wrote into home: %v", entries)
	}
}

func installArgs(t *testing.T, home string) []string {
	t.Helper()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return []string{"--home", home, "--hosts", "codex", "--source", source}
}

func TestInstallConfirmationBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, answer   string
		tty, wantError bool
	}{
		{"cancel", "\nn\n", true, false}, {"default", "\n", true, false}, {"eof", "y", true, false}, {"nonterminal", "y\n", false, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			var out bytes.Buffer
			err := install(installArgs(t, home), strings.NewReader(test.answer), &out, test.tty)
			if (err != nil) != test.wantError {
				t.Fatalf("err=%v output=%s", err, out.String())
			}
			entries, err := os.ReadDir(home)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("unconfirmed install wrote files: %v", entries)
			}
		})
	}
}

func TestInstallUnchangedStillConfirms(t *testing.T) {
	home := t.TempDir()
	var first bytes.Buffer
	args := installArgs(t, home)
	if err := install(args, strings.NewReader("\ny\n"), &first, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(first.String(), "Open new CLI sessions") {
		t.Fatalf("missing restart instruction: %s", first.String())
	}
	instruction := filepath.Join(home, ".codex", "AGENTS.md")
	before, err := os.ReadFile(instruction)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(instruction)
	if err != nil {
		t.Fatal(err)
	}
	var second bytes.Buffer
	if err := install(args, strings.NewReader("\ny\n"), &second, true); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(instruction)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(instruction)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) || !info.ModTime().Equal(afterInfo.ModTime()) {
		t.Fatal("no-op rewrote installed file")
	}
	if !strings.Contains(second.String(), "Hive's core is already up to date") || !strings.Contains(second.String(), "Apply these changes?") {
		t.Fatalf("unchanged core skipped confirmation: %s", second.String())
	}
	// U10: the final result identifier must not print the bare literal
	// "(unchanged)" management.Apply uses internally as its sentinel ID.
	if strings.Contains(second.String(), "(unchanged)") {
		t.Fatalf("leaked the internal sentinel ID: %s", second.String())
	}
	if !strings.Contains(second.String(), "(no changes)") {
		t.Fatalf("missing user-facing wording for the idempotent reinstall result: %s", second.String())
	}
}

func TestInstallWizardCancelsEmptySelection(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.DiscoverHosts = func(management.Options) ([]hostCandidate, error) {
		return []hostCandidate{{Name: "codex", Detected: true}}, nil
	}
	if err = installWithDependencies([]string{"--home", home, "--source", source}, strings.NewReader("\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("empty selection wrote to home: %v %v", entries, err)
	}
	if !strings.Contains(out.String(), "1. codex") || !strings.Contains(out.String(), "Cancelled. No changes applied.") {
		t.Fatal(out.String())
	}
}

func TestInstallWizardBackToHostSelectionKeepsBufferedAnswers(t *testing.T) {
	home := t.TempDir()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.DiscoverHosts = func(management.Options) ([]hostCandidate, error) {
		return []hostCandidate{{Name: "codex", Detected: true}, {Name: "claude", Detected: true}}, nil
	}
	var out bytes.Buffer
	if err := installWithDependencies([]string{"--home", home, "--source", source}, strings.NewReader("1\nb\n2\ny\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatalf("back selection did not install Claude: %v\n%s", err, out.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("back selection retained Codex: %v\n%s", err, out.String())
	}
}

func TestInstallSharedHostClosureRequiresConsent(t *testing.T) {
	home := t.TempDir()
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.RequiredHosts = func(management.Options) ([]string, error) {
		return []string{"codex", "opencode"}, nil
	}
	var out bytes.Buffer
	err := installWithDependencies(installArgs(t, home), strings.NewReader("n\n"), &out, true, dependencies)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("declined shared closure wrote to home: %v %v", entries, err)
	}
	if !strings.Contains(out.String(), "Shared resources require selecting: opencode") {
		t.Fatal(out.String())
	}
	// U8: name the affected shared resource (from the plan data already at
	// hand) and the consequence of accepting, not only the extra host name.
	if !strings.Contains(out.String(), "Affected shared resources:") || !strings.Contains(out.String(), ".agents") {
		t.Fatalf("missing named shared resource: %s", out.String())
	}
	if !strings.Contains(out.String(), "rewrite those resources too, on already-installed hosts that share them") {
		t.Fatalf("missing consequence statement: %s", out.String())
	}
}

func TestInstallRecoversOnboardingBeforeCorePending(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pending.json"), []byte(`{"id":"core"}`), 0600); err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.Pending = func(string) (management.PendingKind, error) { return management.PendingOnboarding, nil }
	adapter := &testOnboardingAdapter{runner: testExternalRunner{}}
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) {
		if adapter.Runner() == nil {
			t.Fatal("onboarding recovery did not receive configured adapter and state")
		}
		return adapter, nil
	}
	calledOnboarding := false
	dependencies.RecoverOnboarding = func(_ string, got onboardingAdapter) (management.OnboardingResult, error) {
		calledOnboarding = true
		if got.Runner() == nil {
			t.Fatal("wrong onboarding recovery hook")
		}
		return management.OnboardingResult{ID: "onboarding", Phase: "partial"}, nil
	}
	dependencies.RecoverCore = func(string) (string, error) {
		t.Fatal("core recovery ran before onboarding recovery")
		return "", nil
	}
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := installWithDependencies([]string{"--home", home, "--state-dir", stateDir, "--source", source}, strings.NewReader("y\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if !calledOnboarding || !strings.Contains(out.String(), "Onboarding recovery") {
		t.Fatalf("onboarding recovery was skipped: %s", out.String())
	}
}

func TestInstallPreviewsExactProviderBeforeConfirmation(t *testing.T) {
	home := t.TempDir()
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture", Effects: []string{"installs the CLI"}}},
		runner: testExternalRunner{},
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(input onboardingInput) (onboardingAdapter, error) {
		if input.DryRun || len(input.Options.Hosts) != 1 || input.Options.Hosts[0] != "codex" {
			t.Fatal("adapter factory did not receive selected install context")
		}
		return adapter, nil
	}
	var out bytes.Buffer
	if err := installWithDependencies(installArgs(t, home), strings.NewReader("1\n0.5.11\ny\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	if len(adapter.requests) != 1 || adapter.requests[0] != (providerRequest{ID: "context7", Version: "0.5.11"}) {
		t.Fatalf("provider request = %#v", adapter.requests)
	}
	selected := strings.Index(out.String(), "Select optional capabilities")
	confirmed := strings.Index(out.String(), "Apply these changes?")
	if selected < 0 || confirmed < 0 || selected > confirmed || !strings.Contains(out.String(), "Optional capability: context7 0.5.11 from official fixture") {
		t.Fatal(out.String())
	}
}

// TestInstallRefusesStalePlanChangedBeforeConfirmation covers T3's staleness
// requirement: the plan captured at the top of the wizard loop must not be
// applied once shared state has moved on, even though the wizard itself
// never re-reads it between building the summary and confirming. The mutation
// runs from the adapter-factory hook, the same point in the flow where a real
// external-provider lookup would also observe current state, right after the
// plan is built and before the summary/confirmation are shown.
func TestInstallRefusesStalePlanChangedBeforeConfirmation(t *testing.T) {
	home := t.TempDir()
	stateDir := filepath.Join(home, "state")
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	mutated := false
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(input onboardingInput) (onboardingAdapter, error) {
		if !mutated {
			mutated = true
			// An unrelated concurrent change to the same shared state
			// (installing a different host) must invalidate the plan this
			// wizard iteration already captured.
			other := management.Options{Scope: "user", Home: home, StateDir: stateDir, Source: source, Hosts: []string{"claude"}}
			op, err := management.BuildPlan("install", other)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := (management.Engine{}).Apply(op); err != nil {
				t.Fatal(err)
			}
		}
		return coreOnlyAdapterFactory(input)
	}
	args := []string{"--home", home, "--hosts", "codex", "--source", source, "--state-dir", stateDir}
	var out bytes.Buffer
	if err = installWithDependencies(args, strings.NewReader("y\n"), &out, true, dependencies); err == nil {
		t.Fatal("expected refusal to apply a plan made stale by a concurrent change")
	}
	if !mutated {
		t.Fatal("mutation hook never ran; test did not exercise staleness")
	}
	if _, statErr := os.Stat(filepath.Join(home, ".codex", "AGENTS.md")); !os.IsNotExist(statErr) {
		t.Fatalf("stale plan still wrote codex destinations: %v", statErr)
	}
	if _, statErr := os.Stat(filepath.Join(home, ".claude", "CLAUDE.md")); statErr != nil {
		t.Fatalf("unrelated concurrent installation did not apply: %v", statErr)
	}
}

// developmentSourceArgs builds a --source checkout containing only content/
// and integrations/ (enough for BuildPlan), deliberately without VERSION or
// release.json — the repro in D1: management.productFromSource returns a nil
// *ProductIdentity for such a checkout, identifying a development build.
func developmentSourceArgs(t *testing.T, home string) []string {
	t.Helper()
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	source := t.TempDir()
	// Resolve macOS's /var -> /private/var symlink before use: unlike --home
	// (canonicalized by management.NormalizeOptions), --source is walked
	// as-is by collectResources's target.Safe check, which rejects any
	// symlink ancestor.
	resolvedSource, err := filepath.EvalSymlinks(source)
	if err != nil {
		t.Fatal(err)
	}
	copyTree(t, filepath.Join(repo, "content"), filepath.Join(resolvedSource, "content"))
	copyTree(t, filepath.Join(repo, "integrations"), filepath.Join(resolvedSource, "integrations"))
	return []string{"--home", home, "--hosts", "codex", "--source", resolvedSource, "--dry-run"}
}

// TestInstallSummaryDevelopmentBuildWithoutProductIdentity covers D1: a
// --source checkout without VERSION or release.json must be shown as a
// development build, never crash showInstallSummary with a nil p.Product
// dereference.
func TestInstallSummaryDevelopmentBuildWithoutProductIdentity(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	if err := install(developmentSourceArgs(t, home), strings.NewReader("\n"), &out, true); err != nil {
		t.Fatalf("development build install errored instead of showing a development summary: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "development build") {
		t.Fatalf("missing development-build label: %s", out.String())
	}
}

// TestInstallUnchangedSummaryOmitsProviderMentionWithoutSelection covers U3:
// an unchanged core with no optional capability selected must not also claim
// that "selected" optional providers still need confirmation.
func TestInstallUnchangedSummaryOmitsProviderMentionWithoutSelection(t *testing.T) {
	home := t.TempDir()
	args := installArgs(t, home)
	if err := install(args, strings.NewReader("\ny\n"), &bytes.Buffer{}, true); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := install(args, strings.NewReader("\n"), &out, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Hive's core is already up to date") {
		t.Fatalf("missing unchanged-core message: %s", out.String())
	}
	if !strings.Contains(out.String(), "Hive files checked:") || strings.Contains(out.String(), "files to install or update") {
		t.Fatalf("an unchanged install must not announce files to install: %s", out.String())
	}
	if strings.Contains(out.String(), "optional capabilities still require confirmation") {
		t.Fatalf("mentioned optional providers although none were selected: %s", out.String())
	}
	if !strings.Contains(out.String(), "Optional capabilities: none selected.") {
		t.Fatalf("missing no-capabilities line: %s", out.String())
	}
}

// TestInstallSummaryUsesUserFacingResourceWording covers U5: the resource
// count line must not leak the internal "Rebuild resources" build term.
func TestInstallSummaryUsesUserFacingResourceWording(t *testing.T) {
	home := t.TempDir()
	var out bytes.Buffer
	if err := install(installArgs(t, home), strings.NewReader("\ny\n"), &out, true); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Hive files to install or update:") {
		t.Fatalf("missing user-facing resource wording: %s", out.String())
	}
	if strings.Contains(out.String(), "Rebuild resources") {
		t.Fatalf("leaked internal build wording: %s", out.String())
	}
	assertTerminalWidth(t, out.String(), home)
}

// assertTerminalWidth fails when a line of fixed installer text exceeds 80
// columns; lines naming the synthetic home are skipped because their length
// depends on the test's temporary path.
func assertTerminalWidth(t *testing.T, output, home string) {
	t.Helper()
	// A terminal echoes the answer and its newline after each prompt; a
	// buffer does not, so restore that break before measuring.
	for _, prompt := range []string{"to skip: ", "to cancel: ", "[N] cancel ", "[y/N] "} {
		output = strings.ReplaceAll(output, prompt, prompt+"\n")
	}
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, home) && len(line) > 80 {
			t.Fatalf("line exceeds 80 columns (%d): %q", len(line), line)
		}
	}
}

// TestInstallSummaryWrapsLongEffectLines covers U7: a long capability effect
// line (a translated ManualReason concatenated across several steps) must be
// wrapped so every printed line stays within a plain 80-column terminal.
func TestInstallSummaryWrapsLongEffectLines(t *testing.T) {
	home := t.TempDir()
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture"}},
		runner: testExternalRunner{},
	}
	longEffect := strings.Repeat("sample effect text long enough to force a line wrap ", 4)
	adapter2 := &longEffectAdapter{testOnboardingAdapter: adapter, effect: longEffect}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter2, nil }
	var out bytes.Buffer
	if err := installWithDependencies(installArgs(t, home), strings.NewReader("1\n0.5.11\ny\n"), &out, true, dependencies); err != nil {
		t.Fatal(err)
	}
	// Check only the wrapped requirement/effect bullets (printed with a
	// 4-space indent): unrelated lines such as the state-dir backup path are
	// filesystem paths outside U7's scope and cannot be safely wrapped.
	wrapped := false
	for _, line := range strings.Split(out.String(), "\n") {
		if !strings.HasPrefix(line, "    ") {
			continue
		}
		wrapped = true
		if len(line) > 80 {
			t.Fatalf("wrapped line exceeds 80 columns (%d): %q", len(line), line)
		}
	}
	if !wrapped {
		t.Fatalf("no wrapped requirement/effect line found: %s", out.String())
	}
	if !strings.Contains(out.String(), "sample effect") {
		t.Fatalf("wrapped effect text missing: %s", out.String())
	}
}

// TestInstallWizardRepromptsInvalidHostSelection covers U1: an invalid host
// selection (out of range, non-numeric, duplicate) must re-prompt, showing
// the CLI's existing English error text, and keep the wizard alive instead
// of ending the whole process with an error.
func TestInstallWizardRepromptsInvalidHostSelection(t *testing.T) {
	home := t.TempDir()
	source, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.DiscoverHosts = func(management.Options) ([]hostCandidate, error) {
		return []hostCandidate{{Name: "codex", Detected: true}, {Name: "claude", Detected: true}}, nil
	}
	var out bytes.Buffer
	// "abc" (non-numeric), then "9" (out of range), then "1,1" (duplicate),
	// then a valid selection, then decline the final confirmation so the
	// test only exercises the selection loop.
	answers := "abc\n9\n1,1\n1\nn\n"
	if err := installWithDependencies([]string{"--home", home, "--source", source}, strings.NewReader(answers), &out, true, dependencies); err != nil {
		t.Fatalf("invalid host entries ended the process instead of re-prompting: %v\n%s", err, out.String())
	}
	if got := strings.Count(out.String(), "select host numbers from 1 to 2"); got != 2 {
		t.Fatalf("expected 2 re-prompts for the non-numeric/out-of-range entries, got %d: %s", got, out.String())
	}
	if got := strings.Count(out.String(), "duplicate host selection"); got != 1 {
		t.Fatalf("expected 1 re-prompt for the duplicate entry, got %d: %s", got, out.String())
	}
	if !strings.Contains(out.String(), "Apply these changes?") {
		t.Fatalf("valid selection after retries never reached confirmation: %s", out.String())
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("declined confirmation wrote to home: %v %v", entries, err)
	}
}

// TestInstallWizardRepromptsInvalidCapabilitySelectionAndEmptyVersion covers
// U1 for the capability step: an invalid capability index list re-prompts,
// and an empty exact version (Enter, no default available) re-prompts for
// that one capability's version while keeping the earlier capability's
// already-entered version.
func TestInstallWizardRepromptsInvalidCapabilitySelectionAndEmptyVersion(t *testing.T) {
	home := t.TempDir()
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{
			{ID: "engram", Name: "Engram", Source: "official fixture"},
			{ID: "context7", Name: "Context7", Source: "official fixture"},
		},
		runner: testExternalRunner{},
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	var out bytes.Buffer
	// "x" then "9" invalid capability lists; "1,2" valid; version "1.0.0" for
	// Engram; an empty line (invalid: no default) then "2.0.0" for Context7.
	answers := "x\n9\n1,2\n1.0.0\n\n2.0.0\ny\n"
	if err := installWithDependencies(installArgs(t, home), strings.NewReader(answers), &out, true, dependencies); err != nil {
		t.Fatalf("invalid capability entries ended the process instead of re-prompting: %v\n%s", err, out.String())
	}
	if got := strings.Count(out.String(), "select capability numbers from 1 to 2 without duplicates"); got != 2 {
		t.Fatalf("expected 2 re-prompts for the invalid capability lists, got %d: %s", got, out.String())
	}
	if got := strings.Count(out.String(), "exact version required for Context7"); got != 1 {
		t.Fatalf("expected 1 re-prompt for the empty version, got %d: %s", got, out.String())
	}
	if len(adapter.requests) != 2 || adapter.requests[0] != (providerRequest{ID: "engram", Version: "1.0.0"}) || adapter.requests[1] != (providerRequest{ID: "context7", Version: "2.0.0"}) {
		t.Fatalf("provider requests lost prior answers across retries: %#v", adapter.requests)
	}
}

// TestInstallWizardEOFDuringVersionEntryCancelsSafely covers U1's EOF
// carve-out: an actual end of input while entering a capability's version
// (not merely an empty line) must cancel the whole install cleanly, without
// error and without writes, exactly like EOF at the host or confirmation
// prompts already does.
func TestInstallWizardEOFDuringVersionEntryCancelsSafely(t *testing.T) {
	home := t.TempDir()
	adapter := &testOnboardingAdapter{
		offers: []providerOffer{{ID: "context7", Name: "Context7", Source: "official fixture"}},
		runner: testExternalRunner{},
	}
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return adapter, nil }
	var out bytes.Buffer
	// Selects capability 1, then the input stream ends mid-version-entry
	// (no trailing newline): a real EOF, not an empty line.
	if err := installWithDependencies(installArgs(t, home), strings.NewReader("1\n0.5"), &out, true, dependencies); err != nil {
		t.Fatalf("EOF during version entry returned an error instead of cancelling safely: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "Cancelled. No changes applied.") {
		t.Fatalf("missing cancellation message: %s", out.String())
	}
	entries, err := os.ReadDir(home)
	if err != nil || len(entries) != 0 {
		t.Fatalf("EOF cancellation wrote to home: %v %v", entries, err)
	}
}

type longEffectAdapter struct {
	*testOnboardingAdapter
	effect string
}

func (a *longEffectAdapter) Plan(p management.Plan, requests []providerRequest) (onboardingPreview, error) {
	preview, err := a.testOnboardingAdapter.Plan(p, requests)
	if err != nil || len(preview.Details) == 0 {
		return preview, err
	}
	preview.Details[0].Effects = []string{a.effect}
	return preview, nil
}

func TestInstallUnchangedCoreWithCapabilityFitsTerminal(t *testing.T) {
	home := t.TempDir()
	if err := install(installArgs(t, home), strings.NewReader("\ny\n"), io.Discard, true); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	err := install(installArgs(t, home), strings.NewReader("1\ny\n"), &out, true)
	if err == nil {
		t.Fatal("a manual capability must leave the result partial")
	}
	if !strings.Contains(out.String(), "Hive's core is already up to date.") {
		t.Fatalf("missing unchanged-core line: %s", out.String())
	}
	assertTerminalWidth(t, out.String(), home)
	if line := "hive: " + err.Error(); len(line) > 80 {
		t.Fatalf("error line exceeds 80 columns (%d): %q", len(line), line)
	}
}
