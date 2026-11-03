package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"tricell-hive/tooling/viewer"
)

// Lifecycle adapters own semantic intent over a real isolated v3 server.
func TestWorkLifecycleStagesAndClosure(t *testing.T) {
	f := newViewFixture(t)
	repo := newProjectRepo(t)
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	c := viewWorkCommand{Project: repo, Change: repo, Key: "lifecycle", Title: "Lifecycle", Summary: "Isolated lifecycle", Host: "codex", NativeSessionID: "fixture-lifecycle"}
	e := &viewWorkEngine{dir: t.TempDir(), info: f.info(), journal: newWorkJournal()}
	s, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	c.WorkID = s.ViewerSessionID
	c.Action = "stage"
	c.Stage = "build"
	c.State = "in_progress"
	c.Reason = "Start implementation"
	start, err := e.lifecycle(c)
	if err != nil {
		t.Fatal(err)
	}
	c.State = "in_review"
	initial, err := e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := e.actor(initial, c)
	if err != nil {
		t.Fatal(err)
	}
	method, path, body, _, err := e.nextLifecycleStep(initial, actor, c)
	if err != nil {
		t.Fatal(err)
	}
	pending, err := e.prepareIntent(c.Action, c.WorkID, workLifecycleSource(initial, actor), viewWorkLifecycleContext{Command: c, FirstStep: &viewWorkStep{Method: method, Path: path, ContentType: "application/json", Body: body}})
	if err != nil {
		t.Fatal(err)
	}
	// Recover the prepared semantic intent before its first append/send.
	if _, err = e.resumeLifecycle(pending); err != nil {
		t.Fatal(err)
	}
	current, err := e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	if current.State == viewer.TimelineClosed {
		t.Fatal("review closed work")
	}
	close := c
	close.Action = "close"
	close.Outcome = "completed"
	if _, err = e.lifecycle(close); err == nil {
		t.Fatal("closed active stage")
	}
	c.State = "completed"
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	c.State = "in_progress"
	if _, err = e.lifecycle(c); err == nil {
		t.Fatal("reentered without explicit predecessor")
	}
	c.PreviousStage = start.StageID
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	c.State = "completed"
	c.PreviousStage = ""
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	if _, err = e.lifecycle(close); err != nil {
		t.Fatal(err)
	}
	current, err = e.session(c.WorkID)
	if err != nil || current.State != viewer.TimelineClosed {
		t.Fatalf("close: %s %v", current.State, err)
	}
	// Binding state remains local, separate from the server inventory.
	if filepath.Base(e.dir) == "viewer" {
		t.Fatal("test publisher shares viewer state")
	}
}

func TestWorkLifecycleTaskBatchAndSegmentRecovery(t *testing.T) {
	f, c := workEngineFixture(t)
	e := &viewWorkEngine{dir: t.TempDir(), info: f.info(), journal: newWorkJournal()}
	s, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	c.WorkID = s.ViewerSessionID
	c.Action = "stage"
	c.Stage = "plan"
	c.State = "in_progress"
	c.Reason = "Declare plan"
	plan, err := e.lifecycle(c)
	if err != nil {
		t.Fatal(err)
	}
	scenario := &timelineScenario{t: t, f: f, input: filepath.Join(t.TempDir(), "request.json"), id: c.WorkID, actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: s.ForegroundProducerID}}
	file := filepath.Join(t.TempDir(), "plan.md")
	if err := os.WriteFile(file, []byte("# Plan\n"), 0600); err != nil {
		t.Fatal(err)
	}
	doc := scenario.result([]string{"document", "add", c.WorkID, "--file", file}, viewer.TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "lifecycle-document", Actor: scenario.actor, MediaType: "text/markdown"})
	s = scenario.current()
	event := scenarioEvent(t, s, plan.StageID)
	revision := scenario.result([]string{"stage", "snapshot", c.WorkID, plan.StageID}, viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: "lifecycle-plan", Actor: scenario.actor, DocumentKey: "plan", ExpectedVersion: event.Version, ExpectedSessionVersion: s.Version, Status: "draft", Tabs: []viewer.TimelineDocumentTab{{Label: "Plan", DocumentID: doc.DocumentID}}})
	register := c
	register.Action = "tasks register"
	register.PlanRevision = revision.DocumentRevisionID
	source := &viewWorkSource{Host: "codex", NativeSessionID: c.NativeSessionID}
	register.Tasks = []viewWorkTaskInput{{Key: "T1", Title: "Implement", Summary: "Build and verify separately", Order: 1, Segments: []viewWorkSegmentInput{{Stage: "build", Order: 1, Assignee: source}, {Stage: "verify", Order: 1, Assignee: source}}}, {Key: "T2", Title: "Document", Order: 2, Segments: []viewWorkSegmentInput{{Stage: "build", Order: 2, Assignee: source}}}}
	// Fail the whole batch before first mutation when the final member conflicts.
	invalid := register
	invalid.Tasks = append([]viewWorkTaskInput{}, register.Tasks...)
	invalid.Tasks[1].Order = 1
	before := scenario.current().Version
	if _, err := e.lifecycle(invalid); err == nil {
		t.Fatal("accepted conflicting final task")
	}
	if scenario.current().Version != before {
		t.Fatal("invalid batch partially mutated")
	}
	// Persist and execute exactly the first task, then recover the remainder from
	// a reloaded journal, reproducing process loss between batch members.
	s = scenario.current()
	actor := scenario.actor
	intent, err := e.prepareIntent(register.Action, c.WorkID, workLifecycleSource(s, actor), viewWorkLifecycleContext{Command: register})
	if err != nil {
		t.Fatal(err)
	}
	method, path, body, _, err := e.nextLifecycleStep(s, actor, register)
	if err != nil {
		t.Fatal(err)
	}
	index, err := e.appendStep(intent, method, path, "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.executeStep(intent, index); err != nil {
		t.Fatal(err)
	}
	e.journal, err = readWorkJournal(e.dir)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a lost acknowledgement after server commit; exact replay must
	// recover the first receipt and continue without registering it twice.
	e.journal.Pending[0].Steps[0].Response = nil
	if _, err = e.resumeLifecycle(e.journal.Pending[0]); err != nil {
		t.Fatal(err)
	}
	s = scenario.current()
	if len(s.Tasks) != 2 {
		t.Fatalf("partial batch recovery created %d tasks", len(s.Tasks))
	}
	before = s.Version
	if _, err = e.lifecycle(register); err != nil {
		t.Fatal(err)
	}
	if scenario.current().Version != before {
		t.Fatal("identical registration mutated registry")
	}
	c.State = "completed"
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	c.Stage = "build"
	c.State = "in_progress"
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	update := c
	update.Action = "task update"
	update.Key = "T1"
	update.State = "in_progress"
	update.Reason = "Start build segment"
	unknown := update
	unknown.NativeSessionID = "unaccepted"
	before = scenario.current().Version
	if _, err = e.lifecycle(unknown); err == nil {
		t.Fatal("unaccepted source updated task")
	}
	if scenario.current().Version != before {
		t.Fatal("unaccepted source changed registry")
	}
	if _, err = e.lifecycle(update); err != nil {
		t.Fatal(err)
	}
	task := workLifecycleTask(scenario.current(), "T1")
	if task.Segments[0].State != "in_progress" || task.Segments[1].State != "not_started" {
		t.Fatal("build transition changed verify segment")
	}
	update.State = "blocked"
	if _, err = e.lifecycle(update); err == nil {
		t.Fatal("accepted incomplete block")
	}
	update.Block = &viewer.TimelineBlock{Reason: "Missing fixture", Owner: "Fixture owner", Steps: []string{"Restore fixture"}, ResumeWhen: "Fixture passes"}
	if _, err = e.lifecycle(update); err != nil {
		t.Fatal(err)
	}
	update.State = "in_progress"
	update.Block = nil
	if _, err = e.lifecycle(update); err == nil {
		t.Fatal("resolved block without evidence")
	}
	update.Resolution = "Fixture restored and checked"
	if _, err = e.lifecycle(update); err != nil {
		t.Fatal(err)
	}
	order := int64(3)
	update.Metadata = &viewWorkTaskMetadata{Order: &order}
	if _, err = e.lifecycle(update); err == nil {
		t.Fatal("changed scope without reference")
	}
	update.Metadata.PlanRevision = register.PlanRevision
	if _, err = e.lifecycle(update); err != nil {
		t.Fatal(err)
	}
	task = workLifecycleTask(scenario.current(), "T1")
	if task.PlanOrder != 3 || task.Segments[0].State != "in_progress" {
		t.Fatal("metadata update reset task progress")
	}
}

func TestWorkLifecycleBindingConflictRetainsReconcileableRequest(t *testing.T) {
	f, c := workEngineFixture(t)
	e := &viewWorkEngine{dir: t.TempDir(), info: f.info(), journal: newWorkJournal()}
	s, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	c.WorkID = s.ViewerSessionID
	c.Action = "stage"
	c.Stage = "build"
	c.State = "in_progress"
	c.Reason = "Start build"
	if _, err = e.lifecycle(c); err != nil {
		t.Fatal(err)
	}
	// Create a task using the existing low-level API; this case owns the gap
	// between display binding and transition rather than plan registration.
	s, err = e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	actor, _ := e.actor(s, c)
	create := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Declare task"), TaskKey: "T1", Title: "Concurrent task", PlanOrder: 1, Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	raw, err := json.Marshal(create)
	if err != nil {
		t.Fatal(err)
	}
	response, _, err := viewWorkHTTP(e.info, "POST", "/api/timeline/sessions/"+c.WorkID+"/tasks", "application/json", raw)
	if err != nil {
		t.Fatal(err)
	}
	made, err := viewer.DecodeTimelineOperationResult(response)
	if err != nil {
		t.Fatal(err)
	}
	c.Action = "task update"
	c.Key = "T1"
	c.State = "in_progress"
	scenario := &timelineScenario{t: t, f: f, input: filepath.Join(t.TempDir(), "request.json"), id: c.WorkID, actor: actor}
	now := scenario.current()
	joined := scenario.result([]string{"producer", "request", c.WorkID}, viewer.TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "lifecycle-join", Source: viewer.TimelineSource{Host: "claude-code", NativeSessionID: "accepted-other", ProjectRoot: c.Project}, ExpectedSessionVersion: now.Version, Reason: "Join isolated test work"})
	scenario.result([]string{"producer", "accept", c.WorkID}, viewer.TimelineProducerAcceptRequest{TimelineNativeOperation: scenario.base("lifecycle-accept"), RequestID: joined.RequestID})
	wrong := c
	wrong.Host = "claude-code"
	wrong.NativeSessionID = "accepted-other"
	if _, err = e.lifecycle(wrong); err == nil {
		t.Fatal("accepted producer updated another producer's segment")
	}
	if len(e.journal.Pending) != 0 {
		t.Fatal("denied assignee left a pending mutation")
	}
	s, err = e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	method, path, body, _, err := e.nextLifecycleStep(s, actor, c)
	if err != nil {
		t.Fatal(err)
	}
	intent, err := e.prepareIntent(c.Action, c.WorkID, workLifecycleSource(s, actor), viewWorkLifecycleContext{Command: c, FirstStep: &viewWorkStep{Method: method, Path: path, ContentType: "application/json", Body: body}})
	if err != nil {
		t.Fatal(err)
	}
	// Interrupt after the confirmed binding, before preparing the transition.
	if _, err = e.appendStep(intent, method, path, "application/json", body); err != nil {
		t.Fatal(err)
	}
	if _, err = e.executeStep(intent, 0); err != nil {
		t.Fatal(err)
	}
	s, err = e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	task := workLifecycleTask(s, "T1")
	title := "External metadata update"
	patch := viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Concurrent metadata"), ExpectedTaskVersion: task.TaskVersion, Title: &title}
	raw, _ = json.Marshal(patch)
	if _, _, err = viewWorkHTTP(e.info, "PATCH", "/api/timeline/sessions/"+c.WorkID+"/tasks/"+made.TaskID, "application/json", raw); err != nil {
		t.Fatal(err)
	}
	if _, err = e.resumeLifecycle(intent); err == nil {
		t.Fatal("overwrote concurrent task update")
	}
	if intent.State != "rejected_conflict" || len(intent.Steps) != 2 || intent.Steps[1].Rejection == "" {
		t.Fatal("conflict lacks exact rejected request for reconciliation")
	}
	if err := e.journal.archiveRejected(intent.ID, "Retain conflict and inspect concurrent update"); err != nil {
		t.Fatal(err)
	}
	s, err = e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	task = workLifecycleTask(s, "T1")
	if task.Title != title || task.Segments[0].State != "not_started" {
		t.Fatal("conflicting transition changed external state")
	}
}
