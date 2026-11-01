package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/viewer"
)

func TestTimelineCLIVisualWorkflowKeepsFocusedEvidenceAcrossRestart(t *testing.T) {
	f := newViewFixture(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	input := filepath.Join(t.TempDir(), "request.json")
	invoke := func(args []string, body any) string {
		t.Helper()
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(input, data, 0600); err != nil {
			t.Fatal(err)
		}
		return f.mustView(append(args, "--json", input)...)
	}
	refs := []viewer.TimelineTicketRef{{Key: "TST-1"}, {Key: "TST-2"}}
	created := invoke([]string{"session", "create"}, viewer.TimelineCreateSessionRequest{SchemaVersion: 2, OperationID: "visual-create", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "fixture-visual"}, Title: "Visual fixture", Summary: "Two screens and a focused revision", Tickets: []viewer.TimelineTicket{{Key: "TST-1"}, {Key: "TST-2"}}})
	response, err := viewer.DecodeTimelineSessionResponse([]byte(created))
	if err != nil {
		t.Fatal(err)
	}
	sid := response.Session.ViewerSessionID
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: response.Session.ForegroundProducerID}
	result := func(raw string) viewer.TimelineOperationResult {
		t.Helper()
		r, err := viewer.DecodeTimelineOperationResult([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	request := result(invoke([]string{"event", "add", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-request", Kind: "request", Title: "Request", Summary: "Two tickets", TicketRefs: refs}))
	version := request.SessionVersion
	stage := result(invoke([]string{"stage", "start", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-build", Kind: "stage", Stage: "build", StageState: "in_progress", ExpectedSessionVersion: &version, Title: "Build", Summary: "Fixture only", TicketRefs: refs}))
	candidate := viewer.TimelineCandidate{Branch: "feature/fixture", Commit: "abc1234", Base: "development"}
	upload := func(operation, name string) viewer.TimelineOperationResult {
		t.Helper()
		file := filepath.Join("..", "viewer", "frontend", "src", "dev", "evidence", name)
		return result(invoke([]string{"artifact", "add", sid, "--file", file}, viewer.TimelineArtifactMetadata{Actor: actor, SchemaVersion: 2, OperationID: operation, CapturedAt: "2026-01-01T00:00:00Z", Candidate: candidate, Viewport: "desktop", Theme: "light", MediaType: "image/png"}))
	}
	profile, chat := upload("visual-profile", "profile-light.png"), upload("visual-chat", "chat-light.png")
	validation := result(invoke([]string{"event", "add", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-validation", Kind: "validation", Title: "First validation", Summary: "Separate screens", TicketRefs: refs, Candidate: &candidate, Views: []viewer.TimelineViewInput{
		{Title: "Profile", TicketRefs: refs[:1], Steps: []string{"Open profile"}, ExpectedResult: "Profile is readable", ImageRefs: []viewer.TimelineImageRef{{ArtifactID: profile.ArtifactID, Caption: "Profile"}}},
		{Title: "Chat", TicketRefs: refs[1:], Steps: []string{"Open chat"}, ExpectedResult: "Chat is readable", ImageRefs: []viewer.TimelineImageRef{{ArtifactID: chat.ArtifactID, Caption: "Chat"}}},
	}}))
	if len(validation.ViewIDs) != 2 {
		t.Fatal("validation lost separate view IDs")
	}
	changed := upload("visual-chat-change", "chat-mobile-after.png")
	candidate.Commit = "def5678"
	iterationRequest := viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-iteration", Kind: "iteration", Title: "Chat correction", Summary: "Only Chat changes", TicketRefs: refs[1:], Candidate: &candidate, RelatedEventID: validation.EventID, Views: []viewer.TimelineViewInput{{ViewID: validation.ViewIDs[1], Title: "Chat", TicketRefs: refs[1:], Steps: []string{"Open corrected chat"}, ExpectedResult: "Compact filter", ImageRefs: []viewer.TimelineImageRef{{ArtifactID: changed.ArtifactID, Caption: "Corrected Chat"}}, Supersedes: &viewer.TimelineSupersedes{EventID: validation.EventID, EvidenceID: validation.EvidenceID, ViewID: validation.ViewIDs[1]}}}}
	iterationRaw := invoke([]string{"event", "add", sid}, iterationRequest)
	iteration := result(iterationRaw)
	result(invoke([]string{"event", "add", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-decision", Kind: "approval_reported", Title: "Reported fixture decision", Summary: "Synthetic approval", TicketRefs: refs[1:], Target: &viewer.TimelineEvidenceTarget{Kind: "evidence", EventID: iteration.EventID, EvidenceID: iteration.EvidenceID, ViewIDs: iteration.ViewIDs}, SourceReference: "Synthetic test fixture, not a real human approval"}))
	result(invoke([]string{"event", "add", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-delivery", Kind: "delivery", Title: "Open fixture PR", Summary: "Open is not closed", TicketRefs: refs, Deliveries: []viewer.TimelineDelivery{{Kind: "pull_request", Label: "PR fixture", URL: "https://example.test/pull/42", Status: "open"}}}))
	read := func() viewer.TimelineSession {
		t.Helper()
		raw := f.mustView("session", "get", sid)
		if strings.Contains(raw, "artifact_urls") {
			t.Fatal("CLI printed browser image capabilities")
		}
		r, err := viewer.DecodeTimelineSessionResponse([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		return r.Session
	}
	before := read()
	if before.State != "active" || before.ValidationCount != 2 {
		t.Fatal("open PR changed work status or lost validation count")
	}
	copyTitle := "First validation, clarified"
	result(invoke([]string{"event", "update", sid, validation.EventID}, viewer.TimelineEventPatchRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-copy", ExpectedVersion: validation.EventVersion, Title: &copyTitle}))
	f.mustView("stop")
	f.mustView("start")
	if again := invoke([]string{"event", "add", sid}, iterationRequest); again != iterationRaw {
		t.Fatal("replay after copy edit/restart duplicated or changed evidence")
	}
	after := read()
	var first viewer.TimelineEvent
	for _, event := range after.Events {
		if event.EventID == validation.EventID {
			first = event
		}
	}
	if first.EvidenceID != validation.EvidenceID || first.Sequence != validation.Sequence || len(first.Views) != 2 || first.Views[0].ImageRefs[0].ArtifactID != profile.ArtifactID {
		t.Fatal("focused iteration or copy edit replaced original evidence")
	}
	completed := viewer.TimelineStageState("completed")
	v := after.Version
	result(invoke([]string{"stage", "update", sid, stage.EventID}, viewer.TimelineEventPatchRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-complete", ExpectedVersion: stage.EventVersion, ExpectedSessionVersion: &v, StageState: &completed}))
	result(invoke([]string{"event", "add", sid}, viewer.TimelineAppendEventRequest{Actor: actor, Reason: "Explicit visual fixture operation", SchemaVersion: 2, OperationID: "visual-close", Kind: "closure", Title: "Fixture complete", Summary: "Isolated check", TicketRefs: refs, Closure: &viewer.TimelineClosure{Reason: "completed", Result: "Verified fixture"}}))
	if read().State != "closed" {
		t.Fatal("explicit closure did not close session")
	}
}
