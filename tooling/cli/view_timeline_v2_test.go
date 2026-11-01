package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

func TestViewTimelineV2NativeCommandShapes(t *testing.T) {
	id := strings.Repeat("a", 64)
	for _, c := range []struct {
		args            []string
		operation, path string
	}{
		{[]string{"producer", "request", id}, "join_request", "/join-requests"},
		{[]string{"producer", "accept", id}, "producer_accept", "/producers/accept"},
		{[]string{"producer", "transfer", id}, "foreground_transfer", "/producers/transfer"},
		{[]string{"task", "create", id}, "task_create", "/tasks"},
		{[]string{"task", "update", id, "task-one"}, "task_patch", "/tasks/task-one"},
		{[]string{"task", "transition", id, "task-one"}, "task_transition", "/tasks/task-one/transitions"},
		{[]string{"wait", "create", id}, "wait_create", "/waits"},
		{[]string{"wait", "attempt", id, "wait-one"}, "wait_attempt_start", "/waits/wait-one/attempts"},
		{[]string{"wait", "observe", id, "wait-one"}, "wait_observe", "/waits/wait-one/observations"},
	} {
		got, rest, err := parseViewTimelineCommand(c.args)
		if err != nil || len(rest) != 0 || got.operation != c.operation || got.path != "/api/timeline/sessions/"+id+c.path {
			t.Fatalf("%v: %+v %v", c.args, got, err)
		}
	}
	if _, _, err := parseViewTimelineCommand([]string{"producer", "recover", id}); err == nil {
		t.Fatal("native CLI exposes local recovery")
	}
}
func TestViewTimelineV2OfflineInput(t *testing.T) {
	raw, err := os.ReadFile("../viewer/testdata/timeline_contract/v2-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Action string `json:"action"`
		Kind   string `json:"kind"`
		Valid  bool   `json:"valid"`
		JSON   string `json:"json"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if c.Action != "shape" || c.Kind == "recover_lead" {
			continue
		}
		if err := validateViewTimelineInput(c.Kind, []byte(c.JSON)); (err == nil) != c.Valid {
			t.Fatalf("%s valid=%v: %v", c.Kind, c.Valid, err)
		}
	}
	if err := validateViewTimelineInput("session_patch", []byte(`{"schema_version":2,"operation_id":"rename","actor":{"kind":"local_viewer"},"expected_version":1,"title":"Updated"}`)); err == nil {
		t.Fatal("native CLI bypassed producer identity")
	}
}

func TestViewTimelineV2RejectsCompleteV1Write(t *testing.T) {
	raw, err := os.ReadFile("../viewer/testdata/timeline_contract/create-session.json")
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	v["schema_version"] = 1
	raw, err = json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateViewTimelineInput("create_session", raw); viewer.TimelineContractCode(err) != "unsupported_protocol" {
		t.Fatal("v1 write was not rejected before transport", err)
	}
}
