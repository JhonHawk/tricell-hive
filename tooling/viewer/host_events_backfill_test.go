package viewer

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHostActivityIsStampedWithTheEventsOwnTimeClampedToTheServerClock(t *testing.T) {
	h := newHostEventHarness(t)
	id := h.work.ViewerSessionID
	h.advance()
	h.advance()
	send := func(occurred time.Time) TimelineSession {
		t.Helper()
		e := TimelineHostEvent{SchemaVersion: 3, Source: h.source, EventType: "activity", OccurredAt: occurred.UTC().Format(time.RFC3339Nano), WorkID: id, Data: TimelineHostEventData{}}
		raw, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		h.mustSend(raw)
		return h.current()
	}
	earlier := h.clock.Add(-30 * time.Second)
	v := send(earlier)
	if want := earlier.Format(time.RFC3339Nano); v.LastActivityAt != want || v.HostActivityAt != want {
		t.Fatalf("a backfilled event is stamped with its own time: last %s host %s, want %s", v.LastActivityAt, v.HostActivityAt, want)
	}
	// An older event never moves activity backwards.
	if v = send(earlier.Add(-time.Minute)); v.LastActivityAt != earlier.Format(time.RFC3339Nano) || v.HostActivityAt != earlier.Format(time.RFC3339Nano) {
		t.Fatalf("activity moved backwards: last %s host %s", v.LastActivityAt, v.HostActivityAt)
	}
	// A time ahead of the server's clock is clamped to it.
	if v = send(h.clock.Add(time.Hour)); v.LastActivityAt != h.clock.Format(time.RFC3339Nano) || v.HostActivityAt != h.clock.Format(time.RFC3339Nano) {
		t.Fatalf("a future event time is clamped to the server's now: last %s host %s", v.LastActivityAt, v.HostActivityAt)
	}
}
