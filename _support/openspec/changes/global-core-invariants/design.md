# Design

## Approaches

Solo hay una forma razonable: para cada bullet que se decidió mover, el núcleo conserva una frase con el invariante y una cita a la hoja; el texto movido se copia sin cambios a la hoja que se lee en ese momento, salvo la cláusula de procesos de la limpieza, que en `close-cleanup` dice «stopping the processes it or its children started, as the global guidance's «Preservation» sets» porque su regla completa se queda en el núcleo. No se reescribe ninguna regla. Mover también `global.md:14`, `:21`, `:30`, `:39` y `:46` se descartó (D29 y el supuesto).

## Moves

| Origen | Destino | Lo que queda en el núcleo |
| --- | --- | --- |
| `global.md:32` (enumeración de la limpieza) | `close/close-cleanup.md`, hoja nueva fuera del paquete de flow-close | «At a work item's close, the cleanup [close cleanup](…) lists is authorized; stop only the processes the work or its children started, by process ID, never by a name or pattern that can match another session's.» seguido, completos, de los «Never…» y de la frase de contenedores y evidencia única |
| `global.md:50` (definiciones de cambio mecánico y de cambio pequeño entendido, y qué se salta un cambio mecánico) | `plan/proportionality.md` | «Choose the smallest sufficient investigation, plan, implementation and verification.», «A simple question needs no formal workflow.», el enrutado a `hive-mode` y la regla de que un cambio pequeño entendido va a `flow-build` y el resto a `flow-plan`, ambos términos con enlace a la hoja, y ««Delivery» decides the review, not the plan.» |
| `global.md:51` (tarea operativa) | `plan/proportionality.md` | Nada: el enrutado ya cita la hoja |
| `global.md:16` (procedimiento del proyecto externo) | `plan/scope-changes.md`, nueva sección «An external project» | El límite (cambiar solo el proyecto actual, leer uno externo sin editarlo) y «When the code the work changes lies in an external project, ask as [scope changes](…#an-external-project) sets before the first plan, branch or edit there.» |

`close-cleanup.md` empieza con «Read at a work item's close, before removing, stopping or deleting anything.»; su línea del índice va bajo «Close». `scope-changes.md` amplía su primera línea con «, or when the reads show that the code the work changes lies in an external project», y su línea del índice cambia igual. `proportionality.md` empieza con «Read when choosing a route, a delivery or a verification for a piece of work, and when a rule names a mechanical change, a small understood change or an operational task.» Su línea del índice, bajo «Plan», cita esa primera frase.

## Leaf size

Ninguna hoja necesita excepción: `close-cleanup`, `proportionality` y `scope-changes` quedan bajo 4.000 B, y `close-hygiene` no crece. El paquete de flow-close no cambia.

## Citations

Además de las citas a «Proportionality»: `global.md:21` (tarea operativa) cita `proportionality`; `defect-fix.md:8` cita `scope-changes` para el proyecto externo; `close-hygiene.md:11` y `flow-close/references/branch-cleanup.md:3` citan `close-cleanup` para lo autorizado al cerrar; `flow-close/SKILL.md` enlaza `close-cleanup` en su paso de limpieza; `flow-build/SKILL.md:22`, `delivery-ask.md:5` y el paso de enrutado de `hive-mode` enlazan `proportionality` donde aplican sus definiciones. `workflow-lifecycle/spec.md:807` sigue siendo cierto («Scope and authorization» fija que se pregunta, como fija `scope-changes`) y no cambia.

Los archivos que citan «Proportionality» de la guía global por una definición movida pasan a citar `[proportionality](…/plan/proportionality.md)`, con la ruta que use cada archivo (`skill:` o relativa). Las citas que usan la lectura grande, el cambio riesgoso o el enrutado no cambian. El cambio de ubicación no cambia ninguna regla: los tests de casa única registran la frase en su nueva casa.

## Impact

`content/guidance/global.md`; 2 hojas nuevas y 3 existentes de `hive-core`; `flow-close` y `branch-cleanup`; hasta 15 archivos de `content/` con citas; `tests/content/single_home_test.go`, `leaf_links_test.go` y `reading_footprint_test.go`; `_support/docs/architecture/instruction-resources.md`; `CHANGELOG.md`. `README.md:407` y `rule-enforcers.md:57` siguen siendo ciertos y no cambian. Los paquetes de etapa no crecen: ninguna hoja nueva entra en un `stage.json`.

## Recovery

Revertir el commit restaura cada texto en el núcleo; el siguiente `hive update` lo despliega.
