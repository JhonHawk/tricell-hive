# Tasks

Base: tricell-hive `development` @ a5691ef8, rama `feat/archive-in-work-pr`, worktree `.claude/worktrees/archive-in-work-pr`. Todas las tareas corren en el hilo principal: son ediciones de reglas enlazadas entre sí, donde cada texto depende de cómo quedó el anterior, y suman unos 10 archivos.

### T1 — El archivado pasa a la rama del trabajo antes del push

- [ ] `hive-rule-close-hygiene` «Archive the change folder» fija los dos momentos (antes del push de la rama del trabajo, o en el cierre para las excepciones), el detalle `--detail "<branch>"`, el renombrado a la hora UTC actual cuando la base trae un archivo posterior del mismo requisito, que las líneas de checkpoint siguen en el `tasks.md` archivado y el checkpoint registra la ruta archivada para recuperarla tras una compactación, que un fix posterior al push edita juntos el spec y el delta archivado, que al traer la base con un archivo posterior del mismo requisito se renombra la carpeta y se reescribe su delta, que con `direct-base` el detalle es el nombre de la base, y que copiar ediciones sin commit y `git restore` aplican solo a las excepciones del cierre; su `description` nombra el momento previo al push; con tareas sin marcar, en `automatic` y `ready` no archiva y no pregunta (D11 A); las líneas de checkpoint posteriores van con el siguiente fix y el cierre descarta con `git restore` las que quedan tras el último push si solo tocan el checkpoint archivado (D12 A)
- [ ] `git-workflow` paso 4 deja de decir «or after the merge with the close's archive» y apunta al paso previo al push; `flow-build/references/spec-delta.md` paso 4 apunta a ese paso en vez de a `flow-close`

**Closes:** AC1

**Depends on:** none

**Locations:** `content/skills/hive-rule-close-hygiene/SKILL.md`, `content/skills/git-workflow/SKILL.md`, `content/skills/flow-build/references/spec-delta.md`

**Execution:** main thread

**Changes:** solo texto de reglas; `apply_deltas.py` y `check_change.py` no cambian. Specs en otro repositorio y `hold` conservan su texto actual.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (tamaños de skill y hoja, frases de casa única); `rg -n 'merge SHA|after the merge with the close' content/` sin coincidencias.

### T2 — El material de soporte decidido viaja en el mismo PR

- [ ] `flow-close/references/support-material.md` gana la sección «Before the push»: inventario, solo commit, archive y promote; en `interactive` la decisión va en la ronda de validación de `flow-verify`, y en `automatic` y `ready` se aplica la opción recomendada y se reporta con cómo revertirla (D11 A); commit en la rama del trabajo; el cierre pregunta solo eliminaciones y lo creado después del push, y usa una rama nueva desde la base solo para eso
- [ ] `flow-close/SKILL.md` pasos 2 y 4 y `hive-rule-close-modes` reflejan que el paso previo al push ya decidió esos ítems y archivó la carpeta

**Closes:** AC2

**Depends on:** T1

**Locations:** `content/skills/flow-close/references/support-material.md`, `content/skills/flow-close/SKILL.md`, `content/skills/hive-rule-close-modes/SKILL.md`

**Execution:** main thread

**Changes:** delete y keep siguen en el cierre; `hold`, pausa y delegación desatendida conservan su comportamiento (la delegación responde con la opción recomendada en su paso 5).

**Verification:** `go -C <worktree> test ./tests/content/...` pasa; lectura cruzada de los tres archivos: un solo orden para el momento del push.

### T3 — `flow-verify` ejecuta los registros antes del CI local

- [ ] `flow-verify` paso 2 suma a la ronda de validación de `interactive` la decisión del material de soporte y la de archivar con tareas sin marcar; el paso 5 abre con el commit del material decidido y el archivado, enlazando `hive-rule-close-hygiene` y `support-material.md` directamente, y después asegura la evidencia del CI local sobre el candidato resultante

**Closes:** AC1, AC2

**Depends on:** T1, T2

**Locations:** `content/skills/flow-verify/SKILL.md`

**Execution:** main thread

**Changes:** sin renumerar pasos: `git-workflow` paso 6 ya cita «the checks of flow-verify step 5», así que un cambio mecánico sin verify también los hace.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa; `rg -n 'flow-verify step 5|flow-verify\]\(.*\) step 5' content/` lista solo referencias que siguen siendo correctas.

### T4 — Spec, sitio, registro de enforcers y changelog

- [ ] `site/src/content/docs/flows/close.md`, `flows/verify.md` y `flows/plan.md` describen el momento nuevo y las excepciones; `project-layout/references/layout.md` y `instruction-resources.md` dejan de atribuir el archivado solo al cierre
- [ ] `_support/docs/harness-engineering/rule-enforcers.md`: la fila «Deltas merge into the specs before a change folder is archived» apunta a la casa nueva
- [ ] `CHANGELOG.md` «Unreleased» lo anota en «Changed»

**Closes:** AC3

**Depends on:** T1, T2, T3

**Locations:** `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/plan.md` (líneas 49 y 73), `content/skills/project-layout/references/layout.md` (línea 56), `_support/docs/architecture/instruction-resources.md` (línea 38), `content/skills/git-workflow/SKILL.md` (línea 23, «closing pull request», solo comprobar), `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** main thread

**Changes:** las páginas describen el comportamiento y enlazan la skill; no copian sus reglas.

**Verification:** `go -C <worktree> test ./...` pasa (incluye los enlaces del sitio); `rg -n -i 'archiv' site/src/content/docs/flows/close.md site/src/content/docs/flows/verify.md` muestra el momento nuevo.

### T5 — Este cambio se archiva en su propia rama antes del push

- [ ] Tras la ronda de fixes de verify, este change folder se archiva con el procedimiento nuevo en `feat/archive-in-work-pr` antes de su primer push: `check_change.py --archive`, `apply_deltas.py … --detail "feat/archive-in-work-pr"`, `--apply`

**Closes:** AC1

**Depends on:** T4

**Locations:** `_support/openspec/specs/workflow-lifecycle/spec.md`, `_support/openspec/changes/archive/<fecha>-archive-in-work-pr/`

**Execution:** main thread

**Changes:** el cierre de seguimiento, cuando el usuario fusione, no abre rama ni PR para el archivado.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (`spec_deltas_test.go` con el archivo nuevo); el diff del PR muestra el delta fundido y la carpeta en `changes/archive/`.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go -C <worktree> test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...` and `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` (native) | The review answer |

Sin parada de validación: la entrega es `ready`, que se detiene con el PR abierto para que el usuario lo fusione.

## Checkpoint

- Updated: 2026-11-20
- Stage: plan · ready for flow-build
- Next: T1

## Review status and progress

2026-11-20 — Plan escrito (D9 A, D10 B).
2026-11-20 — Revisión del plan aplicada; D11 A y D12 A.
