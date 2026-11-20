# analysis-ending-no-replies — El auditor marca un análisis que termina sin bloque de respuestas

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `0c7381c5` · rama `fix/analysis-ending-no-replies` en `.claude/worktrees/analysis-ending-no-replies` |
| Playbook | ninguno (revisión de una sesión real, luego `/flow-build` tecleado por el usuario) |

## Decisions

2026-11-20
- **A** — ¿Cómo seguimos con la segunda observación? Corregir primero el detector `analysis_reply_block`, como pide «Measurement» cuando el watch item nombra un mecanismo; la regla `hive-rule-report-endings` no cambia. Descartado: corregir también la regla, o solo registrar.
- **Supuesto** — Qué cuenta como un análisis sin bloque: el último texto de un turno de la sesión principal que no ofrece respuestas en lista, tiene 6 o más líneas no vacías, no lleva `▶`, pregunta en negrita, comando de flujo, tarjeta ni ronda en texto, y cuyo turno no escribió ni borró archivos fuera de la carpeta temporal. Descartado: marcar cualquier final sin bloque, que incluye respuestas cortas y cierres de build. Se revierte quitando la rama nueva del criterio. El umbral se mide contra sesiones reales antes del cierre.

2026-11-20 (revisión tras medir)
- **Supuesto** — La rama nueva solo evalúa un turno que cargó `hive-mode` (llamada a la skill o lectura de su `SKILL.md` instalado); se conservan las 6 líneas y la exclusión de turnos con escrituras o entregas. Medido sobre 249 sesiones reales del 2026-11-13 al 2026-11-20 (191 de Codex, 58 de Claude Code): con solo las 6 líneas marcaba 239 finales, casi todos líneas de espera, progreso de un build o cierres; con «investigación abierta desde `hive-mode` hasta el siguiente flujo» marcaba 9, entre ellos 3 mensajes de progreso de un build; con el turno que cargó `hive-mode` marca 5: kin, dos revisiones de ticket de cam, una respuesta de Jira de fleetsystem y una revisión adversarial, todas sin bloque, y una de ellas (CAM-25) cierra con una pregunta de aclaración que la regla podría aceptar. Descartado: los otros dos criterios, por sus falsos positivos. No ve un análisis de un turno posterior de la misma investigación, que no vuelve a cargar `hive-mode`. Se revierte quitando la condición `hive` del criterio.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «A» — usuario, 2026-11-20, a la opción «Ajusto el detector `analysis_reply_block` para que también marque un análisis que no cierra con bloque ▶ ni con bloque de siguiente paso, con un caso de prueba sacado de esta sesión».

Origen: la sesión de Codex `01a1f7d9-b225-7b12-bfa1-7fa2b09db920` (proyecto kin, `gpt-6.2-sol`, en el floor) terminó un análisis en prosa, «Queda pendiente tu visto bueno para dibujar las cuatro vistas en Figma.», sin bloque ni respuestas; `go run ./tests/pilot --host codex --audit-session` devolvió `not_observed` para `analysis_reply_block`. Watch: `engram:tricell-hive/watch/analysis-ending-without-reply-block` (cuenta 2).

## Objective

Hoy `analysis_reply_block` solo evalúa un final que ya ofrece respuestas en lista; un análisis que termina sin ninguna respuesta pasa sin verse. Al terminar, el criterio también falla ese final y sigue sin evaluar las respuestas cortas, los turnos que editan archivos y los finales con tarjeta, ronda o comando de flujo.

## Scope and acceptance

Dentro: `tests/pilot/session_audit_endings.go`, sus pruebas y fixtures, el requisito de la spec `workflow-lifecycle`, `rule-enforcers.md`, la rúbrica de revisión de sesiones y el changelog. Fuera: cambiar `hive-rule-report-endings` u otra regla distribuida, y otros criterios del auditor.

- AC1 — Un turno de la sesión principal que cargó `hive-mode` y cuyo último texto es un informe de 6 o más líneas no vacías sin respuestas, sin `▶`, sin pregunta en negrita y sin comando de flujo, en un turno que no escribió archivos, hace fallar `analysis_reply_block` con evidencia que dice que terminó sin bloque y si se leyó la regla · false today because el rollout de kin da `not_observed` · Observed in: `go run ./tests/pilot --host codex --audit-session` sobre el fixture de Codex
- AC2 — Un final corto (menos de 6 líneas no vacías), un turno que escribió un archivo del repositorio, un turno que no cargó `hive-mode` y un final con bloque de siguiente paso no se evalúan por esa rama · false today because la rama no existe, así que no hay prueba que lo fije · Observed in: `go test ./tests/pilot -run 'NewCriteria|ReplyList|AnalysisReply'`
- AC3 — Sobre las sesiones reales recientes de Claude Code y Codex de esta máquina, la rama nueva marca el final de kin y sus otros fallos son finales que la regla también pide cerrar con bloque, o el umbral se ajusta hasta que lo sean · false today because la rama no existe · Observed in: el conteo sobre `~/.claude/projects` y `~/.codex/sessions`, registrado en `tasks.md`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `tests/pilot/session_audit_endings.go:158` `replyListWithoutBlock` devuelve `listed=false` sin dos ítems, y `analysisReplyBlock` (línea 232) salta ese turno. Leído el 2026-11-20 en `0c7381c5`.
- Requisito «The session audit flags an analysis that ends without the reply block» en `_support/openspec/specs/workflow-lifecycle/spec.md:1406`.
- El cambio archivado `2026-11-20-0651-floor-versions-reply-block` midió 200 sesiones para elegir la señal de la lista.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` (respuesta de la sesión, 2026-11-20): revisión, push, PR y merge sin parar. Revisión de código: nativa, `/code-review` de Claude Code. No autoriza cambiar reglas distribuidas ni desplegar contenido.

## Review status and progress

2026-11-20 — Carpeta escrita para un cambio sin plan que altera un comportamiento descrito por la spec; sin `hive-review-plan`, como fija la referencia «spec delta».
