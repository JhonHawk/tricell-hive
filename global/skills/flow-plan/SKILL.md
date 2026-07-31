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
follow; describe the action and let `flow-core/references/harness-mechanics.md` translate the mechanic.

## Subcommands (read the reference for the one you are running)

Route on `$ARGUMENTS[0]`: `research` | `write`. No subcommand → infer from ledger state (no
findings yet → `research`; findings present, no plan → `write`) and state which you chose.
`$ARGUMENTS` after the subcommand scopes the run (epic ID or task IDs).

| Subcommand | Read | Produces | Gate |
|---|---|---|---|
| `research` | `references/research.md` | `<slug>-findings.md` — current state, approaches, gaps, risks, version notes | none (read-only) |
| `write` | `references/write.md` | `<slug>-plan.md` (or an initiative's `plan/` parts) with reconciler metadata + Preflight | the plan gate — the technical counterpart of `/flow-specs review` |

Stable path after deploy: `~/.claude/skills/flow-plan/references/<file>.md`.
