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
	return packageFixtureWithout(t, "")
}

// packageFixtureWithout builds a valid package fixture, leaving out the one
// payload file named by skip (empty keeps every file).
func packageFixtureWithout(t *testing.T, skip string) string {
	t.Helper()
	root := t.TempDir()
	for _, p := range []string{"bin/hive", "bin/hive.sha256", "install.sh", "platform", "content/guidance/global.md", "integrations/agent-profiles.json", "LICENSE", "THIRD_PARTY_NOTICES.md"} {
		if p == skip {
			continue
		}
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

func TestReadManifestDataReturnsExactBytesAndUnwrappedNotExist(t *testing.T) {
	root := packageFixture(t)
	want, err := os.ReadFile(filepath.Join(root, ManifestName))
	if err != nil {
		t.Fatal(err)
	}
	m, got, err := ReadManifestData(root)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("ReadManifestData() bytes = %q, want %q", got, want)
	}
	if m.Version != 1 {
		t.Fatalf("ReadManifestData() manifest = %+v", m)
	}

	absent := t.TempDir()
	if _, _, err := ReadManifestData(absent); !os.IsNotExist(err) {
		t.Fatalf("ReadManifestData() on absent manifest = %v, want os.IsNotExist", err)
	}
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
	for _, name := range []string{"LICENSE", "THIRD_PARTY_NOTICES.md"} {
		t.Run("missing "+name, func(t *testing.T) {
			err := VerifyIfPackaged(packageFixtureWithout(t, name))
			if err == nil || !strings.Contains(err.Error(), "incomplete package: "+name) {
				t.Fatalf("VerifyIfPackaged() = %v, want incomplete package: %s", err, name)
			}
		})
	}
	t.Run("checkout", func(t *testing.T) {
		if err := VerifyIfPackaged(t.TempDir()); err != nil {
			t.Fatal(err)
		}
	})
}
