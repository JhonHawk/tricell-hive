package main

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestNativeTraces(t *testing.T) {
	cases := []struct{ host, trace, model string }{
		{"codex", `{"type":"item.completed","item":{"id":"c1","type":"command_execution","command":"cat README.md","aggregated_output":"hello","exit_code":0,"status":"completed"}}
{"type":"turn.completed","usage":{"input_tokens":12}}`, "not_observed"},
		{"claude", `{"type":"system","subtype":"init","model":"claude-opus-5[1m]","skills":["workspace-conventions"]}
{"type":"assistant","message":{"model":"claude-opus-5[1m]","content":[{"type":"tool_use","name":"Read","id":"c1","input":{"file_path":"README.md"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"c1","content":"hello"}]}}
{"type":"result","subtype":"success","is_error":false,"usage":{"input_tokens":12}}`, "claude-opus-5[1m]"},
		{"grok", `{"type":"system","subtype":"init","model":"grok-4.6","skills":["workspace-conventions"]}
{"type":"assistant","message":{"model":"grok-4.6","content":[{"type":"tool_use","name":"read_file","id":"c1","input":{"target_file":"README.md"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"c1","content":"hello","is_error":false}]}}
{"type":"result","is_error":false,"usage":{"input_tokens":12},"modelUsage":{"grok-4.6":{"modelCalls":1}}}`, "grok-4.6"},
		{"pi", `{"type":"session","id":"s1"}
{"type":"tool_execution_start","toolName":"read","toolCallId":"c1","args":{"path":"README.md"}}
{"type":"tool_execution_end","toolName":"read","toolCallId":"c1","result":{"content":[{"type":"text","text":"hello"}]},"isError":false}
{"type":"message_end","message":{"role":"assistant","provider":"xai","model":"grok-4.6","stopReason":"stop","content":[{"type":"text","text":"done"}],"usage":{"input":12}}}
{"type":"agent_end","messages":[]}`, "grok-4.6"},
		{"opencode", `{"type":"tool_use","part":{"type":"tool","tool":"read","callID":"c1","state":{"status":"completed","input":{"filePath":"README.md"},"output":"hello"}}}
{"type":"step_finish","part":{"reason":"tool-calls","tokens":{"input":10}}}
{"type":"step_finish","part":{"reason":"stop","tokens":{"input":12},"cost":0.01}}`, "not_observed"},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			r := parseTrace(c.host, strings.NewReader(c.trace))
			if got := terminalState(r, 0, ""); got != "completed" {
				t.Fatalf("terminal=%s %+v", got, r)
			}
			if r.Model != c.model {
				t.Fatalf("model=%s want %s", r.Model, c.model)
			}
			if len(r.Events) < 2 || r.Events[0].Success == nil || !*r.Events[0].Success {
				t.Fatalf("tool execution/result not joined: %+v", r.Events)
			}
			if c.host != "codex" && r.Events[0].Path != "README.md" {
				t.Fatalf("native tool path lost: %+v", r.Events[0])
			}
			if len(r.Usage) == 0 {
				t.Fatal("native usage lost")
			}
		})
	}
}

type brokenReader struct{ first bool }

func (r *brokenReader) Read(p []byte) (int, error) {
	if !r.first {
		r.first = true
		return copy(p, "{\"type\":\"turn.completed\"}\n"), nil
	}
	return 0, errors.New("synthetic interrupted stream")
}
func TestTruncationCannotComplete(t *testing.T) {
	for name, input := range map[string]io.Reader{
		"malformed-tail": strings.NewReader("{\"type\":\"turn.completed\"}\n{\"type\":"),
		"scanner-error":  &brokenReader{},
		"oversized-line": strings.NewReader("{\"type\":\"turn.completed\"}\n" + strings.Repeat("x", 16*1024*1024+1)),
	} {
		t.Run(name, func(t *testing.T) {
			if got := terminalState(parseTrace("codex", input), 0, ""); got != "trace_error" {
				t.Fatalf("got %s", got)
			}
		})
	}
}
func TestEmptyAndIntermediateStreamsAreNotVerified(t *testing.T) {
	for _, trace := range []string{"", `{"type":"step_finish","part":{"reason":"tool-calls"}}`} {
		if got := terminalState(parseTrace("opencode", strings.NewReader(trace)), 0, ""); got != "not_verified" {
			t.Fatalf("got %s", got)
		}
	}
}

func TestOpenCodeTextAndZeroExitDoNotProveCompletion(t *testing.T) {
	r := parseTrace("opencode", strings.NewReader(`{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-final","text":"done"}}`))
	if r.SessionID != "ses-1" || r.FinalAssistantID != "msg-final" {
		t.Fatalf("native correlation IDs lost: %+v", r)
	}
	if got := terminalState(r, 0, ""); got != "not_verified" {
		t.Fatalf("text plus successful process exit was treated as completion: %s", got)
	}
}

func TestOpenCodeNativeExportRequiresExactSuccessfulFinalTurn(t *testing.T) {
	r := parseTrace("opencode", strings.NewReader(`{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-final","text":"done"}}`))
	good := `{"info":{"id":"ses-1","outcome":"succeeded","location":{"directory":"/workspace"}},"messages":[{"id":"msg-prior","type":"assistant","finish":"tool-calls"},{"id":"msg-final","type":"assistant","finish":"stop"},{"id":"msg-idle","type":"idle","outcome":"succeeded"}]}`
	evidence, err := parseOpenCodeCompletionExport(r.SessionID, r.FinalAssistantID, "/workspace", strings.NewReader(good))
	if err != nil {
		t.Fatalf("valid native terminal evidence rejected: %v", err)
	}
	r.NativeCompletion = evidence
	if got := terminalState(r, 0, ""); got != "completed" {
		t.Fatalf("native terminal evidence did not complete: %s", got)
	}

	cases := map[string]string{
		"wrong session":           strings.Replace(good, `"id":"ses-1"`, `"id":"ses-other"`, 1),
		"wrong directory":         strings.Replace(good, `"directory":"/workspace"`, `"directory":"/other"`, 1),
		"not final assistant":     strings.Replace(good, `"id":"msg-final"`, `"id":"msg-another"`, 1),
		"unfinished assistant":    strings.Replace(good, `"finish":"stop"`, `"finish":"tool-calls"`, 1),
		"missing successful idle": strings.Replace(good, `,{"id":"msg-idle","type":"idle","outcome":"succeeded"}`, "", 1),
		"failed session":          strings.Replace(good, `"outcome":"succeeded"`, `"outcome":"failed"`, 1),
		"later user message":      strings.Replace(good, `,{"id":"msg-idle","type":"idle","outcome":"succeeded"}`, `,{"id":"msg-later-user","type":"user"},{"id":"msg-idle","type":"idle","outcome":"succeeded"}`, 1),
		"invalid export":          `{invalid`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseOpenCodeCompletionExport(r.SessionID, r.FinalAssistantID, "/workspace", strings.NewReader(raw)); err == nil {
				t.Fatal("invalid native history was accepted as terminal proof")
			}
		})
	}
}

func TestOpenCodeNativeExportFailureStaysUnverified(t *testing.T) {
	r := parseTrace("opencode", strings.NewReader(`{"type":"text","sessionID":"ses-1","part":{"type":"text","messageID":"msg-final","text":"done"}}`))
	if err := verifyOpenCodeCompletion(&r, "/path/that/does/not/exist", "/workspace", nil); err == nil {
		t.Fatal("missing OpenCode binary unexpectedly verified completion")
	}
	if got := terminalState(r, 0, ""); got != "not_verified" {
		t.Fatalf("unavailable native history changed terminal state: %s", got)
	}
	if r.NativeExportError == "" {
		t.Fatal("native export failure was not retained as a sanitized reason")
	}
}

func TestPermissionDenialAndHostError(t *testing.T) {
	trace := `{"type":"tool_execution_start","toolName":"write","toolCallId":"c1","args":{"path":"report.md"}}
{"type":"tool_execution_end","toolCallId":"c1","result":{"content":[{"type":"text","text":"Permission denied by policy"}]},"isError":true}
{"type":"agent_end","messages":[]}`
	if got := terminalState(parseTrace("pi", strings.NewReader(trace)), 0, ""); got != "blocked_permission" {
		t.Fatalf("got %s", got)
	}
	if got := terminalState(parseTrace("grok", strings.NewReader(`{"type":"result","subtype":"error_max_turns","is_error":true}`)), 0, ""); got != "host_error" {
		t.Fatalf("got %s", got)
	}
}
func TestMemoryIsRetainedAsAnObservation(t *testing.T) {
	r := parseTrace("claude", strings.NewReader(`{"type":"assistant","message":{"content":[{"type":"tool_use","id":"m1","name":"mcp__engram__mem_search","input":{}}]}}`))
	if len(r.Events) != 1 || !r.Events[0].Memory {
		t.Fatal("memory event was not retained")
	}
}

func TestSuccessfulReadOfErrorTextIsNotPermissionDenial(t *testing.T) {
	r := parseTrace("pi", strings.NewReader(`{"type":"tool_execution_start","toolName":"read","toolCallId":"r1","args":{"path":"log.txt"}}
{"type":"tool_execution_end","toolCallId":"r1","isError":false,"result":{"content":[{"type":"text","text":"An earlier task said permission denied"}]}}
{"type":"agent_end","messages":[]}`))
	if terminalState(r, 0, "") != "completed" {
		t.Fatal("successful read was mistaken for a denial")
	}
}

func TestGrokNativeShellAndMemoryWrapper(t *testing.T) {
	r := parseTrace("grok", strings.NewReader(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"run_terminal_command","id":"s1","input":{"command":"printf 55"}},{"type":"tool_use","name":"use_tool","id":"m1","input":{"server_name":"engram","tool_name":"mem_save"}}]}}`))
	if len(r.Events) != 2 || r.Events[0].Kind != "shell" || r.Events[0].Command != "printf 55" || !r.Events[1].Memory {
		t.Fatalf("native aliases lost: %+v", r.Events)
	}
}

func TestGrokWrappedMemoryToolResultKeepsSemanticPayloadAndCallID(t *testing.T) {
	trace := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"use_tool","id":"call-memory-1","input":{"tool_name":"engram__mem_search","tool_input":{"query":"navigation"}}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"call-memory-1","content":"{\"type\":\"MCP\",\"tool_name\":\"mem_search\",\"server_name\":\"engram\",\"output\":{\"OkayOutput\":\"{\\\"results\\\":[{\\\"id\\\":7,\\\"title\\\":\\\"Navigation config\\\"}]}\"}}","is_error":false}]}}`
	r := parseTrace("grok", strings.NewReader(trace))
	if len(r.Events) != 2 {
		t.Fatalf("expected a correlated call and result, got %+v", r.Events)
	}
	call, result := r.Events[0], r.Events[1]
	if call.ID != "call-memory-1" || call.Tool != "engram__mem_search" || !call.Memory {
		t.Fatalf("wrapped call details were lost: %+v", call)
	}
	if result.ID != "call-memory-1" || result.Success == nil || !*result.Success || !strings.Contains(result.Text, `"title":"Navigation config"`) {
		t.Fatalf("semantic success result was not unwrapped/correlated: %+v", result)
	}
}

func TestGrokWrappedMemoryToolResultRecognizesSemanticErrorCode(t *testing.T) {
	trace := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"use_tool","id":"call-memory-error","input":{"tool_name":"engram__mem_save","tool_input":{"content":"fixture"}}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"call-memory-error","content":"{\"type\":\"MCP\",\"tool_name\":\"mem_save\",\"server_name\":\"engram\",\"output\":{\"OkayOutput\":\"{\\\"error_code\\\":\\\"unknown_project\\\",\\\"message\\\":\\\"not found\\\"}\"}}","is_error":false}]}}`
	r := parseTrace("grok", strings.NewReader(trace))
	if len(r.Events) != 2 {
		t.Fatalf("expected a correlated call and result, got %+v", r.Events)
	}
	call, result := r.Events[0], r.Events[1]
	if call.ID != "call-memory-error" || call.Tool != "engram__mem_save" || !call.Memory {
		t.Fatalf("wrapped error call details were lost: %+v", call)
	}
	if result.ID != "call-memory-error" || result.Success == nil || *result.Success || !strings.Contains(result.Text, `"error_code":"unknown_project"`) {
		t.Fatalf("semantic error_code was not retained as a failed result: %+v", result)
	}
}

// --- gh-31: question/search kinds, full Input, Role and Message correlation ---

func TestQuestionAndSearchToolKinds(t *testing.T) {
	trace := `{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","name":"AskUserQuestion","id":"q1","input":{"question":"¿Continuar?"}},{"type":"tool_use","name":"Grep","id":"g1","input":{"path":".env","output_mode":"content"}}]}}`
	r := parseTrace("claude", strings.NewReader(trace))
	if len(r.Events) != 2 {
		t.Fatalf("expected two tool events, got %+v", r.Events)
	}
	if r.Events[0].Kind != "question" {
		t.Fatalf("AskUserQuestion was not classified as question: %+v", r.Events[0])
	}
	if r.Events[1].Kind != "search" {
		t.Fatalf("Grep was not classified as search: %+v", r.Events[1])
	}
	if len(r.Events[0].Input) == 0 || !strings.Contains(string(r.Events[0].Input), "Continuar") {
		t.Fatalf("full tool input was not retained: %+v", r.Events[0])
	}
	if len(r.Events[1].Input) == 0 || !strings.Contains(string(r.Events[1].Input), "output_mode") {
		t.Fatalf("full tool input was not retained: %+v", r.Events[1])
	}
}

func TestGrokQuestionAndSearchToolKinds(t *testing.T) {
	trace := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"use_tool","id":"q1","input":{"tool_name":"ask_user_question","tool_input":{"question":"¿Seguro?"}}},{"type":"tool_use","name":"use_tool","id":"g1","input":{"tool_name":"grep","tool_input":{"path":"src"}}}]}}`
	r := parseTrace("grok", strings.NewReader(trace))
	if len(r.Events) != 2 || r.Events[0].Kind != "question" || r.Events[1].Kind != "search" {
		t.Fatalf("Grok native question/search aliases lost: %+v", r.Events)
	}
}

func TestPiQuestionKind(t *testing.T) {
	trace := `{"type":"tool_execution_start","toolName":"ask_user_question","toolCallId":"q1","args":{"question":"¿Confirmas?"}}`
	r := parseTrace("pi", strings.NewReader(trace))
	if len(r.Events) != 1 || r.Events[0].Kind != "question" {
		t.Fatalf("Pi ask_user_question was not classified as question: %+v", r.Events)
	}
}

func TestTraceEventInputIsNeverSerialized(t *testing.T) {
	e := traceEvent{Line: 1, Kind: "question", Input: json.RawMessage(`{"secret":"nope"}`)}
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.ToLower(string(b)), "input") || strings.Contains(string(b), "secret") {
		t.Fatalf("Input leaked into serialized traceEvent: %s", b)
	}
}

func TestClaudeAssistantRoleOnTextAndSharedMessageKey(t *testing.T) {
	trace := `{"type":"assistant","message":{"id":"msg_1","content":[{"type":"text","text":"Antes de preguntar reviso el detalle suficiente en español, con más de cuarenta caracteres."},{"type":"tool_use","name":"AskUserQuestion","id":"q1","input":{"question":"¿Continuar?"}}]}}`
	r := parseTrace("claude", strings.NewReader(trace))
	if len(r.Events) != 2 {
		t.Fatalf("expected text and question events, got %+v", r.Events)
	}
	text, question := r.Events[0], r.Events[1]
	if text.Kind != "text" || text.Role != "assistant" {
		t.Fatalf("assistant text role not set: %+v", text)
	}
	if text.Message == "" || text.Message != question.Message {
		t.Fatalf("text and question from the same line did not share a Message key: %+v %+v", text, question)
	}
}

func TestClaudeSyntheticUserTextIsNotAssistantRole(t *testing.T) {
	trace := `{"type":"user","message":{"content":[{"type":"text","text":"Base directory for this skill: /skills/example\n# Skill: example\nBody"}]}}`
	r := parseTrace("claude", strings.NewReader(trace))
	if len(r.Events) != 1 || r.Events[0].Kind != "text" {
		t.Fatalf("expected a single text event, got %+v", r.Events)
	}
	if r.Events[0].Role == "assistant" {
		t.Fatal("synthetic Claude user text was tagged as assistant role")
	}
}

func TestGrokMessageIDFallsBackToPerAssistantSequence(t *testing.T) {
	trace := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"read_file","id":"c1","input":{"target_file":"a.md"}}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"read_file","id":"c2","input":{"target_file":"b.md"}}]}}`
	r := parseTrace("grok", strings.NewReader(trace))
	if len(r.Events) != 2 {
		t.Fatalf("expected two events, got %+v", r.Events)
	}
	if r.Events[0].Message == "" || r.Events[1].Message == "" || r.Events[0].Message == r.Events[1].Message {
		t.Fatalf("Grok fallback sequence did not distinguish assistant messages: %+v", r.Events)
	}
}

func TestToolResultInheritsCallMessage(t *testing.T) {
	trace := `{"type":"assistant","message":{"id":"msg_1","content":[{"type":"tool_use","name":"Read","id":"c1","input":{"file_path":"a.md"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"c1","content":"hello"}]}}`
	r := parseTrace("claude", strings.NewReader(trace))
	call, result := r.Events[0], r.Events[1]
	if call.Message != "msg_1" || result.Message != "msg_1" {
		t.Fatalf("tool_result did not inherit the call's Message: call=%+v result=%+v", call, result)
	}
}

func TestPiMessageEndAssignsMessageKeyToItsToolCallsAndDistinguishesRole(t *testing.T) {
	// The assistant's message_end reports which tool calls it is about to
	// make before they execute, so it precedes their start/end pair.
	trace := `{"type":"message_end","message":{"role":"assistant","content":[{"type":"text","text":"Reviso el archivo antes de continuar con el resto de la tarea pendiente."},{"type":"toolCall","id":"c1","name":"read","arguments":{}}]}}
{"type":"tool_execution_start","toolName":"read","toolCallId":"c1","args":{"path":"a.md"}}
{"type":"tool_execution_end","toolCallId":"c1","result":{"content":[{"type":"text","text":"hello"}]},"isError":false}
{"type":"message_end","message":{"role":"toolResult","content":[{"type":"text","text":"Resultado interno de cuarenta caracteres o más de longitud total."}]}}`
	r := parseTrace("pi", strings.NewReader(trace))
	var call, result, assistantText, toolResultRoleText *traceEvent
	for i := range r.Events {
		e := &r.Events[i]
		switch {
		case e.Kind == "read" && e.ID == "c1":
			call = e
		case e.Kind == "tool_result" && e.ID == "c1":
			result = e
		case e.Kind == "text" && e.Role == "assistant":
			assistantText = e
		case e.Kind == "text" && e.Role == "toolResult":
			toolResultRoleText = e
		}
	}
	if call == nil || result == nil || assistantText == nil || toolResultRoleText == nil {
		t.Fatalf("expected all four events, got %+v", r.Events)
	}
	if call.Message == "" || call.Message != assistantText.Message {
		t.Fatalf("Pi toolCall backfill did not assign the message_end's key to the call: call=%+v text=%+v", call, assistantText)
	}
	if result.Message != call.Message {
		t.Fatalf("Pi tool_result did not inherit the backfilled call Message: call=%+v result=%+v", call, result)
	}
	if toolResultRoleText.Role == "assistant" {
		t.Fatal("Pi's toolResult-role message_end text was tagged as assistant role")
	}
}
