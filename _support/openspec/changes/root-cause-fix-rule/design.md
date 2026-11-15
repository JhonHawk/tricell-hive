# Design

## Approaches

- **A — Referencia nueva con puntero (elegido, D1 A).** `flow-build/references/defect-fix.md` tiene la regla completa; `flow-build/SKILL.md` la enlaza con su condición. Deja la regla en un solo sitio sin recortar reglas que hoy funcionan y solo se lee cuando hace falta; el costo es que el agente abre un archivo más.
- **B — Todo dentro de `SKILL.md`, recortando ~200 bytes.** Se carga siempre, pero cada recorte necesita revisar que no pierde intención.

## Decisions

- **D2 A** — Stored data y Security boundaries pasan a `references/stored-data.md` y `references/security-boundaries.md`. Ya tienen una condición de lectura propia («Before choosing a technique» sobre datos guardados; «When the change touches untrusted input…»), así que el puntero conserva el disparador. Tests se queda: casi cada tarea lo usa.
- **D6 A** — Se quita la sección «Delegation» (250 B): repetía el puntero a `delegation.md` del paso 2 y de la guía global «Proportionality», y nada enlaza `#delegation`. Margen estimado tras el cambio: ~1.120 bytes. Se quedan Checkpoint y reparación del entorno (aplican en todo build) y la primera vista de UI (paso del flujo).
- **Prioridad del hallazgo de revisión (supuesto):** sin prioridad fija; `hive-review-code` lo clasifica por impacto observable, como cualquier hallazgo.
- **Sin prueba de redacción (supuesto):** la regla tiene excepciones (estado legítimo, causa fuera del alcance); fijar la frase no comprueba comportamiento. Las pruebas de tamaño y enlaces existentes cubren la reestructuración.
- **Marca en el código del parche inevitable (supuesto de diseño):** una línea de comentario que nombra la causa (no el ticket, que puede no llegar a existir). Es lo que permite al siguiente agente reconocerlo y no copiarlo (peldaño 2), que es el mecanismo de propagación del post.
- **Alcance definido (revisión del plan, H3):** la causa está fuera del alcance cuando vive en un módulo que el trabajo no toca, en otro repositorio, o exige cambiar un contrato o datos guardados; las otras instancias de la misma falla se arreglan en los módulos que el trabajo ya toca y las demás se reportan como hallazgos. Así no choca con «Fix only what each finding names» de `flow-verify`, con la P2 de `hive-review-code` ni con «Scope changes».
- **Hijos (H4):** un hijo arregla solo dentro de las rutas de su brief y reporta al padre lo demás; nunca propone tickets.
- **Incidentes (H7):** la mitigación de un incidente en producción entra en el paso 4 (parche marcado y reportado), en lugar de chocar con «no es un arreglo».
- **`global.md` no cambia:** la regla actúa al editar, momento en que el agente ya está en `flow-build`, `flow-verify` o un rol `hive-build-*` que enlaza la referencia.

## Proposed shape

`content/skills/flow-build/references/defect-fix.md` (borrador; el build ajusta redacción y bytes):

```markdown
# Defect fix

Read before the first edit of a fix for a defect, and when a plan weighs the approaches to one: in a build task, a round of fixes, an urgent production fix or a child's task, whether the user reported the defect or a check, a review or a verification found it. When a check fails, first decide as [Tests](../SKILL.md#tests) sets whether the code is wrong; this applies once it is.

1. Name the cause before you edit, and write it in the task's line or the report. The cause is the first place the code does what it should not, or fails to handle a state it should handle; the place where the failure surfaced is the symptom. Show it with the reproduction, the regression test seen failing, or the code you read.
2. Fix it at the cause, with the smallest change that removes it, also when a finding names only the place where the symptom surfaced. Search the cause's literal values and callers for the same fault: fix each instance in the modules the work already touches, and report the others as findings.
3. A guard, retry, fallback, default value, clamp, swallowed error or discarded state added where the symptom surfaced is not a fix, unless the system may legitimately reach that state; then handling it there is the fix.
4. The cause is outside the work's scope when it lies in a module the work does not touch, in another repository, or needs a change to a contract or stored data. Then, and while an incident needs its impact reduced before the cause is fixed, make the smallest fix inside the scope, mark it in the code with one comment line that names the cause, and name it in the report as a workaround with the cause's location. Propose fixing the cause as the global guidance's «Scope changes» sets for a ticket; a change to a contract, stored data or the scope is the user's choice, as its «Scope and authorization» sets for a defect that blocks work.
5. A child fixes only within the paths its brief names, and reports to its parent each other instance, workaround and cause outside them instead of proposing a ticket.
```

Punteros en `flow-build/SKILL.md`:
- Paso 2: «…applying [Reuse before writing](#reuse-before-writing) before writing new code, [defect fix](references/defect-fix.md) before the first edit of a defect fix, and [Tests](#tests)…».
- «Reuse before writing», tras la primera línea: «A task or an approach that fixes a defect first reads [defect fix](references/defect-fix.md).» Así lo alcanzan los roles `hive-build-*` y el paso 2 de `flow-plan`, que enlazan esta sección.
- Tests «When a check fails»: cuando el código es el que está mal, lo arregla como sets [defect fix](references/defect-fix.md) (si el presupuesto lo permite; si no, basta el puntero del paso 2).
- Peldaño 2: «…use it as the surrounding code does, unless it is a workaround, such as one a comment marks or a guard for a state that should never exist: do not copy it, and name it in the report.»
- En lugar de las secciones movidas: «## Stored data» y «## Security boundaries» quedan como una línea cada una con su condición y el enlace, o un único párrafo «Situational rules» con ambas; el build elige lo que preserve las anclas más útiles para quien lee.

`flow-verify/SKILL.md` paso 3: «fix in one round…», con «each defect at its cause, as [defect fix](../flow-build/references/defect-fix.md) sets, even when the finding names only the symptom». Es la excepción explícita a «Fix only what each finding names».

`flow-deploy/references/urgent-fixes.md:9`: la mitigación temporal de un incidente se marca y se reporta como sets [defect fix](skill path relativo) paso 4.

`hive-review-code.md`:
- Paso 2, en la lista de ángulos: «a defect fix placed where the symptom surfaced instead of at its cause, such as a guard, retry or fallback for a state that should never exist, and new code that copies a workaround»; un arreglo en la causa no es «a change the task did not ask for».
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
