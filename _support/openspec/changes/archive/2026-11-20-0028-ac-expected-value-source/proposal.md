# ac-expected-value-source — Valor esperado con fuente independiente en los criterios que muestran datos

| Campo | Valor |
| --- | --- |
| Estado | Completado · entrega `automatic` |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `1cdb5af7` · rama `feat/ac-expected-value-source` |
| Playbook | feature (cambio pequeño sin plan, `/flow-build`) |

## Decisions

2026-11-20
- **D1 A** — ¿Cómo cerramos el hueco del oráculo independiente? En `flow-plan` «Acceptance criteria», la observación de un criterio que muestra datos nombra el valor esperado y una fuente que no corre el código bajo prueba (el seed o una consulta). Descartado: B (una consulta extra a la API o a la base en cada recorrido de verify), C (volver a lanzar `hive-verify-change` por defecto, revirtiendo la decisión del 2026-10-18), D (sin cambio, con watch item). Engram: `decision/ac-independent-oracle`.
- **Entrega** `automatic` · **Revisión** nativa `/code-review` · sin revisión de seguridad.
- **Supuesto** — Para no superar la línea base de `TestRequiredProcedureReadingFootprint` (le quedaba 1 B), se quitan de `flow-plan` dos repeticiones: la frase que reafirma que una tarea operativa conserva su plan en la conversación (ya está en la guía global, «Proportionality») y el puntero «asked as asking a round» del paso 4 (las hojas `hive-rule-delivery-ask` y `hive-rule-external-effects` ya apuntan a `questions`). Se revierte restaurando ambas frases.

## Why

El artículo «5 tips for writing good agentic tests» (e2e.com, 2026-11-13) y la evidencia revisada muestran que juzgar la pantalla falla justo en los datos. El evaluador automático de WebVoyager coincidió con humanos el 73,9–78,7 % de las veces y el mejor juez, alrededor del 85 % (Xue et al., arXiv 2504.01382). Reescribir solo la narración del agente subió los falsos positivos de los jueces hasta un 90 % (arXiv 2601.14691). En Hive, el punto 4 de `hive-rule-tests` ya prohíbe en los tests un valor esperado calculado por el código bajo prueba, pero el recorrido de los criterios no tenía equivalente. pstack no tiene esa regla, y la forma anterior de Hive solo la tenía para los tests de propiedades. Es un hueco (ninguna regla cubría el oráculo del recorrido), así que no requiere recurrencia. `Enforceable: judgment`, porque decidir qué criterio muestra datos requiere criterio.

## Acceptance criteria

- AC1 — `flow-plan` «Acceptance criteria» pide, para la observación de datos, el valor esperado y una fuente que no corre el código bajo prueba · falso hoy porque la sección no lo menciona · Observed in: `content/skills/flow-plan/SKILL.md`
- AC2 — La spec `workflow-lifecycle` «Criteria close on the real effect» y la página `site/.../flows/plan.md` dicen lo mismo, y la spec tiene un escenario de un reporte · falso hoy porque ninguna de las dos lo dice · Observed in: `_support/openspec/specs/workflow-lifecycle/spec.md` y `site/src/content/docs/flows/plan.md`
- AC3 — La lectura requerida de plan y build no supera la línea base · falso si el cambio solo agregara texto · Observed in: `go test ./tests/content -run TestRequiredProcedureReadingFootprint`

## Places checked without change

`flow-verify` paso 1 (ya verifica cada criterio por su observación), la plantilla `proposal.md`, `hive-verify-task`, `hive-verify-change`, `check_change.py` (no puede verificarse con un script), `hive-rule-tests` punto 4 (el mismo principio para los tests), `content/guidance/global.md` (sin cambio de tamaño: 0 B), `_support/docs/architecture/` y `tests/pilot/` (no describen la observación de los criterios).
