package grok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"tricell-hive/integrations/hostevent"
)

// SessionEventsName is the second native hook file Hive owns in <GrokHome>/hooks.
// Grok merges every file in that folder, so it sits beside the read-whole hook
// and neither file knows about the other.
const SessionEventsName = "hive-session-events.json"

// SessionEventsSource names the generated payload of the session events hook.
// Like ReadHookSource it is not a file of content/.
const SessionEventsSource = "integrations/grok/session-events-hook"

// SessionEventsPath returns where the session events hook lives for a Grok home.
func SessionEventsPath(grokHome string) string {
	return filepath.Join(grokHome, "hooks", SessionEventsName)
}

// sessionEventNames are the Grok events registered. Stop is passive for the
// hook, which prints nothing. SubagentStart reports a subagent as activity of
// its parent even when Grok does not also run Claude Code's hooks.
var sessionEventNames = []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "Stop", "SessionEnd"}

type programHandler struct {
	Type    string   `json:"type"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Timeout int      `json:"timeout"`
}

type programGroup struct {
	Hooks []programHandler `json:"hooks"`
}

// RenderSessionEventsHook returns the hook file that reports Grok's session
// events. Each handler names the program in "command" and its arguments in
// "args", which Grok runs without a shell, so the path needs no quoting and a
// space or an apostrophe in it is safe. The path comes from a plan that may have
// been edited by hand, so it is checked here as well.
func RenderSessionEventsHook(executable string) ([]byte, error) {
	if !filepath.IsAbs(executable) || strings.ContainsRune(executable, 0) {
		return nil, fmt.Errorf("Grok session events hook executable must be an absolute path without NUL")
	}
	hooks := map[string][]programGroup{}
	for _, event := range sessionEventNames {
		hooks[event] = []programGroup{{Hooks: []programHandler{{
			Type:    "command",
			Command: executable,
			Args:    []string{"hook", HookHost, event},
			Timeout: hostevent.HookTimeoutSeconds,
		}}}}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(struct {
		Hooks map[string][]programGroup `json:"hooks"`
	}{hooks}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
