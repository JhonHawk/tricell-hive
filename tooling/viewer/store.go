package viewer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	tmpPrefix     = ".tmp-"
	oldPrefix     = ".old-"
	writePrefix   = ".write-"
	maxMetaBytes  = 64 << 10
	maxIndexBytes = 64 << 20
)

var (
	// ErrInvalidKey is returned for a session key that is not 64 lowercase
	// hexadecimal characters. It is checked before any path is built.
	ErrInvalidKey = errors.New("invalid session key")
	// ErrExcluded denies normal document and file reads for a tracked source.
	ErrExcluded = errors.New("session source is tracked")
	// ErrNotFound is returned for a session or file that does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidPath is returned for a file path that breaks the rules of
	// walkthrough.json paths, or that is a link.
	ErrInvalidPath = errors.New("invalid file path")

	keyPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// Options configures Open.
type Options struct {
	// StateDir is the manager's state directory (absolute). Everything lives
	// under <StateDir>/viewer/.
	StateDir string
	// TTL is deprecated and ignored. Retention is manual for every copy.
	TTL time.Duration
	// Now is the clock; nil means time.Now. Tests inject their own.
	Now func() time.Time
	// HiveProject and GitRemote are optional readers used to name and group
	// projects; see their types.
	HiveProject HiveProjectReader
	GitRemote   GitRemoteReader
	// SourceTracking classifies current manifest paths in one batch; nil is unknown.
	SourceTracking SourceTrackingReader
}

// Store is the viewer's persistent state: registry, session copies and the
// index. One server owns a state directory at a time (viewer/server.lock,
// taken by the caller); within it every mutating method is serialized by a
// mutex, and View reads a published snapshot without taking it. Methods are
// safe for concurrent use.
type Store struct {
	vdir     string
	ttl      time.Duration
	now      func() time.Time
	hive     HiveProjectReader
	remote   GitRemoteReader
	tracking SourceTrackingReader

	// afterValidate, when set by a test, runs in ingest between validating a
	// folder and reading it again to hash and copy it.
	afterValidate func(folder string)

	// mu serializes the apply phase of a scan, purge, delete, pin, archive and startup.
	// Lock order: mu, then the flock on viewer/lock (registry only).
	mu            sync.Mutex
	lastPersisted []byte
	// Filesystem boundaries use the production implementations installed by Open.
	writeAtomic func(string, []byte) error
	removeFile  func(string) error
	removeTree  func(string) error
	renamePath  func(string, string) error

	// onHash, when set by a test, is called each time a source's files are hashed.
	onHash func(folder string)

	// batchSize and batchBytes bound how many inspected candidates, and how
	// many bytes of their manifests, Sync keeps open before it applies them
	// (zero means the defaults). onBatch, when set by a test, receives the
	// size of each batch.
	batchSize  int
	batchBytes int64
	onBatch    func(n int)

	// fps remembers, per source folder, the (size, mtime, inode) of its
	// manifest and referenced files when they were last hashed, and the
	// resulting key and content hash. A source whose files all still match,
	// and whose copy is current, is not validated or hashed again. In memory
	// only: a restart hashes once more.
	fpMu sync.Mutex
	fps  map[string]*fingerprint

	// projects caches the resolved project of each repository root for
	// projectRefresh, so a pass does not run the readers (two git
	// processes, AGENTS.md, .engram) for every repository every poll.
	// projRepos is the registry's roots when the cache was last valid.
	projMu    sync.Mutex
	projects  map[string]cachedProject
	projRepos string

	// viewMu guards idx, which is replaced, never modified in place.
	viewMu            sync.RWMutex
	idx               *indexData
	timeline          map[string]*timelineRecord
	timelineIndex     []TimelineAPIIndexEntry
	timelineOwned     map[string]bool
	timelineUncertain bool
	timelineFault     func(phase, path string) error
}

// Open rejects previous or invalid viewer state before initializing a fresh
// v2 namespace and loading API records. It never imports or cleans legacy data.
func Open(o Options) (*Store, error) {
	if o.StateDir == "" || !filepath.IsAbs(o.StateDir) {
		return nil, fmt.Errorf("viewer: state directory must be an absolute path, got %q", o.StateDir)
	}
	s := &Store{
		vdir:        viewerDir(o.StateDir),
		ttl:         0,
		now:         o.Now,
		hive:        o.HiveProject,
		remote:      o.GitRemote,
		tracking:    o.SourceTracking,
		writeAtomic: writeFileAtomic, removeFile: os.Remove, removeTree: os.RemoveAll, renamePath: os.Rename,
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
	s.idx = newIndexData()
	if err := s.timelineLoad(); err != nil {
		return nil, err
	}
	return s, nil
}

// Dir returns <StateDir>/viewer.
func (s *Store) Dir() string { return s.vdir }

// TTL returns zero: content has no age-based expiry.
func (s *Store) TTL() time.Duration { return s.ttl }

func (s *Store) sessionsDir() string { return filepath.Join(s.vdir, "sessions") }

func (s *Store) sessionDir(key string) string { return filepath.Join(s.sessionsDir(), key) }

func (s *Store) indexPath() string { return filepath.Join(s.vdir, "index.json") }

// Startup validates the complete v2 inventory before cleaning owned incomplete
// writes and unreferenced blobs. Legacy cleanup and index rebuilding never run.
func (s *Store) Startup() error {
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

// cleanup removes temporary copy directories and stray atomic-write files,
// and finishes a replacement that crashed between its two renames: an
// .old-<key> directory returns as <key> when the live copy is missing.
func (s *Store) cleanup() error {
	entries, err := os.ReadDir(s.sessionsDir())
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		path := filepath.Join(s.sessionsDir(), name)
		switch {
		case strings.HasPrefix(name, tmpPrefix):
			if err := os.RemoveAll(path); err != nil {
				return err
			}
		case strings.HasPrefix(name, oldPrefix):
			key := strings.TrimPrefix(name, oldPrefix)
			if _, err := os.Lstat(s.sessionDir(key)); keyPattern.MatchString(key) && errors.Is(err, fs.ErrNotExist) {
				if err := os.Rename(path, s.sessionDir(key)); err != nil {
					return err
				}
			} else if err := os.RemoveAll(path); err != nil {
				return err
			}
		}
	}
	for _, dir := range []string{s.vdir, s.sessionsDir()} {
		stray, _ := filepath.Glob(filepath.Join(dir, writePrefix+"*"))
		for _, p := range stray {
			os.Remove(p)
		}
	}
	return nil
}

type meta struct {
	Source      string    `json:"source"`
	ContentHash string    `json:"content_hash"`
	IngestedAt  time.Time `json:"ingested_at"`
	Pinned      bool      `json:"pinned"`
	Archived    bool      `json:"archived,omitempty"`
	ProjectKey  string    `json:"project_key,omitempty"`
	ProjectName string    `json:"project_name,omitempty"`
}

func readMeta(dir string) (meta, error) {
	var m meta
	f, err := os.Open(filepath.Join(dir, "meta.json"))
	if err != nil {
		return m, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxMetaBytes+1))
	if err != nil {
		return m, err
	}
	if len(data) > maxMetaBytes {
		return m, errors.New("meta.json is too large")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return m, fmt.Errorf("meta.json: %w", err)
	}
	if !keyPattern.MatchString(m.ContentHash) || m.Source == "" || m.IngestedAt.IsZero() {
		return m, errors.New("meta.json: incomplete")
	}
	return m, nil
}

// storedSessionStatus distinguishes a missing field from invalid null or empty
// values when rebuilding the index without validating all stored resources.
type storedSessionStatus string

func (s *storedSessionStatus) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil || (value != "in_progress" && value != "closed") {
		return fail("session.status", "must be in_progress or closed")
	}
	*s = storedSessionStatus(value)
	return nil
}

// summary is the part of a stored manifest the index shows.
type summary struct {
	Title     string `json:"title"`
	Kind      string `json:"kind"`
	Ticket    string `json:"ticket"`
	CreatedAt string `json:"created_at"`
	Session   struct {
		ID     string              `json:"id"`
		Host   string              `json:"host"`
		Status storedSessionStatus `json:"status"`
	} `json:"session"`
}

func readSummary(dir string) (summary, error) {
	var sm summary
	data, err := readBounded(filepath.Join(dir, manifestFile), maxManifest)
	if err != nil {
		return sm, err
	}
	if err := json.Unmarshal(data, &sm); err != nil {
		return sm, fmt.Errorf("%s: %w", manifestFile, err)
	}
	sm.Ticket = strings.TrimSpace(sm.Ticket)
	if sm.Session.Status == "" {
		sm.Session.Status = "unknown"
	}
	return sm, nil
}

func readBounded(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("%s is not a regular file within %d bytes", filepath.Base(path), limit)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, limit+1))
}

func sessionFrom(key string, m meta, sm summary) Session {
	return Session{
		Key: key, Source: m.Source, ContentHash: m.ContentHash, IngestedAt: m.IngestedAt.UTC(), Pinned: m.Pinned, Archived: m.Archived,
		SessionID: sm.Session.ID, SessionStatus: string(sm.Session.Status), Host: sm.Session.Host, Ticket: sm.Ticket, Title: sm.Title,
		Kind: sm.Kind, CreatedAt: sm.CreatedAt,
		Repository: repoRootOf(m.Source), ProjectKey: m.ProjectKey, ProjectName: m.ProjectName,
	}
}

// load builds the index from disk: sessions from sessions/*/meta.json (the
// physical source of truth), tombstones and ingest
// errors from a readable index.json. For older metadata without project identity,
// a matching index entry supplies the previously resolved identity.
func (s *Store) load() *indexData {
	d := newIndexData()
	previous := map[string]Session{}
	if prev, ok := s.readIndexFile(); ok {
		for _, sess := range prev.Sessions {
			previous[sess.Key] = sess
		}
		d.Tombstones = append(d.Tombstones, prev.Tombstones...)
		for _, e := range prev.Errors {
			if !strings.HasPrefix(e.Path, s.sessionsDir()+string(filepath.Separator)) {
				d.Errors = append(d.Errors, e)
			}
		}
	}
	entries, err := os.ReadDir(s.sessionsDir())
	if err != nil {
		d.Errors = append(d.Errors, IngestError{Path: s.sessionsDir(), Reason: err.Error()})
	}
	for _, e := range entries {
		if !e.IsDir() || !keyPattern.MatchString(e.Name()) {
			continue
		}
		dir := s.sessionDir(e.Name())
		m, err := readMeta(dir)
		var sm summary
		if err == nil {
			sm, err = readSummary(dir)
		}
		if err != nil {
			d.Errors = append(d.Errors, IngestError{Path: dir, Reason: err.Error()})
			continue
		}
		sess := sessionFrom(e.Name(), m, sm)
		if prev, ok := previous[sess.Key]; ok && sess.ProjectKey == "" && prev.Source == sess.Source && prev.Repository == repoRootOf(sess.Source) {
			sess.Repository, sess.ProjectKey, sess.ProjectName = prev.Repository, prev.ProjectKey, prev.ProjectName
		}
		d.Sessions = append(d.Sessions, sess)
	}
	if reg, err := readRegistry(s.vdir); err == nil {
		d.Repositories = repositoryStatuses(reg)
	}
	s.classifySessions(d, s.classifySources(context.Background(), sessionSourcePaths(d)))
	s.enrich(d)
	_ = s.reconcileGroups(d, false)
	s.recoverPendingGroups(d, false)
	d.sorted()
	return d
}

func (s *Store) readIndexFile() (*indexData, bool) {
	data, err := readBounded(s.indexPath(), maxIndexBytes)
	if err != nil {
		return nil, false
	}
	var d indexData
	if err := json.Unmarshal(data, &d); err != nil || d.Version != indexVersion {
		return nil, false
	}
	return &d, true
}

func repositoryStatuses(reg Registry) []RepositoryStatus {
	out := []RepositoryStatus{}
	for _, r := range reg.Repositories {
		info, err := os.Stat(r.Root)
		out = append(out, RepositoryStatus{Root: r.Root, Absent: err != nil || !info.IsDir()})
	}
	return out
}

// commit publishes d and persists it to index.json when it changed. The
// caller holds s.mu (or is Open).
func (s *Store) commit(d *indexData) error {
	if err := s.reconcileGroups(d, true); err != nil {
		s.viewMu.Lock()
		s.idx = d
		s.viewMu.Unlock()
		return err
	}
	s.recoverPendingGroups(d, false)
	d.sorted()
	previous := map[string]Session{}
	if s.lastPersisted != nil {
		for _, sess := range s.current().Sessions {
			previous[sess.Key] = sess
		}
	}
	for _, sess := range d.Sessions {
		if prev, ok := previous[sess.Key]; ok && prev.ProjectKey == sess.ProjectKey && prev.ProjectName == sess.ProjectName {
			continue
		}
		m, err := readMeta(s.sessionDir(sess.Key))
		if err != nil {
			return err
		}
		if m.ProjectKey == sess.ProjectKey && m.ProjectName == sess.ProjectName {
			continue
		}
		m.ProjectKey, m.ProjectName = sess.ProjectKey, sess.ProjectName
		if err := writeMeta(s.sessionDir(sess.Key), m); err != nil {
			return err
		}
	}
	s.viewMu.Lock()
	s.idx = d
	s.viewMu.Unlock()
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if bytes.Equal(data, s.lastPersisted) {
		return nil
	}
	if err := s.writeAtomic(s.indexPath(), data); err != nil {
		return err
	}
	s.lastPersisted = data
	return nil
}

func (s *Store) current() *indexData {
	s.viewMu.RLock()
	defer s.viewMu.RUnlock()
	return s.idx
}

func validKey(key string) error {
	if !keyPattern.MatchString(key) {
		return ErrInvalidKey
	}
	return nil
}

// removeCopy deletes sessions/<key> atomically: it is renamed away first, so
// a crash never leaves a half-deleted copy under its real name. It touches
// nothing outside sessions/.
func (s *Store) removeCopy(key string) error {
	dir := s.sessionDir(key)
	gone := filepath.Join(s.sessionsDir(), tmpPrefix+"del-"+key)
	s.removeTree(gone)
	if err := s.renamePath(dir, gone); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return err
	}
	return s.removeTree(gone)
}

func (s *Store) addTombstone(d *indexData, sess Session) {
	d.dropTombstone(sess.Key)
	d.Tombstones = append(d.Tombstones, Tombstone{
		Key: sess.Key, ContentHash: sess.ContentHash, Source: sess.Source, DeletedAt: s.now().UTC(),
	})
}

// Pin changes the keep preference of the logical session containing key.
func (s *Store) Pin(key string, pinned bool) error {
	return s.updateGroup(key, func(g *sessionGroup) { g.Pinned = pinned })
}

// Archive changes visibility for the logical session containing key.
func (s *Store) Archive(key string, archived bool) error {
	return s.updateGroup(key, func(g *sessionGroup) { g.Archived = archived })
}

// Purge is a compatibility no-op: retained content is removed only explicitly.
func (s *Store) Purge() ([]string, error) { return nil, nil }

// normalReadAllowed checks the published classification without consulting Git.
// Management operations intentionally retain access to excluded copies.
func (s *Store) normalReadAllowed(key string) error {
	d := s.current()
	if i, ok := d.session(key); ok && d.Sessions[i].SourceStatus == "tracked" {
		return ErrExcluded
	}
	return nil
}

// Manifest returns the stored walkthrough.json bytes of a session. Tracked
// sources return ErrExcluded before any document bytes are read.
func (s *Store) Manifest(key string) ([]byte, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := s.normalReadAllowed(key); err != nil {
		return nil, err
	}
	data, err := readBounded(filepath.Join(s.sessionDir(key), manifestFile), maxManifest)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, ErrNotFound
	}
	return data, err
}

// OpenFile opens a stored capture or recording for serving. rel is a
// slash-separated path under the session; it must satisfy the same rules as
// walkthrough.json paths (safe segments, an allowed image or .mp4 extension,
// no link anywhere on the way), so meta.json and walkthrough.json are not
// reachable through it. The caller closes the file. Errors: ErrInvalidKey,
// ErrInvalidPath, ErrNotFound, ErrExcluded.
func (s *Store) OpenFile(key, rel string) (*os.File, error) {
	if err := validKey(key); err != nil {
		return nil, err
	}
	if err := s.normalReadAllowed(key); err != nil {
		return nil, err
	}
	segments := strings.Split(rel, "/")
	for _, seg := range segments {
		if seg == "." || seg == ".." || !segmentPattern.MatchString(seg) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidPath, rel)
		}
	}
	if ext := extension(segments[len(segments)-1]); !contains(captureExtensions, ext) && !contains(recordingExtensions, ext) {
		return nil, fmt.Errorf("%w: extension of %q", ErrInvalidPath, rel)
	}
	root, err := os.OpenRoot(s.sessionDir(key))
	if err != nil {
		return nil, ErrNotFound
	}
	defer root.Close()
	for i := range segments {
		info, err := root.Lstat(filepath.FromSlash(strings.Join(segments[:i+1], "/")))
		if err != nil {
			return nil, ErrNotFound
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return nil, fmt.Errorf("%w: link in %q", ErrInvalidPath, rel)
		}
		if i == len(segments)-1 && !info.Mode().IsRegular() {
			return nil, ErrNotFound
		}
	}
	f, err := root.Open(filepath.FromSlash(rel))
	if err != nil {
		return nil, ErrNotFound
	}
	return f, nil
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
