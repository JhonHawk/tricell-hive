# Rehacer las skills de flujo con menos ceremonia

| Campo | Valor actual |
| --- | --- |
| Estado | Fase 0 hecha · F1: diseño híbrido estricto elegido (2026-10-30); pendiente el diagrama del flujo y las ceremonias |
| Tracker | Sin issue |
| Git | Rama `refactor/flow-ceremony` desde `bb9604c5` · sin commits · no desplegar hasta que el usuario lo indique |
| Verificación | Pruebas de contenido por fase · medición en sesiones reales en la fase 8 |
| Siguiente paso | Fase 1: el flujo objetivo y las decisiones transversales, para que el usuario decida |

## Request

> Vamos a hacer lo siguiente:
> - Realizar un inventario (.md) y una carpeta para respaldar el estado actual de las skills y reglas, o sea, literal, copiarlas ahí, pero sin tracking, solo son referencia
> - Basémonos un poco en el reporte que mencionas sobre qué rehacer, fusionar o quitar
> - Una vez con la copia, vamos a "limpiar" el contenido de las skills dejando el puro cascarón
> - Realicemos los movimientos de eliminación y fusión (o sea, solo dejemos los skills que tiene sentido que existan)
> - En paralelo planeemos en qué fases podemos ir atacando el refactor, o sea, primero research, luego plan, etc.

(2026-10-30, sesión Claude Code en tricell-hive.)

## Objective

Las skills de flujo actuales piden pasos cuyo costo supera lo que aportan, y les faltan los pasos baratos que habrían evitado los defectos. En la sesión Grok `01a18c56` (cam, COM-829, 2026-10-30) el ticket costó 8 % de la cuota semanal, y la UI llegó a la validación con los botones apilados que el revisor de UX ya había marcado como Medium. El objetivo es reconstruir `flow-research`, `flow-plan`, `flow-build`, `flow-verify`, `flow-walkthrough` y `flow-close` contra un flujo mínimo, trayendo de la línea base solo las reglas que pasen el filtro del inventario.

## Scope

Dentro:
- Las seis skills de flujo y sus referencias.
- Los roles de `content/agents/` que esos flujos lanzan.
- Los enlaces de `global.md` hacia ellas y los disparadores de `hive-core` de lectura excesiva.
- Las páginas de `site/`, `README.md` y `llms.txt` que describen esos flujos.

Fuera:
- `git-workflow`, `harness-audit`, `engram-init-workspace`, `security-advisory`, `flow-report` y `starlight-docs-site`, salvo los enlaces que apuntan a archivos rehechos.
- El gestor (`tooling/`), salvo lo que exijan las pruebas.
- Reanudar los pilotos de comportamiento, que siguen pausados.

Ya hecho en la fase 0:
- `workspace-archive` se fusionó en `workspace-conventions` (`references/archive.md` y `scripts/`).
- `flow-security` se fusionó en `flow-verify/references/security-review.md`.

## Acceptance criteria

- AC1: cada skill rehecha cumple el flujo objetivo de la fase 1, y cada regla de la línea base queda clasificada en el inventario de su fase (se queda, se acorta, se va, o pasa a otra parte), con su procedencia cuando la tiene.
- AC2: ninguna regla de las categorías 1 (seguridad y preservación) o 2 (falla real) se pierde sin una decisión del usuario registrada.
- AC3: el plan decide la región de acciones de una pantalla que gana acciones, y el build muestra la primera vista renderizada de un layout nuevo o reorganizado, con un paso barato antes de construir que no dependa de la verificación final.
- AC4: `go test ./...` y `tests/skills` pasan al cerrar cada fase, sin cascarones enlazados desde una skill ya rehecha.
- AC5: antes de integrar, una medición en sesiones reales compara la versión nueva con la línea base y con la ronda vanilla en tickets comparables: cuota, tokens, duración, llamadas del hilo principal y defectos que llegaron a la validación.

## Delivery

Sin decidir. La rama no se despliega mientras tenga cascarones. Commits, push y PR quedan para cuando el usuario los indique.

## Spec search

Términos: `flow-security`, `workspace-archive`, `verify-task`, `flow-verify`, `flow-walkthrough`, en `_support/openspec/specs/` el 2026-10-30. Solo coincidió `documentation-site/spec.md`: «Published documentation site» exige una página `flows/security/` que enlace el `SKILL.md` de su skill. La fusión de `flow-security` altera ese requisito, así que la fase 7 añade un delta MODIFIED o la página pasa a enlazar la referencia; decide el usuario.

## Context

- Inventario y respaldo (sin seguimiento): `_support/workspace/2026-10-30-skills-rebuild/inventory.md` y `baseline/`.
- Engram: `decisions/rule-removal-criterion`, `decisions/memory-residue-audit`, `watch/*`.
- Rúbrica de sesiones: `_support/sessions/2026-10-30-vanilla-harness-evaluation/vanilla-harness-evaluation.research.md` (sin seguimiento).
