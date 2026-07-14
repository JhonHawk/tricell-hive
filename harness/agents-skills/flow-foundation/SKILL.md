---
name: flow-foundation
description: >
  Create the technical foundation of a client project (F5 of the flow pack): instantiated
  naming table, repo matrix, initial schema, base API contracts, and CI/CD pipelines that
  exist BEFORE the first deploy. Use once per project, after specs and mock are settled
  and before development sessions begin.
---

# /flow-foundation — repos, contracts, pipelines

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This phase **authors**
the infrastructure — repos, contracts, schema, CI, provisioning scripts, naming — and
provisions NO environments: execution belongs to `/flow-deploy setup <env>`, run whenever
client resources are approved (the golden rule — **every environment's first deploy is
performed by the pipeline** — lives there). Foundation leaves everything ready so setup
is a button-press, not a project. Proportional: a single-repo project collapses Phases
2–3 into one derivation + one dispatch — the phases are seams, not ceremony.

Reference stack (confirm against specs, don't assume): NestJS APIs, Next.js frontends
(HeroUI/shadcn), PostgreSQL/MongoDB, Redis; AWS for buckets/email; Hetzner VPS for QA,
AWS for production.

## Phase 1 — Naming table

Instantiate `<project>-specs/conventions/naming.md` from the global rule
(`workflow/infra-naming.md`) using the project token in PROJECT.md and the template at
`~/.agents/skills/flow-core/references/naming-template.md`. Every resource this phase
will create gets its row BEFORE anything is created; client exceptions get documented
with their reason and the user's sign-off. The table includes the **repo branch model**
section: each repo's class and branch→environment mapping per
`workflow/git-workflow.md > Branching` — `/flow-deploy` resolves promotions from it. It
also includes the **code-layer conventions** section: boundary casing (API JSON/DTO, DB
tables/columns, ORM mapping, legacy→target mapping when porting) derived from the stacks
the specs settle — per-language idiomatic casing stays in the global `languages/*` rules,
never restated here.

## Phase 2 — Repo matrix

Derive the repo list from the specs (backend, frontend, transactional services, the
existing mocks repo): names per the naming table (`<project>-<component>`), stack per
repo, QA/prod targets, and repo class (deployable multi-env / platform-native /
specs-mocks / IaC) from the branch-model table. Names that derive cleanly from the Phase 1 table → **create and
report the matrix** — the table already carries the user's sign-off. Gate via
AskUserQuestion only on signal: a name needing a NEW naming exception, genuine ambiguity
in the repo split the specs don't settle, or creation inside a client-owned org (a wrong
repo there is outward-visible and expensive to rename).

## Phase 3 — Parallel dispatch

Dispatch per the handoff protocol
(`~/.agents/skills/flow-core/references/handoff-protocol.md`), in parallel where
independent:

- **system-designer** — base OpenAPI contracts into `<project>-specs/contracts/`, derived
  from the reviewed epics, honoring the code-layer conventions from the naming table
  (API JSON casing; identifiers English). Consumer: every implementation agent in F6
  builds against these exactly.
- **database-specialist** — initial schema + migration baseline in the backend repo(s),
  honoring DB naming from the table (including any invariant-name exception).
- **devops-engineer** — per repo: scaffold (including a minimal repo `AGENTS.md` whose
  conventions block POINTS at the naming table — infra names + code-layer conventions —
  never copies it), the branch model per the repo's class
  (deployable multi-env: `development` as default + `qa` + `production` created with the
  scaffold, protections on `qa`/`production`; specs/mocks: single trunk — never deferred
  to first deploy), CI (lint → typecheck → build → test, fail
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
