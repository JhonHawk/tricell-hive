package cursor

import (
	"path/filepath"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// SessionEventsFile is where Hive registers `hive hook cursor <event>`:
// hooks.json in the Cursor home, `{"version": 1, "hooks": {...}}` with the
// handler objects directly under each event. beforeSubmitPrompt also feeds the
// session map because sessionStart does not fire when a chat resumes. The file
// format is verified in the installed code only; Cursor CLI's use of it is not
// documented.
func SessionEventsFile(c target.Config) hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word:        "cursor",
		Path:        filepath.Join(c.CursorHome, "hooks.json"),
		Flat:        true,
		RootVersion: 1,
		Events:      []string{"sessionStart", "beforeSubmitPrompt", "stop", "sessionEnd"},
	}
}
