package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// Persistence owns corruption, locking and atomic retained archival contracts.
func TestWorkJournalPreservesCorruptionAndUnknownVersion(t *testing.T) {
	for _, raw := range []string{`{`, `{"schema_version":99}`, `{"schema_version":1}`} {
		dir := t.TempDir()
		path := filepath.Join(dir, "journal.json")
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := readWorkJournal(dir); err == nil {
			t.Fatal("accepted invalid journal")
		}
		got, _ := os.ReadFile(path)
		if string(got) != raw {
			t.Fatal("changed corrupt journal")
		}
	}
}
func TestWorkJournalIndependentNonblockingLock(t *testing.T) {
	dir := t.TempDir()
	a, err := lockWorkPublisher(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Release()
	if b, err := lockWorkPublisher(dir); err == nil {
		b.Release()
		t.Fatal("second writer acquired lock")
	}
}
func TestWorkJournalArchiveRequiresDefinitiveRejection(t *testing.T) {
	j := newWorkJournal()
	i := &viewWorkIntent{ID: "intent", State: "needs_reconciliation", Steps: []viewWorkStep{{Method: "POST", Path: "/api/test", Body: []byte("frozen")}}}
	j.Pending = append(j.Pending, i)
	if err := j.archiveRejected("intent", "explicit reason"); err == nil {
		t.Fatal("archived unknown outcome")
	}
	i.Steps[0].Rejection = "current_predecessor_required"
	i.State = "rejected_conflict"
	if err := j.archiveRejected("intent", "explicit reason"); err != nil {
		t.Fatal(err)
	}
	if len(j.Pending) != 0 || len(j.History) != 1 || string(j.History[0].Steps[0].Body) != "frozen" {
		t.Fatal("lost archived bytes")
	}
}

func TestWorkJournalRetainsSingleArchiveClassification(t *testing.T) {
	dir := t.TempDir()
	j := newWorkJournal()
	source := viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}
	request := []byte(`{"schema_version":3,"operation_id":"frozen-patch","actor":{"kind":"native_producer","producer_id":"producer-fixture"},"expected_version":1,"state":"closed","reason":"completed: Retained reason"}`)
	work := strings.Repeat("a", 64)
	receipt, _ := json.Marshal(viewer.TimelineOperationResult{SchemaVersion: 3, OperationID: "frozen-patch", Kind: "session_patch", ViewerSessionID: work, SessionVersion: 2})
	i := &viewWorkIntent{ID: "rejected", WorkID: work, Source: source, Action: "close", Context: json.RawMessage(`{"command":{"Action":"close","WorkID":"` + work + `","Host":"codex","NativeSessionID":"fixture"}}`), State: "rejected_conflict", Steps: []viewWorkStep{{Method: "PATCH", Path: "/api/timeline/sessions/" + work, ContentType: "application/json", Body: request, Attempted: true, Response: receipt}, {Method: "PATCH", Path: "/api/timeline/sessions/" + work, ContentType: "application/json", Body: request, Attempted: true, Rejection: "version_conflict"}}}
	j.Pending = append(j.Pending, i)
	if err := writeWorkJournal(dir, j); err != nil {
		t.Fatal(err)
	}
	// A crash before the rename retains active; after rename it retains archived.
	before, err := readWorkJournal(dir)
	if err != nil || len(before.Pending) != 1 {
		t.Fatal("lost active classification", err)
	}
	if err = j.archiveRejected(i.ID, "Explicit conflict recovery"); err != nil {
		t.Fatal(err)
	}
	if err = writeWorkJournal(dir, j); err != nil {
		t.Fatal(err)
	}
	after, err := readWorkJournal(dir)
	if err != nil || len(after.Pending) != 0 || len(after.History) != 1 {
		t.Fatal("ambiguous archive classification", err)
	}
	if !bytes.Equal(after.History[0].Steps[0].Response, receipt) || !bytes.Equal(after.History[0].Steps[1].Body, request) {
		t.Fatal("lost recovery evidence")
	}
	info, _ := os.Stat(filepath.Join(dir, "journal.json"))
	if info.Mode().Perm() != 0600 {
		t.Fatal("journal mode")
	}
}

func TestWorkPublisherContendedChild(t *testing.T) {
	dir := os.Getenv("HIVE_TEST_PUBLISHER_LOCK")
	if dir == "" {
		return
	}
	lock, err := lockWorkPublisher(dir)
	if os.Getenv("HIVE_TEST_PUBLISHER_BUSY") == "1" {
		if err == nil {
			lock.Release()
			t.Fatal("child acquired held publisher lock")
		}
		if !strings.Contains(err.Error(), "busy") {
			t.Fatal(err)
		}
	} else {
		if err != nil {
			t.Fatal(err)
		}
		lock.Release()
	}
}
func TestWorkPublisherLockAcrossProcesses(t *testing.T) {
	dir := t.TempDir()
	lock, err := lockWorkPublisher(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Release()
	run := func(busy string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWorkPublisherContendedChild$")
		cmd.Env = append(os.Environ(), "HIVE_TEST_PUBLISHER_LOCK="+dir, "HIVE_TEST_PUBLISHER_BUSY="+busy)
		if raw, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("lock subprocess: %v %s", err, raw)
		}
	}
	run("1")
	lock.Release()
	run("0")
}

func TestWorkJournalRejectsMalformedBindingWithoutChangingBytes(t *testing.T) {
	for _, binding := range []string{`{}`, `{"work_id":"bad","project":"/root","change":"/root/change","source":{"project_root":"/root","host":"codex","native_session_id":"native"}}`, `{"work_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","project":"relative","change":"/root/change","source":{"project_root":"/root","host":"codex","native_session_id":"native"}}`} {
		dir := t.TempDir()
		raw := []byte(`{"schema_version":1,"bindings":[` + binding + `],"pending":[],"history":[]}`)
		path := filepath.Join(dir, "journal.json")
		os.WriteFile(path, raw, 0600)
		if _, err := readWorkJournal(dir); err == nil {
			t.Fatal("accepted semantically malformed binding")
		}
		got, _ := os.ReadFile(path)
		if string(got) != string(raw) {
			t.Fatal("changed invalid journal")
		}
	}
}

func TestWorkJournalRejectsUnsafeLinksAndRoutes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")
	raw, _ := json.Marshal(newWorkJournal())
	target := filepath.Join(dir, "target.json")
	os.WriteFile(target, raw, 0600)
	if err := os.Link(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted hardlinked journal")
	}
	os.Remove(path)
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted symlink journal")
	}
	i := &viewWorkIntent{ID: "pending", WorkID: strings.Repeat("a", 64), Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Action: "stage", State: "needs_reconciliation", Context: json.RawMessage(`{}`), Steps: []viewWorkStep{}}
	for _, step := range []viewWorkStep{{Method: "DELETE", Path: "/api/timeline/sessions/" + i.WorkID}, {Method: "POST", Path: "/api/timeline/sessions/" + strings.Repeat("b", 64) + "/events"}, {Method: "POST", Path: "/api/timeline/sessions/" + i.WorkID + "/events?override=1"}} {
		i.Steps = []viewWorkStep{step}
		j := newWorkJournal()
		j.Pending = append(j.Pending, i)
		if err := validateWorkJournal(j); err == nil {
			t.Fatal("accepted corrupt mutation route")
		}
	}
}

func TestWorkJournalRejectsMissingSemanticIntent(t *testing.T) {
	j := newWorkJournal()
	j.Pending = append(j.Pending, &viewWorkIntent{ID: "pending", WorkID: strings.Repeat("a", 64), Source: viewer.TimelineSource{ProjectRoot: "/fixture", Host: "codex", NativeSessionID: "fixture"}, Action: "stage", State: "prepared", Context: json.RawMessage(`{}`), Steps: []viewWorkStep{}})
	dir := t.TempDir()
	raw, _ := json.Marshal(j)
	path := filepath.Join(dir, "journal.json")
	os.WriteFile(path, raw, 0600)
	if _, err := readWorkJournal(dir); err == nil {
		t.Fatal("accepted missing semantic intent fields")
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, raw) {
		t.Fatal("modified corrupt semantic intent")
	}
}
