package viewer

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
)

// TimelineProtocolDescription is offline metadata; it never starts or contacts a viewer.
// Schemas describe exact structural fields from the production Go DTOs. The
// semantic invariants remain authoritative validator rules, not JSON Schema promises.
type TimelineProtocolDescription struct {
	ProtocolVersion int                         `json:"protocol_version"`
	SchemaDialect   string                      `json:"schema_dialect"`
	Capabilities    []string                    `json:"capabilities"`
	Retention       string                      `json:"retention"`
	TTLSeconds      int                         `json:"ttl_seconds"`
	Commands        []TimelineProtocolOperation `json:"commands"`
	Limits          map[string]int64            `json:"limits"`
	Invariants      []string                    `json:"invariants"`
}
type TimelineProtocolOperation struct {
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
	one := int64(1)
	source := TimelineSource{ProjectRoot: "/workspace/example", Host: "codex", NativeSessionID: "native-example"}
	create := TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "create-example", Source: source, Title: "Example", Summary: "Scope", Tickets: []TimelineTicket{}}
	stage := TimelineAppendEventRequest{SchemaVersion: 1, OperationID: "stage-example", Kind: "stage", Title: "Plan", Summary: "Prepare the plan", TicketRefs: []TimelineTicketRef{}, Stage: "plan", StageState: "in_progress", ExpectedSessionVersion: &one}
	state := TimelineStageState("completed")
	patch := TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "complete-example", ExpectedVersion: 1, StageState: &state, ExpectedSessionVersion: &one}
	mutation := TimelineMutationRequest{SchemaVersion: 1, OperationID: "delete-example", ExpectedVersion: 1}
	definitions := []timelineProtocolDefinition{
		{"create_session", "hive view session create --json <file|->", TimelineCreateSessionRequest{}, TimelineSessionResponse{}, create},
		{"session_get", "hive view session get <session-id>", nil, TimelineSessionResponse{}, map[string]any{}},
		{"session_patch", "hive view session update <session-id> --json <file|->", TimelineSessionPatchRequest{}, TimelineOperationResult{}, TimelineSessionPatchRequest{SchemaVersion: 1, OperationID: "update-example", ExpectedVersion: 1, Title: timelineStringPointer("Updated title")}},
		{"session_delete", "hive view session delete <session-id> --json <file|->", TimelineMutationRequest{}, TimelineOperationResult{}, mutation},
		{"append_event", "hive view event add <session-id> --json <file|->", TimelineAppendEventRequest{}, TimelineOperationResult{}, TimelineAppendEventRequest{SchemaVersion: 1, OperationID: "request-example", Kind: "request", Title: "Request", Summary: "Initial scope", TicketRefs: []TimelineTicketRef{}}},
		{"event_patch", "hive view event update <session-id> <event-id> --json <file|->", TimelineEventPatchRequest{}, TimelineOperationResult{}, TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "copy-example", ExpectedVersion: 1, Title: timelineStringPointer("Corrected title")}},
		{"event_withdraw", "hive view event withdraw <session-id> <event-id> --json <file|->", TimelineMutationRequest{}, TimelineOperationResult{}, mutation},
		{"artifact_add", "hive view artifact add <session-id> --json <file|-> --file <image>", TimelineArtifactMetadata{}, TimelineOperationResult{}, TimelineArtifactMetadata{SchemaVersion: 1, OperationID: "image-example", CapturedAt: "2026-01-01T00:00:00Z", Candidate: TimelineCandidate{Branch: "feature/example", Commit: "abc1234", Base: "development"}, Viewport: "desktop", Theme: "light", MediaType: "image/png"}},
		{"document_add", "hive view document add <session-id> --json <file|-> --file <markdown>", TimelineDocumentMetadata{}, TimelineOperationResult{}, TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "document-example", MediaType: "text/markdown"}},
		{"document_get", "hive view document get <session-id> <document-id>", nil, TimelineDocumentResponse{}, map[string]any{}},
		{"stage_start", "hive view stage start <session-id> --json <file|->", TimelineAppendEventRequest{}, TimelineOperationResult{}, stage},
		{"stage_update", "hive view stage update <session-id> <event-id> --json <file|->", TimelineEventPatchRequest{}, TimelineOperationResult{}, patch},
		{"document_snapshot", "hive view stage snapshot <session-id> <event-id> --json <file|->", TimelineDocumentSnapshotRequest{}, TimelineOperationResult{}, TimelineDocumentSnapshotRequest{SchemaVersion: 1, OperationID: "snapshot-example", ExpectedVersion: 1, Status: "draft", Tabs: []TimelineDocumentTab{{Label: "Proposal", DocumentID: "document-example"}}}},
	}
	result := TimelineProtocolDescription{ProtocolVersion: 1, SchemaDialect: "https://json-schema.org/draft/2020-12/schema", Retention: "manual", TTLSeconds: 0, Capabilities: []string{"timeline_api", "conditional_mutations", "immutable_evidence", "stage_lifecycle", "document_snapshots", "manual_retention"}, Commands: []TimelineProtocolOperation{}, Limits: map[string]int64{"body_bytes": TimelineMaxBody, "document_bytes": TimelineMaxDocumentBytes, "document_tabs": TimelineMaxDocumentTabs, "documents": TimelineMaxDocuments, "document_snapshots": TimelineMaxDocumentSnapshots, "aggregate_document_bytes": TimelineMaxAggregateDocumentBytes, "events": TimelineMaxEvents, "views": TimelineMaxViews}, Invariants: []string{
		"Reject unknown or duplicate keys, null except required current_stage_event_id, trailing JSON and unsafe integer literals; require exact case-sensitive JSON keys.",
		"Sources require an absolute project root, canonical supported host and real native session ID. Session identity is server-owned.",
		"Fields depend on event/operation/target kind. Production validators enforce closed unions and forbid mixed branches; optional fields must be omitted rather than null.",
		"All event, view, evidence, artifact and document references belong to the session; decisions target a prior immutable evidence identity or document revision.",
		"Stage start requires expected_session_version and in_progress; stage update requires expected event and session versions. Paused occurrences resume; completed/failed occurrences are terminal and re-entry links a new event to the latest occurrence.",
		"Exactly one in_progress stage matches current_stage_event_id in active sessions. Starting/resuming pauses the old foreground atomically. Completing/pausing/failing clears its pointer. awaiting_validation atomically pauses and clears; resume sets active.",
		"Reject active foreground withdrawal and closure/direct closed with stage_active. Complete/pause first; no finalization envelope.",
		"Server assigns IDs, sequence, versions, started_at/finished_at and receipt times. occurred_at is immutable agent-reported provenance and never sorts the timeline.",
		"Snapshots append immutable ordered tabs, never replace revisions. Documents are UTF-8 text/markdown, bytes/hash/ownership are checked; same-session content deduplicates without cross-session access.",
		"Operation replay uses persisted request digests before state/version checks. Same operation with different canonical request or bytes conflicts. Private receipts/digests never appear in public session output.",
		"Image and document URLs are separate scoped authenticated local routes bound to known IDs; no absolute URLs, credentials or token query strings.",
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
		result.Commands = append(result.Commands, TimelineProtocolOperation{Operation: d.name, Usage: d.usage, InputSchema: inputSchema, OutputSchema: outputSchema, Example: example})
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
		properties := map[string]any{}
		required := []string{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
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
				schema["const"] = 1
			}
			if name == "expected_version" || name == "expected_session_version" || name == "session_version" || name == "event_version" || name == "sequence" || name == "version" {
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
			properties[name] = schema
			if !strings.Contains(flags, "omitempty") {
				required = append(required, name)
			}
		}
		return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
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
