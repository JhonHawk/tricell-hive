package viewer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
)

const (
	writePrefix   = ".write-"
	maxMetaBytes  = 64 << 10
	maxIndexBytes = 64 << 20
)

var (
	// ErrInvalidKey is returned for a session key that is not 64 lowercase
	// hexadecimal characters. It is checked before any path is built.
	ErrInvalidKey = errors.New("invalid session key")
	// ErrNotFound is returned for a session or file that does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidPath is returned for a file path that is not a safe relative
	// path, or that is a link.
	ErrInvalidPath = errors.New("invalid file path")

	keyPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Options configures Open.
type Options struct {
	// StateDir is the manager's state directory (absolute). Everything lives
	// under <StateDir>/viewer/.
	StateDir string
	// Now is the clock; nil means time.Now. Tests inject their own.
	Now func() time.Time
	// HiveProject and GitRemote are optional readers used to name and group
	// projects; see their types.
	HiveProject HiveProjectReader
	GitRemote   GitRemoteReader
}

// Store is the viewer's persistent state: the registry and the timeline
// records. One server owns a state directory at a time (viewer/server.lock,
// taken by the caller); within it every mutating method is serialized by a
// mutex. Methods are safe for concurrent use.
type Store struct {
	vdir   string
	ttl    time.Duration
	now    func() time.Time
	hive   HiveProjectReader
	remote GitRemoteReader

	// mu serializes startup and the timeline mutations.
	// Lock order: mu, then the flock on viewer/lock (registry only).
	mu         sync.Mutex
	removeTree func(string) error
	// mu also guards the first startup maintenance error until Run reports it.
	timelineStartupMaintenanceError error

	// viewMu guards the published timeline index.
	viewMu            sync.RWMutex
	timeline          map[string]*timelineRecord
	timelineIndex     []TimelineAPIIndexEntry
	timelineOwned     map[string]bool
	timelineUncertain bool
	timelineFault     func(phase, path string) error
}

// Open rejects previous or invalid viewer state before initializing a fresh
// v3 namespace and loading API records. It never imports or cleans old data.
func Open(o Options) (*Store, error) {
	if o.StateDir == "" || !filepath.IsAbs(o.StateDir) {
		return nil, fmt.Errorf("viewer: state directory must be an absolute path, got %q", o.StateDir)
	}
	s := &Store{
		vdir:       viewerDir(o.StateDir),
		now:        o.Now,
		hive:       o.HiveProject,
		remote:     o.GitRemote,
		removeTree: os.RemoveAll,
	}
	if s.now == nil {
		s.now = time.Now
	}
	if err := ValidateStateOptions(o); err != nil {
		return nil, err
	}
	if err := initializeState(o.StateDir); err != nil {
		return nil, err
	}
	if err := s.timelineLoad(); err != nil {
		return nil, err
	}
	return s, nil
}

// Dir returns <StateDir>/viewer.
func (s *Store) Dir() string { return s.vdir }

// TTL returns zero: content has no age-based expiry.
func (s *Store) TTL() time.Duration { return s.ttl }

// Startup validates the complete v3 inventory before cleaning owned incomplete
// writes and unreferenced blobs.
func (s *Store) Startup() error {
	if err := s.startupInventory(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.timelineMaintainLocked()
	if timelineMaintenanceCorrupt(err) {
		return err
	}
	// Serving valid records must not depend on successful visibility maintenance.
	// Run reports this first attempt even if its retry subsequently succeeds.
	if err != nil {
		s.timelineStartupMaintenanceError = err
	}
	return nil
}

func (s *Store) startupInventory() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ValidateState(filepath.Dir(s.vdir)); err != nil {
		return err
	}
	if err := s.timelineLoad(); err != nil {
		return err
	}
	return s.timelineCleanup()
}

func repositoryStatuses(reg Registry) []RepositoryStatus {
	out := []RepositoryStatus{}
	for _, r := range reg.Repositories {
		info, err := os.Stat(r.Root)
		out = append(out, RepositoryStatus{Root: r.Root, Absent: err != nil || !info.IsDir()})
	}
	return out
}

// writeJSONAtomic writes v as indented JSON to path with mode 0600: a temp
// file in the same directory, fsync, then rename.
func writeJSONAtomic(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, append(data, '\n'))
}

func writeFileAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), writePrefix+"*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, path)
	}
	if err != nil {
		os.Remove(tmp)
	}
	return err
}
