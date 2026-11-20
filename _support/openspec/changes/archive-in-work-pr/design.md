# Design

## Approaches

- **A — Antes del primer push (D9 A, reemplazado por D14 A).** Archivar y commitear el soporte decidido en la rama del trabajo, en `flow-verify` paso 5. Se construyó (T1–T4) y se revirtió (f18f1977): el ciclo plan→build→verify→close pasaba a leer 2.708 B por encima de su línea base de `reading_footprint_test.go` (961 B aun con #554), y añadía D11 y D12 para los modos sin parada y las líneas de checkpoint posteriores al push.
- **B — Justo antes del merge o de la parada `ready`.** Descartado en D9: un push y una corrida de CI más por trabajo.
- **C — Por lotes en un PR de higiene (elegido, D14 A).** El cierre registra el archivado y las acciones de soporte como pendientes; «haz higiene» los entrega todos en una rama y un PR. Es la convención que la guía de OpenSpec recomienda («a tiny follow-up commit or a scheduled cleanup»). Costo: `specs/` va detrás del código hasta el lote, así que `flow-plan` lee los deltas pendientes.

Peldaño de reutilización: C reutiliza la entrada de higiene que `flow-close` ya tiene («a hygiene request» corre todos sus pasos), `apply_deltas.py` y `check_change.py --archive` sin cambios, y el registro de pendientes en memoria que el cierre ya escribe. El procedimiento del lote va en una referencia de `flow-close` que solo se lee en higiene, fuera del ciclo de cada trabajo.

## Decisions

- **D14 A** — archivado por lotes en un PR de higiene; reemplaza D9 A, D10 B, D11 A y D12 A.
- Supuestos en `proposal.md`: el repositorio como fuente de verdad de lo pendiente, con la memoria para lo que no muestra; eliminaciones en el cierre; orden de llegada a la base.

## Proposed shape

**Cierre (`flow-close`):**
- Paso 2 (`support-material.md`): la decisión se pregunta como hoy. Las eliminaciones se aplican en el cierre; commit, archive y promote se registran como pendientes, sin rama. Los ítems de un worktree se mueven antes a la misma ruta del checkout principal, como hoy.
- Paso 4 (`hive-rule-close-hygiene` «Archive the change folder»): no archiva; escribe o actualiza el registro `archive-pending/<change-id>` (título `Archive waiting:`) (carpeta, PR, commit de merge, acciones de soporte) y el informe nombra el número de pendientes y la respuesta «haz higiene». Sin rama ni PR del cierre para esto, el paso 4 corre siempre de inmediato.
- `hive-rule-close-modes`: el paso 4 pasa a «Runs now» en la fila del punto de cierre; la fila de cierre inmediato deja de depender de él.

**Higiene (`flow-close/references/batch-archive.md`, nueva), leída en una petición de higiene:**
1. Lista los change folders que siguen en `<Specs>/changes/` en la base y los cruza con los registros `archive-pending/*`: uno sin registro se lista con su origen, para que el usuario confirme que su trabajo está integrado.
2. Corre `check_change.py --archive` en cada uno; las tareas sin marcar se preguntan en una sola ronda, con su número por carpeta; bajo delegación desatendida se quedan pendientes.
3. Crea una rama desde la base y, en el orden en que cada carpeta llegó a la base, corre `apply_deltas.py <folder> --specs <Specs> --detail "<PR, merge commit>"` y `--apply`.
4. Aplica las acciones de soporte del registro sobre los ítems que siguen existiendo.
5. Commitea, entrega como dice el modo de entrega vigente para ese repositorio (un PR), y quita del registro lo entregado.

**Plan (`flow-plan` paso 1):** lee también los deltas de los change folders pendientes de las capabilities que toca.

**`git-workflow` paso 4:** las ediciones posteriores al merge de un change folder se commitean con el lote de higiene, no con «the close's archive».

## Impact

- Sin cambios de código ni de contrato de scripts.
- El ciclo de cada trabajo lee menos: la sección de archivado de `hive-rule-close-hygiene` se reduce a registrar lo pendiente.
- `specs/` va detrás del código entre lotes; `spec_deltas_test.go` no se ve afectado, porque solo compara carpetas archivadas.
- Un change folder sin archivar en la base pasa a ser el estado normal entre lotes, no una anomalía.
