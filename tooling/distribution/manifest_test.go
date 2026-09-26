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
	b, err := json.Marshal(Manifest{Version: 1, Platform: runtime.GOOS + "/" + runtime.GOARCH, SourceID: strings.Repeat("a", 64), Files: files})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ManifestName), b, 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestReadManifestAllowsLegacyProductVersion(t *testing.T) {
	root := packageFixture(t)
	m, err := ReadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if m.ProductVersion != "" {
		t.Fatalf("legacy product version = %q", m.ProductVersion)
	}
}

func TestReadManifestReturnsProductVersion(t *testing.T) {
	root := packageFixture(t)
	data, err := os.ReadFile(filepath.Join(root, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	m.ProductVersion = "1.2.3"
	data, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ManifestName), data, 0600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadManifest(root)
	if err != nil || got.ProductVersion != "1.2.3" {
		t.Fatalf("ReadManifest() = %+v, %v", got, err)
	}
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
