# determinism-candidates — Rastrear qué pasa del razonamiento a un mecanismo determinista, en el proyecto y en Hive

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #506, merge c1691248 |
| Tracker | none |
| Git | tricell-hive · base `development` en `8e71ca4b` · rama `feat/determinism-candidates` |

## Decisiones

2026-11-17
- **A** — ¿Quién detecta qué se puede volver determinista? Dos niveles con un mismo formato de candidato: la sesión de proyecto lo anota al cerrar su trabajo y lo cuenta en la memoria del proyecto; una sesión inspectora de Hive lee las sesiones reales con detectores deterministas, los seguimientos y los candidatos de los proyectos. (Engram `decision/determinism-candidates-two-levels`.)
- **D10 A** — ¿Dónde viven los detectores que leen sesiones reales? En `tests/pilot`, con adaptadores de transcripción real y un modo `--audit-session`, para que cada criterio tenga un solo dueño con sus fixtures.
- **D11 A** — ¿Qué marca el detector cuando el informe de cierre llega antes de borrar la rama fusionada? `review`, no `fail`: en el punto de completitud `flow-close` prescribe ese orden mientras queda una rama fusionada por borrar, y solo es falla en los cierres que no difieren nada; el inspector lee ese cierre para decidir. El criterio de reproducción de la revisión manual pasa a AC9.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> «Cómo rastrear e identificar situaciones que pasan del razonamiento (y con ello las variaciones de ejecución o revisión) a formas deterministas de ejecutarse.»
>
> «¿Sería una sesión padre (como esta) que se dedique a inspeccionar otras sesiones? ¿O puede ser el camino de ejecutar las skills en la sesión y que sea esta misma la que diga qué se puede automatizar? Porque hay dos niveles: qué se puede hacer de forma determinista para todo Hive, y qué para el proyecto en particular.»

## Objetivo

Hoy Hive convierte reglas en scripts y tests (la Fase 1 dejó siete), pero encuentra qué convertir a mano y de una sola vez. Los criterios deterministas de `tests/pilot/regression.go` solo leen la salida de los pilotos, que están pausados, así que las revisiones de sesiones reales se hacen con scripts improvisados que se reescriben cada vez. Ninguna sesión de proyecto anota los scripts que improvisó dos veces.

Al terminar, una sesión inspectora corre los criterios sobre transcripciones reales de Claude Code, Codex y Grok sin ejecutar ningún modelo; cada seguimiento dice si un mecanismo puede imponerlo; Hive tiene una tabla de sus reglas con el mecanismo que las impone; y el cierre de un trabajo lista los candidatos del proyecto y los cuenta entre sesiones en la memoria del proyecto.

**Prueba para pasar una regla a un mecanismo** (las cuatro): la respuesta sale de archivos o comandos sin juicio; no tiene excepciones que pidan criterio; ocurrió en dos sesiones o el error es irreversible; el mensaje de falla puede decir la corrección.

## Alcance y aceptación

Incluye: adaptadores de transcripción real y el modo `--audit-session` en `tests/pilot`; tres criterios nuevos para sesiones reales; el procedimiento del inspector en la rúbrica; el campo `Enforceable:` de los seguimientos en `AGENTS.md`; la tabla regla ↔ mecanismo de Hive; la parte «Candidatos» del cierre y su registro en memoria; la página del sitio y el delta de especificación.

Excluye: OpenCode, Pi, Cursor y Antigravity en el modo de auditoría (sus sesiones reales son pocas y no hay extractor probado); detectar automáticamente scripts improvisados «con el mismo propósito» entre sesiones distintas (lo anota la sesión que los corrió); ejecutar pilotos o sesiones de modelo; corregir los tres seguimientos que ya tienen recurrencia (son cambios propios).

- AC1 — `go run ./tests/pilot --audit-session <archivo> --host claude|codex|grok` lee una transcripción real del host (Claude: la del proyecto; Codex: el rollout; Grok: `updates.jsonl`), sin ejecutar ningún modelo, y escribe en stdout un JSON con el `cwd`, el modelo, el rango de tiempo y cada criterio con su estado (`pass`, `fail`, `review` o `not_observed`) y su evidencia por línea, sin texto de comandos ni de mensajes · false today because `tests/pilot/main.go` no tiene ese modo y ningún archivo de `tests/pilot` lee rollouts de Codex (`response_item`) ni el `updates.jsonl` de Grok · Observed in: el comando corrido sobre transcripciones reales de esta máquina en `~/.claude/projects`, `~/.codex/sessions` y `~/.grok/sessions`.
- AC2 — **Retired:** D11 A turns the close criterion into `review`; AC9 replaces it. Sobre las sesiones revisadas a mano el 2026-11-17, la auditoría marca `fail` en `text_round_without_card` para Codex carvento `01a1e864-77eb-7961-bd16-838efde72872` (ronda D2/D3 de 22:06Z) y fleetsystem `01a1eadf-1264-7b52-b775-9e6d6483ca6f` (D3 de 19:04Z), marca `fail` en `close_before_branch_cleanup` para Grok cam `01a1eb37-0f75-7882-b92a-d492997b9c6c` y carvento `01a1eb3d-a6b2-7b70-8503-90e3b2ef0ace` (el informe de cierre llegó antes de borrar la rama fusionada), y no marca `text_round_without_card` en Codex fleetsystem `01a1ea8f-7b42-7370-8c2a-9468640dcaf1`, cuyas rondas D1–D4 fueron por tarjeta · false today because no existe la auditoría · Observed in: las mismas transcripciones reales.
- AC9 — Sobre las sesiones revisadas a mano el 2026-11-17, la auditoría marca `fail` en `text_round_without_card` para Codex carvento `01a1e864-77eb-7961-bd16-838efde72872` (ronda D2/D3 de 22:06Z) y fleetsystem `01a1eadf-1264-7b52-b775-9e6d6483ca6f` (D3 de 19:04Z), marca `review` en `close_before_branch_cleanup` para Grok cam `01a1eb37-0f75-7882-b92a-d492997b9c6c` y carvento `01a1eb3d-a6b2-7b70-8503-90e3b2ef0ace` (el informe de cierre llegó antes de borrar la rama fusionada), y no marca `text_round_without_card` en Codex fleetsystem `01a1ea8f-7b42-7370-8c2a-9468640dcaf1`, cuyas rondas D1–D4 fueron por tarjeta. **Replaces:** AC2 · false today because `closeBeforeBranchCleanup` en `tests/pilot/session_audit.go` marca `fail` y solo reconoce `gh pr merge` · Observed in: las mismas transcripciones reales.
- AC3 — `AGENTS.md` «Measurement» pide a cada seguimiento una línea `Enforceable:` con el mecanismo candidato o `judgment`, y la rúbrica de revisión de sesiones empieza por la pasada automática y por leer los registros `candidate/` de los proyectos, promoviendo a Hive el que aparece en tres proyectos · false today because ninguno de los dos archivos lo dice · Observed in: los dos archivos en el candidato.
- AC4 — `_support/docs/harness-engineering/rule-enforcers.md` lista cada regla de Hive que un mecanismo impone, con su mecanismo, y cada regla que sigue en texto con `judgment` o con el mecanismo candidato; un test de contenido falla si una ruta de mecanismo de la tabla no existe · false today because el archivo no existe · Observed in: `go test ./tests/content/...` en el candidato, con un caso negativo que borra una ruta.
- AC5 — El informe final de `flow-close` tiene una parte «Candidates» que lista, solo cuando hay alguno, cada repetición que el trabajo no convirtió en mecanismo con el punto 5 de `defect-fix.md` (lo que se repitió, cuántas veces y el mecanismo propuesto), sin pregunta nueva · false today because `content/skills/flow-close/SKILL.md` no tiene esa parte · Observed in: `content/skills/flow-close/SKILL.md` del candidato; el comportamiento queda sin verificar hasta el primer cierre real que tenga un candidato, después del merge y la actualización local.
- AC6 — La página `flows/close` del sitio y la especificación `workflow-lifecycle` describen los candidatos del cierre · false today because no los mencionan · Observed in: `npm --prefix site run build` y `check_change.py` en el candidato.
- AC7 — El paso 7 de `flow-close` busca el candidato en la memoria del proyecto por los términos de su propósito y crea o actualiza `candidate/<herramienta>-<objetivo>` con la sesión, la fecha y el conteo · false today because el paso 7 solo escribe el resultado, las cuentas y los pendientes · Observed in: `content/skills/flow-close/SKILL.md` del candidato; sin verificar en una sesión real hasta después del merge.
- AC8 — Un candidato visto en dos sesiones se ofrece como una respuesta propia del pie de cierre («convertir el candidato <slug> en mecanismo»), no como un hallazgo · false today because el pie no tiene esa respuesta · Observed in: `content/skills/flow-close/SKILL.md` del candidato; sin verificar en una sesión real hasta después del merge.

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Las transcripciones reales de Claude Code (`~/.claude/projects/<proyecto>/<sesión>.jsonl`) traen líneas `assistant`/`user` con `message.content` de tipo `tool_use`, `tool_result`, `text` y `thinking`, la misma forma que lee la rama `claude` de `parseTrace` (`tests/pilot/trace.go:373`), más líneas de otros tipos (`attachment`, `api-request`, `mode`…) que no lee. Medido sobre la sesión `b103d2f3`, 2026-11-17.
- Los rollouts de Codex 0.169 (`~/.codex/sessions/AAAA/MM/DD/rollout-*.jsonl`) traen `response_item` con `message`, `function_call` (`request_user_input`, `spawn_agent`…) y `custom_tool_call` `exec`, cuyo `input` es JavaScript que llama `tools.exec_command({cmd: …})`; el primer `session_meta` trae `cwd` y `source` (subagente o no). Medido sobre `01a1e864` y `01a1ec36`, 2026-11-17.
- El `chat_history.jsonl` de Grok 1.0.89 no sirve como fuente: tras una compactación, los comandos anteriores solo quedan dentro del resumen (carvento `01a1eb3d`, líneas 2, 4 y 5 con `synthetic_reason: compaction_meta`) y no trae timestamps. El `updates.jsonl` de la misma carpeta sí los conserva: 761 eventos con `timestamp`, `session/update` `tool_call` (con `rawInput`) y `tool_call_update`, `agent_message_chunk`, `user_message_chunk` y `_x.ai/session/update` `turn_completed`. `summary.json` trae `current_model_id`. Medido sobre carvento `01a1eb3d`, 2026-11-17.
- `regressionCriteria` (`tests/pilot/regression.go:25`) agrupa seis criterios; `merged_branch_deleted` y `close_question_after_report` están en `regression.go:567` y `:413` y solo se llaman desde `flows.go:739–740`, con la base fija en `"main"`. `branchDeletionCommand` (`regression.go:668`) no reconoce `cleanup_branches.py`. `TestRegressionFixtures` parsea cada fixture con el formato de pilotos de su host.
- Los rollouts de fleetsystem `01a1eadf` escriben sus 76 `tools.exec_command` con `cmd:` y backticks, no comillas dobles.
- `AGENTS.md` «Measurement» prohíbe nuevos casos de regresión porque ninguna sesión real pasa por ellos con los pilotos pausados; con la auditoría de sesiones reales esa razón deja de valer para sus criterios.
- Tamaños: `content/skills/flow-close/SKILL.md` 13,318 B (límite 16,000); `content/guidance/global.md` 52,673 B, sin cambios previstos.

## Effects

None: the tasks take no effect on an external system.

## Entrega

`ready`, en uso en esta sesión y repositorio (D4 C de `pr-before-after-captures`): revisión, push, PR no borrador, esperar CI y revisiones, y parar con el PR abierto para que el usuario lo fusione. Revisión de código: Claude Code `/code-review` (D5 A), en uso; aplica porque el cambio toca un parser de formatos externos y guía distribuida. No autoriza fusionar, ni ejecutar modelos o pilotos.

## Estado de revisión y progreso

2026-11-17 — Plan escrito. Revisión `hive-review-plan`: 2 P1, 7 P2, 7 P3, todos aplicados. Grok se lee de `updates.jsonl` (el `chat_history.jsonl` pierde los comandos previos a una compactación); `text_round_without_card` ya no exige la línea `Situación:` y define turno y exenciones por host; la extracción de Codex acepta backticks e interpolación; los criterios existentes reciben `--base`, `cwd` y `cleanup_branches.py` como borrado solo en la auditoría; los fixtures nuevos van en `tests/fixtures/sessions/`; un candidato visto en dos sesiones se ofrece en el pie de cierre, no como hallazgo; AC5 se divide en AC5, AC7 y AC8; IDs completos y un caso negativo con rondas por tarjeta en AC2; T4 depende de T2.
