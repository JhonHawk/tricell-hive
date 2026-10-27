# Diseño

## Contexto verificado (base `20646413`, 2026-10-27; las líneas citadas no cambiaron desde `95077cd4`)

- `content/guidance/global.md:34`: la regla de hallazgos incidentales pide «check the project's tracker for duplicates» y, más adelante, «Before proposing, search persistent memory for the finding, skipping ones saved as suppressed». También dice «Pre-existing and intermittent test failures that do not block the work are incidental findings too». No menciona fallos esquivados.
- `content/skills/flow-close/references/incidental-findings.md:5`: se lee «before listing, asking about or recording the incidental findings that the global guidance says to propose at a work item's close».
- `incidental-findings.md:13`: la sección «narrows the incidental-findings rule ... and names each element it changes; every other element of that rule stands».
- `incidental-findings.md:19`: «**Duplicate check.** Search the tracker for duplicates once per proposal, so once per group, not once per member.» Solo nombra el tracker.
- `AGENTS.md` (Guidance structure): «never restate some of its steps there, because a model follows the shorter version and skips the rest».
- Puntos que dependen de este comportamiento y que el cambio deja como están, porque solo apuntan: `flow-close/SKILL.md:31` (paso 5), `flow-build/references/verification.md:51` y `:56`, `flow-research/SKILL.md:42`, `flow-research/references/ticket-body.md:18`, `_support/docs/architecture/instruction-resources.md:18` (condición de lectura, que se actualiza en T1). En `tests/pilot/` no hay detectores de este comportamiento (`rg -i 'incidental|duplicate'` solo encuentra pruebas ajenas).

## Enfoque

- **H1.** La viñeta *Duplicate check* remite a «the tracker and persistent-memory searches the global rule sets», sin copiar sus pasos. La búsqueda en el tracker se hace una vez por propuesta (una por grupo); la búsqueda en memoria, una por miembro, porque las respuestas se guardan por miembro (`incidental-findings.md:20`) y un miembro suprimido o rechazado dentro de un grupo nuevo debe encontrarse. No repite «skipping suppressed», que ya cubre la línea 9.
- **H2.** La condición de lectura de la referencia pasa a cubrir «proponer un hallazgo incidental durante el trabajo o al cerrar», y la línea 13 deja de limitar su estrechamiento al cierre: se aplica dondequiera que se proponga o liste un hallazgo. La regla global ya apunta a la referencia para proponer («as [incidental findings] sets»), así que no hace falta otro puntero. La descripción de la referencia en `instruction-resources.md:18` se alinea con esa condición.
- **H4.** Una frase en la regla global, junto a la de pruebas que fallan: un fallo de un comando o herramienta que no bloquea el trabajo y que se esquiva, por ejemplo reintentando con otra ruta o flag, es un hallazgo incidental; se reporta con el fallo y el rodeo. El que pertenece al host o a otro proyecto se reporta por la regla de otros proyectos (`global.md:30`). Hereda el alcance de la viñeta («authorized implementation or review»), sin ampliarlo; un fallo que bloquea el trabajo sigue la regla de defectos bloqueantes (`global.md:33`). Va en `global.md` porque debe actuar durante el trabajo, sin que se haya leído la referencia.
- **Sitio.** Close y Build describen el comportamiento y enlazan la regla, sin copiarla.

Alternativa descartada: mover la búsqueda en memoria solo a la referencia y quitarla de `global.md`. Dejaría sin la búsqueda a quien propone durante el trabajo sin leer la referencia, que es el caso de fleetsystem.
