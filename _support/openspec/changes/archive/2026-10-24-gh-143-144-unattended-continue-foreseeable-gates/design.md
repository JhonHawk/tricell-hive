# Diseño

## Contexto verificado (development @ 6c0b5204, 2026-10-24)

- **Expiración de la delegación desatendida.** Se enuncia en dos lugares con redacción distinta:
  - `content/skills/unattended-delegation/SKILL.md:8`: «expires when that objective completes or the user writes again»;
  - `content/guidance/global.md:34`: «expiring when it completes or the user returns», y solo para la fuente «me voy».

  Ninguna prueba ni detector afirma ese texto. El sitio (`site/src/content/docs/index.mdx:90`) describe la activación, no la expiración. `site/src/content/docs/flows/build.md:39` dice que una pregunta de estado se responde en una línea y el trabajo sigue.
- **«Costly».**
  - Aparece como disparador en `global.md:68`, `flow-build/SKILL.md:18`, `flow-plan/SKILL.md:36` y `verification.md:3`, sin criterio en ninguno.
  - `verification.md:39` exige evaluar las suites largas «regardless of whether their command starts locally».
  - `flow-build/SKILL.md:34` dice que la delegación es trabajo interno y no necesita aprobación; `global.md` exime la revisión del plan por subagentes.
- **Momento de las preguntas.**
  - `questions.md:11`: «ask only at that decision point».
  - `reporting.md:98`: nunca meter commits, push, PR o despliegues en una opción de ruta; el flujo los pide en su pregunta de entrega.
  - `reporting.md:106`: «Ask once for a set of actions when one answer can authorize them all… ask separately only for an action whose risk or target needs its own decision, such as a production deployment».
  - `delivery-decisions.md:5`: la ronda de entrega y revisión ocurre antes de declarar listo un plan o antes de la primera edición de un trabajo sin plan.
  - `flow-plan/SKILL.md:70`: la pregunta de ruta es una decisión de ruta, no un permiso para proceder.
  - `verification.md:26` admite «explicit advance authorization» sin decir cuándo pedirla.
- **Tareas agregadas.**
  - `plan-revision.md:10` cubre «a changed task» y no menciona tareas nuevas.
  - `plan-execution.md:17` ata `hive-verify-task` al enfoque `tdd`/`check`.
  - `plan-execution.md:30` hace el commit de cada tarea cuando llega a `[x]`, y no antes.
  - `verification.md:43` dice que las ediciones posteriores a la revisión se informan y se pregunta si revisarlas.
  - En carvento, T19 tenía `Closes: AC18` y `Test approach: check`; las reglas existían y no se cumplieron. Es una corrección adoptada por decisión del usuario (E1-B).
- **Validación por pruebas.**
  - `tests/content/skill_links_test.go` solo valida los enlaces `skill:` de `global.md`.
  - Los enlaces relativos de skills y referencias los valida `tooling/management` dentro de `go test ./...`.
  - Ninguna prueba valida anclas de encabezado (`tooling/management/references.go:13`).
- **Tamaños.**
  - `global.md`: 35 036 bytes.
  - `flow-plan/SKILL.md`: 14 432 bytes (margen de 1 568).
  - `unattended-delegation/SKILL.md`: 4 410 bytes.
  - `flow-build/SKILL.md` (margen de 821) y `flow-close/SKILL.md` (margen de 333) no se tocan.

## Enfoque

Un solo hogar por regla, con punteros donde se aplica (AGENTS.md):

| Regla | Hogar canónico | Punteros o ajustes |
| --- | --- | --- |
| Expiración ante continuar, acusar recibo o preguntar por el estado | `unattended-delegation/SKILL.md:8` | `global.md:34` reemplaza su redacción divergente por el invariante mínimo completo, porque la viñeta debe actuar sin leer la skill |
| Prueba costosa | `global.md:68`, porque el disparador de gates vive ahí y debe actuar sin leer referencias | `verification.md:39` se condiciona al costo |
| Cuándo pedir las autorizaciones previsibles | `verification.md`, en «Gate selection and timing» (línea 26). Aporta solo el momento y la excepción a `questions.md:11` | `reporting.md:106` sigue siendo el hogar de agrupar en una respuesta y de que producción va aparte; `flow-plan/SKILL.md:36` remite a `verification.md` |
| Tareas agregadas tarde | `plan-revision.md`, viñeta nueva junto a la de coherencia | remite a `plan-execution.md` y a `verification.md:43` sin repetir sus pasos |

### Redacción propuesta

Va en inglés porque entra al código fuente. El implementador puede ajustar el estilo sin cambiar el contenido.

- **`unattended-delegation/SKILL.md:8`**, reemplaza «and expires when that objective completes or the user writes again.»:

  > and expires when that objective completes or the user sends a message that changes the objective, the scope or an answer in force, asks anything other than how the work is going, gives a new instruction, or tells you to stop. A message in any language that only tells you to continue, acknowledges progress, or asks how the work is going keeps the delegation for the same objective and expiry: answer a status question in one line, say that the delegation continues, and keep deciding.

- **`global.md:34`**, reemplaza «expiring when it completes or the user returns.»:

  > expiring when it completes or the user writes anything other than telling you to continue, acknowledging progress, or asking how the work is going.

- **`global.md:68`**, reemplaza «or costly load/hardware exercises,» por:

  > or a costly check, meaning one that spends money or a third-party service's quota, uses or changes shared or remote resources, or is expected to run longer than 30 minutes (this session's own model use, its children, and the project's CI do not count),

  y añade después de la frase que la contiene:

  > A local check of minutes with no external cost is routine.

- **`verification.md:26`**, añade después de «its mere inclusion of a check is insufficient.»:

  > Ask for the authorizations the plan foresees for later steps, such as a costly check, publishing an artifact, or preparing a shared environment, while the user is present, in the round that settles the delivery and review questions, before the plan is declared ready or before the first edit of work without a plan; when the repository's settings answer those questions, ask them alone at that same point. Give each its own question ID with its scope, target and limits. This is an exception to asking only at the decision point in [questions](../../flow-plan/references/questions.md): it covers only what the plan, or for work without a plan the request, already names, and decisions they do not foresee are still asked when they arise. Group them and keep a production deployment apart as the "Ask once for a set of actions" rule in [reporting](../../flow-close/references/reporting.md) sets. Under unattended delegation, decide them under its Effects section instead of asking.

- **`verification.md:39`**, reemplaza «need scope and effect assessment, regardless of whether their command starts locally.»:

  > need scope and effect assessment when they are costly as the global guidance defines it, regardless of whether their command starts locally.

- **`flow-plan/SKILL.md:36`**, reemplaza «record whether execution authorization is pending or already sufficient under the shared rule.»:

  > record whether execution authorization is pending or already sufficient under the shared rule, and ask the pending ones in the delivery and review round, as [verification](../flow-build/references/verification.md#gate-selection-and-timing) sets.

- **`plan-revision.md`**, viñeta nueva después de la de coherencia:

  > - A task added during the build or after the user's validation follows the same plan rules as any other: criteria falsifiable at the base commit, a `Closes:` line and a test approach. A `tdd` or `check` task is verified by `hive-verify-task` before its commit, as [plan execution](../../flow-build/references/plan-execution.md) sets, and the completion report never claims independent verification for a task that did not get it. When it lands after the dedicated code review ran, report it and ask whether to review it, as [verification](../../flow-build/references/verification.md) sets for later edits.

Comprobación manual: el encabezado `## Gate selection and timing` existe en `verification.md`, de modo que el ancla `#gate-selection-and-timing` resuelve. El puntero a `reporting.md` cita la frase «Ask once for a set of actions» en lugar de un ancla, porque esa frase está al final de la sección «Closing categories», después de «Reminders», y no bajo un encabezado propio. Ninguna prueba valida anclas.

### Alternativas descartadas

- **Pedirlas en la pregunta de ruta del plan.** `reporting.md:98` prohíbe meter efectos en una opción de ruta, y `flow-plan/SKILL.md:70` dice que esa pregunta no es un permiso.
- **Mover la regla a `questions.md`.** Es la referencia de cómo preguntar; `verification.md` ya contiene «advance authorization».
- **Ampliar `flow-close/SKILL.md:48`.** Está a 333 bytes del límite y su regla ya es coherente.
- **Aplicar la frase D1-B de #45.** No habría atrapado T19, choca con `verification.md:43` y requiere cambiar el contrato de `hive-verify-task`.
