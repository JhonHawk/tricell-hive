package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"tricell-hive/integrations/grok"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// (9) Activity counts once per turn, not once per session.

func TestHookOccurredAtTaskEventsNeverUseTheHookClock(t *testing.T) {
	oldNow := hookNow
	hookNow = func() time.Time { return time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { hookNow = oldNow })
	task := hostevent.Event{Type: hostevent.TaskCreated}
	if got := hookOccurredAt(task, time.Time{}); !got.Equal(time.Unix(0, 0)) {
		t.Fatalf("a task event with no stable datum takes the epoch, got %v", got)
	}
	if got := hookOccurredAt(task, identityTestStart); !got.Equal(identityTestStart) {
		t.Fatalf("a task event takes the host's start, got %v", got)
	}
}

func TestHookOccurredAtActivityUsesTheHookClockWhenTheHostGivesNoTime(t *testing.T) {
	now := time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)
	oldNow := hookNow
	hookNow = func() time.Time { return now }
	t.Cleanup(func() { hookNow = oldNow })
	activity := hostevent.Event{Type: hostevent.Activity}
	if got := hookOccurredAt(activity, identityTestStart); !got.Equal(now) {
		t.Fatalf("activity with no host time takes the hook clock, not the process start: %v", got)
	}
	activity.OccurredAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := hookOccurredAt(activity, identityTestStart); !got.Equal(activity.OccurredAt) {
		t.Fatalf("the host's own time wins, got %v", got)
	}
}

func TestHookGrokActivityEventsAreDistinctOperations(t *testing.T) {
	now := time.Date(2031, 1, 1, 0, 0, 0, 0, time.UTC)
	oldNow := hookNow
	hookNow = func() time.Time { return now }
	t.Cleanup(func() { hookNow = oldNow })
	build := func(prompt string) viewer.TimelineHostEvent {
		payload := `{"sessionId":"grok-session","cwd":"/work/app","promptId":"` + prompt + `"}`
		events := grok.HookEvents(hostevent.Input{Event: "UserPromptSubmit", Payload: hostevent.Decode([]byte(payload))})
		if len(events) != 1 {
			t.Fatalf("%d events", len(events))
		}
		return buildHostEvent(events[0], identityTestStart)
	}
	first, second, again := build("prompt-1"), build("prompt-2"), build("prompt-1")
	if first.Validate() != nil || first.OperationID() == second.OperationID() {
		t.Fatalf("two prompts of one session are one operation (%v)", first.Validate())
	}
	if first.OperationID() != again.OperationID() {
		t.Fatal("the same prompt is the same operation")
	}
	if first.Data.TurnID != "prompt-1" {
		t.Fatalf("data %+v", first.Data)
	}
}

func TestHookTaskEventsCarryNoTurnID(t *testing.T) {
	ev := hostevent.Event{Host: "claude-code", SessionID: "s", Cwd: "/work/app", Type: hostevent.TaskCreated, TaskID: "1", Title: "T", State: "not_started", Turn: "prompt-1"}
	if got := buildHostEvent(ev, identityTestStart); got.Data.TurnID != "" || got.Validate() != nil {
		t.Fatalf("%+v %v", got.Data, got.Validate())
	}
}

// (11) The host that ran the hook is the nearest host process.

func TestHookMapIsNeverWrittenOnAProcessBeyondAnotherHost(t *testing.T) {
	// hook(1000) -> sh(900) -> grok(800) -> claude(700): Grok runs a hook
	// registered for Claude Code.
	withIdentityDeps(t, nil, 1000, 900, 800, 700)
	withProcessNames(t, map[int]string{800: "grok", 700: "claude", 900: "sh"})
	deps := viewIdentityDepsFn()
	if pid, ok := findHostProcess(deps, "grok"); !ok || pid != 800 {
		t.Fatalf("grok: %d %v", pid, ok)
	}
	if pid, ok := findHostProcess(deps, "claude-code"); ok {
		t.Fatalf("walked past grok to %d", pid)
	}
}

// (12)

func TestHookWholeRunBudgetIsTwoSeconds(t *testing.T) {
	if hookBudgetDefault != 2*time.Second {
		t.Fatalf("budget %v", hookBudgetDefault)
	}
}

// (13) Every event carries the resolved work and active stage.

func spyOnViewer(t *testing.T, real serverInfo) (serverInfo, func() []viewer.TimelineHostEvent) {
	t.Helper()
	target, _ := url.Parse(real.url())
	proxy := httputil.NewSingleHostReverseProxy(target)
	direct := proxy.Director
	proxy.Director = func(r *http.Request) { direct(r); r.Host = target.Host }
	var mu sync.Mutex
	var seen []viewer.TimelineHostEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/host-events" {
			raw, _ := io.ReadAll(r.Body)
			var e viewer.TimelineHostEvent
			if json.Unmarshal(raw, &e) == nil {
				mu.Lock()
				seen = append(seen, e)
				mu.Unlock()
			}
			r.Body = io.NopCloser(bytes.NewReader(raw))
		}
		proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	spy := real
	spy.Port = mustPort(t, u.Port())
	return spy, func() []viewer.TimelineHostEvent {
		mu.Lock()
		defer mu.Unlock()
		return append([]viewer.TimelineHostEvent{}, seen...)
	}
}

func mustPort(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

func TestHookBindsEveryEventToTheResolvedWorkAndStage(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("bind-all")
	spy, seen := spyOnViewer(t, h.f.info())
	source := viewer.TimelineSource{ProjectRoot: h.project, Host: "claude-code", NativeSessionID: hookSyncSession}
	for _, typ := range []string{"activity", "session_started"} {
		e := viewer.TimelineHostEvent{SchemaVersion: viewer.TimelineSchemaVersion, Source: source, EventType: typ, OccurredAt: time.Now().UTC().Format(time.RFC3339Nano)}
		if postResolvedHostEvent(h.f.stateDir, spy, e) != hostSendDone {
			t.Fatal(typ + " was not sent")
		}
	}
	got := seen()
	if len(got) != 2 {
		t.Fatalf("%d events reached the viewer", len(got))
	}
	for _, e := range got {
		if e.WorkID != h.work || e.StageEventID != h.stage {
			t.Fatalf("%s named work %q stage %q, want %q %q", e.EventType, e.WorkID, e.StageEventID, h.work, h.stage)
		}
	}
}

// (10) Native tasks from before the current work belong to earlier work.

func TestReconcileSkipsTranscriptTasksOlderThanTheCurrentWork(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("older-tasks")
	home := t.TempDir()
	claudeTranscript(t, home, hookSyncSession,
		toolUse(ts(-3*time.Hour), "toolu_old", "TaskCreate", `{"subject":"From an earlier work"}`),
		toolResult(ts(-3*time.Hour), "toolu_old", "Task #5 created successfully: From an earlier work"),
		toolUse(ts(time.Second), "toolu_new", "TaskCreate", `{"subject":"From this work"}`),
		toolResult(ts(time.Second), "toolu_new", "Task #6 created successfully: From this work"))
	r := reconcileResult(t, h.must("", "reconcile", "--home", home, "--project", h.project))
	if r.Tasks != 1 || h.task("host:5") != nil || h.task("host:6") == nil {
		t.Fatalf("result %+v; old task %+v, new task %+v", r, h.task("host:5"), h.task("host:6"))
	}
}

// (14) A larger transcript is read from its tail, and the result says so.

func TestReconcileReadsTheTailOfALargerTranscript(t *testing.T) {
	home := t.TempDir()
	filler := `{"type":"cost-state","note":"` + strings.Repeat("x", 400) + `"}`
	lines := []string{
		toolUse(ts(-3*time.Minute), "toolu_1", "TaskCreate", `{"subject":"Pushed out of the window"}`),
		toolResult(ts(-3*time.Minute), "toolu_1", "Task #1 created successfully: Pushed out of the window"),
	}
	for i := 0; i < 40; i++ {
		lines = append(lines, filler)
	}
	lines = append(lines, toolUse(ts(-time.Minute), "toolu_2", "TaskUpdate", `{"taskId":"2","status":"completed"}`))
	path := claudeTranscript(t, home, "big-session", lines...)
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	facts, err := readHostSessionFile("claude-code", path, st.Size()/2)
	if err != nil || !facts.Partial {
		t.Fatalf("partial %v err %v", facts.Partial, err)
	}
	if len(facts.Tasks) != 1 || facts.Tasks[0].ID != "2" || facts.LastActivity.IsZero() {
		t.Fatalf("tail facts %+v", facts)
	}
	whole, err := readHostSessionFile("claude-code", path, st.Size())
	if err != nil || whole.Partial || len(whole.Tasks) != 2 {
		t.Fatalf("whole file: partial %v tasks %d err %v", whole.Partial, len(whole.Tasks), err)
	}

	state := t.TempDir()
	source := viewer.TimelineSource{ProjectRoot: "/work/app", Host: "claude-code", NativeSessionID: "big-session"}
	sink := &hostSink{state: state}
	r, err := reconcileHostSession(management.Options{}, sink, source, path, st.Size()/2)
	if err != nil || !r.Partial || r.Note == "" || r.Tasks != 1 {
		t.Fatalf("result %+v err %v", r, err)
	}
	if _, err := os.Stat(filepath.Join(state, "hooks", "queue")); err != nil {
		t.Fatalf("the tail's events should wait in the queue: %v", err)
	}
}
