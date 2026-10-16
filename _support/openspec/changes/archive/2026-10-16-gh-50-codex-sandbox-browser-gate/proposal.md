# Compuertas de navegador bajo el sandbox de Codex (#50)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por la [PR #61](https://github.com/JhonHawk/tricell-hive/pull/61) en `development` |
| Tracker · GitHub Issues | • [#50 — hive-review-ux no puede abrir un navegador en Codex con el perfil observe](https://github.com/JhonHawk/tricell-hive/issues/50) |
| Git | automática · rama `fix/gh-50-codex-sandbox-browser-gate` → `development` · sin parada humana |
| Verificación | `go vet ./...` y `go test ./...` · verificador por tarea (T1) · `/code-review` de Claude Code en la PR |
| Siguiente paso | Ninguno: cierre del registro con la PR #61 |

## Petición

> bien, ataquemos #50

(2026-10-16). Criterios del issue leído el 2026-10-16, sección "Acción sugerida":

> Decidir una de estas, o una combinación:
> - dar a `hive-review-ux` un perfil de acceso que permita lanzar el navegador sin permitir editar el código (por ejemplo, escritura solo fuera del repositorio y red local), o usar `verify` para ese rol;
> - o que el rol detecte el sandbox y reporte la revisión como no verificada indicando la causa, para que el hilo principal ofrezca otra vía.
>
> Comprobar antes si Codex ejecuta los servidores MCP fuera del sandbox del rol.

Decisiones del usuario en la misma sesión: D1-A retirar la clave `sandbox_mode` del perfil `observe` de Codex; D2-B entrega automática; D3-A revisión con `/code-review` de Claude Code.

## Objetivo

El issue suponía que el sandbox `read-only` del rol impedía a `hive-review-ux` abrir el navegador en Codex. La investigación mostró que en Codex 0.163.0 el archivo de rol no fija el sandbox del hijo: al lanzar un agente con rol, Codex restaura el perfil de permisos del padre, así que la clave `sandbox_mode` que Hive escribe en los TOML de los ocho roles `observe` no hace nada, y lo que bloquea el navegador es el sandbox de la sesión. Con el sandbox por defecto (`workspace-write` sin red) ningún hijo llega al servidor de desarrollo local, ni `hive-review-ux` ni `hive-verify-change`, y la guía exige esas compuertas sin decir cómo reportar ese bloqueo.

El cambio deja de prometer en el perfil de Codex algo que el host no hace, documenta la limitación con su evidencia y el ajuste que la levanta, y hace que la guía distribuida pida al hijo reportar la compuerta como no superada nombrando el bloqueo del sandbox, sin tocar él mismo los permisos del host. No añade una regla nueva de comportamiento: corrige texto que contradice su propia intención (una compuerta obligatoria inalcanzable) y un contrato de perfiles que describe algo que no ocurre.

## Alcance y aceptación

Incluido: `integrations/agent-profiles.json` y sus tres copias de testdata, el validador `validateAccess` y sus pruebas y golden, `_support/docs/architecture/agent-delivery.md`, `content/skills/flow-build/references/browser-automation.md` y `verification.md`.

Excluido: escribir `~/.codex/config.toml` desde el instalador (la guía del repo prohíbe tocar configuración global sin instrucción y Hive deja los permisos al host); un perfil de permisos de Codex por rol (no existe por rol en 0.163.0); cambios en `content/guidance/global.md` (su tamaño no cambia); regla sobre servidores MCP (no establecido si corren fuera del sandbox; el driver por defecto es la CLI `agent-browser`).

Restricciones: la suite `go test ./...` sigue en verde; los demás hosts conservan su mapeo `observe`; el renderizado de Codex sigue siendo TOML plano con `name`, `description`, `model`, `model_reasoning_effort` y `developer_instructions`.

- AC1. Un rol `observe` renderizado para Codex no contiene la clave `sandbox_mode`. *Falso en la base cuando* `Render(synthetic-observer, codex)` y `render.golden` (líneas 95 y 179) contienen `sandbox_mode = "read-only"`.
- AC2. `ReadProfiles` sigue aceptando `sandbox_mode` (`read-only` o `workspace-write`) en el perfil `observe` de Codex, para que los perfiles congelados en releases anteriores sigan legibles desde `hive models` sin correr `hive update`, y `Render` para Codex omite esa clave aunque el perfil la declare. *Falso en la base cuando* `Render` con un perfil que declara la clave la emite en el TOML. (Redactado así tras la segunda ronda de `/code-review`, que mostró que rechazarla rompía `hive models` con un binario nuevo sobre la release 0.4.0 instalada.)
- AC3. `browser-automation.md` indica al hijo que, cuando el sandbox del host le impide lanzar el navegador (escritura del perfil temporal) o llegar al servidor local (red de loopback), reporte la comprobación como no verificada nombrando el bloqueo tal como el host lo reporta, sin cambiar él mismo el sandbox, y `verification.md` cuenta ese bloqueo entre los casos en que la compuerta se reporta como no superada. *Falso en la base cuando* ninguno de los dos menciona el sandbox del host.
- AC4. `agent-delivery.md` afirma que en Codex 0.163.0 el hijo hereda el sandbox de la sesión, que `observe` no renderiza ninguna clave en Codex, cita la evidencia (fuente y prueba en vivo del 2026-10-16) y nombra el ajuste de sesión que permite las compuertas de navegador. *Falso en la base cuando* la línea 44 dice que Codex ofrece valores de solo lectura para `observe` y que solo un padre con bypass los sustituye.

## Contexto verificado

- Fuente de Codex en la etiqueta `rust-v0.163.0` (publicada 2026-10-15): `codex-rs/core/src/agent/role.rs:37-48` define `AgentRoleOverrides` sin clave de sandbox ni permisos; `codex-rs/core/src/agent/control/spawn.rs:422-449` captura el perfil de permisos, la política de aprobación y el cwd del padre, aplica el rol y los restaura.
- Prueba en vivo 2026-10-16 (Codex CLI 0.163.0, `CODEX_HOME` sombra, sesión `01a143b5-4749-7730-8e4c-425e91451fe0`): rol `probe-ro` con `sandbox_mode = "read-only"`, padre `codex exec --sandbox workspace-write`; el hijo (rollout `01a143b5-5acb-7641-b0f0-f32e9518e225`) registró `sandbox_policy: workspace-write` y escribió `child-wrote.txt`.
- Sondas con `codex sandbox` contra un servidor HTTP local: `read-only` niega temporales, DNS y loopback; `workspace-write` por defecto permite temporales y repo pero niega loopback y DNS; `workspace-write` con `sandbox_workspace_write.network_access = true` permite todo; un perfil `default_permissions` que extiende `:read-only` con `:tmpdir`/`:slash_tmp` en `write` y `network.enabled = true` permite loopback y temporales y niega el repo, pero `default_permissions` solo existe a nivel de sesión.
- Hive no gestiona `~/.codex/config.toml` (`rg config.toml tooling integrations` solo encuentra Grok). `agent-profiles.json:10` y las copias en `integrations/agents/testdata/profiles.json`, `tooling/cli/testdata/agent-profiles.json` y `tooling/management/testdata/agent-profiles.json` declaran `observe: {"sandbox_mode": "read-only"}`; `agents.go:274-278` lo valida; `agents_test.go:169` y `render.golden:95,179` lo esperan.
- La documentación de Codex muestra `sandbox_mode` dentro de un TOML de agente (`pr_explorer`), lo que contradice su propio código; por eso la clave se retira y no se conserva como intención.
- No establecido: si Codex lanza los servidores MCP fuera del sandbox (`rmcp-client/src/rmcp_client.rs` no menciona el sandbox; no concluyente).

## Diseño propuesto

- Perfil de Codex: `observe` queda `{}` como `implement` y `verify`. `validateAccess` conserva la aceptación de `sandbox_mode` (`read-only`, `workspace-write`) como clave heredada, porque `hive models` resuelve contra los bytes de perfiles congelados en la release instalada y un binario nuevo los rompería antes de `hive update`; `Render` salta esa clave para Codex, así que el TOML renderiza los mismos campos menos ella, también desde una release anterior.
- Guía distribuida, con la capacidad primero y sin nombrar hosts: el hijo que no puede lanzar el navegador o alcanzar el servidor local por el sandbox del host reporta la comprobación como no verificada de inmediato, nombra el bloqueo como el host lo reporta y no cambia permisos; el hilo principal lo lleva al usuario como compuerta no superada con el ajuste del host. El dialecto de Codex vive en `agent-delivery.md`.
- `agent-delivery.md`: sustituye la frase sobre Codex de la línea 44 por la limitación verificada, su evidencia y el ajuste de sesión (`[sandbox_workspace_write] network_access = true` o un perfil de `permissions`), y aclara que Hive no escribe `config.toml` de Codex.

## Entrega

Repositorio `tricell-hive`, base `development`, rama `fix/gh-50-codex-sandbox-browser-gate`. Modo automático (D2-B, 2026-10-16): commits por tarea verificada, push, PR a `development`, revisión con `/code-review` de Claude Code (D3-A), merge cuando la suite y la revisión pasen, sin parada humana. El registro de cambio se versiona al cierre en la misma rama antes del merge. Al integrar: cerrar #50 con el resultado y refrescar la instalación local (`maintainer.local`). Sin despliegue externo ni release.

## Estado de revisión y progreso

Base: `64e9251c` (`development`, 2026-10-15). Sin revisión de plan: cambio pequeño con decisiones ya tomadas por el usuario en la sesión.
