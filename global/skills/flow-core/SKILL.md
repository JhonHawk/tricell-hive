---
name: flow-core
description: >
  Shared protocol and templates for the flow pack: the live skills (flow-plan, flow-build,
  flow-report)
  and the playbooks that preserve the retired flow-* commands' knowledge (bootstrap,
  migration, spec-writing, workspace-hygiene, audit, promotion). Not a workflow itself — it
  is the library every flow-* skill and playbook reads for the flow contract, the
  file-routing rule, and the canonical templates (ledger, handoff protocol, naming table,
  specs structure).
user-invocable: false
---

# flow-core — shared library for the flow pack

This skill is never executed as a workflow. It holds the decisions every `flow-*` skill
must agree on, so they live in exactly one place. Design rationale and full phase specs live
with the pack's own source repository — not something to go looking for from a project
session.

## The flow contract

Every `flow-*` skill follows this contract.

1. **OPEN (DoR)** — read `<project>/_support/PROJECT.md` (the ledger) when it exists. If the
   workspace declares a ledger and it is missing, stop and follow `references/bootstrap-playbook.md`
   to establish it; a standalone repository may use its natural `_support/sessions/` home without
   bootstrapping a ledger merely to plan or execute an explicitly identified plan. Never
   improvise workspace structure. Then **consume the `## Current handoff` section** when present —
   it is the bounded context that phase produced for you (paths, decisions, what changed, the
   repos to treat as input). Verify the phase's entry preconditions; if a needed input is missing,
   that gap is the first thing to surface, not something to work around. Once consumed, collapse
   the handoff to one line in `## Handoff history` (poda — see the phase transition contract below).
2. **EXPLORE (when the phase needs it)** — when you cannot yet name the concrete files,
   paths, or repos a dispatch will need, run a bounded read-only exploration FIRST and let
   it return the map. Proportional: a bootstrap phase (kickoff) explores nothing; a specs
   review or dev session usually must. Dispatch the read-only explorer
   (`references/harness-mechanics.md` translates the mechanic per harness) rather than
   exploring inline — discovery noise stays out of the orchestrator's context.
3. **ROUTE FILES** — before writing any file, apply the file-routing rule below. A plan is a
   portable carrier only when its contract, authorization and execution state use the canonical
   plan format; a native harness plan is merely an optional source to normalize.
4. **ORCHESTRATE** — the main thread routes and synthesizes. It does NOT implement,
   review, or verify by itself. Every substantive work unit goes to an agent in a fresh
   context; only summaries return to the main thread. Declared exceptions: mock work
   units (built via `/flow-plan`/`flow-build` like any other unit — prototype carve-out) and
   flow-build on a harness without specialist agents may implement in the main thread —
   review and verification stay in fresh contexts everywhere.
5. **HANDOFF** — every agent prompt follows `references/handoff-protocol.md`: intent and
   bounded context are what keep the dispatched work aligned.
6. **CLOSE (DoD)** — update PROJECT.md when a ledger exists (phase, artifacts with paths, decisions and whether
   they were promoted, open questions), **then write the `## Current handoff` section**:
   what this phase produced (paths), decisions left, what it changed backward (spec-changes
   triggered), the repos the next phase consumes as input, and the **next phase suggested
   from the ledger state** (which phases are `done` for this epic — not a fixed F+1). Then
   report to the user. The handoff IS the report the next phase reads; the ledger condenses,
   the handoff carries the detail.

## Phase transition contract (handoff between phases)

The contract's frontier is explicit: a phase declares what it produces at CLOSE (DoD) and
what it verifies and consumes at OPEN (DoR). The carrier is the **handoff** — a section in
the ledger, not a separate file (one archive, no extra pointers to break; the DoR already
reads PROJECT.md). It lives in two parts:

- **`## Current handoff`** — only the *last closed phase's* handoff, expanded. The next
  phase consumes it at OPEN.
- **`## Handoff history`** — one line per consumed handoff. When a phase consumes the
  current handoff, it **collapses it to a single line here** (poda). The handoff is
  transitory state that expires on consumption, so the durable ledger never accumulates
  dead detail; history keeps one line per phase indefinitely (negligible cost, full trail).

Form and poda mechanics live in `references/ledger-template.md`.

**Gate proportional to risk.** Whether a phase presents an approval gate before executing
scales to the cost of undoing its writes, never uniform:

| Phase writes | Gate |
|---|---|
| Production / real code (`/flow-plan` + `/flow-build`) | Strong portable plan gate — the contract and implementation authority are explicit |
| Resources derived from a signed naming table (bootstrap-playbook foundation stage, repo matrix) | Signal-gated: proceed-and-report on clean derivation; gate on a new naming exception, unsettled repo split, or client-org blast radius (highest-risk bootstrap gate) |
| A mock work unit (a whole repo cheap-to-rebuild but costly-to-redo) | Light `/flow-plan` gate: epics/screens/stack/order, approved before building |
| A draft that re-enters its own review gate (the spec-writing playbook's `epic` step) | The review gate IS the gate; no separate plan gate |
| Bootstrap-playbook conversational stages (intake scorecard, workspace bootstrap confirm) | Per-stage conversational gate — a scorecard/confirm the user answers, no plan file |
| Promotion to an environment (qa/prod) | Git-workflow safety gates (`git-workflow.md`) + `references/promotion-playbook.md` — not a flow skill gate |
| Deterministic bootstrap (bootstrap-playbook structure, spec-writing playbook's `init` step) | No gate — a plan adds friction without reducing risk |

A gate heavier than the phase's reversibility is ceremony; lighter is a foot-gun.

## Phase artifact persistence (Engram resume-mirror)

When a ledger exists, `PROJECT.md` is the primary, durable phase state and flow skills update it
on CLOSE. A standalone plan without a ledger keeps its durable state in the plan's metadata and
execution evidence; do not create a ledger solely to satisfy this mirror. Engram is a
*resume-mirror*, not the source of truth: it lets a compacted session recover a phase's
intermediate artifact without re-deriving it. When a flow skill mirrors one, use a **deterministic
`topic_key`** so the save upserts instead of duplicating:

```
topic_key = flow/{epic-or-project-slug}/{artifact}
artifact ∈ brainstorm | spec | dev-progress | release
```

Same `topic_key` + `project` + `scope` → UPDATE, not INSERT. Retrieve with `mem_search`
followed by `mem_get_observation(id)` — search alone returns truncated content. If the ledger
and an Engram mirror disagree, the ledger wins; refresh the mirror.

## File-routing rule

Before writing any file in a client workspace, answer in order:

1. Must it have history/versioning? → `<project>-specs/` (the versioned project memory)
2. Is it temporary, sensitive, raw evidence, or scratch? → `_support/`
3. Does it apply to a single repo? → `<repo>/_support/` or the repo's natural location
4. Did a temporary report produce a decision? → promote/summarize it into `<project>-specs/`

`_support/` is the NON-versioned layer: it points and expires, it never archives.
`<project>-specs/` preserves. When in doubt between the two, ask what would be lost if the
workspace folder disappeared — anything that matters belongs in the specs repo.

**No `<project>-specs/` yet** (a standalone single repo, or a workspace before its specs
repo exists): the versioned durable layer is the repo's own committed `_support/` —
repo-level `_support/` is versioned *except* `workspace/` — placed via the session-capture
detection rule in `project-structure.md` (specs repo → `<project>-specs/sessions/`;
standalone → `<repo>/_support/sessions/`; workspace pre-specs → `<project>/_support/sessions/`
as staging, `git mv` into the specs repo once it exists). The invariant never changes:
durable artifacts go to the versioned layer for the current shape, NEVER to gitignored
`_support/workspace/`.

**Session reports** (review/QA/test reports a flow phase renders): route to the versioned
session-capture layer for the current shape — `<project>-specs/sessions/<slug>/reports/`
or `<repo>/_support/sessions/<slug>/reports/` (detection rule: `project-structure.md`).
Raw evidence (screenshots, logs) stays in `_support/evidence/<slug>/` (gitignored),
referenced by path — never embedded in the versioned report. Flow skills cite this
instead of restating it. **Declared carve-out — baseline-comparison reports
(the audit-playbook):** cross-run comparison needs one stable home plus a history table, which
per-slug session folders can't give, so audit runs write to
`<project>-specs/audit/reports/YYYY-MM-DD.html` with the instance README as index
(`references/audit-playbook.md` owns the layout; sanctioned in `specs-structure.md`).
Without a specs repo, the standard session form above applies unchanged.

## Templates (read on demand)

| Reference | When to read it |
|---|---|
| `references/ledger-template.md` | Creating PROJECT.md (per `bootstrap-playbook.md`) or repairing/reconstructing it (`workspace-hygiene-playbook.md`, `migration-playbook.md`) |
| `references/judgment-criteria.md` | Judging workspace artifacts (`workspace-hygiene-playbook.md`'s audit, `migration-playbook.md`) |
| `references/audit-playbook.md` | Running or instantiating a multi-lens preventive audit |
| `references/migration-playbook.md` | Bringing a pre-pack project into the flow |
| `references/handoff-protocol.md` | Before dispatching ANY agent from a flow skill; also the research→write→build→verify phase-handoff chain |
| `references/plan-format.md` | Writing the portable plan (`/flow-plan`) or executing one (`/flow-build`) — the frozen contract, authorization and execution-state contract |
| `references/naming-template.md` | Instantiating the project naming table (bootstrap-playbook foundation stage) or auditing it (promotion `verify`) |
| `references/promotion-playbook.md` | Promoting to qa/prod via git conventions — read by deploy sessions and `devops-engineer`, offered by the session hook |
| `references/ux-rubric.md` | The design/UX gate — consumed by `flow-build`'s design gate and by mock-review work |
| `references/test-report-template.md` | Writing the versioned in-vivo/QA report (`flow-build` gate; the QA promotion walk via `promotion-playbook.md`) |
| `references/specs-structure.md` | Creating the specs repo (`spec-writing-playbook.md`'s `init` step, `migration-playbook.md`) or checking conformance (`workspace-hygiene-playbook.md`) |
| `references/release-notes-template.md` | Drafting client release notes (the QA/prod promotion walk via `promotion-playbook.md`) |
| `references/harness-mechanics.md` | You are NOT Claude Code (Codex/opencode reading these skills from `~/.agents/skills/`) — translates mechanic names before executing any flow skill |

Stable path after deploy: `~/.claude/skills/flow-core/references/<file>.md`.

**Only the files listed above live here.** A `global/rules` file cited by name from inside
one of them belongs to a router skill, not to `flow-core`: `project-structure.md` and
`session-capture.md` → `workspace-conventions/references/`, `memory-routing.md` →
`memory-policy/references/`, language and framework rules → `language-rules/references/`.
Resolving those against `flow-core/references/` fails silently and the work proceeds
without the rule.

## Pack map

The pack is organized by **project stage**, not a fixed phase sequence. The daily work loop
(brainstorm → spec → plan → execute) lives *inside* the `desarrollo` stage and repeats per
unit of work; a mock is just a work TYPE that runs the same loop.

Most former per-stage commands are dissolved: their knowledge lives as playbooks in
`references/`, applied directly rather than invoked. `flow-plan`, `flow-build` and
`flow-report` are the live skills.

| Project stage | How it runs |
|---|---|
| `arranque` | Applied via `references/bootstrap-playbook.md` — greenfield: intake + workspace bootstrap + foundation (repos, naming table, CI/CD), fused into one conversational flow. Pre-pack projects enter via `references/migration-playbook.md` (specs repo, tiering, gated migration manifest) |
| `specs` | Applied via `references/spec-writing-playbook.md` — specs repo (`init`), epic drafting + tracker sync (`epic`), quality gate (`review`) |
| `desarrollo` | the daily chain, per unit of work: idea exploration (converges on proceed/discard/defer per `rules/quality/critical-thinking.md`) → `references/spec-writing-playbook.md` → `/flow-plan` freezes the portable contract and authorization → `/flow-build` executes and verifies. Mock work units run the same chain |
| `operación` | promotion to qa/prod via git conventions (`git-workflow.md`) + `references/promotion-playbook.md` — no dedicated skill |

Transversal (not a stage):

| Skill / playbook | Role |
|---|---|
| `references/workspace-hygiene-playbook.md` | Compensating control: audit/apply workspace hygiene for drift from conversational sessions |
| `references/audit-playbook.md` | Multi-lens preventive audit of runtime repos (epic close, pre-architectural change): parallel readers → dedup → refuters → versioned HTML report |
| `flow-plan` | User-invoked planning skill: creates or normalizes the portable plan and records its approval scope; it never implements or publishes |
| `flow-report` | Shared rendering skill (like flow-core, not a stage): renders substantial human-targeted output as self-contained HTML; auto-invokes per `rules/quality/communication-format.md` |
| `flow-core` | This library: the flow contract, file-routing rule, and canonical templates every flow-* skill and playbook reads |
