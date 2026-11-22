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
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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
	// GrokSessionsRoot is the folder that holds Grok's project folders, where a
	// child's usage.json is looked up (default ~/.grok/sessions).
	GrokSessionsRoot string
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
	// Sidechain: a Claude Code child transcript, which holds only sidechain
	// messages. Only the newer criteria skip it; the older ones keep judging it.
	Sidechain bool
	// QuestionOffered: the host listed its waiting question tool in a mode
	// that allows it (Codex additional_tools in Default or Plan mode).
	QuestionOffered bool
	// Prompts holds, per turn, the text the user typed to open it (injected
	// context excluded), so a criterion can tell a reply from a typed command.
	Prompts     map[int]string
	First, Last string
	Trace       traceReport
	Turn        []int
	// Cost is the informational main-thread cost block; no criterion reads it.
	Cost mainThreadCost
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
	CardAnswers     cardAnswers           `json:"card_answers"`
	MainThreadCost  mainThreadCost        `json:"main_thread_cost"`
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
	var cost mainThreadCost
	switch host {
	case "claude":
		s, err = loadClaudeSession(data)
		cost = claudeCost(data)
	case "codex":
		s, err = loadCodexSession(data)
		cost = codexCost(data)
	default:
		s, err = loadGrokSession(data, path)
		cost = grokCost(data, path, opts.GrokSessionsRoot)
	}
	if err != nil {
		return auditSession{}, err
	}
	s.Cost = cost
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
	s.Prompts = map[int]string{}
	var tr timeRange
	sawMain, sawSide := false, false
	line := 0
	scan := auditScanner(data)
	for scan.Scan() {
		line++
		var ev struct {
			Type, Timestamp, Cwd string
			IsMeta, IsSidechain  bool
			IsCompactSummary     bool
			Message              struct{ Content json.RawMessage }
		}
		if json.Unmarshal(scan.Bytes(), &ev) != nil {
			continue
		}
		tr.add(ev.Timestamp)
		if s.Cwd == "" {
			s.Cwd = ev.Cwd
		}
		if ev.Type == "user" || ev.Type == "assistant" {
			sawSide = sawSide || ev.IsSidechain
			sawMain = sawMain || !ev.IsSidechain
		}
		if ev.Type == "user" && !ev.IsMeta && !ev.IsSidechain && !ev.IsCompactSummary && claudePromptContent(ev.Message.Content) {
			starts = append(starts, line)
			s.Prompts[len(starts)] = claudePromptText(ev.Message.Content)
		}
	}
	if err := scan.Err(); err != nil {
		return s, fmt.Errorf("claude transcript: %w", err)
	}
	s.Sidechain = sawSide && !sawMain
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
// (a string, or blocks that are not all tool results). The interruption notice
// Claude Code writes is not a prompt, and neither is a compact summary (the
// caller checks isCompactSummary).
func claudePromptContent(raw json.RawMessage) bool {
	const interrupted = "[Request interrupted"
	var str string
	if json.Unmarshal(raw, &str) == nil {
		return strings.TrimSpace(str) != "" && !strings.HasPrefix(strings.TrimSpace(str), interrupted)
	}
	var blocks []struct{ Type, Text string }
	if json.Unmarshal(raw, &blocks) != nil {
		return false
	}
	for _, b := range blocks {
		if b.Type == "image" || (b.Type == "text" && !strings.HasPrefix(strings.TrimSpace(b.Text), interrupted)) {
			return true
		}
	}
	return false
}

// claudePromptText returns the typed text of a user prompt's content.
func claudePromptText(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	var blocks []struct{ Type, Text string }
	if json.Unmarshal(raw, &blocks) != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// --- Codex rollout ---

type codexBuilder struct {
	s          auditSession
	turn       int
	msgSeq     int
	toolListed bool
	modeAllows bool
}

func (b *codexBuilder) add(e traceEvent) {
	b.s.Trace.Events = append(b.s.Trace.Events, e)
	b.s.Turn = append(b.s.Turn, b.turn)
}

func loadCodexSession(data []byte) (auditSession, error) {
	b := &codexBuilder{}
	b.s.Prompts = map[int]string{}
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
			// request_user_input says it is "only available in Default or Plan mode".
			if mode := str(object(p["collaboration_mode"])["mode"]); mode == "default" || mode == "plan" {
				b.modeAllows = true
			}
		case "event_msg":
			switch str(p["type"]) {
			case "task_started":
				b.turn++
			case "item_completed":
				b.codexCompletedItem(line, object(p["item"]))
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
	b.s.QuestionOffered = b.toolListed && b.modeAllows
	b.s.First, b.s.Last = tr.first, tr.last
	return b.s, nil
}

func (b *codexBuilder) codexItem(line int, p map[string]any) {
	switch str(p["type"]) {
	case "additional_tools":
		if listsTool(p["tools"], "request_user_input") {
			b.toolListed = true
		}
	case "message":
		var parts []string
		for _, raw := range list(p["content"]) {
			if t := str(object(raw)["text"]); t != "" {
				parts = append(parts, t)
			}
		}
		if str(p["role"]) == "user" {
			// Codex injects its context and the instruction files as user
			// messages that open with markup or a heading; they are not typed.
			if text := strings.TrimSpace(strings.Join(parts, "\n")); text != "" && !strings.HasPrefix(text, "<") && !strings.HasPrefix(text, "# ") {
				b.s.Prompts[b.turn] = strings.TrimSpace(b.s.Prompts[b.turn] + "\n" + text)
			}
			return
		}
		if str(p["role"]) != "assistant" {
			return
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

// codexCompletedItem records what Codex reports after an exec block's call
// ran: an MCP call with its arguments and result, which the block's source
// cannot give, and the command a shell call executed, interpolation resolved.
// They get kinds of their own so the criteria that read calls from the block
// do not count them twice.
func (b *codexBuilder) codexCompletedItem(line int, item map[string]any) {
	switch str(item["type"]) {
	case "McpToolCall":
		args, _ := json.Marshal(item["arguments"])
		var parts []string
		for _, raw := range list(object(item["result"])["content"]) {
			parts = append(parts, str(object(raw)["text"]))
		}
		e := traceEvent{Line: line, Kind: "mcp_call", Tool: "mcp__" + str(item["server"]) + "__" + str(item["tool"]), Input: args, Text: strings.Join(parts, "\n")}
		e.Memory = codexMemoryTool(e.Tool)
		b.add(e)
	case "CommandExecution":
		argv := list(item["command"])
		if len(argv) == 0 {
			return
		}
		b.add(traceEvent{Line: line, Kind: "exec_item", Command: str(argv[len(argv)-1])})
	}
}

// listsTool reports whether a tool list, or a namespace inside it, names the tool.
func listsTool(v any, name string) bool {
	for _, raw := range list(v) {
		tool := object(raw)
		if str(tool["name"]) == name || listsTool(tool["tools"], name) {
			return true
		}
	}
	return false
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
	calls, ambiguous := scanCodexExec(js)
	idx, spawns := 0, 0
	for _, c := range calls {
		switch {
		case c.name == "write_stdin":
			continue
		case c.name == "exec_command":
			e := traceEvent{Line: line, Kind: "shell", Tool: "exec_command", ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id}
			if c.hasLiteral && !c.interpolate && !ambiguous {
				e.Command = c.literal
			} else {
				e.CommandWithheld = true
				if c.hasLiteral && c.interpolate && !ambiguous {
					e.CommandTemplate = c.literal
				}
			}
			b.add(e)
		case c.name == "apply_patch":
			b.codexPatch(line, id, idx, c)
		default:
			e := traceEvent{Line: line, Kind: "tool", Tool: c.name, ID: fmt.Sprintf("%s#%d", id, idx), Message: "codex-call-" + id}
			e.Memory = codexMemoryTool(c.name)
			if c.name == "spawn_agent" {
				// The role a child runs is all a criterion needs from its brief;
				// the k-th spawn of a block gets the k-th role it names.
				if roles := spawnRole.FindAllStringSubmatch(js, -1); spawns < len(roles) {
					e.Text = roles[spawns][1]
				}
				spawns++
			}
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
	// Success is known only when the block reports one exit code per command.
	// Truncated output, or a command still running that returns a session ID,
	// leaves it unknown.
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
func scanCodexExec(js string) ([]codexCall, bool) {
	var calls []codexCall
	i := 0
	var prev byte
	for i < len(js) {
		c := js[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '"' || c == '\'' || c == '`':
			_, i, _ = jsLiteral(js, i)
			prev = '"'
		case c == '/' && i+1 < len(js) && js[i+1] == '/':
			for i < len(js) && js[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(js) && js[i+1] == '*':
			end := strings.Index(js[i+2:], "*/")
			if end < 0 {
				return fallbackCodexCalls(js), true
			}
			i += end + 4
		case c == '/' && regexAllowed(prev):
			end, ok := skipJSRegex(js, i)
			if !ok {
				return fallbackCodexCalls(js), true
			}
			i, prev = end, 'a'
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
					call.literal = interpHole.ReplaceAllString(call.literal, "")
				}
				calls = append(calls, call)
				i, prev = k+1, '('
				continue
			}
			i, prev = k, 'a'
		default:
			prev = c
			i++
		}
	}
	return calls, false
}

var toolsCallPattern = regexp.MustCompile(`\btools\.([A-Za-z_][A-Za-z0-9_]*)\(`)

// fallbackCodexCalls lists the calls of a block the scanner cannot tokenize
// (for example a regular expression literal that never closes). Their commands
// are not extracted: a wrong guess would put text from the wrong place into a
// criterion.
func fallbackCodexCalls(js string) []codexCall {
	var calls []codexCall
	for _, m := range toolsCallPattern.FindAllStringSubmatch(js, -1) {
		calls = append(calls, codexCall{name: m[1]})
	}
	return calls
}

// regexAllowed reports whether a "/" after the last significant byte prev
// starts a regular expression literal rather than a division.
func regexAllowed(prev byte) bool {
	switch prev {
	case 0, '(', ',', '=', ':', '[', '!', '&', '|', '?', '{', '}', ';', '+', '-', '*', '%', '<', '>', '~', '^':
		return true
	}
	return false
}

// skipJSRegex returns the index after the regular expression literal and its
// flags that starts at js[i]; false when it does not close on its line.
func skipJSRegex(js string, i int) (int, bool) {
	inClass := false
	for j := i + 1; j < len(js); j++ {
		switch js[j] {
		case '\\':
			j++
		case '\n':
			return 0, false
		case '[':
			inClass = true
		case ']':
			inClass = false
		case '/':
			if !inClass {
				j++
				for j < len(js) && isJSIdent(js[j]) {
					j++
				}
				return j, true
			}
		}
	}
	return 0, false
}

func skipJSSpace(js string, i int) int {
	for i < len(js) {
		switch {
		case js[i] == ' ' || js[i] == '\t' || js[i] == '\n' || js[i] == '\r':
			i++
		case strings.HasPrefix(js[i:], "//"):
			for i < len(js) && js[i] != '\n' {
				i++
			}
		case strings.HasPrefix(js[i:], "/*"):
			end := strings.Index(js[i+2:], "*/")
			if end < 0 {
				return len(js)
			}
			i += end + 4
		default:
			return i
		}
	}
	return i
}

// skipJSExpr returns the index of the "," or "}" that ends the property value
// starting at js[i], skipping nested brackets, strings and regular expression
// literals; false when the text cannot be tokenized.
func skipJSExpr(js string, i int) (int, bool) {
	depth := 0
	var prev byte
	for i < len(js) {
		c := js[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '"' || c == '\'' || c == '`':
			_, i, _ = jsLiteral(js, i)
			prev = '"'
		case strings.HasPrefix(js[i:], "//") || strings.HasPrefix(js[i:], "/*"):
			i = skipJSSpace(js, i)
		case c == '/' && regexAllowed(prev):
			end, ok := skipJSRegex(js, i)
			if !ok {
				return 0, false
			}
			i, prev = end, 'a'
		case c == '{' || c == '[' || c == '(':
			depth++
			prev = c
			i++
		case c == '}' || c == ']' || c == ')':
			if depth == 0 {
				return i, true
			}
			depth--
			prev = c
			i++
		case c == ',' && depth == 0:
			return i, true
		default:
			prev = c
			i++
		}
	}
	return 0, false
}
func isJSIdent(c byte) bool {
	return c == '_' || c == '$' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// interpHole matches the placeholder jsLiteral writes for each interpolation.
var interpHole = regexp.MustCompile(`__INTERP_[0-9]+__`)

// jsLiteral decodes the string or template literal opening at js[i] and
// returns its value, the index after its closing quote and whether a template
// interpolation was seen; each interpolation becomes a distinct placeholder
// word, __INTERP_<n>__.
func jsLiteral(js string, i int) (value string, end int, interpolated bool) {
	quote := js[i]
	var sb strings.Builder
	j := i + 1
	holes := map[string]int{}
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
			start := j
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
			expr := strings.TrimSpace(js[start:max(start, j-1)])
			n, seen := holes[expr]
			if !seen {
				n = len(holes)
				holes[expr] = n
			}
			fmt.Fprintf(&sb, "__INTERP_%d__", n)
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
	i = skipJSSpace(js, i)
	if i >= len(js) || js[i] != '{' {
		return "", false, false
	}
	j := i + 1
	for {
		j = skipJSSpace(js, j)
		if j >= len(js) || js[j] == '}' {
			return "", false, false
		}
		if js[j] == ',' {
			j++
			continue
		}
		var name string
		switch {
		case js[j] == '"' || js[j] == '\'' || js[j] == '`':
			name, j, _ = jsLiteral(js, j)
		case isJSIdent(js[j]):
			k := j
			for k < len(js) && isJSIdent(js[k]) {
				k++
			}
			name, j = js[j:k], k
		default:
			// A spread or a computed key could hide the property.
			return "", false, false
		}
		j = skipJSSpace(js, j)
		if j >= len(js) || js[j] != ':' {
			// Shorthand property: the value is a variable.
			if name == key {
				return "", false, false
			}
			continue
		}
		j = skipJSSpace(js, j+1)
		if name == key {
			if j >= len(js) || (js[j] != '"' && js[j] != '\'' && js[j] != '`') {
				return "", false, false
			}
			v, end, interp := jsLiteral(js, j)
			end = skipJSSpace(js, end)
			if end < len(js) && js[end] != ',' && js[end] != '}' {
				return "", false, false
			}
			return v, true, interp
		}
		end, good := skipJSExpr(js, j)
		if !good {
			return "", false, false
		}
		j = end
	}
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
	if end < len(js) && js[end] != ',' && js[end] != ')' {
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
	s.Prompts = map[int]string{}
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
			s.Prompts[turn] += str(object(u["content"])["text"])
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
				// Only a UserAnswered result of a completed call is an answer or
				// a card that closed by itself; a decline (UserDeclined, a failed
				// call) is the user leaving the card unanswered.
				answered := object(out["UserAnswered"])
				ok = ok && answered != nil
				text = str(answered["message"])
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

// isHelperScript reports whether tok names the cleanup script.
func isHelperScript(tok string) bool {
	return filepath.Base(tok) == "cleanup_branches.py"
}

var envAssignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// helperInvocation reports whether args run the cleanup script: the script is
// the command, or the first positional argument of a python interpreter, after
// leading VAR=value tokens and the wrappers env, timeout <n> and uv run. A
// command that only names the script as an argument, such as a search, does not.
func helperInvocation(args []string) bool {
	i := 0
	skipAssignments := func() {
		for i < len(args) && envAssignment.MatchString(args[i]) {
			i++
		}
	}
	skipAssignments()
	for i < len(args) {
		switch filepath.Base(args[i]) {
		case "env":
			i++
			for i < len(args) && (strings.HasPrefix(args[i], "-") || envAssignment.MatchString(args[i])) {
				i++
			}
			continue
		case "timeout":
			i++
			for i < len(args) && strings.HasPrefix(args[i], "-") {
				i++
			}
			i++
			continue
		case "uv":
			if i+1 < len(args) && args[i+1] == "run" {
				i += 2
				for i < len(args) && strings.HasPrefix(args[i], "-") {
					i++
				}
				continue
			}
		}
		break
	}
	if i >= len(args) {
		return false
	}
	if isHelperScript(args[i]) {
		return true
	}
	if !strings.HasPrefix(filepath.Base(args[i]), "python") {
		return false
	}
	for i++; i < len(args); i++ {
		switch a := args[i]; {
		case a == "-X" || a == "-W":
			i++
		case strings.HasPrefix(a, "-"):
		default:
			return isHelperScript(a)
		}
	}
	return false
}

// isCleanupApply recognizes the cleanup script run with --apply, which deletes
// merged branches. It counts as a branch deletion only in the audit.
func isCleanupApply(args []string) bool {
	if !helperInvocation(args) {
		return false
	}
	for _, a := range args {
		if a == "--apply" {
			return true
		}
	}
	return false
}

var helperAssignment = regexp.MustCompile(`(?:^|[\s;&|])([A-Za-z_][A-Za-z0-9_]*)=["']?([^\s"';&|]*cleanup_branches\.py)["']?`)

// expandHelperVariables replaces, in a command, each variable the command
// itself assigns to a path ending in cleanup_branches.py with that path, so a
// script run through "$SCRIPT" reads as the script. A variable assigned
// elsewhere, or to another path, is left alone.
func expandHelperVariables(command string) string {
	if !strings.Contains(command, "cleanup_branches.py") {
		return command
	}
	for _, m := range helperAssignment.FindAllStringSubmatch(command, -1) {
		ref := regexp.MustCompile(`\$\{` + regexp.QuoteMeta(m[1]) + `\}|\$` + regexp.QuoteMeta(m[1]) + `\b`)
		command = ref.ReplaceAllLiteralString(command, m[2])
	}
	return command
}

func auditBranchDeletion(args []string) bool {
	return branchDeletionCommand(args) || isCleanupApply(args)
}

// eventDeletesBranch reports whether a successful shell event deletes a branch
// in the audit's sense: a deletion command, or the cleanup helper with --apply
// however the script is invoked.
func eventDeletesBranch(e traceEvent) bool {
	if e.Kind != "shell" || e.Success == nil || !*e.Success {
		return false
	}
	for _, args := range shellSegments(expandHelperVariables(e.Command), 0) {
		if auditBranchDeletion(args) {
			return true
		}
	}
	return false
}

// deletedBranchNames lists the branches a session's deletions name: the
// `--merged-pr BRANCH=SHA` values of the helper run with --apply, the branches
// of `git branch -d|-D|--delete`, those of `git push --delete` or a `:branch`
// refspec, and the ref a `gh api -X DELETE .../git/refs/heads/<branch>` names.
func deletedBranchNames(events []traceEvent) map[string]bool {
	names := map[string]bool{}
	add := func(b string) {
		b = strings.TrimPrefix(strings.TrimPrefix(b, ":"), "refs/heads/")
		if b != "" {
			names[b] = true
		}
	}
	for _, e := range events {
		if !eventDeletesBranch(e) {
			continue
		}
		for _, args := range shellSegments(expandHelperVariables(e.Command), 0) {
			if isCleanupApply(args) {
				for i, a := range args {
					switch {
					case a == "--merged-pr" && i+1 < len(args):
						add(strings.SplitN(args[i+1], "=", 2)[0])
					case strings.HasPrefix(a, "--merged-pr="):
						add(strings.SplitN(strings.TrimPrefix(a, "--merged-pr="), "=", 2)[0])
					}
				}
			}
			if isGhAPIDeleteRef(args) {
				for _, a := range args {
					if i := strings.Index(a, "git/refs/heads/"); i >= 0 {
						add(a[i+len("git/refs/heads/"):])
					}
				}
				continue
			}
			isGit, sub, _ := gitInvocation(args, "")
			if !isGit || sub >= len(args) {
				continue
			}
			rest := args[sub+1:]
			switch args[sub] {
			case "branch":
				if gitBranchDeleteCommand(rest) {
					for _, a := range rest {
						if !strings.HasPrefix(a, "-") {
							add(a)
						}
					}
				}
			case "push":
				if !gitPushDeletesBranch(rest) {
					continue
				}
				deleteFlag := false
				for _, a := range rest {
					if a == "--delete" || a == "-d" {
						deleteFlag = true
					}
				}
				positional := 0
				for _, a := range rest {
					switch {
					case strings.HasPrefix(a, ":"):
						add(a)
					case strings.HasPrefix(a, "-"):
					default:
						positional++
						if deleteFlag && positional > 1 {
							add(a)
						}
					}
				}
			}
		}
	}
	return names
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
				// Without the historical inventory, a path that is a directory
				// now, or that no longer exists, cannot be judged.
				info, err := os.Stat(abs)
				return err != nil || info.IsDir()
			})
		}
	}
	expanded := r.Trace
	expanded.Events = append([]traceEvent(nil), r.Trace.Events...)
	for i, e := range expanded.Events {
		if e.Kind == "shell" {
			expanded.Events[i].Command = expandHelperVariables(e.Command)
		}
	}
	merged := mergedBranchDeletedWith(result{Host: r.Host, Cwd: r.Cwd, Root: r.Root, Trace: expanded}, base, auditBranchDeletion)
	if base == "" && merged.Status == "not_observed" && hasLocalGitMerge(s.Trace.Events) {
		merged.Evidence = []string{"base branch unknown: pass --base or declare Base branch in the AGENTS.md ## Hive section of the session cwd"}
	}
	return append(criteria, mergedAfterClose(s, base, merged))
}

// sessionEndPhrases is the subset of phrases the audit takes as the user ending
// the session. It detects, it does not define: close-modes sets what ends a
// session, and rule-enforcers.md lists this subset.
var sessionEndPhrases = []string{"es todo", "eso es todo", "terminamos", "por hoy", "por el momento es todo", "that's all", "that's it for now", "por hoy es todo", "por ahora es todo"}

// sessionEndClosers are the words that may follow an end phrase: thanks and
// the closing «por el momento» or «por hoy» that an end message carries. Any
// other word means the message goes on with the work.
var sessionEndClosers = map[string]bool{
	"gracias": true, "muchas": true, "por": true, "ahora": true, "el": true, "momento": true, "hoy": true, "ok": true, "listo": true,
	"thanks": true, "thank": true, "you": true, "for": true, "now": true,
	"es": true, "eso": true, "todo": true, "terminamos": true, "that's": true, "all": true, "it": true, "that": true,
	"tu": true, "tus": true, "ayuda": true, "nos": true, "vemos": true, "mañana": true, "de": true, "mi": true, "parte": true, "buen": true, "trabajo": true, "today": true, "bye": true,
}

// isSessionEndMessage reports whether a user message starts with a listed
// phrase, has at most 10 words, holds no question mark and continues, after the
// phrase, only with punctuation and closing words.
func isSessionEndMessage(text string) bool {
	t := strings.ToLower(strings.TrimLeft(strings.TrimSpace(text), " \t\n\"'*\u00a1!.,"))
	t = strings.ReplaceAll(t, "\u2019", "'")
	if strings.Contains(t, "?") || len(strings.Fields(t)) > 10 {
		return false
	}
	for _, phrase := range sessionEndPhrases {
		if !strings.HasPrefix(t, phrase) {
			continue
		}
		rest := t[len(phrase):]
		if r, _ := utf8.DecodeRuneInString(rest); rest != "" && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
			continue
		}
		closing := true
		for _, w := range strings.FieldsFunc(rest, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '\'' }) {
			if !sessionEndClosers[w] {
				closing = false
				break
			}
		}
		if closing {
			return true
		}
	}
	return false
}

// sessionEndAfter returns the position, in the event list, of the first user
// message after events[after] that ends the session, or -1. The message sits
// before the first event of its turn.
func sessionEndAfter(s auditSession, after int) (pos int, text string) {
	var turns []int
	for t := range s.Prompts {
		turns = append(turns, t)
	}
	sort.Ints(turns)
	for _, t := range turns {
		if !isSessionEndMessage(s.Prompts[t]) {
			continue
		}
		pos := len(s.Trace.Events)
		for i := range s.Trace.Events {
			if i < len(s.Turn) && s.Turn[i] >= t {
				pos = i
				break
			}
		}
		if pos > after {
			return pos, s.Prompts[t]
		}
	}
	return -1, ""
}

var (
	environmentsLine = regexp.MustCompile(`^\s*[-*]\s*Environments:\s*(.+)$`)
	environmentSplit = regexp.MustCompile(`\s*(?:\x{2192}|->|,)\s*`)
	ghHeadHeadRef    = regexp.MustCompile(`headRefName"?\s*[:=]\s*"?([^\s",}]+)`)
	keptBranch       = regexp.MustCompile("(?i)\\bkept\\s+`([^`\\n]+)`")
)

// promotionBranches is the base and the branches of the `Environments` line of
// the `## Hive` section of AGENTS.md in cwd.
func promotionBranches(cwd, base string) map[string]bool {
	out := map[string]bool{}
	if base != "" {
		out[base] = true
	}
	if cwd == "" {
		return out
	}
	data, err := os.ReadFile(filepath.Join(cwd, "AGENTS.md"))
	if err != nil {
		return out
	}
	inHive := false
	for _, l := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(l, "## ") {
			inHive = strings.TrimSpace(strings.TrimPrefix(l, "## ")) == "Hive"
			continue
		}
		if !inHive {
			continue
		}
		if m := environmentsLine.FindStringSubmatch(l); m != nil {
			for _, b := range environmentSplit.Split(m[1], -1) {
				if b = strings.Trim(strings.TrimSpace(b), "`"); b != "" {
					out[b] = true
				}
			}
		}
	}
	return out
}

// ghPrTarget returns what a `gh pr <sub>` invocation names as its pull
// request: its first positional argument (a number, a branch or a URL, reduced
// to its last path segment), or "" when it names none and acts on the current
// branch. Flags are skipped with the values of the flags that take one.
func ghPrTarget(args []string) string {
	valued := map[string]bool{"--repo": true, "-R": true, "--body": true, "-b": true, "--subject": true, "-t": true, "--body-file": true, "-F": true, "--match-head-commit": true, "--json": true, "--jq": true, "-q": true, "--template": true, "-T": true}
	for i := 3; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			if valued[a] && !strings.Contains(a, "=") {
				i++
			}
			continue
		}
		return a[strings.LastIndex(a, "/")+1:]
	}
	return ""
}

var pullNumber = regexp.MustCompile(`/pull/(\d+)|#(\d+)`)

// resultOf returns the text of the tool result that answers events[i].
func resultOf(events []traceEvent, i int) string {
	for j := i + 1; j < len(events); j++ {
		if events[j].Kind == "tool_result" && events[j].ID != "" && (events[j].ID == events[i].ID || strings.HasPrefix(events[i].ID, events[j].ID+"#")) {
			return events[j].Text
		}
	}
	return ""
}

// mergeHead returns the head branch the session shows for the merge at
// events[mergeAt]. It is the head of the nearest earlier `gh pr create` that
// opened the merged pull request (a create whose result shows another number is
// skipped; one with no --head gives ""), or a read of headRefName by a
// `gh pr view` that names the same pull request as the merge, both naming none
// included. A merge whose head the session does not show gives "".
func mergeHead(events []traceEvent, mergeAt int) string {
	mergeTarget := ""
	for _, args := range shellSegments(events[mergeAt].Command, 0) {
		if isGhPrSubcommand(args, "merge") {
			mergeTarget = ghPrTarget(args)
		}
	}
	mergeNumber := ""
	if _, err := strconv.Atoi(mergeTarget); err == nil {
		mergeNumber = mergeTarget
	}
	for i := mergeAt - 1; i >= 0; i-- {
		e := events[i]
		switch e.Kind {
		case "shell":
			for _, args := range shellSegments(e.Command, 0) {
				if !isGhPrSubcommand(args, "create") {
					continue
				}
				if m := pullNumber.FindStringSubmatch(resultOf(events, i)); m != nil && mergeNumber != "" && m[1]+m[2] != mergeNumber {
					continue
				}
				for j, a := range args {
					switch {
					case (a == "--head" || a == "-H") && j+1 < len(args):
						return args[j+1]
					case strings.HasPrefix(a, "--head="):
						return strings.TrimPrefix(a, "--head=")
					}
				}
				return ""
			}
		case "tool_result":
			m := ghHeadHeadRef.FindStringSubmatch(e.Text)
			if m == nil {
				continue
			}
			// The read counts only when the one segment of its call that asks
			// for headRefName is a gh pr view of the merged pull request.
			sameCall, readers, viewMatches := false, 0, false
			for j := i - 1; j >= 0; j-- {
				if events[j].Kind != "shell" {
					continue
				}
				if e.ID != "" && events[j].ID != e.ID && !strings.HasPrefix(events[j].ID, e.ID+"#") {
					if sameCall {
						break
					}
					continue
				}
				sameCall = true
				for _, args := range shellSegments(events[j].Command, 0) {
					if !strings.Contains(strings.Join(args, " "), "headRefName") {
						continue
					}
					readers++
					if isGhPrSubcommand(args, "view") && ghPrTarget(args) == mergeTarget {
						viewMatches = true
					}
				}
				if e.ID == "" {
					break
				}
			}
			if readers == 1 && viewMatches {
				return m[1]
			}
		}
	}
	return ""
}

// mergedAfterClose judges a merged_branch_deleted fail by what follows the last
// merge. A close report after it keeps the fail. With no close report, a user
// message that ends the session also fails, except, as review, when the merge
// is a promotion (its head, which the session shows, is the base or a branch of
// Environments) or the last assistant message before that message holds
// the form kept `<branch>` for the merge's branch, seen by the same source. With
// neither a report nor such a message, the session may still be open or the
// merge a promotion, and the status is review. The fail evidence is the last
// merge's own event, which locates it.
func mergedAfterClose(s auditSession, base string, merged criterionAssessment) criterionAssessment {
	events := s.Trace.Events
	if merged.Status != "fail" || len(merged.Evidence) == 0 {
		return merged
	}
	mergeAt := -1
	for i, e := range events {
		if regressionEvidence(e) == merged.Evidence[0] {
			mergeAt = i
		}
	}
	if mergeAt < 0 {
		return merged
	}
	for _, e := range events[mergeAt+1:] {
		if e.Kind == "text" && e.Role == "assistant" && isCloseReport(e.Text) {
			return merged
		}
	}
	endAt, _ := sessionEndAfter(s, mergeAt)
	if endAt < 0 {
		merged.Status = "review"
		merged.Evidence = append(merged.Evidence, "no close report after the last merge: the session may be open or the merge a promotion")
		return merged
	}
	if head := mergeHead(events, mergeAt); head != "" {
		if promotionBranches(s.Cwd, base)[head] {
			merged.Status = "review"
			merged.Evidence = append(merged.Evidence, "the merge is a promotion of "+head+", the base or an environment branch")
			return merged
		}
		for i := endAt - 1; i > mergeAt; i-- {
			if events[i].Kind != "text" || events[i].Role != "assistant" {
				continue
			}
			for _, m := range keptBranch.FindAllStringSubmatch(events[i].Text, -1) {
				if m[1] == head {
					merged.Status = "review"
					merged.Evidence = append(merged.Evidence, "the last assistant message keeps `"+head+"`, the branch of the merge")
					return merged
				}
			}
			break
		}
	}
	merged.Evidence = append(merged.Evidence, "a user message that ends the session follows the merge and no branch deletion follows it")
	return merged
}

// --- criteria that exist only for real sessions ---

// roundTitleLine and roundOptionStar define a decision round written as text: a
// title line "D<n> — ..." or "D<n> · ...", or a session question's fixed header such as
// "**Review** — ...", and at least one option line "A - (★) ...".
var (
	// A session question's header counts only on a title that asks, ending in
	// "?", so a report heading such as "**Review** — Completed" is not a round.
	// Only a D<n> title takes the middle dot: the context line of a message,
	// "**Review · ticket · is it done?**", uses it too.
	roundTitleLine  = regexp.MustCompile(`(?mi)^\s*[#>*\s]*(?:D\d+\s*[—–·-]|(?:` + sessionHeaderNames + `)\**\s*[—–-][^\n]*\?)`)
	roundOptionStar = regexp.MustCompile(`(?m)^[\s#>*-]*[A-Z]\s*[-—–·]\s*\**\s*\(★\)`)
	// grokNoUser is the result Grok gives a question when no user can answer.
	grokNoUser = regexp.MustCompile(`(?i)^\s*no user is available`)
	// claudeRejected matches the tool errors that mean the user turned the card
	// down, which is not a card that closed by itself.
	claudeRejected = regexp.MustCompile(`(?i)doesn't want|does not want|rejected|declined|dismiss|denied|interrupted|cancel`)
)

// A round written as a table has, per decision, a row with a cell that opens
// with the decision ID, such as "**D1 · Which rows?**" or a bare "D1" followed by
// the question's cell, and at least two options, such as "**A ★** — …" and
// "**B** — …" or "A) (★) …", in cells of their own or split by <br>, one starred.
var (
	tableDecisionCell = regexp.MustCompile(`^[\s*_]*D\d+\b`)
	tableBareID       = regexp.MustCompile(`^[\s*_]*D\d+[\s*_]*$`)
	// An option opens with its letter and a separator; "." and ":" count only
	// before a space, markup or the star, so "E.g." is not an option but "**A.**" is.
	tableOptionCell = regexp.MustCompile(`^[\s*_]*([A-Z])\b[\s*_]*(\(?★|[-—–·)]|[.:]([\s*_(★]|$))`)
	tableLineBreak  = regexp.MustCompile(`(?i)<br\s*/?>`)
)

// tableRounds returns the rows of a Markdown table that write a decision, each
// as its title, then its option cells one per line. A table of findings (H1…),
// a recap row with one answered option, such as "| D1 | A ★ interactive |", and
// a sentence that cites an answer, such as "D2 A ★", are not rounds.
func tableRounds(text string) []string {
	var rows []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(line, "|"), "|")
		title := -1
		for i, cell := range cells {
			if tableDecisionCell.MatchString(cell) {
				title = i
				break
			}
		}
		if title < 0 {
			continue
		}
		heading, next := strings.TrimSpace(cells[title]), title+1
		if tableBareID.MatchString(cells[title]) && next < len(cells) {
			heading += " " + strings.TrimSpace(cells[next])
			next++
		}
		var options []string
		letters := map[string]bool{}
		starred := false
		for _, cell := range cells[next:] {
			for _, part := range tableLineBreak.Split(cell, -1) {
				if m := tableOptionCell.FindStringSubmatch(part); m != nil {
					options = append(options, strings.TrimSpace(part))
					letters[m[1]] = true
					starred = starred || strings.Contains(part, "★")
				}
			}
		}
		// A round lists its options from A; a note that opens with another letter is not one.
		if !letters["A"] || !letters["B"] || !starred {
			continue
		}
		rows = append(rows, heading+"\n"+strings.Join(options, "\n"))
	}
	return rows
}

func isTextRound(text string) bool {
	return roundTitleLine.MatchString(text) && roundOptionStar.MatchString(text) || len(tableRounds(text)) > 0
}

// emptyCardResult reports whether a native question result is a card that
// closed by itself, in each host's own result shape; it never looks at what the
// user typed. Codex: an empty output, an empty answers object or a timeout
// status (that last shape is not observed in this machine's rollouts). Claude:
// an empty result or a tool error that is not a rejection. Grok: the "no user
// is available" result of a completed call. A decline or a dismissal is the
// user leaving the card unanswered, so the round must not be written out.
func emptyCardResult(host string, e traceEvent) bool {
	text := strings.TrimSpace(e.Text)
	switch host {
	case "codex":
		if text == "" {
			return true
		}
		var v map[string]any
		if json.Unmarshal([]byte(text), &v) != nil {
			return false
		}
		if answers, ok := v["answers"]; ok {
			m := object(answers)
			return answers == nil || (m != nil && len(m) == 0)
		}
		status := strings.ToLower(str(v["status"]))
		return status == "timeout" || status == "timed_out"
	case "claude":
		if e.Success == nil {
			return false
		}
		if *e.Success {
			return text == ""
		}
		return !claudeRejected.MatchString(text)
	case "grok":
		return e.Success != nil && *e.Success && grokNoUser.MatchString(text)
	}
	return false
}

// unattendedActivation returns the index of the first event that activates
// unattended delegation (a call of its skill, or a read of its SKILL.md), or
// -1. From there the guidance asks for decision rounds in text.
func unattendedActivation(events []traceEvent) int {
	const marker = "unattended-delegation/SKILL.md"
	for i, e := range events {
		switch {
		case e.Kind == "skill_invocation" && e.Text == "unattended-delegation":
			return i
		case (e.Kind == "read" || e.Kind == "search") && strings.Contains(filepath.ToSlash(e.Path), marker):
			return i
		case e.Kind == "shell" && strings.Contains(e.Command, marker):
			return i
		}
	}
	return -1
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
	if !cardAvailable(s) {
		return c
	}
	unattended := unattendedActivation(events)
	var failed, passed []string
	for ri, e := range events {
		if e.Kind != "text" || e.Role != "assistant" || !isTextRound(e.Text) {
			continue
		}
		if unattended >= 0 && ri > unattended {
			continue
		}
		if q := cardAfterInTurn(s, ri); q >= 0 {
			passed = append(passed, regressionEvidence(events[q]))
			continue
		}
		if lastQ := cardBeforeInTurn(s, ri); lastQ >= 0 && cardResultIsEmpty(s.Host, events, lastQ, ri) {
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

// cardAfterInTurn returns the first question card after the event at i in its
// turn, or -1.
func cardAfterInTurn(s auditSession, i int) int {
	for j := i + 1; j < len(s.Trace.Events) && s.Turn[j] == s.Turn[i]; j++ {
		if s.Trace.Events[j].Kind == "question" {
			return j
		}
	}
	return -1
}

// cardBeforeInTurn returns the last question card before the event at i in its
// turn, or -1.
func cardBeforeInTurn(s auditSession, i int) int {
	for j := i - 1; j >= 0 && s.Turn[j] == s.Turn[i]; j-- {
		if s.Trace.Events[j].Kind == "question" {
			return j
		}
	}
	return -1
}

// cardResultIsEmpty reports whether the question at qi has a result before ri
// and that result is empty or timed out.
func cardResultIsEmpty(host string, events []traceEvent, qi, ri int) bool {
	for i := qi + 1; i < ri; i++ {
		if events[i].Kind == "tool_result" && events[i].ID == events[qi].ID {
			return emptyCardResult(host, events[i])
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

// auditMerges returns the indexes of the events that complete a merge without
// deleting the branch: a successful `gh pr merge` without --delete-branch, or a
// successful `git merge` followed by a successful `git push` that names base.
// A command whose success is unknown is not a merge.
func auditMerges(events []traceEvent, base string) []int {
	var merges []int
	pendingGitMerge := false
	for i, e := range events {
		if e.Kind != "shell" || e.Success == nil || !*e.Success {
			continue
		}
		for _, args := range shellSegments(e.Command, 0) {
			if isGhPrSubcommand(args, "merge") && !ghPrMergeDeletesBranch(args) {
				merges = append(merges, i)
				continue
			}
			isGit, sub, _ := gitInvocation(args, "")
			if !isGit || sub >= len(args) {
				continue
			}
			switch args[sub] {
			case "merge":
				pendingGitMerge = true
			case "push":
				if pendingGitMerge && pushTargetsBase(args[sub+1:], base) {
					merges = append(merges, i)
					pendingGitMerge = false
				}
			}
		}
	}
	return merges
}

// closeBeforeBranchCleanup checks every successful merge that does not delete
// its branch (a `gh pr merge` without --delete-branch, or a local `git merge`
// followed by a push of the base; a command whose success is unknown is not a
// merge). An assistant message whose first line is a bold «Cierre» or «Close»
// heading, after the merge and before the first command that deletes a branch
// (the cleanup helper with --apply included, however it is invoked), fails: the
// cleanup belongs at the completion point, before the report. It is review,
// not fail, only when the report holds the form kept `<branch>` for a branch the
// session deleted, as deletedBranchNames reads it from `--merged-pr`, `git branch
// -d|-D`, `git push --delete`, a `:branch` refspec or `gh api -X DELETE`, since such a report names what it leaves. A merge with no later
// deletion is not observed: merged_branch_deleted covers that case.
func closeBeforeBranchCleanup(s auditSession, base string) criterionAssessment {
	c := criterionAssessment{Criterion: "close_before_branch_cleanup", Status: "not_observed", Evidence: []string{}}
	events := s.Trace.Events
	deleted := deletedBranchNames(events)
	var failed, kept, cleaned []string
	seen := map[string]bool{}
	promotions := promotionBranches(s.Cwd, base)
	for _, merge := range auditMerges(events, base) {
		if promotions[mergeHead(events, merge)] {
			continue
		}
		deletion := -1
		for i := merge + 1; i < len(events) && deletion < 0; i++ {
			if eventDeletesBranch(events[i]) {
				deletion = i
			}
		}
		if deletion < 0 {
			continue
		}
		found := false
		for i := merge + 1; i < deletion; i++ {
			if events[i].Kind == "text" && events[i].Role == "assistant" && isCloseReport(events[i].Text) {
				found = true
				ev := regressionEvidence(events[i])
				if seen[ev] {
					continue
				}
				seen[ev] = true
				if reportKeepsDeletedBranch(events[i].Text, deleted) {
					kept = append(kept, ev)
				} else {
					failed = append(failed, ev)
				}
			}
		}
		if !found {
			cleaned = append(cleaned, regressionEvidence(events[deletion]))
		}
	}
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(kept) > 0:
		c.Status, c.Evidence = "review", kept
	case len(cleaned) > 0:
		c.Status, c.Evidence = "pass", cleaned
	}
	return c
}

// reportKeepsDeletedBranch reports whether a report holds the form kept `<branch>`
// for a branch the session deleted.
func reportKeepsDeletedBranch(text string, deleted map[string]bool) bool {
	for _, m := range keptBranch.FindAllStringSubmatch(text, -1) {
		if deleted[m[1]] {
			return true
		}
	}
	return false
}

var (
	hiveSkillName = regexp.MustCompile(`\b(?:flow-[a-z][a-z-]*|git-workflow|harness-audit|docs-cleanup|project-layout|workspace-repo|create-issue|create-prd|create-design)\b`)
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
	return append(existingAuditCriteria(s, base), textRoundWithoutCard(s), closeBeforeBranchCleanup(s, base), skillNamedAsReason(s), reviewQuestionNativeOption(s), leafReadBeforeAction(s),
		flowEnteredWithoutRoute(s), decisionRoundFormat(s), uiChangeWithoutUXReview(s), trackerStatesCacheWritten(s), tryItBlockInReport(s), proseChoiceWithoutCard(s), ticketAssessmentEnding(s), selfFeedingLog(s), analysisReplyBlock(s), unboundedBackgroundWait(s), cardOptionFormat(s))
}

var (
	// The title names a review; the options decide whether it picks a mechanism.
	reviewQuestionTitle = regexp.MustCompile(`(?i)\brevisi(?:[oó]n|ones)\b|\breview(?:s|ers?)?\b`)
	// A decision about a review's findings is not the question that picks it.
	reviewFindingsTitle = regexp.MustCompile(`(?i)hallazgo|finding`)
	nativeReviewOption  = regexp.MustCompile(`(?i)/review\b|codex review|revisi[oó]n nativa|native review|\bnativ[ae]\s*:`)
	// The options of the question that picks the review mechanism name at least one
	// mechanism: the agent, the host's native review or no dedicated review.
	reviewMechanismOption = regexp.MustCompile(`(?i)hive-review-code|/review\b|codex review|revisi[oó]n nativa|native review|no dedicated review|sin revisi[oó]n dedicada`)
	// optionLine is a decision's option line written as text, such as
	// "A - (★) Accept" or "- **B · Keep**"; its Situation, "If you choose it" and
	// recommendation lines are not options.
	optionLine = regexp.MustCompile(`(?m)^[\s#>*-]*[A-Z]\**\s*[-—–·].*$`)
)

// reviewQuestionNativeOption checks, on Codex, that the code review question
// offers the host's native review: a decision title line naming a code review,
// or a card question doing so, whose options name a review mechanism. A report
// that only mentions a review, or a decision about applying a review's findings,
// is not that question.
func reviewQuestionNativeOption(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "review_question_native_option", Status: "not_observed", Evidence: []string{}}
	if s.Host != "codex" {
		return c
	}
	var failed, passed []string
	// mechanisms holds one option per line, where an option names its review
	// mechanism; names is where an option is named (a card's label, not its
	// description). The question picks a mechanism when at least two options name
	// one and its title is not about a review's findings, whose options may name
	// a mechanism too.
	judge := func(e traceEvent, title, mechanisms, names string) {
		named := 0
		for _, option := range strings.Split(mechanisms, "\n") {
			if reviewMechanismOption.MatchString(option) {
				named++
			}
		}
		if !reviewQuestionTitle.MatchString(title) || reviewFindingsTitle.MatchString(title) || named < 2 {
			return
		}
		if nativeReviewOption.MatchString(names) {
			passed = append(passed, regressionEvidence(e))
		} else {
			failed = append(failed, regressionEvidence(e))
		}
	}
	for _, e := range s.Trace.Events {
		switch {
		case e.Kind == "text" && e.Role == "assistant":
			for _, block := range decisionBlocks(e.Text) {
				title, rest, _ := strings.Cut(block, "\n")
				options := strings.Join(optionLine.FindAllString(rest, -1), "\n")
				judge(e, title, options, options)
			}
			for _, row := range tableRounds(e.Text) {
				title, cells, _ := strings.Cut(row, "\n")
				judge(e, title, cells, cells)
			}
		case e.Kind == "question" && len(e.Input) > 0:
			var in struct {
				Questions []struct {
					Question string
					Options  []struct{ Label, Description string }
				}
			}
			if json.Unmarshal(e.Input, &in) != nil {
				continue
			}
			for _, q := range in.Questions {
				var full, labels []string
				for _, o := range q.Options {
					full = append(full, o.Label+" "+o.Description)
					labels = append(labels, o.Label)
				}
				judge(e, q.Question, strings.Join(full, "\n"), strings.Join(labels, "\n"))
			}
		}
	}
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(passed) > 0:
		c.Status, c.Evidence = "pass", passed
	}
	return c
}

// decisionBlocks splits a message into its decisions, each from its title line
// (`D<n> — …`) to the next title line or the end.
func decisionBlocks(text string) []string {
	idx := roundTitleLine.FindAllStringIndex(text, -1)
	var blocks []string
	for i, loc := range idx {
		end := len(text)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		blocks = append(blocks, strings.TrimLeft(text[loc[0]:end], "\n"))
	}
	return blocks
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
		FirstTimestamp: s.First, LastTimestamp: s.Last, Criteria: criteria, CardAnswers: cardAnswersOf(s), MainThreadCost: s.Cost,
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

// --- leaf_read_before_action ---

// leafAction is a moment the guidance's rule leaves name: the first one of its
// class in a session needs the leaf's file read before it.
type leafAction struct {
	class, leaf string
	// readInCommand: an earlier segment of the same shell command names the leaf.
	readInCommand bool
}

// A leaf is named by its short name; the installed file is
// hive-core/references/<name>.md, and sessions recorded before the move to
// hive-core read <oldLeafPrefix><name>/SKILL.md.
const (
	leafGitCommit     = "git-commit"
	leafTicketState   = "ticket-state"
	leafQuestionRound = "questions"
	oldLeafPrefix     = "hive-" + "rule-"
)

// leafFormerNames lists the short names a leaf had before, so transcripts of
// sessions that ran the earlier guidance are judged by the leaf they read.
var leafFormerNames = map[string][]string{leafQuestionRound: {"question-channel"}}

// leafPathPattern matches the full path of a leaf's file: the old
// <prefix><name>/SKILL.md or the current hive-core/references/<name>.md. The
// path is compared whole, so a leaf's own nested reference
// (references/<name>/<f>.md) and hive-core/SKILL.md are no leaf read.
func leafPathPattern(leaf string) *regexp.Regexp {
	names := append([]string{leaf}, leafFormerNames[leaf]...)
	var alts []string
	for _, n := range names {
		alts = append(alts, regexp.QuoteMeta(oldLeafPrefix+n+"/SKILL.md"))
	}
	alts = append(alts, regexp.QuoteMeta("hive-core/references/"+leaf+".md"))
	return regexp.MustCompile(`(?:^|[^\w.-])(?:` + strings.Join(alts, "|") + `)(?:$|[^\w./-])`)
}

// namesLeafPath reports whether text names the leaf's file by its full path.
func namesLeafPath(text, leaf string) bool { return leafPathPattern(leaf).MatchString(text) }

// leafReadVerb names the shell programs that print a file, so a command that
// only mentions a leaf's path, such as a commit message, is not a read.
var leafReadVerb = regexp.MustCompile(`^(cat|sed|head|tail|less|more|bat|nl|awk|grep|rg)$`)

// shellReadsLeaf reports whether a shell segment prints the leaf's file: its
// program prints files, it names the path, and it redirects nothing, so a
// heredoc or a commit message that mentions the path is no read.
func shellReadsLeaf(args []string, leaf string) bool {
	return shellReadsPattern(args, leafPathPattern(leaf))
}

// shellReadsPattern reports whether a shell segment prints a file whose path
// pattern matches: a program that prints files, no redirection, and an argument
// that names the path.
func shellReadsPattern(args []string, path *regexp.Regexp) bool {
	if len(args) < 2 || !leafReadVerb.MatchString(filepath.Base(args[0])) {
		return false
	}
	for _, a := range args[1:] {
		if strings.HasPrefix(a, ">") || strings.HasPrefix(a, "<<") {
			return false
		}
	}
	return path.MatchString(strings.Join(args[1:], " "))
}

// namesLeaf reports whether a call read the leaf's file: the Read tool on its
// path, a shell segment shellReadsLeaf accepts, or another tool whose input
// names it. A call the trace marks as failed is no read.
func namesLeaf(e traceEvent, leaf string) bool {
	if e.Success != nil && !*e.Success {
		return false
	}
	switch e.Kind {
	case "read":
		return namesLeafPath(e.Path, leaf)
	case "tool":
		return namesLeafPath(e.Path+"\n"+string(e.Input), leaf)
	case "shell":
		for _, args := range shellSegments(e.Command, 0) {
			if shellReadsLeaf(args, leaf) {
				return true
			}
		}
	}
	return false
}

// ticketMoveTool matches tracker tools that change an existing ticket; a call
// counts only when it names that ticket and sets a state, so creating a ticket
// is not a move.
var ticketMoveTool = regexp.MustCompile(`(?i)(?:save|update|edit)_issue$|issue_write$|transition`)

// leafActionsIn lists the actions of an event that a leaf governs: a git add or
// commit, also through git_state.py's commit command; a ticket move (gh issue
// close, reopen or a label edit and gh project item-edit, after gh's global
// flags, or a tracker tool that names an existing ticket and sets its state,
// or transitions it); or a question card. A shell command is split into
// segments, so a read earlier in the same command counts as before.
func leafActionsIn(e traceEvent, cwd string) []leafAction {
	var out []leafAction
	switch e.Kind {
	case "question":
		if !e.Memory {
			out = append(out, leafAction{class: "question round", leaf: leafQuestionRound})
		}
	case "tool":
		if !ticketMoveTool.MatchString(e.Tool) {
			return nil
		}
		in := decodeInput(e)
		_, hasState := in["state"]
		_, hasStatus := in["status"]
		named := false
		for _, key := range []string{"id", "identifier", "issue", "issue_id", "issueId", "issue_iid", "issue_number", "issueIdOrKey", "number", "key"} {
			if _, ok := in[key]; ok {
				named = true
			}
		}
		if strings.Contains(strings.ToLower(e.Tool), "transition") || (named && (hasState || hasStatus)) {
			out = append(out, leafAction{class: "ticket move", leaf: leafTicketState})
		}
	case "shell":
		read := map[string]bool{}
		for _, args := range shellSegments(e.Command, 0) {
			for _, leaf := range []string{leafGitCommit, leafTicketState, leafQuestionRound} {
				if shellReadsLeaf(args, leaf) {
					read[leaf] = true
				}
			}
			isCommit := false
			if isGit, sub, _ := gitInvocation(args, cwd); isGit && sub < len(args) && (args[sub] == "add" || args[sub] == "commit") {
				isCommit = true
			}
			for i, a := range args {
				if filepath.Base(a) != "git_state.py" {
					continue
				}
				for _, rest := range args[i+1:] {
					if rest == "commit" {
						isCommit = true
					}
				}
			}
			if isCommit {
				out = append(out, leafAction{class: "git add or commit", leaf: leafGitCommit, readInCommand: read[leafGitCommit]})
			}
			if len(args) > 2 && filepath.Base(args[0]) == "gh" && ticketMoveShell(ghSubcommand(args[1:])) {
				out = append(out, leafAction{class: "ticket move", leaf: leafTicketState, readInCommand: read[leafTicketState]})
			}
		}
	}
	return out
}

// ghSubcommand drops gh's global flags, such as -R owner/repo, before its
// command.
func ghSubcommand(args []string) []string {
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		if (args[0] == "-R" || args[0] == "--repo") && len(args) > 1 {
			args = args[2:]
			continue
		}
		args = args[1:]
	}
	return args
}

func ticketMoveShell(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch {
	case args[0] == "issue" && (args[1] == "close" || args[1] == "reopen"):
		return true
	case args[0] == "issue" && args[1] == "edit":
		for _, a := range args[2:] {
			if strings.HasPrefix(a, "--add-label") || strings.HasPrefix(a, "--remove-label") {
				return true
			}
		}
	case args[0] == "project" && args[1] == "item-edit":
		return true
	}
	return false
}

// leafReadBeforeAction checks that, before the first action of each class a
// rule leaf governs, the transcript shows a read of that leaf's SKILL.md.
// Status: fail when an observed action has no read before it; review when the
// only gap is a shell command the host withheld before it (the read could be
// in it); pass when every observed class was read; not_observed ("not
// exercised") when the session has no such action, or when the only read is of
// a stage packet whose result does not show the leaf whole (packetNamesLeaf). Only the first action of a
// class is checked, so a read lost to a compaction is not seen. Reads are
// those namesLeaf accepts.
func leafReadBeforeAction(s auditSession) criterionAssessment {
	c := criterionAssessment{Criterion: "leaf_read_before_action", Status: "not_observed", Evidence: []string{}}
	seen := map[string]bool{}
	var failed, review, unverified, passed []string
	for i, e := range s.Trace.Events {
		for _, a := range leafActionsIn(e, s.Cwd) {
			if seen[a.class] {
				continue
			}
			seen[a.class] = true
			read, withheld, unseen := a.readInCommand, false, false
			prior := s.Trace.Events[:i]
			for j, p := range prior {
				if namesLeaf(p, a.leaf) {
					read = true
					break
				}
				packetRead, packetUnverified := packetNamesLeaf(prior, j, a.leaf)
				if packetRead {
					read = true
					break
				}
				unseen = unseen || packetUnverified
				if p.Kind == "shell" && p.CommandWithheld {
					withheld = true
				}
			}
			line := regressionEvidence(e) + " " + a.class + " (" + a.leaf + ")"
			switch {
			case read:
				passed = append(passed, line)
			case withheld:
				review = append(review, line)
			case unseen:
				unverified = append(unverified, line+": a stage packet was read but its result does not show the leaf whole")
			default:
				failed = append(failed, line)
			}
		}
	}
	switch {
	case len(failed) > 0:
		c.Status, c.Evidence = "fail", failed
	case len(review) > 0:
		c.Status, c.Evidence = "review", review
	case len(unverified) > 0:
		c.Evidence = unverified
	case len(passed) > 0:
		c.Status, c.Evidence = "pass", passed
	}
	return c
}
