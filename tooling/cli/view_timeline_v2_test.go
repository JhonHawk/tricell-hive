package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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

func TestViewTimelineV2NativeDeletionPointsToBrowser(t *testing.T) {
	var out bytes.Buffer
	err := runViewTimeline([]string{"session", "delete", strings.Repeat("a", 64), "--json", "-", "--state-dir", filepath.Join(t.TempDir(), "missing")}, &out, strings.NewReader(`{"schema_version":2}`))
	if err == nil || !strings.Contains(err.Error(), "authenticated browser") || out.Len() != 0 {
		t.Fatal("native deletion did not preserve browser-only boundary", err)
	}
}

func TestViewTimelineV2OfflineDeletionUsageIsBrowserOnly(t *testing.T) {
	var out bytes.Buffer
	if err := runViewTimeline([]string{"contract", "--operation", "session_delete"}, &out, nil); err != nil {
		t.Fatal(err)
	}
	var contract viewer.TimelineProtocolDescription
	if err := json.Unmarshal(out.Bytes(), &contract); err != nil {
		t.Fatal(err)
	}
	if len(contract.Commands) != 1 || strings.Contains(contract.Commands[0].Usage, "hive view session delete") || !strings.Contains(contract.Commands[0].Usage, "Local viewer only") {
		t.Fatal("offline catalog advertised unsupported native deletion")
	}
}

func TestViewTimelineV2CompatibilityRejectsBeforeTransportMutation(t *testing.T) {
	state, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = viewer.PrepareState(state); err != nil {
		t.Fatal(err)
	}
	paths := newViewPaths(state)
	lock, err := viewer.AcquireLock(paths.lock, false)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	var protocol atomic.Int32
	protocol.Store(2)
	var healthCalls, mutationCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/health" {
			healthCalls.Add(1)
			json.NewEncoder(w).Encode(map[string]any{"pid": os.Getpid(), "state_dir": state, "timeline_protocol_version": protocol.Load()})
			return
		}
		mutationCalls.Add(1)
		w.WriteHeader(500)
	}))
	defer server.Close()
	info := serverInfo{PID: os.Getpid(), Port: server.Listener.Addr().(*net.TCPAddr).Port, Token: "fixture-compatibility-token", StateDir: state}
	if err = writeServerInfo(paths, info); err != nil {
		t.Fatal(err)
	}
	request := viewer.TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "compatibility", Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Title: "Compatibility fixture", Summary: "Must preserve input", Tickets: []viewer.TimelineTicket{}}
	input := filepath.Join(t.TempDir(), "request.json")
	raw, _ := json.Marshal(request)
	os.WriteFile(input, raw, 0600)
	var out bytes.Buffer
	err = runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported_protocol") || healthCalls.Load() != 0 || mutationCalls.Load() != 0 || out.Len() != 0 {
		t.Fatal("v1 input was not rejected offline", err)
	}
	request.SchemaVersion = 2
	raw, _ = json.Marshal(request)
	os.WriteFile(input, raw, 0600)
	protocol.Store(1)
	err = runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "fresh v2 state directory") || healthCalls.Load() != 1 || mutationCalls.Load() != 0 || out.Len() != 0 {
		t.Fatal("v2 client sent mutation to v1 server", err)
	}
	preserved, _ := os.ReadFile(input)
	if !bytes.Equal(preserved, raw) {
		t.Fatal("compatibility failure changed input")
	}
}
