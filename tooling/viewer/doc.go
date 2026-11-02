// Package viewer is the core of `hive view`: the state behind a long-lived
// local server that shows walkthroughs (walkthrough.json manifests) from
// every registered repository, grouped by project, ticket and session.
//
// It has no HTTP server and no command line; tooling/viewer's server and
// tooling/cli's `hive view` call the API below. It imports only the standard
// library and tricell-hive/integrations/target (enforced by
// tooling/management/deps_test.go).
//
// # Layout
//
// Everything lives in <state-dir>/viewer/ (see Store):
//
//	lock             blocking flock between processes (registry read-modify-write)
//	server.lock      exclusive, non-blocking flock held by the one running server
//	                 (taken by the caller with AcquireLock)
//	registry.json    {"version":1,"repositories":[{"root","added_at"}]}
//	index.json       rebuildable cache: sessions, tombstones, ingest errors, repositories
//	sessions/<key>/  walkthrough.json + referenced files + meta.json
//
// <key> is SHA-256("<source folder>\x00<session id>"). meta.json is
// {source, content_hash, ingested_at, pinned, archived?, project_key?, project_name?};
// it, not index.json, is the source of truth for copies and their project identity.
// content_hash covers the manifest
// bytes and each referenced file's path and SHA-256, so rewriting only an
// image changes it and replaces the copy (keeping pinned).
//
// # API
//
// Validation (the Go twin of scripts/render_walkthrough.py, checked against
// the shared fixtures in tests/fixtures/walkthrough):
//
//	ValidateManifest(folder) (*Manifest, error)   // error is a *ManifestError
//
// Registry, usable without a server (`hive view register`):
//
//	ResolveRoot(dir) (string, error)   // <root> of _support/workspace/, else the Git root
//	FindGitRoot(dir) (string, error)
//	RegisterRepository(stateDir, dir) (RegisterResult, error)
//	ReadRegistry(stateDir) (Registry, error)
//
// Locking, for the caller's server.lock:
//
//	AcquireLock(path, wait) (*FileLock, error); ErrLocked; (*FileLock).Release
//
// Store, owned by the server process:
//
//	Open(Options) (*Store, error)        read-only load; deletes nothing
//	(*Store).Startup() error             clean temp dirs, rebuild index, persist
//	(*Store).Run(ctx, RunOptions) error  Startup, then Purge + Sync on intervals
//	(*Store).Sync(ctx) (SyncResult, error)
//	(*Store).Purge() ([]string, error)
//	(*Store).Delete(key) error           logical session + tombstones; sources untouched
//	(*Store).Pin(key, pinned) error       logical session retention
//	(*Store).Archive(key, archived) error logical session visibility
//	(*Store).SetTitle(key, title) error   local logical-session name
//	(*Store).View() View                 projects > tickets > sessions, never blocks
//	(*Store).Manifest(key) ([]byte, error)
//	(*Store).OpenFile(key, rel) (*os.File, error)
//
// Keys are validated (ErrInvalidKey) before any path is built; a missing
// session or file is ErrNotFound; a file path that breaks the manifest path
// rules or crosses a link is ErrInvalidPath.
//
// # Grouping
//
// Options.HiveProject and Options.GitRemote are injected by tooling/cli (they
// wrap parseHiveSection and the git runner there). Project key: the
// normalized remote (NormalizeRemote: host/owner/repo), else the repository
// root path. Project name: the Hive `Project:` value, else Engram's
// project_name (read here, at most 64 KiB, no links), else the last segment of
// the remote, else the folder name. A manifest without a ticket lands in the
// group with an empty ticket name. Worktrees of one remote share a project.
// A previously resolved remote identity is retained in copy metadata when its
// repository disappears or remote lookup fails, including across restarts.
//
// Logical sessions use the exact project/host/session-ID tuple. Preferences live
// in session-groups independently of the reconstructible physical-copy index.
// Legacy flags are imported once per group: archive requires all copies, keep
// requires any. Names do not change identity or source walkthrough titles.
//
// # Concurrency
//
// A sync.Mutex serializes the apply phase of Sync, Purge, session preferences and
// Startup in a process. Sync inspects and hashes sources before taking it and
// skips a source whose files keep their size, mtime and inode; View reads a
// published immutable snapshot and never waits for either.
// Only one server may own a state directory (the caller's server.lock).
// Between processes, only the registry is shared (register versus the
// server) and it is guarded by the blocking flock on viewer/lock.
//
// # Deletion and tombstones
//
// Explicit Delete commits every current group member's tombstone before removing
// its copy directory, then reconciles actual survivors. A pending logical-record
// cleanup is retried by Startup/Sync and gates new ingestion of that identity.
// Partial failures do not imply a complete deletion. Purge remains copy-based.
// Delete and Purge leave a tombstone {key, content_hash, source, deleted_at}
// in index.json; a scan skips a source whose key and content_hash match one,
// so deleted content stays out until it changes. Losing index.json loses the
// tombstones, which only means a live source is ingested once more.
package viewer
