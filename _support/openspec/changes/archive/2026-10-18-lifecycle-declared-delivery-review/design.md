# Diseño

## Contexto verificado

Base: `development` en `e484b47b` (2026-10-17). Fuente externa: [How the Claude Code team does code review](https://claude.dev/blog/solving-code-review-at-anthropic/), leída el 2026-10-17 por resumen web (no verbatim). Registro de la investigación: Engram `research/anthropic-code-review-lifecycle`.

**Mapa por capa** (hechos inspeccionados en el repo):

| Capa del artículo | Hive hoy | Qué cambia |
| --- | --- | --- |
| Revisión fresca pre-PR, rutinaria | `hive-review-code` solo si la pregunta de revisión lo eligió (`delivery-decisions.md:26-34`) | R2: con `Review` declarado corre sin preguntar |
| Aprobación por nivel de riesgo | Modo recomendado por riesgo pero preguntado cada sesión (`delivery-decisions.md:5-10`); solo `Delivery: direct-base` evita la pregunta y solo en repos de soporte (`:18`) | R1: `Delivery: automatic|interactive` declarado es el modo |
| Cola de merge | No existe; CI post-merge se espera después (`verification.md:88`); el merge exige "CI is green" sin decir sobre qué base (`delivery-decisions.md:24`, `git-workflow:38`) | R3: base actual antes del merge |
| Revisor flagea la clase, no cada instancia | `hive-review-code.md:16` da prioridad por hallazgo, sin regla de agrupación | R4 |
| Revisión de seguridad separada | `hive-review-security` solo si el implementador juzga que cruza una frontera (`verification.md:43`) | R6: el revisor nombra la frontera |
| Incidente → qué capa debió atraparlo | Postmortem con "acción preventiva con dueño" (`delivery-decisions.md:44`) | R5 |

**Validación del manager:** `tooling/cli/doctor_project.go:49` fija `hiveDeliveryWant = "direct-base"`; `:354-357` rechaza otro valor (`expected "direct-base", found …`); `tui_project_view.go:313-322` construye el selector de `Delivery` con `{"", direct-base}` y conserva un valor actual distinto como tercera opción. Pruebas: `doctor_project_test.go:271-277`, `tui_project_view_test.go:353-362,377-379,405-407,422`, `project_command_test.go:171`. `hive project --set` y el paso Apply del formulario validan con el mismo `checkHiveText` (`project_write.go:525`; pruebas `project_write_test.go:129-131,183`, que solo comprueban el prefijo `Delivery: expected`). `deployment-manager.md:77` documenta la validación de `doctor` y `:110` el rechazo al escribir. El comentario de `tui_project_view.go:272-275` dice "their one valid value".

**Conflicto a resolver en `global.md:66`:** "a saved preference, listed tool, or planned gate alone is not authorization". Con R2, un `Review` declarado sí autoriza; como la regla de compuertas actúa sin que se lea `delivery-decisions.md`, `global.md` enuncia esa excepción completa (AGENTS.md, "Guidance structure").

**Sitio:** `flows/build.md:41` ("the agent asks once, before the first edit, how you want it delivered and reviewed") y `:64` ("You decide. Say so in your request, or answer the delivery question") describen la pregunta; `:67` menciona el postmortem sin listar su contenido, así que R5 no lo toca.

**Detectores:** `tests/pilot/` no detecta la pregunta de entrega ni de revisión (`rg -l -i "delivery question|review question|direct-base|delivery mode" tests` → solo `tests/fixtures/flows/README.md`).

## Diseño propuesto

Principio: cada regla en un solo hogar; los demás sitios son punteros. Hogares: R1 y R2 en `delivery-decisions.md`; R3 en `git-workflow/SKILL.md`; R4 y R6 en `hive-review-code.md`; R5 en `delivery-decisions.md § Urgent production fixes`. `global.md` repite solo la invariante de autorización (R2) porque la regla de compuertas actúa sin leer su hogar.

Decisión de forma para R1: no se añade una lista de rutas que siempre exigen persona. La regla existente que marca `automatic` no disponible cuando la guía exige validación humana (`delivery-decisions.md:10`) ya cubre el caso del artículo ("some paths always need a person") para la superficie que importa a los proyectos de Hive, la UI.

### T1 — Textos de R1 y R2

`delivery-decisions.md:5`: reemplazar "a repository default or remembered preference only informs the recommendation." por "a remembered preference, or a default stated outside the `## Hive` section, only informs the recommendation." e insertar a continuación:

> A `Delivery: automatic` or `Delivery: interactive` value declared in the repository's `## Hive` section is the mode for work there and authorizes that mode's effects, in the same way as an earlier answer: do not ask the delivery-mode question. An explicit choice made earlier in the session for that repository, or a narrower instruction in the request, such as `hold` or commit only, overrides the declared value. Before the first edit, state in one line: the mode and where the work stops; the review mechanism that will apply; any deployment the destination triggers; and the change record's closing route, which, when the record lives in this repository or in a specs repository without `Delivery: direct-base` and the declared mode applies, is a small closing pull request without dedicated review that you merge. The user can change any of them in reply. The rule below that marks `automatic` unavailable for a change that needs human validation still applies: the work then runs under `interactive`, and the line names that rule and the stop.

En la misma línea, reemplazar "Otherwise ask once, in the session language" por "Otherwise, when no `Delivery` value is declared, ask once, in the session language".

`delivery-decisions.md:18`, reemplazar "Every other repository in the work is a code repository and goes through the question." por "Every other repository in the work is a code repository and goes through each question whose value it does not declare, as the paragraphs above and below set."

`delivery-decisions.md:20`, reemplazar "When the work spans several code repositories, ask one delivery-mode question and one review question for all of them: each option lists every code repository with its base branch, and one answer covers them all." por "When the work spans several code repositories, ask one delivery-mode question for those without a declared `Delivery` value and one review question for those without a declared `Review` value, and name the declared ones with their values in the one-line statement: each option lists every code repository the question covers with its base branch, and one answer covers them all."

`delivery-decisions.md:22`, reemplazar "the same question covers versioning that record once, at close" por "the same question, or the declared value's route, covers versioning that record once, at close".

`delivery-decisions.md:26`, reemplazar "Resolve the implementation code-review mechanism and boundary in the same exchange as the delivery mode, for every mode including hold, so delivery does not ask again. Build the question in the session language:" por:

> Resolve the implementation code-review mechanism and boundary in the same exchange as the delivery mode, for every mode including hold, so delivery does not ask again. A `Review` value declared in the repository's `## Hive` section is the answer and authorizes that mechanism: run it as a routine check of the candidate, at the point the delivery mode sets, without asking, or run no dedicated code review when the value is `none`; name it in the one-line statement above. Confirm without launching it, as the first bullet below sets, that the named mechanism exists on this host; when it does not, or it names another host's command, ask the question below and state that limitation. A declared `Review` value, including `none`, also authorizes `hive-review-security` when the change crosses a boundary in [security boundaries](../../flow-build/references/security-boundaries.md): `none` declines the dedicated code review only, and the line names the security review when it applies. Without a declared value, build the question in the session language:

(La frase sobre fronteras de seguridad es la opción N4-A, confirmada por el usuario el 2026-10-17.)

`verification.md:43`, además del cambio de T4, reemplazar "and within the agreed review authorization" por "and within the agreed review authorization or the authorization a declared `Review` value gives". (Como T4 edita la misma línea, T1 hace este reemplazo y T4 el suyo, después.)

`flow-plan/SKILL.md:18`, reemplazar "to resolve the Git mode and applicable code-review mechanism, asked in the same round as the last design questions that do not depend on them" por "to resolve the Git mode and applicable code-review mechanism: stated in one line when the repository declares them, otherwise asked in the same round as the last design questions that do not depend on them".

`flow-close/SKILL.md:67`, reemplazar "Act only on that recorded answer; repository policy does not replace it." por "Act only on the recorded answer, or on the route a declared `Delivery` value sets, with you as the merger, as [delivery decisions](../flow-plan/references/delivery-decisions.md) sets; repository policy does not replace either." y "If no closing route was recorded, ask for it before publishing." por "If no closing route was recorded and none follows from a declared `Delivery` value, ask for it before publishing."

`delivery-decisions.md:34`, reemplazar "A mere mention is not a preference; a `Review` setting in the repository's `## Hive` section is an explicit preference." por:

> A mere mention is not a preference. A `Review` value in the `## Hive` section is not an input to this question: it is the declared answer handled above.

`project-settings.md:9`, reemplazar "Optional: `Environments`, `Review`, `Delivery: direct-base` (a support repository, such as specs or mocks; see [delivery decisions](delivery-decisions.md)), and `Hive guidance: required`" por:

> Optional: `Environments`; `Review`, the code-review mechanism for work there, or `none` for no dedicated review; `Delivery`, either `direct-base` for a support repository such as specs or mocks, or `automatic` or `interactive` as the delivery mode for a code repository, as [delivery decisions](delivery-decisions.md) sets; and `Hive guidance: required`

Y en el ejemplo, después de `- Review: /code-review`, añadir la línea `- Delivery: automatic`.

`global.md:66`, reemplazar "Offer mechanisms available to the project and current host; a saved preference, listed tool, or planned gate alone is not authorization." por:

> Offer mechanisms available to the project and current host; a saved preference, listed tool, or planned gate alone is not authorization; a `Review` or `Delivery` value declared in the repository's `## Hive` section is authorization for the mechanism or mode it names, as the delivery decisions reference sets.

(`global.md` conserva solo esta invariante porque la regla de compuertas actúa sin leer `delivery-decisions.md`; el "authorizes" completo vive allí, como pide `global.md:28`, "a rule that authorizes a write says so".)

`flow-build/SKILL.md:20`, reemplazar "When the work will end in a Git delivery and neither a plan, an earlier answer, nor the user's request, such as a promotion to a named environment, settled it, read [delivery decisions](../flow-plan/references/delivery-decisions.md) and ask its delivery-mode and review question once, before the first edit." por:

> When the work will end in a Git delivery and neither a plan, an earlier answer, the user's request, such as a promotion to a named environment, nor the repository's declared `Delivery` and `Review` settings settled it, read [delivery decisions](../flow-plan/references/delivery-decisions.md) and ask its delivery-mode and review question once, before the first edit; a declared setting is stated in one line, not asked, as that reference sets.

Y en la oración siguiente, reemplazar "unless a plan, an earlier answer, or the user's request already settled it" por "unless a plan, an earlier answer, the user's request, or a declared `Review` setting already settled it".

`verification.md:33`, reemplazar "Run the dedicated code review the review question settled; propose it here only when no review question was asked." por "Run the dedicated code review that the review question or the declared `Review` setting settled; propose it here only when neither settled it."

### T2 — Manager Go

`doctor_project.go`: reemplazar la constante `hiveDeliveryWant` por `var hiveDeliveryValues = []string{"direct-base", "automatic", "interactive"}` (añadir el import `slices`) y en el `case "Delivery"` rechazar solo `present && !slices.Contains(hiveDeliveryValues, v)` con el mensaje `expected one of "direct-base", "automatic", "interactive", found %q`, construyendo la lista desde el slice para que mensaje y valores no se separen. `tui_project_view.go:313-322`: para `Delivery`, `f.options[i] = append([]string{""}, hiveDeliveryValues...)`; la rama que conserva un valor actual desconocido pasa a `present && cur != "" && !slices.Contains(f.options[i], cur)`, para que un archivo con `automatic` no lo muestre dos veces; `Hive guidance` sigue con `{"", hiveGuidanceWant}`; el comentario de `:272-275` deja de decir "their one valid value". Actualizar las pruebas que esperan `expected "direct-base"` (`doctor_project_test.go:273`) y el ciclo del selector (`none → direct-base → automatic → interactive → none`, y con flecha izquierda `none → interactive`). Pruebas nuevas: `TestDoctorProjectAcceptsDeliveryModes`, con `Delivery: automatic` e `interactive` el proyecto es `Valid` y con `Delivery: pr` el mensaje lista los tres valores; un caso de `checkHiveText` que acepta `automatic`; y un caso de carga del formulario con `Delivery: automatic` que muestra la opción una sola vez. `deployment-manager.md:77`: "and `Delivery` or `Hive guidance` values other than `direct-base` and `required`" pasa a "a `Delivery` value other than `direct-base`, `automatic` or `interactive`, or a `Hive guidance` value other than `required`"; `:110`: "a `Delivery` other than `direct-base`" pasa a "a `Delivery` other than `direct-base`, `automatic` or `interactive`".

### T3 — Texto de R3

`git-workflow/SKILL.md`, en "Verify the authorized effect", nuevo párrafo antes de "Stop at the authorization boundary.":

> Before merging a pull request, confirm that the checks that passed ran against the current base. Check it without changing the branch: run `git fetch` and then `git merge-base --is-ancestor origin/<base> origin/<branch>`, which exits 0 only when the pull request's pushed head already contains the base; `BEHIND` in `gh pr view <number> --json mergeStateStatus` also shows the base moved, but another status does not prove it current. When that check fails, bring the branch up to date with a merge from the base and push it, which the merge authorization covers, wait for the new run, and merge only when it passes; when the base moves again while you wait, repeat once, and after that report the merge as pending with the reason. When the repository requires a rebase instead, the rewrite and its force-push need the explicit authorization that the history rule above sets: ask for it, and when no one can answer, leave the merge pending and report it. When the repository has a merge queue, add the pull request to the queue and wait for its result instead, since the queue makes this check. A green run on a stale base does not satisfy the merge condition.

El campo `mergeStateStatus` existe en `gh` 2.102.0 (comprobado con `gh pr view --json` sin campos). No se verificó si GitHub reporta `BEHIND` sin la protección "require branches to be up to date"; por eso la regla es `git merge-base --is-ancestor` y `BEHIND` solo un indicio positivo.

`delivery-decisions.md:24`, reemplazar "once CI is green, conflicts are resolved" por "once CI is green on the current base, as [git-workflow](../../git-workflow/SKILL.md) sets, conflicts are resolved".

### T4 — Textos de R4 y R6

`hive-review-code.md:16`, al final del paso 4 añadir: "When the same defect appears in several places, report it once as one finding that names the defect class and every location, not one finding per instance."

`hive-review-code.md:20`, reemplazar "End with the number of findings at each priority and a one-line verdict: whether the change is correct as written, with its main reason." por: "End with the number of findings at each priority and a one-line verdict: whether the change is correct as written, with its main reason. Then say whether the diff crosses a boundary listed in [security boundaries](skill:flow-build/references/security-boundaries.md), naming each one it crosses, so the parent can decide on a security review."

`verification.md:43`, reemplazar "add `hive-review-security` only when the change crosses a boundary in [security boundaries](security-boundaries.md) and within the agreed review authorization" por "add `hive-review-security` only when the change crosses a boundary in [security boundaries](security-boundaries.md), including one the code reviewer's verdict names, and within the agreed review authorization".

### T5 — Texto de R5

`delivery-decisions.md:44`, reemplazar "offer a short postmortem: what failed, the impact window, the cause, the fix, and a preventive action with its owner." por "offer a short postmortem: what failed, the impact window, the cause, the fix, which layer should have caught it (a check, a test, a standard, a review criterion, or a deployment check), and a preventive action in that layer with its owner."

### T6 — Sitio

`flows/build.md:41`, añadir tras la primera oración: "When the repository's `## Hive` section declares `Delivery` and `Review`, the agent states them in one line before the first edit instead of asking; a narrower instruction in your request still wins."

`flows/build.md:64`, reemplazar "You decide. Say so in your request, or answer the delivery question the agent asks." por "You decide. Say so in your request, declare it once in the repository's `## Hive` section, or answer the delivery question the agent asks." Y añadir al final del párrafo: "Before merging, the agent makes sure the checks ran on the current base branch, updating the branch and waiting for them again when the base moved, or using the repository's merge queue when it has one."
