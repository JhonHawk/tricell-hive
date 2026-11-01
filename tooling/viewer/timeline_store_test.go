package viewer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTimelineFixture(t *testing.T) (*Store, TimelineCreateSessionRequest) {
	t.Helper()
	state := t.TempDir()
	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := RegisterRepository(state, repo); err != nil {
		t.Fatal(err)
	}
	s, err := Open(Options{StateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	return s, TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "create", Source: TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "native-1"}, Title: "A bounded session", Summary: "Test publication", Tickets: []TimelineTicket{}}
}
func createTimeline(t *testing.T, s *Store, r TimelineCreateSessionRequest) TimelineSession {
	t.Helper()
	v, err := s.TimelineCreate(r)
	if err != nil {
		t.Fatal(err)
	}
	return v.Session
}
func timelineString(v string) *string                        { return &v }
func timelineInt(v int64) *int64                             { return &v }
func timelineState(v TimelineStageState) *TimelineStageState { return &v }
func requireTimelineCode(t *testing.T, err error, code string) {
	t.Helper()
	var e *TimelineError
	var c *TimelineContractError
	if errors.As(err, &e) && e.Code == code {
		return
	}
	if errors.As(err, &c) && c.Code == code {
		return
	}
	t.Fatalf("wanted %s, got %v", code, err)
}
func TestTimelineRuntimeIdentityConcurrentRestartAndBindings(t *testing.T) {
	s, request := newTimelineFixture(t)
	s.remote = func(string) (string, error) { return "git@github.example:team/repository.git", nil }
	var wg sync.WaitGroup
	ids := make(chan string, 12)
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := s.TimelineCreate(request)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- r.Session.ViewerSessionID
		}()
	}
	wg.Wait()
	close(ids)
	id := ""
	for value := range ids {
		if id != "" && id != value {
			t.Fatal("duplicate tuple")
		}
		id = value
	}
	if len(s.timeline) != 1 {
		t.Fatal("duplicate records")
	}
	state := filepath.Dir(s.Dir())
	restarted, err := Open(Options{StateDir: state, GitRemote: func(string) (string, error) { return "", errors.New("temporarily unavailable") }})
	if err != nil {
		t.Fatal(err)
	}
	request.OperationID = "resume"
	response, err := restarted.TimelineCreate(request)
	if err != nil || response.Session.ViewerSessionID != id || *response.Created {
		t.Fatalf("resume %v %v", response, err)
	}
	restarted.remote = func(string) (string, error) { return "https://github.example/team/changed.git", nil }
	request.OperationID = "changed"
	_, err = restarted.TimelineCreate(request)
	requireTimelineCode(t, err, "project_identity_changed")
	alias := filepath.Join(t.TempDir(), "worktree")
	if err = os.MkdirAll(filepath.Join(alias, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err = RegisterRepository(state, alias); err != nil {
		t.Fatal(err)
	}
	restarted.remote = func(string) (string, error) { return "https://github.example/team/repository", nil }
	request.Source.ProjectRoot = alias
	request.OperationID = "alias"
	response, err = restarted.TimelineCreate(request)
	if err != nil || response.Session.ViewerSessionID != id {
		t.Fatalf("equivalent worktree split %v", err)
	}
	restarted.remote = func(string) (string, error) { return "", errors.New("offline") }
	request.OperationID = "alias-offline"
	if _, err = restarted.TimelineCreate(request); err != nil {
		t.Fatal(err)
	}
}
func TestTimelineRuntimeStageDocumentEvidenceAndReplay(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	id := v.ViewerSessionID
	stage, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 1, OperationID: "plan", Kind: "stage", Title: "Plan", Summary: "Prepare", TicketRefs: []TimelineTicketRef{}, Stage: "plan", StageState: "in_progress", ExpectedSessionVersion: timelineInt(v.Version)})
	if err != nil {
		t.Fatal(err)
	}
	metadata := TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "doc", MediaType: "text/markdown"}
	content := []byte("\xef\xbb\xbf# Plan\n")
	doc, err := s.TimelineAddDocument(id, metadata, content)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.TimelineAddDocument(id, metadata, content)
	if err != nil || again.DocumentID != doc.DocumentID || again.SessionVersion != doc.SessionVersion {
		t.Fatal("replay changed receipt", err)
	}
	_, err = s.TimelineAddDocument(id, metadata, []byte("changed"))
	requireTimelineCode(t, err, "operation_conflict")
	snapshot, err := s.TimelineSnapshot(id, stage.EventID, TimelineDocumentSnapshotRequest{SchemaVersion: 1, OperationID: "snapshot", ExpectedVersion: 1, Status: "complete", Tabs: []TimelineDocumentTab{{Label: "Plan", DocumentID: doc.DocumentID}}})
	if err != nil {
		t.Fatal(err)
	}
	read, err := s.TimelineDocument(id, doc.DocumentID)
	if err != nil || read.Content != string(content) {
		t.Fatal("immutable document changed", err)
	}
	current, _ := s.TimelineGet(id)
	wait := TimelineAwaitingValidation
	result, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 1, OperationID: "wait", ExpectedVersion: current.Version, State: &wait})
	if err != nil {
		t.Fatal(err)
	}
	if result.CurrentStageEventID != nil {
		t.Fatal("foreground remained active")
	}
	current, _ = s.TimelineGet(id)
	if current.Events[0].StageState != "paused" {
		t.Fatal("stage not paused atomically")
	}
	resumed, err := s.TimelinePatchEvent(id, stage.EventID, TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "resume-stage", ExpectedVersion: current.Events[0].Version, ExpectedSessionVersion: timelineInt(current.Version), StageState: timelineState("in_progress")})
	if err != nil || resumed.CurrentStageEventID == nil {
		t.Fatal("resume", err)
	}
	current, _ = s.TimelineGet(id)
	_, err = s.TimelinePatchEvent(id, stage.EventID, TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "stale", ExpectedVersion: 1, Title: timelineString("stale")})
	requireTimelineCode(t, err, "version_conflict")
	done, err := s.TimelinePatchEvent(id, stage.EventID, TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "done", ExpectedVersion: current.Events[0].Version, ExpectedSessionVersion: timelineInt(current.Version), StageState: timelineState("completed")})
	if err != nil {
		t.Fatal(err)
	}
	current, _ = s.TimelineGet(id)
	if current.Events[0].Sequence != stage.Sequence || current.Events[0].DocumentSnapshots[0].DocumentRevisionID != snapshot.DocumentRevisionID {
		t.Fatal("status changed evidence/order")
	}
	_, err = s.TimelinePatchEvent(id, stage.EventID, TimelineEventPatchRequest{SchemaVersion: 1, OperationID: "terminal", ExpectedVersion: done.EventVersion, ExpectedSessionVersion: timelineInt(done.SessionVersion), StageState: timelineState("in_progress")})
	requireTimelineCode(t, err, "stage_terminal")
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	read, err = restarted.TimelineDocument(id, doc.DocumentID)
	if err != nil || read.Content != string(content) {
		t.Fatal("restart lost immutable data", err)
	}
	replay, err := restarted.TimelineSnapshot(id, stage.EventID, TimelineDocumentSnapshotRequest{SchemaVersion: 1, OperationID: "snapshot", ExpectedVersion: 1, Status: "complete", Tabs: []TimelineDocumentTab{{Label: "Plan", DocumentID: doc.DocumentID}}})
	if err != nil || replay.DocumentRevisionID != snapshot.DocumentRevisionID {
		t.Fatal("snapshot replay before stale validation", err)
	}
}
func TestTimelineRuntimeDurableWriteFailureRecovery(t *testing.T) {
	for _, phase := range []string{"before_file_sync", "after_file_sync", "before_rename", "after_rename", "before_directory_sync", "after_directory_sync"} {
		t.Run(phase, func(t *testing.T) {
			s, request := newTimelineFixture(t)
			v := createTimeline(t, s, request)
			id := v.ViewerSessionID
			renamed := false
			s.timelineFault = func(at, path string) error {
				if at == "after_rename" && strings.HasSuffix(path, id+".json") {
					renamed = true
				}
				if at == phase && (strings.HasSuffix(path, id+".json") || renamed) {
					return errors.New("injected")
				}
				return nil
			}
			patch := TimelineSessionPatchRequest{SchemaVersion: 1, OperationID: "patch", ExpectedVersion: 1, Title: timelineString("Changed")}
			_, err := s.TimelinePatch(id, patch)
			if err == nil {
				t.Fatal("failed durable write reported success")
			}
			s.timelineFault = nil
			current, err := s.TimelineGet(id)
			if err != nil {
				t.Fatal(err)
			}
			if renamed {
				if current.Title != "Changed" {
					t.Fatal("post-rename outcome not reconciled")
				}
			} else if current.Title != v.Title {
				t.Fatal("pre-rename failure mutated record")
			}
			result, err := s.TimelinePatch(id, patch)
			if err != nil || result.SessionVersion != 2 {
				t.Fatalf("retry: %v %v", result, err)
			}
		})
	}
}
func TestTimelineRuntimeImageAndDeleteTombstone(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	id := v.ViewerSessionID
	var data bytes.Buffer
	if err := png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	meta := TimelineArtifactMetadata{SchemaVersion: 1, OperationID: "image", CapturedAt: "2026-01-01T00:00:00Z", Candidate: TimelineCandidate{Branch: "feature", Commit: "abc", Base: "development"}, Viewport: "desktop", Theme: "dark", MediaType: "image/png"}
	imageResult, err := s.TimelineAddArtifact(id, meta, data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	meta.OperationID = "reuse"
	reused, err := s.TimelineAddArtifact(id, meta, data.Bytes())
	if err != nil || *reused.Created || reused.ArtifactID != imageResult.ArtifactID {
		t.Fatal("content not deduplicated", err)
	}
	raw, _, err := s.TimelineArtifact(id, "artifacts/"+imageResult.ArtifactID+".png")
	if err != nil || !bytes.Equal(raw, data.Bytes()) {
		t.Fatal(err)
	}
	current, _ := s.TimelineGet(id)
	del := TimelineMutationRequest{SchemaVersion: 1, OperationID: "delete", ExpectedVersion: current.Version}
	s.timelineFault = func(phase, path string) error {
		if phase == "before_delete_blobs" {
			return errors.New("injected")
		}
		return nil
	}
	_, err = s.TimelineDelete(id, del)
	requireTimelineCode(t, err, "delete_pending")
	s.timelineFault = nil
	s, err = Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := s.TimelineDelete(id, del)
	if err != nil || deleted.Deleted == nil || !*deleted.Deleted {
		t.Fatal("delete retry", err)
	}
	replay, err := s.TimelineDelete(id, del)
	if err != nil || replay.SessionVersion != deleted.SessionVersion {
		t.Fatal("lost delete response replay", err)
	}
	del.ExpectedVersion++
	_, err = s.TimelineDelete(id, del)
	requireTimelineCode(t, err, "operation_conflict")
	request.OperationID = "resurrect"
	_, err = s.TimelineCreate(request)
	requireTimelineCode(t, err, "session_deleted")
	_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 1, OperationID: "unknown", ExpectedVersion: 1, Title: timelineString("new")})
	requireTimelineCode(t, err, "session_deleted")
}
func TestTimelineRuntimeCorruptionAndNamespaceLinks(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	record := s.timelineRecordPath(v.ViewerSessionID)
	if err := os.WriteFile(record, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(Options{StateDir: filepath.Dir(s.Dir())}); err == nil {
		t.Fatal("corruption accepted")
	}
	if raw, _ := os.ReadFile(record); string(raw) != "{" {
		t.Fatal("corrupt evidence altered")
	}
}
func TestTimelineRuntimeHTTPAuthenticationAndStrictInput(t *testing.T) {
	s, request := newTimelineFixture(t)
	srv, err := NewServer(ServerOptions{Store: s, Token: strings.Repeat("x", 32), Port: 4483})
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(request)
	call := func(method, path string, body []byte, token, origin, scope bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:4483"+path, bytes.NewReader(body))
		if token {
			r.Header.Set(tokenHeader, strings.Repeat("x", 32))
		}
		if origin {
			r.Header.Set("Origin", "http://127.0.0.1:4483")
		}
		if scope {
			r.Header.Set("X-Hive-Action-Scope", "session")
		}
		w := httptest.NewRecorder()
		srv.ServeHTTP(w, r)
		return w
	}
	for _, flags := range [][3]bool{{false, true, true}, {true, false, true}, {true, true, false}} {
		if w := call("POST", "/api/timeline/sessions", body, flags[0], flags[1], flags[2]); w.Code < 400 {
			t.Fatalf("unauthorized write %d", w.Code)
		}
	}
	if w := call("POST", "/api/timeline/sessions", append(body, []byte(` {}`)...), true, true, true); w.Code != 400 {
		t.Fatalf("trailing JSON %d", w.Code)
	}
	w := call("POST", "/api/timeline/sessions", body, true, true, true)
	if w.Code != 200 {
		t.Fatalf("create %d %s", w.Code, w.Body.String())
	}
	response, err := DecodeTimelineSessionResponse(w.Body.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	w = call("GET", "/api/index", nil, true, false, false)
	var index struct {
		TimelineSessions []TimelineAPIIndexEntry `json:"timeline_sessions"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &index); err != nil || len(index.TimelineSessions) != 1 {
		t.Fatal("timeline index absent", err)
	}
	if w := call("GET", "/api/timeline/sessions/"+response.Session.ViewerSessionID, nil, false, false, false); w.Code != 403 {
		t.Fatal("unauthorized read")
	}
}
func TestTimelineRuntimeManualRetentionAndExpectedInventory(t *testing.T) {
	e, _ := groupFixture(t)
	members := e.sessions()
	expected := []ExpectedSessionMember{}
	for _, m := range members {
		expected = append(expected, ExpectedSessionMember{m.Key, m.ContentHash})
	}
	e.advance(365 * 24 * time.Hour)
	if _, err := e.store.Purge(); err != nil {
		t.Fatal(err)
	}
	if len(e.sessions()) != len(members) {
		t.Fatal("age deleted content")
	}
	wrong := append([]ExpectedSessionMember{}, expected...)
	wrong[0].ContentHash = strings.Repeat("0", 64)
	err := e.store.DeleteExpected(members[0].Key, wrong)
	requireTimelineCode(t, err, "members_changed")
	if len(e.sessions()) != len(members) {
		t.Fatal("mismatch removed a member")
	}
	if err = e.store.DeleteExpected(members[0].Key, expected); err != nil {
		t.Fatal(err)
	}
	if len(e.sessions()) != 0 {
		t.Fatal("matched inventory not removed")
	}
	if _, err = e.store.Sync(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(e.sessions()) != 0 {
		t.Fatal("deleted inventory resurrected")
	}
}

func TestTimelineRuntimeLegacyAuthorityAndTombstoneSuppression(t *testing.T) {
	e, repo := groupFixture(t)
	e.remotes[repo] = "" // An explicit successful lookup establishes a remote-less project.
	legacy := e.sessions()[0]
	request := TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "own", Source: TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "shared"}, Title: "API authority", Summary: "Own this tuple", Tickets: []TimelineTicket{}}
	response, err := e.store.TimelineCreate(request)
	if err != nil {
		t.Fatal(err)
	}
	if !e.store.timelineLegacyHidden(legacy) {
		t.Fatal("legacy duplicate not suppressed")
	}
	source := readManifestMap(t, legacy.Source)
	source["title"] = "Changed legacy"
	writeManifestMap(t, legacy.Source, source)
	e.touch(legacy.Source)
	e.sync()
	current, err := e.store.TimelineGet(response.Session.ViewerSessionID)
	if err != nil || current.Title != "API authority" {
		t.Fatal("scanner changed API authority", err)
	}
	_, err = e.store.TimelineDelete(current.ViewerSessionID, TimelineMutationRequest{SchemaVersion: 1, OperationID: "delete-owned", ExpectedVersion: current.Version})
	if err != nil {
		t.Fatal(err)
	}
	source["title"] = "Another legacy update"
	writeManifestMap(t, legacy.Source, source)
	e.touch(legacy.Source)
	e.sync()
	if !e.store.timelineLegacyHidden(legacy) {
		t.Fatal("deleted identity lost suppression")
	}
}
func TestTimelineRuntimeCapacityKeepsLifecycleAndDeletionAvailable(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	id := v.ViewerSessionID
	// Distinct valid receipts fill only content capacity, preserving terminal space.
	s.mu.Lock()
	r := timelineClone(s.timeline[id])
	sample := r.Receipts[0]
	for i := len(r.Receipts); i < TimelineContentReceipts; i++ {
		receipt := sample
		receipt.OperationID = fmt.Sprintf("fill-%d", i)
		receipt.Result.OperationID = receipt.OperationID
		r.Receipts = append(r.Receipts, receipt)
	}
	err := s.timelineSave(r)
	s.mu.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 1, OperationID: "content-full", Kind: "request", Title: "Too much", Summary: "Bounded", TicketRefs: []TimelineTicketRef{}})
	requireTimelineCode(t, err, "capacity")
	closure, err := s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 1, OperationID: "close-full", Kind: "closure", Title: "Finished", Summary: "Done", TicketRefs: []TimelineTicketRef{}, Closure: &TimelineClosure{Reason: "completed", Result: "Complete"}})
	if err != nil {
		t.Fatal("reserved closure capacity", err)
	}
	deleted, err := s.TimelineDelete(id, TimelineMutationRequest{SchemaVersion: 1, OperationID: "delete-full", ExpectedVersion: closure.SessionVersion})
	if err != nil || deleted.Deleted == nil {
		t.Fatal("reserved delete", err)
	}
}
func TestTimelineRuntimeInterruptedBlobNeverPublishesReference(t *testing.T) {
	for _, phase := range []string{"before_file_sync", "after_file_sync", "before_rename", "after_rename", "before_directory_sync", "after_directory_sync"} {
		t.Run(phase, func(t *testing.T) {
			s, request := newTimelineFixture(t)
			v := createTimeline(t, s, request)
			id := v.ViewerSessionID
			if err := s.timelineEnsureDir(s.timelineBlobDir(id)); err != nil {
				t.Fatal(err)
			}
			if err := s.timelineEnsureDir(filepath.Join(s.timelineBlobDir(id), "documents")); err != nil {
				t.Fatal(err)
			}
			renamed := false
			s.timelineFault = func(at, path string) error {
				if at == "after_rename" && strings.HasSuffix(path, ".md") {
					renamed = true
				}
				if at == phase && (strings.HasSuffix(path, ".md") || renamed) {
					return errors.New("injected")
				}
				return nil
			}
			meta := TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "interrupt-doc", MediaType: "text/markdown"}
			_, err := s.TimelineAddDocument(id, meta, []byte("# Safe"))
			if err == nil {
				t.Fatal("failed blob publication reported success")
			}
			s.timelineFault = nil
			current, err := s.TimelineGet(id)
			if err != nil || len(current.Documents) != 0 {
				t.Fatal("record referenced incomplete artifact", err)
			}
			if _, err = s.TimelineAddDocument(id, meta, []byte("# Safe")); err != nil {
				t.Fatal("retry failed", err)
			}
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			files, err := os.ReadDir(filepath.Join(s.timelineBlobDir(id), "documents"))
			if err != nil || len(files) != 1 {
				t.Fatal("uncommitted staging not reclaimed", err, len(files))
			}
		})
	}
}
func TestTimelineRuntimeReturnedValuesCannotMutateAuthority(t *testing.T) {
	s, request := newTimelineFixture(t)
	request.Tickets = []TimelineTicket{{Key: "A-1"}}
	response, err := s.TimelineCreate(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Session.Tickets[0].Key = "modified"
	request.Tickets[0].Key = "also-modified"
	current, err := s.TimelineGet(response.Session.ViewerSessionID)
	if err != nil || current.Tickets[0].Key != "A-1" {
		t.Fatal("caller mutated authority", err)
	}
}

func TestTimelineRuntimeRegisteredNonGitWorkspaceDoesNotBorrowRemote(t *testing.T) {
	state := t.TempDir()
	outer := filepath.Join(t.TempDir(), "outer")
	workspace := filepath.Join(outer, "workspace")
	work := filepath.Join(workspace, "_support", "workspace", "run")
	if err := os.MkdirAll(filepath.Join(outer, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatal(err)
	}
	registration, err := RegisterRepository(state, work)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(Options{StateDir: state, GitRemote: func(root string) (string, error) {
		t.Fatalf("non-Git workspace borrowed remote reader: %s", root)
		return "", nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	response, err := s.TimelineCreate(TimelineCreateSessionRequest{SchemaVersion: 1, OperationID: "workspace", Source: TimelineSource{ProjectRoot: registration.Root, Host: "codex", NativeSessionID: "workspace-native"}, Title: "Workspace", Summary: "No borrowed identity", Tickets: []TimelineTicket{}})
	if err != nil || response.Session.ProjectKey != registration.Root {
		t.Fatal("remote-less workspace failed", err)
	}
}

func TestTimelineRuntimeDirectoryCreationSyncRetries(t *testing.T) {
	for _, phase := range []string{"before_directory_sync", "after_directory_sync"} {
		for _, kind := range []string{"namespace", "blob-session", "blob-document"} {
			t.Run(kind+"/"+phase, func(t *testing.T) {
				s, request := newTimelineFixture(t)
				id := timelineSessionKey(request.Source.ProjectRoot, request.Source)
				target := s.timelineDir()
				var operation func() error
				if kind == "namespace" {
					operation = func() error { _, err := s.TimelineCreate(request); return err }
				} else {
					id = createTimeline(t, s, request).ViewerSessionID
					target = s.timelineBlobDir(id)
					if kind == "blob-document" {
						target = filepath.Join(target, "documents")
					}
					operation = func() error {
						_, err := s.TimelineAddDocument(id, TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "directory-retry", MediaType: "text/markdown"}, []byte("# Durable"))
						return err
					}
				}
				parent := filepath.Dir(target)
				attempts := 0
				s.timelineFault = func(at, path string) error {
					if at == phase && path == parent {
						attempts++
						return errors.New("injected parent sync failure")
					}
					return nil
				}
				if err := operation(); err == nil {
					t.Fatal("initial directory sync failure returned success")
				}
				if info, err := os.Lstat(target); err != nil || !info.IsDir() {
					t.Fatal("regression requires mkdir to have succeeded", err)
				}
				if err := operation(); err == nil {
					t.Fatal("existing directory bypassed failed parent sync on retry")
				}
				if attempts != 2 {
					t.Fatalf("parent sync attempts=%d, want2", attempts)
				}
				s.timelineFault = nil
				if err := operation(); err != nil {
					t.Fatal("successful parent sync did not permit retry", err)
				}
				if kind != "namespace" {
					current, err := s.TimelineGet(id)
					if err != nil || len(current.Documents) != 1 {
						t.Fatal("retry duplicated or lost document", err)
					}
				}
			})
		}
	}
}

func TestTimelineRuntimeViewerAncestorsSyncedBeforeFirstRecord(t *testing.T) {
	for _, phase := range []string{"before_directory_sync", "after_directory_sync"} {
		t.Run(phase, func(t *testing.T) {
			s, request := newTimelineFixture(t)
			viewer, err := filepath.EvalSymlinks(s.vdir)
			if err != nil {
				t.Fatal(err)
			}
			state := filepath.Dir(viewer)
			calls := 0
			s.timelineFault = func(at, path string) error {
				if at == phase && path == state {
					calls++
					return errors.New("injected viewer ancestor sync failure")
				}
				return nil
			}
			for range 2 {
				if _, err = s.TimelineCreate(request); err == nil {
					t.Fatal("unsynced viewer ancestor acknowledged")
				}
			}
			if calls != 2 {
				t.Fatalf("ancestor sync not retried: %d", calls)
			}
			synced := map[string]bool{}
			s.timelineFault = func(at, path string) error {
				if at == "after_directory_sync" {
					synced[path] = true
				}
				return nil
			}
			if _, err = s.TimelineCreate(request); err != nil {
				t.Fatal(err)
			}
			for ancestor := viewer; ; {
				parent := filepath.Dir(ancestor)
				if parent == ancestor {
					break
				}
				if !synced[parent] {
					t.Fatalf("Open-created ancestor not synchronized: %s", parent)
				}
				ancestor = parent
			}
		})
	}
}

func TestTimelineRuntimePrivateReceiptEnvelopeExceedsPublicBodyLimit(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	id := v.ViewerSessionID
	metadata := TimelineDocumentMetadata{SchemaVersion: 1, OperationID: "large-record-document", MediaType: "text/markdown"}
	document, err := s.TimelineAddDocument(id, metadata, []byte("# Retained document\n"))
	if err != nil {
		t.Fatal(err)
	}
	r := timelineClone(s.timeline[id])
	now := s.timelineTime()
	for len(r.Session.Events) < TimelineMaxEvents {
		n := len(r.Session.Events) + 1
		r.Session.Events = append(r.Session.Events, TimelineEvent{EventID: fmt.Sprintf("large-event-%d", n), Sequence: int64(n), Version: 1, Kind: "request", Title: "A meaningful milestone", Summary: strings.Repeat("s", 4000), CreatedAt: now, UpdatedAt: now, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}})
		raw, _ := json.Marshal(r.Session)
		if len(raw) >= TimelineContentBudget-8192 {
			break
		}
	}
	public, err := json.Marshal(r.Session)
	if err != nil {
		t.Fatal(err)
	}
	if len(public) > TimelineContentBudget {
		t.Fatalf("fixture exceeds content budget: %d", len(public))
	}
	sample := r.Receipts[0]
	for len(r.Receipts) < TimelineMaxReceipts {
		i := len(r.Receipts)
		receipt := sample
		receipt.OperationID = fmt.Sprintf("near-capacity-receipt-%d", i)
		receipt.Result.OperationID = receipt.OperationID
		r.Receipts = append(r.Receipts, receipt)
	}
	private, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(private) <= TimelineMaxBody || len(private) > timelineMaxRecordBytes {
		t.Fatalf("fixture must exercise distinct limits: public=%d private=%d", len(public), len(private))
	}
	t.Logf("public session=%d bytes; private receipt envelope=%d bytes", len(public), len(private))
	s.mu.Lock()
	err = s.timelineSave(r)
	s.mu.Unlock()
	if err != nil {
		t.Fatal("valid private envelope could not save", err)
	}
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatalf("valid public%d/private%d could not reopen: %v", len(public), len(private), err)
	}
	current, err := reopened.TimelineGet(id)
	if err != nil || len(current.Events) != len(r.Session.Events) {
		t.Fatal("near-capacity session lost on restart", err)
	}
	read, err := reopened.TimelineDocument(id, document.DocumentID)
	if err != nil || read.Content != "# Retained document\n" {
		t.Fatal("document recovery failed", err)
	}
	replay, err := reopened.TimelineAddDocument(id, metadata, []byte("# Retained document\n"))
	if err != nil || replay.DocumentID != document.DocumentID || replay.SessionVersion != document.SessionVersion {
		t.Fatal("receipt recovery failed", err)
	}
	var decoded timelineRecord
	if err = DecodeTimelineJSON(private, &decoded); err == nil {
		t.Fatal("public HTTP decoder accepted a private envelope over2MiB")
	}
	if err = decodeTimelineJSONBounded(private, &decoded, timelineMaxRecordBytes); err != nil {
		t.Fatal("private bounded decoder rejected valid record", err)
	}
}

func TestTimelineRuntimePrivateRecordLimitAndStrictDecoder(t *testing.T) {
	s, request := newTimelineFixture(t)
	v := createTimeline(t, s, request)
	r := timelineClone(s.timeline[v.ViewerSessionID])
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutated := range [][]byte{
		bytes.Replace(raw, []byte(`"storage_version":1`), []byte(`"storage_version":1,"storage_version":1`), 1),
		bytes.Replace(raw, []byte(`"storage_version":1`), []byte(`"Storage_version":1`), 1),
		bytes.Replace(raw, []byte(`"storage_version":1`), []byte(`"storage_version":1,"unknown":true`), 1),
	} {
		var decoded timelineRecord
		if err = decodeTimelineJSONBounded(mutated, &decoded, timelineMaxRecordBytes); err == nil {
			t.Fatal("private decoder weakened exact/duplicate-field checks")
		}
	}
	oversized := append(append([]byte{}, raw...), bytes.Repeat([]byte(" "), timelineMaxRecordBytes-len(raw)+1)...)
	var decoded timelineRecord
	if err = decodeTimelineJSONBounded(oversized, &decoded, timelineMaxRecordBytes); err == nil {
		t.Fatal("private decoder accepted more than4MiB")
	}
	if err = os.WriteFile(s.timelineRecordPath(v.ViewerSessionID), oversized, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(Options{StateDir: filepath.Dir(s.Dir())}); err == nil {
		t.Fatal("startup accepted private record over4MiB")
	}
	info, err := os.Stat(s.timelineRecordPath(v.ViewerSessionID))
	if err != nil || info.Size() != int64(len(oversized)) {
		t.Fatal("failed startup altered oversized evidence", err)
	}
	publicRequest, _ := json.Marshal(TimelineMutationRequest{SchemaVersion: 1, OperationID: "public-bound", ExpectedVersion: 1})
	publicRequest = append(publicRequest, bytes.Repeat([]byte(" "), TimelineMaxBody-len(publicRequest)+1)...)
	if _, err = DecodeTimelineMutation(publicRequest); err == nil {
		t.Fatal("public decoder no longer enforces2MiB")
	}
}
