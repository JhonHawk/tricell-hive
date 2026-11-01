package viewer

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestTimelineProtocolOfflineExamplesUseProductionDecoders(t *testing.T) {
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	if contract.ProtocolVersion != 1 || contract.Retention != "manual" || contract.TTLSeconds != 0 || len(contract.Commands) != 13 {
		t.Fatalf("invalid contract metadata: %+v", contract)
	}
	for _, command := range contract.Commands {
		t.Run(command.Operation, func(t *testing.T) {
			var err error
			switch command.Operation {
			case "create_session":
				_, err = DecodeTimelineCreateSession(command.Example)
			case "session_patch":
				_, err = DecodeTimelineSessionPatch(command.Example)
			case "session_delete", "event_withdraw":
				_, err = DecodeTimelineMutation(command.Example)
			case "append_event", "stage_start":
				_, err = DecodeTimelineAppendEvent(command.Example)
			case "event_patch", "stage_update":
				_, err = DecodeTimelineEventPatch(command.Example)
			case "artifact_add":
				_, err = DecodeTimelineArtifactMetadata(command.Example)
			case "document_add":
				_, err = DecodeTimelineDocumentMetadata(command.Example)
			case "document_snapshot":
				_, err = DecodeTimelineDocumentSnapshot(command.Example)
			case "session_get", "document_get":
				if string(command.Example) != "{}" {
					t.Fatal("read must have no body")
				}
			default:
				t.Fatal("unknown command")
			}
			if err != nil {
				t.Fatal(err)
			}
			narrowed, err := TimelineProtocolContract(command.Operation)
			if err != nil || len(narrowed.Commands) != 1 || !reflect.DeepEqual(narrowed.Commands[0], command) {
				t.Fatal("narrowed contract drift", err)
			}
		})
	}
	if _, err := TimelineProtocolContract("unsupported"); err == nil {
		t.Fatal("unsupported operation accepted")
	}
	if _, err := json.Marshal(contract); err != nil {
		t.Fatal(err)
	}
}
func TestTimelineProtocolSchemaDerivedFieldsAndNullablePointer(t *testing.T) {
	schema := timelineProtocolSchema(reflect.TypeOf(TimelineSession{}))
	properties := schema["properties"].(map[string]any)
	if len(properties) != reflect.TypeOf(TimelineSession{}).NumField() {
		t.Fatal("reflection lost fields")
	}
	pointer := properties["current_stage_event_id"].(map[string]any)
	if len(pointer["anyOf"].([]any)) != 2 {
		t.Fatal("nullable pointer missing")
	}
	required := schema["required"].([]string)
	if !contains(required, "current_stage_event_id") || !contains(required, "documents") {
		t.Fatal("required session fields missing")
	}
	input := timelineProtocolSchema(reflect.TypeOf(TimelineDocumentSnapshotRequest{}))
	if input["additionalProperties"] != false {
		t.Fatal("schema permits unknown keys")
	}
}

func TestTimelineProtocolArtifactEnumsMatchUploadValidation(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(TimelineArtifactMetadata{}), reflect.TypeOf(TimelineArtifact{})} {
		properties := timelineProtocolSchema(typ)["properties"].(map[string]any)
		for field, want := range map[string][]string{"viewport": timelineViewports, "theme": timelineThemes, "media_type": timelineImageMediaTypes} {
			if got := properties[field].(map[string]any)["enum"]; !reflect.DeepEqual(got, want) {
				t.Fatalf("%s.%s enum mismatch: %v", typ.Name(), field, got)
			}
		}
	}
	contract, err := TimelineProtocolContract("artifact_add")
	if err != nil {
		t.Fatal(err)
	}
	var example TimelineArtifactMetadata
	if err := json.Unmarshal(contract.Commands[0].Example, &example); err != nil {
		t.Fatal(err)
	}
	for _, viewport := range timelineViewports {
		for _, theme := range timelineThemes {
			for _, mediaType := range timelineImageMediaTypes {
				example.Viewport, example.Theme, example.MediaType = viewport, theme, mediaType
				if err := example.Validate(); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestTimelineProtocolResultSchemasUseProductionAllowedFields(t *testing.T) {
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range contract.Commands {
		properties := c.OutputSchema["properties"].(map[string]any)
		kind, ok := properties["kind"].(map[string]any)
		if !ok {
			continue
		}
		operationKind, ok := kind["const"].(TimelineOperationKind)
		if !ok {
			t.Fatal("missing result kind")
		}
		allowed := timelineResultAllowed(operationKind)
		if len(properties) != len(allowed) {
			t.Fatal("schema allowed fields drift", c.Operation)
		}
		for name := range properties {
			if !allowed[name] {
				t.Fatal("schema leaked field", name)
			}
		}
	}
}
