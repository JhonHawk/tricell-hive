# business-rules-record — Las reglas de negocio que un trabajo adopta quedan registradas, en tarjetas junto al diseño y en su ticket

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | ninguno |
| Git | `tricell-hive`, base `development` en `b9113a3c`, rama propuesta `feat/business-rules-record` |

## Decisiones

2026-11-13
- **D1 A** — ¿Cuándo escribe Hive las reglas de negocio en el ticket? En cada aprobación y al cierre: al aprobar el ejemplo de diseño, al empezar el build (la respuesta que aprueba el plan) y al cierre con lo que salió en build y verificación.
- **D2 A** — ¿Hive escribe esa sección sin preguntar cada vez? Sí: es un permiso permanente de la guía global, igual que mover el ticket, para todos los proyectos que usan Hive.
- **D3 (respuesta libre)** — ¿Qué pasa cuando plan o build cambian una regla que ya está en una tarjeta o en un frame aprobado? Se actualiza, pero nunca se borra: la regla anterior queda visible como cancelada, con el motivo y la regla que la reemplaza, para que nadie crea que se pisó el criterio de quien la escribió. Un frame que quedó contradicho no se redibuja fuera de la etapa de diseño: queda anotado y se presenta al cierre. Lectura del agente, mostrada al usuario y no corregida.
- **D4 A** — ¿Dónde viven las reglas en el ticket? En ambos lugares: una sección propia en la descripción con las reglas vigentes y las canceladas con su motivo, sin tocar el resto del texto, y un comentario por cada escritura que dice qué se agregó y qué se canceló y por qué. Además, research lee los comentarios del ticket.
- **D5 B** — ¿Cómo se entrega este cambio? Automática: revisión, push, pull request y merge sin detenerse para validación.
- **D6 A** — ¿Qué revisión de código corre sobre el candidato antes del push? La nativa de Claude Code, `/code-review`.

## Solicitud

> «analiza la sesión en curso de fleetsystem, las dos, a nivel diseño, creo que parte de la skill debe ser que se adjunten las reglas de negocio adhoc a la vista, y si se toman algunas que no vengan en la doc o ticket, que esta misma skill se encargue de actualizar (aunque no se si esto segundo deba venir en la skill de figma o sea mas amplio, porque puede pasar desde el plan o build)»

Tras el análisis, el usuario eligió preparar el cambio completo: tarjetas en `flow-figma-design` y un registro común con escritura al ticket. Sin ticket.

Fuera de alcance: la falta de una salida «ver alternativas» en la ronda aprobar/corregir, el registro de FSR-860 que no se actualizó al aprobar, la numeración de decisiones que se traslapa entre sesiones y los cambios a la librería de Figma sin preguntar (hallazgos H6 a H9 del análisis).

## Objetivo

Hoy solo `flow-figma-design` trata reglas de negocio, y solo las que nadie ha resuelto: las anota en `examples.md`, que vive en `_support/workspace/`. Ninguna regla las dibuja junto al frame ni las lleva al ticket, y plan, build y verificación no tienen dónde registrar una regla que aparece tarde. La única escritura autorizada al tracker es mover el ticket y comentar lo pendiente al cierre.

Con el cambio: una referencia única define qué es una regla de negocio, dónde se registra en cada trabajo y cuándo llega al ticket; la etapa de Figma dibuja una tarjeta de reglas junto a cada frame; y cualquier flujo que adopta, cambia o cancela una regla la registra en el mismo turno, conservando la anterior como cancelada con su motivo.

## Alcance y aceptación

Dentro: la referencia nueva `business-rules.md` y la de tarjetas `rule-cards.md`; la guía global; los punteros en research, Figma, plan, build, verificación y cierre; el rol `hive-review-ux` y el brief del revisor; la plantilla de `proposal.md`; el delta de `workflow-lifecycle`; la referencia `outside-the-repository.md`; las páginas del sitio, `README.md`, `llms.txt` e `instruction-resources.md`.

Fuera: código del manager (`tooling/`, `integrations/`), detectores de `tests/pilot/`, casos de regresión nuevos, una fila nueva en la rúbrica de revisión de sesiones (la decide el usuario), y cualquier corrida de un host para probar el comportamiento (los pilotos siguen en pausa).

Los criterios se observan en el texto distribuido; que un agente los cumpla en una sesión real solo se verá en sesiones posteriores.

- AC1 — Una referencia única define la regla de negocio, el registro de reglas de cada trabajo con el formato de su línea y sus estados, y la escritura al ticket (sección de la descripción más comentario, sus momentos, quién escribe y qué pasa sin ticket) · falso hoy porque la definición solo existe en `content/skills/flow-figma-design/SKILL.md:25`, limitada a lo que un frame muestra, y ningún archivo describe una escritura de reglas al ticket · Observed in: `content/skills/flow-plan/references/business-rules.md`
- AC2 — La guía global autoriza al hilo principal a escribir la sección de reglas en la descripción del ticket y su comentario sin pregunta, aunque cambien datos que la tarea no creó, y a corregir las filas de tarjetas que el trabajo creó; nombra sus tres momentos y apunta a la referencia; la regla de `Other trackers` y la lista de efectos ya autorizados de `outside-the-repository.md` cubren esa escritura · falso hoy porque `content/guidance/global.md:21` autoriza solo mover el ticket y el comentario de pendientes, `:113` dice «move and comment» y `outside-the-repository.md:19` exime solo «the ticket's moves» · Observed in: `content/guidance/global.md`, `content/skills/flow-plan/references/outside-the-repository.md`
- AC3 — Cada momento de escritura está en el paso que lo ejecuta, como puntero a la referencia: la aprobación del ejemplo en `flow-figma-design`, el arranque en `flow-build` junto a mover el ticket, y el paso que mueve el ticket en `flow-close` · falso hoy porque esos tres pasos no mencionan reglas ni la descripción del ticket · Observed in: `content/skills/flow-figma-design/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-close/SKILL.md`
- AC4 — Una regla que se decide, se asume o se cancela después del diseño (en el plan o su revisión, en el build, también por un hijo, o en la validación) entra al registro en el mismo turno; la regla que reemplaza conserva la línea anterior marcada `cancelled` con fecha, motivo y reemplazo · falso hoy porque `flow-plan`, `flow-build` y `flow-verify` no nombran las reglas de negocio fuera de la lectura de `examples.md` en `flow-plan/SKILL.md:22` · Observed in: `content/skills/flow-plan/references/business-rules.md` y los punteros de `content/skills/flow-plan/SKILL.md`, `content/skills/flow-build/SKILL.md` y `content/skills/flow-verify/SKILL.md`
- AC5 — `flow-figma-design` dibuja junto a cada frame que dibuja una tarjeta de anotación con las reglas que el texto del frame no dice, cada fila con su origen, y anota el nodo de cada tarjeta en el registro; un frame sin reglas que anotar no lleva tarjeta · falso hoy porque ningún paso de la skill dibuja anotaciones · Observed in: `content/skills/flow-figma-design/references/rule-cards.md` y su puntero en `content/skills/flow-figma-design/SKILL.md`
- AC6 — Cuando un flujo posterior cambia una regla que está en una tarjeta, corrige la fila (la anterior como cancelada con su motivo, más la nueva) si el host tiene conexión que escribe en Figma, y si no anota `card stale`; un frame contradicho se anota `frame stale`, no se redibuja fuera de la etapa de diseño y queda entre lo que el trabajo deja abierto al cierre · falso hoy porque solo `flow-figma-design` escribe en el archivo de Figma y nada registra una tarjeta o un frame desactualizado · Observed in: `content/skills/flow-figma-design/references/rule-cards.md`
- AC7 — Research lee el ticket con sus comentarios · falso hoy porque `content/skills/flow-research/SKILL.md` paso 1 dice «ticket» sin nombrar los comentarios · Observed in: `content/skills/flow-research/SKILL.md`
- AC8 — El revisor de UX revisa las tarjetas en la segunda pasada, sin haberlas visto en la primera: reporta la fila que repite el texto del frame, la que contradice al frame y la regla del registro que falta en su tarjeta · falso hoy porque `content/agents/review/hive-review-ux.md:10` y «Reviewer brief» no nombran tarjetas · Observed in: `content/agents/review/hive-review-ux.md` y `content/skills/flow-build/references/delegation.md`
- AC9 — El sitio, `README.md` y `llms.txt` describen las tarjetas de reglas y la escritura de reglas al ticket donde hoy describen la etapa de Figma, el build, el cierre y los trackers · falso hoy porque `site/src/content/docs/flows/close.md`, `flows/build.md`, `flows/figma-design.md` y `getting-started.md:44` solo nombran mover y comentar el ticket, y ninguna página nombra tarjetas · Observed in: `site/src/content/docs/`, `README.md`, `llms.txt`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

Leído el 2026-11-13 en `development` (`b9113a3c`):

- `content/skills/flow-figma-design/SKILL.md` (15,456 bytes de 16,000): pasos 2, 5, 6, 7 y 8 tratan las líneas `rule:` de `examples.md`; la línea 8 dice que la etapa escribe solo en el archivo de Figma y en la carpeta del trabajo.
- `content/guidance/global.md:21` y `:113`: las únicas escrituras al tracker autorizadas son mover el ticket y comentar lo pendiente; los hijos nunca escriben trackers (`:23`).
- Bytes libres: `flow-close/SKILL.md` 10, `flow-build/SKILL.md` 529, `flow-figma-design/SKILL.md` 544, `flow-plan/SKILL.md` 2,516, `flow-verify/SKILL.md` 4,308, `flow-research/SKILL.md` 7,135. `tests/content/skill_size_test.go` limita cada `SKILL.md` y cada referencia a 16,000 bytes.
- `content/skills/flow-plan/scripts/check_change.py` valida criterios y la sección de progreso, no los encabezados de `proposal.md`.
- `_support/openspec/specs/workflow-lifecycle/spec.md`: requisitos «Business rules that appear while drawing» (515) y «Tickets follow the tracker their key declares» (27).
- El término «business rule» y las líneas `rule:` solo aparecen en `flow-figma-design/SKILL.md`, `flow-plan/SKILL.md:22`, `delegation.md:42` y `hive-review-ux.md:10`. `examples.md` no figura en `design-record.md`, `artifact-layout.md`, el sitio ni los tests.

Observaciones que motivan el cambio (hueco sin regla y función nueva; las adopta la decisión del usuario):

- Claude Code 2.1.323, `claude-opus-5-5`, sesión `62cf4d39-a32c-4b21-afcf-9839dde244f8`, fleetsystem, 2026-11-13: el usuario pidió tarjetas de reglas junto a los frames (38 tarjetas, 89 filas, aprobadas sin corrección) y a las 08:51Z pidió llevar las reglas a la descripción de FSR-855 a FSR-860; hasta entonces vivían en `examples.md`, Engram y las tarjetas.
- Claude Code 2.1.323, `claude-opus-5-5`, sesión `9ff8a1c8-a4af-46e0-bef8-a7bc544214da`, fleetsystem, 2026-11-13: de 20 decisiones de regla fuera del PDF y del ticket de FSR-859, 3 nacieron en research, 3 en diseño, 12 en el plan y 2 en el build; ninguna se escribió en Jira ni en Figma. El frame 310:1468 quedó contradiciendo una decisión del plan y una tarjeta siguió en «Por decidir».

## Effects

None: the tasks take no effect on an external system.

## Entrega

Automática (D5 B): revisión de código, push de la rama, pull request a `development` y merge sin parada de validación; el punto final es el merge. Revisión de código: `/code-review` de Claude Code (D6 A), antes del push. En el checkout del mantenedor, el merge va seguido del refresco de la instalación local que fija el `AGENTS.md` del repositorio. No autoriza un release ni la promoción a `master`.

## Estado de revisión y progreso

2026-11-13 — Plan escrito. Revisión del plan (`hive-review-plan`): 11 hallazgos, todos aplicados. Cambió el delta (dos requisitos existentes pasan a MODIFIED para no contradecir los nuevos), fijó cuál registro gobierna, declaró la excepción a la regla de efectos críticos y el permiso para corregir tarjetas, eligió el recorte de `flow-close`, dio a cada tarea comprobaciones de su criterio y sacó la fila de la rúbrica.

2026-11-13 — Revisión tras `/code-review` (10 hallazgos, todos aplicados): la sección del ticket lleva línea de cierre y no se edita la descripción si sus marcas faltan o se repiten; la sección y los comentarios que Hive escribe no cuentan como fuente; el plan copia también las líneas de control del ticket y deja fuera las `stated`; el trabajo sin carpeta de cambio usa el archivo de checkpoint desde que arranca su plan o build; la escritura del cierre pasa al punto de finalización; se omite una escritura cuando la sección no cambiaría; las tarjetas de alternativas descartadas se eliminan; `flow-close` paso 5 apunta entero a «Preservation». Cambia el delta (momento del cierre, límites de la sección, criterio de omisión); no invalida tareas.
