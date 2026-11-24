package grok

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Grok Build.
const HookHost = "grok"

// IsHookPayload reports whether a hook invocation comes from Grok: Grok sets
// GROK_HOOK_EVENT, and its payload is camelCase (sessionId) even when it runs a
// hook registered for Claude Code (compat.claude.hooks).
func IsHookPayload(payload map[string]any, getenv func(string) string) bool {
	if getenv != nil && getenv("GROK_HOOK_EVENT") != "" {
		return true
	}
	return hostevent.Has(payload, "sessionId") && !hostevent.Has(payload, "session_id")
}

// HookEvents translates one Grok hook payload (camelCase JSON on stdin:
// sessionId, cwd, workspaceRoot) per _support/docs/architecture/host-session-events.md.
// A Grok subagent runs in the parent's process under its own sessionId, and
// every event that fires inside it carries subagentType. Such an event never
// changes the map or reports another session: the parent's SubagentStart, which
// carries the parent's sessionId and names the child in subagentId, becomes
// subagent activity of the parent, and every other subagent event is dropped.
// Grok has no native task events, so a task event is plain activity. The
// snake_case keys are read too, for a payload shaped like Claude Code's.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:           HookHost,
		SessionID:      hostevent.Str(p, "sessionId", "session_id"),
		Cwd:            hostevent.Str(p, "cwd", "workspaceRoot", "workspace_root"),
		TranscriptPath: hostevent.Str(p, "transcriptPath", "transcript_path"),
		OccurredAt:     hostevent.When(p, "timestamp", "occurredAt", "occurred_at"),
		Turn:           hostevent.Turn(p, "promptId", "prompt_id"),
	}
	if base.SessionID == "" {
		return nil
	}
	event := in.Event
	if native := in.Env("GROK_HOOK_EVENT"); native != "" {
		event = native
	}
	norm := hostevent.Normalize(event)
	if hostevent.Str(p, "subagentType", "subagent_type") != "" {
		if norm != "subagentstart" {
			return nil
		}
		e, ok := base.Subagent(hostevent.SubagentStarted, hostevent.Str(p, "subagentId", "subagent_id"))
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
