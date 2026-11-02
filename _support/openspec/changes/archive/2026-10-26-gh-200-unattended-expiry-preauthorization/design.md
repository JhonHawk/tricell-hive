# Diseño

## Contexto verificado (base `aa283530`, 2026-10-26)

- `content/skills/unattended-delegation/SKILL.md:8` termina la delegación con «any other message than the ones below, such as one that changes the objective, the scope or an answer in force, asks anything other than how the work is going, gives a new instruction, reports something new, or tells you to stop». Solo la mantienen continuar, acusar progreso y preguntar por el estado.
- `content/skills/unattended-delegation/SKILL.md:10` pide declarar la fuente y la cita al activar, pero no dice dónde. En las dos sesiones de hoy quedó solo en el razonamiento.
- `content/guidance/global.md:35` repite el vencimiento para el «me voy»: «expiring when it completes or the user writes anything other than telling you to continue, acknowledging progress, or asking how the work is going».
- `content/skills/flow-build/references/verification.md:27` pide las autorizaciones previstas «while the user is present: in the round that settles the delivery and review questions, before the plan is declared ready or before the first edit». Con un «me voy» durante la investigación, esa ronda todavía no ocurrió.
- `content/skills/flow-plan/references/questions.md`, sección «Use text when no waiting card carries the question», enumera tres condiciones: no hay herramienta, la herramienta vuelve antes de la respuesta, o una pregunta descartada se repite.
- Evidencia de sesión (en #200):
  - **cam** (`c00e1e7f`, Claude Code 2.1.310, `claude-opus-5-5`): la delegación sobrevivió solo porque el aviso de `bb1` incluía «ahora si me voy a dormir». No hubo pregunta de entrega, y el agente planeó un PR igual.
  - **carvento** (`01a177d3`, Pi, `claude-sonnet-5-5`): la delegación paró en «¿hago commit, PR…?» hasta que el usuario pidió «lanzame las preguntas de autorización».
- Supuesto, sin verificar contra la documentación de cada host: una tarjeta nativa que espera respuesta detiene el hilo principal hasta que se contesta. Por eso D3-A pide texto. La regla se escribe por capacidad («a tool that blocks your turn until the answer»), sin nombrar herramientas.

## Enfoque

Cada regla tiene un solo hogar, la skill. `global.md` solo corrige su copia del vencimiento, porque ya la restaba y una copia contradictoria rompería AC1. `questions.md`, `verification.md`, `delivery-decisions.md`, `flow-build/SKILL.md` y `flows/build.md` del sitio solo agregan punteros o una exclusión en el punto donde se aplica la regla. Las dos reglas de `questions.md` que la ronda cambia (no agregar un segundo reporte y preguntar cada opción una vez) se nombran en la skill, que es el único punto de uso.

## Redacción propuesta

**`unattended-delegation/SKILL.md`, párrafo 1, desde «and expires when…»** (sustituye hasta «keep deciding.»):

> and expires when that objective completes or the user sends a message that changes the objective, the scope or an answer in force, asks for other work, or tells you to stop. Only these messages, in any language, keep the delegation for the same objective and expiry: one that tells you to continue, acknowledges progress, or asks how the work is going or about the work; an answer to a question this delegation asked; and one that adds information, a resource, feedback or a correction within the objective and scope. When a message fits both lists, such as a correction that changes the delivery mode, it ends the delegation. Apply what a kept message adds as input to the work, answer a question in one line, say that the delegation continues, and keep deciding. Every other message ends it.

**`unattended-delegation/SKILL.md`, párrafo 2** (sustituye «At activation, state the source…»):

> At activation, state in visible text, not only in your reasoning, the source and the message or answer it rests on, quoted, then the objective, root or repository, references, destination, evidence, and an expiry or deadline. While it runs, every message to the user adds the delegation's state after the progress of the context line that [reporting](../flow-close/references/reporting.md#open-with-the-context-line) sets, such as `· unattended, active`; the message that ends it says that it ended and why. This skill grants no broader tools or host permissions.

**`unattended-delegation/SKILL.md`, sección «Decisions»** (agrega la excepción después de «…left unanswered, which stays unauthorized»):

> , and the questions of the pre-authorization round below, which only the user answers

**Sección nueva, `## Pre-authorization round`, antes de `## Decisions`:**

> When an explicit delegation or a message that the user is leaving activates this mode, and its objective reaches a build or a delivery, find at once which answers the work still needs to reach its delivery point: the delivery mode and review questions of [delivery decisions](../flow-plan/references/delivery-decisions.md), the `UI verification` value for a rendered UI effect, and the foreseen authorizations that [verification](../flow-build/references/verification.md#gate-selection-and-timing) sets, classified as [effect authorization](../flow-plan/references/effect-authorization.md) sets, each critical effect with its exact target and ceiling. Answers in force, from the session or the request, need no question. Ask the missing ones in one round in the activation message, formatted as [questions](../flow-plan/references/questions.md) sets for a question no waiting card carries, never through a tool that blocks your turn until the answer, since the user may already be gone; then keep working. An answer keeps the delegation and sets its delivery point and the effects it authorizes. Until one arrives, and when none does, decide nothing those questions cover: edit and verify as `hold` sets, so the work may start before the first-edit delivery round that [delivery decisions](../flow-plan/references/delivery-decisions.md) sets; run no dedicated review without a review answer; apply the UI gate as it sets for an unresolved `UI verification` value; take the work to a candidate whose tasks are verified at their checkpoints and whose affected tests pass, without commit or publication. At the end, the completion report lists the unanswered round with the same IDs and ends with its pause line: this changes two rules of [questions](../flow-plan/references/questions.md#pause-and-handle-the-answer) for this round only, since that report is not a second report appended to an unanswered question, and repeating the round in it is not asking it again.

**`global.md:35`**, la frase del «me voy» (sustituye desde «expiring when it completes…» hasta «how the work is going.»). Es la invariante completa, porque `global.md` la restata:

> expiring when it completes or the user changes its objective, scope or an answer in force, asks for other work, or tells you to stop. Only these keep it: a message that continues, acknowledges progress, asks how the work is going or about the work, answers a question the delegation asked, or adds information, a resource, feedback or a correction within the objective and scope; a message that fits both lists ends it.

**`questions.md`, condición de la tarjeta (`:47`)**:

> **Condition:** the host offers a native question tool that waits for the answer, and the question is not the round that [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) asks at activation.

**`questions.md`, condición de texto (`:60`)** (agrega el caso a la lista):

> …or a dismissed or timed-out question is asked once more in text; or the round that [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) asks at activation, whose pause line adds that the work continues to a local candidate meanwhile.

**Punteros al momento de la ronda** (una frase en cada lugar que pide resolver la entrega antes de la primera edición):

- `verification.md:27`, después de «…alone at that same point.»: «When the user activates unattended delegation by leaving before that round, it is asked at activation, as [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) sets.»
- `flow-build/SKILL.md:24`, al final del párrafo: «Under unattended delegation, the round is asked as [unattended delegation](../unattended-delegation/SKILL.md#pre-authorization-round) sets, and the work continues without waiting for it.»
- `delivery-decisions.md:5`, después de la primera oración: «Under unattended delegation, it is asked as [unattended delegation](../../unattended-delegation/SKILL.md#pre-authorization-round) sets, and the work continues without waiting for it.»
- `site/src/content/docs/flows/build.md:48`: una frase equivalente para personas.

**Sitio:**

- `index.mdx:97`: sustituir «A message that only tells it to continue, acknowledges progress or asks how it is going does not end that; anything else hands control back to you.» por una frase que diga que el feedback, un recurso o una corrección dentro de la meta tampoco la terminan, y que al irte te pregunta de inmediato, sin detenerse, lo que falta para entregar.
- `skills.md:30`: una cláusula equivalente.

El implementador anota las frases exactas en `tasks.md` antes de verificar.
