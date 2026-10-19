package viewer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
)

func replaceWithHardlink(t *testing.T, path string) {
	t.Helper()
	outside := filepath.Join(t.TempDir(), "linked-source")
	if err := os.WriteFile(outside, readFile(t, path), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(outside, path); err != nil {
		t.Fatal(err)
	}
}

func TestManifestHardlink(t *testing.T) {
	for _, rel := range []string{manifestFile, "images/nav-1.png"} {
		t.Run(rel, func(t *testing.T) {
			folder := copyFixture(t, "valid/final")
			replaceWithHardlink(t, filepath.Join(folder, rel))
			if _, err := ValidateManifest(folder); err == nil || !strings.Contains(err.Error(), "hard link") {
				t.Fatalf("hardlinked %s: got %v, want explicit rejection", rel, err)
			}
		})
	}
	t.Run("regular", func(t *testing.T) {
		if _, err := ValidateManifest(copyFixture(t, "valid/final")); err != nil {
			t.Fatal(err)
		}
	})
}

func TestScanHardlink(t *testing.T) {
	for _, phase := range []string{"manifest", "resource", "after-validation", "after-hash"} {
		for _, previous := range []bool{false, true} {
			name := phase + "/new"
			if previous {
				name = phase + "/replacement"
			}
			t.Run(name, func(t *testing.T) {
				e := newEnv(t)
				repo := newRepo(t)
				folder := e.place(repo, "run1", nil)
				e.register(repo)
				e.start()
				var old []Session
				oldFiles := map[string]string{}
				if previous {
					e.sync()
					old = append(old, e.store.current().Sessions...)
					stored := e.store.sessionDir(old[0].Key)
					err := filepath.WalkDir(stored, func(path string, d os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !d.IsDir() {
							oldFiles[path] = string(readFile(t, path))
						}
						return nil
					})
					if err != nil {
						t.Fatal(err)
					}
					m := readManifestMap(t, folder)
					m["title"] = "Changed candidate"
					writeManifestMap(t, folder, m)
					e.touch(folder)
				}
				replace := func() { replaceWithHardlink(t, filepath.Join(folder, "images/nav-1.png")) }
				switch phase {
				case "manifest":
					replaceWithHardlink(t, filepath.Join(folder, manifestFile))
				case "resource":
					replace()
				case "after-validation":
					e.store.afterValidate = func(string) { replace() }
				case "after-hash":
					e.store.onBatch = func(int) { replace() }
				}
				res := e.sync()
				if res.Errors != 1 || res.Added != 0 || res.Replaced != 0 {
					t.Fatalf("hardlink was accepted: %+v", res)
				}
				if got := e.store.View().Errors; len(got) != 1 || !strings.Contains(got[0].Reason, "hard link") {
					t.Fatalf("expected hardlink diagnostic, got %+v", got)
				}
				if got := e.store.current().Sessions; !reflect.DeepEqual(got, old) && (previous || len(got) != 0) {
					t.Fatalf("sessions changed: got %+v, want %+v", got, old)
				}
				for path, data := range oldFiles {
					if string(readFile(t, path)) != data {
						t.Fatalf("previous copy changed: %s", path)
					}
				}
				entries, err := os.ReadDir(e.store.sessionsDir())
				if err != nil {
					t.Fatal(err)
				}
				want := 0
				if previous {
					want = 1
				}
				if len(entries) != want {
					t.Fatalf("partial copies remain: %v", entries)
				}
			})
		}
	}
}

// hardlinkFileInfo supplies only the synthetic metadata used by the link policy.
type hardlinkFileInfo struct {
	os.FileInfo
	metadata any
}

func (i hardlinkFileInfo) Sys() any { return i.metadata }

func TestManifestHardlinkMetadata(t *testing.T) {
	for _, tc := range []struct {
		name      string
		metadata  any
		wantError string
	}{
		{"missing", nil, "cannot determine"},
		{"unsupported", struct{}{}, "cannot determine"},
		{"nil-stat", (*syscall.Stat_t)(nil), "cannot determine"},
		{"multiple-links", &syscall.Stat_t{Nlink: 2}, "more than one hard link"},
		{"single-link-other-owner", &syscall.Stat_t{Nlink: 1, Uid: ^uint32(0)}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSingleLink(hardlinkFileInfo{metadata: tc.metadata})
			if tc.wantError == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("got %v, want %q", err, tc.wantError)
			}
		})
	}
}

func TestManifestHardlinkRejectedBeforeParsing(t *testing.T) {
	folder := copyFixture(t, "valid/final")
	path := filepath.Join(folder, manifestFile)
	if err := os.WriteFile(path, []byte("not JSON"), 0o600); err != nil {
		t.Fatal(err)
	}
	replaceWithHardlink(t, path)
	if _, err := ValidateManifest(folder); err == nil || !strings.Contains(err.Error(), "more than one hard link") {
		t.Fatalf("hardlink check did not precede parsing: %v", err)
	}
}
