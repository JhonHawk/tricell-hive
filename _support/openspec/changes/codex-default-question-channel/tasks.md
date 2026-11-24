# Tasks

Base `development` en `15ac9aa5`, rama `fix/codex-default-question-channel`, worktree `.claude/worktrees/codex-question-channel`. Todas las tareas corren en el hilo principal: T1 a T3 tocan los mismos archivos del auditor y su tabla de tests, y T4 y T5 son ediciones de texto de pocas líneas. El auditor va primero, como pide `AGENTS.md` «Measurement» cuando el watch item nombra un mecanismo.

### T1 — `prose_choice_without_card` lee una decisión de sesión pendiente en prosa

- [ ] La línea 408 de la sesión `01a20c02` falla, y una frase que solo deja pendiente la entrega da `review`

**Closes:** AC1

**Depends on:** none

**Locations:** `tests/pilot/session_audit_prose.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/prose_choice_without_card/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** En `proseChoices`, un tercer caso de frase sin `?`. Si nombra una decisión de sesión (`sessionChoice`) y la deja pendiente («queda/falta por confirmar», «pendiente de confirmar», "still needs confirming", "awaiting your confirmation"), da `review`. Si recomienda («recomiendo», "I recommend") un texto entre backticks que contiene un modo de entrega (`interactive|automatic|ready|hold|direct-base`) como palabra entera, da `fail`; cada frase se juzga sola, y `` `threshold` `` o `` `already` `` no cuentan. Fixtures nuevas: `codex-rollout-pending-recommend-fail.jsonl` con el texto de L408, `codex-rollout-pending-only-review.jsonl` y `codex-rollout-mode-substring-notobserved.jsonl` (un `` `threshold` `` recomendado). Un informe en pasado («Registré la entrega automática») sigue sin contar, y las fixtures existentes conservan su estado.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; `go run ./tests/pilot --audit-session ~/.codex/sessions/2026/11/23/rollout-2026-11-23T23-23-31-01a20c02-5d3c-76b2-ad76-f221fd88dd04.jsonl --host codex` da `prose_choice_without_card` fail en la línea 408

### T2 — Criterio `async_card_when_waiting_offered`

- [ ] La línea 421 de la sesión `01a20c02` falla; una async sin la tarjeta que espera y con la ronda escrita pasa

**Closes:** AC2

**Depends on:** none

**Locations:** `tests/pilot/session_audit_async.go` (nuevo), `tests/pilot/session_audit.go` (`auditCriteria`), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/async_card_when_waiting_offered/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Por cada llamada a `request_user_input_async` de una sesión principal: `fail` si la sesión ofrece `request_user_input` (`QuestionOffered`, que se calcula para toda la sesión); `fail` si no lo ofrece y ningún texto anterior del mismo turno es una ronda (`isTextRound`); si no, `pass`. Sin llamadas async, `not_observed`; no juzga a un hijo. Tres fixtures: `codex-rollout-waiting-offered-fail.jsonl`, `codex-rollout-async-only-no-round-fail.jsonl`, `codex-rollout-async-only-round-pass.jsonl`. La evidencia nombra la línea, nunca el texto.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; el `--audit-session` de T1 sobre `01a20c02` da `async_card_when_waiting_offered` fail en la línea 421

### T3 — El `$flow-x` tecleado en Codex cuenta como carga, y la ruta al build exige el bloque `▶`

- [ ] La línea 496 de la sesión `01a20c02` falla en `plan_without_build_route`

**Closes:** AC3

**Depends on:** T1, T2 (misma tabla de tests)

**Locations:** `tests/pilot/session_audit.go` (`codexCompletedItem`), `tests/pilot/session_audit_routes.go` (`planWithoutBuildRoute`), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/plan_without_build_route/`

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- `codexCompletedItem` emite `Kind:"skill_invocation"` con el nombre de la skill cuando un ítem `UserMessage` trae en su contenido una entrada `skill`.
- `planWithoutBuildRoute`:
  - Aprueba una etapa en la que un mensaje trae el comando de `flow-build` o `create-design` junto con una línea `▶`.
  - Da `fail`, con la línea de ese mensaje, cuando el comando aparece solo sin `▶`, sin importar cómo termine la etapa.
  - Mantiene los casos actuales cuando ningún mensaje trae el comando.
- Fixtures nuevas: `codex-rollout-typed-skill-no-block-fail.jsonl` y `codex-rollout-typed-skill-block-pass.jsonl`.
- Las cuatro fixtures `pass` actuales tienen `▶` en el mensaje del comando y no cambian.
- No hay fixture con ítems `UserMessage`. Por eso, antes y después del cambio, hay que correr `--audit-session` sobre los diez rollouts Codex principales más recientes de `~/.codex/sessions/2026/11/` y anotar en el progreso cada criterio que cambie de estado, con su causa.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; el `--audit-session` de T1 sobre `01a20c02` da `plan_without_build_route` fail en la línea 496

### T4 — La regla y su descripción

- [ ] La fila del canal y la página del sitio dicen lo que pide AC4

**Closes:** AC4, AC5

**Depends on:** T1, T2, T3

**Locations:** `content/skills/hive-core/messages/questions.md` «Channel», `site/src/content/docs/host-limits.md` «Codex», `_support/docs/architecture/agent-delivery.md` (fila Codex de la línea 283, línea 292, nota fechada tras «Superseded on 2026-11-19»), `README.md:199`, `llms.txt:73`, `_support/docs/architecture/tool-dependencies.md:40`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **La regla.** La fila «No waiting card» (línea 62) nombra la tarjeta async solo «cuando el modo no ofrece `request_user_input`». La viñeta de la línea 64 se reescribe, sin agregar otra, para decir que Hive pregunta por la tarjeta la entrega, la revisión, la validación y toda decisión aunque la plantilla de Codex mande la entrada obligatoria a una frase de texto. Las «solicitudes de permiso» de esa plantilla quedan como las escaladas de sandbox o aprobación del host. La redacción va sin «never», porque no es un invariante.
- **El sitio.** Extiende la viñeta Codex (líneas 63-73) describiendo el comportamiento, sin copiar la regla.
- **`agent-delivery.md`.**
  - La fila Codex de la línea 283 agrega la exclusión de la tarjeta async.
  - La línea 292 queda marcada como superada.
  - Una nota fechada registra la observación `01a20c02` (Codex, `gpt-6.2-sol` high, 2026-11-24) y la decisión D4 A.
- **README, llms.txt y tool-dependencies.** Dejan de limitar la tarjeta a «las preguntas que Codex permite» y dicen que Hive la usa para cada ronda.
- **Paquetes de etapa.** Los que contienen la regla (`hive-mode`, `flow-plan`, la salida de `flow-verify` y `unattended-delegation`) quedan al menos 500 bytes bajo 24.000.

**Verification:** la comprobación de texto de `AGENTS.md` «Checks by path» pasa; `pnpm check`, `pnpm lint` y `pnpm build` en `site/` pasan

### T5 — Rúbrica, enforcers y spec

- [ ] Los criterios nuevos y cambiados tienen su fila y su requisito

**Closes:** AC1, AC2, AC3

**Depends on:** T1, T2, T3

**Locations:** `_support/docs/harness-engineering/session-review-rubric.md` (filas 37, 39 y 42), `_support/docs/harness-engineering/rule-enforcers.md` (filas 50 y 51 y una nueva), `_support/openspec/changes/codex-default-question-channel/specs/workflow-lifecycle/spec.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **Rúbrica.**
  - La fila 37 deja de mandar la entrada obligatoria de Codex Default a la plantilla nativa: solo las escaladas de permisos del host y las tarjetas opcionales sin respuesta siguen la plantilla. Esta fila suma la comprobación de la ronda async con `async_card_when_waiting_offered`, sin agregar otra fila.
  - La fila 42 nombra la recomendación de un modo entre backticks (fail) y la decisión que solo queda pendiente (review).
  - La fila 39 nombra el bloque `▶` y la carga tecleada en Codex.
  - Todas van fechadas 2026-11-24 con la sesión `01a20c02`.
- **`rule-enforcers.md`.**
  - La fila 50 describe el nuevo fail y el nuevo review.
  - La fila 51 suma el requisito de `▶`.
  - La fila nueva se escribe `` `tests/pilot/session_audit_async.go` (`async_card_when_waiting_offered`) ``, con su home en `questions.md`.
- **Spec.** El delta queda ajustado a lo construido.

**Verification:** la comprobación de texto de `AGENTS.md` «Checks by path» pasa

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | Al cerrar cada tarea, en el worktree | El comando de la tarea pasa | Cubierto por la implementación |
| Full suite | Candidata final, en el worktree | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Auditor sobre la sesión real | Candidata final | `--audit-session` sobre `01a20c02`: fail en 408, 421 y la carga de `flow-plan`; L200 y L236 siguen en fail | Cubierto por la implementación |
| Code review | Candidata final, antes del push | `/code-review` de Claude Code | La respuesta de revisión |
| Regla instalada (AC4) | Tras el merge y `hive update` | `rg -n 'request_user_input' ~/.agents/skills/hive-core/messages/questions.md` muestra la condición de la fila async y la línea de la plantilla | La entrega `automatic` y la actualización de mantenedor |

Sin recorrido humano: la entrega es `automatic`, y el efecto de la regla en sesiones reales de Codex se mide después, en el watch `watch/codex-decision-format-async`.

## Checkpoint

- Updated: 2026-11-24
- Stage: flow-build · 2 · T1–T3 (hive-build-backend) y T4–T5 (hive-write-docs) en paralelo
- Last commit: 15ac9aa5
- Open findings: none
- Next: integrar las dos salidas: go test ./tests/pilot/, comprobación de texto, checks del sitio; luego commits por tarea
- Worktree: .claude/worktrees/codex-question-channel · fix/codex-default-question-channel
- Rules: flow-build/stage.md · extra: child-brief
- Principles: none

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-24 — Plan escrito.
