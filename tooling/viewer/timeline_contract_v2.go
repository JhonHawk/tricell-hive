package viewer

// This file defines v2 wire shapes and pure validation. It performs no storage,
// networking or authority mutation. Call graph validators under the store lock.
import (
	"regexp"
	"time"
)

const (
	TimelineMaxProducers        = 16
	TimelineMaxJoinRequests     = 16
	TimelineMaxTasks            = 100
	TimelineMaxTaskSegments     = 5
	TimelineMaxBlockSteps       = 10
	TimelineMaxRequiredChecks   = 50
	TimelineMaxWaitAttempts     = 20
	TimelineMaxWaitObservations = 500
	TimelineMaxWaits            = 50
	TimelineMaxWaitText         = 500
)

type TimelineActor struct {
	Kind       string `json:"kind"`
	ProducerID string `json:"producer_id,omitempty"`
}

func (a TimelineActor) Validate() error {
	switch a.Kind {
	case "native_producer":
		return timelineID(a.ProducerID, "actor.producer_id")
	case "local_viewer":
		if a.ProducerID == "" {
			return nil
		}
	}
	return timelineError("actor", "invalid_actor")
}
func validateTimelineNativeActor(a TimelineActor) error {
	if err := a.Validate(); err != nil {
		return err
	}
	if a.Kind != "native_producer" {
		return timelineError("actor", "native_producer_required")
	}
	return nil
}

type TimelineProducer struct {
	ProducerID string         `json:"producer_id"`
	Source     TimelineSource `json:"source"`
	JoinedAt   string         `json:"joined_at"`
}
type TimelineJoinRequest struct {
	RequestID   string         `json:"request_id"`
	Source      TimelineSource `json:"source"`
	RequestedAt string         `json:"requested_at"`
	Reason      string         `json:"reason"`
}
type TimelineAudit struct {
	Block              *TimelineBlock          `json:"block,omitempty"`
	Resolution         string                  `json:"resolution,omitempty"`
	PreviousSegment    *TimelineTaskSegment    `json:"previous_segment,omitempty"`
	NextSegment        *TimelineTaskSegment    `json:"next_segment,omitempty"`
	PlanRevisionID     string                  `json:"plan_revision_id,omitempty"`
	Operation          string                  `json:"operation"`
	Reason             string                  `json:"reason"`
	PreviousState      string                  `json:"previous_state,omitempty"`
	NextState          string                  `json:"next_state,omitempty"`
	PreviousProducerID string                  `json:"previous_producer_id,omitempty"`
	NextProducerID     string                  `json:"next_producer_id,omitempty"`
	RequestID          string                  `json:"request_id,omitempty"`
	Target             *TimelineDecisionTarget `json:"target,omitempty"`
}
type TimelineBlock struct {
	Reason     string   `json:"reason"`
	Steps      []string `json:"steps"`
	Owner      string   `json:"owner"`
	ResumeWhen string   `json:"resume_when"`
}
type TimelineTaskSegment struct {
	SegmentID           string         `json:"segment_id"`
	PlannedStage        TimelineStage  `json:"planned_stage"`
	Order               int64          `json:"order"`
	State               string         `json:"state"`
	AssignedProducerID  string         `json:"assigned_producer_id"`
	DisplayStageEventID string         `json:"display_stage_event_id,omitempty"`
	Block               *TimelineBlock `json:"block,omitempty"`
	ReviewKind          string         `json:"review_kind,omitempty"`
	Withdrawn           bool           `json:"withdrawn,omitempty"`
}
type TimelineTask struct {
	TaskID         string                `json:"task_id"`
	TaskKey        string                `json:"task_key"`
	Title          string                `json:"title"`
	Summary        string                `json:"summary"`
	TaskVersion    int64                 `json:"task_version"`
	PlanRevisionID string                `json:"plan_revision_id,omitempty"`
	Segments       []TimelineTaskSegment `json:"segments"`
	State          string                `json:"state"`
	Withdrawn      bool                  `json:"withdrawn,omitempty"`
}

// Creation never accepts output IDs, versions, aggregate state or history.
type TimelineTaskSegmentInput struct {
	PlannedStage       TimelineStage `json:"planned_stage"`
	Order              int64         `json:"order"`
	AssignedProducerID string        `json:"assigned_producer_id,omitempty"`
}
type TimelineNativeOperation struct {
	SchemaVersion          int           `json:"schema_version"`
	OperationID            string        `json:"operation_id"`
	Actor                  TimelineActor `json:"actor"`
	ExpectedSessionVersion int64         `json:"expected_session_version"`
	Reason                 string        `json:"reason"`
}
type TimelineJoinRequestInput struct {
	SchemaVersion          int            `json:"schema_version"`
	OperationID            string         `json:"operation_id"`
	Source                 TimelineSource `json:"source"`
	ExpectedSessionVersion int64          `json:"expected_session_version"`
	Reason                 string         `json:"reason"`
}
type TimelineProducerAcceptRequest struct {
	TimelineNativeOperation
	RequestID string `json:"request_id"`
}
type TimelineForegroundTransferRequest struct {
	TimelineNativeOperation
	TargetProducerID string `json:"target_producer_id"`
}
type TimelineRecoverLeadRequest struct {
	TimelineNativeOperation
	RequestID string `json:"request_id"`
}
type TimelineTaskCreateRequest struct {
	TimelineNativeOperation
	TaskKey        string                     `json:"task_key"`
	Title          string                     `json:"title"`
	Summary        string                     `json:"summary"`
	PlanRevisionID string                     `json:"plan_revision_id,omitempty"`
	Segments       []TimelineTaskSegmentInput `json:"segments"`
}
type TimelineTaskPatchRequest struct {
	TimelineNativeOperation
	ExpectedTaskVersion int64                   `json:"expected_task_version"`
	Title               *string                 `json:"title,omitempty"`
	Summary             *string                 `json:"summary,omitempty"`
	PlanRevisionID      string                  `json:"plan_revision_id,omitempty"`
	DecisionTarget      *TimelineDecisionTarget `json:"decision_target,omitempty"`
	SegmentID           string                  `json:"segment_id,omitempty"`
	Order               *int64                  `json:"order,omitempty"`
	AssignedProducerID  string                  `json:"assigned_producer_id,omitempty"`
	DisplayStageEventID string                  `json:"display_stage_event_id,omitempty"`
	Withdrawn           *bool                   `json:"withdrawn,omitempty"`
}
type TimelineTaskTransitionRequest struct {
	TimelineNativeOperation
	ExpectedTaskVersion int64          `json:"expected_task_version"`
	SegmentID           string         `json:"segment_id"`
	State               string         `json:"state"`
	Block               *TimelineBlock `json:"block,omitempty"`
	ReviewKind          string         `json:"review_kind,omitempty"`
	Resolution          string         `json:"resolution,omitempty"`
}

type TimelineRequiredCheck struct {
	CheckKey string `json:"check_key"`
	Label    string `json:"label"`
}
type TimelineCheckPolicy struct {
	PolicyID       string                  `json:"policy_id"`
	Confirmed      bool                    `json:"confirmed"`
	Authority      string                  `json:"authority"`
	SourceURL      string                  `json:"source_url"`
	ObservedAt     string                  `json:"observed_at"`
	RequiredChecks []TimelineRequiredCheck `json:"required_checks"`
}
type TimelineCIExternalRun struct {
	Commit string `json:"commit"`
	RunID  string `json:"run_id"`
	URL    string `json:"url"`
}
type TimelineCIAttempt struct {
	AttemptID   string                `json:"attempt_id"`
	Generation  int64                 `json:"generation"`
	CheckKeys   []string              `json:"check_keys"`
	ExternalRun TimelineCIExternalRun `json:"external_run"`
	SourceURL   string                `json:"source_url"`
	ObservedAt  string                `json:"observed_at"`
	Actor       TimelineActor         `json:"actor"`
}
type TimelineCIObservation struct {
	ObservationID string                `json:"observation_id"`
	AttemptID     string                `json:"attempt_id"`
	CheckKey      string                `json:"check_key"`
	ExternalRun   TimelineCIExternalRun `json:"external_run"`
	State         string                `json:"state"`
	Summary       string                `json:"summary"`
	ObservedAt    string                `json:"observed_at"`
	ReceivedAt    string                `json:"received_at"`
	Actor         TimelineActor         `json:"actor"`
}
type TimelineCIWait struct {
	WaitID              string                  `json:"wait_id"`
	WaitVersion         int64                   `json:"wait_version"`
	StageEventID        string                  `json:"stage_event_id"`
	TaskID              string                  `json:"task_id,omitempty"`
	SegmentID           string                  `json:"segment_id,omitempty"`
	Candidate           TimelineCandidate       `json:"candidate"`
	PullRequestURL      string                  `json:"pull_request_url"`
	Policy              TimelineCheckPolicy     `json:"policy"`
	State               string                  `json:"state"`
	NextAction          string                  `json:"next_action,omitempty"`
	PreviousWaitID      string                  `json:"previous_wait_id,omitempty"`
	ApplicableAttemptID map[string]string       `json:"applicable_attempt_id"`
	Attempts            []TimelineCIAttempt     `json:"attempts"`
	Observations        []TimelineCIObservation `json:"observations"`
}
type TimelineWaitCreateRequest struct {
	TimelineNativeOperation
	TargetStageEventID string              `json:"target_stage_event_id"`
	TaskID             string              `json:"task_id,omitempty"`
	SegmentID          string              `json:"segment_id,omitempty"`
	Candidate          TimelineCandidate   `json:"candidate"`
	PullRequestURL     string              `json:"pull_request_url"`
	Policy             TimelineCheckPolicy `json:"policy"`
	PreviousWaitID     string              `json:"previous_wait_id,omitempty"`
	NextAction         string              `json:"next_action,omitempty"`
}
type TimelineWaitAttemptStartRequest struct {
	TimelineNativeOperation
	ExpectedWaitVersion int64                 `json:"expected_wait_version"`
	CheckKeys           []string              `json:"check_keys"`
	ExternalRun         TimelineCIExternalRun `json:"external_run"`
	SourceURL           string                `json:"source_url"`
	ObservedAt          string                `json:"observed_at"`
}
type TimelineWaitObserveRequest struct {
	TimelineNativeOperation
	ExpectedWaitVersion int64                 `json:"expected_wait_version"`
	AttemptID           string                `json:"attempt_id"`
	CheckKey            string                `json:"check_key"`
	ExternalRun         TimelineCIExternalRun `json:"external_run"`
	State               string                `json:"state"`
	Summary             string                `json:"summary"`
	ObservedAt          string                `json:"observed_at"`
}

var timelineTaskKey = regexp.MustCompile(`^T[1-9][0-9]{0,126}$`)
var timelineDocumentKey = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var timelineTaskStates = []string{"not_started", "in_progress", "in_review", "blocked", "completed"}
var timelineWaitStates = []string{"waiting", "succeeded", "failed", "cancelled", "unknown"}

func timelineReason(s string) error { return timelineText(s, "reason", 500, true) }
func (r TimelineNativeOperation) validate(local bool) error {
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := r.Actor.Validate(); err != nil {
		return err
	}
	if local != (r.Actor.Kind == "local_viewer") {
		return timelineError("actor", "forbidden_actor")
	}
	if err := timelineNumber(r.ExpectedSessionVersion, "expected_session_version"); err != nil {
		return err
	}
	return timelineReason(r.Reason)
}
func (b TimelineBlock) Validate() error {
	if err := timelineReason(b.Reason); err != nil {
		return err
	}
	if len(b.Steps) < 1 || len(b.Steps) > TimelineMaxBlockSteps {
		return timelineError("steps", "invalid_count")
	}
	for _, step := range b.Steps {
		if err := timelineText(step, "steps", 500, true); err != nil {
			return err
		}
	}
	if err := timelineText(b.Owner, "owner", 200, true); err != nil {
		return err
	}
	return timelineText(b.ResumeWhen, "resume_when", 500, true)
}
func validateTimelineSessionPatchActor(r TimelineSessionPatchRequest) error {
	if err := r.Actor.Validate(); err != nil {
		return err
	}
	if r.Actor.Kind == "local_viewer" {
		if r.Summary != nil || r.State != nil || r.Tickets != nil {
			return timelineError("actor", "forbidden_actor")
		}
	} else if r.Title != nil || r.Archived != nil || r.Pinned != nil {
		return timelineError("actor", "local_viewer_required")
	}
	if r.State != nil {
		return timelineReason(r.Reason)
	}
	if r.Reason != "" {
		return timelineReason(r.Reason)
	}
	return nil
}
func validateTimelineAppendV2(r TimelineAppendEventRequest) error {
	if err := validateTimelineNativeActor(r.Actor); err != nil {
		return err
	}
	if contains([]string{"state_transition", "producer_joined", "foreground_transferred", "task_state_transition", "local_action"}, string(r.Kind)) {
		return timelineError("kind", "server_only")
	}
	if r.Kind == "stage" || r.Kind == "closure" {
		if r.TargetStageEventID != "" || r.TaskID != "" || r.SegmentID != "" {
			return timelineError("target_stage_event_id", "forbidden_for_kind")
		}
		if err := timelineReason(r.Reason); err != nil {
			return err
		}
	}
	if r.TargetStageEventID != "" {
		if err := timelineID(r.TargetStageEventID, "target_stage_event_id"); err != nil {
			return err
		}
	}
	if (r.TaskID == "") != (r.SegmentID == "") {
		return timelineError("task_id", "task_segment_pair_required")
	}
	if r.TaskID != "" {
		if err := timelineID(r.TaskID, "task_id"); err != nil {
			return err
		}
		if err := timelineID(r.SegmentID, "segment_id"); err != nil {
			return err
		}
		if r.TargetStageEventID == "" {
			return timelineError("target_stage_event_id", "required_for_task")
		}
	}
	if r.Kind == "review" {
		if !contains([]string{"in_progress", "completed", "failed"}, r.ReviewStatus) {
			return timelineError("review_status", "invalid_state")
		}
		if r.TargetStageEventID == "" {
			return timelineError("target_stage_event_id", "required_for_review")
		}
	} else if r.ReviewStatus != "" {
		return timelineError("review_status", "forbidden_for_kind")
	}
	return nil
}
func validateTimelineSnapshotV2(r TimelineDocumentSnapshotRequest) error {
	if err := validateTimelineNativeActor(r.Actor); err != nil {
		return err
	}
	if !timelineDocumentKey.MatchString(r.DocumentKey) {
		return timelineError("document_key", "invalid_key")
	}
	if r.PredecessorRevisionID != "" {
		if err := timelineID(r.PredecessorRevisionID, "predecessor_revision_id"); err != nil {
			return err
		}
	}
	if err := timelineNumber(r.ExpectedSessionVersion, "expected_session_version"); err != nil {
		return err
	}
	if (r.TaskID == "") != (r.SegmentID == "") {
		return timelineError("task_id", "task_segment_pair_required")
	}
	return nil
}

// DecodeTimelineV2Operation uses closed DTOs, never an open map of payload fields.
func DecodeTimelineV2Operation(kind string, raw []byte) (any, error) {
	var target any
	switch kind {
	case "join_request":
		target = &TimelineJoinRequestInput{}
	case "producer_accept":
		target = &TimelineProducerAcceptRequest{}
	case "foreground_transfer":
		target = &TimelineForegroundTransferRequest{}
	case "recover_lead":
		target = &TimelineRecoverLeadRequest{}
	case "task_create":
		target = &TimelineTaskCreateRequest{}
	case "task_patch":
		target = &TimelineTaskPatchRequest{}
	case "task_transition":
		target = &TimelineTaskTransitionRequest{}
	case "wait_create":
		target = &TimelineWaitCreateRequest{}
	case "wait_attempt_start":
		target = &TimelineWaitAttemptStartRequest{}
	case "wait_observe":
		target = &TimelineWaitObserveRequest{}
	default:
		return nil, timelineError("kind", "unsupported_operation")
	}
	if err := DecodeTimelineJSON(raw, target); err != nil {
		return nil, err
	}
	return target, ValidateTimelineV2Operation(target)
}
func ValidateTimelineV2Operation(value any) error {
	var base TimelineNativeOperation
	switch r := value.(type) {
	case *TimelineJoinRequestInput:
		if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
			return err
		}
		if err := timelineSource(r.Source); err != nil {
			return err
		}
		if err := timelineNumber(r.ExpectedSessionVersion, "expected_session_version"); err != nil {
			return err
		}
		return timelineReason(r.Reason)
	case *TimelineProducerAcceptRequest:
		base = r.TimelineNativeOperation
		if err := timelineID(r.RequestID, "request_id"); err != nil {
			return err
		}
	case *TimelineForegroundTransferRequest:
		base = r.TimelineNativeOperation
		if err := timelineID(r.TargetProducerID, "target_producer_id"); err != nil {
			return err
		}
	case *TimelineRecoverLeadRequest:
		if err := timelineID(r.RequestID, "request_id"); err != nil {
			return err
		}
		return r.TimelineNativeOperation.validate(true)
	case *TimelineTaskCreateRequest:
		base = r.TimelineNativeOperation
		if !timelineTaskKey.MatchString(r.TaskKey) {
			return timelineError("task_key", "invalid_key")
		}
		if err := timelineText(r.Title, "title", 120, true); err != nil {
			return err
		}
		if err := timelineText(r.Summary, "summary", 4000, false); err != nil {
			return err
		}
		if len(r.Segments) < 1 || len(r.Segments) > 5 {
			return timelineError("segments", "invalid_count")
		}
		seen := map[TimelineStage]bool{}
		for _, s := range r.Segments {
			if !contains(timelineStages, string(s.PlannedStage)) || seen[s.PlannedStage] {
				return timelineError("segments", "invalid_stage")
			}
			seen[s.PlannedStage] = true
			if err := timelineNumber(s.Order, "order"); err != nil {
				return err
			}
			if s.AssignedProducerID != "" {
				if err := timelineID(s.AssignedProducerID, "assigned_producer_id"); err != nil {
					return err
				}
			}
		}
	case *TimelineTaskPatchRequest:
		base = r.TimelineNativeOperation
		if r.Summary != nil {
			if err := timelineText(*r.Summary, "summary", 4000, false); err != nil {
				return err
			}
		}
		if err := timelineNumber(r.ExpectedTaskVersion, "expected_task_version"); err != nil {
			return err
		}
		if r.Title == nil && r.Summary == nil && r.PlanRevisionID == "" && r.SegmentID == "" && r.Withdrawn == nil {
			return timelineError("body", "empty_patch")
		}
		if r.Title != nil {
			if err := timelineText(*r.Title, "title", 120, true); err != nil {
				return err
			}
		}
		if r.Order != nil {
			if err := timelineNumber(*r.Order, "order"); err != nil {
				return err
			}
		}
		if (r.Order != nil || r.AssignedProducerID != "" || r.DisplayStageEventID != "") && r.SegmentID == "" {
			return timelineError("segment_id", "required_for_segment_patch")
		}
		if r.DecisionTarget != nil {
			if err := timelineTarget(*r.DecisionTarget); err != nil {
				return err
			}
		}
	case *TimelineTaskTransitionRequest:
		base = r.TimelineNativeOperation
		if err := timelineNumber(r.ExpectedTaskVersion, "expected_task_version"); err != nil {
			return err
		}
		if err := timelineID(r.SegmentID, "segment_id"); err != nil {
			return err
		}
		if !contains(timelineTaskStates, r.State) {
			return timelineError("state", "invalid_state")
		}
		if r.State == "blocked" {
			if r.Block == nil {
				return timelineError("block", "required_for_blocked")
			}
			if err := r.Block.Validate(); err != nil {
				return err
			}
		} else if r.Block != nil {
			return timelineError("block", "forbidden_for_state")
		}
		if r.ReviewKind != "" && (r.State != "in_review" || !contains([]string{"human", "agent"}, r.ReviewKind)) {
			return timelineError("review_kind", "forbidden_for_state")
		}
	case *TimelineWaitCreateRequest:
		base = r.TimelineNativeOperation
		if err := timelineText(r.NextAction, "next_action", TimelineMaxWaitText, false); err != nil {
			return err
		}
		if err := timelineID(r.TargetStageEventID, "target_stage_event_id"); err != nil {
			return err
		}
		if err := timelineCandidate(r.Candidate); err != nil {
			return err
		}
		if err := timelineHTTPS(r.PullRequestURL, "pull_request_url"); err != nil {
			return err
		}
		if err := r.Policy.Validate(); err != nil {
			return err
		}
		if (r.TaskID == "") != (r.SegmentID == "") {
			return timelineError("task_id", "task_segment_pair_required")
		}
	case *TimelineWaitAttemptStartRequest:
		base = r.TimelineNativeOperation
		if err := timelineNumber(r.ExpectedWaitVersion, "expected_wait_version"); err != nil {
			return err
		}
		if len(r.CheckKeys) < 1 || len(r.CheckKeys) > TimelineMaxRequiredChecks {
			return timelineError("check_keys", "invalid_count")
		}
		seen := map[string]bool{}
		for _, k := range r.CheckKeys {
			if err := timelineID(k, "check_keys"); err != nil {
				return err
			}
			if seen[k] {
				return timelineError("check_keys", "duplicate_key")
			}
			seen[k] = true
		}
		if err := timelineRun(r.ExternalRun); err != nil {
			return err
		}
		if err := timelineHTTPS(r.SourceURL, "source_url"); err != nil {
			return err
		}
		if err := timelineTime(r.ObservedAt, "observed_at"); err != nil {
			return err
		}
	case *TimelineWaitObserveRequest:
		base = r.TimelineNativeOperation
		if err := timelineNumber(r.ExpectedWaitVersion, "expected_wait_version"); err != nil {
			return err
		}
		if err := timelineID(r.AttemptID, "attempt_id"); err != nil {
			return err
		}
		if err := timelineID(r.CheckKey, "check_key"); err != nil {
			return err
		}
		if !contains(timelineWaitStates, r.State) {
			return timelineError("state", "invalid_state")
		}
		if err := timelineRun(r.ExternalRun); err != nil {
			return err
		}
		if err := timelineTime(r.ObservedAt, "observed_at"); err != nil {
			return err
		}
		if err := timelineText(r.Summary, "summary", TimelineMaxWaitText, true); err != nil {
			return err
		}
	default:
		return timelineError("kind", "unsupported_operation")
	}
	return base.validate(false)
}
func timelineRun(r TimelineCIExternalRun) error {
	if err := timelineText(r.Commit, "commit", 200, true); err != nil {
		return err
	}
	if err := timelineText(r.RunID, "run_id", 200, true); err != nil {
		return err
	}
	return timelineHTTPS(r.URL, "url")
}
func (p TimelineCheckPolicy) Validate() error {
	if err := timelineID(p.PolicyID, "policy_id"); err != nil {
		return err
	}
	if err := timelineText(p.Authority, "authority", 500, true); err != nil {
		return err
	}
	if err := timelineHTTPS(p.SourceURL, "source_url"); err != nil {
		return err
	}
	if err := timelineTime(p.ObservedAt, "observed_at"); err != nil {
		return err
	}
	if len(p.RequiredChecks) > TimelineMaxRequiredChecks {
		return timelineError("required_checks", "limit")
	}
	seen := map[string]bool{}
	for _, c := range p.RequiredChecks {
		if err := timelineID(c.CheckKey, "check_key"); err != nil {
			return err
		}
		if seen[c.CheckKey] {
			return timelineError("check_key", "duplicate_key")
		}
		seen[c.CheckKey] = true
		if err := timelineText(c.Label, "label", 200, true); err != nil {
			return err
		}
	}
	return nil
}

// AggregateTimelineTaskState excludes withdrawn requirements; callers omit counts
// for an empty set rather than presenting a fictitious 0/0 completion.
func AggregateTimelineTaskState(segments []TimelineTaskSegment) string {
	n, completed, notStarted := 0, 0, 0
	review := false
	for _, s := range segments {
		if s.Withdrawn {
			continue
		}
		n++
		switch s.State {
		case "blocked":
			return "blocked"
		case "in_review":
			review = true
		case "completed":
			completed++
		case "not_started":
			notStarted++
		}
	}
	if review {
		return "in_review"
	}
	if n > 0 && completed == n {
		return "completed"
	}
	if notStarted == n {
		return "not_started"
	}
	return "in_progress"
}
func ValidateTimelineTaskTransition(previous string, r TimelineTaskTransitionRequest) error {
	if err := ValidateTimelineV2Operation(&r); err != nil {
		return err
	}
	if previous == r.State {
		return nil
	}
	allowed := map[string][]string{"not_started": {"in_progress", "blocked"}, "in_progress": {"in_review", "blocked", "completed"}, "in_review": {"in_progress", "blocked", "completed"}, "blocked": {"in_progress"}, "completed": {"in_progress"}}
	if !contains(allowed[previous], r.State) {
		return timelineError("state", "invalid_transition")
	}
	if previous == "blocked" {
		return timelineText(r.Resolution, "resolution", 500, true)
	}
	return nil
}

// ValidateTimelineAuthority is independent of payload shape and checks accepted
// membership on every mutation. Ownership moderation is explicit, never byte edit.
// ownerProducerID is derived from stored content ownership or the wait segment
// assignment, never from a caller-supplied claim. Project-resolution preconditions
// are defined by ValidateTimelineResolvedProject; path equality is not identity.
func ValidateTimelineAuthority(s TimelineSession, operation string, actor TimelineActor, ownerProducerID string) error {
	if err := actor.Validate(); err != nil {
		return err
	}
	local := contains([]string{"session_rename", "session_archive", "session_pin", "session_delete", "recover_lead"}, operation)
	if actor.Kind == "local_viewer" {
		if local {
			return nil
		}
		return timelineError("actor", "forbidden_actor")
	}
	if local {
		return timelineError("actor", "local_viewer_required")
	}
	accepted := false
	for _, p := range s.Producers {
		if p.ProducerID == actor.ProducerID {
			accepted = true
		}
	}
	if !accepted {
		return timelineError("actor", "producer_not_accepted")
	}
	lead := actor.ProducerID == s.ForegroundProducerID
	switch operation {
	case "producer_accept", "foreground_transfer", "session_patch", "stage_start", "stage_update", "closure", "task_create", "task_patch":
		if !lead {
			return timelineError("actor", "lead_required")
		}
	case "event_patch":
		if ownerProducerID != "" && ownerProducerID != actor.ProducerID {
			return timelineError("actor", "content_owner_required")
		}
	case "event_withdraw":
		if ownerProducerID != "" && ownerProducerID != actor.ProducerID && !lead {
			return timelineError("actor", "content_owner_required")
		}
	case "wait_create", "wait_attempt_start", "wait_observe":
		if !lead && (ownerProducerID == "" || actor.ProducerID != ownerProducerID) {
			return timelineError("actor", "segment_assignee_required")
		}
	case "append_event", "artifact_add", "document_add", "document_snapshot", "task_transition":
	default:
		return timelineError("operation", "unsupported_operation")
	}
	return nil
}
func ValidateTimelineTaskReference(s TimelineSession, taskID, segmentID, stageID string) error {
	if taskID == "" && segmentID == "" {
		return nil
	}
	for _, t := range s.Tasks {
		if t.TaskID != taskID || t.Withdrawn {
			continue
		}
		for _, seg := range t.Segments {
			if seg.SegmentID == segmentID && !seg.Withdrawn {
				if seg.DisplayStageEventID != stageID || stageID == "" {
					return timelineError("segment_id", "stage_binding_mismatch")
				}
				for _, e := range s.Events {
					if e.EventID == stageID && e.Kind == "stage" && !e.Withdrawn && e.Stage == seg.PlannedStage {
						return nil
					}
				}
			}
		}
	}
	return timelineError("task_id", "unknown_task_segment")
}

// ValidateTimelineV2Graph checks structure and server-owned membership. Before
// trusting restored authority, follow ValidateTimelineResolvedProject load rules.
func ValidateTimelineV2Graph(s TimelineSession) error {
	if s.SchemaVersion != 2 {
		return timelineError("schema_version", "unsupported_protocol")
	}
	if s.PendingJoinRequests == nil || s.Tasks == nil || s.Waits == nil || len(s.Producers) < 1 || len(s.Producers) > 16 || len(s.PendingJoinRequests) > 16 || len(s.Tasks) > 100 || len(s.Waits) > TimelineMaxWaits {
		return timelineError("session", "v2_collection_limit")
	}
	if (s.State == "in_review" || s.State == "awaiting_validation") && s.CurrentStageEventID == nil {
		return timelineError("current_stage_event_id", "foreground_required")
	}
	if s.CurrentStageEventID != nil {
		for _, e := range s.Events {
			if e.EventID == *s.CurrentStageEventID {
				want := TimelineState(e.StageState)
				if e.StageState == "in_progress" {
					want = TimelineActive
				}
				if s.State != want {
					return timelineError("state", "foreground_state_mismatch")
				}
			}
		}
	}
	members := map[string]bool{}
	sources := map[timelineNativeIdentity]bool{}
	originalAccepted := false
	lead := false
	for _, p := range s.Producers {
		if err := timelineID(p.ProducerID, "producer_id"); err != nil {
			return err
		}
		if err := timelineSource(p.Source); err != nil {
			return err
		}
		if members[p.ProducerID] || sources[timelineNativeSourceIdentity(p.Source)] {
			return timelineError("producers", "duplicate_identity")
		}
		if err := timelineTime(p.JoinedAt, "joined_at"); err != nil {
			return err
		}
		members[p.ProducerID] = true
		sources[timelineNativeSourceIdentity(p.Source)] = true
		originalAccepted = originalAccepted || p.Source == s.Source
		lead = lead || p.ProducerID == s.ForegroundProducerID
	}
	if !lead || !originalAccepted {
		return timelineError("foreground_producer_id", "accepted_original_and_lead_required")
	}
	requests := map[string]bool{}
	for _, r := range s.PendingJoinRequests {
		if err := timelineID(r.RequestID, "request_id"); err != nil {
			return err
		}
		if err := timelineSource(r.Source); err != nil {
			return err
		}
		if sources[timelineNativeSourceIdentity(r.Source)] || requests[r.RequestID] {
			return timelineError("pending_join_requests", "duplicate_identity")
		}
		if err := timelineReason(r.Reason); err != nil {
			return err
		}
		if err := timelineTime(r.RequestedAt, "requested_at"); err != nil {
			return err
		}
		requests[r.RequestID] = true
		sources[timelineNativeSourceIdentity(r.Source)] = true
	}
	acceptedActor := func(a TimelineActor) error {
		if err := validateTimelineNativeActor(a); err != nil {
			return err
		}
		if !members[a.ProducerID] {
			return timelineError("actor", "producer_not_accepted")
		}
		return nil
	}
	for _, d := range s.Documents {
		if err := acceptedActor(d.Actor); err != nil {
			return err
		}
	}
	for _, a := range s.Artifacts {
		if err := acceptedActor(a.Actor); err != nil {
			return err
		}
	}
	stages := map[string]TimelineEvent{}
	revisions := map[string]TimelineEvent{}
	chains := map[string]string{}
	for _, e := range s.Events {
		if err := e.Actor.Validate(); err != nil {
			return err
		}
		if e.Actor.Kind == "native_producer" && !members[e.Actor.ProducerID] {
			return timelineError("actor", "producer_not_accepted")
		}
		if e.Kind == "stage" {
			stages[e.EventID] = e
		}
		for _, r := range e.DocumentSnapshots {
			if err := acceptedActor(r.Actor); err != nil {
				return err
			}
			if !timelineDocumentKey.MatchString(r.DocumentKey) || r.PredecessorRevisionID != chains[r.DocumentKey] {
				return timelineError("document_key", "invalid_chain")
			}
			if prev, ok := revisions[r.PredecessorRevisionID]; ok && prev.EventID != e.EventID && (prev.Stage != e.Stage || e.RelatedEventID != prev.EventID) {
				return timelineError("predecessor_revision_id", "stage_chain_mismatch")
			}
			chains[r.DocumentKey] = r.DocumentRevisionID
			revisions[r.DocumentRevisionID] = e
		}
	}
	ids, keys, segments := map[string]bool{}, map[string]bool{}, map[string]bool{}
	orders := map[TimelineStage]map[int64]bool{}
	for _, t := range s.Tasks {
		if err := timelineID(t.TaskID, "task_id"); err != nil {
			return err
		}
		if ids[t.TaskID] || keys[t.TaskKey] || !timelineTaskKey.MatchString(t.TaskKey) || len(t.Segments) < 1 || len(t.Segments) > 5 {
			return timelineError("tasks", "invalid_identity_or_segments")
		}
		ids[t.TaskID] = true
		keys[t.TaskKey] = true
		if err := timelineText(t.Title, "title", 120, true); err != nil {
			return err
		}
		if err := timelineNumber(t.TaskVersion, "task_version"); err != nil {
			return err
		}
		if t.PlanRevisionID != "" {
			if _, ok := revisions[t.PlanRevisionID]; !ok {
				return timelineError("plan_revision_id", "unknown_revision")
			}
		}
		kinds := map[TimelineStage]bool{}
		for _, seg := range t.Segments {
			if err := timelineID(seg.SegmentID, "segment_id"); err != nil {
				return err
			}
			if segments[seg.SegmentID] || kinds[seg.PlannedStage] || !contains(timelineStages, string(seg.PlannedStage)) || !contains(timelineTaskStates, seg.State) || !members[seg.AssignedProducerID] {
				return timelineError("segments", "invalid_segment")
			}
			segments[seg.SegmentID] = true
			kinds[seg.PlannedStage] = true
			if err := timelineNumber(seg.Order, "order"); err != nil {
				return err
			}
			if orders[seg.PlannedStage] == nil {
				orders[seg.PlannedStage] = map[int64]bool{}
			}
			if !seg.Withdrawn && !t.Withdrawn {
				if orders[seg.PlannedStage][seg.Order] {
					return timelineError("order", "duplicate_stage_order")
				}
				orders[seg.PlannedStage][seg.Order] = true
			}
			if seg.DisplayStageEventID != "" {
				e, ok := stages[seg.DisplayStageEventID]
				if !ok || e.Stage != seg.PlannedStage {
					return timelineError("display_stage_event_id", "stage_binding_mismatch")
				}
			}
			if seg.State == "blocked" {
				if seg.Block == nil {
					return timelineError("block", "required_for_blocked")
				}
				if err := seg.Block.Validate(); err != nil {
					return err
				}
			} else if seg.Block != nil {
				return timelineError("block", "forbidden_for_state")
			}
		}
		if t.State != AggregateTimelineTaskState(t.Segments) {
			return timelineError("state", "aggregate_mismatch")
		}
	}
	for _, e := range s.Events {
		if e.Audit != nil || contains([]string{"state_transition", "producer_joined", "foreground_transferred", "task_state_transition", "local_action"}, string(e.Kind)) {
			if err := validateTimelineAuditReferences(s, e); err != nil {
				return err
			}
		}
		if e.StageEventID != "" {
			stage, ok := stages[e.StageEventID]
			if !ok || e.Kind == "stage" || e.Kind == "closure" || stage.Sequence > e.Sequence {
				return timelineError("stage_event_id", "invalid_stage_reference")
			}
		}
		if err := ValidateTimelineTaskReference(s, e.TaskID, e.SegmentID, e.StageEventID); err != nil {
			return err
		}
	}
	for _, e := range s.Events {
		for _, r := range e.DocumentSnapshots {
			if err := ValidateTimelineTaskReference(s, r.TaskID, r.SegmentID, e.EventID); err != nil {
				return err
			}
		}
	}
	waits := map[string]bool{}
	for _, w := range s.Waits {
		if err := timelineID(w.WaitID, "wait_id"); err != nil {
			return err
		}
		if waits[w.WaitID] {
			return timelineError("wait_id", "duplicate_id")
		}
		if w.PreviousWaitID != "" && !waits[w.PreviousWaitID] {
			return timelineError("previous_wait_id", "unknown_wait")
		}
		waits[w.WaitID] = true
		if err := timelineNumber(w.WaitVersion, "wait_version"); err != nil {
			return err
		}
		if e, ok := stages[w.StageEventID]; !ok || e.Stage != "delivery" {
			return timelineError("stage_event_id", "delivery_required")
		}
		if err := ValidateTimelineTaskReference(s, w.TaskID, w.SegmentID, w.StageEventID); err != nil {
			return err
		}
		for _, a := range w.Attempts {
			if err := acceptedActor(a.Actor); err != nil {
				return err
			}
		}
		for _, o := range w.Observations {
			if err := acceptedActor(o.Actor); err != nil {
				return err
			}
		}
		if err := ValidateTimelineCIWait(w); err != nil {
			return err
		}
	}
	return nil
}
func ValidateTimelineCIWait(w TimelineCIWait) error {
	if err := timelineText(w.NextAction, "next_action", TimelineMaxWaitText, false); err != nil {
		return err
	}
	if err := w.Policy.Validate(); err != nil {
		return err
	}
	if err := timelineCandidate(w.Candidate); err != nil {
		return err
	}
	if err := timelineHTTPS(w.PullRequestURL, "pull_request_url"); err != nil {
		return err
	}
	if !contains(timelineWaitStates, w.State) || len(w.Attempts) > TimelineMaxWaitAttempts || len(w.Observations) > TimelineMaxWaitObservations {
		return timelineError("wait", "invalid_state_or_limit")
	}
	required := map[string]bool{}
	for _, c := range w.Policy.RequiredChecks {
		required[c.CheckKey] = true
	}
	attempts := map[string]TimelineCIAttempt{}
	var gen int64
	for _, a := range w.Attempts {
		if a.ExternalRun.Commit != w.Candidate.Commit {
			return timelineError("commit", "candidate_mismatch")
		}
		if err := timelineID(a.AttemptID, "attempt_id"); err != nil {
			return err
		}
		if _, ok := attempts[a.AttemptID]; ok || a.Generation <= gen {
			return timelineError("generation", "non_monotonic_attempt")
		}
		gen = a.Generation
		if len(a.CheckKeys) < 1 {
			return timelineError("check_keys", "empty")
		}
		seen := map[string]bool{}
		for _, k := range a.CheckKeys {
			if !required[k] || seen[k] {
				return timelineError("check_key", "unknown_or_duplicate")
			}
			seen[k] = true
		}
		if err := timelineRun(a.ExternalRun); err != nil {
			return err
		}
		if err := timelineHTTPS(a.SourceURL, "source_url"); err != nil {
			return err
		}
		if err := timelineTime(a.ObservedAt, "observed_at"); err != nil {
			return err
		}
		if err := validateTimelineNativeActor(a.Actor); err != nil {
			return err
		}
		attempts[a.AttemptID] = a
	}
	for k, id := range w.ApplicableAttemptID {
		a, ok := attempts[id]
		if !required[k] || !ok || !contains(a.CheckKeys, k) {
			return timelineError("applicable_attempt_id", "invalid_attempt")
		}
		for _, other := range w.Attempts {
			if contains(other.CheckKeys, k) && other.Generation > a.Generation {
				return timelineError("applicable_attempt_id", "superseded_attempt")
			}
		}
	}
	observed := map[string]TimelineCIObservation{}
	ids := map[string]bool{}
	for _, o := range w.Observations {
		if err := timelineText(o.Summary, "summary", TimelineMaxWaitText, true); err != nil {
			return err
		}
		a, ok := attempts[o.AttemptID]
		if !ok || !contains(a.CheckKeys, o.CheckKey) || o.ExternalRun != a.ExternalRun || ids[o.ObservationID] {
			return timelineError("observations", "unknown_or_mixed_attempt")
		}
		if err := timelineID(o.ObservationID, "observation_id"); err != nil {
			return err
		}
		ids[o.ObservationID] = true
		if !contains(timelineWaitStates, o.State) {
			return timelineError("state", "invalid_state")
		}
		if err := timelineTime(o.ObservedAt, "observed_at"); err != nil {
			return err
		}
		if err := timelineTime(o.ReceivedAt, "received_at"); err != nil {
			return err
		}
		if err := validateTimelineNativeActor(o.Actor); err != nil {
			return err
		}
		if w.ApplicableAttemptID[o.CheckKey] == o.AttemptID {
			prev, exists := observed[o.CheckKey]
			prevTime, _ := time.Parse(time.RFC3339Nano, prev.ObservedAt)
			nextTime, _ := time.Parse(time.RFC3339Nano, o.ObservedAt)
			if !exists || prevTime.Before(nextTime) {
				observed[o.CheckKey] = o
			}
		}
	}
	if w.State != AggregateTimelineCIWait(w.Policy, w.ApplicableAttemptID, observed) {
		return timelineError("state", "wait_aggregate_mismatch")
	}
	return nil
}
func AggregateTimelineCIWait(policy TimelineCheckPolicy, applicable map[string]string, observed map[string]TimelineCIObservation) string {
	if !policy.Confirmed || len(policy.RequiredChecks) == 0 {
		return "unknown"
	}
	waiting, cancelled, unknown := false, false, false
	for _, c := range policy.RequiredChecks {
		o, ok := observed[c.CheckKey]
		if !ok || applicable[c.CheckKey] == "" || o.AttemptID != applicable[c.CheckKey] {
			unknown = true
			continue
		}
		switch o.State {
		case "failed":
			return "failed"
		case "cancelled":
			cancelled = true
		case "unknown":
			unknown = true
		case "waiting":
			waiting = true
		}
	}
	if cancelled {
		return "cancelled"
	}
	if unknown {
		return "unknown"
	}
	if waiting {
		return "waiting"
	}
	return "succeeded"
}

var timelineV2ResultFields = map[string][]string{
	"join_request": {"request_id"}, "producer_accept": {"request_id", "producer_id"}, "foreground_transfer": {"producer_id"}, "recover_lead": {"request_id", "producer_id"},
	"task_create": {"task_id", "task_version"}, "task_patch": {"task_id", "task_version"}, "task_transition": {"task_id", "task_version"},
	"wait_create": {"wait_id", "wait_version"}, "wait_attempt_start": {"wait_id", "wait_version", "attempt_id", "generation"}, "wait_observe": {"wait_id", "wait_version", "observation_id"},
}

func validateTimelineV2Result(r TimelineOperationResult) error {
	values := map[string]string{"request_id": r.RequestID, "producer_id": r.ProducerID, "task_id": r.TaskID, "wait_id": r.WaitID, "attempt_id": r.AttemptID, "observation_id": r.ObservationID}
	numbers := map[string]int64{"task_version": r.TaskVersion, "wait_version": r.WaitVersion, "generation": r.Generation}
	for _, k := range timelineV2ResultFields[string(r.Kind)] {
		if n, ok := numbers[k]; ok {
			if err := timelineNumber(n, k); err != nil {
				return err
			}
		} else if err := timelineID(values[k], k); err != nil {
			return err
		}
	}
	return nil
}

// Audit is a closed union selected by event kind and operation; a server origin
// alone never legitimizes incomplete or mixed transition branches.
func validateTimelineAuditEvent(e TimelineEvent) error {
	a := e.Audit
	if a == nil || e.Origin != "server" || e.Withdrawn {
		return timelineError("audit", "server_event_required")
	}
	if err := e.Actor.Validate(); err != nil {
		return err
	}
	if err := timelineReason(a.Reason); err != nil {
		return err
	}
	allowed := map[string]bool{"operation": true, "reason": true, "target": true}
	require := func(fields ...string) {
		for _, f := range fields {
			allowed[f] = true
		}
	}
	switch e.Kind {
	case "state_transition":
		if !contains([]string{"stage_start", "stage_update", "session_patch", "closure"}, a.Operation) {
			return timelineError("audit.operation", "invalid_audit_operation")
		}
		if err := validateTimelineNativeActor(e.Actor); err != nil {
			return err
		}
		states := timelineStates
		if a.Operation == "stage_start" || a.Operation == "stage_update" {
			states = timelineStageStates
			if err := timelineID(e.StageEventID, "stage_event_id"); err != nil {
				return err
			}
		}
		if !contains(states, a.PreviousState) || !contains(states, a.NextState) || a.PreviousState == a.NextState {
			return timelineError("audit", "invalid_state_transition")
		}
		require("previous_state", "next_state")
	case "producer_joined":
		if !contains([]string{"producer_accept", "recover_lead"}, a.Operation) {
			return timelineError("audit.operation", "invalid_audit_operation")
		}
		if err := timelineID(a.RequestID, "audit.request_id"); err != nil {
			return err
		}
		if err := timelineID(a.NextProducerID, "audit.next_producer_id"); err != nil {
			return err
		}
		require("request_id", "next_producer_id")
	case "foreground_transferred":
		if !contains([]string{"foreground_transfer", "recover_lead"}, a.Operation) {
			return timelineError("audit.operation", "invalid_audit_operation")
		}
		if err := timelineID(a.PreviousProducerID, "audit.previous_producer_id"); err != nil {
			return err
		}
		if err := timelineID(a.NextProducerID, "audit.next_producer_id"); err != nil {
			return err
		}
		if a.PreviousProducerID == a.NextProducerID {
			return timelineError("audit", "invalid_producer_transition")
		}
		require("previous_producer_id", "next_producer_id")
		if a.Operation == "recover_lead" {
			if err := timelineID(a.RequestID, "audit.request_id"); err != nil {
				return err
			}
			require("request_id")
		}
	case "task_state_transition":
		if a.Operation != "task_transition" {
			return timelineError("audit.operation", "invalid_audit_operation")
		}
		if err := validateTimelineNativeActor(e.Actor); err != nil {
			return err
		}
		for field, id := range map[string]string{"task_id": e.TaskID, "segment_id": e.SegmentID, "stage_event_id": e.StageEventID} {
			if err := timelineID(id, field); err != nil {
				return err
			}
		}
		transitions := map[string][]string{"not_started": {"in_progress", "blocked"}, "in_progress": {"in_review", "blocked", "completed"}, "in_review": {"in_progress", "blocked", "completed"}, "blocked": {"in_progress"}, "completed": {"in_progress"}}
		if !contains(transitions[a.PreviousState], a.NextState) {
			return timelineError("audit", "invalid_state_transition")
		}
		require("previous_state", "next_state", "block", "resolution", "previous_segment", "next_segment", "plan_revision_id")
		if a.NextState == "blocked" && a.Block == nil {
			return timelineError("block", "required_for_blocked")
		}
		if a.Block != nil {
			if a.NextState != "blocked" && a.PreviousState != "blocked" {
				return timelineError("block", "forbidden_for_state")
			}
			if err := a.Block.Validate(); err != nil {
				return err
			}
		}
		if a.PreviousState == "blocked" {
			if err := timelineText(a.Resolution, "resolution", 500, true); err != nil {
				return err
			}
		} else if a.Resolution != "" {
			return timelineError("resolution", "forbidden_for_state")
		}
		for _, seg := range []*TimelineTaskSegment{a.PreviousSegment, a.NextSegment} {
			if seg != nil && seg.SegmentID != e.SegmentID {
				return timelineError("audit.segment_id", "unknown_task_segment")
			}
		}
	case "local_action":
		if !contains([]string{"session_rename", "session_archive", "session_pin", "session_delete"}, a.Operation) || e.Actor.Kind != "local_viewer" {
			return timelineError("audit.operation", "invalid_audit_operation")
		}
	default:
		return timelineError("audit", "server_event_required")
	}
	if a.Operation == "recover_lead" {
		if e.Actor.Kind != "local_viewer" {
			return timelineError("actor", "local_viewer_required")
		}
	} else if e.Kind == "producer_joined" || e.Kind == "foreground_transferred" {
		if err := validateTimelineNativeActor(e.Actor); err != nil {
			return err
		}
	}
	fields := map[string]bool{"block": a.Block != nil, "resolution": a.Resolution != "", "previous_segment": a.PreviousSegment != nil, "next_segment": a.NextSegment != nil, "plan_revision_id": a.PlanRevisionID != "", "previous_state": a.PreviousState != "", "next_state": a.NextState != "", "previous_producer_id": a.PreviousProducerID != "", "next_producer_id": a.NextProducerID != "", "request_id": a.RequestID != ""}
	for f, present := range fields {
		if present && !allowed[f] {
			return timelineError("audit."+f, "forbidden_for_operation")
		}
	}
	if a.Target != nil {
		return timelineTarget(*a.Target)
	}
	return nil
}
func validateTimelineAuditReferences(s TimelineSession, e TimelineEvent) error {
	if err := validateTimelineAuditEvent(e); err != nil {
		return err
	}
	a := e.Audit
	for _, id := range []string{a.PreviousProducerID, a.NextProducerID} {
		if id != "" {
			known := false
			for _, p := range s.Producers {
				known = known || p.ProducerID == id
			}
			if !known {
				return timelineError("audit.producer_id", "producer_not_accepted")
			}
		}
	}
	if a.PlanRevisionID != "" && !timelineHasRevision(s, a.PlanRevisionID) {
		return timelineError("plan_revision_id", "unknown_revision")
	}
	if a.Target != nil {
		prefix := s
		prefix.Events = nil
		for _, event := range s.Events {
			if event.Sequence < e.Sequence {
				prefix.Events = append(prefix.Events, event)
			}
		}
		if err := timelineEventReferences(prefix, TimelineAppendEventRequest{Target: a.Target}); err != nil {
			return err
		}
	}
	return nil
}
func timelineV2ProtocolDefinitions(actor TimelineActor) []timelineProtocolDefinition {
	base := TimelineNativeOperation{SchemaVersion: 2, OperationID: "operation-example", Actor: actor, ExpectedSessionVersion: 1, Reason: "Explicit authorized change"}
	recovery := base
	recovery.Actor = TimelineActor{Kind: "local_viewer"}
	policy := TimelineCheckPolicy{PolicyID: "policy-example", Confirmed: true, Authority: "Repository branch policy", SourceURL: "https://example.test/policy", ObservedAt: "2026-11-01T00:00:00Z", RequiredChecks: []TimelineRequiredCheck{{CheckKey: "test", Label: "Tests"}}}
	run := TimelineCIExternalRun{Commit: "abc1234", RunID: "run-example", URL: "https://example.test/runs/1"}
	return []timelineProtocolDefinition{
		{"join_request", "hive view producer request <session-id> --json <file|->", TimelineJoinRequestInput{}, TimelineOperationResult{}, TimelineJoinRequestInput{SchemaVersion: 2, OperationID: "join-example", Source: TimelineSource{ProjectRoot: "/workspace/example", Host: "claude-code", NativeSessionID: "real-native-session"}, ExpectedSessionVersion: 1, Reason: "Authorized collaboration"}},
		{"producer_accept", "hive view producer accept <session-id> --json <file|->", TimelineProducerAcceptRequest{}, TimelineOperationResult{}, TimelineProducerAcceptRequest{TimelineNativeOperation: base, RequestID: "request-example"}},
		{"foreground_transfer", "hive view producer transfer <session-id> --json <file|->", TimelineForegroundTransferRequest{}, TimelineOperationResult{}, TimelineForegroundTransferRequest{TimelineNativeOperation: base, TargetProducerID: "producer-next"}},
		{"recover_lead", "Local viewer only: recover_lead", TimelineRecoverLeadRequest{}, TimelineOperationResult{}, TimelineRecoverLeadRequest{TimelineNativeOperation: recovery, RequestID: "request-example"}},
		{"task_create", "hive view task create <session-id> --json <file|->", TimelineTaskCreateRequest{}, TimelineOperationResult{}, TimelineTaskCreateRequest{TimelineNativeOperation: base, TaskKey: "T1", Title: "Implement contract", Summary: "Freeze typed boundaries", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}},
		{"task_patch", "hive view task update <session-id> <task-id> --json <file|->", TimelineTaskPatchRequest{}, TimelineOperationResult{}, TimelineTaskPatchRequest{TimelineNativeOperation: base, ExpectedTaskVersion: 1, Title: timelineStringPointer("Updated task")}},
		{"task_transition", "hive view task transition <session-id> <task-id> --json <file|->", TimelineTaskTransitionRequest{}, TimelineOperationResult{}, TimelineTaskTransitionRequest{TimelineNativeOperation: base, ExpectedTaskVersion: 1, SegmentID: "segment-example", State: "in_progress"}},
		{"wait_create", "hive view wait create <session-id> --json <file|->", TimelineWaitCreateRequest{}, TimelineOperationResult{}, TimelineWaitCreateRequest{TimelineNativeOperation: base, TargetStageEventID: "stage-delivery", Candidate: TimelineCandidate{Branch: "feature/example", Commit: "abc1234", Base: "development"}, PullRequestURL: "https://example.test/pull/1", Policy: policy}},
		{"wait_attempt_start", "hive view wait attempt <session-id> <wait-id> --json <file|->", TimelineWaitAttemptStartRequest{}, TimelineOperationResult{}, TimelineWaitAttemptStartRequest{TimelineNativeOperation: base, ExpectedWaitVersion: 1, CheckKeys: []string{"test"}, ExternalRun: run, SourceURL: "https://example.test/runs/1", ObservedAt: "2026-11-01T00:00:00Z"}},
		{"wait_observe", "hive view wait observe <session-id> <wait-id> --json <file|->", TimelineWaitObserveRequest{}, TimelineOperationResult{}, TimelineWaitObserveRequest{TimelineNativeOperation: base, ExpectedWaitVersion: 1, AttemptID: "attempt-example", CheckKey: "test", ExternalRun: run, State: "waiting", Summary: "Tests running", ObservedAt: "2026-11-01T00:00:00Z"}},
	}
}

// ValidateTimelineV2OperationReferences closes authority and same-work references.
// objectID is the task/wait route parameter, never trusted as an ownership claim.
// Registered-project resolution is a separate required server boundary documented
// by ValidateTimelineResolvedProject, not a comparison of Source root strings.
func ValidateTimelineV2OperationReferences(s TimelineSession, kind, objectID string, value any) error {
	if err := ValidateTimelineV2Operation(value); err != nil {
		return err
	}
	if r, ok := value.(*TimelineJoinRequestInput); ok {
		identity := timelineNativeSourceIdentity(r.Source)
		for _, p := range s.Producers {
			if timelineNativeSourceIdentity(p.Source) == identity {
				return timelineError("source", "duplicate_identity")
			}
		}
		for _, pending := range s.PendingJoinRequests {
			if timelineNativeSourceIdentity(pending.Source) == identity {
				return timelineError("source", "duplicate_identity")
			}
		}
		if r.ExpectedSessionVersion != s.Version {
			return timelineError("expected_session_version", "version_conflict")
		}
		if len(s.PendingJoinRequests) >= 16 {
			return timelineError("pending_join_requests", "limit")
		}
		return nil
	}
	var base TimelineNativeOperation
	switch r := value.(type) {
	case *TimelineProducerAcceptRequest:
		base = r.TimelineNativeOperation
	case *TimelineForegroundTransferRequest:
		base = r.TimelineNativeOperation
	case *TimelineRecoverLeadRequest:
		base = r.TimelineNativeOperation
	case *TimelineTaskCreateRequest:
		base = r.TimelineNativeOperation
	case *TimelineTaskPatchRequest:
		base = r.TimelineNativeOperation
	case *TimelineTaskTransitionRequest:
		base = r.TimelineNativeOperation
	case *TimelineWaitCreateRequest:
		base = r.TimelineNativeOperation
	case *TimelineWaitAttemptStartRequest:
		base = r.TimelineNativeOperation
	case *TimelineWaitObserveRequest:
		base = r.TimelineNativeOperation
	default:
		return timelineError("operation", "unsupported_operation")
	}
	if base.ExpectedSessionVersion != s.Version {
		return timelineError("expected_session_version", "version_conflict")
	}
	if s.State == "closed" || s.State == "failed" {
		return timelineError("state", "session_terminal")
	}
	assignee := ""
	switch r := value.(type) {
	case *TimelineWaitCreateRequest:
		assignee = timelineWaitAssignee(s, r.TaskID, r.SegmentID)
	case *TimelineWaitAttemptStartRequest:
		if w, ok := timelineFindWait(s, objectID); ok {
			assignee = timelineWaitAssignee(s, w.TaskID, w.SegmentID)
		}
	case *TimelineWaitObserveRequest:
		if w, ok := timelineFindWait(s, objectID); ok {
			assignee = timelineWaitAssignee(s, w.TaskID, w.SegmentID)
		}
	}
	if err := ValidateTimelineAuthority(s, kind, base.Actor, assignee); err != nil {
		return err
	}
	member := func(id string) bool {
		for _, p := range s.Producers {
			if p.ProducerID == id {
				return true
			}
		}
		return false
	}
	pending := func(id string) error {
		for _, r := range s.PendingJoinRequests {
			if r.RequestID == id {
				if len(s.Producers) >= 16 {
					return timelineError("producers", "limit")
				}
				return nil
			}
		}
		return timelineError("request_id", "pending_request_required")
	}
	switch r := value.(type) {
	case *TimelineProducerAcceptRequest:
		return pending(r.RequestID)
	case *TimelineRecoverLeadRequest:
		return pending(r.RequestID)
	case *TimelineForegroundTransferRequest:
		if !member(r.TargetProducerID) {
			return timelineError("target_producer_id", "producer_not_accepted")
		}
	case *TimelineTaskCreateRequest:
		if len(s.Tasks) >= 100 {
			return timelineError("tasks", "limit")
		}
		for _, t := range s.Tasks {
			if t.TaskKey == r.TaskKey {
				return timelineError("task_key", "duplicate_key")
			}
		}
		if r.PlanRevisionID != "" && !timelineHasRevision(s, r.PlanRevisionID) {
			return timelineError("plan_revision_id", "unknown_revision")
		}
		for _, seg := range r.Segments {
			if seg.AssignedProducerID != "" && !member(seg.AssignedProducerID) {
				return timelineError("assigned_producer_id", "producer_not_accepted")
			}
			for _, t := range s.Tasks {
				for _, old := range t.Segments {
					if !t.Withdrawn && !old.Withdrawn && old.PlannedStage == seg.PlannedStage && old.Order == seg.Order {
						return timelineError("order", "duplicate_stage_order")
					}
				}
			}
		}
	case *TimelineTaskPatchRequest:
		if r.DecisionTarget != nil {
			if err := timelineEventReferences(s, TimelineAppendEventRequest{Target: r.DecisionTarget}); err != nil {
				return err
			}
		}
		t, ok := timelineFindTask(s, objectID)
		if !ok {
			return timelineError("task_id", "unknown_task")
		}
		if t.TaskVersion != r.ExpectedTaskVersion {
			return timelineError("expected_task_version", "version_conflict")
		}
		if r.PlanRevisionID != "" && !timelineHasRevision(s, r.PlanRevisionID) {
			return timelineError("plan_revision_id", "unknown_revision")
		}
		if r.AssignedProducerID != "" && !member(r.AssignedProducerID) {
			return timelineError("assigned_producer_id", "producer_not_accepted")
		}
		if r.SegmentID != "" {
			found := false
			for _, seg := range t.Segments {
				if seg.SegmentID != r.SegmentID {
					continue
				}
				found = true
				if r.DisplayStageEventID != "" {
					match := false
					for _, e := range s.Events {
						if e.EventID == r.DisplayStageEventID && e.Kind == "stage" && !e.Withdrawn && e.Stage == seg.PlannedStage {
							match = true
						}
					}
					if !match {
						return timelineError("display_stage_event_id", "stage_binding_mismatch")
					}
				}
			}
			if !found {
				return timelineError("segment_id", "unknown_task_segment")
			}
		}
	case *TimelineTaskTransitionRequest:
		t, ok := timelineFindTask(s, objectID)
		if !ok {
			return timelineError("task_id", "unknown_task")
		}
		if t.TaskVersion != r.ExpectedTaskVersion {
			return timelineError("expected_task_version", "version_conflict")
		}
		for _, seg := range t.Segments {
			if seg.SegmentID != r.SegmentID || seg.Withdrawn {
				continue
			}
			if base.Actor.ProducerID != s.ForegroundProducerID && base.Actor.ProducerID != seg.AssignedProducerID {
				return timelineError("actor", "segment_assignee_required")
			}
			if err := ValidateTimelineTaskReference(s, t.TaskID, seg.SegmentID, seg.DisplayStageEventID); err != nil {
				return err
			}
			if seg.State == "completed" && r.State != "completed" {
				for _, e := range s.Events {
					if e.EventID == seg.DisplayStageEventID && (e.StageState == "completed" || e.StageState == "failed") {
						return timelineError("display_stage_event_id", "reentry_binding_required")
					}
				}
			}
			return ValidateTimelineTaskTransition(seg.State, *r)
		}
		return timelineError("segment_id", "unknown_task_segment")
	case *TimelineWaitCreateRequest:
		if len(s.Waits) >= TimelineMaxWaits {
			return timelineError("waits", "limit")
		}
		stage := false
		for _, e := range s.Events {
			if e.EventID == r.TargetStageEventID && e.Kind == "stage" && !e.Withdrawn && e.Stage == "delivery" {
				stage = true
			}
		}
		if !stage {
			return timelineError("target_stage_event_id", "delivery_required")
		}
		if err := ValidateTimelineTaskReference(s, r.TaskID, r.SegmentID, r.TargetStageEventID); err != nil {
			return err
		}
		if r.PreviousWaitID != "" {
			found := false
			for _, w := range s.Waits {
				if w.WaitID == r.PreviousWaitID {
					found = true
					if w.Candidate == r.Candidate && w.Policy.PolicyID == r.Policy.PolicyID {
						return timelineError("previous_wait_id", "same_candidate_policy")
					}
				}
			}
			if !found {
				return timelineError("previous_wait_id", "unknown_wait")
			}
		}
	case *TimelineWaitAttemptStartRequest:
		w, ok := timelineFindWait(s, objectID)
		if !ok {
			return timelineError("wait_id", "unknown_wait")
		}
		if w.WaitVersion != r.ExpectedWaitVersion {
			return timelineError("expected_wait_version", "version_conflict")
		}
		if r.ExternalRun.Commit != w.Candidate.Commit {
			return timelineError("commit", "candidate_mismatch")
		}
		if len(w.Attempts) >= TimelineMaxWaitAttempts {
			return timelineError("attempts", "limit")
		}
		for _, k := range r.CheckKeys {
			found := false
			for _, c := range w.Policy.RequiredChecks {
				found = found || c.CheckKey == k
			}
			if !found {
				return timelineError("check_key", "unknown_check")
			}
		}
		for _, a := range w.Attempts {
			if a.ExternalRun == r.ExternalRun {
				return timelineError("external_run", "attempt_already_known")
			}
		}
	case *TimelineWaitObserveRequest:
		w, ok := timelineFindWait(s, objectID)
		if !ok {
			return timelineError("wait_id", "unknown_wait")
		}
		if w.WaitVersion != r.ExpectedWaitVersion {
			return timelineError("expected_wait_version", "version_conflict")
		}
		if len(w.Observations) >= 500 {
			return timelineError("observations", "limit")
		}
		for _, a := range w.Attempts {
			if a.AttemptID == r.AttemptID && contains(a.CheckKeys, r.CheckKey) && a.ExternalRun == r.ExternalRun {
				return nil
			}
		}
		return timelineError("attempt_id", "unknown_or_mixed_attempt")
	}
	return nil
}
func timelineHasRevision(s TimelineSession, id string) bool {
	for _, e := range s.Events {
		for _, r := range e.DocumentSnapshots {
			if r.DocumentRevisionID == id {
				return true
			}
		}
	}
	return false
}
func timelineFindTask(s TimelineSession, id string) (TimelineTask, bool) {
	for _, t := range s.Tasks {
		if t.TaskID == id && !t.Withdrawn {
			return t, true
		}
	}
	return TimelineTask{}, false
}
func timelineFindWait(s TimelineSession, id string) (TimelineCIWait, bool) {
	for _, w := range s.Waits {
		if w.WaitID == id {
			return w, true
		}
	}
	return TimelineCIWait{}, false
}

// Content mutations use the first 500 event slots; eight remaining slots are
// reserved for lifecycle/closure audit. No operation may prune history to fit.
func ValidateTimelineEventCapacity(eventCount, newEvents int, lifecycle bool) error {
	limit := TimelineContentEvents
	if lifecycle {
		limit = TimelineMaxEvents
	}
	if eventCount < 0 || newEvents < 0 || eventCount+newEvents > limit {
		return timelineError("events", "limit")
	}
	return nil
}

// ResolveTimelineStageEventID is a server-only association decision. Explicit
// targets cannot name withdrawn/foreign occurrences; collaborators never fallback.
func ResolveTimelineStageEventID(s TimelineSession, r TimelineAppendEventRequest) (string, error) {
	operation := "append_event"
	if r.Kind == "stage" {
		operation = "stage_start"
	}
	if r.Kind == "closure" {
		operation = "closure"
	}
	if err := ValidateTimelineAuthority(s, operation, r.Actor, ""); err != nil {
		return "", err
	}
	if r.Kind == "stage" || r.Kind == "closure" || r.Kind == "request" || r.Kind == "work_started" {
		if r.TargetStageEventID != "" {
			return "", timelineError("target_stage_event_id", "forbidden_for_kind")
		}
		return "", nil
	}
	target := r.TargetStageEventID
	if target == "" {
		if r.Actor.ProducerID != s.ForegroundProducerID {
			return "", timelineError("target_stage_event_id", "collaborator_target_required")
		}
		if s.CurrentStageEventID != nil {
			target = *s.CurrentStageEventID
		} else {
			for _, e := range s.Events {
				if e.Kind == "stage" && !e.Withdrawn {
					target = e.EventID
				}
			}
		}
	}
	if target != "" {
		found := false
		for _, e := range s.Events {
			if e.EventID == target && e.Kind == "stage" && !e.Withdrawn {
				found = true
			}
		}
		if !found {
			return "", timelineError("target_stage_event_id", "stage_unavailable")
		}
	}
	if err := ValidateTimelineTaskReference(s, r.TaskID, r.SegmentID, target); err != nil {
		return "", err
	}
	return target, nil
}
func validateTimelineSnapshotReferencesV2(s TimelineSession, eventID string, r TimelineDocumentSnapshotRequest) error {
	if err := ValidateTimelineAuthority(s, "document_snapshot", r.Actor, ""); err != nil {
		return err
	}
	if r.ExpectedSessionVersion != s.Version {
		return timelineError("expected_session_version", "version_conflict")
	}
	if err := ValidateTimelineTaskReference(s, r.TaskID, r.SegmentID, eventID); err != nil {
		return err
	}
	var current TimelineEvent
	for _, e := range s.Events {
		if e.EventID == eventID {
			current = e
		}
	}
	latest := ""
	var previous TimelineEvent
	for _, e := range s.Events {
		for _, snap := range e.DocumentSnapshots {
			if snap.DocumentKey == r.DocumentKey {
				latest = snap.DocumentRevisionID
				previous = e
			}
		}
	}
	if r.PredecessorRevisionID != latest {
		return timelineError("predecessor_revision_id", "current_predecessor_required")
	}
	if latest != "" && previous.EventID != eventID && (previous.Stage != current.Stage || current.RelatedEventID != previous.EventID) {
		return timelineError("document_key", "stage_chain_mismatch")
	}
	return nil
}

func timelineWaitAssignee(s TimelineSession, taskID, segmentID string) string {
	for _, t := range s.Tasks {
		if t.TaskID == taskID && !t.Withdrawn {
			for _, seg := range t.Segments {
				if seg.SegmentID == segmentID && !seg.Withdrawn && seg.PlannedStage == "delivery" {
					return seg.AssignedProducerID
				}
			}
		}
	}
	return ""
}

// ValidateTimelineResolvedProject checks a trusted resolver result, never a path
// equality or a client assertion. resolvedProjectKey is contextual, non-JSON
// input obtained by the server's registered-root resolver (Store.timelineProject).
// Preserve Source's actual tuple, including a distinct alias/worktree root.
//
// Required T4/T5 boundaries:
//   - Create: derive the work ProjectKey from resolving the original Source.
//   - Join: resolve the requesting Source before persisting a pending request.
//   - Accept/recover: resolve that stored pending Source again before granting
//     membership, in the same conditional transaction as acceptance/transfer.
//   - Load: resolve the original Source, every accepted producer Source and each
//     pending request Source before accepting restored authority. An unavailable
//     or changed identity must fail without writing; do not trust stored paths.
//
// Within that canonical work, deduplicate host/native-session identities across
// alias roots and retain the first actual Source unchanged.
// Pure graph/authority validators trust membership established at those boundaries;
// they neither access the registry nor infer project identity from root spelling.
// Never deserialize resolvedProjectKey from a publisher request.
func ValidateTimelineResolvedProject(s TimelineSession, source TimelineSource, resolvedProjectKey string) error {
	if err := timelineSource(source); err != nil {
		return err
	}
	if resolvedProjectKey == "" {
		return timelineError("source", "project_identity_unavailable")
	}
	if resolvedProjectKey != s.ProjectKey {
		return timelineError("source", "foreign_project")
	}
	return nil
}

// A work has one server-resolved project identity. Roots describe actual source
// provenance, but never split (project, host, native session) membership.
type timelineNativeIdentity struct {
	Host            TimelineHost
	NativeSessionID string
}

func timelineNativeSourceIdentity(source TimelineSource) timelineNativeIdentity {
	return timelineNativeIdentity{Host: source.Host, NativeSessionID: source.NativeSessionID}
}
