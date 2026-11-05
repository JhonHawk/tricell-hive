package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Session files below follow the layouts observed for each host (see
// view_reconcile.go); they hold no private data.

func claudeTranscript(t *testing.T, home, session string, lines ...string) string {
	t.Helper()
	dir := filepath.Join(home, ".claude", "projects", "-work-app")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, session+".jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func ts(d time.Duration) string { return time.Now().Add(d).UTC().Format(time.RFC3339Nano) }

func toolUse(at, id, name, input string) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"message":{"role":"assistant","content":[{"type":"tool_use","id":%q,"name":%q,"input":%s}]}}`, at, id, name, input)
}

func toolResult(at, id, text string) string {
	return fmt.Sprintf(`{"type":"user","timestamp":%q,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":%q,"content":%q}]}}`, at, id, text)
}

func reconcileResult(t *testing.T, out string) viewWorkReconcileResult {
	t.Helper()
	var r viewWorkReconcileResult
	if err := json.Unmarshal([]byte(out), &r); err != nil {
		t.Fatalf("reply %q: %v", out, err)
	}
	return r
}

func TestReconcileBackfillsNativeTasksAndActivityFromAClaudeTranscript(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("reconcile")
	home := t.TempDir()
	claudeTranscript(t, home, hookSyncSession,
		`{"type":"cost-state","totalCostUSD":0.1}`,
		toolUse(ts(time.Second), "toolu_1", "TaskCreate", `{"subject":"Write the tests","description":"d","activeForm":"Writing"}`),
		toolResult(ts(2*time.Second), "toolu_1", "Task #1 created successfully: Write the tests"),
		toolUse(ts(3*time.Second), "toolu_2", "TaskCreate", `{"subject":"Ship it"}`),
		toolResult(ts(4*time.Second), "toolu_2", "Task #2 created successfully: Ship it"),
		toolUse(ts(5*time.Second), "toolu_3", "TaskUpdate", `{"taskId":"1","status":"in_progress"}`),
		toolUse(ts(6*time.Second), "toolu_4", "TaskUpdate", `{"taskId":"1","status":"completed"}`),
		toolUse(ts(7*time.Second), "toolu_5", "TaskUpdate", `{"taskId":"2","status":"deleted"}`),
		`{"type":"cost-state","totalCostUSD":0.2}`)

	out := h.must("", "reconcile", "--home", home, "--project", h.project)
	r := reconcileResult(t, out)
	if r.Result != "reconciled" || r.SessionFile != hookSyncSession+".jsonl" || !r.ActivitySent || r.Sent != r.Tasks+1 || r.Queued != 0 {
		t.Fatalf("result %+v", r)
	}
	if task := h.task("host:1"); task == nil || task.Title != "Write the tests" || task.State != "completed" {
		t.Fatalf("task 1: %+v", task)
	}
	if task := h.task("host:2"); task == nil || task.Title != "Ship it" || task.State != "not_started" {
		t.Fatalf("task 2 (a deleted update is not a state): %+v", task)
	}
	// Running it again is the same operation: nothing new appears.
	versionBefore := h.session(h.work).Version
	h.must("", "reconcile", "--home", home, "--project", h.project)
	if got := h.session(h.work); len(got.Tasks) != 2 || got.Version != versionBefore {
		t.Fatalf("a second reconcile changed the work: %d tasks, version %d -> %d", len(got.Tasks), versionBefore, got.Version)
	}
	if got := h.session(h.work); got.CurrentStageEventID == nil || *got.CurrentStageEventID != h.stage {
		t.Fatal("reconcile must not change the stage")
	}
}

func TestReconcileSendsOldActivityWithItsOwnTime(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("old-activity")
	home := t.TempDir()
	claudeTranscript(t, home, hookSyncSession, toolUse(ts(-3*time.Hour), "toolu_1", "TaskUpdate", `{"taskId":"9","status":"completed"}`))
	r := reconcileResult(t, h.must("", "reconcile", "--home", home, "--project", h.project))
	if !r.ActivitySent || r.LastActivityAt == "" || r.Note != "" {
		t.Fatalf("an old activity is sent whatever its age, with its own time: %+v", r)
	}
	if strings.Contains(strings.ToLower(r.Note+r.Result), "ended") || strings.Contains(strings.ToLower(r.Note), "dead") {
		t.Fatalf("reconcile never infers a process death: %+v", r)
	}
}

func TestReconcileQueuesWhenTheViewerIsDown(t *testing.T) {
	h := newHookFixture(t)
	h.openBuild("down")
	home := t.TempDir()
	claudeTranscript(t, home, hookSyncSession,
		toolUse(ts(time.Second), "toolu_1", "TaskCreate", `{"subject":"Later"}`),
		toolResult(ts(2*time.Second), "toolu_1", "Task #4 created successfully: Later"))
	h.f.mustView("stop")
	r := reconcileResult(t, h.must("", "reconcile", "--home", home, "--project", h.project))
	if r.Sent != 0 || r.Queued != r.Tasks+1 || len(h.queued()) != r.Queued {
		t.Fatalf("%+v, queue %v", r, h.queued())
	}
	h.f.mustView("start")
	h.hook("UserPromptSubmit", h.payload(``))
	if task := h.task("host:4"); task == nil || task.Title != "Later" {
		t.Fatalf("the queued task never arrived: %+v", task)
	}
}

func TestReconcileReadsEachHostsSessionFile(t *testing.T) {
	now := time.Now().Add(-5 * time.Second)
	cases := []struct {
		host, id, rel, content string
	}{
		{"codex", "01a1a07f-0fdf-74d1-b3c6-b1aa6098e0e9", ".codex/sessions/2026/11/03/rollout-2026-11-03T02-20-44-01a1a07f-0fdf-74d1-b3c6-b1aa6098e0e9.jsonl",
			fmt.Sprintf(`{"timestamp":%q,"ordinal":1,"type":"event_msg","payload":{"type":"task_complete"}}`, now.UTC().Format(time.RFC3339Nano))},
		{"grok", "01a0c2ed-0d20-7f81-9ebe-ab11e013c966", ".grok/sessions/%2Fwork%2Fapp/01a0c2ed-0d20-7f81-9ebe-ab11e013c966/updates.jsonl",
			fmt.Sprintf(`{"timestamp":%d,"method":"_x.ai/session/update","params":{"sessionId":"01a0c2ed-0d20-7f81-9ebe-ab11e013c966"}}`, now.Unix())},
		{"pi", "01a1729a-0cfb-765f-b454-16739f5d2a0e", ".pi/agent/sessions/--work-app--/2026-10-25T10-27-40-923Z_01a1729a-0cfb-765f-b454-16739f5d2a0e.jsonl",
			fmt.Sprintf(`{"type":"message","id":"ae904382","timestamp":%q,"message":{"role":"assistant"}}`, now.UTC().Format(time.RFC3339Nano))},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, filepath.FromSlash(c.rel))
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(`{"type":"session"}`+"\n"+c.content+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			state := t.TempDir()
			var out strings.Builder
			err := runViewWork([]string{"reconcile", "--host", c.host, "--native-session-id", c.id, "--home", home, "--state-dir", state, "--project", t.TempDir()}, &out, strings.NewReader(""))
			if err != nil {
				t.Fatal(err)
			}
			r := reconcileResult(t, out.String())
			want := now.UTC().Truncate(time.Second).Format(time.RFC3339)
			if r.SessionFile != filepath.Base(path) || r.LastActivityAt != want || !r.ActivitySent || r.Tasks != 0 {
				t.Fatalf("%+v, want last activity %s", r, want)
			}
		})
	}
}

func TestReconcileHostsWithoutAReadableFileAreAResultNotAnError(t *testing.T) {
	for _, host := range []string{"opencode", "cursor", "codex"} {
		var out strings.Builder
		err := runViewWork([]string{"reconcile", "--host", host, "--native-session-id", "ses-1", "--home", t.TempDir(), "--state-dir", t.TempDir(), "--project", t.TempDir()}, &out, strings.NewReader(""))
		if err != nil {
			t.Fatalf("%s: %v", host, err)
		}
		if r := reconcileResult(t, out.String()); r.Result != "no_session_file" || r.Sent != 0 {
			t.Fatalf("%s: %+v", host, r)
		}
	}
}

func TestReconcileRejectsASessionIDThatIsNotAFileNamePart(t *testing.T) {
	for _, id := range []string{"../etc", "a/b", "*", "a b"} {
		err := runViewWork([]string{"reconcile", "--host", "claude-code", "--native-session-id", id, "--home", t.TempDir(), "--state-dir", t.TempDir(), "--project", t.TempDir()}, &strings.Builder{}, strings.NewReader(""))
		if err == nil {
			t.Fatalf("%q must be rejected", id)
		}
	}
}

func TestReconcileKeepsItsPendingArchiveForm(t *testing.T) {
	// With --pending the command is still the archive of a rejected intent and
	// still needs its action and reason.
	if _, err := parseViewWorkCommand([]string{"reconcile", "--work-id", workA, "--pending", "op-abc", "--host", "codex", "--native-session-id", "x"}, strings.NewReader("")); err == nil {
		t.Fatal("the archive form without its action and reason must be rejected")
	}
	c, err := parseViewWorkCommand([]string{"reconcile", "--host", "codex", "--native-session-id", "x"}, strings.NewReader(""))
	if err != nil || !hostReconcile(c) {
		t.Fatalf("the host form needs no work or pending: %+v %v", c, err)
	}
}
