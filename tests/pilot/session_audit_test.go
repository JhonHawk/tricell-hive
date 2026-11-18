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
	if s := auditStatus(t, auditFixture(t, "claude", auditOptions{Cwd: t.TempDir()}, "existing_criteria", "claude-project-git-add-dir.jsonl"), "no_broad_git_add").Status; s != "pass" {
		t.Fatalf("pathspec that is not a directory here = %s, want pass", s)
	}
	if s := auditStatus(t, auditFixture(t, "claude", opts, "existing_criteria", "claude-project-git-add-file.jsonl"), "no_broad_git_add").Status; s != "pass" {
		t.Fatalf("file pathspec = %s, want pass", s)
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
	cases := []newCriterionCase{
		{"round claude fail", "claude", round, []string{round, "claude-project-fail.jsonl"}, auditOptions{}, "fail"},
		{"round claude pass", "claude", round, []string{round, "claude-project-pass.jsonl"}, auditOptions{}, "pass"},
		{"round claude no card in session", "claude", round, []string{round, "claude-project-nocard-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round claude title without option star", "claude", round, []string{round, "claude-project-notround-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex fail", "codex", round, []string{round, "codex-rollout-fail.jsonl"}, auditOptions{}, "fail"},
		{"round codex pass", "codex", round, []string{round, "codex-rollout-pass.jsonl"}, auditOptions{}, "pass"},
		{"round codex empty card before round", "codex", round, []string{round, "codex-rollout-empty-card-notobserved.jsonl"}, auditOptions{}, "not_observed"},
		{"round codex subagent", "codex", round, []string{round, "codex-rollout-subagent-notobserved.jsonl"}, auditOptions{}, "not_observed"},
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
