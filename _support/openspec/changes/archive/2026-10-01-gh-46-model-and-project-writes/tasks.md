# Tareas

**Base:** `d4c014d`, fijada al empezar el build el 2026-10-01, igual a la revisión del plan.

La integración ocurre en la rama `feat/gh-46-model-and-project-writes`, en el worktree `.claude/worktrees/gh-46-writes`. Cada línea trabaja en su propio worktree:
- línea A: `.claude/worktrees/gh-46-models`, rama `feat/gh-46-models`;
- línea B: `.claude/worktrees/gh-46-project`, rama `feat/gh-46-project`.

Las rutas son relativas a la raíz del repositorio. Ninguna prueba lee la configuración real del desarrollador: todas usan un home y un estado de prueba (`t.TempDir()`), y los repositorios de prueba se crean con `git init`.

Hay dos líneas de trabajo en paralelo, con escritores separados:
- **Línea A, modelos:** T1 → T2 → T3.
- **Línea B, proyecto:** T4 → T5.

T6 cierra las dos. El único archivo que comparten es `tooling/cli/main.go`: B agrega el `case "project"` y A no lo toca, porque `hive models` ya enruta en `models.go`.

## T1 — Ajuste por rol en el gestor

- [x] El estado guarda ajustes por CLI y rol. Cada plan `install` los vuelve a aplicar, `BuildModelsPlan` los cambia sin tocar nada más y los recibos siguen en `verified`.

**Closes:** AC2, AC3, AC4.

**Depends on:** ninguna.

**Locations:**
- `integrations/agents/agents.go`: `ModelOverride`, `ValidateOverride`, y el parámetro nuevo en `Resolve` y `Render`.
- `tooling/management/types.go`: los campos de `State` y `Plan`.
- `tooling/management/plan.go`: `BuildPlan` copia los ajustes, `nextRecord` los aplica y `validatePlan` los valida.
- `tooling/management/apply.go`: `prepareTransaction` copia, poda y borra los ajustes, y marca `changed`.
- `tooling/management/models.go`: `BuildModelsPlan`, `EffectiveModels` con ajustes y `ModelRow.Override`.
- Las pruebas de cada uno.

**Execution:** delegada a `hive-build-backend` en su propio worktree (línea A). Los contratos están en [design.md](design.md#ajuste-de-modelo-en-el-estado-d1-a-d2-a) y los archivos son solo de esta línea.

**Test approach:** `characterization` primero y luego `tdd`. Antes de tocar `Resolve`, se guarda como golden la salida completa de `Render` (20 roles × 6 CLIs) en la base, y después se exige que salga idéntica con ajuste `nil`.

**Changes:**
1. Guardar el golden de `Render` en la base.
2. Comprobar con Context7, en la documentación oficial actual, si Claude Code acepta `ultra` en `effort` y qué valores acepta `thinking` en Pi.
   - Si un CLI rechaza un valor de la lista, `ValidateOverride` se lo niega a ese CLI y la fuente queda anotada en el reporte.
   - Si la documentación no lo dice, se deja la lista del código y se anota como no verificado.
3. Agregar `ModelOverride`, `ValidateOverride` y el parámetro de `Resolve` y `Render`.
4. Agregar los campos del estado y del plan, y copiarlos y podarlos en `BuildPlan` y en `prepareTransaction`.
5. Escribir `BuildModelsPlan`, con la regla del recibo y los rechazos de design.md.
6. Agregar los ajustes a `EffectiveModels`.

**Verification:**
- `go test ./integrations/agents -run 'Golden|Override'` pasa, con estos casos:
  - el golden es idéntico con `nil`;
  - una tabla de `ValidateOverride` con cada rechazo de AC4: caracteres fuera de la lista, espacio, comilla, `#`, U+202E, 201 caracteres, esfuerzo en Grok y Cursor, un esfuerzo desconocido y un ajuste sin modelo ni esfuerzo. Un ajuste solo de esfuerzo, con `Model` vacío, se acepta;
  - `Resolve` falla con un esfuerzo en OpenCode sobre un perfil sin modelo y sin modelo en el ajuste;
  - el perfil `inherit` de OpenCode (`…#max`) con un ajuste `Effort: high` da `…#high`, no `…#max#high`.
- `go test ./tooling/management -run 'ModelOverride|ModelsPlan|EffectiveModels'` pasa, con estos casos sobre un home de prueba con Claude, Codex, Grok, OpenCode y Pi instalados:
  - **Aplicar un ajuste (AC1):** cambia solo el archivo del rol en ese CLI; se comparan los hashes de todos los demás archivos y de los otros CLIs. Escribe el valor esperado en cada formato:
    - `model` y `effort` en Claude;
    - `model_reasoning_effort` en Codex;
    - `thinking` en Pi;
    - `#variant` en OpenCode.
  - **Recibo (AC2):** `Status` da `verified` después del ajuste y después de un `BuildPlan("install")` desde otra fuente con un perfil cambiado, que además conserva el ajuste.
  - **Vuelta atrás (AC2):** una reinstalación con `ReleaseID` de una release anterior conserva el ajuste, y su estado de versión es igual al de la misma reinstalación sin ajustes.
  - **Quitar ajustes (AC3):**
    - `BuildModelsPlan` con un conjunto vacío deja el archivo idéntico al de una instalación limpia;
    - un `remove` del CLI borra sus ajustes del estado.
  - **Cambios solo de estado:** se guardan, con su `state.json` comprobado después:
    - el `reset` de un ajuste «not applied» (rol ausente de la release);
    - un `set` igual al valor de la release.

    Ningún mapa vacío queda en `state.json`.
  - **Rechazos del constructor:** con un recurso borrado a mano en el CLI, o con un `Legacy` pendiente, `BuildModelsPlan` falla con el mensaje que remite a `hive install` o `hive doctor` y no escribe nada. También falla con un rol inexistente o un CLI no registrado.
  - **Rechazos de `Apply` (AC4):** con planes editados a mano y su `ID` recalculado con `planID(p)`, `Apply` falla con el mensaje concreto de la validación, no con «invalid or legacy plan»:
    - un ajuste inválido;
    - un ajuste distinto del estado para un CLI fuera del plan;
    - un CLI desconocido.
  - **Plan viejo:** un plan guardado antes de cambiar otro ajuste es rechazado por `StateHash`.
  - **Otros planes con ajustes guardados:** `BuildVoicePlan` + `Apply` y un `remove` de otro CLI funcionan y conservan los ajustes de los demás CLIs.
  - **`set --effort` en OpenCode:** sobre un rol con perfil `inherit` (`…#max`), guarda un ajuste con `Model` vacío y escribe `…#<effort>`.
  - **Sin ajustes:** `state.json` es idéntico byte a byte al que escribe la base para la misma instalación.
- `go test -race ./tooling/management` pasa.

## T2 — Comandos `hive models set` y `hive models reset`

- [x]
  reopened by correction rounds: las rondas de corrección 1 y 2 cambiaron el resumen compartido en `models.go`; se verifica en la puerta del candidato aceptado `60d201e`.

 
  reopened by T8: la regla de OpenCode cambió `set --model` sobre un rol de OpenCode. Se reverificó con T8 y AC1, AC3 y AC4 se cumplen en `913bef4`.

  Los dos subcomandos construyen y aplican el plan de T1 con resumen y confirmación, y `hive models` marca los roles ajustados. Los dos subcomandos construyen y aplican el plan de T1 con resumen y confirmación, y `hive models` marca los roles ajustados.

**Closes:** AC1, AC3, AC4.

**Depends on:** T1.

**Locations:**
- `tooling/cli/models.go`: enrutado de los subcomandos, resumen, marca `*`, nota y la línea «not applied».
- `tooling/cli/main.go`: solo el texto de uso, si lo lista.
- Pruebas en `tooling/cli/models_command_test.go`.

**Execution:** la misma delegación de T1, en serie, porque usa su contrato y no comparte archivos con la línea B.

**Test approach:** `tdd`.

**Changes:**
1. Enrutar `set` y `reset` sin cambiar `hive models` sin subcomando.
2. Reutilizar `voicePlanWith`, o una función común extraída de ella, para el resumen, la confirmación, `--dry-run` y `--out`.
3. Mostrar la marca `*`, la nota y los ajustes no aplicados en la tabla.

**Verification:**
- `go test ./tooling/cli -run 'ModelsCommand'` pasa, con estos casos:
  - `set` con terminal falsa aplica, y `hive models` muestra `*` y la nota;
  - `--dry-run` no escribe;
  - `--out` más `hive apply` dan los mismos archivos;
  - `reset --all` vuelve a la instalación limpia;
  - `reset --role R --only effort` sobre un ajuste de modelo y esfuerzo deja solo el modelo, y `--only model` sobre un ajuste solo de modelo equivale a `reset --role R`;
  - cada rechazo de AC4 sale con el mensaje y sin escribir, incluido `--model ""`;
  - un `set` sin cambios dice «Nothing to change»;
  - sin terminal y sin `--dry-run` ni `--out`, falla sin escribir;
  - la salida de `hive models` sin ajustes es idéntica a la de la base (golden de texto).

## T3 — Edición en la vista Models

- [x] La vista Models tiene cursor, panel de edición, `x` para quitar un ajuste, confirmación, pie y barra de ayuda, según [design.md](design.md#edición-en-la-vista-models).

**Closes:** AC5, AC10.

**Depends on:** T2.

**Locations:** `tooling/cli/tui_models_view.go` y `tooling/cli/tui_models_view_test.go`.

**Execution:** la misma delegación, en serie. El presupuesto de filas sale de design.md.

**Test approach:** `tdd` con `appDriver`.

**Verification:**
- `go test ./tooling/cli -run 'ModelsView'` pasa, con estos casos:
  - **Lista:**
    - ↑↓ mueve `>` y llega al rol 20 a 80×24;
    - un rol ajustado muestra `*`;
    - el pie tiene las dos líneas de AC5;
    - después de aplicar, el cursor sigue en el mismo rol.
  - **Panel:**
    - Enter abre el panel con el modelo efectivo;
    - el texto escrito más Enter abre la confirmación con el CLI, el rol, el modelo y el esfuerzo antes «→» después, y la ruta;
    - Apply deja los mismos archivos y el mismo estado que `hive models set`;
    - cambiar solo el esfuerzo deja el mismo `state.json` que `hive models set --effort` sin `--model`;
    - vaciar `Model` deja el mismo `state.json` que `hive models reset --only model`;
    - Cancel dice «Cancelled. No changes applied.», y un plan sin cambios dice «Nothing to change»; en los dos casos `state.json` y los archivos conservan su hash.
  - **Teclas del panel:**
    - Esc en el panel lo cierra y deja la vista abierta;
    - Backspace en el campo borra texto, y `r` y `x` escritos en el campo aparecen en el texto;
    - Backspace con el foco en `Effort` deja la vista y el panel abiertos.
  - **Otros:**
    - en Grok, el esfuerzo dice «not supported by grok»;
    - `x` sobre un rol ajustado lo quita;
    - un error de validación de 200 caracteres ocupa una sola fila.
  - **Tamaño:** `assertFits` pasa a 80×24 y a 120×40 en la tabla, el panel con error y la confirmación. La barra de ayuda a 80 columnas incluye `esc back` y `ctrl+c quit` en la lista, y `esc close` en el panel.
- `go test ./tooling/cli -run 'Contrast'` sigue pasando.

## T4 — Escritura de `## Hive` y `hive project set`

- [x] `checkHiveText`, `editHiveSection`, `writeProjectFile`, `suggestHiveValues`, el aviso de `CLAUDE.md` y el comando funcionan según [design.md](design.md#escritura-de--hive-d3-a).

**Closes:** AC6, AC7, AC8.

**Depends on:** ninguna.

**Locations:**
- `tooling/cli/project_write.go` nuevo.
- `tooling/cli/project_command.go` nuevo.
- `tooling/cli/doctor_project.go`: se extraen el recorrido de la sección con posiciones y `checkHiveText`. `parseHiveSection` y `validateHiveSection` los usan, y su salida no cambia.
- `tooling/cli/main.go`: `case "project"` y el texto de uso.
- Pruebas en `tooling/cli/project_write_test.go` y `tooling/cli/project_command_test.go`.

**Execution:** delegada a `hive-build-backend` en otro worktree (línea B), en paralelo con T1. No comparte archivos con la línea A.

**Test approach:**
- `characterization` sobre `parseHiveSection` y `validateHiveSection`: las pruebas actuales de `doctor_project_test.go` pasan sin cambios después de la extracción.
- `tdd` para el resto.

**Changes:**
1. Comprobar con Context7 la afirmación sobre `CLAUDE.md` y `@AGENTS.md` en la documentación actual de Claude Code. Fijar el texto del aviso según lo que diga, y anotar la fuente en el reporte.
2. Extraer el recorrido de la sección y `checkHiveText`.
3. Escribir `editHiveSection` como función pura.
4. Escribir `writeProjectFile` y `suggestHiveValues`.
5. Escribir el comando con su resumen.

**Verification:**
- `go test ./tooling/cli -run 'HiveSection|HiveText|ProjectWrite|ProjectCommand|Suggest|DoctorProject'` pasa, con estos casos:
  - **Tabla de edición (AC6, AC7):**
    - reemplazar, agregar o quitar una clave;
    - claves desconocidas y prosa conservadas;
    - un bloque de código con `## Hive` dentro, ignorado;
    - CRLF conservado;
    - sin sección, se agrega al final, también en un archivo sin salto de línea final;
    - sección duplicada, rechazada;
    - clave repetida, rechazada.
  - **Tabla de rechazos (AC8):** cada caso comprueba el mensaje y que el archivo conserva su hash:
    - fuera de un repositorio;
    - `AGENTS.md` como directorio, es decir, no regular;
    - enlace simbólico;
    - archivo de más de 1 MiB;
    - obligatorio que falta y obligatorio vacío;
    - un valor con salto de línea, con `\x1b[31m` y con U+202E;
    - `Delivery: pr`;
    - `Hive guidance: optional`;
    - `Base branch: -x`;
    - una clave desconocida en `--set`;
    - un `--unset Project`.
  - **Avisos que no bloquean:**
    - con una rama inexistente y una ruta de `Specs` inexistente, se escribe y la salida lleva los dos avisos;
    - una clave desconocida ya presente se conserva y solo avisa.
  - **Escritura en un repositorio temporal:**
    - conserva los permisos `0600` del archivo existente;
    - crea `AGENTS.md` con `0644` aunque la umask sea `077`;
    - un archivo cambiado entre la lectura y la escritura se rechaza con «AGENTS.md changed since the preview»;
    - un `AGENTS.md` creado por otro entre la lectura y la escritura hace fallar el `os.Link` sin sobrescribirlo;
    - en ningún caso quedan temporales en el directorio;
    - un `CLAUDE.md` sin la importación da el aviso y conserva su hash.
  - **Sugerencias:**
    - con `git@github.com:o/r.git`, `Tracker` es `GitHub Issues · o/r` y `Project` es `r`;
    - con `https://x-access-token:SECRET@github.com/o/r.git`, la sugerencia no contiene `SECRET`;
    - con `https://github.com.evil.com/o/r`, no hay sugerencia de `Tracker`;
    - con `origin/HEAD` apuntando a `main`, `Base branch` es `main`;
    - con `_support/openspec` presente, `Specs` lo sugiere.
  - **Comando:** sin terminal y sin `--dry-run`, falla sin escribir.
  - **Doctor:** las pruebas de `doctor_project_test.go`, incluida la que compara `hiveSettingKeys` con la guía, siguen pasando.

## T5 — Formulario en la vista Project

- [x] La vista Project abre el formulario con `e`, sugiere valores, confirma y vuelve a validar, según [design.md](design.md#formulario-en-la-vista-project).

**Closes:** AC9, AC10.

**Depends on:** T4.

**Locations:** `tooling/cli/tui_project_view.go` y `tooling/cli/tui_project_view_test.go`.

**Execution:** la misma delegación de T4, en serie.

**Test approach:** `tdd` con `appDriver`.

**Verification:**
- `go test ./tooling/cli -run 'ProjectView'` pasa, con estos casos:
  - **Apertura:**
    - en un repositorio sin sección, con remoto de GitHub, `origin/HEAD` y `_support/openspec`, `e` abre las ocho filas con los obligatorios primero;
    - `Project`, `Base branch`, `Tracker` y `Specs` llevan «suggested».
  - **Campos:**
    - ←→ en `Delivery` alterna entre vacío y `direct-base`;
    - un obligatorio vacío da «Project is required» y no abre la confirmación.
  - **Sección existente:** el formulario carga sus valores actuales, sin «suggested».
  - **Confirmación:**
    - muestra las líneas antes y después, y «AGENTS.md has uncommitted changes» con el archivo modificado sin commit;
    - con un `CLAUDE.md` sin la importación, una rama inexistente y una ruta de `Specs` inexistente, muestra los tres avisos.
  - **Aplicar:**
    - Apply deja el mismo archivo que `hive project set` con los mismos valores, y la vista muestra «Valid»;
    - Cancel vuelve al formulario con lo escrito intacto, y el archivo conserva su hash;
    - un archivo cambiado entre la confirmación y Apply da «AGENTS.md changed since the preview» en el formulario y conserva lo escrito.
  - **Teclas:**
    - Esc en el formulario lo cierra y deja la vista abierta;
    - Backspace en un campo borra texto, y `r` y `e` escritos en un campo aparecen en el texto;
    - Backspace con el foco en `Delivery` deja la vista y el formulario abiertos.
  - **Fuera de un repositorio:** `e` no hace nada.
  - **Tamaño:**
    - `assertFits` pasa a 80×24 y a 120×40, con un valor de 200 caracteres en un campo;
    - la barra de ayuda a 80 columnas incluye `esc close` y `ctrl+c quit`.

## T6 — Especificación y documentación

- [x] `deployment-manager.md` y el delta de la especificación describen el cambio.

**Closes:** AC11.

**Depends on:** T3, T5.

**Locations:**
- `_support/docs/architecture/deployment-manager.md`: comandos, vistas, límites y la nota de compatibilidad del estado.
- `specs/versioned-installation/spec.md` de este cambio, si la implementación cambió algo de lo escrito.

**Execution:** hilo principal. Necesita el resultado integrado de las dos líneas y es un cambio de texto corto.

**Test approach:** `check`. Se verifica con `rg -n "models set|project set" _support/docs/architecture/deployment-manager.md`, que debe devolver líneas, y con `rg -n "never write models" _support/docs/architecture/deployment-manager.md`, que no debe devolver nada.

**Verification:** las dos búsquedas dan el resultado esperado y los ejemplos de comandos coinciden con `--help` de los binarios construidos.

## T7 — Lista de modelos de cada CLI

- [x]
  reopened by correction rounds: la ronda 2 agregó `Name` y `Provider` al catálogo; se verifica en la puerta del candidato aceptado `60d201e`.

  `listHostModels` devuelve los modelos de cada CLI con la ejecución acotada de [design.md](design.md#lista-de-modelos-del-cli-d6-a-t7).

**Closes:** AC12, AC13.

**Depends on:** ninguna (parte de la rama integrada `da2ef07`).

**Locations:** `tooling/cli/model_catalog.go` y `tooling/cli/model_catalog_test.go`, nuevos, con `testdata/catalog/<host>.txt` grabados.

**Execution:** delegada a un `hive-build-backend` nuevo en el worktree `.claude/worktrees/gh-46-catalog` (rama `feat/gh-46-catalog`), en paralelo con T8. Es un archivo nuevo con un contrato cerrado.

**Test approach:** `tdd`.

**Changes:**
1. Grabar salidas reales de los cinco comandos en `testdata`, recortadas a unas pocas entradas y sin la línea de sesión de Grok. La de Codex se reduce a `slug`, `visibility` y `supported_reasoning_levels`.
2. Escribir el lector de cada CLI y el runner acotado.

**Verification:** `go test -count=1 ./tooling/cli -run 'Catalog'` pasa, con estos casos:
- **Lectura de cada CLI con su salida grabada:**
  - Codex omite `hide` y da los esfuerzos de cada modelo;
  - OpenCode reintenta una vez con salida vacía y falla tras dos vacías;
  - Pi omite la cabecera;
  - Grok omite la línea de sesión y quita «(default)»;
  - Cursor omite el encabezado y el consejo.
- **Ids inválidos:** no se ofrecen.
- **Claude:** no ejecuta nada y devuelve los cuatro alias.
- **Runner real, con un binario falso en un `PATH` temporal y el límite de tiempo reducido por parámetro en la prueba:**
  - un proceso que duerme más que el límite se corta, y un nieto que lanzó también muere (se busca su PID después);
  - una salida mayor que el tope da error, no una lista parcial;
  - el falso graba su `argv`, su carpeta de trabajo y si su entrada estándar es nula: los argumentos son los fijos, la carpeta es `os.TempDir()` y no la del repositorio, y no hay terminal;
  - con `--home` no se ejecuta nada.
- **Nunca se ejecuta (AC13):** con un runner falso que cuenta llamadas, `hive models`, `hive doctor` y recorrer la vista Models sin abrir un panel dejan el contador en 0.

## T8 — Grupos en el gestor y en los comandos

- [x]
  reopened by correction rounds: la ronda 1 cambió el resumen de `set --group`; se verifica en la puerta del candidato aceptado `60d201e`.

  `ModelRow.Group`, la salida agrupada de `hive models`, y `--group` en `set` y `reset` funcionan según [design.md](design.md#grupos-por-función-d7-a-t8-y-t9).

**Closes:** AC14, AC15.

**Depends on:** ninguna.

**Locations:** `tooling/management/models.go` (`Group`), `tooling/cli/models.go`, y las pruebas de los dos.

**Execution:** delegada al `hive-build-backend` de la línea de modelos, en `.claude/worktrees/gh-46-models` actualizado a la rama integrada, en paralelo con T7.

**Test approach:** `tdd`. El golden de texto de `hive models` se regenera a propósito y el diff se revisa en el reporte.

**Verification:** `go test -count=1 ./tooling/cli ./tooling/management` pasa completo, porque `models.go` es compartido con la vista. Los casos:
- **Gestor:** `Group` sale de la ruta de cada rol.
- **`hive models`:**
  - imprime los grupos en orden con sus cabeceras;
  - la cabecera muestra «mixed» cuando los roles difieren.
- **`set --group review --model opus`:**
  - deja `{model: opus}` en cada rol de `review`;
  - reemplaza el ajuste propio que tenía un rol;
  - el resumen nombra ese rol en «Replaces the own override of».
- **`reset --group`:** quita los ajustes del grupo. Con `--only effort`, quita solo esa parte.
- **Rechazos, sin escribir:** `--group` con `--role`, y un grupo desconocido.
- **Equivalencia:** el estado que deja `--group` es el mismo que dejarían los `--role` equivalentes uno a uno.
- **Filas sin cambios:** quitando las cabeceras de grupo, la salida de `hive models` sin ajustes es idéntica al golden de la base.
- **Regla de OpenCode:**
  - `set --model x` sobre un rol `inherit` escribe `{model: x, effort: max}` y el archivo lleva `x#max`; es un caso de T2 que se reverifica;
  - `set --group … --model x` en un grupo con esfuerzos «mixed» conserva el esfuerzo de cada rol.

## T9 — Selector de modelos y grupos en la vista Models

- [x]
  round 1, AC10: un `display_name` con caracteres de doble ancho hace que el recuadro supere el ancho (recorte por caracteres, no por columnas). Además, `3a84ce4` borró cuatro pruebas de `29f99aa`.
  reopened by correction round 1: el usuario pidió separar visualmente los grupos y hacer legible la confirmación.

  La vista agrupa con cabeceras editables y el campo Model es una lista con «Other…», según [design.md](design.md#lista-de-modelos-del-cli-d6-a-t7) y [design.md](design.md#grupos-por-función-d7-a-t8-y-t9).

**Closes:** AC12, AC14, AC15, AC10.

**Depends on:** T7, T8.

**Locations:** `tooling/cli/tui_models_view.go`, `tooling/cli/tui_app.go` (el caché por sesión en `appConfig`), y las pruebas de la vista.

**Execution:** el mismo implementador de T8, en serie. Antes, el orquestador revisa el diff de T7 y lo integra en el worktree de modelos.

**Test approach:** `tdd` con `appDriver` y un `catalogRunner` falso inyectado por `appConfig`.

**Verification:** `go test -count=1 ./tooling/cli -run 'ModelsView|Contrast'` pasa, con estos casos:
- **Selector (forma final aceptada en las rondas 2–4):**
  - → o una letra en Model abren el recuadro «Select model»:
    - el modelo actual va marcado con `●` y el cursor arranca en él;
    - Claude, Codex, Grok y Cursor muestran una lista plana, y OpenCode y Pi secciones por proveedor;
    - «release default» y «Other…» están en la línea final;
  - la letra queda en la búsqueda, y escribir filtra;
  - Enter fija el modelo y vuelve al panel, y otro Enter abre la revisión;
  - «release default» deja el mismo estado que `reset --only model`;
  - Esc cierra la lista sin cambios;
  - «Other…» abre el campo de texto, donde ←→ mueven el cursor;
  - en Grok, Enter en el panel sigue revisando;
  - con 209 modelos falsos se llega al último con PgDn.
- **Esfuerzos en Codex:** cambian al elegir un modelo con otros niveles, y un esfuerzo que ya no está vuelve a «release default».
- **Fallo del runner:** muestra «Model list unavailable» y deja «Other…».
- **Carga:**
  - el runner se llama una vez por CLI, aunque se abra el panel dos veces o se salga de la vista y se vuelva;
  - un fallo se reintenta en la siguiente apertura;
  - el spinner gira mientras carga;
  - con `--home` no se llama.
- **Esfuerzo por valor:** un catálogo que llega con el panel abierto no cambia el esfuerzo elegido ni aparece un cambio que nadie pidió.
- **Grupos:**
  - las cabeceras aparecen en orden y el cursor cae en ellas;
  - Enter en una cabecera abre el panel del grupo;
  - Apply deja el mismo `state.json` que `hive models set --group`;
  - la confirmación empieza con «Replaces the own override of» y lista los roles;
  - la cabecera lleva `*` cuando algún rol está ajustado;
  - `x` en una cabecera equivale a `reset --group`;
  - un grupo de OpenCode con esfuerzo «mixed» y un modelo nuevo deja el mismo estado que `set --group --model`;
  - Backspace con el foco en Effort o en la lista no cierra la vista.
- **Ancho (AC10):** `assertFits` pasa a 80×24 y a 120×40 con la lista abierta, con un grupo «mixed» y en las confirmaciones de grupo.
- **Cobertura existente:** las pruebas de T3 siguen pasando. Las que dependen de la tabla plana, o de escribir en Model y revisar con Enter (`tui_models_edit_test.go:320-356`, `:425`, `:679`, `:262-268`), se adaptan al selector y quedan anotadas en el reporte.

## T10 — Documentación y especificación de la lista y los grupos

- [x] `deployment-manager.md` y el delta de la especificación describen la lista de modelos y los grupos.

**Closes:** AC16.

**Depends on:** T9.

**Locations:** `_support/docs/architecture/deployment-manager.md`, `specs/versioned-installation/spec.md` de este cambio, y el texto de uso de `tooling/cli/main.go:92`, que debe nombrar `--group`.

**Execution:** hilo principal, porque es un texto corto sobre el resultado integrado.

**Test approach:** `check`, con `rg -n "debug models|--group" _support/docs/architecture/deployment-manager.md`.

**Verification:**
- esa búsqueda devuelve líneas;
- el requisito «Per-role model overrides» del delta nombra la lista de cada CLI, su ejecución acotada y la edición por grupo;
- «Read-only diagnostics» dice que la vista Models ejecuta los comandos de listado solo al abrir un panel;
- `rg -n "No model catalog" specs/versioned-installation/spec.md _support/docs/architecture/deployment-manager.md` no devuelve nada.

## Verificación y revisión humana

| Puerta | Cuándo y dónde | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Por tarea | Al terminar cada tarea, en su worktree | `hive-verify-task` vuelve a correr la verificación de la tarea y da un veredicto por cada `AC<n>` | Parte de la implementación |
| Suite | Tras integrar las dos líneas | `go vet ./...`, `go test ./...` y `go test -race ./...` pasan; `go list -m all` es igual al de la base | Parte de la implementación |
| `hive-review-ux` y `hive-verify-change` | Tras T6, en paralelo, cada uno con su propio home y su `tmux` privado | Ver el detalle debajo | Parte de la implementación |
| Recorrido del usuario | Después de los dos anteriores | Ver el detalle debajo | Según el modo de entrega |

**Detalle de `hive-review-ux` y `hive-verify-change`:**
- **Binarios:** el del worktree integrado, con `go build -o <tmp>/hive ./tooling/cli`, y el de la base `d4c014d`, para comparar la vista Models sin ajustes.
- **Instalación de prueba:** `<tmp>/hive install --home <fixture> --state-dir <fixture>/state --hosts claude,codex,grok,pi,opencode`. El `--home` evita que las variables heredadas (`CODEX_HOME`, `XDG_STATE_HOME`, `PI_CODING_AGENT_DIR`, `XDG_CONFIG_HOME` y otras) la hagan escribir en la configuración real.
- **Apertura de la interfaz:** desde un repositorio de prueba dentro del fixture (`cd <fixture>/repo`), con `env -i HOME=<fixture> PATH=<fixture>/bin:/usr/bin:/bin CLAUDE_CONFIG_DIR=<fixture>/.claude GROK_HOME=<fixture>/.grok TERM=$TERM <tmp>/hive tui --state-dir <fixture>/state`. La vista Project usa el directorio actual, así que nunca apunta a un repositorio real.
- **`tmux`:** cada revisor usa su propio socket, `tmux -L <nombre>`, y nunca el servidor por defecto.
- **Repositorios de prueba:**
  - uno sin `AGENTS.md`, con remoto `git@github.com:o/r.git` y `origin/HEAD`;
  - uno con `## Hive` incompleta y un `CLAUDE.md` sin `@AGENTS.md`.
- **Tamaños y temas:**
  - 80×24 y 120×40, con fondo oscuro y con `NO_COLOR=1`.
  - El tema claro no se puede forzar dentro de `tmux`, porque `tui.go:81-85` lo detecta con `HasDarkBackground`. Lo cubre la prueba de contraste (`tui_contrast_test.go`), y el revisor lo declara como no observado en vivo.
- **`hive-review-ux`:** sigue los criterios de revisión de interfaz de `content/skills/flow-build/references/ui-review-criteria.md` sobre Models (editar, cancelar, quitar un ajuste, errores) y Project (crear el archivo, crear la sección, editar, rechazos y avisos).
- **`hive-verify-change`:** recorre AC1–AC11. Comprueba archivos de agente, `hive status` en `verified`, `AGENTS.md` y `CLAUDE.md` byte a byte, y que cancelar deja `state.json` y el repositorio iguales.

**Recorrido del usuario:** se hace sobre un home y un repositorio de prueba, no sobre la instalación real.
1. `go build -o "$SCRATCH/hive" ./tooling/cli`.
2. `F=$(mktemp -d)`; `"$SCRATCH/hive" install --home "$F" --state-dir "$F/state" --hosts claude,codex,opencode`.
3. `git init -b main "$F/repo"`; `git -C "$F/repo" commit --allow-empty -m init`; `git -C "$F/repo" remote add origin git@github.com:o/r.git`; `git -C "$F/repo" update-ref refs/remotes/origin/main HEAD`; `git -C "$F/repo" symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/main`; `mkdir -p "$F/repo/_support/openspec"`; `cd "$F/repo"`. Así las cuatro sugerencias obligatorias aparecen.
4. `"$SCRATCH/hive" tui --home "$F" --state-dir "$F/state"`.
5. Models → un rol de Claude → Enter → esfuerzo `max` → Enter → Apply. Se espera ver la fila con `*` y `max`.
6. `x` sobre el mismo rol → Apply. Se espera ver la fila con el valor de la release y sin `*`.
7. Project → `e` → revisar las sugerencias → Enter → Apply. Se espera ver «Valid», y `AGENTS.md` en `$F/repo` con la sección.

La implementación deja los comandos exactos, ya probados, en el reporte de entrega.

8. **Opcional, sobre tu instalación real (AC12 en vivo):** `hive tui` sin `--home` → Models → Enter sobre un rol → → para abrir la lista. Vas a ver los modelos que lista cada CLI. Sal con Esc, Esc y Esc, sin aplicar: solo se ejecutan los comandos de listado y no se escribe nada. Las pruebas automáticas no observan las salidas reales de los CLIs, porque usan `--home` o binarios falsos.

## Revisión del plan

**Ronda 1** sobre la revisión `cb590a97cc35` (hash del contenido de la carpeta), con tres `hive-review-plan` de solo lectura despachados de forma nativa en Claude Code. Los tres hallazgos bloqueantes de cada revisor se corrigieron en el plan.
- **Gestor y estado:** 6 hallazgos (3 bloqueantes), todos aceptados.
  - AC2 se reescribió para la vuelta atrás, que hoy deja `legacy`.
  - `prepareTransaction` marca `changed` y poda los mapas vacíos.
  - `BuildModelsPlan` rechaza un plan que haga algo más que reescribir agentes.
  - OpenCode quita el `#variant` antes de poner el nuevo.
  - La prueba de plan editado recalcula su ID.
  - La lista de esfuerzos sale del código, sin reglas por CLI que no tienen fuente.
- **Interfaz:** 8 hallazgos (1 bloqueante), todos aceptados.
  - Se fijaron el pie, el envío del panel, la barra de ayuda por modo, el error de una línea, el flujo de confirmar y cancelar, y las teclas en panel y formulario.
  - Se agregaron las pruebas de AC5 y AC9 que faltaban.
  - Se copió el aislamiento del cambio anterior, con el repositorio de prueba como directorio de arranque.
  - El tema claro queda declarado como no observable en vivo.
- **Seguridad de escritura:** 7 hallazgos (2 bloqueantes), todos aceptados.
  - Se agregaron las pruebas de cada rechazo de AC8.
  - `validatePlan` valida los ajustes de otros CLIs.
  - El modelo admite solo una lista cerrada de caracteres.
  - Los valores de `AGENTS.md` deben quedar iguales tras `sanitizeLine`.
  - El validador trabaja sobre texto y separa bloqueos de avisos.
  - El archivo nuevo se publica con `os.Link`.
  - La URL de `origin` se lee sin credenciales y con el host exacto.
  - Sin hallazgo: `Render` ya escribe el modelo entre comillas en los seis formatos, así que no puede meter claves nuevas.

**Ronda 2** (última, sobre T1–T6). Se reanudaron los mismos revisores de gestor e interfaz sobre las correcciones que no venían de su propia propuesta. Las correcciones de seguridad eran propuestas del revisor aplicadas tal cual, así que no se volvieron a revisar.
- **Gestor:** 5 hallazgos, todos aplicados tal como los propuso el revisor, sin otra ronda:
  - la regla entre CLIs de `validatePlan` solo vale para `install`, y `remove` y `voice` llevan ajustes vacíos;
  - un `Model` vacío no se valida, y un ajuste vacío se rechaza;
  - la regla de OpenCode la aplica `Resolve`;
  - el conjunto de CLIs sale de `validateHosts`;
  - `set` nunca copia el valor de la release dentro del ajuste.
- **Interfaz:** 4 hallazgos, todos aplicados tal como los propuso el revisor:
  - `reset --only model|effort` da su comando equivalente al borrado parcial del panel;
  - Backspace en campos que no son de texto no cierra la vista;
  - el recorrido crea `origin/HEAD` y `_support/openspec`;
  - se agregaron casos a T3 y T5.
- **Límite que queda:** no está verificado que Claude Code acepte `ultra` ni qué valores acepta `thinking` en Pi. T1 lo comprueba con Context7 antes de fijar la validación (paso 2), y si la documentación no lo dice, lo deja anotado.
- **Cobertura:** gestor y estado, interfaz de terminal, y seguridad de escritura. No hay dominio de infraestructura, porque el cambio no despliega nada salvo la reconstrucción local del binario.

**Ronda sobre la ampliación T7–T10** (revisión del 2026-10-01, dos `hive-review-plan` de solo lectura):
- **Seguridad y backend:** 7 hallazgos, todos aceptados.
  - Carpeta de trabajo neutra para no cargar configuración ni extensiones del repositorio.
  - Proceso aislado de la terminal, con el grupo de procesos terminado al cortar.
  - Desbordamiento de salida como error.
  - Las frases de «sin catálogo» se reescriben en la especificación y la propuesta.
  - Regla de OpenCode igual para la pantalla y el comando.
  - La restricción de `hive models` ahora solo protege las filas de los roles.
  - Prueba de cero ejecuciones.
- **Interfaz:** 10 hallazgos, todos aceptados.
  - Enter sigue revisando desde el panel, y → o una letra abren la lista.
  - «release default» va en la lista.
  - El esfuerzo se compara por valor.
  - Valores iniciales «mixed».
  - Caché compartido en `appModel`, con spinner.
  - Reparto de las 10 filas.
  - Pruebas de T3 que se adaptan.
  - T8 corre la suite completa de `tooling/cli`.
  - AC10 nombra las pantallas nuevas.
  - Confirmación de grupo con los reemplazos primero, y cabeceras con `*`.
- **Correcciones del orquestador:**
  - la regla de OpenCode (un modelo nuevo conserva el esfuerzo que muestra cada rol);
  - las teclas del selector.

  Las dos responden a decisiones que el revisor pidió y no a su propuesta literal, pero no cambian ninguna decisión del usuario. Por el límite de una ronda de re-revisión, quedan cubiertas por la verificación por tarea y por las revisiones finales, en lugar de otra ronda de plan.

## Progreso

- **2026-10-01, plan:** plan escrito.
- **2026-10-01, arranque del build:** autorizado por el usuario con `/flow-build`.
  - T1 y T4 lanzados en paralelo, cada uno a un `hive-build-backend`.
  - T4 entregado en `7115d33` y enviado a `hive-verify-task`. T5 arrancado en paralelo con el mismo implementador.
  - Desviaciones de T4 que reportó el implementador, pendientes del veredicto:
    - la comprobación de caracteres de los valores vive en `validateHiveRequest`, para no cambiar la salida de `hive doctor`;
    - una rama que no se puede comprobar bloquea, y solo una rama ausente avisa;
    - un opcional vacío en `--set` se rechaza y remite a `--unset`;
    - el aviso de `CLAUDE.md` dice «by default», según la documentación de Claude Code (memory, sección AGENTS.md, v2.1.277+).
  - **Veredicto de T4:** `hive-verify-task` dio AC6, AC7 y AC8 por cumplidos sobre `7115d33`.
    - Corrió `go test ./tooling/cli` en 46 s y probó el binario con `expect` en repositorios temporales.
    - El orquestador comprobó las líneas citadas (`project_write.go:396`, `:227-238`).
  - **Corrección pedida tras el veredicto:** un valor inseguro que ya estaba en una clave no tocada pasaba sin bloqueo, y la vista previa lo mostraba limpio. Se alinea con design.md comprobando todos los valores resultantes. La corrección va en su propio commit (`125292c`) y la cubren `hive-verify-change` y la suite final.
  - **T5 entregado en `75dd40a` y enviado a `hive-verify-task`:**
    - el código del formulario se escribió antes que sus pruebas, para tener pronto las capturas de la composición; el paso de rojo a verde se observó contra la vista anterior;
    - el formulario carga el archivo y las sugerencias en segundo plano, y un archivo inválido muestra su motivo sin abrir el formulario;
    - un valor de `Delivery` o `Hive guidance` distinto de los válidos aparece como tercera opción, para no perderlo al abrir el formulario.
  - **T1 entregado** en `a2ad7bc`, `2a3ca13`, `0365a18`, `087f328` y `92affb1`, y enviado a `hive-verify-task`. T2 arrancó en paralelo. Desviaciones que reportó el implementador:
    - **Alias:** `management.ModelOverride` es un alias de `agents.ModelOverride`.
    - **Solo ámbito `user`:** los ajustes valen solo ahí, como dice la propuesta; un plan de proyecto no los lleva.
    - **`PlanUnchanged`:** cuenta un cambio de ajustes como cambio (una línea en `migration.go`).
    - **`ultra` en Pi:** se niega. La documentación de pi-subagents vía Context7 y `agent-delivery.md:32` dan `off`–`max`, y un valor desconocido apaga el razonamiento sin avisar.
    - **`ultra` en Claude Code:** sigue aceptado sin verificar. La documentación (model-config, env-vars) lista `low`–`max` y no dice que el frontmatter rechace otro valor.
  - **Límites de T1:**
    - Un ajuste solo de esfuerzo en OpenCode falla en `Resolve` si una release nueva deja ese rol sin modelo; `models reset` lo resuelve.
    - El golden `render_content.golden` sigue el contenido real y se rompería en cada cambio de `content/`. Se retira en T2, porque el golden sintético sigue protegiendo el renderizador.
  - **Veredicto de T1:** `hive-verify-task` dio AC2, AC3 y AC4 por cumplidos en la capa del gestor sobre `92affb1`.
    - Lo comprobó en un checkout fijo, con la suite normal y `-race` (139 s), y con ocho alteraciones del código: cada una hace fallar al menos una prueba.
    - La migración heredada solo se probó de punta a punta en una copia desechable. La prueba se agrega con T2.
    - El orquestador comprobó `models.go:245-252` y `plan.go:391-393`.
  - **Veredicto de T5:** `hive-verify-task` dio AC9 y la parte de Project de AC10 por cumplidos.
    - Comprobó `125292c`, con la suite de `tooling/cli` en 53 s.
    - De 16 alteraciones del código, las pruebas detectaron 13. Las tres restantes son huecos de cobertura sin defecto:
      - nada exige `TextFocused()`;
      - nada prueba una fila de error larga;
      - nada impide enviar claves sin cambios.
    - Se piden pruebas para las dos primeras.
  - **T2 entregado** en `93ab370`, `282a379`, `74f8402`, `11d1e67` y `ae22e45`, y enviado a `hive-verify-task`. T3 arrancó en paralelo.
    - Agrega `management.StoredModelOverrides`, una consulta de solo lectura.
    - Mueve `runModels` a `models.go`.
    - Rechaza también `--effort ""`, que remite a `reset --only effort`.
    - Un `reset` de un rol que no conocen ni la release ni el estado da error.
    - La marca `*` vive en `modelCellsFor`, que comparte con la vista.
  - **Veredicto de T2:** `hive-verify-task` dio AC1, AC3 y AC4 por cumplidos en el nivel del comando sobre `ae22e45`.
    - Hizo un recorrido con el binario sobre una instalación de prueba de los seis CLIs: `set` cambió solo el archivo del rol y `state.json`, y `reset --all` devolvió los hashes de la instalación limpia.
    - Cinco rechazos de AC4 se probaron solo con el binario o en capas inferiores; el comando solo reenvía esos errores.
  - **Restricción «`Render` idéntico sin ajustes (20 roles × 6 CLIs)»:**
    - la cubrió el golden que se verificó en `92affb1`;
    - `agents.go` no cambió después (`git diff 92affb1 ae22e45` vacío), así que la evidencia sigue valiendo aunque T2 retiró ese golden;
    - el golden sintético sigue en la suite.
  - **Detalle de redacción pendiente para T3:** tras un `set` que solo cambia el estado, el comando imprime igualmente «Open new CLI sessions.».
  - **Línea de proyecto integrada** en `cbd03fe`.
  - **T3 entregado** en `cb05172`, `6f1a960` (redacción de T2) y `9c341a4`, y enviado a `hive-verify-task`. Decisiones del implementador:
    - la lista de esfuerzos del panel sale de `agents.ValidateOverride`, así que Pi no ofrece `ultra`;
    - el resultado final aparece en la fila de posición;
    - la barra de ayuda usa el separador ` • ` que ya existe, no el `·` que escribió el plan.
  - **Veredicto de T3:** `hive-verify-task` dio AC5 y la parte de Models de AC10 por cumplidos sobre `9c341a4`.
    - La suite de `tooling/cli` pasó en 52 s.
    - De 13 alteraciones del código, las pruebas detectaron 10. Las tres restantes no esconden defectos, solo pruebas que faltan, y se pidieron:
      - un plan sin cambios con una petición no vacía;
      - un error de 200 caracteres;
      - un modelo con espacios alrededor.
    - Pruebas de cobertura agregadas en `30c4d56`.
    - **Observación de UX:** en un rol `inherit` de OpenCode, elegir a mano «release default» en el esfuerzo pierde el `#max` de la release (la confirmación muestra `max → -`). Es coherente con el diseño, pero la etiqueta puede confundir. Queda para la revisión de UX.
  - **T6:** documentación en `77a1e50`. `rg "models set|project set"` devuelve líneas y `rg "never write models"` no devuelve nada. Los ejemplos coinciden con el texto de ayuda de `main.go`. El delta de la especificación no necesitó cambios.
  - **Integración:** la línea de modelos se integró en `da2ef07`. Hubo un conflicto solo en el texto de ayuda de `main.go`, y se resolvió conservando las dos líneas. Suite completa y `-race` en curso.
  - **Para T3:** un ajuste solo de modelo sobre un rol de OpenCode con perfil `inherit` descarta el `#max` de la release. El panel debe enviar también el esfuerzo efectivo cuando cambia el modelo en OpenCode, para que el resultado coincida con lo que el usuario ve. El equivalente en texto es `set --model x --effort max`.
  - **Hallazgo incidental, fuera del cambio:** en el worktree de modelos, `go test ./tooling/management` sin `-count=1` dejó al comando `go` al 100 % de CPU con unos 4 GB durante varios minutos, dos veces. Con `-count=1` corre en 18 s.
- **Datos persistentes:** el cambio solo agrega el campo opcional `model_overrides` a `state.json`.
  - Hoy ningún estado lo tiene: 0 registros afectados en la única instalación, la local del usuario.
  - No hay migración. Un binario anterior ignora el campo (ver «Compatibilidad» en design.md).
- **Modelos del implementador y del verificador:** son distintos, así que no hace falta otro modelo.
  - Implementador: `hive-build-backend`, con `sonnet` configurado y no observado (`~/.claude/agents/hive-build-backend.md`).
  - Verificador: `hive-verify-task`, con `opus` configurado y no observado (`~/.claude/agents/hive-verify-task.md`).
- **2026-10-01, ampliación T7–T10 (D6-A, D7-A, D8-A):** plan revisado y corregido.
  - T7 y T8 arrancaron en paralelo, cada uno a un `hive-build-backend` nuevo; T7 en el worktree `gh-46-catalog` y T8 en `gh-46-models` sobre `da2ef07`.
  - Suite completa del candidato `da2ef07` antes de la ampliación: `go test ./...` en 1 min 11 s y `go test -race ./...` en 3 min 33 s, ambas en verde.
  - **Recorrido del usuario interrumpido:** con `--home`, la vista CLIs no detecta CLIs. Se instalaron los seis en el home de prueba con `hive plan install` y `hive apply`.
  - **T7 entregado** en `441bb09` y enviado a `hive-verify-task`. Puntos que hay que conocer:
    - en Codex, `supported_reasoning_levels` es una lista de objetos `{effort, description}`;
    - OpenCode devolvió otra vez salida vacía, así que su muestra está escrita a mano con el formato documentado;
    - no se extrajo un ejecutor común con `--version`, porque difieren en la entrada estándar, la carpeta, el grupo de procesos, el desbordamiento y la salida de error;
    - matar el grupo de procesos depende de `Setsid`, que solo existe en Unix.
  - **T8 entregado** en `a01d21b` y `913bef4`, y enviado a `hive-verify-task` junto con la reverificación de T2. Desviaciones:
    - las filas sin cabeceras se comparan ordenadas con el golden de la base, porque el orden por grupo las mueve;
    - la cabecera marca «mixed» parte por parte y lleva `*` si algún rol está ajustado;
    - los mensajes de uso nombran `--group`;
    - «Replaces the own override of» lista solo los roles cuyo ajuste cambia;
    - `--group` reemplaza el ajuste propio y `--role` lo combina, así que son equivalentes desde un estado sin ajustes.
  - **T7 integrado** en la línea de modelos con `6d47d13`, después de revisar el ejecutor: `Setsid`, matar el grupo de procesos, carpeta `os.TempDir()`, entrada nula y desbordamiento como error. T9 arrancó con el implementador de T8.
  - **Veredicto de T7:** `hive-verify-task` dio por cumplidas las partes de AC12 y AC13 que tocan a T7, sobre `441bb09`.
    - Contrastó Codex y OpenCode con su salida real: OpenCode salió vacía una vez y con 312 líneas la otra.
    - De seis alteraciones, las pruebas detectaron cuatro.
    - **Hallazgos que se resuelven en T9:**
      - la guarda de «cero llamadas» no detectaría una conexión a través de `catalogRunnerFor` con `--home`;
      - «sin terminal» se prueba solo con `[ -t 0 ]`.
    - **Desviación aceptada del diseño:** el ejecutor es propio y no extiende el de `--version`, porque aquel mezcla la salida de error y recorta en vez de fallar.
  - **Veredicto de T8:** `hive-verify-task` dio AC14 y AC15 por cumplidos en el nivel del comando sobre `913bef4`, y AC1, AC3 y AC4 de T2 otra vez cumplidos.
    - Rompió el código de diez maneras distintas y las pruebas detectaron las diez.
    - Probó el binario sobre una instalación de los seis CLIs con el contenido real.
    - **Hallazgos:**
      - `hive --help` no nombra `--group`; se agrega en T10;
      - en un rol `inherit` de OpenCode, `set --model` seguido de `reset --only model` deja un ajuste `{effort: max}`. El archivo vuelve al de la release, pero la fila sigue con `*`. Es coherente con la regla de OpenCode y queda para la revisión de UX.
  - **T9 entregado** en `f196839`, `5f751d7` y `5a2ae38`, y enviado a `hive-verify-task`. Decisiones del implementador:
    - el puntero del caché vive en `appConfig.catalog`, creado una vez en `newAppModel`, y lo escribe `appModel.Update`;
    - el panel de grupo conserva una parte no tocada cuando todos los roles ya la tienen como ajuste propio;
    - el filtro siempre deja «Other…»;
    - las cabeceras se alinean con las columnas de la tabla;
    - la ayuda del panel agrega `type model`.
  - **Seguimientos de T7 resueltos en `f196839`:**
    - la guarda de cero llamadas observa `listHostModels` por un gancho;
    - el falso registra su terminal de control con `ps -o tty=`. Esa comprobación solo detecta un fallo cuando la suite corre desde una terminal.
  - **Veredicto de T9:** `hive-verify-task` dio por cumplidas en la vista AC10, AC12, AC13, AC14 y AC15 sobre `5a2ae38`.
    - De 18 alteraciones del código, las pruebas detectaron 12.
    - Bajo una terminal real, quitar `Setsid` hace fallar la prueba.
    - **Pedido tras el veredicto, en un commit pequeño:**
      - pruebas para un grupo «mixed» y para «release default» en las dos partes de un grupo, cuya confirmación debe leerse como un reset;
      - corregir «Replaces the own override of» para que nombre solo los roles que pierden una parte;
      - una prueba con un id largo en la lista.
    - **Diseño ajustado:** con el panel abierto, la lista de esfuerzos solo se recorta al elegir un modelo. Abrir el panel o recibir la lista no mueve el esfuerzo, lo que resuelve la contradicción entre dos frases de design.md.
  - **Retoques de T9** en `29f99aa`: la confirmación de reset de grupo; «Replaces» nombra solo los roles que pierden una parte; pruebas de grupo «mixed» y de id largo.
  - **Línea de modelos integrada** en `2ca0403`.
  - **T10** en el commit de documentación sobre `2ca0403`:
    - `rg "debug models|--group"` devuelve líneas y «No model catalog» ya no aparece;
    - `hive --help` nombra `--group`;
    - el delta de la especificación agrega la lista, los grupos y la excepción de ejecución a «Read-only diagnostics».
  - **Suite completa y `-race`** del candidato integrado, en curso.
  - **Ronda de corrección 1 (validación del usuario, 2026-10-01):**
    - **Problema 1:** las cabeceras de grupo no se distinguían de los roles.
    - **Problema 2:** la confirmación era ilegible, con rutas absolutas por rol, un prefijo repetido y `low → low`.
    - **Corrección:**
      - cabeceras resaltadas con `▾`, roles sangrados y una línea en blanco entre grupos;
      - confirmación como tabla de cambios, con la ruta abreviada una sola vez.
    - T9 se reabre y se verifica en la puerta del candidato aceptado.
  - **Ronda 1 aceptada por el usuario:** tabla agrupada y confirmación en `3a84ce4`, integradas en `85a0598`; documentación alineada en `83158ca`.
  - **Ronda de corrección 2 (D9-A, decidida por el usuario el 2026-10-01): el selector de modelos pasa a un recuadro «Select model».**
    - **Contenido:** búsqueda, la sección «In use on <CLI>» con `●` en el modelo actual, secciones por proveedor (OpenCode, Pi), nombres legibles donde el CLI los da (Codex `display_name`, Cursor), la fila resaltada, y «release default» y «Other…» en la línea final.
    - **Fuera de alcance, porque Hive no tiene esos datos:** «Recent», favoritos y «Free».
    - **Comprobado:** OpenCode 2.0.21 no tiene `--verbose` en `opencode models`, porque imprime la ayuda.
    - **Implementado** en `48960b4` y `580a969`: bajo cada proveedor el id va sin el prefijo y el valor guardado sigue siendo el id completo. Integrado en `bcf51b2` y documentado en `378debd`.
    - Las pruebas de esta ronda se escribieron junto con el rediseño, sin observar el rojo antes; queda anotado.
  - **Ronda de corrección 3, pedida por el usuario:**
    - sin «In use»;
    - lista plana en Claude, Codex, Grok y Cursor;
    - secciones por proveedor con nombres legibles en OpenCode y Pi;
    - un panel sin borde como el de la referencia;
    - `●` y el cursor sobre el modelo actual;
    - un indicador `▸` con la pista «→ choose» en el campo Model.
    - **Error observado por el usuario:** el recuadro era más ancho que la terminal y sus bordes derecho e inferior no se veían. Se agrega una prueba de regresión de ancho.
    - **Implementado** en `9661bdb` e integrado. El implementador no pudo reproducir el ancho excesivo: el recuadro medía justo el ancho de la vista. La hipótesis es que la terminal del usuario mide los caracteres de caja (`┌ ─ │`) como más anchos. El panel nuevo no los usa, y una prueba exige que ninguna pantalla los contenga. La causa queda sin confirmar.
  - **Ronda de corrección 4, pedida por el usuario:**
    - se revierte el panel sin borde con fondo, y vuelve el recuadro con borde y sin fondo;
    - se conserva el resto de la ronda 3;
    - el recuadro deja dos columnas de margen a la derecha, para que un carácter de ancho ambiguo dibujado como doble no haga que el borde se salga.
  - **Candidato aceptado por el usuario:** `60d201e` («perfecto, así me gusta»), después de cuatro rondas de corrección.
  - **Puerta del candidato aceptado, completa porque ninguna había corrido:** suite normal y `-race`, `hive-review-ux`, `hive-verify-change`, y `hive-verify-task` sobre las tareas reabiertas T2, T7, T8 y T9.
  - **Suite del candidato `60d201e`:** pasa. `go vet ./...` limpio, `go test ./...` en 1 min 22 s y `go test -race ./...` en 4 min 34 s. `go.mod` no cambió respecto de `d4c014d`, así que no hay dependencias nuevas.
  - **`hive-review-ux` sobre `60d201e`: no pasa.**
    - **H1 (High, introducido):** a 80 columnas, la tabla de la confirmación se rompe con ids largos de OpenCode y Pi.
    - **Medium, introducidos:**
      - M1: una búsqueda vacía deja aplicar «release default»;
      - M2: el cursor salta al llegar la lista;
      - M3: el error de validación se corta.
    - **Low:** L1–L8. Se corrigen L1, L2, L5 y L7. Quedan anotados sin corregir:
      - L3 y L4, que el diseño especifica;
      - L6, la lista sin posición;
      - L8, porque tras escribir, «Valid» ya no muestra el aviso de `CLAUDE.md`, ya que la validación no revisa `CLAUDE.md`.
    - **Preexistentes:** la ruta larga en la cabecera de Project.
    - **Correcciones** enviadas al implementador; después se vuelve a revisar con el mismo revisor.
  - **Reverificación sobre `60d201e`:** T2, T7 y T8 cumplen todos sus criterios.
    - T9 cumple AC12–AC15, pero no AC10 (H1, nombres de doble ancho).
    - **Pedido al implementador, en la misma ronda que las correcciones de UX:**
      - recorte por ancho visual;
      - devolver las cuatro pruebas borradas;
      - una prueba del cursor inicial sobre el modelo actual;
      - renombrar los restos de la ronda 2.
  - **Plan alineado con el diseño aceptado:**
    - design.md «Selector en el panel» y las confirmaciones describen la forma final;
    - AC5 pide ahora la carpeta común y no la ruta de cada archivo, como aceptó el usuario en la ronda 1.
  - **Correcciones de la revisión de UX** en `182121b`: H1, M1–M3, L1, L2, L5 y L7.
  - **Correcciones de la reverificación de T9** en `d5d325f`:
    - recorte por ancho visual;
    - las cuatro pruebas borradas, devueltas;
    - la prueba del cursor inicial;
    - el renombrado de los restos de la ronda 2.
  - Todo se integró en `9b63680`. En curso:
    - la nueva revisión de UX, acotada a sus hallazgos;
    - la segunda ronda de verificación de T9, solo AC10;
    - la suite completa.
  - **`hive-verify-change` sobre `60d201e`:** AC1–AC16 cumplidos.
    - Usó un fixture de seis CLIs, CLIs falsos con registro de `argv`, `cwd` y `tty`, y gemelos para comparar la interfaz con los comandos.
    - **Hallazgos:**
      - D1 (tabla rota a 80 columnas), D2 (`…` suelto) y D3 (menú) ya estaban corregidos en `9b63680`;
      - D4: `Base branch: bad branch` se acepta porque se valida solo la primera palabra;
      - D5: se imprime «Updated» al crear un archivo y no se sangra la advertencia partida.
    - D4 y D5 se corrigen ahora.
    - **Límites:**
      - el `StateHash` viejo se rechazó por otra regla;
      - la migración heredada no se provocó a propósito;
      - instalar con `--home` y abrir sin él da «legacy migration pending», un efecto del fixture.
  - **Nueva revisión de UX sobre `9b63680`: pasa.**
    - H1, M1–M3, L1, L2, L5 y L7 quedaron corregidos, y los nombres CJK y con emojis caben a 80×24.
    - **N1 (Low, introducido por la corrección de H1):** el recorte por la izquierda puede ocultar dónde difieren dos ids. Se corrige recortando por el medio, junto con D4 y D5.
  - **Verificación de T9, ronda 2, sobre `9b63680`:** AC10 cumplido.
    - Quedan cerrados los huecos de pruebas H2 y H3: las cuatro pruebas devueltas y el cursor inicial. 9 de 10 alteraciones hacen fallar alguna prueba.
    - **H6 (nuevo):** en un cambio solo de esfuerzo, un id largo no se recorta y la fila se parte a 80 columnas. Se corrige junto con D4, D5 y N1.
  - **Suite sobre `9b63680`:** pasa, la normal en 1 min 29 s y `-race` en 4 min 48 s.
  - **Correcciones finales** en `6f0ea55` (D4, D5, N1, H6), integradas en `e7afd4a`. Pruebas con rojo antes del arreglo.
    - Sin otra ronda de verificadores: ninguna es Blocker ni High, y la suite las cubre.
  - **Suite sobre `e7afd4a`:** pasa, la normal en 1 min 25 s y `-race` en 4 min 43 s.
  - **Parada para el recorrido final del usuario** sobre `e7afd4a`, con un home y un repositorio de prueba nuevos.
  - **Recorrido final aceptado por el usuario** («adelante») sobre `e7afd4a`.
  - **Entrega:**
    - push de `feat/gh-46-model-and-project-writes` y [PR #89](https://github.com/JhonHawk/tricell-hive-private/pull/89) a `development`;
    - antes del push se buscaron datos de cuenta y patrones de credenciales en las muestras grabadas, sin coincidencias. El repositorio no declara un escáner de secretos;
    - `/code-review` de Claude Code en curso (D5-A).
  - **`/code-review` (medio) del PR #89 sobre `e7afd4a`:** diez hallazgos.
    - **Se corrigen:**
      - (1) P1 confirmado: `editHiveSection` pierde el salto de línea final o deja un CR suelto cuando la sección vacía cierra el archivo;
      - (2) P2: la regla de OpenCode fijaba el esfuerzo de la release. Pasa a `agents.Resolve`, solo para el `#variant` incrustado;
      - (3) P2: el recorte cuenta la flecha por bytes;
      - (4) P2: el error de un ajuste no nombra el rol;
      - (5) P2: la validación de D4 ejecutaba `git` sin el ejecutor controlado.
    - **Solo se reportan (P3):** (6–10) código muerto, duplicación de reglas y lecturas repetidas.
    - La corrección de (1) cambia lógica, así que habrá una re-revisión del diff de las correcciones.
  - **Correcciones del `/code-review`** en `09dcf95`, integradas en `9299036`.
    - La documentación de la regla de OpenCode se ajustó en `3736f02`, y la especificación y design.md en la carpeta del cambio.
    - **Suite sobre `3736f02`:** pasa, la normal en 1 min 24 s y `-race` en 4 min 45 s.
    - Push hecho. La re-revisión de `/code-review`, acotada al diff `e7afd4a..3736f02`, está en curso.
  - **Re-revisión de `/code-review` sobre `e7afd4a..3736f02`:** las correcciones 1, 2 y 3 se sostienen. Quedan dos restos de severidad baja:
    - F1: `Base branch` acepta espacios Unicode;
    - F2: la pista de reset aparece también en errores de la release.
  - **Decisión del usuario D10-A:** corregirlos y fusionar sin otra revisión.
  - **Correcciones de F1 y F2** en `d00793c`, integradas en `4a06722`. La suite y `-race` pasan.
  - **Merge del [PR #89](https://github.com/JhonHawk/tricell-hive-private/pull/89)** en `development` como `c0cbedb` el 2026-10-01. No hay CI.
  - **Binario local:** `~/.local/bin/hive` se reconstruyó desde `c0cbedb`. El cambio no toca `content/`, así que `hive update` no tiene nada nuevo que desplegar.
  - **Cierre:**
    - delta aplicado a `specs/versioned-installation/spec.md`;
    - carpeta archivada el 2026-10-01;
    - #46 cerrado.

