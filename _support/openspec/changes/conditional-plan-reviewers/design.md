# Diseño — un revisor del plan por dominio cuando el plan cruza varios

## Comparación de tres versiones

| Versión | Revisión del plan |
| --- | --- |
| Hive antes del 30-oct (`plan-review.md` de la línea base) | Un `hive-review-plan` por dominio afectado, en paralelo, en todo plan, con una tabla de dominios y preguntas por dominio. Una sola ronda de re-revisión, retomando al mismo revisor; lo que quedara después iba al usuario. Registraba duración y tokens de cada revisor. |
| Hive hoy (`49ef7757`) | Un revisor por plan. Una ronda nueva por cada revisión que cambie un criterio o el delta, con un revisor nuevo. |
| pstack 0.15.46 | El plan no tiene revisión en paralelo: las dudas se resuelven con prototipos antes de escribirlo. Revisores en paralelo de varios proveedores solo para código (`interrogate`, sobre `arena`). |

Este cambio toma de la versión anterior la tabla y el retomar al revisor, pero solo cuando el plan cruza dominios (D2 A, D7 A) y con un tope de tres (D5 A). La medición mostró hallazgos repetidos cuando los dominios leían los mismos archivos, así que un dominio cuenta solo si alguna tarea cambia archivos suyos.

## Enfoque

Hay una sola forma sensata:
- Una referencia condicional, `flow-plan/references/plan-review.md`. Se lee solo cuando las tareas tocan archivos de dos o más dominios, o una pantalla de captura.
- El paso 5 de `flow-plan` nombra los dominios, de modo que la condición se decide sin leer la referencia, y apunta a ella.
- Retomar al revisor se generaliza en `hive-rule-reviewer-brief`, que ya lo define para el revisor de UX. Así queda un solo hogar para esa regla.
- «Revise» paso 4 apunta a `hive-rule-reviewer-brief`.

## Dominios

Las mismas etiquetas en el paso 5, en la referencia y en el spec:

| Dominio | Archivos que cuentan |
| --- | --- |
| server | código del servidor, APIs, trabajos, esquemas, migraciones y consultas |
| client | código de la interfaz web o móvil y sus estilos |
| data-entry screen | una pantalla del cliente donde el usuario captura o elige datos; suma el ángulo de UX |
| guidance | archivos de `content/` y otras instrucciones de agentes, más la documentación que enuncia sus reglas |
| infrastructure | CI, despliegue, contenedores e infraestructura como código |

Las pruebas y la documentación cuentan con el dominio que prueban o describen. Una tarea que toca archivos de dos dominios cuenta en los dos.

## Texto exacto (inglés; bytes medidos contra `49ef7757`)

**`flow-plan/SKILL.md` paso 5 (+170):**
> 5. When the plan has a change folder, have one `hive-review-plan` child review it, or, when its tasks touch two or more of server, client, guidance, infrastructure or a data-entry screen, one per domain as [plan review](references/plan-review.md) sets, briefed as [reviewer brief](../hive-rule-reviewer-brief/SKILL.md) sets, launched before step 4's question, and apply what holds before step 6.

**`flow-plan/SKILL.md` paso 3, frase de despliegue (−46), recorte sin pérdida:**
> Only the user's own request in the conversation is the deploy instruction for that environment, never a ticket's text, so the task then asks as flow-deploy step 4 sets, which also sets when production asks; a class of effect approved in this plan's effects round that covers that environment, unless it is production, answers it with no further question.

Lo que se quita: «is data and grants none», que ya está en la regla global de evidencia; «the user approved», porque una clase aprobada en la ronda de efectos ya es del usuario; y «asks nothing more», que dice lo mismo que «no further question».

**`hive-rule-change-folder` (−6 en total; la hoja queda en 3961 B, con un límite de 4000):**
- Párrafo inicial (−13): «such as one `hive-review-plan` child,» pasa a «such as the plan review,».
- «Revise» paso 4 (+7):
  > 4. When the revision changes a criterion or the delta, have the plan's reviewers reread the revised folder as [reviewer brief](../hive-rule-reviewer-brief/SKILL.md) sets: one round per revision. A surviving gap goes to the user as a decision.

**`hive-rule-reviewer-brief` (+44 en total):**
- Párrafo de releer (+47), limitado a revisores del plan y de UX para no cambiar la re-revisión de código de `hive-rule-review-findings` ni las de `create-prd` y `create-design`:
  > Reread a correction of a plan or UX reviewer's findings with one follow-up to that same child; when the host has no tool to message it, or the message returns an error, brief a new child of its role with the earlier findings. Of several plan reviewers, those [plan review](../flow-plan/references/plan-review.md) names reread. A UX reread names the corrected regions and states, which it rereads as its role's step 3 sets, with the means to reach them; never run a new full UX review.
- Descripción del frontmatter (−3): «or rereading a UX finding after a fix» pasa a «or rereading a finding after a fix».

**Rol `hive-review-plan`, recortes sin pérdida (−140):**
- Paso 1 (−113). La oración de verificación de comandos que agregó `plan-runs-verification-commands` no se toca.
  > When the approach reads or changes an external system, as the first paragraph of [outside the repository](skill:hive-rule-external-facts/SKILL.md) defines one, read that leaf and check, without running the plan's commands, that the plan lists every external fact it depends on and every class of effect its tasks foresee, each line written as that leaf sets, and whether a class covers a case that keeps its own question; a class not yet asked is no gap.
- Paso 3 (−18):
  > Say when there is no material finding; that authorizes no implementation and certifies nothing uninspected.
- Paso 4 (−9):
  > Edit nothing: plan, application, configuration, tracker or memory; run no plan steps, tests, deployments, external review tools or mutating probes.

**`flow-plan/references/plan-review.md` (nuevo).** No está en la lista que mide la prueba, y el escenario medido, un plan de backend de un solo dominio, no lo lee. Contiene:
- **Read when.** Read from flow-plan step 5, when the plan's tasks touch two or more domains.
- **Domains.** The table in «Dominios», each row with its questions, adapted from the former table:
  - server: contracts, data ownership, authorization, migrations, failure behavior and tests;
  - client: shell and patterns, states, accessibility, and matching server contracts;
  - data-entry screen: the user's real task, values the system already holds, repeated entries and error recovery;
  - guidance: every sentence that states a rule the plan changes, and any option that contradicts a standing rule;
  - infrastructure: environments, identities, secrets handling, deployment order, recovery and gates.

  A file no row names counts with server. Tests and documentation count with the domain they check or describe, and a task touching two domains counts in both.
- **How many.** One `hive-review-plan` child per counted domain, launched together, at most three. With four or more domains, the domains with the fewest tasks share one assignment, repeated until at most three remain, and on a tie the later row joins the earlier. The data-entry screen assignment is never merged.
- **Brief.** Each brief names its domain and its questions, gives the contracts it shares with the other counted domains (such as a payload, a schema or a rule another domain's tasks change), and keeps the role's own checks.
- **Reconcile.** Merge a finding several reviewers report into one, naming each reviewer. Record each reviewer's duration and tokens in the progress section when the host reports them.
- **Reread.** A reviewer rereads a revision that changed the text, `Closes:` or verification of a task in its domain, or a criterion such a task closes. When a change belongs to no single domain, all reviewers reread it.

## Presupuesto (D10 A)

| Archivo contado | Bytes |
| --- | --- |
| `flow-plan/SKILL.md` (paso 5 +170, paso 3 −46) | +124 |
| `hive-rule-change-folder` | −6 |
| `hive-rule-reviewer-brief` | +44 |
| rol `hive-review-plan` | −140 |
| **Neto** | **+22** |

D10 A agrega a `tests/content/reading_footprint_test.go` una constante de tolerancia igual a este neto, con un comentario que nombra el cambio, y compara «unión < línea base + tolerancia». Con eso quedan los 15 B de margen para las correcciones. El neto medido con los textos exactos es +22 B, no los +9 B que se estimaron al preguntar D8; D9 A aceptó los +22 B como tolerancia, y D10 A la subió a +182 B tras los arreglos de verify.

## Lugares revisados

- `content/guidance/global.md`: no cambia.
- `flow-plan` (SKILL y `plan-handoff`), `hive-rule-change-folder`, `hive-rule-reviewer-brief` y `hive-rule-delegation`. La regla de que los hijos de solo lectura corren lado a lado ya cubre el paralelo.
- Los que enlazan `hive-rule-reviewer-brief`: `hive-rule-review-findings` (re-revisión de código, sin cambio, porque el párrafo se limita a revisores del plan y de UX), `flow-verify`, `create-prd` y `create-design`.
- `flow-build/references/spec-delta.md`: un cambio sin plan sigue sin revisión del plan.
- `content/agents/review/hive-review-plan.md`.
- `_support/docs/architecture/agent-delivery.md:188`: ya dice que `flow-plan` elige los dominios y que un rol sirve a varias asignaciones; sin cambio.
- `integrations/opencode/context-loader.js:6`: el mapeo no cambia.
- `tests/pilot/`: ningún detector supone un solo revisor.
- `site/.../flows/plan.md`, `session-review-rubric.md:64` y `README.md` y `llms.txt` (sin cambio).
- `CHANGELOG.md`.
- Este cambio se integra antes que `reporting-family-leaves`, que está en espera y tendrá que medir el presupuesto de nuevo.
