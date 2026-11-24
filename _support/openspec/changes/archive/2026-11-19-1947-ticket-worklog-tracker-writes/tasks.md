# Tasks

Base `development` en `81d5b4f1`. Dos unidades: E1 (gestor y auditor, Go) y E2 (guía, scripts y documentación), en ese orden, porque la guía de E2 nombra `hive time`. T1 va a un hijo `hive-build-backend` por tamaño (más de 3 archivos en `tooling/`); el resto en el hilo principal, porque son reglas que dependen de leer el conjunto y se solapan en `global.md`.

### T1 — `hive time` en el gestor

- [x] `hive time` imprime el tiempo activo, la espera por el usuario, la espera de CI y el convertido de la sesión actual o de un archivo dado, en Claude Code, Codex, Grok y Pi, leyendo la transcripción completa.

**Closes:** AC2

**Depends on:** none

**Locations:** `tooling/cli/` (comando `time`, un lector de horas por host, pruebas), `tooling/cli/testdata/` con fixtures derivados de transcripciones reales sin nombres, rutas ni claves de proyecto

**Execution:** `hive-build-backend`

**Test approach:** tdd — un fixture por host con esperas humanas, una respuesta de tarjeta, una notificación de tarea en segundo plano (Claude), una espera de CI y un archivo truncado; la suma esperada se calcula a mano

**Changes:** comando con `--since`, `--factor` (2.0 por omisión; menor que 1 o mayor que 3 sale con código 2), `--session-file` y `--host`; reusa `resolveViewIdentity` y `locateHostSessionFile`; lectura completa por líneas, no el corte de cola de `readHostSessionFile`; `partial` cuando no lee el archivo entero; `no_session_file` con código 0. Antes de escribir los lectores de Grok y Pi, leer un archivo real de cada uno y fijar cómo distingue un mensaje humano; si no lo distingue, su lector da `partial`. No toca el visor.

**Verification:** `go -C <repo> test ./tooling/...`; `go -C <repo> run ./tooling/cli time` en esta sesión de Claude Code da `result: ok`; con `--session-file` sobre los rollouts `01a1f358` y `01a1f37d`, la espera humana coincide ±1 min con los huecos entre fin de turno y mensaje del usuario medidos a mano, y sobre la transcripción de Claude más grande de la máquina (59.5 MB) da `ok` con una ventana que empieza en su primer evento.

### T2 — Criterios del auditor

- [x] `--audit-session` marca la caché de estados que no se escribe y el «Pruébalo tú» que falta en un cambio de UI.

**Closes:** AC7, AC8

**Depends on:** none

**Locations:** `tests/pilot/session_audit_routes.go` (y el parser de `session_audit.go` si no expone las llamadas de Engram dentro de celdas `exec`) con sus pruebas, `tests/fixtures/sessions/{tracker_states_cache_written,try_it_block_in_report}/`, `_support/docs/harness-engineering/{rule-enforcers,session-review-rubric}.md`

**Execution:** main thread

**Test approach:** tdd — cada fixture falla con el auditor actual

**Changes:** `tracker_states_cache_written`: una búsqueda de `tracker/states/` con resultado vacío (también anidada en una celda `exec` de Codex), seguida de un movimiento de ticket (`acli jira workitem transition`, una herramienta MCP de transición de Atlassian o de estado de Linear, `gh issue edit`/`close`) y sin escritura de ese tema (`mem_save`, `mem_update`, `engram save`) → `fail`. `try_it_block_in_report`: una edición de archivo de interfaz y el último mensaje del agente tras leer `flow-verify` sin encabezado de la lista «Pruébalo tú», «Pruébalo», «Try it» → `fail`; solo mira ese mensaje, no el texto de guías leídas. Mensajes de commit con las observaciones (Codex 0.170.0, `gpt-6.2-sol` configurado, rollout `01a1f37d`, 2026-11-19).

**Verification:** `go -C <repo> test ./tests/pilot` (solo); `go -C <repo> run ./tests/pilot --audit-session <rollout 01a1f37d> --host codex` da `fail` en los dos.

### T3 — Revisión adversarial solo a petición

- [x] La ronda de entrega no pregunta la revisión adversarial ni sondea hosts; `interrogate` se lee cuando el usuario la pide.

**Closes:** AC1

**Depends on:** none

**Locations:** `content/skills/hive-rule-delivery-ask/SKILL.md` (quitar la viñeta y su mención en la descripción), `content/skills/hive-rule-interrogate/SKILL.md` (descripción, primer párrafo y «Availability»: corre cuando el usuario pide la revisión), `content/skills/hive-rule-questions/SKILL.md` (lista de encabezados fijos), `content/skills/flow-verify/SKILL.md`, `content/guidance/global.md:123` (disparador «When the user asks for a review by another provider's agent»), `tests/content/leaf_links_test.go:93` (el enlace a `interrogate` pasa a `flow-verify`), `tests/content/single_home_test.go` si fija frases movidas, `_support/docs/architecture/instruction-resources.md`, `site/src/content/docs/flows/{plan,verify}.md`, delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** check — pruebas de contenido

**Changes:** la revisión por otro proveedor corre solo a petición explícita del usuario, sobre el candidato, con sus límites actuales; no se agrega revisión de planes por otro proveedor; no cambia la pregunta del mecanismo de revisión (D38). `tests/pilot` conserva el encabezado «Revisión adversarial» en `sessionHeaderNames` para auditar transcripciones anteriores.

**Verification:** `go -C <repo> test ./tests/content`; `rg -n -i adversarial content/` solo encuentra `interrogate`, su línea del índice y `flow-verify` a petición.

### T4 — Escritura en Jira por capacidad y token global

- [x] La guía escribe en Jira con `acli`, el MCP o REST con el token global, y un script hace transiciones, worklogs y adjuntos sin exponer el token.

**Closes:** AC6

**Depends on:** none

**Locations:** `content/skills/hive-rule-trackers/SKILL.md` (descripción y orden), `references/jira.md`, `scripts/jira_rest.py` (nuevo), `tests/skills/jira_rest_test.py` (nuevo), `content/guidance/global.md` («Scope and authorization»: la línea de recursos privados admite REST con el token global cuando la CLI y el MCP no tienen la capacidad; «Secrets»: el almacén `~/.config/hive/secrets/` que solo el usuario escribe, comprobado sin mostrar su contenido, nunca pasado a otros agentes), `content/skills/hive-rule-interrogate/SKILL.md` «Limits» (el comando del otro proveedor corre sin acceso al almacén global)

**Execution:** main thread

**Test approach:** tdd — el script contra un servidor HTTP local de prueba (sustituto) y un `acli` falso en `PATH`: sin token en argumentos ni salida, error HTTP sin cabeceras ni cuerpo, permisos del archivo verificados, archivo ausente da código 3 sin crearlo, worklog sin comentario, `worklog-list` filtra por la cuenta

**Changes:** orden por capacidad; transiciones por REST cuando `acli` no las lista; adjuntos por REST; token del archivo global, email y sitio de `acli jira auth status`. No cambia Linear ni GitHub salvo la línea de que no registran tiempo (T5).

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'`; `go -C <repo> test ./tests/content`; `rg` sin «Read private resources only through these authenticated tools» en la referencia de Jira.

### T5 — Tiempo trabajado: medir, convertir y registrar

- [x] La guía mide con `hive time` o las marcas, convierte, registra el worklog en la entrega, lo actualiza al cierre, suma las sesiones previas del plan, guarda el tiempo real local y no escribe razonamiento en el tracker.

**Closes:** AC3, AC4, AC5

**Depends on:** T1, T4

**Locations:** `content/skills/hive-rule-trackers/references/time.md` (nuevo) enlazado desde la hoja, `scripts/work_time.py` (nuevo), `tests/skills/work_time_test.py` (nuevo), `references/{linear,github}.md`, `content/skills/hive-rule-ticket-state/SKILL.md` (cuándo y descripción), `content/skills/flow-plan/SKILL.md` o su referencia del estado del plan (host e ID de sesión en el estado guardado), `content/skills/flow-close/SKILL.md` paso 7 (ID del worklog, tiempo real, factor y tipo en el registro de cierre), `content/guidance/global.md` (autorización del worklog en «Scope and authorization», líneas del índice de ticket-state y trackers que nombran el worklog), delta de `workflow-lifecycle`

**Execution:** main thread

**Test approach:** tdd para `work_time.py` (la secuencia de marcas derivada de un fixture de T1 da los mismos cuatro campos); check para la guía

**Changes:** ventana desde el primer mensaje que trabaja cada ticket (D48, D51); cada ticket trabajado a la vez recibe el tiempo completo (D41); leer los worklogs de la cuenta antes de agregar y actualizar el mismo al cierre; worklog solo con el tiempo convertido, sin descripción; sesiones previas re-medidas desde el estado del plan; tiempo real, factor, tipo e ID en el registro de cierre; nada de tiempo en «Pending work» ni en comentarios; Linear y GitHub no registran tiempo.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'`; `go -C <repo> test ./tests/content`; tamaño de `global.md` y de la hoja de trackers reportados.

### T6 — Documentación

- [x] README, llms.txt, el sitio y la arquitectura describen `hive time`, los scripts, el registro de tiempo y la revisión adversarial a petición.

**Closes:** AC9

**Depends on:** T1, T3, T5

**Locations:** `README.md`, `llms.txt`, `site/src/content/docs/flows/{plan,verify,close}.md`, `_support/docs/architecture/{deployment-manager,instruction-resources}.md`

**Execution:** main thread

**Test approach:** check

**Changes:** describen el comportamiento y enlazan las hojas; no copian sus reglas.

**Verification:** `go -C <repo> test ./tests/content`; build del sitio si su prueba lo pide.

### T7 — Suite completa y revisión

- [x] Suite completa y revisión de código de cada unidad.

**Closes:** AC1, AC2, AC3, AC6, AC7, AC8, AC9

**Depends on:** T1–T6

**Locations:** —

**Execution:** main thread

**Test approach:** check

**Changes:** ninguno.

**Verification:** `go -C <repo> test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` (el repositorio no corre pruebas en CI, `CONTRIBUTING.md:15`), y `/code-review` sobre cada unidad.

### T8 — Observación tras el despliegue

- [ ] Una sesión real posterior registra y lee de vuelta un worklog en Jira, actualiza el mismo al cierre, guarda el tiempo real local y no sondea hosts.

**Closes:** AC1, AC4, AC5, AC6

**Depends on:** T7, el despliegue y el prerrequisito del token

**Locations:** `engram:tricell-hive/pending/ticket-worklog-tracker-writes`

**Execution:** main thread (auditoría con `--audit-session`)

**Test approach:** check

**Changes:** ninguno; queda abierta a propósito hasta que corra una sesión real. Mide los hechos externos no medidos: actualizar un worklog y subir un adjunto por REST, y si el MCP de Atlassian escribe worklogs o adjuntos.

**Verification:** auditoría de la sesión real y lectura del worklog en Jira.

## Delivery units

### E1 — PR gestor y auditor · depends on: none · files: tooling/cli/, tests/pilot/, tests/fixtures/sessions/, _support/docs/harness-engineering/

- [x] unit: `go -C <repo> test ./tooling/... ./tests/pilot`
- [x] live: `hive time` en esta sesión, sobre los dos rollouts y sobre la transcripción de 59.5 MB · pass: `result: ok` y espera humana ±1 min de la medición a mano

### E2 — PR guía, scripts y documentación · depends on: E1 · files: content/, tests/content/, tests/skills/, site/, README.md, llms.txt, _support/docs/architecture/

- [x] unit: `go -C <repo> test ./tests/content` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`
- [x] live: guía instalada tras `hive update` · pass: `rg` confirma `references/time.md`, el orden por capacidad y la ausencia de la pregunta adversarial en `~/.agents/skills/`

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final de cada unidad, local | `go -C <repo> test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Revisión de código | Candidato de cada unidad, antes de la parada | `/code-review` | Respuesta de revisión de la sesión |

Parada humana (`interactive`): antes de cada push, el diff y la salida de `hive time` en esta sesión.

## Checkpoint

- Updated: 2026-11-19
- Stage: close · archive
- Last commit: 2f456f75
- Open findings: none
- Next: archive decision (T8 open)
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/ticket-worklog · docs/archive-ticket-worklog

### Evidence

- before captures: none (no UI)
- AC2: hive time ok on this session, rollouts 01a1f358 (--session-file and --session) and 01a1f37d, 62.3 MB Claude file; --factor 0.5 exits 2 · 9cf91b07
- AC3: work_time sum = hive time on codex fixture (420/150/302/540) · 9cf91b07
- AC7/AC8: --audit-session on 01a1f37d gives fail on both · 11fb1ffd
- blast radius: 55 real sessions, existing criteria identical base vs candidate · 9cf91b07
- tests/skills 275 OK, tests/content ok · 9cf91b07
- E1 unit: go test ./tooling/... ./tests/pilot ok · a610d1e4
- E1 live: hive time ok on this session, 01a1f358, 01a1f37d and the 62.3 MB transcript (the 59.5 MB one also ok); waits match hand measurement · a610d1e4
- E2 unit: tests/content ok, tests/skills 275+ OK · 296ec3f2
- E2 live: installed time.md, jira_rest.py, capability order, no adversarial in delivery-ask after hive update from 2f456f75 · 2f456f75

## Review status and progress

2026-11-19 — Plan escrito y revisado (21 huecos aplicados).
2026-11-19 — T2 · 11fb1ffd · unverified: none
2026-11-19 — T1 · d7ad2d97 · unverified: Codex CI wait totals (271 s, 118 s) not measured independently; Pi always partial
2026-11-19 — T3 · 7a69913d · unverified: AC1 in a real session after deploy (T8)
2026-11-19 — T4 · 9cf91b07 · unverified: AC6 against a real Jira: REST worklog update and attachment upload not measured (T8)
2026-11-19 — T5 · 7a69913d · unverified: AC4, AC5 in a real session (T8); AC3 checked against work_time tests, not field-by-field against a hive time fixture
2026-11-19 — T6 · 918114ea · unverified: none
2026-11-19 — T7 · 296ec3f2 · unverified: non-blocking review findings left for the close: Jira adjustEstimate default, --started on Python <=3.10, duplicate text decoder in time_read.go, MCP moves through fields other than state/status, Codex <user_action> timing
