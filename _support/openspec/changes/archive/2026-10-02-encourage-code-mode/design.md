# Diseño

## Contexto verificado (2026-10-02)

**Pi 1.0.0.** Fuente: los documentos del paquete instalado en `~/.pi/agent/install/releases/1.0.0/node_modules/@earendil-works/pi-coding-agent/docs/`.

- **Qué es:** `codemode` es una extensión integrada que viene registrada pero inactiva. Ejecuta JavaScript en QuickJS y desde ahí llama a cualquier herramienta con `tools.<name>(args)`, incluidas `bash`, `edit` y las de MCP (`cli.md:141-168`).
- **Cómo se activa:**
  - con `"defaultTools": ["+codemode"]` en `~/.pi/agent/settings.json` o en `.pi/settings.json` (`cli.md:150-158`, `settings.md:40-56`);
  - solo, cuando conecta un servidor MCP con exposición `codemode`, que es la predeterminada, salvo que `autoEnableCodemode` sea `false` (`mcp.md:182`, `mcp.md:206`).
  - `/reload` activa las herramientas recién agregadas a `defaultTools`.
- **Herramientas diferidas:** las de exposición `codemode` o `deferred` se pueden llamar desde un script (`mcp.md:204`). El `mcp.json` del usuario declara `linear` y `heroui` con el servidor oculto y herramientas diferidas, así que no hay que cambiarlo.
- **Permisos:** Pi no pide aprobaciones. Las llamadas hechas desde un script pasan por los mismos controles que las directas (investigación del 2026-10-01 sobre `docs/mcp.md` y la fuente instalada).

**OpenCode 2.0.21.** Fuente: el tag `v2.0.21` de `anomalyco/opencode`, investigado el 2026-10-01.

- La herramienta `execute` está activa por defecto para MCP y plugins.
- Las herramientas integradas están fijadas a `codemode: false` en `packages/core/src/tool.ts:234-237`, así que no hay nada que activar. La única palanca de Hive es la guía.
- El `opencode.json` del usuario no tiene reglas `ask` sobre MCP, por lo que los fallos abiertos de permisos dentro de `execute` no le afectan hoy.

**Anuncios.** [Pi 1.0](https://earendil.com/posts/pi-1-0/) presenta el code mode como función estable. ["You Said No MCP!"](https://earendil.com/posts/you-said-no-mcp/) explica que corre del lado de confianza del propio Pi y muestra un análisis de 167 issues de Linear con 331 llamadas desde un solo script.

**`hive doctor`.**

- **Regla de diseño:** la sección Integrations no lee la configuración de ningún host ni proveedor (`tooling/cli/doctor_integrations.go:1-7`, [deployment-manager.md](../../../docs/architecture/deployment-manager.md) línea 73, requisito "Read-only diagnostics" de [versioned-installation](../../specs/versioned-installation/spec.md)).
- **Precedente:** la fila `pi-subagents` ya resuelve el mismo caso con `Found: "not checked"`.
- **Forma de las filas:** `agent-browser` muestra cómo se arma una fila que no está en `providers.Catalog`. Fija `Source` en el código y usa `recordNotTracked`.
- **Vista de la TUI:** `tooling/cli/tui_integrations_view.go:20` fija `integrationsFixedRows = 6`, que es el título, la cabecera y cuatro filas.

**Guía.** `content/guidance/global.md`, sección Proportionality (líneas 79-93). La regla vecina de la línea 86 ya justifica cuidar el contexto: "early context is re-read for the rest of the session".

## Enfoque

### Fila `Pi codemode`

La fila se agrega en `collectIntegrationRows`, entre `pi-subagents` y `agent-browser`, y se arma solo con constantes. Así no puede depender de la configuración de Pi, que es lo que exige AC3.

- **Solo con Pi registrado.** La fila aparece solo cuando Pi está registrado, como se le presentó al usuario en D4-A. Para saberlo se usa `management.RegisteredHosts(o)`, que lee el estado de Hive y no la configuración de Pi. `collectIntegrationRows` ya lee ese directorio de estado con `readLastOnboarding`.
  - Si `RegisteredHosts` falla, la fila se muestra igual. Es solo informativa, y el error de estado ya aparece en la sección Installation.
- **No pasa por el catálogo ni por el registro de onboarding.** No usa `fillSourceFromCatalog` ni `rec.apply`, que la dejarían con `"-"` y "No onboarding record yet". `nextStep` necesita un caso explícito para `pi-codemode`, porque el `default` aplica `agentBrowserNextStep` (`doctor_integrations.go:185-186`).
  - El comentario de `nextStep` que dice que el único comando sugerido es el de Context7 (`doctor_integrations.go:173-174`) se ajusta: el paso nuevo menciona `/reload` dentro de Pi.

- **ID:** `pi-codemode`. **Name:** `Pi codemode`.
- **Found:** `not checked`. **Evidence:** `Not checked: finding it would mean reading Pi's configuration, which Hive does not do.`
- **Record:** `recordNotTracked`. **RecordDetail:** `Not part of onboarding: Hive does not configure Pi's tools.`
- **Source:** `github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/cli.md#enable-codemode`. La URL devolvió HTTP 200 el 2026-10-02. La forma sin esquema sigue la de `agentBrowserSource`.
- **Next**, como un caso más en `nextStep`: `Pi turns codemode on by itself when an MCP server uses codemode exposure, its default. To keep it on in every session, add "defaultTools": ["+codemode"] to Pi's settings.json (by default ~/.pi/agent/settings.json) and run /reload in Pi, or ask Pi to enable codemode; see Source above.`

**La TUI.** `integrationsFixedRows` pasa a 7, el máximo: título, cabecera y cinco filas. La vista ya rellena hasta ese número cuando hay menos filas (`tui_integrations_view.go:158-160`) y recorre `v.rows` sin suponer cuántas hay. El comentario de cabecera dice "four or five rows". La descripción del menú en `tooling/cli/tui_views.go:275` suma "Pi codemode".

### Regla de guía

Se agrega un punto nuevo en Proportionality, justo antes de la regla de delegación (línea 86) para no partir el bloque de delegación (líneas 86 a 91). Va redactado por capacidad, como pide `AGENTS.md`:

> - When the host offers a tool that runs a script able to call other tools, make three or more read-only calls that script can reach, or one whose result you would only filter, sort, or count, from one script that returns only what the answer needs, since every result in the conversation is re-read for the rest of the session. Keep direct a single call whose result you must read whole; delegate a result whose relevant part needs judgment to find, as the next rule says. Calls inside a script may skip the host's approval, so put a write, edit, or state-changing command in one only when it is authorized and confirmed as a direct write, and keep secret values out of the script's text.

Por qué está redactada así:

- **Umbral de tres llamadas:** sale de la propuesta que el usuario aceptó en D1-A.
- **Solo lecturas:** "read-only" y "that script can reach" limitan la regla a lecturas que el script alcanza. En OpenCode, el script no llega a las herramientas integradas.
- **Script o subagente:** la segunda frase separa los dos casos. Un filtro mecánico va al script; lo que necesita juicio va al subagente, según la regla siguiente. Una sola llamada solo queda directa si su resultado hay que leerlo entero, así que no contradice la primera frase.
- **Escrituras y secretos:** la tercera frase cubre que Pi no pide aprobaciones y que OpenCode tiene fallos de permisos dentro de `execute`. "Confirmed as a direct write" conserva la confirmación de registros, cambio y vuelta atrás que exige la sección Scope and authorization. También extiende al texto del script la regla de secretos, que solo nombraba argumentos y URL.
- **Sin nombres de herramientas:** la regla no nombra `execute` ni `codemode`, porque esos nombres cambian entre hosts y versiones.
- **Tamaño:** son 688 bytes. `content/guidance/global.md` pasa de 42615 a unos 43304 bytes, por debajo del tope de 43622 de `tests/content/budget_test.go:12`. No hace falta subirlo.

### Alternativas descartadas

- **Que Hive escriba `defaultTools` (D3-C):** Hive nunca escribe preferencias de Pi; hasta para `pi-subagents` ejecuta `pi install` (`tooling/management/types.go:198`). Haría falta registrar la propiedad de la entrada, detectar conflictos y deshacer cambios para algo que Pi resuelve con una línea.
- **Detectar leyendo la configuración (D4-B):** rompe la regla de diseño y daría falsos "apagado".
