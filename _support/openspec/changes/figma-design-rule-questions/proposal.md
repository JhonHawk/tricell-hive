# figma-design-rule-questions — Figma: las reglas de negocio que surgen al dibujar se preguntan aparte

| Campo | Valor |
| --- | --- |
| Estado | Draft |
| Tracker | ninguno |
| Git | tricell-hive · base `development` @ `98266174` · rama propuesta `feat/figma-design-rule-questions` |

## Decisiones

2026-11-11
- **Origen** — ¿Hay un hueco? Sí. Dibujar una pantalla obliga a decidir reglas de negocio: qué exige una acción, quién la toma, qué estado sigue, qué bloquea, dónde vive un ajuste que afecta a más pantallas. Ninguna regla de la etapa de diseño dice qué hacer con ellas. El usuario eligió planear la regla (respuesta A a «¿Cómo seguimos?»). Como es una situación que ninguna regla cubre, la adopta la decisión del usuario sin esperar recurrencia; agrega preguntas, y el usuario la aprobó.
- **D6 A** — ¿Cómo se resuelve una regla que aparece al dibujar? Dentro de la etapa:
  - Se dibuja la recomendada, se anota como propuesta en el registro de ejemplos y se pregunta como decisión aparte en la ronda de aprobación. Aprobar la pantalla no aprueba la regla.
  - Una regla que añade una capacidad no pedida, cambia datos o ajustes que usan otras pantallas o módulos, o cambia un permiso, detiene las pantallas que moldea y lleva a investigación.
  - El plan toma las reglas aprobadas como decisiones y pregunta las que sigan propuestas.
- **D7 A** — ¿El revisor de UX también busca reglas sin fuente? Sí. Su brief lleva las fuentes que fijan reglas (petición, criterios del ticket, decisiones de investigación y decisiones registradas de la app) y reporta, aparte de sus hallazgos, cada comportamiento de los frames que ninguna fija.
- **Supuesto: decisiones menores** — Una regla cuyas opciones no cambian alcance, contrato, datos guardados, costo, reversibilidad ni quién puede actuar se declara como supuesto en el mensaje y queda registrada; no se pregunta. Así lo pide ya la regla global de preguntas, y evita llenar la ronda de detalles.
- **Supuesto: lo que la app ya hace** — El comportamiento que ya implementan las pantallas hermanas o su código cuenta como fuente. Una regla que la app ya resuelve no se pregunta.

## Solicitud

> En el diseño de los flujos pueden ir surgiendo dudas a resolverse no? como reglas de negocio, qué pasa en estas situaciones?

**Evidencia:** sesión Codex `01a1c93d-7172-74f0-a2cc-581c66a4aeb8` (fleetsystem FSR-858, gpt-6.1-sol, 2026-11-11).
- D20 y D21 llevaban reglas dentro de la aprobación visual: «permite registrar una nueva aprobación de una versión rechazada» y «el comentario es opcional en esta propuesta».
- La política de aprobación de toda la empresa quedó dentro del asistente de un solo workflow, y ni quien dibujó ni el revisor la señalaron.

## Objetivo

Hoy una regla que la pantalla decide queda implícita en el dibujo, y aprobar la pantalla la aprueba sin que el usuario la vea como regla. Con el cambio:
- cada regla sin fuente se registra y se pregunta como decisión propia, o se declara como supuesto cuando es menor;
- el revisor la busca por su cuenta;
- la que cambia el alcance va a investigación;
- la aprobada llega al plan como decisión.

## Alcance y aceptación

**Dentro:**
- `flow-figma-design`;
- el modo de frames de `hive-review-ux` y el brief de revisor en `delegation.md`;
- el paso de UI de `flow-plan`;
- la lista de transiciones de `content/guidance/global.md`;
- la especificación `workflow-lifecycle`;
- las páginas del sitio de la etapa y del plan;
- `CHANGELOG.md` y `instruction-resources.md`.

**Fuera:**
- `flow-research`, que ya resuelve reglas antes de dibujar y vuelve a recomendar la etapa al terminar;
- build y verify;
- las sesiones en curso, que cargaron los skills al arrancar.

- AC1 — `flow-figma-design` define una regla, sus fuentes (petición, criterios del ticket, decisiones de investigación, decisiones registradas de la app, y lo que ya implementan las pantallas hermanas o su código) y la línea de registro (`rule:` con regla, pantalla, estado, respuesta y quién la encontró). Además manda declarar como supuesto la regla menor · false today because `rg -n "rule:" content/skills/flow-figma-design/SKILL.md` no encuentra nada · Observed in: `content/skills/flow-figma-design/SKILL.md` en la rama integrada
- AC2 — En la ronda de aprobación, cada regla propuesta es su propia decisión, y la opción de aprobar la pantalla nombra esa decisión y vale con cualquier respuesta. Una respuesta distinta de la dibujada redibuja esa parte y la pantalla queda pendiente hasta que el usuario la apruebe redibujada · false today because el paso 6 solo pregunta aprobar o corregir cada pantalla · Observed in: `content/skills/flow-figma-design/SKILL.md` en la rama integrada
- AC3 — Una regla de alcance tiene que cumplir alguna de estas condiciones: añade una capacidad no pedida, cambia datos o ajustes que usan otras pantallas o módulos, o cambia un permiso. Cuando la encuentra quien dibuja o el revisor:
  - detiene el dibujo de las pantallas que moldea;
  - se pide la aprobación de las demás;
  - la etapa termina con el bloque de siguiente paso hacia `flow-research` que nombra esas reglas;
  - al volver, la etapa retoma el registro existente y dibuja solo las pantallas detenidas;
  - la salida a build exige que ninguna regla siga propuesta.

  · false today because el paso 8 solo lleva a plan o build y no menciona reglas · Observed in: `content/skills/flow-figma-design/SKILL.md` en la rama integrada
- AC4 — En el modo de frames, `hive-review-ux` reporta, aparte de sus hallazgos, cada comportamiento que no fijan sus requisitos ni las pantallas hermanas, sin reprobar el veredicto. El brief de frames pone las decisiones de investigación entre los requisitos y no incluye la lista de reglas propuestas de quien dibujó · false today because `rg -n -i "rule" content/agents/review/hive-review-ux.md` no encuentra nada · Observed in: `content/agents/review/hive-review-ux.md` y `content/skills/flow-build/references/delegation.md` en la rama integrada
- AC5 — `flow-plan` escribe como decisiones del plan las reglas aprobadas del registro y como supuestos las `assumption`, junto al registro del ejemplo; pregunta las propuestas antes de aprobar el plan y no decide una regla de alcance: la nombra y lleva a investigación, o bajo delegación desatendida la reporta sin resolver · false today because «UI decisions» paso 4 solo lee el estado de cada pantalla · Observed in: `content/skills/flow-plan/SKILL.md` en la rama integrada
- AC6 — La delta de `workflow-lifecycle` agrega el requisito «Business rules that appear while drawing» y ajusta el final de «Screen proposals in Figma before the plan». Las páginas `flows/figma-design` y `flows/plan` describen las preguntas de regla y la vuelta a Research · false today because ninguna lo menciona (`rg -n -i "rule" site/src/content/docs/flows/figma-design.md` sin resultados de reglas de negocio) · Observed in: la rama integrada y el sitio construido con `pnpm build`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- `content/skills/flow-figma-design/SKILL.md` @ `98266174`: 11 485 bytes; pasos 2–8 sin mención de reglas de negocio.
- `content/agents/review/hive-review-ux.md` línea 10: modo de frames.
- `content/skills/flow-build/references/delegation.md`: lista de requisitos en la línea 38 y caso de frames del brief en la línea 41. Las líneas 29 y 41 prohíben dar al revisor listas de lo «ya resuelto».
- `content/skills/flow-plan/SKILL.md` «UI decisions» paso 4: lee `examples.md` (12 585 bytes).
- `content/guidance/global.md` línea 90: la lista de transiciones dice «Figma design to plan or build» y no incluye research (47 400 bytes antes del cambio).
- `content/skills/unattended-delegation/SKILL.md` paso 5: elige la opción recomendada en las demás preguntas.
- `site/src/content/docs/flows/plan.md` línea 26: describe lo que el plan toma del ejemplo aprobado.
- Únicas menciones de `examples.md` en `content/`, `site/`, `tests/` y `_support/docs/`: `flow-plan` y `flow-figma-design`. Ningún detector de `tests/pilot/` lee estos textos.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Modo `automatic` y revisión con `/code-review` de Claude Code, respuestas D4 B y D5 A de esta sesión para este repositorio: revisar, push, PR y merge sin parar. No autoriza desplegar fuera del refresco local que `AGENTS.md` asigna al checkout del mantenedor.

## Estado de revisión y avance

- 2026-11-11 — Plan escrito.
- 2026-11-11 — Revisión del plan (`hive-review-plan`): 9 brechas a corregir y 6 menores, ninguna bloqueante. Se aplicaron todas:
  - **Reglas menores y lo que la app ya hace:** supuestos nuevos en Decisiones.
  - **Ronda:** la opción de aprobar vale con cualquier respuesta de la regla.
  - **Registro:** formato de la línea de regla.
  - **Unión de listas:** quien dibuja une la lista del revisor con la suya en el paso 5.
  - **Reglas del revisor:** también pasan la prueba de alcance.
  - **Vuelta a investigación:** se piden primero las demás pantallas y al volver se retoma el registro.
  - **Salida a build:** exige que no quede regla propuesta.
  - **Desatendido:** no decide reglas de alcance.
  - **Sitio:** se suma la página del plan.
  - **Guía global:** se suma la transición de diseño a investigación.
  - **Brief del revisor:** las decisiones de investigación van entre sus requisitos.
  - **Prueba de alcance:** queda como prueba propia de la etapa, no como clase de «Scope changes».
  - **Delta:** ahora es MODIFIED + ADDED.
  - **Verificaciones:** observables (`pnpm`, ruta real del script).
  - **Notas menores:** el paso 2 junta las fuentes, y se suman `instruction-resources.md` y la sección `[Unreleased]` del changelog.
- 2026-11-11 — Revisión de código (`/code-review`, 10 hallazgos) aplicada en una ronda:
  - **Estados nuevos:** la regla de alcance tiene su propio estado `scope` y no se pregunta en la ronda; las pantallas que moldea quedan `stopped by <regla>`.
  - **Antes de ir a investigación:** se termina el registro y la revisión de las demás pantallas.
  - **Reglas del revisor:** se contrastan con las fuentes antes de registrarlas.
  - **Aprobar con redibujo:** la pantalla queda pendiente aunque se haya elegido aprobarla.
  - **Plan:** lleva los supuestos y manda las reglas `scope` a investigación.
  - **Desatendido:** no termina hacia investigación.
  - **Revisor:** remite a la definición del skill en vez de copiarla.
  - **Brief:** las decisiones de investigación quedan solo en el caso de frames.
  - **Sitio:** el criterio de regla menor queda completo.
  - **AC5:** se ajustó a esos cambios.
