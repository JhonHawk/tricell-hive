package viewer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestTimelineDomainJoinRequestDoesNotGrantAuthority(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	request := TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "join", ExpectedSessionVersion: v.Version, Source: TimelineSource{ProjectRoot: v.Source.ProjectRoot, Host: "claude-code", NativeSessionID: "collaborator"}, Reason: "Explicit shared work"}
	result, err := s.TimelineV3Operation(v.ViewerSessionID, "join_request", "", &request)
	if err != nil {
		t.Fatal(err)
	}
	current, _ := s.TimelineGet(v.ViewerSessionID)
	if len(current.Producers) != 1 || len(current.PendingJoinRequests) != 1 || result.RequestID == "" {
		t.Fatal("join granted membership")
	}
	accept := TimelineProducerAcceptRequest{TimelineNativeOperation: TimelineNativeOperation{SchemaVersion: 3, OperationID: "accept", ExpectedSessionVersion: current.Version, Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, Reason: "Accept contributor"}, RequestID: result.RequestID}
	joined, err := s.TimelineV3Operation(v.ViewerSessionID, "producer_accept", "", &accept)
	if err != nil {
		t.Fatal(err)
	}
	current, _ = s.TimelineGet(v.ViewerSessionID)
	if len(current.Producers) != 2 || len(current.PendingJoinRequests) != 0 || joined.ProducerID == "" || current.Source != v.Source {
		t.Fatal("membership corrupted")
	}
}

func domainCurrent(t *testing.T, s *Store, id string) TimelineSession {
	t.Helper()
	v, err := s.TimelineGet(id)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func domainBase(t *testing.T, s *Store, id, op, producer string) TimelineNativeOperation {
	v := domainCurrent(t, s, id)
	return TimelineNativeOperation{SchemaVersion: 3, OperationID: op, ExpectedSessionVersion: v.Version, Actor: TimelineActor{Kind: "native_producer", ProducerID: producer}, Reason: "Explicit " + op}
}
func domainStage(t *testing.T, s *Store, id string, stage TimelineStage) string {
	t.Helper()
	v := domainCurrent(t, s, id)
	result, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "start-" + string(stage), Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, Reason: "Start " + string(stage), Kind: "stage", Stage: stage, StageState: "in_progress", ExpectedSessionVersion: &v.Version, Title: string(stage), Summary: "Work stage", TicketIDs: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	return result.EventID
}
func domainJoin(t *testing.T, s *Store, id, native string) string {
	t.Helper()
	v := domainCurrent(t, s, id)
	r := TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "join-" + native, ExpectedSessionVersion: v.Version, Source: TimelineSource{ProjectRoot: v.Source.ProjectRoot, Host: "codex", NativeSessionID: native}, Reason: "Join explicitly"}
	res, err := s.TimelineV3Operation(id, "join_request", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	return res.RequestID
}
func domainAccept(t *testing.T, s *Store, id, request string) string {
	t.Helper()
	v := domainCurrent(t, s, id)
	r := TimelineProducerAcceptRequest{TimelineNativeOperation: domainBase(t, s, id, "accept-"+request, v.ForegroundProducerID), RequestID: request}
	res, err := s.TimelineV3Operation(id, "producer_accept", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	return res.ProducerID
}
func domainTask(t *testing.T, s *Store, id string) TimelineTask {
	t.Helper()
	v := domainCurrent(t, s, id)
	r := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "task", v.ForegroundProducerID), TaskKey: "T7", Title: "Verify and deliver", Summary: "One task across stages", Segments: []TimelineTaskSegmentInput{{PlannedStage: "verify", Order: 1}, {PlannedStage: "delivery", Order: 1}}}
	res, err := s.TimelineV3Operation(id, "task_create", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	v = domainCurrent(t, s, id)
	for _, task := range v.Tasks {
		if task.TaskID == res.TaskID {
			return task
		}
	}
	t.Fatal("task absent")
	return TimelineTask{}
}
func domainBind(t *testing.T, s *Store, id, task, segment, stage, op string) {
	t.Helper()
	v := domainCurrent(t, s, id)
	current, _ := timelineFindTask(v, task)
	r := TimelineTaskPatchRequest{TimelineNativeOperation: domainBase(t, s, id, op, v.ForegroundProducerID), ExpectedTaskVersion: current.TaskVersion, SegmentID: segment, DisplayStageEventID: stage, DecisionTarget: nil}
	_, err := s.TimelineV3Operation(id, "task_patch", task, &r)
	if err != nil {
		t.Fatal(err)
	}
}
func domainTransition(t *testing.T, s *Store, id, task, segment, state, op string) *TimelineOperationResult {
	t.Helper()
	v := domainCurrent(t, s, id)
	current, _ := timelineFindTask(v, task)
	r := TimelineTaskTransitionRequest{TimelineNativeOperation: domainBase(t, s, id, op, v.ForegroundProducerID), ExpectedTaskVersion: current.TaskVersion, SegmentID: segment, State: state}
	res, err := s.TimelineV3Operation(id, "task_transition", task, &r)
	if err != nil {
		t.Fatal(err)
	}
	return &res
}

func TestTimelineDomainTaskSegmentsAndLinkedInvalidation(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "verify")
	task := domainTask(t, s, id)
	domainBind(t, s, id, task.TaskID, task.Segments[0].SegmentID, stage, "bind-verify")
	domainTransition(t, s, id, task.TaskID, task.Segments[0].SegmentID, "in_progress", "doing")
	domainTransition(t, s, id, task.TaskID, task.Segments[0].SegmentID, "completed", "verified")
	current := domainCurrent(t, s, id)
	got, _ := timelineFindTask(current, task.TaskID)
	if got.State == "completed" || got.Segments[1].State != "not_started" || current.CurrentStageEventID == nil || *current.CurrentStageEventID != stage {
		t.Fatal("segment transition changed unrelated state")
	}
	before := got.TaskVersion
	event := TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "linked", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, Kind: "review", ReviewStatus: "completed", Title: "Verification evidence", Summary: "A precise result", TicketIDs: []string{}, TargetStageEventID: stage, TaskID: task.TaskID, SegmentID: task.Segments[0].SegmentID}
	result, err := s.TimelineAppend(id, event)
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	got, _ = timelineFindTask(current, task.TaskID)
	if got.TaskVersion != before+1 {
		t.Fatal("linked event did not invalidate task")
	}
	replay, err := s.TimelineAppend(id, event)
	if err != nil || !reflect.DeepEqual(result, replay) {
		t.Fatal("replay changed result", err)
	}
	current = domainCurrent(t, s, id)
	got, _ = timelineFindTask(current, task.TaskID)
	if got.TaskVersion != before+1 {
		t.Fatal("replay changed task")
	}
	r := TimelineTaskTransitionRequest{TimelineNativeOperation: domainBase(t, s, id, "incomplete-block", v.ForegroundProducerID), ExpectedTaskVersion: got.TaskVersion, SegmentID: task.Segments[1].SegmentID, State: "blocked"}
	_, err = s.TimelineV3Operation(id, "task_transition", task.TaskID, &r)
	if err == nil {
		t.Fatal("incomplete block accepted")
	}
}

func TestTimelineDomainRecoveryConcurrencyAndProvenance(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	pending := domainJoin(t, s, id, "next-lead")
	before := domainCurrent(t, s, id)
	request := TimelineRecoverLeadRequest{TimelineNativeOperation: TimelineNativeOperation{SchemaVersion: 3, OperationID: "recover", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedSessionVersion: before.Version, Reason: "Explicit local recovery"}, RequestID: pending}
	var wg sync.WaitGroup
	results := make(chan TimelineOperationResult, 2)
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r := request
			r.OperationID = fmt.Sprintf("recover-%d", n)
			result, err := s.TimelineV3Operation(id, "recover_lead", "", &r)
			if err != nil {
				failures <- err
			} else {
				results <- result
			}
		}(i)
	}
	wg.Wait()
	close(results)
	close(failures)
	if len(results) != 1 || len(failures) != 1 {
		t.Fatal("recovery had multiple winners")
	}
	winner := <-results
	current := domainCurrent(t, s, id)
	if current.Source != v.Source || current.State != v.State || current.ForegroundProducerID != winner.ProducerID || len(current.Producers) != 2 {
		t.Fatal("recovery altered content/provenance")
	}
	request.OperationID = winner.OperationID
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := reopened.TimelineV3Operation(id, "recover_lead", "", &request)
	if err != nil || !reflect.DeepEqual(replay, winner) {
		t.Fatal("recovery receipt lost", err)
	}
	oldLead := TimelineForegroundTransferRequest{TimelineNativeOperation: domainBase(t, reopened, id, "former-lead", v.ForegroundProducerID), TargetProducerID: v.ForegroundProducerID}
	_, err = reopened.TimelineV3Operation(id, "foreground_transfer", "", &oldLead)
	requireTimelineCode(t, err, "lead_required")
}

func TestTimelineDomainCIApplicableAttemptAndTaskSnapshots(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "delivery")
	task := domainTask(t, s, id)
	segment := task.Segments[1].SegmentID
	domainBind(t, s, id, task.TaskID, segment, stage, "bind-delivery")
	policy := TimelineCheckPolicy{PolicyID: "required-v1", Confirmed: true, Authority: "Branch protection", SourceURL: "https://example.com/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []TimelineRequiredCheck{{CheckKey: "unit", Label: "Unit tests"}}}
	create := TimelineWaitCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "wait", v.ForegroundProducerID), TargetStageEventID: stage, TaskID: task.TaskID, SegmentID: segment, CandidateID: "candidate-test", PullRequestDeliveryID: "delivery-test", Policy: policy}
	result, err := s.TimelineV3Operation(id, "wait_create", "", &create)
	if err != nil {
		t.Fatal(err)
	}
	waitID := result.WaitID
	attempt := func(op, run string) TimelineOperationResult {
		current := domainCurrent(t, s, id)
		w, _ := timelineFindWait(current, waitID)
		request := TimelineWaitAttemptStartRequest{TimelineNativeOperation: domainBase(t, s, id, op, v.ForegroundProducerID), ExpectedWaitVersion: w.WaitVersion, CheckKeys: []string{"unit"}, ExternalRun: TimelineCIExternalRun{Commit: "abc", RunID: run, URL: "https://example.com/runs/" + run}, SourceURL: "https://example.com/runs/" + run, ObservedAt: "2026-01-01T00:00:00Z"}
		r, err := s.TimelineV3Operation(id, "wait_attempt_start", waitID, &request)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	observe := func(op, attempt, run, state string) TimelineOperationResult {
		observedAt := "2026-01-02T00:00:00Z"
		if op == "late-first-success" {
			observedAt = "2026-01-03T00:00:00Z"
		}
		current := domainCurrent(t, s, id)
		w, _ := timelineFindWait(current, waitID)
		request := TimelineWaitObserveRequest{TimelineNativeOperation: domainBase(t, s, id, op, v.ForegroundProducerID), ExpectedWaitVersion: w.WaitVersion, AttemptID: attempt, CheckKey: "unit", ExternalRun: TimelineCIExternalRun{Commit: "abc", RunID: run, URL: "https://example.com/runs/" + run}, State: state, Summary: "Observed provider result", ObservedAt: observedAt}
		r, err := s.TimelineV3Operation(id, "wait_observe", waitID, &request)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	first := attempt("attempt-one", "one")
	observe("first-success", first.AttemptID, "one", "succeeded")
	current := domainCurrent(t, s, id)
	w, _ := timelineFindWait(current, waitID)
	if w.State != "succeeded" {
		t.Fatal(w.State)
	}
	second := attempt("attempt-two", "two")
	observe("second-waiting", second.AttemptID, "two", "waiting")
	observe("late-first-success", first.AttemptID, "one", "succeeded")
	current = domainCurrent(t, s, id)
	w, _ = timelineFindWait(current, waitID)
	if w.State != "waiting" || len(w.Attempts) != 2 || len(w.Observations) != 3 {
		t.Fatal("old attempt replaced current outcome")
	}
	before, _ := timelineFindTask(current, task.TaskID)
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, reopened, id)
	w, _ = timelineFindWait(current, waitID)
	after, _ := timelineFindTask(current, task.TaskID)
	if w.State != "waiting" || after.TaskVersion != before.TaskVersion {
		t.Fatal("restart lost attempt authority or task version")
	}
	// Unknown attempts cannot promote themselves, even with the current versions.
	request := TimelineWaitObserveRequest{TimelineNativeOperation: domainBase(t, reopened, id, "unknown-attempt", v.ForegroundProducerID), ExpectedWaitVersion: w.WaitVersion, AttemptID: "unknown", CheckKey: "unit", ExternalRun: TimelineCIExternalRun{Commit: "abc", RunID: "unknown", URL: "https://example.com/runs/unknown"}, State: "succeeded", Summary: "Untrusted attempt mapping", ObservedAt: "2026-01-03T00:00:00Z"}
	_, err = reopened.TimelineV3Operation(id, "wait_observe", waitID, &request)
	if err == nil {
		t.Fatal("unknown attempt promoted")
	}
}

func TestTimelineDomainLifecycleKeepsValidationForegroundAndAuditImmutable(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "build")
	current := domainCurrent(t, s, id)
	wait := TimelineAwaitingValidation
	result, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "ask", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedVersion: current.Version, State: &wait, Reason: "Waiting for concrete user validation"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	if current.State != wait || current.CurrentStageEventID == nil || *current.CurrentStageEventID != stage || current.Events[0].StageState != "awaiting_validation" {
		t.Fatal("validation lost stage")
	}
	_, err = s.TimelinePatchEvent(id, current.Events[len(current.Events)-1].EventID, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "edit-audit", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedVersion: 1, Title: timelineString("Changed audit")})
	requireTimelineCode(t, err, "immutable_audit")
	active := TimelineActive
	_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "resume", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedVersion: result.SessionVersion, State: &active, Reason: "User responded; resume explicitly"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	state := TimelineStageState("completed")
	_, err = s.TimelinePatchEvent(id, stage, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "finish-build", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedVersion: current.Events[0].Version, ExpectedSessionVersion: &current.Version, StageState: &state, Reason: "Build completed"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	if current.CurrentStageEventID != nil || current.State != TimelineActive {
		t.Fatal("completed stage left foreground")
	}
	_, err = s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "close", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, Kind: "closure", Reason: "Explicit completion", Title: "Finished", Summary: "Integrated outcome", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Completed"}})
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if current = domainCurrent(t, reopened, id); current.State != TimelineClosed {
		t.Fatal("closure lost")
	}
}

func TestTimelineDomainHTTPGuardsAndClosedOperationDispatch(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	server, err := NewServer(ServerOptions{Store: s, Port: 45678, Token: testToken})
	if err != nil {
		t.Fatal(err)
	}
	input := TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "http-join", ExpectedSessionVersion: v.Version, Source: TimelineSource{ProjectRoot: v.Source.ProjectRoot, Host: "codex", NativeSessionID: "http-collaborator"}, Reason: "Explicit HTTP join"}
	raw, _ := json.Marshal(input)
	path := "/api/timeline/sessions/" + v.ViewerSessionID + "/join-requests"
	call := func(token, origin, scope string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "http://127.0.0.1:45678"+path, bytes.NewReader(raw))
		r.Header.Set(tokenHeader, token)
		r.Header.Set("Origin", origin)
		r.Header.Set("X-Hive-Action-Scope", scope)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	for _, guard := range []struct{ token, origin, scope string }{{"", "http://127.0.0.1:45678", "session"}, {testToken, "http://foreign.example", "session"}, {testToken, "http://127.0.0.1:45678", ""}} {
		if w := call(guard.token, guard.origin, guard.scope); w.Code != 403 && w.Code != 409 {
			t.Fatal("guard bypass", w.Code)
		}
	}
	w := call(testToken, "http://127.0.0.1:45678", "session")
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	var result TimelineOperationResult
	if err = json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.RequestID == "" {
		t.Fatal("bad domain response", err)
	}
	// A Go caller cannot disguise a transfer as a less privileged kind.
	transfer := TimelineForegroundTransferRequest{TimelineNativeOperation: domainBase(t, s, v.ViewerSessionID, "wrong-kind", v.ForegroundProducerID), TargetProducerID: v.ForegroundProducerID}
	_, err = s.TimelineV3Operation(v.ViewerSessionID, "task_transition", "", &transfer)
	if err == nil {
		t.Fatal("mixed dispatch accepted")
	}
}

func TestTimelineDomainCollaboratorTargetsOwnershipAndForeignProjects(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "build")
	member := domainAccept(t, s, id, domainJoin(t, s, id, "reviewer"))
	before := domainCurrent(t, s, id)
	publication := TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "review", Actor: TimelineActor{Kind: "native_producer", ProducerID: member}, Kind: "review", ReviewStatus: "in_progress", TargetStageEventID: stage, Title: "Parallel review", Summary: "Independent review", TicketIDs: []string{}}
	result, err := s.TimelineAppend(id, publication)
	if err != nil {
		t.Fatal(err)
	}
	current := domainCurrent(t, s, id)
	if current.State != before.State || *current.CurrentStageEventID != stage {
		t.Fatal("collaborator stole foreground")
	}
	publication.OperationID = "no-target"
	publication.TargetStageEventID = ""
	_, err = s.TimelineAppend(id, publication)
	if err == nil {
		t.Fatal("collaborator used implicit stage")
	}
	_, err = s.TimelinePatchEvent(id, result.EventID, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "foreign-edit", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedVersion: result.EventVersion, Title: timelineString("Edit another producer")})
	requireTimelineCode(t, err, "content_owner_required")
	foreign := newRepo(t)
	if _, err = RegisterRepository(filepath.Dir(s.Dir()), foreign); err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	join := TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "foreign-join", ExpectedSessionVersion: current.Version, Source: TimelineSource{ProjectRoot: foreign, Host: "codex", NativeSessionID: "foreign"}, Reason: "Foreign project must fail"}
	before = current
	_, err = s.TimelineV3Operation(id, "join_request", "", &join)
	requireTimelineCode(t, err, "foreign_project")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("denied join mutated work")
	}
}

func domainEvent(t *testing.T, v TimelineSession, id string) TimelineEvent {
	t.Helper()
	for _, event := range v.Events {
		if event.EventID == id {
			return event
		}
	}
	t.Fatal("missing event", id)
	return TimelineEvent{}
}
func TestTimelineDomainTaskReentryPreservesHistoricalEvidenceAndDocuments(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "verify")
	task := domainTask(t, s, id)
	segment := task.Segments[0].SegmentID
	domainBind(t, s, id, task.TaskID, segment, stage, "bind-proof")
	domainTransition(t, s, id, task.TaskID, segment, "in_progress", "work-proof")
	current := domainCurrent(t, s, id)
	before, _ := timelineFindTask(current, task.TaskID)
	metadata := TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "proof-document", Actor: timelineTestActor(s, id), MediaType: "text/markdown"}
	document, err := s.TimelineAddDocument(id, metadata, []byte("# Proof"))
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	after, _ := timelineFindTask(current, task.TaskID)
	if after.TaskVersion != before.TaskVersion {
		t.Fatal("unlinked upload invalidated task")
	}
	snapshot := TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: "proof-snapshot", Actor: timelineTestActor(s, id), ExpectedVersion: domainEvent(t, current, stage).Version, ExpectedSessionVersion: current.Version, DocumentKey: "proof", TaskID: task.TaskID, SegmentID: segment, Status: "complete", Tabs: []TimelineDocumentTab{{Label: "Proof", DocumentID: document.DocumentID}}}
	revision, err := s.TimelineSnapshot(id, stage, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	domainTransition(t, s, id, task.TaskID, segment, "completed", "proof-completed")
	current = domainCurrent(t, s, id)
	completed := TimelineStageState("completed")
	_, err = s.TimelinePatchEvent(id, stage, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "verify-done", Actor: timelineTestActor(s, id), ExpectedVersion: domainEvent(t, current, stage).Version, ExpectedSessionVersion: &current.Version, StageState: &completed, Reason: "Verification complete"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	restartStage, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "verify-again", Actor: timelineTestActor(s, id), Kind: "stage", Stage: "verify", StageState: "in_progress", ExpectedSessionVersion: &current.Version, RelatedEventID: stage, Reason: "Explicit correction", Title: "Verify correction", Summary: "Recheck changed result", TicketIDs: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	domainBind(t, s, id, task.TaskID, segment, restartStage.EventID, "rebind-correction")
	domainTransition(t, s, id, task.TaskID, segment, "in_progress", "reopen-proof")
	current = domainCurrent(t, s, id)
	before, _ = timelineFindTask(current, task.TaskID)
	snapshot.OperationID = "continued-proof"
	snapshot.ExpectedVersion = domainEvent(t, current, restartStage.EventID).Version
	snapshot.ExpectedSessionVersion = current.Version
	snapshot.PredecessorRevisionID = revision.DocumentRevisionID
	latest, err := s.TimelineSnapshot(id, restartStage.EventID, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	after, _ = timelineFindTask(current, task.TaskID)
	if after.TaskVersion != before.TaskVersion+1 {
		t.Fatal("linked revision invalidated task incorrectly")
	}
	snapshot.OperationID = "identical-proof"
	snapshot.ExpectedVersion = latest.EventVersion
	snapshot.ExpectedSessionVersion = current.Version
	snapshot.PredecessorRevisionID = latest.DocumentRevisionID
	noChange, err := s.TimelineSnapshot(id, restartStage.EventID, snapshot)
	if err != nil || noChange.DocumentRevisionID != latest.DocumentRevisionID {
		t.Fatal("identical snapshot created revision", err)
	}
	current = domainCurrent(t, s, id)
	unchanged, _ := timelineFindTask(current, task.TaskID)
	if unchanged.TaskVersion != after.TaskVersion {
		t.Fatal("identical snapshot invalidated task")
	}
	current = domainCurrent(t, s, id)
	retained, _ := timelineFindTask(current, task.TaskID)
	withdrawn := true
	retire := TimelineTaskPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "retire-task", v.ForegroundProducerID), ExpectedTaskVersion: retained.TaskVersion, Withdrawn: &withdrawn, DecisionTarget: &TimelineDecisionTarget{Kind: "document", EventID: restartStage.EventID, DocumentRevisionID: latest.DocumentRevisionID}}
	if _, err = s.TimelineV3Operation(id, "task_patch", task.TaskID, &retire); err != nil {
		t.Fatal("retirement broke historical references", err)
	}
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, reopened, id)
	if domainEvent(t, current, stage).DocumentSnapshots[0].DocumentRevisionID != revision.DocumentRevisionID || domainEvent(t, current, restartStage.EventID).DocumentSnapshots[0].PredecessorRevisionID != revision.DocumentRevisionID {
		t.Fatal("reentry rewrote earlier evidence")
	}
}

func TestTimelineDomainTaskBlockResolutionAndNoop(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "verify")
	task := domainTask(t, s, id)
	segment := task.Segments[0].SegmentID
	domainBind(t, s, id, task.TaskID, segment, stage, "bind-block")
	current := domainCurrent(t, s, id)
	task, _ = timelineFindTask(current, task.TaskID)
	block := &TimelineBlock{Reason: "Required fixture unavailable", Steps: []string{"Restore the fixture in the isolated test directory"}, Owner: "Current lead", ResumeWhen: "Fixture validation succeeds"}
	request := TimelineTaskTransitionRequest{TimelineNativeOperation: domainBase(t, s, id, "block", v.ForegroundProducerID), ExpectedTaskVersion: task.TaskVersion, SegmentID: segment, State: "blocked", Block: block}
	blocked, err := s.TimelineV3Operation(id, "task_transition", task.TaskID, &request)
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	foreground := *current.CurrentStageEventID
	request = TimelineTaskTransitionRequest{TimelineNativeOperation: domainBase(t, s, id, "resolve", v.ForegroundProducerID), ExpectedTaskVersion: blocked.TaskVersion, SegmentID: segment, State: "in_progress", Resolution: "Fixture restored and validated"}
	resolved, err := s.TimelineV3Operation(id, "task_transition", task.TaskID, &request)
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	if *current.CurrentStageEventID != foreground {
		t.Fatal("task resolution changed foreground")
	}
	last := current.Events[len(current.Events)-1]
	if last.Audit == nil || last.Audit.Block == nil || last.Audit.Block.Steps[0] != block.Steps[0] || last.Audit.Resolution != request.Resolution {
		t.Fatal("resolution lost unblock history")
	}
	eventCount := len(current.Events)
	request.TimelineNativeOperation = domainBase(t, s, id, "noop-state", v.ForegroundProducerID)
	request.ExpectedTaskVersion = resolved.TaskVersion
	request.Resolution = ""
	result, err := s.TimelineV3Operation(id, "task_transition", task.TaskID, &request)
	if err != nil || result.TaskVersion != resolved.TaskVersion {
		t.Fatal("no-op advanced task", err)
	}
	current = domainCurrent(t, s, id)
	if len(current.Events) != eventCount {
		t.Fatal("no-op manufactured transition")
	}
	currentTask, _ := timelineFindTask(current, task.TaskID)
	sameOrder := currentTask.Segments[0].Order
	same := TimelineTaskPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "same-assignment", v.ForegroundProducerID), ExpectedTaskVersion: currentTask.TaskVersion, SegmentID: segment, AssignedProducerID: currentTask.Segments[0].AssignedProducerID, Order: &sameOrder}
	sameResult, err := s.TimelineV3Operation(id, "task_patch", task.TaskID, &same)
	if err != nil || sameResult.TaskVersion != currentTask.TaskVersion {
		t.Fatal("no-op assignment required a new scope decision", err)
	}

}

func TestTimelineDomainTransferConcurrentWinnersAndChangedRetry(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	member := domainAccept(t, s, id, domainJoin(t, s, id, "transfer-target"))
	current := domainCurrent(t, s, id)
	base := TimelineForegroundTransferRequest{TimelineNativeOperation: domainBase(t, s, id, "transfer", v.ForegroundProducerID), TargetProducerID: member}
	var wg sync.WaitGroup
	success := make(chan TimelineOperationResult, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			r := base
			r.OperationID = fmt.Sprintf("transfer-%d", n)
			if result, err := s.TimelineV3Operation(id, "foreground_transfer", "", &r); err == nil {
				success <- result
			}
		}(i)
	}
	wg.Wait()
	close(success)
	if len(success) != 1 {
		t.Fatal("competing transfers had wrong winner count")
	}
	winner := <-success
	after := domainCurrent(t, s, id)
	if after.Version != current.Version+1 || after.ForegroundProducerID != member || after.Source != v.Source {
		t.Fatal("transfer did not commit once")
	}
	base.OperationID = winner.OperationID
	base.Reason = "Changed retry payload"
	_, err := s.TimelineV3Operation(id, "foreground_transfer", "", &base)
	requireTimelineCode(t, err, "operation_conflict")
}

func TestTimelineDomainContentEventCapacityReservesClosure(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	record := timelineClone(s.timeline[id])
	actor := timelineTestActor(s, id)
	now := s.timelineTime()
	for i := 0; i < TimelineContentEvents; i++ {
		record.Session.Events = append(record.Session.Events, TimelineEvent{EventID: fmt.Sprintf("event-%d", i), Sequence: int64(i + 1), Version: 1, Kind: "request", Actor: actor, Title: "Request", Summary: "Explicit activity", CreatedAt: now, UpdatedAt: now, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}})
	}
	if err := s.timelineSave(record); err != nil {
		t.Fatal(err)
	}
	_, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "content-overflow", Actor: actor, Kind: "request", Title: "Extra", Summary: "Must fail", TicketIDs: []string{}})
	if err == nil {
		t.Fatal("content used lifecycle reserve")
	}
	result, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "capacity-closure", Actor: actor, Kind: "closure", Reason: "Explicit terminal closure", Title: "Finished", Summary: "Retain all history", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Finished"}})
	if err != nil {
		t.Fatal(err)
	}
	current := domainCurrent(t, s, id)
	if len(current.Events) != TimelineContentEvents+2 || current.State != TimelineClosed {
		t.Fatal("closure dropped history")
	}
	if _, err = s.TimelineDelete(id, TimelineMutationRequest{SchemaVersion: 3, OperationID: "capacity-delete", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: result.SessionVersion}); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineDomainLocalManagementAndLongReasons(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	title := "  Renamed work  "
	archived, pinned := true, true
	request := TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "local-management", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Title: &title, Archived: &archived, Pinned: &pinned, Reason: strings.Repeat("x", 500)}
	result, err := s.TimelinePatch(id, request)
	if err != nil {
		t.Fatal(err)
	}
	current := domainCurrent(t, s, id)
	if current.Title != "Renamed work" || !current.Archived || !current.Pinned || current.Source != v.Source || current.State != v.State || len(current.Events) != 3 {
		t.Fatal("local management changed domain state")
	}
	for _, event := range current.Events {
		if event.Actor.Kind != "local_viewer" || event.Actor.ProducerID != "" || event.Audit == nil || len(event.Audit.Reason) != 500 {
			t.Fatal("local actor or long reason lost")
		}
	}
	replay, err := s.TimelinePatch(id, request)
	if err != nil || !reflect.DeepEqual(replay, result) || len(domainCurrent(t, s, id).Events) != 3 {
		t.Fatal("local replay duplicated audit", err)
	}
	request.OperationID = "native-spoof"
	request.ExpectedVersion = current.Version
	request.Actor = TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}
	_, err = s.TimelinePatch(id, request)
	requireTimelineCode(t, err, "local_viewer_required")
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if current = domainCurrent(t, reopened, id); current.Title != "Renamed work" || !current.Pinned || !current.Archived {
		t.Fatal("preferences lost")
	}
}

func TestTimelineDomainWaitUnknownPolicyAndNewCandidateLink(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	build := domainStage(t, s, id, "build")
	create := TimelineWaitCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "wrong-stage", v.ForegroundProducerID), TargetStageEventID: build, CandidateID: "candidate-test", PullRequestDeliveryID: "delivery-test", Policy: TimelineCheckPolicy{PolicyID: "unconfirmed", Confirmed: false, Authority: "Unknown requirements", SourceURL: "https://example.com/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []TimelineRequiredCheck{}}}
	_, err := s.TimelineV3Operation(id, "wait_create", "", &create)
	requireTimelineCode(t, err, "delivery_required")
	delivery := domainStage(t, s, id, "delivery")
	create.TimelineNativeOperation = domainBase(t, s, id, "unknown-policy", v.ForegroundProducerID)
	create.TargetStageEventID = delivery
	result, err := s.TimelineV3Operation(id, "wait_create", "", &create)
	if err != nil {
		t.Fatal(err)
	}
	current := domainCurrent(t, s, id)
	wait, _ := timelineFindWait(current, result.WaitID)
	if wait.State != "unknown" {
		t.Fatal("unconfirmed requirements became success")
	}
	create.TimelineNativeOperation = domainBase(t, s, id, "new-candidate", v.ForegroundProducerID)
	create.CandidateID = "candidate-def"
	_, err = s.TimelineV3Operation(id, "wait_create", "", &create)
	requireTimelineCode(t, err, "previous_wait_required")
	create.PreviousWaitID = result.WaitID
	next, err := s.TimelineV3Operation(id, "wait_create", "", &create)
	if err != nil || next.WaitID == result.WaitID {
		t.Fatal("new candidate reused old wait", err)
	}
	current = domainCurrent(t, s, id)
	wait, _ = timelineFindWait(current, next.WaitID)
	if wait.PreviousWaitID != result.WaitID || wait.State != "unknown" || len(wait.Observations) != 0 {
		t.Fatal("new candidate inherited old outcome")
	}
}

func TestTimelineDomainFailedWorkRequiresExplicitReopen(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "build")
	current := domainCurrent(t, s, id)
	pausedWork := TimelinePaused
	_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "pause-before-failure", Actor: timelineTestActor(s, id), ExpectedVersion: current.Version, State: &pausedWork, Reason: "Pause before declaring failed work"})
	if err != nil {
		t.Fatal(err)
	}
	if paused := domainCurrent(t, s, id); paused.CurrentStageEventID != nil || domainEvent(t, paused, stage).StageState != "paused" {
		t.Fatal("global pause did not synchronize foreground")
	}
	current = domainCurrent(t, s, id)
	failed := TimelineFailed
	_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "work-failed", Actor: timelineTestActor(s, id), ExpectedVersion: current.Version, State: &failed, Reason: "Explicit terminal failure"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	resumed := TimelineStageState("in_progress")
	request := TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "implicit-reopen", Actor: timelineTestActor(s, id), ExpectedVersion: domainEvent(t, current, stage).Version, ExpectedSessionVersion: &current.Version, StageState: &resumed, Reason: "Cannot resume without reopening work"}
	_, err = s.TimelinePatchEvent(id, stage, request)
	requireTimelineCode(t, err, "session_terminal")
	active := TimelineActive
	_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "explicit-work-reopen", Actor: timelineTestActor(s, id), ExpectedVersion: current.Version, State: &active, Reason: "Explicit authorized continuation"})
	if err != nil {
		t.Fatal(err)
	}
	current = domainCurrent(t, s, id)
	request.OperationID = "resume-after-reopen"
	request.ExpectedSessionVersion = &current.Version
	if _, err = s.TimelinePatchEvent(id, stage, request); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineDomainKnownRunCannotPromoteViaURLChange(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "delivery")
	create := TimelineWaitCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "identity-wait", v.ForegroundProducerID), TargetStageEventID: stage, CandidateID: "candidate-test", PullRequestDeliveryID: "delivery-test", Policy: TimelineCheckPolicy{PolicyID: "required", Confirmed: true, Authority: "Provider policy", SourceURL: "https://example.com/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []TimelineRequiredCheck{{CheckKey: "unit", Label: "Unit"}}}}
	made, err := s.TimelineV3Operation(id, "wait_create", "", &create)
	if err != nil {
		t.Fatal(err)
	}
	waitID := made.WaitID
	for i, state := range []string{"succeeded", "failed"} {
		current := domainCurrent(t, s, id)
		wait, _ := timelineFindWait(current, waitID)
		run := fmt.Sprintf("run-%d", i+1)
		external := TimelineCIExternalRun{Commit: "abc", RunID: run, URL: "https://example.com/runs/" + run}
		start := TimelineWaitAttemptStartRequest{TimelineNativeOperation: domainBase(t, s, id, "start-"+run, v.ForegroundProducerID), ExpectedWaitVersion: wait.WaitVersion, CheckKeys: []string{"unit"}, ExternalRun: external, SourceURL: external.URL, ObservedAt: "2026-01-01T00:00:00Z"}
		attempt, err := s.TimelineV3Operation(id, "wait_attempt_start", waitID, &start)
		if err != nil {
			t.Fatal(err)
		}
		observation := TimelineWaitObserveRequest{TimelineNativeOperation: domainBase(t, s, id, "observe-"+run, v.ForegroundProducerID), ExpectedWaitVersion: attempt.WaitVersion, AttemptID: attempt.AttemptID, CheckKey: "unit", ExternalRun: external, State: state, Summary: "Observed current provider run", ObservedAt: "2026-01-02T00:00:00Z"}
		if _, err = s.TimelineV3Operation(id, "wait_observe", waitID, &observation); err != nil {
			t.Fatal(err)
		}
	}
	before := domainCurrent(t, s, id)
	wait, _ := timelineFindWait(before, waitID)
	if wait.State != "failed" {
		t.Fatal(wait.State)
	}
	old := wait.Attempts[0].ExternalRun
	old.URL += "?view=checks"
	promotion := TimelineWaitAttemptStartRequest{TimelineNativeOperation: domainBase(t, s, id, "url-repromotion", v.ForegroundProducerID), ExpectedWaitVersion: wait.WaitVersion, CheckKeys: []string{"unit"}, ExternalRun: old, SourceURL: old.URL, ObservedAt: "2026-01-03T00:00:00Z"}
	_, err = s.TimelineV3Operation(id, "wait_attempt_start", waitID, &promotion)
	requireTimelineCode(t, err, "attempt_already_known")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("URL-only replay changed applicable attempt")
	}
}

func TestTimelineDomainLifecycleCannotConsumeClosureReserve(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	record := timelineClone(s.timeline[id])
	actor := timelineTestActor(s, id)
	now := s.timelineTime()
	for i := 0; i < TimelineContentEvents; i++ {
		record.Session.Events = append(record.Session.Events, TimelineEvent{EventID: fmt.Sprintf("reserve-%d", i), Sequence: int64(i + 1), Version: 1, Kind: "request", Actor: actor, Title: "Request", Summary: "Content at capacity", CreatedAt: now, UpdatedAt: now, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}})
	}
	if err := s.timelineSave(record); err != nil {
		t.Fatal(err)
	}
	stage := domainStage(t, s, id, "build")
	rejected := false
	for cycle := 0; cycle < 3; cycle++ {
		before := domainCurrent(t, s, id)
		wait := TimelineAwaitingValidation
		_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: fmt.Sprintf("reserve-validation-%d", cycle), Actor: actor, ExpectedVersion: before.Version, State: &wait, Reason: "Explicit validation pause"})
		if err != nil {
			if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
				t.Fatal("rejected validation changed state")
			}
			rejected = true
			break
		}
		before = domainCurrent(t, s, id)
		active := TimelineStageState("in_progress")
		_, err = s.TimelinePatchEvent(id, stage, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: fmt.Sprintf("reserve-resume-%d", cycle), Actor: actor, ExpectedVersion: domainEvent(t, before, stage).Version, ExpectedSessionVersion: &before.Version, StageState: &active, Reason: "Explicit resume"})
		if err != nil {
			if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
				t.Fatal("rejected resume changed state")
			}
			rejected = true
			break
		}
	}
	if !rejected {
		t.Fatal("nonclosing lifecycle consumed guaranteed closure slots")
	}
	current := domainCurrent(t, s, id)
	completed := TimelineStageState("completed")
	_, err := s.TimelinePatchEvent(id, stage, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "reserve-finish-build", Actor: actor, ExpectedVersion: domainEvent(t, current, stage).Version, ExpectedSessionVersion: &current.Version, StageState: &completed, Reason: "Build finished"})
	if err != nil {
		t.Fatal("cannot release foreground within reserve", err)
	}
	_, err = s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "guaranteed-closure", Actor: actor, Kind: "closure", Title: "Finished", Summary: "Keep the complete trace", Reason: "Explicit final closure", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Finished"}})
	if err != nil {
		t.Fatal("guaranteed closure failed", err)
	}
	current = domainCurrent(t, s, id)
	if current.State != TimelineClosed || len(current.Events) > TimelineMaxEvents {
		t.Fatal("closure state/history invalid")
	}
}

func TestTimelineDomainLifecycleCannotConsumeClosureReceipts(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	record := timelineClone(s.timeline[id])
	sample := record.Receipts[0]
	for i := len(record.Receipts); i < TimelineMaxReceipts-1; i++ {
		receipt := sample
		receipt.OperationID = fmt.Sprintf("receipt-%d", i)
		receipt.Result.OperationID = receipt.OperationID
		record.Receipts = append(record.Receipts, receipt)
	}
	if err := s.timelineSave(record); err != nil {
		t.Fatal(err)
	}
	before := domainCurrent(t, s, id)
	same := before.Title
	_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "receipt-overflow", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: before.Version, Title: &same})
	requireTimelineCode(t, err, "capacity")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("receipt exhaustion partially mutated work")
	}
	result, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "receipt-closure", Actor: timelineTestActor(s, id), Kind: "closure", Title: "Finished", Summary: "Keep durable history", Reason: "Explicit terminal closure", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Finished"}})
	if err != nil {
		t.Fatal("closure receipt unavailable", err)
	}
	if _, err = s.TimelineDelete(id, TimelineMutationRequest{SchemaVersion: 3, OperationID: "receipt-delete", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: result.SessionVersion}); err != nil {
		t.Fatal("terminal deletion receipt unavailable", err)
	}
}

func capacityRecord(t *testing.T, s *Store, id string, summaryLength int) *timelineRecord {
	t.Helper()
	record := timelineClone(s.timeline[id])
	actor := timelineTestActor(s, id)
	now := s.timelineTime()
	for i := 0; i < TimelineContentEvents; i++ {
		size := summaryLength
		if i == 0 {
			size = 1
		}
		record.Session.Events = append(record.Session.Events, TimelineEvent{EventID: fmt.Sprintf("bytes-%d", i), Sequence: int64(i + 1), Version: 1, Kind: "request", Actor: actor, Title: "Request", Summary: strings.Repeat("s", size), CreatedAt: now, UpdatedAt: now, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}})
	}
	return record
}
func capacityClose(t *testing.T, s *Store, id, op string) {
	t.Helper()
	_, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: op, Actor: timelineTestActor(s, id), Kind: "closure", Title: "Done", Summary: "Done", Reason: "Done", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Done"}})
	if err != nil {
		t.Fatal("small closure lost reserved serialized capacity", err)
	}
}

func TestTimelineDomainSerializedPublicClosureHeadroom(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	record := capacityRecord(t, s, id, 1)
	raw, _ := json.Marshal(record.Session)
	remaining := TimelineMaxBody - 2048 - len(raw)
	for i := 1; i < len(record.Session.Events) && remaining > 0; i++ {
		add := remaining
		if add > 3999 {
			add = 3999
		}
		record.Session.Events[i].Summary += strings.Repeat("s", add)
		remaining -= add
	}
	if remaining != 0 {
		t.Fatal("cannot construct exact public byte probe")
	}
	if err := s.timelineSave(record); err != nil {
		t.Fatal(err)
	}
	before := stateManifest(t, filepath.Dir(s.Dir()))
	_, err := s.TimelinePatchEvent(id, record.Session.Events[0].EventID, TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "consume-public-headroom", Actor: timelineTestActor(s, id), ExpectedVersion: 1, Summary: timelineString(strings.Repeat("p", 2041))})
	requireTimelineCode(t, err, "capacity")
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("rejected byte growth mutated state")
	}
	capacityClose(t, s, id, "public-reserve-close")
}

func TestTimelineDomainSerializedPrivateLedgerClosureHeadroom(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	record := capacityRecord(t, s, id, 2000)
	sample := record.Receipts[0]
	actor := timelineTestActor(s, id)
	source := record.Source
	for i := len(record.Receipts); i < TimelineMaxReceipts-3; i++ {
		op := fmt.Sprintf("ledger-fill-%d", i)
		receipt := sample
		receipt.OperationID = op
		receipt.Result.OperationID = op
		record.Receipts = append(record.Receipts, receipt)
		record.Audit = append(record.Audit, timelinePrivateAudit{AuditID: fmt.Sprintf("audit-fill-%d", i), Operation: "create_session", OperationID: op, ObjectID: id, Actor: &actor, Source: &source, Reason: strings.Repeat("s", 500), ReceivedAt: s.timelineTime()})
	}
	raw, _ := json.Marshal(record)
	remaining := timelineMaxRecordBytes - 2048 - len(raw)
	for i := 1; i < len(record.Audit) && remaining >= 5; i++ {
		escaped := remaining / 5
		if escaped > 500 {
			escaped = 500
		}
		record.Audit[i].Reason = strings.Repeat("<", escaped) + strings.Repeat("s", 500-escaped)
		remaining -= escaped * 5
	}
	if remaining < 0 || remaining > 4 {
		t.Fatal("cannot construct exact private byte probe", remaining)
	}
	record.Session.Events[0].Summary += strings.Repeat("s", remaining)
	raw, _ = json.Marshal(record)
	if len(raw) != timelineMaxRecordBytes-2048 {
		t.Fatal("private probe size", len(raw))
	}
	if err := s.timelineSave(record); err != nil {
		t.Fatal(err)
	}
	before := stateManifest(t, filepath.Dir(s.Dir()))
	summary := "Small metadata update"
	_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "consume-ledger-headroom", Actor: actor, ExpectedVersion: record.Session.Version, Summary: &summary})
	requireTimelineCode(t, err, "capacity")
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("rejected ledger growth mutated state")
	}
	capacityClose(t, s, id, "private-reserve-close")
}
