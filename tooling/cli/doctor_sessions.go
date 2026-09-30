// doctor_sessions.go is the Sessions section: a one-line summary, scoped to the
// hosts Hive checked, of how many open sessions of Claude Code and Grok started
// before the installed release and should be restarted; one line naming the
// registered CLIs whose sessions Hive cannot see; then each checked CLI's own
// sessions. The session files are internal,
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
	"strings"
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

// hostSessions is what checking one CLI (claude or grok) found.
type hostSessions struct {
	host  string
	lines []string
	live  int // open sessions whose process is alive
	stale int // of those, the ones that started before the installed release
	// checked is false when the sessions could not be read, or could be read but
	// not compared with the release time; the summary then names the host.
	checked bool
	// compared is true when live sessions were compared with the release time,
	// which makes the marks estimates the section explains.
	compared bool
}

// hostDisplayNames are the names the summary uses for the CLIs it checks and
// for the ones it cannot see.
var hostDisplayNames = map[string]string{
	"claude": "Claude Code", "grok": "Grok",
	"codex": "Codex", "pi": "Pi", "cursor": "Cursor",
}

// unseenHosts are the CLIs whose open sessions Hive cannot read, in the order
// the notice names them.
var unseenHosts = []string{"codex", "pi", "cursor"}

// listWords joins names as "A", "A or B" or "A, B or C" (conjunction "or" or
// "and").
func listWords(names []string, conjunction string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " " + conjunction + " " + names[len(names)-1]
}

func sessionsSection(deps doctorDeps, st doctorState) doctorSection {
	sec := doctorSection{Title: "Sessions"}
	if st.err != nil {
		// The CLIs section already explains the unreadable state.
		sec.Err = stateNotChecked
		return sec
	}
	if len(st.registered) == 0 {
		sec.Lines = []string{noHostsText}
		return sec
	}
	home, err := deps.userHome()
	if err != nil {
		sec.Err = "cannot resolve the home directory: " + sanitizeLine(err.Error())
	}
	var (
		checks   []hostSessions
		lines    []string
		estimate bool
	)
	for _, host := range installerHosts {
		if !st.isRegistered(host) {
			continue
		}
		switch host {
		case "claude", "grok":
			if err != nil {
				checks = append(checks, hostSessions{host: host})
				lines = append(lines, fmt.Sprintf("Session check unavailable for %s: no home directory", host))
				continue
			}
			ref, hasRef := st.released[st.hostInstallation(host).Release]
			res := checkHostSessions(deps, host, home, ref, hasRef)
			checks = append(checks, res)
			estimate = estimate || res.compared
			lines = append(lines, res.lines...)
		case "opencode":
			lines = append(lines, "opencode: reloads its instructions on the next message; start a new session for role or skill changes")
		}
	}
	if summary := sessionsSummary(checks); summary != "" {
		sec.Lines = append(sec.Lines, summary)
	}
	var unseen []string
	for _, host := range unseenHosts {
		if st.isRegistered(host) {
			unseen = append(unseen, hostDisplayNames[host])
		}
	}
	if len(unseen) > 0 {
		sec.Lines = append(sec.Lines, "Hive cannot see "+listWords(unseen, "or")+" sessions; restart them after each update.")
	}
	sec.Lines = append(sec.Lines, lines...)
	if estimate {
		sec.Lines = append(sec.Lines, "Restart marks are estimates: they compare each start time with when the installed release was last written.")
	}
	return sec
}

// sessionsSummary answers "do I need to restart sessions?" in one line, for the
// CLIs whose sessions Hive can read (Claude Code and Grok) and names only the
// ones it could check, so it is not read as covering CLIs Hive cannot see. A
// CLI it could not check follows after a semicolon. It is empty when none of
// them is registered, because then nothing was checked.
func sessionsSummary(checks []hostSessions) string {
	if len(checks) == 0 {
		return ""
	}
	stale := 0
	var checked, unavailable []string
	for _, c := range checks {
		stale += c.stale
		if c.checked {
			checked = append(checked, hostDisplayNames[c.host])
		} else {
			unavailable = append(unavailable, hostDisplayNames[c.host])
		}
	}
	note := ""
	if len(unavailable) > 0 {
		note = listWords(unavailable, "and") + " could not be checked"
	}
	if len(checked) == 0 {
		return "Hive could not check whether any open session needs a restart (" + note + ")"
	}
	hosts := listWords(checked, "or")
	var line string
	if stale > 0 {
		line = fmt.Sprintf("%d open %s %s should be restarted", stale, hosts, plural(stale, "session", "sessions"))
	} else {
		line = "No open " + hosts + " session needs a restart"
	}
	if note != "" {
		line += "; " + note + "."
	}
	return line
}

// checkHostSessions checks one CLI (claude or grok) and returns its lines and
// counts.
func checkHostSessions(deps doctorDeps, host, home string, ref time.Time, hasRef bool) hostSessions {
	res := hostSessions{host: host}
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
		res.lines = []string{fmt.Sprintf("Session check unavailable for %s: %s", host, sanitizeLine(err.Error()))}
		return res
	}
	var live []openSession
	for _, s := range sessions {
		if s.PID > 0 && deps.processAlive(s.PID) {
			live = append(live, s)
		}
	}
	sort.Slice(live, func(i, j int) bool { return live[i].PID < live[j].PID })
	res.live = len(live)
	if len(live) == 0 {
		res.checked = true
		res.lines = []string{host + ": no open sessions"}
		return res
	}
	res.compared = hasRef
	res.checked = hasRef
	for _, s := range live {
		if hasRef && s.Started.Before(ref) {
			res.stale++
		}
	}
	count := fmt.Sprintf("%s: %d open %s", host, len(live), plural(len(live), "session", "sessions"))
	switch {
	case !hasRef:
	case res.stale == 0:
		count += ", none to restart"
	default:
		count += fmt.Sprintf(", %d to restart", res.stale)
	}
	res.lines = []string{count}
	if !hasRef {
		res.lines = append(res.lines, "  The installed release's write time is unknown, so sessions are not compared with it.")
	}
	now := deps.now()
	for _, s := range live {
		res.lines = append(res.lines,
			fmt.Sprintf("  pid %d  started %s (%s)", s.PID, s.Started.UTC().Format("2006-01-02 15:04 UTC"), age(now.Sub(s.Started))),
			"    "+sanitizeLine(s.CWD))
		switch {
		case !hasRef:
		case s.Started.Before(ref):
			res.lines = append(res.lines, "    "+staleSessionMark)
		default:
			res.lines = append(res.lines, "    up to date")
		}
	}
	return res
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
