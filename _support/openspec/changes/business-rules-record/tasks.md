# Tasks

Base: `tricell-hive` `development` en `b9113a3c`; rama `feat/business-rules-record`. T1 a T5 corren en el hilo principal: son textos que dependen unos de otros y de las decisiones de la sesión, y cada uno edita tres archivos o menos, salvo T3, que agrega una frase a cuatro. T6 va a un hijo `hive-write-docs`: edita más de tres archivos a partir de fuentes que T1 a T5 ya fijaron.

### T1 — La referencia común de reglas de negocio y el registro en el plan

- [x] Existe `business-rules.md` con la definición, la línea del registro y sus estados, dónde vive el registro y cuál gobierna, la escritura al ticket con sus casos (sin ticket, abandonado, pausado, tickets espejo, herramienta que no edita la descripción) y lo que hace cada flujo; la plantilla de `proposal.md` tiene la sección `## Business rules`, con encabezado en inglés; `flow-plan/SKILL.md` apunta a la referencia en «UI decisions» paso 4, «Change folder» y «Revise a plan in progress».

**Closes:** AC1, AC4

**Depends on:** none

**Locations:** `content/skills/flow-plan/references/business-rules.md` (nuevo), `content/skills/flow-plan/assets/change-folder/proposal.md`, `content/skills/flow-plan/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** lo que fija `design.md` «Proposed shape» (línea del registro, sección del ticket) y sus «Decisions». En `flow-plan/SKILL.md:22` la frase que escribe las reglas «beside the example's record in the `UI` section» pasa a mandarlas a `## Business rules` de `proposal.md`. No cambia `check_change.py` ni las plantillas de `tasks.md` y `design.md`.

**Verification:** `go test ./tests/content/...` pasa (tamaños y enlaces). `rg -n "business-rules.md" content/skills/flow-plan/SKILL.md` muestra tres punteros. `rg -n "beside the example's record" content/skills/flow-plan/SKILL.md` no devuelve nada. `rg -n "cancelled <date>|ticket write:|ticket section:|## Business rules" content/skills/flow-plan/references/business-rules.md` muestra las cuatro. `rg -n "^## Business rules" content/skills/flow-plan/assets/change-folder/proposal.md` muestra una línea.

### T2 — La guía global autoriza la escritura de reglas al ticket

- [x] «Scope and authorization» autoriza la sección de reglas y su comentario con sus tres momentos y el puntero, sin pregunta aunque cambien datos que la tarea no creó, y corregir las filas de tarjetas que el trabajo creó; la frase de la tarea operativa nombra la excepción del diseño que nombra su ticket; «Project settings» cubre esa escritura con `Other trackers`; `outside-the-repository.md` la nombra junto a los movimientos del ticket.

**Closes:** AC2

**Depends on:** T1

**Locations:** `content/guidance/global.md`, `content/skills/flow-plan/references/outside-the-repository.md`

**Execution:** main thread

**Test approach:** check

**Changes:** la regla del ticket (`global.md:21`), las dos frases de `Other trackers` (`:111`, `:113`) y las líneas 3 y 19 de `outside-the-repository.md`. No cambia la definición de efecto crítico (`:16`) ni la regla de delegación (`:23`): la excepción vive en la regla del ticket.

**Verification:** `go test ./tests/content/...` pasa. `rg -n "business-rules.md" content/guidance/global.md` muestra el puntero en «Scope and authorization». `rg -n "business rules" content/skills/flow-plan/references/outside-the-repository.md` muestra dos líneas. `rg -n "move and comment each ticket in" content/guidance/global.md` no devuelve una frase que excluya las reglas. `wc -c content/guidance/global.md` antes y después, anotado para el pull request.

### T3 — Los momentos y los punteros en research, build, verificación y cierre

- [x] `flow-research` paso 1 nombra los comentarios del ticket, también cuando un hijo lo lee; `flow-build` escribe las reglas al arrancar, junto a mover el ticket, y apunta a la referencia para una regla decidida o asumida a mitad del build; `flow-verify` apunta a ella para una corrección que decide una regla; `flow-close` paso 3 lleva su frase y el paso 5 el recorte que `design.md` fija.

**Closes:** AC3, AC4, AC7

**Depends on:** T1

**Locations:** `content/skills/flow-research/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-close/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** solo punteros que nombran la regla y cuándo leerla; ningún paso de la referencia se repite en ellos. En `flow-close`, exactamente los dos textos de la fila `flow-close/SKILL.md` de `design.md`. `flow-research` sigue sin escribir el tracker.

**Verification:** `go test ./tests/content/...` pasa. `rg -c "business-rules.md" content/skills/flow-build/SKILL.md content/skills/flow-verify/SKILL.md content/skills/flow-close/SKILL.md` da 2, 1 y 1. `rg -n "comments" content/skills/flow-research/SKILL.md` muestra el paso 1. `wc -c` de los cuatro `SKILL.md` queda en 16,000 o menos, con `flow-close` en 15,995.

### T4 — Las tarjetas de reglas en la etapa de Figma

- [x] Existe `rule-cards.md`; `flow-figma-design/SKILL.md` define la regla por puntero a la referencia común, dibuja las tarjetas tras la primera revisión, las incluye en la segunda pasada, actualiza la fila y escribe el ticket al aprobarse el ejemplo, manda las líneas `rule:` al registro de reglas del plan y dice qué escribe la etapa en su línea 8.

**Closes:** AC3, AC5, AC6

**Depends on:** T1

**Locations:** `content/skills/flow-figma-design/references/rule-cards.md` (nuevo), `content/skills/flow-figma-design/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** lo que fija `design.md` «Tarjeta de reglas» y la fila de `flow-figma-design` en «Dónde va cada puntero», incluidas «A request for the design itself» y «Unattended delegation». Lo que no quepa en los 16,000 bytes de `SKILL.md` va a `rule-cards.md` detrás del puntero. No cambia cómo se clasifican las reglas `proposed`, `assumption` y `scope`, ni la ronda de aprobación.

**Verification:** `go test ./tests/content/...` pasa. `rg -n "card stale|frame stale|stated:" content/skills/flow-figma-design/references/rule-cards.md` muestra las tres. `rg -n "rule-cards.md|business-rules.md" content/skills/flow-figma-design/SKILL.md` muestra los punteros de los pasos 5 y 7. `rg -n "A rule is a behavior in a situation" content/skills/flow-figma-design/SKILL.md` no devuelve nada. `wc -c content/skills/flow-figma-design/SKILL.md` queda en 16,000 o menos.

### T5 — El revisor de UX y su brief cubren las tarjetas

- [ ] `hive-review-ux.md` dice qué revisa en una tarjeta y toma la definición de regla de la referencia común; «Reviewer brief» da las tarjetas solo en la segunda pasada.

**Closes:** AC8

**Depends on:** T4

**Locations:** `content/agents/review/hive-review-ux.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Test approach:** check

**Changes:** el párrafo de frames en un archivo de diseño de cada archivo. En `hive-review-ux.md:10`, «as step 5 of flow-figma-design defines one» pasa a un enlace `skill:flow-plan/references/business-rules.md`. La primera pasada sigue sin recibir las reglas del dibujante.

**Verification:** `go test ./tests/content/...` pasa, incluido `agent_boundary_test.go`. `rg -n "card" content/agents/review/hive-review-ux.md content/skills/flow-build/references/delegation.md` muestra las tres comprobaciones del revisor y la regla del brief. `rg -n "as step 5 of" content/agents/review/hive-review-ux.md` no devuelve nada.

### T6 — El sitio, README, llms.txt y la documentación de arquitectura

- [ ] Las páginas del sitio, `README.md` y `llms.txt` describen las tarjetas y la escritura de reglas al ticket; `instruction-resources.md` nombra al dueño de cada referencia nueva.

**Closes:** AC9

**Depends on:** T1, T2, T3, T4, T5

**Locations:** `site/src/content/docs/flows/figma-design.md`, `flows/build.md`, `flows/close.md`, `flows/plan.md`, `flows/verify.md`, `flows/research.md`, `flows/overview.md`, `getting-started.md`, `README.md`, `llms.txt`, `_support/docs/architecture/instruction-resources.md`

**Execution:** `hive-write-docs`

**Changes:** las páginas describen el comportamiento y enlazan a la skill; no copian sus reglas. No se agrega fila a la rúbrica de revisión de sesiones: esa fila la decide el usuario.

**Verification:** `go test ./...` pasa. `pnpm --dir site build` pasa. `rg -n -i "business rules|rule cards?" site/src/content/docs/flows/figma-design.md site/src/content/docs/flows/close.md site/src/content/docs/flows/build.md README.md llms.txt` muestra cada archivo. `rg -n -i "moves? and comments?" site README.md llms.txt` no deja una frase que excluya la escritura de reglas.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | Los comandos de la `Verification` de la tarea | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `pnpm --dir site build` | Covered by the implementation |
| Whole-behavior sweep | Final candidate, locally | `rg` de «business rule», `rule:`, «move and comment», «writes only in that Figma file», «the ticket's moves» y `examples.md` en `content/`, `site/`, `_support/docs/architecture/` y `tests/pilot/`; los lugares revisados se listan en el pull request | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` de Claude Code sobre el diff de la rama (D6 A) | The review answer |

El comportamiento de un agente con estas reglas no se comprueba en este trabajo: los pilotos están en pausa. Queda para las sesiones reales posteriores.

## Checkpoint

## Review status and progress

2026-11-13 — Plan escrito y revisado; ninguna tarea iniciada.
2026-11-13 — T1 · 447c5eea · unverified: none
2026-11-13 — T2 · b408e21b · unverified: none
2026-11-13 — T3 · 0838055c · unverified: none
2026-11-13 — T4 · 1dee6fa7 · unverified: none
