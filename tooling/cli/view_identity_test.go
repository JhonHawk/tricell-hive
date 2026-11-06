package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

// The default identity sources read the real environment and process tree of
// the test binary. No test may see them: every test starts with no variable and
// no parent process, and a test that needs a source installs its own.
func init() { viewIdentityDepsFn = func() viewIdentityDeps { return emptyIdentityDeps() } }

var (
	workA = strings.Repeat("a", 64)
	workB = strings.Repeat("b", 64)
)

func emptyIdentityDeps() viewIdentityDeps {
	return viewIdentityDeps{
		getenv:    func(string) string { return "" },
		selfPID:   1000,
		parentPID: func(int) (int, error) { return 0, errors.New("no parent") },
		// Every process started at identityTestStart unless a test says otherwise.
		processStart: func(int) (time.Time, error) { return identityTestStart, nil },
		processName:  func(int) (string, error) { return "", errors.New("no name") },
	}
}

var identityTestStart = time.Date(2026, 11, 4, 8, 0, 0, 0, time.UTC)

// withIdentityDeps installs deps for one test; env names are the only
// variables the resolver can see.
func withIdentityDeps(t *testing.T, env map[string]string, chain ...int) {
	t.Helper()
	old := viewIdentityDepsFn
	t.Cleanup(func() { viewIdentityDepsFn = old })
	parents := map[int]int{}
	for i := 0; i+1 < len(chain); i++ {
		parents[chain[i]] = chain[i+1]
	}
	viewIdentityDepsFn = func() viewIdentityDeps {
		d := emptyIdentityDeps()
		d.getenv = func(k string) string { return env[k] }
		if len(chain) > 0 {
			d.selfPID = chain[0]
		}
		d.parentPID = func(pid int) (int, error) {
			if p, ok := parents[pid]; ok {
				return p, nil
			}
			return 0, errors.New("no parent")
		}
		return d
	}
}

type pidMapSession struct {
	ID      string
	Started time.Time
	Ended   *time.Time
}

// writePidMap writes the fixture T6's hook will write:
// <state>/hooks/sessions/<pid>.json.
func writePidMap(t *testing.T, state, host string, pid int, changed time.Time, sessions ...pidMapSession) string {
	t.Helper()
	type entry struct {
		SessionID string  `json:"session_id"`
		StartedAt string  `json:"started_at"`
		EndedAt   *string `json:"ended_at,omitempty"`
	}
	doc := struct {
		Host      string  `json:"host"`
		PID       int     `json:"pid"`
		ChangedAt string  `json:"changed_at"`
		Started   string  `json:"process_started_at"`
		Sessions  []entry `json:"sessions"`
	}{Host: host, PID: pid, ChangedAt: changed.UTC().Format(time.RFC3339Nano), Started: identityTestStart.Format(time.RFC3339)}
	for _, s := range sessions {
		e := entry{SessionID: s.ID, StartedAt: s.Started.UTC().Format(time.RFC3339Nano)}
		if s.Ended != nil {
			v := s.Ended.UTC().Format(time.RFC3339Nano)
			e.EndedAt = &v
		}
		doc.Sessions = append(doc.Sessions, e)
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(state, "hooks", "sessions")
	if err = os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, strconv.Itoa(pid)+".json")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestResolveViewIdentityFromHostVariables(t *testing.T) {
	for _, tc := range []struct{ env, host string }{
		{"CLAUDE_CODE_SESSION_ID", "claude-code"},
		{"PI_SESSION_ID", "pi"},
		{"OPENCODE_SESSION_ID", "opencode"},
	} {
		t.Run(tc.host, func(t *testing.T) {
			withIdentityDeps(t, map[string]string{tc.env: "native-" + tc.host})
			id, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir())
			if err != nil || id.Host != tc.host || id.NativeSessionID != "native-"+tc.host {
				t.Fatalf("got %+v, %v", id, err)
			}
		})
	}
}

func TestResolveViewIdentityIgnoresUnverifiedCodexVariable(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CODEX_THREAD_ID": "thread-1"})
	if id, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir()); !errors.Is(err, errViewIdentityMissing) {
		t.Fatalf("CODEX_THREAD_ID must not be a source, got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityRejectsTwoHostVariables(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "a", "OPENCODE_SESSION_ID": "b"})
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir()); !errors.Is(err, errViewIdentityAmbiguous) {
		t.Fatalf("two host variables are ambiguous, got %v", err)
	}
}

func TestResolveViewIdentityFromPidMapNearestAncestor(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	// self 1000 -> shell 900 -> host 800 -> terminal 700; only 800 and 700 have maps.
	withIdentityDeps(t, nil, 1000, 900, 800, 700)
	writePidMap(t, state, "grok", 800, now.Add(-time.Hour), pidMapSession{ID: "grok-near", Started: now.Add(-time.Minute)})
	writePidMap(t, state, "codex", 700, now.Add(-time.Hour), pidMapSession{ID: "codex-far", Started: now.Add(-time.Minute)})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.Host != "grok" || id.NativeSessionID != "grok-near" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityIgnoresEndedSessions(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	ended := now.Add(-time.Second)
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "codex", 800, now.Add(-time.Hour),
		pidMapSession{ID: "old", Started: now.Add(-30 * time.Minute), Ended: &ended},
		pidMapSession{ID: "current", Started: now.Add(-time.Minute)})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.NativeSessionID != "current" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityRejectsTwoLiveSessionsUnderOneProcess(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "codex", 800, now.Add(-time.Hour),
		pidMapSession{ID: "main", Started: now.Add(-2 * time.Minute)},
		pidMapSession{ID: "subagent", Started: now.Add(-time.Minute)})
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
		t.Fatalf("two live sessions must be rejected as ambiguous, got %v", err)
	}
}

func TestResolveViewIdentityRejectsMapOlderThanSessionChange(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	// The process changed session (/clear) after this entry was recorded.
	writePidMap(t, state, "claude-code", 800, now.Add(-time.Second),
		pidMapSession{ID: "before-clear", Started: now.Add(-time.Hour)})
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), state); !errors.Is(err, errViewIdentityAmbiguous) {
		t.Fatalf("a map older than the last session change must be rejected, got %v", err)
	}
}

func TestResolveViewIdentityRejectsUnreadableOrForgedMaps(t *testing.T) {
	now := time.Now()
	sessions := func(state string) string { return filepath.Join(state, "hooks", "sessions") }
	for name, prepare := range map[string]func(t *testing.T, state string){
		"invalid JSON": func(t *testing.T, state string) {
			os.MkdirAll(sessions(state), 0700)
			os.WriteFile(filepath.Join(sessions(state), "800.json"), []byte("{"), 0600)
		},
		"pid mismatch": func(t *testing.T, state string) {
			p := writePidMap(t, state, "codex", 801, now.Add(-time.Hour), pidMapSession{ID: "x", Started: now})
			os.Rename(p, filepath.Join(sessions(state), "800.json"))
		},
		"unknown host": func(t *testing.T, state string) {
			writePidMap(t, state, "emacs", 800, now.Add(-time.Hour), pidMapSession{ID: "x", Started: now})
		},
		"invalid session id": func(t *testing.T, state string) {
			writePidMap(t, state, "codex", 800, now.Add(-time.Hour), pidMapSession{ID: "", Started: now})
		},
		"symlink": func(t *testing.T, state string) {
			real := writePidMap(t, state, "codex", 799, now.Add(-time.Hour), pidMapSession{ID: "x", Started: now})
			os.Symlink(real, filepath.Join(sessions(state), "800.json"))
		},
	} {
		t.Run(name, func(t *testing.T) {
			state := t.TempDir()
			withIdentityDeps(t, nil, 1000, 800)
			prepare(t, state)
			if id, err := resolveViewIdentity(viewIdentityDepsFn(), state); err == nil {
				t.Fatalf("resolved %+v from an unusable map", id)
			}
		})
	}
}

func TestResolveViewIdentityWithoutSourceIsMissing(t *testing.T) {
	withIdentityDeps(t, nil, 1000, 800)
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir()); !errors.Is(err, errViewIdentityMissing) {
		t.Fatalf("got %v", err)
	}
}

func TestResolveViewIdentityVariableWinsOverMap(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "from-env"}, 1000, 800)
	writePidMap(t, state, "claude-code", 800, time.Now().Add(-time.Hour), pidMapSession{ID: "from-map", Started: time.Now()})
	id, err := resolveViewIdentity(viewIdentityDepsFn(), state)
	if err != nil || id.NativeSessionID != "from-env" {
		t.Fatalf("got %+v, %v", id, err)
	}
}

func TestResolveViewIdentityStopsAtAncestryLimit(t *testing.T) {
	// A parent cycle must not loop forever.
	withIdentityDeps(t, nil, 1000, 900, 1000)
	if _, err := resolveViewIdentity(viewIdentityDepsFn(), t.TempDir()); !errors.Is(err, errViewIdentityMissing) {
		t.Fatalf("got %v", err)
	}
}

// --- parse: identity sources and precedence ---

func TestParseViewWorkCommandResolvesIdentityWithoutFlags(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "resolved-session"})
	c, err := parseViewWorkCommand([]string{"stage", "--state-dir", t.TempDir(), "--work-id", workA, "--stage", "build", "--state", "in_progress", "--reason", "Start build"}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if c.Host != "claude-code" || c.NativeSessionID != "resolved-session" {
		t.Fatalf("identity not resolved: %q %q", c.Host, c.NativeSessionID)
	}
}

func TestParseViewWorkCommandExplicitFlagsWinOverEveryOtherSource(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "from-env"}, 1000, 800)
	writePidMap(t, state, "grok", 800, time.Now().Add(-time.Hour), pidMapSession{ID: "from-map", Started: time.Now()})
	c, err := parseViewWorkCommand([]string{"stage", "--state-dir", state, "--work-id", workA, "--host", "codex", "--native-session-id", "explicit", "--stage", "build", "--state", "in_progress", "--reason", "Start build"}, strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	if c.Host != "codex" || c.NativeSessionID != "explicit" {
		t.Fatalf("flags lost: %q %q", c.Host, c.NativeSessionID)
	}
}

func TestParseViewWorkCommandWithoutSourceStaysRejected(t *testing.T) {
	_, err := parseViewWorkCommand([]string{"stage", "--state-dir", t.TempDir(), "--work-id", workA, "--stage", "build", "--state", "in_progress", "--reason", "Start build"}, strings.NewReader(""))
	var we *viewWorkError
	if !errors.As(err, &we) || we.Code != "invalid_input" || we.Field != "required semantic field missing" {
		t.Fatalf("no source must be rejected as today, got %v", err)
	}
}

func TestParseViewWorkCommandRejectsAmbiguousMapAndLeavesNoState(t *testing.T) {
	state := t.TempDir()
	now := time.Now()
	withIdentityDeps(t, nil, 1000, 800)
	writePidMap(t, state, "codex", 800, now.Add(-time.Hour),
		pidMapSession{ID: "a", Started: now.Add(-2 * time.Minute)},
		pidMapSession{ID: "b", Started: now.Add(-time.Minute)})
	_, err := parseViewWorkCommand([]string{"close", "--state-dir", state, "--work-id", workA, "--outcome", "completed", "--reason", "Done"}, strings.NewReader(""))
	var we *viewWorkError
	if !errors.As(err, &we) || we.Code != "invalid_input" || !strings.Contains(we.Field, "ambiguous") {
		t.Fatalf("got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(state, "publisher")); !os.IsNotExist(statErr) {
		t.Fatal("a rejected command wrote publisher state")
	}
}

func TestParseViewWorkCommandPartialIdentityFlagsStayRejected(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "from-env"})
	for _, flags := range [][]string{{"--host", "codex"}, {"--native-session-id", "only-id"}} {
		args := append([]string{"stage", "--state-dir", t.TempDir(), "--work-id", workA, "--stage", "build", "--state", "in_progress", "--reason", "Start build"}, flags...)
		if _, err := parseViewWorkCommand(args, strings.NewReader("")); err == nil {
			t.Fatalf("%v: a half-given identity must not be completed from another source", flags)
		}
	}
}

func TestParseViewWorkCommandListNeedsNoIdentity(t *testing.T) {
	root := t.TempDir()
	c, err := parseViewWorkCommand([]string{"list", "--state-dir", t.TempDir(), "--project", root}, strings.NewReader(""))
	if err != nil || c.Host != "" || c.NativeSessionID != "" {
		t.Fatalf("list must keep working without identity: %+v, %v", c, err)
	}
}

func TestParseViewWorkCommandLocalActionsTakeNoIdentity(t *testing.T) {
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "from-env"})
	c, err := parseViewWorkCommand([]string{"pin", "--state-dir", t.TempDir(), "--work-id", workA}, strings.NewReader(""))
	if err != nil || c.Host != "" || c.NativeSessionID != "" {
		t.Fatalf("management commands act as the viewer: %+v, %v", c, err)
	}
}

// --- current work ---

func TestCurrentWorkFileRoundTrip(t *testing.T) {
	state := t.TempDir()
	if ids, err := readCurrentWorks(state, "codex", "s1"); err != nil || len(ids) != 0 {
		t.Fatalf("missing file must mean no current work: %v %v", ids, err)
	}
	if err := addCurrentWork(state, "codex", "s1", workA); err != nil {
		t.Fatal(err)
	}
	if err := addCurrentWork(state, "codex", "s1", workA); err != nil {
		t.Fatal(err)
	}
	if err := addCurrentWork(state, "codex", "s1", workB); err != nil {
		t.Fatal(err)
	}
	if ids, _ := readCurrentWorks(state, "codex", "s1"); strings.Join(ids, ",") != workA+","+workB {
		t.Fatalf("got %v", ids)
	}
	if ids, _ := readCurrentWorks(state, "codex", "other"); len(ids) != 0 {
		t.Fatal("current work leaked to another session")
	}
	if err := releaseCurrentWork(state, "codex", "s1", workA); err != nil {
		t.Fatal(err)
	}
	if ids, _ := readCurrentWorks(state, "codex", "s1"); strings.Join(ids, ",") != workB {
		t.Fatalf("got %v", ids)
	}
	info, err := os.Stat(filepath.Join(state, "publisher", "v1", "current", "codex", "s1.json"))
	if err != nil || info.Mode().Perm()&0077 != 0 {
		t.Fatalf("current file must exist and be private: %v", err)
	}
}

func TestCurrentWorkNeverLetsANativeIDReachThePath(t *testing.T) {
	state := t.TempDir()
	for _, native := range []string{"../escape", "a/b", "Mixed.Case:id", ".hidden"} {
		if err := addCurrentWork(state, "codex", native, workA); err != nil {
			t.Fatalf("%q: %v", native, err)
		}
		if ids, err := readCurrentWorks(state, "codex", native); err != nil || len(ids) != 1 {
			t.Fatalf("%q did not round-trip: %v %v", native, ids, err)
		}
	}
	// Everything stayed inside current/codex, except the session locks, which are
	// named by a hash of the identity.
	var outside []string
	locks := filepath.Join(state, "publisher", "v1", "locks") + string(filepath.Separator)
	filepath.Walk(state, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.HasPrefix(p, locks+"session-") && !strings.HasPrefix(p, filepath.Join(state, "publisher", "v1", "current", "codex")+string(filepath.Separator)) {
			outside = append(outside, p)
		}
		return nil
	})
	if len(outside) != 0 {
		t.Fatalf("files outside the host folder: %v", outside)
	}
	if err := addCurrentWork(state, "emacs", "s1", workA); err == nil {
		t.Fatal("an unknown host must be rejected")
	}
}

func TestParseViewWorkCommandUsesTheCurrentWork(t *testing.T) {
	state := t.TempDir()
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "sess-a"})
	args := []string{"stage", "--state-dir", state, "--stage", "build", "--state", "in_progress", "--reason", "Start build"}
	if _, err := parseViewWorkCommand(args, strings.NewReader("")); err == nil {
		t.Fatal("no current work and no --work-id must be rejected as today")
	}
	if err := addCurrentWork(state, "claude-code", "sess-a", workA); err != nil {
		t.Fatal(err)
	}
	c, err := parseViewWorkCommand(args, strings.NewReader(""))
	if err != nil || c.WorkID != workA {
		t.Fatalf("got %q, %v", c.WorkID, err)
	}
	if err = addCurrentWork(state, "claude-code", "sess-a", workB); err != nil {
		t.Fatal(err)
	}
	_, err = parseViewWorkCommand(args, strings.NewReader(""))
	var we *viewWorkError
	if !errors.As(err, &we) || !strings.Contains(we.Field, "work-id") {
		t.Fatalf("two current works must require --work-id, got %v", err)
	}
	c, err = parseViewWorkCommand(append(args, "--work-id", workB), strings.NewReader(""))
	if err != nil || c.WorkID != workB {
		t.Fatalf("explicit --work-id must win: %q %v", c.WorkID, err)
	}
}

func TestParseViewWorkCommandKeepsCreationWithoutWorkID(t *testing.T) {
	state := t.TempDir()
	root := t.TempDir()
	withIdentityDeps(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "sess-a"})
	if err := addCurrentWork(state, "claude-code", "sess-a", workA); err != nil {
		t.Fatal(err)
	}
	c, err := parseViewWorkCommand([]string{"open", "--state-dir", state, "--project", root, "--change", root, "--key", "K", "--title", "T", "--summary", "S"}, strings.NewReader(""))
	if err != nil || c.WorkID != "" {
		t.Fatalf("open must create, never reuse the current work: %q %v", c.WorkID, err)
	}
}

// --- engine: current work set by open, released by close; project registered ---

func TestWorkOpenFixesCurrentWorkAndCloseReleasesIt(t *testing.T) {
	f, c := workEngineFixture(t)
	opened := engineWorkOpen(t, c)
	ids, err := readCurrentWorks(f.stateDir, c.Host, c.NativeSessionID)
	if err != nil || len(ids) != 1 || ids[0] != opened.WorkID {
		t.Fatalf("open must fix the current work: %v %v", ids, err)
	}
	engineWorkOpen(t, c) // reopening the same work stays one entry
	if ids, _ = readCurrentWorks(f.stateDir, c.Host, c.NativeSessionID); len(ids) != 1 {
		t.Fatalf("reopen duplicated the current work: %v", ids)
	}
	closeCmd := c
	closeCmd.Action, closeCmd.WorkID, closeCmd.Outcome, closeCmd.Reason = "close", opened.WorkID, "abandoned", "Closed by test"
	if err = executeViewWork(closeCmd, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if ids, _ = readCurrentWorks(f.stateDir, c.Host, c.NativeSessionID); len(ids) != 0 {
		t.Fatalf("close must release the current work: %v", ids)
	}
}

func TestWorkOpenRegistersTheProjectRoot(t *testing.T) {
	f := newViewFixture(t)
	root, err := canonicalWorkDir(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err = exec.Command("git", "-C", root, "init", "-q").Run(); err != nil {
		t.Fatal(err)
	}
	f.mustView("start") // the root is deliberately not registered
	change := filepath.Join(root, "A")
	os.Mkdir(change, 0700)
	c := viewWorkCommand{Action: "open", Options: management.Options{Scope: "user", StateDir: f.stateDir}, Project: root, Change: change, Host: "codex", NativeSessionID: "register-fixture", Key: "A", Title: "Work A", Summary: "Registers its project"}
	engineWorkOpen(t, c)
	reg, err := viewer.ReadRegistry(f.stateDir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, r := range reg.Repositories {
		found = found || r.Root == root
	}
	if !found {
		t.Fatalf("open did not register %s: %+v", root, reg.Repositories)
	}
}
