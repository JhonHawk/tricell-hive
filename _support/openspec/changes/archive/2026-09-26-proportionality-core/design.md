# Diseño

## Contexto verificado (2026-09-26, `aea74f3`)

- **Proporcionalidad:** `global.md:78` dice "Simple questions and mechanical edits do not require a formal workflow or a report", sin definir qué es mecánico. `:80` dice "Mechanical edits alone do not require [flow-build]".
- **Pregunta de cierre:** `global.md:33` la pide en todo trabajo terminado; solo se omite si "the user already answered this close question or explicitly asked to end".
- **Hallazgos incidentales:** `global.md:41` dice "Record the rest with evidence, severity, and a suggested action: ... open an issue in the tracker ... without asking first", sin piso de trivialidad.
- **Delegación:** `flow-build/SKILL.md:30` dice "A change spanning several files with its tests, or a diagnosis likely to produce long output, goes to a child by default".
- **Hijos de UI:**
  - `flow-build/SKILL.md:32`: "the independent UI review and in-vivo children ... are required, not proposed".
  - `verification.md:59`: "Any change with a rendered UI effect requires two independent children".
- **Tests:** `flow-build/SKILL.md:40` define una lista de cambios mecánicos para los tests: "a rename, a move, user-facing text, configuration, formatting, a presentation or documentation edit, or wiring that the compiler, type checker, or existing tests already exercise".
- **Agentes de desarrollo:**
  - `frontend-developer.md:12`: "Handle race conditions and relevant loading, empty, error, and success states".
  - `backend-developer.md:12`: "Check authentication, authorization, error semantics, and resource cleanup".
- **Anthropic** (verificado el 2026-09-26): la plantilla "Avoid over-engineering" dice "Don't add error handling, fallbacks, or validation for scenarios that can't happen ... Only validate at system boundaries (user input, external APIs)".
- **Hive legacy:** `critical-thinking.md:5` dice "For trivial changes (typos, one-line fixes, renames, simple config tweaks, formatting) ... go directly to the action".

## Enfoque

La definición de cambio mecánico vive en `global.md`, que se carga siempre, y `flow-build:40` pasa a remitir a ella: así la regla tiene una sola casa. Cada regla pesada lleva además un calificador local corto, porque una excepción lejana no frena a un modelo débil.

La exención de UI no usa "mecánico": la definición global incluye "a presentation edit", y un cambio de layout es presentación pero sí necesita los hijos (D1-A). Se formula aparte (solo texto o un valor de estilo que no altera layout, estado ni flujo), vive solo en `verification.md:59`, y las demás menciones de la regla de UI remiten a ella (revisión B1).

La regla de alcance del código va en Proportionality, junto a la definición, y mantiene un piso explícito: se valida en los límites de confianza, como hace la plantilla de Anthropic.

## Redacción distribuida (inglés)

### `global.md:78` (reemplaza la segunda frase)

"Simple questions and mechanical edits do not require a formal workflow or a report." pasa a:

> A mechanical change is a rename, a move, user-facing text, configuration that changes no permission, security setting, or dependency, formatting, a documentation edit, a text or style change that alters no layout, state, or flow, or wiring that the compiler, type checker, or existing tests already exercise; a change that alters logic, a contract, permissions, security, money, stored data, concurrency, or a dependency is not mechanical, however small. A simple question needs no formal workflow or report. A mechanical change needs no formal workflow, retained plan, or delegation of the edit itself, and no dedicated review unless project guidance requires one; verify it with the checks that already cover it.

### `global.md` (viñeta nueva después de `:78`)

> - Build only what the task requires. Do not add features, abstractions, configurability, fallbacks, or compatibility layers that no requirement or existing consumer needs; an interface with one implementation or a setting for a value that never changes is unrequested. Do not add error handling or validation for cases the code cannot reach, and validate where input crosses a trust boundary, such as user input, external services, files from outside the project, or other external data. Leave code outside the change's scope unrefactored and uncommented, apart from fixes under the incidental-findings rule.

### `global.md:33` (reemplaza la última frase)

"Skip it only when the user already answered this close question or explicitly asked to end." pasa a:

> Skip it only when the user already answered this close question or explicitly asked to end, or after a mechanical change with no Pending item; that report still carries its cleanup label, delivery state, and reminders, so the user need not ask.

### `global.md:80`

"Mechanical edits alone do not require it" pasa a "Mechanical changes alone do not require it", para usar un solo término.

### `global.md:41`

"Record the rest with evidence," pasa a "Record each remaining defect that has a concrete failure scenario with evidence,". Después de "record it in the plan or handoff when no tracker exists." se agrega:

> Mention style or wording nits, and suspected issues without a concrete failure scenario marked as unconfirmed, only in the report.

### `flow-build/SKILL.md:30`

"A change spanning several files with its tests, or a diagnosis" pasa a "A change spanning several files with its tests, unless it is mechanical, or a diagnosis".

### `flow-build/SKILL.md:32`

"the independent UI review and in-vivo children in [verification](references/verification.md) are required, not proposed" pasa a "the independent UI review and in-vivo children that [verification](references/verification.md) requires are required, not proposed". La exención vive solo en `verification.md:59`.

### `flow-build/SKILL.md:40`

"A mechanical change adds no new test: a rename, a move, user-facing text, configuration, formatting, a presentation or documentation edit, or wiring that the compiler, type checker, or existing tests already exercise; verify it with the checks that already cover it." pasa a:

> A mechanical change, as the global guidance defines it, adds no new test; verify it with the checks that already cover it.

### `verification.md:59`

"Any change with a rendered UI effect requires two independent children" pasa a "Any change with a rendered UI effect, other than one that changes only text or a style value without altering layout, state, or flow, requires two independent children". Al final del párrafo se agrega:

> The implementer checks such a text or style change in the rendered page at the affected viewports, with the longest realistic text and in each theme; if the check shows it alters layout, the two children apply.

### Otras apariciones de la regla de UI (remiten a `verification.md:59`)

- `verification.md:31` (fila "Local candidate"): "a rendered UI effect adds the required independent UI review and in-vivo children" pasa a "a rendered UI effect adds the independent UI review and in-vivo children when the UI gate below requires them".
- `flow-plan/SKILL.md:45`: "A change with a rendered UI effect plans the required independent `review-ux` and `sdd-verify` children" pasa a "A change with a rendered UI effect plans the independent `review-ux` and `sdd-verify` children when the verification reference requires them".
- `flow-plan/references/plan-format.md:92`: "<For a rendered UI effect, include" pasa a "<For a rendered UI effect that the verification reference does not exempt, include".
- `review-ux.md:3`: "Use for the independent review of any change with a rendered UI effect before reporting it complete." pasa a "Use for the independent review of a change with a rendered UI effect before reporting it complete."

### `frontend-developer.md:12`

"Handle race conditions and relevant loading, empty, error, and success states." pasa a "Handle race conditions and the loading, empty, error, and success states that the change introduces or alters."

### `backend-developer.md:12`

"Check authentication, authorization, error semantics, and resource cleanup." pasa a "Check authentication, authorization, error semantics, and resource cleanup where the change affects them."

## Riesgos

- **Clasificación de "mecánico":** un modelo podría clasificar como mecánico un cambio de lógica. Lo frenan la lista cerrada y la exclusión explícita "however small"; `flow-build:40` mantiene el desempate a favor de TDD.
- **Regla de alcance frente a hallazgos incidentales:** "Leave code the change does not touch" podría leerse como prohibición de corregir defectos marginales; por eso lleva la salvedad explícita "apart from fixes under the incidental-findings rule".
- **Tamaño:** `global.md` crece; el aumento se mide en T1 y se justifica en el reporte.
