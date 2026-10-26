# Verificación por puntos de control según el riesgo

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #171 (`1543c8bf`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#168 — La verificación de cada tarea del plan frena el build: verificar en puntos de control según el riesgo](https://github.com/JhonHawk/tricell-hive/issues/168) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana · después de gh-167 |
| Verificación | conteo de frases · lectura · `go test ./tests/content/...` · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

Supernota del usuario al final de la sesión `b979e8b2` (2026-10-25):

> El flujo actual es: build task 1 -> Verify task 1 [waiting] -> Build task 2 -> verify task 2 -> ... Y esto esta tomando demasiado tiempo, lo considero un cuello de botella
> > Hay manera de que estas reviciones sean al final? el verify sea un paso? y talvez pueda correr en paralelo? Aunque, por ejemplo, no es lo mismo un plan de 4 task que uno de 16, se pueden hacer checkpoint para correr verifycaciones intermedias y no necesariamente dividir en 4, puedes ser un checkpoint en la task 2, luego en la 8, segun el tipo de tareas y el riesgo de implementarse

Sesión `5470b05e` (2026-10-25): «de hecho la sesión anterior en este mismo proyecto (con claude) tiene mucho de lo que menciono». Respuestas: D1-A (un ticket por tema y `flow-plan` conjunto); D2-A «y ajustemos la regla de los commits para que coincida con estas verificaciones (amenos que me recomiendes otra cosa)»; D4-A (un commit por tarea, en orden del plan); D5-A (un punto de control solo frena a las tareas que dependen de lo verificado ahí); D8-A (commit conjunto, luego reemplazado); D9-B (sin `hive-verify-change` final); D12-A durante el build, tras la re-revisión de `/code-review` del PR #171: cada tarea se commitea al reportarla y los arreglos van en commits `(T<n>)` aparte; la publicación para merge exige todas las tareas en `[x]`; dos tareas que comparten archivo no corren a la vez. Reemplaza D4-A y D8-A.

Criterios del ticket [#168](https://github.com/JhonHawk/tricell-hive/issues/168), leídos el 2026-10-25:

> - AC1: el formato del plan permite marcar en qué punto de control se verifica cada tarea, con un criterio de riesgo para elegirlo, y la revisión del plan bloquea un plan que deja sin verificar una tarea de alto riesgo.
> - AC2: `flow-build` verifica en los puntos de control marcados, no tras cada tarea, y una tarea dependiente puede arrancar antes del veredicto de su predecesora.
> - AC3: la regla de commits coincide con los puntos de control y nunca commitea una tarea sin veredicto independiente o aceptación del usuario.
> - AC4: los textos que dependen de la verificación por tarea (`plan-revision.md`, `ui-gate.md`, `flow-walkthrough`, `verification.md`, `reporting.md`, `site/.../flows/build.md`, `_support/docs/architecture/agent-delivery.md`) quedan coherentes.

## Objetivo

En un plan con criterios numerados, cada tarea `tdd` o `check` se verifica con `hive-verify-task` en cuanto el implementador la reporta, y una tarea que depende de ella espera ese veredicto. En cadenas de tareas dependientes la espera se suma: unos 21 min en la cadena T2→T3→T4 de la sesión `14a79ad0` y unos 6 min de verificadores en serie en `01a1579f`. El cambio deja que el plan elija, por riesgo, en qué puntos se verifica: las tareas de una cadena se construyen sin esperar entre sí, y solo un punto de control frena a lo que depende de él. Se conserva lo que motivó la verificación por tarea (ARK-730): ninguna tarea se da por verificada ni se commitea sin un veredicto independiente o la aceptación del usuario.

## Alcance y aceptación

Incluido: el formato del plan, la revisión del plan, la ejecución en `flow-build`, la regla de commits y los textos que dependen de ellos. Excluido, con su razón:

- Una única verificación al final (D2-C): un defecto temprano se arrastra por todas las tareas dependientes.
- Cambiar el contrato de `hive-verify-task` de una tarea por hijo: en un punto de control se lanza un verificador por tarea, todos juntos (decisión menor, ver diseño).
- `hive-verify-change`, `hive-review-ux` y la revisión de código: siguen en el candidato final como hoy.
- La revisión de seguridad a pedido: [gh-167](../gh-167-security-review-on-request/proposal.md).

Restricciones: un plan cuyas tareas no declaran punto de control sigue verificándose tarea por tarea, como hoy; los `SKILL.md` que se toquen siguen ≤ 16 000 bytes; la suite actual pasa.

- AC1. `plan-format.md` define el campo `**Checkpoint:** C<n>` de cada tarea `tdd` o `check`, la lista de puntos de control de `tasks.md` con la razón de cada uno, el criterio de riesgo para ubicarlos, que una tarea solo depende de tareas anteriores del plan, las líneas `waiting: T<n>`, `reopened by the fix of T<n>` y `blocked:` bajo `[/]` en la tabla de marcadores, y que, sin el campo, cada tarea es su propio punto de control. *Falso en la base cuando* `rg -c Checkpoint content/skills/flow-plan/references/plan-format.md` no encuentra nada.
- AC2. `plan-review.md` hace que toda asignación reporte como bloqueante una tarea que cumple el criterio de riesgo y no cierra un punto de control, y una tarea `tdd` o `check` sin punto de control en un plan que los usa. *Falso en la base cuando* `plan-review.md` no menciona puntos de control.
- AC3. `plan-execution.md` lanza las verificaciones de un punto de control cuando cada una de sus tareas está en `[?]` o fuera de él (bloqueada antes de reportarse o retirada, con su línea), un `hive-verify-task` por tarea y todos juntos, salvo las tareas que comparten archivo, que se verifican una tras otra; la re-verificación de una tarea `[x]` reabierta corre en el momento que fija su propia regla (la puerta del candidato aceptado para una corrección del walkthrough), sin esperar un punto. *Falso en la base cuando* `content/skills/flow-build/references/plan-execution.md:19` lanza el verificador en cuanto el implementador reporta la tarea.
- AC4. En `plan-execution.md`: una tarea que depende de otra del mismo punto de control arranca sobre lo que dejó el implementador; una que depende de tareas de un punto anterior espera a que esas tareas estén en `[x]`; una tarea con todos sus criterios `met` cuya dependencia del mismo punto no está en `[x]` queda en `[?]` con una línea `waiting: T<n>`; y cuando aterriza un arreglo o una reelaboración, las tareas posteriores que dependen de ella y no tienen commit se verifican de nuevo con todos sus criterios; la que falla vuelve a `[/]` con una línea que nombra el arreglo, sin gastar sus rondas. *Falso en la base cuando* `plan-execution.md:26` dice «a task that depends on it waits for `[x]`» sin distinguir puntos de control y no existe `waiting:`.
- AC5. La regla de commits (D12-A) commitea cada tarea al reportarla, antes de verificarla, con sus arreglos en commits `(T<n>)` aparte y sin reescribir historia; el trabajo se publica para merge solo cuando todas las tareas están en `[x]`, y una tarea retirada se deshace con un commit de revert. *Falso en la base cuando* `plan-execution.md:30` dice «commit each task yourself when it reaches `[x]`, and nothing earlier».
- AC8. Dos tareas cuyas Locations comparten un archivo nunca corren a la vez (implementación, ronda de arreglo ni verificación); un arreglo espera a los verificadores de su punto que comparten archivo con él; tras un arreglo se re-verifican las dependientes y las tareas que comparten archivo aún no integradas, y una tercera reapertura va al usuario. *Falso en la base cuando* `plan-execution.md:26` solo pausa la tarea que comparte archivo con un arreglo en curso y lanza juntas solo las verificaciones sin archivos en común.
- AC6. La descripción de `hive-verify-task`, la fila «Build iteration» de `verification.md`, `plan-revision.md:10-11` (una tarea agregada lleva su punto de control cuando el plan los usa; una tarea reabierta se verifica en el momento que fija su propia regla, sin esperar un punto) y `_support/docs/architecture/agent-delivery.md` sitúan la verificación en el punto de control de la tarea. *Falso en la base cuando* la descripción de `content/agents/quality/hive-verify-task.md:3` dice «after the implementer reports a task» y `plan-revision.md:10` dice «under the per-task procedure».
- AC7. Las páginas `site/src/content/docs/flows/plan.md` y `flows/build.md` explican los puntos de control, y `CHANGELOG.md` lo registra con `(#168)`. *Falso en la base cuando* ninguna menciona puntos de control ni `#168`.

Ticket AC1 → AC1 y AC2; ticket AC2 → AC3, AC4 y AC8; ticket AC3 → AC4 (`waiting:`) y AC5; ticket AC4 → AC6 y AC7. D9-B: no se agrega `hive-verify-change` como chequeo final en planes sin interfaz; cada tarea conserva su veredicto antes del commit y el candidato final pasa la suite y `/code-review`. De la lista del ticket, `ui-gate.md`, `flow-walkthrough` y `reporting.md` se revisaron: verifican las tareas cuyos archivos cambian y cuentan tareas verificadas, lo que sigue valiendo; T3 lo confirma con una búsqueda.

## Entrega

Valores declarados en `AGENTS.md` `## Hive` (base `development`): `Delivery: automatic` y `Review: /code-review`. Sin efecto de UI renderizada. Worktree propuesto `.claude/worktrees/gh-168`, rama propuesta `feat/gh-168-verification-checkpoints`; T1–T3 pueden avanzar en paralelo a gh-167 porque no comparten archivos; T4 y el merge van después del merge de gh-167, que comparte `CHANGELOG.md` y `site/.../flows/build.md`. La promoción a `master` queda fuera.

Temas abiertos: las vigilancias S3 (tarea dependiente que espera la verificación de su predecesora, sesión `14a79ad0`) y W16 (verificadores en serie, `01a1579f`) tienen destino en este cambio; se cierran en Engram al integrarlo.
