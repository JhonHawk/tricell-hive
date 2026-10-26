# CodeGraph como complemento de Hive

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en [PR #105](https://github.com/JhonHawk/tricell-hive/pull/105), merge `2c296d74`, el 2026-10-23; desplegado en local y nota local retirada |
| Tracker · GitHub Issues | • [#104 — CodeGraph como base de datos de programa, con índice al día](https://github.com/JhonHawk/tricell-hive/issues/104) |
| Git | Automático (D9-B) · rama `feat/gh-104-codegraph-complement` → PR → `/code-review` → merge a `development` · sin parada humana |
| Verificación | `go vet` · `go test` · unittest de skills · `hive doctor` local · `/code-review` · tamaño de `global.md` |
| Siguiente paso | Ninguno; sin promoción a `master` |

## Solicitud

> La premisa es la siguiente: https://dashbit.co/blog/evolving-ai-era analiza este articulo en contraste con hive

> ojo, valida que ya tenemos un par de proyectos usando codegraph (no recuerdo si cam o carvento) es reciente y parece funcionar

> tal vez, valga la pena valida que codegraph este actualizado y meterlo como un complemento relevante (asi como engram)

Decisiones del usuario (2026-10-23):

- D4-A: adoptarlo por decisión explícita, abrir el issue y planificar.
- D5-A: sincronizar ya los índices de cam y carvento (hecho).
- D6-A: complemento manual en el catálogo.
- D7-A: regla de uso en `content/guidance/global.md`, que sustituye la nota local.
- D8-A: regla de `codegraph sync`, sin hooks nuevos.
- D9-B: entrega automática.
- D10-A: `/code-review` de Claude Code.

El issue #104 no tiene criterios de aceptación propios; los de abajo salen de estas decisiones.

## Objetivo

CodeGraph 1.6.2 se usa desde el 2026-10-09 en cam y carvento para saber quién llama a qué, pero vive fuera de Hive. Su instalador escribe un bloque de instrucciones propio y una nota local añade sus límites. Hive no lo ofrece en su catálogo, `hive doctor` no lo ve, y ninguna regla distribuida dice cómo mantener el índice al día. El 2026-10-23 el índice de cam tenía 314 archivos sin indexar, y las relaciones (quién llama, a qué afecta un cambio) salían de un grafo viejo sin avisar.

El cambio añade CodeGraph como complemento manual: catálogo y diagnóstico en el gestor, una regla corta en la guía global sobre cuándo usarlo, cuándo sincronizarlo y qué confirmar, y la documentación que enumera los complementos. Después del merge, la nota local de los tres hosts se retira, porque Hive pasa a cubrirla.

## Alcance

Incluye:
- la entrada `codegraph` en `tooling/providers`, solo manual;
- la fila de CodeGraph en `hive doctor` y en la vista Integraciones del TUI;
- la regla en la sección Evidence de `global.md`;
- `README.md`, `llms.txt`, `SECURITY.md`, `_support/docs/architecture/{tool-dependencies,installer,deployment-manager}.md`, `tooling/providers/README.md` y `CHANGELOG.md` (ninguna página de `site/` ni `CONTRIBUTING.md` enumera los complementos);
- el delta de la spec `versioned-installation`, que de paso corrige su lista de complementos, desfasada respecto al código;
- tras el merge: reconstruir el binario, `hive update` y retirar la nota local de los tres hosts (autorizado por D7-A).

Excluye:
- una receta ejecutable de `codegraph install` (D6-A), porque rompería `SECURITY.md:38` y `tool-dependencies.md:17`;
- hooks de git o de lefthook (D8-A);
- Grok Build y Pi en el catálogo: Grok no es destino del instalador y no está probado; Pi no habla MCP;
- crear índices en repos de usuarios: indexar sigue siendo decisión del usuario.

Restricciones que se mantienen:
- `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan;
- `hive doctor` no ejecuta el binario de CodeGraph ni lee su configuración;
- el bloque `CODEGRAPH_START/END` del instalador sigue intacto tras `hive update`;
- la PR informa el cambio de tamaño de `global.md` y los lugares revisados.

## Criterios de aceptación

- AC1. `providers.Catalog()` incluye `codegraph` al final, con hosts `claude`, `codex`, `cursor` y `opencode`, fuente `github.com/colbymchenry/codegraph` y un motivo manual. *Falso en la base cuando* el catálogo no tiene ninguna entrada `codegraph`.
- AC2. `providers.ValidateStep` acepta un paso `Manual` `codegraph-manual` y rechaza un paso `Planned` para `codegraph`. *Falso en la base cuando* devuelve `unknown provider` para el paso manual.
- AC3. La sección Integrations de `hive doctor` muestra una fila CodeGraph:
  - `detected` con la ruta del programa cuando `codegraph` está en el PATH, sin ejecutarlo;
  - `not detected` cuando no está;
  - `not checked` con un home sintético.
  
  Su siguiente paso remite a la fuente oficial y no contiene comandos de instalación de paquetes (`npm`, `npx`, `brew`, `curl`). *Falso en la base cuando* no existe la fila.
- AC4. La vista Integraciones del TUI muestra la fila CodeGraph y cabe en 80×24 con el número máximo de filas. *Falso en la base cuando* no existe la fila.
- AC5. `content/guidance/global.md` tiene en Evidence una regla que:
  - manda usar un índice de CodeGraph existente antes de la búsqueda de texto;
  - remite a la búsqueda de texto cuando no hay herramienta de CodeGraph disponible;
  - manda sincronizarlo antes de fiarse de las relaciones obtenidas por el comando `codegraph`;
  - exime de sincronizar a los roles de solo lectura o delegados, que avisan de que las relaciones pueden estar viejas;
  - manda confirmar con búsqueda de texto o leyendo el archivo que importa los enlaces adivinados por nombre;
  - prohíbe crear índices.
  
  *Falso en la base cuando* `rg -n -i "code-graph|codegraph" content/guidance/global.md` no devuelve nada.
- AC6. `README.md`, `llms.txt`, `SECURITY.md`, `tool-dependencies.md`, `installer.md`, `deployment-manager.md` y `tooling/providers/README.md` enumeran CodeGraph como complemento manual. *Falso en la base cuando* `rg -il codegraph` sobre esos archivos no devuelve nada.
- AC7. Tras el merge y el despliegue local:
  - la guía global desplegada en Claude Code, Codex y OpenCode contiene la regla de AC5;
  - la sección `## CodeGraph limits (local note)` ya no existe en `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md` ni `~/.config/opencode/AGENTS.md`;
  - el bloque `CODEGRAPH_START/END` sigue presente.
  
  *Falso en la base cuando* la nota local existe en los tres archivos (líneas 130, 150 y 149).

## Entrega

- **Modo:** automático (D9-B, 2026-10-23). Rama `feat/gh-104-codegraph-complement` desde `development`, commits por tarea verificada, push, PR a `development`, `/code-review` de Claude Code sobre el diff completo (D10-A), y merge cuando pasen las pruebas locales y se resuelvan los hallazgos. No hay CI.
- **Change record:** queda fuera de los commits de las tareas; `flow-close` lo archiva y lo versiona una sola vez al cerrar, con un PR pequeño de cierre, sin revisión dedicada, que mergeo yo.
- **Tras el merge:** el marcador `_support/workspace/maintainer.local` existe, así que reconstruyo el binario (`go build -o "$(command -v hive)" ./tooling/cli`), despliego con `hive update --out <plan>` y `hive apply --plan <plan>` (sin terminal no hay confirmación interactiva) y retiro la nota local (T5).
- **Tracker:** #104 pasa a en curso al empezar la build y se cierra al integrarse en `development`.
- **Nada se promociona a `master`:** queda como recordatorio.
