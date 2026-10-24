# Aprobar los criterios como contrato y declarar en el PR la evidencia de cada criterio (2026-10-24)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · validado por el usuario (D7-A) e integrado en `development` el 2026-10-24 por el [PR #135](https://github.com/JhonHawk/tricell-hive/pull/135), merge `0d12f7c0`; #134 cerrado; registro archivado el mismo día |
| Tracker · GitHub Issues | • [#134 — Plan y PR: aprobar los criterios como contrato y declarar en el PR la evidencia de cada criterio](https://github.com/JhonHawk/tricell-hive/issues/134) |
| Git | `interactive` (D4-A) · rama desde `development` · parada antes del push para que el usuario valide |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...` y `npm run build` en `site/` sobre el candidato final · sin revisión de código dedicada (D5-A) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Mensaje del usuario, 2026-10-24:

> Sigueinte research para hive va sobre PDD

Respuestas siguientes, 2026-10-24, tarjetas nativas:

- D1-A: «Planear aprobación y PR» (el resumen del plan muestra los criterios para aprobarlos y el cuerpo del PR declara por criterio la evidencia, lo no verificado y los supuestos).
- D2-A: abrir un issue en JhonHawk/tricell-hive → [#134](https://github.com/JhonHawk/tricell-hive/issues/134).
- D3-A: la aprobación va en la misma pregunta de ruta; elegir construir aprueba los criterios, una corrección en texto libre revisa el plan antes de construir, y con la implementación ya autorizada solo se muestran.
- D4-A: entrega `interactive`. D5-A: sin revisión de código dedicada.

Criterios de aceptación del issue #134, citados textualmente (leído el 2026-10-24):

> - AC1. El resumen de un plan listo muestra el texto de cada `AC<n>` y pide al usuario aprobarlos o corregirlos en la misma pregunta de ruta.
> - AC2. El cuerpo de un PR autorizado incluye, por criterio, la evidencia que lo respalda, lo no verificado y los supuestos aceptados.
> - AC3. Las páginas de `site/` que describen los flujos de plan y build reflejan ambos cambios.

## Objetivo

La evaluación de Proof-Driven Development (PDD) Kit, de Microsoft Research y GitHub Next, mostró que su valor portable no es el verificador formal (Verus, solo Rust, sin publicar) sino dos hábitos. El primero: el humano aprueba la spec antes del código. El segundo: el PR declara qué quedó probado, qué no y bajo qué supuestos. Los benchmarks señalan que la IA falla más al escribir la spec que al cumplirla ([Verus-SpecGym](https://arxiv.org/abs/2605.26457)), así que la revisión humana rinde más sobre los criterios.

En Hive, los criterios `AC<n>` son esa spec, pero hoy el usuario no los ve al elegir la ruta: el resumen del plan listo muestra objetivo, decisiones, tareas y gates. El PR, por su parte, describe el problema, el comportamiento, los checks y el alcance de un revert, pero no la evidencia por criterio, lo no verificado ni los supuestos. Tras el cambio, el usuario lee y aprueba los criterios en la misma pregunta de ruta, y quien revise el PR ve qué respalda cada criterio y qué queda abierto.

## Alcance y aceptación

Incluido: `content/skills/flow-plan/SKILL.md` (párrafo del resumen del plan listo), `content/skills/git-workflow/SKILL.md` (párrafo del cuerpo del PR), `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/build.md` y `CHANGELOG.md`.

Excluido:
- Escala de evidencia por criterio (prueba formal, test de propiedades, test, observación, lectura de código): R2 de la evaluación, se decidirá aparte.
- Mutación sistemática de criterios: R3 de la evaluación, queda en observación.
- `content/guidance/global.md`: ninguna de las dos reglas vive ahí; su tamaño no cambia.
- `plan-format.md` y `questions.md`: el resumen del plan listo y la pregunta de ruta solo se definen en `flow-plan/SKILL.md`; `questions.md` define la forma de preguntar, que no cambia.
- Deltas de requisitos: ningún spec de `_support/openspec/specs/` describe los flujos de Hive.

Restricciones:
- `content/skills/flow-plan/SKILL.md` se mantiene en 16 000 bytes o menos (`tests/content/skill_size_test.go`).
- `go vet ./...`, `go test ./...` y `npm run build` en `site/` pasan sobre el candidato final.
- Cuando el repositorio tiene plantilla de PR, la plantilla sigue mandando y la evidencia por criterio se acomoda en ella.

- AC1. La regla del resumen del plan listo exige mostrar el texto de cada `AC<n>` y el nombre de cada requisito que la carpeta del cambio agrega, modifica o elimina, y dice que elegir iniciar `flow-build` en la pregunta de ruta aprueba esos criterios como contrato, y que la opción lo dice por sus IDs para que la tarjeta se entienda sola. Una corrección en texto libre no responde la pregunta de ruta: revisa el plan y, tras la revisión (y la re-revisión cuando aplique), muestra los criterios revisados y vuelve a hacer la pregunta de ruta. *Falso en la base cuando* `content/skills/flow-plan/SKILL.md:70` enumera objetivo, decisiones, tareas, gates, entrega, pendientes y ruta, sin el texto de los criterios ni su aprobación.
- AC2. Cuando la ejecución ya está autorizada explícitamente, la misma regla exige mostrar los criterios y continuar, sin pregunta de aprobación. *Falso en la base cuando* la última frase de `content/skills/flow-plan/SKILL.md:70` solo dice «continue as permitted instead of asking», sin mostrar los criterios.
- AC3. La regla del cuerpo del PR exige, para el PR que entrega el cambio cuando la conversación o el registro del cambio fijaron criterios `AC<n>`, listar cada criterio con la evidencia registrada que lo respalda (test, comando u observación), lo que quedó sin verificar (incluido lo aceptado sin verificar por el usuario) y los supuestos aceptados; sin criterios, los checks que corrieron, lo no verificado y los supuestos. El PR que solo publica el registro del cambio queda fuera. *Falso en la base cuando* `content/skills/git-workflow/SKILL.md:32` pide problema, comportamiento, checks y revert, sin evidencia por criterio, sin lo no verificado y sin supuestos.
- AC4. `site/src/content/docs/flows/plan.md` dice que el resumen del plan listo muestra los criterios y que elegir construir los aprueba, y `site/src/content/docs/flows/build.md` dice qué declara el PR por criterio. *Falso en la base cuando* la pregunta frecuente «Does a ready plan mean the agent may start building?» de `plan.md` no menciona los criterios y `build.md` no dice nada del cuerpo del PR.

## Entrega

- Repositorio: `tricell-hive`, base `development`. Modo `interactive` (D4-A, 2026-10-24): rama de trabajo desde `development`, commits locales por tarea verificada, verificación y gates compartidos. Después, parada antes del push con el resultado disponible para que el usuario lo valide, y con despliegue local desde la rama solo si lo pide. Tras su validación: push, PR, merge a `development` con las pruebas locales en verde y `hive update` local por el marcador de maintainer.
- Registro del cambio: se versiona una vez al cierre con `flow-close`, en un PR pequeño de cierre sin revisión dedicada que fusiona el agente.
- Revisión de código: ninguna dedicada (D5-A); el agente revisa el diff y corren las pruebas de contenido. No cruza ningún límite de seguridad.
- UI: sin efecto renderizado en la aplicación; las páginas del sitio cambian solo de texto, así que no hay recorrido de UI.
- Ticket: #134 pasa a en progreso al empezar la implementación y se cierra con el merge a `development`.
