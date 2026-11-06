package main

// view_identity_resolve.go holds the resolution rules of view_identity.go: the
// host variables, the ancestor walk, the session map and the safe-location
// checks. The map format and the writer's rules are the «Session map» section
// of _support/docs/architecture/host-session-events.md.

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"tricell-hive/tooling/viewer"
)

var viewIdentityDepsFn = func() viewIdentityDeps {
	return viewIdentityDeps{
		getenv:       os.Getenv,
		selfPID:      os.Getpid(),
		parentPID:    psParentPID,
		processStart: psProcessStart,
		processName:  psProcessName,
	}
}

// psField asks ps for one field of pid; it works on macOS and Linux. The C
// locale keeps lstart in one fixed layout.
func psField(pid int, field string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ps", "-o", field+"=", "-p", strconv.Itoa(pid))
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

func psParentPID(pid int) (int, error) {
	out, err := psField(pid, "ppid")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(out)
}

func psProcessName(pid int) (string, error) { return psField(pid, "comm") }

func psProcessStart(pid int) (time.Time, error) {
	out, err := psField(pid, "lstart")
	if err != nil {
		return time.Time{}, err
	}
	return parsePsLstart(out)
}

// parsePsLstart parses `ps -o lstart=` ("Tue Nov  4 10:05:09 2026", local time).
func parsePsLstart(s string) (time.Time, error) {
	return time.ParseInLocation("Mon Jan 2 15:04:05 2006", strings.Join(strings.Fields(s), " "), time.Local)
}

// viewIdentityVariables are the shell variables verified for each host, in a
// fixed order. Cursor has none: its conversation ID equals the hook's only by
// inference, so it resolves through the session map.
var viewIdentityVariables = []struct{ env, host string }{
	{"CLAUDE_CODE_SESSION_ID", "claude-code"},
	{"PI_SESSION_ID", "pi"},
	{"OPENCODE_SESSION_ID", "opencode"},
}

// viewHostProcessNames maps the base name of a host CLI's process to its
// canonical host.
var viewHostProcessNames = map[string]string{
	"claude":       "claude-code",
	"codex":        "codex",
	"grok":         "grok",
	"opencode":     "opencode",
	"pi":           "pi",
	"cursor-agent": "cursor",
}

func hostOfProcessName(name string) string {
	return viewHostProcessNames[filepath.Base(strings.TrimSpace(name))]
}

const (
	viewIdentityAncestryLimit = 32
	viewPidMapSessionCap      = 64
	// ps reports the start time in whole seconds; a writer may round or
	// truncate, so two readings of one process differ by at most this.
	viewProcessStartTolerance = 2 * time.Second
)

// walkViewAncestors visits the parents of the current process, nearest first,
// until fn stops, the chain ends or the limit is reached.
func walkViewAncestors(deps viewIdentityDeps, fn func(pid int) (stop bool, err error)) error {
	seen := map[int]bool{}
	pid := deps.selfPID
	for n := 0; n < viewIdentityAncestryLimit && !seen[pid]; n++ {
		seen[pid] = true
		parent, err := deps.parentPID(pid)
		if err != nil || parent <= 1 {
			return nil
		}
		pid = parent
		if stop, err := fn(pid); stop || err != nil {
			return err
		}
	}
	return nil
}

func resolveViewIdentity(deps viewIdentityDeps, state string) (viewIdentity, error) {
	var found []viewIdentity
	for _, v := range viewIdentityVariables {
		value := deps.getenv(v.env)
		if value != "" && validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(v.host), NativeSessionID: value}) == nil {
			found = append(found, viewIdentity{Host: v.host, NativeSessionID: value})
		}
	}
	if len(found) == 1 {
		return confirmVariableHost(deps, state, found[0])
	}
	if len(found) > 1 {
		// Nested hosts inherit each other's variables; nothing says which is nearest.
		return viewIdentity{}, errViewIdentityAmbiguous
	}
	var result viewIdentity
	var resultErr = errViewIdentityMissing
	err := walkViewAncestors(deps, func(pid int) (bool, error) {
		doc, present, err := loadPidMap(deps, state, pid)
		if err != nil {
			resultErr = errViewIdentityAmbiguous
			return true, nil
		}
		if !present {
			return false, nil
		}
		result, resultErr = doc.liveIdentity()
		return true, nil
	})
	if err != nil {
		return viewIdentity{}, err
	}
	return result, resultErr
}

// confirmVariableHost accepts a host variable only when the nearest evidence
// of a host above this process, a session map or a known host CLI process, is
// that variable's own host: a host without its own variable, started from
// another agent's shell, inherits the outer agent's.
func confirmVariableHost(deps viewIdentityDeps, state string, id viewIdentity) (viewIdentity, error) {
	verdict := error(nil)
	err := walkViewAncestors(deps, func(pid int) (bool, error) {
		decided := false
		doc, present, err := loadPidMap(deps, state, pid)
		if err != nil {
			verdict = errViewIdentityAmbiguous
			return true, nil
		}
		if present {
			decided = true
			if doc.Host != id.Host {
				verdict = errViewIdentityAmbiguous
			}
		}
		if deps.processName != nil {
			if name, err := deps.processName(pid); err == nil {
				if h := hostOfProcessName(name); h != "" {
					decided = true
					if h != id.Host {
						verdict = errViewIdentityAmbiguous
					}
				}
			}
		}
		return decided, nil
	})
	if err != nil {
		return viewIdentity{}, err
	}
	if verdict != nil {
		return viewIdentity{}, verdict
	}
	return id, nil
}

type pidMapSessionEntry struct {
	ID      string
	Started time.Time
	Ended   bool
}

type pidMapDoc struct {
	Host     string
	Changed  time.Time
	Sessions []pidMapSessionEntry
}

// liveIdentity: a session is live when it has not ended and did not start
// before changed_at; exactly one live session resolves. changed_at moves only
// when a session replaces another in the same process (/clear, resume into
// another session, end), never when a second session starts alongside, so two
// live sessions stay detectable.
func (d *pidMapDoc) liveIdentity() (viewIdentity, error) {
	var live []viewIdentity
	for _, s := range d.Sessions {
		if s.Ended || s.Started.Before(d.Changed) {
			continue
		}
		live = append(live, viewIdentity{Host: d.Host, NativeSessionID: s.ID})
	}
	if len(live) != 1 {
		return viewIdentity{}, errViewIdentityAmbiguous
	}
	return live[0], nil
}

// loadPidMap reads and validates <state>/hooks/sessions/<pid>.json. present is
// false, with no error, when there is no file or when the file belongs to an
// earlier process that held the same pid (the process_started_at of the map
// differs from the live process's). Any other failure is an error: a map that
// exists but cannot be trusted is never skipped. Layout:
//
//	{"host": "<host>", "pid": N, "process_started_at": "<RFC 3339>", "changed_at": "<RFC 3339>",
//	 "sessions": [{"session_id": "...", "started_at": "<RFC 3339>", "ended_at": "<RFC 3339, optional>"}]}
func loadPidMap(deps viewIdentityDeps, state string, pid int) (*pidMapDoc, bool, error) {
	hooks := filepath.Join(state, "hooks")
	dir := filepath.Join(hooks, "sessions")
	for _, d := range []string{hooks, dir} {
		if err := checkSafeDir(d); err != nil {
			if os.IsNotExist(err) {
				return nil, false, nil
			}
			return nil, false, err
		}
	}
	raw, err := readSafeMapFile(filepath.Join(dir, strconv.Itoa(pid)+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	var doc struct {
		Host             string   `json:"host"`
		PID              *float64 `json:"pid"`
		ProcessStartedAt string   `json:"process_started_at"`
		ChangedAt        string   `json:"changed_at"`
		Sessions         []struct {
			SessionID string `json:"session_id"`
			StartedAt string `json:"started_at"`
			EndedAt   string `json:"ended_at"`
		} `json:"sessions"`
	}
	if json.Unmarshal(raw, &doc) != nil || doc.PID == nil || wholePID(*doc.PID) != pid || len(doc.Sessions) > viewPidMapSessionCap ||
		validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(doc.Host), NativeSessionID: "host-check"}) != nil {
		return nil, false, errViewIdentityAmbiguous
	}
	mapStart, err := time.Parse(time.RFC3339Nano, doc.ProcessStartedAt)
	if err != nil {
		return nil, false, errViewIdentityAmbiguous
	}
	out := &pidMapDoc{Host: doc.Host}
	if out.Changed, err = time.Parse(time.RFC3339Nano, doc.ChangedAt); err != nil {
		return nil, false, errViewIdentityAmbiguous
	}
	for _, s := range doc.Sessions {
		started, err := time.Parse(time.RFC3339Nano, s.StartedAt)
		if err != nil || validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(doc.Host), NativeSessionID: s.SessionID}) != nil {
			return nil, false, errViewIdentityAmbiguous
		}
		entry := pidMapSessionEntry{ID: s.SessionID, Started: started}
		if s.EndedAt != "" {
			if _, err := time.Parse(time.RFC3339Nano, s.EndedAt); err != nil {
				return nil, false, errViewIdentityAmbiguous
			}
			entry.Ended = true
		}
		out.Sessions = append(out.Sessions, entry)
	}
	if deps.processStart == nil {
		return nil, false, errViewIdentityAmbiguous
	}
	live, err := deps.processStart(pid)
	if err != nil {
		return nil, false, errViewIdentityAmbiguous
	}
	if d := live.Sub(mapStart); d > viewProcessStartTolerance || d < -viewProcessStartTolerance {
		return nil, false, nil // the pid was reused by another process
	}
	return out, true, nil
}

// --- safe locations ---

func ownedByCurrentUser(st os.FileInfo) bool {
	stat, ok := st.Sys().(*syscall.Stat_t)
	return ok && int(stat.Uid) == os.Getuid()
}

// checkSafeDir rejects a folder that is a symlink, is not owned by the current
// user or is group- or world-writable. A missing folder returns the not-exist
// error unchanged.
func checkSafeDir(path string) error {
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !st.IsDir() || !ownedByCurrentUser(st) || st.Mode().Perm()&0022 != 0 {
		return workStateError("unsafe_publisher_file")
	}
	return nil
}

// readSafeMapFile reads a session map: a regular file, owned by the current
// user, not group- or world-writable and without extra hard links, opened with
// no-follow semantics and re-checked on the open descriptor.
func readSafeMapFile(path string) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || !ownedByCurrentUser(before) || before.Mode().Perm()&0022 != 0 || workFileHasLinks(before) {
		return nil, errViewIdentityAmbiguous
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errViewIdentityAmbiguous
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil || !os.SameFile(before, opened) || !opened.Mode().IsRegular() || !ownedByCurrentUser(opened) || opened.Mode().Perm()&0022 != 0 || workFileHasLinks(opened) {
		return nil, errViewIdentityAmbiguous
	}
	raw, err := io.ReadAll(io.LimitReader(f, sessionFileLimit+1))
	if err != nil || len(raw) > sessionFileLimit {
		return nil, errors.Join(errViewIdentityAmbiguous, err)
	}
	return raw, nil
}
