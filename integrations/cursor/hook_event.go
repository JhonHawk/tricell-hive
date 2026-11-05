package cursor

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Cursor CLI.
const HookHost = "cursor"

// IsHookPayload reports whether a hook invocation comes from Cursor CLI: its
// payload carries cursor_version or a conversation_id, also when it runs a hook
// registered for Claude Code.
func IsHookPayload(payload map[string]any) bool {
	return hostevent.Has(payload, "cursor_version") || hostevent.Has(payload, "conversation_id") && !hostevent.Has(payload, "session_id")
}

// HookEvents translates one Cursor CLI hook payload (JSON on stdin:
// conversation_id, workspace_roots, transcript_path; session_id on sessionStart
// and sessionEnd) per _support/docs/architecture/host-session-events.md. The
// conversation is the session. sessionStart does not fire when a chat resumes,
// so beforeSubmitPrompt also registers the session in the map. Cursor has no
// native task events.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:           HookHost,
		SessionID:      hostevent.Str(p, "conversation_id", "session_id"),
		Cwd:            hostevent.FirstString(p, "workspace_roots"),
		TranscriptPath: hostevent.Str(p, "transcript_path"),
		OccurredAt:     hostevent.When(p, "timestamp", "occurred_at"),
	}
	if base.Cwd == "" {
		base.Cwd = hostevent.Str(p, "cwd")
	}
	if base.SessionID == "" {
		return nil
	}
	event := in.Event
	if native := hostevent.Str(p, "hook_event_name"); native != "" {
		event = native
	}
	return []hostevent.Event{hostevent.Generic(base, hostevent.Normalize(event), true)}
}
