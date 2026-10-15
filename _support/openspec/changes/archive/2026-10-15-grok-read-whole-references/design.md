# Diseño

## Verified context

Leído el 2026-10-13 en `feat/flash-class-guidance` (`982f9828`); el código de `tooling/` e `integrations/` de Grok coincide con `development`.

- Grok 1.0.55 (`grok --version`). `read_file` (definición de herramienta de la sesión carvento `01a1371f`) lee hasta 1000 líneas, falla si la lectura es demasiado grande e ignora `offset`/`limit` solo en `SKILL.md`, `AGENTS.md` y `CLAUDE.md`.
- Hooks de Grok (`~/.grok/docs/user-guide/10-hooks.md`): los globales viven en `~/.grok/hooks/*.json` y siempre son de confianza (tabla "Hook Locations"); un grupo `PreToolUse` con `matcher: "read_file"` recibe el evento por stdin (`toolName`, `toolInput`, `cwd`); la respuesta `{"hookSpecificOutput":{"hookEventName":"PreToolUse","updatedInput":{…}}}` reemplaza la entrada completa en silencio (líneas 378 y 393); una entrada que no cumple el esquema bloquea la llamada (línea 395); un comando que no arranca o sale distinto de 0 y 2 falla abierto ("Exit Codes"). La documentación muestra la ruta como `target_file` (esquema de la herramienta) y como `path` (`14-headless-mode.md:270`): no verificado cuál llega al hook.
- Inferido, no verificado: el hook corre también en las sesiones de los hijos (`10-hooks.md` describe `UserPromptSubmit` "observe-only" en sesiones de subagente, lo que implica que los hooks se cargan allí). Lo comprueba AC6.
- Manager: los destinos de Grok son un bloque, skills y roles (`integrations/grok/grok.go:23-38`, `:284-292`). No existe un tipo para hooks: `normalizeState` acepta solo `block|skill|agent|symlink` (`tooling/management/ownership.go:48`); `nextRecord` rechaza otros tipos (`tooling/management/plan.go:614`); `transform` borra y protege por archivo solo `skill` y `agent` (`tooling/management/files.go:332-342`); `validSource` admite solo fuentes de `content/` (`plan.go:265-273`).
- Propiedad por archivo en directorios compartidos: los roles en `<GrokHome>/agents/` refusan un archivo ajeno previo (`plan.go:433-435`, `files.go:337-338`, prueba `TestCursorForeignAgentFileConflicts` en `tooling/management/cursor_test.go:160`).
- La carpeta de skills desplegada es `grok.SkillsDir(c)` = `<Home>/.agents/skills` (`integrations/grok/grok.go:17-19`); el manager ya escribe esa ruta absoluta en el contenido desplegado (`skillsDirFor`, `plan.go:33-55`).
- El binario `hive` vive en una ruta por versión (`~/.local/share/hive/packages/hive-<version>-<os>-<arch>/bin/hive`, `README.md:70-71`) y no siempre está en `PATH`.
- Reglas del repositorio: `AGENTS.md` pide una necesidad medida para hooks y no agregar runtime; aquí la medición son 22 sesiones reales y el ejecutor es el binario existente.

## Proposed design

### Subcomando del hook

`hive grok-hook read-whole --skills <dir>` es un comando oculto (no aparece en la ayuda). Lee un evento JSON de stdin y:

1. Toma `toolInput`; si no tiene `offset` ni `limit`, termina sin salida.
2. Toma la ruta de `target_file` (la clave requerida del esquema de `read_file`, verificada en la definición de herramienta de la sesión carvento `01a1371f`) o, si falta, de `path`; si es relativa, la resuelve contra `workspaceRoot` del evento y, si falta, contra `cwd`.
3. Exige que la ruta limpia, con enlaces simbólicos resueltos cuando existen, quede dentro de `<dir>` (también resuelto), que termine en `.md`, que sea un archivo regular, que tenga hasta 1000 líneas y hasta 64 KiB (por debajo del tope de unos 25 000 tokens de lectura de Grok).
4. Si se cumple, imprime la respuesta con `updatedInput` = `toolInput` sin `offset` ni `limit`, conservando las demás claves y valores tal como llegaron.

Ante cualquier error (JSON inválido, archivo ilegible, argumentos inválidos, pánico) no imprime nada y sale con 0: Grok sigue con la lectura original. La salida 2 es `deny` en `PreToolUse` (`10-hooks.md`, "Exit Codes") y un pánico de Go sale con 2, así que la entrada del subcomando recupera pánicos (`recover`) y parsea sus argumentos sin `flag.ExitOnError`. La lógica vive en un paquete probable sin E/S de proceso (por ejemplo `integrations/grok/readhook`), y el comando de la CLI solo conecta stdin, stdout y la bandera.

Límites conocidos: con `toolset.file_toolset = hashline` la herramienta se llama `hashline_read` y el matcher `read_file` no la cubre (falla seguro, sin efecto); las sesiones abiertas no cargan el hook hasta reiniciarse o recargar con `r` en `/hooks` (`10-hooks.md`, "Mid-Session Reload"); los hijos sí lo ejecutan (`10-hooks.md:595`).

### Tipo de recurso `hook`

Sigue el precedente de `plugin` (cargador de contexto de OpenCode, `tooling/management/ownership.go:62-66`, `plan.go:562-573`): un archivo generado con fuente constante.

- `grok.Resolve` agrega siempre, sin depender de la lista de fuentes, un destino `Kind: "hook"` en `<GrokHome>/hooks/hive-read-whole-references.json` con fuente constante `grok.ReadHookSource` (por ejemplo `integrations/grok/read-whole-hook`) y consumidor Grok de alcance `user`. La lista de fuentes de una instalación sale de `releaseSources`, que nunca contiene esa constante (`ownership.go:123-125`), así que copiar el bucle de `plugin` (`opencode.go:43-51`) no instalaría el hook.
- La fuente constante se acepta solo en registros `hook`, no como archivo de una release: `normalizeState` acepta `hook` solo con esa fuente, ese nombre de archivo y consumidor `grok`/`user`, como hace con `plugin`, y `validateRelease` la rechaza como ruta de archivo (`plan.go:325`). `stateSources` la trata como fuente generada, no como archivo de `content/`.
- El manager genera su contenido: un JSON `{"hooks":{"PreToolUse":[{"matcher":"read_file","hooks":[{"type":"command","command":"'<hive>' grok-hook read-whole --skills '<SkillsDir>'","timeout":10}]}]}}`, con `<hive>` la ruta absoluta canónica del binario y `<SkillsDir>` la de `grok.SkillsDir`. Grok corre el comando con `sh -c` (`10-hooks.md:215`), así que las rutas van entre comillas simples con el escape POSIX habitual.
- La ruta del binario la resuelve `management` con una variable exportada sustituible, como `skillsDirFor` (`plan.go:33`): por defecto `os.Executable` con enlaces resueltos; las pruebas de `management` y de `cli` la fijan en sus `TestMain` existentes (`tooling/management/sync_test.go:18`, `tooling/cli/tui_hosts_test.go:61`) a un archivo regular fuera del directorio temporal, por ejemplo bajo `testdata/`, porque el binario de prueba y `t.TempDir()` están bajo `TMPDIR`. Otros paquetes que construyen planes no instalan Grok (`tooling/package/context_loader_test.go:65`, `tests/pilot/guidance_variant_test.go:107-110`); el corredor de pilotos con `go run` (`tests/pilot/guidance_variant.go:214`) rehusará Grok, lo que se anota en la documentación porque los pilotos están en pausa. `buildPlan` la guarda en un campo nuevo `executable` del plan (`omitempty`, así los planes guardados conservan su ID: `types.go:253,273`) para que `validatePlan` y `apply` vuelvan a generar los mismos bytes. Ningún archivo de `tooling/cli` cambia para esto, así que todos los constructores de planes (instalar, `plan`, `update`, TUI, rollback, quitar, `BuildModelsPlan`) reciben la misma ruta.
- Si el ejecutable está bajo `os.TempDir()` o en la caché de compilación de Go (`go run`), un plan que instale el hook falla con un mensaje que pide compilar con `go build`. `validatePlan` exige que `executable` sea absoluto, limpio y un archivo regular cuando el plan trae un cambio de tipo `hook`.
- Propiedad como `agent`: modo 0644 explícito en `nextRecord`; un archivo ajeno previo falla en el plan (la comprobación de `plan.go:451` y `transform` incluyen `hook`: `unowned hook collision`); `owned` rechaza un archivo gestionado editado; quitar lo borra. Los directorios creados por Hive se limpian con `CreatedDirs` solo si quedan vacíos (`apply.go:158-174`), así que un `hooks/` previo con archivos ajenos queda intacto. La recuperación no necesita cambios (`apply.go:692-716`).
- Reinstalar o actualizar con otro ejecutable produce otro contenido gestionado y el plan lo reemplaza (AC5). `hive models` sobre Grok, si el binario cambió de ruta desde la última instalación, rehúsa con el mensaje existente "is not an agent file and would change" (`models.go:284`) hasta correr `hive update`; es aceptable y se documenta.
- `hive doctor` no necesita cambios: el estado nuevo aparece en Installation como cualquier registro.

### Compatibilidad

Un binario anterior rechaza un estado con `hook`. Es aceptable porque el repositorio ya pide reconstruir antes de desplegar; se documenta en `deployment-manager.md`. No hay migración de datos: el registro se crea en la próxima instalación o actualización de Grok. Borrar el paquete de `hive` que instaló el hook lo deja sin efecto (Grok muestra el fallo y sigue la lectura) hasta reinstalar o actualizar; se documenta en el README.

Descartado: un script `python3` o `sh` (D1-B, D1-C), porque agrega un runtime que el núcleo no exige o parsea JSON de forma frágil; una línea en la guía (R1-B), porque los arreglos solo de redacción no han funcionado en Grok.
