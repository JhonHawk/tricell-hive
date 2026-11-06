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

// Payloads Grok 1.0.77 sent for one subagent (probe of 2026-11-06, paths and
// prompts replaced): the parent's SubagentStart names the child in subagentId,
// and every event the child fires carries the child's own sessionId and its
// subagentType.
const (
	grokParent           = "01a1af4a-da28-7392-96ac-a9a742b64c77"
	grokChild            = "01a1af4b-00d7-7e52-9250-02badec6ce88"
	grokSubagentStart    = `{"cwd":"/w","description":"read","hookEventName":"subagent_start","permissionMode":"bypassPermissions","sessionId":"` + grokParent + `","session_id":"` + grokParent + `","subagentId":"` + grokChild + `","subagentType":"general-purpose","timestamp":"2026-11-06T05:18:10.800Z","transcriptPath":"/t/p/updates.jsonl","workspaceRoot":"/w"}`
	grokChildPrompt      = `{"cwd":"/w","hookEventName":"user_prompt_submit","permissionMode":"bypassPermissions","prompt":"x","promptId":"01a1af4b-01b6-7d71-a398-244c4b996772","sessionId":"` + grokChild + `","session_id":"` + grokChild + `","subagentType":"general-purpose","timestamp":"2026-11-06T05:18:10.900Z","workspaceRoot":"/w"}`
	grokChildSubagentEnd = `{"cwd":"/w","hookEventName":"subagent_stop","lastAssistantMessage":"probe","phase":"gate","sessionId":"` + grokChild + `","session_id":"` + grokChild + `","subagentId":"` + grokChild + `","subagentType":"general-purpose","timestamp":"2026-11-06T05:18:18.100Z","transcriptPath":"/t/c/updates.jsonl","workspaceRoot":"/w"}`
	grokChildSessionEnd  = `{"cwd":"/w","hookEventName":"session_end","reason":"shutdown","sessionId":"` + grokChild + `","session_id":"` + grokChild + `","subagentType":"general-purpose","timestamp":"2026-11-06T05:18:18.200Z","transcriptPath":"/t/c/updates.jsonl","workspaceRoot":"/w"}`
)

func TestSubagentStartIsSubagentActivityOfTheParent(t *testing.T) {
	got := translate("SubagentStart", grokSubagentStart, map[string]string{"GROK_HOOK_EVENT": "subagent_start"})
	if len(got) != 1 || got[0].Type != hostevent.SubagentStarted || got[0].SessionID != grokParent || got[0].AgentID != grokChild || got[0].Map != hostevent.MapNone {
		t.Fatalf("got %+v", got)
	}
}

func TestEventsOfTheSubagentsOwnSessionAreDropped(t *testing.T) {
	for _, c := range []struct{ event, native, payload string }{
		{"UserPromptSubmit", "user_prompt_submit", grokChildPrompt},
		{"SubagentStop", "subagent_stop", grokChildSubagentEnd},
		{"SessionEnd", "session_end", grokChildSessionEnd},
		// The same events reaching Hive through a hook registered for Claude Code.
		{"UserPromptSubmit", "", grokChildPrompt},
		{"SessionEnd", "", grokChildSessionEnd},
	} {
		env := map[string]string{"GROK_HOOK_EVENT": c.native}
		if got := translate(c.event, c.payload, env); len(got) != 0 {
			t.Fatalf("%s (%q) must not reach the viewer or the map: %+v", c.event, c.native, got)
		}
	}
}

func TestSubagentEventsWithoutATypeAreDropped(t *testing.T) {
	for _, event := range []string{"SubagentStart", "SubagentStop"} {
		if got := translate(event, `{"sessionId":"g-1","subagentId":"c-1"}`, nil); len(got) != 0 {
			t.Fatalf("%s without subagentType: %+v", event, got)
		}
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
