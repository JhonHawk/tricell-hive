# Design

## Approaches

Para cada regla se compararon tres versiones: cómo lo hace la fuente de referencia de flujos y verificación, cómo lo hacía Hive antes de su reescritura de noviembre y el texto actual. En H1, H2 y H3 Hive tenía antes un texto más acotado que perdió en las reescrituras; la referencia es más austera todavía (no repite revisiones de plan, salta la revisión en cambios de solo redacción, no pregunta lo reversible).

- **H1, suite repetida (D3 A elegida):** tabla ruta → check en `AGENTS.md` (que `CONTRIBUTING.md` enlaza junto a la suite, para no tener dos fuentes), definición de «alcance conocido» en `checks.md` con reutilización del pase de la suite del candidato, y criterio de auditoría. Descartadas: solo el criterio (la pérdida sigue) y CI en `development` (efecto externo y espera de CI en cada PR). Eximir el Markdown de toda suite queda descartado: `tests/content` lee el Markdown de `content/`, así que un cambio de texto sí puede romper los tests.
- **H2, segunda revisión (D4 A elegida):** disparador por tipo de archivo del diff de los arreglos, archivos movidos fuera del alcance y criterio de auditoría. Descartadas: volver a «a lo sumo una tras P0/P1» (no distingue texto de código) y solo el criterio.
- **H3, relecturas del plan (D5 A elegida):** disparador de septiembre (contrato, interfaz, criterio, decisión o supuesto, más hecho externo, verificación y delta), en un solo hogar (`change-folder.md`, que lee todo plan, no solo el de varios dominios); una ronda de relectura durante la planificación y lo que sobreviva al usuario en el resumen. Descartadas: solo el tope y dejarlo igual.
- **H7, preguntas de sesión (D6 A, D7 A, D8 A elegidas):** la revisión preferida por el proyecto se anuncia en una línea; `hold` pendiente de un trabajo anterior hace repreguntar la entrega; «In place» se anuncia y sigue.

## Decisions

- **Medición de las relajaciones (D4, D5, D6):** son decisiones del usuario, tomadas el 2026-11-23 sobre la comparación de sesiones reales que «Measurement» pide: las 28 sesiones auditadas del 20 al 22 de noviembre (suite repetida en ≥5 sesiones, tercera revisión en `7319321c`, cinco relecturas en `aa27a201`, 84–95 % de tarjetas con la opción recomendada) y la medida de `9ab2cf5b` sobre 12 sesiones (5 de 7 re-revisiones sin hallazgos). Después de integrar, T7 da los dos criterios con que se compara; se reabre si una segunda revisión omitida o una relectura cortada deja pasar un defecto que llega a `development`.
- **Alcance de D5:** el «ciclo» es la planificación (hasta el resumen del plan); durante el build cada revisión conserva su ronda. A los cinco disparadores de D5 se suman hecho externo, línea de verificación y delta, que ya disparaban relectura y siguen haciéndolo: relee igual o más, no menos.
- **Mecanismo antes que prosa:** H1 y H2 tienen respuesta en archivos y comandos sin juicio, así que T7 añade sus criterios de auditoría en el mismo PR.
- **`question_after_detail`:** la corrección es de guía, con recurrencia cumplida en Opus 5.5 (89 de 132 tarjetas sin texto visible, en muchas sesiones independientes).

## Impact

- Paquetes de etapa que incluyen hojas cambiadas: `flow-build` (checks, 21.000 B), `flow-verify/stage-entry` (checks), `flow-verify/stage-exit` (reviewer-brief, questions, review-findings, 18.238 B), `flow-plan` (questions, delivery-ask, 20.145 B), `hive-mode` (questions), `unattended-delegation`. Tope 23.500 B; cada tarea lo comprueba.
- `review_question_native_option` solo juzga una pregunta de revisión en Codex; con D6 esa pregunta no aparece cuando el proyecto fija el mecanismo, y el criterio queda `not_observed`.
- La regla personal cambia en todos los hosts donde Hive está instalado, por `personal_rules.py deploy`.

## Recovery

Revertir el PR restaura las hojas, `AGENTS.md` y los criterios. La regla personal se restaura editando `~/.config/hive/rules.md` con el texto anterior (queda en el diff que muestra T8) y desplegando otra vez.
