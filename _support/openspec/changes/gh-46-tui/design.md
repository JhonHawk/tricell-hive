# Diseño

## Contexto verificado

Inspeccionado el 2026-09-28 sobre `32eb03f`. Dos subagentes `sdd-explore` investigaron el asistente actual y la librería. Dos revisores `review-plan` comprobaron el diseño contra el código y contra `huh` v2.0.3, con un experimento desechable.

- **Asistente de `hive install`** (`tooling/cli/install.go`): la entrada común es `runInstallFlow` (159-181). Hace `distribution.VerifyIfPackaged`, `NormalizeOptions` y la comprobación de operaciones pendientes (`handlePendingInstallOperation`, 198-241), y después llama a `runOnboardingWizard` (245-326).
  - Lo comparten el instalador offline y la instalación en línea (`bootstrap.go:220` y `:234`, que agregan `BindRetainedInstaller`).
  - **Preguntas:** `selectInstallerHosts` (499-552), la confirmación de `expandToRequiredHosts` (333-359), `previewInstallOnboarding` (444-455), `selectProviderRequests` (560-629), `readProviderVersion` y `confirmInstall` (907-932).
  - **La regla «sin terminal»** vive dentro de esas funciones: `selectInstallerHosts` falla, `selectProviderRequests` devuelve `(nil, true)` y `confirmInstall` falla.
- **`update`** (`update.go:131`) y **`voice`** (`voice.go:174`) construyen su propio `installTerminal` y reciben `args` e `interactive`.
- **Pruebas:** `install_test.go` tiene 17 pruebas que recorren el asistente como caja negra y comparan el texto exacto. `TestRunUsageErrors` (`main_test.go:278`) llama a `run([]string{})`.
- **`hive` sin argumentos** imprime un error de uso (`main.go:39-41`) y no acepta opciones. Cambiar `HOME` no aísla, porque se siguen leyendo `CODEX_HOME` y `CLAUDE_CONFIG_DIR` (`types.go:244-247`) y se busca en `PATH` (`install.go:489`).
- **`huh` v2.0.3** (`charm.land/huh/v2`, MIT, pide Go 1.25.8), comprobado por el revisor en su código fuente:
  - En modo accesible, `PromptString` crea un `bufio.Scanner` nuevo en cada pregunta, y ese scanner lee por adelantado. Con `strings.NewReader("2\ny\n")`, la segunda pregunta recibe EOF; con `iotest.OneByteReader` funciona.
  - `runAccessible` ignora los errores de cada campo y devuelve `nil`. EOF se toma como «usar el valor por defecto», y un `MultiSelect` con validador vuelve a preguntar sin fin.
  - Fuera del modo accesible, Ctrl-C devuelve `ErrUserAborted`. Los temas de v2 reciben `isDark`, y `Note` no se desplaza.

## Diseño elegido

### Punto de entrada (D14-A)

- **Subcomando nuevo:** `hive tui [--home DIR] [--state-dir DIR] [--source DIR]` abre la interfaz. `hive` sin argumentos es exactamente `hive tui` sin opciones.
  - Así se puede abrir sobre un home de prueba sin tocar la configuración real.
  - `--source` vale `.` por defecto, como en los demás comandos.
  - Las pantallas que necesitan una fuente (Install, Voice) muestran un error y vuelven al menú si esa fuente no tiene catálogo. El error dice: «Run hive from a Hive checkout or package, or pass --source».
- **Detección** inyectable: `openInterface(isTTY, accessible bool, in io.Reader, out io.Writer, o options)`.
  - Con `HIVE_ACCESSIBLE=1`, abre en modo accesible.
  - Si no, con entrada y salida en terminal (`terminalInput(os.Stdin) && terminalInput(os.Stdout)`), abre la interfaz normal.
  - Si no, `hive` sin argumentos imprime el error de uso de hoy, sin cambios, y `hive tui` falla con un mensaje que nombra `HIVE_ACCESSIBLE`.
  - `TestRunUsageErrors` fija el entorno con `t.Setenv("HIVE_ACCESSIBLE", "")` y no depende de la terminal real.
- **Sin cambios:** los demás subcomandos, `install.sh` y `bootstrap.sh`.
- **Ayuda:** `printHelp` menciona `hive tui` y `HIVE_ACCESSIBLE`.

### Separar las preguntas de la lógica

- **Interfaz `prompter`** en `tooling/cli`:
  - `SelectHosts(candidates []hostCandidate) (hosts []string, ok bool, err error)`
  - `SelectProviders(offers []providerOffer) ([]providerRequest, bool, error)`
  - `ProviderVersion(offer providerOffer) (string, bool, error)`
  - `Confirm(prompt string, allowBack bool) (installDecision, error)`
- **`installTerminal` la implementa** con las funciones actuales, incluida la regla «sin terminal», sin cambiar ningún texto.
- **La usan** `runOnboardingWizard`, `expandToRequiredHosts`, `previewInstallOnboarding`, `handlePendingInstallOperation` y variantes internas `updateWith(o, flags, p prompter, out)` y `voicePlanWith(...)`, que `update` y `voice` llaman con su `installTerminal`.
- **`runInstallFlow`** recibe también el `prompter`, así que `bootstrap.go` le pasa el suyo, sin cambiar `BindRetainedInstaller`.
- **La interfaz implementa otra versión con `huh`** y entra por `runInstallFlow`, así que hereda la verificación del paquete y la comprobación de operaciones pendientes.

### El `prompter` de `huh`

- **`Confirm` con `allowBack`:** un `Select` de tres opciones (Apply, Back, Cancel). Sin `allowBack`, un `Confirm`.
- **`SelectHosts`:** un `MultiSelect` sin validador. Nada viene marcado de entrada, igual que en el asistente de texto, y cada opción conserva la etiqueta de estado del asistente (detectado, registrado o instalación antigua).
- **Cancelar:** Ctrl-C o Esc (`ErrUserAborted`) y el fin de la entrada se convierten en cancelación (`ok=false` o `installCancelled`), nunca en un error.
- **Entrada en modo accesible:** la interfaz envuelve la entrada en un lector que entrega un byte por lectura y registra el EOF. Después de cada formulario, el `prompter` lo consulta y lo trata como cancelación. Ningún otro `bufio.Reader` lee esa entrada durante la sesión.
- **Tema:** `huh.ThemeCharm(isDark)`, detectando el fondo con la función de v2 documentada para la versión fijada; se comprueba en T2. Con `NO_COLOR` definido se usa `ThemeBase`.

### La interfaz

- **Ubicación:** vive en `tooling/cli` (paquete `main`), en archivos `tui*.go`, porque reutiliza funciones no exportadas. `tooling/management` y `tooling/distribution` no importan la dependencia. Una prueba recorre sus importaciones con `go/parser`, siguiendo solo `tricell-hive/...`, y falla ante cualquier ruta cuyo primer elemento lleve un punto.
- **Forma:** un menú en bucle, hecho con formularios de `huh` uno tras otro, no una aplicación a pantalla completa.
  - Cada pantalla empieza con un encabezado de una línea.
  - Imprime su resumen con las mismas funciones de los comandos, a la salida normal para que quede en el historial, y pide la confirmación en un formulario aparte. No usa `Note` para los resúmenes.
  - Antes de una operación que puede tardar imprime una línea, por ejemplo «Resolving HEAD…».
- **Menú:** siete entradas fijas (Status, Install CLIs, Remove CLIs, Update, Releases, Voice, Quit) en ese orden y siempre visibles, así que la numeración del modo accesible nunca cambia. Encima se imprime una línea de estado, por ejemplo «2 CLI hosts · release 622087a518ff · voice jarvis», o «No CLI hosts are registered». Quit es la opción por defecto, así que el fin de la entrada sale con 0.
- **Cancelar y errores dentro de una pantalla:**
  - Ctrl-C, Esc, rechazar la confirmación o el fin de la entrada imprimen «Cancelled. No changes applied.» y vuelven al menú.
  - Un error del flujo (por ejemplo, Git ausente, una revisión inválida, un conflicto o una release que el gestor actual no puede validar) imprime el mensaje del comando y vuelve al menú.
  - Ctrl-C en el menú sale con 0.
  - **Excepción en Releases, decidida al construir T4:** `huh` v2.0.3 comprueba la tecla de salir antes de que el campo vea la pulsación. Por eso, en la lista de releases, Esc cierra o limpia el filtro, y solo Ctrl-C cancela la pantalla. Esa lista tiene su propio mapa de teclas.
- **Al abrir:** si hay una operación o un onboarding pendiente, se usa `handlePendingInstallOperation` con el `prompter` de `huh`. Sus mensajes nombran `hive recover` en lugar de «run ./install.sh again».
- **Pantallas y sus estados:**

| Pantalla | Qué hace | Vacío o límite |
| --- | --- | --- |
| Status | Una fila por CLI registrado en alcance de usuario, con la release (ID corto), la versión, cuántos recursos están en `drift` y la voz. Solo lee. | «No CLI hosts are registered. Choose Install CLIs.» |
| Install CLIs | `runInstallFlow` con el `prompter` de `huh`. | Si no se detecta ningún CLI, el mensaje del asistente. |
| Remove CLIs | `MultiSelect` de los CLIs registrados; `removeFlow` arma `BuildPlan("remove")` e imprime un resumen propio: archivos que se quitan, recursos compartidos que se conservan y bloques de voz que se quitan. Confirmar y `Engine.Apply`. | Sin CLIs: el mismo mensaje que Status. |
| Update | `Input` de fuente (`.`) y revisión (`HEAD`), y `updateWith`. | Sin CLIs: el error de `update`. |
| Releases | Un solo `Select` de altura fija, con filtro y etiquetas de 78 columnas o menos (ID corto, fecha, commit y CLIs recortados), que marca la release instalada. Elegir una ofrece volver a ella: `rollbackFlow` arma `BuildPlan("install")` con `ReleaseID` para los CLIs registrados, muestra `showInstallSummary` (incluido su aviso de voz), confirma y aplica. | Sin releases: «No releases are retained yet». Elegir la instalada: «Already installed». |
| Voice | Muestra la voz activa y ofrece las voces de `ListVoices` y Off. Con una voz, pide el tratamiento, el nombre (solo con `name`) y la intensidad, y sigue con `voicePlanWith`. | Sin CLIs: el mismo mensaje que Status. Fuente sin voces: el error de la fuente. |

- **Textos:** en inglés, como el resto de la salida del gestor.

### Dependencia

- **Qué entra:** `charm.land/huh/v2` v2.0.3 y sus dependencias en `go.mod` y `go.sum`.
  - Son módulos Go sin scripts de instalación, verificados con `go mod verify` y los checksums del proxy.
  - Es la primera dependencia externa del gestor, y se registra como excepción en la documentación: `AGENTS.md` prohíbe dependencias para configuración, y esta es para la interfaz de producto que pidió el usuario.
- **Construir una release** pasa a necesitar descargar esos módulos.
- **Instalar** el paquete offline no cambia.

## Límites de seguridad

- **Ningún dato externo nuevo:** los valores escritos (fuente, revisión, nombre) pasan por las mismas validaciones de los comandos.
- **Dependencia:** fuente, versión fijada y ausencia de scripts, según [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md).

## Compatibilidad

- **Comandos:** no cambian su salida ni sus opciones.
- **Formatos:** no cambian los formatos de plan ni de estado.
- **Solo cambia** `hive` sin argumentos en una terminal o con `HIVE_ACCESSIBLE=1`, y aparece el subcomando `hive tui`.
