package hostevent_test

import (
	"testing"
	"tricell-hive/integrations/claude"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/cursor"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/hostevent"
)

func TestTurnReadsOnlyAnIDTheViewerAccepts(t *testing.T) {
	p := hostevent.Decode([]byte(`{"a":"has space","b":"ok-1","c":42}`))
	if got := hostevent.Turn(p, "a", "b"); got != "ok-1" {
		t.Fatalf("got %q", got)
	}
	if got := hostevent.Turn(p, "c"); got != "42" {
		t.Fatalf("got %q", got)
	}
	if got := hostevent.Turn(p, "missing"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestEveryHostKeepsItsTurnDiscriminatorOnActivity(t *testing.T) {
	cases := []struct {
		name     string
		events   []hostevent.Event
		wantTurn string
	}{
		{"grok promptId", grok.HookEvents(hostevent.Input{Event: "UserPromptSubmit", Payload: hostevent.Decode([]byte(`{"sessionId":"s","cwd":"/p","promptId":"pr-1"}`))}), "pr-1"},
		{"cursor generation_id", cursor.HookEvents(hostevent.Input{Event: "beforeSubmitPrompt", Payload: hostevent.Decode([]byte(`{"conversation_id":"c","generation_id":"gen-1","workspace_roots":["/p"]}`))}), "gen-1"},
		{"claude prompt_id", claude.HookEvents(hostevent.Input{Event: "UserPromptSubmit", Payload: hostevent.Decode([]byte(`{"session_id":"s","cwd":"/p","prompt_id":"pr-2"}`))}), "pr-2"},
		{"codex turn_id", codex.HookEvents(hostevent.Input{Event: "Stop", Payload: hostevent.Decode([]byte(`{"session_id":"s","cwd":"/p","turn_id":"turn-3"}`))}), "turn-3"},
	}
	for _, c := range cases {
		if len(c.events) != 1 || c.events[0].Turn != c.wantTurn {
			t.Errorf("%s: %+v", c.name, c.events)
		}
	}
}
