# Diseño · gh-110-change-impact

## Contexto verificado

Base: `development` en `ad6099c1` (2026-10-23).

**Hechos inspeccionados en las sesiones** (transcripts en `~/.claude/projects/-Users-jmartinez-Development-projects-visooal-facti/`):

- `42b8e06a`, 09:06 UTC: el informe de investigación recomendó como R1 «Cerrar 3306 y 22 del MySQL». El usuario respondió en D1: «si el legacy pierde comunicación con él, se cae». La configuración del repositorio del API ya apuntaba a la IP pública de ese MySQL. El agente leyó entonces `flow-research`, cuyo paso de inspección solo cubre el repositorio.
- Misma sesión: el usuario pidió después ampliar la ventana de captura a 24 h, mirar si la base seguía recibiendo escrituras, «una forma rápida de revertirla» y «riesgo 0» antes de desplegar. Los revisores del plan (backend/datos e infraestructura) sí señalaron impactos sobre el portal moderno y la reversión. El code-review y la revisión previa al despliegue hallaron que corregir la recuperación de contraseña reactivaba código muerto y que el despliegue ampliaba el acceso a `admin.*`.
- `ccbed8a5`, 07:19 UTC: el agente informó «No pude descartar llamadores externos … porque no hay logs» y aun así propuso retirar. El usuario pidió un punto de reversión y buscar en otros repositorios.
- `c574e5a1`, 06:13 UTC: el usuario corrigió: «hay un server que solo se dedica al cron y uno al api».
- Las tres sesiones son del mismo usuario y proyecto. No se halló el patrón en otros proyectos.

**Hechos inspeccionados en Hive:**

- `content/skills/flow-research/SKILL.md:16` pide «entry points, callers, tests, configuration, and project guidance», todo dentro del repositorio.
- `content/skills/flow-plan/SKILL.md:22` descarta «A consumer that no evidence shows exists» y «a gate that can only pass by proving an absence, such as a period with no traffic from clients nobody has identified: no observation settles it, so the transition never retires». Nació de dos fallos reales de sobreingeniería (commit `3db867ef`, 2026-09-26, y el caso ARK-604, commit `c0069831`).
- `content/skills/flow-build/SKILL.md:42` repite esa regla al reconciliar transiciones.
- `content/skills/flow-plan/references/plan-review.md:15,17` lista productores/consumidores (backend) y orden de despliegue y recuperación (infra), pero nada sobre consumidores fuera del código.
- `content/agents/review/hive-review-code.md:14` exige que «the change introduced it, not code that was already there».
- `content/guidance/global.md:31` arregla dentro del trabajo los «defects that affect the current work».
- Ninguna capability spec de `_support/openspec/specs/` describe la guía distribuida, así que no hay delta.

**Evidencia externa (afirmaciones de fuentes, no medidas aquí):**

- La práctica para legacy coincide en tres pasos: mapa de efectos (Feathers), análisis dinámico además del estático (Thoughtworks, 2026-03-02) y cambio paralelo con el retiro al final (Fowler, Parallel Change).
- Ningún estudio controlado prueba que un paso dedicado de impacto reduzca regresiones. Añadir contexto genérico aumentó el costo más de 20% sin mejorar el éxito (Gloaguen et al., arXiv 2602.11988).
- Una instrucción concreta de buscar cada llamador mejoró los arreglos en la causa raíz en modelos capaces y no en uno pequeño (benchmark público de un proyecto de skills comparable, 2026-06-22). Por eso la referencia usa pasos operativos con su señal, no una exhortación a «considerar el impacto».

**Supuesto:** las señales elegidas cubren las tres sesiones. No hay medición de que los modelos del perfil `execution` las sigan; con las pruebas piloto en pausa, la validación será observar las próximas sesiones sobre legacy.

## Diseño propuesto

### Hogar y forma

`flow-research` es dueño del método de investigación y ya aloja referencias que otras skills consumen (`delegation.md`). La referencia nueva vive en `content/skills/flow-research/references/change-impact.md`. `flow-plan`, `flow-build` y `plan-review.md` la enlazan con su condición de lectura; los roles la reciben por el brief, como cualquier referencia (sin edición de `hive-research` ni `hive-review-plan`).

Proporcionalidad (restricción del usuario): el disparador decide si se lee la referencia; dentro, cada paso tiene su propia señal y solo corre cuando se observa. El paso 1 es el único que corre siempre bajo la referencia, porque es barato y fue el que habría resuelto D1.

### Texto propuesto de la referencia (en inglés, como el contenido distribuido)

```markdown
# Change impact

## When to read

Read before recommending, planning, or building either of these, unless it is a mechanical change as the global guidance defines it:

- a change that removes, closes, disables, migrates, or renames something that runs in production, such as a dependency, a network rule or port, an endpoint, a scheduled job, or a table or collection, or that moves it to another host or process;
- a behavior change in code that runs in production when no test or CI run covers that behavior.

Reading it does not invoke planning or authorize any effect.

## Scale the check to its signals

Always run step 1. Run each later step only when you observe its signal, and state in one line which signals you found. A change that shows none needs no further step.

1. **Search for the value, not only the symbol.** Search for the literal value the change touches: the host name or IP address, port, connection-string key, table or collection name, environment variable, route, or job name. Include configuration, deployment scripts, container and process-manager files, crontabs, and infrastructure code. Search every repository of the declared workspace when the value is a resource reached over the network, such as a host, an IP address, a port, or a database; this repository is enough for a value only it uses, such as an internal route or a local key. Each hit is a consumer to account for.
2. **Signal: a consumer runs outside this repository's release, or the system runs on several hosts or processes.** Map which host or process runs what, such as an API server and a separate cron server, and which of them read and which write the data the change touches. Use read-only sources: infrastructure state, process lists, configuration, and project guidance.
3. **Signal: the change touches authentication, permissions, or account data; makes reachable code that was disabled, commented out, or unreachable; or creates or migrates records that existing checks read.** Read the code paths the change makes reachable and state what they allow once it ships. List a defect in such a path in the Impact list below as part of this change, since the change makes it reachable.
4. **Signal: the repository cannot rule out a consumer, such as an external client, another server, or a periodic job.** Mark that consumer unverified. Name the read-only observation that would settle it, such as connection or access logs or database counters, a window long enough to cover the system's known periodic jobs, and who can run it, and state how to revert the change. Until that observation runs, do not describe the change as safe or quick.

## Report the impact

Where the finding or plan states the recommendation, add a short Impact list: each consumer or path found, its evidence, what the change does to it, and what remains unverified. In a plan it goes in the verified context, and each unverified item becomes a task or a blocker. When you delegate part of the investigation, pass this reference and name the change in the brief, so the child looks for consumers instead of only answering a narrower question.
```

### Reconciliación de la regla de interfaces (P2)

En `flow-plan/SKILL.md:22`, tras «…so the transition never retires.», añadir:

> Before applying this rule, search for consumers as [change impact](../flow-research/references/change-impact.md) sets, since a consumer nobody looked for is not yet one that no evidence shows exists. When the component is consumed outside the change's release control and that search cannot settle whether a consumer exists, a bounded read-only observation of its real traffic may inform the choice before the plan makes it. That observation decides between one coordinated change and a transition; it never becomes the condition that retires a transition. When it shows no consumer, plan the coordinated change.

En `flow-build/SKILL.md:42`, «or on a gate that can pass only by proving an absence, stop and ask» pasa a «or on a gate that can pass only by proving an absence, which the interface rule in [flow-plan](../flow-plan/SKILL.md) distinguishes from an observation that only chose the approach, stop and ask». Una observación ya usada para elegir no frena el build.

En `flow-build/references/data-changes.md:7`, junto a «Identify readers and writers», añadir «, searching for them as [change impact](../../flow-research/references/change-impact.md) sets when it applies».

La intención original se conserva: un consumidor hipotético no justifica una transición y ninguna transición se retira por demostrar una ausencia. Lo nuevo es que buscar consumidores es un paso previo, y que una observación acotada solo sirve para elegir el enfoque antes de planear.

### Punteros (P3)

- `flow-research/SKILL.md`, en «Frame the investigation», después del párrafo de inspección (línea 16): «Before recommending a change that the "When to read" section of [change impact](references/change-impact.md) covers, read that reference.»
- `flow-plan/SKILL.md`, junto a las lecturas de seguridad y datos (línea 24), un puntero breve que no repite el disparador: «Before a change that the "When to read" section of [change impact](../flow-research/references/change-impact.md) covers, read that reference.»
- `flow-build/SKILL.md:16`, en «Reconcile before execution», dos añadidos: «For work without a plan, read [change impact](../flow-research/references/change-impact.md) before a change its "When to read" section covers.», y que un consumidor o camino alcanzable que esa referencia encuentra y el plan no contempla («does not account for») invalida alcance o contrato y vuelve al plan con el procedimiento de revisión existente.
- `plan-review.md`, filas Backend (línea 15) e Infraestructura (línea 17): ambas añaden «the plan's Impact list against [change impact](../../flow-research/references/change-impact.md) when it applies»; la de Infraestructura añade además «hosts and processes that consume what the change closes or moves».

### Code-review (P4)

En `hive-review-code.md:14`, sustituir «the change introduced it, not code that was already there» por «the change introduced it, or makes reachable or widens access to code that was already there, such as re-enabling a disabled path; not code that stays as reachable as before».

En `flow-build/references/verification.md:43`, tras «Dispose of pre-existing findings under the shared incidental-findings rule», añadir «; a finding the reviewer reports as made reachable by the change belongs to the change, not to that rule». El texto exacto de esa línea se confirma al editar.

### Sitio y changelog

- `research.md`, en «What it produces», un párrafo: cuando la respuesta recomienda cambiar algo que corre en producción o código legacy sin pruebas, el agente revisa antes qué depende de ello, en la medida que indican las señales, y marca como no verificado lo que el código no muestra, con cómo comprobarlo. Enlaza la referencia.
- `plan.md`: una frase equivalente donde describe el contenido del plan.
- `CHANGELOG.md`, `[Unreleased]` → `### Changed`: una entrada con `#110`. Es un cambio de contenido instalado (versión de parche).

### Lugares revisados (para el PR)

`content/guidance/global.md:31,59,73`; `content/skills/flow-research/SKILL.md`; `content/skills/flow-plan/SKILL.md:22,24,26`; `content/skills/flow-plan/references/plan-review.md`, `plan-format.md:79-81`; `content/skills/flow-build/SKILL.md:7,14,16`, `references/verification.md`, `references/data-changes.md`; `content/agents/review/hive-review-code.md`, `hive-research.md`, `hive-review-plan.md`; `site/src/content/docs/flows/*.md`; `_support/docs/architecture/instruction-resources.md`; `tests/pilot/flash_class.go:25`, `tests/content/reading_footprint_test.go` (el escenario de plan/build no incluye referencias condicionales, así que no cambia).
