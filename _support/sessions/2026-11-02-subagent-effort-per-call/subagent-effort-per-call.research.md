# Effort por llamada al delegar subagentes

Fecha de inicio: 2026-11-02 · Estado: pausada, sin ticket. Se retoma para abrir el ticket y planear.

## Pregunta y decisión

¿Puede el hilo principal fijar el effort (nivel de razonamiento) de un subagente **al crearlo**, en cada host de Hive? La respuesta decide si `flow-plan` puede asignar un effort por tarea y si `flow-build` puede pasarlo al delegar, con qué alternativa donde el host no lo permita.

Contexto de origen: en Claude Code el `effort` de la llamada `Agent` gana al `effort:` del archivo del agente. La herramienta solo deja usarlo cuando el usuario, un `CLAUDE.md` o una skill piden un nivel concreto. Un plan aprobado por el usuario que asigne un effort a una tarea cuenta como ese pedido. Hoy ni `flow-plan` ni `flow-build` mencionan el effort; se comprobó con `rg -i '\beffort\b'` sobre las dos skills, sin resultados.

## Lo ya establecido (no se reinvestigó)

`_support/docs/architecture/agent-delivery.md` (sección *Effort*) e `integrations/agent-profiles.json` documentan el effort **por perfil y por rol**: Claude `effort`, Codex `model_reasoning_effort`, Pi `thinking`, OpenCode `#variant` solo con modelo configurado, Grok y Cursor sin efecto. Esta investigación cubre solo el effort **por llamada**.

## Resultado por host

| Host (versión instalada) | ¿Effort al crear? | Mecanismo | Lo que lo frena |
|---|---|---|---|
| Claude Code | Sí | `effort` en la llamada `Agent` | Nada: el valor de la llamada gana al del agente |
| Codex 0.166.0 | Existe, pero no aplica | `reasoning_effort` (y `model`) en `spawn_agent` | El `model_reasoning_effort` del archivo del agente gana, y los agentes de Hive lo traen |
| Pi 1.4.0 (pi-subagents 0.83.0) | Sí | Sufijo en `model`: `provider/id:high` | Pide un modelo concreto; los perfiles de Hive en Pi heredan el modelo de la sesión |
| OpenCode 2.2.5 | Sí | `model: "provider/model#variant"` en la herramienta `subagent` | Pide un modelo concreto; cada modelo ofrece solo ciertas variantes |
| Grok 1.0.73 | No | — | La herramienta `task` acepta `model`, no effort |
| Cursor 2026.10.22 | Solo dentro del nombre del modelo | `model` con el effort incluido (`id[effort=high]` o `gpt-5.3-codex-high`) | No hay campo de effort; que la tarea acepte los corchetes no está probado |

## Hallazgos

### H1 — Codex: el archivo del agente anula el effort de la llamada

- **obs (binario):** `spawn_agent` describe `reasoning_effort` como «Reasoning effort override for the new agent. Omit to inherit the parent effort.» y también acepta `model`. El valor debe estar entre los que admite el modelo; si no, la llamada falla con «Reasoning effort `…` is not supported for model `…`».
- **obs (código, `rust-v0.166.0`):** `agent_child_config.rs` (`prepare_agent_spawn_config`) aplica primero los valores de la llamada y después `apply_spawn_agent_role`, que escribe el `model_reasoning_effort` del rol al final. Por eso el rol gana.
- **obs (instalado):** `~/.codex/agents/hive-build-backend.toml` trae `model_reasoning_effort = "medium"`.
- **obs (código):** en multi-agent v2, los campos dependen de `features.multi_agent_v2.expose_spawn_agent_model_overrides`, activo por defecto (#32749). Los campos se añadieron en #14160 (2026-03-11); la primera versión estable que los trae es `rust-v0.115.0`.
- **inferencia:** para que el effort de la llamada sirva, los agentes de Codex tendrían que no traer effort fijo. Pero entonces, si la llamada no lo pasa, el subagente hereda el effort de la sesión, justo lo que la sección *Effort* de `agent-delivery.md` evita a propósito. Otra opción es un agente distinto por nivel.
- **sin verificar:** que el rol gane no se observó en una ejecución real. Tampoco se comprobó si `fork_turns: "all"` descarta el effort de la llamada.

### H2 — Pi: el nivel va como sufijo del modelo

- **obs (código instalado):** `~/.pi/agent/npm/node_modules/pi-subagents/src/extension/schemas.js:247` dice «Suffix :off/minimal/low/medium/high/xhigh/max overrides agent thinking default.» `src/shared/model-info.js:21` (`resolveEffectiveThinking`) usa primero el sufijo y, si no lo hay, el `thinking` de la cabecera.
- **doc (`docs/models.md`, `docs/tool-reference.md`):** no hay un campo `thinking` al delegar; el nivel va en el sufijo. Cada tarea en paralelo o en cadena acepta su propio `model`. El orden de precedencia es: valor de la llamada, luego rol, luego cabecera, luego modelo de la sesión.
- **inferencia:** `inherit:high` no está documentado para subagentes nativos de Pi, así que habría que nombrar `provider/id`. Quien delega tendría que saber qué modelo está usando.
- **límites del usuario:** si el valor rebasa `subagents.maxThinking` o sale de `subagents.modelScope`, la llamada falla antes de crear el subagente.

### H3 — OpenCode: la variante va en el modelo de la llamada

- **obs (binario `~/.opencode/bin/opencode`):** la herramienta `subagent` (antes `task`) acepta `agent`, `description`, `prompt`, `model`, `sessionID` y `background`. `model` se escribe `providerID/modelID#variant`, y no hay un campo aparte de effort.
- **obs (binario):** precedencia `Q ?? H.model ?? L.model`: primero la llamada, luego el agente, luego la sesión. Una variante que el modelo no ofrece se rechaza: «Variant "x" is not available for "p/m"».
- **inferencia:** sin `model`, el subagente recibe el objeto de modelo de la sesión, variante incluida, así que heredaría la variante activa. No se observó en una ejecución real.
- El redondeo a la variante más cercana ya existe en `Resolve` (`integrations/agents`) para los perfiles. Para un valor por llamada haría falta una regla equivalente que pueda seguir el modelo, o una tabla.

### H4 — Las herramientas avisan al modelo que no fije el modelo

- **obs:** OpenCode: «NEVER set this unless the user explicitly asks for a particular model or variant.» Grok: «ONLY choose an explicit `model` when the user DIRECTLY requests it.» Codex: «Do not set the `model` field unless the user explicitly asks» (en el código, `SPAWN_AGENT_INHERITED_MODEL_GUIDANCE_V2`).
- **inferencia:** un plan aprobado cuenta como ese pedido explícito, pero `flow-build` tendría que decirlo en palabras para que un modelo de ejecución se anime a fijarlo.

### H5 — Grok: no hay effort por llamada; el agente tiene un campo sin documentar

- **obs (binario 1.0.73):** el esquema de `task`/`spawn_subagent` acepta `prompt`, `description`, `run_in_background`, `isolation`, `resume_from`, `cwd` y `model`, sin effort.
- **obs (binario):** `AgentDefinition` lee un campo `effort` (también en 1.0.72). La guía de usuario no lo documenta para agentes; lo documenta solo para skills (`effort`) y personas (`reasoning_effort`).
- **doc:** la guía de usuario dice que la precedencia es «Explicit spawn-time override > Role default > Persona default > Parent session». El primer nivel no tiene un parámetro en la herramienta.
- **contradice:** `_support/docs/architecture/agent-delivery.md:45` («Grok has no effort field»). Corregirlo cuando se trabaje en esto, tras verificar que el campo se aplica.

### H6 — Cursor: el effort va dentro del nombre del modelo

- **obs (`index.js` instalado):** la tarea tiene `TaskV2Params`: `description`, `prompt`, `subagent_type`, `model` (opcional), `name` y `mode`. No hay effort.
- **doc (cursor.com/docs/subagents, 2026-11-02):** «The parent agent can name a model when it launches a subagent.» El effort va entre corchetes en el nombre del modelo, como `claude-opus-5[effort=high]`; `cursor-agent --help` muestra la misma sintaxis. `--list-models` también lista variantes con el effort en el nombre.
- **sin verificar:** que la tarea acepte los corchetes, y qué gana cuando el modelo pedido al delegar choca con el de la cabecera.

## Sin verificar en conjunto

Ninguna conclusión se probó con un subagente real, porque los pilotos están pausados. Antes de comprometerse con un host habría que confirmar:

- En Codex, que el rol gane sobre la llamada.
- En Pi, `inherit` con nivel.
- En OpenCode, la variante heredada de la sesión.
- En Grok, el `effort` del agente.
- En Cursor, los corchetes en el modelo de la tarea.

El visor local reportó un índice incompatible con protocol v3, así que la investigación no se publicó ahí.

## Recomendación al cierre

- El plan asigna un effort por tarea, opcional: solo cuando la tarea necesita un nivel distinto del de su agente, con una línea de justificación.
- `flow-build` lo pasa donde el host lo respeta: Claude Code ya, y Pi y OpenCode con el modelo de la sesión más el nivel.
- Donde no se puede (Codex con agentes que traen effort, Grok y Cursor), usa el valor del agente y lo dice en el reporte.
- Quitar el effort de los agentes de Codex o crear un agente por cada nivel sale caro; se decide en el plan, no aquí.

Es una funcionalidad nueva, así que no necesita que una falla se repita; basta con que el usuario lo decida. Tocaría `flow-plan`, `flow-build`, `agent-delivery.md` y, si cambia el comportamiento descrito, `site/`.

## Para retomar

1. Abrir el ticket en GitHub Issues (`JhonHawk/tricell-hive`) enlazando este archivo.
2. Releer la tabla contra las versiones instaladas en ese momento, porque los parámetros cambian entre versiones.
3. Seguir con `/flow-plan <número>`.

Pendiente aparte, al planear: decidir cómo sabe quien delega cuál es su modelo actual, en Pi y OpenCode, para armar `provider/id:nivel` o `provider/model#variant`.
