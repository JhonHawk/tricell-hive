# Tasks

Base `development` en `1b87f0a7`, rama `feat/audit-round-detection`, worktree `.claude/worktrees/audit-round-detection`. Todas las tareas corren en el hilo principal: tocan el mismo paquete Go y comparten patrones (`roundTitleLine`, `roundOptionStar`, `optionLine`), así que en paralelo chocarían, y cada una es menor que su brief.

`<copia de facti>` es `/private/tmp/claude-501/-Users-jmartinez-Development-projects-tricell-tricell-hive/fe7ea3a1-1f06-4409-be29-f99dfb5d78ac/scratchpad/facti-01a20155.jsonl`. Es una copia local del rollout de Codex `~/.codex/sessions/2026/11/21/rollout-2026-11-21T21-38-50-01a20155-ce5a-7030-a04e-ebef68ef4ca0.jsonl`, 2530 líneas, sha256 `a5e67ee8b84606ed…`, y no se versiona. Si un reinicio la borra, se vuelve a copiar del rollout y se toman sus primeras 2530 líneas. Los fixtures nuevos usan texto sintético: ninguno lleva claves de ticket, correos ni nombres de entornos de ese proyecto.

`<barrido>` corre `--audit-session` sobre cada fixture y escribe una línea por archivo con el estado de cada criterio: `for f in tests/fixtures/sessions/*/*.jsonl; do printf '%s\t' "$f"; <binario> --audit-session "$f" --host "$(basename "$f" | sed 's/[-.].*//')" | python3 -I -c 'import json,sys; d=json.load(sys.stdin); print(" ".join(c["criterion"]+"="+c["status"] for c in d["criteria"]))'; done`, desde la raíz del worktree. El binario de la base se compila antes de la primera edición: `go -C <worktree> build -o <tmp>/pilot-base ./tests/pilot`.

### T1 — Rondas numeradas en texto

- [x] Un título numerado con pregunta cuenta como título de ronda cuando le sigue al menos una opción con estrella antes del siguiente título; `roundOptionStar` y `optionLine` aceptan `(★)` entre la letra y el separador; `reviewQuestionTitle` reconoce «revisamos el código» y «revisar el código»; `reviewQuestionNativeOption` juzga una pregunta de revisión con una sola opción que nombra un mecanismo

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `tests/pilot/session_audit.go`, `tests/pilot/session_audit_routes.go`, `tests/pilot/session_audit_prose.go` (`proseQuestions`), `tests/fixtures/sessions/text_round_without_card/`, `tests/fixtures/sessions/decision_round_format/`, `tests/fixtures/sessions/review_question_native_option/`, `tests/fixtures/sessions/prose_choice_without_card/`, `tests/pilot/session_audit_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:**
- `roundTitleLine` no cambia, porque RE2 no puede exigir «seguido de una opción con estrella». Una función a nivel de mensaje encuentra los títulos numerados con pregunta (`^\s*[#>*\s]*\d+[.)]\s*[^\n]*\?`) y los acepta solo si una línea de opción con `(★)` les sigue antes del siguiente título. `decisionBlocks`, `isTextRound` y `proseQuestions` usan esa función.
- `proseQuestions` deja de saltar un título numerado que no tiene opción con estrella, así que una elección numerada en prosa sin letras sigue contando como pregunta.
- `reviewQuestionTitle` suma solo `revis(?:amos|ar) el c[oó]digo`. Un `revis\w+` suelto queda descartado porque tomaría títulos ajenos.

Fixtures:
- `text_round_without_card`, `decision_round_format` y `review_question_native_option` reciben cada uno una ronda con la forma de la línea 278 de facti, con texto sintético (`fail`).
- Un análisis con una lista numerada de preguntas sin opciones con estrella (`not_observed` en los tres).
- Una elección numerada en prosa sin letras, «**1. ¿interactive o hold?**», que sigue fallando en `prose_choice_without_card`.

**Verification:** `go test ./tests/pilot/ -run 'TestAuditNewCriteria'` pasa con los fixtures nuevos. `go run ./tests/pilot --audit-session <copia de facti> --host codex` da `fail` en la línea 278 para `text_round_without_card`, `decision_round_format` y `review_question_native_option`.

### T2 — Elecciones en prosa: confirmación imperativa y respuestas entre comillas

- [x] `prose_choice_without_card` falla una confirmación imperativa de una elección de sesión y marca `review` en un mensaje que ofrece dos o más respuestas entre comillas

**Closes:** AC4, AC9

**Depends on:** none

**Locations:** `tests/pilot/session_audit_prose.go`, `tests/fixtures/sessions/prose_choice_without_card/`, `tests/pilot/session_audit_test.go`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- Una frase sin «?» cuenta como pregunta cuando abre con una confirmación imperativa («confirma», «confírmame», "confirm"), o la trae después de una coma o dos puntos, y nombra una elección de sesión. `sessionChoice` no cambia.
- Un mensaje que ofrece dos o más respuestas entre «» o "" después de «puedes responder», «responde» o "you can reply" da `review`.
- El rubric (fila de `prose_choice_without_card`) y `rule-enforcers.md` (fila 46) nombran ambos casos.

Fixtures:
- La confirmación imperativa (`fail`).
- Las respuestas entre comillas (`review`).
- «el log confirma que la entrega…» (`not_observed`).

**Verification:** `go test ./tests/pilot/ -run 'TestAuditNewCriteria'` pasa. La auditoría de la copia de facti incluye la línea 266 en la evidencia de `prose_choice_without_card`.

### T3 — Paquetes leídos dentro de una celda `exec` de Codex

- [x] Cuando una celda `exec` trae tantas salidas en texto plano como comandos de shell, cada comando recibe la suya por posición, y `packetNamesLeaf` la usa si el resultado del comando está vacío; con otro número, ningún comando recibe texto

**Closes:** AC6

**Depends on:** none

**Locations:** `tests/pilot/session_audit.go` (`codexExecOutput`, `traceEvent`), `tests/pilot/session_audit_packets.go`, `tests/fixtures/sessions/analysis_reply_block/`, `tests/fixtures/sessions/leaf_read_before_action/`, `tests/pilot/session_audit_packets_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** Un campo nuevo del evento guarda esa salida, y solo la lectura de paquetes lo lee. La regla del paquete completo no cambia: exige la línea final y cada encabezado, y la invalidan una marca de truncado o un filtro (`packetFilter`).

Fixtures:
- Una celda con el paquete completo (lectura contada).
- Una celda truncada (no contada).
- Una celda con un `grep` de los encabezados y una lectura parcial del mismo paquete (no contada).
- Una celda con más salidas planas que comandos (no contada).

**Verification:** `go test ./tests/pilot/ -run 'TestAuditNewCriteria|TestAnalysisReplyBlockEvidenceSaysIfTheRuleWasRead|TestLeafReadBeforeActionEdgeCases|Packet'` pasa. La auditoría de la copia de facti dice «report-endings rule was read» en la evidencia de la línea 83.

### T4 — El reporte que juzga Try it

- [x] `tryItBlockInReport` juzga el último mensaje del asistente en el último turno que cargó `flow-close` después del cambio de UI

**Closes:** AC7

**Depends on:** none

**Locations:** `tests/pilot/session_audit_trackers.go`, `tests/fixtures/sessions/try_it_block_in_report/`

**Execution:** main thread

**Test approach:** check

**Changes:** Un detector propio de la carga de `flow-close`: la llamada a la skill, un Read de su `SKILL.md` instalado, o un comando de shell con `readCommand` que lo nombra, rangos de `sed` incluidos. `flowSkillRead` no cambia, así que `flow_entered_without_route` no se mueve. Sin `flow-close`, el comportamiento actual se mantiene.

Fixtures:
- Un reporte de cierre sin Try it, cargado con un rango de `sed`, seguido de otro trabajo con un último mensaje distinto. La evidencia debe apuntar al reporte.

**Verification:** `go test ./tests/pilot/ -run 'TestAuditNewCriteria'` pasa. La auditoría de la copia de facti da la evidencia de `try_it_block_in_report` en la línea 1784.

### T5 — Nuevo criterio `plan_without_build_route`

- [x] El criterio falla una etapa de plan sin comando de build en ningún mensaje del asistente, pasa con `flow-build` o `create-design` en la sintaxis de cualquier host, da `review` si la sesión termina dentro de la etapa, y no juzga una sesión hija

**Closes:** AC5

**Depends on:** none

**Locations:** `tests/pilot/session_audit_routes.go`, `tests/pilot/session_audit.go` (`auditCriteria`), `tests/fixtures/sessions/plan_without_build_route/`, `tests/pilot/session_audit_test.go`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- La etapa va de la carga de `flow-plan` (`readsFlowSkill`) a la siguiente carga de un flujo distinto, o al final de la sesión. Releer `flow-plan` no la cierra.
- Un mensaje que solo nombra `/flow-plan` u otro flujo no la aprueba.
- La fila del rubric «Moving between flows ends with a next-step block» (línea 39) apunta a `plan_without_build_route` en su columna de observación, en vez de agregar otra fila. `rule-enforcers.md` gana una fila.

Fixtures:
- `fail`.
- `pass` con el bloque y `/flow-build`.
- `review` con la sesión abierta.
- Una relectura de `flow-plan` antes del bloque (`pass`).
- Un mensaje que solo nombra `/flow-plan` (`fail`).
- Una hija (`not_observed`).

**Verification:** `go test ./tests/pilot/ -run 'TestAuditNewCriteria|TestAuditRunsTheSevenExistingCriteria'` pasa. La auditoría de la copia de facti da `plan_without_build_route` en `fail` con evidencia en la línea 210.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./tests/pilot/ ./tests/content/` and `go vet ./tests/pilot/` | Covered by the implementation |
| Fixture verdicts | Final candidate, locally | `<barrido>` with the base binary and with the candidate. They differ only in the rows of the new fixtures and in the new column `plan_without_build_route` | Covered by the implementation |
| Real session | Final candidate, locally | `go run ./tests/pilot --audit-session <copia de facti> --host codex`: AC1–AC7 and AC9 against the lines named | Covered by the implementation |
| Code review | Final candidate, before the push | `Native: Claude Code /code-review` | The review answer |

## Checkpoint

- Updated: 2026-11-21
- Stage: build · local CI
- Last commit: 6818b23f
- Open findings: none
- Next: flow-verify: code review /code-review, then push
- Worktree: .claude/worktrees/audit-round-detection · feat/audit-round-detection
- Rules: flow-build/stage.md
- Principles: none

### Evidence

- before captures: none (no screen)
- base sweep: 329 fixtures read with pilot-base @1b87f0a7 (scratchpad verdicts-base.tsv)
- go test ./... · ok (4m23s) · 6818b23f
- go vet ./tests/pilot/ · ok · 6818b23f
- fixture sweep · 342 fixtures, no verdict change outside new fixtures and the new column · 6818b23f
- facti audit · AC1–AC7 lines 278/278/278/266/210/83/1784 · 6818b23f

## Review status and progress

2026-11-21 — T1 · 3a4454c4 · unverified: none
2026-11-21 — T2 · 742ed7b2 · unverified: none
2026-11-21 — T3 · 2f2de663 · unverified: none
2026-11-21 — T4 · 995e234a · unverified: none
2026-11-21 — T5 · 6818b23f · unverified: none
