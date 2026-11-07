package antigravity

import (
	"path/filepath"
	"testing"

	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

func translate(event, payload string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload))})
}

const samplePayload = `{"conversationId":"ec33ebf9-0cba-4100-8142-c61503f6c587","workspacePaths":["/work/app","/work/other"],"transcriptPath":"/work/app/.gemini/antigravity-cli/transcript.jsonl","modelName":"auto","lastUserInput":"hi","invocationNum":3,"executionNum":1,"terminationReason":"NO_TOOL_CALL"}`

func TestSessionStartIsTheCanonicalStartInWorkspacePathsZero(t *testing.T) {
	got := translate("SessionStart", samplePayload)
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	e := got[0]
	if e.Host != "antigravity" || e.SessionID != "ec33ebf9-0cba-4100-8142-c61503f6c587" || e.Cwd != "/work/app" || e.Type != hostevent.SessionStarted || e.Map != hostevent.MapStart {
		t.Fatalf("got %+v", e)
	}
}

func TestPreInvocationIsActivityThatRefreshesTheMap(t *testing.T) {
	got := translate("PreInvocation", samplePayload)
	if len(got) != 1 || got[0].Type != hostevent.Activity || got[0].Map != hostevent.MapTouch || got[0].Cwd != "/work/app" || got[0].Host != "antigravity" || got[0].Turn != "3" {
		t.Fatalf("got %+v", got)
	}
}

func TestStopIsTheEndOfATurnAndNeverTheEndOfASession(t *testing.T) {
	got := translate("Stop", samplePayload)
	if len(got) != 1 || got[0].Type != hostevent.Activity || got[0].Map != hostevent.MapNone || got[0].Host != "antigravity" || got[0].Cwd != "/work/app" {
		t.Fatalf("got %+v", got)
	}
	for _, name := range []string{"Stop", "SessionEnd", "session_end", "PostInvocation", "PreToolUse", "nope"} {
		for _, e := range translate(name, samplePayload) {
			if e.Type == hostevent.SessionEnded || e.Map == hostevent.MapEnd {
				t.Fatalf("%s must never end a session: %+v", name, e)
			}
		}
	}
	if len(translate("SessionEnd", samplePayload)) != 0 {
		t.Fatal("an event agy does not have produces nothing")
	}
}

func TestPayloadWithoutAConversationProducesNothing(t *testing.T) {
	if got := translate("SessionStart", `{"workspacePaths":["/w"]}`); got != nil {
		t.Fatalf("got %+v", got)
	}
}

func TestReplyIsTheAcceptedObjectOfTheEvent(t *testing.T) {
	for event, want := range map[string]string{"SessionStart": "{}", "PreInvocation": "{}", "Stop": `{"decision":"stop"}`, "stop": `{"decision":"stop"}`, "": "{}", "Weird": "{}"} {
		if got := Reply(event); got != want {
			t.Errorf("Reply(%q) = %s, want %s", event, got, want)
		}
	}
}

func TestSessionEventsFileIsAFlatNamedHookWithoutPreToolUse(t *testing.T) {
	home := t.TempDir()
	f := SessionEventsFile(target.Config{Scope: "user", Home: home, AntigravityHome: filepath.Join(home, ".gemini", "config")})
	if f.Path != filepath.Join(home, ".gemini", "config", "hooks.json") || f.Word != "antigravity" || !f.Flat || f.ContainerKey() != "hive" {
		t.Fatalf("file = %+v", f)
	}
	want := []string{"SessionStart", "PreInvocation", "Stop"}
	if len(f.Events) != len(want) {
		t.Fatalf("events = %v", f.Events)
	}
	for i := range want {
		if f.Events[i] != want[i] {
			t.Fatalf("events = %v, want %v", f.Events, want)
		}
	}
}
