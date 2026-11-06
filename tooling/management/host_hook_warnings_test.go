package management

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostHookWarningsNameAnExecutableInsideAGitCheckout(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// A worktree has a .git file, a clone a .git folder: both count.
	for name, mk := range map[string]func(string){
		"clone":    func(dir string) { _ = os.MkdirAll(filepath.Join(dir, ".git"), 0o700) },
		"worktree": func(dir string) { _ = os.WriteFile(filepath.Join(dir, ".git"), []byte("gitdir: /x\n"), 0o600) },
	} {
		t.Run(name, func(t *testing.T) {
			checkout := filepath.Join(root, name)
			bin := filepath.Join(checkout, "bin", "hive")
			_ = os.MkdirAll(filepath.Dir(bin), 0o700)
			mk(checkout)
			p := Plan{Executable: bin, HostHooks: []HostHooksStep{{Host: "claude"}}}
			w := HostHookWarnings(p)
			if len(w) != 1 || !strings.Contains(w[0], "git checkout \""+name+"\"") {
				t.Fatalf("warnings = %q", w)
			}
		})
	}
	outside := filepath.Join(root, "installed", "bin", "hive")
	_ = os.MkdirAll(filepath.Dir(outside), 0o700)
	if w := HostHookWarnings(Plan{Executable: outside, HostHooks: []HostHooksStep{{Host: "claude"}}}); len(w) != 0 {
		t.Fatalf("a binary outside any checkout warned: %q", w)
	}
	if w := HostHookWarnings(Plan{Executable: filepath.Join(root, "clone", "bin", "hive")}); len(w) != 0 {
		t.Fatalf("a plan with no hook steps warned: %q", w)
	}
}
