package viewer

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"
)

func domainCandidate(t *testing.T, s *Store, id, commit string) string {
	t.Helper()
	v := domainCurrent(t, s, id)
	r := TimelineCandidateCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "candidate-"+commit, v.ForegroundProducerID), TimelineCandidate: TimelineCandidate{Branch: "feature", Commit: commit, Base: "development"}}
	out, err := s.TimelineV3Operation(id, "candidate_create", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	return out.CandidateID
}
func domainDelivery(t *testing.T, s *Store, id string) string {
	t.Helper()
	v := domainCurrent(t, s, id)
	r := TimelineDeliveryCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "delivery-register", v.ForegroundProducerID), RepositoryURL: "https://example.com/repo", ExternalID: "1", TimelineDeliverySnapshot: TimelineDeliverySnapshot{Kind: "pull_request", Label: "PR1", URL: "https://example.com/pull/1", Status: "open"}}
	out, err := s.TimelineV3Operation(id, "delivery_create", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	return out.DeliveryID
}
func TestTimelineCatalogRuntimeDedupCASReplayAndRestart(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	now, _ := time.Parse(time.RFC3339Nano, v.CreatedAt)
	s.now = func() time.Time { return now }
	id := v.ViewerSessionID
	r := TimelineTicketCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "ticket-create", v.ForegroundProducerID), TimelineTicketInput: TimelineTicketInput{ExternalScope: "HTTPS://EXAMPLE.com:443/tracker/", ExternalID: "1", Key: "EX-1", URL: "https://example.com/issue/1"}}
	first, err := s.TimelineV3Operation(id, "ticket_create", "", &r)
	if err != nil {
		t.Fatal(err)
	}
	activity := domainCurrent(t, s, id).LastActivityAt
	for i := 0; i < 4; i++ {
		now = now.Add(time.Hour)
		r.TimelineNativeOperation = domainBase(t, s, id, fmt.Sprintf("ticket-repeat-%d", i), v.ForegroundProducerID)
		out, err := s.TimelineV3Operation(id, "ticket_create", "", &r)
		if err != nil || out.TicketID != first.TicketID || *out.Created {
			t.Fatalf("dedup: %+v %v", out, err)
		}
		if got := domainCurrent(t, s, id); got.LastActivityAt != activity {
			t.Fatal("deduplicated catalog creation advanced activity")
		}
	}
	before := domainCurrent(t, s, id)
	r.TimelineNativeOperation = domainBase(t, s, id, "ticket-conflict", v.ForegroundProducerID)
	r.Key = "different"
	_, err = s.TimelineV3Operation(id, "ticket_create", "", &r)
	requireTimelineCode(t, err, "entity_exists")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("failed registration changed work")
	}
	p := TimelineTicketPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "patch", v.ForegroundProducerID), Key: timelineStringPointer("EX-2")}
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			q := p
			q.OperationID = fmt.Sprintf("race-%d", i)
			_, err := s.TimelineV3Operation(id, "ticket_patch", first.TicketID, &q)
			if err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	if wins != 1 {
		t.Fatal("CAS winners", wins)
	}
	current := domainCurrent(t, s, id)
	if current.LastActivityAt != now.Format(time.RFC3339Nano) {
		t.Fatal("real catalog patch did not advance activity")
	}
	now = now.Add(time.Hour)
	same := TimelineTicketPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "same-ticket-key", v.ForegroundProducerID), Key: timelineStringPointer("EX-2")}
	if _, err = s.TimelineV3Operation(id, "ticket_patch", first.TicketID, &same); err != nil {
		t.Fatal(err)
	}
	if got := domainCurrent(t, s, id); got.LastActivityAt != current.LastActivityAt {
		t.Fatal("unchanged catalog patch advanced activity")
	}
	c := domainCandidate(t, s, id, "abc")
	d := domainDelivery(t, s, id)
	if c == "" || d == "" {
		t.Fatal("missing IDs")
	}
	reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	// Replay retains its original created flag and version after later writes.
	r.Key = "EX-1"
	r.TimelineNativeOperation = TimelineNativeOperation{SchemaVersion: 3, OperationID: "ticket-create", Actor: TimelineActor{Kind: "native_producer", ProducerID: v.ForegroundProducerID}, ExpectedSessionVersion: v.Version, Reason: "Explicit ticket-create"}
	replay, err := reopened.TimelineV3Operation(id, "ticket_create", "", &r)
	if err != nil || !reflect.DeepEqual(first, replay) {
		t.Fatalf("replay %+v %v", replay, err)
	}
}

func TestTimelineCatalogLinkedInvalidationIsAtomicAndHistorical(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	candidate := domainCandidate(t, s, id, "abc")
	delivery := domainDelivery(t, s, id)
	ticketReq := TimelineTicketCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "ticket", v.ForegroundProducerID), TimelineTicketInput: TimelineTicketInput{ExternalScope: "https://example.com/tracker", ExternalID: "one", Key: "EX-1"}}
	ticket, err := s.TimelineV3Operation(id, "ticket_create", "", &ticketReq)
	if err != nil {
		t.Fatal(err)
	}
	stage := domainStage(t, s, id, "delivery")
	task := domainTask(t, s, id)
	domainBind(t, s, id, task.TaskID, task.Segments[1].SegmentID, stage, "bind")
	_, err = s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "linked-delivery", TargetStageEventID: stage, Actor: timelineTestActor(s, id), Kind: "delivery", Title: "PR opened", Summary: "Published open", TaskID: task.TaskID, SegmentID: task.Segments[1].SegmentID, TicketIDs: []string{ticket.TicketID}, DeliveryIDs: []string{delivery}})
	if err != nil {
		t.Fatal(err)
	}
	wait := TimelineWaitCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "wait", v.ForegroundProducerID), TargetStageEventID: stage, TaskID: task.TaskID, SegmentID: task.Segments[1].SegmentID, CandidateID: candidate, PullRequestDeliveryID: delivery, Policy: TimelineCheckPolicy{PolicyID: "policy", Confirmed: false, Authority: "Synthetic", SourceURL: "https://example.com/policy", ObservedAt: "2026-01-01T00:00:00Z", RequiredChecks: []TimelineRequiredCheck{}}}
	_, err = s.TimelineV3Operation(id, "wait_create", "", &wait)
	if err != nil {
		t.Fatal(err)
	}
	before := domainCurrent(t, s, id)
	patch := TimelineDeliveryPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "patch-delivery", v.ForegroundProducerID), URL: timelineStringPointer("https://example.com/pull/corrected"), Status: timelineStringPointer("merged")}
	result, err := s.TimelineV3Operation(id, "delivery_patch", delivery, &patch)
	if err != nil {
		t.Fatal(err)
	}
	after := domainCurrent(t, s, id)
	if after.Tasks[0].TaskVersion != before.Tasks[0].TaskVersion+1 || after.Waits[0].WaitVersion != before.Waits[0].WaitVersion+1 {
		t.Fatal("multiple paths invalidated more than once")
	}
	if !reflect.DeepEqual(before.Events, after.Events) || before.Waits[0].PullRequestRef != after.Waits[0].PullRequestRef || before.Waits[0].CandidateID != after.Waits[0].CandidateID || !reflect.DeepEqual(before.Candidates, after.Candidates) {
		t.Fatal("immutable publication context changed")
	}
	replay, err := s.TimelineV3Operation(id, "delivery_patch", delivery, &patch)
	if err != nil || !reflect.DeepEqual(result, replay) {
		t.Fatal("retry changed receipt", err)
	}
	patch.TimelineNativeOperation = domainBase(t, s, id, "noop-delivery", v.ForegroundProducerID)
	_, err = s.TimelineV3Operation(id, "delivery_patch", delivery, &patch)
	if err != nil {
		t.Fatal(err)
	}
	noop := domainCurrent(t, s, id)
	if noop.Tasks[0].TaskVersion != after.Tasks[0].TaskVersion || noop.Waits[0].WaitVersion != after.Waits[0].WaitVersion {
		t.Fatal("noop invalidated")
	}
	tp := TimelineTicketPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "patch-ticket", v.ForegroundProducerID), Key: timelineStringPointer("EX-2"), URL: timelineStringPointer("https://example.com/issue/one")}
	_, err = s.TimelineV3Operation(id, "ticket_patch", ticket.TicketID, &tp)
	if err != nil {
		t.Fatal(err)
	}
	final := domainCurrent(t, s, id)
	if final.Tasks[0].TaskVersion != noop.Tasks[0].TaskVersion+1 || final.Waits[0].WaitVersion != noop.Waits[0].WaitVersion || !reflect.DeepEqual(before.Events, final.Events) {
		t.Fatal("ticket invalidation or historical mutation")
	}
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(final, domainCurrent(t, restarted, id)) {
		t.Fatal("restart changed canonical data")
	}
}

func TestTimelineCatalogRuntimeAuthorityConflictAndBootstrap(t *testing.T) {
	s, req := newTimelineFixture(t)
	req.Tickets = []TimelineTicketInput{{ExternalScope: "https://example.com/tracker", ExternalID: "one", Key: "SAME"}, {ExternalScope: "https://other.example.com/tracker", ExternalID: "one", Key: "SAME"}}
	created, err := s.TimelineCreate(req)
	if err != nil {
		t.Fatal(err)
	}
	id := created.Session.ViewerSessionID
	restarted, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
	if err != nil {
		t.Fatal(err)
	}
	retry, err := restarted.TimelineCreate(req)
	if err != nil || !reflect.DeepEqual(created.Session, retry.Session) {
		t.Fatal("bootstrap replay lost IDs", err)
	}
	s = restarted
	v := domainCurrent(t, s, id)
	collaborator := domainAccept(t, s, id, domainJoin(t, s, id, "collaborator"))
	r := TimelineTicketCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "forbidden-create", collaborator), TimelineTicketInput: TimelineTicketInput{ExternalScope: "https://example.com/tracker", ExternalID: "two", Key: "EX-2"}}
	_, err = s.TimelineV3Operation(id, "ticket_create", "", &r)
	requireTimelineCode(t, err, "lead_required")
	c := TimelineCandidateCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "collaborator-candidate", collaborator), TimelineCandidate: TimelineCandidate{Branch: "feature", Commit: "abc", Base: "development"}}
	candidate, err := s.TimelineV3Operation(id, "candidate_create", "", &c)
	if err != nil {
		t.Fatal("accepted collaborator cannot capture candidate", err)
	}
	c.TimelineNativeOperation = domainBase(t, s, id, "local-candidate", v.ForegroundProducerID)
	c.Actor = TimelineActor{Kind: "local_viewer"}
	_, err = s.TimelineV3Operation(id, "candidate_create", "", &c)
	if err == nil {
		t.Fatal("browser gained publication authority")
	}
	for i := 0; i < 4; i++ {
		c.TimelineNativeOperation = domainBase(t, s, id, fmt.Sprintf("repeat-candidate-%d", i), collaborator)
		out, err := s.TimelineV3Operation(id, "candidate_create", "", &c)
		if err != nil || out.CandidateID != candidate.CandidateID || *out.Created {
			t.Fatal("candidate dedup", err)
		}
	}
	d := domainDelivery(t, s, id)
	for i := 0; i < 4; i++ {
		current := domainCurrent(t, s, id)
		input := current.Deliveries[0]
		r := TimelineDeliveryCreateRequest{TimelineNativeOperation: domainBase(t, s, id, fmt.Sprintf("repeat-delivery-%d", i), v.ForegroundProducerID), RepositoryURL: input.RepositoryURL, ExternalID: input.ExternalID, TimelineDeliverySnapshot: input.TimelineDeliverySnapshot}
		out, err := s.TimelineV3Operation(id, "delivery_create", "", &r)
		if err != nil || out.DeliveryID != d || *out.Created {
			t.Fatal("delivery dedup", err)
		}
	}
	p := TimelineTicketPatchRequest{TimelineNativeOperation: domainBase(t, s, id, "ticket-url", v.ForegroundProducerID), URL: timelineStringPointer("https://EXAMPLE.com:443/issue/one")}
	_, err = s.TimelineV3Operation(id, "ticket_patch", v.Tickets[0].TicketID, &p)
	if err != nil {
		t.Fatal(err)
	}
	before := domainCurrent(t, s, id)
	p.TimelineNativeOperation = domainBase(t, s, id, "url-collision", v.ForegroundProducerID)
	p.URL = timelineStringPointer("https://example.com/issue/one")
	_, err = s.TimelineV3Operation(id, "ticket_patch", v.Tickets[1].TicketID, &p)
	requireTimelineCode(t, err, "entity_conflict")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("collision partially wrote")
	}
	p.TimelineNativeOperation = domainBase(t, s, id, "unknown-ticket", v.ForegroundProducerID)
	_, err = s.TimelineV3Operation(id, "ticket_patch", d, &p)
	requireTimelineCode(t, err, "unknown_ticket")
	req.OperationID = "duplicate-bootstrap"
	req.Source.NativeSessionID = "another"
	req.Tickets = append(req.Tickets, req.Tickets[0])
	_, err = s.TimelineCreate(req)
	if err == nil {
		t.Fatal("duplicate bootstrap accepted")
	}
}

func TestTimelineCatalogDurableFaultReplay(t *testing.T) {
	for _, phase := range []string{"before_file_sync", "after_file_sync", "before_rename", "after_rename", "before_directory_sync", "after_directory_sync"} {
		t.Run(phase, func(t *testing.T) {
			s, v, _ := newV3Fixture(t)
			id := v.ViewerSessionID
			r := TimelineCandidateCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "fault-candidate", v.ForegroundProducerID), TimelineCandidate: TimelineCandidate{Branch: "feature", Commit: "abc", Base: "development"}}
			renamed := false
			s.timelineFault = func(at, path string) error {
				if path == s.timelineRecordPath(id) && at == "after_rename" {
					renamed = true
				}
				if at == phase && (path == s.timelineRecordPath(id) || renamed) {
					return errors.New("injected")
				}
				return nil
			}
			_, err := s.TimelineV3Operation(id, "candidate_create", "", &r)
			if err == nil {
				t.Fatal("injected failure succeeded")
			}
			if renamed {
				requireTimelineCode(t, err, "outcome_unknown")
			}
			s.timelineFault = nil
			out, err := s.TimelineV3Operation(id, "candidate_create", "", &r)
			if err != nil {
				t.Fatal(err)
			}
			current := domainCurrent(t, s, id)
			if len(current.Candidates) != 1 {
				t.Fatal("fault duplicated registration")
			}
			reopened, err := Open(Options{StateDir: filepath.Dir(s.Dir())})
			if err != nil {
				t.Fatal(err)
			}
			replay, err := reopened.TimelineV3Operation(id, "candidate_create", "", &r)
			if err != nil || !reflect.DeepEqual(out, replay) {
				t.Fatal("fault receipt lost", err)
			}
		})
	}
}

func TestTimelineCatalogCeilingPreservesClosure(t *testing.T) {
	s, v, _ := newV3Fixture(t)
	id := v.ViewerSessionID
	for i := 0; i < TimelineMaxCatalogRecords; i++ {
		domainCandidate(t, s, id, fmt.Sprintf("commit-%d", i))
	}
	before := domainCurrent(t, s, id)
	r := TimelineCandidateCreateRequest{TimelineNativeOperation: domainBase(t, s, id, "over-limit", v.ForegroundProducerID), TimelineCandidate: TimelineCandidate{Branch: "feature", Commit: "over-limit", Base: "development"}}
	_, err := s.TimelineV3Operation(id, "candidate_create", "", &r)
	requireTimelineCode(t, err, "limit")
	if !reflect.DeepEqual(before, domainCurrent(t, s, id)) {
		t.Fatal("limit partially wrote")
	}
	_, err = s.TimelineAppend(id, TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "close", Actor: timelineTestActor(s, id), Reason: "Explicit closure", Kind: "closure", Title: "Complete", Summary: "Retain canonical records", TicketIDs: []string{}, Closure: &TimelineClosure{Reason: "completed", Result: "Verified"}})
	if err != nil {
		t.Fatal("catalog ceiling blocked closure", err)
	}
}

func TestTimelineArtifactRejectsUnresolvedCandidateBeforeAnyWrite(t *testing.T) {
	for _, reuse := range []bool{false, true} {
		for _, kind := range []string{"unknown", "foreign", "wrong-type"} {
			t.Run(fmt.Sprintf("%s/reuse=%v", kind, reuse), func(t *testing.T) {
				s, v, request := newV3Fixture(t)
				id := v.ViewerSessionID
				candidate := domainCandidate(t, s, id, "original")
				var content bytes.Buffer
				if err := png.Encode(&content, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
					t.Fatal(err)
				}
				metadata := TimelineArtifactMetadata{SchemaVersion: 3, OperationID: "seed-artifact", Actor: timelineTestActor(s, id), CandidateID: candidate, CapturedAt: "2026-01-01T00:00:00Z", Viewport: "desktop", Theme: "dark", MediaType: "image/png"}
				if reuse {
					if _, err := s.TimelineAddArtifact(id, metadata, content.Bytes()); err != nil {
						t.Fatal(err)
					}
				}
				switch kind {
				case "unknown":
					metadata.CandidateID = "candidate-unknown"
				case "foreign":
					request.OperationID = "foreign-work"
					request.Source.NativeSessionID = "foreign-work"
					foreign, err := s.TimelineCreate(request)
					if err != nil {
						t.Fatal(err)
					}
					metadata.CandidateID = domainCandidate(t, s, foreign.Session.ViewerSessionID, "foreign")
				case "wrong-type":
					metadata.CandidateID = domainDelivery(t, s, id)
				}
				metadata.OperationID = "invalid-artifact"
				before := stateManifest(t, filepath.Dir(s.Dir()))
				recordBefore := timelineClone(s.timeline[id])
				_, err := s.TimelineAddArtifact(id, metadata, content.Bytes())
				requireTimelineCode(t, err, "unknown_candidate")
				if !reflect.DeepEqual(before, stateManifest(t, filepath.Dir(s.Dir()))) {
					t.Fatal("rejected candidate changed persistent files or mtimes")
				}
				if !reflect.DeepEqual(recordBefore, s.timeline[id]) {
					t.Fatal("rejected candidate changed session, audit or receipts")
				}
			})
		}
	}
}
