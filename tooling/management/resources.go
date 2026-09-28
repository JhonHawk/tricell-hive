package management

import (
	"fmt"
	"os"
	"path/filepath"
	"tricell-hive/integrations/target"
)

// Unlike read, this narrowly inspects the declared alias leaf without following
// it. All ancestors remain subject to Safe. The caller still checks ownership.
func readResource(t target.Target, migration bool) (snapshot, error) {
	if t.Kind != "symlink" {
		return read(t.Path)
	}
	if err := target.Safe(filepath.Dir(t.Path)); err != nil {
		return snapshot{}, err
	}
	info, err := os.Lstat(t.Path)
	if os.IsNotExist(err) {
		return snapshot{}, nil
	}
	if err != nil {
		return snapshot{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(t.Path)
		return snapshot{Exists: true, Kind: "symlink", LinkTarget: link}, err
	}
	if info.IsDir() && migration {
		entries, err := os.ReadDir(t.Path)
		if err != nil {
			return snapshot{}, err
		}
		if len(entries) == 0 {
			return snapshot{Exists: true, Kind: "directory", Mode: uint32(info.Mode().Perm())}, nil
		}
		if len(entries) != 1 || entries[0].Name() != "SKILL.md" {
			return snapshot{}, fmt.Errorf("legacy skill directory contains unowned entries: %s", t.Path)
		}
		s, err := read(filepath.Join(t.Path, "SKILL.md"))
		if err != nil {
			return snapshot{}, err
		}
		return snapshot{Exists: true, Kind: "skill-directory", Data: s.Data, Mode: uint32(info.Mode().Perm()), FileMode: s.Mode}, nil
	}
	return snapshot{}, fmt.Errorf("unowned or changed alias resource: %s", t.Path)
}
func transformResource(s snapshot, ch Change) (snapshot, error) {
	if ch.Replaces != nil {
		if s.Kind != "skill-directory" || ch.After == nil || ch.Target.Kind != "symlink" {
			return snapshot{}, fmt.Errorf("invalid legacy skill migration")
		}
		if err := owned(snapshot{Exists: true, Data: s.Data, Mode: s.FileMode}, *ch.Replaces, hiveMarkers); err != nil {
			return snapshot{}, err
		}
		return snapshot{Exists: true, Kind: "symlink", LinkTarget: ch.Target.LinkTarget}, nil
	}
	return transform(s, ch.Before, ch.After, hiveMarkers)
}

// A directory snapshot is used only for the known V1 Claude directory -> alias
// migration. It carries its single owned file so recovery never follows a link.
func writeResource(t target.Target, expected, s snapshot, migration bool, fail func(string) error) error {
	if t.Kind != "symlink" {
		cur, err := read(t.Path)
		if err != nil {
			return err
		}
		if !same(cur, expected) {
			return fmt.Errorf("concurrent resource change: %s", t.Path)
		}
		return write(t.Path, s)
	}
	cur, err := readResource(t, migration)
	if err != nil {
		return err
	}
	if !same(cur, expected) {
		return fmt.Errorf("concurrent alias change: %s", t.Path)
	}
	if cur.Kind == "skill-directory" {
		if err = os.Remove(filepath.Join(t.Path, "SKILL.md")); err != nil {
			return err
		}
		if fail != nil {
			if err = fail("migration:skill"); err != nil {
				return err
			}
		}
		cur.Kind = "directory"
	}
	if cur.Kind == "directory" {
		if err = os.Remove(t.Path); err != nil {
			return err
		}
		if fail != nil {
			if err = fail("migration:directory"); err != nil {
				return err
			}
		}
		cur = snapshot{}
	}
	if s.Kind == "skill-directory" || s.Kind == "directory" {
		if cur.Exists {
			if err = os.Remove(t.Path); err != nil {
				return err
			}
		}
		if err = os.Mkdir(t.Path, os.FileMode(s.Mode)); err != nil {
			return err
		}
		if s.Kind == "skill-directory" {
			return write(filepath.Join(t.Path, "SKILL.md"), snapshot{Exists: true, Data: s.Data, Mode: s.FileMode})
		}
		return nil
	}
	if !s.Exists {
		if !cur.Exists {
			return nil
		}
		return os.Remove(t.Path)
	}
	if s.Kind != "symlink" || s.LinkTarget != t.LinkTarget {
		return fmt.Errorf("invalid managed link payload")
	}
	// Rename a sibling temporary link; never open the link's referent for writes.
	f, err := os.CreateTemp(filepath.Dir(t.Path), ".hive-link-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	if err = f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err = os.Remove(tmp); err != nil {
		return err
	}
	defer os.Remove(tmp)
	if err = os.Symlink(s.LinkTarget, tmp); err != nil {
		return err
	}
	if err = target.Safe(filepath.Dir(t.Path)); err != nil {
		return err
	}
	latest, err := readResource(t, migration)
	if err != nil {
		return err
	}
	if !same(latest, cur) {
		return fmt.Errorf("concurrent alias change: %s", t.Path)
	}
	if err = os.Rename(tmp, t.Path); err != nil {
		return err
	}
	d, err := os.Open(filepath.Dir(t.Path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
