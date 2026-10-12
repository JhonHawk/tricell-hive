# Diseño: ruta iterativa

## Dónde vive

La pregunta de ruta tiene su hogar en `flow-research` (global la cita: «Investigation findings use `flow-research`'s route question»). La ruta iterativa se define ahí, completa, en su propio párrafo. `global.md` solo cambia donde hoy contradice o define la unidad de trabajo, y `flow-build` solo donde decide cuándo cerrar. El límite de efectos no se redefine: se reutiliza el de `global.md` («local editing and verification»).

## Texto propuesto

### `flow-research/SKILL.md`, párrafo nuevo después de la línea 36

> When the next steps alternate investigation and implementation toward one measurable objective, also offer an iterative route in the same route question. A round is one investigation followed by one implementation and its measurement; count and measure only complete rounds. The option states the objective and how it is measured, the round limit you propose for this case, and its stop conditions, and names `flow-research` and `flow-build` as the flows it alternates. Every step must stay within the local editing and verification that implementation already authorizes, in the current project's working tree, the task's scratch directory, or its support workspace, and each round's change must be a small understood change for `flow-build`. Choosing the route authorizes its rounds without a new route question: the first build round asks `flow-build`'s delivery or review question when it applies, and later rounds reuse that answer. Report each round in one line with its measurement and continue. Stop the route when:
>
> - the objective is met;
> - the round limit is reached;
> - a round ends without a measurable improvement;
> - a step needs an effect the route does not cover and no existing grant authorizes, such as a commit, an installation, or deleting or overwriting material the task did not create;
> - a round's change needs `flow-plan`;
> - a finding changes the objective or scope;
> - a decision is genuinely the user's.
>
> When the route stops with changes in the working tree, run `flow-close`: as the completion point when the objective is met, as a pause otherwise. Without such changes, end with the route question.

La frase existente «Within a broader authorized task…» se mantiene: la ruta iterativa es una de esas tareas.

### `global.md:31` (Recommendations)

Antes: «…`flow-plan` when the route spans several tickets, repositories, or steps, `flow-build` only for a small understood change, and both when either fits; start the chosen flow without a new command. The choice authorizes that flow only: never bundle…»
Después: «…and both when either fits, except the iterative route that `flow-research` defines; start the chosen flow without a new command. The choice authorizes that flow only, or that iterative route's rounds: never bundle…»

### `global.md:34` (unidad de trabajo)

Antes: «A work item is a ticket or the scope the user authorized, such as a plan or one of its phases;»
Después: «A work item is a ticket or the scope the user authorized, such as a plan, one of its phases, or an iterative route;»

### `global.md:42`

Antes: «Repeated approvals are not standing permission.»
Después: «Repeated approvals are not standing permission; an iterative route the user chose, as `flow-research` defines it, authorizes its remaining rounds within its bounds.»

### `flow-build/SKILL.md:75`, tras «Run `[flow-close](…)` when any one of these happens: … or the work is paused.»

> Inside an iterative route that `flow-research` defines, a verified round is not the completion point: report it in one line and continue with the next round; run `flow-close` when the route stops.

### `site/src/content/docs/flows/build.md:31` y `research.md`

`build.md:31` añade que en una ruta iterativa el reporte y la pregunta de cierre llegan cuando la ruta para. `research.md` describe la ruta, su tope y sus paradas.

## Decisiones del orquestador (revisión del plan)

- F4: con un modo de entrega que ya hace commits en la sesión, el commit no detiene la ruta, porque un permiso vigente lo autoriza (`global.md:39`); la parada es solo para efectos que ningún permiso cubre.
- N3 (aceptado sin cambio): si el usuario se va a mitad de ruta, la delegación desatendida se activa con su alcance propio; la delegación desatendida está fuera de este cambio.

## Intenciones que se conservan

- «A question, investigation, or proposal does not authorize implementation»: intacta; la ruta la elige el usuario.
- «never bundle commits, pushes, pull requests, deployments, tracker writes, or secret reads into a route option»: intacta y aplicada a la ruta.
- La regla de elección de flujo de `global.md:31` sigue vigente; la ruta es una excepción nombrada y obliga a que cada vuelta sea un cambio pequeño.
- Las decisiones de diseño dentro de una vuelta siguen preguntándose (última parada).

## Alternativas descartadas

- D4-B, inferir la autorización del objetivo: obliga a reescribir la regla central de autorización y quita al usuario el momento de aceptar el tope.
- D5-B/D5-C: tope fijo o sin tope (ver propuesta).

## Ejemplo en cam

En D8 el agente habría ofrecido: «Ruta iterativa — medir en una copia de CAM, corregir el resolver y volver a medir, hasta 4 vueltas o hasta que una vuelta no reduzca los enlaces falsos; solo en el espacio de trabajo local, sin instalar ni reindexar CAM». Al elegirla, D9–D12 habrían sido líneas de progreso.
