# Diseño

## Contexto verificado

Base: `deb2958` en `rebuild/harness-engineering`, leída el 2026-09-29.

Dónde vive hoy la higiene:

- [`global.md`](../../../../content/guidance/global.md) "Retention and cleanup" (línea 154): limpieza al cerrar la tarea, línea de limpieza en el reporte; línea 144: capturas de UI hasta el fin de sesión; línea 124: fusionar deltas y archivar; línea 40: hallazgos incidentales con dos respuestas (abrir o declinar con conteo).
- [`flow-build/SKILL.md`](../../../../content/skills/flow-build/SKILL.md) "Verify and close": el cierre en tres pasos (limpieza, reporte, pregunta), el cierre del registro vía `change-records.md` y la actualización del estado del plan "version the update at the points the delivery decision names".
- [`git-workflow/SKILL.md`](../../../../content/skills/git-workflow/SKILL.md) "Clean up after a merge": borrado de ramas, `git fetch --prune`, avance de la base y de ramas de entorno; la línea 8 cita esa sección como su única autorización propia.
- [`change-records.md`](../../../../content/skills/flow-plan/references/change-records.md) "Close a change": aplicar deltas, estado cerrado, `git mv` al archivo, `project.md`.
- [`delivery-decisions.md`](../../../../content/skills/flow-plan/references/delivery-decisions.md) líneas 14 y 18: el registro en un repo `direct-base` se commitea y pushea "when the plan is ready, at each delivery milestone the plan records, and at the closing archive"; en otro repo, "when the plan is approved, at each delivery milestone … and at the closing archive".
- [`flow-plan/SKILL.md`](../../../../content/skills/flow-plan/SKILL.md) último párrafo: "Version a saved change folder before returning … commit and push it now".
- [`workspace-conventions/SKILL.md`](../../../../content/skills/workspace-conventions/SKILL.md) línea 22 remite el archivado de cambios a `change-records.md`.

Rondas de corrección: [`verification.md`](../../../../content/skills/flow-build/references/verification.md) línea 67 (commit `8bcbba9`, 2026-09-29 15:54) ya omite los dos hijos de UI entre rondas, pero solo para "layout adjustments", y el gate final no está acotado.

Tests: `verification.md` "Frequency and coverage" pide "existing test-selection tooling where reliable" sin disparadores de ampliación ni tratamiento de una selección vacía; `flow-build` pide la suite completa una vez cuando el CI no la corre.

Evidencia de la auditoría del 2026-09-29 (subagente `hive-research`, transcripciones de cada host):

- ark y sample-project tienen selección de tests que funciona: `vitest related --run --project unit <files>` corrió 92 tests en 1,9 s (ark) y 186 en 2,7 s (sample-project). El `test:related` de ark omite `--project unit` y arrastró 95 archivos de integración con deadlocks `40P01` (Grok, ARK-680).
- Las suites completas de sample-project tardan cerca de un minuto; la integración de ark, de 3 a 23 minutos. OpenCode repitió la integración completa de ark sobre un candidato sin cambios (1377 s).
- `hive-verify-task` tardó de 3 a 6 minutos por pasada para menos de 5 s de tests; ARK-654 T2 tuvo cuatro pasadas (Claude Code, sample-project `5988de05`). Las esperas del CI posterior al merge fueron de 7 a 8 minutos.
- Documentación: Vitest v4.1.6 (ctx7 `/vitest-dev/vitest/v4.1.6`): `related` sigue importaciones estáticas, no dinámicas, y no corre nada cuando no encuentra fuentes relacionadas. No se consultó Vitest 5 (sample-project). La guía no nombra herramientas: pide la selección que el proyecto ya tenga.

## Diseño elegido

### `flow-close`

Una skill nueva en `content/skills/flow-close/SKILL.md`. El catálogo la despliega sola: el gestor recorre `content/skills/*/SKILL.md` (`deployment-manager.md`, línea 131), sin registro aparte.

Disparadores:

1. **Cierre de un ítem de trabajo:** su entrega llegó al punto de cierre (merge a la base con la regla del CI posterior al merge cumplida, push a la base en `direct-base`, working tree verificado en `hold`), o el trabajo se abandona o se pausa.
2. **Fin de sesión:** el usuario termina la sesión.
3. **Petición explícita:** el usuario pide higiene ("haz higiene", "clean up"); se ejecuta lo que aplique de 1 y 2.

Orden al cerrar un ítem (el último paso termina el turno):

1. Tracker: aplicar la regla global de estado del ticket y sus sub-issues.
2. Registro del cambio: aplicar deltas, estado cerrado, archivar y versionarlo una sola vez (ver abajo).
3. Git: la limpieza posterior al merge que hoy está en `git-workflow`, sin cambios de contenido.
4. Temporales y procesos: los de la tarea; conservar evidencia única y material de dueño incierto.
5. Hallazgos incidentales: la lista para la pregunta única del cierre, con las tres respuestas (AC9).
6. Reporte de cierre con su línea de limpieza.
7. Pregunta de cierre global.

Fin de sesión: borrar las capturas de UI creadas por la tarea que el usuario no eligió conservar, detener procesos que sigan vivos y dejar los recordatorios.

Versionado del registro (D1-A, D6-A). Mientras dura el trabajo, `tasks.md` y el estado de `proposal.md` se actualizan en disco en cada hito; nadie los commitea. El cierre ocurre después de integrar el código y, cuando aplica, de que pase el CI posterior al merge (`verification.md` línea 81 no cambia); el estado cerrado cita el commit o PR que integró. Entonces:

- Specs en un repo `direct-base` (ark-specs, sample-project-specs): un commit y push con los deltas aplicados y la carpeta ya en `archive/`.
- Specs dentro del repo de código (este repo): un commit con los deltas aplicados y la carpeta archivada va a la base por la vía de cierre que el usuario eligió en la pregunta de entrega: push directo, o un PR pequeño de cierre sin revisión dedicada, con quién lo mergea. `flow-close` actúa solo sobre esa respuesta registrada; la política del repo no la sustituye (`global.md` línea 37). Orden en este caso: volver a la base y avanzarla, hacer el commit del registro, publicarlo por la vía elegida y después borrar ramas, incluida la del PR de cierre. Como la carpeta nunca se commiteó, la base no tiene una carpeta activa en ningún momento, y ese commit llega después de la revisión del código, así que la regla de re-revisión (`verification.md` línea 41) no se dispara.
- Carpeta que nunca estuvo en Git: se mueve con `mv` y se agrega por ruta; `git mv` solo cuando ya estaba versionada (registros anteriores a este cambio).
- `hold`: el registro queda sin versionar y se reporta como pendiente, igual que el código.
- Pausa: no se archiva; la carpeta queda en disco con su estado y se reporta como pendiente.
- Abandono: se archiva sin fusionar deltas, con la razón en `proposal.md`, y se versiona por la misma vía que un cierre.

Riesgo aceptado por D1-A: mientras dura el trabajo, el plan existe solo en esa copia local. El `git add` por ruta (regla vigente de `git-workflow`) evita que otra sesión lo meta en su commit.

Autorización de borrado (hallazgo B5 de la revisión). Hoy la concede `git-workflow` ("this rule authorizes these deletions", línea 36) y su línea 8 la cita. `global.md` "Retention and cleanup" pasa a conceder en una frase el borrado de las ramas de trabajo mergeadas, con la salvaguarda de no borrar una rama con commits sin integrar, y de sus worktrees solo cuando no tienen archivos sin commitear ni sin seguimiento, sin forzar (bajo D1-A una carpeta de cambio sin seguimiento es la única copia del plan). Texto fijo de la frase: "At close, this rule authorizes deleting merged work branches and their clean worktrees; never delete a branch with unmerged commits or a worktree with uncommitted or untracked files." `flow-close` conserva el procedimiento y cita esa concesión; `git-workflow` línea 8 pasa a decir que la skill no concede acciones Git propias.

Lo que queda en otros archivos:

- `global.md` conserva la autorización y el disparador: qué borra la limpieza del cierre (temporales reproducibles de la tarea, procesos que arrancó, ramas mergeadas) y que el cierre y el fin de sesión corren `flow-close`. Retira el detalle de procedimiento que pasa a la skill. La regla de capturas de UI (línea 144) se queda, porque `flow-build` línea 67 y `verification.md` línea 71 la citan; la línea 123 ("The `flow-plan` skill carries … the closing procedure") y la 124 pasan a nombrar `flow-close`.
- `git-workflow` conserva la verificación de efectos Git y remite la limpieza posterior al merge a `flow-close`; su línea 8 cambia para no citar una sección que ya no está.
- `change-records.md` conserva formatos y deltas; "Close a change" pasa a `flow-close`, que enlaza las secciones de deltas.
- `delivery-decisions.md` deja de fijar puntos de versionado del registro y remite a `flow-close`; cuando las specs viven en un repo de código, cada opción de la pregunta de entrega nombra la vía de cierre del registro (push directo a la base o PR de cierre, y quién lo mergea), para que una sola respuesta la autorice.
- `flow-plan` deja la carpeta sin commitear al devolver el plan.
- `flow-build` sustituye su cierre en tres pasos por "run `flow-close`", y en el párrafo del estado del plan (línea 44) cambia "version the update at the points the delivery decision names" por la actualización en disco.
- Referencias en prosa (hallazgo B4): `delivery-decisions.md` líneas 14 ("the points the change-folder paragraph below names") y 32 ("the change record's versioning points"), `flow-build` líneas 44 y 75 y `global.md` línea 123 se reescriben; el chequeo de enlaces no las ve.
- `workspace-conventions` remite el archivado de cambios a `flow-close`.

### Correcciones durante la validación

`verification.md` "Human UI handoff" generaliza la regla de `8bcbba9`:

- Mientras el usuario pide correcciones en una pausa de validación (texto, comportamiento o layout), el implementador aplica cada una con su test cuando cambia comportamiento del tipo TDD, corre los tests afectados, reconstruye, captura las vistas afectadas y vuelve a preguntar. Entre rondas no corren `hive-review-ux`, `hive-verify-change`, `hive-verify-task` ni la revisión de código dedicada.
- Cuando el usuario acepta, corre un único gate acotado a la diferencia entre el último candidato que pasó el gate y el aceptado: `hive-review-ux` sobre las vistas cambiadas, `hive-verify-change` sobre los recorridos que tocan archivos cambiados y `hive-verify-task` sobre las tareas cuyos archivos cambiaron. Si todavía no había corrido ningún gate, corre completo.
- Un hallazgo Blocker o High de ese gate se corrige, se muestra al usuario y se revisa de nuevo solo con el hijo que lo reportó.
- La revisión de código de la entrega corre sobre el candidato aceptado, en el punto que el modo de entrega ya fija.
- Marcadores (hallazgo N3): una tarea `[x]` cuyos archivos cambia una corrección vuelve a `[?]` con una línea "reopened by correction round <n>", y la verifica `hive-verify-task` en el gate único. Una corrección fuera de las tareas del plan se anota en el avance y la cubre `hive-verify-change` en ese gate.
- TDD no cambia: una corrección que altera comportamiento del tipo que la regla lista lleva su test en la misma ronda; solo se difieren los hijos independientes.

`flow-build` "Work outside the plan's tasks" recibe una frase que remite a esa regla para las correcciones de validación.

### Tests acotados

`verification.md` "Frequency and coverage", primera viñeta, pasa a pedir:

- La selección de tests afectados que el proyecto ya ofrece (por ejemplo, un comando `related` o `--filter` que la guía del proyecto nombre), no la suite completa, mientras se itera.
- Ampliar a toda la suite del paquete cuando cambian el esquema, las migraciones, los archivos de setup o de configuración de tests, o un módulo compartido del que dependen muchos tests; ampliar a todo el repo cuando cambian el lockfile, la configuración del workspace o el CI.
- Una selección que no encuentra ningún test deja el cambio sin verificar, no aprobado.
- No repetir una suite completa sobre un candidato sin cambios; una repetición para detectar inestabilidad es un ejercicio que se declara antes.

`flow-build` paso 3 del ciclo de `hive-verify-task`: la re-verificación revisa solo los criterios que la ronda anterior marcó `not met`, y el brief los nombra. `content/agents/quality/hive-verify-task.md` (líneas 10 y 16, hallazgo B3) cambia en consecuencia: cuando el brief nombra criterios, el veredicto cubre solo esos; sin esa excepción el rol volvería a calificar todos.

### Atascos como hallazgos

`verification.md` agrega una viñeta en "Frequency and coverage": cuenta como hallazgo incidental, con su evidencia, cualquiera de estos:

- un error de deadlock o lock timeout en tests (por ejemplo, `40P01` de Postgres);
- una corrida que pasa 3 minutos sin salida nueva y se detiene;
- un test que falla y pasa al repetirlo sin cambios;
- una corrida local de tests de más de 5 minutos;
- una ronda de verificación (un hijo `hive-verify-task`, `hive-verify-change` o `hive-review-ux`) de más de 5 minutos;
- una espera del CI posterior al merge de más de 5 minutos.

Umbrales absolutos a propósito (hallazgo N2): el orquestador ve la duración de cada hijo y de cada comando, pero no siempre el tiempo de los tests dentro de un hijo. Para estos hallazgos, el escenario de falla que pide `global.md` línea 40 es el costo medido: qué se detuvo, cuánto tardó y cuántas veces.

Los de infraestructura de tests o CI del proyecto (deadlock, inestabilidad, suite lenta, CI lento) se proponen para el tracker del proyecto. Los del costo del propio flujo (rondas de verificación) se reportan como hallazgo de guía para la sesión de Hive, sin ticket en el proyecto, por la regla global de alcance entre proyectos; admiten "no volver a mencionar", no "abrir ticket".

`global.md` línea 40 agrega la tercera respuesta: "don't mention again" guarda el hallazgo como silenciado en la memoria persistente, bajo el proyecto y un tema estable; antes de proponer, la búsqueda existente descarta los silenciados de ese proyecto y tema. Se nombra la capacidad ("persistent memory", como ya dice esa línea), no el producto. Presupuesto de `global.md` (hallazgo N1): tope de +300 bytes; lo agregado (tercera respuesta, disparador de `flow-close`, concesión de ramas) se compensa condensando "Retention and cleanup" y las líneas 123 y 124.
