---
name: flow-start
description: >
  Stand up a NEW client project workspace from scratch — the greenfield wizard. Use when
  beginning a project that has no workspace yet: raw notes or a brief in hand, a client
  just signed, "empecemos el proyecto de <cliente>". Walks three dialogue stages — intake
  (produce the requirements doc), bootstrap (workspace + ledger + memory), foundation
  (naming, repos, contracts, CI) — adapting or skipping any stage already done (retrofit).
---

# /flow-start — greenfield project wizard

An interactive, plan-like wizard that stands a project up end to end, fusing three steps
into one conversation: **intake → bootstrap → foundation**. Each stage is a dialogue with at
most one consolidated question block; present each stage's plan like something the user
approves, not a form you fill silently. Follow the flow contract
(`~/.agents/skills/flow-core/SKILL.md`) — but this is the entry that *creates* the ledger, so
the OPEN "ledger must exist" precondition is what the wizard satisfies, not what it assumes.

**Retrofit, don't refuse.** Inspect what already exists before each stage and adapt: a
requirements doc present → skip intake; a ledger present → skip bootstrap; some repos present
→ foundation covers only the gaps. Refuse only when the workspace is **fully established**
(ledger + specs repo + repos all present) — that is maintenance, not a start: point to
`/flow-hygiene` and stop.

## Resolve group & project

`$1` = group (client/domain, lowercase), `$2` = project (kebab-case). Resolve in order:
explicit arguments → infer from cwd when it matches `…/projects/<group>/<project>` (confirm
the inference in one line before creating anything) → ask for whichever token is missing.

## Stages (read the reference for the stage you are running)

| Stage | Read | Produces | Skip when |
|---|---|---|---|
| A — Intake | `references/stage-a-intake.md` | `requirements-v2.md` + `open-questions.md` (via requirement-analyst); scorecard gate | A satisfactory requirements doc already exists |
| B — Bootstrap | `references/stage-b-bootstrap.md` | Workspace tree, `_support/PROJECT.md`, the AGENTS.md/CLAUDE.md ambient pair, session-lane permissions, memory + tracker/conventions | The ledger already exists (adopt it; fill missing fields) |
| C — Foundation | `references/stage-c-foundation.md` | Naming table, repo matrix, base contracts, schema baseline, CI/CD scaffolding — **no environments provisioned** | Repos, contracts, and CI already exist (covers only the gaps) |

Copy-ready artifacts Stage B writes live in `templates/` — `workspace-agents.md`,
`workspace-claude.md`, `sessions-permissions.json`. Stable paths after deploy:
`~/.agents/skills/flow-start/{references,templates}/<file>`.

## CLOSE

CLOSE per the flow contract: the ledger records what each stage produced (requirements doc,
tree, ledger, naming table path, repos, CI links; environments row = "pending — first deploy
runs via the promotion playbook when resources are confirmed"), the decisions taken, and the
`## Current handoff`. Report the created workspace and offer `/flow-specs init` as the next
step — the specs repo is where the accepted intake assumptions and the base contracts become
durable, versioned project memory.
