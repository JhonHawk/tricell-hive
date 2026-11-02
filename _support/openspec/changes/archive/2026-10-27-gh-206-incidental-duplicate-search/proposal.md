# Búsqueda de duplicados en tracker y memoria, y fallos esquivados como hallazgos

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #231](https://github.com/JhonHawk/tricell-hive/pull/231) (`7372c065`) |
| Tracker · GitHub Issues | • [#206 — Hallazgos incidentales sin buscar duplicados en tracker y memoria, o sin reportar](https://github.com/JhonHawk/tricell-hive/issues/206) |
| Git | automatic · PR a `development` y merge · registro en PR de cierre sin revisión que fusiona el agente · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · `/code-review` de Claude Code |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> «D4-A #206 aquí» (tarjeta, 2026-10-27), tras recomendar #206 como siguiente ticket de Hive.

> «D7-A H1+H2+H4 con flow-plan (Recomendada)» (tarjeta, 2026-10-27), sobre los hallazgos de la investigación de #206.

#206 no declara criterios de aceptación; su «Acción sugerida» (leída el 2026-10-27): «Revisar si el paso de búsqueda queda lejos del momento en que se propone el hallazgo (durante el trabajo y no en el cierre), y si un error esquivado con un rodeo se reconoce como hallazgo.»

## Objetivo

En cuatro sesiones del 2026-10-26, los agentes propusieron hallazgos incidentales sin buscar duplicados en la memoria persistente, o no reportaron un fallo que esquivaron. La guía global pide dos búsquedas antes de proponer (tracker y memoria), pero la referencia que el agente lee al cerrar, `incidental-findings.md`, repite solo la del tracker, y el modelo sigue esa versión corta. Además, esa referencia se lee solo «at a work item's close», así que los hallazgos propuestos durante el trabajo no pasan por ella, y ninguna regla dice que un fallo de un comando o herramienta que se rodea es un hallazgo.

El cambio hace que la referencia nombre las dos búsquedas, amplía su condición de lectura al momento de proponer, y añade a la regla global que un fallo esquivado es un hallazgo incidental que se reporta junto con el rodeo.

## Alcance y aceptación

Incluido: `content/skills/flow-close/references/incidental-findings.md` (condición de lectura y viñeta *Duplicate check*), `content/guidance/global.md:34` (regla de hallazgos incidentales), las páginas Close y Build de `site/` y `CHANGELOG.md`.

Excluido, con su razón:
- H3 (tracker de fleetsystem sin sección `## Hive`): un solo caso, y su causa principal es la falta de esa sección, que corresponde a la sesión de fleetsystem con `harness-audit`. Queda en vigilancia (conteo 1).
- H5 (Pi no reportó un P3 en carvento): `review-and-gates.md:16` ya pide reportar los P3; un solo caso. Queda en vigilancia (conteo 1).
- El paso 5 de `flow-close/SKILL.md` y `verification.md:51`: ya apuntan a la referencia o a la regla global sin repetir sus pasos.

Admisión (AGENTS.md, Measurement): H1 es una corrección observada en dos sesiones independientes (Claude cam `c00e1e7f`, Claude fleetsystem `e2b47f4f`). H2 es un cambio de puntero que no añade regla. H4 es un hueco: ninguna regla cubre un fallo esquivado; lo adopta la decisión del usuario D7-A.

Restricciones:
- `go test ./...` sigue pasando, incluido `tests/content/skill_size_test.go` (referencias ≤ 16 000 bytes).
- Se informa el cambio de tamaño de `content/guidance/global.md` en el PR (41 174 bytes en la base).
- Ninguna otra parte de la regla de hallazgos incidentales cambia: severidad, agrupación, respuestas y recurrencia quedan igual.

Búsqueda en specs (2026-10-27): se buscaron `incidental`, `duplicate`, `persistent memory` y `workaround` en `_support/openspec/specs/`. Solo coincidieron requisitos ajenos (deduplicación de enlaces del visor y archivos de sección duplicados en `versioned-installation`). Ninguna capacidad describe la guía distribuida, así que el cambio no lleva delta.

- AC1. La viñeta *Duplicate check* de `incidental-findings.md` nombra las dos búsquedas de la regla global, en el tracker y en la memoria persistente. *Falso en la base cuando* `rg -n 'Duplicate check' content/skills/flow-close/references/incidental-findings.md` muestra una línea que menciona solo el tracker.
- AC2. La condición de lectura de `incidental-findings.md` incluye proponer un hallazgo durante el trabajo, no solo al cerrar. *Falso en la base cuando* la sección «When to read» dice solo «at a work item's close».
- AC3. La regla global de hallazgos incidentales dice que un fallo de un comando o herramienta que no bloquea el trabajo y se esquiva con un rodeo es un hallazgo incidental que se reporta con el fallo y el rodeo; solo se propone como defecto cuando está en el proyecto actual (no si es del host), y el de otro proyecto se reporta para la sesión de ese proyecto. *Falso en la base cuando* `sed -n 34p content/guidance/global.md | rg -c -i 'work around'` no devuelve nada. (La única aparición de «work around» en la base, en la línea 51, habla de claves privadas.)
- AC4. La página Close del sitio dice que el agente busca duplicados en el tracker y en su memoria antes de proponer un hallazgo, y la página Build dice que un fallo esquivado se reporta como hallazgo. *Falso en la base cuando* `rg -n -i 'duplicate|memory' site/src/content/docs/flows/close.md` no menciona la búsqueda y `rg -n -i 'work(s|ed)? around' site/src/content/docs/flows/build.md` no devuelve nada.

## Temas abiertos

- #206: este cambio; se cierra al integrarse.
- H3 y H5: vigilancia en la memoria del mantenedor, con conteo 1 cada uno.

## Entrega

Respuestas del 2026-10-27, en esta sesión:
- **D8-A `automatic`** en `tricell-hive`, base `development`: rama de trabajo, commits, revisión dedicada sobre el candidato local, push, PR a `development` y merge cuando pasen las comprobaciones. El registro se versiona al cerrar, en un PR pequeño sin revisión que fusiona el agente. Sin parada humana. No hay promoción a `master` en este alcance.
- **D9-A**: el `/code-review` de Claude Code sobre el candidato local, antes del push.
- El ticket #206 pasa a en curso al empezar y se cierra al integrarse, según la regla global de tickets.
- Tras integrar, `hive update` desde este checkout, autorizado por el marcador `maintainer.local` de AGENTS.md.
