package codex

import (
	"path/filepath"
	"testing"

	"tricell-hive/integrations/target"
)

func TestSessionEventsFileNeverRegistersTheStopEvents(t *testing.T) {
	home := t.TempDir()
	f := SessionEventsFile(target.Config{Scope: "user", Home: home, CodexHome: filepath.Join(home, ".codex")})
	if f.Path != filepath.Join(home, ".codex", "hooks.json") || f.Word != "codex" || f.Flat {
		t.Fatalf("file = %+v", f)
	}
	want := map[string]bool{"SessionStart": true, "UserPromptSubmit": true, "SubagentStart": true, "SessionEnd": true}
	if len(f.Events) != len(want) {
		t.Fatalf("events = %v", f.Events)
	}
	for _, e := range f.Events {
		if !want[e] {
			t.Fatalf("unexpected event %s: Stop and SubagentStop expect JSON output and are never registered", e)
		}
	}
}
