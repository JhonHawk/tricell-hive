# question-format-template — Formatos de preguntas y respuestas como plantillas literales, con su criterio de auditoría

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | none |
| Git | tricell-hive, base `development` en `19f3b7be`, rama `feat/question-format-template` (worktree `.claude/worktrees/question-format-template`) |
| Playbook | none (guía distribuida más un criterio del auditor; ningún playbook cubre ambos) |

## Decisions

2026-11-21
- **D1** — ¿Prohibimos la viñeta delante de la letra de cada opción? No: no se trata de convertir cada variante en una prohibición, basta con decir de forma concisa el formato que se necesita; el usuario preguntó si otras reglas tienen el mismo problema (respuesta libre en la tarjeta).
- **D2 A** — ¿El bloque «▶ ¿Cómo seguimos?» lleva un ID de la serie D? Sí, para poder citar después una pregunta que quedó abierta.
- **D3 A** — ¿Qué letra usa el índice de las preguntas de investigación y plan? Se mantiene `D<n>` (no `Q<n>` de grilling), por continuidad con Engram, los watch y la rúbrica.
- **D4 A** — ¿Se separan las decisiones escritas en texto con `---`? Sí, y cada opción lleva la letra sola, sin viñeta.
- **D5 A** — ¿Cómo se escribe cada formato en la guía? Como plantilla literal en su hoja (un bloque con la forma exacta y una línea por parte sobre qué contiene), no como prosa pieza por pieza ni en una hoja aparte.
- **D6 A** — ¿Llevan ID todos los bloques cuyas letras responde el usuario? Sí: el bloque de análisis, el de siguiente paso y el pie de cierre llevan `▶ D<n> — …`.
- **Dónde A** — El trabajo va en un worktree propio, porque otra sesión escribía en el checkout principal.
- **D7 A** — ¿Qué tan estricto es `card_option_format` con la marca de recomendada? Exige `<letra> - ` en cada opción y acepta `(★)` justo después del guion o el marcador del host «(Recommended)» al final de la etiqueta, solo en la primera opción; se descartó exigir solo `(★)` (fallarían 41 de 52 tarjetas de Codex) y medir solo Claude.
- **D8 A** — ¿Cuánto pueden crecer las hojas que lee cada sesión? Cero neto, sin tolerancia nueva: las plantillas ocupan lo mismo o menos que la prosa que reemplazan en `questions.md`, `report-endings.md` y `global.md`; si no caben, el build se detiene con la cifra medida y pregunta.
- **supuesto** — `request_user_input_async` de Codex no se juzga: no es una tarjeta que espera, su ronda va completa en el texto («Channel»), y sus opciones llegan como texto suelto; se descartó leerlas como etiquetas. Se revierte agregando ese nombre de herramienta al criterio.
- **supuesto** — La marca del host aceptada al final de la etiqueta es solo «(Recommended)», sin distinguir mayúsculas, la que piden las herramientas de Claude Code y Codex; se descartó aceptar «(Recomendado)», que ningún host pide. Se revierte en la expresión.
- **Entrega B** — `automatic`: revisión, push, PR y merge sin detenerse.
- **Revisión A** — `/code-review` nativo de Claude Code sobre el candidato antes del push.
- **supuesto** — Una tarjeta de una sola elección sin ninguna marca de recomendada falla, porque la hoja pide la recomendada primera y marcada; se descartó dejarla pasar. Se revierte en la expresión del criterio.
- **supuesto** — El criterio juzga Claude, Codex y Grok (las tres guardan `options[].label` en la entrada de la tarjeta) y cada tarjeta mostrada, tenga o no resultado; se descartó limitarlo a Claude y Codex. Se revierte filtrando por host.
- **supuesto** — La evidencia del criterio da la línea de la tarjeta y el número de la pregunta en ella, nunca su texto ni su `id`, que escribe el modelo, porque la salida del auditor no lleva texto de mensajes (`TestAuditReportCarriesNoCommandOrMessageText`); se descartó citar la pregunta. Se revierte cambiando la evidencia.
- **supuesto** — `close-report.md` queda fuera: ya usa una plantilla literal para su tabla de hallazgos y su reporte final es una lista de partes, no una forma de respuesta; se descartó reescribirlo. Se revierte agregándolo como tarea.
- **supuesto** — La línea mínima de `content/guidance/global.md` cambia «lettered options, the recommended first marked (★)» por el literal `A - (★) <option>`, con un tamaño neto de cerca de 0 bytes; se descartó dejarla igual. Se revierte restaurando la frase.
- **supuesto** — En el texto, cada línea de una decisión va separada por una línea en blanco, porque el Markdown del terminal une las líneas consecutivas de un párrafo; se descartó usar dos espacios al final de línea, que no se ven y se pierden al editar. Se revierte en la plantilla.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Revisar grill-with-docs, aquí lo que nos interesa es el formato con el que hace las preguntas. Lo hace por rondas, tiene un formato que es fácilmente identificable y sí tiene algo en el aspecto de que ya te da como una preselección […]. Creo que en nuestro caso también aporta mucho el hecho de que demos una alternativa.»
>
> «Aquí por ejemplo, tienes doble indicador […]. Ahora, lo que es cierto es que también perdemos el índice, me gustó más Q1, Q2, etc., pero claro que no aplica a toda pregunta, no son lo mismo las preguntas para el research o plan que las de selección de modo de git o revisión.»
>
> (D1) «No se trata de poner como reglas todas las prohibiciones, basta con decir de forma concisa el formato que se necesita; no sé si tengamos más reglas con este problema.»

Sin ticket.

## Objective

Hoy la guía describe en prosa, pieza por pieza, la ronda de decisión, la etiqueta de la tarjeta, el bloque de respuestas, el de siguiente paso y el pie de cierre. Ninguna de esas hojas tiene una plantilla (0 bloques de código), y la forma se rompe de dos maneras observadas. En el texto, una viñeta aparece delante de la letra («- A - (★) …») en 604 mensajes de 117 sesiones de Claude. En la tarjeta, tras una compactación, las opciones pierden la letra y la ★ pasa al final, en 5 sesiones del 19 al 21 de noviembre (`engram:tricell-hive/watch/card-format-after-compaction`). Además, los bloques de respuestas no tienen ID, así que una pregunta abierta no se puede citar.

Al terminar, cada formato está en su hoja como plantilla literal, los bloques de respuestas llevan `▶ D<n> — …`, y `--audit-session` informa `card_option_format` para medir la próxima recaída de la tarjeta antes de agregar más reglas en prosa.

## Scope and acceptance

Entra: `questions.md` («Format», «Channel»), `report-endings.md` (bloque de análisis, siguiente paso, pie de cierre), la línea mínima de `global.md`, `AGENTS.md` «Guidance structure», el criterio `card_option_format` con sus fixtures, la rúbrica de sesiones, `rule-enforcers.md` y las páginas de `site/` que describen esos bloques. No entra: `close-report.md`, el texto de las preguntas de sesión (siguen con su encabezado fijo, sin número), cambiar `D<n>` por `Q<n>`, ni reglas de reapertura tras una compactación (decisión pendiente del watch).

- AC1 — `go run ./tests/pilot --audit-session <transcript> --host claude` informa `card_option_format` con `fail` cuando una opción de una tarjeta de una sola elección no empieza con `<letra> - `, o la marca de recomendada falta o no está solo en la primera opción, y su evidencia da la línea de la tarjeta y el índice de la pregunta, sin su texto · false today because el criterio no existe en la salida del auditor (corrida del 2026-11-21 sobre `aa27a201`) · Observed in: el auditor sobre el registro real `aa27a201-bc12-44e0-af29-ca245d12f8cc.jsonl` (tarjetas en las líneas 3849, 6616 y 6940)
- AC2 — El mismo criterio da `pass` en una sesión cuyas tarjetas siguen `A - (★) …`, incluidas las preguntas de sesión con encabezado fijo, y en una con tarjetas de opción múltiple con letra y varias marcas; da `not_observed` en una sesión sin tarjetas · false today because el criterio no existe · Observed in: el auditor sobre los registros reales `1140329b-708e-457b-83d3-7a550b97972b.jsonl` (13 tarjetas en formato) y `e8ef9855-1d4b-48f7-b8f5-2879ea4fdc14.jsonl` (3 tarjetas, una de opción múltiple con letras y cuatro marcas) y `7bcc87ae-674d-42c9-b444-80a774e5661d.jsonl` (18 tarjetas) de este proyecto
- AC3 — En Codex, una tarjeta `A · Siempre (Recommended)` o `Aprobar conjunto (Recommended)` falla por la letra con guion, y una con `A - (★) …` o `A - <opción> (Recommended)` pasa · false today because el criterio no existe · Observed in: el auditor sobre los rollouts reales `rollout-2026-11-20T01-26-41-01a1f7d9-b225-7b12-bfa1-7fa2b09db920.jsonl` (falla) y `rollout-2026-11-18T14-09-13-01a1f047-170e-7201-8341-51c978d9de47.jsonl` (`codex-tui`, tres tarjetas `request_user_input` como «A - (★) Automática (Recommended)», pasa)
- AC4 — `content/skills/hive-core/references/questions.md` muestra la ronda en texto y la tarjeta como plantillas literales: cada opción en su línea como `A - (★) <opción>` sin marcador de lista, y `---` entre decisiones en el texto · false today because la hoja no tiene ningún bloque de código ni regla `---` (`grep -c '^```'` da 0) · Observed in: la hoja del repositorio y los paquetes de etapa que `tooling/management` genera con ella (flow-plan, flow-verify stage-exit, unattended-delegation, hive-mode)
- AC5 — `content/skills/hive-core/references/report-endings.md` muestra el bloque de análisis, el bloque de siguiente paso y el pie de cierre como plantillas literales con encabezado `▶ D<n> — …` y respuestas con letra sin marcador de lista, y dice, en «The next-step block», que el ID solo sirve para citar el bloque · false today because la hoja no tiene plantillas y sus encabezados no llevan ID · Observed in: la hoja del repositorio y el paquete de etapa de `hive-mode` que `tooling/management` genera con ella
- AC6 — `AGENTS.md` «Guidance structure» dice que un formato que ve el usuario se escribe como plantilla literal en su hoja y que una plantilla no es un ejemplo suelto · false today because la sección solo dice que los ejemplos se usan para marcar límites · Observed in: `AGENTS.md` del repositorio
- AC7 — La página `site/src/content/docs/flows/hive-mode.md` dice que el bloque de respuestas lleva un ID que se puede citar · false today because la línea 61 solo dice que cada respuesta lleva una letra · Observed in: la página del sitio en el repositorio

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Formato de grilling leído en `mattpocock-skills` en `abeeaf9` (2026-11-19); `pstack` no tiene formato de preguntas; Hive antes del rework (`2bb96478`) tenía el mismo formato en `global.md`.
- Auditoría de 570 sesiones del 1 al 22 de noviembre (Claude 189, Codex 215, Grok 166), 2026-11-21: 1 507 de 1 630 preguntas en tarjeta de Claude con el formato completo; patrón de H1 en `bf678880`, `7ee42122`, `aa27a201` (tras compactación), `2faf431f`, `00c7fa30` (antes de leer la hoja). Codex: 41 de 52 preguntas en tarjeta sin `<letra> - `, 49 con «(Recommended)».
- `analysis_reply_block` y `decision_round_format` ya aceptan opciones sin viñeta (`letterItem`, `optionLine`), y `roundTitleLine` no toma `**▶ D7 — …**` por un título de decisión porque `▶` va antes del ID; `**D7 — …**` sin `▶` sí lo tomaría. `cardAnswersOf` ya lee las etiquetas de tarjeta de Claude y Codex.
- Paquetes de etapa instalados que incluyen `questions.md`: flow-plan 20 894 bytes, flow-verify stage-exit 20 129, unattended-delegation 19 995, hive-mode 15 033 (este también con `report-endings.md`); tope 24 000 bytes y 1 000 líneas (`stagePacketMaxBytes`). Margen más chico: 3 106 bytes en flow-plan. `content/guidance/global.md`: 29 610 de 30 000 bytes.
- `go test -count=1 ./tests/pilot/ -run 'TestAudit'` · ok, 0.25 s · 2026-11-21
- `go test -count=1 ./tests/content/` · ok, 2.2 s · 2026-11-21
- `go test -count=1 ./tooling/management/ -run 'Stage'` · ok, 2.3 s · 2026-11-21
- `go run ./tests/pilot --audit-session <transcript> --host claude|codex` · corrió sobre las 570 sesiones (binario compilado de la base) · 2026-11-21
- `go test -count=1 ./tests/content/ -run 'TestRequiredProcedureReadingFootprint'` con 2 bytes agregados a `questions.md` (revertidos) · FAIL: 195 013 B, no por debajo de 184 489 B + 10 523 B; margen real 1 B · 2026-11-21
- `go vet ./...` · exit 0 · 2026-11-21
- `go test -count=1 ./...` · exit 0, 261 s · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · OK, 124 s · 2026-11-21
- `grep -c '^```' content/skills/hive-core/references/questions.md` · 0 (lo que AC4 cambia) · 2026-11-21

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic`: revisión `/code-review` de Claude Code sobre el candidato, push de `feat/question-format-template`, PR a `development` y merge sin detenerse, salvo que un revisor humano tenga pendiente su aprobación. No autoriza un release ni el refresco de otra instalación que la local del mantenedor, que `AGENTS.md` ya fija tras el merge.

## Review status and progress

2026-11-21 — Plan escrito; revisión de plan pendiente.
2026-11-21 — Revisión en dos dominios (`hive-review-plan` server, 52 herramientas, 375 s; guidance, 38 herramientas, 315 s). Aplicado: evidencia sin texto de la pregunta (AC1, T1, spec); marca de recomendada en un lugar fijo y tarjeta sin marca como falla (spec, T1); opción «Open the viewer» con letra (`viewer-open.md` entra a T2, `AGENTS.md:76` a T4); lugares reales para AC2 y AC3; `Observed in` de AC4 y AC5 en el repositorio y los paquetes generados; plantilla propia del pie de cierre con ID cuando tiene respuestas; aclaración de que el ID de un bloque solo sirve para citarlo; plantilla de tarjeta sin descripciones; frase y anclas fijadas por tests en T3; fila 46 de la rúbrica en T3; observaciones en el mensaje de commit; Grok y tarjetas sin resultado en el criterio; fechas locales. Pendiente: el presupuesto de bytes de `reading_footprint_test` (margen de 1 byte) y de los topes de hoja (551 B en `questions.md`, 874 B en `report-endings.md`), que va al usuario como D8.
2026-11-21 — Relectura de los dos revisores (server 13 herramientas, 95 s; guidance 3 herramientas, 33 s). D8 A respondida: cero neto. Aplicado: AC2 con `e8ef9855` y `7bcc87ae` (las sesiones propuestas antes tenían etiquetas del formato anterior), AC3 con el rollout `01a1f047` de `codex-tui` (el anterior era `codex exec`, sin tarjetas), comprobadas con una lectura propia de sus etiquetas; `request_user_input_async` excluido con fixture; una sola marca de host, «(Recommended)», en el spec MODIFIED y ADDED; fixtures separadas de tarjeta rechazada con error y sin resultado; la `(★)` del pie de cierre solo cuando hay una respuesta recomendada; la regla del ID de un bloque vive solo en `report-endings.md`, con puntero desde la línea de IDs de `questions.md`; `design.md` «Impact» apunta a D8.
