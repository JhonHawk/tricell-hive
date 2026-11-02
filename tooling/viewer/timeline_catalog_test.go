package viewer

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTimelineCatalogSharedContract(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "catalog-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Action, Kind, ObjectID, Code string
		Valid                              bool
		JSON                               string
	}
	// Explicit tags matter for object_id; the corpus is shared verbatim with TS.
	var raw []struct {
		Name     string `json:"name"`
		Action   string `json:"action"`
		Kind     string `json:"kind"`
		ObjectID string `json:"object_id"`
		Code     string `json:"code"`
		Valid    bool   `json:"valid"`
		JSON     string `json:"json"`
	}
	if err = json.Unmarshal(timelineFixture(t, "catalog-cases.json"), &raw); err != nil {
		t.Fatal(err)
	}
	for _, v := range raw {
		cases = append(cases, struct {
			Name, Action, Kind, ObjectID, Code string
			Valid                              bool
			JSON                               string
		}{v.Name, v.Action, v.Kind, v.ObjectID, v.Code, v.Valid, v.JSON})
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			var err error
			switch c.Action {
			case "shape":
				_, err = DecodeTimelineV3Operation(c.Kind, []byte(c.JSON))
			case "references":
				var value any
				value, err = DecodeTimelineV3Operation(c.Kind, []byte(c.JSON))
				if err == nil {
					err = ValidateTimelineCatalogReferences(s, c.Kind, c.ObjectID, value)
				}
			case "result":
				_, err = DecodeTimelineOperationResult([]byte(c.JSON))
			case "event":
				var r TimelineAppendEventRequest
				r, err = DecodeTimelineAppendEvent([]byte(c.JSON))
				if err == nil {
					err = ValidateTimelineEventReferences(s, r)
				}
			case "operation_references":
				var value any
				value, err = DecodeTimelineV3Operation(c.Kind, []byte(c.JSON))
				if err == nil {
					err = ValidateTimelineV3OperationReferences(s, c.Kind, c.ObjectID, value)
				}
			case "session":
				_, err = DecodeTimelineSession([]byte(c.JSON))
			default:
				t.Fatal("unsupported action")
			}
			if (err == nil) != c.Valid {
				t.Fatalf("valid=%v got %v", c.Valid, err)
			}
			if err != nil && c.Code != "" && TimelineContractCode(err) != c.Code {
				t.Fatalf("want %s got %v", c.Code, err)
			}
		})
	}
}
func TestTimelineCatalogSnapshotsAndReferences(t *testing.T) {
	s, err := DecodeTimelineSession(timelineFixture(t, "catalog-session.json"))
	if err != nil {
		t.Fatal(err)
	}
	refs, err := CaptureTimelineTicketReferences(s, []string{"catalog-ticket"})
	if err != nil {
		t.Fatal(err)
	}
	old := refs[0]
	for i := range s.Tickets {
		if s.Tickets[i].TicketID == "catalog-ticket" {
			s.Tickets[i].Key = "Corrected"
			s.Tickets[i].URL = "https://example.test/changed"
		}
	}
	next, err := CaptureTimelineTicketReferences(s, []string{"catalog-ticket"})
	if err != nil || refs[0] != old || next[0].Snapshot.Key != "Corrected" {
		t.Fatal("publication context moved", err)
	}
	for _, ids := range [][]string{{"catalog-ticket", "catalog-ticket"}, {"candidate-1"}, {"foreign"}} {
		if _, err := CaptureTimelineTicketReferences(s, ids); err == nil {
			t.Fatal("invalid ID references accepted")
		}
	}
}
func TestTimelineCatalogIdentityNormalization(t *testing.T) {
	for _, c := range []struct{ input, want string }{{"HTTPS://EXAMPLE.test:443/Tracker/", "https://example.test/Tracker"}, {"https://example.test/a%2Fb/", "https://example.test/a%2Fb"}, {"https://[::1]:443/repository/", "https://[::1]/repository"}} {
		got, err := CanonicalTimelineIdentityURL(c.input)
		if err != nil || got != c.want {
			t.Fatalf("normalization %s %v", got, err)
		}
	}
}
func TestTimelineTaskPlanOrderPersistsAndInvalidatesExactlyOnce(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	stage := domainStage(t, s, id, "plan")
	actor := TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}
	doc, err := s.TimelineAddDocument(id, TimelineDocumentMetadata{Actor: actor, SchemaVersion: 3, OperationID: "order-doc", MediaType: "text/markdown"}, []byte("# Recorded task scope"))
	if err != nil {
		t.Fatal(err)
	}
	snap, err := s.TimelineSnapshot(id, stage, TimelineDocumentSnapshotRequest{Actor: actor, SchemaVersion: 3, OperationID: "order-snapshot", ExpectedSessionVersion: doc.SessionVersion, ExpectedVersion: 1, DocumentKey: "plan", Status: "complete", Tabs: []TimelineDocumentTab{{Label: "Tasks", DocumentID: doc.DocumentID}}})
	if err != nil {
		t.Fatal(err)
	}
	create := TimelineTaskCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "order-create", actor.ProducerID), TaskKey: "T1", PlanOrder: 1, Title: "Recorded task", Summary: "Keep one identity", Segments: []TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	made, err := s.TimelineV3Operation(id, "task_create", "", &create)
	if err != nil {
		t.Fatal(err)
	}
	current := domainCurrent(t, s, id)
	if current.Tasks[0].PlanOrder != 1 {
		t.Fatal("create ignored plan_order")
	}
	order := int64(2)
	patch := TimelineTaskPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "order-denied", actor.ProducerID), ExpectedTaskVersion: made.TaskVersion, PlanOrder: &order}
	if _, err = s.TimelineV3Operation(id, "task_patch", made.TaskID, &patch); TimelineContractCode(err) != "scope_reference_required" {
		t.Fatal("scope change without reference", err)
	}
	after := domainCurrent(t, s, id)
	if !reflect.DeepEqual(current, after) {
		t.Fatal("rejected order patch changed state")
	}
	patch.OperationID = "order-patch"
	patch.PlanRevisionID = snap.DocumentRevisionID
	changed, err := s.TimelineV3Operation(id, "task_patch", made.TaskID, &patch)
	if err != nil {
		t.Fatal(err)
	}
	after = domainCurrent(t, s, id)
	if after.Tasks[0].PlanOrder != 2 || after.Tasks[0].TaskVersion != made.TaskVersion+1 || changed.TaskVersion != made.TaskVersion+1 {
		t.Fatal("order change was not persisted/versioned", changed, after.Tasks[0])
	}
	replay, err := s.TimelineV3Operation(id, "task_patch", made.TaskID, &patch)
	if err != nil || !reflect.DeepEqual(changed, replay) {
		t.Fatal("order retry duplicated change", err)
	}
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	restored := domainCurrent(t, restarted, id)
	if restored.Tasks[0].PlanOrder != 2 || restored.Tasks[0].TaskVersion != changed.TaskVersion {
		t.Fatal("order lost after restart")
	}
	replay, err = restarted.TimelineV3Operation(id, "task_patch", made.TaskID, &patch)
	if err != nil || !reflect.DeepEqual(changed, replay) {
		t.Fatal("order receipt lost after restart", err)
	}
	noop := TimelineTaskPatchRequest{TimelineNativeOperation: domainBase(t, restarted, id, "order-noop", actor.ProducerID), ExpectedTaskVersion: changed.TaskVersion, PlanOrder: &order}
	result, err := restarted.TimelineV3Operation(id, "task_patch", made.TaskID, &noop)
	if err != nil || result.TaskVersion != changed.TaskVersion {
		t.Fatal("same order invalidated reader", result, err)
	}
	stale := noop
	stale.OperationID = "order-stale"
	if _, err = restarted.TimelineV3Operation(id, "task_patch", made.TaskID, &stale); TimelineContractCode(err) != "version_conflict" {
		t.Fatal("stale order accepted", err)
	}
}
