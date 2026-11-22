package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// mainThreadCost is the informational breakdown of a session's main thread that
// --audit-session reports: model calls, processed input, compactions, tool
// calls and the share spent reading Hive's installed skills or running their
// scripts. It has no pass or fail and no criterion reads it.
type mainThreadCost struct {
	ModelCalls   int   `json:"model_calls"`
	InputTokens  int64 `json:"input_tokens"`
	CachedTokens int64 `json:"cached_tokens"`
	// PeakContext is the largest single-call input; only hosts that report usage per call have it.
	PeakContext int64 `json:"peak_context,omitempty"`
	// UsageSource is per_call, per_turn, per_turn_with_children (a child's usage was
	// missing, so the parent total still includes its children) or none.
	UsageSource string         `json:"usage_source"`
	Compactions int            `json:"compactions"`
	ToolCalls   int            `json:"tool_calls"`
	HiveReads   int            `json:"hive_reads"`
	HiveScripts int            `json:"hive_scripts"`
	ByTool      map[string]int `json:"by_tool"`
	// RepeatedUsageEvents counts Codex token_count events whose total did not change.
	RepeatedUsageEvents int `json:"repeated_usage_events,omitempty"`
}

func newCost() mainThreadCost { return mainThreadCost{UsageSource: "none", ByTool: map[string]int{}} }

var hiveSkillPath = regexp.MustCompile(`([^\s"'` + "`" + `]*/)skills/(?:[^/\s"'` + "`" + `\\]+/(SKILL\.md|stage(?:-[a-z]+)?\.md|references/|scripts/[^\s"'` + "`" + `\\]+\.py)|hive-core/` + leafDomainDir + `/[^\s"'` + "`" + `\\]*\.md)`)
var hiveReadCommand = regexp.MustCompile(`\b(cat|sed|head|tail|rg|grep)\b`)

// hiveCall classifies a tool call as a read of an installed Hive skill ("read"),
// a run of one of its scripts ("script") or neither (""). Only a skill under a
// hidden folder of the user (such as ~/.agents/skills) counts, never a
// repository's content/skills.
func hiveCall(tool, path, command string) string {
	installed := func(text string) (m []string) {
		for _, found := range hiveSkillPath.FindAllStringSubmatch(text, -1) {
			if strings.Contains(found[1], "/.") && !strings.HasSuffix(found[1], "content/") {
				return found
			}
		}
		return nil
	}
	if toolKind(tool) == "read" {
		if m := installed(path); m != nil && !strings.HasPrefix(m[2], "scripts/") {
			return "read"
		}
		return ""
	}
	if m := installed(command); m != nil {
		if strings.HasPrefix(m[2], "scripts/") {
			return "script"
		}
		if hiveReadCommand.MatchString(command) {
			return "read"
		}
	}
	return ""
}

func (c *mainThreadCost) toolCall(name, path, command string) {
	c.ToolCalls++
	c.ByTool[name]++
	switch hiveCall(name, path, command) {
	case "read":
		c.HiveReads++
	case "script":
		c.HiveScripts++
	}
}

func costLines(data []byte, fn func(line []byte)) {
	scan := auditScanner(data)
	for scan.Scan() {
		fn(scan.Bytes())
	}
}

func claudeCost(data []byte) mainThreadCost {
	c := newCost()
	counted, tools := map[string]bool{}, map[string]bool{}
	costLines(data, func(raw []byte) {
		var ev struct {
			Type, Subtype string
			IsSidechain   bool
			Message       struct {
				ID      string
				Content json.RawMessage
			}
		}
		var shaped struct {
			Message struct {
				Usage struct {
					InputTokens              int64 `json:"input_tokens"`
					CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
					CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
				} `json:"usage"`
			} `json:"message"`
		}
		if json.Unmarshal(raw, &ev) != nil || ev.IsSidechain {
			return
		}
		if ev.Type == "system" && strings.Contains(ev.Subtype, "compact") {
			c.Compactions++
		}
		if ev.Type != "assistant" {
			return
		}
		_ = json.Unmarshal(raw, &shaped)
		u := shaped.Message.Usage
		total := u.InputTokens + u.CacheReadInputTokens + u.CacheCreationInputTokens
		if total > 0 && !counted[ev.Message.ID] {
			if ev.Message.ID != "" {
				counted[ev.Message.ID] = true
			}
			c.ModelCalls++
			c.InputTokens += total
			c.CachedTokens += u.CacheReadInputTokens
			c.PeakContext = max(c.PeakContext, total)
		}
		var blocks []struct {
			Type, ID, Name string
			Input          map[string]any
		}
		if json.Unmarshal(ev.Message.Content, &blocks) != nil {
			return
		}
		for _, b := range blocks {
			if b.Type != "tool_use" || (b.ID != "" && tools[b.ID]) {
				continue
			}
			tools[b.ID] = true
			c.toolCall(b.Name, first(b.Input, "file_path", "filePath", "path"), first(b.Input, "command", "cmd"))
		}
	})
	if c.ModelCalls > 0 {
		c.UsageSource = "per_call"
	}
	return c
}

// codexCallCommand returns the shell command of a Codex function_call, whose
// arguments are a JSON string holding cmd or command as a string or an argv.
func codexCallCommand(arguments string) string {
	var args map[string]json.RawMessage
	if json.Unmarshal([]byte(arguments), &args) != nil {
		return ""
	}
	for _, key := range []string{"cmd", "command"} {
		var text string
		if json.Unmarshal(args[key], &text) == nil {
			return text
		}
		var argv []string
		if json.Unmarshal(args[key], &argv) == nil {
			return strings.Join(argv, " ")
		}
	}
	return ""
}

func codexCost(data []byte) mainThreadCost {
	c := newCost()
	lastTotal := ""
	costLines(data, func(raw []byte) {
		var ev struct {
			Type    string
			Payload struct {
				Type, Name, Input string
				Arguments         string
				Info              *struct {
					Total json.RawMessage `json:"total_token_usage"`
					Last  struct {
						Input  int64 `json:"input_tokens"`
						Cached int64 `json:"cached_input_tokens"`
					} `json:"last_token_usage"`
				}
			}
		}
		if json.Unmarshal(raw, &ev) != nil {
			return
		}
		p := ev.Payload
		switch {
		case ev.Type == "compacted":
			c.Compactions++
		case ev.Type == "event_msg" && p.Type == "token_count" && p.Info != nil:
			if total := string(p.Info.Total); total == lastTotal {
				c.RepeatedUsageEvents++
			} else {
				lastTotal = total
				c.ModelCalls++
				c.InputTokens += p.Info.Last.Input
				c.CachedTokens += p.Info.Last.Cached
				c.PeakContext = max(c.PeakContext, p.Info.Last.Input)
			}
		case ev.Type == "response_item" && p.Type == "custom_tool_call":
			// An exec block's source holds the commands it runs.
			c.toolCall(p.Name, "", p.Input)
		case ev.Type == "response_item" && p.Type == "function_call":
			c.toolCall(p.Name, "", codexCallCommand(p.Arguments))
		}
	})
	if c.ModelCalls > 0 {
		c.UsageSource = "per_call"
	}
	return c
}

func grokCost(data []byte, path, sessionsRoot string) mainThreadCost {
	c := newCost()
	var turns bool
	costLines(data, func(raw []byte) {
		var ev struct {
			Params struct{ Update map[string]any }
		}
		if json.Unmarshal(raw, &ev) != nil {
			return
		}
		u := ev.Params.Update
		switch str(u["sessionUpdate"]) {
		case "auto_compact_completed":
			c.Compactions++
		case "turn_completed":
			if usage := object(u["usage"]); usage != nil {
				turns = true
				c.ModelCalls += int(num(usage["modelCalls"]))
				c.InputTokens += int64(num(usage["inputTokens"]))
				c.CachedTokens += int64(num(usage["cachedReadTokens"]))
			}
		case "tool_call":
			name := str(object(object(u["_meta"])["x.ai/tool"])["name"])
			if name == "" {
				name = str(u["title"])
			}
			args := object(u["rawInput"])
			name, args = grokToolCall(name, args)
			c.toolCall(name, first(args, "file_path", "filePath", "path", "target_file", "target_directory"), first(args, "command", "cmd"))
		}
	})
	if !turns {
		return c
	}
	c.UsageSource = "per_turn"
	calls, input, cached, ok := grokChildrenUsage(filepath.Dir(path), sessionsRoot)
	if !ok {
		c.UsageSource = "per_turn_with_children"
		return c
	}
	c.ModelCalls -= calls
	c.InputTokens -= input
	c.CachedTokens -= cached
	return c
}

func num(v any) float64 {
	f, _ := v.(float64)
	return f
}

// grokChildrenUsage sums the usage.json of every child a Grok session started
// (subagents/*/meta.json, child_session_id), looked up in any project folder of
// the sessions root because a child can live in another directory's folder. ok
// is false when a child's usage cannot be found, so the caller subtracts none.
func grokChildrenUsage(sessionDir, sessionsRoot string) (calls int, input, cached int64, ok bool) {
	metas, _ := filepath.Glob(filepath.Join(sessionDir, "subagents", "*", "meta.json"))
	if len(metas) == 0 {
		return 0, 0, 0, true
	}
	if sessionsRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return 0, 0, 0, false
		}
		sessionsRoot = filepath.Join(home, ".grok", "sessions")
	}
	for _, meta := range metas {
		raw, err := os.ReadFile(meta)
		var m struct {
			Child string `json:"child_session_id"`
		}
		if err != nil || json.Unmarshal(raw, &m) != nil || m.Child == "" || strings.ContainsAny(m.Child, `/\`) {
			return 0, 0, 0, false
		}
		hits, _ := filepath.Glob(filepath.Join(sessionsRoot, "*", m.Child, "usage.json"))
		if len(hits) == 0 {
			return 0, 0, 0, false
		}
		raw, err = os.ReadFile(hits[0])
		var u struct {
			Session struct {
				ModelCalls       int   `json:"modelCalls"`
				InputTokens      int64 `json:"inputTokens"`
				CachedReadTokens int64 `json:"cachedReadTokens"`
			}
		}
		if err != nil || json.Unmarshal(raw, &u) != nil {
			return 0, 0, 0, false
		}
		calls += u.Session.ModelCalls
		input += u.Session.InputTokens
		cached += u.Session.CachedReadTokens
	}
	return calls, input, cached, true
}
