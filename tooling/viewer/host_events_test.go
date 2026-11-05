package viewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These tests exercise POST /api/host-events at the HTTP boundary: the endpoint
// authenticated by the local token without a browser origin, the maintenance-only
// last activity, the mirroring of native tasks and the effects it must never have.

type hostEventHarness struct {
	t      *testing.T
	store  *Store
	server *Server
	work   TimelineSession
	source TimelineSource
	clock  time.Time
	// request is the request that created work, for creating more of the session.
	request TimelineCreateSessionRequest
}

func newHostEventHarness(t *testing.T) *hostEventHarness {
	t.Helper()
	s, request := newTimelineFixture(t)
	h := &hostEventHarness{t: t, store: s, clock: time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)}
	s.now = func() time.Time { return h.clock }
	h.work = createTimeline(t, s, request)
	h.source = request.Source
	h.request = request
	server, err := NewServer(ServerOptions{Store: s, Port: 4483, Token: strings.Repeat("h", 32)})
	if err != nil {
		t.Fatal(err)
	}
	h.server = server
	return h
}

func (h *hostEventHarness) advance() { h.clock = h.clock.Add(time.Minute) }

func (h *hostEventHarness) post(body []byte, token bool, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:4483/api/host-events", bytes.NewReader(body))
	if token {
		r.Header.Set(tokenHeader, strings.Repeat("h", 32))
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	h.server.ServeHTTP(w, r)
	return w
}

func (h *hostEventHarness) event(eventType string, data TimelineHostEventData, workID, stageID string) []byte {
	return h.eventFrom(h.source, eventType, data, workID, stageID)
}

func (h *hostEventHarness) eventFrom(source TimelineSource, eventType string, data TimelineHostEventData, workID, stageID string) []byte {
	e := TimelineHostEvent{SchemaVersion: 3, Source: source, EventType: eventType, OccurredAt: h.clock.Format(time.RFC3339), WorkID: workID, StageEventID: stageID, Data: data}
	raw, err := json.Marshal(e)
	if err != nil {
		h.t.Fatal(err)
	}
	return raw
}

func (h *hostEventHarness) send(raw []byte) (int, TimelineHostEventResult) {
	h.t.Helper()
	w := h.post(raw, true, "")
	var result TimelineHostEventResult
	if w.Code == 200 {
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			h.t.Fatalf("result %s: %v", w.Body.String(), err)
		}
	}
	return w.Code, result
}

func (h *hostEventHarness) mustSend(raw []byte) TimelineHostEventResult {
	h.t.Helper()
	code, result := h.send(raw)
	if code != 200 {
		h.t.Fatalf("host event status %d", code)
	}
	return result
}

func (h *hostEventHarness) current() TimelineSession {
	h.t.Helper()
	return domainCurrent(h.t, h.store, h.work.ViewerSessionID)
}

func (h *hostEventHarness) stage(stage TimelineStage) string {
	h.t.Helper()
	return domainStage(h.t, h.store, h.work.ViewerSessionID, stage)
}

func hostTask(id, title string) TimelineHostEventData {
	return TimelineHostEventData{TaskID: id, Title: title}
}

func hostTaskByKey(v TimelineSession, key string) (TimelineTask, bool) {
	for _, task := range v.Tasks {
		if task.TaskKey == key {
			return task, true
		}
	}
	return TimelineTask{}, false
}

func TestHostEventsCapabilityIsAdvertisedInHealth(t *testing.T) {
	h := newHostEventHarness(t)
	r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:4483/api/health", nil)
	r.Header.Set(tokenHeader, strings.Repeat("h", 32))
	w := httptest.NewRecorder()
	h.server.ServeHTTP(w, r)
	var health struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil || !contains(health.Capabilities, "host_events") || !contains(health.Capabilities, "local_actions_cli") {
		t.Fatalf("capabilities %v (%v)", health.Capabilities, err)
	}
}

func TestHostEventsRequireTheTokenButNoBrowserOrigin(t *testing.T) {
	h := newHostEventHarness(t)
	raw := h.event("activity", TimelineHostEventData{}, "", "")
	if w := h.post(raw, false, ""); w.Code != 403 {
		t.Fatalf("without the token: %d", w.Code)
	}
	if w := h.post(raw, false, "http://127.0.0.1:4483"); w.Code != 403 {
		t.Fatalf("own origin without the token: %d", w.Code)
	}
	if w := h.post(raw, true, "http://evil.example"); w.Code != 403 {
		t.Fatalf("foreign origin with the token: %d", w.Code)
	}
	// The route is for a local process: a request carrying any browser origin is refused,
	// the viewer's own origin included.
	if w := h.post(raw, true, "http://127.0.0.1:4483"); w.Code != 403 {
		t.Fatalf("own browser origin with the token: %d", w.Code)
	}
	if w := h.post(raw, true, ""); w.Code != 200 {
		t.Fatalf("token without an origin: %d %s", w.Code, w.Body.String())
	}
	r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:4483/api/host-events", nil)
	r.Header.Set(tokenHeader, strings.Repeat("h", 32))
	w := httptest.NewRecorder()
	h.server.ServeHTTP(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET: %d", w.Code)
	}
}

func TestHostEventFromANonProducerSessionIsRejectedAndChangesNothing(t *testing.T) {
	h := newHostEventHarness(t)
	before := h.current()
	stranger := h.source
	stranger.NativeSessionID = "not-a-producer"
	h.advance()
	raw := h.eventFrom(stranger, "task_created", hostTask("1", "Intruder"), before.ViewerSessionID, "")
	if code, _ := h.send(raw); code != 403 {
		t.Fatalf("non-producer session: %d", code)
	}
	after := h.current()
	if after.LastActivityAt != before.LastActivityAt || after.Version != before.Version || len(after.Tasks) != 0 {
		t.Fatalf("rejected event changed the work: %+v", after)
	}
}

func TestHostEventBetweenStageDeclarationAndNextMutationKeepsExpectedVersion(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	declared := h.current()
	h.advance()
	h.mustSend(h.event("activity", TimelineHostEventData{}, id, stageID))
	h.mustSend(h.event("task_created", hostTask("1", "Write the tests"), id, stageID))
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "1"}, id, stageID))
	h.mustSend(h.event("session_ended", TimelineHostEventData{}, id, stageID))
	after := h.current()
	if after.Version != declared.Version || len(after.Events) != len(declared.Events) {
		t.Fatalf("host events moved version %d -> %d or events %d -> %d", declared.Version, after.Version, len(declared.Events), len(after.Events))
	}
	// The agent's next conditional mutation still commits with the version of its last receipt.
	_, err := h.store.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "agent-next", Actor: TimelineActor{Kind: "native_producer", ProducerID: declared.ForegroundProducerID}, Reason: "Next stage", Kind: "stage", Stage: "verify", StageState: "in_progress", ExpectedSessionVersion: &declared.Version, Title: "verify", Summary: "Work stage", TicketIDs: []string{}})
	if err != nil {
		t.Fatalf("agent mutation with the declared version failed: %v", err)
	}
	// The agent's own task declaration also still commits.
	v := h.current()
	create := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, h.store, id, "agent-task", v.ForegroundProducerID), TaskKey: "T1", PlanOrder: 1, Title: "Agent task", Summary: "", Segments: []TimelineTaskSegmentInput{{PlannedStage: "verify", Order: 1}}}
	if _, err = h.store.TimelineV3Operation(id, "task_create", "", &create); err != nil {
		t.Fatalf("agent task_create: %v", err)
	}
}

func TestHostActivityIsMaintenanceMetadataNotAnEvent(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	before := h.current()
	h.advance()
	h.mustSend(h.event("activity", TimelineHostEventData{}, id, stageID))
	after := h.current()
	if after.LastActivityAt != h.clock.Format(time.RFC3339Nano) || after.LastActivityAt == before.LastActivityAt {
		t.Fatalf("last activity %s -> %s", before.LastActivityAt, after.LastActivityAt)
	}
	if after.Version != before.Version || len(after.Events) != len(before.Events) || after.State != before.State || after.CurrentStageEventID == nil || *after.CurrentStageEventID != stageID {
		t.Fatalf("activity changed more than last activity: %+v", after)
	}
	// The maintenance pass leaves the work as it is: it never bumps the version over a host event.
	h.advance()
	h.store.mu.Lock()
	err := h.store.timelineMaintainLocked()
	h.store.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if again := h.current(); again.Version != before.Version {
		t.Fatalf("maintenance bumped the version to %d", again.Version)
	}
	// The index (what the viewer lists) carries the new last activity.
	for _, entry := range h.store.TimelineIndex() {
		if entry.ViewerSessionID == id && entry.LastActivityAt != after.LastActivityAt {
			t.Fatalf("index last activity %s", entry.LastActivityAt)
		}
	}
}

func TestHostTasksMirrorIntoAWorkWithoutRegisteredTasks(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	declared := h.current()
	h.advance()
	h.mustSend(h.event("task_created", hostTask("7", "Run typecheck"), id, stageID))
	v := h.current()
	task, ok := hostTaskByKey(v, "host:7")
	if !ok || task.Title != "Run typecheck" || task.State != "not_started" || len(task.Segments) != 1 {
		t.Fatalf("mirrored task: %+v", v.Tasks)
	}
	segment := task.Segments[0]
	if segment.PlannedStage != "build" || segment.DisplayStageEventID != stageID || segment.AssignedProducerID != declared.ForegroundProducerID {
		t.Fatalf("segment not bound to the active stage and the session producer: %+v", segment)
	}
	h.advance()
	h.mustSend(h.event("task_updated", TimelineHostEventData{TaskID: "7", Title: "Run typecheck", State: "in_progress"}, id, stageID))
	if task, _ = hostTaskByKey(h.current(), "host:7"); task.State != "in_progress" {
		t.Fatalf("update not applied: %+v", task)
	}
	h.advance()
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "7"}, id, stageID))
	v = h.current()
	task, _ = hostTaskByKey(v, "host:7")
	if task.State != "completed" || task.Segments[0].State != "completed" || task.TaskVersion != 3 {
		t.Fatalf("completion: %+v", task)
	}
	// Nothing about stages, authority or work state moved.
	if v.Version != declared.Version || v.State != declared.State || v.ForegroundProducerID != declared.ForegroundProducerID || len(v.Producers) != len(declared.Producers) || len(v.Events) != len(declared.Events) || v.CurrentStageEventID == nil || *v.CurrentStageEventID != stageID {
		t.Fatalf("host task events changed lifecycle: %+v", v)
	}
	for _, e := range v.Events {
		if e.EventID == stageID && e.StageState != "in_progress" {
			t.Fatalf("stage state changed to %s", e.StageState)
		}
	}
	if err := ValidateTimelineSession(v); err != nil {
		t.Fatal(err)
	}
}

func TestHostTasksAreNotAddedWhenTheAgentRegisteredTasks(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	create := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, h.store, id, "agent-task", h.work.ForegroundProducerID), TaskKey: "T1", PlanOrder: 1, Title: "Plan task", Summary: "", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	if _, err := h.store.TimelineV3Operation(id, "task_create", "", &create); err != nil {
		t.Fatal(err)
	}
	before := h.current()
	h.advance()
	h.mustSend(h.event("task_created", hostTask("9", "Native task"), id, stageID))
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "9"}, id, stageID))
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "T1"}, id, stageID))
	after := h.current()
	if len(after.Tasks) != 1 || after.Tasks[0].TaskKey != "T1" || after.Tasks[0].State != "not_started" || after.Tasks[0].TaskVersion != before.Tasks[0].TaskVersion {
		t.Fatalf("tasks changed: %+v", after.Tasks)
	}
	if after.Version != before.Version || after.LastActivityAt != h.clock.Format(time.RFC3339Nano) {
		t.Fatalf("only last activity should move: version %d, activity %s", after.Version, after.LastActivityAt)
	}
}

func TestHostTasksAreCappedAtFifty(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	for i := 1; i <= 52; i++ {
		h.advance()
		h.mustSend(h.event("task_created", hostTask(fmt.Sprint(i), fmt.Sprintf("Native task %d", i)), id, stageID))
	}
	v := h.current()
	if len(v.Tasks) != TimelineMaxHostTasks {
		t.Fatalf("%d tasks, want %d", len(v.Tasks), TimelineMaxHostTasks)
	}
	if _, ok := hostTaskByKey(v, "host:51"); ok {
		t.Fatal("task past the cap was mirrored")
	}
	// A task already mirrored still follows the host.
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "50"}, id, stageID))
	if task, _ := hostTaskByKey(h.current(), "host:50"); task.State != "completed" {
		t.Fatalf("mirrored task past 49 stopped following the host: %+v", task)
	}
	if err := ValidateTimelineSession(h.current()); err != nil {
		t.Fatal(err)
	}
}

func TestHostEventWithoutCurrentWorkOrActiveStageOnlyMovesLastActivity(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	before := h.current()
	// No work: the event reaches the works of its session and moves only their activity.
	h.advance()
	h.mustSend(h.event("task_created", hostTask("1", "No current work"), "", ""))
	after := h.current()
	if len(after.Tasks) != 0 || after.Version != before.Version || len(after.Events) != len(before.Events) {
		t.Fatalf("event without a work changed content: %+v", after)
	}
	if after.LastActivityAt != h.clock.Format(time.RFC3339Nano) {
		t.Fatalf("last activity %s", after.LastActivityAt)
	}
	// A work but no active stage.
	h.advance()
	h.mustSend(h.event("task_created", hostTask("2", "No stage"), id, ""))
	after = h.current()
	if len(after.Tasks) != 0 || after.Version != before.Version || after.LastActivityAt != h.clock.Format(time.RFC3339Nano) {
		t.Fatalf("event without a stage changed content: %+v", after)
	}
	// A session that belongs to no work is accepted and changes nothing.
	stranger := h.source
	stranger.NativeSessionID = "no-work-yet"
	h.advance()
	result := h.mustSend(h.eventFrom(stranger, "session_started", TimelineHostEventData{}, "", ""))
	if result.Status != "no_work" {
		t.Fatalf("status %q", result.Status)
	}
	if again := h.current(); again.LastActivityAt != after.LastActivityAt {
		t.Fatal("another session moved this work's activity")
	}
}

func TestHostEventReplayAppliesOnce(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	created := h.event("task_created", hostTask("3", "Replayed"), id, stageID)
	first := h.mustSend(created)
	if first.Status != "applied" || first.OperationID == "" {
		t.Fatalf("first result %+v", first)
	}
	h.advance()
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "3"}, id, stageID))
	settled := h.current()
	// Replaying the creation after the completion neither regresses the task nor moves activity.
	h.advance()
	replay := h.mustSend(created)
	if replay.Status != "replayed" || replay.OperationID != first.OperationID {
		t.Fatalf("replay result %+v", replay)
	}
	after := h.current()
	task, _ := hostTaskByKey(after, "host:3")
	if len(after.Tasks) != 1 || task.State != "completed" || task.TaskVersion != 2 || after.LastActivityAt != settled.LastActivityAt || after.Version != settled.Version {
		t.Fatalf("replay changed the work: %+v activity %s", after.Tasks, after.LastActivityAt)
	}
	// The replay survives a restart of the store.
	h.store.mu.Lock()
	h.store.timeline = map[string]*timelineRecord{}
	err := h.store.timelineLoad()
	h.store.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	h.advance()
	if again := h.mustSend(created); again.Status != "replayed" {
		t.Fatalf("replay after restart: %+v", again)
	}
	if final := h.current(); final.LastActivityAt != settled.LastActivityAt {
		t.Fatal("replay after restart moved activity")
	}
}

func TestHostEventCannotCarryALifecycleChange(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	before := h.current()
	for _, extra := range []string{`"stage_state":"completed"`, `"state":"closed"`, `"closure":{"outcome":"done"}`, `"producer_id":"producer-other"`, `"actor":{"kind":"local_viewer"}`} {
		body := fmt.Sprintf(`{"schema_version":3,"source":{"project_root":%q,"host":%q,"native_session_id":%q},"event_type":"activity","occurred_at":"2026-11-04T10:00:00Z","work_id":%q,"stage_event_id":%q,"data":{},%s}`, h.source.ProjectRoot, h.source.Host, h.source.NativeSessionID, id, stageID, extra)
		if code, _ := h.send([]byte(body)); code != 400 {
			t.Fatalf("%s accepted: %d", extra, code)
		}
	}
	// Data that would complete a stage is not part of the closed task shape either.
	body := fmt.Sprintf(`{"schema_version":3,"source":{"project_root":%q,"host":%q,"native_session_id":%q},"event_type":"activity","occurred_at":"2026-11-04T10:00:00Z","data":{"state":"completed"}}`, h.source.ProjectRoot, h.source.Host, h.source.NativeSessionID)
	if code, _ := h.send([]byte(body)); code != 400 {
		t.Fatalf("activity with task data accepted: %d", code)
	}
	after := h.current()
	if after.Version != before.Version || after.State != before.State || len(after.Events) != len(before.Events) || after.LastActivityAt != before.LastActivityAt {
		t.Fatalf("rejected events changed the work: %+v", after)
	}
}

func TestHostEventOnAClosedOrFailedWorkOnlyMovesLastActivity(t *testing.T) {
	for _, terminal := range []TimelineState{TimelineClosed, TimelineFailed} {
		t.Run(string(terminal), func(t *testing.T) {
			h := newHostEventHarness(t)
			id := h.work.ViewerSessionID
			stageID := h.stage("build")
			now := h.current()
			actor := TimelineActor{Kind: "native_producer", ProducerID: now.ForegroundProducerID}
			var stageEvent TimelineEvent
			for _, event := range now.Events {
				if event.EventID == stageID {
					stageEvent = event
				}
			}
			if _, err := h.store.TimelinePatchEvent(id, stageID, TimelineEventPatchRequest{Actor: actor, Reason: "Finish the stage", SchemaVersion: 3, OperationID: "end-stage", ExpectedVersion: stageEvent.Version, ExpectedSessionVersion: timelineInt(now.Version), StageState: timelineState("completed")}); err != nil {
				t.Fatalf("finishing the stage: %v", err)
			}
			if terminal == TimelineClosed {
				capacityClose(t, h.store, id, "end")
			} else {
				now = h.current()
				state := terminal
				patch := TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "end", Actor: TimelineActor{Kind: "native_producer", ProducerID: now.ForegroundProducerID}, Reason: "End the work", ExpectedVersion: now.Version, State: &state}
				if _, err := h.store.TimelinePatch(id, patch); err != nil {
					t.Fatalf("ending the work: %v", err)
				}
			}
			before := h.current()
			h.advance()
			result := h.mustSend(h.event("task_created", hostTask("1", "After the end"), id, stageID))
			if result.Status != "activity_only" {
				t.Fatalf("status %q", result.Status)
			}
			after := h.current()
			if len(after.Tasks) != 0 || after.Version != before.Version || len(after.Events) != len(before.Events) || after.State != terminal {
				t.Fatalf("terminal work changed beyond last activity: %+v", after)
			}
			if after.LastActivityAt != h.clock.Format(time.RFC3339Nano) {
				t.Fatalf("last activity %s", after.LastActivityAt)
			}
		})
	}
}

func TestHostEventFromAnotherProjectIsRejected(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	before := h.current()
	// Same host and native session ID, but a different registered project.
	other := t.TempDir()
	if err := os.MkdirAll(filepath.Join(other, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := registerRepository(h.store.vdir, other, h.clock); err != nil {
		t.Fatal(err)
	}
	foreign := h.source
	foreign.ProjectRoot = other
	h.advance()
	if code, _ := h.send(h.eventFrom(foreign, "task_created", hostTask("1", "Elsewhere"), id, stageID)); code != 409 {
		t.Fatalf("event of another project naming this work: %d", code)
	}
	// Without a work it must not reach this project's work either.
	if result := h.mustSend(h.eventFrom(foreign, "activity", TimelineHostEventData{}, "", "")); result.Status != "no_work" {
		t.Fatalf("status %q", result.Status)
	}
	// A project root nobody registered resolves to nothing.
	unregistered := h.source
	unregistered.ProjectRoot = t.TempDir()
	if code, _ := h.send(h.eventFrom(unregistered, "task_created", hostTask("1", "Unknown"), id, stageID)); code != 409 {
		t.Fatalf("unregistered project: %d", code)
	}
	after := h.current()
	if after.LastActivityAt != before.LastActivityAt || len(after.Tasks) != 0 || after.Version != before.Version {
		t.Fatalf("rejected events changed the work: %+v", after)
	}
	// A subdirectory of the registered project is the same project.
	sub := filepath.Join(h.source.ProjectRoot, "pkg")
	if err := os.MkdirAll(sub, 0o700); err != nil {
		t.Fatal(err)
	}
	inside := h.source
	inside.ProjectRoot = sub
	h.advance()
	if r := h.mustSend(h.eventFrom(inside, "activity", TimelineHostEventData{}, id, stageID)); r.Status != "applied" {
		t.Fatalf("subdirectory: %+v", r)
	}
}

func TestHostEventReceiptsAreKeptApartFromAgentReceipts(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	raw := h.event("task_created", hostTask("5", "Own receipt"), id, stageID)
	applied := h.mustSend(raw)
	// An agent operation may reuse the ID a host event derived: it is a different operation.
	v := h.current()
	base := domainBase(t, h.store, id, applied.OperationID, v.ForegroundProducerID)
	create := TimelineTaskCreateRequest{TimelineNativeOperation: base, TaskKey: "T1", PlanOrder: 1, Title: "Agent task", Summary: "", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	res, err := h.store.TimelineV3Operation(id, "task_create", "", &create)
	if err != nil || res.TaskID == "" || res.Kind != "task_create" {
		t.Fatalf("agent operation collided with a host receipt: %v %+v", err, res)
	}
	// The host event is still a replay of itself afterwards.
	h.advance()
	if again := h.mustSend(raw); again.Status != "replayed" {
		t.Fatalf("status %q", again.Status)
	}
	// And an agent operation that reuses an agent receipt is not mistaken for a host one.
	h.store.mu.Lock()
	for _, receipt := range h.store.timeline[id].Receipts {
		if receipt.OperationID == applied.OperationID && receipt.Result.Kind != "task_create" {
			t.Errorf("host operation stored among agent receipts: %+v", receipt.Result)
		}
	}
	h.store.mu.Unlock()
}

func TestHostActivityAtIsSetOnlyByHostEvents(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	if h.current().HostActivityAt != "" {
		t.Fatalf("a new work has host activity %q", h.current().HostActivityAt)
	}
	// Agent operations move last activity, never host activity.
	h.advance()
	stageID := h.stage("build")
	agent := h.current()
	if agent.HostActivityAt != "" || agent.LastActivityAt == h.work.LastActivityAt {
		t.Fatalf("agent operation: host %q last %s", agent.HostActivityAt, agent.LastActivityAt)
	}
	h.advance()
	h.mustSend(h.event("activity", TimelineHostEventData{}, id, stageID))
	hostAt := h.clock.Format(time.RFC3339Nano)
	after := h.current()
	if after.HostActivityAt != hostAt || after.LastActivityAt != hostAt {
		t.Fatalf("host event: host %q last %s, want %s", after.HostActivityAt, after.LastActivityAt, hostAt)
	}
	found := false
	for _, entry := range h.store.TimelineIndex() {
		if entry.ViewerSessionID == id {
			found = true
			if entry.HostActivityAt != hostAt {
				t.Fatalf("index host activity %q", entry.HostActivityAt)
			}
		}
	}
	if !found {
		t.Fatal("work missing from the index")
	}
	// A later agent operation moves last activity and leaves host activity as it was.
	h.advance()
	v := h.current()
	create := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, h.store, id, "task-after-host", v.ForegroundProducerID), TaskKey: "T1", PlanOrder: 1, Title: "Agent task", Summary: "", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	if _, err := h.store.TimelineV3Operation(id, "task_create", "", &create); err != nil {
		t.Fatal(err)
	}
	later := h.current()
	if later.HostActivityAt != hostAt || later.LastActivityAt != h.clock.Format(time.RFC3339Nano) {
		t.Fatalf("agent operation: host %q last %s", later.HostActivityAt, later.LastActivityAt)
	}
	// It survives a restart.
	h.store.mu.Lock()
	h.store.timeline = map[string]*timelineRecord{}
	err := h.store.timelineLoad()
	h.store.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	if reloaded := h.current(); reloaded.HostActivityAt != hostAt {
		t.Fatalf("after restart %q", reloaded.HostActivityAt)
	}
	if err := ValidateTimelineSession(h.current()); err != nil {
		t.Fatal(err)
	}
}

func TestHostEventNamingAStageThatIsNoLongerActiveOnlyMovesActivity(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	oldStage := h.stage("build")
	now := h.current()
	var stageEvent TimelineEvent
	for _, event := range now.Events {
		if event.EventID == oldStage {
			stageEvent = event
		}
	}
	actor := TimelineActor{Kind: "native_producer", ProducerID: now.ForegroundProducerID}
	if _, err := h.store.TimelinePatchEvent(id, oldStage, TimelineEventPatchRequest{Actor: actor, Reason: "Finish the stage", SchemaVersion: 3, OperationID: "end-build", ExpectedVersion: stageEvent.Version, ExpectedSessionVersion: timelineInt(now.Version), StageState: timelineState("completed")}); err != nil {
		t.Fatal(err)
	}
	activeStage := h.stage("verify")
	before := h.current()
	h.advance()
	result := h.mustSend(h.event("task_created", hostTask("9", "Late task"), id, oldStage))
	if result.Status != "activity_only" {
		t.Fatalf("status %q", result.Status)
	}
	after := h.current()
	if len(after.Tasks) != 0 || after.Version != before.Version || len(after.Events) != len(before.Events) || after.CurrentStageEventID == nil || *after.CurrentStageEventID != activeStage {
		t.Fatalf("stale stage changed content: %+v", after)
	}
	hostAt := h.clock.Format(time.RFC3339Nano)
	if after.LastActivityAt != hostAt || after.HostActivityAt != hostAt {
		t.Fatalf("activity not recorded: last %s host %s", after.LastActivityAt, after.HostActivityAt)
	}
	// A foreign work and a non-producer are still rejected whole.
	if code, _ := h.send(h.event("activity", TimelineHostEventData{}, strings.Repeat("2", 64), oldStage)); code == 200 {
		t.Fatal("foreign work accepted")
	}
	stranger := h.source
	stranger.NativeSessionID = "not-a-producer"
	if code, _ := h.send(h.eventFrom(stranger, "activity", TimelineHostEventData{}, id, oldStage)); code == 200 {
		t.Fatal("non-producer accepted")
	}
}

func TestHostTaskEventOlderThanTheLastAppliedOneNeverRegressesATask(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	h.mustSend(h.event("task_created", hostTask("4", "Settled"), id, stageID))
	h.advance()
	progress := h.event("task_updated", TimelineHostEventData{TaskID: "4", Title: "Settled", State: "in_progress"}, id, stageID)
	h.mustSend(progress)
	h.advance()
	h.mustSend(h.event("task_completed", TimelineHostEventData{TaskID: "4"}, id, stageID))
	// More operations than the replay memory keeps push the older event out of it.
	for i := 0; i < timelineMaxHostOperations+1; i++ {
		h.advance()
		h.mustSend(h.event("activity", TimelineHostEventData{}, id, stageID))
	}
	settled := h.current()
	h.advance()
	result := h.mustSend(progress)
	if result.Status != "activity_only" {
		t.Fatalf("late event status %q", result.Status)
	}
	after := h.current()
	task, _ := hostTaskByKey(after, "host:4")
	if task.State != "completed" || task.Segments[0].State != "completed" || task.TaskVersion != 3 || after.Version != settled.Version {
		t.Fatalf("late event regressed the task: %+v", task)
	}
	if after.LastActivityAt != settled.LastActivityAt {
		t.Fatalf("a late event moved activity: %s -> %s", settled.LastActivityAt, after.LastActivityAt)
	}
	// The record survives a restart.
	h.store.mu.Lock()
	h.store.timeline = map[string]*timelineRecord{}
	err := h.store.timelineLoad()
	h.store.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	h.advance()
	if again := h.mustSend(progress); again.Status != "activity_only" && again.Status != "replayed" {
		t.Fatalf("after restart %+v", again)
	}
	if task, _ = hostTaskByKey(h.current(), "host:4"); task.State != "completed" {
		t.Fatalf("after restart: %+v", task)
	}
}
