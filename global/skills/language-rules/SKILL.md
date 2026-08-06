---
name: language-rules
description: >
  Codex: load before code write/edit/review/debug/generate — full language conventions
  (TS/JS, React/Next, Angular, Nest, Python, Java/Kotlin, SQL/Prisma/Drizzle, Tailwind,
  shell, Docker/Terraform/GHA, UI). opencode gets language rules by glob via its plugin;
  Claude Code needs none of it — its rules load natively. Match files in play.
---

# language-rules — deterministic router to the full language conventions

The full, canonical rules live in `references/` (injected at build time from
`global/rules/` — single source of truth). This skill exists because Codex has no
glob-conditional rule loading: the COMPLETE rules load only when the stack is touched.

**Per-harness scope — read only what your harness lacks:**

| Harness | What this skill is for | Reference path |
|---|---|---|
| Codex | Everything below — no conditional rule channel exists | `references/<file>` |
| opencode | Browser + quality rows. Language rows arrive via the rules plugin | `references/<file>` |

## Routing table — read every row that matches the files/manifests in play

| Signal (file touched, manifest, or extension) | Read |
|---|---|
| `*.ts`, `*.tsx`, `*.js`, `*.jsx`, `tsconfig*.json` | `references/typescript-standards.md` |
| `next` in package.json, `next.config.*`, `app/` or `pages/` router files | `references/react-nextjs.md` + typescript |
| `angular.json`, `@angular/core` in package.json, `*.component.*` | `references/angular-patterns.md` + typescript |
| `nest-cli.json`, `@nestjs/core` in package.json | `references/nestjs-patterns.md` + typescript |
| `pyproject.toml`, `*.py` | `references/python-standards.md` |
| `pom.xml`, `build.gradle*`, `*.java`, `*.kt` | `references/java-kotlin.md` |
| `schema.prisma`, `drizzle.config.*`, migration dirs, raw `*.sql` | `references/sql-migrations.md` |
| Tailwind markers (`@import "tailwindcss"`, `@theme`, `tailwind.config.*`) | `references/tailwind.md` |
| `*.sh`, shell script edits | `references/shell-standards.md` |
| `Dockerfile*`, `*.tf`, `.github/workflows/*` | `references/iac-devops.md` |
| Building or styling UI (any stack) | `references/ui-visual-design.md` |
| Non-trivial implementation (new feature, refactor — any stack) | `references/development-principles.md` |
| Writing/modifying tests, or any behavior change | `references/testing.md` |
| Non-obvious bug: intermittent, multi-layer, or resists the first fix | `references/debugging.md` |
| Driving a browser / in-vivo verification of a running app | `references/browser-automation.md` |
| Code discovery/search (rg vs jbcontext vs codegraph routing, absence claims, index-hit verification) | `references/code-search.md` |

## Stacking — combinations are the norm, not the exception

- Angular app work → `typescript-standards` + `angular-patterns` (+ `ui-visual-design` if UI).
- Next.js app work → `typescript-standards` + `react-nextjs` + `tailwind` (if markers) + `ui-visual-design` (if UI).
- NestJS API work → `typescript-standards` + `nestjs-patterns` (+ `sql-migrations` if schema).

## Rules of use

- Route by what is ON DISK (manifests, extensions), never by what the prompt names — a
  "fix the validation bug" prompt reveals its stack only after inspection.
- Match against the files in play for THIS task, not every stack present in the repo — a
  one-file fix loads its own row(s) plus the quality rows its triggers actually fire.
- Role-scoped subagents load only their lens's rows: a reviewer/verifier reads the stack
  row(s) of the diff plus the quality row for its lens (testing for test gates,
  browser-automation for in-vivo) — never the full matching set. The dispatcher's handoff
  may name the rows already applied so they are not re-derived.
- A reference you already loaded this session (and not compacted away) does not need
  reloading.
- No matching row → this skill has nothing for the task; proceed without it.
- **Claude Code:** every row is already reaching you — language rows by `paths:` glob, the rest always-on. Do not load this skill.
- **opencode:** the language rows arrive automatically via the rules plugin — read only
  the quality/verification rows (development-principles, testing, debugging,
  browser-automation) from here.
