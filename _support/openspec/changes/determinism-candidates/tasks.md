# Tasks

Base: `tricell-hive` `development` en `8e71ca4b`, rama `feat/determinism-candidates`. T1–T2 van a un solo hijo `hive-build-backend`, en ese orden, porque comparten `tests/pilot/trace.go` y la estructura del reporte y suman más de tres archivos editados; T3–T5 son texto y tablas en el hilo principal, mientras el hijo trabaja; T6 corre en el hilo principal al final.

### T1 — Adaptadores de transcripción real y modo `--audit-session`

- [x] `go run ./tests/pilot --audit-session <archivo> --host claude|codex|grok` lee una transcripción real y escribe el JSON que fija `design.md` «Auditoría de sesiones reales», con los seis criterios de `regressionCriteria` y `merged_branch_deleted`.

**Closes:** AC1

**Depends on:** none

**Locations:** `tests/pilot/main.go`, un archivo nuevo `tests/pilot/session_audit.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/<adaptador>/`

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd — el adaptador es un contrato con formatos externos: un fixture saneado por host (Claude del proyecto; Codex con un `exec` JavaScript que mezcla `cmd` con comillas dobles, backticks e interpolación, más `apply_patch`; Grok `updates.jsonl` con `tool_call` y `tool_call_update`) visto fallar antes de escribir el adaptador

**Changes:** agrega el modo, los adaptadores, `--base` y los ajustes de `design.md` «Criterios existentes en la auditoría», que solo valen en el modo de auditoría; no cambia `parseTrace` para los pilotos, ni `--assess`, ni el comportamiento de ningún criterio en los pilotos.

**Verification:** `go test ./tests/pilot/...` pasa; el comando sobre una transcripción real de cada host en esta máquina termina con código 0 e imprime JSON válido sin texto de comandos.

### T2 — Tres criterios para sesiones reales

- [ ] `text_round_without_card`, `close_before_branch_cleanup` y `skill_named_as_reason` corren en `--audit-session`, con los estados que fija `design.md`.

**Closes:** AC1, AC9

**Depends on:** T1

**Locations:** `tests/pilot/session_audit.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/<criterio>/`

**Execution:** hijo `hive-build-backend`, a continuación de T1

**Test approach:** tdd — por criterio, un caso `pass` y su resultado `fail` o `review`, con al menos dos hosts para `text_round_without_card`, más sus exenciones (tarjeta vacía o expirada antes de la ronda, rollout de subagente) y, para `close_before_branch_cleanup`, el caso sin borrado (`not_observed`); cada `fail` o `review` se ve fallar al revertir la rama del criterio

**Changes:** agrega los criterios solo al modo de auditoría, y `close_before_branch_cleanup` reconoce también un `git merge` seguido del push de la base; `regressionCriteria` de los pilotos no cambia.

**Verification:** `go test ./tests/pilot/...` pasa, incluida la prueba de reversión registrada en el informe del hijo.

### T3 — Medición: campo `Enforceable:` y excepción a los casos nuevos

- [x] `AGENTS.md` «Measurement» pide la línea `Enforceable:` en cada seguimiento, dice que `--audit-session` no ejecuta modelos y no es un piloto pausado, y aclara que la prohibición de casos nuevos en `tests/fixtures/regression/` no alcanza a los fixtures de `tests/fixtures/sessions/`, que prueban una herramienta que sí recibe sesiones reales.

**Closes:** AC3

**Depends on:** none

**Locations:** `AGENTS.md`, `tests/fixtures/regression/README.md` (sus secciones «Add a case» y «Checks before committing» apuntan a la carpeta nueva para los criterios de auditoría)

**Execution:** hilo principal

**Changes:** tres oraciones en «Measurement» y la nota del README; no cambia el umbral de recurrencia, quién decide una corrección, ni los casos de regresión existentes.

**Verification:** lectura del diff contra la prueba de cuatro condiciones de la propuesta; `go test ./tests/content/...` pasa.

### T4 — Rúbrica del inspector y tabla regla ↔ mecanismo

- [x] La rúbrica empieza por la pasada automática y por los `candidate/` de los proyectos, y su fila de la tarjeta vacía dice la regla vigente; `rule-enforcers.md` existe, sembrado como fija `design.md`, con un test de contenido que falla cuando una ruta de mecanismo no existe.

**Closes:** AC3, AC4

**Depends on:** T2 (los archivos de los criterios que la tabla enlaza deben existir)

**Locations:** `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/README.md`, `tests/content/rule_enforcers_test.go`

**Execution:** hilo principal

**Test approach:** check — el test nuevo, con un caso negativo que apunta a una ruta inexistente

**Changes:** agrega la sección inicial de la rúbrica sin quitar filas; la tabla no reemplaza los hogares canónicos de las reglas, solo los enlaza.

**Verification:** `go test ./tests/content/...` pasa y falla con la ruta negativa.

### T5 — Candidatos del proyecto en el cierre

- [x] El informe final de `flow-close` tiene la parte «Candidates», el paso 7 crea o actualiza `candidate/<herramienta>-<objetivo>`, y el pie de cierre ofrece convertir en mecanismo un candidato visto en dos sesiones, como fija `design.md` «Candidatos del proyecto», apuntando a `defect-fix.md` punto 5 sin repetir su escalera; la página del sitio y el delta de `workflow-lifecycle` lo describen.

**Closes:** AC5, AC6, AC7, AC8

**Depends on:** none

**Locations:** `content/skills/flow-close/SKILL.md`, `site/src/content/docs/flows/close.md`, `_support/openspec/changes/determinism-candidates/specs/workflow-lifecycle/spec.md`

**Execution:** hilo principal

**Test approach:** check — los tests de contenido (tamaño y referencias) y la revisión del comportamiento completo del cierre que pide `AGENTS.md` «Guidance structure»

**Changes:** agrega la parte, la escritura de memoria y una respuesta del pie; no cambia «Findings» ni las escrituras de memoria permitidas por la guía global (el cierre ya es un momento permitido).

**Verification:** `go test ./tests/content/...` pasa; `flow-close/SKILL.md` ≤ 16,000 B; `npm --prefix site run check` y `build` terminan en 0.

### T6 — Auditoría de las sesiones revisadas a mano

- [ ] El modo corre sobre las sesiones de AC9 y reproduce los resultados de la revisión manual.

**Closes:** AC9

**Depends on:** T2

**Locations:** ninguna en el repositorio; resultados en `_support/workspace/2026-11-17-determinism-candidates/`

**Execution:** hilo principal

**Changes:** solo lectura de transcripciones; no corre modelos.

**Verification:** los JSON muestran `fail` en `text_round_without_card` para `01a1e864-77eb-7961-bd16-838efde72872` y `01a1eadf-1264-7b52-b775-9e6d6483ca6f`, `review` en `close_before_branch_cleanup` para Grok `01a1eb37-0f75-7882-b92a-d492997b9c6c` y `01a1eb3d-a6b2-7b70-8503-90e3b2ef0ace`, y ningún `fail` de `text_round_without_card` en `01a1ea8f-7b42-7370-8c2a-9468640dcaf1`.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...`, `go vet ./...`, `npm --prefix site run check/lint/build` | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` | Respuesta D5 A, en uso |

Sin parada de validación: la entrega es `ready` y el usuario fusiona el PR.

## Checkpoint

- Updated: 2026-11-17
- Stage: build · T2 (reabierta por D11 A)
- Last commit: bb18b4a5
- Open findings: none
- Next: T2 (ronda de correcciones de la revisión de código y D11 A) → T6 → verify

### Evidence

- go vet ./... && go test ./... -count=1 · all pass except tests/pilot TestE2ESetupDeadlinesOwnDescendantsAndPipes (pre-existing timing flake, also fails on an unmodified git archive of HEAD) · bb18b4a5
- npm --prefix site run check/lint/build · exit 0, 20 pages · bb18b4a5
- before captures: none (no interface change)

## Review status and progress

2026-11-17 — Plan escrito.
2026-11-17 — T3 · ac57ee7b · unverified: none
2026-11-17 — T5 · 1c619f1f · unverified: AC5, AC7, AC8 behavior: observed only in installed text after merge and local refresh, and in the first real close with a candidate
2026-11-17 — T1 · 960855c8 · unverified: none
2026-11-17 — T2 · 960855c8 · unverified: none
2026-11-17 — T4 · bb18b4a5 · unverified: none
2026-11-17 — T6 · uncommitted · unverified: 01a1ea8f's negative case holds because the audit detects no round title in its text (its rounds lived in cards), not because it judged a round as carded
2026-11-17 — Revisión por D11 A: `close_before_branch_cleanup` marca `review`, no `fail`, porque `flow-close` permite cerrar antes de limpiar cuando deja pasos pendientes. AC2 se retira y AC9 lo reemplaza; reabre T2 y T6; la fila de la tabla de reglas pasa a `judgment`; el merge reconoce también `git merge` seguido del push de la base. Revisión del plan sobre el cambio: 2 P2 y 4 P3, aplicados (razón precisa en el diseño, cómo resuelve el inspector un `review`, punto de control, AC9, T2).
