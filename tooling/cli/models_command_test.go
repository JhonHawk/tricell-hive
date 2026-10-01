package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"tricell-hive/integrations/target"
	"tricell-hive/tooling/management"
)

// Tests of `hive models`, `hive models set` and `hive models reset` (#46, T2).
// Every home, state directory and source is synthetic.

var updateModelsGolden = flag.Bool("update-models-golden", false, "rewrite testdata/models_text.golden")

const modelsGoldenPath = "testdata/models_text.golden"

// modelsEnv is a synthetic home with the synthetic agent roles installed on
// five CLIs.
type modelsEnv struct {
	home, stateDir, source string
	hosts                  []string
}

func newModelsEnv(t *testing.T) modelsEnv {
	t.Helper()
	base, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	e := modelsEnv{
		home:     filepath.Join(base, "home"),
		stateDir: filepath.Join(base, "state"),
		source:   modelsTestSource(t),
		hosts:    []string{"claude", "codex", "grok", "opencode", "pi"},
	}
	if err := os.MkdirAll(e.home, 0700); err != nil {
		t.Fatal(err)
	}
	installDirect(t, e.home, e.stateDir, e.source, e.hosts)
	return e
}

func (e modelsEnv) common() []string {
	return []string{"--home", e.home, "--state-dir", e.stateDir}
}

// TestModelsCommandWithoutOverridesPrintsTheBaseText characterizes the text of
// `hive models` for an installation with no override.
func TestModelsCommandWithoutOverridesPrintsTheBaseText(t *testing.T) {
	e := newModelsEnv(t)
	var out bytes.Buffer
	if err := runModels(e.common(), &out); err != nil {
		t.Fatal(err)
	}
	if *updateModelsGolden {
		if err := os.WriteFile(modelsGoldenPath, out.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(modelsGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("hive models output changed without overrides:\n%s\nwant:\n%s", out.String(), want)
	}
}

// --- helpers ----------------------------------------------------------------

func (e modelsEnv) agentFile(host, role string) string {
	dirs := map[string]string{"claude": ".claude/agents", "codex": ".codex/agents", "grok": ".grok/agents", "pi": ".pi/agent/agents", "opencode": ".config/opencode/agents"}
	for _, ext := range []string{".md", ".toml"} {
		p := filepath.Join(e.home, dirs[host], role+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filepath.Join(e.home, dirs[host], role+".md")
}

func (e modelsEnv) read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func (e modelsEnv) state(t *testing.T) string {
	return e.read(t, filepath.Join(e.stateDir, "state.json"))
}

// snapshot hashes every regular file under home and the state directory.
func (e modelsEnv) snapshot(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, root := range []string{e.home, e.stateDir} {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() {
				b, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				out[path] = string(b)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return out
}

func sameFiles(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// write runs `hive models <sub> args…` with the given answers to the prompt.
func (e modelsEnv) write(sub string, interactive bool, answers string, args ...string) (string, error) {
	var out bytes.Buffer
	err := modelsWrite(sub, append(append([]string{}, args...), e.common()...), strings.NewReader(answers), &out, interactive)
	return out.String(), err
}

func (e modelsEnv) modelsText(t *testing.T) string {
	t.Helper()
	var out bytes.Buffer
	if err := runModels(e.common(), &out); err != nil {
		t.Fatal(err)
	}
	return out.String()
}

// --- AC1: set ---------------------------------------------------------------

func TestModelsCommandSetAppliesAndModelsShowsTheMarker(t *testing.T) {
	e := newModelsEnv(t)
	file := e.agentFile("claude", "plain-role")
	before := e.read(t, file)
	cleanText := e.modelsText(t)
	if strings.Contains(cleanText, "*") {
		t.Fatalf("test setup: a clean installation shows a marker:\n%s", cleanText)
	}
	others := e.snapshot(t)
	out, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--model", "opus", "--effort", "max")
	if err != nil {
		t.Fatalf("set: %v\n%s", err, out)
	}
	for _, want := range []string{
		"claude plain-role: model syn-claude-exec → opus, effort medium → max",
		file,
		"Open sessions keep the previous model until they restart.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output has no %q:\n%s", want, out)
		}
	}
	after := e.read(t, file)
	if after == before || !strings.Contains(after, `model: "opus"`) || !strings.Contains(after, `effort: "max"`) {
		t.Fatalf("claude role file after set:\n%s", after)
	}
	for path, text := range e.snapshot(t) {
		if strings.HasPrefix(path, e.stateDir) || path == file {
			continue
		}
		if others[path] != text {
			t.Errorf("set changed %s, which is not the role file or the state", path)
		}
	}
	text := e.modelsText(t)
	if !strings.Contains(text, "plain-role *") || !strings.Contains(text, "opus") || !strings.Contains(text, "* set with hive models set") {
		t.Fatalf("hive models after set:\n%s", text)
	}
	if strings.Count(text, " *\n")+strings.Count(text, " * ") != 1 {
		t.Fatalf("exactly one row should carry the marker:\n%s", text)
	}
}

func TestModelsCommandDryRunWritesNothing(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	for _, interactive := range []bool{true, false} {
		out, err := e.write("set", interactive, "", "--host", "codex", "--role", "plain-role", "--effort", "high", "--dry-run")
		if err != nil {
			t.Fatalf("dry-run: %v\n%s", err, out)
		}
		if !strings.Contains(out, "codex plain-role: model syn-codex-exec → syn-codex-exec, effort medium → high") || !strings.Contains(out, "Preview: nothing was changed.") {
			t.Errorf("dry-run output:\n%s", out)
		}
	}
	if !sameFiles(before, e.snapshot(t)) {
		t.Fatal("--dry-run wrote something")
	}
}

func TestModelsCommandOutThenApplyGivesTheSameFilesAsSet(t *testing.T) {
	saved := newModelsEnv(t)
	direct := newModelsEnv(t)
	planDir, err := target.Canonical(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	planFile := filepath.Join(planDir, "plan.json")
	out, err := saved.write("set", false, "", "--host", "pi", "--role", "plain-role", "--model", "p/m", "--effort", "xhigh", "--out", planFile)
	if err != nil || !strings.Contains(out, "Plan saved to "+planFile) {
		t.Fatalf("--out: %v\n%s", err, out)
	}
	if strings.Contains(saved.read(t, saved.agentFile("pi", "plain-role")), "p/m") {
		t.Fatal("--out applied the plan")
	}
	if err := run([]string{"apply", "--plan", planFile}); err != nil {
		t.Fatal(err)
	}
	if _, err := direct.write("set", true, "y\n", "--host", "pi", "--role", "plain-role", "--model", "p/m", "--effort", "xhigh"); err != nil {
		t.Fatal(err)
	}
	a, b := saved.read(t, saved.agentFile("pi", "plain-role")), direct.read(t, direct.agentFile("pi", "plain-role"))
	if a != b || !strings.Contains(a, `thinking: "xhigh"`) {
		t.Fatalf("saved and applied plan differ from set:\n%s\n---\n%s", a, b)
	}
	var stateA, stateB management.State
	if err := json.Unmarshal([]byte(saved.state(t)), &stateA); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(direct.state(t)), &stateB); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stateA.ModelOverrides, stateB.ModelOverrides) {
		t.Fatalf("overrides differ: %+v vs %+v", stateA.ModelOverrides, stateB.ModelOverrides)
	}
}

func TestModelsCommandSetWithoutTerminalFailsWithoutWriting(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	_, err := e.write("set", false, "", "--host", "claude", "--role", "plain-role", "--effort", "max")
	if err == nil || !strings.Contains(err.Error(), "interactive terminal is required") {
		t.Fatalf("error = %v", err)
	}
	if !sameFiles(before, e.snapshot(t)) {
		t.Fatal("a refused set wrote something")
	}
}

func TestModelsCommandSetDeclinedAndNothingToChange(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	out, err := e.write("set", true, "n\n", "--host", "claude", "--role", "plain-role", "--effort", "max")
	if err != nil || !strings.Contains(out, "Cancelled. No changes applied.") || !sameFiles(before, e.snapshot(t)) {
		t.Fatalf("declined set: %v\n%s", err, out)
	}
	if _, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--effort", "max"); err != nil {
		t.Fatal(err)
	}
	after := e.snapshot(t)
	out, err = e.write("set", false, "", "--host", "claude", "--role", "plain-role", "--effort", "max")
	if err != nil || !strings.Contains(out, "Nothing to change") || !sameFiles(after, e.snapshot(t)) {
		t.Fatalf("repeating a set: %v\n%s", err, out)
	}
}

func TestModelsCommandSetNeverCopiesReleaseValuesIntoTheOverride(t *testing.T) {
	e := newModelsEnv(t)
	if _, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--effort", "max"); err != nil {
		t.Fatal(err)
	}
	stored, err := management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if err != nil {
		t.Fatal(err)
	}
	if got := stored["claude"]["plain-role"]; got != (management.ModelOverride{Effort: "max"}) {
		t.Fatalf("stored = %+v, want an effort-only override", got)
	}
	// A later set of the model keeps the stored effort.
	if _, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--model", "opus"); err != nil {
		t.Fatal(err)
	}
	stored, _ = management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if got := stored["claude"]["plain-role"]; got != (management.ModelOverride{Model: "opus", Effort: "max"}) {
		t.Fatalf("stored = %+v, want the new model and the kept effort", got)
	}
	// On OpenCode an effort over the inherit profile's #max is stored alone.
	if _, err := e.write("set", true, "y\n", "--host", "opencode", "--role", "inherit-role", "--effort", "high"); err != nil {
		t.Fatal(err)
	}
	stored, _ = management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if got := stored["opencode"]["inherit-role"]; got != (management.ModelOverride{Effort: "high"}) {
		t.Fatalf("opencode stored = %+v", got)
	}
	if !strings.Contains(e.read(t, e.agentFile("opencode", "inherit-role")), `model: "syn-oc/inherit#high"`) {
		t.Fatal("OpenCode file does not carry #high")
	}
}

// --- AC3: reset -------------------------------------------------------------

func TestModelsCommandResetAllReturnsToTheCleanInstallation(t *testing.T) {
	e := newModelsEnv(t)
	clean := e.snapshot(t)
	for _, args := range [][]string{
		{"--host", "claude", "--role", "plain-role", "--model", "opus", "--effort", "max"},
		{"--host", "claude", "--role", "inherit-role", "--effort", "low"},
	} {
		if _, err := e.write("set", true, "y\n", args...); err != nil {
			t.Fatal(err)
		}
	}
	out, err := e.write("reset", true, "y\n", "--host", "claude", "--all")
	if err != nil {
		t.Fatalf("reset --all: %v\n%s", err, out)
	}
	// Compare the installed files and state.json; transaction journals and
	// release snapshots in the state directory are history, not installation.
	now := e.snapshot(t)
	for path, text := range clean {
		if strings.HasPrefix(path, e.stateDir) && filepath.Base(path) != "state.json" {
			continue
		}
		if now[path] != text {
			t.Errorf("%s differs from the clean installation", path)
		}
	}
	for path := range now {
		if _, ok := clean[path]; !ok && !strings.HasPrefix(path, e.stateDir) {
			t.Errorf("%s is new", path)
		}
	}
	if strings.Contains(e.modelsText(t), "*") {
		t.Fatal("a marker remains after reset --all")
	}
}

func TestModelsCommandResetOnlyKeepsTheOtherPart(t *testing.T) {
	e := newModelsEnv(t)
	set := []string{"--host", "claude", "--role", "plain-role", "--model", "opus", "--effort", "max"}
	if _, err := e.write("set", true, "y\n", set...); err != nil {
		t.Fatal(err)
	}
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--role", "plain-role", "--only", "effort"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	stored, _ := management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if got := stored["claude"]["plain-role"]; got != (management.ModelOverride{Model: "opus"}) {
		t.Fatalf("after reset --only effort: %+v", got)
	}
	file := e.read(t, e.agentFile("claude", "plain-role"))
	if !strings.Contains(file, `model: "opus"`) || !strings.Contains(file, `effort: "medium"`) {
		t.Fatalf("file after reset --only effort:\n%s", file)
	}
	// --only model on a model-only override is the same as reset --role.
	onlyModel := e.snapshot(t)
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--role", "plain-role", "--only", "model"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	stored, _ = management.StoredModelOverrides(management.Options{Home: e.home, StateDir: e.stateDir})
	if len(stored) != 0 {
		t.Fatalf("an override with no part left remains: %+v", stored)
	}
	if strings.Contains(e.state(t), "model_overrides") || sameFiles(onlyModel, e.snapshot(t)) {
		t.Fatal("reset --only model on a model-only override did not clear it")
	}
}

func TestModelsCommandResetOfAnUnsetRoleChangesNothing(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	out, err := e.write("reset", false, "", "--host", "claude", "--role", "plain-role")
	if err != nil || !strings.Contains(out, "Nothing to change") || !sameFiles(before, e.snapshot(t)) {
		t.Fatalf("%v\n%s", err, out)
	}
}

// --- AC4: rejections -----------------------------------------------------------

func TestModelsCommandRejectionsNameTheProblemAndWriteNothing(t *testing.T) {
	e := newModelsEnv(t)
	before := e.snapshot(t)
	cases := []struct {
		name string
		sub  string
		args []string
		want string
	}{
		{"effort on grok", "set", []string{"--host", "grok", "--role", "plain-role", "--effort", "high"}, "grok does not support effort"},
		{"unknown effort", "set", []string{"--host", "claude", "--role", "plain-role", "--effort", "extreme"}, "effort must be one of"},
		{"model with a space", "set", []string{"--host", "claude", "--role", "plain-role", "--model", "a b"}, "model may only contain"},
		{"model with a hash", "set", []string{"--host", "opencode", "--role", "plain-role", "--model", "a/b#high"}, "model may only contain"},
		{"empty model", "set", []string{"--host", "claude", "--role", "plain-role", "--model", ""}, "--model must not be empty"},
		{"empty effort", "set", []string{"--host", "claude", "--role", "plain-role", "--effort", ""}, "--effort must not be empty"},
		{"neither model nor effort", "set", []string{"--host", "claude", "--role", "plain-role"}, "needs --model or --effort"},
		{"missing host", "set", []string{"--role", "plain-role", "--effort", "high"}, "--host is required"},
		{"missing role", "set", []string{"--host", "claude", "--effort", "high"}, "--role is required"},
		{"unknown role", "set", []string{"--host", "claude", "--role", "ghost-role", "--effort", "high"}, `role "ghost-role" is not in the release installed for claude`},
		{"unregistered CLI", "set", []string{"--host", "cursor", "--role", "plain-role", "--model", "x"}, "cursor is not installed"},
		{"unknown CLI", "set", []string{"--host", "vim", "--role", "plain-role", "--model", "x"}, "unsupported host"},
		{"reset without a target", "reset", []string{"--host", "claude"}, "--role or --all"},
		{"reset all with a role", "reset", []string{"--host", "claude", "--all", "--role", "plain-role"}, "--all cannot be combined"},
		{"reset only without a role", "reset", []string{"--host", "claude", "--all", "--only", "model"}, "--only needs --role"},
		{"reset with a bad part", "reset", []string{"--host", "claude", "--role", "plain-role", "--only", "both"}, "--only must be model or effort"},
		{"reset of an unknown role", "reset", []string{"--host", "claude", "--role", "ghost-role"}, `role "ghost-role" is unknown for claude`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out, err := e.write(c.sub, true, "y\n", c.args...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error = %v (output %q), want %q", err, out, c.want)
			}
			if !sameFiles(before, e.snapshot(t)) {
				t.Fatal("a rejected command wrote something")
			}
		})
	}
}

func TestModelsCommandRefusesAnInstallationItWouldRepair(t *testing.T) {
	e := newModelsEnv(t)
	// Deleting Claude's instruction file by hand: setting a model must not restore it.
	if err := os.Remove(filepath.Join(e.home, ".claude", "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	before := e.snapshot(t)
	_, err := e.write("set", true, "y\n", "--host", "claude", "--role", "plain-role", "--effort", "max")
	if err == nil || !strings.Contains(err.Error(), "hive install") {
		t.Fatalf("error = %v, want a pointer to hive install or hive doctor", err)
	}
	if !sameFiles(before, e.snapshot(t)) {
		t.Fatal("a refused set wrote something")
	}
}

// --- hive models output ---------------------------------------------------------

func TestModelsCommandListsAnOverrideForARoleTheReleaseLostAsNotApplied(t *testing.T) {
	e := newModelsEnv(t)
	var st management.State
	if err := json.Unmarshal([]byte(e.state(t)), &st); err != nil {
		t.Fatal(err)
	}
	st.ModelOverrides = map[string]map[string]management.ModelOverride{"claude": {"retired-role": {Model: "x"}}}
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.stateDir, "state.json"), append(b, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	text := e.modelsText(t)
	if !strings.Contains(text, "retired-role: not applied: role not in the installed release") {
		t.Fatalf("hive models:\n%s", text)
	}
	if strings.Contains(text, "* set with hive models set") {
		t.Fatalf("the note appears without a marked row:\n%s", text)
	}
	// Resetting it removes the line.
	if out, err := e.write("reset", true, "y\n", "--host", "claude", "--role", "retired-role"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if strings.Contains(e.modelsText(t), "retired-role") {
		t.Fatal("the unapplied override is still listed after reset")
	}
}
