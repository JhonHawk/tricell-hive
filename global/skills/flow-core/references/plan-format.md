# Plan format — the executable, harness-agnostic plan

A plan written by `flow-plan write` is **two things at once**: the contract another harness
executes, and the durable state of that execution. Both roles impose the same discipline —
the plan must be self-sufficient.

## Two invariants

1. **Self-contained — written for an engineer with zero context.** The executor may be a
   different model on a different harness that never saw the research, the conversation, or
   the spec discussion. Every task carries its exact files, interfaces, atomic steps, and a
   verification command with its expected output. No "as discussed", no "similar to T2", no
   "TBD". The judgment was front-loaded by the planner; the executor runs and compares.
2. **Harness-neutral.** No step names a Claude-Code-only mechanic. Where a step needs a
   harness mechanic (dispatch, plan-gate, review), describe the *action* and let
   `references/harness-mechanics.md` translate it. The plan must read the same to grok on
   opencode and to Claude Code.

## Plan header

```markdown
# <Feature> — Plan

Status: planned        ← planned | building | built | verified  (see state machine below)
Implements: <epic / task / decision refs>     ← back-reference to the intention layer
Spec: <path to design spec, if any>
Goal: <one line>
Architecture: <2-3 sentences>
Integration: <integration branch · merge mechanics · which CI gates each PR>
```

`Status` is the only line that mutates during execution — keep it on its own line, first,
for cheap reads (`flow-build` reads it before anything else).

## The state machine (the plan IS the state)

The pair *(`Status` header, git)* is the entire execution state — no separate ledger,
nothing in session memory. Any harness reconciles from these two and continues.

```
planned ──build──► building ──all tasks landed──► built ──in-vivo gate──► verified
```

| `Status` | Meaning | Source of truth |
|---|---|---|
| `planned` | written, approved, nothing executed | — |
| `building` | execution in progress | which tasks are done = which appear in `git log` |
| `built` | every task landed in git | git log shows all task IDs |
| `verified` | the in-vivo/review gate passed | the versioned in-vivo report |

- **Task-level state lives in git, never in the plan.** A task is done when its commit is in
  the log — not when a box is checked. This is why the plan stays trustworthy across
  harnesses and compactions: git is the authority, the header is a coordination flag the
  reviewer re-checks against git before trusting (same ledger-vs-git discipline as the rest
  of the pack).
- **`flow-build` advances `Status`**, never the planner. The reviewer advances `built →
  verified`.

## Decisions to close before executing

When any decision blocks task detail, the plan carries a **Decisions to close BEFORE executing**
section above the tasks — the approval gate resolves them so execution is mechanical, never a
drip of mid-task questions (`gap-resolution.md > Decisions to close before executing`).

```markdown
## Decisions to close BEFORE executing

| Decision | Point | What's decided | Type | Recommendation | Blocks |
|---|---|---|---|---|---|
| D1 | <spec/AC ref> | <the choice> | technical | <your call> | T3, T4 |
| D2 | <spec/AC ref> | <the choice> | stakeholder | <your call> | T1 |

> Resolution path: technical rows confirmed with a peer/tool; stakeholder rows ratified in the
> approval gate.
```

- *Technical* rows the planner confirms with a peer/tool (a second model, context7, a quick test);
  *stakeholder* rows fold into the plan-approval question block (`flow-plan write`, step 6).
- A row that blocks no task does not belong here — resolve it inline. Omit the whole section when
  no decision blocks detail; never include it empty.

## Task block

Every task is independently executable and independently trackable.

```markdown
### T<n>: <imperative title>

in-vivo: yes | no        ← does this task need a live walk? planner decides (UI/integration → yes; pure logic → no)
design-review: yes       ← opt-in: user-facing UI task → flow-build's Visual-craft gate walks it; omit the line for non-UI tasks
Agent: <agent-name>      ← optional routing annotation: the routing-table executor when a row covers the task (agent-routing.md); omit when none applies
Files:
  - Create: <exact/path>
  - Modify: <exact/path:lines>
Interfaces:
  - Consumes: <signatures/types from earlier tasks — exact>
  - Produces: <signatures/types later tasks depend on>

- [ ] Step 1: <one action — code, command, or edit>
- [ ] Step 2: ...
Verify: `<command>` — Expected: `<result>`
Commit: feat(<scope>): T<n> <subject>
```

Rules:
- **`<n>` is local to the plan** (`T1`, `T2`…), not a tracker ID. The mapping to Linear/Jira
  is an optional downstream projection — the plan never depends on the tracker to know what
  is done. The commit's `T<n>` tag is what derives task-state from `git log`.
- **`in-vivo:` is mandatory per task** so the gate is never assumed or omitted. `flow-build`
  reads these to know which tasks need a walk; the *timing* (inline vs deferred) is a
  once-per-run decision, not per task (see `flow-build`). **`design-review:` is its opt-in
  sibling** — set `yes` on user-facing UI tasks to route them through flow-build's
  Visual-craft design gate; absent means no.
- **The walk of an `in-vivo: yes` task runs via `in-vivo-qa-tester` wherever the agent roster
  exists** — the same dispatch `flow-build`'s gate makes mechanical, and it includes the role's
  adversarial half, not just the plan's happy-path checklist. `Agent:` makes any other
  routing-row executor plan-visible without breaking harness-neutrality: a harness without the
  roster executes the recipe directly. Running a walk or an `Agent:`-annotated task inline
  when the roster IS available is a **substitution** — reported at close with its one-line
  justification, never silent. This duty travels with the plan: a protocol run by hand outside
  `/flow-build` inherits it identically.
- **`Verify:` pairs a command with its expected output** (handoff-protocol element 4). A step
  whose expected result you cannot state is not a verification step yet. This is what lets a
  cheaper executor verify mechanically instead of judging.
- **`Commit:` uses the `T<n>` tag** — one commit per completed-and-verified task, never a
  batch. The tag is load-bearing: it is how any harness reads execution state from git.

## Large scope — split into parts (an initiative)

When `flow-plan write` finds the scope too dense for one plan, it proposes splitting and
creates an **initiative** (`project-structure.md > Session capture layer`):

```
sessions/<start-date>-<slug>/
  plan/
    <slug>-plan.md     ← master: header + the part index (lists 00-NN with their Status)
    00-<part>.md       ← each part is a full plan body (own Status, own tasks T1..Tn)
    01-<part>.md
```

- **Each part carries its own `Status`.** A part can be `verified` while another is `planned`
  — `flow-build` targets the part you point it at and reconciles that part independently.
- **The master plan's part index is a dashboard.** To avoid drift, the truth of each part's
  status is the part file's own header; the master lists the parts and may mirror their
  status, but the part file wins.
- A single-file plan (one `<slug>-plan.md`, no `plan/` folder) is the default; only split
  when density warrants it.

## Example (single task)

```markdown
### T3: Add idempotency key to the send-message endpoint

in-vivo: yes
Agent: backend-developer
Files:
  - Modify: src/modules/messages/messages.controller.ts
  - Modify: src/modules/messages/dto/send-message.dto.ts
Interfaces:
  - Consumes: SendMessageDto (from T2)
  - Produces: header `Idempotency-Key` honored; duplicate within 10m → 200 with prior result

- [ ] Step 1: add `idempotencyKey?: string` to SendMessageDto with @IsUUID() @IsOptional()
- [ ] Step 2: in the controller, look up the key in Redis before persisting; on hit, return the stored response
- [ ] Step 3: on miss, persist (key → response, TTL 10m) after a successful send
Verify: `pnpm test messages.idempotency.spec` — Expected: `4 passing, 0 failing`
Commit: feat(messages): T3 idempotency key on send endpoint
```
