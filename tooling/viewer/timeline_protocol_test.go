package viewer

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTimelineProtocolOfflineExamplesUseProductionDecoders(t *testing.T) {
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	if contract.ProtocolVersion != 3 || contract.Retention != "manual" || contract.TTLSeconds != 0 || len(contract.Commands) != 28 {
		t.Fatal("invalid contract metadata")
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
				_, err = DecodeTimelineV3Operation(command.Operation, command.Example)
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

func TestTimelineProtocolV3ApprovedLimitsAndAuditUnion(t *testing.T) {
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]int64{"waits": 50, "wait_attempts": 20, "required_checks": 50, "wait_text": 500} {
		if contract.Limits[key] != want {
			t.Fatalf("%s limit: got %d want %d", key, contract.Limits[key], want)
		}
	}
	cases := []struct {
		typ   reflect.Type
		field string
		want  int
	}{{reflect.TypeOf(TimelineSession{}), "waits", 50}, {reflect.TypeOf(TimelineCIWait{}), "attempts", 20}, {reflect.TypeOf(TimelineCheckPolicy{}), "required_checks", 50}, {reflect.TypeOf(TimelineWaitAttemptStartRequest{}), "check_keys", 50}}
	for _, c := range cases {
		props := timelineProtocolSchema(c.typ)["properties"].(map[string]any)
		if props[c.field].(map[string]any)["maxItems"] != c.want {
			t.Fatalf("%s.%s bound drift", c.typ.Name(), c.field)
		}
	}
	audit := timelineProtocolSchema(reflect.TypeOf(TimelineAudit{}))
	branches, ok := audit["oneOf"].([]any)
	if !ok || len(branches) != 6 {
		t.Fatal("audit is not a closed operation union")
	}
	for _, b := range branches {
		if b.(map[string]any)["additionalProperties"] != false {
			t.Fatal("audit permits unknown keys")
		}
	}
	for _, op := range []string{"wait_create", "wait_attempt_start", "wait_observe"} {
		if !strings.Contains(contract.Authority[op], "assigned delivery-segment") {
			t.Fatal("wait authority drift", op)
		}
	}
}

func TestTimelineCandidateSchemaPreservesInheritedBounds(t *testing.T) {
	c, err := TimelineProtocolContract("candidate_create")
	if err != nil {
		t.Fatal(err)
	}
	props := c.Commands[0].InputSchema["properties"].(map[string]any)
	if props["reason"].(map[string]any)["maxLength"] != 500 {
		t.Fatal("candidate schema overrides reason bound", props["reason"])
	}
	for _, field := range []string{"branch", "commit", "base"} {
		if props[field].(map[string]any)["maxLength"] != 256 {
			t.Fatal("candidate descriptor bound missing", field)
		}
	}
	var request TimelineCandidateCreateRequest
	if err = DecodeTimelineJSON(c.Commands[0].Example, &request); err != nil {
		t.Fatal(err)
	}
	request.Reason = strings.Repeat("r", 500)
	if err = ValidateTimelineV3Operation(&request); err != nil {
		t.Fatal("500-character reason rejected", err)
	}
}

func TestTimelineProtocolEveryOperationHasKnownFlow(t *testing.T) {
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, c := range contract.Commands {
		known[c.Operation] = true
	}
	for _, flow := range TimelineProtocolFlows() {
		operations, err := timelineFlowMembers(flow)
		if err != nil || len(operations) == 0 {
			t.Fatalf("flow %s: %v", flow, err)
		}
		for _, name := range operations {
			if !known[name] {
				t.Fatalf("flow %s names unknown operation %s", flow, name)
			}
		}
	}
	list := TimelineProtocolList()
	if len(list.Operations) != len(contract.Commands) {
		t.Fatalf("list has %d entries for %d operations", len(list.Operations), len(contract.Commands))
	}
	for _, entry := range list.Operations {
		if len(entry.Flows) == 0 || !known[entry.Operation] || entry.Usage == "" {
			t.Fatalf("operation %s lacks a flow or is unknown", entry.Operation)
		}
	}
	if _, err := timelineFlowMembers("nope"); err == nil {
		t.Fatal("unknown flow accepted")
	}
}

func TestTimelineProtocolBriefCarriesFieldsEnumsAndExample(t *testing.T) {
	brief, err := TimelineProtocolBrief("plan")
	if err != nil || brief.Flow != "plan" || brief.ProtocolVersion != 3 {
		t.Fatal(err)
	}
	byName := map[string]TimelineProtocolBriefOperation{}
	for _, o := range brief.Operations {
		byName[o.Operation] = o
	}
	for _, want := range []string{"stage_start", "stage_update", "document_add", "ticket_create", "session_get"} {
		if _, ok := byName[want]; !ok {
			t.Fatalf("plan brief lacks %s", want)
		}
	}
	if _, ok := byName["task_create"]; ok {
		t.Fatal("plan brief carries a build-only operation")
	}
	update := byName["stage_update"]
	if update.Usage == "" || len(update.Example) == 0 || !contains(update.Required, "stage_state") || !contains(update.Required, "reason") {
		t.Fatalf("incomplete stage_update brief: %+v", update)
	}
	if len(update.Enums["stage_state"]) == 0 {
		t.Fatal("stage_state enum missing")
	}
	if _, ok := byName["create_session"].Enums["source.host"]; !ok {
		t.Fatal("nested enum path source.host missing")
	}
	raw, _ := json.Marshal(brief)
	if strings.Contains(string(raw), "output_schema") {
		t.Fatal("brief carries output schemas")
	}
	if _, err := TimelineProtocolBrief("nope"); err == nil {
		t.Fatal("unknown flow accepted")
	}
}

func TestTimelineProtocolBriefStaysUnderSizeLimit(t *testing.T) {
	for _, flow := range TimelineProtocolFlows() {
		brief, err := TimelineProtocolBrief(flow)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(brief)
		if len(raw) >= 10000 {
			t.Fatalf("%s brief is %d bytes", flow, len(raw))
		}
	}
}

func TestTimelineProtocolRulesBlockHoldsSharedRules(t *testing.T) {
	rules, err := json.Marshal(TimelineProtocolRules())
	if err != nil {
		t.Fatal(err)
	}
	var keys map[string]json.RawMessage
	_ = json.Unmarshal(rules, &keys)
	for _, k := range []string{"protocol_version", "storage_version", "schema_dialect", "localization", "capabilities", "retention", "ttl_seconds", "authority", "error_codes", "limits", "invariants"} {
		if _, ok := keys[k]; !ok {
			t.Fatalf("rules lack %s", k)
		}
	}
	if _, ok := keys["commands"]; ok {
		t.Fatal("rules carry commands")
	}
}

func TestTimelineProtocolStageReasonRequiredAndConstsTyped(t *testing.T) {
	for _, op := range []string{"stage_start", "stage_update"} {
		full, err := TimelineProtocolContract(op)
		if err != nil || !contains(full.Commands[0].InputSchema["required"].([]string), "reason") {
			t.Fatalf("%s schema does not require reason: %v", op, err)
		}
	}
	for _, flow := range TimelineProtocolFlows() {
		brief, err := TimelineProtocolBrief(flow)
		if err != nil {
			t.Fatal(err)
		}
		for _, o := range brief.Operations {
			if _, ok := o.Enums["schema_version"]; ok {
				t.Fatalf("%s lists the integer schema_version as a string value", o.Operation)
			}
		}
	}
}
