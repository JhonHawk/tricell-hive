# Tasks

Base: tricell-hive `development` en `a1c6c9ac`; rama `feat/flow-figma-design`. T1 a T3 van en el hilo principal porque son la redacción de reglas distribuidas, que el hilo principal debe revisar entera contra el inventario; T4 va a un hijo `hive-write-docs` porque edita más de tres archivos de documentación a partir de fuentes ya fijadas.

El mensaje del commit que integra T1 nombra las observaciones que motivan el cambio: Claude Code, `claude-fable-5-2`, sesión `c71dd314`, 2026-11-10 (la evaluación de un ticket seguida del plan no alcanzó la oferta) y Claude Code, `claude-fable-5-2`, sesión `01acdb18`, 2026-11-10 (sistema de diseño sin cargar, 390 px en móvil, registro sin actualizar, pasos manuales sin avisar). El pull request lista los lugares revisados y el cambio de tamaño de `content/guidance/global.md`.

### T1 — La skill `flow-figma-design`

- [x] Existe la skill con sus ocho pasos en orden y pasa las pruebas de contenido

**Closes:** AC1, AC5

**Depends on:** none

**Locations:** `content/skills/flow-figma-design/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Escribe la skill a partir de la sección actual de `flow-research`, sin perder ninguna de sus intenciones salvo las opciones «text sketches» y «no proposals» de la oferta, que D10 A retira. Agrega los pasos de D5 A, D6 A y D8 A, y dice que la respuesta que inicia la etapa es la instrucción para escribir en el archivo, sin sustituir la pregunta de un efecto crítico. La etapa no edita `DESIGN.md`: guarda el sí en el registro. No cambia `flow-build`.

**Verification:** `go test ./tests/content/...` pasa; `wc -c content/skills/flow-figma-design/SKILL.md` da 16.000 o menos.

### T2 — La investigación recomienda la etapa y la guía global nombra la transición

- [x] `flow-research` ya no tiene la sección ni la oferta y recomienda la etapa en su bloque

**Closes:** AC2, AC6

**Depends on:** T1

**Locations:** `content/skills/flow-research/SKILL.md`, `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Quita «Examples for a screen» y sus cuatro apuntadores; el paso 5 recomienda `flow-figma-design` cuando hay pantalla nueva o recompuesta sin composición fijada y el host tiene una conexión que escribe en Figma, también en la terminación de un ticket evaluado, y nunca bajo delegación desatendida. El paso 0 envía a la etapa, con el bloque de siguiente paso, la petición cuyo entregable es el diseño. En «Questions and routes», la lista de transiciones gana las dos nuevas y la cláusula «Offer a flow's command only when the next step needs that flow's procedure» gana un diseño en Figma. No cambia las tres continuaciones en el mismo turno.

**Verification:** `rg -n "Examples for a screen|examples-for-a-screen" content/skills/flow-research` no devuelve nada; `rg -n "flow-figma-design" content/skills/flow-research/SKILL.md content/guidance/global.md` devuelve la recomendación y la transición; `go test ./tests/content/...` pasa.

### T3 — Los consumidores apuntan a la skill nueva y el rol gana sus límites

- [x] Ningún archivo distribuido apunta a la sección retirada, y el rol y el registro de diseño dicen lo decidido

**Closes:** AC3, AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-plan/assets/change-folder/design.md`, `content/skills/flow-build/references/delegation.md`, `content/agents/design/hive-design-ui.md`, `content/skills/hive-ui-criteria/references/design-record.md`, `content/skills/unattended-delegation/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Cambia cada enlace a la sección por el de la skill nueva, reescribe las frases que dicen que la investigación dibuja o deja el ejemplo («When research left an approved example», el punto «in research» de `delegation.md`, que gana uno propio para la etapa) y quita la rama del boceto en texto donde el ejemplo aprobado ya solo puede ser un frame, incluida la de `flow-verify/SKILL.md:12`. `flow-plan` paso 4 nombra dónde está el registro y agrega la tarea que escribe «Design file» cuando el registro trae esa instrucción. `hive-design-ui` permite crear componentes y variables locales al archivo, con sus identificadores en el registro (D7 A). `design-record.md` dice que la etapa hace la pregunta del registro y que el plan escribe el campo (D6 A). No cambia el comportamiento de comparación de `flow-build`, `flow-verify` ni `hive-review-ux`.

**Verification:** `rg -n -i "examples-for-a-screen|research left an approved|in research, .*design file" content` no devuelve nada; `rg -n "Never touch variables" content/agents/design/hive-design-ui.md` no devuelve nada; `go test ./tests/content/...` pasa.

### T4 — Sitio, README, llms.txt, arquitectura y changelog

- [x] La documentación para personas describe la etapa y el sitio compila

**Closes:** AC3, AC6

**Depends on:** T1, T2, T3

**Locations:** `site/src/content/docs/flows/figma-design.md` (nuevo), `site/astro.config.mjs`, `site/src/content/docs/flows/overview.md`, `site/src/content/docs/flows/research.md`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/skills.md`, `site/src/content/docs/getting-started.md`, `README.md`, `llms.txt`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** hive-write-docs

**Changes:** Página nueva con las secciones que exige `documentation-site` y el enlace a su `SKILL.md`; entrada en la barra lateral y en las tablas y listas de flujos; el diagrama y la lista de `README.md`; la línea de `llms.txt`; el hogar canónico del ejemplo aprobado en `instruction-resources.md`; la entrada «Screen examples in research» de «Unreleased» en `CHANGELOG.md`, reescrita para la etapa nueva en lugar de agregar otra; las frases del sitio que dicen que la investigación ofrece o deja el ejemplo (`flows/overview.md:32`, `flows/plan.md:26`, `flows/research.md:41,44`). Las páginas describen la conducta y enlazan a la skill; no copian sus reglas.

**Verification:** `npm run build` en `site/` termina sin errores; `rg -n -i "examples-for-a-screen|research left an approved|in research, .*design file|offer of screen proposals|Screen examples in research" content site _support/docs/architecture README.md llms.txt CHANGELOG.md` no devuelve nada; `rg -n "figma-design" site/astro.config.mjs README.md llms.txt` devuelve las entradas.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` (unos 4 minutos, medido en la sesión `8bca76f5`) | Covered by the implementation |
| Site build | Final candidate, locally | `npm run build` en `site/` | Covered by the implementation |
| Stale wording | Final candidate, locally | El `rg` de AC3 sobre todos sus lugares no devuelve nada | Covered by the implementation |
| Plan coherence | Final candidate, locally | `python3 <flow-plan-dir>/scripts/check_change.py` sobre esta carpeta sale con 0 | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | El mecanismo que fije la ronda de entrega | The review answer |

Lo que este cambio no verifica: la conducta en una sesión real. Los pilotos están en pausa, así que los criterios se observan en los archivos del candidato; tras el despliegue, las sesiones que pasen por la etapa son los episodios que confirman o reabren el cambio.

## Checkpoint

## Review status and progress

2026-11-10 — Plan escrito.
2026-11-10 — Revisión del plan aplicada: T2, T3 y T4 ampliadas; ninguna tarea hecha invalidada.
2026-11-10 — T1 · c5fa4149 · unverified: conducta en una sesión real: los pilotos están en pausa
2026-11-10 — T2 · bc3455b4 · unverified: conducta en una sesión real: los pilotos están en pausa
2026-11-10 — T3 · ee689b2c · unverified: conducta en una sesión real: los pilotos están en pausa; plan-handoff.md y la plantilla design.md no necesitaron cambio
2026-11-10 — T4 · c45a6c0b · unverified: none
