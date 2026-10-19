package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tricell-hive/tooling/management"
	"tricell-hive/tooling/providers"
)

func cardOffer(t *testing.T, a *nativeProviderAdapter, hosts ...string) (providerOffer, bool) {
	t.Helper()
	offers, err := a.Detect(management.Options{Hosts: hosts})
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range offers {
		if o.ID == string(providers.CodexQuestionCard) {
			return o, true
		}
	}
	return providerOffer{}, false
}

func planCard(t *testing.T, a *nativeProviderAdapter) onboardingPreview {
	t.Helper()
	preview, err := a.Plan(management.Plan{}, []providerRequest{{ID: string(providers.CodexQuestionCard)}})
	if err != nil {
		t.Fatal(err)
	}
	return preview
}

func TestCodexQuestionCardOfferExecutableWithCodexOnPath(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}})
	offer, ok := cardOffer(t, newTestAdapter(t, false, false, true), "codex")
	if !ok {
		t.Fatal("no codex-question-card offer")
	}
	if offer.ManualOnly || offer.Unversioned || !offer.DefaultOn {
		t.Fatalf("offer = %+v, want executable, versioned-less default-on", offer)
	}
	if offer.Effect != "enables Codex's question card in Default mode" {
		t.Fatalf("effect = %q", offer.Effect)
	}
}

func TestCodexQuestionCardOfferManualWithoutCodexOnPath(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64"})
	offer, ok := cardOffer(t, newTestAdapter(t, false, false, true), "codex")
	if !ok || !offer.ManualOnly || offer.DefaultOn {
		t.Fatalf("offer = %+v ok=%v, want manual without default", offer, ok)
	}
}

func TestCodexQuestionCardOfferManualUnderSyntheticHome(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}})
	offer, ok := cardOffer(t, newTestAdapter(t, true, false, true), "codex")
	if !ok || !offer.ManualOnly || offer.DefaultOn {
		t.Fatalf("offer = %+v ok=%v, want manual without default under a synthetic home", offer, ok)
	}
}

func TestCodexQuestionCardAbsentWithoutCodexHost(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}})
	if _, ok := cardOffer(t, newTestAdapter(t, false, false, true), "claude", "pi"); ok {
		t.Fatal("offered without codex among the hosts")
	}
}

func TestCodexQuestionCardPlanExecutable(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}})
	a := newTestAdapter(t, false, false, true)
	cardOffer(t, a, "codex")
	preview := planCard(t, a)
	if got := stepIDs(preview); len(got) != 1 || got[0] != providers.CodexQuestionCardEnableStep {
		t.Fatalf("steps = %v", got)
	}
	var step providers.Step
	if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
		t.Fatal(err)
	}
	if step.Status != providers.Planned || step.Method != providers.MethodCodex || step.Provider != providers.CodexQuestionCard {
		t.Fatalf("step = %+v", step)
	}
	if strings.Join(step.Argv, " ") != "codex features enable default_mode_request_user_input" {
		t.Fatalf("argv = %v", step.Argv)
	}
	if err := providers.ValidateStep(preview.Steps[0].Payload); err != nil {
		t.Fatal(err)
	}
	if len(preview.Details) != 1 {
		t.Fatalf("details = %+v", preview.Details)
	}
	effects := strings.Join(preview.Details[0].Effects, "\n")
	for _, want := range []string{"default_mode_request_user_input", "config.toml", "codex features disable default_mode_request_user_input"} {
		if !strings.Contains(effects, want) {
			t.Fatalf("effects lack %q: %s", want, effects)
		}
	}
}

func TestCodexQuestionCardPlanManualNamesTheCommand(t *testing.T) {
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64"})
	a := newTestAdapter(t, false, false, true)
	cardOffer(t, a, "codex")
	preview := planCard(t, a)
	if got := stepIDs(preview); len(got) != 1 || got[0] != providers.CodexQuestionCardManualStep {
		t.Fatalf("steps = %v", got)
	}
	var step providers.Step
	if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"the `codex` executable was not found in PATH", "codex features enable default_mode_request_user_input"} {
		if !strings.Contains(step.ManualReason, want) {
			t.Fatalf("reason %q lacks %q", step.ManualReason, want)
		}
	}
	if err := providers.ValidateStep(preview.Steps[0].Payload); err != nil {
		t.Fatal(err)
	}
}

func TestCodexQuestionCardExecuteRunsExactArgv(t *testing.T) {
	f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	cardOffer(t, a, "codex")
	preview := planCard(t, a)
	raw, err := a.Runner().Execute(preview.Steps[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(f.execs) != 1 || strings.Join(f.execs[0], " ") != "codex features enable default_mode_request_user_input" {
		t.Fatalf("execs = %v", f.execs)
	}
	var res agentBrowserResult
	if err := json.Unmarshal(raw, &res); err != nil || res.failed() {
		t.Fatalf("result = %s err=%v", raw, err)
	}
}

func TestCodexQuestionCardExecuteMissingFeaturesSubcommand(t *testing.T) {
	for _, output := range []string{"error: unrecognized subcommand 'features'", "error: unexpected argument 'features' found"} {
		f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
		f.onExec = func(*abFake, []string) agentBrowserResult {
			return agentBrowserResult{ExitCode: 2, Output: output}
		}
		useAgentBrowserFake(t, f)
		a := newTestAdapter(t, false, false, true)
		cardOffer(t, a, "codex")
		raw, err := a.Runner().Execute(planCard(t, a).Steps[0])
		if err != nil {
			t.Fatal(err)
		}
		var res agentBrowserResult
		if err := json.Unmarshal(raw, &res); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(res.Message, "no `features` subcommand") || !strings.Contains(res.Message, "0.163.0") {
			t.Fatalf("message = %q for %q", res.Message, output)
		}
	}
}

func TestCodexQuestionCardExecuteSyntheticRunsNothing(t *testing.T) {
	f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	cardOffer(t, a, "codex")
	step := planCard(t, a).Steps[0]
	synthetic := newTestAdapter(t, true, false, true)
	raw, err := synthetic.Runner().Execute(step)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.execs) != 0 {
		t.Fatalf("a synthetic runner ran %v", f.execs)
	}
	status, err := synthetic.Runner().Reconcile(step, raw)
	if err != nil || status != management.StepSkipped {
		t.Fatalf("status = %q err=%v", status, err)
	}
}

func cardRunner(t *testing.T, codexHome string) (providerRunner, management.ExternalStep) {
	t.Helper()
	useAgentBrowserFake(t, &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}})
	a := newTestAdapter(t, false, false, true)
	a.codexHome = codexHome
	cardOffer(t, a, "codex")
	return a.Runner().(providerRunner), planCard(t, a).Steps[0]
}

func TestCodexQuestionCardReconcileReadsConfig(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("[features]\ndefault_mode_request_user_input = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, step := cardRunner(t, home)
	for name, stored := range map[string]json.RawMessage{
		"no stored result": nil,
		"empty result":     json.RawMessage(`{}`),
		"success result":   json.RawMessage(`{"exit_code":0}`),
		"failed result":    json.RawMessage(`{"exit_code":1,"message":"x"}`),
	} {
		status, err := runner.Reconcile(step, stored)
		if err != nil || status != management.StepVerified {
			t.Fatalf("%s: status = %q err=%v, want verified", name, status, err)
		}
	}
}

func TestCodexQuestionCardReconcileFailsWithoutKey(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("[features]\nother = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runner, step := cardRunner(t, home)
	for name, stored := range map[string]json.RawMessage{
		"no stored result": nil,
		"success result":   json.RawMessage(`{"exit_code":0}`),
	} {
		if status, _ := runner.Reconcile(step, stored); status != management.StepFailed {
			t.Fatalf("%s: status = %q, want failed", name, status)
		}
	}
	// An unreadable form is never a pass.
	if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte("features = { default_mode_request_user_input = true }\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if status, _ := runner.Reconcile(step, json.RawMessage(`{"exit_code":0}`)); status != management.StepFailed {
		t.Fatalf("unsupported form: status = %q, want failed", status)
	}
}

func TestCodexQuestionCardReconcileSkippedResult(t *testing.T) {
	runner, step := cardRunner(t, t.TempDir())
	status, err := runner.Reconcile(step, json.RawMessage(`{"skipped":true}`))
	if err != nil || status != management.StepSkipped {
		t.Fatalf("status = %q err=%v", status, err)
	}
}

func TestCodexQuestionCardFactoryResolvesCodexHome(t *testing.T) {
	home := t.TempDir()
	ad, err := nativeProviderAdapterFactory(onboardingInput{Options: management.Options{Scope: "user", Home: home}})
	if err != nil {
		t.Fatal(err)
	}
	canon, _ := filepath.EvalSymlinks(home)
	got := ad.(*nativeProviderAdapter).codexHome
	if got != filepath.Join(canon, ".codex") {
		t.Fatalf("codexHome = %q, want under %q", got, canon)
	}
}

func TestStepResultDetailRecognizesCodexQuestionCardPrefix(t *testing.T) {
	msg, _ := stepResultDetail(management.StepReceipt{
		Step:   management.ExternalStep{ID: providers.CodexQuestionCardEnableStep},
		Status: management.StepFailed,
	})
	if msg != "interrupted before its result was recorded" {
		t.Fatalf("message = %q", msg)
	}
}

func TestCodexQuestionCardOfferProbesCodexOnce(t *testing.T) {
	f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
	useAgentBrowserFake(t, f)
	a := newTestAdapter(t, false, false, true)
	cardOffer(t, a, "codex")
	cardOffer(t, a, "codex")
	planCard(t, a)
	if f.codexProbes != 1 {
		t.Fatalf("probes = %d, want 1", f.codexProbes)
	}
}

func TestCodexQuestionCardUnsupportedCodexIsManual(t *testing.T) {
	const reason = "update Codex to 0.163.0 or later (`codex features list` did not report `default_mode_request_user_input`) and run `codex features enable default_mode_request_user_input`"
	cases := map[string]*abFake{
		"feature missing from the list": {codexFeatures: "other_feature  stable  true\n"},
		"non-zero exit":                 {codexExit: 2, codexFeatures: "default_mode_request_user_input  under development  false\n"},
	}
	for name, f := range cases {
		f.goos, f.goarch, f.path = "darwin", "arm64", map[string]bool{"codex": true}
		useAgentBrowserFake(t, f)
		a := newTestAdapter(t, false, false, true)
		offer, ok := cardOffer(t, a, "codex")
		if !ok || !offer.ManualOnly || offer.DefaultOn || offer.Effect != "" {
			t.Fatalf("%s: offer = %+v ok=%v, want manual without default", name, offer, ok)
		}
		preview := planCard(t, a)
		var step providers.Step
		if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
			t.Fatal(err)
		}
		if step.ID != providers.CodexQuestionCardManualStep || step.ManualReason != reason {
			t.Fatalf("%s: step = %+v", name, step)
		}
		if err := providers.ValidateStep(preview.Steps[0].Payload); err != nil {
			t.Fatal(err)
		}
		if f.codexProbes != 1 {
			t.Fatalf("%s: probes = %d", name, f.codexProbes)
		}
	}
}

func TestCodexQuestionCardProbeDoesNotRunWithoutPrograms(t *testing.T) {
	for name, a := range map[string]func(t *testing.T) *nativeProviderAdapter{
		"synthetic":       func(t *testing.T) *nativeProviderAdapter { return newTestAdapter(t, true, false, true) },
		"dry run":         func(t *testing.T) *nativeProviderAdapter { return newTestAdapter(t, false, true, true) },
		"non-interactive": func(t *testing.T) *nativeProviderAdapter { return newTestAdapter(t, false, false, false) },
	} {
		f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
		useAgentBrowserFake(t, f)
		offer, ok := cardOffer(t, a(t), "codex")
		if f.codexProbes != 0 {
			t.Fatalf("%s: probe ran %d times", name, f.codexProbes)
		}
		wantManual := name == "synthetic"
		if !ok || offer.ManualOnly != wantManual {
			t.Fatalf("%s: offer = %+v ok=%v", name, offer, ok)
		}
	}
}

func executeCardEnable(t *testing.T, codexHome string, config string) agentBrowserResult {
	t.Helper()
	if config != "" {
		if err := os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte(config), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	runner, step := cardRunner(t, codexHome)
	raw, err := runner.Execute(step)
	if err != nil {
		t.Fatal(err)
	}
	var res agentBrowserResult
	if err := json.Unmarshal(raw, &res); err != nil {
		t.Fatal(err)
	}
	return res
}

func TestCodexQuestionCardExecuteUnconfirmedSettingFailsClosed(t *testing.T) {
	home := t.TempDir()
	res := executeCardEnable(t, home, "[features]\nother = true\n")
	want := "Codex reported success but Hive could not confirm the setting in " + home + "/config.toml; run `codex features list` to check it"
	if res.Message != want {
		t.Fatalf("message = %q, want %q", res.Message, want)
	}
	runner, step := cardRunner(t, home)
	if status, _ := runner.Reconcile(step, json.RawMessage(`{"exit_code":0,"message":"x"}`)); status != management.StepFailed {
		t.Fatalf("status = %q, want failed", status)
	}
}

func TestCodexQuestionCardExecuteUnreadableConfigNamesTheError(t *testing.T) {
	home := t.TempDir()
	res := executeCardEnable(t, home, "features = { default_mode_request_user_input = true }\n")
	prefix := "Codex reported success but Hive could not confirm the setting in " + home + "/config.toml: "
	if !strings.HasPrefix(res.Message, prefix) || !strings.HasSuffix(res.Message, "; run `codex features list` to check it") || !strings.Contains(res.Message, "unsupported") {
		t.Fatalf("message = %q", res.Message)
	}
}

func TestCodexQuestionCardExecuteConfirmedSettingHasNoMessage(t *testing.T) {
	home := t.TempDir()
	res := executeCardEnable(t, home, "[\"features\"]\ndefault_mode_request_user_input = true\n")
	if res.Message != "" || res.failed() {
		t.Fatalf("result = %+v", res)
	}
}

func TestCodexQuestionCardStoredOutputIsEmpty(t *testing.T) {
	secretLine := "token = \"s3cr3t-value\" at line 7"
	for name, tc := range map[string]struct {
		exit   int
		output string
	}{
		"success":             {0, secretLine},
		"failure":             {1, secretLine},
		"missing subcommand":  {2, "error: unrecognized subcommand 'features'\n" + secretLine},
		"unexpected argument": {2, "error: unexpected argument 'features' found\n" + secretLine},
	} {
		home := t.TempDir()
		f := &abFake{goos: "darwin", goarch: "arm64", path: map[string]bool{"codex": true}}
		f.onExec = func(*abFake, []string) agentBrowserResult {
			return agentBrowserResult{ExitCode: tc.exit, Output: tc.output}
		}
		useAgentBrowserFake(t, f)
		a := newTestAdapter(t, false, false, true)
		a.codexHome = home
		cardOffer(t, a, "codex")
		raw, err := a.Runner().Execute(planCard(t, a).Steps[0])
		if err != nil {
			t.Fatal(err)
		}
		var res agentBrowserResult
		if err := json.Unmarshal(raw, &res); err != nil {
			t.Fatal(err)
		}
		if res.Output != "" || strings.Contains(string(raw), "s3cr3t") {
			t.Fatalf("%s: stored result keeps output: %s", name, raw)
		}
		if res.Message == "" {
			t.Fatalf("%s: no message stored", name)
		}
	}
}
