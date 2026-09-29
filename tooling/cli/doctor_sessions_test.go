package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tricell-hive/tooling/management"
)

// Tests of the Sessions section (AC4). The release was written at 12:00 UTC;
// the fake clock reads 15:00 UTC the same day.
var (
	sessionRelease = time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	sessionBefore  = time.Date(2026, 9, 29, 10, 0, 0, 0, time.UTC)
	sessionAfter   = time.Date(2026, 9, 29, 13, 0, 0, 0, time.UTC)
)

// sessionsFixture installs the hosts in a synthetic home and fixes the
// release's write time.
func sessionsFixture(t *testing.T, hosts string) (o management.Options, home string, f *doctorFake) {
	t.Helper()
	o, home, stateDir := doctorHome(t, hosts)
	setReleaseTime(t, stateDir, sessionRelease)
	return o, home, newDoctorFake(home)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func claudeSession(t *testing.T, dir, name string, pid int, cwd string, started time.Time) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "sessions", name), fmt.Sprintf(
		`{"pid":%d,"sessionId":"s","cwd":%s,"startedAt":%d,"version":"2.1.284","kind":"interactive","status":"idle","secret":"TOPSECRET"}`,
		pid, jsonString(cwd), started.UnixMilli()))
}

func grokSessions(t *testing.T, dir string, entries ...string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "active_sessions.json"), "["+strings.Join(entries, ",")+"]")
}

func grokEntry(pid int, cwd string, opened time.Time) string {
	return fmt.Sprintf(`{"session_id":"g","pid":%d,"cwd":%s,"opened_at":%s,"extra":{"x":1}}`, pid, jsonString(cwd), jsonString(opened.Format(time.RFC3339Nano)))
}

func jsonString(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// sessionLine returns the block of one session: its "pid" line and the
// indented lines that follow it (working directory and mark).
func sessionLine(t *testing.T, sec doctorSection, needle string) string {
	t.Helper()
	for i, l := range sec.Lines {
		if !strings.Contains(l, needle) {
			continue
		}
		block := []string{l}
		for _, next := range sec.Lines[i+1:] {
			if !strings.HasPrefix(next, "    ") {
				break
			}
			block = append(block, next)
		}
		return strings.Join(block, "\n")
	}
	t.Fatalf("no line with %q in:\n%s", needle, sectionText(sec))
	return ""
}

func TestSessionsClaudeMarksOnlySessionsStartedBeforeTheRelease(t *testing.T) {
	o, home, f := sessionsFixture(t, "claude")
	dir := filepath.Join(home, ".claude")
	claudeSession(t, dir, "101.json", 101, "/work/old", sessionBefore)
	claudeSession(t, dir, "102.json", 102, "/work/new", sessionAfter)
	claudeSession(t, dir, "103.json", 103, "/work/dead", sessionBefore)
	claudeSession(t, dir, "0.json", 0, "/work/zero", sessionBefore)
	claudeSession(t, dir, "104.json", 104, "/work/negative", sessionBefore)
	writeFile(t, filepath.Join(dir, "sessions", "101.abcdef.key"), "KEYMATERIAL")
	writeFile(t, filepath.Join(dir, "sessions", "notes.json"), "not json at all")
	f.alive[101], f.alive[102] = true, true
	f.alive[0] = true // even a fake that says "alive" must not revive pid 0
	f.alive[-1] = true

	sec := collectDoctor(o, t.TempDir(), f.deps()).Sessions
	text := sectionText(sec)
	mustContain(t, text, "claude: 2 open sessions")
	old := sessionLine(t, sec, "pid 101")
	mustContain(t, old, "/work/old", "started 2026-09-29 10:00 UTC", "5 h ago", "started before the installed release; restart it")
	fresh := sessionLine(t, sec, "pid 102")
	mustContain(t, fresh, "/work/new", "started 2026-09-29 13:00 UTC", "up to date")
	mustNotContain(t, fresh, "started before")
	mustNotContain(t, text, "pid 103", "pid 0 ", "/work/zero", "/work/negative", "TOPSECRET", "KEYMATERIAL", "2.1.284")
	mustContain(t, text, "Restart marks are estimates")
}

func TestSessionsClaudeDirectoryFromEnvironmentComesFromDeps(t *testing.T) {
	o, home, _ := sessionsFixture(t, "claude")
	o = nonSyntheticOptions(t, o, home)
	other := filepath.Join(t.TempDir(), "claude-config")
	claudeSession(t, other, "201.json", 201, "/env/dir", sessionBefore)
	claudeSession(t, filepath.Join(home, ".claude"), "202.json", 202, "/default/dir", sessionBefore)
	f := newDoctorFake(home)
	f.env["CLAUDE_CONFIG_DIR"] = other
	f.alive[201], f.alive[202] = true, true

	text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text, "/env/dir")
	mustNotContain(t, text, "/default/dir")

	delete(f.env, "CLAUDE_CONFIG_DIR")
	text = sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text, "/default/dir")
	mustNotContain(t, text, "/env/dir")
}

func TestSessionsSyntheticHomeIgnoresEnvironment(t *testing.T) {
	o, home, f := sessionsFixture(t, "claude,grok")
	f.env["CLAUDE_CONFIG_DIR"] = "/must/not/be/read"
	f.env["GROK_HOME"] = "/must/not/be/read"
	claudeSession(t, filepath.Join(home, ".claude"), "301.json", 301, "/synthetic/claude", sessionBefore)
	grokSessions(t, filepath.Join(home, ".grok"), grokEntry(302, "/synthetic/grok", sessionBefore))
	f.alive[301], f.alive[302] = true, true
	text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text, "/synthetic/claude", "/synthetic/grok")
}

func TestSessionsGrokMarksOnlySessionsOpenedBeforeTheRelease(t *testing.T) {
	o, home, f := sessionsFixture(t, "grok")
	grokSessions(t, filepath.Join(home, ".grok"),
		grokEntry(401, "/g/old", sessionBefore),
		grokEntry(402, "/g/new", sessionAfter.Add(123456789*time.Nanosecond)),
		grokEntry(403, "/g/dead", sessionBefore),
		grokEntry(0, "/g/zero", sessionBefore))
	f.alive[401], f.alive[402], f.alive[0] = true, true, true

	sec := collectDoctor(o, t.TempDir(), f.deps()).Sessions
	text := sectionText(sec)
	mustContain(t, text, "grok: 2 open sessions")
	mustContain(t, sessionLine(t, sec, "pid 401"), "/g/old", "started before the installed release; restart it")
	fresh := sessionLine(t, sec, "pid 402")
	mustContain(t, fresh, "/g/new", "up to date")
	mustNotContain(t, fresh, "started before")
	mustNotContain(t, text, "pid 403", "/g/zero", "session_id")
}

func TestSessionsGrokHomeFromEnvironmentComesFromDeps(t *testing.T) {
	o, home, _ := sessionsFixture(t, "grok")
	o = nonSyntheticOptions(t, o, home)
	other := filepath.Join(t.TempDir(), "grok-home")
	grokSessions(t, other, grokEntry(501, "/grok/env", sessionBefore))
	f := newDoctorFake(home)
	f.env["GROK_HOME"] = other
	f.alive[501] = true
	mustContain(t, sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions), "/grok/env")
}

func TestSessionsNoOpenSessions(t *testing.T) {
	o, home, f := sessionsFixture(t, "claude,grok")
	grokSessions(t, filepath.Join(home, ".grok"))
	f.alive = map[int]bool{}
	text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text, "claude: no open sessions", "grok: no open sessions")
	mustNotContain(t, text, "estimates")
}

func TestSessionsUnavailableWhenGrokFileIsOversizedUnknownOrUnreadable(t *testing.T) {
	for _, tc := range []struct {
		name   string
		write  func(t *testing.T, path string)
		reason string
	}{
		{"oversized", func(t *testing.T, path string) {
			writeFile(t, path, "["+strings.Repeat(" ", 2<<20)+"]")
		}, "larger than 1 MiB"},
		{"not a list", func(t *testing.T, path string) { writeFile(t, path, `{"sessions":[]}`) }, "unknown format"},
		{"missing opened_at", func(t *testing.T, path string) { writeFile(t, path, `[{"pid":5,"cwd":"/x"}]`) }, "unknown format"},
		{"bad date", func(t *testing.T, path string) { writeFile(t, path, `[{"pid":5,"cwd":"/x","opened_at":"yesterday"}]`) }, "unknown format"},
		{"not json", func(t *testing.T, path string) { writeFile(t, path, `garbage`) }, "unknown format"},
		{"unreadable", func(t *testing.T, path string) {
			writeFile(t, path, `[]`)
			if err := os.Chmod(path, 0); err != nil {
				t.Fatal(err)
			}
		}, "permission denied"},
		{"directory", func(t *testing.T, path string) {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
		}, "not a regular file"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "unreadable" && os.Geteuid() == 0 {
				t.Skip("root reads any file")
			}
			o, home, f := sessionsFixture(t, "claude,grok")
			tc.write(t, filepath.Join(home, ".grok", "active_sessions.json"))
			claudeSession(t, filepath.Join(home, ".claude"), "601.json", 601, "/still/shown", sessionBefore)
			f.alive[601] = true
			r := collectDoctor(o, t.TempDir(), f.deps())
			text := sectionText(r.Sessions)
			mustContain(t, text, "Session check unavailable for grok: ", tc.reason)
			mustContain(t, text, "/still/shown") // the other CLI is still checked
			mustContain(t, sectionText(r.CLIs), "claude")
			mustContain(t, sectionText(r.Installation), "No problems found")
		})
	}
}

func TestSessionsUnavailableWhenClaudeFileIsOversizedOrUnknown(t *testing.T) {
	for name, content := range map[string]string{
		"unknown format":    `{"pid":1}`,
		"larger than 1 MiB": `{"pid":1,"startedAt":1,"pad":"` + strings.Repeat("x", 2<<20) + `"}`,
	} {
		t.Run(name, func(t *testing.T) {
			o, home, f := sessionsFixture(t, "claude")
			writeFile(t, filepath.Join(home, ".claude", "sessions", "701.json"), content)
			f.alive[701] = true
			text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
			mustContain(t, text, "Session check unavailable for claude: 701.json", name)
		})
	}
}

func TestSessionsNoticesForHostsHiveCannotSee(t *testing.T) {
	o, _, f := sessionsFixture(t, "codex,pi,cursor,opencode")
	text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text,
		"codex: Hive cannot see its open sessions; restart them after each update",
		"pi: Hive cannot see its open sessions; restart them after each update",
		"cursor: Hive cannot see its open sessions; restart them after each update",
		"opencode: reloads its instructions on the next message; no restart needed")
	mustNotContain(t, text, "claude", "grok", "estimates")
}

func TestSessionsCwdIsSanitized(t *testing.T) {
	o, home, f := sessionsFixture(t, "claude")
	claudeSession(t, filepath.Join(home, ".claude"), "801.json", 801, "/work/\x1b[2Jevil\x07", sessionBefore)
	f.alive[801] = true
	text := sectionText(collectDoctor(o, t.TempDir(), f.deps()).Sessions)
	mustContain(t, text, "/work/evil")
	mustNotContain(t, text, "\x1b", "\x07", "[2J")
}

func TestSessionsWithoutReleaseTimeAreListedButNotMarked(t *testing.T) {
	o, home, f := sessionsFixture(t, "claude")
	claudeSession(t, filepath.Join(home, ".claude"), "901.json", 901, "/w", sessionBefore)
	f.alive[901] = true
	// The file system always gives a write time, so simulate a release whose
	// time could not be read by emptying the map the section consults.
	st := loadDoctorState(o)
	st.released = map[string]time.Time{}
	sec := sessionsSection(f.deps().forOptions(o), st)
	text := sectionText(sec)
	mustContain(t, text, "pid 901", "write time is unknown")
	mustNotContain(t, text, "started before", "up to date")
}
