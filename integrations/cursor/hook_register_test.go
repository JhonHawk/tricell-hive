package cursor

import (
	"path/filepath"
	"testing"

	"tricell-hive/integrations/target"
)

func TestSessionEventsFileIsFlatAndVersioned(t *testing.T) {
	home := t.TempDir()
	f := SessionEventsFile(target.Config{Scope: "user", Home: home, CursorHome: filepath.Join(home, ".cursor")})
	if f.Path != filepath.Join(home, ".cursor", "hooks.json") || f.Word != "cursor" || !f.Flat || f.RootVersion != 1 {
		t.Fatalf("file = %+v", f)
	}
	want := []string{"sessionStart", "beforeSubmitPrompt", "stop", "sessionEnd"}
	if len(f.Events) != len(want) {
		t.Fatalf("events = %v", f.Events)
	}
	for i := range want {
		if f.Events[i] != want[i] {
			t.Fatalf("events = %v, want %v", f.Events, want)
		}
	}
}
