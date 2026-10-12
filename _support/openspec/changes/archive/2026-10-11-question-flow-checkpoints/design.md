# Diseño: puntos de decisión y tarjeta nativa

## Contexto verificado

Base: `development` en `bba2fd48` (2026-10-11). Evidencia reunida el 2026-10-11 por tres investigadores y revisada por el orquestador; está resumida en la memoria `hive/research/communication-flow-2026-10-round2`.

**Reglas actuales** (inspeccionado):
- En `content/guidance/global.md`, la línea 21 exige la elección completa en el texto «even when the options also live in … a native question». La 22 pide además la tarjeta. La 23 fija la línea de pausa. Las líneas 31 y 34 remiten a la oferta de `flow-research`.
- `content/skills/flow-research/SKILL.md:36` termina cada ronda con «A1 (show the recommendations for H1–H3?)». La línea 32 pone la pregunta de ticket en esa misma pausa.
- `content/skills/flow-plan/SKILL.md:16` ya agrupa las preguntas de diseño independientes y presenta en el acto, bajo «Pausa», las decisiones que cambian el diseño.
- El archivo mide 46.036 bytes y el presupuesto de `tests/content/budget_test.go:12` es de 46.040.

**Sesiones reales** (observado; regex sobre los historiales; corte por la hora de inicio de sesión respecto del despliegue del 2026-10-09 00:49Z):
- **La oferta de recomendaciones** apareció 5 veces después del corte: Claude 3, Cursor 1 y Codex 1. Las 5 se contestaron solo con consentimiento («a ver, muestralas», «si», «A1», «si», «Si, por favor»). Esperas medidas: 8,0, 18,1 y 6,8 minutos.
- **Claude, antes y después:**
  - Tarjetas sin texto: de 70 de 184 (38 %) a 2 de 35 (6 %).
  - Mensaje que repite las opciones de la tarjeta: de 11 % a 91 %.
  - Tarjetas con varias preguntas: de 46 % a 29 %.
  - Sesión `1528596c` (2026-10-09): D1, D2 y D3 salieron en tarjetas separadas con 1 a 2 minutos entre sí.
- **Cursor:** 2 de 2 tarjetas de cierre salieron sin el reporte antes (`b6185612`, `7d3001ac`, 2026-10-09). Hubo una pregunta en texto sin línea de pausa ni IDs (`13035f46`, 2026-10-11).
- **Quejas del usuario:**
  - «las preguntas se repiten, es porque son dos bloques diferente» (`b8c6aeeb`, 2026-10-03, sobre Codex).
  - «recuerdame como quedan las opciones» (`5aa1a929`, 2026-10-03: la tarjeta se descartó y las opciones solo estaban en ella).
- **Límites:** después del corte hay 6 sesiones de Claude, 5 de ellas sobre Hive. Grok, OpenCode y Pi no tienen preguntas posteriores al corte.

**Hosts** (versiones instaladas el 2026-10-11: Claude Code 2.1.296, Codex CLI 0.162.1, Grok Build 1.0.50, OpenCode v2.0.24, Cursor Agent 2026.10.01, Pi 1.1.0 con `rpiv-ask-user-question` 2.12.0):
- **Tarjeta que espera la respuesta:** Claude, Grok, OpenCode y Pi con la extensión. En Codex espera solo en modo Plan; en modo normal, `request_user_input_async` vuelve de inmediato y cierra al terminar el turno (src en `rust-v0.162.1`). Según su documentación, Cursor sigue trabajando mientras la pregunta está abierta.
- **Límites por tarjeta:** 4 preguntas en Claude y en Pi con la extensión, 3 en Codex (`request_user_input`), sin límite en Grok y OpenCode. Cursor no tiene campo de descripción ni marca de recomendada.
- **El texto antes de la tarjeta no es fiable:** el modal de Claude tapa la conversación (anthropics/claude-code#89644). Opus 5.x puede volver ese texto razonamiento oculto (anthropics/claude-code#97504). OpenCode puede perderlo (sst/opencode#49708, #52735).
- **Lo que dicen los fabricantes:**
  - Plantilla de Codex en modo normal: «Never write a multiple choice question as a textual assistant message».
  - Prompt de `AskUserQuestion` en Claude 2.1.296: «Use this tool only when you are blocked on a decision that is genuinely the user's to make», y no para «Should I proceed?».

**Referencias externas** (lectura estática, comparación solamente):
- Ninguna esconde contenido detrás de una oferta.
- Patrones portables: preguntar en rondas según dependencias (13 preguntas en unas 3 rondas); recomendada primero; los datos los busca el agente; las decisiones menores se toman y se declaran; una lista corta y cerrada de motivos para detenerse; tarjeta nativa solo si la pregunta cabe completa, y si no, texto.

## Decisiones

| ID | Decisión | Origen |
| --- | --- | --- |
| D1-A | Planear la propuesta P1–P6 | usuario, 2026-10-11 |
| D2-A | Con tarjeta: resumen por pregunta en el texto y opciones solo en la tarjeta | usuario, 2026-10-11 |
| — | Una respuesta puede cambiar la etapa; el agente lo dice y sigue | usuario, 2026-10-11 (mensaje a mitad de turno) |
| — | Codex en modo normal sigue preguntando en texto | orquestador: regla vigente, sin evidencia nueva para cambiarla |
| — | La línea de pausa no cambia | orquestador: fuera de alcance, funciona |
| P1-A | El límite de bytes de `global.md` pasa a ser informativo | usuario, 2026-10-11 |
| — | Rama `feat/guidance-rework` desde `development`, sin commits | usuario, 2026-10-11 (B1 y «no hagas commits») |

Decisión que reemplaza: D2-A del 2026-10-08 (oferta «¿te muestro las recomendaciones?»), registrada en `changes/archive/2026-10-09-gh-9-communication-decision-flow/design.md`.

## Diseño propuesto

### Puntos de decisión por flujo

```
flow-research  fin de ronda ─► ruta: plan / build / seguir / parar
                               (+ ¿abro ticket? en la misma ronda)
flow-plan      ronda de diseño ─► preguntas independientes juntas
               cierre del plan ─► modo de entrega + revisión
               (con las de diseño que queden, si no dependen)
flow-build     solo un bloqueo real o la validación humana
flow-close     siguiente ticket / sesión nueva / terminar
```

No se pregunta fuera de estos puntos. Los datos los averigua el agente. Una decisión menor con un valor convencional la toma el agente y la declara en una línea. Nunca se pregunta solo para pedir permiso de seguir ni para ofrecer mostrar algo.

### Forma de una ronda (D2-A)

```
mensaje                               tarjeta nativa
├─ explicación / hallazgos            ├─ pregunta con su contexto
├─ D1 ¿…? → recomiendo D1-A (razón)   ├─ recomendada primero, marcada
├─ D2 ¿…? → recomiendo D2-B (razón)   ├─ efecto de cada opción
├─ cierres (Pendiente, Recordatorios) └─ más allá del límite:
└─ **Esperando tu decisión:** D1, D2     otra llamada seguida
```

Sin tarjeta que espere, con una tarjeta asíncrona o al repetir una pregunta descartada, las opciones van completas en el texto, una línea cada una.

### Respuesta que cambia el flujo

Después de cada respuesta, incluida una en texto libre o un «Other», el agente dice en una línea qué pasa, antes de actuar:
- **Cierra la pregunta:** sigue.
- **Abre la siguiente ronda:** las preguntas que dependían de esa respuesta.
- **Cambia la etapa:** por ejemplo, una respuesta de diseño que reabre la investigación, o una validación que devuelve la construcción al plan. Dice qué trabajo o respuestas anteriores pierden validez.

Vuelve a preguntar solo si el cambio necesita efectos todavía no autorizados.

```
respuesta ─► ¿qué cambia?
             ├─ nada más ───────► sigue
             ├─ habilita otras ─► siguiente ronda
             └─ cambia la etapa ─► línea: etapa nueva + qué se invalida
                                  └─ ¿efectos nuevos? ─► preguntar
```

### Texto propuesto para `global.md`

El implementador puede ajustar la redacción. Cada regla tiene que nombrar su alcance (AGENTS.md, Guidance structure), sin límite de bytes que bloquee (P1-A). Este texto ya incorpora la revisión del plan (ver tasks.md).

1. **Línea 21, puntos de decisión y contenido** (reemplaza la línea completa):
   «Ask the user only at a decision point: a choice that is genuinely theirs and changes what you do next, such as scope, design, delivery, a route, a validation, or the close. Look up facts yourself. Make a minor call, one with a conventional default that is neither consequential nor an effect needing authorization, yourself and state it in one line. Never ask only for leave to proceed or to show something you could show. At each decision point, ask in one round every open question whose answer does not depend on another open one. Give each question an ID such as `D1`, the context needed to answer, and options with IDs such as `D1-A`, each with its effect, the Recommended option first with its reason; a yes-or-no question needs no option IDs. A consequential question, one whose answer changes scope, contracts, data, or cost, or is hard to reverse, also states the main tradeoff and uncertainty. Never leave a choice's options only in a ticket or document. Reuse explicit decisions already in force; silence, preselection, and a dismissed or timed-out question grant nothing.»
2. **Línea 22, tarjeta y resumen** (reemplaza la línea completa):
   «When the host offers a native question tool that waits for the answer, ask through it, splitting questions or options beyond its limit across consecutive calls, each after a line of text. Write the card to stand alone, since a host may hide the message: the question carries the context to answer, each option its effect in its description, or in its label when the tool has no description field, and the Recommended option is marked. The response that calls it starts, after memory, tracker, and plan updates and with no tool call between, with the message: the explanation, then one line per question with its ID, the question, and the Recommended option with its reason, glossing that option's ID, without the other options; a response holding only the call is the failure. Whenever no waiting card carries the question, such as with no such tool, one that returns before the answer, which you use only while independent work continues, or a question asked again after a dismissal or timeout, which you ask once more in text, write each option in the text, one line each with only what sets it apart; state once what several options share and any route project policy rules out; add no "Other" option or reply example. Take the first answer that arrives. In a question's text and options, state as fact only what you verified in this session and mark the rest as an assumption to confirm.»
3. **Línea 23, después de la respuesta** (se agrega al final de la línea):
   «After each answer, including free text, say in one line before acting whether it settles the question, opens the next round, or returns the work to an earlier flow, such as from plan to research or from build to plan, and what it invalidates; re-ask an answered question only when the change needs effects not yet authorized.»
4. **Línea 24:** «and the possible answers with what each triggers» pasa a «and the possible answers with what each triggers, in the card when one carries the question». Así no choca con D2-A.
5. **Línea 31:**
   - «Investigation findings defer this item to `flow-research`'s offer to show recommendations.» pasa a «Investigation findings use `flow-research`'s route question.».
   - Se borra «label each with the flow it runs, such as "Run with `flow-build`", so the step from research to execution is visible;». Repite «each option naming the flow it starts».
6. **Línea 34:** «use `flow-research`'s offer to show recommendations instead of this close question» pasa a «use `flow-research`'s route question instead of this close question».

**Intención de las líneas reemplazadas y su destino:**
- Los tipos de elección (decisión, ruta, cierre, sí o no) quedan en los puntos de decisión. La «oferta» sale a propósito.
- Opciones con ID y efecto, el contexto para responder, la recomendada con su razón y el costo de las preguntas consecuentes: quedan en la regla 1, para todos los casos.
- «Keep it there even when the options also live in a ticket, a document»: queda como «Never leave a choice's options only in a ticket or document». La parte de la tarjeta se reemplaza por D2-A.
- Una línea por opción en texto, lo que comparten las opciones, la ruta no disponible, y sin «Other» ni ejemplo de respuesta: quedan en la cláusula de texto de la regla 2.
- Que la respuesta con solo la llamada es el fallo, y que ninguna llamada a herramienta separa el mensaje de la tarjeta: quedan abreviadas. Esa frase bajó las tarjetas sin texto de 38 % a 6 %.
- Las opciones que pasan del límite se reparten: quedan, ahora para preguntas y opciones, cada llamada con una línea de texto antes. Esto mantiene `questionAfterDetail` (`tests/pilot/regression.go:103`), que exige texto antes de cada tarjeta.
- La tarjeta asíncrona solo mientras sigue trabajo independiente: queda.
- Lo descartado no cuenta como respuesta y se pregunta una vez más en texto: queda en las reglas 1 y 2. Ahora incluye la tarjeta que se cierra por tiempo, como la de Grok a los 30 minutos.
- La glosa de ID de la línea 18: el resumen glosa el ID de la opción recomendada, porque solo la tarjeta lo define.
- La oferta de una página `flow-report` de la línea 19 no se ve afectada: ofrece crear algo nuevo, no mostrar algo que ya existe.

**Tamaño (medido en bytes sobre `bba2fd48`):**
- Regla 1: +33.
- Regla 2: +273.
- Regla 3: +326.
- Línea 24: +43.
- Línea 31: −148.
- Línea 34: −15.
- Neto: +512, unos 46.548 bytes frente a los 46.040 del presupuesto.

La revisión del plan añadió unos 300 bytes al cerrar huecos para los modelos de ejecución, y el requisito de cambio de etapa del usuario suma 326. **P1-A** (usuario, 2026-10-11): el límite deja de bloquear y pasa a ser informativo. Evidencia, revisada por un investigador sobre las versiones instaladas:
- Ningún host recorta ni ignora el archivo global entre 46 y 50 KB.
- Codex aplica `project_doc_max_bytes` (32 KiB) solo a la cadena del proyecto y lee entero `~/.codex/AGENTS.md`. Se vio en el código de `rust-v0.162.1` y en `codex debug prompt-input`; el usuario lo confirmó.
- Claude Code solo deja de leer un archivo por encima de 4.194.304 bytes. Su aviso de `/doctor` aparece a los 40.000 caracteres en modelos de 200k de contexto, y ese umbral ya se superaba antes de este cambio.

El límite nació en `cba373f6` (2026-09-23) para forzar un crecimiento deliberado, no por un límite de host.

Cambio en `tests/content/budget_test.go`: la prueba informa con `t.Logf` el tamaño en bytes y caracteres y no falla nunca. La constante y su comentario se reemplazan por una explicación de por qué es informativa.

Cambio en `AGENTS.md:24`: «and fits the size budget» pasa a «and states the change in global.md's size in the pull request».

### Skills

- **`flow-research/SKILL.md:30-36`:**
  - La ronda de hallazgos sobre trabajo posible termina con la pregunta de ruta, bajo las reglas globales. Opciones: las alternativas que resuelven los hallazgos, cada una con su flujo (`flow-plan` cuando abarca varios tickets, repositorios o pasos; `flow-build` solo para un cambio pequeño y entendido), seguir investigando y parar en los hallazgos.
  - La pregunta de ticket va en la misma ronda como otra pregunta con ID.
  - Se borran «wait for the offer below» (línea 30), «in the same pause line as the offer below» (línea 32) y «Make the offer once per round» (línea 36). Queda: la pregunta de ruta se hace una vez por ronda y vuelve solo si los hallazgos cambian.
  - Una pregunta que los hallazgos responden por completo, sin trabajo posible, termina sin ruta.
- **`flow-plan/SKILL.md:16`:**
  - Se conserva el bloque «Pausa» para decisiones que cambian el diseño. «the options with their effects, and the recommended one under the global choice rules» pasa a una remisión a las reglas globales, para que con tarjeta las opciones vayan solo en ella.
  - «ask independent design questions together in one exchange…» se reemplaza por una remisión a la regla global de rondas. Se mantiene la frase sobre la pregunta que decide cuáles otras aplican.
  - Cuando no dependen entre sí, las preguntas de entrega y revisión se hacen en la misma ronda que las últimas de diseño.
  - La respuesta que cambia el plan sigue la regla 3.
- **`flow-plan/references/delivery-decisions.md:14`:** con tarjeta, el texto lleva el resumen (D2-A) y la tarjeta lleva la cadena completa de cada modo. Sin tarjeta, las opciones van una línea cada una con lo compartido aparte, como hoy. Las opciones de revisión que pasen del límite se reparten según la regla 2.
- **`flow-build/references/verification.md:65`:** el caso «keep the choice in the message alone and end the turn» lo cubre la regla 2 («whenever no waiting card carries the question»): las opciones van en el texto. Se revisa que la redacción coincida. El recorrido de validación sigue completo en el mensaje, porque es explicación.
- **`flow-close/SKILL.md` paso 7 y `harness-audit/SKILL.md:56`:** la revisión no encontró conflicto. Solo cambia la redacción que dependa de «complete in the message».

### Hosts (`agent-delivery.md`)

Se actualizan la fecha, las versiones y las filas de la tabla con lo observado el 2026-10-11. La regla 2 decide por la herramienta que el agente tiene en la sesión, no por esta tabla. La tabla documenta cómo se aplica D2-A:

| Host | Tarjeta que espera | Dónde van las opciones |
| --- | --- | --- |
| Claude Code, Grok, OpenCode | Sí; Grok la cierra a los 30 min por defecto | Tarjeta; texto con el resumen. Un cierre por tiempo cuenta como descarte |
| Pi | Solo con la extensión `rpiv-ask-user-question` | Con la extensión: tarjeta. Sin ella: texto completo |
| Codex | Solo en modo Plan (`request_user_input`) | Modo Plan: tarjeta. Modo normal: texto completo, sin tarjeta asíncrona para una pregunta que bloquea |
| Cursor | Conflicto: el binario dice que espera por defecto (obs (bin), 2026-10-08); la documentación dice que el agente sigue trabajando ([overview](https://cursor.com/docs/agent/overview), doc, leída el 2026-10-11) | Tarjeta con el efecto en la etiqueta, porque no hay campo de descripción; texto con el resumen |

- **Límites no verificados:** el número de preguntas y opciones de OpenCode, y el esquema de Grok, que viene de una versión anterior a 1.0.50. Quedan marcados así.
- **Codex se aparta de su fabricante:** la plantilla del modo normal pide no escribir opciones como texto. Hive las escribe igual, porque en 29 sesiones ninguna de las 24 tarjetas asíncronas que quedaron abiertas al terminar el turno se contestó por la tarjeta (línea 141).
- **Historia:** la línea 145 («Current rule (change gh-9…)») pasa a historia, marcada como reemplazada por este cambio.

### Medición después del despliegue

Los scripts quedan en `_support/workspace/2026-10-11-question-flow-checkpoints/` (ignorado por Git). Solo devuelven conteos. Se usa el mismo script para la base y para el después:

- **Cortes:**
  - Base de gh-9: del 2026-10-09 00:49Z hasta el despliegue de este cambio.
  - Después: desde la hora del `hive update` que despliegue este cambio.
  - Cada sesión se clasifica por su hora de inicio.
- **Muestra mínima:** 5 sesiones de proyectos reales con al menos un momento de pregunta, por host. Un host con menos se informa como «muestra insuficiente». Las copias de una misma conversación cuentan como una sesión.
- **Métricas:**

| Métrica | Definición y denominador | Umbral después |
| --- | --- | --- |
| M1 ofertas | Mensajes que ofrecen mostrar recomendaciones u opciones; por cada 100 momentos de pregunta | 0 |
| M2 duplicación | Llamadas a tarjeta cuyo texto de la misma respuesta nombra dos o más etiquetas de opción aparte de la recomendada; sobre las llamadas a tarjeta | ≤ 15 % |
| M3 tarjetas sin texto | Igual que `bare_split.py`; sobre las llamadas a tarjeta | ≤ 6 % |
| M4 rondas partidas | Tarjetas de una sola pregunta seguidas de otra a menos de 4 minutos, sin texto del usuario entre ambas; sobre las llamadas a tarjeta | Por debajo de la base recalculada |
| M5 duplicación en Codex | Respuestas con una tarjeta asíncrona y las mismas opciones en el texto; sobre las preguntas de Codex | 0 |

La base de M1, M2, M4 y M5 se recalcula con este script antes del despliegue (T6), porque las cifras del 2026-10-11 salieron de expresiones regulares que no se guardaron.

## Alternativas descartadas

- **Codex: la tarjeta asíncrona como última acción, sin opciones en el texto.** Su mensaje final ya trae las opciones (src), pero en 29 sesiones ninguna de las 24 tarjetas que quedaron pendientes al terminar el turno se contestó por la tarjeta. Sin una sesión real que lo pruebe, se mantiene el texto y no se usa la tarjeta asíncrona para una pregunta que bloquea, lo que evita los dos bloques de `b8c6aeeb`.
- **D2-B (opciones en texto y en tarjeta) y D2-C (solo tarjeta):** descartadas por el usuario.
- **Hook que obligue a escribir texto antes de la tarjeta:** ya se descartó el 2026-10-08 porque `transcript_path` va atrasado.
