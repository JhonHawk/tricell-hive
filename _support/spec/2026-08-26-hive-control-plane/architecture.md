# Architecture

Status: Proposed

This document defines the proposed boundary for a local Hive control plane. It
focuses on deterministic scope resolution, state binding, persistence, and
failure behavior. It deliberately leaves human judgment and publication outside
the executable.

## Boundary

The proposed binary is `hive-control`. A shell hook, skill, or harness
adapter invokes it, receives exactly one JSON envelope, and decides whether to
relay the envelope to the model or execute the exact argv returned in a
transition.

~~~text
Claude / Codex / OpenCode / Grok / shell
                    │
                    │ direct process invocation
                    ▼
             hive-control
       ┌────────────┼────────────┐
       ▼            ▼            ▼
  root resolver  authority    contract
       │            │            │
       ▼            ▼            ▼
 workspace      per-repo      JSON v1
 and targets    state/CAS     envelopes
~~~

The binary owns mechanical facts:

- canonical paths and repository identity;
- target selection and discovery bounds;
- candidate and policy bindings;
- legal state transitions;
- evidence references and verification freshness;
- locking, revision checks, and atomic persistence;
- machine-readable continuation or stop reasons.

The binary does not own:

- interpretation of the user's natural-language request;
- architecture or product decisions;
- classification of a finding as correct, severe, or candidate-caused;
- approval of publication actions;
- Engram observations or tracker status;
- the content of a human-authored Flow artifact.

All machine responses use one versioned envelope with the fields `schema`,
`kind`, `ok`, `request_id`, `observed_at`, `scope`,
`data`, `next_transition`, and `warnings`. An `error` object is
optional and is present when `ok` is false. This common shape lets empty,
partial, and failing workspace operations remain machine-readable without
inventing a second error protocol.

Every non-null `next_transition`, including `collect` and `stop`, contains a
canonical absolute `cwd`. Failures that occur before the requested path can be
canonicalized use `scope.canonical_path=null` and `next_transition=null` while
retaining `scope.requested_path`. An `execute` transition additionally contains
an argv whose first element is
`filepath.EvalSymlinks(os.Executable())`. Failure to resolve and canonicalize
that executable path fails closed. This avoids depending on the adapter's PATH
and makes the executable that produced the transition explicit.

This is a local controller pattern inspired by the reconciliation model described
in the [Kubernetes controller documentation](https://kubernetes.io/docs/concepts/architecture/controller/)
(accessed 2026-08-26). Hive does not need Kubernetes, a cluster, or a remote
control plane.

## Proposed module layout

The first runtime implementation should use a root Go module:

~~~text
tricell-hive/
├── go.mod
├── cmd/
│   └── hive-control/
│       └── main.go
├── internal/
│   ├── control/
│       ├── cli/             # argv parsing and stdout/stderr boundary
│       ├── contract/        # versioned envelopes and schema validation
│       ├── root/            # workspace/repository resolution
│       ├── gitidentity/     # Git dir, common dir, worktree identity
│       ├── authority/       # per-repository transaction state machine
│       ├── workspace/       # multi-repository coordination
│       ├── candidate/       # immutable candidate and policy bindings
│       ├── store/           # atomic files, locks, CAS, quarantine
│       ├── security/        # path, permission, symlink, and redaction rules
│       └── transition/      # legal next-transition derivation
│   └── controltest/         # shared temporary workspace and Git fixtures
├── contracts/
│   └── hive-control/
│       └── v1/
│           ├── schemas/
│           ├── fixtures/
│           └── README.md
~~~

The layout follows the organization described by the
[Go module layout guidance](https://go.dev/doc/modules/layout) (accessed
2026-08-26). The command package must remain thin. Domain decisions belong in
internal packages so tests can exercise them without spawning a process.

The implementation should start with the standard library. A new dependency
requires a separate compatibility and supply-chain review; it must not be added
just to parse flags, write JSON, or lock a local file.

## State model

The state model has two scopes.

### Repository authority

One repository target has one independently mutable authority lineage. The
lineage binds:

- repository canonical worktree root;
- Git directory and Git common directory;
- worktree identity;
- repository identity;
- opaque authority ID;
- selected candidate projection;
- canonical candidate manifest digest;
- policy content digest and contract version;
- creation and current revision;
- selected checks/reviewers;
- collected result identities;
- correction budget and validation budget;
- current lifecycle state.

The proposed states are:

| State | Meaning | Legal next actions |
| --- | --- | --- |
| `preflighted` | Current scope was resolved and a candidate is available | `start`, `abandon`, read-only status |
| `reviewing` | Candidate and policy are frozen for review/verification | `prepare-result`, `capture-check`, `capture-review`, `abandon`, read-only status |
| `correction_required` | A bounded candidate correction is required | `capture-correction`, `abandon` |
| `validating` | A corrected candidate or evidence set is being validated | `prepare-result`, `capture-validation`, `abandon` |
| `approved` | The controlled boundary completed; evidence is final | `burn` or read-only status |
| `escalated` | The bounded path cannot continue automatically | `recover`, `abandon`, read-only status |
| `abandoned` | The owner explicitly stopped this lineage | read-only status |
| `invalidated` | The candidate, policy, or binding changed, or this lineage was superseded by a correction | `start`, `abandon`, or read-only status |

`preflighted` is a read-only projection returned by selectorless status; it is
not persisted and has no authority ID. `review start` consumes that projection
and creates the first persisted state, `reviewing`.

`approved` and `abandoned` are terminal lifecycle states. `approved` is an
evidence result, not a publication permission, and may be burned only after its
evidence has been durably returned to the caller. `abandoned` is terminal but
is not burnable in v1; it retains a minimal inspectable outcome for explicit
cleanup policy. `burn` removes mutable authority, candidate snapshots, intake
files, and reviewer results from an approved lineage. It retains only a
minimal idempotency outcome without candidate content, prompts, or source. If
the project needs a retained report, that report belongs to the ordinary
documentation or session path, not to the authority store. A burn replay with
the same idempotency key returns the original result; a new key for the burned
authority returns `state_absent`.

### Authority identity

`authority_id` is an opaque, cryptographically random identifier allocated
when a lineage starts. The authority record binds it to the canonical
repository root, Git directory, Git common directory, worktree identity,
repository ID, and lineage. Callers must not parse it or derive a path from it.

It is stable for the life of that lineage, including process restarts and
compaction. It is never reused after burn, abandonment, or invalidation. The
store checks the full binding before accepting an ID. A collision or an ID
already bound to a different tuple fails closed with `authority_id_collision`;
the operation does not overwrite either record. Repository ID, worktree ID, and
lineage are verification fields and remain in every bound result, but they are
not authority lookup keys.

### Workspace coordination

A workspace scope is an ordered set of repository target records plus the
resolution facts that produced them:

- requested path and canonical path;
- resolution mode and explicit selectors;
- discovery limits and whether they were reached;
- selected repository IDs;
- per-repository operation state and outcome;
- unresolved paths or unsupported targets;
- workspace run ID and contract version.

The workspace record is an index and coordination record. It is not a synthetic
Git tree and cannot approve a collection of repositories as one transaction.

### Candidate binding

The candidate binding is immutable for a lineage. A later status operation
recomputes the live identity and compares it with the binding. At minimum it
must detect:

- worktree root replacement or canonical path change;
- Git common directory or Git directory change;
- target branch/base identity change;
- any projected tracked or eligible untracked entry change;
- policy bytes or policy hash change;
- contract major-version change.

When a binding differs, the result is `invalidated` with a reason code. The
caller must not silently reuse a green result for the new candidate. A
correction is the explicit exception in the lifecycle: capture-correction
verifies the old binding, then creates a successor authority with a new
authority ID and lineage bound to the corrected candidate. The old lineage is
marked invalidated/superseded; its immutable candidate binding is not edited.

## Candidate projection and digest

The v1 projection is named `worktree-v1`. It represents the exact candidate
available from one canonical worktree at start time. It is a manifest, not a
hash of an unspecified directory and not a copy of `.git` metadata.

The implementation builds the manifest from Git and the filesystem:

1. Read the `HEAD` OID and the selected base OID, if a base is part of the
   command. An absent base is represented explicitly, never as an empty string
   that could collide with a real value.
2. Read `git ls-files --stage -z` and include every tracked path with its
   index stage and index mode.
3. For each tracked path, read the worktree object without following a symlink.
   Include regular-file content, symlink target bytes, or an explicit missing
   marker. A tracked submodule is one entry of type `gitlink` (mode `160000`)
   with its recorded commit OID as payload; its nested contents are not
   included. A tracked path remains included even when an ignore pattern also
   matches it; ignore rules apply only to untracked enumeration.
4. Read non-ignored untracked paths using Git's standard exclude rules. Include
   only regular files and their content. Ignore ignored files, directories,
   symlinks, devices, FIFOs, and sockets. Untracked symlinks therefore have no
   entry; tracked symlinks are represented by step 3.
5. Exclude every `.git` directory or `.git` file and any explicitly configured
   Hive authority or workspace-state subtree if one is below the candidate
   root. There is no broad dot-file or editor-file exclusion: a tracked or
   non-ignored untracked regular file is included unless it is one of those
   exact control paths. This makes the exclusion set testable rather than
   dependent on a tool-specific metadata list.
6. Normalize each relative path to slash-separated bytes relative to the
   canonical worktree root. Sort entries by raw path bytes, then stage, then
   type, then mode. Never sort by display path or locale.

The candidate manifest is serialized with length-prefixed fields so that
concatenation cannot create ambiguous records. `head_oid` is required; when no
base was selected, `base_oid` is the literal tagged value `absent` rather than
an omitted or empty field:

~~~text
manifest =
  SHA256(
    "tricell.hive-control/candidate/v1\0" ||
    len(projection) || projection ||
    len(head_oid) || head_oid ||
    len(base_oid) || base_oid ||
    repeat(entry)
  )

entry =
  len(path_bytes) || path_bytes ||
  len(stage) || stage ||
  len(mode) || mode ||
  len(type) || type ||
  len(content_or_target_bytes) || content_or_target_bytes
~~~

The framing uses unsigned big-endian lengths and byte strings; the domain
separator is part of the hash input. An untracked entry uses stage `untracked`;
this distinguishes it from a tracked index entry. A missing tracked file uses
type `missing` and an empty payload. A symlink uses type `symlink` and
the link target bytes. A regular file uses type `regular` and exact
content bytes. An untracked regular file's mode is normalized to the executable
bit only (`100644` or `100755`) to avoid platform-specific
permission noise; tracked entries retain the index mode. The manifest records
the projection name, entry count, excluded-entry count, and digest alongside
the authority binding.

The policy digest is calculated separately with
`SHA256("tricell.hive-control/policy/v1\0" || policy_bytes)`. A reviewer or
check result must carry both the candidate manifest digest and policy digest.
The candidate digest must never be called merely “candidate bytes” in the
implementation or contract.

## Persistence

### Repository-local authority root

For a Git repository, the proposed authority root is derived from the canonical
Git common directory:

~~~text
<git-common-dir>/tricell-hive/control/v1/
├── lock
├── maintenance.lock
├── repositories/
│   └── <repository-id>/
│       ├── lineage.lock
│       └── <lineage-id>/
│           ├── state.json
│           ├── candidate.json
│           ├── intake/
│           ├── checks/
│           └── reviewers/
├── operations/
│   └── <idempotency-key>.json
├── incidents/
└── quarantine/
~~~

The exact directory names are subject to the v1 contract, but the invariants
are not: it is outside the worktree, keyed by the common directory, and
distinguishes linked worktrees by worktree identity. A linked worktree therefore
shares the repository authority store without making its candidate interchangeable
with another worktree.

The `operations` records are intentionally smaller than review authority:
they retain the idempotency key, request fingerprint, outcome envelope, and
authority ID needed to replay an identical mutation after `burn`. They
must not retain source, prompts, environment values, or raw reviewer results.
Retention and bounded cleanup of these records are a separate maintenance
operation; cleanup cannot make an authority ID reusable.

Git documents `--git-common-dir` as the location shared by worktrees, while
`--git-dir` remains worktree-specific. The resolver must use both values and
validate them instead of inferring one from a path. See the official
[Git `rev-parse` documentation](https://git-scm.com/docs/git-rev-parse)
(accessed 2026-08-26).

### Non-Git workspace state

A workspace that has no Git target can still be resolved and reported. Its
coordination record belongs in a user-local root selected by the platform state
directory policy below:

~~~text
<state-root>/workspaces/v1/<workspace-id>/
├── workspace.json
└── runs/
    └── <run-id>.json
~~~

The default `state-root` is platform-specific and exact. These locations are
Hive policy choices, not behavior inferred from a Go directory API:

| Platform | Default |
| --- | --- |
| macOS | `~/Library/Application Support/Tricell/hive-control` |
| Linux | `$XDG_STATE_HOME/tricell/hive-control`, falling back to `~/.local/state/tricell/hive-control` |
| Windows | `%LOCALAPPDATA%\\Tricell\\hive-control` |

An absolute `--state-root` override is allowed for tests and managed
environments. The root and every existing component below it must be owned by
the current local user, must not be a symlink, and must have owner-only
permissions before a write. The v1 state model is same-local-user only; it
does not claim safe coordination through a shared network filesystem or across
OS user accounts.

It must not silently create authority under an arbitrary non-Git workspace. A
non-Git workspace can coordinate discovery and report `no_targets`; it
cannot create a repository review lineage without an explicit Git target.

The workspace ID is opaque and derived from the canonical workspace root:

~~~text
workspace_id =
  base32url(
    SHA256("tricell.hive-control/workspace/v1\0" || canonical_root_bytes)
  )
~~~

The full digest is used; truncation is not allowed in v1. Renaming a workspace
changes its canonical root and therefore produces a new ID. Existing state
under the old ID is not adopted automatically. V1 permits explicit inspection
or abandonment of the old run; workspace migration is outside the v1 contract.

### Atomicity and concurrency

All mutable records use:

1. an exclusive lock for the smallest relevant scope;
2. a read and validation of the expected current revision;
3. a state transition check;
4. a write to a temporary file in the same directory;
5. flush and close;
6. atomic rename into place;
7. directory synchronization where the platform supports it.

Each successful replacement increments a monotonic revision. A stale caller
gets a `concurrent_update` error and must re-read status; it must not
overwrite the record or fabricate a continuation. Exact retries of an
idempotent request must return the original result and already-applied revision.

The control plane must not claim crash-proof durability beyond what the host
filesystem provides. It can guarantee that a completed replacement is either
the prior valid record or the new valid record where the platform's rename and
sync semantics support that guarantee. A malformed or incomplete record fails
closed and is moved to quarantine only through an explicit repair path.

## Lifecycle operations

Every mutable transition has a corresponding CLI operation and a required
idempotency key:

| Transition | Operation | Result |
| --- | --- | --- |
| create | `review start` | Creates one authority ID and enters `reviewing` |
| collect | `review prepare-result` then one `capture-*` command | Allocates a real intake path, advances revision, and returns the exact capture argv for that new revision |
| correction | `review capture-correction` | Verifies the old binding, records correction evidence, and creates a successor authority in `validating` |
| recovery | `review recover` | Reopens a legal escalated lineage or creates a successor with a new authority ID |
| terminal disposal | `review burn` | Removes the approved authority and evidence, retaining only its idempotency outcome |
| explicit stop | `review abandon` | Marks the lineage abandoned and preserves a minimal inspectable outcome |

`prepare-result` is a mutation. After allocating the owner-only intake file, it
increments the authority revision and returns an `execute` transition containing
the exact capture command, the allocated path, and the new expected revision.
The adapter writes only to that returned path and executes only that argv.

`recover` never guesses a target. It requires the existing
`authority_id`, canonical `--repo`, verification IDs, expected
revision, and an explicit recovery reason. If the candidate or worktree binding
changed, recovery must create a new authority ID after a fresh start rather
than weakening the old binding.

`capture-correction` consumes the old `correction_required` authority's bounded
correction budget. Its result carries the old authority binding and source
candidate digest, while `hive-control` recomputes the corrected worktree-v1
manifest. If the corrected digest matches the result, the operation atomically
marks the old lineage invalidated/superseded and creates a new authority ID and
lineage in `validating`; it never edits the old candidate binding. A mismatched
corrected digest fails before mutation.

`burn` is legal only for a verified `approved` authority. It
must verify the complete binding and expected revision while holding the
lineage lock, persist the idempotency outcome, remove mutable authority and
evidence, and then return a common envelope with
`data.disposition=burned`. It never emits a publication command. A
replay with the same key and identical request fingerprint returns that original
envelope; a different key cannot address the burned authority.

## Workspace and repository ownership

The resolver returns repository targets; the authority layer consumes exactly one
target at a time. A workspace coordinator can fan out work:

~~~text
workspace resolve
        │
        ├── repo A ──> authority A ──> result A
        ├── repo B ──> authority B ──> result B
        └── repo C ──> authority C ──> result C
                         │
                         ▼
                 workspace summary
~~~

There is deliberately no cross-repository atomic transaction in v1. If repo B
fails after repo A succeeds, the result records `partial_failure` and
preserves both per-repository outcomes. It does not undo A, edit B, or report
the workspace as approved. A later version could add an explicit coordinator
protocol, but that would be a new contract and a new rollback problem.

## Security and failure semantics

The executable is local, but its state can influence subsequent agent actions.
The following are mandatory:

- Use argument arrays and `exec.Command`; never build a shell command string from
  user or repository input.
- Require an explicit absolute canonical `--repo PATH` on every
  repository-scoped command. Initial root selection may resolve a user-facing
  alias, but every returned bound argv uses the canonical path.
- Resolve an explicitly selected symlink path to its final target, record both
  requested and canonical paths, and fail on loops, missing components, or
  inaccessible components.
- Do not follow symlink directory entries during discovery. An explicit
  `--repo` path supplied with `--workspace` must resolve inside the
  canonical workspace or return `unsafe_path`.
- Validate `--state-root`, policy, result, and computed authority paths
  with no-follow checks; reject unsafe symlinks, ownership failures, and escapes
  from their expected root.
- Create state directories with owner-only permissions and state files with
  owner read/write permissions.
- Never persist secrets, environment contents, prompts, or full candidate source
  unless a future contract explicitly requires it.
- Redact sensitive values in error messages and preserve structured error codes.
- Refuse to mutate when repository identity, candidate binding, policy hash, or
  state schema cannot be verified.
- Treat malformed JSON, missing locks, stale revisions, ambiguous roots, and
  discovery-limit exhaustion as explicit failures, not empty success.
- Keep stdout machine-readable; human diagnostics go to stderr.
- Do not interpret an `approved` state as commit, push, PR, merge, release, or
  deployment authorization.

The model is adversarially safer when the binary returns one legal transition
with exact arguments instead of expecting the model to reconstruct a command.
The binary is not an authentication boundary against a malicious local actor
who can rewrite the same user's filesystem. That limitation must be explicit in
the threat model and documentation.

The only `execute` transitions the CLI may return are allowlisted Hive
lifecycle operations: `review start`, `review prepare-result`,
`review capture-check`, `review capture-review`,
`review capture-correction`, `review capture-validation`,
`review recover`, `review burn`, `review abandon`, and
their explicit workspace equivalents. It must never return `commit`,
`push`, `PR`, `merge`, `release`, or
`deploy` argv.

## Shadow and cutover boundary

In shadow mode, `hive-control` may resolve scope, compute bindings, and emit
advisory results. Existing Hive rules remain the sole authority. Shadow output
must be bounded and cheap: it should not duplicate full prompts or copy source
content into a second store.

At cutover, replace only the duplicated review/verification procedure:

- status reconstruction;
- candidate freshness;
- check/reviewer bookkeeping;
- bounded continuation counters;
- per-repository result aggregation.

Preserve the existing safety gates, human judgment, publication policy, Engram,
tracker, Flow artifacts, and ordinary Git policy. A cutover is complete only
when the old mechanical path is retired or frozen under a named rollback
milestone; permanent dual authorities are not a supported design.
