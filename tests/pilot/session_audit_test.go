package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func sessionFixture(parts ...string) string {
	return filepath.Join(append([]string{"..", "fixtures", "sessions"}, parts...)...)
}

func loadFixture(t *testing.T, host string, opts auditOptions, parts ...string) auditSession {
	t.Helper()
	s, err := loadAuditSession(sessionFixture(parts...), host, opts)
	if err != nil {
		t.Fatalf("load %v: %v", parts, err)
	}
	return s
}

func eventsOf(s auditSession, kind string) []traceEvent {
	var out []traceEvent
	for _, e := range s.Trace.Events {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

func turnOf(t *testing.T, s auditSession, i int) int {
	t.Helper()
	if len(s.Turn) != len(s.Trace.Events) {
		t.Fatalf("Turn has %d entries for %d events", len(s.Turn), len(s.Trace.Events))
	}
	return s.Turn[i]
}

func assistantTexts(s auditSession) []traceEvent {
	var out []traceEvent
	for _, e := range s.Trace.Events {
		if e.Kind == "text" && e.Role == "assistant" {
			out = append(out, e)
		}
	}
	return out
}

func TestAuditClaudeAdapter(t *testing.T) {
	s := loadFixture(t, "claude", auditOptions{}, "adapters", "claude-project", "session.jsonl")
	if s.Cwd != "project" || s.Model != "claude-example-1" || s.ModelConfigured {
		t.Fatalf("cwd/model = %q/%q configured=%v", s.Cwd, s.Model, s.ModelConfigured)
	}
	if s.First != "2026-01-01T10:00:01.000Z" || s.Last != "2026-01-01T10:00:10.000Z" {
		t.Fatalf("range = %s .. %s", s.First, s.Last)
	}
	shells := eventsOf(s, "shell")
	if len(shells) != 1 || shells[0].Command != "git status --short && echo sentinel-command-text" || shells[0].Success == nil || !*shells[0].Success {
		t.Fatalf("shell events = %+v", shells)
	}
	if len(eventsOf(s, "question")) != 1 {
		t.Fatalf("question events = %+v", eventsOf(s, "question"))
	}
	texts := assistantTexts(s)
	if len(texts) != 3 {
		t.Fatalf("assistant texts = %d", len(texts))
	}
	var shellIdx, lastText int
	for i, e := range s.Trace.Events {
		if e.Kind == "shell" {
			shellIdx = i
		}
		if e.Kind == "text" && e.Text == "Done." {
			lastText = i
		}
	}
	// Turns: the meta user line is not a turn, the second real prompt starts turn 2.
	if turnOf(t, s, shellIdx) != 1 || turnOf(t, s, lastText) != 2 {
		t.Fatalf("turns = %d, %d", turnOf(t, s, shellIdx), turnOf(t, s, lastText))
	}
}

func TestAuditCodexAdapter(t *testing.T) {
	s := loadFixture(t, "codex", auditOptions{}, "adapters", "codex-rollout", "rollout.jsonl")
	if s.Cwd != "project" || s.Model != "gpt-example" || !s.ModelConfigured || s.Subagent {
		t.Fatalf("cwd/model/configured/subagent = %q/%q/%v/%v", s.Cwd, s.Model, s.ModelConfigured, s.Subagent)
	}
	if s.First != "2026-01-01T11:00:01.000Z" || s.Last == "" {
		t.Fatalf("range = %s .. %s", s.First, s.Last)
	}
	shells := eventsOf(s, "shell")
	if len(shells) != 4 {
		t.Fatalf("shell events = %+v", shells)
	}
	wantCommands := []string{
		`git status --short && echo "sentinel-double"`,
		`echo sentinel-single`,
		"git add notes.txt && echo \"sentinel-backtick\"",
		"",
	}
	wantSuccess := []bool{true, false, true, true}
	for i, e := range shells {
		if e.Command != wantCommands[i] {
			t.Fatalf("shell %d command = %q, want %q", i, e.Command, wantCommands[i])
		}
		if e.Success == nil || *e.Success != wantSuccess[i] {
			t.Fatalf("shell %d success = %v, want %v", i, e.Success, wantSuccess[i])
		}
		if e.Line == 0 {
			t.Fatalf("shell %d has no source line", i)
		}
	}
	if !shells[3].CommandWithheld {
		t.Fatalf("interpolated command must be retained without text: %+v", shells[3])
	}
	writes := eventsOf(s, "write")
	deletes := eventsOf(s, "delete")
	if len(writes) != 2 || writes[0].Path != "src/a.txt" || writes[1].Path != "docs/b.md" || len(deletes) != 1 || deletes[0].Path != "old.txt" {
		t.Fatalf("patch events = %+v %+v", writes, deletes)
	}
	var mem *traceEvent
	for i, e := range s.Trace.Events {
		if e.Kind == "tool" && e.Tool == "mcp__engram__mem_search" {
			mem = &s.Trace.Events[i]
		}
		if e.Tool == "write_stdin" {
			t.Fatalf("write_stdin must be ignored: %+v", e)
		}
	}
	if mem == nil || !mem.Memory {
		t.Fatalf("engram tool event missing or not marked as memory: %+v", mem)
	}
	if len(eventsOf(s, "question")) != 1 {
		t.Fatalf("question events = %+v", eventsOf(s, "question"))
	}
	texts := assistantTexts(s)
	if len(texts) != 2 || texts[0].Text != "Round answered." || texts[1].Text != "Second turn." {
		t.Fatalf("texts = %+v", texts)
	}
	var first, second int
	for i, e := range s.Trace.Events {
		if e.Text == "Round answered." {
			first = i
		}
		if e.Text == "Second turn." {
			second = i
		}
	}
	if turnOf(t, s, first) != 1 || turnOf(t, s, second) != 2 {
		t.Fatalf("turns = %d, %d", turnOf(t, s, first), turnOf(t, s, second))
	}
}

func TestAuditCodexSubagentFlag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollout.jsonl")
	line := `{"timestamp":"2026-01-01T00:00:00.000Z","type":"session_meta","payload":{"id":"x","cwd":"project","source":{"subagent":{"thread_spawn":{"depth":1}}}}}` + "\n"
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loadAuditSession(path, "codex", auditOptions{})
	if err != nil || !s.Subagent {
		t.Fatalf("subagent = %v err=%v", s.Subagent, err)
	}
}

func TestAuditGrokAdapter(t *testing.T) {
	s := loadFixture(t, "grok", auditOptions{}, "adapters", "grok-updates", "updates.jsonl")
	if s.Model != "grok-example-1" {
		t.Fatalf("model = %q", s.Model)
	}
	wantFirst := time.Unix(1794948001, 0).UTC().Format("2006-01-02T15:04:05.000Z")
	if s.First != wantFirst {
		t.Fatalf("first = %s, want %s", s.First, wantFirst)
	}
	shells := eventsOf(s, "shell")
	if len(shells) != 2 || shells[0].Command != "git status --short && echo sentinel-command-text" || shells[0].Success == nil || !*shells[0].Success || shells[1].Success == nil || *shells[1].Success {
		t.Fatalf("shell events = %+v", shells)
	}
	if len(eventsOf(s, "question")) != 1 {
		t.Fatalf("question events = %+v", eventsOf(s, "question"))
	}
	var mem bool
	for _, e := range s.Trace.Events {
		if e.Kind == "tool" && e.Tool == "engram__mem_save" && e.Memory {
			mem = true
		}
	}
	if !mem {
		t.Fatalf("engram use_tool not recognized: %+v", s.Trace.Events)
	}
	texts := assistantTexts(s)
	if len(texts) != 2 || texts[0].Text != "Hello there, this is a message split in chunks." || texts[1].Text != "Second." {
		t.Fatalf("texts = %+v", texts)
	}
	var first, second int
	for i, e := range s.Trace.Events {
		if e.Text == "Hello there, this is a message split in chunks." {
			first = i
		}
		if e.Text == "Second." {
			second = i
		}
	}
	if turnOf(t, s, first) != 1 || turnOf(t, s, second) != 2 {
		t.Fatalf("turns = %d, %d", turnOf(t, s, first), turnOf(t, s, second))
	}
}

func TestAuditGrokCwdFromFolderName(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "%2Fwork%2Fexample%2Fproject", "session-1")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "updates.jsonl")
	if err := os.WriteFile(path, []byte(`{"method":"session/update","params":{"update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"x"}}},"timestamp":1794948001}`+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := loadAuditSession(path, "grok", auditOptions{})
	if err != nil || s.Cwd != "/work/example/project" {
		t.Fatalf("cwd = %q err=%v", s.Cwd, err)
	}
}

func TestAuditReportCarriesNoCommandOrMessageText(t *testing.T) {
	for _, tc := range []struct {
		host  string
		parts []string
	}{
		{"claude", []string{"adapters", "claude-project", "session.jsonl"}},
		{"codex", []string{"adapters", "codex-rollout", "rollout.jsonl"}},
		{"grok", []string{"adapters", "grok-updates", "updates.jsonl"}},
	} {
		var out bytes.Buffer
		if err := auditSessionMain(sessionFixture(tc.parts...), tc.host, auditOptions{Base: "development"}, &out); err != nil {
			t.Fatalf("%s: %v", tc.host, err)
		}
		var rep auditReport
		if err := json.Unmarshal(out.Bytes(), &rep); err != nil {
			t.Fatalf("%s: invalid JSON: %v", tc.host, err)
		}
		if rep.Host != tc.host || len(rep.Criteria) == 0 || rep.FirstTimestamp == "" || rep.LastTimestamp == "" {
			t.Fatalf("%s: report = %+v", tc.host, rep)
		}
		for _, secret := range []string{"sentinel", "example output", "Hello there", "Round answered", "Checking.", "Please check"} {
			if strings.Contains(out.String(), secret) {
				t.Fatalf("%s: report leaks %q: %s", tc.host, secret, out.String())
			}
		}
	}
}

func TestAuditRejectsUnsupportedHost(t *testing.T) {
	var out bytes.Buffer
	if err := auditSessionMain(sessionFixture("adapters", "claude-project", "session.jsonl"), "pi", auditOptions{}, &out); err == nil {
		t.Fatal("host pi must be rejected")
	}
}

func auditStatus(t *testing.T, rep auditReport, name string) criterionAssessment {
	t.Helper()
	for _, c := range rep.Criteria {
		if c.Criterion == name {
			return c
		}
	}
	t.Fatalf("criterion %s missing from %+v", name, rep.Criteria)
	return criterionAssessment{}
}

func auditFixture(t *testing.T, host string, opts auditOptions, parts ...string) auditReport {
	t.Helper()
	rep, err := auditSessionFile(sessionFixture(parts...), host, opts)
	if err != nil {
		t.Fatalf("audit %v: %v", parts, err)
	}
	return rep
}

func TestAuditRunsTheSevenExistingCriteria(t *testing.T) {
	rep := auditFixture(t, "claude", auditOptions{}, "adapters", "claude-project", "session.jsonl")
	got := map[string]bool{}
	for _, c := range rep.Criteria {
		got[c.Criterion] = true
	}
	for _, name := range []string{"question_after_detail", "no_broad_git_add", "no_secret_content_read", "cited_id_glossed", "no_bare_url", "ticket_ids_not_packed_in_prose", "merged_branch_deleted"} {
		if !got[name] {
			t.Fatalf("criterion %s missing: %+v", name, rep.Criteria)
		}
	}
}

func TestAuditMergedBranchDeletedCountsCleanupScriptApply(t *testing.T) {
	pass := auditStatus(t, auditFixture(t, "claude", auditOptions{}, "existing_criteria", "claude-project-merged-cleanup-pass.jsonl"), "merged_branch_deleted")
	if pass.Status != "pass" || len(pass.Evidence) == 0 || !strings.HasPrefix(pass.Evidence[0], "claude-project-merged-cleanup-pass.jsonl:") {
		t.Fatalf("pass = %+v", pass)
	}
	fail := auditStatus(t, auditFixture(t, "claude", auditOptions{}, "existing_criteria", "claude-project-merged-cleanup-fail.jsonl"), "merged_branch_deleted")
	if fail.Status != "fail" {
		t.Fatalf("a dry-run cleanup must not count as deletion: %+v", fail)
	}
}

func TestAuditMergedBranchDeletedWithoutCloseIsReview(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, "existing_criteria", "claude-project-merged-no-close-review.jsonl"), "merged_branch_deleted")
	if got.Status != "review" || len(got.Evidence) != 2 || !strings.Contains(got.Evidence[1], "no close report after the last merge") {
		t.Fatalf("a merge with no close after it must be review: %+v", got)
	}
}

func TestAuditLocalMergeNeedsBase(t *testing.T) {
	without := auditStatus(t, auditFixture(t, "claude", auditOptions{Cwd: t.TempDir()}, "existing_criteria", "claude-project-local-merge.jsonl"), "merged_branch_deleted")
	if without.Status != "not_observed" || len(without.Evidence) != 1 || !strings.Contains(without.Evidence[0], "base branch") {
		t.Fatalf("without base = %+v", without)
	}
	with := auditStatus(t, auditFixture(t, "claude", auditOptions{Base: "development", Cwd: t.TempDir()}, "existing_criteria", "claude-project-local-merge.jsonl"), "merged_branch_deleted")
	if with.Status != "fail" {
		t.Fatalf("with base = %+v", with)
	}
}

func TestAuditResolvesBaseFromFlagThenAgents(t *testing.T) {
	dir := t.TempDir()
	if got := resolveAuditBase(dir, ""); got != "" {
		t.Fatalf("no AGENTS.md: %q", got)
	}
	agents := "# Repo\n\n## Hive\n\n- Project: example\n- Base branch: development\n- Tracker: none\n\n## Other\n\n- Base branch: wrong\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(agents), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := resolveAuditBase(dir, ""); got != "development" {
		t.Fatalf("AGENTS.md: %q", got)
	}
	if got := resolveAuditBase(dir, "main"); got != "main" {
		t.Fatalf("flag: %q", got)
	}
}

func TestAuditGitAddResolvesAgainstSessionCwd(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "docs"), 0o700); err != nil {
		t.Fatal(err)
	}
	opts := auditOptions{Cwd: dir}
	// A directory pathspec that exists on disk cannot be checked against the
	// historical inventory: review, never fail.
	if s := auditStatus(t, auditFixture(t, "claude", opts, "existing_criteria", "claude-project-git-add-dir.jsonl"), "no_broad_git_add").Status; s != "review" {
		t.Fatalf("directory pathspec = %s, want review", s)
	}
	// A pathspec that no longer exists cannot be judged either.
	if s := auditStatus(t, auditFixture(t, "claude", auditOptions{Cwd: t.TempDir()}, "existing_criteria", "claude-project-git-add-dir.jsonl"), "no_broad_git_add").Status; s != "review" {
		t.Fatalf("pathspec missing on disk = %s, want review", s)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if s := auditStatus(t, auditFixture(t, "claude", opts, "existing_criteria", "claude-project-git-add-file.jsonl"), "no_broad_git_add").Status; s != "pass" {
		t.Fatalf("file pathspec = %s, want pass", s)
	}
	// A pathspec outside the session root cannot be judged.
	if s := auditStatus(t, auditFixture(t, "claude", opts, "existing_criteria", "claude-project-git-add-outside.jsonl"), "no_broad_git_add").Status; s != "review" {
		t.Fatalf("pathspec outside the root = %s, want review", s)
	}
	if s := auditStatus(t, auditFixture(t, "claude", opts, "existing_criteria", "claude-project-git-add-all.jsonl"), "no_broad_git_add").Status; s != "fail" {
		t.Fatalf("git add -A = %s, want fail", s)
	}
}

type newCriterionCase struct {
	name      string
	host      string
	criterion string
	parts     []string
	opts      auditOptions
	want      string
}

func hiveLikeDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	skill := filepath.Join(dir, "content", "skills", "flow-close")
	if err := os.MkdirAll(skill, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skill, "SKILL.md"), []byte("# flow-close\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "nested", "worktree")
}

func TestAuditNewCriteria(t *testing.T) {
	const round, close, skill = "text_round_without_card", "close_before_branch_cleanup", "skill_named_as_reason"
	const review = "review_question_native_option"
	const leaf = "leaf_read_before_action"
	const route, format, ux = "flow_entered_without_route", "decision_round_format", "ui_change_without_ux_review"
	const cache, tryIt = "tracker_states_cache_written", "try_it_block_in_report"
	const prose = "prose_choice_without_card"
	const asyncCard = "async_card_when_waiting_offered"
	const assess = "ticket_assessment_ending"
	const feed = "self_feeding_log"
	const reply = "analysis_reply_block"
	const wait = "unbounded_background_wait"
	const cards = "card_option_format"
	const folder = "change_folder_uncommitted"
	const planRoute = "plan_without_build_route"
	const suite, rounds = "full_suite_after_text_fix", "review_rounds_per_candidate"
	cases := []newCriterionCase{
		{"async card while the waiting card is offered", "codex", asyncCard, []string{asyncCard, "codex-rollout-waiting-offered-fail.jsonl"}, auditOptions{}, "fail"},
		{"async card while the waiting card is offered, round written first", "codex", asyncCard, []string{asyncCard, "codex-rollout-waiting-offered-round-written-fail.jsonl"}, auditOptions{}, "fail"},
		{"async card as the only tool, no round in the text", "codex", asyncCard, []string{asyncCard, "codex-rollout-async-only-no-round-fail.jsonl"}, auditOptions{}, "fail"},
		{"async card as the only tool, round written first", "codex", asyncCard, []string{asyncCard, "codex-rollout-async-only-round-pass.jsonl"}, auditOptions{}, "pass"},
		{"async card in a child session", "codex", asyncCard, []string{asyncCard, "codex-rollout-child-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"async card, session without an async call", "codex", asyncCard, []string{prose, "codex-rollout-card-pass.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice codex pending delivery with a recommended mode in backticks", "codex", prose, []string{prose, "codex-rollout-pending-recommend-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex delivery only left pending", "codex", prose, []string{prose, "codex-rollout-pending-only-review.jsonl"}, auditOptions{}, "review"},
		{"prose choice codex recommended backticks with no delivery mode word", "codex", prose, []string{prose, "codex-rollout-mode-substring-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"plan codex typed skill, build command without a next-step block", "codex", planRoute, []string{planRoute, "codex-rollout-typed-skill-no-block-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex typed skill, build command without a next-step block, session ends in the stage", "codex", planRoute, []string{planRoute, "codex-rollout-typed-skill-no-block-open-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex typed skill, build command with the next-step block", "codex", planRoute, []string{planRoute, "codex-rollout-typed-skill-block-pass.jsonl"}, auditOptions{}, "pass"},
		{"plan codex stage ended by a build without the command", "codex", planRoute, []string{planRoute, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex next-step block naming /flow-build", "codex", planRoute, []string{planRoute, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"plan codex session ends inside the stage", "codex", planRoute, []string{planRoute, "codex-rollout-open-review.jsonl"}, auditOptions{}, "review"},
		{"plan codex re-read of flow-plan keeps the stage open", "codex", planRoute, []string{planRoute, "codex-rollout-reread-pass.jsonl"}, auditOptions{}, "pass"},
		{"plan codex message naming only /flow-plan", "codex", planRoute, []string{planRoute, "codex-rollout-plan-command-only-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex child session", "codex", planRoute, []string{planRoute, "codex-rollout-child-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"plan codex build in the plan's own turn", "codex", planRoute, []string{planRoute, "codex-rollout-same-turn-review.jsonl"}, auditOptions{}, "review"},
		{"plan codex build command in bold", "codex", planRoute, []string{planRoute, "codex-rollout-bold-command-pass.jsonl"}, auditOptions{}, "pass"},
		{"plan codex grep of flow-build while planning", "codex", planRoute, []string{planRoute, "codex-rollout-grep-build-pass.jsonl"}, auditOptions{}, "pass"},
		{"plan codex card answered in the plan's turn, then the build", "codex", planRoute, []string{planRoute, "codex-rollout-card-then-build-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex build loaded by a sed range chained with rg", "codex", planRoute, []string{planRoute, "codex-rollout-load-with-rg-fail.jsonl"}, auditOptions{}, "fail"},
		{"plan codex grep of the repository's unattended skill", "codex", planRoute, []string{planRoute, "codex-rollout-unattended-grep-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex later tweak without verify keeps the close report", "codex", tryIt, []string{tryIt, "codex-rollout-later-tweak-pass.jsonl"}, auditOptions{}, "pass"},
		{"review codex fixed Review header offering only the agent", "codex", review, []string{review, "codex-rollout-header-agent-only-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex numbered title under a heading marker", "codex", round, []string{round, "codex-rollout-numbered-heading-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex bare Review heading above a numbered round", "codex", round, []string{round, "codex-rollout-review-heading-numbered-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose codex bare Review heading above a numbered round", "codex", prose, []string{round, "codex-rollout-review-heading-numbered-fail.jsonl"}, auditOptions{}, "not_observed"},
		{"format codex D1 round whose Situation lists numbered questions", "codex", format, []string{format, "codex-rollout-numbered-situation-pass.jsonl"}, auditOptions{}, "pass"},
		{"prose choice codex question after a bare Review header", "codex", prose, []string{prose, "codex-rollout-header-then-question-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex later screen change after a closed one", "codex", tryIt, []string{tryIt, "codex-rollout-later-screen-fail.jsonl"}, auditOptions{}, "fail"},
		{"review codex security question with one mechanism", "codex", review, []string{review, "codex-rollout-security-one-mechanism-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex title with a middle dot", "codex", round, []string{round, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex title with a middle dot", "codex", format, []string{format, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex numbered questions with starred options", "codex", round, []string{round, "codex-rollout-numbered-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex numbered questions without Situation", "codex", format, []string{format, "codex-rollout-numbered-fail.jsonl"}, auditOptions{}, "fail"},
		{"review codex numbered review question with one option", "codex", review, []string{review, "codex-rollout-numbered-one-option-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex analysis listing numbered questions", "codex", round, []string{round, "codex-rollout-numbered-analysis-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"format codex analysis listing numbered questions", "codex", format, []string{round, "codex-rollout-numbered-analysis-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review codex analysis listing numbered questions", "codex", review, []string{round, "codex-rollout-numbered-analysis-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice codex numbered bold question without letters", "codex", prose, []string{prose, "codex-rollout-numbered-prose-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex imperative confirmation of delivery", "codex", prose, []string{prose, "codex-rollout-imperative-confirm-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex replies offered in quotation marks", "codex", prose, []string{prose, "codex-rollout-quoted-replies-review.jsonl"}, auditOptions{}, "review"},
		{"prose choice codex sentence that reports a confirmation", "codex", prose, []string{prose, "codex-rollout-reported-confirm-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice codex past-tense Confirmé is no request", "codex", prose, []string{prose, "codex-rollout-past-tense-confirm-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice codex without the card", "codex", prose, []string{prose, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex asked through the card", "codex", prose, []string{prose, "codex-rollout-card-pass.jsonl"}, auditOptions{}, "pass"},
		{"prose choice wrapped over two lines", "codex", prose, []string{prose, "codex-rollout-wrapped-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice with an unclosed ¿ across paragraphs", "codex", prose, []string{prose, "codex-rollout-unclosed-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice card about something else", "codex", prose, []string{prose, "codex-rollout-other-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice findings card headed Revisión", "codex", prose, []string{prose, "codex-rollout-findings-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice english after the context line", "codex", prose, []string{prose, "codex-rollout-english-context-line-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice yes-no review question", "codex", prose, []string{prose, "codex-rollout-yes-no-review-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice ordinary word ready", "codex", prose, []string{prose, "codex-rollout-ready-word-review.jsonl"}, auditOptions{}, "review"},
		{"prose choice before a next-step block", "codex", prose, []string{prose, "codex-rollout-next-step-block-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice in a full-format text round", "codex", prose, []string{prose, "codex-rollout-full-round-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice after another question", "codex", prose, []string{prose, "codex-rollout-after-question-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose session word without a choice", "codex", prose, []string{prose, "codex-rollout-session-word-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice naming /code-review", "codex", prose, []string{prose, "codex-rollout-code-review-slash-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose delivery choice after a lettered block", "codex", prose, []string{prose, "codex-rollout-round-plus-delivery-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex plain yes-no question", "codex", prose, []string{prose, "codex-rollout-plain-question-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"prose choice codex about what the user wants", "codex", prose, []string{prose, "codex-rollout-intent-review.jsonl"}, auditOptions{}, "review"},
		{"prose choice codex card not offered", "codex", prose, []string{prose, "codex-rollout-not-offered-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"cache codex search empty, move, no write", "codex", cache, []string{cache, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"cache codex written after the move", "codex", cache, []string{cache, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"cache codex search found the topic", "codex", cache, []string{cache, "codex-rollout-found-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"cache claude search empty, move, no write", "claude", cache, []string{cache, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"cache claude written through the engram CLI", "claude", cache, []string{cache, "claude-project-cli-save-pass.jsonl"}, auditOptions{}, "pass"},
		{"cache claude creating a ticket is not a move", "claude", cache, []string{cache, "claude-project-create-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"cache claude reading transitions is not a move", "claude", cache, []string{cache, "claude-project-read-transitions-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"try it codex final report without the block", "codex", tryIt, []string{tryIt, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex final report with the block", "codex", tryIt, []string{tryIt, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"try it codex close report judged before later work", "codex", tryIt, []string{tryIt, "codex-rollout-close-then-other-work-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex frontend child, report without the block", "codex", tryIt, []string{tryIt, "codex-rollout-child-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex mobile child, report without the block", "codex", tryIt, []string{tryIt, "codex-rollout-mobile-child-fail.jsonl"}, auditOptions{}, "fail"},
		{"try it codex mobile child, report with the block", "codex", tryIt, []string{tryIt, "codex-rollout-mobile-child-pass.jsonl"}, auditOptions{}, "pass"},
		{"try it codex UI edit without verify", "codex", tryIt, []string{tryIt, "codex-rollout-noverify-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"try it claude final report with the block", "claude", tryIt, []string{tryIt, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"try it claude final report without the block", "claude", tryIt, []string{tryIt, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"route codex reply to a proposal starts build", "codex", route, []string{route, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"route codex reply to a next-step block", "codex", route, []string{route, "codex-rollout-block-pass.jsonl"}, auditOptions{}, "pass"},
		{"route codex typed command after a proposal", "codex", route, []string{route, "codex-rollout-typed-pass.jsonl"}, auditOptions{}, "pass"},
		{"route codex direct request in the first turn", "codex", route, []string{route, "codex-rollout-direct-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"route claude ok to a proposal starts build", "claude", route, []string{route, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment claude card answered, then a bare ending", "claude", assess, []string{assess, "claude-project-card-then-bare-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment claude typed /hive-mode command", "claude", assess, []string{assess, "claude-project-typed-command-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment claude prompt only mentions /hive-mode", "claude", assess, []string{assess, "claude-project-mentioned-command-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"format codex decisions as bullets", "codex", format, []string{format, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex fixed headers without options", "codex", format, []string{format, "codex-rollout-fixedheader-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex full round in text", "codex", format, []string{format, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"format codex card carries the round", "codex", format, []string{format, "codex-rollout-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"format codex report with a review heading", "codex", format, []string{format, "codex-rollout-report-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"ux codex screen edit without the reviewer", "codex", ux, []string{ux, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"ux codex screen edit with the reviewer", "codex", ux, []string{ux, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"ux codex test and logic edits only", "codex", ux, []string{ux, "codex-rollout-test-only-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"ux claude screen edit with the reviewer", "claude", ux, []string{ux, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"route a read of the repository's own flow skill", "codex", route, []string{route, "codex-rollout-repo-read-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"format a recap of decisions in force", "codex", format, []string{format, "codex-rollout-recap-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"route codex flow loaded after a card's answer", "codex", route, []string{route, "codex-rollout-card-fail.jsonl"}, auditOptions{}, "fail"},
		{"route codex new request after a link with a query", "codex", route, []string{route, "codex-rollout-new-request-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"route codex longer reply to a proposal", "codex", route, []string{route, "codex-rollout-long-reply-review.jsonl"}, auditOptions{}, "review"},
		{"route claude typed slash command", "claude", route, []string{route, "claude-project-typed-pass.jsonl"}, auditOptions{}, "pass"},
		{"format a recap marked recommended", "codex", format, []string{format, "codex-rollout-recap-recommended-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"ux a brief that only mentions the UX reviewer", "codex", ux, []string{ux, "codex-rollout-brief-mention-fail.jsonl"}, auditOptions{}, "fail"},
		{"ux a Claude child transcript is not judged", "claude", ux, []string{ux, "claude-subagent-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"route codex async card answered in the next turn", "codex", route, []string{route, "codex-rollout-async-card-fail.jsonl"}, auditOptions{}, "fail"},
		{"route a file test is not a skill load", "codex", route, []string{route, "codex-rollout-file-test-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"format a lettered round without the star", "codex", format, []string{format, "codex-rollout-missing-star-fail.jsonl"}, auditOptions{}, "fail"},
		{"format one of two decisions without its format", "codex", format, []string{format, "codex-rollout-one-of-two-fail.jsonl"}, auditOptions{}, "fail"},
		{"ux a rejected edit changes nothing", "claude", ux, []string{ux, "claude-project-rejected-edit-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"ux an edit in a sibling project", "claude", ux, []string{ux, "claude-project-sibling-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"ux a screen rewritten from the shell", "codex", ux, []string{ux, "codex-rollout-shell-edit-fail.jsonl"}, auditOptions{}, "fail"},
		{"ux a page written to a temporary folder", "claude", ux, []string{ux, "claude-project-tmp-html-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"leaf claude reads before git, ticket and card", "claude", leaf, []string{leaf, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude git add before the read", "claude", leaf, []string{leaf, "claude-project-git-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf claude ticket move without the read", "claude", leaf, []string{leaf, "claude-project-ticket-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf claude card without the read", "claude", leaf, []string{leaf, "claude-project-question-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf claude no governed action", "claude", leaf, []string{leaf, "claude-project-notexercised.jsonl"}, auditOptions{}, "not_observed"},
		{"leaf codex reads before git and card", "codex", leaf, []string{leaf, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf codex reads the single questions leaf", "codex", leaf, []string{leaf, "codex-rollout-questions-leaf-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf codex git add without the read", "codex", leaf, []string{leaf, "codex-rollout-git-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf codex card without the read", "codex", leaf, []string{leaf, "codex-rollout-question-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf claude reads the hive-core path before git, ticket and card", "claude", leaf, []string{leaf, "claude-project-core-path-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude nested reference and hive-core SKILL.md are no leaf read", "claude", leaf, []string{leaf, "claude-project-core-nested-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf codex shell read of the hive-core path before git and card", "codex", leaf, []string{leaf, "codex-rollout-core-path-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude reads the domain path before git, ticket and card", "claude", leaf, []string{leaf, "claude-project-domain-path-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude nested reference under a domain is no leaf read", "claude", leaf, []string{leaf, "claude-project-domain-nested-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf codex shell read of the domain path before git and card", "codex", leaf, []string{leaf, "codex-rollout-domain-path-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf codex no governed action", "codex", leaf, []string{leaf, "codex-rollout-notexercised.jsonl"}, auditOptions{}, "not_observed"},
		{"leaf claude Read of a packet with line numbers and arrows counts its members", "claude", leaf, []string{leaf, "claude-project-packet-read-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude Read of a packet with line numbers and tabs counts its members", "claude", leaf, []string{leaf, "claude-project-packet-read-tab-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude packet read with no line numbers counts its members", "claude", leaf, []string{leaf, "claude-project-packet-unnumbered-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude packet that holds a leaf's own Change folder section", "claude", leaf, []string{leaf, "claude-project-packet-own-section-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf claude complete packet without the governing leaf", "claude", leaf, []string{leaf, "claude-project-packet-without-leaf-fail.jsonl"}, auditOptions{}, "fail"},
		{"leaf claude packet cut at its end", "claude", leaf, []string{leaf, "claude-project-packet-cut-end-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"leaf claude packet cut in the middle with its end line intact", "claude", leaf, []string{leaf, "claude-project-packet-cut-middle-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"leaf codex shell read of a packet counts its members", "codex", leaf, []string{leaf, "codex-rollout-packet-shell-pass.jsonl"}, auditOptions{}, "pass"},
		{"leaf codex packet cut in the middle with its end line intact", "codex", leaf, []string{leaf, "codex-rollout-packet-cut-middle-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round claude fail", "claude", round, []string{round, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"round claude pass", "claude", round, []string{round, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"round claude no card in session", "claude", round, []string{round, "claude-project-nocard-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round claude title without option star", "claude", round, []string{round, "claude-project-notround-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex fail", "codex", round, []string{round, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex pass", "codex", round, []string{round, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"round codex empty card before round", "codex", round, []string{round, "codex-rollout-empty-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex subagent", "codex", round, []string{round, "codex-rollout-subagent-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex card listed in additional_tools", "codex", round, []string{round, "codex-rollout-offered-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex card listed outside Default or Plan", "codex", round, []string{round, "codex-rollout-offered-execute-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex written as a table", "codex", round, []string{round, "codex-rollout-table-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex session question under a fixed header", "codex", round, []string{round, "codex-rollout-fixed-header-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex fixed header in another case", "codex", round, []string{round, "codex-rollout-header-case-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex review report heading with reply letters", "codex", round, []string{round, "codex-rollout-review-report-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review question under the fixed header", "codex", review, []string{review, "codex-rollout-fixed-header-pass.jsonl"}, auditOptions{}, "pass"},
		{"round codex findings table and a cited answer", "codex", round, []string{round, "codex-rollout-table-notround-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review question without the native review", "codex", review, []string{review, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"review native named only outside the options", "codex", review, []string{review, "codex-rollout-situation-fail.jsonl"}, auditOptions{}, "fail"},
		{"review table offering the native review", "codex", review, []string{review, "codex-rollout-table-pass.jsonl"}, auditOptions{}, "pass"},
		{"review card naming the mechanisms only in descriptions", "codex", review, []string{review, "codex-rollout-card-description-fail.jsonl"}, auditOptions{}, "fail"},
		{"review table naming the native review in a situation cell", "codex", review, []string{review, "codex-rollout-table-situation-fail.jsonl"}, auditOptions{}, "fail"},
		{"review table with a separate ID column", "codex", review, []string{review, "codex-rollout-table-idcolumn-pass.jsonl"}, auditOptions{}, "pass"},
		{"review findings card mentioning the agent", "codex", review, []string{review, "codex-rollout-findings-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review question titled with a bare review", "codex", review, []string{review, "codex-rollout-bare-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex table with A) options split by <br>", "codex", round, []string{round, "codex-rollout-table-parens-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex table with A. and B: options", "codex", round, []string{round, "codex-rollout-table-punctuation-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex table with bold **A.** options", "codex", round, []string{round, "codex-rollout-table-bold-letters-fail.jsonl"}, auditOptions{}, "fail"},
		{"review findings decision naming two mechanisms", "codex", review, []string{review, "codex-rollout-findings-mechanisms-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review question with a plural title", "codex", review, []string{review, "codex-rollout-plural-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"review card with the native review", "codex", review, []string{review, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"close report mentioning a review", "codex", review, []string{review, "codex-rollout-close-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"decision about applying review findings", "codex", review, []string{review, "codex-rollout-findings-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round grok fail", "grok", round, []string{round, "grok-updates-fail.jsonl"}, auditOptions{}, "fail"},
		{"round grok pass", "grok", round, []string{round, "grok-updates-pass.jsonl"}, auditOptions{}, "pass"},
		{"round grok no user available", "grok", round, []string{round, "grok-updates-empty-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round grok declined card is not an exemption", "grok", round, []string{round, "grok-updates-declined-card-fail.jsonl"}, auditOptions{}, "fail"},
		{"close claude fail", "claude", close, []string{close, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"close claude pass", "claude", close, []string{close, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"close claude no deletion", "claude", close, []string{close, "claude-project-nodelete-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close claude merge deletes the branch itself", "claude", close, []string{close, "claude-project-delete-branch-flag-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close grok fail", "grok", close, []string{close, "grok-updates-fail.jsonl"}, auditOptions{}, "fail"},
		{"close grok pass", "grok", close, []string{close, "grok-updates-pass.jsonl"}, auditOptions{}, "pass"},
		{"close report keeps a branch the session deleted", "claude", close, []string{close, "claude-project-kept-review.jsonl"}, auditOptions{}, "review"},
		{"close report says a folder stays on master without the kept form", "claude", close, []string{close, "claude-project-stays-on-master-fail.jsonl"}, auditOptions{}, "fail"},
		{"close report keeps a branch the session never deleted", "claude", close, []string{close, "claude-project-kept-other-branch-fail.jsonl"}, auditOptions{}, "fail"},
		{"close deletion by the helper through a variable is a deletion", "claude", close, []string{close, "claude-project-variable-apply-pass.jsonl"}, auditOptions{}, "pass"},
		{"close codex merge with unknown success is not a merge", "codex", close, []string{close, "codex-rollout-truncated-merge-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close earlier merge is checked too", "claude", close, []string{close, "claude-project-two-merges-fail.jsonl"}, auditOptions{}, "fail"},
		{"close local merge and push of base", "claude", close, []string{close, "claude-project-local-merge-fail.jsonl"}, auditOptions{Base: "development"}, "fail"},
		{"close local merge without a base is not recognized", "claude", close, []string{close, "claude-project-local-merge-fail.jsonl"}, auditOptions{Cwd: "elsewhere"}, "not_observed"},
		{"round codex typed timeout is not an empty card", "codex", round, []string{round, "codex-rollout-typed-timeout-fail.jsonl"}, auditOptions{}, "fail"},
		{"round claude typed timeout is not an empty card", "claude", round, []string{round, "claude-project-typed-timeout-fail.jsonl"}, auditOptions{}, "fail"},
		{"round grok typed text is not an empty card", "grok", round, []string{round, "grok-updates-typed-timeout-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex empty answers", "codex", round, []string{round, "codex-rollout-empty-answers-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round claude tool error", "claude", round, []string{round, "claude-project-tool-error-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round claude rejected card is not empty", "claude", round, []string{round, "claude-project-rejected-card-fail.jsonl"}, auditOptions{}, "fail"},
		{"round after unattended delegation (skill call)", "claude", round, []string{round, "claude-project-unattended-after-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round before unattended delegation still fails", "claude", round, []string{round, "claude-project-unattended-before-fail.jsonl"}, auditOptions{}, "fail"},
		{"round after unattended delegation (SKILL.md read)", "codex", round, []string{round, "codex-rollout-unattended-read-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round with a compact summary inside its turn", "claude", round, []string{round, "claude-project-compact-summary-pass.jsonl"}, auditOptions{}, "pass"},
		{"round with an interruption notice inside its turn", "claude", round, []string{round, "claude-project-interrupted-pass.jsonl"}, auditOptions{}, "pass"},
		{"round grok UserDeclined is a decline", "grok", round, []string{round, "grok-updates-userdeclined-fail.jsonl"}, auditOptions{}, "fail"},
		{"round grok failed status is a decline", "grok", round, []string{round, "grok-updates-failed-status-fail.jsonl"}, auditOptions{}, "fail"},
		{"skill claude review", "claude", skill, []string{skill, "claude-project-review.jsonl"}, auditOptions{Cwd: "elsewhere"}, "review"},
		{"skill claude pass", "claude", skill, []string{skill, "claude-project-pass.jsonl"}, auditOptions{Cwd: "elsewhere"}, "pass"},
		{"skill codex spanish verb", "codex", skill, []string{skill, "codex-rollout-review.jsonl"}, auditOptions{Cwd: "elsewhere"}, "review"},
		{"assessment codex ends without a card, block or command", "codex", assess, []string{assess, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment codex ends with the next-step block", "codex", assess, []string{assess, "codex-rollout-block-pass.jsonl"}, auditOptions{}, "pass"},
		{"assessment codex ends with a card", "codex", assess, []string{assess, "codex-rollout-card-pass.jsonl"}, auditOptions{}, "pass"},
		{"assessment codex ends with an open question in text", "codex", assess, []string{assess, "codex-rollout-open-question-review.jsonl"}, auditOptions{}, "review"},
		{"assessment codex ticket without hive-mode", "codex", assess, []string{assess, "codex-rollout-no-hive-mode-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"assessment codex ticket-status playbook", "codex", assess, []string{assess, "codex-rollout-ticket-status-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"assessment codex standard prefix is not a ticket", "codex", assess, []string{assess, "codex-rollout-standard-prefix-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"assessment codex github issue number", "codex", assess, []string{assess, "codex-rollout-github-issue-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment claude skill call, no ending", "claude", assess, []string{assess, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"assessment claude ends with a card", "claude", assess, []string{assess, "claude-project-card-pass.jsonl"}, auditOptions{}, "pass"},
		{"assessment claude ends with the answers block", "claude", assess, []string{assess, "claude-project-answers-block-pass.jsonl"}, auditOptions{}, "pass"},
		{"assessment claude turn that follows a flow command", "claude", assess, []string{assess, "claude-project-flow-command-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed codex interpolated block with a trap that tails its log", "codex", feed, []string{feed, "codex-rollout-interpolated-trap-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed codex reader inside the redirected block", "codex", feed, []string{feed, "codex-rollout-reader-in-block-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed claude exec redirect with a trap that cats the file", "claude", feed, []string{feed, "claude-project-exec-trap-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed claude simple command redirect with a trap", "claude", feed, []string{feed, "claude-project-simple-trap-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude block redirect then a tail after an or", "claude", feed, []string{feed, "claude-project-or-tail-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude output in one command, tail in another", "claude", feed, []string{feed, "claude-project-two-commands-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude run_check helper", "claude", feed, []string{feed, "claude-project-run-check-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed codex the same interpolation is the same file", "codex", feed, []string{feed, "codex-rollout-interpolated-same-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed codex distinct interpolations are different files", "codex", feed, []string{feed, "codex-rollout-interpolated-distinct-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude stderr-only exec redirect", "claude", feed, []string{feed, "claude-project-stderr-exec-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude exec saves fd 3 then redirects stdout", "claude", feed, []string{feed, "claude-project-fd3-save-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed claude reader before the stdout restore", "claude", feed, []string{feed, "claude-project-reader-before-restore-fail.jsonl"}, auditOptions{}, "fail"},
		{"feed claude reader after the stdout restore", "claude", feed, []string{feed, "claude-project-reader-after-restore-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"feed claude pattern inside a python heredoc", "claude", feed, []string{feed, "claude-project-python-heredoc-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude italic bullets without letters", "claude", reply, []string{reply, "claude-project-italic-bullets-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply claude italic bullets after reading the rule", "claude", reply, []string{reply, "claude-project-italic-read-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply claude heading with the arrow and letters", "claude", reply, []string{reply, "claude-project-lettered-block-pass.jsonl"}, auditOptions{}, "pass"},
		{"reply claude heading with its D ID and letters without list markers", "claude", reply, []string{reply, "claude-project-id-heading-pass.jsonl"}, auditOptions{}, "pass"},
		{"format claude a reply block with its D ID is not a decision round", "claude", format, []string{reply, "claude-project-id-heading-pass.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude bold question heading and letters", "claude", reply, []string{reply, "claude-project-bold-question-pass.jsonl"}, auditOptions{}, "pass"},
		{"reply claude letters without a heading", "claude", reply, []string{reply, "claude-project-letters-no-heading-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply claude report ending in bold-labeled findings", "claude", reply, []string{reply, "claude-project-bold-findings-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude decision round written in text", "claude", reply, []string{reply, "claude-project-text-round-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude ending with flow commands", "claude", reply, []string{reply, "claude-project-flow-command-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude an earlier turn ended badly", "claude", reply, []string{reply, "claude-project-earlier-turn-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply claude sidechain transcript is not judged", "claude", reply, []string{reply, "claude-project-sidechain-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply codex italic bullets", "codex", reply, []string{reply, "codex-rollout-italic-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply codex lettered block", "codex", reply, []string{reply, "codex-rollout-lettered-pass.jsonl"}, auditOptions{}, "pass"},
		{"reply codex report ending with no replies", "codex", reply, []string{reply, "codex-rollout-no-replies-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply codex scratch write in /tmp keeps it an analysis", "codex", reply, []string{reply, "codex-rollout-temp-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply codex short answer is not judged", "codex", reply, []string{reply, "codex-rollout-short-answer-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply codex turn that patched a file is not judged", "codex", reply, []string{reply, "codex-rollout-edit-turn-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude turn that wrote a file is not judged", "claude", reply, []string{reply, "claude-project-write-turn-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply claude refused write keeps it an analysis", "claude", reply, []string{reply, "claude-project-denied-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply codex bare report in a turn without hive-mode is not judged", "codex", reply, []string{reply, "codex-rollout-no-hive-mode-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"reply grok italic bullets", "grok", reply, []string{reply, "grok-updates-italic-fail.jsonl"}, auditOptions{}, "fail"},
		{"reply grok lettered block", "grok", reply, []string{reply, "grok-updates-lettered-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a background loop with sleep and no deadline", "claude", wait, []string{"unbounded_background_wait", "claude-project-unbounded-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude timeout wrapping the loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-timeout-wrap-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude timeout only inside the loop body", "claude", wait, []string{"unbounded_background_wait", "claude-project-timeout-body-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a for loop over a finite list", "claude", wait, []string{"unbounded_background_wait", "claude-project-seq-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a counter compared to a limit that exits", "claude", wait, []string{"unbounded_background_wait", "claude-project-counter-exit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude SECONDS only printed", "claude", wait, []string{"unbounded_background_wait", "claude-project-seconds-print-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude SECONDS compared in the until condition", "claude", wait, []string{"unbounded_background_wait", "claude-project-seconds-limit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a foreground loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-foreground-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait claude run_in_background false", "claude", wait, []string{"unbounded_background_wait", "claude-project-background-false-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait claude a background command without a loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-no-loop-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait grok a loop with block_until_ms 0 and no deadline", "grok", wait, []string{"unbounded_background_wait", "grok-updates-unbounded-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait grok timeout wrapping the loop", "grok", wait, []string{"unbounded_background_wait", "grok-updates-timeout-wrap-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait grok a counter compared to a limit that exits", "grok", wait, []string{"unbounded_background_wait", "grok-updates-counter-exit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait grok block_until_ms above 0", "grok", wait, []string{"unbounded_background_wait", "grok-updates-blocking-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait grok no block_until_ms", "grok", wait, []string{"unbounded_background_wait", "grok-updates-no-block-field-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait codex the host does not mark a background command", "codex", wait, []string{"unbounded_background_wait", "codex-rollout-unbounded-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait claude a loop after then", "claude", wait, []string{"unbounded_background_wait", "claude-project-then-while-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a loop in a subshell", "claude", wait, []string{"unbounded_background_wait", "claude-project-subshell-while-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a C-style for with a counter limit", "claude", wait, []string{"unbounded_background_wait", "claude-project-c-style-limit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a C-style for with a constant condition", "claude", wait, []string{"unbounded_background_wait", "claude-project-c-style-always-true-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a C-style for without a condition", "claude", wait, []string{"unbounded_background_wait", "claude-project-c-style-empty-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a counter limit in the while condition", "claude", wait, []string{"unbounded_background_wait", "claude-project-while-condition-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a SECONDS test inside an if in the body", "claude", wait, []string{"unbounded_background_wait", "claude-project-seconds-if-in-body-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a SECONDS test followed by a brace group", "claude", wait, []string{"unbounded_background_wait", "claude-project-seconds-brace-exit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a decrementing counter in the condition", "claude", wait, []string{"unbounded_background_wait", "claude-project-decrement-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude the limit on the left of the comparison", "claude", wait, []string{"unbounded_background_wait", "claude-project-limit-first-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude SECONDS compared with an or-exit", "claude", wait, []string{"unbounded_background_wait", "claude-project-seconds-or-exit-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a C-style for with spaces", "claude", wait, []string{"unbounded_background_wait", "claude-project-c-style-spaced-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude timeout after then", "claude", wait, []string{"unbounded_background_wait", "claude-project-then-timeout-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a counter inside a double-bracket and", "claude", wait, []string{"unbounded_background_wait", "claude-project-double-bracket-and-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude /bin/sleep in a negated condition loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-absolute-sleep-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude an outer loop whose only sleep is in a timeout-wrapped inner loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-outer-nested-timeout-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a loop after a background command", "claude", wait, []string{"unbounded_background_wait", "claude-project-background-before-loop-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude an unbounded outer loop around a bounded inner loop", "claude", wait, []string{"unbounded_background_wait", "claude-project-nested-inner-bounded-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a counter compared with -eq", "claude", wait, []string{"unbounded_background_wait", "claude-project-counter-eq-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a counter compared with == in arithmetic", "claude", wait, []string{"unbounded_background_wait", "claude-project-counter-arith-eq-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a loop keyword in a heredoc line", "claude", wait, []string{"unbounded_background_wait", "claude-project-heredoc-for-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a variable named MAX_SECONDS is not the clock", "claude", wait, []string{"unbounded_background_wait", "claude-project-max-seconds-name-fail.jsonl"}, auditOptions{}, "fail"},
		{"wait claude a C-style for with do on the same line", "claude", wait, []string{"unbounded_background_wait", "claude-project-c-style-do-glued-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a sleep inside an echoed sentence", "claude", wait, []string{"unbounded_background_wait", "claude-project-echoed-sleep-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait claude a loop inside an echoed sentence", "claude", wait, []string{"unbounded_background_wait", "claude-project-echoed-loop-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"wait claude timeout after env", "claude", wait, []string{"unbounded_background_wait", "claude-project-env-timeout-pass.jsonl"}, auditOptions{}, "pass"},
		{"wait claude a sleep in a subshell", "claude", wait, []string{"unbounded_background_wait", "claude-project-subshell-sleep-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude options that lost their letters", "claude", cards, []string{cards, "claude-project-trailing-star-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude cards in format, session questions included", "claude", cards, []string{cards, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"card options claude multiple choice with letters and several marks", "claude", cards, []string{cards, "claude-project-multiselect-pass.jsonl"}, auditOptions{}, "pass"},
		{"card options claude the mark on the second option", "claude", cards, []string{cards, "claude-project-second-star-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a single choice with no mark", "claude", cards, []string{cards, "claude-project-no-mark-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a star quoted inside another option", "claude", cards, []string{cards, "claude-project-quoted-star-pass.jsonl"}, auditOptions{}, "pass"},
		{"card options claude a rejected card without letters", "claude", cards, []string{cards, "claude-project-rejected-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a card with no result without letters", "claude", cards, []string{cards, "claude-project-no-result-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a Spanish marker no host asks for", "claude", cards, []string{cards, "claude-project-spanish-marker-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a child session", "claude", cards, []string{cards, "claude-project-sidechain-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"card options claude no card", "claude", cards, []string{cards, "claude-project-nocard-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"card options codex a middle dot after the letter", "codex", cards, []string{cards, "codex-rollout-middle-dot-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options codex the host's recommended marker", "codex", cards, []string{cards, "codex-rollout-host-marker-pass.jsonl"}, auditOptions{}, "pass"},
		{"card options codex an async card is not judged", "codex", cards, []string{cards, "codex-rollout-async-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"card options grok a card in format", "grok", cards, []string{cards, "grok-updates-pass.jsonl"}, auditOptions{}, "pass"},
		{"card options claude a card the host refused for invalid input", "claude", cards, []string{cards, "claude-project-invalid-input-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"card options claude two options with the same letter", "claude", cards, []string{cards, "claude-project-repeated-letter-fail.jsonl"}, auditOptions{}, "fail"},
		{"card options claude a card closed by another error was shown", "claude", cards, []string{cards, "claude-project-timeout-error-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder codex patched and never committed", "codex", folder, []string{folder, "codex-rollout-patch-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder codex written by Path.write_text in a heredoc", "codex", folder, []string{folder, "codex-rollout-write-text-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder codex versioned exited 1", "codex", folder, []string{folder, "codex-rollout-versioned-failed-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder codex commit of other paths beside a tasks.py record", "codex", folder, []string{folder, "codex-rollout-commit-of-other-paths-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder codex committed naming the folder", "codex", folder, []string{folder, "codex-rollout-redirect-commit-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder codex versioned exited 0", "codex", folder, []string{folder, "codex-rollout-versioned-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder codex no text after the merge", "codex", folder, []string{folder, "codex-rollout-no-text-review.jsonl"}, auditOptions{}, "review"},
		{"change folder codex only a checkpoint after the merge", "codex", folder, []string{folder, "codex-rollout-checkpoint-only-review.jsonl"}, auditOptions{}, "review"},
		{"change folder codex checkpoint written with --out after the merge", "codex", folder, []string{folder, "codex-rollout-checkpoint-out-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder codex folder only read", "codex", folder, []string{folder, "codex-rollout-read-only-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder codex no merge", "codex", folder, []string{folder, "codex-rollout-no-merge-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder codex archived folder", "codex", folder, []string{folder, "codex-rollout-archive-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder codex merge exited non-zero", "codex", folder, []string{folder, "codex-rollout-failed-merge-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder claude written and merged with --delete-branch", "claude", folder, []string{folder, "claude-project-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder claude committed after add", "claude", folder, []string{folder, "claude-project-commit-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder claude folder only read", "claude", folder, []string{folder, "claude-project-read-only-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"change folder claude commit paths built from a shell variable", "claude", folder, []string{folder, "claude-project-variable-commit-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder claude variable names other paths", "claude", folder, []string{folder, "claude-project-variable-other-paths-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder claude renamed then committed under the new name", "claude", folder, []string{folder, "claude-project-rename-commit-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder claude renamed after its commit and not committed again", "claude", folder, []string{folder, "claude-project-rename-uncommitted-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder claude staged in one call and committed in another", "claude", folder, []string{folder, "claude-project-staged-in-another-call-pass.jsonl"}, auditOptions{}, "pass"},
		{"change folder claude git add -A may have taken the folder", "claude", folder, []string{folder, "claude-project-add-all-review.jsonl"}, auditOptions{}, "review"},
		{"change folder claude edited after the commit and before the merge", "claude", folder, []string{folder, "claude-project-edit-after-commit-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder claude add in the specs repository and commit in the code repository", "claude", folder, []string{folder, "claude-project-add-in-specs-commit-in-code-fail.jsonl"}, auditOptions{}, "fail"},
		{"change folder claude code commit first and specs commit last", "claude", folder, []string{folder, "claude-project-code-commit-then-specs-commit-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude a markdown edit between two plain runs", "claude", suite, []string{suite, "claude-project-text-edit-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude two passing runs with nothing between them", "claude", suite, []string{suite, "claude-project-no-change-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude git add and commit are not changes", "claude", suite, []string{suite, "claude-project-commit-between-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude piped runs with a PIPESTATUS echo are not shown to pass", "claude", suite, []string{suite, "claude-project-pipestatus-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude pnpm test runs shown passing by exit=0", "claude", suite, []string{suite, "claude-project-pnpm-exit-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a shell write to a markdown file", "claude", suite, []string{suite, "claude-project-shell-md-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a Python write of a markdown file named in the script", "claude", suite, []string{suite, "claude-project-python-literal-md-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a scratch write in the temporary folder", "claude", suite, []string{suite, "claude-project-temp-write-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude cd and go -C name the same directory", "claude", suite, []string{suite, "claude-project-cd-run-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a code edit beside the markdown edit", "claude", suite, []string{suite, "claude-project-code-edit-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude a shell edit of a Go file", "claude", suite, []string{suite, "claude-project-shell-sed-go-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude a rebase between the runs", "claude", suite, []string{suite, "claude-project-rebase-between-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude piped runs without a pass line", "claude", suite, []string{suite, "claude-project-pipe-unproven-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude the first run failed", "claude", suite, []string{suite, "claude-project-first-failed-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a shell write with an unresolved target", "claude", suite, []string{suite, "claude-project-shell-unknown-target-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a first run launched in the background", "claude", suite, []string{suite, "claude-project-background-run-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude an exit line printed by another command", "claude", suite, []string{suite, "claude-project-other-command-status-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a known code write by a script after the run in the same call", "claude", suite, []string{suite, "claude-project-python-after-run-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude an unknown script write after the run in the same call", "claude", suite, []string{suite, "claude-project-python-variable-after-run-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude printf of the status counts like echo", "claude", suite, []string{suite, "claude-project-printf-status-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a printf whose line never holds the status", "claude", suite, []string{suite, "claude-project-printf-no-status-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a printf whose status argument no conversion prints", "claude", suite, []string{suite, "claude-project-printf-status-argument-unused-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a double-quoted printf format keeps its escapes", "claude", suite, []string{suite, "claude-project-printf-double-quoted-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a pipe inside a bash -c argument", "claude", suite, []string{suite, "claude-project-bash-c-pipe-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a text file under testdata is not text", "claude", suite, []string{suite, "claude-project-testdata-text-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude a Markdown file under tests/fixtures is not text", "claude", suite, []string{suite, "claude-project-fixtures-text-pass.jsonl"}, auditOptions{}, "pass"},
		{"full_suite_after_text_fix claude a first run with no tool result", "claude", suite, []string{suite, "claude-project-no-result-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude status echoes of commands run in the background", "claude", suite, []string{suite, "claude-project-background-group-status-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude two commands print the same status text", "claude", suite, []string{suite, "claude-project-same-status-text-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a PIPESTATUS echo that belongs to another pipeline", "claude", suite, []string{suite, "claude-project-other-pipeline-status-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude an exit line that only contains the run's status text", "claude", suite, []string{suite, "claude-project-contained-status-line-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a status line that ends with the run's status text", "claude", suite, []string{suite, "claude-project-status-suffix-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix codex a plain Codex run, a README patch and a rerun", "codex", suite, []string{suite, "codex-rollout-text-edit-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude the status of the run echoed right after it", "claude", suite, []string{suite, "claude-project-echo-status-fail.jsonl"}, auditOptions{}, "fail"},
		{"full_suite_after_text_fix claude a piped run is not shown by a PIPESTATUS echo beside another exit line", "claude", suite, []string{suite, "claude-project-piped-own-status-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude a Python write whose target is a variable", "claude", suite, []string{suite, "claude-project-python-write-review.jsonl"}, auditOptions{}, "review"},
		{"full_suite_after_text_fix claude one run", "claude", suite, []string{suite, "claude-project-one-run-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude runs filtered by -run", "claude", suite, []string{suite, "claude-project-filtered-run-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude runs of one sub-package", "claude", suite, []string{suite, "claude-project-subpackage-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude a push between the runs", "claude", suite, []string{suite, "claude-project-push-between-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude a merge between the runs", "claude", suite, []string{suite, "claude-project-merge-between-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude a new branch between the runs", "claude", suite, []string{suite, "claude-project-new-branch-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude runs in two directories", "claude", suite, []string{suite, "claude-project-other-directory-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"full_suite_after_text_fix claude suite text inside arguments", "claude", suite, []string{suite, "claude-project-echoed-suite-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review_rounds_per_candidate claude three code-review skill rounds", "claude", rounds, []string{rounds, "claude-project-three-skill-rounds-fail.jsonl"}, auditOptions{}, "fail"},
		{"review_rounds_per_candidate claude four review agent rounds", "claude", rounds, []string{rounds, "claude-project-four-agent-rounds-fail.jsonl"}, auditOptions{}, "fail"},
		{"review_rounds_per_candidate claude two rounds", "claude", rounds, []string{rounds, "claude-project-two-rounds-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude three agents in one message are one round", "claude", rounds, []string{rounds, "claude-project-one-message-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude a push resets the count", "claude", rounds, []string{rounds, "claude-project-push-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude a merge resets the count", "claude", rounds, []string{rounds, "claude-project-merge-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude a flow load resets the count", "claude", rounds, []string{rounds, "claude-project-flow-load-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude a branch switch resets the count", "claude", rounds, []string{rounds, "claude-project-branch-switch-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"rounds claude a typed /code-review and its skill launch are one round", "claude", rounds, []string{rounds, "claude-project-typed-and-skill-one-round-pass.jsonl"}, auditOptions{}, "pass"},
		{"rounds claude git checkout of a path is not a branch change", "claude", rounds, []string{rounds, "claude-project-checkout-paths-fail.jsonl"}, auditOptions{}, "fail"},
		{"rounds claude git restore of paths is not a branch change", "claude", rounds, []string{rounds, "claude-project-restore-paths-fail.jsonl"}, auditOptions{}, "fail"},
		{"rounds claude git checkout of a branch resets the count", "claude", rounds, []string{rounds, "claude-project-checkout-branch-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"rounds claude a branch change between the typed command and its launch", "claude", rounds, []string{rounds, "claude-project-typed-then-reset-fail.jsonl"}, auditOptions{}, "fail"},
		{"rounds claude git checkout of a branch with a dot resets the count", "claude", rounds, []string{rounds, "claude-project-checkout-dotted-branch-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"rounds claude git checkout of a branch followed by -- resets the count", "claude", rounds, []string{rounds, "claude-project-checkout-branch-then-dashes-resets-pass.jsonl"}, auditOptions{}, "pass"},
		{"review_rounds_per_candidate claude a typed /code-review before two more rounds", "claude", rounds, []string{rounds, "claude-project-typed-first-review.jsonl"}, auditOptions{}, "review"},
		{"review_rounds_per_candidate claude no review", "claude", rounds, []string{rounds, "claude-project-no-review-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"review_rounds_per_candidate claude a child session", "claude", rounds, []string{rounds, "claude-project-sidechain-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"rounds codex another host", "codex", rounds, []string{rounds, "codex-rollout-notobserved.jsonl"}, auditOptions{}, "not_observed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := auditFixture(t, tc.host, tc.opts, tc.parts...)
			got := auditStatus(t, rep, tc.criterion)
			if got.Status != tc.want {
				t.Fatalf("%s = %s, want %s (evidence %v)", tc.criterion, got.Status, tc.want, got.Evidence)
			}
			if (tc.want == "fail" || tc.want == "review") && len(got.Evidence) == 0 {
				t.Fatalf("%s %s has no evidence", tc.criterion, tc.want)
			}
			for _, e := range got.Evidence {
				for _, leak := range []string{"Which delivery", "requires your validation", "pide revisar", "Cierre", "Close ·", "scripts/cleanup"} {
					if strings.Contains(e, leak) {
						t.Fatalf("evidence leaks text %q: %s", leak, e)
					}
				}
			}
		})
	}
}

func TestAuditSkillNamedAsReasonSkipsTheHiveRepository(t *testing.T) {
	cwd := hiveLikeDir(t)
	got := auditStatus(t, auditFixture(t, "claude", auditOptions{Cwd: cwd}, "skill_named_as_reason", "claude-project-review.jsonl"), "skill_named_as_reason")
	if got.Status != "not_observed" {
		t.Fatalf("in the Hive repository naming a skill is legitimate: %+v", got)
	}
}

func TestAuditRoundEvidencePointsAtTheRoundMessage(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "codex", auditOptions{}, "text_round_without_card", "codex-rollout-fail.jsonl"), "text_round_without_card")
	if len(got.Evidence) != 1 || !strings.HasPrefix(got.Evidence[0], "codex-rollout-fail.jsonl:") {
		t.Fatalf("evidence = %v", got.Evidence)
	}
}

func TestAuditUnknownSuccessIsNotRecordedAsSuccess(t *testing.T) {
	for _, parts := range [][]string{
		{"close_before_branch_cleanup", "codex-rollout-truncated-merge-notobserved.jsonl"},
		{"existing_criteria", "codex-rollout-running-merge.jsonl"},
	} {
		s := loadFixture(t, "codex", auditOptions{}, parts...)
		var seen bool
		for _, e := range s.Trace.Events {
			if e.Kind == "shell" && strings.HasPrefix(e.Command, "gh pr merge") {
				seen = true
				if e.Success != nil {
					t.Fatalf("%v: success must be unknown, got %v", parts, *e.Success)
				}
			}
		}
		if !seen {
			t.Fatalf("%v: merge command not extracted", parts)
		}
		merged := auditStatus(t, auditFixture(t, "codex", auditOptions{}, parts...), "merged_branch_deleted")
		if merged.Status != "not_observed" {
			t.Fatalf("%v: merged_branch_deleted = %+v", parts, merged)
		}
	}
}

func TestScanCodexExecObjectKeys(t *testing.T) {
	cases := []struct {
		name, js, want string
		withheld       bool
	}{
		{"quoted key", `text(await tools.exec_command({"cmd": "git status", "yield_time_ms": 1000}));`, "git status", false},
		{"identifier value of another key named like the target", "const cmd = \"x\";\ntext(await tools.exec_command({workdir: cmd, cmd: \"git status\"}));", "git status", false},
		{"other key first", `text(await tools.exec_command({workdir: "sub", yield_time_ms: 5, cmd: 'git log'}));`, "git log", false},
		{"nested object value before the key", `text(await tools.exec_command({env: {cmd: "no"}, cmd: "git diff"}));`, "git diff", false},
		{"regex with a quote before the call", "const re = /[\"']x/g;\ntext(await tools.exec_command({cmd: \"git status\"}));", "git status", false},
		{"regex with a quote as a value", `text(await tools.exec_command({pattern: /"/, cmd: "git status"}));`, "git status", false},
		{"variable value is withheld", "const c = \"git status\";\ntext(await tools.exec_command({cmd: c}));", "", true},
		{"concatenation is withheld", `text(await tools.exec_command({cmd: "git " + "status"}));`, "", true},
		{"unterminated regex makes the block ambiguous", "const re = /\"abc\ntext(await tools.exec_command({cmd: \"git status\"}));", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls, ambiguous := scanCodexExec(tc.js)
			if len(calls) != 1 || calls[0].name != "exec_command" {
				t.Fatalf("calls = %+v", calls)
			}
			c := calls[0]
			literal := c.hasLiteral && !c.interpolate && !ambiguous
			if tc.withheld {
				if literal {
					t.Fatalf("command must be withheld, got %q", c.literal)
				}
				return
			}
			if !literal || c.literal != tc.want {
				t.Fatalf("literal = %q ok=%v ambiguous=%v, want %q", c.literal, c.hasLiteral, ambiguous, tc.want)
			}
		})
	}
}

func TestLeafReadBeforeActionEdgeCases(t *testing.T) {
	no := false
	sess := func(events ...traceEvent) auditSession { return auditSession{Trace: traceReport{Events: events}} }
	cases := []struct {
		name string
		s    auditSession
		want string
	}{
		{"helper commit without a read fails", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "python3 -I /x/git-workflow/scripts/git_state.py commit --repo /r -m msg a.go"},
		), "fail"},
		{"leaf named only in a commit message is no read", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "git commit -m \"see " + oldLeafPrefix + "git-commit/SKILL.md\" a.go"},
		), "fail"},
		{"a failed read is no read", sess(
			traceEvent{Line: 1, Kind: "read", ID: "r1", Path: "/s/" + oldLeafPrefix + "git-commit/SKILL.md", Success: &no},
			traceEvent{Line: 3, Kind: "shell", Command: "git add a.go"},
		), "fail"},
		{"creating a ticket with a state is not a move", sess(
			traceEvent{Line: 1, Kind: "tool", Tool: "mcp__linear__save_issue", Input: []byte(`{"title":"x","state":"Backlog"}`)},
		), "not_observed"},
		{"gh -R issue close is a move", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "gh -R owner/repo issue close 12"},
		), "fail"},
		{"helper commit with --repo before commit is seen", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "python3 -I /x/git_state.py --repo /r commit -m m a.go"},
		), "fail"},
		{"a heredoc naming the leaf is no read", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "cat > /tmp/m <<'EOF'\nsee " + oldLeafPrefix + "git-commit/SKILL.md\nEOF"},
			traceEvent{Line: 2, Kind: "shell", Command: "git commit -F /tmp/m a.go"},
		), "fail"},
		{"a GitLab-style update by iid is a move", sess(
			traceEvent{Line: 1, Kind: "tool", Tool: "update_issue", Input: []byte(`{"issue_iid":5,"state":"closed"}`)},
		), "fail"},
		{"cat of the leaf before the helper commit passes", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "cat /s/" + oldLeafPrefix + "git-commit/SKILL.md"},
			traceEvent{Line: 2, Kind: "shell", Command: "python3 -I /x/git_state.py commit -m m a.go"},
		), "pass"},
	}
	for _, tc := range cases {
		if got := leafReadBeforeAction(tc.s).Status; got != tc.want {
			t.Errorf("%s: status %s, want %s", tc.name, got, tc.want)
		}
	}
}

func TestAuditCardAnswers(t *testing.T) {
	cases := []struct {
		name, host, file    string
		answered, recommend int
	}{
		{"claude answered with an option", "claude", "claude-project-answered-pass.jsonl", 2, 1},
		{"claude user-has-answered form", "claude", "claude-project-user-has-answered-pass.jsonl", 2, 2},
		{"claude free text is not an option", "claude", "claude-project-free-text.jsonl", 1, 0},
		{"claude empty card", "claude", "claude-project-empty.jsonl", 0, 0},
		{"claude multiple choice is not counted", "claude", "claude-project-multiselect.jsonl", 0, 0},
		{"codex answered with Recommended and recomendado labels", "codex", "codex-rollout-answered-pass.jsonl", 3, 2},
		{"codex empty answers", "codex", "codex-rollout-empty.jsonl", 0, 0},
		{"codex free text and a missing answer", "codex", "codex-rollout-free-text.jsonl", 1, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := auditFixture(t, tc.host, auditOptions{}, "card_answers", tc.file)
			if rep.CardAnswers.Answered != tc.answered || rep.CardAnswers.Recommended != tc.recommend {
				t.Fatalf("card_answers = %+v, want answered %d recommended %d", rep.CardAnswers, tc.answered, tc.recommend)
			}
			for _, c := range rep.Criteria {
				if c.Criterion == "card_answers" || c.Criterion == "card_star_share" {
					t.Fatalf("card answers must stay out of the criteria list: %+v", c)
				}
			}
		})
	}
}

func TestAuditCardAnswersAppearInTheJSONReport(t *testing.T) {
	var out bytes.Buffer
	if err := auditSessionMain(sessionFixture("card_answers", "claude-project-answered-pass.jsonl"), "claude", auditOptions{}, &out); err != nil {
		t.Fatal(err)
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if got := string(v["card_answers"]); !strings.Contains(got, `"answered": 2`) || !strings.Contains(got, `"recommended": 1`) {
		t.Fatalf("card_answers = %s", got)
	}
}

func TestAuditMainThreadCost(t *testing.T) {
	grokRoot := sessionFixture("main_thread_cost", "grok-sessions")
	cases := []struct {
		name, host, file string
		want             mainThreadCost
	}{
		{"claude counts a message once and skips sidechain lines", "claude", "claude.jsonl", mainThreadCost{
			ModelCalls: 2, InputTokens: 480, CachedTokens: 350, PeakContext: 320, UsageSource: "per_call", Compactions: 1,
			ToolCalls: 4, HiveReads: 1, HiveScripts: 1, ByTool: map[string]int{"Read": 3, "Bash": 1}}},
		{"codex counts a token_count only when the total changed", "codex", "codex.jsonl", mainThreadCost{
			ModelCalls: 2, InputTokens: 350, CachedTokens: 140, PeakContext: 250, UsageSource: "per_call", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"exec": 4, "exec_command": 1}, RepeatedUsageEvents: 1}},
		{"grok subtracts a child found in another project folder", "grok", "grok-sessions/%2Frepo/parent-ok/updates.jsonl", mainThreadCost{
			ModelCalls: 5, InputTokens: 2700, CachedTokens: 800, UsageSource: "per_turn", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"read_file": 2, "run_terminal_command": 3}}},
		{"grok keeps the parent total when a child's usage is missing", "grok", "grok-sessions/%2Frepo/parent-missing/updates.jsonl", mainThreadCost{
			ModelCalls: 7, InputTokens: 3000, CachedTokens: 900, UsageSource: "per_turn_with_children", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"read_file": 2, "run_terminal_command": 3}}},
		{"grok without usage reports none", "grok", "grok-sessions/%2Frepo/parent-none/updates.jsonl", mainThreadCost{
			UsageSource: "none", ByTool: map[string]int{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := auditFixture(t, tc.host, auditOptions{GrokSessionsRoot: grokRoot}, "main_thread_cost", filepath.FromSlash(tc.file))
			got, want := rep.MainThreadCost, tc.want
			gj, _ := json.Marshal(got)
			wj, _ := json.Marshal(want)
			if string(gj) != string(wj) {
				t.Fatalf("main_thread_cost = %s\nwant %s", gj, wj)
			}
		})
	}
}

func TestAuditMainThreadCostLeavesTheCriteriaAlone(t *testing.T) {
	var out bytes.Buffer
	if err := auditSessionMain(sessionFixture("main_thread_cost", "claude.jsonl"), "claude", auditOptions{}, &out); err != nil {
		t.Fatal(err)
	}
	var v map[string]json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(v["main_thread_cost"]), `"model_calls": 2`) {
		t.Fatalf("main_thread_cost = %s", v["main_thread_cost"])
	}
	if strings.Contains(string(v["criteria"]), "main_thread_cost") {
		t.Fatalf("criteria mention main_thread_cost: %s", v["criteria"])
	}
}

func TestAnalysisReplyBlockEvidenceSaysIfTheRuleWasRead(t *testing.T) {
	const reply = "analysis_reply_block"
	for _, tc := range []struct{ file, want string }{
		{"claude-project-italic-bullets-fail.jsonl", "not read"},
		{"claude-project-italic-read-fail.jsonl", "was read"},
		{"codex-rollout-no-replies-fail.jsonl", "without a reply block; report-endings rule was read"},
		{"codex-rollout-temp-write-fail.jsonl", "without a reply block; report-endings rule not read"},
		{"codex-rollout-cell-packet-fail.jsonl", "report-endings rule was read"},
		{"codex-rollout-cell-truncated-fail.jsonl", "report-endings rule not read"},
		{"codex-rollout-cell-filtered-fail.jsonl", "report-endings rule not read"},
		{"codex-rollout-cell-unpaired-fail.jsonl", "report-endings rule not read"},
	} {
		host := strings.SplitN(tc.file, "-", 2)[0]
		got := auditStatus(t, auditFixture(t, host, auditOptions{}, reply, tc.file), reply)
		if got.Status != "fail" || len(got.Evidence) != 1 || !strings.Contains(got.Evidence[0], tc.want) || !strings.HasPrefix(got.Evidence[0], tc.file+":") {
			t.Fatalf("%s: %+v, want one fail evidence line at the file containing %q", tc.file, got, tc.want)
		}
	}
}

// TestReplyListShapes pins how the end of an analysis is read: where the list
// starts, which heading counts and which bullets are not replies.
func TestReplyListShapes(t *testing.T) {
	for _, tc := range []struct {
		name         string
		text         string
		bad, listed  bool
		exemptEnding bool
	}{
		{"italic replies before a waiting line", "Findings.\n\n- *fix it* — patch it\n- *leave it* — no change\n\nWaiting for your reply.", true, true, false},
		{"italic replies split by blank lines", "Findings.\n\n- *fix it* — patch it\n\n- *leave it* — no change", true, true, false},
		{"lettered options under a bold question", "Body.\n\n**▶ How do we continue?**\n- **A** (★) Close\n- **B** See the findings", false, true, false},
		{"bold letters without a heading", "Body.\n\n- **A** Close\n- **B** See the findings", true, true, false},
		{"a bold question far above does not head the list", "**Why does the loader read twice?**\n\nIt reads twice.\n\nSo:\n\nA — fix it\nB — leave it", true, true, false},
		{"hyphenated words are not letters", "Next:\n\n- E-mail the owner\n- X-ray the logs", false, false, false},
		{"paths in underscores are not italic labels", "Updated:\n\n- _support/docs/a_b.md updated\n- _support/openspec/x_y.md added", false, false, false},
		{"a flow command in the body does not exempt the ending", "This would go through /flow-plan later.\n\n- *fix it* — patch it\n- *leave it* — no change", true, true, false},
		{"a next-step block with the command between heading and replies", "Report.\n\n---\n**▶ Next step: plan (recommended)**\nDecisions remain.\n\n`/flow-plan the change`\n\nOther replies:\n- A - Record the cases.\n- B - Open a ticket.", false, true, true},
		{"a flow command in the ending exempts it", "Done.\n\n---\n**▶ Next step: plan**\n\n`/flow-plan the change`\n\n- **A** compact first\n- **B** change the plan", false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad, listed := replyListWithoutBlock(tc.text)
			if bad != tc.bad || listed != tc.listed {
				t.Fatalf("replyListWithoutBlock = bad %v listed %v, want bad %v listed %v", bad, listed, tc.bad, tc.listed)
			}
			if got := endingExempt(tc.text); got != tc.exemptEnding {
				t.Fatalf("endingExempt = %v, want %v", got, tc.exemptEnding)
			}
		})
	}
}

// TestShellWork pins which shell commands make a turn a work turn for the
// no-reply-block branch of analysis_reply_block.
func TestShellWork(t *testing.T) {
	for _, tc := range []struct {
		command string
		work    bool
	}{
		{"awk '$3 > 100 {print}' data.tsv", false},
		{`jq 'select(.n > 5)' a.json`, false},
		{`python3 -c "print(a > b)"`, false},
		{"rg -n foo src > /tmp/ex/hits.txt", false},
		{"ffprobe clip.mp4 > /private/var/folders/xx/T/probe.txt", false},
		{"cd /tmp/ex-frames && ffprobe clip.mp4 > probe.txt", false},
		{"rm /tmp/kin/a.jpg && rmdir /tmp/kin", false},
		{"cat a.go > src/b.go", true},
		{"sed -i 's/x/y/' src/a.go", true},
		{"sed -i.bak 's/x/y/' /tmp/a.txt", false},
		{"echo x | tee notes.md", true},
		{"cp a.go src/a.go", true},
		{"mv /tmp/a.go src/a.go", true},
		{`mkdir -p "$work/images"`, true},
		{"git apply fix.patch", true},
		{"git -c core.hooksPath=x commit -m x", true},
		{"git --no-pager push origin x", true},
		{"python3 -I scripts/git_state.py commit --repo . -m x -- a.go", true},
		{"gh pr create --fill", true},
		{"git log --oneline -3", false},
		{"bash -lc 'cd /repo && git commit -am fix && git push'", true},
		{`jq "select(.n>5)" a.json`, false},
		{`rg "->" src`, false},
		{`sed -i -e "s/x/y/" /tmp/f`, false},
		{"cp src/a.go /tmp/ex/a.go", false},
		{"truncate -s 0 /tmp/x.log", false},
		{"ln -s /repo/f /tmp/l", false},
		{"mkdir -m 700 /tmp/x", false},
		{"python3 - <<'PY'\nif a > b:\n    print(a)\nPY", false},
		{"cat <<EOF > /tmp/n\ndon't\nEOF\nsed -i 's/x/y/' src/a.go", true},
		{"go test ./... 2> build/errors.log", true},
		{"make &> out.log", true},
		{"go test ./... 2>&1 | tail -3", false},
		{"cd /tmp/ex && cd ~/repo && echo x > notes.md", true},
		{"find src -name '*.orig' -delete", true},
		{"find src -name '*.go' -exec cat {} +", false},
		{"env GIT_EDITOR=true git merge x", true},
		{"(git push)", true},
		{"sudo rm -rf build", true},
		{"nohup git push &", true},
		{"git stash list", false},
		{"git clean -fdx", true},
		{"git --git-dir /r/.git --work-tree /r commit -m x", true},
		{"git branch -D x", true},
		{"git branch --show-current", false},
		{"git tag v1", true},
		{"git tag -l", false},
		{"git worktree add ../x", true},
		{"echo 'unclosed", true},
	} {
		if got := shellWork(tc.command); got != tc.work {
			t.Errorf("shellWork(%q) = %v, want %v", tc.command, got, tc.work)
		}
	}
	for _, tc := range []struct {
		template string
		work     bool
	}{
		{"", true},
		{"__INTERP_0__ > __INTERP_1__", true},
		{"cat a.txt > __INTERP_0__/notes.md", true},
		{"rg -n updateMe __INTERP_0__/apps/backend/src | head -30", false},
	} {
		withheld := []traceEvent{{Kind: "shell", CommandWithheld: true, CommandTemplate: tc.template}}
		if got := workEvent(withheld, 0); got != tc.work {
			t.Errorf("workEvent(withheld, template %q) = %v, want %v", tc.template, got, tc.work)
		}
	}
}

func TestAuditUnboundedBackgroundWaitEvidencePointsAtTheLaunch(t *testing.T) {
	for host, want := range map[string]string{"claude": ".jsonl:2 shell", "grok": ".jsonl:2 shell"} {
		name := map[string]string{"claude": "claude-project-unbounded-fail.jsonl", "grok": "grok-updates-unbounded-fail.jsonl"}[host]
		got := auditStatus(t, auditFixture(t, host, auditOptions{}, "unbounded_background_wait", name), "unbounded_background_wait")
		if len(got.Evidence) != 1 || !strings.Contains(got.Evidence[0], want) {
			t.Fatalf("%s evidence = %v, want it to contain %q", host, got.Evidence, want)
		}
	}
}

func TestLeafPathsOldAndNew(t *testing.T) {
	old := "/h/.agents/skills/" + oldLeafPrefix + "report-endings/SKILL.md"
	cur := "/h/.agents/skills/hive-core/references/report-endings.md"
	nested := "/h/.agents/skills/hive-core/references/report-endings/extra.md"
	core := "/h/.agents/skills/hive-core/SKILL.md"
	domain := "/h/.agents/skills/hive-core/messages/report-endings.md"
	domainNested := "/h/.agents/skills/hive-core/messages/report-endings/extra.md"
	for _, tc := range []struct {
		path string
		want bool
	}{{old, true}, {cur, true}, {domain, true}, {nested, false}, {domainNested, false}, {core, false}} {
		if got := readsReportEndings(traceEvent{Kind: "read", Path: tc.path}); got != tc.want {
			t.Errorf("read %s: %v, want %v", tc.path, got, tc.want)
		}
		if got := readsReportEndings(traceEvent{Kind: "shell", Command: "sed -n 1,40p " + tc.path}); got != tc.want {
			t.Errorf("shell %s: %v, want %v", tc.path, got, tc.want)
		}
		if got := hiveCall("Read", tc.path, ""); (got == "read") != (tc.want || tc.path == nested || tc.path == domainNested || tc.path == core) {
			t.Errorf("cost read %s: %q", tc.path, got)
		}
	}
	if namesLeaf(traceEvent{Kind: "read", Path: "/s/hive-core/references/git-commit/x.md"}, leafGitCommit) {
		t.Error("a nested reference counted as the git-commit leaf")
	}
	if !namesLeaf(traceEvent{Kind: "read", Path: "/s/" + oldLeafPrefix + "question-channel/SKILL.md"}, leafQuestionRound) {
		t.Error("the former questions name was not accepted")
	}
}

func TestAuditMergedBranchDeletedThroughHelperVariable(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, "existing_criteria", "claude-project-merged-variable-apply-pass.jsonl"), "merged_branch_deleted")
	if got.Status != "pass" {
		t.Fatalf("the helper run with --apply through a variable is a deletion: %+v", got)
	}
}

func writeAgentsEnvironments(t *testing.T, environments string) string {
	t.Helper()
	dir := t.TempDir()
	body := "# Guidance\n\n## Hive\n\n- Project: example\n- Base branch: development\n- Environments: " + environments + "\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestAuditMergedBranchDeletedAfterTheSessionEnded(t *testing.T) {
	const crit = "merged_branch_deleted"
	const dir = "existing_criteria"
	for _, c := range []struct {
		name, host, file string
	}{
		{"claude", "claude", "claude-project-merged-end-fail.jsonl"},
		{"codex", "codex", "codex-rollout-merged-end-fail.jsonl"},
		{"grok", "grok", "grok-updates-merged-end-fail.jsonl"},
	} {
		got := auditStatus(t, auditFixture(t, c.host, auditOptions{}, dir, c.file), crit)
		if got.Status != "fail" {
			t.Errorf("%s: a merge left after the user ended the session must fail: %+v", c.name, got)
		}
	}
	question := auditStatus(t, auditFixture(t, "claude", auditOptions{}, dir, "claude-project-merged-question-review.jsonl"), crit)
	if question.Status != "review" {
		t.Errorf("a question that holds an end phrase does not end the session: %+v", question)
	}
}

func TestAuditMergedBranchDeletedPromotionAfterTheSessionEnded(t *testing.T) {
	const crit = "merged_branch_deleted"
	const dir = "existing_criteria"
	base := auditOptions{Base: "development"}
	if got := auditStatus(t, auditFixture(t, "claude", base, dir, "claude-project-promotion-head-review.jsonl"), crit); got.Status != "review" {
		t.Errorf("a promotion whose head is the base must be review: %+v", got)
	}
	if got := auditStatus(t, auditFixture(t, "claude", base, dir, "claude-project-promotion-headrefname-review.jsonl"), crit); got.Status != "review" {
		t.Errorf("a promotion whose head a headRefName read shows must be review: %+v", got)
	}
	if got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, dir, "claude-project-promotion-head-review.jsonl"), crit); got.Status != "fail" {
		t.Errorf("without a base the head development is a work branch: %+v", got)
	}
	qa := auditFixture(t, "claude", base, dir, "claude-project-promotion-qa-head-review.jsonl")
	if got := auditStatus(t, qa, crit); got.Status != "fail" {
		t.Errorf("a head outside the base and the Environments is a work merge: %+v", got)
	}
	withEnv := auditOptions{Base: "development", Cwd: writeAgentsEnvironments(t, "development \u2192 qa \u2192 master")}
	if got := auditStatus(t, auditFixture(t, "claude", withEnv, dir, "claude-project-promotion-qa-head-review.jsonl"), crit); got.Status != "review" {
		t.Errorf("a head listed in the Environments of the session AGENTS.md must be review: %+v", got)
	}
}

func TestAuditMergedBranchDeletedKeptBranchAfterTheSessionEnded(t *testing.T) {
	const crit = "merged_branch_deleted"
	const dir = "existing_criteria"
	if got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, dir, "claude-project-merged-kept-end-review.jsonl"), crit); got.Status != "review" {
		t.Errorf("kept for the branch of the merge, seen by --head, must be review: %+v", got)
	}
	if got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, dir, "claude-project-merged-kept-unseen-head-fail.jsonl"), crit); got.Status != "fail" {
		t.Errorf("a merge whose head the session does not show is a work merge: %+v", got)
	}
}

func TestAuditHelperApplyNeedsTheScriptToRun(t *testing.T) {
	const dir = "existing_criteria"
	for _, f := range []string{
		"claude-project-merged-unrelated-apply-fail.jsonl",
		"claude-project-merged-rg-apply-fail.jsonl",
		"claude-project-merged-other-variable-apply-fail.jsonl",
	} {
		if got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, dir, f), "merged_branch_deleted"); got.Status != "fail" {
			t.Errorf("%s: an --apply of another command is not a branch deletion: %+v", f, got)
		}
	}
}

func TestAuditEndPhraseNeedsOnlyClosingWords(t *testing.T) {
	for _, tc := range []struct {
		msg  string
		want bool
	}{
		{"es todo por el momento, gracias!", true},
		{"Es todo", true},
		{"that's it for now, thanks", true},
		{"por el momento es todo", true},
		{"Es todo correcto, sigue con el siguiente ticket", false},
		{"Terminamos el build y luego vemos el deploy", false},
	} {
		if got := isSessionEndMessage(tc.msg); got != tc.want {
			t.Errorf("isSessionEndMessage(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
	for _, f := range []string{"claude-project-merged-es-todo-correcto-review.jsonl", "claude-project-merged-terminamos-el-build-review.jsonl"} {
		if got := auditStatus(t, auditFixture(t, "claude", auditOptions{}, "existing_criteria", f), "merged_branch_deleted"); got.Status != "review" {
			t.Errorf("%s: a message that continues the work does not end the session: %+v", f, got)
		}
	}
}

func TestAuditMergeHeadIsThatOfTheMergedPullRequest(t *testing.T) {
	base := auditOptions{Base: "development"}
	for _, f := range []string{
		"claude-project-merged-create-without-head-fail.jsonl",
		"claude-project-merged-headref-other-pr-fail.jsonl",
	} {
		if got := auditStatus(t, auditFixture(t, "claude", base, "existing_criteria", f), "merged_branch_deleted"); got.Status != "fail" {
			t.Errorf("%s: the head of another pull request is not the merge's head: %+v", f, got)
		}
	}
}

func TestAuditCloseBeforeCleanupSkipsPromotionMerges(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "claude", auditOptions{Base: "development"}, "close_before_branch_cleanup", "claude-project-promotion-then-work-pass.jsonl"), "close_before_branch_cleanup")
	if got.Status != "pass" {
		t.Fatalf("a deploy report after a promotion merge is not a close before cleanup: %+v", got)
	}
}

func TestAuditEndPhraseAcceptsClosingVariants(t *testing.T) {
	for _, tc := range []struct {
		msg  string
		want bool
	}{
		{"Por hoy es todo, gracias", true},
		{"Por ahora es todo", true},
		{"Es todo, gracias por tu ayuda", true},
		{"Terminamos por hoy, nos vemos mañana", true},
		{"that's all for today", true},
		{"that's all, bye", true},
		{"Es todo correcto, sigue con el siguiente ticket", false},
		{"Terminamos el build y luego vemos el deploy", false},
	} {
		if got := isSessionEndMessage(tc.msg); got != tc.want {
			t.Errorf("isSessionEndMessage(%q) = %v, want %v", tc.msg, got, tc.want)
		}
	}
}

func TestAuditMergeHeadMatchesTheMergedPullRequest(t *testing.T) {
	base := auditOptions{Base: "development"}
	for _, tc := range []struct{ file, want string }{
		{"claude-project-merged-create-other-pr-fail.jsonl", "fail"},
		{"claude-project-merged-headref-current-branch-review.jsonl", "review"},
		{"claude-project-merged-headref-branch-positional-review.jsonl", "review"},
	} {
		if got := auditStatus(t, auditFixture(t, "claude", base, "existing_criteria", tc.file), "merged_branch_deleted"); got.Status != tc.want {
			t.Errorf("%s = %s, want %s: %+v", tc.file, got.Status, tc.want, got)
		}
	}
}

func TestAuditHelperInvocationSeesWrappers(t *testing.T) {
	for _, tc := range []struct {
		cmd  string
		want bool
	}{
		{"python3 -I scripts/cleanup_branches.py --repo . --apply", true},
		{"PYTHONDONTWRITEBYTECODE=1 python3 -I scripts/cleanup_branches.py --repo . --apply", true},
		{"env PYTHONDONTWRITEBYTECODE=1 python3 scripts/cleanup_branches.py --apply", true},
		{"timeout 600 python3 -I scripts/cleanup_branches.py --apply", true},
		{"uv run scripts/cleanup_branches.py --apply", true},
		{"uv run python scripts/cleanup_branches.py --apply", true},
		{"python3 -X utf8 scripts/cleanup_branches.py --apply", true},
		{"python3 -W ignore scripts/cleanup_branches.py --apply", true},
		{"rg -n -- --apply scripts/cleanup_branches.py", false},
		{"python3 scripts/apply_deltas.py --apply", false},
	} {
		got := false
		for _, args := range shellSegments(tc.cmd, 0) {
			got = got || isCleanupApply(args)
		}
		if got != tc.want {
			t.Errorf("isCleanupApply(%q) = %v, want %v", tc.cmd, got, tc.want)
		}
	}
}

func TestDeletedBranchNamesMatchesTheDeletionCommands(t *testing.T) {
	ok := true
	shell := func(cmd string) traceEvent { return traceEvent{Kind: "shell", Command: cmd, Success: &ok} }
	got := deletedBranchNames([]traceEvent{
		shell("gh api -X DELETE repos/o/r/git/refs/heads/feat/api-gone"),
		shell("git push origin feat/kept :feat/gone"),
		shell("git push --delete origin feat/flagged"),
	})
	want := map[string]bool{"feat/api-gone": true, "feat/gone": true, "feat/flagged": true}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("deletedBranchNames = %v, want %v", got, want)
	}
}

func TestMergeHeadReadsEverySegmentOfACodexCall(t *testing.T) {
	ok := true
	events := []traceEvent{
		{Kind: "shell", ID: "call_1#0", Command: "gh pr view 42 --json headRefName", Success: &ok},
		{Kind: "shell", ID: "call_1#1", Command: "gh pr checks 42", Success: &ok},
		{Kind: "tool_result", ID: "call_1", Text: `{"headRefName":"development"}`},
		{Kind: "shell", ID: "call_2", Command: "gh pr merge 42 --merge", Success: &ok},
	}
	if got := mergeHead(events, 3); got != "development" {
		t.Fatalf("mergeHead = %q, want development", got)
	}
}

func TestMergeHeadIgnoresAHeadRefNameAnotherSegmentPrinted(t *testing.T) {
	ok := true
	events := []traceEvent{
		{Kind: "shell", ID: "call_1#0", Command: "gh pr view 42 --json title", Success: &ok},
		{Kind: "shell", ID: "call_1#1", Command: "gh pr list --json headRefName,number", Success: &ok},
		{Kind: "tool_result", ID: "call_1", Text: `[{"headRefName":"development","number":41}]`},
		{Kind: "shell", ID: "call_2", Command: "gh pr merge 42 --merge", Success: &ok},
	}
	if got := mergeHead(events, 3); got != "" {
		t.Fatalf("mergeHead = %q, want empty", got)
	}
}

func TestAuditChangeFolderEvidenceNamesTheFolder(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "codex", auditOptions{}, "change_folder_uncommitted", "codex-rollout-patch-fail.jsonl"), "change_folder_uncommitted")
	if len(got.Evidence) != 1 || !strings.Contains(got.Evidence[0], "changes/demo-change") || !strings.HasPrefix(got.Evidence[0], "codex-rollout-patch-fail.jsonl:") {
		t.Fatalf("evidence = %v", got.Evidence)
	}
}

func TestAuditCodexCompletedItemCarriesExitCodeAndPatchPaths(t *testing.T) {
	s := loadFixture(t, "codex", auditOptions{}, "change_folder_uncommitted", "codex-rollout-versioned-failed-fail.jsonl")
	var versioned, merge, patched bool
	for _, e := range s.Trace.Events {
		switch {
		case e.Kind == "exec_item" && strings.Contains(e.Command, " versioned "):
			versioned = e.Success != nil && !*e.Success
		case e.Kind == "exec_item" && strings.HasPrefix(e.Command, "cd /work/code && gh pr merge"):
			merge = e.Success != nil && *e.Success
		case e.Kind == "write_item" && strings.HasSuffix(e.Path, "/changes/demo-change/design.md"):
			patched = true
		}
	}
	if !versioned || !merge || !patched {
		t.Fatalf("versioned failed = %v, merge succeeded = %v, patch path = %v", versioned, merge, patched)
	}
}

func TestAuditExpandAssignmentsIsBoundedAndExpandsAValueOnce(t *testing.T) {
	if got := expandAssignments("D=a/b; OLD=$D/c; echo $OLD"); got != "D=a/b; OLD=a/b/c; echo a/b/c" {
		t.Fatalf("expansion = %q", got)
	}
	var b strings.Builder
	b.WriteString("A00=xx;")
	for i := 1; i < 40; i++ {
		fmt.Fprintf(&b, "A%02d=$A%02d$A%02d;", i, i-1, i-1)
	}
	b.WriteString("echo $A39")
	chain := b.String()
	start := time.Now()
	got := expandAssignments(chain)
	if got != chain {
		t.Fatalf("a chain that outgrows the bound must stay unexpanded, got %d bytes", len(got))
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("expansion took %v", d)
	}
	if big := strings.Repeat("x", maxExpandedCommand+1) + "; F=a; echo $F"; expandAssignments(big) != big {
		t.Fatal("a command past the bound must stay unexpanded")
	}
}

func TestPlanWithoutBuildRouteNamesTheMessageWithoutTheBlock(t *testing.T) {
	got := auditStatus(t, auditFixture(t, "codex", auditOptions{}, "plan_without_build_route", "codex-rollout-typed-skill-no-block-fail.jsonl"), "plan_without_build_route")
	if got.Status != "fail" || len(got.Evidence) != 1 || !strings.Contains(got.Evidence[0], "codex-rollout-typed-skill-no-block-fail.jsonl:6 ") {
		t.Fatalf("plan_without_build_route = %+v, want fail at the message on line 6", got)
	}
}
