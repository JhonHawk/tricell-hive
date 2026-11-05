package main

import (
	"fmt"
	"tricell-hive/tooling/viewer"
)

// `work sync` applies one stage declaration: the stage state, the tasks that
// are missing (bound to the stage), the task states and, with --close, the end
// of the work. It is built from the lifecycle operations, which keep their own
// journaled intents, so every step stays resumable and the low-level commands
// do not change.

// viewWorkSyncTaskReceipt is the identity and state of one declared task in the
// declared stage, read back from the viewer.
type viewWorkSyncTaskReceipt struct {
	Key       string `json:"key"`
	TaskID    string `json:"task_id"`
	SegmentID string `json:"segment_id"`
	State     string `json:"state"`
}

type viewWorkSyncResult struct {
	WorkID         string                    `json:"work_id"`
	Result         string                    `json:"result"`
	SessionVersion int64                     `json:"session_version"`
	StageID        string                    `json:"stage_id,omitempty"`
	State          string                    `json:"state,omitempty"`
	Tasks          []viewWorkSyncTaskReceipt `json:"tasks,omitempty"`
	TasksOmitted   int                       `json:"tasks_omitted,omitempty"`
	Queue          *viewWorkQueueReport      `json:"queue,omitempty"`
}

var viewWorkSyncTaskStates = map[string]bool{"": true, "not_started": true, "in_progress": true, "in_review": true, "blocked": true, "completed": true}

// validateViewWorkSyncTasks checks a declared task list with the validators the
// viewer applies to a task creation and a task transition, so a rejected input
// names its field and rule, never a value.
func validateViewWorkSyncTasks(tasks []viewWorkSyncTask) error {
	if len(tasks) == 0 || len(tasks) > viewer.TimelineMaxTasks {
		return faultf("tasks", "invalid task count")
	}
	seen := map[string]bool{}
	for i, t := range tasks {
		path := fmt.Sprintf("tasks[%d]", i)
		if seen[t.Key] {
			return faultf(path+".key", "duplicate task key")
		}
		seen[t.Key] = true
		r := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workValidationOperation("Declare tasks"), PlanOrder: int64(i + 1), TaskKey: t.Key, Title: t.Title, Summary: syncTaskSummary(t), Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: "build", Order: int64(i + 1)}}}
		if err := viewer.ValidateTimelineV3Operation(&r); err != nil {
			return viewerFault(err, path, map[string]string{"task_key": "key", "plan_order": "order"})
		}
		if !viewWorkSyncTaskStates[t.State] {
			return faultf(path+".state", "unsupported state")
		}
		if t.State != "" && t.State != "not_started" {
			tr := viewer.TimelineTaskTransitionRequest{TimelineNativeOperation: workValidationOperation("Declare tasks"), ExpectedTaskVersion: 1, SegmentID: "validate", State: t.State, Block: t.Block, Resolution: t.Resolution}
			if err := viewer.ValidateTimelineV3Operation(&tr); err != nil {
				return viewerFault(err, path, map[string]string{"segment_id": "state"})
			}
		} else if t.Block != nil || t.Resolution != "" {
			return faultf(path+".state", "block and resolution need a state")
		}
	}
	return nil
}

// syncTaskSummary: a declared task needs no separate summary; its title says it.
func syncTaskSummary(t viewWorkSyncTask) string {
	if t.Summary != "" {
		return t.Summary
	}
	return t.Title
}

// syncStatePath returns the task states to pass through, in order, to go from
// one state to another along the viewer's allowed transitions.
func syncStatePath(from, to string) []string {
	allowed := map[string][]string{"not_started": {"in_progress", "blocked"}, "in_progress": {"in_review", "blocked", "completed"}, "in_review": {"in_progress", "blocked", "completed"}, "blocked": {"in_progress"}, "completed": {"in_progress"}}
	if from == to {
		return nil
	}
	previous := map[string]string{from: ""}
	queue := []string{from}
	for len(queue) > 0 {
		at := queue[0]
		queue = queue[1:]
		for _, next := range allowed[at] {
			if _, ok := previous[next]; ok {
				continue
			}
			previous[next] = at
			if next == to {
				var path []string
				for n := to; n != from; n = previous[n] {
					path = append([]string{n}, path...)
				}
				return path
			}
			queue = append(queue, next)
		}
	}
	return []string{to} // the viewer states why this transition is not allowed
}

// workHasPlan: the work has a change folder bound in this state directory or has
// published a plan, so its tasks are registered against a confirmed plan
// revision.
func (e *viewWorkEngine) workHasPlan(s viewer.TimelineSession) bool {
	for _, b := range e.journal.Bindings {
		if b.WorkID == s.ViewerSessionID && b.Change != "" {
			return true
		}
	}
	for _, event := range s.Events {
		if event.Kind == "stage" && event.Stage == "plan" && !event.Withdrawn && len(event.DocumentSnapshots) > 0 {
			return true
		}
	}
	return false
}

func (e *viewWorkEngine) syncMissingTasks(s viewer.TimelineSession, c viewWorkCommand) []viewWorkTaskInput {
	var planOrder, segmentOrder int64
	for _, t := range s.Tasks {
		planOrder = max(planOrder, t.PlanOrder)
		for _, seg := range t.Segments {
			if string(seg.PlannedStage) == c.Stage {
				segmentOrder = max(segmentOrder, seg.Order)
			}
		}
	}
	var missing []viewWorkTaskInput
	for _, t := range c.SyncTasks {
		if workLifecycleTask(s, t.Key) != nil {
			continue
		}
		planOrder++
		segmentOrder++
		missing = append(missing, viewWorkTaskInput{Key: t.Key, Title: t.Title, Summary: syncTaskSummary(t), Order: planOrder, Segments: []viewWorkSegmentInput{{Stage: viewer.TimelineStage(c.Stage), Order: segmentOrder}}})
	}
	return missing
}

func (e *viewWorkEngine) sync(c viewWorkCommand) (viewWorkSyncResult, error) {
	s, err := e.session(c.WorkID)
	if err != nil {
		return viewWorkSyncResult{}, err
	}
	if s.State == viewer.TimelineClosed && c.CloseOutcome != "" {
		// The declaration already ended the work: repeating it changes nothing.
		return e.syncReceipt(c, s, false)
	}
	changed := false
	apply := func(r viewWorkResult, err error) error {
		if err == nil && r.Result != "unchanged" {
			changed = true
		}
		return err
	}
	stage := c
	stage.Action = "stage"
	// A stage starts in progress. Declaring it in another state starts it first,
	// when this declaration is the one that starts an occurrence.
	latest := workLifecycleStage(s, c.Stage)
	terminal := latest != nil && (latest.StageState == "completed" || latest.StageState == "failed")
	if c.State != "in_progress" && (latest == nil && c.PreviousStage == "" || terminal && c.PreviousStage != "") {
		start := stage
		start.State = "in_progress"
		if err = apply(e.lifecycle(start)); err != nil {
			return viewWorkSyncResult{}, err
		}
		stage.PreviousStage = ""
	}
	// Ending the stage waits for the task states: a task moves only in an
	// active stage.
	ending := c.State == "completed" || c.State == "failed"
	if !ending {
		if err = apply(e.lifecycle(stage)); err != nil {
			return viewWorkSyncResult{}, err
		}
	}
	if len(c.SyncTasks) > 0 {
		if s, err = e.session(c.WorkID); err != nil {
			return viewWorkSyncResult{}, err
		}
		if missing := e.syncMissingTasks(s, c); len(missing) > 0 {
			register := c
			register.Action, register.Tasks = "tasks register", missing
			if err = apply(e.lifecycle(register)); err != nil {
				return viewWorkSyncResult{}, err
			}
		}
	}
	for _, t := range c.SyncTasks {
		if t.State == "" || t.State == "not_started" {
			continue
		}
		if s, err = e.session(c.WorkID); err != nil {
			return viewWorkSyncResult{}, err
		}
		task := workLifecycleTask(s, t.Key)
		if task == nil || task.Withdrawn {
			return viewWorkSyncResult{}, workStateError("task_not_found")
		}
		segment := workLifecycleSegment(task, c.Stage)
		if segment == nil {
			return viewWorkSyncResult{}, workStateError("task_segment_not_found")
		}
		from := segment.State
		for _, next := range syncStatePath(from, t.State) {
			update := c
			update.Action, update.Key, update.State = "task update", t.Key, next
			update.Block, update.Resolution = nil, ""
			if next == "blocked" {
				update.Block = t.Block
			}
			if from == "blocked" {
				update.Resolution = t.Resolution
			}
			if err = apply(e.lifecycle(update)); err != nil {
				return viewWorkSyncResult{}, err
			}
			from = next
		}
	}
	if ending {
		if err = apply(e.lifecycle(stage)); err != nil {
			return viewWorkSyncResult{}, err
		}
	}
	if c.CloseOutcome != "" {
		closing := c
		closing.Action, closing.Outcome = "close", c.CloseOutcome
		if err = apply(e.closeWork(closing)); err != nil {
			return viewWorkSyncResult{}, err
		}
	}
	if s, err = e.session(c.WorkID); err != nil {
		return viewWorkSyncResult{}, err
	}
	return e.syncReceipt(c, s, changed)
}

// syncReceipt reads the declaration back from the viewer: the stage, and every
// declared task bound to that stage in the declared state. A mismatch is an
// error, never a success.
func (e *viewWorkEngine) syncReceipt(c viewWorkCommand, s viewer.TimelineSession, changed bool) (viewWorkSyncResult, error) {
	result := viewWorkSyncResult{WorkID: s.ViewerSessionID, Result: "unchanged", SessionVersion: s.Version}
	if changed {
		result.Result = "updated"
	}
	stage := workLifecycleStage(s, c.Stage)
	if stage == nil {
		return result, workStateError("readback_mismatch")
	}
	result.StageID = stage.EventID
	if s.State == viewer.TimelineClosed {
		result.State = "closed"
	}
	for _, t := range c.SyncTasks {
		task := workLifecycleTask(s, t.Key)
		if task == nil || task.Withdrawn {
			return result, &viewWorkError{Code: "readback_mismatch", Field: "task " + t.Key}
		}
		segment := workLifecycleSegment(task, c.Stage)
		if segment == nil || segment.DisplayStageEventID != stage.EventID || t.State != "" && segment.State != t.State {
			return result, &viewWorkError{Code: "readback_mismatch", Field: "task " + t.Key}
		}
		result.Tasks = append(result.Tasks, viewWorkSyncTaskReceipt{Key: t.Key, TaskID: task.TaskID, SegmentID: segment.SegmentID, State: segment.State})
	}
	// The receipt fits the bounded mutation output: past it, the last tasks are
	// counted, and `work show` lists them all.
	for len(result.Tasks) > 0 && !fitsViewWorkOutput(result) {
		result.Tasks = result.Tasks[:len(result.Tasks)-1]
		result.TasksOmitted++
	}
	return result, nil
}

func fitsViewWorkOutput(value any) bool {
	var out countingWriter
	return writeViewWorkBoundedJSON(&out, value) == nil
}

type countingWriter struct{ n int }

func (w *countingWriter) Write(p []byte) (int, error) { w.n += len(p); return len(p), nil }

// closeWork completes the active stage, closes the work and reads it back.
func (e *viewWorkEngine) closeWork(c viewWorkCommand) (viewWorkResult, error) {
	s, err := e.session(c.WorkID)
	if err != nil {
		return viewWorkResult{}, err
	}
	changed := false
	if s.CurrentStageEventID != nil {
		for _, event := range s.Events {
			if event.EventID != *s.CurrentStageEventID || event.StageState == "completed" || event.StageState == "failed" {
				continue
			}
			stage := c
			stage.Action, stage.Stage, stage.State = "stage", string(event.Stage), "completed"
			if _, err = e.lifecycle(stage); err != nil {
				return viewWorkResult{}, err
			}
			changed = true
		}
	}
	r, err := e.lifecycle(c)
	if err != nil {
		return viewWorkResult{}, err
	}
	if s, err = e.session(c.WorkID); err != nil {
		return viewWorkResult{}, err
	}
	if s.State != viewer.TimelineClosed {
		return viewWorkResult{}, &viewWorkError{Code: "readback_mismatch", Field: "state"}
	}
	r.State = "closed"
	r.SessionVersion = s.Version
	if changed && r.Result == "unchanged" {
		r.Result = "updated"
	}
	return r, nil
}
