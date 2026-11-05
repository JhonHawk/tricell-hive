package claude

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Claude Code.
const HookHost = "claude-code"

// HookEvents translates one Claude Code hook payload (snake_case JSON on stdin:
// session_id, cwd, transcript_path; task_id and task_subject on TaskCreated and
// TaskCompleted; agent_id and agent_type on SubagentStart and SubagentStop) per
// _support/docs/architecture/host-session-events.md. Claude Code runs one
// session per process, so a start always replaces the previous one (/clear and
// resume write a new session_id). Subagents with an empty agent_type are
// internal agents and are ignored.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:           HookHost,
		SessionID:      hostevent.Str(p, "session_id"),
		Cwd:            hostevent.Str(p, "cwd"),
		TranscriptPath: hostevent.Str(p, "transcript_path"),
		OccurredAt:     hostevent.When(p, "timestamp", "occurred_at"),
		Turn:           hostevent.Turn(p, "prompt_id", "turn_id"),
	}
	if base.Cwd == "" {
		base.Cwd = in.Env("CLAUDE_PROJECT_DIR")
	}
	if base.SessionID == "" {
		return nil
	}
	norm := hostevent.Normalize(in.Event)
	switch norm {
	case "taskcreated", "taskcompleted":
		typ, state := hostevent.TaskCreated, hostevent.StateNotStarted
		if norm == "taskcompleted" {
			typ, state = hostevent.TaskCompleted, hostevent.StateCompleted
		}
		e, ok := base.Task(typ, hostevent.ID(p["task_id"]), hostevent.Str(p, "task_subject"), state)
		if !ok {
			return nil
		}
		return []hostevent.Event{e}
	case "subagentstart", "subagentstop":
		if hostevent.Str(p, "agent_type") == "" {
			return nil
		}
		typ := hostevent.SubagentStarted
		if norm == "subagentstop" {
			typ = hostevent.SubagentStopped
		}
		e, ok := base.Subagent(typ, hostevent.Str(p, "agent_id"))
		if !ok {
			return nil
		}
		return []hostevent.Event{e}
	}
	return []hostevent.Event{hostevent.Generic(base, norm, true)}
}
