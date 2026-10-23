# Diseño

## Contexto verificado (2026-10-23)

**CodeGraph 1.6.2.** Es la última versión en npm y GitHub (2026-10-03). Fuente: bundle `~/.codegraph/versions/v1.6.2/lib/dist/`.

- **El servidor MCP mantiene el índice mientras vive su sesión.**
  - Al conectarse se pone al día (`mcp/engine.js:426`, `catchUpSync`) y vigila los archivos (`mcp/project-lifecycle.js:142`).
  - El servicio en segundo plano se cierra 300 s después de que se va el último cliente (`mcp/daemon.js:98`).
  - Ni la línea de comandos (`explore`, `callers`, `impact`) ni el hook `prompt-hook` sincronizan (`bin/codegraph.js:1479`: «this hook never runs» sync).
- **`codegraph status -j`** informa `pendingChanges {added, modified, removed}` e `index.state`. Siempre sale con 0, así que hay que leer el JSON.
- **El bloque del instalador no se puede omitir.**
  - `codegraph install` escribe siempre su bloque `CODEGRAPH_START/END` (`installer/targets/claude.js:185`, `codex.js:137`).
  - `codegraph upgrade` lo reescribe.
  - Su texto dice cuándo usar CodeGraph, pero no habla de límites ni de frescura.
- **Destinos del instalador:** claude, cursor, codex, opencode, hermes, gemini, antigravity, kiro y copilot. Grok Build no lo es (el issue #1347 de CodeGraph sigue abierto). Pi no habla MCP, según el autor.
- **Defecto conocido (Engram #11370):** CodeGraph enlaza por nombre algunos imports de paquetes del workspace y alias que no están en la raíz, sin marcar la suposición.

**Hive.**

- **Catálogo** (`tooling/providers/catalog.go:7-40`).
  - Hoy tiene Engram, Context7 y pi-subagents (manuales), y agent-browser y codex-question-card (con receta).
  - Las entradas nuevas van al final para no mover los índices (comentario en `catalog.go:4-6`).
  - `ValidateStep` filtra por proveedor (`run.go:56`). Un paso manual de un proveedor sin receta se valida sin más código.
- **Diagnóstico** (`tooling/cli/doctor_integrations.go`).
  - Engram se detecta con `programEvidence(deps, synthetic, "engram")` (`:132`), que solo hace `lookPath`; una prueba asegura que nunca ejecuta el programa (`doctor_integrations_test.go:220-235`).
  - `nextStep` (`:227-247`) cae en `default:` con el texto de agent-browser, así que CodeGraph necesita su propio caso.
- **TUI:** `integrationsFixedRows = 8` (`tooling/cli/tui_integrations_view.go:22`) cubre el título, la cabecera y 6 filas; con CodeGraph el máximo pasa a 7. La descripción de `tui_views.go:275` ya está desfasada porque omite la tarjeta de Codex.
- **Bloque gestionado:** Hive conserva lo que hay fuera de su bloque (`tooling/management/files.go:220-247`; `TestLifecyclePreservesUserBytesModesAndExtraFiles`). En esta máquina, el bloque de CodeGraph va después del de Hive en los tres archivos de instrucciones. En Codex y OpenCode hay un bloque de Context7 entre los dos.
- **`hive update` sin terminal:** si el plan cambia archivos, falla (`deployment-manager.md:40`). Desde el shell del agente hay que usar `hive update --out <plan>` y luego `hive apply --plan <plan>`.
- **Roles de solo lectura:** `hive-review-plan.md:17` prohíbe las «mutating probes». `codegraph sync` escribe la base del índice.
- **Spec desfasada:** `versioned-installation` dice que el catálogo «SHALL be limited to Engram, Context7, and pi-subagents» y que «Detecting `agent-browser` SHALL NOT add it to the optional capability catalog». Las dos afirmaciones contradicen el código actual.

## Enfoque elegido

### Catálogo y diagnóstico (D6-A)

- **Entrada manual `codegraph`**, al final de `Catalog()`:
  - hosts `claude`, `codex`, `cursor` y `opencode`;
  - fuente `github.com/colbymchenry/codegraph`;
  - motivo: «Install it and run its own `codegraph install` for your hosts, following the official instructions; Hive does not run it or edit host configuration.»
  - Sin receta: el gestor no ejecuta nada.
- **Fila CodeGraph en `hive doctor`**, en este orden exacto de filas: `engram, context7, pi-subagents, [pi-codemode], agent-browser, codegraph, [codex-question-card]`:
  - `programEvidence(deps, synthetic, "codegraph")` hace solo `lookPath` y nunca ejecuta el binario, como exige la spec «Read-only diagnostics»;
  - la fila entra en el bucle de `fillSourceFromCatalog` y `rec.apply` (`doctor_integrations.go:144-147`) para tener `Source` y el estado del último onboarding.
  - Siguiente paso cuando no se detecta: remite a la fuente oficial y a ejecutar `codegraph install` del propio CodeGraph para los hosts del usuario.
  - Siguiente paso cuando se detecta: indexar un repo con `codegraph init` solo si el usuario lo decide.
  - Ningún texto incluye `npm`, `npx`, `brew` ni `curl`, que el ayudante de pruebas prohíbe.
- **TUI:**
  - se suma la fila al presupuesto (`integrationsFixedRows` de 8 a 9) y se actualizan los comentarios de `tui_integrations_view.go:2-3` y `:18-19`;
  - la descripción del menú (`tui_views.go:275`), hoy desfasada, pasa a un texto genérico de 64 caracteres o menos, como «Check optional integrations and their next steps», porque la línea del menú ya mide 80 columnas (`tui_views.go:314-323`, `TestAppMenuAndGenericViewsFit`).

### Regla de uso (D7-A, D8-A)

Una viñeta en la sección Evidence de `content/guidance/global.md`, después de la de Context7. Nombra la capacidad y usa CodeGraph como ejemplo, como hace la viñeta de Context7. Texto previsto:

> - When a repository has a CodeGraph index, a `.codegraph/` directory at its root, and CodeGraph's MCP tool or `codegraph` command is available, use it to locate symbols, callers, and the impact of a change before text search; otherwise use text search. Never create an index, since indexing is the user's decision. Its MCP server keeps the index current while its session runs; the `codegraph` command does not, so before relying on callers, callees, or impact from that command, run `codegraph sync -q` first; a read-only or delegated role does not sync, and reports those relations as possibly stale. CodeGraph links some imports and calls by name when it cannot resolve them, without marking the guess: confirm with text search, or by reading the importing file, a result that crosses applications in a monorepo or comes from a workspace-package import or a path alias the root `tsconfig.json` does not declare.

Por qué en `global.md` y no en `flow-research`: el bloque del instalador y la nota local actual se cargan siempre, y una pregunta suelta sin skill activo también consulta el índice. La regla sustituye a la nota local (T5), con todo su contenido, y complementa el bloque del instalador. Coincide con él en usar el índice antes de buscar texto, y añade lo que ese bloque no dice: cuándo sincronizar, qué no hacen los roles de solo lectura y qué confirmar.

**Lugares a revisar en la PR, según «Guidance structure» de `AGENTS.md`:**
- `global.md` (Evidence y Proportionality);
- `content/agents/review/hive-research.md`;
- `content/skills/flow-research/SKILL.md`;
- `content/skills/flow-build/references/verification.md`;
- `site/`;
- `_support/docs/architecture/`;
- `tests/pilot/`.

Hoy ninguno tiene una regla sobre CodeGraph. `tests/pilot/flash_class_config_test.go:12,15` lo menciona, pero solo como dato de prueba.

### Spec

`MODIFIED` de dos requisitos de `versioned-installation`:

- **«Optional operation capabilities»:** el catálogo real son seis complementos; CodeGraph es solo manual, sin receta y sin gate nativo pendiente.
- **«Read-only diagnostics»:**
  - añade CodeGraph y la fila de la tarjeta de Codex a la evidencia local;
  - sustituye la frase sobre agent-browser, que el código ya contradice, por una que limita la detección de CodeGraph a `lookPath`.

### Despliegue y retirada de la nota local (T5)

Después del merge, como el marcador de mantenedor existe:
1. reconstruir el binario;
2. `hive update --dry-run`, luego `hive update --out <plan>` y `hive apply --plan <plan>` (sin terminal, `hive update` sin `--out` falla);
3. comprobar que la regla está en los tres archivos desplegados;
4. guardar la sección de cada archivo en `_support/workspace/gh-104-codegraph-complement/backup/` (ignorado);
5. borrar solo la sección `## CodeGraph limits (local note)`. En los tres archivos es la última sección, va desde su título hasta el final del archivo, después de `CODEGRAPH_END`, y las tres copias son idénticas.

Para revertir, se vuelve a insertar la sección guardada al final del archivo, sin restaurar el archivo entero, para no pisar escrituras posteriores. El bloque del instalador no se toca.

## Alternativas descartadas

- **Receta ejecutable (D6-B):** el instalador de CodeGraph escribe configuración de cuatro hosts. Es un tercero que escribe por ellos, fuera del precedente de `pi install` y `codex features enable`, que son comandos de cada host.
- **Hooks de git o de lefthook (D8-B y D8-C):** añaden maquinaria por repo y afectan al equipo si se versionan. La regla de sync cubre el hueco sin instalar nada.
