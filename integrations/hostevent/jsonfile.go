package hostevent

// JSONHookFile describes one JSON hook file that Hive adds its own entries to
// and that other tools share: Claude Code's settings.json, Codex's hooks.json
// and Cursor's hooks.json. The manager edits it entry by entry and never writes
// the rest; the host packages say where the file is and which events to
// register (_support/docs/architecture/host-session-events.md).
type JSONHookFile struct {
	// Word is the host word of the command: `hive hook <Word> <event>`.
	Word string
	Path string
	// Flat is true when an event's array holds the handler objects themselves
	// (Cursor); otherwise each element is a matcher group with a "hooks" array
	// (Claude Code, Codex).
	Flat bool
	// Events are the event names to register, in file order.
	Events []string
	// RootVersion is the "version" a file created by Hive starts with, or 0 for
	// none (Cursor's hooks.json needs `"version": 1`).
	RootVersion int
	// StatusMessages maps an event to the "statusMessage" its entry carries, the
	// name Codex's /hooks screen shows. An event without a text, or a nil map,
	// writes no field: Claude Code's settings.json is also read by Grok, which
	// records a load error for it, and Cursor's file has no such field.
	StatusMessages map[string]string
}

type groupHandler struct {
	Type          string `json:"type"`
	Command       string `json:"command"`
	Timeout       int    `json:"timeout"`
	StatusMessage string `json:"statusMessage,omitempty"`
}

type group struct {
	Hooks []groupHandler `json:"hooks"`
}

type flatHandler struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout"`
}

// Entry returns the value Hive appends to the event's array: a matcher group
// holding one command handler, or the bare handler for a flat file.
func (f JSONHookFile) Entry(executable, event string) any {
	command := Command(executable, f.Word, event)
	if f.Flat {
		return flatHandler{Command: command, Timeout: HookTimeoutSeconds}
	}
	return group{Hooks: []groupHandler{{Type: "command", Command: command, Timeout: HookTimeoutSeconds, StatusMessage: f.StatusMessages[event]}}}
}

// Commands returns the commands of an entry as the file holds them: the one
// handler command of a group, or of a flat handler. Anything else is nil.
func (f JSONHookFile) Commands(entry map[string]any) []string {
	if f.Flat {
		if c, ok := entry["command"].(string); ok {
			return []string{c}
		}
		return nil
	}
	handlers, ok := entry["hooks"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, h := range handlers {
		m, ok := h.(map[string]any)
		if !ok {
			return nil
		}
		c, _ := m["command"].(string)
		out = append(out, c)
	}
	return out
}
