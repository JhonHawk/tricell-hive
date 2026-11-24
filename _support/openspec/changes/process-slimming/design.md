# Design

## Approaches

- **Registro del plan (D18 A, elegido).** Un `plan.md` local con orden fijo, como el archivo de tareas de un proyecto de referencia, con el progreso fuera de git como en otro. El cuerpo del PR es el registro público. Descartados:
  - la carpeta liviana, que conserva parsers y ruido;
  - un `plan.md` commiteado y borrado al cerrar, que suma un archivo y un commit por trabajo;
  - el plan solo en la conversación, que pierde la reanudación y deja a los hijos sin archivo.
- **Entrega (D19 A, D21 A).** Tres PRs en orden de costo, más un cuarto para la migración de `project-layout`; cada uno se revierte solo.
- **Fixtures (D20 A).** Casos dentro de la tabla del test, más la comparación sobre sesiones reales. Descartados: congelarlas, que deja el material publicado, y borrarlas sin migrar, que deja el auditor sin tests propios.

Comparación de tres formas, el 2026-11-24:
- Un proyecto de referencia: sin specs, plan corto solo para trabajo de varios PRs.
- Hive antes del 09-23: plan en la conversación.
- Hive hoy: carpeta de 4 archivos y archivo permanente.

## Proposed shape

`<support>/workspace/<YYYY-MM-DD>-<trabajo>/<YYYY-MM-DD>-<trabajo>.plan.md`, sin commitear, en este orden; la carpeta guarda además los registros `<YYYY-MM-DD>-<trabajo>.<tipo>.<ext>` y las subcarpetas `images/` y `evidence/` (D27 A):

1. `# <id> — <título>` y una tabla con estado, tracker, git y playbook.
2. `## Decisiones`, `## Business rules`.
3. `## Solicitud`, `## Objetivo`, `## Alcance y aceptación` con líneas `AC<n> — … · false today because … · Observed in: …`.
4. `## Prerrequisitos`, `## Contexto verificado`, `## Effects`, `## Entrega`.
5. `## Diseño`, solo con decisiones de diseño: enfoques, UI y copy, impacto, recuperación.
6. `## Tareas`: bloques `### T<n> — …` con caja, `**Closes:**`, `**Depends on:**`, `**Locations:**`, `**Execution:**`, `**Test approach:**`, `**Changes:**` y `**Verification:**`. Después `## Delivery units`, si aplica.
7. `## Checkpoint`, que escribe `tasks.py`, y el enlace del registro de ejecución.
8. `## Review status and progress` al final, con el encabezado que busca `tasks.py done`: una línea por evento.

`## Effects` y `## Business rules` conservan su nombre en inglés, porque las reglas los buscan. Un hijo recibe `ruta absoluta del plan · T<n> · AC<n>…` y lee hasta `## Checkpoint`. Para un sandbox sbx, el plan se copia al workspace ignorado del worktree.

## Impact

- Los proyectos que ya tienen `<Specs>/changes/` conservan sus carpetas hasta que su usuario pide la migración de `project-layout` (D21 A, PR 4). La guía nueva deja de crearlas y de archivarlas.
- `hive doctor` deja de advertir sobre carpetas, y `Specs` sin declarar deja de ser un error.
- Desaparecen `TestSpecDeltasApplied`, `apply_deltas.py` y sus tests, el lote de higiene y los registros `archive-pending/`.

## Recovery

Cada PR se revierte por separado. El archivo borrado vuelve desde git.
