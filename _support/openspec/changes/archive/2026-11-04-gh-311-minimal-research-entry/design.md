# Diseño

## Decisiones

Están en `proposal.md`. Aquí solo la forma que toman en el texto.

## Forma propuesta

### Research mínimo (`flow-research/SKILL.md`)

Una guía de no omisiones, sin visor, en este orden:

0. **Pedido completo.** Si el pedido ya dice todo lo que cambia el resultado, revisar que no haya bloqueos y hacerlo: un cambio al repositorio pasa a `flow-build` en el mismo turno; una acción sin cambio al repositorio se hace y se dice qué se hizo. Las reglas de efectos de la guía global siguen aplicando.
1. **Leer antes de preguntar:** código, ticket, configuración y memoria verificada; nunca preguntar lo que eso ya responde.
2. **Devolver el pedido:** qué, para quién y por qué, con lo encontrado: un conflicto, una premisa falsa o un camino más barato.
3. **Preguntar lo que falta:** cuando falta algo que cambia lo que se construye, una pregunta abierta por mensaje, en texto, con su razón, y la salida «o arranco ya». Las tarjetas quedan para elegir entre opciones.
4. **Investigar lo que no se sabe:**
   - hijos `hive-research` en paralelo cuando un aspecto necesita más de 5 lecturas grandes (ver `flow-build/references/delegation.md`); el explorador integrado del host solo para ubicar archivos, y su resultado se verifica;
   - probar cada causa posible de una falla y reportar las descartadas;
   - nombrar cada sistema que la decisión toca y si se buscó;
   - abrir la fuente de cada afirmación que decide, en vez de recordarla;
   - buscar si ya existe un ticket.
5. **Cerrar el qué:** devolver lo entendido separando lo dicho por el usuario de lo propuesto. Si pidió analizar, explicar o saber por qué, terminar en los hallazgos con un pie de opciones sin flujo recomendado. Si pidió un cambio: a `flow-build` en el mismo turno cuando el cómo está resuelto, o el bloque de siguiente paso a `flow-plan` cuando el cómo tiene una decisión suya o no se puede retomar con el pedido y el diff.

Se conserva la evaluación de ticket como lista corta: impacto, complejidad, preparación y dependencias. Sale: el visor, la oferta de `flow-report`, la ruta obligatoria y la apertura de tickets dentro de una opción de ruta, que pasa a ser una acción suelta.

Peso: hoy 5.075 bytes más ~14.800 del visor que obliga a leer; objetivo 4.000–4.500 bytes sin lecturas obligatorias adicionales (no se cumplió: 4.574 al entregar y 5.252 tras las correcciones de sesiones reales).

### Pie de cierre (`flow-close`)

Al integrar, el ticket se mueve y sale el reporte final. Termina con un pie, no con una tarjeta:

```
---
**▶ Siguiente paso: cerrar (recomendado)**
<una línea con la razón>
Otras respuestas:
- resolver H2 y H3 en esta rama
- explicar H1
- siguiente ticket: <ticket> (<razón>)
- seguir con <lo pendiente>
```

Las alternativas salen del contexto: los hallazgos sin resolver, lo pendiente del trabajo y el siguiente ticket recomendado. Cierre del visor, archivo del registro, limpieza de rama y worktree y escritura de memoria corren al elegir cerrar, al empezar otro trabajo (primero ese cierre, dicho en una línea) o al terminar la sesión. Bajo delegación desatendida el cierre corre completo sin pie.

### Visor

La sesión de visor empieza al entrar a plan o build, con un resumen de casi un párrafo de lo que trata el trabajo. Research no publica.

### Carpeta versionada con el código y pendientes (T8–T11, 2026-11-04)

Enfoques para el primer commit de la carpeta de cambio:

- **A — Con la entrega del código (elegido, D26 A).** Con `Specs` en el mismo repositorio, la carpeta va en la rama del trabajo antes del push y se integra con el PR; con `Specs` aparte, se commitea allí en el punto de completitud. El cierre solo la mueve a `archive/`. Coincide con la regla del usuario y no cambia cómo se escribe el plan.
- **B — Al aprobar el plan.** Como en septiembre de 2026 (`docs(openspec): plan COM-…` en cam-specs). Protege desde el primer día, pero cada revisión del plan deja un commit y, con `direct-base`, publica en la base planes sin construir.
- **C — Solo al pausar o esperar.** Menos commits, pero un trabajo integrado sin pendientes queda sin commit hasta la respuesta del pie (caso COM-915).

Casa canónica: `git-workflow` paso 4 dice cuándo y dónde se commitea, incluido el traslado al worktree; `flow-close` lo ejecuta en la primera parte del cierre para el repositorio aparte, ligado al punto de completitud, y solo enlaza; `plan-handoff.md` apunta. La carpeta se sigue archivando al cerrar aunque queden pendientes: estos viven en el ticket y en el registro. El registro `pending/<change-id>` vive en `flow-close` paso 7; `flow-status` solo lo lee. La guía global dice únicamente las dos autorizaciones (comentario y borrado), que deben actuar sin leer `flow-close`.

## Impacto

- **Guía global:** «Proportionality» (entrada y criterio de plan), «Questions and routes» (bloque de siguiente paso, excepción de preguntas abiertas, continuación en el mismo turno), «Preservation» (momento de la limpieza del cierre), «Scope changes» (criterio de cambio pequeño), Context7.
- **Skills:** `flow-research`, `flow-plan`, `flow-build` (entrada y paso 5), `flow-verify` (paso 6), `flow-close`, `flow-deploy` (paso de cierre), `git-workflow` (paso 10), `flow-walkthrough` y `references/viewer-publication.md`.
- **Rol:** `hive-research` (línea de Context7).
- **Documentación:** `site/src/content/docs/flows/{overview,research,plan,build,close}.md`, `README.md:317-327`, `llms.txt:33-37`, `_support/docs/architecture/instruction-resources.md` y `agent-delivery.md` si nombran estas reglas.
- **Tests:** `tests/content/skill_size_test.go`, `reading_footprint_test.go` y `budget_test.go` por los tamaños; `tests/pilot/regression.go` (detector `close_question_after_report`) si depende de la tarjeta de cierre. Los casos de `tests/fixtures/` no cambian.
- No toca producción: el contenido llega a las máquinas solo con `hive update`.

## Recuperación

Revertir el merge del PR y correr `hive update`. Nada queda irreversible: no hay datos guardados ni contratos de CLI.
