# Opción mínima y tamaño en tiempo del agente en las preguntas que deciden cuánto se construye

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #260](https://github.com/JhonHawk/tricell-hive/pull/260) (`48a616b8`) · desplegado localmente con `hive update` |
| Tracker · GitHub Issues | • [#259 — Las preguntas de alcance no ofrecen la opción mínima y el tamaño estimado mide esfuerzo humano](https://github.com/JhonHawk/tricell-hive/issues/259) |
| Git | automatic · PR a `development` y merge · registro en un PR de cierre pequeño que integra el agente · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · `go vet ./...` · `/code-review` de Claude Code |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> aun asi, son 10 horas para este ticket https://linear.app/tricell/issue/COM-847/retirar-desde-la-plataforma-una-cuenta-o-un-numero-de-whatsapp-que-ya

> D1-B Corregir la guía ya

> y tambien corrijamos lo de los tiempos, porque los da como si lo fuera a hacer un humano y en varias sesiones me da 2-3h o mas y lo resuelve en menos de 10min

> Tiempo del agente, sin embargo si esto dispara sub-investigaciones para calcular ese tiempo, mejor retiremoslo y que ese tiempo de investigación se dedica a algo más

El issue #259 se abrió el 2026-10-28 desde esta sesión; sus criterios son los AC de abajo.

## Objetivo

En la sesión de Codex `01a180de-c680-7f91-81d7-6c50a9e02ef1` (Codex CLI 0.165.0, 2026-10-27/28, proyecto cam, ticket COM-847), el plan recomendó en cada pregunta la opción más pesada (D2-A, D3-A, D4-A y D7-A). Nunca ofreció una opción que resolviera solo el problema del ticket, y la ruta de build (D8) solo permitía construir AC1–AC9 o conservar el plan, sin tamaño. El ticket pedía un endpoint transaccional de soft-delete. Salieron 22 commits, 182 archivos y ~30.6k líneas (~6.6k de fuente escrita a mano) en ~9 h activas. El usuario observa además que, cuando se da un rango de tiempo, se calcula como esfuerzo de una persona (2-3 h) para trabajo que el agente termina en menos de 10 min.

El cambio hace que una pregunta que decide cuánto se construye ofrezca siempre la opción mínima suficiente y justifique una Recomendada mayor. También hace que la ruta de build del plan dé siempre el tamaño del plan, y que el rango de tiempo mida el tiempo del agente con evidencia ya disponible, sin investigar solo para estimarlo.

## Alcance y aceptación

Incluye una referencia nueva de `hive-core` que recibe la condición de tamaño de `questions.md` (el archivo está en 15,992 B, al límite de 16,000) junto con la regla nueva de opción mínima, y los punteros a ella:

- `questions.md`, `flow-close/SKILL.md` y `plan-handoff.md`;
- el índice de `hive-core/SKILL.md`;
- la lista de hogares canónicos de `_support/docs/architecture/instruction-resources.md`;
- la prueba de huella de lectura;
- las páginas `site/` de plan, research y close;
- `CHANGELOG.md`.

Excluye:

- Cambiar el disparador de 10× para las demás preguntas (D3-A).
- Tocar `content/guidance/global.md`: su regla de proporcionalidad ya pide el plan mínimo suficiente; el fallo está en cómo se formulan las opciones, cuyo hogar es la referencia de preguntas.

Restricciones:

- `go test ./...` y `go vet ./...` pasan.
- Cada `SKILL.md` y cada referencia quedan en 16,000 B o menos.
- `content/guidance/global.md` no cambia de tamaño (0 B).
- La condición de tamaño se mueve entera, y `questions.md` conserva solo un puntero con su condición.

Búsqueda de requisitos: en `_support/openspec/specs/*/spec.md` se buscaron `question`, `size`, `estimat`, `option`, `smallest` y `route`. Solo hubo coincidencias ajenas: rutas de API y la lista de páginas de `documentation-site`, que no cambia porque no se agrega página. Se leyó el requisito de páginas de `documentation-site`. El cambio altera la redacción de la guía y no un requisito de capacidad, así que no lleva delta.

- AC1. Una pregunta en la que dos o más opciones empiezan trabajo sobre el mismo pedido y construyen cantidades distintas incluye la opción más pequeña que resuelve el problema del pedido, su ticket o un criterio aprobado. Si la Recomendada construye más, su razón nombra qué deja sin resolver la menor, en esos términos. La ruta de build y la pregunta de cierre no entran: su cantidad ya quedó fijada antes. *Falso en la base cuando* `rg -n "smallest option" content/skills/hive-core/references/` no encuentra nada.
- AC2. El rango de tiempo del tamaño estimado mide lo que tarda el agente de principio a fin, con hijos, pruebas y revisiones, y no el esfuerzo de una persona. *Falso en la base cuando* `rg -n "person's effort" content/skills/` no encuentra nada.
- AC3. El rango se basa solo en evidencia ya disponible; no se lee, busca ni delega solo para estimarlo, y sin esa base se escribe "time not estimated: no basis at hand". *Falso en la base cuando* `rg -n "no basis at hand" content/skills/` no encuentra nada.
- AC4. La opción `flow-build` de la ruta de build del plan siempre da el tamaño estimado del plan, aunque no se cumpla el disparador. *Falso en la base cuando* `rg -n "size" content/skills/flow-plan/references/plan-handoff.md` no encuentra nada.
- AC5. La página del plan en `site/` describe la opción mínima, el tiempo del agente y la ruta de build con tamaño, enlazando la referencia nueva. *Falso en la base cuando* `rg -n "smallest" site/src/content/docs/flows/plan.md` no encuentra nada.

## Decisiones

- D1-B (2026-10-28): corregir la guía ya; la aprobación del usuario sustituye la recurrencia de dos sesiones (`AGENTS.md`, Measurement).
- D2-A (2026-10-28): la regla de opción mínima vive en la referencia de preguntas (aquí, la referencia nueva que se separa de ella) y vale para plan, research y build.
- D3-A (2026-10-28): la ruta de build del plan siempre da el tamaño; el disparador de 10× no cambia.
- D4 (2026-10-28, respuesta libre): tiempo del agente, pero solo con lo que ya está a la mano; si calcularlo exige investigar, se omite el rango.
- D5-A (2026-10-28): entrega `automatic` a `development`.
- D6-A (2026-10-28): `/code-review` de Claude Code sobre el candidato local, antes del push.
- D7-A (2026-10-28): iniciar `flow-build` desde T1; aprueba AC1–AC5.

Decisión menor tomada sin preguntar: separar la referencia `hive-core/references/option-scope.md`, porque `questions.md` no admite el texto nuevo bajo el límite de 16,000 B y `AGENTS.md` pide dividir por responsabilidad.

## Entrega

- `tricell-hive`, base `development`, modo `automatic` (D5-A): rama de trabajo en un worktree, commits locales, `/code-review` sobre el candidato local (D6-A), push, PR a `development` y merge cuando pasen las pruebas locales. No hay CI (retirada 2026-09-30).
- Registro de cambio: se versiona una vez al cierre, en un PR pequeño sin revisión dedicada que integra el agente, como #258.
- Después del merge, actualización de la instalación local con `hive update` (checkout del mantenedor). No es un release.
- El issue #259 pasa a en progreso al empezar el build y se cierra al integrarse.
- El cuerpo del PR lista los lugares revisados (global, skills y referencias, roles, `site/`, `_support/docs/architecture/`, `tests/`) y dice que `content/guidance/global.md` cambia 0 B, como pide `AGENTS.md`.
- El mensaje del commit nombra la sesión Codex `01a180de-c680-7f91-81d7-6c50a9e02ef1` (2026-10-28) y el fallo observado, como pide `AGENTS.md`.
