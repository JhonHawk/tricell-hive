# Design

## Approaches

| Enfoque | Qué es | A favor | En contra |
| --- | --- | --- | --- |
| A. Ampliar E2E | Una variante de varios tickets en `flow-build/references/e2e-coordination.md` | Un solo hogar para el coordinador; registro, `validate-handoff` y reemplazo sin cambios | E2E es un objetivo que avanza etapas sin preguntar; mezclarlo con N tickets con paradas confunde a ambos |
| **B. Skill `swarm` aparte (elegido, D10 A)** | Un skill que activa el usuario y guarda todas las excepciones; usa las mecánicas de E2E por referencia | Los flujos normales no cambian y una prueba lo comprueba | Un skill más, con su tope de 16 KB |
| C. Flag en `## Hive` | `Swarm: enabled` por repositorio | Se activa una vez | Autorización permanente en lugar de explícita por lote |

B es el único que cumple a la vez «las excepciones solo viven en el modo» y «activación explícita». Sigue el patrón de E2E: un modo con su excepción encerrada en su procedimiento y una sola línea en la guía global.

## Decisions

Las del usuario y los supuestos están en `proposal.md` «Decisiones». Las de diseño que no le tocan:

- El skill no copia pasos de `flow-plan`, `flow-build`, `flow-verify` ni de las hojas: el owner los lee por enlace, y el skill nombra solo qué paso hace el owner, cuál el coordinador y cada elemento de la regla que la excepción cambia (regla de `AGENTS.md` para excepciones de un solo punto de uso).
- El registro de ejecución de E2E se reutiliza con una unidad por ticket y un registro por ticket bajo `_support/workspace/<run>/`, creado al aprobar el plan del ticket; `tasks.py` solo amplía los roles aceptados (H1 A).
- Specs en otro repositorio con `Delivery: direct-base`: la ruta del change folder del ticket en ese checkout principal es la única ruta extra asignada al owner (H2 B); el owner no hace commits ahí, y el coordinador hace su commit con `git commit --only`, como ya fija git-workflow paso 3. `validate-handoff` no ve esa ruta: el coordinador revisa su `git status` acotado a la carpeta en cada parada.
- Las líneas `Ruling:` no van al checkpoint, que `tasks.py checkpoint` reescribe: el owner las devuelve y el coordinador las escribe en `decisions.tsv`.

## Proposed shape

```
Activación: el usuario escribe el comando del modo con 2–4 tickets
Coordinador (hilo principal)
  ronda 0: veredicto por ticket + entrega + topes + worktrees  → crea worktrees, asigna puerto a UI
  lanza 1 hive-ticket-owner por ticket (≤ tope simultáneo)
Owner, corrida 1: lee ticket y código, escribe el plan en el change folder
  → PARADA 1: plan, decisiones a–f, efectos, reglas supuestas, rutas, hechos externos
Coordinador: rondas agrupadas; ticket a en progreso; registro del ticket; hive-review-plan solo si es riesgoso; efectos; dependencias en el worktree
Owner, corrida 2 (retomado o relanzado desde el worktree): build, checks, commits locales
  · lo reversible → línea Ruling en su retorno → decisions.tsv
  · bloqueo → termina y lo devuelve → el coordinador pregunta → relanza
  → PARADA 2: head exacto, comandos y resultados, evidencia, lo no verificado
Coordinador: validate-handoff, verificador independiente (+ hive-review-ux en UI),
  vuelve a correr las verificaciones, push/PR por ticket según la entrega, ticket al tracker
Tope alcanzado: no lanza más; cada ticket abierto queda con worktree, head y checkpoint
```

- **Archivos del skill:** `content/skills/swarm/SKILL.md` (activación, ronda 0, deberes del coordinador, topes, recuperación) y `references/owner-stops.md` (lo que el owner hace en cada corrida, los formatos de la parada 1, la parada 2 y el bloqueo, la línea `Ruling`, las excepciones de capturas y puertos).
- **Rol:** `content/agents/development/hive-ticket-owner.md`, `model_profile: reasoning`, `access_profile: implement`, con un bloque de límites propio fijado en `boundaryOverrides`.
- **Guía global:** la línea 24 de `content/guidance/global.md` pasa a nombrar los dos modos y sus roles, sin pasos.

## Impact

Lugares revisados (lista para el PR): `content/guidance/global.md` «Scope and authorization»; `content/skills/flow-build/references/e2e-coordination.md`; `content/skills/flow-plan/references/execution-record.md` y `scripts/tasks.py`; `content/skills/hive-core/delegation/delegation.md` y `child-brief.md`; `content/skills/git-workflow/SKILL.md` paso 3; `content/skills/hive-core/ui/before-captures.md`; `content/skills/hive-core/messages/questions.md`; `content/agents/development/hive-build-owner.md`; `tests/content/agent_boundary_test.go`, `single_home_test.go`, `reading_footprint_test.go`, `leaf_links_test.go`; `tooling/management/hidden_skills.go`; `content/skills/flow-plan/scripts/tasks.py`; `content/skills/audit-harness/references/skills-and-agents.md` (HA-SK-04); `content/skills/hive-core/tracker/ticket-state.md`; `site/src/content/docs/flows/overview.md`; `_support/openspec/specs/workflow-lifecycle/spec.md` («Bounded E2E coordination», «Local deliverable owner exception», «The user decides whether a worktree is created»); `site/src/content/docs/skills.md`; `llms.txt`; `_support/docs/architecture/instruction-resources.md` y `agent-delivery.md`; `_support/docs/harness-engineering/session-review-rubric.md`; detectores de `tests/pilot/`. Ninguno de los flujos normales cambia de texto; su conducta solo cambia dentro de una sesión que activó el modo.

## Recovery

Revertir el PR quita el skill, el rol, la línea global y las pruebas; ningún dato guardado cambia. Una sesión del modo cortada a mitad deja worktrees, ramas y checkpoints; la limpieza de cierre los protege mientras tengan commits sin integrar o archivos sin seguimiento.
