# Diseño

## Enfoques

Comparados en la investigación de esta sesión (2026-11-21):

- **pstack** (`skills/poteto-mode/playbooks/worktree-cleanup.md`): no hay etapa de cierre. La limpieza es un playbook aparte que audita con `worktree-audit.sh` y borra sin preguntar lo que está limpio, fusionado y sin uso; solo pausa ante cambios sin commit o un worktree en uso.
- **Hive antes de `73046b86` (3-11)**: el cierre corría todo de una vez y preguntaba el material de soporte antes de limpiar; al final hacía la pregunta de cierre.
- **Hive actual**: al llegar al merge difiere los pasos 2 y 5 a la respuesta «Close»; desde `c0f92d09` (20-11) el archivo va en una respuesta «Hygiene» aparte.

Un solo enfoque sensato para construirlo: cambiar la tabla de `close-modes` y las reglas que dependen de ella, el helper y la auditoría. La regla vive en esas hojas; no hay otra capa donde ponerla. Se elige el híbrido: el criterio de pstack para la limpieza (limpio, fusionado y sin uso se borra sin preguntar) y el motivo de `73046b86` (no interrumpir con una pregunta al cerrar), que se conserva porque el paso 2 ya no pregunta en el caso normal.

## Decisiones

Ver `proposal.md` «Decisiones».

## Forma propuesta

**`global.md` «Preservation».** La lista de lo autorizado al cierre añade: borrar el material de verificación de la obra que su ticket o el cuerpo de su PR fusionado ya tienen, como fija `support-material`. Cabe en el margen (29.433 de 30.000 B); el PR indica el tamaño.

**`close-modes.md`.** La fila «Completion point reached…» corre los pasos 1 a 7 en orden y termina con el pie; desaparece la fila «when step 2's inventory lists no item…». Queda pendiente, para la respuesta «Close», solo lo que necesita al usuario: lo que el helper conserva por una razón que el usuario resuelve, un volcado o export, y cada recurso remoto, cada uno con su pregunta, nunca preguntados antes del paso 6. La fila de cierres sin espera (continuación, abandono, limpieza de despliegue, desatendido, y el fin de sesión de una obra sin cierre previo) deja de nombrar «step 2's question»; corre todo y pregunta solo esos mismos elementos. El fin de sesión es un mensaje del usuario que la termina sin pedir más trabajo; antes del último mensaje repite la limpieza, y ese mensaje lista cada elemento aún pendiente con su pregunta, cada rama conservada con la forma ``kept `<rama>`: <razón>`` de «Report», y cuenta los archivos en espera sin entregarlos. Un pedido de higiene tras el cierre de una obra corre solo lo pendiente y el lote, y su mensaje reporta solo eso; uno sin cierre previo sigue siendo un cierre completo. «When steps wait» pierde la frase de la pregunta del paso 2 en la primera ronda de la obra siguiente. «Close a work item once» queda.

**`support-material.md`.** El inventario propone solo «Delete», «Promote» y «Keep». «Delete» acepta como copia la evidencia cuyo adjunto muestra la lista de adjuntos del ticket leída en el paso 1, además del cuerpo del PR fusionado. El paso no pregunta: aplica los borrados, mueve lo conservado del worktree al checkout principal como hoy, y nombra una promoción o una carpeta `sessions/` cerrada en «Next» del reporte con el comando de `project-layout`. Se quitan «Commit», «Archive» y el registro de acciones para el lote.

**`trackers.md`.** La relectura del paso 1 lee la lista de adjuntos, no solo el comentario, cuando un borrado depende de ella; una evidencia que solo se lista en el comentario para que el usuario la adjunte no cuenta como copia.

**`layout.md` «Retain evidence», `pr-captures.md`, `viewer-close.md`, `workspace-and-artifacts.md`.** Nombran el ticket, además del cuerpo del PR, como registro que reemplaza una captura o un video.

**`close-hygiene.md`, `batch-archive.md`, `archive.md`, `inventory.md`.** El registro `archive-pending/<id>` guarda la carpeta, el PR y el merge. `batch-archive` quita el paso 6 «Support» y las acciones de soporte de los pasos 1, 4 y 7. `archive.md` e `inventory.md` dejan de nombrar a `flow-close` como quien les pasa acciones; las llama `project-layout`.

**`report-endings.md` «The close footer».** «Close» existe solo mientras `close-modes` deja algo pendiente; «Hygiene» con su número queda igual.

**`close-report.md`.** La fila de un pedido de higiene después del cierre de la obra reporta el lote y lo pendiente que corrió, sin el reporte final de la obra. «Next» nombra la promoción y la carpeta `sessions/` cerrada.

**`flow-close/SKILL.md`, `branch-cleanup.md`, `unattended-delegation/SKILL.md`.** Dejan de nombrar «support decision»: lo conservado se mueve al checkout principal y el inventario se aplica. `branch-cleanup.md` «Report» fija la forma «kept `<rama>`: <razón>» y añade `stray_clones`.

**`cleanup_branches.py`.** Nueva clave `stray_clones` en el plan (con y sin `--apply`): cada subcarpeta directa de `.worktrees/`, `.claude/worktrees/` y `.grok/worktrees/` del checkout principal y de `.worktrees/` de su carpeta padre (resueltas desde el checkout principal, no desde `--repo`) que tenga un directorio `.git` propio y cuyo `origin`, normalizado (host y ruta, sin esquema, usuario ni sufijo `.git`), coincida con el del repositorio; con `path`, `branch`, `dirty` (cambios o archivos sin seguimiento) y `unpushed` (commits que no están en ningún remoto). Nunca la borra ni la pasa a `apply_cleanup`; el docstring lo dice.

**`session_audit.go`.**
- Un borrado por el helper cuenta aunque el script se invoque por una variable: el comando contiene `cleanup_branches.py` y lleva `--apply`.
- `close_before_branch_cleanup`: `closeBeforeCleanupStatus` pasa a `fail`. Queda `review` solo cuando el reporte tiene la forma ``kept `<rama>` `` y esa rama está entre las borradas de la sesión (`--merged-pr`, `git branch -d/-D`, `git push --delete`). El comentario D11 se reescribe con la decisión nueva.
- `merged_branch_deleted`: también falla cuando, después del último merge sin borrado posterior, el usuario escribe un mensaje de fin de sesión y no sigue ningún borrado, salvo, con `review`, cuando ese merge es una promoción cuya cabeza la sesión muestra (`gh pr create --head` o una lectura de `headRefName`) y es la base o una rama de `Environments` del `AGENTS.md` de la sesión, o cuando el último mensaje del agente dice ``kept `<rama>` `` para la rama de ese merge, vista por la misma fuente; un merge cuya cabeza no se ve cuenta como merge de trabajo. Un mensaje de fin de sesión empieza con una frase del subconjunto («es todo», «eso es todo», «terminamos», «por hoy», «por el momento es todo», «that's all», «that's it for now»), tiene 10 palabras o menos y no lleva «?». En Codex el mensaje se ubica con `s.Prompts` y `s.Turn`.

## Impacto

Lo usan `flow-close`, `git-workflow` paso 10 (entrega a la limpieza del cierre), `flow-deploy` (su cierre de limpieza ya no difiere), `unattended-delegation` (nombra la decisión de material de soporte), `stage_manifest_test.go` (miembros del paquete de cierre, 22.099 de 22.500 B), `hive-mode/playbooks/ticket-status.md` (cuenta archivos pendientes; sin cambio), `reading_footprint_test.go` (presupuesto de bytes de las hojas que lee el cierre) y `single_home_test.go` («Close a work item once» sigue en `close-modes.md`). La página `site/src/content/docs/flows/close.md` describe la deferral y el material de soporte. No toca producción: es guía distribuida, un script local y una auditoría de transcripciones.

## Recuperación

Revertir el PR restaura las hojas, el helper y la auditoría; `hive update` vuelve a desplegar la versión anterior. La evidencia que un cierre borró con esta regla no se recupera desde el disco, pero el ticket o el cuerpo del PR tienen su copia por construcción.
