# base-sync-admin-grant-cost — Base local al día al cerrar, permiso de `--admin` por sesión y costo del hilo principal en el auditor

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #532, merge 2f7b7f6a |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `8a4443a2` · rama `feat/base-sync-admin-grant` en `.claude/worktrees/base-sync` |
| Playbook | ninguno (revisión de sesión, luego `/flow-plan` tecleado por el usuario) |

## Decisions

2026-11-19
- **D1 A** — ¿Cómo se actualiza la rama base local después del merge, al cerrar? Lo hace `cleanup_branches.py --apply` con `fetch` y `merge --ff-only` en el checkout principal; cuando lo que llega toca archivos con cambios locales, o las ramas divergieron, no avanza y lo reporta. Descartado: solo texto en la referencia, o dejarlo como está.
- **D2 A** — ¿Cuánto dura el permiso del usuario para mezclar con `--admin`? Toda la sesión, para los pull requests con los checks requeridos en verde en los repositorios del trabajo, nombrado en cada reporte; nunca cubre la rama de producción, un push directo ni una desviación del camino de promoción. Descartado: por repositorio (forma anterior al 2026-10-09) y por merge (regla actual).
- **D3 A** — ¿Qué hacemos con la duración de sesiones como la de PRL-928? Medir primero: un desglose informativo del hilo principal en `--audit-session` y la comparación de 3 sesiones antes de recortar nada. Descartado: delegar más ya, o nada.
- **D4 A** — Con ese permiso vigente, ¿qué pasa cuando un revisor humano aprobó un commit anterior y no la punta? Se mezcla y el reporte nombra al revisor y el commit aprobado; un revisor pedido sin respuesta o una revisión con cambios solicitados siguen reteniendo el merge. Descartado: esperar siempre, o cubrir cualquier retención.
- **Seguridad A** — ¿Corre `hive-review-security` sobre el candidato? Sí, porque D2 cambia cuándo el agente salta la protección de una rama.
- **Supuesto** — El script hace su propio `git fetch origin <base>` con `--apply`. Descartado: confiar en que el agente refresque las referencias antes. Se revierte quitando el fetch.
- **Supuesto** — Solo avanza la base cuando está abierta en el checkout principal; si está en otro checkout o en ninguno, lo reporta sin moverla. Descartado: mover la referencia directamente. Se revierte agregando ese caso.
- **Supuesto** — El desglose de D3 es informativo, sin `pass` ni `fail`, como `card_answers`. Descartado: un umbral. Se revierte agregando un criterio.
- **Supuesto** — «Lectura de Hive» es un archivo `SKILL.md` o `references/` de una skill instalada leído con la herramienta de lectura o con `cat`, `sed`, `head`, `tail` o `rg`; «script de Hive» es un comando que corre `<skill>/scripts/*.py`. Descartado: una lista fija de archivos. Se revierte cambiando el patrón.
- **Supuesto** — El plan termina midiendo 3 sesiones y entregando el desglose en el reporte final, sin cambiar reglas. Descartado: medir en otra sesión. Se revierte quitando esa tarea.

2026-11-20 (revisión del plan)
- **Supuesto** — `hive-rule-merge-gate` entra en las excepciones de tamaño de `tests/content/leaf_skill_test.go` (7000 bytes), porque el texto nuevo la deja en unos 4360 bytes y recortar 360 bytes reescribiría la regla de espera de CI, que este cambio no toca. Descartado: reescribir ese párrafo, o una hoja nueva para el permiso. Se revierte recortando la hoja y quitando la excepción.
- **Supuesto** — `content/guidance/global.md` cambia en tres frases que este cambio contradiría: «repeated approvals are not standing permission» (línea 20) gana la excepción del permiso que una regla limita a la sesión; «a human reviewer's pending approval holds the merge» (línea 140) gana «unless an admin-merge grant lifts it»; y «Preservation» autoriza al cierre avanzar la base del checkout principal como `branch-cleanup.md` lo fija. Descartado: dejar el global sin tocar, que haría que un modelo literal se niegue. Se revierte quitando esas cláusulas.
- **Supuesto** — La rama de producción es la última del encadenamiento `Environments` del `## Hive` cuando no es la base; un repositorio sin encadenamiento no tiene rama de producción para esta regla. Descartado: una lista de nombres como `main` o `master`, que excluiría el `master` de specs, su base. Se revierte nombrando la rama en la regla.
- **Supuesto** — El permiso cubre un pull request cuyos checks requeridos pasaron o que no tiene ninguno (el PR 31 de specs no tenía). Un revisor que aprobó un commit anterior y volvió a ser pedido sigue reteniendo: gana la petición. Descartado: solo PR con checks; que gane la aprobación vieja. Se revierte en la hoja.
- **Supuesto** — El script hace el `fetch` y el avance de la base antes de planear la limpieza, para que una rama mezclada solo en `origin` cuente como mezclada; con la base en otro checkout o sin checkout principal, no la mueve. Corre `status` y `merge` en el checkout principal aunque `--repo` sea un worktree, con `--no-overwrite-ignore`, y cuenta como cambio local un archivo ignorado que lo que llega empieza a seguir. Descartado: avanzar después de borrar. Se revierte moviendo el paso.
- **Supuesto** — En Grok, el uso por turno del padre incluye el de sus hijos: el auditor resta el `usage.json` de cada hijo de `subagents/*/meta.json` (buscado en todas las carpetas de `~/.grok/sessions/`) y marca `per_turn_with_children` cuando falta alguno. En Codex cuenta un `token_count` solo cuando su total cambió. Una «lectura de Hive» es solo de una skill instalada bajo una carpeta oculta del usuario (como `~/.agents/skills/`), no de `content/skills/` del repositorio. Descartado: sumar el uso del padre tal cual. Se revierte en el auditor.
- **Nota de medición** — D2 relaja la regla que `a57af495` endureció tras el push directo a `production` de project-g Grok `01a11cbc`. La única observación es una sesión de un modelo del floor (`grok-4.7-build-fast`, como modelo de las sesiones del usuario; este plan lo dio por error como fuera del floor), y relajar una regla es decisión del usuario, como «Measurement» de `AGENTS.md` fija; las exclusiones de la rama de producción, el push directo y la desviación de promoción cubren ese incidente. Se compararon pstack (permisos por conjunto declarado y por sesión, pausa en despliegues), la forma anterior de Hive (por repositorio) y la actual (por bypass) antes de recomendar.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> D1–D3 «Script (★)», «Toda la sesión (★)», «Medir primero (★)» — usuario, 2026-11-19, tras la revisión de la sesión en project-b: «La higiene no esta contemplando descargar cambios del repositorio; se perdio el hecho de poder conservar un permiso y retulizarlo durante la sesión, por ejemplo el --admin; El tema de esa sesión fue de infra/devops, pero se volvio larguisima».

Origen: sesión Grok Build `01a1f5fc-60ea-7971-8d04-ca82a187fa9e` en project-b (PRL-928, `grok-4.7-build-fast` high, en el floor como modelo de las sesiones del usuario), 2026-11-19 16:45–22:08 CST.
- Al cerrar, el checkout de specs quedó 20 commits detrás de `origin/master` y el usuario pidió «haz el pull de cambios».
- El usuario autorizó `--admin` para el PR 291 del backend a las 18:59 y tuvo que repetirlo a las 20:53 para el PR 31 de specs.
- La sesión hizo 599 llamadas al modelo con 70.8 M de tokens de entrada (unos $32.5 en total); restando los 12 hijos (189 llamadas, 15.7 M), el hilo principal hizo 410 llamadas con 55.1 M. Cerca del 25 % de sus llamadas a herramientas leyeron reglas o corrieron scripts de Hive, y hubo 5 compactaciones.

Decisión: `engram:tricell-hive/decision/fleet-prl928-base-sync-admin-grant-length`.

## Objective

Hoy el cierre deja la base local atrás de `origin`, cada bypass con `--admin` se pide de nuevo aunque el usuario ya lo dio en la sesión, y no hay forma repetible de ver en qué se va el hilo principal. Al terminar, el cierre deja la base al día cuando es seguro, un permiso de admin-merge dura la sesión dentro de sus límites, y `--audit-session` da el desglose del hilo principal.

## Scope and acceptance

Dentro: `cleanup_branches.py` y su referencia; la hoja `hive-rule-merge-gate` y su excepción de tamaño; tres frases de `content/guidance/global.md`; un bloque informativo en `--audit-session` para Claude, Codex y Grok; documentación (referencia, sitio, rúbrica, `rule-enforcers.md`, changelog) y la medición de 3 sesiones. Fuera: cambiar la delegación, la ceremonia de plan o cualquier otra regla por la medición; el camino de promoción, cuyas desviaciones siguen pidiéndose una a una.

- AC1 — En un repositorio cuyo checkout principal está en la base y detrás de `origin`, sin cambios locales en los archivos que llegan, `cleanup_branches.py --apply` deja la base en la punta de `origin/<base>` y reporta `base_update.action: fast_forward` con los commits avanzados · false today because el script nunca toca la base (el checkout de specs de project-b quedó 20 commits atrás) · Observed in: un repositorio temporal con `origin` en `tests/skills/cleanup_branches_test.py`
- AC2 — Cuando lo que llega toca un archivo modificado o sin seguimiento, la base local tiene commits que no están en `origin`, o el checkout principal está en otra rama, el script deja la base donde estaba y reporta `base_update.action: keep` con la razón y los archivos · false today because el plan no tiene ninguna entrada para la base · Observed in: un repositorio temporal con `origin` en `tests/skills/cleanup_branches_test.py`
- AC3 — La hoja `hive-rule-merge-gate` instalada dice que un permiso explícito de admin-merge vale el resto de la sesión, salvo que el usuario lo retire, para los pull requests con los checks requeridos en verde en los repositorios del trabajo, que cubre a un revisor cuya última revisión aprueba un commit anterior (nombrado en el reporte), y que nunca cubre la rama de producción, un push directo, una desviación del camino de promoción, un revisor pedido sin respuesta ni cambios solicitados; la hoja pesa 4000 bytes o menos · false today because la hoja dice «a grant for one bypass does not cover later ones» · Observed in: `~/.agents/skills/hive-rule-merge-gate/SKILL.md` tras el despliegue · **Retired:** el texto no cabe en 4000 bytes y le faltaban los PR sin checks y la rama de producción definida; lo reemplaza AC7
- AC4 — En una sesión real iniciada tras el despliegue, después de que el usuario autoriza un admin-merge, el siguiente admin-merge de la sesión en un repositorio del trabajo se hace sin preguntar y su reporte nombra el permiso · false today because en project-b `01a1f5fc` el agente se detuvo a las 20:37 y el usuario repitió «dale con --admin» · Observed in: una sesión real iniciada tras el despliegue
- AC5 — `go run ./tests/pilot --audit-session` da un bloque `main_thread_cost` con llamadas al modelo, tokens de entrada (con su parte en caché), compactaciones, llamadas a herramientas, lecturas de Hive, scripts de Hive y llamadas por herramienta, para Claude, Codex y Grok; en project-b `01a1f5fc` da 599 llamadas y 70,803,744 tokens de entrada, igual que `usage.json` · false today because el reporte no tiene ese bloque · Observed in: `--audit-session` sobre la sesión Grok `01a1f5fc` y sobre una de Claude y una de Codex · **Retired:** esas cifras de Grok incluyen a los hijos; lo reemplaza AC8
- AC6 — El reporte final da el desglose de 3 sesiones reales nombradas (project-b Grok `01a1f5fc`, project-g Codex `01a1f5d4`, Claude `bf678880`), con la parte de lecturas y scripts de Hive y las compactaciones, sin proponer ni hacer cambios de reglas · false today because no existe ese desglose · Observed in: el reporte final de este trabajo
- AC7 — **Replaces:** AC3 · La hoja `hive-rule-merge-gate` instalada dice que un permiso explícito de admin-merge vale el resto de la sesión, salvo que el usuario lo retire, para los pull requests de los repositorios del trabajo cuyos checks requeridos pasaron o que no tienen ninguno; que levanta la retención de un revisor cuya última revisión aprueba un commit anterior, nombrado con ese commit en el reporte; y que nunca cubre la rama de producción (la última de `Environments` cuando no es la base), un push directo, una desviación del camino de promoción, un revisor pedido (aunque haya aprobado antes) ni cambios solicitados; `content/guidance/global.md` instalado ya no contradice ese permiso, y `go test ./tests/content/...` pasa con la hoja en su excepción de tamaño · false today because la hoja dice «a grant for one bypass does not cover later ones» y el global dice «repeated approvals are not standing permission» · Observed in: `~/.agents/skills/hive-rule-merge-gate/SKILL.md` y la guía global instalada tras el despliegue
- AC8 — **Replaces:** AC5 · `go run ./tests/pilot --audit-session` da un bloque `main_thread_cost` con llamadas al modelo, tokens de entrada (con su parte en caché), compactaciones, llamadas a herramientas, lecturas de skills instaladas de Hive, scripts de Hive y llamadas por herramienta, del hilo principal sin sus hijos, para Claude, Codex y Grok; en project-b `01a1f5fc` da 410 llamadas y 55,115,725 tokens de entrada (599 y 70,803,744 menos los 12 hijos) con `usage_source` `per_turn` · false today because el reporte no tiene ese bloque · Observed in: `--audit-session` sobre la sesión Grok `01a1f5fc` y sobre una de Claude y una de Codex

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/flow-close/scripts/cleanup_branches.py` (444 líneas) planea y aplica la limpieza de worktrees y ramas; no hace `fetch`, ni `merge`, ni reporta la base. `tests/skills/cleanup_branches_test.py` ya arma un `origin` desnudo y un clon.
- `content/skills/flow-close/references/branch-cleanup.md:9` dice «return its checkout to the base», sin actualizarla. `8e5fba29` (2026-09-23) decía «return to the base, prune and fast-forward»; se perdió en `adf076cf` (2026-10-30). Ninguna hoja de `content/` hace fast-forward hoy.
- `content/skills/hive-rule-merge-gate/SKILL.md` pesa 3930 de 4000 bytes; su último punto dice «a grant for one bypass does not cover later ones». Antes de `a57af495` (2026-10-09) el permiso valía «for that repository»; ese commit lo endureció tras el push directo a `production` de project-g Grok `01a11cbc`. `hive-rule-git-commit` ya deja que la aprobación de saltar hooks valga «for the rest of the session unless withdrawn».
- `hive-rule-promotion-path` pide cada desviación como pregunta propia; no cambia.
- `tests/content/single_home_test.go:44` fija la frase «Never bypass branch protection, required reviews or required checks» en la hoja del merge.
- `--audit-session` lee Claude, Codex y Grok (`tests/pilot/session_audit.go`) y ya da un bloque informativo `card_answers`; `traceEvent` no guarda tokens. `_support/sessions/2026-11-03-delegation-cost/measure-delegation.py` mide llamadas, tokens procesados y compactaciones del hilo principal para Codex y Claude, pero no para Grok.
- Grok expone el uso por turno: la suma de `turn_completed.usage.inputTokens` de `01a1f5fc` da 70,803,744 y la de `modelCalls` coincide con `usage.json` (599); sus mensajes no traen uso por llamada (medido el 2026-11-19).

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic`: revisión, push, pull request a `development` y merge con commit de merge, sin parada de validación. Revisión de código con `/code-review` nativo de Claude Code y un hijo `hive-review-security`. Despliegue local de mantenedor con `hive update --source` tras el merge. No autoriza cambios en otros repositorios ni en la configuración global fuera de ese despliegue.

## Review status and progress

2026-11-19 — Plan escrito.
2026-11-20 — Revisión del plan (`hive-review-plan`): 6 P2 y 4 P3. Aplicados: excepción de tamaño de la hoja del merge; tres frases del global; rama de producción definida; PR sin checks y revisor vuelto a pedir; avance de la base antes de planear, en el checkout principal, con `--no-overwrite-ignore` y archivos ignorados; resta de los hijos de Grok y cifras corregidas; `token_count` repetidos de Codex; lecturas de Hive solo de skills instaladas; escenario de producción del delta corregido. AC3 y AC5 retirados, reemplazados por AC7 y AC8.
