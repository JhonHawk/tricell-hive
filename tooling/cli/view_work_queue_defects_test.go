package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// chdirToProject runs the test from inside the project, where a queued
// declaration for a work this session never opened finds its project.
func (q *queueCLI) chdirToProject() {
	q.t.Helper()
	old, err := os.Getwd()
	if err != nil {
		q.t.Fatal(err)
	}
	if err = os.Chdir(q.project); err != nil {
		q.t.Fatal(err)
	}
	q.t.Cleanup(func() { os.Chdir(old) })
}

func decodeResult(t *testing.T, out string) viewWorkResult {
	t.Helper()
	var r viewWorkResult
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("result %q: %v", out, err)
	}
	return r
}

func (q *queueCLI) lockFiles() []string {
	q.t.Helper()
	entries, err := os.ReadDir(filepath.Join(q.f.stateDir, "publisher", "v1", "locks"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		q.t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// Defect 1: re-opening a work while the viewer is down resolves to the work.
func TestWorkQueueReopenWhileViewerDownResolvesToTheOpenWork(t *testing.T) {
	q := newQueueCLI(t)
	opened := decodeResult(t, q.open("same-key"))
	q.f.mustView("stop")
	again := decodeResult(t, q.open("same-key"))
	if again.WorkID != opened.WorkID || again.Result != "opened" {
		t.Fatalf("reopen %+v, want the open work %s", again, opened.WorkID)
	}
	if got := len(q.journal().Queue); got != 0 {
		t.Fatalf("%d entries queued", got)
	}
	if got := q.workIDs(); len(got) != 1 || got[0] != opened.WorkID {
		t.Fatalf("current works %v", got)
	}
	q.f.mustView("start")
	if _, err := q.run("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "No work-id needed"); err != nil {
		t.Fatalf("a command without --work-id failed: %v", err)
	}
}

func TestWorkQueueReopenOfAQueuedOpenQueuesNothingNew(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")
	first := decodeQueued(t, q.open("queued-key"))
	second := decodeQueued(t, q.open("queued-key"))
	if second.WorkID != first.WorkID || second.OperationID != first.OperationID || len(q.journal().Queue) != 1 {
		t.Fatalf("first %+v second %+v queue %d", first, second, len(q.journal().Queue))
	}
}

// Defect 2: a flush sends only the entries its own session queued, and an
// authority rejection does not stop it.
func TestWorkQueueFlushSendsOnlyTheFlushingSessionsEntries(t *testing.T) {
	q := newQueueCLI(t)
	other := q.as("queue-stranger")
	q.chdirToProject()
	work := decodeResult(t, q.open("shared")).WorkID
	q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start")
	q.f.mustView("stop")
	for _, reason := range []string{"Stranger one", "Stranger two"} {
		decodeQueued(t, other.must("", "sync", "--work-id", work, "--stage", "build", "--state", "in_progress", "--reason", reason))
	}
	q.f.mustView("start")

	synced := decodeSync(t, q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Owner again"))
	if synced.Queue != nil {
		t.Fatalf("the owner's flush touched another session's entries: %+v", synced.Queue)
	}
	if got := len(q.journal().Queue); got != 2 {
		t.Fatalf("%d entries, want the stranger's 2", got)
	}

	_, err := other.run("", "sync", "--work-id", work, "--stage", "build", "--state", "in_progress", "--reason", "Stranger now")
	if err == nil || !strings.Contains(err.Error(), "accepted_producer_required") {
		t.Fatalf("the stranger's own command: %v", err)
	}
	if !strings.Contains(err.Error(), "queued declarations rejected") {
		t.Fatalf("the rejection of the queued entries was not reported to their session: %v", err)
	}
	if got := len(q.journal().Queue); got != 0 {
		t.Fatalf("%d reported entries left", got)
	}
}

// Defect 3: a definitive rejection during the flush archives its intent.
func TestWorkQueueDefinitiveRejectionArchivesItsIntent(t *testing.T) {
	q := newQueueCLI(t)
	work := decodeResult(t, q.open("authority")).WorkID
	q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start")
	collab := q.as("queue-collab")
	q.chdirToProject()
	s := q.session(work)
	scenario := &timelineScenario{t: t, f: q.f, input: filepath.Join(t.TempDir(), "request.json"), id: work, actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: s.ForegroundProducerID}}
	joined := scenario.result([]string{"producer", "request", work}, viewer.TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "defect-join", Source: viewer.TimelineSource{Host: queueHost, NativeSessionID: "queue-collab", ProjectRoot: q.project}, ExpectedSessionVersion: scenario.current().Version, Reason: "Accept collaborator"})
	scenario.result([]string{"producer", "accept", work}, viewer.TimelineProducerAcceptRequest{TimelineNativeOperation: scenario.base("defect-accept"), RequestID: joined.RequestID})

	q.f.mustView("stop")
	queued := decodeQueued(t, collab.must("", "sync", "--work-id", work, "--stage", "build", "--state", "in_progress", "--reason", "Collaborator closes", "--close", "completed"))
	q.f.mustView("start")

	// The collaborator may not declare stages itself: its own command fails, and
	// the failure carries the report of the queued declaration it found rejected.
	_, err := collab.run("", "sync", "--work-id", work, "--stage", "build", "--state", "in_progress", "--reason", "Collaborator continues")
	if err == nil || !strings.Contains(err.Error(), "queued declarations rejected: sync lead_required "+queued.OperationID) {
		t.Fatalf("the rejection was not reported to the session that queued it: %v", err)
	}
	j := q.journal()
	if len(j.Pending) != 0 {
		t.Fatalf("the rejected intent stayed pending: %+v", j.Pending[0])
	}
	if _, err := collab.run("", "sync", "--work-id", work, "--stage", "build", "--state", "in_progress", "--reason", "Later"); err == nil || strings.Contains(err.Error(), "pending_operation") {
		t.Fatalf("later lifecycle calls are blocked by the rejected intent: %v", err)
	}
}

func TestWorkQueueArchivesTheDefinitivelyRejectedIntentOfAFlush(t *testing.T) {
	source := viewer.TimelineSource{ProjectRoot: "/p", Host: queueHost, NativeSessionID: queueNative}
	e := &viewWorkEngine{journal: newWorkJournal()}
	i := &viewWorkIntent{ID: "pending-rejected", WorkID: "work-x", Source: source, Action: "close", State: "rejected_conflict", Steps: []viewWorkStep{{Attempted: true, Rejection: "lead_required"}}}
	e.journal.Pending = append(e.journal.Pending, i)
	e.archiveFlushRejection(i.ID, &viewWorkQueued{ID: "op-1"}, "lead_required")
	if len(e.journal.Pending) != 0 || len(e.journal.History) != 1 || e.journal.History[0].State != "archived_rejected" || !strings.Contains(e.journal.History[0].ArchiveReason, "op-1") {
		t.Fatalf("pending %d history %+v", len(e.journal.Pending), e.journal.History)
	}
}

func TestWorkQueueNeverArchivesAnUnknownOutcome(t *testing.T) {
	q := newQueueCLI(t)
	e := &viewWorkEngine{dir: filepath.Join(q.f.stateDir, "publisher", "v1"), journal: newWorkJournal()}
	source := viewer.TimelineSource{ProjectRoot: q.project, Host: queueHost, NativeSessionID: queueNative}
	i := &viewWorkIntent{ID: "pending-unknown", WorkID: "work-x", Source: source, Action: "close", State: "rejected_conflict", Steps: []viewWorkStep{{Attempted: true}}}
	e.journal.Pending = append(e.journal.Pending, i)
	e.archiveFlushRejection(i.ID, &viewWorkQueued{ID: "op-1"}, "lead_required")
	if len(e.journal.Pending) != 1 || len(e.journal.History) != 0 {
		t.Fatal("an intent with an attempted, unanswered step was archived")
	}
}

// Defect 4: lock files do not outlive the work or the session's current work.
func TestWorkLockFilesAreRemovedWithTheirWorkAndSession(t *testing.T) {
	q := newQueueCLI(t)
	a := decodeResult(t, q.open("lock-a")).WorkID
	b := decodeResult(t, q.open("lock-b")).WorkID
	q.must("", "sync", "--work-id", a, "--stage", "build", "--state", "in_progress", "--reason", "Start A")
	if len(q.lockFiles()) == 0 {
		t.Fatal("no lock files at all")
	}
	q.must("", "close", "--work-id", a, "--outcome", "completed", "--reason", "Done A")
	for _, name := range q.lockFiles() {
		if strings.HasPrefix(name, a) || strings.Contains(name, a) {
			t.Fatalf("the closed work's lock file %s remains", name)
		}
	}
	sessionLocks := func() int {
		n := 0
		for _, name := range q.lockFiles() {
			if strings.HasPrefix(name, "session-") {
				n++
			}
		}
		return n
	}
	if sessionLocks() != 1 {
		t.Fatalf("another work is current: session lock files %v", q.lockFiles())
	}
	q.must("", "close", "--work-id", b, "--outcome", "completed", "--reason", "Done B")
	if got := q.lockFiles(); len(got) != 0 {
		t.Fatalf("lock files left after the last work closed: %v", got)
	}
	if got := q.workIDs(); len(got) != 0 {
		t.Fatalf("current works %v", got)
	}
	// A command refused because the work ended leaves no lock file either.
	if _, err := q.run("", "sync", "--work-id", a, "--stage", "build", "--state", "in_progress", "--reason", "After close"); err == nil {
		t.Fatal("a closed work took a declaration")
	}
	if got := q.lockFiles(); len(got) != 0 {
		t.Fatalf("lock files left by a refused command: %v", got)
	}
}

func TestWorkLockFileIsNeverRemovedWhileHeld(t *testing.T) {
	q := newQueueCLI(t)
	dir := filepath.Join(q.f.stateDir, "publisher", "v1")
	scope := strings.Repeat("a", 64)
	if _, _, err := workPublisherDir(viewWorkCommand{Options: management.Options{Scope: "user", StateDir: q.f.stateDir}}); err != nil {
		t.Fatal(err)
	}
	l, err := lockWorkScope(dir, scope)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Release()
	removeIdleWorkLock(dir, scope)
	if _, err := os.Stat(filepath.Join(dir, "locks", scope+".lock")); err != nil {
		t.Fatalf("a held lock file was removed: %v", err)
	}
	l.Release()
	removeIdleWorkLock(dir, scope)
	if _, err := os.Stat(filepath.Join(dir, "locks", scope+".lock")); !os.IsNotExist(err) {
		t.Fatalf("an idle lock file remains: %v", err)
	}
}

// Defect 5: aliases, the per-work queue limit and the admission check.
func TestWorkAliasIsRemovedOnceNothingRefersToItsHandle(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")
	opened := decodeQueued(t, q.open("alias-key"))
	q.f.mustView("start")
	q.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Flush")
	if j := q.journal(); len(j.Queue) != 0 || len(j.Aliases) != 1 {
		t.Fatalf("queue %d aliases %+v: the handle still addresses a current work", len(j.Queue), j.Aliases)
	}
	q.must("", "close", "--outcome", "completed", "--reason", "Done")
	if j := q.journal(); len(j.Aliases) != 0 {
		t.Fatalf("aliases left after the work closed: %+v (handle %s)", j.Aliases, opened.WorkID)
	}
}

func TestWorkQueueLimitIsPerWork(t *testing.T) {
	q := newQueueCLI(t)
	q.f.mustView("stop")
	opened := decodeQueued(t, q.open("limit-a"))
	for n := 1; n < viewWorkQueueLimit; n++ {
		q.must("", "sync", "--work-id", opened.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Fill")
	}
	if _, err := q.run("", "sync", "--work-id", opened.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Over"); err == nil || !strings.Contains(err.Error(), "publisher_capacity") {
		t.Fatalf("the work's 257th entry: %v", err)
	}
	if _, err := q.run("", "open", "--project", q.project, "--key", "limit-b", "--title", "B", "--summary", "Other work"); err != nil {
		t.Fatalf("another work is limited by the first: %v", err)
	}
}

func TestWorkAdmissionReadsTheJournalUnderItsLock(t *testing.T) {
	q := newQueueCLI(t)
	dir := filepath.Join(q.f.stateDir, "publisher", "v1")
	source := viewer.TimelineSource{ProjectRoot: q.project, Host: queueHost, NativeSessionID: queueNative}
	if _, _, err := workPublisherDir(viewWorkCommand{Options: management.Options{Scope: "user", StateDir: q.f.stateDir}}); err != nil {
		t.Fatal(err)
	}
	first := &viewWorkEngine{state: q.f.stateDir, dir: dir}
	second := &viewWorkEngine{state: q.f.stateDir, dir: dir}
	if err := first.load(); err != nil {
		t.Fatal(err)
	}
	if err := second.load(); err != nil {
		t.Fatal(err)
	}
	if _, err := first.prepareIntent("close", "work-one", source, map[string]string{"n": "1"}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if _, err := second.prepareIntent("close", "work-two", source, map[string]string{"n": "2"}); err == nil || !strings.Contains(err.Error(), "publisher_capacity") {
		t.Fatalf("a second process reserved beyond the limit from a stale journal: %v", err)
	}
	if j, err := readWorkJournalChecked(dir, false); err != nil || len(j.Pending) != 1 {
		t.Fatalf("pending intents: %v", err)
	}
}
