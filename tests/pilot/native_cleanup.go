package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const nativeOwnerMarker = ".hive-native-owner"

type nativeOwnedDirectory struct {
	directory, parent      *os.Root
	info                   os.FileInfo
	relative, leaf, marker string
	ancestors              map[string]os.FileInfo
}

// Called before credential import or model dispatch, while only the trusted
// runner has written this fresh home. OpenRoot is never used to reacquire an
// attacker-controlled cleanup path after those writers start.
func (g *guidanceVariant) prepareNativeCleanup() error {
	g.cleanupMu.Lock()
	defer g.cleanupMu.Unlock()
	if g.cleanupDone {
		return fmt.Errorf("native cleanup ownership unavailable after cleanup")
	}
	if g.nativeHome != nil {
		if len(g.nativeDirectories) == 3 {
			return nil
		}
		return fmt.Errorf("native cleanup ownership initialization incomplete")
	}
	info, err := os.Lstat(g.shadowHome)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("native home ownership unavailable")
	}
	home, err := os.OpenRoot(g.shadowHome)
	if err != nil {
		return fmt.Errorf("native home could not be pinned")
	}
	pinned, err := home.Stat(".")
	if err != nil || !os.SameFile(info, pinned) {
		home.Close()
		return fmt.Errorf("native home identity changed")
	}
	g.nativeHome, g.nativeHomeInfo = home, info
	g.nativeDirectories = map[string]*nativeOwnedDirectory{}
	for _, relative := range []string{".local", ".cache", ".local/share/opencode"} {
		if err := home.MkdirAll(relative, 0700); err != nil {
			return fmt.Errorf("native directory could not be prepared")
		}
		owned := &nativeOwnedDirectory{relative: relative, leaf: filepath.Base(relative), ancestors: map[string]os.FileInfo{}}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		for i := range parts {
			name := filepath.Join(parts[:i+1]...)
			stat, err := home.Lstat(name)
			if err != nil || !stat.IsDir() || stat.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("native directory is not owned regular storage")
			}
			owned.ancestors[name] = stat
		}
		owned.info = owned.ancestors[relative]
		owned.parent, err = home.OpenRoot(filepath.Dir(relative))
		if err != nil {
			return fmt.Errorf("native parent could not be pinned")
		}
		owned.directory, err = home.OpenRoot(relative)
		if err != nil {
			owned.parent.Close()
			return fmt.Errorf("native directory could not be pinned")
		}
		nonce := make([]byte, 32)
		if _, err := rand.Read(nonce); err != nil {
			return fmt.Errorf("native ownership marker unavailable")
		}
		owned.marker = hex.EncodeToString(nonce)
		file, err := owned.directory.OpenFile(nativeOwnerMarker, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			owned.directory.Close()
			owned.parent.Close()
			return fmt.Errorf("native ownership marker conflict")
		}
		_, err = file.WriteString(owned.marker)
		closeErr := file.Close()
		if err != nil || closeErr != nil {
			owned.directory.Close()
			owned.parent.Close()
			return fmt.Errorf("native ownership marker write failed")
		}
		g.nativeDirectories[relative] = owned
	}
	return nil
}

func (g *guidanceVariant) nativeIdentityValid(owned *nativeOwnedDirectory, includeLeaf bool) bool {
	info, err := os.Lstat(g.shadowHome)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !os.SameFile(info, g.nativeHomeInfo) {
		return false
	}
	for name, expected := range owned.ancestors {
		if !includeLeaf && name == owned.relative {
			continue
		}
		actual, err := g.nativeHome.Lstat(name)
		if err != nil || actual.Mode()&os.ModeSymlink != 0 || !os.SameFile(expected, actual) {
			return false
		}
	}
	return true
}

func (g *guidanceVariant) removeOwnedNativeDirectory(path string) error {
	if g.nativeHome == nil {
		return fmt.Errorf("native cleanup has no pinned ownership")
	}
	relative, err := filepath.Rel(g.shadowHome, path)
	if err != nil {
		return fmt.Errorf("native cleanup path is unverified")
	}
	owned := g.nativeDirectories[relative]
	if owned == nil || !g.nativeIdentityValid(owned, true) {
		return fmt.Errorf("native cleanup identity or ancestors changed")
	}
	markerInfo, err := owned.directory.Lstat(nativeOwnerMarker)
	if err != nil || !markerInfo.Mode().IsRegular() {
		return fmt.Errorf("native ownership marker changed")
	}
	marker, err := owned.directory.ReadFile(nativeOwnerMarker)
	if err != nil || string(marker) != owned.marker {
		return fmt.Errorf("native ownership marker changed")
	}
	// Atomically detach the name, then verify the detached inode. A raced-in
	// replacement is retained, not recursively deleted. The recursion below uses
	// the already-pinned owned directory, never the reopened detached pathname.
	quarantine := ".hive-native-cleanup-" + owned.marker
	if err := owned.parent.Rename(owned.leaf, quarantine); err != nil {
		return fmt.Errorf("native cleanup detachment failed")
	}
	actual, err := owned.parent.Lstat(quarantine)
	if err != nil || actual.Mode()&os.ModeSymlink != 0 || !os.SameFile(actual, owned.info) {
		return fmt.Errorf("native replacement retained in owned quarantine")
	}
	file, err := owned.directory.Open(".")
	if err != nil {
		return fmt.Errorf("pinned native directory unavailable")
	}
	entries, err := file.ReadDir(-1)
	file.Close()
	if err != nil {
		return fmt.Errorf("pinned native directory listing failed")
	}
	for _, entry := range entries {
		if entry.Name() != nativeOwnerMarker {
			if err := owned.directory.RemoveAll(entry.Name()); err != nil {
				return fmt.Errorf("pinned native cleanup failed")
			}
		}
	}
	// Only remove the empty detached entry, never recurse through its name. Even
	// a subsequent symlink replacement cannot cause deletion of its target.
	actual, err = owned.parent.Lstat(quarantine)
	if err != nil || actual.Mode()&os.ModeSymlink != 0 || !os.SameFile(actual, owned.info) || !g.nativeIdentityValid(owned, false) {
		return fmt.Errorf("native cleanup replacement retained")
	}
	if err := owned.directory.Remove(nativeOwnerMarker); err != nil {
		return fmt.Errorf("native marker cleanup failed")
	}
	if err := owned.parent.Remove(quarantine); err != nil {
		return fmt.Errorf("native detached directory cleanup failed")
	}
	if !g.nativeIdentityValid(owned, false) {
		return fmt.Errorf("native cleanup ancestors changed during deletion")
	}
	return nil
}

func (g *guidanceVariant) closeNativeCleanup() {
	for _, owned := range g.nativeDirectories {
		owned.directory.Close()
		owned.parent.Close()
	}
	if g.nativeHome != nil {
		g.nativeHome.Close()
	}
}
