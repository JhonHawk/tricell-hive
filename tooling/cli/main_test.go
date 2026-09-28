package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"tricell-hive/integrations/target"
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

// captureStdout redirects os.Stdout through an os.Pipe for the duration of fn
// and returns everything written to it. run writes directly to os.Stdout
// (it takes no io.Writer for plan/apply/status/help), so this is the only
// way to observe those subcommands' visible output from a test.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	captured := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		captured <- buf.String()
	}()
	fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = original
	return <-captured
}

// putCharacterization writes a fixture file, creating parent directories as
// needed, mirroring the private helper tooling/management's own tests use
// (unexported, so not reusable across packages).
func putCharacterization(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0640); err != nil {
		t.Fatal(err)
	}
}

// characterizationFixture builds a minimal synthetic home, state directory
// and source checkout for exercising plan/apply/status through run, without
// touching the real user's home or state.
type characterizationFixture struct {
	home, stateDir, source string
}

func newCharacterizationFixture(t *testing.T) characterizationFixture {
	t.Helper()
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := characterizationFixture{
		home:     filepath.Join(base, "home"),
		stateDir: filepath.Join(base, "state"),
		source:   filepath.Join(base, "source"),
	}
	if err := os.MkdirAll(f.home, 0700); err != nil {
		t.Fatal(err)
	}
	putCharacterization(t, filepath.Join(f.source, management.GlobalSource), "# Rules\nKeep user content.\n")
	putCharacterization(t, filepath.Join(f.source, management.SkillSource), "---\nname: workspace-conventions\ndescription: Organize records.\n---\nPreserve evidence.\n")
	return f
}

// expectedPlanInstallBody is the visible plan output shared by the preview
// and the --out cases: the first-setup hint and both managed changes, with
// no trailing "Plan <id>" line (that ID is content- and path-derived and is
// asserted separately, by shape, in each test).
func expectedPlanInstallBody(home string) string {
	return "First setup: Context7 is recommended but optional. Run hive setup for local discovery and official install/update guidance; this plan can proceed without it.\n" +
		fmt.Sprintf("install %s/.codex/AGENTS.md [block]\n", home) +
		"+++ managed after\n<!-- === TRICELL HIVE RULES:BEGIN === -->\n# Rules\nKeep user content.\n<!-- === TRICELL HIVE RULES:END === -->\n\n" +
		fmt.Sprintf("consumer: codex (user, %s)\n", home) +
		fmt.Sprintf("install %s/.agents/skills/workspace-conventions/SKILL.md [skill]\n", home) +
		"+++ managed after\n---\nname: workspace-conventions\ndescription: Organize records.\n---\nPreserve evidence.\n\n" +
		fmt.Sprintf("consumer: codex (user, %s)\n", home)
}

var planIDPreviewSuffix = regexp.MustCompile(`^Plan [0-9a-f]{64}\nPreview only; use --out FILE to save an applicable plan\.\n$`)
var planIDSavedSuffix = regexp.MustCompile(`^Plan [0-9a-f]{64}\n$`)

// TestRunPlanInstallPreviewOnlyPrintsChangesWithoutSaving characterizes
// `hive plan install` invoked without --out: a preview of every managed
// change plus the reminder that nothing was saved, and no plan file left on
// disk.
func TestRunPlanInstallPreviewOnlyPrintsChangesWithoutSaving(t *testing.T) {
	f := newCharacterizationFixture(t)
	out := captureStdout(t, func() {
		if err := run([]string{"plan", "install", "--scope", "user", "--home", f.home, "--source", f.source, "--hosts", "codex", "--state-dir", f.stateDir}); err != nil {
			t.Fatal(err)
		}
	})
	body := expectedPlanInstallBody(f.home)
	if len(out) < len(body) || out[:len(body)] != body {
		t.Fatalf("unexpected plan preview body:\ngot:  %q\nwant prefix: %q", out, body)
	}
	if suffix := out[len(body):]; !planIDPreviewSuffix.MatchString(suffix) {
		t.Fatalf("unexpected plan preview tail: %q", suffix)
	}
}

// TestRunPlanWithOutThenApplySavesAndApplies characterizes `hive plan
// install --out FILE`, which prints the same changes without the preview
// reminder and saves an applicable plan, followed by `hive apply --plan
// FILE`, which reports its result as {"result": "<32-hex-id>"}.
func TestRunPlanWithOutThenApplySavesAndApplies(t *testing.T) {
	f := newCharacterizationFixture(t)
	planFile := filepath.Join(filepath.Dir(f.stateDir), "plan.json")
	planOut := captureStdout(t, func() {
		if err := run([]string{"plan", "install", "--scope", "user", "--home", f.home, "--source", f.source, "--hosts", "codex", "--state-dir", f.stateDir, "--out", planFile}); err != nil {
			t.Fatal(err)
		}
	})
	body := expectedPlanInstallBody(f.home)
	if len(planOut) < len(body) || planOut[:len(body)] != body {
		t.Fatalf("unexpected plan --out body:\ngot:  %q\nwant prefix: %q", planOut, body)
	}
	if suffix := planOut[len(body):]; !planIDSavedSuffix.MatchString(suffix) {
		t.Fatalf("unexpected plan --out tail (should have no preview reminder): %q", suffix)
	}
	if _, err := os.Stat(planFile); err != nil {
		t.Fatalf("expected --out to save a plan file: %v", err)
	}

	applyOut := captureStdout(t, func() {
		if err := run([]string{"apply", "--plan", planFile}); err != nil {
			t.Fatal(err)
		}
	})
	var decoded map[string]string
	if err := json.Unmarshal([]byte(applyOut), &decoded); err != nil {
		t.Fatalf("apply output is not the expected JSON object: %v (%q)", err, applyOut)
	}
	result, ok := decoded["result"]
	if !ok || len(decoded) != 1 {
		t.Fatalf("expected exactly one %q field, got %v", "result", decoded)
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(result) {
		t.Fatalf("expected a 32-character hex result id, got %q", result)
	}
}

// TestRunStatusReportsNotInstalledOnFreshHome characterizes `hive status`
// against a home with nothing installed yet: every managed target is
// reported as not_installed with no release, version or consumers.
func TestRunStatusReportsNotInstalledOnFreshHome(t *testing.T) {
	f := newCharacterizationFixture(t)
	out := captureStdout(t, func() {
		if err := run([]string{"status", "--scope", "user", "--home", f.home, "--hosts", "codex", "--state-dir", f.stateDir}); err != nil {
			t.Fatal(err)
		}
	})
	want := "[\n" +
		"  {\n" +
		fmt.Sprintf("    \"Path\": \"%s/.agents/skills/workspace-conventions/SKILL.md\",\n", f.home) +
		"    \"Host\": \"codex\",\n" +
		"    \"Kind\": \"skill\",\n" +
		"    \"Status\": \"not_installed\",\n" +
		"    \"Release\": \"\",\n" +
		"    \"ProductVersion\": \"\",\n" +
		"    \"VersionStatus\": \"\",\n" +
		"    \"Consumers\": null\n" +
		"  },\n" +
		"  {\n" +
		fmt.Sprintf("    \"Path\": \"%s/.codex/AGENTS.md\",\n", f.home) +
		"    \"Host\": \"codex\",\n" +
		"    \"Kind\": \"block\",\n" +
		"    \"Status\": \"not_installed\",\n" +
		"    \"Release\": \"\",\n" +
		"    \"ProductVersion\": \"\",\n" +
		"    \"VersionStatus\": \"\",\n" +
		"    \"Consumers\": null\n" +
		"  }\n" +
		"]\n"
	if out != want {
		t.Fatalf("unexpected status output:\ngot:  %q\nwant: %q", out, want)
	}
}

// TestRunUsageErrors characterizes the four documented usage-error messages:
// no subcommand at all, `plan` without install/remove, unexpected
// positional arguments after a subcommand's flags, and an unrecognized
// subcommand. HIVE_ACCESSIBLE is cleared so "no subcommand" never depends on
// the ambient environment: with it set, run([]) would open the accessible
// interface instead of returning the usage error (D14-A, tui.go).
func TestRunUsageErrors(t *testing.T) {
	t.Setenv("HIVE_ACCESSIBLE", "")
	cases := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no subcommand",
			args: []string{},
			want: "usage: hive --version | setup | install | plan install|remove | apply --plan FILE | status | recover --state-dir DIR | update | releases | voice list|set|off",
		},
		{
			name: "plan without action",
			args: []string{"plan"},
			want: "plan requires install or remove",
		},
		{
			name: "positional arguments",
			args: []string{"status", "--scope", "user", "extra"},
			want: "unexpected positional arguments",
		},
		{
			name: "unknown command",
			args: []string{"unknown"},
			want: `unknown command "unknown"`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := run(c.args)
			if err == nil {
				t.Fatalf("expected an error for args %v", c.args)
			}
			if err.Error() != c.want {
				t.Fatalf("got error %q, want %q", err.Error(), c.want)
			}
		})
	}
}
