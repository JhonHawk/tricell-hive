# Una oferta al usuario es una pregunta

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por [PR #64](https://github.com/JhonHawk/tricell-hive/pull/64) (`4dc73c28`), 2026-10-17 |
| Tracker · GitHub Issues | Sin issue: nace de la prueba en Carvento del 2026-10-17 |
| Git | rama de prueba activa `feat/delegation-threshold-global` · commit por tarea y despliegue local (C1-A) · sin push ni merge |
| Verificación | comprobaciones `rg` por criterio · `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` |
| Siguiente paso | Ninguno: cambio cerrado y archivado |

## Petición

> R5-A Texto: ofertas = preguntas (Recommended)

(2026-10-17.) La opción decía: "Con flow-plan, en la rama de prueba: toda respuesta que termina pidiéndote algo, incluidas las ofertas, o se hace (si ya está autorizada y es solo lectura) o se pregunta con la tarjeta. Va en el aviso de global.md y en questions.md. Es barato y ataca la causa que comparten los hosts."

Contexto previo del usuario, el mismo día: "Abri una sesión nueva en carvento con claude" y "jeje, no habia nada".

## Objetivo

La guía nueva, desplegada en `aa17fbe1`, fija el orden de una pregunta: primero la línea de espera, después la tarjeta. Pero solo se activa cuando el agente reconoce que está formulando "una decisión". En la primera sesión de prueba (Carvento `1a3ab3ea`, 2026-10-17, claude-opus-5-5, con la guía nueva cargada), el agente cerró con "¿La reviso para que decidas?". No leyó `questions.md`, no puso línea de espera y no usó tarjeta. Además pedía permiso para algo de solo lectura. Lo mismo pasa en otros hosts con las ofertas del tipo "¿quieres que…?":
- OpenCode dened `ses_ebcb6cbd7ffe` (2026-10-16): dos ofertas en texto.
- fleetsystem `02891730` (2026-10-17, 06:42Z): "si abro desde aquí los tickets".

El cambio trata como pregunta toda respuesta que termina pidiéndole algo al usuario, incluidas las ofertas. Si la acción ya está autorizada y solo lee o reporta, el agente la hace en vez de ofrecerla. Si no, pregunta con el procedimiento de `questions.md` y la tarjeta. Siguiendo la norma adoptada hoy (`a2437302`), la regla completa va en `questions.md` y `global.md` solo amplía su aviso.

## Alcance y aceptación

Incluye:
- `content/skills/flow-plan/references/questions.md`: "When to read" y "Decide whether a question is needed".
- El aviso de lectura en `content/guidance/global.md:20`.
- La oferta de una página `flow-report` en `content/skills/flow-close/references/reporting.md:41`.

Queda fuera:
- Un hook del host para forzar la tarjeta (R5-B).
- Las ofertas dentro de un informe que no esperan respuesta, como los recordatorios de `reporting.md`.

Restricciones:
- La suite local pasa.
- El texto se escribe para los modelos del perfil `execution` y sigue la norma de punteros de AGENTS.md.
- Cada commit nombra las sesiones de origen y la decisión R5-A.
- El pull request declara cuánto cambia `global.md`.

- AC1. `questions.md` dice que un mensaje que termina con una oferta o petición que espera la respuesta del usuario, en cualquier forma, tiene tres salidas:
  - si la acción ya está autorizada, incluidas las comprobaciones que la guía global manda no preguntar, se hace;
  - si es una decisión del usuario, o una pregunta que otra regla exige (la de ruta, la de cierre o la de tickets), se pregunta con este procedimiento y nunca como una frase de cierre sola;
  - si no es ninguna de las dos, se omite.

  "When to read" incluye ese caso, y la última oración de la línea 11 se une al bloque nuevo. *Falso en la base cuando* `rg -c "waits on the user's answer" content/skills/flow-plan/references/questions.md` no encuentra coincidencias.
- AC2. El aviso de `global.md:20` manda leer `questions.md` también antes de terminar un mensaje con una oferta o petición que espera la respuesta del usuario, sin un ejemplo único. *Falso en la base cuando* `rg -c "waits on the user's answer" content/guidance/global.md` no encuentra coincidencias.
- AC3. `reporting.md:41` ofrece la página `flow-report` como una opción de la pregunta con que termina el mensaje, como la de ruta o la de cierre, o, si no hay ninguna, la pregunta siguiendo `questions.md`. *Falso en la base cuando* `rg -c 'as an option of the question the message ends with' content/skills/flow-close/references/reporting.md` no encuentra coincidencias.

## Entrega

- **Modo:** C1-A, elegido en esta sesión para este repositorio.
- **Revisión de código:** ninguna dedicada, porque son cambios solo de texto.
- **Registro:** esta carpeta queda sin versionar hasta el merge.
