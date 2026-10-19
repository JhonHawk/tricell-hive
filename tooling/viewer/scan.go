package viewer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Bounds on what a failed ingestion can put in the index (H4).
const (
	maxIngestErrorsPerRepository = 20
	maxErrorPath                 = 512
)

// SyncResult counts what one Sync pass did.
type SyncResult struct {
	// Added is the number of new copies, Replaced the number of copies whose
	// content changed, Errors the number of sources that were rejected.
	Added, Replaced, Errors int
}

// RunOptions configures Run.
type RunOptions struct {
	// PollInterval is the time between scans (default 5 s); PurgeInterval
	// the time between purges (default 1 h).
	PollInterval, PurgeInterval time.Duration
	// OnError, when set, receives failures of a scan or purge. They never
	// stop Run.
	OnError func(error)
}

// Run is the server's background loop. It calls Startup, purges, scans, then
// repeats both on their intervals until ctx is cancelled (it then returns
// nil). A failure in a pass goes to OnError and the loop continues.
func (s *Store) Run(ctx context.Context, o RunOptions) error {
	if o.PollInterval <= 0 {
		o.PollInterval = 5 * time.Second
	}
	if o.PurgeInterval <= 0 {
		o.PurgeInterval = time.Hour
	}
	report := func(err error) {
		if err != nil && ctx.Err() == nil && o.OnError != nil {
			o.OnError(err)
		}
	}
	if err := s.Startup(); err != nil {
		return err
	}
	_, err := s.Purge()
	report(err)
	_, err = s.Sync(ctx)
	report(err)
	poll, purge := time.NewTicker(o.PollInterval), time.NewTicker(o.PurgeInterval)
	defer poll.Stop()
	defer purge.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-poll.C:
			_, err := s.Sync(ctx)
			report(err)
		case <-purge.C:
			_, err := s.Purge()
			report(err)
		}
	}
}

// sessionKey identifies a copy: one per source folder and session id.
func sessionKey(folder, sessionID string) string {
	h := sha256.Sum256([]byte(folder + "\x00" + sessionID))
	return hex.EncodeToString(h[:])
}

// combineHash is content_hash: SHA-256 over the manifest bytes followed, for
// each referenced file in path order, by its path and its own SHA-256.
func combineHash(raw []byte, files []string, fileHashes map[string]string) string {
	h := sha256.New()
	h.Write(raw)
	for _, f := range files {
		fmt.Fprintf(h, "\x00%s\x00%s", f, fileHashes[f])
	}
	return hex.EncodeToString(h.Sum(nil))
}

// fileLimit is the size cap of a referenced file by extension.
func fileLimit(rel string) int64 {
	if contains(recordingExtensions, extension(filepath.Base(rel))) {
		return maxRecording
	}
	return maxCapture
}

// copyFile copies rel from root to dst (nil to only hash), returning its
// SHA-256. It refuses anything that is not a regular file or exceeds the cap.
func copyFile(root *os.Root, rel, dst string) (string, error) {
	src, err := root.Open(filepath.FromSlash(rel))
	if err != nil {
		return "", err
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", rel)
	}
	limit := fileLimit(rel)
	var w io.Writer = io.Discard
	var out *os.File
	if dst != "" {
		if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
			return "", err
		}
		out, err = os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return "", err
		}
		w = out
	}
	var h hash.Hash = sha256.New()
	n, err := io.Copy(io.MultiWriter(w, h), io.LimitReader(src, limit+1))
	if out != nil {
		if cerr := out.Close(); err == nil {
			err = cerr
		}
	}
	if err != nil {
		return "", err
	}
	if n > limit {
		return "", fmt.Errorf("%s exceeds %d MiB", rel, limit>>20)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func hashFiles(root *os.Root, m *Manifest) (string, error) {
	var err error
	sums := map[string]string{}
	for _, f := range m.Files {
		if sums[f], err = copyFile(root, f, ""); err != nil {
			return "", err
		}
	}
	return combineHash(m.Raw, m.Files, sums), nil
}

// install writes the copy of one source as sessions/<key>, replacing any
// previous copy and carrying its pinned flag over. The copy is assembled in
// a temporary directory and swapped in with renames, so a reader sees the
// old or the new copy, never a mix; Startup repairs a crash in between.
func (s *Store) install(key, folder string, root *os.Root, m *Manifest, hash string, now time.Time) (Session, error) {
	tmp, err := os.MkdirTemp(s.sessionsDir(), tmpPrefix)
	if err != nil {
		return Session{}, err
	}
	defer os.RemoveAll(tmp)
	if err := os.WriteFile(filepath.Join(tmp, manifestFile), m.Raw, 0o600); err != nil {
		return Session{}, err
	}
	sums := map[string]string{}
	for _, f := range m.Files {
		if sums[f], err = copyFile(root, f, filepath.Join(tmp, filepath.FromSlash(f))); err != nil {
			return Session{}, err
		}
	}
	if combineHash(m.Raw, m.Files, sums) != hash {
		return Session{}, errors.New("source changed while it was being copied; it will be retried")
	}
	final := s.sessionDir(key)
	pinned := false
	if old, err := readMeta(final); err == nil {
		pinned = old.Pinned
	}
	mt := meta{Source: folder, ContentHash: hash, IngestedAt: now.UTC(), Pinned: pinned}
	if err := writeJSONAtomic(filepath.Join(tmp, "meta.json"), mt); err != nil {
		return Session{}, err
	}
	old := filepath.Join(s.sessionsDir(), oldPrefix+key)
	hadOld := false
	if _, err := os.Lstat(final); err == nil {
		os.RemoveAll(old)
		if err := os.Rename(final, old); err != nil {
			return Session{}, err
		}
		hadOld = true
	}
	if err := os.Rename(tmp, final); err != nil {
		if hadOld {
			os.Rename(old, final)
		}
		return Session{}, err
	}
	if hadOld {
		os.RemoveAll(old)
	}
	sm := summary{Title: m.Title, Kind: m.Kind, Ticket: m.Ticket, CreatedAt: m.CreatedAt}
	sm.Session.ID, sm.Session.Host = m.SessionID, m.SessionHost
	return sessionFrom(key, mt, sm), nil
}

// Sync scans every registered repository once. For each
// <root>/_support/workspace/<folder>/walkthrough.json it validates the
// manifest, and copies the folder's manifest and referenced files into
// sessions/<key>/ when the content is new or changed. Invalid sources are
// listed as ingest errors and do not stop the pass; a source the user
// deleted or that was purged stays out until its content changes; a source
// whose manifest is older than the TTL is skipped; a registered repository
// missing from disk is flagged absent and its copies are kept. Links are
// never followed. A cancelled ctx commits what was done and returns its
// error.
func (s *Store) Sync(ctx context.Context) (SyncResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var res SyncResult
	reg, err := s.readRegistryShared()
	if err != nil {
		return res, err
	}
	now := s.now()
	d := s.current().clone()
	d.Repositories = repositoryStatuses(reg)
	var problems []IngestError
	var perRepo, omitted int
	addProblem := func(path string, err error) {
		res.Errors++
		if perRepo >= maxIngestErrorsPerRepository {
			omitted++
			return
		}
		perRepo++
		problems = append(problems, IngestError{Path: cleanText(path, maxErrorPath), Reason: cleanText(err.Error(), maxErrorText)})
	}
	var cancelled error
	for i, r := range reg.Repositories {
		if d.Repositories[i].Absent {
			continue
		}
		perRepo, omitted = 0, 0
		ws, names, errs := openWorkspace(r.Root)
		for _, e := range errs {
			addProblem(e.Path, errors.New(e.Reason))
		}
		if ws != nil {
			wsDir := filepath.Join(r.Root, "_support", "workspace")
			for _, name := range names {
				if cancelled = ctx.Err(); cancelled != nil {
					break
				}
				s.ingest(ws, name, wsDir, now, d, &res, addProblem)
			}
			ws.Close()
		}
		if omitted > 0 {
			problems = append(problems, IngestError{Path: r.Root, Reason: fmt.Sprintf("%d more ingest errors omitted", omitted)})
		}
		if cancelled != nil {
			break
		}
	}
	if cancelled == nil {
		s.pruneTombstones(d)
	}
	d.Errors = append([]IngestError{}, problems...)
	s.enrich(d)
	err = s.commit(d)
	if cancelled != nil {
		return res, cancelled
	}
	return res, err
}

func (s *Store) readRegistryShared() (Registry, error) {
	lock, err := AcquireLock(filepath.Join(s.vdir, "lock"), true)
	if err != nil {
		return Registry{}, err
	}
	defer lock.Release()
	return readRegistry(s.vdir)
}

// openWorkspace opens <root>/_support/workspace once, as an os.Root, and
// lists its real (non-link) folders. _support and workspace must be real
// directories; a link there is reported and nothing under it is read. Each
// step is checked with Lstat on the root opened before it, so no path is
// resolved twice. The caller closes the returned root.
func openWorkspace(repo string) (*os.Root, []string, []IngestError) {
	dir := repo
	cur, err := os.OpenRoot(repo)
	if err != nil {
		return nil, nil, []IngestError{{Path: repo, Reason: err.Error()}}
	}
	for _, part := range []string{"_support", "workspace"} {
		dir = filepath.Join(dir, part)
		info, err := cur.Lstat(part)
		if errors.Is(err, fs.ErrNotExist) {
			cur.Close()
			return nil, nil, nil
		}
		if err != nil {
			cur.Close()
			return nil, nil, []IngestError{{Path: dir, Reason: err.Error()}}
		}
		if !info.IsDir() {
			cur.Close()
			return nil, nil, []IngestError{{Path: dir, Reason: "must be a real directory, not a link"}}
		}
		next, err := cur.OpenRoot(part)
		cur.Close()
		if err != nil {
			return nil, nil, []IngestError{{Path: dir, Reason: err.Error()}}
		}
		cur = next
	}
	entries, err := fs.ReadDir(cur.FS(), ".")
	if err != nil {
		cur.Close()
		return nil, nil, []IngestError{{Path: dir, Reason: err.Error()}}
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() { // a DirEntry does not follow links: linked folders are skipped
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return cur, names, nil
}

// ingest handles one workspace folder. It opens the folder once from the
// workspace root and uses that descriptor to validate, hash and copy.
func (s *Store) ingest(ws *os.Root, name, wsDir string, now time.Time, d *indexData, res *SyncResult, problem func(string, error)) {
	folder := filepath.Join(wsDir, name)
	manifestPath := filepath.Join(folder, manifestFile)
	if info, err := ws.Lstat(name); err != nil || !info.IsDir() {
		return // removed or replaced by a link since it was listed
	}
	root, err := ws.OpenRoot(name)
	if err != nil {
		problem(folder, err)
		return
	}
	defer root.Close()
	info, err := root.Lstat(manifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		return
	}
	if err != nil {
		problem(manifestPath, err)
		return
	}
	if !info.Mode().IsRegular() {
		problem(manifestPath, errors.New("walkthrough.json must be a regular file, not a link"))
		return
	}
	if now.Sub(info.ModTime()) > s.ttl {
		return
	}
	m, err := validateRoot(root)
	if err != nil {
		problem(manifestPath, err)
		return
	}
	if s.afterValidate != nil {
		s.afterValidate(folder)
	}
	key := sessionKey(folder, m.SessionID)
	hash, err := hashFiles(root, m)
	if err != nil {
		problem(manifestPath, err)
		return
	}
	if ti, ok := d.tombstone(key); ok && d.Tombstones[ti].ContentHash == hash {
		return
	}
	i, exists := d.session(key)
	if exists && d.Sessions[i].ContentHash == hash {
		return
	}
	sess, err := s.install(key, folder, root, m, hash, now)
	if err != nil {
		problem(manifestPath, err)
		return
	}
	d.dropTombstone(key)
	if exists {
		d.Sessions[i] = sess
		res.Replaced++
		return
	}
	d.Sessions = append(d.Sessions, sess)
	res.Added++
}

// pruneTombstones drops the tombstones whose source walkthrough.json no
// longer exists, as long as its repository is reachable (an unmounted
// repository says nothing about its sources).
func (s *Store) pruneTombstones(d *indexData) {
	kept := d.Tombstones[:0:0]
	for _, t := range d.Tombstones {
		if _, err := os.Stat(repoRootOf(t.Source)); err == nil {
			if _, err := os.Lstat(filepath.Join(t.Source, manifestFile)); errors.Is(err, fs.ErrNotExist) {
				continue
			}
		}
		kept = append(kept, t)
	}
	d.Tombstones = kept
}
