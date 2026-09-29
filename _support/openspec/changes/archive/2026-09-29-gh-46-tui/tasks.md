# Tareas

**Commit base del replanteo:** `dd146c0`, en el worktree `.claude/worktrees/gh-46-tui`, rama local `feat/gh-46-tui`. Se publica directo a `rebuild/harness-engineering` (D4-A).

**Decisiones del 2026-09-28, tras el recorrido del usuario:**
- D1-A: aplicación propia con Bubble Tea, sin `huh` y sin `HIVE_ACCESSIBLE`.
- D2-A: Status se integra en CLIs.
- D3-A: altas y bajas en dos pasos.
- D4-A: la misma entrega de antes.

## Orden y ejecución

- **T7 a T10**, en secuencia, delegadas a un mismo hijo `backend-developer`. Todas escriben en `tooling/cli` y T8 a T10 dependen del esqueleto de T7.
- **Después de cada una:** `review-task` contra sus `AC<n>`.
- **En el hilo principal:** T11 (documentación y especificación) y T12 (revisión de la interfaz, prueba en vivo, recorrido del usuario y revisión de código).

Cada hijo recibe:
- `AGENTS.md`;
- este cambio (`proposal.md`, `design.md` y `tasks.md`);
- [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md), por las dependencias de T10.

## Historial: primera versión (T1–T6, formularios secuenciales)

T1 a T5 quedaron verificadas y T6 pasó `review-ux` y `sdd-verify`. El recorrido del usuario pidió cambios de diseño, y T7 a T12 los atienden. Los commits `aec5f41..dd146c0` se conservan.

| Tarea | Resultado | Qué sigue valiendo |
| --- | --- | --- |
| T1 | Interfaz `prompter`; `installTerminal` es la implementación de texto. Pruebas de texto sin cambios. | La caracterización. El `prompter` se reevalúa en T10. |
| T2 | Menú con `huh`, `hive tui`, recuperación al abrir, `deps_test.go`. | `hive tui` con sus opciones, `interfaceStdio`, `deps_test.go` y la detección del fondo. |
| T3 | Install y Remove con homes gemelos idénticos al comando. | `removeFlow`, `showRemoveSummary` y los auxiliares de homes gemelos. |
| T4 | Status, Update, Releases, Voice y `rollbackFlow`. | `rollbackFlow`, las etiquetas de releases y voces, y `statusLine`. |
| T5 | Documentación en `deployment-manager.md`. | Se reescribe en T11. |
| T6 | `review-ux` (F1–F5, N1 y N2 resueltos) y `sdd-verify` (AC1–AC11 de la primera versión cumplidos). | La prueba de contraste WCAG y el script de datos de prueba `_support/workspace/2026-09-28-gh-46-tui/make-fixtures.sh`. |

**Recorrido del usuario (2026-09-28):** pidió los cuatro cambios del objetivo. Captura: `~/Library/Application Support/CleanShot/media/media_xuQsSn6wME/CleanShot 2026-09-28 at 13.19.09@2x.png`.

## Tareas del replanteo

### T7 — Esqueleto de la aplicación, menú y recuperación al abrir

- [x] `hive` y `hive tui` abren la aplicación en la pantalla alterna, con menú, línea de estado, barra de ayuda, teclas globales, aviso de tamaño mínimo y vista de recuperación. El menú lleva a vistas provisionales («Not implemented yet»), que T8 y T9 reemplazan.
  - **`review-task` (2026-09-28):** AC1, AC2 (fuera de una escritura) y AC9 (menú, vistas genéricas y aviso) cumplidos.
    - **Pruebas:** con `-race`, en 170 s; `go vet` limpio; `sha256` de las cuatro pruebas de texto iguales a `dd146c0`.
    - **En `tmux`:** pantalla alterna activa y después apagada, cinco formas de salir sin rastro, código de salida 0, y aviso de tamaño con el estado conservado.
  - **Desviaciones:**
    - la ruta secuencial (`openInterface`, `runMenu`, `checkPendingOnOpen`) sigue en `tui.go` sin ser alcanzable, porque las pruebas protegidas la llaman; T10 la borra;
    - `x/ansi` como dependencia directa ([dependencias](design.md#dependencias)).
  - **Pendientes que pasan a T10:**
    - `TestUsage…` fija `HIVE_ACCESSIBLE` con `t.Setenv`, y T10 quita esos `Setenv` para que la búsqueda de AC10 quede limpia;
    - borrar `TestLegacyHuhThemeContrastMeetsWCAGAA`.
  - **Menores:** ninguna prueba exige la «…» al recortar una fila (T8 y T9 la afirman en sus filas), y por debajo de unas 32 columnas el aviso se recorta.

**Closes:** AC1, AC2 (fuera de una escritura), AC9 (menú, vistas genéricas y aviso de tamaño).

**Depends on:** ninguna.

**Locations:**
- `tooling/cli/tui.go`: `openInterface`, `runBareInterface`, `runInterfaceCommand`, `checkPendingOnOpen` y `interfaceStatusLine`;
- nuevos `tooling/cli/tui_app.go` (modelo raíz, pila, teclas, tamaño), `tooling/cli/tui_views.go` (menú, resumen y confirmación, aviso, recuperación) y `tooling/cli/tui_theme.go` (estilos de `lipgloss`);
- `tooling/cli/main.go`: la ayuda sin `HIVE_ACCESSIBLE`;
- `tooling/cli/tui_test.go`: el conductor de pruebas y las pruebas nuevas.

**Execution:** delegada a `backend-developer`. El diseño fija las teclas, la pila y la entrada.

**Test approach:** tdd.

**Changes:** según [aplicación](design.md#aplicación-d1-a), [vistas](design.md#vistas) (menú, resumen y confirmación, aviso, recuperación) y [pruebas](design.md#pruebas).
- La interfaz secuencial deja de ser alcanzable desde `hive` y `hive tui`. Su código se borra en T10.
- La recuperación al abrir reutiliza la lógica de `handlePendingInstallOperation`, sin su pregunta de texto.
- El tema pasa a estilos propios. `TestThemeContrastMeetsWCAGAA` mide esos estilos en los dos fondos, con la excepción documentada del título si se mantiene el índigo.

Pruebas nuevas:
- la entrada y la salida abren y cierran la pantalla alterna;
- el menú tiene cinco entradas en orden;
- Esc y Backspace en una vista provisional vuelven al menú, y Backspace en el menú no hace nada;
- Ctrl-C sale con 0 desde el menú y desde una vista;
- con menos de 80×24, el aviso; al volver a 80×24, la vista conserva su estado;
- el menú y las vistas genéricas caben a 80×24 y a 120×40;
- con `NO_COLOR`, la salida no tiene secuencias de color y el cursor es `>`;
- el conductor falla por tope de iteraciones ante un `Cmd` que se reprograma, en lugar de colgarse;
- sin terminal, `hive` da el error de uso y `hive tui` su mensaje;
- una operación pendiente en un estado de prueba se ofrece y se recupera con el mismo resultado que `hive recover`.

**Verification:**
- `go test -race -timeout 20m ./tooling/cli -run 'App|Menu|Keys|AltScreen|TooSmall|Recover|Usage|Theme'` pasa, y las pruebas nuevas fallan en `dd146c0`.
- `go vet ./...` pasa.
- `review-task` abre el binario en `tmux` a 80×24, entra y sale de una vista con Esc y con Backspace, cierra con Ctrl-C y comprueba con `capture-pane` que no queda nada impreso.

### T8 — Vista CLIs: estado, casillas, diff y «Uninstall all»

- [x] La vista CLIs muestra el estado por CLI y aplica altas, bajas o ambas en dos pasos, y «Uninstall all». Los resultados son idénticos a los de los comandos.
  - **`review-task` (2026-09-28):** AC2 (durante una escritura), AC3, AC4, AC5, AC8 y AC9 en esta vista, cumplidos.
    - **Pruebas:** 26 pruebas `HostsView` con `-race`; la suite completa con `-race` pasa (`tooling/cli` 702 s); `sha256` de las pruebas de texto iguales.
    - **En `tmux`:** altas y bajas en dos pasos, Esc en el aviso intermedio, «Uninstall all», un rechazo, `kill -INT` y cambios de tamaño durante la aplicación. Cada resultado es igual en archivos y `state.json` al de su comando.
  - **Añadido por el hilo principal antes de la revisión:** Esc y Backspace en el aviso «Step 1 of 2 done» se detienen tras quitar, en lugar de seguir con la instalación.
  - **Aclaraciones:**
    - `hive install` también recorta los espacios de la versión, así que no hay diferencia con el comando.
    - Con `--home` no se detecta ningún CLI por `PATH`, así que un CLI quitado desaparece de la vista (la regla de filas del diseño).
  - **Pendientes que pasan a T9:**
    - N1: el aviso de tamaño dice «press ctrl+c to quit» durante una escritura, cuando Ctrl-C se ignora;
    - N2: la nota de operación pendiente se agrega después de calcular el alto del mensaje (`tui_hosts_view.go:779`), y podría cortar líneas sin «…».

**Closes:** AC2 (durante una escritura), AC3, AC4, AC5, AC8 (en esta vista), AC9 (en esta vista).

**Depends on:** T7.

**Locations:**
- nuevos `tooling/cli/tui_hosts_view.go` y `tooling/cli/tui_hosts_view_test.go`;
- `tooling/cli/tui_hosts.go`: `removeFlow` partido en plan y aplicación, y `showRemoveSummary`;
- `tooling/cli/install.go`: solo si hace falta exponer una pieza de la instalación sin cambiar su comportamiento.

**Execution:** delegada, el mismo hijo, después de T7.

**Test approach:** tdd.

**Changes:** según [vistas](design.md#vistas) (CLIs, capacidades opcionales) y [diff de CLIs](design.md#diff-de-clis). Pruebas con el conductor, sobre homes de prueba y comparadas con homes gemelos:
- las filas muestran la casilla, la etiqueta, la release corta, la versión y el `drift` que da `hive status` para el mismo home;
- alta de Codex y Claude;
- alta con una capacidad opcional que pide versión, con un adaptador de prueba, contra `hive install` con las mismas elecciones;
- alta de Grok con la expansión a Claude;
- marcar un CLI con instalación antigua: el resumen dice que migra, y el resultado es igual al de `hive install`;
- un paquete alterado: el alta falla con el mensaje de `hive install` y no cambia nada;
- baja de Claude con una voz activa;
- baja de Codex y alta de Pi a la vez, contra `plan remove`+`apply` seguido de `install --hosts pi`;
- rechazar el segundo paso deja solo la baja, y un fallo del segundo paso muestra «Removed: codex.» y su error;
- «Uninstall all» arranca en Cancel, ignora `y`, y al confirmar deja cero CLIs y la voz apagada;
- con el `Cmd` del `apply` retenido: `spinner` visible, y Esc, Backspace y Ctrl-C no hacen nada; al liberarlo, la vista muestra el resultado;
- un resumen largo se desplaza con ↓ y PgDn;
- el programa real con `-race` aplica una baja, y descarta una `InterruptMsg` enviada durante la escritura;
- abrir y salir no cambia `state.json`;
- sin cambios, «Nothing to apply»;
- un error del plan aparece como aviso en la vista;
- con seis CLIs, nombres y releases largas, la vista cabe a 80×24 y a 120×40 y el cursor llega a la última fila.

**Verification:**
- `go test -race -timeout 20m ./tooling/cli -run 'HostsView'` pasa, y las pruebas nuevas fallan en `dd146c0`.
- `review-task` repite en `tmux`, a 80×24, las altas y bajas juntas con los datos de `make-fixtures.sh`.

### T9 — Vistas Update, Releases y Voice

- [x] Update, Releases y Voice funcionan en su vista, con ajustes en el lugar, resumen y confirmación. Sus resultados son idénticos a los comandos.
  - **`review-task` (2026-09-28):** AC6, AC7, AC8 y AC9 en estas vistas, cumplidos, y también los pendientes N1 y N2 de T8.
    - **Pruebas:** 54 con `-race`; las pruebas de texto de `update` y `voice` pasan con el mismo `sha256`, así que `planUpdate` no cambió la salida de `hive update`.
    - **En `tmux`:** Update, la vuelta a una release y la voz (set y Off) dan archivos y estado idénticos a los comandos literales. Las 101 releases se recorren hasta la última, y al salir no queda nada impreso.
  - **Decisiones anotadas en el diseño:** con Off solo se ve la fila Voice, y los campos de texto largos se desplazan sin «…».
  - **Pendientes que pasan a T10:**
    - H1: `assertFits` mide la pantalla ya recortada por la raíz, así que no puede fallar por el cuerpo de una vista; hay que medir `View(viewCtx)` contra el alto disponible;
    - H4: Backspace durante la planificación de Voice, con Name enfocado, abandona el plan sin aviso;
    - H5: el mensaje anterior de Voice sigue visible mientras se editan los valores.

**Closes:** AC6, AC7, AC8 (en estas vistas), AC9 (en estas vistas).

**Depends on:** T8.

**Locations:**
- nuevos `tooling/cli/tui_update_view.go`, `tooling/cli/tui_releases_view.go`, `tooling/cli/tui_voice_view.go` y sus `_test.go`;
- `tooling/cli/update.go`: una función extraída, compartida con `updateWith`, según [contexto verificado](design.md#contexto-verificado), sin cambiar la salida de `hive update`;
- `tooling/cli/tui_screens.go`: `rollbackFlow` partido en plan y aplicación.

**Execution:** delegada, el mismo hijo, después de T8.

**Test approach:** tdd.

**Changes:** según [vistas](design.md#vistas). Pruebas con el conductor:
- **Update:** desde un repositorio Git temporal con un commit nuevo, igual que `hive update`; una revisión que empieza con `-` muestra el error en la vista; Backspace en un campo borra y no vuelve atrás.
- **Releases:** la instalada está marcada; volver a una anterior deja sus archivos; sin releases, su mensaje; elegir la instalada da «Already installed»; `/` abre el filtro y filtra por subcadena; Esc con el filtro activo lo limpia y sin filtro vuelve al menú; con 100 releases, el cursor llega a la última a 80×24 y a 120×40.
- **Voice:** la vista se carga con la voz activa; ←→ cambia Address e Intensity; Name aparece solo con `name`; activar Jarvis con `sir` y después Off deja los archivos idénticos a `hive voice set` y `off`; sin cambios, su mensaje.
- **Las tres:** rechazar no cambia nada y deja el mensaje dentro de la vista; al aplicar con éxito, la vista muestra el resultado; caben a 80×24 y a 120×40.

**Verification:**
- `go test -race -timeout 20m ./tooling/cli -run 'UpdateView|ReleasesView|VoiceView'` pasa, y las pruebas nuevas fallan en `dd146c0`.
- Las pruebas de texto de `update_test.go` y `voice_test.go` pasan sin modificarse, con el mismo `sha256` antes y después.

### T10 — Retirar `huh` y la interfaz secuencial

- [x] No queda código, prueba ni dependencia de `huh` ni de `HIVE_ACCESSIBLE`. `bubbletea` es una dependencia directa.
  - **`review-task` (2026-09-28):** AC10 cumplido.
    - **Búsquedas y módulos:** el `rg` no devuelve nada; `go mod tidy -diff` no muestra diferencias; `go list -m all` solo pierde 9 módulos (`huh`, `catppuccin`, `x/conpty`, `x/errors`, `x/exp/*`, `x/xpty`, `creack/pty`, `hashstructure`) y no gana ninguno.
    - **Pruebas:** la suite completa con `-race` pasa (`tooling/cli` 771 s), y el `sha256` de las cuatro pruebas de texto es igual.
    - **Cambios de estructura:** `prompter` volvió a ser `installTerminal`, y `updateWith` quedó integrada en `update()`. `voicePlanWith` sigue porque la usan `voiceSet` y `voiceOff`.
  - **Pendientes de T7 y T9 cerrados:**
    - los `Setenv` de `HIVE_ACCESSIBLE` retirados, y borrada la prueba de contraste de `huh`;
    - H1: `assertFits` mide `View(viewCtx)`;
    - H4: Backspace durante la planificación de Voice;
    - H5: Voice limpia el mensaje al editar.
  - **Cobertura restituida (F1 del revisor, con su propuesta):** tres pruebas de `recoverPending` fijan cuándo el mensaje nombra `hive recover`. Cada una se comprobó rompiendo la regla que protege. También se corrigieron tres comentarios y se quitó `installTerminal.ProviderVersion`, que no tenía llamadas.

**Closes:** AC10.

**Depends on:** T9.

**Locations:**
- `tooling/cli/tui_prompter.go`, `tooling/cli/tui_screens.go` y `tooling/cli/tui_hosts.go`: se borran las partes secuenciales, y los archivos quedan vacíos o se eliminan;
- sus pruebas en `tui_test.go`, `tui_hosts_test.go` y `tui_screens_test.go`;
- `tooling/cli/install.go`, `update.go` y `voice.go`: el `prompter`, si queda con una sola implementación;
- `go.mod` y `go.sum`.

**Execution:** delegada, el mismo hijo. Antes de borrar, lista los archivos y funciones que va a eliminar.

**Test approach:** characterization (las pruebas de texto) y check (`rg` y `go mod`).

**Changes:** según [retiro de la versión secuencial](design.md#retiro-de-la-versión-secuencial) y [dependencias](design.md#dependencias). `go mod tidy`, después `go mod verify`.

**Verification:**
- `rg -n "charm.land/huh|HIVE_ACCESSIBLE|huhPrompter" go.mod tooling/` no devuelve nada.
- `rg -n "charm.land/bubbletea/v2 v2.0.2$" go.mod` devuelve la línea sin `// indirect`.
- `go mod verify`, `go vet ./...` y `go test -race -timeout 20m ./...` pasan.
- `go list -m all` comparado con el de `dd146c0` (sacado con `git worktree add` temporal en el scratchpad) solo pierde módulos, sin agregar ninguno.
- Los `sha256` de `install_test.go`, `update_test.go`, `voice_test.go` y `bootstrap_test.go` coinciden con los de `dd146c0`.

### T11 — Documentación y especificación

- [x] `deployment-manager.md` y la especificación del cambio describen la aplicación de pantalla completa.
  - **`review-task` (2026-09-28):** AC11 cumplido.
    - **Contra el código y el binario en `tmux`:** cada afirmación contrastada; `--help` nombra `hive tui` y no `HIVE_ACCESSIBLE`; la suite con `-race` de `tooling/cli` tardó 771 s, lo que confirma «about thirteen minutes».
    - **Correcciones aplicadas con la redacción del revisor:**
      - C1: una actualización sin cambios se aplica sin preguntar, como `hive update`; la excepción se agregó también a la especificación;
      - C2: «Uninstall all» nombrado, y que no acepta `y`;
      - C3: la línea de estado muestra el número de CLIs;
      - C4: con Off solo se ve la fila de la voz;
      - C5: la fuente solo se exige al agregar CLIs.

**Closes:** AC11.

**Depends on:** T10.

**Locations:**
- `_support/docs/architecture/deployment-manager.md`: el párrafo de dependencias (línea 3) y la sección de la interfaz (línea 66 en adelante);
- `_support/openspec/changes/gh-46-tui/specs/versioned-installation/spec.md`, ya reescrita en el plan: solo se ajusta si la implementación se desvió.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:**
- las vistas y las teclas;
- `hive tui` y sus opciones;
- el comportamiento sin terminal;
- que los comandos de texto son el camino para lectores de pantalla y guiones;
- la excepción de dependencias con Bubble Tea, `bubbles` y `lipgloss`;
- `-timeout 20m` para la suite.

**Verification:**
- `rg -n "HIVE_ACCESSIBLE|huh" _support/docs/architecture/deployment-manager.md` no devuelve nada.
- `rg -n "hive tui|Uninstall all|bubbletea" _support/docs/architecture/deployment-manager.md` devuelve la sección nueva.
- Cada afirmación coincide con el binario en `tmux`.

### T12 — Revisión de la interfaz, prueba en vivo, recorrido del usuario y revisión de código

- [x] `review-ux` y `sdd-verify` aprueban la aplicación en una terminal real, el usuario la recorre y la aprueba, y `/code-review` no deja hallazgos altos abiertos.
  - **Primera ronda (2026-09-28, candidato `076fd3d`):**
    - **`sdd-verify`:** AC1 y AC3 a AC11 cumplidos, comparando cada flujo con el comando de texto en homes gemelos, incluida una operación pendiente real provocada con `kill -9`. AC2 no cumplido, por D1 (alto): una segunda SIGINT o SIGTERM durante un `apply` mata el proceso y deja la terminal en la pantalla alterna. La causa probable es que Bubble Tea v2.0.2 deja de escuchar señales después de la primera.
    - **`review-ux`:** sin hallazgos bloqueantes ni altos. Tres medios introducidos:
      - M1: el mensaje de una instalación parcial se corta y puede perder `hive recover`;
      - M2: no se ven los cambios pendientes en CLIs;
      - M3: los campos largos pierden el comienzo sin señal.
    - **Límites de la revisión:** el tema claro, el campo de versión de capacidades y la fila de instalación antigua no se ejercitaron en vivo.
    - **Ronda de corrección:** D1, M1 a M3, y los menores L3, L5, L6 y L7.
  - **Re-verificación (única ronda):**
    - **`sdd-verify`:** AC2 cumplido. Con 2, 3 y 30 señales mezcladas durante un «Uninstall all» de unos 4 s, la aplicación sobrevive, termina sin operación pendiente y sale con 0 y la terminal restaurada; el resultado es igual al gemelo. No hay regresiones en AC4, AC6, AC7 ni en la recuperación.
    - **`review-ux`:** M1, M2 y M3 resueltos. La corrección de M3 introdujo N1 (medio): con el cursor al final de un valor largo aparece un «…» falso y el cursor no se ve. También dejó menores: N3, la ayuda dice «uninstall» en lugar de «uninstall all»; N4, el aviso de una instalación parcial se titula «Error».
    - **Corrección local de N1, N3 y N4 en el hilo principal:** sin nueva ronda de revisión, lo verifican sus pruebas y el recorrido del usuario.
    - **Para el recorrido del usuario (N2):** Enter en CLIs abre el resumen, y el resumen de una baja arranca en Apply, así que dos Enter aplican una baja.
  - **Recorrido del usuario (2026-09-29):** aprobado («me encanta») sobre los homes `full` y `detected` con el binario de `5ffd00c`. N2 queda como está.
  - **Rebase:** sobre `rebuild/harness-engineering` (`b2b1f63`), sin conflictos. Los 5 commits nuevos de la base son de guía y no tocan `tooling/`.
  - **`/code-review` (medio, verifica sus propios hallazgos), sobre `rebuild/harness-engineering...5ca5945`:** cuatro hallazgos, todos corregidos en la misma rama:
    - medio: con `--home` sin `--state-dir`, el aviso de recuperación decía `hive recover` sin la ruta, y ese comando actuaría sobre el estado real;
    - bajo: la ruta de `--state-dir` en el aviso iba sin comillas;
    - bajo: volver de una vista de solo lectura borraba las marcas pendientes en CLIs;
    - bajo: una voz activa que no está en la fuente se mostraba como Off.

**Depends on:** T11.

**Execution:** hilo principal. Prepara los datos con `make-fixtures.sh` (seis CLIs, 100 releases, una operación pendiente, nombres largos y un home vacío) y lanza los dos hijos en paralelo, cada uno con su propio home.

**Changes:**
- **`review-ux`:** [UI review criteria](../../../../content/skills/flow-build/references/ui-review-criteria.md) traducidos a terminal:
  - `tmux` a 80×24 y 120×40, con `capture-pane -e`;
  - fondo oscuro y `NO_COLOR` en `tmux`;
  - solo teclado;
  - todas las vistas con sus estados vacíos, de error, de cancelación y con listas largas;
  - que no quede nada impreso al salir.

  **Límite:** `tmux` probablemente no reenvía la consulta del color de fondo, así que el tema claro no se ve ahí (inferencia sin comprobar). Lo cubren la prueba de contraste en los dos fondos y, si el usuario usa un perfil claro, su recorrido.

  **Base de comparación:** la captura del recorrido y `dd146c0`, compilado en un `git worktree add` temporal en el scratchpad y borrado al terminar.

  **Capturas:** en `_support/workspace/2026-09-28-gh-46-tui/images/`, enlazadas desde el informe.
- **`sdd-verify`:** recorre AC1 a AC11 de punta a punta con `tmux send-keys`, incluida una operación pendiente real provocada con `kill -9` durante un `apply`.
- **Hallazgos:** se corrigen los de gravedad alta que introdujo el cambio y se vuelve a revisar.
- **Recorrido del usuario:** un bloque por vista con:
  - el comando para compilar;
  - cómo abrir `hive tui --home <home de prueba> --state-dir <estado de prueba> --source <worktree>`;
  - qué debe ver;
  - cómo borrar el home al terminar.
- **`/code-review`:** después de que el usuario apruebe, y antes del push.

**Verification:** los veredictos de los dos hijos, la aprobación del usuario y el resultado de `/code-review`, anotados aquí con su fecha.

## Verificación y revisión humana

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas | Al cerrar cada tarea y al final, en local | `go vet ./...`, `go test -race -timeout 20m ./...` y `go mod verify` en verde; no hay CI | Parte de la implementación |
| Verificación por tarea | Tras T7, T8, T9, T10 y T11 | `review-task` contra sus `AC<n>` | Parte de la implementación |
| Revisión de la interfaz y prueba en vivo | T12 | `review-ux` y `sdd-verify` en `tmux` sobre el binario | Parte de la implementación (interfaz visible) |
| Recorrido del usuario | T12, antes de `/code-review` | El usuario abre `hive tui` sobre un home de prueba | Parte del plan; decide el usuario |
| Revisión de código | Antes del push | `/code-review` (D4-A) | Ya decidida |
| Push | Tras la revisión de código | Rebase sobre la punta de `rebuild/harness-engineering` y push directo | Ya decidida (D4-A) |
| Usar la interfaz sobre la configuración real | Después de integrar | El usuario ejecuta `hive` | Decisión del usuario |

## Estado de la revisión y avance

**Revisión del plan (2026-09-28):** versión revisada `7a8fbe06b569`. Dos revisores `review-plan` en paralelo, como rol nativo de Claude Code y de solo lectura.

- **Arquitectura en Go y concurrencia:** cuatro hallazgos que bloqueaban, corregidos con la propuesta del revisor.
  - `bubbles/list` trae `sahilm/fuzzy`, un módulo nuevo: Releases usa una lista propia con `textinput` y `viewport`.
  - El alta se saltaba `VerifyIfPackaged` y `sourceHasCatalog`: se agregan antes de `RequiredHosts`, y `explicitStateDir` llega a `finalizeInstallResult`.
  - Nada probaba las teclas durante una escritura, y una SIGINT externa cortaba el `apply`: el conductor retiene el `Cmd` y `tea.WithFilter` descarta `InterruptMsg` y `QuitMsg`.
  - El conductor no terminaba con `spinner`, cursor, `Batch` o `Sequence`: tiene reglas y un tope.
  - También se incorporaron: una función de Update compartida con `updateWith`, la copia de `Options` en cada `Cmd`, `tea.WithWindowSize` en la prueba de pantalla alterna, y la vista de recuperación sin `prompter` de relleno.
- **Experiencia en la terminal:** cuatro hallazgos que bloqueaban, corregidos con la propuesta del revisor.
  - AC8 no tenía pruebas de progreso, de teclas ignoradas ni de desplazamiento.
  - AC2 durante una escritura quedaba sin prueba: ahora lo cierra T8.
  - AC9 estaba mal repartido: T7 cierra el menú, las vistas genéricas y el aviso, y se agregan 120×40 y listas largas.
  - AC4 no probaba las capacidades opcionales: se agrega un caso.
  - También se incorporaron:
    - Cancel de entrada en «Uninstall all»;
    - Backspace sin efecto en el menú;
    - marcar una instalación antigua significa migrarla;
    - el mensaje si falla el segundo paso;
    - el foco en las capacidades opcionales;
    - el aviso de éxito (AC8);
    - la carga con `spinner`;
    - el cambio de tamaño;
    - los símbolos con `NO_COLOR`;
    - el mensaje sin terminal completo (AC1);
    - la compilación de la base y el lugar de las capturas en T12.
- **Sin nueva ronda:** todas las correcciones aplican propuestas de los revisores.
- **Límites que quedan:**
  - el tema claro no se revisa en `tmux`;
  - nadie ejecutó código durante la revisión;
  - que Ctrl-C leído de un búfer llegue como tecla es una inferencia que confirma la prueba de T7.

**Avance:** T7 a T12 verificadas.
- **Suite final sobre `a22df95`:** `go vet`, `go mod verify` y `go test -race -timeout 25m ./...` pasan en los 17 paquetes (`tooling/cli` 781 s).
- **Entrega:** push directo a `rebuild/harness-engineering` (D4-A) el 2026-09-29, y cambio cerrado y archivado.

**Siguiente paso:** ninguno en este cambio.
