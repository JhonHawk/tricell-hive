# Diseño — el plan ejecuta los comandos de verificación que escribe

## Comparación de tres versiones

| Versión | Qué pide al plan |
| --- | --- |
| Hive antes del rework (`9f8db1df`, 2026-11-01) | «Read the code, the `## Hive` settings and the existing specs the change touches.» Nada sobre ejecutar. |
| Hive hoy (`04a1f27b`) | Lo mismo, más: la pista del ticket es una afirmación a verificar; los datos externos se miden con una lectura antes de escribir las tareas (`hive-rule-external-facts`). Los comandos de `Verification:` no se ejecutan; tras la revisión, «apply what holds», sin pedir que la corrección se compruebe. |
| pstack 0.15.46 (`poteto-mode`) | `playbooks/multi-phase-plan.md` paso 2: resolver por prototipo las dudas abiertas antes de escribir el plan; preguntar solo lo que ninguna ejecución resuelve. `SKILL.md`: «A fork you could settle by running something → prototype it». *Attack the Premise*: dos correcciones que comparten una premisa fallaron la misma puerta; probar la premisa directamente antes de otra corrección. |

En contra de la regla actual: la sesión `01a1fb69` mostró que el plan, tal como está, convierte al revisor en el depurador de sus comandos. Dos de tres rondas encontraron un P1 que la corrección anterior había introducido. Las dos correcciones compartían la premisa «la prueba de la app corre así», como describe *Attack the Premise*.

En contra del cambio: correr la suite completa en cada plan cuesta tiempo (4 min aquí). Por eso, según D3 A, un comando cuya última corrida medida pasó de 5 minutos se ejecuta solo en su forma más corta con el mismo runner; sin medición previa, se corta a esa forma a los 5 minutos.

## Enfoques (D1)

- **A (elegido).** Extender lo existente. La regla vive en `flow-plan` paso 3. Lo apuntan «Revise a plan in progress», el rol `hive-review-plan`, la plantilla y el punto 7 de `plan-handoff`. Registra en la misma sección «Verified context» que ya usan los datos externos. Es el peldaño 2 de *Reuse before writing*: reusa el formato de línea y el lugar existentes.
- **B.** A más un chequeo en `check_change.py`. Descartado: solo prueba que la línea existe, no que el comando se ejecutó, y los recorridos en pantalla necesitan excepción.
- **C.** Fusionar con `hive-rule-external-facts` en una hoja de «datos del plan». Descartado: reescribe una hoja vigente con sus pruebas de enlaces y de lugar único.

## Texto de la regla (borrador, inglés)

**Nota del build (2026-11-21):** el texto final está en `content/skills/flow-plan/SKILL.md` paso 3, que es su hogar. Es más corto que este borrador por dos límites que miden las pruebas:
- `reading_footprint_test.go` exige que lo que se lee entre plan y build quede por debajo de la línea base; el borrador la pasaba por 1123 B.
- `leaf_skill_test.go` limita la hoja `hive-rule-change-folder` a 4000 B, y la hoja ya tenía 3999.

Por eso la regla de volver a ejecutar en una revisión vive en `flow-plan` paso 3, y «Revise» paso 1 solo la apunta. Del paso 3 de «Revise» se quitó la frase que describía qué revisa `check_change.py`; la misma descripción sigue en el spec `workflow-lifecycle` y en el sitio, y el script imprime cada línea que falla.

El borrador de abajo queda reemplazado por el texto de `content/`. Quedan fuera del texto final: «passed to the runner by path», «locally or in CI» para la última corrida medida, el registro de los nombres de las pruebas que fallan, y «or in the summary of a plan kept in the conversation» en la frase de registro, porque «Where the plan lives» ya pone cada comando en ese resumen. El margen sobre la línea base de `reading_footprint_test.go` quedó en 12 B.

En `flow-plan/SKILL.md`, al final del paso 3:

> Before writing a task's `Verification:` line, run each command it names once on the base, so the plan never prescribes a check that cannot run; an observation or a walk on that line is not a command and is left to verify. When the command needs the change to exist, such as a test file the task adds, run its nearest existing form: the same script or runner on a test it already finds, and, when the new test will import project code, a throwaway test that only imports those modules, written in the host's temporary location and passed to the runner by path; when the runner only finds tests inside the repository, write it there untracked, run it and delete it in the same command. Do not run while planning a command that changes tracked files, a database that is not local, shared or remote resources or an external system, nor a long check as [checks](../hive-rule-checks/SKILL.md#long-checks) sets; the runner's own caches and ignored output, and a local test database already running, do not count. Write such a command as `not measured`, with the task that meets it first, and likewise a command that needs a service the build will start, such as a local database not yet running, with the task that starts it. Record one line per command in the plan's verified context, or in the summary of a plan kept in the conversation, in the format external facts use: the command, its result and the date. A command that cannot run because its script, runner or a dependency is missing, or because a module does not load, is a gap of the plan: change the task's check before the review. One that runs and fails on an existing test is recorded with the failing tests' names.

Frase canónica para `single_home_test.go`: `run each command it names once on the base`.

Umbral (D3 A), una oración más en el mismo párrafo, después de «nor a long check…»:

> A command whose last measured run, locally or in CI, took over 5 minutes runs only in its shortest form with the same runner, such as one test file; with no measured run, it starts in the background and is cut to that form once it passes 5 minutes, in place of the 30-minute question of [checks](../hive-rule-checks/SKILL.md#long-checks) while planning. Start together, each in the background, the commands that share no database, port, lock or output directory, and wait for all of them; run the others one after another.

En `flow-plan` «Where the plan lives», la lista de lo que lleva el resumen de un plan en la conversación suma: «each verification command with its result».

En `flow-plan/SKILL.md` línea 13, los pasos 4 y 5 quedan en líneas separadas (hoy «…sets.5. When…»).

En `hive-rule-change-folder` «Revise a plan in progress», al final del paso 3:

> When the revision changes a task's `Verification:` command or what it rests on, such as its runner, a dependency or a module the test imports, run it again as [flow-plan](../flow-plan/SKILL.md) step 3 sets, against the working tree during a build, and update its line before the next review round, when there is one.

En `hive-review-plan.md` paso 1, una oración nueva:

> Check that each command a task's `Verification:` line names has its line in the plan's verified context, as [flow-plan](skill:flow-plan/SKILL.md) step 3 sets, without running it: a command with no line, or one recorded as run although the project lacks what it needs, such as its script, runner or dependency, is a gap. An observation or a walk on that line is not a command.

En `plan-handoff.md` punto 7: «…and the code review, with each task's verification command written as not measured.»

En la plantilla `proposal.md`, bajo «Verified context», un comentario aparte y sin condición: «One line per command a task's `Verification:` line names, as flow-plan step 3 sets: the command, its result or not measured, and the date.»

## Lugares revisados (para el texto del pull request)

`content/guidance/global.md` (no cambia; 28 593 bytes), `flow-plan/SKILL.md`, `plan-handoff.md`, plantillas `proposal.md` y `tasks.md`, `hive-rule-change-folder`, `hive-rule-external-facts`, `hive-rule-checks`, `hive-rule-verification-floor`, `hive-rule-tests`, `project-layout/references/layout.md`, `content/agents/review/hive-review-plan.md`, `site/src/content/docs/flows/plan.md`, `_support/docs/architecture/` (no menciona «verified context»), `tests/pilot/` (sin detector afectado), `tests/content/single_home_test.go`, `README.md` y `llms.txt` (no cambian), `CHANGELOG.md`.
