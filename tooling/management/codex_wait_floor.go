package management

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"tricell-hive/integrations/codex"
)

// The Codex wait floor sets min_wait_timeout_ms and default_wait_timeout_ms in
// [features.multi_agent_v2] of Codex's config.toml, so a parent agent does not
// wake every 60 seconds. The file may hold secrets (for example in
// [mcp_servers.*.env]), so nothing here stores its bytes: the plan carries a
// fingerprint, the journal carries fingerprints and the lines Hive inserts, and
// every undo is a line-level inverse computed on the file as it is then.

const codexBackupSuffix = ".hive-backup"

func codexConfigPath(p Plan) string { return filepath.Join(p.Config.CodexHome, "config.toml") }

func hostsIncludeCodex(hosts []string) bool {
	for _, h := range hosts {
		if h == "codex" {
			return true
		}
	}
	return false
}

func codexWaitFloorKeys() []CodexWaitFloorKey {
	return []CodexWaitFloorKey{
		{Name: "min_wait_timeout_ms", Value: codex.WaitFloorMinMS},
		{Name: "default_wait_timeout_ms", Value: codex.WaitFloorDefaultMS},
	}
}

// planCodexWaitFloor adds the step when the option asks for it. It fails closed
// with the editor's fixed cause when the file cannot be edited safely.
func planCodexWaitFloor(p *Plan, o Options, state State) error {
	if !o.CodexWaitFloor {
		return nil
	}
	if p.Action != "install" || p.Config.Scope != "user" || !hostsIncludeCodex(p.Hosts) {
		return fmt.Errorf("codex wait floor: needs a user-scope install that includes codex")
	}
	// A record for another path would be replaced by this install and its lines
	// and backup orphaned; the remove side keeps such a record, so refuse here.
	if rec := state.CodexWaitFloor; rec != nil && rec.Path != codexConfigPath(*p) {
		return fmt.Errorf("codex wait floor: the state records lines at another config.toml; remove them there first")
	}
	cur, err := readCodexConfig(p.Config.CodexHome)
	if err != nil {
		return err
	}
	p.CodexWaitFloor = &CodexWaitFloorStep{Path: codexConfigPath(*p), Keys: codexWaitFloorKeys(), Expected: finger(cur)}
	return nil
}

// readCodexConfig reads config.toml after the editor's file-system checks and
// classifies it, so planning and applying refuse the same files.
func readCodexConfig(codexHome string) (snapshot, error) {
	if _, err := codex.CheckWaitFloorTarget(codexHome); err != nil {
		return snapshot{}, fmt.Errorf("codex wait floor: %w", err)
	}
	cur, err := read(filepath.Join(codexHome, "config.toml"))
	if err != nil {
		return snapshot{}, fmt.Errorf("codex wait floor: %w", err)
	}
	if _, _, err := codex.WaitFloor(cur.Data, cur.Exists); err != nil {
		return snapshot{}, fmt.Errorf("codex wait floor: %w", err)
	}
	return cur, nil
}

// linesEntry is the journal entry of a new kind: an edit of a file by lines.
// It holds fingerprints and Hive's own lines, never the file's bytes, so a
// committed journal (which is never deleted) cannot leak what config.toml holds.
type linesEntry struct {
	Path string
	// Remove is true for the inverse that plan remove runs.
	Remove bool `json:",omitempty"`
	// Before and After are the fingerprints of the file the edit was computed
	// from and of the file it must produce (install only).
	Before, After Fingerprint
	// Record is the lines this entry inserts (install) or the owned lines it
	// removes (remove).
	Record CodexWaitFloorRecord
	// BackupWanted is true when this entry creates config.toml.hive-backup.
	BackupWanted bool `json:",omitempty"`
	// Pending is set before the first write of the entry and Done after the
	// verified write; either one makes Recover undo the lines.
	Pending bool `json:",omitempty"`
	Done    bool `json:",omitempty"`
	// Skip is true when the state records a path other than the current Codex
	// config.toml (CODEX_HOME changed): the step does nothing but carry its note.
	Skip bool `json:",omitempty"`
	// Notes are fixed phrases naming keys Hive left in place, for the caller.
	Notes []string `json:",omitempty"`
}

func codexBackupPath(configPath string) string { return configPath + codexBackupSuffix }

// prepareCodexLines builds the journal entry and the next state record for the
// plan's wait floor step, or for the removal of what the state records.
func prepareCodexLines(p Plan, state State, next *State, j *journal) error {
	next.CodexWaitFloor = copyWaitFloorRecord(state.CodexWaitFloor)
	switch {
	case p.CodexWaitFloor != nil:
		step := p.CodexWaitFloor
		cur, err := readCodexConfig(p.Config.CodexHome)
		if err != nil {
			return err
		}
		if finger(cur) != step.Expected {
			return fmt.Errorf("stale target: %s", step.Path)
		}
		st, edit, err := codex.WaitFloor(cur.Data, cur.Exists)
		if err != nil {
			return fmt.Errorf("codex wait floor: %w", err)
		}
		if st == codex.WaitFloorApplied {
			return nil
		}
		mode := cur.Mode
		if !cur.Exists {
			mode = 0600
		}
		after := snapshot{Exists: true, Data: edit.Apply(cur.Data), Mode: mode}
		backup, err := read(codexBackupPath(step.Path))
		if err != nil {
			return err
		}
		le := &linesEntry{Path: step.Path, Before: finger(cur), After: finger(after), BackupWanted: cur.Exists && !backup.Exists}
		le.Record = CodexWaitFloorRecord{Path: step.Path, Lines: edit.Record.Lines, CreatedTable: edit.Record.CreatedTable,
			CreatedFile: edit.Record.CreatedFile, Separator: edit.Record.Separator, BackupCreated: le.BackupWanted}
		merged := le.Record
		if old := state.CodexWaitFloor; old != nil && old.Path == step.Path {
			merged.Lines = append(append([]string(nil), old.Lines...), merged.Lines...)
			merged.CreatedTable = merged.CreatedTable || old.CreatedTable
			merged.CreatedFile = merged.CreatedFile || old.CreatedFile
			merged.Separator = merged.Separator || old.Separator
			merged.BackupCreated = merged.BackupCreated || old.BackupCreated
		}
		next.CodexWaitFloor = &merged
		j.CodexLines = le
	case p.Action == "remove" && hostsIncludeCodex(p.Hosts) && state.CodexWaitFloor != nil:
		rec := *state.CodexWaitFloor
		if rec.Path != codexConfigPath(p) {
			// The lines live in a file this plan does not address. The rest of the
			// uninstall goes on and the record stays, so a later remove that finds
			// that path again can still undo exactly what Hive wrote there.
			j.CodexLines = &linesEntry{Path: rec.Path, Remove: true, Skip: true, Record: rec,
				Notes: []string{"Codex wait floor: not removed; the state records " + rec.Path + ", which is not the current Codex config.toml"}}
			return nil
		}
		j.CodexLines = &linesEntry{Path: rec.Path, Remove: true, Record: rec}
		next.CodexWaitFloor = nil
	}
	return nil
}

func copyWaitFloorRecord(r *CodexWaitFloorRecord) *CodexWaitFloorRecord {
	if r == nil {
		return nil
	}
	c := *r
	c.Lines = append([]string(nil), r.Lines...)
	return &c
}

// codexLinesUnchanged reports whether the wait floor part of p changes nothing.
func codexLinesUnchanged(p Plan, s State) (bool, error) {
	if p.CodexWaitFloor != nil {
		cur, err := readCodexConfig(p.Config.CodexHome)
		if err != nil {
			return false, err
		}
		if finger(cur) != p.CodexWaitFloor.Expected {
			return false, fmt.Errorf("stale target: %s", p.CodexWaitFloor.Path)
		}
		st, _, err := codex.WaitFloor(cur.Data, cur.Exists)
		return err == nil && st == codex.WaitFloorApplied, err
	}
	return !(p.Action == "remove" && hostsIncludeCodex(p.Hosts) && s.CodexWaitFloor != nil && s.CodexWaitFloor.Path == codexConfigPath(p)), nil
}

// validateCodexWaitFloor checks the shape of a (possibly saved) plan's step.
func validateCodexWaitFloor(p Plan) error {
	st := p.CodexWaitFloor
	if st == nil {
		return nil
	}
	if p.Action != "install" || p.Config.Scope != "user" || !hostsIncludeCodex(p.Hosts) {
		return fmt.Errorf("invalid codex wait floor step")
	}
	if st.Path != codexConfigPath(p) || !reflect.DeepEqual(st.Keys, codexWaitFloorKeys()) {
		return fmt.Errorf("invalid codex wait floor step")
	}
	return nil
}

// applyCodexLines performs the journaled write (install) or the line-level
// inverse (remove) after the entries are written and before the state is.
func applyCodexLines(j *journal, jp string, fail func(string) error) error {
	le := j.CodexLines
	if le == nil || le.Skip {
		return nil
	}
	if le.Remove {
		return removeCodexLines(j, jp, fail)
	}
	cur, err := read(le.Path)
	if err != nil {
		return err
	}
	if finger(cur) != le.Before {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	st, edit, err := codex.WaitFloor(cur.Data, cur.Exists)
	if err != nil || st == codex.WaitFloorApplied {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	mode := cur.Mode
	if !cur.Exists {
		mode = 0600
	}
	after := snapshot{Exists: true, Data: edit.Apply(cur.Data), Mode: mode}
	if finger(after) != le.After {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	le.Pending = true
	if err = saveJournal(jp, *j); err != nil {
		return err
	}
	if le.BackupWanted {
		bk := codexBackupPath(le.Path)
		if old, err := read(bk); err != nil {
			return err
		} else if old.Exists {
			return fmt.Errorf("concurrent change: %s", bk)
		}
		if err = write(bk, snapshot{Exists: true, Data: cur.Data, Mode: 0600}); err != nil {
			return err
		}
		if fail != nil {
			if err = fail("codex-lines-backup"); err != nil {
				return err
			}
		}
	}
	if err = os.MkdirAll(filepath.Dir(le.Path), 0700); err != nil {
		return err
	}
	if err = write(le.Path, after); err != nil {
		return err
	}
	if fail != nil {
		if err = fail("codex-lines-write"); err != nil {
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
		return fail("codex-lines-after")
	}
	return nil
}

// undoCodexLinesOnFile removes the owned lines from the file as it is now. It
// reports what it left; a changed or rewritten file is never an error.
func undoCodexLinesOnFile(path string, rec CodexWaitFloorRecord) (codex.WaitFloorUndo, error) {
	cur, err := read(path)
	if err != nil {
		return codex.WaitFloorUndo{}, err
	}
	if !cur.Exists {
		gone := codex.WaitFloorUndo{Skipped: "config.toml is gone"}
		for _, l := range rec.Lines {
			if k, _, ok := strings.Cut(l, "="); ok && !strings.HasPrefix(strings.TrimSpace(l), "[") {
				gone.Kept = append(gone.Kept, strings.TrimSpace(k))
			}
		}
		return gone, nil
	}
	res := codex.UndoWaitFloor(cur.Data, rec.toEditor())
	switch {
	case res.DeleteFile:
		err = write(path, snapshot{})
	case !bytes.Equal(res.Content, cur.Data):
		err = write(path, snapshot{Exists: true, Data: res.Content, Mode: cur.Mode})
	}
	return res, err
}

func (r CodexWaitFloorRecord) toEditor() codex.WaitFloorRecord {
	return codex.WaitFloorRecord{Lines: r.Lines, CreatedTable: r.CreatedTable, CreatedFile: r.CreatedFile, Separator: r.Separator}
}

func removeCodexLines(j *journal, jp string, fail func(string) error) error {
	le := j.CodexLines
	le.Pending = true
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		if err := fail("codex-lines-pending"); err != nil {
			return err
		}
	}
	res, err := undoCodexLinesOnFile(le.Path, le.Record)
	if err != nil {
		return err
	}
	if fail != nil {
		if err = fail("codex-lines-write"); err != nil {
			return err
		}
	}
	// The backup goes only when Hive itself removed every registered line in this
	// run (or the file it created). Whatever else happened, the editor's own
	// result classifies it: Kept covers a value the user changed, a line that was
	// already gone, a rewrite in another form and a file it can no longer read.
	for _, k := range res.Kept {
		le.Notes = append(le.Notes, "Codex wait floor: left "+k+"; its line was changed or was already gone")
	}
	if clean := len(res.Kept) == 0 && res.Skipped == ""; clean {
		if le.Record.BackupCreated {
			if err = write(codexBackupPath(le.Path), snapshot{}); err != nil {
				return err
			}
		}
	} else if le.Record.BackupCreated {
		if bk, err := read(codexBackupPath(le.Path)); err != nil {
			return err
		} else if bk.Exists {
			le.Notes = append(le.Notes, "Codex wait floor: kept "+filepath.Base(codexBackupPath(le.Path)))
		}
	}
	le.Pending, le.Done = false, true
	return saveJournal(jp, *j)
}

// undoCodexInstall rolls back an interrupted install's wait floor write: it
// removes the entry's lines wherever they are in the current file, so an edit
// Codex made elsewhere survives, and the backup this entry created.
func undoCodexInstall(j journal) error {
	le := j.CodexLines
	if le == nil || le.Remove || !(le.Pending || le.Done) {
		return nil
	}
	res, err := undoCodexLinesOnFile(le.Path, le.Record)
	if err != nil {
		return err
	}
	if le.BackupWanted {
		// The write never happened (or the file is back to what it was) when the
		// file equals its journaled Before image; the lines are then not "kept".
		now, err := read(le.Path)
		if err != nil {
			return err
		}
		// Delete only when Hive removed every line or the write never happened;
		// never when the undo skipped, because the lines may still be there.
		if !((len(res.Kept) == 0 && res.Skipped == "") || finger(now) == le.Before) {
			return nil
		}
		bk := codexBackupPath(le.Path)
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

// StoredCodexWaitFloor returns, read-only, the record of the lines Hive wrote
// into Codex's config.toml for the state of o, or nil when Hive wrote none. It
// is a copy the caller may change. The accessor does not touch config.toml.
func StoredCodexWaitFloor(o Options) (*CodexWaitFloorRecord, error) {
	o.Scope = "user"
	_, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	state, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	return copyWaitFloorRecord(state.CodexWaitFloor), nil
}

// withoutCodexRecord returns the snapshot of a state file with its Codex wait
// floor record dropped, or the snapshot itself when it has none.
func withoutCodexRecord(s snapshot) (snapshot, error) {
	var st State
	if len(s.Data) == 0 {
		return s, nil
	}
	if err := json.Unmarshal(s.Data, &st); err != nil {
		return snapshot{}, err
	}
	if st.CodexWaitFloor == nil {
		return s, nil
	}
	st.CodexWaitFloor = nil
	s.Data = encode(st)
	return s, nil
}

// codexRemoveRecordGone reports whether the state may drop the record of an
// interrupted removal: Hive's lines are no longer in config.toml (judged by the
// editor's own preview, without writing) and no backup Hive created is left for
// a later remove to own. When it cannot tell, it keeps the record.
func codexRemoveRecordGone(le *linesEntry) (bool, error) {
	cur, err := read(le.Path)
	if err != nil {
		return false, err
	}
	if cur.Exists {
		preview := codex.UndoWaitFloor(cur.Data, le.Record.toEditor())
		if len(preview.Removed) > 0 || preview.Skipped == codex.WaitFloorSkippedUnreadable {
			return false, nil
		}
	}
	if le.Record.BackupCreated {
		bk, err := read(codexBackupPath(le.Path))
		if err != nil {
			return false, err
		}
		if bk.Exists {
			return false, nil
		}
	}
	return true, nil
}
