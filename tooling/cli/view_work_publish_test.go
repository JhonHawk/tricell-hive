package main

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"tricell-hive/tooling/viewer"
)

type workPublishFixture struct {
	dropNext            atomic.Bool
	dropBefore          atomic.Bool
	dropPath            atomic.Value
	hookMu              sync.Mutex
	beforeSnapshot      func()
	store               *viewer.Store
	state, root, change string
	info                serverInfo
	mutations           atomic.Int32
	server              *httptest.Server
	handler             *viewer.Server
}

func newWorkPublishFixture(t *testing.T) *workPublishFixture {
	t.Helper()
	f := &workPublishFixture{state: t.TempDir(), root: t.TempDir()}
	f.state, _ = filepath.EvalSymlinks(f.state)
	f.root, _ = filepath.EvalSymlinks(f.root)
	if err := os.Mkdir(filepath.Join(f.root, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := viewer.RegisterRepository(f.state, f.root); err != nil {
		t.Fatal(err)
	}
	f.change = filepath.Join(f.root, "change")
	if err := os.Mkdir(f.change, 0700); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"proposal.md": "# Proposal\nOriginal proposal\n", "tasks.md": "# Tasks\nT1: implement\n"} {
		if err := os.WriteFile(filepath.Join(f.change, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	store, err := viewer.Open(viewer.Options{StateDir: f.state})
	if err != nil {
		t.Fatal(err)
	}
	f.server = httptest.NewUnstartedServer(nil)
	port := f.server.Listener.Addr().(*net.TCPAddr).Port
	f.handler, err = viewer.NewServer(viewer.ServerOptions{Store: store, Token: strings.Repeat("x", 32), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	f.store = store
	f.dropPath.Store("")
	f.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			f.mutations.Add(1)
		}
		if strings.HasSuffix(r.URL.Path, "/document-snapshots") {
			f.hookMu.Lock()
			hook := f.beforeSnapshot
			f.beforeSnapshot = nil
			f.hookMu.Unlock()
			if hook != nil {
				hook()
			}
		}
		if path := f.dropPath.Load().(string); path != "" && strings.HasSuffix(r.URL.Path, path) && r.Method == "POST" && f.dropNext.CompareAndSwap(true, false) {
			if f.dropBefore.Load() {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err == nil {
					conn.Close()
				}
				return
			}
			recorder := httptest.NewRecorder()
			f.handler.ServeHTTP(recorder, r)
			if recorder.Code >= 200 && recorder.Code < 300 {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err == nil {
					conn.Close()
				}
				return
			}
			for key, values := range recorder.Header() {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}
			w.WriteHeader(recorder.Code)
			w.Write(recorder.Body.Bytes())
			return
		}
		f.handler.ServeHTTP(w, r)
	})
	f.server.Start()
	t.Cleanup(f.server.Close)
	paths := newViewPaths(f.state)
	lock, err := viewer.AcquireLock(paths.lock, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { lock.Release() })
	f.info = serverInfo{PID: os.Getpid(), Port: port, Token: strings.Repeat("x", 32), StateDir: f.state}
	if err := writeServerInfo(paths, f.info); err != nil {
		t.Fatal(err)
	}
	return f
}
func (f *workPublishFixture) run(t *testing.T, args ...string) (viewWorkResult, error) {
	t.Helper()
	args = append(args, "--host", "codex", "--native-session-id", "publication-native", "--state-dir", f.state)
	var out bytes.Buffer
	err := runViewWork(args, &out, strings.NewReader(""))
	var result viewWorkResult
	if err == nil {
		if out.Len() > 4096 {
			t.Fatal("mutation result exceeds 4096 bytes")
		}
		if e := json.Unmarshal(out.Bytes(), &result); e != nil {
			t.Fatal(e)
		}
	}
	return result, err
}
func TestWorkPublishRealV3NoopAndRevision(t *testing.T) {
	f := newWorkPublishFixture(t)
	first, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan")
	if err != nil {
		t.Fatal(err)
	}
	if first.Result != "published" || first.RevisionID == "" {
		t.Fatalf("unconfirmed publication: %+v", first)
	}
	mutationCount := f.mutations.Load()
	second, err := f.run(t, "publish", "--work-id", first.WorkID, "--change", f.change, "--stage", "plan")
	if err != nil {
		t.Fatal(err)
	}
	if f.mutations.Load() != mutationCount {
		t.Fatal("semantic no-op emitted a mutating request")
	}
	if second.Result != "unchanged" || second.RevisionID != first.RevisionID || second.SessionVersion != first.SessionVersion {
		t.Fatalf("no-op mutated remote work: %+v -> %+v", first, second)
	}
	if err := os.WriteFile(filepath.Join(f.change, "tasks.md"), []byte("# Tasks\nUpdated task description\n"), 0600); err != nil {
		t.Fatal(err)
	}
	third, err := f.run(t, "publish", "--work-id", first.WorkID, "--change", f.change, "--stage", "plan")
	if err != nil {
		t.Fatal(err)
	}
	if third.RevisionID == first.RevisionID || third.ChangedCount == nil || *third.ChangedCount != 1 {
		t.Fatalf("missing single-document revision: %+v", third)
	}
	raw, err := viewTimelineHTTP(f.info, "GET", "/api/timeline/sessions/"+first.WorkID, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(observed.Session.Documents) != 3 {
		t.Fatalf("expected immutable original plus one changed document, got %d", len(observed.Session.Documents))
	}
	snapshots := 0
	for _, event := range observed.Session.Events {
		if event.Kind == "stage" && event.Title != "" {
			t.Fatalf("first publication repeats the stage name in its title: %q", event.Title)
		}
		snapshots += len(event.DocumentSnapshots)
		for _, snapshot := range event.DocumentSnapshots {
			if len(snapshot.Tabs) != 2 || snapshot.Tabs[0].Label != "Proposal" || snapshot.Tabs[1].Label != "Tasks" {
				t.Fatal("wrong tab order")
			}
		}
	}
	if snapshots != 2 {
		t.Fatalf("expected two revisions, got %d", snapshots)
	}
}

func (f *workPublishFixture) observe(t *testing.T, id string) viewer.TimelineSession {
	t.Helper()
	raw, err := viewTimelineHTTP(f.info, "GET", "/api/timeline/sessions/"+id, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	response, err := viewer.DecodeTimelineSessionResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return response.Session
}
func (f *workPublishFixture) pending(t *testing.T) *viewWorkIntent {
	t.Helper()
	j, err := readWorkJournal(filepath.Join(f.state, "publisher", "v1"))
	if err != nil {
		t.Fatal(err)
	}
	if len(j.Pending) != 1 {
		t.Fatalf("expected one pending intent, got %d", len(j.Pending))
	}
	return j.Pending[0]
}
func TestWorkPublishLostResponsesReplayFrozenBytes(t *testing.T) {
	for _, path := range []string{"/documents", "/document-snapshots", "/api/timeline/sessions"} {
		t.Run(path, func(t *testing.T) {
			f := newWorkPublishFixture(t)
			f.dropPath.Store(path)
			f.dropNext.Store(true)
			creation := []string{"publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"}
			if _, err := f.run(t, creation...); err == nil || !strings.Contains(err.Error(), "outcome_unknown") {
				t.Fatalf("missing uncertain outcome: %v", err)
			}
			pending := f.pending(t)
			if path == "/document-snapshots" {
				if _, err := f.run(t, "reconcile", "--work-id", pending.WorkID, "--pending", pending.ID, "--action", "archive-rejected", "--reason", "An uncertain response is not a definitive rejection"); err == nil || !strings.Contains(err.Error(), "outcome_unknown") {
					t.Fatalf("archived uncertain outcome: %v", err)
				}
				f.pending(t)
			}
			// Editing/deleting live sources cannot alter an already prepared request.
			if err := os.WriteFile(filepath.Join(f.change, "tasks.md"), []byte("# Tasks\nDifferent content after interruption\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if path == "/api/timeline/sessions" {
				if err := os.Remove(filepath.Join(f.change, "tasks.md")); err != nil {
					t.Fatal(err)
				}
			}
			var result viewWorkResult
			var err error
			if pending.WorkID == "" {
				result, err = f.run(t, creation...)
			} else {
				result, err = f.run(t, "resume", "--work-id", pending.WorkID)
			}
			if err != nil {
				t.Fatal(err)
			}
			s := f.observe(t, result.WorkID)
			if len(s.Documents) != 2 {
				t.Fatalf("replay duplicated or changed documents: %d", len(s.Documents))
			}
			count := 0
			for _, event := range s.Events {
				for _, snapshot := range event.DocumentSnapshots {
					count++
					for _, tab := range snapshot.Tabs {
						if tab.Label != "Tasks" {
							continue
						}
						raw, err := viewTimelineHTTP(f.info, "GET", "/api/timeline/sessions/"+s.ViewerSessionID+"/documents/"+tab.DocumentID, "", nil, viewer.TimelineMaxBody)
						if err != nil {
							t.Fatal(err)
						}
						doc, err := viewer.DecodeTimelineDocumentResponse(raw)
						if err != nil {
							t.Fatal(err)
						}
						if doc.Content != "# Tasks\nT1: implement\n" {
							t.Fatal("resume read edited live source")
						}
					}
				}
			}
			if count != 1 {
				t.Fatalf("snapshot replay duplicated revisions: %d", count)
			}
			journal, err := readWorkJournal(filepath.Join(f.state, "publisher", "v1"))
			if err != nil {
				t.Fatal(err)
			}
			if len(journal.Pending) != 0 || len(journal.History) != 1 {
				t.Fatal("publication not confirmed")
			}
			var retained workPublicationIntent
			if err := json.Unmarshal(journal.History[0].Context, &retained); err != nil {
				t.Fatal(err)
			}
			for _, doc := range retained.Documents {
				if len(doc.Bytes) != 0 || doc.SHA256 == "" {
					t.Fatal("confirmed history retained disposable bytes or lost identity")
				}
			}
			for _, step := range journal.History[0].Steps {
				if len(step.Body) != 0 || len(step.Response) == 0 {
					t.Fatal("confirmed history retained disposable request body or lost receipt")
				}
			}
		})
	}
}
func TestWorkPublishConflictReconcileAndFreshIntent(t *testing.T) {
	f := newWorkPublishFixture(t)
	var externalRevision string
	var externalMu sync.Mutex
	f.hookMu.Lock()
	f.beforeSnapshot = func() {
		index, err := viewIndex(f.info)
		if err != nil {
			t.Error(err)
			return
		}
		if len(index.TimelineSessions) != 1 {
			t.Error("missing work")
			return
		}
		s := f.observe(t, index.TimelineSessions[0].ViewerSessionID)
		stage, err := publicationStage(s, "plan")
		if err != nil {
			t.Error(err)
			return
		}
		external, err := f.store.TimelineSnapshot(s.ViewerSessionID, stage.EventID, viewer.TimelineDocumentSnapshotRequest{SchemaVersion: 3, OperationID: "external-snapshot", Actor: viewer.TimelineActor{Kind: "native_producer", ProducerID: s.ForegroundProducerID}, ExpectedVersion: stage.Version, ExpectedSessionVersion: s.Version, DocumentKey: "plan", Status: "complete", Tabs: []viewer.TimelineDocumentTab{{Label: "External plan", DocumentID: s.Documents[0].DocumentID}}})
		if err != nil {
			t.Error(err)
			return
		}
		externalMu.Lock()
		externalRevision = external.DocumentRevisionID
		externalMu.Unlock()
	}
	f.hookMu.Unlock()
	_, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan")
	if err == nil || !strings.Contains(err.Error(), "version_conflict") {
		t.Fatalf("external revision was overwritten: %v", err)
	}
	pending := f.pending(t)
	if pending.State != "rejected_conflict" {
		t.Fatalf("expected definitive rejection, got %s", pending.State)
	}
	if _, err = f.run(t, "publish", "--work-id", pending.WorkID, "--change", f.change, "--stage", "plan"); err == nil {
		t.Fatal("replacement bypassed pending conflict")
	}
	if _, err = f.run(t, "reconcile", "--work-id", pending.WorkID, "--pending", pending.ID, "--action", "archive-rejected", "--reason", "Retain the rejected intent and publish the revised current chain"); err != nil {
		t.Fatal(err)
	}
	result, err := f.run(t, "publish", "--work-id", pending.WorkID, "--change", f.change, "--stage", "plan")
	if err != nil {
		t.Fatal(err)
	}
	s := f.observe(t, pending.WorkID)
	tip, _ := publicationTip(s, "plan")
	externalMu.Lock()
	expectedPredecessor := externalRevision
	externalMu.Unlock()
	if tip == nil || tip.DocumentRevisionID != result.RevisionID || tip.PredecessorRevisionID != expectedPredecessor {
		t.Fatal("fresh intent lost the external chain")
	}
	j, err := readWorkJournal(filepath.Join(f.state, "publisher", "v1"))
	if err != nil {
		t.Fatal(err)
	}
	archived := false
	for _, history := range j.History {
		if history.ID == pending.ID && history.State == "archived_rejected" && len(history.Steps) > 0 && len(history.Context) > 0 {
			archived = true
		}
	}
	if !archived {
		t.Fatal("reconciliation discarded frozen evidence")
	}
}
func TestWorkPublishOptionalRemovalAndStatusCreateRevisions(t *testing.T) {
	f := newWorkPublishFixture(t)
	if err := os.WriteFile(filepath.Join(f.change, "design.md"), []byte("# Design\nOriginal design\n"), 0600); err != nil {
		t.Fatal(err)
	}
	first, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(f.change, "design.md")); err != nil {
		t.Fatal(err)
	}
	second, err := f.run(t, "publish", "--work-id", first.WorkID, "--change", f.change, "--stage", "plan")
	if err != nil {
		t.Fatal(err)
	}
	third, err := f.run(t, "publish", "--work-id", first.WorkID, "--change", f.change, "--stage", "plan", "--status", "complete")
	if err != nil {
		t.Fatal(err)
	}
	if first.RevisionID == second.RevisionID || second.RevisionID == third.RevisionID {
		t.Fatal("document removal or status was treated as unchanged")
	}
	s := f.observe(t, first.WorkID)
	if len(s.Documents) != 3 {
		t.Fatal("optional removal deleted historical document")
	}
	tip, _ := publicationTip(s, "plan")
	if len(tip.Tabs) != 2 || tip.Status != "complete" {
		t.Fatal("wrong current snapshot")
	}
}

func TestWorkPublishPreflightAndExplicitResearchSelection(t *testing.T) {
	f := newWorkPublishFixture(t)
	if err := os.Remove(filepath.Join(f.change, "tasks.md")); err != nil {
		t.Fatal(err)
	}
	if _, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"); err == nil {
		t.Fatal("missing required document accepted")
	}
	if f.mutations.Load() != 0 {
		t.Fatal("preflight failure mutated remote state")
	}
	if err := os.WriteFile(filepath.Join(f.change, "notes.md"), []byte("# Research\nRetained original notes\n"), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "research-10", "--title", "Investigate export", "--summary", "Record research findings", "--stage", "research", "--document-key", "research", "--document", "notes.md", "--reason", "Start the requested research")
	if err != nil {
		t.Fatal(err)
	}
	s := f.observe(t, result.WorkID)
	tip, _ := publicationTip(s, "research")
	if tip == nil || len(tip.Tabs) != 1 || tip.Tabs[0].Label != "notes.md" || len(s.Documents) != 1 {
		t.Fatal("explicit selection inferred unrelated change-folder documents")
	}
}
func TestWorkPublishDeletionOverridesPendingIntent(t *testing.T) {
	f := newWorkPublishFixture(t)
	f.hookMu.Lock()
	f.beforeSnapshot = func() {
		index, err := viewIndex(f.info)
		if err != nil {
			t.Error(err)
			return
		}
		if len(index.TimelineSessions) != 1 {
			t.Error("missing work")
			return
		}
		s := f.observe(t, index.TimelineSessions[0].ViewerSessionID)
		_, err = f.store.TimelineDelete(s.ViewerSessionID, viewer.TimelineMutationRequest{SchemaVersion: 3, OperationID: "local-delete", Actor: viewer.TimelineActor{Kind: "local_viewer"}, ExpectedVersion: s.Version})
		if err != nil {
			t.Error(err)
		}
	}
	f.hookMu.Unlock()
	if _, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"); err == nil {
		t.Fatal("deleted work reported published")
	}
	pending := f.pending(t)
	mutationCount := f.mutations.Load()
	if _, err := f.run(t, "resume", "--work-id", pending.WorkID); err == nil {
		t.Fatal("deleted work resumed")
	}
	if f.mutations.Load() != mutationCount {
		t.Fatal("recovery mutated or recreated deleted work")
	}
	index, err := viewIndex(f.info)
	if err != nil {
		t.Fatal(err)
	}
	if len(index.TimelineSessions) != 0 {
		t.Fatal("deleted work recreated")
	}
	f.pending(t)
}

func TestWorkPublishTabLabelPreflightHasNoMutations(t *testing.T) {
	for _, name := range []string{strings.Repeat("a", 90), "control\x01label"} {
		t.Run(name, func(t *testing.T) {
			f := newWorkPublishFixture(t)
			dir := filepath.Join(f.change, "specs", name)
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# Spec\nDocument payload\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := f.run(t, "publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"); err == nil {
				t.Fatal("invalid v3 tab label accepted")
			}
			if f.mutations.Load() != 0 {
				t.Fatalf("invalid v3 tab label caused %d remote mutations", f.mutations.Load())
			}
		})
	}
}

func TestWorkPublishLostCreationThroughAliasUsesFrozenSource(t *testing.T) {
	f := newWorkPublishFixture(t)
	alias := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(f.root, alias); err != nil {
		t.Fatal(err)
	}
	f.dropPath.Store("/api/timeline/sessions")
	f.dropNext.Store(true)
	args := []string{"publish", "--project", alias, "--change", filepath.Join(alias, "change"), "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"}
	if _, err := f.run(t, args...); err == nil || !strings.Contains(err.Error(), "outcome_unknown") {
		t.Fatalf("missing uncertain creation: %v", err)
	}
	if err := os.Remove(filepath.Join(f.change, "tasks.md")); err != nil {
		t.Fatal(err)
	}
	result, err := f.run(t, args...)
	if err != nil {
		t.Fatal(err)
	}
	s := f.observe(t, result.WorkID)
	if len(s.Documents) != 2 {
		t.Fatal("alias recovery did not publish original frozen set")
	}
	index, err := viewIndex(f.info)
	if err != nil {
		t.Fatal(err)
	}
	if len(index.TimelineSessions) != 1 {
		t.Fatal("alias recovery duplicated work")
	}
}

func TestWorkPublishCreationSelectorsResumeSavedReceipt(t *testing.T) {
	for _, checkpoint := range []string{"create receipt before context advance", "later upload"} {
		t.Run(checkpoint, func(t *testing.T) {
			f := newWorkPublishFixture(t)
			if checkpoint == "create receipt before context advance" {
				f.dropPath.Store("/events")
				f.dropBefore.Store(true)
			} else {
				f.dropPath.Store("/documents")
			}
			f.dropNext.Store(true)
			args := []string{"publish", "--project", f.root, "--change", f.change, "--key", "ticket-10", "--title", "Plan export", "--summary", "Prepare export implementation", "--stage", "plan", "--reason", "Start the requested plan"}
			if _, err := f.run(t, args...); err == nil || !strings.Contains(err.Error(), "outcome_unknown") {
				t.Fatalf("expected interrupted publication: %v", err)
			}
			dir := filepath.Join(f.state, "publisher", "v1")
			journal, err := readWorkJournal(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(journal.Pending) != 1 {
				t.Fatal("missing pending publication")
			}
			pending := journal.Pending[0]
			if pending.WorkID == "" || len(pending.Steps[0].Response) == 0 {
				t.Fatal("create receipt was not durably saved")
			}
			if checkpoint == "create receipt before context advance" {
				// Restore the exact journal boundary after executeStep saved the create
				// receipt/ID, before binding or semantic context advanced. The intercepted
				// stage request never reached the real server.
				var original workPublicationIntent
				if err := json.Unmarshal(pending.Context, &original); err != nil {
					t.Fatal(err)
				}
				original.Command.WorkID = ""
				original.Phase = "create"
				original.Step = 0
				original.SessionVersion = 0
				pending.Context, err = json.Marshal(original)
				if err != nil {
					t.Fatal(err)
				}
				pending.Steps = pending.Steps[:1]
				journal.Bindings = []viewWorkBinding{}
				if err := writeWorkJournal(dir, journal); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Remove(filepath.Join(f.change, "tasks.md")); err != nil {
				t.Fatal(err)
			}
			result, err := f.run(t, args...)
			if err != nil {
				t.Fatal(err)
			}
			s := f.observe(t, result.WorkID)
			if len(s.Documents) != 2 {
				t.Fatal("original creation selectors rescanned live files")
			}
			index, err := viewIndex(f.info)
			if err != nil {
				t.Fatal(err)
			}
			if len(index.TimelineSessions) != 1 {
				t.Fatal("receipt recovery duplicated work")
			}
		})
	}
}
