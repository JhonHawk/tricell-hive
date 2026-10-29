# Diseño

## Contexto verificado

Textos leídos en `development` @ `668363e2` (2026-10-28). La guía desplegada coincide con `content/` en las líneas citadas.

- `content/guidance/global.md` (17 773 B):
  - `:13`: "...never an invented one; a one-line progress note takes none."
  - `:63`: "A mechanical change is a rename, a move, user-facing text, configuration that changes no permission, security setting or dependency, formatting, documentation, a text or style change that alters no layout, state or flow, or wiring existing checks exercise".
  - `:70`: "Before adding, updating or installing a package or tool from a registry, a one-off `npx` included, read and run..."
- `content/skills/flow-plan/references/delivery-decisions.md:32` (13 973 B): "It does not authorize a deployment: when a push there triggers one, name it, and ask before that push when no question is otherwise asked."
- `content/skills/flow-walkthrough/SKILL.md` (11 808 B):
  - `:12` lista los disparadores, sin el push de un repositorio de soporte.
  - `:24` define la primera vista renderizada: con parada de validación pregunta, si no, da la ruta del artefacto y sigue.
- `content/skills/flow-build/SKILL.md:62` (14 962 B): "unavailable prerequisites leave the affected criterion unverified".
- `content/skills/flow-plan/references/plan-criteria-and-markers.md:25-26` (4 355 B): `blocked: <prerequisite>` sin definir qué es un prerrequisito.
- `content/skills/flow-build/references/plan-execution.md:22` (12 912 B): "because a prerequisite is missing".
- `content/skills/flow-plan/references/plan-format.md:17` (15 089 B): "accounts/data prerequisites".
- `content/skills/flow-verify/references/review-and-gates.md:46,50`: ya pide "Resolve authorized local setup, accounts, roles and seed data" y "Prepare authorized local accounts, roles and seed data before the walk". Es el hogar canónico; la re-revisión de código le agregó el caso de la configuración local que falla (ver H2).
- `content/skills/flow-close/references/reporting.md:13,15` (15 893 B, margen de 107 B): "except a one-line progress note" y "A message that is only a one-line progress note...".
- `content/skills/flow-build/references/security-boundaries.md:22` (10 042 B): "including a one-off run such as `npx`...". `:34` prohíbe `@latest` en una ejecución puntual.
- `site/src/content/docs/flows/build.md:90` describe la primera vista renderizada. `:42` describe la línea de contexto sin la excepción, así que no la contradice.
- `tests/` y `tests/pilot/` no contienen ninguna de estas frases.

## Enfoque

Cada falla se corrige en el hogar canónico de su regla. Donde `global.md` restate el invariante, se ajusta esa copia para que no contradiga al hogar. No se agregan preguntas: H1 reutiliza la pregunta que ya existe antes del push que despliega, y solo acota qué respuesta la cubre.

### H1: push que despliega desde un repositorio `direct-base`

Alcance (revisión 2 del plan, pedida por el usuario el 2026-10-28 y precisada en la re-revisión): H1 aplica solo cuando el trabajo es únicamente de documentación. Un paso de documentación dentro de un trabajo sobre el proyecto real no agrega fricción.

- **AC1** usa "a documentation update that is part of a change to code or configuration".
- **AC2 y AC3** usan un solo criterio: "work whose changes, other than its change record, are only in repositories that declare `Delivery: direct-base`". Así el registro de cambios que vive en un repositorio de código no saca el trabajo del alcance.
- Esta corrección no agrega fricción al trabajo mixto. El disparador general de la primera vista (`flow-walkthrough/SKILL.md:24`, "When a change adds or rearranges a layout") y el gate de UI siguen como hoy; cambiarlos está fuera de alcance.

- **`global.md:63`.** La condición "que no altere el layout renderizado, el estado ni el flujo" pasa a cubrir también el texto visible y la documentación, salvo que la documentación acompañe otro trabajo. Texto propuesto: "A mechanical change is a rename, a move, configuration that changes no permission, security setting or dependency, formatting, wiring existing checks exercise, or user-facing text, documentation or style that alters no rendered layout, state or flow; a documentation update that is part of a change to code or configuration stays mechanical even when it alters a layout;".
- **`delivery-decisions.md:32`, despliegue.** Texto propuesto: "It does not authorize a deployment: when a push there triggers one, name it, and ask before that push unless an answer given for this change named that deployment; an answer for another change, or a delivery mode reused from earlier work, does not cover it, an exception to the same-kind rule of [effect authorization](../../hive-core/references/effect-authorization.md#what-an-answer-covers). When the work's changes, other than its change record, are only in repositories that declare `Delivery: direct-base` and the change adds or rearranges a rendered layout, show its first rendered view and ask before that push even when an answer named the deployment, as [flow-walkthrough](../../flow-walkthrough/SKILL.md) sets." (D7-A, durante el build)
  - Esto cierra la lectura de la sesión observada, que reusó el "actualicemos" de un cambio anterior.
  - El despliegue se nombra siempre.
  - Un "súbelo" que no supo del despliegue no lo cubre.
  - La excepción se nombra en este punto porque `effect-authorization.md:29` permite reusar una respuesta para "a later effect of the same kind" (`AGENTS.md`: una excepción usada en un solo punto se queda ahí y nombra la regla que cambia).
- **`delivery-decisions.md:32`, trabajo mixto, sin fricción.** "Inside each option, name every support repository, its branch, and when that option publishes it" pasa a "…, and when that option publishes it and any deployment that push triggers, including the closing push of a change record". Así la respuesta de entrega de un trabajo con código nombra el despliegue y cubre ese push sin otra pregunta. Esto incluye el push del registro de cambios al cierre, que `flow-close/SKILL.md:85` hace "without waiting for the code's question".
- **`delivery-decisions.md:30`.** La lista de lo que dice cada opción ("when it publishes each support repository in the work") agrega "and the deployment that push triggers", para que la tarjeta se sostenga sola.
- **`delivery-decisions.md:32`, solo repositorios de soporte.** La misma línea dice "When the implementation touches only support repositories, publish without asking." Pasa a: "…, publish without asking, except a push that triggers a deployment, as above."
- **`flow-walkthrough/SKILL.md`.**
  - `:12` agrega el disparador "before a push that deploys a new or rearranged layout, in work whose changes, other than its change record, are only in repositories that declare `Delivery: direct-base`".
  - `:24` agrega "In work whose changes, other than its change record, are only in repositories that declare `Delivery: direct-base`, before a push that deploys the layout, give the artifact path and ask before that push whether to publish it; approving it authorizes that push." (D7-A y revisión del cambio D7)
  - La captura sigue el procedimiento existente: servidor de desarrollo y una captura por vista.

### H2: prerrequisitos que el agente puede crear

Texto entregado, tras la revisión de código y la re-revisión:

- **Hogar canónico.** `review-and-gates.md:46` › Runtime and UI gates dice: "Resolve authorized local setup, accounts, roles and seed data; a failure while setting one up is a defect that blocks the work, fixed within the stage's fix-round limits; access only the user can supply, or local setup still failing after those rounds, remains a reported prerequisite, named with its cause."
- **`flow-build/SKILL.md:62`.** Dice: "an unavailable prerequisite leaves the affected criterion unverified only when [runtime and UI gates](…#runtime-and-ui-gates) leaves it as a reported prerequisite; resolve first the local setup that rule lists."
- **`plan-criteria-and-markers.md:30`.** Dice: "The `<prerequisite>` of a `blocked:` line is one that [runtime and UI gates](…) leaves as a reported prerequisite, named with its cause; the local setup it lists is resolved first."
- **`plan-execution.md:22`, paso 4.** Dice: "resolve one that [runtime and UI gates](…) lists as local, such as seed data or a local service, as it sets, and re-verify those criteria; for one it leaves as a reported prerequisite, leave `[?]` with a line `blocked: <prerequisite>`…"
- **`plan-format.md:17`.** Dice: "accounts/data prerequisites and how the build creates them locally".

"Do not start a fix round" sigue en el paso 4, igual que en la base. Se refiere a la ronda de corrección de la tarea, no a reparar la configuración local que fija el hogar canónico.

### H4: excepción de la línea de contexto

- **`global.md:13`.** "a one-line progress note takes none" pasa a "only a progress note of a single sentence takes none".
- **`reporting.md:13`.** "except a one-line progress note" pasa a "except a single-sentence progress note".
- **`reporting.md:15`.** "A message that is only a one-line progress note, such as" pasa a "A message that is only a progress note of one sentence, such as", y la oración termina con "a message of two or more sentences is not one".
- **`reporting.md:61` y `:110`.** "reply in one progress line" pasa a "reply in one progress sentence", y "report progress in a line" a "report progress in one sentence". Así no queda viva la lectura por línea que motivó H4. Es un ajuste del revisor de coherencia, no bloqueante, y entra si los bytes alcanzan.
- **Bytes.** Si el total supera el margen de 107 B, se acortan frases del mismo archivo sin cambiar su significado, como el ejemplo del final del párrafo. Se lista en el PR qué se acortó.

### H8: excepción para la CLI de Context7 (D2-B)

- **`security-boundaries.md`.** Al final del primer párrafo de "Check a dependency before installing" agrega: "Running the npm package `ctx7` with its `library` or `docs` subcommand to look up documentation, as the installed documentation skill instructs, such as `npx ctx7@latest docs <id> \"<query>\"`, is exempt from this section; any other package name or subcommand, `login` included, and installing it, such as `npm install -g ctx7`, are not. Its queries still follow the global rule on third-party requests."
  - Nombra el paquete exacto y los subcomandos, para que la exención no se extienda a nombres parecidos ni a otras skills.
  - La skill instalada ofrece `npm install -g ctx7@latest` y `npx ctx7@latest login` (`~/.agents/skills/find-docs/SKILL.md:31-34,149`). Ninguno de los dos queda exento.
- **`global.md:70`.** "a one-off `npx` included, read and run…" pasa a "a one-off `npx` included, read and run … except running `ctx7 library` or `ctx7 docs` through `npx` as the documentation skill instructs; installing it is not excepted." El invariante global conserva así el límite de instalación.
- `evidence-tools.md` no cambia.

### Sitio y registro de cambios

- `site/src/content/docs/flows/build.md:90` pasa de "Under `automatic`, `direct-base` and unattended delegation there is no such stop" a "Under `automatic`, `direct-base` chosen for a code repository, and unattended delegation there is no such stop". Además agrega una oración: en un trabajo que solo toca repositorios de soporte (`Delivery: direct-base`), antes de un push que despliega un layout nuevo o reacomodado, el agente muestra la primera vista y pregunta.
- `CHANGELOG.md`, sección Unreleased, recibe una entrada que resume H1, H2, H4 y H8 con la referencia a #262.

## Alternativas descartadas

- H8 con versión fija por sesión o con instalación global: el usuario eligió la excepción (D2-B).
- H1 como pregunta nueva para todo cambio de documentación: agregaría ceremonia. Basta acotar la pregunta que ya existe y la clasificación de mecánico.
