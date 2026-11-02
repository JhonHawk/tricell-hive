# #312 — Entrega después del push: bots de revisión, base al día y modo `ready`

| Campo | Valor |
| --- | --- |
| Estado | Completed · integrado en `development` el 2026-11-02 con el PR #313 (merge `df2591f9`); #312 cerrada |
| Tracker | [#312](https://github.com/JhonHawk/tricell-hive/issues/312) · cerrada · enhancement · relacionada con #311 |
| Git | `JhonHawk/tricell-hive`, base `development` en `5fa27f76`; se construye en el checkout actual, rama `fix/post-0.7.0-corrections`, que ya lleva `26883dcc` (arreglo de `flow-close`, sin PR); los dos trabajos salen en un solo PR (D9 B) |

## Decisiones

2026-11-01
- ¿Qué ideas del research entran? La entrega después del push (esta issue) y la intención con enfoques (#311, aparte); quedan fuera el freno al crecimiento del diff, los PRs viejos o ajenos y la firma de comentarios. (D1 A, B)
- ¿Entra el freno al crecimiento del diff por la revisión? No: Hive ya limita las rondas y marca como P2 lo que la tarea no pidió. Este plan extiende ese mismo límite de una ronda a los hallazgos de bots después del push. (D4 A)

2026-11-02
- ¿Cómo se comporta el modo nuevo? Como `automatic` hasta el PR: revisa, hace push, abre el PR, espera CI y bots, corrige y se detiene sin merge; el usuario valida en el PR. (D5 A)
- ¿Un bot que solo comenta y aún no revisó el último commit frena el merge? No: al terminar el CI se leen una vez las revisiones y comentarios del último commit; el bot que falte queda como pendiente en el reporte y `automatic` hace el merge igual. (D6 A)
- ¿Cómo se entrega este cambio? `interactive`, con Claude Code `/code-review` antes de la validación. (D7 A, D8 A)
- ¿Sobre qué base se construye? Sobre `fix/post-0.7.0-corrections`, en un solo PR junto con el arreglo de `flow-close` (`26883dcc`); no hay choque en `CHANGELOG.md`. (D9 B)
- ¿Cómo se llama el modo? `ready`, y la regla aclara que el PR se abre como no borrador y que `gh pr ready` no forma parte del modo. (D10 A)
- ¿Qué pasa con los demás punteros rotos de `instruction-resources.md`? Se resuelven en este mismo cambio: cada entrada apunta al hogar actual de su regla o sale de la lista. (D11, respuesta libre)

Supuestos del plan, sin decisión del usuario:
- Bajo `ready` siempre se recomienda una revisión dedicada, porque cada hallazgo que aparece después del push cuesta una ronda de CI y de bots.
- No hay script vigía ni sondeo; se sigue usando la espera bloqueante del CI que ya existe.
- Un bot es una cuenta de tipo `Bot` en la API de GitHub o con login terminado en `[bot]`.
- Un check de bot que falla y no es requerido cuenta como hallazgo, no como fallo del CI.
- `gh run watch --exit-status` lee por REST; se comprueba en el build (sin corridas en este repo para observarlo hoy).

## Pedido

> Después del push, Hive espera el CI con un comando bloqueante y atiende a los revisores humanos, pero no trata los bots de revisión de IA [...], no distingue los checks y comentarios más nuevos que el último push [...], antes del merge solo comprueba que la base sea ancestro de la rama [...] y los modos de entrega no tienen un punto intermedio.

Criterios de la issue, leídos el 2026-11-02:

> 1. La pregunta de entrega ofrece el nuevo modo, y elegirlo deja un PR abierto, con CI verde y hallazgos atendidos, sin merge ni auto-merge.
> 2. Con bots de revisión en el repositorio, el flujo espera a que terminen sobre el último commit, verifica cada hallazgo contra el código y responde con la razón los que descarta.
> 3. Un comentario o check anterior al último push no provoca cambios.
> 4. Antes de un merge o de declarar el PR listo, un PR BEHIND o DIRTY, o con commits nuevos de la base que tocan sus archivos tras el último verde, no avanza hasta rebasar y volver a pasar el CI.
> 5. Sin cuota de GraphQL, la lectura del PR y de sus checks sigue funcionando por REST.
> 6. El cambio pasa la revisión de todos los lugares que `AGENTS.md` exige, y `README.md`/`site/` se actualizan si describen los modos de entrega.

Cambios frente a la issue:
- Criterio 2: se acota por D6 A (se espera a los bots que publican un check; los que solo comentan no frenan).
- Criterio 3: se limita a bots y checks; un hilo abierto de un revisor humano sigue vigente aunque sea anterior al último push.
- Criterio 4: ya se cumple antes del merge (`git-workflow` paso 8 exige que la rama contenga la base, más estricto que BEHIND, y un conflicto al traer la base se pregunta en el paso 2), así que se reescribe como AC5: la misma comprobación antes de declarar `ready`.
- La issue deja fuera un vigía propio, la firma de comentarios, tomar PRs ajenos, refrescar PRs viejos, cambiar la regla del merge admin y la revisión con otro proveedor.

## Objetivo

Hoy, después del push, Hive solo sabe esperar el CI y hacer el merge, o detenerse antes del push. Con el cambio, el usuario puede pedir un PR listo para su merge; los bots de revisión se leen y se atienden con las mismas reglas de la revisión dedicada (una ronda, solo lo bloqueante), solo sobre lo que corresponde al último commit; y cuando GitHub agota la cuota de GraphQL la lectura y la espera siguen por REST.

## Alcance y aceptación

Dentro: la guía global («Delivery»), `flow-verify` paso 6, `git-workflow` pasos 6–9, `flow-close` (punto de cierre, paso 1 del visor y reporte final), las páginas del sitio que describen la integración y el cierre, la sección «Runtime resolution» de `instruction-resources.md` completa (D11) y `CHANGELOG.md`.

Fuera: `tooling/` (el validador ya rechaza cualquier `Delivery` que no sea `direct-base`), `unattended-delegation` (pregunta el modo igual que hoy y `ready` no añade efectos), README y `llms.txt` (no nombran los modos), y qué modo de entrega usa una sesión posterior al archivar el change record de un PR que mergeó otra persona (el hueco ya existe con el revisor humano).

- AC1 — La pregunta de entrega ofrece `ready`, y la guía global le asigna sus efectos: revisar, hacer push, abrir el PR como no borrador, esperar el CI, atender hallazgos y detenerse sin merge · falso hoy porque `content/guidance/global.md:74-77` solo define `interactive`, `automatic`, `direct-base` y `hold`.
- AC2 — Bajo `ready`, la integración termina con el PR abierto y sin merge: el agente no hace merge, no activa auto-merge ni lo añade a una cola de merge · falso hoy porque `content/skills/flow-verify/SKILL.md:19` siempre termina en merge.
- AC3 — Al terminar la espera del CI, las revisiones y comentarios que corresponden al último commit se leen una vez, de personas y de bots; cada hallazgo de un bot se contrasta con el código antes de cambiar nada y el descartado recibe respuesta con la razón · falso hoy porque `content/skills/git-workflow/SKILL.md:19` solo trata revisores asignados o solicitados y `rg -i '\bbots?\b' content/skills content/guidance` no da resultados.
- AC4 — Una revisión o un comentario de bot cuyo `commit_id` no es el commit de cabeza, o un comentario de bot creado antes del último push de este trabajo, no provoca cambios; un hilo abierto de un revisor humano sí, aunque sea anterior · falso hoy porque ninguna regla de `content/` distingue lo anterior al último push (`rg -i 'latest push|last push|newer than' content` sin resultados).
- AC5 — Antes de declarar el PR `ready` se hace la misma comprobación de base al día que antes del merge (`git fetch` y `merge-base --is-ancestor`; si falla, traer la base, hacer push y volver a esperar el CI), y si no queda al día el PR se reporta como no listo; el reporte nombra el commit de base comprobado · falso hoy porque el paso 8 de `content/skills/git-workflow/SKILL.md` la exige solo para el merge.
- AC6 — Cuando `gh` falla con el error de cuota de GraphQL, el PR, sus checks, revisiones y comentarios se leen por REST y la espera del CI usa `gh run watch <run-id> --exit-status`, nunca consultas en bucle · falso hoy porque `rg -i 'graphql|rate limit' content` no da resultados, y `gh pr checks` lee por GraphQL (observado con `GH_DEBUG=api` el 2026-11-02).
- AC7 — Un PR dejado bajo `ready` es el punto de cierre: el ticket sigue en progreso y la sesión que lo encuentre mergeado ejecuta el cierre · falso hoy porque `content/skills/flow-close/SKILL.md` («The completion point is…») solo lo dice de un PR con revisor humano asignado.
- AC8 — La pregunta de revisión recomienda siempre una revisión bajo `ready`, como bajo `automatic` · falso hoy porque `content/guidance/global.md:72` dice «always under `automatic`».
- AC9 — Las páginas `site/src/content/docs/flows/verify.md` y `close.md` describen la parada de `ready` · falso hoy porque `verify.md:30` dice que la integración siempre termina en merge y `close.md:10` solo nombra el revisor humano.
- AC10 — Cada ruta de `content/` que nombra `_support/docs/architecture/instruction-resources.md` existe, y cada entrada de su lista de hogares canónicos apunta al lugar donde hoy vive su regla (la de entrega, a la guía global «Delivery») · falso hoy porque 18 rutas de esa lista no existen, entre ellas `flow-plan/references/delivery-decisions.md` (línea 33), más las de `hive-core/references/` de la línea 37 y `flow-research/references/delegation.md` de la línea 13; la skill `hive-core` ya no existe en `content/skills/`.
- AC11 — Los hallazgos de bots después del push siguen el paso 3 de `flow-verify`: una ronda de correcciones solo para lo bloqueante, lo no bloqueante va a los hallazgos del cierre, cada commit de corrección lleva una línea `Unreviewed: <reason>`, y un hallazgo sobre un commit de corrección se reporta sin abrir otra ronda · falso hoy porque `git-workflow` paso 7 dice «Apply the blocking comments first and the rest in any order» y nada limita las rondas después del push.
- AC12 — Antes de abrir el PR bajo `ready`, una automatización del repositorio que haría merge del PR (un workflow o una app de merge automático) se lee y se pregunta antes de abrirlo · falso hoy porque `git-workflow` paso 6 solo pide leer la automatización que «triggers CI, review, preview or deployment».
- AC13 — Bajo `ready`, o con revisor humano asignado, el cierre deja la sesión del visor en `in_review`, no cerrada · falso hoy porque `flow-close` paso 1 solo conoce cerrada (integrada o abandonada) y pausada.
- AC14 — Un commit en la rama del PR que no hizo este trabajo (por ejemplo, el autofix de un bot) se nombra y se pregunta antes de seguir, y no cuenta como push de este trabajo · falso hoy porque ninguna regla de `content/` trata commits ajenos que aparecen en la rama remota del PR después del push.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

Leído el 2026-11-02 sobre `26883dcc`. Ese commit difiere de `development` en `CHANGELOG.md`, `content/skills/flow-close/SKILL.md`, `_support/docs/architecture/workspace-and-artifacts.md` y `content/skills/flow-walkthrough/references/walkthrough-manifest.md`; por eso el plan cita `flow-close` por su texto y no por número de línea, y por D9 B los dos trabajos comparten la entrada `[Unreleased]` de `CHANGELOG.md`.
- `content/guidance/global.md:70-79` («Delivery»), 21.924 bytes.
- `content/skills/flow-verify/SKILL.md` (4.854 bytes), `content/skills/git-workflow/SKILL.md` (9.406 bytes), `content/skills/flow-close/SKILL.md`, `content/skills/unattended-delegation/SKILL.md:13-19`.
- Todo lo que nombra los modos, la integración, la revisión o el punto de cierre, buscado con `rg`: `flow-deploy` (su promoción es su propia elección de entrega; su paso 26 revisa los commits `Unreviewed:` en la promoción), `flow-plan/references/plan-handoff.md:16`, la plantilla `proposal.md:47`, `harness-audit/references/apply.md:8`, `flow-build:16`, `site/src/content/docs/flows/build.md:29` y `getting-started.md:38-41` son genéricos y no cambian. `tooling/cli/doctor_project.go:397-398` rechaza `Delivery: ready`. `tests/pilot` (`regression.go:543`) solo reacciona a merges.
- Requisitos vigentes: se buscaron «pull request», «merge», «delivery mode» y «bot» en `_support/openspec/specs/` el 2026-11-02. `walkthrough-viewer` (líneas 307-316) fija la etapa Delivery y su espera de CI, que nunca hace merge, y ya tiene el estado `in_review`; ninguna capacidad fija los modos de entrega. No hay deltas.
- Origen: research del 2026-11-01 sobre referencias externas de skills de entrega de PRs (notas locales de referencias). Las reglas se redactan desde cero; si se adapta texto, aplica la atribución que pide su licencia.

## Entrega

`interactive` (D7 A, 2026-11-02): el build se detiene antes del push para que el usuario valide el diff; después push, PR a `development` y merge. Revisión de código: nativa, Claude Code `/code-review`, sobre el candidato final antes de esa validación (D8 A). Incluye `26883dcc` en el mismo PR (D9 B); no autoriza otros commits ajenos, un release, `hive update` fuera de lo que `AGENTS.md` fija para el checkout del mantenedor, ni cambios en `tooling/`.

## Estado de la revisión y avance

2026-11-02 — Plan escrito. `hive-review-plan` encontró 3 bloqueantes, 8 a corregir y 5 menores; todos se verificaron y se aplicaron:
- el filtro de frescura se limita a bots y checks (AC4);
- los hallazgos de bots siguen la ronda única de `flow-verify` (AC11);
- se lee la automatización que haría merge (AC12);
- la parada de `ready` en el paso 8 no hereda instrucciones de merge;
- el estado del visor queda en `in_review` (AC13);
- el reporte vive solo en el reporte final de `flow-close`;
- la frescura se mide con `commit_id` y la hora del push;
- REST cubre la espera, las revisiones y los comentarios (AC6);
- AC10 se acotó a la línea 33 y después, por D11, se amplió a toda la lista de hogares canónicos, con una comprobación que recorre las rutas;
- se citan las líneas por texto;
- se tratan los commits ajenos en la rama (AC14);
- se aclaran el nombre del modo, la razón de la recomendación de revisión y el commit de base en el reporte.

Queda fuera el modo de entrega del archivado en una sesión posterior.

2026-11-02 — Build en 6 commits (`4aaba3fe`…`fe095359`), dos rondas de revisión (`f6f32661`, `d7818c88`), validado por el usuario (D12 A) e integrado con el PR #313 (`df2591f9`).
