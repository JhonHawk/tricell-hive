// Package hostevent holds what every host's hook translator produces: one
// neutral Event per fact a coding host reported about a session. The translators
// live in each host's package (integrations/<host>/hook_event.go); the
// `hive hook` command in tooling/cli sends the events and records the session
// map. Nothing here prints, reads the clock or touches the file system.
package hostevent

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Map says what an event does to the process-to-session map
// (_support/docs/architecture/host-session-events.md, «Session map»).
type Map string

const (
	// MapNone leaves the map alone.
	MapNone Map = ""
	// MapStart registers the session as live. With Replaces the session takes the
	// place of the one the process had (a new session after /clear, a resume into
	// another session); without it the session starts alongside the others.
	MapStart Map = "start"
	// MapTouch registers the session only if the process has no live session
	// naming it, for hosts whose start event does not fire on resume.
	MapTouch Map = "touch"
	// MapEnd marks the session ended.
	MapEnd Map = "end"
)

// Viewer event types (tooling/viewer/timeline_contract_v3.go).
const (
	SessionStarted  = "session_started"
	Activity        = "activity"
	SessionEnded    = "session_ended"
	TaskCreated     = "task_created"
	TaskUpdated     = "task_updated"
	TaskCompleted   = "task_completed"
	SubagentStarted = "subagent_started"
	SubagentStopped = "subagent_stopped"
)

// Task states a host event may carry.
const (
	StateNotStarted = "not_started"
	StateInProgress = "in_progress"
	StateCompleted  = "completed"
)

// Event is one translated fact. Host is the canonical viewer host name
// (claude-code, codex, grok, opencode, pi, cursor, antigravity).
type Event struct {
	Host      string
	SessionID string
	// Cwd is the project directory the host reported; empty when it did not.
	Cwd string
	// Type is a viewer event type.
	Type string
	// OccurredAt is the time the host or a stable datum of the event gives; zero
	// when the payload has none. It is never the hook's clock.
	OccurredAt time.Time
	// TranscriptPath names a host-written file whose modification time is a
	// stable datum when OccurredAt is zero.
	TranscriptPath string
	Map            Map
	Replaces       bool
	TaskID, Title  string
	State          string
	AgentID        string
	// Turn is the host's own ID of the prompt or turn an activity belongs to,
	// when the payload carries one. It keeps two activity events of a host
	// that stamps no time distinct; a task event does not use it.
	Turn string
}

// Input is a hook invocation as the translator sees it: the event name the hook
// command was registered for, the decoded JSON payload and the environment.
type Input struct {
	Event   string
	Payload map[string]any
	Getenv  func(string) string
}

// Env reads a variable; a missing Getenv reads none.
func (in Input) Env(name string) string {
	if in.Getenv == nil {
		return ""
	}
	return in.Getenv(name)
}

// Decode parses one JSON object, keeping numbers exact. Anything else is nil.
func Decode(raw []byte) map[string]any {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	var m map[string]any
	if dec.Decode(&m) != nil {
		return nil
	}
	return m
}

// Normalize folds an event name to lower case letters and digits, so
// SessionStart, session_start, sessionStart and session.start compare equal.
func Normalize(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Str returns the first non-empty string among keys.
func Str(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

// Has reports whether m has key.
func Has(m map[string]any, key string) bool { _, ok := m[key]; return ok }

// FirstString returns the first string of a JSON array under key.
func FirstString(m map[string]any, key string) string {
	if list, ok := m[key].([]any); ok {
		for _, v := range list {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// ID renders a JSON scalar (string or number) as a native ID.
func ID(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case json.Number:
		return x.String()
	}
	return ""
}

// When reads the host's own timestamp under any of keys: RFC 3339 text, or a
// number of Unix seconds or milliseconds. Zero when absent or unreadable.
func When(m map[string]any, keys ...string) time.Time {
	for _, k := range keys {
		switch x := m[k].(type) {
		case string:
			if t, err := time.Parse(time.RFC3339Nano, x); err == nil {
				return t.UTC()
			}
		case json.Number:
			if f, err := strconv.ParseFloat(x.String(), 64); err == nil && f > 0 {
				return FromUnix(f)
			}
		}
	}
	return time.Time{}
}

// FromUnix reads seconds, or milliseconds when the value is past 1e11.
func FromUnix(f float64) time.Time {
	if f >= 1e11 {
		return time.UnixMilli(int64(f)).UTC()
	}
	return time.Unix(int64(f), int64((f-float64(int64(f)))*1e9)).UTC()
}

var nativeID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// Turn reads the host's prompt or turn ID under any of keys; empty when the
// payload has none or it is not an ID the viewer accepts.
func Turn(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if id := ID(m[k]); ValidNativeID(id) {
			return id
		}
	}
	return ""
}

// ValidNativeID is the shape the viewer accepts for a task or agent ID.
func ValidNativeID(id string) bool { return nativeID.MatchString(id) }

// Title cleans a task title for the viewer: control characters become spaces,
// the ends are trimmed and the text is cut at 120 characters.
func Title(s string) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 120 {
		s = strings.TrimSpace(string(r[:120]))
	}
	return s
}

// Task fills the task fields of an event when the ID is usable; a task event
// without a usable ID is dropped by the caller (ok false). A created or updated
// task needs a title, so an empty one is replaced by a neutral name.
func (e Event) Task(typ, id, title, state string) (Event, bool) {
	if !ValidNativeID(id) {
		return e, false
	}
	e.Type, e.TaskID, e.State = typ, id, state
	e.Title = Title(title)
	if e.Title == "" && typ != TaskCompleted {
		e.Title = "Task " + id
	}
	return e, true
}

// Subagent fills the agent fields of a subagent event.
func (e Event) Subagent(typ, agentID string) (Event, bool) {
	if !ValidNativeID(agentID) {
		return e, false
	}
	e.Type, e.AgentID = typ, agentID
	return e, true
}

// Generic translates the event names every host shares: a session's start and
// end, a prompt, anything else as plain activity. replaces says whether a start
// takes the place of the session the process had.
func Generic(base Event, norm string, replaces bool) Event {
	switch norm {
	case "sessionstart", "sessioncreated":
		base.Type, base.Map, base.Replaces = SessionStarted, MapStart, replaces
	case "sessionend", "sessiondeleted", "sessionshutdown":
		base.Type, base.Map = SessionEnded, MapEnd
	case "userpromptsubmit", "beforesubmitprompt":
		base.Type, base.Map, base.Replaces = Activity, MapTouch, replaces
	default:
		base.Type = Activity
	}
	return base
}
