# Diseño

## Contexto verificado

Leído en `development` fa3e3948 el 2026-10-29.

**Hechos inspeccionados (obligaciones a mitad del trabajo):**

- `content/guidance/global.md`:
  - `:13`: la línea de contexto en "every progress or decision message", con exención para una nota de una oración.
  - `:20`: "a finished task gets one progress sentence".
  - `:29`: el defecto que bloquea se arregla y "report it in one line".
- `content/skills/flow-close/references/reporting.md`:
  - `:9`: las reglas de lenguaje.
  - `:15`: la condición "reports progress" y las exenciones.
  - `:57-62`: los pasos del turno de notificación.
  - `:64-66`: el bloque "in-progress item".
  - `:68-70`: el reporte de cierre.
  - `:112-130`: el ejemplo, con un bloque "In progress".
- `content/skills/flow-close/references/work-item-boundary.md:11`: "one progress sentence".
- `content/skills/flow-research/references/delegation.md`:
  - `:7-11`: el título "Split and announce work" y "Tell the user in one sentence".
  - `:57-59`: la línea antes de un lanzamiento que bloquea.
  - `:67`: la oración por tramo.
  - `:89-91`: el ejemplo que la cita.
- `content/skills/flow-build/SKILL.md`:
  - `:12`: la línea de entrada con las etapas y el ticket elegido.
  - `:38`: "Before progress or completion reports".
  - `:42`: "announce the split… announced before its first edit".
  - `:66`: "in one sentence naming the checks it owes".
- `content/skills/flow-verify/SKILL.md`:
  - `:14`: el puntero "progress reports".
  - `:18`: la línea de entrada.
  - `:65`: la lista del reporte y la regla de no mandarlo aparte.
- `content/skills/flow-build/references/iterate-and-integrate.md`:
  - `:9`: "say so in the progress update".
  - `:15`: la lista repetida.
- `content/skills/flow-research/SKILL.md`:
  - `:26`: el anuncio antes de la primera búsqueda delegada.
  - `:48`: "Report each round in one line".
- `content/skills/flow-deploy/SKILL.md`:
  - `:14`: el puntero.
  - `:18`: "Name the run. Say in one line".
- `content/skills/flow-walkthrough/SKILL.md:24`: "opening with the in-progress item" y "in the progress note".
- `content/skills/hive-core/references/questions.md`:
  - `:88`: "say in one line before acting whether it settles the question".
  - `:90`: "goes in that round's progress note".
- `content/skills/unattended-delegation/SKILL.md`:
  - `:10`: la etiqueta del estado en cada mensaje.
  - `:20`: el reporte final con las decisiones, que apunta a `reporting.md#report-a-running-or-completed-item`.
- `content/skills/flow-close/SKILL.md`:
  - paso 7: "a cleanup line stating what was removed and what was kept".
  - excepciones: "still carries its cleanup label".
- `content/voices/preamble.md:9` y `tests/content/voices_test.go:42` ("plain everyday words").
- Los punteros "Before progress reports" en `flow-plan/SKILL.md:67` y `flow-security/SKILL.md:14`.
- Los enlaces al ancla `#split-and-announce-work` en `flow-build/SKILL.md:42`, `flow-plan/SKILL.md:57`, `flow-research/SKILL.md:26` y `security-advisory/SKILL.md:19`.
- Las páginas del sitio en `site/src/content/docs/flows/`:
  - `build.md:6,20,39,42,43`
  - `verify.md:22`
  - `research.md:52`
  - `plan.md:75`
  - `close.md:9,80`
- `_support/docs/architecture/instruction-resources.md:17`.

**Tamaños (bytes, `wc -c`):**

| Archivo | Bytes |
| --- | --- |
| `global.md` | 19,218 |
| `reporting.md` | 15,860 |
| `questions.md` | 15,984 |
| `work-item-boundary.md` | 2,254 |
| `flow-build/SKILL.md` | 14,248 |

`questions.md` tiene 16 bytes de margen: su edición no puede crecer.

**Evidencia de sesiones:** del 22 al 29 de octubre; el hijo `hive-research` revisó los transcripts. Las citas decisivas se verificaron en esta sesión:

- `4afe6d47` contiene "Sigue sin verificar si producción define `INGRESS_RECOVERY_*`".
- `bc3842f3` contiene "Tenías razón: el resumen anterior decía que no quedaban procesos corriendo, y no era cierto."

**Inferencia:** las paradas a mitad del trabajo vienen de mensajes de solo texto o de tomar una pregunta como pausa, no de las notas de progreso. La guía del 28 de octubre en adelante, 47e615f8 y b227122c, solo corrió en una sesión de Claude, así que el efecto de #269 en los reportes de cierre no está medido.

## Enfoque

**Separar dos reglas que hoy van juntas.** La continuidad se queda en su hogar, `work-item-boundary.md`: el texto va con la siguiente llamada y el turno termina solo en los casos listados. La obligación de informar se quita. A mitad del trabajo, un mensaje es opcional y, si existe, no lleva línea de contexto; las respuestas a una pregunta del usuario siguen siendo cortas.

**La línea de contexto** (`global.md:13` y `reporting.md:15`) cubre solo los mensajes que piden una decisión o una validación, entregan el resultado de una etapa, cierran trabajo o reportan un bloqueo. Desaparecen la palabra "progress" y la lista de exenciones para notas de una oración. La cuarta parte de `unattended-delegation` sigue igual: un mensaje sin línea de contexto lleva el estado en su primera línea.

**Espera (D3-A).** El bloque "in-progress item" se reemplaza por una oración al terminar el turno para esperar: qué sigue corriendo y qué decide su resultado. `delegation.md` pierde la oración por tramo y la línea antes de un lanzamiento que bloquea. Conserva la duración del tramo, el "re-await the same process" y la regla del despertar. El ejemplo de espera de `delegation.md` se acorta para no citar la oración retirada.

**Turno de notificación** (`reporting.md`, pasos 1–6). Los pasos y su numeración no cambian, para no romper "as step 6 sets", la condición de `:64` ni la restricción del paso 6.

- El paso 4 pierde la apertura con el "in-progress item".
- El paso 5 deja de pedir una nota de progreso. Si después del resultado sigue corriendo trabajo y el turno termina para esperarlo, ese turno lleva la oración de D3-A, y nada más. Si no queda nada corriendo, continúa con el paso 6.
- El paso 2 (sin mensaje cuando sale un proceso que paraste a propósito, y el 143 no es una falla) sigue igual.
- La condición de `:64` se reescribe como la oración de D3-A y conserva sus excepciones: el paso 2 y el 5.
- El encabezado `## Report a running or completed item` no cambia, porque lo enlazan `delegation.md:63`, `work-item-boundary.md:11`, `flow-walkthrough:24` y `unattended-delegation:20`.

**Reparto entre subagentes.** El título de `delegation.md` pasa a "Split work" y el ancla pasa a `#split-work`. El cambio de título y sus cuatro enlaces (`flow-build:42`, `flow-plan:57`, `flow-research:26`, `security-advisory:19`) van en la misma tarea, T2. `flow-build:42` y `flow-research:26` pierden el anuncio. El reparto no se reporta; es trabajo interno.

**Entradas de etapa.**

- `flow-build:12` pierde la línea de etapas. El ticket elegido por el agente pasa al reporte de cierre como decisión tomada sin preguntar. `global.md:63` deja de decir que la entrada de `flow-build` "also sets the line that tells the user which stages run".
- `flow-verify:18` pierde su línea de entrada. `flow-verify:33` nombra los chequeos que el usuario pidió saltar solo "in the report", no "at entry".
- `flow-deploy:18` deja de "decir en una línea": identifica la corrida. El candidato, los entornos y los tickets van en la pregunta de autorización del paso 4 cuando la hay, y siempre en el reporte del paso 8.
- `flow-build:66` conserva "hand it the candidate with your next tool call", sin "in one sentence".

**Rondas iterativas.** `flow-research:48` deja de pedir una línea por ronda. La medición queda en el estado de la ruta que fija `loops.md`, y una parada reporta ese estado. `questions.md:90` cambia "goes in that round's progress note" por "is recorded in that round's loop state".

**Información que se mueve al cierre (AC7).** Va entre las decisiones tomadas sin preguntar:

- el defecto que bloqueaba y se arregló sin preguntar (`global.md:29`);
- el ticket elegido (`flow-build:12`).

La ruta del primer render sin parada de validación (`flow-walkthrough:24`) también va en el reporte final.

**Reporte de cierre (D5-A).**

- `reporting.md` tiene un único hogar para la lista de lo entregado: lo que cambió; lo no verificado o no cumplido, con el nivel de verificación solo cuando quedó por debajo del acordado; las fallas o la incertidumbre sin resolver; y el siguiente paso cuando hay uno.
- Esa lista aplica a todo mensaje que entrega el resultado de una etapa: la solicitud de validación, el reporte de cierre y la parada por un bloqueo de verify.
- El reporte de cierre la usa y además lleva: qué se completó, en qué punto y qué excluye; el PR; las decisiones sin preguntar; la limpieza que dejó algo vivo o falló; los incidentes; y los temas abiertos con su destino.
- Los elementos condicionales aparecen solo si hay alguno. Nunca se listan los chequeos que pasaron.
- En `flow-verify:65` e `iterate-and-integrate.md:15`, solo la lista de campos pasa a ser un puntero. La regla de colocación se queda en `flow-verify:65`: va en la solicitud de validación o en el reporte de cierre, "not in a message of its own, which can end the turn first", con su excepción para el bloqueo o el resultado autorizado. La restricción de continuidad la protege.
- El paso 7 de `flow-close` y sus excepciones dejan de pedir una línea de limpieza fija.
- La condición "a work item completes" se reescribe con menos texto, no con más.
- `reporting.md:36` y `:78` dicen "reports" en lugar de "progress reports".

**Lenguaje (D8-A).** De `reporting.md:9` se quitan dos oraciones: la de "Write plain, natural prose… explain a technical term the first time" y la de "The English terms these instructions use…". Las reemplaza "Write in the session language." `content/voices/preamble.md:9` pierde la línea de vocabulario y `voices_test.go` pierde esa ancla. Se conservan:

- "Explain necessary context without assuming prior knowledge", que trata del contenido y no del registro;
- la glosa de IDs;
- las reglas de código en inglés.

**Acuses (D2-A).** `questions.md:88` dice que, cuando una respuesta llega, la línea antes de actuar se escribe solo si la respuesta no elige una opción ofrecida, invalida algo o regresa el trabajo a otra etapa; en ese caso dice qué cambia.

- Se deja de anunciar que una respuesta "abre la siguiente ronda", porque D2-A no conserva ese caso.
- El ejemplo de `:118` ("Your answer settles D1…") se reescribe en la forma "qué cambia".
- La edición de `:90` suma unos 7 bytes, así que la de `:88` y la del ejemplo deben ahorrar al menos eso. El archivo debe quedar en 15,984 bytes o menos.

**Punteros.** "Before progress reports, completion reports…" pasa a "Before reports…" en `flow-build`, `flow-verify`, `flow-deploy`, `flow-plan` y `flow-security`. El puntero de `flow-build:38` y `flow-verify:14` a `work-item-boundary` conserva sus dos disparadores: "before the first message" y "before ending a turn". Así coincide con `work-item-boundary.md:5` y `global.md:20`, y protege el caso que #269 corrigió.

**Alternativa descartada.** Conservar la línea de contexto en las notas de progreso opcionales. Se descarta porque los datos muestran que esa línea era el grueso del volumen (unos 1,400 reportes) y el usuario pidió ligereza.

## Riesgos

- Puede haber más "cómo vamos?" durante esperas largas. D3-A lo mitiga y queda en observación con el resto.
- El efecto en los hosts que terminan el turno con texto, Grok y Cursor, depende de que la regla de continuidad siga intacta. AC1 y las restricciones la protegen.
