# Diseño: flujo de decisiones y pausas

## Contexto verificado

Base: `development` en `9c285026` (2026-10-08). Las líneas citadas son de esa revisión.

**Reglas actuales** (inspeccionado):
- La sección de comunicación de `content/guidance/global.md` (líneas 14–33) tiene 2.254 palabras, el 32 % del archivo. El archivo mide 45.134 bytes frente al presupuesto de 45.528 de `tests/content/budget_test.go:12`.
- Desde el 2026-09-15, 53 commits tocaron esa sección. Hay reversiones directas: `6b9e0d49` deshizo `00e784e8`, y `de91b153` fue consecuencia de `9d50665f`.
- Decidir cómo preguntar exige evaluar unas 14 condiciones más ramas por host.
- Conflictos concretos:
  - Cuatro reglas piden abrir el mensaje con algo distinto, sin orden: `:16`, `:17`, `:21` y `:24`.
  - `:21` pide la decisión «before the detail» y la tarjeta «after the supporting detail».
  - `:17` manda enlazar el detalle en vez de reproducirlo, sin excluir las alternativas.
  - `:28` pone las decisiones abiertas dentro de Pendiente.
  - `flow-research/SKILL.md:36` reduce las rutas a una línea con la recomendada, mientras `:30` exige alternativas y sus ventajas y desventajas.

**Sesiones reales del 1 al 8 de octubre de 2026** (observado; detectores por patrones de texto, así que los conteos aproximados son máximos):
- **Tarjetas sin texto:** 82 de 208 respuestas de Claude Code con `AskUserQuestion` no traen texto (Opus 5.5; recontado por el orquestador). Codex: 0 de 18. Grok dejó de fallar al pasar a texto.
- **Decisiones que no quedaron en el chat**, en tres sesiones independientes:
  - Carvento `71388ddc`: las alternativas quedaron solo en COM-786 y el usuario respondió «En lista los tres arreglos».
  - State-research `5aa1a929` #245–#250: las opciones estaban solo en la tarjeta; el usuario la rechazó y escribió «recuérdame cómo quedan las opciones».
  - `35bd0059`.
- **Seguimientos tipo «¿qué queda pendiente?»:** 18 de 329 (5,5 %), sin relación con la longitud del reporte.
- **Pendientes con decisiones:** 51 de 85 secciones de Pendiente contienen palabras de decisión. Es un máximo.
- **Lo que funciona:** responder solo con el identificador («D1-A», «D8-A, D9-A») en Claude Code, Grok y Codex; usar la marca de recomendada; las preguntas en texto con las opciones en el mensaje en Grok.

**Mecánica de las tarjetas por host** (2026-10-08, versiones instaladas). Etiquetas de origen: doc es documentación oficial, src es código fuente, bin es el binario o paquete instalado y local son archivos de sesión de esta máquina.
- **Claude Code 2.1.295:** `AskUserQuestion` admite de 1 a 4 preguntas, de 2 a 4 opciones, una descripción por opción, el sufijo `(Recommended)` y «Other» automático (doc).
  - Bloquea, sin límite de tiempo desde v2.1.200 (doc).
  - Si la rechazas, deja la pregunta y las etiquetas en el historial (bin).
  - El modal tapa la pantalla: anthropics/claude-code#89644, abierto.
  - Con tema claro, el texto de la pregunta puede ser casi invisible: anthropics/claude-code#77242, abierto.
- **Codex 0.161.0:**
  - `request_user_input` solo funciona en modo Plan, salvo con la opción experimental `default_mode_request_user_input` (src).
  - `request_user_input_async` vuelve de inmediato, escribe el título y las opciones como mensaje final y cierra sus controles al terminar el turno; openai/codex#49146 dice que es intencional (src).
- **Grok 1.0.50:** `ask_user_question` admite etiqueta, descripción, vista previa y `(Recommended)` (local, build anterior). El fallo de `9d50665f` fue del modelo, que mandó solo la llamada, no del host.
- **OpenCode v2.0.24:** `question` admite descripciones y la marca `(Recommended)` (src).
  - La tarjeta tapa toda la interfaz: sst/opencode#53506.
  - El texto enviado junto con la tarjeta puede perderse: sst/opencode#49708 y sst/opencode#52735, abiertos.
- **Cursor 2026.10.01:** `AskQuestion` solo tiene etiquetas, sin descripción ni marca de recomendada, y deja todas las opciones en el historial (bin).
- **Pi 1.0.4:** sin tarjeta propia; en esta máquina está la extensión de terceros `rpiv-ask-user-question`.
- **Todos:** rechazar la tarjeta le llega al modelo, en varios hosts, como «usa tu criterio» (local, bin).

**Evidencia externa** (fuentes oficiales y estudios):
- Anthropic documenta que Sonnet 5 sigue las instrucciones al pie de la letra y no generaliza una regla a casos que no nombra. Recomienda redacción en positivo, ejemplos y principios con alcance explícito en lugar de listas de casos.
- OpenAI dice que una guía contradictoria hace que GPT-6 se detenga, y recomienda un AGENTS.md corto.
- xAI dice que Grok sigue mejor las reglas cortas y concretas.
- IFScale muestra que los modelos pequeños degradan con muchas instrucciones. Un estudio factorial de 1.650 sesiones de Claude Code no encontró efecto del tamaño del archivo en modelos frontera.
- Superpowers (v6.4.2, `8ca22db`):
  - Pone la explicación, las opciones y la recomendación juntas y termina con una sola pregunta fija.
  - Pausa muy poco.
  - Atribuye bloqueos de más de ocho horas a una pregunta «enterrada en un resumen» (issues #1545 y #2077).
  - Limita cada aprobación a la etapa presentada.

## Decisiones

| ID | Decisión | Origen |
| --- | --- | --- |
| R1 | Rediseñar el flujo de decisiones con una invariante | usuario, 2026-10-08 |
| D1-A | Tarjeta que se entiende sola, incentivada en todo host que la tenga | usuario |
| D2-A | La investigación termina cada ronda con la pausa «¿te muestro las recomendaciones?» | usuario |
| D3-A | En un plan, una decisión que cambia el diseño va directa con la etiqueta «Pausa» | usuario |
| D4-A | No activar la opción experimental de Codex | usuario |
| S1-A | Alcance: decisiones y pausas; formato, enlaces y glosas sin cambio | usuario |
| E1-A, E2-A | Entrega interactiva y sin revisión dedicada | usuario |
| D5-A | Presupuesto de la guía global +512 bytes como máximo (46.040) | usuario |
| B1-A | Empezar `flow-build` en esta sesión | usuario |

Decisiones derivadas, presentadas al usuario sin objeción:
- La validación manual deja de exigir texto: el recorrido va completo en el mensaje y la tarjeta se entiende sola.
- En Codex, una pregunta que bloquea el resto del trabajo sigue en texto. Se puede revisar cuando una sesión real compruebe que `request_user_input_async` deja las opciones como texto al terminar el turno.
- Los scripts de medición se guardan en `_support/workspace/` sin agregar herramientas al repositorio.

## Diseño propuesto

### Invariante y pausa

```
mensaje (persistente)                      tarjeta (atajo opcional)
├─ explicación / hallazgos                 ├─ pregunta con contexto mínimo
├─ D1: opciones D1-A… con efecto           ├─ opciones con efecto
│      y la recomendada con su razón       │  (en la etiqueta si no hay
├─ cierres (Pendiente, Recordatorios)      │   campo de descripción)
└─ última línea: «Esperando tu decisión:   └─ (Recommended)
   D1 (cómo usamos la tarjeta)»
tarjeta rechazada o cerrada = sin respuesta → se pregunta una vez más
```

Texto en inglés propuesto para la guía. El implementador puede ajustar la redacción, pero cada regla debe nombrar su alcance, porque Sonnet 5 no generaliza, y la guía no puede pasar del presupuesto actual de bytes.

1. **Toda elección completa en el mensaje** (reemplaza `global.md:21`). «Whenever you ask the user to choose, including a decision, a route, the close question, an offer such as showing recommendations, or a yes-or-no, write the choice complete in the message text: the context needed to answer, each option with an ID such as `D1-A` and one line stating its effect (a yes-or-no question needs no option IDs), and the Recommended option with its reason, so the user can answer "D1-A". Keep it there even when the options also live in a ticket, a document, or a native question, since those copies can be closed, hidden, or never opened. A consequential question, one whose answer changes scope, contracts, data, or cost, or is hard to reverse, also states the main tradeoff and uncertainty. In text, give each option one line with only what sets it apart; state once, after the options, what several options share and any route unavailable under project policy with its reason; add no "Other" option or reply example. Reuse explicit decisions already in force; silence, preselection, and a dismissed question grant nothing.»
2. **Tarjeta que se entiende sola** (reemplaza `global.md:22` conservando su primera frase, y absorbe de `:23` la frase de las actualizaciones previas a la tarjeta). «When the host offers a native question tool that waits for the answer, also ask through it. A response that calls the native question tool starts with its full message text in that same response, after memory, tracker, and plan updates, so no tool call separates them; reasoning is not shown to the user, so a response holding only the tool call is the failure. Write the card to stand alone, since a host may hide the message: its question carries the context to answer, each option states its effect in its description, or in its label when the tool has no description field, the Recommended option is marked, and options beyond the tool's limit are split across questions. A question tool that returns before the answer closes at turn end: use it only while independent work continues, and ask a question that blocks the remaining work in text. A dismissed or declined question is no answer: do not decide in its place; ask it once more in text. Take the first answer that arrives. In a question's text and options, state as fact only what you verified in this session and mark the rest as an assumption to confirm.»
3. **Línea de pausa** (nueva). «End every message that waits on the user's choice with one pause line, its last line, after any closing items: a fixed bold label in the session language meaning "waiting for your decision", such as "**Esperando tu decisión:**", then what is asked, by ID with a short gloss when it has one. Use the same label every time. When a native question follows, the pause line ends the message before it. Ask each open choice once; only a dismissed question is asked again.»
4. **Pendiente sin elecciones** (en `:28`). Cambia «An unanswered choice remains an open decision; ask it once with its options» por «An open choice is not a Pending item: it goes in the pause line». Las acciones del usuario siguen en Pendiente.
5. **Precedencia de apertura** (en `:16`). Se agrega: «When several openings apply, open with work still running, then a completed work item, then the answer.» De `:21` desaparece «before the detail that supports it», porque la regla 1 ya no fija el orden.
6. **Lo que el usuario debe sopesar** (en `:17`). Se cambia «Link the retained document for technical detail instead of reproducing it in the conversation» por «Link the retained document for technical detail instead of reproducing it; alternatives, recommendations, and any finding the user must weigh stay in the message». Se borra la frase equivalente de `:23` («State in the message any finding the user must weigh; a linked file supports it, never replaces it»), para que la regla tenga un solo hogar.
7. **Recomendaciones y cierre** (`:30` y `:33`):
   - Recomendaciones: «ask the routes as a choice» en lugar de «through the native question tool», y se agrega «Investigation findings defer this item to `flow-research`'s offer to show recommendations».
   - Cierre: «Ask it as a choice» en lugar de «through the native question tool, or in text with the same options when the session has none or the work is unattended». Se conserva tal cual la frase «Investigation-only findings and follow-ups use `flow-research`'s route rule instead of this close question», cambiando «route rule» por «offer to show recommendations».
   - Lo desatendido sigue cubierto por `unattended-delegation/SKILL.md:16`.
8. **Espera del usuario** (`:23`):
   - Se conserva el pedido autosuficiente y «never a second report».
   - Se borran «Ask a choice through the native question tool, which shows only the choice, after finishing memory, tracker, and plan updates, so no tool call separates this message from it», porque la regla 2 ya lo cubre, y «take the first answer that arrives», que pasa a la regla 2.
   - La validación manual pasa de «goes in the message and ends the turn with a question in text, never the native tool, on every host; any other question that turn, decisions with IDs included, is also in text there» a «goes complete in the message and ends with the pause line».
   - Se borra «Scale the detail to the effort asked», que repite «Match detail to the question» de `:16`.
   - Se borra «Delivery, route and close questions are unchanged.», que pierde sentido cuando esas preguntas siguen las mismas reglas.

**Etiquetas.** «Esperando tu decisión» (o su equivalente en el idioma de la sesión) es la única etiqueta de la línea de pausa. «Pausa» es otra cosa: el encabezado del bloque de una decisión de plan que cambia el diseño (D3-A), que explica por qué se detiene el trabajo. Ese bloque también termina con la línea de pausa.

**Tamaño.** La base mide 45.134 bytes, con 394 de margen hasta el presupuesto de 45.528 (`tests/content/budget_test.go:12`). En la re-revisión se midió el borrador en unos +700 bytes netos. Con estas reglas 2, 3 y 5 más cortas y la eliminación adicional de `:23`, la estimación baja a unos +480. Eso todavía excede el margen en unos 90 bytes, o en unos 300 si la redacción final no logra acortar lo previsto. Decisión D5-A (usuario, 2026-10-08): el presupuesto sube como máximo 512 bytes, a 46.040, y T1 actualiza la constante y su comentario en el mismo commit.

### Investigación (D2-A)

`flow-research` distingue los **hallazgos**, que incluyen las alternativas cuando esas alternativas responden la pregunta investigada (así se conserva `:30`), de las **recomendaciones para resolverlos**: arreglos, rutas y tickets.

Una ronda de hallazgos sobre trabajo posible termina con la línea de pausa ofreciendo las recomendaciones. Cuando además conviene un ticket, cada pregunta lleva su ID, por ejemplo «**Esperando tu decisión:** A1 (¿te muestro las recomendaciones para H1–H3?) y A2 (¿abro un ticket para H2?)».

Si el usuario acepta, o pide hacer el trabajo o continuar, las recomendaciones se presentan como elección, con las reglas 1 a 3:
- las alternativas con su efecto;
- la recomendada;
- una opción por flujo: `flow-plan` cuando abarca varias tareas, `flow-build` solo para un cambio pequeño y entendido, y detenerse en los hallazgos.

La oferta se hace una vez por ronda; vuelve solo si los hallazgos cambian. Esto reemplaza la línea de ruta de `:36` y ubica la pregunta de ticket de `:32` en la misma pausa.

### Plan (D3-A)

`flow-plan` agrega a «Ground the decisions»:
- Una decisión que cambia el diseño del plan o de la investigación en que se apoya se presenta en cuanto aparece. Va en un bloque encabezado con la etiqueta de pausa del idioma de la sesión («Pausa» en español), con la explicación, las opciones con su efecto y la recomendación. Termina con la línea de pausa.
- El trabajo que depende de esa decisión se detiene.
- Las preguntas de diseño independientes siguen juntas en un intercambio, como hoy.

`delivery-decisions.md:14` se reescribe así:
- El mensaje siempre lleva las opciones, una línea cada una, con lo que comparten en una línea aparte.
- La tarjeta lleva la redacción completa por opción.
- En `:30`, «in a text question, omit it» pasa a «in the message text, omit it».
- Si las opciones de revisión pasan del límite de la tarjeta, se dividen según la regla 2.

### Construcción y cierre

`verification.md`, en las líneas `:65`, `:67` y `:77`:
- El traspaso de UI, las rondas de corrección y la validación de la entrega pasan de «question in text» y «without the native question tool» a la línea de pausa, con tarjeta donde el host la ofrezca.
- Se conserva que el trabajo independiente siga en paralelo. Antes de editar, el build verifica en la documentación de cada host que los hijos en segundo plano siguen corriendo mientras una tarjeta espera la respuesta. Donde no se verifique, la pregunta de esa ronda queda en texto, nombrando la capacidad y no el host.

`flow-close` paso 7 enlaza la regla de elección y la línea de pausa sin cambiar las opciones del cierre.

### Hosts

La guía global nombra capacidades, como una tarjeta que espera la respuesta, una que vuelve antes o una sin campo de descripción, y ningún host.

En `agent-delivery.md` se agrega una tabla nueva de tarjetas de preguntas, con su fecha (2026-10-08) y versiones propias. La tabla de presentación existente no cambia sus otras columnas: su columna «Native question» remite a la tabla nueva. Las etiquetas de evidencia siguen las del archivo:
- `doc`;
- `obs`, con el origen entre paréntesis: `obs (src)` para el código fuente, `obs (bin)` para el binario o el paquete y `obs (local)` para archivos de sesión;
- `inf`.

Los párrafos `:139–143`, que afirman que las opciones van solo en la tarjeta, pasan a historia con una nota que remite a este cambio.

### Medición

La línea base del 1 al 8 de octubre de 2026 está arriba. T7 deja en `_support/workspace/2026-10-08-communication-decision-flow/` (ignorado por Git) scripts que solo reportan conteos:
- `bare.py`, con fechas de inicio y fin, reproduce el 82 de 208.
- Los demás números (18 de 329 seguimientos y 51 de 85 pendientes) vienen de los scripts del subagente. El README dice cuáles se reproducen y cuáles no.

Ningún script escribe texto de las sesiones en el directorio retenido; los volcados intermedios van al directorio temporal de cada sesión.

Una semana después del despliegue se repite la medición. Si las tarjetas sin texto, las decisiones perdidas o los seguimientos de pendientes no bajan, o si Grok vuelve a mandar tarjetas vacías, se revisa la regla, con una sola ronda de arreglo.

### Riesgos

- **Tarjetas en Grok:** volver a usarlas reabre el riesgo de tarjetas vacías de `9d50665f`. La mitigación es la regla 2: se conserva la frase del fallo y la tarjeta se entiende sola. La medición posterior lo vigila.
- **Opciones duplicadas:** aparecen en el mensaje y en la tarjeta en Claude Code (ya observado en una sesión del 2026-10-02, `agent-delivery.md:143`), en Codex (la tarjeta que vuelve antes) y en Cursor. Es el costo aceptado de la invariante R1 frente a perder las opciones.
- **Quitar reglas:** sin pruebas con modelos no hay evidencia de que los modelos débiles las cumplan sin ellas. El cambio quita ramas por host y duplicados, conserva la intención de cada regla (la revisión del plan enumeró las que estaban en riesgo y quedaron reubicadas arriba) y deja la medición en sesiones reales como salvaguarda.
