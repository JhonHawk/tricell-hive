# Tasks

Base `development` en `a8b2632e`; rama `feat/hive-local-one-line-facts`, en el worktree `.claude/worktrees/hive-local-one-line-facts`. Todas las tareas corren en el hilo principal: T1 a T3 son texto de guía de pocos archivos, y T4 escribe archivos locales y memoria con la aprobación del usuario, que un hijo no puede pedir.

### T1 — La guía global define la entrada de una línea y su detalle en la memoria persistente

- [x] «Local instructions» define la entrada de una línea, el detalle con su puntero y su ciclo de vida, lo que no se registra y el tope de 50 líneas, y «Persistent memory» permite esa escritura y el proyecto que nombra un puntero

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En «Local instructions», el segundo punto pasa a decir: qué es una entrada (qué tiene, usa o necesita aquí, o una limitación de una herramienta con la solución que funcionó, en una línea de hasta 200 caracteres con su fecha, sin pasos, relato ni versión salvo la que el proyecto fija); cuándo hay detalle (la línea no puede contener el comando con sus opciones, o reutilizar el hecho toma dos o más pasos); dónde va (tema `local/<slug>` nombrado por el asunto, tras buscar uno existente, en el proyecto del archivo o en `hive-machine` para el archivo de máquina) con el puntero al final de la línea; que se lee solo cuando la tarea necesita el hecho y que el hilo principal lo pasa en el brief del hijo que lo necesite; qué pasa cuando el detalle falla o la línea se quita; la línea sola sin Engram; la lista ampliada de lo que no se registra; y que las cuentas de prueba siguen en `ACCOUNTS.md` mientras que la cuenta o identidad que usa una herramienta sí es un hecho del archivo. El ejemplo de ubicación deja de nombrar la versión de una herramienta. El tercer punto baja el tope a 50 líneas. En «Persistent memory»: la lista de momentos de escritura añade el detalle de un hecho local, la frase «Never per task, finding, fix or discovery» nombra esa excepción, y la regla del nombre de proyecto admite el que nombra un puntero. No cambian dónde viven los archivos, quién los lee, la exclusión de Git, los scripts locales ni que los hijos no escriben. El tamaño de `global.md` se anota antes y después para el PR. El commit nombra las observaciones del contexto verificado (host, modelo o alcance, fechas) y el fallo observado.

**Verification:** `go test ./tests/content/... ./integrations/... ./tooling/management/...` pasa, incluidas las pruebas de tamaño de la guía, de su reparto por host y del tope por sección de Antigravity; lectura de los puntos cambiados contra AC1 a AC4.

### T2 — La skill `local-instructions` escribe líneas y parte las entradas largas al curar

- [x] La skill escribe entradas de una línea con su puntero cuando hay detalle, usa el tope de 50 y, al curar, propone para cada entrada existente si queda, se parte, se reubica o sale

**Closes:** AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/local-instructions/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** El paso 3 convierte cada candidato `commands` en una línea con su detalle cuando no cabe, y en modo curación clasifica también cada entrada existente más larga que su línea, de una clase que la guía excluye o que pertenece a otro de los archivos. El paso 4 lista esas propuestas con las demás y deja de decir que nada va a la memoria persistente. El paso 6 escribe la línea como la guía la define y guarda el detalle. El tope citado pasa a 50. La skill sigue apuntando a «Local instructions» como hogar de la regla y no repite sus pasos; su frase «Work only in the current project and the machine file» no cambia. `extract_candidates.py` no cambia.

**Verification:** `go test ./tests/content/...` pasa y la prueba `tests/skills/local_instructions_extract_test.py` pasa con el comando que el repositorio usa para ella; lectura de los pasos 3, 4 y 6 contra AC5.

### T3 — La documentación describe la entrada de una línea

- [x] Cada lugar de `content/`, del sitio, de la arquitectura y de la especificación que afirme qué guarda el archivo dice lo nuevo, y el changelog lo registra

**Closes:** AC7

**Depends on:** T1, T2

**Locations:** `CHANGELOG.md`; el propósito de `_support/openspec/specs/local-instructions/spec.md`; y solo donde afirmen lo que cambió: `llms.txt` (línea 46), `README.md`, `site/src/content/docs/` (`getting-started.md`, `host-limits.md`, `skills.md`, `flows/build.md`, `flows/verify.md`, `index.mdx`), `_support/docs/architecture/workspace-and-artifacts.md`, `content/skills/workspace-conventions/references/artifact-layout.md`, las referencias de `harness-audit`, las skills `flow-build`, `flow-verify`, `flow-research`, `flow-status` y `security-advisory`, y `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Changes:** Revisión de todo el comportamiento, como pide `AGENTS.md`: buscar `HIVE.local`, «Local instructions», `commands` y `start command` en `content/`, `site/src`, `_support/docs`, `_support/openspec/specs`, `llms.txt` y `README.md`, leer cada resultado, corregir el que describa una entrada como comandos, versiones o recetas, y dejar igual el que solo apunte a la regla. Lugares ya vistos que lo afirman: `site/src/content/docs/flows/build.md:30`, `flows/verify.md:19`, `content/skills/flow-build/SKILL.md:22` y el propósito de la especificación. Entradas en `[Unreleased]` del changelog. La lista de lugares revisados va al PR.

**Verification:** `go test ./...` pasa; `npm run build` en `site/` termina sin errores; la búsqueda anterior no deja una descripción de la entrada antigua fuera de los cambios archivados.

### T4 — Los ocho archivos de la máquina quedan limpios

- [ ] Cada `HIVE.local.md` tiene solo líneas de entrada de hasta 200 caracteres, sin entradas obsoletas ni de las clases excluidas, en 50 líneas o menos; el detalle de lo partido está en Engram y una copia literal de los archivos anteriores queda guardada

**Closes:** AC6

**Depends on:** T1, T2

**Locations:** fuera del repositorio: `~/.config/hive/HIVE.local.md` y los siete `HIVE.local.md` de proyecto de esta máquina; Engram de cada proyecto dueño y el proyecto `hive-machine`; la copia y la tabla de trabajo en `_support/workspace/2026-11-13-hive-local-cleanup/` del checkout principal (ignorada por Git)

**Execution:** main thread

**Test approach:** check

**Changes:** Con la regla ya integrada:

1. Copiar los ocho archivos, tal cual, a la carpeta de la limpieza, con modo `0600`.
2. Escribir ahí la tabla, una fila por entrada: su clase de `design.md` «Proposed shape» (queda, se parte, sale, se muda, se reubica), la línea propuesta, el tema de Engram cuando se parte, el destino cuando se reubica, y la marca de la que contiene un valor que debe sustituirse por el archivo que lo guarda. La entrada de archivo y librería de diseño de una app queda como línea hasta que su `DESIGN.md` los tenga. Las limitaciones de herramientas con su solución quedan o se parten, no salen.
3. Mostrar la tabla en la conversación y pedir una decisión por archivo: aprobar lo propuesto o conservar todo, con respuestas que nombren una entrada para cambiar su clase. Cada decisión dice que cubre reescribir ese archivo y guardar su detalle en el Engram de su proyecto.
4. Por cada archivo aprobado: guardar en Engram el detalle de cada entrada partida, bajo `local/<slug>` en el proyecto que nombra el `.engram/config.json` más cercano al archivo, o en `hive-machine` para el archivo de máquina, y cada instrucción del usuario como decisión; sin valores secretos; leer cada guardado de vuelta antes de seguir.
5. Leer el archivo de nuevo justo antes de escribirlo y reescribir solo las entradas de la tabla aprobada, conservando cualquier entrada que otra sesión haya añadido entretanto.
6. Avisar, sin editar nada de esos repositorios, qué archivo y librería de diseño conviene registrar en el `DESIGN.md` de su app, y que las sesiones abiertas con la guía anterior deben reiniciarse para no volver a escribir entradas largas.

No se toca ningún otro archivo de esos proyectos, ni `ACCOUNTS.md`, ni los scripts locales.

**Verification:** un comando que recorre los ocho archivos confirma que ninguna línea que empieza con `- ` pasa de 200 caracteres, que ninguna dice `STALE` y que ninguno pasa de 50 líneas; `engram search` en cada proyecto y en `hive-machine` devuelve cada tema `local/…` que la tabla nombra; una búsqueda de patrones de valor secreto en los detalles guardados no devuelve nada; `git status` en cada repositorio muestra que su archivo sigue sin versionar; la carpeta de la limpieza tiene las ocho copias.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y `npm run build` en `site/` | Covered by the implementation |
| Whole-behavior review | Final candidate, locally | Lista de lugares revisados y tamaño de `content/guidance/global.md` antes y después, para el PR | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |
| Cleanup table | T4, before any file is rewritten | La tabla mostrada y la respuesta del usuario por archivo | The user's answer to that round |

Sin interfaz que recorrer. Los criterios AC1 a AC5 y AC7 se observan en los archivos del repositorio; AC6, en los archivos de la máquina, en Engram y en la carpeta de la limpieza. El comportamiento de una sesión nueva al registrar un hecho no se prueba aquí, porque los pilotos de comportamiento están en pausa.

## Checkpoint

- Updated: 2026-11-13
- Stage: verify · 1
- Last commit: 8c6dfeaa
- Open findings: none
- Next: code review with /code-review of T1-T3, push, pull request, merge, local refresh; then T4: copies, table, the user's approval per file, Engram details, rewrite

### Evidence

- go test ./... · exit 0 · working tree that became 8c6dfeaa
- python3 -I tests/skills/local_instructions_extract_test.py · 6 tests OK · 1d2264a8
- npm run build in site/ · exit 0 · 8c6dfeaa
- content/guidance/global.md · 48751 -> 50690 bytes · 0c4dedf6
- CI: no job applies to a pull request to development

## Review status and progress

2026-11-13 — Plan escrito y revisado; ver el progreso de la propuesta.
2026-11-13 — T1 · 0c4dedf6 · unverified: none
2026-11-13 — T2 · 1d2264a8 · unverified: none
2026-11-13 — T3 · 8c6dfeaa · unverified: none
