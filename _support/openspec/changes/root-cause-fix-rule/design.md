# Design

## Approaches

- **A — Referencia nueva con puntero (elegido, D1 A).** `flow-build/references/defect-fix.md` tiene la regla completa; `flow-build/SKILL.md` la enlaza con su condición. Deja la regla en un solo sitio sin recortar reglas que hoy funcionan y solo se lee cuando hace falta; el costo es que el agente abre un archivo más.
- **B — Todo dentro de `SKILL.md`, recortando ~200 bytes.** Se carga siempre, pero cada recorte necesita revisar que no pierde intención.

## Decisions

- **D2 A** — Stored data y Security boundaries pasan a `references/stored-data.md` y `references/security-boundaries.md`. Ya tienen una condición de lectura propia («Before choosing a technique» sobre datos guardados; «When the change touches untrusted input…»), así que el puntero conserva el disparador. Tests se queda: casi cada tarea lo usa.
- **D6 A** — Se quita la sección «Delegation» (250 B): repetía el puntero a `delegation.md` del paso 2 y de la guía global «Proportionality», y nada enlaza `#delegation`. Margen estimado tras el cambio: ~1.120 bytes. Se quedan Checkpoint y reparación del entorno (aplican en todo build) y la primera vista de UI (paso del flujo).
- **Prioridad del hallazgo de revisión (supuesto):** sin prioridad fija; `hive-review-code` lo clasifica por impacto observable, como cualquier hallazgo.
- **Sin prueba de redacción (supuesto):** la regla tiene excepciones (estado legítimo, bloqueo antes del parche, mitigación de incidentes); fijar la frase no comprueba comportamiento. Las pruebas de tamaño y enlaces existentes cubren la reestructuración.
- **Marca en el código del parche inevitable (supuesto de diseño):** una línea de comentario que nombra la causa (no el ticket, que puede no llegar a existir). Es lo que permite al siguiente agente reconocerlo y no copiarlo (peldaño 2), que es el mecanismo de propagación del post.
- **Alcance (revisión del plan H3, corregido por `/code-review` del candidato):** la causa se arregla donde esté en el repositorio; las otras instancias, en los archivos que el trabajo ya cambia (la misma unidad que «Scope changes»), y las demás se reportan. Cuando arreglar la causa exige un repositorio fuera del proyecto actual, un contrato, datos guardados u otra elección del usuario (como un arreglo que no es un cambio pequeño y entendido), es un bloqueo para el usuario antes de cualquier parche, como fija «Scope and authorization». «Legítimo» se observa: el contrato, los tipos, el esquema o la documentación permiten el estado, o viene de una entrada que cruza un límite de confianza. **Replaces:** la definición por «módulo que el trabajo no toca».
- **Hijos (H4):** un hijo arregla solo dentro de las rutas de su brief y reporta al padre lo demás; nunca propone tickets.
- **Incidentes (H7):** la mitigación de un incidente en producción entra en el paso 4 (parche marcado y reportado), en lugar de chocar con «no es un arreglo».
- **`global.md` no cambia:** la regla actúa al editar, momento en que el agente ya está en `flow-build`, `flow-verify` o un rol `hive-build-*` que enlaza la referencia.

## Proposed shape

`content/skills/flow-build/references/defect-fix.md` es el texto vigente; el borrador del plan quedó reemplazado por la corrección de la revisión del candidato (ver «Alcance»).

Punteros en `flow-build/SKILL.md`:
- Paso 2: «…applying [Reuse before writing](#reuse-before-writing) before writing new code, [defect fix](references/defect-fix.md) before the first edit of a defect fix, and [Tests](#tests)…».
- «Reuse before writing», tras la primera línea: «A task or an approach that fixes a defect first reads [defect fix](references/defect-fix.md).» Así lo alcanzan los roles `hive-build-*` y el paso 2 de `flow-plan`, que enlazan esta sección.
- Tests «When a check fails»: cuando el código es el que está mal, lo arregla como sets [defect fix](references/defect-fix.md) (si el presupuesto lo permite; si no, basta el puntero del paso 2).
- Peldaño 2: «…use it as the surrounding code does, unless it is a workaround, such as one a comment marks or a guard for a state the code's contract forbids: do not copy it, and name it in the report.»
- En lugar de las secciones movidas: «## Stored data» y «## Security boundaries» quedan como una línea cada una con su condición y el enlace, o un único párrafo «Situational rules» con ambas; el build elige lo que preserve las anclas más útiles para quien lee.

`flow-verify/SKILL.md` paso 3: «fix in one round…», con «each defect at its cause, as [defect fix](../flow-build/references/defect-fix.md) sets, even when the finding names only the symptom». Es la excepción explícita a «Fix only what each finding names».

`flow-deploy/references/urgent-fixes.md:9`: la mitigación temporal de un incidente se marca y se reporta como sets [defect fix](skill path relativo) paso 4.

`hive-review-code.md`:
- Paso 2, en la lista de ángulos: «a defect fix placed where the symptom surfaced instead of at its cause, such as a guard, retry or fallback for a state the code's contract forbids, and new code that copies a workaround»; un arreglo en la causa no es «a change the task did not ask for».
- Paso 3, el escenario de esos hallazgos: la causa sigue produciendo el estado malo para sus otros consumidores, o el parche copiado se extiende; sigue el precedente de `CODING_STANDARDS.md`, donde la violación sustituye al escenario.

## Impact

- Enlaces que cambian de ancla a archivo: los 7 de «Verified context» de `proposal.md`; en roles con `skill:flow-build/references/<file>.md`, en skills con ruta relativa.
- `_support/docs/architecture/instruction-resources.md`: el índice de recursos canónicos gana las tres referencias nuevas.
- `site/src/content/docs/flows/build.md`: una frase en la lista de comportamientos y «does not copy a workaround» en la línea 26 (reutilización); `site/src/content/docs/flows/verify.md:25` (ronda de arreglos) se revisa y se ajusta si describe el arreglo.
- `flow-deploy/references/urgent-fixes.md:9`: puntero para la mitigación de un incidente.
- `CHANGELOG.md`: entrada bajo `[Unreleased]` (`CONTRIBUTING.md:148`).
- `tooling/management/voice.go:220` cita un `security-boundaries.md` anterior: se comprueba si el archivo nuevo encaja con lo que dice el comentario; si no, se reescribe el comentario (solo comentario).
- `tests/content/reading_footprint_test.go` cuenta `flow-build/SKILL.md` en el escenario del ciclo: baja su tamaño; ningún umbral debería romperse, se comprueba con la suite.
- No toca `global.md`, código Go ni el manager.

## Recovery

Revertir el commit restaura las secciones y enlaces; no hay datos ni efectos externos.
