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
		// Codex caps SessionEnd at 3 s; larger values emit a startup warning.
		Timeouts: map[string]int{"SessionEnd": 3},
		// Codex's /hooks screen names a hook by its statusMessage.
		StatusMessages: map[string]string{
			"SessionStart":     "Hive: report session start to the viewer",
			"UserPromptSubmit": "Hive: report prompt activity to the viewer",
			"SubagentStart":    "Hive: report subagent start to the viewer",
			"SessionEnd":       "Hive: report session end to the viewer",
		},
	}
}
