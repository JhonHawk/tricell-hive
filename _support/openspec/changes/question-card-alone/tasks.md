# Tasks

Historical base: `development` at `c35c3424` (2026-11-16). Historical trial branch: `feat/question-card-alone`; its previous delivery and deployment approvals are not current-session authority. T9 is a local correction approved on 2026-11-17 in the isolated worktree at detached `af4093ec`. The user subsequently authorized native Codex review, commit/push to the trial branch and local deployment, without merge. T6 remains pending later real-session evidence. Model pilots remain paused.

Reopened tasks retain their historical implementation and evidence below. T9 revalidates the changed local text of T2, T4 and T8; installation checks in T5/T8 passed for T9; T6 remains pending. Existing T1/T3/T7 evidence is not proof of the new behavior.

### T1 — Los scripts de medición y la línea base quedan guardados

Primera tarea del build: su origen es una carpeta temporal que un reinicio de la máquina puede borrar.

- [x] `_support/workspace/2026-11-16-question-card-alone/measure/` contiene los scripts de Claude Code, Codex y los otros hosts con sus salidas del 9 al 16 de noviembre, y un `README.md` con el comando de cada uno

**Closes:** AC11

**Depends on:** none

**Locations:** `_support/workspace/2026-11-16-question-card-alone/measure/` (ignorado por Git); origen: la carpeta temporal de la sesión `5aa0de41`, subcarpetas `cc/`, `codex/` (sin el clon del código de Codex) y `others/`

**Execution:** main thread

**Test approach:** check

**Changes:** copia los scripts y sus salidas; añade a cada extractor la fecha de corte como parámetro y la clasificación que T6 necesita: ronda con o sin bloque de decisión antes de la tarjeta, tarjeta sin contestar, y si el mensaje siguiente del agente escribe las decisiones abiertas. No cambia las cifras de la línea base.

**Verification:** `python3 extract.py && python3 aggregate.py` en `measure/cc/` reproduce 306 rondas, 330 tarjetas y 252 rondas con bloque de decisión antes de la tarjeta; `python3 cards.py && python3 agg.py` en `measure/codex/` reproduce 47 tarjetas con 10 vacías. AC11 queda sin verificar hasta T6.

### T2 — La guía global pide la tarjeta sola y el texto tras una tarjeta sin respuesta

- [x] Los tres puntos de «Questions and routes» dicen la regla de `design.md` «Proposed shape», sin restos del orden anterior

**Closes:** AC9

**Depends on:** none

**Locations:** `content/guidance/global.md` (puntos «Write the message as visible text first», «Write the card to stand alone» y «Ask in the text only»)

**Execution:** main thread

**Test approach:** check

**Historical changes (2026-11-16):** reescribe los tres puntos con la condición observable (herramienta que espera y descripción por opción), el contenido del texto previo (línea de contexto e informe), la respuesta a una tarjeta sin contestar y al turno cortado, y los casos que siguen en texto. Antes de editar, revisa cada lugar que depende de la conducta, como pide `AGENTS.md` «Guidance structure», y anota los lugares revisados y el cambio de tamaño de `global.md` para el pull request. No cambia el formato de una decisión, las tarjetas consecutivas, las preguntas abiertas ni el bloque de siguiente paso.

**Historical verification (2026-11-16; superseded by T9 for local text and superseded deployment evidence):** `go test ./tests/...` pasa; `rg -n "visible text first|without repeating the decisions|round's text before it" content/guidance/global.md` no devuelve nada, y `rg -c "through that tool alone when it also gives each option a description field|When a card comes back unanswered and it carried the round alone" content/guidance/global.md` da 2; el tamaño de `global.md` antes y después queda anotado. AC9 queda sin verificar hasta T6.

### T3 — Las skills que citan el orden apuntan a la regla nueva

- [x] `flow-plan` paso 4, `flow-verify` paso 2 y la referencia del visor dejan de pedir el texto antes de la tarjeta

**Closes:** AC9

**Depends on:** T2

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-verify/references/viewer-publication.md` (el punto «In the card» y la lista de reglas que la referencia dice cambiar, que hoy incluye «a round's text is written in full before its card»)

**Execution:** main thread

**Test approach:** check

**Changes:** cada mención pasa a ser un puntero a «Questions and routes» sin repetir sus pasos; la opción «Open the viewer» se describe como opción de la tarjeta y como última línea de opción cuando la ronda se escribe en texto. No cambia qué se pregunta ni cuándo.

**Verification:** `rg -n -i "visible text before the question tool's card|written in the visible text|written in full before its card" content` no devuelve nada; `go test ./tests/content/...` pasa (tamaño de skills y referencias). AC9 queda sin verificar hasta T6.

### T4 — El sitio, la arquitectura y la rúbrica describen el orden nuevo

- [x] La página de límites por host, `agent-delivery.md` y la rúbrica de revisión de sesiones coinciden con la regla de T2

**Closes:** AC12

**Depends on:** T2

**Locations:** `site/src/content/docs/host-limits.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/session-review-rubric.md`

**Execution:** main thread

**Test approach:** check

**Historical changes (2026-11-16):** en la página, Codex y el paso 3 de la lista final dicen que tras un cierre el agente escribe las decisiones en el chat; en la arquitectura, la regla del 2026-11-04 pasa a historia con las sesiones y cifras que la reemplazan, y la tabla por host cambia «Full options in the text and in the card» por el reparto nuevo, con Codex 0.169.0 revisado; en la rúbrica, la fila de la tarjeta vacía comprueba el texto completo después y se añade la fila de la ronda sin bloque de decisión antes de la tarjeta. `README.md` y `llms.txt` no cambian: no afirman este orden.

**Historical verification (2026-11-16; superseded by T9 for local text and superseded deployment evidence):** `rg -n -i "already in the message above it|also written in the message|full options in the text and in the card|writes each decision's title line and each option line in the visible text|gets one line, not the round repeated" site _support/docs` no devuelve nada fuera de los párrafos marcados como historia; `npm run build` en `site/` termina sin errores y la página `host-limits` construida muestra el texto nuevo.

### T5 — La rama de prueba queda desplegada en la instalación local

- [x] Los archivos de guía instalados de cada host registrado contienen la regla de T2

**Closes:** AC4

**Depends on:** T2, T3, T4, y T7 integrada por el usuario en `development`

**Locations:** la instalación local de Hive (los siete hosts registrados)

**Execution:** main thread

**Test approach:** check

**Historical changes (2026-11-16):** con los commits de la rama hechos y la revisión de código aplicada, nombra los commits aún no desplegados, ejecuta `hive update --source <checkout de la rama> --rev feat/question-card-alone --out FILE` y `hive apply --plan FILE`, y dice al usuario que las sesiones abiertas necesitan reinicio. Antes de desplegar incorpora `development` a la rama, de modo que lleve la línea de T7; el commit que la borra se añade al cerrar la prueba, justo antes de integrar. El binario no se recompila: el cambio es solo de contenido. Los refrescos posteriores de otras sesiones siguen la línea de T7; una sesión que aun así despliegue desde `development` pisa la regla, y T6 lo detecta por las ventanas de despliegue y no cuenta esas rondas.

**Historical verification (2026-11-16; superseded by T9 for local text and superseded deployment evidence):** en el archivo de guía instalado de cada host que `hive status --hosts codex,claude,grok,pi,opencode,cursor,antigravity --scope user` da como instalado (sus rutas las lista `hive doctor`), `rg -q "Write the message as visible text first" <archivo>` termina con código 1 y `rg -q "through that tool alone when it also gives each option a description field" <archivo>` termina con código 0; `hive releases` muestra la instantánea desplegada desde la rama.

### T6 — La prueba se mide y el usuario decide la integración

- [ ] Informe de las rondas posteriores al despliegue autorizado de T9, contra AC9–AC11; la evidencia de versiones anteriores queda histórica

**Closes:** AC9, AC10, AC11, AC4

**Depends on:** T1, T9, an authorized deploy of the corrected version, and real sessions meeting AC11

**Locations:** `_support/workspace/2026-11-16-question-card-alone/measure/`; native host session histories, read-only

**Execution:** main thread; delegate a bounded host review when the applicable delegation rule requires it

**Test approach:** check

**Changes:** Record the new deployment cutoff before counting. Count new sessions after it, or later rounds where a host demonstrably reloaded the corrected guidance. Exclude other guidance versions. Read each unanswered card and potential lost round against the host instructions and the session-review rubric; examine tool identity, receipts, user answers and available UI evidence without inferring queue resolution from chat. Keep the prior 20 Claude Code / 10 Codex rounds and 3 unanswered cards (2 Codex) thresholds. If Codex has two lost rounds, revise the plan without restoring a native-host conflict. Review the existing watch topics at the authorized close. No integration or model pilot is authorized by this task.

**Verification:** Zero lost rounds under AC11; each counted round satisfies AC9 and each unanswered card AC10. The installed version and its persistence after another session's refresh establish AC4. Ordinary tests cannot close these behavior criteria.

### T7 — La línea temporal de refresco está en `development`

- [x] El `AGENTS.md` raíz dice que, mientras la rama `feat/question-card-alone` exista sin integrar, el refresco de la instalación despliega su unión con `development`

**Closes:** AC4

**Depends on:** none

**Locations:** `AGENTS.md` raíz, sección «Refreshing a local installation»

**Execution:** main thread

**Test approach:** check

**Changes:** añade una frase al párrafo del checkout del mantenedor: mientras la rama de prueba `feat/question-card-alone` exista sin integrar, cada refresco construye la unión de esa rama con `development` sin worktree (`git merge-tree --write-tree` y `git commit-tree`), despliega con `--rev` en ese commit y, tras un despliegue correcto, guarda el commit en `refs/hive/trial-join`; no escribe en la rama ni en el remoto. Si la unión falla, no despliega contenido e informa la salida del comando. (Revisado tras la revisión de código, que descartó el worktree temporal.) Rama `chore/trial-refresh-line`, pull request propio en modo `ready`; el usuario lo integra. No toca `content/`, así que no se despliega nada.

**Verification:** `rg -n "feat/question-card-alone" AGENTS.md` muestra la frase en «Refreshing a local installation»; un ensayo de sus comandos (unión con `git merge-tree`, `hive update --rev <unión> --dry-run`) termina con código 0 sin cambios en la rama, y una referencia inexistente hace salir a `git merge-tree` con código 1; `go test ./tests/...` pasa; el pull request tiene CI en verde. AC4 queda sin verificar hasta T6.

### T8 — La regla distingue la tarjeta que se cierra sola de la que cancela el usuario

- [x] El punto «When a card comes back unanswered» de la guía global, sus dependientes y la instalación local siguen D9

**Closes:** AC10, AC11

**Depends on:** T5

**Locations:** `content/guidance/global.md`; `content/skills/harness-audit/references/local-instructions.md`; `site/src/content/docs/host-limits.md`; `_support/docs/architecture/agent-delivery.md`; `_support/docs/harness-engineering/session-review-rubric.md`; la instalación local; `_support/workspace/2026-11-16-question-card-alone/measure/cc/extract.py` y `measure/codex/cards.py`

**Execution:** main thread

**Test approach:** check

**Historical changes (2026-11-16):** reescribe ese punto en dos ramas, «The card closed by itself» y «The user left it unanswered», como fija `design.md` «D9»; alinea la referencia de `harness-audit`, la página de límites por host (sección nueva «If you cancel a card»), la arquitectura y la fila de la rúbrica; añade a los extractores de Claude Code y de Codex, para cada tarjeta sin contestar (en Codex también la que no tiene resultado, que es la cancelada con Esc), si el agente volvió a preguntar por la tarjeta, y el siguiente mensaje del usuario para la lectura a ojo; anota para el pull request el cambio de tamaño de `global.md` y los lugares revisados; pasa `/code-review`, hace commit, push y redespliega con los pasos del párrafo temporal de `AGENTS.md`. Más de 3 archivos, pero en el hilo principal: es una corrección de una frase por archivo sobre texto que esta sesión acaba de escribir.

**Historical verification (2026-11-16; superseded by T9 for local text and superseded deployment evidence):** `go test ./...` y `npm --prefix site run build` terminan con código 0; `rg -c "The card closed by itself:|The user left it unanswered:" content/guidance/global.md` da 2 y la misma búsqueda encuentra ambas frases en la guía instalada de cada host tras el despliegue. AC10 y AC11 quedan sin verificar hasta T6.

### T9 — One channel per question, within native host restrictions

- [x] Correct the canonical rule, align its consumers and retained plan, and verify the local candidate

**Closes:** AC9, AC10, AC11, AC12

**Depends on:** the user's 2026-11-17 approval for the local correction

**Locations:** `content/guidance/global.md`; `site/src/content/docs/host-limits.md`; `_support/docs/architecture/agent-delivery.md`; `_support/docs/architecture/tool-dependencies.md`; `_support/docs/harness-engineering/session-review-rubric.md`; `README.md`; `llms.txt`; this change folder

**Execution:** main thread integrates the canonical rule and bounded documentation edits; an independent `hive-review-plan` review covers this revision

**Test approach:** check

**Changes:** Follow `design.md` «Proposed shape». Retire superseded criteria without deleting them, preserve historical evidence and the existing trial thresholds, and leave deployment and behavior verification pending. Checked consumers: `flow-plan`, `flow-verify` and `viewer-publication`, `harness-audit` local-instructions, agents (no question directives), and `tests/pilot/regression.go` (existing S1 checks surrounding text). No synthetic regression fixture is added. Record the global-guidance size delta with the candidate's final report.

**Verification:** Run `python3 -I <flow-plan-dir>/scripts/check_change.py <absolute-change-folder>`, ordinary content/Go checks and the site build, and review the resulting diff against the canonical rule. A passing local check does not verify AC9–AC11 in real sessions. AC12 requires the built page and aligned text; T6 closes behavior after an authorized deployment. The user subsequently authorized native Codex review, commit/push and local installation, without merge.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` and `npm run build` in `site/` | Covered by the implementation |
| Historical code review | Previous published candidate | Nativa: Claude Code `/code-review` (D5 A) | Historical answer, not current-session publication authority |
| Native code review | T9 candidate before publication | Codex `review --uncommitted`: no actionable findings | Explicit approval, 2026-11-17 |
| Trial | After the authorized T9 deploy, in real sessions on this machine | The report of T6 against AC11 | Local deployment authorized; no model pilots |

Sin cambio de interfaz: no hay recorrido de UI. La validación humana es la prueba misma: el usuario trabaja en sus sesiones con la regla nueva y decide la integración con el informe de T6.

## Checkpoint

- Updated: 2026-11-17
- Stage: ready · corrected trial published and installed; PR #491 stays open without merge
- Candidate: correction 420b2e37, synchronized with development 15b2e402 in df96a4d6; this metadata-only checkpoint is the next branch commit
- Open findings: none from the independent plan review or native Codex review
- Deployment: join 43c79157f2bbb59c8e21eace799afb2189701b33 (trial df96a4d6fd1f82a50365d3bfe2a983b0e0c0e785; base 15b2e402c72543b74d63e50df1bc2b188638212a); release 76eafee953426655964ad112dfa122b5d2ee74d76994b26c1820f91b78040aac; cutoff 2026-11-17T08:49:56Z
- Next: T6 real-session observation against AC9–AC11, plus persistence through a later refresh for AC4. Count only sessions that loaded this rule. Earlier cutoffs are historical. No merge or model pilot is authorized.
- Close hygiene pending: task worktree and reproducible scratch. Review evidence remains locally in `_support/workspace/2026-11-16-question-card-alone/codex-t9-review.log` while the trial is open.

### Evidence

Historical evidence below predates T9 and does not verify its uncommitted candidate.

- go -C <repo> test ./... · exit 0 · 27e752a0 (árbol sin cambios desde la ejecución)
- npm --prefix site run build · exit 0, 19 páginas · 27e752a0
- go test ./tests/... en chore/trial-refresh-line · exit 0 · 27da2b19
- ensayo del refresco: worktree temporal + hive update --dry-run · exit 0, sin worktrees restantes · 27da2b19
- go -C <repo> test ./... · exit 0 · candidato con las correcciones de ambas revisiones
- npm --prefix site run build · exit 0 · mismo candidato
- hive update --source <checkout> --rev aceff0f0 --out FILE + hive apply · exit 0 · unión de 15f74421 y 582997ab, guardada en refs/hive/trial-join
- rg de la regla nueva y de la anterior en los archivos de guía instalados · nueva en 6 archivos (7 hosts), anterior en ninguno
- go -C <repo> test ./... · exit 0 · candidato a1ee85f7
- npm --prefix site run build · exit 0 · candidato a1ee85f7
- hive update --rev 4c5d9bc3 --out FILE + hive apply · exit 0 · unión de ab4db72f y 582997ab, en refs/hive/trial-join
- rg de «The card closed by itself:» y «The user left it unanswered:» en la guía instalada · ambas en los 6 archivos de los 7 hosts; regla anterior en ninguno
- git merge origin/development (007d3999) en la rama + go test ./... + npm --prefix site run build · exit 0 · 6d1c40ed
- hive update --rev fe3c4438 + hive apply · unchanged: la instalación ya tenía ese contenido · refs/hive/trial-join = fe3c4438

## Review status and progress

2026-11-16 — Plan escrito.
2026-11-16 — Revisión tras la revisión del plan: D6 A añade T7 y cambia T5; D7 A sustituye AC3 por AC6 en T1 y T6.
2026-11-16 — T1 · uncommitted · unverified: AC6 hasta T6; la línea base de Codex se reproduce: 63 filas en los archivos, de ellas 47 tarjetas que esperan desde el 2026-11-09, con 10 vacías
2026-11-16 — T7 · 27da2b19 · unverified: AC4 hasta T6; el ensayo de los comandos de refresco con --dry-run terminó sin dejar worktrees; push y pull request pendientes de la revisión de código
2026-11-16 — T2 · 6545eca3 · unverified: AC1 y AC2 hasta T6 (sesiones reales)
2026-11-16 — T3 · 7fab11bf · unverified: AC1 hasta T6
2026-11-16 — T4 · 27e752a0 · unverified: none
2026-11-16 — T5 · 15f74421 · unverified: AC4: la regla nueva está en la guía instalada de los siete hosts (Grok la lee de ~/.claude/CLAUDE.md) y la anterior en ninguno; falta observar que sobrevive a un refresco de otra sesión, en T6
2026-11-16 — Revisión por D9: T8 añadida; T2 deja de cerrar AC2 (retirado); T6 cierra AC7 y AC8.
2026-11-16 — T8 · a1ee85f7 · unverified: AC7 y AC8 hasta T6 (sesiones reales); las correcciones de la segunda pasada de revisión no se revisaron de nuevo

2026-11-17 — T9 added after the user approved a local correction: native host compatibility and one channel per pending question. AC9–AC12 replace AC1, AC7, AC8 and AC5. T2, T4, T5 and T8 reopened; prior checks and deployment evidence retained as history. T6 remains pending a new authorized deploy and real-session evidence. No model pilot, commit, publication or deployment performed for T9.

2026-11-17 — T9 local candidate complete, uncommitted on af4093ec. T2 and T4 locally reverified; T5 and T8 installation remain open. AC12 checked in source and generated `host-limits` HTML; AC9–AC11 remain unverified in real sessions. Independent `hive-review-plan` review found no material gaps; no dedicated code review chosen in this session. No commit, publication, deployment or model pilot.

### T9 local verification

Commands ran against `.claude/worktrees/question-channel-fix`, using its absolute path for repository commands. Evidence covers the uncommitted T9 candidate, not a deployed version.

- `go -C <worktree> vet ./...` — exit 0.
- `go -C <worktree> test ./...` — exit 0, all packages; CLI package 250.259 s. After final content edits, `go -C <worktree> test ./tests/content/...` — exit 0.
- `python3 -m unittest discover -s <worktree>/tests/skills -p '*_test.py'` — exit 0, 119 tests in 13.919 s.
- `npm --prefix <worktree>/site run check` — exit 0, no diagnostics; `run lint` — exit 0, no fixes; `run build` — exit 0, 20 pages. Build warned about empty i18n and missing docs/404, outside the changed page.
- Site scripts used the existing main-checkout dependencies after package.json and pnpm-lock.yaml compared equal. The pnpm wrapper rejected the temporary node_modules symlink; npm ran the same declared scripts without installing dependencies. The generated host-limits HTML contains the new channel and native-mode explanation. Temporary dependency symlink removed after checks.
- `python3 -I <flow-plan-dir>/scripts/check_change.py <change-folder>` and `git -C <worktree> diff --check` — exit 0.
- Independent plan review — no material findings; covers the revised folder and its consistency with canonical guidance and named consumers, not live host behavior or a dedicated code review.
- Global guidance: 56,383 → 57,451 UTF-8 bytes (+1,068) from af4093ec. No change to pilot detectors or synthetic regression fixtures.

2026-11-17 — The user explicitly approved native Codex review, commit/push to the trial branch and local deployment, with no merge. Native `codex review --uncommitted` completed with no actionable defects; it checked the revised rule and consumers, plan coherence and diff whitespace. It did not test live host behavior. The prior local-only authorization and no-review-selected statements above are historical.

### T9 delivery verification

- Native Codex review of the uncommitted correction later committed as 420b2e37: no actionable defects; no live host behavior check.
- Combined candidate df96a4d6 on development 15b2e402: full Go suite exit 0 (CLI 223.207 s), go vet exit 0, 146 Python tests passed in 15.557 s, site check/lint/build exit 0 (20 pages), change validation and secret scan passed. No CI workflow targets development PRs.
- T5 and T8 installation checks passed for all seven hosts. Native channel selection, required unanswered-card response and removal of the old asynchronous-supplement instruction were checked in installed guidance. Grok shares Claude's block; Antigravity splits guidance across files.
- New cutoff: 2026-11-17T08:49:56Z. Installation receipt c581c77f33e29145ad1c53cd8ed0761a; first corrected join 43c79157f2bbb59c8e21eace799afb2189701b33. A metadata-only final refresh does not restart this cutoff.
- AC12 locally verified. AC9–AC11 remain unverified in real sessions; AC4 persistence through a later session's refresh remains pending for this version. No model pilot or merge.
- Final global guidance: 57,162 bytes; development 15b2e402: 52,219 (+4,943 for the whole trial). T9 alone added 1,068 bytes before synchronizing the base.
