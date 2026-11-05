package viewer

import (
	"testing"
)

// Finding 3: a task update that does not know the title (the reconcile saw no
// create) is valid, and never replaces the title the mirror already has.
func TestHostTaskUpdateWithoutTitleKeepsTheKnownTitle(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	h.mustSend(h.event("task_created", hostTask("7", "Run typecheck"), id, stageID))
	h.advance()
	h.mustSend(h.event("task_updated", TimelineHostEventData{TaskID: "7", State: "in_progress"}, id, stageID))
	task, ok := hostTaskByKey(h.current(), "host:7")
	if !ok || task.Title != "Run typecheck" || task.State != "in_progress" {
		t.Fatalf("title or state after an update without a title: %+v", task)
	}
}

func TestHostTaskUpdateWithoutTitleOfAnUnknownTaskIsNamedByItsID(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	h.mustSend(h.event("task_updated", TimelineHostEventData{TaskID: "8", State: "in_progress"}, id, stageID))
	task, ok := hostTaskByKey(h.current(), "host:8")
	if !ok || task.Title != "Native task 8" {
		t.Fatalf("placeholder: %+v", task)
	}
	h.advance()
	h.mustSend(h.event("task_updated", hostTask("8", "Real title"), id, stageID))
	if task, _ = hostTaskByKey(h.current(), "host:8"); task.Title != "Real title" {
		t.Fatalf("a later real title must replace the placeholder: %+v", task)
	}
}

// Finding 6: once the agent registers tasks, the host tasks already mirrored keep
// following the host's state, and no new one is added.
func TestHostTasksAlreadyMirroredKeepUpdatingAfterTheAgentRegistersTasks(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	h.mustSend(h.event("task_created", hostTask("7", "Native one"), id, stageID))
	create := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, h.store, id, "agent-task", h.work.ForegroundProducerID), TaskKey: "T1", PlanOrder: 1, Title: "Plan task", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	if _, err := h.store.TimelineV3Operation(id, "task_create", "", &create); err != nil {
		t.Fatal(err)
	}
	h.advance()
	h.mustSend(h.event("task_updated", TimelineHostEventData{TaskID: "7", State: "in_progress"}, id, stageID))
	h.mustSend(h.event("task_created", hostTask("9", "Native two"), id, stageID))
	v := h.current()
	if task, _ := hostTaskByKey(v, "host:7"); task.State != "in_progress" {
		t.Fatalf("a mirrored task froze when the agent registered tasks: %+v", task)
	}
	if _, added := hostTaskByKey(v, "host:9"); added || len(v.Tasks) != 2 {
		t.Fatalf("a host task was added next to the agent's: %+v", v.Tasks)
	}
}

// Finding 9: an unnamed host event moves the activity of the session's open works
// only; a closed or archived one is touched only when it is named, and then only
// its activity moves.
func TestHostEventUpdatesOnlyOpenWorksOfTheSession(t *testing.T) {
	h := newHostEventHarness(t)
	open := h.work.ViewerSessionID
	other := func(key string) string {
		r := h.request
		r.OperationID, r.WorkKey = "create-"+key, timelineString(key)
		return createTimeline(t, h.store, r).ViewerSessionID
	}
	closed, archived := other("closed"), other("archived")
	capacityClose(t, h.store, closed, "end-closed")
	capacityClose(t, h.store, archived, "end-archived")
	yes := true
	if _, err := h.store.TimelinePatch(archived, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "archive", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: domainCurrent(t, h.store, archived).Version, Archived: &yes}); err != nil {
		t.Fatal(err)
	}
	activity := func(id string) string { return domainCurrent(t, h.store, id).LastActivityAt }
	before := map[string]string{open: activity(open), closed: activity(closed), archived: activity(archived)}
	h.advance()
	h.mustSend(h.event("activity", TimelineHostEventData{}, "", ""))
	if activity(open) == before[open] {
		t.Fatal("the open work did not move")
	}
	if activity(closed) != before[closed] || activity(archived) != before[archived] {
		t.Fatalf("a closed or archived work was rewritten: %s %s", activity(closed), activity(archived))
	}
	h.advance()
	if r := h.mustSend(h.event("activity", TimelineHostEventData{TurnID: "t2"}, closed, "")); r.Status != "activity_only" && r.Status != "applied" {
		t.Fatalf("named closed work: %+v", r)
	}
	if activity(closed) == before[closed] {
		t.Fatal("a closed work named as the current one still takes its activity")
	}
}

// Finding 8: the project is resolved, and Git asked for its remote, outside the
// store lock.
func TestHostEventResolvesTheProjectOutsideTheStoreLock(t *testing.T) {
	h := newHostEventHarness(t)
	calls, held := 0, 0
	h.store.remote = func(string) (string, error) {
		calls++
		if h.store.mu.TryLock() {
			h.store.mu.Unlock()
		} else {
			held++
		}
		return "", nil
	}
	h.advance()
	h.mustSend(h.event("activity", TimelineHostEventData{}, "", ""))
	if calls == 0 {
		t.Fatal("the remote was never read, so the test proves nothing")
	}
	if held != 0 {
		t.Fatalf("%d of %d remote reads ran under the store lock", held, calls)
	}
}
