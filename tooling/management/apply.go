package management

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"time"
	"tricell-hive/integrations/target"
)

type entry struct {
	Change Change
	// Voice is this path's voice-block change in the same plan, or nil when
	// only the Hive block (or only legacy retirement) touches this path.
	// One entry always means one write, so a path with both a Change and a
	// voice change shares this single entry (see prepareEntries).
	Voice         *VoiceChange `json:",omitempty"`
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
	PackageDone             bool `json:",omitempty"`
	PackagePending          bool `json:",omitempty"`
}

func journalHash(j journal) string { j.Integrity = ""; return hash(encode(j)) }
func saveJournal(path string, j journal) error {
	j.Integrity = journalHash(j)
	return writeJSON(path, j)
}

type pending struct{ ID string }

// Engine's failpoint is internal and used only by failure-injection tests.
// nested marks a call made by Onboard (or RecoverOnboarding) on its own
// core transaction: the lock is already held and preflight already ran
// once under it, so Apply (or Recover) must not repeat either. presetState
// carries the State that preflight already validated, so a nested Apply
// never re-reads it.
type Engine struct {
	failpoint     func(string) error
	nested        bool
	presetState   State
	transactionID string
}

func (e Engine) fail(stage string) error {
	if e.failpoint != nil {
		return e.failpoint(stage)
	}
	return nil
}

// preflight is Apply's and Onboard's single authoritative gate. It rejects
// an obviously stale or invalid plan before creating a state directory,
// then re-validates it under the lock together with any pending core
// operation, returning the validated
// state and the still-held lock's release function. Both entry points call
// it exactly once; the nested core transaction Onboard starts trusts this
// result instead of repeating the checks (see Engine.nested).
func preflight(p Plan) (State, func(), error) {
	if err := checkOnboarding(p.StateDir); err != nil {
		return State{}, nil, err
	}
	state, sh, err := readState(p.StateDir)
	if err != nil {
		return State{}, nil, err
	}
	if err = validatePlan(p, state); err != nil {
		return State{}, nil, err
	}
	if err = validateMigration(p, state); err != nil {
		return State{}, nil, err
	}
	if sh != p.StateHash {
		return State{}, nil, fmt.Errorf("stale plan: state changed")
	}
	unlock, err := lock(p.StateDir)
	if err != nil {
		return State{}, nil, err
	}
	state, err = authoritativePreflight(p)
	if err != nil {
		unlock()
		return State{}, nil, err
	}
	return state, unlock, nil
}

// authoritativePreflight re-runs preflight's checks under the lock, where
// they are the ones that matter: nothing observed before the lock was
// acquired is trustworthy once another operation could have run.
func authoritativePreflight(p Plan) (State, error) {
	if err := checkOnboarding(p.StateDir); err != nil {
		return State{}, err
	}
	if s, err := read(filepath.Join(p.StateDir, "pending.json")); err != nil {
		return State{}, err
	} else if s.Exists {
		return State{}, fmt.Errorf("unfinished operation: recover first")
	}
	state, sh, err := readState(p.StateDir)
	if err != nil {
		return State{}, err
	}
	if err = validateMigration(p, state); err != nil {
		return State{}, err
	}
	if sh != p.StateHash {
		return State{}, fmt.Errorf("stale plan: state changed")
	}
	if err = validatePlan(p, state); err != nil {
		return State{}, err
	}
	return state, nil
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

// anyUserScopeHiveBlockRemains reports whether s.Records still has any
// "block"-kind record with at least one user-scope consumer: whether there
// is still somewhere a voice choice could apply.
func anyUserScopeHiveBlockRemains(s State) bool {
	for _, r := range s.Records {
		if r.Target.Kind != "block" {
			continue
		}
		for _, c := range r.Consumers {
			if c.Scope == "user" {
				return true
			}
		}
	}
	return false
}

// prepareTransaction computes the next state, the journal entries, and
// whether anything would actually change for p against state. It touches
// no persistent state beyond the advisory reads prepareEntries and
// overlayRead already perform. transactionID, when set, overrides the
// generated journal ID before any migration receipt is built from it,
// matching the core transaction ID Onboard assigns to its child journal.
func prepareTransaction(p Plan, state State, transactionID string) (journal, State, []string, bool, error) {
	beforeState, err := read(filepath.Join(p.StateDir, "state.json"))
	if err != nil {
		return journal{}, State{}, nil, false, err
	}
	next := emptyState()
	for k, r := range state.Records {
		next.Records[k] = r
	}
	next.CreatedDirs = append([]string(nil), state.CreatedDirs...)
	next.Migrations = append([]MigrationReceipt(nil), state.Migrations...)
	next.Voice = state.Voice
	next.PiSubagentsSource = state.PiSubagentsSource
	if state.VoiceSpans != nil {
		next.VoiceSpans = map[string]VoiceSpan{}
		for k, v := range state.VoiceSpans {
			next.VoiceSpans[k] = v
		}
	}
	j := journal{Version: stateVersion, Phase: "prepared", Plan: p, BeforeState: beforeState}
	changed := state.Version != stateVersion || p.Migration != nil || len(p.Legacy) > 0
	j.Entries, err = prepareEntries(p, state)
	if err != nil {
		return journal{}, State{}, nil, false, err
	}
	var paths []string
	for _, en := range j.Entries {
		if en.After.Exists {
			paths = append(paths, en.Change.Target.Path)
		}
	}
	for _, ch := range p.Changes {
		cur, err := overlayRead(p, ch.Target, ch.Replaces != nil)
		if err != nil {
			return journal{}, State{}, nil, false, err
		}
		after, err := transformResource(cur, ch, p.Action)
		if err != nil {
			return journal{}, State{}, nil, false, err
		}
		if !same(cur, after) || !reflect.DeepEqual(ch.Before, ch.After) {
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
	for _, vc := range p.Voice {
		if vc.Gone || !reflect.DeepEqual(vc.Before, vc.After) {
			changed = true
		}
		if vc.After == nil {
			delete(next.VoiceSpans, vc.Path)
		} else {
			if next.VoiceSpans == nil {
				next.VoiceSpans = map[string]VoiceSpan{}
			}
			next.VoiceSpans[vc.Path] = *vc.After
		}
	}
	if len(next.VoiceSpans) == 0 {
		next.VoiceSpans = nil
	}
	if p.Action == "voice" {
		if !reflect.DeepEqual(state.Voice, p.VoiceSetting) {
			changed = true
		}
		next.Voice = p.VoiceSetting
	} else if p.Action == "remove" && state.Voice != nil && next.VoiceSpans == nil && !anyUserScopeHiveBlockRemains(next) {
		// No voice span is left, and no user-scope Hive block is registered
		// at all any more: the choice has nothing left to apply to, so drop
		// it (design.md "plan remove"). A Hive block that simply has no span
		// yet (e.g. installed while the voice text could not be rendered,
		// see addVoiceChangesForInstall) must not lose the choice: it should
		// still regain a span once rendering works again.
		next.Voice = nil
		changed = true
	}
	if updateProductState(&next, state, p) {
		changed = true
	}
	// A change of overrides is saved even when no file reflects it, such as the
	// reset of an override whose role the release lost.
	next.ModelOverrides = nextModelOverrides(state, p)
	if !sameOverrides(state.ModelOverrides, next.ModelOverrides) {
		changed = true
	}
	if p.PiPackage != nil && p.PiPackage.Action == PackageInstall {
		changed = true
	}
	if p.Action == "remove" && hostsIncludePi(p.Hosts) && state.PiSubagentsSource != "" {
		changed = true
	}
	if !changed {
		return j, next, nil, false, nil
	}
	dirsToCreate, err := missingDirs(paths)
	if err != nil {
		return journal{}, State{}, nil, false, err
	}
	for _, d := range dirsToCreate {
		if !slices.Contains(next.CreatedDirs, d) {
			next.CreatedDirs = append(next.CreatedDirs, d)
		}
	}
	id := make([]byte, 16)
	if _, err = rand.Read(id); err != nil {
		return journal{}, State{}, nil, false, err
	}
	j.ID = hex.EncodeToString(id)
	if transactionID != "" {
		j.ID = transactionID
	}
	if p.Migration != nil {
		receipt := *p.Migration
		receipt.Transaction = j.ID
		var retained []MigrationReceipt
		for _, r := range next.Migrations {
			if r.Config != receipt.Config || !reflect.DeepEqual(r.Hosts, receipt.Hosts) {
				retained = append(retained, r)
			}
		}
		next.Migrations = append(retained, receipt)
	}
	j.AfterState = snapshot{Exists: true, Data: encode(next), Mode: 0600}
	return j, next, dirsToCreate, true, nil
}

// startTransaction durably records intent to run j before any resource
// write: the journal and pending marker Recover needs to reconcile a crash.
func startTransaction(p Plan, j journal, jp, pp string) error {
	for _, d := range []string{"transactions", "releases"} {
		path := filepath.Join(p.StateDir, d)
		if err := target.Safe(path); err != nil {
			return err
		}
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}
	if err := saveJournal(jp, j); err != nil {
		return err
	}
	return writeJSON(pp, pending{j.ID})
}

// commitTransaction performs j's actual writes and verification, updating
// the journal as it goes so Recover can tell exactly how far it got.
func (e Engine) commitTransaction(p Plan, j *journal, state State, next *State, dirsToCreate []string, jp, pp string) error {
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
		if err := saveJournal(jp, *j); err != nil {
			return err
		}
	}
	for i, en := range j.Entries {
		if same(en.Before, en.After) {
			continue
		}
		cur, err := readEntry(en)
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
		if err = writeEntry(en, cur, en.After, e.fail); err != nil {
			return err
		}
		if err = e.fail(fmt.Sprintf("write:%d", i)); err != nil {
			return err
		}
	}
	for _, en := range j.Entries {
		cur, err := readEntry(en)
		if err != nil {
			return err
		}
		if !same(cur, en.After) {
			return fmt.Errorf("post-write verification failed")
		}
	}
	if p.Migration != nil || len(p.Legacy) > 0 {
		remaining, err := scanLegacy(p.Config, p.Hosts, *next)
		if err != nil {
			return err
		}
		if remaining.Detected || len(remaining.Edits) > 0 {
			return fmt.Errorf("legacy retirement verification failed")
		}
	}
	if err := e.fail("pi-package"); err != nil {
		return err
	}
	if err := applyPiPackage(p, next, j, jp, e.fail); err != nil {
		return err
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
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	return os.Remove(pp)
}

// invertSubBlock reverts one marker-delimited sub-block within s from its
// forward "after" state back to its forward "before" state (forwardBefore,
// forwardAfter name the ORIGINAL forward change, not what to produce).
// forwardBefore == forwardAfter == nil means this sub-block was not part of
// the forward change at all: a no-op, s is returned unchanged. If s's
// sub-block already matches forwardBefore, it is returned unchanged too, so
// a prior interrupted recovery attempt that already reverted just this
// sub-block is resumed rather than rejected; recomputing outward from
// whatever s currently is (rather than requiring it to exactly equal the
// forward-after state) is what lets text a user wrote outside every managed
// block, at any point, survive untouched.
// insert, when non-nil, replaces transform's generic EOF-append for the one
// case that needs a specific position instead: reconstructing a block that
// the forward change removed (forwardBefore != nil, forwardAfter == nil).
// Hive's own reconstruction is a correct EOF-append (that mirrors how it
// was first installed), so only the voice call passes insertVoiceSpan.
func invertSubBlock(s snapshot, forwardBefore, forwardAfter *Record, m markers, insert func(snapshot, []byte) (snapshot, error)) (snapshot, error) {
	if forwardBefore == nil && forwardAfter == nil {
		return s, nil
	}
	if atBlockState(s, forwardBefore, m) {
		return s, nil
	}
	if !atBlockState(s, forwardAfter, m) {
		return snapshot{}, fmt.Errorf("unexpected %s block state", m.name)
	}
	if forwardAfter == nil && insert != nil {
		return insert(s, forwardBefore.Managed)
	}
	return transform(s, forwardAfter, forwardBefore, m)
}

// atBlockState reports whether s's marker-m block currently matches rec
// (rec == nil meaning "absent").
func atBlockState(s snapshot, rec *Record, m markers) bool {
	if rec == nil {
		a, _, err := blockRange(s.Data, m)
		return err == nil && a < 0
	}
	return owned(s, *rec, m) == nil
}

// recoverBlockAndVoice computes the inverse of one entry that carries a
// voice change (merged with a Hive Change, or standing alone with
// en.Change.Before == en.Change.After == nil). action is the plan's
// original Action, which fixes the mathematically correct inverse order:
// install composed Hive then voice forward, so its inverse undoes voice
// first, then Hive; remove and "voice" (set/off) composed voice then Hive
// forward, so their inverse undoes Hive first, then voice — the inverse of
// a composition always undoes its outermost (last-applied) step first,
// matching whichever order prepareEntries actually used for that action.
func recoverBlockAndVoice(action string, cur snapshot, en entry) (snapshot, error) {
	path := en.Change.Target.Path
	hive := func(s snapshot) (snapshot, error) {
		return invertSubBlock(s, en.Change.Before, en.Change.After, hiveMarkers, nil)
	}
	voice := func(s snapshot) (snapshot, error) {
		if en.Voice == nil {
			return s, nil
		}
		return invertSubBlock(s, voiceRecordFromSpan(path, en.Voice.Before), voiceRecordFromSpan(path, en.Voice.After), voiceMarkers, insertVoiceSpan)
	}
	if action == "install" {
		s, err := voice(cur)
		if err != nil {
			return snapshot{}, err
		}
		return hive(s)
	}
	s, err := hive(cur)
	if err != nil {
		return snapshot{}, err
	}
	return voice(s)
}

// finishApply prunes directories left empty by the committed transaction.
// This bookkeeping runs after commit succeeds, never inside it: it is not
// part of what Recover must undo.
func finishApply(p Plan, id string, next State) (string, error) {
	next.CreatedDirs = cleanupDirs(next.CreatedDirs)
	if err := writeJSON(filepath.Join(p.StateDir, "state.json"), next); err != nil {
		return id, fmt.Errorf("installed; directory bookkeeping failed: %w", err)
	}
	return id, nil
}

// recordSourceCommit appends p.SourceCommit to releases/<id>.commits.json
// when the plan carries both a source commit and a release and that commit
// is not already listed. The caller must still hold the manager lock, since
// this is otherwise an unguarded read-modify-write of a shared file. Apply's
// core outcome (install/remove, "unchanged" or committed) never depends on
// this: a write failure here is reported as a warning string, not an error,
// because the installation it describes is already committed.
func recordSourceCommit(p Plan) string {
	if p.SourceCommit == "" || p.Release == nil {
		return ""
	}
	dir := filepath.Join(p.StateDir, "releases")
	if err := target.Safe(dir); err != nil {
		return fmt.Sprintf("warning: recording source commit failed: %v", err)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Sprintf("warning: recording source commit failed: %v", err)
	}
	path := filepath.Join(dir, p.Release.ID+".commits.json")
	var rec commitRecord
	s, err := read(path)
	if err != nil {
		return fmt.Sprintf("warning: recording source commit failed: %v", err)
	}
	if s.Exists {
		if err := json.Unmarshal(s.Data, &rec); err != nil {
			return fmt.Sprintf("warning: recording source commit failed: %v", err)
		}
	}
	for _, c := range rec.Commits {
		if c.Commit == p.SourceCommit {
			return ""
		}
	}
	rec.Commits = append(rec.Commits, commitLogEntry{Commit: p.SourceCommit, AppliedAt: time.Now().UTC().Format(time.RFC3339)})
	if err := write(path, snapshot{Exists: true, Data: encode(rec), Mode: 0600}); err != nil {
		return fmt.Sprintf("warning: recording source commit failed: %v", err)
	}
	return ""
}
func (e Engine) Apply(p Plan) (string, error) {
	state := e.presetState
	if !e.nested {
		var (
			unlock func()
			err    error
		)
		state, unlock, err = preflight(p)
		if err != nil {
			return "", err
		}
		defer unlock()
	}
	j, next, dirsToCreate, changed, err := prepareTransaction(p, state, e.transactionID)
	if err != nil {
		return "", err
	}
	if !changed {
		if err := revalidatePiPackage(p); err != nil {
			return "", err
		}
		result := "unchanged"
		if warning := recordSourceCommit(p); warning != "" {
			result += "; " + warning
		}
		return result, nil
	}
	jp := filepath.Join(p.StateDir, "transactions", j.ID+".json")
	pp := filepath.Join(p.StateDir, "pending.json")
	if err = startTransaction(p, j, jp, pp); err != nil {
		return "", err
	}
	if err = e.commitTransaction(p, &j, state, &next, dirsToCreate, jp, pp); err != nil {
		return j.ID, fmt.Errorf("transaction %s requires recover: %w", j.ID, err)
	}
	warning := recordSourceCommit(p)
	id, err := finishApply(p, j.ID, next)
	if err != nil {
		if warning != "" {
			err = fmt.Errorf("%w; %s", err, warning)
		}
		return id, err
	}
	if warning != "" {
		id += "; " + warning
	}
	return id, nil
}
func (e Engine) Recover(stateDir string) (string, error) {
	dir, err := target.Canonical(stateDir)
	if err != nil {
		return "", err
	}
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		return "no_pending_operation", nil
	}
	if !e.nested {
		unlock, err := lock(dir)
		if err != nil {
			return "", err
		}
		defer unlock()
	}
	if !e.nested {
		if err := checkOnboarding(dir); err != nil {
			return "", err
		}
	}
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
	if (j.Version != 1 && j.Version != 2 && j.Version != 3 && j.Version != 4 && j.Version != 5 && j.Version != stateVersion) || j.Version != j.Plan.Version || j.ID != p.ID || j.Plan.StateDir != dir || j.Integrity != journalHash(j) || j.Plan.ID != planID(j.Plan) {
		return "", fmt.Errorf("invalid transaction")
	}
	removeOrphanWrites(dir, j)
	// Rollback restores Hive's own bytes from the journal.
	if j.Phase == "committed" || j.Phase == "recovered" {
		return j.Phase, os.Remove(pp)
	}
	if j.PackageDone || j.PackagePending {
		if err := undoPiPackage(j); err != nil {
			return "", err
		}
		j.PackageDone = false
		if err := saveJournal(jp, j); err != nil {
			return "", err
		}
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
		cur, err := readEntry(en)
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
		// The write completed exactly, with nothing further having touched
		// the file since: restore en.Before exactly, for a voice-carrying
		// entry just as for any other. This must run before the voice
		// block-by-block reconstruction below: that reconstruction composes
		// its result by finding and replacing each managed span in whatever
		// cur currently is, and an insert (a first-time managed span) always
		// appends at the position transform's generic insert branch uses,
		// which does not necessarily reproduce en.Before's original layout
		// (e.g. a removal that leaves trailing user text after the managed
		// spans would otherwise come back with that text first and the
		// spans re-appended after it), and does not preserve en.Before's
		// exact file mode when reconstructing from a fully deleted file.
		if same(cur, en.After) {
			inverses[i] = en.Before
			continue
		}
		if en.Voice != nil {
			if cur.Exists && en.Before.Exists && cur.Mode != en.Before.Mode {
				return "", fmt.Errorf("recovery conflict: %s; preserved", en.Change.Target.Path)
			}
			inv, err := recoverBlockAndVoice(j.Plan.Action, cur, en)
			if err != nil {
				return "", fmt.Errorf("recovery conflict: %s: %w", en.Change.Target.Path, err)
			}
			inverses[i] = inv
			continue
		}
		// A previous recovery attempt may already have restored the managed span
		// while retaining newer text outside it.
		if en.Change.Target.Kind == "block" && cur.Mode == en.Before.Mode && en.Change.Before != nil && owned(cur, *en.Change.Before, hiveMarkers) == nil {
			inverses[i] = cur
			continue
		}
		if en.Change.Target.Kind == "block" && en.Change.Before == nil && len(j.Plan.Legacy) == 0 {
			a, _, parseErr := blockRange(cur.Data, hiveMarkers)
			if parseErr == nil && a < 0 && cur.Exists {
				inverses[i] = cur
				continue
			}
		}
		if partialLegacyTree(cur, en.Before) {
			inverses[i] = en.Before
			continue
		}
		if legacyTouches(j.Plan, en.Change.Target.Path) {
			return "", fmt.Errorf("recovery conflict: migrated target changed: %s; preserved", en.Change.Target.Path)
		}
		if en.Change.Target.Kind != "block" || en.Change.After == nil || cur.Mode != en.After.Mode {
			return "", fmt.Errorf("recovery conflict: %s; preserved", en.Change.Target.Path)
		}
		inv, err := transform(cur, en.Change.After, en.Change.Before, hiveMarkers)
		if err != nil {
			return "", fmt.Errorf("recovery conflict: %w", err)
		}
		inverses[i] = inv
	}
	for i := len(j.Entries) - 1; i >= 0; i-- {
		en := j.Entries[i]
		cur, err := readEntry(en)
		if err != nil {
			return "", err
		}
		if !same(cur, currents[i]) {
			return "", fmt.Errorf("concurrent recovery change")
		}
		if !same(cur, inverses[i]) {
			if err = writeEntry(en, cur, inverses[i], nil); err != nil {
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
