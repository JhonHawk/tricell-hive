# Tasks

Base: `development` en `19f3b7be`; rama `feat/question-format-template` en el worktree `.claude/worktrees/question-format-template`. Todas las tareas en el hilo principal: cada una toca pocos archivos con lecturas acotadas, y T2–T4 cambian guía distribuida, cuya revisión de conducta completa (`AGENTS.md` «Guidance structure») conviene hacer en un solo hilo. T1 va primero para tener el mecanismo antes que el texto, como pide el watch `watch/card-format-after-compaction`.

### T1 — Criterio `card_option_format` del auditor de sesiones

- [x] `--audit-session` informa `card_option_format` en Claude, Codex y Grok, sobre cada tarjeta mostrada: falla una opción sin `<letra> - `, y en una de una sola elección una marca de recomendada que falta, se repite o no está en la primera; pasa una sesión con tarjetas en formato y da `not_observed` sin tarjetas. La evidencia da la línea y el número de la pregunta, sin su texto ni su `id`.

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `tests/pilot/session_audit_assessment.go` (junto a `cardAnswersOf`, reutilizando su lectura de `Questions[].Options[].Label` y `recommendedLabel`), `tests/pilot/session_audit.go` (`auditCriteria`), `tests/pilot/session_audit_test.go` (`TestAuditNewCriteria`), `tests/fixtures/sessions/card_option_format/` (nuevo), `_support/docs/harness-engineering/session-review-rubric.md` (la fila de Questions del 2026-10-31 que hoy no informa el marcador del host apunta al criterio nuevo), `_support/docs/harness-engineering/rule-enforcers.md` (fila nueva), `_support/openspec/changes/question-format-template/specs/workflow-lifecycle/spec.md`.

**Execution:** hilo principal.

**Test approach:** check — extiende la tabla de `TestAuditNewCriteria`, dueña de los criterios del auditor, y su guarda de filtración de texto. Casos con fixtures: Claude con etiquetas `Texto (★)` (fail), Claude con `A - (★) …` y preguntas de Entrega y Revisión (pass), Claude de opción múltiple con dos marcas y letras (pass), Claude con la marca en la segunda opción (fail), Claude sin ninguna marca (fail), Claude con `B - Estricto: solo A - (★)` (pass), Claude con una tarjeta rechazada (resultado de error) y sin letras (fail), Claude con una tarjeta sin resultado al final de la sesión y sin letras (fail), Claude con `Texto (Recomendado)` sin letra (fail), Codex `request_user_input_async` con opciones en texto suelto (not_observed), Codex `A · Siempre (Recommended)` (fail), Codex `A - Siempre (Recommended)` (pass), Grok con `options[].label` en formato (pass), sesión sin tarjeta (not_observed), sesión hija (not_observed). Al menos una fixture que falla pregunta «Which delivery…», para que la guarda de filtración confirme que la evidencia no lleva el texto.

**Changes:** agrega el criterio a la salida con una lectura propia de cada tarjeta mostrada (la de `cardAnswersOf` salta las que no tienen resultado); no cambia los demás criterios ni `card_answers`. La evidencia sale de `regressionEvidence` más el número de la pregunta en la tarjeta.

**Verification:** `go test -count=1 ./tests/pilot/ -run 'TestAudit'` pasa; `go run ./tests/pilot --audit-session ~/.claude/projects/-Users-me-Development-projects-tricell-tricell-hive/aa27a201-bc12-44e0-af29-ca245d12f8cc.jsonl --host claude` da `card_option_format` `fail` en las líneas 3849, 6616 y 6940; el mismo comando sobre `1140329b-708e-457b-83d3-7a550b97972b.jsonl`, `e8ef9855-1d4b-48f7-b8f5-2879ea4fdc14.jsonl` y `7bcc87ae-674d-42c9-b444-80a774e5661d.jsonl` da `pass`; con `--host codex` sobre `~/.codex/sessions/2026/11/20/rollout-2026-11-20T01-26-41-01a1f7d9-b225-7b12-bfa1-7fa2b09db920.jsonl` da `fail` y sobre `~/.codex/sessions/2026/11/18/rollout-2026-11-18T14-09-13-01a1f047-170e-7201-8341-51c978d9de47.jsonl` da `pass`.

### T2 — Plantillas de la ronda y de la tarjeta en `questions.md`

- [x] `questions.md` muestra la ronda en texto y la tarjeta como plantillas literales, con opciones `A - (★) <opción>` sin marcador de lista y `---` entre decisiones en el texto; la línea mínima de `global.md` usa el literal.

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/hive-core/references/questions.md` («Format», «Channel», la línea de IDs), `content/guidance/global.md` («Questions and routes», la frase de la decisión), `content/skills/hive-core/references/viewer-open.md` (líneas 7 y 9: la opción y la respuesta con letra).

**Execution:** hilo principal.

**Test approach:** check

**Changes:** las plantillas de `design.md` «Proposed shape» reemplazan la descripción en prosa de la forma; las reglas de contenido de cada parte quedan en una línea cada una, sin listas de lo que no se debe hacer (D1). La línea de IDs nombra el ID de los bloques de respuestas y apunta a «The next-step block» de `report-endings.md`, donde vive la regla de que solo sirve para citarlo (T3). No cambia qué se pregunta («What to ask»), los canales ni las preguntas de sesión. Antes de editar, revisar todo lo que gobierna esta conducta como pide `AGENTS.md` «Guidance structure» (global, hojas, roles, `site/`, `_support/docs/architecture/`, detectores de `tests/pilot/`) y listar los lugares en el PR, con el tamaño de `global.md`. Tamaño: `questions.md` bajo su tope de 7 000 B, y crecimiento neto cero de las tres hojas de lectura (`questions.md`, `report-endings.md`, `global.md`), por D8 A; si no cabe, detenerse con la cifra medida y preguntar. El mensaje de commit nombra las observaciones (Claude Code, `claude-opus-5-5`, sesiones `bf678880`, `7ee42122`, `aa27a201`, `2faf431f`, `00c7fa30`, 2026-11-19 a 2026-11-21; 604 mensajes con viñeta en 117 sesiones) y el watch `card-format-after-compaction`.

**Verification:** `go test -count=1 ./tests/content/` pasa (tamaño de hoja, tope de `global.md`, paquetes de etapa, enlaces); `go test -count=1 ./tooling/management/ -run 'Stage'` pasa; `grep -c '^```' content/skills/hive-core/references/questions.md` da al menos 4 (dos bloques).

### T3 — Plantillas de los bloques de respuestas en `report-endings.md`

- [x] `report-endings.md` muestra el bloque de análisis, el de siguiente paso y el pie de cierre como plantillas literales con `▶ D<n> — …` y respuestas con letra sin marcador de lista, y el auditor sigue aceptándolos.

**Closes:** AC5

**Depends on:** T2

**Locations:** `content/skills/hive-core/references/report-endings.md` (la fila de análisis, «The next-step block», «The close footer», con la línea 26 del encabezado `▶ Closed`), `tests/fixtures/sessions/analysis_reply_block/` (fixture nueva), `tests/pilot/session_audit_test.go`, `_support/docs/harness-engineering/session-review-rubric.md` (fila de «Reports» de la línea 46: el bloque lleva `▶ D<n>`).

**Execution:** hilo principal.

**Test approach:** check — una fixture de Claude cuyo análisis termina con `**▶ D7 — ¿Cómo seguimos?**` y respuestas `A - (★) …` sin viñeta: `analysis_reply_block` da `pass` y `decision_round_format` `not_observed`.

**Changes:** cambia solo la forma de los tres bloques y agrega su ID (D2, D6); el pie de cierre tiene su propia plantilla, sin comando, con la línea de espera y con `(★)` solo cuando recomienda una respuesta, y `▶ Closed` lleva ID cuando ofrece respuestas. «The next-step block» dice que el ID solo sirve para citar el bloque y que responderle conserva su efecto. No cambia cuándo se usa cada bloque ni sus respuestas. Conserva la frase que fija `tests/content/single_home_test.go:107` («is never asked with the question tool, because a card hides the report it closes») y las anclas `#the-next-step-block` y `#the-close-footer`. Tamaño: bajo el tope de 9 000 B y crecimiento neto cero junto con T2 (D8 A). Mensaje de commit con las mismas observaciones que T2.

**Verification:** `go test -count=1 ./tests/pilot/ -run 'TestAudit'` pasa con la fixture nueva; `go test -count=1 ./tests/content/` pasa.

### T4 — `AGENTS.md`, sitio y suite completa

- [x] `AGENTS.md` dice que un formato que ve el usuario se escribe como plantilla literal en su hoja y que una plantilla no es un ejemplo suelto; las páginas del sitio que describen los bloques de respuestas dicen que llevan un ID que se puede citar.

**Closes:** AC6, AC7

**Depends on:** T3

**Locations:** `AGENTS.md` («Guidance structure», junto a la línea de ejemplos; «Maintainer viewer restart», la respuesta *open the viewer* como respuesta con letra), `site/src/content/docs/flows/hive-mode.md` (línea 61), `site/src/content/docs/flows/close.md` (línea 28) y cada página de `site/` que describa el bloque de siguiente paso o el pie de cierre (buscar «next-step», «footer», «letter»).

**Execution:** hilo principal.

**Test approach:** check

**Changes:** describe la conducta y enlaza la hoja; no copia las plantillas.

**Verification:** `go test -count=1 ./tests/content/` pasa; `go vet ./...` y `go test -count=1 ./...` pasan (suite completa); `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./...`, `go test -count=1 ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Revisión de código | Candidato final, antes de la validación o del push | La respuesta de revisión | La respuesta de revisión |

Recorrido humano: leer las plantillas en `questions.md` y `report-endings.md`; tras el refresco local, la próxima ronda real que veas debe seguir la plantilla, y el auditor la mide.

## Checkpoint

- Updated: 2026-11-21
- Stage: verify · 2 reviews
- Last commit: 09cbbcc9
- Open findings: none
- Next: code review result, then push
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/question-format-template · feat/question-format-template
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- before captures: none (no UI)
- go vet ./... · exit 0 · 09cbbcc9
- go test -count=1 ./... · exit 0 · 09cbbcc9 (tree = 09cbbcc9 plus tasks.md)
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 347 OK, 121 s · 09cbbcc9
- CI-only: .github/workflows/master-source-branch.yml (source-branch policy on master PRs, no tests)
- full suite reused from build at 09cbbcc9 (tree unchanged but tasks.md)
- blast radius: candidate auditor vs base on 570 sessions · 12532 criteria identical; 8 diffs from 3 transcripts that grew, identical on rerun · 09cbbcc9
- AC1 aa27a201 card_option_format fail ×3 (3849, 6616, 6940) · 09cbbcc9
- AC2 1140329b/e8ef9855/7bcc87ae pass (13/3/18 cards) · 09cbbcc9
- AC3 codex 01a1f7d9 fail, 01a1f047 pass · 09cbbcc9
- AC4–AC7 leaves, AGENTS.md, site grep; TestStageManifests and footprint ok · 09cbbcc9

## Review status and progress

2026-11-21 — T1 · 08ea7ace · unverified: none
2026-11-21 — T2 · d6b223a5 · unverified: none
2026-11-21 — T3 · c8183d41 · unverified: none
2026-11-21 — T4 · 09cbbcc9 · unverified: none
