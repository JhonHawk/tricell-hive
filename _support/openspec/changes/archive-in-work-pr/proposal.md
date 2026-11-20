# archive-in-work-pr — El archivado de OpenSpec y el material de soporte viajan en el PR del trabajo

| Campo | Valor |
| --- | --- |
| Estado | In build |
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

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «por cierto, ahorita como esta la regla de commit de los specs o documentación que generamos con openspec? creo que se esta forzando a que vaya en su propio pr (y debe ir junto)» (2026-11-20)

Respuesta a la investigación: «A» (planear el cambio). Sin ticket.

## Objective

Hoy el change folder (`proposal.md`, `design.md`, `tasks.md`, delta) se commitea con el código, pero su archivado (fundir los deltas en `specs/`, marcar la propuesta como completada y moverla a `changes/archive/`) y los commits del material de soporte que decide el cierre van después del merge, en una rama nueva desde la base: un segundo PR por trabajo (por ejemplo #551 y #552). Al terminar, ambos van como commits de la rama del trabajo antes de su primer push, y el cierre solo abre otra rama para las excepciones.

## Scope and acceptance

Dentro: `hive-rule-close-hygiene` («Archive the change folder»), `flow-close/references/support-material.md`, `flow-close/SKILL.md` pasos 2 y 4, `hive-rule-close-modes`, `flow-verify/SKILL.md` paso 5, `git-workflow/SKILL.md` pasos 4 y 6, `flow-build/references/spec-delta.md`, el spec `workflow-lifecycle`, las páginas del sitio `flows/close.md`, `flows/verify.md` y `flows/plan.md`, `project-layout/references/layout.md`, `_support/docs/architecture/instruction-resources.md`, `rule-enforcers.md` y `CHANGELOG.md`.

Fuera: `apply_deltas.py` y `check_change.py` (no cambian: `--detail` ya es texto libre y opcional); Specs en otro repositorio (sigue como hoy: se commitea allí en el punto de cierre); la entrega `hold` (no commitea, sigue en el cierre); el global (`content/guidance/global.md`) no se toca.

Excepciones que siguen archivando en el cierre, en una rama nueva desde la base: tareas sin marcar con un «no» a archivar igual o bajo delegación desatendida; trabajo ya fusionado sin archivar; Specs en otro repositorio; `hold`; material de soporte creado después del push.

- AC1 — Un trabajo con change folder en este repositorio, todas sus tareas marcadas, llega a su parada `ready` con el archivado en su propia rama: el diff del PR funde el delta en `specs/<capability>/spec.md`, mueve la carpeta a `changes/archive/YYYY-MM-DD-HHMM-<change-id>/` con `proposal.md` en `Completado · <rama>`, y el cierre posterior no abre otra rama ni PR para el archivado · false today because `hive-rule-close-hygiene` archiva «on a new branch from the base» después del merge (PR #551 y su archivo #552) · Observed in: el PR de este mismo cambio, archivado siguiendo el procedimiento nuevo a mano (T5), y su cierre de seguimiento; la parte del cierre y el uso por la regla instalada quedan sin verificar en la parada `ready`
- AC2 — Los ítems de soporte que el usuario decide commitear, archivar o promover antes del push son commits de la rama del trabajo y aparecen en el diff de su PR; la pregunta del cierre solo lista eliminaciones y lo creado después del push · false today because `support-material.md` commitea esos ítems «on a new branch from the base» en el cierre · Observed in: la primera sesión real con material de soporte tras instalar el cambio (queda sin verificar en la parada `ready`)
- AC3 — Las páginas `site/src/content/docs/flows/close.md` y `flows/verify.md` dicen que el archivado y el material de soporte decidido viajan en el PR del trabajo antes del push, y nombran las excepciones · false today because `close.md` dice que el archivado corre «when you choose to close» y `verify.md` no lo menciona · Observed in: las dos páginas del sitio en la rama

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/hive-rule-close-hygiene/SKILL.md` «Archive the change folder»: «on a new branch from the base … `--detail "<PR, merge SHA>"`» (a5691ef8).
- `content/skills/flow-close/references/support-material.md`: commit «with step 4's move or … in one commit that touches only the support home; either goes on a new branch from the base».
- `content/skills/git-workflow/SKILL.md` paso 4: «commit it on the work's branch before the push … or after the merge with the close's archive».
- `content/skills/flow-close/scripts/apply_deltas.py`: `--detail` opcional, una línea libre; rechaza una carpeta fuera de `<Specs>/changes/`.
- `_support/openspec/specs/workflow-lifecycle/spec.md`: «Archive confirms unchecked tasks» y «Deltas merge deterministically at archive» dicen «the close SHALL…».
- `tests/content/spec_deltas_test.go`: exige que el delta de la carpeta archivada más reciente de cada requisito coincida con el spec.
- Historia: `flow-close` del 2026-09-29 (7e8a627f) ya archivaba tras integrar, con «a small closing pull request»; pstack (`skills/poteto-mode/playbooks/multi-phase-plan.md`) no tiene specs vivos ni archivado.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`ready` y revisión nativa `/code-review`, respondidas en esta sesión para este repositorio (trabajo global-guidance-slimming) y vigentes: push, PR no borrador, espera de CI y respuesta a sus revisiones, y parada con el PR abierto para que el usuario lo fusione. No autoriza el merge, auto-merge ni cola de merge.

## Review status and progress

2026-11-20 — Plan escrito.
2026-11-20 — revisión del plan (hive-review-plan): 2 P1, 4 P2, 5 P3. P1 llevados al usuario como D11 y D12 (respondidos A y A). P2 aplicados: fix posterior al push edita spec y delta archivado; colisión con otra rama renombra y reescribe el delta; bloque MODIFIED «Archived change order within a day» y excepción de Specs en otro repositorio; cuatro lugares más en el alcance (sitio `plan.md`, `layout.md`, `instruction-resources.md`, `description` de close-hygiene). P3 aplicados: pasos de copia y `git restore` solo en las excepciones; detalle con `direct-base`; AC1 (cierre) y AC2 sin verificar en la parada; impacto en close-modes corregido; ruta archivada en el checkpoint.
