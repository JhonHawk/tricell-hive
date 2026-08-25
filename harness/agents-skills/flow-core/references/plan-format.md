# Plan format — the executable, harness-agnostic plan

A plan written in native plan mode (captured by the plan-capture hook) is **two things at
once**: the contract another harness executes, and the durable state of that execution.
Both roles impose the same discipline — the plan must be self-sufficient.

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
  *stakeholder* rows fold into the plan-approval question block.
- A row that blocks no task does not belong here — resolve it inline. Omit the whole section when
  no decision blocks detail; never include it empty.

## Preflight — resources confirmed at plan time, not at point of use

The plan's approval is the last interruption; a missing credential found mid-execution kills the
autonomy. Derive the resource list from the WHOLE flow (implementation, the in-vivo gate, and what
promoting to qa/prod will need) and resolve it NOW. Check presence, never print values:

| Resource class | Verify |
|---|---|
| Env/config | `.env.<env>` and config files the tasks read exist |
| CLI auth | `gh auth status`, `aws sts get-caller-identity` / `hcloud` for the accounts touched |
| Domain tools | CLIs beyond gh/aws/hcloud (tunnels, webhook simulators, provider CLIs) — `which`/`--version`; missing → ask before installing |
| Services | DB/Redis/queues reachable (or note how they start) |
| Integrations | tracker access works; external sandbox tokens present — including the in-vivo gate's credentials |
| Test baseline | the touched suite runs before T1 (affected subset per `testing.md`); a red baseline is a Preflight decision for the user, never absorbed silently |

What is checkable gets reported `ok`/`missing`; what needs the user goes in a **Preflight section**
of the plan as explicit asks. The plan is not ready for approval while a known-needed resource is
unresolved. This applies to any plan that will be executed, including one written in native plan
mode and captured by the plan-capture hook.

## Delivery pauses — how many times the user sees it before the plan ends

One review of everything at the end is where a plan turns into a pile of rework. **Plans with 4+
file-modifying tasks** (the checkpoint threshold in `git-mechanics.md > Commits`) close this at the
plan gate, BEFORE the first edit — splitting at push time saves nothing, the lines are already
written. The gate proposes, it does not ask for a number:

1. **N pauses**, each one named: after which task it falls and what the user can exercise there.
   The cuts land on seams the plan already declares — a task whose `Interfaces: Produces` closes a
   contract the following ones consume, never an arbitrary count of tasks.
2. **One review when the whole plan is met** — a single delivery.

Accepting pauses IS choosing `interactive` for those stretches: each pause is a validation, and the
chain resumes on it. Under `hold` it does not apply — everything is a diff at close. The chosen
answer goes in the plan and binds execution; a plan that runs past its own pause is a defect.

**Review-workload forecast rides the same gate.** Estimate the change-group's changed lines; past
~400 (the same threshold `agent-routing.md` uses for refuter fan-out — one operative number in the
system), the gate proposes the PR split (chained or stacked, cut at the plan's own seams) alongside
the pauses — the chain strategy is the user's pick. Splitting at push time saves nothing.

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

When the scope proves too dense for one plan during native plan-mode writing, propose
splitting and create an **initiative** (`project-structure.md > Session capture layer`):

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
