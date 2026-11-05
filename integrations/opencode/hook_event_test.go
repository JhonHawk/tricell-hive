package opencode

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload))})
}

func TestSessionEventsNameTheSessionAndNeverTouchTheMap(t *testing.T) {
	for event, want := range map[string]string{
		"session.created": hostevent.SessionStarted, "session.deleted": hostevent.SessionEnded,
		"session.execution.started": hostevent.Activity,
	} {
		got := translate(event, `{"session_id":"ses_1","directory":"/w","occurred_at":"2026-11-04T10:00:00Z"}`)
		if len(got) != 1 || got[0].Type != want || got[0].Map != hostevent.MapNone || got[0].SessionID != "ses_1" || got[0].Cwd != "/w" ||
			!got[0].OccurredAt.Equal(time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)) {
			t.Fatalf("%s: %+v", event, got)
		}
	}
}

func TestTodowriteMapsStatusesWithContentDerivedIDs(t *testing.T) {
	payload := `{"sessionID":"ses_1","directory":"/w","occurred_at":"2026-11-04T10:00:00Z","todos":[
		{"content":"Write tests","status":"pending"},
		{"content":"Fix  the bug","status":"in_progress"},
		{"content":"Ship it","status":"completed"},
		{"content":"Dropped","status":"cancelled"},
		{"content":"","status":"pending"}]}`
	got := translate("todowrite", payload)
	if len(got) != 3 {
		t.Fatalf("got %+v", got)
	}
	want := []struct{ typ, state string }{{hostevent.TaskCreated, hostevent.StateNotStarted}, {hostevent.TaskUpdated, hostevent.StateInProgress}, {hostevent.TaskCompleted, hostevent.StateCompleted}}
	for i, w := range want {
		if got[i].Type != w.typ || got[i].State != w.state || got[i].SessionID != "ses_1" || !strings.HasPrefix(got[i].TaskID, "td-") {
			t.Fatalf("todo %d: %+v", i, got[i])
		}
	}
	if got[1].TaskID != TodoID("Fix the bug") {
		t.Fatalf("whitespace must not change the ID: %q", got[1].TaskID)
	}
	if got[0].TaskID == got[1].TaskID {
		t.Fatal("different content must give different IDs")
	}
}

func TestTodowriteIsBoundedAndFallsBackToActivity(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"session_id":"ses_1","directory":"/w","todos":[`)
	for i := 0; i < 80; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"content":"task %d","status":"pending"}`, i)
	}
	b.WriteString("]}")
	if got := translate("todowrite", b.String()); len(got) != 50 {
		t.Fatalf("a todowrite is capped at 50 events, got %d", len(got))
	}
	got := translate("todowrite", `{"session_id":"ses_1","directory":"/w","todos":[]}`)
	if len(got) != 1 || got[0].Type != hostevent.Activity {
		t.Fatalf("got %+v", got)
	}
}

func TestNoSessionNoEvent(t *testing.T) {
	if translate("session.created", `{"directory":"/w"}`) != nil {
		t.Fatal("the plugin must pass the session ID")
	}
}
