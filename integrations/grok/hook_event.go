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
// Grok subagents share the parent's process and session, and Grok has no native
// task events, so a task or subagent event is plain activity. The snake_case
// keys are read too, for a payload shaped like Claude Code's.
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
	return []hostevent.Event{hostevent.Generic(base, hostevent.Normalize(event), true)}
}
