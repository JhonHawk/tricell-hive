# Verificador en otro modelo y validación manual en texto

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · PR [#68](https://github.com/JhonHawk/tricell-hive/pull/68) mergeado en `rebuild/harness-engineering` (`e9c4c01`) |
| Tracker · GitHub Issues | Sin issue; hallazgos H1 y H2 de la auditoría del 2026-09-29, resueltos sin abrir ticket por decisión del usuario |
| Git | Automático · PR a `rebuild/harness-engineering`, merge tras T4 · cierre del registro por push directo |
| Verificación | TDD en `integrations/agents` · `rg` por criterio · `go vet` · `go test ./...` · `update --dry-run` |
| Siguiente paso | Ninguno. Fuera: el merge del rebuild a `master` |

## Objetivo

`hive-verify-task` debe verificar con un modelo distinto al que implementó la tarea. Esa era la intención del cambio `per-task-verification` (2026-09-26, decisión D11-A: otra familia), pero ninguna regla general lo exige y la configuración ya no lo cumple: desde `9aed6e5` (2026-09-29) Codex, Pi y OpenCode usan GPT-6.1 Sol para implementar y para verificar; en Claude Code el verificador usa Opus, igual que el hilo principal; en Grok y Cursor todo hereda el modelo de la sesión. Además, los roles de Pi fijan `openai-codex/gpt-6.1-sol`, que no aparece en el catálogo de Pi.

Por otro lado, Grok pide la validación manual con la herramienta de preguntas antes de escribir el recorrido: en sample-project el usuario se quejó tres veces ("No me diste los pasos para la revisión", "solo dame los pasos sin lanzar la pregunta") y en ark volvió a pasar con la regla del 2026-09-29 ya cargada.

Este cambio agrega un perfil de modelo `verifier` por host y una comparación antes de lanzar el verificador, y convierte la petición de validación manual en un mensaje con la pregunta en texto, en todos los hosts.

## Alcance y aceptación

Incluye `integrations/agents` (validación y resolución de perfiles), `integrations/agent-profiles.json`, el rol `hive-verify-task`, `flow-build`, `verification.md`, `global.md` y `_support/docs/architecture/agent-delivery.md`.

Queda fuera: garantizar otra familia en Codex (solo modelos de OpenAI) y Grok (solo xAI), donde basta otro modelo; pilotos de comportamiento (pausados por AGENTS.md).

- AC1. Los perfiles aceptan un cuarto perfil de modelo `verifier` y un rol puede declarar `model_profile: "verifier"`; los perfiles del repositorio lo traen en los seis hosts, y un snapshot de una release anterior sin él sigue siendo legible, fallando solo un rol que lo pida. *Falso en la base cuando* `ReadProfiles` rechaza un host con cuatro perfiles de modelo y `Parse` rechaza `model_profile: "verifier"`.
- AC2. `hive-verify-task` resuelve por host al modelo de la tabla D8-A: Claude `opus`, Codex `gpt-6-astra`, Pi `xai/grok-4.7`, OpenCode `github-copilot/claude-opus-5.5` con su variante de esfuerzo, Grok `grok-4.6`, Cursor `inherit`. *Falso en la base cuando* `Resolve` del archivo real `content/agents/quality/hive-verify-task.md` da el perfil `reasoning`.
- AC3. Antes de lanzar `hive-verify-task`, `flow-build` pide comparar su modelo con el del implementador, tratando un modelo desconocido o heredado de la sesión como igual; si coinciden, lanzarlo con otro modelo mediante la opción de modelo del lanzamiento cuando el host la tiene, o preguntar al usuario cuando no, con una respuesta que vale para todo el build; conservar el caso de un verificador que no puede correr con sus tres respuestas; y anotar cada modelo como observado o como configurado sin observar. *Falso en la base cuando* `flow-build` solo prohíbe el mismo modelo para un sustituto.
- AC4. Una petición de validación manual (el recorrido que se entrega para que el usuario valide un resultado, las rondas de corrección y la parada de `interactive` antes del push) se escribe como mensaje y termina el turno con la pregunta en texto, sin la herramienta de preguntas, en todos los hosts; cualquier otra pregunta de ese mismo turno también va en texto; la pregunta de entrega, la de ruta y la de cierre no cambian. *Falso en la base cuando* `global.md` pide la herramienta de preguntas para cualquier espera con respuesta de opción y `verification.md` no exceptúa la validación.

Restricciones:

- `go vet ./...` y `go test ./...` pasan, incluido el tope de `tests/content` (43 000 bytes para `global.md`); si hay que subirlo, se sube en este cambio con la razón.
- `go run ./tooling/cli update --dry-run --source .` valida el catálogo.
- Una release congelada con tres perfiles sigue legible para `hive models`, la vista de modelos, `plan install --release` y los planes guardados (`tooling/management/models.go:24-26`, `plan.go:300`, `apply.go:78`), igual que el precedente de Cursor opcional (`agent-delivery.md:19`).

## Entrega

- **Modo:** automático, reutilizando D4-A de esta sesión para este repo. Rama `feat/verifier-model-and-validation-handoff`, commits por tarea, push, PR a `rebuild/harness-engineering` y merge cuando pase T4 (el repo no tiene CI).
- **Revisión:** sin revisión dedicada, reutilizando D5-C. Este cambio sí toca código Go del gestor; el usuario puede pedir revisión antes del merge.
- **Registro:** esta carpeta no se commitea en la rama; tras el merge se archiva por push directo a la base (D6-A).
- **Después:** cambia el gestor (`integrations/`), así que primero se recompila el binario (`go build -o "$(command -v hive)" ./tooling/cli`) y luego `hive update` (excepción temporal de AGENTS.md).
- **Punto de cierre:** merge, registro archivado, binario recompilado y despliegue local refrescado. Fuera: el merge del rebuild a `master`.

## Decisiones del usuario (2026-09-29)

- Resolver H1 y H2 ahora, sin abrir issues; ruta `flow-plan`.
- D7-A: perfil de verificador por host y comparación antes de lanzar (se llama `verifier` para no confundirlo con el perfil de acceso `verify`).
- D8-A: la tabla de modelos de AC2, y revisar el modelo de Pi que no está en su catálogo (resuelto en la planificación: Pi acepta un identificador fuera del catálogo con un aviso; no se cambia).
- D9-A: validación manual en texto en todos los hosts.
