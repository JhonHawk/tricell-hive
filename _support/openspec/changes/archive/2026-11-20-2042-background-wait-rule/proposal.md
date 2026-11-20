# background-wait-rule — restaurar la regla de esperas en segundo plano y corregir `merged_branch_deleted`

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #555, 277a139a |
| Tracker | ninguno |
| Git | `JhonHawk/tricell-hive` · base `development` en `27875867` · rama propuesta `feat/background-wait-rule` |
| Playbook | feature |

## Decisiones

2026-11-20 (sesión de Claude Code `bc461a86`; registradas en `engram:tricell-hive/decision/background-wait-rule-restore`)
- **D1 A** — ¿Dónde vive la regla restaurada? En una hoja nueva, `hive-rule-background-wait`, con una línea en el índice de hojas de la guía global y un criterio de `--audit-session` que detecta cuando alguien se la salta. Se descartaron meterla en `hive-rule-checks`, cuyo momento de lectura no cubre las esperas de despliegues o builds en la nube, y un script de espera al estilo pstack.
- **D2 A** — ¿Cómo se corrige el falso fallo de `merged_branch_deleted` en la auditoría de sesiones reales? Da fail solo si un mensaje de cierre sigue al último merge sin que se borre ninguna rama; en cualquier otro caso da review. Así deja de marcar sesiones que aún no cierran, entre ellas las que están en medio de una promoción. Un merge de promoción seguido de un cierre sin limpieza sigue dando fail.
- **D3 A** — ¿Cómo convive con la rama sin integrar `refactor/global-guidance-slimming`? Se construye sobre `development` ahora, y esa rama absorbe la línea nueva del índice cuando se rebase.
- **supuesto** — El texto de la hoja restaura el de `7e811e46` y `d34637b8`: la espera termina en todos los resultados, se prefiere el código de salida, y una vigilancia con tope despierta al agente cuando el host solo avisa al salir. Se le añade que una salida no reconocida termina la espera. *Descartado:* redactarla desde cero. *Para revertir:* reescribir la hoja.
- **supuesto** — El criterio marca un bucle en segundo plano sin plazo. Hay plazo si el bucle va envuelto en `timeout`, si recorre una lista finita o si compara un contador o el tiempo transcurrido con un límite. El criterio no juzga si la condición de salida cubre todos los casos, porque eso requiere juicio. *Descartado:* intentar juzgar la condición. *Para revertir:* quitar el criterio.
- **supuesto** — El criterio lee Claude Code (`run_in_background`) y Grok (`block_until_ms: 0`). En Codex da not_observed como límite declarado, porque su ejecución unificada no marca qué comando corre en segundo plano. También es límite declarado que en Grok no sabemos si un valor distinto de 0 pasa el comando a segundo plano. Monitor de Claude Code queda fuera, porque tiene su propio `timeout_ms`. *Descartado:* inferirlo de la duración. *Para revertir:* añadir esos casos en otro cambio.
- **supuesto** — La hoja solo se enlaza desde el índice de la guía global, que es donde vive el momento de lectura, sin punteros desde flow-deploy ni flow-build. *Descartado:* enlazarla también desde esas skills. *Para revertir:* añadir esos enlaces.
- **supuesto** — No se intenta reconocer un merge de promoción por la rama de origen, porque `gh pr merge <n>` no la nombra en la transcripción. *Descartado:* leerla de la salida de `gh pr create`. *Para revertir:* añadirlo en otro cambio.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> «A, y B, validemos como teníamos las reglas anteriores, porque no recuerdo que tuviéramos un problema» (2026-11-20). La A era planear la regla de esperas y su criterio de auditoría; la B, corregir `merged_branch_deleted`.

El origen es la sesión de Grok `01a1f7e8-7f68-70b2-b053-db12af2bb562` en cam (CAM-22, 2026-11-20). Un bucle en segundo plano esperaba dos builds de Amplify y comparaba `"$c$b"` con `SUCCEEDSUCCEED`. Pero `aws amplify list-jobs --max-items 1 --output text` imprime además una línea `None`, así que ninguna rama coincidió nunca. El bucle no tenía plazo, Grok solo despierta al agente cuando la tarea termina, y la sesión quedó quieta 8 h 51 min, hasta que el usuario preguntó.

## Objetivo

Hoy ninguna regla de Hive cubre una espera hecha a mano fuera de CI y de los hijos, y la auditoría de sesiones no la detecta. Además, `merged_branch_deleted` marca como fallo a toda sesión que aún no cierra. Cuando el cambio termine, el agente tendrá una hoja que leer antes de lanzar esa espera, la auditoría marcará el bucle sin plazo, y `merged_branch_deleted` dará fail solo cuando un cierre de verdad no limpió las ramas.

## Alcance y aceptación

Dentro del alcance:
- la hoja nueva y su línea en `content/guidance/global.md`;
- el criterio `unbounded_background_wait` con sus fixtures;
- el ajuste de `merged_branch_deleted` en la auditoría de sesiones reales;
- las filas de `session-review-rubric.md` y `rule-enforcers.md`;
- la delta de `workflow-lifecycle`.

Fuera del alcance:
- un script de espera;
- la detección en Codex;
- el detector `merged_branch_deleted` de los pilotos (`regression.go`), que no cambia;
- la rama de adelgazamiento.

La sesión de cam sigue abierta y crece. Por eso AC1 y AC2 se observan sobre una copia cortada en sus primeras 3909 líneas, la de hasta el «cómo vamos?» de las 13:13. La copia está en `_support/workspace/background-wait-rule/grok-01a1f7e8-3909.jsonl`, que Git ignora; es privada (modo 600), porque trae datos de cam, y no se versiona.

- AC1 — `go run ./tests/pilot --audit-session _support/workspace/background-wait-rule/grok-01a1f7e8-3909.jsonl --host grok` reporta `unbounded_background_wait` en `fail` con la evidencia de la línea 3855 · falso hoy porque la salida no tiene ese criterio · Observed in: la auditoría de la copia de la sesión real de Grok `01a1f7e8`
- AC2 — La misma auditoría reporta `merged_branch_deleted` en `review`, no en `fail` · falso hoy porque reporta `fail` con evidencia en la línea 3776 (corrido el 2026-11-20) · Observed in: la auditoría de la copia de la sesión real de Grok `01a1f7e8`
- AC3 — En una transcripción de Claude Code y en una de Grok, un bucle en segundo plano con `sleep` y sin plazo da `fail` · falso hoy porque el criterio no existe · Observed in: `--audit-session` sobre `tests/fixtures/sessions/unbounded_background_wait/`, con la estructura real de cada host y valores sintéticos
- AC5 — Un bucle en segundo plano acotado da `pass`: envuelto en `timeout N`, recorriendo una lista finita o comparando un contador o el tiempo transcurrido con un límite. Un `timeout` dentro del cuerpo del bucle no cuenta como plazo · falso hoy porque el criterio no existe · Observed in: `--audit-session` sobre `tests/fixtures/sessions/unbounded_background_wait/`
- AC4 — Después de la actualización local, el índice de hojas de la guía instalada en `~/.claude/CLAUDE.md` enlaza `hive-rule-background-wait/SKILL.md` instalado. La hoja exige: terminar en éxito, fallo, objetivo inexistente, proceso colgado y salida no reconocida; un plazo; y tramos de entre 5 y 10 minutos que despierten al agente cuando el host solo avisa al salir · falso hoy porque esa hoja no existe · Observed in: la guía instalada tras `hive update` desde el commit integrado

## Requisitos previos

None: the build needs nothing from outside the repository.

## Contexto verificado

- La regla existió en `content/skills/flow-research/references/delegation.md`, introducida en `7e811e46` (2026-10-16) y ampliada en `d34637b8` (2026-10-20). `adf076cf` (2026-10-30) la vació en la reconstrucción de skills. La fila R-D-10 del inventario `_support/workspace/2026-10-30-skills-rebuild/f2-rule-inventory.md` la conservaba como referencia rara, pero no se recreó. Verificado con `git show adf076cf^:…delegation.md`, líneas 53-57.
- `git grep` en HEAD no encuentra `every outcome` ni `bounded watch` en `content/`.
- pstack (`93b916d4`) resuelve la espera de CI con un script, `watch-pr`: plazo opcional, tope de errores y estados desconocidos tratados como fallo. No tiene ninguna regla escrita para bucles hechos a mano.
- La auditoría registra `merged_branch_deleted` en `tests/pilot/session_audit.go:1110` (`existingAuditCriteria`) y reutiliza `mergedBranchDeletedWith` de `regression.go:596`. `isCloseReport` ya existe.
- `traceEvent.Input` conserva los argumentos del tool. En la sesión de cam, la llamada de Grok que lanzó el bucle trae `block_until_ms: 0` y le sigue `task_backgrounded`.
- `content/guidance/global.md` pesa 32,148 bytes en `27875867`. La rama `refactor/global-guidance-slimming` (punta `089cd08b`, sin integrar) fija un presupuesto de 27,500, y su índice va sin negrita, así que la línea nueva se reformatea cuando esa rama se rebase.
- La observación para el commit, según «Measurement» de AGENTS.md: Grok Build, modelo `grok-4.7-build-fast` (fila del piso, provisional), sesión `01a1f7e8-7f68-70b2-b053-db12af2bb562`, cam, 2026-11-20, 8 h 51 min sin actividad. Las anteriores: Claude Code `e5edd628` y `b6644329` (2026-10-16), Grok `01a1579f` (2026-10-20) y Claude Code `b4b1285f` (2026-11-05).
- Riesgo medido solo en parte: el prompt de Claude Code sugiere `run_in_background` con `until …; do sleep …; done` sin plazo, así que el criterio puede fallar en muchas sesiones de Claude. Es coherente con la hoja, pero no medimos cuántas.
- `aws amplify list-jobs --max-items 1 --query 'jobSummaries[0].status' --output text` imprime `SUCCEED\nNone\n`. Reproducido el 2026-11-20 con lectura de solo lectura.

## Effects

None: the tasks take no effect on an external system.

## Entrega

2026-11-20: entrega `automatic`. Reviso, hago push, abro el pull request hacia `development` y lo fusiono sin pararme a validar. La revisión de código es la nativa, `/code-review` de Claude Code. Después de integrar, el marcador de mantenedor autoriza la actualización local (`hive update` desde el commit integrado). Esta respuesta no autoriza release, despliegue del sitio ni cambios a la rama de adelgazamiento.

## Estado de la revisión y avance

2026-11-20 — Plan escrito y revisado por `hive-review-plan`, con 1 hallazgo P1, 4 P2 y varios P3, todos aplicados:
- AC1 y AC2 se observan sobre una copia fija de la sesión, porque la original sigue abierta.
- Se corrigen los fixtures de T3: `local-merge` sí cambia y el de Codex no.
- Se quita `review` del criterio nuevo, porque Claude y Grok no ocultan el comando.
- La detección de plazo usa `shellSegments`: el `timeout` cuenta solo si envuelve el bucle, una lista finita acota, y el tiempo cuenta solo si se compara.
- Los fixtures llevan valores sintéticos.
- AC3 se divide en AC3 y AC5.
- La hoja recupera de la regla anterior el proceso colgado, el relanzamiento ante «no target» y el tramo de 5 a 10 minutos, y añade parar tras dos tramos sin avance.
- Se añaden el CHANGELOG, la frase de la rúbrica y la fila de rule-enforcers.
- D2 queda acotada.
- La punta de la rama de adelgazamiento pasa a `089cd08b` y se registra la observación para el commit.

2026-11-20 — Revisión durante el build: la delta y design.md cuentan como plazo también la condición del propio bucle y un `for ((…))` con condición. Sin eso, un `while [ $i -lt 60 ]` acotado daba fail. No invalida tareas.
2026-11-20 — Revisión de la delta (hive-review-plan, 0 P0/P1, 2 P2 y 2 P3): el bucle se reconoce después de then/do/(/{; el for C-style exige comparar con un límite; fixtures nuevos; operadores corregidos en design.md; `|| true` y el tiempo leído antes del bucle quedan como límites aceptados.
