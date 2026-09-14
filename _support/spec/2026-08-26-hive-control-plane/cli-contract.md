# CLI Contract

Status: Proposed

This document defines the proposed direct CLI and JSON contract for v1. The
contract is intentionally explicit about scope, exact argv, selectors, and
exit codes so adapters do not reconstruct workflow mechanics from prose.

## Executable and invocation

The executable is `hive-control`. It is a direct child process of a shell
adapter, hook, skill, or harness. There is no network listener, MCP transport,
or resident daemon in v1.

Machine callers must pass `--json`. A successful machine invocation writes one
JSON object followed by one newline to stdout. Human diagnostics and optional
progress go to stderr. A command must never mix human text into JSON stdout.

Every non-null `next_transition`, including `collect` and `stop`, includes a
canonical absolute `cwd`. An `execute` transition additionally includes an argv
array whose first element is the result of canonicalizing `os.Executable()` with
`filepath.EvalSymlinks`. Failure of either operation fails closed. Adapters must
not replace that path with a bare `hive-control` lookup or execute a command
through a shell.

The contract identifier is `tricell.hive-control/v1`. JSON Schemas use the
[JSON Schema Draft 2020-12](https://json-schema.org/draft/2020-12) dialect
(accessed 2026-08-26).

## Global options

These options are accepted at the top level or by the relevant subcommand; the
examples place them after the subcommand for readability:

| Option | Meaning |
| --- | --- |
| `--json` | Require the machine envelope; mandatory for automation |
| `--cwd PATH` | Requested process scope; defaults to the process current directory |
| `--state-root PATH` | Test/managed override for non-Git workspace state |
| `--contract NAME` | Contract name; defaults to `tricell.hive-control/v1` |
| `--quiet` | Suppress non-error stderr diagnostics; never changes JSON |
| `--help` | Human help; does not emit a success envelope unless `--json` is also given |
| `--version` | Human version; `version --json` is preferred for automation |

The executable must reject unknown options, duplicate singleton options, and
relative `--state-root` values. Paths in returned argv are absolute and
canonical where the resolver has authority to canonicalize them.

## Scope selectors

The following selectors are command-specific but share one meaning:

| Selector | Meaning |
| --- | --- |
| `--repo PATH` | Explicit repository target; repeatable |
| `--workspace PATH` | Explicit workspace root; establishes a workspace coordinator when combined with explicit `--repo` selectors |
| `--mode auto\|repo\|workspace` | Root resolution mode |
| `--all` | Select all repositories returned by bounded workspace discovery |
| `--path-prefix PATH` | Select discovered repository roots under an explicit workspace prefix |
| `--max-depth N` | Bounded discovery depth |
| `--max-repositories N` | Bounded number of repository targets |
| `--max-entries N` | Bounded directory entries visited |
| `--authority-id ID` | Opaque authority binding required by bound repository commands |
| `--repo-id ID` | Repository identity verification field |
| `--worktree-id ID` | Worktree identity verification field |
| `--lineage ID` | Lineage verification field |
| `--expected-revision N` | Compare-and-set revision required by mutations |
| `--idempotency-key KEY` | Required on every mutation; reusable only for an exact replay of the same request |

Mutating multi-repository commands require `--all`, at least one `--repo`, or
explicit `--repo-id` selectors. `--workspace` plus repeated `--repo`
selects those repositories in that workspace; it does not make the workspace a
synthetic repository. A status command may report all discovered targets without
selecting them for mutation.

Repository-scoped commands require an explicit absolute canonical
`--repo PATH` so the CLI can locate and validate Git's common directory.
Bound commands additionally require `--authority-id`,
`--repo-id`, `--worktree-id`, and `--lineage`. Mutating bound
commands also require `--expected-revision`; read-only `review inspect`
does not. Those verification fields must match the authority record; none is a
substitute for `authority_id`. Initial
`root resolve` and workspace commands can resolve a user-facing path,
then return canonical paths for subsequent commands.

Every mutation requires `--idempotency-key`. Keys are opaque, bounded,
and restricted to a safe filename alphabet. An identical request fingerprint
returns the original result, including after `burn`. Reusing a key with
different payload, binding, or operation returns `idempotency_key_reuse`
and does not mutate state.

## Commands

### Version and schema

~~~bash
hive-control version --json
hive-control schema status --json
~~~

`version` reports the binary version, contract version, Go toolchain metadata
used to build it, and supported capability names. `schema` reports the
contract identifier and schema resource names. Neither command reads or writes
repository authority.

### Root resolution

~~~bash
hive-control root resolve \
  --cwd /work/product \
  --mode workspace \
  --max-depth 3 \
  --max-repositories 32 \
  --max-entries 4096 \
  --json
~~~

This is read-only. It returns canonical workspace and repository identities,
discovery limits, nested boundaries, and unsupported targets.

### Workspace status

~~~bash
hive-control workspace status \
  --workspace /work/product \
  --mode workspace \
  --all \
  --json
~~~

This is read-only. It aggregates per-repository status without treating the
aggregate as one review authority.

### Workspace review start

Workspace mutation is explicitly namespaced under `workspace review`:

~~~bash
hive-control workspace review start \
  --workspace /work/product \
  --repo /work/product/services/api \
  --repo /work/product/apps/web \
  --policy-file /work/policy/hive-policy.json \
  --idempotency-key start-WR-01 \
  --json
~~~

The CLI canonicalizes the workspace and repository paths, verifies containment,
allocates `workspace_run_id`, and returns one independent target outcome
per selected repository. `--all` is accepted only when
`data.complete_discovery` is `true`. An incomplete discovery
returns exit 31 before any repository authority mutates. Explicit, individually
verified `--repo` targets may proceed even when unrelated discovery is
incomplete.

### Workspace inspection, retry, and abandon

~~~bash
hive-control workspace inspect \
  --workspace /work/product \
  --workspace-run-id WR-01 \
  --json

hive-control workspace retry \
  --workspace /work/product \
  --workspace-run-id WR-01 \
  --idempotency-key retry-WR-01-1 \
  --json

hive-control workspace abandon \
  --workspace /work/product \
  --workspace-run-id WR-01 \
  --idempotency-key abandon-WR-01-1 \
  --reason user_requested \
  --json
~~~

All three commands are keyed by `workspace_run_id`. `inspect` is
read-only. `retry` revalidates the stored selection and attempts only
failed or non-terminal selected targets; completed targets are returned as
`skipped_completed` and are never restarted. `abandon` stops the
workspace run without deleting per-repository outcomes. These commands do not
provide workspace-level approval or cross-repository rollback.

### Review status

Selectorless status preflights the current scope. A bound status includes the
canonical repository path, authority ID, lineage, expected revision, repository
ID, and worktree ID returned by a prior start. `--next-transition` asks
the CLI to return one legal transition.

~~~bash
hive-control review status \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 7 \
  --next-transition \
  --json
~~~

The caller must execute only the returned transition. A forecast or explanatory
text field is descriptive and is not an executable instruction.

### Review start

~~~bash
hive-control review start \
  --repo /work/product/services/api \
  --projection worktree \
  --policy-file /work/policy/hive-policy.json \
  --idempotency-key start-R-api-1 \
  --json
~~~

The policy file is read and hashed by the CLI. It is not copied into a command
string or emitted as secret-bearing content. `--policy-file` must resolve to an
absolute, existing, no-follow regular file within the caller's allowed policy
root; a symlink, escape, inaccessible component, or non-regular file is
rejected before authority mutation.

### Prepare and capture evidence

The v1 contract does not support stdin. A `collect` transition names the
required evidence slot and returns an exact allowlisted
`review prepare-result` argv. That command creates an intake file,
returns its canonical absolute path in `data.intake_path`, increments the
authority revision, and returns an `execute` transition containing the exact
capture argv with that path and new revision. The adapter writes the bounded
result only to the allocated path, then executes only that returned capture
argv. A `collect` transition never invents a result path.

~~~bash
hive-control review prepare-result \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 7 \
  --slot-id affected-tests \
  --idempotency-key prepare-R-api-1 \
  --json
~~~

The intake path is created as an owner-only regular file below the authority
lineage. It is not a symlink, is bounded to 1 MiB, and is returned only after
the file has been allocated successfully. `--result-file` on a capture
command must be an absolute, existing, no-follow regular file no larger than
1 MiB. `--result-file -` is invalid and returns exit 30.

For the prepare request above, a successful response advances the authority to
revision 8 and its `next_transition` is an exact `execute` for
`review capture-check`, including `--expected-revision 8`,
`--result-file <data.intake_path>`, the binding fields, slot ID,
idempotency key, canonical `cwd`, and canonical executable path. Schemas and
fixtures must assert that the prepare response and capture argv agree byte for
byte on those values.

### Capture a check

~~~bash
hive-control review capture-check \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 8 \
  --check-id affected-tests \
  --result-file /work/product/services/api/.git/tricell-hive/control/v1/intake/I-01.json \
  --idempotency-key capture-check-R-api-1 \
  --json
~~~

### Capture a reviewer result

~~~bash
hive-control review capture-review \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 8 \
  --reviewer-id code-reviewer \
  --result-file /absolute/path/review-result.json \
  --idempotency-key capture-review-R-api-1 \
  --json
~~~

The CLI validates the result envelope, schema, subject/candidate digest, policy
digest, authority binding, reviewer or check slot, and lineage/revision before
accepting it. A result for a different candidate is a binding mismatch, not a
warning. The CLI records structured evidence and derives mechanical state; it
does not decide whether a human-authored finding is true. The result's declared
mechanical outcome may leave the authority `reviewing`, enter
`correction_required`, or enter `escalated`; a correction-required outcome is
handled only by the capture-correction command below.

### Capture a correction

~~~bash
hive-control review capture-correction \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 10 \
  --result-file /absolute/path/correction-result.json \
  --idempotency-key capture-correction-R-api-1 \
  --json
~~~

The correction result must identify the old authority binding and source
candidate manifest digest. After the adapter applies the correction, the CLI
recomputes the worktree-v1 manifest and requires the result's
`corrected_candidate_manifest_sha256` to match. It consumes the policy's
bounded correction budget, marks the old lineage invalidated/superseded, and
returns a new `authority_id` and `lineage` in `data` with the successor in
`validating`; it never edits the old binding or silently resets the budget.

### Capture validation

~~~bash
hive-control review capture-validation \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 11 \
  --result-file /absolute/path/validation-result.json \
  --idempotency-key capture-validation-R-api-1 \
  --json
~~~

Validation results must identify the candidate and policy they observed. The
policy controls the validation budget; a result for a different candidate,
authority, or policy is rejected before state mutation.

### Recover an escalated review

~~~bash
hive-control review recover \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 12 \
  --recovery-reason maintainer_authorized \
  --idempotency-key recover-R-api-1 \
  --json
~~~

`recover` is legal only from `escalated` and requires an explicit
reason. It revalidates the repository and candidate binding. If a new target or
projection is needed, the result contains a new `authority_id` and
lineage; the old authority is not weakened or reused. Recovery never returns a
publication command.

### Burn an approved authority

~~~bash
hive-control review burn \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 13 \
  --idempotency-key burn-R-api-1 \
  --json
~~~

`burn` is legal only from a verified `approved` state. While
holding the lineage lock, it verifies the binding and revision, persists the
minimum idempotency outcome, removes mutable state, candidate snapshots,
intakes, checks, and reviewer results, and returns
`data.disposition=burned`. A same-key identical replay returns the
original envelope; a different key returns `state_absent`. Burn never
authorizes commit, push, PR, merge, release, or deploy.

### Abandon and inspect

~~~bash
hive-control review inspect \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --json

hive-control review abandon \
  --repo /work/product/services/api \
  --authority-id A-01 \
  --repo-id R-api \
  --worktree-id W-api \
  --lineage L-01 \
  --expected-revision 13 \
  --reason user_requested \
  --idempotency-key abandon-R-api-1 \
  --json
~~~

`inspect` is read-only but a repository-scoped inspect still requires
the canonical `--repo` and binding fields. `abandon` is an explicit
lifecycle mutation; it never deletes an unrelated lineage. It preserves a
minimal inspectable outcome and idempotency record. Quarantine or cleanup
policy belongs to the authority implementation.

Bound repository commands have the following mandatory verification fields:

~~~text
--repo <canonical absolute path>
--authority-id <opaque authority id>
--repo-id <verification id>
--worktree-id <verification id>
--lineage <verification id>
~~~

Every bound mutation additionally has:

~~~text
--expected-revision <current revision>
--idempotency-key <unique request key>
~~~

The CLI validates all fields against the authority record. The authority ID is
the lookup binding; the other IDs prevent a stale or misrouted result from
being accepted.

## Common machine envelope

Every response, including failures and partial workspace results, has the same
top-level fields. `next_transition` is nullable only when the requested path or
executable cannot be canonicalized:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "schema": "tricell.hive-control/v1",
  "kind": "review.status",
  "ok": true,
  "request_id": "req_01J...",
  "observed_at": "2026-08-26T18:00:00Z",
  "scope": {
    "kind": "repository",
    "requested_path": "/work/product/services/api",
    "canonical_path": "/work/product/services/api"
  },
  "data": {},
  "next_transition": {
    "kind": "collect",
    "operation": "review.capture-review",
    "cwd": "/work/product/services/api",
    "slot_id": "code-reviewer",
    "required_evidence": [
      "candidate_manifest_sha256",
      "policy_sha256",
      "authority_id",
      "review_findings"
    ],
    "prepare": {
      "operation": "review.prepare-result",
      "cwd": "/work/product/services/api",
      "argv": [
        "/usr/local/bin/hive-control",
        "review",
        "prepare-result",
        "--repo",
        "/work/product/services/api",
        "--authority-id",
        "A-01",
        "--repo-id",
        "R-api",
        "--worktree-id",
        "W-api",
        "--lineage",
        "L-01",
        "--expected-revision",
        "7",
        "--slot-id",
        "code-reviewer",
        "--idempotency-key",
        "prepare-review-R-api-1",
        "--json"
      ]
    }
  },
  "warnings": []
}
~~~

The `cwd` field is the canonical absolute directory in which the returned
argv must run. The first argv element is
`filepath.EvalSymlinks(os.Executable())` for the binary that produced the
envelope. The example's
`/usr/local/bin/hive-control` is illustrative; an implementation must
return its actual canonical executable path.

`observed_at` and `request_id` are diagnostic metadata. State decisions use
canonical identity, revision, hashes, and transition fields, not wall-clock
ordering. Conformance fixtures should normalize diagnostic metadata when
comparing deterministic output.

Transition kinds are:

| Kind | Meaning |
| --- | --- |
| `execute` | The caller may execute the exact returned argv |
| `collect` | The caller must provide named evidence through the exact capture contract |
| `stop` | No automatic lifecycle operation is legal |

The returned argv is an array, not a shell string. Adapters must not reorder,
drop, quote, or add arguments. If an argv cannot be represented safely, the CLI
must return `stop` with a reason code.

The only allowed `execute` operations are the Hive lifecycle operations
listed in [Adapter rules](#adapter-rules). The CLI must never return an argv for
commit, push, PR, merge, release, deploy, or an arbitrary policy command.

### Error and partial-result variants

An ordinary error uses the same envelope and adds `error`:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "schema": "tricell.hive-control/v1",
  "kind": "review.capture-review",
  "ok": false,
  "request_id": "req_01J...",
  "observed_at": "2026-08-26T18:00:00Z",
  "scope": {
    "kind": "repository",
    "requested_path": "/work/product/services/api",
    "canonical_path": "/work/product/services/api"
  },
  "data": {},
  "next_transition": {
    "kind": "stop",
    "reason": "concurrent_update",
    "cwd": "/work/product/services/api"
  },
  "warnings": [],
  "error": {
    "code": "concurrent_update",
    "message": "state revision 8 does not match expected revision 7",
    "retryable": true,
    "details": {
      "repo_id": "R-api",
      "lineage": "L-01",
      "current_revision": 8
    }
  }
}
~~~

Workspace partial failure also uses the common envelope. It returns exit 50:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "schema": "tricell.hive-control/v1",
  "kind": "workspace.result",
  "ok": false,
  "request_id": "req_01J...",
  "observed_at": "2026-08-26T18:00:00Z",
  "scope": {
    "kind": "workspace",
    "requested_path": "/work/product",
    "canonical_path": "/work/product"
  },
  "data": {
    "workspace_run_id": "WR-01",
    "complete_discovery": true,
    "target_outcomes": [
      {"repo_id": "R-api", "status": "completed", "authority_id": "A-01"},
      {"repo_id": "R-web", "status": "failed", "error_code": "binding_mismatch"}
    ]
  },
  "next_transition": {
    "kind": "stop",
    "reason": "partial_failure",
    "cwd": "/work/product"
  },
  "warnings": [],
  "error": {
    "code": "partial_failure",
    "message": "one or more selected repository targets did not complete",
    "retryable": true,
    "details": {"failed_repo_ids": ["R-web"]}
  }
}
~~~

Error messages are for humans and logs; callers branch on `error.code`.
Sensitive path components, environment values, and candidate content must not
be included unless the caller explicitly requested a safe diagnostic view.

A failure before canonicalization retains the requested path but uses null for
the unavailable values:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "schema": "tricell.hive-control/v1",
  "kind": "root.resolve",
  "ok": false,
  "request_id": "req_01J...",
  "observed_at": "2026-08-26T18:00:00Z",
  "scope": {
    "kind": "unknown",
    "requested_path": "/missing/workspace",
    "canonical_path": null
  },
  "data": {},
  "next_transition": null,
  "warnings": [],
  "error": {
    "code": "path_missing",
    "message": "requested path does not exist",
    "retryable": false,
    "details": {}
  }
}
~~~

## Root resolution envelope

The `root resolve` result places its payload under `data` like every
other command:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "schema": "tricell.hive-control/v1",
  "kind": "root.resolve",
  "ok": true,
  "request_id": "req_01J...",
  "observed_at": "2026-08-26T18:00:00Z",
  "scope": {
    "kind": "workspace",
    "requested_path": "/work/product",
    "canonical_path": "/work/product"
  },
  "data": {
    "mode": "workspace",
    "complete_discovery": true,
    "limits": {
      "max_depth": 3,
      "max_repositories": 32,
      "max_entries": 4096,
      "visited_entries": 42,
      "limits_reached": []
    },
    "repositories": [
      {
        "repo_id": "R-api",
        "worktree_id": "W-api",
        "root": "/work/product/services/api",
        "git_dir": "/work/product/services/api/.git",
        "git_common_dir": "/work/product/services/api/.git",
        "boundary": "nested",
        "selection": "discovered",
        "capabilities": ["worktree_projection"]
      }
    ],
    "unsupported_targets": [],
    "nested_boundaries": ["/work/product/services/api"]
  },
  "next_transition": {
    "kind": "stop",
    "reason": "resolution_complete",
    "cwd": "/work/product"
  },
  "warnings": []
}
~~~

The envelope's `scope` describes the request; `data` describes the
resolved result. `complete_discovery` is therefore machine-readable and
must be checked before a mutating `--all` operation.

Identity fields are opaque contract values. Callers must not reconstruct them
from path strings. The authority binding rules are:

~~~text
authority_id = opaque unique lineage binding
repo_id      = repository identity verification field
worktree_id  = worktree identity verification field
lineage      = lifecycle lineage verification field
~~~

`authority_id` is allocated at start, remains stable for the lineage, and
is never reused after burn, abandonment, or invalidation. A collision or an
existing ID with a different binding returns `authority_id_collision`
without mutation. Bound repository commands and result payloads require the
canonical `--repo PATH`, `authority_id`, `repo-id`,
`worktree-id`, `lineage`, and expected revision.

Result files have one supported input mechanism in v1:

- no stdin support;
- `--result-file` must be absolute, existing, no-follow, regular, and
  no larger than 1 MiB;
- a symlink, directory, device, FIFO, socket, missing path, oversized file, or
  path that escapes its expected intake root is rejected before mutation;
- the CLI computes a domain-separated SHA-256 of the exact input bytes and
  stores only the bounded structured result and digest required by the contract;
- the result must carry the contract, slot ID, `authority_id`,
  `repo-id`, `worktree-id`, `lineage`,
  candidate manifest digest, and policy digest.
- a correction result additionally carries
  `corrected_candidate_manifest_sha256`; its ordinary candidate digest remains
  the source digest bound to the old authority.

The result digest is exactly
`SHA256("tricell.hive-control/result/v1\0" || result_file_bytes)`.
The submitted result schema must expose that digest (or the CLI must add it to
the accepted evidence record), and the result's candidate and policy digests
must equal the active authority binding. A minimum policy shape is:

~~~json
{
  "schema": "tricell.hive-control/policy/v1",
  "checks": [{"slot_id": "affected-tests"}],
  "reviewers": [{"slot_id": "code-reviewer"}],
  "budgets": {"max_corrections": 2, "max_validations": 1}
}
~~~

Slot IDs are unique within their list and budgets are finite non-negative
integers. The policy digest is calculated over the exact policy file bytes;
unknown policy fields are rejected unless the versioned policy schema marks
them non-semantic.

The policy file supplies named check and reviewer slots and bounded
correction/validation budgets. It is read as an existing safe regular file,
hashed with the policy domain separator, and never treated as a shell command
source. Every mutation, including prepare, capture, recover, burn, abandon,
workspace review start, workspace retry, and workspace abandon, requires a
unique `--idempotency-key`. Identical requests replay the original
envelope; key reuse with a different payload fails with
`idempotency_key_reuse`.

## Exit codes

Exit codes are stable coarse classes; the JSON error code is the precise
branching value.

| Code | Class | Examples |
| ---: | --- | --- |
| 0 | Success | status, empty workspace resolution, accepted capture |
| 2 | Usage | missing selector, conflicting flags, invalid JSON mode |
| 10 | Resolution | path missing, Git identity unavailable |
| 11 | Ambiguous scope | multiple repositories where one was required |
| 12 | Unsupported target | bare repository, unsupported projection |
| 13 | No targets | mutating command has no selected verified repositories |
| 20 | State absent | lineage not found, no active authority |
| 21 | Invalid state | malformed, unsupported, or quarantined state |
| 22 | Binding mismatch | candidate, policy, worktree, or target changed |
| 23 | Illegal transition | operation not legal from current state |
| 24 | Concurrent update | stale revision or lock conflict |
| 25 | Authority collision | opaque authority ID is bound to a different tuple |
| 30 | Input rejected | result schema, subject hash, or slot mismatch |
| 31 | Discovery limit | bounded walk stopped before complete discovery |
| 40 | Permission/security | unsafe path, bad permissions, symlink escape |
| 50 | Partial workspace failure | one or more selected repositories failed |
| 70 | Internal failure | invariant or unexpected implementation error |

The canonical v1 error-code registry is:

| Error code | Exit | Meaning |
| --- | ---: | --- |
| `path_missing` | 10 | Requested path does not exist |
| `symlink_loop` | 10 | Explicit path canonicalization found a link loop |
| `inaccessible_path` | 10 | An explicit path component cannot be inspected |
| `ambiguous_scope` | 11 | One target was required but multiple remain |
| `unsupported_target` | 12 | Target or projection is outside v1 |
| `no_targets` | 13 | Mutation has no selected verified repository |
| `state_absent` | 20 | Requested authority or run does not exist |
| `state_invalid` | 21 | State is malformed, unsupported, or quarantined |
| `binding_mismatch` | 22 | Candidate, policy, repository, worktree, lineage, or result binding differs |
| `illegal_transition` | 23 | Operation is not legal from current state |
| `concurrent_update` | 24 | Expected revision or lock precondition failed |
| `authority_id_collision` | 25 | Opaque ID is already bound elsewhere |
| `result_file_invalid` | 30 | Result is missing, oversized, stdin, or not a regular file |
| `result_schema_invalid` | 30 | Evidence does not match the versioned schema |
| `result_slot_unknown` | 30 | Check or reviewer slot is absent from policy |
| `idempotency_key_reuse` | 30 | A key was reused with a different request fingerprint |
| `discovery_incomplete` | 31 | `--all` cannot prove a complete target set |
| `unsafe_path` | 40 | A path is a forbidden symlink, escape, ownership, or permission case |
| `partial_failure` | 50 | At least one selected workspace target failed |
| `internal_failure` | 70 | An invariant or unexpected implementation error occurred |

A result digest mismatch uses `binding_mismatch`, not a special warning. These
codes are the only strings adapters branch on; schemas and fixtures must reject
an undocumented code.

An empty workspace resolution is successful because it is a verified observation.
A mutating command with no targets returns a typed error even though the
underlying read-only resolution succeeded.

## Versioning and compatibility

The contract major version changes when field meaning, transition semantics,
identity binding, or exit-code behavior becomes incompatible. Additive fields
must be optional and documented. Unknown fields may be ignored only when the
schema declares them non-semantic; unknown transition kinds must fail closed.

The JSON Schemas and fixtures under
`contracts/hive-control/v1/{schemas,fixtures}` are part of the compatibility
surface. The first implementation must test every envelope and exit-code class
against those fixtures.

## Adapter rules

An adapter may:

- invoke the binary with an explicit cwd;
- pass exact selectors;
- relay JSON losslessly to the model or harness;
- execute the exact returned argv;
- provide one named result through a capture command.

An adapter may not:

- parse the repository root or authority path from prose;
- manufacture a lineage, revision, target ID, or subject hash;
- change argv ordering or values;
- convert `stop` into a guessed retry;
- treat `approved` as publication authorization;
- use the workspace root as a repository fallback.
