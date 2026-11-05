package claude

import (
	"path/filepath"
	"testing"

	"tricell-hive/integrations/target"
)

func TestSessionEventsFileRegistersTheNativeTaskEventsInTheUserSettings(t *testing.T) {
	home := t.TempDir()
	f := SessionEventsFile(target.Config{Scope: "user", Home: home, ClaudeHome: filepath.Join(home, ".claude")})
	if f.Path != filepath.Join(home, ".claude", "settings.json") || f.Word != "claude" || f.Flat {
		t.Fatalf("file = %+v", f)
	}
	have := map[string]bool{}
	for _, e := range f.Events {
		have[e] = true
	}
	for _, e := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "TaskCreated", "TaskCompleted", "SessionEnd"} {
		if !have[e] {
			t.Fatalf("missing event %s in %v", e, f.Events)
		}
	}
}
