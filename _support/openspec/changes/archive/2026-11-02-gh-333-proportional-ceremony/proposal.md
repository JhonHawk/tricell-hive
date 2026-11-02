# #333 — Ceremonia proporcional: criterio de entrega, tarea operativa, disparadores del cierre y tickets

| Campo | Valor |
| --- | --- |
| Estado | Completed · integrado por PR #334 (`7194b2e1`) el 2026-11-02 |
| Tracker | [JhonHawk/tricell-hive#333](https://github.com/JhonHawk/tricell-hive/issues/333) · open · enhancement · related #328 |
| Git | tricell-hive · `development` @ `452b4e3e` · rama propuesta `fix/gh-333-proportional-ceremony` |

## Decisiones

2026-11-02
- ¿Se restituye el criterio de 0.6.0 para recomendar el modo de entrega? Sí: `automatic` cuando los checks, la revisión y la verificación establecen el resultado (infra, CI/CD, migraciones, refactors, tooling) e `interactive` cuando hay algo que solo el usuario puede juzgar, con una línea de razón junto a la recomendación. (D1 A)
- ¿Dónde vive el plan? Se agrega la tarea operativa (no deja cambios en el repo): sin carpeta de cambio ni pregunta de entrega, con los efectos críticos aún preguntados, y crear tooling nuevo en el repo para ella es decisión del usuario. El plan acotado vive en el chat, con una instantánea en el visor si está arriba; la carpeta de cambio se escribe solo cuando hay que retomarlo, coordinar entregas, registrar decisiones de peso o el usuario la pide. (D3 A)
- ¿Cuándo corre el cierre? También al terminar una tarea operativa y al abandonar el trabajo; `hold` se cierra cuando se entregue más adelante. La limpieza nombra contenedores, imágenes, servicios y dumps locales que creó la tarea, lista y pregunta los recursos remotos, y un visor caído no frena el resto del cierre. (D4, respuesta «hold es el único que no tiene cierre, hasta después»)
- ¿Cuándo se crea un ticket y cómo se relaciona? Solo para algo que hay que atender pero no se resolverá en esta sesión ni con el contexto actual; un hallazgo pequeño dentro de la zona se recomienda corregirlo ahora; el trabajo actual no necesita ticket (la carpeta usa un slug); todo ticket nuevo se busca antes como duplicado y se relaciona con su origen; una tarea operativa no mueve ni comenta su ticket salvo pedido. (D5 A)
- ¿Se crea un issue para planearlo? Sí, #333. (D6 A)
- ¿Cómo se entrega? `interactive`: parada antes del push para validar el diff, luego push, PR y merge. (D7 A)
- ¿Qué revisión corre? Claude Code `/code-review` sobre el candidato. (D8 A)

## Solicitud

Issue #333 (leído 2026-11-02), cuatro cambios: criterio de recomendación del modo de entrega; tarea operativa y dónde vive el plan; disparadores y alcance de la higiene del cierre; criterio de creación de tickets y su relación con el origen. Su aceptación provisional: cada regla con un solo hogar canónico y punteros en los demás lugares; páginas de `site/` actualizadas para el comportamiento de flujo que cambia; el cambio de tamaño de `global.md` declarado en el pull request.

## Objetivo

Hoy la guía no da criterio para la recomendación del modo de entrega, escribe una carpeta de cambio para todo plan, exige la clave de ticket en el nombre de la carpeta, no define cuándo crear ni cómo relacionar un ticket, y solo llega al cierre después de un merge o un push. Al terminar, la ceremonia sigue al tamaño y tipo real de la tarea: la entrega se recomienda por clase de trabajo, una tarea operativa no genera carpeta ni preguntas de entrega ni movimientos de tracker, el plan acotado queda en el chat, los tickets se crean solo para trabajo diferido y relacionados con su origen, y el cierre limpia también tras tareas operativas y abandonos.

## Alcance y aceptación

Dentro: `content/guidance/global.md`, `content/skills/flow-plan/` (SKILL, plan-handoff, plantilla de proposal), `content/skills/flow-build/SKILL.md`, `content/skills/flow-close/SKILL.md`, y las páginas `site/src/content/docs/flows/{plan,build,close}.md`. Fuera: el visor y su contrato (`work_key` ya es opcional sin ticket), la espera a los hijos antes del reporte que tenía 0.6.0, el aviso de fragmentos de secretos en tarjetas, y los restos del dump en el proyecto que lo originó.

- AC1 — «Delivery» de `global.md` dice qué modo de entrega recomendar: `automatic` cuando los checks, la revisión y la verificación establecen el resultado (infraestructura, CI/CD, migraciones, refactors, tooling o dependencias sin cambio visible), `interactive` cuando hay una superficie que solo el usuario puede juzgar, salvo que la guía del proyecto exija validación humana, y una línea de razón junto a la estrella · falso hoy porque `rg -n "Recommend .automatic|surface only the user" content/` no encuentra nada.
- AC2 — «Proportionality» define la tarea operativa (trabajo que no deja cambios para entregar en un repositorio, como un dump, una exportación o una operación de entorno): no lleva carpeta de cambio ni preguntas de entrega y revisión, sus efectos críticos se siguen preguntando, y agregar archivos de tooling al repositorio para ella es una decisión del usuario · falso hoy porque `rg -n "operational task" content/` no encuentra nada.
- AC3 — `flow-plan` escribe la carpeta de cambio solo cuando el usuario la pide, el trabajo debe retomarse entre sesiones o agentes (incluida la delegación de tareas a hijos), las entregas necesitan coordinación, o el plan cambia un contrato, datos almacenados, seguridad, dinero, concurrencia o una dependencia (dinero y concurrencia sumados en el build: `hive-verify-task` lee la carpeta); si no, el plan queda en el chat (con instantánea en el visor cuando corre) y `flow-build` lo ejecuta con su checkpoint en la conversación · falso hoy porque `flow-plan/SKILL.md` paso 3 siempre escribe la carpeta y `flow-build` exige `tasks.md` como checkpoint.
- AC4 — Una carpeta de cambio sin ticket usa un slug, y ninguna ruta pide crear un ticket para poder empezar un flujo · falso hoy porque `flow-plan/SKILL.md:35` exige la clave del ticket y `global.md:67` pide crear el ticket que falta antes del comando.
- AC5 — Un solo hogar en `global.md` define cuándo se crea un ticket (trabajo que hay que atender pero no se resolverá en esta sesión ni con su contexto), recomienda corregir ahora un hallazgo pequeño dentro de la zona, exige buscar duplicados antes y relacionar el ticket nuevo con su origen (blocks/blocked-by, sub-issue o related); `flow-close` apunta a esa regla en vez de repetir su propio criterio · falso hoy porque «Scope changes» dice «give it a destination, such as a ticket» sin criterio, y `rg -n "relate" content/guidance` no encuentra regla de relación.
- AC6 — Una tarea operativa no mueve ni comenta su ticket salvo que el usuario lo pida · falso hoy porque `global.md:22` y `flow-build` paso 1 mueven todo ticket que se implementa.
- AC7 — `flow-close` corre también al terminar una tarea operativa y al abandonar el trabajo, y declara que el trabajo en `hold` se cierra cuando se entrega; su limpieza cubre contenedores, imágenes, servicios y dumps locales que creó la tarea (autorizado en «Preservation»), lista los recursos remotos que creó con el comando que los quita y pregunta antes de quitarlos; con el visor caído, el cierre sigue y el reporte nombra lo que quedó abierto · falso hoy porque `flow-close` solo define el cierre en el punto de finalización, su paso 4 depende del cierre del visor y no nombra contenedores ni recursos remotos.
- AC8 — Las páginas `site/` de Plan, Build y Close describen ese comportamiento, y ninguna instrucción distribuida contradice las anteriores · falso hoy porque `site/src/content/docs/flows/plan.md` dice que la salida es siempre una carpeta de cambio.

## Prerrequisitos

Ninguno: el build no necesita nada fuera del repositorio.

## Contexto verificado

- 2026-11-02, `development` @ `452b4e3e`, árbol limpio salvo `_support/sessions/2026-11-02-subagent-effort-per-call/` (ajeno).
- Sesiones revisadas: Claude Code cam `32cffc1d`, `9a6a7fe6`; carvento `8d09a5eb`, `b950c7b3`; OpenCode fleetsystem `ses_e62ddddbbffeZnFbV22SUxvq9u` (todas 2026-11-02). Evidencia y decisiones: Engram `tricell-hive/decision/session-review-2026-11-02-ceremony`.
- 0.6.0: `_support/workspace/2026-10-30-skills-rebuild/baseline/content/skills/flow-plan/references/delivery-decisions.md:23-24`, `flow-plan/SKILL.md:69`, `flow-close/references/incidental-findings.md:15`, `flow-research/SKILL.md:42`. El criterio de entrega se quitó en `7a37381b`.
- Hoy el cierre se alcanza solo por `git-workflow/SKILL.md:31` (paso 10) y `flow-deploy`.
- Trackers: `gh issue create --parent/--blocked-by/--blocking` (gh 2.103.3) y `acli jira workitem link create` verificados localmente; Linear `save_issue` con `relatedTo` observado en las sesiones de cam.
- Sin delta de specs: `_support/openspec/specs/` cubre instalación, visor, sitio y entrega de contexto; ninguno declara estas reglas de flujo, como en #328.

## Entrega

`interactive` (D7 A): rama `fix/gh-333-proportional-ceremony`, commits locales, revisión con Claude Code `/code-review` (D8 A) y parada antes del push para que el usuario valide el diff; tras su aceptación, push, PR a `development` y merge. No autoriza el refresco de la instalación local fuera de lo que `AGENTS.md` «Refreshing a local installation» ya fija para el checkout del mantenedor, ni un release.

## Estado de revisión y progreso

2026-11-02 — Plan escrito. Sin `hive-review-plan`: no cambia contrato, datos almacenados, seguridad ni dependencias. Visor no disponible (`hive view status`: `running:false`); no se publicó nada.
2026-11-02 — Build T1–T6, dos rondas de `/code-review` (10 + 9 hallazgos corregidos en `ba782a3e` y `bbfda812`), validación del usuario (D9 A), PR #334 integrado en `7194b2e1`; #333 cerrado.
