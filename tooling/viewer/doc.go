// Package viewer is the core of `hive view`: the state behind a long-lived
// local server that shows the work of agent sessions (the timeline API, v3) from
// every registered repository.
//
// The HTTP server (Server) and the state (Store) live here; the command line
// lives in tooling/cli. The package imports only the standard library and
// tricell-hive/integrations/target (enforced by tooling/management/deps_test.go).
//
// # Layout
//
// Everything lives in <state-dir>/viewer/:
//
//	storage.json        {"storage_version":3}, the marker of a v3 namespace
//	lock                blocking flock between processes (registry read-modify-write)
//	server.lock         exclusive, non-blocking flock held by the one running server
//	                    (taken by the caller with AcquireLock)
//	registry.json       {"version":1,"repositories":[{"root","added_at"}]}
//	timeline-api-v3/    one <id>.json record per work, plus <id>/documents and
//	                    <id>/artifacts for its stored evidence
//
// Any other name, a missing marker or another storage_version is rejected by the
// strict inventory (ValidateState) with invalid_storage, before anything is
// written. No older state is imported, read or migrated.
//
// # API
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
//	Open(Options) (*Store, error)        validates the state and loads the records
//	(*Store).Startup() error             validate, then clean owned temporaries
//	(*Store).Run(ctx, RunOptions) error  Startup, then wait for cancellation
//	(*Store).Timeline*(...)              the timeline operations of the API
//
// Keys are validated (ErrInvalidKey) before any path is built; a missing work
// or file is ErrNotFound; a file path that is not a safe relative path or
// crosses a link is ErrInvalidPath.
//
// # Grouping
//
// Options.HiveProject and Options.GitRemote are injected by tooling/cli (they
// wrap parseHiveSection and the git runner there). Project key: the
// normalized remote (NormalizeRemote: host/owner/repo), else the repository
// root path. Project name: the Hive `Project:` value, else Engram's
// project_name (read here, at most 64 KiB, no links), else the last segment of
// the remote, else the folder name.
//
// # Concurrency
//
// A sync.Mutex serializes Startup and the timeline mutations in a process. Only
// one server may own a state directory (the caller's server.lock). Between
// processes, only the registry is shared (register versus the server) and it is
// guarded by the blocking flock on viewer/lock.
package viewer
