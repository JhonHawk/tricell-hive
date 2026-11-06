package viewer

import (
	"encoding/json"
	"strings"
	"testing"
)

// These tests pin the revision 3 contract for host-observed events: the closed
// wire shape of a host event, its binding to a session that is an accepted
// producer, the operation ID derived from the event, and the provenance value.

func hostEventCases(t *testing.T) []struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	Kind   string `json:"kind"`
	Valid  bool   `json:"valid"`
	JSON   string `json:"json"`
} {
	t.Helper()
	var cases []struct {
		Name   string `json:"name"`
		Action string `json:"action"`
		Kind   string `json:"kind"`
		Valid  bool   `json:"valid"`
		JSON   string `json:"json"`
	}
	if err := json.Unmarshal(timelineFixture(t, "host-event-cases.json"), &cases); err != nil {
		t.Fatal(err)
	}
	return cases
}

func TestTimelineContractRevisionIsThree(t *testing.T) {
	if TimelineContractRevision != 3 {
		t.Fatalf("contract revision = %d, want 3", TimelineContractRevision)
	}
	contract, err := TimelineProtocolContract("")
	if err != nil {
		t.Fatal(err)
	}
	if contract.ContractRevision != 3 || !contains(contract.Capabilities, "host_events") {
		t.Fatalf("protocol revision %d capabilities %v", contract.ContractRevision, contract.Capabilities)
	}
	joined := strings.Join(contract.Invariants, "\n")
	if !strings.Contains(joined, "X-Hive-Timeline-Revision: 3") || strings.Contains(joined, "X-Hive-Timeline-Revision: 2") {
		t.Fatal("revision invariant not updated")
	}
	if contract.Authority["host_event"] == "" {
		t.Fatal("host_event authority is not described")
	}
	if len(contract.Commands) != 28 {
		t.Fatalf("host events must not add or change an agent operation: %d commands", len(contract.Commands))
	}
}

func TestTimelineHostEventSharedFixtures(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "v3-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range hostEventCases(t) {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Action {
			case "shape":
				_, err = DecodeTimelineHostEvent([]byte(c.JSON))
			case "authority":
				var a TimelineActor
				err = DecodeTimelineJSON([]byte(c.JSON), &a)
				if err == nil {
					err = ValidateTimelineAuthority(s, c.Kind, a, "")
				}
			default:
				t.Fatal("unknown fixture action")
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v got %v", c.Valid, err)
			}
		})
	}
}

func hostEventSession(t *testing.T) TimelineSession {
	t.Helper()
	s, err := DecodeTimelineSession(timelineFixture(t, "v3-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	stage := "stage-build"
	s.CurrentStageEventID = &stage
	return s
}

func hostEventFrom(source TimelineSource, eventType string) TimelineHostEvent {
	return TimelineHostEvent{SchemaVersion: 3, Source: source, EventType: eventType, OccurredAt: "2026-11-04T10:00:00Z", Data: TimelineHostEventData{}}
}

func TestTimelineHostEventActsAsTheAcceptedProducerOfItsSession(t *testing.T) {
	s := hostEventSession(t)
	for _, p := range s.Producers {
		actor, err := ValidateTimelineHostEventSession(s, hostEventFrom(p.Source, "activity"))
		if err != nil {
			t.Fatalf("producer %s: %v", p.ProducerID, err)
		}
		if actor.Kind != "native_producer" || actor.ProducerID != p.ProducerID {
			t.Fatalf("actor = %+v, want producer %s", actor, p.ProducerID)
		}
	}
}

func TestTimelineHostEventFromANonProducerSessionIsRejected(t *testing.T) {
	s := hostEventSession(t)
	stranger := TimelineSource{ProjectRoot: "/fixture/project", Host: "codex", NativeSessionID: "not-a-producer"}
	if _, err := ValidateTimelineHostEventSession(s, hostEventFrom(stranger, "activity")); TimelineContractCode(err) != "producer_not_accepted" {
		t.Fatalf("stranger session accepted or wrong code: %v", err)
	}
	// A pending join request is not membership.
	pending := TimelineSource{ProjectRoot: "/fixture/project", Host: "grok", NativeSessionID: "pending-native"}
	s.PendingJoinRequests = append(s.PendingJoinRequests, TimelineJoinRequest{RequestID: "request-pending", Source: pending, RequestedAt: "2026-11-04T10:00:00Z", Reason: "Explicit authorized change"})
	if _, err := ValidateTimelineHostEventSession(s, hostEventFrom(pending, "activity")); TimelineContractCode(err) != "producer_not_accepted" {
		t.Fatalf("pending request accepted or wrong code: %v", err)
	}
	// Same native ID on another host is another session.
	other := s.Producers[0].Source
	other.Host = "pi"
	if _, err := ValidateTimelineHostEventSession(s, hostEventFrom(other, "activity")); TimelineContractCode(err) != "producer_not_accepted" {
		t.Fatalf("other host accepted or wrong code: %v", err)
	}
}

func TestTimelineHostEventMustNameTheWorkAndActiveStage(t *testing.T) {
	s := hostEventSession(t)
	source := s.Producers[0].Source
	e := hostEventFrom(source, "activity")
	e.WorkID = s.ViewerSessionID
	e.StageEventID = "stage-build"
	if _, err := ValidateTimelineHostEventSession(s, e); err != nil {
		t.Fatal(err)
	}
	foreignWork := e
	foreignWork.WorkID = strings.Repeat("2", 64)
	if _, err := ValidateTimelineHostEventSession(s, foreignWork); TimelineContractCode(err) != "session_mismatch" {
		t.Fatalf("foreign work: %v", err)
	}
	inactive := e
	inactive.StageEventID = "stage-verify"
	if _, err := ValidateTimelineHostEventSession(s, inactive); TimelineContractCode(err) != "stage_binding_mismatch" {
		t.Fatalf("not the active stage: %v", err)
	}
	s.CurrentStageEventID = nil
	if _, err := ValidateTimelineHostEventSession(s, e); TimelineContractCode(err) != "stage_binding_mismatch" {
		t.Fatalf("no active stage: %v", err)
	}
}

func TestTimelineHostEventOperationIDIsDerivedFromTheEvent(t *testing.T) {
	raw := []byte(`{"schema_version":3,"source":{"project_root":"/fixture/project","host":"claude-code","native_session_id":"native-session-alpha"},"event_type":"task_created","occurred_at":"2026-11-04T10:00:00Z","work_id":"` + strings.Repeat("1", 64) + `","stage_event_id":"stage-build","data":{"task_id":"7","title":"Run typecheck"}}`)
	first, err := DecodeTimelineHostEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	id := first.OperationID()
	if err := timelineOperation(TimelineSchemaVersion, id); err != nil {
		t.Fatalf("derived id %q is not a valid operation id: %v", id, err)
	}
	// Re-sending the same bytes, or the same event spelled with other whitespace, derives the same ID.
	again, err := DecodeTimelineHostEvent(raw)
	if err != nil || again.OperationID() != id {
		t.Fatalf("resend derived %q, want %q (%v)", again.OperationID(), id, err)
	}
	spaced, err := DecodeTimelineHostEvent([]byte(strings.ReplaceAll(string(raw), `,"`, ", \n\"")))
	if err != nil || spaced.OperationID() != id {
		t.Fatalf("respelled event derived %q, want %q (%v)", spaced.OperationID(), id, err)
	}
	// Any difference in the event is a different operation.
	for name, mutate := range map[string]func(*TimelineHostEvent){
		"time":    func(e *TimelineHostEvent) { e.OccurredAt = "2026-11-04T10:00:01Z" },
		"type":    func(e *TimelineHostEvent) { e.EventType = "task_updated" },
		"task":    func(e *TimelineHostEvent) { e.Data.TaskID = "8" },
		"title":   func(e *TimelineHostEvent) { e.Data.Title = "Other" },
		"session": func(e *TimelineHostEvent) { e.Source.NativeSessionID = "native-session-beta" },
		"host":    func(e *TimelineHostEvent) { e.Source.Host = "codex" },
		"work":    func(e *TimelineHostEvent) { e.WorkID = strings.Repeat("3", 64) },
		"stage":   func(e *TimelineHostEvent) { e.StageEventID = "stage-verify" },
	} {
		changed := first
		mutate(&changed)
		if changed.OperationID() == id {
			t.Fatalf("changing %s kept the operation ID", name)
		}
	}
}

func TestTimelineHostEventTaskKey(t *testing.T) {
	if got := TimelineHostTaskKey("7"); got != "host:7" {
		t.Fatalf("task key = %q", got)
	}
	// The stored graph accepts the host key; an agent cannot create one.
	s := hostEventSession(t)
	if len(s.Tasks) == 0 {
		t.Fatal("fixture has no task")
	}
	s.Tasks[0].TaskKey = TimelineHostTaskKey("7")
	if err := ValidateTimelineV3Graph(s); err != nil {
		t.Fatalf("host-keyed task rejected in the stored graph: %v", err)
	}
	s.Tasks[0].TaskKey = "host:" + strings.Repeat("a", 101)
	if err := ValidateTimelineV3Graph(s); err == nil {
		t.Fatal("oversized host key accepted")
	}
	s.Tasks[0].TaskKey = "host:"
	if err := ValidateTimelineV3Graph(s); err == nil {
		t.Fatal("empty host key accepted")
	}
	create := []byte(`{"schema_version":3,"operation_id":"op-one","actor":{"kind":"native_producer","producer_id":"producer-original"},"expected_session_version":5,"reason":"Explicit authorized change","plan_order":2,"task_key":"host:7","title":"Mine","summary":"","segments":[{"planned_stage":"build","order":1}]}`)
	if _, err := DecodeTimelineV3Operation("task_create", create); err == nil {
		t.Fatal("agent task_create accepted a host-keyed task")
	}
}

func TestTimelineHostObservedOrigin(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "v3-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	for i := range s.Events {
		if s.Events[i].Origin == "agent_reported" {
			s.Events[i].Origin = "host_observed"
		}
	}
	if err := ValidateTimelineSession(s); err != nil {
		t.Fatalf("host_observed origin rejected: %v", err)
	}
	s.Events[len(s.Events)-1].Origin = "hook"
	if err := ValidateTimelineSession(s); TimelineContractCode(err) != "invalid_origin" {
		t.Fatalf("unknown origin accepted or wrong code: %v", err)
	}
}

func TestTimelineHostActivityAtIsOptionalAndValidated(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "v3-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.HostActivityAt != "" {
		t.Fatalf("fixture carries host activity %q", s.HostActivityAt)
	}
	raw, err := json.Marshal(s)
	if err != nil || strings.Contains(string(raw), "host_activity_at") {
		t.Fatalf("absent host activity is serialized: %v", err)
	}
	s.HostActivityAt = s.UpdatedAt
	if err := ValidateTimelineSession(s); err != nil {
		t.Fatalf("host activity at the update time: %v", err)
	}
	s.HostActivityAt = "yesterday"
	if err := ValidateTimelineSession(s); err == nil {
		t.Fatal("malformed host activity accepted")
	}
	s.HostActivityAt = "2099-01-01T00:00:00Z"
	if err := ValidateTimelineSession(s); TimelineContractCode(err) != "server_time_mismatch" {
		t.Fatalf("host activity after the update time: %v", err)
	}
	decoded, err := DecodeTimelineIndexEntry(timelineFixture(t, "api-index-entry.json"))
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := decoded.(TimelineAPIIndexEntry)
	if !ok {
		t.Fatalf("fixture decoded as %T", decoded)
	}
	entry.HostActivityAt = "2099-01-01T00:00:00Z"
	if err := validateTimelineIndexEntry(entry); TimelineContractCode(err) != "server_time_mismatch" {
		t.Fatalf("index entry host activity after the update time: %v", err)
	}
}
