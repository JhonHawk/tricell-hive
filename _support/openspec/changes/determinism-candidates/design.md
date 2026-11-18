# Diseño

## Enfoques considerados (D10)

| Enfoque | A favor | En contra |
| --- | --- | --- |
| **Extender `tests/pilot` (elegido)** | Un solo dueño por criterio; reutiliza `traceReport`, `traceEvent` y `criterionAssessment` | Go; el adaptador de Codex extrae comandos de JavaScript |
| Herramienta Python en `_support/scripts` | Iteración rápida; los extractores de `measure/` ya existen | Criterios duplicados entre pilotos y sesiones reales, que se desalinean |
| Puente Python → Go | No toca el parser | Dos piezas; `--assess` espera un directorio de corrida con metadatos que una sesión real no trae |

## Auditoría de sesiones reales

**Entrada.** `--audit-session <archivo> --host claude|codex|grok [--base <rama>]`. La auditoría no ejecuta ningún modelo: lee un archivo que ya existe.

- **Claude:** la transcripción del proyecto (`~/.claude/projects/<p>/<sesión>.jsonl`), por la rama `claude` existente, que ignora los tipos de línea que no conoce. El `cwd` sale del campo `cwd` de sus líneas.
- **Codex:** el rollout. El `cwd` y si es un subagente salen del primer `session_meta`; el modelo, del primer `turn_context`, reportado como configurado porque Codex no expone el servido.
- **Grok:** el `updates.jsonl` de la carpeta de sesión, no `chat_history.jsonl`, porque este pierde los comandos anteriores a una compactación (solo quedan dentro del resumen) y no tiene timestamps. `updates.jsonl` es el flujo de eventos del protocolo del cliente (`session/update` con `tool_call` y `rawInput`, `tool_call_update` con su estado, `agent_message_chunk`, `user_message_chunk`, y `_x.ai/session/update` `turn_completed`), con `timestamp` en cada línea. El `cwd` sale del nombre decodificado de la carpeta del proyecto; el modelo, de `summary.json`.
- **Base:** `--base`, o el `Base branch` del `## Hive` del `AGENTS.md` en el `cwd` cuando existe; sin ninguno, los criterios que la necesitan devuelven `not_observed` con esa razón.

**Normalización.** Cada adaptador produce el `traceReport` existente, con los tipos de evento que el parser ya usa (`shell`, `read`, `write`, `question`, `skill_invocation`, `tool_result` y `text` con su `Role`), más un límite de turno por evento:

- Turno: Codex `task_started` … `task_complete` o `turn_aborted`; Claude, una línea `user` que no es `tool_result` ni meta; Grok, `user_message_chunk` hasta `turn_completed`.
- Codex `exec`: de cada `tools.exec_command({cmd: …})` se extrae el literal entre `"`, `'` o backticks; uno con interpolación `${…}` queda como comando retenido sin texto. `tools.apply_patch` es `write`; `tools.mcp__*` es una herramienta, con la marca de memoria para Engram; `tools.write_stdin` se ignora. El éxito de un comando dentro de un bloque solo se da por cierto cuando la salida del bloque no trae error; si no, queda sin éxito conocido. `function_call` `request_user_input` y `request_user_input_async` son `question`.
- Grok: `tool_call` con `rawInput.command` es `shell`; la herramienta de preguntas (`ask_user_question`, como la lee `measure/others/grok_rounds.py`) es `question`; los `agent_message_chunk` consecutivos se unen en un `text` de rol asistente.

**Salida.** JSON en stdout: host, archivo, `cwd`, modelo (con `configured` cuando no se ve el servido), primer y último timestamp (Grok: de `updates.jsonl`), y por criterio `status` (`pass`, `fail`, `review`, `not_observed`) y `evidence` (línea, tipo de evento, herramienta). Nunca el texto de un comando o de un mensaje, igual que `regressionEvidence`, porque pueden traer secretos.

**Criterios existentes en la auditoría.** Corren los seis de `regressionCriteria` y `merged_branch_deleted`, con estos ajustes solo en el modo de auditoría:

- `merged_branch_deleted` usa la base de arriba y cuenta `cleanup_branches.py … --apply` como borrado de rama.
- `no_broad_git_add` resuelve rutas contra el `cwd` de la sesión; sin inventario de archivos, una ruta de directorio que no puede comprobar queda en `review`, no en `fail`.
- El README de los fixtures dice qué criterios pierden precisión sin inventario.

**Criterios nuevos (solo en la auditoría).**

1. `text_round_without_card` — Una ronda es un mensaje del asistente con una línea de título que cumple `^\s*[#>*\s]*D\d+\s*[—–-]` y al menos una línea de opción con `(★)`. `fail` cuando, después de esa ronda y antes de que termine el turno, no hay llamada a una herramienta nativa de preguntas, y la sesión llamó a una en otro momento (prueba de que estaba permitida). No cuenta (`not_observed` para esa ronda):
   - cuando el resultado de la última pregunta del turno antes de la ronda está vacío o expiró: la guía global pide entonces la ronda en texto;
   - en un rollout de subagente (`session_meta.source.subagent`);
   - en una sesión sin ninguna llamada a herramienta de preguntas.
2. `close_before_branch_cleanup` — Tras cada merge exitoso (`gh pr merge` sin `--delete-branch`, o `git merge` seguido del push de la base): `review`, nunca `fail` (D11 A), cuando un mensaje del asistente cuya primera línea en negrita empieza por `Cierre` o `Close` llega antes del primer comando que borra una rama, porque en el punto de completitud `flow-close` prescribe ese orden cuando queda una rama fusionada por borrar (corre los pasos 1, 3, 6 y 7 y deja 2, 4 y 5 pendientes), y solo los cierres que no difieren nada (el de seguimiento, el abandonado, el de limpieza de un despliegue, la higiene, el fin de sesión y el no supervisado) deben borrar antes del informe. Cuenta como borrado `git branch -d/-D`, `git push … --delete` o `cleanup_branches.py … --apply`. `not_observed` cuando no hay ningún borrado después; ese caso lo cubre `merged_branch_deleted`.
3. `skill_named_as_reason` — `review`, nunca `fail`: una oración del asistente que nombra una skill de Hive (`flow-*`, `git-workflow`, `harness-audit`, `docs-cleanup`, `create-issue`, `create-prd`) junto a un verbo de exigencia (`pide`, `exige`, `establece`, `requires`, `asks`, `sets`). No corre cuando el `cwd` es el repositorio de Hive, donde nombrar una skill es legítimo porque el trabajo la cambia.

**Fixtures.** Van en `tests/fixtures/sessions/<criterio-o-adaptador>/<host>-<formato>-<pass|fail>.jsonl` (`claude-project`, `codex-rollout`, `grok-updates`), con su propio test, porque `--host codex` y `--host grok` en `tests/fixtures/regression/` ya significan los formatos de los pilotos (`codex exec --json`, `grok -p`). Se sanean como fija `tests/fixtures/regression/README.md`. No son casos de regresión de una corrección de la guía: son las pruebas de una herramienta que sí recibe sesiones reales.

## Inspector (nivel Hive)

La rúbrica (`session-review-rubric.md`) gana una sección inicial:

1. Correr la auditoría sobre cada sesión a revisar.
2. Leer los `candidate/` de los proyectos de Engram, con una búsqueda por tema en cada proyecto.
3. Leer las transcripciones solo donde un criterio marcó `fail` o `review`. Un `review` de `close_before_branch_cleanup` es falla solo cuando ese cierre era de los que no difieren nada, o cuando su pie no ofrece la limpieza pendiente como respuesta.

Un candidato que aparece en tres proyectos se vuelve candidato de Hive. Cada seguimiento gana una línea `Enforceable: <mecanismo> | judgment`, decidida con la prueba de cuatro condiciones de la propuesta. Además se corrige la fila «An empty or timed-out card gets one line, not the round repeated» (2026-10-31), que contradice la regla vigente: una tarjeta que se cerró sola recibe la ronda escrita en texto.

## Candidatos del proyecto (nivel proyecto)

- **Qué cuenta:** una repetición que el trabajo no convirtió en mecanismo con el punto 5 de `defect-fix.md`, que es el hogar canónico de la escalera de capas. Es decir, un script en línea o una cadena de comandos que el trabajo corrió dos o más veces para el mismo fin, o una corrección que el usuario dio dos veces en el trabajo. Lo anota la sesión que lo hizo, porque sabe para qué servía.
- **Informe:** una parte «Candidates» del informe final de `flow-close`, una línea por candidato con lo que se repitió, cuántas veces y el mecanismo propuesto, sin pregunta nueva. El mecanismo se elige con el punto 5 de `defect-fix.md`; el cierre solo apunta ahí, sin repetir la escalera.
- **Memoria:** el paso 7 busca el candidato en la memoria del proyecto por los términos de su propósito antes de crear un registro, y lo crea o actualiza con la sesión, la fecha y el conteo. El tema es `candidate/<herramienta>-<objetivo>`, en kebab-case, por ejemplo `candidate/link-check-docs`.
- **Propuesta:** un candidato visto en dos sesiones se ofrece como una respuesta propia del pie de cierre, «convertir el candidato <slug> en mecanismo», que entra en `flow-build` como un arreglo. No es un hallazgo, porque «Findings» de `flow-close` define un hallazgo como un defecto.

## Tabla regla ↔ mecanismo de Hive

`_support/docs/harness-engineering/rule-enforcers.md` tiene tres columnas:

- la regla, con un enlace a su hogar canónico;
- el mecanismo (la ruta del script, el test o el criterio de auditoría), o `judgment`;
- el estado: `enforced`, `candidate` o `judgment`.

Se siembra con los siete scripts de la Fase 1, los tests de contenido que imponen reglas (tamaño de skills, referencias, piso de modelos) y los criterios de la auditoría. Un test de contenido comprueba que cada ruta de la columna de mecanismo exista.
