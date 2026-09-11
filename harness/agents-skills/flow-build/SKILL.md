---
name: flow-build
description: >
  Execute an approved plan (the executing stage of the daily dev chain) as a state-driven
  reconciler — resumable, never re-doing landed work. `verify` jumps straight to the
  verification gate. Runs on any harness. Point it at a plan or part.
disable-model-invocation: true
---

# /flow-build — execute the portable plan

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This is the *doing* half of
the dev chain: it executes a portable Hive plan produced or normalized by `/flow-plan`. It is a
**reconciler** — it reads the frozen contract, authorization and execution evidence, then
converges toward the next pending state. Native harness Plan Mode is optional and has no
authority over this flow.

`/flow-build` requires an explicit user request for the requested next action. The default
execution path requires an `implement` grant and authorizes implementation only to the extent
recorded in the plan's Authorization section. The `verify` subcommand uses its explicit invocation
as verification authority: OPEN inspects the bound authorization and requires status `valid` or
`conditional` with observed plan status `built` or `verified`; it does not call
`validate --action verify` and requires no implementation grant. A prior explicit delivery grant
is honored when its digest, target and conditions still match; a preauthorized condition may be
recorded as satisfied with fresh evidence without asking again. If a condition changed, expired or
failed, stop and resolve it before the action. Mandatory repository and safety gates still apply.
Invoking `/flow-build`, passing a green review, or reaching `verified` never grants publication by
itself. A `verify` invocation never starts implementation or publication.

**Runs on any harness.** Where specialist agents are available, dispatch each task to its
specialist (`agent-routing.md`) in a fresh context and keep the main thread orchestrating — the
flow-pack rule. Where they are not (a lighter harness), implement the task's recipe directly:
the plan is self-contained (`flow-core/references/plan-format.md`) precisely so a cheaper executor can run
it. The in-vivo gate is **not** frontier-only — it runs here too; a teammate who only has this
harness still gates their work.

## Stages (read the reference for the stage you are in)

`$ARGUMENTS`: an optional `verify` subcommand plus an optional plan-or-part path. The default
run walks OPEN → Execute → Gate → CLOSE; `verify` resolves the plan at OPEN, then jumps
straight to the Gate (it assumes `built`) and closes without reconciling delivery.

| Stage | Read | When |
|---|---|---|
| OPEN — reconcile | `references/reconcile.md` | Every invocation: resolve the plan, normalize legacy content before approval, validate authority and read the pending state from disk/Git/evidence |
| Execute (`planned`/`building` → `built`) | `references/execute.md` | The reconciled state is `planned` or `building` and implementation is explicitly authorized; skipped by `verify` |
| Gate (`built` → `verified`) | `references/verify-gate.md` | The reconciled state is `built`, or the `verify` subcommand — two-stage review, in-vivo gate, design gate, integrated smoke |
| CLOSE | `references/close.md` | Every invocation, after the run's last stage |

Stable path after deploy: `~/.agents/skills/flow-build/references/<file>.md`.
