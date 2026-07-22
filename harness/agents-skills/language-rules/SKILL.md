---
name: language-rules
description: >
  Mandatory before any task that writes, edits, reviews, debugs, or generates code on a
  harness WITHOUT conditional rule loading (Codex — load this; opencode and Claude Code
  receive these rules automatically through their own conditional channels, do not load
  there). Covers TypeScript/JavaScript (.ts/.tsx/.js/.jsx, tsconfig), React/Next.js
  (next.config.*, app/ router), Angular (angular.json), NestJS (nest-cli.json,
  @nestjs/core), Python (pyproject.toml, .py), Java/Kotlin/Spring (pom.xml,
  build.gradle*), SQL migrations/Prisma/Drizzle (schema.prisma, drizzle.config, migration
  dirs), Tailwind CSS, shell scripts, Docker/Terraform/GitHub Actions, and UI visual
  craft — plus depth references for implementation principles, testing gates, debugging
  discipline, and browser-driven verification. Load after repository discovery even when
  the user does not name the stack; read every reference matching the files in play for
  THIS task — not the repo's whole stack inventory.
  opencode: language rules arrive via its rules plugin — load this skill only for the
  quality/verification references.
---

# language-rules — deterministic router to the full language conventions

The full, canonical language rules live in `references/` (injected at build time from
`global/rules/languages/` — single source of truth). This skill exists because Codex has
no glob-conditional rule loading: the always-on floor was removed in favor of loading
the COMPLETE rules only when the stack is actually touched.

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
- opencode: the language rows arrive automatically via the rules plugin — read only the
  quality/verification rows (development-principles, testing, debugging,
  browser-automation) from here.
