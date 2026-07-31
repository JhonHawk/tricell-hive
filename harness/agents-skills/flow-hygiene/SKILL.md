---
name: flow-hygiene
description: >
  Audit and repair workspace health — git state, stray files, broken pointers; convention
  is suggested, not enforced. Use at milestone closes (epic done, pre-delivery) or whenever
  a workspace feels hard to navigate. audit reports proposed actions; apply executes the
  approved ones; migrate is the opt-in path for bringing a pre-pack project into the flow.
---

# /flow-hygiene — workspace health

The flow contract keeps flow-skill writes clean; conversational sessions are where disorder
accumulates. This skill is the compensating control: not a gate, an on-demand sweep guided by
judgment. Default subcommand: `audit`.

It is also the **adoption path for pre-pack workspaces**: `/flow-start` refuses fully-established
workspaces by design, so a project that predates the pack enters the flow through `migrate` here
(or `audit` + `apply` for a lighter retrofit) — reconstructing the ledger and the ambient pair,
after which any phase gate (`/flow-plan`, `/flow-build`) works normally.

## Subcommands (read the references for the one you are running)

| Subcommand | Read | Does |
|---|---|---|
| `audit` (default) | `references/judgment-criteria.md` → `references/audit.md` | Dispatches workspace-custodian and saves the actions manifest; proposes, never executes |
| `apply` | `references/apply.md` | Gates the manifest by risk and executes only the approved actions |
| `migrate` | `references/judgment-criteria.md` → `references/migrate.md` (+ `flow-core/references/migration-playbook.md`, the source of truth) | Brings a pre-pack project into the flow: specs repo, tiering, artifact routing |

Stable path after deploy: `~/.agents/skills/flow-hygiene/references/<file>.md`.
