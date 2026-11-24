# worktree-gate — Preguntar antes de crear un worktree y podar las reglas de worktree que generan ruido

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #479, 86cf5b1b |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `fb753487` · rama propuesta `feat/worktree-gate` |

## Decisiones

2026-11-15
- **Investigación A** — ¿Se añade una regla para que los comandos de una tarea se ejecuten desde el repositorio o worktree que contiene el cambio? Sí: ninguna regla lo dice hoy, y seis sesiones dejaron residuos de pnpm en la raíz de un workspace.
- **Investigación (gate)** — ¿Quién decide si se crea un worktree? El usuario: el agente lo pregunta en vez de decidirlo con criterios.
- **D1 A** — ¿Qué reglas de worktree se quitan? Se pregunta antes de crear un worktree, que va solo a la ruta nativa del host; se quitan los tres criterios de cuándo crearlo, la alternativa "donde diga la guía del repositorio" y el registro obligatorio de cada worktree en la memoria persistente (el worktree queda nombrado en el plan o el checkpoint de la tarea). Se mantienen: nunca forzar el borrado, rescatar los archivos ignorados, leer `HIVE.local.md` y `ACCOUNTS.md` del checkout principal, no tocar la rama de otra sesión y no usar `git stash` mientras otro agente trabaja en el repositorio.
- **D2 A** — ¿Qué memorias de worktree se borran? Solo las que tratan de un worktree, fuera de los proyectos `tricell-hive` y `hive-machine`: los registros de Engram cuyo título o tema es un worktree, y los pasajes sobre rutas o reglas de worktree en las memorias de los hosts, globales y por proyecto (Claude Code, Grok, Codex). Antes, una copia de respaldo; se salta todo registro de un worktree que siga existiendo. El criterio de qué registro se borra y cuál se edita lo fija D5 A.
- **D5 A** — Fija el criterio de selección que D2 A dejaba abierto: ahora es la ruta. ¿Qué se hace con los registros de Engram que listan una ruta de worktree (260 fuera de Hive)? Lo que ensucia el contexto es la ruta. Los 55 que tratan del worktree mismo se borran. En los 205 restantes (decisiones, resúmenes de sesión, hallazgos) se quita solo la ruta y se conserva el resto. El usuario valida la lista de ambos grupos antes de tocar nada. El mismo criterio de la ruta rige para las memorias de los hosts.
- **D6 A** — ¿Qué pasa con "or in a separate worktree" en los roles de subagente? Se quita: los hijos comparan con `git diff` o `git show` y nunca crean un worktree. Se actualizan el test y los roles.
- **D7 A** — ¿Qué hace el trabajo desatendido si otra sesión escribe en el checkout? Detenerse y reportarlo.
- **D8** — **Replaces:** D7 A. En una sesión desatendida, el agente decide en el momento si crea el worktree y elige lo que no lo detiene. Reporta la elección y su motivo.
- **Ampliación (corrección dentro de D2 A)** — Las memorias que cada host guarda por proyecto entran en la limpieza: la memoria de Claude Code de cada proyecto y los temas de cada workspace de Grok.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Y creo que lo convertiremos en un gate, o sea, que las reglas vayan sobre preguntar si se crea o no el worktree. Lo que quiero es quitar reglas de uso de worktree que generen ruido.» (2026-11-15)
>
> «Y además consideremos ELIMINAR todas las memorias de los hosts que hablen de worktrees (excepto las de hive).» · «y también las memorias de los hosts que tienen por proyecto» (2026-11-15)

Antes, en la misma sesión, el usuario aprobó (opción A) añadir la regla del directorio de trabajo de los comandos.

## Objective

Hoy la guía decide sola cuándo crear un worktree (tres criterios), deja ambigua su ubicación ("la función nativa del host; si no, donde diga la guía del repositorio") y obliga a guardar cada worktree en la memoria persistente. Ningún proyecto de org-a declara una ubicación y los registros de memoria se copian como precedente: aparecieron worktrees en `project-a/.worktrees`, `project-g/_support/workspace/<fecha>/project-g-monorepo` y `project-b/_support/worktrees`, y un worktree integrado (CAR-3) quedó olvidado fuera del repositorio. Ninguna regla dice desde dónde se ejecutan los comandos: seis sesiones ejecutaron pnpm en la raíz de un workspace y dejaron `package.json`, `pnpm-lock.yaml` y `node_modules/`.

Al terminar: el agente pregunta antes de crear un worktree y usa la ruta nativa de su host; los comandos de la tarea se ejecutan desde el repositorio o worktree del cambio, con ruta absoluta; ninguna regla pide registrar los worktrees en memoria; y las memorias que hoy sirven de precedente ya no están.

## Scope and acceptance

Dentro: `content/skills/git-workflow/SKILL.md`, `content/guidance/global.md`, `content/skills/flow-close/references/branch-cleanup.md`, `content/skills/flow-build/references/delegation.md`, el delta de `workflow-lifecycle`, las páginas de `site/` que describen esos flujos, `_support/docs/architecture/workspace-and-artifacts.md`, y la limpieza de memorias de D2 A.

Fuera: las reglas de protección que D1 A mantiene; los worktrees que hoy existan; la memoria de Antigravity (sin entradas de worktree, 2026-11-15); el archivo `archive/` y los `MEMORY.md` generados de Grok, que su documentación no permite editar; los proyectos `tricell-hive` y `hive-machine`.

- AC1 — Antes de crear un worktree, `git-workflow` pide preguntar al usuario como decisión, con la frase "user's decision whether", salvo dos casos: el trabajo ya enviado a build y autorizado con un worktree pedido por el usuario, o una sesión que ya trabaja en uno. Ya no contiene los tres criterios de cuándo crearlo · false today because `rg -F "user's decision whether" content/` no encuentra nada y el paso 3 contiene "Create a worktree only when…" · Observed in: el `git-workflow/SKILL.md` instalado en cada host tras `hive update`
- AC2 — La ubicación de un worktree es solo la "native worktree location" del host; si el host no tiene una, o no está verificada, se pregunta la ruta. Desaparece "where the repository's guidance places worktrees" · false today because `rg -F "native worktree location" content/` no encuentra nada y la frase vieja está en el paso 3 · Observed in: el `git-workflow/SKILL.md` instalado tras `hive update`
- AC3 — La guía global establece que un comando que actúa sobre un repositorio (gestor de paquetes, build, test, Git) recibe el repositorio o worktree del cambio por ruta absoluta. Lo hace mediante la "own directory option" de la herramienta (`git -C`, `pnpm --dir`, `go -C`) y, solo para una herramienta sin ella, `cd <ruta absoluta> && <comando>` en la misma llamada donde el host lo permite. Nunca se apoya en un `cd` de una llamada anterior, ni se ejecuta desde la raíz de un workspace declarado ni tras una "chain of relative" `cd ..` · false today because `rg -F "own directory option" content/` y `rg -F "chain of relative" content/` no encuentran nada, y seis sesiones ejecutaron pnpm en la raíz de un workspace · Observed in: la guía global instalada tras `hive update`
- AC4 — Ninguna regla pide escribir en la memoria persistente al crear o quitar un worktree. `tasks.py checkpoint --worktree` guarda la ruta y la rama del worktree, y las llamadas posteriores la conservan. Otra sesión lo encuentra con `git worktree list`, y el checkpoint lo confirma · false today because «Persistent memory» incluye "when you create or remove a worktree", `git-workflow` paso 3 pide "Record its path, branch and purpose in persistent memory" y `tasks.py` no tiene la opción `--worktree` · Observed in: la guía global, los skills y `tasks.py` instalados tras `hive update`
- AC5 — El sitio describe la pregunta antes de crear un worktree y la ruta nativa, y la documentación de arquitectura registra la decisión que reemplaza los criterios del 13 oct · false today because ninguna página de `site/` menciona esa pregunta y `workspace-and-artifacts.md` describe los criterios como vigentes · Observed in: el sitio construido localmente (`astro build`) y el archivo `_support/docs/architecture/workspace-and-artifacts.md`
- AC6 — Fuera de `tricell-hive` y `hive-machine`, ningún registro activo de Engram de la lista de T6 contiene una ruta de worktree. Hay tres excepciones: los registros que el usuario excluyó al validar, los de un worktree que sigue existiendo y los temas `local/*` a los que apunta un `HIVE.local.md`. Los registros que tratan del worktree mismo quedan borrados de forma recuperable. Los demás conservan su contenido sin la ruta. Existe un export de cada proyecto afectado como respaldo · false today because 260 registros activos contienen una ruta de worktree · Observed in: la base de Engram de esta máquina
- AC7 — Los pasajes sobre rutas o reglas de worktree aprobados en T6 ya no están en las memorias de Claude Code por proyecto ni en los temas de los workspaces de Grok (salvo el de `tricell-hive`), con respaldo de cada archivo editado, y existe una nota de Codex que pide quitar los suyos · false today because esos pasajes siguen en 3 archivos de Claude Code, en temas de Grok de project-g, project-a y project-b, y no existe la nota · Observed in: los almacenes de memoria de los hosts en esta máquina

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Base `development` en `fb753487` (2026-11-15). `git-workflow/SKILL.md` ocupa 15957 bytes, a 43 del límite de 16000 que comprueba `tests/content/skill_size_test.go`: el cambio debe reducirlo.
- La decisión D29 (2026-10-12) quitó una ubicación fija de Hive porque chocaba con la de cada host; los criterios de creación entraron tras una sesión de project-a con cuatro worktrees innecesarios (Claude Code, 66c195b6, 2026-10-13). Fuente: `_support/docs/architecture/workspace-and-artifacts.md` línea 54.
- Sesiones que ejecutaron pnpm en la raíz de un workspace: Grok `01a18466` grok-4.7-build-fast (project-a, 2026-10-28), Grok `01a186ad` grok-4.7-build-fast (project-g, 2026-10-29), Codex `01a1a96b` gpt-6.1-sol (project-g, 2026-11-04), Grok `01a1b29f` grok-4.7 (project-g, 2026-11-06), Grok `01a1cbd1` grok-4.7-build-fast (project-b, 2026-11-11), subagente de Claude Code `a06066a5` Sonnet 5.5 de la sesión `b14d8f28` (project-b, 2026-11-13). Registro: `engram:tricell-hive/watch/pnpm-at-workspace-root`.
- Engram 3.7.2: `engram delete <id>` borra de forma recuperable por defecto. Primer recuento, ya sustituido por el criterio de la ruta de D5 A (260 registros con ruta: unos 55 tratan del worktree mismo y unos 205 lo citan): registros activos con título o tema de worktree fuera de Hive, 68 (`org-a-project-g` 47, `org-a-project-a` 10, `org-a-project-b` 6, `org-b-project-f` 2, `engram` 2, `tricell-landing` 1); otros ~585 solo lo mencionan (consulta de solo lectura a `~/.engram/engram.db`, 2026-11-15).
- Memorias de los hosts con "worktree" (2026-11-15): Claude Code, 3 archivos de memoria por proyecto (orca-sns-frontend-web, org-a-whatsapp-hub, org-a-internal-apps); Grok, temas en `workspaces/project-g-893b42f2` (6), `project-a-bd41b901` (3), `project-b-63775406` (3), además de `archive/` y del workspace `tricell-hive-c8460796`, que quedan fuera; Codex, `memories/` (archivos y una base `memories_1.sqlite`); Antigravity, ninguna.
- Engram 3.7.2 tiene `export [file] --project` e `import <file>`, pero no tiene orden de restauración ni de recuperar un borrado; la restauración se prueba en T6 antes de borrar (`engram --help`, 2026-11-15). Candidatos por substring "worktree" en título o tema fuera de Hive: 69 (26 `decision`, 21 `config`, 18 `discovery`, 2 `bugfix`, 1 `pattern`, 1 `project`); varios son decisiones o estado de trabajo que solo mencionan un worktree, y `org-a-project-b local/worktree-env` es el detalle de una entrada de `project-b/HIVE.local.md` (revisión del plan, 2026-11-15).
- Rutas nativas de worktree por host (`_support/docs/architecture/workspace-and-artifacts.md` línea 54): Claude Code dentro del repositorio en `.claude/worktrees/`; Codex, Grok y Cursor en su propio directorio de usuario; Antigravity, Pi y OpenCode sin verificar, por lo que en ellos se pregunta la ruta.
- Rutas de edición documentadas (`harness-audit/references/memory.md`): en Claude Code se editan los Markdown; en Grok solo `topics/` y `observations/_inbox/`, nunca el `MEMORY.md` generado, y su consolidación puede volver a aprender un dato borrado; en Codex se añade una nota en `memories/extensions/ad_hoc/notes/` que pide el cambio.

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **D3 B** — Entrega `automatic`: revisión, push, PR y merge a `development` sin parada de validación. Tras integrar, el checkout de mantenedor actualiza la instalación local (`hive update`) antes de T6.
- **D4 A** — Revisión nativa de Claude Code `/code-review` sobre el candidato antes del push.
- El mensaje de commit nombra cada sesión observada (host, modelo, sesión, fecha) y el fallo, como fija «Measurement» del `AGENTS.md` raíz. Tras integrar, se revisa `engram:tricell-hive/watch/pnpm-at-workspace-root`.
- Si la sesión termina antes de T6–T8, el cierre las registra como trabajo pendiente.
- No autoriza: borrar ramas remotas ajenas a este trabajo, una release ni promover a `master`. Los borrados de memoria de T7–T8 los autoriza D2 A, y solo para los registros y pasajes de la lista de T6.

## Review status and progress

2026-11-15 — Plan escrito.
2026-11-15 — Primera revisión del plan: 14 hallazgos. Se aplicaron 11; tres pasaron a decisiones (D5, D6, D7).
2026-11-15 — Decisiones D5 A, D6 A y D8 (D8 reemplaza a D7 A). Segunda revisión: 12 hallazgos, aplicados todos.
