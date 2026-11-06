package management

import (
	"os"
	"path/filepath"
	"testing"
	"tricell-hive/integrations/target"
)

func TestReleasePathsRespectSyntheticHomeAndXDG(t *testing.T) {
	home, _ := target.Canonical(t.TempDir())
	other, _ := target.Canonical(t.TempDir())
	t.Setenv("HOME", home)
	for _, xdg := range []string{other, "relative/data", ""} {
		t.Setenv("XDG_DATA_HOME", xdg)
		store, err := ReleasePackageStore("")
		want := filepath.Join(home, ".local", "share", "hive", "packages")
		if filepath.IsAbs(xdg) {
			want = filepath.Join(xdg, "hive", "packages")
		}
		if err != nil || store != want {
			t.Fatalf("XDG %q: %q, %v; want %q", xdg, store, err, want)
		}
		synthetic, err := ReleasePackageStore(home)
		if err != nil || synthetic != filepath.Join(home, ".local", "share", "hive", "packages") {
			t.Fatalf("synthetic: %q %v", synthetic, err)
		}
		command, err := ManagedCommandPath(home)
		if err != nil || command != filepath.Join(home, ".local", "bin", "hive") {
			t.Fatalf("command: %q %v", command, err)
		}
	}
	link := filepath.Join(t.TempDir(), "home")
	if err := os.Symlink(home, link); err != nil {
		t.Fatal(err)
	}
	store, err := ReleasePackageStore(link)
	if err != nil || store != filepath.Join(home, ".local", "share", "hive", "packages") {
		t.Fatalf("canonical home: %q %v", store, err)
	}
}
