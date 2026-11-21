# Inventario de oraciones y su destino (T1)

Base del build: `development` en `7f6277e0`. Cada línea citada es la actual en ese árbol, ubicada por su contenido; los números de `tasks.md` T1 venían de `2376cbbb`. Evidencia: observado, leído de los archivos del worktree `.claude/worktrees/reporting-family-leaves` el 2026-11-21. Ningún archivo de `content/` se modificó.

## Cómo se contó

- Oración: texto separado en `. ` y `; ` y por viñeta, como fija `tasks.md` T1. Una fila de tabla de `close-modes` cuenta por su celda «Ends with» (5 filas); la plantilla de la tabla de hallazgos de `close-footer` cuenta una vez (`S2.T`). Los encabezados `#`, el frontmatter y las líneas en blanco no cuentan.
- Las líneas de cada fuente son las que nombra T1, resueltas al árbol actual: `flow-deploy` línea 18 es el paso 1 y los pasos 6 y 8 son las líneas 29 y 31-33; `flow-close` pasos 6 y 8 son las líneas 19 y 21; `global.md` «Questions and routes» pasó de las líneas 133-137 a 108-112, y sus tres viñetas son las 110-112. `unattended-delegation` 15-16, 20 y 28 y `merge-gate` 9 y 14 no se movieron.
- Destino: `se queda` = la oración sigue en su archivo (T4 y T5 solo cambian cómo la enlaza o cuánto repite); `se borra` = duplicado de otro destino. `hive-rule-report-endings` y `hive-rule-close-report` son los dos hogares; sus secciones son las de `design.md`.
- Modo para los bytes: K = el texto llega entero, C = se funde con otra oración (se cuenta al 75 %), D/S = no suma bytes al hogar.

## Resumen

### Oraciones por fuente

| Fuente | Oraciones |
| --- | --- |
| hive-rule-next-step-block | 18 |
| hive-rule-close-footer | 27 |
| hive-rule-report-endings | 14 |
| hive-rule-close-report | 30 |
| hive-rule-close-modes (columna «Ends with») | 5 |
| git-workflow pasos 9–10 | 7 |
| flow-deploy línea 18 y pasos 6 y 8 | 13 |
| flow-close descripción y pasos 6 y 8 | 9 |
| unattended-delegation líneas 15-16, 20, 28 | 30 |
| hive-rule-merge-gate líneas 9 y 14 | 15 |
| hive-mode líneas 14, 25-26 | 12 |
| flow-build líneas 14, 20 | 13 |
| plan-handoff líneas 21, 28 | 4 |
| global.md «Questions and routes» (110-112) | 10 |
| **Total** | **207** |

### Oraciones por destino

| Destino | Oraciones |
| --- | --- |
| `hive-rule-report-endings` (hogar de cómo termina el mensaje) | 45 |
| `hive-rule-close-report` (hogar del reporte final) | 42 |
| se quedan en su archivo de origen (flujos, `close-modes`, `merge-gate`, `global.md`) | 116 |
| se borran por duplicado | 4 |
| **Total** | **207** |

Detalle de los dos hogares por sección:

| Hogar | Sección o fila | Oraciones |
| --- | --- | --- |
| hive-rule-close-report | «Final report» | 11 |
| hive-rule-close-report | «Findings at the close» | 10 |
| hive-rule-close-report | «Findings» | 10 |
| hive-rule-close-report | «Candidates» | 7 |
| hive-rule-close-report | «Findings at the close» › «Find the next ticket» | 1 |
| hive-rule-close-report | línea «Read» | 1 |
| hive-rule-close-report | «Variants» fila «PR abierto en ready o esperando a un revisor» | 1 |
| hive-rule-close-report | «Variants» fila «Despliegue» | 1 |
| hive-rule-report-endings | fila «Cierre con footer» | 15 |
| hive-rule-report-endings | fila «Paso a otro flujo» | 10 |
| hive-rule-report-endings | «Same-turn continuations» | 9 |
| hive-rule-report-endings | fila «Acción única» | 5 |
| hive-rule-report-endings | fila «Análisis, explicación, revisión o diagnóstico» | 2 |
| hive-rule-report-endings | línea «Read» | 1 |
| hive-rule-report-endings | frase de apertura de la tabla | 1 |
| hive-rule-report-endings | fila «Evaluación de un ticket elegido» | 1 |
| hive-rule-report-endings | fila «Cierre con footer» / fila «Cierre sin footer» | 1 |

### Las 62 líneas de «Rule leaves»

`global.md` «Rule leaves» tiene 40 viñetas (líneas 63-102) con 62 enlaces; el 62 de T1 cuenta enlaces, no viñetas, y el inventario los cubre por viñeta con su número de enlaces. Solo 3 viñetas cambian en este plan: 67 (`close-hygiene` y `close-footer`), 71 (`report-endings`) y 72 (`next-step-block`); 4 enlaces tocados, 58 intactos. La prueba literal del disparador es de `hive-core-stage-packets` (D7 A) y aquí no se toca.

## Estimación de bytes

Método (inferencia sobre el texto actual, no un build): tamaño actual de la hoja menos el texto de sus oraciones da el esqueleto (frontmatter, título, viñetas); se suman las oraciones que el inventario le asigna (K entera, C al 75 %), más el armazón nuevo de `design.md` (encabezados, tabla de variantes de 828 B, tabla de formas de terminar).

| Hoja | Hoy | Estimado | Límite | Resultado |
| --- | --- | --- | --- | --- |
| `hive-rule-report-endings` | 2,602 B | ≈ 7,779 B (con T5: ≈ 8,000 – 8,300 B) | 7,000 B | no cabe: faltan ≈ 800 B, y T5 pide además llevar a «Same-turn continuations» las condiciones de `hive-mode`, `flow-build` y `plan-handoff` que `S1.9` solo resume |
| `hive-rule-close-report` | 3,776 B | ≈ 5,825 B | 4,000 B | pasa de 4,000 B; cabe en 7,000 B, así que entra a las excepciones de `leaf_skill_test.go` (T2) |

Desglose de `report-endings`: esqueleto 274 B + las 14 de `report-endings` 2,215 B + las de `next-step-block` 3,385 B + las de `close-footer` (encabezado y respuestas) 1,554 B + armazón de tabla 351 B. Desglose de `close-report`: esqueleto 352 B + sus oraciones 3,424 B + «Findings at the close» 1,135 B + tabla de variantes y encabezado 828 B + 60 B de encabezado.

Palancas para que `report-endings` quepa, sin quitar una regla: (1) `S1.2` enumera los pares de flujos (≈ 330 B) que la tabla de formas ya nombra; (2) `S1.9` y `S3.6` repiten las cinco continuaciones que «Same-turn continuations» fija una sola vez (≈ 600 B); (3) `S2.5`-`S2.16` pueden ser una lista de respuestas de una línea cada una. Si aun así pasa de 7,000 B, la salida es sacar la sintaxis por host y la compactación a su propia hoja, decisión del padre porque contradice D1 A.

Efecto en `global.md` (28,593 B hoy): se pierden la viñeta 72 (101 B), la cláusula del footer en la 67 (80 B) y `S14.10` (100 B); la viñeta 71 gana ≈ 45 B para nombrar el paso a otro flujo y el cierre. Neto ≈ −235 B, cerca de 28,360 B: no crece sobre la base.

## Oraciones por fuente

### S1 · hive-rule-next-step-block (18)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S1.1 | `hive-rule-next-step-block:9` | Read before you end a report that moves from one flow to the next, or a work item's close that starts another. | se borra: duplicado de hive-rule-report-endings línea «Read» (S3.1 ampliada) |  |
| S1.2 | `hive-rule-next-step-block:11` | Moving from one flow to the next (research's status report to a ticket's research, research to create-design, plan or… | hive-rule-report-endings · fila «Paso a otro flujo» |  |
| S1.3 | `hive-rule-next-step-block:11` | End the report with a separate next-step block: a `---` rule, a bold heading such as **▶ Next step: plan (recommended… | hive-rule-report-endings · fila «Paso a otro flujo» (forma del bloque) |  |
| S1.4 | `hive-rule-next-step-block:11` | Offer a flow's command only when the next step needs that flow's procedure: an investigation, a design in Figma, a pl… | hive-rule-report-endings · fila «Paso a otro flujo» |  |
| S1.5 | `hive-rule-next-step-block:11` | No flow needs a ticket: when the work has none, the command's argument is a short description of the work or its chan… | hive-rule-report-endings · fila «Paso a otro flujo» |  |
| S1.6 | `hive-rule-next-step-block:11` | Add the other replies the user can give, such as keep investigating or what to change, as a short list with a letter… | hive-rule-report-endings · fila «Paso a otro flujo» | se funde con S3.3 (respuestas con letra) |
| S1.7 | `hive-rule-next-step-block:11` | Only the user's reply to this block, or a flow command they type, starts the next flow; | hive-rule-report-endings · fila «Paso a otro flujo» | el enlace al footer pasa a «Findings at the close» de close-report |
| S1.8 | `hive-rule-next-step-block:11` | a reply to a question or the acceptance of a proposal only settles or redirects the scope, except a fix the [close fo… | hive-rule-report-endings · fila «Paso a otro flujo» | el enlace al footer pasa a «Findings at the close» de close-report |
| S1.9 | `hive-rule-next-step-block:11` | The other exceptions are stages covered by the explicit [E2E grant](../flow-build/references/e2e-coordination.md) and… | hive-rule-report-endings · «Same-turn continuations» | hogar único de las cinco continuaciones y del E2E |
| S1.10 | `hive-rule-next-step-block:11` | typing its command loads that skill whole for that flow, and a flow or skill entered by continuation is read whole be… | hive-rule-report-endings · «Same-turn continuations» | hogar único de las cinco continuaciones y del E2E |
| S1.11 | `hive-rule-next-step-block:11` | Keep the question tool for decisions inside a flow: scope, design, delivery and review, validation and findings. | hive-rule-report-endings · fila «Paso a otro flujo» |  |
| S1.12 | `hive-rule-next-step-block:12` | When the block leads from an approved plan to its build in the same session, or starts another work item after a clos… | hive-rule-report-endings · fila «Paso a otro flujo» (respuesta de compactar) |  |
| S1.13 | `hive-rule-next-step-block:12` | That compaction is not one «during work» of the [checkpoint](../hive-rule-checkpoint/SKILL.md): after it, end again w… | hive-rule-report-endings · fila «Paso a otro flujo» (respuesta de compactar) |  |
| S1.14 | `hive-rule-next-step-block:13` | Command syntax by host, for that block: `/flow-x` in Claude Code, Grok, Cursor and Antigravity; | hive-rule-report-endings · «Same-turn continuations» (sintaxis por host, como fija design.md) |  |
| S1.15 | `hive-rule-next-step-block:13` | `$flow-x` in Codex; | hive-rule-report-endings · «Same-turn continuations» (sintaxis por host, como fija design.md) |  |
| S1.16 | `hive-rule-next-step-block:13` | `/skill:flow-x` in Pi; | hive-rule-report-endings · «Same-turn continuations» (sintaxis por host, como fija design.md) |  |
| S1.17 | `hive-rule-next-step-block:13` | `@flow-x` in OpenCode. | hive-rule-report-endings · «Same-turn continuations» (sintaxis por host, como fija design.md) |  |
| S1.18 | `hive-rule-next-step-block:13` | Write the commands of several hosts only when you cannot tell which host you run in. | hive-rule-report-endings · «Same-turn continuations» (sintaxis por host, como fija design.md) |  |

### S2 · hive-rule-close-footer (27)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S2.1 | `hive-rule-close-footer:9` | Read at step 8 of [flow-close](../flow-close/SKILL.md), before the footer. | se borra: duplicado de hive-rule-report-endings línea «Read» (cubre «before you end a message that reports, hands over or closes») |  |
| S2.2 | `hive-rule-close-footer:11` | End the final report with [the next-step block](../hive-rule-next-step-block/SKILL.md), never with the question tool. | hive-rule-report-endings · fila «Cierre con footer» | se funde con S1.3 (bloque, nunca la herramienta de preguntas) |
| S2.3 | `hive-rule-close-footer:11` | Its heading recommends **close** with its reason, unless a proposed finding is P1 or higher, when it recommends seein… | hive-rule-report-endings · fila «Cierre con footer» (encabezado) |  |
| S2.4 | `hive-rule-close-footer:11` | with no step pending and neither case, its heading is **▶ Closed**, with no recommendation. | hive-rule-report-endings · fila «Cierre con footer» (encabezado) |  |
| S2.5 | `hive-rule-close-footer:11` | Its replies come from this work's context, each only when it applies: | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.6 | `hive-rule-close-footer:13` | **Close:** runs the pending steps, as [close modes](../hive-rule-close-modes/SKILL.md) sets; | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.7 | `hive-rule-close-footer:13` | only while one is pending. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.8 | `hive-rule-close-footer:14` | **Resolve findings now**, naming them, such as "resolve H2 and H3 in this branch": they enter `flow-build` on the wor… | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.9 | `hive-rule-close-footer:15` | **See the findings**, with their count and highest severity, such as "see the 3 findings (1 P1)". | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.10 | `hive-rule-close-footer:16` | **Explain a finding or a result**, naming it. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.11 | `hive-rule-close-footer:17` | **Turn a candidate into a mechanism**, naming it: only for a candidate whose record counts two sessions or more; | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.12 | `hive-rule-close-footer:17` | it enters `flow-build` as **Resolve findings now** does. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.13 | `hive-rule-close-footer:18` | **Find the next ticket**, as **Find the next ticket** below sets, run only when the user picks it. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.14 | `hive-rule-close-footer:19` | **Hygiene**, with the count of pending items as [batch archive](../flow-close/references/batch-archive.md) step 1 cou… | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.15 | `hive-rule-close-footer:19` | only while one is pending. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.16 | `hive-rule-close-footer:20` | **Continue** with what remains of the work, naming it, such as the next phase of a plan in progress. | hive-rule-report-endings · fila «Cierre con footer» (respuestas) |  |
| S2.17 | `hive-rule-close-footer:22` | **See the findings** shows one table, after checking each finding for an existing ticket as [create-issue](../create-… | hive-rule-close-report · «Findings at the close» |  |
| S2.T | `hive-rule-close-footer:24-26` | Plantilla de la tabla de hallazgos: columnas ID, Finding, Detail and impact, Recommendation, con una fila de ejemplo H1. | hive-rule-close-report · «Findings at the close» |  |
| S2.18 | `hive-rule-close-footer:28` | Recommend fixing now or a ticket as the ticket rule of [scope changes](../hive-rule-scope-changes/SKILL.md) sets; | hive-rule-close-report · «Findings at the close» |  |
| S2.19 | `hive-rule-close-footer:28` | recommend ignoring when the finding did not recur or has no impact. | hive-rule-close-report · «Findings at the close» |  |
| S2.20 | `hive-rule-close-footer:28` | After the table, ask the findings as a round, as [asking a round](../hive-rule-questions/SKILL.md) sets: each finding… | hive-rule-close-report · «Findings at the close» |  |
| S2.21 | `hive-rule-close-footer:28` | the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". | hive-rule-close-report · «Findings at the close» |  |
| S2.22 | `hive-rule-close-footer:28` | A ticket answer creates it as that skill sets; | hive-rule-close-report · «Findings at the close» |  |
| S2.23 | `hive-rule-close-footer:28` | a fix-now answer goes as **Resolve findings now** sets; | hive-rule-close-report · «Findings at the close» |  |
| S2.24 | `hive-rule-close-footer:28` | an ignore answer is noted in the report. | hive-rule-close-report · «Findings at the close» |  |
| S2.25 | `hive-rule-close-footer:28` | When every finding has an answer, end again with the close footer, without the findings replies. | hive-rule-close-report · «Findings at the close» |  |
| S2.26 | `hive-rule-close-footer:30` | **Find the next ticket** runs research's [status report](../hive-mode/playbooks/ticket-status.md#find-the-next-ticket… | hive-rule-close-report · «Findings at the close» › «Find the next ticket» |  |

### S3 · hive-rule-report-endings (14)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S3.1 | `hive-rule-report-endings:9` | Read before you end an analysis, an explanation, a review, a diagnosis, a ticket assessment or a work item's close, o… | hive-rule-report-endings línea «Read» (reescrita: «Read before you end a message that reports, hands over or closes») |  |
| S3.2 | `hive-rule-report-endings:11` | Three reports end differently. | hive-rule-report-endings · frase de apertura de la tabla (la fijan session_audit_endings.go:14 y single_home_test.go:82) |  |
| S3.3 | `hive-rule-report-endings:11` | An analysis, explanation, review or diagnosis the user asked for without asking for a change ends at its findings, wi… | hive-rule-report-endings · fila «Análisis, explicación, revisión o diagnóstico» |  |
| S3.4 | `hive-rule-report-endings:11` | it never recommends a flow the user did not ask for, and when the findings confirm a defect in the current project, o… | hive-rule-report-endings · fila «Análisis, explicación, revisión o diagnóstico» |  |
| S3.5 | `hive-rule-report-endings:11` | The assessment of a ticket the user picked for research, such as "research COM-918", ends as [hive-mode](../hive-mode… | hive-rule-report-endings · fila «Evaluación de un ticket elegido» (remite a hive-mode paso 5) |  |
| S3.6 | `hive-rule-report-endings:11` | When the user's own request asked for a change and named it with its target, not whether or how to make one, and rese… | hive-rule-report-endings · «Same-turn continuations» (investigación a flow-build) |  |
| S3.7 | `hive-rule-report-endings:11` | A reply to a question or a proposal never counts as that request. | hive-rule-report-endings · «Same-turn continuations» (investigación a flow-build) |  |
| S3.8 | `hive-rule-report-endings:11` | A work item's close ends with the block [flow-close](../flow-close/SKILL.md) sets; | hive-rule-report-endings · fila «Cierre con footer» / fila «Cierre sin footer» | apunta a close-modes en vez de a flow-close |
| S3.9 | `hive-rule-report-endings:11` | the other hand-overs end with [the next-step block](../hive-rule-next-step-block/SKILL.md). | se borra: duplicado de hive-rule-report-endings · fila «Paso a otro flujo» (S1.2) |  |
| S3.10 | `hive-rule-report-endings:12` | A single action that changes no repository file and is not a tracker write «Scope and authorization» already authoriz… | hive-rule-report-endings · fila «Acción única» |  |
| S3.11 | `hive-rule-report-endings:12` | the answer authorizes only that action. | hive-rule-report-endings · fila «Acción única» |  |
| S3.12 | `hive-rule-report-endings:12` | Creating the ticket then follows [create-issue](../create-issue/SKILL.md), in the same turn. | hive-rule-report-endings · fila «Acción única» |  |
| S3.13 | `hive-rule-report-endings:12` | An answer to another decision authorizes a ticket or other single action only when the chosen option's label or descr… | hive-rule-report-endings · fila «Acción única» |  |
| S3.14 | `hive-rule-report-endings:12` | text around the question does not. | hive-rule-report-endings · fila «Acción única» |  |

### S4 · hive-rule-close-report (30)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S4.1 | `hive-rule-close-report:9` | Read when you write the final report of a close or classify a finding. | hive-rule-close-report línea «Read» (reescrita: cubre toda variante) |  |
| S4.2 | `hive-rule-close-report:13` | Report the whole work item in one message, each part only when it has content: | hive-rule-close-report · «Final report» |  |
| S4.3 | `hive-rule-close-report:15` | **Integrated:** what changed and where: the pull request, the commits, the ticket's state; | hive-rule-close-report · «Final report» |  |
| S4.4 | `hive-rule-close-report:15` | for a pull request left open, its head commit, the base commit its current-base check passed on, and the CI result. | hive-rule-close-report · «Variants» fila «PR abierto en ready o esperando a un revisor» | sale de la parte Integrated hacia la fila (design.md) |
| S4.5 | `hive-rule-close-report:16` | **Resolved:** each acceptance criterion with how it was shown, the corrections the user asked for, and the defect's p… | hive-rule-close-report · «Final report» |  |
| S4.6 | `hive-rule-close-report:17` | **Try it:** the block the verification floor sets, when the change can be seen or run. | hive-rule-close-report · «Final report» |  |
| S4.7 | `hive-rule-close-report:18` | **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs. | hive-rule-close-report · «Final report» |  |
| S4.8 | `hive-rule-close-report:19` | **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data. | hive-rule-close-report · «Final report» |  |
| S4.9 | `hive-rule-close-report:20` | **Support material:** each item with its path and the action applied, or listed as pending, as [support material](../… | hive-rule-close-report · «Final report» |  |
| S4.10 | `hive-rule-close-report:21` | **Findings:** the defects not fixed, as [Findings](#findings) sets. | hive-rule-close-report · «Final report» |  |
| S4.11 | `hive-rule-close-report:22` | **Candidates:** the work's repetitions, as [Candidates](#candidates) sets. | hive-rule-close-report · «Final report» |  |
| S4.12 | `hive-rule-close-report:23` | **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user, review bots or bas… | hive-rule-close-report · «Final report» (Next); el merge pendiente y de quién también en la fila «ready» |  |
| S4.13 | `hive-rule-close-report:25` | When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `create-report` page… | hive-rule-close-report · «Final report» | T2 hecho: dice que la página se ofrece como una respuesta del bloque de respuestas del cierre y enlaza `report-endings`. T3 añade a la fila «Cierre con footer» de `report-endings` la respuesta **Offer a `create-report` page** (solo cuando el usuario pide guardar o compartir el reporte, o este no cabe en una pantalla), nombrada así en `close-report` |
| S4.14 | `hive-rule-close-report:29` | A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there,… | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.15 | `hive-rule-close-report:29` | Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, a… | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.16 | `hive-rule-close-report:29` | P1 is blocking, as [review findings](../hive-rule-review-findings/SKILL.md) defines it, and must be fixed before merge; | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.17 | `hive-rule-close-report:29` | P2 should be fixed but does not block; | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.18 | `hive-rule-close-report:29` | P3 is minor. | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.19 | `hive-rule-close-report:29` | A finding of P2 or higher with a concrete failure scenario is proposed; | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.20 | `hive-rule-close-report:29` | the rest get one line each in the report. | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.21 | `hive-rule-close-report:29` | P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.22 | `hive-rule-close-report:29` | findings of P1 or higher are proposed one by one. | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.23 | `hive-rule-close-report:29` | Never open a ticket without the user's answer, which unattended delegation gives at its step 5. | hive-rule-close-report · «Findings» (se queda: definición y prioridades; el ancla #findings) |  |
| S4.24 | `hive-rule-close-report:33` | A candidate is a repetition the work did not turn into a mechanism under [defect fix](../hive-rule-defect-fix/SKILL.m… | hive-rule-close-report · «Candidates» |  |
| S4.25 | `hive-rule-close-report:33` | The report lists each one on one line, with what repeated, how many times and the mechanism that point would choose,… | hive-rule-close-report · «Candidates» |  |
| S4.26 | `hive-rule-close-report:33` | a work with none has no such part. | hive-rule-close-report · «Candidates» |  |
| S4.27 | `hive-rule-close-report:33` | [Script lifecycle](../project-layout/references/script-lifecycle.md) sets where a script goes at each level. | hive-rule-close-report · «Candidates» |  |
| S4.28 | `hive-rule-close-report:35` | At step 7, search the project's persistent memory for each candidate by the terms of its purpose, then create or upda… | hive-rule-close-report · «Candidates» |  |
| S4.29 | `hive-rule-close-report:35` | A candidate seen in two sessions or more is offered in the close footer; | hive-rule-close-report · «Candidates» | enlaza la respuesta «Turn a candidate» de la fila de footer de report-endings |
| S4.30 | `hive-rule-close-report:35` | it is not a finding, since it is no defect. | hive-rule-close-report · «Candidates» |  |

### S5 · hive-rule-close-modes (columna «Ends with») (5)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S5.1 | `hive-rule-close-modes:13` | Step 8's footer | se queda en close-modes columna «Ends with» | el texto pasa a un enlace a hive-rule-report-endings · fila «Cierre con footer» |
| S5.2 | `hive-rule-close-modes:14` | Step 8's footer | se queda en close-modes columna «Ends with» | el texto pasa a un enlace a hive-rule-report-endings · fila «Cierre con footer» |
| S5.3 | `hive-rule-close-modes:15` | The report, without the footer | se queda en close-modes columna «Ends with» | T2 hecho: la frase que abre «Variants» de close-report dice que abandono, higiene, fin de sesión y cierres desatendidos usan la fila «Close at the completion point»; T3 enlaza esa fila |
| S5.4 | `hive-rule-close-modes:16` | The report, without the footer | se queda en close-modes columna «Ends with» | enlaza hive-rule-report-endings · fila «Cierre sin footer» y close-report · fila `hold` |
| S5.5 | `hive-rule-close-modes:17` | The report, without the footer | se queda en close-modes columna «Ends with» | enlaza hive-rule-report-endings · fila «Cierre sin footer» y close-report · fila Pausa |

### S6 · git-workflow pasos 9–10 (7)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S6.1 | `git-workflow:22` | **Verify each effect.** After each authorized action, check the resulting state instead of the command's intent: the… | se queda en git-workflow paso 9 (verificar efectos) |  |
| S6.2 | `git-workflow:22` | When `gh` fails with a GraphQL rate-limit error, read the pull request and its checks through REST, as the state read… | se queda en git-workflow paso 9 (verificar efectos) |  |
| S6.3 | `git-workflow:22` | Never query the REST endpoints in a loop. | se queda en git-workflow paso 9 (verificar efectos) |  |
| S6.4 | `git-workflow:22` | After a merge or a push to the base, check the CI it triggered there, waiting and handling a failure as [merge gate](… | se queda en git-workflow paso 9 (verificar efectos) |  |
| S6.5 | `git-workflow:22` | Complete the remaining effects the delivery already authorizes, then report what you observed, the checks omitted and… | git-workflow paso 9: «Complete the remaining effects the delivery already authorizes» | la cláusula «then report what you observed, the checks omitted and the steps left, as the final report sets» pasa a «entrega lo observado a flow-close»; su contenido ya es Integrated, Omitted y Next de close-report |
| S6.6 | `git-workflow:23` | **Close.** After verifying a merge you performed, a push to the base, or a `ready` stop, or after step 8's stop for a… | se queda en git-workflow paso 10 (cierre) |  |
| S6.7 | `git-workflow:23` | when flow-close is already running, as when it merges the change record's closing pull request, continue it instead. | se queda en git-workflow paso 10 (cierre) |  |

### S7 · flow-deploy línea 18 y pasos 6 y 8 (13)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S7.1 | `flow-deploy:18` | **Name the run.** Identify the candidate, the environments in the declared order, and the tickets it carries; | se queda en flow-deploy paso 1 |  |
| S7.2 | `flow-deploy:18` | state them in step 4's question when one is asked, and in the report. | flow-deploy paso 1 | «and in the report» queda cubierto por la fila de despliegue de close-report («la ejecución») |
| S7.3 | `flow-deploy:18` | Without an `Environments` setting, ask once which environments the request reaches. | se queda en flow-deploy paso 1 |  |
| S7.4 | `flow-deploy:29` | **Check what is served.** After each deployment, verify the served identity and the health of every affected target a… | se queda en flow-deploy paso 6 | T4: la espera de recuperación (S7.6-S7.7) se nombra mensaje de bloqueo (reasons-and-blockers); S7.8 queda como estado, y su vocabulario entra en la fila de despliegue |
| S7.5 | `flow-deploy:29` | A trigger that fires is reported to the user with its evidence. | se queda en flow-deploy paso 6 | T4: la espera de recuperación (S7.6-S7.7) se nombra mensaje de bloqueo (reasons-and-blockers); S7.8 queda como estado, y su vocabulario entra en la fila de despliegue |
| S7.6 | `flow-deploy:29` | Run its recovery action when the answer of step 4 covers it; | se queda en flow-deploy paso 6 | T4: la espera de recuperación (S7.6-S7.7) se nombra mensaje de bloqueo (reasons-and-blockers); S7.8 queda como estado, y su vocabulario entra en la fila de despliegue |
| S7.7 | `flow-deploy:29` | otherwise ask for it as step 4 sets and run nothing meanwhile. | se queda en flow-deploy paso 6 | T4: la espera de recuperación (S7.6-S7.7) se nombra mensaje de bloqueo (reasons-and-blockers); S7.8 queda como estado, y su vocabulario entra en la fila de despliegue |
| S7.8 | `flow-deploy:29` | A deployment with an unmatched identity, an unfinished rollout, or a failed check or error observed after it, such as… | se queda en flow-deploy paso 6 | T4: la espera de recuperación (S7.6-S7.7) se nombra mensaje de bloqueo (reasons-and-blockers); S7.8 queda como estado, y su vocabulario entra en la fila de despliegue |
| S7.9 | `flow-deploy:31` | **Report** the run named in step 1 and, per environment, the identity served, the health and the functional acceptanc… | hive-rule-close-report · «Variants» fila «Despliegue» | paso 8 pasa a entregar lo observado a flow-close, sin escribir el reporte |
| S7.10 | `flow-deploy:31` | Record the promotion as step 4 sets. | se queda en flow-deploy paso 8 (registro y Completion) |  |
| S7.11 | `flow-deploy:32` | **Completion:** when the repository's `## Hive` declares `Completion` (`<state> on <ref> · …`), once the deployment i… | se queda en flow-deploy paso 8 (registro y Completion) |  |
| S7.12 | `flow-deploy:32` | a ticket whose due check did not run stays out of it. | se queda en flow-deploy paso 8 (registro y Completion) |  |
| S7.13 | `flow-deploy:33` | Then run [flow-close](../flow-close/SKILL.md) for the deploy's cleanup: the tickets and the change record closed at i… | flow-deploy paso 8 (continúa a flow-close) | «and its final report, run at once» pasa a «flow-close escribe el reporte con la fila de despliegue» |

### S8 · flow-close descripción y pasos 6 y 8 (9)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S8.1 | `flow-close:3` | Close a work item and its session hygiene. | se queda en flow-close descripción (disparadores) |  |
| S8.2 | `flow-close:3` | Use when a work item reaches its completion point, is abandoned or is paused; | se queda en flow-close descripción (disparadores) |  |
| S8.3 | `flow-close:3` | when an operational task ends; | se queda en flow-close descripción (disparadores) |  |
| S8.4 | `flow-close:3` | when the user ends the session; | se queda en flow-close descripción (disparadores) |  |
| S8.5 | `flow-close:3` | or when the user asks for session hygiene, such as "haz higiene"; | se queda en flow-close descripción (disparadores) |  |
| S8.6 | `flow-close:3` | not to clean up, prune or reorganize existing docs or support material, which project-layout does. | se queda en flow-close descripción (disparadores) |  |
| S8.7 | `flow-close:3` | Closes the change record, cleans merged branches, task temporaries and processes, proposes what to do with the work's… | flow-close descripción | cambia «a close footer whose reply starts any cleanup left» a «as close-modes sets: with the footer or the report» |
| S8.8 | `flow-close:19` | Write the final report as [Final report](../hive-rule-close-report/SKILL.md#final-report) sets. | flow-close paso 6 (único escritor del reporte final) | elige la fila de close-report que aplica |
| S8.9 | `flow-close:21` | End with the close footer as [close footer](../hive-rule-close-footer/SKILL.md) sets. | flow-close paso 8 | enlaza la fila de close-modes y de report-endings que aplica en lugar de close-footer |

### S9 · unattended-delegation líneas 15-16, 20, 28 (30)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S9.1 | `unattended-delegation:15` | **Close the open answers at activation.** For an explicit delegation or a leaving message whose objective reaches a b… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.2 | `unattended-delegation:15` | State the answers already in force, which include each class of effect the user approved in the plan's effects round,… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.3 | `unattended-delegation:15` | A class the user refused stays refused: skip its effects and report them. | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.4 | `unattended-delegation:15` | A class left unanswered adds no authorization, is not an open answer of this step and puts no work under `hold`: step… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.5 | `unattended-delegation:15` | Ask the missing ones in one round, in the activation message, as [asking in text](../hive-rule-questions/SKILL.md) fo… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.6 | `unattended-delegation:15` | this is the named exception to the rule of [asking a round](../hive-rule-questions/SKILL.md) that the turn ends with… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.7 | `unattended-delegation:15` | A reply that answers or changes an answer keeps it. | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.8 | `unattended-delegation:15` | Until the answers arrive, work under `hold` and decide nothing they cover: edit and verify the working tree as [flow-… | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.9 | `unattended-delegation:15` | The `hive-review-ux` child still runs for a UI change; | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.10 | `unattended-delegation:15` | it publishes nothing. | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.11 | `unattended-delegation:15` | The work is then incomplete; | se queda en unattended-delegation paso 4 (no es contenido de reporte) |  |
| S9.12 | `unattended-delegation:15` | the end report (step 9) names the review and any child that did not run as not run, lists the unanswered questions wi… | unattended-delegation paso 4 / paso 9 | su «end report (step 9) names…, lists the unanswered questions… ends with its pause line» es contenido del resumen del paso 9 |
| S9.13 | `unattended-delegation:16` | **Decide everything else yourself.** Take the recommended option for every other question, log each decision with its… | unattended-delegation paso 9 (resumen de sesión) | efectos con su reversión y decisiones tomadas |
| S9.14 | `unattended-delegation:16` | This includes questions whose own rule says never to act without the user's answer: adding test infrastructure, as [f… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.15 | `unattended-delegation:16` | This step is the instruction for those effects, which the objective needs. | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.16 | `unattended-delegation:16` | Opening such a ticket is allowed (closing it as not planned reverts it); | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.17 | `unattended-delegation:16` | a parent ticket with children is never created here and is reported as unresolved, as [breakdown](../create-issue/ref… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.18 | `unattended-delegation:16` | step 6 bounds the rest, so a deletion with no verified way back is skipped and reported. | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.19 | `unattended-delegation:16` | The end report presents each with how to revert it; | unattended-delegation paso 9 (resumen de sesión) | efectos con su reversión y decisiones tomadas |
| S9.20 | `unattended-delegation:16` | the close's memory write records them as your decisions, not the user's, and not the test-infrastructure choice as a… | unattended-delegation paso 9 (resumen de sesión) | efectos con su reversión y decisiones tomadas |
| S9.21 | `unattended-delegation:16` | The exceptions are a question whose own rule says what unattended work does with it, such as the external-project que… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.22 | `unattended-delegation:16` | Research offers no proposals while active: the plan settles the UI. | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.23 | `unattended-delegation:16` | While the delegation is active, capture and publish the first rendered view through [walkthrough](../hive-rule-walkth… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.24 | `unattended-delegation:16` | this replaces only the first-view question of [first view](../flow-build/references/first-view.md), not its evidence… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.25 | `unattended-delegation:16` | Delegate and wait as [delegation](../hive-rule-delegation/SKILL.md) and [merge gate](../hive-rule-merge-gate/SKILL.md… | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.26 | `unattended-delegation:16` | Require no branch, hook or tracking system for unattended status. | se queda en unattended-delegation paso 5 (decisiones, no reporte) |  |
| S9.27 | `unattended-delegation:20` | **End.** At expiry or when the objective is complete, stop unattended effects, then report the completed and unresolv… | unattended-delegation paso 9 | T4: resume efectos con reversión, decisiones y lo no resuelto y enlaza el reporte de cada cierre; «as close report sets» sobre «completed … work» sale (duplicado de Resolved y Next de close-report) |
| S9.28 | `unattended-delegation:28` | The task record keeps only progress. | se queda en unattended-delegation «Decision log» |  |
| S9.29 | `unattended-delegation:28` | Never commit or publish the log unless the user asks; | se queda en unattended-delegation «Decision log» |  |
| S9.30 | `unattended-delegation:28` | the end report cites it by path. | se queda en unattended-delegation «Decision log» |  |

### S10 · hive-rule-merge-gate líneas 9 y 14 (15)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S10.1 | `hive-rule-merge-gate:9` | Wait for CI with one command that blocks until the run ends, such as `gh pr checks --watch` or `gh run watch --exit-s… | se queda en merge-gate (espera de CI) |  |
| S10.2 | `hive-rule-merge-gate:9` | When the host stops the command before the run ends, start the same command again, with no message between runs. | se queda en merge-gate (espera de CI) |  |
| S10.3 | `hive-rule-merge-gate:9` | Never redirect it to a file to read at intervals, never wait with a loop of sleeps and status queries, and send one m… | se queda en merge-gate (espera de CI) |  |
| S10.4 | `hive-rule-merge-gate:9` | Then answer the reviews as [PR reviews](../hive-rule-pr-reviews/SKILL.md) sets, and merge as the bullets below set: u… | se queda en merge-gate (espera de CI) |  |
| S10.5 | `hive-rule-merge-gate:9` | When a CI job fails, diagnose it from its log first and fix its cause; | se queda en merge-gate (espera de CI) |  |
| S10.6 | `hive-rule-merge-gate:9` | rerun it only when the log shows a transient failure. | se queda en merge-gate (espera de CI) |  |
| S10.7 | `hive-rule-merge-gate:9` | After the merge or a push to the base, wait for the CI run it triggered there with a blocking command on that run, su… | se queda en merge-gate (espera de CI) |  |
| S10.8 | `hive-rule-merge-gate:9` | When the host reports a background command's exit, run it in the background and run `flow-close` meanwhile: the close… | se queda en merge-gate (mensaje del CI de la base) | T2 hecho: la fila `ready` de «Variants» enlaza `merge-gate` para el CI de la base y el merge pendiente, sin copiarlos |
| S10.9 | `hive-rule-merge-gate:9` | When it ends, report its result; | se queda en merge-gate (mensaje del CI de la base) | T4: no es un reporte final; S10.8 menciona el reporte del cierre: se reescribe a «la sección Next de close-report» |
| S10.10 | `hive-rule-merge-gate:9` | handle a failure as this leaf sets, on a new branch from the base. | se queda en merge-gate (mensaje del CI de la base) | T4: no es un reporte final; S10.8 menciona el reporte del cierre: se reescribe a «la sección Next de close-report» |
| S10.11 | `hive-rule-merge-gate:14` | A human reviewer holds the merge in every delivery mode. | se queda en merge-gate (retención por revisor humano) |  |
| S10.12 | `hive-rule-merge-gate:14` | Right before merging, read `gh api repos/<owner>/<repo>/pulls/<n>` and the reviews as [PR reviews](../hive-rule-pr-re… | se queda en merge-gate (retención por revisor humano) |  |
| S10.13 | `hive-rule-merge-gate:14` | The hold lasts while `requested_reviewers` names a `User` or `requested_teams` a team, or while a person who reviewed… | se queda en merge-gate (retención por revisor humano) |  |
| S10.14 | `hive-rule-merge-gate:14` | an approval counts on the head commit, or on one followed only by this step's base merges. | se queda en merge-gate (retención por revisor humano) |  |
| S10.15 | `hive-rule-merge-gate:14` | During the hold, stop with the pull request open, never wait or poll, and tell the user in the report that the merge… | se queda en merge-gate (retención por revisor humano) | T2 hecho: la fila `ready` de «Variants» enlaza `merge-gate` para el CI de la base y el merge pendiente, sin copiarlos |

### S11 · hive-mode líneas 14, 25-26 (12)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S11.1 | `hive-mode:14` | **Complete request.** When the request already says everything that changes the result, such as "add this user to the… | se queda en hive-mode paso 0 |  |
| S11.2 | `hive-mode:14` | An action that changes no repository file is then done and reported, under the global guidance's effect rules; | se queda en hive-mode paso 0 |  |
| S11.3 | `hive-mode:14` | a change to repository files takes step 5's route. | se queda en hive-mode paso 0 |  |
| S11.4 | `hive-mode:14` | A request whose deliverable is a design in Figma, such as replicating a screen, is not drawn here: check the connecti… | hive-mode paso 0 (paso de la continuación a create-design) | T5: la condición va a report-endings «Same-turn continuations»; el paso se queda |
| S11.5 | `hive-mode:14` | with it, continue into [create-design](../create-design/SKILL.md) in the same turn, which draws it and does not repor… | hive-mode paso 0 (paso de la continuación a create-design) | T5: la condición va a report-endings «Same-turn continuations»; el paso se queda |
| S11.6 | `hive-mode:14` | A request to review a design someone already drew goes to its [review entry](../create-design/references/review.md) i… | hive-mode paso 0 | no es una de las cinco continuaciones del hogar; se queda |
| S11.7 | `hive-mode:14` | A request whose deliverable is a requirement document, such as the PRD, the functional requirements or the business r… | hive-mode paso 0 (paso de la continuación a create-prd) | T5: condición a report-endings «Same-turn continuations» |
| S11.8 | `hive-mode:14` | When one bounded lookup (as «Proportionality» defines it) answers a question, report the answer with its source and u… | se queda en hive-mode paso 0 |  |
| S11.9 | `hive-mode:25` | When the user picked a ticket for research without asking for a change, and the findings confirm work it asks for, en… | hive-mode paso 5 (hogar del final de la evaluación de un ticket) | la fila «Evaluación de un ticket» de report-endings remite aquí |
| S11.10 | `hive-mode:25` | Every other analysis, explanation, review, diagnosis or assessment without a requested change ends as [report endings… | hive-mode paso 5 | puntero a report-endings; no cambia |
| S11.11 | `hive-mode:26` | When the user's own request asked for a change and named it with its target, not whether or how to make one, and its… | hive-mode paso 5 (paso de la continuación a flow-build) | T5: la condición va a report-endings «Same-turn continuations»; el paso se queda |
| S11.12 | `hive-mode:26` | A reply to your question, or the acceptance of a proposal you made, settles or redirects the scope and never starts a… | hive-mode paso 5 | la regla «una respuesta no inicia un flujo» ya está en S1.7-S1.8 y S3.7; queda solo «close the what again and end with the next-step block» |

### S12 · flow-build líneas 14, 20 (13)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S12.1 | `flow-build:14` | Enter with a ready plan, in a change folder or kept in the conversation as [flow-plan](../flow-plan/SKILL.md#where-th… | se queda en flow-build (entrada) | el enlace «close footer» pasa a close-report «Findings at the close» |
| S12.2 | `flow-build:14` | other fixes go to `flow-plan`. | se queda en flow-build (entrada) | el enlace «close footer» pasa a close-report «Findings at the close» |
| S12.3 | `flow-build:14` | Without a plan, ask the delivery and review questions as [delivery questions](../hive-rule-delivery-ask/SKILL.md) set… | se queda en flow-build (entrada) | el enlace «close footer» pasa a close-report «Findings at the close» |
| S12.4 | `flow-build:14` | when it changes a behavior an existing `<Specs>/specs/<capability>/spec.md` describes, first follow [spec delta](refe… | se queda en flow-build (entrada) | el enlace «close footer» pasa a close-report «Findings at the close» |
| S12.5 | `flow-build:20` | When every task is done, reproduce the repository's CI as [local CI](../hive-rule-checks/SKILL.md#local-ci) sets. | se queda en flow-build paso 5 |  |
| S12.6 | `flow-build:20` | Then continue into `flow-verify` in the same turn, with no next-step block: read its skill whole, then hand it the ca… | flow-build paso 5 (paso de la continuación a flow-verify) | T5: la condición va a report-endings «Same-turn continuations» |
| S12.7 | `flow-build:20` | When the user asked to stop after build, end instead with the next-step block toward `flow-verify`; | flow-build paso 5 | «end instead with the next-step block toward flow-verify» enlaza report-endings fila «Paso a otro flujo» |
| S12.8 | `flow-build:20` | when they asked to skip verify, integrate as the delivery says and run `flow-close`. | se queda en flow-build paso 5 |  |
| S12.9 | `flow-build:20` | Verify owns [check reuse](../hive-rule-checks/SKILL.md#check-reuse), checks the acceptance criteria in its walk, and… | se queda en flow-build paso 5 |  |
| S12.10 | `flow-build:20` | A mechanical change skips verify: check its criteria, if it has any, then, under `interactive`, hold the delivery's v… | se queda en flow-build paso 5 |  |
| S12.11 | `flow-build:20` | integrate as the delivery says, and run `flow-close`. | se queda en flow-build paso 5 |  |
| S12.12 | `flow-build:20` | An operational task, as «Proportionality» defines it, skips the CI run and verify: check its criteria and run `flow-c… | se queda en flow-build paso 5 |  |
| S12.13 | `flow-build:20` | End the final report with one line naming each principle read from the code principles that changed a decision, or none. | flow-build paso 5 | PENDIENTE: ordena una línea del reporte final («one line naming each principle…») que close-report no tiene; ver sección de oraciones sin colocar |

### S13 · plan-handoff líneas 21, 28 (4)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S13.1 | `flow-plan/references/plan-handoff.md:21` | End the summary with the [next-step block](../../hive-rule-next-step-block/SKILL.md), with its compaction reply for a… | plan-handoff (cierre del resumen) | enlaza report-endings fila «Paso a otro flujo» con su respuesta de compactar; el texto de la forma del bloque no se copia |
| S13.2 | `flow-plan/references/plan-handoff.md:21` | The build route is a route, not a request for leave to proceed: a plan does not authorize implementation, and only th… | plan-handoff | «only the user's reply starts flow-build, except…» duplica S1.7 y S1.9; queda el hecho propio: el plan no autoriza implementar |
| S13.3 | `flow-plan/references/plan-handoff.md:28` | Continue to `flow-build` in the same turn when the explicit [E2E grant](../../flow-build/references/e2e-coordination.… | plan-handoff (paso de la continuación a flow-build) | T5: la condición va a report-endings «Same-turn continuations» |
| S13.4 | `flow-plan/references/plan-handoff.md:28` | Show the summary the same way, then continue. | plan-handoff |  |

### S14 · global.md «Questions and routes» (110-112) (10)

| ID | Origen | Oración | Destino | Nota |
| --- | --- | --- | --- | --- |
| S14.1 | `guidance/global.md:110` | Ask only for a choice that is the user's and changes what happens next ([questions](skill:hive-rule-questions/SKILL.m… | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.2 | `guidance/global.md:110` | a message that asks for a decision or validation, delivers a result or a blocker, or closes work opens with a bold co… | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.3 | `guidance/global.md:110` | a decision has an ID or a session question's fixed header, a `Situation:` line, lettered options, the recommended fir… | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.4 | `guidance/global.md:110` | Give the effect of a step as its reason, never that an instruction requires it; | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.5 | `guidance/global.md:110` | end a blocker or a failure with the smallest fix you can see, and do what your access and authorization allow instead… | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.6 | `guidance/global.md:111` | Ask a round through the host's native question tool when it waits for the answer and the host permits it for that que… | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.7 | `guidance/global.md:111` | a card the user left unanswered is never replaced by the round in the text. | se queda en global.md «Questions and routes» (invariante de preguntas) | no es duplicado de report-endings; las lee questions |
| S14.8 | `guidance/global.md:112` | Moving from one flow to the next is never asked with the question tool: end the report with a next-step block that na… | global.md «Questions and routes» (invariante mínimo de pasar a otro flujo) | T3: el enlace pasa a report-endings; la cola «except the continuations and fixes that leaf lists» apunta a «Same-turn continuations» |
| S14.9 | `guidance/global.md:112` | only the user's reply to it, or a flow command they type, starts the next flow, except the continuations and fixes th… | global.md «Questions and routes» (invariante mínimo de pasar a otro flujo) | T3: el enlace pasa a report-endings; la cola «except the continuations and fixes that leaf lists» apunta a «Same-turn continuations» |
| S14.10 | `guidance/global.md:112` | An analysis, explanation, review or diagnosis ends with a block of lettered replies and no command. | se borra: duplicado de hive-rule-report-endings · fila «Análisis, explicación, revisión o diagnóstico» (la línea del índice de report-endings nombra ese momento) |  |

## Líneas del índice «Rule leaves» (`global.md` 63-102)

| Línea | Enlaces | Hojas | Estado |
| --- | --- | --- | --- |
| 63 | 1 | hive-rule-git-commit | se queda como está |
| 64 | 1 | hive-rule-merge-gate | se queda como está |
| 65 | 1 | hive-rule-pr-reviews | se queda como está |
| 66 | 1 | hive-rule-promotion-path | se queda como está |
| 67 | 3 | hive-rule-close-modes, hive-rule-close-hygiene, hive-rule-close-footer | TOCADA: close-hygiene pasa de «steps 4 to 7» a los pasos 4, 5 y 7 (T6); la cláusula «before the close footer: close footer» se borra (T3) |
| 68 | 1 | hive-rule-close-report | se queda como está |
| 69 | 1 | hive-rule-questions | se queda como está |
| 70 | 1 | hive-rule-reasons-and-blockers | se queda como está |
| 71 | 1 | hive-rule-report-endings | TOCADA: report-endings sigue y su disparador absorbe «handing over to the next flow» y «closing», que dejan las líneas 72 y 67 (T3); el texto exacto lo fija el plan hive-core-stage-packets (D7 A) |
| 72 | 1 | hive-rule-next-step-block | BORRADA: next-step-block desaparece (T3) |
| 73 | 1 | hive-rule-delivery-ask | se queda como está |
| 74 | 1 | hive-rule-foreign-commits | se queda como está |
| 75 | 2 | hive-rule-local-record, hive-rule-local-write | se queda como está |
| 76 | 1 | hive-rule-local-script | se queda como está |
| 77 | 1 | hive-rule-test-accounts | se queda como está |
| 78 | 2 | hive-rule-ticket-state, hive-rule-trackers | se queda como está |
| 79 | 1 | hive-rule-scope-changes | se queda como está |
| 80 | 1 | hive-rule-project-settings | se queda como está |
| 81 | 1 | hive-rule-reuse-before-writing | se queda como está |
| 82 | 1 | hive-rule-verification-floor | se queda como está |
| 83 | 2 | hive-rule-tests, hive-rule-tests-no-runner | se queda como está |
| 84 | 1 | hive-rule-defect-fix | se queda como está |
| 85 | 3 | hive-rule-delegation, hive-rule-child-brief, hive-rule-reviewer-brief | se queda como está |
| 86 | 3 | hive-rule-business-rules, hive-rule-business-record, hive-rule-business-ticket | se queda como está |
| 87 | 2 | hive-rule-external-facts, hive-rule-external-effects | se queda como está |
| 88 | 3 | hive-rule-ui-plan, hive-rule-design-stage, hive-rule-ui-example | se queda como está |
| 89 | 3 | hive-rule-before-captures, hive-rule-ui-verification, hive-rule-a11y-audit | se queda como está |
| 90 | 4 | hive-rule-walkthrough, hive-rule-viewer, hive-rule-viewer-open, hive-rule-viewer-close | se queda como está |
| 91 | 1 | hive-rule-change-folder | se queda como está |
| 92 | 1 | hive-rule-checkpoint | se queda como está |
| 93 | 2 | hive-rule-checks, hive-rule-background-wait | se queda como está |
| 94 | 2 | hive-rule-review-findings, hive-rule-review-evidence | se queda como está |
| 95 | 2 | hive-rule-project-verification, hive-rule-project-verification-format | se queda como está |
| 96 | 1 | hive-rule-stored-data | se queda como está |
| 97 | 1 | hive-rule-security-boundaries | se queda como está |
| 98 | 3 | hive-rule-suppression-review, hive-rule-no-comments, hive-rule-blast-radius | se queda como está |
| 99 | 1 | hive-rule-unattended-effects | se queda como está |
| 100 | 1 | hive-rule-loops | se queda como está |
| 101 | 1 | hive-rule-prose | se queda como está |
| 102 | 1 | hive-rule-interrogate | se queda como está |

## Oraciones sin colocar o con decisión pendiente

Ninguna oración quedó sin destino: las 207 tienen uno. Cuatro piden una decisión del padre antes de T2-T5:

1. `S12.13` (`flow-build/SKILL.md:20`, «End the final report with one line naming each principle read from the code principles…»): ordena una parte del reporte final que `close-report` no tiene, y T4 deja a `flow-close` paso 6 como único escritor. Hay que decidir si pasa a una parte nueva de «Final report» o se borra; se queda por ahora en `flow-build`.
2. `S4.13` (`close-report/SKILL.md:25`, ofrecer una página `create-report` «as one reply of the close footer»): la lista de respuestas de `close-footer` (`S2.6`-`S2.16`) no trae esa respuesta. Al fundirla en la fila de footer de `report-endings`, hay que añadir esa respuesta o cambiar la frase.
3. `S5.3` (`close-modes/SKILL.md:15`): una sola fila agrupa seis situaciones y T3 la enlaza a «Cierre sin footer» más «la fila de `close-report` que aplica»; las filas de `close-report` no cubren abandono, higiene ni fin de sesión. Propuesta: tratarlas con la fila «Cierre en el punto de terminación» y registrarlo en la tabla.
4. `S10.8` y `S10.15` (`merge-gate/SKILL.md:9,14`): dicen qué pone el reporte de cierre sobre el CI de la base y el merge pendiente, que ahora también están en la fila «ready» de `close-report` (`S4.4`, `S4.12`). Se dejan en `merge-gate` (T4 así lo manda) y la fila de `close-report` los enlaza en lugar de copiarlos.

## Texto del PR (para cuando la entrega salga de `hold`)

**Title:** refactor(guidance): give reporting two homes, the final report and how a message ends

**What changed.** The final report lives in `hive-rule-close-report` (one row per variant plus the findings round of the close). How each message ends lives in `hive-rule-report-endings` (a row per way of ending, the close footer's replies and the same-turn continuations). `hive-rule-next-step-block` and `hive-rule-close-footer` are deleted and every link points to the section that took their text. `flow-close` step 6 writes the only final report; `git-workflow`, `flow-deploy` and `unattended-delegation` hand over or summarize. A close ends as its `close-modes` row sets. The 13 playbooks name their rows. `close-hygiene` covers steps 4, 5 and 7. `site/`, `llms.txt`, `README.md`, `CHANGELOG.md`, the architecture page and the rule-enforcers table follow.

**Why.** One work item wrote two final reports, `flow-close` said the close always ends with the footer while `close-modes` listed six closes without it, and the same-turn continuation was written in six places.

**Size of `global.md`.** 28,593 B on base `7f6277e0`; 28,312 B now (`wc -c`), 281 B smaller. The cap of 30,000 B is untouched. `report-endings` is 8,661 B (own limit 9,000 B, D11 A) and `close-report` 6,420 B (7,000 B exception). Each playbook is at most 7,000 B (largest `ticket-status`, 5,917 B).

**Reading footprint.** The plan/build union grows by +9,053 B (+4.91%) without the moments the baseline did not read, against an allowance of 182 B + 8,880 B (D12 A). The test logs: `plan/build union at 2bb96478 (its declared list): 184489 B; candidate with every selected leaf: 200736 B (+16247 B, +8.81%), of which new sources 94363 B`.

**How it was checked.** `go test -count=1 ./tests/content/...`, `./tests/pilot/`, `check_change.py` on this folder; `inventory.md` maps each source sentence to its destination, input for `/code-review`.

**Places checked.** `content/guidance/global.md` («Rule leaves», «Questions and routes»); the 13 playbooks; `hive-rule-close-modes`, `hive-rule-close-hygiene`, `hive-rule-merge-gate`; `flow-close`, `flow-deploy`, `flow-build`, `flow-plan/references/plan-handoff.md`, `git-workflow`, `unattended-delegation`, `hive-mode`; `site/src/content/docs/flows/{close,deploy,overview,build,hive-mode,create-design}.md`; `llms.txt`; `README.md`; `CHANGELOG.md`; `_support/docs/architecture/{instruction-resources,workspace-and-artifacts}.md`; `_support/docs/harness-engineering/{rule-enforcers,session-review-rubric}.md`; `tests/content/*`, `tests/pilot/session_audit_endings.go` and its fixtures (name `report-endings` kept); `_support/openspec/specs/workflow-lifecycle/spec.md` and `project-layout/spec.md:54` (kept, still true).
