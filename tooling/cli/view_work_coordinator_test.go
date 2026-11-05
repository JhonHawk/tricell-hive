package main

import (
	"fmt"
	"image/color"
	"testing"
)

// A single sync moves the tasks and then ends the stage; repeating it changes
// nothing.
func TestWorkSyncAppliesTaskStatesBeforeCompletingTheStage(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "finish", "--title", "Finish", "--summary", "One declaration finishes the stage"))
	args := []string{"sync", "--stage", "build", "--state", "completed", "--reason", "Build done", "--tasks", "-"}
	first := decodeSync(t, s.must(syncThreeDone, args...))
	if first.Result != "updated" || len(first.Tasks) != 3 {
		t.Fatalf("first: %+v", first)
	}
	work := s.session(open.WorkID)
	if stage := workLifecycleStage(work, "build"); stage == nil || stage.StageState != "completed" || work.CurrentStageEventID != nil {
		t.Fatalf("stage %+v", stage)
	}
	for _, task := range work.Tasks {
		if task.Segments[0].State != "completed" {
			t.Fatalf("task %s is %s", task.TaskKey, task.Segments[0].State)
		}
	}
	again := decodeSync(t, s.must(syncThreeDone, args...))
	if again.Result != "unchanged" || again.SessionVersion != first.SessionVersion {
		t.Fatalf("repeat: %+v after %+v", again, first)
	}
}

// The reply of an evidence with many captures fits the bounded output: what
// does not fit is counted, never an error after publishing.
func TestWorkEvidenceReplyIsCappedAfterPublishing(t *testing.T) {
	s := newSyncCLI(t)
	dir := t.TempDir()
	s.must("", "open", "--project", s.project, "--key", "many", "--title", "Many", "--summary", "Many captures")
	s.must("", "sync", "--stage", "verify", "--state", "awaiting_validation", "--reason", "Ready")
	args := []string{"evidence", "--stage", "verify", "--view-title", "Everything", "--candidate", "abc123", "--branch", "feature", "--base", "development"}
	for n := 0; n < 100; n++ {
		path := testPNG(t, dir, fmt.Sprintf("shot-%03d.png", n), color.RGBA{R: uint8(n), G: uint8(n * 2), B: 7, A: 255})
		args = append(args, "--image", path)
	}
	out := s.must("", args...)
	r := decodeEvidence(t, out)
	if r.Result != "updated" || r.EventID == "" || r.ArtifactsOmitted == 0 || len(r.ArtifactIDs)+r.ArtifactsOmitted != 100 {
		t.Fatalf("%d bytes: %+v", len(out), r)
	}
}
