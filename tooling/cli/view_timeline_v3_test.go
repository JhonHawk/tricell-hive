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

func TestViewTimelineV3NativeCommandShapes(t *testing.T) {
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
func TestViewTimelineV3OfflineInput(t *testing.T) {
	raw, err := os.ReadFile("../viewer/testdata/timeline_contract/v3-cases.json")
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
	if err := validateViewTimelineInput("session_patch", []byte(`{"schema_version":3,"operation_id":"rename","actor":{"kind":"local_viewer"},"expected_version":1,"title":"Updated"}`)); err == nil {
		t.Fatal("native CLI bypassed producer identity")
	}
}

func TestViewTimelineV3RejectsCompleteV1Write(t *testing.T) {
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

func TestViewTimelineV3NativeDeletionPointsToBrowser(t *testing.T) {
	var out bytes.Buffer
	err := runViewTimeline([]string{"session", "delete", strings.Repeat("a", 64), "--json", "-", "--state-dir", filepath.Join(t.TempDir(), "missing")}, &out, strings.NewReader(`{"schema_version":3}`))
	if err == nil || !strings.Contains(err.Error(), "authenticated browser") || out.Len() != 0 {
		t.Fatal("native deletion did not preserve browser-only boundary", err)
	}
}

func TestViewTimelineV3OfflineDeletionUsageIsBrowserOnly(t *testing.T) {
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

func TestViewTimelineV3CompatibilityRejectsBeforeTransportMutation(t *testing.T) {
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
	request := viewer.TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "compatibility", Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Title: "Compatibility fixture", Summary: "Must preserve input", Tickets: []viewer.TimelineTicketInput{}}
	input := filepath.Join(t.TempDir(), "request.json")
	raw, _ := json.Marshal(request)
	os.WriteFile(input, raw, 0600)
	var out bytes.Buffer
	err = runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported_protocol") || healthCalls.Load() != 0 || mutationCalls.Load() != 0 || out.Len() != 0 {
		t.Fatal("v1 input was not rejected offline", err)
	}
	request.SchemaVersion = 3
	raw, _ = json.Marshal(request)
	os.WriteFile(input, raw, 0600)
	protocol.Store(1)
	err = runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "fresh v3 state directory") || healthCalls.Load() != 1 || mutationCalls.Load() != 0 || out.Len() != 0 {
		t.Fatal("v3 client sent mutation to v1 server", err)
	}
	protocol.Store(2)
	err = runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "fresh v3 state directory") || healthCalls.Load() != 2 || mutationCalls.Load() != 0 {
		t.Fatal("v3 client sent mutation to v2 server", err)
	}
	preserved, _ := os.ReadFile(input)
	if !bytes.Equal(preserved, raw) {
		t.Fatal("compatibility failure changed input")
	}
}

func TestViewContractPartsAreCompactAndExclusive(t *testing.T) {
	run := func(args ...string) (string, error) {
		var out bytes.Buffer
		err := runViewTimeline(append([]string{"contract"}, args...), &out, nil)
		return out.String(), err
	}
	full, err := run()
	if err != nil || strings.Contains(full, "\n  ") || strings.Count(strings.TrimSuffix(full, "\n"), "\n") != 0 {
		t.Fatalf("full contract is not compact: %v", err)
	}
	one, err := run("--operation", "stage_start")
	if err != nil || strings.Contains(one, "\n  ") {
		t.Fatalf("operation output is not compact: %v", err)
	}
	var indented bytes.Buffer
	if json.Indent(&indented, []byte(one), "", "  ") != nil {
		t.Fatal("operation output is not JSON")
	}
	brief, err := run("--for", "build")
	var parsed struct {
		Flow       string           `json:"flow"`
		Operations []map[string]any `json:"operations"`
	}
	if err != nil || json.Unmarshal([]byte(brief), &parsed) != nil || parsed.Flow != "build" || len(parsed.Operations) == 0 {
		t.Fatalf("invalid --for output: %v", err)
	}
	for _, o := range parsed.Operations {
		if _, ok := o["output_schema"]; ok {
			t.Fatal("brief carries output_schema")
		}
	}
	list, err := run("--list")
	var listed struct {
		Operations []struct{ Operation string } `json:"operations"`
	}
	var complete viewer.TimelineProtocolDescription
	if err != nil || json.Unmarshal([]byte(full), &complete) != nil || len(complete.Commands) == 0 || json.Unmarshal([]byte(list), &listed) != nil || len(listed.Operations) != len(complete.Commands) {
		t.Fatalf("invalid --list output: %v", err)
	}
	rules, err := run("--section", "rules")
	if err != nil || !strings.Contains(rules, `"error_codes"`) || strings.Contains(rules, `"commands"`) {
		t.Fatalf("invalid rules block: %v", err)
	}
	for _, args := range [][]string{{"--for", "build", "--list"}, {"--for", "build", "--operation", "stage_start"}, {"--list", "--section", "rules"}, {"--operation", "stage_start", "--section", "rules"}, {"--for", "nope"}, {"--section", "nope"}, {"--for="}, {"--section="}, {"--list=false"}} {
		out, err := run(args...)
		if err == nil || out != "" {
			t.Fatalf("%v accepted", args)
		}
	}
	if _, err := run("--for", "nope"); err == nil || !strings.Contains(err.Error(), "research") {
		t.Fatalf("unknown flow error does not list valid values: %v", err)
	}
}

func TestViewStageStartWithoutReasonNamesTheField(t *testing.T) {
	input := `{"actor":{"kind":"native_producer","producer_id":"producer-example"},"schema_version":3,"operation_id":"stage-example","kind":"stage","title":"","summary":"Prepare","ticket_ids":[],"stage":"plan","stage_state":"in_progress","expected_session_version":1}`
	var out bytes.Buffer
	err := runViewTimeline([]string{"stage", "start", strings.Repeat("a", 64), "--json", "-", "--state-dir", filepath.Join(t.TempDir(), "missing")}, &out, strings.NewReader(input))
	if err == nil || err.Error() != "invalid timeline input: invalid_text on reason" {
		t.Fatalf("got %v", err)
	}
}

// Operations whose example the native CLI validator rejects by design, or that
// have no input, so the brief check cannot run on them. Any other skipped
// operation fails the test.
var viewBriefSkipped = map[string]string{
	"session_delete": "local_viewer actor is rejected by the native validator",
	"recover_lead":   "local_viewer actor is rejected by the native validator",
	"session_get":    "no input",
	"document_get":   "no input",
}

// Operations that must change at least one field: dropping the example's only
// optional field leaves an empty patch, which the validator rejects.
var viewBriefMustChange = map[string]bool{
	"session_patch": true, "event_patch": true, "ticket_patch": true, "task_patch": true, "delivery_patch": true,
}

// The brief's required and optional labels must match the production
// validators: dropping a required field from an operation's example is
// rejected, and dropping an optional one is accepted, except for the
// operations in viewBriefMustChange when it is the example's only optional field.
func TestViewContractBriefMatchesValidator(t *testing.T) {
	seen := map[string]bool{}
	for _, flow := range viewer.TimelineProtocolFlows() {
		brief, err := viewer.TimelineProtocolBrief(flow)
		if err != nil {
			t.Fatal(err)
		}
		for _, op := range brief.Operations {
			if seen[op.Operation] {
				continue
			}
			seen[op.Operation] = true
			var example map[string]json.RawMessage
			if json.Unmarshal(op.Example, &example) != nil || len(example) == 0 || validateViewTimelineInput(op.Operation, op.Example) != nil {
				if _, allowed := viewBriefSkipped[op.Operation]; !allowed {
					t.Errorf("%s: example skipped but not in the allow-list (%v)", op.Operation, validateViewTimelineInput(op.Operation, op.Example))
				}
				continue
			}
			if reason, skipped := viewBriefSkipped[op.Operation]; skipped {
				t.Errorf("%s is allow-listed (%s) but its example validates; remove it from the list", op.Operation, reason)
			}
			var optionalInExample []string
			for name := range example {
				if contains(op.Optional, name) {
					optionalInExample = append(optionalInExample, name)
				}
			}
			for name := range example {
				without := map[string]json.RawMessage{}
				for k, v := range example {
					if k != name {
						without[k] = v
					}
				}
				raw, _ := json.Marshal(without)
				rejected := validateViewTimelineInput(op.Operation, raw) != nil
				switch {
				case contains(op.Required, name) && !rejected:
					t.Errorf("%s: brief marks %s required, validator accepts it missing", op.Operation, name)
				case contains(op.Optional, name) && rejected && !(viewBriefMustChange[op.Operation] && len(optionalInExample) == 1):
					t.Errorf("%s: brief marks %s optional, validator rejects it missing", op.Operation, name)
				}
			}
		}
	}
}

func TestViewContractReportsRealBuildErrorsNotUnknownFlow(t *testing.T) {
	var out bytes.Buffer
	err := runViewTimeline([]string{"contract", "--for", "nope"}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "unknown flow") || !strings.Contains(err.Error(), "local") {
		t.Fatalf("unknown flow error: %v", err)
	}
	out.Reset()
	if err := runViewTimeline([]string{"contract", "--for", "local"}, &out, nil); err != nil || !strings.Contains(out.String(), `"session_delete"`) {
		t.Fatalf("local flow: %v", err)
	}
}
