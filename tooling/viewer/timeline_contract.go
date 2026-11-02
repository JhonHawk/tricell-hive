package viewer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// TimelineSchemaVersion versions the API contract independently of legacy manifests.
const TimelineSchemaVersion = 3
const (
	TimelineMaxBody                         = 2 << 20
	TimelineContentBudget                   = 7 << 18
	TimelineMaxEvents                       = 508
	TimelineContentEvents                   = 500
	TimelineMaxViews                        = 50
	TimelineMaxImageRefs                    = 100
	TimelineMaxReceipts                     = 1000
	TimelineContentReceipts                 = 992
	TimelineMaxDocumentBytes                = 256 << 10
	TimelineMaxDocumentTabs                 = 8
	TimelineMaxDocuments                    = 64
	TimelineMaxDocumentSnapshots            = 50
	TimelineMaxAggregateDocumentBytes       = 16 << 20
	TimelineMaxNumber                 int64 = 2147483647
)

type TimelineHost string
type TimelineState string
type TimelineEventKind string
type TimelineStage string
type TimelineStageState string

const (
	TimelineActive             TimelineState = "active"
	TimelineInReview           TimelineState = "in_review"
	TimelineFailed             TimelineState = "failed"
	TimelineAwaitingValidation TimelineState = "awaiting_validation"
	TimelinePaused             TimelineState = "paused"
	TimelineClosed             TimelineState = "closed"
)

var timelineHosts = []string{"codex", "claude-code", "grok", "pi", "opencode", "cursor"}
var timelineStates = []string{"active", "in_review", "awaiting_validation", "paused", "closed", "failed"}
var timelineViewports = []string{"desktop", "mobile"}
var timelineThemes = []string{"light", "dark", "unspecified"}
var timelineImageMediaTypes = []string{"image/png", "image/jpeg", "image/webp"}
var timelineEventKinds = []string{"request", "work_started", "stage", "validation", "changes_reported", "iteration", "approval_reported", "delivery", "closure", "review", "state_transition", "producer_joined", "foreground_transferred", "task_state_transition", "local_action"}
var timelineOpaqueID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$`)
var timelineOperationID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:-]{0,127}$`)
var timelineSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)
var timelineTimestampPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,9})?(Z|[+-]\d{2}:\d{2})$`)

// TimelineContractError reports schema facts, never caller-supplied values.
type TimelineContractError struct {
	Field string
	Code  string
}

func (e *TimelineContractError) Error() string { return e.Field + ": " + e.Code }
func timelineError(field, code string) error   { return &TimelineContractError{Field: field, Code: code} }

type TimelineSource struct {
	ProjectRoot     string       `json:"project_root"`
	Host            TimelineHost `json:"host"`
	NativeSessionID string       `json:"native_session_id"`
}
type TimelineTicketInput struct {
	ExternalScope string `json:"external_scope"`
	ExternalID    string `json:"external_id"`
	Key           string `json:"key"`
	Title         string `json:"title,omitempty"`
	URL           string `json:"url,omitempty"`
}
type TimelineTicket struct {
	TicketID string `json:"ticket_id"`
	TimelineTicketInput
}
type TimelineTicketSnapshot struct {
	Key   string `json:"key"`
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}
type TimelineTicketRef struct {
	TicketID string                 `json:"ticket_id"`
	Snapshot TimelineTicketSnapshot `json:"snapshot"`
}
type TimelineCandidate struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	Base   string `json:"base"`
}
type TimelineCandidateRecord struct {
	CandidateID string `json:"candidate_id"`
	TimelineCandidate
}

var timelineStages = []string{"research", "plan", "build", "verify", "delivery"}
var timelineStageStates = []string{"in_progress", "in_review", "awaiting_validation", "completed", "paused", "failed"}

type TimelineDecisionTarget struct {
	Kind               string   `json:"kind"`
	EventID            string   `json:"event_id"`
	EvidenceID         string   `json:"evidence_id,omitempty"`
	ViewIDs            []string `json:"view_ids,omitempty"`
	DocumentRevisionID string   `json:"document_revision_id,omitempty"`
}
type TimelineEvidenceTarget = TimelineDecisionTarget
type TimelineDocument struct {
	Actor      TimelineActor `json:"actor"`
	DocumentID string        `json:"document_id"`
	SessionID  string        `json:"session_id"`
	SHA256     string        `json:"sha256"`
	Size       int64         `json:"size"`
	MediaType  string        `json:"media_type"`
	ReceivedAt string        `json:"received_at"`
}
type TimelineDocumentTab struct {
	Label      string `json:"label"`
	DocumentID string `json:"document_id"`
}
type TimelineDocumentSnapshot struct {
	TaskID                string                `json:"task_id,omitempty"`
	SegmentID             string                `json:"segment_id,omitempty"`
	Actor                 TimelineActor         `json:"actor"`
	DocumentKey           string                `json:"document_key"`
	PredecessorRevisionID string                `json:"predecessor_revision_id,omitempty"`
	DocumentRevisionID    string                `json:"document_revision_id"`
	Status                string                `json:"status"`
	CreatedAt             string                `json:"created_at"`
	Tabs                  []TimelineDocumentTab `json:"tabs"`
}
type TimelineDocumentResponse struct {
	SchemaVersion int              `json:"schema_version"`
	Document      TimelineDocument `json:"document"`
	Content       string           `json:"content"`
}
type TimelineDocumentMetadata struct {
	Actor         TimelineActor `json:"actor"`
	SchemaVersion int           `json:"schema_version"`
	OperationID   string        `json:"operation_id"`
	MediaType     string        `json:"media_type"`
}
type TimelineDocumentSnapshotRequest struct {
	DocumentKey            string                `json:"document_key"`
	PredecessorRevisionID  string                `json:"predecessor_revision_id,omitempty"`
	ExpectedSessionVersion int64                 `json:"expected_session_version"`
	TaskID                 string                `json:"task_id,omitempty"`
	SegmentID              string                `json:"segment_id,omitempty"`
	Actor                  TimelineActor         `json:"actor"`
	SchemaVersion          int                   `json:"schema_version"`
	OperationID            string                `json:"operation_id"`
	ExpectedVersion        int64                 `json:"expected_version"`
	Status                 string                `json:"status"`
	Tabs                   []TimelineDocumentTab `json:"tabs"`
}
type TimelineSupersedes struct {
	EventID    string `json:"event_id"`
	EvidenceID string `json:"evidence_id"`
	ViewID     string `json:"view_id"`
}
type TimelineImageRef struct {
	ArtifactID string `json:"artifact_id"`
	Caption    string `json:"caption"`
	Alt        string `json:"alt,omitempty"`
}
type TimelineView struct {
	ViewID         string              `json:"view_id"`
	Title          string              `json:"title"`
	TicketRefs     []TimelineTicketRef `json:"ticket_refs"`
	Steps          []string            `json:"steps"`
	ExpectedResult string              `json:"expected_result"`
	ImageRefs      []TimelineImageRef  `json:"image_refs"`
	Supersedes     *TimelineSupersedes `json:"supersedes,omitempty"`
}

// TimelineViewInput permits an omitted view_id only for a server-assigned new view.
type TimelineViewInput struct {
	ViewID         string              `json:"view_id,omitempty"`
	Title          string              `json:"title"`
	TicketIDs      []string            `json:"ticket_ids"`
	Steps          []string            `json:"steps"`
	ExpectedResult string              `json:"expected_result"`
	ImageRefs      []TimelineImageRef  `json:"image_refs"`
	Supersedes     *TimelineSupersedes `json:"supersedes,omitempty"`
}
type TimelineDeliverySnapshot struct {
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	URL    string `json:"url"`
	Status string `json:"status"`
}
type TimelineDelivery struct {
	DeliveryID    string `json:"delivery_id"`
	RepositoryURL string `json:"repository_url"`
	ExternalID    string `json:"external_id"`
	TimelineDeliverySnapshot
}
type TimelineDeliveryReference struct {
	DeliveryID string                   `json:"delivery_id"`
	Snapshot   TimelineDeliverySnapshot `json:"snapshot"`
}
type TimelineClosure struct {
	Reason string `json:"reason"`
	Result string `json:"result"`
}
type TimelineEvent struct {
	StageEventID      string                      `json:"stage_event_id,omitempty"`
	TaskID            string                      `json:"task_id,omitempty"`
	SegmentID         string                      `json:"segment_id,omitempty"`
	Reason            string                      `json:"reason,omitempty"`
	ReviewStatus      string                      `json:"review_status,omitempty"`
	Audit             *TimelineAudit              `json:"audit,omitempty"`
	Actor             TimelineActor               `json:"actor"`
	EventID           string                      `json:"event_id"`
	Sequence          int64                       `json:"sequence"`
	Version           int64                       `json:"version"`
	Kind              TimelineEventKind           `json:"kind"`
	Title             string                      `json:"title"`
	Summary           string                      `json:"summary"`
	CreatedAt         string                      `json:"created_at"`
	UpdatedAt         string                      `json:"updated_at"`
	Origin            string                      `json:"origin"`
	TicketRefs        []TimelineTicketRef         `json:"ticket_refs"`
	Withdrawn         bool                        `json:"withdrawn,omitempty"`
	CandidateID       string                      `json:"candidate_id,omitempty"`
	EvidenceID        string                      `json:"evidence_id,omitempty"`
	Views             []TimelineView              `json:"views,omitempty"`
	RelatedEventID    string                      `json:"related_event_id,omitempty"`
	Target            *TimelineEvidenceTarget     `json:"target,omitempty"`
	SourceReference   string                      `json:"source_reference,omitempty"`
	DeliveryRefs      []TimelineDeliveryReference `json:"delivery_refs,omitempty"`
	Closure           *TimelineClosure            `json:"closure,omitempty"`
	Stage             TimelineStage               `json:"stage,omitempty"`
	StageState        TimelineStageState          `json:"stage_state,omitempty"`
	StartedAt         string                      `json:"started_at,omitempty"`
	FinishedAt        string                      `json:"finished_at,omitempty"`
	OccurredAt        string                      `json:"occurred_at,omitempty"`
	DocumentSnapshots []TimelineDocumentSnapshot  `json:"document_snapshots,omitempty"`
}
type TimelineArtifact struct {
	Actor       TimelineActor `json:"actor"`
	ArtifactID  string        `json:"artifact_id"`
	SessionID   string        `json:"session_id"`
	SHA256      string        `json:"sha256"`
	Size        int64         `json:"size"`
	MediaType   string        `json:"media_type"`
	ReceivedAt  string        `json:"received_at"`
	CapturedAt  string        `json:"captured_at"`
	CandidateID string        `json:"candidate_id"`
	Viewport    string        `json:"viewport"`
	Theme       string        `json:"theme"`
}
type TimelineSession struct {
	Deliveries           []TimelineDelivery        `json:"deliveries"`
	Candidates           []TimelineCandidateRecord `json:"candidates"`
	Producers            []TimelineProducer        `json:"producers"`
	ForegroundProducerID string                    `json:"foreground_producer_id"`
	PendingJoinRequests  []TimelineJoinRequest     `json:"pending_join_requests"`
	Tasks                []TimelineTask            `json:"tasks"`
	Waits                []TimelineCIWait          `json:"waits"`
	Reason               string                    `json:"reason,omitempty"`
	SchemaVersion        int                       `json:"schema_version"`
	ViewerSessionID      string                    `json:"viewer_session_id"`
	ProjectKey           string                    `json:"project_key"`
	ProjectName          string                    `json:"project_name"`
	Source               TimelineSource            `json:"source"`
	Version              int64                     `json:"version"`
	Title                string                    `json:"title"`
	Summary              string                    `json:"summary"`
	State                TimelineState             `json:"state"`
	Archived             bool                      `json:"archived"`
	Pinned               bool                      `json:"pinned"`
	CreatedAt            string                    `json:"created_at"`
	UpdatedAt            string                    `json:"updated_at"`
	ValidationCount      int                       `json:"validation_count"`
	Tickets              []TimelineTicket          `json:"tickets"`
	Events               []TimelineEvent           `json:"events"`
	Artifacts            []TimelineArtifact        `json:"artifacts"`
	Documents            []TimelineDocument        `json:"documents"`
	CurrentStageEventID  *string                   `json:"current_stage_event_id" timeline:"nullable"`
}

// TimelineAPIIndexEntry is the API v3 work summary.
type TimelineAPIIndexEntry struct {
	CurrentStageEventID *string          `json:"current_stage_event_id" timeline:"nullable"`
	Mode                string           `json:"mode"`
	ViewerSessionID     string           `json:"viewer_session_id"`
	ProjectKey          string           `json:"project_key"`
	ProjectName         string           `json:"project_name"`
	Source              TimelineSource   `json:"source"`
	Version             int64            `json:"version"`
	Title               string           `json:"title"`
	Summary             string           `json:"summary"`
	State               TimelineState    `json:"state"`
	Archived            bool             `json:"archived"`
	Pinned              bool             `json:"pinned"`
	CreatedAt           string           `json:"created_at"`
	UpdatedAt           string           `json:"updated_at"`
	ValidationCount     int              `json:"validation_count"`
	Tickets             []TimelineTicket `json:"tickets"`
}

// TimelineIndexEntry decoders accept only timeline_api v3. The legacy wrapper
// below remains solely for the old runtime to compile until its T4 removal.
type TimelineIndexEntry interface{ timelineIndexMode() string }

func (TimelineAPIIndexEntry) timelineIndexMode() string { return "timeline_api" }

type TimelineLegacySession = indexSession
type TimelineLegacyIndexEntry struct {
	Mode    string                `json:"mode"`
	Session TimelineLegacySession `json:"session"`
}

func (TimelineLegacyIndexEntry) timelineIndexMode() string { return "legacy" }

type TimelineOperationKind string

// TimelineOperationResult contains public IDs/versions only. Request digests and
// completed operation receipts remain private persistence records.
type TimelineOperationResult struct {
	TicketID            string                `json:"ticket_id,omitempty"`
	DeliveryID          string                `json:"delivery_id,omitempty"`
	CandidateID         string                `json:"candidate_id,omitempty"`
	RequestID           string                `json:"request_id,omitempty"`
	ProducerID          string                `json:"producer_id,omitempty"`
	TaskID              string                `json:"task_id,omitempty"`
	TaskVersion         int64                 `json:"task_version,omitempty"`
	WaitID              string                `json:"wait_id,omitempty"`
	WaitVersion         int64                 `json:"wait_version,omitempty"`
	AttemptID           string                `json:"attempt_id,omitempty"`
	Generation          int64                 `json:"generation,omitempty"`
	ObservationID       string                `json:"observation_id,omitempty"`
	CurrentStageEventID *string               `json:"current_stage_event_id" timeline:"nullable"`
	DocumentID          string                `json:"document_id,omitempty"`
	DocumentRevisionID  string                `json:"document_revision_id,omitempty"`
	SHA256              string                `json:"sha256,omitempty"`
	Size                *int64                `json:"size,omitempty"`
	SchemaVersion       int                   `json:"schema_version"`
	OperationID         string                `json:"operation_id"`
	Kind                TimelineOperationKind `json:"kind"`
	ViewerSessionID     string                `json:"viewer_session_id"`
	SessionVersion      int64                 `json:"session_version"`
	EventID             string                `json:"event_id,omitempty"`
	EventVersion        int64                 `json:"event_version,omitempty"`
	Sequence            int64                 `json:"sequence,omitempty"`
	EvidenceID          string                `json:"evidence_id,omitempty"`
	ViewIDs             []string              `json:"view_ids,omitempty"`
	ArtifactID          string                `json:"artifact_id,omitempty"`
	Created             *bool                 `json:"created,omitempty"`
	Withdrawn           *bool                 `json:"withdrawn,omitempty"`
	Deleted             *bool                 `json:"deleted,omitempty"`
}
type TimelineOperationReceipt struct {
	OperationID   string                  `json:"operation_id"`
	RequestSHA256 string                  `json:"request_sha256"`
	CompletedAt   string                  `json:"completed_at"`
	Result        TimelineOperationResult `json:"result"`
}

type TimelineSessionResponse struct {
	SchemaVersion int               `json:"schema_version"`
	Mode          string            `json:"mode"`
	Session       TimelineSession   `json:"session"`
	ArtifactURLs  map[string]string `json:"artifact_urls,omitempty"`
	DocumentURLs  map[string]string `json:"document_urls,omitempty"`
	Created       *bool             `json:"created,omitempty"`
}
type TimelineCreateSessionRequest struct {
	SchemaVersion int                   `json:"schema_version"`
	OperationID   string                `json:"operation_id"`
	Source        TimelineSource        `json:"source"`
	Title         string                `json:"title"`
	Summary       string                `json:"summary"`
	Tickets       []TimelineTicketInput `json:"tickets"`
}
type TimelineSessionPatchRequest struct {
	Reason          string         `json:"reason,omitempty"`
	Actor           TimelineActor  `json:"actor"`
	SchemaVersion   int            `json:"schema_version"`
	OperationID     string         `json:"operation_id"`
	ExpectedVersion int64          `json:"expected_version"`
	Title           *string        `json:"title,omitempty"`
	Summary         *string        `json:"summary,omitempty"`
	State           *TimelineState `json:"state,omitempty"`
	Archived        *bool          `json:"archived,omitempty"`
	Pinned          *bool          `json:"pinned,omitempty"`
}
type TimelineMutationRequest struct {
	Actor           TimelineActor `json:"actor"`
	SchemaVersion   int           `json:"schema_version"`
	OperationID     string        `json:"operation_id"`
	ExpectedVersion int64         `json:"expected_version"`
}
type TimelineEventPatchRequest struct {
	Reason                 string              `json:"reason,omitempty"`
	Actor                  TimelineActor       `json:"actor"`
	StageState             *TimelineStageState `json:"stage_state,omitempty"`
	ExpectedSessionVersion *int64              `json:"expected_session_version,omitempty"`
	SchemaVersion          int                 `json:"schema_version"`
	OperationID            string              `json:"operation_id"`
	ExpectedVersion        int64               `json:"expected_version"`
	Title                  *string             `json:"title,omitempty"`
	Summary                *string             `json:"summary,omitempty"`
}
type TimelineAppendEventRequest struct {
	TargetStageEventID     string                  `json:"target_stage_event_id,omitempty"`
	TaskID                 string                  `json:"task_id,omitempty"`
	SegmentID              string                  `json:"segment_id,omitempty"`
	ReviewStatus           string                  `json:"review_status,omitempty"`
	Reason                 string                  `json:"reason,omitempty"`
	Actor                  TimelineActor           `json:"actor"`
	Stage                  TimelineStage           `json:"stage,omitempty"`
	StageState             TimelineStageState      `json:"stage_state,omitempty"`
	ExpectedSessionVersion *int64                  `json:"expected_session_version,omitempty"`
	OccurredAt             string                  `json:"occurred_at,omitempty"`
	SchemaVersion          int                     `json:"schema_version"`
	OperationID            string                  `json:"operation_id"`
	Kind                   TimelineEventKind       `json:"kind"`
	Title                  string                  `json:"title"`
	Summary                string                  `json:"summary"`
	TicketIDs              []string                `json:"ticket_ids"`
	CandidateID            string                  `json:"candidate_id,omitempty"`
	Views                  []TimelineViewInput     `json:"views,omitempty"`
	RelatedEventID         string                  `json:"related_event_id,omitempty"`
	Target                 *TimelineEvidenceTarget `json:"target,omitempty"`
	SourceReference        string                  `json:"source_reference,omitempty"`
	DeliveryIDs            []string                `json:"delivery_ids,omitempty"`
	Closure                *TimelineClosure        `json:"closure,omitempty"`
}
type TimelineArtifactMetadata struct {
	Actor         TimelineActor `json:"actor"`
	SchemaVersion int           `json:"schema_version"`
	OperationID   string        `json:"operation_id"`
	CapturedAt    string        `json:"captured_at"`
	CandidateID   string        `json:"candidate_id"`
	Viewport      string        `json:"viewport"`
	Theme         string        `json:"theme"`
	MediaType     string        `json:"media_type"`
}

// DecodeTimelineJSON rejects duplicate keys (including escaped aliases), unknown
// fields, nulls and trailing values before any runtime-boundary validation.
func DecodeTimelineJSON(raw []byte, target any) error {
	return decodeTimelineJSONBounded(raw, target, TimelineMaxBody)
}

// decodeTimelineJSONBounded shares exact-field, duplicate-key and encoding
// validation between HTTP payloads and the larger private receipt envelope.
func decodeTimelineJSONBounded(raw []byte, target any, maxBytes int) error {
	if len(raw) == 0 || len(raw) > maxBytes || !utf8.Valid(raw) {
		return timelineError("body", "invalid_size_or_encoding")
	}
	scan := json.NewDecoder(bytes.NewReader(raw))
	scan.UseNumber()
	if err := timelineJSONValue(scan, 0); err != nil {
		return err
	}
	if _, err := scan.Token(); err != io.EOF {
		return timelineError("body", "trailing_data")
	}
	targetType := reflect.TypeOf(target)
	if targetType == nil || targetType.Kind() != reflect.Pointer || reflect.ValueOf(target).IsNil() {
		return timelineError("body", "invalid_decode_target")
	}
	var envelope map[string]json.RawMessage
	if json.Unmarshal(raw, &envelope) == nil {
		if version, ok := envelope["schema_version"]; ok && string(version) != "3" {
			return timelineError("schema_version", "unsupported_protocol")
		}
	}
	if err := timelineRequiredFields(raw, targetType.Elem()); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return timelineError("body", "invalid_fields_or_types")
	}
	return nil
}
func timelineJSONFields(t reflect.Type) []reflect.StructField {
	var fields []reflect.StructField
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous && field.Tag.Get("json") == "" {
			fields = append(fields, timelineJSONFields(field.Type)...)
		} else {
			fields = append(fields, field)
		}
	}
	return fields
}
func timelineRequiredFields(raw []byte, t reflect.Type) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return timelineError("body", "invalid_fields_or_types")
		}
		declared := timelineJSONFields(t)
		allowed := map[string]bool{}
		for _, field := range declared {
			name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name != "" && name != "-" {
				allowed[name] = true
			}
		}
		// encoding/json accepts case-insensitive aliases. Check exact tag names
		// independently so a second differently-cased key can never overwrite one.
		for name := range fields {
			if !allowed[name] {
				return timelineError("body", "unknown_field")
			}
		}
		legacy := t == reflect.TypeOf(indexSession{})
		legacyRequired := map[string]bool{"key": true, "title": true, "session_id": true, "source": true, "content_hash": true}
		for _, field := range declared {
			name, options, _ := strings.Cut(field.Tag.Get("json"), ",")
			if name == "-" || name == "" {
				continue
			}
			value, ok := fields[name]
			if !ok {
				if (!legacy && !strings.Contains(options, "omitempty")) || (legacy && legacyRequired[name]) {
					return timelineError(name, "required_field")
				}
				continue
			}
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				if field.Tag.Get("timeline") == "nullable" {
					continue
				}
				return timelineError(name, "null_not_allowed")
			}
			if !legacy && !(name == "title" && (t == reflect.TypeOf(TimelineTicketInput{}) || t == reflect.TypeOf(TimelineTicketSnapshot{}) || t == reflect.TypeOf(TimelineTicketCreateRequest{}) || t == reflect.TypeOf(TimelineTicket{}))) && field.Type.Kind() == reflect.String && strings.Contains(options, "omitempty") && bytes.Equal(value, []byte(`""`)) {
				return timelineError(name, "empty_optional_field")
			}
			if err := timelineRequiredFields(value, field.Type); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if t.Elem().Kind() != reflect.Struct && t.Elem().Kind() != reflect.Pointer {
			return nil
		}
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return timelineError("body", "invalid_fields_or_types")
		}
		for _, value := range values {
			if err := timelineRequiredFields(value, t.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
func timelineJSONValue(dec *json.Decoder, depth int, nullable ...bool) error {
	if depth > 32 {
		return timelineError("body", "too_deep")
	}
	token, err := dec.Token()
	if err != nil {
		return timelineError("body", "invalid_json")
	}
	if token == nil {
		if len(nullable) > 0 && nullable[0] {
			return nil
		}
		return timelineError("body", "null_not_allowed")
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for dec.More() {
			key, err := dec.Token()
			if err != nil {
				return timelineError("body", "invalid_json")
			}
			name, ok := key.(string)
			if !ok || seen[name] {
				return timelineError("body", "duplicate_key")
			}
			seen[name] = true
			if err := timelineJSONValue(dec, depth+1, name == "current_stage_event_id"); err != nil {
				return err
			}
		}
	case '[':
		for dec.More() {
			if err := timelineJSONValue(dec, depth+1); err != nil {
				return err
			}
		}
	default:
		return timelineError("body", "invalid_json")
	}
	end, err := dec.Token()
	if err != nil || (delim == '{' && end != json.Delim('}')) || (delim == '[' && end != json.Delim(']')) {
		return timelineError("body", "invalid_json")
	}
	return nil
}
func timelineText(value, field string, limit int, required bool) error {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) > limit || required && strings.TrimSpace(value) == "" {
		return timelineError(field, "invalid_text")
	}
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return timelineError(field, "invalid_text")
		}
	}
	return nil
}
func timelineNumber(n int64, field string) error {
	if n < 1 || n > TimelineMaxNumber {
		return timelineError(field, "invalid_positive_integer")
	}
	return nil
}
func timelineID(id, field string) error {
	if !timelineOpaqueID.MatchString(id) {
		return timelineError(field, "invalid_id")
	}
	return nil
}
func timelineTime(value, field string) error {
	if !timelineTimestampPattern.MatchString(value) {
		return timelineError(field, "invalid_timestamp")
	}
	if !strings.HasSuffix(value, "Z") {
		zone := value[len(value)-6:]
		if zone[1:3] > "23" || zone[4:6] > "59" {
			return timelineError(field, "invalid_timestamp")
		}
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return timelineError(field, "invalid_timestamp")
	}
	return nil
}
func timelineOperation(schema int, id string) error {
	if schema != TimelineSchemaVersion {
		return timelineError("schema_version", "unsupported_protocol")
	}
	if !timelineOperationID.MatchString(id) {
		return timelineError("operation_id", "invalid_id")
	}
	return nil
}
func timelineSource(source TimelineSource) error {
	if err := timelineText(source.ProjectRoot, "source.project_root", 4096, true); err != nil {
		return err
	}
	if !filepath.IsAbs(source.ProjectRoot) || strings.IndexByte(source.ProjectRoot, 0) >= 0 {
		return timelineError("source.project_root", "absolute_root_required")
	}
	if !contains(timelineHosts, string(source.Host)) {
		return timelineError("source.host", "unsupported_host")
	}
	if !utf8.ValidString(source.NativeSessionID) || utf8.RuneCountInString(source.NativeSessionID) > 256 || source.NativeSessionID == "" {
		return timelineError("source.native_session_id", "native_id_required")
	}
	for _, r := range source.NativeSessionID {
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return timelineError("source.native_session_id", "native_id_required")
		}
	}
	return nil
}
func timelineHTTPS(raw, field string) error {
	if utf8.RuneCountInString(raw) > 4096 || (!safeDeliveryURL(raw) || !timelineURLHost(raw)) {
		return timelineError(field, "invalid_https_url")
	}
	return nil
}
func TimelineTicketIdentity(ref TimelineTicketRef) string { return ref.TicketID }
func timelineTicketRef(ref TimelineTicketRef, field string) error {
	if e := timelineID(ref.TicketID, field+".ticket_id"); e != nil {
		return e
	}
	if e := timelineText(ref.Snapshot.Key, field+".snapshot.key", 100, true); e != nil {
		return e
	}
	if e := timelineText(ref.Snapshot.Title, field+".snapshot.title", 300, false); e != nil {
		return e
	}
	if ref.Snapshot.URL != "" {
		return timelineHTTPS(ref.Snapshot.URL, field+".snapshot.url")
	}
	return nil
}
func timelineTickets(tickets []TimelineTicket) error {
	if tickets == nil || len(tickets) > 100 {
		return timelineError("tickets", "array_required_or_limit")
	}
	for _, t := range tickets {
		if e := timelineID(t.TicketID, "ticket_id"); e != nil {
			return e
		}
		if e := timelineTicketInput(t.TimelineTicketInput); e != nil {
			return e
		}
	}
	return nil
}
func timelineTicketBootstrap(tickets []TimelineTicketInput) error {
	if tickets == nil || len(tickets) > 100 {
		return timelineError("tickets", "array_required_or_limit")
	}
	seen := map[string]bool{}
	urls := map[string]bool{}
	for _, t := range tickets {
		if e := timelineTicketInput(t); e != nil {
			return e
		}
		scope, _ := CanonicalTimelineIdentityURL(t.ExternalScope)
		key := scope + "\x00" + t.ExternalID
		if seen[key] {
			return timelineError("tickets", "duplicate_identity")
		}
		seen[key] = true
		if t.URL != "" {
			url := timelineLocator(t.URL)
			if urls[url] {
				return timelineError("tickets", "entity_conflict")
			}
			urls[url] = true
		}
	}
	return nil
}
func timelineRefs(refs []TimelineTicketRef, field string) error {
	if refs == nil || len(refs) > 100 {
		return timelineError(field, "array_required_or_limit")
	}
	seen := map[string]bool{}
	for _, ref := range refs {
		if e := timelineTicketRef(ref, field); e != nil {
			return e
		}
		if seen[ref.TicketID] {
			return timelineError(field, "duplicate_id")
		}
		seen[ref.TicketID] = true
	}
	return nil
}
func timelineCandidate(c TimelineCandidate) error {
	for _, v := range []struct{ field, value string }{{"branch", c.Branch}, {"commit", c.Commit}, {"base", c.Base}} {
		if err := timelineText(v.value, "candidate."+v.field, 256, true); err != nil {
			return err
		}
	}
	return nil
}
func timelineTarget(t TimelineEvidenceTarget) error {
	if err := timelineID(t.EventID, "target.event_id"); err != nil {
		return err
	}
	switch t.Kind {
	case "document":
		if t.EvidenceID != "" || t.ViewIDs != nil {
			return timelineError("target", "mixed_target")
		}
		return timelineID(t.DocumentRevisionID, "target.document_revision_id")
	case "evidence":
		if t.DocumentRevisionID != "" {
			return timelineError("target", "mixed_target")
		}
		if err := timelineID(t.EvidenceID, "target.evidence_id"); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, id := range t.ViewIDs {
			if err := timelineID(id, "target.view_ids"); err != nil {
				return err
			}
			if seen[id] {
				return timelineError("target.view_ids", "duplicate_id")
			}
			seen[id] = true
		}
		if len(t.ViewIDs) > TimelineMaxViews {
			return timelineError("target.view_ids", "limit")
		}
		return nil
	default:
		return timelineError("target.kind", "unsupported_target")
	}
}
func timelineView(v TimelineViewInput) error {
	if v.ViewID != "" {
		if err := timelineID(v.ViewID, "views.view_id"); err != nil {
			return err
		}
	}
	if err := timelineText(v.Title, "views.title", 200, true); err != nil {
		return err
	}
	if err := timelineIDs(v.TicketIDs, "views.ticket_ids", false); err != nil {
		return err
	}
	if v.Steps == nil || len(v.Steps) == 0 || len(v.Steps) > 100 {
		return timelineError("views.steps", "nonempty_array_required_or_limit")
	}
	for _, step := range v.Steps {
		if err := timelineText(step, "views.steps", 2000, true); err != nil {
			return err
		}
	}
	if err := timelineText(v.ExpectedResult, "views.expected_result", 4000, true); err != nil {
		return err
	}
	if v.ImageRefs == nil || len(v.ImageRefs) == 0 || len(v.ImageRefs) > TimelineMaxImageRefs {
		return timelineError("views.image_refs", "nonempty_array_required_or_limit")
	}
	seen := map[string]bool{}
	for _, image := range v.ImageRefs {
		if err := timelineID(image.ArtifactID, "views.image_refs.artifact_id"); err != nil {
			return err
		}
		if err := timelineText(image.Caption, "views.image_refs.caption", 1000, true); err != nil {
			return err
		}
		if err := timelineText(image.Alt, "views.image_refs.alt", 2000, image.Alt != ""); err != nil {
			return err
		}
		if seen[image.ArtifactID] {
			return timelineError("views.image_refs", "duplicate_artifact")
		}
		seen[image.ArtifactID] = true
	}
	if v.Supersedes != nil {
		if err := timelineTarget(TimelineEvidenceTarget{Kind: "evidence", EventID: v.Supersedes.EventID, EvidenceID: v.Supersedes.EvidenceID, ViewIDs: []string{v.Supersedes.ViewID}}); err != nil {
			return err
		}
		if v.ViewID != "" && v.ViewID != v.Supersedes.ViewID {
			return timelineError("views.supersedes", "view_mismatch")
		}
	}
	return nil
}

func DecodeTimelineCreateSession(raw []byte) (TimelineCreateSessionRequest, error) {
	var r TimelineCreateSessionRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineCreateSessionRequest) Validate() error {
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := timelineSource(r.Source); err != nil {
		return err
	}
	if err := timelineText(r.Title, "title", 200, true); err != nil {
		return err
	}
	if err := timelineText(r.Summary, "summary", 4000, true); err != nil {
		return err
	}
	return timelineTicketBootstrap(r.Tickets)
}
func DecodeTimelineSessionPatch(raw []byte) (TimelineSessionPatchRequest, error) {
	var r TimelineSessionPatchRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineSessionPatchRequest) Validate() error {
	if err := validateTimelineSessionPatchActor(r); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := timelineNumber(r.ExpectedVersion, "expected_version"); err != nil {
		return err
	}
	if r.Title == nil && r.Summary == nil && r.State == nil && r.Archived == nil && r.Pinned == nil {
		return timelineError("body", "empty_patch")
	}
	if r.Title != nil {
		if err := timelineText(*r.Title, "title", 200, true); err != nil {
			return err
		}
	}
	if r.Summary != nil {
		if err := timelineText(*r.Summary, "summary", 4000, true); err != nil {
			return err
		}
	}
	if r.State != nil && !contains(timelineStates, string(*r.State)) {
		return timelineError("state", "unsupported_state")
	}

	return nil
}
func DecodeTimelineEventPatch(raw []byte) (TimelineEventPatchRequest, error) {
	var r TimelineEventPatchRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineEventPatchRequest) Validate() error {
	if err := validateTimelineNativeActor(r.Actor); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := timelineNumber(r.ExpectedVersion, "expected_version"); err != nil {
		return err
	}
	if r.Title == nil && r.Summary == nil && r.StageState == nil {
		return timelineError("body", "empty_patch")
	}
	if r.Title != nil {
		if err := timelineText(*r.Title, "title", 200, false); err != nil {
			return err
		}
	}
	if r.Summary != nil {
		if err := timelineText(*r.Summary, "summary", 4000, true); err != nil {
			return err
		}
	}
	if r.StageState != nil {
		if err := timelineReason(r.Reason); err != nil {
			return err
		}
		if !contains(timelineStageStates, string(*r.StageState)) {
			return timelineError("stage_state", "unsupported_stage_state")
		}
		if r.ExpectedSessionVersion == nil {
			return timelineError("expected_session_version", "required_for_stage")
		}
		return timelineNumber(*r.ExpectedSessionVersion, "expected_session_version")
	}
	if r.ExpectedSessionVersion != nil {
		return timelineError("expected_session_version", "forbidden_without_stage_transition")
	}
	return nil
}
func DecodeTimelineMutation(raw []byte) (TimelineMutationRequest, error) {
	var r TimelineMutationRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineMutationRequest) Validate() error {
	if err := r.Actor.Validate(); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	return timelineNumber(r.ExpectedVersion, "expected_version")
}
func DecodeTimelineArtifactMetadata(raw []byte) (TimelineArtifactMetadata, error) {
	var r TimelineArtifactMetadata
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineArtifactMetadata) Validate() error {
	if err := validateTimelineNativeActor(r.Actor); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := timelineTime(r.CapturedAt, "captured_at"); err != nil {
		return err
	}
	if err := timelineID(r.CandidateID, "candidate_id"); err != nil {
		return err
	}
	if !contains(timelineViewports, r.Viewport) {
		return timelineError("viewport", "unsupported_viewport")
	}
	if !contains(timelineThemes, r.Theme) {
		return timelineError("theme", "unsupported_theme")
	}
	if !contains(timelineImageMediaTypes, r.MediaType) {
		return timelineError("media_type", "unsupported_media_type")
	}
	return nil
}
func DecodeTimelineAppendEvent(raw []byte) (TimelineAppendEventRequest, error) {
	var r TimelineAppendEventRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func (r TimelineAppendEventRequest) Validate() error {
	return r.validate(r.Kind == "stage")
}

// Stored stages reuse the common payload checks after their lifecycle fields are
// validated separately; retain their original title-context rule in that path.
func (r TimelineAppendEventRequest) validate(stageContext bool) error {
	if err := validateTimelineAppendV3(r); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if !contains(timelineEventKinds, string(r.Kind)) {
		return timelineError("kind", "unsupported_kind")
	}
	if err := timelineText(r.Title, "title", 200, !stageContext); err != nil {
		return err
	}
	if err := timelineText(r.Summary, "summary", 4000, true); err != nil {
		return err
	}
	if err := timelineIDs(r.TicketIDs, "ticket_ids", false); err != nil {
		return err
	}
	if r.CandidateID != "" {
		if err := timelineID(r.CandidateID, "candidate_id"); err != nil {
			return err
		}
	}
	if r.RelatedEventID != "" {
		if err := timelineID(r.RelatedEventID, "related_event_id"); err != nil {
			return err
		}
	}
	if r.OccurredAt != "" {
		if err := timelineTime(r.OccurredAt, "occurred_at"); err != nil {
			return err
		}
	}
	if r.Kind == "stage" {
		if !contains(timelineStages, string(r.Stage)) {
			return timelineError("stage", "unsupported_stage")
		}
		if r.StageState != "in_progress" {
			return timelineError("stage_state", "stage_start_required")
		}
		if r.ExpectedSessionVersion == nil {
			return timelineError("expected_session_version", "required_for_stage")
		}
		if err := timelineNumber(*r.ExpectedSessionVersion, "expected_session_version"); err != nil {
			return err
		}
	} else if r.Stage != "" || r.StageState != "" || r.ExpectedSessionVersion != nil {
		return timelineError("stage", "forbidden_for_kind")
	}
	evidence := r.Kind == "validation" || r.Kind == "iteration"
	decision := r.Kind == "changes_reported" || r.Kind == "approval_reported"
	if evidence {
		if len(r.Views) == 0 || len(r.Views) > TimelineMaxViews {
			return timelineError("views", "nonempty_array_required_or_limit")
		}
		seen := map[string]bool{}
		hasPrior := false
		for _, v := range r.Views {
			if err := timelineView(v); err != nil {
				return err
			}
			if v.ViewID != "" {
				if seen[v.ViewID] {
					return timelineError("views.view_id", "duplicate_id")
				}
				seen[v.ViewID] = true
			}
			hasPrior = hasPrior || v.Supersedes != nil
		}
		if r.Kind == "iteration" && !hasPrior {
			return timelineError("views.supersedes", "iteration_reference_required")
		}
	} else if r.Views != nil {
		return timelineError("views", "forbidden_for_kind")
	}
	if decision {
		if r.Target == nil {
			return timelineError("target", "required_for_decision")
		}
		if err := timelineTarget(*r.Target); err != nil {
			return err
		}
		if err := timelineText(r.SourceReference, "source_reference", 1000, true); err != nil {
			return err
		}
	} else if r.Target != nil || r.SourceReference != "" {
		return timelineError("target", "forbidden_for_kind")
	}
	if r.Kind == "delivery" {
		if e := timelineIDs(r.DeliveryIDs, "delivery_ids", true); e != nil {
			return e
		}
	} else if r.DeliveryIDs != nil {
		return timelineError("delivery_ids", "forbidden_for_kind")
	}
	if r.Kind == "closure" {
		if r.Closure == nil || !contains([]string{"completed", "paused", "abandoned"}, r.Closure.Reason) {
			return timelineError("closure", "invalid_reason")
		}
		if err := timelineText(r.Closure.Result, "closure.result", 4000, true); err != nil {
			return err
		}
	} else if r.Closure != nil {
		return timelineError("closure", "forbidden_for_kind")
	}
	return nil
}

// ValidateTimelineEventReferences validates membership separately from payload
// decoding. It requires the current session snapshot under the eventual store lock.
func ValidateTimelineEventReferences(session TimelineSession, r TimelineAppendEventRequest) error {
	if _, err := ResolveTimelineStageEventID(session, r); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if err := timelineAppendLifecycle(session, r); err != nil {
		return err
	}
	return timelineEventReferences(session, r)
}
func timelineEventReferences(session TimelineSession, r TimelineAppendEventRequest) error {
	refs := func(ids []string) error { _, err := CaptureTimelineTicketReferences(session, ids); return err }
	if r.TicketIDs != nil {
		if err := refs(r.TicketIDs); err != nil {
			return err
		}
	}
	if r.CandidateID != "" {
		if _, err := TimelineCandidateByID(session, r.CandidateID); err != nil {
			return err
		}
	}
	if r.DeliveryIDs != nil {
		if _, err := CaptureTimelineDeliveryReferences(session, r.DeliveryIDs); err != nil {
			return err
		}
	}
	events := map[string]TimelineEvent{}
	views := map[string]bool{}
	artifacts := map[string]bool{}
	for _, e := range session.Events {
		events[e.EventID] = e
		for _, v := range e.Views {
			views[v.ViewID] = true
		}
	}
	for _, a := range session.Artifacts {
		if a.SessionID == session.ViewerSessionID {
			artifacts[a.ArtifactID] = true
		}
	}
	target := func(t TimelineEvidenceTarget) error {
		e, ok := events[t.EventID]
		if t.Kind == "document" {
			if !ok || e.Kind != "stage" || e.Withdrawn {
				return timelineError("target", "unknown_document_revision")
			}
			for _, snapshot := range e.DocumentSnapshots {
				if snapshot.DocumentRevisionID == t.DocumentRevisionID {
					return nil
				}
			}
			return timelineError("target", "unknown_document_revision")
		}
		if !ok || e.EvidenceID == "" || e.EvidenceID != t.EvidenceID || e.Withdrawn {
			return timelineError("target", "unknown_evidence")
		}
		owned := map[string]bool{}
		for _, v := range e.Views {
			owned[v.ViewID] = true
		}
		for _, id := range t.ViewIDs {
			if !owned[id] {
				return timelineError("target.view_ids", "unknown_view")
			}
		}
		return nil
	}
	if r.RelatedEventID != "" {
		if _, ok := events[r.RelatedEventID]; !ok {
			return timelineError("related_event_id", "unknown_event")
		}
	}
	if r.Target != nil {
		if err := target(*r.Target); err != nil {
			return err
		}
	}
	for _, v := range r.Views {
		if err := refs(v.TicketIDs); err != nil {
			return err
		}
		if v.ViewID != "" && !views[v.ViewID] {
			return timelineError("views.view_id", "unknown_view")
		}
		for _, image := range v.ImageRefs {
			if !artifacts[image.ArtifactID] {
				return timelineError("views.image_refs", "unknown_artifact")
			}
		}
		if v.Supersedes != nil {
			if err := target(TimelineEvidenceTarget{Kind: "evidence", EventID: v.Supersedes.EventID, EvidenceID: v.Supersedes.EvidenceID, ViewIDs: []string{v.Supersedes.ViewID}}); err != nil {
				return err
			}
		}
	}
	return nil
}

func DecodeTimelineSession(raw []byte) (TimelineSession, error) {
	var s TimelineSession
	err := DecodeTimelineJSON(raw, &s)
	if err == nil {
		err = ValidateTimelineSession(s)
	}
	return s, err
}

// ValidateTimelineSession validates public output and restored records, without
// side effects. Persistence and authoritative project resolution belong to the store.
func ValidateTimelineSession(s TimelineSession) error {
	if err := ValidateTimelineV3Graph(s); err != nil {
		return err
	}
	if s.SchemaVersion != TimelineSchemaVersion {
		return timelineError("schema_version", "unsupported_protocol")
	}
	if !timelineSHA256.MatchString(s.ViewerSessionID) {
		return timelineError("viewer_session_id", "invalid_id")
	}
	if err := timelineSource(s.Source); err != nil {
		return err
	}
	if err := timelineText(s.ProjectKey, "project_key", 2048, true); err != nil {
		return err
	}
	if err := timelineText(s.ProjectName, "project_name", 200, true); err != nil {
		return err
	}
	if err := timelineNumber(s.Version, "version"); err != nil {
		return err
	}
	if err := timelineText(s.Title, "title", 200, true); err != nil {
		return err
	}
	if err := timelineText(s.Summary, "summary", 4000, true); err != nil {
		return err
	}
	if !contains(timelineStates, string(s.State)) {
		return timelineError("state", "unsupported_state")
	}
	if err := timelineTime(s.CreatedAt, "created_at"); err != nil {
		return err
	}
	if err := timelineTime(s.UpdatedAt, "updated_at"); err != nil {
		return err
	}
	if err := validateTimelineCatalog(s); err != nil {
		return err
	}
	if err := timelineTickets(s.Tickets); err != nil {
		return err
	}
	if s.Events == nil || len(s.Events) > TimelineMaxEvents || s.Artifacts == nil {
		return timelineError("session", "invalid_collection")
	}
	artifacts := map[string]bool{}
	for _, a := range s.Artifacts {
		if err := timelineID(a.ArtifactID, "artifacts.artifact_id"); err != nil {
			return err
		}
		if a.SessionID != s.ViewerSessionID || artifacts[a.ArtifactID] {
			return timelineError("artifacts", "wrong_owner_or_duplicate")
		}
		artifacts[a.ArtifactID] = true
		if !timelineSHA256.MatchString(a.SHA256) || a.Size < 1 || a.Size > maxCapture {
			return timelineError("artifacts", "invalid_hash_or_size")
		}
		if err := timelineTime(a.ReceivedAt, "artifacts.received_at"); err != nil {
			return err
		}
		m := TimelineArtifactMetadata{Actor: a.Actor, SchemaVersion: 3, OperationID: "restore", CapturedAt: a.CapturedAt, CandidateID: a.CandidateID, Viewport: a.Viewport, Theme: a.Theme, MediaType: a.MediaType}
		if err := m.Validate(); err != nil {
			return err
		}
	}
	prefix := s
	prefix.Events = []TimelineEvent{}
	ids := map[string]bool{}
	evidenceIDs := map[string]bool{}
	sequence := int64(0)
	count := 0
	for _, e := range s.Events {
		if err := timelineID(e.EventID, "events.event_id"); err != nil {
			return err
		}
		if ids[e.EventID] {
			return timelineError("events.event_id", "duplicate_id")
		}
		ids[e.EventID] = true
		if err := timelineNumber(e.Sequence, "events.sequence"); err != nil {
			return err
		}
		if e.Sequence <= sequence {
			return timelineError("events.sequence", "not_increasing")
		}
		sequence = e.Sequence
		if err := timelineNumber(e.Version, "events.version"); err != nil {
			return err
		}
		if e.Origin != "agent_reported" && e.Origin != "server" {
			return timelineError("events.origin", "invalid_origin")
		}
		if err := timelineTime(e.CreatedAt, "events.created_at"); err != nil {
			return err
		}
		if err := timelineTime(e.UpdatedAt, "events.updated_at"); err != nil {
			return err
		}
		if err := timelineRefs(e.TicketRefs, "ticket_refs"); err != nil {
			return err
		}
		for _, ref := range e.DeliveryRefs {
			if err := timelineID(ref.DeliveryID, "delivery_id"); err != nil {
				return err
			}
			if err := timelineDeliverySnapshot(ref.Snapshot); err != nil {
				return err
			}
		}
		for _, view := range e.Views {
			if err := timelineRefs(view.TicketRefs, "views.ticket_refs"); err != nil {
				return err
			}
		}
		r := TimelineAppendEventRequest{Actor: e.Actor, Reason: e.Reason, TargetStageEventID: e.StageEventID, TaskID: e.TaskID, SegmentID: e.SegmentID, ReviewStatus: e.ReviewStatus, SchemaVersion: 3, OperationID: "restore", Kind: e.Kind, Title: e.Title, Summary: e.Summary, TicketIDs: timelineTicketIDs(e.TicketRefs), CandidateID: e.CandidateID, RelatedEventID: e.RelatedEventID, Target: e.Target, SourceReference: e.SourceReference, DeliveryIDs: timelineDeliveryIDs(e.DeliveryRefs), Closure: e.Closure}
		if e.Kind == "validation" || e.Kind == "iteration" {
			if err := timelineID(e.EvidenceID, "events.evidence_id"); err != nil {
				return err
			}
			if evidenceIDs[e.EvidenceID] {
				return timelineError("events.evidence_id", "duplicate_id")
			}
			evidenceIDs[e.EvidenceID] = true
			if !e.Withdrawn {
				count++
			}
		} else if e.EvidenceID != "" {
			return timelineError("events.evidence_id", "forbidden_for_kind")
		}
		viewIDs := map[string]bool{}
		if e.Views != nil {
			r.Views = []TimelineViewInput{}
		}
		for _, v := range e.Views {
			if err := timelineID(v.ViewID, "events.views.view_id"); err != nil {
				return err
			}
			if viewIDs[v.ViewID] {
				return timelineError("events.views.view_id", "duplicate_id")
			}
			viewIDs[v.ViewID] = true
			r.Views = append(r.Views, TimelineViewInput{ViewID: v.ViewID, Title: v.Title, TicketIDs: timelineTicketIDs(v.TicketRefs), Steps: v.Steps, ExpectedResult: v.ExpectedResult, ImageRefs: v.ImageRefs, Supersedes: v.Supersedes})
		}
		topTickets := map[string]bool{}
		for _, ref := range e.TicketRefs {
			topTickets[TimelineTicketIdentity(ref)] = true
		}
		for _, v := range e.Views {
			for _, ref := range v.TicketRefs {
				if !topTickets[TimelineTicketIdentity(ref)] {
					return timelineError("events.ticket_refs", "missing_view_ticket")
				}
			}
		}
		// A newly allocated view is valid in its creating output, while references
		// to prior evidence remain checked against the prefix, never future events.
		context := prefix
		context.Events = append([]TimelineEvent{}, prefix.Events...)
		for i := range context.Events {
			context.Events[i].Withdrawn = false
		}
		context.Events = append(context.Events, TimelineEvent{Views: e.Views})
		if e.Kind == "stage" {
			if err := validateTimelineStageEvent(e); err != nil {
				return err
			}
			r.Kind = "request"
		}
		if e.Audit == nil {
			if err := r.validate(e.Kind == "stage"); err != nil {
				return err
			}
		} else if err := validateTimelineAuditEvent(e); err != nil {
			return err
		}
		if err := timelineEventReferences(context, r); err != nil {
			return err
		}
		prefix.Events = append(prefix.Events, e)
	}
	if err := validateTimelineR3Session(s); err != nil {
		return err
	}
	if s.ValidationCount != count {
		return timelineError("validation_count", "count_mismatch")
	}
	return nil
}

// TimelineContractCode extracts a bounded code for callers' HTTP error mapping.
func TimelineContractCode(err error) string {
	var target *TimelineContractError
	if errors.As(err, &target) {
		return target.Code
	}
	return "invalid_request"
}

// IsSafeTimelineArtifactURL accepts only same-server capability URLs for this
// session. Filesystem ownership is still enforced by the serving boundary.
func IsSafeTimelineArtifactURL(raw, sessionID string) bool {
	if !timelineSHA256.MatchString(sessionID) || !strings.HasPrefix(raw, "/files/"+sessionID+"/") {
		return false
	}
	parts := strings.Split(strings.TrimPrefix(raw, "/files/"+sessionID+"/"), "/")
	if len(parts) < 2 || !timelineSHA256.MatchString(parts[0]) {
		return false
	}
	for _, part := range parts[1:] {
		if part == "." || part == ".." || !segmentPattern.MatchString(part) {
			return false
		}
	}
	return true
}
func DecodeTimelineSessionResponse(raw []byte) (TimelineSessionResponse, error) {
	var r TimelineSessionResponse
	if err := DecodeTimelineJSON(raw, &r); err != nil {
		return r, err
	}
	if r.SchemaVersion != 3 || r.Mode != "timeline_api" {
		return r, timelineError("mode", "unsupported_version_or_mode")
	}
	if err := ValidateTimelineSession(r.Session); err != nil {
		return r, err
	}
	artifacts := map[string]bool{}
	for _, a := range r.Session.Artifacts {
		artifacts[a.ArtifactID] = true
	}
	for key, url := range r.ArtifactURLs {
		if !artifacts[key] || !IsSafeTimelineArtifactURL(url, r.Session.ViewerSessionID) {
			return r, timelineError("artifact_urls", "invalid_artifact_url")
		}
	}
	documents := map[string]bool{}
	for _, document := range r.Session.Documents {
		documents[document.DocumentID] = true
	}
	for id, url := range r.DocumentURLs {
		if !documents[id] || !IsSafeTimelineDocumentURL(url, r.Session.ViewerSessionID, id) {
			return r, timelineError("document_urls", "invalid_document_url")
		}
	}
	return r, nil
}

func DecodeTimelineIndexEntry(raw []byte) (TimelineIndexEntry, error) {
	var discriminator map[string]json.RawMessage
	if err := DecodeTimelineJSON(raw, &discriminator); err != nil {
		return nil, err
	}
	var mode string
	if err := json.Unmarshal(discriminator["mode"], &mode); err != nil {
		return nil, timelineError("mode", "required_field")
	}
	switch mode {
	case "legacy":
		return nil, timelineError("mode", "unsupported_protocol")
	case "timeline_api":
		var entry TimelineAPIIndexEntry
		if err := DecodeTimelineJSON(raw, &entry); err != nil {
			return nil, err
		}
		rawSummary, err := json.Marshal(entry)
		if err != nil {
			return nil, timelineError("body", "invalid_summary")
		}
		var obj map[string]json.RawMessage
		if err = json.Unmarshal(rawSummary, &obj); err != nil {
			return nil, timelineError("body", "invalid_summary")
		}
		delete(obj, "mode")
		// The summary carries the stage pointer but not its event, so the
		// synthetic session below cannot prove the foreground. Check the
		// pointer on the entry and validate the rest as an active session.
		if entry.State == TimelineInReview || entry.State == TimelineAwaitingValidation {
			if entry.CurrentStageEventID == nil {
				return nil, timelineError("current_stage_event_id", "foreground_required")
			}
			obj["state"] = json.RawMessage(`"active"`)
		}
		obj["schema_version"] = json.RawMessage(`3`)
		obj["producers"], _ = json.Marshal([]TimelineProducer{{ProducerID: "summary", Source: entry.Source, JoinedAt: entry.CreatedAt}})
		obj["foreground_producer_id"] = json.RawMessage(`"summary"`)
		obj["pending_join_requests"] = json.RawMessage(`[]`)
		obj["tasks"] = json.RawMessage(`[]`)
		obj["waits"] = json.RawMessage(`[]`)
		obj["deliveries"] = json.RawMessage(`[]`)
		obj["candidates"] = json.RawMessage(`[]`)
		obj["events"] = json.RawMessage(`[]`)
		obj["artifacts"] = json.RawMessage(`[]`)
		obj["documents"] = json.RawMessage(`[]`)
		obj["current_stage_event_id"] = json.RawMessage(`null`)
		obj["validation_count"] = json.RawMessage(`0`)
		rawSummary, err = json.Marshal(obj)
		if err != nil {
			return nil, timelineError("body", "invalid_summary")
		}
		if _, err = DecodeTimelineSession(rawSummary); err != nil {
			return nil, err
		}
		if entry.ValidationCount < 0 || entry.ValidationCount > TimelineMaxEvents {
			return nil, timelineError("validation_count", "invalid_count")
		}
		if entry.CurrentStageEventID != nil {
			if err := timelineID(*entry.CurrentStageEventID, "current_stage_event_id"); err != nil {
				return nil, err
			}
		}
		return entry, nil
	default:
		return nil, timelineError("mode", "unsupported_mode")
	}
}

func (r TimelineOperationResult) Validate() error {
	if r.TicketID != "" && r.Kind != "ticket_create" && r.Kind != "ticket_patch" {
		return timelineError("ticket_id", "forbidden_for_kind")
	}
	if r.DeliveryID != "" && r.Kind != "delivery_create" && r.Kind != "delivery_patch" {
		return timelineError("delivery_id", "forbidden_for_kind")
	}
	if r.CandidateID != "" && r.Kind != "candidate_create" {
		return timelineError("candidate_id", "forbidden_for_kind")
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if !timelineSHA256.MatchString(r.ViewerSessionID) {
		return timelineError("viewer_session_id", "invalid_id")
	}
	if err := timelineNumber(r.SessionVersion, "session_version"); err != nil {
		return err
	}
	event := r.Kind == "append_event" || r.Kind == "event_patch" || r.Kind == "event_withdraw" || r.Kind == "document_snapshot"
	switch r.Kind {
	case "ticket_create", "delivery_create", "candidate_create", "ticket_patch", "delivery_patch":
		if err := validateTimelineCatalogResult(r); err != nil {
			return err
		}
	case "create_session":
		if r.Created == nil {
			return timelineError("created", "required_field")
		}
	case "session_patch":
	case "join_request", "producer_accept", "foreground_transfer", "recover_lead", "task_create", "task_patch", "task_transition", "wait_create", "wait_attempt_start", "wait_observe":
		if err := validateTimelineV3Result(r); err != nil {
			return err
		}
	case "session_delete":
		if r.Deleted == nil || !*r.Deleted {
			return timelineError("deleted", "completed_delete_required")
		}
	case "append_event", "event_patch", "event_withdraw", "document_snapshot":
		if err := timelineID(r.EventID, "event_id"); err != nil {
			return err
		}
		if err := timelineNumber(r.EventVersion, "event_version"); err != nil {
			return err
		}
		if err := timelineNumber(r.Sequence, "sequence"); err != nil {
			return err
		}
		if r.Kind == "event_withdraw" && (r.Withdrawn == nil || !*r.Withdrawn) {
			return timelineError("withdrawn", "completed_withdrawal_required")
		}
	case "document_add":
		if err := timelineID(r.DocumentID, "document_id"); err != nil {
			return err
		}
		if !timelineSHA256.MatchString(r.SHA256) || r.Size == nil || *r.Size < 0 || *r.Size > TimelineMaxDocumentBytes || r.Created == nil {
			return timelineError("document", "invalid_result")
		}
	case "artifact_add":
		if err := timelineID(r.ArtifactID, "artifact_id"); err != nil {
			return err
		}
		if r.Created == nil {
			return timelineError("created", "required_field")
		}
	default:
		return timelineError("kind", "unsupported_operation")
	}
	if !event && (r.EventID != "" || r.EventVersion != 0 || r.Sequence != 0 || r.EvidenceID != "" || r.ViewIDs != nil || r.Withdrawn != nil) {
		return timelineError("event_id", "forbidden_for_kind")
	}
	if r.Kind != "artifact_add" && r.ArtifactID != "" {
		return timelineError("artifact_id", "forbidden_for_kind")
	}
	if r.Kind != "session_delete" && r.Deleted != nil {
		return timelineError("deleted", "forbidden_for_kind")
	}
	if r.Kind != "create_session" && r.Kind != "artifact_add" && r.Kind != "document_add" && r.Kind != "ticket_create" && r.Kind != "delivery_create" && r.Kind != "candidate_create" && r.Created != nil {
		return timelineError("created", "forbidden_for_kind")
	}
	if r.Kind != "event_withdraw" && r.Withdrawn != nil {
		return timelineError("withdrawn", "forbidden_for_kind")
	}
	if r.EvidenceID != "" {
		if err := timelineID(r.EvidenceID, "evidence_id"); err != nil {
			return err
		}
	}
	if r.ViewIDs != nil {
		if r.EvidenceID == "" || len(r.ViewIDs) > TimelineMaxViews {
			return timelineError("view_ids", "invalid_evidence_scope")
		}
		seen := map[string]bool{}
		for _, id := range r.ViewIDs {
			if err := timelineID(id, "view_ids"); err != nil {
				return err
			}
			if seen[id] {
				return timelineError("view_ids", "duplicate_id")
			}
			seen[id] = true
		}
	}
	if r.CurrentStageEventID != nil {
		if err := timelineID(*r.CurrentStageEventID, "current_stage_event_id"); err != nil {
			return err
		}
	}
	if r.Kind == "document_snapshot" && (r.EvidenceID != "" || r.ViewIDs != nil) {
		return timelineError("evidence_id", "forbidden_for_kind")
	}
	if r.Kind == "document_snapshot" {
		if err := timelineID(r.DocumentRevisionID, "document_revision_id"); err != nil {
			return err
		}
	}
	if r.Kind != "document_add" && (r.DocumentID != "" || r.SHA256 != "" || r.Size != nil) {
		return timelineError("document_id", "forbidden_for_kind")
	}
	if r.Kind != "document_snapshot" && r.DocumentRevisionID != "" {
		return timelineError("document_revision_id", "forbidden_for_kind")
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return timelineError("body", "invalid_result")
	}
	return timelineResultFields(raw, r.Kind)
}
func DecodeTimelineOperationResult(raw []byte) (TimelineOperationResult, error) {
	var r TimelineOperationResult
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	if err == nil {
		err = timelineResultFields(raw, r.Kind)
	}
	return r, err
}
func (r TimelineOperationReceipt) Validate() error {
	if !timelineOperationID.MatchString(r.OperationID) {
		return timelineError("operation_id", "invalid_id")
	}
	if !timelineSHA256.MatchString(r.RequestSHA256) {
		return timelineError("request_sha256", "invalid_digest")
	}
	if err := timelineTime(r.CompletedAt, "completed_at"); err != nil {
		return err
	}
	if err := r.Result.Validate(); err != nil {
		return err
	}
	if r.Result.OperationID != r.OperationID {
		return timelineError("result.operation_id", "operation_mismatch")
	}
	return nil
}
func DecodeTimelineOperationReceipt(raw []byte) (TimelineOperationReceipt, error) {
	var r TimelineOperationReceipt
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	if err == nil {
		var fields map[string]json.RawMessage
		if e := json.Unmarshal(raw, &fields); e != nil {
			err = timelineError("body", "invalid_receipt")
		} else {
			err = timelineResultFields(fields["result"], r.Result.Kind)
		}
	}
	return r, err
}

// ValidateTimelineOperationRetry compares contract data only; receipt storage and
// retry dispatch remain the responsibility of the later server implementation.
func ValidateTimelineOperationRetry(receipt TimelineOperationReceipt, operationID, requestSHA256 string) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	if receipt.OperationID != operationID || receipt.RequestSHA256 != requestSHA256 {
		return timelineError("operation_id", "operation_conflict")
	}
	return nil
}

// Shared by strict result decoding and offline schema generation.
func timelineResultAllowed(kind TimelineOperationKind) map[string]bool {
	allowed := map[string]bool{"schema_version": true, "operation_id": true, "kind": true, "viewer_session_id": true, "session_version": true, "current_stage_event_id": true}
	switch kind {
	case "ticket_create":
		allowed["ticket_id"] = true
		allowed["created"] = true
	case "ticket_patch":
		allowed["ticket_id"] = true
	case "delivery_create":
		allowed["delivery_id"] = true
		allowed["created"] = true
	case "delivery_patch":
		allowed["delivery_id"] = true
	case "candidate_create":
		allowed["candidate_id"] = true
		allowed["created"] = true
	case "create_session":
		allowed["created"] = true
	case "session_patch":
	case "session_delete":
		allowed["deleted"] = true
	case "append_event", "event_patch", "event_withdraw", "document_snapshot":
		for _, name := range []string{"event_id", "event_version", "sequence", "evidence_id", "view_ids"} {
			allowed[name] = true
		}
		if kind == "event_withdraw" {
			allowed["withdrawn"] = true
		}
	case "document_add":
		for _, name := range []string{"document_id", "sha256", "size", "created"} {
			allowed[name] = true
		}
	case "artifact_add":
		allowed["artifact_id"] = true
		allowed["created"] = true
	}
	if kind == "document_snapshot" {
		allowed["document_revision_id"] = true
	}
	if kind == "document_snapshot" {
		delete(allowed, "evidence_id")
		delete(allowed, "view_ids")
	}
	for _, field := range timelineV3ResultFields[string(kind)] {
		allowed[field] = true
	}
	return allowed
}
func timelineResultFields(raw []byte, kind TimelineOperationKind) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return timelineError("body", "invalid_result")
	}
	allowed := timelineResultAllowed(kind)
	for name := range fields {
		if !allowed[name] {
			return timelineError("body", "forbidden_result_field")
		}
	}
	return nil
}

// Document and stage validators are pure contract checks. Store transactions
// must apply their side effects and compare the same versions under its lock.
func ValidateTimelineDocumentBytes(content []byte) error {
	if len(content) > TimelineMaxDocumentBytes || !utf8.Valid(content) {
		return timelineError("document", "invalid_utf8_or_size")
	}
	return nil
}
func (r TimelineDocumentMetadata) Validate() error {
	if err := validateTimelineNativeActor(r.Actor); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if r.MediaType != "text/markdown" {
		return timelineError("media_type", "unsupported_media_type")
	}
	return nil
}
func DecodeTimelineDocumentMetadata(raw []byte) (TimelineDocumentMetadata, error) {
	var r TimelineDocumentMetadata
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func timelineDocument(d TimelineDocument) error {
	if err := validateTimelineNativeActor(d.Actor); err != nil {
		return err
	}
	if err := timelineID(d.DocumentID, "documents.document_id"); err != nil {
		return err
	}
	if !timelineSHA256.MatchString(d.SessionID) || !timelineSHA256.MatchString(d.SHA256) || d.Size < 0 || d.Size > TimelineMaxDocumentBytes {
		return timelineError("documents", "invalid_metadata")
	}
	if d.MediaType != "text/markdown" {
		return timelineError("documents.media_type", "unsupported_media_type")
	}
	return timelineTime(d.ReceivedAt, "documents.received_at")
}
func timelineDocumentTabs(tabs []TimelineDocumentTab) error {
	if len(tabs) < 1 || len(tabs) > TimelineMaxDocumentTabs {
		return timelineError("tabs", "invalid_count")
	}
	labels := map[string]bool{}
	for _, tab := range tabs {
		if err := timelineText(tab.Label, "tabs.label", 100, true); err != nil {
			return err
		}
		if labels[tab.Label] {
			return timelineError("tabs.label", "duplicate_label")
		}
		labels[tab.Label] = true
		if err := timelineID(tab.DocumentID, "tabs.document_id"); err != nil {
			return err
		}
	}
	return nil
}
func (r TimelineDocumentSnapshotRequest) Validate() error {
	if err := validateTimelineSnapshotV3(r); err != nil {
		return err
	}
	if err := timelineOperation(r.SchemaVersion, r.OperationID); err != nil {
		return err
	}
	if err := timelineNumber(r.ExpectedVersion, "expected_version"); err != nil {
		return err
	}
	if !contains([]string{"draft", "complete"}, r.Status) {
		return timelineError("status", "invalid_snapshot_status")
	}
	return timelineDocumentTabs(r.Tabs)
}
func DecodeTimelineDocumentSnapshot(raw []byte) (TimelineDocumentSnapshotRequest, error) {
	var r TimelineDocumentSnapshotRequest
	err := DecodeTimelineJSON(raw, &r)
	if err == nil {
		err = r.Validate()
	}
	return r, err
}
func DecodeTimelineDocumentResponse(raw []byte) (TimelineDocumentResponse, error) {
	var r TimelineDocumentResponse
	err := DecodeTimelineJSON(raw, &r)
	if err != nil {
		return r, err
	}
	if r.SchemaVersion != 3 {
		return r, timelineError("schema_version", "unsupported_protocol")
	}
	if err = timelineDocument(r.Document); err != nil {
		return r, err
	}
	if err = ValidateTimelineDocumentBytes([]byte(r.Content)); err != nil {
		return r, err
	}
	digest := sha256.Sum256([]byte(r.Content))
	if r.Document.Size != int64(len([]byte(r.Content))) || r.Document.SHA256 != hex.EncodeToString(digest[:]) {
		return r, timelineError("document", "content_mismatch")
	}
	return r, nil
}
func IsSafeTimelineDocumentURL(raw, sessionID, documentID string) bool {
	return timelineSHA256.MatchString(sessionID) && timelineOpaqueID.MatchString(documentID) && raw == "/api/timeline/sessions/"+sessionID+"/documents/"+documentID
}
func validateTimelineStageEvent(e TimelineEvent) error {
	if !contains(timelineStages, string(e.Stage)) || !contains(timelineStageStates, string(e.StageState)) {
		return timelineError("stage", "invalid_stage_state")
	}
	if err := timelineTime(e.StartedAt, "started_at"); err != nil {
		return err
	}
	started, _ := time.Parse(time.RFC3339Nano, e.StartedAt)
	created, _ := time.Parse(time.RFC3339Nano, e.CreatedAt)
	updated, _ := time.Parse(time.RFC3339Nano, e.UpdatedAt)
	if !started.Equal(created) || updated.Before(started) {
		return timelineError("started_at", "server_time_mismatch")
	}
	terminal := e.StageState == "completed" || e.StageState == "failed"
	if terminal {
		if err := timelineTime(e.FinishedAt, "finished_at"); err != nil {
			return err
		}
		finished, _ := time.Parse(time.RFC3339Nano, e.FinishedAt)
		if finished.Before(started) || finished.After(updated) {
			return timelineError("finished_at", "server_time_mismatch")
		}
	} else if e.FinishedAt != "" {
		return timelineError("finished_at", "forbidden_for_stage_state")
	}
	for _, snapshot := range e.DocumentSnapshots {
		if err := timelineID(snapshot.DocumentRevisionID, "document_revision_id"); err != nil {
			return err
		}
		if !contains([]string{"draft", "complete"}, snapshot.Status) {
			return timelineError("status", "invalid_snapshot_status")
		}
		if err := timelineTime(snapshot.CreatedAt, "document_snapshots.created_at"); err != nil {
			return err
		}
		at, _ := time.Parse(time.RFC3339Nano, snapshot.CreatedAt)
		if at.Before(started) || at.After(updated) {
			return timelineError("document_snapshots.created_at", "server_time_mismatch")
		}
		if err := timelineDocumentTabs(snapshot.Tabs); err != nil {
			return err
		}
	}
	return nil
}
func validateTimelineR3Session(s TimelineSession) error {
	if s.Documents == nil || len(s.Documents) > TimelineMaxDocuments {
		return timelineError("documents", "array_required_or_limit")
	}
	docs := map[string]bool{}
	hashes := map[string]bool{}
	total := int64(0)
	for _, d := range s.Documents {
		if err := timelineDocument(d); err != nil {
			return err
		}
		if d.SessionID != s.ViewerSessionID || docs[d.DocumentID] {
			return timelineError("documents", "wrong_owner_or_duplicate")
		}
		if hashes[d.SHA256] {
			return timelineError("documents.sha256", "duplicate_content")
		}
		hashes[d.SHA256] = true
		docs[d.DocumentID] = true
		total += d.Size
	}
	if total > TimelineMaxAggregateDocumentBytes {
		return timelineError("documents", "aggregate_limit")
	}
	snapshots := map[string]bool{}
	latestStages := map[TimelineStage]TimelineEvent{}
	active := ""
	for _, e := range s.Events {
		if e.OccurredAt != "" {
			if err := timelineTime(e.OccurredAt, "occurred_at"); err != nil {
				return err
			}
		}
		if e.Kind != "stage" {
			if e.Stage != "" || e.StageState != "" || e.StartedAt != "" || e.FinishedAt != "" || e.DocumentSnapshots != nil {
				return timelineError("stage", "forbidden_for_kind")
			}
			continue
		}
		if err := validateTimelineStageEvent(e); err != nil {
			return err
		}
		if !e.Withdrawn {
			if previous, ok := latestStages[e.Stage]; ok {
				if (previous.StageState != "completed" && previous.StageState != "failed") || e.RelatedEventID != previous.EventID {
					return timelineError("related_event_id", "stage_reentry_reference_required")
				}
			}
			latestStages[e.Stage] = e
		}
		if contains([]string{"in_progress", "in_review", "awaiting_validation"}, string(e.StageState)) {
			if active != "" || e.Withdrawn {
				return timelineError("current_stage_event_id", "multiple_or_withdrawn_foreground")
			}
			active = e.EventID
		}
		for _, snapshot := range e.DocumentSnapshots {
			if snapshots[snapshot.DocumentRevisionID] {
				return timelineError("document_revision_id", "duplicate_id")
			}
			snapshots[snapshot.DocumentRevisionID] = true
			for _, tab := range snapshot.Tabs {
				if !docs[tab.DocumentID] {
					return timelineError("tabs.document_id", "unknown_document")
				}
			}
		}
	}
	if len(snapshots) > TimelineMaxDocumentSnapshots {
		return timelineError("document_snapshots", "session_limit")
	}
	if s.CurrentStageEventID == nil {
		if active != "" {
			return timelineError("current_stage_event_id", "foreground_mismatch")
		}
	} else {
		if err := timelineID(*s.CurrentStageEventID, "current_stage_event_id"); err != nil {
			return err
		}
		if active != *s.CurrentStageEventID {
			return timelineError("current_stage_event_id", "foreground_mismatch")
		}
	}
	if active != "" && s.State != TimelineActive && s.State != "in_review" && s.State != TimelineAwaitingValidation {
		return timelineError("state", "foreground_requires_active_session")
	}
	return nil
}
func timelineAppendLifecycle(s TimelineSession, r TimelineAppendEventRequest) error {
	if s.State == TimelineClosed || s.State == "failed" {
		return timelineError("state", "session_closed")
	}
	if r.Kind == "closure" && s.CurrentStageEventID != nil {
		return timelineError("current_stage_event_id", "stage_active")
	}
	if r.Kind != "stage" {
		return nil
	}
	if r.ExpectedSessionVersion == nil || *r.ExpectedSessionVersion != s.Version {
		return timelineError("expected_session_version", "version_conflict")
	}
	var latest *TimelineEvent
	for i := range s.Events {
		e := &s.Events[i]
		if e.Kind == "stage" && e.Stage == r.Stage && !e.Withdrawn {
			latest = e
		}
	}
	if latest != nil {
		if latest.StageState != "completed" && latest.StageState != "failed" {
			return timelineError("stage", "stage_resume_required")
		}
		if r.RelatedEventID != latest.EventID {
			return timelineError("related_event_id", "stage_reentry_reference_required")
		}
	}
	return nil
}
func ValidateTimelineEventPatch(s TimelineSession, eventID string, r TimelineEventPatchRequest) error {
	if err := r.Validate(); err != nil {
		return err
	}
	var event *TimelineEvent
	for i := range s.Events {
		if s.Events[i].EventID == eventID {
			event = &s.Events[i]
			break
		}
	}
	if event == nil {
		return timelineError("event_id", "unknown_event")
	}
	if r.Title != nil {
		if err := timelineText(*r.Title, "title", 200, event.Kind != "stage"); err != nil {
			return err
		}
	}
	if event.Audit != nil {
		return timelineError("event_id", "immutable_audit")
	}
	operation := "event_patch"
	if event.Kind == "stage" {
		operation = "stage_update"
	}
	if err := ValidateTimelineAuthority(s, operation, r.Actor, event.Actor.ProducerID); err != nil {
		return err
	}
	if event.Version != r.ExpectedVersion {
		return timelineError("expected_version", "version_conflict")
	}
	if r.StageState == nil {
		return nil
	}
	if event.Kind != "stage" {
		return timelineError("stage_state", "forbidden_for_kind")
	}
	if r.ExpectedSessionVersion == nil || *r.ExpectedSessionVersion != s.Version {
		return timelineError("expected_session_version", "version_conflict")
	}
	if s.State == TimelineClosed || event.Withdrawn {
		return timelineError("stage_state", "stage_unavailable")
	}
	if event.StageState == *r.StageState {
		return nil
	}
	if event.StageState == "completed" || event.StageState == "failed" {
		return timelineError("stage_state", "stage_terminal")
	}
	return nil
}
func ValidateTimelineSessionPatch(s TimelineSession, r TimelineSessionPatchRequest) error {
	operation := "session_patch"
	if r.Actor.Kind == "local_viewer" {
		operation = "session_rename"
		if r.Archived != nil {
			operation = "session_archive"
		}
		if r.Pinned != nil {
			operation = "session_pin"
		}
	}
	if err := ValidateTimelineAuthority(s, operation, r.Actor, ""); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if r.ExpectedVersion != s.Version {
		return timelineError("expected_version", "version_conflict")
	}
	if r.State != nil && *r.State == TimelineClosed && s.CurrentStageEventID != nil {
		return timelineError("current_stage_event_id", "stage_active")
	}
	return nil
}
func ValidateTimelineEventWithdrawal(s TimelineSession, eventID string, r TimelineMutationRequest) error {
	if err := r.Validate(); err != nil {
		return err
	}
	for _, e := range s.Events {
		if e.EventID != eventID {
			continue
		}
		if e.Audit != nil {
			return timelineError("event_id", "immutable_audit")
		}
		operation := "event_withdraw"
		if e.Kind == "stage" {
			operation = "stage_update"
		}
		if err := ValidateTimelineAuthority(s, operation, r.Actor, e.Actor.ProducerID); err != nil {
			return err
		}
		if e.Version != r.ExpectedVersion {
			return timelineError("expected_version", "version_conflict")
		}
		if e.Kind == "stage" && e.StageState == "in_progress" {
			return timelineError("current_stage_event_id", "stage_active")
		}
		return nil
	}
	return timelineError("event_id", "unknown_event")
}
func ValidateTimelineDocumentUpload(s TimelineSession, r TimelineDocumentMetadata, content []byte) error {
	if err := ValidateTimelineAuthority(s, "document_add", r.Actor, ""); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if err := ValidateTimelineDocumentBytes(content); err != nil {
		return err
	}
	if s.State == TimelineClosed || s.State == "failed" {
		return timelineError("state", "session_closed")
	}
	digest := sha256.Sum256(content)
	hash := hex.EncodeToString(digest[:])
	total := int64(len(content))
	for _, d := range s.Documents {
		if d.SHA256 == hash {
			return nil
		}
		total += d.Size
	}
	if len(s.Documents) >= TimelineMaxDocuments || total > TimelineMaxAggregateDocumentBytes {
		return timelineError("documents", "capacity")
	}
	return nil
}
func ValidateTimelineDocumentSnapshot(s TimelineSession, eventID string, r TimelineDocumentSnapshotRequest) error {
	if err := validateTimelineSnapshotReferencesV3(s, eventID, r); err != nil {
		return err
	}
	if err := r.Validate(); err != nil {
		return err
	}
	if s.State == TimelineClosed || s.State == "failed" {
		return timelineError("state", "session_closed")
	}
	var event *TimelineEvent
	count := 0
	for i := range s.Events {
		count += len(s.Events[i].DocumentSnapshots)
		if s.Events[i].EventID == eventID {
			event = &s.Events[i]
		}
	}
	if event == nil || event.Kind != "stage" || event.Withdrawn {
		return timelineError("event_id", "stage_required")
	}
	if event.Version != r.ExpectedVersion {
		return timelineError("expected_version", "version_conflict")
	}
	if count >= TimelineMaxDocumentSnapshots {
		return timelineError("document_snapshots", "session_limit")
	}
	docs := map[string]bool{}
	for _, d := range s.Documents {
		if d.SessionID == s.ViewerSessionID {
			docs[d.DocumentID] = true
		}
	}
	for _, tab := range r.Tabs {
		if !docs[tab.DocumentID] {
			return timelineError("tabs.document_id", "unknown_document")
		}
	}
	return nil
}
