# Tasks

Base `development` en `ba160011`, rama `refactor/audit-prefix`. Todo en el hilo principal: es un reemplazo mecánico con script.

### T1 — Renombrar las skills y sus referencias

- [x] Las tres carpetas, el catálogo, el `testdata` y los tests renombrados; cada uso literal actualizado fuera del historial

**Closes:** AC1, AC3

**Depends on:** none

**Locations:** `content/`, `site/`, `README.md`, `llms.txt`, `_support/docs/`, `integrations/opencode/`, `tooling/`, `tests/`, `.gitignore`, `CHANGELOG.md`

**Execution:** main thread

**Changes:** `git mv` de las carpetas y archivos con el nombre; reemplazo de `comment-audit`, `test-audit` y `harness-audit` (y `<harness-audit-dir>`); identificadores Go y Python derivados. No cambia el historial ni el upstream de OpenClaw.

**Verification:** `rg 'comment-audit|test-audit|harness-audit'` fuera del historial solo devuelve las menciones a propósito; `go test ./...` y los tests de `integrations/opencode` pasan.

### T2 — Retirar los nombres viejos

- [x] `retiredSkills` incluye los tres nombres; la URL del upstream de OpenClaw sigue legal

**Closes:** AC2, AC4

**Depends on:** T1

**Locations:** `tests/content/retired_names_test.go`, `tests/pilot/session_audit.go`

**Execution:** main thread

**Changes:** añadir los nombres a `retiredSkills`, una excepción para la URL del upstream y su caso legal; `hiveSkillName` reconoce el nombre nuevo y conserva el viejo para las sesiones ya grabadas.

**Verification:** `go test ./tests/content/ ./tests/pilot/` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | CI local (jobs del workflow) | Covered by the implementation |
| Code review | — | Sin revisión: cambio mecánico | The review answer |

## Checkpoint

- Updated: 2026-11-22
- Stage: build · 2
- Last commit: ba160011
- Open findings: none
- Next: T1 rename
- Rules: flow-build/stage.md · extra: spec-delta, change-folder, delivery-ask, proportionality
- Principles: none

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-22 — T1 · uncommitted · unverified: none
2026-11-22 — T2 · uncommitted · unverified: none
