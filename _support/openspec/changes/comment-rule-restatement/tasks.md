# Tasks

Base `development` en `f550f3e6`; worktree `.claude/worktrees/comment-rule-restatement`, rama `feat/comment-rule-restatement`. Las dos tareas corren en el hilo principal: editan seis archivos de texto y una fila de un test.

### T1 — El criterio de borrado tiene una sola casa

- [ ] El Kill y la línea de reporte de `no-comments.md` y la Tautology de `audit-comments/SKILL.md` dicen lo que fija `design.md` «Texto», y la fila nueva de `single_home_test.go` fija la cláusula en el leaf.

**Closes:** AC1, AC2

**Depends on:** none

**Locations:** `content/skills/hive-core/review/no-comments.md`, `content/skills/audit-comments/SKILL.md`, `tests/content/single_home_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** La viñeta Kill, la línea de reporte, la viñeta Tautology y una fila de `homePhrases`. El paso 4 de `no-comments` no cambia.

**Verification:** el text check de `AGENTS.md` «Checks by path» pasa, con `TestRealStagePacketsKeepAMarginUnderTheCap` bajo el tope; `go test -count=1 -run 'TestRulesHaveOneHome' ./tests/content/` pasa; `rg -c 'restates a rule the code' content/` da una sola coincidencia, en `no-comments.md`; `rg -c 'gets a detail wrong' content/` no da coincidencias; `rg -c 'restated-rule Kill' content/skills/hive-core/review/no-comments.md` da 1.

### T2 — La auditoría deja la regla de comentarios en el AGENTS.md del proyecto

- [ ] Los pasos 9 y 10, la descripción, `project-skill.md`, `instruction-resources.md`, la página de skills del sitio y `CHANGELOG.md` dicen lo que fija `design.md` «Texto».

**Closes:** AC3, AC4

**Depends on:** T1

**Locations:** `content/skills/audit-comments/SKILL.md`, `content/skills/audit-comments/references/project-skill.md`, `_support/docs/architecture/instruction-resources.md`, `site/src/content/docs/skills.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** La plantilla y su viñeta en «Register it», el puntero del punto 5, las frases de los pasos 9 y 10 y de la descripción, la línea de la arquitectura, la del sitio y la entrada del changelog. `## Skills` y el chequeo del registro no cambian.

**Verification:** el text check pasa; `pnpm check`, `pnpm lint` y `pnpm build` en `site/` pasan y `rg -c 'comment rule' site/dist/skills/index.html` da 1 o más; en `content/skills/audit-comments/references/project-skill.md`, `rg -c` da 1 o más para cada una de `instead of adding another`, `declared workspace does not get it`, `HA-HI-03`, `paths the lint checks`, `until one exists, an entry in` y `section of .AGENTS.md. for the workaround marker`; `rg -c 'comment rule written to its' content/skills/audit-comments/SKILL.md` da 1.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | `Native: Claude Code /code-review` | The review answer |

AC1–AC3 se observan en los archivos instalados tras el merge y el refresco local (marcador de mantenedor); AC4 en el build del sitio del candidato.

## Checkpoint

- Updated: 2026-11-23
- Stage: build · T1
- Last commit: f550f3e6
- Open findings: none
- Next: T1: Kill clause, report line, Tautology pointer, single-home row
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/comment-rule-restatement · feat/comment-rule-restatement
- Rules: flow-build/stage.md

### Evidence

- no before captures (no screen)

## Review status and progress
