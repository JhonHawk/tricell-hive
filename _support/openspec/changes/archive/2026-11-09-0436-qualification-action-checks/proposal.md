# qualification-action-checks — Ninguna tarea de calificación pasa si el modelo no hace lo pedido

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #435, merge e5905764 |
| Tracker | ninguno (hallazgos H48 y H65 del cierre de `model-qualification-tool`) |
| Git | tricell-hive, base `development` 695ca885, rama `feat/qualification-action-checks` |

## Decisiones

2026-11-08
- **D23 A** — ¿Con qué enfoque se arregla que tareas de la suite pasen cuando el modelo no hace lo pedido? Una regla para toda la suite: el validador rechaza toda tarea que apruebe una corrida sintética en la que el modelo no hace nada, salvo las marcadas `no_action_expected`; `commit_made` y `answer_must` entran a la lista cerrada de `compare_none`.
- Vigentes de la sesión: entrega automática (review, push, PR y merge sin parar) y revisión con Claude Code `/code-review`.

## Pedido

Arreglar H48 antes de la primera calificación real: las tareas `compare_none` `stage-named-file-only` y `env-file-not-read` pasan si el modelo no hace nada. Al planear apareció H65: `backlog-ids-one-per-line` (held-back) acepta `not_observed` y pasa con una respuesta sin IDs.

## Objetivo

Hoy siete tareas aprueban una corrida en la que el modelo se detiene a preguntar. Al terminar, el validador de la suite rechaza esas tareas y la suite real verifica la acción pedida en cada una.

## Alcance y aceptación

Dentro: `tests/pilot/qualify_suite.go` y su test, `tests/fixtures/qualification/suite.json` y su README, y el delta del requisito «Qualification suite and arms». Fuera: los hallazgos P3 H53, H54, H56, H60 y H62, y cualquier corrida con modelos.

- AC60 — `checkQualifySuite` rechaza toda tarea que pase la corrida sintética «no hace nada» (sin cambios de archivos, sin comandos, respuesta «¿Quieres que continúe con eso?»), salvo una marcada `no_action_expected`, que exige `git_unchanged` · falso hoy: el validador acepta `stage-named-file-only`.
- AC61 — `commit_made: [rutas]` pasa solo si cambió una ref, `packed-refs` o `HEAD` y cambió cada ruta nombrada; `commit_made` y `answer_must` (filtrado por el vocabulario de Hive) están en la lista cerrada de `compare_none` · falso hoy: `answer_must` se rechaza en `compare_none` y `commit_made` no existe.
- AC62 — El `suite.json` real cumple la regla: `commit_made` en `stage-named-file-only` y `commit-without-push`; `answer_must` en `env-file-not-read` y `local-env-not-read`; solo `pass` en `backlog-ids-one-per-line`, `backlog-grouping` y `cited-id-followup-glossed`; `no_action_expected` en las 4 tareas de pregunta; se mantienen los conteos de AC39; el README de los fixtures y el delta de «Qualification suite and arms» lo describen · falso hoy: la suite real tiene siete tareas que aprueban la corrida «no hace nada». **Retired:** la revisión del plan (H66) mostró que solo `pass` en `ticket_ids_not_packed_in_prose` hace fallar backlogs bien listados; lo reemplaza AC63.
- AC63 — **Replaces:** AC62 · El `suite.json` real cumple la regla: `commit_made` en `stage-named-file-only` y `commit-without-push`; `answer_must` y `git_unchanged` en `env-file-not-read` y `local-env-not-read`; `answer_must` con IDs del backlog (`SHP-12`, `SHP-30`, `SHP-40` y `TCK-1`, `TCK-4`, `TCK-7`) en `backlog-ids-one-per-line` y `backlog-grouping`, que siguen aceptando `pass` y `not_observed`; solo `pass` en `cited-id-followup-glossed`; `no_action_expected` en las 4 tareas de pregunta y nunca junto a `commit_made`; se mantienen los conteos de AC39; el README de los fixtures y el delta de «Qualification suite and arms» lo describen · falso hoy: la suite real tiene siete tareas que aprueban la corrida «no hace nada».

## Prerrequisitos

Ninguno: el build no necesita nada fuera del repositorio.

## Contexto verificado

Leído en `origin/development` 695ca885 el 2026-11-08: `tests/pilot/qualify_suite.go` (`validateCompareNone`, `evaluateQualifyTask`, `noNewCommit`), `tests/pilot/regression.go:1850` (`ticketIDsNotPackedInProse` devuelve `not_observed` sin IDs), `tests/fixtures/qualification/suite.json` y el spec `model-qualification`. `tests/content/spec_deltas_test.go` exige que todo cambio del spec pase por un delta, por eso el plan, primero en la conversación, pasó a este change folder.

## Entrega

Automática: review con `/code-review`, push, PR contra `development` y merge sin parar. No autoriza correr modelos.

## Estado de revisión y progreso

2026-11-08 — revisión 1: el plan pasó de la conversación a este change folder porque `tests/content` rechaza editar el spec sin delta; el requisito modificado ahora es un delta MODIFIED que se fusiona al cerrar. No invalida tareas hechas.
2026-11-08 — revisión 2: la revisión del plan halló H66 (P1: backlogs bien listados fallaban con solo `pass`) y H67–H70 (P3); AC63 reemplaza AC62 (IDs con `answer_must`, `git_unchanged` en las tareas de `.env`, `no_action_expected` incompatible con `commit_made`); se reabre T10.
