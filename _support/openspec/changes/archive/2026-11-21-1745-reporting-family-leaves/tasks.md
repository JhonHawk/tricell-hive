# Tasks

Base: `development` en `2376cbbb`; rama local `refactor/reporting-family-leaves` (entrega `hold`: sin commits). Todas las tareas corren en el hilo principal: editan las mismas hojas y enlaces, así que no son independientes, y la fusión de reglas necesita el inventario de T1 a la vista. Ninguna tarea es riesgosa según «Proportionality» (no toca contratos públicos, datos, seguridad ni dependencias): no hay `hive-verify-task`.

Restricción de todas las tareas: `wc -c content/guidance/global.md` no crece sobre su tamaño en la base del build (28,593 B en `7f6277e0`; tope de la prueba 30,000 B), porque `hive-core-stage-packets` cuenta con ese margen; cada hoja ≤ 4,000 B salvo las excepciones de `tests/content/leaf_skill_test.go` (≤ 7,000 B); cada playbook ≤ 7,000 B.

### T1 — Inventario de oraciones y su destino

- [x] `inventory.md` en esta carpeta lista, oración por oración, cada regla de estas fuentes con su archivo y línea de origen y su destino (archivo y sección o fila), o «se borra: duplicado de <destino>», y estima los bytes de `report-endings` y `close-report` resultantes. Fuentes: `hive-rule-next-step-block`, `hive-rule-close-footer`, `hive-rule-report-endings`, `hive-rule-close-report`, la columna «Ends with» de `hive-rule-close-modes`; `git-workflow` pasos 9–10; `flow-deploy` pasos 6 y 8 y la línea 18; `flow-close` descripción y pasos 6 y 8; `unattended-delegation` líneas 15-16, 20 y 28; `hive-rule-merge-gate` líneas 9 y 14; las copias de la continuación (`hive-mode/SKILL.md:14,25-26`, `flow-build/SKILL.md:14,20`, `flow-plan/references/plan-handoff.md:21,28`); `global.md` «Questions and routes» (líneas 133-137) y las 62 líneas de «Rule leaves».

**Closes:** AC9

**Depends on:** none

**Locations:** `_support/openspec/changes/reporting-family-leaves/inventory.md`

**Execution:** hilo principal

**Changes:** solo el inventario; ningún archivo de `content/`.

**Verification:** cada oración de cada fuente aparece una vez en el inventario (se cuenta por oración, separando en `. `, `; ` y viñetas); las estimaciones dicen si `report-endings` cabe en 7,000 B y si `close-report` pasa de 4,000 B.

### T2 — `close-report` como hogar del reporte final

- [x] `hive-rule-close-report` tiene la tabla de variantes de `design.md` con la fila de despliegue completa, la sección «Findings at the close» con la tabla de hallazgos, su ronda y «Find the next ticket» que venían de `close-footer`, y su línea de lectura cubre toda variante; «Findings» conserva solo la definición y las prioridades.

**Closes:** AC1

**Depends on:** T1

**Locations:** `content/skills/hive-rule-close-report/SKILL.md`, `tests/content/leaf_skill_test.go` (solo si pasa de 4,000 B), `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check — `single_home_test.go` fija la frase que abre la tabla de variantes en `close-report`.

**Changes:** se conservan las frases que fija `single_home_test.go:49-50` y el ancla `#findings` que leen `content/agents/review/*` y `global.md`.

**Verification:** `go test ./tests/content/ -run 'SingleHome|LeafSkill|LeafLinks'` pasa; `rg -n 'round' content/skills/hive-rule-close-report/SKILL.md` solo aparece bajo «Findings at the close».

### T3 — `report-endings` como hogar de cómo termina el mensaje

- [x] `hive-rule-report-endings` tiene la tabla de formas de terminar de `design.md`, la fila de cierre con el footer, sus respuestas y el enlace a la línea de espera de `hive-rule-questions` «Format», y la sección «Same-turn continuations» con la sintaxis por host; `hive-rule-next-step-block` y `hive-rule-close-footer` se borran y todo enlace a ellas apunta a la sección de `report-endings` o `close-report` que tomó su texto; la columna «Ends with» de `close-modes`, el paso 8 de `flow-close` y su descripción enlazan la fila que aplica; `global.md` pierde las dos entradas del índice y las líneas de «Questions and routes» que T1 marcó como duplicadas, conservando el invariante mínimo.

**Closes:** AC3, AC4, AC8

**Depends on:** T1, T2

**Locations:** `content/skills/hive-rule-report-endings/SKILL.md`, los dos directorios borrados, `content/skills/hive-rule-close-modes/SKILL.md`, `content/skills/flow-close/SKILL.md`, `content/guidance/global.md`, los 11 archivos que enlazan las hojas borradas, `tests/content/leaf_skill_test.go`, `tests/content/leaf_links_test.go:30` (la arista `close-footer → flow-close` pasa a `report-endings → flow-close`), `tests/content/single_home_test.go`, `tests/content/reading_footprint_test.go`

**Execution:** hilo principal

**Test approach:** check — `leaf_skill_test.go` da a `hive-rule-report-endings` su propia excepción de 9,000 B (D11 A); `leaf_links_test.go` cambia la arista; `single_home_test.go` fija la frase de la tabla de finales.

**Changes:** `report-endings` conserva su nombre y la frase «Three reports end differently» (o se actualiza junto con `single_home_test.go:82`), porque `tests/pilot/session_audit_endings.go:14` y sus fixtures la citan por nombre.

**Verification:** `rg 'hive-rule-(next-step-block|close-footer)' --glob '!_support/openspec/changes/archive/**' --glob '!_support/sessions/**' --glob '!_support/openspec/changes/reporting-family-leaves/**'` sin resultados; `go test ./tests/content/ ./tests/pilot/` pasa.

### T4 — Un solo escritor del reporte final

- [x] `git-workflow` paso 9 y `flow-deploy` paso 8 verifican y entregan lo observado a `flow-close` sin escribir reporte; `flow-close` paso 6 escribe el reporte con la fila de `close-report` que aplica; `unattended-delegation` paso 9 resume la sesión (efectos con su reversión, decisiones tomadas, lo no resuelto) y enlaza el reporte de cada cierre; `flow-deploy` paso 6 nombra su espera de recuperación como mensaje de bloqueo.

**Closes:** AC1, AC2

**Depends on:** T2

**Locations:** `content/skills/git-workflow/SKILL.md`, `content/skills/flow-deploy/SKILL.md`, `content/skills/flow-deploy/references/deployment-checks.md`, `content/skills/unattended-delegation/SKILL.md`, `content/skills/flow-close/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** no cambian los pasos de cierre ni cuándo corren; `hive-rule-merge-gate` conserva su mensaje del CI de la base, que no es un reporte final.

**Verification:** `rg -n 'final report|close-report' content/skills/{git-workflow,flow-deploy,unattended-delegation,flow-close}` muestra el paso 6 de `flow-close` como único escritor; `go test ./tests/content/` pasa.

### T5 — La continuación en el mismo turno tiene un hogar

- [x] `hive-mode`, `flow-build`, `plan-handoff` y `global.md` nombran la continuación con un enlace a «Same-turn continuations» de `report-endings` y conservan solo su propio paso; `single_home_test.go` fija la frase de la continuación de investigación a `flow-build` en `report-endings`.

**Closes:** AC5

**Depends on:** T3

**Locations:** `content/skills/hive-mode/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/guidance/global.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** cada flujo conserva el paso que ejecuta la continuación (`flow-build` paso 5, `plan-handoff`, `hive-mode` paso 0); solo deja de reescribir sus condiciones.

**Verification:** `go test ./tests/content/ -run SingleHome` pasa; `rg -n 'same turn' content/` muestra el texto entero solo en `report-endings`.

### T6 — `close-hygiene` deja el paso 6 a `close-report`

- [x] La línea de `close-hygiene` en `## Rule leaves` y la hoja `hive-rule-close-hygiene` (descripción y cuerpo) ya no incluyen el paso 6 de `flow-close`. La prueba literal del índice ya no es de este plan: la hace `hive-core-stage-packets` (D7 A).

**Closes:** AC10

**Depends on:** T3

**Locations:** `content/guidance/global.md` «Rule leaves» (línea de `close-hygiene`), `content/skills/hive-rule-close-hygiene/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo el rango de pasos de esa línea y esa hoja; ninguna otra descripción cambia.

**Verification:** `rg -n 'steps 4 to 7' <checkout>/content` no imprime nada; `go -C <checkout> test -count=1 ./tests/content/...` pasa; `wc -c content/guidance/global.md` no pasa de su tamaño en la base del build (28,593 B en `7f6277e0`).

### T7 — Cada playbook nombra su fila

- [x] Los 13 archivos de `content/skills/hive-mode/playbooks/` tienen una línea que enlaza su fila de `report-endings` y «la fila de `close-report` que elige el modo de entrega vigente»; `pr-review-deploy` enlaza además la fila de despliegue; ninguna nombra un modo.

**Closes:** AC7

**Depends on:** T2, T3

**Locations:** `content/skills/hive-mode/playbooks/*.md`, `tests/content/leaf_skill_test.go` (caso que exige la línea), `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** hilo principal

**Test approach:** check — un caso exige la línea en cada playbook; el caso existente de la frase del modo vigente sigue pasando.

**Changes:** ningún playbook copia texto de las hojas; solo enlaza.

**Verification:** `go test ./tests/content/` pasa y cada playbook ≤ 7,000 B.

### T8 — Spec, sitio, `llms.txt`, changelog y documentación

- [x] El delta de `workflow-lifecycle` refleja lo construido; `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/deploy.md` y `llms.txt` describen los dos hogares y el cierre según su modo; `CHANGELOG.md` `[Unreleased]` tiene la entrada; `_support/docs/architecture/instruction-resources.md` y las otras páginas que citaban las hojas borradas apuntan a su destino; el texto del PR (para cuando la entrega salga de `hold`) queda en `inventory.md` con los lugares revisados y el tamaño de `global.md` antes y después.

**Closes:** AC3, AC4

**Depends on:** T3, T4, T5, T6, T7

**Locations:** `specs/workflow-lifecycle/spec.md` de esta carpeta, `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/deploy.md`, `llms.txt`, `CHANGELOG.md`, `_support/docs/architecture/instruction-resources.md`, `_support/docs/architecture/workspace-and-artifacts.md`, `_support/docs/harness-engineering/session-review-rubric.md`

**Execution:** hilo principal

**Changes:** mecánico; las páginas describen y enlazan, no copian reglas.

**Verification:** `go test ./tests/content/` (enlaces) pasa; `python3 -I <flow-plan-dir>/scripts/check_change.py _support/openspec/changes/reporting-family-leaves` sale 0.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, local | El comando de su «Verification» pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...` (≈ 251 s) | Cubierto por la implementación |
| Inventario sin pérdidas | Candidato final | `inventory.md` sin oraciones sin destino | Cubierto por la implementación |
| Revisión de código | Candidato final, antes de entregar el árbol de trabajo | `/code-review` de Claude Code; `inventory.md` como insumo para AC9 | La respuesta de revisión |
| Instalación | Solo cuando el cambio llegue a `development` (fuera de `hold`) | `hive update --source <checkout> --out FILE` y `hive apply --plan FILE`; leer las dos hojas instaladas y confirmar que las borradas ya no están | Refresco del mantenedor (`maintainer.local`) |

Caminata humana: no hay pantalla; la validación es leer las dos hojas resultantes y el inventario.

## Checkpoint

- Updated: 2026-11-21
- Stage: follow-up fixes · hallazgos H11–H18 y revisión acotada
- Last commit: 856149db
- Open findings: none
- Next: push, PR, merge, despliegue; archivo pendiente (higiene)
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/reporting-family-leaves · fix/reporting-findings

### Evidence

- before captures: none (sin pantalla)
- go test -count=1 ./... · exit 0, 262 s · 36699815 (dirty: solo tasks.md)
- python3 -I -m unittest discover -s tests/skills · exit 0, 115 s · 36699815
- go vet ./... · exit 0 · 36699815
- local CI: solo master-source-branch (guarda de PRs a master), no aplica a un PR contra development
- go test -count=1 ./... · exit 0, 259 s · aaf32526
- python3 tests/skills · exit 0 · 36699815, sin cambios de Python desde entonces (reuso)

## Review status and progress

2026-11-20 — Tareas escritas; reescritas tras la revisión del plan (orden: el inventario, luego los dos hogares, luego los flujos, la continuación, el índice y los playbooks).
2026-11-21 — T1 · ab07dabd · unverified: none
2026-11-21 — Revisión durante el build: D11 A da a `report-endings` una excepción de 9,000 B. Se decidieron sin preguntar `S12.13`, `S4.13`, `S5.3` y `S10.8`/`S10.15` (ver la respuesta del T2). No cambia ningún criterio ni el delta, y no invalida ninguna tarea.
2026-11-21 — T2 · 94a48a14 · unverified: none
2026-11-21 — T3 · 6d9aad92 · unverified: none
2026-11-21 — T4 · cc6b183a · unverified: none
2026-11-21 — T5 · 7e3fdc95 · unverified: none
2026-11-21 — T6 · e2451b7b · unverified: none
2026-11-21 — T7 · 90e85220 · unverified: none
2026-11-21 — T8 · 36699815 · unverified: none
