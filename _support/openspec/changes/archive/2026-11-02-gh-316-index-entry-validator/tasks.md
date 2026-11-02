# Tareas

Base: `tricell-hive` en `87a394fa` (`origin/development`), rama `refactor/gh-316-index-entry-validator`. Todas las tareas van en el hilo principal: son tres cambios cortos sobre los mismos dos archivos, y delegarlos costaría más que hacerlos.

### T1 — Casos de caracterización de la entrada del índice

- [x] `cases.json` tiene un caso `index_entry` con su código por cada regla de AC2, y los dos runners pasan contra el código actual.

**Closes:** AC2

**Depends on:** ninguna

**Locations:** `tooling/viewer/testdata/timeline_contract/cases.json`

**Execution:** hilo principal

**Test approach:** characterization

**Changes:** solo casos nuevos, cada uno partiendo de la entrada válida «API index discriminant» con un único error. Los códigos se toman de lo que da hoy cada lenguaje. No cambia código.

**Verification:** `go test ./tooling/viewer -run TestTimelineContractSharedFixtures` y `pnpm run test` en `tooling/viewer/frontend` pasan sobre `87a394fa` más los casos.

### T2 — Validador dedicado en Go

- [x] `DecodeTimelineIndexEntry` valida la entrada con sus propias reglas y una función de tickets compartida con la sesión, sin armar una sesión.

**Closes:** AC1, AC2

**Depends on:** T1

**Locations:** `tooling/viewer/timeline_contract.go`, `tooling/viewer/timeline_catalog.go`

**Execution:** hilo principal

**Test approach:** characterization

**Changes:** se extrae la validación de tickets de `validateTimelineCatalog` y se reescribe el caso `timeline_api` de `DecodeTimelineIndexEntry` en el orden de `design.md`. `ValidateTimelineSession` no cambia de comportamiento.

**Verification:** `go test ./tooling/viewer ./tooling/cli` pasa, incluida la prueba del store que decodifica cada entrada que emite el servidor. `rg -n "DecodeTimelineSession" tooling/viewer/timeline_contract.go` no aparece dentro de `DecodeTimelineIndexEntry`.

### T3 — Validador dedicado en TypeScript

- [x] `parseTimelineIndexEntry` valida la entrada con sus propias reglas y una función de tickets compartida con la sesión, sin armar una sesión.

**Closes:** AC1, AC2

**Depends on:** T1

**Locations:** `tooling/viewer/frontend/src/timeline-contract.ts`, `tooling/viewer/frontend/src/timeline-catalog.ts`, el bundle generado en `tooling/viewer/static/app/`

**Execution:** hilo principal

**Test approach:** characterization

**Changes:** igual que T2 en TypeScript, y se regenera el bundle. `parseTimelineSession` no cambia de comportamiento.

**Verification:** `pnpm run check && pnpm run lint && pnpm run test && pnpm run build && pnpm run verify:dist` en `tooling/viewer/frontend` pasan. `parseTimelineSession` no aparece dentro de `parseTimelineIndexEntry`.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `gofmt -l tooling/ && go vet ./tooling/... && go test ./tooling/... ./tests/...` y la cadena de `pnpm` de T3 | Cubierta por la implementación |
| Conservación | Candidato final, en local | Los 8 casos `index_entry` previos y los de T1 dan el mismo veredicto y código que en `87a394fa`; el visor local real lista sus sesiones sin avisos al cargar el índice | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push o de la parada de validación | Claude Code `/code-review` | La respuesta de revisión |

No hay recorrido de UI: la biblioteca no cambia de aspecto ni de comportamiento.

## Estado de revisión y avance

2026-11-02 — Plan escrito; ninguna tarea empezada.
2026-11-02 — T1 hecha en `e6bd135c`: 26 casos nuevos, todos pasan en Go y TypeScript sobre el código actual. Cuatro reglas ya daban un código distinto en cada lenguaje (`archived` con tipo inválido: `invalid_fields_or_types` en Go y `invalid_value` en TS; `validation_count` negativo o > 508: `invalid_count` y `invalid_positive_integer`; más de 100 tickets: `array_required_or_limit` e `invalid_array`); esos casos fijan solo el rechazo. AC2 queda sin cumplirse en esas cuatro hasta decidir si se alinean.
2026-11-02 — T2 hecha en `7a7bd84d`: `go test ./tooling/viewer ./tooling/cli` pasa; los 26 casos dan en Go el mismo código que antes.
2026-11-02 — T3 hecha en `977cb85d`: la cadena `pnpm` y `verify:dist` pasan; los 26 casos dan en TS el mismo código que antes; bundle regenerado.
2026-11-02 — Suite completa: todo pasa salvo dos fallos ajenos al cambio. `gofmt -l` marca `tests/pilot/regression.go`, que ya estaba así en `origin/development`. `TestExecApplyPiInterruptKillsDescendants` (`tooling/management`) falló una vez bajo carga y pasa 3 de 3 cuando se corre aislado.
2026-11-02 — Verify sobre `977cb85d`. AC1 se cumple: ningún decodificador de entradas llama a la validación de sesión. AC2 se cumple en 30 de 34 casos `index_entry`; quedan los 4 con código divergente. Camino real: un visor de prueba con el binario del candidato y una copia del estado (puerto libre, sin abrir el navegador) reporta 0 errores; las 4 entradas reales pasan por `parseTimelineIndexEntry`. Copia y visor de prueba eliminados.
2026-11-02 — `/code-review` (medium) sobre `977cb85d`: 8 hallazgos, ninguno bloqueante. Los dos primeros (código de `validation_count` distinto y límite de tickets en otro orden) se corrigen con D6 B en `da9a9812`, que alinea ambos códigos y los fija en los casos. Pasan a los hallazgos del cierre: el orden de chequeos en entradas con varios errores (puntero y `validation_count`, que ya diferían antes del cambio), los chequeos de cabecera copiados de la validación de sesión en ambos lenguajes (riesgo de que se desalineen) y los tickets validados dos veces (ya pasaba en el camino de sesión).
2026-11-02 — Segunda `/code-review` sobre `da9a9812`: sin hallazgos bloqueantes; los no bloqueantes van a los hallazgos del cierre. Suite completa de Go en verde. PR #320 sin CI ni revisiones; mergeado en `e1ebb0c3`.

Checkpoint — cerrado.
