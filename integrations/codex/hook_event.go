package codex

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Codex.
const HookHost = "codex"

// HookEvents translates one Codex hook payload (snake_case JSON on stdin:
// session_id, cwd, transcript_path; agent_id on hooks a subagent triggers, with
// the parent's session_id) per _support/docs/architecture/host-session-events.md.
// An event carrying agent_id belongs to a subagent and never changes the main
// session: SubagentStart and SubagentStop report the agent, every other event
// of a subagent is dropped. Codex has no native task events.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:           HookHost,
		SessionID:      hostevent.Str(p, "session_id"),
		Cwd:            hostevent.Str(p, "cwd"),
		TranscriptPath: hostevent.Str(p, "transcript_path"),
		OccurredAt:     hostevent.When(p, "timestamp", "occurred_at"),
	}
	if base.SessionID == "" {
		return nil
	}
	norm := hostevent.Normalize(in.Event)
	if agent := hostevent.ID(p["agent_id"]); agent != "" {
		typ := hostevent.SubagentStarted
		switch norm {
		case "subagentstart":
		case "subagentstop":
			typ = hostevent.SubagentStopped
		default:
			return nil
		}
		e, ok := base.Subagent(typ, agent)
		if !ok {
			return nil
		}
		return []hostevent.Event{e}
	}
	if norm == "subagentstart" || norm == "subagentstop" {
		return nil
	}
	return []hostevent.Event{hostevent.Generic(base, norm, true)}
}
