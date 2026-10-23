package viewer

import (
	"os"
	"path/filepath"
)

// SourceFingerprint identifies the current validated revision of a source.
// It contains no store state or access credentials.
type SourceFingerprint struct {
	Source, SessionID, Key, ContentHash, SessionStatus string
}

// FingerprintSource validates and hashes a source without opening a Store or
// writing anything. Validation and hashing share one rooted directory handle.
func FingerprintSource(folder string) (SourceFingerprint, error) {
	folder, err := filepath.Abs(folder)
	if err != nil {
		return SourceFingerprint{}, err
	}
	folder, err = filepath.EvalSymlinks(folder)
	if err != nil {
		return SourceFingerprint{}, err
	}
	root, err := os.OpenRoot(folder)
	if err != nil {
		return SourceFingerprint{}, err
	}
	defer root.Close()
	m, err := validateRoot(root)
	if err != nil {
		return SourceFingerprint{}, err
	}
	hash, err := hashFiles(root, m)
	if err != nil {
		return SourceFingerprint{}, err
	}
	return SourceFingerprint{Source: folder, SessionID: m.SessionID, Key: sessionKey(folder, m.SessionID), ContentHash: hash, SessionStatus: m.SessionStatus}, nil
}
