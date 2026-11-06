package viewer

import (
	"encoding/json"
	"testing"
	"time"
)

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// A native task from before the work existed belongs to earlier work in the
// same session: it is not mirrored into this one.
func TestHostTaskEventBeforeTheWorkWasCreatedIsIgnored(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	stageID := h.stage("build")
	h.advance()
	before := h.current()
	old := TimelineHostEvent{SchemaVersion: 3, Source: h.source, EventType: "task_created", OccurredAt: h.work.CreatedAt, WorkID: id, StageEventID: stageID, Data: hostTask("old-1", "Earlier work")}
	created, err := time.Parse(time.RFC3339Nano, h.work.CreatedAt)
	if err != nil {
		t.Fatal(err)
	}
	old.OccurredAt = created.Add(-time.Hour).UTC().Format(time.RFC3339Nano)
	raw := mustJSON(t, old)
	if result := h.mustSend(raw); result.Status != "activity_only" {
		t.Fatalf("status %q", result.Status)
	}
	if after := h.current(); len(after.Tasks) != len(before.Tasks) || after.Version != before.Version {
		t.Fatalf("an earlier task was mirrored: %+v", after.Tasks)
	}
	fresh := old
	fresh.OccurredAt = h.clock.UTC().Format(time.RFC3339Nano)
	fresh.Data = hostTask("new-1", "This work")
	if result := h.mustSend(mustJSON(t, fresh)); result.Status != "applied" {
		t.Fatalf("status %q", result.Status)
	}
	if _, ok := hostTaskByKey(h.current(), "host:new-1"); !ok {
		t.Fatal("a task of this work was not mirrored")
	}
}

func TestHostEventTurnIDDistinguishesActivityOperations(t *testing.T) {
	base := TimelineHostEvent{SchemaVersion: 3, Source: TimelineSource{ProjectRoot: "/fixture/project", Host: "grok", NativeSessionID: "s"}, EventType: "activity", OccurredAt: "2026-11-04T10:00:00Z"}
	a, b := base, base
	a.Data.TurnID, b.Data.TurnID = "prompt-1", "prompt-2"
	if a.Validate() != nil || b.Validate() != nil || a.OperationID() == b.OperationID() {
		t.Fatalf("turn IDs must keep two activity events apart: %v %v", a.Validate(), b.Validate())
	}
	task := base
	task.EventType = "task_created"
	task.Data = TimelineHostEventData{TaskID: "1", Title: "T", TurnID: "prompt-1"}
	if TimelineContractCode(task.Validate()) != "invalid_id" {
		t.Fatalf("a task event takes no turn ID: %v", task.Validate())
	}
	bad := base
	bad.Data.TurnID = "has space"
	if bad.Validate() == nil {
		t.Fatal("an invalid turn ID was accepted")
	}
}
