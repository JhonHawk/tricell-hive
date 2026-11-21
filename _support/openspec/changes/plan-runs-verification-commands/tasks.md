# Tareas

Base `04a1f27b` de `development`, rama `feat/plan-runs-verification-commands`. Todo corre en el hilo principal: son ediciones de texto en 9 archivos y ninguna tarea pasa de 5 lecturas grandes.

### T1 — La regla en su hogar y sus punteros

- [x] `flow-plan` paso 3 pide ejecutar los comandos de verificación; «Revise», el revisor, la plantilla y `plan-handoff` lo apuntan; `single_home_test.go` protege la frase.

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/hive-rule-change-folder/SKILL.md`, `content/agents/review/hive-review-plan.md`, `content/skills/flow-plan/assets/change-folder/proposal.md`, `content/skills/flow-plan/references/plan-handoff.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** los textos de `design.md` «Texto de la regla», incluida la separación de los pasos 4 y 5 de `flow-plan/SKILL.md`. No cambia `global.md`, `hive-rule-external-facts` ni `check_change.py`.

**Verification:** `go test ./tests/content/...` pasa. Con la frase `run each command it names once on the base` copiada temporalmente en `content/skills/hive-rule-change-folder/SKILL.md`, `go test ./tests/content/ -run TestRulesHaveOneHome` falla; después se borra la copia y vuelve a pasar (AC4). AC1 a AC3 se observan en los archivos instalados tras `hive update`, después de integrar.

### T2 — Spec, sitio y registros de revisión

- [x] El delta de `workflow-lifecycle`, la página del plan, la fila de la rúbrica y la fila de `rule-enforcers.md` describen la regla.

**Closes:** AC1

**Depends on:** T1

**Locations:** `specs/workflow-lifecycle/spec.md` de esta carpeta, `site/src/content/docs/flows/plan.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en el sitio, una viñeta en «What it produces» (los comandos de verificación ejecutados y su línea), y los párrafos de «Revising a plan in progress» (volver a ejecutar un comando cambiado) y «How it ends» (el resumen nombra los comandos `not measured`); una fila «Plan» en la rúbrica; una fila `judgment` en `rule-enforcers.md`; una entrada en «Changed» de `[Unreleased]` en `CHANGELOG.md`. No toca `README.md` ni `llms.txt`: no cambian nada de lo que dicen.

**Verification:** `go test ./tests/content/...` pasa, incluidos `spec_deltas_test.go` y `rule_enforcers_test.go`; `pnpm --dir site check && pnpm --dir site lint && pnpm --dir site build` pasa; `python3 -I ~/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/plan-runs-verification-commands` sale 0.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, en local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, en local | `go vet ./... && go test ./...` (4 min 14 s en la base) | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push o de la parada de validación | Según la respuesta de revisión | La respuesta de revisión |

Tras integrar, la actualización local (`hive update --source`) instala el contenido; AC1 a AC3 se leen en los archivos instalados.

## Checkpoint

- Updated: 2026-11-21
- Stage: build · tasks done
- Last commit: b645ea7a
- Open findings: none
- Next: local CI, then flow-verify

### Evidence

- before captures: none (no interface)
- go test ./tests/content/... · pass · b645ea7a
- pnpm --dir site check/lint/build · pass · b645ea7a

## Estado de la revisión y avance

2026-11-21 — T1 · ccbfddf2 · unverified: AC1-AC3 are read in the installed files after hive update
2026-11-21 — T2 · b645ea7a · unverified: none
