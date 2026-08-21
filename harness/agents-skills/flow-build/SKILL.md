---
name: flow-build
description: >
  Execute an approved plan (the executing stage of the daily dev chain) as a state-driven
  reconciler — resumable, never re-doing landed work. `verify` jumps straight to the
  verification gate. Runs on any harness. Point it at a plan or part.
disable-model-invocation: true
---

# /flow-build — execute the plan

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This is the *doing* half of
the dev chain: it executes the plan native plan mode (captured by the plan-capture hook)
produced. It is a **reconciler** — it reads the
desired state (the plan) and the observed state (git), and converges toward the next pending
state. It never re-does landed work and never trusts a header over git.

**Runs on any harness.** Where specialist agents are available, dispatch each task to its
specialist (`agent-routing.md`) in a fresh context and keep the main thread orchestrating — the
flow-pack rule. Where they are not (a lighter harness), implement the task's recipe directly:
the plan is self-contained (`flow-core/references/plan-format.md`) precisely so a cheaper executor can run
it. The in-vivo gate is **not** frontier-only — it runs here too; a teammate who only has this
harness still gates their work.

## Stages (read the reference for the stage you are in)

`$ARGUMENTS`: an optional `verify` subcommand plus an optional plan-or-part path. The default
run walks OPEN → Execute → Gate → CLOSE; `verify` resolves the plan at OPEN, then jumps
straight to the Gate (it assumes `built`) and closes.

| Stage | Read | When |
|---|---|---|
| OPEN — reconcile | `references/reconcile.md` | Every invocation: resolve the plan, adopt an organic one, read the pending state from git, decide in-vivo timing |
| Execute (`planned`/`building` → `built`) | `references/execute.md` | The reconciled state is `planned` or `building`; skipped by `verify` |
| Gate (`built` → `verified`) | `references/verify-gate.md` | The reconciled state is `built`, or the `verify` subcommand — two-stage review, in-vivo gate, design gate, integrated smoke |
| CLOSE | `references/close.md` | Every invocation, after the run's last stage |

Stable path after deploy: `~/.agents/skills/flow-build/references/<file>.md`.
