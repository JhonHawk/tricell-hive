# Diseño

## Veredicto de un ticket (D1)

| Enfoque | Qué hace | Costo | Elegido |
| --- | --- | --- | --- |
| Juicio con formato fijo | El agente evalúa los 8 puntos de `create-issue/references/ticket-types.md` «Ready list» y escribe `Listo: no — fallan 2, 5, 6` con lo que falta en cada uno; el auditor verifica que la evaluación cierre con tarjeta o bloque | Sin código de validación nuevo; reusa la lista y los helpers del auditor | Sí (D1 A) |
| Modo para tickets existentes en `check_issue.py` | Revisa criterios, «por definir» y fuera de alcance sin depender de encabezados | Código nuevo, frágil con formatos ajenos | No |
| `check_issue.py` tal cual | Valida encabezados del borrador | Ruido de estructura en tickets que no escribimos | No |

## Dónde vive cada regla

- **Canónico:** `hive-rule-questions` «What to ask» — criterios a–f, la lista de supuestos, su tope y cómo se aceptan.
- **Invariante en `global.md:131`:** «decide lo que tiene una ganadora clara y reversible y lístalo; pregunta solo lo que cambia alcance, una regla de negocio o algo difícil de revertir», sin repetir a–f.
- **`plan-handoff.md` punto 3:** cambia «minor calls» por la lista de supuestos, con puntero a la hoja.
- **`hive-rule-business-record`:** la prueba de `assumption` pasa a «implícita en el ticket, las specs o las palabras del usuario»; `hive-rule-business-rules:22` solo cambia el nombre de la prueba si hace falta.
- **`hive-mode` «Assess a chosen ticket»:** los bloques y su orden; apunta a la hoja de preguntas para a–f.

## Forma de la evaluación de un ticket

```
**Revisión · CAM-20 · no listo**            ← línea de contexto
Resumen                                       ← como hoy: impacto, complejidad, dependencias
Veredicto   Listo: no — fallan 2 (criterios), 5 (fuera de alcance), 6 (límite de solicitudes)
Criterios propuestos   AC1…ACn
Lo que no estaba definido y decidí   ≤5 líneas: decisión · descartado · cómo se revierte
Te pregunto   → tarjeta, solo lo que cumple a–f
Siguiente paso   → bloque ▶ con el comando; en el mismo mensaje si no hay nada que preguntar
```

Cuando hay preguntas, el texto antes de la tarjeta lleva los bloques y la línea de espera; el bloque ▶ va en el mensaje que sigue a las respuestas.

## Auditor

- `ticket_assessment_ending`: tras cargar `hive-mode` (skill o lectura de `hive-mode/SKILL.md`) en un turno cuyo mensaje del usuario nombra una clave de ticket (`[A-Z][A-Z0-9]+-\d+`), cada mensaje del asistente que cierra el turno antes de una tarjeta o un comando de flujo debe terminar con tarjeta, bloque ▶ o comando de flujo; si no, `fail` con su línea. Sin carga de `hive-mode` sobre un ticket, `not_observed`.
- `card_star_share` (informativo): para cada tarjeta contestada, si la respuesta coincide con la etiqueta que lleva `(★)`; reporta `answered` y `star`, sin estado. Claude: `AskUserQuestion` y su resultado; Codex: `request_user_input` y la respuesta `send_user_message_question_reply` o el resultado de la herramienta.
