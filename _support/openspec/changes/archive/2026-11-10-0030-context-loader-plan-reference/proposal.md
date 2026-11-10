# context-loader-plan-reference — La documentación del cargador de OpenCode nombra el recurso que el código entrega

| Campo | Valor |
| --- | --- |
| Estado | Completado |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `77e3417e` · rama `fix/context-loader-plan-reference-docs` |

## Solicitud

Del usuario, 2026-11-09, al cierre de `off-repo-facts-and-effects`: resolver el hallazgo H9, que señaló dos documentos de arquitectura que nombran un `plan-review.md` inexistente.

## Objetivo

El cargador opcional de OpenCode entrega `flow-plan/SKILL.md` al rol `hive-review-plan` desde el commit `7a37381b`, que retiró `flow-plan/references/plan-review.md`. Dos documentos de arquitectura y un requisito de la spec `context-delivery` seguían nombrando el archivo retirado. Este cambio los iguala al código; no cambia el código ni el contenido distribuido.

Queda fuera: que el cargador entregue referencias condicionales. Por diseño no las precarga; el rol `hive-review-plan` manda leer `outside-the-repository.md` por su ruta instalada cuando el plan toca un sistema externo.

## Criterios

- **AC1** — `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/deployment-manager.md` y el requisito «Fixed-role context delivery without authority expansion» nombran `flow-plan/SKILL.md` como recurso de `hive-review-plan`, igual que `integrations/opencode/context_snapshot.go`. Observed in: el repositorio.
