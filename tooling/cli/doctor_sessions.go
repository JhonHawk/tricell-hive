// doctor_sessions.go is the Sessions section: which open sessions of Claude
// Code and Grok started before the installed release, and what to do for the
// CLIs whose sessions Hive cannot see. The session files are internal,
// undocumented formats of those CLIs: they are read defensively (size limit,
// tolerant decoding) and only pid, working directory and start time are shown.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	sessionFileLimit = 1 << 20 // bytes read from any one session file
	sessionFileCap   = 500     // most session files read for one CLI
)

// staleSessionMark is shown for a session that started before the release
// the CLI has installed.
const staleSessionMark = "started before the installed release; restart it"

// openSession is what Hive shows about one live session.
type openSession struct {
	PID     int
	CWD     string
	Started time.Time
}

func sessionsSection(deps doctorDeps, st doctorState) doctorSection {
	sec := doctorSection{Title: "Sessions"}
	if st.err != nil {
		sec.Err = "cannot read the installation state: " + sanitizeLine(st.err.Error())
		return sec
	}
	if len(st.registered) == 0 {
		sec.Lines = []string{"No CLI hosts are registered"}
		return sec
	}
	home, err := deps.userHome()
	if err != nil {
		sec.Err = "cannot resolve the home directory: " + sanitizeLine(err.Error())
	}
	estimate := false
	for _, host := range installerHosts {
		if !st.isRegistered(host) {
			continue
		}
		switch host {
		case "claude", "grok":
			if err != nil {
				sec.Lines = append(sec.Lines, fmt.Sprintf("Session check unavailable for %s: no home directory", host))
				continue
			}
			ref, hasRef := st.released[st.hostInstallation(host).Release]
			lines, compared := hostSessionLines(deps, host, home, ref, hasRef)
			estimate = estimate || compared
			sec.Lines = append(sec.Lines, lines...)
		case "opencode":
			sec.Lines = append(sec.Lines, "opencode: reloads its instructions on the next message; no restart needed")
		default:
			sec.Lines = append(sec.Lines, host+": Hive cannot see its open sessions; restart them after each update")
		}
	}
	if estimate {
		sec.Lines = append(sec.Lines, "Restart marks are estimates: they compare each start time with when the installed release was last written.")
	}
	return sec
}

// hostSessionLines checks one CLI (claude or grok) and returns its lines, and
// whether it compared live sessions with the release time (which makes the
// marks estimates the section explains).
func hostSessionLines(deps doctorDeps, host, home string, ref time.Time, hasRef bool) (lines []string, compared bool) {
	var (
		sessions []openSession
		err      error
	)
	switch host {
	case "claude":
		sessions, err = readClaudeSessions(deps, home)
	default:
		sessions, err = readGrokSessions(deps, home)
	}
	if err != nil {
		return []string{fmt.Sprintf("Session check unavailable for %s: %s", host, sanitizeLine(err.Error()))}, false
	}
	var live []openSession
	for _, s := range sessions {
		if s.PID > 0 && deps.processAlive(s.PID) {
			live = append(live, s)
		}
	}
	sort.Slice(live, func(i, j int) bool { return live[i].PID < live[j].PID })
	if len(live) == 0 {
		return []string{host + ": no open sessions"}, false
	}
	lines = []string{fmt.Sprintf("%s: %d open %s", host, len(live), plural(len(live), "session", "sessions"))}
	if !hasRef {
		lines = append(lines, "  The installed release's write time is unknown, so sessions are not compared with it.")
	}
	now := deps.now()
	for _, s := range live {
		lines = append(lines,
			fmt.Sprintf("  pid %d  started %s (%s)", s.PID, s.Started.UTC().Format("2006-01-02 15:04 UTC"), age(now.Sub(s.Started))),
			"    "+sanitizeLine(s.CWD))
		switch {
		case !hasRef:
		case s.Started.Before(ref):
			lines = append(lines, "    "+staleSessionMark)
		default:
			lines = append(lines, "    up to date")
		}
	}
	return lines, hasRef
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// age is a coarse "how long ago" for a start time.
func age(d time.Duration) string {
	switch {
	case d < 0:
		return "in the future"
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d/time.Minute))
	case d < 48*time.Hour:
		return fmt.Sprintf("%d h ago", int(d/time.Hour))
	}
	return fmt.Sprintf("%d days ago", int(d/(24*time.Hour)))
}

// readSessionFile reads a regular file of at most sessionFileLimit bytes.
func readSessionFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", filepath.Base(path))
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, sessionFileLimit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > sessionFileLimit {
		return nil, fmt.Errorf("%s is larger than 1 MiB", filepath.Base(path))
	}
	return data, nil
}

// configDir is the CLI's own directory: its environment variable, else the
// default under home. An unset variable comes back empty from deps.getenv.
func configDir(deps doctorDeps, envName, home, fallback string) string {
	if v := deps.getenv(envName); v != "" {
		return v
	}
	return filepath.Join(home, fallback)
}

var claudeSessionFile = regexp.MustCompile(`^[0-9]+\.json$`)

// readClaudeSessions reads <config dir>/sessions/<pid>.json. Each file holds
// pid, cwd and startedAt (milliseconds since the Unix epoch); other keys are
// ignored. A missing directory means no sessions.
func readClaudeSessions(deps doctorDeps, home string) ([]openSession, error) {
	dir := filepath.Join(configDir(deps, "CLAUDE_CONFIG_DIR", home, ".claude"), "sessions")
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && claudeSessionFile.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	if len(names) > sessionFileCap {
		return nil, fmt.Errorf("more than %d session files", sessionFileCap)
	}
	var out []openSession
	for _, name := range names {
		data, err := readSessionFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		var raw struct {
			PID       *float64 `json:"pid"`
			CWD       string   `json:"cwd"`
			StartedAt *float64 `json:"startedAt"`
		}
		if err := json.Unmarshal(data, &raw); err != nil || raw.PID == nil || raw.StartedAt == nil || *raw.StartedAt <= 0 {
			return nil, fmt.Errorf("%s has an unknown format", name)
		}
		out = append(out, openSession{PID: wholePID(*raw.PID), CWD: raw.CWD, Started: time.UnixMilli(int64(*raw.StartedAt))})
	}
	return out, nil
}

// readGrokSessions reads <grok home>/active_sessions.json, a list of
// {session_id, pid, cwd, opened_at} where opened_at is an RFC 3339 time. A
// missing file means no sessions.
func readGrokSessions(deps doctorDeps, home string) ([]openSession, error) {
	path := filepath.Join(configDir(deps, "GROK_HOME", home, ".grok"), "active_sessions.json")
	data, err := readSessionFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var raw []struct {
		PID      *float64 `json:"pid"`
		CWD      string   `json:"cwd"`
		OpenedAt *string  `json:"opened_at"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, errors.New("active_sessions.json has an unknown format")
	}
	var out []openSession
	for _, r := range raw {
		if r.PID == nil || r.OpenedAt == nil {
			return nil, errors.New("active_sessions.json has an unknown format")
		}
		opened, err := time.Parse(time.RFC3339Nano, *r.OpenedAt)
		if err != nil {
			return nil, errors.New("active_sessions.json has an unknown format")
		}
		out = append(out, openSession{PID: wholePID(*r.PID), CWD: r.CWD, Started: opened})
	}
	return out, nil
}

// wholePID converts a decoded JSON number to a pid, or 0 (never alive) when it
// is not a whole number in the pid range.
func wholePID(f float64) int {
	if f != float64(int64(f)) || f < 1 || f > 1<<31-1 {
		return 0
	}
	return int(f)
}
