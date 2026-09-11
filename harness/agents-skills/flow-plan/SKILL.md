---
name: flow-plan
description: >
  Prepare and approve a portable Hive plan for a bounded unit of work. Use when the user
  explicitly invokes `/flow-plan` or clearly asks to plan an objective; the plan owns phases,
  artifacts and authorization independently of the harness's native plan mode.
disable-model-invocation: true
---

# /flow-plan — prepare the plan

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This is the planning half of
the daily chain. It may inspect the workspace and write the portable plan and, when a ledger
exists, its handoff, but it never edits project code itself. It does not commit, push, merge,
publish or deploy. `/flow-build` is user-gated (`disable-model-invocation`): never invoke it from
this skill, even when the approval also authorizes implementation and delivery — complete the
approval record, then ask the user to run `/flow-build` (or take the direct route if they choose it).

The user must invoke `/flow-plan` or make an equivalent explicit request such as “planeemos X”.
When a harness does not expose the slash command, that request follows this same procedure
directly. Do not infer that a question, a native harness Plan Mode transition, or a proposed
approach is a request to create or approve a Hive plan. Native Plan Mode remains an optional
drafting aid; it never changes the Hive plan, authorization or phase state.

## OPEN — resolve the scope

1. Read `~/.agents/skills/flow-core/references/plan-format.md` before creating or normalizing a
   plan. Use its exact contract delimiters, metadata fields, state transitions and read-only
   validator CLI; do not invent a parallel schema. The same reference is available from the
   universal skill root on non-Claude harnesses.
2. If `<project>/_support/PROJECT.md` exists, read it and consume `## Current handoff` as
   required by the flow contract. If it is absent, use the standalone repository's natural
   `_support/sessions/` home and require an explicit plan path or an unambiguous objective; do
   not bootstrap a ledger merely to plan this work.
3. Resolve the argument as an objective, an existing portable plan, or a session plan path.
   A single unambiguous candidate may be continued; if several active candidates exist, list all
   of them and require the user to choose, even when one is newer. A plan with `Session: no` may
   be discussed in the current conversation, but it carries no durable cross-harness or resume
   guarantee.
4. Run only the bounded read-only exploration needed to identify the affected files, repos,
   contracts, prerequisites and applicable specialists. Preserve the distinction between
   facts, decisions and open questions.
5. **Offer research when the exploration leaves a question only official docs or the web can
   settle** (a library capability, a standard, an ecosystem claim): ONE prose offer of an
   evidence pass — `sdd-explore` in `evidence` mode — before tasks are written. Offered, never
   auto-run; a no holds for this plan. Accepted: `resolved` claims enter the contract's
   decisions with their citations; a `partial` outcome is an open question at the gate, never
   a decision (`gap-resolution.md > Consequential-design classes carry their citations`).

## BUILD — write the portable plan

Shape the smallest plan that makes execution mechanical and remains proportional to the work.
Use the canonical plan format for one Markdown file with three separate concerns:

- **Contract:** human-readable Markdown containing the objective, scope, exclusions, decisions,
  interfaces, tasks and verification criteria. Keep the existing task/interface lists in the
  delimited contract; do not replace them with a JSON schema. Once approved, its delimited bytes
  are frozen and receive the canonical digest.
- **Authorization:** explicit user evidence, approved contract digest, permitted actions,
  targets and conditions. Implementation and delivery permissions are separate; delivery may
  be authorized up front when the user says so.
- **Execution:** status, task evidence, verification results and delivery reconciliation.
  Progress belongs here, never in the frozen contract. Use ordinary ordered steps; do not turn
  the contract's steps into progress checkboxes.
  Keep recovery attempts in the same plan, outside the contract, using the plan-format schema.
  Initialize a new, never-started plan with known empty history; reconcile an existing plan
  before recording its history as known. Missing history is not evidence of zero attempts.

Use the proportional phase model:

```text
explore → define → design/plan → implement → verify/review → deliver/close
```

A bounded change may combine define and design/plan. A larger change may need separate
specification, design or contract artifacts. A review gate remains the gate when the artifact
being reviewed is itself the deliverable; do not manufacture a second equivalent approval.

Before presenting the gate, resolve technical prerequisites that are discoverable and record
missing user-held prerequisites explicitly. Do not write implementation tasks whose files,
interfaces, test approach, verification command or expected result remain unknown. Every task
declares `Test approach: tdd`, `characterization`, or `not-applicable` according to
`quality/testing.md`; a `not-applicable` choice includes its concrete reason in the task and
mutable test-evidence table.

## APPROVAL — freeze the contract

Present the complete candidate and ask for one explicit decision covering the contract. Keep
these outcomes distinct:

- approval of the problem, scope or design permits planning to continue but does not authorize
  implementation;
- explicit implementation consent permits `/flow-build` to edit the authorized scope;
- explicit delivery consent is the session git mode chosen at the gate, recorded as the grant
  set `plan-format.md > Delivery grants from the session mode` maps it to (`interactive` and
  `automatic` both include `merge` into the base; `deploy` only when the user names it), within
  the repository's mandatory gates.

Record the user's evidence and the resulting action/target scope in Authorization, including any
conditions that delivery must satisfy. Later fresh evidence may mark a preauthorized condition
pending or satisfied without new consent; only a changed, expired or failed condition requires
resolution before the action. A status line, the existence of a plan, a native Plan Mode approval,
a green review or an invocation of `/flow-plan` is not consent to implement or publish. If the user
authorizes implementation and delivery in one instruction, retain both grants for later
reconciliation without asking again unless the scope, target, condition or mandatory safety gate
changes.

The implementation grant covers the normal RED→GREEN→refactor cycle declared by each task; do not
request approval between those phases. If execution later discovers that implementation already
exists for a `tdd` task without observed RED evidence, preserve the work and stop the task at the
exception gate described in `quality/testing.md`; the explicit user exception is separate from
ordinary plan approval. For direct work or `Session: no`, keep the approach and run evidence in the current
conversation and report that no durable cross-harness resume guarantee exists. An accepted
exception permits completion only with the actual fail-to-pass comparison, pass-to-pass evidence
and applicable independent checks, labeled `exception-accepted` rather than strict TDD.

Set `draft` while the contract is being formed or awaiting approval. Before setting `planned`,
run the shared read-only validator against the structure, contract digest and authorization
record. `planned` makes a plan eligible for execution; it does not invent an implementation or
delivery grant that is absent from Authorization.

Material changes to the objective, contract scope, interfaces, tasks or verification create a
new contract revision and require new approval. Expanding Authorization to an operation or target
already covered by the frozen contract records new user evidence and increments the authorization
revision without changing the contract digest. Narrowing or revoking a grant updates Authorization
or revocations without a contract revision; an operation or target outside the frozen contract
scope requires one. Editorial progress or evidence updates do not change the contract digest. Do
not silently adopt an old `Status: planned` plan as authorized: normalize legacy content first,
then record the current explicit request as grant evidence when it covers the unchanged scope, or
clarify the scope before editing.

## CLOSE — hand off or return the plan

When a ledger exists, update its phase, artifact index and `## Current handoff` with the plan path,
contract revision, decisions, open questions and the next suggested action. Report whether the
plan is `draft` or `planned`, which permissions were explicitly granted, and which remain absent.

If the approval includes implementation authority, stop after the approval record and ask the
user to run `/flow-build`, naming the direct route as the alternative; the recorded authorization
evidence and delivery grants carry over to whichever route they pick. Never call `/flow-build`
yourself — the harness refuses it. If it is planning-only, return the plan and offer the next
execution route; a user may also use the direct route for a small change. The plan does not force
a full ceremony where proportionality says it adds no value.

Stable path after deploy: `~/.agents/skills/flow-plan/SKILL.md`.
