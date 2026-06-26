---
name: flow-foundation
description: >
  Create the technical foundation of a client project (F5 of the flow pack): instantiated
  naming table, repo matrix, initial schema, base API contracts, and CI/CD pipelines that
  exist BEFORE the first deploy. Use once per project, after specs and mock are settled
  and before development sessions begin.
argument-hint: ""
disable-model-invocation: true
---

# /flow-foundation — repos, contracts, pipelines

Follow the flow contract (`~/.claude/skills/flow-core/SKILL.md`). This phase **authors**
the infrastructure — repos, contracts, schema, CI, provisioning scripts, naming — but
does NOT provision environments: foundation usually starts before the client's
accounts/resources are confirmed, and coupling it to provisioning would block it on
exactly what the user doesn't control. Environment execution belongs to
`/flow-deploy setup <env>`, which runs whenever resources are approved. The golden rule
(**every environment's first deploy is performed by the pipeline**) lives there;
foundation's job is leaving everything ready so setup is a button-press, not a project.

Reference stack (confirm against specs, don't assume): NestJS APIs, Next.js frontends
(HeroUI/shadcn), PostgreSQL/MongoDB, Redis; AWS for buckets/email; Hetzner VPS for QA,
AWS for production.

## Phase 1 — Naming table

Instantiate `<project>-specs/conventions/naming.md` from the global rule
(`workflow/infra-naming.md`) using the project token in PROJECT.md and the template at
`~/.claude/skills/flow-core/references/naming-template.md`. Every resource this phase
will create gets its row BEFORE anything is created; client exceptions get documented
with their reason and the user's sign-off.

## Phase 2 — Repo matrix (gate)

Derive the repo list from the specs (backend, frontend, transactional services, the
existing mocks repo) and present it via AskUserQuestion: names per the naming table
(`<project>-<component>`), stack per repo, QA/prod targets. Nothing is created until the
user approves the matrix — repos are cheap to create and expensive to rename.

## Phase 3 — Parallel dispatch

Dispatch per the handoff protocol
(`~/.claude/skills/flow-core/references/handoff-protocol.md`), in parallel where
independent:

- **system-designer** — base OpenAPI contracts into `<project>-specs/contracts/`, derived
  from the reviewed epics. Consumer: every implementation agent in F6 builds against
  these exactly.
- **database-specialist** — initial schema + migration baseline in the backend repo(s),
  honoring DB naming from the table (including any invariant-name exception).
- **devops-engineer** — per repo: scaffold, CI (lint → typecheck → build → test, fail
  fast) running green from the first commit, the CD workflow **defined with its target
  environment parametrized** (ready to point at whatever env setup creates), and
  **scripted, idempotent** provisioning AUTHORED and versioned in
  `<repo>/_support/infrastructure/` — written now, executed later by
  `/flow-deploy setup` when resources are approved. Every resource name comes from the
  naming table, and every script step pairs its command with the expected output
  (`Run: <command> — Expected: <result>`) so setup detects drift instead of
  interpreting it.

Verify each agent's claims via Bash before accepting (contract lints, migrations run
locally, CI green on the nearly-empty repos).

## Phase 4 — Close

F5 closes **without a live environment** — that's deliberate. CLOSE per the contract:
ledger updated (F5 done; repos listed; naming table path; CI links; environments row =
"pending — run `/flow-deploy setup <env>` when resources are confirmed"). Development
sessions (`/flow-plan`) can start immediately against local services; the first
environment arrives via setup whenever the client side unblocks.
