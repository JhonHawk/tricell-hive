package viewer

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
)

// TimelineLocalizationMetadata describes closed system labels, not producer text.
type TimelineLocalizationMetadata struct {
	DefaultLanguage    string   `json:"default_language"`
	SupportedLanguages []string `json:"supported_languages"`
	ReservedTitleKeys  []string `json:"reserved_title_keys"`
}

// TimelineProtocolDescription is offline metadata; it never starts or contacts a viewer.
// Schemas describe exact structural fields from the production Go DTOs. The
// semantic invariants remain authoritative validator rules, not JSON Schema promises.
type TimelineProtocolDescription struct {
	Localization    TimelineLocalizationMetadata `json:"localization"`
	StorageVersion  int                          `json:"storage_version"`
	Authority       map[string]string            `json:"authority"`
	ErrorCodes      []string                     `json:"error_codes"`
	ProtocolVersion int                          `json:"protocol_version"`
	SchemaDialect   string                       `json:"schema_dialect"`
	Capabilities    []string                     `json:"capabilities"`
	Retention       string                       `json:"retention"`
	TTLSeconds      int                          `json:"ttl_seconds"`
	Commands        []TimelineProtocolOperation  `json:"commands"`
	Limits          map[string]int64             `json:"limits"`
	Invariants      []string                     `json:"invariants"`
}
type TimelineProtocolOperation struct {
	Method       string          `json:"method,omitempty"`
	Path         string          `json:"path,omitempty"`
	Operation    string          `json:"operation"`
	Usage        string          `json:"usage"`
	InputSchema  map[string]any  `json:"input_schema"`
	OutputSchema map[string]any  `json:"output_schema"`
	Example      json.RawMessage `json:"example"`
}
type timelineProtocolDefinition struct {
	name, usage   string
	input, output any
	example       any
}

// TimelineProtocolContract narrows to an exact operation, or returns all operations
// when operation is empty. Examples are typed values checked by production decoders.
func TimelineProtocolContract(operation string) (TimelineProtocolDescription, error) {
	actor := TimelineActor{Kind: "native_producer", ProducerID: "producer-example"}
	one := int64(1)
	source := TimelineSource{ProjectRoot: "/workspace/example", Host: "codex", NativeSessionID: "native-example"}
	create := TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "create-example", Source: source, Title: "Example", Summary: "Scope", Tickets: []TimelineTicketInput{}, WorkKey: timelineStringPointer("COM-123")}
	stage := TimelineAppendEventRequest{Actor: actor, SchemaVersion: 3, OperationID: "stage-example", Reason: "Start the plan", Kind: "stage", Title: "", Summary: "Prepare the plan", TicketIDs: []string{}, Stage: "plan", StageState: "in_progress", ExpectedSessionVersion: &one}
	state := TimelineStageState("completed")
	patch := TimelineEventPatchRequest{Actor: actor, SchemaVersion: 3, OperationID: "complete-example", Reason: "Plan completed", ExpectedVersion: 1, StageState: &state, ExpectedSessionVersion: &one}
	mutation := TimelineMutationRequest{Actor: actor, SchemaVersion: 3, OperationID: "delete-example", ExpectedVersion: 1}
	definitions := []timelineProtocolDefinition{
		{"create_session", "hive view session create --json <file|->", TimelineCreateSessionRequest{}, TimelineSessionResponse{}, create},
		{"session_get", "hive view session get <session-id>", nil, TimelineSessionResponse{}, map[string]any{}},
		{"session_patch", "hive view session update <session-id> --json <file|->", TimelineSessionPatchRequest{}, TimelineOperationResult{}, TimelineSessionPatchRequest{Actor: actor, SchemaVersion: 3, OperationID: "update-example", ExpectedVersion: 1, Summary: timelineStringPointer("Updated summary")}},
		{"session_delete", "Local viewer only: DELETE /api/timeline/sessions/<session-id>", TimelineMutationRequest{}, TimelineOperationResult{}, TimelineMutationRequest{SchemaVersion: 3, OperationID: "delete-example", ExpectedVersion: 1, Actor: TimelineActor{Kind: "local_viewer"}}},
		{"append_event", "hive view event add <session-id> --json <file|->", TimelineAppendEventRequest{}, TimelineOperationResult{}, TimelineAppendEventRequest{Actor: actor, SchemaVersion: 3, OperationID: "request-example", Kind: "request", Title: "Request", Summary: "Initial scope", TicketIDs: []string{}}},
		{"event_patch", "hive view event update <session-id> <event-id> --json <file|->", TimelineEventPatchRequest{}, TimelineOperationResult{}, TimelineEventPatchRequest{Actor: actor, SchemaVersion: 3, OperationID: "copy-example", ExpectedVersion: 1, Title: timelineStringPointer("Corrected title")}},
		{"event_withdraw", "hive view event withdraw <session-id> <event-id> --json <file|->", TimelineMutationRequest{}, TimelineOperationResult{}, mutation},
		{"artifact_add", "hive view artifact add <session-id> --json <file|-> --file <image>", TimelineArtifactMetadata{}, TimelineOperationResult{}, TimelineArtifactMetadata{Actor: actor, SchemaVersion: 3, OperationID: "image-example", CapturedAt: "2026-01-01T00:00:00Z", CandidateID: "candidate-example", Viewport: "desktop", Theme: "light", MediaType: "image/png"}},
		{"document_add", "hive view document add <session-id> --json <file|-> --file <markdown>", TimelineDocumentMetadata{}, TimelineOperationResult{}, TimelineDocumentMetadata{Actor: actor, SchemaVersion: 3, OperationID: "document-example", MediaType: "text/markdown"}},
		{"document_get", "hive view document get <session-id> <document-id>", nil, TimelineDocumentResponse{}, map[string]any{}},
		{"stage_start", "hive view stage start <session-id> --json <file|->", TimelineAppendEventRequest{}, TimelineOperationResult{}, stage},
		{"stage_update", "hive view stage update <session-id> <event-id> --json <file|->", TimelineEventPatchRequest{}, TimelineOperationResult{}, patch},
		{"document_snapshot", "hive view stage snapshot <session-id> <event-id> --json <file|->", TimelineDocumentSnapshotRequest{}, TimelineOperationResult{}, TimelineDocumentSnapshotRequest{Actor: actor, SchemaVersion: 3, OperationID: "snapshot-example", DocumentKey: "plan", ExpectedSessionVersion: 1, ExpectedVersion: 1, Status: "draft", Tabs: []TimelineDocumentTab{{Label: "Proposal", DocumentID: "document-example"}}}},
	}
	definitions = append(definitions, timelineV3ProtocolDefinitions(actor)...)
	definitions = append(definitions, timelineCatalogProtocolDefinitions(actor)...)
	result := TimelineProtocolDescription{Localization: TimelineLocalizationMetadata{DefaultLanguage: "en", SupportedLanguages: []string{"en", "es"}, ReservedTitleKeys: []string{"start", "research", "plan", "build", "verify", "delivery", "tasks", "ci_wait"}}, StorageVersion: 3, Authority: map[string]string{
		"create_session":  "native source bootstrap; server creates original producer and lead",
		"join_request":    "same-project native source; no membership until acceptance",
		"producer_accept": "lead; re-resolve pending Source against work.ProjectKey", "foreground_transfer": "lead to accepted producer", "recover_lead": "local_viewer; pending same-project request; nonterminal work",
		"session_patch":  "lead for summary/lifecycle; local_viewer for title/archive/pin only",
		"session_delete": "local_viewer", "stage_start": "lead", "stage_update": "lead", "closure": "lead",
		"append_event": "accepted producer; collaborator requires explicit stage target", "event_patch": "content owner; lead for stage lifecycle", "event_withdraw": "content owner or lead moderation; immutable audit forbidden",
		"artifact_add": "accepted producer", "document_add": "accepted producer", "document_snapshot": "accepted producer; explicit stage and chain predecessor",
		"ticket_create": "lead", "ticket_patch": "lead", "delivery_create": "lead", "delivery_patch": "lead", "candidate_create": "accepted native producer",
		"task_create": "lead", "task_patch": "lead", "task_transition": "lead or assigned segment producer", "wait_create": "lead or assigned delivery-segment producer; lead only when unbound", "wait_attempt_start": "lead or assigned delivery-segment producer; lead only when unbound", "wait_observe": "lead or assigned delivery-segment producer; lead only when unbound",
	}, ErrorCodes: []string{"scope_reference_required", "entity_exists", "entity_conflict", "duplicate_plan_order", "unknown_ticket", "unknown_delivery", "unknown_candidate", "duplicate_id", "pull_request_required", "project_identity_unavailable", "duplicate_identity", "segment_assignee_required", "invalid_audit_operation", "invalid_state_transition", "forbidden_for_operation", "server_event_required", "invalid_producer_transition", "unsupported_protocol", "cutoff_required", "version_conflict", "operation_conflict", "producer_not_accepted", "foreign_project", "lead_required", "forbidden_actor", "immutable_audit", "stage_binding_mismatch", "candidate_mismatch", "current_predecessor_required", "pending_request_required", "unknown_or_mixed_attempt", "superseded_attempt", "wait_aggregate_mismatch", "limit"}, ProtocolVersion: 3, SchemaDialect: "https://json-schema.org/draft/2020-12/schema", Retention: "manual", TTLSeconds: 0, Capabilities: []string{"canonical_registries", "timeline_api", "conditional_mutations", "immutable_evidence", "stage_lifecycle", "document_snapshots", "manual_retention"}, Commands: []TimelineProtocolOperation{}, Limits: map[string]int64{"body_bytes": TimelineMaxBody, "document_bytes": TimelineMaxDocumentBytes, "document_tabs": TimelineMaxDocumentTabs, "documents": TimelineMaxDocuments, "document_snapshots": TimelineMaxDocumentSnapshots, "aggregate_document_bytes": TimelineMaxAggregateDocumentBytes, "events": TimelineMaxEvents, "content_events": TimelineContentEvents, "producers": 16, "pending_join_requests": 16, "tasks": 100, "task_segments": 5, "required_checks": TimelineMaxRequiredChecks, "wait_attempts": TimelineMaxWaitAttempts, "wait_observations": TimelineMaxWaitObservations, "waits": TimelineMaxWaits, "wait_text": TimelineMaxWaitText, "tickets": 100, "deliveries": 100, "candidates": 100, "plan_order": TimelineMaxNumber, "external_id": 256, "identity_url": 4096, "views": TimelineMaxViews}, Invariants: []string{
		"Work-local catalogs use typed natural identities and stable opaque IDs. Ticket/delivery metadata patches preserve immutable publication snapshots. Candidate records are immutable and every publication selects its candidate_id explicitly. No current-candidate pointer exists.",
		"Registry patches advance every affected task_version and wait_version once in the same durable commit; unrelated changes, no-ops and exact operation replays do not invalidate readers. New entity operations consume content capacity and closure reserves.",
		"Reject unknown or duplicate keys, null except required current_stage_event_id, trailing JSON and unsafe integer literals; require exact case-sensitive JSON keys.",
		"Sources require an absolute project root, canonical supported host and real native session ID. Session identity is server-owned.",
		"Project identity comes from resolving the actual Source against the trusted registered-root registry to canonical ProjectKey at create/join/accept/recover/load. Distinct alias/worktree roots can share a project; path equality is not project proof. Accepted membership is server-owned; deduplicate (canonical project, host, native session) across alias roots while preserving the first actual Source. No request accepts caller-asserted project proof.",
		"Fields depend on event/operation/target kind. Production validators enforce closed unions and forbid mixed branches; optional fields must be omitted rather than null.",
		"All event, view, evidence, artifact and document references belong to the session; decisions target a prior immutable evidence identity or document revision.",
		"Stage start requires expected_session_version and in_progress; stage update requires expected event and session versions. Paused occurrences resume; completed/failed occurrences are terminal and re-entry links a new event to the latest occurrence.",
		"Exactly one in_progress/in_review/awaiting_validation stage matches current_stage_event_id and work state. Completing/pausing/failing clears its pointer. Parallel reviews do not change foreground.",
		"Reject active foreground withdrawal and closure/direct closed with stage_active. Complete/pause first; no finalization envelope.",
		"Server assigns IDs, sequence, versions, started_at/finished_at and receipt times. occurred_at is immutable agent-reported provenance and never sorts the timeline.",
		"Snapshots append immutable ordered tabs, never replace revisions. Documents are UTF-8 text/markdown, bytes/hash/ownership are checked; same-session content deduplicates without cross-session access.",
		"Operation replay uses persisted request digests before state/version checks. Same operation with different canonical request or bytes conflicts. Private receipts/digests never appear in public session output.",
		"Image and document URLs are separate scoped authenticated local routes bound to known IDs; no absolute URLs, credentials or token query strings.",
		"Native actors declare provenance, not cryptographic process identity. local_viewer never publishes or conducts stage lifecycle; recovery is a separate conditional authority-only operation.",
		"Tasks have one stable work identity and up to five explicit required stage segments; aggregate state and counts never infer approval or cascade lifecycle. Linked snapshot changes increment each affected task_version atomically once, including evidence/documents/CI, excluding no-ops and retries.",
		"CI results use only applicable_attempt_id per required check; server generations advance explicitly. Late old-attempt observations extend history without restoring success. Policy must be confirmed and required checks nonempty for succeeded; a different candidate/policy creates a linked wait.",
		"v3 rejects old protocol and prior state with unsupported_protocol/cutoff_required; no import, reader, conversion or migration is promised.",
		"All content uses manual retention and ttl_seconds zero; capacity rejects new content without pruning history. Persistence/atomic transition behavior is implemented by the server, not this description.",
	}}
	for _, d := range definitions {
		if operation != "" && operation != d.name {
			continue
		}
		example, err := json.Marshal(d.example)
		if err != nil {
			return result, err
		}
		inputSchema := timelineProtocolSchema(reflect.TypeOf(d.input))
		outputSchema := timelineProtocolSchema(reflect.TypeOf(d.output))
		if d.name == "stage_start" {
			properties := inputSchema["properties"].(map[string]any)
			properties["kind"].(map[string]any)["const"] = "stage"
			properties["stage_state"].(map[string]any)["const"] = "in_progress"
			inputSchema["required"] = append(inputSchema["required"].([]string), "stage", "stage_state", "expected_session_version")
		}
		if d.name == "stage_update" {
			inputSchema["required"] = append(inputSchema["required"].([]string), "stage_state", "expected_session_version")
		}
		if reflect.TypeOf(d.output) == reflect.TypeOf(TimelineOperationResult{}) {
			operationKind := TimelineOperationKind(d.name)
			if d.name == "stage_start" {
				operationKind = "append_event"
			}
			if d.name == "stage_update" {
				operationKind = "event_patch"
			}
			properties := outputSchema["properties"].(map[string]any)
			allowed := timelineResultAllowed(operationKind)
			required := outputSchema["required"].([]string)
			for name := range properties {
				if !allowed[name] {
					delete(properties, name)
				} else if name != "evidence_id" && name != "view_ids" && !contains(required, name) {
					required = append(required, name)
				}
			}
			// Map iteration must not change machine-readable contracts between runs.
			sort.Strings(required)
			outputSchema["required"] = required
			properties["kind"].(map[string]any)["const"] = operationKind
		}
		result.Commands = append(result.Commands, TimelineProtocolOperation{Method: timelineCatalogRoute(d.name).method, Path: timelineCatalogRoute(d.name).path, Operation: d.name, Usage: d.usage, InputSchema: inputSchema, OutputSchema: outputSchema, Example: example})
	}
	if len(result.Commands) == 0 {
		return result, timelineError("operation", "unsupported_operation")
	}
	return result, nil
}
func timelineStringPointer(s string) *string { return &s }
func timelineProtocolSchema(t reflect.Type) map[string]any {
	if t == nil {
		return map[string]any{"type": "object", "properties": map[string]any{}, "additionalProperties": false}
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		if t == reflect.TypeOf(TimelineAudit{}) {
			return timelineAuditSchema()
		}
		if t == reflect.TypeOf(TimelineActor{}) {
			return map[string]any{"oneOf": []any{
				map[string]any{"type": "object", "properties": map[string]any{"kind": map[string]any{"const": "native_producer"}, "producer_id": map[string]any{"type": "string", "pattern": timelineOpaqueID.String()}}, "required": []string{"kind", "producer_id"}, "additionalProperties": false},
				map[string]any{"type": "object", "properties": map[string]any{"kind": map[string]any{"const": "local_viewer"}}, "required": []string{"kind"}, "additionalProperties": false},
			}}
		}
		properties := map[string]any{}
		required := []string{}
		for _, f := range timelineJSONFields(t) {
			name, flags, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" || name == "-" {
				continue
			}
			schema := timelineProtocolSchema(f.Type)
			if f.Tag.Get("timeline") == "nullable" {
				schema = map[string]any{"anyOf": []any{schema, map[string]any{"type": "null"}}}
			}
			if name == "mode" {
				schema["const"] = "timeline_api"
			}
			if name == "schema_version" {
				schema["const"] = 3
			}
			if name == "plan_order" || name == "expected_task_version" || name == "task_version" || name == "expected_wait_version" || name == "wait_version" || name == "generation" || name == "order" || name == "expected_version" || name == "expected_session_version" || name == "session_version" || name == "event_version" || name == "sequence" || name == "version" {
				schema["minimum"] = 1
				schema["maximum"] = TimelineMaxNumber
			}
			if strings.HasSuffix(name, "_at") {
				schema["format"] = "date-time"
			}
			if name == "sha256" || name == "request_sha256" {
				schema["pattern"] = timelineSHA256.String()
			}
			if t == reflect.TypeOf(TimelineDocument{}) && name == "size" {
				schema["minimum"] = 0
				schema["maximum"] = TimelineMaxDocumentBytes
			}
			if t == reflect.TypeOf(TimelineDocumentMetadata{}) || t == reflect.TypeOf(TimelineDocument{}) {
				if name == "media_type" {
					schema["const"] = "text/markdown"
				}
			}
			if t == reflect.TypeOf(TimelineArtifactMetadata{}) || t == reflect.TypeOf(TimelineArtifact{}) {
				switch name {
				case "viewport":
					schema["enum"] = timelineViewports
				case "theme":
					schema["enum"] = timelineThemes
				case "media_type":
					schema["enum"] = timelineImageMediaTypes
				}
			}
			if t == reflect.TypeOf(TimelineDocumentSnapshotRequest{}) || t == reflect.TypeOf(TimelineDocumentSnapshot{}) {
				if name == "tabs" {
					schema["minItems"] = 1
					schema["maxItems"] = TimelineMaxDocumentTabs
				}
				if name == "status" {
					schema["enum"] = []string{"draft", "complete"}
				}
			}
			if name == "planned_stage" || name == "stage" {
				schema["enum"] = timelineStages
			}
			if name == "stage_state" {
				schema["enum"] = timelineStageStates
			}
			if name == "document_key" {
				schema["pattern"] = timelineDocumentKey.String()
			}
			if name == "task_key" {
				schema["pattern"] = timelineTaskKey.String()
			}
			if name == "reason" {
				schema["minLength"] = 1
				schema["maxLength"] = 500
			}
			if name == "state" {
				switch t {
				case reflect.TypeOf(TimelineSession{}), reflect.TypeOf(TimelineSessionPatchRequest{}):
					schema["enum"] = timelineStates
				case reflect.TypeOf(TimelineTask{}), reflect.TypeOf(TimelineTaskSegment{}), reflect.TypeOf(TimelineTaskTransitionRequest{}):
					schema["enum"] = timelineTaskStates
				case reflect.TypeOf(TimelineCIWait{}), reflect.TypeOf(TimelineCIObservation{}), reflect.TypeOf(TimelineWaitObserveRequest{}):
					schema["enum"] = timelineWaitStates
				}
			}
			if name == "segments" {
				schema["minItems"] = 1
				schema["maxItems"] = 5
			}
			if name == "steps" && t == reflect.TypeOf(TimelineBlock{}) {
				schema["minItems"] = 1
				schema["maxItems"] = 10
			}
			if name == "producers" {
				schema["minItems"] = 1
				schema["maxItems"] = 16
			}
			if name == "pending_join_requests" {
				schema["maxItems"] = 16
			}
			if name == "tasks" {
				schema["maxItems"] = 100
			}
			if name == "required_checks" || name == "check_keys" {
				schema["maxItems"] = TimelineMaxRequiredChecks
			}
			if name == "waits" {
				schema["maxItems"] = TimelineMaxWaits
			}
			if name == "next_action" || (name == "summary" && (t == reflect.TypeOf(TimelineWaitObserveRequest{}) || t == reflect.TypeOf(TimelineCIObservation{}))) {
				schema["maxLength"] = TimelineMaxWaitText
			}
			if name == "attempts" {
				schema["maxItems"] = TimelineMaxWaitAttempts
			}
			if name == "observations" {
				schema["maxItems"] = 500
			}
			switch name {
			case "ticket_ids", "delivery_ids":
				schema["maxItems"] = 100
				schema["uniqueItems"] = true
				schema["items"] = map[string]any{"type": "string", "pattern": timelineOpaqueID.String()}
				if name == "delivery_ids" {
					schema["minItems"] = 1
				}
			case "tickets", "deliveries", "candidates", "ticket_refs", "delivery_refs":
				schema["maxItems"] = 100
			case "ticket_id", "delivery_id", "candidate_id", "pull_request_delivery_id":
				schema["pattern"] = timelineOpaqueID.String()
			case "external_id":
				schema["minLength"] = 1
				schema["maxLength"] = 256
				schema["pattern"] = `^[^\s\p{Cc}]+$`
			case "work_key":
				schema["minLength"] = 1
				schema["maxLength"] = 128
				schema["pattern"] = `^[^\s\p{Cc}]+$`
				schema["description"] = "Optional name of one work inside a native session, such as a ticket key. The same key from the same native session resumes that work; without it, each new operation_id creates a new work."
			case "external_scope", "repository_url":
				schema["format"] = "uri"
				schema["maxLength"] = 4096
				schema["description"] = "Canonical HTTPS namespace: lowercase scheme/host, no default port, trailing slash, credentials, query, fragment or dot segments. Path case is preserved."
			}
			if t == reflect.TypeOf(TimelineTicketInput{}) || t == reflect.TypeOf(TimelineTicket{}) || t == reflect.TypeOf(TimelineTicketCreateRequest{}) || t == reflect.TypeOf(TimelineTicketSnapshot{}) || t == reflect.TypeOf(TimelineTicketPatchRequest{}) {
				switch name {
				case "key":
					schema["minLength"] = 1
					schema["maxLength"] = 100
				case "title":
					schema["maxLength"] = 300
				case "url":
					schema["maxLength"] = 4096
				}
			}
			if t == reflect.TypeOf(TimelineDeliverySnapshot{}) || t == reflect.TypeOf(TimelineDelivery{}) || t == reflect.TypeOf(TimelineDeliveryCreateRequest{}) || t == reflect.TypeOf(TimelineDeliveryPatchRequest{}) {
				switch name {
				case "kind":
					schema["enum"] = []string{"pull_request", "commit"}
				case "status":
					schema["enum"] = []string{"open", "merged", "local"}
				case "label":
					schema["minLength"] = 1
					schema["maxLength"] = 200
				case "url":
					schema["format"] = "uri"
					schema["maxLength"] = 4096
				}
			}
			if (t == reflect.TypeOf(TimelineCandidate{}) || t == reflect.TypeOf(TimelineCandidateRecord{}) || t == reflect.TypeOf(TimelineCandidateCreateRequest{})) && contains([]string{"branch", "commit", "base"}, name) {
				schema["minLength"] = 1
				schema["maxLength"] = 256
			}
			if name == "title" && (t == reflect.TypeOf(TimelineAppendEventRequest{}) || t == reflect.TypeOf(TimelineEvent{}) || t == reflect.TypeOf(TimelineEventPatchRequest{})) {
				schema["maxLength"] = 200
				schema["description"] = "For stages, title is verbatim context only: do not repeat the reserved stage heading; use an empty string when no extra context is needed. Other event kinds require a nonempty free title. Patch omission preserves context; an empty string clears stage context only, validated against the target event."
			}
			properties[name] = schema
			if !strings.Contains(flags, "omitempty") {
				required = append(required, name)
			}
		}
		result := map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
		if t == reflect.TypeOf(TimelineAppendEventRequest{}) || t == reflect.TypeOf(TimelineEvent{}) {
			result["allOf"] = []any{map[string]any{
				"if":   map[string]any{"properties": map[string]any{"kind": map[string]any{"const": "stage"}}, "required": []string{"kind"}},
				"else": map[string]any{"properties": map[string]any{"title": map[string]any{"minLength": 1}}},
			}}
		}
		return result
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": "array", "items": timelineProtocolSchema(t.Elem())}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": timelineProtocolSchema(t.Elem())}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int64:
		return map[string]any{"type": "integer"}
	default:
		schema := map[string]any{"type": "string"}
		switch t {
		case reflect.TypeOf(TimelineHost("")):
			schema["enum"] = timelineHosts
		case reflect.TypeOf(TimelineState("")):
			schema["enum"] = timelineStates
		case reflect.TypeOf(TimelineEventKind("")):
			schema["enum"] = timelineEventKinds
		case reflect.TypeOf(TimelineStage("")):
			schema["enum"] = timelineStages
		case reflect.TypeOf(TimelineStageState("")):
			schema["enum"] = timelineStageStates
		}
		return schema
	}
}

func timelineAuditSchema() map[string]any {
	fields := map[string]reflect.Type{}
	for _, f := range timelineJSONFields(reflect.TypeOf(TimelineAudit{})) {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		fields[name] = f.Type
	}
	branch := func(ops, required, optional []string) map[string]any {
		properties := map[string]any{"operation": map[string]any{"type": "string", "enum": ops}, "reason": map[string]any{"type": "string", "minLength": 1, "maxLength": 500}, "target": timelineProtocolSchema(fields["target"])}
		for _, k := range append(append([]string{}, required...), optional...) {
			properties[k] = timelineProtocolSchema(fields[k])
		}
		return map[string]any{"type": "object", "properties": properties, "required": append([]string{"operation", "reason"}, required...), "additionalProperties": false}
	}
	return map[string]any{"oneOf": []any{
		branch([]string{"stage_start", "stage_update", "session_patch", "closure"}, []string{"previous_state", "next_state"}, nil),
		branch([]string{"producer_accept", "recover_lead"}, []string{"request_id", "next_producer_id"}, nil),
		branch([]string{"foreground_transfer"}, []string{"previous_producer_id", "next_producer_id"}, nil),
		branch([]string{"recover_lead"}, []string{"request_id", "previous_producer_id", "next_producer_id"}, nil),
		branch([]string{"task_transition"}, []string{"previous_state", "next_state"}, []string{"block", "resolution", "previous_segment", "next_segment", "plan_revision_id"}),
		branch([]string{"session_rename", "session_archive", "session_pin", "session_delete"}, nil, nil),
	}}
}
