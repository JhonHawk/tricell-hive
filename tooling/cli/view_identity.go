package main

// view_identity.go resolves the native identity of a semantic `hive view work`
// command when no --host/--native-session-id flag is given, and keeps the
// current work of each native session.
//
// Identity sources, first one that exists wins:
//  1. explicit flags (handled by the caller; this file is never reached);
//  2. the session variable the host exposes to its agent's shell, only the ones
//     in host-session-events.md marked usable: CLAUDE_CODE_SESSION_ID,
//     PI_SESSION_ID and OPENCODE_SESSION_ID, and only when no nearer host
//     process or session map belongs to another host (a nested host inherits the
//     outer variable). Codex's CODEX_THREAD_ID and Cursor's
//     CURSOR_CONVERSATION_ID are unverified and deliberately not read;
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

// addCurrentWork records work as open in the session. Callers hold the
// publisher lock, which serializes this read-modify-write.
func addCurrentWork(state, host, native, work string) error {
	if !viewTimelineSessionID.MatchString(work) {
		return workInput("work-id")
	}
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

// releaseCurrentWork removes work from the session's open works.
func releaseCurrentWork(state, host, native, work string) error {
	ids, err := readCurrentWorks(state, host, native)
	if err != nil {
		return err
	}
	kept := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != work {
			kept = append(kept, id)
		}
	}
	if len(kept) == len(ids) {
		return nil
	}
	return writeCurrentWorks(state, host, native, kept)
}

// viewWorkActionNeedsWork lists the semantic actions that act on an existing
// work and so may omit --work-id while exactly one work is current. open and
// publish create when it is omitted and never reuse the current work.
func viewWorkActionNeedsWork(action string) bool {
	switch action {
	case "stage", "tasks register", "task update", "close", "resume", "reconcile", "bind", "show":
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
	if c.Host == "" && c.NativeSessionID == "" && c.Action != "list" && c.Action != "show" {
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
		}
	}
	if c.WorkID == "" && c.Host != "" && c.NativeSessionID != "" && viewWorkActionNeedsWork(c.Action) {
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
		case 1:
			c.WorkID = ids[0]
		default:
			return workInput("work-id required: several works are open in this session")
		}
	}
	return nil
}
