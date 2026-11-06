package claude

import (
	"path/filepath"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// SessionEventsFile is where Hive registers `hive hook claude <event>`: the
// user settings file, which other tools share. SessionEnd shares a 1.5 s budget
// unless its entry sets a larger timeout, and every entry carries one
// (hostevent.HookTimeoutSeconds). TaskCreated and TaskCompleted are the native
// task events; the rest report the session and its activity.
func SessionEventsFile(c target.Config) hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word: "claude",
		Path: filepath.Join(c.ClaudeHome, "settings.json"),
		Events: []string{
			"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop",
			"TaskCreated", "TaskCompleted", "SessionEnd",
		},
	}
}
