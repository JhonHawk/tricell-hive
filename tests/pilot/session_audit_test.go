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
	cases := []newCriterionCase{
		{"round codex title with a middle dot", "codex", round, []string{round, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"format codex title with a middle dot", "codex", format, []string{format, "codex-rollout-dot-title-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex without the card", "codex", prose, []string{prose, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"prose choice codex asked through the card", "codex", prose, []string{prose, "codex-rollout-card-pass.jsonl"}, auditOptions{}, "pass"},
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
