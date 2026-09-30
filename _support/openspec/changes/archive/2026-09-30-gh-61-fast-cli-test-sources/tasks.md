# Tareas

Base del cambio: `3530d13` (`origin/development`). Worktree: `.claude/worktrees/gh-61`, rama `feat/gh-61-fast-cli-test-sources`.

## T1 — `openMenuEntry` dibuja la vista antes de devolver el control (#59)

- [x] `no_CLI_hosts` deja de esperar 60 s. `openMenuEntry` llama a `d.screen()` después de `enter`. `hive-verify-task` (Sonnet) dio AC1 por cumplido:
  - la subprueba tarda 0,54 s, y 61,07 s si se quitan las líneas;
  - verificó que el mecanismo es el del parpadeo del cursor (`textinput.go:673-675`, `cursor.go:192-203`, `tui_test.go:151-172`);
  - las 17 llamadas a `openMenuEntry` se benefician del arreglo.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `tooling/cli/tui_update_view_test.go:23` (`openMenuEntry`).

**Execution:** hilo principal, porque son unas pocas líneas y la causa ya está verificada.

**Test approach:** check. La subprueba medida antes y después.

**Verification:** `go test -count=1 -run 'TestUpdateViewErrorsShowInsideTheView' -v ./tooling/cli` muestra `no_CLI_hosts` en 2 s o menos; en la base tarda 61 s.

## T2 — Fuente mínima en `testdata/` y pruebas de `cli` movidas a ella

- [x] `tooling/cli/testdata/minimal-source/` reemplaza a `minimalCatalogSource`. Evidencia, revisada por el orquestador:
  - `Abs("../..")` solo queda en `voice_test.go:67`, `"..", "..", "content"` solo en `doctor_project_test.go:380`, y ya no existen `minimalCatalogSource` ni `OnceValues`;
  - `testdata` tiene 4 archivos y no está ignorado por Git;
  - `go test -count=1 ./tooling/cli` pasa en 43,7 s; ninguna prueba de instalación o bootstrap supera 0,5 s.

  Cambios de afirmación y su motivo:
  - la vista de modelos pasa a 22 roles sintéticos exactos, exige `scrollable()` en 80 columnas y se renombra `TestModelsViewFitsAndScrollsWithManyRolesOnSixCLIs`;
  - `TestInstallSharedHostClosureRequiresConsent` conserva sus afirmaciones, con su propia copia de la fuente y una skill sintética;
  - `developmentSourceArgs` usa la fuente mínima.

  Implementó `hive-build-backend` (Sonnet, configurado).
**Closes:** AC2, AC3.

**Depends on:** ninguna. No toca los archivos de T1: si `tui_update_view_test.go` necesita otro cambio, se coordina con el hilo principal.

**Locations:**
- `tooling/cli/tui_hosts_test.go` (`minimalCatalogSource`, `minimalTestSource`, `TestMain`);
- `tooling/cli/tui_test.go:35` y `:1012`;
- `tooling/cli/install_test.go` (8 lugares, incluido `installArgs:110`);
- `tooling/cli/provider_adapter_test.go` (4);
- `tooling/cli/main_test.go:50`;
- `tooling/cli/bootstrap_test.go:39`, `:338`, `:597` y `:638`;
- `tooling/cli/tui_models_view_test.go:382-420`;
- `tooling/cli/testdata/minimal-source/` (nuevo).

**Execution:** delegada a `hive-build-backend`: la interfaz está decidida en [design.md](design.md#fuente-mínima-fija-d2-a) y solo escribe en `tooling/cli`, sin cruzarse con T3.

**Test approach:** characterization. Las pruebas existentes son la especificación: deben seguir pasando con las mismas afirmaciones de comportamiento. Cada afirmación que dependía de lo que traía la fuente y se ajuste queda anotada aquí con su motivo.

**Verification:**
- `rg -n 'Abs\("\.\./\.\."\)' tooling/cli/*_test.go` solo devuelve `voice_test.go`.
- `rg -n '"\.\.", "\.\.", "content"' tooling/cli/*_test.go` solo devuelve `doctor_project_test.go:380`.
- `rg -n 'minimalCatalogSource|OnceValues' tooling/cli` no devuelve nada.
- `TestInstallSharedHostClosureRequiresConsent` sigue exigiendo `Affected shared resources:` y `.agents`, con la skill sintética de [design.md](design.md#fuente-mínima-fija-d2-a).
- La vista de modelos exige `scrollable()` en 80×24.
- `go vet ./tooling/cli` y `go test -count=1 ./tooling/cli` en verde.
- Cada prueba de la lista de 19 lentas sigue existiendo, o se registra el cambio de nombre de la vista de modelos, y pasa.

## T3 — Una prueba con contenido real y chequeo de binario de prueba

- [x] (`hive-verify-task`, Opus: AC4 y AC5 cumplidos. El plan real instala 72 archivos, y las cuatro mutaciones hacen fallar las pruebas. El límite del chequeo con `--help` queda anotado: la garantía real es `go doc testing.Init`.) `TestRepositoryCatalogueInstructionReferences` instala y aplica el catálogo real, y `DisableDiskSyncForTests` entra en pánico fuera de un binario de prueba. Evidencia del implementador:
  - antes del cambio falla con «did not panic outside a test binary» y después pasa;
  - con la condición invertida falla el paquete entero, porque entra en pánico `TestMain`;
  - `hive --help` no muestra flags `-test.*`;
  - `-race` pasa en 91 s, en paralelo con T2.

  Para el catálogo real se usó el host `claude`: sus skills van a `~/.agents/skills` y sus agentes a `~/.claude/agents`.

**Closes:** AC4, AC5.

**Depends on:** ninguna. Solo escribe en `tooling/management`, así que corre en paralelo con T2.

**Locations:** `tooling/management/references_test.go:60`, `tooling/management/files.go` (`DisableDiskSyncForTests`) y `tooling/management/sync_test.go`.

**Execution:** delegada a `hive-build-backend`, un segundo hijo en paralelo con T2: otro paquete y ningún archivo en común.

**Test approach:** tdd. Primero la prueba del pánico, que falla en la base porque la función no comprueba nada, y la ampliación de la prueba del catálogo; después el cambio.

**Verification:**
- `go test -count=1 -run 'TestRepositoryCatalogueInstructionReferences|TestDisableDiskSync' -v ./tooling/management` en verde.
- La prueba del catálogo muestra el apply y el segundo plan sin cambios.
- Con la condición del pánico invertida, la prueba falla; después se restaura.
- `go build -o <scratch>/hive ./tooling/cli`, y `<scratch>/hive --help` no muestra flags `-test.*`.
- `go test -count=1 -race ./tooling/management` en verde, porque la prueba cambia estado de paquete.

## T4 — Tiempos y CI

- [x] Los tiempos de `tooling/cli` y de la CI cumplen los límites. [Corrida de CI 36684332838](https://github.com/JhonHawk/tricell-hive/actions/runs/36684332838) del PR #71:
  - total 8 min 20 s (antes 27 min 13 s);
  - `go test -race`: 7 min 12 s (antes 26 min 14 s), así que AC7 se cumple;
  - `tooling/cli` 268 s (antes 1464 s);
  - `tooling/management` 248 s (antes 132 s): ahora aplica el catálogo real y es casi el camino crítico.

  Las mediciones locales: Local (AC6), corriendo solo: 44,6 s sin `-race` (base 140 s) y 115,9 s con `-race` (base 839 s). La suite completa sin `-race` tarda 45,5 s (base 2 min 23 s). AC7 queda pendiente de la CI del PR.

**Closes:** AC6, AC7.

**Depends on:** T1, T2 y T3; la parte de la CI, del PR.

**Locations:** ninguna; es medición.

**Execution:** hilo principal, porque mide con el paquete corriendo solo y depende del PR que abre el hilo principal.

**Test approach:** check.

**Verification:**
- `go test -count=1 ./tooling/cli` tarda 60 s o menos, y `go test -count=1 -race ./tooling/cli` 240 s o menos, corriendo cada uno solo.
- En el PR, el paso `go test -race` de la corrida `CI` tarda 12 min o menos, según `gh run view <id> --json jobs`.

## Verificación y revisión

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | T1 a T3, en local | `hive-verify-task`, con otro modelo que el implementador | Parte de `flow-build` |
| Suite local | Una vez sobre el candidato final | `go vet ./...` y `go test ./...` en verde | Parte de la implementación |
| Revisión de código | Antes del push | `/code-review` de Claude Code (D6-A de esta sesión) | Autorizada |
| CI del PR | Tras el push | `CI` en verde; su tiempo cierra AC7 | Autorizada (D5-A) |

Sin efecto visible en la interfaz: T1 y T2 solo tocan pruebas, y T3 solo cambia un camino de error del código de producción.

## Estado de la revisión y avance

Plan escrito el 2026-09-30 sobre `3530d13`. Lo revisó un `hive-review-plan` nativo de Claude Code, de solo lectura, en el dominio de código de pruebas en Go (todo el cambio). Además de leer el plan, hizo una prueba de concepto en una copia desechable del repositorio.

Aceptados:
- **B1:** `TestInstallSharedHostClosureRequiresConsent` necesita una skill sintética y no se relaja su afirmación.
- **B2:** el límite de AC6 sin `-race` pasa de 45 a 60 s; la prueba de concepto dio 48,1 s.
- **N1:** la fuente mínima incluye `agent-profiles.json`.
- **N2:** 22 roles sintéticos y la prueba exige `scrollable()`.
- **N3:** nuevo patrón de búsqueda en AC2.
- **N4:** un solo CLI y `PlanUnchanged` en T3.
- **N5:** las referencias a T5 y T6 corregidas.

Construcción (2026-09-30): commits `c45a740` (T1), `901d1b7` (chequeo de T3) y `14eb3fc` (T2 y la prueba del catálogo real), y [PR #71](https://github.com/JhonHawk/tricell-hive/pull/71).

`/code-review` devolvió diez hallazgos.

Aplicados por `hive-build-backend`:
- **F4:** roles del perfil verificador y de los perfiles de acceso en la vista de modelos.
- **F5:** plan del catálogo real con `codex` y `claude`, y apply solo con `claude`.
- **F6:** `VERSION` en `testdata`; `doctorHome` usa una copia sin ella.
- **F7:** `EvalSymlinks` en la ruta de la fuente.
- **F8:** la prueba del pánico da un solo mensaje cuando falla.
- **F9:** una sola forma de copiar la fuente.
- **F10:** los fixtures de bootstrap añaden una skill anidada, un agente y los perfiles reales.

No aplicados:
- **F1 (importar `testing`):** refutado. El binario crece 16,8 KB (+0,08 %), `flag` ya estaba enlazado y `testing.Testing` está pensado para código que no es de prueba.
- **F2 (chequeo redundante):** es la decisión D3-A del usuario.
- **F3 (parpadeo en el controlador de pruebas):** fuera del alcance.

Suite completa sin `-race` tras los arreglos: en verde.

La cita de la guía de Google sobre ganchos exportados quedó matizada. Todas las correcciones aplican la propuesta del revisor, así que no hubo segunda ronda.

Límites que siguen abiertos:
- no se midió `-race` con la fuente mínima;
- no se probó la instalación de agentes en un solo CLI en T3;
- el tiempo de la CI se mide en el PR.
