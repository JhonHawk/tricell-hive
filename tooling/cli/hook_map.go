package main

// hook_map.go writes the process-to-session map that view_identity_resolve.go
// reads. The format and every rule here are the «Session map» section of
// _support/docs/architecture/host-session-events.md; a change to one side changes
// the other.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"tricell-hive/integrations/hostevent"
)

// hookProcessArgs returns the command line of a process; tests replace it.
var hookProcessArgs = func(pid int) (string, error) { return psField(pid, "args") }

// hookMapHosts are the hosts that identify their session through the map. The
// OpenCode and Pi hooks name their session themselves, and their processes do
// not tell sessions apart (OpenCode runs every session in one shared service).
func hookUsesMap(host string) bool {
	switch host {
	case "claude-code", "codex", "grok", "cursor":
		return true
	}
	return false
}

// hostProcessMatches says whether a process is the host CLI: its command name
// is the host's, or it is an interpreter running the host's script.
func hostProcessMatches(host, comm, args string) bool {
	return processHost(comm, args) == host
}

// processHost is the host CLI a process is, or "" when it is none.
func processHost(comm, args string) string {
	if h := hostOfProcessName(comm); h != "" {
		return h
	}
	switch filepath.Base(strings.TrimSpace(comm)) {
	case "node", "bun", "deno":
		fields := strings.Fields(args)
		for i := 1; i < len(fields) && i <= 3; i++ {
			name := filepath.Base(fields[i])
			name = strings.TrimSuffix(name, filepath.Ext(name))
			if h := hostOfProcessName(name); h != "" {
				return h
			}
		}
	}
	return ""
}

// findHostProcess walks up from the hook to the nearest ancestor that is the
// host CLI. The map is named after that process, never the hook's own `sh -c`
// wrapper or a short-lived child; a hook that cannot tell finds none.
func findHostProcess(deps viewIdentityDeps, host string) (int, bool) {
	found := 0
	_ = walkViewAncestors(deps, func(pid int) (bool, error) {
		if deps.processName == nil {
			return true, nil
		}
		comm, err := deps.processName(pid)
		if err != nil {
			return false, nil
		}
		args := ""
		if base := filepath.Base(strings.TrimSpace(comm)); base == "node" || base == "bun" || base == "deno" {
			args, _ = hookProcessArgs(pid)
		}
		switch processHost(comm, args) {
		case "":
			return false, nil
		case host:
			found = pid
		}
		// The nearest host CLI is the one that ran the hook: a hook registered
		// for another host and run by this one (Grok and Cursor run Claude
		// Code's hooks) writes no map, and never one on a process further up.
		return true, nil
	})
	return found, found != 0
}

type hookMapSession struct {
	SessionID string `json:"session_id"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type hookMapDoc struct {
	Host             string           `json:"host"`
	PID              int              `json:"pid"`
	ProcessStartedAt string           `json:"process_started_at"`
	ChangedAt        string           `json:"changed_at"`
	Sessions         []hookMapSession `json:"sessions"`
}

func (d *hookMapDoc) index(id string) int {
	for i, s := range d.Sessions {
		if s.SessionID == id {
			return i
		}
	}
	return -1
}

// live: no ended_at and started_at at or after changed_at, the reader's rule.
func (d *hookMapDoc) live(s hookMapSession) bool {
	if s.EndedAt != "" {
		return false
	}
	started, err1 := time.Parse(time.RFC3339Nano, s.StartedAt)
	changed, err2 := time.Parse(time.RFC3339Nano, d.ChangedAt)
	return err1 == nil && err2 == nil && !started.Before(changed)
}

func (d *hookMapDoc) liveCount() int {
	n := 0
	for _, s := range d.Sessions {
		if d.live(s) {
			n++
		}
	}
	return n
}

// apply records one event at time now and reports whether the document changed.
// changed_at moves only when a session replaces another in the process: a
// session that starts with Replaces, or the end of the last live session. A
// session starting alongside the others leaves it, so two live sessions stay
// detectable as ambiguous; ending a session that an earlier replacement had
// already superseded leaves it too, so the order of /clear's end and start
// events does not matter.
func (d *hookMapDoc) apply(ev hostevent.Event, now time.Time) bool {
	if changed, err := time.Parse(time.RFC3339Nano, d.ChangedAt); err == nil && now.Before(changed) {
		now = changed
	}
	ts := now.UTC().Format(time.RFC3339Nano)
	switch ev.Map {
	case hostevent.MapStart, hostevent.MapTouch:
		i := d.index(ev.SessionID)
		if i >= 0 && d.live(d.Sessions[i]) {
			// Already live: a repeated start or a touch moves nothing.
			return false
		}
		if i >= 0 {
			d.Sessions = append(d.Sessions[:i], d.Sessions[i+1:]...)
		}
		if ev.Replaces || len(d.Sessions) == 0 || d.liveCount() == 0 {
			d.ChangedAt = ts
		}
		d.Sessions = append(d.Sessions, hookMapSession{SessionID: ev.SessionID, StartedAt: ts})
		d.trim()
		return true
	case hostevent.MapEnd:
		i := d.index(ev.SessionID)
		if i < 0 || d.Sessions[i].EndedAt != "" {
			return false
		}
		wasLive := d.live(d.Sessions[i])
		d.Sessions[i].EndedAt = ts
		if wasLive && d.liveCount() == 0 {
			d.ChangedAt = ts
		}
		return true
	}
	return false
}

// trim keeps at most viewPidMapSessionCap sessions, dropping the oldest ended
// ones first and then the oldest of all.
func (d *hookMapDoc) trim() {
	for len(d.Sessions) > viewPidMapSessionCap {
		drop := 0
		for i, s := range d.Sessions {
			if s.EndedAt != "" {
				drop = i
				break
			}
		}
		d.Sessions = append(d.Sessions[:drop], d.Sessions[drop+1:]...)
	}
}

// recordSessionMap applies a start, touch or end to the map of the host process
// the hook runs under. It writes nothing when no ancestor is the host, and
// returns the host's start time for use as a stable event datum.
func recordSessionMap(state string, deps viewIdentityDeps, ev hostevent.Event, now time.Time) (time.Time, error) {
	pid, ok := findHostProcess(deps, ev.Host)
	if !ok || deps.processStart == nil {
		return time.Time{}, errViewIdentityMissing
	}
	started, err := deps.processStart(pid)
	if err != nil {
		return time.Time{}, err
	}
	if ev.Map == hostevent.MapNone {
		return started, nil
	}
	dir, err := ensureHookDir(state, "sessions")
	if err != nil {
		return started, err
	}
	l, err := acquirePrivateWorkLock(filepath.Join(dir, strconv.Itoa(pid)+".lock"), true)
	if err != nil {
		return started, err
	}
	defer l.Release()
	path := filepath.Join(dir, strconv.Itoa(pid)+".json")
	doc := &hookMapDoc{Host: ev.Host, PID: pid, ProcessStartedAt: started.UTC().Format(time.RFC3339), Sessions: []hookMapSession{}}
	raw, err := readSafeMapFile(path)
	switch {
	case err == nil:
		var old hookMapDoc
		if json.Unmarshal(raw, &old) == nil && old.Host == ev.Host && old.PID == pid && len(old.Sessions) <= viewPidMapSessionCap {
			if was, perr := time.Parse(time.RFC3339Nano, old.ProcessStartedAt); perr == nil {
				if d := started.Sub(was); d <= viewProcessStartTolerance && d >= -viewProcessStartTolerance {
					doc = &old
				}
			}
		}
	case os.IsNotExist(err):
	default:
		return started, err // a map that exists but cannot be trusted is never replaced
	}
	if !doc.apply(ev, now) {
		return started, nil
	}
	out, err := json.Marshal(doc)
	if err != nil {
		return started, err
	}
	return started, writePrivateWorkFile(dir, path, out)
}

// ensureHookDir creates <state>/hooks/<name> at mode 0700 and applies the
// safe-folder check to it and to hooks/.
func ensureHookDir(state, name string) (string, error) {
	hooks := filepath.Join(state, "hooks")
	dir := filepath.Join(hooks, name)
	for _, d := range []string{hooks, dir} {
		if err := os.Mkdir(d, 0700); err != nil && !os.IsExist(err) {
			return "", err
		}
		if err := checkSafeDir(d); err != nil {
			return "", err
		}
	}
	return dir, nil
}
