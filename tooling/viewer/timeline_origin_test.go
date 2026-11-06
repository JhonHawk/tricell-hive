package viewer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// These tests pin the declared origin of the local_viewer actor (browser or
// cli). They read stored state back as raw JSON so they describe the persisted
// shape, not the Go type.

type originEnv struct {
	t      *testing.T
	store  *Store
	server *Server
	id     string
}

func newOriginEnv(t *testing.T) *originEnv {
	t.Helper()
	store, v, _ := newV3Fixture(t)
	server, err := NewServer(ServerOptions{Store: store, Port: 4483, Token: strings.Repeat("z", 32)})
	if err != nil {
		t.Fatal(err)
	}
	return &originEnv{t: t, store: store, server: server, id: v.ViewerSessionID}
}

func (e *originEnv) call(method, suffix, body string) *httptest.ResponseRecorder {
	e.t.Helper()
	r := httptest.NewRequest(method, "http://127.0.0.1:4483/api/timeline/sessions/"+e.id+suffix, bytes.NewReader([]byte(body)))
	r.Header.Set(tokenHeader, strings.Repeat("z", 32))
	r.Header.Set("Origin", "http://127.0.0.1:4483")
	r.Header.Set("X-Hive-Action-Scope", "session")
	r.Header.Set("X-Hive-Timeline-Revision", "3")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	e.server.ServeHTTP(w, r)
	return w
}

func (e *originEnv) version() int64 {
	e.t.Helper()
	v, err := e.store.TimelineGet(e.id)
	if err != nil {
		e.t.Fatal(err)
	}
	return v.Version
}

// actor builds the JSON of a local_viewer actor; origin "" omits the field.
func originActor(origin string) string {
	if origin == "" {
		return `{"kind":"local_viewer"}`
	}
	return `{"kind":"local_viewer","origin":"` + origin + `"}`
}

func (e *originEnv) patch(op, origin string, fields string) *httptest.ResponseRecorder {
	e.t.Helper()
	body := `{"schema_version":3,"operation_id":"` + op + `","expected_version":` + jsonInt(e.version()) + `,"actor":` + originActor(origin) + `,` + fields + `}`
	return e.call(http.MethodPatch, "", body)
}

func jsonInt(n int64) string { b, _ := json.Marshal(n); return string(b) }

// rawRecord reads the persisted record file as generic JSON.
func (e *originEnv) rawRecord() map[string]any {
	e.t.Helper()
	raw, err := os.ReadFile(e.store.timelineRecordPath(e.id))
	if err != nil {
		e.t.Fatal(err)
	}
	var out map[string]any
	if err = json.Unmarshal(raw, &out); err != nil {
		e.t.Fatal(err)
	}
	return out
}

func originOf(t *testing.T, actor any) string {
	t.Helper()
	m, ok := actor.(map[string]any)
	if !ok {
		t.Fatalf("actor missing: %v", actor)
	}
	origin, _ := m["origin"].(string)
	return origin
}

// lastLocalEvent returns the actor of the newest event with the given audit operation.
func (e *originEnv) lastEventActor(operation string) any {
	e.t.Helper()
	session, _ := e.rawRecord()["session"].(map[string]any)
	events, _ := session["events"].([]any)
	for i := len(events) - 1; i >= 0; i-- {
		ev := events[i].(map[string]any)
		if audit, ok := ev["audit"].(map[string]any); ok && audit["operation"] == operation {
			return ev["actor"]
		}
	}
	e.t.Fatalf("no %s event", operation)
	return nil
}

func (e *originEnv) receiptActor(slot string) any {
	e.t.Helper()
	m, _ := e.rawRecord()["maintenance"].(map[string]any)
	receipt, _ := m[slot].(map[string]any)
	if receipt == nil {
		e.t.Fatalf("no %s", slot)
	}
	return receipt["actor"]
}

func TestTimelineRenameKeepsDeclaredOriginInPublicEvent(t *testing.T) {
	e := newOriginEnv(t)
	if w := e.patch("rename-cli", "cli", `"title":"From the CLI"`); w.Code != 200 {
		t.Fatalf("origin cli rename = %d %s", w.Code, w.Body.String())
	}
	if got := originOf(t, e.lastEventActor("session_rename")); got != "cli" {
		t.Fatalf("rename event origin = %q, want cli", got)
	}
	if w := e.patch("rename-default", "", `"title":"From the browser"`); w.Code != 200 {
		t.Fatalf("rename without origin = %d", w.Code)
	}
	if got := originOf(t, e.lastEventActor("session_rename")); got != "" {
		t.Fatalf("rename event without origin = %q, want it stored absent", got)
	}
	if w := e.patch("rename-browser", "browser", `"title":"Explicit browser"`); w.Code != 200 {
		t.Fatalf("rename origin browser = %d", w.Code)
	}
	if got := originOf(t, e.lastEventActor("session_rename")); got != "" {
		t.Fatalf("rename event with origin browser = %q, want it stored absent", got)
	}
}

func TestTimelineAdministrativeReceiptsKeepDeclaredOrigin(t *testing.T) {
	e := newOriginEnv(t)
	cases := []struct {
		op, origin, fields, slot, want string
	}{
		{"archive-cli", "cli", `"archived":true`, "archive_receipt", "cli"},
		{"restore-default", "", `"archived":false`, "restore_receipt", ""},
		{"pin-cli", "cli", `"pinned":true`, "pin_receipt", "cli"},
	}
	for _, c := range cases {
		if w := e.patch(c.op, c.origin, c.fields); w.Code != 200 {
			t.Fatalf("%s = %d %s", c.op, w.Code, w.Body.String())
		}
		if got := originOf(t, e.receiptActor(c.slot)); got != c.want {
			t.Fatalf("%s origin = %q, want %s", c.slot, got, c.want)
		}
	}
}

func TestTimelineUnknownOriginIsRejectedWithoutChange(t *testing.T) {
	e := newOriginEnv(t)
	before := e.version()
	for _, origin := range []string{"robot", "CLI", "agent", " cli"} {
		if w := e.patch("bad-"+strings.TrimSpace(origin), origin, `"title":"Nope"`); w.Code < 400 || w.Code >= 500 {
			t.Fatalf("origin %q = %d, want a 4xx rejection", origin, w.Code)
		}
	}
	if w := e.patch("empty-origin", "", `"title":"Nope"`); w.Code != 200 {
		t.Fatalf("control request = %d", w.Code)
	}
	if e.version() != before+1 {
		t.Fatalf("version %d, want %d: a rejected origin changed the work", e.version(), before+1)
	}
	if actor := (TimelineActor{Kind: "native_producer", ProducerID: "producer-1"}); actor.Validate() != nil {
		t.Fatal("control native actor must validate")
	}
	var native TimelineActor
	if err := json.Unmarshal([]byte(`{"kind":"native_producer","producer_id":"producer-1","origin":"cli"}`), &native); err == nil && native.Validate() == nil {
		t.Fatal("native_producer must not carry an origin")
	}
}

func TestTimelineRecoverLeadKeepsDeclaredOriginInPublicEvent(t *testing.T) {
	e := newOriginEnv(t)
	pending := domainJoin(t, e.store, e.id, "next-lead")
	body := `{"schema_version":3,"operation_id":"recover-cli","actor":` + originActor("cli") + `,"expected_session_version":` + jsonInt(e.version()) + `,"reason":"Lead session ended","request_id":"` + pending + `"}`
	if w := e.call(http.MethodPost, "/recover-lead", body); w.Code != 200 {
		t.Fatalf("recover with origin cli = %d %s", w.Code, w.Body.String())
	}
	if got := originOf(t, e.lastEventActor("recover_lead")); got != "cli" {
		t.Fatalf("recover event origin = %q, want cli", got)
	}
}

func TestTimelineDeletionAuditKeepsOriginAfterDeletion(t *testing.T) {
	e := newOriginEnv(t)
	body := `{"schema_version":3,"operation_id":"delete-cli","expected_version":` + jsonInt(e.version()) + `,"actor":` + originActor("cli") + `}`
	if w := e.call(http.MethodDelete, "", body); w.Code != 200 {
		t.Fatalf("delete with origin cli = %d %s", w.Code, w.Body.String())
	}
	record := e.rawRecord()
	if record["session"] != nil {
		t.Fatal("deleted record still holds the public session")
	}
	audits, _ := record["audit"].([]any)
	for _, a := range audits {
		audit := a.(map[string]any)
		if audit["operation"] == "session_delete" {
			if got := originOf(t, audit["actor"]); got != "cli" {
				t.Fatalf("deletion audit origin = %q, want cli", got)
			}
			return
		}
	}
	t.Fatal("deletion audit did not survive deletion")
}

func TestTimelineDeletionWithoutOriginIsRecordedAsBrowser(t *testing.T) {
	e := newOriginEnv(t)
	body := `{"schema_version":3,"operation_id":"delete-default","expected_version":` + jsonInt(e.version()) + `,"actor":` + originActor("") + `}`
	if w := e.call(http.MethodDelete, "", body); w.Code != 200 {
		t.Fatalf("delete = %d", w.Code)
	}
	audits, _ := e.rawRecord()["audit"].([]any)
	for _, a := range audits {
		audit := a.(map[string]any)
		if audit["operation"] == "session_delete" {
			if got := originOf(t, audit["actor"]); got != "" {
				t.Fatalf("deletion audit origin = %q, want it stored absent", got)
			}
			return
		}
	}
	t.Fatal("no deletion audit")
}

func TestOriginDoesNotChangeAuthorityOrRevision(t *testing.T) {
	if TimelineContractRevision != 3 {
		t.Fatalf("contract revision = %d, want 3", TimelineContractRevision)
	}
	e := newOriginEnv(t)
	// A local_viewer still cannot publish content, whatever its origin.
	body := `{"schema_version":3,"operation_id":"forbidden","expected_version":` + jsonInt(e.version()) + `,"actor":` + originActor("cli") + `,"summary":"Not allowed"}`
	if w := e.call(http.MethodPatch, "", body); w.Code < 400 {
		t.Fatalf("local_viewer summary patch = %d, want a rejection", w.Code)
	}
	// A native producer cannot use the local-only actions.
	producer := domainCurrent(t, e.store, e.id).ForegroundProducerID
	body = `{"schema_version":3,"operation_id":"native-rename","expected_version":` + jsonInt(e.version()) + `,"actor":{"kind":"native_producer","producer_id":"` + producer + `"},"title":"Native"}`
	if w := e.call(http.MethodPatch, "", body); w.Code < 400 {
		t.Fatalf("native rename = %d, want a rejection", w.Code)
	}
}

func TestHealthAnnouncesLocalActionsCLICapability(t *testing.T) {
	s := newSrvEnv(t)
	w := s.do(http.MethodGet, "/api/health", reqOpts{token: testToken})
	if w.Code != http.StatusOK {
		t.Fatalf("health = %d", w.Code)
	}
	var h struct {
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &h); err != nil {
		t.Fatal(err)
	}
	if len(h.Capabilities) != 2 || h.Capabilities[0] != "local_actions_cli" || h.Capabilities[1] != "host_events" {
		t.Fatalf("capabilities = %v, want [local_actions_cli host_events]", h.Capabilities)
	}
}

func TestPublishedSchemaAcceptsOriginAndUsageNoLongerLocalOnly(t *testing.T) {
	full, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range full.Commands {
		if (c.Operation == "session_delete" || c.Operation == "recover_lead") && strings.Contains(c.Usage, "Local viewer only") {
			t.Fatalf("%s usage still says Local viewer only: %q", c.Operation, c.Usage)
		}
		if c.Operation == "session_delete" && !strings.Contains(c.Usage, "hive view work delete") {
			t.Fatalf("session_delete usage does not offer the CLI: %q", c.Usage)
		}
		if c.Operation == "recover_lead" && !strings.Contains(c.Usage, "hive view work recover") {
			t.Fatalf("recover_lead usage does not offer the CLI: %q", c.Usage)
		}
	}
	del, err := TimelineProtocolContract("session_delete")
	if err != nil || len(del.Commands) != 1 {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(del.Commands[0].InputSchema)
	if !strings.Contains(string(raw), `"origin"`) || !strings.Contains(string(raw), `"browser"`) || !strings.Contains(string(raw), `"cli"`) {
		t.Fatalf("session_delete input schema does not accept origin: %s", raw)
	}
}

func TestMaintenanceBudgetHoldsWithOriginOnEveryReceipt(t *testing.T) {
	e := newOriginEnv(t)
	for _, c := range []struct{ op, fields string }{
		{"archive", `"archived":true`}, {"restore", `"archived":false`}, {"pin", `"pinned":true`},
	} {
		if w := e.patch(c.op, "browser", c.fields); w.Code != 200 {
			t.Fatalf("%s = %d %s", c.op, w.Code, w.Body.String())
		}
	}
	m, _ := json.Marshal(e.rawRecord()["maintenance"])
	// The three receipts are the only growth origin adds; a hundred bytes of
	// headroom under the cap leaves the policy's own metadata room.
	if len(m)+len(`,"maintenance":`) > TimelineMaintenanceMaxBytes/2 {
		t.Fatalf("maintenance metadata is %d bytes, over half of the %d-byte cap", len(m), TimelineMaintenanceMaxBytes)
	}
}
