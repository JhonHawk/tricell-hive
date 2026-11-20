# Tasks

Base: `development` en `1cdb5af7`, rama `feat/floor-versions-reply-block` en `.claude/worktrees/floor-versions`. T2 edita más de 3 archivos del auditor (criterio nuevo, registro, pruebas, fixtures) y va a un hijo `hive-build-backend`. T1 y T3 quedan en el hilo principal mientras el hijo trabaja, porque sus archivos no se cruzan con los del auditor.

Los commits de T2 nombran la observación: Claude Code, `claude-opus-5-5`, sesión `bc3b8e52-20f4-47fb-a275-1e3fb437e088`, tricell-hive, 2026-11-20, dos finales de análisis sin bloque `▶` ni letras; aprobado por el usuario sin recurrencia.

### T1 — Perfiles y tabla del floor en las versiones de hoy

- [x] `agent-profiles.json` y la tabla de `model-floor.md` nombran `gpt-6.2-sol` y `grok-4.7` donde nombraban `gpt-6.1-sol` y `grok-4.6`, con las pruebas que leen los perfiles al día.

**Closes:** AC1, AC2, AC3

**Depends on:** ninguna

**Locations:** `integrations/agent-profiles.json`, `_support/docs/harness-engineering/model-floor.md`, `_support/docs/architecture/agent-delivery.md` (solo frases de estado actual), las pruebas que fallen por leer los perfiles

**Execution:** hilo principal

**Test approach:** check

**Changes:** Los supuestos de `proposal.md` sobre Codex, Grok y la tabla: Codex `execution`, `reasoning` y `verifier` a `gpt-6.2-sol` con sus esfuerzos; `focused` sin cambio; Grok `verifier` a `grok-4.7`. En la tabla, la fila del objetivo de escritura de `gpt-6.2-sol` absorbe la del modelo de las sesiones del usuario y suma `max`; `gpt-6.1-sol` y `grok-4.6` se quedan con `Writing target` en `no` y un origen que dice hasta cuándo fueron perfil; una línea «Row changes» bajo la tabla da la fecha. Antes de editar, `rg -n -i 'gpt-?6\.1[ -]sol|grok[ -]4\.6'` en `content/`, `integrations/`, `tooling/`, `tests/` y `_support/docs/` separa las menciones que leen los perfiles o describen el estado actual (como la tabla de perfiles de `agent-delivery.md`, líneas 13 a 23) de las que son datos de una prueba o historia fechada; solo cambian las primeras.

**Verification:** `go -C <worktree> test ./tests/content/... ./tests/pilot/... ./integrations/... ./tooling/...` pasa. Tras el commit, `hive update --source <worktree> --out <tmp>/plan.json` y leer en ese plan los bytes `managed` de un rol de Codex de cada perfil y de `hive-verify-task` de Grok: nombran los modelos nuevos. AC1 y AC2 se observan en los roles instalados tras el merge, al recompilar y desplegar.

### T2 — Criterio `analysis_reply_block` del auditor

- [x] `--audit-session` marca el final de un turno que ofrece respuestas en una lista sin letras o sin encabezado `▶`.

**Closes:** AC4

**Depends on:** ninguna

**Locations:** `tests/pilot/session_audit.go` (registro), un archivo nuevo `tests/pilot/session_audit_endings.go`, `tests/pilot/session_audit_test.go`, fixtures nuevas en `tests/fixtures/sessions/analysis_reply_block/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** La regla del supuesto de `proposal.md`: el último texto del agente en cada turno de la sesión principal, el que va antes del siguiente mensaje del usuario (los turnos se cortan como en `claudePromptContent`) o del fin de la sesión, falla cuando termina con dos o más ítems de lista cuyo rótulo va en cursiva y no es una letra sola, o con ítems de letra (`A`, `B`…) sin una línea encima que tenga `▶` o sea una pregunta en negrita. No se evalúa un turno que terminó con una tarjeta de preguntas o un comando de flujo, un texto que `isTextRound` reconoce como ronda de decisiones, ni una sesión hija (como `ticket_assessment_ending`). La evidencia nombra la línea y dice si antes se leyó el `SKILL.md` instalado de `hive-rule-report-endings`, bajo una carpeta oculta como `hiveModeRead`, no `content/skills/`. `not_observed` cuando ningún turno termina con una lista así. Fixtures sintéticas: un final con viñetas en cursiva (falla), un final con `▶` y letras (pasa), un final con `**¿Cómo seguimos?**` y letras sin `▶` (pasa), un final con letras sin encabezado (falla), un reporte que termina con hallazgos rotulados en negrita como `- **F7:** …` (no observado), una ronda de decisiones en texto (no observado), y Codex y Grok con un caso cada uno.

**Verification:** `go -C <worktree> test ./tests/pilot/...` y `go -C <worktree> vet ./tests/pilot/...` pasan, cada caso nuevo visto fallar antes. `--audit-session` sobre `bc3b8e52` da `fail` en las líneas 50 y 97 con la lectura de la regla ausente; sobre `7ee42122` no da `fail`. Como control, los arreglos `criteria` de `bc3b8e52`, `7ee42122`, Grok `01a1f5fc` y Codex `01a1f5d4` no cambian en ningún otro criterio frente a un `git archive` de `1cdb5af7`, y sobre las 40 sesiones de Claude más recientes de tricell-hive el criterio no da `fail` fuera de `bc3b8e52`.

### T3 — Documentación

- [x] La rúbrica, `rule-enforcers.md` y el changelog dicen lo que cambió.

**Closes:** AC3, AC4

**Depends on:** T1, T2

**Locations:** `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Changes:** La rúbrica gana una fila (Reports) para `analysis_reply_block`; `rule-enforcers.md` anota que el auditor detecta el final de un análisis sin bloque de respuestas; el changelog nombra el criterio y las versiones nuevas de los perfiles.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...`, `go -C <worktree> vet ./...` and `python3 -m unittest discover -s <worktree>/tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |

Sin parada de validación (`automatic`). Tras el merge: recompilar el binario y desplegar, y leer los roles instalados para AC1 y AC2.

## Checkpoint

- Updated: 2026-11-20
- Stage: verify · integrate
- Last commit: 314b7317
- Open findings: none
- Next: push, PR, merge; rebuild binary; deploy; read installed roles; flow-close
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/floor-versions · feat/floor-versions-reply-block

### Evidence

- before captures: none (no screen)
- go -C <wt> test ./tests/content/... ./integrations/... ./tooling/... · ok · 960b79d5
- go -C <wt> test ./tests/pilot/... · ok · 9fa629a3
- python3 -m unittest discover -s <wt>/tests/skills · OK · bc805892
- go -C <wt> vet ./... · ok · bc805892

## Review status and progress

2026-11-20 — T1 · 960b79d5 · unverified: AC1 and AC2: installed roles after rebuild and deploy
2026-11-20 — T2 · 9fa629a3 · unverified: none
2026-11-20 — T3 · bc805892 · unverified: none

2026-11-20 — Revisiones: /code-review (3 importantes del detector y 5 menores corregidos en c8fbd249 y c9a3828b; las 40 sesiones recientes fallan solo en bc3b8e52 y en un bloque de siguiente paso sin encabezado de bd34920c:972); segunda /code-review sobre bc805892..c9a3828b: 5 importantes de precisión del detector (regla --- en el cuerpo, letras con puntuación, líneas de continuación, comando en el encabezado, listas del cuerpo unidas) y 5 menores, no bloqueantes, quedan para el cierre; textos de prueba en inglés y la referencia del floor corregidos.
