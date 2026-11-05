package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// hookFixture is a viewer that runs, a Claude Code session that opened a work
// without a plan, and a process tree for the hook.
type hookFixture struct {
	*syncCLI
	work  string
	stage string
}

const hookSyncSession = "sync-session-one" // the session newSyncCLI resolves

func newHookFixture(t *testing.T) *hookFixture {
	s := newSyncCLI(t)
	withHookHost(t, "claude", map[string]string{"CLAUDE_CODE_SESSION_ID": hookSyncSession})
	return &hookFixture{syncCLI: s}
}

func (h *hookFixture) openBuild(key string) {
	h.t.Helper()
	open := decodeSync(h.t, h.must("", "open", "--project", h.project, "--key", key, "--title", "Build "+key, "--summary", "Host events"))
	h.work = open.WorkID
	decodeSync(h.t, h.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build"))
	h.stage = *h.session(h.work).CurrentStageEventID
}

func (h *hookFixture) hook(event, payload string) {
	h.t.Helper()
	mustHook(h.t, h.f.stateDir, payload, "claude", event)
}

func (h *hookFixture) payload(extra string) string {
	return `{"session_id":"` + hookSyncSession + `","cwd":"` + h.project + `"` + extra + `}`
}

func (h *hookFixture) task(key string) *viewer.TimelineTask {
	for _, task := range h.session(h.work).Tasks {
		if task.TaskKey == key {
			t := task
			return &t
		}
	}
	return nil
}

func (h *hookFixture) queued() []string {
	return hostEventQueueNames(filepath.Join(h.f.stateDir, "hooks", "queue"))
}

// AC4: a task created and completed in the host's list shows up in the current
// work, in its active stage, with no call from the agent.
func TestHookMirrorsNativeTasksIntoTheCurrentWork(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("native-tasks")
	h.hook("SessionStart", h.payload(`,"source":"startup"`))
	h.hook("TaskCreated", h.payload(`,"task_id":"7","task_subject":"Run typecheck"`))
	task := h.task("host:7")
	if task == nil || task.Title != "Run typecheck" || task.State != "not_started" || task.Segments[0].DisplayStageEventID != h.stage {
		t.Fatalf("created task %+v (stage %s)", task, h.stage)
	}
	h.hook("TaskCompleted", h.payload(`,"task_id":"7","task_subject":"Run typecheck"`))
	if task = h.task("host:7"); task == nil || task.State != "completed" || len(h.session(h.work).Tasks) != 1 {
		t.Fatalf("completed task %+v", task)
	}
	if got := h.session(h.work); got.CurrentStageEventID == nil || *got.CurrentStageEventID != h.stage {
		t.Fatal("a host event must not change the stage")
	}
	if len(h.queued()) != 0 {
		t.Fatalf("nothing should wait: %v", h.queued())
	}
}

func TestHookDoesNotAddTasksToAWorkWithPlanTasks(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("plan-tasks")
	h.must(syncThreeTasks, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Plan tasks", "--tasks", "-")
	before := len(h.session(h.work).Tasks)
	h.hook("TaskCreated", h.payload(`,"task_id":"9","task_subject":"Native one"`))
	if after := len(h.session(h.work).Tasks); after != before || h.task("host:9") != nil {
		t.Fatalf("tasks went from %d to %d", before, after)
	}
}

func TestHookActivityMovesWithoutAWorkNamedAndChangesNoVersion(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("activity")
	versionBefore := h.session(h.work).Version
	h.hook("UserPromptSubmit", h.payload(`,"prompt":"hello"`))
	got := h.session(h.work)
	if got.Version != versionBefore {
		t.Fatalf("host activity must not invalidate agent versions: %d -> %d", versionBefore, got.Version)
	}
}

// T6: with the viewer down the event waits in the host event queue; the next run
// with the viewer up sends it, resolving work and stage at that moment.
func TestHookQueuesWhileTheViewerIsDownAndSendsLater(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("viewer-down")
	h.f.mustView("stop")
	h.hook("TaskCreated", h.payload(`,"task_id":"1","task_subject":"Queued task"`))
	h.hook("TaskCreated", h.payload(`,"task_id":"1","task_subject":"Queued task"`)) // the same event twice is one
	if got := h.queued(); len(got) != 1 {
		t.Fatalf("queue %v", got)
	}
	if _, err := viewTimelineRunning(newViewPaths(h.f.stateDir)); err == nil {
		t.Fatal("a hook started the viewer")
	}
	h.f.mustView("start")
	h.hook("UserPromptSubmit", h.payload(`,"prompt":"go on"`))
	if got := h.queued(); len(got) != 0 {
		t.Fatalf("the queue should be sent: %v", got)
	}
	if task := h.task("host:1"); task == nil || task.Title != "Queued task" {
		t.Fatalf("the queued task never arrived: %+v", task)
	}
}

// T6: the hook flushes the work's semantic queue (T4) when the viewer is up, so
// a work opened while the viewer was down exists before the host event binds to it.
func TestHookFlushesTheSemanticQueueBeforeItsEvent(t *testing.T) {
	h := newHookFixture(t)
	h.f.mustView("stop")
	queued := decodeQueued(t, h.must("", "open", "--project", h.project, "--key", "queued-open", "--title", "Queued open", "--summary", "Opened with the viewer down"))
	decodeQueued(t, h.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build"))
	h.hook("TaskCreated", h.payload(`,"task_id":"3","task_subject":"Before the viewer"`))
	if len(h.queued()) != 1 {
		t.Fatalf("the event should wait: %v", h.queued())
	}
	h.f.mustView("start")
	h.hook("SessionStart", h.payload(`,"source":"resume"`))

	ids, err := readCurrentWorks(h.f.stateDir, "claude-code", hookSyncSession)
	if err != nil || len(ids) != 1 || ids[0] == queued.WorkID {
		t.Fatalf("the queued open should have become a real work: %v %v", ids, err)
	}
	h.work = ids[0]
	if got := readJournalQueue(t, h.f.stateDir); got != 0 {
		t.Fatalf("%d declarations still queued", got)
	}
	stage := workLifecycleStage(h.session(h.work), "build")
	if stage == nil || stage.StageState != "in_progress" {
		t.Fatalf("the queued stage was not created: %+v", stage)
	}
	if len(h.queued()) != 0 {
		t.Fatalf("the host event should be sent after the declarations: %v", h.queued())
	}
	if task := h.task("host:3"); task == nil || task.Segments[0].DisplayStageEventID != stage.EventID {
		t.Fatalf("the task should bind to the flushed stage: %+v", task)
	}
}

func readJournalQueue(t *testing.T, state string) int {
	t.Helper()
	j, err := readWorkJournal(filepath.Join(state, "publisher", "v1"))
	if err != nil {
		t.Fatal(err)
	}
	return len(j.Queue)
}

func TestHookEventForAClosedWorkOnlyMovesActivity(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("closing")
	h.must("", "close", "--outcome", "completed", "--reason", "Done")
	h.hook("TaskCreated", h.payload(`,"task_id":"4","task_subject":"After close"`))
	if h.task("host:4") != nil {
		t.Fatal("a closed work takes no native task")
	}
	if len(h.queued()) != 0 {
		t.Fatalf("an event the viewer accepted as activity must not wait: %v", h.queued())
	}
}

func TestHookQueueIsBoundedAndPrivate(t *testing.T) {
	state := t.TempDir()
	event := viewer.TimelineHostEvent{SchemaVersion: viewer.TimelineSchemaVersion, Source: viewer.TimelineSource{ProjectRoot: "/w", Host: "claude-code", NativeSessionID: "cc-1"}, EventType: "activity"}
	for i := 0; i < hostEventQueueLimit+10; i++ {
		event.OccurredAt = identityTestStart.Add(time.Duration(i) * time.Second).Format("2006-01-02T15:04:05Z07:00")
		queueHostEvent(state, event)
	}
	dir := filepath.Join(state, "hooks", "queue")
	names := hostEventQueueNames(dir)
	if len(names) != hostEventQueueLimit {
		t.Fatalf("%d queued, want %d", len(names), hostEventQueueLimit)
	}
	st, err := os.Stat(filepath.Join(dir, names[0]))
	if err != nil || st.Mode().Perm() != 0600 {
		t.Fatalf("queued file mode %v %v", st.Mode(), err)
	}
	if !strings.HasSuffix(names[0], ".json") {
		t.Fatal(names[0])
	}
}
