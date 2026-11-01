package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"tricell-hive/tooling/viewer"
)

// timelineLocalJSON exercises browser-only local-viewer endpoints without
// leaking access headers into test output or impersonating a native producer.
func timelineLocalJSON(t *testing.T, f *viewFixture, method, path string, value any) string {
	t.Helper()
	status, body := timelineLocalStatus(t, f, method, path, value)
	if status != 200 {
		t.Fatalf("local viewer request status %d", status)
	}
	return body
}
func timelineLocalStatus(t *testing.T, f *viewFixture, method, path string, value any) (int, string) {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	info := f.info()
	request, err := http.NewRequest(method, info.url()+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Hive-Token", info.Token)
	request.Header.Set("Origin", info.url())
	request.Header.Set("X-Hive-Action-Scope", "session")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal("local viewer request failed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		t.Fatal("local viewer response incomplete")
	}
	return response.StatusCode, string(body)
}

type timelineScenario struct {
	t      *testing.T
	f      *viewFixture
	input  string
	id     string
	actor  viewer.TimelineActor
	serial int
}

func (s *timelineScenario) call(args []string, value any) (string, error) {
	s.t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		s.t.Fatal(err)
	}
	if err = os.WriteFile(s.input, raw, 0600); err != nil {
		s.t.Fatal(err)
	}
	return s.f.view(append(args, "--json", s.input)...)
}
func (s *timelineScenario) invoke(args []string, value any) string {
	s.t.Helper()
	raw, err := s.call(args, value)
	if err != nil {
		s.t.Fatal(args, err)
	}
	return raw
}
func (s *timelineScenario) result(args []string, value any) viewer.TimelineOperationResult {
	s.t.Helper()
	raw := s.invoke(args, value)
	result, err := viewer.DecodeTimelineOperationResult([]byte(raw))
	if err != nil {
		s.t.Fatal(err)
	}
	return result
}
func (s *timelineScenario) current() viewer.TimelineSession {
	s.t.Helper()
	raw := s.f.mustView("session", "get", s.id)
	v, err := viewer.DecodeTimelineSessionResponse([]byte(raw))
	if err != nil {
		s.t.Fatal(err)
	}
	return v.Session
}
func (s *timelineScenario) op(prefix string) string {
	s.serial++
	return fmt.Sprintf("%s-%d", prefix, s.serial)
}
func (s *timelineScenario) base(prefix string) viewer.TimelineNativeOperation {
	v := s.current()
	return viewer.TimelineNativeOperation{SchemaVersion: 2, OperationID: s.op(prefix), Actor: s.actor, ExpectedSessionVersion: v.Version, Reason: "Explicit isolated CLI " + prefix}
}
func (s *timelineScenario) stage(kind viewer.TimelineStage) string {
	s.t.Helper()
	v := s.current()
	return s.result([]string{"stage", "start", s.id}, viewer.TimelineAppendEventRequest{SchemaVersion: 2, OperationID: s.op("stage"), Actor: s.actor, Reason: "Start " + string(kind), Kind: "stage", Stage: kind, StageState: "in_progress", ExpectedSessionVersion: &v.Version, Title: string(kind), Summary: "Isolated end-to-end stage", TicketRefs: []viewer.TimelineTicketRef{}}).EventID
}
func scenarioEvent(t *testing.T, v viewer.TimelineSession, id string) viewer.TimelineEvent {
	t.Helper()
	for _, e := range v.Events {
		if e.EventID == id {
			return e
		}
	}
	t.Fatal("event missing")
	return viewer.TimelineEvent{}
}
func scenarioTask(t *testing.T, v viewer.TimelineSession, id string) viewer.TimelineTask {
	t.Helper()
	for _, task := range v.Tasks {
		if task.TaskID == id {
			return task
		}
	}
	t.Fatal("task missing")
	return viewer.TimelineTask{}
}
func (s *timelineScenario) finish(stage string) {
	v := s.current()
	event := scenarioEvent(s.t, v, stage)
	done := viewer.TimelineStageState("completed")
	s.result([]string{"stage", "update", s.id, stage}, viewer.TimelineEventPatchRequest{SchemaVersion: 2, OperationID: s.op("finish"), Actor: s.actor, ExpectedVersion: event.Version, ExpectedSessionVersion: &v.Version, StageState: &done, Reason: "Stage completed explicitly"})
}
func (s *timelineScenario) bind(task, segment, stage string) {
	v := s.current()
	current := scenarioTask(s.t, v, task)
	s.result([]string{"task", "update", s.id, task}, viewer.TimelineTaskPatchRequest{TimelineNativeOperation: s.base("bind"), ExpectedTaskVersion: current.TaskVersion, SegmentID: segment, DisplayStageEventID: stage})
}
func (s *timelineScenario) transition(task, segment, state string, block *viewer.TimelineBlock, resolution string) {
	v := s.current()
	current := scenarioTask(s.t, v, task)
	s.result([]string{"task", "transition", s.id, task}, viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: s.base("transition"), ExpectedTaskVersion: current.TaskVersion, SegmentID: segment, State: state, Block: block, Resolution: resolution})
}
func (s *timelineScenario) denied(args []string, value any) {
	s.t.Helper()
	before := s.current()
	_, err := s.call(args, value)
	if err == nil {
		s.t.Fatal("denied operation succeeded", args)
	}
	after := s.current()
	if !reflect.DeepEqual(before, after) {
		s.t.Fatal("denied operation changed state")
	}
	raw, _ := json.Marshal(value)
	stored, _ := os.ReadFile(s.input)
	if !bytes.Equal(raw, stored) {
		s.t.Fatal("denied operation changed its input")
	}
}
func completedScenarioTasks(v viewer.TimelineSession) int {
	n := 0
	for _, task := range v.Tasks {
		if !task.Withdrawn && task.State == "completed" {
			n++
		}
	}
	return n
}

func TestTimelineCLIV2SharedTasksCIAndRestart(t *testing.T) {
	f := newViewFixture(t)
	repo := newProjectRepo(t)
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	s := &timelineScenario{t: t, f: f, input: filepath.Join(t.TempDir(), "request.json")}
	create := viewer.TimelineCreateSessionRequest{SchemaVersion: 2, OperationID: "shared-create", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "fixture-A"}, Title: "Shared CLI fixture", Summary: "Synthetic evidence; no provider calls or integration", Tickets: []viewer.TimelineTicket{}}
	response, err := viewer.DecodeTimelineSessionResponse([]byte(s.invoke([]string{"session", "create"}, create)))
	if err != nil {
		t.Fatal(err)
	}
	s.id = response.Session.ViewerSessionID
	s.actor = viewer.TimelineActor{Kind: "native_producer", ProducerID: response.Session.ForegroundProducerID}
	originalActor := s.actor
	legacy := create
	legacy.SchemaVersion = 1
	legacy.OperationID = "old-client-write"
	legacy.Source.NativeSessionID = "fixture-v1"
	status, legacyBody := timelineLocalStatus(t, f, "POST", "/api/timeline/sessions", legacy)
	var legacyFailure struct {
		Code string `json:"code"`
	}
	json.Unmarshal([]byte(legacyBody), &legacyFailure)
	if status != 400 || legacyFailure.Code != "unsupported_protocol" {
		t.Fatal("v1 write was accepted by v2 viewer")
	}
	plan := s.stage("plan")
	sourceB := viewer.TimelineSource{ProjectRoot: repo, Host: "claude-code", NativeSessionID: "fixture-B"}
	current := s.current()
	join := viewer.TimelineJoinRequestInput{SchemaVersion: 2, OperationID: "shared-join", Source: sourceB, ExpectedSessionVersion: current.Version, Reason: "Explicitly share the selected work"}
	joinedRaw := s.invoke([]string{"producer", "request", s.id}, join)
	joined, err := viewer.DecodeTimelineOperationResult([]byte(joinedRaw))
	if err != nil {
		t.Fatal(err)
	}
	s.denied([]string{"event", "add", s.id}, viewer.TimelineAppendEventRequest{SchemaVersion: 2, OperationID: "pending-cannot-publish", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: "pending"}, Kind: "review", ReviewStatus: "in_progress", TargetStageEventID: plan, Title: "Rejected pending review", Summary: "Must not publish", TicketRefs: []viewer.TimelineTicketRef{}})
	accepted := s.result([]string{"producer", "accept", s.id}, viewer.TimelineProducerAcceptRequest{TimelineNativeOperation: s.base("accept"), RequestID: joined.RequestID})
	actorB := viewer.TimelineActor{Kind: "native_producer", ProducerID: accepted.ProducerID}
	s.actor = actorB
	s.result([]string{"event", "add", s.id}, viewer.TimelineAppendEventRequest{SchemaVersion: 2, OperationID: "parallel-review", Actor: actorB, Kind: "review", ReviewStatus: "completed", TargetStageEventID: plan, Title: "Parallel review", Summary: "Does not take foreground", TicketRefs: []viewer.TimelineTicketRef{}})
	if v := s.current(); v.ForegroundProducerID != originalActor.ProducerID || v.CurrentStageEventID == nil || *v.CurrentStageEventID != plan {
		t.Fatal("parallel review took foreground")
	}
	s.actor = originalActor
	docs := map[string]viewer.TimelineOperationResult{}
	revisions := map[string]string{}
	contents := map[string]string{}
	for _, key := range []string{"architecture", "scope"} {
		file := filepath.Join(t.TempDir(), key+".md")
		contents[key] = "# " + key + "\n\nImmutable fixture document.\n"
		if err = os.WriteFile(file, []byte(contents[key]), 0600); err != nil {
			t.Fatal(err)
		}
		docs[key] = s.result([]string{"document", "add", s.id, "--file", file}, viewer.TimelineDocumentMetadata{SchemaVersion: 2, OperationID: s.op("upload"), Actor: s.actor, MediaType: "text/markdown"})
		v := s.current()
		event := scenarioEvent(t, v, plan)
		snapshot := s.result([]string{"stage", "snapshot", s.id, plan}, viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 2, OperationID: s.op("snapshot"), Actor: s.actor, DocumentKey: key, ExpectedVersion: event.Version, ExpectedSessionVersion: v.Version, Status: "complete", Tabs: []viewer.TimelineDocumentTab{{Label: key, DocumentID: docs[key].DocumentID}}})
		revisions[key] = snapshot.DocumentRevisionID
	}
	s.result([]string{"producer", "transfer", s.id}, viewer.TimelineForegroundTransferRequest{TimelineNativeOperation: s.base("transfer"), TargetProducerID: actorB.ProducerID})
	s.actor = actorB
	// B continues A's document chain without replacing the independent scope chain.
	previousArchitecture := docs["architecture"]
	previousContent := contents["architecture"]
	revisedFile := filepath.Join(t.TempDir(), "architecture-revised.md")
	contents["architecture"] = "# architecture\n\nExplicit continuation by producer B.\n"
	if err = os.WriteFile(revisedFile, []byte(contents["architecture"]), 0600); err != nil {
		t.Fatal(err)
	}
	docs["architecture"] = s.result([]string{"document", "add", s.id, "--file", revisedFile}, viewer.TimelineDocumentMetadata{SchemaVersion: 2, OperationID: s.op("revised-document"), Actor: s.actor, MediaType: "text/markdown"})
	current = s.current()
	planEvent := scenarioEvent(t, current, plan)
	revised := s.result([]string{"stage", "snapshot", s.id, plan}, viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 2, OperationID: s.op("continued-chain"), Actor: s.actor, DocumentKey: "architecture", PredecessorRevisionID: revisions["architecture"], ExpectedVersion: planEvent.Version, ExpectedSessionVersion: current.Version, Status: "complete", Tabs: []viewer.TimelineDocumentTab{{Label: "architecture", DocumentID: docs["architecture"].DocumentID}}})
	revisions["architecture"] = revised.DocumentRevisionID
	retained, decodeErr := viewer.DecodeTimelineDocumentResponse([]byte(f.mustView("document", "get", s.id, previousArchitecture.DocumentID)))
	if decodeErr != nil || retained.Content != previousContent {
		t.Fatal("continued chain changed original bytes", decodeErr)
	}
	s.finish(plan)
	// A phase-looking title never supplies the missing explicit segment phase.
	s.denied([]string{"task", "create", s.id}, viewer.TimelineTaskCreateRequest{TimelineNativeOperation: s.base("missing-phase"), TaskKey: "T8", Title: "Build implementation", Summary: "No inferred phase", Segments: []viewer.TimelineTaskSegmentInput{}})
	tasks := []viewer.TimelineTask{}
	for i := 1; i <= 7; i++ {
		segments := []viewer.TimelineTaskSegmentInput{{PlannedStage: "build", Order: int64(i)}}
		if i == 7 {
			segments = []viewer.TimelineTaskSegmentInput{{PlannedStage: "verify", Order: 1}, {PlannedStage: "delivery", Order: 1}}
		}
		result := s.result([]string{"task", "create", s.id}, viewer.TimelineTaskCreateRequest{TimelineNativeOperation: s.base("declare"), TaskKey: fmt.Sprintf("T%d", i), Title: fmt.Sprintf("Explicit task %d", i), Summary: "Declared phases, never inferred", Segments: segments, PlanRevisionID: revisions["scope"]})
		tasks = append(tasks, scenarioTask(t, s.current(), result.TaskID))
	}
	build := s.stage("build")
	for i := 0; i < 6; i++ {
		task := tasks[i]
		segment := task.Segments[0].SegmentID
		s.bind(task.TaskID, segment, build)
		if i == 2 {
			currentTask := scenarioTask(t, s.current(), task.TaskID)
			s.result([]string{"task", "update", s.id, task.TaskID}, viewer.TimelineTaskPatchRequest{TimelineNativeOperation: s.base("assign-original-producer"), ExpectedTaskVersion: currentTask.TaskVersion, SegmentID: segment, AssignedProducerID: originalActor.ProducerID, PlanRevisionID: revisions["scope"]})
			s.actor = originalActor
		}
		if i == 0 {
			s.transition(task.TaskID, segment, "blocked", &viewer.TimelineBlock{Reason: "Synthetic access prerequisite", Steps: []string{"Restore the isolated fixture"}, Owner: "Fixture lead", ResumeWhen: "Fixture validation passes"}, "")
			s.transition(task.TaskID, segment, "in_progress", nil, "Fixture restored and validated")
		} else {
			s.transition(task.TaskID, segment, "in_progress", nil, "")
		}
		s.transition(task.TaskID, segment, "completed", nil, "")
		if i == 0 {
			s.transition(task.TaskID, segment, "in_progress", nil, "")
			s.transition(task.TaskID, segment, "completed", nil, "")
		}
		s.actor = actorB
	}
	current = s.current()
	if completedScenarioTasks(current) != 6 || len(current.Tasks) != 7 {
		t.Fatal("build did not remain 6/7 unique tasks")
	}
	wait := viewer.TimelineAwaitingValidation
	s.result([]string{"session", "update", s.id}, viewer.TimelineSessionPatchRequest{SchemaVersion: 2, OperationID: s.op("ask"), Actor: s.actor, ExpectedVersion: current.Version, State: &wait, Reason: "Synthetic user validation pause"})
	current = s.current()
	if current.CurrentStageEventID == nil || *current.CurrentStageEventID != build {
		t.Fatal("validation lost foreground")
	}
	s.result([]string{"event", "add", s.id}, viewer.TimelineAppendEventRequest{SchemaVersion: 2, OperationID: s.op("response"), Actor: s.actor, Kind: "approval_reported", TargetStageEventID: build, Title: "Synthetic reported response", Summary: "Fixture evidence only", TicketRefs: []viewer.TimelineTicketRef{}, Target: &viewer.TimelineEvidenceTarget{Kind: "document", EventID: plan, DocumentRevisionID: revisions["architecture"]}, SourceReference: "Synthetic fixture, not an authenticated human approval"})
	current = s.current()
	active := viewer.TimelineActive
	s.result([]string{"session", "update", s.id}, viewer.TimelineSessionPatchRequest{SchemaVersion: 2, OperationID: s.op("resume"), Actor: s.actor, ExpectedVersion: current.Version, State: &active, Reason: "Explicit continuation after fixture response"})
	s.finish(build)
	verify := s.stage("verify")
	last := tasks[6]
	s.bind(last.TaskID, last.Segments[0].SegmentID, verify)
	s.transition(last.TaskID, last.Segments[0].SegmentID, "in_progress", nil, "")
	s.transition(last.TaskID, last.Segments[0].SegmentID, "completed", nil, "")
	current = s.current()
	if completedScenarioTasks(current) != 6 || scenarioTask(t, current, last.TaskID).State == "completed" {
		t.Fatal("verify alone completed T7 globally")
	}
	s.finish(verify)
	delivery := s.stage("delivery")
	s.bind(last.TaskID, last.Segments[1].SegmentID, delivery)
	policy := viewer.TimelineCheckPolicy{PolicyID: "unknown", Confirmed: false, Authority: "No confirmed access", SourceURL: "https://example.test/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []viewer.TimelineRequiredCheck{}}
	waiting := viewer.TimelineWaitCreateRequest{TimelineNativeOperation: s.base("unknown-policy"), TargetStageEventID: delivery, TaskID: last.TaskID, SegmentID: last.Segments[1].SegmentID, Candidate: viewer.TimelineCandidate{Branch: "feature", Commit: "unknown", Base: "development"}, PullRequestURL: "https://example.test/pr/1", Policy: policy}
	unknown := s.result([]string{"wait", "create", s.id}, waiting)
	if s.current().Waits[0].State != "unknown" {
		t.Fatal("unconfirmed checks became green")
	}
	waiting.TimelineNativeOperation = s.base("known-policy")
	waiting.PreviousWaitID = unknown.WaitID
	waiting.Candidate.Commit = "abc"
	waiting.Policy = viewer.TimelineCheckPolicy{PolicyID: "required", Confirmed: true, Authority: "Synthetic provider snapshot", SourceURL: "https://example.test/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []viewer.TimelineRequiredCheck{{CheckKey: "unit", Label: "Unit"}, {CheckKey: "lint", Label: "Lint"}}}
	made := s.result([]string{"wait", "create", s.id}, waiting)
	wrongCandidate := viewer.TimelineWaitAttemptStartRequest{TimelineNativeOperation: s.base("wrong-candidate"), ExpectedWaitVersion: s.current().Waits[1].WaitVersion, CheckKeys: []string{"unit"}, ExternalRun: viewer.TimelineCIExternalRun{Commit: "other-commit", RunID: "wrong", URL: "https://example.test/runs/wrong"}, SourceURL: "https://example.test/runs/wrong", ObservedAt: "2026-01-01T00:00:00Z"}
	s.denied([]string{"wait", "attempt", s.id, made.WaitID}, wrongCandidate)
	var first viewer.TimelineOperationResult
	var late viewer.TimelineWaitObserveRequest
	var lateRaw string
	for attempt := 1; attempt <= 3; attempt++ {
		current = s.current()
		w := current.Waits[1]
		run := fmt.Sprintf("run-%d", attempt)
		external := viewer.TimelineCIExternalRun{Commit: "abc", RunID: run, URL: "https://example.test/runs/" + run}
		started := s.result([]string{"wait", "attempt", s.id, made.WaitID}, viewer.TimelineWaitAttemptStartRequest{TimelineNativeOperation: s.base("attempt"), ExpectedWaitVersion: w.WaitVersion, CheckKeys: []string{"unit", "lint"}, ExternalRun: external, SourceURL: external.URL, ObservedAt: "2026-01-01T00:00:00Z"})
		if attempt == 1 {
			first = started
			// Insufficient provider visibility is reported as unknown, never success.
			s.result([]string{"wait", "observe", s.id, made.WaitID}, viewer.TimelineWaitObserveRequest{TimelineNativeOperation: s.base("provider-access-denied"), ExpectedWaitVersion: s.current().Waits[1].WaitVersion, AttemptID: started.AttemptID, CheckKey: "unit", ExternalRun: external, State: "unknown", Summary: "Synthetic provider snapshot: permission denied", ObservedAt: "2026-01-01T01:00:00Z"})
			if s.current().Waits[1].State == "succeeded" {
				t.Fatal("provider access denial became green")
			}
		}
		for _, check := range []string{"unit", "lint"} {
			current = s.current()
			state := "succeeded"
			if attempt == 2 {
				state = "failed"
			}
			observed := viewer.TimelineWaitObserveRequest{TimelineNativeOperation: s.base("observe"), ExpectedWaitVersion: current.Waits[1].WaitVersion, AttemptID: started.AttemptID, CheckKey: check, ExternalRun: external, State: state, Summary: "Synthetic provider outcome", ObservedAt: "2026-01-02T00:00:00Z"}
			s.result([]string{"wait", "observe", s.id, made.WaitID}, observed)
		}
		if attempt == 2 {
			current = s.current()
			late = viewer.TimelineWaitObserveRequest{TimelineNativeOperation: s.base("late-success"), ExpectedWaitVersion: current.Waits[1].WaitVersion, AttemptID: first.AttemptID, CheckKey: "unit", ExternalRun: viewer.TimelineCIExternalRun{Commit: "abc", RunID: "run-1", URL: "https://example.test/runs/run-1"}, State: "succeeded", Summary: "Late older-attempt success", ObservedAt: "2026-01-03T00:00:00Z"}
			lateRaw = s.invoke([]string{"wait", "observe", s.id, made.WaitID}, late)
			if s.current().Waits[1].State != "failed" {
				t.Fatal("late success replaced current failed attempt")
			}
			f.mustView("stop")
			f.mustView("start")
			if s.invoke([]string{"wait", "observe", s.id, made.WaitID}, late) != lateRaw || s.current().Waits[1].State != "failed" {
				t.Fatal("restart lost applicable attempt or receipt")
			}
		}
	}
	current = s.current()
	if current.Waits[1].State != "succeeded" || completedScenarioTasks(current) != 6 {
		t.Fatal("CI outcome fabricated task completion")
	}
	s.transition(last.TaskID, last.Segments[1].SegmentID, "in_progress", nil, "")
	s.transition(last.TaskID, last.Segments[1].SegmentID, "completed", nil, "")
	if completedScenarioTasks(s.current()) != 7 {
		t.Fatal("final task did not reach 7/7")
	}
	s.finish(delivery)
	s.result([]string{"event", "add", s.id}, viewer.TimelineAppendEventRequest{SchemaVersion: 2, OperationID: s.op("close"), Actor: s.actor, Kind: "closure", Title: "Fixture finished", Summary: "No provider calls, reruns or merge performed", Reason: "Explicit fixture closure", TicketRefs: []viewer.TimelineTicketRef{}, Closure: &viewer.TimelineClosure{Reason: "completed", Result: "Isolated flow verified; no integration claimed"}})
	f.mustView("stop")
	f.mustView("start")
	current = s.current()
	if current.Source != create.Source || len(current.Producers) != 2 || current.ForegroundProducerID != actorB.ProducerID || current.State != viewer.TimelineClosed || completedScenarioTasks(current) != 7 {
		t.Fatal("restart changed work identity or task completion")
	}
	for key, document := range docs {
		raw := f.mustView("document", "get", s.id, document.DocumentID)
		read, err := viewer.DecodeTimelineDocumentResponse([]byte(raw))
		if err != nil || read.Content != contents[key] {
			t.Fatal("independent document chain lost bytes", err)
		}
	}
	if s.invoke([]string{"producer", "request", s.id}, join) != joinedRaw {
		t.Fatal("join receipt changed after restart and closure")
	}
	f.mustView("stop")
}

func TestTimelineCLIV2LocalRecoveryHTTPConcurrencyAndRestart(t *testing.T) {
	f := newViewFixture(t)
	repo := newProjectRepo(t)
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	s := &timelineScenario{t: t, f: f, input: filepath.Join(t.TempDir(), "recovery.json")}
	create := viewer.TimelineCreateSessionRequest{SchemaVersion: 2, OperationID: "recovery-create", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "recovery-original"}, Title: "Isolated recovery fixture", Summary: "No process-authentication claim", Tickets: []viewer.TimelineTicket{}}
	initial, err := viewer.DecodeTimelineSessionResponse([]byte(s.invoke([]string{"session", "create"}, create)))
	if err != nil {
		t.Fatal(err)
	}
	s.id = initial.Session.ViewerSessionID
	s.actor = viewer.TimelineActor{Kind: "native_producer", ProducerID: initial.Session.ForegroundProducerID}
	stage := s.stage("plan")
	requests := []string{}
	for _, native := range []string{"recovery-peer-B", "recovery-peer-C"} {
		v := s.current()
		joined := s.result([]string{"producer", "request", s.id}, viewer.TimelineJoinRequestInput{SchemaVersion: 2, OperationID: s.op("request"), ExpectedSessionVersion: v.Version, Source: viewer.TimelineSource{ProjectRoot: repo, Host: "claude-code", NativeSessionID: native}, Reason: "Explicit pending fixture participant"})
		requests = append(requests, joined.RequestID)
	}
	before := s.current()
	info := f.info()
	type outcome struct {
		status  int
		body    string
		request viewer.TimelineRecoverLeadRequest
	}
	results := make(chan outcome, 2)
	var group sync.WaitGroup
	for i, requestID := range requests {
		group.Add(1)
		go func(index int, requestID string) {
			defer group.Done()
			value := viewer.TimelineRecoverLeadRequest{TimelineNativeOperation: viewer.TimelineNativeOperation{SchemaVersion: 2, OperationID: fmt.Sprintf("local-recover-%d", index), Actor: viewer.TimelineActor{Kind: "local_viewer"}, ExpectedSessionVersion: before.Version, Reason: "Explicit isolated browser recovery"}, RequestID: requestID}
			raw, _ := json.Marshal(value)
			request, _ := http.NewRequest("POST", info.url()+"/api/timeline/sessions/"+s.id+"/recover-lead", bytes.NewReader(raw))
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("X-Hive-Token", info.Token)
			request.Header.Set("Origin", info.url())
			request.Header.Set("X-Hive-Action-Scope", "session")
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				results <- outcome{status: 0}
				return
			}
			defer response.Body.Close()
			body, _ := io.ReadAll(io.LimitReader(response.Body, viewer.TimelineMaxBody))
			results <- outcome{status: response.StatusCode, body: string(body), request: value}
		}(i, requestID)
	}
	group.Wait()
	close(results)
	winners, conflicts := 0, 0
	var winner outcome
	for result := range results {
		switch result.status {
		case 200:
			winners++
			winner = result
		case 409:
			conflicts++
		default:
			t.Fatal("unexpected recovery status", result.status)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatal("local recovery did not have exactly one winner")
	}
	after := s.current()
	if after.Source != before.Source || after.State != before.State || after.CurrentStageEventID == nil || *after.CurrentStageEventID != stage || len(after.Producers) != 2 || len(after.PendingJoinRequests) != 1 {
		t.Fatal("local recovery changed declared state or accepted both requests")
	}
	f.mustView("stop")
	f.mustView("start")
	replay := timelineLocalJSON(t, f, "POST", "/api/timeline/sessions/"+s.id+"/recover-lead", winner.request)
	var originalResult, replayedResult viewer.TimelineOperationResult
	if json.Unmarshal([]byte(winner.body), &originalResult) != nil || json.Unmarshal([]byte(replay), &replayedResult) != nil || !reflect.DeepEqual(originalResult, replayedResult) {
		t.Fatal("local recovery receipt changed after restart")
	}
	if !reflect.DeepEqual(after, s.current()) {
		t.Fatal("recovery replay changed current state")
	}
	f.mustView("stop")
}
