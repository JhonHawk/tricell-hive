# Diseño

## Contexto verificado

**Guía actual (observado en `ec20d91`):**
- `content/skills/flow-build/SKILL.md:65` pide cerrar la cobertura de regresión "through the project's actual CI gate". `:18` dice que el modo de entrega cubre "waiting for the CI its merge depends on", o sea el CI **anterior** al merge.
- `content/skills/flow-build/references/verification.md:75` pide cumplir los gates de regresión al integrar y confirmar qué ejecuta el CI. No menciona las corridas que dispara el merge o el push. `:85` ya retiene tickets desde esta referencia ("its tickets stay out of their completion state").
- `content/skills/git-workflow/SKILL.md:16` manda leer `verification.md` antes de una acción de Git que dispare CI. `:30` pide inspeccionar "the pull request, checks, or merge state" tras la acción.
- `content/guidance/global.md:41`: un ticket está hecho cuando se integra en la rama base, "move it then", "unless the project declares a later completion point". Esa misma línea autoriza la escritura en el tracker ("this rule authorizes that tracker write"), el patrón que usa aquí el relanzamiento.
- `content/skills/flow-plan/references/delivery-decisions.md:12`: un modo autoriza solo sus efectos de entrega elegidos, "not … unspecified external effects by implication". Por eso el relanzamiento no puede colgar del modo.

**ARK-730, sample-project (fuente: cuerpo de #43, no reverificado aquí):** la suite completa del backend corre solo en el push a `development`; en la PR aparece como `skipping`. La corrida posterior al merge (36283836687) falló al instalar pnpm por un timeout de registry.npmjs.org. El cambio incluía una migración con test de integración de backend.

**ARK-734, ark (observado 2026-09-27 con `gh`, solo lectura):**
- La PR #142 tuvo `Integration` y `Docker` en `skipping`; el resto pasó.
- La corrida posterior al merge (36307256396, push a `development`) ejecutó `Integration` en verde; `Docker` volvió a saltarse.
- En `.github/workflows/ci.yml`, `Integration` se dispara con `any_code == 'true'`, que incluye los frontends. Sin embargo, solo ejecuta `turbo run test:integration --filter=ark-backend` y `--filter=ark-transactional`.
- El diff integrado (`602fa761`) cambia una línea en `apps/customer-portal/.../invite-member-modal.tsx`.
- ark tiene además `trunk-red-notifier.yml`, que abre un issue cuando el CI de `development` queda en rojo.

**Sesión de ARK-734 (transcript de Codex `01a0e1fb`, observado):** esperó con `gh run watch` desde las 02:48 hasta las 02:55 hora local. Movió el ticket a *To Deploy* antes de esa espera y cerró al terminar.

## Regla nueva en `verification.md`

Dos viñetas nuevas en "Frequency and coverage", justo después de la viñeta "At integration or promotion…". El texto:

> - After a merge or push you performed to the base branch, find the CI runs its commit triggered there. Wait for a run before reporting the integration complete only when it executes a job that tests a changed file and did not already run on the pull request, such as one shown there as skipped or absent; without a pull request, every such job counts. A job tests a changed file when the file is in the packages it tests or in their dependencies, or when it runs the whole repository's suite. Judge a job by what it tests, not by the filter that triggers it, which can be broader. Also wait when the workflow does not show what a job tests, or when a workflow that the push should trigger has no run listed yet. Otherwise, do not wait: give the run's link as a reminder. A deployment or publication job that the push triggers falls under the deployment checks below, not under this rule. Wait on the run's completion signal under the shared waiting rule. Until the run passes, the ticket move that global guidance makes at integration and the closing of the change folder wait; if the turn must end first, report the open run as a pending item.
> - When a run you waited for fails, read its failed jobs' logs. A failure that the base branch's previous run already had is not this change's: handle it under the shared incidental-findings rule. For an infrastructure failure, such as a registry or network timeout or an unavailable runner, rerun only the failed test jobs, once; this rule authorizes that rerun, not a rerun of a deployment or publication job. When the rerun fails too, or the failure comes from the change's code, the integration is not complete and its ticket stays out of its completion state: report the failed jobs, the diagnosis, and the run link as a pending item.

**Por qué así:**
- **Criterio por lo que prueba el job, no por su disparo.** En ark, el disparador de `Integration` incluye los frontends. Con un criterio por disparo, ARK-734 habría esperado igual. El caso de ark muestra que ese error ocurre en un repositorio real, no que sea hipotético.
- **Merge o push (D5-A).** En `direct-base` la corrida del push es el único CI. Sin PR no hay con qué comparar, así que cuenta todo job que pruebe un archivo cambiado.
- **"Ante la duda, espera", incluida la corrida que aún no aparece.** GitHub tarda en registrar la corrida de un push. Un modelo que no ve ninguna y cierra repite ARK-730. Solo un workflow cuyo filtro excluye el cambio significa "no hay corrida".
- **Retención explícita del ticket y del change folder.** `global.md:41` está siempre en contexto y dice "move it then". ARK-734 movió el ticket antes de esperar. La viñeta nombra lo que retrasa, como ya hace `verification.md:85`, sin editar `global.md`.
- **Pendiente si el turno termina.** La regla compartida de espera permite terminar el turno con un aviso de trabajo en curso. Sin esta frase, una corrida abierta al cierre no dejaría ninguna obligación registrada.
- **Relanzamiento autorizado por la regla, solo para tests.** El modo de entrega no cubre efectos no nombrados (`delivery-decisions.md:12`), y en `direct-base` no hay merge del que colgar. Relanzar un job de despliegue o publicación podría repetir un efecto externo, así que queda fuera.
- **Fallo previo como hallazgo incidental (D6-A).** Un test que ya fallaba en la corrida anterior de la base bloquearía el ticket por algo ajeno al cambio.
- **Espera según la regla compartida.** `flow-build/SKILL.md:48` ya usa el término y `global.md` define cómo esperar. Aquí solo se remite a ella.

**Comprobación con los dos casos:**

| Caso | Job que no corrió en la PR | ¿Prueba un archivo cambiado? | Resultado de la regla |
| --- | --- | --- | --- |
| ARK-730 | Suite completa del backend | Sí: la migración está en el backend | Esperar; ante el timeout de npm (infraestructura), relanzar una vez los jobs de test fallidos; el ticket espera |
| ARK-734 | `Integration` | No: prueba `ark-backend` y `ark-transactional`; el cambio está en `customer-portal` | No esperar; recordatorio con el enlace; el ticket se mueve al integrar |

## Referencia en `git-workflow`

En `content/skills/git-workflow/SKILL.md:30`, tras "…the pull request, checks, or merge state when those effects were authorized.", se agrega:

> After a merge or push to the base branch, also check the CI it triggered there, as [verification](../flow-build/references/verification.md) describes under frequency and coverage.

`:16` hace leer `verification.md` **antes** de la acción. Esta línea lo recuerda en el momento del fallo de ARK-730: al verificar el merge. Es un puntero, no una segunda copia de la regla. El enlace va sin ancla porque `skill_links_test.go` solo valida enlaces de `global.md`, y `:16` ya usa esa misma forma.

## Relación con la regla global de ticket completado

`global.md:41` mueve el ticket al integrar. La regla nueva no declara un punto de cierre de proyecto. Retrasa ese movimiento, y el archivo del change folder, hasta que pase la corrida que cubre el cambio, y lo dice explícitamente. No se edita `global.md`, que tiene un presupuesto de bytes ajustado y se carga en cada sesión. La condición solo aplica en la frontera de integración, donde `git-workflow` y `flow-build` ya remiten a `verification.md`.

## Procedencia en el commit

`AGENTS.md` ("Measurement") pide que el mensaje de commit de un cambio de guía que arregla una falla observada nombre la sesión (host, ID, fecha) y la falla, sin agregar un caso de regresión. Aquí hay dos:
- sample-project, OpenCode `ses_f20143eacffe8d44SvMbTnjkr7`, 2026-09-26: ARK-730 se cerró con la corrida posterior al merge en rojo, y esa corrida era la única que ejecutaba la suite del backend que cubría el cambio.
- ark, Codex `01a0e1fb-2578-73b3-b5df-e5e037dc5759`, 2026-09-27: ARK-734 esperó 7 minutos una corrida posterior al merge cuyo único job nuevo no probaba el cambio.
