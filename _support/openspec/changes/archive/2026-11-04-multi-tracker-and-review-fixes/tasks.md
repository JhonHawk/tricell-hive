# Tareas

Base: `origin/development` 048ad0c3 más `4a6b0ea6` en `feat/gh-311-minimal-research-entry`. T6, la primera en orden, fija la definición en la guía, que un test de `tooling/cli` exige, antes de lanzar hijos. T1 y luego T2 corren en un mismo hijo `hive-build-backend`, en orden, porque comparten el paquete `tooling/cli`; mientras, el hilo principal hace T3 y T4, que son redacción que el usuario juzga.

### T6 — Definición de la clave en «Project settings»

- [x] «Project settings» de `global.md` define `Other trackers` y `Completion` por tracker con la gramática de `design.md`.

**Closes:** AC4 (parte de la definición)

**Depends on:** respuestas D9 y D10

**Locations:** `content/guidance/global.md` «Project settings»

**Execution:** hilo principal

**Test approach:** check

**Verification:** `go test ./tooling/cli/ -run HiveSettingKeys` pasa una vez añadida la clave en T1.

### T1 — `Other trackers` y `Completion` por tracker en el manager

- [x] `hive doctor`, `hive project set` y la vista Project aceptan, validan y escriben `Other trackers` y `Completion` por prefijo.

**Closes:** AC1, AC2, AC3

**Depends on:** T6

**Locations:** `tooling/cli/doctor_project.go`, `tooling/cli/doctor_project_test.go`, `tooling/cli/tui_project_view.go`, `tooling/cli/tui_project_view_test.go`, `tooling/cli/project_write.go` y su test solo para el texto del aviso

**Execution:** hijo `hive-build-backend`, con `design.md` como contrato congelado

**Test approach:** tdd (la forma de `## Hive` es un contrato persistido)

**Changes:** añade `{"Other trackers", false}` después de `Tracker`; valida cada tracker, sus prefijos y cada valor de `Completion` por separado, con la gramática de `design.md`; los hallazgos nuevos son avisos. Actualiza los tests existentes que dependen del número y orden de claves (`formRows`, la navegación por posición, «Known keys», la lista de opcionales). No cambia la forma de `Tracker` ni las claves retiradas. Usa solo marcadores en los tests.

**Verification:** `go test ./tooling/cli/` pasa, con casos que fallan antes del cambio: `Other trackers` desconocida, un valor de `Completion` mal formado tras `; ` (hoy aceptado porque el patrón se aplica a la línea entera), un prefijo no declarado y un prefijo repetido.

### T2 — Errores de `tasks register` con ruta y regla

- [x] `invalid_input` de `hive view work tasks register` y `task update --json` nombra la ruta y la regla, sin valores.

**Closes:** AC8

**Depends on:** T1 terminada (mismo hijo, mismo paquete)

**Locations:** `tooling/cli/view_work.go` y su test; `tooling/viewer` no cambia

**Execution:** el mismo hijo `hive-build-backend`, después de T1

**Test approach:** tdd (defecto)

**Changes:** decodificación elemento a elemento y rutas `tasks[<i>].<campo>` / `metadata.<campo>`, como `design.md`. El código `invalid_input` y la sugerencia del contrato no cambian; el contrato de `hive view work contract` documenta la convención y dos reglas que hoy no dice: la clave de tarea es `T1` en adelante (`T0` se rechaza) y el `order` de un segmento es único por etapa entre todas las tareas, no dentro de cada una. Ambas se observaron en este build (2026-11-04): el registro falló tres veces con «invalid semantic JSON».

**Verification:** tests que fallan antes del cambio: dos tareas con la misma clave esperan `tasks[1].key: duplicate task key`; dos tareas con el segmento `build` en `order` 1 nombran `tasks[1].segments[0].order` y la regla; la clave `T0` nombra `tasks[0].key`; un campo desconocido nombra su ruta sin repetir el valor.

### T3 — Guía y skills: trackers, H11 y H12

- [x] Las reglas que mueven, leen o crean tickets usan el tracker de cada clave, y se corrigen H11 y H12.

**Closes:** AC4, AC5, AC6, AC7

**Depends on:** T6

**Locations:** `content/guidance/global.md` («Scope and authorization», «Questions and routes», «Scope changes»), `content/skills/flow-close/SKILL.md` (pasos 3 y búsqueda de hallazgos; ≤ 16 000 bytes), `content/skills/flow-deploy/SKILL.md` «Completion», `content/skills/flow-status/SKILL.md`, `content/skills/flow-research/SKILL.md`, `content/skills/harness-audit/references/instruction-files.md` HA-IF-17 y `_support/docs/harness-engineering/harness-audit-rules.md`, `content/skills/flow-build/references/delegation.md` «Reviewer brief» (línea canónica, conciliada con «nothing settled» de su línea 28)

**Execution:** hilo principal

**Test approach:** check

**Changes:** la regla de trackers vive en «Project settings»; los demás lugares solo la apuntan. HA-IF-17 reporta un tracker nombrado fuera de `## Hive`, y una regla del proyecto que exige confirmar escrituras al tracker como decisión del usuario (mantenerla o quitarla), nunca como `stale`.

**Verification:** `go test ./tests/...` y `go test ./tooling/cli/ -run HiveSettingKeys` pasan; búsqueda de cada lugar que nombra el tracker, listada en el commit con el tamaño de `global.md`.

### T4 — Specs, documentación, sitio y CHANGELOG

- [x] Los deltas de spec, la arquitectura, el sitio y el CHANGELOG describen la clave nueva, con marcadores.

**Closes:** AC9

**Depends on:** T1 (forma final de los avisos)

**Locations:** `specs/versioned-installation/spec.md` y `specs/workflow-lifecycle/spec.md` de esta carpeta, `_support/docs/architecture/deployment-manager.md` (líneas 76, 110 y 299), `site/src/content/docs/getting-started.md`, `site/src/content/docs/flows/deploy.md`, `site/src/content/docs/flows/close.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Verification:** build del sitio pasa; los deltas se leen contra «Read-only diagnostics».

### T5 — Verificación del conjunto

- [x] Suite completa y verificación independiente de T1.

**Closes:** AC1–AC9

**Depends on:** T1, T2, T3, T4

**Execution:** hilo principal; un hijo `hive-verify-task` verifica T1 por ser un contrato persistido

**Verification:** `go test ./...` y `go vet ./...` pasan; `hive doctor --project` sobre un repositorio temporal con el ejemplo de `design.md`.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...`, `go vet ./...`, build del sitio | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D8 A) | La respuesta de revisión |

Antes del push se nombran los commits de la rama que no son de esta obra (gh-311 y `4a6b0ea6`), como pide «Delivery». El mensaje de cada commit que corrige una falla observada nombra sus sesiones: H11 (Codex cam 01a1a5b7, 01a1a457; tricell-hive 01a1a55e; 2026-11-04) y H12 (Claude Code facti d3ff7bd2, tricell-hive f9f55a91; Grok cam 01a1a4f3; 2026-11-04).

## Estado de la revisión y avance

2026-11-04 — plan escrito; revisión independiente: 13 hallazgos, R1–R6 y R9–R13 aplicados; R7 y R8 resueltos por el usuario (D9 A, D10 A).
- Checkpoint: plan listo; visor work `6e89cf6319e299588cc1e7ea9dcdfe616d2b0a9b5facd34b66c721884687646e`, snapshot `revision-56451fce1fa04f5cbbce4e8d4959dca6`. Siguiente: `/flow-build` (T6, luego el hijo de T1→T2 en paralelo con T3 y T4).
- 2026-11-04 · T6+T3 6b80fc65, T1 053e159b, T2 1fa51ea5, T4 21298fad. `go vet ./...` limpio, `go test ./...` ok, build del sitio ok, binario candidato probado con `hive doctor --project` y `tasks register`. Abierto: AC3/spec dicen «después de Tracker», pero el editor añade una clave nueva al final de una sección existente (decisión del usuario). En curso: hive-verify-task de T1 y /code-review.
- 2026-11-04 · Verificación de T1 (hive-verify-task): AC1 y AC2 cumplidos; AC3 cumplido con D11 A (clave al final de una sección existente). Primera /code-review: 10 hallazgos, 9 corregidos en c710f1f9; la duplicación del localizador de JSON con el visor queda listada. Segunda /code-review sobre c710f1f9: 7 hallazgos, 4 corregidos en dda11f4d (`Unreviewed:`), 3 de mantenimiento listados. `go test ./...`, `go vet ./...` y el build del sitio pasan.
- Hallazgos abiertos (P3, mantenimiento): el localizador de fallas de JSON repite el recorrido cerrado del visor (arreglo de fondo: ruta en el error de contrato del visor); el chequeo de null vive en dos ramas; el bucle de separación repite el del placeholder del TUI; una falla de campo no localizada cae en «body: invalid fields or types».
- Checkpoint: build y verificación completos; push de la rama hecho. Siguiente: PR a development y despliegue solo cuando el usuario lo pida (D5 C); después, /harness-audit en los proyectos afectados.
- 2026-11-04 · PR #361 fusionado como b37ad0c9 y desplegado (binario recompilado; hive update a todos los hosts). Deltas fusionados en specs/versioned-installation y specs/workflow-lifecycle; carpeta archivada.
