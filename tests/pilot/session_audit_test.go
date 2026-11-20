package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	const assess = "ticket_assessment_ending"
	const feed = "self_feeding_log"
	const reply = "analysis_reply_block"
	cases := []newCriterionCase{
		{"round codex title with a middle dot", "codex", round, []string{round, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex title with a middle dot", "codex", format, []string{format, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
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
		{"try it codex frontend child, report without the block", "codex", tryIt, []string{tryIt, "codex-rollout-child-fail.jsonl"}, auditOptions{}, "fail"},
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
		{"leaf codex no governed action", "codex", leaf, []string{leaf, "codex-rollout-notexercised.jsonl"}, auditOptions{}, "not_observed"},
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
		{"close claude review", "claude", close, []string{close, "claude-project-review.jsonl"}, auditOptions{}, "review"},
		{"close claude pass", "claude", close, []string{close, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"close claude no deletion", "claude", close, []string{close, "claude-project-nodelete-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close claude merge deletes the branch itself", "claude", close, []string{close, "claude-project-delete-branch-flag-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close grok review", "grok", close, []string{close, "grok-updates-review.jsonl"}, auditOptions{}, "review"},
		{"close grok pass", "grok", close, []string{close, "grok-updates-pass.jsonl"}, auditOptions{}, "pass"},
		{"close codex merge with unknown success is not a merge", "codex", close, []string{close, "codex-rollout-truncated-merge-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"close earlier merge is checked too", "claude", close, []string{close, "claude-project-two-merges-review.jsonl"}, auditOptions{}, "review"},
		{"close local merge and push of base", "claude", close, []string{close, "claude-project-local-merge-review.jsonl"}, auditOptions{Base: "development"}, "review"},
		{"close local merge without a base is not recognized", "claude", close, []string{close, "claude-project-local-merge-review.jsonl"}, auditOptions{Cwd: "elsewhere"}, "not_observed"},
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
			traceEvent{Line: 1, Kind: "shell", Command: `git commit -m "see hive-rule-git-commit/SKILL.md" a.go`},
		), "fail"},
		{"a failed read is no read", sess(
			traceEvent{Line: 1, Kind: "read", ID: "r1", Path: "/s/hive-rule-git-commit/SKILL.md", Success: &no},
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
			traceEvent{Line: 1, Kind: "shell", Command: "cat > /tmp/m <<'EOF'\nsee hive-rule-git-commit/SKILL.md\nEOF"},
			traceEvent{Line: 2, Kind: "shell", Command: "git commit -F /tmp/m a.go"},
		), "fail"},
		{"a GitLab-style update by iid is a move", sess(
			traceEvent{Line: 1, Kind: "tool", Tool: "update_issue", Input: []byte(`{"issue_iid":5,"state":"closed"}`)},
		), "fail"},
		{"cat of the leaf before the helper commit passes", sess(
			traceEvent{Line: 1, Kind: "shell", Command: "cat /s/hive-rule-git-commit/SKILL.md"},
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
	} {
		if got := shellWork(tc.command); got != tc.work {
			t.Errorf("shellWork(%q) = %v, want %v", tc.command, got, tc.work)
		}
	}
}
