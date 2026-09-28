# Tareas

**Commit base:** `32eb03f`. Worktree `.claude/worktrees/gh-46-tui`, rama local `feat/gh-46-tui`; se publica directo a `rebuild/harness-engineering` (D7-A). La suite completa con `-race` pasó en `3b5a254`; `32eb03f` solo agrega el cierre del cambio de voz. D17-A (2026-09-28): `hive` sin argumentos y `hive tui` abren la interfaz.

## Orden y ejecución

- **Primero, T1:** separa las preguntas de la lógica.
- **Después, T2:** trae la dependencia, la entrada, el `prompter` de `huh` y el menú.
- **Luego, T3 y T4:** las pantallas. Las hace el mismo hijo en secuencia, porque todas escriben en `tooling/cli`.
- **En el hilo principal:** T5 (documentación) y T6 (revisión de la interfaz, prueba en vivo y recorrido del usuario).

Cada hijo recibe `AGENTS.md`, este cambio y [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md). T2 lo necesita por la dependencia.

### T1 — Interfaz `prompter` para las preguntas

- [x] `install`, `update` y `voice` obtienen las elecciones a través de `prompter`, y `installTerminal` es su implementación de texto.
  - **Evidencia:** los hashes de `install_test.go`, `update_test.go`, `voice_test.go` y `bootstrap_test.go` son idénticos antes y después, y `go test -race ./tooling/cli` y `go vet ./...` pasan.
  - **Resultado:** `prompter` está en `install.go:45`, y existen `updateWith` y `voicePlanWith`.
  - **Desviación:** `runInstallFlow` conserva su firma, porque `bootstrap_test.go` la llama directamente. T2 agrega `runInstallFlowWith`, que recibe el `prompter`, para que la interfaz entre por la verificación del paquete y la comprobación de operaciones pendientes.

**Closes:** AC8.

**Depends on:** ninguna.

**Locations:**
- `tooling/cli/install.go`: `installTerminal`, `runInstallFlow`, `runOnboardingWizard`, `expandToRequiredHosts`, `previewInstallOnboarding`, `handlePendingInstallOperation`, `selectInstallerHosts`, `selectProviderRequests`, `readProviderVersion`, `confirmInstall`.
- `tooling/cli/bootstrap.go:234`: pasa su `prompter`.
- `tooling/cli/update.go` y `tooling/cli/voice.go`: nuevas `updateWith` y `voicePlanWith`, que reciben el `prompter`.

**Execution:** delegada a `backend-developer`, porque la interfaz ya está fijada en el diseño.

**Test approach:** characterization.

**Changes:** según [separar las preguntas de la lógica](design.md#separar-las-preguntas-de-la-lógica). La regla «sin terminal» se queda en `installTerminal`. Las pruebas de `install_test.go`, `update_test.go`, `voice_test.go` y `bootstrap_test.go` son la caracterización: tienen que pasar antes y después sin modificarse.

**Verification:**
- Se anota el `sha256` de esos archivos de prueba antes y después, y debe coincidir.
- `go test -race ./tooling/cli` pasa.
- `rg -n "type prompter" tooling/cli` devuelve la interfaz.

### T2 — Dependencia, entrada, `prompter` de `huh` y menú

- [x] `hive` sin argumentos y `hive tui` abren el menú en una terminal o en modo accesible; `hive` conserva el error de uso sin terminal; la interfaz ofrece recuperar una operación pendiente al abrir; y el núcleo sigue sin dependencias externas.
  - Ronda 1, AC1, AC7 y AC10 no cumplidos:
    - una respuesta inválida seguida del fin de la entrada hace caer el programa con un índice fuera de rango (salida 2);
    - Esc no cancela;
    - `deps_test` no sigue las importaciones de `tricell-hive/...`;
    - faltan pruebas del orden del menú, de la línea de estado, del subcomando `tui` y de `hive` con `HIVE_ACCESSIBLE`.
  - **Ronda 2:** AC1, AC7 y AC10 cumplidos. `review-task` repitió en `tmux` a 80×24 las entradas que hacían caer el programa y la tecla Esc. También rompió a propósito el código de las formas que antes nadie detectaba, y ahora cada una hace fallar una prueba. `go mod tidy` lo corrió el hilo principal al cerrar la tarea.
  - **Pendientes que pasan a T4:**
    - Esc dentro del filtro de una lista cierra toda la pantalla en vez de limpiar el filtro;
    - ninguna prueba falla si se quita `.WithKeyMap(formKeyMap)` de `runForm`;
    - la prueba de `hive` sin argumentos lee el estado real, solo para leer, y conviene darle un home de prueba.

**Closes:** AC1, AC7, AC10 (el menú y el `prompter`).

**Depends on:** T1.

**Locations:**
- `go.mod` y `go.sum`;
- `tooling/cli/main.go`: la rama sin argumentos, `tui` en el despacho y la ayuda, sin superar las 40 líneas de `run`;
- `tooling/cli/main_test.go`: `TestRunUsageErrors` fija `HIVE_ACCESSIBLE`;
- nuevos: `tooling/cli/tui.go`, `tooling/cli/tui_prompter.go` y `tooling/cli/tui_test.go`;
- nuevo: `tooling/management/deps_test.go`.

**Execution:** delegada a `backend-developer`.

**Test approach:** tdd.

**Changes:** según [punto de entrada](design.md#punto-de-entrada-d14-a), [el `prompter` de `huh`](design.md#el-prompter-de-huh) y [la interfaz](design.md#la-interfaz).
- `go get charm.land/huh/v2@v2.0.3`, y después `go mod tidy` y `go mod verify`.
- Comprobar en la documentación de la versión fijada cómo se detecta el fondo para el tema, y anotarlo aquí.
  - **Anotado el 2026-09-28:** en `huh` v2.0.3 el tema por defecto ya es `ThemeCharm(hasDarkBg)`, pero solo detecta el fondo dentro del bucle de Bubble Tea (`tea.BackgroundColorMsg`). La interfaz lo resuelve antes con `lipgloss.HasDarkBackground(in, out)` de `charm.land/lipgloss/v2` v2.0.1 (`query.go:83`), que por eso es una dependencia directa. Con `NO_COLOR` usa `ThemeBase`.

Pruebas, todas en modo accesible con entrada guionada:
- elegir Quit sale con 0;
- el fin de la entrada en el menú sale con 0;
- una pantalla cancelada vuelve al menú;
- `ErrUserAborted` se traduce en cancelación, con una prueba unitaria del `prompter`;
- `openInterface` sin terminal y sin `HIVE_ACCESSIBLE` devuelve el error de uso de hoy;
- una operación pendiente en un estado de prueba se ofrece y se recupera;
- `deps_test.go` recorre con `go/parser` las importaciones de `tooling/management` y `tooling/distribution`.

**Verification:** `go test -race ./tooling/cli ./tooling/management`, `go mod verify` y `awk '/^func run\(/,/^}/' tooling/cli/main.go | wc -l` con 40 o menos. Las pruebas nuevas fallan en la base.

### T3 — Pantallas Install CLIs y Remove CLIs

- [x] Instalar y quitar CLIs desde la interfaz produce los mismos archivos y el mismo estado que los comandos equivalentes, con sus estados vacíos y su cancelación.
  - Ronda 1, AC2 y AC3 no cumplidos: en la terminal normal los `MultiSelect` esconden siempre su última opción (`huh` v2.0.3 resta la línea del título sin una altura explícita). Además: sin pruebas para las etiquetas de estado ni para la lista de archivos de Remove, la normalización de `state.json` más amplia de lo necesario, el texto de `--dry-run` que la interfaz no tiene, y las pruebas nuevas suben la suite con `-race` a 857 s, por encima del límite de 10 minutos.
  - **Ronda 2:** AC2, AC3, AC10 y AC11 cumplidos.
    - En `tmux`, a 80×24 y 100×45, las listas muestran todas sus opciones.
    - Con el árbol `content/` real, la interfaz y los comandos dejan homes y `state.json` idénticos: 140 entradas en instalar y 105 en quitar.
    - Las roturas deliberadas del código hacen fallar las pruebas nuevas.
    - `go test -race ./tooling/cli` pasa en 555 s con el límite por defecto.
  - **Pendientes que pasan a T4 y T5:**
    - la recuperación dentro de Install no nombra `--state-dir`;
    - el margen de tiempo de la suite es de solo 45 s: T5 documenta `-timeout` explícito.

**Closes:** AC2, AC3, AC10 (en estas pantallas), AC11 (en estas pantallas).

**Depends on:** T2.

**Locations:** nuevos `tooling/cli/tui_hosts.go` y `tooling/cli/tui_hosts_test.go`, y `removeFlow` con su resumen propio.

**Execution:** delegada, el mismo hijo que T2, en secuencia.

**Test approach:** tdd.

**Changes:** según [la interfaz](design.md#la-interfaz). Install entra por `runInstallFlow`. Pruebas en modo accesible, sobre homes de prueba; cada una compara los archivos y `state.json` con un home gemelo hecho con el comando equivalente:
- instalar Codex y Claude;
- instalar Grok con la expansión obligatoria a Claude;
- rechazar la confirmación (nada cambia y vuelve al menú);
- quitar Claude con una voz activa (se quita su bloque de voz);
- Remove sin CLIs registrados (su mensaje);
- una fuente sin catálogo (su error y vuelta al menú).

**Verification:** `go test -race -timeout 20m ./tooling/cli -run 'InstallScreen|RemoveScreen|ConfigureForm|RunBareArgs|Menu'`. Las pruebas nuevas fallan en la base. `go test -race ./tooling/cli` debe seguir bajo el límite por defecto de 10 minutos.

### T4 — Pantallas Status, Update, Releases y Voice

- [x] Status, Update, Releases (con la vuelta a una release) y Voice funcionan con vista previa y confirmación, dan los mismos resultados que sus comandos, y tienen sus estados vacíos y su cancelación.
  - Ronda 1, AC6 no cumplido: la lista de voces calcula su altura por número de opciones y no por líneas; con las descripciones largas esconde a Jarvis y a Mentor en una terminal normal. Además:
    - las pruebas de Status y Voice no fallan si los valores están mal;
    - la pista sobre bajar de versión se añade a cualquier error de la vuelta a una release;
    - Update dentro de la interfaz muestra el texto de `--dry-run`;
    - faltan los encabezados de pantalla del diseño.
  - **Ronda 2:** AC4, AC5, AC6, AC10 y AC11 cumplidos.
    - En `tmux`, a 80×24 y 100×45, la lista de voces muestra sus cuatro opciones en una sola línea.
    - Update, la vuelta a una release y Voice dan homes y estado idénticos a sus comandos con el árbol real.
    - La salida de `hive update` fuera de la interfaz es idéntica a la de la base.
  - **Añadido por el hilo principal:** «nonportable personal path in» como error de validación de una release, con su prueba (falló antes y pasa después).
  - **Suite con `-race`:** `go test -race -timeout 20m ./...` pasa en los 16 paquetes; `tooling/cli` tardó 608 s en paralelo con otras cargas, lo que confirma que hace falta el `-timeout` explícito documentado.

**Closes:** AC4, AC5, AC6, AC10 (en estas pantallas), AC11 (en estas pantallas).

**Depends on:** T3.

**Locations:** nuevos `tooling/cli/tui_screens.go` y `tooling/cli/tui_screens_test.go`, y `rollbackFlow`.

**Execution:** delegada, el mismo hijo, después de T3.

**Test approach:** tdd.

**Changes:** según [la interfaz](design.md#la-interfaz). Pruebas en modo accesible, sobre homes de prueba:
- **Status:** muestra la release, la versión, el `drift` y la voz, y deja `state.json` y el home idénticos; sin CLIs, su mensaje.
- **Update:** desde un repositorio Git temporal con un commit nuevo, con el mismo resultado que `hive update`; con una revisión que empieza con `-`, el error y la vuelta al menú.
- **Releases:** lista con la instalada marcada; volver a una release anterior deja sus archivos; sin releases, su mensaje; elegir la instalada, «Already installed».
- **Voice:** activa Jarvis con `sir` y después la apaga, dejando los archivos idénticos.
- **Todas:** rechazar la confirmación no cambia nada y vuelve al menú.

**Verification:** `go test -race -timeout 20m ./tooling/cli -run 'StatusScreen|UpdateScreen|ReleasesScreen|VoiceScreen|Rollback|ScreenDecline|ReleaseLabel|ReleaseSelect|ReleasesSelectKeyMap'` y `go test ./tooling/cli`. Las pruebas nuevas fallan en la base.

### T5 — Documentación

- [x] `deployment-manager.md` documenta la interfaz, `hive tui`, el modo accesible y la dependencia.
  - **Evidencia:** `review-task` da AC9 cumplido, con cada afirmación contrastada con el código y con el binario en modo accesible. Con sus observaciones se añadieron:
    - las dependencias directas `lipgloss` y `bubbles`;
    - cuándo hace falta `--source`;
    - Esc y Ctrl-C en el menú;
    - `NO_COLOR` y el tema según el fondo;
    - Esc en Releases sin filtro;
    - Back en la confirmación de Install;
    - que `--state-dir` solo aparece si se pasó uno.
  - **Verificación con `-race`:** 568 s con `-timeout 20m`, así que el `-timeout` explícito documentado está justificado.

**Closes:** AC9.

**Depends on:** T4.

**Locations:** `_support/docs/architecture/deployment-manager.md`.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:**
- la entrada sin argumentos y `hive tui` con sus opciones;
- las pantallas;
- la cancelación;
- la recuperación al abrir;
- `HIVE_ACCESSIBLE`;
- que `hive install` sigue siendo el instalador de texto;
- la dependencia como excepción explícita, con su efecto en la construcción de releases;
- la actualización de la frase «using only the standard library».

**Verification:** `rg -n "HIVE_ACCESSIBLE|charm.land/huh|hive tui" _support/docs/architecture/deployment-manager.md` devuelve la sección nueva, y `go run ./tooling/cli --help` menciona `hive tui` y `HIVE_ACCESSIBLE`.

### T6 — Revisión de la interfaz, prueba en vivo y recorrido del usuario

- [/] `review-ux` y `sdd-verify` aprueban la interfaz sobre el binario compilado en una terminal real, y el usuario la recorre y la aprueba.
  - **`sdd-verify` (2026-09-28):** AC1 a AC11 cumplidos, en `tmux` y en modo accesible, comparando cada pantalla con su comando sobre homes gemelos. Incluye una operación pendiente real, provocada matando `apply` con `kill -9`, que la interfaz ofreció recuperar y recuperó. Sin verificar en vivo: la recuperación de un paso opcional interrumpido, que cubre una prueba unitaria.
  - **`review-ux`, primera ronda:** cinco hallazgos, uno de ellos bloqueante.
    - F1: contraste casi nulo de las opciones sobre fondo oscuro, porque `huh` v2.0.3 invierte sus pares de colores claro/oscuro.
    - F2: filas de Status partidas a 80 columnas.
    - F3: la fila de la release instalada no cabía.
    - F4: un error crudo en Voice sin voces.
    - F5: la voz activa perdía su configuración al volver a elegirla.
  - **`review-ux`, segunda ronda:** F1 a F5 resueltos. Aparecieron dos puntos medios nuevos, corregidos después:
    - N1: la inversión completa del tema bajó el contraste de los títulos; se sustituyó por ajustes solo en los colores invertidos.
    - N2: texto claro sobre fucsia en el botón enfocado.
  - **Prueba de contraste:** `TestThemeContrastMeetsWCAGAA` calcula el contraste WCAG de diez estilos en los dos temas. Excepción documentada: el título conserva el índigo de `huh` y da 4,36 a 1 frente a `#1e1e1e`.
  - **Suite:** `go test -race -timeout 20m ./...` pasa en los 16 paquetes; `tooling/cli` tarda 585 s.
  - **Pendiente:** el recorrido del usuario.

**Depends on:** T5.

**Execution:** hilo principal. Primero prepara los datos de prueba y después lanza los dos hijos en paralelo, cada uno con su propio home.

**Changes:**
- **Datos de prueba:** un script desechable en el scratchpad de la sesión prepara un home con seis CLIs, 100 releases retenidas (generadas desde commits de un clon temporal), una operación pendiente y nombres largos. También deja un home vacío.
- **`review-ux`:**
  - Recibe [UI review criteria](../../../../content/skills/flow-build/references/ui-review-criteria.md) traducidos a terminal:
    - usa `tmux` a 80×24 y 120×40, con `capture-pane -e` para conservar los colores;
    - nada recortado ni partido a 80 columnas, y las listas caben en 24 filas;
    - contraste de los colores del tema frente a un fondo claro y uno oscuro;
    - solo teclado, con `NO_COLOR` y en modo accesible.
  - Revisa las siete pantallas con sus estados, incluidos el vacío, el error, la cancelación y las listas largas.
  - **Comparación de contenido:** no hay una versión anterior con interfaz, así que el contenido se compara contra lo que imprimen `hive install`, `update`, `voice set`, `status` y `releases`. La maquetación queda sin versión base, y así se anota.
- **`sdd-verify`:** recorre cada pantalla de principio a fin con `tmux send-keys` o `expect`, incluida la recuperación al abrir, y además un recorrido completo en modo accesible.
- **Hallazgos:** se corrigen los de gravedad alta que introdujo el cambio y se vuelve a revisar.
- **Recorrido del usuario:** un bloque por pantalla, con el comando para compilar el binario, cómo abrirlo con `hive tui --home <home de prueba> --state-dir <estado de prueba> --source <worktree>`, qué debe ver y cómo borrar el home al terminar.

**Verification:** los veredictos de los dos hijos y la aprobación del usuario, anotados aquí con su fecha.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Al cerrar cada tarea y al final, en local | `go vet ./...`, `go test -race ./...` y `go mod verify` en verde; no hay CI | Parte de la implementación |
| Verificación por tarea | Tras T1, T2, T3, T4 y T5 | `review-task` contra sus `AC<n>` | Parte de la implementación |
| Revisión de la interfaz y prueba en vivo | T6 | `review-ux` y `sdd-verify` con `tmux` sobre el binario | Parte de la implementación (interfaz visible) |
| Recorrido del usuario | T6, antes de la revisión de código | El usuario abre `hive tui` sobre un home de prueba | Parte del plan; decide el usuario |
| Revisión de código | Antes del push | `/code-review` (D8-A) | Ya decidida |
| Usar la interfaz sobre la configuración real | Después de integrar | El usuario ejecuta `hive` | Decisión del usuario |

## Estado de la revisión y avance

**Revisión del plan** (2026-09-28): versión revisada `0149afee6783`. Dos revisores `review-plan` en paralelo, como rol nativo de Claude Code y de solo lectura.

- **Arquitectura en Go:** cuatro hallazgos que bloqueaban, corregidos con la propuesta del revisor.
  - En modo accesible, `huh` lee por adelantado y descarta respuestas; se usa un lector de un byte por lectura.
  - `huh` no señala el fin de la entrada ni la cancelación; lo registra ese lector, y Quit es la opción por defecto.
  - No había forma de abrir la interfaz sobre un home de prueba; se agrega `hive tui` con sus opciones.
  - Install se saltaba la verificación del paquete y la comprobación de operaciones pendientes; ahora entra por `runInstallFlow`.
  - Se incorporaron las sugerencias sobre las funciones que faltaban en T1, la prueba de uso aislada del entorno, el recorrido de importaciones con `go/parser`, el resumen propio de Remove, la recuperación con `handlePendingInstallOperation` y la redacción de la especificación.
- **Experiencia en la terminal:** cinco hallazgos que bloqueaban, corregidos.
  - Cancelar o fallar dentro de una pantalla vuelve al menú (AC10 nuevo).
  - Estados vacíos y límite con menú fijo (AC11 nuevo).
  - La especificación se contradecía con el diseño sobre `HIVE_ACCESSIBLE`.
  - La revisión se traduce a terminal con `tmux`.
  - La comparación de contenido se hace contra los comandos de texto.
  - Se incorporaron el tema según el fondo y `NO_COLOR`, el resumen impreso seguido de la confirmación, la lista de releases filtrable, que ningún CLI venga marcado de entrada, la línea de espera, el script de datos de prueba, la ayuda y la comprobación de que Status no escribe.
- **Cambio del usuario durante la revisión:** el subcomando explícito se llama `hive tui` en vez de `hive ui`.
- **Sin nueva ronda:** todas las correcciones aplican propuestas de los revisores. Queda sin verificar cómo se comporta `huh` con un lector de pantalla real.

**Avance:** plan listo para implementar; ninguna tarea empezada.

**Siguiente paso:** `flow-build`, empezando por T1.
