package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/tooling/viewer"
)

// Sample payloads are derived from the documented fields in
// _support/docs/architecture/host-session-events.md; none holds private data.

const (
	hookHostPID  = 800
	hookShellPID = 900
	hookSelfPID  = 1000
)

// withHookHost installs a process tree hook(1000) -> sh(900) -> <host>(800) and
// a clock that advances one second per read, so the map's times are distinct.
func withHookHost(t *testing.T, hostProcess string, env map[string]string) {
	t.Helper()
	withIdentityDeps(t, env, hookSelfPID, hookShellPID, hookHostPID)
	withProcessNames(t, map[int]string{hookHostPID: hostProcess, hookShellPID: "sh"})
	oldNow, oldBudget := hookNow, hookBudget
	var mu sync.Mutex
	tick := identityTestStart.Add(time.Minute)
	hookNow = func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		tick = tick.Add(time.Second)
		return tick
	}
	t.Cleanup(func() { hookNow, hookBudget = oldNow, oldBudget })
}

// runHookCLI runs `hive hook` through the real entry point with stdin from a
// file, and returns what reached stdout and stderr, which must always be empty.
func runHookCLI(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	in, e := os.CreateTemp(t.TempDir(), "stdin")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = in.WriteString(stdin); e != nil {
		t.Fatal(e)
	}
	in.Seek(0, io.SeekStart)
	outR, outW, _ := os.Pipe()
	errR, errW, _ := os.Pipe()
	oldIn, oldOut, oldErr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = in, outW, errW
	func() {
		defer func() { os.Stdin, os.Stdout, os.Stderr = oldIn, oldOut, oldErr }()
		err = run(append([]string{"hook"}, args...))
	}()
	outW.Close()
	errW.Close()
	o, _ := io.ReadAll(outR)
	e2, _ := io.ReadAll(errR)
	in.Close()
	return string(o), string(e2), err
}

// mustHook runs a hook and asserts the contract every path keeps: no output, exit 0.
func mustHook(t *testing.T, state, stdin string, args ...string) {
	t.Helper()
	out, errOut, err := runHookCLI(t, stdin, append(args, "--state-dir", state)...)
	if out != "" || errOut != "" || err != nil {
		t.Fatalf("hook %v printed %q / %q or failed: %v", args, out, errOut, err)
	}
}

func mapSessions(t *testing.T, state string) hookMapDoc {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(state, "hooks", "sessions", "800.json"))
	if err != nil {
		t.Fatal(err)
	}
	var doc hookMapDoc
	if err = json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func resolvedSession(t *testing.T, state string) (viewIdentity, error) {
	t.Helper()
	return resolveViewIdentity(viewIdentityDepsFn(), state)
}

func TestHookSessionStartWritesTheMapAndPrintsNothing(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/work/app","transcript_path":"/none","hook_event_name":"SessionStart","source":"startup"}`, "claude", "SessionStart")
	doc := mapSessions(t, state)
	if doc.Host != "claude-code" || doc.PID != hookHostPID || doc.ProcessStartedAt == "" || len(doc.Sessions) != 1 || doc.Sessions[0].SessionID != "cc-1" {
		t.Fatalf("map %+v", doc)
	}
	if doc.ChangedAt != doc.Sessions[0].StartedAt {
		t.Fatalf("the first session's started_at is changed_at: %+v", doc)
	}
	id, err := resolvedSession(t, state)
	if err != nil || id.Host != "claude-code" || id.NativeSessionID != "cc-1" {
		t.Fatalf("resolver: %+v %v", id, err)
	}
	for _, p := range []string{state + "/hooks", state + "/hooks/sessions"} {
		if st, _ := os.Stat(p); st.Mode().Perm() != 0700 {
			t.Fatalf("%s mode %v", p, st.Mode().Perm())
		}
	}
	if st, _ := os.Stat(state + "/hooks/sessions/800.json"); st.Mode().Perm() != 0600 {
		t.Fatalf("map mode %v", st.Mode().Perm())
	}
}

// AC3 / T6: /clear in Claude Code changes the map's session, in either order of
// the end of the old session and the start of the new one.
func TestHookClearChangesTheMapSession(t *testing.T) {
	for name, events := range map[string][][2]string{
		"end then start": {{"SessionEnd", "cc-1"}, {"SessionStart", "cc-2"}},
		"start then end": {{"SessionStart", "cc-2"}, {"SessionEnd", "cc-1"}},
		"start only":     {{"SessionStart", "cc-2"}},
	} {
		t.Run(name, func(t *testing.T) {
			withHookHost(t, "claude", nil)
			state := t.TempDir()
			mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
			for _, e := range events {
				mustHook(t, state, `{"session_id":"`+e[1]+`","cwd":"/w","source":"clear"}`, "claude", e[0])
			}
			id, err := resolvedSession(t, state)
			if err != nil || id.NativeSessionID != "cc-2" {
				t.Fatalf("after /clear the session must be cc-2: %+v %v (%+v)", id, err, mapSessions(t, state))
			}
		})
	}
}

func TestHookSessionEndLeavesNoLiveSession(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w","reason":"exit"}`, "claude", "SessionEnd")
	doc := mapSessions(t, state)
	if doc.Sessions[0].EndedAt == "" {
		t.Fatalf("ended_at must be written on end: %+v", doc)
	}
	if _, err := resolvedSession(t, state); err == nil {
		t.Fatal("an ended session must not resolve")
	}
}

func TestHookRepeatedStartAndPromptDoNotRewriteTheMap(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
	before, _ := os.ReadFile(state + "/hooks/sessions/800.json")
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w","prompt":"x"}`, "claude", "UserPromptSubmit")
	after, _ := os.ReadFile(state + "/hooks/sessions/800.json")
	if string(before) != string(after) {
		t.Fatalf("map rewritten:\n%s\n%s", before, after)
	}
}

func TestHookWritesNothingWhenNoAncestorIsTheHost(t *testing.T) {
	withHookHost(t, "zsh", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
	if _, err := os.Stat(filepath.Join(state, "hooks", "sessions")); !os.IsNotExist(err) {
		t.Fatalf("a hook that cannot tell which ancestor is the host writes no map (%v)", err)
	}
}

func TestHookMapIsNamedAfterTheHostProcessNotTheWrapper(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w"}`, "claude", "SessionStart")
	entries, _ := os.ReadDir(filepath.Join(state, "hooks", "sessions"))
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") && e.Name() != "800.json" {
			t.Fatalf("unexpected map %s", e.Name())
		}
	}
}

func TestHookDetectsTheRealHostFromPayloadAndEnvironment(t *testing.T) {
	cases := []struct {
		name, process, hint, event, payload, wantHost, wantSession string
		env                                                        map[string]string
	}{
		{"claude settings run by Claude Code", "claude", "claude", "SessionStart", `{"session_id":"cc-1","cwd":"/w"}`, "claude-code", "cc-1", nil},
		{"codex snake_case", "codex", "codex", "SessionStart", `{"session_id":"cx-1","cwd":"/w","hook_event_name":"SessionStart"}`, "codex", "cx-1", nil},
		{"grok camelCase through its own hook", "grok", "grok", "sessionStart", `{"sessionId":"gk-1","cwd":"/w","workspaceRoot":"/w"}`, "grok", "gk-1", nil},
		{"grok running a Claude Code hook (camelCase)", "grok", "claude", "SessionStart", `{"sessionId":"gk-2","cwd":"/w"}`, "grok", "gk-2", nil},
		{"grok running a Claude Code hook (GROK_HOOK_EVENT)", "grok", "claude", "SessionStart", `{"session_id":"gk-3","cwd":"/w"}`, "grok", "gk-3", map[string]string{"GROK_HOOK_EVENT": "session_start"}},
		{"cursor running a Claude Code hook", "cursor-agent", "claude", "SessionStart", `{"conversation_id":"cu-1","workspace_roots":["/w"],"cursor_version":"2026.10.22"}`, "cursor", "cu-1", nil},
		{"cursor beforeSubmitPrompt", "cursor-agent", "cursor", "beforeSubmitPrompt", `{"conversation_id":"cu-2","workspace_roots":["/w"],"prompt":"x"}`, "cursor", "cu-2", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withHookHost(t, c.process, c.env)
			state := t.TempDir()
			mustHook(t, state, c.payload, c.hint, c.event)
			doc := mapSessions(t, state)
			if doc.Host != c.wantHost || len(doc.Sessions) != 1 || doc.Sessions[0].SessionID != c.wantSession {
				t.Fatalf("map %+v, want host %s session %s", doc, c.wantHost, c.wantSession)
			}
		})
	}
}

func TestHookOpenCodeAndPiNameTheirSessionAndWriteNoMap(t *testing.T) {
	for _, c := range []struct{ process, hint, event, payload string }{
		{"opencode", "opencode", "session.created", `{"session_id":"ses_1","directory":"/w","occurred_at":"2026-11-04T10:00:00Z"}`},
		{"node", "pi", "session_start", `{"session_id":"01a1","cwd":"/w","occurred_at":"2026-11-04T10:00:00Z"}`},
	} {
		withHookHost(t, c.process, nil)
		state := t.TempDir()
		mustHook(t, state, c.payload, c.hint, c.event)
		if _, err := os.Stat(filepath.Join(state, "hooks", "sessions")); !os.IsNotExist(err) {
			t.Fatalf("%s must not write a pid map (%v)", c.hint, err)
		}
		queued := hostEventQueueNames(filepath.Join(state, "hooks", "queue"))
		if len(queued) != 1 {
			t.Fatalf("%s: the event should wait in the queue with the viewer down: %v", c.hint, queued)
		}
	}
}

func TestHookCodexSubagentEventsDoNotChangeTheMainSession(t *testing.T) {
	withHookHost(t, "codex", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cx-1","cwd":"/w"}`, "codex", "SessionStart")
	before, _ := os.ReadFile(state + "/hooks/sessions/800.json")
	mustHook(t, state, `{"session_id":"cx-sub","agent_id":"ag-1","cwd":"/w"}`, "codex", "SessionStart")
	mustHook(t, state, `{"session_id":"cx-1","agent_id":"ag-1","cwd":"/w"}`, "codex", "SubagentStart")
	mustHook(t, state, `{"session_id":"cx-1","agent_id":"ag-1","cwd":"/w"}`, "codex", "SubagentStop")
	after, _ := os.ReadFile(state + "/hooks/sessions/800.json")
	if string(before) != string(after) {
		t.Fatalf("a subagent changed the main session's map:\n%s\n%s", before, after)
	}
	id, err := resolvedSession(t, state)
	if err != nil || id.NativeSessionID != "cx-1" {
		t.Fatalf("resolver: %+v %v", id, err)
	}
}

func TestHookClaudeIgnoresInternalSubagents(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w","agent_id":"a1","agent_type":""}`, "claude", "SubagentStart")
	if names := hostEventQueueNames(filepath.Join(state, "hooks", "queue")); len(names) != 0 {
		t.Fatalf("an internal agent must send nothing: %v", names)
	}
}

// --- the contract every path keeps: silent, exit 0 ---

func TestHookIsSilentOnEveryPath(t *testing.T) {
	withHookHost(t, "claude", nil)
	oversized := `{"session_id":"cc-1","cwd":"/w","pad":"` + strings.Repeat("x", hookInputLimit) + `"}`
	cases := map[string]struct {
		stdin string
		args  []string
	}{
		"valid, viewer down":    {`{"session_id":"cc-1","cwd":"/w"}`, []string{"claude", "SessionStart"}},
		"empty input":           {``, []string{"claude", "SessionStart"}},
		"not JSON":              {`not json at all`, []string{"claude", "SessionStart"}},
		"JSON array":            {`[1,2,3]`, []string{"claude", "SessionStart"}},
		"truncated JSON":        {`{"session_id":"cc-1"`, []string{"claude", "SessionStart"}},
		"over 1 MiB":            {oversized, []string{"claude", "SessionStart"}},
		"unknown host":          {`{"session_id":"x"}`, []string{"nope", "SessionStart"}},
		"missing event":         {`{"session_id":"x"}`, []string{"claude"}},
		"no arguments":          {`{}`, nil},
		"extra argument":        {`{}`, []string{"claude", "SessionStart", "extra"}},
		"bad session id":        {`{"session_id":"has space","cwd":"/w"}`, []string{"claude", "SessionStart"}},
		"relative cwd":          {`{"session_id":"cc-1","cwd":"rel/path"}`, []string{"claude", "UserPromptSubmit"}},
		"grok camelCase":        {`{"sessionId":"gk-1","cwd":"/w"}`, []string{"grok", "sessionStart"}},
		"claude snake_case":     {`{"session_id":"cc-1","cwd":"/w"}`, []string{"claude", "SessionEnd"}},
		"codex snake_case":      {`{"session_id":"cx-1","cwd":"/w"}`, []string{"codex", "UserPromptSubmit"}},
		"task without id":       {`{"session_id":"cc-1","task_subject":"x"}`, []string{"claude", "TaskCreated"}},
		"opencode no session":   {`{"directory":"/w"}`, []string{"opencode", "session.created"}},
		"unknown event":         {`{"session_id":"cc-1","cwd":"/w"}`, []string{"claude", "SomethingElse"}},
		"flag without a value":  {`{}`, []string{"claude", "SessionStart", "--state-dir"}},
		"state dir not a path":  {`{"session_id":"cc-1","cwd":"/w"}`, []string{"claude", "SessionStart", "--state-dir", "/nonexistent/\x00bad"}},
		"transcript is missing": {`{"session_id":"cc-1","cwd":"/w","transcript_path":"/definitely/missing"}`, []string{"claude", "UserPromptSubmit"}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			args := c.args
			if !strings.Contains(strings.Join(args, " "), "--state-dir") {
				args = append(append([]string{}, args...), "--state-dir", t.TempDir())
			}
			out, errOut, err := runHookCLI(t, c.stdin, args...)
			if out != "" || errOut != "" || err != nil {
				t.Fatalf("stdout %q, stderr %q, err %v", out, errOut, err)
			}
		})
	}
}

func TestHookOversizedInputWritesNothing(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-1","cwd":"/w","pad":"`+strings.Repeat("x", hookInputLimit)+`"}`, "claude", "SessionStart")
	if entries, _ := os.ReadDir(state); len(entries) != 0 {
		t.Fatalf("an oversized payload must be ignored whole, state has %d entries", len(entries))
	}
}

func TestHookStopsAtTheBudgetEvenWhenStdinNeverCloses(t *testing.T) {
	withHookHost(t, "claude", nil)
	hookBudget = 150 * time.Millisecond
	r, w := io.Pipe()
	start := time.Now()
	runHook([]string{"claude", "SessionStart", "--state-dir", t.TempDir()}, r)
	if time.Since(start) > 3*time.Second {
		t.Fatal("the hook outlived its budget")
	}
	// Let the abandoned reader finish before the test restores package state.
	w.Close()
	time.Sleep(300 * time.Millisecond)
}

// --- occurred_at never comes from the hook's clock ---

func TestHookOccurredAtIsStableAcrossRuns(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	transcript := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(transcript, []byte("{}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 11, 4, 9, 30, 0, 123, time.UTC)
	if err := os.Chtimes(transcript, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	payload := `{"session_id":"cc-1","cwd":"/w","task_id":"5","task_subject":"Run lint","transcript_path":"` + transcript + `"}`
	mustHook(t, state, payload, "claude", "TaskCreated")
	mustHook(t, state, payload, "claude", "TaskCreated") // the clock moved on; the event did not
	names := hostEventQueueNames(filepath.Join(state, "hooks", "queue"))
	if len(names) != 1 {
		t.Fatalf("the same event twice must be one queued event, got %v", names)
	}
	raw, _ := os.ReadFile(filepath.Join(state, "hooks", "queue", names[0]))
	var e viewer.TimelineHostEvent
	if err := json.Unmarshal(raw, &e); err != nil || e.OccurredAt != "2026-11-04T09:30:00.000000123Z" {
		t.Fatalf("occurred_at %q (%v) must be the transcript's modification time", e.OccurredAt, err)
	}
	if e.WorkID != "" || e.StageEventID != "" {
		t.Fatalf("a queued event names no work: %+v", e)
	}
}

func TestHookOccurredAtFallsBackToHostStartThenEpoch(t *testing.T) {
	ev := hostevent.Event{}
	if got := hookOccurredAt(ev, time.Time{}); !got.Equal(time.Unix(0, 0)) {
		t.Fatalf("got %v", got)
	}
	if got := hookOccurredAt(ev, identityTestStart); !got.Equal(identityTestStart) {
		t.Fatalf("got %v", got)
	}
	ev.OccurredAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := hookOccurredAt(ev, identityTestStart); !got.Equal(ev.OccurredAt) {
		t.Fatalf("the host's own time wins, got %v", got)
	}
}

// --- the map writer's rules ---

func TestHookMapAlongsideSessionKeepsChangedAtAndStaysAmbiguous(t *testing.T) {
	t0 := identityTestStart
	d := &hookMapDoc{Host: "codex", PID: 800}
	start := func(id string, replaces bool, at time.Time) {
		d.apply(hostevent.Event{Host: "codex", SessionID: id, Map: hostevent.MapStart, Replaces: replaces}, at)
	}
	start("a", false, t0)
	changed := d.ChangedAt
	start("b", false, t0.Add(time.Minute))
	if d.ChangedAt != changed || d.liveCount() != 2 {
		t.Fatalf("a session starting alongside must not move changed_at: %+v", d)
	}
	start("c", true, t0.Add(2*time.Minute))
	if d.liveCount() != 1 || d.ChangedAt == changed {
		t.Fatalf("a replacing start leaves one live session and moves changed_at: %+v", d)
	}
}

func TestHookMapEndOfASupersededSessionDoesNotMoveChangedAt(t *testing.T) {
	d := &hookMapDoc{Host: "claude-code", PID: 800}
	d.apply(hostevent.Event{SessionID: "old", Map: hostevent.MapStart, Replaces: true}, identityTestStart)
	d.apply(hostevent.Event{SessionID: "new", Map: hostevent.MapStart, Replaces: true}, identityTestStart.Add(time.Second))
	moved := d.ChangedAt
	d.apply(hostevent.Event{SessionID: "old", Map: hostevent.MapEnd}, identityTestStart.Add(2*time.Second))
	if d.ChangedAt != moved || d.liveCount() != 1 {
		t.Fatalf("%+v", d)
	}
}

func TestHookMapCapsSessionsAndDropsEndedFirst(t *testing.T) {
	d := &hookMapDoc{Host: "claude-code", PID: 800}
	at := identityTestStart
	for i := 0; i < 70; i++ {
		id := "s" + string(rune('A'+i%26)) + string(rune('a'+i/26))
		at = at.Add(time.Second)
		d.apply(hostevent.Event{SessionID: id, Map: hostevent.MapStart, Replaces: true}, at)
		at = at.Add(time.Second)
		d.apply(hostevent.Event{SessionID: id, Map: hostevent.MapEnd}, at)
	}
	if len(d.Sessions) > viewPidMapSessionCap {
		t.Fatalf("%d sessions", len(d.Sessions))
	}
}

func TestHookMapOfAReusedPidStartsFresh(t *testing.T) {
	withHookHost(t, "claude", nil)
	state := t.TempDir()
	mustHook(t, state, `{"session_id":"cc-old","cwd":"/w"}`, "claude", "SessionStart")
	// The pid is reused by a process that started an hour later.
	withHookHost(t, "claude", nil)
	withProcessStart(t, func(int) time.Time { return identityTestStart.Add(time.Hour) })
	mustHook(t, state, `{"session_id":"cc-new","cwd":"/w"}`, "claude", "SessionStart")
	doc := mapSessions(t, state)
	if len(doc.Sessions) != 1 || doc.Sessions[0].SessionID != "cc-new" {
		t.Fatalf("a crashed host's map must not survive pid reuse: %+v", doc)
	}
}

func TestHostProcessMatchesInterpreterScripts(t *testing.T) {
	for _, c := range []struct {
		host, comm, args string
		want             bool
	}{
		{"claude-code", "/Users/x/.local/bin/claude", "", true},
		{"codex", "/opt/codex-cli/codex", "", true},
		{"pi", "/usr/local/bin/node", "node /usr/local/lib/node_modules/pi/dist/pi.js -p x", true},
		{"pi", "/usr/local/bin/node", "node /srv/app/server.js", false},
		{"claude-code", "zsh", "", false},
		{"codex", "claude", "", false},
	} {
		if got := hostProcessMatches(c.host, c.comm, c.args); got != c.want {
			t.Fatalf("%+v got %v", c, got)
		}
	}
}
