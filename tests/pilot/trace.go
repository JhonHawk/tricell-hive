package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type traceEvent struct {
	Line    int    `json:"line"`
	Kind    string `json:"kind"`
	Tool    string `json:"tool,omitempty"`
	ID      string `json:"id,omitempty"`
	Path    string `json:"path,omitempty"`
	Command string `json:"command,omitempty"`
	Text    string `json:"text,omitempty"`
	Success *bool  `json:"success,omitempty"`
	Memory  bool   `json:"memory,omitempty"`
}
type traceReport struct {
	Model, ModelAtInit                        string
	TerminalSeen, HostError, PermissionDenied bool
	DiscoveryPresent                          bool
	ParseError                                string   `json:",omitempty"`
	MalformedLines                            []int    `json:",omitempty"`
	Skills                                    []string `json:",omitempty"`
	Events                                    []traceEvent
	Usage                                     []map[string]any `json:",omitempty"`
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
	}
	return "tool"
}

func parseTrace(host string, input io.Reader) traceReport {
	r := traceReport{Model: "not_observed", ModelAtInit: "not_observed"}
	calls := map[string]int{}
	addTool := func(line int, name, id string, args map[string]any) {
		e := traceEvent{Line: line, Kind: toolKind(name), Tool: name, ID: id, Path: first(args, "file_path", "filePath", "path", "target_file", "target_directory"), Command: first(args, "command", "cmd")}
		e.Memory = strings.Contains(strings.ToLower(name), "engram") || strings.Contains(strings.ToLower(name), "memory") || strings.HasPrefix(strings.ToLower(name), "mem_")
		if name == "use_tool" {
			e.Memory = e.Memory || strings.Contains(strings.ToLower(str(args["server_name"])), "engram") || strings.HasPrefix(strings.ToLower(str(args["tool_name"])), "mem_")
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
		if i, found := calls[id]; found {
			r.Events[i].Success = &ok
		}
		r.Events = append(r.Events, traceEvent{Line: line, Kind: "tool_result", ID: id, Text: output, Success: &ok})
		if failed && permissionDenied(output) {
			r.PermissionDenied = true
			r.Events = append(r.Events, traceEvent{Line: line, Kind: "permission_denied", ID: id})
		}
	}
	message := func(line int, m map[string]any) {
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
				addTool(line, str(c["name"]), str(c["id"]), object(c["input"]))
			case "tool_result":
				finishTool(line, str(c["tool_use_id"]), textContent(c["content"]), truth(c["is_error"]))
			case "text":
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(c["text"])})
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
				message(line, object(ev["message"]))
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
					addTool(line, "command_execution", id, i)
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
					addTool(line, first(i, "tool", "name"), id, object(i["arguments"]))
					finishTool(line, id, textContent(i["result"]), str(i["status"]) == "failed")
				case "agent_message":
					r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(i["text"])})
				}
			}
		case "pi":
			switch typ {
			case "tool_execution_start":
				addTool(line, str(ev["toolName"]), str(ev["toolCallId"]), object(ev["args"]))
			case "tool_execution_end":
				finishTool(line, str(ev["toolCallId"]), textContent(ev["result"]), truth(ev["isError"]))
			case "message_end":
				message(line, object(ev["message"]))
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
			p := object(ev["part"])
			switch typ {
			case "tool_use":
				s := object(p["state"])
				id := first(p, "callID", "id")
				addTool(line, str(p["tool"]), id, object(s["input"]))
				if status := str(s["status"]); status == "completed" || status == "error" {
					finishTool(line, id, first(s, "output", "error"), status == "error")
				}
			case "text":
				r.Events = append(r.Events, traceEvent{Line: line, Kind: "text", Text: str(p["text"])})
			case "step_finish":
				if reason := str(p["reason"]); reason == "stop" {
					r.TerminalSeen = true
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
	return "not_verified"
}
