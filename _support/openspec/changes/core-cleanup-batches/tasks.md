# Tasks

Base: `development` en `8aa361cb`. Bajo `hold` no hay commit, así que no hace falta rama: el build propone, antes de su primera edición, trabajar en el sitio sobre `development` o en un worktree si otra sesión ocupa la checkout (git-workflow paso 3); el árbol verificado queda donde se elija. T1, T2, T3 y T5 van en el hilo principal: fijan reglas, pruebas y paquetes que las demás tareas usan, y T3 aplica supuestos que pueden necesitar una pregunta. T4, T6 y T7 van, en ese orden, a un hijo `hive-write-docs` cada uno, porque cada uno edita más de 10 archivos de guía; van uno tras otro porque varias filas tocan los mismos archivos (`flow-verify/SKILL.md`, `flow-build/SKILL.md`, `child-brief.md`). En cada edición que toque una línea con «never» que no sea invariante, la tarea la reescribe con alcance y razón (D5). Entrega `hold`: sin commits; si el usuario luego pide entregarlo, un commit por lote: T1+T2, T3, T4+T5, T6, T7.

### T1 — AGENTS.md fija la evidencia de D4 y la redacción de D5

- [x] `AGENTS.md` «Measurement» dice lo que fija D4 y «Guidance structure» lo que fija D5, cada cosa una sola vez, sin contradecir las líneas que quedan.

**Closes:** AC1, AC2

**Depends on:** none

**Locations:** `AGENTS.md`

**Execution:** main thread

**Test approach:** check

**Changes:** reescribe la última oración de «Measurement» sobre relajar o quitar una regla, y agrega una oración en «Guidance structure» junto a la del alcance explícito. Los deltas de spec ya están en la carpeta de cambio. No toca `content/guidance/global.md`: es guía del repositorio, no distribuida.

**Verification:** `rg -n -i 'audit-session' AGENTS.md` muestra la línea nueva de «Measurement»; `rg -n -i 'only for invariants' AGENTS.md` muestra la de «Guidance structure»; lectura de ambas secciones completas sin otra frase que diga lo contrario.

### T2 — Una prueba detecta las citas a secciones o pasos que no existen

- [x] Con la base, la prueba nueva falla nombrando `content/skills/hive-core/plan/scope-changes.md:12`; con PL-1 arreglado en esta misma tarea, pasa.

**Closes:** AC3

**Depends on:** none

**Locations:** `tests/content/cited_targets_test.go` (nuevo), `content/skills/hive-core/plan/scope-changes.md`

**Execution:** main thread

**Test approach:** tdd

**Changes:** recorre los `.md` de `content/`; para cada enlace a un `.md` seguido de `«Nombre»` exige en el destino un título `#` con ese nombre o una etiqueta en negrita al inicio de un ítem o párrafo, `**Nombre**` seguido de `.`, `:` o `,` (así cubre `**Decide and list.**` y `**When a check fails**,`); para cada enlace a un `SKILL.md` con «step N» en su texto o justo después exige un ítem `N.` en el destino. El mensaje nombra archivo, línea, cita y destino, y dice qué corregir. Reutiliza la forma de leer enlaces de `skill_links_test.go`. No verifica que el destino diga lo que la cita le atribuye. Después de verla fallar, arregla PL-1 (`scope-changes.md:12` enlaza `change-folder.md#revise-a-plan-in-progress`).

**Verification:** `go test ./tests/content/ -run TestCitedSectionsAndStepsExist -count=1` falla en la base con una línea que nombra `scope-changes.md:12` y `flow-plan/SKILL.md`, y con ninguna otra; después del arreglo de PL-1, la misma orden pasa y `go test ./tests/content/... -count=1` ok.

### T3 — Lote 1: referencias colgantes, destinos equivocados, circulares y conflictos

- [x] Cada fila de «Lote 1» de `design.md` resuelta en su casa, con los demás lugares como enlace con ancla.

**Closes:** AC4

**Depends on:** T2

**Locations:** las de cada fila de `design.md` «Lote 1» (hojas de `content/skills/hive-core/**`, skills, referencias y playbooks citados); `tests/content/single_home_test.go` y `leaf_links_test.go` si registran una frase o un enlace que se mueve

**Execution:** `hive-write-docs` (revisado en el build: edita ~40 archivos; el hijo se detiene y reporta toda fila que necesite una decisión, y el hilo principal revisa el diff y mide la huella)

**Test approach:** check

**Changes:** una edición por fila (salvo PL-1, que hizo T2), como dice la fila y los supuestos de `design.md`; si la primera oración de una hoja cambia, también su línea del índice en `content/guidance/global.md` (`index_trigger_test.go`). Una hoja que pase de 4.000 B lleva lo nuevo a una referencia propia (supuesto). No cambia lo que las reglas autorizan. Cada línea con «never» que reescribe por D5 queda listada en el progreso con su texto anterior y el nuevo, y el alcance no se estrecha (un alcance más estrecho es relajar, D4). Al cerrar, mide la unión plan/build de `reading_footprint_test.go`; si crece, se detiene y pregunta.

**Verification:** `go test ./tests/content/... -count=1` ok, incluida la prueba de T2; `rg -n 'Tests section' content` sin resultados; `rg -n 'git-workflow step 5' content` sin resultados; `go test ./tests/content/ -run TestRequiredProcedureReadingFootprint -count=1 -v` con la unión igual o menor; lectura fila por fila de cada frase citada en el candidato.

### T4 — Lote 2a: repeticiones parciales a punteros

- [x] Cada fila de «Lote 2a» de `design.md` queda como enlace con ancla a su casa, sin pasos, condiciones ni excepciones de ella.

**Closes:** AC5

**Depends on:** T3

**Locations:** las de cada fila de `design.md` «Lote 2a», incluidos `content/agents/review/hive-review-code.md` y `content/agents/quality/hive-verify-task.md`

**Execution:** `hive-write-docs`

**Test approach:** check

**Changes:** reemplaza cada repetición por el momento y el enlace a la casa, como fija la spec «A rule is cited outside its home, never restated»; donde una repetición tenía un matiz que la casa no tiene, lo lleva a la casa en vez de borrarlo, y lo nombra en su reporte. No toca los párrafos «Reporting» de los playbooks (CL-6, descartado). El hijo lista cada línea con «never» que reescribió por D5, con su texto anterior y el nuevo.

**Verification:** `go test ./tests/content/... -count=1` ok; lectura fila por fila en el candidato; el hilo principal revisa el diff de cada fila contra la casa y confirma que ninguna reescritura de D5 estrechó un alcance.

### T5 — Lote 2b: paquetes sin reglas condicionales y con la que el cierre necesita

- [x] Los cuatro paquetes de «Lote 2b» cambian como dice `design.md`, y la línea «it holds the leaves» de cada `SKILL.md` cuyo paquete cambia lista exactamente su `stage.json`.

**Closes:** AC9

**Depends on:** T4

**Locations:** `content/skills/flow-verify/assets/stage.json`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-build/assets/stage.json`, `content/skills/flow-build/SKILL.md` (línea 8), `content/skills/flow-close/assets/stage.json`, `content/skills/flow-close/SKILL.md` (línea 8), `content/skills/hive-core/delegation/reviewer-brief.md`, `content/skills/create-design/references/review.md`, una referencia nueva de create-design para el brief de marcos dibujados con sus lectores `create-design/references/example-record.md` y `rule-cards.md`, `tests/content/reading_footprint_test.go`, `tests/content/stage_manifest_test.go`, `tests/content/leaf_links_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** PK-1 a PK-4, con PK-4 partido en tres destinos como dice `design.md`. Actualiza la línea 8 de `flow-build/SKILL.md` y `flow-close/SKILL.md` (y la de `flow-verify/SKILL.md`) para que nombre solo las hojas de su paquete. Agrega a `stage_manifest_test.go` una comprobación de que esa lista coincide con `stage.json`, para que no vuelva a desalinearse. Ajusta solo las expectativas de las pruebas que listan esos paquetes o enlaces; si la unión de huella crece, se detiene y pregunta (supuesto).

**Verification:** `go test ./tests/content/... -count=1` ok, incluida la comprobación nueva de `stage_manifest_test.go`; `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRealStagePacketsKeepAMarginUnderTheCap' -count=1` ok; lectura de los tres `stage.json`, las tres líneas «it holds the leaves» y `reviewer-brief.md`.

### T6 — Lote 3: cada regla enlazada donde aplica

- [x] Cada fila de «Lote 3» de `design.md` queda enlazada en el paso del flujo donde aplica, y `security-boundaries.md` lista los ocho temas (D7).

**Closes:** AC7

**Depends on:** T5

**Locations:** las de cada fila de `design.md` «Lote 3», incluidos `content/agents/review/hive-review-security.md`, `content/agents/quality/hive-read-state.md` y el traslado de `flow-verify/references/host-session-identity.md` a `content/skills/hive-core/ui/viewer/host-session-identity.md`

**Execution:** `hive-write-docs`

**Test approach:** check

**Changes:** un enlace condicional por fila, con el momento en que aplica. D7: la primera oración de `security-boundaries.md` lista los ocho temas, su línea del índice en `content/guidance/global.md` cambia igual, y `flow-build/SKILL.md:28`, el paso 1 de `hive-review-security.md` y el enlace de B-10 apuntan a la hoja sin repetir temas. UI-5 mueve el archivo a `hive-core/ui/viewer/host-session-identity.md`; `hive-core/assets/domains.json` no cambia. Ninguna skill pasa de 16.000 B ni hoja de 4.000 B (supuesto). El hijo lista cada «never» reescrito por D5 con su texto anterior y el nuevo. Al cerrar, el hilo principal mide la unión de huella; si crece, se detiene y pregunta.

**Verification:** `go test ./tests/content/... -count=1` ok; `go test ./tooling/management/ -run TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink -count=1` ok; `rg -c 'git-workflow' content/skills/flow-verify/SKILL.md content/skills/flow-deploy/SKILL.md` cuenta al menos 1 en cada uno; `rg -n -i 'authentication' content/skills/hive-core/build/security-boundaries.md content/guidance/global.md` muestra la primera oración y su línea del índice; `go test ./tests/content/ -run TestRequiredProcedureReadingFootprint -count=1 -v` con la unión igual o menor; lectura fila por fila.

### T7 — Documentación de lo que cambió

- [x] `instruction-resources.md`, `rule-enforcers.md`, `CHANGELOG.md` y cada página de `site/` (o `README.md`/`llms.txt`) que describa un comportamiento que cambió lo describen como queda.

**Closes:** AC7

**Depends on:** T6

**Locations:** `_support/docs/architecture/instruction-resources.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`, `site/src/content/docs/viewer.md:68` (ruta de `host-session-identity.md`), `site/src/content/docs/flows/security.md:14` (temas de seguridad) y las demás páginas de `site/` que la búsqueda encuentre

**Execution:** `hive-write-docs`

**Test approach:** check

**Changes:** registra la prueba de T2 y la comprobación nueva de `stage_manifest_test.go` como mecanismos en `rule-enforcers.md`; actualiza en `instruction-resources.md` lo que describa los paquetes que cambió T5; busca en `site/`, `README.md` y `llms.txt` los temas de seguridad (en cualquier orden), el registro `Playbook`, el worklog al cerrar, la ruta movida de UI-5 y cualquier otra frase que T3–T6 cambien, y las alinea. Las páginas describen y enlazan, no copian reglas.

**Verification:** `go test ./tests/content/... -count=1` ok (incluye `rule_enforcers_test.go`); `rg -n 'flow-verify/references/host-session-identity' content site README.md llms.txt _support/docs` sin resultados; `rg -n -i -l 'process execution' content site README.md llms.txt _support/docs` y lectura de cada archivo que liste: todos nombran los ocho temas o enlazan la hoja; lectura de cada página cambiada.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Change folder | Final candidate, locally | `python3 -I content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/core-cleanup-batches` sale 0 | Covered by the implementation |
| Code review | Final candidate, before handing over the `hold` result | Claude Code `/code-review` sobre el diff contra `8aa361cb` | The review answer |

## Checkpoint

- Updated: 2026-11-22
- Stage: deliver · push + PR
- Last commit: c8e60a70
- Open findings: none
- Next: push refactor/core-cleanup-batches, open PR to development, wait CI, merge, refresh install
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- before captures · none (no screen) · 8aa361cb
- go test ./tests/content/ -run TestCitedSectionsAndStepsExist · FAIL on base naming scope-changes.md:12 only, PASS after PL-1; mutation step 55 caught · uncommitted on 8aa361cb
- T3 · go test ./tests/content/... ok; tooling/management real-content+packet tests ok; rg Tests section / git-workflow step 5 empty · uncommitted on 8aa361cb
- T5 · tests/content ok (incl. TestPacketHoldsLinesMatchManifests, mutation caught); tooling real-content+packet margin ok; footprint union 193354 B (project-verification moved to E2E route) · uncommitted
- T6 · tests/content ok; tooling ok; global.md 25909 B; footprint 195800 B · uncommitted
- local CI · go vet ./... ok; go test ./... 22/23 ok, tests/content failed only mid-T7 (rule-enforcers.md being edited) and passes rerun after T7 (go test ./tests/content/... ok); python3 -m unittest tests/skills 412 OK; site npm run build ok · uncommitted on 8aa361cb
- CI-only jobs · .github/workflows/master-source-branch.yml runs on push to master only; not applicable to hold
- AC walk · AC1 AC2 AC3 AC4 AC5 AC7 AC9 observations hold on candidate (rg + tests); suppression inventory: none; blast radius: TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink + packet margin + integrations/antigravity ok; no tooling/integrations ref to moved file · uncommitted
- after fix round · tests/content ok; tooling real-content + packet margin ok; footprint 195922 B · uncommitted
- final candidate · go vet ok; go test ./... 23/23 ok; python tests/skills OK; check_change.py 0 · uncommitted on 8aa361cb
- H15-H20 fixes · tests/content ok; tooling real-content + packet margin ok · uncommitted

## Review status and progress

2026-11-22 — T1 · uncommitted · unverified: none
2026-11-22 — T2 · uncommitted · unverified: none
2026-11-22 — T3 · uncommitted · unverified: PL-4 recovery clause move deferred to T5 (flow-build packet margin)
2026-11-22 — T4 · uncommitted · unverified: G-C5 step 5 merge skipped (renumbering of steps 6-10 cited elsewhere); CL-11 moved to T5; TR-6 time sentence kept in ticket-state.md:7
2026-11-22 — T5 · uncommitted · unverified: none
2026-11-22 — T6 · uncommitted · unverified: none
2026-11-22 — T7 · uncommitted · unverified: site build (astro) not run
