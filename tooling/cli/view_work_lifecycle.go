package main

import (
	"encoding/json"
	"strings"
	"tricell-hive/tooling/viewer"
)

// Lifecycle context retains semantic input; exact wire requests are separate
// journal steps. Registration resolves default assignees before preparing intent.
type viewWorkLifecycleBinding struct {
	TaskKey string               `json:"task_key"`
	Stage   viewer.TimelineStage `json:"stage"`
	EventID string               `json:"event_id,omitempty"`
}

type viewWorkLifecycleContext struct {
	Command      viewWorkCommand            `json:"command"`
	FirstStep    *viewWorkStep              `json:"first_step,omitempty"`
	Bindings     []viewWorkLifecycleBinding `json:"bindings,omitempty"`
	PrimaryStage bool                       `json:"primary_stage,omitempty"`
}

func workLifecycleSource(s viewer.TimelineSession, a viewer.TimelineActor) viewer.TimelineSource {
	for _, p := range s.Producers {
		if p.ProducerID == a.ProducerID {
			return p.Source
		}
	}
	return viewer.TimelineSource{}
}
func workLifecycleStage(s viewer.TimelineSession, stage string) *viewer.TimelineEvent {
	var latest *viewer.TimelineEvent
	for n := range s.Events {
		event := &s.Events[n]
		if event.Kind == "stage" && string(event.Stage) == stage && !event.Withdrawn {
			latest = event
		}
	}
	return latest
}
func workLifecycleTask(s viewer.TimelineSession, key string) *viewer.TimelineTask {
	for n := range s.Tasks {
		if s.Tasks[n].TaskKey == key {
			return &s.Tasks[n]
		}
	}
	return nil
}
func workLifecycleSegment(task *viewer.TimelineTask, stage string) *viewer.TimelineTaskSegment {
	for n := range task.Segments {
		if string(task.Segments[n].PlannedStage) == stage && !task.Segments[n].Withdrawn {
			return &task.Segments[n]
		}
	}
	return nil
}
func workLifecycleBase(s viewer.TimelineSession, a viewer.TimelineActor, reason string) viewer.TimelineNativeOperation {
	return viewer.TimelineNativeOperation{SchemaVersion: 3, OperationID: workID("op-"), Actor: a, ExpectedSessionVersion: s.Version, Reason: reason}
}
func workLifecycleProducer(s viewer.TimelineSession, source viewWorkSource) (string, error) {
	for _, p := range s.Producers {
		if p.Source.Host == source.Host && p.Source.NativeSessionID == source.NativeSessionID {
			return p.ProducerID, nil
		}
	}
	return "", workStateError("accepted_assignee_required")
}
func workLifecyclePlan(s viewer.TimelineSession, id string) bool {
	for _, event := range s.Events {
		if event.Kind == "stage" && event.Stage == "plan" && !event.Withdrawn {
			for _, snapshot := range event.DocumentSnapshots {
				if snapshot.DocumentRevisionID == id {
					return true
				}
			}
		}
	}
	return false
}

func (e *viewWorkEngine) lifecycle(c viewWorkCommand) (viewWorkResult, error) {
	s, err := e.session(c.WorkID)
	if err != nil {
		return viewWorkResult{}, err
	}
	actor, err := e.actor(s, c)
	if err != nil {
		return viewWorkResult{}, err
	}
	if s.State == viewer.TimelineClosed || s.State == "failed" {
		return viewWorkResult{}, workStateError("session_closed")
	}
	for _, pending := range e.journal.Pending {
		if pending.WorkID == c.WorkID {
			return viewWorkResult{}, &viewWorkError{Code: "pending_operation", Pending: pending.ID, Next: "resume the original intent before another lifecycle operation"}
		}
	}
	var bindings []viewWorkLifecycleBinding
	// Prevalidate the entire batch, including remote collisions and accepted
	// assignees, before saving an intent or creating its first task.
	if c.Action == "tasks register" {
		if err := validateViewWorkTasks(c.Tasks); err != nil {
			return viewWorkResult{}, err
		}
		// A work without a change folder or a published plan has no revision to
		// confirm; any other registers against a confirmed one.
		if (c.PlanRevision != "" || e.workHasPlan(s)) && !workLifecyclePlan(s, c.PlanRevision) {
			return viewWorkResult{}, workStateError("confirmed_plan_revision_required")
		}
		if c.Reason == "" {
			c.Reason = "Register declared plan tasks"
		}
		for n := range c.Tasks {
			for m := range c.Tasks[n].Segments {
				if c.Tasks[n].Segments[m].Assignee == nil {
					for _, p := range s.Producers {
						if p.ProducerID == s.ForegroundProducerID {
							source := viewWorkSource{Host: p.Source.Host, NativeSessionID: p.Source.NativeSessionID}
							c.Tasks[n].Segments[m].Assignee = &source
						}
					}
				}
			}
		}
		if err := e.validateLifecycleRegistration(s, actor, c); err != nil {
			return viewWorkResult{}, err
		}
		for _, task := range c.Tasks {
			for _, segment := range task.Segments {
				stage := workLifecycleStage(s, string(segment.Stage))
				if stage != nil && stage.StageState != "completed" && stage.StageState != "failed" {
					bindings = append(bindings, viewWorkLifecycleBinding{TaskKey: task.Key, Stage: segment.Stage, EventID: stage.EventID})
				}
			}
		}
		missing := make([]viewWorkTaskInput, 0, len(c.Tasks))
		for _, task := range c.Tasks {
			if workLifecycleTask(s, task.Key) == nil {
				missing = append(missing, task)
			}
		}
		c.Tasks = missing
	}
	// Build and validate the first mutation before creating durable pending state.
	// Invalid transitions and missing prerequisites do not leave a blocking intent.
	method, path, body, _, err := e.nextLifecycleStep(s, actor, c)
	if err != nil {
		return viewWorkResult{}, err
	}
	primaryStage := c.Action == "stage" && method != ""
	if c.Action == "stage" && c.State != "completed" && c.State != "failed" {
		eventID := ""
		if stage := workLifecycleStage(s, c.Stage); stage != nil && (!primaryStage || stage.StageState != "completed" && stage.StageState != "failed") {
			eventID = stage.EventID
		}
		for _, task := range s.Tasks {
			if task.Withdrawn {
				continue
			}
			if segment := workLifecycleSegment(&task, c.Stage); segment != nil {
				bindings = append(bindings, viewWorkLifecycleBinding{TaskKey: task.TaskKey, Stage: segment.PlannedStage, EventID: eventID})
			}
		}
	}
	if method == "" {
		method, path, body, err = e.nextLifecycleBinding(s, actor, c.Reason, bindings, 0)
		if err != nil {
			return viewWorkResult{}, err
		}
	}
	if method == "" {
		return viewWorkResult{WorkID: c.WorkID, Result: "unchanged", SessionVersion: s.Version}, nil
	}
	context := viewWorkLifecycleContext{Command: c, Bindings: bindings, PrimaryStage: primaryStage, FirstStep: &viewWorkStep{Method: method, Path: path, ContentType: "application/json", Body: body}}
	i, err := e.prepareIntent(c.Action, c.WorkID, workLifecycleSource(s, actor), context)
	if err != nil {
		return viewWorkResult{}, err
	}
	if _, err = e.appendStep(i, method, path, "application/json", body); err != nil {
		return viewWorkResult{}, err
	}
	return e.resumeLifecycle(i)
}

func (e *viewWorkEngine) lifecycleRequest(s viewer.TimelineSession, kind, object, method, path string, request any) (string, string, []byte, error) {
	var err error
	switch r := request.(type) {
	case *viewer.TimelineAppendEventRequest:
		err = r.Validate()
		if err == nil {
			err = viewer.ValidateTimelineEventReferences(s, *r)
		}
	case *viewer.TimelineEventPatchRequest:
		err = viewer.ValidateTimelineEventPatch(s, object, *r)
	default:
		err = viewer.ValidateTimelineV3Operation(request)
		if err == nil {
			err = viewer.ValidateTimelineV3OperationReferences(s, kind, object, request)
		}
	}
	if err != nil {
		return "", "", nil, err
	}
	raw, err := json.Marshal(request)
	return method, path, raw, err
}

func (e *viewWorkEngine) nextLifecycleStep(s viewer.TimelineSession, a viewer.TimelineActor, c viewWorkCommand) (string, string, []byte, string, error) {
	basePath := "/api/timeline/sessions/" + s.ViewerSessionID
	makeRequest := func(kind, object, method, path string, r any) (string, string, []byte, string, error) {
		m, p, b, err := e.lifecycleRequest(s, kind, object, method, path, r)
		return m, p, b, object, err
	}
	switch c.Action {
	case "stage":
		if err := viewer.ValidateTimelineAuthority(s, "stage_update", a, ""); err != nil {
			return "", "", nil, "", err
		}
		latest := workLifecycleStage(s, c.Stage)
		if latest != nil && string(latest.StageState) == c.State && c.PreviousStage == "" && (c.Summary == "" || c.Summary == latest.Summary) {
			return "", "", nil, latest.EventID, nil
		}
		if latest == nil || latest.StageState == "completed" || latest.StageState == "failed" {
			if c.State != "in_progress" {
				return "", "", nil, "", workStateError("stage_start_required")
			}
			if latest != nil && c.PreviousStage != latest.EventID || latest == nil && c.PreviousStage != "" {
				return "", "", nil, "", workStateError("explicit_stage_predecessor_required")
			}
			summary := c.Summary
			if summary == "" {
				summary = c.Reason
			}
			r := &viewer.TimelineAppendEventRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: a, Reason: c.Reason, Kind: "stage", Stage: viewer.TimelineStage(c.Stage), StageState: "in_progress", ExpectedSessionVersion: &s.Version, Summary: summary, TicketIDs: []string{}, RelatedEventID: c.PreviousStage}
			return makeRequest("stage_start", "", "POST", basePath+"/events", r)
		}
		if c.PreviousStage != "" {
			return "", "", nil, "", workStateError("previous_stage_only_for_reentry")
		}
		if string(latest.StageState) == c.State && (c.Summary == "" || c.Summary == latest.Summary) {
			return "", "", nil, latest.EventID, nil
		}
		state := viewer.TimelineStageState(c.State)
		r := &viewer.TimelineEventPatchRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: a, Reason: c.Reason, ExpectedVersion: latest.Version, ExpectedSessionVersion: &s.Version, StageState: &state}
		if c.Summary != "" {
			r.Summary = &c.Summary
		}
		return makeRequest("stage_update", latest.EventID, "PATCH", basePath+"/events/"+latest.EventID, r)
	case "close":
		if s.CurrentStageEventID != nil {
			return "", "", nil, "", workStateError("stage_active")
		}
		if err := viewer.ValidateTimelineAuthority(s, "closure", a, ""); err != nil {
			return "", "", nil, "", err
		}
		r := &viewer.TimelineAppendEventRequest{SchemaVersion: 3, OperationID: workID("op-"), Actor: a, Reason: c.Reason, Kind: "closure", Title: "Work " + c.Outcome, Summary: c.Reason, TicketIDs: []string{}, Closure: &viewer.TimelineClosure{Reason: c.Outcome, Result: c.Reason}}
		return makeRequest("closure", "", "POST", basePath+"/events", r)
	case "tasks register":
		if err := e.validateLifecycleRegistration(s, a, c); err != nil {
			return "", "", nil, "", err
		}
		for _, input := range c.Tasks {
			if workLifecycleTask(s, input.Key) != nil {
				continue
			}
			r, err := workLifecycleCreate(s, a, c, input)
			if err != nil {
				return "", "", nil, "", err
			}
			return makeRequest("task_create", "", "POST", basePath+"/tasks", r)
		}
		return "", "", nil, "", nil
	case "task update":
		task := workLifecycleTask(s, c.Key)
		if task == nil || task.Withdrawn {
			return "", "", nil, "", workStateError("task_not_found")
		}
		if c.Metadata != nil {
			r, err := workLifecycleMetadata(s, a, c, *task)
			if err != nil {
				return "", "", nil, "", err
			}
			return makeRequest("task_patch", task.TaskID, "PATCH", basePath+"/tasks/"+task.TaskID, r)
		}
		segment := workLifecycleSegment(task, c.Stage)
		if segment == nil {
			return "", "", nil, "", workStateError("task_segment_not_found")
		}
		if a.ProducerID != s.ForegroundProducerID && a.ProducerID != segment.AssignedProducerID {
			return "", "", nil, "", workStateError("segment_assignee_required")
		}
		r := &viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: workLifecycleBase(s, a, c.Reason), ExpectedTaskVersion: task.TaskVersion, SegmentID: segment.SegmentID, State: c.State, Block: c.Block, Resolution: c.Resolution, ReviewKind: c.ReviewKind}
		if err := viewer.ValidateTimelineTaskTransition(segment.State, *r); err != nil {
			return "", "", nil, "", err
		}
		latest := workLifecycleStage(s, c.Stage)
		if latest == nil || latest.StageState == "completed" || latest.StageState == "failed" {
			return "", "", nil, "", workStateError("active_stage_required")
		}
		if segment.DisplayStageEventID != latest.EventID {
			patch := &viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workLifecycleBase(s, a, c.Reason), ExpectedTaskVersion: task.TaskVersion, SegmentID: segment.SegmentID, DisplayStageEventID: latest.EventID}
			return makeRequest("task_patch", task.TaskID, "PATCH", basePath+"/tasks/"+task.TaskID, patch)
		}
		if segment.State == c.State {
			return "", "", nil, task.TaskID, nil
		}
		return makeRequest("task_transition", task.TaskID, "POST", basePath+"/tasks/"+task.TaskID+"/transitions", r)
	}
	return "", "", nil, "", workStateError("unknown_lifecycle_action")
}

func workLifecycleCreate(s viewer.TimelineSession, a viewer.TimelineActor, c viewWorkCommand, input viewWorkTaskInput) (*viewer.TimelineTaskCreateRequest, error) {
	r := &viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workLifecycleBase(s, a, c.Reason), TaskKey: input.Key, Title: input.Title, Summary: input.Summary, PlanOrder: input.Order, PlanRevisionID: c.PlanRevision}
	for _, seg := range input.Segments {
		id := ""
		var err error
		if seg.Assignee != nil {
			id, err = workLifecycleProducer(s, *seg.Assignee)
			if err != nil {
				return nil, err
			}
		}
		r.Segments = append(r.Segments, viewer.TimelineTaskSegmentInput{PlannedStage: seg.Stage, Order: seg.Order, AssignedProducerID: id})
	}
	return r, nil
}
func (e *viewWorkEngine) validateLifecycleRegistration(s viewer.TimelineSession, a viewer.TimelineActor, c viewWorkCommand) error {
	if err := viewer.ValidateTimelineAuthority(s, "task_create", a, ""); err != nil {
		return err
	}
	simulated := s
	simulated.Tasks = append([]viewer.TimelineTask{}, s.Tasks...)
	for _, input := range c.Tasks {
		r, err := workLifecycleCreate(simulated, a, c, input)
		if err != nil {
			return err
		}
		if old := workLifecycleTask(simulated, input.Key); old != nil {
			if old.Withdrawn || old.Title != input.Title || old.Summary != input.Summary || old.PlanOrder != input.Order || old.PlanRevisionID != c.PlanRevision || len(old.Segments) != len(r.Segments) {
				return workStateError("task_metadata_conflict")
			}
			for n, seg := range old.Segments {
				want := r.Segments[n]
				if seg.Withdrawn || seg.PlannedStage != want.PlannedStage || seg.Order != want.Order || seg.AssignedProducerID != want.AssignedProducerID {
					return workStateError("task_metadata_conflict")
				}
			}
			continue
		}
		if err := viewer.ValidateTimelineV3OperationReferences(simulated, "task_create", "", r); err != nil {
			return err
		}
		dummy := viewer.TimelineTask{TaskKey: input.Key, PlanOrder: input.Order}
		for _, seg := range r.Segments {
			dummy.Segments = append(dummy.Segments, viewer.TimelineTaskSegment{PlannedStage: seg.PlannedStage, Order: seg.Order})
		}
		simulated.Tasks = append(simulated.Tasks, dummy)
	}
	return nil
}

func workLifecycleMetadata(s viewer.TimelineSession, a viewer.TimelineActor, c viewWorkCommand, task viewer.TimelineTask) (*viewer.TimelineTaskPatchRequest, error) {
	m := c.Metadata
	r := &viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workLifecycleBase(s, a, c.Reason), ExpectedTaskVersion: task.TaskVersion, PlanOrder: m.Order, Title: m.Title, Summary: m.Summary, Withdrawn: m.Withdrawn, PlanRevisionID: m.PlanRevision, Order: m.SegmentOrder}
	if m.PlanRevision != "" && !workLifecyclePlan(s, m.PlanRevision) {
		return nil, workStateError("confirmed_plan_revision_required")
	}
	if m.Stage != "" {
		segment := workLifecycleSegment(&task, string(m.Stage))
		if segment == nil {
			return nil, workStateError("task_segment_not_found")
		}
		r.SegmentID = segment.SegmentID
	}
	if m.Assignee != nil {
		id, err := workLifecycleProducer(s, *m.Assignee)
		if err != nil {
			return nil, err
		}
		r.AssignedProducerID = id
	}
	if m.DecisionReference != "" {
		for _, event := range s.Events {
			if event.EventID == m.DecisionReference && !event.Withdrawn && event.Target != nil && (event.Kind == "changes_reported" || event.Kind == "approval_reported") {
				target := *event.Target
				r.DecisionTarget = &target
			}
			if event.EventID == m.DecisionReference && !event.Withdrawn && event.EvidenceID != "" {
				r.DecisionTarget = &viewer.TimelineDecisionTarget{Kind: "evidence", EventID: event.EventID, EvidenceID: event.EvidenceID}
			}
		}
		if r.DecisionTarget == nil {
			return nil, workStateError("confirmed_decision_reference_required")
		}
	}
	scopeChanged := m.Order != nil && *m.Order != task.PlanOrder
	if r.SegmentID == "" {
		scopeChanged = scopeChanged || m.Withdrawn != nil && *m.Withdrawn != task.Withdrawn
	} else {
		for _, segment := range task.Segments {
			if segment.SegmentID != r.SegmentID {
				continue
			}
			scopeChanged = scopeChanged || m.Withdrawn != nil && *m.Withdrawn != segment.Withdrawn || m.SegmentOrder != nil && *m.SegmentOrder != segment.Order || r.AssignedProducerID != "" && r.AssignedProducerID != segment.AssignedProducerID
		}
	}
	if scopeChanged && r.PlanRevisionID == "" && r.DecisionTarget == nil {
		return nil, workStateError("scope_reference_required")
	}

	return r, nil
}

func (e *viewWorkEngine) resumeLifecycle(i *viewWorkIntent) (viewWorkResult, error) {
	var context viewWorkLifecycleContext
	if err := json.Unmarshal(i.Context, &context); err != nil || context.Command.Action != i.Action || context.Command.WorkID != i.WorkID {
		return viewWorkResult{}, workStateError("invalid_lifecycle_context")
	}
	c := context.Command
	if c.Host != string(i.Source.Host) || c.NativeSessionID != i.Source.NativeSessionID {
		return viewWorkResult{}, workStateError("source_mismatch")
	}
	result := viewWorkResult{WorkID: i.WorkID, Result: "updated"}
	for {
		if len(i.Steps) == 0 {
			if context.FirstStep != nil {
				step := context.FirstStep
				if _, err := e.appendStep(i, step.Method, step.Path, step.ContentType, step.Body); err != nil {
					return viewWorkResult{}, err
				}
			} else {
				return viewWorkResult{}, workStateError("incomplete_lifecycle_preparation")
			}
		}
		for index := range i.Steps {
			raw, err := e.executeStep(i, index)
			if err != nil {
				return viewWorkResult{}, err
			}
			receipt, err := viewer.DecodeTimelineOperationResult(raw)
			if err != nil {
				return viewWorkResult{}, err
			}
			result.SessionVersion = receipt.SessionVersion
			if receipt.TaskID != "" {
				result.TaskID = receipt.TaskID
			}
			if receipt.EventID != "" {
				result.StageID = receipt.EventID
			}
		}
		s, err := e.session(i.WorkID)
		if err != nil {
			return viewWorkResult{}, err
		}
		actor, err := e.actor(s, c)
		if err != nil {
			return viewWorkResult{}, err
		}
		if workLifecycleSource(s, actor) != i.Source {
			return viewWorkResult{}, workStateError("source_mismatch")
		}
		last := i.Steps[len(i.Steps)-1]
		creates := 0
		for _, step := range i.Steps {
			if step.Method == "POST" && strings.HasSuffix(step.Path, "/tasks") {
				creates++
			}
		}
		if c.Action == "stage" || c.Action == "tasks register" && creates == len(c.Tasks) {
			bindings := append([]viewWorkLifecycleBinding{}, context.Bindings...)
			if context.PrimaryStage {
				receipt, err := viewer.DecodeTimelineOperationResult(i.Steps[0].Response)
				if err != nil {
					return viewWorkResult{}, err
				}
				for n := range bindings {
					if bindings[n].EventID == "" {
						bindings[n].EventID = receipt.EventID
					}
				}
			}
			receipt, err := viewer.DecodeTimelineOperationResult(last.Response)
			if err != nil {
				return viewWorkResult{}, err
			}
			method, path, body, err := e.nextLifecycleBinding(s, actor, c.Reason, bindings, receipt.SessionVersion)
			if err != nil {
				return viewWorkResult{}, err
			}
			if method != "" {
				if _, err = e.appendStep(i, method, path, "application/json", body); err != nil {
					return viewWorkResult{}, err
				}
				continue
			}
		}
		done := c.Action == "stage" || c.Action == "close" || c.Action == "tasks register" && creates == len(c.Tasks) || c.Action == "task update" && (c.Metadata != nil || strings.HasSuffix(last.Path, "/transitions"))
		if done {
			result.SessionVersion = s.Version
			if err := e.finishIntent(i); err != nil {
				return viewWorkResult{}, err
			}
			return result, nil
		}
		if c.Action == "task update" && c.Metadata == nil {
			receipt, err := viewer.DecodeTimelineOperationResult(last.Response)
			if err != nil {
				return viewWorkResult{}, err
			}
			var binding viewer.TimelineTaskPatchRequest
			if json.Unmarshal(last.Body, &binding) != nil || binding.SegmentID == "" {
				return viewWorkResult{}, workStateError("invalid_lifecycle_binding")
			}
			// Use versions confirmed by our binding receipt, rather than a newer
			// actor's state. A competing write receives a definitive server
			// rejection, preserving an exact request that can be reconciled.
			base := workLifecycleBase(s, actor, c.Reason)
			base.ExpectedSessionVersion = receipt.SessionVersion
			transition := viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: base, ExpectedTaskVersion: receipt.TaskVersion, SegmentID: binding.SegmentID, State: c.State, Block: c.Block, Resolution: c.Resolution, ReviewKind: c.ReviewKind}
			raw, err := json.Marshal(transition)
			if err != nil {
				return viewWorkResult{}, err
			}
			if _, err = e.appendStep(i, "POST", last.Path+"/transitions", "application/json", raw); err != nil {
				return viewWorkResult{}, err
			}
			continue
		}
		if c.Action == "tasks register" && creates < len(c.Tasks) {
			receipt, err := viewer.DecodeTimelineOperationResult(last.Response)
			if err != nil {
				return viewWorkResult{}, err
			}
			if s.Version != receipt.SessionVersion {
				request, err := workLifecycleCreate(s, actor, c, c.Tasks[creates])
				if err != nil {
					return viewWorkResult{}, err
				}
				request.ExpectedSessionVersion = receipt.SessionVersion
				raw, err := json.Marshal(request)
				if err != nil {
					return viewWorkResult{}, err
				}
				if _, err = e.appendStep(i, "POST", "/api/timeline/sessions/"+i.WorkID+"/tasks", "application/json", raw); err != nil {
					return viewWorkResult{}, err
				}
				continue
			}
		}
		method, path, body, _, err := e.nextLifecycleStep(s, actor, c)
		if err != nil {
			return viewWorkResult{}, err
		}
		if method == "" {
			result.SessionVersion = s.Version
			if err := e.finishIntent(i); err != nil {
				return viewWorkResult{}, err
			}
			return result, nil
		}
		// A partial registration resumes only missing tasks. A transition resumes
		// after its separate display binding, with versions read afresh.
		if _, err = e.appendStep(i, method, path, "application/json", body); err != nil {
			return viewWorkResult{}, err
		}
	}
}

// Only the already accepted lead performs these metadata writes. Assignees
// subsequently use their own producer for transitions without changing authority.
func (e *viewWorkEngine) nextLifecycleBinding(s viewer.TimelineSession, actor viewer.TimelineActor, reason string, bindings []viewWorkLifecycleBinding, expected int64) (string, string, []byte, error) {
	for _, binding := range bindings {
		task := workLifecycleTask(s, binding.TaskKey)
		if task == nil || task.Withdrawn {
			continue
		}
		segment := workLifecycleSegment(task, string(binding.Stage))
		if segment == nil || segment.DisplayStageEventID == binding.EventID {
			continue
		}
		if binding.EventID == "" {
			return "", "", nil, workStateError("stage_binding_required")
		}
		request := &viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workLifecycleBase(s, actor, reason), ExpectedTaskVersion: task.TaskVersion, SegmentID: segment.SegmentID, DisplayStageEventID: binding.EventID}
		if expected == 0 || expected == s.Version {
			if err := viewer.ValidateTimelineV3OperationReferences(s, "task_patch", task.TaskID, request); err != nil {
				return "", "", nil, err
			}
		} else if err := viewer.ValidateTimelineV3Operation(request); err != nil {
			return "", "", nil, err
		}
		if expected > 0 {
			request.ExpectedSessionVersion = expected
		}
		raw, err := json.Marshal(request)
		return "PATCH", "/api/timeline/sessions/" + s.ViewerSessionID + "/tasks/" + task.TaskID, raw, err
	}
	return "", "", nil, nil
}
