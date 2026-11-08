package main

// view_identity.go resolves the native identity of a semantic `hive view work`
// command when no --host/--native-session-id flag is given, and keeps the
// current work of each native session.
//
// Identity sources, first one that exists wins:
//  1. explicit flags (handled by the caller; this file is never reached);
//  2. the session variable the host exposes to its agent's shell, only the ones
//     in host-session-events.md marked usable: CLAUDE_CODE_SESSION_ID,
//     PI_SESSION_ID, OPENCODE_SESSION_ID and CODEX_SESSION_ID. A nested host
//     inherits the outer variable, so the variable of the nearest host process
//     or session map wins; when that host has none among them, only its own map
//     applies. Codex's CODEX_THREAD_ID (a subagent's own thread) and Cursor's
//     CURSOR_CONVERSATION_ID are deliberately not read;
//  3. the process-to-session map <state>/hooks/sessions/<pid>.json written by
//     the host's session-start hook (this file only reads it), looked up on the
//     nearest process ancestor that has one. The format and writer rules are
//     the «Session map» section of host-session-events.md.
//
// A source that exists but cannot name exactly one live session is a rejection,
// never a guess; the caller falls back to the explicit flags.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"time"
	"tricell-hive/tooling/management"
	"tricell-hive/tooling/viewer"
)

var (
	// errViewIdentityMissing: no source applies.
	errViewIdentityMissing = errors.New("native identity unavailable")
	// errViewIdentityAmbiguous: a source applies but does not name exactly one
	// live session, or cannot be trusted.
	errViewIdentityAmbiguous = errors.New("native identity ambiguous")
)

type viewIdentity struct{ Host, NativeSessionID string }

// viewIdentityDeps are the process facts the resolver reads, injectable so
// tests never see the real environment or process tree.
type viewIdentityDeps struct {
	getenv    func(string) string
	selfPID   int
	parentPID func(int) (int, error)
	// processStart is the start time of a live process; processName its
	// command name (only the base name is used).
	processStart func(int) (time.Time, error)
	processName  func(int) (string, error)
}

// --- current work ---

const viewCurrentWorkCap = 16

type viewCurrentWorks struct {
	WorkIDs []string `json:"work_ids"`
}

// currentWorkPath is <state>/publisher/v1/current/<host>/<native-id>.json. The
// host and the native ID are validated first, because they become path parts.
func currentWorkPath(state, host, native string) (string, error) {
	if validateViewWorkSource(viewWorkSource{Host: viewer.TimelineHost(host), NativeSessionID: native}) != nil {
		return "", workInput("invalid native source")
	}
	name := native
	if !viewCurrentFileName.MatchString(native) {
		// The contract admits IDs that are not safe file names; such an ID is
		// keyed by its digest.
		sum := sha256.Sum256([]byte(native))
		name = "sha256-" + hex.EncodeToString(sum[:])
	}
	return filepath.Join(state, "publisher", "v1", "current", host, name+".json"), nil
}

var viewCurrentFileName = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,127}$`) // lower case only: names must not collide on case-insensitive file systems

// checkCurrentWorkDirs applies the safe-folder check to current/ and
// current/<host>/ before a current-work file is read or written; a missing
// folder returns the not-exist error.
func checkCurrentWorkDirs(file string) error {
	hostDir := filepath.Dir(file)
	for _, d := range []string{filepath.Dir(hostDir), hostDir} {
		if err := checkSafeDir(d); err != nil {
			return err
		}
	}
	return nil
}

// readCurrentWorks returns the open works of a native session; a missing file
// is none.
func readCurrentWorks(state, host, native string) ([]string, error) {
	path, err := currentWorkPath(state, host, native)
	if err != nil {
		return nil, err
	}
	if err = checkCurrentWorkDirs(path); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	raw, err := readPrivateWorkFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var doc viewCurrentWorks
	if json.Unmarshal(raw, &doc) != nil {
		return nil, workStateError("invalid_current_work")
	}
	for _, id := range doc.WorkIDs {
		if !viewTimelineSessionID.MatchString(id) {
			return nil, workStateError("invalid_current_work")
		}
	}
	return doc.WorkIDs, nil
}

func writeCurrentWorks(state, host, native string, ids []string) error {
	path, err := currentWorkPath(state, host, native)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err = checkCurrentWorkDirs(path); err != nil {
		return err
	}
	raw, err := json.Marshal(viewCurrentWorks{WorkIDs: ids})
	if err != nil {
		return err
	}
	return writePrivateWorkFile(dir, path, raw)
}

// addCurrentWork records work as open in the session, under the session lock.
func addCurrentWork(state, host, native, work string) error {
	if !viewTimelineSessionID.MatchString(work) {
		return workInput("work-id")
	}
	return withSessionLock(state, host, native, func() error { return addCurrentWorkLocked(state, host, native, work) })
}

// withSessionLock serializes the read-modify-write of a session's current-work
// file between processes: the per-work locks do not, since one session's
// several works share that file.
func withSessionLock(state, host, native string, fn func() error) error {
	return withSessionLockFile(state, host, native, func() (bool, error) { return false, fn() })
}

// withSessionLockFile also removes the lock file when fn reports that the
// session has no current work left; the lock is still held then, so a held lock
// file is never removed.
func withSessionLockFile(state, host, native string, fn func() (remove bool, err error)) error {
	_, dir, err := workPublisherDir(viewWorkCommand{Options: management.Options{Scope: "user", StateDir: state}})
	if err != nil {
		return err
	}
	if err = ensureWorkSubdir(dir, "locks"); err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(host + "\x00" + native))
	path := filepath.Join(dir, "locks", "session-"+hex.EncodeToString(sum[:])+".lock")
	l, err := acquirePrivateWorkLock(path, true)
	if err != nil {
		return err
	}
	defer l.Release()
	remove, err := fn()
	if remove && err == nil {
		os.Remove(path)
	}
	return err
}

// currentWorkRefs lists every work ID that any session's current-work file names.
func currentWorkRefs(state string) (map[string]bool, error) {
	files, err := filepath.Glob(filepath.Join(state, "publisher", "v1", "current", "*", "*.json"))
	if err != nil {
		return nil, err
	}
	refs := map[string]bool{}
	for _, file := range files {
		raw, err := readPrivateWorkFile(file)
		if err != nil {
			return nil, err
		}
		var doc viewCurrentWorks
		if json.Unmarshal(raw, &doc) != nil {
			return nil, workStateError("invalid_current_work")
		}
		for _, id := range doc.WorkIDs {
			refs[id] = true
		}
	}
	return refs, nil
}

func addCurrentWorkLocked(state, host, native, work string) error {
	ids, err := readCurrentWorks(state, host, native)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == work {
			return nil
		}
	}
	ids = append(ids, work)
	if len(ids) > viewCurrentWorkCap {
		ids = ids[len(ids)-viewCurrentWorkCap:]
	}
	return writeCurrentWorks(state, host, native, ids)
}

// releaseCurrentWork removes work from the session's open works. A session left
// with none loses its current-work file and its lock file.
func releaseCurrentWork(state, host, native, work string) error {
	return withSessionLockFile(state, host, native, func() (bool, error) { return releaseCurrentWorkLocked(state, host, native, work) })
}

func releaseCurrentWorkLocked(state, host, native, work string) (bool, error) {
	ids, err := readCurrentWorks(state, host, native)
	if err != nil {
		return false, err
	}
	kept := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != work {
			kept = append(kept, id)
		}
	}
	if len(kept) == len(ids) {
		return len(kept) == 0, nil
	}
	if len(kept) == 0 {
		path, err := currentWorkPath(state, host, native)
		if err != nil {
			return false, err
		}
		if err = os.Remove(path); err != nil && !os.IsNotExist(err) {
			return false, err
		}
		return true, nil
	}
	return false, writeCurrentWorks(state, host, native, kept)
}

// viewWorkActionNeedsWork lists the semantic actions that act on an existing
// work and so may omit --work-id while exactly one work is current. open and
// publish create when it is omitted and never reuse the current work.
func viewWorkActionNeedsWork(action string) bool {
	switch action {
	case "stage", "sync", "evidence", "tasks register", "task update", "close", "resume", "reconcile", "bind", "show":
		return true
	}
	return false
}

// resolveViewWorkContext fills the identity and the work of a parsed semantic
// command from the resolved sources. Explicit flags are never overridden, and a
// half-given identity is left for validation to reject.
func resolveViewWorkContext(c *viewWorkCommand) error {
	if isViewWorkLocal(c.Action) {
		return nil
	}
	var state string
	stateDir := func() (string, error) {
		if state == "" {
			_, s, err := management.NormalizeOptions(c.Options)
			if err == nil {
				s, err = filepath.Abs(s)
			}
			if err != nil {
				return "", err
			}
			state = s
		}
		return state, nil
	}
	// list reads by project and show by work ID; show without --work-id needs the
	// session's identity to find its current work.
	needsIdentity := c.Action != "list" && (c.Action != "show" || c.WorkID == "")
	if c.Host == "" && c.NativeSessionID == "" && needsIdentity {
		dir, err := stateDir()
		if err != nil {
			return err
		}
		id, err := resolveViewIdentity(viewIdentityDepsFn(), dir)
		switch {
		case err == nil:
			c.Host, c.NativeSessionID = id.Host, id.NativeSessionID
		case errors.Is(err, errViewIdentityAmbiguous):
			return workInput("native identity ambiguous: pass --host and --native-session-id")
		case errors.Is(err, errViewIdentityMissing):
			return workInput("native identity missing: pass --host and --native-session-id")
		default:
			return workInput("native identity unresolved: pass --host and --native-session-id")
		}
	}
	// A capture acts on the session's one open work and opens one when it has
	// none, so it reads the current works but never reports their absence.
	capture := c.Action == "capture"
	if c.WorkID == "" && c.Host != "" && c.NativeSessionID != "" && (viewWorkActionNeedsWork(c.Action) || capture) && !hostReconcile(*c) {
		dir, err := stateDir()
		if err != nil {
			return err
		}
		ids, err := readCurrentWorks(dir, c.Host, c.NativeSessionID)
		if err != nil {
			return err
		}
		switch len(ids) {
		case 0:
			if !capture {
				return workInput("work-id required: this session has no open work")
			}
		case 1:
			c.WorkID = ids[0]
		default:
			return workInput("work-id required: several works are open in this session")
		}
	}
	return nil
}

// hostReconcile is `reconcile` in its host-session form: no pending intent
// named, so it reads the host's session file instead of archiving a rejection.
func hostReconcile(c viewWorkCommand) bool {
	return c.Action == "reconcile" && c.Pending == "" && c.ReconcileAction == ""
}
