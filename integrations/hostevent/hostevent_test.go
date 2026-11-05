package hostevent

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeFoldsEveryHostSpelling(t *testing.T) {
	for _, name := range []string{"SessionStart", "session_start", "sessionStart", "session-start", "SESSION.START"} {
		if Normalize(name) != "sessionstart" {
			t.Fatalf("%q -> %q", name, Normalize(name))
		}
	}
}

func TestWhenReadsTextSecondsAndMilliseconds(t *testing.T) {
	m := Decode([]byte(`{"a":"2026-11-04T10:00:00Z","b":1792477992,"c":1792477992000,"d":"nope"}`))
	want := time.Unix(1792477992, 0).UTC()
	if !When(m, "a").Equal(time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)) || !When(m, "b").Equal(want) || !When(m, "c").Equal(want) {
		t.Fatalf("unexpected times %v %v %v", When(m, "a"), When(m, "b"), When(m, "c"))
	}
	if !When(m, "d", "missing").IsZero() {
		t.Fatal("unreadable text must give no time")
	}
}

func TestDecodeRejectsNonObjects(t *testing.T) {
	for _, raw := range []string{"", "[]", "7", "not json", "{"} {
		if Decode([]byte(raw)) != nil {
			t.Fatalf("%q must not decode", raw)
		}
	}
}

func TestTaskCleansTitleAndRejectsBadID(t *testing.T) {
	e, ok := Event{}.Task(TaskCreated, "7", "  Run\n\ttypecheck  "+strings.Repeat("x", 200), StateNotStarted)
	if !ok || strings.ContainsAny(e.Title, "\n\t") || len([]rune(e.Title)) > 120 {
		t.Fatalf("title not cleaned: %q", e.Title)
	}
	if _, ok = (Event{}).Task(TaskCreated, "a b", "t", ""); ok {
		t.Fatal("an ID with a space must be refused")
	}
	if e, _ = (Event{}).Task(TaskCreated, "7", "  ", ""); e.Title != "Task 7" {
		t.Fatalf("empty created title must be neutral, got %q", e.Title)
	}
	if e, _ = (Event{}).Task(TaskCompleted, "7", "", ""); e.Title != "" {
		t.Fatalf("a completed task may have no title, got %q", e.Title)
	}
}
