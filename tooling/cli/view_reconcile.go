package main

// view_reconcile.go backfills what the hooks missed from the files a host keeps
// for a session: Claude Code and Codex JSONL, Grok updates.jsonl and Pi JSONL.
// It sends host events (the same ones a hook sends), so everything the viewer does
// with them stays the viewer's rule: activity as metadata, native tasks only in a
// work without registered tasks, never a stage or a closure. It never infers that
// a process died, and it sends no session_ended.
//
// Locations are observed on the maintainer's machine, not documented as a stable
// interface; a layout that changes makes the file not found, which is a result,
// not an error.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

const (
	// reconcileReadLimit bounds the part of a session file read.
	reconcileReadLimit = 16 << 20
	// reconcileHookReadLimit is the bound when a session-start hook reconciles.
	reconcileHookReadLimit = 4 << 20
	reconcileTailBytes     = 256 << 10
	reconcileMaxEvents     = 120
)

var reconcileSessionID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type viewWorkReconcileResult struct {
	Result         string `json:"result"`
	Host           string `json:"host"`
	SessionFile    string `json:"session_file,omitempty"`
	LastActivityAt string `json:"last_activity_at,omitempty"`
	ActivitySent   bool   `json:"activity_sent,omitempty"`
	Tasks          int    `json:"tasks,omitempty"`
	Sent           int    `json:"sent,omitempty"`
	Queued         int    `json:"queued,omitempty"`
	Note           string `json:"note,omitempty"`
	// Partial is set when the session file was larger than the read limit and
	// only its last part was read.
	Partial bool `json:"partial,omitempty"`
}

// executeHostReconcile runs `hive view work reconcile` without --pending: the
// identity was resolved by the parser, as for any semantic command.
func executeHostReconcile(c viewWorkCommand, w io.Writer) error {
	_, state, err := management.NormalizeOptions(c.Options)
	if err != nil {
		return err
	}
	if state, err = filepath.Abs(state); err != nil {
		return err
	}
	project := c.Project
	if project == "" {
		if project, err = defaultViewWorkProject(); err != nil {
			return workInput("project: run inside the repository or pass --project")
		}
	}
	source := viewer.TimelineSource{ProjectRoot: project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	sink := openHostSink(c.Options, state, source)
	r, err := reconcileHostSession(c.Options, sink, source, c.SessionFile, reconcileReadLimit)
	if err != nil {
		return err
	}
	return writeViewWorkBoundedJSON(w, r)
}

// reconcileHostSession reads the session's file and delivers what it finds.
func reconcileHostSession(opts management.Options, sink *hostSink, source viewer.TimelineSource, explicit string, limit int64) (viewWorkReconcileResult, error) {
	r := viewWorkReconcileResult{Result: "reconciled", Host: string(source.Host)}
	path := explicit
	if path == "" {
		var err error
		if path, err = locateHostSessionFile(opts, string(source.Host), source.NativeSessionID); err != nil {
			return r, err
		}
	}
	if path == "" {
		r.Result = "no_session_file"
		r.Note = "this host keeps no session file Hive reads, or none was found for the session"
		return r, nil
	}
	r.SessionFile = filepath.Base(path)
	facts, err := readHostSessionFile(string(source.Host), path, limit)
	if err != nil {
		return r, workStateError("session_file_unreadable")
	}
	if facts.Partial {
		r.Partial = true
		r.Note = fmt.Sprintf("the session file is larger than %d MiB: only its last %d MiB were read", limit>>20, limit>>20)
	}
	// Native tasks from before the current work was created belong to earlier
	// work in the same session; the viewer ignores them too.
	var created time.Time
	if sink.up {
		if work, ok := resolveHostEventWork(sink.state, sink.info, source); ok {
			created, _ = time.Parse(time.RFC3339Nano, work.CreatedAt)
		}
	}
	var events []viewer.TimelineHostEvent
	build := func(typ string, at time.Time, task reconcileTask) viewer.TimelineHostEvent {
		if task.ID != "" && task.Title == "" && typ != hostevent.TaskCompleted {
			task.Title = "Task " + task.ID
		}
		return viewer.TimelineHostEvent{
			SchemaVersion: viewer.TimelineSchemaVersion, Source: source, EventType: typ,
			OccurredAt: at.UTC().Format(time.RFC3339Nano),
			Data:       viewer.TimelineHostEventData{TaskID: task.ID, Title: task.Title, State: task.State},
		}
	}
	for _, t := range facts.Tasks {
		if t.At.Before(created) {
			continue
		}
		typ := hostevent.TaskUpdated
		switch t.State {
		case hostevent.StateCompleted:
			typ = hostevent.TaskCompleted
		case hostevent.StateNotStarted:
			typ = hostevent.TaskCreated
		}
		events = append(events, build(typ, t.At, t))
	}
	if len(events) > reconcileMaxEvents {
		events = events[len(events)-reconcileMaxEvents:]
	}
	r.Tasks = len(events)
	if !facts.LastActivity.IsZero() {
		r.LastActivityAt = facts.LastActivity.UTC().Format(time.RFC3339)
		// The viewer stamps activity with the event's own time, clamped to its clock.
		events = append(events, build(hostevent.Activity, facts.LastActivity, reconcileTask{}))
		r.ActivitySent = true
	}
	for _, e := range events {
		if e.Validate() != nil {
			continue
		}
		if sink.deliver(e) {
			r.Sent++
		} else {
			r.Queued++
		}
	}
	return r, nil
}

// --- locating the file ---

func locateHostSessionFile(opts management.Options, host, session string) (string, error) {
	if !reconcileSessionID.MatchString(session) {
		return "", workInput("native session ID is not a safe file name part")
	}
	cfg, _, err := management.NormalizeOptions(opts)
	if err != nil {
		return "", err
	}
	var patterns []string
	switch host {
	case "claude-code":
		patterns = []string{filepath.Join(cfg.ClaudeHome, "projects", "*", session+".jsonl")}
	case "codex":
		patterns = []string{filepath.Join(cfg.CodexHome, "sessions", "*", "*", "*", "rollout-*-"+session+".jsonl")}
	case "grok":
		patterns = []string{filepath.Join(cfg.GrokHome, "sessions", "*", session, "updates.jsonl")}
	case "pi":
		patterns = []string{filepath.Join(cfg.PiHome, "sessions", "*", "*_"+session+".jsonl")}
	default:
		return "", nil
	}
	var found []string
	for _, p := range patterns {
		matches, _ := filepath.Glob(p)
		for _, m := range matches {
			if st, err := os.Lstat(m); err == nil && st.Mode().IsRegular() && ownedByCurrentUser(st) {
				found = append(found, m)
			}
		}
	}
	if len(found) == 0 {
		return "", nil
	}
	sort.Strings(found)
	return found[len(found)-1], nil
}

// --- reading the file ---

type reconcileTask struct {
	ID, Title, State string
	At               time.Time
}

type reconcileFacts struct {
	LastActivity time.Time
	Tasks        []reconcileTask
	// Partial: the file was larger than the limit and only its tail was read.
	Partial bool
}

// openSessionFile opens without following links and checks the descriptor.
func openSessionFile(path string) (*os.File, os.FileInfo, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, nil, err
	}
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || !ownedByCurrentUser(st) {
		f.Close()
		return nil, nil, fmt.Errorf("unsafe session file")
	}
	return f, st, nil
}

func readHostSessionFile(host, path string, limit int64) (reconcileFacts, error) {
	f, st, err := openSessionFile(path)
	if err != nil {
		return reconcileFacts{}, err
	}
	defer f.Close()
	var facts reconcileFacts
	// Last activity: the newest timestamp in the tail of the file.
	tail := int64(reconcileTailBytes)
	from := st.Size() - tail
	if from < 0 {
		from = 0
	}
	buf := make([]byte, st.Size()-from)
	if _, err = f.ReadAt(buf, from); err != nil && err != io.EOF {
		return facts, err
	}
	for _, line := range bytes.Split(buf, []byte("\n")) {
		if m := hostevent.Decode(line); m != nil {
			if t := hostevent.When(m, "timestamp"); t.After(facts.LastActivity) {
				facts.LastActivity = t
			}
		}
	}
	if host == "claude-code" {
		start := int64(0)
		if st.Size() > limit {
			// A larger transcript is read from its last limit bytes, from the
			// first whole line on.
			start = st.Size() - limit
			facts.Partial = true
		}
		var in io.Reader = io.NewSectionReader(f, start, st.Size()-start)
		if start > 0 {
			br := bufio.NewReaderSize(in, 1<<20)
			if _, rerr := br.ReadBytes('\n'); rerr != nil {
				return facts, nil
			}
			in = br
		}
		facts.Tasks, err = readClaudeTasks(in)
	}
	return facts, err
}

var claudeTaskCreated = regexp.MustCompile(`Task #([0-9]+) created`)

// readClaudeTasks follows the native task tools in a Claude Code transcript:
// a TaskCreate tool_use gives the subject and its tool_result the number
// ("Task #N created successfully: ..."), and a TaskUpdate gives taskId and
// status. Observed in transcripts of Claude Code 2.1.x; not documented.
func readClaudeTasks(r io.Reader) ([]reconcileTask, error) {
	type block struct {
		Type      string          `json:"type"`
		ID        string          `json:"id"`
		Name      string          `json:"name"`
		Input     json.RawMessage `json:"input"`
		ToolUseID string          `json:"tool_use_id"`
		Content   json.RawMessage `json:"content"`
	}
	var line struct {
		Timestamp string `json:"timestamp"`
		Message   struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	pending := map[string]string{} // tool_use id -> subject
	titles := map[string]string{}
	var out []reconcileTask
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1<<20), 4<<20)
	for sc.Scan() {
		raw := sc.Bytes()
		if !bytes.Contains(raw, []byte("Task")) {
			continue
		}
		line.Message.Content, line.Timestamp = nil, ""
		if json.Unmarshal(raw, &line) != nil {
			continue
		}
		at, err := time.Parse(time.RFC3339Nano, line.Timestamp)
		if err != nil {
			continue
		}
		var blocks []block
		if json.Unmarshal(line.Message.Content, &blocks) != nil {
			continue
		}
		for _, b := range blocks {
			switch {
			case b.Type == "tool_use" && b.Name == "TaskCreate":
				var in struct {
					Subject string `json:"subject"`
				}
				if json.Unmarshal(b.Input, &in) == nil {
					pending[b.ID] = in.Subject
				}
			case b.Type == "tool_use" && b.Name == "TaskUpdate":
				var in struct {
					TaskID any    `json:"taskId"`
					Status string `json:"status"`
				}
				if json.Unmarshal(b.Input, &in) != nil {
					continue
				}
				id := strings.TrimSpace(fmt.Sprint(in.TaskID))
				state := map[string]string{"in_progress": hostevent.StateInProgress, "completed": hostevent.StateCompleted}[in.Status]
				if state != "" && hostevent.ValidNativeID(id) {
					out = append(out, reconcileTask{ID: id, Title: titles[id], State: state, At: at})
				}
			case b.Type == "tool_result":
				subject, ok := pending[b.ToolUseID]
				if !ok {
					continue
				}
				delete(pending, b.ToolUseID)
				if m := claudeTaskCreated.FindStringSubmatch(toolResultText(b.Content)); m != nil && hostevent.ValidNativeID(m[1]) {
					titles[m[1]] = hostevent.Title(subject)
					out = append(out, reconcileTask{ID: m[1], Title: titles[m[1]], State: hostevent.StateNotStarted, At: at})
				}
			}
		}
	}
	if err := sc.Err(); err != nil && err != bufio.ErrTooLong {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

func toolResultText(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	var parts []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &parts) == nil {
		var b strings.Builder
		for _, p := range parts {
			b.WriteString(p.Text)
		}
		return b.String()
	}
	return ""
}
