package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Real-session audit: reads a transcript a host already wrote and runs the
// deterministic criteria on it. It launches no model and writes nothing but
// the JSON report on stdout. Pilots never reach this file: parseTrace and the
// pilot criteria are shared, not changed.

type auditOptions struct {
	// Base is the base branch the criteria that need one use (--base).
	Base string
	// Cwd overrides the working directory read from the transcript (--cwd),
	// for a session whose repository has moved since it ran.
	Cwd string
}

// auditSession is a transcript normalized to the pilot trace model. Turn[i] is
// the turn number (1-based; 0 before the first) of Trace.Events[i].
type auditSession struct {
	Host            string
	File            string
	Cwd             string
	Model           string
	ModelConfigured bool
	Subagent        bool
	First, Last     string
	Trace           traceReport
	Turn            []int
}

type auditReport struct {
	Host            string                `json:"host"`
	File            string                `json:"file"`
	Cwd             string                `json:"cwd"`
	Model           string                `json:"model"`
	ModelConfigured bool                  `json:"model_configured,omitempty"`
	Subagent        bool                  `json:"subagent,omitempty"`
	FirstTimestamp  string                `json:"first_timestamp"`
	LastTimestamp   string                `json:"last_timestamp"`
	Criteria        []criterionAssessment `json:"criteria"`
}

const auditTimeLayout = "2006-01-02T15:04:05.000Z"

func validAuditHost(host string) bool { return host == "claude" || host == "codex" || host == "grok" }

func readAuditFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read session file: %w", err)
	}
	return data, nil
}

func auditScanner(data []byte) *bufio.Scanner {
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 65536), 64*1024*1024)
	return scan
}

type timeRange struct{ first, last string }

func (t *timeRange) add(ts string) {
	if ts == "" {
		return
	}
	if t.first == "" || ts < t.first {
		t.first = ts
	}
	if ts > t.last {
		t.last = ts
	}
}

func loadAuditSession(path, host string, opts auditOptions) (auditSession, error) {
	if !validAuditHost(host) {
		return auditSession{}, fmt.Errorf("--host must be claude, codex or grok")
	}
	data, err := readAuditFile(path)
	if err != nil {
		return auditSession{}, err
	}
	var s auditSession
	switch host {
	case "claude":
		s, err = loadClaudeSession(data)
	case "codex":
		s, err = loadCodexSession(data)
	default:
		s, err = loadGrokSession(data, path)
	}
	if err != nil {
		return auditSession{}, err
	}
	s.Host, s.File = host, path
	if opts.Cwd != "" {
		s.Cwd = opts.Cwd
	}
	return s, nil
}

// --- Claude Code project transcript ---

func loadClaudeSession(data []byte) (auditSession, error) {
	s := auditSession{}
	s.Trace = parseTrace("claude", bytes.NewReader(data))
	if s.Trace.ParseError != "" {
		return s, fmt.Errorf("claude transcript: %s", s.Trace.ParseError)
	}
	s.Model = s.Trace.Model
	var starts []int
	var tr timeRange
	line := 0
	scan := auditScanner(data)
	for scan.Scan() {
		line++
		var ev struct {
			Type, Timestamp, Cwd string
			IsMeta, IsSidechain  bool
			Message              struct{ Content json.RawMessage }
		}
		if json.Unmarshal(scan.Bytes(), &ev) != nil {
			continue
		}
		tr.add(ev.Timestamp)
		if s.Cwd == "" {
			s.Cwd = ev.Cwd
		}
		if ev.Type == "user" && !ev.IsMeta && !ev.IsSidechain && claudePromptContent(ev.Message.Content) {
			starts = append(starts, line)
		}
	}
	if err := scan.Err(); err != nil {
		return s, fmt.Errorf("claude transcript: %w", err)
	}
	s.First, s.Last = tr.first, tr.last
	s.Turn = make([]int, len(s.Trace.Events))
	for i, e := range s.Trace.Events {
		for _, st := range starts {
			if st <= e.Line {
				s.Turn[i]++
			}
		}
	}
	return s, nil
}

// claudePromptContent reports whether a user line's content is a typed prompt
// (a string, or blocks that are not all tool results).
func claudePromptContent(raw json.RawMessage) bool {
	var str string
	if json.Unmarshal(raw, &str) == nil {
		return strings.TrimSpace(str) != ""
	}
	var blocks []struct{ Type string }
	if json.Unmarshal(raw, &blocks) != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "text" || b.Type == "image" {
			return true
		}
	}
	return false
}

// --- Codex rollout ---

type codexBuilder struct {
	s      auditSession
	turn   int
	msgSeq int
}

func (b *codexBuilder) add(e traceEvent) {
	b.s.Trace.Events = append(b.s.Trace.Events, e)
	b.s.Turn = append(b.s.Turn, b.turn)
}

func loadCodexSession(data []byte) (auditSession, error) {
	b := &codexBuilder{}
	b.s.Trace = traceReport{Model: "not_observed", ModelAtInit: "not_observed"}
	var tr timeRange
	line := 0
	sawMeta, sawContext := false, false
	scan := auditScanner(data)
	for scan.Scan() {
		line++
		var ev struct {
			Timestamp, Type string
			Payload         map[string]any
		}
		if json.Unmarshal(scan.Bytes(), &ev) != nil {
			b.s.Trace.MalformedLines = append(b.s.Trace.MalformedLines, line)
			continue
		}
		tr.add(ev.Timestamp)
		p := ev.Payload
		switch ev.Type {
		case "session_meta":
			if !sawMeta {
				sawMeta = true
				b.s.Cwd = str(p["cwd"])
				if src := object(p["source"]); src != nil {
					_, b.s.Subagent = src["subagent"]
				}
			}
		case "turn_context":
			if !sawContext {
				sawContext = true
				b.s.Model = str(p["model"])
				b.s.ModelConfigured = b.s.Model != ""
			}
		case "event_msg":
			if str(p["type"]) == "task_started" {
				b.turn++
			}
		case "response_item":
			b.codexItem(line, p)
		}
	}
	if err := scan.Err(); err != nil {
		return b.s, fmt.Errorf("codex rollout: %w", err)
	}
	if b.s.Model == "" {
		b.s.Model = "not_observed"
	}
	b.s.First, b.s.Last = tr.first, tr.last
	return b.s, nil
}

func (b *codexBuilder) codexItem(line int, p map[string]any) {
	switch str(p["type"]) {
	case "message":
		if str(p["role"]) != "assistant" {
			return
		}
		var parts []string
		for _, raw := range list(p["content"]) {
			if t := str(object(raw)["text"]); t != "" {
				parts = append(parts, t)
			}
		}
		b.msgSeq++
		key := str(p["id"])
		if key == "" {
			key = fmt.Sprintf("codex-msg-%d", b.msgSeq)
		}
		b.add(traceEvent{Line: line, Kind: "text", Role: "assistant", Text: strings.Join(parts, "\n"), Message: key})
	case "function_call":
		name, id := str(p["name"]), str(p["call_id"])
		kind := toolKind(name)
		if name == "request_user_input" || name == "request_user_input_async" {
			kind = "question"
		}
		e := traceEvent{Line: line, Kind: kind, Tool: name, ID: id, Message: "codex-call-" + id}
		e.Input = json.RawMessage(str(p["arguments"]))
		if !json.Valid(e.Input) {
			e.Input = nil
		}
		e.Memory = codexMemoryTool(name)
		b.add(e)
	case "function_call_output":
		id := str(p["call_id"])
		ok := true
		b.add(traceEvent{Line: line, Kind: "tool_result", ID: id, Text: textContent(p["output"]), Success: &ok, Message: "codex-call-" + id})
	case "custom_tool_call":
		if str(p["name"]) == "exec" {
			b.codexExec(line, str(p["call_id"]), str(p["input"]))
		}
	case "custom_tool_call_output":
		b.codexExecOutput(line, str(p["call_id"]), list(p["output"]))
	}
}

func codexMemoryTool(name string) bool {
	n := strings.ToLower(name)
	return strings.Contains(n, "engram") || strings.Contains(n, "memory") || strings.HasPrefix(n, "mem_")
}

// codexCall is one tools.<name>(...) call found in an exec block.
type codexCall struct {
	name        string
	literal     string
	hasLiteral  bool
	interpolate bool
}

func (b *codexBuilder) codexExec(line int, id, js string) {
	calls := scanCodexExec(js)
	idx := 0
	for _, c := range calls {
		switch {
		case c.name == "write_stdin":
			continue
		case c.name == "exec_command":
			e := traceEvent{Line: line, Kind: "shell", Tool: "exec_command", ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id}
			if c.hasLiteral && !c.interpolate {
				e.Command = c.literal
			} else {
				e.CommandWithheld = true
			}
			b.add(e)
		case c.name == "apply_patch":
			b.codexPatch(line, id, idx, c)
		default:
			e := traceEvent{Line: line, Kind: "tool", Tool: c.name, ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id}
			e.Memory = codexMemoryTool(c.name)
			b.add(e)
		}
		idx++
	}
}

var patchFile = regexp.MustCompile(`(?m)^\*\*\* (Add|Update|Delete) File: (.+?)\s*$`)

func (b *codexBuilder) codexPatch(line int, id string, idx int, c codexCall) {
	if !c.hasLiteral {
		b.add(traceEvent{Line: line, Kind: "write", Tool: "apply_patch", ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id})
		return
	}
	matches := patchFile.FindAllStringSubmatch(c.literal, -1)
	if len(matches) == 0 {
		b.add(traceEvent{Line: line, Kind: "write", Tool: "apply_patch", ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id})
		return
	}
	for n, m := range matches {
		kind := "write"
		if m[1] == "Delete" {
			kind = "delete"
		}
		b.add(traceEvent{Line: line, Kind: kind, Tool: "apply_patch", ID: fmt.Sprintf("%s#%d.%d", id, idx, n), Path: m[2], Message: "codex-call-" + id})
	}
}

// codexExecOutput pairs a block's output with the events its calls produced.
// A command's success is known only when the block reports exactly one exit
// code per command; otherwise it is known true only when every reported code
// is zero, and unknown (nil) otherwise.
func (b *codexBuilder) codexExecOutput(line int, id string, items []any) {
	type result struct {
		code     int
		hasCode  bool
		rejected bool
		output   string
	}
	var results []result
	header := ""
	for n, raw := range items {
		text := str(object(raw)["text"])
		if n == 0 {
			header = text
			continue
		}
		var v map[string]any
		if json.Unmarshal([]byte(text), &v) != nil || v == nil {
			continue
		}
		if str(v["status"]) == "rejected" {
			results = append(results, result{rejected: true})
			continue
		}
		if val := object(v["value"]); val != nil && str(v["status"]) == "fulfilled" {
			v = val
		}
		if code, ok := v["exit_code"].(float64); ok {
			results = append(results, result{code: int(code), hasCode: true, output: str(v["output"])})
		}
	}
	var execIdx []int
	for i, e := range b.s.Trace.Events {
		if strings.HasPrefix(e.ID, id+"#") && (e.Kind == "shell" || e.Kind == "tool" || e.Kind == "write" || e.Kind == "delete") {
			execIdx = append(execIdx, i)
		}
	}
	shellCount := 0
	for _, i := range execIdx {
		if b.s.Trace.Events[i].Kind == "shell" {
			shellCount++
		}
	}
	allZero := strings.HasPrefix(header, "Script completed")
	for _, r := range results {
		if r.rejected || !r.hasCode || r.code != 0 {
			allZero = false
		}
	}
	paired := len(results) == shellCount
	shellN := 0
	for _, i := range execIdx {
		e := &b.s.Trace.Events[i]
		var success *bool
		text := ""
		switch e.Kind {
		case "shell":
			if paired {
				r := results[shellN]
				ok := !r.rejected && r.hasCode && r.code == 0
				success, text = &ok, r.output
			} else if allZero {
				ok := true
				success = &ok
			}
			shellN++
		default:
			if strings.HasPrefix(header, "Script completed") {
				ok := true
				success = &ok
			}
		}
		e.Success = success
		b.add(traceEvent{Line: line, Kind: "tool_result", ID: e.ID, Text: text, Success: success, Message: e.Message})
	}
}

// scanCodexExec lists the tools.<name>(...) calls of a JavaScript exec block
// in source order. It skips string, template and comment text so a command that
// mentions "tools." is not a call, and extracts the literal of the first
// argument (apply_patch) or of the cmd property (exec_command), accepting
// double quotes, single quotes and backticks. A template literal with ${...}
// interpolation is reported as interpolated.
func scanCodexExec(js string) []codexCall {
	var calls []codexCall
	i := 0
	for i < len(js) {
		c := js[i]
		switch {
		case c == '"' || c == '\'' || c == '`':
			_, end, _ := jsLiteral(js, i)
			i = end
		case c == '/' && i+1 < len(js) && js[i+1] == '/':
			for i < len(js) && js[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(js) && js[i+1] == '*':
			end := strings.Index(js[i+2:], "*/")
			if end < 0 {
				return calls
			}
			i += end + 4
		case strings.HasPrefix(js[i:], "tools.") && (i == 0 || !isJSIdent(js[i-1])):
			j := i + len("tools.")
			k := j
			for k < len(js) && isJSIdent(js[k]) {
				k++
			}
			if k < len(js) && js[k] == '(' && k > j {
				call := codexCall{name: js[j:k]}
				switch call.name {
				case "exec_command":
					call.literal, call.hasLiteral, call.interpolate = jsObjectLiteral(js, k+1, "cmd")
				case "apply_patch":
					call.literal, call.hasLiteral, call.interpolate = jsFirstArgLiteral(js, k+1)
				}
				calls = append(calls, call)
				i = k + 1
				continue
			}
			i = k
		default:
			i++
		}
	}
	return calls
}

func isJSIdent(c byte) bool {
	return c == '_' || c == '$' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// jsLiteral decodes the string or template literal opening at js[i] and
// returns its value, the index after its closing quote and whether a template
// interpolation was seen (the value then stops at the interpolation).
func jsLiteral(js string, i int) (value string, end int, interpolated bool) {
	quote := js[i]
	var sb strings.Builder
	j := i + 1
	for j < len(js) {
		c := js[j]
		switch {
		case c == '\\' && j+1 < len(js):
			j++
			switch js[j] {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '0':
				sb.WriteByte(0)
			case 'u':
				if j+4 < len(js) {
					if n, err := strconv.ParseUint(js[j+1:j+5], 16, 32); err == nil {
						sb.WriteRune(rune(n))
						j += 4
					}
				}
			case 'x':
				if j+2 < len(js) {
					if n, err := strconv.ParseUint(js[j+1:j+3], 16, 32); err == nil {
						sb.WriteRune(rune(n))
						j += 2
					}
				}
			default:
				sb.WriteByte(js[j])
			}
			j++
		case c == quote:
			return sb.String(), j + 1, interpolated
		case quote == '`' && c == '$' && j+1 < len(js) && js[j+1] == '{':
			interpolated = true
			depth := 1
			j += 2
			for j < len(js) && depth > 0 {
				switch js[j] {
				case '{':
					depth++
					j++
				case '}':
					depth--
					j++
				case '"', '\'', '`':
					_, j, _ = jsLiteral(js, j)
				default:
					j++
				}
			}
		default:
			sb.WriteByte(c)
			j++
		}
	}
	return sb.String(), len(js), interpolated
}

// jsObjectLiteral finds the top-level property key in the object literal that
// starts at or after js[i] and returns its string literal value. A property
// that is not a plain literal (a variable, a concatenation) is not literal.
func jsObjectLiteral(js string, i int, key string) (value string, ok, interpolated bool) {
	for i < len(js) && (js[i] == ' ' || js[i] == '\n' || js[i] == '\t') {
		i++
	}
	if i >= len(js) || js[i] != '{' {
		return "", false, false
	}
	depth := 0
	for j := i; j < len(js); {
		c := js[j]
		switch {
		case c == '"' || c == '\'' || c == '`':
			_, j, _ = jsLiteral(js, j)
		case c == '{' || c == '[' || c == '(':
			depth++
			j++
		case c == '}' || c == ']' || c == ')':
			depth--
			if depth <= 0 {
				return "", false, false
			}
			j++
		case depth == 1 && strings.HasPrefix(js[j:], key) && (j == 0 || !isJSIdent(js[j-1])) && (j+len(key) >= len(js) || !isJSIdent(js[j+len(key)])):
			k := j + len(key)
			for k < len(js) && (js[k] == ' ' || js[k] == '\n' || js[k] == '\t') {
				k++
			}
			if k >= len(js) || js[k] != ':' {
				return "", false, false
			}
			k++
			for k < len(js) && (js[k] == ' ' || js[k] == '\n' || js[k] == '\t') {
				k++
			}
			if k >= len(js) || (js[k] != '"' && js[k] != '\'' && js[k] != '`') {
				return "", false, false
			}
			v, end, interp := jsLiteral(js, k)
			for end < len(js) && (js[end] == ' ' || js[end] == '\n' || js[end] == '\t') {
				end++
			}
			if end < len(js) && js[end] == '+' {
				return "", false, false
			}
			return v, true, interp
		default:
			j++
		}
	}
	return "", false, false
}

func jsFirstArgLiteral(js string, i int) (value string, ok, interpolated bool) {
	for i < len(js) && (js[i] == ' ' || js[i] == '\n' || js[i] == '\t') {
		i++
	}
	if i >= len(js) || (js[i] != '"' && js[i] != '\'' && js[i] != '`') {
		return "", false, false
	}
	v, end, interp := jsLiteral(js, i)
	for end < len(js) && (js[end] == ' ' || js[end] == '\n' || js[end] == '\t') {
		end++
	}
	if end < len(js) && js[end] == '+' {
		return "", false, false
	}
	return v, true, interp
}

// --- Grok updates.jsonl ---

func loadGrokSession(data []byte, path string) (auditSession, error) {
	s := auditSession{}
	s.Trace = traceReport{Model: "not_observed", ModelAtInit: "not_observed"}
	s.Cwd = grokCwd(path)
	s.Model = "not_observed"
	if raw, err := os.ReadFile(filepath.Join(filepath.Dir(path), "summary.json")); err == nil {
		var sum struct {
			CurrentModelID string `json:"current_model_id"`
		}
		if json.Unmarshal(raw, &sum) == nil && sum.CurrentModelID != "" {
			s.Model, s.ModelConfigured = sum.CurrentModelID, true
		}
	}
	var tr timeRange
	turn, inTurn := 0, false
	var pendingText strings.Builder
	pendingLine, msgSeq := 0, 0
	calls := map[string]int{}
	finished := map[string]bool{}
	add := func(e traceEvent) {
		s.Trace.Events = append(s.Trace.Events, e)
		s.Turn = append(s.Turn, turn)
	}
	flush := func() {
		if pendingText.Len() == 0 {
			return
		}
		msgSeq++
		add(traceEvent{Line: pendingLine, Kind: "text", Role: "assistant", Text: pendingText.String(), Message: fmt.Sprintf("grok-msg-%d", msgSeq)})
		pendingText.Reset()
	}
	line := 0
	scan := auditScanner(data)
	for scan.Scan() {
		line++
		var ev struct {
			Method    string
			Params    struct{ Update map[string]any }
			Timestamp float64
		}
		if json.Unmarshal(scan.Bytes(), &ev) != nil {
			s.Trace.MalformedLines = append(s.Trace.MalformedLines, line)
			continue
		}
		if ev.Timestamp > 0 {
			tr.add(time.Unix(int64(ev.Timestamp), 0).UTC().Format(auditTimeLayout))
		}
		u := ev.Params.Update
		switch str(u["sessionUpdate"]) {
		case "user_message_chunk":
			flush()
			if !inTurn {
				turn++
				inTurn = true
			}
		case "turn_completed":
			flush()
			inTurn = false
		case "agent_message_chunk":
			if pendingText.Len() == 0 {
				pendingLine = line
			}
			pendingText.WriteString(str(object(u["content"])["text"]))
		case "tool_call":
			flush()
			id := str(u["toolCallId"])
			name := str(object(object(u["_meta"])["x.ai/tool"])["name"])
			if name == "" {
				name = str(u["title"])
			}
			args := object(u["rawInput"])
			name, args = grokToolCall(name, args)
			e := traceEvent{Line: line, Kind: toolKind(name), Tool: name, ID: id, Path: first(args, "file_path", "filePath", "path", "target_file", "target_directory"), Command: first(args, "command", "cmd"), Message: "grok-call-" + id}
			if encoded, err := json.Marshal(args); err == nil {
				e.Input = json.RawMessage(encoded)
			}
			e.Memory = codexMemoryTool(name)
			if str(object(u["rawInput"])["tool_name"]) != "" {
				e.Memory = e.Memory || strings.Contains(strings.ToLower(str(object(u["rawInput"])["server_name"])), "engram")
			}
			add(e)
			calls[id] = len(s.Trace.Events) - 1
		case "tool_call_update":
			status := str(u["status"])
			id := str(u["toolCallId"])
			out := object(u["rawOutput"])
			if (status != "completed" && status != "failed") || out == nil || finished[id] {
				continue
			}
			i, known := calls[id]
			if !known {
				continue
			}
			finished[id] = true
			ok := status == "completed"
			text := ""
			if s.Trace.Events[i].Kind == "shell" {
				code, hasCode := out["exit_code"].(float64)
				ok = ok && hasCode && code == 0 && !truth(out["timed_out"])
				text = str(out["output"])
				if !hasCode {
					s.Trace.Events[i].Success = nil
					add(traceEvent{Line: line, Kind: "tool_result", ID: id, Text: text, Message: s.Trace.Events[i].Message})
					continue
				}
			}
			if s.Trace.Events[i].Kind == "question" {
				text = str(object(out["UserAnswered"])["message"])
			}
			s.Trace.Events[i].Success = &ok
			add(traceEvent{Line: line, Kind: "tool_result", ID: id, Text: text, Success: &ok, Message: s.Trace.Events[i].Message})
		}
	}
	if err := scan.Err(); err != nil {
		return s, fmt.Errorf("grok updates: %w", err)
	}
	flush()
	s.First, s.Last = tr.first, tr.last
	return s, nil
}

// grokCwd decodes the project folder name two levels above updates.jsonl
// (sessions/<url-encoded project>/<session>/updates.jsonl). A name that is not
// an encoded absolute path gives no cwd.
func grokCwd(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	name := filepath.Base(filepath.Dir(filepath.Dir(abs)))
	decoded, err := url.PathUnescape(name)
	if err != nil || !strings.HasPrefix(decoded, "/") {
		return ""
	}
	return decoded
}

// --- base branch ---

var baseBranchLine = regexp.MustCompile(`^\s*[-*]\s*Base branch:\s*(\S+)`)

// resolveAuditBase returns the --base value, or the Base branch of the
// `## Hive` section of AGENTS.md in cwd, or "".
func resolveAuditBase(cwd, flagBase string) string {
	if flagBase != "" {
		return flagBase
	}
	if cwd == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(cwd, "AGENTS.md"))
	if err != nil {
		return ""
	}
	inHive := false
	for _, l := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(l, "## ") {
			inHive = strings.TrimSpace(strings.TrimPrefix(l, "## ")) == "Hive"
			continue
		}
		if inHive {
			if m := baseBranchLine.FindStringSubmatch(l); m != nil {
				return m[1]
			}
		}
	}
	return ""
}

// --- criteria ---

// isCleanupApply recognizes the cleanup script run with --apply, which deletes
// merged branches. It counts as a branch deletion only in the audit.
func isCleanupApply(args []string) bool {
	script := false
	apply := false
	for _, a := range args {
		if filepath.Base(a) == "cleanup_branches.py" {
			script = true
		}
		if a == "--apply" {
			apply = true
		}
	}
	return script && apply
}

func auditBranchDeletion(args []string) bool {
	return branchDeletionCommand(args) || isCleanupApply(args)
}

func hasLocalGitMerge(events []traceEvent) bool {
	for _, e := range events {
		if e.Kind != "shell" || e.Success == nil || !*e.Success {
			continue
		}
		for _, args := range shellSegments(e.Command, 0) {
			if isGit, sub, _ := gitInvocation(args, ""); isGit && sub < len(args) && args[sub] == "merge" {
				return true
			}
		}
	}
	return false
}

func existingAuditCriteria(s auditSession, base string) []criterionAssessment {
	r := result{Host: s.Host, Cwd: s.Cwd, Root: s.Cwd, Trace: s.Trace}
	criteria := regressionCriteria(r)
	for i, c := range criteria {
		if c.Criterion == "no_broad_git_add" {
			criteria[i] = noBroadGitAddCheck(r, func(abs string) bool {
				info, err := os.Stat(abs)
				return err == nil && info.IsDir()
			})
		}
	}
	merged := mergedBranchDeletedWith(r, base, auditBranchDeletion)
	if base == "" && merged.Status == "not_observed" && hasLocalGitMerge(s.Trace.Events) {
		merged.Evidence = []string{"base branch unknown: pass --base or declare Base branch in the AGENTS.md ## Hive section of the session cwd"}
	}
	return append(criteria, merged)
}

// --- criteria that exist only for real sessions ---

// roundTitleLine and roundOptionStar define a decision round written as text: a
// title line "D<n> — ..." and at least one option line "A - (★) ...".
var (
	roundTitleLine  = regexp.MustCompile(`(?m)^\s*[#>*\s]*D\d+\s*[—–-]`)
	roundOptionStar = regexp.MustCompile(`(?m)^[\s#>*-]*[A-Z]\s*[-—–]\s*\**\s*\(★\)`)
	emptyCardResult = regexp.MustCompile(`(?i)timed out|timeout|time out|expired|no user is available|non-interactive`)
)

func isTextRound(text string) bool {
	return roundTitleLine.MatchString(text) && roundOptionStar.MatchString(text)
}

// emptyCardOutput reports a native question result that carries no answer
// because the card closed by itself: an empty output, an empty answers object,
// or a timeout or no-user message. A declined or dismissed card is not empty:
// the user left it unanswered and the round must not be written out.
func emptyCardOutput(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" || emptyCardResult.MatchString(trimmed) {
		return true
	}
	var v map[string]any
	if json.Unmarshal([]byte(trimmed), &v) == nil {
		if answers, ok := v["answers"]; ok {
			m := object(answers)
			return answers == nil || (m != nil && len(m) == 0)
		}
	}
	return false
}

// textRoundWithoutCard fails when an assistant message writes a decision round
// as text and no native question call follows it before its turn ends, in a
// session that called one at another time (proof that the tool was allowed).
// A round does not count when the last card of its turn before the round came
// back empty or timed out (the guidance then asks for the round as text), in a
// subagent rollout, or in a session with no native question call at all.
func textRoundWithoutCard(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "text_round_without_card", Status: "not_observed", Evidence: []string{}}
	if s.Subagent {
		c.Evidence = []string{"subagent session: the host may not allow native questions"}
		return c
	}
	events := s.Trace.Events
	anyQuestion := false
	for _, e := range events {
		if e.Kind == "question" {
			anyQuestion = true
		}
	}
	if !anyQuestion {
		return c
	}
	var failed, passed []string
	for ri, e := range events {
		if e.Kind != "text" || e.Role != "assistant" || !isTextRound(e.Text) {
			continue
		}
		turn := s.Turn[ri]
		answered, lastQ := false, -1
		for i := range events {
			if events[i].Kind != "question" || s.Turn[i] != turn {
				continue
			}
			if i > ri {
				answered = true
				passed = append(passed, regressionEvidence(events[i]))
				break
			}
			lastQ = i
		}
		if answered {
			continue
		}
		if lastQ >= 0 && cardResultIsEmpty(events, lastQ, ri) {
			continue
		}
		failed = append(failed, regressionEvidence(e))
	}
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(passed) > 0:
		c.Status, c.Evidence = "pass", passed
	}
	return c
}

// cardResultIsEmpty reports whether the question at qi has a result before ri
// and that result is empty or timed out.
func cardResultIsEmpty(events []traceEvent, qi, ri int) bool {
	for i := qi + 1; i < ri; i++ {
		if events[i].Kind == "tool_result" && events[i].ID == events[qi].ID {
			return emptyCardOutput(events[i].Text)
		}
	}
	return false
}

var closeReportFirstLine = regexp.MustCompile(`^\s*\*\*\s*(Cierre|Close)\b`)

func isCloseReport(text string) bool {
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		return closeReportFirstLine.MatchString(l)
	}
	return false
}

// closeBeforeBranchCleanup fails when, after a successful `gh pr merge` that
// does not delete the branch, an assistant message whose first line is a bold
// «Cierre» or «Close» heading comes before the first command that deletes a
// branch. It is not observed without a later deletion: merged_branch_deleted
// covers that case.
func closeBeforeBranchCleanup(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "close_before_branch_cleanup", Status: "not_observed", Evidence: []string{}}
	events := s.Trace.Events
	merge := -1
	for i, e := range events {
		if e.Kind != "shell" || e.Success == nil || !*e.Success {
			continue
		}
		for _, args := range shellSegments(e.Command, 0) {
			if isGhPrSubcommand(args, "merge") && !ghPrMergeDeletesBranch(args) {
				merge = i
			}
		}
	}
	if merge < 0 {
		return c
	}
	deletion := -1
	for i := merge + 1; i < len(events) && deletion < 0; i++ {
		e := events[i]
		if e.Kind != "shell" || e.Success == nil || !*e.Success {
			continue
		}
		for _, args := range shellSegments(e.Command, 0) {
			if auditBranchDeletion(args) {
				deletion = i
				break
			}
		}
	}
	if deletion < 0 {
		return c
	}
	for i := merge + 1; i < deletion; i++ {
		if events[i].Kind == "text" && events[i].Role == "assistant" && isCloseReport(events[i].Text) {
			c.Status = "fail"
			c.Evidence = []string{regressionEvidence(events[i])}
			return c
		}
	}
	c.Status = "pass"
	c.Evidence = []string{regressionEvidence(events[deletion])}
	return c
}

var (
	hiveSkillName = regexp.MustCompile(`\b(?:flow-[a-z][a-z-]*|git-workflow|harness-audit|docs-cleanup|create-issue|create-prd)\b`)
	demandVerb    = regexp.MustCompile(`(?i)\b(?:pide|piden|exige|exigen|establece|establecen|requires?|asks?|sets?)\b`)
	sentenceBreak = regexp.MustCompile(`[.!?]+\s+|\n+`)
)

// isHiveRepository reports whether cwd, or a parent of it, holds Hive's skills,
// where naming a skill is legitimate because the work changes it.
func isHiveRepository(cwd string) bool {
	if cwd == "" || !filepath.IsAbs(cwd) {
		return false
	}
	for dir := filepath.Clean(cwd); ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "content", "skills", "flow-close", "SKILL.md")); err == nil {
			return true
		}
		if dir == filepath.Dir(dir) {
			return false
		}
	}
}

// skillNamedAsReason flags, for review and never as a failure, an assistant
// sentence that names a Hive skill next to a verb of demand: the reason for a
// step should be its effect on the user, not the instruction that asks for it.
func skillNamedAsReason(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "skill_named_as_reason", Status: "not_observed", Evidence: []string{}}
	if isHiveRepository(s.Cwd) {
		c.Evidence = []string{"the session cwd is the Hive repository, where naming a skill is legitimate"}
		return c
	}
	observed := false
	for _, e := range s.Trace.Events {
		if e.Kind != "text" || e.Role != "assistant" {
			continue
		}
		observed = true
		for _, sentence := range sentenceBreak.Split(e.Text, -1) {
			if hiveSkillName.MatchString(sentence) && demandVerb.MatchString(sentence) {
				c.Status = "review"
				c.Evidence = append(c.Evidence, regressionEvidence(e))
				break
			}
		}
	}
	if observed && c.Status != "review" {
		c.Status = "pass"
	}
	return c
}

func auditCriteria(s auditSession, base string) []criterionAssessment {
	return append(existingAuditCriteria(s, base), textRoundWithoutCard(s), closeBeforeBranchCleanup(s), skillNamedAsReason(s))
}

func auditSessionFile(path, host string, opts auditOptions) (auditReport, error) {
	s, err := loadAuditSession(path, host, opts)
	if err != nil {
		return auditReport{}, err
	}
	base := resolveAuditBase(s.Cwd, opts.Base)
	criteria := auditCriteria(s, base)
	prefix := filepath.Base(path) + ":"
	for i := range criteria {
		ev := make([]string, 0, len(criteria[i].Evidence))
		for _, e := range criteria[i].Evidence {
			ev = append(ev, strings.Replace(e, "stdout.jsonl:", prefix, 1))
		}
		criteria[i].Evidence = ev
	}
	return auditReport{
		Host: host, File: path, Cwd: s.Cwd, Model: s.Model, ModelConfigured: s.ModelConfigured, Subagent: s.Subagent,
		FirstTimestamp: s.First, LastTimestamp: s.Last, Criteria: criteria,
	}, nil
}

// auditSessionMain implements `--audit-session`: one JSON report on stdout.
func auditSessionMain(path, host string, opts auditOptions, stdout io.Writer) error {
	rep, err := auditSessionFile(path, host, opts)
	if err != nil {
		return err
	}
	out, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(stdout, string(out))
	return err
}
