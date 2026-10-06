// Package distribution verifies complete offline installer packages.
package distribution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const ManifestName = "release.json"

type Manifest struct {
	Version        int               `json:"version"`
	Platform       string            `json:"platform"`
	SourceID       string            `json:"source_id"`
	ProductVersion string            `json:"product_version,omitempty"`
	Files          map[string]string `json:"files"`
}

func Digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Files inventories every regular file except the manifest itself. Links and
// special files are rejected so validation and later loading see the same tree.
func Files(root string) (map[string]string, error) {
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("link not allowed in the package: %s", path)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("special file in the package: %s", path)
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ManifestName {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		result[rel] = Digest(b)
		return nil
	})
	return result, err
}

// A source checkout remains usable for development. A directory with package
// sentinels must have its manifest; losing it must not disable verification.
func VerifyIfPackaged(root string) error {
	_, err := ReadManifest(root)
	if os.IsNotExist(err) {
		for _, p := range []string{"bin/hive", "platform"} {
			if _, e := os.Lstat(filepath.Join(root, p)); e == nil {
				return fmt.Errorf("incomplete package: missing %s", ManifestName)
			} else if !os.IsNotExist(e) {
				return e
			}
		}
		return nil
	}
	if err != nil {
		return err
	}
	return verifyManifest(root)
}

// ReadManifest reads a package manifest without treating an absent manifest as an error.
// Callers that need package integrity should use VerifyIfPackaged.
func ReadManifest(root string) (Manifest, error) {
	m, _, err := ReadManifestData(root)
	return m, err
}

// ReadManifestData reads a package manifest like ReadManifest and also
// returns the exact bytes it decoded. An absent manifest returns the Lstat
// error unwrapped, so os.IsNotExist still recognizes it.
func ReadManifestData(root string) (Manifest, []byte, error) {
	info, err := os.Lstat(filepath.Join(root, ManifestName))
	if err != nil {
		return Manifest{}, nil, err
	}
	if !info.Mode().IsRegular() {
		return Manifest{}, nil, fmt.Errorf("package manifest is not a regular file")
	}
	data, err := os.ReadFile(filepath.Join(root, ManifestName))
	if err != nil {
		return Manifest{}, nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, nil, fmt.Errorf("invalid package manifest: %w", err)
	}
	return m, data, nil
}

func verifyManifest(root string) error {
	m, err := ReadManifest(root)
	if err != nil {
		return err
	}
	if m.Version != 1 || m.Platform != runtime.GOOS+"/"+runtime.GOARCH || len(m.SourceID) != 64 || len(m.Files) == 0 {
		return fmt.Errorf("incompatible package or invalid manifest")
	}
	if _, err := hex.DecodeString(m.SourceID); err != nil {
		return fmt.Errorf("invalid package identity")
	}
	for _, required := range []string{"bin/hive", "bin/hive.sha256", "install.sh", "platform", "content/guidance/global.md", "integrations/agent-profiles.json", "LICENSE", "THIRD_PARTY_NOTICES.md"} {
		if _, ok := m.Files[required]; !ok {
			return fmt.Errorf("incomplete package: %s", required)
		}
	}
	actual, err := Files(root)
	if err != nil {
		return err
	}
	if len(actual) != len(m.Files) {
		return fmt.Errorf("package has missing or additional files; download a complete copy")
	}
	for path, want := range m.Files {
		if filepath.IsAbs(path) || strings.Contains(path, "\\") || path == "." || filepath.ToSlash(filepath.Clean(path)) != path || strings.HasPrefix(path, "../") || actual[path] != want {
			return fmt.Errorf("invalid package integrity: %s", path)
		}
	}
	return nil
}
