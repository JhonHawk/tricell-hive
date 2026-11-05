package pi

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload))})
}

func TestExtensionEventsNameTheSessionAndNeverTouchTheMap(t *testing.T) {
	for event, want := range map[string]string{
		"session_start": hostevent.SessionStarted, "agent_settled": hostevent.Activity, "session_shutdown": hostevent.SessionEnded,
	} {
		got := translate(event, `{"session_id":"01a1","cwd":"/w"}`)
		if len(got) != 1 || got[0].Type != want || got[0].Map != hostevent.MapNone || got[0].Host != "pi" || got[0].SessionID != "01a1" {
			t.Fatalf("%s: %+v", event, got)
		}
	}
}

func TestNoSessionNoEvent(t *testing.T) {
	if translate("session_start", `{"cwd":"/w"}`) != nil {
		t.Fatal("the extension must pass the session ID")
	}
}
