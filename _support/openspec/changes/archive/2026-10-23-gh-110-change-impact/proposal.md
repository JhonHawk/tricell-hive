# Impacto proporcional antes de cambiar lo que corre en producción

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en [PR #111](https://github.com/JhonHawk/tricell-hive/pull/111), merge `1000ee23`, el 2026-10-23; desplegado en local |
| Tracker · GitHub Issues | • [#110 — Guía: análisis de impacto proporcional antes de cambiar componentes que corren en producción (legacy)](https://github.com/JhonHawk/tricell-hive/issues/110) |
| Git | Interactivo (D3-A) · `feat/gh-110-change-impact` → PR → `/code-review` → merge a `development` · parada antes del push |
| Verificación | `go vet` · `go test` · unittest de skills · `rg` por cláusula · `hive-verify-task` por tarea · `/code-review` (2 rondas) · `global.md` +102 bytes (T5) |
| Siguiente paso | Ninguno; sin promoción a `master` |

## Solicitud

> Analiza la sesión de facti con claude code, algo que creo no habiamos considerado en la etapa de research es el impacto, ese ejemplo es de una parte legacy, y tuve que redirigir varias veces en temas que deba y ebe revisar por la naturaleza legacy de lkos componentes. Deep research y dame una propuesta

> recuerda que esto no es una regla fija, hay puntos que pueden o no escalar dependiendo la complejidad d el atarea

Decisiones del usuario (2026-10-23):

- D1-A: planear P1–P4 con `flow-plan`.
- D2-A: abrir la incidencia #110 (abierta).
- Restricción: la regla no es fija; cada paso escala según la complejidad y el riesgo de la tarea.

La incidencia #110 no declara criterios de aceptación propios; los criterios salen de su «Acción sugerida» (leída el 2026-10-23).

## Objetivo

En tres sesiones reales de Claude Code del 2026-10-23 sobre un API legacy sin pruebas ni CI (`42b8e06a`, `ccbed8a5`, `c574e5a1`), el agente recomendó o planeó cambios sobre componentes que corren en producción sin preguntar qué dependía de ellos en ejecución: recomendó cerrar el acceso de red a una base de datos que el propio API usa por su IP pública, propuso retirar endpoints sin poder descartar llamadores externos y no consideró la separación entre el servidor del API y el de crons. El usuario tuvo que redirigirlo, y el code-review y la revisión previa al despliegue hallaron tarde caminos rotos o inseguros que el cambio volvía alcanzables.

Este cambio da a Hive una referencia única de impacto del cambio, condicionada a señales y proporcional: una búsqueda barata del valor afectado siempre que aplica, y pasos más caros (topología, caminos alcanzables, observación de solo lectura) solo cuando aparece su señal. La investigación, el plan, la revisión del plan, el build y el code-review la enlazan en el momento en que actúan, y la regla de interfaces de `flow-plan` deja de desalentar la búsqueda de consumidores sin perder su intención contra la sobreingeniería.

## Alcance y aceptación

Incluye:

- P1: `content/skills/flow-research/references/change-impact.md`, hogar canónico.
- P2: reconciliar la regla de interfaces en `content/skills/flow-plan/SKILL.md` y su copia en `content/skills/flow-build/SKILL.md`.
- P3: punteros con condición de lectura desde `flow-research`, `flow-plan`, `flow-build`, `plan-review.md` y `flow-build/references/data-changes.md`.
- P4: `content/agents/review/hive-review-code.md` cuenta como introducido un defecto previo que el cambio vuelve alcanzable, y `flow-build/references/verification.md` no lo devuelve a la regla de incidentales.
- Las páginas `site/src/content/docs/flows/research.md` y `plan.md`, y una entrada en `CHANGELOG.md`.

Excluye, con su motivo:

- Un ajuste `Legacy` en `## Hive`: costaría código del gestor y las señales ya se observan en el repositorio.
- Un revisor dedicado de impacto: no hay evidencia de que reduzca regresiones y suma costo.
- Pruebas de caracterización obligatorias: ya existen para refactors; exigirlas aquí sería ceremonia.
- Cambios en la regla de incidentales de `content/guidance/global.md`: ya cubre «defects that affect the current work». El único cambio en `global.md` es la excepción de T5 (D7-A).
- Cambios en facti: es otro proyecto.

Criterios:

- AC1. Existe `content/skills/flow-research/references/change-impact.md` con: una sección «When to read» con el disparador acotado (quitar, cerrar, desactivar, migrar, renombrar o mover a otro host o proceso algo que corre en producción; o un cambio de comportamiento en código de producción sin pruebas ni CI que lo cubran) y la exclusión del renombre de un símbolo que solo usa este código; el paso 1 (buscar el valor literal y, en un cambio de comportamiento, cada llamador), que siempre corre; tres pasos con su señal (mapear hosts y procesos y quién lee o escribe; leer los caminos que el cambio vuelve alcanzables; marcar como no verificado lo no descartable, con observación de solo lectura, ventana y reversión); y la frase de que un cambio sin señales no lleva más pasos. *Falso en la base cuando* el archivo no existe.
- AC2. `content/skills/flow-research/SKILL.md` enlaza `references/change-impact.md` con su condición de lectura antes de recomendar un cambio así. *Falso en la base cuando* `rg change-impact content/skills/flow-research/SKILL.md` no devuelve nada.
- AC3. `content/skills/flow-plan/SKILL.md` tiene un puntero de lectura a la referencia junto a las lecturas de seguridad y datos, y su regla de interfaces pide buscar consumidores antes de aplicarse y declara que una observación acotada solo decide entre cambio coordinado y transición y nunca retira una transición. *Falso en la base cuando* `rg change-impact content/skills/flow-plan/SKILL.md` no devuelve nada.
- AC4. `content/skills/flow-build/SKILL.md` tiene la condición de lectura para trabajo sin plan; su reconciliación devuelve al plan un consumidor o camino alcanzable que la referencia encuentra y el plan no contempla; y su copia de la regla de interfaces distingue la compuerta de ausencia de una observación que solo eligió el enfoque. *Falso en la base cuando* `rg change-impact content/skills/flow-build/SKILL.md` no devuelve nada.
- AC5. En `content/skills/flow-plan/references/plan-review.md`, las filas Backend / data / integration e Infrastructure / delivery enlazan la referencia para revisar la lista de impacto del plan. *Falso en la base cuando* `rg -c change-impact` sobre ese archivo devuelve 0.
- AC6. `content/agents/review/hive-review-code.md` paso 3 reporta como introducido un defecto previo que el cambio vuelve alcanzable o amplía, y `content/skills/flow-build/references/verification.md` no lo devuelve a la regla de incidentales. *Falso en la base cuando* el paso 3 excluye todo «code that was already there».
- AC9. `content/guidance/global.md` declara que renombrar o mover algo que lee o fija un host, un proceso o una configuración fuera del código no es un cambio mecánico (D7-A). *Falso en la base cuando* la definición de cambio mecánico de `global.md:73` incluye todo renombre y movimiento sin excepción.
- AC8. `content/skills/flow-build/references/data-changes.md` remite a la búsqueda de lectores y escritores de la referencia. *Falso en la base cuando* `rg change-impact` sobre ese archivo no devuelve nada.
- AC7. `site/src/content/docs/flows/research.md` y `plan.md` describen el chequeo de impacto proporcional sin copiar sus reglas, y `CHANGELOG.md` lo registra en `[Unreleased]` con `#110`. *Falso en la base cuando* ninguna de las tres fuentes lo menciona.

Restricciones que deben seguir siendo ciertas:

- `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan; la validación de enlaces de la release resuelve la referencia nueva.
- `content/guidance/global.md` solo cambia por T5 (D7-A); el PR declara su tamaño.
- La regla de interfaces conserva su intención: un consumidor que no se buscó ni se observó no justifica una transición por fases.
- La referencia no nombra herramientas de un host concreto.

## Entrega

Decisiones del 2026-10-23:

- D3-A, interactivo: rama `feat/gh-110-change-impact` desde `development`, commits locales y verificación; parada antes del push para que el usuario lea la redacción (despliegue local desde la rama solo si lo pide). Tras su validación: push, PR, `/code-review`, merge a `development` y refresco de la instalación local (marca de mantenedor). La carpeta de cambio se versiona al cierre con un PR pequeño sin revisión que el agente fusiona. Sin promoción a `master`.
- D4-A: `/code-review` de Claude Code sobre el diff antes del merge.
- D5-A: `flow-build` con esta carpeta.
- D7-A (tras la re-revisión de `/code-review`): añadir a `global.md` que renombrar o mover algo leído o fijado fuera del código no es mecánico.

El commit nombra las sesiones de origen (host Claude Code, `42b8e06a-80ad-410b-8cba-fe4feadb5165`, `ccbed8a5-93cf-4c6c-b667-f9387bec1e2e`, `c574e5a1-ec3e-44da-995e-2a9a89ba2e6b`, 2026-10-23), como pide `AGENTS.md` («Measurement»).
