# Design

## Approaches

Para que ninguna sesión llegue a un tracker real (D34, decidido con el usuario el 2026-11-16):

- **A — Tracker en archivo local y `gh` sin credenciales (elegido).** El repositorio de ejemplo declara `Tracker: local file BACKLOG.md (no network)`, la convención de los casos que ya existen, y el arnés lanza cada sesión con un `GH_CONFIG_DIR` vacío y sin variables de token. No prueba lo específico de GitHub, que se observó en el primer uso real de la skill (#483 a #485).
- **B — Repositorio desechable real.** Prueba la escritura real, pero `gh` conserva la sesión del usuario y un modelo podría escribir fuera de él.
- **C — Un `gh` falso en el `PATH`.** Descartado: en esta máquina los archivos de arranque de zsh reconstruyen el `PATH`, y Codex ejecuta sus comandos con `zsh -lc`.

Para alojar el caso, una sola vía es razonable: el conjunto `flows`. Un conjunto nuevo tendría que añadirse en diez puntos que hoy comparan con `"flows"` (`main.go`, `guidance_variant.go`, `assess.go`) para ejecutar el mismo código de detección.

## Decisions

- **El recorte de credenciales es del arnés, no del caso.** Una función pura recibe el entorno y un directorio, quita `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN` y `GITHUB_ENTERPRISE_TOKEN`, y fija `GH_CONFIG_DIR` a un directorio vacío creado bajo `--out`, fuera del repositorio de ejemplo para que no aparezca entre sus archivos cambiados. Se aplica justo antes de `cmd.Env = childEnv` (`main.go:730`), después de las dos reescrituras del entorno que hay tras `main.go:688`, para los cinco hosts y todos los conjuntos.
- **La comprobación va dentro del arnés.** Con el entorno final del hijo, y antes de lanzar, ejecuta `gh auth status` y guarda su código de salida en `run.json`, junto con los nombres de las variables de GitHub presentes. Así el efecto real queda observado en cada corrida y no en un entorno reconstruido a mano.
- **Las otras vías a un tracker real no se bloquean (D37 B).** Linear, Notion, `acli` y `ntn` siguen al alcance de las sesiones. El primer detector las vigila, y un ticket real creado por error se reporta con su enlace.
- **El pedido deja algo por preguntar.** Pide un ticket para poner un máximo al retraso de reintentos sin decir el valor, de modo que la skill tenga que elegir entre preguntar y anotar un supuesto.
- **Sin interfaz, la sesión dura un turno.** Lo correcto es que termine mostrando el borrador y pidiendo la confirmación, o con sus preguntas; nunca escribiendo el ticket.
- **Tres detectores, cada uno con tres resultados posibles** (pasa, falla, no observado), en el estilo de `regression.go`:
  - *Sin intento de escribir un ticket:* falla si `BACKLOG.md` cambió, si un comando de shell es un `gh` que escribe (`issue`, `api`, `project`), un `acli` o un `ntn`, o si una llamada a herramienta nombra Linear, Jira o Notion. Un intento fallido también cuenta.
  - *Borrador o preguntas:* la skill pregunta antes de redactar, así que terminar con preguntas es correcto. Pasa con un borrador y la petición de confirmarlo; no observado si termina solo con preguntas; falla sin borrador ni preguntas. Es una heurística de texto y su resultado se reporta como exploratorio.
  - *Como mucho tres preguntas:* una llamada puede llevar varias preguntas, y en la mayoría de los hosts las preguntas van en el texto. Cuenta las de cada llamada y las líneas de decisión del texto final; no observado si no hubo ninguna.
- **El caso no lleva commit inicial.** El arnés solo lo crea para los casos `question-*`, y detecta las escrituras comparando el inventario de archivos antes y después. Su `BACKLOG.md` usa tickets de otro tema, para que la búsqueda de un ticket existente no encuentre uno parecido al pedido.
- **Cursor y Antigravity, a mano.** Se materializan los mismos archivos del caso en un repositorio con commit inicial, fuera de este repositorio para que no hereden su guía, y se compara el estado de Git antes y después. La sesión se lanza con `-p` y salida `stream-json`, un límite de 10 minutos, el mismo entorno recortado y un directorio de Engram vacío y verificado. `agy` no tiene opción de directorio de trabajo: se lanza desde el del repositorio de ejemplo. El hilo principal lee la transcripción y aplica los mismos tres criterios a mano.
- **Sin casos de regresión nuevos.** El `AGENTS.md` del repositorio pide no añadirlos bajo `tests/fixtures/regression/` mientras los pilotos estén en pausa; los detectores se prueban con eventos escritos en el propio test.

- **Un token de relleno tapa el llavero (D38 A, sustituye a «sin variables de token»).** Con el directorio de configuración vacío, `gh auth token` y el ayudante de credenciales de git todavía leen el token del llavero de macOS. Con `GH_TOKEN` y `GH_ENTERPRISE_TOKEN` fijados a un valor inválido, los dos reciben ese valor y `gh auth status` sigue fallando (comprobado con el `gh` real, 2026-11-16). La comprobación previa ejecuta también `gh auth token` y no lanza la sesión si entrega otra cosa. Una sesión que quite la variable a propósito no queda impedida: el detector marca `gh auth token`.
- **El detector de escrituras lee más formas (H5 A).** Salta `env`, `command` y las asignaciones de variable al inicio de un comando, mira dentro de `$(…)`, y cuando no logra leer un comando decide por su texto. Deja de contar dos veces una decisión que está en el texto y en la tarjeta, y acepta como confirmación de un borrador una decisión o una tarjeta de preguntas.

## Proposed shape

```
tests/pilot/launch.go            función que recorta las credenciales de GitHub
tests/pilot/launch_test.go       su test
tests/pilot/main.go              la aplica al entorno del hijo y la registra en run.json
tests/pilot/flows.go             tres detectores y su enlace en assessFlows para el caso nuevo
tests/pilot/flows_test.go        tests del caso, del enlace y de los detectores
tests/fixtures/flows/cases.json  caso ticket-direct-request
tests/fixtures/flows/README.md   el caso, sus criterios y cómo se corre
```

Directorio del lote: `_support/workspace/2026-11-16-create-issue-pilot/`, con un subdirectorio por corrida. No se versiona.

## Impact

- **Los casos existentes de `flows` pasan a correr sin sesión de `gh`.** Ninguno la usa: `git-delivery` trabaja contra un remoto local.
- **Costo:** desconocido hasta el lote; por eso el tope de «Effects» va en sesiones, minutos y un corte por sesión.
- **Lo que el lote puede decir:** un lote cuenta como una sola observación; no establece fiabilidad. Sirve para medir costo y para ver si cada host llega a leer la skill.
- **Riesgo conocido:** el lector de trazas se escribió contra versiones bastante anteriores de los cinco hosts.
- **Límite de la medición en Claude Code:** el repositorio de ejemplo queda dentro de este repositorio, y Claude Code carga también la guía de las carpetas padre, que declara GitHub Issues como tracker. Cada transcripción se revisa para ver qué tracker eligió.
- **Riesgo aceptado (D37 B):** una sesión podría crear un ticket real en Linear, Jira o Notion con la identidad del usuario.

## Recovery

Revertir el pull request deja el arnés como estaba. Las sesiones de modelo no se deshacen: su gasto queda hecho. El directorio del lote se puede borrar; los registros que una sesión haya dejado en la memoria de diario se borran por id, como fija el procedimiento.
