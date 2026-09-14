# Implementation Plan

Status: Proposed

This plan describes a future implementation sequence. It is not an approval to
start coding, modify deploy-global, install files globally, or change the
current Hive rules. The current session is hold.

## Prerequisites and gates

Before implementation begins:

1. Approve the boundary in [architecture.md](architecture.md).
2. Approve the workspace-resolution defaults and explicit-selector semantics in
   [workspace-resolution.md](workspace-resolution.md).
3. Approve the v1 CLI name, contract ID, common envelope, schemas, lifecycle
   commands, and exit-code classes in [cli-contract.md](cli-contract.md).
4. Confirm that review/verification bookkeeping is the first controlled
   boundary. Engram, Flow, tracker, publication, and deployment remain outside.
5. Confirm the replacement shape: retain the current procedure as the
   authoritative shadow path, then completely replace only the duplicated
   review/verification procedure after A/B evidence. No permanent dual
   authorities.
6. Confirm the binary distribution policy. Any extension to deploy-global or
   installation under a user-global directory requires separate explicit
   authorization.

Environment facts already verified:

- /usr/local/go/bin/go reports go1.27.0 darwin/arm64.
- go is not on the current session PATH; this plan does not repair PATH.
- Hive currently has no go.mod, Go files, runtime CI, or runtime binary.
- Hive is documented as configuration-only today.

The first implementation changes that repository classification. The
repository's AGENTS guidance, validation instructions, generated-profile
classification, and README must be updated in the same change group as the
runtime, subject to the normal review gate. Those files are outside this
proposal's hold-only scope.

## Proposed A/B cutover thresholds

These are proposed thresholds, not accepted release criteria. The approval
owner is the Hive repository/architecture owner; the named approver and target
values must be recorded before Phase 5. A cutover is not justified by a
smaller prompt alone.

- Zero wrong-root or false-empty outcomes in the scenario corpus.
- Zero publication or safety violations: no returned publication argv and no
  bypass of the existing authorization gates.
- 100% detection of stale-result, candidate, and linked-worktree mismatches.
- Zero unexplained disagreements between the current rules and the control
  plane on selected targets, state, or legal transition.
- A measured target of at least 20% median reduction in lifecycle-control tokens
  per comparable session, without hiding judgment or policy text.
- Bounded discovery p95 at or below 500 ms on a local SSD fixture containing
  32 repositories and 4,096 visited entries, with no bound silently exceeded.
- At least three repetitions per resolver/lifecycle fixture and at least 30
  comparable repository/workspace sessions across the A/B lanes.

The owner must approve the threshold results and replacement boundary. A failed
threshold keeps the current procedure authoritative and leaves the CLI in
shadow or removes the adapter; it does not authorize a weakening of a safety
gate.

Before Phase 5 begins, the named approver and accepted values must be recorded
in `_support/spec/2026-08-26-hive-control-plane/cutover-decision.md`.
Phase 5 must refuse to start while that record is absent or still proposed.

## Integration impact inventory

The following are candidate Hive surfaces, based on the initial repository
review, not a verified call-site map:

- global/core-sections/work-style-delegation.md;
- global/core-sections/work-style-execution.md;
- global/rules-situational/agent-routing.md;
- global/rules-situational/git-mechanics.md;
- global/skills/flow-build/references/verify-gate.md;
- global/hooks/reviewer-guard/reviewer-guard.sh;
- global/hooks/reviewer-guard/test-reviewer-guard.sh; and
- generated harness outputs produced by harness/build.py.

Before Phase 4, run a bounded rg inventory from the Hive root, inspect the
deciding call sites, and map each one to a contract operation. Record the
verified result before editing any integration surface. No sibling repository
changes are in scope. Extending deploy-global, copying a binary to a global
directory, or changing global installation remains a separate gated work item.

## Phases

### Phase 0 — Contract and fixtures

Deliver:

- a root go.mod using the approved Go version policy;
- contracts/hive-control/v1/schemas;
- contracts/hive-control/v1/fixtures;
- typed Go structures for the one common envelope, optional error, scope,
  data, and all transition kinds;
- policy and evidence schemas, including named slots, digests, authority
  binding, and bounded correction/validation budgets;
- candidate fixtures for the exact worktree-v1 projection;
- golden tests for stable fields, selectors, lifecycle reachability, and exit
  classes.

The common envelope must always expose schema, kind, ok, request_id,
observed_at, scope, data, next_transition, and warnings; error is present only
when ok=false. Do not add hooks, deploy changes, or global installation. This
phase proves the wire contract without giving the binary authority over the
current workflow.

### Phase 1 — Workspace and Git resolver

Implement:

- explicit --workspace precedence, repeated explicit --repo selectors,
  auto, repo, and workspace modes;
- the marker search for .engram/config.json and _support/PROJECT.md, then
  nearest Git worktree, then canonical cwd workspace;
- absolute canonicalization, requested/canonical path reporting, final-target
  symlink resolution, no-follow discovery, containment, and unsafe-path errors;
- bounded sorted discovery, complete_discovery, and --all refusal with exit
  31 before any mutation;
- zero/one/many target handling, no_targets exit 13, and explicit verified
  repository behavior when discovery is incomplete;
- nested repository boundaries, linked worktree identity, bare-repository
  rejection, and Git common-directory derivation;
- the canonical-root workspace ID and exact platform state-root defaults.

Test with temporary directory trees and real Git repositories. Include symlink
aliases, loops, inaccessible components, symlink directory entries, paths with
spaces, nested repositories, linked worktrees, broad-root suppression, and
explicit repositories spanning roots without --workspace. Every resolver result
must expose applied limits, complete_discovery, nested boundaries,
requested/canonical paths, and explicit-versus-discovered selection.

### Phase 2 — Candidate projection and authority store

Implement:

- the exact worktree-v1 manifest: HEAD/base OIDs, staged paths with index
  stages/modes, tracked worktree content read without following symlinks,
  tracked symlink targets and missing markers, gitlink submodules, and
  non-ignored untracked regular files;
- raw-byte path ordering, tagged absent base OID, length-prefixed framing, and
  the domain-separated candidate SHA-256; ignored files, .git metadata,
  authority/workspace control paths, and untracked non-regular files must be
  excluded exactly as specified;
- separate domain-separated policy digest and policy parsing for named check and
  reviewer slots plus correction/validation budgets;
- opaque authority_id allocation and full repo/worktree/lineage binding,
  stable only for the lineage, collision detection, and no reuse after burn,
  abandon, or invalidation;
- per-repository authority under Git common dir, non-Git workspace coordination
  under the user-local state root, owner-only permissions, no unsafe symlinks,
  same-local-user scope, and workspace rename/new-ID behavior;
- atomic writes, locks, revision CAS, malformed-state quarantine, idempotent
  mutation replay, and no cross-repository atomic transaction;
- every lifecycle operation: start, prepare-result, capture-check,
  capture-review, capture-correction, capture-validation, recover, burn,
  inspect, and abandon;
- workspace review start, inspect, retry, and abandon keyed by
  workspace_run_id, with per-target outcomes and partial failure.

Test that all state actions have a command, that legal transitions reject
illegal states, and that burn removes mutable authority/candidate/intake/result
state while retaining only the minimum idempotency outcome. Verify that a
different key after burn returns state_absent, while an identical replay
returns the original burn envelope.

### Phase 3 — Direct CLI

Implement the commands and exact argv rules in the CLI contract:

- version, schema, and root resolve;
- workspace status;
- workspace review start, workspace inspect, workspace retry, and
  workspace abandon;
- review status, review start, and review prepare-result;
- review capture-check, review capture-review,
  review capture-correction, and review capture-validation;
- review recover, review burn, review inspect, and review abandon.

Require explicit canonical --repo PATH on repository-scoped commands and all
verification fields on bound commands. Require an idempotency key on every
mutation. Return one newline-delimited common envelope on stdout, diagnostics
only on stderr, stable error codes, and the prescribed exit status. Execute
transitions must return canonical absolute cwd and the canonical absolute
EvalSymlinks(os.Executable()) path as argv[0]. A collect transition must return named
evidence plus the exact review prepare-result argv; it must not return a
nonexistent or invented result path.

Run each command in temporary fixtures, validate every envelope against the
2020-12 schemas, and assert stdout contains exactly one JSON object. Verify
result-file no-follow/max 1 MiB checks, no-stdin rejection, idempotency replay
and key-reuse rejection, complete-discovery guard, workspace retry selection,
and all lifecycle/burn semantics.

### Phase 4 — Shadow adapter

Before changing an entry point, complete the verified rg inventory above and
map the selected surfaces to the CLI contract. Add a thin integration adapter
that invokes hive-control from existing review/verification entry points
without changing their authority. Shadow mode must:

- be opt-in and reversible;
- report scope, binding, proposed state, and next transition;
- preserve the common envelope and exact argv;
- avoid duplicate prompts or full source copies;
- never block, approve, commit, push, or deploy;
- record bounded comparison evidence outside the main prompt where possible.

The current Hive rules remain authoritative throughout this phase. No sibling
repo files are edited, and no generated/global installation is changed by this
proposal.

### Phase 5 — Shadow evaluation and A/B

Define comparable sessions across:

- repository-root cwd and subdirectory cwd;
- non-Git workspace with zero, one, and many repositories;
- explicit workspace, marker-selected workspace, nearest-Git fallback, and
  canonical-cwd fallback;
- nested parent/child repositories and linked worktrees;
- explicit symlink aliases, rejected symlink directory traversal, path escape,
  and broad-root suppression;
- --all with complete and incomplete discovery;
- candidate mutation during review and reviewer result for a different candidate;
- policy mutation, result-file violations, and idempotency key replay/reuse;
- correction-required to validating, escalated recovery, approved burn, and
  abandoned lineage;
- workspace partial failure, inspect, retry with completed targets, and abandon;
- compaction/resume, stale writer, process interruption, malformed state, and
  concurrent invocation.

Measure:

- context tokens attributable to repeated lifecycle instructions;
- number of model-reconstructed commands;
- transition disagreements between current rules and the CLI;
- stale-result, candidate, worktree, and policy mismatch detections;
- wrong-root and false-empty outcomes;
- time and filesystem cost of bounded discovery;
- recovery success after compaction;
- partial-failure reporting quality;
- publication/safety violations (the target is zero).

Use the proposed minimum repetitions and threshold set above. Keep the current
rules authoritative until the owner approves the complete evidence package.

### Phase 6 — Cutover

Cut over only after the A/B criteria are met and the owner approves the
replacement boundary. The cutover may retire duplicated review/verification
bookkeeping, including:

- manual status reconstruction;
- repeated candidate freshness prose;
- duplicated review slot counters;
- duplicated bounded-retry counters;
- workspace target aggregation.

The cutover must preserve:

- safety and authorization gates;
- model and human judgment;
- publication policy and separate commit/push/PR/merge authorization;
- Engram and memory policy;
- tracker semantics;
- Flow plans, ledgers, and durable artifacts;
- ordinary repository Git policy;
- security and secrets rules.

The old bookkeeping path should be frozen or removed at the cutover boundary,
not left as a second live authority. If rollback is needed, disable the CLI
adapter and return to the preserved current rules; do not run both authorities
against the same lineage.

### Phase 7 — Distribution and maintenance

Only after local verification and a separate explicit deployment decision:

- build the binary for supported platforms;
- install atomically into an approved user-local binary location;
- retain a rollback copy and version manifest;
- update deploy-global only if that responsibility is approved;
- add upgrade and uninstall behavior;
- document that the current Go binary may be installed without changing PATH.

Extending deploy-global or changing any global tool directory is a scope
change. A global deploy remains user-initiated.

## Verification matrix

The runtime needs more than a happy-path command test:

| Area | Required checks |
| --- | --- |
| Contract | 2020-12 schemas; one common envelope; root payload under data; full success/error/partial fixtures; unknown-field behavior; exit codes |
| Resolver | explicit workspace precedence; marker/Git/cwd fallback; zero/one/many repos; bounded depth/count; complete_discovery; sorted traversal; no symlink directory traversal; no_targets and exit 31 |
| Paths | requested/canonical path pair; null canonical path/transition before canonicalization; symlink loops/inaccessible paths; workspace containment; unsafe state-root/policy/result/authority paths; owner-only state |
| Git | repository root/subdirectory; .git file; linked worktrees; common-dir identity; bare repo; worktree replacement |
| Nested repos | independent parent/child targets; ordinary descendant discovery within bounds; no duplicate scan; no synthetic combined tree |
| Candidate | exact worktree-v1 manifest fixtures; HEAD/base binding; stages/modes; tracked content/symlink/missing; gitlinks; untracked regular files; ignored/control exclusions; byte ordering; SHA-256 vectors |
| Store | atomic writes; lock contention; CAS stale writer; idempotent replay/key reuse; malformed state; quarantine; burn cleanup and retained receipt |
| Binding | authority collision; candidate/policy mutation; worktree replacement; target mismatch; revision mismatch; reviewer result for another candidate |
| Lifecycle | command for every legal state/action; correction/validation budgets; escalated recover; approved burn; abandon; state-absent after burn |
| Evidence/policy | no stdin; existing no-follow regular result; 1 MiB limit; schema/slot/subject/policy digest; named slots; bounded budgets |
| Workspace | workspace_run_id; fan-out; all target outcomes; partial failure envelope/exit 50; retry only failed/non-terminal; completed targets skipped; no rollback claim |
| Security | argument-array execution; exact allowlist; no publication argv; path escape refusal; owner-only permissions; redaction; no secret persistence |
| CLI | exact prepare-to-capture argv/revision/path; canonical cwd and EvalSymlinks(os.Executable()) argv[0]; symlink-launched binary; stdout/stderr separation; --json; deterministic ordering |
| Resumption | status after interruption; compaction-equivalent fresh invocation; terminal cleanup; same-key burn replay |
| Packaging | build, local install/rollback, version output, no PATH mutation, supported OS matrix |

For the Go runtime, the affected verification boundary should include:

~~~bash
/usr/local/go/bin/go test ./...
/usr/local/go/bin/go test -race ./...
/usr/local/go/bin/go vet ./...
/usr/local/go/bin/go build ./cmd/hive-control
~~~

The commands are illustrative until a root go.mod exists. They must run from
the Hive root after implementation. Existing Markdown, generated-tree, and
shell checks remain required; adding a runtime does not remove them.

## Acceptance criteria

### AC-1 — Non-Git scope is first-class

Given a non-Git directory, when root resolve --mode workspace --json runs,
then the common envelope has kind=root.resolve, ok=true,
scope.kind=workspace, a canonical scope.canonical_path, and
data.complete_discovery with an explicit boolean value. It does not fail
merely because git rev-parse cannot resolve the workspace itself.

### AC-2 — Empty scope is explicit

Given a non-Git workspace with no repositories, root resolve returns
data.repositories=[], data.complete_discovery=true, and a stop transition
whose reason is no_targets with exit 0. A mutating workspace review start
returns ok=false, error.code=no_targets, and exit 13; no Git authority
directory is created.

### AC-3 — Root precedence is observable

Given --workspace W and a cwd inside repository R, resolution uses W and
returns scope.canonical_path=W; it does not silently narrow to R. Without
--workspace, auto reports the nearest .engram/config.json or
_support/PROJECT.md marker, otherwise the nearest Git worktree, otherwise the
canonical cwd workspace. It never crawls a broad home or volume root.

### AC-4 — Nested repositories are independent

Given a parent repository and an independently nested child repository, bounded
workspace discovery returns two targets with distinct repository/worktree
verification fields and boundary=nested. No combined repository identity or
child-content projection is emitted for the parent target.

### AC-5 — Discovery completeness guards fan-out

Given a discovery result with data.complete_discovery=false, a mutating
--all command returns error.code=discovery_incomplete, exit 31, and no
authority state revision changes. An explicitly canonicalized and Git-verified
--repo target may proceed independently.

### AC-6 — Path policy is fail-closed

Given an explicit symlink alias, the response records both
scope.requested_path and scope.canonical_path; a symlink loop or inaccessible
component returns error.code=symlink_loop or inaccessible_path, with
scope.canonical_path=null and next_transition=null, without mutation. A canonical
--repo outside explicit workspace containment returns error.code=unsafe_path
and exit 40. Discovery does not enter symlink directory entries.

### AC-7 — Evidence intake and exact argv are executable

Given a collect transition, next_transition.kind=collect includes
required_evidence and an exact allowlisted review prepare-result argv, but
contains no result-file path. When that argv runs with its returned canonical
cwd, it returns an existing owner-only absolute data.intake_path, increments
the revision, and returns the exact capture argv containing that path and new
expected revision. The adapter writes only to that path and executes only that
argv. Every returned execute argv has the canonical absolute
EvalSymlinks(os.Executable()) path as argv[0];
paths containing spaces, quotes, or shell metacharacters cannot alter the
operation.

### AC-8 — Candidate and reviewer freshness is mechanical

Given an accepted check or reviewer result bound to candidate digest X, policy
digest P, authority A, worktree W, and lineage L, a candidate, policy,
authority, worktree, or lineage change produces an observable binding_mismatch
or invalidated error (exit 22 where applicable) before state mutation. The
explicit capture-correction successor path is the only exception: it verifies
the old digest, recomputes the corrected digest, and returns a new authority.
A reviewer result for another candidate is never accepted as a warning.

### AC-9 — Stale writers fail closed

Given two callers with expected-revision 7, when caller A advances to
revision 8 first, caller B receives error.code=concurrent_update, exit 24,
and its state revision remains unchanged.

### AC-10 — Every lifecycle action has a command

Given each persisted state, next_transition names only an operation documented
in the CLI contract. correction_required accepts review capture-correction,
which verifies the old binding, marks it invalidated/superseded, and creates a
new authority in validating; escalated accepts review recover or review abandon;
approved accepts review burn or read-only status. review burn returns
data.disposition=burned, removes mutable evidence, replays only for the same
idempotency key, and returns state_absent for a different key.

### AC-11 — Evidence and policy contracts are bounded

Given a result submitted through stdin, --result-file -, a missing path,
symlink, non-regular file, or file over 1 MiB, the CLI returns
result_file_invalid (exit 30), or unsafe_path (exit 40) for a symlink/escape,
before mutation. A result missing the
contract, slot, authority binding, candidate digest, or policy digest returns
result_schema_invalid, result_slot_unknown, or binding_mismatch using
the exit mapping in the CLI contract. Policy data exposes named check/reviewer
slots and finite correction/validation budgets; exceeding a budget enters the
documented correction/escalation path rather than resetting it.

### AC-12 — Idempotency is observable

Given a mutation and idempotency key K, an identical replay returns the original
envelope and does not add a revision. Reusing K with a different payload,
binding, or operation returns error.code=idempotency_key_reuse and does not
mutate state. This remains true for the retained burn outcome.

### AC-13 — Workspace runs are resumable per target

Given workspace_run_id=WR-01 with completed, failed, and non-terminal targets,
workspace inspect returns all target outcomes. workspace retry returns
skipped_completed for completed targets and attempts only failed or
non-terminal selected targets. workspace abandon stops the run without
deleting per-repository outcomes and requires its idempotency key.

### AC-14 — Partial workspace failure uses the common envelope

Given selected repositories A and B where A completes and B fails, the response
has kind=workspace.result, ok=false, data.target_outcomes containing both
outcomes, error.code=partial_failure, and exit 50. It never reports the
workspace as approved or rolls back A.

### AC-15 — Execute authorization is allowlisted

Given any status, review, recovery, or workspace result, every execute argv
operation is one of the documented Hive lifecycle operations. No result
contains commit, push, PR, merge, release, or deploy argv. approved remains
an evidence state and does not authorize publication.

### AC-16 — Workspace identity and state are local

Given a canonical workspace root, workspace_id equals the full
domain-separated SHA-256 projection specified in architecture.md. Renaming
the root produces a new ID; old state is not adopted and remains available only
for explicit inspect or abandon. Default state roots match the macOS, Linux,
and Windows contract, existing state components are owner-only and non-symlink,
and a different local OS user is outside v1. Workspace migration is outside v1.

### AC-17 — Shadow is reversible

Given the recorded pre-shadow baseline fixture, when the same fixture runs with
shadow mode disabled, the observable state, stdout/stderr diagnostics, and
external side-effect log match the baseline; no control-plane authority is
written and no shadow transition is executed.

### AC-18 — Publication remains separate

Given an approved authority and a publication-authorization fixture returning
denied, requesting the next transition returns no publication argv, the process
observer records no commit, push, PR, merge, release, or deployment invocation,
and the lifecycle ends at evidence disposition or explicit burn.

## Operational rollback

Before cutover, rollback is disabling the shadow adapter; the current rules
remain authoritative. After cutover, rollback is:

1. disable the CLI adapter;
2. preserve the authority store for inspection;
3. return control to the preserved current procedure;
4. do not delete state or rewrite Git history;
5. record the reason and affected contract revision.

If the store is corrupt or identity cannot be verified, fail closed and require
an explicit inspection or abandonment path. Do not repair by guessing a root,
lineage, candidate, or revision.

## References

Official sources accessed on 2026-08-26:

- [Go module layout](https://go.dev/doc/modules/layout)
- [Go 1.27 release notes](https://go.dev/doc/go1.27)
- [Git rev-parse and common directories](https://git-scm.com/docs/git-rev-parse)
- [Kubernetes controller pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
- [JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12)

The optional reference project RDD comparison uses the local ../reference/optional reference project snapshot
7afe50d1 (v2.5.0-rc.1), especially
docs/review-integration.md and docs/architecture/organic-rdd.md. It is
reference material, not a dependency or an implementation source for Hive.
