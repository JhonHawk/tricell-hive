# Tareas

Base del build: `feat/flash-class-guidance` en `982f9828` (registrada el 2026-10-13 al arrancar; `go vet ./...` y `go test ./...` en verde en esa base). Todas las órdenes corren desde el worktree `.claude/worktrees/feat-flash-class-guidance`.

### T1 — Subcomando del hook

- [x] `hive grok-hook read-whole --skills <dir>` reescribe las lecturas que cumplen y deja pasar las demás.

**Closes:** AC1, AC2.

**Depends on:** none.

**Locations:** paquete nuevo `integrations/grok/readhook` (lógica y pruebas); `tooling/cli/main.go` (`run`, rama `grok-hook` sin entrada en `printHelp`); archivo nuevo `tooling/cli/grok_hook.go`.

**Execution:** delegada a `hive-build-backend`, en paralelo con T2: T2 no toca `tooling/cli` salvo el `TestMain` de `tui_hosts_test.go`, y el contrato de línea de comandos está fijado en `design.md`.

**Test approach:** tdd.

**Changes:** según "Subcomando del hook" en `design.md`. Las pruebas cubren cada caso de AC1 (clave `target_file`, clave `path`, ruta relativa contra `workspaceRoot` y, sin él, contra `cwd`, otras claves conservadas) y de AC2 (fuera de la carpeta, enlace simbólico que sale de ella, no `.md`, más de 1000 líneas, más de 64 KiB, inexistente, sin `offset`/`limit`, JSON inválido, argumentos inválidos, pánico simulado con salida 0 y stdout vacío).

**Verification:** `go test ./integrations/grok/... ./tooling/cli/ -run 'ReadWhole|GrokHook'` pasa; sin el cambio, las pruebas nuevas fallan.

### T2 — Tipo de recurso `hook` para Grok

- [x] Instalar, actualizar y quitar Grok gestiona `<GrokHome>/hooks/hive-read-whole-references.json` con propiedad por archivo y un estado que se vuelve a leer.

**Closes:** AC3, AC4, AC5, AC7.

**Depends on:** none (usa la línea de comandos fijada en `design.md`, no el código de T1).

**Locations:** `integrations/grok/grok.go` (`Resolve`, constante `ReadHookSource`, render del JSON); `tooling/management/plan.go` (variable exportada del ejecutable junto a `skillsDirFor`, `buildPlan`, `nextRecord`, `validSource`, comprobación de archivo ajeno en `:451`, `validatePlan`); `tooling/management/types.go` (`Plan.Executable`); `tooling/management/ownership.go` (`normalizeState`, `stateSources`); `tooling/management/files.go` (`owned`, `transform`); la validación de la release si rechaza fuentes generadas; prueba nueva `tooling/management/grok_hook_test.go` siguiendo `cursor_test.go`; los `TestMain` existentes `tooling/management/sync_test.go:18` y `tooling/cli/tui_hosts_test.go:61` para fijar el ejecutable a un archivo regular fuera del directorio temporal; `validateRelease` para rechazar la fuente constante como archivo de release.

**Execution:** delegada a `hive-build-backend`, en paralelo con T1.

**Test approach:** tdd.

**Changes:** según "Tipo de recurso `hook`" en `design.md`. Pruebas: instalar escribe el JSON con matcher `read_file`, la ruta del ejecutable entre comillas y la carpeta de skills, lo registra como `hook` con modo 0644, y `status`, `update` y quitar vuelven a leer el estado sin error; quitar lo borra y deja un `engram-subagent-stop.json` ajeno intacto, también cuando `hooks/` existía antes; un archivo previo ajeno en la ruta falla con colisión; un archivo gestionado editado se rechaza; reinstalar con otro ejecutable reescribe el comando; `BuildModelsPlan` sobre Grok con el mismo ejecutable no rehúsa; un ejecutable bajo el directorio temporal falla al planificar; `apply` de un plan con `executable` relativo o inexistente falla; un plan guardado sin el campo conserva su ID.

**Verification:** `go test ./tooling/management/ ./integrations/grok/ ./tooling/cli/ -run 'GrokHook|Grok|Models'` pasa; sin el cambio, las pruebas nuevas fallan.

### T3 — Documentación del manager

- [x] La documentación describe el hook, su propiedad, su compatibilidad, su dependencia de la ruta del binario, cuándo lo carga Grok y cómo quitarlo.

**Depends on:** T1, T2.

**Locations:** `_support/docs/architecture/deployment-manager.md` (ejemplos con `go run` en las líneas 3 y 16, que ahora necesitan `go build` para Grok; "Content and native destinations" con el tipo `hook` y la fuente generada; el corredor de pilotos con `go run` ya no instala Grok; párrafo de Grok; "Preservation and state"); `_support/docs/architecture/repository-and-distribution.md` ("Core, references and optional native supply", donde esta rama trata los hooks); `README.md` (fila de Grok en "Hosts"; línea 226, borrar paquetes viejos deja el hook sin efecto hasta reinstalar o actualizar; reiniciar Grok o recargar con `r` en `/hooks` tras instalar o quitar); `CHANGELOG.md`.

**Execution:** hilo principal: depende del resultado de T1 y T2 y el brief sería más largo que el cambio.

**Test approach:** check — `rg -n 'hive-read-whole-references|grok-hook' _support/docs README.md CHANGELOG.md` encuentra cada sección nombrada; `llms.txt` y `site/` no cambian porque no afirman nada sobre lo que se instala en Grok (`rg -i grok site/src` vacío el 2026-10-13; `llms.txt` no menciona hooks ni rutas).

**Verification:** la orden anterior lista las secciones; `AGENTS.md` "README and llms.txt" queda cumplido.

### T4 — Prueba real en Grok

- [x] Una sesión headless de Grok lee completa una referencia pedida con `limit: 40`, en la sesión principal y en un hijo.

**Closes:** AC6.

**Depends on:** T1, T2.

**Execution:** hilo principal (D2-A). Antes de correrla se avisa al usuario qué se escribe y dónde, y que las sesiones de Grok que arranquen durante la prueba cargarán el hook hasta reiniciarse.

**Test approach:** check — en vivo.

**Changes:**

1. Compilar el binario con `go build -o <scratchpad>/hive-hooktest ./tooling/cli` en el scratchpad de la sesión, no en `/tmp`.
2. Guardar `shasum -a 256` de `~/Library/Application Support/tricell-hive/state.json`, el estado de la instalación en macOS (`deployment-manager.md`, "Preservation and state").
3. Escribir a mano, sin el manager, `~/.grok/hooks/hive-read-whole-references.json` con el JSON de `design.md` apuntando a ese binario y a `~/.agents/skills`.
4. Correr `env SHELL=/bin/zsh grok -p` en un directorio del scratchpad pidiendo leer `~/.agents/skills/flow-build/references/verification.md` con `limit: 40` y lanzar un hijo que haga la misma lectura con `limit: 40`.
5. Borrar primero el JSON y después el binario temporal; comprobar que el hash del estado no cambió.

**Verification:** en el `chat_history.jsonl` de la sesión (`~/.grok/sessions/<cwd codificado>/<id>/`) y en el del hijo (localizado por `subagents/*/meta.json`), la llamada registrada lleva `limit: 40` y su resultado incluye la línea 110 del archivo. Se anotan aquí los IDs de ambas sesiones y la clave de ruta que llegó al hook si difiere de `target_file`.

## Verification and human review

| Gate | Timing and environment | Mechanism and expected evidence | Authority / pending condition |
| --- | --- | --- | --- |
| Pruebas Go completas | Candidato final, worktree | `go vet ./...` y `go test ./...` en verde | Parte del build |
| Revisión de código | Candidato final, antes de pedir el commit | `/code-review` de Claude Code sobre el diff completo; hallazgos corregidos o refutados con evidencia | D4-A, elegida el 2026-10-13 |
| Prueba real en Grok | Tras T1–T3 | T4 | D2-A, elegida el 2026-10-13 |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` con un modelo distinto del implementador | Parte del build |

No hay interfaz gráfica: no aplican `hive-review-ux` ni `hive-verify-change` con navegador.

## Review status and progress

Revisión del plan, ronda 1 (2026-10-13, candidato `c3e482d2`/`a32ed0eb`/`d712e7e3`), dos `hive-review-plan` en paralelo:

- Manager y contrato del hook: tres bloqueantes aceptados y corregidos: fuente constante con validación como `plugin` (sin ella el estado guardado no se vuelve a leer, `ownership.go:59-61`); ejecutable resuelto en `management` para todos los constructores de planes, incluido `BuildModelsPlan`; T1 y T2 sin escritor compartido en `tooling/cli/main.go`. No bloqueantes aceptados: prueba de `update`, `validatePlan` del ejecutable, modo 0644 y comprobación de archivo ajeno.
- Entrega en Grok y prueba real: dos bloqueantes aceptados y corregidos: T4 escribe el hook a mano y comprueba que el estado no cambie; binario temporal (`go run`) rechazado y dependencia de la ruta documentada. No bloqueantes aceptados: pánico que sale con 2 (deny), evidencia de `limit: 40` en la llamada registrada, recarga de sesiones abiertas, límite `hashline_read`, rutas relativas contra `workspaceRoot`, línea 226 del README.

Re-revisión del dominio del manager (2026-10-13, candidato `eb0ddc4a`/`548bf002`/`990badd6`, mismo revisor reanudado): sin bloqueantes; los cuatro no bloqueantes se aplicaron tal como los propuso (los `TestMain` existentes y su ejecutable fuera del temporal, `Resolve` emite el hook siempre, la fuente constante rechazada como archivo de release, nota del corredor de pilotos), así que no hay otra ronda. Las correcciones de entrega aplicaron lo que propuso su revisor y no necesitaron re-revisión. Límite: ningún revisor ejecutó código ni Grok.

Avance: build arrancado el 2026-10-13 (B1-A). Modelos: implementadores `hive-build-backend` = sonnet (configurado, no observado); verificador `hive-verify-task` = opus (configurado, no observado); difieren.

- T1: implementada (RED 6 pruebas positivas sin salida y `undefined: runGrokHook`; GREEN con `go test ./integrations/grok/... ./tooling/cli/ -run 'ReadWhole|GrokHook'`). Desvíos aceptados: un evento con `toolInputTruncated: true` pasa sin cambios; `--skills` y la base de una ruta relativa deben ser absolutos; el JSON reescrito ordena sus claves. Verificada por `hive-verify-task` (Opus 5.5, observado): AC1 y AC2 met; abrí sus localizadores (`integrations/grok/readhook/readhook.go:32-118`, `tooling/cli/grok_hook.go:13-34`). Su orden filtrada dio FAIL solo porque `./integrations/grok/...` incluye la prueba roja de T2 (`undefined: RenderReadHook`); repetida tras T2: verde.
- T4: hecha el 2026-10-13 con Grok 1.0.55 (`grok --yolo -p`, `grok-4.7-build-fast`). Sesión `01a1378c-f578-7423-9f00-be8b5b0bdbf1` y su hijo `hive-research` `01a1378d-3110-7e92-bce1-2a3105bae29a`: en ambos `chat_history.jsonl` la llamada registrada es `{"target_file":"…/flow-build/references/verification.md","limit":40}` y su resultado (28 371 caracteres) llega a la línea 110. AC6 met. La ruta llegó en `target_file`. Limpieza: JSON del hook borrado primero, binario temporal después; hash del estado de Hive igual antes y después (`76c9dbe151a75045`).
- Gates finales (candidato con T1–T4): `gofmt -l` vacío, `go vet ./...` y `go test ./...` en verde (18 paquetes). `/code-review` medium de Claude Code: P1 `RequiredHosts` (`tooling/management/versioning.go:206`) previsualiza `nextRecord` sin `Executable` y aborta el siguiente install/update con Grok una vez registrado el hook; P2 sin bit de ejecución exigido, `TestMain` sin resolver enlaces, registro `hook` sin comprobar su carpeta `hooks/`. Ronda de corrección enviada al implementador de T2. P3 solo reportados: `hive models` rehúsa tras mover el binario (diseño aceptado y documentado), rutas relativas contra `workspaceRoot` (el esquema de `read_file` dice "relative to the workspace"), carrera entre la comprobación y la lectura, hook atado a la ruta del binario (documentado), listas de tipos duplicadas.
- Ronda de corrección (implementador de T2): `RequiredHosts` saltaba los grupos `hook`; bit de ejecución exigido; `TestMain` con enlaces resueltos; registro `hook` exige carpeta `hooks/`; cada una con prueba roja antes del arreglo. Re-revisión única de `/code-review` sobre esas correcciones: confirmó el P1 corregido y señaló que saltar el hook omitía la puerta "shared resource crosses scopes". Corregido sin revisión (local, sin cambio de contrato): la vista previa de `RequiredHosts` resuelve el ejecutable como `buildPlan` en lugar de saltar el hook (`tooling/management/versioning.go:207-213`), con una prueba que mueve el binario y otra que exige error sin ejecutable (esta falla con la versión que saltaba); la prueba del bit de ejecución exige ese motivo. Solo reportados (P3): el bit de ejecución no se compara con el usuario actual; el registro `hook` no se compara con `ReadHookPath` del home; la prueba escribe un archivo temporal en `testdata/`; con el checkout bajo `$TMPDIR` las pruebas de Grok fallan por la regla de binarios temporales; `checkHookExecutable` corre varias veces por instalación.
- Candidato final: `gofmt -l` vacío, `go vet ./...` y `go test ./...` en verde.
- T2: implementada (RED sobre stubs en 13 pruebas nuevas; GREEN con `go vet ./...` y `go test ./tooling/management/ ./integrations/... ./tooling/cli/`). Desvíos aceptados: `checkHookExecutable` también rechaza rutas que cruzan un enlace simbólico; el chequeo de temporales cubre `os.TempDir()`, `GOTMPDIR`, `GOCACHE` y `UserCacheDir()/go-build`; `validatePlan` rechaza `executable` en planes que no escriben el hook; solo las instalaciones de usuario con Grok resuelven el ejecutable. Pruebas existentes ajustadas porque el hook es un destino más: `TestResolveGrokUserTargets`, `TestFiveHostSharedConsumersAndPartialRemoval`, el golden de `model_override_state_test.go` (excluye el hook) y los ayudantes de `tooling/cli/tui_hosts_test.go` (normalizan la ruta del home dentro del hook), además de los dos `TestMain`. Verificada por `hive-verify-task` (Opus 5.5, observado), también con la CLI real sobre un home sintético: AC3, AC4, AC5 y AC7 met; abrí sus localizadores (`tooling/management/plan.go:72-125`, `ownership.go:69-73`, `integrations/grok/hook.go:40-63`). Huecos menores anotados por el verificador: ninguna prueba ejerce la rama `GOCACHE`/`UserCacheDir()/go-build` (go run en Go 1.27 usa `$TMPDIR`, sí cubierto), y `collectFiles` de `tui_hosts_test.go` ahora normaliza el home en todos los archivos.
- T3: hecha y comprobada con `rg -n 'read-whole' _support/docs README.md CHANGELOG.md`, que encuentra `deployment-manager.md` (nota de `go run` en Commands y sección "Grok read-whole hook" con los desvíos de T2), `repository-and-distribution.md:63`, `README.md:177` y la advertencia de borrar paquetes, y `CHANGELOG.md:15`; `site/src` y `llms.txt` no mencionan Grok ni hooks. El enlace de ancla y los documentos los cubre `tests/content` en el gate final.
