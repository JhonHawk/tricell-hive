# Tasks

Base: `development` en `15ac9aa5`; rama propuesta `feat/swarm-related-tickets`. El delta modifica requisitos que solo existen en el delta de `gh-576-swarm-mode`: archivar `gh-576-swarm-mode` antes que este. T1 y T2 corren en el hilo principal: texto de reglas que depende de las decisiones de esta sesión. T3 va a un hijo `hive-write-docs`.

### T1 — Grupos en el skill y en las paradas del owner

- [ ] La ronda 0 propone grupos de tickets relacionados con su razón y su tamaño, el usuario los confirma, y todo lo que hoy es «por ticket» (owner, worktree, rama, change folder, registro, verificador, PR) pasa a ser por grupo; el owner planea y construye los tickets de su grupo en un plan.

**Closes:** AC1, AC2, AC3, AC4, AC8, AC10

**Depends on:** none

**Locations:** `content/skills/swarm/SKILL.md`, `content/skills/swarm/references/owner-stops.md`.

**Execution:** hilo principal.

**Test approach:** check

**Changes:** «Round 0» define relacionados (archivos estimados que se cruzan sin contar los comunes como CHANGELOG, lockfiles, manifiestos e índices generados, el mismo módulo, una dependencia), el tamaño del grupo (≤ 3 tickets, ≤ 10 archivos estimados en total), que tickets cuyos archivos se cruzan van al mismo grupo o no se toman juntos, que el lote cuenta tickets, y que con un solo grupo se recomienda el flujo normal. El resto del skill dice grupo donde dice ticket para owner, worktree, rama, change folder (con la clave de su primer ticket), registro, verificador, PR, corridas y su tope, puerto de UI, campos del checkpoint del lote, rondas de decisiones, `flow-verify`, `flow-close` y la `description`; cada ticket se mueve en el tracker por separado. La parada del plan compara las «Paths to write» de cada grupo con las rutas asignadas de los otros grupos, o con sus estimaciones de la ronda 0 cuando aún no tienen, sin contar un grupo ya integrado; si se cruzan fuera de los archivos comunes, detiene el grupo que llega y el otro en su próximo retorno, como bloqueo, y pregunta si reagrupar o apartar uno. Donde el skill cita la regla de delegación para lanzar owners juntos, aclara que los archivos son disjuntos entre worktrees y que los comunes se resuelven en merge gate; un choque en un archivo común lo resuelve el coordinador al traer la base antes del merge. `owner-stops.md` escribe cada criterio como `- AC<n> (<clave>) — …` y construye los tickets del grupo en orden de dependencia. Un ticket solo es un grupo de uno. No cambia ninguna excepción ni permiso.

**Verification:** la comprobación de texto de `AGENTS.md` «Checks by path» pasa; AC1–AC4, AC8 y AC10 los alcanza la primera sesión real del modo.

### T2 — El rol habla de su grupo

- [ ] La descripción, los pasos y el bloque de límites de `hive-ticket-owner` nombran el grupo de tickets asignado; la prueba de límites y la de hogar único siguen fijando ese bloque.

**Closes:** AC3

**Depends on:** T1

**Locations:** `content/agents/development/hive-ticket-owner.md`, `tests/content/agent_boundary_test.go`, `tests/content/single_home_test.go`.

**Execution:** hilo principal.

**Test approach:** check — se actualiza la constante del bloque y su frase fijada; los permisos no cambian.

**Changes:** «your ticket's change folder» pasa a «the change folder of your assigned tickets» en el bloque y en su constante; ningún permiso cambia.

**Verification:** la comprobación de texto pasa, y `go test -count=1 -run 'TestAgentBoundaryBlock|TestRulesHaveOneHome' ./tests/content/` pasa.

### T3 — Documentación

- [ ] La página de skills y la vista general del sitio, `llms.txt`, `agent-delivery.md` y el CHANGELOG describen un owner por grupo; la rúbrica de sesiones reescribe sus filas del modo por grupo y gana una que comprueba la propuesta de grupos y la comparación de rutas en la parada del plan.

**Closes:** AC9

**Depends on:** T1

**Locations:** `site/src/content/docs/skills.md`, `site/src/content/docs/flows/overview.md`, `llms.txt`, `CHANGELOG.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/session-review-rubric.md`.

**Execution:** hijo `hive-write-docs`: seis archivos con fuentes que T1 y T2 fijan.

**Test approach:** check

**Changes:** describen la conducta y enlazan el skill sin copiar sus reglas.

**Verification:** `pnpm --dir site check`, `pnpm --dir site lint` y `pnpm --dir site build` pasan; la comprobación de texto pasa; `rg -n -e '(owner|worktree|pull request)[^.]{0,40}per ticket|per ticket[^.]{0,40}(owner|worktree|pull request)|each ticket on its own' llms.txt site/src/content/docs _support/docs/architecture/agent-delivery.md CHANGELOG.md _support/docs/harness-engineering/session-review-rubric.md` no muestra frases de un owner, un worktree o un PR por ticket.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the validation stop | Native: Claude Code `/code-review` | The review answer |
| Real session | After the integration and the local refresh | The first real swarm session with two related tickets, reviewed with the session rubric | The user's own session |

Validación humana (`interactive`): antes del push, el usuario lee el diff de `content/skills/swarm/SKILL.md` y `references/owner-stops.md`. AC1–AC4, AC8 y AC10 quedan sin verificar hasta la primera sesión real del modo con dos tickets relacionados.

## Checkpoint

- Updated: 2026-11-24
- Stage: build · start
- Last commit: 15ac9aa5
- Open findings: none
- Next: T1 groups in swarm skill and owner stops
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no screen changed)

## Review status and progress
