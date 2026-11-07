package antigravity

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Antigravity (agy and
// the App).
const HookHost = "antigravity"

// StopReply is what `hive hook antigravity Stop` prints on every path: agy lets
// the agent stop on any decision other than "continue".
const StopReply = `{"decision":"stop"}`

// EmptyReply is the smallest object agy accepts from SessionStart and
// PreInvocation, and what Hive prints for an event it does not know.
const EmptyReply = `{}`

// Reply is the JSON object `hive hook antigravity <event>` prints for event
// whatever happens: agy reads stdout as the hook's verdict.
func Reply(event string) string {
	if hostevent.Normalize(event) == "stop" {
		return StopReply
	}
	return EmptyReply
}

// HookEvents translates one agy hook payload (camelCase JSON on stdin:
// conversationId, workspacePaths, transcriptPath; invocationNum on PreInvocation
// and executionNum on Stop) per _support/docs/architecture/host-session-events.md.
// The conversation is the session. The payload does not name its event, so the
// command argument decides. agy runs hooks in the folder of hooks.json, so the
// directory is workspacePaths[0], never the working directory.
//
// PreInvocation also registers the session in the map, since SessionStart fires
// only for a new conversation. Stop fires after every turn
// (terminationReason NO_TOOL_CALL) and agy has no session-end event, so Stop is
// plain activity and nothing here ever ends a session.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:           HookHost,
		SessionID:      hostevent.Str(p, "conversationId"),
		Cwd:            hostevent.FirstString(p, "workspacePaths"),
		TranscriptPath: hostevent.Str(p, "transcriptPath"),
		OccurredAt:     hostevent.When(p, "timestamp", "occurredAt"),
		Turn:           hostevent.Turn(p, "invocationNum", "executionNum"),
	}
	if base.SessionID == "" {
		return nil
	}
	switch hostevent.Normalize(in.Event) {
	case "sessionstart":
		return []hostevent.Event{hostevent.Generic(base, "sessionstart", true)}
	case "preinvocation":
		return []hostevent.Event{hostevent.Generic(base, "userpromptsubmit", true)}
	case "stop":
		base.Type = hostevent.Activity
		return []hostevent.Event{base}
	}
	return nil
}
