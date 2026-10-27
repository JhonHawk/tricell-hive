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
	"reflect"
	"sort"
	"syscall"
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
	if info.Size() > limit {
		return "", fmt.Errorf("%s exceeds %d MiB", rel, limit>>20)
	}
	if err := validateSingleLink(info); err != nil {
		return "", fmt.Errorf("%s: %w", rel, err)
	}
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
// previous copy. Logical preferences are applied separately. The copy is
// assembled in a temporary directory and swapped in with renames, so a reader sees the
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
	mt := meta{Source: folder, ContentHash: hash, IngestedAt: now.UTC()}
	if old, err := readMeta(final); err == nil {
		mt.ProjectKey, mt.ProjectName = old.ProjectKey, old.ProjectName
	}
	if err := writeMeta(tmp, mt); err != nil {
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
	sm.Session.Status = storedSessionStatus(m.SessionStatus)
	return sessionFrom(key, mt, sm), nil
}

// statKey is the cheap identity of a file between passes. ctime and dev are
// part of it because an unprivileged user can set mtime, but not ctime: an
// in-place rewrite that restores size and mtime still changes ctime.
type statKey struct {
	size  int64
	mtime int64
	ino   uint64
	dev   uint64
	ctime int64
}

func statOf(info os.FileInfo) statKey {
	k := statKey{size: info.Size(), mtime: info.ModTime().UnixNano()}
	if st, ok := info.Sys().(*syscall.Stat_t); ok { // unix only, like the flock
		k.ino, k.dev, k.ctime = uint64(st.Ino), uint64(st.Dev), ctimeOf(st)
	}
	return k
}

// ctimeOf reads the status-change time from a Stat_t. The field is named
// Ctimespec on darwin and Ctim on linux, so it is found by reflection rather
// than by a build-tagged file.
func ctimeOf(st *syscall.Stat_t) int64 {
	v := reflect.ValueOf(st).Elem()
	for _, name := range []string{"Ctimespec", "Ctim"} {
		if f := v.FieldByName(name); f.IsValid() {
			return f.FieldByName("Sec").Int()*1e9 + f.FieldByName("Nsec").Int()
		}
	}
	return 0
}

// fingerprint is what the last hash of a source looked like.
type fingerprint struct {
	manifest statKey
	files    map[string]statKey
	key      string
	hash     string
}

// unchanged reports whether folder's manifest and every file the last hash
// covered still have the same (size, mtime, inode), and the current index
// already reflects that hash (a copy with it, or a tombstone for it), so the
// source needs no validation or hashing. It is a cheap heuristic: content
// rewritten without changing size, mtime and inode goes unnoticed until one
// of them changes.
func (s *Store) unchanged(folder string, root *os.Root, manifest statKey, snap *indexData) bool {
	s.fpMu.Lock()
	fp := s.fps[folder]
	s.fpMu.Unlock()
	if fp == nil || fp.manifest != manifest {
		return false
	}
	for rel, want := range fp.files {
		info, err := root.Lstat(filepath.FromSlash(rel))
		if err != nil || !info.Mode().IsRegular() || statOf(info) != want {
			return false
		}
	}
	if i, ok := snap.session(fp.key); ok && snap.Sessions[i].ContentHash == fp.hash {
		return true
	}
	if i, ok := snap.tombstone(fp.key); ok && snap.Tombstones[i].ContentHash == fp.hash {
		return true
	}
	return false
}

func (s *Store) remember(folder string, fp *fingerprint) {
	s.fpMu.Lock()
	if s.fps == nil {
		s.fps = map[string]*fingerprint{}
	}
	s.fps[folder] = fp
	s.fpMu.Unlock()
}

// pruneFingerprints drops the fingerprints of folders this pass did not list:
// sources that disappeared, and repositories that are absent or no longer
// registered.
func (s *Store) pruneFingerprints(seen map[string]bool) {
	s.fpMu.Lock()
	for folder := range s.fps {
		if !seen[folder] {
			delete(s.fps, folder)
		}
	}
	s.fpMu.Unlock()
}

func (s *Store) forget(folder string) {
	s.fpMu.Lock()
	delete(s.fps, folder)
	s.fpMu.Unlock()
}

// candidate is a source found by the unlocked inspection phase of Sync. root
// stays open (on the folder as it was validated) until the apply phase has
// copied from it.
type candidate struct {
	repo        int
	folder      string
	problemPath string
	err         error
	root        *os.Root
	m           *Manifest
	key, hash   string
}

// Defaults for how many inspected candidates, and how many bytes of their raw
// manifests, one batch of a scan retains before it is applied.
const (
	defaultBatchSize  = 32
	defaultBatchBytes = 64 << 20
)

// problemLog collects ingest errors with the per-repository bounds of H4.
type problemLog struct {
	items   []IngestError
	count   []int
	omitted []int
	errors  int
}

func (p *problemLog) add(repo int, path string, err error) {
	p.errors++
	if p.count[repo] >= maxIngestErrorsPerRepository {
		p.omitted[repo]++
		return
	}
	p.count[repo]++
	p.items = append(p.items, IngestError{Path: cleanText(path, maxErrorPath), Reason: cleanText(err.Error(), maxErrorText)})
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
//
// A pass has two phases. Inspection (validate, fingerprint, hash) runs
// without the store mutex, so Delete, Pin and Purge never wait for it; a
// source whose files keep their (size, mtime, inode) and whose copy is
// current is skipped without being read. The apply phase then takes the
// mutex, re-checks tombstones and the current index, and copies.
func (s *Store) Sync(ctx context.Context) (SyncResult, error) {
	var res SyncResult
	reg, err := s.readRegistryShared()
	if err != nil {
		return res, err
	}
	now := s.now()
	s.noteRepositories(reg)
	snap := s.current()
	// Discover paths without reading manifests, then classify before cache or TTL.
	paths := sessionSourcePaths(snap)
	namesByRoot := make(map[string][]string)
	for _, r := range reg.Repositories {
		ws, names, _ := openWorkspace(r.Root)
		if ws != nil {
			ws.Close()
		}
		namesByRoot[r.Root] = names
		for _, name := range names {
			paths = append(paths, filepath.Join(r.Root, "_support", "workspace", name, manifestFile))
		}
	}
	classification := s.classifySources(ctx, paths)
	repos := repositoryStatuses(reg)
	log := &problemLog{count: make([]int, len(reg.Repositories)), omitted: make([]int, len(reg.Repositories))}

	batchSize, batchBytes := s.batchSize, s.batchBytes
	if batchSize <= 0 {
		batchSize = defaultBatchSize
	}
	if batchBytes <= 0 {
		batchBytes = defaultBatchBytes
	}
	var cands []*candidate
	var retained int64
	closeBatch := func() {
		for _, c := range cands {
			if c.root != nil {
				c.root.Close()
			}
		}
		cands, retained = nil, 0
	}
	defer closeBatch()
	// flush applies the batch under the mutex, then closes its roots, so a
	// pass never keeps more than one bounded batch of open folders and raw
	// manifests. Each batch is committed on its own.
	flush := func() error {
		if len(cands) == 0 {
			return nil
		}
		if s.onBatch != nil {
			s.onBatch(len(cands))
		}
		s.mu.Lock()
		defer s.mu.Unlock()
		d := s.current().clone()
		for _, c := range cands {
			if c.err != nil {
				log.add(c.repo, c.problemPath, c.err)
				continue
			}
			if err := s.apply(c, now, d, &res); err != nil {
				log.add(c.repo, filepath.Join(c.folder, manifestFile), err)
			}
		}
		s.classifySessions(d, classification)
		closeBatch()
		s.enrich(d)
		return s.commit(d)
	}
	seen := map[string]bool{}
	cancelled := ctx.Err()
	var flushErr error
	for i, r := range reg.Repositories {
		if cancelled != nil {
			break
		}
		if repos[i].Absent {
			continue
		}
		ws, _, errs := openWorkspace(r.Root)
		names := namesByRoot[r.Root]
		for _, e := range errs {
			log.add(i, e.Path, errors.New(e.Reason))
		}
		if ws != nil {
			wsDir := filepath.Join(r.Root, "_support", "workspace")
			for _, name := range names {
				if cancelled = ctx.Err(); cancelled != nil {
					break
				}
				folder := filepath.Join(wsDir, name)
				seen[folder] = true
				if classification[filepath.Join(folder, manifestFile)].Status == "tracked" {
					continue
				}
				if c := s.inspect(ws, name, wsDir, now, snap); c != nil {
					c.repo = i
					cands = append(cands, c)
					if c.m != nil {
						retained += int64(len(c.m.Raw))
					}
					if len(cands) >= batchSize || retained >= batchBytes {
						if err := flush(); err != nil && flushErr == nil {
							flushErr = err
						}
					}
				}
			}
			ws.Close()
		}
		if cancelled != nil {
			break
		}
	}
	if err := flush(); err != nil && flushErr == nil {
		flushErr = err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	d := s.current().clone()
	s.classifySessions(d, classification)
	d.Repositories = repos
	for i, r := range reg.Repositories {
		if log.omitted[i] > 0 {
			log.items = append(log.items, IngestError{Path: r.Root, Reason: fmt.Sprintf("%d more ingest errors omitted", log.omitted[i])})
		}
	}
	res.Errors = log.errors
	if cancelled == nil {
		s.pruneTombstones(d)
		s.pruneFingerprints(seen)
	}
	d.Errors = append([]IngestError{}, log.items...)
	s.enrich(d)
	err = s.commit(d)
	if cancelled != nil {
		return res, cancelled
	}
	if flushErr != nil {
		return res, flushErr
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

// inspect is the unlocked phase for one workspace folder. It returns nil
// when there is nothing to do, a candidate with err set for a rejected
// source, or a candidate ready to be applied. It opens the folder once from
// the workspace root and uses that descriptor to validate and hash, and keeps
// it for the copy.
func (s *Store) inspect(ws *os.Root, name, wsDir string, now time.Time, snap *indexData) *candidate {
	folder := filepath.Join(wsDir, name)
	manifestPath := filepath.Join(folder, manifestFile)
	if info, err := ws.Lstat(name); err != nil || !info.IsDir() {
		return nil // removed or replaced by a link since it was listed
	}
	root, err := ws.OpenRoot(name)
	if err != nil {
		return &candidate{folder: folder, problemPath: folder, err: err}
	}
	fail := func(path string, err error) *candidate {
		root.Close()
		s.forget(folder)
		return &candidate{folder: folder, problemPath: path, err: err}
	}
	info, err := root.Lstat(manifestFile)
	if errors.Is(err, fs.ErrNotExist) {
		root.Close()
		return nil
	}
	if err != nil {
		return fail(manifestPath, err)
	}
	if !info.Mode().IsRegular() {
		return fail(manifestPath, errors.New("walkthrough.json must be a regular file, not a link"))
	}
	if now.Sub(info.ModTime()) > s.ttl {
		root.Close()
		return nil
	}
	manifest := statOf(info)
	if s.unchanged(folder, root, manifest, snap) {
		root.Close()
		return nil
	}
	m, err := validateRoot(root)
	if err != nil {
		return fail(manifestPath, err)
	}
	if s.afterValidate != nil {
		s.afterValidate(folder)
	}
	fp := &fingerprint{manifest: manifest, files: map[string]statKey{}}
	for _, rel := range m.Files {
		if fi, err := root.Lstat(filepath.FromSlash(rel)); err == nil && fi.Mode().IsRegular() {
			fp.files[rel] = statOf(fi)
		} else {
			fp = nil
			break
		}
	}
	key := sessionKey(folder, m.SessionID)
	if s.onHash != nil {
		s.onHash(folder)
	}
	hash, err := hashFiles(root, m)
	if err != nil {
		return fail(manifestPath, err)
	}
	if fp != nil {
		fp.key, fp.hash = key, hash
		s.remember(folder, fp)
	}
	return &candidate{folder: folder, root: root, m: m, key: key, hash: hash}
}

// apply is the locked phase for one inspected source: it re-checks the
// tombstones and the current copy, then installs.
func (s *Store) apply(c *candidate, now time.Time, d *indexData, res *SyncResult) error {
	if ti, ok := d.tombstone(c.key); ok && d.Tombstones[ti].ContentHash == c.hash {
		return nil
	}
	i, exists := d.session(c.key)
	if exists && d.Sessions[i].ContentHash == c.hash {
		return nil
	}
	// Resolve the same project identity used by publication before touching disk.
	probe := Session{Key: c.key, Source: c.folder, SessionID: c.m.SessionID, Host: c.m.SessionHost}
	if exists {
		probe = d.Sessions[i]
		probe.SessionID, probe.Host = c.m.SessionID, c.m.SessionHost
	}
	identities := newIndexData()
	identities.Sessions = []Session{probe}
	s.enrich(identities)
	group, groupErr := s.readGroup(identityOf(identities.Sessions[0]))
	if groupErr != nil && !errors.Is(groupErr, fs.ErrNotExist) {
		return fmt.Errorf("session preferences: %w", groupErr)
	}
	sess, err := s.install(c.key, c.folder, c.root, c.m, c.hash, now)
	if err != nil {
		s.forget(c.folder)
		return err
	}
	if groupErr == nil {
		sess.Pinned, sess.Archived, sess.SessionTitle = group.Pinned, group.Archived, group.TitleOverride
	}
	d.dropTombstone(c.key)
	if exists {
		previous := d.Sessions[i]
		sess.Repository, sess.ProjectKey, sess.ProjectName = previous.Repository, previous.ProjectKey, previous.ProjectName
		d.Sessions[i] = sess
		res.Replaced++
		return nil
	}
	d.Sessions = append(d.Sessions, sess)
	res.Added++
	return nil
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
