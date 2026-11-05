package viewer

import (
	"encoding/json"
	"reflect"
	"time"
)

// timelinePrivateAudit records generic operations without duplicating content.
// Lifecycle and membership transitions already have public immutable events.
type timelinePrivateAudit struct {
	AuditID     string                `json:"audit_id"`
	Operation   TimelineOperationKind `json:"operation"`
	OperationID string                `json:"operation_id"`
	ObjectID    string                `json:"object_id"`
	Actor       *TimelineActor        `json:"actor,omitempty"`
	Source      *TimelineSource       `json:"source,omitempty"`
	Reason      string                `json:"reason"`
	ReceivedAt  string                `json:"received_at"`
}

func (s *Store) timelineAudit(r *timelineRecord, kind TimelineOperationKind, op, object string, actor *TimelineActor, source *TimelineSource, reason string) error {
	id, err := timelineRandomID("audit-")
	if err != nil {
		return err
	}
	if reason == "" {
		reason = "Explicit " + string(kind)
	}
	if actor != nil && actor.Kind == "native_producer" && r.Session != nil {
		for _, producer := range r.Session.Producers {
			if producer.ProducerID == actor.ProducerID {
				actual := producer.Source
				source = &actual
				break
			}
		}
	}
	if actor != nil {
		recorded := actor.recorded()
		actor = &recorded
	}
	r.Audit = append(r.Audit, timelinePrivateAudit{AuditID: id, Operation: kind, OperationID: op, ObjectID: object, Actor: actor, Source: source, Reason: reason, ReceivedAt: s.timelineTime()})
	return nil
}
func (s *Store) timelineAuditEvent(v *TimelineSession, kind TimelineEventKind, actor TimelineActor, stage, task, segment string, a TimelineAudit, lifecycle bool) error {
	if err := ValidateTimelineEventCapacity(len(v.Events), 1, lifecycle); err != nil {
		return err
	}
	id, err := timelineRandomID("event-")
	if err != nil {
		return err
	}
	now := s.timelineTime()
	title := []rune(a.Reason)
	if len(title) > 200 {
		title = title[:200]
	}
	v.Events = append(v.Events, TimelineEvent{EventID: id, Sequence: int64(len(v.Events) + 1), Version: 1, Kind: kind, Actor: actor.recorded(), StageEventID: stage, TaskID: task, SegmentID: segment, Title: string(title), Summary: a.Reason, CreatedAt: now, UpdatedAt: now, Origin: "server", TicketRefs: []TimelineTicketRef{}, Audit: &a})
	return nil
}
func timelineBumpTask(v *TimelineSession, ids ...string) error {
	set := map[string]bool{}
	for _, id := range ids {
		if id != "" {
			set[id] = true
		}
	}
	for i := range v.Tasks {
		if set[v.Tasks[i].TaskID] {
			if v.Tasks[i].TaskVersion >= TimelineMaxNumber {
				return timelineFailure("capacity", 413)
			}
			v.Tasks[i].TaskVersion++
		}
	}
	return nil
}
func timelineTaskIndex(v *TimelineSession, id string) int {
	for i := range v.Tasks {
		if v.Tasks[i].TaskID == id {
			return i
		}
	}
	return -1
}
func timelineWaitIndex(v *TimelineSession, id string) int {
	for i := range v.Waits {
		if v.Waits[i].WaitID == id {
			return i
		}
	}
	return -1
}
func timelineWaitAggregate(w TimelineCIWait) string {
	observed := map[string]TimelineCIObservation{}
	for _, o := range w.Observations {
		if w.ApplicableAttemptID[o.CheckKey] != o.AttemptID {
			continue
		}
		previous, exists := observed[o.CheckKey]
		oldTime, _ := time.Parse(time.RFC3339Nano, previous.ObservedAt)
		nextTime, _ := time.Parse(time.RFC3339Nano, o.ObservedAt)
		if !exists || oldTime.Before(nextTime) {
			observed[o.CheckKey] = o
		}
	}
	return AggregateTimelineCIWait(w.Policy, w.ApplicableAttemptID, observed)
}
func timelineOperationBase(value any) (TimelineNativeOperation, string) {
	raw, _ := json.Marshal(value)
	var base TimelineNativeOperation
	_ = json.Unmarshal(raw, &base)
	return base, base.OperationID
}
func (s *Store) timelineBind(r *timelineRecord, source TimelineSource) error {
	root, key, _, err := s.timelineProject(source)
	if err != nil {
		return err
	}
	if err = ValidateTimelineResolvedProject(*r.Session, source, key); err != nil {
		return err
	}
	for _, candidate := range timelineSourceRoots(root, source.ProjectRoot) {
		found := false
		for _, old := range r.Roots {
			found = found || old == candidate
		}
		if !found {
			if len(r.Roots) >= 64 {
				return timelineFailure("capacity", 413)
			}
			r.Roots = append(r.Roots, candidate)
		}
	}
	return nil
}

// TimelineV3Operation runs one conditional domain transaction, including replay,
// authority, source resolution, history and receipt, under the store mutex.
func (s *Store) TimelineV3Operation(id, kind, objectID string, value any) (TimelineOperationResult, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	value, err = DecodeTimelineV3Operation(kind, raw)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	base, op := timelineOperationBase(value)
	s.mu.Lock()
	defer s.mu.Unlock()
	old, err := s.timelineGetLocked(id)
	if err != nil {
		return TimelineOperationResult{}, err
	}
	digest := timelineDigest(kind, id, objectID, value, nil)
	if result, done, err := s.timelineReplay(old, op, digest); done || err != nil {
		return result, err
	}
	if err = ValidateTimelineV3OperationReferences(*old.Session, kind, objectID, value); err != nil {
		return TimelineOperationResult{}, err
	}
	if old.Session.State == TimelineClosed || old.Session.State == "failed" {
		return TimelineOperationResult{}, timelineFailure("session_terminal", 409)
	}
	r := timelineClone(old)
	v := r.Session
	result := timelineResult(r, TimelineOperationKind(kind), op)
	publicAudit := false
	switch request := value.(type) {
	case *TimelineJoinRequestInput:
		if err = s.timelineBind(r, request.Source); err != nil {
			return result, err
		}
		result.RequestID, err = timelineRandomID("request-")
		if err != nil {
			return result, err
		}
		v.PendingJoinRequests = append(v.PendingJoinRequests, TimelineJoinRequest{RequestID: result.RequestID, Source: request.Source, RequestedAt: s.timelineTime(), Reason: request.Reason})
		if err = s.timelineAudit(r, TimelineOperationKind(kind), op, result.RequestID, nil, &request.Source, request.Reason); err != nil {
			return result, err
		}
	case *TimelineTicketCreateRequest, *TimelineTicketPatchRequest, *TimelineDeliveryCreateRequest, *TimelineDeliveryPatchRequest, *TimelineCandidateCreateRequest:
		err = s.timelineWriteCatalog(v, value, objectID, &result)
	case *TimelineProducerAcceptRequest:
		result.RequestID = request.RequestID
		result.ProducerID, err = s.timelineAccept(r, request.RequestID, base.Actor, kind, base.Reason, false)
		publicAudit = true
	case *TimelineRecoverLeadRequest:
		result.RequestID = request.RequestID
		result.ProducerID, err = s.timelineAccept(r, request.RequestID, base.Actor, kind, base.Reason, true)
		publicAudit = true
	case *TimelineForegroundTransferRequest:
		result.ProducerID = request.TargetProducerID
		if v.ForegroundProducerID == request.TargetProducerID {
			return result, timelineFailure("already_in_state", 409)
		}
		previous := v.ForegroundProducerID
		v.ForegroundProducerID = request.TargetProducerID
		err = s.timelineAuditEvent(v, "foreground_transferred", base.Actor, "", "", "", TimelineAudit{Operation: kind, Reason: base.Reason, PreviousProducerID: previous, NextProducerID: request.TargetProducerID}, true)
		publicAudit = true
	case *TimelineTaskCreateRequest:
		result.TaskID, err = timelineRandomID("task-")
		if err != nil {
			return result, err
		}
		task := TimelineTask{TaskID: result.TaskID, TaskKey: request.TaskKey, PlanOrder: request.PlanOrder, Title: request.Title, Summary: request.Summary, TaskVersion: 1, PlanRevisionID: request.PlanRevisionID, State: "not_started", Segments: []TimelineTaskSegment{}}
		for _, input := range request.Segments {
			sid, e := timelineRandomID("segment-")
			if e != nil {
				return result, e
			}
			assignee := input.AssignedProducerID
			if assignee == "" {
				assignee = v.ForegroundProducerID
			}
			task.Segments = append(task.Segments, TimelineTaskSegment{SegmentID: sid, PlannedStage: input.PlannedStage, Order: input.Order, AssignedProducerID: assignee, State: "not_started"})
		}
		v.Tasks = append(v.Tasks, task)
		result.TaskVersion = 1
	case *TimelineTaskPatchRequest:
		i := timelineTaskIndex(v, objectID)
		task := &v.Tasks[i]
		before := *task
		before.Segments = append([]TimelineTaskSegment{}, task.Segments...)
		if request.PlanOrder != nil {
			task.PlanOrder = *request.PlanOrder
		}
		if request.Title != nil {
			task.Title = *request.Title
		}
		if request.Summary != nil {
			task.Summary = *request.Summary
		}
		if request.PlanRevisionID != "" {
			task.PlanRevisionID = request.PlanRevisionID
		}
		if request.SegmentID == "" {
			if request.Withdrawn != nil {
				task.Withdrawn = *request.Withdrawn
			}
		} else {
			for j := range task.Segments {
				seg := &task.Segments[j]
				if seg.SegmentID != request.SegmentID {
					continue
				}
				if request.Order != nil {
					seg.Order = *request.Order
				}
				if request.AssignedProducerID != "" {
					seg.AssignedProducerID = request.AssignedProducerID
				}
				if request.DisplayStageEventID != "" && seg.DisplayStageEventID != "" && request.DisplayStageEventID != seg.DisplayStageEventID {
					linked := false
					for _, event := range v.Events {
						linked = linked || event.EventID == request.DisplayStageEventID && event.RelatedEventID == seg.DisplayStageEventID
					}
					if !linked {
						return result, timelineFailure("reentry_binding_required", 409)
					}
				}
				if request.DisplayStageEventID != "" {
					seg.DisplayStageEventID = request.DisplayStageEventID
				}
				if request.Withdrawn != nil {
					seg.Withdrawn = *request.Withdrawn
				}
			}
		}
		scopeChanged := before.Withdrawn != task.Withdrawn || before.PlanOrder != task.PlanOrder
		for j, previous := range before.Segments {
			next := task.Segments[j]
			scopeChanged = scopeChanged || previous.AssignedProducerID != next.AssignedProducerID || previous.Order != next.Order || previous.Withdrawn != next.Withdrawn
		}
		if scopeChanged && request.PlanRevisionID == "" && request.DecisionTarget == nil {
			return result, timelineFailure("scope_reference_required", 409)
		}
		task.State = AggregateTimelineTaskState(task.Segments)
		if !reflect.DeepEqual(before, *task) {
			if err = timelineBumpTask(v, objectID); err != nil {
				return result, err
			}
		}
		result.TaskID = objectID
		result.TaskVersion = task.TaskVersion
	case *TimelineTaskTransitionRequest:
		i := timelineTaskIndex(v, objectID)
		task := &v.Tasks[i]
		for j := range task.Segments {
			seg := &task.Segments[j]
			if seg.SegmentID != request.SegmentID {
				continue
			}
			previous := *seg
			if seg.State == request.State {
				result.TaskID = objectID
				result.TaskVersion = task.TaskVersion
				break
			}
			seg.State = request.State
			seg.Block = request.Block
			seg.ReviewKind = request.ReviewKind
			a := TimelineAudit{Operation: kind, Reason: base.Reason, PreviousState: previous.State, NextState: seg.State, PreviousSegment: &previous, NextSegment: seg, Resolution: request.Resolution, Block: request.Block, PlanRevisionID: task.PlanRevisionID}
			if previous.State == "blocked" {
				a.Block = previous.Block
			}
			if err = s.timelineAuditEvent(v, "task_state_transition", base.Actor, seg.DisplayStageEventID, objectID, seg.SegmentID, a, true); err != nil {
				return result, err
			}
			if err = timelineBumpTask(v, objectID); err != nil {
				return result, err
			}
			result.TaskID = objectID
			result.TaskVersion = task.TaskVersion
		}
		task.State = AggregateTimelineTaskState(task.Segments)
		publicAudit = true
	case *TimelineWaitCreateRequest:
		for _, w := range v.Waits {
			if w.CandidateID == request.CandidateID && w.Policy.PolicyID == request.Policy.PolicyID {
				return result, timelineFailure("wait_already_exists", 409)
			}
		}
		previous := ""
		for _, wait := range v.Waits {
			if wait.PullRequestRef.DeliveryID == request.PullRequestDeliveryID && wait.TaskID == request.TaskID && wait.SegmentID == request.SegmentID {
				previous = wait.WaitID
			}
		}
		if previous != "" && request.PreviousWaitID != previous {
			return result, timelineFailure("previous_wait_required", 409)
		}
		result.WaitID, err = timelineRandomID("wait-")
		if err != nil {
			return result, err
		}
		refs, e := CaptureTimelineDeliveryReferences(*v, []string{request.PullRequestDeliveryID})
		if e != nil {
			return result, e
		}
		wait := TimelineCIWait{WaitID: result.WaitID, WaitVersion: 1, StageEventID: request.TargetStageEventID, TaskID: request.TaskID, SegmentID: request.SegmentID, CandidateID: request.CandidateID, PullRequestRef: refs[0], Policy: request.Policy, PreviousWaitID: request.PreviousWaitID, NextAction: request.NextAction, ApplicableAttemptID: map[string]string{}, Attempts: []TimelineCIAttempt{}, Observations: []TimelineCIObservation{}}
		wait.State = timelineWaitAggregate(wait)
		v.Waits = append(v.Waits, wait)
		result.WaitVersion = 1
		err = timelineBumpTask(v, wait.TaskID)
	case *TimelineWaitAttemptStartRequest:
		i := timelineWaitIndex(v, objectID)
		w := &v.Waits[i]
		if w.WaitVersion >= TimelineMaxNumber {
			return result, timelineFailure("capacity", 413)
		}
		result.AttemptID, err = timelineRandomID("attempt-")
		if err != nil {
			return result, err
		}
		generation := int64(len(w.Attempts) + 1)
		w.Attempts = append(w.Attempts, TimelineCIAttempt{AttemptID: result.AttemptID, Generation: generation, CheckKeys: request.CheckKeys, ExternalRun: request.ExternalRun, SourceURL: request.SourceURL, ObservedAt: request.ObservedAt, Actor: base.Actor})
		for _, key := range request.CheckKeys {
			w.ApplicableAttemptID[key] = result.AttemptID
		}
		w.WaitVersion++
		w.State = timelineWaitAggregate(*w)
		result.WaitID = objectID
		result.WaitVersion = w.WaitVersion
		result.Generation = generation
		err = timelineBumpTask(v, w.TaskID)
	case *TimelineWaitObserveRequest:
		i := timelineWaitIndex(v, objectID)
		w := &v.Waits[i]
		if w.WaitVersion >= TimelineMaxNumber {
			return result, timelineFailure("capacity", 413)
		}
		duplicate := ""
		for _, previous := range w.Observations {
			if previous.AttemptID == request.AttemptID && previous.CheckKey == request.CheckKey && previous.ObservedAt == request.ObservedAt && previous.State != request.State {
				return result, timelineFailure("observation_conflict", 409)
			}
			if previous.AttemptID == request.AttemptID && previous.CheckKey == request.CheckKey && previous.ExternalRun == request.ExternalRun && previous.State == request.State && previous.Summary == request.Summary && previous.ObservedAt == request.ObservedAt && previous.Actor == base.Actor {
				duplicate = previous.ObservationID
				break
			}
		}
		if duplicate != "" {
			result.ObservationID = duplicate
			result.WaitID = objectID
			result.WaitVersion = w.WaitVersion
			break
		}
		result.ObservationID, err = timelineRandomID("observation-")
		if err != nil {
			return result, err
		}
		w.Observations = append(w.Observations, TimelineCIObservation{ObservationID: result.ObservationID, AttemptID: request.AttemptID, CheckKey: request.CheckKey, ExternalRun: request.ExternalRun, State: request.State, Summary: request.Summary, ObservedAt: request.ObservedAt, ReceivedAt: s.timelineTime(), Actor: base.Actor})
		w.WaitVersion++
		w.State = timelineWaitAggregate(*w)
		result.WaitID = objectID
		result.WaitVersion = w.WaitVersion
		err = timelineBumpTask(v, w.TaskID)
	}
	if err != nil {
		return result, err
	}
	if _, join := value.(*TimelineJoinRequestInput); !join && !publicAudit {
		object := objectID
		if object == "" {
			object = result.TicketID
			if object == "" {
				object = result.DeliveryID
			}
			if object == "" {
				object = result.CandidateID
			}
			if object == "" {
				object = result.TaskID
			}
			if object == "" {
				object = result.WaitID
			}
		}
		if err = s.timelineAudit(r, TimelineOperationKind(kind), op, object, &base.Actor, nil, base.Reason); err != nil {
			return result, err
		}
	}
	if err = timelineBump(v, s.timelineTime()); err != nil {
		return result, err
	}
	result.SessionVersion = v.Version
	result.CurrentStageEventID = v.CurrentStageEventID
	return s.timelineCommit(r, TimelineOperationKind(kind), op, digest, result, false)
}
func (s *Store) timelineAccept(r *timelineRecord, requestID string, actor TimelineActor, kind, reason string, recover bool) (string, error) {
	v := r.Session
	for i, pending := range v.PendingJoinRequests {
		if pending.RequestID != requestID {
			continue
		}
		if err := s.timelineBind(r, pending.Source); err != nil {
			return "", err
		}
		id, err := timelineRandomID("producer-")
		if err != nil {
			return "", err
		}
		v.Producers = append(v.Producers, TimelineProducer{ProducerID: id, Source: pending.Source, JoinedAt: s.timelineTime()})
		v.PendingJoinRequests = append(v.PendingJoinRequests[:i], v.PendingJoinRequests[i+1:]...)
		if err = s.timelineAuditEvent(v, "producer_joined", actor, "", "", "", TimelineAudit{Operation: kind, Reason: reason, RequestID: requestID, NextProducerID: id}, true); err != nil {
			return "", err
		}
		if recover {
			previous := v.ForegroundProducerID
			v.ForegroundProducerID = id
			if err = s.timelineAuditEvent(v, "foreground_transferred", actor, "", "", "", TimelineAudit{Operation: kind, Reason: reason, RequestID: requestID, PreviousProducerID: previous, NextProducerID: id}, true); err != nil {
				return "", err
			}
		}
		return id, nil
	}
	return "", timelineFailure("pending_request_required", 409)
}

func validateTimelinePrivateAudit(r *timelineRecord) error {
	if len(r.Audit) > TimelineMaxReceipts+1 {
		return timelineFailure("corrupt_storage", 500)
	}
	ids, ops := map[string]bool{}, map[string]bool{}
	allowed := []string{"ticket_create", "ticket_patch", "delivery_create", "delivery_patch", "candidate_create", "create_session", "join_request", "task_create", "task_patch", "wait_create", "wait_attempt_start", "wait_observe", "session_patch", "event_patch", "event_withdraw", "document_add", "artifact_add", "session_delete"}
	for _, a := range r.Audit {
		if timelineID(a.AuditID, "audit_id") != nil || !contains(allowed, string(a.Operation)) || !timelineOperationID.MatchString(a.OperationID) || ids[a.AuditID] || ops[a.OperationID] || timelineReason(a.Reason) != nil || timelineTime(a.ReceivedAt, "received_at") != nil {
			return timelineFailure("corrupt_storage", 500)
		}
		ids[a.AuditID] = true
		ops[a.OperationID] = true
		matched := r.DeletePending && a.OperationID == r.DeleteOperation && a.Operation == "session_delete"
		for _, receipt := range r.Receipts {
			if receipt.OperationID == a.OperationID && receipt.Result.Kind == a.Operation {
				matched = true
			}
		}
		if !matched {
			return timelineFailure("corrupt_storage", 500)
		}
		if a.Actor == nil {
			if a.Operation != "join_request" || a.Source == nil || timelineSource(*a.Source) != nil {
				return timelineFailure("corrupt_storage", 500)
			}
		} else {
			if a.Actor.Validate() != nil {
				return timelineFailure("corrupt_storage", 500)
			}
			if a.Actor.Kind == "local_viewer" && a.Source != nil {
				return timelineFailure("corrupt_storage", 500)
			}
			if a.Actor.Kind == "native_producer" && (a.Source == nil || timelineSource(*a.Source) != nil) {
				return timelineFailure("corrupt_storage", 500)
			}
			if a.Actor.Kind == "native_producer" && r.Session != nil {
				found := false
				for _, p := range r.Session.Producers {
					found = found || p.ProducerID == a.Actor.ProducerID && a.Source != nil && *a.Source == p.Source
				}
				if !found {
					return timelineFailure("corrupt_storage", 500)
				}
			}
		}
	}
	return nil
}
