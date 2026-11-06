package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// Persistence owns corruption, locking and atomic retained archival contracts.
func TestWorkJournalPreservesCorruptionAndUnknownVersion(t *testing.T) {
	for _, raw := range []string{`{`, `{"schema_version":99}`, `{"schema_version":1}`} {
		dir := t.TempDir()
		path := filepath.Join(dir, "journal.json")
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readWorkJournal(dir); err == nil {
			t.Fatal("accepted invalid journal")
		}
		got, _ := os.ReadFile(path)
		if string(got) != raw {
			t.Fatal("changed corrupt journal")
		}
	}
}
func TestWorkJournalIndependentNonblockingLock(t *testing.T) {
	dir := t.TempDir()
	a, err := lockWorkPublisher(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Release()
	if b, err := lockWorkPublisher(dir); err == nil {
		b.Release()
		t.Fatal("second writer acquired lock")
	}
}
func TestWorkJournalArchiveRequiresDefinitiveRejection(t *testing.T) {
	j := newWorkJournal()
	i := &viewWorkIntent{ID: "intent", State: "needs_reconciliation", Steps: []viewWorkStep{{Method: "POST", Path: "/api/test", Body: []byte("frozen")}}}
	j.Pending = append(j.Pending, i)
	if err := j.archiveRejected("intent", "explicit reason"); err == nil {
		t.Fatal("archived unknown outcome")
	}
	i.Steps[0].Rejection = "current_predecessor_required"
	i.State = "rejected_conflict"
	if err := j.archiveRejected("intent", "explicit reason"); err != nil {
		t.Fatal(err)
	}
	if len(j.Pending) != 0 || len(j.History) != 1 || string(j.History[0].Steps[0].Body) != "frozen" {
		t.Fatal("lost archived bytes")
	}
}

func TestWorkJournalRetainsSingleArchiveClassification(t *testing.T) {
	dir := t.TempDir()
	j := newWorkJournal()
	source := viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}
	request := []byte(`{"schema_version":3,"operation_id":"frozen-patch","actor":{"kind":"native_producer","producer_id":"producer-fixture"},"expected_version":1,"state":"closed","reason":"completed: Retained reason"}`)
	work := strings.Repeat("a", 64)
	receipt, _ := json.Marshal(viewer.TimelineOperationResult{SchemaVersion: 3, OperationID: "frozen-patch", Kind: "session_patch", ViewerSessionID: work, SessionVersion: 2})
	i := &viewWorkIntent{ID: "rejected", WorkID: work, Source: source, Action: "close", Context: json.RawMessage(`{"command":{"Action":"close","WorkID":"` + work + `","Host":"codex","NativeSessionID":"fixture"}}`), State: "rejected_conflict", Steps: []viewWorkStep{{Method: "PATCH", Path: "/api/timeline/sessions/" + work, ContentType: "application/json", Body: request, Attempted: true, Response: receipt}, {Method: "PATCH", Path: "/api/timeline/sessions/" + work, ContentType: "application/json", Body: request, Attempted: true, Rejection: "version_conflict"}}}
	j.Pending = append(j.Pending, i)
	if err := writeWorkJournal(dir, j); err != nil {
		t.Fatal(err)
	}
	// A crash before the rename retains active; after rename it retains archived.
	before, err := readWorkJournal(dir)
	if err != nil || len(before.Pending) != 1 {
		t.Fatal("lost active classification", err)
	}
	if err = j.archiveRejected(i.ID, "Explicit conflict recovery"); err != nil {
		t.Fatal(err)
	}
	if err = writeWorkJournal(dir, j); err != nil {
		t.Fatal(err)
	}
	after, err := readWorkJournal(dir)
	if err != nil || len(after.Pending) != 0 || len(after.History) != 1 {
		t.Fatal("ambiguous archive classification", err)
	}
	if !bytes.Equal(after.History[0].Steps[0].Response, receipt) || !bytes.Equal(after.History[0].Steps[1].Body, request) {
		t.Fatal("lost recovery evidence")
	}
	info, _ := os.Stat(filepath.Join(dir, "journal.json"))
	if info.Mode().Perm() != 0600 {
		t.Fatal("journal mode")
	}
}

func TestWorkPublisherContendedChild(t *testing.T) {
	dir := os.Getenv("HIVE_TEST_PUBLISHER_LOCK")
	if dir == "" {
		return
	}
	lock, err := lockWorkPublisher(dir)
	if os.Getenv("HIVE_TEST_PUBLISHER_BUSY") == "1" {
		if err == nil {
			lock.Release()
			t.Fatal("child acquired held publisher lock")
		}
		if !strings.Contains(err.Error(), "busy") {
			t.Fatal(err)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
		lock.Release()
	}
}
func TestWorkPublisherLockAcrossProcesses(t *testing.T) {
	dir := t.TempDir()
	lock, err := lockWorkPublisher(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	run := func(busy string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWorkPublisherContendedChild$")
		cmd.Env = append(os.Environ(), "HIVE_TEST_PUBLISHER_LOCK="+dir, "HIVE_TEST_PUBLISHER_BUSY="+busy)
		if raw, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("lock subprocess: %v %s", err, raw)
		}
	}
	run("1")
	lock.Release()
	run("0")
}

func TestWorkJournalRejectsMalformedBindingWithoutChangingBytes(t *testing.T) {
	for _, binding := range []string{`{}`, `{"work_id":"bad","project":"/root","change":"/root/change","source":{"project_root":"/root","host":"codex","native_session_id":"native"}}`, `{"work_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","project":"relative","change":"/root/change","source":{"project_root":"/root","host":"codex","native_session_id":"native"}}`} {
		dir := t.TempDir()
		raw := []byte(`{"schema_version":1,"bindings":[` + binding + `],"pending":[],"history":[]}`)
		path := filepath.Join(dir, "journal.json")
		os.WriteFile(path, raw, 0600)
		if _, err := readWorkJournal(dir); err == nil {
			t.Fatal("accepted semantically malformed binding")
		}
		got, _ := os.ReadFile(path)
		if string(got) != string(raw) {
			t.Fatal("changed invalid journal")
		}
	}
}

func TestWorkJournalRejectsUnsafeLinksAndRoutes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")
	raw, _ := json.Marshal(newWorkJournal())
	target := filepath.Join(dir, "target.json")
	os.WriteFile(target, raw, 0600)
	if err := os.Link(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted hardlinked journal")
	}
	os.Remove(path)
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted symlink journal")
	}
	i := &viewWorkIntent{ID: "pending", WorkID: strings.Repeat("a", 64), Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Action: "stage", State: "needs_reconciliation", Context: json.RawMessage(`{}`), Steps: []viewWorkStep{}}
	for _, step := range []viewWorkStep{{Method: "DELETE", Path: "/api/timeline/sessions/" + i.WorkID}, {Method: "POST", Path: "/api/timeline/sessions/" + strings.Repeat("b", 64) + "/events"}, {Method: "POST", Path: "/api/timeline/sessions/" + i.WorkID + "/events?override=1"}} {
		i.Steps = []viewWorkStep{step}
		j := newWorkJournal()
		j.Pending = append(j.Pending, i)
		if err := validateWorkJournal(j); err == nil {
			t.Fatal("accepted corrupt mutation route")
		}
	}
}

func TestWorkJournalRejectsMissingSemanticIntent(t *testing.T) {
	j := newWorkJournal()
	j.Pending = append(j.Pending, &viewWorkIntent{ID: "pending", WorkID: strings.Repeat("a", 64), Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Action: "stage", State: "prepared", Context: json.RawMessage(`{}`), Steps: []viewWorkStep{}})
	dir := t.TempDir()
	raw, _ := json.Marshal(j)
	path := filepath.Join(dir, "journal.json")
	os.WriteFile(path, raw, 0600)
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted missing semantic intent fields")
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, raw) {
		t.Fatal("modified corrupt semantic intent")
	}
}

func TestWorkJournalClosureRequestSurvivesReloadAndReplay(t *testing.T) {
	for _, committed := range []bool{false, true} {
		t.Run(fmt.Sprint("committed-", committed), func(t *testing.T) {
			_, c := workEngineFixture(t)
			opened := engineWorkOpen(t, c)
			c.Action, c.WorkID, c.Outcome, c.Reason = "close", opened.WorkID, "completed", "Implementation verified and delivered"
			e, err := openViewWorkEngine(c)
			if err != nil {
				t.Fatal(err)
			}
			defer e.Close()
			s, err := e.session(c.WorkID)
			if err != nil {
				t.Fatal(err)
			}
			actor, err := e.actor(s, c)
			if err != nil {
				t.Fatal(err)
			}
			method, path, body, _, err := e.nextLifecycleStep(s, actor, c)
			if err != nil {
				t.Fatal(err)
			}
			step := viewWorkStep{Method: method, Path: path, ContentType: "application/json", Body: body}
			i, err := e.prepareIntent("close", c.WorkID, workLifecycleSource(s, actor), viewWorkLifecycleContext{Command: c, FirstStep: &step})
			if err != nil {
				t.Fatal(err)
			}
			if _, err = e.appendStep(i, method, path, "application/json", body); err != nil {
				t.Fatal(err)
			}
			i.Steps[0].Attempted = true
			i.State = "needs_reconciliation"
			if err = e.save(); err != nil {
				t.Fatal(err)
			}
			if committed {
				if _, _, err = viewWorkHTTP(e.info, method, path, "application/json", body); err != nil {
					t.Fatal(err)
				}
			}
			e.journal, err = readWorkJournal(e.dir)
			if err != nil {
				t.Fatal("valid uncertain closure rejected on reload", err)
			}
			c.Action = "resume"
			if err = e.recover(c, io.Discard); err != nil {
				t.Fatal("closure recovery failed", err)
			}
			current, err := e.session(c.WorkID)
			if err != nil || current.State != viewer.TimelineClosed {
				t.Fatal("closure was not confirmed", err)
			}
			// Earlier candidates retained bodies in confirmed history. Those valid v3
			// bytes must remain readable, even though new confirmations compact them.
			retained := e.journal.History[len(e.journal.History)-1]
			retained.Steps[0].Body = body
			if err = e.save(); err != nil {
				t.Fatal(err)
			}
			if _, err = readWorkJournal(e.dir); err != nil {
				t.Fatal("valid retained closure body rejected", err)
			}
		})
	}
}

func TestWorkJournalReconcilesScopeReferenceRejection(t *testing.T) {
	_, c := workEngineFixture(t)
	opened := engineWorkOpen(t, c)
	c.WorkID = opened.WorkID
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	s, err := e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	actor, err := e.actor(s, c)
	if err != nil {
		t.Fatal(err)
	}
	create := viewer.TimelineTaskCreateRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Declare task"), TaskKey: "T1", Title: "Scope change recovery", PlanOrder: 1, Segments: []viewer.TimelineTaskSegmentInput{{PlannedStage: "build", Order: 1}}}
	raw, _ := json.Marshal(create)
	response, _, err := viewWorkHTTP(e.info, "POST", "/api/timeline/sessions/"+c.WorkID+"/tasks", "application/json", raw)
	if err != nil {
		t.Fatal(err)
	}
	created, err := viewer.DecodeTimelineOperationResult(response)
	if err != nil {
		t.Fatal(err)
	}
	s, err = e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	order := int64(2)
	request := viewer.TimelineTaskPatchRequest{TimelineNativeOperation: workLifecycleBase(s, actor, "Change declared scope"), ExpectedTaskVersion: created.TaskVersion, PlanOrder: &order}
	body, _ := json.Marshal(request)
	path := "/api/timeline/sessions/" + c.WorkID + "/tasks/" + created.TaskID
	c.Action, c.Key, c.Reason = "task update", "T1", "Change declared scope"
	c.Metadata = &viewWorkTaskMetadata{Order: &order}
	step := viewWorkStep{Method: "PATCH", Path: path, ContentType: "application/json", Body: body}
	i, err := e.prepareIntent(c.Action, c.WorkID, workLifecycleSource(s, actor), viewWorkLifecycleContext{Command: c, FirstStep: &step})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.appendStep(i, step.Method, path, step.ContentType, body); err != nil {
		t.Fatal(err)
	}
	if _, err = e.executeStep(i, 0); err == nil {
		t.Fatal("server accepted missing scope reference")
	}
	if i.State != "rejected_conflict" || i.Steps[0].Rejection != "scope_reference_required" {
		t.Fatal("structured scope rejection remained uncertain")
	}
	// Older candidates retained this exact rejection as an unknown outcome.
	// Reconciliation must replay the original request and recover that evidence.
	i.State = "needs_reconciliation"
	i.Steps[0].Rejection = ""
	if err = e.save(); err != nil {
		t.Fatal(err)
	}
	e.journal, err = readWorkJournal(e.dir)
	if err != nil {
		t.Fatal(err)
	}
	reconcile := c
	reconcile.Action, reconcile.Pending, reconcile.Reason = "reconcile", i.ID, "Archive the exact request rejected for its missing scope reference"
	if err = e.recover(reconcile, io.Discard); err != nil {
		t.Fatal(err)
	}
	if len(e.journal.Pending) != 0 {
		t.Fatal("reconciliation left permanent pending intent")
	}
	j, err := readWorkJournal(e.dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, archived := range j.History {
		if archived.ID == i.ID {
			found = true
			if archived.State != "archived_rejected" || archived.Steps[0].Rejection != "scope_reference_required" || !bytes.Equal(archived.Steps[0].Body, body) {
				t.Fatal("archive lost rejection evidence")
			}
		}
	}
	if !found {
		t.Fatal("archive missing")
	}
	current, err := e.session(c.WorkID)
	if err != nil {
		t.Fatal(err)
	}
	if workLifecycleTask(current, "T1").PlanOrder != 1 {
		t.Fatal("rejected replay mutated scope")
	}
}
func TestWorkJournalUncertainOutcomesStayUnarchiveable(t *testing.T) {
	for _, code := range []string{"operation_conflict", "outcome_unknown", "", "internal", "malformed_response", "unrecognized_future_error"} {
		if definitiveWorkRejection(code) {
			t.Fatalf("uncertain outcome %q classified as rejection", code)
		}
	}
}

func TestWorkJournalLegacyCreateResponsesSurviveLoadAndExplicitResume(t *testing.T) {
	_, c := workEngineFixture(t)
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	source := viewer.TimelineSource{ProjectRoot: c.Project, Host: viewer.TimelineHost(c.Host), NativeSessionID: c.NativeSessionID}
	request := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "original-preupgrade-create", Source: source, WorkKey: &c.Key, Title: c.Title, Summary: c.Summary, Tickets: []viewer.TimelineTicketInput{}}
	body, _ := json.Marshal(request)
	intent, err := e.prepareIntent("open", "", source, viewWorkOpenContext{Binding: viewWorkBinding{Key: c.Key, Project: c.Project, Change: c.Change, Source: source}, Request: body})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.appendStep(intent, "POST", "/api/timeline/sessions", "application/json", body); err != nil {
		t.Fatal(err)
	}
	raw, err := e.executeStep(intent, 0)
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]json.RawMessage
	json.Unmarshal(raw, &envelope)
	var session map[string]json.RawMessage
	json.Unmarshal(envelope["session"], &session)
	for _, field := range []string{"last_activity_at", "closed_at", "autoarchive_at", "archive_origin"} {
		delete(session, field)
	}
	envelope["session"], _ = json.Marshal(session)
	legacy, _ := json.Marshal(envelope)
	if _, err = viewer.DecodeTimelineSessionResponse(legacy); err == nil {
		t.Fatal("public decoder accepted old response")
	}
	intent.Steps[0].Response = legacy
	historical := compactConfirmedWorkIntent(intent)
	historical.ID += "-history"
	historical.Steps[0].Body = append([]byte(nil), body...)
	e.journal.History = append(e.journal.History, historical)
	dir, workID := e.dir, intent.WorkID
	original, _ := json.Marshal(e.journal)
	path := filepath.Join(dir, "journal.json")
	if err = os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	e.Close()
	loaded, err := readWorkJournal(dir)
	if err != nil {
		t.Fatal("valid legacy pending/history journal rejected", err)
	}
	if len(loaded.Pending) != 1 || len(loaded.History) != 1 {
		t.Fatal("lost journal entries")
	}
	for _, retained := range []*viewWorkIntent{loaded.Pending[0], loaded.History[0]} {
		if !bytes.Equal(retained.Steps[0].Response, legacy) || !bytes.Equal(retained.Steps[0].Body, body) {
			t.Fatal("load replaced frozen response or request")
		}
	}
	disk, _ := os.ReadFile(path)
	if !bytes.Equal(disk, original) {
		t.Fatal("read rewrote legacy journal")
	}
	// Missing additions are compatible; present corruption and old schema
	// versions are not, in either pending or confirmed retained history.
	for _, location := range []string{"pending", "history"} {
		for field, value := range map[string]json.RawMessage{
			"last_activity_at": json.RawMessage(`"not-a-time"`), "closed_at": json.RawMessage(`123`), "autoarchive_at": json.RawMessage(`"not-a-time"`), "archive_origin": json.RawMessage(`"unknown"`), "schema_version": json.RawMessage(`2`),
		} {
			t.Run(location+"/"+field, func(t *testing.T) {
				var corruptEnvelope map[string]json.RawMessage
				json.Unmarshal(legacy, &corruptEnvelope)
				var corruptSession map[string]json.RawMessage
				json.Unmarshal(corruptEnvelope["session"], &corruptSession)
				corruptSession[field] = value
				corruptEnvelope["session"], _ = json.Marshal(corruptSession)
				corruptResponse, _ := json.Marshal(corruptEnvelope)
				var corrupt viewWorkJournal
				json.Unmarshal(original, &corrupt)
				target := corrupt.Pending[0]
				if location == "history" {
					target = corrupt.History[0]
				}
				target.Steps[0].Response = corruptResponse
				bad, _ := json.Marshal(corrupt)
				if err := os.WriteFile(path, bad, 0600); err != nil {
					t.Fatal(err)
				}
				if _, err := readWorkJournal(dir); err == nil {
					t.Fatal("accepted corrupt retained response")
				}
				retained, _ := os.ReadFile(path)
				if !bytes.Equal(retained, bad) {
					t.Fatal("changed corrupt journal")
				}
			})
		}
	}
	if err = os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}
	// A caller mismatch must not alter the retained pending operation.
	resume := c
	resume.Action = "resume"
	resume.WorkID = workID
	resume.NativeSessionID += "-other"
	if err = executeViewWork(resume, io.Discard); err == nil || !strings.Contains(err.Error(), "source_mismatch") {
		t.Fatal("resume accepted foreign caller", err)
	}
	disk, _ = os.ReadFile(path)
	if !bytes.Equal(disk, original) {
		t.Fatal("source mismatch changed journal")
	}
	resume.NativeSessionID = c.NativeSessionID
	if err = executeViewWork(resume, io.Discard); err != nil {
		t.Fatal("explicit resume could not consume legacy create receipt", err)
	}
	loaded, err = readWorkJournal(dir)
	if err != nil || len(loaded.Pending) != 0 || len(loaded.History) != 2 {
		t.Fatal("resume lost retained history", err)
	}
	for _, retained := range loaded.History {
		if !bytes.Equal(retained.Steps[0].Response, legacy) {
			t.Fatal("resume rewrote retained response")
		}
	}
}
