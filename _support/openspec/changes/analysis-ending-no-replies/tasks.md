# Tasks

Base `development` en `0c7381c5`, rama `fix/analysis-ending-no-replies`. Todo en el hilo principal: dos tareas pequeñas sobre un solo criterio.

### T1 — La rama nueva de `analysis_reply_block` con sus pruebas

- [x] `analysis_reply_block` falla un análisis de 6 o más líneas que termina sin respuestas en un turno sin escrituras, y no evalúa los finales cortos ni los turnos de trabajo

**Closes:** AC1, AC2

**Depends on:** none

**Locations:** `tests/pilot/session_audit_endings.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/analysis_reply_block/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** una rama nueva en `analysisReplyBlock` para el final sin lista, con su evidencia; fixtures sintéticos de Codex (falla sin bloque, final corto) y de Claude (turno con escritura). No cambia la rama de la lista ni otros criterios.

**Verification:** `go test ./tests/pilot -run 'NewCriteria|ReplyList|AnalysisReply'` pasa, visto fallar antes de la rama nueva; `go run ./tests/pilot --host codex --audit-session <rollout de kin>` da `fail`.

### T2 — Medición sobre sesiones reales y documentación

- [x] El conteo sobre las sesiones reales recientes queda registrado y la spec, `rule-enforcers.md`, la rúbrica y el changelog nombran la rama nueva

**Closes:** AC3

**Depends on:** T1

**Locations:** `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** correr el auditor sobre las sesiones recientes de `~/.claude/projects` y `~/.codex/sessions`, leer cada final marcado por la rama nueva y ajustar el umbral si marca finales que la regla no gobierna; actualizar las filas de la documentación.

**Verification:** el conteo y la lectura de los marcados quedan en «Review status and progress»; `go test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y el CI local | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` de Claude Code | The review answer |

## Checkpoint

- Updated: 2026-11-20
- Stage: flow-verify · code review
- Last commit: 45147880
- Open findings: none
- Next: read /code-review findings, then push and PR
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/analysis-ending-no-replies · fix/analysis-ending-no-replies

### Evidence

- go vet ./... · pass · 917a5b20
- go test ./... · pass, 22 packages · b9935a58 tree
- python3 -m unittest discover -s tests/skills -p '*_test.py' · pass, 314 tests · b9935a58 tree
- blast radius: base vs candidate auditor over 249 real sessions · other criteria changed 0, analysis_reply_block changed 5 · 45147880

## Review status and progress

2026-11-20 — T1 · b9935a58 · unverified: none
2026-11-20 — T2 · 917a5b20 · unverified: CAM-25 flag is a clarifying question the rule may accept; analyses in later turns of a research are not seen
2026-11-20 — Medición (AC3): 249 sesiones reales del 2026-11-13 al 2026-11-20 (191 de Codex en `~/.codex/sessions`, 58 de Claude Code en `~/.claude/projects`), auditadas con el binario de `917a5b20`. La rama nueva marca 5 finales: kin `01a1f7d9` línea 69, CAM-20 `01a1f5d4` línea 64, CAM-25 `01a1f5f6` línea 89, una respuesta de Jira de fleetsystem (línea 84) y una revisión adversarial (línea 107). La rama de la lista sigue con 4 fallos y 204 aprobados, igual que antes del cambio.
