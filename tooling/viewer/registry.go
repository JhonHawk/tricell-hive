package viewer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"tricell-hive/integrations/target"
)

const registryVersion = 1

// ErrNotGitRepository is returned by RegisterRepository for a directory that
// is neither inside a Git working tree nor inside a _support/workspace folder.
var ErrNotGitRepository = errors.New("not inside a Git repository or a _support/workspace folder")

// Registry is viewer/registry.json: the repositories the server scans.
type Registry struct {
	Version      int          `json:"version"`
	Repositories []Repository `json:"repositories"`
}

// Repository is one registered Git root.
type Repository struct {
	Root    string    `json:"root"`
	AddedAt time.Time `json:"added_at"`
}

// RegisterResult reports what RegisterRepository did.
type RegisterResult struct {
	// Root is the canonical Git root that is now registered.
	Root string
	// Added is false when the root was already registered; nothing is
	// written in that case.
	Added bool
}

func viewerDir(stateDir string) string { return filepath.Join(stateDir, "viewer") }

// FindGitRoot returns the canonical root of the Git working tree containing
// dir: the nearest ancestor holding a .git directory, or a .git file (a
// linked worktree). It runs no git process.
func FindGitRoot(dir string) (string, error) {
	canonical, err := canonicalDir(dir)
	if err != nil {
		return "", err
	}
	return gitRootOf(canonical, dir)
}

func canonicalDir(dir string) (string, error) {
	canonical, err := target.Canonical(dir)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return canonical, nil
}

func gitRootOf(canonical, original string) (string, error) {
	for p := canonical; ; p = filepath.Dir(p) {
		if _, err := os.Lstat(filepath.Join(p, ".git")); err == nil {
			return p, nil
		}
		if p == filepath.Dir(p) {
			return "", fmt.Errorf("%s: %w", original, ErrNotGitRepository)
		}
	}
}

// ResolveRoot returns the canonical root that dir belongs to. If dir (after
// resolving links) is inside <root>/_support/workspace/, at any depth, the
// root is <root>, whether or not it is a Git repository: a multi-repository
// workspace keeps its walkthroughs there while its repositories are nested
// folders. Otherwise it is the Git root of dir (FindGitRoot). Otherwise the
// error wraps ErrNotGitRepository.
func ResolveRoot(dir string) (string, error) {
	canonical, err := canonicalDir(dir)
	if err != nil {
		return "", err
	}
	for p := canonical; p != filepath.Dir(p); p = filepath.Dir(p) {
		support := filepath.Dir(p)
		if filepath.Base(p) == "workspace" && filepath.Base(support) == "_support" && support != filepath.Dir(support) {
			return filepath.Dir(support), nil
		}
	}
	return gitRootOf(canonical, dir)
}

// ReadRegistry reads <stateDir>/viewer/registry.json. A missing file is an
// empty registry; an unreadable or unknown-version file is an error.
func ReadRegistry(stateDir string) (Registry, error) {
	return readRegistry(viewerDir(stateDir))
}

func readRegistry(vdir string) (Registry, error) {
	data, err := os.ReadFile(filepath.Join(vdir, "registry.json"))
	if errors.Is(err, os.ErrNotExist) {
		return Registry{Version: registryVersion}, nil
	}
	if err != nil {
		return Registry{}, err
	}
	var r Registry
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&r); err != nil {
		return Registry{}, fmt.Errorf("registry.json: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return Registry{}, errors.New("registry.json: trailing data")
	}
	if r.Version != registryVersion {
		return Registry{}, fmt.Errorf("registry.json: unsupported version %d", r.Version)
	}
	return r, nil
}

// RegisterRepository adds the root of dir (see ResolveRoot: a workspace root or a Git root) to <stateDir>/viewer/registry.json.
// It needs no server: a blocking flock on viewer/lock serializes the
// read-modify-write against other processes, and an already registered root
// is only read, never rewritten. Outside a Git repository and a
// _support/workspace folder it returns an error wrapping ErrNotGitRepository.
func RegisterRepository(stateDir, dir string) (RegisterResult, error) {
	return registerRepository(viewerDir(stateDir), dir, time.Now())
}

func registerRepository(vdir, dir string, now time.Time) (RegisterResult, error) {
	root, err := ResolveRoot(dir)
	if err != nil {
		return RegisterResult{}, err
	}
	if err := os.MkdirAll(vdir, 0o700); err != nil {
		return RegisterResult{}, err
	}
	lock, err := AcquireLock(filepath.Join(vdir, "lock"), true)
	if err != nil {
		return RegisterResult{}, err
	}
	defer lock.Release()
	reg, err := readRegistry(vdir)
	if err != nil {
		return RegisterResult{}, err
	}
	for _, r := range reg.Repositories {
		if r.Root == root {
			return RegisterResult{Root: root}, nil
		}
	}
	reg.Version = registryVersion
	reg.Repositories = append(reg.Repositories, Repository{Root: root, AddedAt: now.UTC()})
	if err := writeJSONAtomic(filepath.Join(vdir, "registry.json"), reg); err != nil {
		return RegisterResult{}, err
	}
	return RegisterResult{Root: root, Added: true}, nil
}
