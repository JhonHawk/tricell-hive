# Tareas

Base `development` en `d23f9a88`, rama `feat/viewer-passive-capture`, un solo pull request (D7 B). `gh-311-minimal-research-entry` quedó archivado (PR #370).

**Quién hace cada tarea.**
- **Hijos `hive-build-backend`:** las tareas de binario y servidor, porque cada una supera 5 lecturas grandes o 3 archivos editados. Corren una tras otra, salvo T5 y T6.
- **T5 y T6 a la vez:** cuando T1 congela el contrato tipado, porque escriben archivos distintos; la comprobación de integración entre las dos va en T10.
- **Verificación por tarea:** las de riesgo (contrato, concurrencia, configuración global) llevan un hijo `hive-verify-task`.
- **Hilo principal:** T0, T8 y T10; ejecuta las comprobaciones, hace los commits y marca las tareas.

### T0 — Comprobación por host, contra la documentación y la versión instalada

- [x] Un registro por host de cada afirmación que usa el diseño, marcando qué está documentado, qué observado y qué inferido, con la URL o ruta, la fecha y la versión.

**Closes:** AC2, AC3, AC8 (condición previa)

**Depends on:** ninguna

**Locations:** `_support/docs/architecture/agent-delivery.md` (sección de hooks por host) o una referencia nueva en `_support/docs/architecture/`

**Execution:** hilo principal (con `hive-research` si hacen falta más de 5 lecturas grandes por host)

**Changes:** cubre, por host:
- **Claude Code:** eventos y `task_subject` en `TaskCreated`/`TaskCompleted`.
- **Codex:** `SessionStart`, `SubagentStart`/`SubagentStop`, `SessionEnd`, la variable de sesión del shell y `trusted_hash`.
- **Grok:** un segundo archivo en `~/.grok/hooks/` y la carga en camelCase.
- **OpenCode:** la carga automática de `plugins/`, `todo.updated` y `OPENCODE_SESSION_ID`.
- **Pi:** `PI_SESSION_ID` y `agent_settled`.
- **Cursor CLI:** hooks y sus campos.

Lo que no se verifique queda marcado como no verificado en el diseño y en el informe del gestor.

**Verification:** cada afirmación del diseño tiene fuente o queda marcada como no verificada.

### T1 — Contrato: revisión 3, eventos del host y delta de la spec

- [ ] El contrato tipado y la spec describen `host_observed`, `POST /api/host-events`, la capacidad `host_events`, la cola semántica y las órdenes `work` nuevas.

**Closes:** AC4, AC9 (parte de contrato)

**Depends on:** T0

**Locations:** `tooling/viewer/timeline_contract.go`, `timeline_contract_v3.go`, `timeline_protocol.go`, `tooling/viewer/frontend/src/timeline-contract*.ts`, `_support/openspec/changes/viewer-passive-capture/specs/walkthrough-viewer/spec.md`

**Execution:** hive-build-backend; después hive-verify-task

**Test approach:** tdd (contrato público)

**Changes:**
- `TimelineContractRevision` pasa a 3.
- Se añaden el origen `host_observed` y el evento del host, que actúa como el productor aceptado de su sesión y lleva un ID de operación derivado del evento.

No cambia ninguna operación existente.

**Verification:**
- `go test ./tooling/viewer/...` con estos casos:
  - un evento del host que intente cambiar una etapa, la autoridad o el estado del trabajo se rechaza;
  - un evento de una sesión que no es productora se rechaza;
  - reenviar el mismo evento lo aplica una sola vez.
- `go test ./tests/content/...` valida el delta.

### T2 — Identidad resuelta y trabajo actual

- [x] `hive view work` funciona sin `--host` ni `--native-session-id` cuando hay una fuente unívoca, y si no la hay lo rechaza.

**Closes:** AC2

**Depends on:** T0

**Locations:** `tooling/cli/view_work.go`, `view_work_engine.go`, nuevo `tooling/cli/view_identity.go`, `tooling/cli/doctor_sessions.go` (lectura compartida de mapas pid)

**Execution:** hive-build-backend; después hive-verify-task

**Test approach:** tdd

**Changes:**
- Resolución en el orden del diseño, rechazando un mapa ambiguo o anterior al último cambio de sesión.
- Archivo del trabajo actual por (host, ID nativo).
- Registro automático del proyecto en `work open`.

Solo se usan las variables que T0 verifique.

**Verification:**
- Tests con:
  - variable de entorno;
  - mapa pid unívoco;
  - dos sesiones bajo el mismo proceso: rechaza;
  - un mapa anterior a un `/clear`: rechaza;
  - ninguna fuente: rechaza, como hoy.
- Comprobación real: `hive view work list` en una sesión de Claude Code sin banderas.

### T3 — `work` sin plan, `work sync`, `work evidence` y cierre que completa la etapa

- [x] Un build sin plan se publica con `work open`, `work sync` y `work sync --close`, que lee el trabajo al cerrar.
- [x] Las capturas, con su candidato y una vista que sustituye a otra, se publican con un solo `work evidence`.

**Closes:** AC1, AC7

**Depends on:** T2

**Locations:** `tooling/cli/view_work.go`, `view_work_lifecycle.go`, `view_work_publish.go`, `view_work_contract.go`, nuevo `view_work_sync.go`, nuevo `view_work_evidence.go`

**Execution:** hive-build-backend; después hive-verify-task

**Test approach:** tdd; antes, un test de characterization sobre `work stage` y `work task update` para conservar su comportamiento

**Changes:**
- `--change` pasa a ser opcional, y la revisión del plan deja de exigirse sin carpeta de cambio.
- Las respuestas llevan los IDs de segmento.
- Los segmentos se vinculan a la etapa activa.
- `work contract` se acorta a las órdenes nuevas.

Las operaciones de bajo nivel no cambian. La tarea comprueba si el fallo de #358 también ocurre en `work task update` y lo anota en la propuesta.

**Verification:** tests de extremo a extremo contra un visor de prueba:
- build sin plan con 3 tareas en ≤ 3 órdenes, sin `stage_binding_mismatch`;
- repetir `work sync` devuelve `unchanged`;
- `work evidence` con dos imágenes, candidato y `--supersedes`.

### T4 — Cola semántica y lock por trabajo

- [x] Con el visor parado, `work open`, `work sync` y `work evidence` responden `queued`, y la siguiente orden con el visor activo los envía en orden, construidos sobre el estado actual.
- [x] Dos trabajos publican a la vez.

**Closes:** AC5, AC6

**Depends on:** T3

**Locations:** `tooling/cli/view_work_journal.go`, `view_work_engine.go`, nuevo `view_work_queue.go`

**Execution:** hive-build-backend; después hive-verify-task

**Test approach:** tdd; `go test -race`

**Changes:**
- Cola de declaraciones semánticas en el journal, con su ID de operación.
- Al enviarlas, las peticiones se registran con sus bytes exactos antes de transmitirse.
- Lo que ya no aplique se informa como rechazado, sin sobrescribir.
- Lock por trabajo; el lock global queda para la creación y la migración.
- Los journals y las intenciones pendientes anteriores siguen pudiéndose reanudar.

**Verification:** `go test -race ./tooling/cli/...` con estos casos:
- visor parado al abrir;
- reanudación tras matar el proceso;
- una etapa ya completada por otro: rechazada;
- dos procesos sobre trabajos distintos;
- dos procesos sobre el mismo trabajo: `busy`.

### T5 — Servidor: eventos del host y última actividad; frontend con procedencia

- [x] `POST /api/host-events` actualiza la última actividad como metadato y, en un trabajo sin tareas registradas, refleja las tareas nativas como `host_observed` hasta 50.
- [x] La interfaz distingue esa procedencia.

**Closes:** AC4, AC9

**Depends on:** T1

**Locations:** `tooling/viewer/server.go`, nuevo `tooling/viewer/host_events.go`, `timeline_domain.go`, `timeline_store.go`, `timeline_maintenance.go`, `tooling/viewer/frontend/src/**` (tareas, cabecera del trabajo), `tooling/viewer/static/app/` y `bundle-manifest.json` (bundle regenerado)

**Execution:** hive-build-backend para el servidor; después hive-build-frontend para la interfaz. Al ser una interfaz, `flow-verify` la recorre y un `hive-review-ux` la revisa.

**Test approach:** tdd en el servidor; check en el frontend (extender `timeline-tasks.test.tsx`)

**Changes:**
- Capacidad `host_events` en `/api/health`.
- La actividad se guarda como metadato, no como evento.
- Sin efectos en etapas, autoridad ni cierre.

**Verification:**
- `go test ./tooling/viewer/...` con estos casos:
  - un evento del host no invalida el `expected_version` de una mutación del agente;
  - con tareas registradas no se añade ninguna;
  - el tope de 50 se respeta;
  - sin trabajo actual, solo se mueve la última actividad.
- `pnpm check && pnpm lint && pnpm test && pnpm build && pnpm verify:dist`.
- Capturas de escritorio y móvil, en claro y oscuro, de un trabajo con tareas del host.

### T6 — `hive hook` y conciliación

- [x] `hive hook <host> <evento>` registra el mapa pid, resuelve el trabajo actual y su etapa, envía o encola el evento, no imprime nada y sale con 0.
- [x] `hive view work reconcile` recupera la última actividad y, cuando corresponde, las tareas nativas desde los archivos de sesión.

**Closes:** AC3, AC4

**Depends on:** T1, T4

**Locations:** `tooling/cli/main.go`, nuevo `tooling/cli/hook_command.go`, un traductor de la carga en cada `integrations/<host>/` (`claude`, `codex`, `grok`, `opencode`, `pi`, `cursor`), nuevo `tooling/cli/view_reconcile.go`

**Execution:** hive-build-backend; puede correr a la vez que T5

**Test approach:** tdd con cargas de ejemplo por host tomadas de T0, sin datos privados

**Changes:**
- El mapa pid se escribe tal como fija «Session map» en `_support/docs/architecture/host-session-events.md` (pid del proceso del host, `process_started_at`, reglas de `changed_at`, topes y modos).
- Límite de 1 MiB en la entrada y 2 s de tiempo límite.
- Cola de eventos del host cuando el visor está caído.
- Nunca deduce la muerte del proceso.
- `occurred_at` sale del host o de un dato estable del evento, nunca de la hora del hook: el ID de operación del evento se deriva de su contenido, y una hora nueva haría que un reenvío se aplicara dos veces.

**Verification:** `go test ./tooling/cli/... ./integrations/...` con estos casos, sin salida en stdout:
- entrada inválida;
- visor caído;
- carga en camelCase de Grok y en snake_case de Claude Code y Codex;
- `/clear` en Claude Code cambia la sesión del mapa.

### T7 — Registro por host en el gestor

- [ ] `hive update` instala y la desinstalación quita solo las entradas de Hive en los seis hosts.

**Closes:** AC8

**Depends on:** T6

**Locations:**
- `integrations/{claude,codex,grok,opencode,pi,cursor}/`
- `tooling/management/{plan,ownership,files,versioning}.go`
- nuevo `tooling/management/json_hooks.go` (editor seguro para `settings.json` y `hooks.json`)
- `tooling/management/*_test.go`

**Execution:** hive-build-backend; después hive-verify-task (escribe configuración global)

**Test approach:** tdd sobre un HOME temporal

**Changes:**
- **Editor JSON:** conserva el resto del archivo, guarda una copia y reconoce como propias solo las entradas con `'<hive>' hook`.
- **Codex:** no escribe `trusted_hash`; el informe pide aprobar el hook en `/hooks`.
- **Grok:** segundo archivo nativo.
- **OpenCode:** plugin de eventos en `plugins/`.
- **Pi:** extensión propia.
- **Cursor:** `hooks.json`, marcado como no verificado si T0 no lo confirma.

**Verification:**
- Tests con un HOME temporal que ya tiene hooks de herdr, rex, codegraph y Engram: instalar, actualizar y desinstalar sin tocarlos; conflicto de colisión.
- `hive update --dry-run --source <checkout>` con el binario de la rama sobre un HOME temporal muestra los seis destinos.

### T8 — Guía de `flow-walkthrough` y de los flujos

- [x] El procedimiento publica una vez por cambio de etapa con `work open`, `work sync`, `work evidence` y `work close`, sin bajo nivel ni banderas de identidad en la vía normal.

**Closes:** AC10

**Depends on:** T4, T5, T6, T7

**Locations:**
- `content/skills/flow-walkthrough/SKILL.md` y `references/viewer-publication.md`
- `references/host-session-identity.md`: se reduce a la vía explícita por host cuando el gestor no resuelve la identidad
- `content/skills/flow-build/SKILL.md:8,20`
- `flow-plan`, `flow-verify`, `flow-close` y `flow-research`
- `flow-verify/references/ui-verification.md:25`, `content/agents/review/hive-review-ux.md:14`

**Execution:** hilo principal (guía distribuida)

**Changes:**
- **Tareas del host:** no se registran a mano cuando el host las informa y el trabajo no tiene plan.
- **Lo que sigue siendo del agente:** las etapas y las validaciones.
- **Revisión completa que pide `AGENTS.md`:** global, skills, roles, `site/`, `_support/docs/architecture/`, `tests/pilot/` e `instruction-resources.md`. El PR incluye la lista de lugares revisados y el cambio de tamaño de `global.md`.

**Verification:**
- `go test ./tests/content/...` (tamaños ≤ 16.000 bytes).
- Búsqueda de `--native-session-id`, `contract --operation` y `session create` en `content/` sin resultados fuera de la vía explícita de identidad y de las operaciones avanzadas.

### T9 — Documentación

- [x] La documentación describe las órdenes nuevas, la cola, los hooks por host y sus límites.

**Closes:** AC8, AC10 (documentación)

**Depends on:** T7, T8

**Locations:**
- `site/src/content/docs/viewer.md` y `flows/*.md`
- `README.md:417-477`, `llms.txt`
- `_support/docs/architecture/deployment-manager.md:114-267,333-410,453-500`, `agent-delivery.md`
- `_support/docs/harness-engineering/session-review-rubric.md` (fila del visor)
- `CHANGELOG.md`

**Execution:** hive-write-docs

**Verification:** búsqueda de afirmaciones que dejan de ser ciertas («no watcher, new server endpoint», «never activates»); compilación del sitio si el repositorio la tiene configurada.

### T10 — Verificación del conjunto

- [x] La suite completa pasa en local, y en esta Mac los hooks instalados se disparan en una sesión real de cada host.

**Closes:** AC1–AC10

**Depends on:** T1–T9

**Execution:** hilo principal

**Verification:**
- **Local:** `go vet ./...`, `go test -race ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, las órdenes `pnpm` de T5, y una prueba de integración entre T5 y T6 (hook → servidor → tarea visible).
- **En esta Mac, antes de la parada de validación** (D14 ajustada por el usuario):
  1. Antes de tocar nada, se nombran los commits que se despliegan y se guarda una copia de `~/.claude/settings.json`, `~/.codex/hooks.json`, `~/.cursor/hooks.json` y las carpetas de hooks y plugins.
  2. El binario se compila desde la rama a la ruta instalada y se ejecuta `hive update` desde esa rama. Esto despliega también la guía de la rama en todos los hosts. Las sesiones abiertas la cargan al reiniciar.
  3. Se reinicia el visor con su puerto, carpeta de estado y TTL, porque exige la revisión 3. Las pestañas abiertas pierden el acceso.
  4. Se lanza una sesión corta real por host desde una carpeta temporal, con un prompt trivial (`claude -p`, `codex exec`, `grok -p`, `opencode run`, `pi -p`, `cursor-agent -p`). Se comprueba en el visor y en el registro de hooks de cada host que se dispararon `session_started`, la actividad y una tarea nativa donde el host las tenga.
  5. Se publica un build sin plan de prueba con ≤ 3 órdenes.
  6. Si algo falla y no se corrige en la ronda, se vuelve a `development` con `hive update` y el binario de `development`, y se restauran las copias.
- **Después del merge:** observación de las sesiones normales con `_support/sessions/2026-11-02-viewer-cost/measure-viewer-cost.py` frente a la línea base (~6,7 % en Claude Code). No forma parte de la aceptación.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, en local | La orden de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go vet ./...`, `go test -race ./...`, unittest y `pnpm` | Cubierta por la implementación |
| Interfaz | Candidato final, en local | Escritorio y móvil, claro y oscuro; `hive-review-ux` | Cubierta por la implementación |
| Hooks reales en esta Mac | Candidato final, antes de la parada de validación | Salida de `hive update`, registros de hooks por host, eventos visibles en el visor, build de prueba con ≤ 3 órdenes | D14 ajustada; aprobación del hook en Codex |
| Revisión de código | Candidato final, antes de la parada de validación | `/code-review` | D2 A |
| Observación tras el merge | Sesiones normales del usuario | Medición de costo frente a la línea base | Marca de mantenedor; aprobar el hook en Codex |

## Estado de revisión y progreso

2026-11-04 — T1 hecha en `49a9af6c` (revisión 3, `host_observed`, bundle regenerado). Verificación de T2: AC2 cumplido; dos P2 y P3 en corrección.

2026-11-05 — Codex verificado en real tras aprobar los hooks (identidad por mapa pid). Suite completa con `-race -timeout 40m` en verde. Segunda revisión: cinco locales corregidos en `79bd5e75`; UX: cinco corregidos en `f77c8765`; Codex: hooks con nombre en `885897b2`; H39 corregido en `db7bbaed` y comprobado con `hive view start` (13 s).

2026-11-05 — Prueba real en esta Mac con sesiones cortas: Claude Code (haiku; identidad por variable, tarea nativa `host:1` completada, `host_activity_at`), Grok (identidad por mapa pid), Pi (`PI_SESSION_ID`), Cursor (mapa pid desde `beforeSubmitPrompt`) y OpenCode (`OPENCODE_SESSION_ID`, `todowrite` como tarea del host completada); todos abren, declaran la etapa y cierran con 3 órdenes. Lanzadas desde el shell de Claude Code, Grok y Pi se rechazaron como identidad ambigua (correcto: host anidado); se repitieron sin la variable heredada.

2026-11-04 — T2 hecha en `56fbb304`. Comprobación real en Claude Code: `work stage` sin banderas resuelve la identidad (`unchanged`); sin la variable, se rechaza. El mapa pid aún no se ha probado con `ps` real (lo escribe T6).

2026-11-04 — T0 hecha: `_support/docs/architecture/host-session-events.md`; diseño corregido (detección del host real, tareas solo en Claude Code y OpenCode `todowrite`, identidad explícita en OpenCode y Pi, sin `Stop`/`SubagentStop` en Codex). Sin ejecutar sesiones: `CODEX_THREAD_ID`, `CURSOR_CONVERSATION_ID` y los hooks en `codex exec` quedan sin verificar.

2026-11-04 — Plan escrito. Revisión `hive-review-plan`: 4 P1, 9 P2 y 6 P3, aplicados (decisiones D12 A, D13 A y D14 B).

## Checkpoint

- Etapa: cerrado. Integrado en #374 (dd729041); hallazgos H40–H43, H46 y H47 en #376 (9a1f3231); H44 en #375; H45 ignorado.
- Visor: trabajo `50457f3803d3122ef9c5294ee314a04b7caa03a46924c73bd3ea949a081b4a30`, etapa build `event-0031e0d2ade9120dcd2f535d5f41f8b3`, tareas registradas con la revisión `revision-69dbb00740f6bd76e426531cdab267a0` (la tarea T0 del plan es la clave `T11` en el visor, que no admite `T0`).
- Hallazgos abiertos (al cierre, no bloquean): latencia cuando el lock de la cola está ocupado; clave de proyecto desfasada entre dos pasos; lógica de proyecto duplicada; escritura a disco de cada evento; arranque del visor de 13 s con 53 trabajos; atribución de Cursor a Claude Code, plausible; base de UX (títulos de fila sin truncar, truncado del registro). Corregidos: 10 de la primera revisión, 5 de la segunda, 5 de UX, nombres de hooks en Codex y H39 (`hive view start`).
- Siguiente: ninguno; medir el nuevo costo de contexto con sesiones normales (rúbrica).
