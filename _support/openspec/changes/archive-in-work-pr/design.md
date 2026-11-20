# Design

## Approaches

- **A — Antes del primer push (elegido, D9 A).** `flow-verify` paso 5, tras la validación y la ronda de fixes y antes del CI local: decide el material de soporte, archiva el change folder y corre el CI local sobre el resultado. Un solo push, sin corrida de CI extra; el PR muestra el spec ya fusionado. El detalle del estado es el nombre de la rama. Un fix posterior al push (CI, bot) que cambie el comportamiento edita juntos el spec fusionado y el delta archivado, y anota la línea en el `tasks.md` archivado.
- **B — Justo antes del merge o de la parada `ready`.** Tras CI y revisiones del PR: refleja todo y el detalle puede ser `PR #N`, a cambio de un push y una corrida de CI más por trabajo y de repetir la comprobación de base actual.
- Descartado: mantener el archivado en el cierre (lo que el usuario pidió cambiar).

Peldaño de reutilización: A reutiliza `apply_deltas.py`, `check_change.py --archive` y el inventario de `support-material.md` sin cambios de código; solo mueve el momento en que se invocan.

## Decisions

- **D9 A** — momento: antes del primer push.
- **D11 A** — `automatic` y `ready` no se detienen: opción recomendada del material de soporte, reportada con cómo revertirla; con tareas sin marcar no archiva. `interactive`: ambas preguntas en la ronda de validación.
- **D12 A** — checkpoint posterior al archivado: con el siguiente fix; tras el último push, el cierre lo descarta.
- **D10 B** — el material de soporte decidido (commit, archive, promote) viaja en el mismo PR; las eliminaciones siguen en el cierre (supuesto en `proposal.md`).

## Proposed shape

Orden en `flow-verify` paso 5, antes del CI local, para un trabajo cuya entrega commitea (no `hold`) y con `Specs` en este repositorio:

1. Inventario del material de soporte (`support-material.md`, sección nueva «Before the push»): solo las acciones commit, archive y promote; la ronda de esa decisión se une a la pregunta de «archivar igual» cuando `check_change.py --archive` lista tareas sin marcar. Delete y keep no se preguntan aquí.
2. Commit de los ítems respondidos en la rama del trabajo.
3. Archivado (`hive-rule-close-hygiene` «Archive the change folder»): `check_change.py --archive`, `apply_deltas.py <folder> --specs <Specs> --detail "<branch>"`, revisar, `--apply`, commit en la rama del trabajo.
4. CI local sobre el candidato resultante (paso 5 actual).

Después del archivado, las líneas de checkpoint van al `tasks.md` de la carpeta archivada.

El cierre (`flow-close`):
- paso 2 pregunta solo lo que el paso previo al push no decidió: eliminaciones y material creado después del push;
- paso 4 archiva solo un change folder que su rama no archivó (excepciones), en una rama nueva desde la base, como hoy.

`git-workflow`: el paso 4 deja de decir «or after the merge with the close's archive» y apunta al paso previo al push; el paso 6 ya ejecuta «the checks of flow-verify step 5», que pasa a incluir estos registros, así que un cambio mecánico sin verify también los hace.

## Impact

- Sin cambios de código ni de contrato de scripts.
- Riesgo: dos ramas que cambian el mismo requisito y se fusionan en orden distinto al de su archivado. `spec_deltas_test.go` falla en local al traer la base; la corrección es renombrar la carpeta archivada a la hora UTC actual (supuesto en `proposal.md`).
- `hive-rule-close-modes`: el paso 4 del cierre normalmente ya no tiene carpeta que archivar; el paso 2 sigue esperando cuando hay material de verificación en `workspace/` que borrar, así que la fila de cierre inmediato solo se cumple algo más a menudo.
- Fix posterior al push: edita juntos el spec y el delta archivado (supuesto en `proposal.md`).
