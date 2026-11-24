# Tasks

Base: `development` en `fb753487` (2026-11-15), rama `feat/worktree-gate`.

Todas las tareas corren en el hilo principal: cada una edita entre 1 y 4 archivos ya identificados, y las de memoria (T6–T8) trabajan sobre almacenes locales de esta máquina.

T1–T5 forman la entrega del repositorio. T6–T8 corren después de integrar y desplegar, para que ninguna sesión nueva vuelva a escribir registros de worktree mientras se limpian. Si la sesión termina antes, el cierre las deja como trabajo pendiente.

Los comandos de este plan siguen la regla que añaden: cada uno indica el repositorio con la opción de directorio de la herramienta (`git -C`, `pnpm --dir`) o con rutas absolutas, y nunca depende de un `cd` hecho en una llamada anterior.

### T1 — `git-workflow`: preguntar antes de crear un worktree, solo ruta nativa, sin registro en memoria

- [x] El paso 3 pide preguntar antes de crear un worktree, fija la ruta nativa del host y nombra el worktree donde fija T3. Ya no tiene los tres criterios, la alternativa de ubicación ni el registro en memoria. Los pasos 2 y 3 no dan dos hogares a la misma pregunta.

**Closes:** AC1, AC2, AC4

**Depends on:** none

**Locations:** `content/skills/git-workflow/SKILL.md` (pasos 2 y 3)

**Execution:** main thread

**Test approach:** check

**Changes:**

Paso 3:
- El trabajo va en el checkout actual. Crear un worktree es una decisión del usuario y se pregunta como decisión, con lo observado como `Situación:`; por ejemplo, cambios ajenos en el checkout, otra sesión escribiendo o la necesidad de dos checkouts.
- No se pregunta en dos casos:
  - cuando el trabajo ya está enviado a build y autorizado y el usuario pidió el worktree; una pregunta como "¿podemos tomarlo en un worktree?" no cuenta;
  - cuando la sesión ya trabaja en uno.
- Con trabajo desatendido (D8, que reemplaza a D7 A), la pregunta se resuelve como fija el paso 5 de `unattended-delegation`: el agente elige la opción que no detiene el trabajo y la reporta. Aquí basta un puntero a ese paso; no se repite su regla.
- La pregunta contiene la frase "user's decision whether", y la ubicación, la frase "native worktree location", para que AC1 y AC2 tengan una comprobación positiva.
- Si se acepta, el worktree va en la ruta nativa del host; si el host no tiene, o no está verificada, se pregunta la ruta. Si la ruta queda dentro del repositorio, se añade a `.git/info/exclude`.
- El worktree se nombra donde fija T3 y en el reporte, y se quita como fija `branch-cleanup`.
- Se reescribe la cláusula de `direct-base`, que hoy "narrows step 2's rule to work in a worktree", para que no apunte a una regla que desaparece.

Paso 2:
- "do your own work in a worktree" pasa a la pregunta del paso 3.
- La lista de elecciones genéricas deja de incluir "or a worktree" y apunta al paso 3.

No cambia la carpeta de cambio que se mueve al worktree.

El archivo debe quedar por debajo de los 15957 bytes actuales.

**Verification:**
- `rg -n "Create a worktree only when|where the repository's guidance places worktrees|Record its path, branch and purpose in persistent memory|branch or a worktree" content/skills/git-workflow/SKILL.md` no imprime nada.
- `rg -F "user's decision whether" content/skills/git-workflow/SKILL.md` y `rg -F "native worktree location" content/skills/git-workflow/SKILL.md` imprimen una línea cada uno.
- `go test ./tests/content/...` pasa, incluido el límite de tamaño.
- `wc -c` del archivo es menor que 15957.

### T2 — Guía global y rúbrica: directorio de trabajo de los comandos y fin del registro de worktrees en memoria

- [x] `global.md` tiene la regla del directorio de trabajo, acotada; «Persistent memory» y la rúbrica de revisión de sesiones ya no cuentan crear o quitar un worktree como momento de escritura.

**Closes:** AC3, AC4

**Depends on:** none

**Locations:** `content/guidance/global.md` («Scope and authorization» o «Preservation», y «Persistent memory»), `_support/docs/harness-engineering/session-review-rubric.md` (fila «Memory»)

**Execution:** main thread

**Test approach:** check

**Changes:**
- **Nueva invariante en `global.md`.** Un comando que actúa sobre un repositorio (gestor de paquetes, build, test, Git) recibe el repositorio o worktree del cambio por ruta absoluta.
  - Formas: la "own directory option" de la herramienta (`git -C`, `pnpm --dir`, `go -C`); solo para una herramienta sin ella, `cd <ruta absoluta> && <comando>` en la misma llamada o, donde el host lo rechaza, rutas absolutas a los archivos que lee. Nunca argumentos absolutos que el gestor de paquetes lea como paquete.
  - Nunca se apoya en un `cd` de una llamada anterior: en un hijo, el directorio vuelve al de inicio entre llamadas.
  - Nunca desde la raíz de un workspace declarado ni tras una "chain of relative" `cd ..`.
  - Un cambio que abarca varios repositorios ejecuta cada comando en el suyo.
  - Las lecturas que recorren el workspace y las tareas operativas sin cambio en un repositorio no quedan cubiertas.
- **«Persistent memory».** Se quita "when you create or remove a worktree".
- **Rúbrica.** Se quita "a worktree change" de la fila «Memory».
- **Tamaño.** Se mide `global.md` antes y después, para el PR.

**Verification:**
- `rg -n "create or remove a worktree" content/guidance/global.md` y `rg -n "a worktree change" _support/docs/harness-engineering/session-review-rubric.md` no imprimen nada.
- `rg -F "own directory option" content/guidance/global.md` y `rg -F "chain of relative" content/guidance/global.md` imprimen la nueva regla.
- `go test ./tests/content/...` pasa.

### T3 — Checkpoint, cierre, delegación y roles: el worktree se encuentra sin registro en memoria

- [x] El checkpoint nombra el worktree. Si no hay checkpoint, se encuentra con `git worktree list`. El barrido de `branch-cleanup` tiene una fuente coherente. Los hijos y roles siguen lo que fije D6.

**Closes:** AC4

**Depends on:** T1

**Locations:** `content/skills/flow-plan/scripts/tasks.py` y `tests/skills/tasks_test.py` (opción `--worktree`), `content/skills/flow-build/SKILL.md` (lista del checkpoint), `content/skills/flow-close/references/branch-cleanup.md` (líneas 38 y 44), `content/skills/flow-build/references/delegation.md`, `tests/content/agent_boundary_test.go` y los roles de `content/agents/` (D6 A)

**Execution:** main thread

**Test approach:** tdd para `tasks.py`; check para el resto

**Changes:**
- **`tasks.py checkpoint`.** Gana `--worktree "<ruta> · <rama>"`, que escribe una línea `Worktree:` en el checkpoint. Las llamadas posteriores sin la opción la conservan, como hoy conservan `### Evidence`. El test lo cubre antes del cambio.
- **Checkpoint de `flow-build`.** Lista el worktree entre lo que guarda (`--worktree`). Otra sesión encuentra el worktree con `git worktree list`, y el checkpoint, que vive dentro de él cuando la carpeta de cambio se movió, solo lo confirma.
- **Barrido de `branch-cleanup.md`.**
  - Línea 38: la fuente de ramas del usuario en repositorios sin PR pasa a ser "a close record in persistent memory, or a plan or checkpoint naming the branch".
  - Línea 44: se conserva también toda rama que un plan o checkpoint nombra mientras su trabajo no está cerrado, para que el barrido no quite el worktree recién creado de otra sesión.
- **`delegation.md`.** Gana un puntero: lanzar un hijo aislado en su propio worktree es crear un worktree, y le aplica la pregunta de `git-workflow` paso 3.
- **Texto de frontera de los roles (D6 A).** Se quita ", or in a separate worktree" de la constante `boundaryGit` de `tests/content/agent_boundary_test.go` y de cada rol de `content/agents/` que la repite. Los hijos comparan solo con `git diff` o `git show <base>:<path>` y nunca crean un worktree. Es un cambio mecánico.
- **No cambian:** las condiciones para conservar un worktree, la regla de no forzar el borrado ni el rescate de archivos ignorados.

**Verification:**
- `rg -n "worktree record|in a separate worktree" content/` no imprime nada.
- `python3 -m unittest tests.skills.tasks_test` pasa con el caso nuevo de `--worktree`.
- `go test ./tests/content/...` pasa, incluido `agent_boundary_test.go`.
- `wc -c content/skills/flow-build/SKILL.md` es menor que 16000.

### T4 — Delta de spec, sitio y documentación de arquitectura

- [x] El delta de `workflow-lifecycle` refleja la pregunta y el directorio de trabajo; el sitio y `workspace-and-artifacts.md` los describen.

**Closes:** AC5

**Depends on:** T1, T2, T3

**Locations:** `_support/openspec/changes/worktree-gate/specs/workflow-lifecycle/spec.md` (ajustar a la redacción final de T1–T3; incluye el MODIFIED de «Close cleanup sweeps the repositories the work touched»), `site/src/content/docs/flows/build.md`, `site/src/content/docs/index.mdx` si su línea de `git-workflow` queda inexacta, `_support/docs/architecture/workspace-and-artifacts.md` (línea 54)

**Execution:** main thread

**Test approach:** check

**Changes:**
- **`build.md`** gana una frase: el agente pregunta antes de abrir un worktree y ejecuta los comandos desde el checkout del cambio. Enlaza al skill y no copia sus reglas.
- **`workspace-and-artifacts.md`** registra la decisión del 2026-11-15, que sustituye los criterios del 13 oct y el registro en memoria, con sus motivos (ubicaciones improvisadas, memorias que funcionan como precedente, residuos de pnpm). D29 se mantiene.
- **`README.md` y `llms.txt`** no cambian; se confirma.

**Verification:**
- `pnpm --dir site run build` y `pnpm --dir site run check` pasan.
- `rg -n "2026-11-15" _support/docs/architecture/workspace-and-artifacts.md` muestra la decisión.
- `python3 ~/.claude/skills/flow-plan/scripts/check_change.py _support/openspec/changes/worktree-gate` sale con 0.
- `rg -n -i worktree README.md llms.txt` no muestra nada que deje de ser cierto.

### T5 — Revisión completa de la conducta y suite

- [x] Cada lugar que dependía de los criterios, de la ubicación o del registro en memoria está actualizado o justificado, y la suite pasa.

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** T1, T2, T3, T4

**Locations:** búsqueda en `content/`, `site/`, `_support/docs/`, `_support/openspec/specs/`, `tests/pilot/`, `tests/fixtures/`

**Execution:** main thread

**Test approach:** check

**Changes:** Solo las correcciones que la búsqueda encuentre. Los casos `question-worktree` de `tests/fixtures/flows/cases.json` y `question-worktree-no-prep` de `tests/fixtures/qualification/suite.json` siguen siendo coherentes con la pregunta, porque una pregunta del usuario no autoriza crear el worktree. Se confirma y no se cambian. El PR lista cada lugar revisado.

**Verification:**
- `rg -n -i "worktree" content/ site/ _support/docs/ _support/openspec/specs/`, revisado línea a línea.
- `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan.

### T6 — Inventario, clasificación y respaldo de las memorias de worktree (tras desplegar)

- [x] Existe la lista de registros y pasajes con una ruta de worktree, dividida en "borrar" y "quitar la ruta" (D5 A) y validada por el usuario, con respaldo y una restauración probada.

**Closes:** AC6, AC7

**Depends on:** T5 integrada y desplegada con `hive update`

**Locations:** `_support/workspace/2026-11-15-worktree-gate/memory-backup/` (ignorado por Git); lectura de `~/.engram/engram.db`, `~/.claude/projects/*/memory/`, `~/.grok/memory-v2/workspaces/`, `~/.codex/memories/`

**Execution:** main thread

**Test approach:** check

**Changes:**

Engram:
- Candidatos: registros activos fuera de `tricell-hive` y `hive-machine` cuyo contenido lista una ruta de worktree: `/worktrees/`, `.worktrees/`, `_support/worktrees/`, o una carpeta de `_support/workspace/` usada como worktree. El 2026-11-15 eran 260.
- Se dividen en dos grupos:
  - **Borrar:** el registro trata del worktree mismo (su creación, ruta, configuración o retirada); unos 55.
  - **Quitar la ruta:** decisiones, resúmenes de sesión, hallazgos y demás; unos 205. Se elimina la frase o línea con la ruta y se conserva el resto. Si quitarla deja la frase sin sentido, se reescribe esa frase sin la ruta.
- Se mide por qué vía se edita un registro sin perder su id: `mem_update` con `find`/`replace` para el contenido, y sus campos `title` y `topic_key` cuando la ruta está ahí. Se prueba en una copia del almacén, nunca en un registro real.
- Siempre se conservan:
  - los temas `local/*` a los que apunta un `HIVE.local.md`, como `org-a-project-b local/worktree-env`;
  - los registros de un worktree que sigue en algún `git worktree list` de los repositorios de su proyecto. Cada proyecto de Engram corresponde a los repositorios de la carpeta de su workspace; por ejemplo, `org-a-project-g` corresponde a `org-a/project-g/*`.
- Antes de borrar:
  - se exporta cada proyecto afectado entero con `engram export <archivo> --project <proyecto>` al respaldo;
  - se prueba la restauración en una copia del almacén (el directorio de datos de Engram copiado y apuntado con su opción de directorio de datos, medida aquí). En la copia se hace `engram delete <id>`, después `engram import <export>`, y se comprueba que el registro vuelve a estar activo con el mismo id. Si no vuelve, se para y se pregunta otra vía antes de T7. Ningún registro real se toca antes de que el usuario valide la lista;
  - el respaldo queda con la carpeta en modo `0700` y los archivos en `0600`, porque contiene memorias completas de otros proyectos. Es evidencia única: el cierre pregunta antes de quitarlo.
- La lista, con id, proyecto, título, grupo y la ruta encontrada, se guarda en el respaldo. Se pide al usuario que la valide como decisión antes de T7; T7 aplica solo lo que el usuario valida.

Hosts:
- Los pasajes de Claude Code y de los temas de Grok que mencionan "worktree" se clasifican con el criterio de D5 A:
  - se borran las rutas de worktree y las reglas de uso que las fijan;
  - se conservan los datos sobre otro tema, como la columna de OpenCode o la limitación `isolation=worktree` de Grok.
- Quedan fuera el workspace de Grok `tricell-hive-c8460796`, `archive/` y los `MEMORY.md` generados.
- Cada archivo que se va a editar se copia al respaldo.
- Para Codex se listan los pasajes que se van a quitar, incluidos `rollout_summaries/` y `skills/`.

Antes de leer cada almacén se aplica la sección «Secrets» de la guía global, y no se copia ningún secreto a los reportes.

**Verification:**
- Cada registro de la lista tiene un motivo de borrado y figura en el export.
- La restauración de prueba devolvió el registro a activo.
- Cada archivo de host que se va a editar tiene su copia.

### T7 — Aplicar la lista validada en Engram

- [x] Los registros del grupo "borrar" quedan borrados de forma recuperable; los del grupo "quitar la ruta" conservan su id y su contenido sin la ruta.

**Closes:** AC6

**Depends on:** T6, con la lista validada por el usuario

**Locations:** Engram de esta máquina

**Execution:** main thread

**Test approach:** check

**Changes:**
- Grupo "borrar": `engram delete <id>`, sin `--hard`, para cada registro.
- Grupo "quitar la ruta": la vía de edición medida en T6, sobre cada registro.
- No se toca nada fuera de la lista.
- Para deshacerlo: `engram import` con el export de T6, y para las ediciones, la versión anterior de cada registro (tabla `observation_versions`) o el export.

**Verification:**
- `sqlite3 -readonly ~/.engram/engram.db "SELECT id FROM observations WHERE id IN (<grupo borrar>) AND deleted_at IS NULL"` no devuelve filas.
- La consulta de rutas de T6, repetida y limitada a los ids de la lista, solo devuelve los registros excluidos al validar. Fuera de la lista quedan, sin tocar, los temas `local/*` conservados y los registros que otras sesiones escribieron después de T6.
- Los registros editados siguen activos con el mismo id.

### T8 — Limpiar las memorias de los hosts

- [x] Los pasajes aprobados en T6 ya no están en Claude Code ni en Grok, y existe la nota de Codex que los pide quitar.

**Closes:** AC7

**Depends on:** T6

**Locations:** los archivos de `~/.claude/projects/*/memory/` y los temas de `~/.grok/memory-v2/workspaces/{project-g-893b42f2,project-a-bd41b901,project-b-63775406}/topics/` de la lista de T6, una nota nueva en `~/.codex/memories/extensions/ad_hoc/notes/`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **Claude Code y Grok:** se quitan solo los pasajes aprobados. Si un archivo o tema trata solo de eso, se borra entero; en Claude Code también se quita su línea de `MEMORY.md`.
- **Codex:** se escribe la nota `<timestamp>-remove-worktree-memories.md`, que enumera cada pasaje. Codex la consolida más tarde.
- **Reporte:** avisa de que Grok puede volver a aprender un dato borrado a partir de sesiones posteriores, y de que en Codex el cambio queda pendiente hasta que consolide la nota.

**Verification:**
- `rg -n -i worktree` sobre los archivos editados solo muestra pasajes clasificados para conservar.
- La nota de Codex existe y nombra cada pasaje.
- Cada archivo editado tiene su copia en el respaldo.

## Verification and human review

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Checks de cada tarea | Al cerrar cada tarea, en local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, en local | `go test ./...`; `pnpm --dir site run build`; `pnpm --dir site run check`; `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D4 A) | La respuesta de revisión |
| Despliegue local | Tras integrar en `development` | `hive update --source <checkout> --out <plan>` y `hive apply --plan <plan>`; sobre los skills instalados, `rg -F` de "user's decision whether", "native worktree location", "own directory option" y `--worktree` | Checkout de mantenedor: actualizar la instalación al integrar |

Entrega `automatic` (D3 B): sin parada de validación.

## Checkpoint

- Updated: 2026-11-15
- Stage: close · archive
- Last commit: 86cf5b1b
- Open findings: H27: children cannot run checks on base (accepted cost of D6 A, P3)
- Next: archive PR, delete feat/worktree-gate, remove scratchpad temporaries

### Evidence

- go test ./... · pass (4m17s) · 3f98f048
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 90 tests OK · 3f98f048
- pnpm --dir site run build · pass · 3f98f048
- pnpm --dir site run check · 0 errors · 3f98f048
- check_change.py worktree-gate · exit 0 · 3f98f048
- go test ./tests/content/... · pass · be7ab2ef
- python3 -m unittest tests.skills.tasks_test · OK · be7ab2ef
- pnpm --dir site run build · pass · be7ab2ef
- go test ./... · 22 ok · 82dec2ab
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 90 OK · 82dec2ab
- pnpm --dir site run build · pass · 82dec2ab
- pnpm --dir site run check · 0 errors · 82dec2ab
- engram import restore test on store copy · not restored · 86cf5b1b
- mem_update find/replace on throwaway 17640 · same id, version kept, hard-deleted after · 86cf5b1b
- installed copies AC1-AC4 rg · pass · 86cf5b1b

## Review status and progress

2026-11-15 — Plan escrito.
2026-11-15 — Revisión del plan (`hive-review-plan`, 14 hallazgos): se aplicaron H2, H3, H4, H6, H7, H9–H14; H1, H5 y H8 pasan a decisiones D5, D6 y D7.
2026-11-15 — Revisión: D5 A cambia AC6 y T6–T7; D6 A cambia T3; D8 reemplaza a D7 A en T1 y en el delta. Segunda revisión del plan (12 hallazgos, H15–H26), aplicados todos:
- `--worktree` en `tasks.py`;
- comandos que no dependen de un `cd` anterior;
- condición de conservación del barrido y su MODIFIED;
- comprobaciones positivas de AC1–AC4;
- excepciones `local/*` en AC6;
- prueba de restauración en una copia;
- respaldo protegido;
- recuentos y modelos.
No invalida tareas hechas (ninguna lo está).
2026-11-15 — T1 · cbb79dd4 · unverified: AC1/AC2 in installed copies until hive update
2026-11-15 — T2 · f7dfb692 · unverified: AC3 in installed copies until hive update
2026-11-15 — T3 · 891d71c9 · unverified: AC4 in installed copies until hive update
2026-11-15 — T4 · 3f98f048 · unverified: none
2026-11-15 — T5 · 3f98f048 · unverified: none
2026-11-15 — Verify: two /code-review rounds fixed (be7ab2ef and the next commit); T5's suite evidence at 3f98f048 is superseded by the rerun on the final candidate. AC1–AC4 stay unverified in installed copies until hive update.
2026-11-15 — T7 · uncommitted · unverified: none
2026-11-15 — T8 · uncommitted · unverified: Codex applies the note when it consolidates; Grok may relearn from later sessions
2026-11-15 — T6 · uncommitted · unverified: restore of a soft-delete only by SQL deleted_at=NULL; engram import does not restore
