# Diseño

## Decisiones

- **Comportamiento de `ready` (D5 A).** Opciones: igual que `automatic` hasta el PR y parada antes del merge (elegida), o dos paradas (validación antes del push y parada antes del merge). Se eligió la primera porque la issue pide un PR listo para el merge del usuario y la segunda repite la validación que el usuario hace en el PR.
- **Bots que solo comentan (D6 A).** Opciones: no frenan el merge (elegida) o lo frenan. Hive prohíbe sondear, y un bot sin check no tiene una espera bloqueante; un bot lento o caído no debe bloquear `automatic`. Los bots que publican un check ya quedan dentro de la espera del CI.
- **Límite de rondas después del push.** Los hallazgos de bots y de checks posteriores al push usan la regla de una ronda del paso 3 de `flow-verify`; no se crea un límite nuevo. Así un bot no puede abrir un ciclo de push, CI y hallazgos sin fin, y cada corrección sin revisión dedicada queda marcada con `Unreviewed:`, que `flow-deploy` ya revisa al promover.
- **Frescura solo para bots y checks.** Un hilo humano abierto sigue siendo un pedido vigente aunque sea anterior al último push; filtrarlo por fecha rompería el paso 7 actual.
- **Hogar canónico de cada regla.**
  - Los efectos de cada modo viven solo en la guía global («Delivery»).
  - Cómo se leen y atienden revisores humanos y bots, y su límite de rondas, vive en `git-workflow` paso 7.
  - La condición de base al día, y que un bot pendiente no frena el merge, viven en el paso 8.
  - La lectura y la espera por REST viven en el paso 9.
  - La automatización que haría merge vive en el paso 6.
  - `flow-verify` paso 6 solo dice dónde termina cada modo y apunta a esos pasos.
  - El reporte de la parada vive en el reporte final de `flow-close`, no en `flow-verify`.
- **Punto de cierre y visor.** `flow-close` ya define un PR abierto que espera a una persona como punto de cierre; `ready` se añade a esa misma frase, y su paso 1 deja la sesión del visor en `in_review` en ese caso, un estado que el contrato del visor ya tiene.

## Forma propuesta

Texto aproximado de cada regla, en inglés como la guía; el build ajusta la redacción respetando el tamaño.

- **`global.md` «Delivery»**, viñeta nueva tras `automatic`: «`ready`: review, push, open the pull request as non-draft, wait for CI and handle the reviews as [git-workflow](…) sets, then stop with the pull request open and unmerged for the user to merge; never merge, enable auto-merge or add it to a merge queue.» En la recomendación de revisión: «always under `automatic` or `ready`».
- **`git-workflow` paso 6:** la lista de automatización que se lee antes de una acción pasa a «CI, review, preview, deployment or merge»; bajo `ready`, una automatización que haría merge del PR al abrirlo o al quedar verde se pregunta antes de abrirlo.
- **`git-workflow` paso 7** («Answer reviewers»):
  - Cuándo se lee: tras la espera del CI, leer una vez las revisiones, comentarios y checks del PR.
  - Qué cuenta como bot: una cuenta de tipo `Bot` o con login terminado en `[bot]`.
  - Qué hallazgos de bot cuentan: una revisión o comentario de revisión cuyo `commit_id` es el commit de cabeza, o un comentario creado después de la hora del último push de este trabajo, que el agente anota al hacer push. Lo demás de bots y los checks de commits anteriores no provocan cambios.
  - Los hilos humanos: un hilo humano abierto sigue el texto actual, sin filtro de fecha.
  - Cómo se atiende un hallazgo de bot: se contrasta con el código antes de cambiar nada. Uno descartado recibe respuesta con la razón. Un check de bot que falla y no es requerido es un hallazgo.
  - Rondas: los hallazgos de bots y checks después del push siguen el paso 3 de `flow-verify`. Hay una ronda solo para lo bloqueante, y cada commit de corrección lleva `Unreviewed: <reason>`. Lo no bloqueante va a los hallazgos del cierre, y un hallazgo sobre un commit de corrección se reporta sin otra ronda.
  - Commits ajenos: un commit en la rama remota del PR que no hizo este trabajo, como el autofix de un bot, se nombra y se pregunta como fija «Delivery» para commits ajenos, y no cuenta como push de este trabajo.
- **`git-workflow` paso 8:**
  - La viñeta de base al día pasa a valer para «merge, or stop a `ready` pull request». Bajo `ready`, el push que trae la base lo cubre la autorización de entrega, y cuando la base vuelve a moverse, «report the pull request as not ready».
  - Se añade: «a review bot that has not reviewed the head commit does not hold the merge; name it as pending».
  - El método de merge, la cola de merge y el merge admin no se aplican a `ready`.
- **`git-workflow` paso 9:** cuando `gh` sale con el error de cuota de GraphQL («API rate limit exceeded» para GraphQL), hay que leer por REST:
  - el PR: `gh api repos/<owner>/<repo>/pulls/<n>`;
  - los checks: `.../commits/<sha>/check-runs`;
  - las revisiones y los comentarios: `.../pulls/<n>/reviews`, `.../pulls/<n>/comments` y `.../issues/<n>/comments`;
  - y esperar el CI con `gh run watch <run-id> --exit-status`.
  
  Nunca se consulta en bucle.
- **`flow-verify` paso 6:** «… and merge; under `ready`, stop instead of merging once git-workflow step 8's current-base check passes, and run `flow-close`.» Las revisiones tras el CI apuntan a `git-workflow` paso 7.
- **`flow-close`:**
  - El punto de cierre: «When the delivery is `ready`, or a human reviewer is assigned to the pull request, opening it is the completion point…».
  - Paso 1: «…and `in_review` when the pull request waits for the user's merge or a human reviewer».
  - Reporte final: en **Integrated**, el enlace del PR, el commit de cabeza, el commit de base comprobado y el resultado del CI. En **Next**, el merge que queda al usuario y los bots pendientes.

## Impacto

- **Dónde se leen los modos:** en `flow-plan` paso 4, en `flow-build` antes de la primera edición, en `unattended-delegation` paso 4 y en `flow-verify` paso 6. Solo cambia este último.
- **Lo que no cambia:**
  - `flow-deploy` usa su propia elección de entrega.
  - `tooling/cli` rechaza `Delivery: ready` como valor de `## Hive`.
  - La etapa Delivery del visor ya prohíbe que su espera haga merge.
- **Alcance de los efectos:** no toca producción ni datos guardados.

## Recuperación

Revertir el merge del PR deshace todo: son cambios de texto en la guía distribuida. Las instalaciones que ya recibieron la guía con `hive update` vuelven al texto anterior con el siguiente `hive update`.
