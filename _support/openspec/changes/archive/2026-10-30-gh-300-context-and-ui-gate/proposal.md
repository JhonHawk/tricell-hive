# Contexto por fase, UI-gate sin efecto visual y límite de roles en Cursor

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por PR #301 (merge 0ade8ca9, 2026-10-30) |
| Tracker · GitHub Issues | • [#300 — El contexto releído domina el costo de cada ticket y la UI-gate se abre en cambios sin efecto visual](https://github.com/JhonHawk/tricell-hive/issues/300) |
| Git | automatic · PR a `development` y merge · sin parada humana |
| Verificación | tests de contenido · verificador C1 · suite completa · sin revisión dedicada (D7-C) |
| Siguiente paso | Ninguno; archivado |

## Request

> Haz research sobre todo los hosts y todas las sesiones posibles, considerando breakpoints o disparos en el consumo y si lo genero algun cambio nuestro y genera un reporte html con gráficas comparativas al respecto, por favor

> La premisa tambien es ver si la chamba de los validadores realmente tiene valor, y catalogarlo (tambien con grafica)

> y si da pistas a que si, extendamos las revisiones a las demas sesiones de los hosts

Respuestas de la sesión (2026-10-29/30):
- D2-A, D2-B y D2-C. En D2-A: «en grok, dejemoslo por default, no forcemos la config sin fast y 256k, por si en algun momento los desarrolladores cambian esas configuraciones, lo notemos».
- D3-A (regla de contexto en Hive), D4-A (corregir la UI-gate y vigilar lo demás) y D5-A (documentar el límite de Cursor y mantener el fallback).

Criterios de aceptación de [#300](https://github.com/JhonHawk/tricell-hive/issues/300), citados textualmente (leído el 2026-10-30):

> - AC1: una regla de manejo de contexto en `hive-core` dice cuándo sugerir la compactación al cerrar una fase y con qué capacidad del host. No agrega preguntas y no contradice «un flujo en curso nunca se parte» (`flow-close`) ni la regla de relectura tras compactación (`global.md`).
> - AC2: los mensajes de cierre de fase de `flow-plan`, `flow-build` y `flow-verify` apuntan a esa regla.
> - AC3: la UI-gate y la decisión de verificación de UI excluyen, sin ambigüedad, los cambios sin efecto visual y los que solo cambian texto o estilo, antes de preguntar el nivel de verificación de UI.
> - AC4: `agent-delivery.md` registra el límite de Cursor CLI 2026.10.22 y el fallback.

## Objective

Releer el contexto acumulado es entre el 49% y el 89% del costo de un ticket en los seis hosts. Las sesiones llegan a 300–966 k tokens antes de que el host compacte solo, y ninguna regla de Hive sugiere compactar al cerrar una fase. Compactar resultó seguro: 26 de 27 compactaciones reales salieron limpias, porque el estado vive en el registro de cambio. Pero el agente no puede compactar por sí mismo en ningún host; solo el usuario puede lanzar el comando, y solo cuando el turno del agente termina. El cambio agrega una regla que, en los turnos que terminan esperando al usuario en un límite de fase, sugiere en una línea el comando de compactación del host cuando el contexto pasa de la mitad de la ventana.

Además, la UI-gate se abrió en dos sesiones independientes de Grok, del 28 y el 29 de octubre, para trabajo de backend o solo de texto, aunque la regla vigente lo excluye. #300 dice 4 sesiones; al verificarlas, dos eran de UI real. El cambio obliga a decidir y declarar en una línea si la gate aplica, en la ronda de entrega y antes de preguntar el nivel de verificación de UI. Por último, registra que Cursor CLI 2026.10.22 no carga los roles de `~/.cursor/agents`.

## Scope and acceptance

Incluido: una referencia nueva `content/skills/hive-core/references/context.md` y su condición de lectura en `hive-core/SKILL.md`; punteros de una línea en `flow-plan/references/plan-handoff.md`, `flow-verify/SKILL.md` (validación del usuario) y `flow-close/SKILL.md` (pregunta de cierre con un flujo en curso); la ronda de entrega (`flow-plan/references/delivery-decisions.md`), la decisión de verificación de UI (`flow-plan/references/ui-verification-decision.md`) y el alcance de la gate (`flow-verify/references/ui-gate.md`); `_support/docs/architecture/agent-delivery.md`, `instruction-resources.md` y `README.md`; las páginas `site/src/content/docs/flows/plan.md`, `verify.md` y `close.md`; el escenario del ciclo en `tests/content/reading_footprint_test.go`.

Excluido, con destino:
- La cadencia de escritura en memoria: la causa es el protocolo del plugin de Engram y los hooks (Engram `watch/memory-without-memory-md`, decisión D9 de no cambiarla).
- Umbrales de auto-compactación en la configuración de los hosts: se respetan los defaults (D2-A, D3-A).
- Instalar los roles de Cursor por proyecto o como plugin (D5-A).
- Otros recortes de validadores, que quedan como vigilancias en Engram (D4-A): revisor UX fuera de su disparador, lint omitido antes del push con la regla de #278, checkpoints de una sola tarea y consultas de `hive-design-architecture` sin disparador.
- AC2 de #300 nombra `flow-build`, pero el paso de build a verify no termina el turno, así que el usuario no puede compactar ahí. El puntero va en los tres turnos de límite de fase que sí esperan al usuario: la entrega del plan, la validación de verify y la pregunta de cierre. Ver la decisión S1 de `design.md`.

Búsqueda de especificaciones: se buscaron `compact`, `context`, `UI gate`, `UI verification`, `Cursor` y `agents` en `_support/openspec/specs/` (capacidades `context-delivery`, `documentation-site`, `versioned-installation` y `walkthrough-viewer`). Solo `context-delivery` toca la entrega de guía, y su requisito habla de canales de entrega y del puntero de proyecto de Cursor, no de conducta de los flujos. El cambio altera la redacción de la guía y la documentación interna, sin requisitos de capacidad, así que no lleva delta.

- AC1. `content/skills/hive-core/references/context.md` dice que, en un turno que termina esperando al usuario en un límite de fase y con el contexto usado por encima de la mitad de la ventana del host, el mensaje agrega una sola línea que sugiere la capacidad de compactación del host, después de dejar el estado en disco; que no es una pregunta, no parte el flujo ni la sesión y no escribe memoria. *Falso en la base cuando* el archivo no existe.
- AC2. `plan-handoff.md`, la sección de validación de `flow-verify/SKILL.md` y la pregunta de cierre de `flow-close/SKILL.md` apuntan a esa referencia, y `hive-core/SKILL.md` lista su condición de lectura. *Falso en la base cuando* `rg -n "references/context.md" content` no devuelve nada.
- AC3. La ronda de entrega (`delivery-decisions.md`), que corre para todo trabajo, exige declarar en una línea que la UI-gate no aplica, y no preguntar el nivel de verificación de UI. Aplica al trabajo de backend o infraestructura sin efecto visual y al cambio que solo altera texto o un valor de estilo sin cambiar layout, estado ni flujo. El alcance de `ui-gate.md` nombra el texto de un mensaje, aviso o notificación como cambio de texto. *Falso en la base cuando* `delivery-decisions.md:40` solo condiciona la pregunta («for a rendered UI effect») y `ui-gate.md` no tiene ese ejemplo.
- AC4. `agent-delivery.md` y `README.md` registran que Cursor CLI 2026.10.22 solo carga subagentes de `<workspace>/.cursor/agents` (y de `.claude/agents` y `.grok/agents` del workspace con la extensibilidad de terceros), que los roles de `~/.cursor/agents` no aparecen en `Task` y que rige el fallback genérico. *Falso en la base cuando* la línea de observaciones de Cursor dice «not yet observed».

## Delivery

Respuestas del 2026-10-30: D6-A (`automatic`) y D7-C (sin revisión dedicada).
- **Ruta:** rama de trabajo desde `development` y commits locales. Pruebas de contenido por tarea, verificador de C1 y suite completa una vez. Luego push, PR a `development` y merge cuando pasen los checks, sin parada humana.
- **Revisión:** sin revisión dedicada. Solo la inspección del diff por el implementador y las pruebas. Con `automatic`, nadie revisa de forma independiente antes del merge.
- **Después del merge:** se refresca la instalación local de Hive (`hive update` desde este checkout), como manda la regla de mantenedor de `AGENTS.md`, y se reporta el commit desplegado.
- **Registro de cambio:** cierra con un PR pequeño sin revisión dedicada, que integra el agente.
- **Tracker:** #300 pasa a en progreso al iniciar el build y se cierra con la integración.
