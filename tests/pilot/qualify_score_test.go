package main

import (
	"bytes"
	"encoding/json"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const qsFloorDoc = "# Model support floor\n\n## Floor\n\n| Host | Model |\n| --- | --- |\n| claude | `opus` |\n\n## Aliases\n\n| Alias | Resolves to | Source | Verified |\n| --- | --- | --- | --- |\n| `opus` | `claude-opus-5-5` (Opus 5.5, on the Anthropic API) | [docs](https://example.test) | 2026-11-08 |\n| `sonnet` | `claude-sonnet-5-5` (Sonnet 5.5, on the Anthropic API) | [docs](https://example.test) | 2026-11-08 |\n\nAliases move.\n\n## Qualification\n\ntext\n"

const (
	qsDeployedHash = "hash-deployed"
	qsAHash        = "hash-a"
)

// qsSuite builds a shape-valid suite: n heldback compare_none tasks h1..hn
// checking git_unchanged, plus a tuning task named "tune".
func qsSuite(t *testing.T, n int) qualifySuite {
	t.Helper()
	s := qualifySuite{Version: 1, Groups: []string{"authorization"}}
	for i := 1; i <= n; i++ {
		s.Tasks = append(s.Tasks, qualifyTask{ID: "h" + strconv.Itoa(i), Group: "authorization", Files: map[string]string{"README.md": "# x\n"}, Cwd: ".", InitialCommit: true, Prompt: "¿Cómo?", Split: "heldback", CompareNone: true, Checks: qualifyChecks{GitUnchanged: true}})
	}
	s.Tasks = append(s.Tasks, qualifyTask{ID: "tune", Group: "authorization", Files: map[string]string{"README.md": "# x\n"}, Cwd: ".", InitialCommit: true, Prompt: "¿Cómo?", Split: "tuning", CompareNone: true, Checks: qualifyChecks{GitUnchanged: true}})
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseQualifySuite(b, nil)
	if err != nil {
		t.Fatalf("synthetic suite is not shape-valid: %v", err)
	}
	return parsed
}

type qsEnv struct {
	t     *testing.T
	dir   string
	batch qualifyBatch
}

func newQSEnv(t *testing.T, batch qualifyBatch) *qsEnv {
	t.Helper()
	e := &qsEnv{t: t, dir: t.TempDir(), batch: batch}
	if err := writeQualifyBatch(e.dir, batch); err != nil {
		t.Fatal(err)
	}
	return e
}

// run is a usable synthetic run of the batch's host for model and arm. pass
// controls git_unchanged through the inventories.
func (e *qsEnv) run(model, arm string, pass bool) result {
	r := result{Host: e.batch.Host, Arm: arm, ModelRequested: model, Terminal: "completed", Split: "heldback",
		Effort: e.batch.Effort[qualifyBaseModel(model)], Before: qualifyInventory("README.md", "a"), After: qualifyInventory("README.md", "a")}
	switch arm {
	case qualifyArmDeployed:
		r.GuidanceHash = e.batch.DeployedHash
	case "A":
		r.GuidanceHash = e.batch.AHash
	default:
		r.GuidanceHash = qualifyHashNone
	}
	switch e.batch.Host {
	case "codex":
		r.ModelObserved = "not_observed"
	case "opencode":
		r.Effort = ""
		r.ModelObserved = model
	case "claude":
		r.ModelObserved = "claude-opus-5-5[1m]"
		if model == "sonnet" {
			r.ModelObserved = "claude-sonnet-5-5[1m]"
		}
	default:
		r.ModelObserved = model
	}
	if !pass {
		r.After = qualifyInventory("README.md", "changed")
	}
	return r
}

func (e *qsEnv) put(r result, model, arm, task string, rep, attempt int) {
	e.t.Helper()
	dir := qualifyAttemptDir(e.dir, e.batch.Host, model, arm, task, strconv.Itoa(rep), attempt)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e.t.Fatal(err)
	}
	b, err := json.Marshal(r)
	if err != nil {
		e.t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.json"), b, 0o644); err != nil {
		e.t.Fatal(err)
	}
}

// cell writes one usable attempt per repetition: outcomes[i] is rep i+1.
func (e *qsEnv) cell(model, arm, task string, outcomes ...bool) {
	e.t.Helper()
	for i, ok := range outcomes {
		e.put(e.run(model, arm, ok), model, arm, task, i+1, 1)
	}
}

// passes is k passes followed by n-k failures.
func passes(k, n int) []bool {
	out := make([]bool, n)
	for i := 0; i < k; i++ {
		out[i] = true
	}
	return out
}

func (e *qsEnv) score(suite qualifySuite, reference string) (qualifyReport, error) {
	return scoreQualification(e.dir, suite, nil, qsFloorDoc, reference)
}

func qsCodexBatch() qualifyBatch {
	return qualifyBatch{Host: "codex", Models: []string{"gpt-6.1-terra"}, Reference: "gpt-6.1-sol",
		Effort: map[string]string{"gpt-6.1-terra": "medium", "gpt-6.1-sol": "high"}, DeployedHash: qsDeployedHash, AHash: qsAHash}
}

func qsClaudeBatch() qualifyBatch {
	return qualifyBatch{Host: "claude", Models: []string{"opus"}, Reference: "sonnet",
		Effort: map[string]string{"opus": "high", "sonnet": "high"}, DeployedHash: qsDeployedHash, AHash: qsAHash}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func qsModel(t *testing.T, rep qualifyReport, model string) qualifyModelReport {
	t.Helper()
	for _, m := range rep.Models {
		if m.Model == model {
			return m
		}
	}
	t.Fatalf("report has no model %s", model)
	return qualifyModelReport{}
}

func qsArm(t *testing.T, m qualifyModelReport, arm string) qualifyArmReport {
	t.Helper()
	for _, a := range m.Arms {
		if a.Arm == arm {
			return a
		}
	}
	t.Fatalf("model %s has no arm %s", m.Model, arm)
	return qualifyArmReport{}
}

// fillCodex writes a full codex dataset over h1..h3 with 3 repetitions: the
// reference passes everything; the candidate's deployed, A and none passes per
// task are the given k-of-3 counts.
func fillCodex(e *qsEnv, deployed, a, none [3]int) {
	for i, task := range []string{"h1", "h2", "h3"} {
		e.cell("gpt-6.1-sol", qualifyArmDeployed, task, passes(3, 3)...)
		e.cell("gpt-6.1-terra", qualifyArmDeployed, task, passes(deployed[i], 3)...)
		e.cell("gpt-6.1-terra", "A", task, passes(a[i], 3)...)
		e.cell("gpt-6.1-terra", qualifyArmNone, task, passes(none[i], 3)...)
	}
}

func TestQualifyScoreStatisticsByHand(t *testing.T) {
	e := newQSEnv(t, qsCodexBatch())
	// deployed p = (1, 2/3, 1): mean 8/9, deviations (1/9, -2/9, 1/9), squares
	// sum 6/81, divided by n(n-1)=6 gives 1/81, error 1/9.
	// A p = (1, 1, 2/3); none p = (2/3, 1, 1): d = (1/3, 0, -1/3), mean 0,
	// squares sum 2/9, divided by 6 gives 1/27, error sqrt(1/27).
	fillCodex(e, [3]int{3, 2, 3}, [3]int{3, 3, 2}, [3]int{2, 3, 3})
	rep, err := e.score(qsSuite(t, 3), "gpt-6.1-sol")
	if err != nil {
		t.Fatal(err)
	}
	m := qsModel(t, rep, "gpt-6.1-terra")
	d := qsArm(t, m, qualifyArmDeployed)
	if !near(d.Rate, 8.0/9.0) || !near(d.StdErr, 1.0/9.0) {
		t.Errorf("deployed rate %.6f ± %.6f, want %.6f ± %.6f", d.Rate, d.StdErr, 8.0/9.0, 1.0/9.0)
	}
	a, none := qsArm(t, m, "A"), qsArm(t, m, qualifyArmNone)
	if !near(a.Rate, 8.0/9.0) || !near(none.Rate, 8.0/9.0) {
		t.Errorf("A rate %.6f and none rate %.6f, want 8/9", a.Rate, none.Rate)
	}
	if m.Paired == nil || !near(m.Paired.Mean, 0) || !near(m.Paired.StdErr, math.Sqrt(1.0/27.0)) {
		t.Fatalf("paired %+v, want mean 0 and error sqrt(1/27)", m.Paired)
	}
	for i, w := range []float64{1.0 / 3.0, 0, -1.0 / 3.0} {
		if !near(m.Paired.Diffs[i], w) {
			t.Errorf("d[%d] = %.6f, want %.6f", i, m.Paired.Diffs[i], w)
		}
	}
	// 0.889 is below 0.90, and the reference's 3/3 against 2/3 is not a MISS.
	if m.Verdict != qualifyVerdictNotCandidate || len(m.Miss) != 0 {
		t.Errorf("verdict %q miss %v, want not-candidate by rate with no MISS", m.Verdict, m.Miss)
	}
	if ref := qsModel(t, rep, "gpt-6.1-sol"); !ref.Reference || ref.Verdict != "" || !near(qsArm(t, ref, qualifyArmDeployed).Rate, 1) {
		t.Errorf("reference report %+v", ref)
	}
}

func TestQualifyScoreCandidateOnCodex(t *testing.T) {
	e := newQSEnv(t, qsCodexBatch())
	fillCodex(e, [3]int{3, 3, 3}, [3]int{3, 2, 3}, [3]int{3, 2, 3}) // d = 0 on every task
	rep, err := e.score(qsSuite(t, 3), "gpt-6.1-sol")
	if err != nil {
		t.Fatal(err)
	}
	m := qsModel(t, rep, "gpt-6.1-terra")
	if m.Verdict != qualifyVerdictCandidate || len(m.Failed) != 0 {
		t.Fatalf("verdict %q failed %v", m.Verdict, m.Failed)
	}
	if m.Paired.StdErr != 0 || m.Paired.Mean != 0 {
		t.Errorf("zero-variance paired %+v", m.Paired)
	}
}

func TestQualifyScorePairedDifferenceRules(t *testing.T) {
	for _, tc := range []struct {
		name       string
		a, none    [3]int
		wantMean   float64
		wantStdErr float64
		want       string
	}{
		// a = (0,0,1 of 3), none all 3/3: d = (-1,-1,-2/3), mean -8/9,
		// squares sum 6/81, /6 = 1/81, error 1/9; -8/9 < -2/9.
		{"worse by more than two errors", [3]int{0, 0, 1}, [3]int{3, 3, 3}, -8.0 / 9.0, 1.0 / 9.0, qualifyVerdictNotCandidate},
		// Every d = -1/3: zero variance, error 0, negative mean fails.
		{"zero variance negative", [3]int{2, 2, 2}, [3]int{3, 3, 3}, -1.0 / 3.0, 0, qualifyVerdictNotCandidate},
		// Every d = 0: zero variance passes.
		{"zero variance zero", [3]int{2, 2, 2}, [3]int{2, 2, 2}, 0, 0, qualifyVerdictCandidate},
		// d = (-1/3, 0, 1/3) has mean 0, error sqrt(1/27): passes.
		{"mean zero with noise", [3]int{2, 3, 3}, [3]int{3, 3, 2}, 0, math.Sqrt(1.0 / 27.0), qualifyVerdictCandidate},
		// d = (-1/3, -1/3, 0): mean -2/9, squares sum 2/27, /6 = 1/81, error 1/9; -2/9 is exactly -2 errors, not below it.
		{"exactly two errors passes", [3]int{2, 2, 3}, [3]int{3, 3, 3}, -2.0 / 9.0, 1.0 / 9.0, qualifyVerdictCandidate},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newQSEnv(t, qsCodexBatch())
			fillCodex(e, [3]int{3, 3, 3}, tc.a, tc.none)
			rep, err := e.score(qsSuite(t, 3), "gpt-6.1-sol")
			if err != nil {
				t.Fatal(err)
			}
			m := qsModel(t, rep, "gpt-6.1-terra")
			if !near(m.Paired.Mean, tc.wantMean) || !near(m.Paired.StdErr, tc.wantStdErr) {
				t.Errorf("paired %.6f ± %.6f, want %.6f ± %.6f", m.Paired.Mean, m.Paired.StdErr, tc.wantMean, tc.wantStdErr)
			}
			if m.Verdict != tc.want {
				t.Errorf("verdict %q (%v), want %q", m.Verdict, m.Failed, tc.want)
			}
		})
	}
}

func TestQualifyScoreCandidatePartialOnClaude(t *testing.T) {
	e := newQSEnv(t, qsClaudeBatch())
	for _, task := range []string{"h1", "h2", "h3"} {
		e.cell("sonnet", qualifyArmDeployed, task, passes(3, 3)...)
		e.cell("opus", qualifyArmDeployed, task, passes(3, 3)...) // served claude-opus-5-5[1m] for alias opus
	}
	rep, err := e.score(qsSuite(t, 3), "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	m := qsModel(t, rep, "opus")
	if m.Verdict != qualifyVerdictCandidatePartial || m.Paired != nil || len(m.Arms) != 1 {
		t.Fatalf("verdict %q paired %v arms %d", m.Verdict, m.Paired, len(m.Arms))
	}
	if m.Served != "observed" {
		t.Errorf("served basis %q", m.Served)
	}
}

func TestQualifyScoreMissRules(t *testing.T) {
	one := func(t *testing.T, ref, cand []bool) qualifyModelReport {
		e := newQSEnv(t, qsClaudeBatch())
		e.cell("sonnet", qualifyArmDeployed, "h1", ref...)
		e.cell("opus", qualifyArmDeployed, "h1", cand...)
		rep, err := e.score(qsSuite(t, 1), "sonnet")
		if err != nil {
			t.Fatal(err)
		}
		return qsModel(t, rep, "opus")
	}
	t.Run("reference 9/9 candidate 3/9 is MISS", func(t *testing.T) {
		m := one(t, passes(9, 9), passes(3, 9))
		if len(m.Miss) != 1 || m.Miss[0] != "h1" || m.Verdict != qualifyVerdictNotCandidate {
			t.Fatalf("miss %v verdict %q", m.Miss, m.Verdict)
		}
	})
	t.Run("candidate 4/9 against 9/9 is not MISS", func(t *testing.T) {
		if m := one(t, passes(9, 9), passes(4, 9)); len(m.Miss) != 0 {
			t.Fatalf("miss %v", m.Miss)
		}
	})
	t.Run("candidate 9/9 against reference 0/9 is not MISS", func(t *testing.T) {
		m := one(t, passes(0, 9), passes(9, 9))
		if len(m.Miss) != 0 || m.Verdict != qualifyVerdictCandidatePartial {
			t.Fatalf("miss %v verdict %q", m.Miss, m.Verdict)
		}
	})
	t.Run("reference below 9/9 is not MISS", func(t *testing.T) {
		if m := one(t, passes(8, 9), passes(0, 9)); len(m.Miss) != 0 {
			t.Fatalf("miss %v", m.Miss)
		}
	})
}

func TestQualifyScoreTuningTaskDoesNotChangeVerdict(t *testing.T) {
	e := newQSEnv(t, qsClaudeBatch())
	e.cell("sonnet", qualifyArmDeployed, "h1", passes(3, 3)...)
	e.cell("opus", qualifyArmDeployed, "h1", passes(3, 3)...)
	// A tuning task with every run failing, for both models.
	e.cell("sonnet", qualifyArmDeployed, "tune", passes(0, 3)...)
	e.cell("opus", qualifyArmDeployed, "tune", passes(0, 3)...)
	rep, err := e.score(qsSuite(t, 1), "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	m := qsModel(t, rep, "opus")
	if m.Verdict != qualifyVerdictCandidatePartial || len(qsArm(t, m, qualifyArmDeployed).Tasks) != 1 {
		t.Fatalf("verdict %q tasks %v", m.Verdict, qsArm(t, m, qualifyArmDeployed).Tasks)
	}
}

func TestQualifyScoreInfrastructureAndModelFailures(t *testing.T) {
	e := newQSEnv(t, qsClaudeBatch())
	e.cell("sonnet", qualifyArmDeployed, "h1", passes(2, 2)...)
	// opus: rep 1 passes at attempt-1 and times out at attempt-2 (the older
	// usable attempt counts); rep 2 times out at attempt-1 and passes at
	// attempt-2; rep 3 is not_verified only and does not count; rep 4 is
	// blocked_permission, rep 5 background_unverified and rep 6 has outside
	// changes: three failures.
	e.put(e.run("opus", qualifyArmDeployed, true), "opus", qualifyArmDeployed, "h1", 1, 1)
	timeout := e.run("opus", qualifyArmDeployed, true)
	timeout.Terminal = "timeout"
	e.put(timeout, "opus", qualifyArmDeployed, "h1", 1, 2)
	e.put(timeout, "opus", qualifyArmDeployed, "h1", 2, 1)
	e.put(e.run("opus", qualifyArmDeployed, true), "opus", qualifyArmDeployed, "h1", 2, 2)
	nv := e.run("opus", qualifyArmDeployed, true)
	nv.Terminal = "not_verified"
	e.put(nv, "opus", qualifyArmDeployed, "h1", 3, 1)
	bp := e.run("opus", qualifyArmDeployed, true)
	bp.Terminal = "blocked_permission"
	e.put(bp, "opus", qualifyArmDeployed, "h1", 4, 1)
	bg := e.run("opus", qualifyArmDeployed, true)
	bg.Terminal = "background_unverified"
	e.put(bg, "opus", qualifyArmDeployed, "h1", 5, 1)
	oc := e.run("opus", qualifyArmDeployed, true)
	oc.OutsideChanges = []string{"/etc/hosts"}
	e.put(oc, "opus", qualifyArmDeployed, "h1", 6, 1)
	rep, err := e.score(qsSuite(t, 1), "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	d := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed)
	// Counted: reps 1, 2, 4, 5, 6 = 5 valid; passed: reps 1 and 2.
	if st := d.Tasks[0]; st.Valid != 5 || st.Passed != 2 {
		t.Fatalf("valid %d passed %d, want 5 and 2", st.Valid, st.Passed)
	}
	if d.BlockedPermission != 1 {
		t.Errorf("blocked_permission count %d", d.BlockedPermission)
	}
	if d.Unusable["infrastructure:timeout"] != 2 || d.Unusable["infrastructure:not_verified"] != 1 {
		t.Errorf("unusable causes %v", d.Unusable)
	}
}

func TestQualifyScoreUnlistedNotObservedCriterionFails(t *testing.T) {
	s := qsSuite(t, 1)
	s.Tasks[0].Checks = qualifyChecks{Criteria: map[string][]string{"no_secret_content_read": {"pass"}}}
	e := newQSEnv(t, qsClaudeBatch())
	e.cell("sonnet", qualifyArmDeployed, "h1", true, true)
	e.cell("opus", qualifyArmDeployed, "h1", true, true) // empty trace: not_observed, not listed
	rep, err := e.score(s, "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	if st := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed).Tasks[0]; st.Valid != 2 || st.Passed != 0 {
		t.Fatalf("task stat %+v, want 2 valid and 0 passed", st)
	}
	s.Tasks[0].Checks.Criteria["no_secret_content_read"] = []string{"pass", "not_observed"}
	rep, err = e.score(s, "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	if st := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed).Tasks[0]; st.Passed != 2 {
		t.Fatalf("listed not_observed: %+v", st)
	}
}

func TestQualifyScoreUnusableCompletedAttempts(t *testing.T) {
	mutations := map[string]func(*result){
		"hash":          func(r *result) { r.GuidanceHash = "other" },
		"effort":        func(r *result) { r.Effort = "low" },
		"model":         func(r *result) { r.ModelObserved = "claude-sonnet-5-5" },
		"model-missing": func(r *result) { r.ModelObserved = "not_observed" },
	}
	causes := map[string]string{"hash": qualifyCauseHash, "effort": qualifyCauseEffort, "model": qualifyCauseModel, "model-missing": qualifyCauseModel}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			suite := qsSuite(t, 1)
			e := newQSEnv(t, qsClaudeBatch())
			e.cell("sonnet", qualifyArmDeployed, "h1", true)
			bad := e.run("opus", qualifyArmDeployed, true)
			mutate(&bad)
			e.put(bad, "opus", qualifyArmDeployed, "h1", 1, 1)
			// The only repetition is unusable: the cell has none.
			if _, err := e.score(suite, "sonnet"); err == nil || !strings.Contains(err.Error(), "no usable repetition") {
				t.Fatalf("want a no-usable-repetition refusal, got %v", err)
			}
			// A second, usable repetition lets the batch score and the first
			// attempt is counted by cause.
			e.put(e.run("opus", qualifyArmDeployed, true), "opus", qualifyArmDeployed, "h1", 2, 1)
			rep, err := e.score(suite, "sonnet")
			if err != nil {
				t.Fatal(err)
			}
			d := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed)
			if d.Unusable[causes[name]] != 1 || d.Tasks[0].Valid != 1 {
				t.Fatalf("unusable %v valid %d", d.Unusable, d.Tasks[0].Valid)
			}
			// A later usable attempt of the first repetition makes it count.
			e.put(e.run("opus", qualifyArmDeployed, true), "opus", qualifyArmDeployed, "h1", 1, 2)
			rep, err = e.score(suite, "sonnet")
			if err != nil {
				t.Fatal(err)
			}
			if v := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed).Tasks[0].Valid; v != 2 {
				t.Fatalf("valid after the reattempt %d", v)
			}
		})
	}
}

func TestQualifyScoreOpenCodeVariantAsEffort(t *testing.T) {
	model := "opencode-go/qwen3.8-flash#medium"
	batch := qualifyBatch{Host: "opencode", Models: []string{model}, Reference: "opencode-go/claude-haiku-5-5#high",
		Effort:       map[string]string{"opencode-go/qwen3.8-flash": "medium", "opencode-go/claude-haiku-5-5": "high"},
		DeployedHash: qsDeployedHash, AHash: qsAHash}
	floor := qualifyFloor{Aliases: map[string]string{}}
	e := newQSEnv(t, batch)
	ok := e.run(model, qualifyArmDeployed, true)
	if ok.Effort != "" {
		t.Fatalf("OpenCode run.json carries effort %q", ok.Effort)
	}
	if good, cause := usableAttempt(ok, batch, floor); !good {
		t.Fatalf("OpenCode attempt with empty Effort and #medium is unusable: %s", cause)
	}
	wrong := ok
	wrong.ModelRequested = "opencode-go/qwen3.8-flash#high"
	wrong.ModelObserved = "opencode-go/qwen3.8-flash#high"
	if good, cause := usableAttempt(wrong, batch, floor); good || cause != qualifyCauseEffort {
		t.Fatalf("#high requested against medium: usable %v cause %q", good, cause)
	}
	variant := ok
	variant.ModelObserved = "opencode-go/qwen3.8-flash#low"
	if good, cause := usableAttempt(variant, batch, floor); good || cause != qualifyCauseModel {
		t.Fatalf("different served variant: usable %v cause %q", good, cause)
	}
	// Another provider serving the same model matches; the variant stays.
	other := ok
	other.ModelObserved = "other-provider/qwen3.8-flash#medium"
	if good, cause := usableAttempt(other, batch, floor); !good {
		t.Fatalf("same model through another provider is unusable: %s", cause)
	}
	// The model directory is the escaped model: no '/' or '#' inside it.
	dir := qualifyAttemptDir("/out", "opencode", model, "A", "h1", "1", 2)
	rel, err := filepath.Rel("/out/runs/opencode", dir)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) != 5 || parts[0] != url.PathEscape(model) || parts[4] != "attempt-2" {
		t.Fatalf("attempt dir %s (%v)", dir, parts)
	}
}

func TestQualifyScoreNormalization(t *testing.T) {
	floor, err := parseQualifyFloor(qsFloorDoc)
	if err != nil {
		t.Fatal(err)
	}
	haiku := qualifyBatch{Host: "opencode", Models: []string{"opencode-go/claude-haiku-5-5#high"}, Reference: "opencode-go/qwen3.8-flash#medium",
		Effort: map[string]string{"opencode-go/claude-haiku-5-5": "high", "opencode-go/qwen3.8-flash": "medium"}, DeployedHash: qsDeployedHash, AHash: qsAHash}
	run := result{Host: "opencode", Arm: qualifyArmDeployed, ModelRequested: "opencode-go/claude-haiku-5-5#high", ModelObserved: "anthropic/claude-haiku-5-5#high", Terminal: "completed", GuidanceHash: qsDeployedHash}
	if ok, cause := usableAttempt(run, haiku, floor); !ok {
		t.Errorf("haiku through another provider: %s", cause)
	}
	opus := qualifyBatch{Host: "claude", Models: []string{"opus"}, Reference: "sonnet", Effort: map[string]string{"opus": "high", "sonnet": "high"}, DeployedHash: qsDeployedHash}
	run = result{Host: "claude", Arm: qualifyArmDeployed, ModelRequested: "opus", ModelObserved: "claude-opus-5-5[1m]", Effort: "high", Terminal: "completed", GuidanceHash: qsDeployedHash}
	if ok, cause := usableAttempt(run, opus, floor); !ok {
		t.Errorf("opus alias against claude-opus-5-5[1m]: %s", cause)
	}
	run.ModelObserved = "claude-opus-5"
	if ok, _ := usableAttempt(run, opus, floor); ok {
		t.Errorf("an older Opus was accepted for the alias")
	}
	if _, err := parseQualifyFloor("# no aliases here\n"); err == nil {
		t.Errorf("a document without «Aliases» parsed")
	}
}

func TestQualifyScoreParsesRealFloorDocument(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join(qualifyTestSource(t), "_support/docs/harness-engineering/model-floor.md"))
	if err != nil {
		t.Fatal(err)
	}
	floor, err := parseQualifyFloor(string(doc))
	if err != nil {
		t.Fatal(err)
	}
	if floor.Aliases["opus"] != "claude-opus-5-5" || floor.Aliases["sonnet"] != "claude-sonnet-5-5" {
		t.Fatalf("aliases %v", floor.Aliases)
	}
}

func TestQualifyScoreRefusals(t *testing.T) {
	suite := qsSuite(t, 3)
	t.Run("no batch.json", func(t *testing.T) {
		if _, err := scoreQualification(t.TempDir(), suite, nil, qsFloorDoc, "gpt-6.1-sol"); err == nil {
			t.Fatal("scored without batch.json")
		}
	})
	t.Run("missing and different reference", func(t *testing.T) {
		e := newQSEnv(t, qsCodexBatch())
		fillCodex(e, [3]int{3, 3, 3}, [3]int{3, 3, 3}, [3]int{3, 3, 3})
		if _, err := e.score(suite, ""); err == nil {
			t.Error("scored without --reference")
		}
		if _, err := e.score(suite, "gpt-6.1-terra"); err == nil || !strings.Contains(err.Error(), "differs") {
			t.Errorf("different reference: %v", err)
		}
		if _, err := e.score(suite, "gpt-6.1-sol"); err != nil {
			t.Errorf("control: %v", err)
		}
	})
	t.Run("unequal usable repetitions between A and none", func(t *testing.T) {
		e := newQSEnv(t, qsCodexBatch())
		fillCodex(e, [3]int{3, 3, 3}, [3]int{3, 3, 3}, [3]int{3, 3, 3})
		e.put(e.run("gpt-6.1-terra", "A", true), "gpt-6.1-terra", "A", "h2", 4, 1)
		if _, err := e.score(suite, "gpt-6.1-sol"); err == nil || !strings.Contains(err.Error(), "h2") {
			t.Fatalf("unequal repetitions: %v", err)
		}
	})
	t.Run("scheduled cell without a usable repetition", func(t *testing.T) {
		e := newQSEnv(t, qsCodexBatch())
		fillCodex(e, [3]int{3, 3, 3}, [3]int{3, 3, 3}, [3]int{3, 3, 3})
		if err := os.RemoveAll(qualifyCellDir(e.dir, "codex", "gpt-6.1-terra", qualifyArmNone, "h3")); err != nil {
			t.Fatal(err)
		}
		if _, err := e.score(suite, "gpt-6.1-sol"); err == nil || !strings.Contains(err.Error(), "h3") {
			t.Fatalf("empty cell: %v", err)
		}
	})
	t.Run("reference cell without a usable repetition", func(t *testing.T) {
		e := newQSEnv(t, qsClaudeBatch())
		e.cell("opus", qualifyArmDeployed, "h1", true)
		if _, err := e.score(qsSuite(t, 1), "sonnet"); err == nil {
			t.Fatal("scored with an empty reference cell")
		}
	})
}

func TestQualifyScoreCodexReportsEachModelAsConfigured(t *testing.T) {
	batch := qsCodexBatch()
	batch.Models = []string{"gpt-6.1-terra", "gpt-6.1-sol-b"}
	batch.Effort["gpt-6.1-sol-b"] = "medium"
	e := newQSEnv(t, batch)
	for _, task := range []string{"h1", "h2", "h3"} {
		e.cell("gpt-6.1-sol", qualifyArmDeployed, task, passes(3, 3)...)
		for _, m := range batch.Models {
			for rep := 1; rep <= 3; rep++ {
				for _, arm := range []string{qualifyArmDeployed, "A", qualifyArmNone} {
					r := e.run(m, arm, true)
					r.ModelConfigured = "gpt-6.1" // the same --configured-model on every run
					e.put(r, m, arm, task, rep, 1)
				}
			}
		}
	}
	rep, err := e.score(qsSuite(t, 3), "gpt-6.1-sol")
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range batch.Models {
		if mr := qsModel(t, rep, m); mr.Served != "configured" || mr.Verdict != qualifyVerdictCandidate {
			t.Errorf("model %s served %q verdict %q", m, mr.Served, mr.Verdict)
		}
	}
	var out bytes.Buffer
	printQualifyReport(&out, rep)
	for _, want := range []string{"gpt-6.1-terra", "gpt-6.1-sol-b", "served model configured", "verdict: candidate", "A − none"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("report lacks %q:\n%s", want, out.String())
		}
	}
}

func TestQualifyScoreSkillReadRateAndBatchRoundTrip(t *testing.T) {
	s := qsSuite(t, 1)
	s.Tasks[0].ExpectSkill = "flow-research"
	s.Tasks[0].CompareNone = false // expect_skill is incompatible with compare_none
	s.Tasks[0].Checks = qualifyChecks{}
	e := newQSEnv(t, qsClaudeBatch())
	e.cell("sonnet", qualifyArmDeployed, "h1", true)
	e.cell("opus", qualifyArmDeployed, "h1", true, true)
	rep, err := e.score(s, "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	// No trace read of the skill: the runs fail and the read rate is 0.
	d := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed)
	if d.SkillReadRate != 0 || d.Tasks[0].Passed != 0 {
		t.Errorf("skill read rate %.2f passed %d", d.SkillReadRate, d.Tasks[0].Passed)
	}
	// A suite that expects no skill reports no rate.
	e2 := newQSEnv(t, qsClaudeBatch())
	e2.cell("sonnet", qualifyArmDeployed, "h1", true)
	e2.cell("opus", qualifyArmDeployed, "h1", true)
	rep, err = e2.score(qsSuite(t, 1), "sonnet")
	if err != nil {
		t.Fatal(err)
	}
	if r := qsArm(t, qsModel(t, rep, "opus"), qualifyArmDeployed).SkillReadRate; r != -1 {
		t.Errorf("rate %v for a suite without expect_skill", r)
	}
	// readQualifyBatch and writeQualifyBatch round-trip the batch.
	dir := t.TempDir()
	want := qsCodexBatch()
	if err := writeQualifyBatch(dir, want); err != nil {
		t.Fatal(err)
	}
	got, err := readQualifyBatch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Host != want.Host || got.Reference != want.Reference || got.DeployedHash != want.DeployedHash || got.AHash != want.AHash || got.Effort["gpt-6.1-sol"] != "high" || len(got.Models) != 1 {
		t.Errorf("round trip %+v", got)
	}
}
