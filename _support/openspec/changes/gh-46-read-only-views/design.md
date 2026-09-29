# Diseño

## Contexto verificado

Inspeccionado el 2026-09-29 sobre `82f2dcc`. Dos subagentes `sdd-explore` leyeron el código y la documentación oficial de los hosts. Las afirmaciones consecuentes se contrastaron con las fuentes. Las rutas son relativas a la raíz del repositorio.

**Lo que ya existe y se reutiliza:**
- **Aplicación:** la interfaz `view` y la estructura `action` (`tooling/cli/tui_app.go:62-111`), el menú fijo `mainMenuItems` (`tooling/cli/tui_views.go:167`) y el `scrollBox` con `viewport` (`tui_views.go:241`).
- **Detección de CLIs:** `detectInstallerHosts` (`tooling/cli/install.go:560`) usa `exec.LookPath(host)` sobre `installerHosts` (`install.go:21`). Con `--home` sintético no detecta nada.
- **Estado por recurso:** `management.Status` (`tooling/management/plan.go:741`) devuelve `StatusEntry` (`types.go:194`). Sus estados son `installed`, `drift`, `retained_shared`, `shadowed`, `unowned`, `unowned_or_conflicting`, `migration_required` y `recovery_required`. Los marcadores rotos o duplicados salen como `unowned_or_conflicting` en un bloque que Hive no registró, y como `drift` en uno registrado (`plan.go:790-801`). Con una operación pendiente, `Status` marca todas las filas como `recovery_required` (`plan.go:806-808`).
- **Releases:** `management.Releases` (`plan.go:883`) da `LastWrittenAt` (la fecha de modificación de `releases/<id>.json`) y los consumidores por release.
- **Modelos:** `agents.Render` (`integrations/agents/agents.go:238`) resuelve `perfil → {model, effort}` y deja que el `effort` del rol gane en Claude, Codex y Pi (`acceptsEffort`, `:226`). El snapshot de cada release guarda `Files` (`Payload{Path, Data}`) y `Profiles` (`types.go:82-92`).
- **Integraciones:**
  - `providers.Catalog` (`tooling/providers/catalog.go:7`) lista Engram, Context7 y pi-subagents, todas `manual`.
  - El registro se escribe en `onboarding/<id>.json` (`management/onboarding.go:154`) y no tiene lector exportado.
  - `context7Candidates` (`tooling/cli/setup.go:77`) da las rutas de skills que ya revisa `hive setup`.
- **`## Hive`:** no hay código que la lea. Su única definición es `content/guidance/global.md`, sección «Project settings».

**Evidencia sobre los hosts** (documentación consultada el 2026-09-29, con versiones instaladas claude 2.1.284, codex 0.158.0, grok 1.0.45, pi 0.87.1, opencode 2.0.19 y cursor-agent 2026.09.23):
- **Claude Code:** guarda un archivo por sesión en `~/.claude/sessions/<pid>.json`. Se observaron las claves `pid`, `sessionId`, `cwd`, `startedAt` (número, milisegundos Unix), `version`, `kind` y `status`.
  - Por documentación, lee sus instrucciones al arrancar y `/compact` vuelve a leer el `CLAUDE.md` del proyecto ([sub-agents](https://code.claude.com/docs/en/sub-agents), [memory](https://code.claude.com/docs/en/memory)).
  - Junto a cada archivo hay otro `<pid>.<hash>.key`, que no se lee.
- **Grok:** guarda las sesiones en `~/.grok/active_sessions.json`, una lista de `{session_id, pid, cwd, opened_at}` donde `opened_at` es una cadena de fecha. Inyecta `AGENTS.md` al inicio de la conversación.
- **Codex:** lee `AGENTS.md` una vez por sesión y no tiene un registro de sesiones vivas fiable ([AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md)).
- **OpenCode v2:** recarga `AGENTS.md` en el siguiente mensaje ([instructions](https://opencode.ai/v2/docs/instructions)).
- **Pi y Cursor:** no hay registro ni documentación de recarga que sirvan.
- **Límite:** los archivos de sesión de Claude Code y Grok son internos y no documentados. Pueden cambiar con cualquier versión.

## Diseño elegido

### Dónde vive cada parte

| Parte | Paquete | Motivo |
| --- | --- | --- |
| Modelo efectivo por rol | `integrations/agents` (`Resolve`) y `tooling/management` (`EffectiveModels`) | Datos del snapshot de la release; biblioteca estándar solamente |
| Lector del último registro de integraciones | `tooling/management` (`LastOnboarding`) | El formato del registro es de `management` |
| Versiones, sesiones, detección de integraciones y validación de `## Hive` | `tooling/cli`, archivos `doctor*.go` | Ejecutan procesos y leen archivos de terceros: no pertenecen a `management` |
| Vistas | `tooling/cli`, archivos `tui_*_view.go` | Como las vistas existentes |

### Contratos nuevos

- **`agents.Resolve(source string, data, profiles []byte, host string) (profile string, m Model, err error)`:** devuelve el perfil del rol y el `Model` efectivo, con la misma regla de esfuerzo que `Render`. `Render` pasa a llamarla, así que hay una sola regla.
  - Las pruebas actuales de `Render` solo buscan subcadenas (`agents_test.go:211-335`), así que no bastan como protección.
  - Antes de extraer la función se guarda como golden la salida completa de `Render` para los 20 roles en los 6 CLIs, y la prueba la compara byte a byte después.
- **`management.EffectiveModels(o Options) ([]ModelRow, error)`:**
  - `ModelRow` tiene `Host`, `Role`, `Profile`, `Model` y `Effort`.
  - Recorre los registros de agentes del estado en alcance `user`. Para cada registro lee el snapshot de su propio `Record.Release`, porque un recurso compartido puede conservar una release anterior (`plan.go:493`). Toma el `Payload` del rol (ruta que cumple `agents.IsSource`) y llama a `Resolve` con los `Profiles` de ese snapshot.
  - Ordena por CLI y luego por rol.
  - Un CLI sin agentes instalados no da filas.
  - Sin CLIs registrados, devuelve una lista vacía y ningún error.
  - **Presentación** (en `cli`, no en el contrato):
    - un `Model` vacío se muestra como «host default»;
    - un `inherit` literal, como «inherit (parent session)»;
    - en OpenCode, el sufijo `#variant` se separa del modelo y se muestra como esfuerzo.
- **`management.LastOnboarding(stateDir string) (OnboardingResult, time.Time, bool, error)`:**
  - Lee el registro terminado más reciente de `onboarding/`, por fecha de modificación.
  - `loadOnboarding` tiene fija la ruta del registro pendiente (`onboarding.go:129-133`). Por eso se extrae su validación a una función que recibe el registro ya decodificado y el directorio de estado canónico, y la usan las dos.
  - `LastOnboarding` resuelve `stateDir` con `target.Canonical`, igual que `normalize` (`types.go:277`). Sin eso, un directorio que pasa por un enlace simbólico, como `/tmp` o `t.TempDir()` en macOS, daría por inválido cada registro.
  - Solo considera archivos cuyo nombre cumple `^[a-f0-9]{32}\.json$` y coincide con el `ID` del registro.
  - Devuelve `false` cuando no hay ninguno.
- **`doctorDeps`** (en `cli`): `lookPath`, `runVersion(bin) (string, error)`, `processAlive(pid) bool`, `now`, `userHome` y `getenv`.
  - Las pruebas los reemplazan todos, así que ninguna prueba lee la configuración real del desarrollador.
  - Con `--home` sintético, `getenv` devuelve vacío, como en `setup.go:38-41`. Además, `lookPath` y `runVersion` no se usan: nada se detecta ni se ejecuta, como en `detectInstallerHosts`.
- **`collectDoctor(o, project, deps) doctorReport`:** arma las secciones CLIs, Installation, Sessions, Integrations y Project.
  - Un error de una sección queda dentro de ella y no aborta las demás.
  - T4 crea la función con las cinco secciones. Integrations y Project quedan como funciones de relleno en `doctor_integrations.go` y `doctor_project.go`, que T2 y T3 completan sin tocar `doctor.go`.
- **`renderDoctorText(report, w)` y `renderModelsText(rows, w)`:** la vista y el comando comparten el mismo texto por sección.

### Reglas por sección

- **CLIs:**
  - `installerHosts` en su orden.
  - Binario para la versión: el nombre del CLI, salvo `cursor`, que usa `cursor-agent`.
  - La versión solo se pide a un CLI detectado, con estas condiciones:
    - `exec.CommandContext` con argumentos fijos `--version` y sin shell;
    - 3 s de límite y `cmd.WaitDelay` de 1 s, para que un proceso nieto que retiene la salida no alargue la espera;
    - la primera línea recortada a 80 runas.
  - Los seis CLIs se consultan en paralelo, así que la carga dura como máximo unos 4 s.
  - Si `cursor-agent` no está en `PATH` aunque `cursor` sí, se muestra «CLI version unavailable: cursor-agent not found».
  - Release y estado de la instalación desde las filas de `Status` de ese CLI.
- **Installation:**
  - Filas de `Status` con estado distinto de `installed`, `retained_shared` y `not_installed` (`plan.go:772`).
  - Con una operación pendiente (`management.Pending`), las filas `recovery_required` se resumen en una línea con `hive recover`, en lugar de listarse una por una.
  - Cada estado tiene una frase fija en inglés, y un estado desconocido muestra su código tal cual.
  - Con al menos una fila `drift`, una sola línea explica que Hive no puede reparar un archivo cambiado y que el usuario debe deshacer el cambio, arreglar permisos o restaurarlo, y luego revisar con `hive status` (D1-B, M4).
- **Sessions:**
  - **Resumen:** con Claude Code o Grok registrado, la primera línea responde si hay que reiniciar algo («N open sessions should be restarted» o «No open session needs a restart», con el CLI que no se pudo comprobar), y la línea de cada CLI suma su cuenta («claude: 3 open sessions, 1 to restart») (D1-B, M4).
  - **Hora de referencia:** el `LastWrittenAt` de la release que el CLI tiene instalada.
  - **Claude Code:** el directorio sale de `CLAUDE_CONFIG_DIR` o de `~/.claude`, `sessions/*.json`. Se ignoran los archivos que no son `<número>.json`.
  - **Grok:** el directorio sale de `GROK_HOME` o de `~/.grok`, `active_sessions.json`.
  - **Directorios:** `CLAUDE_CONFIG_DIR` y `GROK_HOME` se leen con `doctorDeps.getenv`.
  - **Formatos:** en Claude Code, `startedAt` son milisegundos Unix. En Grok, `opened_at` se lee con RFC3339Nano. Los campos desconocidos se toleran.
  - **Sesiones vivas:** se cuentan las de `pid` vivo, comprobado con `syscall.Kill(pid, 0)`, donde `nil` o `EPERM` significan vivo. Un `pid` de 0 o menos se descarta, porque `Kill` lo daría por vivo.
  - **Marca por sesión:** desactualizada si empezó antes de la hora de referencia, al día si no.
  - **Datos que se muestran:** solo `pid`, `cwd` y la hora de inicio, nunca el contenido de otros campos.
  - **Codex, Pi y Cursor:** una línea por CLI registrado, que dice que Hive no ve sus sesiones y que hay que reiniciarlas tras actualizar.
  - **OpenCode:** una línea que dice que recarga en el siguiente mensaje.
  - **Límites declarados:** la marca es una estimación, por estas razones:
    - `releases/<id>.json` se reescribe en cada transacción que lleva esa release (`apply.go:399-407`), como agregar otro CLI. Ningún comando repara un `drift`: `install`, `update` y `plan remove` se niegan mientras un archivo gestionado difiere (`TestNoCommandRepairsADriftedManagedFile`). Eso adelanta la hora de referencia y puede marcar sesiones de más.
    - Un `/compact` en Claude Code o un cambio de voz no la mueven.
    - `LastWrittenAt` tiene precisión de segundos.
    - Un `pid` reutilizado puede hacer pasar por viva una sesión muerta.
- **Integrations:**
  - **Engram:** `lookPath("engram")`.
  - **Context7:** los archivos de `context7Candidates`.
  - **pi-subagents:** solo el registro, porque detectarlo exige leer la configuración de Pi.
  - **`agent-browser`:** `lookPath("agent-browser")` y `skills/agent-browser/SKILL.md` en las mismas raíces que `context7Candidates`.
  - A cada fila se le agrega el estado del último registro, cuando lo hay, y la fuente de `providers.Catalog`. El siguiente paso se calcula por fila y por caso (encontrado, no encontrado, estado del registro); el único comando que sugiere es `npx ctx7@latest setup --cli` (D1-B, M1). `agent-browser` no entra al catálogo de instalación: su fuente se fija en el código de la vista.
- **Project:**
  - **Raíz del repositorio:** `git -C <dir> rev-parse --show-toplevel`, donde `<dir>` es el directorio actual o `--project`.
  - **Sección:** en `AGENTS.md` de esa raíz se busca la línea exacta `## Hive`. Sus elementos son las líneas `- Clave: valor` hasta el siguiente encabezado `#`.
  - **Claves conocidas:** una tabla `hiveSettingKeys` con `Project`, `Base branch`, `Tracker` y `Specs` como obligatorias, y `Environments`, `Review`, `Delivery` y `Hive guidance` como opcionales.
  - **Specs:** la ruta es el texto antes de ` · `, resuelto desde la raíz del repositorio.
  - **Base branch:** se valida la primera palabra del valor.
    - Un valor que empieza con `-` o contiene `@{` se rechaza como inválido sin llamar a `git`.
    - Luego se valida con `git check-ref-format --branch`.
    - Por último se busca con `git -C <raíz> show-ref --verify --quiet refs/heads/<b>` o `refs/remotes/origin/<b>`.
  - **Entorno de `git`:** todas las llamadas usan el entorno de `filteredGitEnv` (`tooling/cli/update.go:224`), para que un `GIT_DIR` heredado no anule `-C`.
  - **Formato de los hallazgos:** cada uno es una línea con su clave.
- **Texto que viene de fuera:** versión, `cwd`, rutas y valores de `AGENTS.md` pasan por `sanitizeLine`, que quita caracteres de control y secuencias de escape antes de mostrarse. Así, un archivo ajeno no puede alterar la terminal.

### Vistas

Las cuatro son de solo lectura y siguen el patrón de carga de CLIs y Releases:
- **Carga:** un `tea.Cmd` con mensaje dirigido a la vista (`owned{v}`) y un número de secuencia (`seq`), como en `tui_hosts_view.go:40,229-305` y `tui_releases_view.go:85-143`. Una recarga doble o un resultado que llega después de salir de la vista se descartan.
- **`spinner`:** se muestra mientras cargan.
- **`r`:** recarga la vista. El raíz solo trata Ctrl-C, Esc y Backspace, así que no hay conflicto.
- **Esc y Backspace:** cada vista devuelve `action{}`, no `navNone`, para que el raíz vuelva al menú.
- **Error de carga:** un error de `EffectiveModels`, de `LastOnboarding` o de `Status` se muestra dentro de la vista con «r to retry», no impreso.
- **Sin color:** el cursor es `>`, y el CLI elegido en Models va entre corchetes (`[claude]`), como en el diseño anterior (`archive/2026-09-29-gh-46-tui/design.md:89`).

**Presupuesto de filas.** El área de una vista a 80×24 es de 21 filas (24 menos las 3 del marco). `scrollBox` recibe lo que queda después de las filas fijas de cada vista:

| Vista | Filas fijas | Contenido y teclas | Vacío y error |
| --- | --- | --- | --- |
| Diagnostics | Título y posición (2) | Las secciones CLIs, Installation y Sessions con encabezados, en un solo `scrollBox`; ↑↓, PgUp y PgDn desplazan. Con el estado ilegible, la explicación completa va solo en CLIs, e Installation y Sessions dicen «Not checked: Hive's state could not be read (see above).» | Sin CLIs registrados, Installation y Sessions dicen «No CLI hosts are registered» |
| Models | Título, fila de CLIs, encabezado de la tabla, posición y pie de 2 líneas (6) | ←→ cambia de CLI; tabla Role, Profile, Model y Effort en `scrollBox`. Anchos a 80 columnas: Role tan ancho como el rol más largo (hoy 24), Profile 10, Model 28 como mínimo y Effort 8, con separaciones de 2; Model se recorta con «…» y Role nunca se recorta; a 120 columnas, Model crece. Pie: cómo cambiarlos (D2-A) | Sin CLIs: «No CLI hosts are registered». Un CLI sin agentes: «No agents installed for <host>» |
| Integrations | Título, encabezado y las 4 filas de la lista (6) | ↑↓ mueve el cursor `>`. Debajo, el detalle de la fila elegida (fuente, estado del registro y siguiente paso) ocupa las filas restantes (15 a 80×24) en un `scrollBox` que ajusta el texto sin recortarlo y se desplaza con PgUp y PgDn | Sin registro: «No onboarding record yet» en la columna de estado |
| Project | Título y posición (2) | La ruta del `AGENTS.md`, o `Directory: …` fuera de un repositorio, como primera línea del `scrollBox`, ajustada con sangría en las continuaciones en lugar de recortada; después «Valid» o la lista de hallazgos, y los valores leídos | Fuera de un repositorio, dos líneas: «Not inside a Git repository; run hive from a repository.» y «Workspace-level ## Hive sections are not checked.» |

El menú agrega las cuatro entradas antes de Quit: 9 filas más el título ocupan 11 de las 21. Las pruebas del menú cambian así:
- `menuRowPattern` (`tui_test.go:313`) reconoce los nueve nombres;
- `TestAppMenuHasFiveEntriesInOrder` pasa a exigir nueve entradas en orden;
- la prueba de Esc y Backspace (`:413-414`) incluye las cuatro vistas nuevas.

## Especificación

Un requisito nuevo en `versioned-installation` (`specs/versioned-installation/spec.md` de este cambio): diagnóstico de solo lectura. Dice lo siguiente:
- Las vistas y comandos no escriben.
- Solo se ejecuta `--version` de CLIs detectados.
- No se leen configuraciones de terceros.
- `agent-browser` se detecta sin entrar al catálogo de instalación.
- La marca de sesión es una estimación.

El requisito «Interactive terminal interface» se reemplaza completo para nombrar las vistas nuevas en su lista.

## Límites de seguridad

- **Ejecución de procesos:** binarios con nombre fijo, resueltos por `PATH` y solo si están detectados. Argumentos fijos, sin shell y con límite de tiempo. `git` recibe la ruta del proyecto como argumento de `-C` y el nombre de rama ya validado. Nunca se interpola en un shell.
- **Datos no confiables:** archivos de sesión, `AGENTS.md` y salida de `--version`. Se aplican límite de tamaño (1 MiB), decodificación tolerante y `sanitizeLine`. No se leen los `.key` ni otros campos.
- **Dependencias:** ninguna nueva.

## Compatibilidad

- Los comandos existentes no cambian su salida ni sus opciones.
- El estado y los formatos de plan no cambian.
- Los ajustes de modelos y la edición de `## Hive` quedan para el segundo cambio de #46. Allí se decidirá si el ajuste vive en el estado, como la voz, o en la fuente.
