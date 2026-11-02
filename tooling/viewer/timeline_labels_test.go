package viewer

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestTimelineReservedLabelMetadataAndSchema(t *testing.T) {
	c, err := TimelineProtocolContract("stage_start")
	if err != nil {
		t.Fatal(err)
	}
	want := TimelineLocalizationMetadata{DefaultLanguage: "en", SupportedLanguages: []string{"en", "es"}, ReservedTitleKeys: []string{"start", "research", "plan", "build", "verify", "delivery", "tasks", "ci_wait"}}
	if !reflect.DeepEqual(c.Localization, want) {
		t.Fatal(c.Localization)
	}
	for _, typ := range []reflect.Type{reflect.TypeOf(TimelineAppendEventRequest{}), reflect.TypeOf(TimelineEvent{}), reflect.TypeOf(TimelineEventPatchRequest{})} {
		schema := timelineProtocolSchema(typ)
		props := schema["properties"].(map[string]any)
		if _, ok := props["title_key"]; ok {
			t.Fatal("duplicate key field")
		}
		title := props["title"].(map[string]any)
		if title["type"] != "string" || !strings.Contains(title["description"].(string), "do not repeat") {
			t.Fatal(title)
		}
		if typ != reflect.TypeOf(TimelineEventPatchRequest{}) && (schema["allOf"] == nil || !contains(schema["required"].([]string), "title")) {
			t.Fatal("missing conditional title rule")
		}
	}
	var example TimelineAppendEventRequest
	if err = json.Unmarshal(c.Commands[0].Example, &example); err != nil || example.Title != "" {
		t.Fatal("stage example repeats heading", err)
	}
}

func TestTimelineStageContextPersistenceReplayAndRejection(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	actor := TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}
	start := TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "start-context", Actor: actor, Reason: "Start contextual stage", Kind: "stage", Stage: "build", StageState: "in_progress", ExpectedSessionVersion: &v.Version, Title: "", Summary: "Scope", TicketIDs: []string{}}
	res, err := s.TimelineAppend(v.ViewerSessionID, start)
	if err != nil {
		t.Fatal(err)
	}
	context := "Native context / contexto"
	patch := TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "set-context", Actor: actor, ExpectedVersion: res.EventVersion, Title: &context}
	updated, err := s.TimelinePatchEvent(v.ViewerSessionID, res.EventID, patch)
	if err != nil {
		t.Fatal(err)
	}
	before := domainCurrent(t, s, v.ViewerSessionID)
	stale := patch
	stale.OperationID = "stale-context"
	if _, err = s.TimelinePatchEvent(v.ViewerSessionID, res.EventID, stale); err == nil {
		t.Fatal("stale patch accepted")
	}
	if !reflect.DeepEqual(before, domainCurrent(t, s, v.ViewerSessionID)) {
		t.Fatal("stale patch changed work")
	}
	summary := "Updated summary"
	omitted := TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "preserve-context", Actor: actor, ExpectedVersion: updated.EventVersion, Summary: &summary}
	preserved, err := s.TimelinePatchEvent(v.ViewerSessionID, res.EventID, omitted)
	if err != nil {
		t.Fatal(err)
	}
	if domainCurrent(t, s, v.ViewerSessionID).Events[0].Title != context {
		t.Fatal("omission cleared context")
	}
	empty := ""
	clear := TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "clear-context", Actor: actor, ExpectedVersion: preserved.EventVersion, Title: &empty}
	cleared, err := s.TimelinePatchEvent(v.ViewerSessionID, res.EventID, clear)
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if err = restarted.Startup(); err != nil {
		t.Fatal(err)
	}
	replay, err := restarted.TimelinePatchEvent(v.ViewerSessionID, res.EventID, clear)
	if err != nil || !reflect.DeepEqual(cleared, replay) {
		t.Fatal("replay changed receipt", err)
	}
	restored := domainCurrent(t, restarted, v.ViewerSessionID)
	if restored.Events[0].Title != "" {
		t.Fatal("clear did not persist")
	}
	req := start
	req.OperationID = "custom-event"
	req.Kind = "request"
	req.Stage = ""
	req.StageState = ""
	req.ExpectedSessionVersion = nil
	req.Title = "Custom native title"
	custom, err := restarted.TimelineAppend(v.ViewerSessionID, req)
	if err != nil {
		t.Fatal(err)
	}
	before = domainCurrent(t, restarted, v.ViewerSessionID)
	rejected := clear
	rejected.OperationID = "reject-empty-custom"
	rejected.ExpectedVersion = custom.EventVersion
	if _, err = restarted.TimelinePatchEvent(v.ViewerSessionID, custom.EventID, rejected); err == nil {
		t.Fatal("empty nonstage patch accepted")
	}
	req.OperationID = "reject-empty-create"
	req.Title = ""
	if _, err = restarted.TimelineAppend(v.ViewerSessionID, req); err == nil {
		t.Fatal("empty nonstage create accepted")
	}
	if !reflect.DeepEqual(before, domainCurrent(t, restarted, v.ViewerSessionID)) {
		t.Fatal("rejection changed work")
	}
	t.Logf("created empty stage; set/preserve/clear context; stale rejected unchanged; restart receipt replay identical; nonstage create/patch rejected unchanged; stage=%s version=%d", res.EventID, restored.Version)
}
