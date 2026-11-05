package cursor

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload))})
}

func TestConversationIsTheSession(t *testing.T) {
	got := translate("sessionStart", `{"conversation_id":"c-1","session_id":"s-9","workspace_roots":["/w","/x"],"hook_event_name":"sessionStart","cursor_version":"2026.10.22"}`)
	if len(got) != 1 || got[0].Host != "cursor" || got[0].SessionID != "c-1" || got[0].Cwd != "/w" || got[0].Map != hostevent.MapStart {
		t.Fatalf("got %+v", got)
	}
}

func TestBeforeSubmitPromptRegistersTheSessionToo(t *testing.T) {
	got := translate("beforeSubmitPrompt", `{"conversation_id":"c-1","workspace_roots":["/w"],"prompt":"x"}`)
	if got[0].Type != hostevent.Activity || got[0].Map != hostevent.MapTouch || !got[0].Replaces {
		t.Fatalf("Cursor must also map on beforeSubmitPrompt: %+v", got)
	}
}

func TestSessionEndUsesTheConversation(t *testing.T) {
	got := translate("sessionEnd", `{"conversation_id":"c-1","session_id":"s-9","workspace_roots":["/w"]}`)
	if got[0].SessionID != "c-1" || got[0].Map != hostevent.MapEnd {
		t.Fatalf("got %+v", got)
	}
}

func TestIsHookPayload(t *testing.T) {
	if !IsHookPayload(hostevent.Decode([]byte(`{"conversation_id":"c"}`))) || !IsHookPayload(hostevent.Decode([]byte(`{"cursor_version":"1","session_id":"s"}`))) {
		t.Fatal("conversation_id or cursor_version is Cursor")
	}
	if IsHookPayload(hostevent.Decode([]byte(`{"session_id":"s"}`))) {
		t.Fatal("a Claude-shaped payload is not Cursor")
	}
}
