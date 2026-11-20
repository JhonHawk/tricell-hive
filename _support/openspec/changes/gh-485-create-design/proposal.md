# #485 — create-design: una sola skill de diseño que dibuja, revisa y guarda los criterios de UI

| Campo | Valor |
| --- | --- |
| Estado | En build · desde el 2026-11-20 |
| Tracker | [JhonHawk/tricell-hive#485](https://github.com/JhonHawk/tricell-hive/issues/485) · abierta · enhancement · relacionada: #475 (resuelta en la práctica, sin cerrar) |
| Git | `JhonHawk/tricell-hive` · base `development` en `34e58d40` · rama propuesta `feat/create-design` |
| Playbook | ninguno (plan pedido con `/flow-plan` sobre la issue) |

## Decisiones

2026-11-16 (sesión de investigación, registradas en `engram:tricell-hive/decision/create-design-skill`)
- **D5 A** — ¿Los criterios de UI siguen siendo una skill aparte? No: pasan a ser referencias de la skill de diseño.
- **D6 A** — ¿Cómo se llama la skill de diseño? `create-design`; el comando `/flow-figma-design` desaparece.
- **D7 A** — ¿Hay una entrada para revisar un diseño que ya existe? Sí: recibe un link de Figma, frames o un mock, devuelve hallazgos, estados faltantes, reglas sin decidir como preguntas y un veredicto, sin dibujar nada; la ejecuta `hive-review-ux` en un hilo aparte.

2026-11-20 (esta sesión)
- **D1 A** — ¿A qué ancho se dibuja una app solo móvil? Una sola vez, a 440 px, aunque los frames existentes del archivo usen otro ancho; una app web sigue con su ancho de escritorio más el frame de 440 px.
- **D2 B** — ¿Dónde se hace el ajuste del ancho móvil? Dentro de este cambio, no en uno aparte.
- **D3 A** — ¿Dónde viven los criterios de UI que leen la revisión, el build, el plan y verify? El `SKILL.md` de `hive-ui-criteria` se mueve entero a `create-design/references/ui-criteria.md` y sus cuatro referencias se mueven tal cual; los consumidores solo cambian la ruta del enlace. La única excepción al mover sin cambios es el texto que D4 cambia en `design-record.md`. El usuario preguntó antes si esos usos seguían funcionando: sí, porque todos leen por enlace.
- **D4 A** — ¿La etapa de diseño sigue entregando la edición de `DESIGN.md` (archivo o biblioteca de Figma) a mitad del diseño? No: la anota como pendiente en `examples.md` y el plan agrega la tarea que la escribe; un pedido que es solo diseño la nombra pendiente.
- **D8 A** — Con D4, ¿un sí al archivo o la biblioteca de Figma cambia la ruta al terminar el diseño? Envía la obra al plan, igual que hoy una convención; `approval.md` «Close the stage» no cambia.
- **Supuesto** — La entrada «corregir» de D7 es el caso actual «un pedido cuyo entregable es el diseño» con frames de destino; descartado: una entrada nueva. Se revierte separándolas.
- **Supuesto** — La entrada de revisión acepta un link de Figma (basta leer) o una imagen; sin conexión, solo imágenes; los hallazgos quedan en la conversación. Descartado: guardarlos en `_support/`.
- **Supuesto** — La página `flows/figma-design/` del sitio pasa a `flows/create-design/` con una redirección desde la ruta anterior, porque `README.md` y enlaces externos apuntan a ella. Descartado: conservar la URL vieja.
- **Supuesto** — El detector de rutas de `tests/pilot/session_audit_routes.go` reconoce los dos nombres, para que las sesiones grabadas antes del cambio sigan contando. Descartado: solo el nombre nuevo.
- **Supuesto** — Las entradas históricas de `CHANGELOG.md` no se tocan; el cambio agrega su línea en `[Unreleased]`. El sitio, que publica el changelog, seguirá nombrando las skills viejas solo ahí.
- **Supuesto** — El brief del revisor para un diseño que dibujó otra persona vive en `hive-rule-reviewer-brief`, su único hogar: archivo y nodos o la imagen, el pedido si lo hay, las pantallas existentes y las decisiones registradas, sin lista de hermanos del paso 3, inventario de biblioteca ni ancho. Descartado: escribirlo en `review.md`, que lo duplicaría.
- **Supuesto** — Señal de plataforma para D1: una app que corre solo en móvil es una app nativa, o una cuyo shell en el código y en los frames hermanos es de teléfono sin layout de escritorio; una app web con layout adaptable (un breakpoint o una etiqueta meta viewport, la señal de `hive-rule-ui-verification`) se sirve en móvil. Sin código ni frames, decide la plataforma que nombran el pedido o el ticket; sin ninguna, se dibuja como web con su frame de 440 px y se dice en el mensaje de aprobación. Descartado: preguntarlo siempre.
- **Supuesto** — Revisar una imagen no necesita conexión a Figma: `hive-mode` y el playbook `deliverable` no la frenan con el chequeo de conexión, que sigue aplicando a dibujar y a revisar un link. Descartado: exigir la conexión siempre.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> teniamos pendeinte un refactor de la skill de figma no? … hagamos el plan para la skill

(2026-11-20, Claude Code.) Ajustes pedidos en la misma sesión, observados en kin (Codex, `gpt-6.2-sol`, sesiones `01a1e3e1` del 2026-11-16 y `01a1f7d9` del 2026-11-20):

> 1) Se diseñaron versiones de 390px y 440px, es innecesario, bastaba con la de 440px
> 2) La agrupación se hizo siguiente ese par de vistas, osea, agrupaba los dos frames del login y luego agrupaba las dos vistas de permisos, etc

Criterios de la issue #485, leídos el 2026-11-20:

> 1. `flow-figma-design` pasa a llamarse `create-design`, y el comando anterior deja de existir en la misma versión.
> 2. Los criterios de UI, el registro de diseño y el refinamiento quedan como referencias de `create-design`, y `hive-ui-criteria` deja de ser una skill.
> 3. Plan, build, verify y los roles de diseño, frontend y revisión de UX enlazan esas referencias directamente.
> 4. Revisar un diseño que ya existe devuelve hallazgos por prioridad, los estados que faltan, las reglas sin decidir como preguntas con respuesta recomendada y un veredicto, sin dibujar ni mover nada.
> 5. Esa revisión la hace el revisor de UX en un hilo aparte.
> 6. Al dibujar, la skill termina en el diseño aprobado con su handoff y solo ofrece el siguiente paso.
> 7. El sitio, `llms.txt` y la lista de hogares canónicos nombran `create-design` y ya no nombran las dos skills anteriores.

Fuera de alcance según la issue: «Cambiar cómo se dibuja una pantalla y cambiar los criterios de UI mismos». D2 B mete en este cambio el ancho móvil y D4 A la entrega de `DESIGN.md`, que sí cambian cómo se dibuja; el resto de esa exclusión se mantiene.

## Objetivo

Hoy hay dos skills de UI: `flow-figma-design` dibuja y `hive-ui-criteria` guarda los criterios. La segunda no se carga nunca por su descripción (0 de 25 sesiones medidas el 2026-11-16); los roles la leen por enlace. No hay entrada para revisar un diseño que dibujó otra persona: en cinco pedidos se corrió el flujo completo o revisó el mismo hilo que dibujó. Una app solo móvil se dibuja dos veces (al ancho de los frames existentes y a 440 px), y las secciones apilan cada par. La etapa hace commit y PR de `DESIGN.md` a mitad del diseño.

El agrupamiento por pares de kin vino de la sesión del 2026-11-16, antes de que existiera la regla de secciones: a pedido del usuario agrupó cada vista con sus dos anchos en un frame blanco. No rompe ninguna regla vigente; con un solo ancho, ese agrupamiento ya no forma pares.

Al terminar: una skill `create-design` con cuatro entradas (dibujar, corregir, revisar un diseño existente, crear o actualizar `DESIGN.md`), los criterios como sus referencias, una app móvil dibujada una vez a 440 px, y la edición de `DESIGN.md` entregada por el plan.

## Alcance y aceptación

Dentro: la carpeta de la skill y sus referencias, los roles y hojas de regla que la enlazan, `hive-mode` y sus playbooks, las pruebas que nombran las skills, el sitio, `README.md`, `llms.txt`, `_support/docs/architecture/instruction-resources.md`, las specs `workflow-lifecycle`, `ui-criteria` y `documentation-site`, y `CHANGELOG.md`.

Fuera: el contenido de los criterios de UI (precedencia, pisos, modos), el resto del procedimiento de dibujo, el gestor (`tooling/`), y cerrar la #475.

Criterio 6 de la issue: ya se cumple en `development` (`content/skills/flow-figma-design/references/approval.md` «Close the stage» termina en el bloque de siguiente paso), así que no gradúa nada y no tiene criterio aquí.

- AC1 — Tras el refresco local, la skill `create-design` está instalada con sus referencias `ui-criteria.md`, `criteria.md`, `design-record.md`, `refinement.md` y `notice.md`, y ya no hay `flow-figma-design` ni `hive-ui-criteria` instaladas · falso hoy porque `ls ~/.agents/skills` lista `flow-figma-design` y `hive-ui-criteria` y no `create-design` · Observed in: `~/.agents/skills/` tras `hive update` desde el commit integrado
- AC2 — Ningún archivo distribuido, de `README.md`, `llms.txt`, de `_support/docs/architecture/` ni del sitio salvo el changelog nombra `flow-figma-design` ni `hive-ui-criteria`, ni el sitio enlaza `/flows/figma-design/` o llama al flujo «Figma design», y cada enlace que apuntaba a ellas resuelve a una ruta de `create-design` que existe; en los archivos desplegados, los enlaces `skill:` de los roles `hive-review-ux`, `hive-build-frontend` y `hive-design-ui` apuntan a `create-design/references/` · falso hoy porque `rg -l "flow-figma-design|hive-ui-criteria"` devuelve 30 archivos fuera de las dos carpetas de las skills, `CHANGELOG.md` y `_support/openspec/`, entre ellos esos tres roles · Observed in: `rg` sobre el repositorio y `~/.claude/agents/hive-review-ux.md` tras el refresco
- AC3 — La descripción y la tabla de entradas de `create-design` nombran sus cuatro entradas (dibujar, corregir, revisar un diseño existente, crear o actualizar `DESIGN.md`), y `create-design` tiene una entrada para revisar un diseño existente, dado como link de Figma o imagen: un hijo `hive-review-ux` lo revisa en un hilo aparte y la entrada devuelve hallazgos por prioridad, estados faltantes, reglas sin decidir como preguntas con respuesta recomendada y un veredicto, sin crear, mover ni cambiar nodos, y termina con un bloque de respuestas sin comando; `hive-mode` envía a ella un pedido de revisar un diseño, sin exigir conexión a Figma cuando es una imagen · falso hoy porque ni `flow-figma-design` ni el playbook `deliverable` tienen esa entrada (el playbook solo envía a dibujar, tras el chequeo de conexión) · Observed in: `~/.agents/skills/create-design/` y `~/.agents/skills/hive-mode/playbooks/deliverable.md` tras el refresco
- AC4 — En una app que corre solo en móvil, `create-design` dibuja cada pantalla una vez, a 440 px, aunque los frames existentes del archivo usen otro ancho, y no coloca un segundo frame debajo de cada pantalla; una app web conserva su ancho de escritorio más el frame de 440 px cuando se sirve en móvil, con la señal de plataforma de las decisiones · falso hoy porque el paso 5 toma el ancho de los frames existentes (390 en kin), el paso 6 agrega 440 px a cada pantalla y `sections.md` coloca ese frame debajo de cada una · Observed in: `~/.agents/skills/create-design/SKILL.md`, `references/draw.md`, `references/approval.md`, `references/rule-cards.md` y `references/sections.md` tras el refresco
- AC5 — Un sí a registrar el archivo o la biblioteca de Figma en `DESIGN.md` no escribe ni entrega nada durante el diseño: queda `record: … · pending` en `examples.md`, la etapa no hace la pregunta de entrega por ello, el plan agrega la tarea que lo escribe, y ningún texto de la skill dice que la etapa escribe esos campos · falso hoy porque `design-library.md` «Record it» escribe el campo y lo entrega de inmediato con git-workflow, y `SKILL.md`, `approval.md` y `design-record.md` lo repiten · Observed in: `~/.agents/skills/create-design/` (`SKILL.md`, `references/design-library.md`, `approval.md`, `design-record.md`) y `~/.agents/skills/hive-rule-ui-example/SKILL.md` tras el refresco
- AC6 — La spec describe lo nuevo: los deltas de esta carpeta nombran `create-design`, el ancho móvil, la entrega por el plan, la entrada de revisión y la página `flows/create-design/` · falso hoy porque las specs nombran `flow-figma-design`, `hive-ui-criteria` y `flows/figma-design/` · Observed in: `_support/openspec/specs/` tras el cierre, que aplica los deltas

## Requisitos previos

None: the build needs nothing from outside the repository.

## Contexto verificado

- `content/skills/flow-figma-design/SKILL.md` mide 8.928 bytes (2026-11-20, `wc -c`); el límite es 16.000: no hace falta partirlo. `hive-ui-criteria`: `SKILL.md` 4.114 bytes y cuatro referencias (13.857 bytes).
- Consumidores leídos (`rg`, 2026-11-20): roles `hive-review-ux`, `hive-build-frontend`, `hive-design-ui`; hojas `hive-rule-design-stage`, `-reviewer-brief`, `-ui-example`, `-ui-plan`, `-ui-verification`, `-business-record`, `-business-rules`, `-business-ticket`, `-checkpoint`, `-delegation`; `hive-mode` y sus playbooks `deliverable` y `ui-change`; `create-prd`, `flow-verify`, `unattended-delegation`; pruebas `tests/content/leaf_links_test.go` y `tests/pilot/session_audit_routes.go`; sitio, `README.md`, `llms.txt`, `instruction-resources.md`. `content/guidance/global.md` no nombra ninguna de las dos skills.
- Retirar una skill es borrar su carpeta: `hive update` quita sus archivos, como pasó con `flow-research` en `8de4a4a5`.
- La sesión de kin del 2026-11-20 ya corrigió su diseño: retiró las 40 variantes de 390 px y dejó las 40 vistas de 440 px organizadas por flujo.
- Medición del 2026-11-16 (`engram:tricell-hive/decision/create-design-skill`): 62 hilos cargaron `hive-ui-criteria`, 50 de ellos roles por enlace, 0 por descripción.

## Effects

None: the tasks take no effect on an external system.

## Entrega

2026-11-20 — Entrega `automatic`: revisión, push, PR e integración en `development` sin detenerse; una aprobación humana pendiente en el PR sigue reteniendo el merge. Revisión: nativa, Claude Code `/code-review`. No autoriza desplegar ni publicar una versión; el refresco local de la instalación tras el merge lo autoriza `AGENTS.md` «Refreshing a local installation» con el marcador de mantenedor.

## Estado de revisión y avance

2026-11-20 — Revisión en verify (hallazgos de `/code-review`): el delta de «Screen proposals in Figma before the plan» precisa que un shell de teléfono sin escritorio cuenta como solo móvil aunque tenga meta viewport o breakpoints, y que la verificación de una app solo móvil captura y compara solo 440 px. No cambia ningún criterio; no invalida tareas cerradas. El ancho, lo que se puede cambiar en el archivo y el dibujo por script pasan a `create-design/references/draw.md` para que `SKILL.md` quede bajo los 8.000 bytes que Codex lee.

2026-11-20 — Plan escrito; revisión de `hive-review-plan`: 1 P1, 7 P2 y varios P3, aplicados. Cambiaron: D4 cubre también `design-record.md`, `SKILL.md` y `approval.md` (F1); T5 suma cinco páginas del sitio (F2); T4 suma `flowCommand` (F3); los chequeos de T1–T4 corren en su orden y T4 depende de T3 (F4); el brief de la revisión vive en `hive-rule-reviewer-brief` (F5); señal de plataforma para D1 (F6); AC4 reescrito, porque la mitad de agrupación ya era cierta en la base y los pares de kin vinieron de la sesión del 11-16, previa a las secciones (F7); T6 revisa el plan en vez de editar los deltas en silencio (F8); revisar una imagen no exige conexión; AC2 deja fuera el changelog del sitio; AC3 nombra la cuarta entrada; base `f60d1f54`.
