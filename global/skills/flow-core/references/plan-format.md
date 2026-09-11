# Plan format — the executable, harness-agnostic plan

A Hive plan is a human-readable Markdown contract plus mutable approval and
execution metadata. The contract is the part another harness executes; its
exact bytes are bound to the approval by SHA-256. The metadata lets a plan
advance, pause, resume, or record delivery without changing the approved
instructions.

## Layout and ownership

The plan has five independent areas:

1. `Status:` is one mutable line outside the contract. It is one of `draft`,
   `planned`, `building`, `built`, or `verified`.
2. The contract is ordinary Markdown between the contract markers. It contains
   the goal, scope, decisions, interfaces, tasks, and verification recipe.
3. Authorization is one fenced JSON block outside the contract. It binds the
   contract digest to user evidence and exact action/target grants.
4. The execution table is ordinary Markdown outside the contract. It records
   task progress and delivery evidence; edits to it do not change the digest.
5. Recovery is an optional fenced JSON block outside the contract and
   authorization. It records meaningful attempts so a resumed run can
   reconcile an interrupted operation before repeating it.

The contract markers are exact and must occur once each:

```markdown
<!-- hive-plan:contract:start -->
<!-- hive-plan:contract:end -->
```

The authorization markers are exact and must occur once each:

```markdown
<!-- hive-plan:authorization:start -->
<!-- hive-plan:authorization:end -->
```

The recovery markers are optional, but when present they are exact and must
occur once each:

```markdown
<!-- hive-plan:recovery:start -->
<!-- hive-plan:recovery:end -->
```

Recovery metadata must stay outside both the contract and authorization
blocks. A plan created before recovery support has no recovery history; that
absence is reported as `unknown` and is never treated as zero attempts.

The bytes hashed are every UTF-8 byte after the contract start marker and
before the contract end marker, including their leading and trailing
whitespace. Status, authorization, and execution text are outside that byte
range.

The contract stays Markdown. Do not replace it with a JSON schema or put
machine metadata inside it. The validator checks only that the contract is
valid UTF-8, non-empty, and contains a Markdown heading; the flow and review
stages own its domain-specific completeness.

## Canonical shape

````markdown
# Feature — Plan

Status: planned

<!-- hive-plan:contract:start -->
## Contract

Implements: <epic, issue, or decision reference>
Spec: <path to the design spec, if any>
Goal: <one line>
Scope: <what changes>
Exclusions: <what remains outside this plan>
Architecture: <the selected approach and its boundaries>
Integration: <branch, merge mechanics, and applicable CI gates>

## Decisions to close before executing

| Decision | Resolution | Type | Blocks |
|---|---|---|---|
| D1 | <resolved choice> | technical or stakeholder | T1 |

### T1: <imperative title>

in-vivo: yes | no
design-review: yes
Agent: <agent name, when a routing row applies>
Files:
  - Create: <exact path>
  - Modify: <exact path>
Interfaces:
  - Consumes: <known input>
  - Produces: <output consumed later>

- Step 1: <one action>
- Step 2: <one action>
Verify: `<command>` — Expected: `<result>`
Commit: feat(<scope>): T1 <subject>
<!-- hive-plan:contract:end -->

## Execution

| Task | State | Evidence |
|---|---|---|
| T1 | pending | |

<!-- hive-plan:recovery:start -->
```json
{
  "schema": "hive-plan/recovery.v1",
  "plan_id": "feature-slug",
  "attempts": []
}
```
<!-- hive-plan:recovery:end -->

<!-- hive-plan:authorization:start -->
```json
{
  "schema": "hive-plan/authorization.v1",
  "plan_id": "feature-slug",
  "contract_sha256": "<64 lowercase hexadecimal characters>",
  "approval": {
    "evidence": "The user's explicit authorization or an exact reference to it.",
    "date": "2026-09-10",
    "revision": 1
  },
  "grants": [
    {
      "id": "grant-implement",
      "action": "implement",
      "targets": ["repo:feature-slug"],
      "conditions": [],
      "evidence": "The user's explicit instruction authorizing implementation."
    }
  ],
  "revocations": []
}
```
<!-- hive-plan:authorization:end -->
````

`plan.py digest <plan>` computes the digest after the contract has been
written. The resulting value is copied into `authorization.contract_sha256`.
The authorization block is then updated without touching the contract.

## Authorization metadata

The authorization object is intentionally narrow:

- `schema` must be `hive-plan/authorization.v1`.
- `plan_id` identifies the plan instance and is reported during inspection.
- `contract_sha256` must equal the digest of the exact current contract bytes.
- `approval.evidence` records the user evidence. It documents provenance; it
  is not a cryptographic signature or proof of human identity.
- `approval.date` is an ISO-like local date or timestamp, and `revision` is a
  positive integer incremented when the authorization is replaced.
- The `grants` list may be empty for an approved design-only plan. Each grant
  has a unique `id`, one lowercase action name, one or more exact `targets`, a
  `conditions` list, and non-empty `evidence` binding that grant to the user
  instruction. Targets are compared by exact equality;
  wildcard characters are rejected and no implicit path, branch, or resource
  inheritance exists.
- A revocation names an existing grant by `grant_id`, records a non-empty
  `reason`, `evidence`, and `revoked_at`, and takes effect for the current authorization.
  To grant the same action again, add a new grant with a new ID and a new
  authorization revision.

Each condition is an object with a non-empty `requirement` and an `evidence`
string. Empty evidence means the condition is pending and validation fails
closed. Non-empty evidence records that the condition was satisfied; the flow
still owns freshness and independent verification. A satisfied condition does
not require asking for the same user approval again. This small validator does
not attest tests, deployments, or human consent.

`validate <plan> --action <action> --target <target>` is the authority check
for one concrete action. It succeeds only when:

1. the document, authorization metadata and optional recovery metadata are well-formed;
2. the plan is not `draft`;
3. the authorization digest matches the current contract;
4. the exact action and target occur in an active grant; and
5. that grant has no unresolved conditions.

Action/state compatibility is closed by the validator: `implement` is valid
only in `planned` or `building`; `verify` is valid in `built` or `verified`;
`commit`, `push`, `pr`, `merge`, and `deploy` are valid only in `verified`.
Unknown actions and actions at another status fail closed. This keeps a
publication grant from bypassing verification and prevents `verified` from
silently reopening implementation.

An action absent from all grants is `undeclared_operation`. An action present
with a different target is `undeclared_destination`. A matching grant covered
by a revocation is `grant_revoked`. These are separate machine-readable error
codes in the CLI result.

## Recovery metadata

The recovery block is mutable operational evidence. It is not part of the
contract digest, does not authenticate the user, and does not grant or revoke
any action. The coordinator is the only writer; delegated workers return
evidence to the coordinator instead of editing the plan concurrently.

`recovery.plan_id` must match the bound authorization's `plan_id`; this binds
the history to the same plan instance without claiming that the history is
tamper-proof.

`attempts` is an ordered list. Each entry has a unique `id`, strictly
increasing positive `sequence`, one canonical `kind` (`task`, `delegation`,
`fix`, `review`, or `remote`), a stable semantic `scope`, the
`contract_sha256` observed when the attempt began, `started_at`, `outcome`, and
non-empty `evidence`. A completed entry also has `ended_at`. `started` means
the attempt is still open; `unknown` and `interrupted` mean its result must be
reconciled before another attempt. Optional `task` and `predecessor` fields
connect an entry to a plan task or an earlier attempt. A `predecessor` must
refer to an earlier entry.

The same semantic `kind` and `scope` retain their counters across contract
digests. A different scope has an independent budget. The validator does not
infer that a changed digest is a new budget or that an absent history is an
empty budget; the reconciler decides that from observed evidence. A remote
failed attempt may carry one `exception` with type `verified_transient` and
its evidence. It is excluded from the remote failed-operation count only as
that narrow exception.

The bounded budgets are:

| Kind | Budget | Counted attempts |
|---|---:|---|
| `task` | none | ordinary task state; no retry budget is inferred |
| `delegation` | 2 total | initial dispatch plus one rerun |
| `fix` | 3 | failed attempts for the same problem scope |
| `review` | 2 total | correction rounds for the same review scope, including a successful closure |
| `remote` | 2 | failed operations, excluding one verified transient exception |

These limits do not authorize a run. They only answer whether the recorded
scope can continue. An open or ambiguous attempt requires reconciliation of
the workspace, Git, process, or remote state before repeating it. A successful
`review` attempt returns `completed` and closes that review kind/scope
permanently, including across later contract digests; use a new semantic scope
for an independent review. Successful task, delegation, fix and remote
attempts remain evidence while their scope may continue according to its own
budget. A completed result remains evidence even when a later delivery was
interrupted.

## Status and execution state

The status line is outside the frozen contract and can change without
changing its digest:

```text
draft ──approval──> planned ──start──> building ──tasks done──> built ──gate──> verified
```

| Status | Meaning |
|---|---|
| `draft` | Contract is being prepared; execution is not authorized. |
| `planned` | Contract and authorization are bound; implementation has not started. |
| `building` | An authorized implementation is in progress. |
| `built` | Planned implementation tasks have landed and their task checks ran. |
| `verified` | Applicable verification and review gates passed. Delivery may still be pending. |

The execution table is a mutable report, not an authorization source. It can
record `pending`, `in_progress`, `done`, or `blocked` per task, plus evidence
for commits, tests, review, and delivery. The reconciler compares those
claims with the workspace, Git, and executed checks before advancing status.
Changing a progress row must leave `contract_sha256` unchanged.

`verified` means quality gates passed. It does not grant commit, push, merge,
release, or deploy permission. Those operations require corresponding active
grants and the session's existing Git and deployment gates. A plan may carry
those grants from the initial approval, so one explicit authorization can
cover implementation and publication when the user clearly requested both.

## Task block

Every task is independently executable and independently trackable. The
contract must be self-contained for an engineer on another harness:

- `in-vivo:` is mandatory. Set `yes` when the task needs a live walk; otherwise
  set `no`.
- `design-review: yes` opts a user-facing UI task into the visual-craft gate.
- `Agent:` is an optional routing annotation. A harness without that roster
  executes the recipe directly and reports the substitution.
- `Verify:` pairs a command with its expected output. A step without an
  observable expected result is not a verification step.
- `Commit:` carries the local task tag (`T1`, `T2`, and so on). It describes
  the intended commit; the execution table and Git remain the evidence that it
  happened.

When a decision blocks task detail, resolve it in `Decisions to close before
executing` and include the resolution in the approved contract. Do not leave
`TBD` decisions for the implementation stage.

## Proportionality and delivery

Small work may combine definition and planning in one short contract. Larger
work can include separate specification, design, and task sections or split
into parts. The approval boundary remains the contract digest and its active
grants; adding a file or editing the execution table does not silently expand
the approved scope.

An approval can grant implementation and publication together. A material change to the
objective, contract scope, interfaces, tasks, or verification requires a new contract revision
and new user approval. Expanding Authorization to an operation or target already covered by the
frozen contract records new user evidence and increments the authorization revision without
changing the contract digest. Narrowing or revoking a grant updates Authorization or revocations
without a contract revision; an operation or target outside the frozen contract scope requires
one. A positive review is evidence of quality, not a new grant.

When adopting a legacy or organic plan, normalize it before treating it as executable. A legacy
status, location or native harness approval supplies no authority. If the normalized contract is
unchanged and the current user request explicitly covers implementation, record that request as
the `implement` grant evidence; otherwise clarify the exact scope before editing.

## CLI

The standard-library validator is read-only and emits JSON. It never writes a
plan, changes status, updates progress, or authenticates the human evidence.

```sh
python3 path/to/plan.py inspect path/to/plan.md
python3 path/to/plan.py digest path/to/plan.md
python3 path/to/plan.py validate path/to/plan.md \
  --action implement --target repo:feature-slug
python3 path/to/plan.py recovery-check path/to/plan.md \
  --kind delegation --scope T1
```

Successful commands exit zero. Rejected or malformed plans emit an `error`
object with a stable `code` and exit non-zero. `inspect` distinguishes a
well-formed bound authorization from an `unbound` digest, while `can_implement`
is true only for a non-draft plan with an active, condition-satisfied
`implement` grant. A design-only plan may be bound and inspectable while
returning `can_implement: false`.

`inspect` includes a recovery summary. `status: known` means the block is
present, including an explicitly initialized empty `attempts` list;
`status: unknown` means the plan predates recovery metadata and must be
reconciled before repeating work. `recovery-check` is read-only and returns a
structured `continue`, `reconcile`, `completed`, or `exhausted` decision. It
exits zero only for `continue`; unresolved, completed, or exhausted recovery
exits non-zero. It never checks or creates action grants, so callers still run
`validate` for an authorized implementation or delivery action.

Verification uses `inspect`, not `validate --action verify`: it requires the current digest,
authorization status `valid` or `conditional`, and observed status `built` or `verified`. An
explicit `verify` request or an explicit normal `/flow-build` invocation reaching the Gate is
the verification authority; no `implement` grant is required. `validate` is reserved for an
action and target that require an active grant, including implementation and delivery.
