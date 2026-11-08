# Tasks

Base `development` en `b14ae746`; rama `feat/local-scripts-worktree-writes`. Todas las tareas corren en el hilo principal: son siete archivos de texto con ediciones cortas, y ninguna llega al umbral de delegación.

### T1 — Escritura por shell y scripts locales en la guía global

- [x] `content/guidance/global.md` lleva las ediciones de «Proposed shape» en `design.md`: las viñetas 1 y 2, la de autorización y la nueva de «Local instructions», el puntero en «Test accounts» y las excepciones en «Support folder» y «Proportionality».

**Closes:** AC9, AC12, AC13, AC17

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** solo esas frases. La viñeta de autorización sigue cubriendo solo los archivos `HIVE.local.md` en cualquier flujo; solo gana el caso del worktree. «Secrets» y `git-workflow` no cambian.

**Verification:**
- `rg -c "through a shell command" content/guidance/global.md` devuelve 4 (viñeta 1, viñeta de autorización, viñeta de scripts y «Test accounts»); hoy devuelve 0.
- `rg -o "_support/scripts/<name>|except the local scripts|local script «Local instructions» sets is not|git worktree list --porcelain|ends with a newline|running each Git command on its own|the .HIVE.local.md. files and creating" content/guidance/global.md` muestra siete coincidencias nuevas, una de cada frase; en la base no muestra ninguna.
- La viñeta de scripts va después de la que empieza por «This rule authorizes, in any flow».
- `go test ./tests/content/...` pasa.

### T2 — `scripts/` en el layout del hogar de apoyo

- [x] `artifact-layout.md` lista `scripts/` como destino nunca versionado que remite a «Local instructions».

**Closes:** AC5

**Depends on:** T1

**Locations:** `content/skills/workspace-conventions/references/artifact-layout.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** una viñeta después de `secrets/`; las demás reglas del layout no cambian.

**Verification:**
- `rg -n "^- \`scripts/\`" content/skills/workspace-conventions/references/artifact-layout.md` muestra una línea.
- `go test ./tests/content/...` pasa.

### T3 — Documentación para personas y arquitectura

- [x] `site/src/content/docs/getting-started.md` describe en dos o tres frases los scripts locales y la escritura por shell desde un worktree aislado.
- [x] `site/src/content/docs/flows/build.md:28` agrega los scripts locales a lo que build registra.
- [x] `site/src/content/docs/host-limits.md`, sección «Claude Code», nombra el aislamiento que rechaza `Edit`/`Write` al checkout principal, la ruta que prevé el host («edítalo tú o abre una sesión ahí»), la escritura por shell que usa Hive como hueco conocido, y que con el sandbox activo esa escritura también puede rechazarse, en cuyo caso el informe final la nombra.
- [x] `_support/docs/architecture/workspace-and-artifacts.md` explica el porqué en «Machine facts» (la prueba del 2026-11-08 y el caso de la base por worktree), y su línea 56 sobre «permanent maintenance scripts» admite la excepción de los scripts locales.

**Closes:** AC14

**Depends on:** T1

**Locations:** las cuatro páginas anteriores.

**Execution:** hilo principal

**Test approach:** check

**Changes:** describen el comportamiento y enlazan a la guía; no copian la regla. `README.md` y `llms.txt` no cambian: no afirman nada de esto.

**Verification:**
- Cada búsqueda encuentra el texto nuevo, y hoy no encuentra nada:
  - `rg -n "_support/scripts/" site/src/content/docs/getting-started.md`;
  - `rg -n "local script" site/src/content/docs/flows/build.md`;
  - `rg -n "Edit.*Write|Write.*Edit" site/src/content/docs/host-limits.md`, dentro de la sección `## Claude Code`;
  - `rg -n "except" _support/docs/architecture/workspace-and-artifacts.md`, en la línea sobre «permanent maintenance scripts».
- `npm --prefix site run build` pasa (el comando `build` de `site/package.json`, `astro build`).

### T4 — La limpieza de ramas aplica lo pendiente antes de borrar el worktree

**Retired:** D7 A eliminó la escritura diferida; ya no hay nada pendiente que aplicar y la tarea no se ejecuta.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en local | Su comando pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` | Cubierta por la implementación |
| Barrido de dependencias | Candidato final | Búsqueda en la guía global, skills, referencias, roles, `site/`, `_support/docs/architecture/` y `tests/pilot/` de las menciones de escribir en el checkout principal, de scripts y de `HIVE.local.md`; los lugares revisados quedan listados en el PR | Cubierta por la implementación |
| Tamaño de la guía global | Candidato final | Bytes de `content/guidance/global.md` antes (41 843) y después, escritos en el PR | Cubierta por la implementación |
| Sesión observada | Mensaje del commit | Nombra la prueba de H1: Claude Code 2.1.319, sesión `b3015fee`, 2026-11-08, y el rechazo observado | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D5 A) | La respuesta de revisión |

Entrega `automatic` (D4 B): no hay parada para validar. Después de la revisión vienen el push, el PR y el merge.

## Review status and progress

- 2026-11-08 — Plan escrito.
- 2026-11-08 — Primera revisión del plan aplicada: se agrega T4 y se actualizan T1 y T3.
- 2026-11-08 — D7 A: se retira T4; T1 y T3 se rehacen sobre la escritura por shell; se agregan las páginas del sitio de H15 y las verificaciones concretas de H16.
- 2026-11-08 — Tercera revisión: T1 cierra AC15 en lugar de AC11, y las verificaciones de T1 y T3 comprueban texto que hoy no existe.
- 2026-11-08 — Cuarta revisión: T1 cierra AC16 en lugar de AC15 y suma la viñeta 2 y la de autorización; T3 describe la escritura por shell en `host-limits.md` como hueco conocido.
- 2026-11-08 — Quinta revisión: H29 (el conteo de T1 pasa a 4), H30 (el caso del worktree va en una frase aparte de la viñeta de autorización), H31 (restaurar moviendo el `.bak` de vuelta) y H32 (`host-limits.md` menciona el sandbox). H33 (el modo del script antes de tu revisión) queda listado. No cambian criterios ni deltas.
- 2026-11-08 — T1 hecha en `2535db59`: `rg -c "through a shell command"` da 4; las cinco frases nuevas aparecen una vez cada una; `go test ./tests/content/...` pasa; `global.md` pasa de 41 843 a 44 160 bytes.
- 2026-11-08 — T2 hecha en `2c88e669`: aparece la viñeta `scripts/`; `go test ./tests/content/...` pasa.
- 2026-11-08 — T3 hecha en `6531003a`: las cuatro búsquedas encuentran el texto nuevo; `pnpm run build` en `site/` construye 17 páginas (dependencias instaladas con `pnpm install --frozen-lockfile` en el worktree).

## Checkpoint

Etapa: verify, paso 2. Candidato `1e39c910` en `feat/local-scripts-worktree-writes`, en el worktree `.claude/worktrees/local-scripts-worktree-writes`.

Evidencia reutilizada:
- `go test ./...` pasó en `6531003a`. `1e39c910` solo cambia `tasks.md`, así que el resultado sigue valiendo.
- `pnpm run build` en `site/` construyó 17 páginas.
- El CI solo tiene `master-source-branch`, que no aplica a los PR contra `development`.

Recorrido de los criterios, desde el worktree aislado:
- `git check-ignore -v` resuelve el exclude compartido: `HIVE.local.md` y `_support/secrets/` están ignorados y `_support/scripts/` no, así que la regla agregaría la entrada al exclude.
- Sobre un archivo de prueba del checkout principal: el append con `>>` funciona, `sed -i.bak '2s/.*/…/'` cambia una sola línea (lo confirma `diff`), y el modo `600` se conserva.
- Nota: Claude Code rechaza `$(git rev-parse …)` dentro de un comando compuesto; corrido como comando aparte, funciona.

Primera `/code-review` (`development...1e39c910`): H34–H42.
- Arreglados en `554bd12d`: H34, H36, H39, H40 y H42.
- Resueltos por D9 A (solo agregar): H35, H37 y H38.
- H41 ya estaba decidido en D8 A.

Segunda `/code-review`, limitada a los arreglos (`1e39c910...554bd12d`): H44–H51.
- Arreglados con una línea `Unreviewed:` en el commit:
  - H44: las filas de `ACCOUNTS.md` solo se agregan; los cambios van al informe.
  - H45: no se cambia ni se borra nada; el informe sustituye a la marca stale y al recorte por el tope.
  - H46: un repositorio bare se pregunta.
  - H48: un script nuevo solo se escribe si su ruta no existe.
  - H49: el salto de línea se comprueba solo si el archivo existe y no está vacío.
  - H50: se etiqueta la observación en `host-limits.md`.
- Pasan a los hallazgos del cierre:
  - H47: «Secrets» todavía escribe `$(git rev-parse …)`; Claude Code lo rechaza y su error indica separar los comandos.
  - H51: `getting-started.md` resume el comportamiento.
  - H43: Claude Code también rechaza comandos compuestos que no usan Git (observado en esta sesión); su error indica separarlos.

La segunda revisión no cubre lo que la primera no vio. Siguiente acción: suite, push, PR, merge y `flow-close`.
