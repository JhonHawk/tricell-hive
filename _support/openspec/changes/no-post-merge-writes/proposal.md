# no-post-merge-writes — Nada se escribe en una carpeta de cambio después de su merge

| Field | Value |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | none |
| Git | tricell-hive · base `development` en `1b87f0a7` · rama `fix/no-post-merge-writes` · worktree `.claude/worktrees/no-post-merge-writes` |
| Playbook | none |

## Decisions

2026-11-21
- **D8 A** — ¿Cuándo se archiva? Al cerrar cada trabajo; reemplazada por D9 A el mismo día, que mantiene el lote de D14 A.
- **D9 A** — ¿Cómo evitamos que el repo quede sucio después del merge? Se mantiene el archivo en lote (D14 A) y, después del merge, el cierre no escribe en la carpeta: su checkpoint va al temporal del host; el estado «Completado» solo lo escribe el lote dentro de su commit. **Replaces:** D8 A.
- **D10 A** — ¿Cómo se hace cumplir? El script; la condición de esta línea quedó reemplazada por el primer «Supuesto (revisión del plan)» de abajo. `tasks.py checkpoint` y `done` se niegan a escribir en `tasks.md` cuando el checkout está en la base o en un commit sin rama y la carpeta ya está en `origin/<base>`, salvo con `Delivery: direct-base`; `checkpoint --out <archivo>` escribe el checkpoint en ese archivo.
- **Dónde A** — Worktree propio, porque hay otras sesiones activas.
- **D11 A** — ¿Se cambia la condición de la negativa? Sí: **Replaces:** D10 A en su condición (el script la sigue haciendo cumplir). La negativa no mira la rama: se dispara cuando la carpeta ya está en `origin/<base>` y todas sus tareas están marcadas en esa copia integrada, en cualquier checkout y sin excepción por `Delivery`; descartado: «checkout en la base o sin rama, salvo `Delivery: direct-base`», que no cubre el caso más común (un checkout que sigue en la rama ya integrada) y bloquea un `direct-base` respondido en la sesión, que el script no puede leer. Con tareas abiertas (varios PRs, una tarea posterior al merge, `direct-base` a mitad del build) se sigue escribiendo como hoy. Un error al leer Git nunca niega. Se revierte volviendo a la condición por rama.
- **Supuesto** (revisión del plan) — Lo que quedó sin commit antes del merge (un checkpoint escrito después del último push) lo resuelve la salida de `versioned`, que el cierre ya sigue (`flow-close` paso 3): si solo cambió el checkpoint, indica `git restore`; si cambió algo más, lo nombra para el reporte. Descartado: mantener la frase de `git restore` en `close-hygiene`, que no cabe en el presupuesto de lectura con la frase nueva.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «lo que no quiero es que se quede el repo sucio...»
>
> «que no se necesite actualizar el estado despues de mergeado (o se haga, pero en el proceso bulk de archivar), porque openspec no actualiza esos archivos despues de cerrado el pr o si?»

Sin ticket. Fuera de alcance: cuándo se archiva (queda en lote, D14 A) y qué escribe el lote.

## Objective

Hoy el cierre corre `tasks.py checkpoint` sobre la carpeta ya integrada y deja `tasks.md` modificado sin commit (pasó en el cierre de `project-skill-registry-mobile`); `close-hygiene` manda hacer `git restore` y `git-workflow` paso 4 dice lo contrario, que esas ediciones se integran con el archivo del lote. Al terminar, el script no deja escribir en una carpeta integrada con todas sus tareas marcadas, desde ningún checkout; el cierre guarda su checkpoint en el temporal del host; y las reglas del repositorio del trabajo ya no hablan de ediciones posteriores al merge (`batch-archive.md` mantiene a propósito el caso de un repositorio de specs aparte).

## Scope and acceptance

- AC1 — En un checkout en la base, o en un commit sin rama, con la carpeta ya en `origin/<base>` y sin `Delivery: direct-base`, `tasks.py <folder> checkpoint …` sale con código distinto de 0. **Retired:** la condición pasó a «integrada y con todas sus tareas marcadas», ver AC6.
- AC6 — Con la carpeta ya en `origin/<base>` (merge normal o por squash) y todas sus tareas marcadas, `tasks.py <folder> checkpoint …` sale con 1 desde la base, desde un commit sin rama y desde la rama ya integrada, nombra `--out` y deja `tasks.md` idéntico byte a byte, y `done` sale con 1 sin escribir; con una tarea abierta, sin commits, sin remoto o con la carpeta solo en su rama, escribe como hoy · false today because el comando escribe el checkpoint y sale 0 en todos esos casos · Observed in: pruebas con un repositorio temporal y su remoto desnudo **Replaces:** AC1
- AC2 — En el caso de AC6, `checkpoint … --out <archivo>` crea ese archivo con solo el bloque del checkpoint, imprime su ruta absoluta, conserva sus líneas de evidencia en la siguiente llamada y no toca `tasks.md`; rechaza un `--out` dentro del repositorio de la carpeta · false today because `--out` no existe · Observed in: pruebas con un repositorio temporal
- AC3 — En ese mismo caso, `tasks.py <folder> done T<n> …` sale con código distinto de 0. **Retired:** con todas las tareas marcadas, `done` no tiene nada que hacer; queda dentro de AC6.
- AC4 — `close-hygiene.md` ya no manda `git restore` de la carpeta y dice que después del merge el cierre escribe y lee su checkpoint en `<temporal del host>/<change-id>-checkpoint.md` con `--out`; `git-workflow` paso 4 y `site/.../flows/close.md` ya no dicen que las ediciones posteriores al merge esperan al archivo; `tasks.py versioned`, con `tasks.md` cambiado solo en su checkpoint, imprime `git restore -- <ruta>` en lugar de «later edits wait for the archive», y con otros cambios los nombra para el reporte · false today because las frases están en `close-hygiene.md:11`, `git-workflow/SKILL.md:18`, `flows/close.md:39` y `tasks.py:41,884` · Observed in: el texto del candidato y la prueba de `versioned`
- AC5 — Tras el cierre de este mismo cambio, `git status` del checkout donde corrió el cierre no muestra cambios en la carpeta de cambio · false today because el cierre de `project-skill-registry-mobile` dejó `tasks.md` modificado · Observed in: el cierre de este cambio, con la versión desplegada

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `python3 -m unittest discover -s tests/skills -p 'tasks_test.py'` en la base · pasa · ~22 s · 2026-11-21
- `go test ./tests/content/...` en la base · pasa · ~2 s · 2026-11-21
- `grep -c 'git restore' content/skills/hive-core/references/close-hygiene.md` en la base · 1 · 2026-11-21
- `go test ./tests/pilot/ -run TestAuditNewCriteria` en la base `1b87f0a7` (worktree) · pasa · 2026-11-21
- `go vet ./...` · pasa en `0fa546c2` · 2026-11-21
- Presupuestos (lectura del revisor del plan, misma fórmula que las pruebas) · lectura de plan y build 10 B libres (incluye `close-hygiene.md` y `checkpoint.md`); paquete de `flow-build` 29 B; `flow-close` 268 B; `flow-verify` entrada 1,064 B · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` y `go test ./...` en la base · pasaron en `0fa546c2` (antes de #582, que solo archiva) · 2026-11-21
- OpenSpec (`/Volumes/TRICELL_VAULT/dev-resources/reference/openspec` en `0c737dc1`) · `src/core/archive.ts` solo escribe los specs principales y mueve la carpeta; `docs/commands.md` «Preserves all artifacts» · nunca edita la carpeta después del PR · 2026-11-21

## Effects

None: the tasks take no effect on an external system.

## Delivery

Respuestas en vigor de esta sesión y repositorio: `automatic`, revisión nativa `/code-review`. Sin revisión de seguridad: no toca una frontera de seguridad.

## Review status and progress

2026-11-21 — plan escrito.
2026-11-21 — relectura de ambos revisores: resueltos los P1 y P2; nuevos aplicados (la negativa lee la copia integrada de `tasks.md`; `git-workflow` paso 4 con la excepción de tareas abiertas; D10 A reemplazada por D11 A con la confirmación del usuario).
2026-11-21 — revisión por dominio: guía (2 P1, 4 P2, 3 P3; ~138k tokens, ~2.7 min) y servidor (1 P1, 5 P2, 3 P3; ~143k tokens, ~4 min). Aplicados: condición sin rama y sin excepción por `Delivery` (P1 de ambos, P2 de rama integrada y de escrituras legítimas); un error de Git nunca niega; `--out` crea el archivo, solo el bloque, imprime la ruta, rechaza rutas dentro del repositorio; `versioned` resuelve lo que quedó sin commit antes del merge; ruta fija del checkpoint del cierre y edición neta cero de `checkpoint.md`; página del sitio y detector de sesiones en el alcance; presupuestos en el contexto verificado.
