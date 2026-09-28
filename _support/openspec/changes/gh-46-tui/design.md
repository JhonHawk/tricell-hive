# Diseño

## Contexto verificado

Inspeccionado el 2026-09-28 sobre `dd146c0`. Dos subagentes `sdd-explore` leyeron los flujos de `tooling/cli` y el código fuente fijado de `huh` v2.0.3, Bubble Tea v2.0.2 y `bubbles` v2.0.0, en el module cache. Las rutas de código son relativas a `tooling/`.

**Qué se reutiliza de los flujos:**
- **Instalar:** las piezas ya son funciones separadas. El orden es:
  1. `DiscoverHosts`;
  2. `RequiredHosts` (`cli/install.go:438`, dentro de `expandToRequiredHosts`);
  3. `BuildPlan("install")` y `PlanUnchanged`;
  4. `adapter.Detect` y `adapter.Plan` (`previewInstallOnboarding`, `:575`);
  5. `showInstallSummary` (`:798`);
  6. `applyInstallOnboarding` (`:588`) y `finalizeInstallResult` (`:472`).

  Solo `runOnboardingWizard` (`:348`) las encadena con preguntas bloqueantes, y la aplicación no lo llama.
- **Quitar:** `removeFlow` (`cli/tui_hosts.go:80`) ya separa `BuildPlan("remove")`, `showRemoveSummary`, la confirmación y `Engine.Apply`.
- **Volver a una release:** `rollbackFlow` (`cli/tui_screens.go:201`) tiene la misma forma, con `ReleaseID`.
- **Voz:** `BuildVoicePlan` (`management/voice.go:391`), `showVoiceSummary` y `Apply`. Los valores permitidos son: `Address` en `sir`, `name` o `none`; `Intensity` en `subtle` o `marked`; `Name` solo con `name`. `ListVoices` y `CurrentVoice` dan las opciones y el valor activo.
- **Actualizar:** `updateWith` (`cli/update.go:74-165`) mezcla la resolución del commit, `planFromCommit`, el resumen y la confirmación. `planFromCommit` borra su directorio temporal antes de devolver el plan (`:192`), así que planificar y aplicar en `Cmd` distintos es seguro.
  - Se extrae una función que hace los pasos 1 a 3 (`:77-102`) y devuelve el plan, la línea «Source commit» (`:124`) y si no hay cambios.
  - `updateWith` y la vista la usan las dos, y la salida de `hive update` no cambia.
  - Sin cambios, la vista aplica directamente, como el comando (`:138-145`).
- **Los resúmenes** escriben en un `io.Writer`: la aplicación los renderiza en un `bytes.Buffer` y los muestra en un `viewport`.

**Planes que mezclan altas y bajas:**
- `BuildPlan` acepta una sola acción, `install` o `remove` (`management/plan.go:317-321`).
- Instalar nunca quita otros CLIs (`ownership.go:179-208`).
- Con una operación pendiente, `BuildPlan` falla (`plan.go:333`).
- Por eso altas y bajas son dos planes aplicados en secuencia, y el segundo se construye después de aplicar el primero.

**Recursos compartidos y voz al quitar:**
- `RequiredHosts` (`management/versioning.go:161`) solo expande al instalar.
- Quitar un CLI conserva los recursos compartidos que otro sigue usando (`nextRecord`, `plan.go:437`); `showRemoveSummary` los lista.
- La voz se retira cuando no queda ningún bloque Hive de usuario (`apply.go:279`).
- «Uninstall all» equivale a `BuildPlan("remove")` con `RegisteredHosts` (`types.go:342`).

**Límites de `huh` v2.0.3 (verificados en su código):**
- `Form.Update` evalúa la tecla de salir antes de reenviarla al campo (`form.go:563-568`). Si Backspace volviera atrás, un `Input` ya no podría borrar texto.
- Un `Select` en línea solo navega entre campos con Tab o Enter, y Enter en el último campo cierra el formulario (`field_select.go:241-247`, `group.go:222-226`).
- El modo accesible solo funciona con `Form.Run`, no dentro de otro programa (`form.go:676-681`).

**Qué da Bubble Tea v2.0.2:**
- La pantalla alterna es el campo `tea.View.AltScreen` (`tea.go:149-161`). Se sale de ella al terminar el programa (`cursed_renderer.go:314-322`).
- Cada `tea.Cmd` corre en su propia goroutine, y su resultado vuelve como mensaje (`tea.go:679-720`).
- `tea.WithInput`, `tea.WithOutput` (`options.go:30,40`) y `Program.Send` permiten probar el programa sin terminal.
- `bubbles` v2.0.0 da `textinput`, `viewport`, `spinner`, `key` y `help`, que no traen módulos nuevos.
- `bubbles/list` importa `github.com/sahilm/fuzzy` (`list/list.go:17`), que no está en `go.sum`, así que no se usa.
- `teatest/v2` solo existe como pseudo-versión, así que no se agrega.
- `tea.WithFilter` (`options.go:133`) filtra mensajes antes del modelo. Una SIGINT externa llega como `InterruptMsg`, y `Run` termina con `ErrInterrupted` (`tea.go:627-652,749`). En modo crudo, Ctrl-C llega como una tecla.
- `tea.WithWindowSize` (`options.go:163`) fija el tamaño sin terminal. Sin él, el tamaño es 0×0.

**Chequeos que la instalación hace antes de planificar:** `distribution.VerifyIfPackaged` solo se llama dentro de `runInstallFlowWith` (`cli/install.go:218`), y `sourceHasCatalog` dentro de `installScreen` (`cli/tui_hosts.go:27`). `BuildPlan` ya normaliza las opciones y rechaza una operación pendiente (`management/plan.go:329-338`).

**Pruebas actuales:**
- Unas 8 comparan la interfaz con el comando equivalente en homes gemelos (`installViaText`, `collectFiles`, `assertHomesMatch`, `assertStateJSONMatches` en `cli/tui_hosts_test.go:113-214`). Solo dependen del estado final.
- Unas 40 dependen de los formularios secuenciales: guiones numéricos, textos exactos de `huh`, el lector de un byte y la recuperación de su pánico.
- `TestThemeContrastMeetsWCAGAA` calcula el contraste con funciones genéricas, pero sobre estilos de `huh`.

## Diseño elegido

### Aplicación (D1-A)

- **Programa único:** `tea.NewProgram` con un modelo raíz que guarda una pila de vistas. La vista de arriba recibe las teclas y dibuja; el raíz compone la línea de estado, la vista y la barra de ayuda en un `tea.View` con `AltScreen: true`.
- **Tamaño mínimo:** con menos de 80×24, en lugar de la vista se muestra «Terminal too small: needs 80×24».
- **Operaciones largas:** planificar, aplicar, resolver un commit y recuperar corren como `tea.Cmd` y devuelven un mensaje con el resultado.
  - El modelo solo cambia en `Update`, nunca desde la goroutine.
  - Mientras una operación de escritura corre, la vista muestra un `spinner` e ignora las teclas, Ctrl-C incluido. Un `tea.WithFilter` descarta `InterruptMsg` y `QuitMsg` mientras el modelo raíz marca una escritura en curso, así que una SIGINT externa tampoco corta un `apply` a medias. Una SIGKILL sigue dejando la operación pendiente, que se recupera al abrir.
  - Descubrir CLIs, calcular el estado y armar la línea de estado también corren como `tea.Cmd`, con un `spinner` mientras cargan.
  - Cada `tea.Cmd` recibe una copia de las `Options` (con `Hosts` copiado), nunca el valor que el modelo sigue modificando.
- **Cambio de tamaño:** bajar de 80×24 muestra el aviso, y volver restaura la vista con su estado.
- **Campos de texto largos:** un `textinput` (Source, Name, el filtro) desplaza su contenido alrededor del cursor, sin «…». El valor completo sigue accesible moviendo el cursor. Para AC9 no cuenta como una fila cortada (decidido al construir T9).
- **Teclas globales:**

| Tecla | Efecto |
| --- | --- |
| Esc | Vuelve a la vista anterior (en el menú, sale). En Releases con el filtro activo, primero limpia el filtro. |
| Backspace | Igual que Esc cuando ningún campo de texto tiene el foco; dentro de uno, borra. En el menú no hace nada. |
| Ctrl-C | Sale de la aplicación con 0, salvo mientras se aplica un cambio. |
| ↑↓ · ←→ · Enter · Espacio | Mover, cambiar un valor, aceptar, marcar una casilla. |

- **Entrada:**
  - `hive` sin argumentos es `hive tui` sin opciones, y las opciones de `hive tui` no cambian.
  - Sin terminal en la entrada y la salida, `hive` imprime el error de uso de hoy, y `hive tui` falla con «hive tui needs a terminal; use the text commands: hive status, install, update, releases, voice, plan/apply to remove hosts, recover».
  - Se retiran `HIVE_ACCESSIBLE`, el lector de un byte y la recuperación del pánico de `huh`.
- **Tema:**
  - Estilos propios de `lipgloss` con pares claro/oscuro.
  - El fondo se detecta con `lipgloss.HasDarkBackground` antes de arrancar el programa, como hoy.
  - Con `NO_COLOR`, estilos sin color. La casilla (`[x]`, `[ ]`), el cursor (`>`) y el botón elegido (`[Apply]`) siempre usan símbolos, así que nada depende solo del color.
  - La prueba de contraste pasa a medir estos estilos, con las mismas funciones WCAG.
- **Ubicación:**
  - Todo vive en `tooling/cli`, en archivos `tui*.go`, porque usa funciones no exportadas.
  - `tooling/management` y `tooling/distribution` siguen sin dependencias externas.
- **Textos:** en inglés, como el resto de la salida del gestor.

### Vistas

Cada vista es un modelo con `Update` y `View`. Resúmenes, confirmaciones, avisos y resultados son vistas genéricas que se apilan encima de la vista de origen. Al terminar, el resultado se muestra como un aviso dentro de la vista de origen, que se refresca desde el estado.

| Vista | Contenido e interacción | Vacío, límite y error |
| --- | --- | --- |
| Menú | CLIs, Update, Releases, Voice, Quit. La línea de estado encima, por ejemplo «2 CLI hosts · release 622087a518ff · voice jarvis». | Sin CLIs: «No CLI hosts are registered». |
| CLIs | Una fila por CLI detectado o registrado: casilla, nombre, etiqueta (detected, registered, legacy install), release corta, versión y `drift`. Marcar un CLI con instalación antigua significa migrarla e instalar, como en `hive install` (`cli/install.go:822`). Espacio marca o desmarca, `a` aplica y `u` abre «Uninstall all». La barra de ayuda lo dice. Al volver a la vista, las casillas se recargan desde el estado y se consulta si quedó una operación pendiente. | Mientras carga, un `spinner`. Ningún CLI detectado ni registrado: el mensaje del asistente. Sin cambios marcados, `a` dice «Nothing to apply». |
| Aplicar CLIs | Se describe en el [diff de CLIs](#diff-de-clis). | Un error del plan o de la aplicación vuelve a CLIs con el mensaje del comando. |
| Update | Dos `textinput`, Source (`.`) y Revision (`HEAD`). ↑↓ o Tab cambian de campo y Enter resuelve el commit con un `spinner` antes del resumen. | Sin CLIs: el error de `update`. Revisión inválida o sin Git: el mensaje del comando en la vista. |
| Releases | Lista propia: filas en un `viewport` que sigue al cursor y, encima, un `textinput` de filtro que se abre con `/` y filtra por subcadena sin distinguir mayúsculas. Filas de 78 columnas o menos: ID corto, fecha, commit y CLIs recortados con «…». Marca la instalada. Enter lleva al resumen de `rollbackFlow`. | Sin releases: «No releases are retained yet». La instalada: «Already installed». Un error de validación: su mensaje. |
| Voice | Filas Voice (Off y `ListVoices`), Address (`none`, `sir`, `name`), Name (`textinput`, solo con `name`) e Intensity (`subtle`, `marked`), cargadas desde `CurrentVoice`. Con Off solo se muestra la fila Voice, porque los demás valores no significan nada (decidido al construir T9). ↑↓ cambia de fila, ←→ cambia el valor y Enter lleva al resumen. | Sin CLIs: el mensaje de Status. Fuente sin voces: el error de la fuente. Sin cambios: «Voice is already set this way». |
| Recuperar | Al abrir, si hay una operación u onboarding pendiente: qué se interrumpió y las opciones «Recover» y «Leave it». Usa la lógica de `handlePendingInstallOperation`, sin su pregunta de texto. | Si la recuperación falla, el mensaje del comando, que nombra `hive recover`. |
| Resumen y confirmación | Un `viewport` con el resumen del comando (↑↓, PgUp y PgDn desplazan), más Apply y Cancel. ←→ elige y Enter acepta; también `y` y `n`. Arranca en Apply, salvo en «Uninstall all», que arranca en Cancel y no acepta `y`. | Esc y Cancel vuelven sin cambios, con «Cancelled. No changes applied.» dentro de la vista de origen. Al terminar con éxito, la vista de origen se refresca y muestra el resultado del comando. |
| Aviso | Un texto y Continue, para la expansión de CLIs requeridos y el paso intermedio del diff. | — |
| Capacidades opcionales | Casillas de `adapter.Detect`, ninguna marcada. ↑↓ mueve y Espacio marca. Una capacidad marcada que pide versión (no `ManualOnly`) abre debajo un `textinput` con el foco; Enter o Tab lo cierran y vuelven a la lista. Con la lista enfocada, Enter sigue al resumen. | Sin ofertas, se salta. |

### Diff de CLIs

Al aplicar (`a`), marcadas = conjunto deseado, y registradas = conjunto actual.

1. **Bajas** (registradas y desmarcadas):
   1. `BuildPlan("remove")` con esos CLIs.
   2. El resumen de `showRemoveSummary`.
   3. La confirmación.
   4. `Engine.Apply`.
   Si el `apply` falla y deja una operación pendiente, el paso de altas se omite y se abre la vista de recuperación.
2. **Altas** (marcadas y sin registrar), construidas sobre el estado posterior a las bajas:
   1. `distribution.VerifyIfPackaged` y `sourceHasCatalog`, los mismos chequeos que `runInstallFlowWith` e `installScreen`. Si fallan, su mensaje vuelve a CLIs.
   2. `RequiredHosts`. Si agrega CLIs, el aviso con el texto de `expandToRequiredHosts`, más Accept y Cancel.
   3. `BuildPlan("install")`.
   4. Las capacidades opcionales.
   5. `adapter.Plan`.
   6. El resumen de `showInstallSummary`.
   7. La confirmación.
   8. `applyInstallOnboarding` y `finalizeInstallResult`, con `fromInterface` en verdadero y el `explicitStateDir` real.
3. **Con bajas y altas a la vez (D3-A):**
   - Cada resumen dice «Step 1 of 2: remove» o «Step 2 of 2: install».
   - Rechazar el segundo paso vuelve a CLIs con «Removed: <CLIs>. Install cancelled; no further changes».
   - Si el segundo paso falla, vuelve a CLIs con «Removed: <CLIs>.» seguido del mensaje del comando.

«Uninstall all» (`u`) salta las casillas: `BuildPlan("remove")` con todos los registrados, el resumen, la confirmación y `Apply`.

### Retiro de la versión secuencial

Se eliminan, sin coexistir:
- `huhPrompter` y sus métodos;
- `runMenu` y `runMenuEntry`;
- las pantallas `installScreen`, `removeScreen`, `statusScreen`, `updateScreen`, `releasesScreen` y `voiceScreen`;
- el lector de un byte;
- el tema de `huh` y las pruebas que dependen de todo eso.

Se conservan y se reutilizan:
- `removeFlow`, `rollbackFlow` y los resúmenes, partidos en «armar el plan» y «aplicar» donde haga falta;
- `interfaceStatusLine`, `statusLine` y las etiquetas de releases y voces;
- los auxiliares de homes gemelos.

La vista de recuperación llama a las funciones de detección y recuperación que usa `handlePendingInstallOperation`, no a esa función, para no necesitar un `prompter` de relleno. Si al terminar el `prompter` queda con una sola implementación (`installTerminal`), la interfaz se elimina y sus usuarios vuelven a recibir el tipo concreto. Las variantes `updateWith` y `voicePlanWith` se eliminan si nada más las usa. Las pruebas de texto sin modificar son la caracterización.

### Dependencias

- **Salen:** `charm.land/huh/v2` y sus dependencias exclusivas, con `go mod tidy`.
- **Quedan directas:** `charm.land/bubbletea/v2` v2.0.2 (hoy indirecta), `charm.land/bubbles/v2` v2.0.0 y `charm.land/lipgloss/v2` v2.0.1.
- **También directa (decidido al construir T7):** `github.com/charmbracelet/x/ansi`, que ya estaba en `go.mod` como indirecta. Se usa por `ansi.Truncate` y `ansi.Wrap`, porque `lipgloss` v2 no recorta con «…» ni ajusta respetando las secuencias ANSI. No agrega ningún módulo.
- **Sin módulos nuevos:** los tres ya están en `go.sum`. Solo se usan los paquetes de `bubbles` que no traen módulos nuevos (`textinput`, `viewport`, `spinner`, `key` y `help`), no `list`. T10 compara `go list -m all` con el de `dd146c0`.
- **Excepción:** la excepción a la regla de dependencias de `AGENTS.md` se mantiene, con la lista actualizada, porque la interfaz es una función de producto pedida por el usuario.

## Pruebas

- **Conductor propio en `tui_test.go`:** crea el modelo raíz sobre un home de prueba, envía `tea.KeyPressMsg` y ejecuta los `tea.Cmd` devueltos en el mismo goroutine de la prueba. Después afirma sobre el texto de `View()` y sobre el estado final. No se agrega `teatest`. Reglas del conductor:
  - expande `tea.BatchMsg`;
  - la aplicación no usa `tea.Sequence`, cuyo mensaje no es exportado;
  - descarta los mensajes periódicos del `spinner` y del cursor de `textinput`;
  - un tope de iteraciones hace fallar la prueba en lugar de colgarla;
  - puede retener el `Cmd` de una escritura. Así una prueba afirma el `spinner` y que Esc, Backspace y Ctrl-C no hacen nada, y después lo libera.
- **Un programa real:** una prueba corre `tea.NewProgram` con `tea.WithInput`, `tea.WithOutput` y `tea.WithWindowSize(80, 24)`, aplica una baja sobre un home de prueba y sale. Con `-race` cubre la goroutine real del `apply`. La misma familia comprueba que la salida abre (`\x1b[?1049h`) y cierra (`\x1b[?1049l`) la pantalla alterna, y que el filtro descarta una `InterruptMsg` durante una escritura.
- **Homes gemelos:** las comparaciones de archivos y `state.json` con el comando equivalente se conservan con el conductor nuevo:
  - instalar dos CLIs y la expansión de Grok a Claude;
  - quitar con una voz activa;
  - altas y bajas juntas, comparadas contra `plan remove`+`apply` seguido de `install`;
  - «Uninstall all»;
  - Update, la vuelta a una release y la voz.
- **Pantalla alterna:** una prueba arranca el programa con `tea.WithInput` y `tea.WithOutput` sobre búferes, sale con Ctrl-C y comprueba que la salida abre (`\x1b[?1049h`) y cierra (`\x1b[?1049l`) la pantalla alterna.
- **Ancho y alto:** cada vista se renderiza a 80×24 y a 120×40 con datos largos. Ninguna línea supera el ancho (`lipgloss.Width`), la vista no supera el alto, y las listas largas (6 CLIs, 100 releases) muestran el cursor al bajar hasta el final.

## Límites de seguridad

- **Ningún dato externo nuevo:** fuente, revisión, nombre y versiones de capacidades pasan por las mismas validaciones de los comandos.
- **Dependencias:** solo se quitan módulos. Las que quedan ya estaban fijadas y verificadas con `go mod verify`, según [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md).

## Compatibilidad

- **Comandos:** no cambian su salida ni sus opciones.
- **Formatos:** no cambian los formatos de plan ni de estado.
- **Retirado sin transición:** `HIVE_ACCESSIBLE` y la interfaz secuencial nunca se publicaron (la rama es local), así que no tienen consumidores.
