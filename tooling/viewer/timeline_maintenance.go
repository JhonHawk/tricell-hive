package viewer

import (
	"encoding/json"
	"errors"
	"time"
)

func timelineLater(a, b string) string {
	at, _ := time.Parse(time.RFC3339Nano, a)
	bt, _ := time.Parse(time.RFC3339Nano, b)
	if bt.After(at) {
		return b
	}
	return a
}

// Reconstruct only server receipt times. Event sequence establishes episode order;
// client occurrence dates and filesystem times cannot establish closure.
func timelineRefreshMaintenance(r *timelineRecord, now string) {
	v := r.Session
	if v == nil {
		return
	}
	m := r.Maintenance
	fresh := m == nil
	if fresh {
		m = &timelineMaintenance{PolicyVersion: timelineMaintenancePolicyVersion, LastActivityAt: v.CreatedAt}
		r.Maintenance = m
	}
	last := m.LastActivityAt
	var episode *TimelineEvent
	var episodeBoundary int64
	for i := range v.Events {
		e := &v.Events[i]
		if fresh && e.Actor.Kind == "native_producer" && !contains([]string{"producer_joined", "foreground_transferred", "local_action"}, string(e.Kind)) {
			last = timelineLater(last, e.CreatedAt)
			last = timelineLater(last, e.UpdatedAt)
		}
		if e.Kind == "state_transition" && e.Audit != nil && e.Audit.NextState == string(TimelineClosed) && e.StageEventID == "" {
			episode = e
			episodeBoundary = e.Sequence
		} else if e.Kind == "state_transition" && e.Audit != nil && e.StageEventID == "" && e.Audit.Operation == "session_patch" {
			episode = nil
			episodeBoundary = e.Sequence
		} else if e.Kind == "closure" {
			episode = e
			episodeBoundary = e.Sequence
		}
	}
	// A private legacy session_patch receipt does not retain its old payload,
	// so it cannot distinguish a summary change from an unchanged request.
	// Explicit lifecycle changes remain proved by their public transitions.
	for _, a := range r.Audit {
		if fresh && a.Actor != nil && a.Actor.Kind == "native_producer" && !contains([]string{"create_session", "join_request", "producer_accept", "foreground_transfer", "recover_lead", "session_delete", "session_patch"}, string(a.Operation)) {
			last = timelineLater(last, a.ReceivedAt)
		}
	}
	m.LastActivityAt = last
	// Some historic operations sampled their audit timestamp after UpdatedAt.
	// Preserve truthful activity and advance general modification time as needed.
	v.UpdatedAt = timelineLater(v.UpdatedAt, last)
	if v.State != TimelineClosed {
		if m.ArchiveOrigin != nil && *m.ArchiveOrigin == "automatic" {
			v.Archived = false
			m.ArchiveOrigin = nil
		}
		m.EpisodeID = nil
		m.ClosedAt = nil
		m.EligibilityStartedAt = nil
		m.RestoreExempt = false
	} else {
		id := "episode-unknown"
		var closed *string
		if episode != nil {
			id = episode.EventID
			closed = timelineStringPointer(episode.CreatedAt)
		}
		if m.EpisodeID == nil || *m.EpisodeID != id {
			m.EpisodeID = timelineStringPointer(id)
			m.ClosedAt = closed
			m.RestoreExempt = false
			m.EligibilityStartedAt = nil
			if closed == nil {
				m.EligibilityStartedAt = timelineStringPointer(now)
				v.UpdatedAt = timelineLater(v.UpdatedAt, now)
			}
		}
		if fresh {
			// Preserve a retained restore even when the closure date is unknown.
			// A later explicit lifecycle transition invalidates older choices;
			// absence of the original closure event does not revoke a choice.
			start := episodeBoundary
			for _, e := range v.Events {
				if e.Sequence > start && e.Kind == "local_action" && e.Audit != nil && e.Audit.Operation == "session_archive" && e.Actor.Kind == "local_viewer" {
					m.RestoreExempt = !v.Archived
				}
			}
		}
	}
	if v.Archived && m.ArchiveOrigin == nil {
		m.ArchiveOrigin = timelineStringPointer("manual")
	}
	if !v.Archived {
		m.ArchiveOrigin = nil
	}
	if m.ClosedAt != nil {
		v.UpdatedAt = timelineLater(v.UpdatedAt, *m.ClosedAt)
	}
	timelineProjectMaintenance(r)
}

func (s *Store) timelineMaintainLocked() error {
	if err := s.timelineReconcile(); err != nil {
		return err
	}
	var failures []error
	for _, old := range s.timeline {
		if old.Session == nil || old.DeletePending {
			continue
		}
		r := timelineClone(old)
		baseline := timelineScalarBaseline(r.Session)
		now := s.timelineTime()
		timelineRefreshMaintenance(r, now)
		changed := !timelineEqual(old.Maintenance, r.Maintenance) || !timelineEqual(old.Session, r.Session)
		due := false
		if at := r.Session.AutoarchiveAt; at != nil {
			deadline, _ := time.Parse(time.RFC3339Nano, *at)
			current, _ := time.Parse(time.RFC3339Nano, now)
			due = !current.Before(deadline)
		}
		if !changed && !due {
			continue
		}
		if r.Maintenance.ContentBaseline == nil {
			r.Maintenance.ContentBaseline = baseline
		}
		// Reserve a version for manual reversal. Never archive a record which the
		// viewer could not subsequently restore because its counter is exhausted.
		if due && r.Session.Version >= TimelineMaxNumber-1 {
			failures = append(failures, timelineFailure("capacity", 413))
			continue
		}
		if err := timelineBump(r.Session, timelineLater(r.Session.UpdatedAt, now)); err != nil {
			failures = append(failures, err)
			continue
		}
		if due {
			r.Session.Archived = true
			r.Maintenance.ArchiveOrigin = timelineStringPointer("automatic")
			r.Maintenance.AutomaticAction = &timelineAutomaticAction{EpisodeID: *r.Maintenance.EpisodeID, CompletedAt: now, CommittedVersion: r.Session.Version}
		}
		if err := s.timelineSave(r); err != nil {
			failures = append(failures, err)
			if s.timelineUncertain {
				break
			}
		}
	}
	return errors.Join(failures...)
}

func timelineAdministrativePatch(request TimelineSessionPatchRequest) bool {
	return request.Actor.Kind == "local_viewer" && request.Title == nil && request.Summary == nil && request.State == nil && (request.Archived != nil || request.Pinned != nil)
}
func (s *Store) timelineAdministrativeCommit(r *timelineRecord, request TimelineSessionPatchRequest, digest string) (TimelineOperationResult, error) {
	now := timelineLater(r.Session.UpdatedAt, s.timelineTime())
	baseline := timelineScalarBaseline(r.Session)
	timelineRefreshMaintenance(r, now)
	if r.Maintenance.ContentBaseline == nil {
		r.Maintenance.ContentBaseline = baseline
	}
	if request.Archived != nil {
		r.Session.Archived = *request.Archived
	}
	if request.Pinned != nil {
		r.Session.Pinned = *request.Pinned
	}
	if request.Archived != nil {
		if *request.Archived {
			r.Maintenance.ArchiveOrigin = timelineStringPointer("manual")
		} else {
			r.Maintenance.ArchiveOrigin = nil
			if r.Session.State == TimelineClosed {
				r.Maintenance.RestoreExempt = true
			}
		}
	}
	if r.Session.Archived && r.Session.Version >= TimelineMaxNumber-1 {
		return TimelineOperationResult{}, timelineFailure("capacity", 413)
	}
	if err := timelineBump(r.Session, now); err != nil {
		return TimelineOperationResult{}, err
	}
	result := timelineResult(r, "session_patch", request.OperationID)
	receipt := &timelineAdministrativeReceipt{Actor: request.Actor.recorded(), OperationID: request.OperationID, RequestSHA256: digest, ExpectedVersion: request.ExpectedVersion, CommittedVersion: r.Session.Version, CompletedAt: now, Result: result}
	if request.Archived != nil {
		if *request.Archived {
			r.Maintenance.ArchiveReceipt = receipt
		} else {
			r.Maintenance.RestoreReceipt = receipt
		}
	} else {
		r.Maintenance.PinReceipt = receipt
	}
	return result, s.timelineSave(r)
}

// Existing canonical activity advances only on new substantive receipts. Generic
// session patches and event patches can commit unchanged results; exclude those.
func timelineCommittedActivity(r, old *timelineRecord, kind TimelineOperationKind) {
	if old == nil || old.Session == nil || r.Maintenance == nil || contains([]string{"create_session", "join_request", "producer_accept", "foreground_transfer", "recover_lead", "session_delete"}, string(kind)) {
		return
	}
	if timelineUnchangedDomainContent(r.Session, old.Session, kind) {
		return
	}
	if kind == "session_patch" && r.Session.State == old.Session.State && r.Session.Summary == old.Session.Summary {
		return
	}
	if kind == "event_patch" && timelineEqual(r.Session.Events, old.Session.Events) {
		return
	}
	if kind == "session_patch" && r.Session.State == old.Session.State && len(r.Audit) == len(old.Audit) {
		return
	}
	at := r.Session.UpdatedAt
	for _, a := range r.Audit[len(old.Audit):] {
		at = timelineLater(at, a.ReceivedAt)
	}
	for _, e := range r.Session.Events[len(old.Session.Events):] {
		at = timelineLater(at, e.CreatedAt)
	}
	r.Maintenance.LastActivityAt = timelineLater(r.Maintenance.LastActivityAt, at)
	r.Session.UpdatedAt = timelineLater(r.Session.UpdatedAt, r.Maintenance.LastActivityAt)
	timelineProjectMaintenance(r)
}

func timelineScalarBaseline(v *TimelineSession) *timelineMaintenanceBaseline {
	return &timelineMaintenanceBaseline{Version: v.Version, UpdatedAt: v.UpdatedAt, Archived: v.Archived, Pinned: v.Pinned}
}

// Only the four administrative scalars can grow into policy headroom. Ordinary
// publication clears the baseline before its existing content admission checks.
// Normalizing these fields for accounting never changes the truthful stored data.
func timelineMaintenanceContentBytes(r *timelineRecord) ([]byte, int, error) {
	actual, err := timelineRecordContentBytes(r)
	if err != nil || r.Session == nil || r.Maintenance == nil || r.Maintenance.ContentBaseline == nil {
		return actual, 0, err
	}
	copy := *r
	session := *r.Session
	copy.Session = &session
	b := r.Maintenance.ContentBaseline
	session.Version = b.Version
	session.UpdatedAt = b.UpdatedAt
	session.Archived = b.Archived
	session.Pinned = b.Pinned
	baseline, err := timelineRecordContentBytes(&copy)
	if err != nil {
		return nil, 0, err
	}
	body, err := timelineSessionContentBytes(&session)
	if err != nil {
		return nil, 0, err
	}
	if len(body) > TimelineMaxBody {
		return nil, 0, timelineFailure("capacity", 413)
	}
	growth := max(0, len(actual)-len(baseline))
	metadata, _ := json.Marshal(r.Maintenance)
	if len(metadata)+len(`,"maintenance":`)+growth > TimelineMaintenanceMaxBytes {
		return nil, 0, timelineFailure("capacity", 413)
	}
	return baseline, growth, nil
}

// Domain collections contain the accepted work effect. A receipt/private audit,
// a general modification timestamp and derived counters alone are not activity.
// Keep every business field, including evidence timestamps and provenance: a new
// observation, attempt, plan binding or catalog value remains substantive.
func timelineUnchangedDomainContent(next, previous *TimelineSession, kind TimelineOperationKind) bool {
	switch kind {
	case "ticket_create", "ticket_patch", "delivery_create", "delivery_patch", "candidate_create", "task_create", "task_patch", "task_transition", "wait_create", "wait_attempt_start", "wait_observe":
	default:
		return false
	}
	type content struct {
		Tickets    []TimelineTicket
		Deliveries []TimelineDelivery
		Candidates []TimelineCandidateRecord
		Tasks      []TimelineTask
		Waits      []TimelineCIWait
	}
	normalize := func(v *TimelineSession) content {
		tasks := append([]TimelineTask{}, v.Tasks...)
		for i := range tasks {
			tasks[i].TaskVersion = 0
		}
		waits := append([]TimelineCIWait{}, v.Waits...)
		for i := range waits {
			waits[i].WaitVersion = 0
		}
		return content{v.Tickets, v.Deliveries, v.Candidates, tasks, waits}
	}
	return timelineEqual(normalize(next), normalize(previous))
}

// Inventory corruption remains fatal, including when it emerges during uncertain
// write reconciliation. Capacity and physical maintenance failures are retried.
func timelineMaintenanceCorrupt(err error) bool {
	if err == nil {
		return false
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, child := range joined.Unwrap() {
			if timelineMaintenanceCorrupt(child) {
				return true
			}
		}
		return false
	}
	if wrapped, ok := err.(interface{ Unwrap() error }); ok {
		return timelineMaintenanceCorrupt(wrapped.Unwrap())
	}
	var failure *TimelineError
	if errors.As(err, &failure) && (failure.Code == "corrupt_storage" || failure.Code == "invalid_storage") {
		return true
	}
	var contract *TimelineContractError
	return errors.As(err, &contract)
}
