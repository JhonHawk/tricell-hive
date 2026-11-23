# Tasks

Base `development` en `0aad427f`; worktree `.claude/worktrees/vendored-exclusion`, rama `feat/vendored-exclusion`. Las dos tareas corren en el hilo principal: T1 cambia una función y un test, T2 edita cuatro archivos de texto.

### T1 — `--same-code` rechaza cambios en rutas excluidas

- [x] `sameCode` reporta un cambio bajo una ruta de `EXCLUDED`, con los casos nuevos de `design.md` vistos fallar antes del cambio.

**Closes:** AC1

**Depends on:** none

**Locations:** `content/skills/audit-comments/assets/ratchet/check-comments.mjs`, `content/skills/audit-comments/assets/ratchet/check-comments.test.mjs`

**Execution:** main thread

**Test approach:** tdd

**Changes:** La función `excludedChanges`, su reporte al principio de `sameCode` y los cuatro casos de la suite.

**Verification:** `go test -count=1 -v ./tests/content/ -run TestAuditCommentsTemplate` falla con los casos nuevos antes del cambio del script y pasa después; la salida muestra que corrieron los subtests `babel8` y `babel7` y nombra los casos nuevos.

### T2 — El texto de la skill y sus referencias

- [ ] Los pasos 1, 2, 4 y 7 de `SKILL.md`, `ratchet.md`, `project-skill.md`, el nombre del test de Go y `CHANGELOG.md` dicen lo que fija `design.md` «Texto».

**Closes:** AC2, AC3

**Depends on:** T1

**Locations:** `content/skills/audit-comments/SKILL.md`, `content/skills/audit-comments/references/ratchet.md`, `content/skills/audit-comments/references/project-skill.md`, `_support/docs/architecture/instruction-resources.md`, `tests/content/testdata/audit-comments/package.json`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Los de `design.md` «Texto». Antes de editar, buscar en `content/`, `site/`, `_support/docs/architecture/` y `tests/pilot/` cada lugar que nombre `EXCLUDED`, `--same-code` o el código vendorizado, y listarlos en el PR.

**Verification:** `go test ./...` pasa; `rg -c 'changed under an' content/skills/audit-comments/references/ratchet.md` da 1 o más; `rg -c 'vendored' content/skills/audit-comments/SKILL.md` da 1 o más; `rg -c '0600' content/skills/audit-comments/SKILL.md` da 1 o más; `rg -c 'record commit' content/skills/audit-comments/references/project-skill.md` da 1; `rg -c 'comment_audit_template_test' _support/docs/architecture/instruction-resources.md tests/content/testdata/audit-comments/package.json` no da coincidencias.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` | Covered by the implementation |
| Code review | Final candidate, before the push | `Native: Claude Code /code-review` | Review in force |

AC1 se observa en la suite del candidato; AC2 y AC3, en los archivos instalados tras el merge (marcador de mantenedor presente).

## Checkpoint

- Updated: 2026-11-23
- Stage: build · T1
- Last commit: 0aad427f
- Open findings: none
- Next: T1: tdd tests then excludedChanges
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/vendored-exclusion · feat/vendored-exclusion
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no UI)

## Review status and progress

2026-11-23 — T1 · 6f8cf47f · unverified: none
