# Tareas

Base `49ef7757` de `development`, rama `feat/conditional-plan-reviewers`. Todo corre en el hilo principal: son ediciones de texto en 11 archivos y ninguna tarea pasa de 5 lecturas grandes.

### T1 — La referencia, el paso 5, «Revise» y los recortes

- [x] `plan-review.md` existe con la tabla, el tope y la reconciliación; el paso 5 y «Revise» la apuntan; los recortes y la tolerancia de D10 mantienen el presupuesto.

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `content/skills/flow-plan/references/plan-review.md` (nuevo), `content/skills/flow-plan/SKILL.md`, `content/skills/hive-rule-change-folder/SKILL.md`, `content/skills/hive-rule-reviewer-brief/SKILL.md`, `content/agents/review/hive-review-plan.md` (solo recortes), `tests/content/reading_footprint_test.go` (tolerancia)

**Execution:** hilo principal

**Test approach:** check

**Changes:** los textos de `design.md`. No cambia el rol `hive-review-plan` salvo los recortes, ni `spec-delta.md`, ni el mapeo de OpenCode.

**Verification:** `go test ./tests/content/...` pasa: la unión plan/build queda por debajo de la línea base más la tolerancia, `hive-rule-change-folder` en 4000 B o menos (`TestLeafSkillInvariants`) y la referencia en 16 000 B o menos (`skill_size_test.go`). `go test ./tooling/management/ -run 'Reference|Anchor'` pasa: revisa los enlaces y las anclas de la referencia nueva. AC1 a AC3 se leen en los archivos instalados tras `hive update`, después de integrar.

### T2 — Spec, sitio, rúbrica y changelog

- [x] El delta de `workflow-lifecycle`, la página del plan, la fila de delegación de la rúbrica y el changelog describen el reparto por dominio.

**Closes:** AC1 (lo describe en el sitio y el spec; lo cierra T1)

**Depends on:** T1

**Locations:** `specs/workflow-lifecycle/spec.md` de esta carpeta, `site/src/content/docs/flows/plan.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en el sitio, la frase «one independent reviewer checks the folder» (línea 47) y la de «Revising a plan in progress» (línea 59); la fila 64 de la rúbrica; una entrada en «Changed» de `[Unreleased]`. No toca `README.md` ni `llms.txt`: no cambian nada de lo que dicen.

**Verification:** `go test ./tests/content/...` pasa, incluido `spec_deltas_test.go`; `pnpm --dir site check && pnpm --dir site lint && pnpm --dir site build` pasa; `python3 -I ~/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/conditional-plan-reviewers` sale 0.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, en local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, en local | `go vet ./... && go test ./...` (3 min 51 s en la base) | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | `/code-review` | La respuesta de revisión |

Tras integrar, la actualización local (`hive update --source`) instala el contenido; AC1 a AC3 se leen en los archivos instalados.

## Checkpoint

- Updated: 2026-11-21
- Stage: verify · fixes done
- Last commit: a474c23b
- Open findings: none
- Next: push, PR, merge, refresh, close

### Evidence

- before captures: none (no interface)
- go test ./tests/content/... · pass · febf6665
- pnpm --dir site check/lint/build · pass
- go test ./tooling/management -run Reference|Anchor · pass · 2da238dd
- go vet ./... && go test ./... · pass · febf6665
- go test ./tests/content/... · pass · after limited-review fixes

## Estado de la revisión y avance

2026-11-21 — T1 · 2da238dd · unverified: AC1-AC3 are read in the installed files after hive update
2026-11-21 — T2 · febf6665 · unverified: none
