package opencode

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"tricell-hive/integrations/hostevent"
)

// HookHost is the canonical host name the viewer uses for OpenCode.
const HookHost = "opencode"

// maxTodos bounds the todos one todowrite call turns into events; the viewer
// reflects at most 50 native tasks per work.
const maxTodos = 50

// TodoID derives the native task ID of a todo from its content, because
// OpenCode 2.2.5 gives todos none. The same text is the same task across
// todowrite calls.
func TodoID(content string) string {
	sum := sha256.Sum256([]byte(strings.Join(strings.Fields(content), " ")))
	return "td-" + hex.EncodeToString(sum[:])[:16]
}

// HookEvents translates the envelope the OpenCode event plugin pipes to
// `hive hook opencode <event>`. All OpenCode sessions share one background
// service, so a process-ancestry map cannot tell them apart: the plugin names
// the session itself and no event touches the map. The envelope is
// {"session_id", "directory", "occurred_at"?, "todos"?:[{"content","status"}]};
// session.created, session.deleted and session.execution.* come from the plugin's
// event subscription, and todowrite from its execute.after tool hook. The
// plugin fills occurred_at from the event's own time (the hook never reads its
// clock). Camel-case keys (sessionID) are read too.
func HookEvents(in hostevent.Input) []hostevent.Event {
	p := in.Payload
	base := hostevent.Event{
		Host:       HookHost,
		SessionID:  hostevent.Str(p, "session_id", "sessionID", "sessionId"),
		Cwd:        hostevent.Str(p, "directory", "cwd"),
		OccurredAt: hostevent.When(p, "occurred_at", "timestamp", "time"),
	}
	if base.SessionID == "" {
		return nil
	}
	norm := hostevent.Normalize(in.Event)
	if norm != "todowrite" {
		e := hostevent.Generic(base, norm, false)
		e.Map = hostevent.MapNone
		return []hostevent.Event{e}
	}
	todos, _ := p["todos"].([]any)
	var out []hostevent.Event
	for _, raw := range todos {
		if len(out) >= maxTodos {
			break
		}
		todo, _ := raw.(map[string]any)
		content := hostevent.Str(todo, "content", "title")
		if content == "" {
			continue
		}
		typ, state := hostevent.TaskCreated, hostevent.StateNotStarted
		switch strings.ToLower(hostevent.Str(todo, "status")) {
		case "in_progress":
			typ, state = hostevent.TaskUpdated, hostevent.StateInProgress
		case "completed":
			typ, state = hostevent.TaskCompleted, hostevent.StateCompleted
		case "pending", "":
		default:
			// cancelled and unknown states are not a task state the viewer has.
			continue
		}
		if e, ok := base.Task(typ, TodoID(content), content, state); ok {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		// The call still proves the session is active.
		base.Type = hostevent.Activity
		return []hostevent.Event{base}
	}
	return out
}
