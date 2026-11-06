package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"tricell-hive/tooling/viewer"
)

func TestTimelineCLIContractIsOfflineAndNarrowed(t *testing.T) {
	var out bytes.Buffer
	if err := runView([]string{"contract"}, &out, viewDeps{}); err != nil {
		t.Fatal(err)
	}
	var full viewer.TimelineProtocolDescription
	if err := json.Unmarshal(out.Bytes(), &full); err != nil || len(full.Commands) != 28 || full.ProtocolVersion != 3 || full.Retention != "manual" {
		t.Fatalf("invalid full contract: %v", err)
	}
	for _, operation := range full.Commands {
		out.Reset()
		if err := runViewTimeline([]string{"contract", "--operation", operation.Operation}, &out, nil); err != nil {
			t.Fatal(err)
		}
		var narrowed viewer.TimelineProtocolDescription
		if err := json.Unmarshal(out.Bytes(), &narrowed); err != nil || len(narrowed.Commands) != 1 || narrowed.Commands[0].Operation != operation.Operation {
			t.Fatalf("invalid narrowed contract for %s", operation.Operation)
		}
	}
	out.Reset()
	if err := runViewTimeline([]string{"contract", "--operation", "unknown"}, &out, nil); err == nil || out.Len() != 0 {
		t.Fatal("unknown operation produced output")
	}
}

func TestTimelineCLIParsesOnlyScopedCommandPaths(t *testing.T) {
	sid := strings.Repeat("a", 64)
	cases := []struct {
		args                    []string
		operation, path, method string
	}{
		{[]string{"session", "create"}, "create_session", "/api/timeline/sessions", "POST"},
		{[]string{"session", "get", sid}, "session_get", "/api/timeline/sessions/" + sid, "GET"},
		{[]string{"stage", "start", sid}, "stage_start", "/api/timeline/sessions/" + sid + "/events", "POST"},
		{[]string{"stage", "snapshot", sid, "event_1"}, "document_snapshot", "/api/timeline/sessions/" + sid + "/events/event_1/document-snapshots", "POST"},
		{[]string{"document", "get", sid, "doc_1"}, "document_get", "/api/timeline/sessions/" + sid + "/documents/doc_1", "GET"},
	}
	for _, tc := range cases {
		args := append(append([]string{}, tc.args...), "--state-dir", "/fixture-state")
		got, rest, err := parseViewTimelineCommand(args)
		if err != nil || got.operation != tc.operation || got.method != tc.method || got.path != tc.path || len(rest) != 2 {
			t.Fatalf("bad command %v: %+v %v", tc.args, got, err)
		}
	}
	for _, args := range [][]string{
		{"session", "get", "../outside"}, {"event", "update", sid, "../event"},
		{"document", "get", sid, "doc?token=value"}, {"stage", "start"},
		{"session", "unknown"}, {"artifact", "get", sid},
	} {
		if _, _, err := parseViewTimelineCommand(args); err == nil {
			t.Fatalf("accepted invalid command %v", args)
		}
	}
}

func timelineCLIFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "viewer", "testdata", "timeline_contract", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestTimelineCLIInvalidOrUnavailablePublicationPreservesInputs(t *testing.T) {
	state := t.TempDir()
	var out bytes.Buffer
	if err := runViewTimeline([]string{"session", "create", "--json", "-", "--state-dir", state}, &out, strings.NewReader(`{"schema_version":1,"schema_version":1}`)); err == nil || out.Len() != 0 {
		t.Fatal("invalid input was accepted")
	}
	input := filepath.Join(t.TempDir(), "request.json")
	raw := timelineCLIFixture(t, "create-session.json")
	if err := os.WriteFile(input, raw, 0600); err != nil {
		t.Fatal(err)
	}
	err := runViewTimeline([]string{"session", "create", "--json", input, "--state-dir", state}, &out, nil)
	if err == nil || !strings.Contains(err.Error(), "unavailable") || out.Len() != 0 {
		t.Fatalf("unexpected unavailable outcome: %v", err)
	}
	after, readErr := os.ReadFile(input)
	if readErr != nil || !bytes.Equal(raw, after) {
		t.Fatal("publication changed input")
	}
	entries, err := os.ReadDir(state)
	if err != nil || len(entries) != 0 {
		t.Fatal("offline publication created server state")
	}
	link := filepath.Join(t.TempDir(), "link.json")
	if err := os.Symlink(input, link); err != nil {
		t.Fatal(err)
	}
	if _, err := viewTimelineReadFile(link, viewer.TimelineMaxBody); err == nil {
		t.Fatal("followed input symlink")
	}
	if _, err := viewTimelineReadInput("-", strings.NewReader(strings.Repeat("x", 12)), 10); err == nil {
		t.Fatal("accepted oversized stdin")
	}
}

func timelineCLIInfo(t *testing.T, server *httptest.Server) serverInfo {
	t.Helper()
	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		t.Fatal(err)
	}
	n, err := strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}
	return serverInfo{PID: 123, Port: n, Token: "synthetic-timeline-test-access", StateDir: "/fixture-state"}
}

func TestTimelineCLITransportGuardsAndProtocolIdentity(t *testing.T) {
	var revision atomic.Bool
	revision.Store(true)
	var mutations atomic.Int32
	var protocolVersion atomic.Int32
	protocolVersion.Store(3)
	var info serverInfo
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if revision.Load() {
			w.Header().Set("X-Hive-Timeline-Revision", "3")
		}
		if r.Header.Get("X-Hive-Token") != info.Token {
			t.Error("missing access header")
		}
		if r.URL.Path == "/api/health" {
			fmt.Fprintf(w, `{"pid":123,"state_dir":"/fixture-state","timeline_protocol_version":%d}`, protocolVersion.Load())
			return
		}
		if r.Method == "PATCH" {
			mutations.Add(1)
		}
		if r.Method == "PATCH" && (r.Header.Get("Origin") != info.url() || r.Header.Get("X-Hive-Action-Scope") != "session") {
			t.Error("missing mutation guards")
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	info = timelineCLIInfo(t, server)
	if _, err := viewTimelineReadHealth(info, "/fixture-state"); err != nil {
		t.Fatal(err)
	}
	protocolVersion.Store(1)
	if _, err := viewTimelineReadHealth(info, "/fixture-state"); err == nil {
		t.Fatal("accepted obsolete v1 server")
	}
	protocolVersion.Store(3)
	if _, err := viewTimelineReadHealth(info, "/different-state"); err == nil {
		t.Fatal("accepted wrong server identity")
	}
	if _, err := viewTimelineHTTP(info, "PATCH", "/api/timeline/sessions/test", "application/json", strings.NewReader(`{}`), 1024); err != nil {
		t.Fatal(err)
	}
	revision.Store(false)
	before := mutations.Load()
	if _, err := viewTimelineHTTP(info, "PATCH", "/api/timeline/sessions/test", "application/json", strings.NewReader(`{"operation_id":"original"}`), 1024); err == nil || !strings.Contains(err.Error(), "reload_required") || mutations.Load() != before {
		t.Fatal("old server received a write", err)
	}
	if _, code, err := viewWorkHTTP(info, "PATCH", "/api/timeline/sessions/test", "application/json", []byte(`{"operation_id":"original"}`)); err == nil || code != "reload_required" || mutations.Load() != before {
		t.Fatal("publisher wrote to old server", code, err)
	}
	revision.Store(true)
	if _, err := viewTimelineHTTP(info, "GET", "/api/test?token=forbidden", "", nil, 1024); err == nil {
		t.Fatal("accepted credential query")
	}
	if _, err := viewTimelineHTTP(info, "GET", "/api/test", "", nil, 2); err == nil {
		t.Fatal("accepted oversized response")
	}
}

func TestTimelineCLIRedirectAndErrorNeverExposeCredentials(t *testing.T) {
	reached := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached++; w.Write([]byte(`{}`)) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Hive-Timeline-Revision", "3")
		if r.URL.Path == "/api/redirect" {
			http.Redirect(w, r, target.URL, http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"token":"synthetic-private-response","code":"unsafe-echo"}`))
	}))
	defer server.Close()
	info := timelineCLIInfo(t, server)
	for _, path := range []string{"/api/redirect", "/api/error"} {
		out, err := viewTimelineHTTP(info, "GET", path, "", nil, 1024)
		if err == nil || len(out) != 0 || strings.Contains(err.Error(), info.Token) || strings.Contains(err.Error(), "synthetic-private-response") {
			t.Fatalf("unsafe error: %v", err)
		}
	}
	if reached != 0 {
		t.Fatal("redirect forwarded access to another listener")
	}
}

func TestTimelineCLIPublicReadbackRejectsForeignAndPrivateResults(t *testing.T) {
	raw := timelineCLIFixture(t, "operation-result.json")
	result, err := viewer.DecodeTimelineOperationResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	input, _ := json.Marshal(map[string]any{"operation_id": result.OperationID})
	command := viewTimelineCommand{operation: string(result.Kind), sessionID: result.ViewerSessionID}
	var out bytes.Buffer
	if err := viewTimelineDecodeOutput(&out, command, input, raw); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	command.sessionID = strings.Repeat("f", 64)
	if err := viewTimelineDecodeOutput(&out, command, input, raw); err == nil || out.Len() != 0 {
		t.Fatal("accepted foreign session result")
	}
	if err := viewTimelineDecodeOutput(&out, command, input, []byte(`{"token":"synthetic-private-response"}`)); err == nil || out.Len() != 0 {
		t.Fatal("printed private response")
	}
	doc := timelineCLIFixture(t, "document-response-bom.json")
	response, err := viewer.DecodeTimelineDocumentResponse(doc)
	if err != nil {
		t.Fatal(err)
	}
	command = viewTimelineCommand{operation: "document_get", sessionID: response.Document.SessionID, objectID: response.Document.DocumentID}
	if err := viewTimelineDecodeOutput(&out, command, nil, doc); err != nil {
		t.Fatal(err)
	}
}

// The response owner exercises both authenticated reading and strict decoding.
func TestTimelineCLIPublicReadbackAtLegacyContentLimit(t *testing.T) {
	session, err := viewer.DecodeTimelineSession(timelineCLIFixture(t, "session.json"))
	if err != nil {
		t.Fatal(err)
	}
	template := session.Events[0]
	session.Events = nil
	session.ValidationCount = 0
	for n := 0; n < 460; n++ {
		event := template
		event.EventID = fmt.Sprintf("boundary-%d", n)
		event.Sequence = int64(n + 1)
		event.Version = 1
		event.Title = strings.Repeat("t", 200)
		event.Summary = strings.Repeat("s", 3000)
		session.Events = append(session.Events, event)
	}
	sessionRaw, _ := json.Marshal(session)
	var legacy map[string]json.RawMessage
	json.Unmarshal(sessionRaw, &legacy)
	for _, field := range []string{"last_activity_at", "closed_at", "autoarchive_at", "archive_origin"} {
		delete(legacy, field)
	}
	legacyRaw, _ := json.Marshal(legacy)
	remaining := viewer.TimelineMaxBody - len(legacyRaw)
	if remaining < 0 {
		t.Fatal("fixture already exceeds original content budget")
	}
	for n := range session.Events {
		add := min(1000, remaining)
		session.Events[n].Summary += strings.Repeat("s", add)
		remaining -= add
	}
	if remaining != 0 {
		t.Fatal("fixture cannot reach original content limit", remaining)
	}
	if err = viewer.ValidateTimelineSession(session); err != nil {
		t.Fatal("invalid boundary fixture", err)
	}
	response := viewer.TimelineSessionResponse{SchemaVersion: 3, Mode: "timeline_api", Session: session, ArtifactURLs: map[string]string{}, DocumentURLs: map[string]string{}}
	raw, _ := json.Marshal(response)
	if len(raw) <= viewer.TimelineMaxBody || len(raw) > viewer.TimelineMaxBody+(16<<10) {
		t.Fatal("fixture missed response boundary", len(raw))
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(viewer.TimelineRevisionHeader, viewer.TimelineRevisionValue)
		w.Write(raw)
	}))
	defer server.Close()
	info := timelineCLIInfo(t, server)
	data, err := viewTimelineHTTP(info, "GET", "/api/timeline/sessions/"+session.ViewerSessionID, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	workData, _, workErr := viewWorkHTTP(info, "GET", "/api/timeline/sessions/"+session.ViewerSessionID, "", nil)
	if workErr != nil || !bytes.Equal(workData, data) {
		t.Fatal("publisher rejected bounded session response", workErr)
	}
	var out bytes.Buffer
	if err = viewTimelineDecodeOutput(&out, viewTimelineCommand{operation: "session_get", sessionID: session.ViewerSessionID}, nil, data); err != nil {
		t.Fatal("valid legacy content plus response projection rejected", err)
	}
	oversizedResponse := append(append([]byte{}, raw...), bytes.Repeat([]byte(" "), viewer.TimelineMaxResponseBody-len(raw)+1)...)
	if _, err = viewer.DecodeTimelineSessionResponse(oversizedResponse); err == nil {
		t.Fatal("response exceeded fixed headroom")
	}
	raw = oversizedResponse
	if _, err = viewTimelineHTTP(info, "GET", "/api/timeline/sessions/"+session.ViewerSessionID, "", nil, viewer.TimelineMaxBody); err == nil {
		t.Fatal("low-level transport accepted oversized response")
	}
	if _, _, err = viewWorkHTTP(info, "GET", "/api/timeline/sessions/"+session.ViewerSessionID, "", nil); err == nil {
		t.Fatal("publisher transport accepted oversized response")
	}
	// The larger response budget must never change request admission.
	oversized := append(timelineCLIFixture(t, "create-session.json"), bytes.Repeat([]byte(" "), viewer.TimelineMaxBody)...)
	if _, err = viewer.DecodeTimelineCreateSession(oversized); err == nil {
		t.Fatal("oversized request accepted")
	}
}

func TestTimelineCLIGitReaderDistinguishesMissingOriginFromFailure(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	read := viewGitRemote(newGitRunner())
	if remote, err := read(repo); err != nil || remote != "" {
		t.Fatal("missing origin was not a verified empty identity")
	}
	if err := os.WriteFile(filepath.Join(repo, ".git", "config"), []byte("[invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := read(repo); err == nil {
		t.Fatal("corrupt Git metadata became a path fallback")
	}
}

func TestTimelineCLIRealServerDocumentsStagesReplayAndTarget(t *testing.T) {
	f := newViewFixture(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	input := filepath.Join(t.TempDir(), "request.json")
	invoke := func(args []string, body any) string {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
		return f.mustView(append(args, "--json", input)...)
	}
	create := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "cli-create", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "fixture-native"}, Title: "CLI fixture", Summary: "Isolated integration", Tickets: []viewer.TimelineTicketInput{}}
	raw := invoke([]string{"session", "create"}, create)
	session, err := viewer.DecodeTimelineSessionResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	sid := session.Session.ViewerSessionID
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: session.Session.ForegroundProducerID}
	var status struct {
		TimelineSessions int    `json:"timeline_sessions"`
		Retention        string `json:"retention"`
	}
	if err := json.Unmarshal([]byte(f.mustView("status")), &status); err != nil || status.TimelineSessions != 1 || status.Retention != "manual" {
		t.Fatal("status omitted API sessions or manual retention")
	}
	if again := invoke([]string{"session", "create"}, create); again != raw {
		t.Fatal("create replay changed public result")
	}
	version := session.Session.Version
	stageRaw := invoke([]string{"stage", "start", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit fixture operation", SchemaVersion: 3, OperationID: "cli-plan", Kind: "stage", Title: "Plan", Summary: "Document fixture", TicketIDs: []string{}, Stage: "plan", StageState: "in_progress", ExpectedSessionVersion: &version})
	stage, err := viewer.DecodeTimelineOperationResult([]byte(stageRaw))
	if err != nil {
		t.Fatal(err)
	}
	markdown := filepath.Join(t.TempDir(), "plan.md")
	content := "\ufeff# Fixture plan\n\nImmutable Markdown.\n"
	if err := os.WriteFile(markdown, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	docRaw := invoke([]string{"document", "add", sid, "--file", markdown}, viewer.TimelineDocumentMetadata{Actor: actor, SchemaVersion: 3, OperationID: "cli-doc", MediaType: "text/markdown"})
	doc, err := viewer.DecodeTimelineOperationResult([]byte(docRaw))
	if err != nil {
		t.Fatal(err)
	}
	snapshotRaw := invoke([]string{"stage", "snapshot", sid, stage.EventID}, viewer.TimelineDocumentSnapshotRequest{Actor: actor, DocumentKey: "cli-plan", ExpectedSessionVersion: doc.SessionVersion, SchemaVersion: 3, OperationID: "cli-snapshot", ExpectedVersion: stage.EventVersion, Status: "complete", Tabs: []viewer.TimelineDocumentTab{{Label: "Plan", DocumentID: doc.DocumentID}}})
	snapshot, err := viewer.DecodeTimelineOperationResult([]byte(snapshotRaw))
	if err != nil {
		t.Fatal(err)
	}
	documentRaw := f.mustView("document", "get", sid, doc.DocumentID)
	document, err := viewer.DecodeTimelineDocumentResponse([]byte(documentRaw))
	if err != nil || document.Content != content {
		t.Fatal("document bytes did not round-trip")
	}
	state := viewer.TimelineStageState("completed")
	current := snapshot.SessionVersion
	complete := viewer.TimelineEventPatchRequest{Actor: actor, Reason: "Explicit fixture operation", SchemaVersion: 3, OperationID: "cli-complete", ExpectedVersion: snapshot.EventVersion, StageState: &state, ExpectedSessionVersion: &current}
	completeRaw := invoke([]string{"stage", "update", sid, stage.EventID}, complete)
	if _, err := viewer.DecodeTimelineOperationResult([]byte(completeRaw)); err != nil {
		t.Fatal(err)
	}
	opened := f.mustView("open", "--session-id", sid, "--event-id", stage.EventID)
	if len(f.opened) != 1 || strings.Contains(opened, "#code=") {
		t.Fatal("opener output exposed access or opened the wrong number of views")
	}
	target, err := url.Parse(f.opened[0])
	if err != nil || target.Query().Get("session_id") != sid || target.Query().Get("event_id") != stage.EventID || !strings.HasPrefix(target.Fragment, "code=") {
		t.Fatal("target was not carried through one-use access")
	}
	if _, err := f.view("open", "--session-id", sid, "--event-id", "foreign-event"); err == nil || len(f.opened) != 1 {
		t.Fatal("invalid event silently opened another target")
	}
	f.mustView("stop")
	f.mustView("start")
	if again := invoke([]string{"stage", "update", sid, stage.EventID}, complete); again != completeRaw {
		t.Fatal("mutation replay after restart changed IDs/result")
	}
	readRaw := f.mustView("session", "get", sid)
	read, err := viewer.DecodeTimelineSessionResponse([]byte(readRaw))
	if err != nil || len(read.Session.Events) != 2 || len(read.Session.Documents) != 1 || read.Session.CurrentStageEventID != nil {
		t.Fatal("restart lost stage/documents")
	}
	deletion := viewer.TimelineMutationRequest{Actor: viewer.TimelineActor{Kind: "local_viewer"}, SchemaVersion: 3, OperationID: "cli-delete", ExpectedVersion: read.Session.Version}
	deleted := timelineLocalJSON(t, f, "DELETE", "/api/timeline/sessions/"+sid, deletion)
	f.mustView("stop")
	f.mustView("start")
	if again := timelineLocalJSON(t, f, "DELETE", "/api/timeline/sessions/"+sid, deletion); again != deleted {
		t.Fatal("deletion replay after restart failed")
	}
	if got, err := os.ReadFile(markdown); err != nil || string(got) != content {
		t.Fatal("CLI altered source Markdown")
	}
}

func TestTimelineCLICreateReadbackChecksTheWorkKey(t *testing.T) {
	request := func(key any) []byte {
		var body map[string]any
		if err := json.Unmarshal(timelineCLIFixture(t, "create-session.json"), &body); err != nil {
			t.Fatal(err)
		}
		if key != nil {
			body["work_key"] = key
		}
		raw, _ := json.Marshal(body)
		return raw
	}
	response := func(key any) []byte {
		var session map[string]any
		if err := json.Unmarshal(timelineCLIFixture(t, "session.json"), &session); err != nil {
			t.Fatal(err)
		}
		session["work_key"] = key
		raw, _ := json.Marshal(map[string]any{"schema_version": 3, "mode": "timeline_api", "session": session, "created": true})
		return raw
	}
	command := viewTimelineCommand{operation: "create_session"}
	for _, c := range []struct {
		name         string
		sent, echoed any
		ok           bool
	}{
		{"no key sent and null returned", nil, nil, true},
		{"key sent and echoed", "COM-1", "COM-1", true},
		{"key sent and null returned", "COM-1", nil, false},
		{"key sent and another returned", "COM-1", "COM-2", false},
		{"no key sent and a key returned", nil, "COM-1", false},
	} {
		var out bytes.Buffer
		err := viewTimelineDecodeOutput(&out, command, request(c.sent), response(c.echoed))
		if (err == nil) != c.ok || (!c.ok && out.Len() != 0) {
			t.Fatalf("%s: ok=%v err=%v", c.name, c.ok, err)
		}
	}
}

// A forbidden_for_kind the viewer returns for an event operation carries the
// kinds that accept the field, inside the request failure and before its
// recovery advice.
func TestTimelineCLIViewerRejectionNamesTheKindsThatAcceptAField(t *testing.T) {
	f := newViewFixture(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	input := filepath.Join(t.TempDir(), "request.json")
	write := func(body any) {
		t.Helper()
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(input, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "hint-create", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "fixture-native"}, Title: "Hint fixture", Summary: "Server rejection hint", Tickets: []viewer.TimelineTicketInput{}})
	session, err := viewer.DecodeTimelineSessionResponse([]byte(f.mustView("session", "create", "--json", input)))
	if err != nil {
		t.Fatal(err)
	}
	sid := session.Session.ViewerSessionID
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: session.Session.ForegroundProducerID}
	write(viewer.TimelineAppendEventRequest{Actor: actor, SchemaVersion: 3, OperationID: "hint-request", Kind: "request", Title: "Request", Summary: "Initial scope", TicketIDs: []string{}})
	added, err := viewer.DecodeTimelineOperationResult([]byte(f.mustView("event", "add", sid, "--json", input)))
	if err != nil {
		t.Fatal(err)
	}
	state := viewer.TimelineStageState("completed")
	write(viewer.TimelineEventPatchRequest{Actor: actor, SchemaVersion: 3, OperationID: "hint-patch", ExpectedVersion: 1, StageState: &state, ExpectedSessionVersion: &added.SessionVersion, Reason: "Close it"})
	_, err = f.view("event", "update", sid, added.EventID, "--json", input)
	if err == nil || !strings.Contains(err.Error(), "(HTTP 400: forbidden_for_kind on stage_state; only kind stage accepts it); preserve inputs") {
		t.Fatalf("got %v", err)
	}
}

func TestTimelineCLIViewerRejectionHintOnlyForEventOperations(t *testing.T) {
	for _, c := range []struct {
		operation string
		err       *viewHTTPStatusError
	}{
		{"session_patch", &viewHTTPStatusError{Status: 400, Detail: ": forbidden_for_kind on stage_state", Code: "forbidden_for_kind", Field: "stage_state"}},
		{"event_patch", &viewHTTPStatusError{Status: 409, Detail: ": version_conflict on stage_state", Code: "version_conflict", Field: "stage_state"}},
	} {
		if got, want := viewTimelineWithKindHint(c.operation, c.err).Error(), c.err.Error(); got != want {
			t.Errorf("%s %s: message changed to %q", c.operation, c.err.Code, got)
		}
	}
}
