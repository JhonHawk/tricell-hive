package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

const queueHost, queueNative = "codex", "queue-session"

// queueCLI runs `hive view work` commands as one native session with explicit
// identity, against the fixture's viewer.
type queueCLI struct {
	t       *testing.T
	f       *viewFixture
	project string
	// native is the native session ID; empty is queueNative.
	native string
}

// as returns the same fixture driven by another native session of the host.
func (q *queueCLI) as(native string) *queueCLI {
	other := *q
	other.native = native
	return &other
}

func newQueueCLI(t *testing.T) *queueCLI {
	f, c := workEngineFixture(t)
	return &queueCLI{t: t, f: f, project: c.Project}
}

func (q *queueCLI) args(args ...string) []string {
	native := q.native
	if native == "" {
		native = queueNative
	}
	return append(append([]string{}, args...), "--host", queueHost, "--native-session-id", native, "--state-dir", q.f.stateDir)
}

func (q *queueCLI) run(stdin string, args ...string) (string, error) {
	q.t.Helper()
	var out bytes.Buffer
	err := runViewWork(q.args(args...), &out, strings.NewReader(stdin))
	return out.String(), err
}

func (q *queueCLI) must(stdin string, args ...string) string {
	q.t.Helper()
	out, err := q.run(stdin, args...)
	if err != nil {
		q.t.Fatalf("hive view work %s: %v", strings.Join(args, " "), err)
	}
	return out
}

func (q *queueCLI) open(key string) string {
	return q.must("", "open", "--project", q.project, "--key", key, "--title", "Work "+key, "--summary", "Queued publication")
}

func (q *queueCLI) journal() *viewWorkJournal {
	q.t.Helper()
	j, err := readWorkJournal(filepath.Join(q.f.stateDir, "publisher", "v1"))
	if err != nil {
		q.t.Fatal(err)
	}
	return j
}

func (q *queueCLI) session(id string) viewer.TimelineSession {
	q.t.Helper()
	r, err := viewer.DecodeTimelineSessionResponse([]byte(q.f.mustView("session", "get", id)))
	if err != nil {
		q.t.Fatal(err)
	}
	return r.Session
}

func (q *queueCLI) workCount() int {
	q.t.Helper()
	index, err := viewIndex(q.f.info())
	if err != nil {
		q.t.Fatal(err)
	}
	return len(index.TimelineSessions)
}

func (q *queueCLI) workIDs() []string {
	q.t.Helper()
	ids, err := readCurrentWorks(q.f.stateDir, queueHost, queueNative)
	if err != nil {
		q.t.Fatal(err)
	}
	return ids
}

func decodeQueued(t *testing.T, out string) viewWorkQueuedResult {
	t.Helper()
	var r viewWorkQueuedResult
	if err := json.Unmarshal([]byte(out), &r); err != nil || r.Result != "queued" {
		t.Fatalf("not a queued reply: %q %v", out, err)
	}
	return r
}

func stageEvents(s viewer.TimelineSession, stage string) int {
	n := 0
	for _, e := range s.Events {
		if e.Kind == "stage" && string(e.Stage) == stage && !e.Withdrawn {
			n++
		}
	}
	return n
}

func (q *queueCLI) assertViewerNotStarted() {
	q.t.Helper()
	if _, err := viewTimelineRunning(newViewPaths(q.f.stateDir)); err == nil {
		q.t.Fatal("a queued command started the viewer")
	}
}

// AC5: with the viewer stopped, open and sync answer queued and never start
// it; the next command with the viewer running creates the work and the stage
// once, in order.
func TestWorkQueueStoppedViewerThenRunningCreatesWorkAndStageOnce(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")

	opened := decodeQueued(t, q.open("queued-build"))
	synced := decodeQueued(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build"))
	if synced.WorkID != opened.WorkID || opened.Position != 1 || synced.Position != 2 || opened.OperationID == synced.OperationID {
		t.Fatalf("open %+v sync %+v", opened, synced)
	}
	q.assertViewerNotStarted()
	if got := q.workIDs(); len(got) != 1 || got[0] != opened.WorkID {
		t.Fatalf("current works %v", got)
	}
	if got := len(q.journal().Queue); got != 2 {
		t.Fatalf("%d queued entries", got)
	}

	q.f.mustView("start")
	third := decodeSync(t, q.must(`[{"key":"T1","title":"First","state":"in_progress"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Declare tasks", "--tasks", "-"))
	if third.Queue == nil || third.Queue.Sent != 2 || len(third.Queue.Rejected) != 0 || third.WorkID == opened.WorkID || len(third.Tasks) != 1 {
		t.Fatalf("third: %+v", third)
	}
	if n := q.workCount(); n != 1 {
		t.Fatalf("%d works, want 1", n)
	}
	work := q.session(third.WorkID)
	if stageEvents(work, "build") != 1 || work.WorkKey == nil || *work.WorkKey != "queued-build" {
		t.Fatalf("stages %d key %v", stageEvents(work, "build"), work.WorkKey)
	}
	// The work existed before its stage did.
	stage := workLifecycleStage(work, "build")
	if stage == nil || stage.CreatedAt < work.CreatedAt {
		t.Fatalf("stage %+v before work %s", stage, work.CreatedAt)
	}
	j := q.journal()
	if len(j.Queue) != 0 || len(j.Pending) != 0 || len(j.Aliases) != 1 || j.Aliases[0].Alias != opened.WorkID || j.Aliases[0].WorkID != third.WorkID {
		t.Fatalf("journal queue %d pending %d aliases %+v", len(j.Queue), len(j.Pending), j.Aliases)
	}
	if got := q.workIDs(); len(got) != 1 || got[0] != third.WorkID {
		t.Fatalf("current works %v, want the viewer's ID", got)
	}
	// The handle keeps addressing the work.
	if again := decodeSync(t, q.must("", "sync", "--work-id", opened.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Again")); again.WorkID != third.WorkID || again.Result != "unchanged" {
		t.Fatalf("again: %+v", again)
	}
}

// A queued sync that names a task the work lacks, without the title needed to
// create it, is rejected when it is sent and leaves the queue, so it never
// blocks the commands that follow.
func TestWorkQueueRejectsANewTaskWithoutATitle(t *testing.T) {
	q := newQueueCLI(t)
	q.open("untitled")
	started := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start"))
	q.f.mustView("stop")
	decodeQueued(t, q.must(`[{"key":"T9","state":"in_progress"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Untitled", "--tasks", "-"))
	q.f.mustView("start")

	next := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_review", "--reason", "Review"))
	if next.Queue == nil || next.Queue.Sent != 0 || len(next.Queue.Rejected) != 1 || next.Queue.Rejected[0].Reason != "task_title_required" {
		t.Fatalf("next sync: %+v", next)
	}
	if work := q.session(started.WorkID); len(work.Tasks) != 0 || workLifecycleStage(work, "build").StageState != "in_review" {
		t.Fatalf("work %+v", work)
	}
	again := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Back"))
	if again.Queue != nil && len(again.Queue.Rejected) != 0 {
		t.Fatalf("the rejected declaration came back: %+v", again.Queue)
	}
}

// A declaration that no longer applies is reported rejected, not applied, and
// the stage keeps the state another producer gave it.
func TestWorkQueueRejectsADeclarationTheStageMovedPast(t *testing.T) {
	q := newQueueCLI(t)
	q.open("moved-on")
	started := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start"))
	q.f.mustView("stop")
	decodeQueued(t, q.must("", "sync", "--stage", "build", "--state", "completed", "--reason", "Finished here"))
	q.f.mustView("start")

	// Another producer completes the stage while the declaration waits. The
	// engine call skips the queue, as a different session would.
	e, err := openViewWorkEngine(viewWorkCommand{Action: "stage", WorkID: started.WorkID, Options: management.Options{Scope: "user", StateDir: q.f.stateDir}})
	if err != nil {
		t.Fatal(err)
	}
	completed, err := e.lifecycle(viewWorkCommand{Action: "stage", WorkID: started.WorkID, Host: queueHost, NativeSessionID: queueNative, Stage: "build", State: "completed", Reason: "Completed elsewhere"})
	e.Close()
	if err != nil || completed.Result != "updated" {
		t.Fatalf("%+v %v", completed, err)
	}
	before := q.session(started.WorkID)
	stageBefore := workLifecycleStage(before, "build")

	closed := viewWorkResult{}
	if err := json.Unmarshal([]byte(q.must("", "close", "--outcome", "completed", "--reason", "Done")), &closed); err != nil {
		t.Fatal(err)
	}
	if closed.Queue == nil || closed.Queue.Sent != 0 || len(closed.Queue.Rejected) != 1 || closed.Queue.Rejected[0].Action != "sync" || closed.Queue.Rejected[0].Reason != "stage_already_terminal" {
		t.Fatalf("close: %+v", closed)
	}
	after := q.session(started.WorkID)
	stageAfter := workLifecycleStage(after, "build")
	if stageAfter.StageState != "completed" || stageAfter.Version != stageBefore.Version || stageEvents(after, "build") != 1 {
		t.Fatalf("stage %+v after %+v", stageAfter, stageBefore)
	}
	if again := decodeSync(t, q.must("", "sync", "--work-id", started.WorkID, "--stage", "build", "--state", "completed", "--reason", "Check", "--close", "completed")); again.Queue != nil {
		t.Fatalf("the rejection was reported twice: %+v", again.Queue)
	}
	if got := len(q.journal().Queue); got != 0 {
		t.Fatalf("%d entries left", got)
	}
}

// queueChild is the process the tests kill or hold: it runs one command with a
// checkpoint that kills it at the Nth transmitted request, or holds the work
// lock until released.
func TestWorkQueueChild(t *testing.T) {
	raw := os.Getenv("HIVE_TEST_QUEUE_ARGS")
	if raw == "" {
		return
	}
	var args []string
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		t.Fatal(err)
	}
	killAt := 0
	fmt.Sscan(os.Getenv("HIVE_TEST_QUEUE_KILL_AT"), &killAt)
	hold := os.Getenv("HIVE_TEST_QUEUE_HOLD")
	sent := 0
	viewWorkCheckpoint = func(point string) {
		switch {
		case point == "sent" && killAt > 0:
			if sent++; sent == killAt {
				syscall.Kill(os.Getpid(), syscall.SIGKILL)
			}
		case point == "locked" && hold != "":
			os.WriteFile(filepath.Join(hold, "locked"), nil, 0600)
			for n := 0; n < 600; n++ {
				if _, err := os.Stat(filepath.Join(hold, "release")); err == nil {
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
			t.Error("never released")
		}
	}
	var out bytes.Buffer
	if err := runViewWork(args, &out, strings.NewReader("")); err != nil {
		t.Fatal(err)
	}
}

func (q *queueCLI) child(env []string, args ...string) *exec.Cmd {
	raw, _ := json.Marshal(q.args(args...))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	q.t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWorkQueueChild$")
	// The fixture's serve switch would turn the test binary into the CLI.
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "HIVE_VIEW_TEST_SERVE=") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	cmd.Env = append(append(cmd.Env, "HIVE_TEST_QUEUE_ARGS="+string(raw)), env...)
	return cmd
}

func (q *queueCLI) killed(env []string, args ...string) {
	q.t.Helper()
	out, err := q.child(env, args...).CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if !ok || exit.Sys().(syscall.WaitStatus).Signal() != syscall.SIGKILL {
		q.t.Fatalf("the process was not killed: %v\n%s", err, out)
	}
}

// Killing the process while it sends the queue leaves the exact request it was
// sending; the next command replays it and the work and stage exist once.
func TestWorkQueueResumesAfterTheProcessIsKilledMidFlush(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")
	opened := decodeQueued(t, q.open("killed-flush"))
	decodeQueued(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build"))
	q.f.mustView("start")

	// The second transmitted request is the stage start: sent, never recorded.
	q.killed([]string{"HIVE_TEST_QUEUE_KILL_AT=2"}, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Declare")
	j := q.journal()
	if len(j.Queue) != 1 || j.Queue[0].Action != "sync" || j.Queue[0].State != "started" || len(j.Pending) != 1 || !j.Pending[0].Steps[0].Attempted || len(j.Pending[0].Steps[0].Response) != 0 {
		t.Fatalf("after the kill: queue %d pending %d", len(j.Queue), len(j.Pending))
	}

	r := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Declare"))
	if r.Queue == nil || r.Queue.Sent != 1 || r.WorkID == opened.WorkID {
		t.Fatalf("resume: %+v", r)
	}
	work := q.session(r.WorkID)
	if stageEvents(work, "build") != 1 {
		t.Fatalf("%d stage events", stageEvents(work, "build"))
	}
	if n := q.workCount(); n != 1 {
		t.Fatalf("%d works, want 1", n)
	}
	if j = q.journal(); len(j.Queue) != 0 || len(j.Pending) != 0 {
		t.Fatalf("queue %d pending %d", len(j.Queue), len(j.Pending))
	}
}

// Two commands on different works run together; two on one work do not.
func TestWorkLocksArePerWork(t *testing.T) {
	q := newQueueCLI(t)
	a := decodeSync(t, q.open("work-a"))
	b := decodeSync(t, q.open("work-b"))
	hold := t.TempDir()
	cmd := q.child([]string{"HIVE_TEST_QUEUE_HOLD=" + hold}, "sync", "--work-id", a.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Holding A")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	released := false
	release := func() {
		if !released {
			released = true
			os.WriteFile(filepath.Join(hold, "release"), nil, 0600)
		}
	}
	defer release()
	for n := 0; ; n++ {
		if _, err := os.Stat(filepath.Join(hold, "locked")); err == nil {
			break
		}
		if n > 600 {
			t.Fatal("the child never took the lock")
		}
		time.Sleep(50 * time.Millisecond)
	}

	if out, err := q.run("", "sync", "--work-id", b.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "B runs"); err != nil {
		t.Fatalf("another work was blocked: %v", err)
	} else if decodeSync(t, out).WorkID != b.WorkID {
		t.Fatal(out)
	}
	if _, err := q.run("", "sync", "--work-id", a.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "A again"); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("the same work was not busy: %v", err)
	}
	release()
	if err := <-done; err != nil {
		t.Fatalf("the holder failed: %v", err)
	}
	if stageEvents(q.session(a.WorkID), "build") != 1 || stageEvents(q.session(b.WorkID), "build") != 1 {
		t.Fatal("a stage is missing")
	}
}

// A journal and a pending intent written before the queue existed (no queue, no
// aliases, no locks folder) still resume.
func TestWorkQueuePreQueueJournalWithPendingIntentStillResumes(t *testing.T) {
	q := newQueueCLI(t)
	opened := decodeSync(t, q.open("old-journal"))
	q.killed([]string{"HIVE_TEST_QUEUE_KILL_AT=1"}, "stage", "--work-id", opened.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Start")

	dir := filepath.Join(q.f.stateDir, "publisher", "v1")
	// A journal from before the queue holds no alias; a real open now records one.
	old := q.journal()
	old.Aliases = nil
	if err := writeWorkJournal(dir, old); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "journal.json"))
	if err != nil || bytes.Contains(raw, []byte(`"queue"`)) || bytes.Contains(raw, []byte(`"aliases"`)) {
		t.Fatalf("not a pre-queue journal: %v", err)
	}
	if err = os.RemoveAll(filepath.Join(dir, "locks")); err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(dir, "journal.lock"))

	var resumed viewWorkResult
	if err = json.Unmarshal([]byte(q.must("", "resume", "--work-id", opened.WorkID)), &resumed); err != nil || resumed.WorkID != opened.WorkID {
		t.Fatalf("%+v %v", resumed, err)
	}
	if stageEvents(q.session(opened.WorkID), "build") != 1 || len(q.journal().Pending) != 0 {
		t.Fatal("the intent did not complete once")
	}
}

func fileSHA(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// The bytes of a queued capture are frozen when it is queued.
func TestWorkQueueEvidenceFreezesTheCapture(t *testing.T) {
	q := newQueueCLI(t)
	opened := decodeSync(t, q.open("frozen"))
	q.must("", "sync", "--stage", "verify", "--state", "awaiting_validation", "--reason", "Ready")
	dir := t.TempDir()
	shot := testPNG(t, dir, "screen.png", color.RGBA{R: 200, A: 255})
	original := fileSHA(t, shot)

	q.f.mustView("stop")
	queued := decodeQueued(t, q.must("", "evidence", "--stage", "verify", "--view-title", "Screen", "--image", shot+":Screen", "--candidate", "abc123", "--branch", "feature", "--base", "development"))
	if queued.WorkID != opened.WorkID {
		t.Fatalf("%+v", queued)
	}
	q.assertViewerNotStarted()
	testPNG(t, dir, "screen.png", color.RGBA{B: 200, A: 255}) // the capture changes afterwards
	if fileSHA(t, shot) == original {
		t.Fatal("the fixture did not change the capture")
	}
	j := q.journal()
	if len(j.Queue) != 1 || len(j.Queue[0].Frozen) != 1 {
		t.Fatalf("queue %+v", j.Queue)
	}
	frozen := filepath.Join(q.f.stateDir, "publisher", "v1", j.Queue[0].Frozen[0])
	if st, err := os.Stat(frozen); err != nil || st.Mode().Perm() != 0600 || fileSHA(t, frozen) != original {
		t.Fatalf("frozen copy: %v", err)
	}

	q.f.mustView("start")
	flushed := decodeSync(t, q.must("", "sync", "--stage", "verify", "--state", "awaiting_validation", "--reason", "Ready"))
	if flushed.Queue == nil || flushed.Queue.Sent != 1 {
		t.Fatalf("%+v", flushed)
	}
	work := q.session(opened.WorkID)
	if len(work.Artifacts) != 1 || work.Artifacts[0].SHA256 != original {
		t.Fatalf("artifacts %+v, want the queued bytes %s", work.Artifacts, original)
	}
	if _, err := os.Stat(frozen); !os.IsNotExist(err) {
		t.Fatal("the frozen copy was kept")
	}
}

// work evidence records the exact request of each step before sending it: a
// process killed after sending leaves it for the next run to replay.
func TestWorkEvidenceIsDurableAcrossAKill(t *testing.T) {
	q := newQueueCLI(t)
	opened := decodeSync(t, q.open("durable-evidence"))
	q.must("", "sync", "--stage", "verify", "--state", "awaiting_validation", "--reason", "Ready")
	shot := testPNG(t, t.TempDir(), "durable.png", color.RGBA{G: 200, A: 255})
	args := []string{"evidence", "--stage", "verify", "--view-title", "Durable", "--image", shot, "--candidate", "abc123", "--branch", "feature", "--base", "development"}

	// The first request is the candidate's registration.
	q.killed([]string{"HIVE_TEST_QUEUE_KILL_AT=1"}, args...)
	j := q.journal()
	if len(j.Pending) != 1 || j.Pending[0].Action != "evidence" || !j.Pending[0].Steps[0].Attempted || len(j.Pending[0].Steps[0].Body) == 0 || len(j.Pending[0].Steps[0].Response) != 0 {
		t.Fatalf("pending %+v", j.Pending)
	}
	r := decodeEvidence(t, q.must("", args...))
	if r.Result != "updated" || r.EventID == "" || len(r.ArtifactIDs) != 1 {
		t.Fatalf("%+v", r)
	}
	work := q.session(opened.WorkID)
	if len(work.Candidates) != 1 || len(work.Artifacts) != 1 || len(q.journal().Pending) != 0 {
		t.Fatalf("candidates %d artifacts %d pending %d", len(work.Candidates), len(work.Artifacts), len(q.journal().Pending))
	}
}

// The per-work locks do not protect the file of a session's current works,
// which its several works share.
func TestWorkCurrentWorkFileSurvivesConcurrentWriters(t *testing.T) {
	state := t.TempDir()
	var group sync.WaitGroup
	var ids []string
	for n := 0; n < 12; n++ {
		ids = append(ids, fmt.Sprintf("%064x", n+1))
	}
	for _, id := range ids {
		group.Add(1)
		go func() {
			defer group.Done()
			if err := addCurrentWork(state, queueHost, queueNative, id); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()
	got, err := readCurrentWorks(state, queueHost, queueNative)
	if err != nil || len(got) != len(ids) {
		t.Fatalf("%d of %d works kept: %v", len(got), len(ids), err)
	}
}

// Processes working on other works change the journal between a process's
// reads and writes; a save keeps their changes.
func TestWorkJournalSaveKeepsAnotherWorksChanges(t *testing.T) {
	q := newQueueCLI(t)
	entry := func(handle string) *viewWorkQueued {
		source := viewer.TimelineSource{ProjectRoot: q.project, Host: queueHost, NativeSessionID: queueNative}
		return &viewWorkQueued{ID: workID("op-"), Action: "sync", WorkID: handle, Source: source, State: "queued", Command: viewWorkCommand{Action: "sync", Host: queueHost, NativeSessionID: queueNative, WorkID: handle}}
	}
	command := func(handle string) viewWorkCommand {
		return viewWorkCommand{Action: "sync", WorkID: handle, Options: management.Options{Scope: "user", StateDir: q.f.stateDir}}
	}
	a, b := strings.Repeat("a", 64), strings.Repeat("b", 64)
	ea, err := openViewWorkEngine(command(a))
	if err != nil {
		t.Fatal(err)
	}
	defer ea.Close()
	eb, err := openViewWorkEngine(command(b))
	if err != nil {
		t.Fatal(err)
	}
	defer eb.Close()
	ea.journal.Queue = append(ea.journal.Queue, entry(a))
	eb.journal.Queue = append(eb.journal.Queue, entry(b))
	if err = ea.save(); err != nil {
		t.Fatal(err)
	}
	if err = eb.save(); err != nil {
		t.Fatal(err)
	}
	if got := len(q.journal().Queue); got != 2 {
		t.Fatalf("%d entries, want both", got)
	}
	ea.removeQueued(ea.journal.Queue[0].ID)
	if err = ea.save(); err != nil {
		t.Fatal(err)
	}
	if j := q.journal(); len(j.Queue) != 1 || j.Queue[0].WorkID != b {
		t.Fatalf("queue %+v", j.Queue)
	}
}
