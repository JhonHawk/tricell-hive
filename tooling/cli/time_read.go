package main

// time_read.go reads a host's whole transcript for `hive time`, line by line,
// keeping only times, the role or origin that tells a person's message from a
// machine's, and the commands that wait for CI. It does not reuse the tail read
// of readHostSessionFile: a total needs every line. The formats are observed
// on the maintainer's machine (Claude Code 2.1.x, Codex rollouts of
// 2026-11, Grok updates.jsonl, Pi version 3), not documented.
//
// A user wait is the span from the end of the agent's turn to the next message
// a person wrote, and the span of a question card from its call to its answer.
// A CI wait is the span of a command that waits for CI or a remote deployment.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

type timeFacts struct {
	first, last time.Time
	waits       []timeSpan
	ci          []timeSpan
	partial     bool
	note        string
}

// ciCommand matches the commands that wait for CI or a remote deployment.
var ciCommand = regexp.MustCompile(`gh\s+pr\s+checks\b[^;&|\n]*--watch|gh\s+run\s+watch\b`)

func (f *timeFacts) see(t time.Time) {
	if t.IsZero() {
		return
	}
	if f.first.IsZero() || t.Before(f.first) {
		f.first = t
	}
	if t.After(f.last) {
		f.last = t
	}
}

func (f *timeFacts) markPartial(note string) {
	if !f.partial {
		f.partial, f.note = true, note
	}
}

// idleTracker follows whether the agent's turn has ended, so that the next
// message a person writes closes a user wait.
type idleTracker struct {
	idle  bool
	since time.Time
}

func (i *idleTracker) end(t time.Time) { i.idle, i.since = true, t }
func (i *idleTracker) busy()           { i.idle = false }

func (i *idleTracker) human(f *timeFacts, t time.Time) {
	if i.idle && t.After(i.since) {
		f.waits = append(f.waits, timeSpan{i.since, t})
	}
	i.idle = false
}

// readTimeFile reads the whole file; what it cannot read whole makes the
// result partial, never a smaller total passed off as complete.
func readTimeFile(host, path string) timeFacts {
	var f timeFacts
	file, _, err := openSessionFile(path)
	if err != nil {
		f.markPartial("the session file could not be opened")
		return f
	}
	defer file.Close()
	var handle func(line []byte) error
	switch host {
	case "claude-code":
		handle = newClaudeTimeReader(&f)
	case "codex":
		handle = newCodexTimeReader(&f)
	case "grok":
		handle = newGrokTimeReader(&f)
	case "pi":
		handle = newPiTimeReader(&f)
		f.markPartial("Pi does not mark the prompts its extensions inject, so a user wait cannot be told from a machine wait")
	default:
		f.markPartial("no reader for this host")
		return f
	}
	br := bufio.NewReaderSize(file, 1<<20)
	for n := 1; ; n++ {
		line, rerr := br.ReadBytes('\n')
		complete := rerr == nil
		if len(line) > 0 {
			trimmed := trimLine(line)
			if len(trimmed) > 0 {
				if herr := handle(trimmed); herr != nil {
					if complete {
						f.markPartial(fmt.Sprintf("line %d of the session file is unreadable", n))
					} else {
						f.markPartial("the session file ends in a truncated line")
					}
				}
			}
		}
		if rerr != nil {
			if rerr != io.EOF {
				f.markPartial("the session file could not be read to its end")
			}
			break
		}
	}
	return f
}

func trimLine(b []byte) []byte {
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == '\r') {
		b = b[:len(b)-1]
	}
	return b
}

func parseTimeStamp(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// textOf is the text of a content that is a string or blocks with text.
func textOf(raw json.RawMessage) (text string, hasImage bool, hasToolResult bool) {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, false, false
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return "", false, false
	}
	var parts []string
	for _, b := range blocks {
		switch b.Type {
		case "text", "input_text":
			parts = append(parts, b.Text)
		case "image":
			hasImage = true
		case "tool_result":
			hasToolResult = true
		}
	}
	return strings.Join(parts, "\n"), hasImage, hasToolResult
}

// --- Claude Code ---

var claudeToolUseID = regexp.MustCompile(`<tool-use-id>([^<]+)</tool-use-id>`)

// claudeMachinePrefixes open the user entries Claude Code writes itself when
// they carry no origin (older versions).
var claudeMachinePrefixes = []string{"<task-notification", "<system-reminder", "<local-command", "[Request interrupted", "<tools-used"}

type claudeTimeEntry struct {
	Type             string          `json:"type"`
	Subtype          string          `json:"subtype"`
	Timestamp        string          `json:"timestamp"`
	IsMeta           bool            `json:"isMeta"`
	IsSidechain      bool            `json:"isSidechain"`
	IsCompactSummary bool            `json:"isCompactSummary"`
	Origin           json.RawMessage `json:"origin"`
	Message          struct {
		Content    json.RawMessage `json:"content"`
		StopReason string          `json:"stop_reason"`
	} `json:"message"`
}

func newClaudeTimeReader(f *timeFacts) func([]byte) error {
	var idle idleTracker
	questions := map[string]time.Time{}
	ciCalls := map[string]time.Time{}
	ciBackground := map[string]time.Time{}
	return func(line []byte) error {
		var e claudeTimeEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return err
		}
		ts := parseTimeStamp(e.Timestamp)
		// The summary Claude Code writes when the person is away comes after the
		// session's last activity, not as part of it.
		if !(e.Type == "system" && e.Subtype == "away_summary") {
			f.see(ts)
		}
		if e.IsSidechain || ts.IsZero() {
			return nil
		}
		switch e.Type {
		case "assistant":
			var blocks []struct {
				Type  string          `json:"type"`
				ID    string          `json:"id"`
				Name  string          `json:"name"`
				Input json.RawMessage `json:"input"`
			}
			if json.Unmarshal(e.Message.Content, &blocks) == nil {
				for _, b := range blocks {
					if b.Type != "tool_use" {
						continue
					}
					switch b.Name {
					case "AskUserQuestion":
						questions[b.ID] = ts
					case "Bash":
						var in struct {
							Command    string `json:"command"`
							Background bool   `json:"run_in_background"`
						}
						if json.Unmarshal(b.Input, &in) == nil && ciCommand.MatchString(in.Command) {
							if in.Background {
								ciBackground[b.ID] = ts
							} else {
								ciCalls[b.ID] = ts
							}
						}
					}
				}
			}
			if e.Message.StopReason == "end_turn" {
				idle.end(ts)
			} else {
				idle.busy()
			}
		case "system":
			if e.Subtype == "turn_duration" {
				idle.end(ts)
			}
		case "user":
			claudeUser(f, &idle, e, ts, questions, ciCalls, ciBackground)
		}
		return nil
	}
}

func claudeUser(f *timeFacts, idle *idleTracker, e claudeTimeEntry, ts time.Time, questions, ciCalls, ciBackground map[string]time.Time) {
	var blocks []struct {
		Type      string `json:"type"`
		ToolUseID string `json:"tool_use_id"`
	}
	hasResult := false
	if json.Unmarshal(e.Message.Content, &blocks) == nil {
		for _, b := range blocks {
			if b.Type != "tool_result" {
				continue
			}
			hasResult = true
			if start, ok := questions[b.ToolUseID]; ok {
				f.waits = append(f.waits, timeSpan{start, ts})
				delete(questions, b.ToolUseID)
			}
			if start, ok := ciCalls[b.ToolUseID]; ok {
				f.ci = append(f.ci, timeSpan{start, ts})
				delete(ciCalls, b.ToolUseID)
			}
		}
	}
	if hasResult {
		idle.busy()
		return
	}
	var origin struct {
		Kind string `json:"kind"`
	}
	hasOrigin := len(e.Origin) > 0 && string(e.Origin) != "null" && json.Unmarshal(e.Origin, &origin) == nil
	if hasOrigin && origin.Kind == "task-notification" {
		text, _, _ := textOf(e.Message.Content)
		if m := claudeToolUseID.FindStringSubmatch(text); m != nil {
			if start, ok := ciBackground[m[1]]; ok {
				f.ci = append(f.ci, timeSpan{start, ts})
				delete(ciBackground, m[1])
			}
		}
		idle.busy()
		return
	}
	if e.IsMeta || e.IsCompactSummary {
		return
	}
	text, hasImage, _ := textOf(e.Message.Content)
	if hasOrigin {
		if origin.Kind == "human" {
			idle.human(f, ts)
		}
		return
	}
	trimmed := strings.TrimSpace(text)
	if strings.HasPrefix(trimmed, "[Request interrupted") {
		idle.end(ts)
		return
	}
	if trimmed == "" && !hasImage {
		return
	}
	for _, p := range claudeMachinePrefixes {
		if strings.HasPrefix(trimmed, p) {
			return
		}
	}
	// No origin: the entry of an older version, or a command a person typed.
	idle.human(f, ts)
}

// --- Codex ---

var (
	codexSessionID = regexp.MustCompile(`"session_id":\s*(\d+)`)
	codexCellID    = regexp.MustCompile(`Script running with cell ID (\S+)`)
	codexNumber    = regexp.MustCompile(`\d{3,}`)
)

type codexTimeLine struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Payload   struct {
		Type      string          `json:"type"`
		Role      string          `json:"role"`
		Content   json.RawMessage `json:"content"`
		Name      string          `json:"name"`
		CallID    string          `json:"call_id"`
		Arguments string          `json:"arguments"`
		Input     json.RawMessage `json:"input"`
		Output    json.RawMessage `json:"output"`
		Message   string          `json:"message"`
	} `json:"payload"`
}

type codexPending struct {
	start    time.Time
	question bool
	ci       bool
}

// codexHumanText says whether a user message is one a person typed: Codex
// injects its context and the instruction files as user messages that open
// with markup or the instruction heading.
// codexMachinePrefixes open the user-role messages Codex injects itself. Any
// other user text is a person's, including the tags Codex wraps a person's
// image, review action or card answer in.
var codexMachinePrefixes = []string{"# AGENTS.md instructions", "<environment_context", "<skill", "<user_instructions", "<turn_aborted"}

func codexHumanText(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	for _, prefix := range codexMachinePrefixes {
		if strings.HasPrefix(t, prefix) {
			return false
		}
	}
	return true
}

func newCodexTimeReader(f *timeFacts) func([]byte) error {
	var idle idleTracker
	pending := map[string]*codexPending{}
	sessions := map[string]bool{}
	cells := map[string]bool{}
	execNames := map[string]bool{"exec": true, "exec_command": true, "shell": true, "shell_command": true}
	return func(line []byte) error {
		var e codexTimeLine
		if err := json.Unmarshal(line, &e); err != nil {
			return err
		}
		ts := parseTimeStamp(e.Timestamp)
		f.see(ts)
		if ts.IsZero() {
			return nil
		}
		p := e.Payload
		switch e.Type {
		case "event_msg":
			switch p.Type {
			case "task_complete", "turn_aborted":
				idle.end(ts)
			case "user_message":
				if codexHumanText(p.Message) {
					idle.human(f, ts)
				}
			}
		case "response_item":
			switch p.Type {
			case "message":
				if p.Role == "user" {
					text, _, _ := textOf(p.Content)
					if codexHumanText(text) {
						idle.human(f, ts)
					}
				} else if p.Role == "assistant" {
					idle.busy()
				}
			case "reasoning":
				idle.busy()
			case "function_call", "custom_tool_call":
				idle.busy()
				input := p.Arguments
				if p.Type == "custom_tool_call" {
					_ = json.Unmarshal(p.Input, &input)
				}
				switch {
				case p.Name == "request_user_input":
					pending[p.CallID] = &codexPending{start: ts, question: true}
				case p.Name == "wait":
					var a struct {
						Cell string `json:"cell_id"`
					}
					if json.Unmarshal([]byte(p.Arguments), &a) == nil && cells[a.Cell] {
						pending[p.CallID] = &codexPending{start: ts, ci: true}
					}
				case execNames[p.Name] || p.Name == "write_stdin":
					if ciCommand.MatchString(input) {
						pending[p.CallID] = &codexPending{start: ts, ci: true}
					} else if len(sessions) > 0 {
						for _, n := range codexNumber.FindAllString(input, -1) {
							if sessions[n] {
								pending[p.CallID] = &codexPending{start: ts, ci: true}
								break
							}
						}
					}
				}
			case "function_call_output", "custom_tool_call_output":
				idle.busy()
				pd, ok := pending[p.CallID]
				if !ok {
					break
				}
				delete(pending, p.CallID)
				if pd.question {
					f.waits = append(f.waits, timeSpan{pd.start, ts})
					break
				}
				f.ci = append(f.ci, timeSpan{pd.start, ts})
				out := toolResultText(p.Output)
				for _, m := range codexSessionID.FindAllStringSubmatch(out, -1) {
					sessions[m[1]] = true
				}
				if m := codexCellID.FindStringSubmatch(out); m != nil {
					cells[m[1]] = true
				}
			}
		}
		return nil
	}
}

// --- Grok ---

type grokTimeLine struct {
	Timestamp float64 `json:"timestamp"`
	Params    struct {
		Meta struct {
			AgentTimestampMs float64 `json:"agentTimestampMs"`
		} `json:"_meta"`
		Update struct {
			Kind       string          `json:"sessionUpdate"`
			Status     string          `json:"status"`
			ToolCallID string          `json:"toolCallId"`
			Title      string          `json:"title"`
			RawInput   json.RawMessage `json:"rawInput"`
			Content    struct {
				Text string `json:"text"`
			} `json:"content"`
			Meta struct {
				Synthetic bool `json:"syntheticPrompt"`
			} `json:"_meta"`
			TaskID       string `json:"task_id"`
			TaskCallID   string `json:"tool_call_id"`
			TaskSnapshot struct {
				TaskID string `json:"task_id"`
			} `json:"task_snapshot"`
		} `json:"update"`
	} `json:"params"`
}

func grokTime(l grokTimeLine) time.Time {
	if ms := l.Params.Meta.AgentTimestampMs; ms > 0 {
		return time.UnixMilli(int64(ms)).UTC()
	}
	if l.Timestamp > 0 {
		return time.Unix(int64(l.Timestamp), 0).UTC()
	}
	return time.Time{}
}

func newGrokTimeReader(f *timeFacts) func([]byte) error {
	var idle idleTracker
	questions := map[string]time.Time{}
	ciCalls := map[string]time.Time{}
	ciTasks := map[string]time.Time{}
	return func(line []byte) error {
		var l grokTimeLine
		if err := json.Unmarshal(line, &l); err != nil {
			return err
		}
		ts := grokTime(l)
		f.see(ts)
		if ts.IsZero() {
			return nil
		}
		u := l.Params.Update
		switch u.Kind {
		case "user_message_chunk":
			if u.Meta.Synthetic {
				idle.busy()
			} else {
				idle.human(f, ts)
			}
		case "turn_completed":
			idle.end(ts)
		case "agent_message_chunk", "agent_thought_chunk":
			idle.busy()
		case "tool_call":
			idle.busy()
			if u.Title == "ask_user_question" {
				questions[u.ToolCallID] = ts
				break
			}
			var in struct {
				Command string `json:"command"`
				Cmd     string `json:"cmd"`
			}
			if json.Unmarshal(l.Params.Update.RawInput, &in) == nil && (ciCommand.MatchString(in.Command) || ciCommand.MatchString(in.Cmd)) {
				ciCalls[u.ToolCallID] = ts
			}
		case "tool_call_update":
			if u.Status != "completed" && u.Status != "failed" {
				break
			}
			if start, ok := questions[u.ToolCallID]; ok {
				f.waits = append(f.waits, timeSpan{start, ts})
				delete(questions, u.ToolCallID)
			}
			if start, ok := ciCalls[u.ToolCallID]; ok {
				f.ci = append(f.ci, timeSpan{start, ts})
				delete(ciCalls, u.ToolCallID)
			}
		case "task_backgrounded":
			if start, ok := ciCalls[u.TaskCallID]; ok {
				ciTasks[u.TaskID] = start
				delete(ciCalls, u.TaskCallID)
			}
		case "task_completed":
			if start, ok := ciTasks[u.TaskSnapshot.TaskID]; ok {
				f.ci = append(f.ci, timeSpan{start, ts})
				delete(ciTasks, u.TaskSnapshot.TaskID)
			}
		}
		return nil
	}
}

// --- Pi ---

type piTimeLine struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		Role       string          `json:"role"`
		StopReason string          `json:"stopReason"`
		ToolCallID string          `json:"toolCallId"`
		Content    json.RawMessage `json:"content"`
	} `json:"message"`
}

// newPiTimeReader reads a Pi transcript. A message with role user is taken as a
// person's, but Pi extensions inject user prompts of their own with nothing
// that tells them apart, so readTimeFile marks every Pi result partial.
func newPiTimeReader(f *timeFacts) func([]byte) error {
	var idle idleTracker
	questions := map[string]time.Time{}
	ciCalls := map[string]time.Time{}
	return func(line []byte) error {
		var l piTimeLine
		if err := json.Unmarshal(line, &l); err != nil {
			return err
		}
		ts := parseTimeStamp(l.Timestamp)
		f.see(ts)
		if ts.IsZero() || l.Type != "message" {
			return nil
		}
		switch l.Message.Role {
		case "user":
			idle.human(f, ts)
		case "assistant":
			var blocks []struct {
				Type      string          `json:"type"`
				ID        string          `json:"id"`
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if json.Unmarshal(l.Message.Content, &blocks) == nil {
				for _, b := range blocks {
					if b.Type != "toolCall" {
						continue
					}
					if b.Name == "ask_user_question" {
						questions[b.ID] = ts
					} else if b.Name == "bash" {
						var in struct {
							Command string `json:"command"`
						}
						if json.Unmarshal(b.Arguments, &in) == nil && ciCommand.MatchString(in.Command) {
							ciCalls[b.ID] = ts
						}
					}
				}
			}
			if l.Message.StopReason == "toolUse" {
				idle.busy()
			} else {
				idle.end(ts)
			}
		case "toolResult":
			idle.busy()
			if start, ok := questions[l.Message.ToolCallID]; ok {
				f.waits = append(f.waits, timeSpan{start, ts})
				delete(questions, l.Message.ToolCallID)
			}
			if start, ok := ciCalls[l.Message.ToolCallID]; ok {
				f.ci = append(f.ci, timeSpan{start, ts})
				delete(ciCalls, l.Message.ToolCallID)
			}
		}
		return nil
	}
}
