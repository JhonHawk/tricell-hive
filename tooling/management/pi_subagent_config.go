package management

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"tricell-hive/integrations/pi"
	"tricell-hive/integrations/target"
)

// The Pi subagent activation sets "toolActivation": "eager" in pi-subagents'
// own config.json under the Pi home, so the subagent tool is active from a
// session's first request. Hive owns that one key only, and only when it wrote
// it. The file is the user's: the plan carries a fingerprint, the journal
// carries fingerprints and flags, and every undo is a key-level inverse
// computed on the file as it is then. The only file bytes a journal may hold
// are those of an empty JSON object, which is whitespace and braces.

const (
	piConfigBackupSuffix = ".hive-backup"
	piConfigEagerJSON    = `"` + pi.EagerActivation + `"`
	piConfigUnsupported  = "the pi-subagents config is not a JSON object with unique keys"
	piConfigShownRunes   = 64
	piConfigBeforeLimit  = 4096
)

func piConfigPath(p Plan) string {
	return filepath.Join(p.Config.PiHome, filepath.FromSlash(pi.SubagentConfigRel))
}

func piConfigBackupPath(path string) string { return path + piConfigBackupSuffix }

// readPiConfig reads config.json as it is on disk after the integration's
// file-system checks, so planning and applying refuse the same files.
func readPiConfig(piHome string) (snapshot, error) {
	if _, err := pi.CheckSubagentConfigTarget(piHome); err != nil {
		return snapshot{}, err
	}
	return read(filepath.Join(piHome, filepath.FromSlash(pi.SubagentConfigRel)))
}

// piConfigAfterMigration is the file as the plan's legacy migration leaves it,
// or as it is on disk when the migration does not edit it.
func piConfigAfterMigration(p Plan) (snapshot, error) {
	path := piConfigPath(p)
	disk, err := readPiConfig(p.Config.PiHome)
	if err != nil {
		return snapshot{}, err
	}
	for _, edit := range p.Legacy {
		if edit.Path == path {
			return legacyAfter(edit), nil
		}
	}
	return disk, nil
}

func shownPiValue(v string) string {
	var b strings.Builder
	n := 0
	for _, r := range v {
		if n >= piConfigShownRunes {
			b.WriteString("...")
			break
		}
		if unicode.IsControl(r) {
			r = ' '
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

// planPiSubagentConfig classifies the key and adds the step. Release updates
// never plan it. It fails only on a file-system fault; a file Hive must not
// edit becomes a manual step.
func planPiSubagentConfig(p *Plan, state State) error {
	if p.ReleaseUpdate {
		p.PiSubagentConfig = nil
		return nil
	}
	if p.Action != "install" || p.Config.Scope != "user" || !hostsIncludePi(p.Hosts) {
		return nil
	}
	path := piConfigPath(*p)
	step := &PiSubagentConfigStep{Path: path, Key: pi.ToolActivationKey, Value: piConfigEagerJSON}
	p.PiSubagentConfig = step
	rec := state.PiSubagentConfig
	if rec != nil && rec.Path != path {
		step.Status, step.Detail = PiConfigOtherHome, rec.Path
		return nil
	}
	cur, err := piConfigAfterMigration(*p)
	if err != nil {
		var refusal *pi.SubagentConfigRefusal
		if !errors.As(err, &refusal) {
			return fmt.Errorf("pi subagent activation: %w", err)
		}
		step.Status, step.Detail = PiConfigManual, refusal.Cause
		return nil
	}
	status, value := pi.ClassifySubagentConfig(cur.Data, cur.Exists)
	switch status {
	case pi.ConfigUnsupported:
		step.Status, step.Detail = PiConfigManual, piConfigUnsupported
	case pi.KeyPresent:
		if pi.IsEagerValue(value) && rec != nil && rec.Value == piConfigEagerJSON {
			step.Status = PiConfigApplied
		} else {
			step.Status, step.Value = PiConfigUserSet, shownPiValue(value)
		}
	default:
		step.Status, step.Expected = PiConfigWrite, finger(cur)
	}
	return nil
}

// validatePiSubagentConfig checks the shape of a (possibly saved) plan's step.
func validatePiSubagentConfig(p Plan) error {
	st := p.PiSubagentConfig
	if st == nil {
		return nil
	}
	if p.ReleaseUpdate || p.Action != "install" || p.Config.Scope != "user" || !hostsIncludePi(p.Hosts) {
		return fmt.Errorf("invalid pi subagent config step")
	}
	if st.Path != piConfigPath(p) || st.Key != pi.ToolActivationKey {
		return fmt.Errorf("invalid pi subagent config step")
	}
	switch st.Status {
	case PiConfigWrite, PiConfigApplied:
		if st.Value != piConfigEagerJSON {
			return fmt.Errorf("invalid pi subagent config step")
		}
	case PiConfigManual:
		if st.Value != piConfigEagerJSON || st.Detail == "" {
			return fmt.Errorf("invalid pi subagent config step")
		}
	case PiConfigOtherHome:
		if st.Detail == "" || st.Detail == st.Path {
			return fmt.Errorf("invalid pi subagent config step")
		}
	case PiConfigUserSet:
	default:
		return fmt.Errorf("invalid pi subagent config step")
	}
	if st.Status != PiConfigWrite && st.Expected != (Fingerprint{}) {
		return fmt.Errorf("invalid pi subagent config step")
	}
	return nil
}

// piConfigEntry is the journal entry of the toolActivation edit.
type piConfigEntry struct {
	Path string
	// Remove is true for the inverse that plan remove runs.
	Remove bool `json:",omitempty"`
	// Before and After are the fingerprints of the file the edit was computed
	// from and of the file it must produce (install only).
	Before, After Fingerprint
	// Record is what this entry creates (install) or what it may remove
	// (remove).
	Record PiSubagentConfigRecord
	// BackupWanted is true when this entry creates config.json.hive-backup.
	BackupWanted bool `json:",omitempty"`
	// BeforeEmpty holds the bytes of the file when it was an empty JSON object,
	// so a recovery can give them back exactly. It never holds anything else.
	BeforeEmpty []byte `json:",omitempty"`
	// Pending is set before the first write of the entry and Done after the
	// verified write; either one makes Recover undo the entry.
	Pending bool `json:",omitempty"`
	Done    bool `json:",omitempty"`
	// Skip is true when the state records a path other than the current
	// config.json (the Pi home changed): the entry does nothing but carry its note.
	Skip bool `json:",omitempty"`
	// Notes are fixed phrases naming what Hive left in place, for the caller.
	Notes []string `json:",omitempty"`
}

func copyPiRecord(r *PiSubagentConfigRecord) *PiSubagentConfigRecord {
	if r == nil {
		return nil
	}
	c := *r
	c.CreatedDirs = append([]string(nil), r.CreatedDirs...)
	return &c
}

// preparePiConfig builds the journal entry and the next state record for the
// plan's step, or for the removal of what the state records.
func preparePiConfig(p Plan, state State, next *State, j *journal) error {
	next.PiSubagentConfig = copyPiRecord(state.PiSubagentConfig)
	switch {
	case p.PiSubagentConfig != nil && p.PiSubagentConfig.Status == PiConfigWrite:
		step := p.PiSubagentConfig
		cur, err := piConfigAfterMigration(p)
		if err != nil {
			return fmt.Errorf("pi subagent activation: %w", err)
		}
		if finger(cur) != step.Expected {
			return fmt.Errorf("stale target: %s", step.Path)
		}
		st, _ := pi.ClassifySubagentConfig(cur.Data, cur.Exists)
		if st != pi.ConfigAbsent && st != pi.KeyMissing {
			return fmt.Errorf("stale target: %s", step.Path)
		}
		data, err := pi.InsertToolActivation(cur.Data)
		if err != nil {
			return fmt.Errorf("pi subagent activation: %w", err)
		}
		mode := cur.Mode
		if !cur.Exists {
			mode = 0600
		}
		after := snapshot{Exists: true, Data: data, Mode: mode}
		backup, err := read(piConfigBackupPath(step.Path))
		if err != nil {
			return err
		}
		le := &piConfigEntry{Path: step.Path, Before: finger(cur), After: finger(after), BackupWanted: cur.Exists && !backup.Exists}
		if cur.Exists && pi.SubagentConfigEmpty(cur.Data) && len(cur.Data) <= piConfigBeforeLimit {
			le.BeforeEmpty = append([]byte(nil), cur.Data...)
		}
		le.Record = PiSubagentConfigRecord{Path: step.Path, Value: piConfigEagerJSON, CreatedFile: !cur.Exists, BackupCreated: le.BackupWanted}
		merged := le.Record
		if old := state.PiSubagentConfig; old != nil && old.Path == step.Path {
			merged.CreatedFile = merged.CreatedFile || old.CreatedFile
			merged.BackupCreated = merged.BackupCreated || old.BackupCreated
			merged.CreatedDirs = append([]string(nil), old.CreatedDirs...)
		}
		next.PiSubagentConfig = &merged
		j.PiConfig = le
	case p.Action == "remove" && hostsIncludePi(p.Hosts) && state.PiSubagentConfig != nil:
		rec := *copyPiRecord(state.PiSubagentConfig)
		if rec.Path != piConfigPath(p) {
			// The key lives in a file this plan does not address. The rest of the
			// uninstall goes on and the record stays, so a later remove that finds
			// that path again can still undo exactly what Hive wrote there.
			j.PiConfig = &piConfigEntry{Path: rec.Path, Remove: true, Skip: true, Record: rec,
				Notes: []string{"Pi subagent activation: not removed; the state records " + rec.Path + ", which is not the current pi-subagents config.json"}}
			return nil
		}
		j.PiConfig = &piConfigEntry{Path: rec.Path, Remove: true, Record: rec}
		next.PiSubagentConfig = nil
	}
	return nil
}

// finishPiConfigDirs records, once the transaction knows which folders it
// creates itself, the ones below the Pi home that only this step creates.
func finishPiConfigDirs(p Plan, next *State, j *journal, dirsToCreate []string) error {
	le := j.PiConfig
	if le == nil || le.Remove || le.Skip {
		return nil
	}
	missing, err := missingDirs([]string{le.Path})
	if err != nil {
		return err
	}
	for _, d := range missing {
		if strings.HasPrefix(d, p.Config.PiHome+string(os.PathSeparator)) && !slices.Contains(dirsToCreate, d) {
			le.Record.CreatedDirs = append(le.Record.CreatedDirs, d)
			if !slices.Contains(next.PiSubagentConfig.CreatedDirs, d) {
				next.PiSubagentConfig.CreatedDirs = append(next.PiSubagentConfig.CreatedDirs, d)
			}
		}
	}
	return nil
}

// piConfigUnchanged reports whether the toolActivation part of p changes nothing.
func piConfigUnchanged(p Plan, s State) (bool, error) {
	if st := p.PiSubagentConfig; st != nil && st.Status == PiConfigWrite {
		cur, err := piConfigAfterMigration(p)
		if err != nil {
			return false, fmt.Errorf("pi subagent activation: %w", err)
		}
		if finger(cur) != st.Expected {
			return false, fmt.Errorf("stale target: %s", st.Path)
		}
		return false, nil
	}
	return !(p.Action == "remove" && hostsIncludePi(p.Hosts) && s.PiSubagentConfig != nil && s.PiSubagentConfig.Path == piConfigPath(p)), nil
}

// applyPiConfig performs the journaled write (install) or the key-level
// inverse (remove) after the entries are written and before the state is.
func applyPiConfig(j *journal, jp string, fail func(string) error) error {
	le := j.PiConfig
	if le == nil || le.Skip {
		return nil
	}
	if le.Remove {
		return removePiConfig(j, jp, fail)
	}
	piHome := j.Plan.Config.PiHome
	cur, err := readPiConfig(piHome)
	if err != nil {
		return fmt.Errorf("pi subagent activation: %w", err)
	}
	if finger(cur) != le.Before {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	data, err := pi.InsertToolActivation(cur.Data)
	if err != nil {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	mode := cur.Mode
	if !cur.Exists {
		mode = 0600
	}
	after := snapshot{Exists: true, Data: data, Mode: mode}
	if finger(after) != le.After {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	le.Pending = true
	if err = saveJournal(jp, *j); err != nil {
		return err
	}
	if le.BackupWanted {
		bk := piConfigBackupPath(le.Path)
		if old, err := read(bk); err != nil {
			return err
		} else if old.Exists {
			return fmt.Errorf("concurrent change: %s", bk)
		}
		if err = write(bk, snapshot{Exists: true, Data: cur.Data, Mode: 0600}); err != nil {
			return err
		}
		if fail != nil {
			if err = fail("pi-config-backup"); err != nil {
				return err
			}
		}
	}
	if err = os.MkdirAll(piHome, 0700); err != nil {
		return err
	}
	for _, d := range le.Record.CreatedDirs {
		if err = target.Safe(d); err != nil {
			return err
		}
		if err = os.Mkdir(d, 0700); err != nil && !os.IsExist(err) {
			return err
		}
	}
	if err = write(le.Path, after); err != nil {
		return err
	}
	if fail != nil {
		if err = fail("pi-config-write"); err != nil {
			return err
		}
	}
	got, err := read(le.Path)
	if err != nil {
		return err
	}
	if finger(got) != le.After {
		return fmt.Errorf("post-write verification failed: %s", le.Path)
	}
	le.Pending, le.Done = false, true
	if err = saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		return fail("pi-config-after")
	}
	return nil
}

// removeCreatedPiDirs removes, deepest first, the folders the record says Hive
// created, only while they are empty.
func removeCreatedPiDirs(dirs []string) {
	sorted := append([]string(nil), dirs...)
	slices.SortFunc(sorted, func(a, b string) int { return len(b) - len(a) })
	for _, d := range sorted {
		if target.Safe(d) == nil {
			_ = os.Remove(d)
		}
	}
}

func removePiConfig(j *journal, jp string, fail func(string) error) error {
	le := j.PiConfig
	le.Pending = true
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		if err := fail("pi-config-pending"); err != nil {
			return err
		}
	}
	name := filepath.Base(le.Path)
	removed := false
	cur, err := readPiConfig(j.Plan.Config.PiHome)
	var refusal *pi.SubagentConfigRefusal
	if errors.As(err, &refusal) {
		le.Notes = append(le.Notes, "Pi subagent activation: left toolActivation; "+refusal.Cause)
		err = nil
	} else if err != nil {
		return err
	}
	switch {
	case refusal != nil:
		// Nothing is read or written: the file cannot be edited safely.
	case !cur.Exists:
		le.Notes = append(le.Notes, "Pi subagent activation: left toolActivation; "+name+" is gone")
		removeCreatedPiDirs(le.Record.CreatedDirs)
	default:
		st, value := pi.ClassifySubagentConfig(cur.Data, true)
		switch {
		case st == pi.KeyPresent && pi.IsEagerValue(value) && le.Record.Value == piConfigEagerJSON:
			out, _, err := pi.RemoveToolActivation(cur.Data)
			if err != nil {
				return err
			}
			if le.Record.CreatedFile && pi.SubagentConfigEmpty(out) {
				err = write(le.Path, snapshot{})
			} else {
				err = write(le.Path, snapshot{Exists: true, Data: out, Mode: cur.Mode})
			}
			if err != nil {
				return err
			}
			removed = true
			if fail != nil {
				if err = fail("pi-config-write"); err != nil {
					return err
				}
			}
			if le.Record.CreatedFile && pi.SubagentConfigEmpty(out) {
				removeCreatedPiDirs(le.Record.CreatedDirs)
			} else if le.Record.CreatedFile {
				le.Notes = append(le.Notes, "Pi subagent activation: left "+name+"; it holds other settings")
			} else if pi.SubagentConfigEmpty(out) {
				le.Notes = append(le.Notes, "Pi subagent activation: left "+name+"; it is now an empty object")
			}
		case st == pi.KeyPresent:
			le.Notes = append(le.Notes, "Pi subagent activation: left toolActivation; its value was changed after Hive wrote it")
		default:
			le.Notes = append(le.Notes, "Pi subagent activation: left toolActivation; it was already gone or "+name+" can no longer be read")
		}
	}
	bk := piConfigBackupPath(le.Path)
	if le.Record.BackupCreated {
		if removed {
			if err = write(bk, snapshot{}); err != nil {
				return err
			}
		} else if b, err := read(bk); err != nil {
			return err
		} else if b.Exists {
			le.Notes = append(le.Notes, "Pi subagent activation: kept "+filepath.Base(bk))
		}
	}
	le.Pending, le.Done = false, true
	return saveJournal(jp, *j)
}

// undoneInstall is the file once an interrupted install's key is taken out: the
// bytes the key-level inverse leaves, or no file when Hive created it and
// nothing else is in it. removed reports whether a key was taken out.
func undoneInstall(le *piConfigEntry, cur snapshot) (after snapshot, removed bool, err error) {
	if !cur.Exists {
		return cur, false, nil
	}
	st, value := pi.ClassifySubagentConfig(cur.Data, true)
	if st != pi.KeyPresent || !pi.IsEagerValue(value) {
		return cur, false, nil
	}
	out, _, err := pi.RemoveToolActivation(cur.Data)
	if err != nil {
		return cur, false, err
	}
	switch {
	case le.Record.CreatedFile && pi.SubagentConfigEmpty(out):
		return snapshot{}, true, nil
	case le.BeforeEmpty != nil && pi.SubagentConfigEmpty(out):
		return snapshot{Exists: true, Data: le.BeforeEmpty, Mode: cur.Mode}, true, nil
	}
	return snapshot{Exists: true, Data: out, Mode: cur.Mode}, true, nil
}

// piConfigAfterUndo is what the journal entry at path will read as once
// undoPiConfigInstall has run, so Recover's preflight of the other entries (the
// legacy migration may have edited this very file) judges the file it will
// find. It writes nothing.
func piConfigAfterUndo(j journal, path string, cur snapshot) (snapshot, error) {
	le := j.PiConfig
	if le == nil || le.Remove || !(le.Pending || le.Done) || le.Path != path {
		return cur, nil
	}
	after, _, err := undoneInstall(le, cur)
	return after, err
}

// undoPiConfigInstall rolls back an interrupted install's write: it removes
// the key wherever it is in the current file, so an edit Pi or the user made
// elsewhere survives, the file and folders the entry created, and the backup.
func undoPiConfigInstall(j journal) error {
	le := j.PiConfig
	if le == nil || le.Remove || !(le.Pending || le.Done) {
		return nil
	}
	cur, err := read(le.Path)
	if err != nil {
		return err
	}
	after, removed, err := undoneInstall(le, cur)
	if err != nil {
		return err
	}
	if removed {
		if err = write(le.Path, after); err != nil {
			return err
		}
	}
	if !after.Exists {
		removeCreatedPiDirs(le.Record.CreatedDirs)
	}
	if le.BackupWanted && (removed || finger(after) == le.Before) {
		bk := piConfigBackupPath(le.Path)
		b, err := read(bk)
		if err != nil {
			return err
		}
		if b.Exists && hash(b.Data) == le.Before.Hash {
			return write(bk, snapshot{})
		}
	}
	return nil
}

// restorePiConfigRemoval rolls back an interrupted removal: it puts back the
// key it deleted, and the file and folders it deleted when Hive had created
// them, without rewriting anything else. The backup it may have deleted is not
// restored: the journal holds no file bytes.
func restorePiConfigRemoval(j journal) error {
	le := j.PiConfig
	if le == nil || !le.Remove || le.Skip || !(le.Pending || le.Done) {
		return nil
	}
	cur, err := read(le.Path)
	if err != nil {
		return err
	}
	if !cur.Exists {
		if !le.Record.CreatedFile {
			return nil
		}
		if err = os.MkdirAll(j.Plan.Config.PiHome, 0700); err != nil {
			return err
		}
		dirs := append([]string(nil), le.Record.CreatedDirs...)
		slices.SortFunc(dirs, func(a, b string) int { return len(a) - len(b) })
		for _, d := range dirs {
			if err = target.Safe(d); err != nil {
				return err
			}
			if err = os.Mkdir(d, 0700); err != nil && !os.IsExist(err) {
				return err
			}
		}
		data, err := pi.InsertToolActivation(nil)
		if err != nil {
			return err
		}
		return write(le.Path, snapshot{Exists: true, Data: data, Mode: 0600})
	}
	if st, _ := pi.ClassifySubagentConfig(cur.Data, true); st != pi.KeyMissing {
		return nil
	}
	data, err := pi.InsertToolActivation(cur.Data)
	if err != nil {
		return err
	}
	return write(le.Path, snapshot{Exists: true, Data: data, Mode: cur.Mode})
}

// StoredPiSubagentConfig returns, read-only, the record of the toolActivation
// key Hive wrote into pi-subagents' config.json for the state of o, or nil when
// Hive wrote none. It is a copy the caller may change. The accessor does not
// touch config.json.
func StoredPiSubagentConfig(o Options) (*PiSubagentConfigRecord, error) {
	o.Scope = "user"
	_, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	return copyPiRecord(state.PiSubagentConfig), nil
}
