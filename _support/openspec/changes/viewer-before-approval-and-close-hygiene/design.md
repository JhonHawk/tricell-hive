# Diseño

## Enfoques

**Visor (D1).** Se presentaron tres enfoques y se eligió A.
- **A, la elegida: opción dentro de la propia pregunta y volver a preguntar** (v0.6.0). El usuario controla si se abre el visor y no se agrega ninguna tarjeta. Exige una excepción, declarada en su punto de uso, a la regla de no repetir una pregunta.
- **B: abrir el visor sin preguntar.** Rompe «nunca abrirlo por tu cuenta».
- **C: una ronda previa solo para el visor.** Agrega una tarjeta.

**Restos de un trabajo con el cierre pendiente (D5).** Se eligió A: se limpian cuando el registro tiene más de 24 horas. B (solo reportar) obliga a responder; C (limpiar siempre) puede romper una sesión paralela.

**Pausa en un repo de specs sin `direct-base` (D6).** Se eligió A: queda pendiente y se recomienda declarar `direct-base`. B abre rama y PR para un plan pausado; C no orienta.

**Momento de la limpieza (D2).** Se presentaron dos enfoques y se eligió A.
- **A, la elegida: mantener la espera al pie del cierre (#311) y añadir el barrido de ramas ya mergeadas.** #311 nació de observar que en 12 de 21 cierres el usuario siguió trabajando en la misma rama. El barrido limpia en el cierre siguiente lo que haya quedado.
- **B: limpiar en cuanto se hace el merge**, como en legacy. Revierte parte de #311.

Los demás puntos tienen un solo camino razonable:
- **Repos de specs `direct-base` en su checkout principal:** así lo hacía legacy, y `direct-base` ya autoriza el push a la base.
- **Commit al pausar:** es la única forma de no perder un plan revisado.
- **Recorrido de los repos que tocó el trabajo:** sin él, los restos de specs no se ven.
- **Promociones:** decisión del usuario.

## Forma propuesta

El texto se escribe en inglés en `content/` y en español en el sitio. Las frases de abajo son el contenido acordado; la redacción final cabe en los límites de tamaño.

### T1 — La opción de abrir el visor (`viewer-publication.md` «Open the viewer»)

Reemplaza «add "Open the viewer?" as one more question of that round» por esta regla:
- **En la tarjeta:** cuando la parada pregunta con la herramienta de preguntas del host, la pregunta propia de esa parada (aprobar, aceptar o validar) lleva una opción más, «Open the viewer». No es una pregunta aparte, ni tiene ID propio, ni otra llamada a la herramienta.
- **Al elegirla:**
  - se ejecuta `hive view open` una sola vez, con `--session-id ID` si hay varias sesiones abiertas;
  - se vuelve a hacer la misma pregunta, sin cambios y con el mismo ID, pero sin la opción del visor;
  - las respuestas a las demás preguntas de esa ronda se mantienen.
- **Excepción declarada:** elegir la opción no responde la decisión. En ese punto se nombran los elementos de «Questions and routes» que cambian:
  - la opción del visor nunca va primera ni lleva ★, así que «ok» nunca abre el visor;
  - volver a preguntar con el mismo ID es la misma decisión, no un ID reutilizado;
  - el mensaje de la segunda vez es una sola línea («El visor está abierto; D7 otra vez») seguida de la misma tarjeta, sin repetir el texto de la ronda;
  - si la pregunta ya usa todas las opciones que admite la herramienta, la oferta pasa a ser la línea de respuesta.
- **Sin tarjeta, o con la ronda ya preguntada:** se mantiene la línea de respuesta «or reply *open the viewer*»; si el usuario la usa, se abre el visor y la parada sigue esperando su respuesta.
- **Sin cambios:** una vez abierto, no se vuelve a ofrecer en la conversación; si el usuario no lo abre, se ofrece de nuevo en la siguiente parada.

Los punteros cambian «open-viewer question» por «open-viewer option»: `flow-build/SKILL.md:15`, `flow-verify/SKILL.md:15` y `flow-walkthrough/SKILL.md:10`. `plan-handoff.md:21` ya habla de una línea de respuesta y no cambia. El ancla `#open-the-viewer` se mantiene.

### T2 — Limpieza del cierre (`flow-close`, `references/branch-cleanup.md`, `global.md:31`)

- **Hacer espacio.** La sección «Branch cleanup» pasa a `content/skills/flow-close/references/branch-cleanup.md`. El paso 5 enlaza a esa referencia y dice cuándo leerla.
- **Contenido de la referencia:**
  - Los pasos 1 a 4 actuales, sin cambios.
  - **Alcance:** los repositorios donde el trabajo creó una rama, un worktree o un commit, o escribió su change folder, más el de `Specs`. Para una higiene que pide el usuario: el repositorio actual, su `Specs` y los repositorios que nombran los registros de cierre pendiente en la memoria.
  - **Barrido**, después de la rama del trabajo, en cada uno de esos repositorios. Revisa:
    - las ramas mergeadas en la base cuyo pull request creó el usuario autenticado (`gh api user`) y cuyo `headRefOid` coincide con la cabeza local;
    - en un repo sin pull requests, las ramas locales que `git branch --merged <base>` lista y que tienen al menos un commit propio que ya está en la base;
    - las ramas temporales o de preview que creó el trabajo.
  - **El barrido nunca cambia la rama de un checkout.** Es una excepción, nombrada en la referencia, a «never switch, reset or delete its branch» del paso 2 de `git-workflow`. Solo quita worktrees limpios y borra ramas que no tiene activas ningún otro checkout.
  - **Se conserva y se reporta con su motivo:**
    - una rama con commits sin merge;
    - una rama sin commits propios fuera de la base, como una recién creada;
    - una rama activa en un checkout que no es un worktree limpio que se va a quitar;
    - un worktree con archivos sin commit o sin trackear;
    - un worktree con archivos ignorados fuera de las carpetas de dependencias o de build, porque `git worktree remove` borra los ignorados;
    - un worktree con un proceso cuyo directorio de trabajo está dentro (`lsof -d cwd`). Si `lsof` no está disponible, también se conserva;
    - un worktree o una rama que la memoria registra para un trabajo con el cierre pendiente desde hace menos de 24 horas (D5 A). Con más de 24 horas se limpia y su registro pendiente se actualiza como fija «Pending work».
  - **Reporte:** las ramas revisadas, las borradas y las conservadas con su motivo.
- **Promociones** (en `flow-close`, junto a «Pending work»): una promoción no hecha nunca es motivo para conservar una rama, un worktree, un proceso ni material de soporte. Solo una tarea del plan que promueve a un entorno donde el ticket o la solicitud piden el cambio queda como registro de trabajo pendiente, como ya fija `flow-close:41` («a task of its plan runs after the merge»).
- **Pausa fuera de `hold`** (`flow-close`): una frase define el cierre de pausa de un trabajo que no está en `hold`, por ejemplo un plan pausado antes de construir. Pausa el visor, lista el material de soporte sin actuar, detiene los procesos que inició, commitea el change folder como fija el paso 4 de `git-workflow`, y reporta y registra lo pendiente. Va sin pie de cierre.
- **`global.md:31`:** «deleting merged work branches and their clean worktrees» pasa a cubrir las ramas de este trabajo o de trabajos anteriores del usuario, en cada repositorio que tocó el trabajo. No se toca nada más: las prohibiciones sobre ramas sin merge y worktrees con cambios se mantienen.
- **Ancla:** el enlace de `git-workflow:12` a `../flow-close/SKILL.md#branch-cleanup` pasa a apuntar a `../flow-close/references/branch-cleanup.md`. Lo hace T3. El validador de referencias ignora los `#fragmentos`, así que se comprueba con `rg`.

### T3 — Specs y promociones en el plan (`git-workflow` pasos 3 y 4, `flow-plan`)

- **Paso 3.** Si `Specs` está en otro repositorio que declara `Delivery: direct-base`, el change folder se edita en el checkout principal de ese repositorio, sobre su base, sin rama ni worktree, aunque otra sesión escriba ahí. Es una excepción, nombrada en ese punto, a la regla del worktree del paso 2. En ese caso:
  - se commitean solo las rutas del folder con `git commit --only -- <paths>`;
  - antes del push, `git log origin/<base>..HEAD` y la pregunta de «Delivery» sobre commits de otro trabajo;
  - si el push es rechazado, se aplica la decisión del paso 2, nunca un rebase sobre un árbol con cambios.

  Sin `direct-base`, ese repositorio sigue las reglas generales del paso.
- **Paso 4,** en la viñeta del change folder, para un cierre de pausa:
  - en un repo de specs `direct-base`, se commitea el folder y se hace push a su base, con las mismas comprobaciones;
  - con `Specs` en el mismo repositorio y la rama del trabajo ya creada, se commitea en esa rama sin push;
  - con `Specs` en el mismo repositorio y sin rama todavía (pausado antes de construir), el folder sigue sin commit y se registra como pendiente con su ruta;
  - en un repo de specs separado sin `direct-base`, el folder sigue sin commit, se registra como pendiente con su ruta, y el reporte recomienda declarar `Delivery: direct-base` en ese repo (D6 A);
  - bajo `hold`, el folder sigue sin commit.
- **`flow-plan`.** Cuando el ticket o la solicitud piden el cambio en un entorno nombrado, el plan incluye la promoción a ese entorno como tarea, que corre con `flow-deploy`. Esa petición es la instrucción de despliegue solo para ese entorno; producción sigue llevando su propia pregunta, como fija `flow-deploy`. Sin esa petición, el plan no trae promociones: se agrupan y corren cuando el usuario las pide.

## Impacto

Lo encontró el hijo `hive-research` en `6670e799`.

- **Visor:** las definiciones y los punteros de arriba, y `_support/openspec/specs/walkthrough-viewer/spec.md:134-147`. Ningún test ni detector valida esa redacción.
- **Cierre:**
  - `global.md:31`;
  - `site/src/content/docs/flows/close.md` (líneas 10, 22, 27 y 30);
  - `site/src/content/docs/flows/plan.md:45`;
  - `_support/openspec/specs/workflow-lifecycle/spec.md:11`, que sigue siendo cierto con D2 A;
  - `README.md:375-376` y `llms.txt:37` cambian: dicen «task-owned resources», y el barrido también quita ramas de trabajos anteriores;
  - `site/src/content/docs/flows/close.md:3` («clean up only what that work created») y `:56` («The plan stays where it is») cambian;
  - `site/src/content/docs/viewer.md:8` sigue siendo cierto.
- **Tamaño:** `flow-close/SKILL.md` está a 67 bytes de su tope. Por eso T2 mueve antes «Branch cleanup» a la referencia.
- **Pilotos:** los detectores de `tests/pilot/` (`closeSequenceBranchAbsent`, `closeSequenceCleanupLabel`) siguen válidos. Los pilotos están pausados y no se tocan.
- **Fuera de alcance:** los enlaces rotos de `_support/docs/architecture/workspace-and-artifacts.md:23,25`, que es un hallazgo aparte.

## Recuperación

Revertir el merge deshace todo el cambio, porque solo toca texto distribuido. Las sesiones abiertas conservan la guía que cargaron hasta que se reinician. Un barrido que ya corrió con la regla nueva solo borra ramas mergeadas y worktrees limpios, que se pueden recrear desde la base.
