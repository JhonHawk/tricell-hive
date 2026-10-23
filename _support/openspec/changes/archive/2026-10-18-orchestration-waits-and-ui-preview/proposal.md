# Esperas del orquestador, mensajes con contexto y preview de UI en HTML (C1–C6 del research de la sesión cam e2c8d39f)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · fusionado en `development` por [#73](https://github.com/JhonHawk/tricell-hive/pull/73) (`bb4fd503`) el 2026-10-18, a decisión del usuario antes de las pruebas en sesiones reales previstas; desplegado en local |
| Tracker · GitHub Issues | Sin issue propio: C1–C6 se adoptaron por decisión del usuario (AD1); el único cambio no adoptado, C7, quedó en • [#71 — Etiquetas del host en inglés con la regla de idioma ya cargada](https://github.com/JhonHawk/tricell-hive/issues/71) |
| Git | E1-A: rama de prueba `feat/gh-67-flow-walkthrough-skill` · un commit por tarea verificada · sin push ni merge · redespliegue local solo a petición |
| Verificación | `rg` por criterio · `go test ./...` · `python3 -m unittest discover -s tests/skills -p '*_test.py'` · `self_check.py` sobre el esqueleto · `pnpm -C site build` · tamaños · auditoría de coherencia del conjunto (`hive-review-harness`) |
| Siguiente paso | Ninguno: registro archivado |

## Petición

> Analiza la sesión de cam actual, aunque no esta usando la versión desplegada, debemos responder:
>
> - Qué etapas generaron demasiado tiempo de iteración?
> - qué mensajes de espera para el usuario no generaron la pausa necesaria (o se generaron mientras subagentes seguian trabajando) generando nuevos turnos y practicamente rezagando el mensaje de espera.
> - La estructura de mensajes, resumentes, entrega, es adecuado? Deja claro a que se refiere cada etapa? recordemos que es probable que el usuario no solo este atendiendo este tema y este en varios al mismo tiempo
> - Al presentar el hmtl de preview, deberia abrirse en automatico para evita que el usuario lo tenga que pedir?

> Sobre el preview, yo creo que solo el html debe abrirse, ahi estan las capturas

Respuestas del 2026-10-18, tras el research: AD1 = C1, C3, C4, C2, C5 y C6 adoptados como reglas por decisión explícita del usuario (sustituye el umbral de dos sesiones de `AGENTS.md`, sección Measurement); RQ1-A = planear con `flow-plan`; TK1-A = un issue por cambio no adoptado (solo C7 → #71).

Cambios adoptados, tal como se presentaron:

- **C1** — la primera vista se entrega en una página HTML con las capturas dentro y la página se abre sola antes de preguntar.
- **C2** — un turno de notificación que no cambia nada responde en una línea, sin re-listar procesos ni repetir la pausa.
- **C3** — un subagente que devuelve «esperando su propio trabajo» se trata como detenido: se comprueba su estado o se reanuda de inmediato.
- **C4** — una tarea con interfaz se compromete después de la primera vista y de la revisión de interfaz; la suite completa y la revisión de seguridad corren antes de pedirte validación.
- **C5** — el brief de un subagente lleva las skills de dominio que activan las reglas condicionales del proyecto.
- **C6** — una línea fija de contexto (etapa · ticket · avance) al inicio de cada mensaje de progreso o decisión.

## Objetivo

En la sesión de Claude Code `e2c8d39f` del proyecto cam (2026-10-18, opus-5-5, Hive desplegado `9c131d3b`), de 7 h 14 min el orquestador trabajó 31 min, esperó a subagentes 151 y al usuario 253. El tiempo perdido no vino de la iteración en sí, sino de cuatro patrones que Hive hoy no impide:

- una verificación que terminó su turno esperando su propio vigilante quedó parada 68 min mientras el orquestador repetía «sigue en curso»;
- una tarea de interfaz comprometida antes de su primera vista y de la revisión de interfaz se rehízo tres veces;
- la suite completa y la revisión de seguridad corrieron después de la validación humana y cambiaron código ya validado;
- la primera vista llegó como PNG sueltos que el usuario tuvo que pedir abrir, y la página del recorrido tampoco se abrió sola.

A eso se suman 34 mensajes de estado (11 repetición pura) y mensajes largos que no dicen etapa, ticket ni avance, para un usuario que atiende varias sesiones a la vez.

El cambio fija una regla con una sola casa para cada patrón (C1–C6), de modo que el orquestador detecte un hijo detenido en el acto, ordene las puertas de UI antes del commit y de la validación, entregue y abra la vista en HTML, calle cuando nada cambió y abra cada mensaje con su contexto.

Durante el plan, el usuario pidió además (2026-10-18):

> Recuerda que debemos evaluarlo en su conjunto, no como cambios individuales, para evitar contradicciones o reglas que van por el mismo camino pero condiferentes alcances

Por eso el diseño ([design.md](design.md)) agrupa C1–C6 en cuatro momentos (M1–M4), cada uno con una sola casa y un solo orden, y una auditoría de coherencia revisa el conjunto antes de cerrar.

## Alcance y aceptación

Incluye las reglas de los cuatro momentos de [design.md](design.md), sus punteros, la página del sitio `flows/build.md`, `CHANGELOG.md` y las notas de mantenedor `_support/docs/architecture/agent-delivery.md` e `instruction-resources.md`. `README.md` y `llms.txt` no cambian: no mencionan nada de esto.

Criterios. La base es `8c15f59a`; cada observación «falso en la base» se comprobó en ella el 2026-10-18.

- AC1 (M2, apertura). `content/skills/flow-close/references/reporting.md` tiene una sola condición que hace abrir cada mensaje de progreso o de decisión con una línea de contexto `<stage> · <ticket(s)> · <progress>`. Esa condición fija el orden: línea de contexto, luego la apertura que toque (en curso, completado, respuesta o bloque «Pausa»), el cuerpo, y la línea de pausa al final. Toda nota de progreso de una línea es la línea de contexto seguida del cambio. La condición dice que la línea de contexto es un encabezado y no la primera oración, y `content/voices/preamble.md` lo aclara («after any context line»). Exime las respuestas cortas directas y los retornos de subagentes. Hace empezar la pregunta de una tarjeta nativa con etapa y ticket. El ejemplo resuelto de `reporting.md` abre con la línea de contexto. *Falso en la base cuando* `rg -n -F "context line" content/skills/flow-close/references/reporting.md` no devuelve nada.
- AC2 (M1, turno de notificación). `reporting.md` tiene una sola condición para el turno que abre una notificación de subagente o de proceso en segundo plano, con tres salidas en este orden:
  1. el aviso dice que el hijo se detuvo esperando trabajo propio, y se aplica la comprobación de `delegation.md` (puntero);
  2. el resultado decide algo, trae un bloqueo nuevo o cambia lo que se pide, y se reporta con el ítem en curso cuando otro trabajo sigue corriendo. Una petición que cambia es una pregunta nueva con ID nuevo que reemplaza a la anterior. Con una pregunta abierta, un resultado que no la cambia toma la salida 3 (D2-A);
  3. si nada de eso, la línea de contexto con el cambio, en una línea, sin re-listar lo que corre ni repetir una pregunta abierta o su pausa.

  Esa condición absorbe la cláusula de notificaciones de la sección «Work-item boundary», y la condición del ítem en curso deja de aplicarse solo a la salida 3. `questions.md` («Pause and handle the answer») nombra en su punto de uso la excepción: la línea de la salida 3 no lleva pausa ni cuenta como segundo reporte, y una petición que cambia es una pregunta nueva que reemplaza a la anterior. *Falso en la base cuando* `rg -n -F "changes nothing" content/skills/flow-close/references/reporting.md` no devuelve nada.
- AC3 (M1+M2, invariante global). `content/guidance/global.md`, en Communication, enuncia la invariante completa: la línea de contexto en los mensajes de progreso o de decisión, la nota de una línea como esa línea con el cambio, la respuesta de una línea al turno de notificación sin cambios, y que no aplica al retorno de un subagente ni a una respuesta corta sin trabajo en curso. *Falso en la base cuando* `rg -n -F "context line" content/guidance/global.md` no devuelve nada.
- AC4 (M1, punteros). Ningún otro lugar exige el ítem en curso completo en un turno de notificación ni repetir lo que espera. La línea por tramo de espera de `delegation.md` usa la forma de la línea de contexto. *Falso en la base cuando* `rg -n -F "ends its turn with the in-progress item" content/skills/flow-research/references/delegation.md` y `rg -n -F "nothing is repeated for them" content/skills/flow-walkthrough/SKILL.md` devuelven una línea cada uno.
- AC5 (M1, hijo detenido). La sección «Check returned work» de `delegation.md` dice que un hijo que vuelve, o un aviso del host que dice que se detuvo, mientras espera trabajo propio (un vigilante, una espera o un comando en segundo plano) está detenido. Pide comprobar en el acto si siguen vivos él y ese trabajo, reanudar el mismo hijo o repetir la comprobación, y no reportarlo como en curso. La frase de vida de `flow-build/SKILL.md` apunta ahí. *Falso en la base cuando* `rg -n -F "work of its own" content/` no devuelve nada.
- AC6 (M4, brief). La sección «Deliver and resolve instructions» de `delegation.md` dice que el brief lleva las skills y los documentos del proyecto que sus reglas indican para ese tipo de trabajo, evaluando cada regla condicional contra lo que el host expone. `flow-build/SKILL.md` (brief de tarea) y `flow-walkthrough/references/ui-gate.md` (brief de revisores) apuntan ahí, y el segundo aclara que pasarlas no es pre-calificar hallazgos. *Falso en la base cuando* `rg -n -F "conditional rule" content/` no devuelve nada.
- AC7 (M3, página de la primera vista). La sección «First rendered view» de `flow-walkthrough/SKILL.md` escribe la primera vista en `walkthrough.html` con un llenado parcial del esqueleto, corre `self_check.py`, abre la página según AC8 y después pregunta. *Falso en la base cuando* la sección «First rendered view» no contiene `walkthrough.html`: `awk '/^## First rendered view/{f=1;next}/^## /{f=0}f' content/skills/flow-walkthrough/SKILL.md | rg -c walkthrough.html` no imprime nada.
- AC8 (M3, abrir la página). `flow-walkthrough/SKILL.md` tiene una regla «Open the page» con estas partes:
  - abre el archivo local en el visor predeterminado del sistema, solo el HTML, una vez por versión nueva de la página;
  - lo hace antes del mensaje que pide la decisión o la validación;
  - el mensaje da siempre la ruta absoluta y, cuando el host no pudo abrirla (remoto, sin interfaz, rechazo del sandbox o de un permiso), lo dice;
  - nunca publica.

  La frase de que el usuario abre la página desaparece. *Falso en la base cuando* `rg -n -F "Open the page" content/skills/flow-walkthrough/SKILL.md` no devuelve nada y `rg -n -F "the user opens the page" content/skills/flow-walkthrough/SKILL.md` devuelve la línea 44.
- AC9 (M3, commit de tareas con interfaz, D1-A). La regla de commit por tarea de `flow-build/SKILL.md` deja una tarea con efecto de UI verificada y sin commit hasta que la primera vista del cambio, cuando se mostró, fue aceptada y la revisión de interfaz del cambio pasó. Después hace un commit por tarea, en su orden. Bajo `automatic` espera solo la revisión; bajo `UI verification: none`, la comprobación renderizada del implementador. Una tarea sin UI que comparte archivos con una de esas, o depende de ella, también espera. Los arreglos de la suite completa o de la revisión de seguridad sobre una tarea ya comprometida, y las correcciones de la validación humana, van en un commit adicional `(T<n>)` por tarea afectada, sin amend, y si el trabajo se pausa antes, el reporte nombra las tareas verificadas sin commit. *Falso en la base cuando* `rg -n -F "first rendered view and the UI gate" content/skills/flow-build/SKILL.md` no devuelve nada.
- AC10 (M3, antes de pedir validación). Una sola casa en `flow-build/SKILL.md` dice que, antes de pedir al usuario que valide, corren dos cosas sobre el candidato: la suite completa de la aplicación afectada, en local, y `hive-review-security` cuando el cambio cruza una frontera, con la autorización que ya fija `verification.md`. La revisión de seguridad que nombre después el veredicto del revisor de código sigue en el momento de ese revisor, y un arreglo de seguridad que toca archivos de UI repite la revisión de interfaz sobre ese diff antes de pedir la validación. La revisión de código sigue donde la pone el modo de entrega. `verification.md` (fila «Local candidate» y párrafo de revisores), `delivery-decisions.md` (modo `interactive`) y `ui-gate.md` («When the gate runs») apuntan ahí con la frase «before asking the user», en vez de repetirlo. *Falso en la base cuando* `rg -n -F "before asking the user" content/` no devuelve nada.
- AC11 (documentación). `site/src/content/docs/flows/build.md` describe la línea de contexto, la respuesta de una línea a una notificación sin cambios, el hijo detenido, la página que se abre sola y el commit tras la revisión de interfaz. `close.md`, `research.md` y `plan.md` del sitio describen la línea de contexto donde hoy describen la pausa o la tarjeta. `CHANGELOG.md` tiene una entrada de este cambio. `agent-delivery.md` documenta el aviso de Claude Code sobre un hijo detenido con trabajo propio pendiente, como observado en la sesión `e2c8d39f`. *Falso en la base cuando* `rg -n -F "context line" site/src` no devuelve nada.

Restricciones:

- `flow-walkthrough/SKILL.md` queda por debajo de 1.800 palabras (tope de gh-67).
- `ui-gate.md` queda por debajo de 880 palabras; el tope de gh-67 era 850 y este cambio lo sube 30 para el puntero de M4.
- El cambio de tamaño de `global.md` (30.717 bytes y 4.699 palabras en la base) se informa.
- Pasan `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `self_check.py` sobre el esqueleto y `pnpm -C site build`.
- Las anclas `#report-a-running-or-completed-item` (citada por `delegation.md` y `unattended-delegation/SKILL.md`) y las demás citadas siguen existiendo.

Fuera de alcance:

- C7 (etiquetas del host en inglés): issue #71.
- El comportamiento del host de Claude Code que retiene notificaciones mientras una tarjeta espera: observado una vez, no documentado; no se legisla sobre él.
- El aislamiento de sesiones de navegador entre revisores de UI (punto a vigilar, conteo 1).

## Entrega

Respuestas del 2026-10-18:

- **E1-A (modo):** rama de prueba `feat/gh-67-flow-walkthrough-skill`, con base `8c15f59a`. Un commit por tarea verificada, sin push ni merge, y redespliegue local solo a petición. El merge a `development` va tras las pruebas del usuario. Es una instrucción más estrecha que `interactive`.
  - C1 modifica `flow-walkthrough`, que solo existe en esa rama.
  - El registro del cambio queda sin commit hasta `flow-close`.
  - Cada commit nombra la sesión que motivó la regla (Claude Code, `e2c8d39f`, 2026-10-18) y el fallo observado, como pide `AGENTS.md` (Measurement).
- **R1-A (revisión de código):** ninguna dedicada, y sin revisión de seguridad porque no se cruza ninguna frontera. La auditoría de coherencia del conjunto con `hive-review-harness` responde al pedido de evaluarlo en conjunto. No es una revisión de código del diff.
- **UI verification:** no aplica. El esqueleto HTML solo cambia un comentario y no hay efecto de UI renderizado.
