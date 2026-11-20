# Tareas

Base `development` en `27875867`, rama `feat/background-wait-rule`. T1 y T4 corren en el hilo principal, porque cada una toca 3 archivos o menos. T2 corre en un hijo `hive-build-backend`, porque toca más de 3 archivos y su interfaz ya está fijada en `design.md`. T3 también corre en un hijo, pero después de T2: las dos escriben `session_audit.go` y `session_audit_test.go`.

### T1 — Hoja `hive-rule-background-wait` y su línea en la guía global

- [x] La hoja existe con las cinco reglas de `design.md` y el índice de hojas la enlaza en el momento de lanzar una espera.

**Closes:** AC4

**Depends on:** none

**Locations:** `content/skills/hive-rule-background-wait/SKILL.md` (nueva), `content/guidance/global.md`, `tests/content/single_home_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** Se añaden la hoja y la línea del índice, y se registra la frase canónica en `single_home_test.go`. No se añaden enlaces desde otras skills (supuesto en `proposal.md`). No cambian ni `merge gate` ni `child brief`.

**Verification:** `go test ./tests/content/...` pasa. Esto incluye el tamaño de la hoja, los enlaces y el presupuesto de lectura. El AC4 se observa después de integrar, en `~/.claude/CLAUDE.md` y `~/.agents/skills/hive-rule-background-wait/SKILL.md` tras `hive update`.

### T2 — Criterio `unbounded_background_wait`

- [x] `--audit-session` reporta el criterio para Claude Code y Grok, y da `not_observed` en Codex.

**Closes:** AC1, AC3, AC5

**Depends on:** none

**Locations:** `tests/pilot/session_audit_waits.go` (nuevo), `tests/pilot/session_audit.go` (registro), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/unbounded_background_wait/`

**Execution:** child `hive-build-backend`

**Test approach:** check: fixtures con la estructura real de cada host y valores sintéticos. Casos:
- Claude y Grok: bucle sin plazo → `fail`;
- `timeout N` que envuelve el bucle → `pass`;
- `timeout` dentro del cuerpo → `fail`;
- `for` sobre `$(seq 1 60)` → `pass`;
- contador comparado que lleva a `exit` → `pass`;
- `$SECONDS` solo impreso → `fail`;
- bucle en primer plano → `not_observed`;
- Codex → `not_observed`.

**Changes:** Se añade el criterio tal como lo describe `design.md`. No juzga si la condición de salida cubre todos los resultados.

**Verification:** `go test ./tests/pilot/...` pasa. `go run ./tests/pilot --audit-session _support/workspace/background-wait-rule/grok-01a1f7e8-3909.jsonl --host grok` muestra `unbounded_background_wait` en `fail` con la línea 3855.

### T3 — `merged_branch_deleted` da fail solo después de un cierre

- [x] En la auditoría, una sesión sin mensaje de cierre después de su último merge da `review`, y una con cierre y sin borrado sigue dando `fail`.

**Closes:** AC2

**Depends on:** T2

**Locations:** `tests/pilot/session_audit.go` (`existingAuditCriteria`), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/existing_criteria/`

**Execution:** child `hive-build-backend`

**Test approach:** check: `claude-project-merged-cleanup-fail.jsonl` y `claude-project-local-merge.jsonl` reciben un cierre después del merge y siguen dando `fail`. Un fixture nuevo sin cierre da `review`. Los casos de Codex (`session_audit_test.go:641`) no cambian.

**Changes:** Solo cambia la auditoría de sesiones reales; `regression.go` no se toca.

**Verification:** `go test ./tests/pilot/...` pasa. `go run ./tests/pilot --audit-session _support/workspace/background-wait-rule/grok-01a1f7e8-3909.jsonl --host grok` muestra `merged_branch_deleted` en `review`.

### T4 — Rúbrica, registro de enforcers y CHANGELOG

- [x] La rúbrica y `rule-enforcers.md` tienen la fila del criterio nuevo. La rúbrica explica cómo leer el `review` de `merged_branch_deleted`, la fila de limpieza de ramas nombra `session_audit.go`, y `CHANGELOG.md` tiene su entrada.

**Closes:** none

**Depends on:** T2, T3

**Locations:** `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** main thread

**Changes:**
- Una fila del criterio nuevo en cada documento.
- Una frase en la línea 7 de la rúbrica para el `review` de `merged_branch_deleted`.
- `session_audit.go` añadido en la fila de limpieza de ramas de `rule-enforcers.md`.
- La entrada en `[Unreleased]` de `CHANGELOG.md`.

El sitio no cambia: ninguna página describe esperas fuera de CI.

**Verification:** `go test ./tests/content/...` pasa, incluido `rule_enforcers_test.go`.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, en local | `go vet ./...` y `go test ./...` | Cubierto por la implementación |
| Auditoría de la sesión real | Candidato final, en local | AC1 y AC2 con el comando de T2 | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push o de la parada de validación | <mecanismo elegido> | La respuesta de revisión |

## Checkpoint

- Updated: 2026-11-20
- Stage: verify · 3 (fix round)
- Last commit: 582be0ff
- Open findings: CR1-6 bounded waits read as fail · fixing (simplified deadline rule); CR7 /bin/sleep and nested sleep missed · fixing; CR8 CHANGELOG Changed heading · fixed; CR9 leaf deadline vs stretches wording · fixed; CR10 merge located by evidence string · non-blocking, close findings
- Next: On child return: verify, commit fix round (detector + leaf/spec/design/CHANGELOG), focused re-review of the fixes' diff, push/PR/merge

### Evidence

- before captures: none (no screen)
- go test ./tests/content/... · pass · 95075c97
- go test ./tests/pilot/... · pass · 99e3ee80
- audit copy: unbounded_background_wait fail :3855 · 99e3ee80
- go vet ./... && go test ./... · pass · 95fbdfb3
- go vet ./... && go test ./... · pass · 582be0ff
- audit copy: unbounded_background_wait fail :3855, merged_branch_deleted review · 582be0ff
- AC1 audit copy unbounded_background_wait fail :3855 · 582be0ff
- AC2 audit copy merged_branch_deleted review · 582be0ff
- AC3/AC5 audit CLI on fixtures: unbounded fail (claude, grok), timeout-wrap/seq/while-condition pass, timeout-body fail · 582be0ff
- comments: 0 deleted, all Go doc comments kept (declared limits); suppressions: none
- blast radius: real-session path only; pilot detector unchanged, go test ./... pass (rung 4); README/llms need no change

## Estado de la revisión y avance

2026-11-20 — T1 · 95075c97 · unverified: AC4 is observed in the installed guidance after hive update, after integration
2026-11-20 — T2 · 99e3ee80 · unverified: none
2026-11-20 — T3 · 86d12e91 · unverified: none
2026-11-20 — T4 · 95fbdfb3 · unverified: none
