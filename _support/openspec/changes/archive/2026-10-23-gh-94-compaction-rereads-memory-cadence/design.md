# Diseño

## Contexto verificado

El mapa de impacto se hizo el 2026-10-20 sobre `ce6392ba`, en la rama de prueba. Es un hecho inspeccionado salvo donde se marca lo contrario.

- **Compactación:** ninguna regla distribuida la menciona. La sección Continuity de `global.md:90` ("On resume, reconcile the record…") trata registros de trabajo, no instrucciones. `global.md:76` dice "do not assume … a reference was read"; `global.md:81` y `delegation.md:35` dicen que los hijos no heredan las lecturas del padre.
- **Hook de Grok:** `hive-read-whole-references` (`integrations/grok/hook.go`, `deployment-manager.md:342-364`) convierte una lectura parcial de una referencia Markdown en lectura completa. Por eso P1 elige qué archivos releer, no qué secciones.
- **Cómo exponen los hosts una compactación:** OpenCode V2 la marca en `http.request`; Claude Code, Codex, Grok, Pi y Cursor no la exponen a la guía (`_support/openspec/changes/archive/2026-10-15-flash-class-guidance/design.md:128-129, 384`). La regla tiene que ser un criterio del agente: notar que su contexto empieza con un resumen.
- **Memoria:**
  - `global.md:91` fija el destino (Engram o memoria del host) y prevalece sobre las instrucciones de memoria del host, solo en el destino.
  - `global.md:92` ordena seguir los avisos de conflicto de Engram.
  - `flow-close/SKILL.md:24` (paso 3) actualiza la memoria cada vez que retira un worktree.
  - `flow-close/SKILL.md:26` (paso 5) guarda cada respuesta sobre un hallazgo.
  - `questions.md:35` pide completar las actualizaciones de memoria antes del mensaje con la tarjeta.
  - Hive no distribuye texto del protocolo de Engram (`doctor_integrations.go:265`). El "PROACTIVE SAVE RULE" y el "mem_session_summary — MANDATORY" vienen de las instrucciones del servidor MCP y de la skill `engram:memory`.
- **Pruebas:** ninguna afirma estos textos. `tests/content/budget_test.go` registra el tamaño de `global.md`, que hoy pesa 31 550 bytes.
- **Sitio:** ninguna página describe la cadencia ni las relecturas. `site/src/content/docs/flows/close.md:45` describe los desenlaces de un hallazgo y no cambia.

## Diseño propuesto

### P1 — Releer tras compactar

**Casa:** `global.md`, sección Proportionality, a continuación de la viñeta de `:76`.

**Texto:** cuando el contexto empieza con un resumen porque el host compactó o resumió la conversación, las referencias leídas antes no están garantizadas. El agente relee solo la referencia cuya condición aplica a su siguiente acción, cuando aplica, y no todas las que leyó antes en la etapa. Aplica al hilo que compactó; un hijo lee lo que nombra su brief. El `SKILL.md` de la etapa en curso se relee cuando sus pasos no están en el resumen, porque sin ellos se pierde el orden del procedimiento.

**Conflictos que se resuelven:**
- Es coherente con `:76`: no asume que lo leído siga disponible, pero tampoco relee por adelantado.
- No exime a `reporting.md` cuando el siguiente mensaje es un informe, porque su condición aplica entonces.
- No pide leer por secciones, porque el hook de Grok fuerza la lectura completa.

### P2 — Ritmo de escritura en memoria

**Casa:** `global.md`, sección Continuity, una viñeta nueva después de `:91`.

**Texto:**
- Se escribe en memoria persistente en cuatro momentos:
  - cuando el usuario toma una decisión;
  - cuando el plan queda listo, con el estado de trabajo y sin copiar el plan;
  - en el cierre de un work item, que es el punto durable;
  - en el resumen de sesión, solo cuando el usuario termina la sesión y la memoria admite resúmenes. Si la sesión termina sin aviso, no se intenta nada.
- Entre esos momentos no se escribe por tarea, hallazgo, arreglo ni descubrimiento.
- Este ritmo prevalece sobre las instrucciones de guardado de las herramientas de memoria, como la regla de un plugin de guardar después de cada decisión (D20-A).

**Puntos de diseño:**
- **Juicios de conflicto:** `:92` no cambia. La API de Engram pide un juicio por candidato después de cada guardado, así que bajan solo si hay menos guardados.
- **Paso 3 de `flow-close`:** "Each time you remove a worktree … update its record" pasa a actualizar los registros de los worktrees retirados en la pasada de memoria del cierre, que ocurre antes del reporte (paso 6).
- **Paso 5 de `flow-close`:** los conteos de recurrencia de hallazgos ya declinados van en esa misma pasada. Las respuestas nuevas (declinar o no volver a mencionar) llegan en la pregunta de cierre (paso 7) y se guardan después de esa respuesta, como escritura de decisión del usuario. Corrección de la revisión del plan (B1).
- **`questions.md:35`:** sigue válido sin cambios, porque una decisión del usuario es uno de los momentos de escritura.
- **Delegación desatendida:** no escribe en Engram (registra en el task record), así que no hay conflicto.

**Hallazgo de la verificación de T2:** `git-workflow/SKILL.md:18` registra cada worktree en memoria apenas se crea, para coordinar sesiones paralelas. Esa escritura se conserva como quinto momento (crear o retirar un worktree), en vez de prohibirla.

**Tamaño:** se estiman entre 700 y 900 bytes para las dos viñetas de `global.md`, con cada regla en una oración y sin ejemplos, salvo el del plugin.

**Riesgo:** el plugin puede seguir ganando, porque ningún host declara precedencia entre canales (`repository-and-distribution.md:89`). Se observa en las próximas sesiones y no se agrega un mecanismo.
