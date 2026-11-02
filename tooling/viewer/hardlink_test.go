package viewer

import (
	"os"
	"path/filepath"
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
