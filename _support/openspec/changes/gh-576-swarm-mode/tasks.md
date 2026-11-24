# Tasks

Base: `development` en `d5d9ba65`; rama propuesta `feat/swarm-mode`. T1–T3 y T5 corren en el hilo principal: son texto de reglas, pruebas y una validación del registro que dependen de las decisiones de esta sesión. T4 va a un hijo `hive-write-docs`: edita siete archivos con fuentes ya fijadas por T1–T3 y T5.

### T1 — Skill `swarm`

- [x] Existe `content/skills/swarm/` con su `SKILL.md` (activación, ronda 0 y veredicto por ticket, deberes del coordinador, topes, recuperación) y `references/owner-stops.md` (corridas del owner, formatos de parada 1, parada 2 y bloqueo, líneas `Ruling:`, excepciones de capturas previas, puertos y ruta del change folder), más `assets/hive-managed.json`; el skill no se invoca solo; HA-SK-04 de `audit-harness` ya no lo marcaría.

**Closes:** AC2, AC5, AC6, AC8, AC11, AC12

**Depends on:** none

**Locations:** `content/skills/swarm/SKILL.md` (nuevo), `content/skills/swarm/references/owner-stops.md` (nuevo), `content/skills/swarm/assets/hive-managed.json` (nuevo), `content/skills/audit-harness/references/skills-and-agents.md` (línea 10, HA-SK-04). Antes de escribir, leer `_support/docs/architecture/instruction-resources.md` «Authoring», como pide `AGENTS.md` para un skill nuevo.

**Execution:** hilo principal.

**Test approach:** check

**Changes:**
- Frontmatter con `disable-model-invocation: true` y una `description` que limita el uso al pedido explícito del modo.
- Enlaza, sin copiar pasos, `e2e-coordination.md` (registro, `validate-handoff`, reemplazo, recursos exclusivos), `execution-record.md`, `questions.md`, `delivery-ask.md`, `external-effects.md`, `reviewer-brief.md`, `before-captures.md`, `ticket-state.md`, el log de decisiones de `unattended-delegation`, `git-workflow` paso 3 y la lista de ticket listo de `create-issue`.
- Cada excepción nombra todos los elementos de la regla que cambia, y por nombre las líneas que contradice fuera del modo: `e2e-coordination.md:13` (solo `hive-build-owner`), `:17` (el coordinador escribe el plan; el change folder nunca en un worktree de owner), `:21`; `child-brief.md:13` (solo el owner de E2E hace commits); `before-captures.md:5` (el brief de un hijo no pide capturas); la revisión de plan para cada change folder; la creación de worktrees solo para la obra en curso; `global.md:23` (los hijos se quedan en su working tree), que la nueva línea 24 matiza para la única ruta de change folder asignada.
- El owner hace en línea lo que un flujo normal delegaría (aspectos de research, hechos externos como comando con extracto); el coordinador lanza los revisores y verificadores.
- Deberes del coordinador: mover cada ticket a en progreso antes de su build; instalar dependencias en cada worktree antes de servirlo o construir; crear el registro de cada ticket al aprobar su plan; escribir cada `Ruling:` devuelto en `decisions.tsv`; hacer el commit del change folder escrito en otro repositorio.
- El modo no es delegación desatendida y el silencio del usuario no lo extiende.
- HA-SK-04 pasa a eximir «los skills ocultos de Hive, que el manager acompaña de una política de Codex generada», sin nombrar ninguno.

**Verification:** la comprobación de texto de `AGENTS.md` «Checks by path» pasa, y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa (fila de `content/skills/**/assets/**`).

### T2 — Rol `hive-ticket-owner`

- [x] Existe `content/agents/development/hive-ticket-owner.md` (`model_profile: reasoning`, `access_profile: implement`) con sus dos corridas por enlace a `owner-stops.md`, y termina con un bloque de límites propio que la prueba de límites fija y que solo él lleva.

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/agents/development/hive-ticket-owner.md` (nuevo), `tests/content/agent_boundary_test.go` (entrada en `boundaryOverrides` y su bloque de escritura), `tests/content/single_home_test.go` (frase del permiso de commits del rol con su hogar en `owner-stops.md`), `tests/content/leaf_links_test.go` (solo si el rol enlaza el índice de principios como `hive-build-owner`).

**Execution:** hilo principal.

**Test approach:** check — la prueba de límites ya existe; se extiende su tabla con el rol nuevo.

**Changes:** `hive-build-owner.md` no cambia. El bloque del rol permite escribir su change folder y su checkpoint en su worktree o en la única ruta asignada del checkout de Specs, y hacer commits con el helper de commit con ámbito solo en su worktree, rama y rutas asignadas; prohíbe commits en el repositorio de Specs, push, PRs, tracker, memoria persistente, lanzar hijos, instalar dependencias y crear o cambiar worktrees o ramas.

**Verification:** la comprobación de texto pasa, y `go test -count=1 -run 'TestAgentBoundaryBlock|TestRulesHaveOneHome' ./tests/content/` falla si se borra la entrada del rol en `boundaryOverrides`.

### T3 — Línea global y prueba de alcance del modo

- [x] `content/guidance/global.md:24` nombra E2E y el modo swarm con sus roles, enlaza `skill:swarm/SKILL.md` para cuando el usuario nombra el modo y nombra la única ruta de change folder asignada como el solo lugar donde un owner de ticket escribe fuera de su worktree, y una prueba de contenido nueva falla cuando un archivo distribuido fuera del skill, el rol, esa línea y `tasks.py` nombra el modo o el rol.

**Closes:** AC1, AC2, AC4

**Depends on:** T1, T2, T5

**Locations:** `content/guidance/global.md` (línea 24), `tests/content/swarm_scope_test.go` (nuevo).

**Execution:** hilo principal.

**Test approach:** tdd — la prueba fija una regla de alcance (permisos); se escribe primero, se la ve fallar con una mención temporal en `content/skills/flow-build/SKILL.md` y pasar al quitarla.

**Changes:** la línea global sigue siendo un invariante sin pasos y no crece más de 400 bytes; `content/guidance/global.md` queda bajo sus 30,000 bytes. La prueba lee el texto distribuido, que es lo que el proyecto entrega.

**Verification:** `go test -count=1 ./tests/content/...` y la comprobación de texto pasan; con la mención temporal, `go test -count=1 -run TestSwarm ./tests/content/` falla y nombra el archivo.

### T4 — Documentación y rúbrica

- [x] La página de skills y la vista general de flujos del sitio, `llms.txt`, `CHANGELOG.md`, `_support/docs/architecture/instruction-resources.md` y `agent-delivery.md` nombran el modo y el rol; la rúbrica de sesiones gana las filas que auditan una sesión del modo (ronda 0, paradas, topes, puertos, `Ruling` en el log y verificación por ticket).

**Closes:** AC9

**Depends on:** T1, T2, T3, T5

**Locations:** `site/src/content/docs/skills.md`, `site/src/content/docs/flows/overview.md`, `llms.txt`, `CHANGELOG.md`, `_support/docs/architecture/instruction-resources.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/session-review-rubric.md`.

**Execution:** hijo `hive-write-docs`.

**Test approach:** check

**Changes:** las páginas describen la conducta y enlazan el skill sin copiar sus reglas. `README.md` no cambia: no lista skills. Las filas de la rúbrica llevan fecha y el cambio que las motiva.

**Verification:** `pnpm check`, `pnpm lint` y `pnpm build` en `site/` pasan; la comprobación de texto pasa.

### T5 — El registro de ejecución acepta al owner de ticket

- [x] `tasks.py` acepta una unidad con rol `hive-ticket-owner` y un registro en el `_support/workspace` ignorado del checkout del coordinador aunque el change folder esté en otro repositorio o worktree; sigue rechazando otros roles y ubicaciones inseguras; `validate-handoff` comprueba esa unidad como a un owner de E2E.

**Closes:** AC10

**Depends on:** none

**Locations:** `content/skills/flow-plan/scripts/tasks.py` (línea 622, la constante de roles y `execution_location`, líneas 473–490), `tests/skills/tasks_test.py`, `content/skills/flow-plan/references/execution-record.md` (solo la línea que fija la ubicación del registro, si la nombra).

**Execution:** hilo principal.

**Test approach:** tdd — cambia validaciones de permisos y de ubicación del registro: primero fallan los casos que aceptan `hive-ticket-owner` y un registro en el checkout del coordinador con el change folder en otro repositorio; luego pasan; siguen rechazados un rol cualquiera, un registro rastreado, uno no ignorado y uno fuera de `_support/workspace`.

**Changes:** la lista de roles aceptados y la regla de ubicación, que pasa a admitir el `_support/workspace` del repositorio desde donde se corre el comando cuando el change folder está en otro repositorio o worktree, con las mismas exigencias (ignorado, sin seguimiento, sin enlaces simbólicos que escapen); ningún otro campo del registro ni de `validate-handoff` cambia. Revisar también el bloqueo por deriva del plan (`execution-record.md:61`) frente a ediciones del owner tras crear el registro, y anotarlo en la recuperación de T1.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the validation stop | Native: Claude Code `/code-review` | The review answer |
| Security review | Final candidate, before the validation stop | `hive-review-security` | Seguridad A |
| Installed skill | After the integration and the maintainer refresh | `~/.agents/skills/swarm/agents/openai.yaml`, `~/.agents/skills/swarm/SKILL.md` y el enlace `~/.claude/skills/swarm` | Maintainer refresh rule of `AGENTS.md` |
| Real session | After the refresh | The first real swarm session, reviewed with the session rubric | The user's own session |

Validación humana (`interactive`): antes del push, el usuario lee `content/skills/swarm/SKILL.md`, `references/owner-stops.md`, el rol y la línea global. AC5, AC6, AC8, AC11 y AC12 quedan sin verificar hasta la primera sesión real del modo; AC2, hasta el refresco.

## Checkpoint

- Updated: 2026-11-23
- Stage: verify · validation stop
- Last commit: f8aa2d8f
- Open findings: none
- Next: user validation; code review and security review running
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- before captures: none (no screen changed)
- T5 hive-verify-task a9fbbdc7: AC10 met; follow-up fix 5274596a (cross-checkout pointer, symlink test)
- text check · pass · e4ad1eb7
- python3 -m unittest discover -s tests/skills -p '*_test.py' · pass 416 · 5274596a
- hive-verify-task T1 e4ad1eb7: 4 text gaps fixed in 53d9d34b
- hive-verify-task T2 aa3d1133: AC4 met
- hive-verify-task T3 c364d687: AC1 met, AC4 met (T3 part), AC2 pending refresh
- text check · pass · f8aa2d8f
- pnpm site check/lint/build · pass · f8aa2d8f (T4 child)
- full suite (go vet ./..., go test ./..., python unittest) · pass · f8aa2d8f · 254 s
- local CI: only workflow is master-source-branch (PRs to master), nothing to reproduce for development

## Review status and progress

2026-11-23 — T5 · a9fbbdc7 · unverified: hive-verify-task pending
2026-11-23 — T1 · e4ad1eb7 · unverified: AC2 hasta el refresco; AC3, AC5–AC8 hasta la primera sesión real; hive-verify-task pendiente
2026-11-23 — T2 · aa3d1133 · unverified: none
2026-11-23 — T3 · c364d687 · unverified: AC2 instalación tras el refresco
2026-11-23 — T4 · f8aa2d8f · unverified: none
