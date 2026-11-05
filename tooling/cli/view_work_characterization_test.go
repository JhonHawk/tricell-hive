package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

// Characterization of `work stage` and `work task update` as T3 found them.
// They pin the behavior that T3 must keep: the new commands are built over
// these operations, which do not change.

func charRun(t *testing.T, c viewWorkCommand) (viewWorkResult, error) {
	t.Helper()
	var out bytes.Buffer
	if err := executeViewWork(c, &out); err != nil {
		return viewWorkResult{}, err
	}
	var r viewWorkResult
	if err := json.Unmarshal(out.Bytes(), &r); err != nil {
		t.Fatalf("invalid output %q", out.String())
	}
	return r, nil
}

func charStage(c viewWorkCommand, id, stage, state, reason string) viewWorkCommand {
	c.Action, c.WorkID, c.Stage, c.State, c.Reason = "stage", id, stage, state, reason
	return c
}

func TestWorkCharacterizationStageTransitions(t *testing.T) {
	_, c := workEngineFixture(t)
	opened := engineWorkOpen(t, c)
	id := opened.WorkID

	started, err := charRun(t, charStage(c, id, "build", "in_progress", "Start build"))
	if err != nil || started.Result != "updated" || started.StageID == "" {
		t.Fatalf("start: %+v %v", started, err)
	}
	again, err := charRun(t, charStage(c, id, "build", "in_progress", "Start build"))
	if err != nil || again.Result != "unchanged" || again.SessionVersion != started.SessionVersion {
		t.Fatalf("repeat must be unchanged without a mutation: %+v %v", again, err)
	}
	review, err := charRun(t, charStage(c, id, "build", "in_review", "Review"))
	if err != nil || review.Result != "updated" || review.SessionVersion <= started.SessionVersion {
		t.Fatalf("review: %+v %v", review, err)
	}
	done, err := charRun(t, charStage(c, id, "build", "completed", "Done"))
	if err != nil || done.Result != "updated" {
		t.Fatalf("complete: %+v %v", done, err)
	}
	// A terminal stage restarts only with an explicit predecessor.
	if _, err = charRun(t, charStage(c, id, "build", "in_progress", "Again")); err == nil || !strings.Contains(err.Error(), "explicit_stage_predecessor_required") {
		t.Fatalf("reentry without predecessor: %v", err)
	}
	// A stage that never started cannot be completed.
	if _, err = charRun(t, charStage(c, id, "verify", "completed", "Skip")); err == nil || !strings.Contains(err.Error(), "stage_start_required") {
		t.Fatalf("complete before start: %v", err)
	}
}

// charLowLevelTask registers a task through the viewer's task operation, with no
// binding of its segment to a stage occurrence: what a low-level publisher does.
func charLowLevelTask(t *testing.T, e *viewWorkEngine, id, key string, order int64, stage viewer.TimelineStage, assignee string) viewer.TimelineSession {
	t.Helper()
	s, err := e.session(id)
	if err != nil {
		t.Fatal(err)
	}
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: s.ForegroundProducerID}
	r := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Register explicit task"), TaskKey: key, Title: "Task " + key, PlanOrder: order, Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: stage, Order: order, AssignedProducerID: assignee}}}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = viewWorkHTTP(e.info, "POST", "/api/timeline/sessions/"+id+"/tasks", "application/json", raw); err != nil {
		t.Fatal(err)
	}
	s, err = e.session(id)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestWorkCharacterizationTaskUpdate(t *testing.T) {
	f, c := workEngineFixture(t)
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	id := opened.ViewerSessionID
	e.Close()

	// A task registered before its stage starts, with no binding.
	e, _ = openViewWorkEngine(c)
	s := charLowLevelTask(t, e, id, "T1", 1, "build", "")
	e.Close()
	if s.Tasks[0].Segments[0].DisplayStageEventID != "" {
		t.Fatal("the low-level registration unexpectedly bound the segment")
	}
	update := func(state string) (viewWorkResult, error) {
		u := c
		u.Action, u.WorkID, u.Key, u.Stage, u.State, u.Reason = "task update", id, "T1", "build", state, "Move task"
		return charRun(t, u)
	}
	// Without an active stage there is nothing to bind to.
	if _, err = update("in_progress"); err == nil || !strings.Contains(err.Error(), "active_stage_required") {
		t.Fatalf("update without a stage: %v", err)
	}
	if _, err = charRun(t, charStage(c, id, "build", "in_progress", "Start build")); err != nil {
		t.Fatal(err)
	}
	// `stage` start binds the segment, so the low-level binding error does not occur.
	first, err := update("in_progress")
	if err != nil || first.Result != "updated" || first.TaskID == "" {
		t.Fatalf("update after stage start: %+v %v", first, err)
	}
	e, _ = openViewWorkEngine(c)
	s, _ = e.session(id)
	seg := s.Tasks[0].Segments[0]
	if seg.State != "in_progress" || seg.DisplayStageEventID != *s.CurrentStageEventID {
		t.Fatalf("segment %q bound to %q", seg.State, seg.DisplayStageEventID)
	}
	e.Close()
	if again, err := update("in_progress"); err != nil || again.Result != "unchanged" {
		t.Fatalf("repeat: %+v %v", again, err)
	}
	// A task cannot skip from not_started to completed through the viewer rules.
	charLowLevelTaskOrder := func() {
		e, _ := openViewWorkEngine(c)
		defer e.Close()
		charLowLevelTask(t, e, id, "T2", 2, "build", "")
	}
	charLowLevelTaskOrder()
	skip := c
	skip.Action, skip.WorkID, skip.Key, skip.Stage, skip.State, skip.Reason = "task update", id, "T2", "build", "completed", "Skip"
	if _, err = charRun(t, skip); err == nil {
		t.Fatal("not_started to completed accepted")
	}
	// A task registered after the stage started and left unbound is bound by
	// the update itself.
	late := c
	late.Action, late.WorkID, late.Key, late.Stage, late.State, late.Reason = "task update", id, "T2", "build", "in_progress", "Late task"
	if r, err := charRun(t, late); err != nil || r.Result != "updated" {
		t.Fatalf("late task: %+v %v", r, err)
	}
	_ = f
}

func TestWorkCharacterizationLowLevelTransitionFailsUnbound(t *testing.T) {
	// The failure of #358, at the layer where it lives: the viewer rejects a
	// transition of a segment that is not bound to the stage occurrence.
	_, c := workEngineFixture(t)
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	id := opened.ViewerSessionID
	e.Close()
	if _, err = charRun(t, charStage(c, id, "build", "in_progress", "Start build")); err != nil {
		t.Fatal(err)
	}
	e, _ = openViewWorkEngine(c)
	s := charLowLevelTask(t, e, id, "T1", 1, "build", "")
	task, seg := s.Tasks[0], s.Tasks[0].Segments[0]
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: s.ForegroundProducerID}
	r := viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Move task"), ExpectedTaskVersion: task.TaskVersion, SegmentID: seg.SegmentID, State: "in_progress"}
	raw, _ := json.Marshal(r)
	_, code, err := viewWorkHTTP(e.info, "POST", "/api/timeline/sessions/"+id+"/tasks/"+task.TaskID+"/transitions", "application/json", raw)
	e.Close()
	if err == nil || code != "stage_binding_mismatch" {
		t.Fatalf("low-level transition: code %q err %v", code, err)
	}
}

func TestWorkCharacterizationTaskUpdateByAssigneeNeedsLeadBinding(t *testing.T) {
	// The finding for #358 through `work task update`: the update binds the
	// segment itself, and a binding is a lead-only write. A non-lead assignee
	// whose segment nobody bound therefore gets lead_required, not a binding
	// error; a segment the lead already bound moves.
	f, c := workEngineFixture(t)
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	id := opened.ViewerSessionID
	e.Close()
	scenario := &timelineScenario{t: t, f: f, input: filepath.Join(t.TempDir(), "request.json"), id: id, actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: opened.ForegroundProducerID}}
	current := scenario.current()
	joined := scenario.result([]string{"producer", "request", id}, viewer.TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "char-join", Source: viewer.TimelineSource{Host: "claude-code", NativeSessionID: "char-collaborator", ProjectRoot: c.Project}, ExpectedSessionVersion: current.Version, Reason: "Accept collaborator"})
	scenario.result([]string{"producer", "accept", id}, viewer.TimelineProducerAcceptRequest{TimelineNativeOperation: scenario.base("char-accept"), RequestID: joined.RequestID})
	if _, err = charRun(t, charStage(c, id, "build", "in_progress", "Start build")); err != nil {
		t.Fatal(err)
	}
	s := scenario.current()
	var collaborator string
	for _, p := range s.Producers {
		if p.Source.NativeSessionID == "char-collaborator" {
			collaborator = p.ProducerID
		}
	}
	e, _ = openViewWorkEngine(c)
	charLowLevelTask(t, e, id, "T1", 1, "build", collaborator)
	e.Close()
	other := c
	other.Host, other.NativeSessionID = "claude-code", "char-collaborator"
	other.Action, other.WorkID, other.Key, other.Stage, other.State, other.Reason = "task update", id, "T1", "build", "in_progress", "Collaborator starts"
	if _, err = charRun(t, other); err == nil || !strings.Contains(err.Error(), "lead_required") {
		t.Fatalf("assignee on an unbound segment: %v", err)
	}
	// The lead binds (a plain update of the lead's own view), then the assignee moves.
	lead := c
	lead.Action, lead.WorkID, lead.Key, lead.Stage, lead.State, lead.Reason = "task update", id, "T1", "build", "not_started", "Lead binds"
	if _, err = charRun(t, lead); err != nil {
		t.Fatal(err)
	}
	if r, err := charRun(t, other); err != nil || r.Result != "updated" {
		t.Fatalf("assignee after the lead bound: %+v %v", r, err)
	}
}
