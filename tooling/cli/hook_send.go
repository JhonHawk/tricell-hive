package main

// hook_send.go reports a host event to the viewer, or leaves it in the host
// event queue when the viewer is down. The queue holds the event as the host
// reported it, without a work or a stage: both are resolved when the event is
// sent, from the manager's current-work record, so an event queued before its
// work existed still binds to it once the semantic queue has created it.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

const (
	hostEventQueueLimit = 256
	hostEventFlushLimit = 64
	hostEventBodyLimit  = 8 << 10
)

// sendHookEvent sends one translated event. Order matters when the viewer is up:
// the work's queued semantic declarations first (they may create the work and
// its stage), then the queued host events, then this one.
func sendHookEvent(opts management.Options, state string, ev hostevent.Event, hostStart time.Time) {
	event := buildHostEvent(ev, hostStart)
	if event.Validate() != nil {
		return
	}
	sink := openHostSink(opts, state, event.Source)
	sink.deliver(event)
	if ev.Type == hostevent.SessionStarted && sink.up {
		// A session that starts again (resume, a restart after a crash) may have
		// done work the hooks missed while the viewer or the hook was away.
		_, _ = reconcileHostSession(opts, sink, event.Source, "", reconcileHookReadLimit)
	}
}

// buildHostEvent is the viewer's event for a translated one. An activity-only
// event carries the host's turn ID when there is one, so each is distinct.
func buildHostEvent(ev hostevent.Event, hostStart time.Time) viewer.TimelineHostEvent {
	cwd := ev.Cwd
	if !filepath.IsAbs(cwd) {
		if wd, err := os.Getwd(); err == nil {
			cwd = wd
		}
	}
	data := viewer.TimelineHostEventData{TaskID: ev.TaskID, Title: ev.Title, State: ev.State, AgentID: ev.AgentID}
	if !strings.HasPrefix(ev.Type, "task_") {
		data.TurnID = ev.Turn
	}
	return viewer.TimelineHostEvent{
		SchemaVersion: viewer.TimelineSchemaVersion,
		Source:        viewer.TimelineSource{ProjectRoot: cwd, Host: viewer.TimelineHost(ev.Host), NativeSessionID: ev.SessionID},
		EventType:     ev.Type,
		OccurredAt:    hookOccurredAt(ev, hostStart).Format(time.RFC3339Nano),
		Data:          data,
	}
}

// hostSink delivers host events to the viewer, or to the host event queue when
// it is not running. Opening it with the viewer up sends the work's queued
// semantic declarations first (they may create the work and its stage), then the
// queued host events.
type hostSink struct {
	state string
	info  serverInfo
	up    bool
}

func openHostSink(opts management.Options, state string, source viewer.TimelineSource) *hostSink {
	s := &hostSink{state: state}
	info, health, err := viewTimelineRunningHealth(newViewPaths(state))
	if err != nil || !containsString(health.Capabilities, "host_events") {
		return s
	}
	s.info, s.up = info, true
	flushSemanticQueue(opts, state, source)
	flushHostEventQueue(state, info)
	return s
}

// deliver reports whether the event was sent; otherwise it waits in the queue.
func (s *hostSink) deliver(event viewer.TimelineHostEvent) bool {
	if s.up && postResolvedHostEvent(s.state, s.info, event) == hostSendDone {
		return true
	}
	queueHostEvent(s.state, event)
	return false
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// flushSemanticQueue sends the queued declarations of the session's current
// works. A work another command holds (busy) or any failure is left for the
// next run: the queue is durable.
func flushSemanticQueue(opts management.Options, state string, source viewer.TimelineSource) {
	ids, err := readCurrentWorks(state, string(source.Host), source.NativeSessionID)
	if err != nil {
		return
	}
	for _, id := range ids {
		c := viewWorkCommand{Action: "resume", Options: opts, WorkID: id, Host: string(source.Host), NativeSessionID: source.NativeSessionID}
		e, err := openViewWorkEngineQueueing(c, true)
		if err != nil {
			continue
		}
		func() {
			defer e.Close()
			if e.up && e.hasQueueFor(e.scope) {
				_ = e.flushQueue(c)
			}
		}()
	}
}

func (e *viewWorkEngine) hasQueueFor(scope string) bool {
	for _, q := range e.journal.Queue {
		if e.scopeOf(q.WorkID) == scope {
			return true
		}
	}
	return false
}

type hostSendResult int

const (
	hostSendDone hostSendResult = iota
	hostSendRetry
)

// postResolvedHostEvent binds the event to the session's current work and its
// active stage, whatever its type, when the session has one, and posts it. A
// rejected binding is retried once without it, so the activity still counts (an
// event that names no work moves the last activity of every work of the
// session). Transport failures and server errors are retried later from the
// queue; any other answer is final.
func postResolvedHostEvent(state string, info serverInfo, event viewer.TimelineHostEvent) hostSendResult {
	event.WorkID, event.StageEventID = resolveHostEventTarget(state, info, event.Source)
	status, err := postHostEvent(info, event)
	if err != nil || status >= 500 {
		return hostSendRetry
	}
	if status >= 400 && event.WorkID != "" {
		event.WorkID, event.StageEventID = "", ""
		status, err = postHostEvent(info, event)
		if err != nil || status >= 500 {
			return hostSendRetry
		}
	}
	return hostSendDone
}

// resolveHostEventWork reads the session's one current work from the viewer.
func resolveHostEventWork(state string, info serverInfo, source viewer.TimelineSource) (viewer.TimelineSession, bool) {
	ids, err := readCurrentWorks(state, string(source.Host), source.NativeSessionID)
	if err != nil || len(ids) != 1 {
		return viewer.TimelineSession{}, false
	}
	raw, err := viewTimelineHTTP(info, http.MethodGet, "/api/timeline/sessions/"+ids[0], "", nil, viewer.TimelineMaxBody)
	if err != nil {
		return viewer.TimelineSession{}, false
	}
	r, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil || r.Session.ViewerSessionID != ids[0] {
		return viewer.TimelineSession{}, false
	}
	return r.Session, true
}

// resolveHostEventTarget finds the session's one current work and that work's
// active stage. Several current works, a work still queued locally, a finished
// work or an unreadable one name nothing.
func resolveHostEventTarget(state string, info serverInfo, source viewer.TimelineSource) (work, stage string) {
	s, ok := resolveHostEventWork(state, info, source)
	if !ok {
		return "", ""
	}
	if s.State == viewer.TimelineClosed || s.State == viewer.TimelineFailed || s.CurrentStageEventID == nil {
		return s.ViewerSessionID, ""
	}
	return s.ViewerSessionID, *s.CurrentStageEventID
}

// postHostEvent calls POST /api/host-events. The route takes the local token and
// no Origin header, because a local process calls it and not the browser.
func postHostEvent(info serverInfo, event viewer.TimelineHostEvent) (int, error) {
	body, err := json.Marshal(event)
	if err != nil || len(body) > viewer.TimelineMaxBody {
		return 0, fmt.Errorf("host event body")
	}
	req, err := http.NewRequest(http.MethodPost, info.url()+"/api/host-events", bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-Hive-Token", info.Token)
	req.Header.Set(viewer.TimelineRevisionHeader, viewer.TimelineRevisionValue)
	req.Header.Set("Content-Type", "application/json")
	client := *viewClient // 2 s timeout
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, hostEventBodyLimit))
	return resp.StatusCode, nil
}

// --- host event queue ---

func hostEventOperation(e viewer.TimelineHostEvent) string { return e.OperationID() }

// queueHostEvent keeps the event until a run finds the viewer up. The same event
// is the same file; past the limit the oldest event is dropped.
func queueHostEvent(state string, event viewer.TimelineHostEvent) {
	dir, err := ensureHookDir(state, "queue")
	if err != nil {
		return
	}
	suffix := "-" + strings.TrimPrefix(hostEventOperation(event), "host-") + ".json"
	names := hostEventQueueNames(dir)
	for _, n := range names {
		if strings.HasSuffix(n, suffix) {
			return
		}
	}
	raw, err := json.Marshal(event)
	if err != nil {
		return
	}
	// The name orders the queue; it is not part of the event.
	name := fmt.Sprintf("%020d%s", hookNow().UnixNano(), suffix)
	if writePrivateWorkFile(dir, filepath.Join(dir, name), raw) != nil {
		return
	}
	for len(names) >= hostEventQueueLimit {
		_ = os.Remove(filepath.Join(dir, names[0]))
		names = names[1:]
	}
}

func hostEventQueueNames(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if n := e.Name(); strings.HasSuffix(n, ".json") && !strings.HasPrefix(n, ".") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	return names
}

// flushHostEventQueue sends the queued events oldest first, at most
// hostEventFlushLimit per run, and stops at the first transient failure. An event
// the viewer refuses, or a file that is not an event, is dropped.
func flushHostEventQueue(state string, info serverInfo) {
	dir := filepath.Join(state, "hooks", "queue")
	if checkSafeDir(filepath.Join(state, "hooks")) != nil || checkSafeDir(dir) != nil {
		return
	}
	l, err := acquirePrivateWorkLock(filepath.Join(state, "hooks", "queue.lock"), false)
	if err != nil {
		return // another run is sending the queue
	}
	defer l.Release()
	for n, name := range hostEventQueueNames(dir) {
		if n >= hostEventFlushLimit {
			return
		}
		path := filepath.Join(dir, name)
		raw, err := readSafeMapFile(path)
		var event viewer.TimelineHostEvent
		if err != nil || json.Unmarshal(raw, &event) != nil || event.Validate() != nil {
			_ = os.Remove(path)
			continue
		}
		event.WorkID, event.StageEventID = "", ""
		if postResolvedHostEvent(state, info, event) == hostSendRetry {
			return
		}
		_ = os.Remove(path)
	}
}
