# Diseño

## Enfoques

- **A — Pista en el error y esquema condicional desde una sola tabla (elegido, D1 A).** Cubre al productor que solo lee el brief (la pista llega en el primer rechazo) y al que lee la operación completa (el esquema lo declara antes de enviar). Es el que más pruebas mueve.
- **B — Solo la pista en el error.** Es lo mínimo que corta el bucle observado, pero el contrato sigue sin declarar los campos por tipo.
- **C — Solo el esquema condicional.** Ayuda solo a quien lee `--operation append_event`; el error quedaría corregido en el nombre pero sin pista.

Los briefs por flujo quedan fuera de los tres: `--for delivery` mide 9 999 bytes y no cabe nada.

## Forma propuesta

Una tabla exportada en `tooling/viewer`, por ejemplo `TimelineEventKindFields`, que mapea cada campo de evento dependiente del tipo a los tipos que lo aceptan:

| Campo | Tipos que lo aceptan |
| --- | --- |
| `stage`, `stage_state`, `expected_session_version` | `stage` |
| `views` | `validation`, `iteration` |
| `target`, `source_reference` | `changes_reported`, `approval_reported` |
| `delivery_ids` | `delivery` |
| `closure` | `closure` |
| `review_status` | `review` |
| `target_stage_event_id`, `task_id`, `segment_id` | `validation`, `changes_reported`, `iteration`, `approval_reported`, `delivery`, `review` |

Un campo solo entra en la tabla si su validador lo rechaza con `forbidden_for_kind` para los demás tipos. Las ramas de `required_*` siguen en el validador, que no se reescribe: solo cambia el campo que nombra cada rechazo, y se conserva el orden actual de los chequeos (primero `validateTimelineAppendV3`, después las ramas de `validate`).

- **Validador (T1):** `timelineError(<campo real>, "forbidden_for_kind")` en `timeline_contract.go:1141`, `1180` y `2092` y en `timeline_contract_v3.go:347`, que hoy nombra `target_stage_event_id` también cuando sobran `task_id` o `segment_id`. Además, `validateTimelineAppendV3` pasa a rechazar esos tres campos también en `request` y `work_started`, como ya hace el servidor en `ResolveTimelineStageEventID` (`timeline_contract_v3.go:1866-1868`): así la validación local, la tabla y el servidor coinciden, y ninguna petición que hoy el servidor acepta pasa a fallar. Una prueba recorre la tabla × `timelineEventKinds` y comprueba que el validador rechaza con `forbidden_for_kind` en ese campo exactamente donde la tabla lo prohíbe; para la fila de destino de etapa también llama a `ResolveTimelineStageEventID` sobre una sesión de prueba. Los campos solo de estado guardado (`started_at`, `finished_at`, `document_snapshots`, línea 2092) no entran en la tabla: se prueban aparte.
- **CLI (T2):** en `tooling/cli/view_timeline.go`, para las operaciones `append_event` y `event_patch`, con código `forbidden_for_kind` y un campo de la tabla, el mensaje agrega `; only kind <a> accepts it` o `; only kinds <a>, <b> accept it`.
- **Errores del servidor (T4, D4 B):** `event update` no valida el tipo localmente; su único `forbidden_for_kind` lo devuelve el visor (`timeline_contract.go:2210-2211`, `stage_state` en un evento que no es de etapa), y el CLI lo imprime con `viewErrorDetail` (`tooling/cli/view_timeline.go:603-616`). `viewHTTPStatusError` pasa a guardar el código y el campo, y `runViewTimeline`, el único que conoce la operación, agrega la pista de T2 al detalle para `append_event`, `event_patch` y `stage_update` (este último recibe el mismo rechazo del servidor cuando apunta a un evento que no es de etapa). Las firmas de `viewTimelineHTTP` y `viewErrorDetail` no cambian, así que sus otros llamadores tampoco.
- **Esquema (T3):** `timelineProtocolSchema` agrega a `allOf` de `TimelineAppendEventRequest`, por cada fila, `if: {properties: {kind: {not: {enum: <tipos>}}}, required: [kind]}` con un `then` que replica lo que rechaza el validador: `{properties: {<campo>: {maxLength: 0}}}` para los campos de texto que el validador acepta vacíos (`stage`, `source_reference`, `review_status`, `target_stage_event_id`, `task_id`, `segment_id`; `stage_state` según lo que acepte el decodificador), y `{not: {required: [<campo>]}}` para los que se prohíben con solo estar (`expected_session_version`, `views`, `target`, `delivery_ids`, `closure`). Se suma a la condición del título que ya existe. Forma obligatoria: ningún `enum` ni `const` fuera de `if`/`then`/`not`, porque `timelineCollectEnums` recorre las ramas de `allOf` (`timeline_protocol.go:592`) y un `enum` suelto filtraría valores de `kind` a todos los briefs. `TimelineEvent`, que es salida, no cambia. La prueba evalúa esa forma con un ayudante propio de pocas líneas, sin agregar una biblioteca de JSON Schema.

## Impacto

- El texto `forbidden_for_kind on stage` lo producen hoy `timeline_contract.go:1141` y `2092`; `on target`, la línea 1180. Ninguna prueba de Go, fixture compartido (`tooling/viewer/testdata/timeline_contract/*.json`, que también leen las pruebas del espejo en TypeScript), archivo de `content/` ni página de `site/` contiene esos textos, así que no hay aserciones que actualizar.
- `view work evidence`, `publish` y `lifecycle` arman peticiones válidas y no dependen del campo que se nombra.
- Nada en `content/` ni en `site/` cita estos mensajes (búsqueda del 2026-11-05).
- El espejo del frontend (`timeline-contract.ts:920`) conserva `stage`; en ejecución nada usa esa rama.
- El campo que nombra un rechazo cambia solo en los casos mal atribuidos; los códigos de error no cambian. Es una corrección compatible con el spec vigente.

## Recuperación

Revertir el commit; no hay datos guardados ni migraciones.
