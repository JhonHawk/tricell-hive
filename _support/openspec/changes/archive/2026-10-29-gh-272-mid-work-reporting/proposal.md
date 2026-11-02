# Dejar de exigir reportes a mitad del trabajo y simplificar el reporte de cierre

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el PR #273 (merge 9f25ff36) |
| Tracker · GitHub Issues | • [#272 — La guía obliga a reportar a mitad del trabajo y los reportes de entrega salen largos y con IDs sin glosa](https://github.com/JhonHawk/tricell-hive/issues/272) |
| Git | `automatic` · PR a `development` y merge al pasar las pruebas · sin parada humana |
| Verificación | `go test ./...` · búsquedas de frases retiradas · `/code-review` |
| Siguiente paso | Ninguno; observar en sesiones reales los casos E4–E6 y los "cómo vamos?" en esperas |

## Pedido

> La premisa es la siguiente:
>
> - En la sesión anterior resolvimos temas sobre el punto de reporteria, ahora pienso que realmente no es necesario obligar a reportar, talvez debamos quitar esas reglas, eso lo haria más ligero, lo que si debemos revisar es el formato cuando si se hace en las fases de entrega, validemos la sesiones y demas al respecto

> de hecho con E2 no tengo problema, que salgan naturalmente los terminos o verbos que usamos en nuestro glosario, ahora pensandolo bien, me parece bien, distingue el concepto de metodología de la prosa

> no, al contrario, quitemos las reglas que obligan al lenguaje sobre este tema, que sea natural, si acabo dejemos la regla de que se contexte en lenguaje de la sesión aaaah y las reglas que dictan sobre el codigo en ingles y convenciones, sobre desarrollo, esas no se tocan

Ticket [#272](https://github.com/JhonHawk/tricell-hive/issues/272), leído el 2026-10-29. Sus criterios de aceptación, textuales:

> - Ninguna regla distribuida obliga a enviar un mensaje de progreso, de entrada a una etapa, de tramo de espera o de ronda a mitad del trabajo.
> - Siguen escritas la regla que manda el texto junto con la siguiente llamada y la lista de casos en que puede terminar el turno.
> - Siguen escritos los avisos obligatorios de bloqueo, secreto, seguridad, preguntas y autorizaciones reutilizadas.
> - La lista del reporte de entrega aparece en un solo lugar; los demás la señalan con un puntero.
> - El reporte de cierre prescribe menos campos obligatorios y prohíbe citar IDs internos sin glosa.
> - La regla de vocabulario distingue los conceptos de la metodología (permitidos con su nombre) de la prosa (en el idioma de la sesión).
> - Las páginas de `site/` y `_support/docs/architecture/` que describen estos comportamientos coinciden con la guía.
> - `go test ./...` pasa.

El criterio de vocabulario del ticket quedó superado por D4 y D8: no hay distinción ni glosario; se quitan las reglas de lenguaje de los reportes (ver AC5).

## Objetivo

La guía obliga al agente a escribir a mitad del trabajo: anunciar el reparto entre subagentes, una frase por cada tarea terminada, una línea al entrar a build y a verify, un bloque "en curso" al esperar, una línea por tramo de espera y por ronda iterativa, y la línea de contexto en cada mensaje de progreso. En unas 190 sesiones reales del 22 al 29 de octubre, en seis CLIs, esos mensajes fueron muy frecuentes (solo en Claude, unas 1,300 notas de una línea y unos 1,400 reportes intermedios en 55 sesiones). No recibieron quejas y ninguno causó una parada.

El cambio quita esas obligaciones. Conserva la regla que de verdad evita las paradas: el texto va junto con la siguiente llamada y el turno solo termina en los casos de la lista. También deja el reporte de cierre con solo lo que cambia lo que el usuario hará después, en un único lugar, y quita las reglas que imponen un registro de lenguaje a los reportes, dejando solo el idioma de la sesión.

## Alcance y aceptación

**Incluido:**

- Quitar las obligaciones de reportar a mitad del trabajo en `content/`.
- Mover al reporte final la información que el usuario necesita.
- Dejar la lista del reporte de entrega en un solo lugar.
- Dar forma al reporte de cierre según D5-A.
- Quitar las reglas de lenguaje de los reportes según D8-A.
- Ajustar la línea "tu respuesta resuelve Dx" según D2-A.
- Alinear las páginas de `site/` y `_support/docs/architecture/`.

**Excluido:**

- E4 (una afirmación de cierre sin verificar, Claude `bc3842f3`), E5 (una revisión no hecha puesta como recordatorio, cam `4afe6d47`) y E6 (tres mensajes de cierre seguidos, Grok `01a1855f`). Se vieron en una sola sesión cada uno y quedan en observación: un watch item de Engram en `decisions/mid-work-reporting-removal`.
- Las reglas de código en inglés, i18n, commits y convenciones de desarrollo (`global.md` Language, `hive-core/references/language.md`), por decisión del usuario.
- La glosa de IDs y el formato de enlaces, que se mantienen.
- El vocabulario del proyecto en los tickets (`ticket-body.md:11`) y el del usuario en la revisión de UI (`ui-review-criteria.md:67`), porque tratan del producto y no de los reportes.

**Restricciones que deben seguir siendo verdad:**

- La regla "manda el texto junto con la siguiente llamada", la lista de casos en que el turno puede terminar (`work-item-boundary.md`), el paso 6 del turno de notificación y la regla de `flow-verify:65` de no mandar el reporte en un mensaje propio siguen escritos.
- Siguen obligatorios los avisos de bloqueo, de integración caída, de secreto expuesto y de hallazgo de seguridad o crítico; las preguntas; la línea de autorizaciones reutilizadas (`delivery-decisions.md:11`); la oferta de revisión de seguridad (`security-boundaries.md:16`); la línea de clasificación de un tema nuevo (`global.md:18`); y la línea de qué respuesta en vigor cubre una pregunta (`questions.md:19`).
- Cada `SKILL.md` y cada referencia se quedan en 16,000 bytes o menos, y `content/guidance/global.md` no crece.
- `go test ./...` pasa.

**Búsqueda en los requisitos vigentes** (`_support/openspec/specs/`, 2026-10-29). Términos buscados: voice, vocabular, progress, report, context line, mid-work, turn.

- `versioned-installation` "Optional voice layer" dice que la voz no cambia las reglas de comunicación. Quitar la línea de vocabulario del preámbulo no lo altera, porque esa regla deja de existir y la voz sigue sin cambiar ninguna.
- `context-delivery` "Declared source footprint without token claims" cuenta la referencia de reporting como cota superior y no fija su contenido.
- Ningún requisito enumera los mensajes de progreso ni el formato del reporte.
- El cambio no necesita deltas.

**Criterios:**

- AC1. Ninguna regla de `content/` exige un mensaje a mitad del trabajo para: el progreso de una tarea, la entrada a build o a verify, el reparto entre subagentes, un lanzamiento que bloquea, un tramo de espera, una ronda iterativa, el nombre de la corrida de un deploy, ni una verificación más grande que el cambio. *Falso en la base cuando* `rg` encuentra "one progress sentence" (`work-item-boundary.md:11`), "announce the split" (`flow-build/SKILL.md:42`), "At entry, say in one line" (`flow-verify/SKILL.md:18`) o "Report each round in one line" (`flow-research/SKILL.md:48`).
- AC2. Cuando el turno termina para esperar trabajo que sigue corriendo, la guía pide una oración que nombre lo que sigue corriendo y lo que decide su resultado (D3-A), y ya no el bloque "en curso" ni una línea por tramo. *Falso en la base cuando* `reporting.md:66` pide "a labeled in-progress item" y `delegation.md:67` "one progress sentence" por tramo.
- AC3. La lista de lo que cambió, lo que se verificó y a qué nivel, y lo que quedó sin cumplir o sin verificar vive solo en `reporting.md`. `flow-verify/SKILL.md` e `iterate-and-integrate.md` apuntan a ella. *Falso en la base cuando* la lista aparece en `flow-verify/SKILL.md:65` y en `iterate-and-integrate.md:15`.
- AC4. El reporte de cierre siempre lleva: qué se completó, en qué punto, y qué excluye; el PR; y el destino de los temas abiertos. Lleva lo no verificado o no cumplido, las decisiones tomadas sin preguntar, la limpieza que dejó algo vivo o falló, y los incidentes, solo cuando hay alguno. No lista los chequeos que pasaron. *Falso en la base cuando* `flow-close/SKILL.md` (paso 7) pide siempre "a cleanup line stating what was removed and what was kept".
- AC5. La única regla de lenguaje para el texto de los reportes es escribirlo en el idioma de la sesión. *Falso en la base cuando* `reporting.md:9` contiene "everyday words" y "are not the user's vocabulary", y `content/voices/preamble.md:9` contiene "plain everyday words".
- AC6. La línea "tu respuesta resuelve Dx" solo se exige cuando la respuesta no elige una opción ofrecida, invalida algo o regresa el trabajo a otra etapa (D2-A). *Falso en la base cuando* `questions.md:88` la pide para toda respuesta.
- AC7. Lo que hoy se avisa a mitad del trabajo y el usuario necesita conocer va en el reporte final, entre las decisiones tomadas sin preguntar: un defecto que bloqueaba y se arregló sin preguntar, y el ticket elegido cuando el usuario lo dejó al agente. La ruta del primer render, cuando no hay parada de validación, va en el reporte final. *Falso en la base cuando* `global.md:29` pide "report it in one line" y `flow-walkthrough/SKILL.md:24` da la ruta "in the progress note".
- AC8. Las páginas de `site/` y `_support/docs/architecture/instruction-resources.md` describen el comportamiento nuevo. *Falso en la base cuando* `site/src/content/docs/flows/build.md:43` describe la respuesta de una línea a cada aviso, `verify.md:22` la línea de entrada y `plan.md:75` "every progress or decision message".

## Entrega

- **Repositorio:** `tricell-hive`. Rama de trabajo propuesta `fix/gh-272-mid-work-reporting`, desde `development` fa3e3948.
- **Modo:** `automatic` (D6-A, respuesta del usuario del 2026-10-29). Commits, `/code-review` de Claude Code sobre el candidato local (D7-A), push, PR a `development` y merge cuando pasen las pruebas.
- **Cierre del change record:** un PR pequeño sin revisión dedicada, con merge del agente.
- **Al integrar:** el ticket #272 se cierra y se corre `hive update` (marcador del maintainer, `AGENTS.md`).
- **Sin efectos adicionales previstos:** no hay despliegue, chequeo costoso ni escritura fuera del repositorio.
- **Requisitos de `AGENTS.md` para el PR:** listar los lugares revisados y el cambio de tamaño de `content/guidance/global.md`. El commit nombra las sesiones que lo respaldan.
