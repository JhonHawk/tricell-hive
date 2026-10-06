# Impulsar el code mode

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado el 2026-10-02 en `development` por el [PR #101](https://github.com/JhonHawk/tricell-hive-private/pull/101) (`edcaa72`); desplegado con `hive update` (release `c1c15fcc0b75`) |
| Tracker · GitHub Issues | Sin issue: el cambio nace de una conversación del 2026-10-02 |
| Git | `automatic` a `development` · sin parada humana · sin revisión dedicada |
| Verificación | `go vet ./...` · `go test ./...` · `hive doctor` local · `hive update` y lectura del bloque desplegado |
| Siguiente paso | Ninguno en este cambio; queda observar sesiones reales de OpenCode y Pi (D2-A) |

## Objetivo

OpenCode V2 y Pi 1.0 traen "code mode": una herramienta que ejecuta un script, y desde ese script el modelo llama a otras herramientas. Al modelo solo le llega lo que el script devuelve. Sirve cuando hay que encadenar muchas llamadas o filtrar resultados grandes, como leer decenas de tickets de Linear para un informe. Sin code mode, cada resultado entra entero a la conversación y se vuelve a leer durante el resto de la sesión.

El usuario quiere que Hive impulse su uso. Hoy ningún archivo de Hive lo menciona, y en Pi viene apagado salvo que un servidor MCP lo active.

Este cambio hace dos cosas:

- **Regla en la guía global:** cuando el host ofrezca una herramienta de este tipo, el modelo la usa para encadenar tres o más llamadas o filtrar resultados antes de que entren al contexto, y mantiene directas las llamadas sueltas. Se redacta por capacidad, así que sirve en cualquier host que la tenga y no estorba en los que no.
- **Aviso en `hive doctor`:** cuando Pi está registrado en Hive, una fila "Pi codemode" en Integrations explica cuándo lo activa Pi solo y da la línea de `settings.json` para tenerlo siempre encendido. No lee la configuración de Pi.

## Decisiones del usuario (2026-10-02)

- **D1-A:** alcance P1 + P2. P1 es que Pi lo active; P2, la regla de guía. P3, darlo a los subagentes de Pi, queda fuera hasta comprobar que un script no alcanza las herramientas que un rol excluye (`write`, `edit` en el perfil `observe`).
- **D2-A:** sin prueba con modelos. Se observan las próximas sesiones reales de OpenCode y Pi.
- **D3-B:** Hive avisa en vez de escribir la configuración de Pi.
- **D4-A:** el aviso no lee la configuración del host. Así lo exige la regla de diseño de `hive doctor`, y una detección leyendo solo `~/.pi/agent/settings.json` daría falsos "apagado", porque Pi también lo activa por MCP, por la configuración del proyecto o por `--tools`.
- **Entrega:** `automatic` y sin revisión dedicada.

## Alcance

Incluye:

- la fila nueva de `hive doctor` y de la vista Integrations de la TUI, con sus pruebas;
- la documentación del gestor y el delta de especificación;
- la regla en `content/guidance/global.md`.

No incluye:

- escribir la configuración de Pi o de OpenCode (D3-B);
- dar el code mode a los subagentes de Pi (P3);
- cambiar la exposición MCP del usuario. Su `~/.pi/agent/mcp.json` ya deja que los scripts lleguen a las herramientas diferidas (`docs/mcp.md:204` de Pi 1.0.0).

Cambios en el contrato de `hive doctor`: con Pi registrado, la vista Integrations pasa de cuatro a cinco filas. El texto de `hive doctor` cambia, pero ningún consumidor fuera de este repositorio lo analiza; `hive doctor` y la vista comparten el texto.

## Criterios de aceptación

- **AC1.** Con Pi registrado, la sección Integrations de `hive doctor` imprime una fila `Pi codemode` con hallazgo `not checked` y registro `not part of onboarding`. Su fuente es la documentación de la CLI de Pi, y su siguiente paso contiene `"defaultTools": ["+codemode"]` y dice que Pi lo activa solo con servidores MCP de exposición `codemode`. *Falso en la base si* `hive doctor` no imprime esa fila.
- **AC2.** Con Pi registrado, la vista Integrations de la TUI lista cinco filas en este orden: Engram, Context7, pi-subagents, Pi codemode, agent-browser. El detalle de la fila nueva coincide con el texto de `hive doctor`. Sin Pi registrado, lista las cuatro filas de hoy. *Falso en la base si* con Pi registrado la vista lista cuatro filas.
- **AC3.** La fila `Pi codemode` no cambia según la configuración de Pi. Con el fixture no sintético `integrationsFixture`, cuyo `HOME` es la carpeta de prueba, y un `<home>/.pi/agent/settings.json` que ya declara `"defaultTools": ["+codemode"]`, la fila sigue en `not checked`. *Falso en la base si* la fila no existe. Límite: solo detecta una implementación que lea ese archivo y además cambie `Found`; la garantía de fondo es que la fila se arma solo con constantes (design.md).
- **AC4.** `content/guidance/global.md` contiene, en Proportionality, la regla de code mode de [design.md](design.md#regla-de-guía). *Falso en la base si* `rg -n "runs a script able to call other tools" content/guidance/global.md` no encuentra nada.

Restricciones que deben seguir cumpliéndose:

- `go vet ./...` y `go test ./...` pasan;
- `hive doctor` sigue sin escribir archivos ni ejecutar programas fuera de los ya permitidos;
- la sección Integrations sigue buscando en `PATH` solo `engram` y `agent-browser` (prueba existente, `doctor_integrations_test.go:146`).

## Entrega

- **Repositorio:** `tricell-hive`, base `development`, rama propuesta `feat/encourage-code-mode`.
- **Modo:** `automatic`, elegido por el usuario el 2026-10-02. Rama, commits, push, PR a `development` y merge cuando pasen `go vet ./...` y `go test ./...`. El repositorio no tiene CI (la retiró el 2026-09-30), así que las pruebas locales son la condición del merge. El merge no dispara despliegues.
- **Revisión:** sin revisión dedicada. Basta la inspección del diff por el implementador y las pruebas.
- **Después del merge:** actualizar `development` en el checkout principal, recompilar el binario (`go build -o "$(command -v hive)" ./tooling/cli`) porque cambia `tooling/`, y ejecutar `hive update`, como pide la excepción temporal de `AGENTS.md`.
- **Registro del cambio:** se versiona una vez, al cierre, con `flow-close`, en un PR pequeño a `development` sin revisión dedicada, que fusiona el agente.
