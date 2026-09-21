package management

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"tricell-hive/integrations/target"
)

type entry struct {
	Change        Change
	Before, After snapshot
}
type journal struct {
	Version                 int
	ID, Phase               string
	Plan                    Plan
	Entries                 []entry
	BeforeState, AfterState snapshot
	CreatedDirs             []string
	Integrity               string
}

func journalHash(j journal) string { j.Integrity = ""; return hash(encode(j)) }
func saveJournal(path string, j journal) error {
	j.Integrity = journalHash(j)
	return writeJSON(path, j)
}

type pending struct{ ID string }

// Engine's failpoint is internal and used only by failure-injection tests.
type Engine struct{ failpoint func(string) error }

func (e Engine) fail(stage string) error {
	if e.failpoint != nil {
		return e.failpoint(stage)
	}
	return nil
}
func missingDirs(paths []string) ([]string, error) {
	set := map[string]bool{}
	for _, path := range paths {
		for p := filepath.Dir(path); ; p = filepath.Dir(p) {
			info, err := os.Stat(p)
			if err == nil {
				if !info.IsDir() {
					return nil, fmt.Errorf("not a directory: %s", p)
				}
				break
			}
			if !os.IsNotExist(err) {
				return nil, err
			}
			set[p] = true
		}
	}
	var out []string
	for p := range set {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) == len(out[j]) {
			return out[i] < out[j]
		}
		return len(out[i]) < len(out[j])
	})
	return out, nil
}
func cleanupDirs(dirs []string) []string {
	dirs = append([]string(nil), dirs...)
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i]) > len(dirs[j]) })
	var retained []string
	for _, d := range dirs {
		if target.Safe(d) != nil {
			retained = append(retained, d)
			continue
		}
		err := os.Remove(d)
		if err != nil && !os.IsNotExist(err) {
			retained = append(retained, d)
		}
	}
	sort.Strings(retained)
	return retained
}
func (e Engine) Apply(p Plan) (string, error) {
	// Validate before creating a state directory, then repeat under the lock.
	state, sh, err := readState(p.StateDir)
	if err != nil {
		return "", err
	}
	if err = validatePlan(p, state); err != nil {
		return "", err
	}
	if sh != p.StateHash {
		return "", fmt.Errorf("stale plan: state changed")
	}
	unlock, err := lock(p.StateDir)
	if err != nil {
		return "", err
	}
	defer unlock()
	pp := filepath.Join(p.StateDir, "pending.json")
	if s, err := read(pp); err != nil {
		return "", err
	} else if s.Exists {
		return "", fmt.Errorf("unfinished operation: recover first")
	}
	state, sh, err = readState(p.StateDir)
	if err != nil {
		return "", err
	}
	if sh != p.StateHash {
		return "", fmt.Errorf("stale plan: state changed")
	}
	if err = validatePlan(p, state); err != nil {
		return "", err
	}
	beforeState, err := read(filepath.Join(p.StateDir, "state.json"))
	if err != nil {
		return "", err
	}
	next := emptyState()
	for k, r := range state.Records {
		next.Records[k] = r
	}
	next.CreatedDirs = append([]string(nil), state.CreatedDirs...)
	j := journal{Version: 3, Phase: "prepared", Plan: p, BeforeState: beforeState}
	var paths []string
	changed := state.Version != 3
	for _, ch := range p.Changes {
		s, err := readResource(ch.Target, ch.Replaces != nil)
		if err != nil {
			return "", err
		}
		if finger(s) != ch.Expected {
			return "", fmt.Errorf("stale plan: target changed: %s", ch.Target.Path)
		}
		after, err := transformResource(s, ch)
		if err != nil {
			return "", err
		}
		j.Entries = append(j.Entries, entry{ch, s, after})
		if after.Exists {
			paths = append(paths, ch.Target.Path)
		}
		if !same(s, after) || !reflect.DeepEqual(ch.Before, ch.After) {
			changed = true
		}
		if ch.After == nil {
			delete(next.Records, ch.Target.Path)
		} else {
			next.Records[ch.Target.Path] = *ch.After
		}
		if ch.Replaces != nil {
			delete(next.Records, ch.Replaces.Target.Path)
			var kept []string
			for _, d := range next.CreatedDirs {
				if d != ch.Target.Path {
					kept = append(kept, d)
				}
			}
			next.CreatedDirs = kept
		}
	}
	if !changed {
		return "unchanged", nil
	}
	dirsToCreate, err := missingDirs(paths)
	if err != nil {
		return "", err
	}
	next.CreatedDirs = append(next.CreatedDirs, dirsToCreate...)
	j.AfterState = snapshot{Exists: true, Data: encode(next), Mode: 0600}
	id := make([]byte, 16)
	if _, err = rand.Read(id); err != nil {
		return "", err
	}
	j.ID = hex.EncodeToString(id)
	for _, d := range []string{"transactions", "releases"} {
		path := filepath.Join(p.StateDir, d)
		if err = target.Safe(path); err != nil {
			return "", err
		}
		if err = os.MkdirAll(path, 0700); err != nil {
			return "", err
		}
	}
	jp := filepath.Join(p.StateDir, "transactions", j.ID+".json")
	if err = saveJournal(jp, j); err != nil {
		return "", err
	}
	if err = writeJSON(pp, pending{j.ID}); err != nil {
		return "", err
	}
	runErr := func() error {
		if err := e.fail("prepared"); err != nil {
			return err
		}
		for _, d := range dirsToCreate {
			if err := target.Safe(d); err != nil {
				return err
			}
			if err := os.Mkdir(d, 0700); err != nil {
				return err
			}
			j.CreatedDirs = append(j.CreatedDirs, d)
			if err := saveJournal(jp, j); err != nil {
				return err
			}
		}
		for i, en := range j.Entries {
			if same(en.Before, en.After) {
				continue
			}
			cur, err := readResource(en.Change.Target, en.Change.Replaces != nil)
			if err != nil {
				return err
			}
			if !same(cur, en.Before) {
				return fmt.Errorf("concurrent change: %s", en.Change.Target.Path)
			}
			// Resolve again to catch newly introduced overrides/imports before each write.
			if err = validatePlan(p, state); err != nil {
				return err
			}
			if err = writeResource(en.Change.Target, cur, en.After, en.Change.Replaces != nil, e.fail); err != nil {
				return err
			}
			if err = e.fail(fmt.Sprintf("write:%d", i)); err != nil {
				return err
			}
		}
		for _, en := range j.Entries {
			cur, err := readResource(en.Change.Target, en.Change.Replaces != nil)
			if err != nil {
				return err
			}
			if !same(cur, en.After) {
				return fmt.Errorf("post-write verification failed")
			}
		}
		if p.Release != nil {
			path := filepath.Join(p.StateDir, "releases", p.Release.ID+".json")
			if old, err := read(path); err != nil {
				return err
			} else if old.Exists && string(old.Data) != string(encode(p.Release)) {
				return fmt.Errorf("release cache conflict")
			}
			if err := writeJSON(path, p.Release); err != nil {
				return err
			}
		}
		if err := write(filepath.Join(p.StateDir, "state.json"), j.AfterState); err != nil {
			return err
		}
		if err := e.fail("state"); err != nil {
			return err
		}
		j.Phase = "committed"
		if err := saveJournal(jp, j); err != nil {
			return err
		}
		if err := os.Remove(pp); err != nil {
			return err
		}
		return nil
	}()
	if runErr != nil {
		return j.ID, fmt.Errorf("transaction %s requires recover: %w", j.ID, runErr)
	}
	// Only remove recorded empty directories after the transaction has committed.
	next.CreatedDirs = cleanupDirs(next.CreatedDirs)
	if err = writeJSON(filepath.Join(p.StateDir, "state.json"), next); err != nil {
		return j.ID, fmt.Errorf("installed; directory bookkeeping failed: %w", err)
	}
	return j.ID, nil
}
func (e Engine) Recover(stateDir string) (string, error) {
	dir, err := target.Canonical(stateDir)
	if err != nil {
		return "", err
	}
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return "no_pending_operation", nil
	}
	unlock, err := lock(dir)
	if err != nil {
		return "", err
	}
	defer unlock()
	pp := filepath.Join(dir, "pending.json")
	var p pending
	if err = decodeFile(pp, &p); os.IsNotExist(err) {
		return "no_pending_operation", nil
	} else if err != nil {
		return "", err
	}
	if len(p.ID) != 32 {
		return "", fmt.Errorf("invalid transaction ID")
	}
	if _, err = hex.DecodeString(p.ID); err != nil {
		return "", err
	}
	jp := filepath.Join(dir, "transactions", p.ID+".json")
	var j journal
	if err = decodeFile(jp, &j); err != nil {
		return "", err
	}
	if (j.Version != 1 && j.Version != 2 && j.Version != 3) || j.Version != j.Plan.Version || j.ID != p.ID || j.Plan.StateDir != dir || j.Integrity != journalHash(j) || j.Plan.ID != planID(j.Plan) {
		return "", fmt.Errorf("invalid transaction")
	}
	if j.Phase == "committed" || j.Phase == "recovered" {
		return j.Phase, os.Remove(pp)
	}
	stateNow, err := read(filepath.Join(dir, "state.json"))
	if err != nil {
		return "", err
	}
	if !same(stateNow, j.BeforeState) && !same(stateNow, j.AfterState) {
		return "", fmt.Errorf("state changed after interruption; preserved")
	}
	// Preflight every inverse before touching any resource. Outside-block edits survive.
	inverses := make([]snapshot, len(j.Entries))
	currents := make([]snapshot, len(j.Entries))
	for i, en := range j.Entries {
		cur, err := readResource(en.Change.Target, en.Change.Replaces != nil)
		if err != nil {
			return "", err
		}
		currents[i] = cur
		if same(cur, en.Before) {
			inverses[i] = cur
			continue
		}
		if en.Change.Replaces != nil && (!cur.Exists || (cur.Kind == "directory" && cur.Mode == en.Before.Mode)) {
			// These are the two journaled intermediate states of the one
			// allowed directory-to-link migration.
			inverses[i] = en.Before
			continue
		}
		// A previous recovery attempt may already have restored the managed span
		// while retaining newer text outside it.
		if en.Change.Target.Kind == "block" && cur.Mode == en.Before.Mode && en.Change.Before != nil && owned(cur, *en.Change.Before) == nil {
			inverses[i] = cur
			continue
		}
		if en.Change.Target.Kind == "block" && en.Change.Before == nil {
			a, _, parseErr := blockRange(cur.Data)
			if parseErr == nil && a < 0 && cur.Exists {
				inverses[i] = cur
				continue
			}
		}
		if same(cur, en.After) {
			inverses[i] = en.Before
			continue
		}
		if en.Change.Target.Kind != "block" || en.Change.After == nil || cur.Mode != en.After.Mode {
			return "", fmt.Errorf("recovery conflict: %s; preserved", en.Change.Target.Path)
		}
		inv, err := transform(cur, en.Change.After, en.Change.Before)
		if err != nil {
			return "", fmt.Errorf("recovery conflict: %w", err)
		}
		inverses[i] = inv
	}
	for i := len(j.Entries) - 1; i >= 0; i-- {
		en := j.Entries[i]
		cur, err := readResource(en.Change.Target, en.Change.Replaces != nil)
		if err != nil {
			return "", err
		}
		if !same(cur, currents[i]) {
			return "", fmt.Errorf("concurrent recovery change")
		}
		if !same(cur, inverses[i]) {
			if err = writeResource(en.Change.Target, cur, inverses[i], en.Change.Replaces != nil, nil); err != nil {
				return "", err
			}
			if err = e.fail(fmt.Sprintf("recover:%d", i)); err != nil {
				return "", err
			}
		}
	}
	if err = write(filepath.Join(dir, "state.json"), j.BeforeState); err != nil {
		return "", err
	}
	cleanupDirs(j.CreatedDirs)
	j.Phase = "recovered"
	if err = saveJournal(jp, j); err != nil {
		return "", err
	}
	if err = os.Remove(pp); err != nil {
		return "", err
	}
	return "recovered:" + j.ID, nil
}
