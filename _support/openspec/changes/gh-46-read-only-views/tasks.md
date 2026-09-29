# Tareas

**Base:** `7f06e70`, fijada al empezar el build el 2026-09-29.
- El plan se escribió y revisó sobre `82f2dcc`.
- La base avanzó a `eddcf41` (#51–#53: nombres `hive-*` de los agentes y pruebas de catálogo) y luego a `7f06e70`, que cambia una línea de la definición de Pending en `content/guidance/global.md`. Ninguno de los dos toca un archivo del plan ni la sección «Project settings» que lee AC10.

Las rutas son relativas a la raíz del repositorio. Los comandos se corren desde la raíz del worktree. Ninguna prueba lee la configuración real: todas usan `doctorDeps` falsos o un home de prueba.

## T1 — Modelos efectivos por rol y vista Models

- [/] La vista Models muestra el modelo y el esfuerzo efectivos por rol y CLI, calculados con la misma regla que `agents.Render`.

**Closes:** AC5.

**Depends on:** ninguna.

**Locations:**
- `integrations/agents/agents.go` (`Resolve` nueva, `Render`) y su golden de salida;
- `tooling/management/models.go` nuevo (`EffectiveModels`, `ModelRow`);
- `tooling/cli/models.go` nuevo (`renderModelsText` y la presentación de [design.md](design.md#contratos-nuevos));
- `tooling/cli/tui_models_view.go` nuevo.

**Execution:** delegada a `hive-build-backend`, en su propio worktree y en paralelo con T4. Los contratos están fijados y los archivos son propios.

**Test approach:** `tdd`. Para `Resolve`, `characterization` primero: el golden de `Render` se guarda antes de extraerla.

**Changes:**
1. Guardar el golden de `Render` (20 roles × 6 CLIs).
2. Extraer `Resolve` sin cambiar el golden.
3. `EffectiveModels` por registro con su propio `Record.Release`.
4. `renderModelsText` y la vista, con el presupuesto de filas de design.md.

**Verification:**
- `go test ./integrations/agents -run 'Resolve|RenderGolden'` pasa. El golden es idéntico antes y después de extraer `Resolve`.
- `go test ./tooling/management -run EffectiveModels` pasa, con estos casos:
  - un home de prueba con Claude, Codex, Pi y OpenCode instalados: para cada archivo de agente instalado, su `model` y su esfuerzo (`effort`, `model_reasoning_effort`, `thinking` o el sufijo `#variant`) coinciden con la fila;
  - un rol con `effort` propio muestra ese esfuerzo en Claude, Codex y Pi;
  - un registro compartido con una release anterior usa esa release;
  - sin CLIs registrados, da una lista vacía.
- `go test ./tooling/cli -run ModelsView` pasa, con estos casos:
  - ←→ cambia de CLI y el elegido va entre corchetes;
  - Grok muestra «host default», y el `inherit` de Claude muestra «inherit (parent session)»;
  - en OpenCode, un modelo `…#max` se muestra sin el sufijo en la columna Model y con `max` en la columna Effort;
  - el pie de la vista contiene `integrations/agent-profiles.json` y `hive update`;
  - sin CLIs: «No CLI hosts are registered»;
  - `state.json` y el home quedan idénticos.

## T4 — Diagnóstico: CLIs, instalación y sesiones

- [/] La vista Diagnostics muestra las secciones CLIs, Installation y Sessions según las reglas de [design.md](design.md#reglas-por-sección).

**Closes:** AC2, AC3, AC4.

**Depends on:** ninguna.

**Locations:**
- `tooling/cli/doctor.go` nuevo (`doctorDeps`, `doctorReport`, `sanitizeLine`, `collectDoctor` con las cinco secciones);
- `tooling/cli/doctor_integrations.go` y `tooling/cli/doctor_project.go`, con funciones de relleno;
- `tooling/cli/doctor_sessions.go` nuevo;
- `tooling/cli/tui_doctor_view.go` nuevo.

**Execution:** delegada a `hive-build-backend`, en su propio worktree y en paralelo con T1. Define los tipos compartidos y los rellenos que completan T2 y T3.

**Test approach:** `tdd`.

**Changes:**
1. `doctorDeps`, con `getenv` y con implementaciones reales y falsas.
2. Sección CLIs, con versiones en paralelo.
3. Sección Installation.
4. Sección Sessions.
5. `sanitizeLine`.
6. La vista, con el patrón de carga `owned{v}` y `seq`.

**Verification:** `go test ./tooling/cli -run 'Doctor|Sessions|SanitizeLine'` pasa, con estos casos:
- **AC2:**
  - un CLI detectado muestra su versión, su release corta y el estado de su instalación;
  - `cursor` ejecuta `cursor-agent`, y sin `cursor-agent` dice «version unavailable: cursor-agent not found»;
  - un `--version` falso que tarda 5 s da «version unavailable» y la sección termina en menos de 5 s;
  - un CLI no detectado no se ejecuta: el ejecutable falso escribe un marcador si corre, y el marcador no existe;
  - con `--home`, no se ejecuta nada.
- **AC3:**
  - un recurso en `drift` y un bloque con marcadores duplicados aparecen con su frase y su ruta;
  - los recursos `not_installed` no aparecen;
  - una operación pendiente da una sola línea con `hive recover`;
  - una instalación limpia dice «No problems found».
- **AC4:**
  - una sesión viva de Claude Code con `startedAt` anterior a la release se marca, y otra posterior no;
  - con Grok pasa lo mismo;
  - un `pid` muerto se ignora, y un `pid` 0 también;
  - un `active_sessions.json` de 2 MiB, otro con formato desconocido y otro sin permiso de lectura dan «Session check unavailable», y las demás secciones se muestran;
  - con Codex, Pi y Cursor registrados aparece el aviso de reiniciar, y con OpenCode, la nota de recarga.
- **General:**
  - una versión con `\x1b[2J` se muestra sin la secuencia;
  - `state.json` y el home quedan idénticos;
  - `getenv` falso con `CLAUDE_CONFIG_DIR` apunta a la carpeta de prueba.

## T2 — Integraciones

- [ ] La vista Integrations muestra Engram, Context7, pi-subagents y `agent-browser` con evidencia local, estado del último registro, fuente y siguiente paso.

**Closes:** AC6.

**Depends on:** T4 (`doctorDeps`, `sanitizeLine`, `doctorReport` y el relleno de `doctor_integrations.go`).

**Locations:**
- `tooling/management/onboarding.go` (`LastOnboarding` y la validación extraída de `loadOnboarding`);
- `tooling/cli/doctor_integrations.go`;
- `tooling/cli/tui_integrations_view.go` nuevo.

**Execution:** delegada a `hive-build-backend`, en paralelo con T3, en otro worktree cortado después de integrar T4.

**Test approach:** `tdd`.

**Changes:**
1. `LastOnboarding`.
2. La sección.
3. La vista, con ↑↓ y el detalle siempre visible.

**Verification:**
- `go test ./tooling/management -run 'LastOnboarding|Onboarding'` pasa, con estos casos:
  - sin registros da `false`;
  - con dos registros elige el más reciente;
  - un directorio de estado al que se llega por un enlace simbólico lee el registro como válido;
  - un archivo con otro nombre en `onboarding/` se ignora;
  - uno manipulado da error;
  - las pruebas actuales de `loadOnboarding` siguen pasando.
- `go test ./tooling/cli -run Integrations` pasa, con estos casos:
  - `engram` y `agent-browser` falsos en `PATH` escriben un marcador si se ejecutan: se detectan y el marcador no existe;
  - el `SKILL.md` de Context7 en `.agents/skills/find-docs/` se muestra con su ruta;
  - un registro con un paso `manual` de Engram muestra ese estado;
  - pi-subagents muestra solo el estado del registro;
  - cada fila muestra la fuente de `providers.Catalog`, o la fijada para `agent-browser`, y el siguiente paso;
  - sin registro: «No onboarding record yet»;
  - un registro manipulado muestra el error dentro de la vista;
  - `state.json` y el home quedan idénticos.

## T3 — Validación de `## Hive`

- [ ] La vista Project valida la sección `## Hive` del repositorio actual y reporta cada hallazgo.

**Closes:** AC7, AC10.

**Depends on:** T4 (`doctorDeps`, `sanitizeLine`, `doctorReport` y el relleno de `doctor_project.go`).

**Locations:**
- `tooling/cli/doctor_project.go` (`hiveSettingKeys`, `validateHiveSection`);
- `tooling/cli/tui_project_view.go` nuevo;
- `tooling/cli/doctor_project_test.go`, que también lee `content/guidance/global.md` por su ruta en el repositorio.

**Execution:** delegada a `hive-build-backend`, en paralelo con T2, en otro worktree cortado después de integrar T4.

**Test approach:** `tdd`.

**Changes:**
1. El validador, con los comandos `git` y el entorno filtrado de design.md.
2. La prueba de sincronía de AC10.
3. La vista.

**Verification:** `go test ./tooling/cli -run 'HiveSection|ProjectView|HiveSettingKeys'` pasa. Los repositorios de prueba se crean con `git init` en `t.TempDir()`. Casos:
- fuera de Git, con las dos líneas de design.md, incluida la de secciones de workspace;
- sin `AGENTS.md`;
- sin sección, y con la sección duplicada;
- falta `Tracker`, y `Project` vacío;
- `Specs` inexistente;
- `Base branch` inexistente, una rama que existe solo en `origin` (válida), `a..b`, `-x` y `@{-1}` (inválidas, sin llamar a `git` en los dos últimos);
- clave desconocida;
- `Delivery: pr` y `Hive guidance: optional`;
- un `GIT_DIR` heredado que apunta a otro repositorio no cambia el resultado;
- una sección válida dice «Valid»;
- el `AGENTS.md` de prueba queda idéntico byte a byte.

La prueba de sincronía falla si se quita `Review` de `hiveSettingKeys`: se comprueba a mano una vez, antes de cerrar la tarea.

## T5 — Comandos, menú y tamaños

- [ ] `hive doctor` y `hive models` imprimen las secciones, el menú abre las cuatro vistas y todas caben a 80×24 y a 120×40.

**Closes:** AC1, AC8, AC9.

**Depends on:** T1, T2, T3, T4.

**Locations:**
- `tooling/cli/main.go` (comandos y ayuda);
- `tooling/cli/tui_views.go` (`mainMenuItems`);
- el mensaje sin terminal de `hive tui` (`tooling/cli/tui.go`);
- `tooling/cli/tui_test.go` (`menuRowPattern`, la prueba de entradas y la de Esc y Backspace).

**Execution:** hilo principal. Integra los worktrees y toca archivos que comparten todas las vistas.

**Test approach:** `tdd`.

**Changes:**
1. Los comandos con `--home`, `--state-dir` y, en `doctor`, `--project`.
2. Las entradas del menú antes de Quit.
3. El mensaje sin terminal.
4. Las pruebas de menú, comandos y tamaños.

**Verification:** `go test ./tooling/cli -run 'Menu|EscAndBackspace|DoctorCommand|ModelsCommand|Fits|NoTerminal'` pasa, con estos casos:
- **AC1:**
  - el menú tiene 9 entradas en el orden de AC1 (`menuRowPattern` las reconoce todas);
  - Enter abre cada vista nueva;
  - Esc y Backspace vuelven al menú desde cada una.
- **AC8:**
  - `hive doctor --home <tmp>` termina con 0 con un recurso en `drift`;
  - `hive doctor --project <repo>` valida ese repositorio;
  - la salida de `hive doctor` es igual, sección por sección, al texto de las vistas, porque ambas llaman a los mismos renderizadores;
  - `hive models --home <tmp>` imprime las filas y termina con 0;
  - `hive doctor --state-dir <archivo que no es directorio>` termina con distinto de 0;
  - sin terminal, `hive tui` nombra `doctor` y `models`.
- **AC9:**
  - cada vista nueva se mide con `assertFits` y `viewFitProblem` sobre su propia salida, no sobre la pantalla compuesta, a 80×24 y a 120×40, con 6 CLIs, 20 roles, sesiones y rutas largas;
  - en Diagnostics, Models y Project, ↓ y PgDn llegan a la última fila, y ↑ y PgUp vuelven a la primera;
  - en Integrations, un siguiente paso largo se ajusta sin «…» y PgDn llega a su última línea;
  - en Models, a 80 columnas, `hive-design-architecture` se muestra completo.

**Observación adicional:** `go run ./tooling/cli doctor` en este checkout muestra las cinco secciones sin error. Solo lee.

## T6 — Documentación y especificación

- [ ] `deployment-manager.md` documenta las vistas, `hive doctor`, `hive models` y sus límites. La especificación del cambio está completa.

**Closes:** AC11.

**Depends on:** T5.

**Locations:**
- `_support/docs/architecture/deployment-manager.md`, sección «Terminal interface» y una sección nueva de diagnóstico;
- [specs/versioned-installation/spec.md](specs/versioned-installation/spec.md) de este cambio.

**Execution:** hilo principal. El texto depende de lo que T5 deja integrado.

**Test approach:** `check`, con `rg -n "hive doctor|hive models" _support/docs/architecture/deployment-manager.md`.

**Verification:**
- El `rg` devuelve líneas en ambas secciones.
- La lista de vistas del documento nombra las nueve entradas del menú.
- Los límites de sesiones de design.md aparecen en el documento.
- `go test ./tests/...` pasa, por si alguna prueba de contenido revisa ese documento.

## Verificación y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Al cerrar cada tarea | `hive-verify-task` vuelve a correr la verificación de la tarea y da un veredicto por criterio | Parte de la implementación |
| Suite completa | Al terminar T6, en el worktree integrado | `go vet ./...` y `go test -race -timeout 20m ./...` pasan. `go list -m all` es igual al de la base (ningún módulo nuevo) | Parte de la implementación |
| `hive-review-ux` y `hive-verify-change` | Al terminar T6, en paralelo, cada uno con su propio home de prueba y su sesión de `tmux` | Ver el detalle debajo | Parte de la implementación |
| Recorrido del usuario | Después de los dos anteriores | Ver el detalle debajo | Parte de la entrega D5-B |
| `/code-review` | Después del recorrido, sobre el diff completo contra la punta de la base | Hallazgos atendidos o refutados con evidencia | D6-A |

**Detalle de `hive-review-ux` y `hive-verify-change`:**
- **Binario:** el que se compila en el worktree con `go build -o <tmp>/hive ./tooling/cli`. La base `82f2dcc` se compila aparte para comparar el menú.
- **Instalación de prueba con `--home`:** la instalación de Hive del home de prueba se hace con `--home <fixture> --state-dir <fixture>/state`. Así ignora las variables heredadas (`CODEX_HOME`, `XDG_STATE_HOME`, `PI_CODING_AGENT_DIR`, `XDG_CONFIG_HOME` y otras), que la harían escribir en la configuración real.
- **Vistas sin `--home` y con entorno limpio:** con `--home` no se detecta ni se ejecuta nada. Por eso las vistas se abren con `env -i HOME=<fixture> PATH=<fixture>/bin:/usr/bin:/bin CLAUDE_CONFIG_DIR=<fixture>/.claude GROK_HOME=<fixture>/.grok TERM=$TERM <tmp>/hive tui --state-dir <fixture>/state`.
- **Contenido de los homes:**
  - 6 CLIs falsos en `<fixture>/bin` que responden `--version`;
  - la instalación de prueba descrita arriba, con un recurso alterado para que haya `drift`;
  - archivos de sesión de Claude Code y Grok: los «vivos» apuntan al `pid` de un `sleep` que el revisor lanza y termina, y los «muertos» a un `pid` inexistente;
  - un registro de integraciones;
  - un repositorio con `## Hive` incompleto, desde el que se abre la aplicación.
- **Tamaños y temas:** 80×24 y 120×40; fondo oscuro, claro y `NO_COLOR`.
- **`hive-review-ux`:** sigue [UI review criteria](../../../../content/skills/flow-build/references/ui-review-criteria.md).
- **`hive-verify-change`:** recorre AC1 a AC9 y comprueba que `state.json`, el home y el repositorio quedan iguales.

**Detalle del recorrido del usuario:**
- Se corre `go run ./tooling/cli` desde el worktree sobre la configuración real. Es seguro porque las vistas solo leen.
- Recorrido:
  - Diagnostics: versiones reales y sesiones abiertas de Claude Code, incluida la de esta sesión, y una sesión de Grok abierta para el recorrido. Hoy `~/.grok/active_sessions.json` es una lista vacía, así que el formato de `opened_at` solo se confirma con una sesión viva;
  - Models: Claude y OpenCode;
  - Integrations;
  - Project, en este repositorio.
- Resultado esperado: `~/.config/hive` sin cambios y `git status` limpio en el repositorio.

## Estado de la revisión y avance

- **Revisión del plan (2026-09-29):**
  - Dos `review-plan`, uno de backend e integración y otro de interfaz de terminal, revisaron la primera versión.
  - Encontraron criterios con verificación incompleta y varios problemas: el filtro de Installation marcaba los recursos `not_installed`, faltaba aislar las pruebas de la configuración real, y los gates de interfaz no podían observar versiones con `--home`.
  - También señalaron el reparto de `doctor.go` entre tareas paralelas, el presupuesto de filas, el patrón de carga y la ejecución de `git` y de `--version`.
  - Todos se aplicaron en la segunda versión.
- **Segunda ronda, la única permitida (2026-09-29):** los mismos dos revisores confirmaron las correcciones y señalaron cinco puntos más, que se aplicaron tal como los propusieron:
  - el detalle de Integrations se desplaza en vez de cortarse;
  - la variante de OpenCode se muestra en la columna de esfuerzo;
  - la columna de rol no se recorta;
  - la instalación de prueba usa `--home` y las vistas se abren con `env -i`;
  - `LastOnboarding` resuelve la ruta canónica.
  
  También se precisaron la regla de extracción de AC10 y el caso de un archivo de sesión ilegible. Por ser propuestas de los propios revisores aplicadas sin cambios, no hay otra ronda.
- **Límites de la revisión:** no se ejecutó ninguna prueba ni se observó el comportamiento real. La revisión cubre el backend, la integración y la interfaz de terminal, las dos áreas afectadas.
- **Ajustes antes del build (2026-09-29):** el análisis previo agregó al recorrido una sesión de Grok (H3) y excluyó la validación de secciones de workspace, con su mensaje (H4). El usuario los aceptó y autorizó el build.
- **Avance:** T1 y T4 en curso.
- **Siguiente paso:** integrar T4 y lanzar T2 y T3.
