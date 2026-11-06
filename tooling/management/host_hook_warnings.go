package management

import (
	"fmt"
	"os"
	"path/filepath"
)

// gitCheckoutOf returns the nearest folder above path that holds a .git entry (a
// folder in a clone, a file in a worktree), if any.
func gitCheckoutOf(path string) (string, bool) {
	for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir, true
		}
		if dir == filepath.Dir(dir) {
			return "", false
		}
	}
}

// HostHookWarnings returns what the user should know about where the hooks
// point: every hook runs the registered hive executable, so one built inside a
// git checkout or worktree breaks them all when that folder is removed. It says
// nothing for a plan that registers no hooks.
func HostHookWarnings(p Plan) []string {
	if len(p.HostHooks) == 0 || p.Executable == "" {
		return nil
	}
	root, ok := gitCheckoutOf(p.Executable)
	if !ok {
		return nil
	}
	// Names, not paths: the folder is what the user recognizes, and a long
	// absolute path cannot be wrapped in a terminal.
	return []string{fmt.Sprintf("Hooks: the registered hive executable lives inside the git checkout %q, so every hook stops working if that checkout or worktree is removed. Build the binary outside it and run hive update --source . from this checkout again.", filepath.Base(root))}
}
