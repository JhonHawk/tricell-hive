# Tareas

Commit base: `b49c2abb` (`development`, 2026-10-23). Rama de trabajo: `feat/gh-104-codegraph-complement`.

Reparto:
- **T1 y T2 se delegan juntas** a un implementador de backend Go: su interfaz está fijada, solo escriben en `tooling/providers/*.go` y `tooling/cli/*.go`, y T2 depende de T1.
- **T3 y T4 van en el hilo principal en paralelo:** son cambios de redacción que no comparten archivos con T1 y T2.
- **T5 corre después del merge.**

### T1 — CodeGraph en el catálogo

- [x] `Catalog()` incluye `codegraph` como proveedor manual y `ValidateStep` lo admite.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:** `tooling/providers/types.go:13-21` (constante de ID), `tooling/providers/catalog.go:7-40`, `tooling/providers/run.go:56`, `tooling/providers/providers_test.go:30-38`, `tooling/providers/run_test.go`.

**Execution:** delegada junto con T2 (`hive-build-backend`); interfaz fijada, sin escritores solapados.

**Test approach:** tdd.

**Changes:**
- Añadir `CodeGraph ID = "codegraph"`.
- Añadir la entrada al final de `Catalog()` con los hosts, la fuente y el motivo de design.md.
- Añadir `CodeGraph` al `case` de `ValidateStep`, sin admitir pasos `Planned` (sin receta).

**Verification:** `go test ./tooling/providers/` pasa, con estos casos nuevos que fallan en la base:
- `wantOrder` incluye `codegraph` al final;
- la entrada tiene hosts `claude, codex, cursor, opencode` y fuente `github.com/colbymchenry/codegraph`;
- `codegraph-manual` (Manual) se acepta;
- un paso `Planned` con `Argv` para `codegraph` se rechaza.

### T2 — Fila de CodeGraph en diagnóstico y TUI

- [x] `hive doctor` y la vista Integraciones muestran la fila de CodeGraph.

**Closes:** AC3, AC4.

**Depends on:** T1.

**Locations:** `tooling/cli/doctor_integrations.go` (`:131-148` filas y bucle de fuente y registro `:144-147`, `:227-247` `nextStep`, `:349-358` `programEvidence`), `tooling/cli/tui_integrations_view.go` (`:2-3`, `:18-19`, `:22`), `tooling/cli/tui_views.go:275`, `tooling/cli/doctor_integrations_test.go` (`:144`, `:163`, `:178`, `:233`, `:378-379`, `:746`), `tooling/cli/tui_integrations_view_test.go` (`:46`, `:74`, `:88-95`, `:107`), `tooling/cli/provider_adapter_test.go` (modelo en `:56` y `:83`).

**Execution:** delegada con T1.

**Test approach:** tdd.

**Changes:**
- Fila `codegraph` con `programEvidence(deps, synthetic, "codegraph")`, en el orden exacto de design.md, e incluida en el bucle de `fillSourceFromCatalog` y `rec.apply`.
- Caso propio en `nextStep` (`codegraphNextStep`), con los textos de design.md y sin `npm`, `npx`, `brew` ni `curl`.
- `integrationsFixedRows` pasa a 9, y se actualizan los comentarios de `tui_integrations_view.go:2-3` y `:18-19`.
- La descripción de `tui_views.go:275` pasa a un texto genérico de 64 caracteres o menos.
- Actualizar las listas de IDs, los conteos de filas y el nombre de la prueba de `:88-95`, que pasa de seis a siete filas.

**Verification:**
- `go test ./tooling/cli/` pasa, con casos que fallan en la base:
  - con `codegraph` en el PATH falso, la fila sale `detected` con la ruta, y la prueba de `:220-235` confirma que no se ejecuta;
  - sin él, sale `not detected`;
  - con home sintético, sale `not checked`;
  - la fila muestra su `Source`; el mapa de `TestIntegrationsEveryRowShowsItsSourceAndNextStep` incluye `codegraph`;
  - un registro de onboarding con `codegraph-manual` aparece en la fila;
  - el texto del siguiente paso pasa `assertNoCatalogReasonOrInventedCommand`;
  - una instalación con `--hosts codex` que elige la oferta de CodeGraph no ejecuta ningún proceso, deja el paso `codegraph-manual` y sale con código distinto de cero;
  - la vista cabe en 80×24 con el número máximo de filas, y `TestAppMenuAndGenericViewsFit` sigue pasando.
- Observación local: `go run ./tooling/cli doctor` muestra la fila CodeGraph `detected` con la ruta que devuelve `command -v codegraph` en el mismo shell.

### T3 — Regla de uso en la guía global

- [x] `global.md` (Evidence) tiene la viñeta de design.md.

Revisión R1 (`/code-review` de la PR #105): AC5 cambia; T3 se verifica de nuevo.

**Closes:** AC5.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:58` (insertar después de la viñeta de Context7).

**Execution:** hilo principal; el brief sería más largo que el cambio.

**Test approach:** check: búsqueda y lectura.

**Changes:**
- Insertar la viñeta.
- Revisar los lugares que lista design.md («Lugares a revisar en la PR»).

**Verification:**
- `rg -n -i "code-graph|codegraph" content/guidance/global.md` encuentra la viñeta; en la base no hay ninguna coincidencia.
- La lectura confirma las seis partes de AC5:
  - usarlo antes de la búsqueda de texto;
  - volver a la búsqueda de texto cuando no hay herramienta de CodeGraph;
  - sincronizar al usar el comando `codegraph`;
  - los roles de solo lectura o delegados no sincronizan y avisan;
  - confirmar los enlaces adivinados;
  - no crear índices.
- `wc -c content/guidance/global.md` da el tamaño nuevo (base: 32.693 bytes) para la PR.

### T4 — Documentación y delta de la spec

- [x] La documentación enumera CodeGraph como complemento manual y el delta de la spec está escrito.

**Closes:** AC6.

**Depends on:** ninguna (los textos siguen design.md).

**Locations:** `README.md:156-160`, `llms.txt:43`, `SECURITY.md:41-42`, `_support/docs/architecture/tool-dependencies.md` (tabla `:24-30`), `_support/docs/architecture/installer.md:101-106`, `_support/docs/architecture/deployment-manager.md:76,79`, `tooling/providers/README.md:3-18`, `CHANGELOG.md` (`[Unreleased]`), `specs/versioned-installation/spec.md` de esta carpeta (ya escrito).

**Execution:** hilo principal.

**Test approach:** check: búsqueda.

**Changes:**
- Añadir CodeGraph como manual, con su motivo, en cada lista.
- En `deployment-manager.md:76`:
  - la fila CodeGraph;
  - «room for six rows» pasa a siete;
  - se quita «`agent-browser` is detected but is not part of the optional capability catalog», que el código contradice.
- En `deployment-manager.md:79`, nombrar CodeGraph entre los programas que no se buscan con `--home`.
- Entrada en `CHANGELOG.md`.

**Verification:** `rg -il codegraph README.md llms.txt SECURITY.md _support/docs/architecture/tool-dependencies.md _support/docs/architecture/installer.md _support/docs/architecture/deployment-manager.md tooling/providers/README.md` lista los siete archivos; en la base no hay ninguno.

### T5 — Despliegue local y retirada de la nota

- [x] La guía desplegada contiene la regla y la nota local ya no existe en los tres hosts.

**Closes:** AC7.

**Depends on:** T1–T4 (catálogo, fila de diagnóstico, regla y documentación) integradas en `development`.

**Locations:** `~/.claude/CLAUDE.md:130`, `~/.codex/AGENTS.md:150`, `~/.config/opencode/AGENTS.md:149`.

**Execution:** hilo principal, después del merge (marcador `maintainer.local` presente).

**Test approach:** check: lectura de los archivos desplegados.

**Changes:** según design.md, «Despliegue y retirada de la nota local»:
- `go build -o "$(command -v hive)" ./tooling/cli`;
- `hive update --dry-run`, luego `hive update --out <plan>` y `hive apply --plan <plan>`;
- guardar la sección en `_support/workspace/gh-104-codegraph-complement/backup/`;
- borrar en cada archivo solo la sección `## CodeGraph limits (local note)`.

**Verification:**
- `rg -c "CodeGraph index" ~/.claude/CLAUDE.md ~/.codex/AGENTS.md ~/.config/opencode/AGENTS.md` devuelve 1 por archivo.
- `rg "CodeGraph limits" ...` no imprime nada y sale con código 1.
- `rg -c "CODEGRAPH_START" ...` devuelve 1 por archivo.
- `hive doctor` muestra la fila CodeGraph.

## Verificación compartida

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, rama de trabajo | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, todos en verde (~50 s) | Incluida en la implementación |
| Recorrido local | Tras T2 | `go run ./tooling/cli doctor` y la vista Integraciones del TUI con la fila CodeGraph | Incluida en la implementación |
| Revisión de código | PR abierta, antes del merge | `/code-review` de Claude Code sobre el diff completo, una vez; los P0 y P1 se corrigen | D10-A |
| Despliegue local | Tras el merge | T5 | Marcador de mantenedor y D7-A |

No hay cambio de UI renderizada en la web: el TUI cubre el ajuste de filas con su prueba de 80×24.

## Revisión y progreso

- **Revisión del plan:** revisión `28fa2b0e190b`, dos dominios con `hive-review-plan` nativo y en paralelo.
  - **Gestor Go (T1 y T2):** 1 bloqueante y 5 no bloqueantes; 111 s y unos 119k tokens.
  - **Guía, documentación, spec y despliegue (T3–T5):** 0 bloqueantes y 8 no bloqueantes; 150 s y unos 126k tokens.
- **Hallazgos aceptados y aplicados.** Todos son la propuesta del propio revisor, así que se reconcilian sin otra ronda:
  - la descripción del menú pasa a un texto genérico de 64 caracteres o menos (era el bloqueante);
  - la fila de CodeGraph entra en el bucle de fuente y registro, con orden exacto de filas;
  - prueba del flujo de instalación manual;
  - hosts y fuente comprobados en T1;
  - la ruta esperada sale de `command -v`;
  - la regla exime a los roles de solo lectura de sincronizar;
  - el antecedente «from them» se aclara y se nombra `pendingChanges`;
  - el defecto de enlazar por nombre se acota a CodeGraph y se conserva «leyendo el archivo que importa»;
  - T5 usa `--out` y `apply`, y revierte reinsertando la sección;
  - `SECURITY.md` y `deployment-manager.md:76` entran en T4;
  - en la spec, CodeGraph queda solo manual y la tarjeta de Codex entra en el diagnóstico;
  - las verificaciones usan `rg` y `wc -c`.
- **Límites:** ningún revisor ejecutó pruebas ni `hive`. Nadie midió el asistente de hosts (`tui_hosts_view.go:1002`) con seis ofertas; queda para el recorrido local de T2.
- **Revisión de código (`/code-review`, PR #105, unos 150k tokens y 130 s):** 8 hallazgos, ninguno P0 ni P1.
  - **Corregidos, introducidos por el cambio y locales:**
    - la regla vuelve a la búsqueda de texto cuando no hay herramienta;
    - el disparador `pendingChanges` se elimina: con el comando siempre se sincroniza, y la sesión MCP se pone al día al conectarse;
    - `deployment-manager.md:216` («four or five»);
    - el comentario de `provider_adapter.go:13`.
  - **Descartados:**
    - el registro del cambio fuera de la PR: lo versiona `flow-close` al cerrar, por diseño;
    - la fila de CodeGraph mostrada sin hosts compatibles: Engram y Context7 ya se muestran siempre, así que es el mismo patrón;
    - la regla repite el bloque del instalador: decidido en D7-A, y Grok y Pi no tienen ese bloque.
  - **Solo informado (P3):** el relleno de la vista quita una línea al detalle en instalaciones con menos de siete filas.
- **Progreso (build):**
  - Modelos de T1 y T2, configurados y no observados: implementador `hive-build-backend` en sonnet con esfuerzo medio; verificadores `hive-verify-task` en opus con esfuerzo alto.
  - T3 verificada (AC5 met) y confirmada en `c10952d5`. Se cambió «the command always exits 0» por «its exit code does not reflect pending changes», a propuesta del verificador.
  - T4 verificada (AC6 met) y confirmada en `8766958e`. Se añadieron la fecha de comprobación de CodeGraph en `tool-dependencies.md` y la condición «registration cannot be read» en el delta.
  - T1 y T2 implementadas con TDD: RED observado (5 de 6 entradas del catálogo, `unknown provider`, 21 pruebas de `tooling/cli`) y luego GREEN. Verificadas (AC1–AC4 met) y confirmadas en `2886523e` y `224eead5`. `go run ./tooling/cli doctor` muestra CodeGraph `detected` con la misma ruta que `command -v codegraph`.
  - T3 se verificó de nuevo tras R1 (AC5 met) y se confirmó en un commit adicional `(T3)`.
  - Suite completa en el candidato: `go vet` limpio, 20 paquetes ok y 55 pruebas de skills OK.
  - PR #105 mergeada en `2c296d74` y #104 cerrado.
  - T5:
    - `hive update --out` y `hive apply` desde `2c296d74`;
    - la regla aparece una vez por archivo en los tres hosts;
    - la nota local (416 bytes por archivo) está retirada, con copia de la sección en `_support/workspace/gh-104-codegraph-complement/backup/`;
    - `CODEGRAPH_START` sigue presente;
    - `hive doctor` muestra la fila CodeGraph.
