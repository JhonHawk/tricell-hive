package main

import (
	"strings"
	"testing"
	"time"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/tooling/viewer"
)

// Defect 2: a task that would be created with no title is skipped, not sent
// invalid and counted.
func TestReconcileSkipsACreateWithNoTitleAndCountsOnlySentTasks(t *testing.T) {
	source := viewer.TimelineSource{ProjectRoot: "/p", Host: "claude-code", NativeSessionID: "s"}
	at := time.Now().Add(-time.Minute)
	facts := reconcileFacts{Tasks: []reconcileTask{
		{ID: "7", State: hostevent.StateNotStarted, At: at},
		{ID: "8", Title: "Known", State: hostevent.StateNotStarted, At: at},
	}}
	events, tasks := reconcileHostEvents(source, facts, time.Time{})
	if tasks != 1 || len(events) != 1 || events[0].Data.TaskID != "8" {
		t.Fatalf("%d tasks, events %+v", tasks, events)
	}
	if err := events[0].Validate(); err != nil {
		t.Fatal(err)
	}
}

// Finding 4: a line past the scanner's limit (a huge tool result) is skipped, and
// the tasks after it are still read.
func TestReadClaudeTasksSkipsOversizedLinesAndContinues(t *testing.T) {
	huge := `{"type":"user","timestamp":"` + ts(-time.Hour) + `","message":{"content":"Task ` + strings.Repeat("x", 5<<20) + `"}}`
	lines := []string{
		toolUse(ts(-3*time.Minute), "toolu_1", "TaskCreate", `{"subject":"Before the huge line"}`),
		toolResult(ts(-3*time.Minute), "toolu_1", "Task #1 created successfully: Before"),
		huge,
		toolUse(ts(-2*time.Minute), "toolu_2", "TaskCreate", `{"subject":"After the huge line"}`),
		toolResult(ts(-2*time.Minute), "toolu_2", "Task #2 created successfully: After"),
		toolUse(ts(-time.Minute), "toolu_3", "TaskUpdate", `{"taskId":"2","status":"completed"}`),
	}
	got, err := readClaudeTasks(strings.NewReader(strings.Join(lines, "\n") + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].ID != "1" || got[1].ID != "2" || got[2].ID != "2" || got[2].State != hostevent.StateCompleted {
		t.Fatalf("tasks %+v", got)
	}
}

// Finding 3: an update whose create was not read carries no title.
func TestReconcileSendsNoTitleForAnUpdateWhoseCreateWasNotSeen(t *testing.T) {
	source := viewer.TimelineSource{ProjectRoot: "/p", Host: "claude-code", NativeSessionID: "s"}
	at := time.Now().Add(-time.Minute)
	facts := reconcileFacts{Tasks: []reconcileTask{
		{ID: "5", State: hostevent.StateInProgress, At: at},
		{ID: "5", State: hostevent.StateCompleted, At: at.Add(time.Second)},
		{ID: "6", Title: "Known", State: hostevent.StateNotStarted, At: at},
	}}
	events, tasks := reconcileHostEvents(source, facts, time.Time{})
	if tasks != 3 || len(events) != 3 {
		t.Fatalf("%d tasks, %d events", tasks, len(events))
	}
	for _, e := range events {
		if e.Validate() != nil {
			t.Fatalf("invalid event %+v: %v", e, e.Validate())
		}
	}
	if events[0].Data.Title != "" || events[1].Data.Title != "" || events[2].Data.Title != "Known" {
		t.Fatalf("titles %q %q %q", events[0].Data.Title, events[1].Data.Title, events[2].Data.Title)
	}
}
