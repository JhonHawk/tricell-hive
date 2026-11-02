# Tareas

Base: `tricell-hive` en `93b0ce79` (`origin/development`), rama `fix/gh-316-review-findings`. Todo en el hilo principal: son cambios cortos sobre los mismos archivos y sus briefs no serían más cortos que los cambios.

### T1 — Casos de dos fallos y del límite de tickets

- [ ] `cases.json` fija los códigos de AC3 y AC4 y esos casos fallan sobre el código actual.

**Closes:** AC3, AC4

**Depends on:** ninguna

**Locations:** `tooling/viewer/testdata/timeline_contract/cases.json`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** cinco casos nuevos: dos de entrada del índice con dos fallos (AC3); uno de sesión con 101 tickets y un duplicado, y uno de bootstrap de creación con 101 tickets (AC4); y uno de entrada del índice con 101 tickets y un duplicado, que hoy ya da `array_required_or_limit` en ambos lenguajes y protege el cambio de T3.

**Verification:** sobre `93b0ce79`, TypeScript falla en los dos casos de AC3 y en los dos de AC4; Go falla solo en el caso de sesión; el caso de entrada con 101 tickets y un duplicado pasa en ambos.

### T2 — Go: cabecera compartida y tickets una vez

- [ ] La sesión y la entrada del índice comparten `validateTimelineSummaryHeader`; `addTickets` solo registra identidades; la sesión revisa el límite antes que las identidades.

**Closes:** AC1, AC2, AC4

**Depends on:** T1

**Locations:** `tooling/viewer/timeline_contract.go`, `tooling/viewer/timeline_catalog.go`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** como en `design.md`; el límite de tickets usa la constante existente `TimelineMaxCatalogRecords`. No cambia ningún caso de un solo fallo.

**Verification:** `go test ./tooling/viewer ./tooling/cli` pasa, incluidos los casos de T1.

### T3 — TypeScript: cabecera compartida, orden, tickets una vez y código del límite

- [ ] `parseTimelineSession` y `parseTimelineIndexEntry` comparten `summaryHeader`; la entrada sigue el orden de Go; `addTickets` no revalida; `tickets()` completo da `array_required_or_limit`.

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** T1

**Locations:** `tooling/viewer/frontend/src/timeline-contract.ts`, `tooling/viewer/frontend/src/timeline-catalog.ts`, el bundle en `tooling/viewer/static/app/`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** como en `design.md`; se regenera el bundle.

**Verification:** `pnpm run check && pnpm run lint && pnpm run test && pnpm run build && pnpm run verify:dist` pasan, incluidos los casos de T1.

### T4 — Test de interrupción determinista

- [ ] El falso `pi` publica su PID con un renombrado atómico.

**Closes:** AC5

**Depends on:** ninguna

**Locations:** `tooling/management/pi_package_exec_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo el script de `fakeHangingPi`.

**Verification:** `go test ./tooling/management -run 'TestExecApplyPiInterruptKillsDescendants$' -count=50` pasa mientras corre `go test ./tooling/...` en paralelo. Queda un riesgo de tiempo que este cambio no cubre: bajo carga, `TestExecApplyPiTimesOutAndKillsDescendants` (300 ms) puede matar el script antes de que publique el PID.

### T5 — `gofmt` limpio

- [ ] El comentario de `inlineCodeSpan` describe el límite sin la secuencia de backticks que `gofmt` reescribe.

**Closes:** AC6

**Depends on:** ninguna

**Locations:** `tests/pilot/regression.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo el texto del comentario.

**Verification:** `gofmt -l tooling/ tests/` no imprime nada; `go test ./tests/pilot` pasa.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `gofmt -l tooling/ tests/ && go vet ./tooling/... ./tests/... && go test ./tooling/... ./tests/...` y la cadena de `pnpm` de T3 | Cubierta por la implementación |
| Conservación | Candidato final, en local | Todos los casos de un solo fallo previos conservan veredicto y código; un visor de prueba con el binario del candidato sobre una copia del estado reporta 0 errores, sin abrir el navegador, y `parseTimelineIndexEntry` (con `tsx`) acepta cada entrada de su índice | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push o de la parada de validación | Claude Code `/code-review` | La respuesta de revisión |

No hay recorrido de UI: la biblioteca no cambia de aspecto ni de comportamiento.

## Estado de revisión y avance

2026-11-02 — Plan escrito; ninguna tarea empezada.
