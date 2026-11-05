package grok

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string, env map[string]string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload)), Getenv: func(k string) string { return env[k] }})
}

func TestCamelCasePayload(t *testing.T) {
	got := translate("sessionStart", `{"sessionId":"g-1","cwd":"/w","workspaceRoot":"/w"}`, nil)
	if len(got) != 1 || got[0].Host != "grok" || got[0].SessionID != "g-1" || got[0].Cwd != "/w" || got[0].Map != hostevent.MapStart || !got[0].Replaces {
		t.Fatalf("got %+v", got)
	}
}

func TestNativeEventNameWinsOverTheRegisteredOne(t *testing.T) {
	// A hook registered for Claude Code's SessionStart runs under Grok's own event.
	got := translate("SessionStart", `{"sessionId":"g-1","cwd":"/w"}`, map[string]string{"GROK_HOOK_EVENT": "session_end"})
	if got[0].Type != hostevent.SessionEnded || got[0].Map != hostevent.MapEnd {
		t.Fatalf("got %+v", got)
	}
}

func TestWorkspaceRootIsTheFallbackDirectory(t *testing.T) {
	got := translate("UserPromptSubmit", `{"sessionId":"g-1","workspaceRoot":"/ws"}`, nil)
	if got[0].Cwd != "/ws" || got[0].Type != hostevent.Activity {
		t.Fatalf("got %+v", got)
	}
}

func TestTaskEventsAreOnlyActivity(t *testing.T) {
	got := translate("TaskCreated", `{"sessionId":"g-1","task_id":"1","task_subject":"x"}`, nil)
	if got[0].Type != hostevent.Activity || got[0].TaskID != "" {
		t.Fatalf("Grok has no native tasks: %+v", got)
	}
}

func TestIsHookPayload(t *testing.T) {
	none := func(string) string { return "" }
	if !IsHookPayload(hostevent.Decode([]byte(`{"sessionId":"x"}`)), none) {
		t.Fatal("camelCase sessionId is Grok")
	}
	if !IsHookPayload(hostevent.Decode([]byte(`{"session_id":"x"}`)), func(k string) string {
		if k == "GROK_HOOK_EVENT" {
			return "session_start"
		}
		return ""
	}) {
		t.Fatal("GROK_HOOK_EVENT is Grok")
	}
	if IsHookPayload(hostevent.Decode([]byte(`{"session_id":"x"}`)), none) {
		t.Fatal("a snake_case payload with no Grok variable is not Grok")
	}
}
