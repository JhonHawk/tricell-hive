package viewer

import (
	"strings"
)

// validateTimelineSerializedClosureReserve projects the necessary terminal
// operations from the complete candidate, including its ledger and receipts.
// Identifiers/timestamps below are size placeholders only: they are never
// issued or persisted. Required human text uses its maximum JSON-escaped size;
// optional closure references are omitted, so an unreferenced closure remains
// available even when content is full.
func validateTimelineSerializedClosureReserve(candidate *timelineRecord) error {
	if candidate.Session == nil || candidate.Session.State == TimelineClosed {
		return nil
	}
	projected := timelineClone(candidate)
	v := projected.Session
	actor := TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}
	reason := strings.Repeat("<", 500)
	const timestamp = "9999-12-31T23:59:59.999999999Z"
	id := func(letter string) string { return "event-" + strings.Repeat(letter, 32) }
	appendAudit := func(operation, stage, previous, next, letter string) {
		v.Events = append(v.Events, TimelineEvent{EventID: id(letter), Sequence: int64(len(v.Events) + 1), Version: 1, Kind: "state_transition", Actor: actor, StageEventID: stage, Title: strings.Repeat("<", 200), Summary: reason, CreatedAt: timestamp, UpdatedAt: timestamp, Origin: "server", TicketRefs: []TimelineTicketRef{}, Audit: &TimelineAudit{Operation: operation, Reason: reason, PreviousState: previous, NextState: next}})
	}
	appendReceipt := func(kind TimelineOperationKind, letter string, event *TimelineEvent) error {
		if v.Version >= TimelineMaxNumber {
			return timelineFailure("capacity", 413)
		}
		v.Version++
		v.UpdatedAt = timestamp
		op := strings.Repeat(letter, 128)
		result := timelineResult(projected, kind, op)
		if event != nil {
			result.EventID = event.EventID
			result.EventVersion = event.Version
			result.Sequence = event.Sequence
		}
		projected.Receipts = append(projected.Receipts, TimelineOperationReceipt{OperationID: op, RequestSHA256: strings.Repeat("0", 64), CompletedAt: timestamp, Result: result})
		return timelineProjectedTerminalBytes(projected)
	}
	if v.CurrentStageEventID != nil {
		stageID := *v.CurrentStageEventID
		found := false
		for i := range v.Events {
			e := &v.Events[i]
			if e.EventID != stageID {
				continue
			}
			found = true
			if e.Version >= TimelineMaxNumber {
				return timelineFailure("capacity", 413)
			}
			previous := e.StageState
			e.StageState = "completed"
			e.Reason = reason
			e.Version++
			e.UpdatedAt = timestamp
			e.FinishedAt = timestamp
			snapshot := *e
			v.CurrentStageEventID = nil
			v.State = TimelineActive
			v.Reason = reason
			// A stage completion may invalidate linked task snapshots. Account for
			// their counter growth too; global pause is an alternative release path.
			related := map[string]bool{}
			for _, taskID := range timelineRelatedTasks(*v, stageID) {
				related[taskID] = true
			}
			for j := range v.Tasks {
				if related[v.Tasks[j].TaskID] && v.Tasks[j].TaskVersion < TimelineMaxNumber {
					v.Tasks[j].TaskVersion++
				}
			}
			appendAudit("stage_update", stageID, string(previous), "completed", "a")
			if err := appendReceipt("event_patch", "r", &snapshot); err != nil {
				return err
			}
			break
		}
		if !found {
			return timelineFailure("corrupt_storage", 500)
		}
	} else if v.State == TimelineFailed {
		v.State = TimelineActive
		v.Reason = reason
		appendAudit("session_patch", "", string(TimelineFailed), string(TimelineActive), "b")
		if err := appendReceipt("session_patch", "r", nil); err != nil {
			return err
		}
	}
	previous := v.State
	closure := TimelineEvent{EventID: id("c"), Sequence: int64(len(v.Events) + 1), Version: 1, Kind: "closure", Actor: actor, Reason: reason, Title: strings.Repeat("<", 200), Summary: strings.Repeat("<", 4000), CreatedAt: timestamp, UpdatedAt: timestamp, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}, Closure: &TimelineClosure{Reason: "completed", Result: strings.Repeat("<", 4000)}}
	v.Events = append(v.Events, closure)
	v.State = TimelineClosed
	v.Reason = reason
	appendAudit("closure", "", string(previous), string(TimelineClosed), "d")
	return appendReceipt("append_event", "q", &closure)
}

func timelineProjectedTerminalBytes(r *timelineRecord) error {
	if len(r.Receipts) > TimelineMaxReceipts || len(r.Session.Events) > TimelineMaxEvents {
		return timelineFailure("capacity", 413)
	}
	public, err := timelineSessionContentBytes(r.Session)
	if err != nil {
		return err
	}
	if len(public) > TimelineMaxBody {
		return timelineFailure("capacity", 413)
	}
	private, err := timelineRecordContentBytes(r)
	if err != nil {
		return err
	}
	if len(private) > timelineMaxRecordBytes {
		return timelineFailure("capacity", 413)
	}
	return nil
}
