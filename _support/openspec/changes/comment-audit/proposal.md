# comment-audit — Skill global que audita comentarios, deja un lint gradual y genera la skill del proyecto

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`32bef8c8`, avanzada desde `8aa361cb` al integrarse `core-cleanup-batches`) · worktree `.claude/worktrees/comment-audit` · rama `feat/comment-audit` |
| Playbook | feature |

## Decisions

2026-11-22
- **D3 A** — ¿Cómo genera la skill el inventario y el lint gradual? Con una plantilla JS/TS incluida en la skill y el algoritmo descrito para los demás lenguajes, que el agente implementa con el parser propio de cada proyecto. El parser de la plantilla lo reemplaza D5.
- **D4 A** — ¿Se ajusta la regla `no-comments`? Sí: en un proyecto cuyo lint corre el chequeo de comentarios, un comentario Keep y una restricción que no se pudo convertir en check van al registro de decisiones del proyecto, con la ruta afectada, en lugar de quedarse en el código. Las directivas, los encabezados de licencia y la marca de atajo de D6 no se mueven. En los demás proyectos la regla no cambia.
- **D5 A** — ¿Qué parser usa la plantilla? `@babel/parser` con los plugins `typescript` y `jsx`, porque TypeScript 7.0.2 ya no exporta el parser clásico; la skill lo agrega como devDependency, tras el check de paquete, si el proyecto no lo tiene (8.x con Node ≥ 22.18 o ≥ 24.11; 7.29.x con Node anterior). **Replaces:** D3 (solo el parser de la plantilla).
- **D6 A** — ¿Qué pasa con la marca que `defect-fix` exige en cada atajo aceptado, en un proyecto con el lint? El lint permite exactamente una línea con el token de comentario de línea del lenguaje seguido de `workaround: <destino>` (en JS/TS, `// workaround: <ticket>`), sin prosa; la excusa se borra, se abre el issue y queda esa marca sobre el código del atajo. La forma vive solo en el leaf `no-comments`; `defect-fix`, `suppression-review` y la página de build del sitio la enlazan para los proyectos con el lint.
- **Supuesto** — Mientras el ticket no existe, o si el usuario no quiere uno, el destino de la marca es la entrada del registro de decisiones (`<ruta>#<ancla>`); una excusa sin código de atajo debajo se borra sin marca cuando su issue existe. Descartado: exigir el ticket antes del atajo, que obliga a escribir en el tracker durante un incidente, o dejar el atajo sin marca. Revertir: cambiar la frase de `defect-fix`.
- **Dónde A** — Worktree `.claude/worktrees/comment-audit` desde `development`, porque el checkout principal tiene sin commitear el trabajo de `core-cleanup-batches`.
- **Delivery B** — `automatic`.
- **Review A** — `Native: Claude Code /code-review`.
- **Security A** — También corre `hive-review-security`, por la plantilla que ejecuta `git` con rutas del repositorio.
- **Supuesto** — La plantilla se escribe como código nuevo en Hive a partir del algoritmo, no se copia el script del repositorio del cliente donde se probó, y ningún archivo versionado nombra ese proyecto. Descartado: copiarlo tal cual, porque ese código vive en un repositorio privado de un cliente. Revertir: reemplazar la plantilla.
- **Supuesto** — La skill generada se llama `delete-comment` por defecto, como en el prompt original; el usuario puede darle otro nombre al ejecutarla. Descartado: un nombre fijo. Revertir: cambiar el valor por defecto.
- **Supuesto** — La skill generada vive en `.agents/skills/<nombre>/`, sin `assets/hive-managed.json`, con un enlace relativo en `.claude/skills/<nombre>`, otro en `.grok/skills/<nombre>` solo si el proyecto ya tiene `.grok/`, y una línea en el `## Skills` del `AGENTS.md` del proyecto, repetida con prefijo en el workspace declarado si lo hay, como fija HA-IF-19. Descartado: copias por host, que se desincronizan. Revertir: cambiar `references/project-skill.md`.
- **Supuesto** — Los registros de decisión van al registro que el proyecto ya tenga; si no tiene ninguno, la skill pregunta y propone `<support home>/docs/decisions/` con su línea en el índice, tras leer el índice y la línea `Layout:` como fija `project-layout`. Descartado: ADRs sueltos por defecto. Revertir: cambiar la propuesta.
- **Supuesto** — Las excusas de una tanda se registran como un ticket padre con hijos, mediante el desglose de `create-issue`, con una sola aprobación. Descartado: un ticket suelto por excusa, que pide decenas de confirmaciones. Revertir: un ticket por excusa.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Ahi lo que quiero ver es si podemos convertir la skill en algo global» · «osea, creariamos una skill que es equivalente a mi prompt inicial en esa sesión no? la idea es que se cree su skill personalizada como ya funciono, y el lint y lo necesario, eso si, que lo haga para .agents y para .claude» (2026-11-22)

El prompt original, de la sesión de Codex `01a204d1-e082` del 2026-11-22: revisar cada comentario de código, clasificarlo como Tautología, Excusa, Obsoleto o Registro de decisión; borrar tautologías y obsoletos; borrar las excusas y abrir un issue para el atajo; mover los registros de decisión a un ADR; crear una regla de lint que prohíba comentarios con enfoque gradual, asumiendo que no todos se borran en ese PR; crear una skill `/delete-comment` en el repositorio para repetir el proceso; estimar cuántas líneas se ahorran borrando todos los comentarios.

## Objective

Hoy Hive revisa solo los comentarios que agrega o edita cada cambio (leaf `no-comments`). No tiene un procedimiento para auditar todos los comentarios de un repositorio, dejar un lint que impida comentarios nuevos ni generar la skill del proyecto para repetir la limpieza. Al terminar, la skill `comment-audit` hace eso en cualquier proyecto, y las reglas de revisión y de atajos dejan de pedir comentarios que el lint del proyecto rechaza.

## Scope and acceptance

Dentro: la skill `comment-audit` con su plantilla JS/TS y sus referencias; el ajuste de los leaves `no-comments`, `defect-fix` y `suppression-review` y de la página de build del sitio; el catálogo (`llms.txt`, el sitio), `instruction-resources.md` y `CHANGELOG.md`; un test Go que ejecuta la prueba de la plantilla con `@babel/parser` 8.1.2 y 7.29.14 fijados en `tests/content/testdata/comment-audit/`.

Fuera: correr la skill en un proyecto real (los pilotos están en pausa); plantillas para lenguajes que no sean JS/TS; cambios a `harness-audit` o a `skill_registry.py`; un enlace desde `hive-mode/references/code-audit.md`; las copias de ReUI en este repositorio (decisión del 2026-11-22: se quedan).

- AC1 — Instalar Hive desde este cambio entrega la skill `comment-audit`: `~/.agents/skills/comment-audit/SKILL.md` existe con su `name` y `description`, `~/.claude/skills/comment-audit` enlaza a ella, y `llms.txt`, `site/src/content/docs/skills.md` e `index.mdx` la listan · false today because ni `content/skills/comment-audit/` ni esas entradas existen en `8aa361cb` · Observed in: la instalación local tras el merge (`hive update --source <checkout>`, `ls -l ~/.claude/skills/comment-audit`) y los tres archivos del catálogo en `development`
- AC2 — La plantilla del lint gradual, copiada a un repositorio Git temporal, con `@babel/parser` 8.1.2 y también con 7.29.14: sale con 0 cuando un cambio no agrega comentarios o agrega solo una directiva permitida (como `/// <reference types="node" />`) o una marca `// workaround: <ticket>`; sale con 1 y escribe `archivo:línea:columna` por cada comentario nuevo, duplicado o reintroducido tras mover la base, también al renombrar un archivo, con nombres no ASCII y al correr desde una subcarpeta; parsea un `.ts` con un cast `<string>x` y una clase con decoradores; ignora texto con forma de comentario en strings, template literals, regex y texto JSX; salta las rutas excluidas; sale con 2 y un mensaje ante una base que no resuelve o sin `@babel/parser` instalado; `--inventory` imprime una línea JSON por comentario con `file`, `line`, `column`, `start`, `end`, `fullLine`, `directive` y `text`; y `--same-code --base <rev>` sale con 0 si los archivos cambiados solo perdieron comentarios, también al borrar `{/* x */}` en JSX, y con 1 si cambió otro token · false today because la plantilla no existe · Observed in: `go test ./tests/content -run TestCommentAuditTemplate -count=1 -v`
- AC3 — La estructura que prescribe la skill generada pasa `skill_registry.py check` · **Retired:** ya se cumple en la base con una estructura armada a mano; lo reemplaza AC6.
- AC4 — El leaf `no-comments` instalado dice que, en un proyecto cuyo lint corre el chequeo de comentarios, un Keep y una Constraint sin check van al registro de decisiones con la ruta afectada, nombra que eso reemplaza el «keep» de sus pasos 4 y 5 y el «revert» del paso 6 para la entrada del registro, excluye directivas, licencias y la marca de atajo, y conserva «Delete every Kill» y el comportamiento para los demás proyectos · false today because el leaf de `8aa361cb` siempre deja los Keep en el código · Observed in: `~/.agents/skills/hive-core/review/no-comments.md` y `~/.agents/skills/flow-verify/stage-entry.md` instalados tras el merge
- AC5 — Cada cláusula del prompt original tiene en `comment-audit` un paso que la cumple: las cuatro categorías con su acción y su correspondencia con Kill, Keep y Constraint, el issue de cada excusa creado antes de borrarla, el registro de decisión escrito antes de borrar su comentario, el lint gradual que asume una limpieza parcial, la skill del proyecto en `.agents` y `.claude`, y la estimación de líneas (líneas completas aparte de las mixtas, sin directivas) · false today because la skill no existe · Observed in: la tabla de cláusula a paso, con número de línea de `SKILL.md`, que la verificación escribe en `tasks.md` «Evidence» y que revisa `/code-review`
- AC6 — Un repositorio temporal construido siguiendo solo el texto de `comment-audit/references/project-skill.md` (sin pasos de este plan) hace que `skill_registry.py check` salga con 0 sobre el repositorio y sobre un workspace declarado que lo contiene · false today because ninguna referencia de Hive describe cómo registrar una skill generada · Observed in: el comando de la verificación de T2 · **Replaces:** AC3
- AC7 — El leaf `no-comments` instalado define, para un proyecto cuyo lint corre el chequeo de comentarios, la marca de un atajo aceptado: una línea sola con el token de comentario de línea del lenguaje seguido de `workaround: <ticket>`, o de la entrada del registro de decisiones mientras no hay ticket, y que la revisión confirma su destino; `defect-fix` y `suppression-review` instalados y la página `site/src/content/docs/flows/build.md` enlazan esa definición sin repetirla · **Replaces:** la versión anterior de AC7, que repetía la forma en cada archivo · false today because `defect-fix.md:8` pide una línea que nombra la causa en prosa · Observed in: `~/.agents/skills/hive-core/review/no-comments.md`, `build/defect-fix.md` y `review/suppression-review.md` instalados tras el merge, y la página en `development`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/test-audit/SKILL.md` (9.058 B) es el modelo de auditoría por tandas. Leído 2026-11-22.
- `content/skills/hive-core/review/no-comments.md` (2.413 B, límite de leaf 4.000 B) lo leen `flow-verify` (`assets/stage.json`) y el índice de `content/guidance/global.md:122`; `tests/content/single_home_test.go:92` exige «Delete every Kill»; `tests/content/leaf_links_test.go:95` exige el enlace desde `flow-verify`.
- `content/skills/hive-core/build/defect-fix.md:8` pide marcar el atajo «with one comment line that names the cause and, once it exists, the cause's ticket»; `review/suppression-review.md:7` pide conservar ese comentario; `site/src/content/docs/flows/build.md:28` lo describe. Leído 2026-11-22.
- `create-report/assets/starlight/*/astro.config.mjs`: precedente de plantillas de código en `assets/`; `tooling/management/plan.go` acepta `.mjs` en una skill (revisor del plan, 2026-11-22).
- `harness-audit/scripts/skill_registry.py check` reporta una skill en `.agents/skills` sin línea en `## Skills` (sale con 1) y acepta la skill con su línea y el enlace en `.claude/skills` (sale con 0); exime las skills con `assets/hive-managed.json`. Medido en un repositorio temporal, 2026-11-22. HA-IF-19 (`instruction-files.md:45`) exige repetir la línea con prefijo en un workspace declarado.
- `harness-audit/references/stack-skills.md`: `.agents/skills/<nombre>` lo leen Codex, Cursor, OpenCode, Pi y Antigravity; `.claude/skills` lleva un enlace; Grok lee `.grok/skills`. La documentación actual de OpenCode (opencode.ai/docs/skills, 2026-11-22) confirma que lee `.claude/skills` y `.agents/skills`.
- `npm view typescript`: `latest` 7.0.2, cuyo export `.` es solo `./lib/version.cjs` (2026-11-22).
- `@babel/parser@8.1.2`: publicado 2026-11-11 por GitHub Actions, 11 días, sin advisories, engines Node `^22.18.0 || >=24.11.0`, depende de `@babel/types ^8.1.2`. `@babel/parser@7.29.14`: 2026-11-11, GitHub Actions, sin advisories, Node `>=6`. Ambos con `check_package.py` exit 0 (2026-11-22). `@babel/types` 8.1.3 se publicó el 2026-11-20 (2 días): la instalación usa `npm install --before 2026-11-14` para no tomarlo.
- El catálogo de skills aparece en `llms.txt:53`, `site/src/content/docs/skills.md:46` e `index.mdx:106`; ningún test cuenta las skills.
- El repositorio no tiene CI que corra Go; los tests que necesitan Node lo saltan si falta (`tests/pilot/flows_test.go:108`). `tests/pilot/e2e_case.go:594` resuelve el checkout principal con `git rev-parse --path-format=absolute --git-common-dir`.
- Node v24.22.3. `core-cleanup-batches` se integró en `development` como `32bef8c8` (PR #595) durante este plan; el worktree avanzó por fast-forward a ese commit. En él, `defect-fix.md` mide 3.960 B (40 B libres bajo el límite de 4.000) y `suppression-review.md` 1.611 B; `no-comments.md` no cambió (2.413 B). `go test ./tests/content -count=1` en `32bef8c8`: ok (2026-11-22).

Comandos de verificación:

- `go test ./... -count=1` en el worktree · exit 0, 23 paquetes `ok`, 176 s · 2026-11-22
- `go vet ./...` en el worktree · exit 0 · 2026-11-22
- `go test ./tests/content -count=1` en el worktree · incluido en la corrida completa anterior, exit 0 · 2026-11-22
- `go test ./tests/content -run TestCommentAuditTemplate -count=1 -v` · not measured: el test y su `package.json` los agrega T1
- `python3 -I content/skills/harness-audit/scripts/skill_registry.py check <repo temporal>` · 0 con la línea de `## Skills`, 1 sin ella · 2026-11-22
- `python3 -I content/skills/harness-audit/scripts/skill_registry.py check <workspace temporal>` (tabla `## Repositories` con `app`, línea con prefijo `app/` en su `## Skills`) · 0 con la línea, 1 sin ella («## Skills lacks app/.agents/skills/delete-comment/SKILL.md») · 2026-11-22
- `rg -n 'no-comments|Delete every Kill|cause comment|workaround' content site _support/docs --glob '!_support/docs/history/**'` · 15 líneas en 13 archivos; las que tratan la marca de atajo son `defect-fix.md:8`, `suppression-review.md:7`, `site/.../flows/build.md:28` y `reuse-before-writing.md:6`, que ya acepta una marca que nombra su ticket y no cambia · 2026-11-22
- `npm run build` en `site/` con `node_modules` enlazado al checkout principal y quitado después · exit 0, 20 páginas · 2026-11-22
- `python3 -I <flow-plan-dir>/scripts/check_change.py _support/openspec/changes/comment-audit` · exit 0 · 2026-11-22

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic`: revisión, push, PR y merge sin parada de validación. Revisión de código: `Native: Claude Code /code-review`, más `hive-review-security`. Tras el merge, la actualización local de la instalación que `AGENTS.md` autoriza en el checkout del maintainer. No autoriza release ni cambios en otros repositorios.

## Review status and progress

2026-11-22 — Plan escrito.
2026-11-22 — Revisión del plan: dos `hive-review-plan` (servidor 195 s, 125k tokens; guía 176 s, 131k tokens). Revisión 1: D5 (parser Babel, porque TS 7 ya no exporta el parser) y D6 (marca `// workaround: <ticket>`) del usuario; AC2 ampliado; AC3 retirado y reemplazado por AC6; AC7 nuevo; T3 nombra qué pasos del leaf reemplaza; directivas, licencias y salidas de error en la plantilla; registro en workspaces; ruta de decisiones según `project-layout`.
2026-11-22 — Relectura de ambos revisores (servidor 88 s, 147k tokens; guía 55 s, 154k tokens). Revisión 2: plugins de Babel por extensión con `decorators-legacy` y `sourceType: "unambiguous"`; parser cargado con `import()` para salir con 2; `--same-code` tolera borrar comentarios en JSX; `ALLOWED` ampliado y extensible desde el inventario; la marca de atajo usa el comentario de línea de cada lenguaje, va solo sobre código de atajo y apunta al registro mientras no hay ticket; sin registro nombrado, el Keep se reporta abierto; el delta alineado con AC4 y AC6; medidos el build del sitio y el registro en workspace.
2026-11-22 — Segunda relectura (servidor 31 s, 157k tokens: sin P1 ni P2; guía 34 s, 165k tokens: `defect-fix` no tiene espacio). Revisión 3: la forma de la marca de atajo vive solo en `no-comments` y los demás la enlazan (AC7 reescrito); la revisión confirma el destino de la marca y la marca pasa a nombrar el ticket cuando se crea; caso de Babel ausente en un directorio sin `node_modules` y caso `<p>a {/* x */} b</p>`.
2026-11-22 — Tercera relectura (guía 21 s, 170k tokens; servidor 19 s, 164k tokens): sin P1. Corregidos: la cláusula de `defect-fix` reemplaza explícitamente la línea que nombra la causa; D6 dice que los tres archivos enlazan la forma; AC7 observa también `no-comments`. Plan listo.
2026-11-22 — `development` avanzó a `32bef8c8`; el worktree se adelantó por fast-forward y `defect-fix` quedó con 40 B libres: T3 recorta palabras de su paso 4 para que quepa la cláusula. `go test ./tests/content` pasa en la nueva base.
