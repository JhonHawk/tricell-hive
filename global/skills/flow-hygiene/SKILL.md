---
name: flow-hygiene
description: >
  Audit and repair workspace health — git state, stray files, broken pointers; convention
  is suggested, not enforced. Use at milestone closes (epic done, pre-delivery) or whenever
  a workspace feels hard to navigate. audit reports proposed actions; apply executes the
  approved ones. Bringing a pre-pack project into the flow is /flow-adopt, not this skill.
argument-hint: "[audit | apply]"
disable-model-invocation: true
---

# /flow-hygiene — workspace health

The flow contract keeps flow-skill writes clean; conversational sessions are where disorder
accumulates. This skill is the compensating control: not a gate, an on-demand sweep guided by
judgment. Default subcommand: `audit`.

Adoption of a pre-pack workspace (specs-repo creation, tiering, full restructure) lives in
`/flow-adopt`; `audit` + `apply` here cover the lighter retrofit — ledger and ambient-pair
repair without specs-repo creation — whether the workspace is already in the flow or pre-pack.

## Subcommands (read the references for the one you are running)

| Subcommand | Read | Does |
|---|---|---|
| `audit` (default) | `flow-core/references/judgment-criteria.md` → `references/audit.md` | Dispatches workspace-custodian and saves the actions manifest; proposes, never executes |
| `apply` | `references/apply.md` | Gates the manifest by risk and executes only the approved actions |

Stable path after deploy: `~/.claude/skills/flow-hygiene/references/<file>.md`.
