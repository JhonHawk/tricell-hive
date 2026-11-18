# Tasks

Base: tricell-hive `development` en `9f67fead` (2026-11-17). Rama `feat/project-layout-skills` en el checkout principal, sin worktree. T1, T2, T5–T7 corren en el hilo principal porque mueven y reescriben guía cuyo hogar canónico tiene que quedar único; T3 y T4 (scripts con sus tests) pueden ir a un hijo `hive-build-backend` con el brief de esta carpeta; T8 en el hilo principal porque toca `tooling/`. Comando de tests de Python del repositorio: `python3 -m unittest discover -s tests/skills -p '*_test.py'` (`CONTRIBUTING.md:23`).

### T1 — `project-layout` existe y todos sus consumidores apuntan a ella

- [x] `git mv content/skills/docs-cleanup content/skills/project-layout`; `SKILL.md` con `name: project-layout`, descripción que cubre convención, check, organizar y migrar, y el router de los tres modos; el procedimiento de limpieza se conserva entero.
- [x] `git mv content/skills/flow-close/references/artifact-layout.md content/skills/project-layout/references/layout.md`, ampliado con la raíz, el índice y su línea `Layout:` que declara las casas propias del proyecto y manda sobre los destinos por defecto, `data/`, el código fuera de `_support/` y la prohibición para el código de la aplicación; conserva el encabezado «Retain evidence» y su ancla.
- [x] `references/script-lifecycle.md`: dueño de los niveles 0, 1 y 3; los niveles 2 y 4 enlazan a flow-close «Candidates» y a defect-fix punto 5. flow-close «Candidates» suma un solo enlace a él.
- [x] Actualizar en el mismo commit cada consumidor de `artifact-layout` y `docs-cleanup`: `content/guidance/global.md` (enlace y la excepción de AC9 en una frase), `content/guidance/README.md`, `content/agents/**`, `content/skills/**`, `AGENTS.md:53`, `_support/docs/harness-engineering/rule-enforcers.md:25` (ruta de `archive_sessions.py`), `_support/docs/architecture/instruction-resources.md`, `tests/content/reading_footprint_test.go` (nuevas rutas en `added`), `tests/skills/workspace_archive_test.py`, `tests/fixtures/workspace-conventions/README.md`; en `tests/pilot/session_audit.go:1237` se conserva `docs-cleanup` en la regex de transcripts y se suman `project-layout|workspace-repo`.
- [x] `CHANGELOG.md` `[Unreleased]`: la entrada de `docs-cleanup`, aún no publicada, pasa a describir `project-layout` y se agrega `workspace-repo`; sin nota de migración porque `docs-cleanup` nunca salió en un release.

**Closes:** AC1, AC2, AC3, AC9, AC13

**Depends on:** none

**Locations:** `content/skills/project-layout/`, `content/skills/flow-close/`, `content/guidance/`, `content/agents/`, `AGENTS.md`, `CHANGELOG.md`, `_support/docs/**`, `tests/**`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en `global.md` solo el enlace y la frase de la excepción; el PR informa su tamaño antes y después. La retención de evidencia se mueve sin cambios de fondo.

**Verification:** `go vet ./...` y `go test ./...` pasan (enlaces, tamaño ≤16.000 B, frontmatter, footprint, rule-enforcers); `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa; `rg -l 'artifact-layout' content site/src tests` vacío; `rg -n 'data/|script-lifecycle|README.md' content/skills/project-layout/references/layout.md` muestra cada destino de AC2; `rg -n 'Layout:' content/skills/project-layout/references/layout.md` tiene un resultado.

### T2 — El layout sale de harness-audit

- [x] Mover HA-PL-01…05 y el manifiesto de migración a `content/skills/project-layout/references/migrate.md`, con la pregunta del hub de cuatro opciones, la propuesta de `_support/docs` (D16) y la escritura de la línea `Layout:` del índice cuando el usuario la declina, el test de viabilidad y la parte compartida de una migración (texto en `design.md` «Insumos copiados de memoria»).
- [x] Borrar `harness-audit/references/project-layout.md`. Su paso 4 apunta a `project-layout`; el modo de onboarding (`SKILL.md:35`) lee HA-PL-01 en `migrate.md`; se corrige «workspace section, which `hive doctor` does not check» para el caso del coordinador; la descripción de harness-audit y `content/agents/review/hive-review-harness.md:3` dejan de ofrecer preparar la migración de layout y nombran `project-layout`.
- [x] `tests/skills/harness_audit_rules_test.py`: `DEPLOYED` suma `content/skills/project-layout`; las filas HA-PL de `harness-audit-rules.md` siguen activas y su columna de casa dice `project-layout`.

**Closes:** AC6, AC7, AC13

**Depends on:** T1

**Locations:** `content/skills/harness-audit/`, `content/agents/review/hive-review-harness.md`, `content/skills/project-layout/references/migrate.md`, `_support/docs/harness-engineering/harness-audit-rules.md`, `tests/skills/harness_audit_rules_test.py`

**Execution:** hilo principal

**Test approach:** check

**Changes:** las demás reglas de harness-audit no cambian (D18 A).

**Verification:** `rg -n 'Layout:' content/skills/project-layout/references/migrate.md` tiene un resultado; `rg -n 'HA-PL-0[2-5]' content` solo en `project-layout`, y `HA-PL-01` además en el modo de onboarding como puntero; `go test ./tests/content/...` y el comando de Python pasan.

### T3 — `check_layout.py`

- [x] Script con las siete reglas de `design.md` «check_layout.py», que busca el índice en el orden definido y lee su línea `Layout:` antes de aplicarlas, y su test con fixtures temporales: un caso por regla, uno por exención de la regla 3, uno con solo `docs/README.md` y `Layout: docs at docs/ · tools at scripts/`, el mismo sin la línea (regla 7), uno con `Index:` en `AGENTS.md`, y un fixture limpio.
- [x] Corrida de solo lectura sobre cam, carvento, fleetsystem y facti; registrar en el checkpoint los hallazgos contra la lista de «Verified context» de `proposal.md`.

**Closes:** AC5, AC13, AC15

**Depends on:** T1

**Locations:** `content/skills/project-layout/scripts/check_layout.py`, `tests/skills/layout_check_test.py`

**Execution:** hijo `hive-build-backend` o hilo principal

**Test approach:** tdd

**Changes:** no lee `secrets/` ni el contenido de `data/`; no escribe en los proyectos medidos.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'layout_check_test.py'` pasa y, revertida la rama de una regla, su caso falla; las cuatro corridas reales reportan lo de AC5.

### T4 — `workspace-repo`

- [x] `content/skills/workspace-repo/SKILL.md` con el procedimiento de `design.md`, las preguntas del `Tracker` y de `Delivery: direct-base`, el aviso sobre la clave del viewer antes de agregar `origin`, el ascenso del repo de specs como paso manual, y sus limitaciones declaradas.
- [x] `scripts/workspace_repo.py` (plan por defecto, `--apply` con `--tracker` y `--direct-base` explícitos, modo setup que clona lo que falta de la tabla) y su test con un workspace temporal de dos hijos, un repo de specs y un `HIVE.local.md`.

**Closes:** AC8, AC10

**Depends on:** T3

**Locations:** `content/skills/workspace-repo/`, `tests/skills/workspace_repo_test.py`

**Execution:** hijo `hive-build-backend` o hilo principal

**Test approach:** tdd

**Changes:** nunca crea remotos ni hace push; no toca ningún workspace real.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'workspace_repo_test.py'` pasa, cubriendo el plan sin escritura, el `git status` limpio de hijos, la búsqueda desde la raíz, `check_layout.py` en 0 y `go run ./tooling/cli doctor` (o el binario construido desde el checkout) sin hallazgos bloqueantes sobre esa misma salida.

### T5 — Documentación de arquitectura y del repo

- [x] `_support/docs/architecture/workspace-and-artifacts.md` (layouts, coordinador, ciclo de vida, por qué la convención vuelve a una skill propia), `repository-and-distribution.md` si lista skills, `rule-enforcers.md` (filas de `check_layout.py` junto a las de `hive doctor`).
- [x] `llms.txt` nombra `project-layout` y `workspace-repo` (`README.md` solo lista flows).

**Closes:** AC14

**Depends on:** T1, T2, T4

**Locations:** `_support/docs/**`, `llms.txt`

**Execution:** hilo principal

**Verification:** `go test ./tests/content/...` (incluye `rule_enforcers_test`) pasa; `rg -l docs-cleanup llms.txt AGENTS.md` vacío.

### T6 — Sitio

- [x] `site/src/content/docs/skills.md`, `index.mdx` y `flows/close.md`: las dos skills y los enlaces nuevos; describen el comportamiento y enlazan la skill, sin copiar reglas.

**Closes:** AC14

**Depends on:** T1, T4

**Locations:** `site/src/content/docs/**`

**Execution:** hilo principal

**Verification:** `cd site && pnpm run check && pnpm run lint && pnpm run build` pasa; `rg -l docs-cleanup site/src` vacío.

### T7 — Delta y coherencia del cambio

- [x] Revisar el delta `specs/project-layout/spec.md` contra lo construido.

**Closes:** AC2, AC7, AC8, AC9, AC15

**Depends on:** T1, T2, T3, T4

**Locations:** `_support/openspec/changes/project-layout-skills/`

**Execution:** hilo principal

**Verification:** `python3 content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/project-layout-skills` sale 0.

### T8 — Viewer y doctor con un coordinador

- [x] Test de caracterización de `timelineProjectKey` para un workspace con trabajos guardados: la clave no cambia con `git init` sin remoto y cambia, con 409 `project_identity_changed`, al agregar `origin`.
- [x] Test de `hive doctor` con un coordinador cuyo `.gitignore` incluye `HIVE.local.md` y cuyos hijos están ignorados.

**Closes:** AC10

**Depends on:** none

**Locations:** `tooling/viewer/*_test.go`, `tooling/cli/doctor_project_test.go`; código solo si el test muestra una falla, con delta MODIFIED de `walkthrough-viewer`

**Execution:** hilo principal

**Test approach:** characterization, y tdd si aparece una falla

**Changes:** ningún cambio de comportamiento si los tests pasan sobre el código actual.

**Verification:** `go test ./tooling/...` pasa; si se cambió código, el test falla sobre el código anterior.

### T9 — Artifact «Skills de Hive»

- [x] Publicar la versión con `project-layout` y `workspace-repo` cuando el PR se abre, como estado propuesto, y marcarla integrada cuando el usuario lo integre.

**Closes:** AC12

**Depends on:** T5

**Locations:** artifact en claude.ai

**Execution:** hilo principal

**Verification:** `Artifact read` del URL muestra los dos nombres y ningún `docs-cleanup` vigente.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `cd site && pnpm run check && pnpm run lint && pnpm run build` | Covered by the implementation |
| Real-path check | Final candidate, read-only on cam, carvento, fleetsystem, facti | `check_layout.py` contra la lista de «Verified context» | Covered by the implementation (solo lectura) |
| Installed hosts | Final candidate, locally | `hive update --source <checkout> --rev <head> --dry-run` lista `docs-cleanup` como retirado | Covered by the implementation (sin aplicar) |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |

Sin parada de validación humana: la entrega es `ready`, que se detiene con el PR abierto para que el usuario lo integre.

## Checkpoint

- Updated: 2026-11-18
- Stage: verify · ready stop
- Last commit: 610f7765
- Open findings: none
- Next: user merges PR #508; then close: archive change, refresh install, artifact to integrated

### Evidence

- AC5 read-only check_layout on 16 real roots · cam 10 LAYOUT-01; carvento-monorepo 46 LAYOUT-02; facti-2020-new 65, facti-backend-client 16 LAYOUT-02; facti LAYOUT-03 cancel-payment-complements.js and global-emision.js; fleetsystem evidence/ and migration-data/ LAYOUT-04, no LAYOUT-01 · fc8099f6
- before captures: none (no UI)
- full suite: go vet ok, go test ok except known flake TestE2ESetupDeadlinesOwnDescendantsAndPipes, python unittest ok, site check/lint/build ok · 610f7765
- /code-review: 10 findings fixed in f994d453 · plan reviews 4 rounds applied
- no PR CI on development (workflow only for master)

## Review status and progress

2026-11-18 — Revisión: AC13 (línea `Layout:`) suma a T1, T2 y T3.
2026-11-18 — Plan escrito y revisado: T1 y T2 originales unidas en T1 para que los tests no se rompan entre tareas; numeración nueva.
2026-11-18 — T8 · 220bb97c · unverified: none
2026-11-18 — T1 · fd85bd49 · unverified: none
2026-11-18 — T2 · fd85bd49 · unverified: AC7 behavior in a real session pending observation
2026-11-18 — T5 · 6a2f43aa · unverified: none
2026-11-18 — T6 · 6a2f43aa · unverified: site lint and build run with the full suite
2026-11-18 — T3 · fc8099f6 · unverified: none
2026-11-18 — T4 · fc8099f6 · unverified: none
2026-11-18 — T7 · 001d9b61 · unverified: none
2026-11-18 — T9 · uncommitted · unverified: artifact shows the PR as open; mark it integrated after the merge
