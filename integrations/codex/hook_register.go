package codex

import (
	"path/filepath"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// SessionEventsFile is where Hive registers `hive hook codex <event>`:
// hooks.json in the Codex home. Stop and SubagentStop are never registered:
// Codex expects JSON on their standard output, and the hook prints nothing.
// Hive never writes the trust Codex asks for; the user approves the hook in
// /hooks.
func SessionEventsFile(c target.Config) hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word:   "codex",
		Path:   filepath.Join(c.CodexHome, "hooks.json"),
		Events: []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SessionEnd"},
	}
}
