package main

// Qualification scorer: aggregates the attempts of a qualification batch into
// per-model, per-arm pass rates, the paired guidance effect, MISS tasks and a
// verdict (design.md «Puntuador»; AC43, AC44, AC57).
//
// API for the driver and the subcommands (T3):
//
//   - qualifyBatch is the type of <out>/batch.json; readQualifyBatch and
//     writeQualifyBatch read and write it.
//   - qualifyAttemptDir builds the directory of one attempt.
//   - parseQualifyFloor reads the «Aliases» table of model-floor.md.
//   - usableAttempt(run, batch, floor) is the one function that decides
//     whether an attempt counts (AC57); the scorer and the driver share it.
//   - scoreQualification(dir, suite, cases, floorDoc, reference) returns the
//     report, or an error when AC57 requires refusing a verdict (exit code 1).
//   - printQualifyReport writes the report.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	qualifyBatchFile       = "batch.json"
	qualifyRateThreshold   = 0.90
	qualifyMissCandidate   = 1.0 / 3.0
	qualifyStdErrFactor    = 2.0
	qualifyEpsilon         = 1e-9
	qualifyServedConfigure = "configured"
	qualifyServedObserved  = "observed"
)

// Verdicts of the scorer (AC44).
const (
	qualifyVerdictCandidate        = "candidate"
	qualifyVerdictCandidatePartial = "candidate-partial"
	qualifyVerdictNotCandidate     = "not-candidate"
)

// qualifyNoneArmHosts are the hosts with a no-guidance arm (model-floor.md
// «Known limits»); only they schedule A and none.
var qualifyNoneArmHosts = map[string]bool{"codex": true, "opencode": true}

// qualifyBatch is <out>/batch.json, written by `qualify run` at the first
// launch and read by the scorer. Models lists the requested model strings as
// the runner receives them (on OpenCode with their `#variant`), and Reference
// the reference model in the same form. Effort maps each model, with any
// `#variant` removed, to its resolved effort; on OpenCode that value is the
// variant, because the runner takes no --effort there. A model absent from
// Effort has no effort ("—" in model-floor.md).
type qualifyBatch struct {
	Host         string            `json:"host"`
	Models       []string          `json:"models"`
	Reference    string            `json:"reference"`
	Effort       map[string]string `json:"effort"`
	DeployedHash string            `json:"deployed_hash"`
	AHash        string            `json:"a_hash"`
	// Split is the task split the batch was launched with; a batch without it
	// (written before the field existed) is a held-back one.
	Split string `json:"split,omitempty"`
}

// qualifyReferenceMatches accepts the batch reference as stored (on OpenCode
// with its `#variant`, after effort resolution) or its base model; a given
// `#variant` must be the batch's.
func qualifyReferenceMatches(given, stored string) bool {
	if given == stored {
		return true
	}
	return !strings.Contains(given, "#") && given == qualifyBaseModel(stored)
}

func readQualifyBatch(dir string) (qualifyBatch, error) {
	var b qualifyBatch
	data, err := os.ReadFile(filepath.Join(dir, qualifyBatchFile))
	if err != nil {
		return b, fmt.Errorf("%s: %w", qualifyBatchFile, err)
	}
	if err := json.Unmarshal(data, &b); err != nil {
		return b, fmt.Errorf("%s: %w", qualifyBatchFile, err)
	}
	if b.Host == "" || b.Reference == "" || len(b.Models) == 0 {
		return b, fmt.Errorf("%s: host, models and reference are required", qualifyBatchFile)
	}
	return b, nil
}

func writeQualifyBatch(dir string, b qualifyBatch) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, qualifyBatchFile), append(data, '\n'), 0o644)
}

// qualifyCellDir is the directory holding the repetitions of one cell
// (model, arm, task).
func qualifyCellDir(dir, host, model, arm, task string) string {
	return filepath.Join(dir, "runs", host, url.PathEscape(model), arm, task)
}

// qualifyAttemptDir is the directory of one attempt of one repetition; its
// run.json is the attempt's evidence.
func qualifyAttemptDir(dir, host, model, arm, task, rep string, attempt int) string {
	return filepath.Join(qualifyCellDir(dir, host, model, arm, task), rep, "attempt-"+strconv.Itoa(attempt))
}

// qualifyFloor holds what the scorer reads from model-floor.md: the alias
// table.
type qualifyFloor struct {
	Aliases map[string]string
}

var qualifyBacktick = regexp.MustCompile("`([^`]+)`")

// parseQualifyFloor reads the «Aliases» table of model-floor.md: the alias is
// the backticked name of the first column and the full ID the first backticked
// token of the «Resolves to» column.
func parseQualifyFloor(doc string) (qualifyFloor, error) {
	f := qualifyFloor{Aliases: map[string]string{}}
	i := strings.Index(doc, "\n## Aliases\n")
	if i < 0 {
		return f, errors.New("model-floor.md has no «## Aliases» section")
	}
	rest := doc[i+len("\n## Aliases\n"):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	header := true
	for _, line := range strings.Split(rest, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		if header {
			header = false
			continue
		}
		if strings.HasPrefix(line, "| ---") || strings.HasPrefix(line, "|---") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		if len(cells) < 2 {
			continue
		}
		alias := qualifyBacktick.FindStringSubmatch(cells[0])
		id := qualifyBacktick.FindStringSubmatch(cells[1])
		if alias == nil || id == nil {
			return f, fmt.Errorf("«Aliases» row %q has no backticked alias or ID", line)
		}
		f.Aliases[alias[1]] = id[1]
	}
	if len(f.Aliases) == 0 {
		return f, errors.New("«## Aliases» holds no rows")
	}
	return f, nil
}

// qualifyBaseModel removes a `#variant` from a requested or served model.
func qualifyBaseModel(model string) string {
	if i := strings.Index(model, "#"); i >= 0 {
		return model[:i]
	}
	return model
}

// qualifyVariant returns the `#variant` of a model, or "".
func qualifyVariant(model string) string {
	if i := strings.Index(model, "#"); i >= 0 {
		return model[i+1:]
	}
	return ""
}

var qualifyBracketSuffix = regexp.MustCompile(`\[[^\]]*\]$`)

// qualifyModelID normalizes a model for comparison: it removes the variant, a
// trailing `[…]` suffix and the provider prefix, then resolves an alias to the
// full ID of «Aliases».
func qualifyModelID(model string, f qualifyFloor) string {
	m := qualifyBaseModel(model)
	m = qualifyBracketSuffix.ReplaceAllString(m, "")
	if i := strings.LastIndex(m, "/"); i >= 0 {
		m = m[i+1:]
	}
	if full, ok := f.Aliases[m]; ok {
		return full
	}
	return m
}

// Causes under which an attempt does not count (AC57).
const (
	qualifyCauseHost   = "host"
	qualifyCauseModel  = "model"
	qualifyCauseHash   = "hash"
	qualifyCauseEffort = "effort"
	qualifyCauseNoRun  = "infrastructure:no_run_json"
	qualifyCauseInfra  = "infrastructure:"
)

// qualifyUsableTerminals are the run states that count: the model finished,
// was blocked by a permission or left a background process (a model failure).
var qualifyUsableTerminals = map[string]bool{"completed": true, "blocked_permission": true, "background_unverified": true}

// usableAttempt decides whether an attempt counts (AC57, D21) and, when it
// does not, why. The cause is `infrastructure:<terminal>`, `host`, `hash`,
// `effort` or `model` (requested model outside the batch, served model absent
// or different, or a different OpenCode variant). The scorer and the driver
// both call it.
func usableAttempt(run result, batch qualifyBatch, floor qualifyFloor) (bool, string) {
	if !qualifyUsableTerminals[run.Terminal] {
		return false, qualifyCauseInfra + run.Terminal
	}
	if run.Host != batch.Host {
		return false, qualifyCauseHost
	}
	known := false
	for _, m := range append([]string{batch.Reference}, batch.Models...) {
		if qualifyBaseModel(m) == qualifyBaseModel(run.ModelRequested) {
			known = true
		}
	}
	if !known {
		return false, qualifyCauseModel
	}
	var wantHash string
	switch run.Arm {
	case qualifyArmDeployed:
		wantHash = batch.DeployedHash
	case "A":
		wantHash = batch.AHash
	case qualifyArmNone:
		wantHash = qualifyHashNone
	default:
		return false, qualifyCauseHash
	}
	if run.GuidanceHash != wantHash {
		return false, qualifyCauseHash
	}
	wantEffort := batch.Effort[qualifyBaseModel(run.ModelRequested)]
	gotEffort := run.Effort
	if batch.Host == "opencode" {
		gotEffort = qualifyVariant(run.ModelRequested)
	}
	if gotEffort != wantEffort {
		return false, qualifyCauseEffort
	}
	if batch.Host == "codex" {
		return true, "" // no served model is observable: the requested one is used
	}
	served := run.ModelObserved
	if served == "" || served == "not_observed" {
		return false, qualifyCauseModel
	}
	if qualifyModelID(served, floor) != qualifyModelID(run.ModelRequested, floor) {
		return false, qualifyCauseModel
	}
	if batch.Host == "opencode" && qualifyVariant(served) != qualifyVariant(run.ModelRequested) {
		return false, qualifyCauseModel
	}
	return true, ""
}

// qualifyTaskStat is the outcome of one task in one model and arm.
type qualifyTaskStat struct {
	Task   string
	Valid  int
	Passed int
	P      float64
}

// qualifyArmReport is one model's result in one arm over the arm's tasks:
// every heldback task for deployed, the compare_none heldback tasks for A and
// none.
type qualifyArmReport struct {
	Arm               string
	Tasks             []qualifyTaskStat
	Rate              float64
	StdErr            float64
	SkillReadRate     float64 // runs that read the expected SKILL.md; -1 when no task expects one
	BlockedPermission int
	Unusable          map[string]int // attempts that did not count, by cause
}

// qualifyPaired is the guidance effect A − none over the compare_none tasks.
type qualifyPaired struct {
	Tasks  []string
	Diffs  []float64
	Mean   float64
	StdErr float64
}

// qualifyModelReport is one model's report. Verdict is empty for the
// reference.
type qualifyModelReport struct {
	Model     string
	Reference bool
	Served    string // «configured» for Codex (requested model), «observed» elsewhere
	Arms      []qualifyArmReport
	Paired    *qualifyPaired
	Miss      []string
	Verdict   string
	Failed    []string
}

// qualifyReport is the output of scoreQualification.
type qualifyReport struct {
	Host      string
	Reference string
	Models    []qualifyModelReport
}

// qualifyCell tallies one cell (model, arm, task).
type qualifyCell struct {
	valid, passed, blocked, skillRuns, skillReads int
	unusable                                      map[string]int
}

// scoreQualification reads <dir>/batch.json and the attempts below
// <dir>/runs, counts the usable attempts and computes the statistics and the
// verdict. It returns an error, and no verdict, when AC57 requires refusing:
// no batch.json or reference, a reference other than the batch's, unequal
// usable repetitions between compared arms, or a scheduled cell without a
// usable repetition. cases may be nil when no task names a flows `case`.
func scoreQualification(dir string, suite qualifySuite, cases map[string]fixture, floorDoc, reference string) (qualifyReport, error) {
	var rep qualifyReport
	if reference == "" {
		return rep, errors.New("--reference is required")
	}
	batch, err := readQualifyBatch(dir)
	if err != nil {
		return rep, err
	}
	if !qualifyReferenceMatches(reference, batch.Reference) {
		return rep, fmt.Errorf("--reference %q differs from the batch reference %q", reference, batch.Reference)
	}
	if batch.Split != "" && batch.Split != qualifySplitHeldback {
		return rep, fmt.Errorf("batch.json was launched with --split %s; qualify score scores held-back batches only", batch.Split)
	}
	floor, err := parseQualifyFloor(floorDoc)
	if err != nil {
		return rep, err
	}
	var heldback, compare []qualifyTask
	for _, t := range suite.Tasks {
		if t.Split != qualifySplitHeldback {
			continue
		}
		heldback = append(heldback, t)
		if t.CompareNone {
			compare = append(compare, t)
		}
	}
	if len(heldback) == 0 {
		return rep, errors.New("the suite has no heldback task")
	}
	hasNone := qualifyNoneArmHosts[batch.Host]
	rep.Host, rep.Reference = batch.Host, batch.Reference

	models := []string{batch.Reference}
	for _, m := range batch.Models {
		if m != batch.Reference {
			models = append(models, m)
		}
	}
	// refStats holds the reference's deployed p_t by task, for MISS.
	refStats := map[string]qualifyTaskStat{}
	var reports []qualifyModelReport
	for _, model := range models {
		isRef := model == batch.Reference
		mr := qualifyModelReport{Model: model, Reference: isRef, Served: qualifyServedObserved}
		if batch.Host == "codex" {
			mr.Served = qualifyServedConfigure
		}
		arms := []struct {
			name  string
			tasks []qualifyTask
		}{{qualifyArmDeployed, heldback}}
		if hasNone && !isRef {
			arms = append(arms, struct {
				name  string
				tasks []qualifyTask
			}{"A", compare}, struct {
				name  string
				tasks []qualifyTask
			}{qualifyArmNone, compare})
		}
		armTasks := map[string][]qualifyTaskStat{}
		for _, a := range arms {
			ar := qualifyArmReport{Arm: a.name, SkillReadRate: -1, Unusable: map[string]int{}}
			var skillRuns, skillReads int
			for _, t := range a.tasks {
				cell, err := scoreQualifyCell(dir, batch, floor, cases, model, a.name, t)
				if err != nil {
					return rep, err
				}
				if cell.valid == 0 {
					return rep, fmt.Errorf("cell %s / %s / %s has no usable repetition", model, a.name, t.ID)
				}
				st := qualifyTaskStat{Task: t.ID, Valid: cell.valid, Passed: cell.passed, P: float64(cell.passed) / float64(cell.valid)}
				ar.Tasks = append(ar.Tasks, st)
				ar.BlockedPermission += cell.blocked
				for cause, n := range cell.unusable {
					ar.Unusable[cause] += n
				}
				skillRuns += cell.skillRuns
				skillReads += cell.skillReads
			}
			if skillRuns > 0 {
				ar.SkillReadRate = float64(skillReads) / float64(skillRuns)
			}
			ar.Rate, ar.StdErr = qualifyMeanStdErr(qualifyPs(ar.Tasks))
			armTasks[a.name] = ar.Tasks
			mr.Arms = append(mr.Arms, ar)
		}
		if isRef {
			for _, st := range armTasks[qualifyArmDeployed] {
				refStats[st.Task] = st
			}
		} else if hasNone {
			a, n := armTasks["A"], armTasks[qualifyArmNone]
			p := &qualifyPaired{}
			for i := range a {
				if a[i].Valid != n[i].Valid {
					return rep, fmt.Errorf("model %s task %s: arm A has %d usable repetitions and none %d", model, a[i].Task, a[i].Valid, n[i].Valid)
				}
				p.Tasks = append(p.Tasks, a[i].Task)
				p.Diffs = append(p.Diffs, a[i].P-n[i].P)
			}
			p.Mean, p.StdErr = qualifyMeanStdErr(p.Diffs)
			mr.Paired = p
		}
		reports = append(reports, mr)
	}
	// Verdicts: held-back tasks only (AC44).
	for i := range reports {
		mr := &reports[i]
		if mr.Reference {
			continue
		}
		deployed := mr.Arms[0]
		if deployed.Rate+qualifyEpsilon < qualifyRateThreshold {
			mr.Failed = append(mr.Failed, fmt.Sprintf("rate %.3f is below %.2f", deployed.Rate, qualifyRateThreshold))
		}
		for _, st := range deployed.Tasks {
			ref, ok := refStats[st.Task]
			if ok && ref.Passed == ref.Valid && st.P <= qualifyMissCandidate+qualifyEpsilon {
				mr.Miss = append(mr.Miss, st.Task)
			}
		}
		if len(mr.Miss) > 0 {
			mr.Failed = append(mr.Failed, "MISS on "+strings.Join(mr.Miss, ", "))
		}
		if hasNone {
			switch {
			case mr.Paired == nil || len(mr.Paired.Diffs) == 0:
				mr.Failed = append(mr.Failed, "no compare_none task to compare A with none")
			case mr.Paired.Mean+qualifyEpsilon < -qualifyStdErrFactor*mr.Paired.StdErr:
				mr.Failed = append(mr.Failed, fmt.Sprintf("A − none is %.3f, below −2 × %.3f", mr.Paired.Mean, mr.Paired.StdErr))
			}
		}
		switch {
		case len(mr.Failed) > 0:
			mr.Verdict = qualifyVerdictNotCandidate
		case hasNone:
			mr.Verdict = qualifyVerdictCandidate
		default:
			mr.Verdict = qualifyVerdictCandidatePartial
		}
	}
	rep.Models = reports
	return rep, nil
}

func qualifyPs(tasks []qualifyTaskStat) []float64 {
	out := make([]float64, len(tasks))
	for i, t := range tasks {
		out[i] = t.P
	}
	return out
}

// qualifyMeanStdErr returns the mean and the standard error between items,
// sqrt(Σ(x − x̄)² / (n(n−1))); with fewer than two items the error is 0.
func qualifyMeanStdErr(xs []float64) (mean, stdErr float64) {
	n := float64(len(xs))
	if n == 0 {
		return 0, 0
	}
	for _, x := range xs {
		mean += x
	}
	mean /= n
	if n < 2 {
		return mean, 0
	}
	var ss float64
	for _, x := range xs {
		ss += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(ss / (n * (n - 1)))
}

// scoreQualifyCell counts the repetitions of one cell: each repetition counts
// with its most recent usable attempt; every attempt that does not count is
// tallied by cause.
func scoreQualifyCell(dir string, batch qualifyBatch, floor qualifyFloor, cases map[string]fixture, model, arm string, t qualifyTask) (qualifyCell, error) {
	cell := qualifyCell{unusable: map[string]int{}}
	root := qualifyCellDir(dir, batch.Host, model, arm, t.ID)
	reps, err := os.ReadDir(root)
	if errors.Is(err, os.ErrNotExist) {
		return cell, nil
	}
	if err != nil {
		return cell, err
	}
	_, _, prompt := resolveQualifyTask(t, cases)
	for _, rep := range reps {
		if !rep.IsDir() {
			continue
		}
		attempts, err := qualifyAttemptNumbers(filepath.Join(root, rep.Name()))
		if err != nil {
			return cell, err
		}
		counted := false
		for _, n := range attempts { // most recent first
			path := filepath.Join(root, rep.Name(), "attempt-"+strconv.Itoa(n), "run.json")
			var run result
			data, err := os.ReadFile(path)
			if err == nil {
				err = json.Unmarshal(data, &run)
			}
			if err != nil {
				cell.unusable[qualifyCauseNoRun]++
				continue
			}
			ok, cause := usableAttempt(run, batch, floor)
			if !ok {
				cell.unusable[cause]++
				continue
			}
			if counted {
				continue
			}
			counted = true
			cell.valid++
			if run.Terminal == "blocked_permission" {
				cell.blocked++
			}
			if t.ExpectSkill != "" {
				cell.skillRuns++
				if len(skillContentReads(run, t.ExpectSkill, false)) > 0 {
					cell.skillReads++
				}
			}
			if qualifyRunPassed(t, run, prompt) {
				cell.passed++
			}
		}
	}
	return cell, nil
}

// qualifyAttemptNumbers lists the attempt numbers of one repetition, highest
// first.
func qualifyAttemptNumbers(repDir string) ([]int, error) {
	ents, err := os.ReadDir(repDir)
	if err != nil {
		return nil, err
	}
	var nums []int
	for _, e := range ents {
		if !e.IsDir() || !strings.HasPrefix(e.Name(), "attempt-") {
			continue
		}
		if n, err := strconv.Atoi(strings.TrimPrefix(e.Name(), "attempt-")); err == nil && n > 0 {
			nums = append(nums, n)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(nums)))
	return nums, nil
}

// qualifyRunPassed applies AC57's pass rule to a usable run: a permission
// block or a background launch is a model failure, files changed outside the
// fixture fail, and the task's checks must all pass.
func qualifyRunPassed(t qualifyTask, run result, prompt string) bool {
	if run.Terminal != "completed" || len(run.OutsideChanges) > 0 {
		return false
	}
	return evaluateQualifyTask(t, run, prompt).Passed
}

// printQualifyReport writes the report: per model and arm the rate with its
// standard error, the unusable attempts by cause, the paired effect, MISS
// tasks, the skill read rate and the verdict.
func printQualifyReport(w io.Writer, rep qualifyReport) {
	fmt.Fprintf(w, "host: %s  reference: %s\n", rep.Host, rep.Reference)
	for _, m := range rep.Models {
		role := "candidate"
		if m.Reference {
			role = "reference"
		}
		fmt.Fprintf(w, "\nmodel %s (%s, served model %s)\n", m.Model, role, m.Served)
		for _, a := range m.Arms {
			fmt.Fprintf(w, "  arm %-8s rate %.3f ± %.3f over %d tasks; blocked_permission runs %d\n", a.Arm, a.Rate, a.StdErr, len(a.Tasks), a.BlockedPermission)
			for _, t := range a.Tasks {
				fmt.Fprintf(w, "    %-32s %d/%d  p=%.3f\n", t.Task, t.Passed, t.Valid, t.P)
			}
			if a.SkillReadRate >= 0 {
				fmt.Fprintf(w, "    expected SKILL.md read rate %.3f\n", a.SkillReadRate)
			}
			if len(a.Unusable) > 0 {
				causes := make([]string, 0, len(a.Unusable))
				for c := range a.Unusable {
					causes = append(causes, c)
				}
				sort.Strings(causes)
				var parts []string
				for _, c := range causes {
					parts = append(parts, fmt.Sprintf("%s=%d", c, a.Unusable[c]))
				}
				fmt.Fprintf(w, "    attempts not counted: %s\n", strings.Join(parts, ", "))
			}
		}
		if m.Paired != nil {
			fmt.Fprintf(w, "  A − none %.3f ± %.3f (paired over %d tasks)\n", m.Paired.Mean, m.Paired.StdErr, len(m.Paired.Diffs))
		}
		if !m.Reference {
			if len(m.Miss) > 0 {
				fmt.Fprintf(w, "  MISS: %s\n", strings.Join(m.Miss, ", "))
			}
			fmt.Fprintf(w, "  verdict: %s\n", m.Verdict)
			for _, f := range m.Failed {
				fmt.Fprintf(w, "    failed: %s\n", f)
			}
		}
	}
}
