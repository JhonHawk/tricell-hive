# #311 — Research mínimo como entrada, plan como el cómo y cierre con pie

| Campo | Valor |
| --- | --- |
| Estado | In review · PR #352 |
| Tracker | [#311](https://github.com/JhonHawk/tricell-hive/issues/311) · abierto · lo absorbe este cambio |
| Git | tricell-hive · base `development` en `15863834` al planear; `b35d7544` tras integrar `development` dos veces · rama propuesta `feat/gh-311-minimal-research-entry` |

## Decisiones

2026-11-03 (sesión Claude Code `335e6514`)
- ¿Qué pasa con el visor en research? Sale del research; el trabajo del visor se crea al entrar a plan o build con una nota de casi un párrafo sobre la sesión, no con el documento de investigación.
- ¿Cómo entra un pedido de trabajo? Todo pedido de trabajo entra a un research mínimo en el hilo principal: intención, exploración proporcional, causas probadas, cobertura, y respuesta, build o plan. (D1 A)
- ¿Qué resuelve cada etapa? Research resuelve el qué: pasar de lo que el usuario pide a lo que necesita, que a veces coincide y a veces no, como lo hace el brainstorming. Plan resuelve el cómo. Un pedido directo, como «agrega este usuario al repo de git», se revisa por bloqueos y se hace. (D2, respuesta libre)
- ¿Cuándo pasa a plan? Solo si el cómo tiene una decisión del usuario o el trabajo no se puede retomar con el pedido y el `git diff`; si no, del research a build. La lista de revisión deja de ser filtro de plan. (D2 A)
- ¿Cómo termina un análisis? Si el usuario pidió analizar, explicar o saber por qué, en los hallazgos, con un pie de opciones sin flujo recomendado; recomienda un flujo solo si pidió un cambio. (D3 A)
- ¿Cómo se cierra un trabajo? Al integrar se mueve el ticket y sale el reporte con un pie «cerrar (recomendado)» y alternativas del contexto (resolver hallazgos en esta rama, explicar uno, siguiente ticket, seguir); lo destructivo espera la respuesta; el pie es obligatorio para no repetir #208. (D4 A)
- ¿Cómo pregunta research por la intención? Una pregunta abierta por mensaje, en texto, con su razón y la salida «o arranco ya», antes de las tarjetas; las tarjetas quedan para elegir. (D5 A, implícita en «prácticamente lo que hace el brainstorming»)
- ¿Quién hace las búsquedas pesadas? `hive-research` investiga y sintetiza; el explorador integrado del host solo ubica archivos y su resultado se verifica. (D6 A)
- ¿Qué se quita de Context7 en las herramientas? Nada: se quedan `hive setup` y el instalador opcional, que deja las reglas por defecto de Context7; en el contenido de Hive solo una o dos líneas en la guía global que invitan a usarlo. (D7, respuesta libre)
- ¿Cuándo corre la limpieza pendiente sin elegir cerrar? Al elegir cerrar, al empezar otro trabajo (primero el cierre del anterior, dicho en una línea) o al terminar la sesión; si la sesión se corta, la recoge el siguiente cierre o higiene. (D8 A)
- ¿Hay que escribir `/flow-build` cuando el research dejó resuelto el cómo de un cambio pedido? No: sigue a build en el mismo turno; el comando escrito queda para research→plan y plan→build. (D9 A)
2026-11-04 (sesión Claude Code `f9f55a91`, tras ver cinco carpetas sin versionar en cam-specs)
- ¿Qué pasa con una carpeta de cambio cuando el desarrollo está completo? Se commitea aunque queden tareas abiertas: «es mejor tenerlo y editarlo que perderlo por algún accidente». (regla del usuario)
- ¿Qué pasa con lo que queda abierto a propósito? Queda como comentario en el ticket y en Engram, y `flow-status` lo lista desde la memoria; una vez resuelto, el registro se borra. (regla del usuario)
- ¿Cuándo se commitea la carpeta por primera vez? Con la entrega del código: en la rama del trabajo antes del push cuando `Specs` vive en el mismo repositorio; en el repositorio de specs al llegar al punto de completitud cuando vive aparte. (D26 A)
- ¿`flow-status` busca carpetas sin versionar en el repositorio de specs? Sí, con un `git status` filtrado en `Specs`. (D27 A)
- Supuestos sin objeción: el registro es uno por trabajo, `pending/<change-id>` (sin carpeta, la clave del ticket o un slug corto, como el `change-id` de `flow-plan`) con «Pending:» en el título, y dice qué falta, por qué, quién y qué lo destraba, nombrando los tickets o condiciones; se actualiza al resolver un pendiente y se borra con borrado suave (`engram delete <id>`; la documentación de Engram dice que oculta el registro de búsquedas y contexto) cuando no queda ninguno, también al abandonar el trabajo; el comentario en el ticket lo autoriza la guía global como el movimiento del ticket, y sin ticket no hay comentario; una tarea operativa no lo escribe; una promoción por la ruta declarada no es pendiente, porque el estado del ticket ya la registra; los pasos diferidos del cierre en dos partes no van en `pending/`; la carpeta se archiva al cerrar como hoy, aunque queden pendientes, porque estos ya viven en el ticket y en Engram (revisión del plan, hallazgos 1, 6 y 7).

- Supuestos sin objeción: «Scope changes» y la regla de tickets usan el nuevo criterio de cambio pequeño; bajo delegación desatendida el cierre corre completo (decisión del 2026-10-30: la delegación termina con el trabajo); la excepción de preguntas abiertas vive en «Questions and routes»; la línea de Context7 del rol `hive-research` pasa a la guía global, que es su única casa.

## Pedido

Del usuario, 2026-11-03: «Abuso de modo plan… muchas solicitudes entraban directo a build sin plan, y ahora casi todo pide plan»; «Propuesta de entrar a plan sin definición… ni siquiera terminé de definir qué es lo que quiero»; «Propuesta de cierre apresurada… al acabar su tarea de inmediato solicita el cierre y lo ejecuta… tal vez solo pongamos el footer como en plan o build para recomendarlo, pero también poner alternativas dinámicas al contexto de revisar hallazgos, explicar alguno, buscar más tickets». Después: «el default debería ser research y de ahí definir si se necesita un plan o ejecutar directo build».

#311 (leído 2026-11-03): research lee el proyecto primero, pregunta abierto una vez por mensaje y devuelve lo entendido por partes separando lo dicho de lo propuesto; plan presenta 2–3 enfoques en un mensaje propio antes de las tareas, con uno recomendado, registrado en `design.md` o `proposal.md`. Fuera de #311: skill o rol nuevo, mockups en navegador, spikes, variantes en paralelo (#188). Su AC4 mantiene los cambios pequeños directo a build.

## Objetivo

Hoy un pedido de trabajo va a plan salvo que cumpla un filtro estricto (una tarea y nada de la lista de revisión), cada reporte de research termina recomendando plan aunque el usuario solo pidió analizar, y `flow-close` corre solo tras el merge y ofrece una tarjeta fija. Al terminar, todo pedido de trabajo entra a un research mínimo que define el qué con el usuario y decide: hacerlo, build en el mismo turno, o plan solo cuando el cómo tiene decisiones; un análisis termina en los hallazgos; el cierre termina en un pie con alternativas y lo destructivo espera la respuesta.

## Alcance y aceptación

Entra: la guía global, `flow-research`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy`, `git-workflow`, `flow-walkthrough` y `viewer-publication.md`, el rol `hive-research`, las páginas del sitio de flujos, `README.md`, `llms.txt`, la documentación de arquitectura que nombra estas reglas, los detectores de `tests/pilot/` que dependan de ellas y `CHANGELOG.md`. Se sumaron durante las sesiones reales del 2026-11-03: `flow-status`, `security-advisory`, `test-audit`, `unattended-delegation`, los roles `hive-review-*` y `hive-verify-task` (vía la referencia), `AGENTS.md` («Maintainer viewer restart»), `host-limits.md` y la referencia nueva `flow-build/references/delegation.md`.

No entra: el código del visor y su protocolo (la etapa `research` sigue existiendo para otros productores), `tooling/` (incluido `hive setup`), las reglas que instala Context7, los casos de `tests/fixtures/` (pilotos en pausa) y los pilotos.

- AC1 — La guía global manda todo pedido de trabajo a `flow-research` y define que va a plan solo si el cómo tiene una decisión del usuario o el trabajo no se puede retomar con el pedido y el `git diff`; ninguna regla usa la lista de revisión para mandar a plan · falso hoy porque `content/guidance/global.md` «Proportionality» manda a build solo un cambio que «fits in one task, and touches none of the review list».
- AC2 — `flow-research` tiene una vía para el pedido completo: revisa bloqueos y, si es un cambio al repositorio, pasa a `flow-build` en el mismo turno sin comando escrito; si no deja cambio, hace la acción y dice qué hizo · falso hoy porque su paso 5 siempre termina con el bloque de siguiente paso y el comando que el usuario escribe.
- AC3 — `flow-research` lee antes de preguntar, devuelve el pedido (qué, para quién, por qué, más conflicto, premisa falsa o camino más barato) y, cuando falta algo que cambia el resultado, hace una pregunta abierta por mensaje en texto con su razón y la salida «o arranco ya»; «Questions and routes» permite esa pregunta fuera de la tarjeta · falso hoy porque `flow-research` paso 1 solo dice «Ask the user only for intent that changes the conclusion» y toda pregunta va por la tarjeta.
- AC4 — Un pedido de analizar, explicar o saber por qué termina en los hallazgos con un pie de opciones sin flujo recomendado; un flujo se recomienda solo cuando el usuario pidió un cambio · falso hoy porque «Questions and routes» y `flow-research` paso 5 ponen la ruta siempre que «the findings leave work».
- AC5 — `flow-research` no publica en el visor; `flow-walkthrough` da sesión de visor desde plan o build, creada con un resumen de un párrafo; `viewer-publication.md` no tiene snapshot de research ni pregunta de abrir el visor tras un reporte de research · falso hoy por `flow-walkthrough/SKILL.md:8`, `viewer-publication.md:35` y `:47`.
- AC6 — `flow-research` delega a hijos `hive-research` cuando un aspecto necesita más de 5 lecturas grandes (documento completo o más de 100 líneas, definido en «Proportionality»; criterio actualizado en d9b51eb9 y movido a la guía global en 8812fe1c) y usa el explorador integrado solo para ubicar archivos, verificando lo que devuelve; nombra cada sistema que toca y si lo buscó; abre la fuente de cada afirmación que decide; busca un ticket existente; crear un ticket es una acción suelta que se pregunta · falso hoy porque no pide cobertura por sistema ni búsqueda de ticket, y sus rutas incluyen abrir tickets.
- AC7 — `flow-plan` presenta 2–3 enfoques en un mensaje propio antes de las tareas cuando el cómo tiene más de una forma razonable, con uno recomendado, y los registra en `design.md` · falso hoy porque `flow-plan` no tiene paso de enfoques.
- AC8 — Tras integrar (verify paso 6, build paso 5 para cambio mecánico o tarea operativa, `git-workflow` paso 10, `flow-deploy`), el agente mueve el ticket, reporta y termina con el pie «cerrar (recomendado)» y alternativas del contexto; cierre del visor, archivo del registro, limpieza de rama y worktree y escritura de memoria corren solo al elegir cerrar, al empezar otro trabajo o al terminar la sesión; bajo delegación desatendida el cierre corre completo · falso hoy porque `flow-verify/SKILL.md:19` dice «Then run `flow-close`» y la pregunta de cierre es una tarjeta fija de tres opciones.
- AC9 — La guía global tiene una o dos líneas que invitan a usar Context7 para documentación de librerías, frameworks y APIs; el rol `hive-research` ya no lo menciona; `tooling/` no cambia · falso hoy porque la guía global no menciona Context7 y `hive-research.md:14` sí.
- AC10 — El README, `llms.txt` y las páginas del sitio de flujos muestran research como entrada, el fin de un análisis en los hallazgos y el cierre con pie · falso hoy porque el diagrama de `README.md:317-327` dice «A small authorized change → Build» y `site/src/content/docs/flows/close.md` describe el cierre automático.

Se sumó el 2026-11-04 (tarea agregada, D26–D27): `git-workflow` paso 4, `flow-close` pasos 3, 4 y 7, `flow-plan/references/plan-handoff.md`, `flow-status`, la guía global («Scope and authorization», «Preservation») y las páginas `plan`, `close` y `status` del sitio.

- AC11 — La carpeta de cambio se commitea con la entrega del código: con `Specs` en el mismo repositorio, en la rama del trabajo antes del push (con worktree, la carpeta se mueve a la misma ruta del worktree antes del primer commit y se edita solo ahí); con `Specs` aparte, en ese repositorio al llegar al punto de completitud, incluida la parada de `ready`, entregada según su `Delivery` o el modo del trabajo; bajo `hold` queda sin commit · falso hoy porque `plan-handoff.md:30` dice «Leave a change folder uncommitted» y `flow-close` paso 4 la versiona solo al archivarla.
- AC12 — Cuando un trabajo deja algo abierto a propósito (una tarea posterior al merge o una fase siguiente; no una promoción por la ruta declarada ni los pasos diferidos del cierre), su cierre, también el de pausa, lo escribe en un registro de memoria `pending/<change-id>` titulado «Pending: <change-id>» con qué falta, por qué, quién y los tickets o condiciones que lo destraban, y lo comenta en el ticket cuando hay uno; la guía global autoriza ese comentario · falso hoy porque `flow-close` paso 7 solo dice «Write the outcome to persistent memory» y `global.md:22` autoriza solo mover el ticket.
- AC13 — `flow-status` lista los registros «Pending:» del proyecto y las carpetas de cambio sin versionar bajo `Specs`, y marca como desactualizado el registro cuyo ticket está en un estado terminal y cuyas condiciones se cumplen, o cuya carpeta ya está en `archive/` · falso hoy porque `flow-status/SKILL.md:8` lee solo `## Hive`, el tracker y Git del repositorio actual.
- AC14 — Cada cierre busca los registros «Pending:» que nombran su ticket o su `change-id`, los actualiza y borra con borrado suave el que queda sin pendientes; el cierre por abandono borra el registro del trabajo; sin Engram, borra la entrada en la memoria del host o dice que no pudo; la guía global permite ese borrado · falso hoy porque «Preservation» preserva los datos de memoria y ninguna regla borra un registro.
- AC15 — Las páginas `plan`, `close` y `status` del sitio y `CHANGELOG.md` describen la carpeta versionada con el código, los pendientes en ticket y memoria y su listado en status · falso hoy porque `site/src/content/docs/flows/plan.md:45` dice «A change folder stays uncommitted while the work runs».

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- Research de la sesión `335e6514` (2026-11-03) sobre 27 sesiones del 2 y 3 de noviembre (Claude Code, Codex, OpenCode, Grok): H1–H16 en la conversación y en Engram `decision/research-minimal-entry`.
- `v0.6.0`: `flow-build` entraba con «a request the user made»; `flow-research:46` decía «For an investigation-only request, stop at the findings» (quitado en `adf076cf`); el bloque «▶ Next step» llegó en `c0424c02`; el filtro estricto en `99b685cd` (#268) y `d0e7c4b7`.
- `flow-research/SKILL.md` pesa 5.075 bytes; cada research obliga además a leer `flow-walkthrough` (1.778) y `viewer-publication.md` (~13.000).
- Las specs no describen el ruteo entre flujos. `walkthrough-viewer` sí pedía preguntar si abrir el visor tras un reporte de research: el delta `specs/walkthrough-viewer/spec.md` modifica ese requisito (hallazgo de `/code-review`).

## Entrega

- Modo `ready` (D10 C): revisión, push y PR no draft hacia `development`, espera del CI y respuesta a sus revisiones; se detiene con el PR abierto para que el usuario haga el merge. Nunca merge, auto-merge ni cola de merge.
- Revisión: Native, Claude Code `/code-review` sobre el candidato antes del push (D11 A).
- T8–T11 (2026-11-04): `ready` dentro del PR #352 (D28 A) y Native `/code-review` (D29 A).
- No autoriza: el merge, `hive update`, un release, cerrar #311 antes del merge ni escribir en otros repositorios.

## Estado de la revisión y avance

2026-11-03 — Plan escrito. Sin `hive-review-plan`: no cambia un contrato, datos guardados, seguridad ni dependencias.
