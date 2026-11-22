# audit-round-detection — el auditor de sesiones ve las rondas numeradas, el plan sin ruta y los paquetes leídos en celdas de Codex

| Campo | Valor |
| --- | --- |
| Estado | In build · verify |
| Tracker | ninguno · watch items `engram:tricell-hive/watch/codex-decision-format-async` (#17780) y `watch/plan-skips-delivery-round-and-handoff` (#17933) |
| Git | tricell-hive · base `development` en `1b87f0a7` · rama `feat/audit-round-detection` · worktree `.claude/worktrees/audit-round-detection` |
| Playbook | infra-ci |

## Decisiones

2026-11-21
- **D1 A** — ¿Qué hago con los hallazgos del análisis de la sesión Codex de facti `01a20155`? Registrar los watch items y planear el arreglo del auditor, que detecte la falla antes de tocar cualquier regla.
- **D2** — Se escribió `/flow-plan` con el alcance del arreglo: rondas numeradas en texto, plan sin ronda de entrega ni handoff, hojas leídas en `stage.md`, y el falso `try_it_block_in_report`.
- **D3 A** — ¿Dónde se trabaja? En un worktree nuevo desde `development`, porque hay otras sesiones activas en el proyecto.
- **Entrega A** — `automatic`.
- **Revisión A** — `Native: Claude Code /code-review`.
- **supuesto** — Un título numerado con pregunta («**1. ¿…?**») cuenta como título de ronda cuando le sigue al menos una opción con `(★)` antes del siguiente título. Descartado: exigir A y B, que dejaría fuera la pregunta de revisión de facti, que tiene una sola opción. También descartado: aceptar cualquier pregunta numerada, que marcaría como ronda la lista de preguntas de un análisis.
- **supuesto** — `A (★)` sin separador entre la letra y la estrella es la opción recomendada, tanto en `roundOptionStar` como en `optionLine`. Descartado: seguir exigiendo el separador, que es justo lo que dejó pasar la ronda de facti.
- **supuesto** — `reviewQuestionTitle` suma solo «revisamos el código» y «revisar el código». Descartado: un `revis\w+` suelto, que tomaría títulos ajenos.
- **supuesto** — Una confirmación imperativa («confirma», «confírmame», "confirm") con una elección de sesión cuenta como elección en prosa, solo al inicio de la frase o después de una coma o dos puntos. `sessionChoice` no cambia. Descartado: la palabra en cualquier posición, porque «el log confirma que…» no pregunta nada; y ampliar `sessionChoice` a «entregamos», que volvería `fail` preguntas que hoy son `review`.
- **supuesto** — Dos o más respuestas entre comillas después de «puedes responder» o "you can reply" dan `review`, no `fail`, porque puede ser una validación que la regla permite en texto cuando no hay tarjeta.
- **supuesto** — En una celda `exec` de Codex, cada comando recibe por posición una salida en texto plano, solo cuando la celda trae tantas salidas planas como comandos de shell. Solo la lectura de paquetes la usa. Descartado: dar a cada comando el texto de toda la celda, que juntaría una lectura filtrada con otra parcial y contaría como completa una lectura que no lo es. También descartado: `aggregated_output`, que puede traer salida que el modelo no vio.
- **supuesto** — `try_it_block_in_report` juzga el último mensaje del último turno que cargó `flow-close` después del cambio de UI, con un detector propio de esa carga. Descartado: sumar `flow-close` a `flowSkillRead`, que movería `flow_entered_without_route`.
- **supuesto** — `plan_without_build_route` aprueba solo con `flow-build` o `create-design`. Releer `flow-plan` no cierra la etapa, y la sesión que termina dentro de la etapa da `review`. Descartado: aprobar con cualquier comando de flujo, porque «voy a usar `/flow-plan`» aprobaría.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> a (D1 A: registrar los watch items y planear el arreglo del auditor)

> /flow-plan arreglo del auditor --audit-session: rondas numeradas en texto, plan sin ronda de entrega ni handoff, hojas leídas vía stage.md (sesión Codex facti 01a20155)

Sin ticket. Queda fuera cambiar las reglas de preguntas, entrega o handoff. El auditor va primero, como pide «Measurement» de `AGENTS.md` cuando el watch item nombra un mecanismo.

## Objective

Hoy `--audit-session` da `not_observed` en `text_round_without_card`, `decision_round_format` y `review_question_native_option` para la ronda de entrega escrita en texto en facti (línea 278). No ve el plan que terminó sin ronda de entrega ni bloque de siguiente paso. Dice que `report-endings` no se leyó, aunque llegó dentro de `hive-mode/stage.md`. Y `try_it_block_in_report` apunta a un mensaje de otro trabajo.

Con el cambio, el auditor marca esas fallas en la sesión real y los fixtures existentes conservan su veredicto.

## Scope and acceptance

**Dentro:** `tests/pilot/session_audit*.go`, sus fixtures en `tests/fixtures/sessions/`, el delta de `workflow-lifecycle`, y las filas del rubric y de `rule-enforcers.md` sobre estos criterios.

**Fuera:** el texto de las reglas distribuidas en `content/`, el sitio y los criterios del piloto.

**Fuente:** la sesión real es la copia congelada del rollout descrita en `tasks.md`. Las líneas citadas son de esa copia.

- AC1 — `text_round_without_card` falla un mensaje de Codex que escribe una ronda con títulos numerados («**1. ¿…?**») y opciones «- **A (★) · …**», cuando `request_user_input` estaba disponible y ninguna tarjeta lo sigue en su turno · false today because en la copia de facti da `not_observed` · Observed in: `go run ./tests/pilot --audit-session <copia de facti> --host codex`, con evidencia en la línea 278
- AC2 — `decision_round_format` falla esa misma ronda por no tener la línea `Situation:` · false today because da `not_observed` en la copia de facti · Observed in: la misma auditoría, con evidencia en la línea 278
- AC3 — `review_question_native_option` falla en Codex una pregunta de revisión escrita en texto («¿Cómo revisamos el código?») cuya única opción es `hive-review-code` · false today because da `not_observed` en la copia de facti · Observed in: la misma auditoría, con evidencia en la línea 278
- AC4 — `prose_choice_without_card` falla una confirmación imperativa de una elección de sesión («confirma `…, entrega local (hold)`») sin tarjeta después · false today because la línea 266 de la copia no aparece en su evidencia · Observed in: la misma auditoría, con la línea 266 en la evidencia de `fail`
- AC5 — El nuevo `plan_without_build_route` falla una etapa de plan (de la carga de `flow-plan` a la siguiente carga de otro flujo) en la que ningún mensaje del asistente trae `flow-build` o `create-design`, y pasa cuando uno lo trae · false today because el criterio no existe · Observed in: la misma auditoría, `fail` con evidencia en la carga de `flow-plan` de la línea 210
- AC6 — La lectura de un `stage.md` dentro de una celda `exec` de Codex cuenta como lectura de sus hojas cuando la salida que le toca a ese comando trae la línea final del paquete y todos sus encabezados · false today because `analysis_reply_block` dice «report-endings rule not read» en la línea 83 de la copia · Observed in: la misma auditoría, con «report-endings rule was read» en la evidencia de la línea 83
- AC7 — `try_it_block_in_report` juzga el último mensaje del último turno que cargó `flow-close` después del cambio de UI · false today because su evidencia apunta a la línea 2527, un mensaje posterior sobre otro trabajo · Observed in: la misma auditoría, con evidencia en la línea 1784
- AC8 — **Retired:** era verdadero en la base por construcción, así que no medía nada. Lo reemplaza la fila «Fixture verdicts» de la verificación en `tasks.md`.
- AC9 — `prose_choice_without_card` da `review` en un mensaje que ofrece dos o más respuestas entre comillas después de «puedes responder», sin tarjeta y sin otra elección que falle · false today because la línea 930 de la copia no aparece en ninguna evidencia del criterio · Observed in: el fixture `prose_choice_without_card/codex-rollout-quoted-replies-review.jsonl`, porque en la copia de facti un `fail` tapa el `review`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

Leído el 2026-11-21 en `01e43b50`; los archivos no cambiaron en `1b87f0a7`, salvo una línea de `session_audit_trackers.go` ajena a este cambio.

- `tests/pilot/session_audit.go:1610-1611`: `roundTitleLine` solo acepta `D<n>` o un encabezado fijo con pregunta. `roundOptionStar` exige un separador entre la letra y `(★)`.
- `tests/pilot/session_audit.go:2003`: `optionLine` tampoco acepta `A (★) ·`, así que `lettered()` y `reviewQuestionNativeOption` no ven las opciones de facti.
- `tests/pilot/session_audit.go:1993`: `reviewQuestionTitle` no reconoce «revisamos el código».
- `tests/pilot/session_audit.go:2029`: `reviewQuestionNativeOption` exige al menos dos opciones que nombren un mecanismo.
- `tests/pilot/session_audit_prose.go:52-70`: `proseQuestions` solo toma frases que terminan en «?» y salta las líneas que `roundTitleLine` reconoce.
- `tests/pilot/session_audit.go:497-565`: `codexExecOutput` une a cada comando solo los resultados JSON con `exit_code`. Confirmado con una prueba desechable sobre la copia de facti: la lectura de `hive-mode/stage.md` de la línea 29 queda con resultado vacío. La salida de esa celda (línea 39) trae dos salidas planas para dos comandos de shell, y la segunda es el paquete completo con su línea final.
- `tests/pilot/session_audit_trackers.go:120-151`: `tryItBlockInReport` juzga el último mensaje del asistente de toda la sesión. En la copia, `flow-close` se carga en la línea 1639 con un rango de `sed`, y `flowSkillRead` no lo reconoce.
- En la copia, la línea 210 es la lectura de `flow-plan/SKILL.md`. La línea 218 solo cuenta líneas de `stage.md`.
- `_support/openspec/specs/workflow-lifecycle/spec.md:1446`: el requisito de `analysis_reply_block` ya pide contar una hoja leída en un paquete de etapa. Su excepción para las rondas en texto coincide con el cambio.
- `_support/docs/harness-engineering/session-review-rubric.md:39` ya tiene la fila «Moving between flows ends with a next-step block». La línea 42 y `rule-enforcers.md:46` describen `prose_choice_without_card`.

Comandos medidos el 2026-11-21:
- `go test ./tests/pilot/` · ok, 62 s.
- `go test ./tests/pilot/ -run 'TestAuditNewCriteria|TestAnalysisReplyBlockEvidenceSaysIfTheRuleWasRead|TestLeafReadBeforeActionEdgeCases|TestAuditCodexAdapter'` · ok, 0.4 s.
- `go test ./tests/content/` · ok, 2 s.
- `go vet ./tests/pilot/` · ok, 0.3 s.
- `<barrido>` con el binario de la base en `1b87f0a7` · 329 fixtures en 5.7 s. Los dos `main_thread_cost/{claude,codex}.jsonl` necesitan el `sed` de `tasks.md` para sacar el host; con él, los 329 se leen.
- `go run ./tests/pilot --audit-session <copia de facti> --host codex` · corre. Da `not_observed` en `text_round_without_card`, `decision_round_format` y `review_question_native_option`; `analysis_reply_block` falla con «not read»; `try_it_block_in_report` falla en la línea 2527.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` (respuesta de Entrega, 2026-11-21): reviso, hago push, abro el PR hacia `development` y lo fusiono sin parar. No hay instalación local que refrescar, porque `tests/` no forma parte del binario ni del contenido desplegado. La revisión es `Native: Claude Code /code-review`, sobre el candidato antes del push. Esto no autoriza cambiar reglas de `content/`, publicar un release ni tocar el rollout de facti.

## Review status and progress

2026-11-21 — Plan escrito. Lo revisó un `hive-review-plan` en 341 s y 175 k tokens: 4 P1, 5 P2 y 6 P3, todos aplicados.
- AC3 suma «revisamos» al título y su forma de una opción.
- El título numerado ya no exige A y B: le basta una opción con estrella.
- `optionLine` acepta `A (★)`.
- AC5 cita la línea 210.
- AC8 se retira y queda como el barrido de veredictos.
- La celda `exec` se empareja por posición.
- La etapa del plan solo la cierra otro flujo y solo `flow-build` o `create-design` la aprueban.
- `flow-close` tiene un detector propio.
- `proseQuestions` usa la función nueva de títulos.
- No se amplía `sessionChoice`.
- AC4 se separa en AC4 y AC9.
- La fila del rubric se reutiliza.
- Se mide `go vet`.
- Se registra el checksum de la copia.
- El delta de la spec se ordena.
2026-11-22 — Build T1–T5 (`3a4454c4`, `742ed7b2`, `2f2de663`, `995e234a`, `6818b23f`). Comparar 65 sesiones reales entre la base y el candidato mostró un falso positivo con «Confirmé»; se corrigió en `4137c556`. La primera `/code-review` dejó 7 hallazgos bloqueantes, cada uno reproducido con un fixture; se corrigieron en `35a94a21`. La segunda, limitada a esos arreglos, dejó otros 7, que se corrigieron en `d2916d67` con una línea `Unreviewed:`. Quedan para los hallazgos del cierre los no bloqueantes: «responde» en indicativo con comillas dobles, ítems no textuales de una celda `exec` contados como salidas, `readsFlowSkill` y `loadsFlowClose` que toman un grep como carga (previo) y `flowCommand` sin `*_` como `buildRoute`. La relectura de `flow-plan` en otro turno quedó cubierta por el arreglo de `d2916d67`, que compara con la última carga.
