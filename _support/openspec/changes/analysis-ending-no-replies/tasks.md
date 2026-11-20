# Tasks

Base `development` en `0c7381c5`, rama `fix/analysis-ending-no-replies`. Todo en el hilo principal: dos tareas pequeñas sobre un solo criterio.

### T1 — La rama nueva de `analysis_reply_block` con sus pruebas

- [ ] `analysis_reply_block` falla un análisis de 6 o más líneas que termina sin respuestas en un turno sin escrituras, y no evalúa los finales cortos ni los turnos de trabajo

**Closes:** AC1, AC2

**Depends on:** none

**Locations:** `tests/pilot/session_audit_endings.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/analysis_reply_block/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** una rama nueva en `analysisReplyBlock` para el final sin lista, con su evidencia; fixtures sintéticos de Codex (falla sin bloque, final corto) y de Claude (turno con escritura). No cambia la rama de la lista ni otros criterios.

**Verification:** `go test ./tests/pilot -run 'NewCriteria|ReplyList|AnalysisReply'` pasa, visto fallar antes de la rama nueva; `go run ./tests/pilot --host codex --audit-session <rollout de kin>` da `fail`.

### T2 — Medición sobre sesiones reales y documentación

- [ ] El conteo sobre las sesiones reales recientes queda registrado y la spec, `rule-enforcers.md`, la rúbrica y el changelog nombran la rama nueva

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
- Stage: flow-build · T1
- Last commit: 0c7381c5
- Open findings: none
- Next: T1: fixtures and failing tests
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/analysis-ending-no-replies · fix/analysis-ending-no-replies

### Evidence


## Review status and progress
