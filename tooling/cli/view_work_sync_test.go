package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// syncCLI runs one `hive view work` command the way an agent does: parsed from
// its arguments, with the identity and the current work resolved, against the
// fixture's viewer.
type syncCLI struct {
	t        *testing.T
	f        *viewFixture
	project  string
	commands int
}

func newSyncCLI(t *testing.T) *syncCLI {
	f, c := workEngineFixture(t)
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "sync-session-one"})
	return &syncCLI{t: t, f: f, project: c.Project}
}

func (s *syncCLI) run(stdin string, args ...string) (string, error) {
	s.t.Helper()
	s.commands++
	var out bytes.Buffer
	err := runViewWork(append(args, "--state-dir", s.f.stateDir), &out, strings.NewReader(stdin))
	return out.String(), err
}

func (s *syncCLI) must(stdin string, args ...string) string {
	s.t.Helper()
	out, err := s.run(stdin, args...)
	if err != nil {
		s.t.Fatalf("hive view work %s: %v", strings.Join(args, " "), err)
	}
	if len(out) > 4096 {
		s.t.Fatalf("receipt of %d bytes exceeds the bounded output", len(out))
	}
	return out
}

func (s *syncCLI) session(id string) viewer.TimelineSession {
	s.t.Helper()
	r, err := viewer.DecodeTimelineSessionResponse([]byte(s.f.mustView("session", "get", id)))
	if err != nil {
		s.t.Fatal(err)
	}
	return r.Session
}

func decodeSync(t *testing.T, out string) viewWorkSyncResult {
	t.Helper()
	var r viewWorkSyncResult
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("receipt %q: %v", out, err)
	}
	return r
}

const syncThreeTasks = `[{"key":"T1","title":"Parse the input","state":"in_progress"},{"key":"T2","title":"Write the output"},{"key":"T3","title":"Cover with tests"}]`
const syncThreeDone = `[{"key":"T1","title":"Parse the input","state":"completed"},{"key":"T2","title":"Write the output","state":"completed"},{"key":"T3","title":"Cover with tests","state":"completed"}]`

// AC1: a build without a plan publishes its stage, three tasks with their
// states and the close, with the final read included, in three commands and
// with no binding error.
func TestWorkSyncBuildWithoutPlanInThreeCommands(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "no-plan-build", "--title", "Build without a plan", "--summary", "Three tasks, no change folder"))
	if open.Result != "opened" || open.WorkID == "" {
		t.Fatalf("open: %+v", open)
	}
	start := decodeSync(t, s.must(syncThreeTasks, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build", "--tasks", "-"))
	if start.Result != "updated" || start.StageID == "" || len(start.Tasks) != 3 {
		t.Fatalf("first sync: %+v", start)
	}
	finish := decodeSync(t, s.must(syncThreeDone, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Finish the build", "--tasks", "-", "--close", "completed"))
	if finish.Result != "updated" || finish.State != "closed" || finish.WorkID != open.WorkID {
		t.Fatalf("closing sync: %+v", finish)
	}
	if s.commands != 3 {
		t.Fatalf("%d commands, want 3", s.commands)
	}

	// What the viewer shows: the closed work, the stage and three tasks bound to it.
	work := s.session(open.WorkID)
	if work.State != viewer.TimelineClosed || len(work.Tasks) != 3 || work.CurrentStageEventID != nil {
		t.Fatalf("work %s, %d tasks, current stage %v", work.State, len(work.Tasks), work.CurrentStageEventID)
	}
	stage := workLifecycleStage(work, "build")
	if stage == nil || stage.StageState != "completed" {
		t.Fatalf("stage %+v", stage)
	}
	for _, task := range work.Tasks {
		seg := task.Segments[0]
		if seg.State != "completed" || seg.DisplayStageEventID != stage.EventID {
			t.Fatalf("task %s: %s bound to %q", task.TaskKey, seg.State, seg.DisplayStageEventID)
		}
	}
	// The receipt names every task and segment.
	for _, task := range finish.Tasks {
		stored := workLifecycleTask(work, task.Key)
		if stored == nil || stored.TaskID != task.TaskID || stored.Segments[0].SegmentID != task.SegmentID || task.State != "completed" {
			t.Fatalf("receipt %+v does not match the viewer", task)
		}
	}
	// The work is no longer current: a command without --work-id has none to act on.
	if _, err := s.run("", "close", "--outcome", "completed", "--reason", "Again"); err == nil || !strings.Contains(err.Error(), "work-id required: this session has no open work") {
		t.Fatalf("a closed work stayed current: %v", err)
	}
}

func TestWorkSyncRepeatedDeclarationIsUnchanged(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "repeat", "--title", "Repeat", "--summary", "Idempotent sync"))
	args := []string{"sync", "--stage", "build", "--state", "in_progress", "--reason", "Start the build", "--tasks", "-"}
	first := decodeSync(t, s.must(syncThreeTasks, args...))
	again := decodeSync(t, s.must(syncThreeTasks, args...))
	if again.Result != "unchanged" || again.SessionVersion != first.SessionVersion || len(again.Tasks) != 3 || again.StageID != first.StageID {
		t.Fatalf("repeat: %+v after %+v", again, first)
	}
	// The viewer saw nothing new either.
	if got := s.session(open.WorkID).Version; got != first.SessionVersion {
		t.Fatalf("version moved from %d to %d", first.SessionVersion, got)
	}
}

func TestWorkSyncAddsTasksLaterAndWalksThroughStates(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "later", "--title", "Later", "--summary", "Tasks arrive in turns"))
	s.must(`[{"key":"T1","title":"First"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start", "--tasks", "-")
	// T1 completes directly from not_started (the sync walks through in_progress); T2 arrives now.
	r := decodeSync(t, s.must(`[{"key":"T1","title":"First","state":"completed"},{"key":"T2","title":"Second","state":"in_review"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Progress", "--tasks", "-"))
	if r.Result != "updated" || len(r.Tasks) != 2 || r.Tasks[0].State != "completed" || r.Tasks[1].State != "in_review" {
		t.Fatalf("%+v", r)
	}
	work := s.session(open.WorkID)
	if len(work.Tasks) != 2 || work.Tasks[0].PlanOrder == work.Tasks[1].PlanOrder {
		t.Fatalf("tasks %+v", work.Tasks)
	}
	// A task blocked with its reason, and resumed with a resolution.
	blocked := `[{"key":"T2","title":"Second","state":"blocked","block":{"reason":"Waiting for access","steps":["Ask the owner"],"owner":"Maintainer","resume_when":"Access granted"}}]`
	if b := decodeSync(t, s.must(blocked, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Blocked", "--tasks", "-")); b.Tasks[0].State != "blocked" {
		t.Fatalf("%+v", b)
	}
	resumed := `[{"key":"T2","title":"Second","state":"completed","resolution":"Access granted by the owner"}]`
	if b := decodeSync(t, s.must(resumed, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Resumed", "--tasks", "-")); b.Tasks[0].State != "completed" {
		t.Fatalf("%+v", b)
	}
}

// A registered task moves with its key and state alone, as the publication
// reference tells agents; only a task sync has to create needs a title, and a
// declaration that would create one without it writes nothing.
func TestWorkSyncMovesARegisteredTaskByKeyAndState(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "by-key", "--title", "By key", "--summary", "State updates name the task by key"))
	s.must(syncThreeTasks, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start", "--tasks", "-")
	r := decodeSync(t, s.must(`[{"key":"T1","state":"completed"},{"key":"T2","state":"in_progress"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "Progress", "--tasks", "-"))
	if r.Result != "updated" || len(r.Tasks) != 2 || r.Tasks[0].State != "completed" || r.Tasks[1].State != "in_progress" {
		t.Fatalf("%+v", r)
	}
	work := s.session(open.WorkID)
	if got := workLifecycleTask(work, "T1"); got == nil || got.Title != "Parse the input" || got.Segments[0].State != "completed" {
		t.Fatalf("T1 %+v", got)
	}
	before := work.Version
	_, err := s.run(`[{"key":"T1","state":"completed"},{"key":"T4","state":"in_progress"}]`, "sync", "--stage", "build", "--state", "in_progress", "--reason", "New task", "--tasks", "-")
	if err == nil || !strings.Contains(err.Error(), "task_title_required") {
		t.Fatalf("a new task without a title: %v", err)
	}
	if after := s.session(open.WorkID); after.Version != before || len(after.Tasks) != 3 {
		t.Fatalf("the rejected declaration wrote: version %d -> %d, %d tasks", before, after.Version, len(after.Tasks))
	}
}

// A sync on a closed work reports the closed work, not the title a task it
// names lacks, so the agent fixes the real condition first.
func TestWorkSyncOnAClosedWorkReportsItBeforeAMissingTitle(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "closed-first", "--title", "Closed first", "--summary", "Error order"))
	s.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Start", "--close", "completed")
	_, err := s.run(`[{"key":"T5","state":"in_progress"}]`, "sync", "--work-id", open.WorkID, "--stage", "build", "--state", "in_progress", "--reason", "Late", "--previous-stage", "build", "--tasks", "-")
	if err == nil || !strings.Contains(err.Error(), "session_closed") {
		t.Fatalf("got %v, want session_closed", err)
	}
}

// show acts on the session's current work when --work-id is omitted, as the
// publication reference says.
func TestWorkShowUsesTheCurrentWork(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "show-current", "--title", "Show current", "--summary", "No work ID"))
	out := s.must("", "show")
	if !strings.Contains(out, open.WorkID) {
		t.Fatalf("show without --work-id: %.200s", out)
	}
}

func TestWorkSyncRejectsInvalidDeclarationsOffline(t *testing.T) {
	id := strings.Repeat("a", 64)
	base := []string{"sync", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "build", "--state", "in_progress", "--reason", "Sync", "--tasks", "-"}
	for _, tc := range []struct{ name, raw, want string }{
		{"empty", `[]`, "tasks: invalid task count"},
		{"duplicate key", `[{"key":"T1","title":"A"},{"key":"T1","title":"B"}]`, "tasks[1].key: duplicate task key"},
		{"key zero", `[{"key":"T0","title":"A"}]`, "tasks[0].key"},
		{"state", `[{"key":"T1","title":"A","state":"done"}]`, "tasks[0].state: unsupported state"},
		{"blocked without block", `[{"key":"T1","title":"A","state":"blocked"}]`, "tasks[0]"},
		{"unknown field", `[{"key":"T1","title":"A","order":1}]`, "tasks[0].order: unknown field"},
		{"empty title", `[{"key":"T1","title":" "}]`, "tasks[0].title"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseViewWorkCommand(base, strings.NewReader(tc.raw))
			if err == nil || !strings.Contains(err.Error(), "invalid semantic JSON: "+tc.want) && !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %v, want %q", err, tc.want)
			}
		})
	}
	for _, args := range [][]string{
		{"sync", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "build", "--state", "in_progress"},
		{"sync", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "deploy", "--state", "in_progress", "--reason", "x"},
		{"sync", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "build", "--state", "in_progress", "--reason", "x", "--close", "done"},
	} {
		if _, err := parseViewWorkCommand(args, strings.NewReader("")); err == nil {
			t.Fatalf("%v accepted", args)
		}
	}
}

func TestWorkOpenWithoutChangeFolder(t *testing.T) {
	s := newSyncCLI(t)
	// The project defaults to the repository the command runs in.
	t.Chdir(filepath.Join(s.project))
	first := decodeSync(t, s.must("", "open", "--key", "default-project", "--title", "Default project", "--summary", "Runs where it is"))
	again := decodeSync(t, s.must("", "open", "--key", "default-project", "--title", "Edited title", "--summary", "Runs where it is"))
	if first.WorkID == "" || again.WorkID != first.WorkID || again.SessionVersion != first.SessionVersion {
		t.Fatalf("reopen: %+v then %+v", first, again)
	}
	other := decodeSync(t, s.must("", "open", "--key", "second", "--title", "Second", "--summary", "A second work in the same session"))
	if other.WorkID == first.WorkID {
		t.Fatal("two keys collapsed")
	}
	// The journal that was written is still readable, with no binding for either.
	e, err := openViewWorkEngine(viewWorkCommand{Options: management.Options{Scope: "user", StateDir: s.f.stateDir}})
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if len(e.journal.Bindings) != 0 {
		t.Fatalf("a work without a change folder was bound: %+v", e.journal.Bindings)
	}
	// With two current works a command must name one.
	if _, err := s.run("", "close", "--outcome", "completed", "--reason", "Done"); err == nil || !strings.Contains(err.Error(), "work-id required") {
		t.Fatalf("close with two current works: %v", err)
	}
}

func TestWorkTasksRegisterNeedsAPlanOnlyWithAChangeFolder(t *testing.T) {
	f, c := workEngineFixture(t)
	register := func(workID string) error {
		var out bytes.Buffer
		args := []string{"tasks", "register", "--work-id", workID, "--host", c.Host, "--native-session-id", c.NativeSessionID, "--state-dir", f.stateDir, "--json", "-"}
		return runViewWork(args, &out, strings.NewReader(`[{"key":"T1","title":"Task","summary":"Task","order":1,"segments":[{"stage":"build","order":1}]}]`))
	}
	withFolder := engineWorkOpen(t, c)
	if err := register(withFolder.WorkID); err == nil || !strings.Contains(err.Error(), "confirmed_plan_revision_required") {
		t.Fatalf("with a change folder: %v", err)
	}
	c.Change, c.Key, c.Title = "", "no-folder", "No folder"
	without := engineWorkOpen(t, c)
	if err := register(without.WorkID); err != nil {
		t.Fatalf("without a change folder: %v", err)
	}
}

func TestWorkCloseCompletesTheActiveStageAndReadsBack(t *testing.T) {
	s := newSyncCLI(t)
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "close", "--title", "Close", "--summary", "Close completes the stage"))
	s.must("", "sync", "--stage", "build", "--state", "in_review", "--reason", "Review the build")
	closed := decodeSync(t, s.must("", "close", "--outcome", "completed", "--reason", "Integrated and verified"))
	if closed.State != "closed" || closed.Result != "updated" {
		t.Fatalf("%+v", closed)
	}
	work := s.session(open.WorkID)
	if work.State != viewer.TimelineClosed || workLifecycleStage(work, "build").StageState != "completed" {
		t.Fatalf("work %s, stage %+v", work.State, workLifecycleStage(work, "build"))
	}
	// The same declaration with --close, repeated after the close, changes nothing.
	again, err := s.run("", "sync", "--work-id", open.WorkID, "--stage", "build", "--state", "completed", "--reason", "Review the build", "--close", "completed")
	if err != nil {
		t.Fatal(err)
	}
	if r := decodeSync(t, again); r.Result != "unchanged" || r.State != "closed" {
		t.Fatalf("%+v", r)
	}
}

// testPNG writes a small PNG whose bytes depend on its color.
func testPNG(t *testing.T, dir, name string, c color.RGBA) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for x := 0; x < 8; x++ {
		for y := 0; y < 8; y++ {
			img.Set(x, y, c)
		}
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, buffer.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func decodeEvidence(t *testing.T, out string) viewWorkEvidenceResult {
	t.Helper()
	var r viewWorkEvidenceResult
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("receipt %q: %v", out, err)
	}
	return r
}

// AC7: two captures, their candidate and a view that supersedes an earlier one,
// each in one command.
func TestWorkEvidencePublishesCapturesCandidateAndSupersedingView(t *testing.T) {
	s := newSyncCLI(t)
	dir := t.TempDir()
	profile := testPNG(t, dir, "profile.png", color.RGBA{R: 200, A: 255})
	chat := testPNG(t, dir, "chat.png", color.RGBA{G: 200, A: 255})
	fixed := testPNG(t, dir, "chat-fixed.png", color.RGBA{B: 200, A: 255})
	open := decodeSync(t, s.must("", "open", "--project", s.project, "--key", "evidence", "--title", "Evidence", "--summary", "Captures for validation"))
	s.must("", "sync", "--stage", "verify", "--state", "awaiting_validation", "--reason", "Ready for validation")
	// Evidence explicitly targets verify while build owns the foreground.
	s.must("", "sync", "--stage", "build", "--state", "in_progress", "--reason", "Parallel build")

	// Without a candidate there is nothing to attach the captures to.
	if _, err := s.run("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", chat); err == nil || !strings.Contains(err.Error(), "candidate_required") {
		t.Fatalf("no candidate: %v", err)
	}
	if _, err := s.run("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", filepath.Join(dir, "missing.png"), "--candidate", "abc123", "--branch", "feature", "--base", "development"); err == nil || !strings.Contains(err.Error(), "image_unreadable") {
		t.Fatalf("missing image: %v", err)
	}

	first := decodeEvidence(t, s.must("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", profile+":Profile", "--image", chat, "--candidate", "abc123", "--branch", "feature", "--base", "development", "--step", "Open the chat", "--expected", "The filter is compact", "--viewport", "mobile", "--theme", "dark"))
	if first.Result != "updated" || first.CandidateID == "" || first.EventID == "" || len(first.ViewIDs) != 1 || len(first.ArtifactIDs) != 2 {
		t.Fatalf("first evidence: %+v", first)
	}
	work := s.session(open.WorkID)
	if len(work.Candidates) != 1 || work.Candidates[0].Commit != "abc123" || work.Candidates[0].Branch != "feature" || len(work.Artifacts) != 2 {
		t.Fatalf("candidates %+v artifacts %d", work.Candidates, len(work.Artifacts))
	}
	for _, a := range work.Artifacts {
		if a.CandidateID != first.CandidateID || a.Viewport != "mobile" || a.Theme != "dark" {
			t.Fatalf("artifact %+v", a)
		}
	}
	var event viewer.TimelineEvent
	for _, e := range work.Events {
		if e.EventID == first.EventID {
			event = e
		}
	}
	if event.StageEventID != workLifecycleStage(work, "verify").EventID {
		t.Fatalf("evidence attached to wrong stage: %+v", event)
	}
	if event.Kind != "validation" || len(event.Views) != 1 || len(event.Views[0].ImageRefs) != 2 || event.Views[0].ImageRefs[0].Caption != "Profile" || event.Views[0].ImageRefs[1].Caption != "chat" {
		t.Fatalf("event %+v", event)
	}

	// The same command again uploads and appends nothing.
	again := decodeEvidence(t, s.must("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", profile+":Profile", "--image", chat, "--candidate", "abc123", "--viewport", "mobile", "--theme", "dark"))
	if again.Result != "unchanged" || again.EventID != first.EventID || s.session(open.WorkID).Version != work.Version {
		t.Fatalf("repeat: %+v", again)
	}

	// The store preserves the first capture metadata for identical bytes.
	beforeMetadata := s.session(open.WorkID).Version
	if _, err := s.run("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", profile+":Profile", "--image", chat, "--candidate", "abc123", "--viewport", "desktop", "--theme", "light"); err == nil || !strings.Contains(err.Error(), "capture_metadata_conflict") {
		t.Fatalf("contradictory metadata: %v", err)
	}
	if s.session(open.WorkID).Version != beforeMetadata {
		t.Fatal("metadata conflict changed the stored work")
	}

	// A corrected capture of the Chat view, on a new candidate, supersedes it.
	second := decodeEvidence(t, s.must("", "evidence", "--stage", "verify", "--view-title", "Chat corrected", "--image", fixed, "--candidate", "def456", "--branch", "feature", "--base", "development", "--supersedes", first.ViewIDs[0], "--viewport", "mobile", "--theme", "dark"))
	if second.Result != "updated" || second.CandidateID == first.CandidateID || second.EventID == first.EventID {
		t.Fatalf("second evidence: %+v", second)
	}
	work = s.session(open.WorkID)
	var iteration viewer.TimelineEvent
	for _, e := range work.Events {
		if e.EventID == second.EventID {
			iteration = e
		}
	}
	v := iteration.Views[0]
	if iteration.Kind != "iteration" || iteration.RelatedEventID != first.EventID || v.Supersedes == nil || v.Supersedes.ViewID != first.ViewIDs[0] || v.Supersedes.EventID != first.EventID || len(v.ImageRefs) != 1 {
		t.Fatalf("iteration %+v", iteration)
	}
	if _, err := s.run("", "evidence", "--stage", "verify", "--view-title", "Chat", "--image", fixed, "--supersedes", "no-such-view"); err == nil || !strings.Contains(err.Error(), "supersedes_view_not_found") {
		t.Fatalf("unknown view: %v", err)
	}
}

func TestWorkEvidenceRejectsBadInputOffline(t *testing.T) {
	id := strings.Repeat("a", 64)
	base := []string{"evidence", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "verify"}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"no image", append(base, "--view-title", "Chat")},
		{"no title", append(base, "--image", "a.png")},
		{"bad stage", []string{"evidence", "--work-id", id, "--host", "codex", "--native-session-id", "s1", "--stage", "deploy", "--view-title", "Chat", "--image", "a.png"}},
		{"bad viewport", append(base, "--view-title", "Chat", "--image", "a.png", "--viewport", "tablet")},
		{"bad theme", append(base, "--view-title", "Chat", "--image", "a.png", "--theme", "sepia")},
		{"bad supersedes", append(base, "--view-title", "Chat", "--image", "a.png", "--supersedes", "a b")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := parseViewWorkCommand(tc.args, strings.NewReader("")); err == nil {
				t.Fatal("accepted")
			}
		})
	}
	if _, err := parseViewWorkCommand(append(base, "--view-title", "Chat", "--image", "a.png", "--image", "b.png"), strings.NewReader("")); err != nil {
		t.Fatalf("repeated --image rejected: %v", err)
	}
}

func TestViewWorkContractDocumentsTheSemanticCommands(t *testing.T) {
	for _, want := range []string{"open --key KEY", "sync --stage STAGE", "evidence --stage STAGE", "--supersedes VIEW", "tasks_omitted", "Advanced commands"} {
		if !strings.Contains(viewWorkContract, want) {
			t.Errorf("contract lacks %q", want)
		}
	}
	if len(viewWorkContract) > 9000 {
		t.Errorf("contract is %d bytes: shorten the advanced part, not the new commands", len(viewWorkContract))
	}
}
