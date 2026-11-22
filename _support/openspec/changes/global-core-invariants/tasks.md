# Tasks

Base: `development` en `c61c10e3`. T1 y T2 corren en el hilo principal, en ese orden, porque T2 cita las hojas de T1. T3 va a un hijo `hive-write-docs` porque edita más de 3 archivos y no se cruza con T4. T4 corre en el hilo principal.

### T1 — Las hojas reciben el procedimiento

- [x] `plan/proportionality.md` y `close/close-cleanup.md` existen con su texto movido y su línea del índice; `scope-changes` tiene la sección «An external project» y su primera línea nombra ese momento.

**Closes:** AC2, AC4, AC5

**Depends on:** none

**Locations:** `content/skills/hive-core/plan/proportionality.md` y `content/skills/hive-core/close/close-cleanup.md` (nuevos), `content/skills/hive-core/plan/scope-changes.md`, y en `content/guidance/global.md` solo las líneas del índice de las dos hojas nuevas y la de `scope-changes`

**Execution:** main thread

**Test approach:** check

**Changes:** como `design.md` «Moves» y «Leaf size»; el texto se copia sin cambios. Las líneas del índice van en esta tarea para que el test de hojas pase.

**Verification:** `go test ./tests/content/ -run 'Leaf|Index' -count=1` ok (tamaños, primera línea «Read», una línea de índice por hoja).

### T2 — El núcleo conserva solo el invariante

- [x] `global.md` tiene las frases de «Moves» y mide 26.000 B o menos.

**Closes:** AC1, AC4, AC5, AC6

**Depends on:** T1

**Locations:** `content/guidance/global.md` (incluida la cita de `:21`), `tests/content/single_home_test.go`, `tests/content/reading_footprint_test.go` (lecturas del escenario plan/build)

**Execution:** main thread

**Test approach:** check

**Changes:** como `design.md` «Moves»; `single_home_test.go` registra una frase de cada movimiento con su casa nueva: de la limpieza, la de los registros de trabajo pendiente (no la de procesos, que se queda en el núcleo); del cambio mecánico; de la tarea operativa; del proyecto externo.

**Verification:** `wc -c content/guidance/global.md` ≤ 26000; `go test ./tests/content/... -count=1` ok.

### T3 — Las citas apuntan a la hoja nueva

- [x] Cada cita de una definición movida apunta a `proportionality.md`.

**Closes:** AC2, AC3, AC4, AC5

**Depends on:** T2

**Locations:** `tests/content/leaf_links_test.go`, y los archivos de `content/` que lista `rg -l "«Proportionality»" content`, salvo `global.md`, y los de `design.md` «Citations»: `defect-fix.md`, `close-hygiene.md`, `flow-close/SKILL.md`, `flow-close/references/branch-cleanup.md`, `flow-build/SKILL.md`, `delivery-ask.md` y `hive-mode/SKILL.md`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** como `design.md` «Citations»; `tests/content/leaf_links_test.go` registra `close-cleanup` en `flow-close/SKILL.md` y `proportionality` en `flow-build/SKILL.md`.

**Verification:** el `rg` de AC3; `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRealStagePacketsKeepAMarginUnderTheCap|TestRepositoryCatalogueInstructionReferences' -count=1` ok.

### T4 — La documentación nombra las casas nuevas

- [x] `instruction-resources.md` y `CHANGELOG.md` nombran `proportionality.md`, `close-cleanup.md` y la nueva ubicación del proyecto externo.

**Closes:** AC2, AC4

**Depends on:** T2

**Locations:** `_support/docs/architecture/instruction-resources.md` (incluida la sección «Scope changes» que cita en `:30` y no existe), `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** las referencias a esas secciones del núcleo apuntan a su hoja; una entrada en `[Unreleased]`. `README.md:407` y `rule-enforcers.md:57` siguen siendo ciertos y no cambian.

**Verification:** lectura del diff; `go test ./tests/content/... -count=1` ok.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./... -count=1` and `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| UI check | n/a: no screen changes | — | — |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | The session's review answer |

## Checkpoint

- Updated: 2026-11-22
- Stage: verify · integrating
- Last commit: c61c10e3
- Open findings: /code-review: 4 important (citations, flow-close step 2 link, reading footprint cycle) and 4 nits fixed
- Next: push, PR, merge, deploy, close
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- go test ./tests/content/... · ok after T2
- global.md · 25856 B
- go test ./... · ok
- tests/skills · 412 OK
- go test ./... · ok final

## Review status and progress

2026-11-22 — plan escrito.
2026-11-22 — T1 · uncommitted · unverified: none
2026-11-22 — T2 · uncommitted · unverified: none
2026-11-22 — T4 · uncommitted · unverified: none
2026-11-22 — T3 · uncommitted · unverified: none
