package codex

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckWaitFloorTarget(t *testing.T) {
	write := func(t *testing.T, path, body string, mode os.FileMode) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), mode); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("missing file in a writable home", func(t *testing.T) {
		exists, err := CheckWaitFloorTarget(t.TempDir())
		if err != nil || exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
	})
	t.Run("missing home", func(t *testing.T) {
		exists, err := CheckWaitFloorTarget(filepath.Join(t.TempDir(), "absent"))
		if err != nil || exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
	})
	t.Run("regular file", func(t *testing.T) {
		home := t.TempDir()
		write(t, filepath.Join(home, "config.toml"), "x = 1\n", 0600)
		exists, err := CheckWaitFloorTarget(home)
		if err != nil || !exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
	})
	refusals := map[string]func(t *testing.T) string{
		"symlinked file": func(t *testing.T) string {
			home := t.TempDir()
			target := filepath.Join(t.TempDir(), "real.toml")
			write(t, target, "x = 1\n", 0600)
			if err := os.Symlink(target, filepath.Join(home, "config.toml")); err != nil {
				t.Fatal(err)
			}
			return home
		},
		"dangling symlinked file": func(t *testing.T) string {
			home := t.TempDir()
			if err := os.Symlink(filepath.Join(home, "nowhere"), filepath.Join(home, "config.toml")); err != nil {
				t.Fatal(err)
			}
			return home
		},
		"symlinked home": func(t *testing.T) string {
			real := t.TempDir()
			link := filepath.Join(t.TempDir(), "codex")
			if err := os.Symlink(real, link); err != nil {
				t.Fatal(err)
			}
			return link
		},
		"directory instead of file": func(t *testing.T) string {
			home := t.TempDir()
			if err := os.Mkdir(filepath.Join(home, "config.toml"), 0700); err != nil {
				t.Fatal(err)
			}
			return home
		},
		"read-only file": func(t *testing.T) string {
			home := t.TempDir()
			write(t, filepath.Join(home, "config.toml"), "x = 1\n", 0400)
			return home
		},
		"file over 1 MiB": func(t *testing.T) string {
			home := t.TempDir()
			write(t, filepath.Join(home, "config.toml"), strings.Repeat("#\n", 600*1024), 0600)
			return home
		},
		"read-only home without file": func(t *testing.T) string {
			home := filepath.Join(t.TempDir(), "codex")
			if err := os.Mkdir(home, 0500); err != nil {
				t.Fatal(err)
			}
			return home
		},
	}
	for name, setup := range refusals {
		t.Run(name, func(t *testing.T) {
			home := setup(t)
			exists, err := CheckWaitFloorTarget(home)
			var ref *WaitFloorRefusal
			if !errors.As(err, &ref) || ref.State != WaitFloorUnsupported || ref.Cause == "" {
				t.Fatalf("want an unsupported refusal, got exists=%v err=%v", exists, err)
			}
			if strings.Contains(err.Error(), home) {
				t.Fatalf("the refusal must not repeat the path: %v", err)
			}
		})
	}
}
