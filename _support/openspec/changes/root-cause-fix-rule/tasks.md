# Tasks

Base `development` en `6746ceee`, rama `feat/root-cause-fix-rule`. Todas las tareas corren en el hilo principal: son ediciones de texto pequeñas sobre archivos que se solapan (T1 y T2 escriben `flow-build/SKILL.md`), por debajo de los umbrales de delegación.

### T1 — Regla de arreglo en la causa y sus punteros

- [x] `flow-build/references/defect-fix.md` existe con la regla de AC7
- [x] `flow-build/SKILL.md` la enlaza en el paso 2 y en «Reuse before writing» (tarea o enfoque), y el peldaño 2 excluye copiar un parche
- [x] `flow-verify/SKILL.md` paso 3 la enlaza como excepción a «Fix only what each finding names»
- [x] `flow-deploy/references/urgent-fixes.md` la enlaza para la mitigación de un incidente

**Closes:** AC2, AC3, AC7

**Depends on:** none

**Locations:** `content/skills/flow-build/references/defect-fix.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-deploy/references/urgent-fixes.md`

**Execution:** main thread

**Test approach:** check

**Changes:** añade la referencia y sus punteros; no cambia «smallest sufficient» de `global.md` ni las reglas de la sección Tests más allá de un puntero.

**Verification:** `rg -n "defect-fix.md" content/` muestra los enlaces; `go test ./tests/content/` y `go test ./tooling/management/ -run 'Anchor|Reference'` pasan.

### T2 — Secciones situacionales a referencias

- [x] `flow-build/references/stored-data.md` y `security-boundaries.md` contienen las secciones tal cual, con su condición de lectura al inicio
- [x] `flow-build/SKILL.md` las enlaza con su condición y ya no las contiene, y ya no tiene la sección «Delegation» (D6 A)
- [x] Los 7 enlaces entrantes apuntan a los archivos nuevos
- [x] El comentario de `tooling/management/voice.go:220` sigue siendo cierto, o se reescribe

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-build/references/stored-data.md`, `content/skills/flow-build/references/security-boundaries.md`, `content/agents/design/hive-design-architecture.md`, `content/agents/development/hive-build-data.md`, `content/agents/ops/hive-build-infra.md`, `content/agents/review/hive-review-code.md`, `content/agents/review/hive-review-security.md`, `content/skills/flow-deploy/SKILL.md`, `content/skills/flow-deploy/references/deployment-checks.md`, `tooling/management/voice.go` (solo comentario)

**Execution:** main thread

**Test approach:** check

**Changes:** mueve el texto sin reescribirlo; solo cambia el título a `#` y añade la línea de condición.

**Verification:** `rg -n "#stored-data|#security-boundaries" content site` no devuelve nada; `wc -c content/skills/flow-build/SKILL.md` muestra al menos 1.000 bytes libres bajo 16.000; `go test ./tests/content/` y `go test ./tooling/management/ -run 'Anchor|Reference'` pasan.

### T3 — Ángulo de revisión de código

- [x] `hive-review-code.md` paso 2 incluye el arreglo en el síntoma y el código que copia un parche, y aclara que arreglar en la causa no es un cambio no pedido
- [x] Paso 3 dice cuál es el escenario de fallo de esos hallazgos

**Closes:** AC5

**Depends on:** none

**Locations:** `content/agents/review/hive-review-code.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una cláusula en la lista de ángulos y una en el filtro de hallazgos; no fija prioridad.

**Verification:** lectura de los pasos 2 y 3; `go test ./tests/content/` pasa.

### T4 — Sitio, índice de recursos y changelog

- [x] `site/src/content/docs/flows/build.md` describe el arreglo en la causa y el parche marcado, con enlace a `defect-fix.md`, y la línea de reutilización dice que no copia un parche
- [x] `site/src/content/docs/flows/verify.md` revisado; ajustado si describe el arreglo de la ronda
- [x] `_support/docs/architecture/instruction-resources.md` lista las tres referencias nuevas con su dueño y condición
- [x] `CHANGELOG.md` tiene la entrada bajo `[Unreleased]`

**Closes:** AC6

**Depends on:** T1, T2

**Locations:** `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** describe el comportamiento y enlaza la regla, sin copiarla.

**Verification:** `pnpm --dir site check`, `pnpm --dir site lint` y `pnpm --dir site build` pasan (`CONTRIBUTING.md:114-121`).

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | Los comandos de cada tarea | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...`, `go vet ./...` y el build del sitio | Covered by the implementation |
| Revisión de todo el comportamiento | Final candidate, locally | Lista de lugares comprobados (global, skills, referencias, roles, `site/`, `_support/docs/architecture/`, `tests/pilot/`) y tamaño de `global.md` sin cambio, para el PR | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` (D5 A) | The review answer |

## Checkpoint

- Updated: 2026-11-15
- Stage: build · 5
- Last commit: 665991d4
- Open findings: none
- Next: flow-verify: /code-review on candidate, whole-behavior review, then ready PR

### Evidence

- go vet ./... · clean · 665991d4
- go test ./... · all packages ok · 665991d4
- pnpm --dir site check/lint/build · 0 errors, 18 pages · 665991d4
- CI: only master-source-branch.yml (PRs to master); none runs on a PR to development

## Review status and progress

2026-11-15 — T1 · a5ec06a8 · unverified: none
2026-11-15 — T2 · a5ec06a8 · unverified: none
2026-11-15 — T3 · 0fa1456f · unverified: none
2026-11-15 — T4 · 665991d4 · unverified: none
