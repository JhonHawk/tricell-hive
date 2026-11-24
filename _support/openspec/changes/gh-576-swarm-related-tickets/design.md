# Design

## Approaches

| Enfoque | Qué hace | A favor | En contra |
| --- | --- | --- | --- |
| **A. Grupo con un owner (elegido, D20 A)** | La ronda 0 propone grupos de tickets relacionados; un owner por grupo, un worktree, una rama y un PR por grupo | Sin colisiones entre owners por construcción; el owner ve la intención de todo el grupo | PR más grande; los tickets del grupo se integran juntos |
| B. Serializar | Un owner por ticket; el relacionado espera a que el otro se integre | PR por ticket | Más lento; dos owners sobre el mismo código |
| C. Grupo con PRs apilados | Un owner por grupo, un PR por ticket en cadena | PRs chicos | Rebases y force-push que hoy necesitan autorización aparte |

## Decisions

Las decisiones están en `proposal.md` «Decisiones». De diseño, sin pregunta: un ticket solo es un grupo de uno, así el skill tiene un solo camino; `tasks.py` no cambia, porque una unidad ya lleva listas de tareas y criterios y un verificador que cubre todos los criterios de la unidad.

## Impact

Lugares revisados: `content/skills/swarm/SKILL.md` y `references/owner-stops.md` (todo lo que dice «ticket» para owner, worktree, rama, change folder, registro, verificador, PR y topes); `content/agents/development/hive-ticket-owner.md`; `tests/content/agent_boundary_test.go` y `single_home_test.go`; `content/skills/flow-plan/scripts/tasks.py` (`validate_verifier`, campos de unidad); el delta pendiente de `gh-576-swarm-mode`; `site/src/content/docs/skills.md`; `CHANGELOG.md`; la rúbrica de sesiones. Los flujos normales no cambian, y la prueba de alcance lo sigue comprobando.

## Recovery

Revertir el PR devuelve el modo a un owner por ticket; nada guardado cambia.
