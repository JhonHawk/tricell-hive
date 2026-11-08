# Tareas

Base: `development` en `19912662`; rama propuesta `feat/model-support-floor`. Una sola tarea en el hilo principal: edita documentación y agrega un test de contenido (menos de 5 lecturas grandes). T1–T4 y T6 pasaron a `model-qualification-tool` por D13 A; ninguna estaba hecha.

### T5 — Regla, documento del piso y entradas para personas

- [ ] `AGENTS.md` separa el objetivo de escritura del piso y cuenta observaciones en candidatos con al menos una sesión real; `model-floor.md` lista los nueve modelos con su normalización; un test de contenido lo alinea con `agent-profiles.json` y `flashModels`; la rúbrica, `pilot-procedure.md`, el índice de `harness-engineering`, README, `llms.txt` y la portada del sitio apuntan a él.

**Closes:** AC27, AC31, AC33, AC34, AC35, AC36, AC37, AC38

**Depends on:** ninguno

**Locations:** `AGENTS.md`, `_support/docs/harness-engineering/model-floor.md` (nuevo), `_support/docs/harness-engineering/pilot-procedure.md`, `_support/docs/harness-engineering/README.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `README.md`, `llms.txt`, `site/src/content/docs/index.mdx`, `tests/content/model_floor_test.go` (nuevo)

**Execution:** hilo principal

**Test approach:** tdd para `model_floor_test.go` (primero falla sin el documento); check para el texto

**Changes:** lo descrito en `design.md` «Forma propuesta». Los IDs completos de `opus` y `sonnet` se toman de la documentación actual de modelos de Claude Code (Context7 o la documentación oficial), con su URL y fecha junto a la tabla y en el pull request; no se escriben de memoria ni se corre un modelo. Antes de editar, buscar en `content/`, `site/`, `_support/docs/`, `_support/openspec/specs/` y `tests/` los términos `least capable`, `two independent`, `real session`, `models in use`, `smaller model` y `execution profile`, y listar en el pull request los lugares revisados y el tamaño de `content/guidance/global.md` (sin cambios). Ningún cambio en `content/`.

**Verification:** con `GS` = `awk '/^## Guidance structure/,/^## README/' AGENTS.md` y `MS` = `awk '/^## Measurement/,/^## Repository map/' AGENTS.md`:
- AC27: `GS` contiene `model-floor`, `execution`, `focused`, `reasoning` y `verifier`; `rg -n "least capable|whatever model the user's session runs" AGENTS.md` no encuentra nada.
- AC34: `GS` contiene `Measurement` y `frontier model's prompting guide`; no contiene `those models` ni `behave correctly without it`.
- AC35: `MS` contiene `at least one`, `real session`, `authorized`, `qualification`, `pilot`, `current version`, `no-guidance arm`, `at most one observation` y `user's approval to adopt it at once`; no contiene `two independent real sessions`.
- AC36: `MS` contiene `outside the floor`, `model-floor`, `configured model` y `current version of the governing rule`.
- AC37: `MS` contiene `user's decision`, `real-session evidence`, `floor's candidates` y `not enough on their own` (o la frase elegida, citada en el pull request); no contiene `models in use`.
- AC31: `MS` contiene, en el párrafo del watch item, `model` y `run directory`, y `episodes` junto a `authorized runs`; en el párrafo del commit, `host, model`.
- AC38: `go test ./tests/content/ -run ModelFloor` pasa y sus casos con una fila quitada, un esfuerzo cambiado y un alias sin ID fallan dentro del test; en `model-floor.md`, una búsqueda por separado encuentra cada uno de: `model-qualification-tool`, la normalización, `resolves to` con la URL y la fecha de la documentación leída, `90%` con `compare_none`, `authorization` con `spending cap`, `provisional`, `ninguna` (o `none`) y los límites (`no-guidance arm`).
- AC33: `rg -n "model-floor" README.md llms.txt site/src/content/docs/index.mdx _support/docs/harness-engineering/README.md _support/docs/harness-engineering/pilot-procedure.md _support/docs/harness-engineering/session-review-rubric.md` encuentra cada archivo; `rg -n "two independent sessions" _support/docs/harness-engineering/session-review-rubric.md` no encuentra nada; desde `site/`, `pnpm check && pnpm build` pasa.
- Lectura del candidato: cada cláusula de AC27, AC31, AC33 y AC34–AC38 se marca contra el texto final en el pull request.
- `go test ./tests/content/...` pasa.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de la tarea | Al cerrar T5, local | Los comandos de T5 pasan | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./... && go test ./...` y `pnpm check && pnpm build` en `site/` | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code (D6 A) | La respuesta de revisión |

Sin UI propia: la portada del sitio cambia solo de texto. Sin modelos reales.

## Checkpoint

## Estado de revisión y progreso

- 2026-11-08 — plan escrito.
- 2026-11-08 — revisión 1 y revisión 2 sobre la versión con herramienta; ninguna tarea estaba hecha.
- 2026-11-08 — revisión 3: D12 y D13; T1–T4 y T6 pasan a `model-qualification-tool`; T5 cierra AC19–AC21.
- 2026-11-08 — revisión 4: D14; T5 cierra AC20–AC22.
- 2026-11-08 — revisión 5: el piso suma los frontier y «Measurement» cambia en las líneas 86, 88 y 90; T5 cierra AC23–AC26.
- 2026-11-08 — revisión 6: D15; objetivo de escritura separado del piso, `focused`, pertenencia por modelo, rúbrica; T5 cierra AC27–AC33.
- 2026-11-08 — revisión 7: D16 y D17; T5 cierra AC27, AC31, AC33 y AC34–AC38.
- 2026-11-08 — séptima revisión de plan: 2 P1 en el texto de T5 (lectura final con el rango viejo; fuente de los IDs de alias). Revisión 8: corregidos en T5 y en Prerrequisitos, sin cambiar criterios ni el delta.
