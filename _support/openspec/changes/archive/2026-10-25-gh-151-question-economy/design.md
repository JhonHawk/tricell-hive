# Diseño

## Contexto verificado

El análisis de sesiones se hizo el 2026-10-25 con cuatro hijos `hive-research`, y las frases decisivas se comprobaron en las transcripciones. Lo que encontró:

- **Pi carvento `01a17176`.** El agente preguntó D1–D5 en 53 min para un solo conjunto de datos de prueba en QA. El texto dice «conviene autorizarla por etapas» y «antes de escribir, confirmaré registros exactos, cambios y reversión». Esta segunda frase es la regla de `global.md:28`.
- **Codex cam `01a1679c`.** D42–D44 (imágenes, copia al VPS, QA) eran pasos del plan aprobado en D38. Los merges de cada PR de corrección (D33, D34, D45, D47, D48 y D54) vinieron de la condición del usuario «el merge se mantiene sin aprobación hasta que regrese», que el agente siguió aplicando después de esa vuelta.
- **Claude cam `6b6eb569` (COM-878).** Hubo 13 preguntas en 43 min. D59, D60, D66 y D67 eran decisiones menores, y D65–D67 salieron de la revisión del plan. En D57 se preguntó «¿con qué seguimos?» cuando la memoria ya tenía D56-A. Los datos locales se crearon sin preguntar, como dice `verification.md:51` y `:55`.
- **Claude facti `c574e5a1` y `42b8e06a`.** El mismo acceso a producción se pidió cuatro veces para el security group y dos para Atlas.
- **Tickets en cam Codex.** D26 y D50 fueron la pregunta de ticket que `flow-research/SKILL.md:38` exige en cada ronda. D36 fue un hallazgo incidental a mitad del build, un solo caso.

Reglas implicadas, leídas en `2bdf9357`:
- `global.md:28`, que confirma registros exactos para datos compartidos o de producción sin definir «compartido», y `global.md:68`, que define la prueba costosa.
- `global.md:76`, que pide el ticket y la ruta en una ronda.
- `verification.md:26`, que adelanta solo tres casos y dice «decisions they do not foresee are still asked when they arise».
- `delivery-decisions.md:5` y `:24`.
- `flow-build/SKILL.md:20`, que en trabajo sin entrega Git omite las autorizaciones previstas, al contrario de `delivery-decisions.md:5`.
- `questions.md:11`, `:19` y `:23`.
- `plan-review.md:37`.
- `flow-research/SKILL.md:38` y `:42`.
- `reporting.md:98`, que prohíbe efectos en una opción de ruta, y `reporting.md:106`.

Fuera de `content/`, el comportamiento aparece en `site/src/content/docs/flows/{research,plan,build,close}.md` y en `CHANGELOG.md`. Ninguna prueba compara la redacción de las reglas. `tests/content/reading_footprint_test.go` solo mide el tamaño de lo que se lee. `_support/docs/architecture/instruction-resources.md` lista los hogares canónicos.

## Enfoque

1. **Un hogar nuevo para autorizar efectos:** `flow-plan/references/effect-authorization.md`, que dice cómo se clasifican, qué pregunta lleva cada clase y qué cubre una respuesta. `global.md:28` conserva el invariante completo en una frase, porque estas autorizaciones también aparecen fuera de los flujos, por ejemplo en la investigación de facti. `verification.md:26` sigue siendo el hogar del *momento*, la ronda de entrega y revisión, y remite a la referencia para el *qué*.
2. **Agrupar y reutilizar** son reglas generales de cualquier pregunta y van en `questions.md`.
3. **Decisiones menores de la revisión del plan:** van en `plan-review.md`, que es donde se concilian los hallazgos, y el resumen del plan las muestra.
4. **Vigencia de las condiciones y PRs de corrección:** van en `delivery-decisions.md:24`, junto a la regla de merge.
5. **Ticket dentro de la ruta:** `flow-research/SKILL.md:38` es su hogar. `global.md:76` y `reporting.md:98` se alinean con él.

Alternativa descartada: ampliar `verification.md:26`. Esa referencia se lee solo para integración, runtime, UI y gates, y los accesos de facti ocurrieron durante una investigación.

**Revisión 2026-10-25, a pedido del usuario.** Los entornos varían por proyecto: los de Tricell y sus clientes son `local → dev → qa → prod`, y facti es `local → gamma → prod`. Por eso las clases se toman de `Environments` y no de los nombres dev o QA. Además, los permisos que desbloquean la validación se piden al planear, para cada entorno que alcanza el alcance. Cuando el trabajo pasa a un entorno posterior que el ticket no cubría, como la validación en QA de carvento cuando el ticket solo llegaba a development, se piden todos juntos al empezar esa etapa.

## Redacción propuesta

Si el implementador ajusta el estilo de una frase, actualiza la frase de comprobación en [tasks.md](tasks.md) antes de verificar.

### Nueva `content/skills/flow-plan/references/effect-authorization.md`

```markdown
# Effect authorization

## When to read

Read before asking the user to authorize an effect beyond this machine, and before writing the foreseen authorizations of a plan or of the first build question. Reading it authorizes no effect and does not invoke the planning workflow. Format each question as [questions](questions.md) sets; when to ask is set in [verification](../../flow-build/references/verification.md#gate-selection-and-timing).

## Classify each effect

Classify every effect the work and its verification need by where it lands and whether it can be undone. When an effect fits several classes, the strictest one applies. Classify environments as [project settings](project-settings.md) defines the `Environments` setting. Without that setting, an environment is non-production only when its name or configuration shows it, such as dev, qa, staging or gamma; otherwise treat it as production. Reading data is not an effect; opening access in order to read, such as allowing an address through a firewall, is an effect of the class where that access lands.

| Class | Covers | What it takes |
| --- | --- | --- |
| Local | This machine: the working tree, local services, local databases, accounts and seed data | No question; it belongs to authorized implementation and verification |
| Non-production | Records or resources with identifiable names that the task creates in a non-production environment and removes at its end, such as a test partner, branch office or listing; existing accounts or records there used only as inputs, such as signing in with them or attaching new records to them; and reversible configuration there | One question for the whole set |
| Production | Reversible effects in production, such as temporary access, a service reload, or a configuration change with a known revert | One question per kind of effect for the task, such as access, reload or configuration |
| Critical | Any environment: an irreversible or hard-to-reverse effect, such as deleting data or resources without a backup, sending messages to real people, publishing to third parties, or rotating credentials; spending real money or paid quota; destroying or replacing infrastructure, such as a destroy or replace, deleting an environment, or changing DNS; and modifying or deleting data the task did not create in any non-local database, including migrations and backfills | Its own question for each effect |

## Ask once for each set

- **Non-production:** list the whole set the task or its verification needs in that environment: each record or resource with an identifiable name, what uses it, and how it is removed at the end. Offer, as options: create the set; first look for existing accounts or records that could serve as inputs, saying that the search takes time and that it comes back as a new question only when it changes the set; and skip what depends on the set, naming what stays unverified. Do not split the set into stages.
- **Production:** for each kind of effect, name every target, the change, how to revert it and how long it lasts.
- **Critical:** ask each in its own question, never folded into another question's option or into a set, naming the exact target, the estimated cost and a ceiling when it spends, and how to revert it, or saying it cannot be reverted. It holds even when the user's request names the effect.

When declining makes the authorized task or one of its criteria impossible, say so in the question.

## What an answer covers

An answer authorizes the effects it lists, of that kind, on those targets, for the current task in this session. A later effect of the same kind in the same environment and within the listed set, or on a listed production target, needs no new question. A different environment, a target outside the set, a critical effect not already authorized for that exact target, or going past a stated ceiling needs a new one. An authorization recorded in memory or in a plan from an earlier session does not cover a new session, unless the plan records it as advance authorization for that task.

## Under unattended delegation

The Effects section of [unattended delegation](../../unattended-delegation/SKILL.md) sets how critical effects and the other classes are handled instead of asking.
```

### `content/guidance/global.md:28`

Reemplazar «Before a write to shared or production data, confirm the exact records, the change, and how to revert it, even when the user's question names that write; under unattended delegation, record them instead, as its skill says.» por:

> Before asking to authorize an effect beyond this machine, read [effect authorization](skill:flow-plan/references/effect-authorization.md): local effects take no question; the test data a non-production environment needs, meaning one before the last of the repository's `Environments`, takes one question for its whole set; reversible production effects take one question per kind of effect and task, with each target, change and revert; and each critical effect takes its own question, even when the user's request names it. A critical effect is an irreversible one, spending real money or paid quota, destroying or replacing infrastructure, or changing or deleting data the task did not create in a non-local database. Under unattended delegation, skip a critical effect not authorized in advance for that exact target and ceiling and report it, and record the target, change and revert of the others, as that reference and its skill say.

### `content/guidance/global.md:76`

Reemplazar «end with the ticket question and the route question in one round» por «end with the route question, whose options say whether they open a ticket».

### `content/skills/flow-build/references/verification.md:26`

- Reemplazar «such as a costly check, publishing an artifact, or preparing a shared environment, while the user is present» por «such as a costly check, publishing an artifact, or every effect beyond this machine that the work and its verification will need in each environment its scope reaches, classified as [effect authorization](../../flow-plan/references/effect-authorization.md) sets, while the user is present».
- Después de la frase de lo no previsto, añadir: «When the work reaches an environment its plan did not cover, such as a validation requested in a later environment after the ticket's scope ended in an earlier one, including through a request to promote or deploy there, which covers only its delivery chain, ask when the work in that environment starts, before its first effect, for every effect its verification there will need, in one round.»

### `content/skills/flow-plan/references/delivery-decisions.md:5`

Reemplazar «it excludes the functional walk, which runs only as [verification](../../flow-build/references/verification.md) allows, so do not ask the delivery mode» por «it excludes the functional walk, which runs only as [verification](../../flow-build/references/verification.md) allows, and the effects verification needs in that environment, which that reference asks when the work there starts, so do not ask the delivery mode».

### `content/skills/flow-plan/references/project-settings.md:9`

Reemplazar «Optional: `Environments`;» por «Optional: `Environments`, the ordered environments from the base branch to production, such as `development → qa → production`: the last one is production and every earlier one is non-production, and `local` is never one of them even when listed;».
- Reemplazar «and group them as the "Ask once for a set of actions" rule in [reporting](../../flow-close/references/reporting.md) sets, which keeps a production deployment apart» por «and group effects as [effect authorization](../../flow-plan/references/effect-authorization.md) sets and the other actions as the "Ask once for a set of actions" rule in [reporting](../../flow-close/references/reporting.md) sets, which keeps a production deployment apart».
- Reemplazar «it covers only what the plan, or for work without a plan the request, already names, and decisions they do not foresee are still asked when they arise» por «it covers what the plan, or for work without a plan the request, already names, and what those answers cover as [effect authorization](../../flow-plan/references/effect-authorization.md#what-an-answer-covers) sets; an effect neither covers is asked when it arises, together with every other effect that same step needs. When a gate this paragraph confirms, such as a deployed-environment walk, needs effects beyond this machine, ask for them in that gate's question».
- `global.md:68` no cambia. Una prueba costosa que cambia recursos compartidos sigue pidiendo confirmación, y sus efectos van en esa misma pregunta, según la frase anterior. Las clases dicen qué se lista. La prueba costosa dice cuándo hace falta el gate.

### `content/skills/flow-build/SKILL.md:20`

Reemplazar «ask only that reference's review question and, for a rendered UI effect, its `UI verification` value» por «ask only that reference's review question, the foreseen authorizations, and, for a rendered UI effect, its `UI verification` value».

### `content/skills/flow-plan/references/questions.md`

- En `:19`, reemplazar «Reuse explicit decisions already in force; silence, preselection,» por «Silence, preselection,». La intención pasa a la condición nueva.
- Después de la primera condición de «Decide whether a question is needed», añadir:

```markdown
**Condition:** about to ask a question.

**Actions:** first check whether an answer in force already settles it: an earlier answer in this session, the effects an authorization already covers, or a route, scope or design choice recorded in the plan or in persistent memory from an earlier session. When one does, act on it and say in one line which answer covers it. An authorization of an effect carries over from an earlier session only as [effect authorization](effect-authorization.md#what-an-answer-covers) sets.
```

- Después de la condición «several open questions do not depend on one another», añadir:

```markdown
**Condition:** one objective needs several authorizations or choices in sequence, such as the records of one test setup or the steps of an operation the user approved.

**Actions:** ask them as one question that lists the whole set, even when a later step depends on an earlier one's result; ask again only when a result changes the set, naming what changed. A critical effect, as [effect authorization](effect-authorization.md) defines it, and a production deployment keep their own question.
```

- En `:15`, reemplazar «such as the route, close or ticket question» por «such as the route or close question».

### `content/skills/flow-plan/references/plan-review.md:37`

Antes de «Ask the user only for consequential choices not already settled.», añadir:

> Apply without asking a finding whose fix has a conventional default and changes no scope, contract, acceptance criterion, or decision the user made: it is a minor call under [questions](questions.md), and you list it in the progress section as a decision taken without asking, with its reason.

### `content/skills/flow-plan/SKILL.md` (resumen del plan)

Reemplazar «the decisions in force with their IDs;» por «the decisions in force with their IDs, and the minor calls taken without asking with their reasons;».

### `content/skills/flow-plan/references/delivery-decisions.md:24`

Al final del párrafo, añadir:

> A follow-up pull request that fixes the same ticket's change, such as a CI or defect fix after its merge, belongs to the same delivery: the mode's merge rule covers it without a new question. A condition the user attaches to a grant, such as "no merge until I return", lasts only for the scope it states; when that scope ends, the mode's own rule applies again, and a handoff to another session records the condition with its scope, never as a standing restriction.

### `content/skills/flow-research/SKILL.md:38` y `:42`

- En `:38`, reemplazar «ask whether to open one in the same round as the route question below, as its own question with an ID; do not block the investigation on the answer.» por «the route question below carries the ticket: each option that acts on that work says whether it opens a ticket for it, and one more option opens the ticket and stops; choosing an option that names a ticket authorizes opening it. Ask no separate ticket question.».
- En `:38`, reemplazar «Open the ticket only after the user agrees» por «Open the ticket only after the user chooses such an option».
- En `:42`, reemplazar «plus one option to keep investigating and one to stop at the findings» por «plus the ticket option above when it applies, one option to keep investigating and one to stop at the findings».

### `content/skills/flow-close/references/reporting.md:98`

Reemplazar «never bundle commits, pushes, pull requests, deployments, tracker writes, or secret reads into a route option.» por «never bundle commits, pushes, pull requests, deployments, tracker writes, or secret reads into a route option, except opening the ticket that a research route option names, as [flow-research](../../flow-research/SKILL.md#deliver-the-finding) sets.».

### `content/skills/unattended-delegation/SKILL.md` (sección «Effects»)

D10-A, 2026-10-25. Después de la primera oración, añadir: «A critical effect, as [effect authorization](../flow-plan/references/effect-authorization.md) defines it, is taken only when it was authorized in advance for that exact target and ceiling; otherwise skip it, continue the work that does not depend on it, and report it with its target, cost and revert.» Este es el hogar completo de D10-A. La referencia y `global.md:28` solo remiten a él, y `global.md` conserva su invariante. El ejemplo de la migración pasa a «A database migration that was authorized in advance is allowed after a verified backup, …». El de crear un recurso en la nube pasa a «Enabling or creating a cloud resource that only this work uses and that spends no money or paid quota is allowed».

### `_support/docs/architecture/instruction-resources.md`

En la lista de hogares canónicos, añadir: «`flow-plan/references/effect-authorization.md`: before asking to authorize an effect beyond this machine, as `global.md` sets.»

### `tests/content/reading_footprint_test.go`

Añadir `skills+"flow-plan/references/effect-authorization.md"` al escenario `cycle`, después de `delivery-decisions.md`, porque el plan y el build lo leen. La prueba solo mide.
