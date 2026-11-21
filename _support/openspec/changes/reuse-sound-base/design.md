# Diseño

## Enfoques

Un solo camino sensato: el usuario eligió el enfoque en la investigación (Reply A). Se mantiene la escalera de reutilización, porque también evita el fallo contrario (escribir otra versión de algo que existe), y se endurece el peldaño 2. Los otros dos se descartaron allí:

- Adoptar el «Redesign from First Principles» de pstack: sin límite ni prueba observable, choca con el criterio **a** de `questions` (tocar otro módulo sin pedirlo).
- Dejar la excepción actual: solo cubre el workaround marcado y no da ruta para sanear.

## Forma propuesta

### `content/skills/hive-core/references/reuse-before-writing.md`, peldaño 2

Reemplaza su primera frase (la que termina en «name it in the report.») por este texto, y conserva intacta la frase «Prefer deleting to adding: …»:

> 2. The codebase already has it, such as a helper: use it as the surrounding code does when it is a sound base. It is not when (a) it is a workaround, such as one a comment marks or a guard [defect fix](defect-fix.md) rejects as a fix; (b) it breaks `CODING_STANDARDS.md`, lint rules or a documented convention; (c) it needs a boolean flag or branch only the new code uses; or (d) the same logic lives elsewhere, diverging. Name it and its condition in the report, and do not copy or extend it, except a workaround whose comment names its cause, which defect fix step 4 already reported: build on it, naming its ticket or proposing one when none exists. Except for (d), fix the base first, within the paths a child's brief names, when that is a small understood change, pinning its callers' behavior with a [tests](tests.md) characterization test; that fix is part of the task, not a module touched against the request. Otherwise a plan offers it as an approach, and a build, or a child to its parent, reports it before building on it, for that step's choice for a cause; building on it once the user accepts follows that step's workaround handling.

Tras las dos revisiones de código y la decisión D2 A mide 1148 B frente a los 270 B de la frase que reemplaza (paquete 22.488 B). Límite: el paquete de `flow-build` queda en 22.500 B o menos (antes 21.610 B). Se puede acortar la redacción sin quitar ninguna de las cuatro condiciones ni de las rutas: sanear primero, enfoque del plan, reporte del build o del hijo, y construir encima marcado si el usuario lo acepta.

La frase «that fix is part of the task, not a module touched against the request» es la excepción al criterio **a** de `questions`, enunciada en su punto de uso.

### `content/agents/review/hive-review-code.md`

- Paso 2: «and that new code does not copy a workaround» → «and that new code does not copy or extend a base that [reuse before writing](skill:hive-core/references/reuse-before-writing.md) does not count as sound»; y la exención que sigue, «a fix at the cause, and the other instances of the fault that reference has the author fix, is not a change the task did not ask for», pasa a incluir «or a base that reuse before writing has the author fix first».
- Paso 3: «For a fix placed at the symptom or a copied workaround, the failure scenario is the bad state the cause still produces for its other consumers, or the workaround's spread.» → «… or a copied or extended unsound base, … or that base's spread.»

### `tests/content/single_home_test.go`

Añade la frase «It is not when (a) it is a workaround» con su hogar `content/skills/hive-core/references/reuse-before-writing.md`, para que las condiciones no se copien a otro archivo.

### `tests/content/reading_footprint_test.go`

Por D1 A: una constante `reuseSoundBaseAllowance` con su comentario (el cambio, la fecha y la decisión D1 A), sumada a `allowance`. Su valor es el menor que hace pasar el test en los dos órdenes de integración: se mide con `-v` en la rama del build creada desde `development` y, después del commit de T1, en un worktree temporal bajo `.claude/worktrees/` con esa rama fusionada con `feat/close-cleans-at-completion` sin commit, que se quita con `git worktree remove` tras medir; se toma el mayor. El comentario lo dice. Si `feat/close-cleans-at-completion` cambia otra vez antes del merge, se vuelve a medir, y el cuerpo del PR avisa que la rama que se integre segunda corre otra vez `go test ./tests/content/`.

### `site/src/content/docs/flows/build.md`, línea 26

Reemplaza la frase «It does not copy an existing workaround, such as a guard a comment marks as temporary, and names it in its report instead.» por una sola frase de comportamiento, sin listar las condiciones (viven en la hoja):

> It reuses only code that is sound: when what it would reuse is a workaround or works against your standards, it fixes that first when the fix is small, and otherwise asks you before building on it.

## Impacto

La hoja la leen `flow-build` (paquete de etapa), `flow-plan` paso 2, los roles de build y los playbooks `feature` y `refactor`. El cambio puede hacer que un build sin plan se detenga más a menudo en repositorios con mucha deuda (condición b); el límite «cambio pequeño y entendido» y el reporte antes de construir lo acotan. No toca producción ni datos.

## Recuperación

Revertir el commit restaura el texto anterior; no hay efectos fuera del repositorio. Tras integrar, `hive update` despliega el contenido; una reversión necesita otro `hive update`.
