package claude

import (
	"testing"
	"tricell-hive/integrations/hostevent"
)

func translate(event, payload string, env map[string]string) []hostevent.Event {
	return HookEvents(hostevent.Input{Event: event, Payload: hostevent.Decode([]byte(payload)), Getenv: func(k string) string { return env[k] }})
}

func TestSessionStartReplacesTheMapSession(t *testing.T) {
	got := translate("SessionStart", `{"session_id":"s-1","cwd":"/work/app","transcript_path":"/t/s-1.jsonl","source":"clear"}`, nil)
	if len(got) != 1 {
		t.Fatalf("got %+v", got)
	}
	e := got[0]
	if e.Host != "claude-code" || e.SessionID != "s-1" || e.Cwd != "/work/app" || e.Type != hostevent.SessionStarted || e.Map != hostevent.MapStart || !e.Replaces {
		t.Fatalf("got %+v", e)
	}
}

func TestSessionEndAndPromptAndProjectDirFallback(t *testing.T) {
	end := translate("SessionEnd", `{"session_id":"s-1","cwd":"/w","reason":"clear"}`, nil)
	if end[0].Type != hostevent.SessionEnded || end[0].Map != hostevent.MapEnd {
		t.Fatalf("got %+v", end)
	}
	prompt := translate("UserPromptSubmit", `{"session_id":"s-1","prompt":"secret text"}`, map[string]string{"CLAUDE_PROJECT_DIR": "/proj"})
	if prompt[0].Type != hostevent.Activity || prompt[0].Cwd != "/proj" {
		t.Fatalf("got %+v", prompt)
	}
}

func TestTaskEventsCarryIDAndSubject(t *testing.T) {
	c := translate("TaskCreated", `{"session_id":"s-1","cwd":"/w","task_id":"3","task_subject":"Write the tests"}`, nil)
	if c[0].Type != hostevent.TaskCreated || c[0].TaskID != "3" || c[0].Title != "Write the tests" || c[0].State != hostevent.StateNotStarted {
		t.Fatalf("got %+v", c)
	}
	d := translate("TaskCompleted", `{"session_id":"s-1","cwd":"/w","task_id":3,"task_subject":"Write the tests"}`, nil)
	if d[0].Type != hostevent.TaskCompleted || d[0].TaskID != "3" || d[0].State != hostevent.StateCompleted {
		t.Fatalf("a numeric task_id must work: %+v", d)
	}
	if translate("TaskCreated", `{"session_id":"s-1","task_subject":"no id"}`, nil) != nil {
		t.Fatal("a task without an ID must be dropped")
	}
}

func TestSubagentsWithEmptyAgentTypeAreIgnored(t *testing.T) {
	if translate("SubagentStart", `{"session_id":"s-1","agent_id":"a1","agent_type":""}`, nil) != nil || translate("SubagentStop", `{"session_id":"s-1","agent_id":"a1"}`, nil) != nil {
		t.Fatal("internal agents must be ignored")
	}
	got := translate("SubagentStart", `{"session_id":"s-1","agent_id":"a1","agent_type":"Explore"}`, nil)
	if len(got) != 1 || got[0].Type != hostevent.SubagentStarted || got[0].AgentID != "a1" || got[0].Map != hostevent.MapNone {
		t.Fatalf("got %+v", got)
	}
}

func TestNoSessionIDNoEvent(t *testing.T) {
	if translate("SessionStart", `{"cwd":"/w"}`, nil) != nil {
		t.Fatal("a payload without a session must give no event")
	}
}
