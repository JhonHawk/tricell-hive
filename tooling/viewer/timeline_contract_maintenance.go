package viewer

import (
	"encoding/json"
	"time"
)

const timelineMaintenancePolicyVersion = 1

// Private maintenance is separately bounded and never consumes publisher receipts.
// One retained receipt per administrative action permits exact retry, not history.
type timelineMaintenanceBaseline struct {
	Version   int64  `json:"version"`
	UpdatedAt string `json:"updated_at"`
	Archived  bool   `json:"archived"`
	Pinned    bool   `json:"pinned"`
}

type timelineMaintenance struct {
	ContentBaseline      *timelineMaintenanceBaseline   `json:"content_baseline,omitempty"`
	PolicyVersion        int                            `json:"policy_version"`
	EpisodeID            *string                        `json:"episode_id" timeline:"nullable"`
	ClosedAt             *string                        `json:"closed_at" timeline:"nullable"`
	EligibilityStartedAt *string                        `json:"eligibility_started_at" timeline:"nullable"`
	RestoreExempt        bool                           `json:"restore_exempt"`
	ArchiveOrigin        *string                        `json:"archive_origin" timeline:"nullable"`
	LastActivityAt       string                         `json:"last_activity_at"`
	AutomaticAction      *timelineAutomaticAction       `json:"automatic_action" timeline:"nullable"`
	ArchiveReceipt       *timelineAdministrativeReceipt `json:"archive_receipt" timeline:"nullable"`
	RestoreReceipt       *timelineAdministrativeReceipt `json:"restore_receipt" timeline:"nullable"`
	PinReceipt           *timelineAdministrativeReceipt `json:"pin_receipt" timeline:"nullable"`
}
type timelineAutomaticAction struct {
	EpisodeID        string `json:"episode_id"`
	CompletedAt      string `json:"completed_at"`
	CommittedVersion int64  `json:"committed_version"`
}
type timelineAdministrativeReceipt struct {
	Actor            TimelineActor           `json:"actor"`
	OperationID      string                  `json:"operation_id"`
	RequestSHA256    string                  `json:"request_sha256"`
	ExpectedVersion  int64                   `json:"expected_version"`
	CommittedVersion int64                   `json:"committed_version"`
	CompletedAt      string                  `json:"completed_at"`
	Result           TimelineOperationResult `json:"result"`
}

func validateTimelineLifecycleProjection(h timelineSummaryHeader) error {
	if err := timelineTime(h.LastActivityAt, "last_activity_at"); err != nil {
		return err
	}
	last, _ := time.Parse(time.RFC3339Nano, h.LastActivityAt)
	updated, _ := time.Parse(time.RFC3339Nano, h.UpdatedAt)
	if last.After(updated) {
		return timelineError("last_activity_at", "server_time_mismatch")
	}
	for _, f := range []struct {
		name  string
		value *string
	}{{"closed_at", h.ClosedAt}, {"autoarchive_at", h.AutoarchiveAt}} {
		if f.value != nil {
			if err := timelineTime(*f.value, f.name); err != nil {
				return err
			}
		}
	}
	if h.ClosedAt != nil {
		closed, _ := time.Parse(time.RFC3339Nano, *h.ClosedAt)
		if h.State != TimelineClosed || closed.After(updated) {
			return timelineError("closed_at", "server_time_mismatch")
		}
	}
	if h.AutoarchiveAt != nil && h.ClosedAt != nil {
		closed, _ := time.Parse(time.RFC3339Nano, *h.ClosedAt)
		deadline, _ := time.Parse(time.RFC3339Nano, *h.AutoarchiveAt)
		if !deadline.Equal(closed.Add(2 * time.Hour)) {
			return timelineError("autoarchive_at", "server_time_mismatch")
		}
	}
	if h.AutoarchiveAt != nil && (h.State != TimelineClosed || h.Archived || h.Pinned) {
		return timelineError("autoarchive_at", "ineligible_state")
	}
	if h.ArchiveOrigin != nil && (!h.Archived || !contains([]string{"manual", "automatic"}, *h.ArchiveOrigin)) {
		return timelineError("archive_origin", "invalid_origin")
	}
	if h.Archived && h.ArchiveOrigin == nil {
		return timelineError("archive_origin", "required")
	}
	return nil
}

func validateTimelineMaintenance(r *timelineRecord) error {
	m := r.Maintenance
	if m == nil {
		return nil
	}
	bad := func() error { return timelineFailure("corrupt_storage", 500) }
	raw, err := json.Marshal(m)
	if err != nil || len(raw)+len(`,"maintenance":`) > TimelineMaintenanceMaxBytes || m.PolicyVersion != timelineMaintenancePolicyVersion {
		return bad()
	}
	if timelineTime(m.LastActivityAt, "last_activity_at") != nil {
		return bad()
	}
	if b := m.ContentBaseline; b != nil {
		if timelineNumber(b.Version, "version") != nil || timelineTime(b.UpdatedAt, "updated_at") != nil {
			return bad()
		}
		if r.Session != nil {
			baseline, _ := time.Parse(time.RFC3339Nano, b.UpdatedAt)
			updated, _ := time.Parse(time.RFC3339Nano, r.Session.UpdatedAt)
			if baseline.After(updated) || b.Version > r.Session.Version {
				return bad()
			}
		}
	}
	if m.EpisodeID != nil && !timelineOpaqueID.MatchString(*m.EpisodeID) {
		return bad()
	}
	for _, at := range []*string{m.ClosedAt, m.EligibilityStartedAt} {
		if at != nil && timelineTime(*at, "maintenance") != nil {
			return bad()
		}
	}
	if (m.ClosedAt != nil || m.EligibilityStartedAt != nil || m.RestoreExempt) && m.EpisodeID == nil {
		return bad()
	}
	if m.ArchiveOrigin != nil && !contains([]string{"manual", "automatic"}, *m.ArchiveOrigin) {
		return bad()
	}
	if a := m.AutomaticAction; a != nil {
		if !timelineOpaqueID.MatchString(a.EpisodeID) || timelineTime(a.CompletedAt, "completed_at") != nil || timelineNumber(a.CommittedVersion, "committed_version") != nil {
			return bad()
		}
	}
	if r.Session != nil {
		updated, _ := time.Parse(time.RFC3339Nano, r.Session.UpdatedAt)
		for _, at := range []*string{m.ClosedAt, m.EligibilityStartedAt} {
			if at != nil {
				parsed, _ := time.Parse(time.RFC3339Nano, *at)
				if parsed.After(updated) {
					return bad()
				}
			}
		}
		if m.AutomaticAction != nil && m.AutomaticAction.CommittedVersion > r.Session.Version {
			return bad()
		}
	}
	seen := map[string]bool{}
	for _, receipt := range []*timelineAdministrativeReceipt{m.ArchiveReceipt, m.RestoreReceipt, m.PinReceipt} {
		if receipt == nil {
			continue
		}
		if receipt.Actor.Kind != "local_viewer" || receipt.Actor.ProducerID != "" || receipt.Actor.Validate() != nil || !timelineOperationID.MatchString(receipt.OperationID) || !timelineSHA256.MatchString(receipt.RequestSHA256) || timelineTime(receipt.CompletedAt, "completed_at") != nil || timelineNumber(receipt.ExpectedVersion, "expected_version") != nil || timelineNumber(receipt.CommittedVersion, "committed_version") != nil || receipt.CommittedVersion != receipt.ExpectedVersion+1 || receipt.Result.Validate() != nil || receipt.Result.Kind != "session_patch" || receipt.Result.OperationID != receipt.OperationID || receipt.Result.SessionVersion != receipt.CommittedVersion || receipt.Result.ViewerSessionID != r.ID || seen[receipt.OperationID] {
			return bad()
		}
		if r.Session != nil {
			completed, _ := time.Parse(time.RFC3339Nano, receipt.CompletedAt)
			updated, _ := time.Parse(time.RFC3339Nano, r.Session.UpdatedAt)
			if receipt.CommittedVersion > r.Session.Version || completed.After(updated) {
				return bad()
			}
		}
		seen[receipt.OperationID] = true
	}
	return nil
}

// Projection is the sole translation of private policy into public fields.
// Legacy reads use conservative defaults; maintenance initializes proved episode metadata
// under the store lock before making any eligibility decision.
func timelineProjectMaintenance(r *timelineRecord) {
	if r.Session == nil {
		return
	}
	v := r.Session
	v.LastActivityAt = v.CreatedAt
	v.ClosedAt = nil
	v.AutoarchiveAt = nil
	v.ArchiveOrigin = nil
	if v.Archived {
		v.ArchiveOrigin = timelineStringPointer("manual")
	}
	m := r.Maintenance
	if m == nil {
		return
	}
	v.LastActivityAt = m.LastActivityAt
	v.ClosedAt = m.ClosedAt
	v.ArchiveOrigin = m.ArchiveOrigin
	if v.State == TimelineClosed && !v.Archived && !v.Pinned && !m.RestoreExempt {
		at := m.ClosedAt
		if at == nil {
			at = m.EligibilityStartedAt
		}
		if at != nil {
			if start, err := time.Parse(time.RFC3339Nano, *at); err == nil {
				v.AutoarchiveAt = timelineStringPointer(start.Add(2 * time.Hour).UTC().Format(time.RFC3339Nano))
			}
		}
	}
}

// Derived fields are absent from persisted content and its unchanged budgets.
func timelineSessionContentBytes(v *TimelineSession) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	for _, key := range []string{"last_activity_at", "closed_at", "autoarchive_at", "archive_origin"} {
		delete(fields, key)
	}
	return json.Marshal(fields)
}
func (r timelineRecord) MarshalJSON() ([]byte, error) {
	type record timelineRecord
	raw, err := json.Marshal(record(r))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	if r.Session != nil {
		fields["session"], err = timelineSessionContentBytes(r.Session)
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(fields)
}

// Content is bounded independently so publishers cannot borrow policy capacity.
func timelineRecordContentBytes(r *timelineRecord) ([]byte, error) {
	copy := *r
	copy.Maintenance = nil
	return json.Marshal(copy)
}
