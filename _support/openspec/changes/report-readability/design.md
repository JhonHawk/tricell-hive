# Diseño

## Contexto verificado (2026-09-26)

**Guía actual:**
- `content/guidance/global.md:16`: "Lead with the answer or outcome"; etiquetas descriptivas cortas para tickets, issues y PRs enlazados.
- `global.md:17`: un bullet con ~7 reglas; la glosa es su última frase ("When citing an ID from an earlier message, add a short gloss…; IDs defined in the same message need none"). La forma de los reportes: "short labeled paragraphs or a compact list", y "its own short bold label" por área.
- `global.md:20`: decisiones por la pregunta nativa "keeping the option IDs in its labels", sin glosa.
- `global.md:21`: la pregunta nativa "shows only the choice", así que el detalle va antes. Esto justifica que la glosa también vaya dentro de la pregunta.
- `global.md:31`: pregunta de cierre, sin glosa.
- `content/skills/flow-report/SKILL.md:8`: "Keep ordinary answers, status updates, and handoffs in prose." En contexto significa "no en HTML".
- `content/skills/flow-build/SKILL.md:46`: ya pide la lista de tareas nativa para trabajo de varios pasos, con "A retained plan's record, when it exists, still governs".
- `tests/content/budget_test.go:12`: `globalGuidanceBudget = 38015`, igual al tamaño actual (`wc -c`); falla por encima del presupuesto o más de 1024 bytes por debajo.

**Fallos observados (Claude Code, `claude-opus-5-5`, sesiones del 2026-09-23 al 26):**

| Caso | Sesión | Superficie | Texto |
| --- | --- | --- | --- |
| S1–S6 sin glosa en ≥8 preguntas | tricell-hive `876d776d` | Pregunta de continuación y cierre | "S1–S5 está completo y desplegado. ¿Seguimos con P5…?" |
| D9 citado 9 bloques después | ark `a49e7ddc` | Descripción de opción | "El 30 % aplica solo con D9-A…" |
| R2 reutilizado en 4 turnos | tricell-hive `644d495b` / `26ac2ab4` | Pregunta y texto | "R2 quedó aplicado en el diff. ¿Cómo seguimos?" |
| Cumple | tricell-hive `9d7ef3ee` | Pregunta | "¿Cómo seguimos con D2-A (esfuerzo en todos los roles)…?" |

En 163 llamadas a `AskUserQuestion`, 67 bloques citaban un ID de otro bloque; ~1/3 sin glosa, ~1/3 con glosa y el resto dudoso. No se inspeccionaron los otros hosts.

**URLs sueltas:** en 60 sesiones de Claude Code recientes hay URLs fuera de un enlace Markdown en el texto del asistente. En tricell-hive `0f38c529`, 10 líneas de texto del asistente tienen URLs y 5 son enlaces Markdown (conteo, sin leer). Se prefiere esa sesión a la de ark para no arrastrar URLs de clientes.

**Capacidades nativas por host** (versiones instaladas; detalle y fuentes en T3): lista de tareas en Claude Code, Codex (`update_plan`), Grok Build y Cursor, no en OpenCode v2 ni en Pi sin extensión; pregunta nativa en Claude Code, Grok Build y Cursor, experimental en Codex, ausente en OpenCode y Pi; Mermaid en el chat en Pi, Grok Build y Cursor, no en Claude Code, Codex ni OpenCode; enlaces con etiqueta en Codex, Grok Build, Cursor y Pi (documentado o en el binario), en Claude Code (observado por el usuario el 2026-09-26) y en OpenCode inferido del binario, que agrega "(url)" cuando la terminal no negocia hipervínculos.

**Evidencia de diseño de reportes:** BLUF, Minto, GOV.UK y los estudios de lectura de NN/g coinciden en conclusión primero y encabezados que se entienden solos. Las tablas ayudan a comparar elementos por atributos (ensayo con "fact boxes", PMC7137953) y estorban en comparaciones simples. El análisis de estilo de LMSYS (2024-08-28) muestra que la longitud y el Markdown inflan la preferencia sin mejorar la calidad. Las plantillas rígidas se abandonan (guía de postmortem de Google SRE). No hay estudios controlados de reportes de agentes en terminal.

**Modelo de trazas:**
- `traceEvent.Input` (`tests/pilot/trace.go:37-41`, `:329-331`) guarda los argumentos completos de la herramienta.
- En Claude, `Message` es `message.id` (`trace.go:435-442`): una llamada al modelo, no una respuesta completa. Los prompts del usuario guardados como string no producen eventos (`trace.go:372`).
- Los fixtures Claude existentes usan una entrada plana `{"question": …}`. La entrada real de `AskUserQuestion` es `{questions:[{header, multiSelect, question, options:[{label, description}]}]}` (solo se revisaron los nombres de clave en `876d776d`).
- `flows.go:841` aplica cada criterio de `regressionCriteria` a cada caso de flujo; `TestRegressionCriteriaReturnsAllThree` (`regression_test.go:823-837`) fija la cantidad y los nombres.

## Diseño elegido

### Guía (`global.md`)

1. **Glosa (D1-A):** se quita la frase final de `:17` y se agrega un bullet propio justo después:

   > Whenever you cite an ID outside the message that defines it, add a short gloss of what it is, every time, such as "R2 (split the PR)"; give a range one gloss, such as "S1–S6 (close checks)". This applies to message text, the close question, and relayed subagent results. A native question counts as its own message, because it shows only the choice: gloss every ID it cites that was defined outside it, including in the text just before it; option IDs it defines, such as `D1-A`, need none.

2. **Forma de los reportes (D2-A, D3-A):**
   - `:16`: "Give linked tickets, issues, and pull requests a short descriptive label…" pasa a "Write each link as a Markdown link with a short descriptive label rather than a bare URL, including tickets, issues, and pull requests, using known context…". Los requisitos de URL de `:17`, `:21` y `:22` se cumplen porque el enlace lleva la URL.
   - `:17`: "with short labeled paragraphs or a compact list" pasa a "with short labeled sections", y "its own short bold label" pasa a "a short bold label that reads on its own". La apertura sigue en `:16`; no se agrega otra regla de "open with".
   - Bullet nuevo después del de glosa:

   > In chat replies, shape each part by its content: a table only to compare three or more items across two or more attributes, with at most four short columns; a list for parallel items; prose for causes and reasoning. A format that a skill or template prescribes takes precedence. State an item's status in words in the session language, such as done, in progress, blocked, or pending on the user, not emoji. Add no formatting that carries no structure. Put diagrams in a file or published page, not the chat, since several hosts show their source raw. When the user will share or keep a comparison, explainer, or long report, offer a `flow-report` page in one line.

   La lista de tareas nativa no entra: ya vive en `flow-build/SKILL.md:46`.

3. **Presupuesto:** se mide con `wc -c` después de editar; se compensa lo posible recortando redundancias de `:17` y se sube `globalGuidanceBudget` al tamaño exacto, con la justificación en el commit.

### `flow-report`

`SKILL.md:8`: "Keep ordinary answers, status updates, and handoffs in prose." pasa a "Keep ordinary answers, status updates, and handoffs out of HTML; the Communication rules of the global guidance shape chat replies." El resto del skill no cambia.

### `agent-delivery.md`

Sección nueva "Presentation capabilities" después de "Native user questions": tabla por host con cinco columnas de capacidad (lista de tareas, pregunta nativa y vista previa, enlaces con etiqueta, Mermaid en el chat, página publicable), versión instalada, fuente y etiqueta documentado/observado/inferido, fechada 2026-09-26. Explica por qué la guía nombra capacidades y deja los diagramas fuera del chat.

### Criterios deterministas (`tests/pilot/regression.go`)

**`cited_id_glossed`:**
- **Normalización:** antes de buscar IDs se quitan `**` y `__` y se ignoran los spans de código y los bloques cercados (``` y `~~~`).
- **Token:** primero los rangos `([A-Z])(\d+)[–-]([A-Z])(\d+)` con la misma letra en los dos extremos, comprobada en código porque el `regexp` de Go no tiene backreferences; después los IDs sueltos `\b[A-Z]\d{1,2}(?:-[A-Z])?\b`. El guion de un rango no cuenta como glosa del primer ID.
- **Definición:** un ID queda definido por el asistente cuando aparece al inicio de una línea (tras espacios, un marcador de lista `-`, `*`, `N.`, o `|` de celda) en el texto del asistente, o como etiqueta de opción de una pregunta nativa. Solo se revisan IDs definidos así; los que define el usuario o un archivo no se ven.
- **Superficies:** texto del asistente (`Kind=="text"`, `Role=="assistant"`, sin `final`) y eventos `question`, que leen `questions[].question`, `options[].label` y `options[].description`. `header` no se revisa: es una etiqueta de ≤12 caracteres.
- **Cita:** aparición de un ID definido en otro `Message`, cuando el mensaje actual no lo redefine. Un evento `question` es siempre su propio mensaje.
- **Glosado:** la aparición va seguida, en el mismo campo, de espacio opcional y `(`, `—`, `–` o `:`; o es una etiqueta de opción cuya descripción no está vacía; o está en una línea que la define.
- **Resultado:** `fail` con la primera cita sin glosa (evidencia: línea, superficie e ID, sin texto libre); `pass` si todas están glosadas; `not_observed` si no hay citas.
- **Límites declarados:** `Message` es una llamada al modelo, así que una lista definida al principio de una respuesta y citada después en la misma respuesta, tras una herramienta, cuenta como cita; tokens que no son IDs de lista pero aparecen al inicio de línea (`H2`) pueden generar falsos positivos; un "(" que no es glosa pasa.

**`no_bare_url`:**
- **Superficie:** texto del asistente (`Kind=="text"`, sin `final`).
- **Excluye:** URLs dentro de `[…](…)`, `<…>`, código en línea y bloques cercados (``` y `~~~`).
- **Resultado:** `fail` con la primera URL suelta; `pass` si hay URLs y todas están excluidas; `not_observed` si no hay URLs en el texto del asistente.
- **Evidencia:** línea y superficie, nunca la URL (puede llevar tokens firmados).
- **Límite declarado:** no juzga si la etiqueta es descriptiva.

**Registro:**
- Ambos en `regressionCriteria`. Se actualizan `TestRegressionCriteriaReturnsAllThree` (cantidad, nombre y nombres esperados), los doc comments de `regression.go:10` y `flows.go:839`, y el README de fixtures con dos filas.
- Esperado: volver a correr `--assess` sobre corridas existentes puede convertir un caso de flujo que pasaba en `fail`.

**Fixtures (formato Claude que lee `parseTrace`):**
- La entrada de `AskUserQuestion` usa la forma anidada real.
- Se derivan leyendo campo por campo el texto del asistente y las entradas de preguntas, que no son `tool_result`; `provenance.md` lo declara. Nunca se leen `tool_result` ni líneas completas.
- `cited_id_glossed`: el fallo sale de `876d776d` si el asistente definió S1–S6 (primera mención en la línea 563 del texto del asistente; confirmar que es una definición). Si no, se usa el R2 de `644d495b`. El paso sale de `9d7ef3ee`. Variantes: pregunta que cita sin glosa un ID definido en el texto inmediatamente anterior; rango con glosa y sin ella.
- `no_bare_url`: el fallo sale de `0f38c529` con la URL reemplazada por `https://example.com/…`; el paso tiene una URL en cada forma excluida.
- Tests unitarios: "S1–S5 está" falla, "S1–S5 (x)" pasa, `**D1** (x)` pasa y `` `D1` `` se ignora.

## Riesgos acotados

- **Codex:** su prompt de sistema desaconseja tablas salvo que se pidan. La regla de Hive las limita a comparaciones reales en el chat; no está verificado cuál gana. Se observa en el monitoreo de sesiones, sin piloto.
- **OpenCode:** que muestre "etiqueta (url)" sin hipervínculos está inferido del binario, no documentado. Si mostrara solo la etiqueta sin poder abrirla, la URL quedaría oculta; se revisa en la próxima sesión de OpenCode monitoreada.
- **Presupuesto:** los cambios suman ~1 KB al archivo que se carga en todas las sesiones de los seis hosts.
- **Eficacia:** los criterios verifican trazas, no que el modelo cambie su conducta. Esa evidencia sale del monitoreo posterior al despliegue.
