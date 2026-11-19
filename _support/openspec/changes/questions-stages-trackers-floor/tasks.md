# Tasks

Base `development` en `262b64e2`. Cada unidad va en su rama y su PR, en orden E1 → E4, porque E2, E3 y E4 tocan `global.md`. Las tareas corren en el hilo principal: son reescrituras de reglas que dependen de leer el conjunto, y sus escrituras se solapan en `global.md` y en los enlaces. La documentación del sitio de cada unidad puede ir a un hijo `hive-write-docs` cuando pase de 3 archivos.

### T1 — Criterios del auditor para la sesión del 2026-11-19

- [x] `--audit-session` detecta la entrada a un flujo sin bloque, la ronda de decisiones en texto sin formato ni tarjeta, y un cambio de interfaz verificado sin `hive-review-ux`.

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `tests/pilot/session_audit.go` y sus pruebas, `tests/fixtures/sessions/<criterio>/` (fixtures mínimos derivados del rollout 01a1f2f5, sin nombres ni claves del proyecto), `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** main thread

**Test approach:** tdd — cada fixture falla con el auditor actual antes del cambio

**Changes:** un criterio nuevo de entrada a flujo sin ruta, con el disparador observable de AC1; `text_round_without_card` reconoce rondas `D<n>` en viñetas sin `Situation`/`(★)`; un criterio nuevo de cambio de UI sin revisor UX, con el disparador de AC3. Antes de escribirlos, confirmar que el parser ve las lecturas y ediciones hechas dentro de las llamadas `exec` de Codex; si no, ampliarlo en esta tarea. No cambia los criterios existentes que ya pasan. Cada criterio nuevo queda en `rule-enforcers.md`; los mensajes de commit nombran la observación (Codex 0.170.0, `gpt-6.2-sol`, rollout `01a1f2f5`, 2026-11-19).

**Verification:** `go -C <repo> test ./tests/pilot` (solo) y `go -C <repo> run ./tests/pilot --audit-session <rollout 01a1f2f5> --host codex` dan `fail` en los tres criterios.

### T2 — Hoja única de preguntas

- [x] `hive-rule-questions` reemplaza las cuatro hojas de preguntas, con el canal por host incluido Codex Default por `request_user_input`.

**Closes:** AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/hive-rule-questions/SKILL.md` (nueva, excepción nombrada ≤7,000), borrar `content/skills/hive-rule-{decision-format,question-channel,question-text,card-unanswered}/`, todos los archivos de `content/` que los enlazan, `content/guidance/global.md` («Questions and routes», su línea «each decision has an ID», y el índice), `tests/content/leaf_skill_test.go` (excepción), `leaf_links_test.go`, `single_home_test.go`, `reading_footprint_test.go`, `tests/pilot/session_audit.go` (`leafQuestionRound` acepta la hoja nueva y la anterior) con los fixtures de `leaf_read_before_action`, comentarios de `tests/pilot/regression.go` y `regression_test.go`, `_support/docs/harness-engineering/{rule-enforcers,session-review-rubric}.md`, `_support/docs/architecture/instruction-resources.md`, `site/`, delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** check — pruebas de contenido y el criterio de AC2

**Changes:** primero leer la documentación oficial vigente de Codex sobre `request_user_input` en modo Default y anotar el resultado en el contexto verificado; luego una sola hoja con formato, canal por host, texto y tarjeta sin respuesta, recortada sin perder intención; la tarjeta que espera lleva toda ronda de decisiones también en Codex Default. Las frases de hogar único de las hojas borradas pasan a la hoja nueva. No cambia el bloque de siguiente paso (T3).

**Verification:** `go -C <repo> test ./tests/content`; `rg` sin referencias a las cuatro hojas; tamaño de la hoja ≤7,000 y de `global.md` ≤32,000.

### T3 — Paso entre etapas, preguntas de sesión y revisión adversarial

- [x] Ninguna respuesta a una pregunta arranca un flujo; las preguntas de sesión van sin ID; la revisión no se recomienda «sin revisión» salvo en mecánico; `interrogate` deja elegir el proveedor.

**Closes:** AC6, AC7, AC8

**Depends on:** T2

**Locations:** `content/skills/hive-mode/SKILL.md`, `hive-rule-next-step-block`, `hive-rule-report-endings`, `hive-rule-delivery-ask` (reescritura ≤4,000, conservando su frase de hogar único), `hive-rule-interrogate` (conservando la suya), `flow-verify` (pregunta adversarial), `content/guidance/global.md` («Delivery»), `tests/pilot/session_audit.go` (`roundTitleLine` o un reconocedor nuevo de los encabezados fijos, `decisionBlocks`, `reviewQuestionNativeOption`) con un fixture de ronda en texto con encabezados fijos que debe dar `fail`, `site/`, delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** check — pruebas de contenido y los criterios de AC1 y `review_question_native_option`

**Changes:** la condición de continuación directa exige que la petición nombre el cambio y su destino; encabezados fijos para entrega, revisión y adversarial; la pregunta de revisión elige mecanismo y solo ofrece «sin revisión» en mecánico; disponibilidad de `interrogate` como lista de agentes instalados y con sesión.

**Verification:** `go -C <repo> test ./tests/content ./tests/pilot` (pilot solo); fixtures de `review_question_native_option` en verde con los nombres nuevos.

### T4 — Hogar de trackers y caché de estados

- [x] `hive-rule-trackers` con referencias por servicio; transiciones leídas antes de mover; caché `tracker/states/<tracker>/<clave>` en Engram.

**Closes:** AC9, AC10

**Depends on:** T3

**Locations:** `content/skills/hive-rule-trackers/` (hoja ≤4,000 y `references/{jira,linear,github}.md`), `content/skills/create-issue/references/trackers.md` (conserva la creación y enlaza las referencias por servicio), `hive-rule-ticket-state` (enlaza la hoja de trackers para el cómo; la línea del índice sigue siendo una), `content/guidance/global.md` («Persistent memory»), pruebas de contenido, `site/`, delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** check

**Changes:** reglas cortas de movimiento: leer las transiciones en vivo antes de mover, usar la caché `tracker/states/<tracker>/<prefijo>` (nombres y categorías de estado) para elegir, tomar la equivalente ante una falla; los comandos por servicio viven en sus referencias; un momento nuevo de escritura en memoria solo para esa caché. Orden: `hive-rule-ticket-state` decide cuándo y enlaza a `hive-rule-trackers` para el cómo. No cambia cuándo se mueve un ticket.

**Verification:** `go -C <repo> test ./tests/content`; `create-issue` paso 8 sigue leyendo su referencia de creación.

### T5 — Evidencia en el ticket antes de limpiar

- [x] El cierre documenta el ticket con sus evidencias antes de cualquier limpieza.

**Closes:** AC11

**Depends on:** T4

**Locations:** `content/skills/flow-close/SKILL.md`, `hive-rule-close-modes`, `hive-rule-trackers`, `content/guidance/global.md` («Scope and authorization»), `site/src/content/docs/flows/close.md`

**Execution:** main thread

**Test approach:** check

**Changes:** un paso de evidencia (capturas, salida de pruebas, enlaces) en el ticket antes de los pasos que borran material; la autorización de ese comentario en `global.md`; cada tracker adjunta según su referencia; ningún adjunto lleva valores de secretos ni identificadores de cuentas, y una captura que los muestre se recorta o se omite. No cambia qué material se borra.

**Verification:** `go -C <repo> test ./tests/content`; lectura del orden de pasos de `flow-close`.

### T6 — Hoja del piso de pruebas

- [x] `hive-rule-verification-floor` fija lo que siempre corre por tipo de cambio y la enlazan plan, build, verify y los playbooks.

**Closes:** AC12

**Depends on:** T5

**Locations:** `content/skills/hive-rule-verification-floor/SKILL.md` (nueva, ≤4,000; si no cabe, segunda excepción nombrada ≤7,000 en `leaf_skill_test.go`; define también la salida fallando y pasando y el bloque «Pruébalo tú»), `flow-plan`, `flow-build`, `flow-verify`, `content/skills/hive-mode/playbooks/*.md`, `hive-rule-security-boundaries`, `content/agents/review/hive-review-security.md`, `hive-review-code.md`, `hive-mode/playbooks/dependency-security.md`, `hive-rule-tests` (reescritura, 3 bytes libres), `hive-principle-explain-the-number` (enlace desde `perf.md`), `content/guidance/global.md` (índice), `tests/content/leaf_links_test.go`, delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** check

**Changes:** una tabla que apunta a los hogares existentes sin repetir sus frases (`single_home_test.go`); seguridad recomendada en cambios de riesgo; `hive-verify-task` en tareas de riesgo. El cambio mecánico sigue sin verify ni revisión.

**Verification:** `go -C <repo> test ./tests/content`.

### T7 — Evidencia y «Pruébalo tú» en el reporte

- [x] El reporte final lleva la salida fallando y pasando y el bloque «Pruébalo tú» en todo modo.

**Closes:** AC13

**Depends on:** T6

**Locations:** `hive-rule-close-report` (451 bytes libres: una línea que enlaza el piso, conservando su frase de hogar único), `hive-rule-close-modes` (fila de `hold`), `hive-rule-reasons-and-blockers`, `flow-verify`

**Execution:** main thread

**Test approach:** check

**Changes:** dos partes nuevas del reporte final; la pausa de `hold` las incluye. No agrega preguntas.

**Verification:** `go -C <repo> test ./tests/content`.

### T8 — README, llms.txt y sitio

- [x] La documentación para personas describe la revisión y la seguridad nuevas.

**Closes:** AC14

**Depends on:** T6, T7

**Locations:** `README.md` (L390, L440-441), `llms.txt` (L39), `site/src/content/docs/index.mdx` (L61), `site/src/content/docs/flows/{build,verify,plan,security}.md`, `content/skills/unattended-delegation/SKILL.md`, `content/skills/hive-mode/references/code-audit.md`

**Execution:** `hive-write-docs` cuando pase de 3 archivos; si no, main thread

**Test approach:** check

**Changes:** texto alineado con el piso; las páginas describen y enlazan, no copian reglas.

**Verification:** build del sitio desde la copia principal (`npm run build` en `site/`) y `rg -U -i 'only when (you|the user)\s+ask'` sin resultados en `README.md`, `llms.txt`, `site/` y `content/`.

### T9 — Observación tras el despliegue

- [ ] Los criterios de sesión real quedan observados o registrados como pendientes.

**Closes:** AC5, AC6, AC9, AC10, AC11, AC13

**Depends on:** T8

**Locations:** ninguna en el repositorio; Engram `pending/questions-stages-trackers-floor`

**Execution:** main thread

**Test approach:** check

**Changes:** tras cada merge se refresca la instalación (marcador `maintainer.local`); al terminar E4 se auditan con `--audit-session` las primeras sesiones reales de Codex y Claude Code, se marcan las casillas `live` de cada unidad y se registra en Engram lo que siga sin observar.

**Verification:** salida de `--audit-session` por sesión con los criterios de AC1, AC2 y AC3 en `pass`, y la caché de Engram de la transición observada.

## Delivery units

### E1 — PR criterios del auditor · depends on: none · files: tests/pilot/, tests/fixtures/sessions/, _support/docs/

- [x] unit: go -C <repo> test ./tests/pilot
- [x] live: auditar el rollout 01a1f2f5 · pass: los tres criterios nuevos o ajustados dan fail
- [x] perf: n/a: el auditor no tiene ruta de rendimiento que cambie

### E2 — PR preguntas y etapas · depends on: E1 · files: content/skills/hive-rule-questions/, content/skills/hive-mode/, content/guidance/global.md, tests/

- [x] unit: go -C <repo> test ./tests/content ./tests/pilot
- [ ] live: auditar la primera sesión real de Codex y de Claude Code tras el despliegue · pass: AC1 y AC2 en pass y una ronda por tarjeta que espera
- [x] perf: tamaño de global.md · base: 30,800 bytes · fail: más de 32,000 bytes

### E3 — PR trackers y evidencia · depends on: E2 · files: content/skills/hive-rule-trackers/, content/skills/flow-close/, content/guidance/global.md

- [x] unit: go -C <repo> test ./tests/content
- [ ] live: una sesión real que mueva un ticket de Jira con estados en español · pass: lee transiciones, mueve a «En curso» y deja la caché en Engram
- [x] perf: n/a: sin ruta de rendimiento

### E4 — PR piso de pruebas · depends on: E3 · files: content/skills/hive-rule-verification-floor/, content/skills/flow-*, README.md, site/

- [x] unit: go -C <repo> test ./tests/content && python3 -m unittest discover -s tests/skills -p '*_test.py'
- [ ] live: auditar una entrega hold real de UI tras el despliegue · pass: AC3 en pass, el reporte trae la salida fallando y pasando y «Pruébalo tú»
- [x] perf: tamaño de global.md · base: 30,800 bytes · fail: más de 32,000 bytes

Las casillas `live` se marcan en T9, después de los merges; ninguna unidad espera la casilla `live` de la anterior.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Each unit's candidate, locally | `go test` en todos los paquetes (`tests/pilot` solo) y `tests/skills` | Covered by the implementation |
| Code review | Each unit, before the validation stop | Claude Code `/code-review` | The review answer |
| Adversarial review | Each unit, after the code review | `interrogate` con el proveedor elegido | La respuesta «sí» y la elección de proveedor |

En cada parada de validación: el diff de la unidad, el tamaño de `global.md`, la salida del auditor sobre el rollout 01a1f2f5 y los lugares revisados que pide `AGENTS.md`.

## Checkpoint

- Updated: 2026-11-19
- Stage: flow-verify · E3+E4 joint validation stop
- Last commit: 0e3fa537
- Open findings: none
- Next: user validates E3+E4 together; then merge E3 then E4, deploy, close
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/questions-floor · feat/tracker-home

### Evidence

- before captures: none (no interface change)
- go test -count=1 ./tests/pilot · pass · db6a98bd
- go test ./tests/content ./tooling/... ./integrations/... · pass · db6a98bd
- audit rollout 01a1f2f5 · 3 new criteria fail · db6a98bd
- go test full suite (pilot alone) · pass · E2
- global.md 30,888 bytes (≤32,000)
- full suite E3+E4 · pass
- site build · pass
- global.md 31,444 bytes (≤32,000)

## Review status and progress

2026-11-19 — Plan escrito.
2026-11-19 — T1 · 5b9b9b3a · unverified: none
2026-11-19 — T2 · 4e9a8fe1 · unverified: AC5 (Codex round through request_user_input) waits for a real Codex session after deploy
2026-11-19 — T3 · 2cbc0706 · unverified: AC6 real-session half waits for a session after deploy
2026-11-19 — T4 · 84f3ea4c · unverified: AC9-AC10 wait for a real session that moves a ticket
2026-11-19 — T5 · 84f3ea4c · unverified: AC11 waits for a real close
2026-11-19 — T6 · 6d177c29 · unverified: AC13 waits for a real hold delivery after deploy
2026-11-19 — T7 · 6d177c29 · unverified: AC13 waits for a real hold delivery after deploy
2026-11-19 — T8 · 57a7fbdb · unverified: none
