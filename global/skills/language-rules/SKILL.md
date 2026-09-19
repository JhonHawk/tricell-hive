---
name: language-rules
description: >
  Codex/Grok/opencode/PI: load before code write/edit/review/debug/generate — full language
  conventions (TS/JS, React/Next, Angular, Nest, Python, Java/Kotlin, SQL/Prisma/Drizzle,
  Tailwind, shell, Docker/Terraform/GHA, UI, patterns, devops). Claude Code: only the browser
  row (`browser-automation-reference.md` — the CLI mechanics are not always-on). Match files
  in play. Exception on every harness: naming identifiers BEFORE the file exists (new fields, enums, table/spec properties) fires
  nothing — load this skill and read identifier-language.md before choosing the names.
---

# language-rules — deterministic router to the full language conventions

The full, canonical rules live in `references/` (injected at build time from
`global/rules-situational/` and `global/rules/` — single source of truth). This skill exists
because no harness loads these rules by file kind: they are not deployed to any harness's
conditional-rule directory.

**Per-harness scope — read only what your harness lacks:**

| Harness | What this skill is for | Reference path |
|---|---|---|
| Codex | Everything below — no conditional rule channel exists | `references/<file>` |
| PI | Everything below — Hive exposes references through skills; the `rule-delivery` hold names the same files | `references/<file>` |
| Grok | Everything below — it receives the always-on rules and nothing else | `references/<file>` |
| opencode | Everything below — its rules plugin is retired, this skill is the only channel | `references/<file>` |
| Claude Code | Only the browser row — `browser-automation-reference.md`. Language rows arrive through the `rule-delivery` hold on a matching write | `references/<file>` |

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
| `Dockerfile*`, `*.tf`, `.github/workflows/*` | `references/iac-devops.md` + `references/devops-principles.md` |
| Building or styling UI (any stack) | `references/ui-visual-design.md` |
| Non-trivial implementation (new feature, refactor — any stack) | `references/development-principles.md` + `references/patterns-antipatterns.md` |
| Writing/modifying tests, or any behavior change | `references/testing.md` |
| Non-obvious bug: intermittent, multi-layer, or resists the first fix | `references/debugging.md` |
| Driving a browser / in-vivo verification of a running app | `references/browser-automation-reference.md` (full CLI mechanics) + `references/browser-automation.md` (the always-on gate stub, if not already in context) |
| Code discovery/search (routing by operation, absence claims, hit verification) | `references/code-search.md` |
| Naming identifiers with no file yet on disk — new fields, enums, table/column names, spec/API properties | `references/identifier-language.md` |

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
- **`rule-delivery` holds your first matching write** on Claude Code, Grok, Codex and PI,
  naming the reference to read — reading it from here first skips the hold. On Codex the
  hold does not re-arm after a compaction; opencode has no hook channel at all.
- **Claude Code:** the quality rows are always-on and the language rows arrive through that
  hold. Load this skill unprompted only for `browser-automation-reference.md` (the CLI
  mechanics behind the always-on gate stub).
- **Grok:** always-on rules arrive via `~/.grok/rules/`; the language/quality/devops rows do
  not — read every matching row below from `references/`.
- **opencode:** every matching row is read from here — the rules plugin that once delivered
  the language rows is retired.
- **Naming identifiers has no write to hold:** the pre-file naming row applies on every
  harness — the always-on gate states the rule; `identifier-language.md` carries the
  judgment (domain translation, false friends) that the gate alone has been measured to miss.
