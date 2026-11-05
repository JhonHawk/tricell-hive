package main

// hook_send.go reports a host event to the viewer, or leaves it in the host
// event queue when the viewer is down. The queue holds the event as the host
// reported it, without a work or a stage: both are resolved when the event is
// sent, from the manager's current-work record, so an event queued before its
// work existed still binds to it once the semantic queue has created it.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
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

// sendHookEvents reports the events of one hook run. Every valid event is queued
// first, so a run the host kills at its budget loses nothing; then one sink (one
// health check) flushes the queues in order, with the events of this run last,
// and whatever is not sent stays queued. Order matters when the viewer is up: the
// work's queued semantic declarations first (they may create the work and its
// stage), then the queued host events.
func sendHookEvents(opts management.Options, state string, events []viewer.TimelineHostEvent) {
	var sources []viewer.TimelineSource
	seen := map[string]bool{}
	started := -1
	for n, event := range events {
		if event.Validate() != nil {
			continue
		}
		queueHostEvent(state, event)
		if key := hostSourceKey(event.Source); !seen[key] {
			seen[key] = true
			sources = append(sources, event.Source)
		}
		if event.EventType == hostevent.SessionStarted {
			started = n
		}
	}
	if len(sources) == 0 {
		return
	}
	sink := openHostSink(opts, state, sources...)
	if started >= 0 && sink.up {
		// A session that starts again (resume, a restart after a crash) may have
		// done work the hooks missed while the viewer or the hook was away.
		_, _ = reconcileHostSession(opts, sink, events[started].Source, "", reconcileHookReadLimit)
	}
}

func hostSourceKey(s viewer.TimelineSource) string {
	return string(s.Host) + "\x00" + s.NativeSessionID
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
	state   string
	info    serverInfo
	up      bool
	targets *hostTargetCache
}

// openHostSink checks the viewer once. With it up, it flushes the semantic queue
// of each source's current works and then the host event queue.
func openHostSink(opts management.Options, state string, sources ...viewer.TimelineSource) *hostSink {
	s := &hostSink{state: state, targets: &hostTargetCache{}}
	info, health, err := viewTimelineRunningHealth(newViewPaths(state))
	if err != nil || !containsString(health.Capabilities, "host_events") {
		return s
	}
	s.info, s.up = info, true
	for _, source := range sources {
		flushSemanticQueue(opts, state, source)
	}
	flushHostEventQueueWith(state, info, s.targets)
	return s
}

// deliver reports whether the event was sent; otherwise it waits in the queue.
func (s *hostSink) deliver(event viewer.TimelineHostEvent) bool {
	if s.up && postResolvedHostEventWith(s.state, s.info, event, s.targets) == hostSendDone {
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

// hostTargetCache remembers, for the length of one hook run or one flush, the
// current work (and so its active stage) each session resolved to, so a batch of
// events reads the work once per session instead of once per event. The nil
// cache resolves every time.
type hostTargetCache struct {
	byID map[string]hostTarget
}

type hostTarget struct {
	session viewer.TimelineSession
	ok      bool
}

// work reads the session's current work once per cache.
func (c *hostTargetCache) work(state string, info serverInfo, source viewer.TimelineSource) (viewer.TimelineSession, bool) {
	if c == nil {
		return resolveHostEventWork(state, info, source)
	}
	key := hostSourceKey(source)
	if t, ok := c.byID[key]; ok {
		return t.session, t.ok
	}
	s, ok := resolveHostEventWork(state, info, source)
	if c.byID == nil {
		c.byID = map[string]hostTarget{}
	}
	c.byID[key] = hostTarget{s, ok}
	return s, ok
}

func (c *hostTargetCache) resolve(state string, info serverInfo, source viewer.TimelineSource) (work, stage string) {
	s, ok := c.work(state, info, source)
	return hostEventTargetOf(s, ok)
}

// postResolvedHostEventWith binds the event to the session's current work and its
// active stage, whatever its type, when the session has one, and posts it. A
// rejected binding is retried once without it, so the activity still counts (an
// event that names no work moves the last activity of every work of the
// session). Transport failures and server errors are retried later from the
// queue; any other answer is final.
func postResolvedHostEventWith(state string, info serverInfo, event viewer.TimelineHostEvent, cache *hostTargetCache) hostSendResult {
	event.WorkID, event.StageEventID = cache.resolve(state, info, event.Source)
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
// Current works the viewer reports closed, failed, deleted or not found are
// released on the way, so they stop making the session's current work ambiguous;
// a local handle of a queued open is kept, since the viewer cannot know it yet.
func resolveHostEventWork(state string, info serverInfo, source viewer.TimelineSource) (viewer.TimelineSession, bool) {
	host := string(source.Host)
	ids, err := readCurrentWorks(state, host, source.NativeSessionID)
	if err != nil || len(ids) == 0 {
		return viewer.TimelineSession{}, false
	}
	var live viewer.TimelineSession
	candidates := 0 // works that are open or whose state could not be read
	for _, id := range ids {
		raw, err := viewTimelineHTTP(info, http.MethodGet, "/api/timeline/sessions/"+id, "", nil, viewer.TimelineMaxBody)
		if err != nil {
			var status *viewHTTPStatusError
			if errors.As(err, &status) && (status.Status == http.StatusGone || status.Status == http.StatusNotFound) {
				if releaseGoneHostWork(state, host, source.NativeSessionID, id, status.Status == http.StatusNotFound) {
					continue
				}
			}
			candidates++
			continue
		}
		r, err := viewer.DecodeTimelineSessionResponse(raw)
		if err != nil || r.Session.ViewerSessionID != id {
			candidates++
			continue
		}
		if r.Session.State == viewer.TimelineClosed || r.Session.State == viewer.TimelineFailed {
			_ = releaseCurrentWork(state, host, source.NativeSessionID, id)
			continue
		}
		live = r.Session
		candidates++
	}
	if candidates != 1 || live.ViewerSessionID == "" {
		return viewer.TimelineSession{}, false
	}
	return live, true
}

// releaseGoneHostWork removes a current work the viewer reported gone, under the
// session lock and so atomically with the check that a 404 handle has no queued
// open left. It reports whether the work is no longer current.
func releaseGoneHostWork(state, host, native, id string, keepQueuedHandle bool) bool {
	released := false
	err := withSessionLockFile(state, host, native, func() (bool, error) {
		if keepQueuedHandle && hasQueuedOpenHandle(state, id) {
			return false, nil
		}
		remove, err := releaseCurrentWorkLocked(state, host, native, id)
		released = err == nil
		return remove, err
	})
	return err == nil && released
}

// hasQueuedOpenHandle says whether the journal holds a queued open under this
// local handle. Anything unreadable counts as yes: the handle is kept.
func hasQueuedOpenHandle(state, handle string) bool {
	j, err := readWorkJournal(filepath.Join(state, "publisher", "v1"))
	if err != nil {
		return true
	}
	for _, q := range j.Queue {
		if q.Action == "open" && q.WorkID == handle && q.State != "rejected" {
			return true
		}
	}
	return false
}

func hostEventTargetOf(s viewer.TimelineSession, ok bool) (work, stage string) {
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

// hostEventSeq numbers the events one run queues.
var hostEventSeq atomic.Uint64

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
	// The per-run sequence keeps events of one timestamp in emission order.
	name := fmt.Sprintf("%020d-%06d%s", hookNow().UnixNano(), hostEventSeq.Add(1), suffix)
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

// flushHostEventQueueWith sends the queued events oldest first, at most
// hostEventFlushLimit per run, and stops at the first transient failure. An event
// the viewer refuses, or a file that is not an event, is dropped.
func flushHostEventQueueWith(state string, info serverInfo, cache *hostTargetCache) {
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
		if postResolvedHostEventWith(state, info, event, cache) == hostSendRetry {
			return
		}
		_ = os.Remove(path)
	}
}
