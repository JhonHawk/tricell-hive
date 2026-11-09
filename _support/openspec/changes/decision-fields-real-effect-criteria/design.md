# Diseño

## Enfoques

**Contexto de una pregunta (D6).**

- A, elegido — Campos con etiqueta. Cada decisión lleva líneas fijas. Un modelo literal las cumple y una revisión de sesión las comprueba a simple vista; cada pregunta ocupa más.
- B — Prosa con pruebas observables («ningún identificador sin explicar», «la razón es una consecuencia»). Conserva el aspecto actual; es más difícil de comprobar.
- C — Lector independiente: un hijo lee la ronda en frío antes de enviarla. Mide comprensión real; cuesta unos 40 mil tokens y minutos por ronda.

Se eligió A porque convierte el contexto en algo tan verificable como el formato, que es lo que hoy sí se cumple.

**Verificación contra el efecto real (D7).**

- A, elegido — En el criterio: cada criterio dice dónde se observa y un sustituto no lo cierra.
- B — En los roles verificadores: clasifican su evidencia como real o sustituta. Solo actúa cuando se lanza un verificador.
- C — Solo medir con la rúbrica.

Se eligió A porque es donde falló la sesión de origen y ninguna regla lo cubría.

## Decisiones

- El formato con campos aplica a las decisiones `D<n>`. Los hallazgos `H<n>` conservan su tabla, que ya trae qué, dónde y evidencia, y las respuestas del bloque de siguiente paso conservan su lista con letras.
- La sección «Questions and routes» crece. Los campos añaden entre 450 y 650 bytes sobre 11,047 y no se paga moviendo otra regla: la única candidata, la respuesta de compactación, debe actuar después de compactar, cuando una referencia ya no está en contexto (`_support/docs/architecture/compaction-survival.md:65-67`). El build ajusta la redacción de los puntos que edita para crecer lo menos posible, sin quitar ningún requisito. Si la sección pasa de 11,700 bytes o `global.md` de 45,150, el build no recorta por su cuenta: lo dice en la validación y el usuario decide.
- La cuenta de decisiones restantes va dentro de la línea de espera, no en una línea propia: la ronda sigue cerrando con una sola línea antes de la tarjeta.
- En la tarjeta, la razón de la recomendación va al final de la descripción de la opción recomendada, porque la tarjeta debe entenderse sola y no tiene otro lugar para ella.
- Una tarea que no alcanza el lugar de un criterio lo sigue nombrando en `Closes:`, con la etapa que lo alcanza, y lo reporta como no verificado hasta entonces. Así cada criterio conserva su tarea, `check_change.py` y `hive doctor` no cambian su comprobación de cobertura, y no hace falta inventar tareas de verify o de deploy.
- Qué cuenta como efecto real y qué como sustituto (D9 A): el candidato servido en local con el build real del proyecto y datos sembrados de volumen realista es el efecto real de una pantalla o una API que el repositorio sirve, que es como `flow-verify` ya recorre el cambio; la emulación o el mock de un servicio que el repositorio no posee, como una nube, un proveedor de pagos o un sistema de otro equipo, es un sustituto.
- Un criterio que nombra un entorno desplegado no declara el recorrido funcional de ese despliegue (D10): tras desplegar basta la comprobación de que la aplicación responde, que `flow-deploy` ya hace, y recorrer todo o puntos concretos es a petición del usuario, porque varios tickets se acumulan antes de un despliegue. Ese criterio queda sin verificar hasta que el usuario pida el recorrido. Por eso la regla pide nombrar un entorno desplegado solo cuando el efecto no existe en ningún otro lugar, como en un cambio de infraestructura; es una consecuencia de esa respuesta, no una decisión aparte.
- La comprobación de `Observed in:` no corre con `--archive`. Una carpeta en curso de otro proyecto recibe el hallazgo en su siguiente revisión, cuando el agente ya está editando criterios; una carpeta terminada se archiva sin tocarla. Es la decisión de compatibilidad del formato de la carpeta de cambio, y `CHANGELOG.md` la anota como nota de migración.
- No se añade un detector en `tests/pilot`: los pilotos están pausados y `AGENTS.md` pide no añadir casos de regresión para un fallo visto en sesión real. La comprobación vive en las filas del 2026-11-08 de la rúbrica de revisión.

## Forma propuesta

Texto en inglés, como todo el contenido distribuido. Redacción de partida; el build la ajusta sin cambiar lo que exige, y conserva las frases que las verificaciones buscan.

**`global.md`, punto del formato de decisión.** Sustituye «followed by the context needed to answer it: what you verified and what to check, with its target», «a one-line description of its effect» y la frase «For a choice that changes scope … what remains uncertain»:

> Give each decision an ID, `D1`, `D2`…, on its own title line that holds the question itself, such as `D2 — Do you accept the mobile search width?`, never a bare topic. Under it write a `Situation:` line: what happened or what you found, what you verified and what to check, with its target, in words the user has used or that you explain there; name no ID, file, branch or internal name the same message does not explain. Then give each option a line of its own, a letter, a dash and its name, the recommended option first with `(★)` after the dash, and under it an `If you choose it:` line saying what the user gets and what it costs them in time, money, risk or what stops working. Close the decision with an `I recommend <letter>:` line whose reason is a consequence for the user, never a quotation of a rule or a skill: when a rule makes you ask, say the effect that needs the answer, such as "this creates 85 resources at about USD 300 a month"; for a choice that changes scope, a contract, stored data or cost, or is hard to reverse, add what remains uncertain. Write the three labels in the session language. In the round's waiting line, say also how many decisions you know are still to come.

El resto del punto (hechos verificados frente a supuestos, hallazgos `H<n>`, IDs y estrella en todo idioma, numeración, glosa) no cambia.

**`global.md`, punto de la tarjeta.** «each question opens with the stage and ticket; the question carries the situation, with what the user needs to answer, such as the plan path; each option's label is `A - (★) <two to five words>` and its description carries the `If you choose it:` text, the starred one ending with the reason of the `I recommend` line».

**`global.md`, «Delivery».** «Give the reason in one line next to the star» pasa a decir que la razón va en la línea `I recommend`; «followed by one line on what the change touches» pasa a decir que lo que toca el cambio va en la línea `Situation:`; las dos menciones «in the context» (una revisión que el repositorio ya corre en cada pull request, y la revisión preferida que no corre en este host) pasan a «in the `Situation:` line». Los nombres fijos de las opciones de revisión no cambian.

**`viewer-publication.md`.** La opción «Open the viewer» no lleva línea `If you choose it:`; la frase «This changes only three rules» pasa a «four rules» y nombra la cuarta; al volver a preguntar tras abrir el visor se repite la tarjeta con su situación y sus opciones, y el texto visible sigue siendo una línea.

**`flow-plan` «Acceptance criteria».** La frase «Every criterion appears in the `Closes:` line of at least one task, whose verification covers it» pasa a «…whose verification covers it or names the stage that reaches its place», y se añade:

> End each criterion with `Observed in: <place>`: where its real effect shows. For a screen or an API the repository serves, the candidate served locally with the project's real build and seeded data of realistic volume is the real effect. A mock, a stub or an emulation of a service the repository does not own, such as a cloud account, a payment provider or another team's system, and a dry run, are substitutes: a check against one supports a task and never closes a criterion. A task that cannot reach a criterion's place still names it in `Closes:` and reports it as not verified until the stage that reaches it: the walk of `flow-verify`, or, for a deployed environment, the functional walk of `flow-deploy`, which runs when the user asks for it. Name a deployed environment only when the effect exists nowhere else, such as a change to infrastructure.

En «Revise a plan in progress», la frase «Both markers stay in English» nombra también `Observed in:`.

**`check_change.py`.** Un hallazgo `proposal.md:<n>: AC<k> names no **Observed in:** place` por cada criterio activo (ni retirado ni reemplazado) cuya viñeta, con sus líneas de continuación, no trae el marcador, aceptado como `Observed in:`, `**Observed in:**` o `**Observed in**:`. No se evalúa con `--archive`. Las comprobaciones existentes y los códigos de salida no cambian; los criterios de las carpetas de prueba existentes ganan el marcador para seguir pasando.

**Punteros.**

- `flow-build` paso 2: un criterio cuyo lugar la tarea no alcanzó se nombra en `--unverified`.
- `hive-verify-task`: `met` exige evidencia del lugar que nombra el criterio. La frase «When the task's verification does not cover a criterion the task closes, the verdict is `not met` because of the plan» se limita a los criterios cuyo lugar la tarea alcanza; cuando el lugar es de una etapa posterior o la evidencia es de un sustituto, el veredicto es `cannot verify`, con el lugar o el sustituto nombrado, y es el resultado esperado.
- `hive-verify-change`: una línea que apunta a la misma regla, porque califica cada criterio por su comportamiento observado.
- `flow-verify` paso 1: cada criterio se comprueba en el lugar que nombra, y los que nombran un entorno desplegado quedan sin verificar en su reporte.
- `flow-deploy` `deployment-checks.md`, «Functional walk»: cuando el recorrido corre, cierra los criterios cuyo `Observed in:` nombra ese entorno; cuando no corre, siguen sin verificar y el reporte del despliegue los nombra. Cuándo corre no cambia.
- `flow-walkthrough`: los datos stub sirven a la primera vista del build; el recorrido de verify usa el candidato servido con datos sembrados.

**`CHANGELOG.md`.** Una entrada en `[Unreleased]` con los campos de una decisión y el marcador, y la nota de migración de la carpeta de cambio.

## Ajustes de la revisión de código

La redacción final vive en los archivos; estos puntos corrigen la «Forma propuesta» de arriba donde difieren.

- `flow-plan`: la regla habla de lo que «the project» sirve y posee, que cubre el repositorio y su workspace declarado. El recorrido de `flow-verify` cierra cada criterio en su lugar; solo el criterio con evidencia de un sustituto, o cuyo lugar es un entorno desplegado, se reporta sin verificar. Nombrar el entorno no pide el recorrido de deploy, y ningún texto fuera de `deployment-checks.md` repite cuándo corre.
- `hive-verify-task`: `met` y `not met` por el plan quedan como estaban; `cannot verify` se añade para la evidencia de un sustituto y para un criterio que nombra un entorno desplegado.
- `hive-verify-change`: observa en el entorno que su encargo le asigna y dice si es el lugar que el criterio nombra.
- `flow-verify` y los verificadores: un criterio sin marcador, de una carpeta anterior a este cambio, se verifica por su ruta real como antes.
- `global.md`: los hallazgos de una ronda llevan su título y sus opciones, sin las tres líneas con etiqueta; en la tarjeta, una opción sin línea de consecuencia lleva su efecto en una línea.
- `check_change.py`: lee las líneas indentadas, sub-viñetas y texto envuelto del criterio, y exige un lugar tras el marcador; un marcador vacío o un `<…>` sin rellenar no cuenta.
- Sitio: la página de plan nombra la comprobación nueva y la de deploy dice qué cierra su recorrido.
- `flow-close`: registrar el criterio sin verificar como trabajo pendiente se intentó y se revirtió tras su revisión de código; el hueco sale de este cambio y va a un ticket (D12 A).

## Impacto

Del barrido del 2026-11-08 (ripgrep sobre `content/`, `site/`, `_support/docs/`, `_support/openspec/specs/`, `tests/`, `tooling/`, `README.md`, `llms.txt`; `integrations/` solo por cadenas de tarjeta) y de las dos revisiones del plan:

- Se editan: `content/guidance/global.md` («Questions and routes» y «Delivery»); `content/skills/flow-plan/SKILL.md`, su plantilla `proposal.md` y `scripts/check_change.py`; `content/skills/flow-build/SKILL.md`; `content/skills/flow-verify/SKILL.md`; `content/skills/flow-deploy/references/deployment-checks.md`; `content/skills/flow-walkthrough/SKILL.md` y `references/viewer-publication.md`; `content/agents/quality/hive-verify-task.md` y `hive-verify-change.md`; `tests/skills/change_check_test.py` y `tests/skills/tasks_test.py`; `site/src/content/docs/flows/plan.md` y `verify.md`, y `deploy.md` solo si su descripción del recorrido queda inexacta; `_support/docs/architecture/agent-delivery.md` e `instruction-resources.md`; `CHANGELOG.md`.
- Se confirman sin cambio: los ocho punteros de skills a «Questions and routes» (nombran la sección, no su formato); `flow-plan/references/plan-handoff.md` y `flow-close/SKILL.md` (la respuesta de compactación no se mueve); `local-instructions/SKILL.md:47-49` (define opciones fijas, compatible); `content/voices/preamble.md`; `flow-verify/references/ui-verification.md` (sus «realistic data» coinciden con la definición de efecto real); `tooling/cli/doctor_project_changes.go` (su comprobación de cobertura de criterios sigue valiendo, porque cada criterio conserva una tarea en `Closes:`); `site/.../host-limits.md`, `README.md:175-179` y `llms.txt:31` (capacidad de tarjetas por host); `_support/openspec/specs/context-delivery/spec.md` (el núcleo global sigue alojando preguntas y rutas) y `walkthrough-viewer/spec.md:134` (la opción del visor sigue siendo la última y sin estrella).
- Pruebas que pueden romperse: `tests/skills/tasks_test.py:179-198` y varios casos de `change_check_test.py`, cuyos criterios de prueba no llevan el marcador y esperan salida 0; `tests/content/skill_size_test.go` (cada `SKILL.md` y referencia bajo 16,000), `skill_links_test.go` (enlaces `skill:` de `global.md`), `reading_footprint_test.go` (fuentes de los escenarios), `agent_boundary_test.go` (bloque de `hive-verify-task`); `tooling/management/antigravity_test.go:330`, que cubre el reparto de `global.md` en archivos de Antigravity con un presupuesto de 22,000 bytes por parte. Los detectores de `tests/pilot/regression.go` leen IDs y texto previo a la pregunta, que no cambian; `qualify_suite_test.go:321` afirma `D1 — ¿Aceptas el cambio?` y `A - (★) Sí`, que siguen siendo válidos. Ninguna prueba hace cumplir los topes de tamaño de la sección: los mide la verificación de T1.
- No toca producción ni datos. Otros proyectos lo reciben solo cuando se confirme y despliegue.

## Recuperación

Con entrega local, `git restore` de los archivos listados devuelve el estado anterior; no hay nada que revertir fuera del árbol de trabajo. Una vez desplegado, se revierte con el commit y un nuevo `hive update`.
