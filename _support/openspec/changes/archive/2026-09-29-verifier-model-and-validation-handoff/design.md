# Diseño

## Contexto verificado

Base: `f348588` en `rebuild/harness-engineering`, leída el 2026-09-29. Investigación delegada a `hive-research` el mismo día; sus citas se comprobaron donde deciden el diseño.

- [`integrations/agents/agents.go`](../../../../integrations/agents/agents.go) `Parse` (línea 101) acepta solo `execution`, `reasoning` e `inherit` como `model_profile`; `ReadProfiles` (líneas 145–160) exige exactamente tres perfiles de modelo por host y los valida por nombre; `Resolve` (líneas 255–279) toma `h.Models[r.ModelProfile]`, aplica el esfuerzo del rol y, en OpenCode, lo codifica como variante `#esfuerzo`; `Render` emite `model` cuando el perfil lo trae.
- [`integrations/agent-profiles.json`](../../../../integrations/agent-profiles.json): Claude `execution` sonnet y `reasoning` opus; Codex y Pi `gpt-6.1-sol` en ambos; OpenCode `github-copilot/gpt-6.1-sol` en ambos e `inherit` DeepSeek; Grok sin modelos; Cursor `inherit` en todos.
- `hive-verify-task` usa `reasoning`; `hive-build-data` también; el resto de implementadores usa `execution`.
- [`flow-build/SKILL.md`](../../../../content/skills/flow-build/SKILL.md) paso 5 del ciclo solo prohíbe un sustituto en el mismo modelo que el implementador. [`agent-delivery.md`](../../../../_support/docs/architecture/agent-delivery.md) línea 17 describe la regla y el caso de Grok y Cursor.
- Opción de modelo al lanzar un subagente: Claude Code 2.1.285, parámetro `model` del Agent tool, que manda sobre el frontmatter (docs de subagentes, ctx7 `/websites/code_claude`) [verificado]; OpenCode 2.0.19, parámetro `model` de la herramienta `subagent` [verificado en el binario]; Codex 0.159.1, campo `model` de `spawn_agent`, cuya descripción pide no usarlo salvo petición del usuario [verificado]; Pi 0.87.1 con pi-subagents 0.67.0, "per-run override" [según su documentación]; Grok 1.0.45, argumento `model` de `spawn_subagent`, oculto cuando todos los modelos son de xAI y la herencia está activa [según su documentación]; Cursor 2026.09.28, parámetro `model` de Task observado en una sesión de ark (`claude-sonnet-5-5-high`), sin documentación.
- `pi --list-models` (2026-09-29) lista en `openai-codex` `gpt-6-sol`, `gpt-6-astra`, `gpt-6-luna` y `gpt-5.6-*`, pero no `gpt-6.1-sol`, que fijan los perfiles `execution` y `reasoning` de Pi. No se comprobó si Pi lo resuelve igual en tiempo de ejecución.
- Grok en sample-project (`01a0e9b9`, 2026-09-29): tres preguntas de validación sin pasos previos, dos después de la queja del usuario. Grok en ark (`01a0ee0e`, ARK-680): pregunta de validación sin mensaje previo, con `c323789` ya cargado.
- `tests/content/budget_test.go`: `global.md` tiene un tope de 43 000 bytes con 1 024 de holgura hacia abajo; hoy mide 42 988.

## Diseño elegido

### Perfil `verifier`

Un cuarto perfil de modelo, `verifier`, en cada host de `agent-profiles.json`, agregado después de `inherit` (la prueba `TestCursorObserveRendersInheritModelAndReadonly` reemplaza una subcadena exacta de la línea de Cursor, `agents_test.go:157`). `hive-verify-task` pasa a `model_profile: "verifier"`. El nombre evita confundirlo con el perfil de acceso `verify`, que el mismo rol ya usa.

| Host | Modelo | Esfuerzo | Evidencia de catálogo (2026-09-29) |
| --- | --- | --- | --- |
| Claude | `opus` | `high` | alias de Claude Code |
| Codex | `gpt-6-astra` | `high` | `models_cache.json` de Codex |
| Pi | `xai/grok-4.7` | `high` | `pi --list-models` lista `xai grok-4.7` |
| OpenCode | `github-copilot/claude-opus-5.5` | `high` (variante) | `opencode models` lista el modelo; la variante `high` no se comprobó |
| Grok | `grok-4.6` | sin esfuerzo (Grok no lo acepta en perfil) | catálogo de Grok |
| Cursor | `inherit` | sin esfuerzo | la comparación elige el modelo al lanzar |

Contrato (hallazgo B1 de la revisión del gestor, reconciliado con su opción (a), que sigue el precedente de Cursor opcional):

- `Parse` acepta `verifier` en `model_profile`.
- `ReadProfiles` acepta tres o cuatro perfiles por host: `verifier` es opcional al leer, con las mismas validaciones que los otros cuando está. Así, un snapshot de una release anterior sigue legible para `hive models`, la vista de modelos, `plan install --release` y los planes guardados.
- `Resolve` de un rol que pide `verifier` en un host cuyo perfil no lo trae falla con un error claro que nombra el host y el perfil.
- Una prueba sobre los perfiles del repositorio exige `verifier` en los seis hosts, y otra carga perfiles de tres entradas (modelada sobre la de cinco hosts, `agents_test.go:347-378`).

Claude queda con `opus`, igual que `reasoning`: la mayoría de las tareas las implementan hijos en Sonnet, y cuando implementa el hilo principal en Opus, o `hive-build-data`, la comparación cambia el modelo al lanzar.

Modelo de Pi, resuelto en la planificación: `core/model-resolver.js` (líneas 447–454 de `@earendil-works/pi-coding-agent` instalado) acepta un modelo fuera del catálogo cuando el proveedor sí está, como identificador propio y con el aviso "Using custom model id"; `gpt-6.1-sol` existe en el catálogo de Codex para la misma cuenta. Los perfiles `execution` y `reasoning` de Pi no cambian; `agent-delivery.md` anota esa evidencia.

Riesgo sin verificar: si Grok respeta un `model` en la definición del agente (su documentación dice que sí; `Render` ya lo emite cuando el perfil lo trae, `agents.go:304-306`).

### Comparación antes de lanzar

`flow-build`, ciclo de `hive-verify-task`: el paso 5 actual se reemplaza por esta regla, que conserva su caso y sus respuestas, y el paso 6 sigue aplicando las decisiones del usuario.

- Antes del paso 1, comparar el modelo del implementador (el del hilo principal o el del hijo) con el del verificador. Anotar cada uno como observado (lo informa el host o la herramienta) o como configurado sin observar (sale del perfil).
- Tratar como iguales un modelo desconocido o uno que hereda el de la sesión (`inherit`, o Grok sin modelo).
- Si son iguales, lanzar el verificador con otro modelo disponible mediante la opción de modelo del lanzamiento cuando el host la ofrece. Si el host la desaconseja salvo pedido del usuario, o no la ofrece, preguntar una vez al usuario: usar un modelo que él nombre, o aceptar las tareas sin verificación independiente. Esa respuesta vale para el resto del build.
- Si el verificador no puede correr (modelo no disponible, sin cuota), dejar la tarea en `[?]` y preguntar si esperar y reintentar, verificar con un modelo que el usuario nombre, o aceptarla sin verificar. Nunca sustituirlo por un hijo en el modelo del implementador.

`agent-delivery.md` actualiza la línea 17 y la tabla de perfiles, anota que `verifier` y el perfil de acceso `verify` son distintos, y registra la evidencia de catálogo de la tabla de arriba.

### Validación manual en texto

Alcance: una petición de validación manual es el recorrido que se entrega para que el usuario valide un resultado, las rondas de corrección de `verification.md` y la parada de `interactive` antes del push (`delivery-decisions.md` línea 7, `flow-build` línea 67).

- `global.md`, regla de esperas del usuario: una petición de validación manual se escribe en el mensaje y termina el turno con la pregunta en texto, sin la herramienta de preguntas, en todos los hosts. Cualquier otra pregunta de ese mismo turno, incluidas decisiones con IDs, va en texto en el mismo mensaje. La pregunta de entrega, la de ruta y la de cierre no cambian.
- `verification.md` "Human UI handoff": el recorrido termina con esa pregunta en texto; en la vista temprana de una composición y en las rondas de corrección, "preguntar" es esa pregunta en texto, y el turno termina con ella aunque sigan corriendo hijos, con el aviso de trabajo en curso al inicio.
- `global.md` está en 42 988 bytes con tope 43 000: lo agregado se compensa en la misma regla si se puede; si no, se sube `globalGuidanceBudget` en `tests/content/budget_test.go` en este cambio, con la razón en el commit.
