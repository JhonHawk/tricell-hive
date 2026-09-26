package distribution

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func packageFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"bin/hive", "bin/hive.sha256", "install.sh", "platform", "content/guidance/global.md", "integrations/agent-profiles.json"} {
		path := filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := Files(root)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(Manifest{1, runtime.GOOS + "/" + runtime.GOARCH, strings.Repeat("a", 64), files})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ManifestName), b, 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestPackageIntegrity(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := VerifyIfPackaged(packageFixture(t)); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("edited payload", func(t *testing.T) {
		root := packageFixture(t)
		os.WriteFile(filepath.Join(root, "content/guidance/global.md"), []byte("changed"), 0600)
		if VerifyIfPackaged(root) == nil {
			t.Fatal("accepted edited payload")
		}
	})
	t.Run("unlisted skill", func(t *testing.T) {
		root := packageFixture(t)
		os.WriteFile(filepath.Join(root, "content/new.md"), []byte("new"), 0600)
		if VerifyIfPackaged(root) == nil {
			t.Fatal("accepted unlisted payload")
		}
	})
	t.Run("lost manifest", func(t *testing.T) {
		root := packageFixture(t)
		os.Remove(filepath.Join(root, ManifestName))
		if VerifyIfPackaged(root) == nil {
			t.Fatal("lost manifest bypassed verification")
		}
	})
	t.Run("symlink", func(t *testing.T) {
		root := packageFixture(t)
		os.Remove(filepath.Join(root, "install.sh"))
		os.Symlink("platform", filepath.Join(root, "install.sh"))
		if VerifyIfPackaged(root) == nil {
			t.Fatal("accepted symlink")
		}
	})
	t.Run("checkout", func(t *testing.T) {
		if err := VerifyIfPackaged(t.TempDir()); err != nil {
			t.Fatal(err)
		}
	})
}
