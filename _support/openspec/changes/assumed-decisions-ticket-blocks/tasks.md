# Tasks

Base: `development` en `a0e9b280`, rama `feat/assumed-decisions-ticket-blocks` en un worktree de `.claude/worktrees/`. T1 y T2 editan más de 3 archivos del auditor con una interfaz ya fijada (`criterionAssessment` y el formato de fixtures), así que van a un hijo `hive-build-backend`, uno tras otro porque comparten `session_audit.go` y `session_audit_test.go`. T3 a T7 editan como máximo 3 archivos cada una y quedan en el hilo principal, que corre a la vez que el hijo porque sus archivos no se cruzan.

### T1 — Criterio `ticket_assessment_ending` del auditor

- [x] El auditor marca `fail` una evaluación de ticket que cierra el turno sin tarjeta, bloque ▶ ni comando de flujo; `pass` cuando cierra con tarjeta, con el bloque de siguiente paso o con el bloque de respuestas ▶; `review` cuando cierra con una pregunta abierta en texto; y `not_observed` sin evaluación de ticket.

**Closes:** AC1, AC2

**Depends on:** ninguna

**Locations:** `tests/pilot/session_audit_assessment.go` (nuevo), `tests/pilot/session_audit.go` (registro), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/ticket_assessment_ending/`

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd

**Changes:** Una evaluación de ticket es un turno cuyo mensaje del usuario nombra una clave de ticket (Jira o Linear `[A-Z]{2,}[A-Z0-9]*-\d+`, sin `UTF-`, `SHA-`, `ISO-` ni otros prefijos de estándar; o un issue de GitHub `#\d+` junto a «issue» o «ticket») y en el que el agente carga `hive-mode` (skill o lectura de `hive-mode/SKILL.md`, también dentro de un `cat` agrupado de Codex), con un helper nuevo porque `readsFlowSkill` solo reconoce flujos. Se excluyen los turnos que cargan `hive-mode/playbooks/ticket-status.md` y los que siguen a un comando de flujo. Reusa `nextStepBlock`, `flowCommand` y la detección de tarjetas. Grok queda como `not_observed` si su traza no expone la carga de la skill. Fixtures: Codex `fail` (forma de `01a1f5d4` a las 22:04Z), Codex `pass` con bloque ▶ y comando, Claude `pass` con tarjeta, Claude `pass` con «▶ ¿Cómo seguimos?», `review` con pregunta abierta, `not_observed` con una clave en un turno sin `hive-mode` y `not_observed` en un turno de `ticket-status`.

**Verification:** `go -C <worktree> test ./tests/pilot/...` pasa con cada fixture visto fallar antes de la implementación; `go -C <worktree> run ./tests/pilot --audit-session ~/.codex/sessions/2026/11/19/rollout-2026-11-19T16-01-30-01a1f5d4-3ec7-7631-97d8-ba540210ac98.jsonl --host codex` da `fail` con la línea del mensaje de las 22:04:22Z; sobre la sesión Claude `bf678880` da `not_observed` o `pass`, nunca `fail`.

### T2 — Dato informativo `card_answers`

- [x] La salida del auditor trae, por sesión y fuera de `criteria`, cuántas tarjetas se contestaron con una opción y cuántas eligieron la recomendada.

**Closes:** AC3

**Depends on:** T1

**Locations:** `tests/pilot/session_audit.go`, `tests/pilot/session_audit_assessment.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/card_answers/`

**Execution:** hijo `hive-build-backend` (el mismo de T1, después de T1)

**Test approach:** tdd

**Changes:** Campo `card_answers: {answered, recommended}` en la salida JSON. La recomendada es la opción cuya etiqueta lleva `(★)`, «(Recommended)» o «(recomendado)». Cuenta las decisiones y las preguntas de sesión; no cuenta tarjetas vacías (`{"answers":{}}`), respuestas de texto libre ni opciones múltiples. Claude: `AskUserQuestion` y su resultado en sus dos formas («User has answered…» y «Your questions have been answered…»); Codex: `request_user_input` y su respuesta. Grok, sin tarjeta parseable, da `answered: 0`.

**Verification:** `go -C <worktree> test ./tests/pilot/...` pasa con fixtures de Claude (las dos formas), Codex con «(Recommended)» y Codex vacío; sobre `7ee42122` reporta `answered: 4, recommended: 3`; sobre `01a1f5d4` reporta números que coinciden con un conteo a mano.

### T3 — Regla de supuestos y criterios a–f

- [x] `hive-rule-questions` «What to ask» define a–f con sus umbrales, a qué rondas no se aplican, la lista de supuestos y su aceptación; `global.md` y `plan-handoff.md` ya no hablan de «minor calls».

**Closes:** AC4

**Depends on:** ninguna

**Locations:** `content/skills/hive-rule-questions/SKILL.md`, `content/guidance/global.md` (línea 131), `content/skills/flow-plan/references/plan-handoff.md` (punto 3)

**Execution:** hilo principal

**Test approach:** check

**Changes:** Sustituye «minor call» por: decidir la opción que gana con claridad (umbral de `proposal.md`) y listarla, como máximo 5 líneas con la decisión, lo descartado y cómo se revierte, y el resto en una sola pregunta; preguntar solo lo que cumple a–f; la respuesta a la tarjeta o al bloque acepta cada supuesto que no corrige, sin volverlo una decisión ni una autorización. Nombra las rondas que conservan su pregunta. Ajusta la línea 33 («a decision a reply leaves out stays open») para que no choque con los supuestos. `global.md` dice «decide y lista lo que gana con claridad; pregunta lo que cumple un criterio de [questions] «What to ask»», sin enumerar a–f a medias. El PR da el cambio de tamaño de `global.md`.

**Verification:** `rg -n 'minor call' <worktree>/content <worktree>/site` solo devuelve `hive-rule-ui-example` hasta T4; `rg -c '^\s*- \*\*[a-f]\.' <worktree>/content/skills/hive-rule-questions/SKILL.md` da 6; `go -C <worktree> test ./tests/content/...` pasa.

### T4 — Prueba de `assumption` de las reglas de negocio

- [x] La prueba de `assumption` remite a a–f, y las hojas que la citan dejan de decir «minor call».

**Closes:** AC4, AC5

**Depends on:** T3

**Locations:** `content/skills/hive-rule-business-record/SKILL.md` (línea 18), `content/skills/hive-rule-business-rules/SKILL.md` (línea 22, si nombra la prueba vieja), `content/skills/hive-rule-ui-example/SKILL.md` (línea 11)

**Execution:** hilo principal

**Test approach:** check

**Changes:** «`assumption`: adoptada sin preguntar, solo cuando no se cumple ninguno de los criterios a–f de [questions] «What to ask»». La regla aceptada con la respuesta sigue como `assumption` hasta que el usuario la contesta.

**Verification:** `rg -n 'minor call' <worktree>/content <worktree>/site` sin resultados; `rg -n 'a–f|a-f' <worktree>/content/skills/hive-rule-business-record/SKILL.md` da la línea de la prueba; `go -C <worktree> test ./tests/content/...` pasa.

### T5 — Figma y plantilla del plan

- [x] La aprobación de Figma, su página y la plantilla `proposal.md` aplican la misma prueba y separan «Supuesto» (decisión) de la línea `assumption` (regla de negocio).

**Closes:** AC5

**Depends on:** T4

**Locations:** `content/skills/flow-figma-design/references/example-record.md`, `site/src/content/docs/flows/figma-design.md` (línea 31), `content/skills/flow-plan/assets/change-folder/proposal.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** `example-record.md` remite a la prueba de la hoja y conserva su definición de `scope`. La página de Figma describe la prueba sin copiarla. El comentario de `## Decisions` de la plantilla dice que un supuesto se escribe con la palabra del idioma de la sesión para «supuesto» y lo descartado, y que una regla de negocio supuesta va además como línea `assumption` en `## Business rules`.

**Verification:** `python3 -I <worktree>/content/skills/flow-plan/scripts/check_change.py <esta carpeta>` sale con 0; `go -C <worktree> test ./tests/content/...` pasa.

### T6 — Evaluación de un ticket por bloques

- [x] `hive-mode` «Assess a chosen ticket» fija los bloques, el veredicto con los números de la lista que fallan, el orden y «el plan decide el cómo, nunca qué entra»; las páginas del sitio lo describen.

**Closes:** AC6

**Depends on:** T3

**Locations:** `content/skills/hive-mode/SKILL.md`, `site/src/content/docs/flows/hive-mode.md` (línea 51), `site/src/content/docs/flows/plan.md` (línea 61, punto 3 del resumen)

**Execution:** hilo principal

**Test approach:** check

**Changes:** Recupera de `c423e5ef` las líneas con etiqueta y el orden explícito, con los bloques de `design.md`. El paso 5 queda sin ambigüedad: la ronda va antes del bloque de siguiente paso; el bloque va en el mismo mensaje cuando no hay nada que preguntar; si los hallazgos no confirman trabajo, el cierre es el de `report-endings`. «No listo» es fallar algún punto de la lista de `ticket-types.md`. Las páginas describen el comportamiento y enlazan a la skill.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (tamaño de `hive-mode` ≤ 16,000 bytes); lectura cruzada de que `hive-mode`, `hive-rule-questions`, `report-endings` y `next-step-block` dan un solo orden para el cierre de una evaluación.

### T7 — Rúbrica, enforcers y una sola casa

- [x] La rúbrica, `rule-enforcers.md` y `single_home_test.go` reflejan los dos comportamientos y el criterio nuevo.

**Closes:** AC1, AC4

**Depends on:** T1, T3

**Locations:** `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Dos filas de rúbrica con fecha y la sesión Codex `01a1f5d4`; una fila en `rule-enforcers.md` para `ticket_assessment_ending`, y la fila «Ask only for a choice that is the user's» pasa a nombrar la hoja; un `homePhrase` con una frase literal de la regla de supuestos y `hive-rule-questions` como casa.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (incluye `rule_enforcers_test.go` y `TestRulesHaveOneHome`).

### T8 — Changelog

- [x] `CHANGELOG.md` tiene la entrada del cambio bajo `[Unreleased]`.

**Closes:** AC4, AC6

**Depends on:** T3, T6

**Locations:** `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Una entrada que nombra la lista de supuestos, los criterios para preguntar, la evaluación de tickets por bloques y el criterio nuevo del auditor.

**Verification:** `python3 -I <worktree>/content/skills/flow-plan/scripts/check_change.py <esta carpeta>` sale con 0; `go -C <worktree> test ./...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...` | Covered by the implementation |
| Auditor over real sessions | Final candidate, locally | `--audit-session` sobre `01a1f5d4`, `bf678880` y `7ee42122` con los resultados de AC1–AC3 | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` de Claude Code | Delivery `automatic` (2026-11-19) |

Sin parada de validación (entrega `automatic`). Queda abierto a propósito tras el merge y el despliegue: AC4, AC5 y AC6 en la primera sesión real que decide, supone una regla o evalúa un ticket.

## Checkpoint

- Updated: 2026-11-19
- Stage: verify · push
- Last commit: 6f39a06
- Open findings: none
- Next: Push, PR to development, merge gate, merge, then flow-close and local refresh
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/assumed-decisions · feat/assumed-decisions-ticket-blocks

### Evidence

- before captures: none (no screen change)
- go -C <wt> test ./... exit 0 @6cb6ada
- go -C <wt> vet ./... exit 0; gofmt -l clean @6cb6ada
- local CI: the only workflow (master-source-branch) runs on PRs to master; CI-only, not applicable to development
- audit 01a1f5d4 ticket_assessment_ending=fail L64, card_answers 4/3; bf678880 not_observed 49/43; 7ee42122 not_observed 4/3
- go -C <wt> test ./... exit 0 and vet exit 0 @0584c28
- go -C <wt> test ./... exit 0 @6f39a06 (flaky TestE2ESetupDeadlinesOwnDescendantsAndPipes fails 1/5 on base a0e9b280 too)

## Review status and progress

2026-11-19 — T3 · f74ff1c · unverified: AC4 en una sesión real tras el despliegue
2026-11-19 — T4 · 854408f9 · unverified: AC5 en una sesión real tras el despliegue
2026-11-19 — T5 · 8ac65c95 · unverified: AC5 en una sesión real de Figma tras el despliegue
2026-11-19 — T6 · 4892cc99 · unverified: AC6 en la primera evaluación real de un ticket tras el despliegue
2026-11-19 — T1 · 4deeae30 · unverified: none
2026-11-19 — T2 · 4deeae30 · unverified: none
2026-11-19 — T7 · 6cb6ada7 · unverified: none
2026-11-19 — T8 · ac73730 · unverified: none
