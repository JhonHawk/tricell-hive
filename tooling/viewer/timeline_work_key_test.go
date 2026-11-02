package viewer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func workKeyRequest(base TimelineCreateSessionRequest, operation string, key *string) TimelineCreateSessionRequest {
	base.OperationID = operation
	base.WorkKey = key
	return base
}

func TestTimelineWorkKeysSplitAndResumeWorkInOneNativeSession(t *testing.T) {
	s, base := newTimelineFixture(t)
	first, err := s.TimelineCreate(workKeyRequest(base, "op-a", timelineString("COM-1")))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.TimelineCreate(workKeyRequest(base, "op-b", timelineString("COM-2")))
	if err != nil {
		t.Fatal(err)
	}
	if first.Session.ViewerSessionID == second.Session.ViewerSessionID {
		t.Fatal("two work keys shared one viewer session")
	}
	resumed, err := s.TimelineCreate(workKeyRequest(base, "op-c", timelineString("COM-1")))
	if err != nil {
		t.Fatal(err)
	}
	if resumed.Session.ViewerSessionID != first.Session.ViewerSessionID || resumed.Created == nil || *resumed.Created {
		t.Fatalf("same work key did not resume: %+v", resumed.Created)
	}
	if first.Session.WorkKey == nil || *first.Session.WorkKey != "COM-1" || resumed.Session.WorkKey == nil || *resumed.Session.WorkKey != "COM-1" {
		t.Fatal("response did not echo the requested work key")
	}
	if _, err = s.TimelineGet(first.Session.ViewerSessionID); err != nil {
		t.Fatal(err)
	}
	other := base
	other.Source.NativeSessionID = "native-2"
	elsewhere, err := s.TimelineCreate(workKeyRequest(other, "op-d", timelineString("COM-1")))
	if err != nil || elsewhere.Session.ViewerSessionID == first.Session.ViewerSessionID {
		t.Fatal("work key crossed native sessions", err)
	}
}

func TestTimelineWithoutWorkKeyCreatesOneWorkPerOperation(t *testing.T) {
	s, base := newTimelineFixture(t)
	first, err := s.TimelineCreate(workKeyRequest(base, "op-one", nil))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.TimelineCreate(workKeyRequest(base, "op-two", nil))
	if err != nil {
		t.Fatal(err)
	}
	if first.Session.ViewerSessionID == second.Session.ViewerSessionID {
		t.Fatal("two operation IDs without a key shared one work")
	}
	if first.Session.WorkKey != nil || second.Session.WorkKey != nil {
		t.Fatal("effective operation key leaked into the response")
	}
	raw, err := json.Marshal(first.Session)
	if err != nil || !strings.Contains(string(raw), `"work_key":null`) {
		t.Fatalf("session must serialize work_key as null: %s", raw)
	}
	replay, err := s.TimelineCreate(workKeyRequest(base, "op-one", nil))
	if err != nil || replay.Session.ViewerSessionID != first.Session.ViewerSessionID {
		t.Fatal("identical replay did not return the same work", err)
	}
	if got := len(s.TimelineIndex()); got != 2 {
		t.Fatalf("index holds %d works, want 2", got)
	}
}

func TestTimelineWorkKeysSurviveRestart(t *testing.T) {
	s, base := newTimelineFixture(t)
	keyed := createTimeline(t, s, workKeyRequest(base, "op-k", timelineString("COM-9")))
	anonymous := createTimeline(t, s, workKeyRequest(base, "op-x", nil))
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range []TimelineSession{keyed, anonymous} {
		got, err := restarted.TimelineGet(v.ViewerSessionID)
		if err != nil || got.ViewerSessionID != v.ViewerSessionID {
			t.Fatal("restart lost a work", err)
		}
	}
	again, err := restarted.TimelineCreate(workKeyRequest(base, "op-later", timelineString("COM-9")))
	if err != nil || again.Session.ViewerSessionID != keyed.ViewerSessionID || *again.Created {
		t.Fatal("restart changed the work key derivation", err)
	}
}

// legacyRecord rewrites a stored record into the shape written before work
// keys existed: the pre-change ID, and no work_key anywhere.
func legacyRecord(t *testing.T, s *Store, v TimelineSession, oldID string) {
	t.Helper()
	path := s.timelineRecordPath(v.ViewerSessionID)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var tree any
	if err = json.Unmarshal(raw, &tree); err != nil {
		t.Fatal(err)
	}
	var walk func(any) any
	walk = func(node any) any {
		switch n := node.(type) {
		case map[string]any:
			delete(n, "work_key")
			for k, child := range n {
				n[k] = walk(child)
			}
		case []any:
			for i, child := range n {
				n[i] = walk(child)
			}
		case string:
			return strings.ReplaceAll(n, v.ViewerSessionID, oldID)
		}
		return node
	}
	out, err := json.Marshal(walk(tree))
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(s.timelineRecordPath(oldID), out, 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(s.timelineBlobDir(v.ViewerSessionID), s.timelineBlobDir(oldID)); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestTimelineRecordStoredWithTheOldDerivationLoads(t *testing.T) {
	s, base := newTimelineFixture(t)
	v := createTimeline(t, s, workKeyRequest(base, "op-old", nil))
	oldID := preWorkKeyID(v.ProjectKey, v.Source)
	if oldID == v.ViewerSessionID {
		t.Fatal("test needs a derivation that differs from the old one")
	}
	legacyRecord(t, s, v, oldID)
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	got, err := restarted.TimelineGet(oldID)
	if err != nil || got.ViewerSessionID != oldID || got.WorkKey != nil {
		t.Fatal("record with the old derivation did not load unchanged", err)
	}
}

// preWorkKeyID writes out the derivation used before work keys existed, apart
// from timelineSessionKey, so a change to that function cannot pass unnoticed.
func preWorkKeyID(project string, source TimelineSource) string {
	raw, _ := json.Marshal([]any{"timeline-api-v3", project, source.Host, source.NativeSessionID})
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// Sessions stored before work keys must keep their IDs: this pins the value.
func TestTimelineEmptyWorkKeyKeepsThePinnedDerivation(t *testing.T) {
	source := TimelineSource{Host: "codex", NativeSessionID: "native-session-alpha"}
	const want = "e83c9a92a9e50f9d8e8141d9b45b7302aeb5ca69ef4b73316e8ab534f604f3c2"
	if got := timelineSessionKey("example.com/fixture/project", source, ""); got != want {
		t.Fatalf("old derivation changed: got %s, want %s", got, want)
	}
	if got := preWorkKeyID("example.com/fixture/project", source); got != want {
		t.Fatalf("test helper drifted: got %s, want %s", got, want)
	}
}

func TestTimelineCreateRejectsInvalidWorkKeys(t *testing.T) {
	s, base := newTimelineFixture(t)
	for _, bad := range []string{"", " COM-1", "COM 1", "COM\u0007", strings.Repeat("k", 129)} {
		_, err := s.TimelineCreate(workKeyRequest(base, "op-bad", timelineString(bad)))
		requireTimelineCode(t, err, "invalid_text")
	}
}
