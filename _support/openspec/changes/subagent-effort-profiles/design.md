# Diseño

## Contexto verificado (2026-09-25)

- **Formato actual.** `integrations/agents/agents.go` acepta `claude_effort` (`low`…`max`) y lo aplica solo cuando `host == "claude"` (líneas 247–248). El esfuerzo del perfil se escribe como `effort` en Claude, `model_reasoning_effort` en Codex y `thinking` en Pi (líneas 261–269). `ReadProfiles` rechaza un esfuerzo en Grok, OpenCode y Cursor (líneas 157–158). `Version = "1"` (línea 17).
- **Perfiles** (`integrations/agent-profiles.json`). En Claude: ejecución `sonnet` sin esfuerzo, razonamiento `opus`/`high`, inherit `inherit`/`high`. En Codex y Pi: ejecución Terra/`high`, razonamiento Astra/`medium`, inherit `high`.
- **Lo instalado en `~/.claude/agents/`** (Claude Code 2.1.282) confirma que `backend-developer`, `frontend-developer`, `kotlin-multiplatform-developer`, `devops-engineer`, `test-engineer` y `sdd-spec-writer` no tienen `effort`. La [documentación de subagentes de Claude Code](https://code.claude.com/docs/en/sub-agents) dice: "Overrides the session effort level. Default: inherits from session", y que los niveles disponibles dependen del modelo. La sesión por defecto está en `effortLevel: high` en `~/.claude/settings.json`.
- **Codex** (0.157.0 instalado; [documentación de subagentes](https://learn.chatgpt.com/docs/agent-configuration/subagents)). `model_reasoning_effort` del archivo del agente tiene prioridad. En modelos GPT-6 acepta `low`, `medium`, `high`, `xhigh`, `max` y `ultra`, "when supported". La documentación dice "start with `low` for GPT-6 Astra", `medium` para Sol y `high` para Luna, y no menciona GPT-5.6 Terra. El Hive anterior fijaba Astra en `medium` por esa política (`harness/codex/README.md` del clon de referencia), así que la inversión frente a Claude es a propósito y se mantiene.
- **Pi** (0.87.1, pi-subagents 0.67.0 declarado en `~/.pi/agent/settings.json`). El renderizador ya escribe `thinking`, y native-agent-dispatch (historical evidence omitted from public history) observó un subagente en Astra/`medium`. No encontré localmente la documentación de pi-subagents con sus niveles válidos: se verifica en T1.
- **Gestor.** `validateRelease` corre `agents.Validate` sobre cada fuente de agente (`tooling/management/plan.go:298`) antes de comprobar la versión del renderizador (`plan.go:309-310`). `plan install --release` y `validatePlan` (`plan.go:536`) pasan por ahí. Los perfiles quedan congelados dentro del release y entran en su hash (`types.go:96-104`). `plan remove` no renderiza (`plan.go:418-428`), `status` compara con los bytes congelados (`plan.go:639`) y `recover` usa los snapshots del journal (`apply.go:284-340`).

## Diseño elegido

### Dónde vive el esfuerzo (D4-A)

- El esfuerzo por defecto de cada rol es el de su perfil. Se agrega `"effort": "high"` al perfil de ejecución de Claude, el único que hoy no lo fija, y así ningún rol renderizado para Claude depende de la sesión.
- Un rol declara `effort` solo cuando su tarea pide otro nivel en cualquier host. El valor es absoluto y reemplaza al del perfil. Las excepciones dependen de la tarea (consulta de estado, redacción, seguridad), no del modelo, y por eso tienen sentido en todos los hosts.
- Quedan cuatro excepciones: `review-security` `max`, `state-fetcher` `low`, `review-plan` `medium` y `sdd-spec-writer` `medium`. Salen los `claude_effort: high` de `sdd-verify`, `review-ux` y `sdd-explore`, que el perfil nuevo vuelve redundantes.

### Contrato del campo

- `effort`: opcional, una línea, uno de `low`, `medium`, `high`, `xhigh`, `max`. Cualquier otro valor falla la validación.
- `claude_effort` se elimina. Como es un campo desconocido, falla la validación igual que cualquier otro campo no declarado. No queda un alias: las fuentes versionadas se migran en este mismo cambio, y para los releases anteriores vale el límite de la siguiente sección.
- Traducción por host (D5-A). Un nivel que el host no tiene se traduce a su nivel más alto documentado. Según T1, Claude, Codex y Pi aceptan los cinco niveles en su escala, así que hoy la traducción es identidad:

| Nivel del rol | Claude (`effort`) | Codex (`model_reasoning_effort`) | Pi (`thinking`) | Grok, Cursor, OpenCode |
| --- | --- | --- | --- | --- |
| `low`…`max` | igual | igual | igual | se ignora |

- El nivel efectivo es el `effort` del rol si existe; si no, el del perfil. Grok, Cursor y OpenCode no emiten esfuerzo ni con un `effort` de rol: la regla actual de `ReadProfiles` se conserva y el renderizador omite el campo.

### Niveles esperados por rol (A2)

Formato: Claude / Codex / Pi.

| Rol | Perfil | Nivel |
| --- | --- | --- |
| `backend-developer`, `frontend-developer`, `kotlin-multiplatform-developer`, `devops-engineer`, `test-engineer`, `sdd-verify`, `review-ux`, `sdd-explore` | execution | high / high / high |
| `sdd-spec-writer` | execution + `medium` | medium / medium / medium |
| `state-fetcher` | execution + `low` | low / low / low |
| `database-specialist`, `performance-engineer`, `review-code`, `review-harness` | reasoning | high / medium / medium |
| `review-plan` | reasoning + `medium` | medium / medium / medium |
| `solution-architect` (absorbe `cloud-architect`, D8-B), `visual-designer`, `review-refuter` | inherit | high / high / high |
| `review-security` | inherit + `max` | max / max / max |

Cambios frente a lo instalado. Claude: seis roles de ejecución pasan de heredar a `high`, y `review-security` pasa de `high` a `max`. Codex y Pi: `sdd-spec-writer` baja de `high` a `medium`, `state-fetcher` de `high` a `low`, y `review-security` sube de `high` a `max`.

### Compatibilidad y versiones

- `agents.Version` se queda en `"1"`. La revisión del plan mostró que, con las mismas fuentes y perfiles, el código nuevo produce los mismos bytes o rechaza la fuente; no genera una salida distinta sin avisar. Subir la versión solo cambiaría el ID del release.
- Un release anterior, que trae `claude_effort`, falla con el binario nuevo en `plan install --release <hash>` y al aplicar un plan guardado con el binario anterior. El error es `unsupported agent field "claude_effort"`. Para volver a un release anterior hay que usar el binario del commit que lo produjo; se documenta en `deployment-manager.md`.
- Las instalaciones existentes se siguen actualizando, quitando y recuperando: `plan install` sobre el estado actual, `plan remove`, `status` y `recover` no vuelven a renderizar el release anterior. T4 lo comprueba en un home sintético.

### Riesgos acotados

- Un modelo que no soporte `max` (resuelto en T1, 2026-09-25). `review-security` usa el perfil `inherit`, así que su modelo lo decide la sesión, no el renderizador. La traducción de D5-A usa la escala del host, no la del modelo.
  - Claude: [model-config](https://code.claude.com/docs/en/model-config) dice "If you set a level the active model does not support, Claude Code falls back to the highest supported level at or below the one you set". Sonnet 5 y Opus 5.5 soportan los cinco niveles. No queda riesgo.
  - Codex: la documentación no dice qué pasa con un nivel no soportado, y no documenta `max` para `gpt-6-astra`, el modelo de sesión instalado; solo recomienda empezar Astra en `low`. Queda como límite documentado en `agent-delivery.md`, sin capa de respaldo.
  - Pi: pi-subagents 0.67.0 acepta `off`…`max` (`src/shared/model-info.ts:3`) y agrega el nivel al modelo resuelto sin validarlo (`src/runs/shared/child-tool-plan.ts:128-142`). `getSupportedThinkingLevels` no participa al lanzar el subagente. Si el proveedor acepta el nivel se decide fuera de pi-subagents, y no está verificado.
