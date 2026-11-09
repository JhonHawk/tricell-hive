# Tasks

Base: `development` en `997c5485`; rama `feat/design-examples-in-research`. T1 a T5 y T7 corren en el hilo principal: cada una edita tres archivos o menos y su texto depende del de las otras. T6 edita seis archivos de documentación y va a un hijo `hive-write-docs`.

### T1 — Propuestas de pantalla en research

- [x] `flow-research` tiene la sección «Examples for a screen», el rol `hive-design-ui` entrega ejemplos en el archivo de diseño que su encargo le asigna y la referencia de delegación nombra ese trabajo

**Closes:** AC1, AC2, AC3, AC4, AC7, AC11, AC12

**Depends on:** none

**Locations:** `content/skills/flow-research/SKILL.md`, `content/agents/design/hive-design-ui.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en `flow-research`, la sección nueva con la condición, la oferta y sus opciones, el número de ejemplos, el límite de escritura, la elección, el registro y el pedido de solo diseño, como fija `design.md` «Proposed shape»; la excepción en el primer párrafo; un puntero desde los pasos 0, 3 y 5. La sección dice además que la opción del archivo de diseño nombra la escritura, que esa respuesta es su instrucción explícita, que una escritura de pago pide su propia pregunta, que la oferta solo se hace cuando el usuario pidió un cambio, y que tras aprobar un pedido de solo diseño research termina con el bloque de siguiente paso hacia el cierre. En el rol, el paso 1 (línea 11, bocetos «for the `UI` section of `design.md`») admite la página del archivo de diseño que el encargo asigna y dice que cuenta entre las rutas que el encargo nombra; los pasos 3 y 4 dicen que la evidencia de un frame es su imagen y sus nodos. El párrafo final del rol (línea 19) no se toca: `tests/content/agent_boundary_test.go` lo fija palabra por palabra. En la referencia de delegación, un punto bajo el trabajo que un hijo hace por costo. No se renumeran los pasos de research. No cambia `global.md`.

**Verification:** leer la sección contra AC1, AC2, AC3 y AC11, cláusula por cláusula, y anotar la línea de cada una en el registro de la tarea, incluidas las de la respuesta que autoriza la escritura, la escritura de pago y el bloque hacia el cierre; `git diff -- content/agents/design/hive-design-ui.md` no toca el párrafo final del rol; además: `rg -n "Examples for a screen" content/skills/flow-research/SKILL.md` muestra la sección y los punteros de los pasos 0, 3 y 5 (cuatro líneas o más); `rg -n -i "design file" content/skills/flow-research/SKILL.md content/agents/design/hive-design-ui.md` muestra ambos; `rg -n "hive-design-ui" content/skills/flow-build/references/delegation.md` muestra el punto nuevo; `awk 'NR==8' content/skills/flow-research/SKILL.md` nombra los ejemplos; `wc -c` de `flow-research/SKILL.md` y de `delegation.md` da 16,000 o menos; `git diff --stat -- content/guidance/global.md` no muestra nada; `go test ./tests/content/...` pasa.

### T2 — Campo del archivo de diseño en el registro de diseño

- [x] El registro de diseño describe el campo `**Design file:**`

**Closes:** AC5

**Depends on:** none

**Locations:** `content/skills/hive-ui-criteria/references/design-record.md`

**Execution:** main thread

**Test approach:** check

**Changes:** el campo en la lista de la sección `## Hive` y en el ejemplo, con la herramienta y la ubicación; que se escribe solo por instrucción del usuario, que no lleva huella ni cuenta para la caducidad, y que no autoriza escrituras por sí mismo. No cambia la detección.

**Verification:** `rg -n "Design file" content/skills/hive-ui-criteria/references/design-record.md` muestra el campo en la lista y en el ejemplo (dos líneas o más), y la lectura de esas líneas confirma las cuatro cláusulas de AC5; `go test ./tests/content/...` pasa.

### T3 — El plan usa el ejemplo aprobado

- [x] «UI decisions», el resumen del plan y la plantilla `design.md` toman el ejemplo aprobado como referencia

**Closes:** AC6

**Depends on:** none

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-plan/assets/change-folder/design.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en «UI decisions», el ejemplo aprobado es el boceto «después» de las regiones que muestra y su referencia o su boceto va a la sección `UI`; el paso 5 no lanza el hijo para una composición ya aprobada. En el resumen del plan, la referencia del ejemplo en lugar del boceto de esas regiones. En la plantilla, el comentario de la sección `UI`. Los pasos 1 a 3 no cambian.

**Verification:** `rg -n "approved example" content/skills/flow-plan` muestra los tres archivos, y la lectura de «UI decisions» confirma que los pasos 4 y 5 lo nombran; `wc -c content/skills/flow-plan/SKILL.md` da 16,000 o menos; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.

### T4 — Build y verify comparan contra el ejemplo

- [x] La primera vista del build y el recorrido de verify comparan contra el ejemplo aprobado

**Closes:** AC8

**Depends on:** none

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en `flow-build`, la primera vista se compara contra el ejemplo aprobado cuando el plan tiene uno. En `flow-verify`, el recorrido compara la captura de cada región con ejemplo; el texto define qué es una diferencia como `design.md` «Diferencia» fija, la trata como defecto de UI que el cambio introdujo, y lista las que queden en el mensaje de validación cuando la entrega se detiene y en el reporte final cuando no. La validación del usuario y sus condiciones no cambian.

**Verification:** `rg -n "approved example" content/skills/flow-build/SKILL.md content/skills/flow-verify/SKILL.md` muestra ambos; en `flow-verify`, la lectura de las líneas encontradas confirma la comparación en el recorrido, la definición de diferencia y los dos destinos de la lista (mensaje de validación y reporte final), y `git diff -- content/skills/flow-verify/SKILL.md` no toca la frase «ask the user to approve or request corrections»; `wc -c` de los dos da 16,000 o menos; `go test ./tests/content/...` pasa.

### T5 — La revisión de UX compara contra el ejemplo

- [x] `hive-review-ux` y el encargo del revisor incluyen la imagen del ejemplo aprobado

**Closes:** AC9

**Depends on:** none

**Locations:** `content/agents/review/hive-review-ux.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en el rol se cambia la frase existente «judge the result against the approved sketches as well» para que incluya el ejemplo aprobado y su imagen, sin añadir una frase paralela, y reporte cada diferencia, como la define `flow-verify`, con las prioridades que ya usa; «Reviewer brief» nombra esa imagen entre los requisitos que se le dan. No cambia la severidad de los hallazgos ni lo demás del encargo.

**Verification:** `rg -n "approved example" content/agents/review/hive-review-ux.md content/skills/flow-build/references/delegation.md` muestra ambos; `rg -c "approved sketches" content/agents/review/hive-review-ux.md` sigue dando 1; `go test ./tests/content/...` pasa.

### T6 — Sitio y hogares canónicos

- [x] Las páginas de research, plan, build y verify y `instruction-resources.md` describen el comportamiento nuevo

**Closes:** AC10

**Depends on:** T1, T2, T3, T4, T5, T7

**Locations:** `site/src/content/docs/flows/research.md`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/overview.md`, `_support/docs/architecture/instruction-resources.md`

**Execution:** `hive-write-docs`

**Changes:** cada página describe el comportamiento y enlaza a la skill, sin copiar sus reglas; `instruction-resources.md` añade la sección de research a sus hogares canónicos. En `overview.md`, la fila de research nombra la oferta de propuestas entre lo que pregunta. `skills.md` se confirma sin cambio o se ajusta en la misma tarea si enuncia algo que cambió.

**Verification:** `rg -n -i "screen example|approved example" site/src/content/docs/flows/research.md site/src/content/docs/flows/plan.md site/src/content/docs/flows/build.md site/src/content/docs/flows/verify.md` muestra las cuatro páginas; `rg -n -i "proposal" site/src/content/docs/flows/overview.md` muestra la fila de research; `rg -n "Examples for a screen" _support/docs/architecture/instruction-resources.md` muestra el hogar.

### T7 — La delegación desatendida no responde la oferta ni aprueba un ejemplo

- [x] `unattended-delegation` nombra las dos excepciones

**Closes:** AC13

**Depends on:** none

**Locations:** `content/skills/unattended-delegation/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en el paso 5, dos excepciones más a «Take the option you would recommend for every other question», con puntero a la sección de research: con la delegación activa no se hace la oferta de propuestas, y la aprobación del ejemplo de un pedido de solo diseño queda sin responder y se reporta como pendiente. No cambia la excepción de la primera vista ni lo demás del paso.

**Verification:** `rg -n "Examples for a screen" content/skills/unattended-delegation/SKILL.md` muestra el puntero, y la lectura de esa línea confirma las dos excepciones; `wc -c content/skills/unattended-delegation/SKILL.md` da 16,000 o menos; `go test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...`, `go vet ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Sizes | Final candidate, locally | `global.md` sigue en 45,095 bytes; cada `SKILL.md` y referencia en 16,000 o menos | Covered by the implementation; un cambio en `global.md` lo decide el usuario |
| Places of the sweep | Final candidate, locally | Cada lugar de `design.md` «Impact» editado o confirmado sin cambio, listado en el reporte y en el pull request | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` de Claude Code sobre el diff completo (D4 A) | The review answer |

## Checkpoint

- Updated: 2026-11-09
- Stage: flow-verify · 1
- Last commit: f55eba4c
- Open findings: none
- Next: Read the results of the full suite (go test ./..., go vet ./...) and of /code-review; fix in one round; then push, open the pull request, merge to development and run hive update (D3 B)

### Evidence

- ac-walk: rg probes for AC1-AC13 on the candidate files, every count as expected · f55eba4c
- sizes: global.md 45,095 bytes unchanged against origin/development; largest SKILL.md flow-close 15,995 (not touched), largest touched flow-build 13,619 · f55eba4c
- go test ./tests/content/... ok and python3 -m unittest discover -s tests/skills: 89 tests OK · on the working tree before ca36e5cc, same content

## Review status and progress

2026-11-09 — Plan escrito.
2026-11-09 — Revisión 1 del plan: T1 gana la referencia de delegación, los punteros del paso 0 y las dos frases del rol; T5 cambia la frase existente del revisor; las verificaciones observan cada cláusula. Ninguna tarea estaba hecha.
2026-11-09 — Revisión 2 del plan: T1 deja intacto el párrafo final del rol y gana las cláusulas de autorización y de cierre; T4 y T5 definen la diferencia y su destino; T6 gana `overview.md`; T7 es nueva. Ninguna tarea estaba hecha.
2026-11-09 — T1 · 6e36bdfc · unverified: none
2026-11-09 — T2 · bad61441 · unverified: none
2026-11-09 — T3 · ca36e5cc · unverified: none
2026-11-09 — T4 · ca36e5cc · unverified: none
2026-11-09 — T5 · ca36e5cc · unverified: none
2026-11-09 — T7 · 49b6075f · unverified: none
2026-11-09 — T6 · f55eba4c · unverified: none
