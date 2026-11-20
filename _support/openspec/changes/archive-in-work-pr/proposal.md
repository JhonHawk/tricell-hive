# archive-in-work-pr — El archivado de OpenSpec y el material de soporte se entregan por lotes en un PR de higiene

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` @ a5691ef8 · rama `feat/archive-in-work-pr` (worktree `.claude/worktrees/archive-in-work-pr`) |
| Playbook | feature |

## Decisions

2026-11-20
- **Investigación a** — ¿Se cambia la regla para que el archivado vaya en el mismo PR que el código? Sí: planear el cambio, con las excepciones que siguen archivando en el cierre.
- **D9 A** — ¿En qué momento se archiva dentro de la rama del trabajo? Antes del primer push, en `flow-verify`, tras la validación y la ronda de fixes y antes del CI local; el detalle del estado lleva el nombre de la rama, porque el PR aún no existe.
- **D10 B** — ¿Entran los commits de material de soporte que hoy decide el cierre? Sí: la decisión de commitear, archivar o promover ese material también se toma antes del push y viaja en el mismo PR; las eliminaciones siguen en el cierre.
- **D11 A** — ¿Cómo se resuelven en `automatic` y `ready`, que no se detienen antes del push, la decisión del material de soporte y la de archivar con tareas sin marcar? Se aplica la opción recomendada de cada ítem de soporte y se reporta con cómo revertirla; con tareas sin marcar no se archiva y lo hace el cierre en su rama. En `interactive` ambas preguntas van en la ronda de validación de `flow-verify`.
- **D12 A** — ¿Adónde van las líneas del checkpoint escritas tras el archivado? Con el siguiente commit de fix, si lo hay; las que quedan tras el último push no se commitean: el cierre las descarta con `git restore`, porque solo tocan el checkpoint de una carpeta archivada, y el cuerpo del PR y la memoria guardan ese resultado.
- **supuesto** — Las eliminaciones de material de soporte se quedan en el cierre, no antes del push: su condición («un PR fusionado cuyo cuerpo reemplaza el material») solo se cumple después del merge. Alternativa descartada: preguntarlas antes del push y aplicarlas en el cierre. Se revierte moviendo la línea de **Delete** al paso previo al push.
- **supuesto** — Cuando dos ramas cambian el mismo requisito y, al traer la base (antes del push o, en `merge gate`, antes del merge o de la parada `ready`), llega el archivo de la otra con hora posterior, la rama renombra su carpeta archivada a la hora UTC actual y reescribe su delta de ese requisito al bloque resuelto del spec; `spec_deltas_test.go` lo detecta en local (no hay CI que corra `go test`). Alternativa descartada: volver al archivado tras el merge para esos casos. Se revierte quitando esa línea de `hive-rule-close-hygiene`.
- **supuesto** — Un fix posterior al push que cambia un requisito ya fundido edita juntos el spec y el delta de la carpeta archivada, porque `spec_deltas_test.go` exige que coincidan. Alternativa descartada: editar solo el spec. Se revierte quitando esa línea.
- **supuesto** — Con entrega `direct-base` la rama del trabajo es la base: el detalle es el nombre de la base. Alternativa descartada: el SHA, que no existe antes del commit.
- **D13** — El archivado antes del push pasaba el presupuesto de lectura del ciclo plan→build por 2.708 B (961 B aun con #554). Respuesta del usuario: marcar en memoria el archivado pendiente y tener un archivado propio que mueva todos los recursos a la vez en su propio PR; se formuló como D14.
- **D14 A** — ¿Archivado por lotes en su propio PR en vez de dentro del PR de cada trabajo? Sí: el cierre ya no archiva ni commitea material de soporte; lo registra como pendiente en memoria, y la entrada de higiene de `flow-close` («haz higiene») archiva todos los change folders pendientes del repositorio y aplica las acciones de soporte pendientes en una rama desde la base y un PR. `flow-plan` lee los deltas pendientes de las capabilities que toca. **Replaces:** D9 A, D10 B, D11 A, D12 A
- **supuesto** — El repositorio es la fuente de verdad de lo pendiente: un change folder que sigue en `<Specs>/changes/` en la base está pendiente de archivo; la memoria guarda, un registro por cambio en `archive-pending/<change-id>` (fuera de `pending/`, que es trabajo dejado abierto a propósito, y para que dos cierres en paralelo no se pisen), lo que el repositorio no muestra: PR, commit del merge y acciones de soporte. Una carpeta sin registro se archiva solo si el usuario confirma que su trabajo está integrado. Alternativa descartada: un registro único `pending/archives`. Se revierte quitando la reconciliación del lote.
- **D15 A** — ¿Qué hace el cierre con las líneas del checkpoint escritas después del último push? Las descarta con `git restore` cuando solo tocan el checkpoint del change folder; el cuerpo del PR y la memoria guardan ese resultado.
- **D16 A** — ¿Cómo se autoriza la entrega del lote de higiene sin respuesta de entrega vigente? Pregunta solo el modo de entrega antes de su primer cambio (el lote es mecánico, sin pregunta de revisión); bajo `hold` deja todo pendiente, bajo `direct-base` hace push a la base, en los demás modos abre un PR.
- **D17 A** — ¿Specs en otro repositorio también pasa al lote? No: sigue archivando en el cierre como dice el `Delivery` de ese repositorio.
- **D18 A** — ¿El reporte de estado lista los archivados pendientes? Sí, una línea con su número y la respuesta «haz higiene».
- **supuesto** — En `ready`, cada registro se vacía cuando el archivado llega a la base (cierre de seguimiento del PR del lote), no al abrir el PR; el lote salta carpetas que ya archiva un PR de higiene abierto. Alternativa descartada: vaciar al abrir el PR. Se revierte quitando esa línea.
- **supuesto** — Entre carpetas del mismo lote que cambian el mismo requisito, el lote espera al minuto UTC siguiente antes de archivar la segunda, porque `apply_deltas.py` nombra el archivo con el minuto. Alternativa descartada: declarar perdido el orden. Se revierte quitando esa línea.
- **supuesto** — Las eliminaciones de material de soporte se siguen aplicando en el cierre: no necesitan PR. Solo commit, archive y promote esperan al lote. Alternativa descartada: llevar también las eliminaciones al lote. Se revierte moviendo **Delete** al lote.
- **supuesto** — El lote archiva en el orden en que cada carpeta llegó a la base (`git log --diff-filter=A` del `proposal.md`), para que un delta escrito sobre otro pendiente se aplique encima. Alternativa descartada: orden alfabético. Se revierte quitando esa línea.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «por cierto, ahorita como esta la regla de commit de los specs o documentación que generamos con openspec? creo que se esta forzando a que vaya en su propio pr (y debe ir junto)» (2026-11-20)

Respuesta a la investigación: «A» (planear el cambio). Sin ticket.

## Objective

Hoy el change folder (`proposal.md`, `design.md`, `tasks.md`, delta) se commitea con el código, pero su archivado (fundir los deltas en `specs/`, marcar la propuesta como completada y moverla a `changes/archive/`) y los commits del material de soporte que decide el cierre van después del merge, en una rama nueva desde la base: un segundo PR por trabajo (por ejemplo #551 y #552). Al terminar, el cierre registra ambos como pendientes y no abre rama; una petición de higiene entrega todo lo pendiente del repositorio en un solo PR.

## Scope and acceptance

Dentro: `hive-rule-close-hygiene` («Archive the change folder» pasa a registrar lo pendiente), `flow-close/SKILL.md` (pasos 2 y 4 y la entrada de higiene), una referencia nueva `flow-close/references/batch-archive.md`, `flow-close/references/support-material.md`, `flow-close/references/branch-cleanup.md`, `hive-rule-close-modes`, `hive-rule-close-footer`, `hive-rule-close-report`, `hive-mode/playbooks/ticket-status.md`, `flow-plan/SKILL.md` pasos 1 y «Where the plan lives», `flow-build/references/spec-delta.md`, `git-workflow/SKILL.md` paso 4, el spec `workflow-lifecycle`, `README.md`, `llms.txt`, las páginas del sitio `flows/close.md`, `flows/plan.md` y `flows/overview.md`, `project-layout/references/layout.md`, `_support/docs/architecture/instruction-resources.md` y `workspace-and-artifacts.md`, `rule-enforcers.md` y `CHANGELOG.md`.

Fuera: `apply_deltas.py` y `check_change.py` (no cambian); Specs en otro repositorio (D17 A: sigue archivando en el cierre); el global (`content/guidance/global.md`, en el PR #554); `flow-verify`.

- AC1 — Un trabajo con change folder en este repositorio, todas sus tareas marcadas, llega a su parada `ready` con el archivado en su propia rama: el diff del PR funde el delta en `specs/<capability>/spec.md`, mueve la carpeta a `changes/archive/YYYY-MM-DD-HHMM-<change-id>/` con `proposal.md` en `Completado · <rama>`, y el cierre posterior no abre otra rama ni PR para el archivado · false today because `hive-rule-close-hygiene` archiva «on a new branch from the base» después del merge (PR #551 y su archivo #552) · Observed in: el PR de este mismo cambio, archivado siguiendo el procedimiento nuevo a mano (T5), y su cierre de seguimiento; la parte del cierre y el uso por la regla instalada quedan sin verificar en la parada `ready` · **Retired:** D14 A reemplaza el archivado dentro del PR por lotes
- AC2 — Los ítems de soporte que el usuario decide commitear, archivar o promover antes del push son commits de la rama del trabajo y aparecen en el diff de su PR; la pregunta del cierre solo lista eliminaciones y lo creado después del push · false today because `support-material.md` commitea esos ítems «on a new branch from the base» en el cierre · Observed in: la primera sesión real con material de soporte tras instalar el cambio (queda sin verificar en la parada `ready`) · **Retired:** D14 A reemplaza el archivado dentro del PR por lotes
- AC3 — Las páginas `site/src/content/docs/flows/close.md` y `flows/verify.md` dicen que el archivado y el material de soporte decidido viajan en el PR del trabajo antes del push, y nombran las excepciones · false today because `close.md` dice que el archivado corre «when you choose to close» y `verify.md` no lo menciona · Observed in: las dos páginas del sitio en la rama · **Retired:** D14 A reemplaza el archivado dentro del PR por lotes
- AC4 — Al cerrar un trabajo integrado con change folder en su propio repositorio, el cierre no archiva ni abre rama o PR para el archivado ni para el material de soporte que se decida commitear, archivar o promover: escribe el registro de memoria `archive-pending/<change-id>` con la carpeta, su PR y commit de merge y esas acciones, descarta las ediciones sin commit que solo tocan el checkpoint, y el informe y el pie nombran cuántos archivados quedan pendientes y la respuesta «haz higiene» · false today because `hive-rule-close-hygiene` archiva «on a new branch from the base» en cada cierre (PR #551 y su archivo #552) · Observed in: el cierre de seguimiento de este mismo cambio cuando el usuario fusione su PR (sin verificar en la parada `ready`)
- AC5 — Una petición de higiene archiva, en una rama desde la base y una sola entrega bajo el modo vigente (un PR, o push a la base con `direct-base`), cada change folder que sigue en `<Specs>/changes/` de la base con registro o con su trabajo confirmado como integrado, en el orden en que llegó a la base, aplica las acciones de soporte registradas y vacía cada registro cuando el archivado llega a la base · false today because la higiene corre el cierre de la sesión, sin lote, y cada trabajo archiva en su propio PR · Observed in: la primera petición de higiene en este repositorio después de instalar el cambio, con este change folder pendiente (sin verificar en la parada `ready`)
- AC6 — El paso 1 de `flow-plan` y `flow-build/references/spec-delta.md` leen, además de los specs, los deltas de los change folders que siguen en `<Specs>/changes/` de la base para las capabilities que tocan · false today because el paso 1 dice «the existing specs the change touches» solamente · Observed in: el texto de `content/skills/flow-plan/SKILL.md` en la rama y el primer plan real sobre una capability con delta pendiente (sin verificar en la parada `ready`)
- AC7 — Las páginas `site/src/content/docs/flows/close.md`, `flows/plan.md` y `flows/overview.md`, `README.md` y `llms.txt` dicen que el cierre deja el archivado pendiente y que «haz higiene» lo entrega por lotes en un PR · false today because `close.md` dice que el cierre archiva el registro y `plan.md` que «Close records them and archives the folder» · Observed in: las dos páginas del sitio en la rama · **Replaces:** AC3

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/hive-rule-close-hygiene/SKILL.md` «Archive the change folder»: «on a new branch from the base … `--detail "<PR, merge SHA>"`» (a5691ef8).
- `content/skills/flow-close/references/support-material.md`: commit «with step 4's move or … in one commit that touches only the support home; either goes on a new branch from the base».
- `content/skills/git-workflow/SKILL.md` paso 4: «commit it on the work's branch before the push … or after the merge with the close's archive».
- `content/skills/flow-close/scripts/apply_deltas.py`: `--detail` opcional, una línea libre; rechaza una carpeta fuera de `<Specs>/changes/`.
- `_support/openspec/specs/workflow-lifecycle/spec.md`: «Archive confirms unchecked tasks» y «Deltas merge deterministically at archive» dicen «the close SHALL…».
- `tests/content/spec_deltas_test.go`: exige que el delta de la carpeta archivada más reciente de cada requisito coincida con el spec.
- `tests/content/reading_footprint_test.go`: el ciclo plan→build→verify→close no puede leer más que su línea base (184.489 B); hoy lee 184.487 B, un margen de 2 B (cálculo del revisor del plan, 2026-11-20); el archivado antes del push lo pasaba por 2.708 B (D13).
- `hive-rule-close-modes`: «a hygiene request» ya corre todos los pasos del cierre; `flow-close` se usa «when the user asks for session hygiene, such as "haz higiene"».
- Guía de OpenSpec (`docs/team-workflow.md`, Context7 `/fission-ai/openspec`, 2026-11-20): recomienda archivar tras el merge, «often a tiny follow-up commit or a scheduled cleanup».
- Historia: `flow-close` del 2026-09-29 (7e8a627f) ya archivaba tras integrar, con «a small closing pull request»; pstack (`skills/poteto-mode/playbooks/multi-phase-plan.md`) no tiene specs vivos ni archivado.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`ready` y revisión nativa `/code-review`, respondidas en esta sesión para este repositorio (trabajo global-guidance-slimming) y vigentes: push, PR no borrador, espera de CI y respuesta a sus revisiones, y parada con el PR abierto para que el usuario lo fusione. No autoriza el merge, auto-merge ni cola de merge.

## Review status and progress

2026-11-20 — Plan escrito.
2026-11-20 — revisión del plan (hive-review-plan): 2 P1, 4 P2, 5 P3. P1 llevados al usuario como D11 y D12 (respondidos A y A). P2 aplicados: fix posterior al push edita spec y delta archivado; colisión con otra rama renombra y reescribe el delta; bloque MODIFIED «Archived change order within a day» y excepción de Specs en otro repositorio; cuatro lugares más en el alcance (sitio `plan.md`, `layout.md`, `instruction-resources.md`, `description` de close-hygiene). P3 aplicados: pasos de copia y `git restore` solo en las excepciones; detalle con `direct-base`; AC1 (cierre) y AC2 sin verificar en la parada; impacto en close-modes corregido; ruta archivada en el checkpoint.
2026-11-20 — build T1–T4 hecho y luego revertido (f18f1977): el archivado antes del push pasaba `TestRequiredProcedureReadingFootprint` (D13). Revisión del plan por D14 A: archivado por lotes en un PR de higiene; reemplaza D9–D12, retira AC1–AC3, añade AC4–AC7, reescribe tareas y delta.
2026-11-20 — segunda revisión del plan (hive-review-plan, revisión D14): 2 P1, 7 P2, 8 P3. P1 y dos P2/P3 llevados al usuario como D15–D18 (todos A). Aplicados: registro por cambio `archive-pending/<change-id>`; vaciado al llegar a la base; respuesta en `close-footer` e informe; `spec-delta.md` lee deltas pendientes; `README.md`, `llms.txt`, `overview.md`, `workspace-and-artifacts.md`, `branch-cleanup.md` en el alcance; margen de 2 B del presupuesto con tope por tarea; minuto UTC entre carpetas del mismo requisito; carpeta sin registro solo con confirmación.
