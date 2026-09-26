# Diseño

Las líneas se refieren a `content/guidance/global.md` en `e7906a4` (41177 bytes). Cada cambio lleva un ID para las tareas y la revisión. Los textos entre comillas son el reemplazo exacto propuesto; la implementación puede ajustar la redacción solo si la revisión independiente confirma que conserva todas las restricciones del original.

## Criterios comunes

- **Condensar** es decir lo mismo con menos palabras. Si una reescritura suelta una restricción, deja de ser condensación y se descarta o se trata como decisión aparte.
- **Duplicado dentro del archivo:** se quita la copia que no es la casa canónica, y solo cuando la otra dice lo mismo.
- **Mover:** en global queda el disparador y la parte que motivó un fallo observado. El skill que recibe el texto lo enlaza desde su `SKILL.md` con la condición de lectura (`_support/docs/architecture/instruction-resources.md`).
- **Evidencia:** la procedencia de cada regla sale de `git log -L`/`-S` y de las carpetas de cambio archivadas, según la auditoría del 2026-09-26. Una regla nacida de un fallo en un modelo de ejecución o monitoreada en sesiones reales no se quita, solo se condensa.

## C1 — Condensar sin cambiar el significado

### Language (`:9`, `:11`, `:12`)

- **L1 `:9`:** "- Preserve code, paths, identifiers, and verbatim evidence." La frase sobre texto superado repite `:111` (Continuity), y el caso de reanudar en otro idioma ya está en `:8`. Ahorro ~142 B.
- **L2 `:11`:** "- Write placeholders in examples and templates in English, even within session-language prose, such as `<project>-specs/openspec/specs/<capability>/spec.md`." Ahorro ~15 B.
- **L3 `:12`:** "- Separate identifiers from localized content: use English semantic i18n keys; user-facing text follows the product language. Preserve established domain values and externally defined or persisted identifiers when compatibility requires them, keep each domain vocabulary consistent, and rename contracts or migrate values only under an authorized compatibility plan." Conserva la cláusula de preservación con su condición. Ahorro ~40 B.

### Communication (`:16`–`:26`)

- **M1 `:16`:** quitar solo "Use clear prose"; conservar "explain necessary context without assuming prior knowledge or belaboring the obvious", que no es relleno para un modelo pequeño. El resto queda igual. Hipótesis sin medir: "clear prose" pudo empujar los párrafos de la sesión de Grok `01a0dd4c`. Ahorro ~20 B.
- **M2 `:17`:** solo la frase de PR pasa a "Link each newly opened PR or PR awaiting CI or review, and relevant runs when useful." La obligación de distinguir "outcome, material findings, pending decisions, and next action" queda intacta (revisión del plan, H3). Ahorro ~40 B.
- **M3 `:20`+`:21`:** se fusionan en el bullet de C2 (B1). `:20` desaparece como bullet propio.
- **M4 `:22`:** sin cambio. La reescritura propuesta ponía todo el bullet bajo "consequential", y un modelo literal habría limitado a esas preguntas los IDs, la pregunta nativa y "silence and preselection do not grant authorization" (revisión del plan, H2).
- **M5 `:25`:** "- Give reasoned advice. Challenge an unsupported premise without flattery or automatic agreement. For consequential choices, explain the main reason, tradeoff, and uncertainty; when a premise changes, reassess the recommendation instead of silently substituting a new justification." Ahorro ~66 B.

### Scope, Evidence, Verification gates, Proportionality (`:41`–`:90`)

- **S1 `:41`:** el texto de la auditoría, pero **conservando** "Review alone does not authorize fixes." (`:38` cubre la investigación, no la revisión) y "to the user" en la escalación:
  > - Triage incidental findings from authorized implementation or review; do not silently absorb or drop them. Escalate security or critical issues to the user at once. Fix defects that affect the current work within it, and other defects only when the fix is small, local, low-risk, and covered by the checks already running, reported apart from the requested change; while planning, propose such a defect as an addition for the user to accept instead. For each remaining defect with a concrete failure scenario, check for duplicates, then open an issue in the tracker the plan or project names, with evidence, severity, and a suggested action, without asking first, since this rule authorizes that write, and report its link; without a tracker, record it in the plan or handoff. Put style or wording nits, and suspected issues without a failure scenario marked unconfirmed, only in the report. Pre-existing and intermittent test failures are incidental findings too: save the failing output before rerunning. Review alone does not authorize fixes, and repeated approvals are not standing permission. Stop at the authorized outcome, reporting any unmet criterion.
- **S2 `:43`:** "- Unattended work requires an explicit delegation or a matching declared job with a bounded objective and expiry; then use the `unattended-delegation` skill. Silence, task length, a missing user, or a background or scheduled turn without a matching declared job never activates it, and it does not expand normal authorization or host permissions." Ahorro ~24 B.
- **S3 `:65`:** "- For version-sensitive external technical documentation, prefer Context7 as the first lookup through the installed documentation skill or integration, matching the relevant version; when access or coverage falls short, use current official documentation and state the verification limits. Inspect repository-local facts directly. A recommendation about configuring or using an external library, tool, or service names the documentation and version it rests on, or says that none was consulted." Pierde "Context7 is recommended, not a prerequisite"; "prefer" y "when access or coverage falls short" lo cubren (revisión del plan). Ahorro ~75 B.
- **S4 `:73`:** "- Bounded read-only subagent feedback on a saved implementation plan is part of planning and needs no separate gate approval; only the orchestrator edits the plan, and this exception does not authorize implementation code review, hosted review services, or execution of plan steps." Ahorro ~25 B. Queda en global porque es una excepción de autorización.
- **S5 `:74`:** quitar "Reuse authorization while its scope, target, and conditions remain valid.". Para las puertas de verificación lo cubre `:72` ("unless explicitly authorized already"), y para los efectos de entrega, `:38` ("honor an existing grant while its scope, target, and conditions still apply"). Ahorro ~74 B.
- **S6 `:84`:** el texto de la auditoría (la parte del hijo queda condensada, no quitada, porque los hijos genéricos solo la reciben por esta vía):
  > - Select a child by its deliverable and allowed effects, not just a language label. Its brief gives the objective, repository or declared root, authorized effects, the files or paths it may modify or delete, relevant contracts, expected output, and acceptance evidence; delegation grants no additional authority. A child works only in its assigned working tree and named paths, lists planned deletions before applying them, and stops and reports when it needs a change outside them. It may commit locally in its own working tree and branch only when the brief allows, and never pushes, opens or merges pull requests, or writes to trackers, even when the delivery mode authorizes those effects; the main thread reviews its diff before integrating it and owns those effects.
- **S7 `:86`:** el texto de la auditoría, agregando "and existing project conventions" al conjunto que se pasa:
  > - Pass accessible project-guidance paths, task decisions, existing project conventions, and the exact skill or resource identities or paths the child needs; it does not inherit the parent's conversation or skills already read. A `skill:owner/path` link names a resource inside that skill, read relative to the skill's directory as found through the host's catalog or loader or an explicit task path, not the current directory or agent file; it is not a host command or auto-expanded URI, and reading it does not invoke the skill's whole workflow. The child reads these sources before dependent work and reports any unavailable one before proceeding with what depends on it. Pass the smallest relevant set, not whole manuals or unrelated stacks.
- **S8 `:87`:** "- Label delegated premises as verified facts, source claims, or assumptions, with the relevant evidence or missing check." La segunda frase repite `:88`. Ahorro ~72 B.
- **S9 `:90`:** sin cambio. `flow-plan/references/plan-review.md:26` la cita como "the shared dispatch evidence requirements", y `harness-audit` no se carga al delegar. Moverla la dejaría sin cargar en `flow-plan`, `flow-build` y `unattended-delegation` (revisión del plan, H1 de ambos dominios).

### Project settings, Specs, Support folder (`:94`–`:155`)

- **P1 `:94`:** las frases `Optional:` pasan a "Optional: `Environments`, `Review`, and `Hive guidance: required` (every session there, CI included, runs with this guidance)." Ahorro ~105 B.
- **P2 `:119`+`:122`:** "It holds `project.md` (the project ledger), `specs/<capability>/spec.md` (current requirements per capability, unless a product map, `product/<module>/<view>.md` beside the `openspec` directory, holds them), `changes/<change-id>/` (one active planned change with its requirement deltas), and `changes/archive/YYYY-MM-DD-<change-id>/` (closed changes)." `:122` desaparece; `change-records.md:46` sigue siendo la casa del mapa de producto. Ahorro ~125 B.
- **P3 `:123`:** quitar la segunda frase ("Requirements there describe current behavior; never leave a change's deltas unmerged after it closes."), que repite la primera. Ahorro ~102 B.
- **P4 `:132`:** "- Respect an existing documented destination, including a documentation repository. Do not initialize Git automatically." `:117` ya dice que nunca se crea un repositorio de specs. Ahorro ~30 B.
- **P5 `:139`:** quitar la cola que repite `:136`. Ahorro ~62 B.
- **P6 `:147`+`:148`:**
  > - Name new records under `sessions/` `<topic>.<type>.<extension>`, with a descriptive English `kebab-case` topic shared by related records and an English type: `research` for retained investigations (including findings called a report) and `report` for distinct execution, closure, or presentation deliverables. Use Markdown for ordinary research; honor an explicitly requested format.
  > - Dated folders use the work's initial ISO date, except archived changes, which use the closing date; internal filenames do not repeat it. Continuing work keeps its initial date, scope, and location. Do not automatically rename or migrate existing material to this layout.

  Se pierde "Preserve tool-defined filenames and established source-code conventions", que ya cubre `:10` ("tool-required names"). Ahorro ~273 B.
- **P7 `:149`:** "- Date human-authored records and folders with the user's local calendar date, checking the clock and timezone when uncertain. Preserve original evidence timestamps, with explicit offsets when time affects interpretation." Se pierde "Follow protocol or system requirements where UTC is required"; la revisión debe confirmar que es seguro, porque ese caso no está verificado como conducta por defecto. Ahorro ~76 B.
- **P8 `:155`:** "- For requested organization of existing support material, evidence curation, promotion of findings into durable guidance, or archiving, use the `workspace-conventions` skill." La segunda frase repite la descripción del skill. Ahorro ~106 B.

## C2 — Mover a skills

### B1 — Cuerpo del reporte de backlog a `flow-research`

- **En global** (reemplaza `:20` y `:21`) quedan el disparador, el corte que falló en Grok `01a0dd4c` y el agrupamiento por módulo, que motivó la regla original:
  > - When the user asks for the project's status or a listing of its backlog, give the delivery state and, when the project declares a tracker, a backlog report grouped by product module, queried across all non-terminal states, never as a flat list or per-state dump, and never grouped by tracker state, application, repository, or layer; read [backlog report](skill:flow-research/references/backlog-report.md) before writing it. When they ask for another cut of the tickets, such as those that fit one session, those blocking a release, or a single module's tickets, structure the answer by that cut instead, and include state counts or delivery state only when they change the answer.
- **Nuevo** `content/skills/flow-research/references/backlog-report.md`, con el resto de `:21` sin cambios de significado: umbral de cinco tickets, primera respuesta, origen de los módulos, trabajo transversal, iniciativas, aplazados, bloqueadores, IDs, línea de estados, estado del workspace y orden de recomendación.
- **`flow-research/SKILL.md`**, en "Frame the investigation": "When the question is the project's status or a backlog listing, read [backlog report](references/backlog-report.md) before reporting."
- **Riesgo aceptado (K1-B):** que un modelo no lea la referencia en una pregunta corta de estado. Mitigación:
  - la instrucción de leerla va en global, junto al disparador;
  - el agrupamiento por módulo, las prohibiciones que más fallan y el corte no dependen de la referencia (revisión del plan, H4);
  - el localizador usa la forma `skill:`, que `:86` define.
- **Sin validar:** la validación de enlaces del release no cubre `global.md` (`tooling/management/references.go:17-20`), así que el puntero no se valida (revisión del plan, distribución H2). Ahorro estimado ~0,9 KB.

### B2 — Contenido de `project.md` a `change-records.md`

- **En global** (`:124`): "- Local or sensitive pointers stay in `_support/README.md`. Updating `project.md` or `openspec/` is a versioned change delivered under that repository's Git rules."
- **En `content/skills/flow-plan/references/change-records.md`**, en la sección general "Files", no en el paso 4 de "Close a change", porque ese paso solo aplica al cerrar: "Keep in `project.md` only what no other source states: ticket state comes from the tracker, delivery state from Git, and the handoff from the active change."
- **Por qué no es una regla de ubicación:** dice qué contenido va en `project.md`, no dónde vive un artefacto. La ubicación de `project.md` sigue en global (`:119`).
- **Riesgo:** una edición conversacional de `project.md` sin skill cargado. Ahorro ~157 B.

### B3 — Frase `Review` de `:106`

- Quitar "A declared `Review` is the recommended mechanism when a dedicated review applies, not a substitute for the question." `flow-plan/references/delivery-decisions.md:22,26` ya pide ofrecer los mecanismos que declara el `AGENTS.md` y marcar como Recommended el preferido. Como `:26` dice "A mere mention is not a preference", se agrega ahí: "A `Review` setting in the repository's `## Hive` section is an explicit preference." Ahorro en global ~117 B.

## V1 — Rama base sin respaldo del workspace (K2-A)

`:42` pasa a:
> - A ticket is done when its changes are integrated into the repository's base branch; move it then, since this rule authorizes that tracker write, unless the project declares a later completion point. Record promotion to later environments separately with the project's mechanism, such as a release, an environment label, or a comment, without reopening the ticket. The base branch is the one work branches are cut from and merged back into, as the repository's `Base branch` setting declares; with environment branches it is the first of them, such as `development`, and later ones receive changes only by promotion.

Si falta el valor, aplica `:106` (se pregunta cuando la tarea lo necesita y se registra en la sección del repositorio). Ahorro ~147 B.

## V2 — Imágenes de revisión UI (K3-A)

En `content/skills/flow-build/references/verification.md:67`, "put review captures inside the established session folder" pasa a remitir a la regla global sin repetir la ubicación: "put review captures where the global UI-image rule places them". Lo que sigue en la línea se conserva.

## Presupuesto

El ahorro estimado es de ~3,4 KB tras la revisión del plan (sin S9, M2 y M4 reducidos, L3 con su condición), a medir después de editar. `globalGuidanceBudget` baja al tamaño nuevo en `tests/content/budget_test.go`, en el mismo commit.
