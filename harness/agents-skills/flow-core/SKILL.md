---
name: flow-core
description: >
  Shared protocol and templates for the flow pack (flow-intake, flow-kickoff, flow-specs,
  flow-mock, flow-foundation, flow-plan, flow-build, flow-deploy, flow-hygiene, flow-report). Not a workflow itself —
  it is the library every flow-* skill reads for the flow contract, the file-routing rule,
  and the canonical templates (ledger, handoff protocol, naming table, specs structure).
---

# flow-core — shared library for the flow pack

This skill is never executed as a workflow. It holds the decisions every `flow-*` skill
must agree on, so they live in exactly one place. Design rationale and full phase specs:
`tricell-hive/_support/spec/flow-pack-design.md`.

## The flow contract

Every `flow-*` skill follows this contract. It exists because the failure mode it prevents
is well documented in this workspace's history: the main thread doing all the work itself,
agents sitting unused, files landing in improvised locations.

1. **OPEN (DoR)** — read `<project>/_support/PROJECT.md` (the ledger). If it does not
   exist and the current skill is not `flow-kickoff`: stop and suggest `/flow-kickoff`.
   Never improvise workspace structure. Then **consume the `## Current handoff` section**
   if the previous phase left one — it is the bounded context that phase produced for you
   (paths, decisions, what changed, the repos to treat as input). Verify the phase's
   entry preconditions; if a needed input is missing, that gap is the first thing to
   surface, not something to work around. Once consumed, collapse the handoff to one line
   in `## Handoff history` (poda — see the phase transition contract below).
2. **EXPLORE (when the phase needs it)** — when you cannot yet name the concrete files,
   paths, or repos a dispatch will need, run a bounded read-only exploration FIRST and let
   it return the map. Proportional: a bootstrap phase (kickoff) explores nothing; a specs
   review or dev session usually must. Dispatch the read-only explorer
   (`references/harness-mechanics.md` translates the mechanic per harness) rather than
   exploring inline — discovery noise stays out of the orchestrator's context.
3. **ROUTE FILES** — before writing any file, apply the file-routing rule below.
4. **ORCHESTRATE** — the main thread routes and synthesizes. It does NOT implement,
   review, or verify by itself. Every substantive work unit goes to an agent in a fresh
   context; only summaries return to the main thread.
5. **HANDOFF** — every agent prompt follows `references/handoff-protocol.md`. A handoff
   missing intent or bounded context produces misaligned work — the protocol is not
   optional ceremony.
6. **CLOSE (DoD)** — update PROJECT.md (phase, artifacts with paths, decisions and whether
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
| Production / real code (flow-plan/flow-build, flow-deploy) | Strong plan gate — already defined in those skills |
| A whole repo cheap-to-rebuild but costly-to-redo (flow-mock `build`) | Light plan gate: epics/screens/stack/order, approved before building |
| A draft that re-enters its own review gate (flow-specs `epic`) | The review gate IS the gate; no separate plan gate |
| Deterministic bootstrap (flow-kickoff, flow-specs `init`) | No gate — a plan adds friction without reducing risk |

A gate heavier than the phase's reversibility is ceremony; lighter is a foot-gun.

## Phase artifact persistence (Engram resume-mirror)

`PROJECT.md` (the ledger) is the primary, durable phase state — always update it on CLOSE.
Engram is a *resume-mirror*, not the source of truth: it lets a compacted session recover a
phase's intermediate artifact without re-deriving it. When a flow skill mirrors one, use a
**deterministic `topic_key`** so the save upserts instead of duplicating:

```
topic_key = flow/{epic-or-project-slug}/{artifact}
artifact ∈ intake | spec | mock | foundation | dev-progress | deploy-report
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

## Templates (read on demand)

| Reference | When to read it |
|---|---|
| `references/ledger-template.md` | Creating PROJECT.md (`flow-kickoff`) or repairing it (`flow-hygiene`) |
| `references/handoff-protocol.md` | Before dispatching ANY agent from a flow skill; also the research→write→build→verify phase-handoff chain |
| `references/plan-format.md` | Writing an executable plan (`flow-plan write`) or executing one (`flow-build`) — the plan-as-state contract |
| `references/naming-template.md` | Instantiating the project naming table (`flow-foundation`) or auditing it (`flow-deploy verify`) |
| `references/specs-structure.md` | Creating the specs repo (`flow-specs init`) or checking conformance (`flow-hygiene`) |
| `references/release-notes-template.md` | Drafting client release notes (`flow-deploy qa`/`prod`) |
| `references/harness-mechanics.md` | You are NOT Claude Code (Codex/opencode reading these skills from `~/.agents/skills/`) — translates mechanic names before executing any flow skill |

Stable path after deploy: `~/.agents/skills/flow-core/references/<file>.md`.

## Pack map

| Skill | Phase | Purpose |
|---|---|---|
| `/flow-intake` | F1 | Analyze/improve a client requirements document; surface blocking questions |
| `/flow-kickoff` | F2 | Bootstrap the workspace: 3-level structure, ledger, workspace CLAUDE.md |
| `/flow-specs` | F3 | Specs repo (init), epic drafting + tracker sync (epic), quality gate (review) |
| `/flow-mock` | F4 | Navigable prototype (build), UX friction review (review) |
| `/flow-foundation` | F5 | Repos, schema, contracts, naming table, CI/CD before the first deploy |
| `/flow-plan` | F6 | Plan the dev session: research the codebase (`research`), write the executable, harness-agnostic plan (`write`) |
| `/flow-build` | F6 | Execute the plan: state-driven reconciler that dispatches implementation, then verifies (verify gate) and closes |
| `/flow-deploy` | F7 | Gated deploys (qa/prod), post-deploy verification + naming audit (verify) |
| `/flow-hygiene` | — | Compensating control: audit/apply workspace hygiene for drift from conversational sessions |
| `flow-report` | — | Shared rendering skill (like flow-core, not a phase): renders substantial human-targeted output as self-contained HTML; auto-invokes per `rules/quality/communication-format.md` |
