package viewer

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type stateEvidence struct {
	Mode     os.FileMode
	Size     int64
	Modified int64
	Hash     string
}

func stateManifest(t *testing.T, dir string) map[string]stateEvidence {
	t.Helper()
	out := map[string]stateEvidence{}
	err := filepath.WalkDir(dir, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := os.Lstat(p)
		if err != nil {
			return err
		}
		v := stateEvidence{Mode: i.Mode(), Size: i.Size(), Modified: i.ModTime().UnixNano()}
		if i.Mode().IsRegular() {
			raw, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			v.Hash = fmt.Sprintf("%x", sha256.Sum256(raw))
		}
		rel, _ := filepath.Rel(dir, p)
		out[rel] = v
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Previous-generation state is not recognized any more: it falls into the strict
// inventory and fails with the generic invalid_storage, before anything is written.
func TestV3PreviousStateFailsInvalidStorageWithoutWriting(t *testing.T) {
	setup := map[string]func(dir string){
		"index.json":         func(dir string) { os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"version":1}`), 0600) },
		"sessions":           func(dir string) { os.Mkdir(filepath.Join(dir, "sessions"), 0700) },
		"session-groups":     func(dir string) { os.Mkdir(filepath.Join(dir, "session-groups"), 0700) },
		"timeline-api-v1":    func(dir string) { os.Mkdir(filepath.Join(dir, "timeline-api-v1"), 0700) },
		"timeline-api-v2":    func(dir string) { os.Mkdir(filepath.Join(dir, "timeline-api-v2"), 0700) },
		"registry-no-marker": func(dir string) { os.WriteFile(filepath.Join(dir, "registry.json"), []byte(`{"version":1}`), 0600) },
		"marker-v1":          func(dir string) { os.WriteFile(filepath.Join(dir, stateMarker), []byte(`{"storage_version":1}`), 0600) },
		"marker-v2":          func(dir string) { os.WriteFile(filepath.Join(dir, stateMarker), []byte(`{"storage_version":2}`), 0600) },
		"marker-v3-with-index": func(dir string) {
			os.WriteFile(filepath.Join(dir, stateMarker), []byte(`{"storage_version":3}`), 0600)
			os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"version":1}`), 0600)
		},
	}
	for name, prepare := range setup {
		t.Run(name, func(t *testing.T) {
			state := t.TempDir()
			dir := viewerDir(state)
			os.Mkdir(dir, 0700)
			prepare(dir)
			before := stateManifest(t, state)
			requireTimelineCode(t, ValidateState(state), "invalid_storage")
			_, err := Open(Options{StateDir: state})
			requireTimelineCode(t, err, "invalid_storage")
			_, err = RegisterRepository(state, newRepo(t))
			requireTimelineCode(t, err, "invalid_storage")
			if !reflect.DeepEqual(before, stateManifest(t, state)) {
				t.Fatal("rejected state was mutated")
			}
		})
	}
}
func TestV3UnknownCorruptInventoryPreserved(t *testing.T) {
	for _, scenario := range []string{"unmarked", "bad-marker", "unknown-file", "bad-record", "symlink", "hardlink", "bad-registry", "unknown-blob"} {
		t.Run(scenario, func(t *testing.T) {
			state := t.TempDir()
			s, err := Open(Options{StateDir: state})
			if err != nil {
				t.Fatal(err)
			}
			switch scenario {
			case "unmarked":
				os.Remove(filepath.Join(s.Dir(), stateMarker))
				os.WriteFile(filepath.Join(s.Dir(), "server.json"), []byte("{}"), 0600)
			case "bad-marker":
				os.WriteFile(filepath.Join(s.Dir(), stateMarker), []byte(`{"storage_version":3} {}`), 0600)
			case "unknown-file":
				os.WriteFile(filepath.Join(s.Dir(), "unknown"), nil, 0600)
			case "bad-record":
				os.Mkdir(s.timelineDir(), 0700)
				os.WriteFile(s.timelineRecordPath(string64()), []byte("{}"), 0600)
			case "symlink":
				os.Symlink(t.TempDir(), s.timelineDir())
			case "hardlink":
				outside := filepath.Join(t.TempDir(), "outside")
				os.WriteFile(outside, []byte(`{"version":2,"repositories":[]}`), 0600)
				os.Link(outside, filepath.Join(s.Dir(), "registry.json"))
			case "bad-registry":
				os.WriteFile(filepath.Join(s.Dir(), "registry.json"), []byte(`{"version":2,"repositories":[]} {}`), 0600)
			case "unknown-blob":
				os.MkdirAll(filepath.Join(s.timelineDir(), string64(), "documents"), 0700)
				os.WriteFile(filepath.Join(s.timelineDir(), string64(), "documents", "foreign"), nil, 0600)
			}
			before := stateManifest(t, state)
			if err = s.Startup(); err == nil {
				t.Fatal("accepted invalid state")
			}
			if !reflect.DeepEqual(before, stateManifest(t, state)) {
				t.Fatal("changed invalid state")
			}
		})
	}
}
func string64() string { return fmt.Sprintf("%064d", 1) }
func newV3Fixture(t *testing.T) (*Store, TimelineSession, TimelineCreateSessionRequest) {
	t.Helper()
	state := t.TempDir()
	repo := newRepo(t)
	registration, err := RegisterRepository(state, repo)
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(Options{StateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	request := TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "create", Source: TimelineSource{ProjectRoot: registration.Root, Host: "codex", NativeSessionID: "v2-native"}, Title: "V3 work", Summary: "Persistence fixture", Tickets: []TimelineTicketInput{}}
	response, err := s.TimelineCreate(request)
	if err != nil {
		t.Fatal(err)
	}
	return s, response.Session, request
}
func TestV3PersistenceUploadReplayRestart(t *testing.T) {
	s, v, request := newV3Fixture(t)
	actor := TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}
	upload := TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "doc", Actor: actor, MediaType: "text/markdown"}
	result, err := s.TimelineAddDocument(v.ViewerSessionID, upload, []byte("# Durable v2"))
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if err = restarted.Startup(); err != nil {
		t.Fatal(err)
	}
	replay, err := restarted.TimelineAddDocument(v.ViewerSessionID, upload, []byte("# Durable v2"))
	if err != nil || !reflect.DeepEqual(result, replay) {
		t.Fatalf("replay %v %v", replay, err)
	}
	response, err := restarted.TimelineCreate(request)
	if err != nil || *response.Created != true || response.Session.ViewerSessionID != v.ViewerSessionID {
		t.Fatalf("create replay %v", err)
	}
	doc, err := restarted.TimelineDocument(v.ViewerSessionID, result.DocumentID)
	if err != nil || doc.SchemaVersion != 3 || doc.Content != "# Durable v2" {
		t.Fatalf("read %v", err)
	}
	raw, _ := json.Marshal(response)
	if !json.Valid(raw) {
		t.Fatal("invalid response")
	}
	if _, err = os.Stat(filepath.Join(s.Dir(), "sessions")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("legacy copy directory exists")
	}
}
func TestV3DurableReceiptFaultRecovery(t *testing.T) {
	for _, phase := range []string{"before_file_sync", "after_file_sync", "before_rename", "after_rename", "before_directory_sync", "after_directory_sync"} {
		t.Run(phase, func(t *testing.T) {
			s, v, _ := newV3Fixture(t)
			metadata := TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "fault", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, MediaType: "text/markdown"}
			renamed := false
			s.timelineFault = func(at, path string) error {
				if path == s.timelineRecordPath(v.ViewerSessionID) && at == "after_rename" {
					renamed = true
				}
				if at == phase && (path == s.timelineRecordPath(v.ViewerSessionID) || renamed) {
					return errors.New("injected")
				}
				return nil
			}
			_, err := s.TimelineAddDocument(v.ViewerSessionID, metadata, []byte("# Receipt"))
			if err == nil {
				t.Fatal("fault succeeded")
			}
			if renamed {
				requireTimelineCode(t, err, "outcome_unknown")
			}
			s.timelineFault = nil
			result, err := s.TimelineAddDocument(v.ViewerSessionID, metadata, []byte("# Receipt"))
			if err != nil {
				t.Fatal(err)
			}
			restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
			if err != nil {
				t.Fatal(err)
			}
			replay, err := restarted.TimelineAddDocument(v.ViewerSessionID, metadata, []byte("# Receipt"))
			if err != nil || !reflect.DeepEqual(result, replay) {
				t.Fatal("receipt lost", err)
			}
		})
	}
}
func TestV3RunWaitsForCancellation(t *testing.T) {
	s, _, _ := newV3Fixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, RunOptions{PollInterval: time.Millisecond}) }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestV3ResolvedIdentityAndAliases(t *testing.T) {
	state := t.TempDir()
	repo := newRepo(t)
	registration, err := RegisterRepository(state, repo)
	if err != nil {
		t.Fatal(err)
	}
	remote := "https://github.example/team/project"
	s, err := Open(Options{StateDir: state, GitRemote: func(string) (string, error) { return remote, nil }})
	if err != nil {
		t.Fatal(err)
	}
	request := TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "remote", Source: TimelineSource{ProjectRoot: registration.Root, Host: "codex", NativeSessionID: "remote-native"}, Title: "Remote work", Summary: "Identity", Tickets: []TimelineTicketInput{}, WorkKey: timelineString("remote-work")}
	response, err := s.TimelineCreate(request)
	if err != nil {
		t.Fatal(err)
	}
	v := response.Session
	alias := newRepo(t)
	if _, err = RegisterRepository(filepath.Dir(s.Dir()), alias); err != nil {
		t.Fatal(err)
	}
	request.Source.ProjectRoot = alias
	request.OperationID = "alias"
	response, err = s.TimelineCreate(request)
	if err != nil || response.Session.ViewerSessionID != v.ViewerSessionID || response.Session.Source != v.Source {
		t.Fatal("alias changed provenance", err)
	}
	offline := func(string) (string, error) { return "", errors.New("offline") }
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir()), GitRemote: offline})
	if err != nil {
		t.Fatal(err)
	}
	request.OperationID = "offline"
	response, err = restarted.TimelineCreate(request)
	if err != nil || response.Session.ViewerSessionID != v.ViewerSessionID {
		t.Fatal("lost durable binding", err)
	}
	before := stateManifest(t, filepath.Dir(s.Dir()))
	_, err = Open(Options{StateDir: filepath.Dir(s.Dir()), GitRemote: func(string) (string, error) { return "https://github.example/team/changed", nil }})
	if err == nil {
		t.Fatal("changed identity accepted")
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("identity failure wrote state")
	}
}

func TestV3HTTPIndexAndRetiredRoutes(t *testing.T) {
	s, _, _ := newV3Fixture(t)
	server, err := NewServer(ServerOptions{Store: s, Token: "v2-test-secret-is-not-production", Port: 45555})
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path string, token bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "http://127.0.0.1:45555"+path, nil)
		r.Header.Set(TimelineRevisionHeader, TimelineRevisionValue)
		if token {
			r.Header.Set(tokenHeader, "v2-test-secret-is-not-production")
		}
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		return w
	}
	if w := request("GET", "/api/index", false); w.Code != 403 {
		t.Fatal("index leaked")
	}
	w := request("GET", "/api/index", true)
	if w.Code != 200 {
		t.Fatal(w.Code)
	}
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	keys := []string{}
	for key := range envelope {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	expected := []string{"errors", "protocol_version", "repositories", "retention", "timeline_sessions", "ttl_seconds"}
	if !reflect.DeepEqual(keys, expected) {
		t.Fatal(keys)
	}
	if string(envelope["protocol_version"]) != "3" || string(envelope["errors"]) != "[]" {
		t.Fatal("wrong v2 index")
	}
	before := stateManifest(t, filepath.Dir(s.Dir()))
	w = request("POST", "/api/sessions/"+string64()+"/pin", true)
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("retired route changed state")
	}
	w = request("GET", "/api/health", true)
	if !strings.Contains(w.Body.String(), `"timeline_protocol_version":3`) {
		t.Fatal("wrong protocol health")
	}
}

func TestV3ValidationPrecedesCleanup(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	orphan := filepath.Join(s.timelineDir(), ".timeline-write-incomplete")
	if err := os.WriteFile(orphan, []byte("incomplete"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(s.timelineRecordPath(string64()), []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	before := stateManifest(t, filepath.Dir(s.Dir()))
	if err := s.Startup(); err == nil {
		t.Fatal("corruption accepted")
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("cleanup ran before validation")
	}
	os.Remove(s.timelineRecordPath(string64()))
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(orphan); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("owned temporary not cleaned")
	}
	if _, err := s.TimelineGet(v.ViewerSessionID); err != nil {
		t.Fatal(err)
	}
}

func TestV3BlobHardlinksRejectedAtReadAndRestart(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	result, err := s.TimelineAddDocument(v.ViewerSessionID, TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "hardlink-doc", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, MediaType: "text/markdown"}, []byte("# Kept"))
	if err != nil {
		t.Fatal(err)
	}
	replaceWithHardlink(t, filepath.Join(s.timelineBlobDir(v.ViewerSessionID), "documents", result.DocumentID+".md"))
	before := stateManifest(t, filepath.Dir(s.Dir()))
	if _, err = s.TimelineDocument(v.ViewerSessionID, result.DocumentID); err == nil {
		t.Fatal("hardlinked blob read")
	}
	if _, err = Open(Options{StateDir: filepath.Dir(s.Dir())}); err == nil {
		t.Fatal("hardlinked blob restored")
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("hardlink rejection wrote state")
	}
}

func TestV3OriginalSourceAliasIsNeverRewritten(t *testing.T) {
	state := t.TempDir()
	repo := newRepo(t)
	if _, err := RegisterRepository(state, repo); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "repo-alias")
	if err := os.Symlink(repo, alias); err != nil {
		t.Fatal(err)
	}
	s, err := Open(Options{StateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	source := TimelineSource{ProjectRoot: alias, Host: "codex", NativeSessionID: "actual-native"}
	response, err := s.TimelineCreate(TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "alias-create", Source: source, Title: "Actual source", Summary: "Canonical identity, actual provenance", Tickets: []TimelineTicketInput{}})
	if err != nil {
		t.Fatal(err)
	}
	if response.Session.Source != source || response.Session.Producers[0].Source != source {
		t.Fatal("source rewritten")
	}
	reopened, err := Open(Options{StateDir: state})
	if err != nil {
		t.Fatal(err)
	}
	v, err := reopened.TimelineGet(response.Session.ViewerSessionID)
	if err != nil || v.Source != source {
		t.Fatal("source changed on restart", err)
	}
}

func TestV3PendingDeletionAndReceiptSurviveRestart(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	meta := TimelineDocumentMetadata{SchemaVersion: 3, OperationID: "pending-doc", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, MediaType: "text/markdown"}
	result, err := s.TimelineAddDocument(v.ViewerSessionID, meta, []byte("# To delete"))
	if err != nil {
		t.Fatal(err)
	}
	request := TimelineMutationRequest{SchemaVersion: 3, OperationID: "delete", ExpectedVersion: result.SessionVersion, Actor: TimelineActor{Kind: "local_viewer"}}
	s.timelineFault = func(phase, path string) error {
		if phase == "before_delete_blobs" {
			return errors.New("interrupted")
		}
		return nil
	}
	_, err = s.TimelineDelete(v.ViewerSessionID, request)
	requireTimelineCode(t, err, "delete_pending")
	// Simulate a crash in recursive deletion, after the pending marker committed.
	if err = os.RemoveAll(s.timelineBlobDir(v.ViewerSessionID)); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if err = reopened.Startup(); err != nil {
		t.Fatal(err)
	}
	deleted, err := reopened.TimelineDelete(v.ViewerSessionID, request)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err = Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	replay, err := reopened.TimelineDelete(v.ViewerSessionID, request)
	if err != nil || !reflect.DeepEqual(replay, deleted) {
		t.Fatal("delete replay lost", err)
	}
	if len(reopened.TimelineIndex()) != 0 {
		t.Fatal("deleted work resurfaced")
	}
}

func TestV3ManualRetentionNeverScansOrExpires(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	s.now = func() time.Time { return time.Now().Add(365 * 24 * time.Hour) }
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	current, err := s.TimelineGet(v.ViewerSessionID)
	if err != nil || !reflect.DeepEqual(current, v) {
		t.Fatal("age changed API content", err)
	}
}

func TestV3RestoredBindingSurvivesUnavailableSource(t *testing.T) {
	for _, scenario := range []string{"deleted-alias", "moved-repository"} {
		t.Run(scenario, func(t *testing.T) {
			s, original, request := newV3Fixture(t)
			alias := filepath.Join(t.TempDir(), "source-alias")
			if err := os.Symlink(original.Source.ProjectRoot, alias); err != nil {
				t.Fatal(err)
			}
			request.Source.ProjectRoot = alias
			request.Source.NativeSessionID = "alias-native"
			request.OperationID = "alias-source"
			response, err := s.TimelineCreate(request)
			if err != nil {
				t.Fatal(err)
			}
			expected := response.Session
			if scenario == "deleted-alias" {
				if err = os.Remove(alias); err != nil {
					t.Fatal(err)
				}
			} else {
				if err = os.Rename(original.Source.ProjectRoot, original.Source.ProjectRoot+"-temporarily-moved"); err != nil {
					t.Fatal(err)
				}
			}
			before := stateManifest(t, filepath.Dir(s.Dir()))
			reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
			if err != nil {
				t.Fatal("durable read blocked by unavailable source", err)
			}
			if err = reopened.Startup(); err != nil {
				t.Fatal(err)
			}
			current, err := reopened.TimelineGet(expected.ViewerSessionID)
			if err != nil || !reflect.DeepEqual(current, expected) {
				t.Fatal("restoration changed source or work", err)
			}
			request.OperationID = "new-publication"
			_, err = reopened.TimelineCreate(request)
			requireTimelineCode(t, err, "project_unavailable")
			if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
				t.Fatal("restoration or rejected publication wrote state")
			}
		})
	}
}

func TestV3UnavailableAliasDoesNotHideContradictoryIdentity(t *testing.T) {
	state := t.TempDir()
	repo := newRepo(t)
	if _, err := RegisterRepository(state, repo); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(repo, alias); err != nil {
		t.Fatal(err)
	}
	remote := "https://github.example/team/original"
	reader := func(string) (string, error) { return remote, nil }
	s, err := Open(Options{StateDir: state, GitRemote: reader})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.TimelineCreate(TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "bound", Source: TimelineSource{ProjectRoot: alias, Host: "codex", NativeSessionID: "native"}, Title: "Bound work", Summary: "Keep original provenance", Tickets: []TimelineTicketInput{}})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	remote = "https://github.example/team/contradictory"
	before := stateManifest(t, state)
	_, err = Open(Options{StateDir: state, GitRemote: reader})
	requireTimelineCode(t, err, "project_identity_changed")
	if !reflect.DeepEqual(before, stateManifest(t, state)) {
		t.Fatal("contradictory identity altered state")
	}
}

func replaceWithHardlink(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "linked-source")
	if err := os.WriteFile(outside, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(outside, path); err != nil {
		t.Fatal(err)
	}
}
