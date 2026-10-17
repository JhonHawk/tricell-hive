# Diseño

## Contexto verificado

Revisado en `b8d2b785` (`feat/delegation-threshold-global`) el 2026-10-16.

### Evidencia de las sesiones

Comprobado en esta sesión salvo donde se indica.

- **pnpm `e5edd628` (Claude Code, claude-fable-5-1, 2026-10-16)**
  - **U1:** a las 20:53Z, el hijo `hive-build-backend`, limitado a `pnpm11/`, ejecutó `git stash -q`, luego un `git stash pop` que falló y luego `git checkout stash@{0} -- pnpm11`, todo en el checkout compartido. Con eso guardó en el stash los cambios del hilo principal en `pnpm/`. El hilo principal hizo `git stash drop` comprobando solo que una cadena siguiera presente. El stash `1100eebc` aún existe y todas sus líneas están en el árbol, así que no se perdió nada.
  - **U2:** los vigilantes usaban `grep -q -E "^(Finished|error: could not compile|...)"`. Cargo imprime `    Finished` con sangría, y el error de dashmap (`error: patch location ...`) no coincidía. Hubo esperas de 15 minutos (21:18–21:33Z) y de 27 minutos (21:59–22:26Z); en la segunda el check ya había terminado en 14 s.
  - **U3:** los 8 commits usaron `--no-verify` sin preguntar. El hook de pre-commit contiene la guarda del proyecto contra commits de Claude en `main`. `pnpm/scripts/pre-push-rust.sh:11` se salta todos los chequeos de Rust cuando falta `cargo`, y Dylint falló en CI cuatro rondas (archivo de más de 400 líneas, orden de ítems, función de más de 40 líneas).
- **carvento `039a9a2b` (2026-10-13), U1:** un hijo `hive-build-frontend` ejecutó `git stash -q; git stash pop -q` en el monorepo compartido mientras otro hijo editaba `apps/backend`. Esto viene de la observación de Engram #15433; no lo releí en esta sesión.
- **fleetsystem `b6644329` (Claude Code, claude-opus-5-5, 2026-10-16), U2:** se lanzaron en segundo plano `gh pr checks 186 --watch --interval 30` (00:13:45Z y 00:17:44Z), y la sesión los sustituyó por bucles que primero esperan a que exista el run (00:40:56Z). En esta sesión comprobé esos comandos. La salida "no checks reported" la reporta el hijo auditor.

### Texto actual

- **Pie de los 10 roles que escriben** (`hive-build-backend`, `hive-build-frontend`, `hive-build-data`, `hive-build-infra`, `hive-build-kmp`, `hive-design-architecture`, `hive-design-ui`, `hive-write-docs`, `hive-write-tests`, `hive-tune-performance`): una línea idéntica que incluye "Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them." Los roles de solo lectura no la llevan.
- **`delegation.md:29`:** "A child works only in its assigned working tree and named paths, ...".
- **`delegation.md:45`:** "Give each external process you start a completion signal, such as the host's monitor or watch tool, before your turn ends."
- **`delegation.md:61`:** la espera bloqueante con tramos de 5 a 10 minutos.
- **`git-workflow/SKILL.md`:** la sección "Prepare a coherent delivery" (líneas 22–28) regula el staging y el commit, y no menciona los hooks. La línea 38 regula los bypass de protección de ramas.
- **`verification.md:87`:** "Confirm what CI actually executes; do not assume it covers a missing check." Trata la cobertura de CI en la integración, no los hooks locales. No se duplica: U3 queda en `git-workflow` y no repite esta regla.
- **`global.md:79`:** "children stay in their assigned working tree, named paths and authorized effects". No cambia.
- **Sitio:** `site/src/content/docs/flows/build.md:63-64` ("Who commits and merges?") describe quién hace commit. `build.md:37` dice que los checks se vigilan "until they finish", lo que U2 no contradice y por eso no se edita. Ninguna página describe la delegación.
- **Detectores de `tests/pilot`:** ninguno mide stash, esperas ni hooks. No se añaden casos de regresión, como pide AGENTS.md mientras los pilotos están en pausa.
- **`_support/docs/architecture/`:** no menciona el texto del pie de los roles, ni las esperas, ni los hooks.

## Diseño propuesto

### U1: Git que reescribe el árbol (D2-A)

En el pie de los 10 roles, a continuación de "Modify or delete only the paths the brief names, ...", se añade esta frase:

> A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree.

En `delegation.md:29` (brief), después de la primera frase, se añade esta otra:

> A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, changes paths outside the child's own: the child never runs one in a working tree shared with other writers, and neither does the main thread while a child works in that tree; both compare with the base through read-only Git commands or a separate worktree.

Por qué en los dos sitios: el hijo lee su pie, no `delegation.md`, y el hilo principal lee `delegation.md` para preparar el brief. Esto ya pasa hoy con "named paths". La frase del hilo principal también cubre el `git stash` que hizo el hilo principal de fleetsystem a las 01:00:32Z mientras un hijo leía el árbol.

### U2: una señal de fin que cubra todos los resultados

En `delegation.md` se añade un bloque nuevo después de la línea 45, antes del lanzamiento en segundo plano:

> **Condition:** giving an external process a completion signal.
>
> **Actions:** make the signal end on every outcome the process can reach: success, failure, and a target that does not exist yet. Prefer the process's exit status, or a command that reports a final state, over matching its output; when you must match text, first check the pattern against the tool's actual success and failure output, including indentation. When the target may not exist yet, such as checks not yet registered on a pull request you just pushed, use one command that waits for the target to appear and then watches it to a final state, under the wait-stretch rule below. A watch that exits reporting no target, such as `no checks reported`, has not finished: rerun that same command, never a new short sleep-and-check cycle.

(Redacción ajustada por B1 de la revisión del plan, tal como la propuso el revisor.)

Por qué en `delegation.md`: es el único hogar de las esperas, y `global.md:79` ya obliga a leerlo antes de una espera. No hace falta cambiar ningún rol, porque los hijos corren sus comandos largos en primer plano (pie de los roles).

### U3: hooks y chequeos omitidos (D1-A, D3-A)

En `git-workflow/SKILL.md` se añade un párrafo nuevo en "Prepare a coherent delivery", después del párrafo del commit (línea 26):

> Let the repository's Git hooks run on every commit and push. Never skip one, such as with `--no-verify`, `git commit -n`, a variable that disables the hook manager such as `HUSKY=0` or `LEFTHOOK=0`, or an overridden `core.hooksPath`, without the user's explicit approval for that repository; that approval holds for later commits and pushes there for the rest of the session unless the user withdraws it, and you name it when reporting each commit it covers. When a hook fails because this environment lacks a tool it needs, or passes while skipping checks for that reason, read the hook to learn which checks it runs and run them another way with tools already available, since installing a tool is a separate effect that needs its own authorization. Report any check that still cannot run as not run, name the missing tool, and propose installing it. When any check remains not run, ask before pushing. When no one can answer, as under an automatic delivery or unattended delegation, push only after confirming in the workflow files that a job triggered by that push or pull request runs the same commands the hook runs, since a job name alone does not suffice, as [verification](../flow-build/references/verification.md) describes for what CI actually executes; otherwise keep the commit local and report the push as pending.

En `git-workflow/SKILL.md:38`, después de "Use any other bypass only when an explicit grant for that repository and that action covers it, and name the grant when reporting the result; a grant for one bypass, such as creating a branch, does not cover later pushes.", se añade:

> Skipping a Git hook follows the hooks paragraph above instead.

(Re-revisión, aplicada tal como la propuso el revisor:
- R1: la excepción explícita en la línea 38.
- N1: "When any check remains not run".
- N2: la evidencia de CI.

N3 se descarta: la respuesta a D4 contestó la pregunta tal como se formuló, que incluía el modo automático.)

Ajustes por la revisión del plan:
- `git commit -n`, porque en `git push` `-n` significa `--dry-run` (comprobado con `git commit -h` y `git push -h` en git 2.56.0) (N2).
- La cláusula sobre usar solo herramientas ya disponibles (N3).
- El alcance de la aprobación, por D5-A (resto de la sesión en ese repositorio).
- La rama sin respuesta, por la respuesta propia del usuario a D4: "Antes de empujar confirma que hay CI que lo pueda atrapar, si no queda pendiente, en el reporte se incluye la propuesta de instalar las herramienta".

Por qué en `git-workflow`: es la skill que se lee antes de cualquier commit o push, y la línea 24 ya prohíbe alterar historia sin autorización. Su descripción incluye "verifying". El hilo principal pasa la ruta de la skill a un hijo que hace commit, según la regla de delegación.

En `site/src/content/docs/flows/build.md` se añade una frase en "Who commits and merges?":

> Commits and pushes run the repository's hooks. The agent skips one only with your approval, which lasts for the session in that repository. When a hook cannot run its checks in this environment, the agent runs them another way or tells you which did not run and which tools are missing, before pushing; when you cannot answer, it pushes only if the repository's CI runs those checks.

(N4 de la re-revisión.)

### Alternativas descartadas

- **Worktrees obligatorios para hijos que escriben en paralelo (D2-B):** se descartaron en D2 por el coste de preparación e integración en cada delegación.
- **Endurecer `verification.md:87` para U3:** esa regla trata la cobertura de CI en la integración. La omisión de pnpm pasó antes del push y la causó el hook, así que su hogar es `git-workflow`.

## Decisiones tomadas tras la revisión

- **D4 (U3 cuando nadie puede responder), respuesta propia del usuario:** se empuja solo si el CI corre esos mismos chequeos. Si no los corre, el push queda pendiente. En los dos casos, el reporte propone instalar las herramientas que faltan. Resuelve B3: `unattended-delegation/SKILL.md:16` decide sin preguntar.
- **D5-A (alcance de la aprobación):** vale el resto de la sesión en ese repositorio, como las demás elecciones de flujo de `global.md` ("A workflow choice the user made in this session … holds for later work in the same repository"). Es una excepción deliberada frente a `git-workflow/SKILL.md:38`, que pide una concesión por acción para saltarse la protección de ramas. Saltar un hook es otro tipo de bypass y conserva su propio hogar en el párrafo nuevo (N1).

## Mensaje de commit

Según AGENTS.md, nombra las sesiones de origen:
- pnpm `e5edd628` (Claude Code, 2026-10-16);
- carvento `039a9a2b` (Claude Code, 2026-10-13);
- fleetsystem `b6644329` (Claude Code, 2026-10-16).

Y la decisión del usuario D1-A para U3, que tiene un solo caso.
