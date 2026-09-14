# Hive Control Plane

Status: Parked (2026-09-13)

Date: 2026-08-26

This folder describes a possible deterministic control plane for Tricell Hive. It
is a design proposal, not an implementation record and not an approval to change
Hive. It was parked on 2026-08-31: the 2026-08-26 measurement found the cost it
targets (re-deriving lifecycle state from conversation) near zero against a
fixed rule load of ~27,200 tokens per session. A re-measurement over a fresh
corpus started on 2026-09-13; its verdict is recorded in
`_support/archive/audits/2026-09-13-hive-control-remeasurement.md`. Until that
verdict says proceed, no implementation, deployment, or global installation is
authorized.

## Executive summary

The proposal is a small Go command named `hive-control`. It would be invoked
directly by shell adapters, hooks, or skills and would return versioned JSON. It
would own only mechanical workflow facts that are currently reconstructed from
conversation:

- which workspace or repository is in scope;
- which candidate and policy were observed;
- which checks or reviewers have run;
- which transition is legal next;
- whether an earlier result is still valid after the candidate changes.

It would not become an MCP server, a background daemon, a second memory system,
or a publication authority. The model would still interpret intent, make design
judgments, assess findings, and request or perform separately authorized
publication actions.

The most important requirement is workspace-first operation. The process current
directory may be:

- a non-Git workspace containing no repositories;
- a non-Git workspace containing one repository;
- a non-Git workspace containing many nested repositories;
- a repository root or a subdirectory of one repository;
- a linked worktree;
- a directory containing independent nested Git repositories.

The resolver must represent those cases explicitly instead of assuming that
`cwd` can always be passed to `git` as the one project root.

~~~text
agent or hook
     │  exact argv, no shell interpolation
     ▼
hive-control
     ├── resolve workspace and repository targets
     ├── bind candidate, policy, and worktree identity
     ├── read/write local authority with locks and CAS
     └── return one typed next transition
     │
     ├── per-repository review/verification state
     └── workspace coordination state
~~~

## Problem, actors, and trust boundary

The problem is not that Hive lacks a place to put prose. The problem is that
long sessions repeatedly reconstruct scope, candidate freshness, reviewer
slots, bounded retries, and legal next actions from conversation. That
reconstruction costs context and can drift after compaction or a changed
worktree.

The proposed actors are:

| Actor | Responsibility and trust |
| --- | --- |
| User | Chooses scope, resolves ambiguity, supplies judgment, and authorizes publication |
| Model and harness adapter | Interprets intent and transports exact CLI results; treated as unable to invent authority |
| `hive-control` | Local mechanical authority for identity, evidence binding, transitions, and persistence |
| Git and filesystem | Sources of repository/worktree state and local durability; not a hostile remote service |
| Engram, tracker, and Flow | Independent memory, tracking, and durable-artifact authorities; not replaced |

The CLI is a trust boundary for workflow mechanics, not an authentication
boundary against a malicious actor with the same user's filesystem access. It
must fail closed on unverified identity, path, policy, evidence, or state, while
leaving user judgment and publication authorization outside the binary.

## Current baseline

The baseline was checked on 2026-08-26:

| Item | Observed state |
| --- | --- |
| Go toolchain | `/usr/local/go/bin/go version` reports `go1.27.0 darwin/arm64` |
| Shell lookup | `go` is not currently on this session PATH; this proposal does not change PATH |
| Hive module | No `go.mod` exists |
| Hive runtime | No Go files or runtime binary exist |
| CI | Hive has no runtime CI/build pipeline |
| Repository class | Configuration-only repo; current validation is document and consistency review |
| Git mode | Hold; no commit or push for this proposal |

Introducing the runtime would be a deliberate change to the repository class.
The first implementation must update the repository documentation and validation
model in the same change group; this proposal does not make that change.

## Proposed decisions

These are recommendations for review. They are not accepted decisions.

1. **Use a root Go module.** Add `go.mod`, `cmd/hive-control`, and
   `internal/control/*` at the Hive root. Keep versioned JSON Schemas and
   conformance fixtures under `contracts/hive-control/v1/`.
2. **Use a direct CLI for v1.** The integration boundary is an executable and
   JSON on stdout. MCP and a long-lived daemon are intentionally excluded from
   the first version.
3. **Resolve scope before Git operations.** A resolver classifies the requested
   path as a repository, workspace, or unsupported target. It can return zero,
   one, or many canonical repository targets.
4. **Store authority outside worktrees.** Repository authority is anchored at
   Git's common directory, so linked worktrees share the authority root while
   retaining distinct worktree identities. A workspace without Git uses a
   user-local workspace state root.
5. **Separate repository authority from workspace coordination.** Each
   repository gets an independent transaction. A multi-repository workspace gets
   an aggregate plan and per-repository outcomes, but v1 does not offer a
   cross-repository atomic transaction.
6. **Keep current Hive rules authoritative in shadow mode.** The first live
   integration observes and compares proposed transitions without replacing the
   existing rules.
7. **Cut over only the duplicated mechanical procedure.** At cutover, the
   control plane may replace the repeated review/verification bookkeeping.
   Safety gates, human judgment, publication authorization, Engram, tracker,
   Flow artifacts, and ordinary Git policy remain authoritative.
8. **Use one common JSON envelope.** Success, error, empty-scope, and partial
   workspace results share the same top-level fields; `error` is optional and
   `data` carries the operation-specific payload.
9. **Bind every repository lifecycle call to an opaque `authority_id`.** Bound
   calls also carry an explicit canonical `--repo PATH`; `repo_id`,
   `worktree_id`, and `lineage` verify the binding but do not locate it.
10. **Model the full lifecycle.** `capture-correction`, `recover`, and
    `burn` are first-class commands. Burn removes mutable authority while
    retaining only the minimum idempotency outcome needed for safe replay.
11. **Make workspace operations explicit.** `workspace review start`,
    `workspace inspect`, `workspace retry`, and
    `workspace abandon` are keyed by `workspace_run_id`; retries never
    restart completed targets.
12. **Require complete discovery for implicit fan-out.** `complete_discovery=false`
    makes `--all` mutation fail with exit 31 before any target mutates.
13. **Define a testable `worktree-v1` candidate projection.** A sorted,
    domain-separated SHA-256 manifest binds Git OIDs, index stages/modes,
    tracked content, and non-ignored untracked regular files.

The detailed rationale and state model are in [architecture.md](architecture.md).
The workspace cases are in [workspace-resolution.md](workspace-resolution.md).
The machine-facing interface is in [cli-contract.md](cli-contract.md).
The delivery sequence and acceptance criteria are in
[implementation-plan.md](implementation-plan.md).

## Expected value

The control plane could reduce context when a session is long or resumes after
compaction. Instead of restating a lifecycle, the model would receive a compact
status envelope containing the current binding, evidence summary, and one legal
next transition. That can also make the following facts deterministic:

- a check result belongs to an exact candidate and policy revision;
- a changed candidate invalidates a previous gate;
- a stale writer cannot silently overwrite a newer state;
- a workspace result names exactly which repositories were selected;
- a partial multi-repository run reports each repository independently;
- a capture result can be replayed or rejected deterministically by idempotency
  key, authority binding, policy hash, and candidate digest.

This is context reduction by moving mechanical state to a verified local record.
It is not context reduction by hiding safety policy or compressing a human
decision into an opaque status code.

## Non-goals

The first version must not:

- infer user intent or choose architecture;
- decide whether a finding is correct or candidate-caused;
- authorize commits, pushes, pull requests, merges, releases, or deployments;
- replace Engram or the tracker;
- make a non-Git workspace behave as if it were one synthetic repository;
- create a cross-repository commit, rollback, or transaction;
- require an MCP server, network service, or background process;
- write configuration into global tool directories without a separate explicit
  deployment authorization.

## Relationship to optional reference project RDD

The design borrows the useful boundary from the inspected optional reference project snapshot
`7afe50d1` (`v2.5.0-rc.1`; the local reference checkout is at `v2.8.2`,
`266574b0`, as of 2026-09-13): a native executable owns candidate identity,
state transitions, immutable evidence, bounded continuation, and local
authority; adapters transport opaque commands and results. The proposal is not a
port of optional reference project's review lifecycle.

In particular, Hive must adapt the boundary to a workspace that may not be a Git
repository. Git identity is one target type, not the precondition for entering
the control plane. The proposed workspace resolver and the explicit
no-cross-repository-atomicity rule are Hive-specific design constraints.

## Review gates before implementation

Before writing Go code, the owner must confirm:

- the contract name and versioning policy;
- whether review/verification is the first controlled boundary;
- the default discovery limits and explicit selection rules;
- the user-local state-root policy for non-Git workspaces;
- the binary distribution policy and whether `deploy-global` may be extended;
- the shadow metrics that justify an A/B cutover;
- the exact replacement boundary and the artifacts that must survive it.

Approval of this proposal would authorize a later implementation plan only when
the owner explicitly says to proceed. It does not authorize Go installation,
PATH changes, changes to `deploy-global`, changes to hooks, or a global deploy.

## Integration impact inventory

The following are candidate Hive surfaces for a future Phase 4 adapter. This is
an initial inventory, not a verified call-site map and not authorization to edit
any of them:

- `global/core-sections/work-style-delegation.md`;
- `global/core-sections/work-style-execution.md`;
- `global/rules-situational/agent-routing.md`;
- `global/rules-situational/git-mechanics.md`;
- `global/skills/flow-build/references/verify-gate.md`;
- `global/hooks/reviewer-guard/reviewer-guard.sh`;
- `global/hooks/reviewer-guard/test-reviewer-guard.sh`; and
- generated harness outputs produced by `harness/build.py`.

Before Phase 4, the implementation must run a bounded, `rg`-based inventory
from the current Hive tree, map each call site to the contract operation it
would consume, and record the result. No sibling repository changes are in
scope. Extending `deploy-global` or installing a binary into a global
tool directory is a separate gated change.

## Sources

Official sources were accessed on 2026-08-26:

- [Go module layout](https://go.dev/doc/modules/layout) — guidance for
  organizing a Go module with commands and internal packages.
- [Go 1.27 release notes](https://go.dev/doc/go1.27) — release and toolchain
  reference for the installed `go1.27.0` toolchain.
- [Go `os.Executable`](https://pkg.go.dev/os#Executable) and
  [`filepath.EvalSymlinks`](https://pkg.go.dev/path/filepath#EvalSymlinks) —
  executable lookup and symlink canonicalization for returned argv.
- [Git `rev-parse` documentation](https://git-scm.com/docs/git-rev-parse) —
  `--git-dir`, `--git-common-dir`, and worktree-aware identity resolution.
- [Kubernetes controller pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
  — conceptual reference for reconciling observed state to one next action;
  Hive would use the idea locally, not Kubernetes itself.
- [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12) —
  schema dialect for the versioned CLI envelopes and fixtures.

The optional reference project implementation facts were read from the local reference checkout
at `../reference/optional reference project` (a sibling of this repository) on the `v2.5.0-rc.1`
snapshot. The principal local
references are `docs/review-integration.md`,
`docs/architecture/organic-rdd.md`,
`internal/reviewtransaction/compact_store.go`, and
`internal/reviewtransaction/compact_burn.go`.
