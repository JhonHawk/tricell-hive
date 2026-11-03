package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

func workEngineFixture(t *testing.T) (*viewFixture, viewWorkCommand) {
	t.Helper()
	f := newViewFixture(t)
	root, err := canonicalWorkDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = exec.Command("git", "-C", root, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("register", "--dir", root)
	f.mustView("start")
	dir := filepath.Join(root, "A")
	if err = os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	return f, viewWorkCommand{Action: "open", Options: management.Options{Scope: "user", StateDir: f.stateDir}, Project: root, Change: dir, Host: "codex", NativeSessionID: "work-engine-fixture", Key: "A", Title: "Work A", Summary: "Isolated work engine contract", Limit: 20}
}
func engineWorkOpen(t *testing.T, c viewWorkCommand) viewWorkResult {
	t.Helper()
	var out bytes.Buffer
	if err := executeViewWork(c, &out); err != nil {
		t.Fatal(err)
	}
	var r viewWorkResult
	if json.Unmarshal(out.Bytes(), &r) != nil {
		t.Fatal("invalid output")
	}
	return r
}
func TestWorkEngineBindingsAndReadOnlyListing(t *testing.T) {
	f, c := workEngineFixture(t)
	a := engineWorkOpen(t, c)
	b := c
	b.Change = filepath.Join(c.Project, "B")
	os.Mkdir(b.Change, 0700)
	b.Key, b.Title = "B", "Work B"
	other := engineWorkOpen(t, b)
	if a.WorkID == other.WorkID {
		t.Fatal("two keys collapsed")
	}
	c.Title = "Edited local title"
	if again := engineWorkOpen(t, c); again.WorkID != a.WorkID || again.SessionVersion != a.SessionVersion {
		t.Fatal("reopen mutated work")
	}
	mismatch := c
	mismatch.Key = "wrong"
	if err := executeViewWork(mismatch, &bytes.Buffer{}); err == nil {
		t.Fatal("directory silently selected different key")
	}
	before, _ := viewIndex(f.info())
	list := c
	list.Action = "list"
	var output bytes.Buffer
	if err := executeViewWork(list, &output); err != nil {
		t.Fatal(err)
	}
	after, _ := viewIndex(f.info())
	if len(before.TimelineSessions) != len(after.TimelineSessions) {
		t.Fatal("list mutated inventory")
	}
	if !strings.Contains(output.String(), a.WorkID) || !strings.Contains(output.String(), other.WorkID) {
		t.Fatal("list omitted independent works")
	}
	c.Action, c.WorkID = "open", a.WorkID
	c.Change = b.Change
	if err := executeViewWork(c, &bytes.Buffer{}); err == nil {
		t.Fatal("explicit mismatch accepted")
	}
	c.Change = filepath.Join(c.Project, "moved")
	os.Mkdir(c.Change, 0700)
	c.Action, c.Reason = "bind", "Explicit source migration"
	engineWorkOpen(t, c)
	c.Action = "open"
	engineWorkOpen(t, c)
	if err := viewer.PrepareState(f.stateDir); err != nil {
		t.Fatal("publisher corrupted server inventory", err)
	}
}
func TestWorkEngineOriginalCreateReplay(t *testing.T) {
	_, c := workEngineFixture(t)
	e, err := openViewWorkEngine(c)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	source := viewer.TimelineSource{ProjectRoot: c.Project, Host: "codex", NativeSessionID: c.NativeSessionID}
	req := viewer.TimelineCreateSessionRequest{SchemaVersion: 3, OperationID: "original-create", Source: source, Title: c.Title, Summary: c.Summary, Tickets: []viewer.TimelineTicketInput{}, WorkKey: &c.Key}
	raw, _ := json.Marshal(req)
	i, err := e.prepareIntent("open", "", source, viewWorkOpenContext{Binding: viewWorkBinding{Project: c.Project, Change: c.Change, Key: c.Key, Source: source}, Request: raw})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.appendStep(i, "POST", "/api/timeline/sessions", "application/json", raw); err != nil {
		t.Fatal(err)
	}
	// Commit remotely, then lose the acknowledgment as if the process exited.
	if _, _, err = viewWorkHTTP(e.info, "POST", "/api/timeline/sessions", "application/json", raw); err != nil {
		t.Fatal(err)
	}
	c.Title = "Edited after uncertain response"
	s, err := e.open(c)
	if err != nil {
		t.Fatal(err)
	}
	if s.Title != req.Title || s.Version != 1 {
		t.Fatal("retry changed initial request")
	}
	if len(e.journal.Pending) != 0 || len(e.journal.History) != 1 {
		t.Fatal("recovery not durably classified")
	}
}
func TestWorkEngineRejectsForeignCreateReceipt(t *testing.T) {
	f, c := workEngineFixture(t)
	opened := engineWorkOpen(t, c)
	raw, err := viewTimelineHTTP(f.info(), "GET", "/api/timeline/sessions/"+opened.WorkID, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write(raw) }))
	defer server.Close()
	dir := t.TempDir()
	e := &viewWorkEngine{dir: dir, journal: newWorkJournal(), info: timelineCLIInfo(t, server)}
	source := viewer.TimelineSource{ProjectRoot: "/different", Host: "codex", NativeSessionID: "different"}
	i, err := e.prepareIntent("open", "", source, nil)
	if err != nil {
		t.Fatal(err)
	}
	request := viewer.TimelineCreateSessionRequest{Source: source, OperationID: "different"}
	body, _ := json.Marshal(request)
	e.appendStep(i, "POST", "/api/timeline/sessions", "application/json", body)
	if _, err = e.executeStep(i, 0); err == nil {
		t.Fatal("accepted unrelated successful response as original create")
	}
}

func TestWorkEngineReconcileAuthenticatesBeforeReplay(t *testing.T) {
	f, c := workEngineFixture(t)
	opened := engineWorkOpen(t, c)
	raw, err := viewTimelineHTTP(f.info(), "GET", "/api/timeline/sessions/"+opened.WorkID, "", nil, viewer.TimelineMaxBody)
	if err != nil {
		t.Fatal(err)
	}
	mutations := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Write(raw)
			return
		}
		mutations++
		w.WriteHeader(409)
		w.Write([]byte(`{"code":"version_conflict"}`))
	}))
	defer server.Close()
	e := &viewWorkEngine{dir: t.TempDir(), info: timelineCLIInfo(t, server), journal: newWorkJournal()}
	i := &viewWorkIntent{ID: "pending-reconcile", WorkID: opened.WorkID, Source: viewer.TimelineSource{ProjectRoot: c.Project, Host: "codex", NativeSessionID: c.NativeSessionID}, Action: "stage", State: "needs_reconciliation", Steps: []viewWorkStep{{Method: "POST", Path: "/api/timeline/sessions/" + opened.WorkID + "/events", ContentType: "application/json", Body: []byte(`{"operation_id":"uncertain-stage"}`), Attempted: true}}}
	e.journal.Pending = append(e.journal.Pending, i)
	c.Action, c.WorkID, c.Pending, c.Reason = "reconcile", opened.WorkID, i.ID, "Explicit conflict archive"
	c.NativeSessionID = "foreign-unaccepted"
	if err = e.recover(c, &bytes.Buffer{}); err == nil {
		t.Fatal("foreign source accepted")
	}
	if mutations != 0 {
		t.Fatalf("foreign source replayed %d mutations before membership check", mutations)
	}
}

func TestWorkEngineConfirmedHistoryReleasesFrozenBytes(t *testing.T) {
	e := &viewWorkEngine{dir: t.TempDir(), journal: newWorkJournal()}
	source := viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}
	frozen := bytes.Repeat([]byte("x"), 8*256*1024)
	context, _ := json.Marshal(workPublicationIntent{Documents: []workDocument{{Bytes: frozen}}})
	for n := 0; n < 4; n++ {
		i, err := e.prepareIntent("publish", strings.Repeat("a", 64), source, json.RawMessage(context))
		if err != nil {
			t.Fatal("allowed next revision blocked by confirmed history", err)
		}
		if _, err = e.appendStep(i, "POST", "/api/timeline/sessions/"+i.WorkID+"/documents", "application/json", frozen); err != nil {
			t.Fatal("allowed revision exhausted journal", err)
		}
		if n == 0 {
			before, _ := os.ReadFile(filepath.Join(e.dir, "journal.json"))
			if _, err := e.prepareIntent("publish", strings.Repeat("b", 64), source, json.RawMessage(context)); err == nil {
				t.Fatal("admitted work without acknowledgment reserve")
			}
			after, _ := os.ReadFile(filepath.Join(e.dir, "journal.json"))
			if !bytes.Equal(before, after) || len(e.journal.Pending) != 1 {
				t.Fatal("capacity rejection changed pending evidence")
			}
		}
		if err = e.finishIntent(i); err != nil {
			t.Fatal("confirmed revision retained frozen bytes", err)
		}
	}
	raw, err := os.ReadFile(filepath.Join(e.dir, "journal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > 64*1024 {
		t.Fatal("confirmed local history retained publication bodies")
	}
}

func TestWorkEnginePublishesThreeMaximumDocumentSets(t *testing.T) {
	f := newWorkPublishFixture(t)
	work := ""
	for revision := 0; revision < 3; revision++ {
		args := []string{"publish", "--change", f.change, "--stage", "plan", "--explicit-documents"}
		if work == "" {
			args = append(args, "--project", f.root, "--key", "large-work", "--title", "Large bounded publication", "--summary", "Three valid maximum document sets", "--reason", "Start explicit plan publication")
		} else {
			args = append(args, "--work-id", work)
		}
		for n := 0; n < 8; n++ {
			name := fmt.Sprintf("document-%d.md", n)
			data := bytes.Repeat([]byte{byte('a' + revision), byte('0' + n)}, 128*1024)
			if err := os.WriteFile(filepath.Join(f.change, name), data, 0600); err != nil {
				t.Fatal(err)
			}
			args = append(args, "--document", name)
		}
		result, err := f.run(t, args...)
		if err != nil {
			t.Fatalf("maximum publication %d: %v", revision+1, err)
		}
		work = result.WorkID
	}
	j, err := readWorkJournal(filepath.Join(f.state, "publisher", "v1"))
	if err != nil || len(j.Pending) != 0 {
		t.Fatal("maximum publication recovery state", err)
	}
	if len(f.observe(t, work).Documents) != 24 {
		t.Fatal("maximum revisions lost documents")
	}
}
