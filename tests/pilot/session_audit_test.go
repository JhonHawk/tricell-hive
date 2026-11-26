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

// Sessions are written inside the tests as transcript text, one file per case,
// in the format of their host. They carry no real names, paths or ticket keys.
var sessionFileName = map[string]string{"claude": "session.jsonl", "codex": "rollout.jsonl", "grok": "updates.jsonl"}

// writeFiles writes each file (the key is a slash path under a new temporary
// folder) and returns the folder.
func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for rel, body := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(strings.TrimPrefix(body, "\n")), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func writeSession(t *testing.T, host, transcript string) string {
	t.Helper()
	name := sessionFileName[host]
	if name == "" {
		name = "session.jsonl"
	}
	return filepath.Join(writeFiles(t, map[string]string{name: transcript}), name)
}

func loadSession(t *testing.T, host string, opts auditOptions, transcript string) auditSession {
	t.Helper()
	s, err := loadAuditSession(writeSession(t, host, transcript), host, opts)
	if err != nil {
		t.Fatalf("load %s session: %v", host, err)
	}
	return s
}

func auditTranscript(t *testing.T, host string, opts auditOptions, transcript string) auditReport {
	t.Helper()
	rep, err := auditSessionFile(writeSession(t, host, transcript), host, opts)
	if err != nil {
		t.Fatalf("audit %s session: %v", host, err)
	}
	return rep
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
	s := loadSession(t, "claude", auditOptions{}, claudeAdapterSession)
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
	s := loadSession(t, "codex", auditOptions{}, codexAdapterRollout)
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
	root := writeFiles(t, map[string]string{"updates.jsonl": grokAdapterUpdates, "summary.json": `{"current_model_id":"grok-example-1"}`})
	s, err := loadAuditSession(filepath.Join(root, "updates.jsonl"), "grok", auditOptions{})
	if err != nil {
		t.Fatal(err)
	}
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
	for _, tc := range []struct{ host, transcript string }{
		{"claude", claudeAdapterSession},
		{"codex", codexAdapterRollout},
		{"grok", grokAdapterUpdates},
	} {
		var out bytes.Buffer
		if err := auditSessionMain(writeSession(t, tc.host, tc.transcript), tc.host, auditOptions{Base: "development"}, &out); err != nil {
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
	if err := auditSessionMain(writeSession(t, "claude", claudeAdapterSession), "pi", auditOptions{}, &out); err == nil {
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

func TestAuditRunsTheSevenExistingCriteria(t *testing.T) {
	rep := auditTranscript(t, "claude", auditOptions{}, claudeAdapterSession)
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

// sessionCase is one transcript and the status one criterion must report on it.
type sessionCase struct {
	name       string
	host       string
	criterion  string
	want       string
	opts       auditOptions
	disk       []string // files, and folders with a trailing slash, created under the session cwd
	hive       bool     // the session cwd is inside a Hive repository
	evidence   []string // substrings the evidence must hold
	transcript string
}

// TestAuditCriteria keeps, for each criterion, one case that fails and one for
// each other status it emits; not_observed has its own shared empty session.
func TestAuditCriteria(t *testing.T) {
	const round, close, skill = "round_text_and_card", "close_before_branch_cleanup", "skill_named_as_reason"
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
	const planRoute = "plan_without_build_route"
	const suite, rounds = "full_suite_after_text_fix", "review_rounds_per_candidate"
	cases := []sessionCase{
		{name: "round claude written out after a declined card", host: "claude", criterion: round, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "Build · EX-1 · ❓ D1 - Which delivery mode? Situation: one file.", "header": "D1", "multiSelect": false, "options": [{"label": "A - (★) interactive", "description": "> you validate first"}, {"label": "B - hold", "description": "> nothing is published"}]}]}}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "The user doesn't want to proceed with this tool use.", "is_error": true}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Build · EX-1 · delivery**\n\n❓ **D1** - **Which delivery mode?**\n\nSituation: one file.\n\nA - (★) **interactive**: you validate first.\n\nB - **hold**: nothing is published.\n\n➡️ A: you see it first.\n\nWaiting for D1."}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "prose codex decision title with no options", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000e1", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Continue"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "**Build · EX-1 · delivery**\n\n❓ **D3** - **¿Seguimos con interactive o con hold?**"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "reply claude italic bullets without letters", host: "claude", criterion: reply, want: "fail", evidence: []string{"not read"}, transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Analyze the cache."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Analysis · EX · the cache is read twice**\n\nThe loader reads the cache before and after the lock.\n\nWaiting for your reply. You can answer:\n\n- *Fix it* — I remove the second read.\n- *Leave it* — I record it as known."}]}}
`},
		{name: "reply claude italic bullets after reading the rule", host: "claude", criterion: reply, want: "fail", evidence: []string{"was read"}, transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Analyze the cache."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t2", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/hive-rule-report-endings/SKILL.md"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t2", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Analysis · EX · the cache is read twice**\n\nThe loader reads the cache before and after the lock.\n\nWaiting for your reply. You can answer:\n\n- *Fix it* — I remove the second read.\n- *Leave it* — I record it as known."}]}}
`},
		{name: "reply codex report ending with no replies", host: "codex", criterion: reply, want: "fail", evidence: []string{"without a reply block; report-endings rule was read"}, transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000cc", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Analyze the video and propose a few screens."}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "call_id": "p0", "input": "text(await tools.exec_command({cmd:\"cat /home/example/.agents/skills/hive-mode/SKILL.md /home/example/.agents/skills/hive-rule-report-endings/SKILL.md\"}));"}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "p0", "output": "ok"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "**Análisis · EX · Propuesta**\n\nEl estilo del video encaja con la app.\n\n1. **Antes de checar.** Hora actual y **Checar entrada**.\n2. **Confirmar.** Un arco que se llena al mantener presionado.\n3. **Jornada en curso.** Tiempo registrado hoy.\n4. **Comida en curso.** Tiempo desde la pausa.\n\nEl alcance termina en el análisis. Queda pendiente tu visto bueno para dibujar las cuatro vistas."}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "reply claude heading with the arrow and letters", host: "claude", criterion: reply, want: "pass", transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Analyze the cache."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Analysis · EX · the cache is read twice**\n\nThe loader reads the cache before and after the lock.\n\n**▶ How do we continue?**\n\nA — (★) Fix it: removes the second read.\nB — Leave it: record it as known."}]}}
`},
		{name: "async card as the only tool, no round in the text", host: "codex", criterion: asyncCard, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000d1", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input_async"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Continue"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "Asking the delivery now."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input_async", "call_id": "q1", "arguments": "{\"questions\": [{\"id\": \"a\", \"header\": \"D1\", \"question\": \"Which delivery mode?\", \"options\": [\"Interactive\", \"Hold\"]}]}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "q1", "output": "{\"accepted\": true}"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "async card as the only tool, round written first", host: "codex", criterion: asyncCard, want: "pass", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000d1", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input_async"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Continue"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "**Build · EX-1 · delivery**\n\nD1 — Which delivery mode?\n\nSituation: the change touches one example file.\n\nA - (★) Interactive\nIf you choose it: you validate before the push.\n\nB - Hold\nIf you choose it: nothing is published.\n\nI recommend A: you see it first.\n\nWaiting for D1."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input_async", "call_id": "q1", "arguments": "{\"questions\": [{\"id\": \"a\", \"header\": \"D1\", \"question\": \"Which delivery mode?\", \"options\": [\"Interactive\", \"Hold\"]}]}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "q1", "output": "{\"accepted\": true}"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "card options claude options that lost their letters", host: "claude", criterion: cards, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Round."}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 — Which delivery mode?", "header": "D1", "multiSelect": false, "options": [{"label": "Worktree propio (★)", "description": "x"}, {"label": "Checkout principal", "description": "x"}]}]}}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "Your questions have been answered.", "is_error": false}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "card options claude cards in format, session questions included", host: "claude", criterion: cards, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Round."}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 — Which approach?", "header": "D1", "multiSelect": false, "options": [{"label": "A - (★) Template", "description": "x"}, {"label": "B - Inline", "description": "x"}]}, {"question": "Delivery — Which delivery mode?", "header": "Delivery", "multiSelect": false, "options": [{"label": "A - (★) interactive", "description": "x"}, {"label": "B - automatic", "description": "x"}]}, {"question": "Review — Which review?", "header": "Review", "multiSelect": false, "options": [{"label": "A - (★) Native: Claude Code /code-review", "description": "x"}, {"label": "B - Agent: hive-review-code", "description": "x"}]}]}}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "Your questions have been answered.", "is_error": false}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "close claude fail", host: "claude", criterion: close, want: "fail", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"**Close · EXAMPLE · integrated**\n\nThe change is merged into the base branch."}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"d1","name":"Bash","input":{"command":"git branch -d topic","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"d1","content":"deleted","is_error":false}]}}
`},
		{name: "close claude pass", host: "claude", criterion: close, want: "pass", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"d1","name":"Bash","input":{"command":"git push origin --delete topic","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"d1","content":"deleted","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"**Close · EXAMPLE · integrated**\n\nThe change is merged into the base branch."}]}}
`},
		{name: "close report keeps a branch the session deleted", host: "claude", criterion: close, want: "review", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"**Close · EXAMPLE · integrated**\n\nThe change is merged. kept \u0060topic\u0060: the user asked to keep it until the release."}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"c3","name":"Bash","input":{"command":"git branch -d topic","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"c3","content":"deleted","is_error":false}]}}
`},
		{name: "format codex decisions as bullets", host: "codex", criterion: format, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000aa", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t3"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "go"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m5", "content": [{"type": "output_text", "text": "**EX-1 · Approved behavior**\n\nBefore editing, two delivery decisions remain:\n\n- **D1 · Delivery.** I recommend **\u0060hold\u0060**, which keeps the change local. Choose **\u0060interactive\u0060** to validate before the PR.\n- **D2 · Review.** I recommend **no dedicated review**. Choose **\u0060hive-review-code\u0060** for a second review.\n\nReply **with the recommended** to continue."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t6"}}
`},
		{name: "format codex full round in text", host: "codex", criterion: format, want: "pass", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000aa", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t2"}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "go"}]}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m4", "content": [{"type": "output_text", "text": "**Build · EX-1 · delivery**\n\nD1 — Which delivery mode?\n\nSituation: the change touches one example file.\n\nA - (★) Interactive\nIf you choose it: you validate before the push.\n\nB - Hold\nIf you choose it: nothing is published.\n\nI recommend A: you see it first.\n\nWaiting for D1."}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t5"}}
`},
		{name: "route claude ok to a proposal starts build", host: "claude", criterion: route, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Review ticket EX-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Proposal · EX-1 · return to Channels**\n\nThe arrow would go to Channels and keep the attempt.\n\nWaiting for your approval of the arrow's destination?"}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": "ok"}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r1", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/flow-build/SKILL.md"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "route claude typed slash command", host: "claude", criterion: route, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "Review ticket EX-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Proposal · EX-1 · return to Channels**\n\nThe arrow would go to Channels and keep the attempt.\n\nWaiting for your approval of the arrow's destination?"}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": "<command-message>flow-build</command-message>\n<command-name>/flow-build</command-name>"}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r1", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/flow-build/SKILL.md"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "route codex longer reply to a proposal", host: "codex", criterion: route, want: "review", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000aa", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t3"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Review ticket EX-1"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m5", "content": [{"type": "output_text", "text": "**Proposal · EX-1 · return to Channels**\n\nThe arrow would go to Channels and keep the attempt.\n\nWaiting for your approval of the arrow's destination?"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t6"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t7"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Return to Channels and also log the attempt"}]}}
{"timestamp": "2026-01-01T11:00:10.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "call_id": "call_1", "input": "text(await tools.exec_command({cmd:\"cat /home/example/.agents/skills/flow-build/SKILL.md\",yield_time_ms:1000}));"}}
{"timestamp": "2026-01-01T11:00:11.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_1", "output": [{"type": "input_text", "text": "Script completed"}, {"type": "input_text", "text": "{\"output\": \"ok\", \"exit_code\": 0}"}]}}
{"timestamp": "2026-01-01T11:00:12.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t11"}}
`},
		{name: "full_suite_after_text_fix claude two passing runs with nothing between them", host: "claude", criterion: suite, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Work on it."}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r1", "name": "Bash", "input": {"command": "go test -count=1 ./...", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r2", "name": "Bash", "input": {"command": "go test -count=1 ./...", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r2", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}, "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "full_suite_after_text_fix claude a code edit beside the markdown edit", host: "claude", criterion: suite, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "Work on it."}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r1", "name": "Bash", "input": {"command": "go test -count=1 ./...", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "e1", "name": "Edit", "input": {"file_path": "/work/project/README.md", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "e1", "content": "edited", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "e2", "name": "Edit", "input": {"file_path": "/work/project/main.go", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "e2", "content": "edited", "is_error": false}]}, "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r2", "name": "Bash", "input": {"command": "go test -count=1 ./...", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:08.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r2", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T10:00:09.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_5", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}, "timestamp": "2026-01-01T10:00:10.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "full_suite_after_text_fix claude piped runs with a PIPESTATUS echo are not shown to pass", host: "claude", criterion: suite, want: "review", transcript: `
{"type": "user", "message": {"role": "user", "content": "Work on it."}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r1", "name": "Bash", "input": {"command": "go test -count=1 ./... 2>&1 | tail -3; echo PIPESTATUS=${PIPESTATUS[0]}", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r1", "content": "ok  \t./...\nPIPESTATUS=0", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "e1", "name": "Edit", "input": {"file_path": "/work/project/docs/guide.mdx", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "e1", "content": "edited", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "r2", "name": "Bash", "input": {"command": "go test -count=1 ./... 2>&1 | tail -3; echo PIPESTATUS=${PIPESTATUS[0]}", "description": "example"}}]}, "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "r2", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}, "timestamp": "2026-01-01T10:00:08.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "leaf claude git add before the read", host: "claude", criterion: leaf, want: "fail", transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Do the work."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t0", "name": "Bash", "input": {"command": "git add notes.txt"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t0", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t1", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/hive-rule-git-commit/SKILL.md"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t1", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t2", "name": "Bash", "input": {"command": "git commit -m example"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t2", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:08.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_end", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}}
`},
		{name: "leaf claude reads before git, ticket and card", host: "claude", criterion: leaf, want: "pass", transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Do the work."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t0", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/hive-rule-git-commit/SKILL.md"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t0", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t1", "name": "Bash", "input": {"command": "git add notes.txt"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t1", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t2", "name": "Bash", "input": {"command": "git commit -m example"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t2", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:08.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t3", "name": "Bash", "input": {"command": "sed -n 1,40p /home/example/.agents/skills/hive-rule-ticket-state/SKILL.md"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:09.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t3", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:10.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t4", "name": "Bash", "input": {"command": "gh issue close 5"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:11.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t4", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:12.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_5", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t5", "name": "Read", "input": {"file_path": "/home/example/.agents/skills/hive-rule-question-channel/SKILL.md"}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:13.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t5", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:14.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_6", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "t6", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 — Which delivery mode?", "header": "Delivery", "options": [{"label": "A - (★) Interactive", "description": "x"}, {"label": "B - Automatic", "description": "y"}], "multiSelect": false}]}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:15.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "t6", "content": "ok", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:16.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_end", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}}
`},
		{name: "plan codex typed skill, build command without a next-step block", host: "codex", criterion: planRoute, want: "fail", evidence: []string{"rollout.jsonl:6 "}, transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-000000000042", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "id": "u5", "content": [{"type": "input_text", "text": "$flow-plan EX-1"}]}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "event_msg", "payload": {"type": "item_completed", "thread_id": "t", "turn_id": "t", "item": {"type": "UserMessage", "id": "i5", "content": [{"type": "text", "text": "$flow-plan EX-1", "text_elements": [{"byte_range": {"start": 0, "end": 10}, "placeholder": "$flow-plan"}]}, {"type": "skill", "name": "flow-plan", "path": "/home/u/.agents/skills/flow-plan/SKILL.md"}]}}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m3", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EX-1**\n\nEl plan queda en esta conversación.\n\n**Siguiente paso · Implementar EX-1**\n\n\u0060$flow-build EX-1\u0060"}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:10.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t2"}}
{"timestamp": "2026-01-01T11:00:11.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "id": "u11", "content": [{"type": "input_text", "text": "$flow-build EX-1"}]}}
{"timestamp": "2026-01-01T11:00:11.000Z", "type": "event_msg", "payload": {"type": "item_completed", "thread_id": "t", "turn_id": "t", "item": {"type": "UserMessage", "id": "i11", "content": [{"type": "text", "text": "$flow-build EX-1", "text_elements": [{"byte_range": {"start": 0, "end": 11}, "placeholder": "$flow-build"}]}, {"type": "skill", "name": "flow-build", "path": "/home/u/.agents/skills/flow-build/SKILL.md"}]}}}
{"timestamp": "2026-01-01T11:00:14.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m6", "phase": "final_answer", "content": [{"type": "output_text", "text": "Empiezo el build."}]}}
{"timestamp": "2026-01-01T11:00:15.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t2"}}
`},
		{name: "plan codex next-step block naming /flow-build", host: "codex", criterion: planRoute, want: "pass", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-000000000042", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "id": "u1", "content": [{"type": "input_text", "text": "planea EX-1"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "id": "ctc_2", "status": "completed", "call_id": "call_2", "name": "exec", "input": "text(await tools.exec_command({cmd:\"cat /home/u/.agents/skills/flow-plan/SKILL.md\"}));"}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_2", "output": [{"type": "input_text", "text": "Script completed"}, {"type": "input_text", "text": "{\"exit_code\": 0, \"output\": \"# skill\"}"}]}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m3", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan conjunto · EX-1**\n\n| Tarea | Trabajo |\n|---|---|\n| T1 | Cambiar el filtro de ejemplo |\n\nEl plan queda en esta conversación.\n\n---\n**▶ D2 — Siguiente paso: flow-build (recomendado)**\n\nEl plan no tiene decisiones abiertas.\n\n\u0060/flow-build EX-1\u0060\n\nEspero el comando de build."}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:10.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t2"}}
{"timestamp": "2026-01-01T11:00:11.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "id": "u4", "content": [{"type": "input_text", "text": "/flow-build EX-1"}]}}
{"timestamp": "2026-01-01T11:00:12.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "id": "ctc_5", "status": "completed", "call_id": "call_5", "name": "exec", "input": "text(await tools.exec_command({cmd:\"cat /home/u/.agents/skills/flow-build/SKILL.md\"}));"}}
{"timestamp": "2026-01-01T11:00:13.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_5", "output": [{"type": "input_text", "text": "Script completed"}, {"type": "input_text", "text": "{\"exit_code\": 0, \"output\": \"# skill\"}"}]}}
{"timestamp": "2026-01-01T11:00:14.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m6", "phase": "final_answer", "content": [{"type": "output_text", "text": "Empiezo el build."}]}}
{"timestamp": "2026-01-01T11:00:15.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t2"}}
`},
		{name: "plan codex session ends inside the stage", host: "codex", criterion: planRoute, want: "review", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-000000000042", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "id": "u1", "content": [{"type": "input_text", "text": "planea EX-1"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "id": "ctc_2", "status": "completed", "call_id": "call_2", "name": "exec", "input": "text(await tools.exec_command({cmd:\"cat /home/u/.agents/skills/flow-plan/SKILL.md\"}));"}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_2", "output": [{"type": "input_text", "text": "Script completed"}, {"type": "input_text", "text": "{\"exit_code\": 0, \"output\": \"# skill\"}"}]}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m3", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EX-1 · Decisiones**\n\n**1. ¿Hasta dónde entregamos?**\n\n- **A (★) · Local.** Sin publicar.\n- **B · PR.** Abro el PR.\n\nResponde 1A o 1B."}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice codex without the card", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Option A is approved. Before editing, two decisions remain. ¿Prefieres entrega \u0060interactive\u0060 para validar antes de publicar, o \u0060hold\u0060 para quedarte con el cambio local verificado? ¿Confirmas revisión de código y de seguridad de la API?"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice one quoted reply, delivery in the paragraph", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Antes de editar falta decidir la entrega. Voy a dejar el cambio local. Si te parece bien, responde «automatic» y arranco."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery modes as lettered bullets with a bold reply offer", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Preparación · EX-1**\n\n**Entrega**\n\n- **A — \u0060interactive\u0060 (recomendado).** Te muestro el cambio antes del push.\n- **B — \u0060automatic\u0060.** Publico e integro sin pausa.\n- **C — \u0060hold\u0060.** Dejo el cambio local.\n\nPuedes responder **entrega A, revisión A**."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery modes as bullets left pending", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**EX-1 · Plan**\n\n**Entrega propuesta**\n\n- **Interactiva (recomendada)** — Te muestro el cambio antes de publicar.\n- **Ready** — Dejo el PR listo para que lo integres.\n- **Automática** — Integro sin pausa.\n- **Local** — Sin commit ni publicación.\n\nQuedo pendiente del modo de entrega para empezar."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery round in the text with its header", host: "codex", criterion: prose, want: "not_observed", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EX-1 · entrega**\n\n❓ **Entrega** - **¿Cómo se entrega este cambio?**\n\nSituation: un archivo de ejemplo.\n\nA - (★) **interactive**: te muestro el cambio antes del push.\n\nB - **automatic**: publico e integro sin pausa.\n\nC - **hold**: dejo el cambio local.\n\n➡️ A: lo ves antes.\n\nEspero la entrega."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list after a round on another decision", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EX-1**\n\n❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n➡️ A: se revisa antes.\n\n**Entrega**\n\n- **A — \u0060interactive\u0060 (recomendado).** Te lo muestro antes.\n- **B — \u0060hold\u0060.** Queda local."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice masculine delivery modes as bullets", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**EX-1 · Plan**\n\n- **Modo interactivo (recomendado)** — te muestro el cambio.\n- **Automático** — integro sin pausa.\n\nQuedo pendiente de la entrega."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice status report with local and ready items", host: "codex", criterion: prose, want: "not_observed", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Build · EX-1 · listo**\n\n- **Local** checks: 12 pass (recommended rerun).\n- **Ready** for merge: PR #9.\n- a) ready to push."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list after a round closed by a bare arrow", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n➡ A: se revisa antes.\n\n**Entrega**\n\n- **A — \u0060interactive\u0060 (recomendado).** Te lo muestro antes.\n- **B — \u0060hold\u0060.** Queda local."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list after a round with no arrow line", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n**Entrega**\n\n- **A — \u0060interactive\u0060 (recomendado).** Te lo muestro antes.\n- **B — \u0060hold\u0060.** Queda local."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery modes with lowercase letter labels", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Entrega**\n\n- a) \u0060interactive\u0060 (recomendado): antes.\n- b) \u0060hold\u0060: local."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice titled list with the star after the label, not a round", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **Delivery** - **How is this change delivered?**\n\nSituation: one file.\n\n- **A — \u0060interactive\u0060** (★) show first.\n- **B — \u0060hold\u0060** local.\n\n➡️ A: you see it first."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list under an unstarred title after a round", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n❓ **D2** - **¿Cómo se entrega?**\n\n- **A — \u0060interactive\u0060** (recomendado)\n- **B — \u0060hold\u0060**"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice lettered list of ready and local only", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Entrega**\n\n- **A — \u0060ready\u0060** (recommended): PR open.\n- **B — \u0060local\u0060**: stays."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice status report with lowercase ready and local items", host: "codex", criterion: prose, want: "not_observed", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Build · EX-1**\n\n- a) ready to push.\n- b) local checks pass."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list right after a round's options", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n- **A — \u0060interactive\u0060 (recomendado).**\n- **B — \u0060hold\u0060.**"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice round with indented option descriptions and no arrow", host: "codex", criterion: prose, want: "not_observed", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **Entrega** - **¿Cómo se entrega?**\n\nA - (★) **interactive**\n  te lo muestro antes.\nB - **hold**\n  queda local.\nC - **automatic**\n  integro sin pausa.\n\nEspero la entrega."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice delivery list before a later arrow", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "❓ **D1** - **¿Qué alcance?**\n\nSituation: dos módulos.\n\nA - (★) **Uno**: menos riesgo.\n\nB - **Ambos**: más trabajo.\n\n**Entrega**\n\n- **A — \u0060interactive\u0060 (recomendado)**\n- **B — \u0060hold\u0060**\n\nSiguiente ➡ build"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice report explaining each delivery mode", host: "claude", criterion: prose, want: "not_observed", transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "A"}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Validation · EX · D1 decided**\n\nYou chose D1 A: the findings round comes before publishing.\n- **interactive**: before the push, with the validation.\n- **ready**: at the stop with the pull request open.\n- **automatic** and **hold**: unchanged.\n\nNothing was edited yet."}]}}
`},
		{name: "prose choice report naming one delivery mode", host: "codex", criterion: prose, want: "not_observed", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EX-1 · listo**\n\n- **Entrega:** \u0060automatic\u0060.\n- **Revisión:** nativa.\n\nEl plan vive en la conversación."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice on a host with no card tool", host: "codex", criterion: prose, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Antes de editar falta decidir la entrega. Voy a dejar el cambio local. Si te parece bien, responde «automatic» y arranco."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "async card, session without an async call", host: "codex", criterion: prose, want: "pass", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**EX-1 · Delivery**\n\nThe delivery and review questions follow in the card."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input", "call_id": "q1", "arguments": "{\"questions\": [{\"id\": \"delivery\", \"header\": \"Delivery\", \"question\": \"Delivery \\u2014 how is the change delivered?\", \"options\": [{\"label\": \"A - (\\u2605) Interactive\", \"description\": \"You validate before the push.\"}, {\"label\": \"B - Hold\", \"description\": \"Local verified change only.\"}]}]}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "q1", "output": "{\"answers\": {\"delivery\": {\"answers\": [\"A - (\\u2605) Interactive\"]}}}"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Delivery: interactive."}]}}
{"timestamp": "2026-01-01T11:00:10.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "prose choice ordinary word ready", host: "codex", criterion: prose, want: "review", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "continuemos"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Done. Should I merge now or wait until CI is ready?"}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "review question without the native review", host: "codex", criterion: review, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-000000000009", "cwd": "project", "source": "vscode"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input", "description": "only available in Default or Plan mode"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "**Plan · EXAMPLE · ready**\n\n**D4 — ¿Qué revisión de código hacemos?**\n\nSituación: el cambio mueve scripts.\n\n- **A · (★) Agente \u0060hive-review-code\u0060.** Revisa la selección de CI.\n- **B · Sin revisión dedicada.** Se apoya en las pruebas.\n\nEspero D4."}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "review card with the native review", host: "codex", criterion: review, want: "pass", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-000000000009", "cwd": "project", "source": "vscode"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input", "description": "only available in Default or Plan mode"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input", "call_id": "call_9", "arguments": "{\"questions\": [{\"id\": \"review\", \"header\": \"Revisión\", \"question\": \"Integración · EXAMPLE · D1 — ¿Qué revisión de código hacemos antes del push?\", \"options\": [{\"label\": \"Codex review (Recommended)\", \"description\": \"x\"}, {\"label\": \"Agent hive-review-code\", \"description\": \"y\"}, {\"label\": \"No dedicated review\", \"description\": \"z\"}]}]}"}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "call_9", "output": "{\"answers\":{\"review\":{\"answers\":[\"Codex review (Recommended)\"]}}}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m1", "phase": "final_answer", "content": [{"type": "output_text", "text": "Applying the review you chose."}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "review_rounds_per_candidate claude three code-review skill rounds", host: "claude", criterion: rounds, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Work on it."}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s1", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s1", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s2", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s2", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s3", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s3", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "review_rounds_per_candidate claude two rounds", host: "claude", criterion: rounds, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "Work on it."}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s1", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s1", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s2", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s2", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "review_rounds_per_candidate claude a typed /code-review before two more rounds", host: "claude", criterion: rounds, want: "review", transcript: `
{"type": "user", "message": {"role": "user", "content": "<command-message>code-review</command-message>\n<command-name>/code-review</command-name>"}, "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s1", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s1", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s2", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s2", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s3", "name": "Skill", "input": {"skill": "code-review", "args": "origin/development..HEAD"}}]}, "timestamp": "2026-01-01T10:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s3", "content": "reviewed", "is_error": false}]}, "timestamp": "2026-01-01T10:00:07.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "feed claude exec redirect with a trap that cats the file", host: "claude", criterion: feed, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false, "type": "user", "message": {"role": "user", "content": "run the checks"}}
{"timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false, "type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u2", "name": "Bash", "input": {"command": "set -eu\nci_log=$(mktemp -t build-check)\nexec >\"$ci_log\" 2>&1\ntrap 'cat \"$ci_log\"' EXIT\nmake unit\nmake build"}}]}}
{"timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false, "type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u2", "content": "ok", "is_error": false}]}}
{"timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false, "type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}}
`},
		{name: "skill claude review", host: "claude", criterion: skill, want: "review", opts: auditOptions{Cwd: "elsewhere"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Continue."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"I stop before the push because flow-build requires your validation first."}]}}
`},
		{name: "skill claude pass", host: "claude", criterion: skill, want: "pass", opts: auditOptions{Cwd: "elsewhere"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Continue."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"I stop before the push because the change alters a screen you have to judge. I ran flow-build earlier."}]}}
`},
		{name: "skill named as reason in the Hive repository is legitimate", host: "claude", criterion: skill, want: "not_observed", hive: true, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Continue."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"I stop before the push because flow-build requires your validation first."}]}}
`},
		{name: "round text and the same card in one turn", host: "claude", criterion: round, want: "fail", evidence: []string{"session.jsonl:"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"❓ **D8** - **Which approach does the plan follow?**\n\nSituation: the change touches one example file.\n\nA - (★) **Restore text**: restores the text.\n\nB - **Restore and hook**: adds a hook.\n\n➡️ A: it is the smaller change; the hook stays uncertain."},{"type":"tool_use","id":"q1","name":"AskUserQuestion","input":{"questions":[{"question":"Entrega · EX-1 · ❓ D8 - Which approach? Situation: one file.","header":"D8","options":[{"label":"A - (★) Restore text","description":"> restores the text"},{"label":"B - Restore and hook","description":"> adds a hook"}],"multiSelect":false}]}}]}}
`},
		{name: "round card alone", host: "claude", criterion: round, want: "pass", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Working on it."},{"type":"tool_use","id":"q1","name":"AskUserQuestion","input":{"questions":[{"question":"Entrega · EX-1 · ❓ D8 - Which approach? Situation: one file.","header":"D8","options":[{"label":"A - (★) Restore text","description":"> restores the text"},{"label":"B - Restore and hook","description":"> adds a hook"}],"multiSelect":false}]}}]}}
`},
		{name: "round full text with no card", host: "claude", criterion: round, want: "pass", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"❓ **D8** - **Which approach does the plan follow?**\n\nSituation: the change touches one example file.\n\nA - (★) **Restore text**: restores the text.\n\nB - **Restore and hook**: adds a hook.\n\n➡️ A: it is the smaller change; the hook stays uncertain."}]}}
`},
		{name: "round plain message", host: "claude", criterion: round, want: "not_observed", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Working on it."}]}}
`},
		{name: "format new text round in full", host: "claude", criterion: format, want: "pass", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"❓ **D8** - **Which approach does the plan follow?**\n\nSituation: the change touches one example file.\n\nA - (★) **Restore text**: restores the text.\n\nB - **Restore and hook**: adds a hook.\n\n➡️ A: it is the smaller change; the hook stays uncertain."}]}}
`},
		{name: "format new text title without Situation or star", host: "claude", criterion: format, want: "fail", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_0","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"❓ **D8** - **Which approach does the plan follow?**\n\nA - **Restore text**: restores the text.\n\nB - **Restore and hook**: adds a hook."}]}}
`},
		{name: "assessment claude skill call, no ending", host: "claude", criterion: assess, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "Review ticket EX-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s0", "name": "Skill", "input": {"skill": "hive-mode"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s0", "content": "Launching skill: hive-mode", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Review · EX-1 · defect confirmed in code**\n\nThe form saves even when the name did not change.\n\n**Readiness.** I would pin three observable criteria; the rate limit still needs a threshold."}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "assessment claude ends with the answers block", host: "claude", criterion: assess, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "Review ticket EX-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "s0", "name": "Skill", "input": {"skill": "hive-mode"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "s0", "content": "Launching skill: hive-mode", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "**Review · EX-1 · defect confirmed in code**\n\nThe form saves even when the name did not change.\n\n**Readiness.** I would pin three observable criteria; the rate limit still needs a threshold.\n\n▶ ¿Cómo seguimos?\n- A: planificar\n- B: dejarlo"}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "assessment codex ends with an open question in text", host: "codex", criterion: assess, want: "review", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Review ticket EX-1"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "call_id": "c1", "input": "text(await tools.exec_command({cmd:\"cat /home/example/.agents/skills/hive-mode/SKILL.md\",yield_time_ms:1000}));"}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "c1", "output": [{"type": "input_text", "text": "Script completed"}, {"type": "input_text", "text": "{\"output\": \"ok\", \"exit_code\": 0}"}]}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "**Review · EX-1 · defect confirmed in code**\n\nThe form saves even when the name did not change.\n\n**Readiness.** I would pin three observable criteria; the rate limit still needs a threshold.\n\nShould the rate limit stay inside EX-1 or become its own ticket?"}]}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{name: "cache claude search empty, move, no write", host: "claude", criterion: cache, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "build ABC-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u2", "name": "mcp__engram__mem_search", "input": {"project": "p", "query": "tracker/states/jira/ABC"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u2", "content": "No memories found for: \"tracker/states/jira/ABC\"", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u4", "name": "Bash", "input": {"command": "acli jira workitem transition --key ABC-1 --status \"En curso\" --yes"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u4", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_8", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Moved."}]}, "timestamp": "2026-01-01T11:00:08.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "cache claude written through the engram CLI", host: "claude", criterion: cache, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "build ABC-1"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u2", "name": "mcp__engram__mem_search", "input": {"project": "p", "query": "tracker/states/jira/ABC"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u2", "content": "No memories found for: \"tracker/states/jira/ABC\"", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u4", "name": "Bash", "input": {"command": "acli jira workitem transition --key ABC-1 --status \"En curso\" --yes"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u4", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_6", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u6", "name": "Bash", "input": {"command": "engram save --project p --type decision --topic tracker/states/jira/ABC 'states' 'x'"}}]}, "timestamp": "2026-01-01T11:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u6", "content": "saved", "is_error": false}]}, "timestamp": "2026-01-01T11:00:07.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_8", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Moved."}]}, "timestamp": "2026-01-01T11:00:08.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "try it claude final report without the block", host: "claude", criterion: tryIt, want: "fail", transcript: `
{"type": "user", "message": {"role": "user", "content": "fix the button"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u2", "name": "Edit", "input": {"file_path": "/work/project/src/Button.tsx", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u2", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u4", "name": "Skill", "input": {"skill": "flow-verify"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u4", "content": "loaded", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_6", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done. PR #2 merged."}]}, "timestamp": "2026-01-01T11:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "try it claude final report with the block", host: "claude", criterion: tryIt, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "fix the button"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u2", "name": "Edit", "input": {"file_path": "/work/project/src/Button.tsx", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u2", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_4", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "u4", "name": "Skill", "input": {"skill": "flow-verify"}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "u4", "content": "loaded", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_6", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done.\n\n### Try it\n- Open http://localhost:3000\n- Click Save.\n- The toast appears."}]}, "timestamp": "2026-01-01T11:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "ux codex screen edit without the reviewer", host: "codex", criterion: ux, want: "fail", transcript: `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000aa", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t3"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "fix"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "call_id": "call_1", "input": "text(await tools.apply_patch(\"*** Begin Patch\\n*** Update File: apps/web/src/app/page.tsx\\n@@\\n-a\\n+b\\n*** End Patch\"));"}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_1", "output": [{"type": "input_text", "text": "Script completed"}]}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "custom_tool_call", "name": "exec", "call_id": "call_2", "input": "text(await tools.apply_patch(\"*** Begin Patch\\n*** Update File: apps/web/e2e/page.e2e.ts\\n@@\\n-a\\n+b\\n*** End Patch\"));"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "response_item", "payload": {"type": "custom_tool_call_output", "call_id": "call_2", "output": [{"type": "input_text", "text": "Script completed"}]}}
{"timestamp": "2026-01-01T11:00:10.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m9", "content": [{"type": "output_text", "text": "Done."}]}}
{"timestamp": "2026-01-01T11:00:11.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t10"}}
`},
		{name: "ux claude screen edit with the reviewer", host: "claude", criterion: ux, want: "pass", transcript: `
{"type": "user", "message": {"role": "user", "content": "fix"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "e1", "name": "Edit", "input": {"file_path": "/work/project/src/Button.tsx", "old_string": "a", "new_string": "b"}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "e1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_3", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "g1", "name": "Agent", "input": {"subagent_type": "hive-review-ux", "description": "UX review", "prompt": "Review."}}]}, "timestamp": "2026-01-01T11:00:04.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "g1", "content": "ok", "is_error": false}]}, "timestamp": "2026-01-01T11:00:05.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_5", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Done."}]}, "timestamp": "2026-01-01T11:00:06.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{name: "wait claude a background loop with sleep and no deadline", host: "claude", criterion: wait, want: "fail", evidence: []string{".jsonl:2 shell"}, transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Wait for the deploy."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "w1", "name": "Bash", "input": {"command": "while :; do status=$(example-cli deploy status --job 1 --output text); if [ \"$status\" = DONE ]; then break; fi; sleep 20; done", "description": "example", "run_in_background": true}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "w1", "content": "started", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Waiting."}]}}
`},
		{name: "wait grok a loop with block_until_ms 0 and no deadline", host: "grok", criterion: wait, want: "fail", evidence: []string{".jsonl:2 shell"}, transcript: `
{"method": "session/update", "params": {"sessionId": "00000000-0000-0000-0000-000000000003", "update": {"sessionUpdate": "user_message_chunk", "content": {"type": "text", "text": "Wait for the deploy."}}}, "timestamp": 1794948001, "timestamp_us": 1794948001000000}
{"method": "session/update", "params": {"sessionId": "00000000-0000-0000-0000-000000000003", "update": {"sessionUpdate": "tool_call", "toolCallId": "call-1", "title": "run_terminal_command", "rawInput": {"command": "while :; do status=$(example-cli deploy status --job 1 --output text); if [ \"$status\" = DONE ]; then break; fi; sleep 20; done", "description": "example", "block_until_ms": 0}, "_meta": {"x.ai/tool": {"version": 1, "name": "run_terminal_command", "kind": "execute", "namespace": "grok_build"}}}}, "timestamp": 1794948002, "timestamp_us": 1794948002000000}
{"method": "session/update", "params": {"sessionId": "00000000-0000-0000-0000-000000000003", "update": {"sessionUpdate": "tool_call_update", "toolCallId": "call-1", "status": "completed", "rawOutput": {"type": "Terminal", "command": "while :; do status=$(example-cli deploy status --job 1 --output text); if [ \"$status\" = DONE ]; then break; fi; sleep 20; done", "exit_code": 0, "output": "started", "timed_out": false}}}, "timestamp": 1794948003, "timestamp_us": 1794948003000000}
{"method": "session/update", "params": {"sessionId": "00000000-0000-0000-0000-000000000003", "update": {"sessionUpdate": "agent_message_chunk", "content": {"type": "text", "text": "Waiting."}}}, "timestamp": 1794948004, "timestamp_us": 1794948004000000}
{"method": "_x.ai/session/update", "params": {"sessionId": "00000000-0000-0000-0000-000000000003", "update": {"sessionUpdate": "turn_completed", "prompt_id": "p1", "stop_reason": "end_turn", "elapsed_ms": 1}}, "timestamp": 1794948005, "timestamp_us": 1794948005000000}
`},
		{name: "wait claude timeout wrapping the loop", host: "claude", criterion: wait, want: "pass", transcript: `
{"type": "user", "timestamp": "2026-01-01T10:00:01.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": "Wait for the deploy."}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:02.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_1", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "w1", "name": "Bash", "input": {"command": "timeout 600 bash -c 'until [ \"$(example-cli deploy status --job 1 --output text)\" = DONE ]; do sleep 20; done'", "description": "example", "run_in_background": true}}]}}
{"type": "user", "timestamp": "2026-01-01T10:00:03.000Z", "cwd": "project", "isSidechain": false, "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "w1", "content": "started", "is_error": false}]}}
{"type": "assistant", "timestamp": "2026-01-01T10:00:04.000Z", "cwd": "project", "isSidechain": false, "message": {"id": "msg_2", "model": "claude-example-1", "role": "assistant", "content": [{"type": "text", "text": "Waiting."}]}}
`},
		{name: "merged branch deleted: a merge left after the user ended the session, claude", host: "claude", criterion: "merged_branch_deleted", want: "fail", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"c1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"c1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Merged."}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"es todo por el momento, gracias!"}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Bye."}]}}
`},
		{name: "merged branch deleted: a merge left after the user ended the session, codex", host: "codex", criterion: "merged_branch_deleted", want: "fail", transcript: `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"00000000-0000-0000-0000-000000000001","cwd":"project","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"project"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"Merge it."}]}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c1","input":"text(await tools.exec_command({cmd:\"gh pr merge 1 --repo example/repo --merge\",yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c1","output":[{"type":"input_text","text":"Script completed"},{"type":"input_text","text":"{\"output\": \"merged\", \"exit_code\": 0}"}]}}
{"timestamp":"2026-01-01T11:00:07.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"m1","content":[{"type":"output_text","text":"Merged."}]}}
{"timestamp":"2026-01-01T11:00:08.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t1"}}
{"timestamp":"2026-01-01T11:00:09.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t2"}}
{"timestamp":"2026-01-01T11:00:10.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"es todo por el momento, gracias!"}]}}
{"timestamp":"2026-01-01T11:00:11.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"m2","content":[{"type":"output_text","text":"Bye."}]}}
{"timestamp":"2026-01-01T11:00:12.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t2"}}
`},
		{name: "merged branch deleted: a merge left after the user ended the session, grok", host: "grok", criterion: "merged_branch_deleted", want: "fail", transcript: `
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Merge it."}}},"timestamp":1794948001,"timestamp_us":1794948001000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call","toolCallId":"call-1","title":"run_terminal_command","rawInput":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"},"_meta":{"x.ai/tool":{"version":1,"name":"run_terminal_command","kind":"execute","namespace":"grok_build"}}}},"timestamp":1794948002,"timestamp_us":1794948002000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call_update","toolCallId":"call-1","status":"completed","rawOutput":{"type":"Terminal","command":"gh pr merge 1 --repo example/repo --merge","exit_code":0,"output":"merged","timed_out":false}}},"timestamp":1794948003,"timestamp_us":1794948003000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Merged."}}},"timestamp":1794948004,"timestamp_us":1794948004000000}
{"method":"_x.ai/session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"turn_completed","prompt_id":"p1","stop_reason":"end_turn","elapsed_ms":1}},"timestamp":1794948005,"timestamp_us":1794948005000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"es todo por el momento, gracias!"}}},"timestamp":1794948006,"timestamp_us":1794948006000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Bye."}}},"timestamp":1794948007,"timestamp_us":1794948007000000}
{"method":"_x.ai/session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"turn_completed","prompt_id":"p2","stop_reason":"end_turn","elapsed_ms":1}},"timestamp":1794948008,"timestamp_us":1794948008000000}
`},
		{name: "merged branch deleted: the cleanup helper run with --apply", host: "claude", criterion: "merged_branch_deleted", want: "pass", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"d1","name":"Bash","input":{"command":"python3 -I scripts/cleanup_branches.py --repo . --base development --merged-pr topic=abc --apply","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"d1","content":"done","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Finished."}]}}
`},
		{name: "merged branch deleted: no close report after the last merge", host: "claude", criterion: "merged_branch_deleted", want: "review", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Merged, promoting next."}]}}
`},
		{name: "merged branch deleted: the cleanup helper run without --apply is a dry run", host: "claude", criterion: "merged_branch_deleted", want: "fail", transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_m1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 1 --repo example/repo --merge","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_d1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"d1","name":"Bash","input":{"command":"python3 -I scripts/cleanup_branches.py --repo . --base development --merged-pr topic=abc","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"d1","content":"done","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Finished."}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:07.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"**Close** - work item done."}]}}
`},
		{name: "merged branch deleted: the head is that of another pull request", host: "claude", criterion: "merged_branch_deleted", want: "fail", opts: auditOptions{Base: "development"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_v1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"v1","name":"Bash","input":{"command":"gh pr view 41 --repo example/repo --json headRefName","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"v1","content":"{\"headRefName\":\"development\"}","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_m1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"gh pr merge 42 --repo example/repo --squash","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"merged","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Merged."}]}}
{"type":"user","timestamp":"2026-01-01T10:00:07.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"es todo"}}
{"type":"assistant","timestamp":"2026-01-01T10:00:08.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Bye."}]}}
`},
		{name: "merged branch deleted: a local merge with no base branch is not observed", host: "claude", criterion: "merged_branch_deleted", want: "not_observed", evidence: []string{"base branch"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Merge locally."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_m1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"m1","name":"Bash","input":{"command":"git merge topic","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"m1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_p1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"p1","name":"Bash","input":{"command":"git push origin development","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"p1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Finished."}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:07.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"**Close** - work item done."}]}}
`},
		{name: "leaf claude stage packet cut in the middle", host: "claude", criterion: leaf, want: "not_observed", evidence: []string{"does not show the leaf whole"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Do the work."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_r1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"r1","name":"Read","input":{"file_path":"/home/example/.agents/skills/flow-build/stage.md"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"r1","content":"     1\u2192## hive-core/git/git-commit\n     2\u2192\n     3\u2192# Git commit\n     4\u2192\n\u202612 lines truncated\u2026\n\n<!-- hive packet end: stage \u00b7 hive-core/git/git-commit -->\n","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_g1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"g1","name":"Bash","input":{"command":"git add notes.txt"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"g1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Done."}]}}
`},
		{name: "leaf claude stage packet cut before its end line", host: "claude", criterion: leaf, want: "not_observed", evidence: []string{"does not show the leaf whole"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Do the work."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_r1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"r1","name":"Read","input":{"file_path":"/home/example/.agents/skills/flow-build/stage.md"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"r1","content":"     1\u2192## hive-core/git/git-commit\n     2\u2192\n     3\u2192# Git commit\n     4\u2192\n     5\u2192Rules.","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_g1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"g1","name":"Bash","input":{"command":"git add notes.txt"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"g1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_e1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Done."}]}}
`},
		{name: "leaf codex git add after a shell command the host withheld", host: "codex", criterion: leaf, want: "review", transcript: `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"00000000-0000-0000-0000-000000000042","cwd":"project","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"project"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"Do the work."}]}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c1","input":"text(await tools.exec_command({cmd:\"cat \" + dir + \"/notes.txt\"}));"}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c1","output":[{"type":"input_text","text":"Script completed"},{"type":"input_text","text":"{\"output\": \"x\", \"exit_code\": 0}"}]}}
{"timestamp":"2026-01-01T11:00:07.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c2","input":"text(await tools.exec_command({cmd:\"git add notes.txt\"}));"}}
{"timestamp":"2026-01-01T11:00:08.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"c2","output":[{"type":"input_text","text":"Script completed"},{"type":"input_text","text":"{\"output\": \"ok\", \"exit_code\": 0}"}]}}
{"timestamp":"2026-01-01T11:00:09.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"m1","content":[{"type":"output_text","text":"Done."}]}}
{"timestamp":"2026-01-01T11:00:10.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t1"}}
`},
		{name: "git add -A", host: "claude", criterion: "no_broad_git_add", want: "fail", disk: []string{"docs/", "notes.txt"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Stage it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"a1","name":"Bash","input":{"command":"git add -A","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"a1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Staged."}]}}
`},
		{name: "git add of a directory that exists on disk", host: "claude", criterion: "no_broad_git_add", want: "review", disk: []string{"docs/", "notes.txt"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Stage it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"a1","name":"Bash","input":{"command":"git add docs","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"a1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Staged."}]}}
`},
		{name: "git add of a file that exists on disk", host: "claude", criterion: "no_broad_git_add", want: "pass", disk: []string{"docs/", "notes.txt"}, transcript: `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Stage it."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"a1","name":"Bash","input":{"command":"git add notes.txt","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"a1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Staged."}]}}
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.opts
			if len(tc.disk) > 0 {
				opts.Cwd = t.TempDir()
				for _, p := range tc.disk {
					full := filepath.Join(opts.Cwd, filepath.FromSlash(p))
					if strings.HasSuffix(p, "/") {
						if err := os.MkdirAll(full, 0o700); err != nil {
							t.Fatal(err)
						}
					} else if err := os.WriteFile(full, []byte("x"), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			if tc.hive {
				opts.Cwd = hiveLikeDir(t)
			}
			got := auditStatus(t, auditTranscript(t, tc.host, opts, tc.transcript), tc.criterion)
			if got.Status != tc.want {
				t.Fatalf("%s = %s, want %s (evidence %v)", tc.criterion, got.Status, tc.want, got.Evidence)
			}
			if (tc.want == "fail" || tc.want == "review") && len(got.Evidence) == 0 {
				t.Fatalf("%s %s has no evidence", tc.criterion, tc.want)
			}
			for _, want := range tc.evidence {
				if len(got.Evidence) != 1 || !strings.Contains(got.Evidence[0], want) {
					t.Fatalf("evidence = %v, want one line holding %q", got.Evidence, want)
				}
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

// TestAuditEmptySessionObservesNothing is the shared not_observed case: a
// session with no turn gives every criterion nothing to judge, in every host.
func TestAuditEmptySessionObservesNothing(t *testing.T) {
	for _, host := range []string{"claude", "codex", "grok"} {
		rep := auditTranscript(t, host, auditOptions{Base: "development"}, "")
		if len(rep.Criteria) == 0 {
			t.Fatalf("%s: no criteria in the report", host)
		}
		for _, c := range rep.Criteria {
			if c.Status != "not_observed" {
				t.Errorf("%s: %s = %s on an empty session", host, c.Criterion, c.Status)
			}
		}
	}
}

func TestAuditUnknownSuccessIsNotRecordedAsSuccess(t *testing.T) {
	for _, tc := range []struct{ name, transcript string }{
		{"truncated merge", `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"00000000-0000-0000-0000-000000000001","cwd":"project","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"project"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t2"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_1","input":"text(await tools.exec_command({cmd:\"gh pr merge 1 --repo example/repo --merge\",yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_1","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"Warning: truncated output (original token count: 5000)"}]}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"msg2","phase":"final_answer","content":[{"type":"output_text","text":"**Close · EXAMPLE · integrated**\n\nThe change is merged."}]}}
{"timestamp":"2026-01-01T11:00:07.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_3","input":"text(await tools.exec_command({cmd:\"git branch -d topic\",yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:08.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_3","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"chunk_id\": \"1\", \"exit_code\": 0, \"output\": \"example output\", \"wall_time_seconds\": 0.1}"}]}}
{"timestamp":"2026-01-01T11:00:09.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t8"}}
`},
		{"running merge", `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"00000000-0000-0000-0000-000000000001","cwd":"project","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"project"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t2"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_1","input":"text(await tools.exec_command({cmd:\"gh pr merge 1 --repo example/repo --merge\",yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_1","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"session_id\": 7, \"wall_time_seconds\": 1.0, \"output\": \"still running\"}"}]}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t5"}}
`},
	} {
		s := loadSession(t, "codex", auditOptions{}, tc.transcript)
		var seen bool
		for _, e := range s.Trace.Events {
			if e.Kind == "shell" && strings.HasPrefix(e.Command, "gh pr merge") {
				seen = true
				if e.Success != nil {
					t.Fatalf("%s: success must be unknown, got %v", tc.name, *e.Success)
				}
			}
		}
		if !seen {
			t.Fatalf("%s: merge command not extracted", tc.name)
		}
		merged := auditStatus(t, auditTranscript(t, "codex", auditOptions{}, tc.transcript), "merged_branch_deleted")
		if merged.Status != "not_observed" {
			t.Fatalf("%s: merged_branch_deleted = %+v", tc.name, merged)
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
		name, host          string
		answered, recommend int
		transcript          string
	}{
		{"claude answered with an option", "claude", 2, 1, `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 - Which verdict?", "header": "D1", "multiSelect": false, "options": [{"label": "A - (★) Judgment", "description": "x"}, {"label": "B - Mode", "description": "y"}]}, {"question": "Delivery - Where does it stop?", "header": "Delivery", "multiSelect": false, "options": [{"label": "A - (★) interactive", "description": "x"}, {"label": "B - automatic", "description": "y"}]}]}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "Your questions have been answered: \"D1 - Which verdict?\"=\"A - (★) Judgment\", \"Delivery - Where does it stop?\"=\"B - automatic\". You can now continue with these answers in mind.", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{"claude free text is not an option", "claude", 1, 0, `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 - Which verdict?", "header": "D1", "multiSelect": false, "options": [{"label": "A - (★) Judgment", "description": "x"}, {"label": "B - Mode", "description": "y"}]}, {"question": "Delivery - Where does it stop?", "header": "Delivery", "multiSelect": false, "options": [{"label": "A - (★) interactive", "description": "x"}, {"label": "B - automatic", "description": "y"}]}]}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "Your questions have been answered: \"D1 - Which verdict?\"=\"neither, do something else\", \"Delivery - Where does it stop?\"=\"B - automatic\". You can now continue with these answers in mind.", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
`},
		{"codex answered with Recommended and recomendado labels", "codex", 3, 2, `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Continue"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "Round."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input", "call_id": "q1", "arguments": "{\"questions\": [{\"id\": \"a\", \"header\": \"A\", \"question\": \"Scope?\", \"options\": [{\"label\": \"Fix (Recommended)\", \"description\": \"x\"}, {\"label\": \"Limit\", \"description\": \"y\"}]}, {\"id\": \"b\", \"header\": \"B\", \"question\": \"Review?\", \"options\": [{\"label\": \"Agent (recomendado)\", \"description\": \"x\"}, {\"label\": \"Native\", \"description\": \"y\"}]}, {\"id\": \"c\", \"header\": \"C\", \"question\": \"Security?\", \"options\": [{\"label\": \"Include (Recommended)\", \"description\": \"x\"}, {\"label\": \"General\", \"description\": \"y\"}]}]}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "q1", "output": "{\"answers\": {\"a\": {\"answers\": [\"Fix (Recommended)\"]}, \"b\": {\"answers\": [\"Native\"]}, \"c\": {\"answers\": [\"Include (Recommended)\"]}}}"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
		{"codex empty answers", "codex", 0, 0, `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000bb", "cwd": "project", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "project", "collaboration_mode": {"mode": "default"}}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "response_item", "payload": {"type": "additional_tools", "id": "at1", "role": "developer", "tools": [{"type": "namespace", "name": "functions", "tools": [{"type": "function", "name": "exec"}, {"type": "function", "name": "request_user_input"}]}]}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "Continue"}]}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m0", "content": [{"type": "output_text", "text": "Round."}]}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "response_item", "payload": {"type": "function_call", "name": "request_user_input", "call_id": "q1", "arguments": "{\"questions\": [{\"id\": \"a\", \"header\": \"A\", \"question\": \"Scope?\", \"options\": [{\"label\": \"Fix (Recommended)\", \"description\": \"x\"}, {\"label\": \"Limit\", \"description\": \"y\"}]}, {\"id\": \"b\", \"header\": \"B\", \"question\": \"Review?\", \"options\": [{\"label\": \"Agent (recomendado)\", \"description\": \"x\"}, {\"label\": \"Native\", \"description\": \"y\"}]}, {\"id\": \"c\", \"header\": \"C\", \"question\": \"Security?\", \"options\": [{\"label\": \"Include (Recommended)\", \"description\": \"x\"}, {\"label\": \"General\", \"description\": \"y\"}]}]}"}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "function_call_output", "call_id": "q1", "output": "{\"answers\":{}}"}}
{"timestamp": "2026-01-01T11:00:09.000Z", "type": "event_msg", "payload": {"type": "task_complete", "turn_id": "t1"}}
`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep := auditTranscript(t, tc.host, auditOptions{}, tc.transcript)
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
	if err := auditSessionMain(writeSession(t, "claude", `
{"type": "user", "message": {"role": "user", "content": "Continue"}, "timestamp": "2026-01-01T11:00:01.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "assistant", "message": {"id": "msg_0", "model": "claude-example-1", "role": "assistant", "content": [{"type": "tool_use", "id": "q0", "name": "AskUserQuestion", "input": {"questions": [{"question": "D1 - Which verdict?", "header": "D1", "multiSelect": false, "options": [{"label": "A - (★) Judgment", "description": "x"}, {"label": "B - Mode", "description": "y"}]}, {"question": "Delivery - Where does it stop?", "header": "Delivery", "multiSelect": false, "options": [{"label": "A - (★) interactive", "description": "x"}, {"label": "B - automatic", "description": "y"}]}]}}]}, "timestamp": "2026-01-01T11:00:02.000Z", "cwd": "/work/project", "isSidechain": false}
{"type": "user", "message": {"role": "user", "content": [{"type": "tool_result", "tool_use_id": "q0", "content": "Your questions have been answered: \"D1 - Which verdict?\"=\"A - (★) Judgment\", \"Delivery - Where does it stop?\"=\"B - automatic\". You can now continue with these answers in mind.", "is_error": false}]}, "timestamp": "2026-01-01T11:00:03.000Z", "cwd": "/work/project", "isSidechain": false}
`), "claude", auditOptions{}, &out); err != nil {
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
	cases := []struct {
		name, host, file string
		files            map[string]string
		want             mainThreadCost
	}{
		{"claude counts a message once and skips sidechain lines", "claude", "claude.jsonl", map[string]string{"claude.jsonl": claudeCostSession}, mainThreadCost{
			ModelCalls: 2, InputTokens: 480, CachedTokens: 350, PeakContext: 320, UsageSource: "per_call", Compactions: 1,
			ToolCalls: 4, HiveReads: 1, HiveScripts: 1, ByTool: map[string]int{"Read": 3, "Bash": 1}}},
		{"codex counts a token_count only when the total changed", "codex", "codex.jsonl", map[string]string{"codex.jsonl": codexCostSession}, mainThreadCost{
			ModelCalls: 2, InputTokens: 350, CachedTokens: 140, PeakContext: 250, UsageSource: "per_call", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"exec": 4, "exec_command": 1}, RepeatedUsageEvents: 1}},
		{"grok subtracts a child found in another project folder", "grok", "grok-sessions/%2Frepo/parent-ok/updates.jsonl", grokCostSessions, mainThreadCost{
			ModelCalls: 5, InputTokens: 2700, CachedTokens: 800, UsageSource: "per_turn", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"read_file": 2, "run_terminal_command": 3}}},
		{"grok keeps the parent total when a child's usage is missing", "grok", "grok-sessions/%2Frepo/parent-missing/updates.jsonl", grokCostSessions, mainThreadCost{
			ModelCalls: 7, InputTokens: 3000, CachedTokens: 900, UsageSource: "per_turn_with_children", Compactions: 1,
			ToolCalls: 5, HiveReads: 2, HiveScripts: 1, ByTool: map[string]int{"read_file": 2, "run_terminal_command": 3}}},
		{"grok without usage reports none", "grok", "grok-sessions/%2Frepo/parent-none/updates.jsonl", grokCostSessions, mainThreadCost{
			UsageSource: "none", ByTool: map[string]int{}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := writeFiles(t, tc.files)
			opts := auditOptions{GrokSessionsRoot: filepath.Join(root, "grok-sessions")}
			rep, err := auditSessionFile(filepath.Join(root, filepath.FromSlash(tc.file)), tc.host, opts)
			if err != nil {
				t.Fatal(err)
			}
			gj, _ := json.Marshal(rep.MainThreadCost)
			wj, _ := json.Marshal(tc.want)
			if string(gj) != string(wj) {
				t.Fatalf("main_thread_cost = %s\nwant %s", gj, wj)
			}
		})
	}
}

func TestAuditMainThreadCostLeavesTheCriteriaAlone(t *testing.T) {
	var out bytes.Buffer
	if err := auditSessionMain(writeSession(t, "claude", claudeCostSession), "claude", auditOptions{}, &out); err != nil {
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
		{"python3 scripts/sync_labels.py --apply", false},
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

func TestAuditCodexCompletedItemCarriesExitCodeAndPatchPaths(t *testing.T) {
	s := loadSession(t, "codex", auditOptions{}, `
{"timestamp": "2026-01-01T11:00:01.000Z", "type": "session_meta", "payload": {"id": "00000000-0000-0000-0000-0000000000cf", "cwd": "/work/code", "source": "cli"}}
{"timestamp": "2026-01-01T11:00:02.000Z", "type": "turn_context", "payload": {"model": "gpt-example", "cwd": "/work/code"}}
{"timestamp": "2026-01-01T11:00:03.000Z", "type": "event_msg", "payload": {"type": "task_started", "turn_id": "t1"}}
{"timestamp": "2026-01-01T11:00:04.000Z", "type": "response_item", "payload": {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "build DEMO-1"}]}}
{"timestamp": "2026-01-01T11:00:05.000Z", "type": "event_msg", "payload": {"type": "item_completed", "thread_id": "th", "turn_id": "t1", "item": {"type": "FileChange", "id": "p4", "changes": {"/work/specs/openspec/changes/demo-change/design.md": {"type": "add", "content": "x"}}}}}
{"timestamp": "2026-01-01T11:00:06.000Z", "type": "event_msg", "payload": {"type": "item_completed", "thread_id": "th", "turn_id": "t1", "item": {"type": "CommandExecution", "id": "c5", "command": ["/bin/zsh", "-lc", "cd /work/code && gh pr merge 7 --merge --match-head-commit abc123"], "status": "completed", "exit_code": 0, "aggregated_output": ""}}}
{"timestamp": "2026-01-01T11:00:07.000Z", "type": "event_msg", "payload": {"type": "item_completed", "thread_id": "th", "turn_id": "t1", "item": {"type": "CommandExecution", "id": "c6", "command": ["/bin/zsh", "-lc", "python3 -I /skills/flow-plan/scripts/tasks.py /work/specs/openspec/changes/demo-change versioned --work-repo /work/code"], "status": "failed", "exit_code": 1, "aggregated_output": ""}}}
{"timestamp": "2026-01-01T11:00:08.000Z", "type": "response_item", "payload": {"type": "message", "role": "assistant", "id": "m7", "content": [{"type": "output_text", "text": "Closed. The pull request is merged."}]}}
`)
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

const claudeAdapterSession = `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Please check the project."}}
{"type":"attachment","timestamp":"2026-01-01T10:00:02.000Z","cwd":"project"}
{"type":"assistant","timestamp":"2026-01-01T10:00:03.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Checking."},{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"git status --short && echo sentinel-command-text","description":"example"}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:04.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t1","content":"ok","is_error":false}]}}
{"type":"assistant","timestamp":"2026-01-01T10:00:05.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"One question."},{"type":"tool_use","id":"t2","name":"AskUserQuestion","input":{"questions":[{"question":"D1 — Which delivery mode?","header":"Delivery","options":[{"label":"A - (★) Interactive","description":"x"},{"label":"B - Automatic","description":"y"}],"multiSelect":false}]}}]}}
{"type":"user","timestamp":"2026-01-01T10:00:06.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"t2","content":"User has answered your questions: D1=A","is_error":false}]}}
{"type":"user","timestamp":"2026-01-01T10:00:07.000Z","cwd":"project","isSidechain":false,"isMeta":true,"message":{"role":"user","content":[{"type":"text","text":"meta note that is not a turn"}]}}
{"type":"mode","timestamp":"2026-01-01T10:00:08.000Z","cwd":"project"}
{"type":"user","timestamp":"2026-01-01T10:00:09.000Z","cwd":"project","isSidechain":false,"message":{"role":"user","content":"Next request."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:10.000Z","cwd":"project","isSidechain":false,"message":{"id":"msg_3","model":"claude-example-1","role":"assistant","content":[{"type":"text","text":"Done."}]}}
`

const codexAdapterRollout = `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"00000000-0000-0000-0000-000000000001","cwd":"project","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"project"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t2"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_1","input":"text(await tools.exec_command({cmd:\"git status --short && echo \\\"sentinel-double\\\"\",yield_time_ms:1000}));\ntext(await tools.exec_command({cmd:'echo sentinel-single',yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_1","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"chunk_id\": \"1\", \"exit_code\": 0, \"output\": \"o\", \"wall_time_seconds\": 0.1}"},{"type":"input_text","text":"{\"chunk_id\": \"2\", \"exit_code\": 1, \"output\": \"o\", \"wall_time_seconds\": 0.1}"}]}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_2","input":"text(await tools.exec_command({cmd:\u0060git add notes.txt && echo \"sentinel-backtick\"\u0060,yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:07.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_2","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"chunk_id\": \"3\", \"exit_code\": 0, \"output\": \"o\", \"wall_time_seconds\": 0.1}"}]}}
{"timestamp":"2026-01-01T11:00:08.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_3","input":"const name='x';\ntext(await tools.exec_command({cmd:\u0060echo ${name}\u0060,yield_time_ms:1000}));"}}
{"timestamp":"2026-01-01T11:00:09.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_3","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"chunk_id\": \"4\", \"exit_code\": 0, \"output\": \"o\", \"wall_time_seconds\": 0.1}"}]}}
{"timestamp":"2026-01-01T11:00:10.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_4","input":"tools.write_stdin({session_id:1,chars:\"\"});"}}
{"timestamp":"2026-01-01T11:00:11.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_4","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{}"}]}}
{"timestamp":"2026-01-01T11:00:12.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_5","input":"text(await tools.apply_patch(\"*** Begin Patch\\n*** Update File: src/a.txt\\n@@\\n-old\\n+new\\n*** Add File: docs/b.md\\n+hello\\n*** Delete File: old.txt\\n*** End Patch\"));"}}
{"timestamp":"2026-01-01T11:00:13.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_5","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{}"}]}}
{"timestamp":"2026-01-01T11:00:14.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"call_6","input":"text(await tools.mcp__engram__mem_search({query:\"example\",project:\"p\"}));"}}
{"timestamp":"2026-01-01T11:00:15.000Z","type":"response_item","payload":{"type":"custom_tool_call_output","call_id":"call_6","output":[{"type":"input_text","text":"Script completed\nWall time 0.1 seconds"},{"type":"input_text","text":"{\"content\":[]}"}]}}
{"timestamp":"2026-01-01T11:00:16.000Z","type":"response_item","payload":{"type":"function_call","name":"request_user_input","call_id":"call_7","arguments":"{\"questions\": [{\"id\": \"d1\", \"header\": \"Delivery\", \"question\": \"D1 \\u2014 Which delivery mode?\", \"options\": [{\"label\": \"A - (\\u2605) Interactive\", \"description\": \"x\"}, {\"label\": \"B - Automatic\", \"description\": \"y\"}]}]}"}}
{"timestamp":"2026-01-01T11:00:17.000Z","type":"response_item","payload":{"type":"function_call_output","call_id":"call_7","output":"{\"answers\":{\"d1\":{\"answers\":[\"A\"]}}}"}}
{"timestamp":"2026-01-01T11:00:18.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"msg8","phase":"final_answer","content":[{"type":"output_text","text":"Round answered."}]}}
{"timestamp":"2026-01-01T11:00:19.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t18"}}
{"timestamp":"2026-01-01T11:00:20.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t19"}}
{"timestamp":"2026-01-01T11:00:21.000Z","type":"response_item","payload":{"type":"message","role":"assistant","id":"msg9","phase":"final_answer","content":[{"type":"output_text","text":"Second turn."}]}}
{"timestamp":"2026-01-01T11:00:22.000Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"t21"}}
`

const grokAdapterUpdates = `
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Please check."}}},"timestamp":1794948001,"timestamp_us":1794948001000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call","toolCallId":"call-1","title":"run_terminal_command","rawInput":{"command":"git status --short && echo sentinel-command-text","description":"example"},"_meta":{"x.ai/tool":{"version":1,"name":"run_terminal_command","kind":"execute","namespace":"grok_build"}}}},"timestamp":1794948002,"timestamp_us":1794948002000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call_update","toolCallId":"call-1","status":"completed","rawOutput":{"type":"Terminal","command":"git status --short && echo sentinel-command-text","exit_code":0,"output":"example output","timed_out":false}}},"timestamp":1794948003,"timestamp_us":1794948003000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call","toolCallId":"call-2","title":"run_terminal_command","rawInput":{"command":"git branch -d example-branch","description":"example"},"_meta":{"x.ai/tool":{"version":1,"name":"run_terminal_command","kind":"execute","namespace":"grok_build"}}}},"timestamp":1794948004,"timestamp_us":1794948004000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call_update","toolCallId":"call-2","status":"completed","rawOutput":{"type":"Terminal","command":"git branch -d example-branch","exit_code":1,"output":"example output","timed_out":false}}},"timestamp":1794948005,"timestamp_us":1794948005000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call","toolCallId":"call-3","title":"ask_user_question","rawInput":{"questions":[{"question":"D1 — Which delivery mode?","options":[{"label":"A - (★) Interactive"},{"label":"B - Automatic"}]}]},"_meta":{"x.ai/tool":{"version":1,"name":"ask_user_question","kind":"other","namespace":"grok_build"}}}},"timestamp":1794948006,"timestamp_us":1794948006000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call_update","toolCallId":"call-3","status":"completed","rawOutput":{"UserAnswered":{"message":"User has answered your questions: \"D1\"=\"A\"","answers":{"D1":"A"}},"type":"AskUserQuestion"}}},"timestamp":1794948007,"timestamp_us":1794948007000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"tool_call","toolCallId":"call-4","title":"use_tool","rawInput":{"tool_name":"engram__mem_save","tool_input":{"title":"x"}},"_meta":{"x.ai/tool":{"version":1,"name":"use_tool","kind":"use_tool","namespace":"grok_build"}}}},"timestamp":1794948008,"timestamp_us":1794948008000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Hello there, th"}}},"timestamp":1794948009,"timestamp_us":1794948009000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"is is a message"}}},"timestamp":1794948010,"timestamp_us":1794948010000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":" split in chunk"}}},"timestamp":1794948011,"timestamp_us":1794948011000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"s."}}},"timestamp":1794948012,"timestamp_us":1794948012000000}
{"method":"_x.ai/session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"turn_completed","prompt_id":"p1","stop_reason":"end_turn","elapsed_ms":1}},"timestamp":1794948013,"timestamp_us":1794948013000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Second prompt."}}},"timestamp":1794948014,"timestamp_us":1794948014000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Sec"}}},"timestamp":1794948015,"timestamp_us":1794948015000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"ond"}}},"timestamp":1794948016,"timestamp_us":1794948016000000}
{"method":"session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"."}}},"timestamp":1794948017,"timestamp_us":1794948017000000}
{"method":"_x.ai/session/update","params":{"sessionId":"00000000-0000-0000-0000-000000000002","update":{"sessionUpdate":"turn_completed","prompt_id":"p1","stop_reason":"end_turn","elapsed_ms":1}},"timestamp":1794948018,"timestamp_us":1794948018000000}
`

const claudeCostSession = `
{"type":"user","timestamp":"2026-01-01T10:00:01.000Z","cwd":"/repo","isSidechain":false,"message":{"role":"user","content":"Go."}}
{"type":"assistant","timestamp":"2026-01-01T10:00:02.000Z","cwd":"/repo","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"t1","name":"Read","input":{"file_path":"/home/u/.agents/skills/hive-rule-tests/SKILL.md"}}],"usage":{"input_tokens":100,"cache_read_input_tokens":50,"cache_creation_input_tokens":10,"output_tokens":5}}}
{"type":"assistant","timestamp":"2026-01-01T10:00:03.000Z","cwd":"/repo","isSidechain":false,"message":{"id":"msg_1","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"t2","name":"Bash","input":{"command":"python3 -I /home/u/.agents/skills/flow-close/scripts/cleanup_branches.py --apply"}}],"usage":{"input_tokens":100,"cache_read_input_tokens":50,"cache_creation_input_tokens":10,"output_tokens":5}}}
{"type":"system","subtype":"compact_boundary","timestamp":"2026-01-01T10:00:04.000Z","cwd":"/repo"}
{"type":"assistant","timestamp":"2026-01-01T10:00:05.000Z","cwd":"/repo","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"t3","name":"Read","input":{"file_path":"/repo/.claude/worktrees/w/content/skills/x/SKILL.md"}}],"usage":{"input_tokens":20,"cache_read_input_tokens":300,"cache_creation_input_tokens":0,"output_tokens":5}}}
{"type":"assistant","timestamp":"2026-01-01T10:00:06.000Z","cwd":"/repo","isSidechain":false,"message":{"id":"msg_2","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"t4","name":"Read","input":{"file_path":"/repo/README.md"}}],"usage":{"input_tokens":20,"cache_read_input_tokens":300,"cache_creation_input_tokens":0,"output_tokens":5}}}
{"type":"assistant","timestamp":"2026-01-01T10:00:07.000Z","cwd":"/repo","isSidechain":true,"message":{"id":"msg_side","model":"claude-example-1","role":"assistant","content":[{"type":"tool_use","id":"t5","name":"Read","input":{"file_path":"/home/u/.agents/skills/hive-rule-tests/SKILL.md"}}],"usage":{"input_tokens":9999,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}
`

const codexCostSession = `
{"timestamp":"2026-01-01T11:00:01.000Z","type":"session_meta","payload":{"id":"c1","cwd":"/repo","source":"cli"}}
{"timestamp":"2026-01-01T11:00:02.000Z","type":"turn_context","payload":{"model":"gpt-example","cwd":"/repo"}}
{"timestamp":"2026-01-01T11:00:03.000Z","type":"event_msg","payload":{"type":"task_started","turn_id":"t1"}}
{"timestamp":"2026-01-01T11:00:04.000Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":40},"last_token_usage":{"input_tokens":100,"cached_input_tokens":40}}}}
{"timestamp":"2026-01-01T11:00:05.000Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":100,"cached_input_tokens":40},"last_token_usage":{"input_tokens":100,"cached_input_tokens":40}}}}
{"timestamp":"2026-01-01T11:00:05.500Z","type":"event_msg","payload":{"type":"token_count","info":null}}
{"timestamp":"2026-01-01T11:00:06.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c1","input":"text(await tools.exec_command({cmd:\"cat /home/u/.agents/skills/hive-rule-tests/SKILL.md\"}));"}}
{"timestamp":"2026-01-01T11:00:07.000Z","type":"compacted","payload":{}}
{"timestamp":"2026-01-01T11:00:08.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c2","input":"text(await tools.exec_command({cmd:\"python3 -I /home/u/.agents/skills/flow-close/scripts/cleanup_branches.py --apply\"}));"}}
{"timestamp":"2026-01-01T11:00:09.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c3","input":"text(await tools.exec_command({cmd:\"cat /repo/.claude/worktrees/w/content/skills/x/SKILL.md\"}));"}}
{"timestamp":"2026-01-01T11:00:10.000Z","type":"response_item","payload":{"type":"custom_tool_call","name":"exec","call_id":"c4","input":"text(await tools.exec_command({cmd:\"cat README.md\"}));"}}
{"timestamp":"2026-01-01T11:00:10.500Z","type":"response_item","payload":{"type":"function_call","name":"exec_command","call_id":"c5","arguments":"{\"cmd\":[\"cat\",\"/home/u/.agents/skills/flow-build/SKILL.md\"]}"}}
{"timestamp":"2026-01-01T11:00:11.000Z","type":"event_msg","payload":{"type":"token_count","info":{"total_token_usage":{"input_tokens":350,"cached_input_tokens":140},"last_token_usage":{"input_tokens":250,"cached_input_tokens":100}}}}
`

var grokCostSessions = map[string]string{
	"grok-sessions/%2Fother/child-a/usage.json": `
{"sessionId": "child-a", "session": {"modelCalls": 2, "inputTokens": 300, "cachedReadTokens": 100}}`,
	"grok-sessions/%2Frepo/parent-missing/subagents/child-a/meta.json": `
{"parent_session_id": "parent-missing", "child_session_id": "child-a"}`,
	"grok-sessions/%2Frepo/parent-missing/subagents/child-gone/meta.json": `
{"parent_session_id": "parent-missing", "child_session_id": "child-gone"}`,
	"grok-sessions/%2Frepo/parent-missing/summary.json": `
{"current_model_id":"grok-example-1"}`,
	"grok-sessions/%2Frepo/parent-missing/updates.jsonl": `
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Go."}}},"timestamp":1794948001}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"tool_call","toolCallId":"call-1","title":"read_file","rawInput":{"path":"/home/u/.agents/skills/hive-rule-tests/SKILL.md"},"_meta":{"x.ai/tool":{"name":"read_file"}}}},"timestamp":1794948002}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"tool_call","toolCallId":"call-2","title":"run_terminal_command","rawInput":{"command":"python3 -I /home/u/.agents/skills/flow-close/scripts/cleanup_branches.py"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948003}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"tool_call","toolCallId":"call-3","title":"read_file","rawInput":{"path":"/repo/.claude/worktrees/w/content/skills/x/SKILL.md"},"_meta":{"x.ai/tool":{"name":"read_file"}}}},"timestamp":1794948004}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"tool_call","toolCallId":"call-4","title":"run_terminal_command","rawInput":{"command":"sed -n 1,20p /home/u/.agents/skills/flow-close/references/branch-cleanup.md"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948005}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"tool_call","toolCallId":"call-5","title":"run_terminal_command","rawInput":{"command":"cat README.md"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948006}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"turn_completed","stop_reason":"end_turn","usage":{"modelCalls":3,"inputTokens":1000,"cachedReadTokens":400}}},"timestamp":1794948007}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"auto_compact_completed"}},"timestamp":1794948008}
{"method":"session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"More."}}},"timestamp":1794948009}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-missing","update":{"sessionUpdate":"turn_completed","stop_reason":"end_turn","usage":{"modelCalls":4,"inputTokens":2000,"cachedReadTokens":500}}},"timestamp":1794948010}
`,
	"grok-sessions/%2Frepo/parent-none/updates.jsonl": `
{"method":"session/update","params":{"sessionId":"p","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Hi"}}},"timestamp":1794948001}
`,
	"grok-sessions/%2Frepo/parent-ok/subagents/child-a/meta.json": `
{"parent_session_id": "parent-ok", "child_session_id": "child-a"}`,
	"grok-sessions/%2Frepo/parent-ok/summary.json": `
{"current_model_id":"grok-example-1"}`,
	"grok-sessions/%2Frepo/parent-ok/updates.jsonl": `
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"Go."}}},"timestamp":1794948001}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"tool_call","toolCallId":"call-1","title":"read_file","rawInput":{"path":"/home/u/.agents/skills/hive-rule-tests/SKILL.md"},"_meta":{"x.ai/tool":{"name":"read_file"}}}},"timestamp":1794948002}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"tool_call","toolCallId":"call-2","title":"run_terminal_command","rawInput":{"command":"python3 -I /home/u/.agents/skills/flow-close/scripts/cleanup_branches.py"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948003}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"tool_call","toolCallId":"call-3","title":"read_file","rawInput":{"path":"/repo/.claude/worktrees/w/content/skills/x/SKILL.md"},"_meta":{"x.ai/tool":{"name":"read_file"}}}},"timestamp":1794948004}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"tool_call","toolCallId":"call-4","title":"run_terminal_command","rawInput":{"command":"sed -n 1,20p /home/u/.agents/skills/flow-close/references/branch-cleanup.md"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948005}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"tool_call","toolCallId":"call-5","title":"run_terminal_command","rawInput":{"command":"cat README.md"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}},"timestamp":1794948006}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"turn_completed","stop_reason":"end_turn","usage":{"modelCalls":3,"inputTokens":1000,"cachedReadTokens":400}}},"timestamp":1794948007}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"auto_compact_completed"}},"timestamp":1794948008}
{"method":"session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"user_message_chunk","content":{"type":"text","text":"More."}}},"timestamp":1794948009}
{"method":"_x.ai/session/update","params":{"sessionId":"parent-ok","update":{"sessionUpdate":"turn_completed","stop_reason":"end_turn","usage":{"modelCalls":4,"inputTokens":2000,"cachedReadTokens":500}}},"timestamp":1794948010}
`,
}

func TestRoundTitleLineRecognizesCurrentAndOlderTitles(t *testing.T) {
	cases := []struct {
		title string
		want  bool
	}{
		{"❓ **D8** - **¿Qué enfoque sigue el plan?**", true},
		{"❓ **Entrega** - **¿Cómo se entrega este cambio?**", true},
		{"❓ **Revisión de seguridad** - **¿Cómo se revisa la seguridad?**", true},
		{"❓ **Validación** - **¿Validas el cambio?**", true},
		{"❓ **Security review** - **How is security reviewed?**", true},
		{"❓ **Validation** - **Do you validate the change?**", true},
		{"D1 — Which delivery mode?", true},
		{"**Review** — Which review?", true},
		{"**Validación** — Completed", false},
		{"❓ **Entrega** - **Report heading**", false},
	}
	for _, tc := range cases {
		if got := roundTitleLine.MatchString(tc.title); got != tc.want {
			t.Errorf("roundTitleLine(%q) = %v, want %v", tc.title, got, tc.want)
		}
	}
	if !isLooseRound("❓ **Entrega** - ¿interactive o hold? ➡️ interactive: valida antes de publicar; si prefieres hold, dilo") {
		t.Error("a line that recommends with the arrow marker is a loose round")
	}
}
