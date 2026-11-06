package codex

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload))})
}

func TestMainThreadEventsMapTheSession(t *testing.T) {
	start := translate("SessionStart", `{"session_id":"t-1","cwd":"/w","transcript_path":"/r.jsonl","hook_event_name":"SessionStart"}`)
	if len(start) != 1 || start[0].Host != "codex" || start[0].Map != hostevent.MapStart || !start[0].Replaces || start[0].Type != hostevent.SessionStarted {
		t.Fatalf("got %+v", start)
	}
	if end := translate("SessionEnd", `{"session_id":"t-1","cwd":"/w"}`); end[0].Map != hostevent.MapEnd {
		t.Fatalf("got %+v", end)
	}
	if p := translate("UserPromptSubmit", `{"session_id":"t-1","cwd":"/w"}`); p[0].Type != hostevent.Activity {
		t.Fatalf("got %+v", p)
	}
}

func TestSubagentEventsDoNotChangeTheMainSession(t *testing.T) {
	// SubagentStart/Stop report the agent under the parent's session and never touch the map.
	got := translate("SubagentStart", `{"session_id":"t-1","cwd":"/w","agent_id":"ag-9"}`)
	if len(got) != 1 || got[0].Type != hostevent.SubagentStarted || got[0].AgentID != "ag-9" || got[0].Map != hostevent.MapNone || got[0].Replaces {
		t.Fatalf("got %+v", got)
	}
	if got = translate("SubagentStop", `{"session_id":"t-1","agent_id":"ag-9"}`); got[0].Type != hostevent.SubagentStopped {
		t.Fatalf("got %+v", got)
	}
	// Any other event carrying agent_id is the subagent's own: dropped.
	for _, event := range []string{"SessionStart", "SessionEnd", "UserPromptSubmit"} {
		if translate(event, `{"session_id":"t-1","agent_id":"ag-9"}`) != nil {
			t.Fatalf("%s with agent_id must not change the main session", event)
		}
	}
	// A subagent event with no agent_id cannot be attributed.
	if translate("SubagentStart", `{"session_id":"t-1"}`) != nil {
		t.Fatal("a subagent event without agent_id must be dropped")
	}
}
