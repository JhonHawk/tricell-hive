package antigravity

import (
	"path/filepath"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// HookName is Hive's own named hook in hooks.json: agy keeps every tool's hooks
// under a top-level name, so Hive adds, changes and removes only this key.
const HookName = "hive"

// SessionEventsFile is where Hive registers `hive hook antigravity <event>`:
// hooks.json in the Antigravity home, `{"hive": {"SessionStart": [...], ...}}`
// with the handler objects directly under each event (agy's flat events). The
// event is not in the payload, so each command carries it as an argument.
// PreToolUse is not registered: its output must hold a `decision`, and a hook
// that fails there would block tools. The format is the one in agy's own
// hooks documentation (agy-customizations, observed with agy 1.5.2).
func SessionEventsFile(c target.Config) hostevent.JSONHookFile {
	return hostevent.JSONHookFile{
		Word:        "antigravity",
		Path:        filepath.Join(c.AntigravityHome, "hooks.json"),
		Flat:        true,
		Container:   HookName,
		HandlerType: "command",
		Events:      []string{"SessionStart", "PreInvocation", "Stop"},
	}
}
