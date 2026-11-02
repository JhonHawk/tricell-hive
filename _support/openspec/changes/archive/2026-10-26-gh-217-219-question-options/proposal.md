# Opciones completas en la pregunta de entrega y en la pregunta de ruta

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #220](https://github.com/JhonHawk/tricell-hive/pull/220) (`b5a4358e`) |
| Tracker · GitHub Issues | • [#217 — Pregunta de entrega sin la ruta de cierre del registro](https://github.com/JhonHawk/tricell-hive/issues/217)<br>• [#219 — Pregunta de ruta sin la opción de arreglar hallazgos que ya cumplen la admisión](https://github.com/JhonHawk/tricell-hive/issues/219) |
| Git | automatic · PR a `development` y merge · registro en PR de cierre que integra el agente · sin parada humana |
| Verificación | pruebas de contenido · suite completa · `/code-review` |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> Analiza las ultimas sesiones despues de hoy a las 4 am en los harness y valida el detalle de preguntas, dos cosas curiosas
> - Muchas preugntas en prosa (me gusta, pero me asusta)
> - Las preguntas de git estan saliendo extrañas, casi hasta el momento donde se va a ejecutar y no durante la planificación o siento que no se cumplen

> ok, ahi duda, que no deberiamos ir a un plan y luego build? porque cierras como si solo la investigación bastara? es un tema de reglas?

> creeo que si es un huevo, porque ya ha pasado varias veces (no recuerdo las sesiones)

Respuestas en la tarjeta: R4-A (abrir el issue del patrón y planear #217 y #219), D1-B (#219 solo en `flow-research`, sin tocar AGENTS.md; con una frase en el sitio), D2-A (automatic) y D3-A (`/code-review`). Tras la revisión del plan: E1, «evalua la mejor opción para agregar la clausa y que no sea un hueco de skill quien lo provoque, si no hay nadamas, si que sea E1-A», y E2-A (corregir en T1 la contradicción previa de los repositorios de soporte).

Criterios de aceptación de los tickets, leídos el 2026-10-26:

- [#217](https://github.com/JhonHawk/tricell-hive/issues/217):
  > - En una sesión que planea trabajo en un repositorio de código cuyo registro vive en ese mismo repositorio (o en un repositorio de specs sin `Delivery: direct-base`), cada opción de la pregunta de entrega dice cómo se versiona el registro al cerrar.
  > - Esa sesión no hace otra pregunta sobre cómo versionar el registro, ni al arrancar el build ni al cerrar.
- [#219](https://github.com/JhonHawk/tricell-hive/issues/219):
  > - En una investigación cuyos hallazgos dejan trabajo ejecutable, cada pregunta de ruta, incluida la que se repite cuando cambian los hallazgos, ofrece al menos una alternativa que los resuelve y nombra su flujo.
  > - En tricell-hive, cuando un hallazgo cumple la regla de admisión, la pregunta de ruta ofrece arreglarlo además de abrir el issue.

Los criterios de los tickets describen el comportamiento de sesiones reales. Este cambio no puede medirlos, porque los pilotos están en pausa (AGENTS.md, Measurement); los AC de abajo verifican el texto que debe producir ese comportamiento. Por D1-B, el segundo criterio de #219 queda cubierto por la regla general de `flow-research`, sin ancla propia en AGENTS.md.

## Objetivo

Las dos reglas ya existen, pero los modelos las incumplieron en sesiones reales. En #217, la exigencia de incluir la ruta de cierre del registro en cada opción de entrega está seis párrafos después del texto que dice qué lleva cada opción de la tarjeta, y Opus 5.5 la omitió en dos de tres planes. Eso provocó una pregunta git extra al arrancar el build o después del merge. En #219, la pregunta de ruta repetida tras nuevos hallazgos dejó solo opciones de «abrir issue y parar», y el agente cerró como si la investigación bastara.

El cambio pone cada exigencia en la frase que el agente sigue al redactar las opciones. En `flow-research` agrega además una invariante explícita: abrir el ticket y parar nunca sustituye a las alternativas que resuelven el trabajo.

## Alcance y aceptación

Incluye:

- `content/skills/flow-plan/references/delivery-decisions.md`: el párrafo de la tarjeta y el de texto nombran la ruta de cierre del registro.
- `content/skills/flow-research/SKILL.md`: la invariante del ticket en la pregunta de ruta, incluida la pregunta repetida.
- `site/src/content/docs/flows/research.md`: una frase que describe ese comportamiento.
- `content/guidance/global.md:24`: la invariante mínima en la copia de la pregunta de ruta (E1-A). En la sesión `8100a51b` no se leyó `flow-research` antes de la pregunta fallida, así que esa copia era la única regla presente.
- `delivery-decisions.md:16`, variante de texto: la publicación de los repositorios de soporte pasa a ir dentro de cada opción, como ya exige la línea 18 (E2-A; defecto previo que se reporta aparte).
- `CHANGELOG.md`: una entrada.

Excluye:

- AGENTS.md (decisión D1-B).
- #205 y #207 (otros fallos de preguntas, con sus propios issues).

Búsqueda de specs: se buscaron en `_support/openspec/specs/` los términos `route`, `delivery`, `closing`, `question`, `admission` y `watch item`. Solo aparecieron requisitos de `context-delivery` (entrega de contexto y fuentes de los paquetes), `walkthrough-viewer` (enlaces de entrega del visor) y `versioned-installation` (`hive doctor` y la sección `## Hive`). Ninguno describe el contenido de la pregunta de entrega ni el de la pregunta de ruta, ni los prohíbe. Sin delta de specs.

- AC1. El párrafo de `delivery-decisions.md` que dice qué lleva cada opción de la tarjeta, y su variante en texto, nombran la ruta de cierre del registro del cambio. *Falso en la base cuando* la línea 16 enumera solo «the full chain inside each option and the support repositories named in each» y, en texto, «the support repositories' publication, the ticket moves, and an unavailable route».
- AC2. `flow-research/SKILL.md` dice que abrir el ticket y parar nunca sustituye a las alternativas que resuelven el trabajo, en cada pregunta de ruta, incluida la que se repite cuando cambian los hallazgos. *Falso en la base cuando* `rg -n "never replaces|asked again" content/skills/flow-research/SKILL.md` no encuentra esa regla (la línea 46 solo dice «ask it again only when the findings change»).
- AC3. `site/src/content/docs/flows/research.md` dice que la pregunta de ruta, también cuando se repite, siempre ofrece al menos una forma de actuar sobre los hallazgos además de abrir el ticket y parar. *Falso en la base cuando* la línea 47 no menciona la pregunta repetida.

- AC4. `content/guidance/global.md:24` dice que cada pregunta de ruta, incluida la que se repite, conserva al menos una alternativa que resuelve el trabajo. *Falso en la base cuando* la línea 24 no menciona la pregunta repetida.

Restricciones: `go test ./tests/content/...` y la suite completa pasan; `SKILL.md` y las referencias siguen en 16 000 bytes o menos; el PR declara cuánto crece `global.md`.

## Contexto verificado (base `26012ce0`, 2026-10-26)

- `delivery-decisions.md:16` dice qué lleva cada opción de la tarjeta, y `:22` exige la ruta de cierre dentro de cada opción cuando el registro vive en un repositorio sin `Delivery: direct-base`. `flow-close/SKILL.md:78` aplica esa ruta al cerrar y, si no se registró, la pregunta antes de publicar: es el respaldo que produjo la pregunta D6 en #200.
- `flow-research/SKILL.md:46` define las opciones de la pregunta de ruta y dice «Ask it once per round; ask it again only when the findings change». `content/guidance/global.md:24` resume esa regla.
- Evidencia de las sesiones: en los cuerpos de #217 y #219.

## Diseño

- **#217:** se agrega «the change record's closing route that the change-folder paragraph below sets» a la enumeración del párrafo de la tarjeta. En la variante de texto, la ruta de cierre va dentro de lo que distingue a cada opción («such as where it stops, whether it publishes, or how it versions the change record»), no en la línea compartida. La línea 22 la exige dentro de cada opción, y bajo `hold` el registro queda sin versionar (corrección de la revisión del plan). El párrafo de la línea 22 sigue siendo la casa canónica de la regla; la línea 16 solo la nombra en el punto de uso.
- **#219:** después de «the recommended one first with its reason» se agrega una frase: toda pregunta de ruta, incluida la que se repite cuando cambian los hallazgos, ofrece al menos una alternativa que resuelve el trabajo, y abrir el ticket y parar nunca la sustituye. `global.md:24` lo dice con la misma forma (revisión 1, por `/code-review`).
- **#217, revisión 1:** `delivery-decisions.md:5` hace que una respuesta reutilizada sin ruta de cierre pregunte solo esa ruta antes del primer edit. La variante de texto pide la ruta solo cuando aplica el párrafo del registro. `site/.../build.md:48` lo describe.

## Entrega

- Repositorio `tricell-hive`, base `development`, rama propuesta `fix/gh-217-219-question-options` en un worktree nativo (`.claude/worktrees/gh-217-219`).
- D2-A automatic (2026-10-26): commits, `/code-review` (D3-A) sobre el candidato local, push, PR a `development` y merge con la suite local en verde. Sin parada humana.
- Registro: PR de cierre pequeño, sin revisión dedicada, que integra el agente.
- Después del merge: `hive update` local, autorizado por el marcador de mantenedor (`_support/workspace/maintainer.local`).
- Sin CI en el repositorio; la suite local es el gate.
