package cursor

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

// sessionStart and sessionEnd carry both conversation_id and session_id, and
// may carry no cursor_version: Cursor is still told apart.
func TestIsHookDetectsCursorBeyondTheVersionField(t *testing.T) {
	none := func(string) string { return "" }
	cursorPayloads := []string{
		`{"conversation_id":"c","session_id":"s"}` + "", // only the environment can tell
		`{"conversation_id":"c","session_id":"s","generation_id":"g"}`,
		`{"session_id":"s","workspace_roots":["/p"]}`,
	}
	if IsHook(hostevent.Decode([]byte(cursorPayloads[0])), none) {
		t.Fatal("nothing tells this payload from Claude Code's without the environment")
	}
	if !IsHook(hostevent.Decode([]byte(cursorPayloads[0])), func(k string) string {
		if k == "CURSOR_VERSION" {
			return "1.2.3"
		}
		return ""
	}) {
		t.Fatal("CURSOR_VERSION")
	}
	if !IsHook(hostevent.Decode([]byte(cursorPayloads[0])), func(k string) string {
		if k == "CURSOR_PROJECT_DIR" {
			return "/p"
		}
		return ""
	}) {
		t.Fatal("CURSOR_PROJECT_DIR")
	}
	for _, raw := range cursorPayloads[1:] {
		if !IsHook(hostevent.Decode([]byte(raw)), none) {
			t.Fatalf("not detected: %s", raw)
		}
	}
	if IsHook(hostevent.Decode([]byte(`{"session_id":"s","cwd":"/p"}`)), none) {
		t.Fatal("a Claude Code payload is not Cursor's")
	}
}
