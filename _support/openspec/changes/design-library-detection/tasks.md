# Tasks

Base `development` en `aaa0458e`; rama `feat/design-library-detection`, en el worktree `.claude/worktrees/design-library-detection`. Todas las tareas corren en el hilo principal: son texto de guía que depende de las decisiones de esta sesión, y ninguna pasa de cinco archivos.

### T1 — La etapa de Figma fija, inventaría y registra la librería, y valida la primera pantalla

- [x] `references/design-library.md` existe con sus cuatro partes, `SKILL.md` apunta a ella desde el paso 3, dibuja primero una pantalla cuando hay varias y sigue en 16,000 bytes o menos, y `rule-cards.md` habla de la librería fijada

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/flow-figma-design/references/design-library.md` (nuevo), `content/skills/flow-figma-design/SKILL.md`, `content/skills/flow-figma-design/references/rule-cards.md`

**Execution:** main thread

**Test approach:** check

**Changes:** La referencia lleva completa la regla de la librería, en el orden de `design.md` «Proposed shape»: qué librería obliga y su detección, el inventario, dibujar con sus piezas, y la pregunta, la escritura de los dos campos y su entrega, con el caso sin `DESIGN.md` y la delegación desatendida. En `SKILL.md`, con presupuesto de bytes: se liberan unos 680 (la frase de la librería del paso 5, unos 420; la pregunta del archivo del paso 6, unos 240; y los 22 de margen) y se añaden como mucho esos mismos: un solo puntero en el paso 3 con su condición de lectura (antes del paso 4, y dice que la referencia fija la librería, el inventario, el dibujo con sus piezas y el registro en los pasos 6 y 7; unos 300), la regla de la primera pantalla en el paso 5 (unos 280) y la excepción de la línea de apertura (unos 60). Los pasos 4, 7 y 8 y «A request for the design itself» cambian palabras sin crecer («declared» por «bound», y lo pendiente pasa a ser solo lo que la etapa no pudo escribir); «Unattended delegation» no cambia, porque su regla para la librería vive en la referencia. Si el archivo aún excede el tope, la frase del paso 3 que busca los registros de las hermanas pasa a la parte del inventario de la referencia, que ya lee a las hermanas. La primera pantalla: con más de una, se dibuja a su ancho de escritorio la que usa más piezas del inventario, se revisa, se corrige y se da su ubicación sin preguntar ni terminar el turno; la revisión del conjunto la incluye. `examples.md` se crea con el inventario antes del primer dibujo. En `rule-cards.md:22`, «the declared design library» pasa a ser la librería fijada. No cambia la aprobación única del conjunto, el contenido de las tarjetas ni las reglas de negocio.

**Verification:** `go test ./tests/content/...` pasa, incluido el tope de tamaño y la validez de los enlaces; `wc -c` de `SKILL.md` da 16,000 o menos; `rg -n "declared" content/skills/flow-figma-design` no deja ninguna regla de la librería condicionada a que esté declarada; lectura de la referencia contra AC1, AC2 y AC4 y del paso 5 contra AC3.

### T2 — El registro de diseño y el plan reflejan quién escribe archivo y librería

- [x] `design-record.md` dice que la etapa de Figma pregunta y escribe los dos campos, y `flow-plan` solo añade la tarea para un sí pendiente

**Closes:** AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/hive-ui-criteria/references/design-record.md`, `content/skills/flow-plan/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En `design-record.md`, los campos `**Design file:**` y `**Design library:**` de «Format» y la sección «Writing» llevan solo un puntero de una línea que nombra esos dos campos y la referencia nueva, que es el hogar de lo que la etapa pregunta, escribe y entrega; no repiten sus pasos. En `flow-plan/SKILL.md` «UI decisions» paso 4, la tarea del plan cubre cada convención y, solo con un sí pendiente, los dos campos. Las convenciones no cambian.

**Verification:** `go test ./tests/content/...` pasa; `rg -n "Design file|Design library" content/` no muestra ningún lugar que siga diciendo que el plan escribe el archivo o que la etapa no edita `DESIGN.md`.

### T3 — El revisor de UX sube de perfil y compara contra hermanas y librería

- [x] `hive-review-ux` usa el perfil `reasoning`, y su texto y su brief piden la imagen de las hermanas, la librería fijada, el inventario y la búsqueda de cada candidata

**Closes:** AC6, AC7

**Depends on:** T1

**Locations:** `content/agents/review/hive-review-ux.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Test approach:** check

**Changes:** `model_profile` pasa de `focused` a `reasoning`. El párrafo de marcos en un archivo de diseño sustituye «when the app's `DESIGN.md` declares a design library» por la librería que el brief nombra, añade la lectura de la imagen de cada hermana, la búsqueda de cada pieza `candidate:` y reportar como no corrida la comparación que no se hizo. «Reviewer brief» da la librería fijada con el inventario y las `candidate:`. Los pasos 1 a 4 del rol, que revisan una interfaz construida, no cambian.

**Verification:** `go test ./...` pasa; después del commit de esta tarea, porque el gestor despliega solo contenido con commit, el plan de despliegue que genera el gestor compilado del candidato (`go build -o <temporal> ./tooling/cli`, luego `<temporal> update --source . --out FILE`, sin aplicar) muestra para `hive-review-ux` `opus`/`high` en Claude y `gpt-6.1-sol`/`high` en Codex, leído sin sobrescrituras locales de modelo para ese rol o nombrando las que haya.

### T4 — La documentación describe el cambio

- [x] La entrega de agentes, el mapa de recursos de instrucciones y las páginas del sitio de la etapa y del plan dicen lo nuevo

**Closes:** AC8

**Depends on:** T1, T3

**Locations:** `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/instruction-resources.md`, `site/src/content/docs/flows/figma-design.md`, `site/src/content/docs/flows/plan.md`

**Execution:** main thread

**Changes:** `agent-delivery.md` saca a `hive-review-ux` del perfil `focused`, lo nombra en `reasoning`, fecha el cambio con sus dos observaciones y dice que alcanza también a la revisión de UX de `flow-verify` y que en Pi sube el esfuerzo de medio a alto. `instruction-resources.md` añade la entrada de `design-library.md` y le pasa la propiedad de la escritura de los dos campos (línea 19). `flows/figma-design.md` describe la librería detectada, el inventario, la primera pantalla, la pregunta única por archivo y librería, su escritura y entrega (líneas 25 a 28, 34, 43 y 49). `flows/plan.md:26` deja de decir que fuera de un plan el registro solo se escribe a petición. `README.md` y `llms.txt` no afirman nada que cambie. La tabla del piso de modelos no cambia.

**Verification:** `go test ./...` pasa; `npm run build` en `site/` termina sin errores.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y `npm run build` en `site/` | Covered by the implementation |
| Whole-behavior review | Final candidate, locally | Lista de lugares revisados y tamaño de `content/guidance/global.md` (sin cambio), para el PR | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |
| Commit message | Each commit of T1 and T3 | Nombra las dos observaciones del contexto verificado (host, modelo, sesión, fecha) y el fallo observado | Covered by the delivery answer |

Sin interfaz que recorrer: los criterios se observan en los archivos y en el plan de despliegue que genera el gestor. El comportamiento en una sesión real de diseño no se prueba aquí, porque los pilotos de comportamiento están en pausa.

## Checkpoint

- Updated: 2026-11-13
- Stage: verify · 6
- Last commit: c59f7b88
- Open findings: none
- Next: commit the second review round (Unreviewed), push, open the pull request to development, merge, refresh the local installation, close

### Evidence

- go test ./... · exit 0, every package ok · working tree of b3d59c03 (content, roles); T4 changed only _support/docs and site
- go test ./tests/content/... ./integrations/... · exit 0 · 67da38bc
- npm run build in site/ · 18 pages built · 67da38bc
- candidate manager: update --source . --out FILE (not applied) · hive-review-ux after: opus/high (Claude), gpt-6.1-sol/high (Codex); before: sonnet/high, gpt-6.1-terra/medium; no model overrides in the plan's config · b3d59c03
- wc -c content/skills/flow-figma-design/SKILL.md · 15977 (limit 16000) · 67da38bc
- CI: the repository's only workflow is master-source-branch.yml, which runs on pull requests to master; no job applies to this pull request
- /code-review on 67da38bc · 10 findings, all confirmed and fixed in c59f7b88
- /code-review on 67da38bc..c59f7b88 and hive-review-plan on the revised folder · no P0 or P1; contradictions fixed in the commit after c59f7b88, which gets no further review
- go test ./... · exit 0 · working tree after c59f7b88 (final candidate)
- npm run build in site/ · exit 0, 18 pages · final candidate
- wc -c content/skills/flow-figma-design/SKILL.md · 15984 (limit 16000) · final candidate

## Review status and progress

2026-11-13 — Plan escrito.
2026-11-13 — T1 · bee308a0 · unverified: none
2026-11-13 — T2 · d1b3cfe5 · unverified: none
2026-11-13 — T3 · b3d59c03 · unverified: none
2026-11-13 — T4 · 67da38bc · unverified: none
