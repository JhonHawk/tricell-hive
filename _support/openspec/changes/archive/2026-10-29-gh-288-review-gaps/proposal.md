# Huecos de la segunda revisión del tope de rondas

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por PR #292 (merge 83008381, 2026-10-29) |
| Tracker · GitHub Issues | • [#288 — Revisión dedicada: cerrar los huecos que dejó la segunda revisión del tope de rondas](https://github.com/JhonHawk/tricell-hive/issues/288) |
| Git | automatic · PR a `development` y merge · sin parada humana |
| Verificación | tests de contenido · suite completa una vez · `/code-review` |
| Siguiente paso | Ninguno; archivado |

## Request

> Seguir con #288 (Recomendado)

> flow-plan (Recomendado)

Respuestas de la sesión (2026-10-29): D4-A (seguir con #288 en esta sesión), D5-A (ruta `flow-plan`), D6-A, D7-A y D8-A (ver Entrega y `design.md`).

Criterios de aceptación de #288, citados textualmente (leído el 2026-10-29):

> - Una corrección de validación hecha antes de la segunda revisión recibe una sola revisión.
> - Una edición posterior sin revisar queda registrada aunque no haya PR ni registro de cambio.
> - El reporte de revisión nombra el hallazgo y el arreglo, o la edición, que abre la segunda revisión.
> - La segunda revisión de seguridad tiene un tope explícito frente a las ediciones posteriores.
> - La regla de revisión dedicada vive en una referencia que deja margen bajo el límite de bytes.

## Objective

La segunda revisión de `/code-review` sobre el PR #286 dejó seis huecos sin arreglar en la regla de revisión dedicada (`content/skills/flow-verify/references/review-and-gates.md`), porque la regla acotada manda proponer los P2 en vez de corregirlos. Una corrección de validación se revisa dos veces; una edición sin revisar puede quedar sin registro cuando no hay PR ni registro de cambio, y la promoción la daría por revisada; el reporte puede cumplir «qué abre la segunda revisión» repitiendo la regla; la segunda revisión de seguridad no tiene tope frente a ediciones posteriores; y un revisor puede saltarse el refutador con solo declarar que verificó. Además, el archivo está a 9 bytes del límite de 16 000, así que ningún arreglo cabe sin dividirlo.

El cambio divide la referencia por responsabilidad y luego cierra los seis huecos con la redacción mínima, de modo que cada edición posterior y cada corrección tenga exactamente una revisión o un registro, y todo P0/P1 pase por el refutador.

## Scope and acceptance

Incluido: dividir `review-and-gates.md` (las secciones «Runtime and UI gates» y «Human UI handoff» pasan a `runtime-gates.md`), las reglas de revisión de código y de seguridad que quedan en él, la regla de promoción de `content/skills/flow-plan/references/delivery-decisions.md`, cada enlace y texto que repita o dependa de lo que cambia (incluidos `site/`, `_support/docs/architecture/instruction-resources.md` y `tests/content/reading_footprint_test.go`).

Excluido, con destino:
- Los límites de rondas en sí (una revisión completa y una segunda): decisiones del usuario D3-A y D4-A del cambio `review-loop-bounds`; no se tocan.
- La cadencia de la suite completa: en observación aparte (Engram `watch/full-suite-per-review-round`).
- R5 de #290 (un hijo por cada búsqueda de duplicados en un cierre): P3 reportado, sin destino de trabajo.

Búsqueda de requisitos: en `_support/openspec/specs/` (context-delivery, documentation-site, versioned-installation, walkthrough-viewer) se buscaron `re-review`, `dedicated review`, `refut`, `promotion`, `unreviewed`, `commit message`, `review-and-gates`, `16,000`, `footprint` y `byte limit`. Solo coincide «Declared source footprint without token claims» (context-delivery): exige que cada escenario nombre sus fuentes y las cuente una vez, sin listar archivos; añadir `runtime-gates.md` al escenario del ciclo lo cumple sin alterarlo. El cambio solo altera redacción de guía: no lleva delta.

Restricciones: cada referencia tocada queda en 16 000 bytes o menos (`tests/content/skill_size_test.go`); la regla nombra la capacidad y no un comando de host (AGENTS.md); `content/guidance/global.md` no cambia; la suite existente sigue pasando.

- AC1. `review-and-gates.md` deja al menos 3 000 bytes libres bajo el límite y la regla de runtime y UI vive en `runtime-gates.md`, con todos los enlaces resueltos. *Falso en la base cuando* `wc -c` da 15 991 bytes y no existe `runtime-gates.md`.
- AC2. Una corrección pedida al validar, hecha antes de que corra la segunda revisión, no abre ni se une a esa segunda revisión: recibe solo la revisión de correcciones de «Dedicated review timing». *Falso en la base cuando* la viñeta «Later edits» incluye toda edición que no sea arreglo de revisión.
- AC3. Una edición posterior sin revisar se anota en el mensaje de su commit además del cuerpo del PR o el registro de cambio, y la regla de promoción revisa también los commits que un mensaje de commit marca sin revisar; un commit sigue contando como revisado solo cuando el registro de cambio o el cuerpo de su PR lo registra así. *Falso en la base cuando* la viñeta «Later edits» y `delivery-decisions.md` solo nombran el cuerpo del PR y el registro de cambio.
- AC4. El reporte de revisión nombra el hallazgo P0/P1 y su arreglo, o la edición posterior, que abre la segunda revisión, o dice que ninguno la abre. *Falso en la base cuando* la viñeta «Severity» solo pide registrar «what starts a re-review».
- AC5. Después de la segunda revisión de seguridad, una edición posterior que no cruza un límite de seguridad no abre otra revisión de seguridad y se reporta sin revisar; una que cruza un límite recibe la revisión que `flow-security` fija para los límites que aparecen más tarde, que es la única excepción. *Falso en la base cuando* «Security review findings» solo dice «re-review at most once under the same limits».
- AC6. Todo P0/P1, y todo hallazgo que bloquearía el merge, pasa por `hive-refute-claim` antes de actuar, sea cual sea el mecanismo; ningún texto exime a un revisor por declarar que verificó. *Falso en la base cuando* `review-and-gates.md:25` exime a la revisión cuya salida declara que reprodujo o releyó cada hallazgo.

## Delivery

- **Modo:** `automatic`, reutilizado de esta sesión para `JhonHawk/tricell-hive` y `development` (respuesta a «todo el flujo para desplegar», 2026-10-29): rama de trabajo, commits por tarea, `/code-review` de Claude Code sobre el candidato local, push, PR a `development` y merge con la suite completa local en verde (el repo no tiene CI). Después, `hive update` del checkout del mantenedor.
- **Revisión:** `/code-review` de Claude Code (D1-A, reutilizada). Sus P0/P1 pasan por `hive-refute-claim`.
- **Registro de cambio (D8-A):** al cierre, tras el merge, un PR pequeño a `development` sin revisión dedicada que integro yo, con los deltas (ninguno) aplicados y la carpeta archivada.
- **Ticket:** #288 pasa a en progreso al empezar el build y se cierra con el merge.
- **Excluido:** promoción a `master`.

## Status

Completado: integrado por [PR #292](https://github.com/JhonHawk/tricell-hive/pull/292) (merge `83008381`) y desplegado con `hive update`. Sin deltas que aplicar.
