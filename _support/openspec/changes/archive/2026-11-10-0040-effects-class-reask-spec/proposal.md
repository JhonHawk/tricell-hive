# effects-class-reask-spec — La spec recoge que una clase de efecto cambiada se vuelve a preguntar

| Campo | Valor |
| --- | --- |
| Estado | Completado |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `b2cb3ec7` · rama `fix/effects-class-reask-spec` |

## Solicitud

Del usuario, 2026-11-09: resolver el hallazgo H10 del cierre de `off-repo-facts-and-effects`.

## Objetivo

`content/skills/flow-plan/references/outside-the-repository.md` ya establece, desde el PR #446, que cuando una revisión del plan cambia el tipo de acción, el destino, el techo o las tareas de una clase aprobada, su respuesta vuelve a «sin responder» y la clase se pregunta de nuevo antes de otro efecto. El requisito «A plan authorizes foreseen effects by class» de la spec `workflow-lifecycle` no lo decía. Este cambio lo añade, con su escenario; no cambia el contenido distribuido.

## Criterios

- **AC1** — El requisito «A plan authorizes foreseen effects by class» establece la vuelta a «sin responder» y la nueva pregunta, con un escenario, igual que la referencia. Observed in: el repositorio.
