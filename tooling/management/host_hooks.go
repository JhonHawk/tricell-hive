package management

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	"tricell-hive/integrations/claude"
	"tricell-hive/integrations/codex"
	"tricell-hive/integrations/cursor"
	"tricell-hive/integrations/hostevent"
	"tricell-hive/integrations/target"
)

// Host hooks: a user-scope install registers `hive hook <host> <event>` in the
// shared JSON hook files of Claude Code, Codex and Cursor, through the editor
// in json_hooks.go. The files may hold other tools' hooks and, in Claude Code's
// settings.json, secrets in "env", so nothing here stores their bytes: the plan
// carries a fingerprint, the journal carries fingerprints and Hive's own
// entries, and every undo is an entry-level inverse computed on the file as it
// is then. Grok's second hook file, OpenCode's plugin and Pi's extension are
// whole files Hive owns, so they are records of kind "hook" (generated_hooks.go).

// hostHookSpec returns the file Hive registers hooks in for a manager host.
func hostHookSpec(host string, c target.Config) (hostevent.JSONHookFile, bool) {
	switch host {
	case "claude":
		return claude.SessionEventsFile(c), true
	case "codex":
		return codex.SessionEventsFile(c), true
	case "cursor":
		return cursor.SessionEventsFile(c), true
	}
	return hostevent.JSONHookFile{}, false
}

const maxHookFileBytes = 4 << 20

// readHookFile reads a shared hook file after the file-system checks every
// managed read makes, and refuses one too large to be a hook configuration.
func readHookFile(path string) (snapshot, error) {
	s, err := read(path)
	if err != nil {
		return snapshot{}, err
	}
	if len(s.Data) > maxHookFileBytes {
		return snapshot{}, fmt.Errorf("%s is larger than %d MiB", path, maxHookFileBytes>>20)
	}
	return s, nil
}

func priorHostHooks(state State, path string) *HostHooksRecord {
	if r, ok := state.HostHooks[path]; ok {
		c := copyHostHooksRecord(r)
		return &c
	}
	return nil
}

// planHostHooks adds a step for each selected host that has a shared hook file.
// It fails closed, naming the file and the cause, when a file cannot be edited
// safely, so a plan never promises an edit apply would refuse.
func planHostHooks(p *Plan, state State) error {
	if p.Action != "install" || p.Config.Scope != "user" {
		return nil
	}
	for _, h := range p.Hosts {
		spec, ok := hostHookSpec(h, p.Config)
		if !ok {
			continue
		}
		cur, err := readHookFile(spec.Path)
		if err != nil {
			return fmt.Errorf("%s hooks: %w", h, err)
		}
		if _, err = editJSONHooks(cur.Data, cur.Exists, spec, p.Executable, priorHostHooks(state, spec.Path)); err != nil {
			return fmt.Errorf("%s hooks: %s: %w; fix the file or add the Hive hooks by hand", h, spec.Path, err)
		}
		p.HostHooks = append(p.HostHooks, HostHooksStep{Host: h, Path: spec.Path, Expected: finger(cur)})
	}
	return nil
}

// validateHostHooks checks the shape of a (possibly saved) plan's steps against
// what the plan's hosts and configuration resolve to.
func validateHostHooks(p Plan) error {
	var want []HostHooksStep
	if p.Action == "install" && p.Config.Scope == "user" {
		for _, h := range p.Hosts {
			if spec, ok := hostHookSpec(h, p.Config); ok {
				want = append(want, HostHooksStep{Host: h, Path: spec.Path})
			}
		}
	}
	if len(want) != len(p.HostHooks) {
		return fmt.Errorf("invalid host hooks step")
	}
	for i, st := range p.HostHooks {
		if st.Host != want[i].Host || st.Path != want[i].Path {
			return fmt.Errorf("invalid host hooks step")
		}
	}
	return nil
}

// validateHostHooksState checks the records of a state read from disk.
func validateHostHooksState(m map[string]HostHooksRecord) error {
	for path, r := range m {
		if path != r.Path || !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return fmt.Errorf("invalid recorded host hooks path")
		}
		if r.Host != "claude" && r.Host != "codex" && r.Host != "cursor" {
			return fmt.Errorf("invalid recorded host hooks host")
		}
	}
	return nil
}

// hostHooksEntry is the journal entry of one shared hook file.
type hostHooksEntry struct {
	Host, Path string
	// Remove is true for the inverse that plan remove runs.
	Remove bool `json:",omitempty"`
	// Before and After are the fingerprints of the file the edit was computed
	// from and of the file it must produce (install only).
	Before, After Fingerprint
	// Record is what an install writes (the entries it adds and what it creates)
	// or what a remove deletes (everything the state recorded).
	Record HostHooksRecord
	// Known are the commands the state recorded before an install, so a moved
	// binary's old entries are still replaced.
	Known []string `json:",omitempty"`
	// BackupWanted is true when this entry creates the file's backup.
	BackupWanted bool `json:",omitempty"`
	// Pending is set before the first write of the entry and Done after the
	// verified write; either one makes Recover undo an install.
	Pending bool `json:",omitempty"`
	Done    bool `json:",omitempty"`
	// Notes are phrases for the caller's report.
	Notes []string `json:",omitempty"`
}

func copyHostHooksRecord(r HostHooksRecord) HostHooksRecord {
	r.Commands = append([]string(nil), r.Commands...)
	r.CreatedEvents = append([]string(nil), r.CreatedEvents...)
	return r
}

func copyHostHooks(m map[string]HostHooksRecord) map[string]HostHooksRecord {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]HostHooksRecord, len(m))
	for k, v := range m {
		out[k] = copyHostHooksRecord(v)
	}
	return out
}

// mergeHostHooks is the record after an edit: the commands the file now holds
// for Hive and everything Hive created, now or before.
func mergeHostHooks(prior *HostHooksRecord, delta HostHooksRecord, host string) HostHooksRecord {
	out := copyHostHooksRecord(delta)
	out.Host = host
	sort.Strings(out.Commands)
	if prior != nil {
		out.CreatedFile = out.CreatedFile || prior.CreatedFile
		out.CreatedHooks = out.CreatedHooks || prior.CreatedHooks
		out.BackupCreated = out.BackupCreated || prior.BackupCreated
		out.CreatedEvents = append(out.CreatedEvents, prior.CreatedEvents...)
	}
	sort.Strings(out.CreatedEvents)
	var uniq []string
	for i, e := range out.CreatedEvents {
		if i == 0 || e != out.CreatedEvents[i-1] {
			uniq = append(uniq, e)
		}
	}
	out.CreatedEvents = uniq
	return out
}

func hostInList(hosts []string, host string) bool {
	for _, h := range hosts {
		if h == host {
			return true
		}
	}
	return false
}

func sortedHostHookPaths(m map[string]HostHooksRecord) []string {
	paths := make([]string, 0, len(m))
	for p := range m {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

const (
	codexHookNote  = "Codex hooks: open /hooks in Codex and approve the Hive hook before it runs. Hive never writes the trust entry."
	cursorHookNote = "Cursor hooks: registered in the installed code's format only. Cursor CLI's use of hooks.json is not documented, so the events may not arrive."
)

// HostHookPaths lists the shared JSON hook files p edits, for a preview: they
// are steps, not file Changes, so a summary of p.Changes does not show them.
func HostHookPaths(p Plan) []string {
	var out []string
	for _, st := range p.HostHooks {
		out = append(out, st.Path)
	}
	return out
}

// HostHookNotes returns what the user must still do or know after p is applied:
// Codex skips a new or changed hook until the user approves it in /hooks, and
// Cursor CLI's use of hooks.json is verified in the installed code only. It
// lists a note only for a file p would change, so an update that finds the
// entries current says nothing. Call it before Apply: after it the entries are
// already in place. The notes are not part of Apply's result, whose callers
// read it as a transaction ID.
func HostHookNotes(p Plan) ([]string, error) {
	var notes []string
	for _, st := range p.HostHooks {
		spec, ok := hostHookSpec(st.Host, p.Config)
		if !ok {
			return nil, fmt.Errorf("invalid host hooks step")
		}
		cur, err := readHookFile(st.Path)
		if err != nil {
			return nil, err
		}
		state, _, err := readState(p.StateDir)
		if err != nil {
			return nil, err
		}
		res, err := editJSONHooks(cur.Data, cur.Exists, spec, p.Executable, priorHostHooks(state, st.Path))
		if err != nil {
			return nil, err
		}
		if !res.Changed {
			continue
		}
		switch st.Host {
		case "codex":
			notes = append(notes, codexHookNote)
		case "cursor":
			notes = append(notes, cursorHookNote)
		}
	}
	return notes, nil
}

// prepareHostHooks builds the journal entries and the next state records for
// the plan's steps, or for the removal of what the state records.
func prepareHostHooks(p Plan, state State, next *State, j *journal) error {
	next.HostHooks = copyHostHooks(state.HostHooks)
	if next.HostHooks == nil {
		next.HostHooks = map[string]HostHooksRecord{}
	}
	switch {
	case p.Action == "install":
		for _, st := range p.HostHooks {
			spec, ok := hostHookSpec(st.Host, p.Config)
			if !ok {
				return fmt.Errorf("invalid host hooks step")
			}
			cur, err := readHookFile(st.Path)
			if err != nil {
				return fmt.Errorf("%s hooks: %w", st.Host, err)
			}
			if finger(cur) != st.Expected {
				return fmt.Errorf("stale target: %s", st.Path)
			}
			prior := priorHostHooks(state, st.Path)
			res, err := editJSONHooks(cur.Data, cur.Exists, spec, p.Executable, prior)
			if err != nil {
				return fmt.Errorf("%s hooks: %s: %w", st.Host, st.Path, err)
			}
			merged := mergeHostHooks(prior, res.Delta, st.Host)
			if res.Changed {
				mode := cur.Mode
				if !cur.Exists {
					mode = 0600
				}
				backup, err := read(codexBackupPath(st.Path))
				if err != nil {
					return err
				}
				le := &hostHooksEntry{
					Host: st.Host, Path: st.Path, Before: finger(cur),
					After:        finger(snapshot{Exists: true, Data: res.Data, Mode: mode}),
					Record:       res.Delta,
					BackupWanted: cur.Exists && !backup.Exists,
				}
				le.Record.Host = st.Host
				if prior != nil {
					le.Known = prior.Commands
				}
				merged.BackupCreated = merged.BackupCreated || le.BackupWanted
				j.HostHooks = append(j.HostHooks, le)
			}
			next.HostHooks[st.Path] = merged
		}
	case p.Action == "remove" && p.Config.Scope == "user":
		for _, path := range sortedHostHookPaths(state.HostHooks) {
			rec := state.HostHooks[path]
			if !hostInList(p.Hosts, rec.Host) {
				continue
			}
			j.HostHooks = append(j.HostHooks, &hostHooksEntry{Host: rec.Host, Path: path, Remove: true, Record: copyHostHooksRecord(rec)})
			delete(next.HostHooks, path)
		}
	}
	if len(next.HostHooks) == 0 {
		next.HostHooks = nil
	}
	return nil
}

// hostHooksChanged reports whether the host hooks part of a transaction writes
// a file or changes what the state records.
func hostHooksChanged(state, next State, j journal) bool {
	if len(j.HostHooks) > 0 {
		return true
	}
	return !reflect.DeepEqual(state.HostHooks, next.HostHooks)
}

// hostHooksUnchanged reports whether the host hooks part of p changes nothing.
func hostHooksUnchanged(p Plan, s State) (bool, error) {
	if p.Action != "install" {
		if p.Action != "remove" || p.Config.Scope != "user" {
			return true, nil
		}
		for _, r := range s.HostHooks {
			if hostInList(p.Hosts, r.Host) {
				return false, nil
			}
		}
		return true, nil
	}
	for _, st := range p.HostHooks {
		spec, ok := hostHookSpec(st.Host, p.Config)
		if !ok {
			return false, fmt.Errorf("invalid host hooks step")
		}
		cur, err := readHookFile(st.Path)
		if err != nil {
			return false, fmt.Errorf("%s hooks: %w", st.Host, err)
		}
		if finger(cur) != st.Expected {
			return false, fmt.Errorf("stale target: %s", st.Path)
		}
		prior := priorHostHooks(s, st.Path)
		res, err := editJSONHooks(cur.Data, cur.Exists, spec, p.Executable, prior)
		if err != nil {
			return false, err
		}
		if res.Changed || prior == nil || !reflect.DeepEqual(*prior, mergeHostHooks(prior, res.Delta, st.Host)) {
			return false, nil
		}
	}
	return true, nil
}

// applyHostHooks performs each entry's journaled write (install) or entry-level
// inverse (remove) after the files are written and before the state is.
func applyHostHooks(j *journal, jp string, fail func(string) error) error {
	for _, le := range j.HostHooks {
		var err error
		if le.Remove {
			err = removeHostHooksEntry(j, jp, le, fail)
		} else {
			err = installHostHooksEntry(j, jp, le, fail)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func installHostHooksEntry(j *journal, jp string, le *hostHooksEntry, fail func(string) error) error {
	spec, ok := hostHookSpec(le.Host, j.Plan.Config)
	if !ok {
		return fmt.Errorf("invalid host hooks step")
	}
	cur, err := readHookFile(le.Path)
	if err != nil {
		return err
	}
	if finger(cur) != le.Before {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	res, err := editJSONHooks(cur.Data, cur.Exists, spec, j.Plan.Executable, &HostHooksRecord{Commands: le.Known})
	if err != nil || !res.Changed {
		return fmt.Errorf("concurrent change: %s", le.Path)
	}
	mode := cur.Mode
	if !cur.Exists {
		mode = 0600
	}
	after := snapshot{Exists: true, Data: res.Data, Mode: mode}
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
			if err = fail("host-hooks-backup"); err != nil {
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
		if err = fail("host-hooks-write"); err != nil {
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
		return fail("host-hooks-after")
	}
	return nil
}

func removeHostHooksEntry(j *journal, jp string, le *hostHooksEntry, fail func(string) error) error {
	le.Pending = true
	if err := saveJournal(jp, *j); err != nil {
		return err
	}
	if fail != nil {
		if err := fail("host-hooks-pending"); err != nil {
			return err
		}
	}
	spec, ok := hostHookSpec(le.Host, j.Plan.Config)
	if !ok {
		return fmt.Errorf("invalid host hooks record")
	}
	spec.Path = le.Path
	cur, err := readHookFile(le.Path)
	if err != nil {
		return err
	}
	clean := true
	if cur.Exists {
		res := undoJSONHooks(cur.Data, spec, le.Record)
		switch {
		case res.Skipped != "":
			clean = false
			le.Notes = append(le.Notes, fmt.Sprintf("%s hooks: left %s as it is, because %s. Remove the Hive entries by hand.", le.Host, le.Path, res.Skipped))
		case res.DeleteFile:
			err = write(le.Path, snapshot{})
		case string(res.Data) != string(cur.Data):
			err = write(le.Path, snapshot{Exists: true, Data: res.Data, Mode: cur.Mode})
		}
		if err != nil {
			return err
		}
	}
	if fail != nil {
		if err = fail("host-hooks-write"); err != nil {
			return err
		}
	}
	bk := codexBackupPath(le.Path)
	if le.Record.BackupCreated {
		if clean {
			if err = write(bk, snapshot{}); err != nil {
				return err
			}
		} else if b, err := read(bk); err == nil && b.Exists {
			le.Notes = append(le.Notes, fmt.Sprintf("%s hooks: kept %s", le.Host, filepath.Base(bk)))
		}
	}
	le.Pending, le.Done = false, true
	return saveJournal(jp, *j)
}

// undoHostHooksInstall rolls back an interrupted install's writes: it removes
// the entries the journal says an install wrote, wherever they are in the file
// as it is now, so an edit another tool made elsewhere survives, and the backup
// the entry created.
func undoHostHooksInstall(j journal) error {
	for _, le := range j.HostHooks {
		if le.Remove || !(le.Pending || le.Done) {
			continue
		}
		spec, ok := hostHookSpec(le.Host, j.Plan.Config)
		if !ok {
			return fmt.Errorf("invalid host hooks record")
		}
		cur, err := read(le.Path)
		if err != nil {
			return err
		}
		clean := true
		if cur.Exists && finger(cur) != le.Before {
			rec := le.Record
			rec.Commands = append(append([]string(nil), le.Known...), rec.Commands...)
			res := undoJSONHooks(cur.Data, spec, rec)
			clean = res.Skipped == ""
			switch {
			case res.DeleteFile:
				err = write(le.Path, snapshot{})
			case string(res.Data) != string(cur.Data):
				err = write(le.Path, snapshot{Exists: true, Data: res.Data, Mode: cur.Mode})
			}
			if err != nil {
				return err
			}
		}
		if !le.BackupWanted {
			continue
		}
		now, err := read(le.Path)
		if err != nil {
			return err
		}
		if !clean && finger(now) != le.Before {
			continue
		}
		bk := codexBackupPath(le.Path)
		b, err := read(bk)
		if err != nil {
			return err
		}
		if b.Exists && hash(b.Data) == le.Before.Hash {
			if err = write(bk, snapshot{}); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}
