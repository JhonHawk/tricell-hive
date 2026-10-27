package viewer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
	"time"
)

// env is a throwaway viewer: a state directory, a controllable clock and
// injected project readers.
type env struct {
	t           *testing.T
	hiveCalls   map[string]int // reader invocations per root
	remoteCalls map[string]int
	state       string
	mu          sync.Mutex
	now         time.Time
	hive        map[string]string
	remotes     map[string]string
	store       *Store
}

func newEnv(t *testing.T) *env {
	t.Helper()
	e := &env{
		t:           t,
		state:       canon(t, t.TempDir()),
		now:         time.Now().UTC().Truncate(time.Second),
		hive:        map[string]string{},
		remotes:     map[string]string{},
		hiveCalls:   map[string]int{},
		remoteCalls: map[string]int{},
	}
	e.store = e.open()
	return e
}

func (e *env) clock() time.Time {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.now
}

func (e *env) advance(d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.now = e.now.Add(d)
}

// open creates a new Store on the same state directory (a "restart").
func (e *env) open() *Store {
	e.t.Helper()
	s, err := Open(Options{
		StateDir: e.state,
		Now:      e.clock,
		HiveProject: func(root string) (string, bool) {
			e.mu.Lock()
			defer e.mu.Unlock()
			e.hiveCalls[root]++
			n, ok := e.hive[root]
			return n, ok
		},
		GitRemote: func(root string) (string, error) {
			e.mu.Lock()
			defer e.mu.Unlock()
			e.remoteCalls[root]++
			if r, ok := e.remotes[root]; ok {
				return r, nil
			}
			return "", errors.New("no remote")
		},
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return s
}

func (e *env) start() {
	e.t.Helper()
	if err := e.store.Startup(); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) register(repo string) {
	e.t.Helper()
	if _, err := RegisterRepository(e.state, repo); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) sync() SyncResult {
	e.t.Helper()
	res, err := e.store.Sync(context.Background())
	if err != nil {
		e.t.Fatal(err)
	}
	return res
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, r)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// place copies valid/final into <repo>/_support/workspace/<name>, gives it
// session id "sess-<name>" and applies mutate to the manifest. It returns
// the source folder.
func (e *env) place(repo, name string, mutate func(m map[string]any)) string {
	e.t.Helper()
	folder := filepath.Join(repo, "_support", "workspace", name)
	copyTree(e.t, filepath.Join(fixturesRoot, "valid", "final"), folder)
	m := readManifestMap(e.t, folder)
	m["session"] = map[string]any{"id": "sess-" + name, "host": "claude-code"}
	if mutate != nil {
		mutate(m)
	}
	writeManifestMap(e.t, folder, m)
	e.touch(folder)
	return folder
}

// touch sets the manifest mtime to the fake clock.
func (e *env) touch(folder string) {
	e.t.Helper()
	now := e.clock()
	if err := os.Chtimes(filepath.Join(folder, "walkthrough.json"), now, now); err != nil {
		e.t.Fatal(err)
	}
}

func (e *env) sessions() []Session {
	var out []Session
	for _, p := range e.store.View().Projects {
		for _, tk := range p.Tickets {
			out = append(out, tk.Sessions...)
		}
	}
	return out
}

func (e *env) onlySession() Session {
	e.t.Helper()
	s := e.sessions()
	if len(s) != 1 {
		e.t.Fatalf("want exactly one session, got %d", len(s))
	}
	return s[0]
}

func (e *env) sessionDir(key string) string {
	return filepath.Join(e.state, "viewer", "sessions", key)
}

func sha(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

type metaFile struct {
	Source      string    `json:"source"`
	ContentHash string    `json:"content_hash"`
	IngestedAt  time.Time `json:"ingested_at"`
	Pinned      bool      `json:"pinned"`
}

func (e *env) meta(key string) metaFile {
	e.t.Helper()
	var m metaFile
	if err := json.Unmarshal(readFile(e.t, filepath.Join(e.sessionDir(key), "meta.json")), &m); err != nil {
		e.t.Fatal(err)
	}
	return m
}

func TestStoreKeyIsTheHashOfSourceFolderAndSessionID(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	s := e.onlySession()
	if want := sha(folder + "\x00" + "sess-run1"); s.Key != want {
		t.Fatalf("key = %s, want %s", s.Key, want)
	}
	if s.Source != folder || s.SessionID != "sess-run1" || s.Host != "claude-code" {
		t.Fatalf("session fields: %+v", s)
	}
}

func TestStoreCopiesTheManifestAndEveryReferencedFileWithMetadata(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	res := e.sync()
	if res.Added != 1 || res.Errors != 0 {
		t.Fatalf("result = %+v", res)
	}
	s := e.onlySession()
	dir := e.sessionDir(s.Key)
	for _, rel := range []string{"walkthrough.json", "images/nav-1.png", "images/nav-2.png", "images/nav-3.png", "images/settings-1.png"} {
		got, want := readFile(t, filepath.Join(dir, rel)), readFile(t, filepath.Join(folder, rel))
		if !bytes.Equal(got, want) {
			t.Fatalf("%s differs from its source", rel)
		}
	}
	meta := e.meta(s.Key)
	if meta.Source != folder || meta.Pinned || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(meta.ContentHash) {
		t.Fatalf("meta = %+v", meta)
	}
	if !meta.IngestedAt.Equal(e.clock()) {
		t.Fatalf("ingested_at = %v, want the injected clock %v", meta.IngestedAt, e.clock())
	}
	if s.ContentHash != meta.ContentHash {
		t.Fatal("view hash differs from meta.json")
	}
	if info, _ := os.Stat(filepath.Join(dir, "meta.json")); info.Mode().Perm() != 0o600 {
		t.Fatalf("meta.json mode = %v", info.Mode().Perm())
	}
}

func TestStoreContentHashChangesWhenOnlyAReferencedFileChanges(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	first := e.onlySession()
	manifestBefore := readFile(t, filepath.Join(folder, "walkthrough.json"))

	if err := os.WriteFile(filepath.Join(folder, "images", "nav-1.png"), []byte("new image bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	e.advance(time.Minute)
	res := e.sync()
	if res.Replaced != 1 || res.Added != 0 {
		t.Fatalf("result = %+v, want one replacement", res)
	}
	second := e.onlySession()
	if second.Key != first.Key {
		t.Fatal("a content change must keep the key")
	}
	if second.ContentHash == first.ContentHash {
		t.Fatal("content_hash did not change after rewriting only images/nav-1.png")
	}
	if !bytes.Equal(manifestBefore, readFile(t, filepath.Join(folder, "walkthrough.json"))) {
		t.Fatal("test bug: the manifest changed")
	}
	if got := readFile(t, filepath.Join(e.sessionDir(second.Key), "images", "nav-1.png")); string(got) != "new image bytes" {
		t.Fatalf("copy not replaced: %q", got)
	}
	if !second.IngestedAt.After(first.IngestedAt) {
		t.Fatal("replacement must refresh ingested_at")
	}
	entries, _ := os.ReadDir(filepath.Join(e.state, "viewer", "sessions"))
	if len(entries) != 1 {
		t.Fatalf("leftover entries after a replacement: %v", entries)
	}
}

func TestStoreReplaceKeepsPinned(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(folder, "images", "nav-2.png"), []byte("changed"), 0o644)
	e.advance(time.Minute)
	e.sync()
	if e.meta(key).Pinned || !e.onlySession().Pinned {
		t.Fatal("pinned lost on replacement")
	}
}

func TestStorePinSurvivesARestartAndCanBeUndone(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}
	e.store = e.open()
	e.start()
	if !e.onlySession().Pinned {
		t.Fatal("pin lost across restart")
	}
	if err := e.store.Pin(key, false); err != nil {
		t.Fatal(err)
	}
	if e.meta(key).Pinned || e.onlySession().Pinned {
		t.Fatal("pin not removed")
	}
}

func TestStoreDeleteRemovesOnlyTheCopy(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	if err := e.store.Delete(key); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(e.sessionDir(key)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("copy still present: %v", err)
	}
	for _, rel := range []string{"walkthrough.json", "images/nav-1.png"} {
		if _, err := os.Stat(filepath.Join(folder, rel)); err != nil {
			t.Fatalf("repository file %s touched: %v", rel, err)
		}
	}
	if len(e.sessions()) != 0 {
		t.Fatal("deleted session still listed")
	}
	if err := e.store.Delete(key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete err = %v, want ErrNotFound", err)
	}
}

func TestStoreRejectsMalformedKeysBeforeTouchingTheDisk(t *testing.T) {
	e := newEnv(t)
	e.start()
	sentinel := filepath.Join(e.state, "viewer", "registry.json")
	os.WriteFile(sentinel, []byte(`{"version":1,"repositories":[]}`), 0o600)
	for _, key := range []string{"", "..", "../registry.json", "abc", sha("x")[:63], sha("x") + "0", "G" + sha("x")[1:], sha("x")[:62] + "/.", "ABCDEF" + sha("x")[6:]} {
		if err := e.store.Delete(key); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("Delete(%q) = %v, want ErrInvalidKey", key, err)
		}
		if err := e.store.Pin(key, true); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("Pin(%q) = %v, want ErrInvalidKey", key, err)
		}
		if _, err := e.store.Manifest(key); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("Manifest(%q) = %v, want ErrInvalidKey", key, err)
		}
		if _, err := e.store.OpenFile(key, "images/a.png"); !errors.Is(err, ErrInvalidKey) {
			t.Fatalf("OpenFile(%q) = %v, want ErrInvalidKey", key, err)
		}
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatal("a malformed key reached outside sessions/")
	}
	if err := e.store.Pin(sha("unknown"), true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Pin(unknown) = %v", err)
	}
}

func TestStoreServesTheManifestAndOnlyValidatedFiles(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key

	raw, err := e.store.Manifest(key)
	if err != nil || !bytes.Equal(raw, readFile(t, filepath.Join(folder, "walkthrough.json"))) {
		t.Fatalf("Manifest = %d bytes, %v", len(raw), err)
	}
	f, err := e.store.OpenFile(key, "images/nav-1.png")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(f)
	f.Close()
	if !bytes.Equal(data, readFile(t, filepath.Join(folder, "images/nav-1.png"))) {
		t.Fatal("served bytes differ")
	}
	for _, rel := range []string{"", "meta.json", "walkthrough.json", "../meta.json", "images/../meta.json",
		"/etc/passwd", "images//nav-1.png", "images/nav-1.gif", "images/missing.png", "images", "images/nav 1.png"} {
		if f, err := e.store.OpenFile(key, rel); err == nil {
			f.Close()
			t.Fatalf("OpenFile(%q) succeeded", rel)
		} else if !errors.Is(err, ErrInvalidPath) && !errors.Is(err, ErrNotFound) {
			t.Fatalf("OpenFile(%q) = %v, want ErrInvalidPath or ErrNotFound", rel, err)
		}
	}
	// A link planted inside the copy must not be followed.
	outside := filepath.Join(t.TempDir(), "secret.png")
	os.WriteFile(outside, []byte("secret"), 0o600)
	link := filepath.Join(e.sessionDir(key), "images", "evil.png")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if f, err := e.store.OpenFile(key, "images/evil.png"); err == nil {
		f.Close()
		t.Fatal("symlink followed")
	}
}

func TestStoreStartupCleansTemporaryDirectoriesAndRestoresAnInterruptedReplace(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	sessions := filepath.Join(e.state, "viewer", "sessions")

	// A failed copy and a stray atomic-write file.
	os.MkdirAll(filepath.Join(sessions, ".tmp-123", "images"), 0o700)
	os.WriteFile(filepath.Join(sessions, ".tmp-123", "images", "x.png"), []byte("x"), 0o600)
	os.WriteFile(filepath.Join(e.state, "viewer", ".write-999"), []byte("x"), 0o600)
	// An .old copy while the live one exists: removed.
	os.MkdirAll(filepath.Join(sessions, ".old-"+key), 0o700)
	// A crash between the two renames of a replacement: live copy missing.
	other := sha("other")
	os.MkdirAll(filepath.Join(sessions, ".old-"+other), 0o700)
	os.WriteFile(filepath.Join(sessions, ".old-"+other, "meta.json"), []byte("{}"), 0o600)

	e.store = e.open()
	e.start()
	for _, gone := range []string{filepath.Join(sessions, ".tmp-123"), filepath.Join(e.state, "viewer", ".write-999"), filepath.Join(sessions, ".old-"+key), filepath.Join(sessions, ".old-"+other)} {
		if _, err := os.Lstat(gone); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s survived startup (%v)", gone, err)
		}
	}
	if _, err := os.Stat(filepath.Join(sessions, other, "meta.json")); err != nil {
		t.Fatalf("interrupted replacement not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sessions, key, "meta.json")); err != nil {
		t.Fatal("live copy damaged")
	}
}

func TestStoreConcurrentOperationsAreSerialized(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.place(repo, "run2", nil)
	e.register(repo)
	e.start()
	e.sync()
	keys := []string{e.sessions()[0].Key, e.sessions()[1].Key}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(4)
		go func() { defer wg.Done(); e.store.Sync(context.Background()) }()
		go func() { defer wg.Done(); e.store.Pin(keys[0], true); e.store.Pin(keys[0], false) }()
		go func() { defer wg.Done(); _ = e.store.View(); e.store.Purge() }()
		go func(i int) {
			defer wg.Done()
			os.WriteFile(filepath.Join(folder, "images", "nav-3.png"), []byte{byte(i)}, 0o644)
		}(i)
	}
	wg.Wait()
	e.sync()
	if got := len(e.sessions()); got != 2 {
		t.Fatalf("sessions = %d, want 2", got)
	}
	// The view is served without waiting for a running ingestion.
	e.store.mu.Lock()
	done := make(chan struct{})
	go func() { e.store.View(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("View blocked behind the ingestion mutex")
	}
	e.store.mu.Unlock()
}

func TestStoreOpenCreatesPrivateDirectories(t *testing.T) {
	e := newEnv(t)
	for _, p := range []string{"viewer", "viewer/sessions"} {
		info, err := os.Stat(filepath.Join(e.state, p))
		if err != nil || info.Mode().Perm() != 0o700 {
			t.Fatalf("%s: %v %v", p, info, err)
		}
	}
	if _, err := Open(Options{StateDir: ""}); err == nil {
		t.Fatal("empty state dir accepted")
	}
	if _, err := Open(Options{StateDir: "relative/dir"}); err == nil {
		t.Fatal("relative state dir accepted")
	}
}

// ---- Purge ----

func TestPurgeRemovesExpiredCopiesAndTheyDoNotReturn(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key

	e.advance(15 * 24 * time.Hour)
	e.touch(folder) // the source is "fresh" for the advanced clock: only the tombstone can keep it out
	purged, err := e.store.Purge()
	if err != nil || len(purged) != 1 || purged[0] != key {
		t.Fatalf("Purge = %v, %v", purged, err)
	}
	if _, err := os.Stat(e.sessionDir(key)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("expired copy not removed")
	}
	if len(e.sessions()) != 0 {
		t.Fatal("expired session still listed")
	}
	e.advance(time.Minute)
	e.touch(folder)
	if res := e.sync(); res.Added != 0 || len(e.sessions()) != 0 {
		t.Fatalf("purged copy came back: %+v", res)
	}
	if _, err := os.Stat(e.sessionDir(key)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("purged copy re-created")
	}
	// Changed content is a new fact: it is ingested again.
	os.WriteFile(filepath.Join(folder, "images", "nav-1.png"), []byte("changed"), 0o644)
	if res := e.sync(); res.Added != 1 || len(e.sessions()) != 1 {
		t.Fatalf("changed source not re-ingested: %+v", res)
	}
}

func TestPurgeKeepsPinnedAcrossLostIndexAndChangedSource(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	key := e.onlySession().Key
	if err := e.store.Pin(key, true); err != nil {
		t.Fatal(err)
	}

	e.advance(15 * 24 * time.Hour)
	if err := os.Remove(filepath.Join(e.state, "viewer", "index.json")); err != nil {
		t.Fatal(err)
	}
	e.store = e.open()
	e.start() // rebuilds the index from sessions/*/meta.json
	if purged, err := e.store.Purge(); err != nil || len(purged) != 0 {
		t.Fatalf("pinned copy purged: %v %v", purged, err)
	}
	if !e.onlySession().Pinned {
		t.Fatal("pin lost with the index")
	}

	os.WriteFile(filepath.Join(folder, "images", "nav-1.png"), []byte("changed after 15 days"), 0o644)
	e.touch(folder)
	if res := e.sync(); res.Replaced != 1 {
		t.Fatalf("changed source not replaced: %+v", res)
	}
	if purged, _ := e.store.Purge(); len(purged) != 0 {
		t.Fatalf("pinned copy purged after replacement: %v", purged)
	}
	if s := e.onlySession(); !s.Pinned || s.Key != key {
		t.Fatalf("pinned session = %+v", s)
	}
}

func TestPurgeUsesIngestionTimeNotTheManifestCreatedAt(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", func(m map[string]any) { m["created_at"] = "2001-01-01T00:00:00Z" })
	e.register(repo)
	e.start()
	e.sync()
	if purged, _ := e.store.Purge(); len(purged) != 0 {
		t.Fatalf("a copy was purged because of an agent-written created_at: %v", purged)
	}
}

func TestPurgeBoundaryIsStrictlyOlderThanTheTTL(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.advance(14 * 24 * time.Hour)
	if purged, _ := e.store.Purge(); len(purged) != 0 {
		t.Fatalf("a copy exactly TTL old was purged: %v", purged)
	}
	e.advance(time.Second)
	if purged, _ := e.store.Purge(); len(purged) != 1 {
		t.Fatalf("a copy older than the TTL survived: %v", purged)
	}
}

func TestPurgeHonoursAConfiguredTTLAndDefaultsToFourteenDays(t *testing.T) {
	e := newEnv(t)
	if got := e.store.View().TTLSeconds; got != 14*24*3600 {
		t.Fatalf("default ttl = %d", got)
	}
	s, err := Open(Options{StateDir: t.TempDir(), TTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if got := s.View().TTLSeconds; got != 3600 {
		t.Fatalf("configured ttl = %d", got)
	}
}

func TestPurgeRunsAtStartAndOnTheInterval(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.advance(15 * 24 * time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- e.store.Run(ctx, RunOptions{PollInterval: 10 * time.Millisecond, PurgeInterval: 10 * time.Millisecond})
	}()
	waitFor(t, "the expired copy to be purged by Run", func() bool { return len(e.sessions()) == 0 })

	// A copy that expires later is purged by the periodic pass.
	e.place(repo, "run2", nil)
	waitFor(t, "run2 to be ingested", func() bool { return len(e.sessions()) == 1 })
	e.advance(15 * 24 * time.Hour)
	waitFor(t, "run2 to be purged on the interval", func() bool { return len(e.sessions()) == 0 })
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop on cancel")
	}
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// Gap from the T3 verification: only walkthrough.json changes (same
// session.id, same images) and the copy must still be replaced.
func TestStoreContentHashChangesWhenOnlyTheManifestChanges(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	folder := e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	first := e.onlySession()
	imageBefore := readFile(t, filepath.Join(folder, "images", "nav-1.png"))

	m := readManifestMap(t, folder)
	m["title"] = "a retitled walkthrough"
	writeManifestMap(t, folder, m)
	e.advance(time.Minute)
	e.touch(folder)
	res := e.sync()
	if res.Replaced != 1 || res.Added != 0 {
		t.Fatalf("result = %+v, want one replacement", res)
	}
	second := e.onlySession()
	if second.Key != first.Key || second.SessionID != first.SessionID {
		t.Fatal("a manifest-only change must keep the key and session id")
	}
	if second.ContentHash == first.ContentHash {
		t.Fatal("content_hash did not change after rewriting only walkthrough.json")
	}
	if second.Title != "a retitled walkthrough" {
		t.Fatalf("copy not replaced, title = %q", second.Title)
	}
	if !bytes.Equal(imageBefore, readFile(t, filepath.Join(e.sessionDir(second.Key), "images", "nav-1.png"))) {
		t.Fatal("images must be untouched by the replacement")
	}
	if !second.IngestedAt.After(first.IngestedAt) {
		t.Fatal("replacement must refresh ingested_at")
	}
}

// Gap from the T3 verification: with a purge interval that cannot fire during
// the wait, only the purge Run performs at startup can remove the copy.
func TestPurgeRunsAtStartupEvenWhenTheIntervalIsLong(t *testing.T) {
	e := newEnv(t)
	repo := newRepo(t)
	e.place(repo, "run1", nil)
	e.register(repo)
	e.start()
	e.sync()
	e.advance(15 * 24 * time.Hour)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- e.store.Run(ctx, RunOptions{PollInterval: time.Hour, PurgeInterval: time.Hour})
	}()
	waitFor(t, "the startup purge to remove the expired copy", func() bool { return len(e.sessions()) == 0 })
	if len(e.store.View().Tombstones) != 1 {
		t.Fatalf("tombstones = %d, want 1", len(e.store.View().Tombstones))
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop on cancel")
	}
}
