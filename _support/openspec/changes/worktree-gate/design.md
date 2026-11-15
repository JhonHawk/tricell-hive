# Diseño

## Approaches

Se presentaron tres formas de podar las reglas de worktree, todas con la pregunta antes de crear uno:

- **A — Pregunta + ruta nativa, sin registro en memoria (elegida, D1 A).** Quita los tres criterios, la alternativa de ubicación y el registro obligatorio en memoria; el worktree se nombra en el plan o checkpoint de la tarea. Mantiene las protecciones contra pérdida de archivos. Coste: otro host no encuentra en la memoria un worktree de una sesión cortada, solo en el plan de la tarea.
- **B — Pregunta + ruta nativa, con registro en memoria.** Permite retomar desde otro host, pero sigue produciendo los registros que funcionan como precedente de ruta.
- **C — Poda máxima.** Solo la pregunta y "nunca forzar el borrado". Quita también las reglas del checkout principal y las excepciones de archivos ignorados: `git worktree remove` borraría sin avisar archivos ignorados únicos, como un `.env`.

Se elige A porque la fuente de ruido observada son los precedentes de memoria y la ubicación ambigua, mientras que las protecciones de C evitan pérdidas de datos reales.

## Decisions

- **D1 A** — reglas de worktree: ver arriba.
- **D2 A** — memorias: se borran solo las que tratan de un worktree, fuera de `tricell-hive` y `hive-machine`, con respaldo. Se descartaron B (unos 650 registros de Engram, con pérdida de decisiones de proyecto que solo mencionan un worktree de paso) y C (solo hosts, con los 68 registros de Engram como precedente).
- **Ampliación** — las memorias por proyecto de cada host entran en D2 A.
- **D5 A** — sustituye el criterio de D2 A: lo que ensucia el contexto es la ruta del worktree. Se borran los registros que tratan del worktree mismo (unos 55). A los demás que la citan (unos 205) se les quita solo la ruta. El usuario valida la lista antes. Se descartaron borrar los 260, porque se perderían 70 decisiones y 132 resúmenes, y borrar solo los 55, porque las rutas seguirían en las decisiones.
- **D6 A** — los hijos nunca crean worktrees y comparan con `git diff` o `git show`. Se descartó una excepción para worktrees de comparación porque mantenía worktrees que el usuario no pidió.
- **D8** (reemplaza a D7 A) — en trabajo desatendido el agente decide sin preguntar, como fija el paso 5 de `unattended-delegation`, y elige lo que no detiene el trabajo.

## Proposed shape

- **Dónde vive cada regla.**
  - La pregunta y la ruta nativa viven en `git-workflow` paso 3, su hogar canónico.
  - `delegation.md` solo apunta a ella para los hijos aislados en un worktree.
  - La regla del directorio de trabajo vive en `global.md` como invariante mínima, porque toda sesión ejecuta comandos sin cargar `git-workflow`.
- **Formato de la pregunta.** Es una decisión con el formato de «Questions and routes». La `Situación:` dice qué se observó: cambios ajenos en el checkout, otra sesión escribiendo o la necesidad de dos checkouts. Los antiguos criterios pasan a ser lo que la `Situación:` informa, no una condición que el agente evalúe por su cuenta.
- **Sin pregunta.** No se pregunta cuando el usuario pidió el worktree, cuando la sesión ya trabaja en uno, ni para el repositorio de specs con `direct-base`, que nunca usa worktree. Con trabajo desatendido (D8, reemplaza a D7 A), el agente decide en el momento sin preguntar. Elige lo que no detiene el trabajo y lo reporta con su motivo.
- **Hijos (D6 A).** Ningún hijo crea un worktree: compara con la base mediante `git diff` o `git show`. Se quita "or in a separate worktree" de los roles y de su test.

## Impact

- **Criterios de creación** (paso 3 de `git-workflow`): los usa solo ese paso. En `workspace-and-artifacts.md` línea 54 aparecen como historia y se actualizan.
- **Registro en memoria:**
  - `global.md` «Persistent memory» y `git-workflow` paso 3 lo piden.
  - `branch-cleanup.md` línea 38 lo da como ejemplo de fuente de ramas del usuario en repositorios sin PR.
  - El registro de cierre pendiente (`branch-cleanup.md` línea 44) es otro registro y se queda.
- **Fixtures:** el caso de calificación `question-worktree-no-prep` espera que no se ejecute `git worktree add` ante una pregunta del usuario. Con la pregunta antes de crear sigue siendo coherente.
- **Producción:** no la toca. El efecto llega a las sesiones nuevas tras `hive update`.

## Recovery

- **Guía:** se revierte con `git revert` del merge y `hive update`.
- **Engram:** un borrado recuperable se restaura con el respaldo JSON de T6.
- **Archivos de host:** se restauran desde su copia en `_support/workspace/2026-11-15-worktree-gate/memory-backup/`.
- **Codex:** la nota de borrado se quita del directorio de notas antes de que Codex la consolide; si ya lo hizo, se restaura con una nota inversa.
- **Lo que no se deshace:** lo que Grok vuelva a aprender de sesiones posteriores no depende de esta limpieza.
