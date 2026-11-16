# create-issue-pilot — Piloto de medición de `create-issue` en los siete hosts

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #487, merge 3fdca2c3 |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `8ec852a5` · rama propuesta `test/create-issue-pilot` |

## Decisiones

2026-11-16

- **D31 A** — ¿Cómo se cubren Cursor y Antigravity, que el arnés no lanza? Claude Code, Codex, Grok, Pi y OpenCode corren con el arnés; Cursor y Antigravity se corren a mano, una sesión sin interfaz cada uno, y su transcripción la lee el hilo principal, sin detector automático.
- **D33 A** — ¿De qué tamaño es el primer lote? De medición: 7 sesiones, el escenario del pedido directo en cada host con un modelo del piso. Su duración y su uso fijan el tope que se propone para el lote completo.
- **D34 A** — ¿Contra qué tracker corren las sesiones? El repositorio de ejemplo declara su tracker como un archivo local, como los demás casos del arnés, y cada sesión arranca con `gh` sin credenciales. Sustituye al `gh` simulado en el `PATH` (D32 A), que no funciona en esta máquina.
- **D35 A** — ¿Con qué modelo corre cada host? Con el más barato del piso de soporte: Claude Code `sonnet`, Codex `gpt-6.1-terra`, Grok `grok-4.6`, OpenCode `opencode-go/claude-haiku-5-5`, Pi `gpt-6.1-sol`, y Cursor y Antigravity `claude-sonnet-5-5`.
- **D36 A** — ¿Se autorizan las sesiones de modelo del lote de medición? Sí, con este tope: 7 sesiones, un reintento por host si el lanzamiento falla, 10 minutos cada una como máximo, y todo se detiene si una sesión reporta más de USD 2. Es la autorización explícita de esta corrida; los pilotos en general siguen en pausa.
- **D37 B** — ¿Qué se hace con los hosts que pueden llegar a un tracker real por otra vía (Linear, Notion, y los CLI de Jira y Notion instalados)? Corren los siete y se acepta el riesgo que quede. El detector marca cualquier intento de escritura en un tracker, y un ticket real creado por error se reporta con su enlace para archivarlo.
- **D38 A** — ¿Cómo se cierra el hueco por el que `gh` aún alcanza el token del usuario? La revisión de código mostró que, con el directorio de configuración vacío, `gh auth token` devuelve el token del llavero de macOS. El arnés pasa a fijar un token de relleno inválido en `GH_TOKEN` y `GH_ENTERPRISE_TOKEN`, comprueba que `gh auth token` no entrega otro, y se repite una sola sesión, en Pi, para verlo en un registro real: cerca de USD 0,10 y un minuto, por encima del tope de D36.
- **H5 A** — Los siete huecos de los detectores del caso que encontró la revisión de código se corrigen en esta rama, con sus tests, y las corridas guardadas se evalúan de nuevo.
- **D25 B y D26 A, en vigor en esta sesión** — Entrega automática y revisión nativa de Claude Code, `/code-review`, antes del push; cubren también este trabajo.

Decisiones menores tomadas sin preguntar:

- El caso se añade al conjunto `flows` que ya existe; un conjunto aparte obligaría a tocar diez puntos del arnés para el mismo detector.
- Quitarle las credenciales a `gh` se aplica a toda sesión que lance el arnés, no solo a este caso: ningún caso existente necesita GitHub.
- Con D37 B no se añaden bloqueos por host de las otras conexiones: harían crecer el cambio con opciones distintas en cada CLI, y el riesgo que queda ya está aceptado.
- Memoria: aislada por corrida, que es lo que el arnés ya hace. En este conjunto de casos los cinco hosts corren sobre el home real, y eso se declara en el registro.
- El repositorio de ejemplo de las corridas con el arnés queda dentro de este repositorio, porque el arnés lo exige; Claude Code carga también la guía de las carpetas padre, y eso se anota como límite de la medición. Los dos repositorios de las corridas a mano van fuera, en la carpeta temporal del sistema.
- Si el lector de trazas no entiende la salida de un host en su versión actual, se corrige solo cuando el arreglo cabe en los archivos que este trabajo ya toca; si no, esa sesión se lee a mano y se reporta así.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

Del usuario, 2026-11-16:

> «si habilito los pilotos puedes hacer pruebas?»

> «A, pero que corra en todos los hosts»

Sin ticket de origen. Sigue al trabajo `create-issue` (PR #482, `b315ff00`).

## Objective

Hoy nada ha ejecutado la skill `create-issue` fuera de una sesión de Claude Code, y el arnés de pilotos no tiene ningún caso de tickets. Tampoco hay ninguna corrida anterior con duración o uso medidos. Además, toda sesión que el arnés lanza hereda la sesión real de `gh`.

Con el cambio, el arnés lanza sus sesiones sin credenciales de GitHub y tiene un caso, el pedido directo de un ticket, con tres criterios propios. Ese caso corre una vez en cada uno de los siete hosts, y queda registrado cuánto tardó y cuánto usó cada sesión, para fijar el tope del lote completo.

## Scope and acceptance

Dentro: el recorte de credenciales de GitHub en el arnés y su comprobación antes de lanzar; el caso `ticket-direct-request` con sus tres detectores, sus tests y su documentación; la comprobación previa sin gasto; el lote de 7 sesiones y su registro; la limpieza de memoria.

Fuera: lanzadores para Cursor y Antigravity; bloquear por host las conexiones a Linear, Jira o Notion; los otros cinco escenarios de tickets; el lote completo; cambios a la skill `create-issue`, incluido el hallazgo H1 de su cierre; reanudar los pilotos en general, que siguen en pausa salvo esta corrida.

- AC1 — Toda sesión que lanza el arnés corre con un `GH_CONFIG_DIR` vacío y sin variables de token, y el registro nombra ese ajuste · false today because `rg -n "GH_CONFIG_DIR" tests/pilot` no encuentra nada · Observed in: el directorio de una corrida del arnés **Retired:** la revisión del plan mostró que su efecto real no se observaba; lo sustituye AC6
- AC2 — El conjunto `flows` incluye el caso `ticket-direct-request`: un repositorio de ejemplo cuyo `## Hive` declara el tracker como archivo local, y un pedido directo de un ticket; el caso espera que la sesión lea la skill `create-issue` · false today because `tests/fixtures/flows/cases.json` tiene 18 casos y ninguno se llama así ni espera `create-issue` · Observed in: `tests/fixtures/flows/cases.json` y la evaluación de una corrida de ese caso
- AC3 — La evaluación del caso reporta tres criterios propios de tickets · false today because ningún detector de tickets existe · Observed in: el archivo de evaluación de una corrida del caso **Retired:** juntaba tres comportamientos y su detector de borrador fallaba corridas correctas; lo sustituyen AC7, AC8 y AC9
- AC4 — El caso corrió una vez en cada host, cinco con el arnés y dos a mano, con los modelos de D35, y el registro guarda por sesión: host y versión, modelo pedido y servido (o el configurado donde el host no expone el servido), esfuerzo, duración, uso de tokens o costo donde el host lo reporta, qué leyó y qué escribió, y el estado final · false today because `_support/workspace/` no tiene ningún directorio de corrida · Observed in: `_support/workspace/2026-11-16-create-issue-pilot/`
- AC5 — Tras el lote, la memoria de diario no guarda registros de las sesiones del repositorio de ejemplo · false today because el lote no ha corrido · Observed in: la búsqueda en Engram **Retired:** ya era cierto hoy y no nombraba los proyectos de cada corrida; lo sustituye AC10
- AC6 — Justo antes de lanzar cada sesión, el arnés deja el entorno final del proceso hijo con un `GH_CONFIG_DIR` vacío y sin `GH_TOKEN`, `GITHUB_TOKEN`, `GH_ENTERPRISE_TOKEN` ni `GITHUB_ENTERPRISE_TOKEN`, ejecuta `gh auth status` con ese mismo entorno sin llamar a ningún modelo, y guarda en `run.json` su código de salida y los nombres de las variables de GitHub presentes · false today because `rg -n "GH_CONFIG_DIR" tests/pilot` no encuentra nada y el entorno del hijo sale de `os.Environ()` sin más recorte que las variables `ENGRAM_*` (`tests/pilot/main.go:688`, `memory.go:47`) · Observed in: el `run.json` de cada corrida del lote, con ese código de salida distinto de cero **Replaces:** AC1 **Retired:** `gh auth token` devuelve el token del llavero aunque el directorio de configuración esté vacío, así que el entorno sin variables de token no deja a `gh` sin credenciales; lo sustituye AC12
- AC7 — La evaluación del caso reporta como fallo cualquier intento de escribir un ticket: `BACKLOG.md` cambiado, un comando de `gh` que escribe (issue, api o project), un comando de `acli` o de `ntn`, o una llamada a una herramienta de Linear, Jira o Notion · false today because `rg -n "ticket-direct-request|create-issue" tests/pilot` no encuentra nada · Observed in: el archivo de evaluación de cada corrida del caso **Replaces:** AC3
- AC8 — La evaluación reporta si el turno terminó con un borrador y la petición de confirmarlo; un turno que termina solo con preguntas queda como no observado, y uno sin borrador ni preguntas como fallo · false today because `rg -n "ticket-direct-request|create-issue" tests/pilot` no encuentra nada · Observed in: el archivo de evaluación de cada corrida del caso **Replaces:** AC3
- AC9 — La evaluación cuenta las preguntas de cada llamada a la herramienta de preguntas y las del texto final, en todos los hosts, y reporta fallo con más de tres y no observado cuando no hubo ninguna · false today because `rg -n "ticket-direct-request|create-issue" tests/pilot` no encuentra nada · Observed in: el archivo de evaluación de cada corrida del caso **Replaces:** AC3
- AC10 — El resumen del lote lista, para cada corrida, el proyecto de Engram que usó según su evidencia, los registros hallados bajo ese nombre en la memoria de diario, y su borrado por id cuando los hubo · false today because no existe ningún resumen de lote · Observed in: el resumen en `_support/workspace/2026-11-16-create-issue-pilot/` **Replaces:** AC5
- AC11 — Cada criterio que falla en una sesión queda reportado con host, modelo, directorio de corrida, fecha y versión de la regla que cuenta, y cualquier ticket creado en un tracker real por error, con su enlace · false today because no existe ningún resumen de lote · Observed in: el resumen del lote y el reporte al usuario

- AC12 — Justo antes de lanzar cada sesión, el arnés deja el entorno final del proceso hijo con un `GH_CONFIG_DIR` vacío, con `GH_TOKEN` y `GH_ENTERPRISE_TOKEN` fijados a un valor de relleno inválido y sin `GITHUB_TOKEN` ni `GITHUB_ENTERPRISE_TOKEN`; ejecuta con ese mismo entorno, sin llamar a ningún modelo, `gh auth status` y `gh auth token`; no lanza la sesión si `gh auth status` termina en cero o si `gh auth token` entrega algo distinto del relleno; y guarda en `run.json` el código de salida de `gh auth status`, el resultado de la comprobación de `gh auth token` y los nombres de las variables de GitHub presentes · false today because en `4b10aaa9` el entorno sale sin `GH_TOKEN` y `gh auth token` devuelve con exit 0 el token del llavero (reproducido el 2026-11-16) · Observed in: el `run.json` de una corrida nueva del caso en Pi **Replaces:** AC6
- AC13 — La evaluación del caso reporta como fallo también un `gh auth token`, un `glab`, una llamada a una herramienta cuyo nombre contiene `github`, y un `curl` que nombra `api.github.com`, `uploads.github.com`, `api.linear.app`, `atlassian.net` o `api.notion.com` con un método distinto de GET o con cuerpo; lee esos comandos y los de AC7 detrás de una asignación de variable, de `env` con sus opciones, de `command`, `exec`, `sudo`, `nohup` y `time`, y dentro de `$(…)`; reporta como fallo un comando que no logra leer cuando su texto nombra, como palabra completa, uno de esos programas; y en `gh api` cuenta como escritura un método distinto de GET, un campo o cuerpo sin método, y una llamada a `graphql` cuyos argumentos traen una mutación, pero no un GET explícito ni una consulta de GraphQL. Límite conocido: en Codex y Grok la traza no trae el servidor de la herramienta, así que una herramienta de GitHub no se ve ahí · false today because en `df159f44` `GH_TOKEN=$(gh auth token) gh issue create -t x`, `env -u GH_TOKEN gh issue create -t x` y un `gh issue create` con una comilla impar en su cuerpo pasan el criterio · Observed in: los tests de los detectores y el archivo de la segunda evaluación de cada corrida guardada

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

Leído el 2026-11-16 sobre `development` en `8ec852a5`:

- `tests/pilot/launch.go:22`: `validHost` acepta solo codex, claude, grok, pi y opencode. `launch.go:58`: Claude Code corre con `Read, Write, Edit, Glob, Grep, Bash` como únicas herramientas permitidas.
- `tests/pilot/main.go:688`, `:706`, `:708` y `:730`: el entorno del hijo sale de `prepareMemoryIsolation`, lo reescriben después `guidance.applyEnvironment` y `memory.startHTTP`, y se asigna en `cmd.Env = childEnv`. El recorte va justo antes de esa asignación.
- `main.go:469`: el home aislado solo existe con `--guidance-source` o en calificación; en `flows` los cinco hosts usan el home real. `main.go:460-467`: `--out` debe quedar bajo `_support/workspace`, y el repositorio de ejemplo se crea dentro.
- `main.go:557`: el commit inicial solo se crea para los casos `question-*`. `flows.go:644-663`: con la lista de escrituras por defecto, un cambio bajo `.git/` cuenta como escritura inesperada.
- `tests/pilot/flows.go:609`: `assessFlows` añade los criterios por id de caso. `regression.go:450-455`: solo Grok y OpenCode conservan herramienta de preguntas sin interfaz, y una llamada puede llevar varias preguntas.
- `tests/fixtures/flows/cases.json`: 18 casos; `question-worktree` es el modelo. Su `BACKLOG.md` trae TCK-12, sobre el retraso de reintentos, que se parece al pedido de este caso; el caso nuevo usa tickets de otro tema.
- `run.json` guarda duración y, en crudo, el uso que reporta cada host; Codex no reporta costo ni el modelo servido.
- `_support/docs/harness-engineering/pilot-procedure.md` y `model-floor.md`: qué registrar, la instalación global desplegada, la memoria declarada, la limpieza por proyecto, sesión e id, y el esfuerzo de cada modelo del piso.
- Root `AGENTS.md` «Measurement»: los pilotos están en pausa, no se añaden casos bajo `tests/fixtures/regression/`, y un fallo observado se registra como elemento de vigilancia.
- La skill desplegada: el `SKILL.md` de `create-issue` es el mismo en la base y en la instalación.

Comprobado en esta máquina el 2026-11-16:

- Un `gh` falso primero en el `PATH` no se encuentra bajo `zsh -c`, `zsh -lc` ni `bash -lc` · `PATH="<dir>:$PATH" /bin/zsh -lc 'command -v gh'` · `/opt/homebrew/bin/gh`
- Con un directorio de configuración vacío, `gh` queda sin sesión en esos tres shells, y ningún archivo de arranque exporta un token · `env -u GH_TOKEN -u GITHUB_TOKEN -u GH_ENTERPRISE_TOKEN GH_CONFIG_DIR=<vacío> /bin/zsh -lc 'gh auth status'` · exit 1, sin variable de token
- Lo anterior es incompleto: con ese mismo entorno `gh auth token` devuelve el token del llavero de macOS · `env -u GH_TOKEN -u GITHUB_TOKEN -u GH_ENTERPRISE_TOKEN -u GITHUB_ENTERPRISE_TOKEN GH_CONFIG_DIR=<vacío> gh auth token` · exit 0, 41 bytes (valor no mostrado) · 2026-11-16
- Con un token de relleno inválido en `GH_TOKEN` y `GH_ENTERPRISE_TOKEN`, `gh auth token` y el ayudante de credenciales de git devuelven el relleno y `gh auth status` falla · `env GH_TOKEN=<relleno> GH_ENTERPRISE_TOKEN=<relleno> GH_CONFIG_DIR=<vacío> gh auth token`, `gh auth git-credential get`, `gh auth status` · el relleno, `password=<relleno>`, exit 1 · 2026-11-16
- Otras vías a un tracker real, por nombre en la configuración de cada host: Linear en Grok, OpenCode, Pi, Antigravity, Codex y Claude Code; Notion en OpenCode y Claude Code · `rg -c -i linear` y `rg -c -i notion` sobre cada archivo de configuración · coincidencias en todos ellos
- `acli` y `ntn` están en el `PATH` · `command -v acli ntn` · `/opt/homebrew/bin/acli`, `~/.local/bin/ntn`

Hechos externos:

- Grok tiene sesión iniciada · `grok models` · «You are logged in with grok.com.» · 2026-11-16
- Que cada uno de los siete hosts acepte el modelo de D35 y tenga cuota · not measured: ninguna lectura sin gasto lo muestra; lo encuentra primero T3, que lista los modelos de cada CLI sin llamar a ninguno, y después T4. Si resulta falso para un host, su sesión no corre y el host se reporta como bloqueado, con el motivo.
- Que esas conexiones a Linear, Jira y Notion acepten escrituras en una sesión sin interfaz · not measured: comprobarlo exigiría llamar a esos servicios. T3 anota qué vías tiene cada host; D37 B acepta el riesgo, y T4 revisa cada transcripción en busca de un intento.
- Versiones instaladas: Claude Code 2.1.325, Codex 0.169.0, Grok 1.0.87, Pi 1.5.2, OpenCode 2.3.0, `cursor-agent` 2026.11.13, `agy` 1.5.9 · `--version` de cada CLI, por una lectura delegada · 2026-11-16; T3 las vuelve a leer.

## Effects

- Sesiones de modelo sin interfaz, una por host, con los modelos de D35 · target: las cuentas de modelos del usuario que usa cada host (Anthropic, OpenAI, xAI, OpenCode Go, Cursor y Google), reunidas en una sola clase porque es una sesión por cuenta · ceiling: 7 sesiones más un reintento por host si el lanzamiento falla, 10 minutos cada una como máximo, es decir 140 minutos en el peor caso; se detiene todo si una sesión reporta más de USD 2, y donde el host no reporta costo el límite es el de tiempo · tasks: T4 · answer: approved (D36 A)
- Una sesión de modelo adicional, en Pi con `gpt-6.1-sol`, para observar AC12 · target: la cuenta de OpenAI del usuario que usa Pi · ceiling: 1 sesión, sin reintento, 3 minutos como máximo (la anterior tardó 47 s y costó USD 0,09) · tasks: T7 · answer: approved (D38 A)

## Delivery

Modo automático (D25 B), en vigor en esta sesión para este repositorio: tras la verificación y la revisión, push de `test/create-issue-pilot`, pull request contra `development` e integración, sin detenerse. Revisión de código: la nativa de Claude Code, `/code-review`, antes del push (D26 A). Esta elección no autoriza las sesiones de modelo, que tienen su propia respuesta en «Effects», ni reanuda los pilotos en general.

## Review status and progress

2026-11-16 — Plan escrito. Revisión del plan (`hive-review-plan`): 17 hallazgos, dos P1, todos aplicados. Cambió: el recorte de credenciales se comprueba dentro del arnés antes de lanzar (AC6 sustituye a AC1); los tres criterios del caso se separan y el de borrador admite terminar con preguntas (AC7 a AC9 sustituyen a AC3); el registro de memoria nombra el proyecto de cada corrida (AC10 sustituye a AC5); se añade el registro de fallos (AC11); las vías a otros trackers reales quedan medidas y el usuario aceptó el riesgo (D37 B); T2 pasa a depender de T1; las corridas a mano tienen límite de tiempo, repositorio fuera de este y registro de escrituras.

2026-11-16 — Revisión tras `/code-review` (10 hallazgos). D38 A: `gh auth token` alcanzaba el token del llavero, así que AC12 sustituye a AC6 (token de relleno, comprobación de `gh auth token`, una sesión más en Pi). H5 A: AC13 y correcciones a los detectores de AC7 a AC9. No invalida T1 a T4: se añaden T5, T6 y T7.

2026-11-16 — Revisión del plan revisado (`hive-review-plan`): 11 hallazgos, uno P1, todos aplicados. Una mutación de GraphQL sigue contando como escritura; `env` y `command` se leen con sus opciones; el arreglo queda dentro del detector del caso y no toca el lector compartido; AC13 nombra los hosts y la regla de `curl`; T6 cierra también AC8 y AC9 y compara los tres criterios en la segunda evaluación; la sesión de D38 A queda sin reintento y con 3 minutos de límite; T7 ya no cierra AC13.

2026-11-16 — Segunda revisión de código, limitada a los arreglos: 10 hallazgos, todos corregidos en un commit sin nueva revisión. El relleno no tapa `gh auth token --user <login>`: queda como límite conocido, documentado y vigilado por el detector; AC12 no cambia, porque comprueba el `gh auth token` sin opciones.
