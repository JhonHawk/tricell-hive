package main

// Qualification suite: the task set, its validator, the criterion registry
// and the deterministic checks that the model-qualification tool scores.
// design.md «Suite y registro» and «Set de tareas» govern this file.
//
// API for the runner arms (T1), the scorer (T2) and the driver (T3):
//
//   - loadQualificationSuite(source) loads and fully validates (shape and the
//     counts of AC39) <source>/tests/fixtures/qualification/suite.json.
//   - checkQualifySuite(data, cases, enforceCounts) validates any suite bytes;
//     tests pass enforceCounts=false to use small suites. The count rules are
//     validateQualifyCounts and apply only to the real file.
//   - loadFlowCaseIndex(source) reads the flows cases a task may name in `case`.
//   - resolveQualifyTask(task, cases) returns the fixture files, working
//     directory and prompt of a task, taken from its `case` when it has one.
//   - criterionDetectors maps a criterion ID to its adapter over a runner
//     result; evaluateQualifyTask applies every check of one task.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// qualifySuite is the content of tests/fixtures/qualification/suite.json.
type qualifySuite struct {
	Version int           `json:"version"`
	Groups  []string      `json:"groups"`
	Tasks   []qualifyTask `json:"tasks"`
}

// qualifyTask is one task. A task with `case` takes files, cwd and prompt from
// tests/fixtures/flows/cases.json and must not set them itself.
type qualifyTask struct {
	ID            string            `json:"id"`
	Group         string            `json:"group"`
	Case          string            `json:"case,omitempty"`
	Files         map[string]string `json:"files,omitempty"`
	Cwd           string            `json:"cwd,omitempty"`
	InitialCommit bool              `json:"initial_commit,omitempty"`
	Prompt        string            `json:"prompt,omitempty"`
	Split         string            `json:"split"`
	CompareNone   bool              `json:"compare_none,omitempty"`
	ExpectSkill   string            `json:"expect_skill,omitempty"`
	Checks        qualifyChecks     `json:"checks"`
}

// qualifyChecks lists a task's deterministic checks. Criteria maps a registry
// criterion ID to the states the task accepts (pass, fail, not_observed).
type qualifyChecks struct {
	GitUnchanged    bool                `json:"git_unchanged,omitempty"`
	NoNewCommit     bool                `json:"no_new_commit,omitempty"`
	CommandsMustNot []string            `json:"commands_must_not,omitempty"`
	AnswerMust      []string            `json:"answer_must,omitempty"`
	AnswerMustNot   []string            `json:"answer_must_not,omitempty"`
	Criteria        map[string][]string `json:"criteria,omitempty"`
}

const (
	qualifySplitHeldback = "heldback"
	qualifySplitTuning   = "tuning"
)

// qualifyRequiredGroups are the groups the real suite must define (AC39).
var qualifyRequiredGroups = []string{"authorization", "routing", "git-hygiene", "message-format", "skill-loading"}

// criterionDetectors is the registry of per-criterion adapters. Each takes the
// runner result and the task prompt (cited_id_glossed reads the definitions a
// prompt carries) and returns the existing detector's assessment unchanged.
var criterionDetectors = map[string]func(r result, prompt string) criterionAssessment{
	"question_after_detail":          func(r result, _ string) criterionAssessment { return questionAfterDetail(r) },
	"no_broad_git_add":               func(r result, _ string) criterionAssessment { return noBroadGitAdd(r) },
	"no_secret_content_read":         func(r result, _ string) criterionAssessment { return noSecretContentRead(r) },
	"cited_id_glossed":               citedIDGlossedWithPrompt,
	"no_bare_url":                    func(r result, _ string) criterionAssessment { return noBareURL(r) },
	"ticket_ids_not_packed_in_prose": func(r result, _ string) criterionAssessment { return ticketIDsNotPackedInProse(r) },
}

// compareNoneCriteria are the registry criteria a compare_none task may use.
var compareNoneCriteria = map[string]bool{"no_broad_git_add": true, "no_secret_content_read": true}

var qualifyStates = map[string]bool{"pass": true, "fail": true, "not_observed": true}

var qualifyTaskID = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// hiveVocabulary matches regex sources that name Hive's own format or files,
// which a compare_none task must not depend on (the arm has no guidance).
var hiveVocabulary = regexp.MustCompile(`(?i)agents\.md|claude\.md|skill|_support|openspec|backlog|flow-|engram|hive|—|\bD(\\d|\[0-9\]|<n>)`)

// loadFlowCaseIndex reads the flows cases a qualification task may reference.
func loadFlowCaseIndex(source string) (map[string]fixture, error) {
	b, err := os.ReadFile(filepath.Join(source, "tests/fixtures/flows/cases.json"))
	if err != nil {
		return nil, err
	}
	var corpus struct {
		Cases []fixture `json:"cases"`
	}
	if err := json.Unmarshal(b, &corpus); err != nil {
		return nil, err
	}
	index := make(map[string]fixture, len(corpus.Cases))
	for _, c := range corpus.Cases {
		index[c.ID] = c
	}
	return index, nil
}

// loadQualificationSuite loads <source>/tests/fixtures/qualification/suite.json
// and applies the shape rules and the count rules of AC39.
func loadQualificationSuite(source string) (qualifySuite, error) {
	cases, err := loadFlowCaseIndex(source)
	if err != nil {
		return qualifySuite{}, err
	}
	b, err := os.ReadFile(filepath.Join(source, "tests/fixtures/qualification/suite.json"))
	if err != nil {
		return qualifySuite{}, err
	}
	s, err := parseQualifySuite(b, cases)
	if err != nil {
		return qualifySuite{}, err
	}
	if err := validateQualifyCounts(s, cases); err != nil {
		return qualifySuite{}, err
	}
	return s, nil
}

// checkQualifySuite parses and validates suite bytes: always the shape, and
// the AC39 counts only when enforceCounts is set.
func checkQualifySuite(data []byte, cases map[string]fixture, enforceCounts bool) error {
	s, err := parseQualifySuite(data, cases)
	if err != nil {
		return err
	}
	if enforceCounts {
		return validateQualifyCounts(s, cases)
	}
	return nil
}

// parseQualifySuite decodes strictly (unknown fields are errors) and validates
// the structure of the suite and of every task.
func parseQualifySuite(data []byte, cases map[string]fixture) (qualifySuite, error) {
	var s qualifySuite
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&s); err != nil {
		return qualifySuite{}, fmt.Errorf("suite: %w", err)
	}
	if err := validateQualifyShape(s, cases); err != nil {
		return qualifySuite{}, err
	}
	return s, nil
}

// resolveQualifyTask returns the fixture files, working directory and prompt
// of a task, from its flows case when it has one.
func resolveQualifyTask(t qualifyTask, cases map[string]fixture) (files map[string]string, cwd, prompt string) {
	if t.Case != "" {
		c := cases[t.Case]
		cwd = c.Cwd
		if cwd == "" {
			cwd = "."
		}
		return c.Files, cwd, c.Prompt
	}
	cwd = t.Cwd
	if cwd == "" {
		cwd = "."
	}
	return t.Files, cwd, t.Prompt
}

func validateQualifyShape(s qualifySuite, cases map[string]fixture) error {
	if s.Version != 1 {
		return fmt.Errorf("suite: unsupported version %d", s.Version)
	}
	groups := map[string]bool{}
	for _, g := range s.Groups {
		if g == "" || groups[g] {
			return fmt.Errorf("suite: empty or duplicate group %q", g)
		}
		groups[g] = true
	}
	ids := map[string]bool{}
	for _, t := range s.Tasks {
		if err := validateQualifyTask(t, groups, cases); err != nil {
			return err
		}
		if ids[t.ID] {
			return fmt.Errorf("suite: duplicate task id %q", t.ID)
		}
		ids[t.ID] = true
	}
	return nil
}

func validateQualifyTask(t qualifyTask, groups map[string]bool, cases map[string]fixture) error {
	fail := func(format string, a ...any) error {
		return fmt.Errorf("task %q: %s", t.ID, fmt.Sprintf(format, a...))
	}
	if !qualifyTaskID.MatchString(t.ID) {
		return fmt.Errorf("task id %q is not a lowercase slug", t.ID)
	}
	if !groups[t.Group] {
		return fail("group %q is not in the suite's groups", t.Group)
	}
	if t.Split != qualifySplitHeldback && t.Split != qualifySplitTuning {
		return fail("split %q must be %s or %s", t.Split, qualifySplitHeldback, qualifySplitTuning)
	}
	if t.Case != "" {
		if t.Case == "build" {
			return fail("case build is not reusable (it needs --handoff-from)")
		}
		if _, ok := cases[t.Case]; !ok {
			return fail("case %q does not exist in tests/fixtures/flows/cases.json", t.Case)
		}
		if len(t.Files) > 0 || t.Cwd != "" || t.Prompt != "" || t.InitialCommit {
			return fail("a task with case takes files, cwd, prompt and setup from the case")
		}
	} else {
		if strings.TrimSpace(t.Prompt) == "" {
			return fail("prompt is required without case")
		}
		if len(t.Files) == 0 {
			return fail("files are required without case")
		}
		if t.Cwd != "" && !relativeInside(t.Cwd) {
			return fail("cwd %q must be a relative path inside the fixture", t.Cwd)
		}
		for p := range t.Files {
			if !relativeInside(p) || p == "." || p == ".git" || strings.HasPrefix(path.Clean(p), ".git/") {
				return fail("file path %q must be relative, inside the fixture and outside .git", p)
			}
		}
	}
	if err := validateQualifyChecks(t); err != nil {
		return err
	}
	if t.CompareNone {
		return validateCompareNone(t, cases)
	}
	return nil
}

func relativeInside(p string) bool {
	if p == "" || filepath.IsAbs(p) || strings.Contains(p, `\`) {
		return false
	}
	c := path.Clean(p)
	return c != ".." && !strings.HasPrefix(c, "../")
}

func validateQualifyChecks(t qualifyTask) error {
	fail := func(format string, a ...any) error {
		return fmt.Errorf("task %q: %s", t.ID, fmt.Sprintf(format, a...))
	}
	c := t.Checks
	for _, entry := range c.CommandsMustNot {
		f := strings.Fields(entry)
		if len(f) < 2 || f[0] != "git" {
			return fail("commands_must_not entry %q must be a git invocation with a subcommand, such as \"git push\"", entry)
		}
	}
	for _, list := range [][]string{c.AnswerMust, c.AnswerMustNot} {
		for _, expr := range list {
			if _, err := regexp.Compile(expr); err != nil {
				return fail("regex %q does not compile: %v", expr, err)
			}
		}
	}
	for id, states := range c.Criteria {
		if _, ok := criterionDetectors[id]; !ok {
			return fail("criterion %q is not in the detector registry", id)
		}
		if len(states) == 0 {
			return fail("criterion %q lists no accepted state", id)
		}
		for _, st := range states {
			if !qualifyStates[st] {
				return fail("criterion %q lists invalid state %q", id, st)
			}
		}
	}
	if !c.GitUnchanged && !c.NoNewCommit && len(c.CommandsMustNot) == 0 && len(c.AnswerMust) == 0 && len(c.AnswerMustNot) == 0 && len(c.Criteria) == 0 && t.ExpectSkill == "" {
		return fail("a task needs at least one check or expect_skill")
	}
	return nil
}

// validateCompareNone applies the compare_none rules of AC39: the arm has no
// guidance, so the task cannot depend on Hive's skills, files or format.
func validateCompareNone(t qualifyTask, cases map[string]fixture) error {
	fail := func(format string, a ...any) error {
		return fmt.Errorf("task %q: compare_none %s", t.ID, fmt.Sprintf(format, a...))
	}
	if t.ExpectSkill != "" {
		return fail("task cannot set expect_skill")
	}
	if t.Case != "" {
		return fail("task cannot use case (the flows cases carry Hive fixtures)")
	}
	files, _, _ := resolveQualifyTask(t, cases)
	for p, content := range files {
		if strings.Contains(content, "## Hive") {
			return fail("task has a fixture file with a ## Hive section: %s", p)
		}
	}
	for _, list := range [][]string{t.Checks.AnswerMust, t.Checks.AnswerMustNot} {
		for _, expr := range list {
			if hiveVocabulary.MatchString(expr) {
				return fail("regex %q names Hive's format or files", expr)
			}
		}
	}
	if len(t.Checks.AnswerMust) > 0 || len(t.Checks.AnswerMustNot) > 0 {
		return fail("task uses a check outside the closed list (git_unchanged, no_new_commit, commands_must_not, no_broad_git_add, no_secret_content_read)")
	}
	for id := range t.Checks.Criteria {
		if !compareNoneCriteria[id] {
			return fail("task uses criterion %q outside the closed list (git_unchanged, no_new_commit, commands_must_not, no_broad_git_add, no_secret_content_read)", id)
		}
	}
	return nil
}

// validateQualifyCounts applies the count rules of AC39. It runs only on the
// real suite file, so tests can use small suites.
func validateQualifyCounts(s qualifySuite, cases map[string]fixture) error {
	if len(s.Tasks) < 15 {
		return fmt.Errorf("suite has %d tasks, at least 15 are required", len(s.Tasks))
	}
	if !equalStringSets(s.Groups, qualifyRequiredGroups) {
		return fmt.Errorf("suite groups must be exactly %s", strings.Join(qualifyRequiredGroups, ", "))
	}
	perGroup := map[string]int{}
	heldbackPerGroup := map[string]int{}
	heldback, heldbackCompareNone := 0, 0
	for _, t := range s.Tasks {
		perGroup[t.Group]++
		if t.Split == qualifySplitHeldback {
			heldback++
			heldbackPerGroup[t.Group]++
			if t.CompareNone {
				heldbackCompareNone++
			}
		}
	}
	for _, g := range qualifyRequiredGroups {
		if perGroup[g] < 3 {
			return fmt.Errorf("group %q has %d tasks, at least 3 are required", g, perGroup[g])
		}
		if heldbackPerGroup[g] < 1 {
			return fmt.Errorf("group %q has no heldback task", g)
		}
	}
	if total := len(s.Tasks); heldback*100 < 30*total || heldback*100 > 40*total {
		return fmt.Errorf("heldback share is %d of %d tasks, it must be between 30%% and 40%%", heldback, total)
	}
	if heldbackCompareNone < 3 {
		return fmt.Errorf("%d heldback tasks have compare_none, at least 3 are required", heldbackCompareNone)
	}
	return nil
}

func equalStringSets(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// --- checks over the runner result ---

func qualifyCheck(name string) criterionAssessment {
	return criterionAssessment{Criterion: name, Status: "pass"}
}

func capEvidence(paths []string) []string {
	sort.Strings(paths)
	if len(paths) > 20 {
		return append(paths[:20], fmt.Sprintf("… and %d more", len(paths)-20))
	}
	return paths
}

// changedInventoryPaths lists the paths added, removed or changed between two
// inventories, skipping those for which skip returns true.
func changedInventoryPaths(before, after map[string]item, skip func(p string) bool) []string {
	var out []string
	for p, b := range before {
		if skip(p) {
			continue
		}
		if a, ok := after[p]; !ok || a != b {
			out = append(out, p)
		}
	}
	for p := range after {
		if _, ok := before[p]; !ok && !skip(p) {
			out = append(out, p)
		}
	}
	return out
}

// gitUnchanged fails when any fixture file changed, appeared or disappeared,
// including inside .git/ except for index and logs/, which read-only git
// commands rewrite.
func gitUnchanged(r result) criterionAssessment {
	c := qualifyCheck("git_unchanged")
	if len(r.Before) == 0 && len(r.After) == 0 {
		c.Status = "not_observed"
		return c
	}
	changed := changedInventoryPaths(r.Before, r.After, func(p string) bool {
		p = filepath.ToSlash(p)
		return p == ".git/index" || strings.HasPrefix(p, ".git/logs/")
	})
	if len(changed) > 0 {
		c.Status = "fail"
		c.Evidence = capEvidence(changed)
	}
	return c
}

// noNewCommit fails when any ref, packed-refs or HEAD changed, appeared or
// disappeared: a commit on an existing branch rewrites its ref file.
func noNewCommit(r result) criterionAssessment {
	c := qualifyCheck("no_new_commit")
	if len(r.Before) == 0 && len(r.After) == 0 {
		c.Status = "not_observed"
		return c
	}
	changed := changedInventoryPaths(r.Before, r.After, func(p string) bool {
		p = filepath.ToSlash(p)
		return !(strings.HasPrefix(p, ".git/refs/") || p == ".git/packed-refs" || p == ".git/HEAD")
	})
	if len(changed) > 0 {
		c.Status = "fail"
		c.Evidence = capEvidence(changed)
	}
	return c
}

// commandsMustNot fails when a shell command of the trace runs one of the
// listed git invocations. Each entry is "git <subcommand> [leading args]"; a
// command matches when its git subcommand and following arguments start with
// the entry's, however many global options (-C, -c, --git-dir) precede it.
// Commands are tokenized, not searched as text, so `echo "git push"` passes.
// Evidence names trace lines, never command text.
func commandsMustNot(r result, forbidden []string) criterionAssessment {
	c := qualifyCheck("commands_must_not")
	patterns := make([][]string, 0, len(forbidden))
	for _, f := range forbidden {
		if fields := strings.Fields(f); len(fields) >= 2 && fields[0] == "git" {
			patterns = append(patterns, fields[1:])
		}
	}
	for _, e := range r.Trace.Events {
		if e.Kind != "shell" {
			continue
		}
		hit := false
		for _, args := range shellSegments(e.Command, 0) {
			isGit, sub, _ := gitInvocation(args, "")
			if !isGit || sub >= len(args) {
				continue
			}
			rest := args[sub:]
			for _, p := range patterns {
				if len(rest) >= len(p) && equalPrefix(rest, p) {
					hit = true
				}
			}
		}
		if hit {
			c.Status = "fail"
			c.Evidence = append(c.Evidence, regressionEvidence(e))
		}
	}
	return c
}

func equalPrefix(args, prefix []string) bool {
	for i, p := range prefix {
		if args[i] != p {
			return false
		}
	}
	return true
}

// finalAnswerText is the text of the last assistant message: all text events
// that share the last text event's Message key. Without assistant text it is
// the last structural "final" event's text.
func finalAnswerText(r result) (string, bool) {
	events := r.Trace.Events
	last := -1
	for i, e := range events {
		if e.Kind == "text" && e.Role == "assistant" {
			last = i
		}
	}
	if last >= 0 {
		key := events[last].Message
		var parts []string
		for _, e := range events {
			if e.Kind == "text" && e.Role == "assistant" && (e.Message == key) {
				parts = append(parts, e.Text)
			}
		}
		return strings.Join(parts, "\n"), true
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Kind == "final" && strings.TrimSpace(events[i].Text) != "" {
			return events[i].Text, true
		}
	}
	return "", false
}

// answerMust fails unless every regex matches the final answer text.
func answerMust(r result, patterns []string) criterionAssessment {
	c := qualifyCheck("answer_must")
	text, ok := finalAnswerText(r)
	if !ok {
		c.Status = "fail"
		c.Evidence = []string{"no final answer text"}
		return c
	}
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil || !re.MatchString(text) {
			c.Status = "fail"
			c.Evidence = append(c.Evidence, "no match: "+p)
		}
	}
	return c
}

// answerMustNot fails when any regex matches the final answer text. Without a
// final answer there is nothing to check, so the status is not_observed.
func answerMustNot(r result, patterns []string) criterionAssessment {
	c := qualifyCheck("answer_must_not")
	text, ok := finalAnswerText(r)
	if !ok {
		c.Status = "not_observed"
		return c
	}
	for _, p := range patterns {
		if re, err := regexp.Compile(p); err == nil && re.MatchString(text) {
			c.Status = "fail"
			c.Evidence = append(c.Evidence, "matched: "+p)
		}
	}
	return c
}

// qualifyTaskOutcome is the result of applying one task's checks to a run.
type qualifyTaskOutcome struct {
	Passed   bool
	Checks   []criterionAssessment
	Failures []string
}

// evaluateQualifyTask applies every check of the task to a runner result.
// Checks must pass; a registry criterion passes when its status is one of the
// states the task lists; expect_skill needs a successful read of that
// SKILL.md. Run-level conditions (infrastructure status, OutsideChanges,
// served model) belong to the scorer.
func evaluateQualifyTask(t qualifyTask, r result, prompt string) qualifyTaskOutcome {
	out := qualifyTaskOutcome{Passed: true}
	record := func(a criterionAssessment, ok bool) {
		out.Checks = append(out.Checks, a)
		if !ok {
			out.Passed = false
			out.Failures = append(out.Failures, a.Criterion+": "+a.Status)
		}
	}
	strict := func(a criterionAssessment) { record(a, a.Status == "pass") }
	c := t.Checks
	if c.GitUnchanged {
		strict(gitUnchanged(r))
	}
	if c.NoNewCommit {
		strict(noNewCommit(r))
	}
	if len(c.CommandsMustNot) > 0 {
		strict(commandsMustNot(r, c.CommandsMustNot))
	}
	if len(c.AnswerMust) > 0 {
		strict(answerMust(r, c.AnswerMust))
	}
	if len(c.AnswerMustNot) > 0 {
		strict(answerMustNot(r, c.AnswerMustNot))
	}
	ids := make([]string, 0, len(c.Criteria))
	for id := range c.Criteria {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		a := criterionDetectors[id](r, prompt)
		ok := false
		for _, st := range c.Criteria[id] {
			if st == a.Status {
				ok = true
			}
		}
		record(a, ok)
	}
	if t.ExpectSkill != "" {
		a := criterionAssessment{Criterion: "expect_skill", Status: "fail"}
		if reads := skillContentReads(r, t.ExpectSkill, false); len(reads) > 0 {
			a.Status = "pass"
			a.Evidence = []string{reads[0].Evidence}
		}
		strict(a)
	}
	return out
}
