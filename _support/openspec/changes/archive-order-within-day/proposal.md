# archive-order-within-day — Orden de los cambios archivados el mismo día

| Campo | Valor |
| --- | --- |
| Estado | En build |
| Tracker | Ninguno |
| Git | tricell-hive · base `development` en `94ed2264` · rama propuesta `fix/archive-order-within-day` |

## Decisiones

2026-11-04
- **D6 A** — ¿Cómo se fija el orden de dos cambios archivados el mismo día? Con la hora en el nombre: las carpetas nuevas se llaman `YYYY-MM-DD-HHMM-<change-id>`, con la fecha y la hora del archivo. El test ordena de forma exacta cuando las dos carpetas tienen hora.
- **D7 A** — ¿Se renombran carpetas ya archivadas? Solo el par en conflicto del 2026-11-04, con la hora de su commit de archivo: `viewer-session-lifecycle-and-autoarchive` (14:03) y `viewer-local-actions-cli` (20:03). Las otras 125 conservan su nombre y la tolerancia del mismo día.
- **D8 A** — ¿En qué zona horaria van la fecha y la hora? En UTC (`date -u +%Y-%m-%d-%H%M`), para que el orden sea el mismo en cualquier host. Un archivo hecho después de las 18:00 hora local lleva la fecha del día siguiente. El par de D7 queda como `2026-11-04-2003-viewer-session-lifecycle-and-autoarchive` y `2026-11-05-0203-viewer-local-actions-cli`.
- El usuario eligió corregir H25 en el cierre de ux-review-loop; H29–H36 entran en el plan de H25, como se recomendó.

## Solicitud

El usuario, el 2026-11-04: «/flow-plan convención de orden del archivo de cambios (H25) y ajustes del test de deltas (H29–H36)».

Lo que queda abierto, según `pending/ux-review-loop` en Engram:

- **H25:** los nombres `YYYY-MM-DD-<id>` no dan el orden dentro de un día, y `TestSpecDeltasApplied` acepta cualquier delta del mismo día.
- **H29:** el primer bloque de una carpeta con un duplicado entra en `SameDay` y oculta el error de otra carpeta del mismo día.
- **H30:** los subtests no prueban que una carpeta posterior se compare después de un duplicado.
- **H31:** un duplicado en un archivo ya superado deja el test en rojo, sin forma de corregirlo.
- **H32:** el bloque que conserva un delta duplicado depende del orden del archivo.
- **H33:** el comentario de `checkSpecDeltas` está desactualizado.
- **H34:** la detección de duplicados está mezclada con la recolección.
- **H35:** el bucle de coincidencias es poco legible.
- **H36:** hay una copia de slice innecesaria.

## Objetivo

Hoy dos cambios archivados el mismo día que modifican un requisito pasan el test aunque la especificación conserve el texto del que se archivó primero. Al terminar, los archivos nuevos se ordenan por su hora, el par del 2026-11-04 vuelve al chequeo estricto, y el test reporta los duplicados solo cuando importan, sin dejar que oculten otro error.

## Alcance y aceptación

**Dentro:**

- `flow-close` paso 4;
- `tests/content/spec_deltas_test.go`;
- el renombre de dos carpetas archivadas;
- el `CHANGELOG`;
- un delta de `workflow-lifecycle`.

**Fuera:**

- renombrar las otras 125 carpetas;
- `hive doctor`: su patrón `^\d{4}-\d{2}-\d{2}-[a-z0-9]+(-[a-z0-9]+)*$` ya acepta `YYYY-MM-DD-HHMM-<id>`;
- las carpetas de `sessions/`, `workspace/` y `evidence/`.

- AC1 — `flow-close` paso 4 mueve la carpeta a `<Specs>/changes/archive/YYYY-MM-DD-HHMM-<change-id>/`, con la fecha y la hora en UTC (24 horas) del momento del archivo. · Hoy es falso porque `content/skills/flow-close/SKILL.md:15` dice `YYYY-MM-DD-<change-id>`.
- AC2 — Cuando dos carpetas del mismo día modifican un requisito y las dos llevan hora, el test exige el delta de la hora más tardía: una especificación igual al delta más temprano da `DIFFERS`. · Hoy es falso porque la tolerancia del mismo día acepta cualquiera de los dos.
- AC3 — El par del 2026-11-04 se llama `2026-11-04-2003-viewer-session-lifecycle-and-autoarchive` y `2026-11-05-0203-viewer-local-actions-cli` (horas UTC de sus commits de archivo), y el test real exige que «Local viewer management actor» coincida con el delta de `viewer-local-actions-cli`. · Hoy es falso porque esas carpetas no llevan hora y el test acepta el delta de las 14:03.
- AC4 — Un requisito que un delta archivado lista más de una vez se reporta solo mientras esa carpeta tenga el delta más reciente de ese requisito. Lo supera cualquier carpeta que ordene después, también una del mismo día sin hora, y entonces deja de reportarse. Un cambio activo no lo supera. · Hoy es falso porque se reporta siempre (H31).
- AC5 — Los bloques de un requisito duplicado no participan en ninguna comparación. Una carpeta posterior del mismo día que no coincide con la especificación se reporta, aunque el primer bloque duplicado sí coincidiera. · Hoy es falso porque ese bloque entra en `SameDay` y oculta el error (H29 y H32).

Dos carpetas del mismo día en las que al menos una no tiene hora siguen aceptando cualquiera de sus deltas. Esto ya es así hoy y se conserva como regresión en el test, no como criterio.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- **Choques del mismo día:** hay 9 requisitos tocados por dos o más archivos del mismo día. El único en el que el delta más reciente difiere y que es el último de su requisito es «Local viewer management actor», con `bb61f71f` (20:03) después de `bdb2e8f4` (14:03) el 2026-11-04.
- **Qué lee o escribe el nombre:**
  - `tooling/cli/doctor_project_changes.go:19`, sin cambio;
  - `content/skills/flow-close/SKILL.md:15`;
  - `tests/content/spec_deltas_test.go`;
  - `_support/docs/architecture/deployment-manager.md:77`, que describe el patrón de `doctor` y sigue siendo cierto.
- **Referencias a las dos carpetas que se renombran:** ninguna en el repositorio. La coincidencia en el `tasks.md` archivado es una ruta de `_support/workspace/`. `site/` no describe el formato del nombre.
- **Copia de la regla:** `content/skills/workspace-conventions/references/artifact-layout.md:34` («archived changes, which use the closing date»). Pasa a ser un puntero en T1.
- **Entrada del `CHANGELOG`:** la de `CHANGELOG.md:63–66`, sobre la tolerancia del mismo día y los duplicados, queda inexacta. Se reescribe en su sitio.
- **Incidental:** el comentario de la plantilla `proposal.md` menciona un «archive helper» que no existe. Queda fuera de este plan; se lista al cierre.

## Entrega

- **Entrega:** automática (D4 A, en vigor en esta sesión).
- **Revisión:** Claude Code `/code-review` (D5 A).
- **No autoriza:** una release ni escribir en trackers.

## Estado de revisión y avance

2026-11-04 — Plan escrito. La revisión del plan (`hive-review-plan`) encontró 9 huecos, aplicados como detalla `tasks.md`; la zona horaria se resolvió con D8 A.
