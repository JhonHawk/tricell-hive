---
name: monorepo-cutover
description: >
  Migrate a multi-repo product to a monorepo, or bootstrap a new project born in that
  format — intake decisions with the project owner, staged execution to QA, and a
  done-checklist that survives audits ("migrar a monorepo", "unificar los repos",
  "cutover", "nuevo proyecto formato monorepo"). Stack-aware: the JS/TS lane
  (Turborepo + pnpm) is fully specified; a polyglot mix (Go/Python/Java) forks at
  intake into an orchestrator decision instead. Subcommands: intake | execute | verify.
---

# /monorepo-cutover — replicable multi-repo → monorepo migration

Born from the ARK cutover (2026-08); the generalized playbook lives in
`references/cutover-playbook.md`. Read it in full before the first phase — every rule in
it was paid for by a real failure. Stable path after deploy:
`~/.agents/skills/monorepo-cutover/references/` (Claude Code, Grok) or
`~/.agents/skills/monorepo-cutover/references/` (Codex, opencode).

## Step 0 — the stack fork (before anything else)

Ask, never assume, what the monorepo will contain:

- **JS/TS only** (Node services, Next apps, shared TS contracts) → the playbook's lane
  applies end to end: Turborepo + pnpm workspaces, `turbo prune --docker`, remote cache.
- **Polyglot** (Go, Python, Java/Kotlin mixed in) → the orchestrator is an OPEN intake
  decision with the project owner: Bazel / Pants / Nx, or a turbo+make hybrid where the
  JS slice keeps the playbook and other languages get their own task runner. The
  playbook's SEQUENCE (decisions → hoist → contracts → images → CI → promotion) still
  governs; its tool table does not. Record the choice as an intake decision — never
  default it from this file.
- **Greenfield** (new project born monorepo) → skip the hoist; the playbook's layout,
  CI/CD shape, and done-checklist apply from day one. `/flow-start` stage C routes here.

## `intake` — close the decisions before moving code

1. Read `references/cutover-playbook.md > Decisions to close before moving code` and walk
   the table WITH the user (contracts resolution, image naming, registry auth, CD shape,
   CI stages, local compose, production out of scope). A row left open at execute time is
   a stop, not a guess.
2. Register the candidate in Engram — one observation per project, deterministic
   `topic_key: migration/<project>-monorepo-cutover`, upserted (never a new obs per
   session): decisions closed, stack lane, current phase, blockers. This is the
   cross-session tracking; there is no tracking agent.
3. Confirm scope: runtime repos in, specs/mocks/real secrets out; old remotes archive,
   never delete; production is explicitly a separate plan.

## `execute` — the staged sequence

Follow the playbook's execution order (hoist → workspace contracts → prune images → CI →
promotion → Turbo hygiene). Non-negotiables carried from the source case:

- Do not compact the image/CI steps — each one surfaces a different failure.
- QA green (health + a real login, and the container-image assert — a healthz 200 can be
  the OLD container) BEFORE archiving old remotes.
- The hygiene batch (`--affected`, sherif, env hash, remote cache) is a separate PR after
  QA, never mixed with the first deploy.
- Every phase close updates the candidate's Engram observation (same `topic_key`).
- Git/deploy gates unchanged: promotions confirm per `git-workflow.md`; nothing here is
  standing-authorized by the skill's invocation.

## `verify` — the done-checklist

Run `references/cutover-playbook.md > How to know the migration is done` item by item and
report per-criterion (done / blocked / not-reached — never a rounded-up "complete").
Update the Engram observation to its final state. "The repo exists" is not done; the
lockfile shape, the CD auth path, the deployed SHA, and the archived remotes are.

## Guardrails (learned, not theoretical)

- Never rename container images to match the new repo name — image names are
  `<project>-<component>`, decoupled from the git remote.
- A member PAT that "fixes" registry auth is a bridge, not a design — the workflow token
  with explicit registry write access is the end state.
- Recovery mechanisms (overlay, version pins) get DELETED after cutover, not kept "just
  in case" — leftovers silently become the default again.
- Reusable tokens (remote cache, registries) follow `security.md > Authentication &
  Secrets`: secret manager (`op`, else OS keychain), fetched once per session into an
  ephemeral `0600` copy.
