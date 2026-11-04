package viewer

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const stateMarker = "storage.json"

type viewerStorage struct {
	StorageVersion int `json:"storage_version"`
}

// ValidateState performs a strictly read-only preflight before even lock/log
// creation. Manager files outside viewer/ are deliberately outside this boundary.
func ValidateState(stateDir string) error { return ValidateStateOptions(Options{StateDir: stateDir}) }

// ValidateStateOptions additionally re-resolves source identity with runtime readers.
func ValidateStateOptions(o Options) error {
	stateDir := o.StateDir
	if !filepath.IsAbs(stateDir) {
		return timelineFailure("invalid_storage", 500)
	}
	vdir := viewerDir(stateDir)
	info, err := os.Lstat(vdir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || !info.IsDir() {
		return timelineFailure("invalid_storage", 500)
	}
	entries, err := os.ReadDir(vdir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	marked := false
	for _, e := range entries {
		if e.Name() == stateMarker {
			marked = true
		}
	}
	if !marked {
		return timelineFailure("invalid_storage", 500)
	}
	raw, err := readStateFile(filepath.Join(vdir, stateMarker), 1024)
	if err != nil {
		return err
	}
	var marker viewerStorage
	if decodeTimelineJSONBounded(raw, &marker, 1024) != nil {
		return timelineFailure("invalid_storage", 500)
	}
	if marker.StorageVersion != 3 {
		return timelineFailure("invalid_storage", 500)
	}
	for _, e := range entries {
		switch e.Name() {
		case "registry.json":
			if _, err = readStateFile(filepath.Join(vdir, e.Name()), maxIndexBytes); err != nil {
				return err
			}
		case stateMarker, "lock", "server.lock", "server.json", "server.log":
			if _, err = readStateFile(filepath.Join(vdir, e.Name()), maxMetaBytes); err != nil {
				return err
			}
		case "timeline-api-v3":
			if err = validateTimelineInventory(filepath.Join(vdir, e.Name())); err != nil {
				return err
			}
		default:
			if strings.HasPrefix(e.Name(), writePrefix) {
				if _, err = readStateFile(filepath.Join(vdir, e.Name()), maxIndexBytes); err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				continue
			}
			return timelineFailure("invalid_storage", 500)
		}
	}
	_, err = readRegistry(vdir)
	if err != nil {
		return err
	}
	s := &Store{vdir: vdir, now: time.Now, remote: o.GitRemote, hive: o.HiveProject, timeline: map[string]*timelineRecord{}}
	return s.timelineLoad()
}

// readStateFile rejects links, shared inodes, oversize/truncated input and races
// between the name inspection and opened descriptor inspection.
func readStateFile(path string, limit int64) ([]byte, error) {
	return readStateFileAttempt(path, limit, 3)
}
func readStateFileAttempt(path string, limit int64, attempts int) ([]byte, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > limit {
		return nil, timelineFailure("invalid_storage", 500)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	after, err := f.Stat()
	if err != nil {
		return nil, err
	}
	stat, ok := after.Sys().(*syscall.Stat_t)
	if !os.SameFile(before, after) && attempts > 1 {
		f.Close()
		return readStateFileAttempt(path, limit, attempts-1)
	}
	if !os.SameFile(before, after) || !after.Mode().IsRegular() || !ok || stat.Nlink != 1 {
		return nil, timelineFailure("invalid_storage", 500)
	}
	raw, err := readBoundedDescriptor(f, limit)
	return raw, err
}

func validateTimelineInventory(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil || !info.IsDir() {
		return timelineFailure("invalid_storage", 500)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if strings.HasSuffix(e.Name(), ".json") && keyPattern.MatchString(strings.TrimSuffix(e.Name(), ".json")) || strings.HasPrefix(e.Name(), ".timeline-write-") {
			if _, err = readStateFile(path, timelineMaxRecordBytes+TimelineMaintenanceMaxBytes); err != nil {
				return err
			}
			continue
		}
		if !keyPattern.MatchString(e.Name()) {
			return timelineFailure("invalid_storage", 500)
		}
		if _, err := os.Lstat(path + ".json"); err != nil {
			return timelineFailure("corrupt_storage", 500)
		}
		info, err = os.Lstat(path)
		if err != nil || !info.IsDir() {
			return timelineFailure("invalid_storage", 500)
		}
		subs, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, sub := range subs {
			if sub.Name() != "documents" && sub.Name() != "artifacts" {
				return timelineFailure("invalid_storage", 500)
			}
			subpath := filepath.Join(path, sub.Name())
			info, err = os.Lstat(subpath)
			if err != nil || !info.IsDir() {
				return timelineFailure("invalid_storage", 500)
			}
			files, err := os.ReadDir(subpath)
			if err != nil {
				return err
			}
			for _, file := range files {
				if !timelineOrphanName(file.Name(), sub.Name()) && !strings.HasPrefix(file.Name(), ".timeline-write-") {
					return timelineFailure("invalid_storage", 500)
				}
				if _, err = readStateFile(filepath.Join(subpath, file.Name()), maxCapture); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func initializeState(stateDir string) error {
	if err := ValidateState(stateDir); err != nil {
		return err
	}
	dir := viewerDir(stateDir)
	if _, err := os.Lstat(filepath.Join(dir, stateMarker)); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return err
	}
	// Publish a complete namespace atomically. Concurrent first registration
	// cannot observe a partial marker or mistake its write temporary for legacy.
	staged, err := os.MkdirTemp(stateDir, ".viewer-init-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staged)
	if err = writeJSONAtomic(filepath.Join(staged, stateMarker), viewerStorage{StorageVersion: 3}); err != nil {
		return err
	}
	stagedDir, err := os.Open(staged)
	if err != nil {
		return err
	}
	syncErr := stagedDir.Sync()
	stagedDir.Close()
	if syncErr != nil {
		return syncErr
	}
	if err = os.Rename(staged, dir); err != nil {
		if check := ValidateState(stateDir); check != nil {
			return check
		}
		if _, check := os.Lstat(filepath.Join(dir, stateMarker)); check != nil {
			return err
		}
	}
	parent, err := os.Open(stateDir)
	if err != nil {
		return err
	}
	defer parent.Close()
	return parent.Sync()
}

func readBoundedDescriptor(f *os.File, limit int64) ([]byte, error) {
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, timelineFailure("invalid_storage", 500)
	}
	return raw, nil
}

// PrepareState initializes only a fresh, validated v3 namespace.
func PrepareState(stateDir string) error { return initializeState(stateDir) }
