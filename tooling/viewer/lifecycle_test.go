package viewer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManifestSessionStatus(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value any
		want  string
		valid bool
	}{
		{"missing", nil, "unknown", true}, {"active", "in_progress", "in_progress", true}, {"closed", "closed", "closed", true},
		{"unknown", "unknown", "", false}, {"empty", "", "", false}, {"null", nil, "", false}, {"number", 1, "", false}, {"boolean", true, "", false},
		{"object", map[string]any{}, "", false}, {"array", []string{"closed"}, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			folder := copyFixture(t, "valid/final")
			m := readManifestMap(t, folder)
			if tc.name != "missing" {
				m["session"].(map[string]any)["status"] = tc.value
			}
			writeManifestMap(t, folder, m)
			got, err := ValidateManifest(folder)
			if !tc.valid {
				if err == nil {
					t.Fatal("invalid status accepted")
				}
				if _, err := readSummary(folder); err == nil {
					t.Fatal("stored summary accepted invalid status")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.SessionStatus != tc.want {
				t.Fatalf("status = %q, want %q", got.SessionStatus, tc.want)
			}
		})
	}
}

func TestSourceFingerprintIncludesResourcesAndCanonicalIdentity(t *testing.T) {
	folder := copyFixture(t, "valid/final")
	first, err := FingerprintSource(folder)
	if err != nil {
		t.Fatal(err)
	}
	canonical, _ := filepath.EvalSymlinks(folder)
	if first.Source != canonical || first.Key != sessionKey(canonical, first.SessionID) || first.SessionStatus != "unknown" {
		t.Fatalf("identity = %+v", first)
	}
	m, err := ValidateManifest(folder)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) == 0 {
		t.Fatal("fixture must include a resource")
	}
	if err := os.WriteFile(filepath.Join(folder, m.Files[0]), []byte("changed resource"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := FingerprintSource(folder)
	if err != nil {
		t.Fatal(err)
	}
	if first.ContentHash == second.ContentHash || first.Key != second.Key {
		t.Fatal("resource change must change only the revision")
	}
}
