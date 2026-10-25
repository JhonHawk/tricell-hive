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

const (
	floorApplied  = "[features.multi_agent_v2]\nmin_wait_timeout_ms = 300000\ndefault_wait_timeout_ms = 600000\n"
	floorConflict = "token = \"hunter2\"\n[features.multi_agent_v2]\nmin_wait_timeout_ms = 1000\n"
)

// floorFixture returns an adapter wired to a temporary Codex home holding
// config (no file when config is empty) and a fake machine where codex reports
// version (not on PATH when version is empty).
func floorFixture(t *testing.T, config, version string, synthetic, dry, interactive bool) (*nativeProviderAdapter, *abFake, string) {
	t.Helper()
	f := &abFake{goos: "darwin", goarch: "arm64", codexVersion: version, path: map[string]bool{}}
	if version != "" {
		f.path["codex"] = true
	}
	useAgentBrowserFake(t, f)
	home := t.TempDir()
	if config != "" {
		if err := os.WriteFile(filepath.Join(home, "config.toml"), []byte(config), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a := newTestAdapter(t, synthetic, dry, interactive)
	a.codexHome = home
	return a, f, home
}

func floorOffer(t *testing.T, a *nativeProviderAdapter, o management.Options) (providerOffer, bool) {
	t.Helper()
	offers, err := a.Detect(o)
	if err != nil {
		t.Fatal(err)
	}
	for _, offer := range offers {
		if offer.ID == string(providers.CodexWaitFloor) {
			return offer, true
		}
	}
	return providerOffer{}, false
}

var codexUser = management.Options{Scope: "user", Hosts: []string{"codex"}}

func TestWaitFloorAlreadyAppliedIsNotOfferedAndNeverProbesTheVersion(t *testing.T) {
	for name, mode := range map[string][3]bool{
		"interactive": {false, false, true},
		"dry run":     {false, true, true},
		"synthetic":   {true, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			a, f, _ := floorFixture(t, floorApplied, "codex-cli 0.164.0", mode[0], mode[1], mode[2])
			if _, ok := floorOffer(t, a, codexUser); ok {
				t.Fatal("offered although the table is already applied")
			}
			if f.versionProbes != 0 {
				t.Fatalf("version probed %d times", f.versionProbes)
			}
		})
	}
}

func TestWaitFloorOfferedWithCodex0164(t *testing.T) {
	a, f, _ := floorFixture(t, "", "codex-cli 0.164.0\n", false, false, true)
	offer, ok := floorOffer(t, a, codexUser)
	if !ok {
		t.Fatal("no offer with Codex 0.164.0")
	}
	if offer.ManualOnly || !offer.DefaultOn || offer.Unversioned || offer.Effect == "" {
		t.Fatalf("offer = %+v, want executable-free default-on with an effect", offer)
	}
	if f.versionProbes != 1 {
		t.Fatalf("version probed %d times, want 1", f.versionProbes)
	}
}

func TestWaitFloorNotOfferedBelow0164OrWithoutCodexNotEvenAsManual(t *testing.T) {
	for name, version := range map[string]string{
		"0.163.9":      "codex-cli 0.163.9\n",
		"0.99.0":       "codex-cli 0.99.0\n",
		"unparseable":  "something else\n",
		"not on PATH":  "",
		"empty output": "\n",
	} {
		for _, config := range []string{"", floorConflict} {
			t.Run(name, func(t *testing.T) {
				a, f, _ := floorFixture(t, config, version, false, false, true)
				if name == "empty output" {
					f.path["codex"] = true
				}
				if _, ok := floorOffer(t, a, codexUser); ok {
					t.Fatal("offered")
				}
				preview, err := a.Plan(management.Plan{}, nil)
				if err != nil || len(preview.Steps) != 0 {
					t.Fatalf("steps = %+v err=%v", preview.Steps, err)
				}
			})
		}
	}
}

func TestWaitFloorNotOfferedWhereNoProgramRuns(t *testing.T) {
	for name, mode := range map[string][3]bool{
		"non-interactive": {false, false, false},
		"dry run":         {false, true, true},
		"synthetic":       {true, false, true},
	} {
		t.Run(name, func(t *testing.T) {
			a, f, _ := floorFixture(t, "", "codex-cli 0.164.0", mode[0], mode[1], mode[2])
			if _, ok := floorOffer(t, a, codexUser); ok {
				t.Fatal("offered without being able to verify the version")
			}
			if f.versionProbes != 0 {
				t.Fatalf("version probed %d times", f.versionProbes)
			}
		})
	}
}

func TestWaitFloorNotOfferedOutsideUserScopeOrWithoutCodexHost(t *testing.T) {
	a, _, _ := floorFixture(t, "", "codex-cli 0.164.0", false, false, true)
	if _, ok := floorOffer(t, a, management.Options{Scope: "project", Hosts: []string{"codex"}}); ok {
		t.Fatal("offered in project scope")
	}
	if _, ok := floorOffer(t, a, management.Options{Scope: "user", Hosts: []string{"claude", "pi"}}); ok {
		t.Fatal("offered without codex among the hosts")
	}
}

func TestWaitFloorRefusalBecomesAManualStepNamingTheCause(t *testing.T) {
	a, _, _ := floorFixture(t, floorConflict, "codex-cli 0.164.0", false, false, true)
	offer, ok := floorOffer(t, a, codexUser)
	if !ok || !offer.ManualOnly || offer.DefaultOn {
		t.Fatalf("offer = %+v ok=%v, want manual without default", offer, ok)
	}
	preview, err := a.Plan(management.Plan{}, []providerRequest{{ID: string(providers.CodexWaitFloor)}})
	if err != nil {
		t.Fatal(err)
	}
	if got := stepIDs(preview); len(got) != 1 || got[0] != providers.CodexWaitFloorManualStep {
		t.Fatalf("steps = %v", got)
	}
	if err := providers.ValidateStep(preview.Steps[0].Payload); err != nil {
		t.Fatal(err)
	}
	var step providers.Step
	if err := json.Unmarshal(preview.Steps[0].Payload, &step); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(step.ManualReason, "min_wait_timeout_ms has a different value") {
		t.Fatalf("reason = %q", step.ManualReason)
	}
	text := string(preview.Steps[0].Payload) + strings.Join(preview.Details[0].Effects, "\n")
	if strings.Contains(text, "hunter2") {
		t.Fatalf("file text reached the plan: %s", text)
	}
}

func TestWaitFloorSymlinkedCodexHomeIsAManualStep(t *testing.T) {
	a, _, home := floorFixture(t, "", "codex-cli 0.164.0", false, false, true)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(home, link); err != nil {
		t.Fatal(err)
	}
	a.codexHome = link
	offer, ok := floorOffer(t, a, codexUser)
	if !ok || !offer.ManualOnly {
		t.Fatalf("offer = %+v ok=%v", offer, ok)
	}
	preview, err := a.Plan(management.Plan{}, []providerRequest{{ID: string(providers.CodexWaitFloor)}})
	if err != nil || len(preview.Steps) != 1 {
		t.Fatalf("preview = %+v err=%v", preview, err)
	}
	var step providers.Step
	_ = json.Unmarshal(preview.Steps[0].Payload, &step)
	if !strings.Contains(step.ManualReason, "symbolic link") {
		t.Fatalf("reason = %q", step.ManualReason)
	}
}

func TestWaitFloorPlanEmitsNoStepAndSaysWhatItWritesAndHowToRevert(t *testing.T) {
	a, _, _ := floorFixture(t, "", "codex-cli 0.164.0", false, false, true)
	floorOffer(t, a, codexUser)
	preview, err := a.Plan(management.Plan{}, []providerRequest{{ID: string(providers.CodexWaitFloor)}})
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Steps) != 0 {
		t.Fatalf("steps = %v, want none", stepIDs(preview))
	}
	if len(preview.Details) != 1 || preview.Details[0].ID != string(providers.CodexWaitFloor) {
		t.Fatalf("details = %+v", preview.Details)
	}
	effects := strings.Join(preview.Details[0].Effects, "\n")
	for _, want := range []string{"[features.multi_agent_v2]", "min_wait_timeout_ms = 300000", "default_wait_timeout_ms = 600000", "config.toml.hive-backup", "hive plan remove"} {
		if !strings.Contains(effects, want) {
			t.Fatalf("effects lack %q: %s", want, effects)
		}
	}
}

func TestParseCodexVersionAndMinimum(t *testing.T) {
	for in, want := range map[string]bool{
		"codex-cli 0.164.0":         true,
		"codex-cli 0.164.0\n":       true,
		"0.164.1":                   true,
		"codex-cli 0.165.0-alpha.2": true,
		"codex-cli 1.0.0":           true,
		"codex-cli 0.163.9":         false,
		"0.99.0":                    false,
		"codex":                     false,
		"":                          false,
	} {
		if got := codexSupportsWaitFloor(in); got != want {
			t.Errorf("codexSupportsWaitFloor(%q) = %v, want %v", in, got, want)
		}
	}
}

// floorAdapter offers the wait floor and, like the native adapter, plans no
// step for it. It records the plan it was given.
type floorAdapter struct {
	offer    providerOffer
	requests []providerRequest
	plan     management.Plan
}

func newFloorAdapter(manual bool) *floorAdapter {
	return &floorAdapter{offer: providerOffer{ID: string(providers.CodexWaitFloor), Name: string(providers.CodexWaitFloor), Source: "github.com/openai/codex", DefaultOn: !manual, ManualOnly: manual, Effect: "keeps Codex's parent agent waiting longer"}}
}

func (a *floorAdapter) Detect(management.Options) ([]providerOffer, error) {
	return []providerOffer{a.offer}, nil
}
func (a *floorAdapter) Plan(p management.Plan, requests []providerRequest) (onboardingPreview, error) {
	a.plan, a.requests = p, append([]providerRequest(nil), requests...)
	if len(requests) == 0 {
		return onboardingPreview{}, nil
	}
	return onboardingPreview{Details: []providerDetail{{ID: requests[0].ID, Source: "github.com/openai/codex", Effects: []string{"writes the wait floor"}}}}, nil
}
func (a *floorAdapter) Runner() management.ExternalRunner { return testExternalRunner{} }

func installWithFloor(t *testing.T, a *floorAdapter, home, input string) string {
	t.Helper()
	dependencies := defaultInstallDependencies(coreOnlyAdapterFactory)
	dependencies.AdapterFactory = func(onboardingInput) (onboardingAdapter, error) { return a, nil }
	var out strings.Builder
	if err := installWithDependencies(installArgs(t, home), strings.NewReader(input), &out, true, dependencies); err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	return out.String()
}

func TestInstallRebuildsThePlanWithTheWaitFloorOnceChosen(t *testing.T) {
	home := t.TempDir()
	config := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(config), 0o700); err != nil {
		t.Fatal(err)
	}
	before := "model = \"x\"\n"
	if err := os.WriteFile(config, []byte(before), 0o600); err != nil {
		t.Fatal(err)
	}
	a := newFloorAdapter(false)
	out := installWithFloor(t, a, home, "\ny\n")
	if a.plan.CodexWaitFloor == nil {
		t.Fatalf("the plan given to the adapter has no wait floor step:\n%s", out)
	}
	got, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(got), before) || !strings.Contains(string(got), "[features.multi_agent_v2]\nmin_wait_timeout_ms = 300000\ndefault_wait_timeout_ms = 600000\n") {
		t.Fatalf("config.toml = %q", got)
	}
	backup, err := os.ReadFile(config + ".hive-backup")
	if err != nil || string(backup) != before {
		t.Fatalf("backup = %q err=%v", backup, err)
	}
}

func TestInstallWritesNothingToCodexConfigWhenTheWaitFloorIsSkippedOrManual(t *testing.T) {
	for name, tc := range map[string]struct {
		manual bool
		input  string
	}{
		"declined": {false, "none\ny\n"},
		"manual":   {true, "1\ny\n"},
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			a := newFloorAdapter(tc.manual)
			installWithFloor(t, a, home, tc.input)
			if a.plan.CodexWaitFloor != nil {
				t.Fatal("the plan carries a wait floor step")
			}
			mustNotExist(t, filepath.Join(home, ".codex", "config.toml"))
		})
	}
}

func TestInstallWaitFloorMakesTheSummaryAskForConfirmationEvenWhenHiveIsUpToDate(t *testing.T) {
	home := t.TempDir()
	installWithFloor(t, newFloorAdapter(false), home, "none\ny\n") // installs Hive's content only
	a := newFloorAdapter(false)
	out := installWithFloor(t, a, home, "\ny\n")
	if a.plan.CodexWaitFloor == nil {
		t.Fatal("no wait floor step")
	}
	if strings.Contains(out, "Hive was already installed and verified (no changes)") {
		t.Fatalf("the install reported no changes although it writes config.toml:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "config.toml")); err != nil {
		t.Fatal(err)
	}
}
