# naming.md — instantiated naming table template

Written by `/flow-start` (foundation stage) to `<project>-specs/conventions/naming.md`,
consumed by `devops-engineer`, the promotion walk (gate + `verify` audit, via
`promotion-playbook.md`), contract authors (system-designer, specs-stage spec sessions),
and any session that creates an infra resource or defines a cross-layer identifier. The generic rule lives in the global rule
`workflow/infra-naming.md`; this file is its **instantiation**: the concrete name of every
resource this project will have, plus the **code-layer boundary conventions** (API JSON
casing, DB casing, ORM mapping) that per-language idioms cannot resolve alone.

Why instantiate instead of deriving on the fly: an abstract rule evaluated per-session
produces divergent interpretations (`bo` vs `backoffice`, `media` vs `assets`). A table
written once produces the same name every time. The historical inconsistency this prevents
came precisely from naming resources at deploy-request time without a written table.

## Rules of use

- **Gate:** any new resource is validated against this table BEFORE creation. Not listed →
  add the row first (user sign-off if it needs an exception), then create.
- Names derive from the global rule using the **project token** recorded in PROJECT.md.
- Exceptions are documented in this file with their reason — an undocumented exception is
  indistinguishable from drift, which is exactly what the audit flags.
- The promotion `verify` walk (`promotion-playbook.md`) diffs reality (aws/hcloud/gh CLI
  listings) against this table and reports drift. Keep `status` current so the audit stays
  meaningful.
- **Code-layer conventions gate at spec-writing time:** an OpenAPI property, DTO field, or
  column defined in the wrong boundary casing is caught against this table at spec review
  — after implementation it costs a migration, not an edit. Per-language idiomatic casing
  is NOT recorded here (the global `languages/*` rules own it — restating it invites
  drift); only the boundaries where two stacks meet.

## Template

```markdown
# <project> — Infrastructure naming

Project token: `<project-token>` · Derived from global rule `workflow/infra-naming.md`
Environments: `development | qa | production` (full token, always last)

## Resource table

| Resource type | Template | Concrete name | Env | Status | Notes |
|---|---|---|---|---|---|
| VPS / server | `<token>-<env>` | <token>-qa | qa | exists | Hetzner project <name> |
| S3 bucket | `<token>-<content>-<env>` | <token>-media-production | production | planned | |
| IAM user | `<token>-<role>-iam-<env>` | … | | | |
| Secrets path | `<token>/<env>/<resource>` | … | | | |
| Subdomain (non-prod) | `<env-short>-<app>.<domain>` | qa-api.<domain> | qa | exists | DNS exception: short env prefix |
| Subdomain (prod) | `<app>.<domain>` | api.<domain> | production | planned | clean domain IS production |
| Database | `<token>_<env>` or invariant | … | | | see exceptions if invariant |
| GitHub repo | `<token>-<component>` | <token>-backend | — | exists | repos never carry env |
| CI/CD workflow env | `development\|qa\|production` | — | — | — | stage vocabulary only |

## Repo branch model

Classes and semantics: global rule `workflow/git-workflow.md > Branching`. This mapping is
what the promotion walk (`promotion-playbook.md`) reads to resolve promotion source/target
branches.

| Repo | Class | Long-lived branches | Branch → environment |
|---|---|---|---|
| <token>-backend | deployable multi-env | development (default) · qa · production | each branch deploys its same-named env |
| <token>-frontend | platform-native (Vercel) | main | main → production; PR previews |
| <token>-specs | specs/config | master | — |

## Code-layer conventions

Boundary decisions the per-language idioms cannot resolve alone, derived from the stacks
in the repo matrix. Idiomatic casing inside one language is not restated here (global
`languages/*` rules own it). Identifiers are always English; domain values may stay
Spanish per bounded domain (global `CLAUDE.md > Code Layer`).

| Boundary | Convention | Driven by |
|---|---|---|
| API JSON / DTO fields | camelCase | <producer + consumer, e.g. Nest API + React front> |
| DB tables & columns | snake_case, English | <e.g. PostgreSQL idiom> |
| ORM ↔ DB mapping | explicit (`@map` / `@Column({ name })`) — never the ORM's implicit naming strategy | camelCase props over snake_case columns |
| Event/queue payload keys | <casing> | <same ecosystem as the API, or broker convention> |

### Legacy → target mapping (only when porting a legacy surface)

| Legacy (source) | Target | Layer |
|---|---|---|
| <is_superadmin (JSON)> | <isSuperadmin prop ↔ is_superadmin column> | DTO ↔ DB |
| <nombre> | <name / key> | field rename (Spanish → English) |

## Project exceptions (sealed — each with its reason)

| Exception | Reason | Sealed on |
|---|---|---|
| <e.g. DB names invariant across envs> | <e.g. one Postgres per env per VPS — env lives in DATABASE_URL host> | YYYY-MM-DD |

## Untouchable legacies (do NOT rename)

| Resource | Why it stays |
|---|---|
| <name> | <migration cost exceeds benefit / live data / external references> |
```
