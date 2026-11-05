# viewer-before-approval-and-close-hygiene — Ver el visor antes de aprobar e higiene de cierre

| Campo | Valor |
| --- | --- |
| Estado | Completed · entregado en el PR #381 (merge en `development` como `7ee0fbbc`); archivado al cerrar el 2026-11-05 |
| Tracker | ninguno: el trabajo nace de la revisión de sesiones del 2026-11-05 |
| Git | JhonHawk/tricell-hive · base `development` en `6670e799` · rama propuesta `fix/viewer-before-approval-and-close-hygiene` en el worktree `.claude/worktrees/viewer-close-hygiene` |

## Decisiones

2026-11-05
- **Corregir el visor** — ¿Se corrige el orden entre la aprobación y la oferta del visor? Sí: el usuario debe poder ver antes de decidir.
- **Higiene en el mismo plan** — ¿Entra la higiene de los repos de specs al mismo plan? Sí: entran los puntos de cuándo corre, alcance, repos de specs, plan pausado, repos del workspace y promociones.
- **Promociones** — ¿Una promoción no hecha es un pendiente? No. Los despliegues consumen recursos y se agrupan para hacerse juntos, salvo que el ticket o la solicitud pidan un entorno, y en ese caso el plan lo trae como tarea.
- **`Completion`** — ¿Se redefine `Completion`? No: lo define cada proyecto y es difícil generalizarlo. El plan no toca su definición.
- **D1 A** — ¿Cómo se deja ver el visor antes de decidir una aprobación? Con el patrón de v0.6.0: «Abrir el visor» es una opción dentro de la propia pregunta de aprobación. Al elegirla, el agente abre el visor y vuelve a hacer la misma pregunta, con el mismo ID y sin esa opción.
- **D3 A y D4 A** — ¿Cómo se entrega? En modo `automatic`, con la revisión nativa `/code-review` de Claude Code.
- **D5 A** — ¿Qué hace el barrido con el worktree mergeado de un trabajo cuyo cierre sigue pendiente en memoria? Lo limpia cuando el registro tiene más de 24 horas; antes, lo conserva y lo reporta.
- **D6 A** — ¿Qué pasa al pausar un plan en un repo de specs separado sin `direct-base`? El folder queda sin commit, se registra como pendiente y el reporte recomienda declarar `Delivery: direct-base`. Así, fac-235 se resuelve cuando facti-specs declare `direct-base`.
- **D2 A** — ¿Cuándo corre la limpieza de ramas y worktrees de un trabajo ya mergeado? Se mantiene la espera al pie del cierre que fijó #311. Se añaden el barrido de las otras ramas propias ya mergeadas (de v0.6.0) en cada cierre y en cada higiene, el recorrido de los repos del workspace, y que una promoción no justifique conservar nada.

## Solicitud

> Ayudame a analizar la siguientes premisas
> - La pregunta del visor sale después de preguntar si se aprueban los cambios, entonces los aceptas o deniegas y luego te pregunta si quieres verlos
> - No se esta haciendo higiene de de los repos de specs y se quedan en worktrees o sin commitear

Después el usuario eligió corregir las dos y pidió una propuesta para cada punto de la comparación con v0.6.0 y legacy. Precisó además: «una promocion no hecha no debe considerarse como un pendiente, el deploy consume recursos que se puede agrupar para hacer en conjunto, amenos que el ticket o solicitud indiquen una solicitud en un entorno y eso debe venir en el plan», y «Completion debe definirse segun el proyecto, no con esta reglas».

Fuera de alcance:
- El `openspec/` sin versionar de fleetsystem-specs, que es un hallazgo aparte.
- La definición de `Completion`.
- Los detectores de `tests/pilot/`: los pilotos están pausados y sus casos no se tocan.

## Objetivo

Hoy, en la parada de validación, «¿Abro el visor?» va como otra pregunta de la misma tarjeta (`viewer-publication.md:52`). El usuario tiene que responder la aprobación antes de poder ver. Lo observamos en Claude Code 5ff8bac1, 062c567a y 98476eca, y en Codex 01a1a00a y 01a1a30f.

La limpieza del cierre solo cubre la rama del trabajo actual y solo su repositorio. Un repo de specs `direct-base` recibe ramas y worktrees por ticket, que es el caso de cam-9 en cam-specs. Un plan pausado queda sin versionar, como fac-235 en facti-specs desde el 2026-10-29. Además, una promoción pendiente se usa como motivo para conservar worktrees («por el seguimiento de QA»).

Cuando el cambio esté hecho:
- El usuario abre el visor desde la propia pregunta de aprobación y después decide.
- Cada cierre e higiene limpia también las ramas propias ya mergeadas y sus worktrees limpios que quedaron de trabajos anteriores, en cada repo que tocó el trabajo.
- Los repos de specs `direct-base` se editan en su checkout principal, y un plan pausado queda versionado en ellos.
- En un repo de specs sin `direct-base`, el plan pausado queda registrado como pendiente y el cierre recomienda declararlo.
- Ninguna promoción no planeada retiene nada.

## Alcance y aceptación

Dentro del alcance:
- La regla de abrir el visor y sus punteros.
- La limpieza de ramas de `flow-close`, que pasa a una referencia propia.
- La autorización de limpieza de `global.md`.
- Los pasos 3 y 4 de `git-workflow`.
- La regla de tareas de promoción en `flow-plan`.
- Las páginas del sitio que describen esto, los deltas de spec y el CHANGELOG.

- **AC1** — La regla «Open the viewer» de `viewer-publication.md` pone la oferta como una opción dentro de la pregunta propia de la parada (aprobar o validar). Al elegirla, se ejecuta `hive view open` y se vuelve a hacer la misma pregunta, con el mismo ID y sin esa opción. La oferta nunca es una pregunta aparte de la ronda.
  Hoy es falso: la línea 52 dice «add "Open the viewer?" as one more question of that round».
- **AC2** — Ningún archivo de `content/` presenta la oferta del visor como una pregunta. `flow-build`, `flow-verify` y `flow-walkthrough` hablan de la opción de abrir el visor.
  Hoy es falso: los tres dicen «open-viewer question».
- **AC3** — La limpieza de ramas del cierre recorre los repositorios donde el trabajo creó una rama, un worktree o un commit, o escribió su change folder, más el de `Specs`. En cada uno, además de la rama del trabajo, borra las ramas propias del usuario ya mergeadas en la base y sus worktrees limpios, sin cambiar nunca la rama de un checkout. Conserva, y reporta con su motivo:
  - una rama con commits sin merge, o sin commits propios fuera de la base;
  - una rama activa en otro checkout;
  - un worktree con archivos sin commit, sin trackear, o ignorados fuera de las carpetas de dependencias o de build;
  - un worktree con un proceso cuyo directorio de trabajo está dentro, o cuando no se puede comprobar;
  - un trabajo con el cierre pendiente en memoria desde hace menos de 24 horas.
  Hoy es falso: «Branch cleanup» (`flow-close:87-92`) solo trata «the branch» del trabajo.
- **AC4** — `global.md` autoriza borrar las ramas de trabajo mergeadas, de este trabajo o de trabajos anteriores del usuario, en cada repositorio que tocó el trabajo, junto con sus worktrees limpios.
  Hoy es falso: la línea 31 solo dice «deleting merged work branches and their clean worktrees», sin alcance de trabajos anteriores ni de repositorios.
- **AC5** — Si `Specs` está en otro repositorio que declara `Delivery: direct-base`, `git-workflow` manda editar el change folder en el checkout principal de ese repo, sobre su base, sin rama ni worktree. Si otra sesión escribe ahí, se commitean solo las rutas propias.
  Hoy es falso: el paso 3 solo regula el caso «When `Specs` is in this repository».
- **AC6** — `flow-close` define el cierre de pausa de un trabajo que no está en `hold`. En él, el change folder se trata así:
  - en un repo de specs `direct-base`, commit y push a su base;
  - con `Specs` en el mismo repo y la rama del trabajo ya creada, commit en esa rama sin push;
  - con `Specs` en el mismo repo y sin rama todavía, queda sin commit y se registra como pendiente con su ruta;
  - en un repo de specs separado sin `direct-base`, queda sin commit, se registra como pendiente con su ruta, y el reporte recomienda declarar `Delivery: direct-base` en ese repo;
  - bajo `hold`, sigue sin commit.

  Hoy es falso: `flow-close:10` solo define la pausa de `hold`, y `git-workflow:16` solo commitea al llegar al punto de cierre.
- **AC7** — Una promoción no hecha nunca es un motivo para conservar una rama, un worktree, un proceso o material de soporte. Solo una tarea del plan que promueve a un entorno donde el ticket o la solicitud piden el cambio queda como registro de trabajo pendiente. `flow-plan` incluye esa promoción como tarea cuando hay esa petición; la petición es la instrucción de despliegue solo para ese entorno, y producción sigue llevando su propia pregunta.
  Hoy es falso: `flow-close:41` solo excluye la promoción de los registros de pendientes, y `flow-plan` no tiene ninguna regla sobre entornos.
- **AC8** — Los cambios están documentados:
  - `README.md` y `llms.txt` dejan de limitar la limpieza a «task-owned resources»;
  - `site/src/content/docs/flows/close.md` describe el barrido, los repos del workspace, los specs sin worktree, el commit al pausar y la regla de promociones, y corrige sus líneas 3 y 56;
  - `site/src/content/docs/flows/plan.md` describe la tarea de promoción;
  - los deltas de `walkthrough-viewer` y `workflow-lifecycle` validan;
  - el CHANGELOG tiene la entrada en Unreleased.

  Hoy es falso: ninguna de esas páginas menciona estos puntos.

## Prerrequisitos

Ninguno: la construcción no necesita nada de fuera del repositorio.

## Contexto verificado

- Las reglas se leyeron en `development` (`6670e799`): `content/skills/flow-close/SKILL.md` (15 933 bytes, con un tope de 16 000 que fija `tests/content/skill_size_test.go`), `content/skills/git-workflow/SKILL.md`, `content/skills/flow-walkthrough/references/viewer-publication.md` y `content/guidance/global.md:31`.
- Un hijo `hive-research` buscó todos los lugares que declaran estas reglas o dependen de ellas; están en `design.md` «Impacto».
- **Historia.**
  - v0.6.0 tenía el patrón de opción y volver a preguntar (`flow-walkthrough/SKILL.md:48-56`, commit `6cefc16a`). Lo borró el commit del rebuild `adf076cf` (2026-10-30), y `d54fb51d` (2026-11-02) lo reemplazó por el diseño actual.
  - v0.6.0 tenía el paso «Check the remaining branches» en `references/branch-cleanup.md`, que el rebuild también quitó.
  - Legacy (`tricell-hive-master`) trabajaba los repos de specs en su rama principal.
  - Ninguna de las dos versiones cubría la pausa.
- **Evidencia en sesiones.**
  - Visor: los episodios citados en el Objetivo.
  - Higiene: cam-specs tiene el worktree de cam-9, limpio y ya mergeado, aunque cam-specs declara `Delivery: direct-base`. En Codex 01a1aa61 el usuario tuvo que pedir la higiene. fac-235 sigue sin commit en facti-specs.

## Entrega

- **Modo:** `automatic` (D3 A). Revisión, push, PR a `development` y merge, sin parada de validación.
- **Revisión de código:** Claude Code `/code-review` (D4 A), sobre el candidato antes del push.
- **No autoriza:** publicar una release, desplegar ni tocar otros repositorios. La actualización de la instalación local que fija `AGENTS.md` para el checkout del mantenedor sí corre después del merge.

## Estado de revisión y avance

2026-11-05 — Plan escrito. La revisión de `hive-review-plan` dejó 13 hallazgos. Se aplicaron H3 a H13 y la parte de H1 que no necesita decisión. D5 A y D6 A resolvieron H1 y H2. Estado: Ready for flow-build.
