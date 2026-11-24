package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/viewer"
)

// Antigravity reads every hook's stdout as a verdict, so `hive hook antigravity`
// is the one hook that prints: the smallest object its event accepts, on every
// path. The payload follows agy's hooks documentation (camelCase, no event name).

const agyPayload = `{"conversationId":"agy-conv-1","workspacePaths":["/work/app","/work/other"],"transcriptPath":"/none","modelName":"auto","lastUserInput":"hi","invocationNum":2,"executionNum":1,"terminationReason":"NO_TOOL_CALL"}`

// runHookCapture runs runHook with os.Stdout replaced and returns what it printed.
func runHookCapture(t *testing.T, stdin io.Reader, args ...string) string {
	t.Helper()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = outW
	func() {
		defer func() { os.Stdout = old }()
		runHook(args, stdin)
	}()
	outW.Close()
	o, _ := io.ReadAll(outR)
	return string(o)
}

func agyHook(t *testing.T, state, payload string, args ...string) string {
	t.Helper()
	return runHookCapture(t, strings.NewReader(payload), append(args, "--state-dir", state)...)
}

func queuedHostEvents(t *testing.T, state string) []viewer.TimelineHostEvent {
	t.Helper()
	dir := filepath.Join(state, "hooks", "queue")
	var out []viewer.TimelineHostEvent
	for _, name := range hostEventQueueNames(dir) {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		var e viewer.TimelineHostEvent
		if err := json.Unmarshal(raw, &e); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func TestAntigravityEachEventQueuesItsCanonicalEventAndPrintsItsReply(t *testing.T) {
	for _, c := range []struct{ event, reply, wantType string }{
		{"SessionStart", "{}\n", "session_started"},
		{"PreInvocation", "{}\n", "activity"},
		{"Stop", `{"decision":"stop"}` + "\n", "activity"},
	} {
		t.Run(c.event, func(t *testing.T) {
			withHookHost(t, "agy", nil)
			state := t.TempDir()
			if out := agyHook(t, state, agyPayload, "antigravity", c.event); out != c.reply {
				t.Fatalf("stdout %q, want %q", out, c.reply)
			}
			events := queuedHostEvents(t, state)
			if len(events) != 1 {
				t.Fatalf("queued %+v", events)
			}
			e := events[0]
			if e.EventType != c.wantType || e.Source.Host != "antigravity" || e.Source.NativeSessionID != "agy-conv-1" || e.Source.ProjectRoot != "/work/app" {
				t.Fatalf("event %+v", e)
			}
		})
	}
}

func TestAntigravityStopNeverEndsASession(t *testing.T) {
	withHookHost(t, "agy", nil)
	state := t.TempDir()
	agyHook(t, state, agyPayload, "antigravity", "SessionStart")
	agyHook(t, state, agyPayload, "antigravity", "PreInvocation")
	agyHook(t, state, agyPayload, "antigravity", "Stop")
	agyHook(t, state, agyPayload, "antigravity", "SessionEnd") // agy has no such event
	for _, e := range queuedHostEvents(t, state) {
		if e.EventType == "session_ended" {
			t.Fatalf("no antigravity hook may end a session: %+v", e)
		}
	}
	doc := mapSessions(t, state)
	if doc.Host != "antigravity" || len(doc.Sessions) != 1 || doc.Sessions[0].SessionID != "agy-conv-1" || doc.Sessions[0].EndedAt != "" {
		t.Fatalf("map %+v", doc)
	}
	if id, err := resolvedSession(t, state); err != nil || id.Host != "antigravity" || id.NativeSessionID != "agy-conv-1" {
		t.Fatalf("resolver %+v %v", id, err)
	}
}

func TestAntigravityPreInvocationRefreshesTheMapWhenSessionStartDidNotFire(t *testing.T) {
	withHookHost(t, "agy", nil)
	state := t.TempDir()
	agyHook(t, state, agyPayload, "antigravity", "PreInvocation")
	if doc := mapSessions(t, state); doc.Host != "antigravity" || len(doc.Sessions) != 1 || doc.Sessions[0].SessionID != "agy-conv-1" {
		t.Fatalf("map %+v", doc)
	}
}

func TestAntigravityTakesTheDirectoryFromWorkspacePathsNotTheWorkingDirectory(t *testing.T) {
	withHookHost(t, "agy", nil)
	state := t.TempDir()
	t.Chdir(t.TempDir())
	agyHook(t, state, `{"conversationId":"agy-conv-2","workspacePaths":["/elsewhere/project"]}`, "antigravity", "SessionStart")
	events := queuedHostEvents(t, state)
	if len(events) != 1 || events[0].Source.ProjectRoot != "/elsewhere/project" {
		t.Fatalf("events %+v", events)
	}
}

func TestAntigravityPrintsTheAcceptedObjectOnEveryPath(t *testing.T) {
	cases := map[string]struct {
		stdin string
		args  []string
		want  string
	}{
		"unknown event":          {agyPayload, []string{"antigravity", "Nope"}, "{}\n"},
		"no event":               {agyPayload, []string{"antigravity"}, "{}\n"},
		"extra argument":         {agyPayload, []string{"antigravity", "Stop", "extra"}, `{"decision":"stop"}` + "\n"},
		"agy alias":              {agyPayload, []string{"agy", "Stop"}, `{"decision":"stop"}` + "\n"},
		"empty input":            {``, []string{"antigravity", "SessionStart"}, "{}\n"},
		"not JSON on Stop":       {`not json`, []string{"antigravity", "Stop"}, `{"decision":"stop"}` + "\n"},
		"no conversation":        {`{"workspacePaths":["/w"]}`, []string{"antigravity", "PreInvocation"}, "{}\n"},
		"flag without a value":   {agyPayload, []string{"antigravity", "Stop", "--state-dir"}, `{"decision":"stop"}` + "\n"},
		"state dir not a path":   {agyPayload, []string{"antigravity", "SessionStart", "--state-dir", "/nonexistent/\x00bad"}, "{}\n"},
		"relative workspace dir": {`{"conversationId":"c1","workspacePaths":["rel"]}`, []string{"antigravity", "SessionStart"}, "{}\n"},
		"flag before the event":  {agyPayload, []string{"antigravity", "--home", "/nonexistent/home", "Stop"}, `{"decision":"stop"}` + "\n"},
		"other host, agy value":  {agyPayload, []string{"claude", "--home", "agy", "SessionStart"}, ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			withHookHost(t, "agy", nil)
			args := c.args
			if !strings.Contains(strings.Join(args, " "), "--state-dir") {
				args = append(append([]string{}, args...), "--state-dir", t.TempDir())
			}
			out, errOut, err := runHookCLI(t, c.stdin, args...)
			if out != c.want || errOut != "" || err != nil {
				t.Fatalf("stdout %q, stderr %q, err %v; want %q", out, errOut, err, c.want)
			}
		})
	}
}

func TestAntigravityWithoutAnAgyAncestorStillPrintsAndReports(t *testing.T) {
	// The App is not `agy`: its hook finds no host process, so no map is written,
	// and the event still reaches the queue.
	withHookHost(t, "Antigravity", nil)
	state := t.TempDir()
	if out := agyHook(t, state, agyPayload, "antigravity", "SessionStart"); out != "{}\n" {
		t.Fatalf("stdout %q", out)
	}
	if _, err := os.Stat(filepath.Join(state, "hooks", "sessions")); !os.IsNotExist(err) {
		t.Fatalf("no agy ancestor must leave no map (%v)", err)
	}
	if events := queuedHostEvents(t, state); len(events) != 1 || events[0].Source.Host != "antigravity" {
		t.Fatalf("events %+v", events)
	}
	if out := agyHook(t, state, agyPayload, "antigravity", "Stop"); out != `{"decision":"stop"}`+"\n" {
		t.Fatalf("stdout %q", out)
	}
}

func TestAntigravityPrintsAfterARecoveredPanic(t *testing.T) {
	withHookHost(t, "agy", nil)
	old := viewIdentityDepsFn
	viewIdentityDepsFn = func() viewIdentityDeps { panic("forced") }
	t.Cleanup(func() { viewIdentityDepsFn = old })
	if out := runHookCapture(t, strings.NewReader(agyPayload), "antigravity", "Stop", "--state-dir", t.TempDir()); out != `{"decision":"stop"}`+"\n" {
		t.Fatalf("stdout %q", out)
	}
	if out := runHookCapture(t, strings.NewReader(agyPayload), "antigravity", "SessionStart", "--state-dir", t.TempDir()); out != "{}\n" {
		t.Fatalf("stdout %q", out)
	}
}

func TestAntigravityPrintsWhenTheBudgetEnds(t *testing.T) {
	withHookHost(t, "agy", nil)
	hookBudget = 150 * time.Millisecond
	r, w := io.Pipe() // stdin never closes
	start := time.Now()
	out := runHookCapture(t, r, "antigravity", "Stop", "--state-dir", t.TempDir())
	if time.Since(start) > 3*time.Second {
		t.Fatal("the hook outlived its budget")
	}
	if out != `{"decision":"stop"}`+"\n" {
		t.Fatalf("stdout %q", out)
	}
	w.Close()
	time.Sleep(300 * time.Millisecond)
}

func TestOtherHostsStillPrintNothingOnEveryPath(t *testing.T) {
	withHookHost(t, "claude", nil)
	for _, args := range [][]string{{"claude", "Nope"}, {"cursor", "Stop"}, {"nope"}, {}} {
		if out := runHookCapture(t, strings.NewReader(`{"session_id":"cc-1","cwd":"/w"}`), append(args, "--state-dir", t.TempDir())...); out != "" {
			t.Fatalf("%v printed %q", args, out)
		}
	}
}
