# Tasks

Base: `JhonHawk/tricell-hive`, `development` en `8ec852a5`; rama propuesta `test/create-issue-pilot`. T1 corre en el hilo principal. T2 edita cuatro archivos del mismo paquete Go, así que espera a T1 y la corre un hijo `hive-build-backend`. T3 y T4 corren en el hilo principal: T4 gasta cuota y escribe el registro del lote. T5 a T7 salen de la revisión de código (D38 A y H5 A) y corren en el hilo principal.

### T1 — El arnés lanza sus sesiones sin credenciales de GitHub y lo comprueba

- [x] El entorno final de todo proceso hijo lleva un `GH_CONFIG_DIR` vacío y ninguna variable de token de GitHub, y `run.json` guarda el resultado de `gh auth status` bajo ese entorno

**Closes:** AC12

T1 dejó el entorno sin variables de token, el criterio anterior que se retiró; T5 lo cambia por el relleno y es la tarea que entrega AC12.

**Depends on:** none

**Locations:** `tests/pilot/launch.go`, `tests/pilot/launch_test.go`, `tests/pilot/main.go`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Una función pura en `launch.go` recibe el entorno y un directorio, quita `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN` y `GITHUB_ENTERPRISE_TOKEN` y fija `GH_CONFIG_DIR`. `main.go` crea ese directorio vacío bajo `--out`, fuera del repositorio de ejemplo, y aplica la función justo antes de `cmd.Env = childEnv`, después de las reescrituras de la guía y de la memoria. Con ese mismo entorno ejecuta `gh auth status`, sin llamar a ningún modelo, y guarda en `run.json` su código de salida y los nombres, nunca los valores, de las variables de GitHub presentes; cuando `gh` no está instalado lo anota así. No cambia el `PATH` ni el home de ningún host.

**Verification:** El test de la función falla antes del cambio y pasa después: con un entorno que trae las cuatro variables y otro `GH_CONFIG_DIR`, el resultado no trae ninguna de las cuatro y trae el directorio dado una sola vez. Un segundo test cubre la comprobación previa con un `gh` falso que termina con error, en un directorio temporal, como hace `memory_test.go` con `engram`. `go test ./tests/pilot/...` pasa. El efecto real lo observa T4 en el `run.json` de cada corrida.

### T2 — El caso del pedido directo y sus tres detectores

- [x] `flows` tiene el caso `ticket-direct-request`, y su evaluación reporta los tres criterios de tickets

**Closes:** AC2, AC7, AC8, AC9

**Depends on:** T1

**Locations:** `tests/fixtures/flows/cases.json`, `tests/pilot/flows.go`, `tests/pilot/flows_test.go`, `tests/fixtures/flows/README.md`

**Execution:** `hive-build-backend`

**Test approach:** check

**Changes:** El caso sigue la forma de `question-worktree`: un `AGENTS.md` con `## Hive` y `Tracker: local file BACKLOG.md (no network)`, un `BACKLOG.md` con dos tickets de otro tema, que no se parezcan al pedido, y un pedido directo de un ticket que deja sin decir un valor; espera `skill_read: create-issue`. No lleva commit inicial y su id no empieza por `question-`, para no heredar los criterios de esos casos. En `flows.go`, tres detectores con los resultados pasa, falla y no observado, enlazados en `assessFlows` solo para este caso:

- *Sin intento de escribir un ticket:* falla si `BACKLOG.md` cambió, si un comando de shell es un `gh` que escribe (`issue`, `api`, `project`), un `acli` o un `ntn`, o si una llamada a herramienta nombra Linear, Jira o Notion; un intento que falló también cuenta.
- *Borrador o preguntas:* pasa con un borrador y la petición de confirmarlo; no observado si el turno termina solo con preguntas; falla sin borrador ni preguntas.
- *Como mucho tres preguntas:* cuenta las preguntas dentro de cada llamada a la herramienta de preguntas y las líneas de decisión del texto final, en todos los hosts; no observado si no hubo ninguna.

El README describe el caso, sus criterios, que los dos últimos son heurísticas de texto de resultado exploratorio, y el comando para correrlo. No se añade nada bajo `tests/fixtures/regression/`.

**Verification:** `go test ./tests/pilot/...` pasa, incluidos los tests que recorren todos los casos de `flows`. Los tests nuevos cubren, con eventos escritos en el propio test: `BACKLOG.md` cambiado, `gh issue create`, `acli` y una llamada a una herramienta de Linear como fallos del primer detector; un texto final con borrador, uno solo con preguntas y uno sin ninguno de los dos para el segundo; cuatro preguntas en una sola llamada como fallo y ninguna pregunta como no observado para el tercero. La evaluación de una corrida real la observa T4.

### T3 — Comprobación previa, sin gasto

- [x] Queda escrito con qué versión, modelo, esfuerzo y comando corre cada uno de los siete hosts, y qué vías a un tracker real tiene cada uno

**Closes:** AC4

**Depends on:** T1, T2

**Locations:** `_support/workspace/2026-11-16-create-issue-pilot/` (no se versiona)

**Execution:** main thread

**Changes:** Construye el arnés con `go build -o _support/workspace/bin/pilot ./tests/pilot`. Lee de nuevo la versión de cada CLI, y la versión y los hashes de contenido de la instalación de Hive, que el procedimiento pide registrar. Lista los modelos que cada CLI ofrece, sin llamar a ninguno, para confirmar el nombre exacto del modelo de D35 en cada host, y toma el esfuerzo de cada uno de la tabla del piso; en OpenCode el esfuerzo va en el nombre del modelo. Anota, por nombre, las conexiones a Linear, Jira o Notion de cada host. Para Cursor y Antigravity lee la ayuda vigente y deja escritos los dos comandos: salida `stream-json`, su modo de permisos más restrictivo que aún deje leer archivos, un límite de 10 minutos (el propio de `agy`, y un proceso con tiempo máximo para `cursor-agent`), las variables de memoria que usa el arnés (`ENGRAM_DATA_DIR` vacío y verificado, sincronización en la nube desactivada) y el mismo recorte de `gh`. Un host cuyo modelo no aparece queda marcado como bloqueado, con el motivo, y no corre en T4.

**Verification:** El archivo de la comprobación nombra, para los siete hosts, versión, modelo, esfuerzo, comando y vías a un tracker, o el motivo del bloqueo. Ninguna sesión de modelo se inició.

### T4 — El lote de medición y su registro

- [x] El caso corrió una vez en cada host no bloqueado, y el registro y el reporte dicen cuánto tardó y usó cada sesión

**Closes:** AC2, AC4, AC7, AC8, AC9, AC10, AC11

**Depends on:** T3

**Locations:** `_support/workspace/2026-11-16-create-issue-pilot/` (no se versiona); la carpeta temporal del sistema para los dos repositorios de las corridas a mano; la memoria de diario, solo para comprobarla y limpiarla

**Execution:** main thread

**Changes:** Antes de la primera sesión, anota el destino, el cambio y que el gasto no se deshace. Corre el caso con el arnés en Claude Code, Codex, Grok, Pi y OpenCode, una sesión cada uno y de una en una, cada una en su propio directorio de salida, con los modelos y esfuerzos de T3; un reintento usa un directorio nuevo. Después las dos sesiones a mano: materializa los archivos del caso en un repositorio con commit inicial fuera de este repositorio, lanza la sesión y compara el estado de Git antes y después. Un lanzamiento que falla se reintenta una vez; una sesión que reporta más de USD 2 detiene el lote. Evalúa cada corrida y lee cada transcripción contra los tres criterios, anotando en las de Claude Code qué tracker eligió, porque carga también la guía de este repositorio. Si el lector de trazas no entiende la salida de un host, aplica la decisión menor del plan. Al terminar, toma de la evidencia de cada corrida el proyecto de Engram que usó, busca ese nombre en la memoria de diario y borra por id lo que encuentre. Escribe el resumen del lote. Cada criterio fallado se reporta con host, modelo, directorio, fecha y la versión del `SKILL.md` de `create-issue`, y se registra como elemento de vigilancia al cierre. Un ticket creado en un tracker real por error se reporta con su enlace y no se borra sin preguntar.

**Verification:** El directorio tiene un subdirectorio por corrida con su `run.json`, o con la transcripción y el estado de Git en las dos a mano. Cada `run.json` del arnés trae un código de salida distinto de cero para `gh auth status`. El resumen trae, por sesión, host y versión, modelo pedido y servido o configurado, esfuerzo, duración, uso o costo donde se reporta, lecturas, escrituras, estado final, el resultado de los tres criterios y el proyecto de Engram con lo hallado y lo borrado. El reporte al usuario da la duración y el uso por host, cada fallo en la forma indicada y el tope que propone para el lote completo.

### T5 — Un token de relleno tapa el llavero de `gh`

- [x] El entorno del hijo lleva un token de relleno inválido, y el arnés no lanza la sesión si `gh auth token` entrega otro

**Closes:** AC12

**Depends on:** T1

**Locations:** `tests/pilot/launch.go`, `tests/pilot/launch_test.go`, `tests/fixtures/flows/README.md`

**Execution:** main thread

**Test approach:** tdd

**Changes:** La función que recorta el entorno fija `GH_TOKEN` y `GH_ENTERPRISE_TOKEN` a un valor de relleno constante, que no es un secreto, y sigue quitando `GITHUB_TOKEN` y `GITHUB_ENTERPRISE_TOKEN`. La comprobación previa ejecuta además `gh auth token` con ese entorno, falla si la salida no es exactamente el relleno, y guarda el resultado en `run.json`. El README deja de decir que un `gh` no puede llegar a un tracker real y dice qué cubre el recorte. El test de la comprobación arma su entorno con una lista fija, no con el del proceso, para no depender de las variables `GH_` o `GITHUB_` de la máquina.

**Verification:** Los tests fallan antes del cambio y pasan después: el entorno resultante trae el relleno una sola vez en cada variable; un `gh` falso que entrega otro token hace fallar la comprobación; uno que entrega el relleno la deja pasar. `go test ./tests/pilot/...` pasa. Con el `gh` real y el entorno recortado, `gh auth token` devuelve el relleno. El efecto en una corrida lo observa T7.

### T6 — Los detectores del caso leen las formas que la revisión encontró

- [ ] La evaluación del caso marca los intentos que antes pasaban y deja de fallar turnos correctos

**Closes:** AC8, AC9, AC13

**Depends on:** T2

**Locations:** `tests/pilot/flows.go`, `tests/pilot/flows_test.go`, `tests/fixtures/flows/README.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En el detector de escrituras: saltar `env`, `command` y las asignaciones de variable al inicio de un comando y mirar dentro de `$(…)`; marcar `gh auth token`, `glab`, un `curl` que escribe en la API de un tracker y una herramienta cuyo nombre contiene `github`; decidir por el texto cuando el comando no se puede leer; leer el método y los campos de `gh api` también pegados a su opción, no contar como escritura un `gh api` con `-X GET` ni una consulta de GraphQL, y seguir contando una mutación de GraphQL. Todo dentro del detector del caso: `shellSegments`, en `regression.go`, no cambia. En el de borrador: aceptar como confirmación una decisión o una tarjeta de preguntas, quitar las frases que no piden confirmar, y conservar el texto anterior a una tarjeta rechazada. En el de preguntas: no contar dos veces una decisión que está en el texto y en la tarjeta. Las corridas guardadas del lote se evalúan de nuevo en un archivo de evaluación aparte.

**Verification:** `go test ./tests/pilot/...` pasa, con una fila nueva en los tests de cada detector por cada forma corregida, incluidas una consulta y una mutación de GraphQL y `env -u GH_TOKEN gh issue create`. La segunda evaluación de las cinco corridas del arnés se compara con la primera en los tres criterios, y cada cambio se explica en el resumen del lote y en su lista de fallos.

### T7 — Una sesión más, en Pi, para observar el recorte nuevo

- [ ] El caso corrió una vez en Pi con el arnés nuevo y su registro muestra el relleno

**Closes:** AC12

**Depends on:** T5, T6

**Locations:** `_support/workspace/2026-11-16-create-issue-pilot/pi-rev2/` (no se versiona)

**Execution:** main thread

**Changes:** Es el efecto que autoriza D38 A: una sola sesión del caso en Pi con `gpt-6.1-sol`, esfuerzo `medium` y límite de 3 minutos, sin reintento, con el arnés reconstruido. Antes de lanzarla se anota el destino, el cambio y que el gasto no se deshace. Después se evalúa, se lee la transcripción, se comprueba la memoria de diario por el proyecto de la corrida y se añade al resumen del lote.

**Verification:** El `run.json` de la corrida trae un código distinto de cero para `gh auth status`, la comprobación de `gh auth token` como pasada, y `GH_TOKEN`, `GH_ENTERPRISE_TOKEN` y `GH_CONFIG_DIR` entre las variables. El resumen del lote tiene la fila de esta sesión.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/pilot/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` and `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Preflight | Before the batch, locally | The preflight file names the seven hosts; no model session started | Covered by the implementation |
| Measuring batch | After the preflight, up to 140 minutes in the worst case | One run record per host and the summary | The effects answer (D36 A), with the accepted risk of D37 B |
| Extra session | After T5 and T6, under 3 minutes | The run record of `pi-rev2` | The effects answer (D38 A) |
| Code review | Final candidate, before the push | Claude Code `/code-review` on the candidate's diff | The review answer in force (D26 A) |

## Checkpoint

- Updated: 2026-11-16
- Stage: verify · ronda de hallazgos de /code-review, en espera del usuario
- Last commit: f9048e65
- Open findings: H4 P1 · gh auth token entrega el token del llavero con GH_CONFIG_DIR vacío (reproducido) · decisión D38 pendiente; H5 P2 · siete huecos de los detectores de tickets (cuenta doble, confirmación por decisión, gh tras env/asignación, comando ilegible pasa, GitHub por MCP/glab/curl, gh api mal clasificado, frases de confirmación amplias) · pendiente; H6 P3 · lista fija de herramientas de Linear porque trace.go pierde el servidor en Codex y Grok · se lista
- Next: Aplicar la respuesta a D38 (hueco del llavero de gh: relleno en GH_TOKEN, revisar AC6) y a H4 (siete huecos de los detectores); arreglar sin preguntar el test frágil de launch_test.go (os.Environ con variables GH_/GITHUB_ ajenas); después push, PR, merge y cierre

### Evidence

- go test ./tests/pilot/... · ok · caf1d155
- go vet ./... && go test ./... · ok · df159f44
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 90 OK · df159f44

## Review status and progress

2026-11-16 — Plan escrito y revisado; las tareas recogen los 17 hallazgos de la revisión del plan.
2026-11-16 — T1 · 4b10aaa9 · unverified: AC6: el efecto real (gh auth status distinto de cero en cada run.json) lo observa T4
2026-11-16 — T2 · caf1d155 · unverified: AC2, AC7, AC8, AC9: la evaluación de una corrida real la observa T4
2026-11-16 — T3 · uncommitted · unverified: AC4: la cuota de cada host no se mide; lo observa T4
2026-11-16 — T4 · df159f44 · unverified: none
2026-11-16 — Revisión tras `/code-review`: AC12 sustituye a AC6 y se añade AC13; T5, T6 y T7 nuevas; T1 a T4 no se reabren.
2026-11-16 — T5 · abd1dec3 · unverified: AC12: el run.json de una corrida real lo observa T7
