package viewer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTimelineMaintenanceEpisodesAndRestore(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	now = now.Add(3 * time.Hour)
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	current, _ := s.TimelineGet(v.ViewerSessionID)
	if current.Archived || current.Version != v.Version || current.LastActivityAt != v.LastActivityAt {
		t.Fatal("idle active work was mutated")
	}
	now = now.Add(time.Minute)
	closed := TimelineClosed
	_, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.ClosedAt == nil || v.AutoarchiveAt == nil || v.LastActivityAt != now.Format(time.RFC3339Nano) {
		t.Fatalf("missing closure/activity projections: %+v", v)
	}
	now = now.Add(7199 * time.Second)
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.Archived {
		t.Fatal("archived before deadline")
	}
	now = now.Add(time.Second)
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if !v.Archived || v.ArchiveOrigin == nil || *v.ArchiveOrigin != "automatic" {
		t.Fatal("deadline did not archive")
	}
	last := v.LastActivityAt
	no := false
	restore := TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "restore", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &no}
	result, err := s.TimelinePatch(v.ViewerSessionID, restore)
	if err != nil {
		t.Fatal(err)
	}
	if replay, e := s.TimelinePatch(v.ViewerSessionID, restore); e != nil || !timelineEqual(replay, result) {
		t.Fatalf("restore retry: %+v %v", replay, e)
	}
	now = now.Add(24 * time.Hour)
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.Archived || v.AutoarchiveAt != nil || v.LastActivityAt != last {
		t.Fatal("restore exemption/activity lost")
	}
	active := TimelineActive
	_, err = s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "reopen", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &active, Reason: "New work"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.ClosedAt != nil {
		t.Fatal("reopen kept old episode")
	}
	now = now.Add(time.Minute)
	_, err = s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close-again", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &closed, Reason: "New delivery"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.AutoarchiveAt == nil {
		t.Fatal("new close inherited exemption")
	}
}

func TestTimelineMaintenanceLegacyFallbackAndActivity(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	// A legacy explicit closed state with no retained closure proof must receive
	// one durable full window, without interpreting its general UpdatedAt.
	r := timelineClone(s.timeline[v.ViewerSessionID])
	r.Session.State = TimelineClosed
	r.Maintenance = nil
	if err := s.timelineSave(r); err != nil {
		t.Fatal(err)
	}
	now = now.Add(24 * time.Hour)
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if v.ClosedAt != nil || v.AutoarchiveAt == nil || v.Archived {
		t.Fatal("legacy fallback was not conservative")
	}
	deadline := *v.AutoarchiveAt
	now = now.Add(time.Hour)
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if *v.AutoarchiveAt != deadline {
		t.Fatal("restart moved fallback deadline")
	}
	now = now.Add(time.Hour)
	if err := s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	if !v.Archived {
		t.Fatal("persisted fallback did not expire")
	}
}

func TestTimelineMaintenanceAdministrativeCapacity(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	closed := TimelineClosed
	_, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
	if err != nil {
		t.Fatal(err)
	}
	r := timelineClone(s.timeline[v.ViewerSessionID])
	// Fill the retained native receipt ledger. Local reversal must neither consume
	// a native receipt nor append an otherwise exhausted public event stream.
	for len(r.Receipts) < TimelineMaxReceipts {
		receipt := r.Receipts[0]
		receipt.OperationID = fmt.Sprintf("prior-%d", len(r.Receipts))
		receipt.Result.OperationID = receipt.OperationID
		r.Receipts = append(r.Receipts, receipt)
	}
	if err = s.timelineSave(r); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(r.Receipts)
	events, _ := json.Marshal(r.Session.Events)
	now = now.Add(2 * time.Hour)
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(v.ViewerSessionID)
	no, yes := false, true
	patch := TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "restore-1", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &no}
	first, err := s.TimelinePatch(v.ViewerSessionID, patch)
	if err != nil {
		t.Fatal(err)
	}
	patch2 := patch
	patch2.OperationID = "archive-2"
	patch2.ExpectedVersion = first.SessionVersion
	patch2.Archived = &yes
	second, err := s.TimelinePatch(v.ViewerSessionID, patch2)
	if err != nil {
		t.Fatal(err)
	}
	patch3 := patch
	patch3.OperationID = "restore-3"
	patch3.ExpectedVersion = second.SessionVersion
	third, err := s.TimelinePatch(v.ViewerSessionID, patch3)
	if err != nil {
		t.Fatal(err)
	}
	if replay, e := s.TimelinePatch(v.ViewerSessionID, patch3); e != nil || !timelineEqual(replay, third) {
		t.Fatalf("exact retry %v", e)
	}
	_, err = s.TimelinePatch(v.ViewerSessionID, patch)
	requireTimelineCode(t, err, "version_conflict")
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	r = s.timeline[v.ViewerSessionID]
	after, _ := json.Marshal(r.Receipts)
	afterEvents, _ := json.Marshal(r.Session.Events)
	if !bytes.Equal(before, after) || !bytes.Equal(events, afterEvents) {
		t.Fatal("maintenance consumed or pruned public history")
	}
}

func TestTimelineMaintenanceNoopDoesNotAdvanceActivity(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	now = now.Add(time.Hour)
	_, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "unchanged", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, Summary: &v.Summary})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := s.TimelineGet(v.ViewerSessionID)
	if got.LastActivityAt != v.LastActivityAt {
		t.Fatal("unchanged native mutation advanced activity")
	}
	legacy := timelineClone(s.timeline[v.ViewerSessionID])
	legacy.Maintenance = nil
	if err = s.timelineSave(legacy); err != nil {
		t.Fatal(err)
	}
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	got, _ = s.TimelineGet(v.ViewerSessionID)
	if got.LastActivityAt != v.LastActivityAt {
		t.Fatalf("legacy unchanged native mutation advanced activity: got %s, want %s", got.LastActivityAt, v.LastActivityAt)
	}
}

func TestTimelineMaintenancePinReopenAndManualOrigin(t *testing.T) {
	for _, manual := range []bool{false, true} {
		t.Run(fmt.Sprint(manual), func(t *testing.T) {
			s, req := newTimelineFixture(t)
			now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
			s.now = func() time.Time { return now }
			v := createTimeline(t, s, req)
			id := v.ViewerSessionID
			closed := TimelineClosed
			_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
			if err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			yes, no := true, false
			_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "pin", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Pinned: &yes})
			if err != nil {
				t.Fatal(err)
			}
			now = now.Add(3 * time.Hour)
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if v.Archived || v.AutoarchiveAt != nil {
				t.Fatal("pinned work became eligible")
			}
			_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "unpin", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Pinned: &no})
			if err != nil {
				t.Fatal(err)
			}
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if !v.Archived {
				t.Fatal("unpin ignored original expired deadline")
			}
			if manual {
				_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "manual-archive", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &yes})
				if err != nil {
					t.Fatal(err)
				}
				v, _ = s.TimelineGet(id)
			}
			active := TimelineActive
			_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "reopen", Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &active, Reason: "Resume explicit work"})
			if err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if v.Archived != manual {
				t.Fatal("reopen lost archive origin policy")
			}
		})
	}
}

func TestTimelineMaintenanceFailureAndVersionExhaustion(t *testing.T) {
	for _, phase := range []string{"before_rename", "after_rename", "version"} {
		t.Run(phase, func(t *testing.T) {
			s, req := newTimelineFixture(t)
			now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
			s.now = func() time.Time { return now }
			v := createTimeline(t, s, req)
			id := v.ViewerSessionID
			closed := TimelineClosed
			_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
			if err != nil {
				t.Fatal(err)
			}
			if phase == "version" {
				r := timelineClone(s.timeline[id])
				r.Session.Version = TimelineMaxNumber - 1
				if err = s.timelineSave(r); err != nil {
					t.Fatal(err)
				}
			} else {
				s.timelineFault = func(at, path string) error {
					if at == phase {
						return errors.New("injected physical failure")
					}
					return nil
				}
			}
			now = now.Add(2 * time.Hour)
			if err = s.Startup(); err == nil {
				t.Fatal("failure reported successful maintenance")
			}
			s.timelineFault = nil
			if phase != "after_rename" {
				v, _ = s.TimelineGet(id)
				if v.Archived {
					t.Fatal("unsafe failed archive changed visibility")
				}
			}
			if phase == "version" {
				return
			}
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if !v.Archived {
				t.Fatal("retry/reconciliation did not recover")
			}
			version := v.Version
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if v.Version != version {
				t.Fatal("reconciliation archived twice")
			}
		})
	}
}

func TestTimelineMaintenanceLegacyHistory(t *testing.T) {
	for _, scenario := range []struct{ restoreCurrent, unknownClosure bool }{{true, false}, {false, false}, {true, true}, {false, true}} {
		restoreCurrent := scenario.restoreCurrent
		t.Run(fmt.Sprintf("restore=%v/unknown=%v", restoreCurrent, scenario.unknownClosure), func(t *testing.T) {
			s, req := newTimelineFixture(t)
			now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
			s.now = func() time.Time { return now }
			v := createTimeline(t, s, req)
			id := v.ViewerSessionID
			closed, active := TimelineClosed, TimelineActive
			closeWork := func(op string, state TimelineState) {
				t.Helper()
				v, _ = s.TimelineGet(id)
				_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: op, Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &state, Reason: "Explicit delivery transition"})
				if err != nil {
					t.Fatal(err)
				}
			}
			closeWork("close", closed)
			now = now.Add(time.Minute)
			r := timelineClone(s.timeline[id])
			if err := s.timelineAuditEvent(r.Session, "local_action", TimelineActor{Kind: "local_viewer"}, "", "", "", TimelineAudit{Operation: "session_archive", Reason: "Historical manual restore"}, true); err != nil {
				t.Fatal(err)
			}
			r.Session.UpdatedAt = s.timelineTime()
			if err := s.timelineSave(r); err != nil {
				t.Fatal(err)
			}
			if !restoreCurrent {
				closeWork("reopen", active)
				now = now.Add(time.Minute)
				closeWork("new-close", closed)
			}
			r = timelineClone(s.timeline[id])
			r.Maintenance = nil
			// Unknown-time legacy work can retain a manual restore even when
			// its original close event is absent. Retain reopen transitions.
			if scenario.unknownClosure {
				events := []TimelineEvent{}
				for _, event := range r.Session.Events {
					if event.Audit != nil && event.Audit.NextState == string(TimelineClosed) {
						continue
					}
					event.Sequence = int64(len(events) + 1)
					events = append(events, event)
				}
				r.Session.Events = events
			}
			expectedActivity := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
			if !restoreCurrent {
				expectedActivity = now.Format(time.RFC3339Nano)
			}
			// Historic private server receipts can truthfully follow general UpdatedAt.
			audit := r.Audit[0]
			audit.Operation = "session_patch"
			audit.OperationID = "legacy-summary"
			audit.AuditID = "audit-legacy-summary"
			audit.ReceivedAt = now.Add(time.Microsecond).Format(time.RFC3339Nano)
			receipt := r.Receipts[0]
			receipt.OperationID = audit.OperationID
			receipt.Result.OperationID = audit.OperationID
			receipt.Result.Kind = "session_patch"
			receipt.Result.Created = nil
			r.Audit = append(r.Audit, audit)
			r.Receipts = append(r.Receipts, receipt)
			// Write a legacy encoding, which lacked derived-field validation.
			raw, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(s.timelineRecordPath(id), raw, 0600); err != nil {
				t.Fatal(err)
			}
			now = now.Add(3 * time.Hour)
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if (!scenario.unknownClosure || restoreCurrent) && v.Archived == restoreCurrent {
				t.Fatal("restore from wrong closure episode applied")
			}
			if !scenario.unknownClosure && v.LastActivityAt != expectedActivity {
				t.Fatalf("ambiguous summary audit replaced proved activity: got %s want %s", v.LastActivityAt, expectedActivity)
			}
			// Advance past the persisted fallback window and reload again.
			now = now.Add(2 * time.Hour)
			if err = s.Startup(); err != nil {
				t.Fatal(err)
			}
			v, _ = s.TimelineGet(id)
			if v.Archived == restoreCurrent {
				t.Fatal("fallback window ignored current manual restore or later reopen")
			}

		})
	}
}

func TestTimelineMaintenanceRunTickAndCancellation(t *testing.T) {
	s, req := newTimelineFixture(t)
	initial := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	var clock atomic.Int64
	clock.Store(initial.UnixNano())
	s.now = func() time.Time { return time.Unix(0, clock.Load()).UTC() }
	v := createTimeline(t, s, req)
	closed := TimelineClosed
	_, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx, RunOptions{PollInterval: time.Millisecond}) }()
	clock.Store(initial.Add(2 * time.Hour).UnixNano())
	deadline := time.NewTimer(3 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(time.Millisecond)
	defer poll.Stop()
	for {
		select {
		case <-deadline.C:
			t.Fatal("runtime sweep did not archive")
		case <-poll.C:
			got, e := s.TimelineGet(v.ViewerSessionID)
			if e != nil {
				t.Fatal(e)
			}
			if got.Archived {
				cancel()
				select {
				case e = <-done:
					if e != nil {
						t.Fatal(e)
					}
				case <-time.After(time.Second):
					t.Fatal("Run ignored cancellation")
				}
				return
			}
		}
	}
}

func TestTimelineMaintenanceConcurrentChoices(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	id := v.ViewerSessionID
	closed, active := TimelineClosed, TimelineActive
	_, err := s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(id)
	now = now.Add(2 * time.Hour)
	yes, no := true, false
	requests := []TimelineSessionPatchRequest{
		{SchemaVersion: 3, OperationID: "pin", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Pinned: &yes},
		{SchemaVersion: 3, OperationID: "restore", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &no},
		{SchemaVersion: 3, OperationID: "reopen", Actor: timelineTestActor(s, id), ExpectedVersion: v.Version, State: &active, Reason: "Resume work"},
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, req := range requests {
		wg.Add(1)
		go func(r TimelineSessionPatchRequest) {
			defer wg.Done()
			<-start
			_, e := s.TimelinePatch(id, r)
			if e != nil {
				var ce *TimelineContractError
				if !errors.As(e, &ce) || ce.Code != "version_conflict" {
					t.Errorf("concurrent patch: %v", e)
				}
			}
		}(req)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		s.mu.Lock()
		e := s.timelineMaintainLocked()
		s.mu.Unlock()
		if e != nil {
			t.Errorf("sweep: %v", e)
		}
	}()
	close(start)
	wg.Wait()
	v, err = s.TimelineGet(id)
	if err != nil {
		t.Fatal(err)
	}
	if v.Archived && (v.Pinned || v.State != TimelineClosed || s.timeline[id].Maintenance.RestoreExempt) {
		t.Fatal("sweep overwrote concurrent explicit choice")
	}
	if err = ValidateTimelineSession(v); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineMaintenanceRejectsStoredDerivedFields(t *testing.T) {
	s, req := newTimelineFixture(t)
	v := createTimeline(t, s, req)
	raw, err := os.ReadFile(s.timelineRecordPath(v.ViewerSessionID))
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	json.Unmarshal(raw, &fields)
	var session map[string]json.RawMessage
	json.Unmarshal(fields["session"], &session)
	session["last_activity_at"] = json.RawMessage(`"not-a-timestamp"`)
	fields["session"], _ = json.Marshal(session)
	raw, _ = json.Marshal(fields)
	if err = os.WriteFile(s.timelineRecordPath(v.ViewerSessionID), raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(Options{StateDir: filepath.Dir(s.Dir())}); err == nil {
		t.Fatal("corrupt stored derived field silently replaced")
	}
}

func TestTimelineMaintenanceExactLegacyLimits(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	v := createTimeline(t, s, req)
	id := v.ViewerSessionID
	document, err := s.TimelineAddDocument(id, TimelineDocumentMetadata{Actor: timelineTestActor(s, id), SchemaVersion: 3, OperationID: "retained-document", MediaType: "text/markdown"}, []byte("# Immutable retained evidence\n"))
	if err != nil {
		t.Fatal(err)
	}
	var imageBytes bytes.Buffer
	if err = png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	artifact, err := s.TimelineAddArtifact(id, TimelineArtifactMetadata{Actor: timelineTestActor(s, id), SchemaVersion: 3, OperationID: "retained-image", CapturedAt: now.Format(time.RFC3339Nano), CandidateID: domainCandidate(t, s, id, "retained-sha"), Viewport: "desktop", Theme: "dark", MediaType: "image/png"}, imageBytes.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	docPath := filepath.Join(s.timelineBlobDir(id), "documents", document.DocumentID+".md")
	imagePath := filepath.Join(s.timelineBlobDir(id), "artifacts", artifact.ArtifactID+".png")
	docBytes, _ := os.ReadFile(docPath)
	imageRaw, _ := os.ReadFile(imagePath)
	docHash, imageHash := sha256.Sum256(docBytes), sha256.Sum256(imageRaw)
	r := timelineClone(s.timeline[id])
	r.Maintenance = nil
	r.Session.State = TimelineClosed
	r.Session.Version = 9
	actor := timelineTestActor(s, id)
	for len(r.Session.Events) < TimelineMaxEvents {
		n := len(r.Session.Events) + 1
		r.Session.Events = append(r.Session.Events, TimelineEvent{Actor: actor, EventID: fmt.Sprintf("legacy-event-%d", n), Sequence: int64(n), Version: 1, Kind: "request", Title: "Retained work", Summary: strings.Repeat("s", 4000), CreatedAt: v.CreatedAt, UpdatedAt: v.CreatedAt, Origin: "agent_reported", TicketRefs: []TimelineTicketRef{}})
	}
	body, _ := timelineSessionContentBytes(r.Session)
	excess := len(body) - TimelineMaxBody
	for i := len(r.Session.Events) - 1; i >= 0 && excess > 0; i-- {
		n := min(excess, len(r.Session.Events[i].Summary)-1)
		r.Session.Events[i].Summary = r.Session.Events[i].Summary[:len(r.Session.Events[i].Summary)-n]
		excess -= n
	}
	body, _ = timelineSessionContentBytes(r.Session)
	if len(body) != TimelineMaxBody {
		t.Fatalf("fixture body %d", len(body))
	}
	for len(r.Receipts) < TimelineMaxReceipts {
		n := len(r.Receipts)
		op := fmt.Sprintf("legacy-operation-%d", n)
		receipt := r.Receipts[0]
		receipt.OperationID = op
		receipt.Result.OperationID = op
		r.Receipts = append(r.Receipts, receipt)
		audit := r.Audit[0]
		audit.OperationID = op
		audit.AuditID = fmt.Sprintf("audit-legacy-%d", n)
		audit.Reason = strings.Repeat("<", 500)
		r.Audit = append(r.Audit, audit)
	}
	raw, _ := timelineRecordContentBytes(r)
	excess = len(raw) - timelineMaxRecordBytes
	for i := len(r.Audit) - 1; i >= 0 && excess > 0; i-- {
		reason := r.Audit[i].Reason
		// A '<' encodes as six bytes. Replacing it with ASCII frees five;
		// deleting a resulting ASCII character frees the remaining single byte.
		n := min(excess/5, len(reason))
		reason = strings.Repeat("s", n) + reason[n:]
		excess -= n * 5
		if excess > 0 && excess < 5 && n >= excess {
			reason = reason[excess:]
			excess = 0
		}
		r.Audit[i].Reason = reason
	}
	raw, _ = timelineRecordContentBytes(r)
	if len(raw) != timelineMaxRecordBytes {
		t.Fatalf("fixture record %d excess %d", len(raw), excess)
	}
	if err := s.timelineSave(r); err != nil {
		t.Fatal(err)
	}
	publicHistory, _ := json.Marshal(r.Session.Events)
	receipts, _ := json.Marshal(r.Receipts)
	now = now.Add(time.Nanosecond)
	if err := s.Startup(); err != nil {
		t.Fatalf("initialize exact legacy limits: %v", err)
	}
	now = now.Add(2 * time.Hour)
	if err := s.Startup(); err != nil {
		t.Fatalf("archive exact legacy limits: %v", err)
	}
	v, _ = s.TimelineGet(id)
	if !v.Archived {
		t.Fatal("full legacy work remained visible")
	}
	no := false
	restore := TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "restore-full", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &no}
	first, err := s.TimelinePatch(id, restore)
	if err != nil {
		t.Fatalf("restore full: %v", err)
	}
	if replay, e := s.TimelinePatch(id, restore); e != nil || !timelineEqual(replay, first) {
		t.Fatalf("full exact retry: %v", e)
	}
	stale := restore
	stale.OperationID = "stale-full"
	_, err = s.TimelinePatch(id, stale)
	requireTimelineCode(t, err, "version_conflict")
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
	v, _ = s.TimelineGet(id)
	if v.Archived || v.AutoarchiveAt != nil {
		t.Fatal("full restore exemption lost")
	}
	after, _ := json.Marshal(s.timeline[id].Session.Events)
	afterReceipts, _ := json.Marshal(s.timeline[id].Receipts)
	if !bytes.Equal(publicHistory, after) || !bytes.Equal(receipts, afterReceipts) {
		t.Fatal("full history changed")
	}
	docBytes, _ = os.ReadFile(docPath)
	imageRaw, _ = os.ReadFile(imagePath)
	if sha256.Sum256(docBytes) != docHash || sha256.Sum256(imageRaw) != imageHash {
		t.Fatal("retained blob hashes changed")
	}
	t.Logf("exact legacy limits preserved: body=%d record=%d events=%d receipts=%d; document SHA256=%x image SHA256=%x", TimelineMaxBody, timelineMaxRecordBytes, len(r.Session.Events), len(r.Receipts), docHash, imageHash)
	// Rename retains its normal public event admission limits.
	title := v.Title + " with more content"
	_, err = s.TimelinePatch(id, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "cannot-borrow-headroom", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Title: &title})
	requireTimelineCode(t, err, "limit")
}

func TestTimelineMaintenancePreservesExplicitDeletion(t *testing.T) {
	s, req := newTimelineFixture(t)
	v := createTimeline(t, s, req)
	yes := true
	result, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "manual-archive", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: v.Version, Archived: &yes})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.TimelineDelete(v.ViewerSessionID, TimelineMutationRequest{SchemaVersion: 3, OperationID: "explicit-delete", Actor: TimelineActor{Kind: "local_viewer"}, ExpectedVersion: result.SessionVersion})
	if err != nil {
		t.Fatal("maintenance blocked explicit deletion", err)
	}
	if err = s.Startup(); err != nil {
		t.Fatal(err)
	}
}

func TestTimelineMaintenanceRunReportsFailureAndContinues(t *testing.T) {
	s, req := newTimelineFixture(t)
	now := time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	ids := []string{}
	for i := range 2 {
		request := req
		request.OperationID = fmt.Sprintf("create-%d", i)
		request.WorkKey = timelineString(fmt.Sprintf("work-%d", i))
		v := createTimeline(t, s, request)
		closed := TimelineClosed
		_, err := s.TimelinePatch(v.ViewerSessionID, TimelineSessionPatchRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, v.ViewerSessionID), ExpectedVersion: v.Version, State: &closed, Reason: "Completed delivery"})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.ViewerSessionID)
	}
	now = now.Add(2 * time.Hour)
	s.timelineFault = func(at, path string) error {
		if at == "before_rename" && strings.HasSuffix(path, ids[0]+".json") {
			return errors.New("injected archive failure")
		}
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reported := false
	err := s.Run(ctx, RunOptions{OnError: func(err error) { reported = err != nil; cancel() }})
	if err != nil || !reported {
		t.Fatalf("maintenance callback: reported=%v err=%v", reported, err)
	}
	first, _ := s.TimelineGet(ids[0])
	second, _ := s.TimelineGet(ids[1])
	if first.Archived || !second.Archived {
		t.Fatal("failed work hid or prevented safe unrelated maintenance")
	}
}
