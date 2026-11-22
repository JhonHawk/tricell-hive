# Design

## Approaches

1. **El script lo hace cumplir** — elegido (D10 A, con la condición de D11 A). `tasks.py` ya calcula si la carpeta está en `origin/<base>` para `versioned` (`tasks.py:850-905`); `checkpoint` y `done` reutilizan ese cálculo. Cubre el fallo observado: el cierre corrió con el worktree en el merge commit, sin rama.
2. **Solo texto** — descartado: depende de que el agente lo recuerde, y hoy no lo hizo.

## Decisions

- **Cuándo se niega** (`checkpoint` y `done`): la carpeta tiene commit, su último commit (o su contenido, para un merge por squash) está en `origin/<base>`, como ya calcula `versioned`, y todas las casillas de sus tareas `### T<n>` están marcadas en la copia integrada, y la copia de trabajo coincide con ella fuera de `## Checkpoint` (`git show origin/<base>:<carpeta>/tasks.md`), no en la copia de trabajo: así un `done` en la rama de una entrega posterior, o a mitad de un build `direct-base`, no dispara la negativa antes de su commit. No importa la rama ni `Delivery`: una tarea abierta (otro PR del mismo plan, una tarea posterior al merge, `direct-base` a mitad del build) sigue escribiendo como hoy. Un error al leer Git, un repositorio sin commits, sin remoto o sin base resoluble no niega. La comprobación corre solo al ejecutar `checkpoint` o `done`, nunca al importar el módulo (`check_change.py` lo importa).
- **Mensaje de `checkpoint`:** `refused: <folder> is on origin/<base> with every task done; nothing is written to it until the hygiene batch archives it. Write the checkpoint with --out <host temp>/<repo>-<change-id>-checkpoint.md; to continue the work, reopen its task on a new branch first.` Sale con 1.
- **Mensaje de `done`:** el mismo sin la frase de `--out`. Sale con 1.
- **`--out <archivo>`:** solo en `checkpoint`. El archivo guarda solo el bloque `## Checkpoint`: lo lee si existe, si no toma el de `tasks.md`, aplica la misma actualización, lo crea si falta o está vacío (sin copiar permisos de un archivo inexistente), rechaza un archivo que tenga otra cosa, nunca escribe `tasks.md`, e imprime `checkpoint written to <ruta absoluta>`. Rechaza una ruta dentro del repositorio de la carpeta, porque lo volvería a ensuciar.
- **`versioned`:** para un `tasks.md` modificado cuyo diff contra `HEAD` toca solo la sección `## Checkpoint`: `warning: modified <path>: only its checkpoint changed; restore it with git -C <checkout> restore -- <path>`. En los otros worktrees del repositorio, el archivo solo se nombra para el reporte, sin `restore`, porque puede ser trabajo de otra sesión. Para otro archivo modificado o sin seguimiento en el repositorio del trabajo: `warning: modified|untracked <path>: name it in the close report`. El docstring y el uso del módulo se actualizan.
- **Texto:** `close-hygiene` «Clean up» cambia la frase del `git restore` por: después del merge, el cierre escribe y lee su checkpoint con `--out` en `<temporal del host>/<change-id>-checkpoint.md`, que también lee tras una compactación. `checkpoint.md:9` cambia «the checkpoint file of a plan kept in the conversation» por «a checkpoint file in the host's temporary location», neta cero, porque el paquete de `flow-build` tiene 29 B libres y la lectura de plan y build 10 B. `git-workflow` paso 4 cambia «or after the merge with the hygiene batch's archive» por «; once it is merged with every task done, nothing is written to it until the hygiene batch archives it», y agrega que bajo `direct-base` el último `done` entra antes del commit final. `site/.../flows/close.md:39` se reescribe. `batch-archive.md:16` (Specs en otro repositorio) no cambia: copia las ediciones sin commit de un repositorio de specs aparte, que sigue siendo válido.
- **Detector de sesiones:** `commandWrites` (`tests/pilot/session_audit_specs.go`) no cuenta `checkpoint … --out` como escritura en la carpeta, y el comentario de `changeFolderUncommitted` deja de decir que las ediciones esperan al archivo; su veredicto no cambia.

## Impact

`tasks.py` lo usan `flow-build`, `flow-verify`, `flow-close` y `hive-verify-task`; antes del merge su comportamiento no cambia. Ningún otro script llama a `checkpoint` ni a `done`.

## Recovery

Revertir el merge vuelve al comportamiento anterior; no hay datos que migrar.
