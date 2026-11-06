# viewer-timeline-kind-errors — Errores del visor que nombran el campo real y contrato con campos por tipo

| Campo | Valor |
| --- | --- |
| Estado | Draft · plan escrito, en revisión |
| Tracker | ninguno: la investigación no encontró un ticket que lo cubra (búsqueda en GitHub Issues, 2026-11-05) |
| Git | tricell-hive · `development` @ `fbf3922e` (plan escrito sobre `084728fa`) · rama propuesta `fix/viewer-timeline-kind-errors` |

## Decisiones

2026-11-05
- **D1 A** — ¿Cómo se expresa qué campos acepta cada tipo de evento? Con una sola tabla en Go que alimenta tanto la pista del error del CLI (qué tipos aceptan el campo rechazado) como las condiciones por tipo del esquema de `append_event`; los briefs por flujo no cambian.
- **D2 A** — ¿Qué modo de entrega? Automático: revisión, push, PR y merge sin parar.
- **D3 A** — ¿Qué revisión de código? Claude Code `/code-review` antes del push.
- **D4 B** — ¿La pista del error cubre también los rechazos que devuelve el servidor, como el de `event update`? Sí: se extiende al camino de los errores del servidor, con su propia tarea y prueba.

## Pedido

El usuario, tras el diagnóstico de los rechazos `forbidden_for_kind` del visor (2026-11-05): «A» — arreglar las dos causas encontradas y planearlo:

- **Causa 1:** el error culpa al campo equivocado. `TimelineAppendEventRequest.validate` (`tooling/viewer/timeline_contract.go:1131-1142`) rechaza `stage`, `stage_state` y `expected_session_version` en eventos que no son `stage` y siempre nombra `stage`; `target` y `source_reference` se nombran juntos como `target` (línea 1180); la validación de estado (línea 2092) nombra `stage` para `stage_state`, `started_at`, `finished_at` y `document_snapshots`.
- **Causa 2:** el contrato no dice qué campos acepta cada tipo: `hive view contract --operation append_event` los lista como opcionales sin condición por `kind`.

## Objetivo

Hoy un productor que manda un campo que su tipo de evento no acepta recibe el nombre de otro campo y reintenta a ciegas: en cam (Codex, 2026-11-04 21:43) un evento `validation` con `expected_session_version` falló cuatro veces seguidas con `forbidden_for_kind on stage`. En total hubo unos 37 rechazos `invalid timeline input` en 14 sesiones de Codex entre el 3 y el 5 de noviembre. Al terminar, el error nombra el campo que sobra y dice qué tipos lo aceptan, y la operación completa del contrato declara esos campos por tipo.

## Alcance y aceptación

Dentro: los validadores de Go de `append_event` y del estado de la sesión, el mensaje de error del CLI para los rechazos locales de `event add` y para los `forbidden_for_kind` que devuelve el servidor en `event add`, `event update` y `stage update`, el esquema de entrada de `append_event` y una delta de spec.

Fuera: el validador espejo del frontend (`tooling/viewer/frontend/src/timeline-contract.ts`; en ejecución nada usa su rama de `append_event`, solo sus tipos), regenerar el bundle, los briefs por flujo (`--for`), los otros códigos de error y cualquier campo que dependa de tipos que no son de evento (tickets, artefactos, entregas).

- AC1 — `hive view event add` con `kind: validation` y `expected_session_version` falla con `invalid timeline input: forbidden_for_kind on expected_session_version` · hoy es falso porque reporta `on stage` (observado en cam, 2026-11-04).
- AC2 — Cada campo que depende del tipo se nombra a sí mismo cuando se rechaza: `stage_state` y `source_reference` en `append_event`, `task_id` y `segment_id` en eventos `stage` o `closure` (hoy `timeline_contract_v3.go:347` los nombra `target_stage_event_id`), y `stage_state`, `started_at`, `finished_at` y `document_snapshots` en la validación de estado · hoy es falso porque se agrupan bajo `stage` o `target`.
- AC3 — Cuando la validación local de `event add` rechaza con `forbidden_for_kind` un campo que depende del tipo, el mensaje del CLI agrega qué tipos lo aceptan, como `invalid timeline input: forbidden_for_kind on expected_session_version; only kind stage accepts it` · hoy es falso porque el mensaje termina en el nombre del campo.
- AC6 — Cuando el visor rechaza `event add`, `event update` o `stage update` con `forbidden_for_kind` en un campo que depende del tipo, como `stage_state` en un evento que no es de etapa, el mensaje del CLI agrega qué tipos lo aceptan, como `viewer request failed (HTTP 400: forbidden_for_kind on stage_state; only kind stage accepts it)` · hoy es falso porque `viewErrorDetail` (`tooling/cli/view_timeline.go:603-616`) termina en el nombre del campo.
- AC4 — El esquema de entrada de `hive view contract --operation append_event` declara, para cada campo que depende del tipo, los tipos que lo aceptan (condiciones `if`/`then` sobre `kind`), y ese esquema acepta y rechaza los mismos ejemplos que el validador, incluidos los campos de texto vacíos que el validador acepta · hoy es falso porque solo tiene la condición del título.
- AC5 — Una prueba falla si el validador acepta o prohíbe un campo para un tipo de forma distinta a la tabla · hoy es falso porque no existe tabla ni prueba.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- Código leído el 2026-11-05 en `084728fa`: `tooling/viewer/timeline_contract.go:1087-1200` y `2080-2096`, `tooling/viewer/timeline_contract_v3.go:338-385`, `tooling/viewer/timeline_protocol.go:340-392` y `510-560`, `tooling/cli/view_timeline.go:105-150`.
- `timelineCollectEnums` recorre `properties`, `items` y las ramas `anyOf`, `oneOf` y `allOf`, pero no entra en `if`, `then` ni `not` (`timeline_protocol.go:592-631`): una condición nueva no cambia los briefs solo si no deja `enum` ni `const` fuera de esas palabras clave.
- Tamaño de los briefs del binario instalado (2026-11-05): research 6 466 bytes, plan 6 462, build 9 023, verify 8 727, delivery 9 999, close 1 497, collab 1 624, local 1 180. `timeline_protocol_test.go:272-282` exige menos de 10 000.
- `ResolveTimelineStageEventID` (`timeline_contract_v3.go:1866-1868`), que corre en el servidor antes de `Validate`, rechaza `target_stage_event_id` también en `request` y `work_started`; la validación local del CLI no.
- `development` avanzó a `fbf3922e` durante el plan; ese commit solo toca `tooling/viewer/navigation*.go`.
- Spec `walkthrough-viewer`, requisito «Self-describing protocol and confirmed-input hygiene»: «A rejected input SHALL name its field and code».
- Diagnóstico: `_support/sessions/2026-11-05-token-economizer-headroom/token-economizer-headroom.research.md`.

## Entrega

Automática (D2 A): revisión, push, pull request y merge a `development` sin detenerse para validar. Revisión de código: Claude Code `/code-review` (D3 A) sobre el candidato antes del push. No autoriza promover a `master`, publicar una release ni refrescar la instalación local más allá de lo que fija `AGENTS.md`.

## Estado de revisión y progreso

2026-11-05 — plan escrito; primera revisión de `hive-review-plan` (2 P1, 5 P2, 4 P3) aplicada salvo partir AC2 (opcional); D4 B agregó AC6 y T4; segunda revisión, de esa revisión, (2 P2, 3 P3) aplicada salvo el caso opcional contra un visor real.
