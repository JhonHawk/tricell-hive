# Diseño

## Contexto verificado

Leído en `development` `70d2bb13` el 2026-10-25. Los hallazgos vienen de cuatro auditorías `hive-research`; los que sostienen las decisiones los revisó el hilo principal abriendo sus locators.

- **Revisión dedicada sin momento fijo.** «as the delivery mode sets» o «at the point the delivery mode sets» aparece en `content/skills/flow-verify/SKILL.md:30,46,62`, `flow-verify/references/review-and-gates.md:9`, `flow-verify/references/ui-gate.md:38` y `flow-plan/references/delivery-decisions.md:26`. La tabla de modos (`delivery-decisions.md:9-12`) solo dice «the agreed review» para `interactive` y «agreed checks/review» para `automatic`. `review-and-gates.md:9-10` nombra dos momentos: el candidato local y la publicación.
- **`automatic` con UI.** `delivery-decisions.md:10` dice: «when project guidance requires one for the change, such as a visual check of a changed screen, mark automatic unavailable». En cambio, `project-settings.md:9` dice: «under `Delivery: automatic`, `none` merges a UI change with only the implementer's checks and no person». `flow-walkthrough/SKILL.md:24` pregunta por la primera vista sin condición de modo, y `flow-build/SKILL.md:12,62` exige que esa vista quede aceptada.
- **Punto de cierre posterior.** `global.md:34` dice «unless the project declares a later completion point», y `project-settings.md` no lo define. Tres proyectos lo declaran en prosa: CAM (`comsys/cam/AGENTS.md:45`), Carvento (`comsys/carvento/AGENTS.md:58`) y Facti (`visooal/facti/AGENTS.md:86`), con «the completion point is QA».
- **`direct-base`.** Lo declaran `cam-specs`, `carvento-specs` y `cam-frontend-mocks` (`AGENTS.md` de cada uno). El TUI y `doctor` conocen las claves de `## Hive` (`tooling/cli/tui_project_view.go:313`, `doctor_project_test.go:492`); una clave nueva exige cambios en Go.
- **Preguntas en secuencia.** `flow-plan/references/questions.md:25-27` agrupa las preguntas independientes. `:29-31` pide una sola pregunta para elecciones en secuencia «even when a later step depends on an earlier one's result».
- **Tamaño.** `flow-close/SKILL.md` pesa 15 955 bytes y `flow-plan/SKILL.md` 15 799. `tests/content/skill_size_test.go:13` fija el límite en 16 000. `global.md` (38 338 bytes) no tiene límite; `tests/content/budget_test.go` solo lo registra.
- **Portal.** `master` = `v0.5.0` (`8c2882bd`, 2026-10-23). Hay 284 commits de `development` sin publicar. `https://hive.tricell.tech/flows/verify/` respondió 404 el 2026-10-25, y `README.md:344` la enlaza. El sitio se publica solo con una release (`CONTRIBUTING.md:136-137`; `_support/openspec/specs/documentation-site/spec.md:9-18`).
- **Pruebas.** No hay CI. `go test ./...` cubre `tests/content` (tamaños, enlaces de la guía global, deltas aplicados) y `tooling/management` (`TestRepositoryCatalogueInstructionReferences`, enlaces de recursos). Ninguna prueba comprueba las anclas de encabezado ni la redacción de las reglas.

## Enfoque

Cada regla queda en un solo hogar y los demás lugares apuntan a él (`AGENTS.md`, Guidance structure). Para cada área:

### Entrega y revisión (`delivery-decisions.md`, hogar)

- La tabla de modos y un párrafo nuevo fijan el momento de la revisión dedicada (D4-A). Corre sobre el candidato local, después de la suite completa, en este orden según el modo:
  - **`interactive`:** antes de pedir la validación.
  - **`automatic` y `direct-base` en un repo de código:** antes del push.
  - **`hold`:** sobre el árbol de trabajo.
  - Con `interactive`, las correcciones que pide la validación reciben una sola revisión, acotada al diff entre el candidato revisado y el aceptado, junto con la compuerta del candidato aceptado (`ui-gate.md:38,42`, `flow-verify/SKILL.md:54`). Ninguna ronda de corrección corre la revisión completa.
  - La regla de promoción queda así: se revisan solo los commits no revisados antes de su integración. La evidencia es el registro de la revisión en el cuerpo del PR o en el registro del cambio (`git-workflow`, que lista las comprobaciones corridas).
  - Una edición posterior del mismo candidato, incluso con alcance añadido, no vuelve a preguntar si `Review` o una respuesta previa cubre la nueva revisión. Esa revisión se acota a lo cambiado desde el commit revisado (`review-and-gates.md:16`). En esa sección, «pull request» pasa a «candidato».
- Cuando `automatic` es el modo o una opción de la ronda, sin `Review` declarado, la pregunta de revisión siempre se hace, salvo para un cambio mecánico. Su opción recomendada siempre es un mecanismo de revisión; «Sin revisión dedicada» sigue disponible pero nunca recomendada (D10-A). La opción `automatic` se redacta para que valga con cualquier respuesta a la pregunta de revisión, por ejemplo «con la revisión que elijas en D<n>» (AC16). El efecto de cada valor de `UI verification` se escribe por modo, de modo que la opción también valga con cualquier modo.
- `automatic` se descarta solo cuando una regla del proyecto exige validación humana para ese cambio, y la línea de entrega la nombra. Se quita el ejemplo de la pantalla cambiada (D5-A).
- Una frase distingue `direct-base` elegido, que mantiene la revisión y la espera de CI, del valor declarado, que concede las excepciones de repo de soporte.
- Las definiciones de `UI verification` (`full`, `functional`, `none`) quedan aquí. `project-settings.md` lista la clave y apunta. Se corrige que `none` «mantiene la entrega humana» bajo `automatic`.
- La regla de revisión del PR de promoción se queda aquí; `flow-deploy` y `review-and-gates.md` apuntan a ella.

### Punto de cierre posterior (`project-settings.md`, hogar; D6-A)

Un proyecto lo declara en su guía, nombrando el ambiente y su condición, como «Done = on `origin/qa`» en CAM y Carvento o `origin/gamma` en Facti. El comportamiento es uno solo:
- La obra se cierra al integrarse en la rama base. `flow-close` la cierra ahí.
- El ticket va al estado intermedio que la guía del proyecto nombre.
- Pasa al estado final cuando se cumple la condición que nombra la guía del proyecto. `flow-deploy` aplica ese movimiento cuando corre la promoción; si la promoción la hizo otra persona o sesión, la siguiente sesión que vea cumplida la condición lo aplica.

`global.md:34` apunta aquí y dice que esa regla autoriza los dos movimientos del ticket, el intermedio y el final (`global.md:29` exige que una regla que autoriza una escritura lo diga). `flow-close` (Entry, When to run it), `reporting.md:102` y `flow-deploy` apuntan aquí.

### `flow-close` (tamaño y finales sin merge)

- «Close the current walkthrough» pasa a `flow-close/references/walkthrough-close.md` y se deja un puntero, para liberar unos 2 000 bytes.
- «When to run it» define tres finales:
  - **`hold`:** sigue siendo un punto de cierre, porque es el destino acordado (`plan-format.md:57`, `flow-close/SKILL.md:18`). El ticket queda en curso porque no se integró. La entrega, que decidiste no hacer, va al reporte como recordatorio, no como pendiente.
  - **PR que espera revisión humana:** la obra llega a su punto de cierre al abrir el PR. El ticket queda en curso hasta el merge, y el reporte lo da como recordatorio.
  - **Después de una promoción:** solo corre la limpieza de ramas y procesos, sin informe de cierre ni pregunta de cierre.
- El paso 5 de «Clean up after a merge» se limita a las ramas de PRs cuyo autor es el usuario autenticado (`gh api user`).
- Una rama remota se borra solo cuando su cabeza remota es igual al `headRefOid` del PR fusionado o es ancestro de la base. Así no se pierden commits pusheados después del merge desde otra máquina o por un coautor (`flow-close/SKILL.md:88,92`).
- La pregunta de qué hallazgos abrir como tickets ofrece también, por hallazgo o en conjunto, resolverlos en esta obra, nombrando la ruta: `flow-build` si es un cambio pequeño y entendido, `flow-plan` si no (D11-A). La pregunta de qué sigue va en un turno posterior cuando sus opciones dependen de esa respuesta (AC16). El hogar es `incidental-findings.md`; `flow-close` (Close question) y `global.md:33` apuntan a él.

### Verify, walkthrough y build

- `flow-verify` apunta al momento de `delivery-decisions.md`. La prueba en vivo considerada innecesaria pasa a ser una línea del reporte con la razón, no una pregunta. `review-and-gates.md` deja un solo límite para la revisión.
- `flow-walkthrough`: bajo `automatic` y bajo delegación desatendida, la primera vista se captura y se publica en el visor. Su ruta va en la línea de avance, sin pregunta (D5-A). Lo mismo vale para la entrega a pedido y para la grabación con `none`. Se agregan las paradas de validación de `direct-base` en un repo de código.
- `flow-build`:
  - Con `interactive`, se detiene antes del push aunque verify se omita, mostrando el diff y las comprobaciones (D7-A).
  - Un cambio mecánico solo recibe la pregunta de entrega.
  - El criterio de delegación apunta a `delegation.md` (D9-A).
  - El puntero a `git-workflow` pasa a la sección de ejecución.
  - El ticket pasa a en curso antes de la pregunta de entrega.
  - Un formulario o diálogo de captura dispara la planificación de UI.
- `delegation.md`, «Split and announce work», queda como el único criterio (D9-A): lo pequeño y entendido queda directo, y se delega una unidad con interfaz fija, sin escrituras superpuestas y con un brief más corto que el cambio. «Launch and wait» se limita al hilo principal.

### Guía global, desatendido, git y deploy

- `global.md`:
  - `:34` apunta al punto de cierre de `project-settings.md`.
  - `:35` incorpora D8-A como invariante mínima (con `Delivery` y `Review` declarados, cuenta el silencio después de la respuesta de ruta) y apunta a la skill.
  - `:84` limita el umbral de seis lecturas a una consulta sin skill y apunta a `delegation.md`.
  - `:117` se reescribe así: «deleting merged work branches from this work or the user's earlier work… never delete a branch with unmerged commits, unless flow-close confirms a squash or rebase merge against its pull request, and never delete a worktree with uncommitted or untracked files».
- `unattended-delegation`: D8-A queda redactado como «después de responder la pregunta de ruta, con cada valor de entrega y revisión que la obra necesita declarado o respondido». Cubre un valor declarado y otro respondido, y obras en varios repos. Se aplica en la `description` (`:3`), en `:8`, en `global.md:35` y en `site/src/content/docs/skills.md:29`. «Silence grants nothing» (`global.md:20`, `questions.md:23`) gana una excepción que nombra esta fuente. Una `UI verification` sin responder se decide con la opción recomendada y se reporta, como ya pide `unattended-delegation/SKILL.md:16` para toda pregunta. Se agregan además el puntero a la regla de la primera vista y el puntero de `--force-with-lease` a `git-workflow`, que pasa a ser el hogar de esa excepción.
- `git-workflow`: la excepción desatendida de `--force-with-lease`, y que sin CI la base desactualizada se resuelve volviendo a correr las comprobaciones locales.
- `flow-deploy`: los punteros al punto de cierre y a la regla de revisión de la promoción, y un solo conjunto de estados de aceptación.
- `artifact-layout.md:43` apunta a `flow-close`, «Clean up after a merge».
- `flow-research` y `security-advisory` apuntan a `delegation.md` en lugar del umbral global.

### Preguntas (`questions.md`, hogar; AC16)

Una condición nueva en «Decide whether a question is needed»: una pregunta cuyas opciones, contexto o pertinencia dependen de la respuesta a otra, como las que cambian el rumbo de una investigación o abren otro esquema de preguntas, se hace en un turno posterior, después de esa respuesta. En la misma ronda van solo las preguntas cuyas opciones valen con cualquier respuesta a las demás. La regla de conjunto en secuencia se limita a las autorizaciones de un objetivo ya decidido. `flow-plan/SKILL.md:22`, que ya dice «asking one separately only when its answer determines which other questions apply», apunta a la condición nueva en lugar de repetirla.

### Roles y skills de soporte

- `diagnosis.md`: el hijo `hive-research` reúne evidencia de solo lectura, y el hilo principal corre los experimentos dentro de sus efectos autorizados.
- `hive-read-state`: acepta «a tracker query by project and states».
- `hive-verify-change`: modo de solo grabar cuando el brief lo limita.
- `ui-gate.md`: el hilo principal construye y sirve la revisión base en un puerto propio y le pasa la URL al revisor.
- `hive-review-ux`: la lectura de browser automation pasa al paso 1, y las colas de los roles de verificación y observación ganan la prohibición de reescribir el árbol con git y la de dejar procesos.

### Documentación y portal

Las páginas describen el comportamiento final con enlaces a las skills.
- `flows/build.md` deja el detalle del visor al README y las reglas a las skills.
- Las páginas de research, plan y close dejan de copiar las reglas de preguntas.
- `skills.md` agrega `flow-walkthrough`.
- Getting started agrega una sección corta de claves de `## Hive` que apunta a `project-settings.md`.
- README y `llms.txt` agregan `hive voice`, `hive project set` y `hive releases`, y `llms.txt` las skills que no son flujos.
- `CONTRIBUTING.md` agrega `pnpm build` antes de `wrangler deploy`.
- `documentation-site` pasa a nombrar las siete páginas de flujo (delta en [specs](specs/documentation-site/spec.md)).

## Decisiones menores

Tomadas sin preguntar; se revierten editando la regla correspondiente.

- **H8:** se mantiene el nombre `direct-base`, porque tres repos lo declaran, y basta una frase para distinguir los dos usos.
- **H15:** «autor del PR igual al usuario autenticado» es el límite verificable más simple. `global.md` ya acota el borrado a «this work or earlier».
- **H17:** se mueve a referencias el texto que solo se lee en un momento: el cierre del walkthrough y el resumen al entregar el plan (este último a `plan-format.md`, «Hand off a ready plan»).
- **AC14:** el detalle del visor se queda solo en el README, porque es la entrada para personas que describe la instalación y el visor.
- **Revisión del plan:** se aceptaron las propuestas de los revisores tal como vinieron. Son la revisión acotada después de la validación, la redacción de la regla de promoción, la redacción completa de D8-A, el borrado de ramas remotas solo con cabeza igual o ancestro, la condición del estado final según la guía del proyecto, `hold` como punto de cierre, y la renumeración de puntos de control. La condición del estado final ajusta la redacción de D6-A, que decía «al verificar la promoción», a la condición de git que ya declaran los tres proyectos. Se informa al usuario en la entrega del plan.

## Hallazgos menores: destino

| Hallazgo | Destino |
| --- | --- |
| Disparadores de planificación de UI distintos (`flow-build/SKILL.md:28`) | Corregir en T4 |
| Texto previo a #169 (`flow-plan/SKILL.md:40`, `plan-format.md:136`) | Corregir en T2 |
| «smaller model» en `plan-execution.md:23` | Corregir en T4: «a smaller model or a lower effort than the implementer's, such as a smaller model of the same family». `integrations/` no se despliega con el contenido, así que no se cita |
| Criterios falsos «at the base commit» antes de registrar la base | Corregir en T2: al planear, se usa la punta actual de la rama base |
| Tarjeta de aprobación con AC sin glosa (`flow-plan/SKILL.md:74`) | Corregir en T2 |
| `browser-automation.md:37,39` remite a reglas inexistentes | Corregir en T4 |
| Instalación global de agent-browser sin verificar la dependencia | Corregir en T4 (puntero) |
| Momento del movimiento del ticket (`global.md:34` frente a `flow-close/SKILL.md:26`) | Corregir en T3: el paso 1 confirma el movimiento hecho al integrar |
| Lockfile con avisos HIGH preguntado en cada repo JS | Descartar: es una decisión de seguridad deliberada (#147) |
| Condiciones no literales (`questions.md:37`, `delivery-decisions.md:34`) | Descartar: piden juicio a propósito, y una fórmula numérica sería peor |
| Concesiones entre sesiones (`effect-authorization.md:28` frente a `global.md:29`) | Corregir en T5: `global.md:29` apunta a la vigencia que fija `effect-authorization.md` |
| Pausa de validación leída como pausa de cierre (`flow-build/SKILL.md:68`) | Corregir en T4 |
| Readiness de `flow-plan/SKILL.md:48` frente a `plan-review.md:22` | Corregir en T2 (puntero a los marcadores del formato) |
| Puntero a `git-workflow` solo en `flow-build/SKILL.md:64` | Corregir en T4 |
| Sin regla de revisión para planes que quedan en la conversación | Corregir en T2: la revisión del plan corre solo para un plan guardado, y un plan en la conversación se revisa a sí mismo con el listón de readiness |
| Ticket en curso sin orden respecto de la pregunta de entrega | Corregir en T4 |
| Descarte de un criterio que vino del ticket sin avisar (`plan-format.md:26`) | Corregir en T2 |
| Borrado de ramas repetido (`flow-close/SKILL.md:79`) | Corregir en T3 |
| «one memory write» frente a `incidental-findings.md:9` | Corregir en T3 |
| Estados de aceptación de `flow-deploy/SKILL.md:10` frente a `:25` | Corregir en T5 |
| Sin disparador de `flow-close` después de una promoción | Corregir en T3 |
| Walkthrough contradictorio bajo `automatic` (`:13,30,34`) | Corregir en T4 |
| `delegation.md` mezcla instrucciones del padre y de los hijos | Corregir en T4 |
| Referencias de nombres sin locator (`hive-design-architecture.md:12`, `hive-build-infra.md:17`) | Corregir en T6 |
| Referencia hacia adelante en `hive-build-data.md:10` | Corregir en T6 |
| Autoridad para commitear (`harness-audit` frente a `workspace-conventions`) | Corregir en T6: «only under the repository's Git rules and an explicit commit instruction» en ambas |
| Lecturas innecesarias en `harness-audit/SKILL.md:12` | Corregir en T6 |
| Scripts sin línea de invocación (engram-init, workspace-archive) | Corregir en T6 |
| Vocabulario de estados de workspace-archive | Corregir en T6: acepta también Completed y la fase abandonada que fija `plan-format.md` |
| Brief de advisory en `flow-security` | Corregir en T6 |
| `flow-report/SKILL.md:10` presupone un registro | Corregir en T6 |
| `hive-refute-claim` pensado solo para código | Corregir en T6 |
| Vocabulario de severidad entre `hive-review-plan` y `plan-review.md` | Corregir en T6 |
| Reporte de backlog frente a la pregunta de ruta | Corregir en T6: un reporte de estado termina con recomendaciones; la pregunta de ruta solo aparece cuando el usuario pidió qué tomar |
| `uv run` sin verificación de dependencias (`native-tools.md:19`, `starlight-docs-site/SKILL.md:18`) | Corregir en T6 |
| Operación `clone` del manifiesto sin clasificar | Corregir en T6: `clone` es `ask` |
| Colas de los roles de verificación y observación | Corregir en T6 |
| Duplicación de las colas en 10 roles | Descartar: es intencional, porque los hijos pueden no cargar la capa global; se vigila (conteo 1) |
| Nueve roles sin skill que los nombre | Descartar: la selección por descripción es intencional |
| `index.mdx:48` apunta a una sección equivocada del README | Corregir en T7 |
| Encabezados de `index.mdx:104-139` | Corregir en T7 |
| `skills.md:29`, fuentes de activación desatendida | Corregir en T7 |
| «Esperando tu decisión:» literal en `research.md:47` | Corregir en T7 |
| `llms.txt` sin las skills que no son flujos | Corregir en T7 |
| Fecha de revisión vieja en `repository-and-distribution.md:3` | Corregir en T7 |
| Comportamiento de Deploy en `build.md:46,50` | Corregir en T7 |
| `plan.md` no menciona `UI verification` | Corregir en T7 |
| Dos PR por cambio (registro de cierre) | Descartar: es una decisión de diseño vigente |
| `uv` y `pnpm build` en `wrangler deploy` | Corregido por AC14 (T7) |

## Evidencia de sesiones

Revisión de las conversaciones principales de los seis hosts, del 2026-10-01 al 2026-10-25, sin hilos de subagentes ni fixtures de pilotos. La hizo un hijo `hive-research`; el hilo principal comprobó el ejemplo de `c119a25b`. La dependencia se juzgó por el texto de las preguntas y las opciones, con un solo lector. Los scripts están en el scratchpad de la sesión.

- **H18, confirmado.** 20 rondas con una dependencia clara, de 316 rondas con varias preguntas (cerca del 6 %): Claude 17, Grok 1, Pi 1, OpenCode 1. En 8, la respuesta del usuario a la primera pregunta dejó sin sentido o equivocada la segunda. Ejemplos:
  - Claude `c119a25b` (2026-10-24): D25 ofrecía recortes pensados para el límite de 18 000 bytes de D24, y el usuario eligió 16 000.
  - Claude `c574e5a1` (2026-10-23, facti): D12 daba por abierto el ticket de H5, y el usuario eligió «Medir antes en prod».
  - Grok `01a0fa92` (2026-10-02, cam): «Si abriste H3…», y el usuario descartó H3.
  - Además, en 11 rondas de 6 sesiones el usuario respondió con texto libre («resolvamos / resolvamoslo de una vez») en lugar de una opción, cuando una pregunta de «¿abro tickets para estos hallazgos?» iba junto a otra de «¿qué sigue?».
- **AC3.** De 25 entregas `automatic`, la pregunta de revisión se hizo en todas. En 7, el agente recomendó «Sin revisión dedicada», el usuario la eligió y se integró sin revisión.
- **H3, confirmado.** 3 rondas descartaron `automatic` por pantallas nuevas: Claude `94440d0a`, Claude `c6b1930e` y Codex `01a16bf8`.
- **H12, confirmado.** 19 de 469 mensajes de Claude que terminan en la línea de pausa también abren con «Pausa» (9 sesiones).
- **H4, sin casos.** No se encontró la pregunta de la primera vista bajo `automatic` ni desatendida.

## Lugares a revisar por cada regla cambiada

Antes del PR, por cada comportamiento cambiado, una búsqueda en `content/guidance/global.md`, `content/skills/**`, `content/agents/**`, `site/src/content/docs/**`, `_support/docs/architecture/**` y `tests/pilot/**`, con la lista de lugares en el PR (`AGENTS.md`, Guidance structure).
