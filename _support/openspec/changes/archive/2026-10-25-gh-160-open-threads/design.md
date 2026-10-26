# Diseño

## Contexto verificado (base `e438e14b`, 2026-10-25)

- `flow-plan/SKILL.md:16` es la única regla sobre temas añadidos («material expansion beyond that outcome»), y solo rige al planear.
- `global.md:32` ya da destinos a los hallazgos incidentales; esta regla no los cubre.
- `questions.md:15` exige revisar una respuesta vigente antes de preguntar; la regla nueva la reutiliza.
- `flow-close/SKILL.md:36` omite la pregunta de cierre en un cambio mecánico sin pendientes.
- `global.md` pesa 36 990 bytes; `tests/content/budget_test.go` solo registra el tamaño.

## Regla (hogar: `global.md`, «Communication», E8-A)

- **Alcance:** un tema que el usuario agrega mientras hay trabajo abierto, en investigación, plan o build. La respuesta a una pregunta pendiente no es un tema nuevo. Un hijo no aplica la regla. Los hallazgos incidentales siguen su propia regla (`global.md:32`).
- **Clasificar en una línea:** aclara un tema abierto (actualiza el trabajo en curso), abre uno relacionado (se resuelve como alcance antes del trabajo que depende de él, preguntando solo cuando ninguna respuesta vigente lo resuelve, según `questions.md`) o es aparte (espera su destino).
- **Destinos:** ticket, plan o tarea; descartado con el motivo que se le dice al usuario; en vigilancia con la condición que lo reabre, registrada en el plan o el traspaso. Ningún tema se cierra sin destino.
- **Cobertura:** cada reporte de hallazgos, pregunta de ruta o de cierre, traspaso que escribe el agente y primer mensaje tras un resumen de compactación nombra cada tema abierto con su destino propuesto. No hace falta una opción de tarjeta por tema.
- **Cierre:** un tema abierto sin destino cuenta como pendiente, así que la excepción de `flow-close/SKILL.md:36` no lo salta.
- **Presupuesto:** unos 700 bytes en `global.md`.

## Clasificación según AGENTS.md, «Measurement»

Es un hueco: ninguna regla trata los temas añadidos fuera del plan. Lo adopta la decisión del usuario D8-A del 2026-10-25.
