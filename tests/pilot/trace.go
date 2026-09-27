package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type traceEvent struct {
	Line    int    `json:"line"`
	Kind    string `json:"kind"`
	Tool    string `json:"tool,omitempty"`
	ID      string `json:"id,omitempty"`
	Path    string `json:"path,omitempty"`
	Command string `json:"command,omitempty"`
	Text    string `json:"text,omitempty"`
	// Role distinguishes assistant-authored text from other roles (Claude/Grok
	// synthetic "user" messages, Pi's toolResult-role message_end) so a
	// criterion can require assistant-authored detail specifically.
	Role string `json:"role,omitempty"`
	// Message is the key of the source message that produced this event: for
	// Claude/Grok it is message.id (or a per-assistant-event sequence number
	// when the id is absent), for Pi it is the index of the last observed
	// assistant message_end. finishTool copies it from the originating call
	// onto its tool_result so a criterion can group a message's calls and
	// results without depending on line adjacency.
	Message string `json:",omitempty"`
	Success *bool  `json:"success,omitempty"`
	Memory  bool   `json:"memory,omitempty"`
	// Input is the tool call's full, already-unwrapped (for Grok) argument
	// object. It is deliberately excluded from JSON: run.json/events.json
	// never carry it, and --assess always reparses stdout.jsonl, so nothing
	// depends on this surviving a round trip.
	Input json.RawMessage `json:"-"`
}
type traceReport struct {
	Model, ModelAtInit                        string
	TerminalSeen, HostError, PermissionDenied bool
	DiscoveryPresent                          bool
	SessionID                                 string                      `json:",omitempty"`
	SessionIDConflict                         bool                        `json:",omitempty"`
	FinalAssistantID                          string                      `json:",omitempty"`
	TerminalSource                            string                      `json:",omitempty"`
	NativeCompletion                          *openCodeCompletionEvidence `json:",omitempty"`
	NativeExportError                         string                      `json:",omitempty"`
	ParseError                                string                      `json:",omitempty"`
	MalformedLines                            []int                       `json:",omitempty"`
	Skills                                    []string                    `json:",omitempty"`
	Events                                    []traceEvent
	Usage                                     []map[string]any `json:",omitempty"`
}

type openCodeCompletionEvidence struct {
	SessionID        string `json:"session_id"`
	Directory        string `json:"directory"`
	FinalAssistantID string `json:"final_assistant_id"`
	AssistantFinish  string `json:"assistant_finish"`
	SessionOutcome   string `json:"session_outcome"`
	FinalEvent       string `json:"final_event"`
}

const openCodeExportLimit = 32 * 1024 * 1024

type limitedBuffer struct {
	bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.limit-b.Len() {
		return 0, fmt.Errorf("native session export exceeds %d bytes", b.limit)
	}
	return b.Buffer.Write(p)
}

// verifyOpenCodeCompletion checks the exact session recorded in the CLI JSONL
// against OpenCode's read-only native export. The transcript is parsed in
// memory and only the bounded proof metadata is retained in traceReport.
func verifyOpenCodeCompletion(r *traceReport, binary, directory string, env []string) error {
	if r == nil {
		return fmt.Errorf("missing trace report")
	}
	r.NativeCompletion = nil
	r.NativeExportError = ""
	if r.SessionID == "" || r.SessionIDConflict {
		return setNativeExportError(r, "session ID missing or inconsistent")
	}
	if r.FinalAssistantID == "" {
		return setNativeExportError(r, "final assistant message ID not observed")
	}
	if binary == "" {
		return setNativeExportError(r, "OpenCode binary unavailable")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "session", "export", "--standalone", r.SessionID)
	cmd.Dir = directory
	if env != nil {
		cmd.Env = env
	}
	var stdout limitedBuffer
	stdout.limit = openCodeExportLimit
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return setNativeExportError(r, "native session export timed out")
		}
		return setNativeExportError(r, "native session export command failed")
	}
	evidence, err := parseOpenCodeCompletionExport(r.SessionID, r.FinalAssistantID, directory, bytes.NewReader(stdout.Bytes()))
	if err != nil {
		return setNativeExportError(r, err.Error())
	}
	r.NativeCompletion = evidence
	r.TerminalSource = "native_export"
	return nil
}

func setNativeExportError(r *traceReport, message string) error {
	r.NativeExportError = message
	return fmt.Errorf("OpenCode native completion unverified: %s", message)
}

func parseOpenCodeCompletionExport(sessionID, finalAssistantID, directory string, input io.Reader) (*openCodeCompletionEvidence, error) {
	data, err := io.ReadAll(io.LimitReader(input, openCodeExportLimit+1))
	if err != nil {
		return nil, fmt.Errorf("native session export could not be read")
	}
	if len(data) > openCodeExportLimit {
		return nil, fmt.Errorf("native session export exceeds %d bytes", openCodeExportLimit)
	}
	var exported struct {
		Info struct {
			ID       string `json:"id"`
			Outcome  string `json:"outcome"`
			Location struct {
				Directory string `json:"directory"`
			} `json:"location"`
		} `json:"info"`
		Messages []struct {
			ID      string `json:"id"`
			Type    string `json:"type"`
			Finish  string `json:"finish"`
			Outcome string `json:"outcome"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &exported); err != nil {
		return nil, fmt.Errorf("native session export is invalid JSON")
	}
	if exported.Info.ID != sessionID {
		return nil, fmt.Errorf("native session ID does not match trace")
	}
	if !sameAbsolutePath(exported.Info.Location.Directory, directory) {
		return nil, fmt.Errorf("native session directory does not match run directory")
	}
	var lastAssistant struct {
		ID     string
		Finish string
	}
	var lastConversationMessageID string
	var lastConversationMessageType string
	var finalEvent string
	var finalOutcome string
	for _, m := range exported.Messages {
		finalEvent = m.Type
		finalOutcome = m.Outcome
		if m.Type != "idle" {
			lastConversationMessageID = m.ID
			lastConversationMessageType = m.Type
		}
		if m.Type == "assistant" {
			lastAssistant.ID = m.ID
			lastAssistant.Finish = m.Finish
		}
	}
	if lastAssistant.ID != finalAssistantID || lastConversationMessageType != "assistant" || lastConversationMessageID != finalAssistantID {
		return nil, fmt.Errorf("native final assistant message does not match trace")
	}
	if lastAssistant.Finish != "stop" {
		return nil, fmt.Errorf("native final assistant finish is not stop")
	}
	if len(exported.Messages) == 0 || finalEvent != "idle" || finalOutcome != "succeeded" || exported.Info.Outcome != "succeeded" {
		return nil, fmt.Errorf("native session did not end successfully")
	}
	return &openCodeCompletionEvidence{
		SessionID:        exported.Info.ID,
		Directory:        filepath.Clean(exported.Info.Location.Directory),
		FinalAssistantID: lastAssistant.ID,
		AssistantFinish:  lastAssistant.Finish,
		SessionOutcome:   exported.Info.Outcome,
		FinalEvent:       finalEvent,
	}, nil
}

func sameAbsolutePath(actual, expected string) bool {
	if !filepath.IsAbs(actual) || !filepath.IsAbs(expected) {
		return false
	}
	return filepath.Clean(actual) == filepath.Clean(expected)
}

func object(v any) map[string]any { m, _ := v.(map[string]any); return m }
func str(v any) string            { s, _ := v.(string); return s }
func list(v any) []any            { a, _ := v.([]any); return a }
func truth(v any) bool            { b, _ := v.(bool); return b }
func first(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s := str(m[k]); s != "" {
			return s
		}
	}
	return ""
}
func textContent(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if m, ok := v.(map[string]any); ok {
		if s := str(m["text"]); s != "" {
			return s
		}
		return textContent(m["content"])
	}
	parts := []string{}
	for _, c := range list(v) {
		if s := textContent(c); s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n")
}
func permissionDenied(s string) bool {
	s = strings.ToLower(s)
	for _, p := range []string{"permission denied", "permission_denied", "permission was denied", "permission request rejected", "user rejected", "requires approval", "approval required", "not allowed by", "auto mode blocked", "denied by", "permission error"} {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
func toolKind(name string) string {
	switch strings.ToLower(name) {
	case "read", "read_file":
		return "read"
	case "write", "write_file":
		return "write"
	case "edit", "search_replace", "multiedit", "apply_patch":
		return "write"
	case "bash", "run_terminal_cmd", "run_terminal_command", "shell", "command_execution", "powershell":
		return "shell"
	case "skill":
		return "skill_invocation"
	case "askuserquestion", "ask_user_question", "question":
		return "question"
	case "grep":
		return "search"
	}
	return "tool"
}

func grokToolCall(name string, args map[string]any) (string, map[string]any) {
	if name != "use_tool" {
		return name, args
	}
	if toolName := str(args["tool_name"]); toolName != "" {
		if toolInput := object(args["tool_input"]); toolInput != nil {
			return toolName, toolInput
		}
		return toolName, map[string]any{}
	}
	return name, args
}

func grokToolResult(v any) (string, bool) {
	var wrapper map[string]any
	switch x := v.(type) {
	case map[string]any:
		wrapper = x
	case string:
		if json.Unmarshal([]byte(x), &wrapper) != nil {
			return x, false
		}
	default:
		return textContent(v), false
	}
	if str(wrapper["type"]) != "MCP" {
		return textContent(v), false
	}
	output := object(wrapper["output"])
	okay := str(output["OkayOutput"])
	if okay == "" {
		return textContent(output), false
	}
	var semantic any
	if json.Unmarshal([]byte(okay), &semantic) != nil {
		return okay, false
	}
	encoded, err := json.Marshal(semantic)
	if err != nil {
		return okay, false
	}
	payload := object(semantic)
	failed := str(payload["error_code"]) != ""
	return string(encoded), failed
}

func parseTrace(host string, input io.Reader) traceReport {
	r := traceReport{Model: "not_observed", ModelAtInit: "not_observed"}
	calls := map[string]int{}
	// assistantSeq is the per-assistant-event fallback for Claude/Grok's
	// Message key when message.id is absent (Grok's id field is unverified).
	assistantSeq := 0
	// Pi has no message.id. currentPiMessage is the key of the last observed
	// assistant message_end; piMessageIndex counts those to produce it.
	// pendingPiToolMessage maps a not-yet-started tool call's ID (read from
	// the message_end's own toolCall list) to that message's key, since the
	// message_end reporting a turn's tool calls precedes their execution.
	piMessageIndex := 0
	currentPiMessage := ""
	pendingPiToolMessage := map[string]string{}
	addTool := func(line int, name, id string, args map[string]any, messageKey string) {
		e := traceEvent{Line: line, Kind: toolKind(name), Tool: name, ID: id, Path: first(args, "file_path", "filePath", "path", "target_file", "target_directory"), Command: first(args, "command", "cmd"), Message: messageKey}
		if encoded, err := json.Marshal(args); err == nil {
			e.Input = json.RawMessage(encoded)
		}
		e.Memory = strings.Contains(strings.ToLower(name), "engram") || strings.Contains(strings.ToLower(name), "memory") || strings.HasPrefix(strings.ToLower(name), "mem_")
		if name == "use_tool" {
			toolName := strings.ToLower(str(args["tool_name"]))
			e.Memory = e.Memory || strings.Contains(strings.ToLower(str(args["server_name"])), "engram") || strings.Contains(toolName, "engram") || strings.HasPrefix(toolName, "mem_")
		}
		if e.Kind == "skill_invocation" {
			e.Text = first(args, "skill", "name")
		}
		if e.Kind == "write" && e.Path == "" {
			e.Text = first(args, "patch", "patchText", "input")
		}
		r.Events = append(r.Events, e)
		if id != "" {
			calls[id] = len(r.Events) - 1
		}
	}
	finishTool := func(line int, id, output string, failed bool) {
		ok := !failed
		msgKey := ""
		if i, found := calls[id]; found {
			r.Events[i].Success = &ok
			msgKey = r.Events[i].Message
		}
		r.Events = append(r.Events, traceEvent{Line: line, Kind: "tool_result", ID: id, Text: output, Success: &ok, Message: msgKey})
		if failed && permissionDenied(output) {
			r.PermissionDenied = true
			r.Events = append(r.Events, traceEvent{Line: line, Kind: "permission_denied", ID: id})
		}
	}
	message := func(line int, role, messageKey string, m map[string]any) {
		if model := str(m["model"]); model != "" {
			r.Model = model
		}
		if usage := object(m["usage"]); usage != nil {
			r.Usage = append(r.Usage, map[string]any{"line": line, "source": "message", "usage": usage})
		}
		if stop := first(m, "stopReason", "stop_reason"); stop == "error" || stop == "aborted" {
			r.HostError = true
		}
		for _, raw := range list(m["content"]) {
			c := object(raw)
			switch str(c["type"]) {
			case "tool_use":
				name, args := str(c["name"]), object(c["input"])
				if host == "grok" {
					name, args = grokToolCall(name, args)
				}
				addTool(line, name, str(c["id"]), args, messageKey)
			case "tool_result":
				output, failed := textContent(c["content"]), truth(c["is_error"])
				if host == "grok" {
					var semanticFailure bool
					output, semanticFailure = grokToolResult(c["content"])
					failed = failed || semanticFailure
				}
				finishTool(line, str(c["tool_use_id"]), output, failed)
			case "text":
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(c["text"]), Role: role, Message: messageKey})
			}
		}
	}
	scan := bufio.NewScanner(input)
	scan.Buffer(make([]byte, 65536), 16*1024*1024)
	line := 0
	for scan.Scan() {
		line++
		if strings.TrimSpace(scan.Text()) == "" {
			continue
		}
		var ev map[string]any
		if json.Unmarshal(scan.Bytes(), &ev) != nil || ev == nil {
			r.MalformedLines = append(r.MalformedLines, line)
			continue
		}
		typ := str(ev["type"])
		if typ == "system" && str(ev["subtype"]) == "init" {
			_, r.DiscoveryPresent = ev["skills"]
			if model := str(ev["model"]); model != "" {
				r.ModelAtInit = model
			}
			for _, s := range list(ev["skills"]) {
				if name := str(s); name != "" {
					r.Skills = append(r.Skills, name)
				} else if name := str(object(s)["name"]); name != "" {
					r.Skills = append(r.Skills, name)
				}
			}
		}
		if typ == "error" || typ == "turn.failed" {
			r.HostError = true
		}
		if strings.Contains(typ, "permission") {
			encoded, _ := json.Marshal(ev)
			if permissionDenied(string(encoded)) || strings.Contains(typ, "denied") {
				r.PermissionDenied = true
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "permission_denied"})
			}
		}
		switch host {
		case "claude", "grok":
			if typ == "assistant" || typ == "user" {
				m := object(ev["message"])
				key := str(m["id"])
				if typ == "assistant" {
					assistantSeq++
					if key == "" {
						key = fmt.Sprintf("assistant-%d", assistantSeq)
					}
				}
				message(line, typ, key, m)
			}
			// Claude delivers Skill bodies as native synthetic user messages after a
			// successful Skill invocation. A launch acknowledgment alone is not a read.
			if host == "claude" && typ == "user" && truth(ev["isSynthetic"]) {
				for _, raw := range list(object(ev["message"])["content"]) {
					c := object(raw)
					body := str(c["text"])
					const prefix = "Base directory for this skill: "
					if str(c["type"]) != "text" || !strings.HasPrefix(body, prefix) {
						continue
					}
					parts := strings.SplitN(strings.TrimPrefix(body, prefix), "\n", 2)
					if len(parts) != 2 || !strings.HasPrefix(strings.TrimSpace(parts[1]), "# ") || len(strings.TrimSpace(parts[1])) < 100 {
						continue
					}
					directory := strings.TrimSpace(parts[0])
					name := filepath.Base(directory)
					for i := len(r.Events) - 1; i >= 0; i-- {
						call := r.Events[i]
						if call.Kind == "skill_invocation" && call.Text == name && call.Success != nil && *call.Success {
							okay := true
							r.Events = append(r.Events, traceEvent{Line: line, Kind: "read", Tool: "native_skill_source", ID: call.ID, Path: filepath.Join(directory, "SKILL.md"), Success: &okay})
							break
						}
					}
				}
			}
			if typ == "result" {
				r.TerminalSeen = true
				if truth(ev["is_error"]) || strings.HasPrefix(str(ev["subtype"]), "error") {
					r.HostError = true
				}
				if len(list(ev["permission_denials"])) > 0 {
					r.PermissionDenied = true
				}
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "final", Text: str(ev["result"])})
				usage := map[string]any{}
				for _, k := range []string{"usage", "modelUsage", "total_cost_usd", "num_turns", "duration_ms", "duration_api_ms"} {
					if v, ok := ev[k]; ok {
						usage[k] = v
					}
				}
				if len(usage) > 0 {
					r.Usage = append(r.Usage, map[string]any{"line": line, "source": "result", "usage": usage})
				}
			}
		case "codex":
			if typ == "turn.completed" {
				r.TerminalSeen = true
				if u := object(ev["usage"]); u != nil {
					r.Usage = append(r.Usage, map[string]any{"line": line, "source": typ, "usage": u})
				}
			}
			if typ == "item.completed" {
				i := object(ev["item"])
				id := str(i["id"])
				switch str(i["type"]) {
				case "command_execution":
					addTool(line, "command_execution", id, i, "")
					failed := str(i["status"]) == "failed"
					if n, ok := i["exit_code"].(float64); ok && n != 0 {
						failed = true
					}
					finishTool(line, id, str(i["aggregated_output"]), failed)
				case "file_change":
					for _, raw := range list(i["changes"]) {
						c := object(raw)
						kind := "write"
						if str(c["kind"]) == "delete" {
							kind = "delete"
						}
						ok := str(i["status"]) != "failed"
						r.Events = append(r.Events, traceEvent{Line: line, Kind: kind, Tool: "file_change", ID: id, Path: str(c["path"]), Success: &ok})
					}
				case "mcp_tool_call":
					addTool(line, first(i, "tool", "name"), id, object(i["arguments"]), "")
					finishTool(line, id, textContent(i["result"]), str(i["status"]) == "failed")
				case "agent_message":
					// codex exec 0.157.0 has no request_user_input in exec mode, so a
					// Codex close question is only ever observed as assistant text.
					r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(i["text"]), Role: "assistant"})
				}
			}
		case "pi":
			switch typ {
			case "tool_execution_start":
				id := str(ev["toolCallId"])
				addTool(line, str(ev["toolName"]), id, object(ev["args"]), pendingPiToolMessage[id])
			case "tool_execution_end":
				finishTool(line, str(ev["toolCallId"]), textContent(ev["result"]), truth(ev["isError"]))
			case "message_end":
				m := object(ev["message"])
				role := str(m["role"])
				// The assistant's message_end carries the tool calls it is about
				// to make as `toolCall` content blocks (pi-ai ToolCall) before
				// their start/end pairs are observed, so the message key is
				// registered forward for the upcoming addTool calls to pick up.
				if role == "assistant" {
					piMessageIndex++
					currentPiMessage = fmt.Sprintf("pi-%d", piMessageIndex)
					for _, raw := range list(m["content"]) {
						tc := object(raw)
						if str(tc["type"]) != "toolCall" {
							continue
						}
						if tid := str(tc["id"]); tid != "" {
							pendingPiToolMessage[tid] = currentPiMessage
						}
					}
				}
				message(line, role, currentPiMessage, m)
			case "agent_end":
				r.TerminalSeen = true
				for _, m := range list(ev["messages"]) {
					msg := object(m)
					if stop := str(msg["stopReason"]); stop == "error" || stop == "aborted" {
						r.HostError = true
					}
				}
			}
		case "opencode":
			if sessionID := str(ev["sessionID"]); sessionID != "" {
				if r.SessionID == "" {
					r.SessionID = sessionID
				} else if r.SessionID != sessionID {
					r.SessionIDConflict = true
				}
			}
			p := object(ev["part"])
			switch typ {
			case "tool_use":
				s := object(p["state"])
				id := first(p, "callID", "id")
				// The part's own messageID is carried as Message so S1
				// (question_after_detail) and close_question_after_report can
				// group a message's calls/results without line adjacency,
				// matching Claude/Grok/Pi's Message correlation.
				addTool(line, str(p["tool"]), id, object(s["input"]), str(p["messageID"]))
				if status := str(s["status"]); status == "completed" || status == "error" {
					finishTool(line, id, first(s, "output", "error"), status == "error")
				}
			case "text":
				messageID := str(p["messageID"])
				if messageID != "" {
					r.FinalAssistantID = messageID
				}
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(p["text"]), Role: "assistant", Message: messageID})
			case "step_finish":
				if reason := str(p["reason"]); reason == "stop" {
					r.TerminalSeen = true
					r.TerminalSource = "stream"
					if messageID := str(p["messageID"]); messageID != "" {
						r.FinalAssistantID = messageID
					}
				} else if reason == "error" {
					r.HostError = true
				}
				u := map[string]any{}
				for _, k := range []string{"tokens", "cost"} {
					if v, ok := p[k]; ok {
						u[k] = v
					}
				}
				if len(u) > 0 {
					r.Usage = append(r.Usage, map[string]any{"line": line, "source": typ, "usage": u})
				}
			}
		}
	}
	if err := scan.Err(); err != nil {
		r.ParseError = fmt.Sprintf("JSONL scanner: %v", err)
	}
	return r
}

func terminalState(t traceReport, exitCode int, processError string) string {
	if processError == "timeout" {
		return "timeout"
	}
	if t.ParseError != "" || len(t.MalformedLines) > 0 {
		return "trace_error"
	}
	if t.PermissionDenied {
		return "blocked_permission"
	}
	if exitCode != 0 {
		return "process_error"
	}
	if t.HostError {
		return "host_error"
	}
	if t.TerminalSeen {
		return "completed"
	}
	if t.NativeCompletion != nil {
		return "completed"
	}
	return "not_verified"
}
