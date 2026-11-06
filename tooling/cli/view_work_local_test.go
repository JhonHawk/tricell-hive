package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

// Local management commands run against a real viewer handler in a temporary
// state directory, never against the user's viewer.

func (f *workPublishFixture) publishWork(t *testing.T) viewWorkResult {
	t.Helper()
	r, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan")
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// local runs a management command, which carries no native producer identity.
func (f *workPublishFixture) local(args ...string) (viewWorkResult, string, error) {
	var out bytes.Buffer
	err := runViewWork(append(args, "--state-dir", f.state), &out, strings.NewReader(""))
	var result viewWorkResult
	if err == nil {
		json.Unmarshal(out.Bytes(), &result)
	}
	return result, out.String(), err
}

func (f *workPublishFixture) sentActor(t *testing.T, key string) map[string]any {
	t.Helper()
	f.bodyMu.Lock()
	raw := f.bodies[key]
	f.bodyMu.Unlock()
	var body struct {
		Actor map[string]any `json:"actor"`
	}
	if err := json.Unmarshal(raw, &body); err != nil || body.Actor == nil {
		t.Fatalf("no actor sent for %s: %q", key, raw)
	}
	return body.Actor
}

func requireCLIActor(t *testing.T, actor map[string]any) {
	t.Helper()
	if len(actor) != 2 || actor["kind"] != "local_viewer" || actor["origin"] != "cli" {
		t.Fatalf("actor is not local_viewer with origin cli and no producer: %v", actor)
	}
}

func TestWorkLocalRenameArchivePinCycle(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	path := "PATCH /api/timeline/sessions/" + work.WorkID

	renamed, _, err := f.local("rename", "--work-id", work.WorkID, "--title", "Renamed from the CLI")
	if err != nil || renamed.Result != "renamed" || renamed.WorkID != work.WorkID {
		t.Fatalf("rename: %+v %v", renamed, err)
	}
	requireCLIActor(t, f.sentActor(t, path))
	s := f.observe(t, work.WorkID)
	if s.Title != "Renamed from the CLI" || s.Version != renamed.SessionVersion {
		t.Fatalf("title not changed or version mismatch: %q v%d / v%d", s.Title, s.Version, renamed.SessionVersion)
	}
	origin := ""
	for _, e := range s.Events {
		if e.Actor.Kind == "local_viewer" {
			origin = e.Actor.Origin
		}
	}
	if origin != "cli" {
		t.Fatalf("rename event does not record origin cli: %q", origin)
	}

	for _, step := range []struct {
		command, result  string
		archived, pinned bool
	}{
		{"archive", "archived", true, false},
		{"restore", "restored", false, false},
		{"pin", "pinned", false, true},
		{"unpin", "unpinned", false, false},
	} {
		r, _, err := f.local(step.command, "--work-id", work.WorkID)
		if err != nil || r.Result != step.result {
			t.Fatalf("%s: %+v %v", step.command, r, err)
		}
		requireCLIActor(t, f.sentActor(t, path))
		s = f.observe(t, work.WorkID)
		if s.Archived != step.archived || s.Pinned != step.pinned {
			t.Fatalf("%s: archived=%v pinned=%v", step.command, s.Archived, s.Pinned)
		}
	}
}

func TestWorkLocalRecoverMakesRequestLeadAndItClosesTheWork(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	current := f.observe(t, work.WorkID)
	join, _ := json.Marshal(viewer.TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "recover-join", Source: viewer.TimelineSource{Host: "claude-code", NativeSessionID: "new-session", ProjectRoot: f.root}, ExpectedSessionVersion: current.Version, Reason: "Continue the work"})
	raw, err := viewTimelineHTTP(f.info, "POST", "/api/timeline/sessions/"+work.WorkID+"/join-requests", "application/json", bytes.NewReader(join), viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	joined, err := viewer.DecodeTimelineOperationResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	before := f.observe(t, work.WorkID).ForegroundProducerID

	r, _, err := f.local("recover", "--work-id", work.WorkID, "--request-id", joined.RequestID, "--reason", "The leading session ended")
	if err != nil || r.Result != "recovered" {
		t.Fatalf("recover: %+v %v", r, err)
	}
	requireCLIActor(t, f.sentActor(t, "POST /api/timeline/sessions/"+work.WorkID+"/recover-lead"))
	after := f.observe(t, work.WorkID)
	if after.ForegroundProducerID == before || len(after.PendingJoinRequests) != 0 {
		t.Fatalf("lead not reassigned in one commit: %q -> %q, pending %d", before, after.ForegroundProducerID, len(after.PendingJoinRequests))
	}
	origin, reason := "", ""
	for _, e := range after.Events {
		if e.Audit != nil && e.Actor.Kind == "local_viewer" {
			origin, reason = e.Actor.Origin, e.Audit.Reason
		}
	}
	if origin != "cli" || reason != "The leading session ended" {
		t.Fatalf("audit origin/reason: %q %q", origin, reason)
	}

	newLead := []string{"--host", "claude-code", "--native-session-id", "new-session", "--state-dir", f.state}
	var out bytes.Buffer
	if err = runViewWork(append([]string{"stage", "--work-id", work.WorkID, "--stage", "plan", "--state", "completed", "--reason", "Plan done"}, newLead...), &out, nil); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err = runViewWork(append([]string{"close", "--work-id", work.WorkID, "--outcome", "completed", "--reason", "Closed by the recovered lead"}, newLead...), &out, nil); err != nil {
		t.Fatal(err)
	}
	if f.observe(t, work.WorkID).State != viewer.TimelineClosed {
		t.Fatal("recovered lead did not close the work")
	}
}

func TestWorkLocalDeleteNeedsConfirmationEqualToWorkID(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	mutations := f.mutations.Load()
	for name, args := range map[string][]string{
		"missing":  {"delete", "--work-id", work.WorkID},
		"mismatch": {"delete", "--work-id", work.WorkID, "--confirm", strings.Repeat("b", 64)},
	} {
		if _, out, err := f.local(args...); err == nil || out != "" || !strings.Contains(err.Error(), "invalid_input") {
			t.Fatalf("%s confirmation: %q %v", name, out, err)
		}
	}
	if f.mutations.Load() != mutations || f.observe(t, work.WorkID).Version != work.SessionVersion {
		t.Fatal("unconfirmed delete sent or changed something")
	}

	r, _, err := f.local("delete", "--work-id", work.WorkID, "--confirm", work.WorkID)
	if err != nil || r.Result != "deleted" || r.WorkID != work.WorkID {
		t.Fatalf("delete: %+v %v", r, err)
	}
	requireCLIActor(t, f.sentActor(t, "DELETE /api/timeline/sessions/"+work.WorkID))
	index, err := viewIndex(f.info)
	if err != nil || len(index.TimelineSessions) != 0 {
		t.Fatalf("work still listed: %v", err)
	}
}

func TestWorkLocalOlderViewerReceivesNothing(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	f.hideCapability.Store(true)
	mutations := f.mutations.Load()
	for _, args := range [][]string{
		{"rename", "--work-id", work.WorkID, "--title", "Never sent"},
		{"archive", "--work-id", work.WorkID},
		{"pin", "--work-id", work.WorkID},
		{"recover", "--work-id", work.WorkID, "--request-id", "request-one", "--reason", "Never sent"},
		{"delete", "--work-id", work.WorkID, "--confirm", work.WorkID},
	} {
		_, out, err := f.local(args...)
		if err == nil || out != "" || !strings.Contains(err.Error(), "viewer_update_required") || !strings.Contains(err.Error(), "the user") || strings.Contains(err.Error(), "hive view stop") {
			t.Fatalf("%s on an older viewer: %q %v", args[0], out, err)
		}
	}
	if f.mutations.Load() != mutations || f.observe(t, work.WorkID).Title != "Plan export" {
		t.Fatal("an older viewer received a write")
	}
}

func TestWorkLocalUncertainResponseReplaysTheSameRequest(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)

	// Response lost after commit: the replay returns the stored receipt.
	f.dropLocal.Store("after")
	r, _, err := f.local("rename", "--work-id", work.WorkID, "--title", "Committed, response lost")
	if err != nil || r.Result != "renamed" || r.SessionVersion != f.observe(t, work.WorkID).Version {
		t.Fatalf("lost response after commit not settled by replay: %+v %v", r, err)
	}

	// Request lost before the viewer saw it: the replay applies it once.
	f.dropLocal.Store("before")
	r, _, err = f.local("rename", "--work-id", work.WorkID, "--title", "Arrived on replay")
	if err != nil || f.observe(t, work.WorkID).Title != "Arrived on replay" {
		t.Fatalf("lost request not applied by the replay: %+v %v", r, err)
	}

	// Both attempts lost before arriving: re-read, nothing changed.
	f.dropLocal.Store("before-always")
	version := f.observe(t, work.WorkID).Version
	_, out, err := f.local("rename", "--work-id", work.WorkID, "--title", "Never arrived")
	f.dropLocal.Store("")
	if err == nil || out != "" || !strings.Contains(err.Error(), "not_applied") {
		t.Fatalf("lost request reported as applied: %q %v", out, err)
	}
	if s := f.observe(t, work.WorkID); s.Title != "Arrived on replay" || s.Version != version {
		t.Fatal("work changed although no request arrived")
	}

	// A delete whose response is lost is answered by the stored receipt.
	f.dropLocal.Store("after")
	r, _, err = f.local("delete", "--work-id", work.WorkID, "--confirm", work.WorkID)
	if err != nil || r.Result != "deleted" || r.SessionVersion < 1 {
		t.Fatalf("lost delete response not settled by replay: %+v %v", r, err)
	}
}

func TestWorkLocalUnansweredDeleteReportsGoneWithoutAttribution(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	f.dropLocal.Store("after-always")
	_, out, err := f.local("delete", "--work-id", work.WorkID, "--confirm", work.WorkID)
	f.dropLocal.Store("")
	if err == nil || out != "" || !strings.Contains(err.Error(), "work_gone") || strings.Contains(err.Error(), "session_version") {
		t.Fatalf("unanswered delete attributed or reported as success: %q %v", out, err)
	}
}

func (f *workPublishFixture) read(id string) (viewer.TimelineSession, string, error) {
	return viewWorkLocal{info: f.info}.read(id)
}

func TestWorkLocalNonDefinitiveFailureReplaysInsteadOfClaimingNothingChanged(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	f.dropLocal.Store("503")
	r, _, err := f.local("rename", "--work-id", work.WorkID, "--title", "Committed, then 503")
	if err != nil || r.Result != "renamed" || f.observe(t, work.WorkID).Title != "Committed, then 503" {
		t.Fatalf("503 outcome_unknown after commit not settled: %+v %v", r, err)
	}
	// delete_pending: the replay finishes the cleanup and returns the receipt.
	f.dropLocal.Store("pending")
	r, _, err = f.local("delete", "--work-id", work.WorkID, "--confirm", work.WorkID)
	if err != nil || r.Result != "deleted" || r.SessionVersion < 1 {
		t.Fatalf("delete_pending not settled by replay: %+v %v", r, err)
	}
}

func TestWorkLocalDefinitiveRejectionShowsTheViewerCode(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	// Any 4xx with a contract code is definitive, whether or not the code is one
	// the publisher journal lists.
	f.dropLocal.Store("reject")
	_, out, err := f.local("rename", "--work-id", work.WorkID, "--title", "Conflicting")
	if err == nil || out != "" || !strings.Contains(err.Error(), "invalid_text") || !strings.Contains(err.Error(), "nothing was changed") {
		t.Fatalf("4xx contract code not shown as a definitive rejection: %q %v", out, err)
	}
	if f.observe(t, work.WorkID).Title != "Plan export" {
		t.Fatal("rejected request changed the work")
	}
}

func TestWorkLocalRenameWithPaddedTitleMatchesTheStoredTitle(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	f.dropLocal.Store("after")
	r, _, err := f.local("rename", "--work-id", work.WorkID, "--title", "  Padded title  ")
	if err != nil || r.Result != "renamed" {
		t.Fatalf("padded title not confirmed by readback: %+v %v", r, err)
	}
	if f.observe(t, work.WorkID).Title != "Padded title" {
		t.Fatal("title not stored trimmed")
	}
}

func TestWorkLocalReadsHealthOncePerCommand(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := f.publishWork(t)
	before := f.healthReads.Load()
	if _, _, err := f.local("rename", "--work-id", work.WorkID, "--title", "One health read"); err != nil {
		t.Fatal(err)
	}
	if reads := f.healthReads.Load() - before; reads != 1 {
		t.Fatalf("health read %d times for one command", reads)
	}
}

func TestWorkLocalRejectsInvalidInputOffline(t *testing.T) {
	id := strings.Repeat("a", 64)
	for name, args := range map[string][]string{
		"rename without title":        {"rename", "--work-id", id},
		"rename with blank title":     {"rename", "--work-id", id, "--title", "  "},
		"rename without work":         {"rename", "--title", "T"},
		"archive with native source":  {"archive", "--work-id", id, "--host", "codex", "--native-session-id", "s"},
		"archive with reason":         {"archive", "--work-id", id, "--reason", "no reason belongs here"},
		"recover without reason":      {"recover", "--work-id", id, "--request-id", "request-one"},
		"recover without request":     {"recover", "--work-id", id, "--reason", "r"},
		"delete with reason":          {"delete", "--work-id", id, "--confirm", id, "--reason", "r"},
		"unpin with confirm":          {"unpin", "--work-id", id, "--confirm", id},
		"invalid work id":             {"pin", "--work-id", "short"},
		"duplicated confirm":          {"delete", "--work-id", id, "--confirm", id, "--confirm", id},
		"restore with positional arg": {"restore", "--work-id", id, "extra"},
	} {
		var out bytes.Buffer
		err := runViewWork(append(args, "--state-dir", t.TempDir()), &out, nil)
		if err == nil || !strings.HasPrefix(err.Error(), "invalid_input:") || out.Len() != 0 {
			t.Fatalf("%s: expected offline rejection, got %q %v", name, out.String(), err)
		}
	}
}

func TestWorkLocalContractDescribesTheCommands(t *testing.T) {
	var out bytes.Buffer
	if err := runViewWork([]string{"contract"}, &out, nil); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	for _, want := range []string{
		"rename --work-id ID --title TEXT",
		"archive --work-id ID",
		"restore --work-id ID",
		"pin --work-id ID",
		"unpin --work-id ID",
		"recover --work-id ID --request-id ID --reason TEXT",
		"delete --work-id ID --confirm ID",
		"origin cli",
		"local_actions_cli",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("contract missing %q", want)
		}
	}
	if !strings.Contains(text, "«Shared work and authority»") || strings.Contains(text, "Run them only when") {
		t.Fatal("contract restates the agent rule instead of pointing to its home")
	}
	if !strings.Contains(text, "work_gone and delete_pending are errors with a non-zero exit") {
		t.Fatal("contract does not say work_gone and delete_pending are errors")
	}
	if strings.Contains(text, "All mutations require --host") {
		t.Fatal("contract still says every mutation needs a native producer")
	}
}
