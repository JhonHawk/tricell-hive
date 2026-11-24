# gh-576-swarm-related-tickets — modo swarm: un owner por grupo de tickets relacionados

| Campo | Valor |
| --- | --- |
| Estado | Borrador · plan en revisión |
| Tracker | [#576](https://github.com/JhonHawk/tricell-hive/issues/576) · abierto · enhancement |
| Git | tricell-hive · base `development` en `15ac9aa5` · rama propuesta `feat/swarm-related-tickets` |
| Playbook | feature |

## Decisiones

2026-11-23
- **D20 A** — ¿Cómo se trabajan los tickets relacionados en el modo swarm? Un owner por grupo de tickets relacionados, con un worktree, una rama y un pull request por grupo; descartados serializar un owner por ticket y PRs apilados por ticket.
- **D21 A** — ¿Cómo decide el hilo principal que dos tickets están relacionados? Lo están cuando sus archivos estimados se cruzan, tocan el mismo módulo o uno depende del otro; el coordinador propone los grupos con su razón en la ronda 0 y el usuario los confirma o cambia.
- **D22 A** — ¿Cuánto cabe en un grupo? Hasta 3 tickets y 10 archivos estimados en total; un grupo mayor no califica y va a su propia sesión.
- **D23 A** — ¿Un change folder por grupo o por ticket? Uno por grupo, con los criterios de cada ticket marcados con su clave; un registro de ejecución con una unidad y un verificador por grupo.
- **D24 A** — ¿El límite de 2 a 4 cuenta tickets u owners? Tickets, como hoy.
- **D25 A** — ¿Cuentan como cruce los archivos que casi todo cambio toca (CHANGELOG, lockfiles, manifiestos, índices generados)? No: no agrupan tickets; un choque en ellos al integrar lo resuelve el coordinador al traer la base antes del merge, como ya hace merge gate.
- **D26 A** — ¿Qué pasa si todo el lote queda en un solo grupo? El coordinador recomienda trabajar esos tickets en la sesión con el flujo normal, y el usuario decide.
- **Supuesto** — Un grupo cuenta sus corridas, su puerto de UI, sus campos del checkpoint del lote y sus rondas como hoy los cuenta un ticket. Alternativa descartada: topes por ticket dentro del grupo. Se revierte en el skill.
- **Supuesto** — El change folder de un grupo empieza con la clave de su primer ticket y cada criterio se escribe `- AC<n> (<clave>) — …`, para que `tasks.py` y `check_change.py` lo sigan leyendo. Alternativa descartada: un slug sin clave. Se revierte en `owner-stops.md` y en el delta, que lo fija.
- **Supuesto** — Volver a verificar un ticket contra la base después de que otro se integra ya lo hace merge gate (trae la base y vuelve a correr las verificaciones antes de cada merge); este cambio no lo toca. Alternativa descartada: un paso propio del modo. Se revierte añadiéndolo.
- **Supuesto** — Un ticket solo es un grupo de uno: el modo deja de distinguir «owner por ticket» de «owner por grupo». Alternativa descartada: dos caminos. Se revierte separando el texto.
- **Supuesto** — Cada ticket de un grupo se mueve en el tracker por separado al integrarse el PR del grupo, y el PR nombra cada ticket. Alternativa descartada: mover solo el primero. Se revierte en una línea del skill.
- **Supuesto** — El bloque de límites del rol dice «the change folder of your assigned tickets» en lugar de «your ticket's change folder»; los permisos no cambian. Alternativa descartada: dejar el singular. Se revierte con el texto anterior y su prueba.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «H1, que el hilo principal detecte si un agente, puede tomar mas de un ticket relacionado»

H1 del cierre de `gh-576-swarm-mode` (P2): dos tickets del mismo repositorio pueden tocar los mismos archivos sin que nada lo detecte, y un ticket no se vuelve a verificar contra la base después de que otro se integra.

## Objective

Hoy el modo swarm da un owner a cada ticket, y dos tickets que tocan los mismos archivos chocan al integrarse. Con el cambio, la ronda 0 propone agrupar los tickets relacionados bajo un solo owner, que los planea y construye juntos en un worktree, de modo que dos owners nunca comparten un archivo estimado.

## Scope and acceptance

Dentro: el skill `swarm` y su referencia de paradas, el texto del rol `hive-ticket-owner` y sus pruebas, el delta de tres requisitos del modo, la página de skills y la vista general del sitio, `llms.txt`, `agent-delivery.md`, el CHANGELOG y la rúbrica de sesiones. Fuera: `tasks.py` (una unidad por grupo ya cabe), los flujos normales, `sbx`, y volver a verificar contra la base tras otra integración, que ya hace merge gate.

- AC1 — En la ronda 0 el coordinador propone grupos de tickets relacionados (archivos estimados que se cruzan, el mismo módulo o una dependencia), cada uno con su razón, y el usuario los confirma o cambia antes de crear worktrees · false today because el modo asigna un owner a cada ticket sin mirar si se relacionan · Observed in: la primera sesión real del modo con dos tickets relacionados
- AC2 — Un grupo tiene a lo sumo 3 tickets y 10 archivos estimados en total; uno mayor se nombra no apto con su razón y se deja para su propia sesión · false today because no existen grupos · Observed in: la primera sesión real del modo con dos tickets relacionados
- AC3 — Cada grupo tiene un owner, un worktree, una rama, un change folder con los criterios de cada ticket marcados con su clave, una unidad de registro con un verificador, y un pull request que nombra cada ticket; cada ticket se mueve en el tracker al integrarse · false today because cada ticket tiene su propio owner, rama y PR · Observed in: la primera sesión real del modo con dos tickets relacionados
- AC4 — Dos owners del mismo lote nunca comparten un archivo estimado, sin contar los que casi todo cambio toca: tickets cuyos archivos estimados se cruzan van al mismo grupo o no se toman juntos · false today because nada compara los archivos estimados entre tickets · Observed in: la primera sesión real del modo con dos tickets relacionados
- AC5 — El lote sigue contando de 2 a 4 tickets, agrupados o no · false today because no existen grupos, y con ellos el límite podría leerse por owner · Observed in: la primera sesión real del modo **Retired:** ya es cierto en la base; se cumple dentro de AC2
- AC6 — La página de skills del sitio y el CHANGELOG describen los grupos, y el sitio compila · false today because ninguno los menciona · Observed in: `pnpm --dir site check`, `lint` y `build` y la comprobación de texto sobre el candidato **Retired:** reemplazado por AC9
- AC7 — En la parada del plan, el coordinador compara las «Paths to write» de cada grupo con las rutas asignadas a los otros grupos y, si se cruzan fuera de los archivos comunes, detiene ambos grupos y te pregunta si reagrupar o apartar uno · false today because la parada solo revisa escrituras fuera de las rutas del propio owner · Observed in: la primera sesión real del modo con dos tickets relacionados **Retired:** reemplazado por AC10 (qué grupos compara y cuándo se detienen)
- AC8 — Con un solo grupo en el lote, la ronda 0 recomienda el flujo normal para esos tickets · false today because no existen grupos · Observed in: la primera sesión real del modo con un lote de un grupo
- AC9 — La página de skills y la vista general del sitio, `llms.txt`, `agent-delivery.md`, el CHANGELOG y la rúbrica de sesiones describen un owner por grupo, sin frases que digan un owner o un PR por ticket · **Replaces:** AC6 · false today because `llms.txt:58`, `flows/overview.md:27`, `agent-delivery.md:9,129` y la rúbrica (98, 101, 103) dicen por ticket · Observed in: la página de skills y la vista general renderizadas por `pnpm --dir site build`, y `rg` de esas frases sobre el candidato
- AC10 — En la parada del plan, el coordinador compara las «Paths to write» de cada grupo con las rutas asignadas de los otros grupos, o con sus estimaciones de la ronda 0 cuando aún no tienen, sin contar un grupo ya integrado; si se cruzan fuera de los archivos comunes, detiene el grupo que llega y el otro en su próximo retorno, y te pregunta si reagrupar o apartar uno · **Replaces:** AC7 · false today because la parada solo revisa escrituras fuera de las rutas del propio owner · Observed in: la primera sesión real del modo con dos tickets relacionados **Retired:** reemplazado por AC11 (planes de la misma ronda, bloqueos y grupos ya construidos)
- AC11 — En la parada del plan, y por cada ruta que añade un bloqueo del build, el coordinador compara las rutas a escribir del grupo con las de cada otro grupo no integrado (su plan devuelto, si no las rutas de su registro, si no sus estimaciones de la ronda 0) y repite la comparación con cada plan que llega después; si se cruzan fuera de los archivos comunes, detiene el grupo que llega, detiene al otro en su próximo retorno o retiene su publicación si ya pasó su build, y te pregunta si reagrupar o apartar uno · **Replaces:** AC10 · false today because la parada compara solo contra estimaciones y rutas asignadas, y un bloqueo del build no compara · Observed in: la primera sesión real del modo con dos tickets relacionados
- AC12 — Los tickets de un grupo comparten un repositorio; tickets relacionados de repositorios distintos se nombran como no tomables juntos con esa razón · false today because nada impide un grupo entre repositorios · Observed in: la primera sesión real del modo con tickets relacionados en dos repositorios

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/swarm/SKILL.md` en `15ac9aa5`: la ronda 0 da un veredicto por ticket y un owner por ticket; «Who writes and commits» y el resto de excepciones hablan de «a ticket».
- `content/skills/flow-plan/scripts/tasks.py`: una unidad lleva listas `tasks` y `criteria` y un solo `verifier` que cubre todos los criterios de la unidad, así que una unidad por grupo no necesita cambio de código.
- El árbol de `15ac9aa5` es idéntico al de `36ea7644` (`git diff --quiet`), donde la comprobación de texto pasó.

- `rg -n -e '(owner|worktree|pull request)[^.]{0,40}per ticket|per ticket[^.]{0,40}(owner|worktree|pull request)|each ticket on its own' llms.txt site/src/content/docs _support/docs/architecture/agent-delivery.md CHANGELOG.md _support/docs/harness-engineering/session-review-rubric.md` · sobre la base muestra `llms.txt:58`, `CHANGELOG.md:17`, `session-review-rubric.md:98,103`, `flows/overview.md:27`, `skills.md:53` (falso hoy, como se espera) · 2026-11-23
- Comprobación de texto de `AGENTS.md` «Checks by path» · pasa sobre el árbol de la base (`36ea7644`, idéntico a `15ac9aa5`) · 2026-11-23
- `go test -count=1 ./tests/content/...` · pasa sobre el mismo árbol · 2026-11-23
- `pnpm --dir site check`, `lint`, `build` · pasan sobre el mismo árbol · 2026-11-23

## Effects

None: the tasks take no effect on an external system.

## Delivery

`interactive` (respuesta en vigor de esta sesión y repositorio), con la revisión nativa de Claude Code (`/code-review`). El cambio no toca un límite de seguridad nuevo: el bloque de límites del rol cambia de redacción, no de permisos.

## Review status and progress

2026-11-24 — Revisión en build (`/code-review`): cuatro hallazgos bloqueantes corregidos (planes de la misma ronda, grupo ya construido, bloqueos del build, grupos entre repositorios) y la frase del delta; AC10 → AC11, AC12 nuevo; T1 cierra AC11 y AC12.
2026-11-23 — Relectura (~44 s, 139k tokens): seis P3; aplicados tal como los propuso el revisor, sin otra relectura: búsqueda de frases por ticket ampliada y medida, singular del delta, alcance de la comparación entre grupos (AC7 → AC10), aclaración de archivos disjuntos entre worktrees, MODIFIED del revisor de plan por grupo, y nota de que la forma de criterios también está en el delta.
2026-11-23 — Plan escrito. Revisión (`hive-review-plan`, ~143 s, 125k tokens): 6 P2 y 5 P3; D25 A y D26 A decididos; el resto aplicado: textos por ticket en `llms.txt`, vista general, `agent-delivery.md` y rúbrica; MODIFIED de la excepción del owner y del registro; tope, puerto, checkpoint y rondas por grupo; comprobación de rutas entre grupos en la parada del plan; nombre y marca de criterios; AC5 y AC6 retirados, AC7–AC9 nuevos; orden de archivo.
