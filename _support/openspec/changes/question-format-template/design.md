# Design

## Approaches

1. **Plantilla literal por formato, en su hoja (elegida, D5 A).** Un bloque de código con la forma exacta reemplaza la descripción en prosa de la forma; lo que contiene cada parte queda en una línea. Es lo que hace grilling (`skills/productivity/grilling/SKILL.md`) y lo que ya hace `close-report.md` con su tabla de hallazgos, el peldaño más temprano de reuse-before-writing: un patrón que el proyecto ya usa. Costo: las hojas cambian de forma, y `AGENTS.md` dice que un ejemplo suelto se lee como toda la regla, así que se agrega que una plantilla de formato es el formato, no un ejemplo.
2. **Patrones en línea, sin bloque.** Menos cambio, pero sigue describiendo la forma por piezas, que es lo que dejó pasar la viñeta.
3. **Hoja aparte con las cinco plantillas.** Las reúne, pero cada momento lee una hoja más y separa la forma de la regla que la usa.

## Decisions

- Cada plantilla vive en la hoja que ya es el hogar canónico de su formato: la ronda y la tarjeta en `questions.md`; el bloque de análisis, el de siguiente paso y el pie de cierre en `report-endings.md`. Ninguna otra hoja copia la plantilla. `viewer-open.md` cambia su opción a `<letter> - Open the viewer` y su respuesta en el bloque de siguiente paso pasa a ser una respuesta con letra; `harness-audit/references/local-instructions.md` ya usa `A - (★)` y no cambia.
- Las etiquetas de la plantilla se escriben en inglés con la nota existente de que se escriben en el idioma de la sesión; IDs, letras, `(★)`, `▶` y `---` quedan como están.
- El encabezado de un bloque de respuestas lleva `▶` antes del ID (`**▶ D7 — How do we continue?**`), para que el auditor no lo tome por una ronda de decisión sin `Situation:`.
- El ID de un bloque de respuestas solo permite citarlo: responder a un bloque de siguiente paso, con su letra, su ID o «ok», sigue teniendo el efecto de ese bloque, y no es la respuesta a una pregunta que nunca inicia un flujo. Lo dice «The next-step block» de `report-endings.md`; la línea de IDs de `questions.md` apunta ahí.
- `card_option_format` (D7 A): en cada tarjeta que espera, mostrada con o sin resultado, de Claude, Codex o Grok, cada opción abre con `^[A-Z] - ` y con una letra distinta; una tarjeta que el host rechazó por entrada inválida (InputValidationError) no se mostró y no se juzga; `request_user_input_async` no es una tarjeta que espera y no se juzga. En una de una sola elección, la marca de recomendada es `(★)` justo después de `<letter> - ` o «(Recommended)» al final de la etiqueta, en la primera opción y solo en ella; sin marca, falla. Una de opción múltiple exige solo la letra. La evidencia da la línea de la tarjeta y el número de la pregunta en ella, sin su texto ni su `id`, que escribe el modelo. Una `(★)` citada dentro del nombre de otra opción, como «B - Estricto: solo A - (★)», no cuenta como marca.
- En una tarjeta sin descripción por opción, la etiqueta lleva además el efecto de la opción (`A - (★) <name>: <effect>`), como ya pide «Channel».

**Presupuesto de bytes.** `questions.md` tiene 551 B libres (tope 7 000) y `report-endings.md` 874 B (tope 9 000), y `TestRequiredProcedureReadingFootprint` deja 1 B de margen en la suma de lectura de un plan y un build, que incluye las dos hojas y `global.md`. Las plantillas reemplazan prosa con crecimiento neto cero en las tres hojas (D8 A); si no caben, el build se detiene con la cifra medida.

## Proposed shape

Ronda en texto (`questions.md` «Format»), cada línea separada por una en blanco:

```
**<Stage> · <ticket> · <progress>**

<the report the round needs>

D<n> — <the question?>

Situation: <what happened or was found, what was verified, what to check>

A - (★) <option>

If you choose it: <what the user gets and what it costs>

B - <option>

If you choose it: <…>

I recommend A: <the consequence for the user>; <what stays uncertain>

---

D<n+1> — <the question?>

…

<what you wait for · how many known decisions are still to come>
```

Tarjeta (`questions.md` «Channel»):

```
question:    <Stage> · <ticket> · D<n> — <the question?> Situation: <…>
label:       A - (★) <two to five words>
description: If you choose it: <…>; the starred one ends with the recommendation's reason and what stays uncertain
label:       B - <two to five words>
```

Bloque de análisis (`report-endings.md`):

```
**▶ D<n> — How do we continue?**

A - (★) <reply>: <why it is the better one>

B - <reply>

<what you wait for>
```

Bloque de siguiente paso:

```
---
**▶ D<n> — Next step: <flow> (recommended)**

<the reason in one line>

`<the command with a real argument>`

A - <another reply>

B - <another reply>
```

Pie de cierre (sin comando; cerrar es una respuesta):

```
---
**▶ D<n> — <Close (recommended) | See the findings (recommended) | Continue (recommended) | Closed>**

<the reason in one line, when one is recommended>

A - <reply, with (★) after the dash only when one is recommended>

B - <reply>

<what you wait for · known decisions still to come>
```

`**▶ Closed**` sin ID ni letras solo cuando no queda ninguna respuesta que ofrecer.

Tarjeta sin descripción por opción: `label: A - (★) <name>: <effect>`.

## Impact

- Lectores de la forma: `decision_round_format`, `text_round_without_card`, `analysis_reply_block`, `ticket_assessment_ending`, `prose_choice_without_card` en `tests/pilot/`; sus expresiones ya aceptan la letra sin viñeta y `▶` antes del ID (verificado en el código el 2026-11-22). T3 agrega una fixture que lo fija.
- Paquetes de etapa que incluyen `questions.md` (flow-plan, flow-verify stage-exit, unattended-delegation, hive-mode) y `report-endings.md` (hive-mode): tope 24 000 bytes; margen mínimo 3 106 bytes. Ata menos que la suma de lectura (1 B) y los topes de hoja, así que el límite real es el crecimiento neto cero de D8.
- `content/guidance/global.md` está a 390 bytes de su tope; su cambio es de tamaño neto cercano a 0.
- Spec `workflow-lifecycle`: se modifica «Decisions carry labelled context» y se agrega el requisito del criterio.

## Recovery

Revertir el PR. No hay datos guardados ni efectos externos; la guía instalada vuelve con el siguiente refresco local.
