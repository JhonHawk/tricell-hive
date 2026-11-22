# Tasks

Base `013a9725` en `development`, rama `refactor/size-limits`, worktree `.claude/worktrees/size-limits`. Todo en el hilo principal: cada tarea edita pocos archivos ya leídos.

### T1 — Tests de tamaño por concepto

- [x] Hojas, principios, playbooks y referencias comparten el tope de lectura entera de 24.000 B; `SKILL.md` mantiene 16.000 B; cada constante cita su fuente y cada mensaje dice qué hacer

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `tests/content/leaf_skill_test.go`, `tests/content/skill_size_test.go`, `tests/content/reading_footprint_test.go`, `tests/content/budget_test.go`, `tests/content/voices_test.go`

**Execution:** main thread

**Test approach:** check

**Changes:** quita `leafSizeLimit`, `leafExceptionSizeLimit`, `leafReportEndingsSizeLimit`, `playbookSizeLimit` y `leafSizeExceptions`; agrega `wholeReadLimit = 24000` usado por hojas, principios, playbooks y referencias; los tests de frontera existentes pasan a 24.000/24.001. No toca `tooling/`.

**Verification:** `go test ./tests/content/...` pasa, y los tests de frontera prueban 24.000 (pasa) y 24.001 (falla con el mensaje que nombra qué hacer).

### T2 — Textos que repiten los topes

- [x] Ningún texto del repositorio fuera de los cambios archivados da 16.000 B para referencias ni 4.000/7.000/9.000 B para hojas

**Closes:** AC5

**Depends on:** T1

**Locations:** `AGENTS.md`, `content/skills/harness-audit/references/skills-and-agents.md`, `_support/docs/harness-engineering/harness-audit-rules.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** main thread

**Changes:** actualiza cada frase al tope nuevo y su razón; agrega la entrada de `CHANGELOG.md`. No toca el texto de cambios archivados.

**Verification:** `rg -n '16,000|4,000|7,000|9,000'` sobre esos archivos muestra solo el tope de `SKILL.md`; `go test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` and the local CI jobs | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |

## Checkpoint

- Updated: 2026-11-22
- Stage: flow-build · 5
- Last commit: aa87439c
- Open findings: none
- Next: flow-verify: /code-review, then push, PR and merge (automatic)
- Worktree: .claude/worktrees/size-limits · refactor/size-limits
- Rules: flow-build/stage.md · extra: spec-delta, change-folder, prose, delivery-ask
- Principles: none

### Evidence

- go vet ./... && go test ./... · pass (23 packages) · aa87439c
- before captures: none (no UI)
- CI: only master-source-branch.yml (PR guard toward master), no job to run locally

## Review status and progress

2026-11-22 — T1 · b6a7ecaa · unverified: none
2026-11-22 — T2 · aa87439c · unverified: none
