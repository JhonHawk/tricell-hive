package main

import (
	"testing"
	"tricell-hive/tooling/viewer"
)

// Finding 6: tasks mirrored from the host (host:<id>, ordered from 1,000,000)
// do not move the plan and segment orders of the tasks the agent declares.
func TestSyncMissingTasksIgnoresMirroredHostTasks(t *testing.T) {
	s := viewer.TimelineSession{Tasks: []viewer.TimelineTask{{
		TaskKey: "host:7", PlanOrder: 1_000_000,
		Segments: []viewer.TimelineTaskSegment{{PlannedStage: "build", Order: 1_000_000}},
	}}}
	c := viewWorkCommand{Stage: "build", SyncTasks: []viewWorkSyncTask{{Key: "T1", Title: "First"}, {Key: "T2", Title: "Second"}}}
	got := (&viewWorkEngine{}).syncMissingTasks(s, c)
	if len(got) != 2 || got[0].Order != 1 || got[1].Order != 2 || got[0].Segments[0].Order != 1 || got[1].Segments[0].Order != 2 {
		t.Fatalf("orders %+v", got)
	}
}
