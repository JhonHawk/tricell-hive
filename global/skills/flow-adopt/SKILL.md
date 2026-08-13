---
name: flow-adopt
description: >
  Bring an existing pre-pack project (current or past) into the flow: specs repo creation,
  activity tiering, per-artifact routing, ledger and ambient-pair reconstruction. Use when a
  project predates the pack — legacy workspace, no ledger, "metamos este proyecto al flow".
  /flow-start retrofits a bare codebase and points fully-established workspaces to
  /flow-workspace; a project with accumulated documentary material starts here. A lighter
  retrofit without specs-repo creation is /flow-workspace audit + apply. Produces a risk-gated
  migration manifest and executes only the approved actions.
argument-hint: "[<project-path>]"
disable-model-invocation: true
---

# /flow-adopt — bring a pre-pack project into the flow

One-time adoption, distinct from recurring hygiene: it may create the specs repo and
restructure the workspace. Mechanics: `references/adopt.md` (which reads
`flow-core/references/migration-playbook.md` — the source of truth — and judges artifacts
with `flow-core/references/judgment-criteria.md`).

Re-running is safe: already-conforming artifacts produce no findings.

Stable path after deploy: `~/.claude/skills/flow-adopt/references/adopt.md`.
