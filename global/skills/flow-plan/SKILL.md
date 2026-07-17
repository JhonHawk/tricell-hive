---
name: flow-plan
description: >
  The planning stage of the daily dev chain: captures/adopts the approved native plan — one
  plan per unit of work. `research` runs a read-only technical investigation of a scope and
  writes findings; `write` turns findings + spec into an executable, harness-agnostic plan.
  Use before `/flow-build`. Scope with an epic ID or explicit task IDs.
argument-hint: "[research | write] [epic-id | TASK-IDs ...]"
disable-model-invocation: true
---

# /flow-plan — research and write the plan

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`). This skill is the *thinking*
half of a dev unit's chain: it explores and plans, it never executes — execution is `/flow-build`.

**Positioning — the formal/epic track.** The organic path produces the same artifact: a
native plan-mode plan approved in a flow workspace is captured to
`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` (on Claude Code via the `flow-plan-capture`
hook; on other harnesses per the workspace conventions) and `/flow-build` adopts it
directly. Invoke `/flow-plan` when the work warrants the full formal pass — epic-scoped
research findings, preflight verification, and a plan born with reconciler metadata —
not as a prerequisite for every session. It produces
durable artifacts (`<slug>-findings.md`, `<slug>-plan.md`) that a *different harness* may pick
up cold, so everything it writes is self-contained.

**Model and harness are advisory, never a lock.** `research` and `write` benefit from the most
capable model available — use it if you have it. But neither is bound to a harness: a teammate
on a cheaper plan runs both on whatever they have. Never write a step that only Claude Code can
follow; describe the action and let `references/harness-mechanics.md` translate the mechanic.

Route on `$ARGUMENTS[0]`: `research` | `write`. No subcommand → infer from ledger state (no
findings yet → `research`; findings present, no plan → `write`) and state which you chose.

## research — technical investigation (read-only, no gate)

Reduces the unknowns so `write` can produce a complete recipe. Scoped to **implementation**, not
product or UX — those are `flow-specs` (epics/ACs) and mock work units' UX material, already upstream.

1. **OPEN** per the contract: read `<project>/_support/PROJECT.md`; missing → suggest
   `/flow-start` and stop. Consume any `## Current handoff`.
2. Read the target epic in `<project>-specs/epics/` (PRODUCT.md, tasks.md, TECH.md if present)
   and the naming table if infra is in scope. Collect the epic's **parked technical questions**
   (Open Questions marked `technical — resolves in TECH.md` by the business gate) — resolving
   them is part of this investigation, and the resolutions land in the epic's TECH.md, not in
   PRODUCT.md. `$ARGUMENTS` after the subcommand overrides scope (epic ID or task IDs).
3. **EXPLORE in subagents, never inline** (context hygiene — discovery noise stays out of the
   orchestrator). Dispatch read-only explorers per the handoff protocol to establish:
   - **Current state** — how this repo already does the thing; patterns, conventions, and
     existing code to reuse (`development-principles.md > Search/Observe before creating`).
   - **Approaches** — 2-3 technical options with trade-offs and a recommendation.
   - **Gaps and prerequisites** — missing interfaces, uninstalled libs, migrations, env/config,
     credentials (the `gap-resolution.md` Investigate step, materialized as an artifact).
   - **Risks** — failure modes and where the plan will need judgment.
   - **Version-sensitive surfaces** — query context7 for the installed versions so `write` can
     cite exact signatures.
4. **Write `<slug>-findings.md`** to the session/initiative home (per `project-structure.md`):
   current-state map, approaches + recommendation, gaps/prerequisites, risks, version notes.
   Read-only — no source changes, no gate.
5. **Divergence routes back.** If research finds the epic infeasible as specced, or a clearly
   better approach implies changing the spec, surface it and route to `/flow-specs` — do not
   silently deviate (`gap-resolution.md > Divergence Between Sources`).
6. **CLOSE**: update the ledger handoff with the findings path; suggest `/flow-plan write`.

## write — the executable plan (plan gate)

Crystallizes findings + spec into a plan another harness executes. The plan is the contract
**and** the execution state — write it to the standard in `references/plan-format.md`.

1. **OPEN**: read PROJECT.md, the `<slug>-findings.md` if `research` ran, and the target epic.
2. **Produce the plan per `plan-format.md`**: the `Status` header, and one self-contained block
   per task — local ID `T<n>`, exact Files, Interfaces (consume/produce), atomic steps, the
   `in-vivo: yes/no` flag (you decide it: UI/integration → yes, pure logic → no — so the gate is
   never assumed or omitted downstream), a `Verify:` command paired with its expected output, and
   the `Commit: feat(<scope>): T<n> …` tag. Declare the integration semantics (branch, merge
   mechanics, which CI gates each PR) in the header. Routing each task to a specialist is
   `flow-build`'s job at execution time — the plan stays harness-neutral. When any decision
   blocks task detail, add a **Decisions to close BEFORE executing** table above the tasks
   (`plan-format.md`): technical rows you confirm with a peer/tool, stakeholder rows folded into
   the approval gate at step 6 — so execution never drips questions mid-task. **This gate is the
   technical gate** (counterpart of `/flow-specs review`, the business gate): the epic's parked
   technical questions must be closed — in its TECH.md, or as rows in the Decisions table — before
   the plan is ready for approval; an open parked question is a plan defect, never something
   execution absorbs silently.
3. **Preflight — resources confirmed at plan time, not at point of use.** The plan's approval is
   the last interruption; a missing credential found mid-execution kills the autonomy. Derive
   from the WHOLE flow (implementation, the in-vivo gate, and what promoting to qa/prod will
   need) and resolve NOW. Check presence, never print values:

   | Resource class | Verify |
   |---|---|
   | Env/config | `.env.<env>` and config files the tasks read exist |
   | CLI auth | `gh auth status`, `aws sts get-caller-identity` / `hcloud` for the accounts touched |
   | Domain tools | CLIs beyond gh/aws/hcloud (tunnels, webhook simulators, provider CLIs) — `which`/`--version`; missing → ask before installing |
   | Services | DB/Redis/queues reachable (or note how they start) |
   | Integrations | tracker access works; external sandbox tokens present — including the in-vivo gate's credentials |
   | Test baseline | the touched suite runs before T1 (affected subset per `testing.md`); a red baseline is a Preflight decision for the user, never absorbed silently |

   What's checkable gets reported `ok`/`missing`; what needs the user goes in a **Preflight
   section** of the plan as explicit asks. The plan is not ready for approval while a known-needed
   resource is unresolved.
4. **Large scope → split into an initiative.** When the scope is too dense for one plan, propose
   splitting into numbered parts and create the **initiative** folder (`flow-core/references/
   specs-structure.md > Session & initiative conventions`): `plan/` with a master plan + parts `00-NN`, each part its own `Status`.
   One-off scope stays a single `<slug>-plan.md`.
5. **Write the plan to the session/initiative home** (detection rule in `project-structure.md`):
   the specs repo if present, else `<repo>/_support/sessions/`. Add an `Implements:` line (epic/
   task refs) and an `in-progress` row to the sessions index.
6. **Gate the plan for approval**, by harness mode (`references/harness-mechanics.md`):
   - **Native plan mode active**: the plan document is the presentation — finalize and exit via
     ExitPlanMode; do not duplicate it as a summary.
   - **Any other mode**: present the plan FIRST (tasks, preflight `ok`/`missing`, integration
     semantics), then gate with the structured-question mechanic.
7. **CLOSE**: update the ledger handoff (plan path, next phase). **Offer the next step
   explicitly** — the exact `/flow-build` invocation and the plan path, phrased as an offer to
   run it now, plus the model/harness recommendation (capable model for review; a cheaper harness
   MAY run the build) framed as a cost choice, not a requirement. Unresolved Preflight asks the
   user must clear are the only user to-dos — list them with why they are the user's.
