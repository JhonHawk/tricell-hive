package viewer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeGroupingConfig(t *testing.T, root, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, ".engram"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".engram", "config.json"), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPGroupedNavigationHistoricalNamesRemainOneFallbackGroup(t *testing.T) {
	s, request := newTimelineFixture(t)
	root := request.Source.ProjectRoot
	writeGroupingConfig(t, root, `{"project_name":"Historical old name"}`)
	first := createTimeline(t, s, request)
	writeGroupingConfig(t, root, `{"project_name":"Historical new name"}`)
	request.Source.NativeSessionID, request.OperationID = "renamed", "renamed"
	second := createTimeline(t, s, request)
	if first.ProjectKey != second.ProjectKey || first.ProjectName == second.ProjectName {
		t.Fatal("fixture must preserve different historical names for the same identity")
	}
	if err := os.Remove(filepath.Join(root, ".engram", "config.json")); err != nil {
		t.Fatal(err)
	}
	explicitRoot := newRepo(t)
	writeGroupingConfig(t, explicitRoot, `{"project_name":"Explicit group"}`)
	if _, err := RegisterRepository(filepath.Dir(s.Dir()), explicitRoot); err != nil {
		t.Fatal(err)
	}
	request.Source.ProjectRoot, request.Source.NativeSessionID, request.OperationID = explicitRoot, "explicit", "explicit"
	explicit := createTimeline(t, s, request)
	before := stateManifest(t, filepath.Dir(s.Dir()))
	legacy := s.TimelineIndex()
	expectedName := ""
	for _, session := range legacy {
		if session.ProjectKey == first.ProjectKey {
			expectedName = session.ProjectName
			break
		}
	}
	server, err := NewServer(ServerOptions{Store: s, Port: 4483, Token: strings.Repeat("z", 32)})
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest("GET", "http://127.0.0.1:4483/api/index?navigation=grouped", nil)
	r.Header.Set(tokenHeader, strings.Repeat("z", 32))
	r.Header.Set(TimelineRevisionHeader, TimelineRevisionValue)
	w := httptest.NewRecorder()
	server.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("index: %d %s", w.Code, w.Body.String())
	}
	var response struct {
		Sessions   []TimelineAPIIndexEntry `json:"timeline_sessions"`
		Navigation Navigation              `json:"navigation"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(response.Sessions, legacy) || len(response.Navigation.Entries) != 3 {
		t.Fatal("projection changed historical index")
	}
	for _, entry := range response.Navigation.Entries {
		if entry.ViewerSessionID == explicit.ViewerSessionID {
			if entry.ConfigStatus != "valid" || entry.GroupName != "Explicit group" || !strings.HasPrefix(entry.GroupKey, "explicit:") {
				t.Fatalf("explicit group lost: %+v", entry)
			}
		} else if entry.GroupKey != navigationKey("project", first.ProjectKey) || entry.GroupName != expectedName || entry.ConfigStatus != "missing" {
			t.Fatalf("inconsistent historical fallback group: %+v; expected name %q", entry, expectedName)
		}
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("normalizing presentation rewrote historical state")
	}
}

func TestHTTPGroupedNavigationRefreshPreservesPublication(t *testing.T) {
	s, request := newTimelineFixture(t)
	originalRoot := request.Source.ProjectRoot
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(originalRoot, alias); err != nil {
		t.Fatal(err)
	}
	request.Source.ProjectRoot = alias
	first := createTimeline(t, s, request)
	other := newRepo(t)
	survivor := newRepo(t)
	conflict := newRepo(t)
	s.remote = func(root string) (string, error) {
		if root == other || root == survivor || root == conflict {
			return "https://example.com/team/repo.git", nil
		}
		return "", nil
	}
	if _, err := RegisterRepository(filepath.Dir(s.Dir()), other); err != nil {
		t.Fatal(err)
	}
	if _, err := RegisterRepository(filepath.Dir(s.Dir()), survivor); err != nil {
		t.Fatal(err)
	}
	if _, err := RegisterRepository(filepath.Dir(s.Dir()), conflict); err != nil {
		t.Fatal(err)
	}
	request.Source.ProjectRoot, request.Source.NativeSessionID, request.OperationID = other, "other", "other"
	second := createTimeline(t, s, request)
	archived := true
	if _, err := s.TimelinePatch(second.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "archive", ExpectedVersion: second.Version, Actor: TimelineActor{Kind: "local_viewer"}, Reason: "Keep archived sessions in navigation", Archived: &archived}); err != nil {
		t.Fatal(err)
	}
	content := []byte("# Immutable historical document\n")
	document, err := s.TimelineAddDocument(first.ViewerSessionID, TimelineDocumentMetadata{Actor: timelineTestActor(s, first.ViewerSessionID), SchemaVersion: 3, OperationID: "document", MediaType: "text/markdown"}, content)
	if err != nil {
		t.Fatal(err)
	}
	first = domainCurrent(t, s, first.ViewerSessionID)
	// Both historical names are identical, but neither declares grouping yet.
	if first.ProjectName != second.ProjectName {
		t.Fatal("fixture must exercise equal legacy display names")
	}
	server, err := NewServer(ServerOptions{Store: s, Port: 4483, Token: strings.Repeat("z", 32)})
	if err != nil {
		t.Fatal(err)
	}
	get := func(query string) map[string]json.RawMessage {
		t.Helper()
		r := httptest.NewRequest("GET", "http://127.0.0.1:4483/api/index"+query, nil)
		r.Header.Set(tokenHeader, strings.Repeat("z", 32))
		r.Header.Set(TimelineRevisionHeader, TimelineRevisionValue)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("index: %d %s", w.Code, w.Body.String())
		}
		var out map[string]json.RawMessage
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	legacy := get("")
	before := stateManifest(t, filepath.Dir(s.Dir()))
	type entry struct {
		ID     string `json:"viewer_session_id"`
		Key    string `json:"group_key"`
		Name   string `json:"group_name"`
		Kind   string `json:"origin_kind"`
		Label  string `json:"origin_label"`
		Status string `json:"config_status"`
	}
	expectedEntries := 2
	decode := func() []entry {
		t.Helper()
		out := get("?navigation=grouped")
		raw, ok := out["navigation"]
		if !ok {
			t.Fatal("opt-in index has no navigation projection")
		}
		delete(out, "navigation")
		if !reflect.DeepEqual(out, get("")) {
			t.Fatal("grouped index changed legacy fields")
		}
		var projection struct {
			Version int     `json:"version"`
			Entries []entry `json:"entries"`
		}
		if err := json.Unmarshal(raw, &projection); err != nil {
			t.Fatal(err)
		}
		if projection.Version != 1 || len(projection.Entries) != expectedEntries {
			t.Fatalf("projection: %s", raw)
		}
		var sessions []TimelineAPIIndexEntry
		if err := json.Unmarshal(out["timeline_sessions"], &sessions); err != nil {
			t.Fatal(err)
		}
		for i, e := range projection.Entries {
			if e.ID != sessions[i].ViewerSessionID {
				t.Fatal("projection IDs/order differ from index")
			}
		}
		return projection.Entries
	}
	entries := decode()
	if entries[0].Key == entries[1].Key || entries[0].Status != "missing" {
		t.Fatal("legacy names implicitly grouped")
	}
	for _, root := range []string{first.Source.ProjectRoot, other} {
		writeGroupingConfig(t, root, `{"project_name":"  Example 雪  ","ignored":{"anything":true}}`)
	}
	entries = decode()
	hash := sha256.Sum256([]byte("Example 雪"))
	expected := "explicit:" + hex.EncodeToString(hash[:])
	for _, e := range entries {
		if e.Key != expected || e.Name != "Example 雪" || e.Status != "valid" || e.Kind != "repository" || e.Label != "repo" {
			t.Fatalf("entry: %+v", e)
		}
	}
	join := TimelineJoinRequestInput{SchemaVersion: 3, OperationID: "foreign-join", ExpectedSessionVersion: first.Version, Source: TimelineSource{ProjectRoot: other, Host: "codex", NativeSessionID: "foreign"}, Reason: "Same display group grants no membership"}
	_, err = s.TimelineV3Operation(first.ViewerSessionID, "join_request", "", &join)
	requireTimelineCode(t, err, "foreign_project")
	writeGroupingConfig(t, other, `{"project_name":"example 雪"}`)
	entries = decode()
	if entries[0].Key == entries[1].Key {
		t.Fatal("case folded explicit names")
	}
	writeGroupingConfig(t, other, `{"project_name":false}`)
	entries = decode()
	for _, e := range entries {
		if e.ID == second.ViewerSessionID && (e.Status != "invalid" || !strings.HasPrefix(e.Key, "project:")) {
			t.Fatalf("invalid fallback: %+v", e)
		}
	}
	if err := os.Remove(filepath.Join(other, ".engram", "config.json")); err != nil {
		t.Fatal(err)
	}
	entries = decode()
	for _, e := range entries {
		if e.ID == second.ViewerSessionID && e.Status != "missing" {
			t.Fatal("missing config did not refresh")
		}
	}
	if err := os.Remove(filepath.Join(other, ".git")); err != nil {
		t.Fatal(err)
	}
	entries = decode()
	for _, e := range entries {
		if e.ID == second.ViewerSessionID && e.Kind != "workspace" {
			t.Fatal("non-Git origin misclassified")
		}
	}
	if err := os.WriteFile(filepath.Join(other, ".git"), []byte("gitdir: /example/worktree\n"), 0600); err != nil {
		t.Fatal(err)
	}
	entries = decode()
	for _, e := range entries {
		if e.ID == second.ViewerSessionID && e.Kind != "repository" {
			t.Fatal("worktree origin misclassified")
		}
	}
	if err := os.Rename(other, other+"-unavailable"); err != nil {
		t.Fatal(err)
	}
	entries = decode()
	for _, e := range entries {
		if e.ID == second.ViewerSessionID && (e.Kind != "unavailable" || e.Status != "unavailable") {
			t.Fatalf("unavailable: %+v", e)
		}
	}
	unregistered := canon(t, t.TempDir())
	writeGroupingConfig(t, unregistered, `{"project_name":"Must not be read through a retargeted source"}`)
	if err := os.Remove(alias); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(unregistered, alias); err != nil {
		t.Fatal(err)
	}
	entries = decode()
	for _, e := range entries {
		if e.ID == first.ViewerSessionID && (e.Status != "unavailable" || !strings.HasPrefix(e.Key, "project:")) {
			t.Fatalf("retargeted source read outside registry: %+v", e)
		}
	}
	if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("navigation rewrote persisted publication state")
	}
	if !reflect.DeepEqual(get("")["timeline_sessions"], legacy["timeline_sessions"]) {
		t.Fatal("legacy identity changed")
	}
	// A retired worktree can resolve presentation through a live registered
	// source of the same persisted project identity, without relabeling history.
	writeGroupingConfig(t, survivor, `{"project_name":"Recovered project"}`)
	request.Source.ProjectRoot, request.Source.NativeSessionID, request.OperationID = survivor, "survivor", "survivor"
	request.WorkKey = timelineString("survivor")
	live := createTimeline(t, s, request)
	if live.ProjectKey != second.ProjectKey {
		t.Fatal("fixture must retain the worktree project identity")
	}
	expectedEntries = 3
	retained := stateManifest(t, filepath.Dir(s.Dir()))
	var liveEntry, retiredEntry entry
	for _, e := range decode() {
		if e.ID == live.ViewerSessionID {
			liveEntry = e
		}
		if e.ID == second.ViewerSessionID {
			retiredEntry = e
		}
	}
	if retiredEntry.Key != liveEntry.Key || retiredEntry.Name != "Recovered project" || retiredEntry.Status != "valid" || retiredEntry.Kind != "unavailable" {
		t.Fatalf("retired source lost presentation group: %+v; live %+v", retiredEntry, liveEntry)
	}
	if !reflect.DeepEqual(retained, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("resolving a retired source rewrote history")
	}
	// Conflicting declarations for the same identity must keep the diagnostic.
	writeGroupingConfig(t, conflict, `{"project_name":"Different declaration"}`)
	request.Source.ProjectRoot, request.Source.NativeSessionID, request.OperationID = conflict, "conflict", "conflict"
	request.WorkKey = timelineString("conflict")
	createTimeline(t, s, request)
	expectedEntries = 4
	retained = stateManifest(t, filepath.Dir(s.Dir()))
	for _, e := range decode() {
		if e.ID == second.ViewerSessionID && (e.Status != "unavailable" || !strings.HasPrefix(e.Key, "project:")) {
			t.Fatalf("ambiguous declarations silently selected: %+v", e)
		}
	}
	if !reflect.DeepEqual(retained, stateManifest(t, filepath.Dir(s.Dir()))) {
		t.Fatal("ambiguous lookup changed history")
	}
	read, err := s.TimelineDocument(first.ViewerSessionID, document.DocumentID)
	if err != nil || read.Content != string(content) {
		t.Fatalf("historical content changed: %v", err)
	}
}

func TestRegistryExplicitWorkspaceValidationAndIdempotence(t *testing.T) {
	for _, tc := range []struct {
		name, config string
		valid        bool
	}{
		{"unicode", `{"project_name":" 雪 "}`, true},
		{"limit", `{"project_name":"` + strings.Repeat("雪", 200) + `"}`, true},
		{"too_long", `{"project_name":"` + strings.Repeat("雪", 201) + `"}`, false},
		{"empty", `{"project_name":"  "}`, false},
		{"type", `{"project_name":2}`, false},
		{"control", `{"project_name":"bad\nname"}`, false},
		{"edge_control", `{"project_name":"\tvalid"}`, false},
		{"invalid_utf8", "{\"project_name\":\"\xff\"}", false},
		{"slash", `{"project_name":"bad/name"}`, false},
		{"backslash", `{"project_name":"bad\\name"}`, false},
		{"trailing", `{"project_name":"valid"} {}`, false},
		{"array", `[{"project_name":"valid"}]`, false},
		{"json", `{"project_name":`, false},
		{"oversize", `{"project_name":"valid","other":"` + strings.Repeat("x", 65536) + `"}`, false},
		{"byte_limit", `{"project_name":"valid"}` + strings.Repeat(" ", 65536-len(`{"project_name":"valid"}`)), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, state := canon(t, t.TempDir()), t.TempDir()
			writeGroupingConfig(t, root, tc.config)
			result, err := RegisterRepository(state, root)
			if !tc.valid {
				if err == nil || !strings.Contains(err.Error(), ".engram/config.json") {
					t.Fatalf("invalid registration: %+v %v", result, err)
				}
				if _, err := os.Stat(filepath.Join(state, "viewer")); !os.IsNotExist(err) {
					t.Fatal("invalid registration wrote state")
				}
				return
			}
			if err != nil || !result.Added || result.Root != root {
				t.Fatalf("explicit registration: %+v %v", result, err)
			}
			before := stateManifest(t, state)
			result, err = RegisterRepository(state, root)
			if err != nil || result.Added {
				t.Fatalf("repeat: %+v %v", result, err)
			}
			if !reflect.DeepEqual(before, stateManifest(t, state)) {
				t.Fatal("repeat changed registry/added_at")
			}
			alias := filepath.Join(t.TempDir(), "alias")
			if err := os.Symlink(root, alias); err != nil {
				t.Fatal(err)
			}
			result, err = RegisterRepository(state, alias)
			if err != nil || result.Root != root || result.Added {
				t.Fatalf("canonical registration: %+v %v", result, err)
			}
			sub := filepath.Join(root, "child")
			if err := os.Mkdir(sub, 0700); err != nil {
				t.Fatal(err)
			}
			if _, err := ResolveRoot(sub); err == nil {
				t.Fatal("inherited parent configuration")
			}
		})
	}
	for _, kind := range []string{"directory_link", "file_link"} {
		t.Run(kind, func(t *testing.T) {
			root, state, external := canon(t, t.TempDir()), t.TempDir(), t.TempDir()
			writeGroupingConfig(t, external, `{"project_name":"valid"}`)
			if kind == "directory_link" {
				if err := os.Symlink(filepath.Join(external, ".engram"), filepath.Join(root, ".engram")); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Mkdir(filepath.Join(root, ".engram"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(filepath.Join(external, ".engram", "config.json"), filepath.Join(root, ".engram", "config.json")); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := RegisterRepository(state, root); err == nil {
				t.Fatal("linked configuration accepted")
			}
		})
	}
	root := newRepo(t)
	sub := filepath.Join(root, "sub")
	writeGroupingConfig(t, sub, `{"project_name":"child"}`)
	if got, err := ResolveRoot(sub); err != nil || got != root {
		t.Fatalf("Git precedence: %s %v", got, err)
	}
}
