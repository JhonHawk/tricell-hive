package viewer

import "reflect"

func timelineRelatedTasks(v TimelineSession, eventID string) []string {
	ids := []string{}
	for _, e := range v.Events {
		if e.EventID == eventID || e.StageEventID == eventID || e.Target != nil && e.Target.EventID == eventID {
			ids = append(ids, e.TaskID)
			for _, snapshot := range e.DocumentSnapshots {
				ids = append(ids, snapshot.TaskID)
			}
		}
	}
	return ids
}
func (s *Store) timelineSetStage(v *TimelineSession, id string, state TimelineStageState, actor TimelineActor, reason string) error {
	for i := range v.Events {
		e := &v.Events[i]
		if e.EventID != id {
			continue
		}
		if e.Version >= TimelineMaxNumber {
			return timelineFailure("capacity", 413)
		}
		previous := e.StageState
		if previous == state {
			return nil
		}
		now := s.timelineTime()
		e.StageState = state
		e.Reason = reason
		e.Version++
		e.UpdatedAt = now
		e.FinishedAt = ""
		if state == "completed" || state == "failed" {
			e.FinishedAt = now
		}
		v.Reason = reason
		switch state {
		case "in_progress", "in_review", "awaiting_validation":
			v.CurrentStageEventID = &e.EventID
			if state == "in_progress" {
				v.State = TimelineActive
			} else {
				v.State = TimelineState(state)
			}
		default:
			if v.CurrentStageEventID != nil && *v.CurrentStageEventID == id {
				v.CurrentStageEventID = nil
			}
			switch state {
			case "completed":
				v.State = TimelineActive
			default:
				v.State = TimelinePaused
			}
			v.Reason = reason
		}
		return s.timelineAuditEvent(v, "state_transition", actor, id, "", "", TimelineAudit{Operation: "stage_update", Reason: reason, PreviousState: string(previous), NextState: string(state)}, true)
	}
	return ErrNotFound
}
func (s *Store) timelineSuspendOther(v *TimelineSession, except string, actor TimelineActor, reason string) error {
	if v.CurrentStageEventID == nil || *v.CurrentStageEventID == except {
		return nil
	}
	return s.timelineSetStage(v, *v.CurrentStageEventID, "paused", actor, reason)
}
func (s *Store) timelinePatchState(v *TimelineSession, request TimelineSessionPatchRequest) error {
	if request.State == nil || *request.State == v.State {
		return nil
	}
	next := *request.State
	previous := v.State
	stage := ""
	if next == "in_review" || next == TimelineAwaitingValidation {
		if v.CurrentStageEventID == nil {
			return timelineFailure("foreground_required", 409)
		}
	}
	if next == "failed" && v.CurrentStageEventID != nil {
		return timelineFailure("stage_active", 409)
	}
	if v.CurrentStageEventID != nil {
		stage = *v.CurrentStageEventID
		for i := range v.Events {
			e := &v.Events[i]
			if e.EventID != stage {
				continue
			}
			state := TimelineStageState(next)
			if next == TimelineActive {
				state = "in_progress"
			}
			if e.StageState != state {
				e.StageState = state
				e.Reason = request.Reason
				e.UpdatedAt = s.timelineTime()
				e.Version++
			}
		}
	}
	if next == TimelinePaused {
		v.CurrentStageEventID = nil
	}
	v.State = next
	v.Reason = request.Reason
	return s.timelineAuditEvent(v, "state_transition", request.Actor, stage, "", "", TimelineAudit{Operation: "session_patch", Reason: request.Reason, PreviousState: string(previous), NextState: string(next)}, true)
}
func (s *Store) timelineLocalAudit(v *TimelineSession, request TimelineSessionPatchRequest, old TimelineSession) error {
	changes := []struct {
		changed   bool
		operation string
	}{{old.Title != v.Title, "session_rename"}, {old.Archived != v.Archived, "session_archive"}, {old.Pinned != v.Pinned, "session_pin"}}
	for _, change := range changes {
		if !change.changed {
			continue
		}
		reason := request.Reason
		if reason == "" {
			reason = "Local viewer " + change.operation
		}
		if err := s.timelineAuditEvent(v, "local_action", request.Actor, "", "", "", TimelineAudit{Operation: change.operation, Reason: reason}, true); err != nil {
			return err
		}
	}
	return nil
}
func timelineSameEventCopy(a, b TimelineEvent) bool {
	return a.Title == b.Title && a.Summary == b.Summary && a.StageState == b.StageState
}
func timelineEqual(a, b any) bool { return reflect.DeepEqual(a, b) }

func timelineDocumentChainTasks(v TimelineSession, key string) []string {
	ids := []string{}
	revisions := map[string]bool{}
	for _, event := range v.Events {
		for _, snapshot := range event.DocumentSnapshots {
			if snapshot.DocumentKey == key {
				ids = append(ids, snapshot.TaskID)
				revisions[snapshot.DocumentRevisionID] = true
			}
		}
	}
	for _, event := range v.Events {
		if event.Target != nil && event.Target.Kind == "document" && revisions[event.Target.DocumentRevisionID] {
			ids = append(ids, event.TaskID)
		}
	}
	for _, task := range v.Tasks {
		if revisions[task.PlanRevisionID] {
			ids = append(ids, task.TaskID)
		}
	}
	return ids
}
