package codex

import (
	"fmt"
	"os"
	"path/filepath"
)

// CheckWaitFloorTarget maps the file system facts about Codex's config.toml
// under home to the editor's refusal states, without reading the file. It
// reports whether the file exists. It refuses, with a fixed cause that never
// repeats the path, when home or the file is a symbolic link (the writer does
// not follow links), when the file is not a regular file, is larger than the
// scanner accepts, or is not writable by the owner, and when the file is
// absent from a home the owner cannot write. Ancestors of home are not
// inspected: on macOS /tmp and /var are links, and Codex's own home is the
// path the user chose. Mode bits are a hint and do not replace the write
// itself, which still reports a permission failure.
func CheckWaitFloorTarget(home string) (exists bool, err error) {
	if info, lerr := os.Lstat(home); lerr == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return false, refuse(WaitFloorUnsupported, "the Codex home folder is a symbolic link")
		}
		if !info.IsDir() {
			return false, refuse(WaitFloorUnsupported, "the Codex home is not a folder")
		}
	} else if os.IsNotExist(lerr) {
		return false, nil
	} else {
		return false, fmt.Errorf("inspect Codex home: %w", lerr)
	}
	path := filepath.Join(home, "config.toml")
	info, lerr := os.Lstat(path)
	if os.IsNotExist(lerr) {
		if hi, herr := os.Stat(home); herr == nil && hi.Mode().Perm()&0200 == 0 {
			return false, refuse(WaitFloorUnsupported, "the Codex home folder is not writable")
		}
		return false, nil
	}
	if lerr != nil {
		return false, fmt.Errorf("inspect Codex config: %w", lerr)
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return true, refuse(WaitFloorUnsupported, "config.toml is a symbolic link")
	case !info.Mode().IsRegular():
		return true, refuse(WaitFloorUnsupported, "config.toml is not a regular file")
	case info.Size() > maxConfigLine:
		return true, refuse(WaitFloorUnsupported, "config.toml is larger than 1 MiB")
	case info.Mode().Perm()&0200 == 0:
		return true, refuse(WaitFloorUnsupported, "config.toml is not writable")
	}
	return true, nil
}
