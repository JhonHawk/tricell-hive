package cursor

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Cursor CLI.
const HookHost = "cursor"

// IsHookPayload reports whether a payload alone says Cursor CLI ran the hook:
// it carries cursor_version, generation_id, workspace_roots, or a
// conversation_id without a session_id, also when it runs a hook registered for
// Claude Code. sessionStart and sessionEnd carry both IDs and may carry no
// cursor_version, so the environment is read too (IsHook).
func IsHookPayload(payload map[string]any) bool {
	return hostevent.Has(payload, "cursor_version") || hostevent.Has(payload, "generation_id") || hostevent.Has(payload, "workspace_roots") ||
		hostevent.Has(payload, "conversation_id") && !hostevent.Has(payload, "session_id")
}

// IsHook is IsHookPayload plus the variables Cursor sets for its hooks.
func IsHook(payload map[string]any, getenv func(string) string) bool {
	if IsHookPayload(payload) {
		return true
	}
	return getenv != nil && (getenv("CURSOR_VERSION") != "" || getenv("CURSOR_PROJECT_DIR") != "")
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
		Turn:           hostevent.Turn(p, "generation_id"),
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
