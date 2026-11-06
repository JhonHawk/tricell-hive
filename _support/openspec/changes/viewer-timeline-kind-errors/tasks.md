# Tasks

Base: `development` @ `fbf3922e` (plan escrito sobre `084728fa`); rama `fix/viewer-timeline-kind-errors`. Todas las tareas corren en el hilo principal: cada una toca dos o tres archivos con lecturas acotadas, T2 y T3 dependen de la tabla de T1, y T4 de la función de pista de T2 (mismo archivo).

### T1 — Tabla de campos por tipo y rechazos que nombran el campo real

- [x] Existe la tabla exportada; los rechazos `forbidden_for_kind` de `append_event` y de la validación de estado nombran el campo que sobra, y una prueba compara la tabla con el validador.

**Closes:** AC1, AC2, AC5

**Depends on:** none

**Locations:** `tooling/viewer/timeline_contract.go`, `tooling/viewer/timeline_contract_v3.go`, `tooling/viewer/timeline_contract_test.go`

**Execution:** main thread

**Test approach:** tdd

**Changes:** la tabla; el campo nombrado en `timeline_contract.go:1141`, `1180` y `2092` y en `timeline_contract_v3.go:347`; y `validateTimelineAppendV3` rechaza también en `request` y `work_started` los campos de destino de etapa, como el servidor. No cambian los códigos, el orden de los chequeos, las ramas `required_*` ni el espejo del frontend.

**Verification:** antes de la primera edición, guardar en el directorio temporal del host los bytes de `go run ./tooling/cli view contract --for <flujo>` para los ocho flujos (la línea base de T3). Después, `go test ./tooling/viewer/ -run 'Timeline'` pasa con: la prueba de tabla × tipos, que falla si se quita una fila o se cambia un tipo, y que para la fila de destino de etapa también llama a `ResolveTimelineStageEventID`; un caso `validation` con `expected_session_version` que espera `expected_session_version: forbidden_for_kind`; y la carga de una sesión con un evento que no es de etapa y lleva `stage_state`, `started_at`, `finished_at` o `document_snapshots`, que espera cada campo nombrado.

### T2 — Pista del CLI con los tipos que aceptan el campo

- [x] El rechazo local de `hive view event add` agrega qué tipos aceptan el campo rechazado.

**Closes:** AC3

**Depends on:** T1

**Locations:** `tooling/cli/view_timeline.go`, su prueba en `tooling/cli/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** solo el formato del mensaje para `forbidden_for_kind` en campos de la tabla; los demás mensajes no cambian.

**Verification:** `go test ./tooling/cli/ -run 'Timeline'` pasa con un caso que espera `invalid timeline input: forbidden_for_kind on expected_session_version; only kind stage accepts it`, y con el binario construido en el directorio temporal del host, `<bin> view event add <id> --json <archivo>` con ese payload imprime el mensaje sin enviar nada.

### T4 — Pista en los rechazos que devuelve el visor

- [ ] Un `forbidden_for_kind` del servidor en `event add`, `event update` o `stage update`, sobre un campo de la tabla, se imprime con los tipos que lo aceptan.

**Closes:** AC6

**Depends on:** T2

**Locations:** `tooling/cli/view_timeline.go`, `tooling/cli/view_error_detail_test.go` o `tooling/cli/view_timeline_test.go`

**Execution:** main thread

**Test approach:** tdd

**Changes:** `viewHTTPStatusError` guarda el `Code` y el `Field` que `viewErrorDetail` ya lee. En `runViewTimeline`, justo después de `viewTimelineHTTP` (`view_timeline.go:190`), `errors.As` agrega a `Detail` la pista de la función de T2 cuando `command.operation` es `append_event`, `event_patch` o `stage_update`. No cambian las firmas de `viewTimelineHTTP` ni de `viewErrorDetail`, ni sus otros llamadores (`hook_send.go`, `view_timeline_open.go`, `view_work_engine.go`, `view_work_publish.go`), ni el resto del texto del error, incluido el consejo de recuperación.

**Verification:** `go test ./tooling/cli/ -run 'Timeline|ViewErrorDetail'` pasa con una prueba que responde con un servidor HTTP simulado `{"code":"forbidden_for_kind","field":"stage_state"}` a `event update` y espera `forbidden_for_kind on stage_state; only kind stage accepts it`, y con otra que responde lo mismo a una operación que no es de eventos y espera el mensaje de hoy, sin pista.

### T3 — Condiciones por tipo en el esquema de append_event

- [x] `hive view contract --operation append_event` declara, por cada fila de la tabla, que los otros tipos no aceptan el campo, y los briefs no cambian.

**Closes:** AC4

**Depends on:** T1

**Locations:** `tooling/viewer/timeline_protocol.go`, `tooling/viewer/timeline_protocol_test.go`

**Execution:** main thread

**Test approach:** tdd

**Changes:** entradas nuevas en `allOf` de `TimelineAppendEventRequest`, generadas desde la tabla; la condición del título y el esquema de `TimelineEvent` no cambian.

**Verification:** `go test ./tooling/viewer/ -run 'Protocol'` pasa con una prueba que, con un ayudante propio sin dependencias nuevas, evalúa las condiciones generadas (`if.properties.kind.not.enum` y su `then`) contra los ejemplos de la prueba de T1, incluidos los campos de texto vacíos: cada rechazo de la tabla viola una condición y cada combinación aceptada no viola ninguna. Además, `go run ./tooling/cli view contract --for <flujo> | wc -c` da los mismos bytes que la línea base guardada en T1 para los ocho flujos.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, local | `go test ./...` y `go vet ./...` | Cubierta por la implementación |
| Briefs sin cambios | Candidato final, local | bytes de `view contract --for` en los ocho flujos, contra la línea base de T1 | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` | D3 A |

## Estado de revisión y progreso

2026-11-05 — plan escrito y revisado por `hive-review-plan`.
2026-11-05 — T1 hecha: tabla `timelineKindFields`, rechazos con el campo real y prueba de deriva (dos mutaciones atrapadas); `go test ./tooling/viewer/` pasa. Línea base de briefs (bytes): research 6466, plan 6462, build 9023, verify 8727, delivery 9999, close 1497, collab 1624, local 1180.
2026-11-05 — T2 hecha: `viewTimelineKindHint` en el rechazo local; la prueba falló sin la pista y pasa con ella; el binario candidato imprime `forbidden_for_kind on expected_session_version; only kind stage accepts it` sin crear estado; `go test ./tooling/cli/ -run 'Timeline'` pasa.
2026-11-05 — T1 verificada por `hive-verify-task`: AC1, AC2 y AC5 se cumplen; por su sugerencia, la prueba de deriva recorre todos los tipos publicables de `timelineEventKinds` en vez de una lista fija.
2026-11-05 — T2 verificada por `hive-verify-task`: AC3 se cumple.
2026-11-05 — T3 hecha: 12 condiciones por tipo en `allOf` de `append_event`; la prueba compara esquema y validador en toda la matriz de campos × tipos, más los textos vacíos; una mutación (`maxLength` → `minLength`) da 47 discrepancias; los bytes de los briefs de los ocho flujos coinciden con la línea base.

## Checkpoint

- Etapa: build · paso 2 · siguiente: T4 (pista en los rechazos del visor); verificación de T3 en curso.
- Worktree `.claude/worktrees/viewer-timeline-kind-errors`, rama `fix/viewer-timeline-kind-errors` desde `fbf3922e`.
- Visor: work `viewer-timeline-kind-errors`, etapa build en curso.
- Hallazgos abiertos: ninguno.
