package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The fixtures in testdata/time are minimal transcripts in the layouts the
// readers follow; their expected sums are worked by hand in the comments.

func runTimeJSON(t *testing.T, args ...string) timeResult {
	t.Helper()
	var buf bytes.Buffer
	if err := runTime(args, &buf); err != nil {
		t.Fatalf("hive time %v: %v", args, err)
	}
	var r timeResult
	if err := json.Unmarshal(buf.Bytes(), &r); err != nil {
		t.Fatalf("output is not JSON: %q", buf.String())
	}
	return r
}

func TestTimeFixtures(t *testing.T) {
	// claude.jsonl: window 10:00:00 to 10:11:30 = 690 s. The question card
	// (10:06:30 to 10:07:30) is 60 s and the person's message after the turn
	// that ended at 10:08:00 is 120 s: 180 s of user wait. The background
	// `gh pr checks --watch` launched at 10:00:10 and notified at 10:06:20 is
	// 370 s of CI. The sidechain human entry, the task notification, the away
	// summary and the compaction summary end no wait. Active 690 - 180 = 510;
	// converted (510 - 370) * 2 + 370 = 650 s, 660 s rounded up.
	//
	// codex.jsonl: window 10:00:00 to 10:09:30 = 570 s. User waits: the card
	// 30 s, the turn gap 10:06:20 to 10:07:20 60 s, and the user_message after
	// 10:08:00 60 s (the injected instructions, the skill text, the mid-turn
	// steering message and the duplicate of the message end none): 150 s. CI:
	// the exec 1 s, the write_stdin poll 1 s and the wait on its cell 300 s:
	// 302 s. Active 420; converted (420 - 302) * 2 + 302 = 538, 540 s.
	//
	// grok.jsonl: window 490 s. User waits: the card 30 s and the gap 420 to 480
	// 60 s (the synthetic prompt at 410 is the machine's): 90 s. CI: the
	// backgrounded `gh run watch` 40 to 341 = 301 s. Active 400; converted
	// (400 - 301) * 2 + 301 = 499, 540 s.
	//
	// pi.jsonl: window 410 s, a 60 s gap, 300 s of CI. Active 350; converted
	// (350 - 300) * 2 + 300 = 400, 420 s. Always partial.
	cases := []struct {
		file, host, result          string
		active, user, ci, converted int64
		since, until                string
	}{
		{"claude.jsonl", "claude-code", "ok", 510, 180, 370, 660, "2026-01-01T10:00:00Z", "2026-01-01T10:11:30Z"},
		{"codex.jsonl", "codex", "ok", 420, 150, 302, 540, "2026-01-01T10:00:00Z", "2026-01-01T10:09:30Z"},
		{"grok.jsonl", "grok", "ok", 400, 90, 301, 540, "2026-01-01T10:00:00Z", "2026-01-01T10:08:10Z"},
		{"pi.jsonl", "pi", "partial", 350, 60, 300, 420, "2026-01-01T10:00:00Z", "2026-01-01T10:06:50Z"},
	}
	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			r := runTimeJSON(t, "--session-file", filepath.Join("testdata", "time", c.file), "--host", c.host)
			if r.Result != c.result || r.Host != c.host {
				t.Fatalf("result %q host %q, want %q %q (%s)", r.Result, r.Host, c.result, c.host, r.Note)
			}
			if r.ActiveSeconds != c.active || r.UserWaitSeconds != c.user || r.CIWaitSeconds != c.ci || r.ConvertedSeconds != c.converted {
				t.Fatalf("active %d user %d ci %d converted %d, want %d %d %d %d", r.ActiveSeconds, r.UserWaitSeconds, r.CIWaitSeconds, r.ConvertedSeconds, c.active, c.user, c.ci, c.converted)
			}
			if r.Factor != 2 {
				t.Fatalf("factor %v, want 2", r.Factor)
			}
			if c.host != "grok" && (r.Since != c.since || r.Until != c.until) {
				t.Fatalf("window %s to %s, want %s to %s", r.Since, r.Until, c.since, c.until)
			}
		})
	}
}

func TestTimeGrokWindowStartsAtFirstEvent(t *testing.T) {
	r := runTimeJSON(t, "--session-file", filepath.Join("testdata", "time", "grok.jsonl"), "--host", "grok")
	first, err := time.Parse(time.RFC3339, r.Since)
	if err != nil || first.Unix() != 1767261600 {
		t.Fatalf("since %q", r.Since)
	}
}

func TestTimeTruncatedFileIsPartial(t *testing.T) {
	r := runTimeJSON(t, "--session-file", filepath.Join("testdata", "time", "claude-truncated.jsonl"), "--host", "claude-code")
	if r.Result != "partial" || r.Note == "" {
		t.Fatalf("result %q note %q, want partial with a note", r.Result, r.Note)
	}
}

func TestTimeFileWithoutTimedEventsIsPartial(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rollout-x-01a00000-0000-7000-8000-000000000001.jsonl")
	if err := os.WriteFile(path, []byte("{\"type\":\"session_meta\",\"payload\":{}}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := runTimeJSON(t, "--session-file", path, "--host", "codex")
	if r.Result != "partial" || r.ActiveSeconds != 0 {
		t.Fatalf("result %q active %d, want partial", r.Result, r.ActiveSeconds)
	}
}

func TestTimeCodexPastedMarkupEndsTheUserWait(t *testing.T) {
	// The turn ends at 10:01:00 and the person pastes an HTML page at 10:31:00:
	// 1800 s of user wait, not active time.
	lines := []string{
		`{"timestamp":"2026-01-01T10:00:00Z","type":"event_msg","payload":{"type":"task_started"}}`,
		`{"timestamp":"2026-01-01T10:00:30Z","type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":"Done."}]}}`,
		`{"timestamp":"2026-01-01T10:01:00Z","type":"event_msg","payload":{"type":"task_complete"}}`,
		`{"timestamp":"2026-01-01T10:31:00Z","type":"response_item","payload":{"type":"message","role":"user","content":[{"type":"input_text","text":"<!DOCTYPE html><p>error</p>"}]}}`,
		`{"timestamp":"2026-01-01T10:31:05Z","type":"event_msg","payload":{"type":"task_started"}}`,
		`{"timestamp":"2026-01-01T10:32:00Z","type":"event_msg","payload":{"type":"task_complete"}}`,
	}
	path := filepath.Join(t.TempDir(), "rollout-x-01a00000-0000-7000-8000-000000000002.jsonl")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	r := runTimeJSON(t, "--session-file", path, "--host", "codex")
	if r.UserWaitSeconds != 1800 || r.ActiveSeconds != 120 {
		t.Fatalf("user wait %d active %d, want 1800 and 120", r.UserWaitSeconds, r.ActiveSeconds)
	}
}

func TestTimeUnreadableLineIsPartial(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "time", "codex.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "rollout-x-01a00000-0000-7000-8000-000000000000.jsonl")
	broken := append([]byte("{not json\n"), data...)
	if err := os.WriteFile(path, broken, 0o600); err != nil {
		t.Fatal(err)
	}
	r := runTimeJSON(t, "--session-file", path, "--host", "codex")
	if r.Result != "partial" {
		t.Fatalf("result %q, want partial", r.Result)
	}
	if r.Session != "01a00000-0000-7000-8000-000000000000" {
		t.Fatalf("session %q", r.Session)
	}
}

func TestTimeSinceClipsTheWindow(t *testing.T) {
	// From 10:05:00: window 390 s; waits 60 + 120 = 180 s; the CI span clipped
	// to 10:05:00 to 10:06:20 is 80 s. Active 210; converted
	// (210 - 80) * 2 + 80 = 340 s, 360 s rounded up.
	r := runTimeJSON(t, "--session-file", filepath.Join("testdata", "time", "claude.jsonl"), "--host", "claude-code", "--since", "2026-01-01T10:05:00Z")
	if r.Since != "2026-01-01T10:05:00Z" || r.ActiveSeconds != 210 || r.UserWaitSeconds != 180 || r.CIWaitSeconds != 80 || r.ConvertedSeconds != 360 {
		t.Fatalf("%+v", r)
	}
}

func TestTimeFactor(t *testing.T) {
	file := filepath.Join("testdata", "time", "claude.jsonl")
	// (510 - 370) * 3 + 370 = 790 s, 840 s rounded up.
	if r := runTimeJSON(t, "--session-file", file, "--host", "claude-code", "--factor", "3"); r.ConvertedSeconds != 840 || r.Factor != 3 {
		t.Fatalf("%+v", r)
	}
	// With factor 1 the converted time is the active time, 510 s, 540 s rounded up.
	if r := runTimeJSON(t, "--session-file", file, "--host", "claude-code", "--factor", "1"); r.ConvertedSeconds != 540 {
		t.Fatalf("%+v", r)
	}
	for _, bad := range []string{"0.9", "3.1", "nan"} {
		err := runTime([]string{"--session-file", file, "--host", "claude-code", "--factor", bad}, &bytes.Buffer{})
		var ee *exitError
		if !errors.As(err, &ee) || ee.code != 2 {
			t.Fatalf("factor %s: error %v, want exit code 2", bad, err)
		}
	}
}

func TestTimeInputErrorsExitTwo(t *testing.T) {
	file := filepath.Join("testdata", "time", "claude.jsonl")
	for _, args := range [][]string{
		{"--session-file", file},
		{"--host", "claude-code"},
		{"--session-file", file, "--host", "opencode"},
		{"--session-file", file, "--host", "claude-code", "--since", "yesterday"},
		{"--bogus"},
	} {
		err := runTime(args, &bytes.Buffer{})
		var ee *exitError
		if !errors.As(err, &ee) || ee.code != 2 {
			t.Fatalf("%v: error %v, want exit code 2", args, err)
		}
	}
}

func fakeIdentity(t *testing.T, env map[string]string) {
	t.Helper()
	old := viewIdentityDepsFn
	viewIdentityDepsFn = func() viewIdentityDeps {
		return viewIdentityDeps{
			getenv:       func(k string) string { return env[k] },
			selfPID:      100,
			parentPID:    func(int) (int, error) { return 0, nil },
			processStart: func(int) (time.Time, error) { return time.Time{}, errors.New("no process") },
			processName:  func(int) (string, error) { return "", errors.New("no process") },
		}
	}
	t.Cleanup(func() { viewIdentityDepsFn = old })
}

func TestTimeCurrentSessionResolvesItsTranscript(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".claude", "projects", "-work-app")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("testdata", "time", "claude.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sess-1.jsonl"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	fakeIdentity(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "sess-1"})
	r := runTimeJSON(t, "--home", home, "--state-dir", filepath.Join(home, "state"))
	if r.Result != "ok" || r.Host != "claude-code" || r.Session != "sess-1" || r.ActiveSeconds != 510 {
		t.Fatalf("%+v", r)
	}
}

func TestTimeSessionIDLocatesAnEarlierTranscript(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, ".codex", "sessions", "2026", "01", "01")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("testdata", "time", "codex.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	id := "01a00000-0000-7000-8000-0000000000aa"
	if err := os.WriteFile(filepath.Join(dir, "rollout-2026-01-01T10-00-00-"+id+".jsonl"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	state := filepath.Join(home, "state")
	r := runTimeJSON(t, "--home", home, "--state-dir", state, "--host", "codex", "--session", id)
	if r.Result != "ok" || r.Session != id || r.ActiveSeconds != 420 || r.UserWaitSeconds != 150 {
		t.Fatalf("%+v", r)
	}
	if r := runTimeJSON(t, "--home", home, "--state-dir", state, "--host", "codex", "--session", "unknown-1"); r.Result != "no_session_file" {
		t.Fatalf("%+v", r)
	}
	for _, args := range [][]string{
		{"--session", id},
		{"--host", "codex", "--session", id, "--session-file", "x"},
		{"--host", "codex", "--session", "../etc"},
	} {
		var ee *exitError
		if err := runTime(append([]string{"--home", home, "--state-dir", state}, args...), &bytes.Buffer{}); !errors.As(err, &ee) || ee.code != 2 {
			t.Fatalf("%v: error %v, want exit code 2", args, err)
		}
	}
}

func TestTimeHostWithoutTranscriptExitsZero(t *testing.T) {
	home := t.TempDir()
	fakeIdentity(t, map[string]string{"OPENCODE_SESSION_ID": "sess-2"})
	r := runTimeJSON(t, "--home", home, "--state-dir", filepath.Join(home, "state"))
	if r.Result != "no_session_file" || r.Host != "opencode" || r.Session != "sess-2" || r.Factor != 2 {
		t.Fatalf("%+v", r)
	}
	// A host Hive reads, whose transcript is not on disk, gives the same result.
	fakeIdentity(t, map[string]string{"CLAUDE_CODE_SESSION_ID": "missing-1"})
	if r := runTimeJSON(t, "--home", home, "--state-dir", filepath.Join(home, "state")); r.Result != "no_session_file" {
		t.Fatalf("%+v", r)
	}
}

func TestConvertTimeRoundsUpToTheMinute(t *testing.T) {
	for _, c := range []struct {
		active, ci int64
		factor     float64
		want       int64
	}{
		{0, 0, 2, 0},
		{1, 0, 2, 60},
		{30, 0, 2, 60},
		{31, 0, 2, 120},
		{3600, 0, 2, 7200},
		{3600, 3600, 3, 3600},
		{100, 40, 1.5, 180},
	} {
		if got := convertTime(c.active, c.ci, c.factor); got != c.want {
			t.Errorf("convertTime(%d, %d, %v) = %d, want %d", c.active, c.ci, c.factor, got, c.want)
		}
	}
}
