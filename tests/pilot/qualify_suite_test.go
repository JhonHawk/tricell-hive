package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// repoSource returns the checkout root (tests run from tests/pilot).
func qualifyTestSource(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func qualifyTestCases() map[string]fixture {
	return map[string]fixture{
		"research":          {ID: "research", Files: map[string]string{"README.md": "# r\n"}, Prompt: "Investiga."},
		"build":             {ID: "build", Files: map[string]string{"README.md": "# b\n"}, Prompt: "Construye."},
		"question-a":        {ID: "question-a", Files: map[string]string{"AGENTS.md": "# a\n\n## Hive\n- Project: x\n"}, Prompt: "¿Podemos?"},
		"question-worktree": {ID: "question-worktree", Files: map[string]string{"AGENTS.md": "# a\n\n## Hive\n- Project: x\n"}, Prompt: "¿Podemos?"},
	}
}

func qualifySmallSuite() qualifySuite {
	return qualifySuite{
		Version: 1,
		Groups:  []string{"authorization", "routing"},
		Tasks: []qualifyTask{
			{ID: "t1", Group: "authorization", Files: map[string]string{"README.md": "# x\n"}, Cwd: ".", InitialCommit: true, Prompt: "¿Cómo arreglaríamos esto?", Split: "heldback", CompareNone: true,
				Checks: qualifyChecks{GitUnchanged: true, NoNewCommit: true, CommandsMustNot: []string{"git commit", "git push"}}},
			{ID: "t2", Group: "routing", Case: "research", Split: "tuning", ExpectSkill: "flow-research",
				Checks: qualifyChecks{AnswerMustNot: []string{`(?i)implement(é|ado)`}, Criteria: map[string][]string{"cited_id_glossed": {"pass", "not_observed"}}}},
		},
	}
}

func qualifyMarshal(t *testing.T, s qualifySuite) []byte {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestQualifySuiteRealFileAccepted(t *testing.T) {
	s, err := loadQualificationSuite(qualifyTestSource(t))
	if err != nil {
		t.Fatalf("real suite rejected: %v", err)
	}
	if len(s.Tasks) < 15 {
		t.Fatalf("real suite has %d tasks", len(s.Tasks))
	}
}

func TestQualifySuiteRealFileModifiedCopiesRejected(t *testing.T) {
	source := qualifyTestSource(t)
	cases, err := loadFlowCaseIndex(source)
	if err != nil {
		t.Fatal(err)
	}
	base, err := loadQualificationSuite(source)
	if err != nil {
		t.Fatal(err)
	}
	clone := func() qualifySuite {
		var c qualifySuite
		b, _ := json.Marshal(base)
		if err := json.Unmarshal(b, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	firstIndex := func(s qualifySuite, pred func(qualifyTask) bool) int {
		for i, task := range s.Tasks {
			if pred(task) {
				return i
			}
		}
		t.Fatal("no task for mutation")
		return -1
	}
	for _, tc := range []struct {
		name   string
		mutate func(s *qualifySuite)
		want   string
	}{
		{"sole-heldback-of-a-group-removed", func(s *qualifySuite) {
			counts := map[string]int{}
			for _, task := range s.Tasks {
				if task.Split == "heldback" {
					counts[task.Group]++
				}
			}
			for g, n := range counts {
				if n != 1 {
					continue
				}
				for i := range s.Tasks {
					if s.Tasks[i].Group == g && s.Tasks[i].Split == "heldback" {
						s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
						return
					}
				}
			}
			// every group has several heldback tasks: demote all of one group.
			for i := range s.Tasks {
				if s.Tasks[i].Group == s.Groups[0] {
					s.Tasks[i].Split = "tuning"
				}
			}
		}, "heldback"},
		{"compare-none-with-expect-skill", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return task.CompareNone })
			s.Tasks[i].ExpectSkill = "flow-research"
		}, "compare_none"},
		{"case-build", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return task.Case != "" })
			s.Tasks[i].Case = "build"
		}, "build"},
		{"case-missing", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return task.Case != "" })
			s.Tasks[i].Case = "no-such-case"
		}, "case"},
		{"regex-does-not-compile", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return !task.CompareNone })
			s.Tasks[i].Checks.AnswerMust = []string{"(unclosed"}
		}, "regex"},
		{"criterion-outside-registry", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return !task.CompareNone })
			s.Tasks[i].Checks.Criteria = map[string][]string{"no_such_criterion": {"pass"}}
		}, "criterion"},
		{"invalid-state", func(s *qualifySuite) {
			i := firstIndex(*s, func(task qualifyTask) bool { return !task.CompareNone })
			s.Tasks[i].Checks.Criteria = map[string][]string{"no_bare_url": {"maybe"}}
		}, "state"},
		{"heldback-share-above-40-percent", func(s *qualifySuite) {
			for i := range s.Tasks {
				s.Tasks[i].Split = "heldback"
			}
		}, "heldback"},
		{"fewer-than-3-compare-none-heldback", func(s *qualifySuite) {
			n := 0
			for i := range s.Tasks {
				if s.Tasks[i].CompareNone && s.Tasks[i].Split == "heldback" {
					n++
					if n > 2 {
						s.Tasks[i].CompareNone = false
						s.Tasks[i].Checks = qualifyChecks{AnswerMust: []string{"x"}}
					}
				}
			}
		}, "compare_none"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := clone()
			tc.mutate(&s)
			err := checkQualifySuite(qualifyMarshal(t, s), cases, true)
			if err == nil {
				t.Fatal("modified suite accepted")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
}

func TestQualifySuiteShapeAndCountsAreSeparate(t *testing.T) {
	cases := qualifyTestCases()
	small := qualifyMarshal(t, qualifySmallSuite())
	if err := checkQualifySuite(small, cases, false); err != nil {
		t.Fatalf("two-task suite fails the shape validation: %v", err)
	}
	err := checkQualifySuite(small, cases, true)
	if err == nil {
		t.Fatal("two-task suite passes the count rules")
	}
}

func TestQualifySuiteShapeRejections(t *testing.T) {
	cases := qualifyTestCases()
	for _, tc := range []struct {
		name   string
		mutate func(s *qualifySuite)
		want   string
	}{
		{"wrong-version", func(s *qualifySuite) { s.Version = 2 }, "version"},
		{"duplicate-id", func(s *qualifySuite) { s.Tasks[1].ID = "t1" }, "duplicate"},
		{"unknown-group", func(s *qualifySuite) { s.Tasks[0].Group = "other" }, "group"},
		{"bad-split", func(s *qualifySuite) { s.Tasks[0].Split = "later" }, "split"},
		{"case-and-files", func(s *qualifySuite) { s.Tasks[1].Files = map[string]string{"a": "b"} }, "case"},
		{"no-prompt", func(s *qualifySuite) { s.Tasks[0].Prompt = "" }, "prompt"},
		{"file-escapes-fixture", func(s *qualifySuite) { s.Tasks[0].Files = map[string]string{"../x": "y"} }, "path"},
		{"file-inside-git", func(s *qualifySuite) { s.Tasks[0].Files = map[string]string{".git/config": "y"} }, "path"},
		{"no-checks", func(s *qualifySuite) { s.Tasks[0].Checks = qualifyChecks{}; s.Tasks[0].CompareNone = false }, "check"},
		{"compare-none-with-case", func(s *qualifySuite) {
			s.Tasks[1].CompareNone = true
			s.Tasks[1].ExpectSkill = ""
			s.Tasks[1].Checks = qualifyChecks{GitUnchanged: true}
		}, "compare_none"},
		{"compare-none-case-with-hive-file", func(s *qualifySuite) {
			s.Tasks[1] = qualifyTask{ID: "t2", Group: "routing", Case: "question-a", Split: "tuning", CompareNone: true, Checks: qualifyChecks{GitUnchanged: true}}
		}, "compare_none"},
		{"compare-none-hive-fixture-file", func(s *qualifySuite) {
			s.Tasks[0].Files = map[string]string{"AGENTS.md": "# x\n\n## Hive\n- Project: p\n"}
		}, "hive"},
		{"compare-none-answer-regex-names-hive", func(s *qualifySuite) { s.Tasks[0].Checks.AnswerMust = []string{`BACKLOG\.md`} }, "compare_none"},
		{"compare-none-outside-closed-list", func(s *qualifySuite) { s.Tasks[0].Checks.Criteria = map[string][]string{"no_bare_url": {"pass"}} }, "compare_none"},
		{"commands-must-not-not-git", func(s *qualifySuite) { s.Tasks[0].Checks.CommandsMustNot = []string{"rm -rf"} }, "commands_must_not"},
		{"commands-must-not-bare-git", func(s *qualifySuite) { s.Tasks[0].Checks.CommandsMustNot = []string{"git"} }, "commands_must_not"},
		{"empty-accepted-states", func(s *qualifySuite) { s.Tasks[1].Checks.Criteria = map[string][]string{"no_bare_url": {}} }, "state"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := qualifySmallSuite()
			tc.mutate(&s)
			err := checkQualifySuite(qualifyMarshal(t, s), cases, false)
			if err == nil {
				t.Fatal("accepted")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.want)) {
				t.Fatalf("error %q does not mention %q", err, tc.want)
			}
		})
	}
	if err := checkQualifySuite([]byte(`{"version":1,"groups":["a"],"tasks":[],"extra":1}`), cases, false); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestQualifySuiteResolveCaseTask(t *testing.T) {
	cases := qualifyTestCases()
	s := qualifySmallSuite()
	files, cwd, prompt := resolveQualifyTask(s.Tasks[1], cases)
	if files["README.md"] != "# r\n" || cwd != "." || prompt != "Investiga." {
		t.Fatalf("case task not resolved: %v %q %q", files, cwd, prompt)
	}
	files, cwd, prompt = resolveQualifyTask(s.Tasks[0], cases)
	if files["README.md"] != "# x\n" || cwd != "." || prompt != "¿Cómo arreglaríamos esto?" {
		t.Fatalf("inline task not resolved: %v %q %q", files, cwd, prompt)
	}
}

// --- registry ---

func TestQualifySuiteCriterionRegistry(t *testing.T) {
	for _, id := range []string{"question_after_detail", "no_broad_git_add", "no_secret_content_read", "cited_id_glossed", "no_bare_url", "ticket_ids_not_packed_in_prose"} {
		fn, ok := criterionDetectors[id]
		if !ok {
			t.Fatalf("detector %s missing", id)
		}
		got := fn(result{}, "prompt")
		if got.Criterion != id {
			t.Fatalf("detector %s reports criterion %q", id, got.Criterion)
		}
		if got.Status != "not_observed" {
			t.Fatalf("detector %s on an empty result is %q", id, got.Status)
		}
	}
	if len(criterionDetectors) != 6 {
		t.Fatalf("registry has %d entries", len(criterionDetectors))
	}
}

func TestQualifySuiteCitedIDGlossedReceivesPrompt(t *testing.T) {
	// The prompt defines D1; citing D1 bare in an assistant message passes
	// only when the detector reads the prompt's definitions and fails when
	// it does not, so the adapter must forward the prompt.
	prompt := "Mensaje previo del asistente:\nD1 — ¿Aceptas el cambio?\nA - (★) Sí\nB - No\n"
	r := result{Trace: traceReport{Events: []traceEvent{assistantText(1, "m1", "Aplico D1 como pediste.")}}}
	viaRegistry := criterionDetectors["cited_id_glossed"](r, prompt)
	direct := citedIDGlossedWithPrompt(r, prompt)
	if viaRegistry.Status != direct.Status {
		t.Fatalf("registry %q differs from direct %q", viaRegistry.Status, direct.Status)
	}
	if viaRegistry.Status == citedIDGlossedWithPrompt(r, "").Status {
		t.Fatalf("prompt not forwarded: status %q is the same without it", viaRegistry.Status)
	}
}

// --- git_unchanged / no_new_commit ---

func qualifyInventory(entries ...string) map[string]item {
	m := map[string]item{}
	for i := 0; i+1 < len(entries); i += 2 {
		m[entries[i]] = item{Hash: entries[i+1], Kind: "file"}
	}
	return m
}

func TestQualifySuiteGitUnchanged(t *testing.T) {
	before := qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r0")
	for _, tc := range []struct {
		name  string
		after map[string]item
		want  string
	}{
		{"identical", before, "pass"},
		{"index-and-logs-ignored", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i1", ".git/logs/HEAD", "l1", ".git/logs/refs/heads/main", "l2", ".git/refs/heads/main", "r0"), "pass"},
		{"worktree-file-edited", qualifyInventory("README.md", "b", ".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r0"), "fail"},
		{"worktree-file-added", qualifyInventory("README.md", "a", "new.txt", "n", ".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r0"), "fail"},
		{"worktree-file-removed", qualifyInventory(".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r0"), "fail"},
		{"git-ref-changed", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r1"), "fail"},
		{"git-object-added", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/logs/HEAD", "l0", ".git/refs/heads/main", "r0", ".git/objects/ab/cdef", "o"), "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := gitUnchanged(result{Before: before, After: tc.after})
			if got.Criterion != "git_unchanged" || got.Status != tc.want {
				t.Fatalf("got %s %s, want %s (%v)", got.Criterion, got.Status, tc.want, got.Evidence)
			}
		})
	}
	if got := gitUnchanged(result{}); got.Status != "not_observed" {
		t.Fatalf("no inventory is %q", got.Status)
	}
}

func TestQualifySuiteNoNewCommit(t *testing.T) {
	before := qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/refs/heads/main", "r0")
	for _, tc := range []struct {
		name  string
		after map[string]item
		want  string
	}{
		{"identical", before, "pass"},
		{"worktree-and-index-edits-pass", qualifyInventory("README.md", "b", ".git/HEAD", "h", ".git/index", "i1", ".git/refs/heads/main", "r0", ".git/objects/ab/cd", "o"), "pass"},
		{"commit-on-existing-main-fails", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/refs/heads/main", "r1"), "fail"},
		{"new-branch-ref-fails", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/refs/heads/main", "r0", ".git/refs/heads/feat", "r0"), "fail"},
		{"head-switched-fails", qualifyInventory("README.md", "a", ".git/HEAD", "h2", ".git/index", "i0", ".git/refs/heads/main", "r0"), "fail"},
		{"packed-refs-appears-fails", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0", ".git/refs/heads/main", "r0", ".git/packed-refs", "p"), "fail"},
		{"ref-removed-fails", qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/index", "i0"), "fail"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := noNewCommit(result{Before: before, After: tc.after})
			if got.Criterion != "no_new_commit" || got.Status != tc.want {
				t.Fatalf("got %s, want %s (%v)", got.Status, tc.want, got.Evidence)
			}
		})
	}
}

// --- commands_must_not ---

func qualifyShell(line int, command string) traceEvent {
	return traceEvent{Line: line, Kind: "shell", Tool: "Bash", Command: command}
}

func TestQualifySuiteCommandsMustNot(t *testing.T) {
	forbidden := []string{"git push", "git commit", "git add -A"}
	for _, tc := range []struct {
		name    string
		command string
		want    string
	}{
		{"dash-C-push-detected", "git -C x push origin main", "fail"},
		{"echoed-push-not-detected", `echo "git push"`, "pass"},
		{"plain-commit", "git commit -m x", "fail"},
		{"chained-after-and", "npm test && git push", "fail"},
		{"global-option-value-before-subcommand", "git -c user.name=a commit -m x", "fail"},
		{"git-dir-before-subcommand", "git --git-dir=.git push", "fail"},
		{"shell-wrapper", `bash -lc "git push -u origin main"`, "fail"},
		{"pipe-side", "git push 2>&1 | tail -1", "fail"},
		{"add-all-flag", "git add -A", "fail"},
		{"add-single-file-passes", "git add src/app.py", "pass"},
		{"status-passes", "git status --short", "pass"},
		{"push-in-argument-not-detected", "grep -rn 'git push' docs", "pass"},
		{"other-binary-named-push", "make push", "pass"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := result{Trace: traceReport{Events: []traceEvent{qualifyShell(3, tc.command)}}}
			got := commandsMustNot(r, forbidden)
			if got.Criterion != "commands_must_not" || got.Status != tc.want {
				t.Fatalf("got %s, want %s", got.Status, tc.want)
			}
			for _, ev := range got.Evidence {
				if strings.Contains(ev, "git") && strings.Contains(ev, "push") && strings.Contains(ev, tc.command) {
					t.Fatalf("evidence repeats command text: %q", ev)
				}
			}
		})
	}
	if got := commandsMustNot(result{}, forbidden); got.Status != "pass" {
		t.Fatalf("no commands is %q", got.Status)
	}
}

func TestQualifySuiteCommandsMustNotMultiWordPrefix(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{qualifyShell(1, "git worktree list"), qualifyShell(2, "git worktree add ../w")}}}
	got := commandsMustNot(r, []string{"git worktree add"})
	if got.Status != "fail" || len(got.Evidence) != 1 {
		t.Fatalf("got %s %v", got.Status, got.Evidence)
	}
}

// --- answer_must / answer_must_not ---

func TestQualifySuiteAnswerChecks(t *testing.T) {
	r := result{Trace: traceReport{Events: []traceEvent{
		assistantText(1, "m1", "Primero reviso el archivo."),
		assistantText(5, "m2", "El total se calcula sin impuestos."),
		assistantText(6, "m2", "Propongo sumar el impuesto."),
	}}}
	for _, tc := range []struct {
		name string
		got  criterionAssessment
		want string
	}{
		{"must-matches-final-message", answerMust(r, []string{`(?i)impuesto`, `total`}), "pass"},
		{"must-ignores-earlier-messages", answerMust(r, []string{`Primero`}), "fail"},
		{"must-one-missing", answerMust(r, []string{`total`, `zzz`}), "fail"},
		{"must-not-clean", answerMustNot(r, []string{`(?i)implementé`}), "pass"},
		{"must-not-hit", answerMustNot(r, []string{`(?i)propongo`}), "fail"},
		{"must-not-ignores-earlier-messages", answerMustNot(r, []string{`Primero`}), "pass"},
		{"must-without-answer", answerMust(result{}, []string{`x`}), "fail"},
		{"must-not-without-answer", answerMustNot(result{}, []string{`x`}), "not_observed"},
	} {
		if tc.got.Status != tc.want {
			t.Fatalf("%s: got %s, want %s", tc.name, tc.got.Status, tc.want)
		}
	}
	final := result{Trace: traceReport{Events: []traceEvent{{Line: 2, Kind: "final", Text: "Listo, todo bien."}}}}
	if got := answerMust(final, []string{`todo bien`}); got.Status != "pass" {
		t.Fatalf("final event not used when there is no assistant text: %s", got.Status)
	}
}

// --- per-task evaluation ---

func TestQualifySuiteEvaluateTask(t *testing.T) {
	yes := true
	cleanInv := qualifyInventory("README.md", "a", ".git/HEAD", "h", ".git/refs/heads/main", "r0")
	task := qualifySmallSuite().Tasks[0]
	ok := result{Before: cleanInv, After: cleanInv}
	if out := evaluateQualifyTask(task, ok, task.Prompt); !out.Passed {
		t.Fatalf("clean run fails: %v", out.Failures)
	}
	bad := result{Before: cleanInv, After: qualifyInventory("README.md", "b", ".git/HEAD", "h", ".git/refs/heads/main", "r0")}
	if out := evaluateQualifyTask(task, bad, task.Prompt); out.Passed || len(out.Failures) == 0 {
		t.Fatal("edited fixture passes")
	}

	skillTask := qualifySmallSuite().Tasks[1]
	skillTask.Checks.AnswerMustNot = nil
	noSkill := result{}
	if out := evaluateQualifyTask(skillTask, noSkill, "p"); out.Passed {
		t.Fatal("task expecting a skill passes with no read")
	}
	// A criterion that is not_observed passes only when the task lists it.
	bareURL := qualifyTask{ID: "u", Group: "message-format", Split: "tuning", Checks: qualifyChecks{Criteria: map[string][]string{"no_bare_url": {"pass"}}}}
	if out := evaluateQualifyTask(bareURL, result{}, "p"); out.Passed {
		t.Fatal("unlisted not_observed passes")
	}
	bareURL.Checks.Criteria["no_bare_url"] = []string{"pass", "not_observed"}
	if out := evaluateQualifyTask(bareURL, result{}, "p"); !out.Passed {
		t.Fatalf("listed not_observed fails: %v", out.Failures)
	}
	_ = yes
}

// A Codex agent_message has an empty Message key: the final answer is the last
// text event alone, not every assistant message of the session.
func TestQualifySuiteAnswerChecksCodexEmptyMessageKey(t *testing.T) {
	r := result{Host: "codex", Trace: traceReport{Events: []traceEvent{
		assistantText(1, "", "I will implement the change now."),
		assistantText(2, "", "I only propose a plan."),
	}}}
	if got := answerMustNot(r, []string{`(?i)implement`}); got.Status != "pass" {
		t.Fatalf("an early message counted in answer_must_not: %s", got.Status)
	}
	if got := answerMust(r, []string{`(?i)implement`}); got.Status != "fail" {
		t.Fatalf("an early message satisfied answer_must: %s", got.Status)
	}
	if got := answerMust(r, []string{`(?i)plan`}); got.Status != "pass" {
		t.Fatalf("the last message is not the answer: %s", got.Status)
	}
}

// Only Codex lacks a Message key by design: another host's unkeyed text parts
// still join into one answer.
func TestQualifySuiteAnswerJoinsUnkeyedPartsOutsideCodex(t *testing.T) {
	r := result{Host: "opencode", Trace: traceReport{Events: []traceEvent{
		assistantText(1, "", "First part of the answer."),
		assistantText(2, "", "Second part of the answer."),
	}}}
	if got := answerMust(r, []string{`First part`, `Second part`}); got.Status != "pass" {
		t.Fatalf("unkeyed parts were not joined: %s %v", got.Status, got.Evidence)
	}
}
