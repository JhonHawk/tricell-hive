# Tasks

Base: `origin/development` en `da54bfec` (el `development` local está en `fbf3922e`, detrás). Rama `fix/session-identity-subagents` en un worktree bajo `.claude/worktrees/`. Todas las tareas corren en el hilo principal: son cambios chicos en archivos que se conocen, y ningún aspecto supera el umbral de delegación. T4 va antes que T3 porque una discrepancia detiene T3, y la ejecuta el hilo principal porque gasta cuota autorizada por D2 A.

Orden: T4 → T1 → T2 → T3 → T5.

### T1 — Grok: eventos de subagente con el patrón Codex

- [x] `SubagentStart` produce `subagent_started` con la sesión padre y el `subagentId`, y Hive lo registra en el hook propio de Grok; todo otro evento con `subagentType` no produce eventos; los eventos sin `subagentType` se traducen como hoy.

**Closes:** AC2

**Depends on:** none

**Locations:** `integrations/grok/hook_event.go`, `integrations/grok/hook_event_test.go`, `integrations/grok/hook_register.go`, `integrations/grok/hook_register_test.go` (y el test de reporte del hook que enumere los eventos registrados)

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- El traductor de Grok lee `subagentType` y `subagentId`, como en `design.md` «Proposed shape».
- `sessionEventNames` agrega `SubagentStart`, para que el evento llegue aunque los hooks de Claude Code no estén instalados o la compatibilidad de Grok con ellos esté apagada. Con la compatibilidad activa, el mismo evento llega dos veces con el mismo contenido y el visor aplica una sola (ID de operación derivado del contenido).
- Los tests usan los payloads capturados por la sonda (con `cwd` y rutas sustituidas por rutas de fixture): `SubagentStart` del padre, `UserPromptSubmit`, `SubagentStop` y `SessionEnd` del hijo, y `SessionStart`, `UserPromptSubmit` y `Stop` del padre.
- El comentario deja de decir que los subagentes comparten la sesión.
- No cambia `IsHookPayload` ni la detección de host.
- El hook desplegado se actualiza con `hive update` al refrescar la instalación después del merge.

**Verification:** `go test ./integrations/grok/...` pasa; los casos nuevos fallan antes del cambio.

### T2 — Regresión de punta a punta del mapa con un subagente de Grok

- [x] La secuencia real de hooks de un subagente de Grok deja a la sesión principal como única viva del mapa, y el resolver devuelve su identidad durante y después del subagente, por el hook de Grok y por la ruta de compatibilidad de Claude Code.

**Closes:** AC1

**Depends on:** T1

**Locations:** `tooling/cli/view_identity_map_test.go` (o un test nuevo junto a `processHook`)

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- Un test pasa por `processHook` los payloads de la sonda en orden, con un proceso host `grok` falso y un directorio de estado temporal: inicio del padre, prompt del padre, `SubagentStart`, prompt del hijo, `SubagentStop` y `SessionEnd` del hijo.
- Lo hace dos veces: con el host `grok` y con el host `claude` más `GROK_HOOK_EVENT` definido, que es como Grok ejecuta los hooks de `~/.claude/settings.json`.
- Después de cada paso, `resolveViewIdentity` devuelve `grok` con el id del padre, y el mapa nunca contiene el id del hijo.
- No cambia `hook_map.go` ni `apply()`.

**Verification:**
- `go test ./tooling/cli -run <nombre del test>` pasa.
- El test falla con el `hook_event.go` de la base. Para comprobarlo, se restaura temporalmente en el worktree con `git show origin/development:integrations/grok/hook_event.go` y después se vuelve a poner el cambio. Nunca con `git stash`: la lista de stash es compartida entre worktrees.

### T3 — Codex: `CODEX_SESSION_ID`, selección entre variables y vuelta al mapa

- [x] Una shell de Codex resuelve por `CODEX_SESSION_ID`. Con varias variables, gana la del host del proceso más cercano, reconocido también cuando corre bajo un intérprete (`node`, `bun`, `deno`). Un host más cercano sin variable entre las presentes resuelve por su propio mapa. `CODEX_THREAD_ID` sigue sin ser fuente.

**Closes:** AC3, AC4, AC6

**Depends on:** T4

**Locations:** `tooling/cli/view_identity_resolve.go`, `tooling/cli/view_identity.go` (comentario de la línea 13), `tooling/cli/view_identity_test.go`

**Execution:** main thread

**Test approach:** tdd

**Changes:**
- Agrega la variable de Codex a `viewIdentityVariables`, más la selección por host más cercano y la vuelta al mapa descritas en `design.md`.
- La selección y `confirmVariableHost` reconocen el proceso host igual que el escritor del mapa: nombre del comando más argumentos para intérpretes (`processHost`).
- Tests:
  - `CODEX_SESSION_ID` bajo un proceso `codex` cuyo mapa tiene otra sesión viva devuelve la variable.
  - `CLAUDE_CODE_SESSION_ID` y `CODEX_SESSION_ID` bajo `claude` anidado en `codex` devuelven la de Claude Code.
  - `PI_SESSION_ID` y `CLAUDE_CODE_SESSION_ID` bajo `node …/pi` anidado en `claude` devuelven la de Pi.
  - `CODEX_SESSION_ID` bajo un proceso `grok` con mapa de una sesión viva devuelve la sesión de Grok del mapa.
  - Dos variables sin proceso host conocido siguen siendo ambiguas.
- El test de `CODEX_THREAD_ID` queda igual.

**Verification:** `go test ./tooling/cli -run 'ResolveViewIdentity'` pasa; los casos nuevos fallan antes del cambio.

### T4 — Prueba en vivo de `CODEX_SESSION_ID` en Codex 0.167

- [x] Una ejecución real de Codex muestra que `CODEX_SESSION_ID` en su shell es igual al `session_id` que su hook escribió en el mapa de Hive y al id de hilo que reporta Codex.

**Closes:** AC3

**Depends on:** none

**Locations:** ninguno en el repositorio; fixture desechable en el temporal de la sesión

**Execution:** main thread (gasta una ejecución corta de cuota de Codex, autorizada por D2 A)

**Test approach:** check

**Changes:**
- En un repositorio desechable, `codex exec --json` ejecuta `printenv CODEX_SESSION_ID`.
- Se compara la salida con dos fuentes: el `session_id` del mapa `hooks/sessions/<pid>.json` que escribió el hook de Codex para ese proceso, y el id de hilo de los eventos JSONL. Si el hook no escribe mapa (confianza de hooks o sandbox), basta el id de los eventos y se anota la falta.
- Se registran la versión, el comando y el resultado en el progreso.
- No publica ninguna work: `list` y `show` no resuelven identidad, y `open` crearía una work.

**Verification:** los ids son iguales. Si difieren, T3 se detiene y vuelve como decisión.

### T5 — Spec y referencias de identidad

- [x] El delta del spec, `host-session-identity.md` y `host-session-events.md` describen la variable de Codex, la selección entre variables, la vuelta al mapa y los subagentes de Grok.

**Closes:** AC5

**Depends on:** T1, T3, T4

**Locations:** `_support/openspec/changes/session-identity-subagents/specs/walkthrough-viewer/spec.md`, `content/skills/flow-walkthrough/references/host-session-identity.md`, `_support/docs/architecture/host-session-events.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **`host-session-identity.md`:** la fila de Codex pasa a `CODEX_SESSION_ID` («code and observed», con la versión), y una línea explica la regla del host más cercano. La fila de Grok sigue en «none found»: Grok resuelve por el mapa.
- **`host-session-events.md`:** la fila de Grok describe los payloads de subagente. Cita la documentación de Grok 1.0.77 (`~/.grok/docs/user-guide/10-hooks.md:96`, `:106`, `:110`) como «doc» y la sonda como «obs», por separado. Se agregan `SubagentStart` en el registro de Grok, el daemon de Codex y la variable a «Variables versus the map».
- Revisión del comportamiento completo según `AGENTS.md`: guía global, skills, roles, `site/`, arquitectura y detectores de `tests/pilot/`.
- No cambian `viewer-publication.md` ni `site/src/content/docs/viewer.md`; se verifica que no lo necesitan.
- El mensaje de commit nombra la sesión que falló: Grok 1.0.77, `01a1aed2-ced4-79d2-bc52-4b55d46873bf`, 2026-11-06 UTC (2026-11-05 hora local).

**Verification:**
- `python3 <flow-plan-dir>/scripts/check_change.py _support/openspec/changes/session-identity-subagents` termina con 0.
- `go test ./tests/content/...` pasa (tamaño de referencias).
- `rg -n 'CODEX_THREAD_ID|share the parent' content _support/docs tooling integrations` solo muestra los usos que siguen siendo correctos.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, localmente | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, localmente | `go test ./...` y `go vet ./...` | Cubierto por la implementación |
| Prueba en vivo de Codex | T4, localmente | Ids iguales registrados en el progreso | D2 A |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` | D4 A |

No hay superficie de UI. Hay dos comprobaciones humanas opcionales, que no bloquean:
- En un hilo de la app de Codex, `printenv CODEX_SESSION_ID` debería coincidir con la sesión viva del mapa de su daemon.
- En tu próxima sesión real de Grok con subagentes, el visor debería seguir recibiendo las etapas.

## Review status and progress

2026-11-05 — Plan escrito.
2026-11-05 — Revisión `hive-review-plan` aplicada (ver `proposal.md`).
2026-11-06 — T4 hecha, sin commit propio porque no cambia el repositorio. Codex CLI 0.167.0 ejecutó `codex exec --json -s read-only` en un fixture desechable a las 06:00Z. `printenv CODEX_SESSION_ID`, el `thread_id` de `thread.started` y el `session_id` del mapa `hooks/sessions/10901.json` dieron el mismo valor: `01a1af71-85ed-72b1-bd93-713a1ccb12d0`. Queda sin verificar en un hilo del daemon `app-server` (comprobación humana opcional).

2026-11-06 — T1 hecha en `704c0f30`: traductor de Grok y `SubagentStart` en el hook registrado. Los tests nuevos fallaron antes del cambio.
2026-11-06 — T2 hecha en `29a4a6ea`: regresión por las rutas `grok` y `claude`. Falla con el `hook_event.go` de la base (restaurado temporalmente con `git show`, sin stash).
2026-11-06 — T3 hecha en `73ced80a`. Se cambió `TestResolveViewIdentityRejectsNestedHostThatInheritsTheOuterVariable` (ahora `…NestedHostNeverUsesTheOuterVariable`) porque AC6 cambia el resultado esperado: de «ambiguous» a resolver por el mapa propio, o «missing» si no hay mapa. Sigue comprobando que nunca se publica como la sesión externa. `findHostProcess` reutiliza `hostOfProcess`.

2026-11-06 — T5 hecha (ver commit `docs(identity)`). Comportamiento revisado en: `content/guidance/global.md` (no menciona la resolución), skills y referencias de `content/` (solo `host-session-identity.md` cambia; `viewer-publication.md` sigue correcta), `site/src/content/docs/viewer.md:49` (correcta), `_support/docs/architecture/deployment-manager.md:198` (correcta), `host-session-events.md` (actualizado) y `tests/pilot/` (ningún detector depende de esto). `check_change.py` terminó en 0 y `go test ./tests/content/...` pasó.

2026-11-06 — Verificación `hive-verify-task`: T1 cumple AC2; T3 cumple AC3, AC4 y AC6. Observación de T3 corregida en el mismo task: con dos variables y un host interno no reconocido (Pi bajo `node` con la línea de comandos ilegible), ganaba la variable del host externo. Ahora esa variable solo se usa si el host de cada otra variable aparece más arriba; si no, el resultado es «ambiguous». Se agregaron tests para ese caso y para un mapa no confiable en la ruta con variables (falló el primero antes del arreglo). Spec delta y `host-session-events.md` actualizados.
2026-11-06 — Suite completa en `370699ef`: `go vet ./...` sin salida y `go test ./...` sin fallos. El único workflow de CI (`master-source-branch.yml`) controla la rama de origen de `master` y no corre tests.

2026-11-06 — Verify sobre `23dd9d37`. Comprobaciones reutilizadas: `go vet ./...` y `go test ./...` sobre `370699ef`; después solo cambiaron `tooling/cli`, el spec y la documentación, y `go test ./tooling/cli` completo pasó con el código final. Recorrido por la ruta real con el binario candidato y estado temporal (D5 A):
- Desde la shell de Claude Code, sin flags: resuelve la sesión `6efcce79…`. Con `PI_SESSION_ID` o `CODEX_SESSION_ID` heredados y sin ese host en la ascendencia: «ambiguous».
- Grok 1.0.77, un turno con subagente; padre `01a1af85-9efc…`, hijo `01a1af85-daf4…`. El mapa solo tiene al padre. El `open` ejecutado después del subagente queda a nombre del padre. Los 6 eventos encolados son del padre, y el hijo solo aparece como `agent_id` de `subagent_started`. Grok se lanzó desde Claude Code con `CLAUDE_CODE_SESSION_ID` heredado y aun así resolvió su propio mapa (AC1, AC2, AC6).
- Codex 0.167.0, `codex exec -s workspace-write`. Primera corrida lanzada desde Claude Code: «ambiguous» correcto, porque había dos variables y el sandbox no deja ejecutar `ps` (`seatbelt_base_policy.sbpl:16`: process-info solo same-sandbox), así que no hay ascendencia visible. Segunda corrida sin la variable heredada: el hilo `01a1af8c…` abre la work a su nombre (AC3).
- Sandbox local (`sandbox-exec`): la base da «required semantic field missing» con `CODEX_SESSION_ID`; el candidato resuelve. Probable causa del aviso `watch/viewer-identity-unresolved-skips-publication` (Codex, 2026-11-04).

2026-11-06 — Revisión `/code-review` sobre `23dd9d37`: 9 hallazgos.
- Corregidos en esta ronda:
  - **H12** — Bloqueante. En la vuelta al mapa no se comprobaba de dónde venían las variables: un host interno no reconocido publicaba como la sesión del mapa del host externo. Ahora la regla de origen aplica a las dos ramas.
  - **H13** — `hostsAbove` solo corre cuando hay otras variables y se detiene al encontrarlas.
  - **H14** — El mapa del host más cercano ya no se carga dos veces.
  - **H15** — `deployment-manager.md` lista `SubagentStart`.
  - **H16** — `subagentId` también se lee como `subagent_id`.
- A los hallazgos del cierre:
  - **H17** — Codex sin `CODEX_SESSION_ID` (antes de 0.167) sigue resolviendo por el mapa al último hilo. Ya pasaba en la base.
  - **H18** — Una variable de Codex heredada y vencida, por ejemplo en un servidor de tmux, se acepta sin ascendencia. Es el mismo riesgo que ya tienen las otras variables.
- Reportado sin corregir:
  - **H19** — Plausible, sin confirmar: un hook de compatibilidad sin `GROK_HOOK_EVENT` mandaría los eventos del hijo al traductor de Claude Code. En la sonda, `$GROK_HOOK_EVENT` estuvo definido en todos los eventos.
- Sin acción:
  - **H20** — Grok no reporta `subagent_stopped`; es lo que se decidió en D1 A.
- Pruebas: los dos casos nuevos de H12 fallaron antes del arreglo. El caso snake_case de H16 se escribió después del cambio, así que no se vio fallar.

**Checkpoint:** verify · ronda de correcciones hecha · siguiente: segunda revisión limitada al diff de las correcciones, suite completa, push, PR y merge (D3 A).
