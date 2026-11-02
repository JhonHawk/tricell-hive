package viewer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

// Historical age used to prove that legacy expiry thresholds no longer apply.
const defaultTTL = 14 * 24 * time.Hour

// ---- Manual retention ----

func TestManualRetentionIgnoresLegacyConfiguredTTL(t *testing.T) {
	e := newEnv(t)
	if e.store.View().TTLSeconds != 0 {
		t.Fatal("default expiry remains")
	}
	s, err := Open(Options{StateDir: t.TempDir(), TTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	if s.TTL() != 0 || s.View().TTLSeconds != 0 {
		t.Fatal("legacy TTL reactivated expiry")
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

// Gap from the T3 verification: with a purge interval that cannot fire during
// the wait, only the purge Run performs at startup can remove the copy.
