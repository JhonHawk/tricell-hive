package pi

import "tricell-hive/integrations/hostevent"

// HookHost is the canonical host name the viewer uses for Pi.
const HookHost = "pi"

// HookEvents translates the envelope Hive's Pi extension pipes to
// `hive hook pi <event>`: {"session_id", "cwd", "occurred_at"?}, with the event
// being session_start, agent_settled or session_shutdown. The extension takes the
// session ID from ctx.sessionManager.getSessionId() and passes it explicitly, so
// no event touches the process map. Pi has no native task events. The extension
// fills occurred_at from the event's own time; the hook never reads its clock.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:       HookHost,
		SessionID:  hostevent.Str(p, "session_id", "sessionId"),
		Cwd:        hostevent.Str(p, "cwd"),
		OccurredAt: hostevent.When(p, "occurred_at", "timestamp"),
	}
	if base.SessionID == "" {
		return nil
	}
	e := hostevent.Generic(base, hostevent.Normalize(in.Event), false)
	e.Map = hostevent.MapNone
	return []hostevent.Event{e}
}
