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

func TestTimelineCLIStageContextRoundTrip(t *testing.T) {
	f := newViewFixture(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = exec.Command("git", "-C", repo, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("register", "--dir", repo)
	f.mustView("start")
	input := filepath.Join(t.TempDir(), "request.json")
	invoke := func(args []string, body any) string {
		t.Helper()
		raw, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(input, raw, 0600); e != nil {
			t.Fatal(e)
		}
		return f.mustView(append(args, "--json", input)...)
	}
	raw := invoke([]string{"session", "create"}, viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "create-context", Source: viewer.TimelineSource{ProjectRoot: repo, Host: "codex", NativeSessionID: "context-native"}, Title: "Context fixture", Summary: "Isolated CLI context", Tickets: []viewer.TimelineTicketInput{}})
	session, err := viewer.DecodeTimelineSessionResponse([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	sid := session.Session.ViewerSessionID
	actor := viewer.TimelineActor{Kind: "native_producer", ProducerID: session.Session.ForegroundProducerID}
	version := session.Session.Version
	start := viewer.TimelineAppendEventRequest{SchemaVersion: 3, OperationID: "start-empty-context", Actor: actor, Reason: "Start stage", Kind: "stage", Stage: "plan", StageState: "in_progress", ExpectedSessionVersion: &version, Title: "", Summary: "Scope", TicketIDs: []string{}}
	res, err := viewer.DecodeTimelineOperationResult([]byte(invoke([]string{"stage", "start", sid}, start)))
	if err != nil {
		t.Fatal(err)
	}
	context := "Native contextual title / título"
	patch := viewer.TimelineEventPatchRequest{SchemaVersion: 3, OperationID: "set-context", Actor: actor, ExpectedVersion: res.EventVersion, Title: &context}
	res, err = viewer.DecodeTimelineOperationResult([]byte(invoke([]string{"event", "update", sid, res.EventID}, patch)))
	if err != nil {
		t.Fatal(err)
	}
	patch.OperationID = "clear-context"
	patch.ExpectedVersion = res.EventVersion
	empty := ""
	patch.Title = &empty
	clear := invoke([]string{"event", "update", sid, res.EventID}, patch)
	before := f.mustView("session", "get", sid)
	stale := patch
	stale.OperationID = "stale-context"
	b, _ := json.Marshal(stale)
	if err = os.WriteFile(input, b, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = f.view("event", "update", sid, res.EventID, "--json", input); err == nil {
		t.Fatal("stale patch accepted")
	}
	if f.mustView("session", "get", sid) != before {
		t.Fatal("stale patch changed state")
	}
	f.mustView("stop")
	f.mustView("start")
	if got := invoke([]string{"event", "update", sid, res.EventID}, patch); got != clear {
		t.Fatal("restart replay changed receipt")
	}
	restored, err := viewer.DecodeTimelineSessionResponse([]byte(f.mustView("session", "get", sid)))
	if err != nil {
		t.Fatal(err)
	}
	if restored.Session.Events[0].Title != "" || strings.Contains(string(clear), context) {
		t.Fatal("context clear did not persist")
	}
	t.Log("CLI live isolated server: empty stage create, verbatim set, clear, stale reject without changes, stop/start, identical receipt replay")
}
