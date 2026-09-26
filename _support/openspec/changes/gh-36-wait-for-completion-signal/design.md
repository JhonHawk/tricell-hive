# Diseño

## Contexto verificado

**Codex, código fuente en `rust-v0.157.1` (`36650394`), la versión instalada (`codex-cli 0.157.1`):**
- `wait_agent` (multi-agent v2) recibe un solo parámetro, `timeout_ms`, y espera actividad de *cualquier* subagente vivo. Los límites son: por defecto 30 s, mínimo 10 s (un valor menor se sube al mínimo) y máximo 3 600 000 ms (un valor mayor da error). El prompt del host ya pide "prefer longer waits (minutes) to avoid busy polling" (`codex-rs/prompts/src/multi_agent_instructions.rs:10`, verificado).
- Shell: `UnifiedExec` está activo por defecto. `exec_command` devuelve el control tras `yield_time_ms` (250–30 000 ms) sin matar el proceso y entrega un `session_id`. `write_stdin` con entrada vacía espera hasta 300 000 ms. El patrón correcto para CI es un solo `gh run watch` y esperas reutilizadas sobre esa misma sesión.
- `codex exec --json` emite un item `collab_tool_call` con estos campos: `tool` (`spawn_agent`, `send_input`, `wait`, `close_agent`), `sender_thread_id`, `receiver_thread_ids`, `prompt`, `agents_states` (mapa hilo → `{status, message}`) y `status` (`in_progress`, `completed`, `failed`). No trae el `timeout_ms` ni una marca de "agotado". `ResumeAgent` también se emite como `wait`, mientras que `list_agents`, `send_message`, `followup_task` e `interrupt_agent` no se emiten (`codex-rs/exec/src/exec_events.rs:122-124,210-257`; `event_processor_with_jsonl_output.rs:236-257`).
- Bug abierto [openai/codex#24951](https://github.com/openai/codex/issues/24951): un `wait_agent(300000)` tardó 7.5 h en volver. Esperar en tramos acota ese riesgo.

**Transcripts** (rollouts interactivos en `~/.codex/sessions/2026/09/`; solo conteos y parámetros, sin contenido):
- ark `01a0b087`: 56 `wait_agent` (60 000 ms ×45, 10 000 ms ×11); se agotó el 66 %.
- tricell-hive `01a0c10d`: 166 `wait_agent`, se agotó el 54 %, cadenas de hasta 7 seguidas.
- sample-project `01a0db24`: `gh run view` ×87 y `gh run list` ×13, contra `gh run watch` ×4.
- ark `01a0dab3`: `gh run view` ×145 contra `gh run watch` ×1, con intervalo mediano de 55.7 s.
- Ninguna sesión pidió un `timeout_ms` mayor a 60 000.

**Otros hosts** (documentación consultada con ctx7 sobre las versiones instaladas; la revisé en la síntesis, no la ejecuté):
- **Espera bloqueante propia:**
  - Claude Code 2.1.283: `Monitor`, además de comandos en segundo plano con aviso al terminar.
  - Pi 0.87.1: `job wait` con `budgetMs`; los subagentes que exceden el límite se encolan.
- **Segundo plano con aviso al terminar:**
  - Grok Build 1.0.38. El esquema de su herramienta dice: "you are notified on completion, so do not poll or sleep-wait for it".
  - OpenCode v2.0.18.
  - Cursor Agent 2026.09.23.

**Precedente:** optional external research source resuelve lo mismo en Codex: no esperar si queda trabajo local; si no, tramos de 5 a 10 min con una línea de estado tras cada uno.

**Guía actual:** [global.md:82](../../../../content/guidance/global.md) dice solo "Continue useful independent work while children run, or wait when none remains". Ninguna regla dice cómo esperar. `global.md` pesa 38 352 bytes, exactamente el valor de `globalGuidanceBudget` ([budget_test.go:12](../../../../tests/content/budget_test.go)).

## Regla nueva en `global.md`

Reemplaza la cola de `:82`, desde "Continue useful independent work while children run, or wait when none remains." hasta el final del punto. Esa cola incluye también las frases sobre resultados descartados o entregados en un turno posterior y la del hijo delegado. El texto nuevo es:

> Continue useful independent work while children or other external processes, such as CI runs or deployments, are running. When none remains, wait on the completion signal instead of polling. When the host delivers that signal to a later turn, such as a child's result or a background command's completion notification, the main thread ends its turn with the in-progress item described under Communication; when the host drops results that arrive after you return, collect them before returning. Otherwise keep one blocking wait, such as a subagent wait call or a watch command like `gh run watch`, and in stretches of about five to ten minutes, or the host's maximum when lower, re-await the same running process or rerun the same watch, never starting new short sleep-and-check cycles. After each stretch that ends without the result, the main thread gives the user one line naming what is still running and what its result decides, in the same response as the next wait so the turn stays open. A delegated child still returns its final result, not interim status.

**Qué conserva:** las tres frases que reemplaza, reescritas: resultados que el host descarta, resultados entregados en un turno posterior (ahora también para procesos, no solo para subagentes) y el hijo que devuelve su resultado final. Con notificación, terminar el turno pasa de "may" a indicativo, porque "may" frente a "wait" dejaba que un modelo débil se quedara esperando (revisión del plan, B1). La línea va en la misma respuesta que la siguiente espera porque, en Codex, un mensaje sin llamada a herramienta cierra el turno (B2). Codex no despierta al hilo principal cuando llega el aviso de un hijo con el turno ya cerrado (`codex-tools.md:57-60` de optional reference project), así que ahí aplica la rama de espera activa.

**Por qué ahí:** la parte de delegación que toca al hilo principal es global, porque los hijos genéricos no reciben otra cosa (hallazgo de gh-38). Las esperas de CI ocurren también fuera de `flow-build`, y la entrega por skills llega solo en ~45–70 % de las sesiones medidas. Se descartó ponerla en `flow-build` por ese motivo.

**Redacción:** se nombra la capacidad (notificación de fin, espera bloqueante, comando de seguimiento). `gh run watch` es solo un ejemplo que marca el límite; no se nombra ninguna herramienta de un host. El tramo cubre los tres tipos de host: con notificación no hay tramo, y con espera propia el tramo es la espera larga. Así el aviso al usuario tiene un momento concreto en vez de "después de unos minutos".

**Presupuesto:** `globalGuidanceBudget` sube al tamaño medido tras la edición, como en `backlog-report-scope`.

## Parser de trazas de Codex

En la rama `codex` de `parseTrace` ([trace.go:496-524](../../../../tests/pilot/trace.go)), un item `item.completed` con `type: "collab_tool_call"` produce:
- `addTool(line, "collab_" + tool, id, item, "")`. El `Input` conserva el item, de donde el criterio lee `agents_states`.
- `finishTool(line, id, "", status == "failed")`.

`toolKind` deja estos nombres como `tool`. `item.started` se ignora a propósito, igual que en las demás ramas: los ids de inicio y fin salen del mismo mapeo. Los criterios que usan `tool_result` como límite solo actúan sobre eventos `question`, que Codex no emite, así que ningún fixture ni corrida existente cambia de estado (revisión del plan, N2).

## Criterio `no_poll_wait_chain`

Función `noPollWaitChain(r result) criterionAssessment` en `tests/pilot/regression.go`. Queda **declarada pero no conectada** a `regressionCriteria` ni a `assessFlows`, igual que `flowSkillReadBeforeDelivery`, y `TestRegressionCriteriaReturnsAllSix` no cambia. Hay tres motivos: los repos de fixture de los casos flows no tienen CI de GitHub; un `fail` cambiaría el estado del caso con falsos positivos sin medir; y `close_question_after_report` y `merged_branch_deleted` siguen el mismo modelo (revisión del plan, N3).

Se recorren los eventos en orden, **ignorando los `tool_result`**. Hay dos patrones:

1. **Ciclo de sondeo por shell (cualquier host).**
   - **Consulta de estado:** un segmento de `shellSegments` que, tras saltar asignaciones iniciales `VAR=valor` y el flag `-R <repo>`/`--repo <repo>`, es `gh run view`, `gh run list`, `gh pr checks` o `gh pr view` con `statusCheckRollup` en sus argumentos. No cuenta como consulta si lleva `--log` o `--log-failed` (es diagnóstico), ni `--watch` (es una espera bloqueante). `gh run watch` y `gh pr checks --watch` se reconocen como espera saltando los mismos prefijos. Cualquier evento de herramienta que no sea un ciclo corta la cadena de shell, incluidos los que fallan.
   - **Ciclo:** hay dos formas.
     - Un comando de shell con un segmento `sleep` y un segmento de consulta.
     - Un comando formado solo por segmentos `sleep`, cuyo siguiente evento de shell es una consulta.
   - **Cadena:** dos ciclos sin ningún otro evento de herramienta entre ellos. El texto del asistente no la corta: avisar mientras se sigue sondeando sigue siendo sondeo. **Falla** en el segundo ciclo de una cadena.
2. **Cadena de esperas de subagente (Codex).** Dos eventos `collab_wait` con `Success` verdadero, sin texto del asistente ni otro evento de herramienta entre ellos. Una espera fallida (`status: failed`) no cuenta como espera ni corta la cadena. **Falla** en el segundo, salvo que la primera espera haya traído un resultado: algún agente aparece en `agents_states` en estado terminal (`completed`, `errored`, `shutdown`, `interrupted`, `not_found`) y no lo estaba en el evento `collab_*` anterior. La comparación con el anterior evita contar dos veces un agente terminado que Codex siga listando. No depende de qué agentes incluya el handler de `wait`, que no se verificó (revisión del plan, ronda 2, N1).

**Resultado:**
- **Pasa** cuando hay algún evento de espera (un `gh run watch`, un ciclo aislado o un `collab_wait`) y ninguna cadena.
- Queda **no observado** cuando no hay ninguno.
- La evidencia sigue a `regressionEvidence` (línea y herramienta), nunca el comando ni el texto.

**Límites declarados:**
- **Duración de los tramos:** sin marcas de tiempo, el criterio no distingue un tramo corto de uno largo.
- **Bucles dentro de un comando:** un bucle como `until … do sleep 30; done` no se juzga, y a propósito. Es una sola espera bloqueante, sin vueltas del modelo, así que la regla no lo prohíbe.
- **`write_stdin`:** las esperas de `write_stdin` de `UnifiedExec` no se observan como comando. Sondeos repetidos sobre la misma sesión no cuentan como ciclo (es el patrón correcto).
- **Mensajes intermedios de Codex:** que `codex exec` emita items `agent_message` entre llamadas `collab` de un mismo turno se toma del esquema, no se observó en una corrida real.
- **Otros hosts:** Claude, Grok, Pi, OpenCode y Cursor no emiten un evento de espera de subagente en el modelo de trazas; ahí solo aplica el patrón de shell.
- **Otros CI:** la lista de consultas de estado cubre GitHub Actions. `kubectl rollout status` o un `curl` a un endpoint de estado no se juzgan.
- **Formato de traza:** `parseTrace` no lee los rollouts interactivos de `~/.codex/sessions`. El criterio no puede reevaluar las sesiones de origen; solo sirve sobre `codex exec --json` y trazas de otros hosts (N4).

## Fixtures

Carpeta `tests/fixtures/regression/no_poll_wait_chain/`, en el formato de `codex exec --json` que lee `parseTrace`:
| Fixture | Contenido | Esperado | Rama que prueba |
| --- | --- | --- | --- |
| `codex-ci-fail.jsonl` | tres `sleep 60 && gh run view <id> --json status` seguidos (forma de sample-project `01a0db24`, ark `01a0dab3`) | fail | ciclo en un solo comando |
| `codex-ci-bare-sleep-fail.jsonl` | `sleep 60` → `gh run view <id>`, dos veces | fail | ciclo con `sleep` suelto |
| `codex-ci-pass.jsonl` | un `gh run watch <id> --exit-status` | pass | espera correcta |
| `codex-ci-single-pass.jsonl` | un ciclo, trabajo local (`go test`), otro ciclo | pass | umbral de cadena y corte por otra herramienta |
| `codex-ci-log-failed-pass.jsonl` | un ciclo seguido de inmediato por `sleep 10 && gh run view <id> --log-failed` | pass | exclusión de `--log-failed` |
| `codex-wait-fail.jsonl` | tres `collab_tool_call` `wait` seguidos, con agentes en `running` y sin `agent_message` entre ellos (forma de ark `01a0b087`) | fail | cadena de esperas |
| `codex-wait-pass.jsonl` | `wait` con agentes en `running` → `agent_message` de una línea → `wait` | pass | corte por texto |
| `codex-wait-failed-pass.jsonl` | `wait` con `status: failed` → `wait` | pass | la espera fallida no cuenta |
| `codex-wait-result-pass.jsonl` | `wait` que devuelve A en `completed` (antes `running`) y B en `running` → `wait` sin texto | pass | excepción por resultado nuevo |

Un test unitario cubre el caso **no observado** (una traza sin esperas).

**Supuesto de formato:** las sesiones de origen son rollouts interactivos, no `codex exec --json`. Los fixtures reconstruyen la forma con el esquema de `exec_events.rs` en `rust-v0.157.1`, y `provenance.md` lo declara.
