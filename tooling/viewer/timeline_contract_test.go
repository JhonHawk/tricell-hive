package viewer

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type timelineFixtureCase struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Valid bool   `json:"valid"`
	JSON  string `json:"json"`
	// Code, when set, is the contract error code an invalid case must return.
	Code string `json:"code,omitempty"`
}

func timelineFixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/timeline_contract/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func timelineFixtureSession(t *testing.T) TimelineSession {
	t.Helper()
	s, err := DecodeTimelineSession(timelineFixture(t, "session.json"))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func timelineFixtureDecode(kind string, raw []byte) error {
	switch kind {
	case "create_session":
		_, err := DecodeTimelineCreateSession(raw)
		return err
	case "session_patch":
		_, err := DecodeTimelineSessionPatch(raw)
		return err
	case "event_patch":
		_, err := DecodeTimelineEventPatch(raw)
		return err
	case "mutation":
		_, err := DecodeTimelineMutation(raw)
		return err
	case "append_event":
		_, err := DecodeTimelineAppendEvent(raw)
		return err
	case "document_metadata":
		_, err := DecodeTimelineDocumentMetadata(raw)
		return err
	case "document_snapshot":
		_, err := DecodeTimelineDocumentSnapshot(raw)
		return err
	case "artifact_metadata":
		_, err := DecodeTimelineArtifactMetadata(raw)
		return err
	case "session":
		_, err := DecodeTimelineSession(raw)
		return err
	case "index_entry":
		_, err := DecodeTimelineIndexEntry(raw)
		return err
	case "operation_result":
		_, err := DecodeTimelineOperationResult(raw)
		return err
	case "operation_receipt":
		_, err := DecodeTimelineOperationReceipt(raw)
		return err
	default:
		return timelineError("fixture", "unknown_kind")
	}
}
func TestTimelineContractSharedFixtures(t *testing.T) {
	var cases []timelineFixtureCase
	if err := json.Unmarshal(timelineFixture(t, "cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			err := timelineFixtureDecode(c.Kind, []byte(c.JSON))
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v, got %v", c.Valid, err)
			}
			if c.Code != "" && TimelineContractCode(err) != c.Code {
				t.Fatalf("code=%s, got %v", c.Code, err)
			}
		})
	}
}
func TestTimelineContractSessionRoundTrip(t *testing.T) {
	s := timelineFixtureSession(t)
	if s.ValidationCount != 2 || len(s.Events) != 6 || s.State != TimelineActive {
		t.Fatalf("unexpected fixture shape: events=%d/count=%d/state=%s", len(s.Events), s.ValidationCount, s.State)
	}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeTimelineSession(raw); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("artifact_urls")) {
		t.Fatal("capability URLs leaked into session metadata")
	}
}
func TestTimelineContractEvidenceReferencesSurviveCopyEdit(t *testing.T) {
	s := timelineFixtureSession(t)
	evidenceID := s.Events[2].EvidenceID
	s.Events[2].Version++
	s.Events[2].Title = "Corrected first validation title"
	s.Events[2].UpdatedAt = "2026-10-31T19:00:00-06:00"
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := DecodeTimelineSession(raw)
	if err != nil {
		t.Fatal(err)
	}
	if restored.Events[2].EvidenceID != evidenceID || restored.Events[3].Views[0].Supersedes.EvidenceID != evidenceID || restored.Events[2].Sequence != 3 {
		t.Fatal("copy edit changed evidence identity or order")
	}
	s.Events[2].Withdrawn = true
	s.ValidationCount--
	if err := ValidateTimelineSession(s); err != nil {
		t.Fatalf("historical references must resolve after withdrawal: %v", err)
	}
}
func TestTimelineContractRejectsCorruptSession(t *testing.T) {
	cases := []struct {
		name   string
		change func(*TimelineSession)
	}{
		{"foreign artifact owner", func(s *TimelineSession) { s.Artifacts[0].SessionID = strings.Repeat("2", 64) }},
		{"unknown ticket", func(s *TimelineSession) {
			s.Events[2].Views[0].TicketRefs = []TimelineTicketRef{{TicketID: "unknown-ticket", Snapshot: TimelineTicketSnapshot{Key: "OTHER"}}}
		}},
		{"foreign image", func(s *TimelineSession) { s.Events[2].Views[0].ImageRefs[0].ArtifactID = "another_session_image" }},
		{"invalid evidence target", func(s *TimelineSession) { s.Events[3].Views[0].Supersedes.EvidenceID = "other_evidence" }},
		{"wrong view target", func(s *TimelineSession) { s.Events[3].Views[0].Supersedes.ViewID = "other_view" }},
		{"future event target", func(s *TimelineSession) { s.Events[3].Views[0].Supersedes.EventID = "event_approval" }},
		{"duplicate event", func(s *TimelineSession) { s.Events[3].EventID = s.Events[2].EventID }},
		{"duplicate evidence", func(s *TimelineSession) { s.Events[3].EvidenceID = s.Events[2].EvidenceID }},
		{"out of order", func(s *TimelineSession) { s.Events[3].Sequence = 2 }},
		{"human actor spoof", func(s *TimelineSession) { s.Events[2].Origin = "human" }},
		{"bad validation count", func(s *TimelineSession) { s.ValidationCount = 0 }},
		{"bad size", func(s *TimelineSession) { s.Artifacts[0].Size = 21 << 20 }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := timelineFixtureSession(t)
			c.change(&s)
			if err := ValidateTimelineSession(s); err == nil {
				t.Fatal("corrupt session accepted")
			}
		})
	}
}
func TestTimelineContractReferencesUseSessionSnapshot(t *testing.T) {
	s := timelineFixtureSession(t)
	r, err := DecodeTimelineAppendEvent(timelineFixture(t, "append-validation.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = ValidateTimelineEventReferences(s, r); err != nil {
		t.Fatal(err)
	}
	r.Views[0].ImageRefs[0].ArtifactID = "foreign_image"
	if err = ValidateTimelineEventReferences(s, r); err == nil {
		t.Fatal("foreign artifact reference accepted")
	}
	r.Views[0].ImageRefs[0].ArtifactID = "artifact_chat"
	r.Views[0].TicketIDs = []string{"unknown-ticket"}
	if err = ValidateTimelineEventReferences(s, r); err == nil {
		t.Fatal("equal key at another tracker was treated as same ticket")
	}
}
func TestTimelineContractRejectsExplicitNullAndMissingOutputFields(t *testing.T) {
	raw := timelineFixture(t, "session.json")
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	delete(obj, "pinned")
	missing, _ := json.Marshal(obj)
	if _, err := DecodeTimelineSession(missing); err == nil {
		t.Fatal("missing false boolean accepted as a declared output")
	}
	if err := DecodeTimelineJSON([]byte(`{"a":{"x":1,"\u0078":2}}`), &map[string]any{}); err == nil {
		t.Fatal("nested duplicate accepted")
	}
	if err := DecodeTimelineJSON([]byte(`{"a":null}`), &map[string]any{}); err == nil {
		t.Fatal("null accepted")
	}
	if err := DecodeTimelineJSON([]byte(`{} {}`), &map[string]any{}); err == nil {
		t.Fatal("trailing object accepted")
	}
}
func TestTimelineContractLimitsAndRedactedErrors(t *testing.T) {
	raw := []byte(strings.Repeat(" ", TimelineMaxBody+1))
	if err := DecodeTimelineJSON(raw, &TimelineCreateSessionRequest{}); err == nil {
		t.Fatal("oversized body accepted")
	}
	_, err := DecodeTimelineCreateSession([]byte(`{"schema_version":3,"operation_id":"create","source":{"project_root":"/private-sensitive-root","host":"unknown-sensitive-host","native_session_id":"private-sensitive-id"},"title":"Test","summary":"Scope","tickets":[]}`))
	if err == nil || strings.Contains(err.Error(), "sensitive") {
		t.Fatalf("error includes supplied data: %v", err)
	}
	if TimelineContractCode(err) != "unsupported_host" {
		t.Fatal("bounded error code lost")
	}
	deep := []byte(strings.Repeat("[", 34) + "1" + strings.Repeat("]", 34))
	if err := DecodeTimelineJSON(deep, &[]any{}); err == nil {
		t.Fatal("unbounded nesting accepted")
	}
}

func TestTimelineContractCapabilityURLs(t *testing.T) {
	s := timelineFixtureSession(t)
	rawURL := "/files/" + s.ViewerSessionID + "/" + strings.Repeat("a", 64) + "/artifact_profile.png"
	for _, raw := range []string{rawURL} {
		if !IsSafeTimelineArtifactURL(raw, s.ViewerSessionID) {
			t.Fatal("safe capability URL rejected")
		}
	}
	for _, raw := range []string{strings.Replace(rawURL, "artifact_profile.png", "../outside", 1), rawURL + "?token=value", "https://example.com/image.png"} {
		if IsSafeTimelineArtifactURL(raw, s.ViewerSessionID) {
			t.Fatal("unsafe capability URL accepted")
		}
	}
	if IsSafeTimelineArtifactURL(rawURL, strings.Repeat("2", 64)) {
		t.Fatal("foreign session capability accepted")
	}
	response := TimelineSessionResponse{SchemaVersion: 3, Mode: "timeline_api", Session: s, ArtifactURLs: map[string]string{"artifact_profile": rawURL}}
	raw, _ := json.Marshal(response)
	if _, err := DecodeTimelineSessionResponse(raw); err != nil {
		t.Fatal(err)
	}
	response.ArtifactURLs = map[string]string{"foreign_image": rawURL}
	raw, _ = json.Marshal(response)
	if _, err := DecodeTimelineSessionResponse(raw); err == nil {
		t.Fatal("URL for unknown artifact accepted")
	}
}

func TestTimelineContractExactKeysBeforeDecode(t *testing.T) {
	r := TimelineCreateSessionRequest{Title: "preserved"}
	raw := timelineFixture(t, "create-session.json")
	aliased := append(append([]byte{}, bytes.TrimSpace(raw)[:len(bytes.TrimSpace(raw))-1]...), []byte(`,"Title":"overwritten"}`)...)
	if err := DecodeTimelineJSON(aliased, &r); err == nil {
		t.Fatal("case-insensitive alias accepted")
	}
	if r.Title != "preserved" {
		t.Fatal("invalid alias mutated target before rejection")
	}
}
func TestTimelineContractReceiptAndRetryContracts(t *testing.T) {
	var replay struct {
		Receipt       json.RawMessage `json:"receipt"`
		MatchingRetry struct {
			OperationID   string `json:"operation_id"`
			RequestSHA256 string `json:"request_sha256"`
		} `json:"matching_retry"`
		ConflictingRetry struct {
			OperationID   string `json:"operation_id"`
			RequestSHA256 string `json:"request_sha256"`
		} `json:"conflicting_retry"`
	}
	if err := json.Unmarshal(timelineFixture(t, "operation-replay.json"), &replay); err != nil {
		t.Fatal(err)
	}
	receipt, err := DecodeTimelineOperationReceipt(replay.Receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err = ValidateTimelineOperationRetry(receipt, replay.MatchingRetry.OperationID, replay.MatchingRetry.RequestSHA256); err != nil {
		t.Fatal(err)
	}
	if err = ValidateTimelineOperationRetry(receipt, replay.ConflictingRetry.OperationID, replay.ConflictingRetry.RequestSHA256); TimelineContractCode(err) != "operation_conflict" {
		t.Fatalf("changed retry digest accepted: %v", err)
	}
	if receipt.Result.EventID != "event_validation" || receipt.Result.EvidenceID != "evidence_initial" {
		t.Fatal("retry fixture lost the committed IDs")
	}
	raw, err := json.Marshal(receipt.Result)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("request_sha256")) || bytes.Contains(raw, []byte("completed_at")) {
		t.Fatal("private receipt leaked into public result")
	}
	sessionRaw, err := json.Marshal(timelineFixtureSession(t))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sessionRaw, []byte("request_sha256")) || bytes.Contains(sessionRaw, []byte("operation_receipts")) {
		t.Fatal("private receipts leaked into public session")
	}
}
func TestTimelineContractIndexRejectsLegacy(t *testing.T) {
	if _, err := DecodeTimelineIndexEntry(timelineFixture(t, "legacy-index-entry.json")); TimelineContractCode(err) != "unsupported_protocol" {
		t.Fatal("legacy entry accepted", err)
	}
	api, err := DecodeTimelineIndexEntry(timelineFixture(t, "api-index-entry.json"))
	if err != nil {
		t.Fatal(err)
	}
	current, ok := api.(TimelineAPIIndexEntry)
	if !ok || current.ValidationCount != 2 || current.Source.NativeSessionID != "native-session-alpha" {
		t.Fatal("API index variant did not preserve new identity")
	}
}

func TestTimelineContractStageTransitionsShared(t *testing.T) {
	var cases []struct {
		Name    string          `json:"name"`
		Action  string          `json:"action"`
		Session json.RawMessage `json:"session"`
		EventID string          `json:"event_id"`
		Request json.RawMessage `json:"request"`
		Valid   bool            `json:"valid"`
	}
	if err := json.Unmarshal(timelineFixture(t, "transition-cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			s, err := DecodeTimelineSession(c.Session)
			if err != nil {
				t.Fatal("bad scenario session", err)
			}
			switch c.Action {
			case "event_patch":
				r, e := DecodeTimelineEventPatch(c.Request)
				err = e
				if err == nil {
					err = ValidateTimelineEventPatch(s, c.EventID, r)
				}
			case "session_patch":
				r, e := DecodeTimelineSessionPatch(c.Request)
				err = e
				if err == nil {
					err = ValidateTimelineSessionPatch(s, r)
				}
			case "withdraw":
				r, e := DecodeTimelineMutation(c.Request)
				err = e
				if err == nil {
					err = ValidateTimelineEventWithdrawal(s, c.EventID, r)
				}
			case "append":
				r, e := DecodeTimelineAppendEvent(c.Request)
				err = e
				if err == nil {
					err = ValidateTimelineEventReferences(s, r)
				}
			case "snapshot":
				r, e := DecodeTimelineDocumentSnapshot(c.Request)
				err = e
				if err == nil {
					err = ValidateTimelineDocumentSnapshot(s, c.EventID, r)
				}
			default:
				t.Fatal("unknown scenario")
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v: %v", c.Valid, err)
			}
		})
	}
}
func TestTimelineContractDocumentBytesAndReadback(t *testing.T) {
	response, err := DecodeTimelineDocumentResponse(timelineFixture(t, "document-response.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateTimelineDocumentBytes(bytes.Repeat([]byte("x"), TimelineMaxDocumentBytes)); err != nil {
		t.Fatal(err)
	}
	for _, content := range [][]byte{bytes.Repeat([]byte("x"), TimelineMaxDocumentBytes+1), {0xff}} {
		if ValidateTimelineDocumentBytes(content) == nil {
			t.Fatal("invalid document accepted")
		}
	}
	if err := ValidateTimelineDocumentBytes(nil); err != nil {
		t.Fatal("empty UTF8 document must deduplicate", err)
	}
	response.Content += "modified"
	raw, _ := json.Marshal(response)
	if _, err := DecodeTimelineDocumentResponse(raw); err == nil {
		t.Fatal("modified bytes accepted")
	}
	s, err := DecodeTimelineSession(timelineFixture(t, "stage-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	meta := TimelineDocumentMetadata{Actor: TimelineActor{Kind: "native_producer", ProducerID: "producer-original"}, SchemaVersion: 3, OperationID: "doc", MediaType: "text/markdown"}
	if err := ValidateTimelineDocumentUpload(s, meta, []byte("new")); err != nil {
		t.Fatal(err)
	}
	s.State = TimelineClosed
	if ValidateTimelineDocumentUpload(s, meta, []byte("new")) == nil {
		t.Fatal("closed upload accepted")
	}
}
func TestTimelineContractDocumentURLsScopedToKnownMetadata(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "stage-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/timeline/sessions/" + s.ViewerSessionID + "/documents/doc-plan"
	response := TimelineSessionResponse{SchemaVersion: 3, Mode: "timeline_api", Session: s, DocumentURLs: map[string]string{"doc-plan": path}}
	raw, _ := json.Marshal(response)
	if _, err := DecodeTimelineSessionResponse(raw); err != nil {
		t.Fatal(err)
	}
	for _, url := range []string{path + "?token=private", "https://example.com" + path, "/api/timeline/sessions/" + strings.Repeat("a", 64) + "/documents/doc-plan", path + "/../other"} {
		response.DocumentURLs["doc-plan"] = url
		raw, _ := json.Marshal(response)
		if _, err := DecodeTimelineSessionResponse(raw); err == nil {
			t.Fatal("unsafe document URL accepted")
		}
	}
	response.DocumentURLs = map[string]string{"unknown": path}
	raw, _ = json.Marshal(response)
	if _, err := DecodeTimelineSessionResponse(raw); err == nil {
		t.Fatal("unknown document URL accepted")
	}
}

func TestTimelineContractDocumentResponsePreservesUTF8BOM(t *testing.T) {
	response, err := DecodeTimelineDocumentResponse(timelineFixture(t, "document-response-bom.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(response.Content, "\ufeff") {
		t.Fatal("BOM stripped from content")
	}
	if response.Document.Size != int64(len([]byte(response.Content))) || response.Document.Size != 32 {
		t.Fatal("BOM bytes omitted from size")
	}
	if response.Document.SHA256 != "d58e2f6d5d846c4960387b9afdac5ec7aba37d6785782cd7d6f9d057990eeb43" {
		t.Fatal("BOM bytes omitted from digest")
	}
}
