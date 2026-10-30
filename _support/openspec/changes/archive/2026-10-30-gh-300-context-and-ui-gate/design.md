# Diseño

## Verified context

Leído el 2026-10-30 sobre `83c9d3b5` (`development`).

- **Costo del contexto (hecho).** Análisis local de 11 sesiones de un ticket en Claude Code, Codex, Grok, Cursor y Pi, del 2026-10-15 al 29. Releer el contexto pesa 49–60% del costo en Claude Code, 71–80% en Codex y Cursor y 76–89% en Grok y Pi. Las corridas proporcionales estimadas cuestan 41–90% del real, casi todo por compactar en cada fase. El reporte está en `_support/sessions/2026-10-29-token-consumption/token-consumption.report.html`.
- **Compactar es seguro (hecho).** En 27 compactaciones reales, 26 salieron limpias y 1 tuvo una pérdida menor (una escritura rechazada). Ninguna repitió preguntas ni contradijo decisiones. En los 25 casos con trabajo posterior había un registro de cambio o una entrada de Engram, y el agente lo usó. Reconstruir el contexto costó 3–67 k tokens.
- **Controles de los hosts (documentación y `--help`, versiones instaladas).**
  - En ninguno de los seis hosts (Claude Code 2.1.314, Codex 0.165.0, Grok 1.0.73, Pi 1.4.0, OpenCode 2.2.1, Cursor 2026.10.22) el agente puede compactar su propia sesión con una herramienta; es un comando del usuario: `/compact`, o `/summarize` en Cursor. Pi lo permite solo con una extensión (`ctx.compact()`), que no trae.
  - Umbrales por defecto: Claude Code ~967 k con modelos de 1 M, Codex ~90% de la ventana, Grok 85%, y Pi y OpenCode cerca del límite.
  - Solo Claude Code y Pi aceptan instrucciones al compactar.
  - Claude Code documenta que compactar con el caché tibio cuesta poco y después de la vida del caché reprocesa el historial sin caché (https://code.claude.com/docs/en/prompt-caching, leído el 2026-10-29).
  - `_support/docs/architecture/compaction-survival.md` registra qué sobrevive en cada host.
- **Texto vigente que la regla no debe contradecir (hecho).**
  - `content/skills/flow-close/SKILL.md:55`: un flujo en curso nunca se parte.
  - `content/skills/flow-plan/references/plan-handoff.md:9`: nunca recomendar construir el plan recién hecho en otra sesión.
  - `content/guidance/global.md:68`: tras un resumen de compactación, releer solo la referencia cuya condición aplica y el `SKILL.md` activo.
  - `global.md:73`: lecturas completas.
  - `global.md:83`: cadencia de memoria.
  - `flow-close/references/work-item-boundary.md`: no hay parada entre tareas.
  - Preferencia del usuario (memoria del host, 2026-09-22): no cortar la sesión entre plan y build; un `/compact` manual solo como opción del usuario.
- **UI-gate (hecho).**
  - `content/skills/flow-verify/references/ui-gate.md:7` ya excluye el backend sin efecto visual y los cambios solo de texto o estilo.
  - `content/skills/flow-plan/references/ui-verification-decision.md:7` condiciona la pregunta, pero no exige declarar la exclusión.
  - En 2 sesiones independientes con la versión vigente de la condición (sin cambios desde `d1cc2a97`, 2026-10-26), el agente abrió la pregunta y lanzó `hive-verify-change` en trabajo sin efecto visual o solo de texto:
    - Grok cam `01a1855f-cdbe-7200-8434-7323a59ff52b` (COM-901, 2026-10-28): backend, cuyo plan decía «The plan changes no screen». Costó $6.58 más ~$2.7 del hilo principal.
    - Grok cam `01a18abb-b9c6-7db3-a80a-31bbc5fad8fb` (COM-838, 2026-10-29): el texto de un aviso. Hubo 5 recorridos, 4 de ellos reintentos.
    - Codex cam `01a180de` y Cursor cam `afaeb558` reintentaron recorridos sobre un candidato sin servir, pero su trabajo sí tenía UI. Son otro patrón y no cuentan aquí.
  - La lectura de la decisión está condicionada a «a rendered UI effect» en `ui-verification-decision.md:5`, `delivery-decisions.md:40`, `flow-build/SKILL.md:25,27` y `flow-verify/SKILL.md:20`, así que en trabajo de backend la decisión no se lee. La línea de no aplica debe vivir en la ronda que sí corre siempre: `delivery-decisions.md`.
- **Cursor (hecho).** El bundle instalado de `cursor-agent` 2026.10.22 (`~/.local/share/cursor-agent/versions/2026.10.22-27dcf5f/`) construye las fuentes de subagentes con `computeAgentsDirs()`, que devuelve `<workspace>/.cursor/agents` y, con la extensibilidad de terceros, `<workspace>/.claude/agents` y `<workspace>/.grok/agents`. Ninguna fuente de usuario carga agentes. En la sesión de Cursor del 28-oct los roles corrieron como hijos genéricos con el contrato incluido, sin efecto observado en los resultados.

## Proposed design

**S1 — Dónde actúa la regla de contexto.** Solo el usuario puede compactar y solo cuando el turno del agente termina. Por eso la regla actúa en los turnos que terminan esperando al usuario en un límite de fase:
- la pregunta de ruta al entregar un plan listo (`plan-handoff.md`);
- la parada de validación del usuario en verify (`flow-verify/SKILL.md`, «Validation by the user»);
- la pregunta de cierre o pausa con un flujo todavía en curso (`flow-close/SKILL.md`), donde hoy se recomienda continuar y no partir.

El paso de build a verify no termina el turno, así que no lleva puntero; se desvía de la letra de AC2 de #300 por esa razón. La recomendación de sesión nueva que `flow-close` ya da al cerrar, cuando no queda flujo en curso, no cambia.

**S2 — Qué dice la regla.** Referencia nueva `hive-core/references/context.md`, canónica:
- **Condición:** un turno termina esperando la respuesta del usuario en uno de los límites de S1, y el contexto usado pasa de aproximadamente la mitad de la ventana del modelo, según lo muestra el host o según tu estimación de la sesión.
- **Acciones, en orden:**
  1. Confirmar que el estado para retomar ya está en disco: el progreso y el siguiente paso del registro de cambio, o, sin plan retenido, lo que el mensaje ya resume. No se escribe memoria nueva para esto, porque rige la cadencia de memoria de la guía global.
  2. Agregar al mensaje, antes de la línea de pausa, una sola línea que sugiera la capacidad de compactación del host, nombrando el comando conocido como ejemplo (`/compact`, o `/summarize` en Cursor). La línea dice que compactar enseguida, con el caché tibio, cuesta menos que después de una espera larga. Cuando la pregunta va en una tarjeta nativa que espera la respuesta (D8-A, 2026-10-30), la línea agrega cómo hacerlo: descartar la tarjeta, correr el comando y luego escribir cualquier mensaje, o responder directamente. El agente vuelve a preguntar en texto, como ya pide la regla de tarjeta descartada de `questions.md`. `questions.md` no cambia.
- **Límites:** la línea no es una pregunta, no lleva opciones, no bloquea y no reemplaza la línea de pausa. No parte el flujo ni la sesión, y no se repite en turnos que no son límite de fase.
- **Tras una compactación:** aplica la regla de relectura de la guía global; esta referencia no la restate.

Se nombra la capacidad, con los comandos como ejemplo, porque `AGENTS.md` pide capacidad más fallback.

**S2b — Puntos cerrados.** `context.md` nombra los tres puntos de S1 como lista cerrada y excluye de forma explícita todos los demás:
- las rondas de diseño y de entrega;
- la pregunta de ruta de `flow-research`, porque sus hallazgos viven solo en la conversación y compactar los perdería;
- los turnos de un flujo en curso que no son límite de fase.

**S3 — Condición de lectura.** `hive-core/SKILL.md` agrega: «Before ending the turn at a plan handoff, at verify's user validation stop, or at a close question while a flow is still in progress, read [context](references/context.md)». Los tres punteros de S1 son una línea cada uno. `flow-close/SKILL.md` está en 15,367 bytes de 16,000 y `flow-plan/SKILL.md` no se toca.

**S4 — UI-gate.** La línea vive en `delivery-decisions.md:40`, la ronda que corre para todo trabajo. Antes de resolver el valor de `UI verification`, se decide si la gate aplica según el alcance de `ui-gate.md`. Cuando no aplica, la línea de la ronda dice que la UI-gate no aplica y por qué, y no se pregunta el nivel; verify no lanza los hijos de la gate. Los casos son trabajo de backend o infraestructura sin efecto visual, o un cambio solo de texto o de un valor de estilo que no altera layout, estado ni flujo. `ui-gate.md` (Scope) agrega el ejemplo que separa el caso límite observado: el texto de un mensaje, aviso o notificación es un cambio de texto. El trabajo sin entrega Git no pasa por `delivery-decisions.md` y llega directo a `ui-verification-decision.md:7`. Por eso `flow-build/SKILL.md:27` recibe un puntero incondicional de una frase: antes de esa ronda, decidir si la UI-gate aplica como dice `delivery-decisions.md`. Las copias de `flow-build/SKILL.md:25`, `flow-verify/SKILL.md:20` y `ui-verification-decision.md:5,7` son condiciones de lectura coherentes y no cambian; los demás caminos ya leen `delivery-decisions.md` (re-revisión R1). Es una corrección con recurrencia cumplida: 2 sesiones independientes con la versión vigente.

**S5 — Cursor.** `agent-delivery.md` actualiza cuatro lugares: la fila de destino (línea ~114), la nota de ubicaciones (~118), la fila de `Task` (~167) y la fila de observaciones (~197). `README.md:211` («User-level role selection remains unverified.») pasa a decir lo observado. Registran lo que carga 2026.10.22, que los roles de usuario no aparecen, que rige el fallback de `delegation.md` sin efecto observado en resultados y las alternativas descartadas por D5-A. La documentación de Cursor (2026-09-25) que nombra `~/.cursor/agents` queda como conflicto documentado.
