package management

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"tricell-hive/integrations/target"
	"tricell-hive/tooling/legacy"
)

type MigrationReceipt struct {
	Config      target.Config `json:"config"`
	Hosts       []string      `json:"hosts"`
	Detector    string        `json:"detector"`
	Result      string        `json:"result"`
	Release     string        `json:"release"`
	Transaction string        `json:"transaction,omitempty"`
}

func migrationNeeded(p Plan, s State) bool {
	if p.Action != "install" || p.Config.Scope != "user" {
		return false
	}
	if len(p.Legacy) > 0 {
		return true
	}
	for _, host := range p.Hosts {
		covered := false
		for _, r := range s.Migrations {
			if r.Config == p.Config && r.Detector == legacy.Version {
				for _, h := range r.Hosts {
					if h == host {
						covered = true
					}
				}
			}
		}
		if !covered {
			return true
		}
	}
	return false
}

func scanMigration(p *Plan, s State) error {
	if p.Action != "install" || p.Config.Scope != "user" {
		return nil
	}
	result, err := scanLegacy(p.Config, p.Hosts, s)
	if err != nil {
		return err
	}
	p.Legacy = result.Edits
	if migrationNeeded(*p, s) {
		status := "clean"
		if result.Detected {
			status = "migrated"
		}
		p.Migration = &MigrationReceipt{Config: p.Config, Hosts: p.Hosts, Detector: legacy.Version, Result: status, Release: p.Release.ID}
	}
	return nil
}
func validateMigration(p Plan, s State) error {
	expected := p
	expected.Legacy = nil
	expected.Migration = nil
	if err := scanMigration(&expected, s); err != nil {
		return err
	}
	if !reflect.DeepEqual(expected.Legacy, p.Legacy) || !reflect.DeepEqual(expected.Migration, p.Migration) {
		return fmt.Errorf("legacy inventory changed; regenerate plan")
	}
	return nil
}
func legacyAfter(edit legacy.Edit) snapshot {
	if edit.Delete {
		return snapshot{}
	}
	return snapshot{Exists: true, Data: edit.After, Mode: edit.Mode}
}
func overlayRead(p Plan, t target.Target, replaces bool) (snapshot, error) {
	for _, edit := range p.Legacy {
		if edit.Path == t.Path {
			return legacyAfter(edit), nil
		}
	}
	return readResource(t, replaces)
}
func deletedByLegacy(p Plan, path string) bool {
	for _, e := range p.Legacy {
		if e.Path == path && e.Delete {
			return true
		}
	}
	return false
}

// PlanUnchanged performs a read-only preflight; it does not acquire a write lock.
func PlanUnchanged(p Plan) (bool, error) {
	s, sh, err := readState(p.StateDir)
	if err != nil {
		return false, err
	}
	if sh != p.StateHash {
		return false, fmt.Errorf("stale state")
	}
	if err = validatePlan(p, s); err != nil {
		return false, err
	}
	if err = validateMigration(p, s); err != nil {
		return false, err
	}
	if p.Migration != nil || len(p.Legacy) > 0 || s.Version != 5 {
		return false, nil
	}
	for _, ch := range p.Changes {
		cur, err := readResource(ch.Target, ch.Replaces != nil)
		if err != nil {
			return false, err
		}
		if finger(cur) != ch.Expected {
			return false, fmt.Errorf("stale target: %s", ch.Target.Path)
		}
		after, err := transformResource(cur, ch)
		if err != nil {
			return false, err
		}
		if !same(cur, after) || !reflect.DeepEqual(ch.Before, ch.After) {
			return false, nil
		}
	}
	return true, nil
}

// Trees are restricted to catalog-proven directory replacements. Every leaf is
// represented and checked before removal; no generic recursive delete is used.
type treeItem struct {
	Path      string
	Data      []byte
	Mode      uint32
	Directory bool
}

func readLegacyTree(path string) (snapshot, error) {
	if err := target.Safe(path); err != nil {
		return snapshot{}, err
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return snapshot{}, nil
	}
	if err != nil {
		return snapshot{}, err
	}
	if !info.IsDir() {
		return snapshot{}, fmt.Errorf("legacy directory changed: %s", path)
	}
	s := snapshot{Exists: true, Kind: "legacy-directory", Mode: uint32(info.Mode().Perm())}
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if p == path {
			return nil
		}
		rel, _ := filepath.Rel(path, p)
		if d.IsDir() {
			i, e := d.Info()
			if e != nil {
				return e
			}
			s.Tree = append(s.Tree, treeItem{Path: rel, Mode: uint32(i.Mode().Perm()), Directory: true})
			return nil
		}
		v, e := read(p)
		if e != nil {
			return e
		}
		s.Tree = append(s.Tree, treeItem{Path: rel, Data: v.Data, Mode: v.Mode})
		return nil
	})
	return s, err
}
func treeProven(p Plan, path string, s snapshot) error {
	edits := map[string]legacy.Edit{}
	for _, ed := range p.Legacy {
		edits[ed.Path] = ed
	}
	root, ok := edits[path]
	if !ok || !root.Delete || root.Kind != "directory" || root.Mode != s.Mode {
		return fmt.Errorf("unproven directory: %s", path)
	}
	for _, item := range s.Tree {
		ed, ok := edits[filepath.Join(path, item.Path)]
		if !ok || !ed.Delete || ed.Mode != item.Mode || item.Directory != (ed.Kind == "directory") || (!item.Directory && !bytes.Equal(item.Data, ed.Before)) {
			return fmt.Errorf("legacy directory entry changed: %s", item.Path)
		}
	}
	// A removed leaf is also a concurrent inventory change, not authorization to
	// silently reduce the frozen catalog operation.
	for _, ed := range p.Legacy {
		if strings.HasPrefix(ed.Path, path+string(os.PathSeparator)) {
			found := false
			for _, item := range s.Tree {
				if filepath.Join(path, item.Path) == ed.Path {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("legacy directory entry missing: %s", ed.Path)
			}
		}
	}
	return nil
}

func readEntry(en entry) (snapshot, error) {
	if en.Before.Kind == "legacy-directory" || en.After.Kind == "legacy-directory" {
		info, err := os.Lstat(en.Change.Target.Path)
		if os.IsNotExist(err) {
			return snapshot{}, nil
		}
		if err != nil {
			return snapshot{}, err
		}
		if info.IsDir() {
			return readLegacyTree(en.Change.Target.Path)
		}
	}
	return readResource(en.Change.Target, en.Change.Replaces != nil)
}
func writeEntry(en entry, cur, after snapshot, fail func(string) error) error {
	if cur.Kind != "legacy-directory" && after.Kind != "legacy-directory" {
		return writeResource(en.Change.Target, cur, after, en.Change.Replaces != nil, fail)
	}
	// Recheck before each destructive tree operation. Recovery accepts only exact
	// before/after trees, or a catalog-owned partial tree after a failed write.
	path := en.Change.Target.Path
	latest, err := readEntry(en)
	if err != nil {
		return err
	}
	if !same(latest, cur) {
		return fmt.Errorf("concurrent tree change: %s", path)
	}
	if cur.Kind == "legacy-directory" {
		for i := len(cur.Tree) - 1; i >= 0; i-- {
			item := cur.Tree[i]
			p := filepath.Join(path, item.Path)
			if err = target.Safe(p); err != nil {
				return err
			}
			if !item.Directory {
				v, e := read(p)
				if e != nil {
					return e
				}
				if !bytes.Equal(v.Data, item.Data) || v.Mode != item.Mode {
					return fmt.Errorf("tree file changed: %s", p)
				}
			}
			if err = os.Remove(p); err != nil {
				return err
			}
			if fail != nil {
				if err = fail("legacy:remove:" + filepath.ToSlash(item.Path)); err != nil {
					return err
				}
			}
		}
		if err = os.Remove(path); err != nil {
			return err
		}
		if fail != nil {
			if err = fail("legacy:directory"); err != nil {
				return err
			}
		}
	} else if cur.Exists {
		if err = os.Remove(path); err != nil {
			return err
		}
	}
	if !after.Exists {
		return nil
	}
	if after.Kind == "legacy-directory" {
		if err = os.Mkdir(path, os.FileMode(after.Mode)); err != nil {
			return err
		}
		for _, item := range after.Tree {
			p := filepath.Join(path, item.Path)
			if err = target.Safe(p); err != nil {
				return err
			}
			if item.Directory {
				err = os.Mkdir(p, os.FileMode(item.Mode))
			} else {
				err = write(p, snapshot{Exists: true, Data: item.Data, Mode: item.Mode})
			}
			if err != nil {
				return err
			}
		}
		return nil
	}
	return writeResource(en.Change.Target, snapshot{}, after, false, fail)
}
func prepareEntries(p Plan) ([]entry, error) {
	var entries []entry
	consumed := map[string]bool{}
	// Core resources can replace a legacy file or whole skill directory. Store one
	// inverse per physical path, never separate overlapping before-images.
	for _, ch := range p.Changes {
		cur, err := overlayRead(p, ch.Target, ch.Replaces != nil)
		if err != nil {
			return nil, err
		}
		if finger(cur) != ch.Expected {
			return nil, fmt.Errorf("stale target: %s", ch.Target.Path)
		}
		after, err := transformResource(cur, ch)
		if err != nil {
			return nil, err
		}
		var before snapshot
		if ch.Target.Kind == "symlink" && deletedByLegacy(p, ch.Target.Path) {
			before, err = readLegacyTree(ch.Target.Path)
			if err == nil {
				err = treeProven(p, ch.Target.Path, before)
			}
			for _, ed := range p.Legacy {
				if ed.Path == ch.Target.Path || strings.HasPrefix(ed.Path, ch.Target.Path+string(os.PathSeparator)) {
					consumed[ed.Path] = true
				}
			}
		} else {
			before, err = readResource(ch.Target, ch.Replaces != nil)
			if err == nil {
				for _, ed := range p.Legacy {
					if ed.Path == ch.Target.Path && (!before.Exists || before.Mode != ed.Mode || !bytes.Equal(before.Data, ed.Before)) {
						err = fmt.Errorf("legacy target changed: %s", ed.Path)
						break
					}
				}
			}
			consumed[ch.Target.Path] = true
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry{ch, before, after})
	}
	// Keep ancestor directories used by the new release; their recognized leaves
	// still retire individually and unrelated siblings were already rejected.
	for _, ed := range p.Legacy {
		if ed.Kind == "directory" {
			for _, ch := range p.Changes {
				if ch.After != nil && strings.HasPrefix(ch.Target.Path, ed.Path+string(os.PathSeparator)) {
					consumed[ed.Path] = true
				}
			}
		}
	}
	var preliminary []entry
	// Ancestor directories replace their descendant edits as one tree operation.
	var dirs []legacy.Edit
	for _, ed := range p.Legacy {
		if ed.Kind == "directory" && !consumed[ed.Path] {
			dirs = append(dirs, ed)
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return len(dirs[i].Path) < len(dirs[j].Path) })
	for _, ed := range dirs {
		if consumed[ed.Path] {
			continue
		}
		s, err := readLegacyTree(ed.Path)
		if err != nil {
			return nil, err
		}
		if err = treeProven(p, ed.Path, s); err != nil {
			return nil, err
		}
		preliminary = append(preliminary, entry{Change: Change{Target: target.Target{Path: ed.Path, Kind: "legacy-directory"}}, Before: s})
		for _, other := range p.Legacy {
			if other.Path == ed.Path || strings.HasPrefix(other.Path, ed.Path+string(os.PathSeparator)) {
				consumed[other.Path] = true
			}
		}
	}
	// Config deregistration is ordered by the scanner before executable retirement.
	var files []entry
	for _, ed := range p.Legacy {
		if consumed[ed.Path] {
			continue
		}
		s, err := read(ed.Path)
		if err != nil {
			return nil, err
		}
		if !s.Exists || !bytes.Equal(s.Data, ed.Before) || s.Mode != ed.Mode {
			return nil, fmt.Errorf("legacy target changed: %s", ed.Path)
		}
		files = append(files, entry{Change: Change{Target: target.Target{Path: ed.Path, Kind: "legacy-file"}}, Before: s, After: legacyAfter(ed)})
	}
	return append(append(files, preliminary...), entries...), nil
}
func partialLegacyTree(cur, before snapshot) bool {
	if before.Kind != "legacy-directory" {
		return false
	}
	if !cur.Exists {
		return true
	}
	if cur.Kind != "legacy-directory" || cur.Mode != before.Mode {
		return false
	}
	known := map[string]treeItem{}
	for _, item := range before.Tree {
		known[item.Path] = item
	}
	for _, item := range cur.Tree {
		old, ok := known[item.Path]
		if !ok || !reflect.DeepEqual(old, item) {
			return false
		}
	}
	return true
}

func scanLegacy(c target.Config, hosts []string, s State) (legacy.Result, error) {
	var trusted []string
	for _, r := range s.Records {
		current, err := readResource(r.Target, false)
		if err != nil {
			continue
		}
		if err = owned(current, r); err != nil {
			continue
		}
		trusted = append(trusted, r.Target.Path)
	}
	return legacy.ScanExcluding(c, hosts, trusted)
}
func DetectLegacyHosts(o Options) ([]string, error) {
	c, dir, err := normalize(o)
	if err != nil {
		return nil, err
	}
	s, _, err := readState(dir)
	if err != nil {
		return nil, err
	}
	r, err := scanLegacy(c, []string{"claude", "codex", "grok", "opencode", "pi"}, s)
	return r.Hosts, err
}

func legacyTouches(p Plan, path string) bool {
	for _, ed := range p.Legacy {
		if ed.Path == path {
			return true
		}
	}
	return false
}
