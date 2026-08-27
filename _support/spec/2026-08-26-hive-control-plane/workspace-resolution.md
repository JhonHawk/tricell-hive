# Workspace Resolution

Status: Proposed

Workspace resolution is the critical difference between Hive and a
repository-only review tool. The process current directory is an input, not a
guarantee that one Git repository exists.

## Resolution goals

Given a requested path and a mode, the resolver must return a typed result that
answers all of the following:

- What canonical path was evaluated?
- Is it a workspace, a Git worktree, a bare repository, or unsupported?
- Which repository roots are in scope?
- Were targets explicit or discovered?
- Which discovery limits were applied or reached?
- Which nested boundaries were encountered?
- What Git identity does each target have?
- Where will per-repository authority live?

The result must be reproducible from the same filesystem state and arguments.
Two callers must not get a different target set because one happened to run from
inside a child repository.

## Modes and precedence

The v1 CLI has three resolution modes:

| Mode | Meaning |
| --- | --- |
| `auto` | Resolve explicit workspace context first; otherwise use the nearest declared workspace marker, then the nearest Git worktree, then the canonical cwd as a workspace |
| `repo` | Require one explicit or nearest Git worktree; do not discover siblings or children |
| `workspace` | Treat the requested path as a workspace and discover nested repository roots within bounds |

The selector precedence is:

1. An explicit `--workspace PATH` establishes workspace context. Optional
   `--repo PATH`, `--repo-id`, `--all`, or
   `--path-prefix` selectors constrain the targets inside that workspace.
2. Without `--workspace`, repeated explicit `--repo PATH` selectors
   select repository targets without discovery.
3. Without either explicit root, `--cwd PATH` or the process current
   directory supplies the requested path. In `auto` mode, walk upward from
   its canonical path for the nearest declared workspace marker:
   `.engram/config.json` or `_support/PROJECT.md`. These are compatibility
   markers, not control-plane authority.
4. If no marker is found, `auto` falls back to the nearest Git worktree.
   If the path is outside Git and has no marker, the canonical cwd itself is the
   workspace root.
5. The selected resolution mode and its defaults apply to that root.

Conflicting selectors are usage errors. An explicit repository path never
silently expands into a workspace. An explicit workspace never silently narrows
to the repository containing the process current directory.

The upward marker search inspects only the finite ancestor chain of the
requested path; it never crawls sibling directories or the user's home. A
canonical home directory, filesystem root, or other broad volume root is not
automatically crawled. In that case `auto` returns a workspace result
with `complete_discovery=false` and a `broad_root_requires_explicit_workspace`
reason. The caller must provide an explicit, narrower `--workspace` before
using `--all` for mutation.

## Canonicalization

Canonicalization occurs before identity or deduplication:

1. Require an existing directory unless a command explicitly permits a
   not-yet-created target.
2. Convert to an absolute, cleaned path.
3. Resolve `.` and `..`.
4. For an explicitly supplied workspace or repository path, resolve the final
   target through all existing symlink components and record both requested and
   canonical paths.
5. Fail with `symlink_loop` or `inaccessible_path` when resolution
   cannot complete. Do not continue with a partially canonicalized path.
6. Use filesystem identity where available to detect aliases.
7. Preserve a display path separately from the authority key.

The authority key must not depend on a user-facing spelling such as a relative
path, a symlink alias, or a case variant on a case-insensitive filesystem.
An explicit `--repo PATH` supplied with `--workspace PATH` is safe only
when its canonical target is inside the canonical workspace root. Otherwise
the command returns `unsafe_path`. Repeated explicit `--repo` options
without `--workspace` may intentionally span unrelated workspace roots.

Discovery never follows a symlink directory entry. A symlink can be an explicit
target, but it is not an implicit discovery edge. The same no-follow and
containment checks apply to `--state-root`, policy files, result files,
and computed authority paths. A state, policy, or result path that is a symlink,
escapes its expected root, is inaccessible, or is not the required file/directory
type fails closed.

## Automatic discovery

When `auto` receives a non-Git path, discovery is bounded and deterministic.
The proposed defaults are:

- maximum depth: 3 directory levels below the canonical workspace;
- maximum discovered repositories: 32;
- maximum visited directory entries: 4,096;
- no traversal through symlinks;
- skip `.git` internals and common generated or dependency directories such as
  `node_modules`, `vendor`, `dist`, `build`, `.venv`, and `target`;
- sort entries by canonical name before visiting them;
- return a typed `discovery_limit_reached` result if a bound is exceeded.

These values are proposed defaults, not hidden behavior. The contract must expose
the applied limits and accept explicit bounded overrides for a controlled run.
An implementation must never turn an unbounded recursive walk into the default
behavior.

The result includes `complete_discovery`. It is `true` only when the
walk visited every eligible directory within the requested bounds and verified
all discovered candidates. It is `false` when a limit was reached, a
directory was inaccessible, an implicit broad-root crawl was suppressed, or a
candidate could not be verified. If `complete_discovery` is `false`,
`--all` on any mutating workspace command fails with exit 31 before any
repository authority mutation. Explicit repository paths that were individually
canonicalized and Git-verified may proceed independently, even when a workspace
discovery result is incomplete.

Discovery looks for Git roots using Git's own metadata rather than guessing from
directory names. A `.git` directory or `.git` file is a candidate marker; the
resolver then asks Git for the canonical Git directory, common directory, and
worktree information. If those values cannot be verified, the candidate is
reported as unsupported or invalid rather than selected as a normal repository.

Workspace discovery may continue through ordinary worktree directories, within
the configured bounds, after emitting a repository target. It never enters
`.git` internals and does not traverse symlink directory entries. This
allows an independently nested repository to be found. Once a repository is
selected for authority, its own candidate and authority operations treat nested
repositories as opaque independent boundaries; workspace discovery and
repository projection are separate concerns.

## Zero, one, and many repositories

### Zero repositories

For a non-Git workspace with no discovered repositories:

- `root resolve` succeeds with `scope=workspace`, an empty target list, and
  `data.complete_discovery=true` and `next_transition=stop` with reason
  `no_targets`;
- `workspace status` succeeds and reports an empty set;
- `workspace review start` returns a common-envelope error with
  `error.code=no_targets` and exit 13 unless an explicit verified
  repository selector was supplied;
- no Git authority directory is created.

An empty workspace is a valid scope observation, not an implicit repository.

### One repository

For a workspace containing one discovered repository, the resolver still returns
`scope=workspace` unless the caller selected `repo` mode or an explicit
`--repo`. The distinction matters because workspace semantics include the
discovery facts and future sibling detection.

### Many repositories

For multiple targets, the resolver returns an ordered list and stable target IDs.
A workspace coordinator can create one independent operation per target. A
caller must choose one of:

- an explicit list of `--repo PATH` selectors;
- `--all` discovered repositories;
- `--repo-id ID` selectors from a prior resolution;
- a bounded `--path-prefix PATH` selector.

The default for a mutating command is no implicit fan-out. `status` may describe
all discovered targets, but `workspace review start` requires an explicit
selector or `--all`. `--all` is legal only when
`complete_discovery=true`.

If discovery is incomplete, an explicit list of individually verified
repositories can still be started independently. The workspace coordinator
records the unresolved or skipped targets and does not claim that the workspace
was completely reviewed.

## Nested repositories

Each Git repository is an independent boundary. If a parent workspace contains a
child repository:

~~~text
/work/product/                 workspace
├── .git/                       repository A
├── services/api/               repository B
│   └── .git
└── tools/                      non-Git directory
~~~

The resolver returns A and B as separate targets. It does not synthesize a
combined tree. A repository authority operation rooted at A uses A's Git view;
an operation rooted at B uses B's Git view. The child repository contents are
opaque to workspace aggregation unless the parent repository independently
tracks them according to Git's own rules.

When discovery reaches a repository boundary, it records the boundary and does
not enter that repository's `.git` metadata or any resolved Git common/object
directory. It may continue through ordinary worktree directories and other
descendants, within the same bounds, so an independently nested repository can
also be emitted as a separate target. This prevents walking an object database
accidentally while keeping nested repository discovery complete. Repository
projection remains opaque at each boundary and never incorporates a child
repository's contents.

Nested repositories with the same canonical Git common directory are deduplicated
only when they also identify the same worktree. Different linked worktrees
remain different targets even though they share a common directory.

## Linked worktrees

Git linked worktrees share a common repository administration directory but have
different worktree-specific Git directories and candidate files. The resolver
must capture all three:

- canonical worktree root;
- canonical Git directory;
- canonical Git common directory.

The repository ID can be derived from the common repository identity, but the
worktree ID must be included in the lineage binding. A result from worktree A
must never be replayed against worktree B merely because their branch names or
lineage text match.

The proposed authority layout therefore groups by common directory while
partitioning lineages by worktree identity:

~~~text
<git-common-dir>/tricell-hive/control/v1/
└── repositories/
    └── <repository-id>/
        ├── worktrees/
        │   ├── <worktree-id-A>/
        │   └── <worktree-id-B>/
        └── lineages/
~~~

The exact layout can be normalized in the CLI contract, but the identity
invariant must be tested. Git's official `rev-parse` documentation describes the
worktree-aware Git directory and common directory options:
[Git `rev-parse`](https://git-scm.com/docs/git-rev-parse) (accessed
2026-08-26).

## Bare repositories

A bare repository has Git metadata but no worktree candidate. Automatic
discovery reports it as `unsupported_target` and continues when other targets
remain. A mutating review or verification operation cannot select it in v1
because the proposed candidate model is worktree-based.

An explicit future contract could add a committed-object mode, but that is a
different candidate projection and must not be inferred from a bare repository.

## Workspace coordination and failure

Workspace coordination is an ordered fan-out/fan-in operation:

~~~text
resolve
  │
  ├── start repository A ──> approved
  ├── start repository B ──> concurrent_update
  └── start repository C ──> unsupported_target
  │
  ▼
partial_failure
~~~

The summary must include every selected target and its terminal or non-terminal
result. There is no rollback of A because B failed. There is no workspace-level
approval. The caller may retry B using its exact repository binding or abandon
the workspace run; it must not silently restart A.

Workspace lifecycle commands are keyed by `workspace_run_id`:

| Command | Behavior |
| --- | --- |
| `workspace inspect` | Read the immutable selection and every per-target outcome |
| `workspace retry` | Retry only selected targets whose outcome is failed or non-terminal; completed targets are never restarted |
| `workspace abandon` | Explicitly stop the workspace run and preserve per-target outcomes |

`workspace review start` creates the run and its independent target
operations. `workspace retry` requires the canonical workspace path,
the existing run ID, and an idempotency key. It revalidates the stored
selection; if a target's repository, worktree, policy, or candidate binding
changed, that target returns `binding_mismatch` and is not restarted.
The workspace run itself is not approval authority.

## Examples

### Non-Git workspace with nested repositories

~~~bash
hive-control root resolve \
  --cwd /work/product \
  --mode workspace \
  --max-depth 3 \
  --max-repositories 32 \
  --json
~~~

Expected result: a workspace envelope with zero or more canonical repository
targets, discovery limits, nested-boundary metadata, `data.complete_discovery`,
and no mutation.

### Explicit repository selection from a non-Git workspace

~~~bash
hive-control review status \
  --cwd /work/product \
  --repo /work/product/services/api \
  --next-transition \
  --json
~~~

Expected result: one repository-scoped status bound to the canonical API
worktree. The parent workspace is not used as a fallback.

### Explicit multi-repository selection

~~~bash
hive-control workspace review start \
  --workspace /work/product \
  --repo /work/product/services/api \
  --repo /work/product/apps/web \
  --json
~~~

Expected result: a workspace operation containing a `workspace_run_id` and
two independent start transitions. If either start fails, the other result
remains independently visible; no cross-repository atomicity is claimed.

### Retry only failed targets

~~~bash
hive-control workspace retry \
  --workspace /work/product \
  --workspace-run-id WR-01 \
  --idempotency-key retry-WR-01-1 \
  --json
~~~

The retry response includes all targets in `data.target_outcomes`, marks
completed targets as `skipped_completed`, and attempts only failed or
non-terminal selected targets.
