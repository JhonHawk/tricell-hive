package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Fake Codex that completes: the runner sees turn.completed.
const fakeCodexCompletes = `#!/bin/sh
echo "$(basename "$0") $*" >> "$FAKE_HOST_LOG"
case "$1" in
  --version) echo "fake 1.0"; exit 0 ;;
esac
cat > /dev/null
echo '{"type":"turn.completed","usage":{}}'
exit 0
`

// qrEnv is one `qualify run` environment: a temporary checkout with the real
// suite, the real model-floor.md and fake hosts on PATH, and a temporary HOME.
type qrEnv struct {
	*qualifyRun
	recorded [][]string // runner arguments seen by a recording hook
}

func newQREnv(t *testing.T) *qrEnv {
	t.Helper()
	q := newQualifyRun(t)
	data, err := os.ReadFile(filepath.Join(repoRoot, "_support/docs/harness-engineering/model-floor.md"))
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFiles(t, q.source, map[string]string{"_support/docs/harness-engineering/model-floor.md": string(data)})
	writeFixtureFiles(t, q.bin, map[string]string{"codex": fakeCodexCompletes})
	if err := os.Chmod(filepath.Join(q.bin, "codex"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", q.home)
	t.Setenv("PATH", q.bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FAKE_HOST_LOG", q.log)
	t.Setenv("ENGRAM_DATA_DIR", filepath.Join(q.home, "everyday"))
	for _, k := range []string{"CLAUDE_CONFIG_DIR", "XDG_CONFIG_HOME", "XDG_CACHE_HOME", "XDG_DATA_HOME"} {
		t.Setenv(k, "")
	}
	return &qrEnv{qualifyRun: q}
}

// useRunner re-executes the test binary as the pilot runner.
func (e *qrEnv) useRunner() {
	e.t.Helper()
	old := qualifyRunnerCommand
	e.t.Cleanup(func() { qualifyRunnerCommand = old })
	qualifyRunnerCommand = func(args []string) *exec.Cmd {
		e.recorded = append(e.recorded, args)
		cmd := exec.Command(os.Args[0], "-test.run=^TestQualificationRunnerSubprocess$")
		cmd.Env = append(os.Environ(), "QUALIFY_RUNNER_SUBPROCESS=1", "QUALIFY_RUNNER_ARGS="+strings.Join(args, "\n"))
		return cmd
	}
}

// useRecorder replaces the runner with a recorder that launches nothing; with
// writeRun it leaves a run.json that stopped before the host launched (exit 1),
// as the runner does at main.go:679.
func (e *qrEnv) useRecorder(writeRun bool) {
	e.t.Helper()
	old := qualifyRunnerCommand
	e.t.Cleanup(func() { qualifyRunnerCommand = old })
	qualifyRunnerCommand = func(args []string) *exec.Cmd {
		e.recorded = append(e.recorded, args)
		if writeRun {
			out := qrArg(args, "--out")
			_ = os.MkdirAll(out, 0o700)
			_ = os.WriteFile(filepath.Join(out, "run.json"), []byte(`{"Terminal":"not_verified"}`), 0o600)
			return exec.Command("false")
		}
		return exec.Command("true")
	}
}

func qrArg(args []string, name string) string {
	for i, a := range args {
		if a == name && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func qrHas(args []string, name string) bool {
	for _, a := range args {
		if a == name {
			return true
		}
	}
	return false
}

// suite builds the suite of the named real tasks.
func (e *qrEnv) suite(t *testing.T, ids ...string) qualifySuite {
	t.Helper()
	full, err := loadQualificationSuite(e.source)
	if err != nil {
		t.Fatal(err)
	}
	var s qualifySuite
	s.Version, s.Groups = full.Version, full.Groups
	for _, id := range ids {
		found := false
		for _, task := range full.Tasks {
			if task.ID == id {
				s.Tasks = append(s.Tasks, task)
				found = true
			}
		}
		if !found {
			t.Fatalf("task %s not in the real suite", id)
		}
	}
	return s
}

func (e *qrEnv) loader(s qualifySuite) qualifySuiteLoader {
	return func(string) (qualifySuite, error) { return s, nil }
}

func (e *qrEnv) outDir(name string) string {
	return filepath.Join(e.source, "_support/workspace", name)
}

// call runs `qualify run` with the common flags plus args.
func (e *qrEnv) call(s qualifySuite, out string, args ...string) (int, string, string) {
	var so, se bytes.Buffer
	full := append([]string{"--source", e.source, "--guidance-source", e.source, "--out", out, "--configured-model", "configured"}, args...)
	code := qualifyRunCmd(full, e.loader(s), &so, &se)
	return code, so.String(), se.String()
}

func (e *qrEnv) codexArgs(runs int, extra ...string) []string {
	return append([]string{"--host", "codex", "--models", "gpt-6.1-terra", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=high", "--runs", strconv.Itoa(runs), "--split", "heldback"}, extra...)
}

func (e *qrEnv) codexLaunches() int {
	return strings.Count(e.launches(), "codex -a never exec")
}

const (
	qrCompare = "question-no-edit"
	qrPlain   = "project-state-routes-research"
)

func TestQualifyRunRequiresConfirmationEqualToTheSessionsItLaunches(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRunner()
	s := e.suite(t, qrCompare, qrPlain)
	out := e.outDir("confirm")

	// Calendar: reference 2 tasks x 3 + candidate deployed 2 x 3 + A 1 x 3 + none 1 x 3 = 18.
	for _, extra := range [][]string{nil, {"--confirm-sessions", "17"}, {"--confirm-sessions", "19"}} {
		code, so, se := e.call(s, out, append(e.codexArgs(3), extra...)...)
		if code != 2 {
			t.Fatalf("%v: exit %d, want 2\n%s%s", extra, code, so, se)
		}
		if !strings.Contains(so+se, "18") {
			t.Fatalf("the refusal does not state the session count: %s%s", so, se)
		}
		if e.codexLaunches() != 0 || len(e.recorded) != 0 {
			t.Fatalf("a refused call launched: %d recorded", len(e.recorded))
		}
		if _, err := os.Stat(filepath.Join(out, "batch.json")); err == nil {
			t.Fatal("a refused call wrote batch.json")
		}
	}
	code, so, se := e.call(s, out, e.codexArgs(3, "--confirm-sessions", "18")...)
	if code != 0 {
		t.Fatalf("exit %d\n%s%s", code, so, se)
	}
	if len(e.recorded) != 18 || e.codexLaunches() != 18 {
		t.Fatalf("launched %d runners and %d hosts, want 18", len(e.recorded), e.codexLaunches())
	}
	// Each repetition rotates the order of the arms.
	order := map[int][]string{}
	for _, a := range e.recorded {
		if qrArg(a, "--task") == qrCompare && qrArg(a, "--model") == "gpt-6.1-terra" {
			rep := filepath.Base(filepath.Dir(qrArg(a, "--out")))
			n, _ := strconv.Atoi(rep)
			order[n] = append(order[n], qrArg(a, "--arm"))
		}
	}
	if len(order) != 3 || strings.Join(order[1], ",") == strings.Join(order[2], ",") {
		t.Fatalf("arm order does not rotate by repetition: %v", order)
	}
}

func TestQualifyRunBatchWrittenAndCheckedBeforeLaunching(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRecorder(false)
	s := e.suite(t, qrCompare)
	out := e.outDir("batch")
	// reference 1 + candidate deployed 1 + A 1 + none 1 = 4 with one repetition.
	if code, so, se := e.call(s, out, e.codexArgs(1, "--confirm-sessions", "4")...); code != 0 {
		t.Fatalf("first call: %d\n%s%s", code, so, se)
	}
	b, err := readQualifyBatch(out)
	if err != nil {
		t.Fatal(err)
	}
	if b.Host != "codex" || b.Reference != "gpt-6.1-sol" || len(b.Models) != 1 || b.Models[0] != "gpt-6.1-terra" ||
		b.Effort["gpt-6.1-terra"] != "medium" || b.Effort["gpt-6.1-sol"] != "high" ||
		b.DeployedHash != mustHash(t, "codex", e.home) || len(b.AHash) != 64 {
		t.Fatalf("batch.json: %+v", b)
	}
	launched := len(e.recorded)

	refuse := func(name string, want []string, args ...string) {
		t.Helper()
		code, so, se := e.call(s, out, append(args, "--confirm-sessions", "4")...)
		if code != 2 {
			t.Fatalf("%s: exit %d, want 2\n%s%s", name, code, so, se)
		}
		for _, w := range want {
			if !strings.Contains(so+se, w) {
				t.Fatalf("%s: message lacks %q\n%s%s", name, w, so, se)
			}
		}
		if len(e.recorded) != launched {
			t.Fatalf("%s: a refused call launched", name)
		}
	}
	// A changed --reference, --effort or --models.
	refuse("reference", []string{"gpt-6.1-sol", "gpt-6.1-terra"}, "--host", "codex", "--models", "gpt-6.1-sol", "--reference", "gpt-6.1-terra", "--effort", "gpt-6.1-sol=high", "--runs", "1", "--split", "heldback")
	refuse("effort", []string{"high", "medium"}, "--host", "codex", "--models", "gpt-6.1-terra", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=medium", "--runs", "1", "--split", "heldback")
	refuse("models", []string{"gpt-6.1-terra"}, "--host", "codex", "--models", "gpt-6.1-terra,gpt-6.1-sol", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=high", "--runs", "1", "--split", "heldback")

	// A managed file of the deployed home changes after the first call.
	block, _, err := guidanceReadPaths("codex", e.home)
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(block)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(block, []byte(strings.Replace(string(original), "Body.", "Edited.", 1)), 0600); err != nil {
		t.Fatal(err)
	}
	edited := mustHash(t, "codex", e.home)
	refuse("deployed hash", []string{b.DeployedHash, edited}, e.codexArgs(1)...)
	if err := os.WriteFile(block, original, 0600); err != nil {
		t.Fatal(err)
	}
	// The unchanged call is accepted again (nothing was usable, so all 4 relaunch).
	if code, so, se := e.call(s, out, e.codexArgs(1, "--confirm-sessions", "4")...); code != 0 {
		t.Fatalf("unchanged call: %d\n%s%s", code, so, se)
	}
}

// seedAttempt writes a synthetic run.json for one attempt of one cell.
func (e *qrEnv) seedAttempt(t *testing.T, out string, b qualifyBatch, model, arm, task string, rep, attempt int, mutate func(*result)) {
	t.Helper()
	r := result{Host: b.Host, Suite: qualificationSuite, Task: task, Split: "heldback", Arm: arm, ModelRequested: model, Terminal: "completed",
		Effort: b.Effort[qualifyBaseModel(model)], ModelObserved: "not_observed"}
	switch arm {
	case qualifyArmDeployed:
		r.GuidanceHash = b.DeployedHash
	case "A":
		r.GuidanceHash = b.AHash
	default:
		r.GuidanceHash = qualifyHashNone
	}
	if mutate != nil {
		mutate(&r)
	}
	dir := qualifyAttemptDir(out, b.Host, model, arm, task, strconv.Itoa(rep), attempt)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(r)
	if err := os.WriteFile(filepath.Join(dir, "run.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestQualifyRunRelaunchesOnlyRepetitionsWithoutAUsableAttemptAndScores(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRunner()
	s := e.suite(t, qrCompare, qrPlain)
	out := e.outDir("resume")

	// The batch of the first call, as the driver would have written it.
	aHash, err := qualifyExpectedAHash("codex", e.source)
	if err != nil {
		t.Fatal(err)
	}
	b := qualifyBatch{Host: "codex", Models: []string{"gpt-6.1-terra"}, Reference: "gpt-6.1-sol",
		Effort: map[string]string{"gpt-6.1-terra": "medium", "gpt-6.1-sol": "high"}, DeployedHash: mustHash(t, "codex", e.home), AHash: aHash}
	if err := os.MkdirAll(out, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := writeQualifyBatch(out, b); err != nil {
		t.Fatal(err)
	}
	// Five usable repetitions...
	e.seedAttempt(t, out, b, "gpt-6.1-sol", qualifyArmDeployed, qrCompare, 1, 1, nil)
	e.seedAttempt(t, out, b, "gpt-6.1-sol", qualifyArmDeployed, qrPlain, 1, 1, nil)
	e.seedAttempt(t, out, b, "gpt-6.1-terra", qualifyArmDeployed, qrCompare, 1, 1, nil)
	e.seedAttempt(t, out, b, "gpt-6.1-terra", qualifyArmDeployed, qrPlain, 1, 1, nil)
	e.seedAttempt(t, out, b, "gpt-6.1-terra", "A", qrCompare, 1, 1, nil)
	// ...and a completed attempt that does not count (another effort): it is
	// counted among the sessions to launch and relaunched in attempt-2.
	e.seedAttempt(t, out, b, "gpt-6.1-terra", qualifyArmNone, qrCompare, 1, 1, func(r *result) { r.Effort = "high" })

	if code, so, se := e.call(s, out, e.codexArgs(3, "--confirm-sessions", "18")...); code != 2 {
		t.Fatalf("the full count is stale after 5 usable: exit %d\n%s%s", code, so, se)
	}
	if len(e.recorded) != 0 {
		t.Fatal("a refused call launched")
	}
	code, so, se := e.call(s, out, e.codexArgs(3, "--confirm-sessions", "13")...)
	if code != 0 {
		t.Fatalf("exit %d\n%s%s", code, so, se)
	}
	if len(e.recorded) != 13 || e.codexLaunches() != 13 {
		t.Fatalf("launched %d runners, %d hosts, want 13", len(e.recorded), e.codexLaunches())
	}
	for _, a := range e.recorded {
		o := qrArg(a, "--out")
		if strings.HasSuffix(filepath.Dir(o), "/1") && strings.Contains(o, "/gpt-6.1-sol/deployed/"+qrCompare+"/1/") {
			t.Fatalf("a repetition with a usable attempt was relaunched: %s", o)
		}
	}
	rerun := qualifyAttemptDir(out, "codex", "gpt-6.1-terra", qualifyArmNone, qrCompare, "1", 2)
	if _, err := os.Stat(filepath.Join(rerun, "run.json")); err != nil {
		t.Fatalf("the unusable repetition was not relaunched in attempt-2: %v", err)
	}
	if _, err := os.Stat(filepath.Join(qualifyAttemptDir(out, "codex", "gpt-6.1-terra", qualifyArmNone, qrCompare, "1", 1), "run.json")); err != nil {
		t.Fatal("the previous attempt was not kept")
	}
	// Now every repetition has a usable attempt: nothing to launch.
	if code, so, se := e.call(s, out, e.codexArgs(3, "--confirm-sessions", "0")...); code != 0 || e.codexLaunches() != 13 {
		t.Fatalf("a complete batch launched: exit %d\n%s%s", code, so, se)
	}

	// A run of the runner with the real checkout installed in the isolated home
	// is usable: its A hash is the one of batch.json.
	var ran result
	data, err := os.ReadFile(filepath.Join(qualifyAttemptDir(out, "codex", "gpt-6.1-terra", "A", qrCompare, "2", 1), "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &ran); err != nil {
		t.Fatal(err)
	}
	floor, _ := parseQualifyFloor(mustRead(t, filepath.Join(e.source, "_support/docs/harness-engineering/model-floor.md")))
	if ok, cause := usableAttempt(ran, b, floor); !ok || ran.GuidanceHash != b.AHash {
		t.Fatalf("an A attempt of the runner is not usable: %v %q hash %s want %s", ok, cause, ran.GuidanceHash, b.AHash)
	}

	// qualify score over that output prints the report with a verdict.
	var so2, se2 bytes.Buffer
	if code := qualifyScoreCmd([]string{"--source", e.source, "--out", out, "--reference", "gpt-6.1-sol"}, e.loader(s), &so2, &se2); code != 0 {
		t.Fatalf("score exit %d\n%s%s", code, so2.String(), se2.String())
	}
	if !strings.Contains(so2.String(), "verdict:") || !strings.Contains(so2.String(), "gpt-6.1-terra") {
		t.Fatalf("report: %s", so2.String())
	}
	// A score without the batch's reference is refused with code 1 and no verdict.
	so2.Reset()
	if code := qualifyScoreCmd([]string{"--source", e.source, "--out", out, "--reference", "gpt-6.1-terra"}, e.loader(s), &so2, &se2); code != 1 || strings.Contains(so2.String(), "verdict:") {
		t.Fatalf("score with another reference: exit %d\n%s", code, so2.String())
	}
}

func mustRead(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestQualifyRunPassesFlagsPerHostAndArm(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRecorder(false)
	s := e.suite(t, qrCompare)
	if code, so, se := e.call(s, e.outDir("flags"), e.codexArgs(1, "--confirm-sessions", "4")...); code != 0 {
		t.Fatalf("%d\n%s%s", code, so, se)
	}
	for _, a := range e.recorded {
		arm := qrArg(a, "--arm")
		if (arm == "A") != qrHas(a, "--guidance-source") {
			t.Errorf("--guidance-source must be passed only to A: %v", a)
		}
		if qrArg(a, "--source") != e.source || qrArg(a, "--suite") != qualificationSuite || qrArg(a, "--delivery") != "deployed-global" || qrArg(a, "--configured-model") != "configured" {
			t.Errorf("runner flags: %v", a)
		}
		if qrHas(a, "--provider") {
			t.Errorf("--provider only on Pi: %v", a)
		}
		want := "medium"
		if qrArg(a, "--model") == "gpt-6.1-sol" {
			want = "high"
		}
		if qrArg(a, "--effort") != want {
			t.Errorf("effort %q, want %q: %v", qrArg(a, "--effort"), want, a)
		}
	}
	// Pi gets --provider and, having no table row, needs --effort.
	e2 := newQREnv(t)
	installFor(t, e2.source, e2.home, "pi")
	e2.useRecorder(false)
	s2 := e2.suite(t, qrPlain)
	args := []string{"--host", "pi", "--models", "m1", "--reference", "m2", "--provider", "prov", "--runs", "1", "--split", "heldback", "--confirm-sessions", "2"}
	if code, _, se := e2.call(s2, e2.outDir("pi"), args...); code != 2 || !strings.Contains(se, "no row for host pi") {
		t.Fatalf("a Pi model outside the table without --effort must exit 2 naming the model: %d %s", code, se)
	}
	args = append(args, "--effort", "m1=low", "--effort", "m2=high")
	if code, so, se := e2.call(s2, e2.outDir("pi"), args...); code != 0 {
		t.Fatalf("%d\n%s%s", code, so, se)
	}
	if len(e2.recorded) != 2 {
		t.Fatalf("pi launched %d", len(e2.recorded))
	}
	for _, a := range e2.recorded {
		if qrArg(a, "--provider") != "prov" {
			t.Errorf("pi without --provider: %v", a)
		}
	}
}

func TestQualifyRunResolvesEffortFromTheFloorTable(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		models  string
		ref     string
		efforts []string
		code    int
		model   string // runner --model expected for the first model
		effort  string // runner --effort expected ("" = absent)
	}{
		{"claude single value", "claude", "sonnet", "opus", nil, 0, "sonnet", "high"},
		{"codex several values need --effort", "codex", "gpt-6.1-sol", "gpt-6.1-terra", nil, 2, "", ""},
		{"codex chosen value", "codex", "gpt-6.1-sol", "gpt-6.1-terra", []string{"gpt-6.1-sol=high"}, 0, "gpt-6.1-sol", "high"},
		{"grok has no effort", "grok", "grok-4.6", "grok-4.6", nil, 0, "grok-4.6", ""},
		{"reference outside the table", "claude", "sonnet", "haiku", nil, 2, "", ""},
		{"reference outside the table with --effort", "claude", "sonnet", "haiku", []string{"haiku=low"}, 0, "sonnet", "high"},
		{"opencode uses the variant", "opencode", "opencode-go/claude-haiku-5-5", "opencode-go/qwen3.8-flash", nil, 0, "opencode-go/claude-haiku-5-5#high", ""},
		{"opencode suffix equal to the value", "opencode", "opencode-go/claude-haiku-5-5#high", "opencode-go/qwen3.8-flash", nil, 0, "opencode-go/claude-haiku-5-5#high", ""},
		{"opencode contradicting suffix", "opencode", "opencode-go/claude-haiku-5-5#medium", "opencode-go/qwen3.8-flash", nil, 2, "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newQREnv(t)
			installFor(t, e.source, e.home, c.host)
			e.useRecorder(false)
			s := e.suite(t, qrPlain)
			args := []string{"--host", c.host, "--models", c.models, "--reference", c.ref, "--runs", "1", "--split", "heldback"}
			for _, ef := range c.efforts {
				args = append(args, "--effort", ef)
			}
			// Same model as reference runs once; two distinct models run twice.
			n := 2
			if c.models == c.ref {
				n = 1
			}
			args = append(args, "--confirm-sessions", strconv.Itoa(n))
			code, so, se := e.call(s, e.outDir("effort"), args...)
			if code != c.code {
				t.Fatalf("exit %d, want %d\n%s%s", code, c.code, so, se)
			}
			if c.code == 2 {
				if len(e.recorded) != 0 {
					t.Fatal("a refused call launched")
				}
				return
			}
			var first []string
			for _, a := range e.recorded {
				if qualifyBaseModel(qrArg(a, "--model")) == qualifyBaseModel(c.model) {
					first = a
				}
			}
			if first == nil {
				t.Fatalf("no launch for %s: %v", c.model, e.recorded)
			}
			if qrArg(first, "--model") != c.model {
				t.Errorf("model %q, want %q", qrArg(first, "--model"), c.model)
			}
			if c.host == "opencode" && qrHas(first, "--effort") {
				t.Errorf("OpenCode must never receive --effort: %v", first)
			}
			if got := qrArg(first, "--effort"); got != c.effort {
				t.Errorf("effort %q, want %q", got, c.effort)
			}
			b, err := readQualifyBatch(e.outDir("effort"))
			if err != nil {
				t.Fatal(err)
			}
			if c.host == "opencode" && b.Effort["opencode-go/claude-haiku-5-5"] != "high" {
				t.Errorf("batch effort for OpenCode is the variant: %v", b.Effort)
			}
		})
	}
}

func TestQualifyRunRefusesOutsideWorkspaceAndEveryPreLaunchProblemWithCode2(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRecorder(false)
	s := e.suite(t, qrCompare)
	for name, out := range map[string]string{
		"outside the checkout": filepath.Join(t.TempDir(), "out"),
		"the checkout root":    filepath.Join(e.source, "out"),
		"the workspace itself": filepath.Join(e.source, "_support/workspace"),
		"a dot-dot escape":     filepath.Join(e.source, "_support/workspace/../elsewhere"),
		"sibling of workspace": filepath.Join(e.source, "_support/workspace-other"),
	} {
		code, so, se := e.call(s, out, e.codexArgs(1, "--confirm-sessions", "4")...)
		if code != 2 || len(e.recorded) != 0 {
			t.Errorf("%s: exit %d launched %d\n%s%s", name, code, len(e.recorded), so, se)
		}
	}
	// Missing required flags and an unknown host also refuse.
	for name, args := range map[string][]string{
		"no host":         {"--models", "gpt-6.1-terra", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=high", "--confirm-sessions", "4"},
		"no models":       {"--host", "codex", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=high", "--confirm-sessions", "4"},
		"no reference":    {"--host", "codex", "--models", "gpt-6.1-terra", "--confirm-sessions", "4"},
		"unknown host":    {"--host", "nope", "--models", "a", "--reference", "b", "--confirm-sessions", "1"},
		"unknown flag":    {"--bogus"},
		"effort for none": {"--host", "codex", "--models", "gpt-6.1-terra", "--reference", "gpt-6.1-sol", "--effort", "other=high", "--confirm-sessions", "4"},
	} {
		code, so, se := e.call(s, e.outDir("refusals"), args...)
		if code != 2 || len(e.recorded) != 0 {
			t.Errorf("%s: exit %d launched %d\n%s%s", name, code, len(e.recorded), so, se)
		}
	}
	// Dispatch through qualifyMain: unknown subcommand is a usage error.
	var so, se bytes.Buffer
	if code := qualifyMain([]string{"bogus"}, &so, &se); code != 2 {
		t.Errorf("unknown subcommand: exit %d", code)
	}
}

func TestQualifyRunKeepsPreviousAttemptsAndEncodesModelPaths(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "opencode")
	e.useRecorder(true) // every runner leaves only a pre-launch run.json and exits 1
	s := e.suite(t, qrPlain)
	out := e.outDir("keep")
	args := []string{"--host", "opencode", "--models", "opencode-go/claude-haiku-5-5", "--reference", "opencode-go/qwen3.8-flash", "--runs", "1", "--split", "heldback", "--confirm-sessions", "2"}
	code, so, se := e.call(s, out, args...)
	if code != 1 {
		t.Fatalf("failed launches must end with exit 1: %d\n%s%s", code, so, se)
	}
	for _, a := range e.recorded {
		o := qrArg(a, "--out")
		if !strings.Contains(o, "opencode-go%2Fclaude-haiku-5-5%23high") && !strings.Contains(o, "opencode-go%2Fqwen3.8-flash%23medium") {
			t.Errorf("model path not encoded: %s", o)
		}
		if filepath.Base(o) != "attempt-1" {
			t.Errorf("first attempt is %s", o)
		}
	}
	// The pre-launch run.json is not usable: the second call counts and relaunches in attempt-2.
	e.recorded = nil
	code, so, se = e.call(s, out, args...)
	if code != 1 || len(e.recorded) != 2 {
		t.Fatalf("second call: %d, %d launches\n%s%s", code, len(e.recorded), so, se)
	}
	for _, a := range e.recorded {
		if filepath.Base(qrArg(a, "--out")) != "attempt-2" {
			t.Errorf("second attempt is %s", qrArg(a, "--out"))
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(qrArg(a, "--out")), "attempt-1", "run.json")); err != nil {
			t.Error("attempt-1 was not kept")
		}
	}
}

func TestQualifyRunRecordsAndChecksTheSplit(t *testing.T) {
	e := newQREnv(t)
	installFor(t, e.source, e.home, "codex")
	e.useRecorder(false)
	s := e.suite(t, qrCompare)
	s.Tasks[0].Split = qualifySplitTuning
	out := e.outDir("split")
	args := func(split string) []string {
		return append([]string{"--host", "codex", "--models", "gpt-6.1-terra", "--reference", "gpt-6.1-sol", "--effort", "gpt-6.1-sol=high", "--runs", "1", "--split", split}, "--confirm-sessions", "4")
	}
	if code, so, se := e.call(s, out, args(qualifySplitTuning)...); code != 0 {
		t.Fatalf("tuning call: %d\n%s%s", code, so, se)
	}
	b, err := readQualifyBatch(out)
	if err != nil {
		t.Fatal(err)
	}
	if b.Split != qualifySplitTuning {
		t.Fatalf("batch.json split = %q", b.Split)
	}
	s.Tasks[0].Split = qualifySplitHeldback
	code, so, se := e.call(s, out, args(qualifySplitHeldback)...)
	if code != 2 || !strings.Contains(so+se, "split: batch.json has") {
		t.Fatalf("a changed split was not refused: %d\n%s%s", code, so, se)
	}
}
