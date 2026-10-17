# Diseño

## Contexto verificado

Investigación del 2026-10-16 y 2026-10-17, con tres subagentes `hive-research` y comprobaciones directas. Resumen en Engram #15662 (`research/native-memory-vs-engram`).

**Engram (observado):**
- La 3.4.0 (2026-10-12) es la última versión publicada y es la instalada.
- Desde la 3.1.0, cada sobrescritura por `topic_key` guarda las versiones anteriores, consultables con `mem_get_observation` e `include_history`. No hay herramienta para restaurarlas.
- No hay campo de referencias. La procedencia es solo `session_id` y `tool_name`.
- El protocolo de Engram no dice nada sobre otras memorias.

**Memoria nativa por host:**

| Host (versión) | Memoria propia | Engram |
| --- | --- | --- |
| Claude Code 2.1.302 | Auto memory: la escribe el modelo; `MEMORY.md` se carga siempre (200 líneas o 25 KB). Documentado en code.claude.com/docs/en/memory | MCP + plugin con hooks |
| Grok Build 1.0.62 | memory-v2: captura tras cada turno; el índice se carga una vez por sesión. Documentado como experimental (`~/.grok/docs/user-guide/13-memory.md`), pero activo en las sesiones de cam | Por `compat.claude.mcps`, que por defecto lee la configuración MCP de Claude (`26-config-reference.md:122`). Observado: `engram__mem_save` en la sesión `01a138cb` |
| Codex 0.163.0 | `memories`, encendida el 2026-10-17. Documentado como "generated state", sin edición manual recomendada | MCP `engram` y plugin con hooks, activos desde el 2026-10-17 (`codex mcp list`) |
| OpenCode 2.0.32, Cursor 2026.10.14, Pi 1.1.0 | Ninguna documentada | MCP (OpenCode, Cursor) o paquete `gentle-engram` (Pi) |

**Divergencia (muestra de conveniencia de 33 hechos, del 12 al 16 de octubre):**
- 15 iguales.
- 4 viejos en la memoria nativa y 3 viejos en Engram.
- 3 contradictorios: la reversión de `/register` en cam no llegó a Grok.
- 4 solo en la nativa y 4 solo en Engram.

La memoria nativa ya cita a Engram en varios sitios, con `topic_key`, `#ID` o `sync_id`, sin un formato estable. Engram cita a la nativa solo en prosa.

**Sesiones reales donde se actuó sobre memoria vieja o duplicada:**
- **Grok cam `01a138cb` (2026-10-13) y `01a139b4` (2026-10-14):** la restricción contra `/register` quedó en Engram (#15503, #15514, #15549) y en `cam-whatsapp-meta.md`, `preferences.md` y `cam-production.md`. La sesión del 14 actuó sobre ella. La reversión D5 del 2026-10-16 llegó a Engram y no a Grok hasta la sesión `01a1483c` del usuario.
- **Carvento (Engram #13424, 2026-09-25):** un traspaso entre sesiones salió de memoria vieja y se corrigió al consultar Linear.
- **El patrón "memoria presentada como estado actual"** ya tenía conteo 2 (carvento y rfp) en la auditoría del 2026-10-16.

**Lo que la guía ya dice:**
- `global.md:31`: "search persistent memory for the finding".
- `git-workflow/SKILL.md:18`: "the project's persistent memory, when one is available".
- `flow-close/SKILL.md:24` y `:26`: guardar en memoria persistente.
- `global.md:37`: preservar los datos de memoria.
- `global.md:39`: no migrar memoria.
- `global.md:59`: la memoria recordada es una afirmación sobre el pasado.

Ninguna de esas líneas dice qué memoria es la persistente ni qué hacer cuando dos se contradicen.

## Diseño elegido

R2-A fija el resultado. El texto va en `global.md` porque escribir memoria pasa en cualquier tarea, sin skill activa, igual que la ubicación de artefactos. Va en "Continuity", después del párrafo "On resume…", y no nombra hosts: dice "the host's own memory" y lo describe.

Texto exacto: dos puntos nuevos al final de la lista de "Continuity", en este orden.

```
- Persistent memory in this guidance means Engram when its tools work in this session, and otherwise the host's own memory: the store the host writes for itself and reloads across sessions, not its instruction files. When Engram is configured but fails to connect, authenticate, or write, report it under the integration-failure rule in Communication instead of falling back silently. When you write memory yourself and the host has both, write the project's decisions, restrictions, and work or delivery state to Engram, each under its own `topic_key`, and keep in the host's own memory only what concerns that host and, for a project fact, the pointer `engram:<project>/<topic_key>` instead of a copy. This holds over the host's own memory instructions, except when the user asks to save something in the host's own memory.
- A host's automatic capture may still copy project facts. When two memories, or a memory and a current source, disagree, act on either, and correct an entry, only when a current source settles which one is stale; otherwise report both without editing. Correct a stale Engram entry by updating it, following Engram's conflict prompts when it raises them, and a stale fact in the host's own memory, where the host lets you edit it, by replacing only that fact with the pointer; this rule authorizes that edit. Change only the entries the task touches.
```

Tras la re-revisión, el texto se partió en dos puntos (N1) y se quitó "this is not a memory migration", porque `global.md:39` ya lo dice.

**Decisiones resueltas sin pregunta:**
- **El puntero es `engram:<project>/<topic_key>`, no el `#ID`.** Una sobrescritura por `topic_key` conserva el tema y cambia el contenido, así que el puntero lleva a la versión vigente. El `topic_key` hace upsert dentro de un mismo proyecto y scope, así que el proyecto va en el puntero. Cada hecho del proyecto lleva su propio `topic_key`, para que dos hechos no se sobrescriban entre sí (N3).
- **La regla gobierna lo que escribe el agente, no la captura automática del host (B1).** Grok captura cada turno en segundo plano (`13-memory.md:126-130`) y Codex genera sus memorias solo. En esos hosts la defensa es la cláusula de contradicción: verificar el estado actual antes de actuar.
- **Solo se corrige con una fuente que resuelva cuál entrada está vieja (B2).** Sin esa fuente, se informan las dos y no se edita nada. En Engram se actualiza la entrada y se siguen sus avisos de conflicto cuando aparecen. `mem_judge` solo salta si `mem_save` devuelve `judgment_required` (re-revisión N4). En la memoria del host se reemplaza solo el hecho y se conserva el resto de la entrada.
- **La regla autoriza esa edición (re-revisión N3).** `global.md:37-38` pide preservar los datos de memoria y autorizar las sobrescrituras irreversibles. La memoria nativa no guarda historial, así que la regla lo dice de forma explícita, como pide `global.md:28` para una regla que autoriza una escritura. Lo único que se pierde es el hecho viejo; el vigente sigue en Engram.
- **Engram "falla" también al escribir (re-revisión N2).** Los casos son una identidad en conflicto (`global.md:39`) o un `ambiguous_project`, no solo un fallo de conexión.
- **La excepción para el usuario se limita a guardar en la memoria del host (re-revisión N5).** Un "recuerda X" sin más va a Engram.
- **Solo se corrigen las entradas que la tarea toca.** Así lo exige `global.md:39`, que prohíbe migrar memoria.
- **"Where the host lets you edit it" cubre a Codex.** Su documentación trata las memorias como estado generado. En ese caso basta con corregir Engram.
- **Tener Engram significa que sus herramientas funcionan en la sesión (N2).** Si está configurado pero falla, se informa con la regla de integraciones caídas de "Communication" y no se pasa en silencio a la memoria del host.
- **Sin Engram, la memoria propia del host es la persistente.** Engram es opcional en Hive (`tool-dependencies.md`), y la regla tiene que valer también en instalaciones sin él. La descripción excluye los archivos de instrucciones como `AGENTS.md` (N1).
- **La regla prevalece sobre las instrucciones de memoria que trae el host, salvo una petición directa del usuario (N4).** Ejemplos: el prompt de auto memory de Claude, que pide guardar hechos de tipo "project", y `/remember` en Grok (`13-memory.md:156-170`).
- **Fuera de alcance: si un hijo delegado puede escribir memoria.** `delegation.md` no lo dice. Queda como límite, sin regla nueva.

**Fila nueva de `tool-dependencies.md:22`:** columna "Relevant hosts" = `Claude, Codex, Cursor, OpenCode, Pi; Grok reads Claude's MCP configuration`. Las demás columnas no cambian.

## Lugares revisados

- **Afectados:** `global.md` (líneas 31, 37, 39, 59, 61, 86-89), `git-workflow/SKILL.md:18` y `flow-close/SKILL.md:24,26`. Con la definición nueva, "persistent memory" en esas líneas significa Engram cuando existe. No se contradicen.
- **Sin cambio:**
  - **Roles:** `hive-review-plan.md:17` y `hive-review-harness.md:16` prohíben editar memoria a esos roles, y la regla nueva no lo cambia.
  - **`questions.md:31` y `delivery-decisions.md:30`.**
  - **`harness-audit`:** sus menciones de memoria son de Claude `/memory` y `/context`.
  - **Sitio, README y llms.txt:** ninguna página describe el comportamiento de la memoria. `site/src/content/docs/index.mdx:86` solo nombra `engram-init-workspace`, y `README.md:155` presenta Engram como opcional.
  - **`tests/pilot/`:** no tiene detectores de memoria.

## Nota para el commit de T1

`feat(guidance): make Engram the persistent memory and keep host memory to pointers (T1)`. El cuerpo nombra:
- Grok cam `01a138cb` (2026-10-13) y `01a139b4` (2026-10-14): la restricción duplicada y la reversión que no llegó a la memoria de Grok.
- Carvento (Engram #13424, 2026-09-25).
- La decisión R2-A del usuario (2026-10-17).
- El crecimiento de `global.md` en bytes.
