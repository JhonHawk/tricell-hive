package cursor

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

// sessionStart and sessionEnd carry both conversation_id and session_id and may
// carry no cursor_version, so a payload alone cannot tell them from Claude
// Code's: the hook's host argument decides then. The environment never does,
// because a Claude Code or Codex session started from a Cursor terminal
// inherits Cursor's variables.
func TestIsHookPayloadDetectsCursorFromPayloadFieldsOnly(t *testing.T) {
	ambiguous := `{"conversation_id":"c","session_id":"s"}`
	if IsHookPayload(hostevent.Decode([]byte(ambiguous))) {
		t.Fatal("nothing in this payload tells it from Claude Code's")
	}
	for _, raw := range []string{
		`{"conversation_id":"c","session_id":"s","generation_id":"g"}`,
		`{"session_id":"s","workspace_roots":["/p"]}`,
		`{"conversation_id":"c"}`,
		`{"cursor_version":"1.2.3","session_id":"s"}`,
	} {
		if !IsHookPayload(hostevent.Decode([]byte(raw))) {
			t.Fatalf("not detected: %s", raw)
		}
	}
	if IsHookPayload(hostevent.Decode([]byte(`{"session_id":"s","cwd":"/p"}`))) {
		t.Fatal("a Claude Code payload is not Cursor's")
	}
}
