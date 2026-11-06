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
// fixed order. Codex's is the root session ID its shell tools receive
// (codex-rs/core/src/exec_env.rs inject_session_env, observed equal to the
// hook's session_id on 0.167.0); CODEX_THREAD_ID names a subagent's own thread
// and is not one. Cursor and Grok have none and resolve through the session map.
var viewIdentityVariables = []struct{ env, host string }{
	{"CLAUDE_CODE_SESSION_ID", "claude-code"},
	{"PI_SESSION_ID", "pi"},
	{"OPENCODE_SESSION_ID", "opencode"},
	{"CODEX_SESSION_ID", "codex"},
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
	if len(found) > 0 {
		return chooseVariable(deps, state, found)
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

// chooseVariable picks among the host variables of this shell by the nearest
// evidence of a host above it, a session map or a host CLI process: hosts
// started from another agent's shell inherit the outer agent's variable. Every
// variable that is not the nearest host's must come from a host found further
// up; otherwise it may belong to an inner host the walk did not recognize, and
// the result is ambiguous. Then the nearest host's own variable wins, and when
// it has none among them, only that host's own map, on that process, applies.
// With no host evidence, a single variable is used and several are ambiguous.
func chooseVariable(deps viewIdentityDeps, state string, found []viewIdentity) (viewIdentity, error) {
	near, err := nearestHost(deps, state)
	if err != nil {
		return viewIdentity{}, err
	}
	if near.host == "" {
		if len(found) == 1 {
			return found[0], nil
		}
		return viewIdentity{}, errViewIdentityAmbiguous
	}
	var others []string
	for _, id := range found {
		if id.Host != near.host {
			others = append(others, id.Host)
		}
	}
	if len(others) > 0 && !hostsAbove(deps, state, near.pid, others) {
		return viewIdentity{}, errViewIdentityAmbiguous
	}
	for _, id := range found {
		if id.Host == near.host {
			return id, nil
		}
	}
	if near.doc == nil {
		return viewIdentity{}, errViewIdentityMissing
	}
	return near.doc.liveIdentity()
}

// nearestEvidence is the nearest ancestor that has a session map or is a host
// CLI: its host, its pid and its map when it has one. host is "" when no
// ancestor is either.
type nearestEvidence struct {
	host string
	pid  int
	doc  *pidMapDoc
}

// nearestHost finds the nearest evidence of a host. A map that cannot be
// trusted, or a process whose map and name disagree, is ambiguous.
func nearestHost(deps viewIdentityDeps, state string) (nearestEvidence, error) {
	var near nearestEvidence
	verdict := error(nil)
	err := walkViewAncestors(deps, func(pid int) (bool, error) {
		doc, present, err := loadPidMap(deps, state, pid)
		if err != nil {
			verdict = errViewIdentityAmbiguous
			return true, nil
		}
		if present {
			near = nearestEvidence{host: doc.Host, pid: pid, doc: doc}
		}
		if h := hostOfProcess(deps, pid); h != "" {
			if near.host != "" && h != near.host {
				verdict = errViewIdentityAmbiguous
			}
			near.host, near.pid = h, pid
		}
		return near.host != "", nil
	})
	if err == nil {
		err = verdict
	}
	if err != nil {
		return nearestEvidence{}, err
	}
	return near, nil
}

// hostsAbove reports whether every host in need is found among the ancestors
// of pid, by session map or host CLI process, and stops as soon as it is; a map
// that cannot be read names nothing.
func hostsAbove(deps viewIdentityDeps, state string, pid int, need []string) bool {
	missing := map[string]bool{}
	for _, h := range need {
		missing[h] = true
	}
	from := deps
	from.selfPID = pid
	_ = walkViewAncestors(from, func(p int) (bool, error) {
		if doc, present, err := loadPidMap(deps, state, p); err == nil && present {
			delete(missing, doc.Host)
		}
		if h := hostOfProcess(deps, p); h != "" {
			delete(missing, h)
		}
		return len(missing) == 0, nil
	})
	return len(missing) == 0
}

// hostOfProcess is the host CLI a process is, recognized as the map writer
// recognizes it: by its command name, or by its script under an interpreter.
func hostOfProcess(deps viewIdentityDeps, pid int) string {
	if deps.processName == nil {
		return ""
	}
	comm, err := deps.processName(pid)
	if err != nil {
		return ""
	}
	args := ""
	if base := filepath.Base(strings.TrimSpace(comm)); base == "node" || base == "bun" || base == "deno" {
		args, _ = hookProcessArgs(pid)
	}
	return processHost(comm, args)
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
