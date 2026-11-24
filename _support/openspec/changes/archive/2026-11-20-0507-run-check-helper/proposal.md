# run-check-helper — Helper con tope para los checks cuya salida se filtra o se guarda

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #530, merge bd1bbcc1 |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `9b9450c5` · rama propuesta `feat/run-check-helper` |
| Playbook | ninguno (`/flow-plan` tecleado por el usuario) |

## Decisions

2026-11-19
- **Decisión de alcance (A)** — ¿Cómo evitamos otro log que se alimenta a sí mismo? Un helper en `hive-rule-checks/scripts/` que corre el check con la salida a un archivo con tope, cierra el archivo y solo después muestra el final, más un criterio del auditor. Descartado: solo texto, o esperar una segunda observación.
- **D3 A** — ¿En qué forma hacemos el helper? Script Python `run_check.py`, con el comando como argumentos y pruebas en `unittest`. Descartado: script de shell con `ulimit -f` y solo texto.
- **Seguridad A** — ¿Corre `hive-review-security` sobre el candidato? Sí, porque el script es un punto nuevo de ejecución de procesos.
- **Supuesto (reemplazado en la revisión del plan)** — Tope de 256 MiB por log, cambiable con `--max-bytes`. Descartado: sin tope. Se revierte cambiando el valor por defecto.
- **Supuesto (reemplaza al anterior)** — Tope de 64 MiB por log, porque cada corrida deja un log en `$TMPDIR` hasta el cierre. Descartado: 256 MiB. Se revierte cambiando el valor por defecto.
- **Supuesto** — Al terminar muestra las últimas 40 líneas (`--tail`) o las que coinciden con `--grep`; sale con el código del check, o con 125 y un mensaje con la ruta del log cuando salta el tope. Descartado: imprimir todo.
- **Supuesto** — El log queda en `$TMPDIR` como `hive-check-*` con su ruta impresa primero; el cierre lo borra como temporal de la tarea. Descartado: borrarlo al terminar.
- **Supuesto** — La hoja exige el helper para un check cuya salida se guarda; `${PIPESTATUS[0]}` sigue para un pipe directo, que no escribe en un archivo que se lea a sí mismo. Descartado: exigirlo también para pipes. El detalle va en una referencia nueva, porque la huella de lectura de plan-build tiene 17 B de margen. Descartado: recortar otras hojas.
- **Supuesto** — El auditor lee el texto de un `exec` interpolado de Codex en un campo nuevo, `CommandTemplate`, sin cambiar `Command` ni `CommandWithheld`, así ningún otro criterio cambia. Descartado: dejar de retener esos comandos para todos los criterios.
- **Supuesto** — Comportamiento del script fijado en `design.md`: compatible con Python 3.9, stdin vacío, grupo muerto en toda salida, espera de 5 s a nietos, `SIGKILL` a los 5 s, final acotado a 16 KiB, 127 si el comando no existe, `--log` con `O_EXCL`. Descartado: dejarlo al build.
- **Supuesto** — Criterio `self_feeding_log`: detecta un log que se alimenta a sí mismo de verdad (un bloque compuesto o `exec >F` redirigido a `F` que un `trap` o un comando del bloque lee), no un comando simple con redirección. Descartado: detectar solo `tail`, y marcar cualquier redirección más `trap`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «a» — usuario, 2026-11-19, eligiendo: «Planear un helper `run_check` en `hive-rule-checks` que corra el comando con la salida a un archivo temporal con tope de tamaño, cierre el archivo y solo después muestre las últimas líneas conservando el estado, y que la hoja mande usarlo; más un criterio del auditor que detecte un `trap` que lee el mismo archivo que redirige.»

Origen: sesión Codex en project-g `01a1f5d4` (PRK-20, `gpt-6.2-sol`, 2026-11-19 ~22:3xZ). El envoltorio `{ … } >"$project_g_log" 2>&1` con `set -e` y `trap 'tail -35 "$project_g_log"' EXIT` hizo que `tail` escribiera en el log que leía; el archivo llegó a 1.53 TB y llenó el disco. El agente usó ese envoltorio en nueve comandos de la sesión. Watch: `engram:tricell-hive/watch/self-feeding-check-log`. Decisión: `engram:tricell-hive/decision/run-check-helper`.

## Objective

Hoy la hoja pide guardar la salida completa de un check en un archivo y filtrarlo, sin forma segura ni tope, y cada agente escribe su propio envoltorio. Al terminar, el agente corre esos checks con `run_check.py`, que no puede escribir en su propio log ni pasar del tope, y el auditor señala el patrón peligroso si reaparece.

## Scope and acceptance

Dentro: `content/skills/hive-rule-checks/` (hoja, `references/run-check.md` y `scripts/run_check.py`), `tests/skills/run_check_test.py`, `_support/docs/architecture/instruction-resources.md`, el auditor (`tests/pilot`), `rule-enforcers.md`, la rúbrica, `CHANGELOG.md` y la spec `workflow-lifecycle`. Fuera: cambiar la CI del repositorio, otros comandos de shell que no son checks, y Windows.

- AC1 — `python3 -I content/skills/hive-rule-checks/scripts/run_check.py -- <comando>` imprime primero la ruta del log, guarda en él la salida completa del comando, imprime al final como mucho 40 líneas y 16 KiB fuera del log y sale con el código del comando (127 si no existe, 128+n si lo terminó una señal); el log termina con el tamaño de la salida del comando, sin copias de su final, y ningún proceso del grupo sobrevive · false today because el script no existe · Observed in: `python3 -m unittest discover -s tests/skills -p 'run_check_test.py'` y una corrida a mano sobre `go -C <repo> test ./tests/content/...` con un test roto a propósito
- AC2 — Un comando cuya salida supera `--max-bytes` termina (con su grupo de procesos), el log se detiene cerca del tope y el script sale con 125 y un mensaje con la ruta y el tope · false today because no hay tope · Observed in: `run_check_test.py` con un productor sin fin y un tope de 1 MiB, y una corrida a mano con `yes`
- AC3 — `hive-rule-checks` exige `run_check.py` para un check cuya salida se guarda, su referencia da el respaldo sin `python3`, y ningún archivo de `content/` ni `site/` repite la forma «save the whole output to a file … and filter the file» · false today because la hoja la tiene en «Check reuse» · Observed in: el candidato (`rg` y `go test ./tests/content/...`) y la hoja instalada tras `hive update`; la adopción en sesiones reales se sigue en el watch
- AC4 — `--audit-session` da `self_feeding_log` = `fail` en la transcripción Codex `01a1f5d4` con la línea 535 (`prk20-hygiene` y `prk20-turbo`) entre su evidencia y ninguna línea que solo tenga comandos simples con redirección, `not_observed` en la sesión Claude `bf678880`, y los demás criterios no cambian sobre `01a1f5d4` · false today because el criterio no existe · Observed in: la salida del auditor sobre `~/.codex/sessions/2026/11/19/rollout-2026-11-19T16-01-30-01a1f5d4-3ec7-7631-97d8-ba540210ac98.jsonl`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/hive-rule-checks/SKILL.md:15` (2,900 bytes; tope de una hoja: 4,000) pide «save the whole output to a file in the host's temporary location and filter the file»; ningún otro archivo de `content/` o `site/` lo repite (`rg`, 2026-11-19, `9b9450c5`).
- Convención de scripts: Python con `python3 -I <dir>/scripts/<x>.py` (`hive-rule-trackers/scripts/work_time.py`, `jira_rest.py`); pruebas en `tests/skills/*_test.py` con `unittest` (`CONTRIBUTING.md:23`); `_support/docs/architecture/instruction-resources.md:36` describe `hive-rule-checks`.
- `hive-rule-security-boundaries`: argumentos de proceso estructurados, nunca interpolados en un shell.
- El envoltorio que falló está en el rollout `01a1f5d4` (cuatro usos: `prk20-hygiene`, `prk20-turbo`, `prk20-integration`, `prk20-turbo-retry`).

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` con revisión nativa `/code-review` (respuestas vigentes en esta sesión y repositorio, 2026-11-19), más `hive-review-security` sobre el candidato (Seguridad A). No autoriza promover a `master` ni publicar una versión.

## Review status and progress

2026-11-19 — Plan escrito. Revisión `hive-review-plan` (2 P1, 6 P2, 8 P3), todo aplicado: el auditor lee el `exec` interpolado de Codex en `CommandTemplate` (F1); el detalle va a `references/run-check.md` por la huella de lectura (F2); pipes con `PIPESTATUS` siguen (F3); el criterio detecta auto-alimentación real con más fixtures (F4); comportamiento del script fijado y probado (F5, F6); fixtures sintéticos (F7); respaldo sin `python3` (F8); AC3 se observa en el candidato, AC4 nombra la línea 535, nueve usos, ruta completa de `bf678880`, commits con la observación, `<hive-rule-checks-dir>`, tope de 64 MiB, `site/` sin cambio con su razón. `check_change.py` sale con 0. Estado: listo para `flow-build`.
2026-11-20 — Build T1–T4 (`5c5b1b2`…`ae8be52`). `/code-review` (5 importantes) y `hive-review-security` (1 P2, 8 P3): corregidos en `66b1364` y `379e15c` el `--grep` inválido, el comando huérfano si el helper muere por SIGKILL (guardián con pipe), la señal durante el arranque y cuatro de precisión del detector. Segunda revisión: 3 importantes corregidos en local con prueba (`449b39c`, `1aa389b`, con `Unreviewed:`): guardián con doble fork para no bloquear a un comando que espera a todos sus hijos, cierre de descriptores por `/dev/fd`, 127 si falla el arranque, marcadores iguales para la misma interpolación y objetivo de redirección con espacio. Hallazgos de cierre (no bloqueantes): estados 2 y 125 que coinciden con los de otras herramientas; `--grep` con retroceso exponencial ignora señales; el temporal puede caer en el directorio actual; los logs quedan hasta el cierre; consejo sobre datos en `bash -c`; señal antes de `Popen` lanza igual el comando; `--grep ''` muestra el inicio; detector: nombres como subcadena, here-strings, `while`/`if`/funciones en `trap`, `exec >&N` en texto citado, `<` y `>|` cortan la búsqueda, tiempo cuadrático con llaves sin pareja. `TestE2ESetupDeadlinesOwnDescendantsAndPipes` sigue siendo intermitente en la base.
