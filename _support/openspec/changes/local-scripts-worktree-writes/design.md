# Design

## Approaches

**Escritura en el checkout principal desde un worktree (H1).** Primero se eligió diferir, con lo pendiente en un archivo ignorado del worktree. La segunda revisión mostró que diferir depende de una salida del worktree que en Claude Code solo ocurre a pedido del usuario, y que `ExitWorktree` con `remove` borra lo pendiente. Diferir, además, necesitaba cuatro reglas más (H11–H14).

D7 A lo reemplaza, y D8 A lo confirma: escribir por shell, bajo los permisos del usuario. No es la ruta que prevé Claude Code («edit it yourself or start an interactive session there»), sino un hueco que su aislamiento no vigila; `host-limits.md` lo describe así.
- Lo que se gana: una frase, que sirve en todos los flujos.
- Lo que se pierde: editar una línea por shell es más propenso a errores que `Edit`.
- Opción descartada: no escribir desde el worktree (D7 C), que pierde hechos y cuentas.

**Ubicación de los scripts.**
- A, elegida: `_support/scripts/`, ignorado.
- B: `~/.config/hive/scripts/<proyecto>/`, invisible desde el proyecto.
- C: ninguna carpeta.

**Autorización de los scripts.**
- B, elegida: el agente los crea sin preguntar en los flujos listados y los nombra en el informe final.
- A: una aprobación por cada script.

## Decisions

Ver la sección de decisiones de `proposal.md`: D2 A, D3 B y D7 A. D7 A reemplaza la decisión de diferir y la de dónde guardar lo pendiente.

## Proposed shape

**Hogar de cada regla.** «Local instructions» de la guía global es el hogar canónico de las dos piezas, porque debe actuar sin que se lea otro archivo:
- la escritura por shell: se amplía la primera viñeta;
- los scripts locales: una viñeta nueva, después de la viñeta de autorización («This rule authorizes, in any flow, read-only ones included, writing these files…»), para que «these files» no abarque los scripts (H19).

Las demás menciones son punteros:
- «Test accounts» remite a «Local instructions».
- «Support folder» y «Proportionality» exceptúan el script local y nombran el elemento que cambian.
- `artifact-layout.md` lista `scripts/`.

**Texto propuesto** (inglés, como toda la guía distribuida):

- Viñeta 1 de «Local instructions»: «…; in a worktree, read and write the main checkout's file. When the host refuses that edit from the worktree, make the same write through a shell command, which the host's permission settings govern: append a new entry with `>>`; to change one existing line, rewrite that line, addressed by its line number, with `sed -i.bak`, confirm with `diff <file>.bak <file>` that only that line changed, then remove the `.bak` file, or, when the check fails, restore the file by moving the `.bak` file back. When the shell write is refused too, write nothing and name the update in the final report. Never write a copy of the file inside the worktree.»
- Viñeta 2 de «Local instructions» (H28): «Write only in those files of the current project, the local scripts below, and in the machine file.»
- Viñeta de autorización de «Local instructions» (H22, H30), una frase nueva después de la que termina en «…when that folder is versioned.»: «From a worktree whose host refuses commands in the main checkout, run that check from the worktree on the same relative path and add a missing entry to that exclude file through a shell command.»
- Viñeta nueva: «A project's `HIVE.local.md` may name a local script, a helper for this machine that the agent wrote, such as one that creates a database per worktree; the machine file names none. The script lives at `_support/scripts/<name>` under the root of the file that names it, kept out of Git as the bullet above sets for these files, and the entry names it by that path; make a new script executable. From a worktree, run a repository's script by the same path under the main checkout, the folder that holds `$(git rev-parse --path-format=absolute --git-common-dir)`, with the worktree as working directory; when the worktree is a standalone copy that path does not reach, ask the user for the main checkout. The main thread writes or changes a script without asking only in `flow-build`, including the fixes `flow-verify` makes, or in an operational task, and from a worktree through a shell command when the host refuses the edit, as above; never in research, status, plan, a review or a child, which reports the script it needs instead. A script holds no secret value: it reads the secret location's files. The final report names each script the work created or changed. A script every clone needs is not local: name it as a candidate for the repository's tracked code, as a fact that holds for every clone is.»
- «Test accounts»: «…; in a worktree, read and write the main checkout's file, through a shell command when the host refuses the edit, as «Local instructions» sets.»
- «Support folder»: «Application code, scripts, and configuration stay in their source locations, even when an agent wrote them, except the local scripts «Local instructions» sets.»
- «Proportionality»: «Adding files to a repository so the task can run, such as a script or an image definition, is the user's choice: ask it before the first such file; a local script «Local instructions» sets is not such a file.»
- `artifact-layout.md`, después de `secrets/`: «`scripts/`: never versioned, the local scripts that `HIVE.local.md` names, as the global guidance's «Local instructions» sets.»

## Impact

- Lectores de las viñetas que cambian:
  - `content/skills/flow-build/references/delegation.md:28`: da a los hijos la ruta del `HIVE.local.md` del checkout principal; no cambia.
  - `content/skills/local-instructions/SKILL.md`: remite a «Local instructions»; no cambia (H10 queda como límite conocido).
  - `content/skills/harness-audit/references/project-layout.md` HA-PL-05: los scripts están ignorados; no cambia.
  - Detectores `tests/pilot/flows.go:557-573`: exentan `HIVE.local.md` y el exclude en flujos de solo lectura. Los scripts nunca se escriben en solo lectura. No se verificó si el detector reconoce la ruta absoluta del checkout principal cuando la escritura viene de un worktree; con los pilotos pausados, no cambia.
  - `flow-close` y `branch-cleanup.md`: sin cambios, porque ya no hay nada pendiente que aplicar.
- El cambio no toca producción ni código ejecutable de Hive: solo guía, specs y documentación.

## Recovery

Revertir el commit restaura el texto anterior. Los scripts que los agentes ya hayan escrito en `_support/scripts/` de otros proyectos quedan en disco: borrarlos requiere la autorización del usuario («Preservation»).
