# Tasks

Base: tricell-hive `development` @ a5691ef8, rama `feat/archive-in-work-pr`, worktree `.claude/worktrees/archive-in-work-pr`. Todas las tareas corren en el hilo principal: son ediciones de reglas enlazadas entre sí, donde cada texto depende de cómo quedó el anterior, y suman unos 13 archivos. T1–T5 son del enfoque dentro del PR (D9 A), retirado por D14 A y revertido en f18f1977; las tareas vigentes son T6–T10.

### T1 — (retirada) El archivado pasa a la rama del trabajo antes del push

- [x] `hive-rule-close-hygiene` «Archive the change folder» fija los dos momentos (antes del push de la rama del trabajo, o en el cierre para las excepciones), el detalle `--detail "<branch>"`, el renombrado a la hora UTC actual cuando la base trae un archivo posterior del mismo requisito, que las líneas de checkpoint siguen en el `tasks.md` archivado y el checkpoint registra la ruta archivada para recuperarla tras una compactación, que un fix posterior al push edita juntos el spec y el delta archivado, que al traer la base con un archivo posterior del mismo requisito se renombra la carpeta y se reescribe su delta, que con `direct-base` el detalle es el nombre de la base, y que copiar ediciones sin commit y `git restore` aplican solo a las excepciones del cierre; su `description` nombra el momento previo al push; con tareas sin marcar, en `automatic` y `ready` no archiva y no pregunta; las líneas de checkpoint posteriores van con el siguiente fix y el cierre descarta con `git restore` las que quedan tras el último push si solo tocan el checkpoint archivado
- [x] `git-workflow` paso 4 deja de decir «or after the merge with the close's archive» y apunta al paso previo al push; `flow-build/references/spec-delta.md` paso 4 apunta a ese paso en vez de a `flow-close`

**Closes:** none · **Retired:** D14 A (era AC1)

**Depends on:** none

**Locations:** `content/skills/hive-rule-close-hygiene/SKILL.md`, `content/skills/git-workflow/SKILL.md`, `content/skills/flow-build/references/spec-delta.md`

**Execution:** main thread

**Changes:** solo texto de reglas; `apply_deltas.py` y `check_change.py` no cambian. Specs en otro repositorio y `hold` conservan su texto actual.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (tamaños de skill y hoja, frases de casa única); `rg -n 'merge SHA|after the merge with the close' content/` sin coincidencias.

### T2 — (retirada) El material de soporte decidido viaja en el mismo PR

- [x] `flow-close/references/support-material.md` gana la sección «Before the push»: inventario, solo commit, archive y promote; en `interactive` la decisión va en la ronda de validación de `flow-verify`, y en `automatic` y `ready` se aplica la opción recomendada y se reporta con cómo revertirla; commit en la rama del trabajo; el cierre pregunta solo eliminaciones y lo creado después del push, y usa una rama nueva desde la base solo para eso
- [x] `flow-close/SKILL.md` pasos 2 y 4 y `hive-rule-close-modes` reflejan que el paso previo al push ya decidió esos ítems y archivó la carpeta

**Closes:** none · **Retired:** D14 A (era AC2)

**Depends on:** T1

**Locations:** `content/skills/flow-close/references/support-material.md`, `content/skills/flow-close/SKILL.md`, `content/skills/hive-rule-close-modes/SKILL.md`

**Execution:** main thread

**Changes:** delete y keep siguen en el cierre; `hold`, pausa y delegación desatendida conservan su comportamiento (la delegación responde con la opción recomendada en su paso 5).

**Verification:** `go -C <worktree> test ./tests/content/...` pasa; lectura cruzada de los tres archivos: un solo orden para el momento del push.

### T3 — (retirada) `flow-verify` ejecuta los registros antes del CI local

- [x] `flow-verify` paso 2 suma a la ronda de validación de `interactive` la decisión del material de soporte y la de archivar con tareas sin marcar; el paso 5 abre con el commit del material decidido y el archivado, enlazando `hive-rule-close-hygiene` y `support-material.md` directamente, y después asegura la evidencia del CI local sobre el candidato resultante

**Closes:** none · **Retired:** D14 A (era AC1, AC2)

**Depends on:** T1, T2

**Locations:** `content/skills/flow-verify/SKILL.md`

**Execution:** main thread

**Changes:** sin renumerar pasos: `git-workflow` paso 6 ya cita «the checks of flow-verify step 5», así que un cambio mecánico sin verify también los hace.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa; `rg -n 'flow-verify step 5|flow-verify\]\(.*\) step 5' content/` lista solo referencias que siguen siendo correctas.

### T4 — (retirada) Spec, sitio, registro de enforcers y changelog

- [x] `site/src/content/docs/flows/close.md`, `flows/verify.md` y `flows/plan.md` describen el momento nuevo y las excepciones; `project-layout/references/layout.md` y `instruction-resources.md` dejan de atribuir el archivado solo al cierre
- [x] `_support/docs/harness-engineering/rule-enforcers.md`: la fila «Deltas merge into the specs before a change folder is archived» apunta a la casa nueva
- [x] `CHANGELOG.md` «Unreleased» lo anota en «Changed»

**Closes:** none · **Retired:** D14 A (era AC3)

**Depends on:** T1, T2, T3

**Locations:** `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/plan.md` (líneas 49 y 73), `content/skills/project-layout/references/layout.md` (línea 56), `_support/docs/architecture/instruction-resources.md` (línea 38), `content/skills/git-workflow/SKILL.md` (línea 23, «closing pull request», solo comprobar), `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** main thread

**Changes:** las páginas describen el comportamiento y enlazan la skill; no copian sus reglas.

**Verification:** `go -C <worktree> test ./...` pasa (incluye los enlaces del sitio); `rg -n -i 'archiv' site/src/content/docs/flows/close.md site/src/content/docs/flows/verify.md` muestra el momento nuevo.

### T5 — (retirada) Este cambio se archiva en su propia rama antes del push

Retirada sin hacer: tras la ronda de fixes de verify, este change folder se archivaba con el procedimiento nuevo en `feat/archive-in-work-pr` antes de su primer push: `check_change.py --archive`, `apply_deltas.py … --detail "feat/archive-in-work-pr"`, `--apply`

**Closes:** none · **Retired:** D14 A (era AC1)

**Depends on:** T4

**Locations:** `_support/openspec/specs/workflow-lifecycle/spec.md`, `_support/openspec/changes/archive/<fecha>-archive-in-work-pr/`

**Execution:** main thread

**Changes:** el cierre de seguimiento, cuando el usuario fusione, no abre rama ni PR para el archivado.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (`spec_deltas_test.go` con el archivo nuevo); el diff del PR muestra el delta fundido y la carpeta en `changes/archive/`.

### T6 — El cierre registra el archivado y el soporte como pendientes

- [ ] `hive-rule-close-hygiene` (título, `description` y «Archive the change folder»): con Specs en el repositorio del trabajo, el cierre no archiva; escribe el registro de memoria `archive-pending/<change-id>` (carpeta, PR, commit de merge, acciones de soporte) y descarta con `git restore` las ediciones sin commit que solo tocan el checkpoint (D15 A); con Specs en otro repositorio archiva como hoy (D17 A)
- [ ] `hive-rule-close-footer` gana la respuesta «haz higiene» con el número de archivados pendientes, y `hive-rule-close-report` la nombra en «Next» para los cierres que terminan sin pie
- [ ] `flow-close/references/branch-cleanup.md`: un archivado pendiente no es un cierre pendiente
- [ ] `flow-close/references/support-material.md`: las eliminaciones se aplican en el cierre; commit, archive y promote se registran en `archive-pending/<change-id>`, sin rama, con los ítems de un worktree movidos a la misma ruta del checkout principal
- [ ] `flow-close/SKILL.md` pasos 2 y 4, `hive-rule-close-modes` (paso 4 corre de inmediato) y `git-workflow` paso 4 («with the close's archive» pasa al lote de higiene)

**Closes:** AC4

**Depends on:** none

**Locations:** `content/skills/hive-rule-close-hygiene/SKILL.md`, `content/skills/hive-rule-close-footer/SKILL.md`, `content/skills/hive-rule-close-report/SKILL.md`, `content/skills/flow-close/references/support-material.md`, `content/skills/flow-close/references/branch-cleanup.md`, `content/skills/flow-close/SKILL.md`, `content/skills/hive-rule-close-modes/SKILL.md`, `content/skills/git-workflow/SKILL.md`

**Execution:** main thread

**Changes:** solo texto de reglas; `hold`, pausa y Specs en otro repositorio conservan su comportamiento. Presupuesto: el ciclo de `reading_footprint_test.go` tiene 2 B de margen, así que la suma de los archivos del ciclo que T6 toca (close-hygiene, close-modes, `flow-close/SKILL.md`, `support-material.md`, close-report) no crece más de 1 B; el párrafo de archivado de close-hygiene (937 B) es el que se reduce.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa, incluido `TestRequiredProcedureReadingFootprint` bajo su línea base; `rg -n "new branch from the base" content/skills/hive-rule-close-hygiene content/skills/flow-close/references/support-material.md` sin coincidencias.

### T7 — La higiene entrega los pendientes por lotes

- [ ] Referencia nueva `flow-close/references/batch-archive.md`: listar los change folders que siguen en `<Specs>/changes/` de la base y reconciliarlos con los registros `archive-pending/*` (sin registro: solo con confirmación del usuario; desatendido, pendiente), saltar los que ya archiva un PR de higiene abierto, `check_change.py --archive` con una sola ronda para tareas sin marcar, una rama desde la base creada en el checkout que tiene los ítems de soporte, `apply_deltas.py` en orden de llegada con `--detail "<PR, merge commit>"` y un minuto UTC nuevo entre carpetas que cambian el mismo requisito, las acciones de soporte registradas, la entrega bajo el modo vigente (sin respuesta, solo la pregunta de entrega; `hold` deja todo pendiente; `direct-base` push a la base; los demás un PR, D16 A), y vaciar cada registro cuando su archivado llega a la base
- [ ] `hive-mode/playbooks/ticket-status.md` lista el número de archivados pendientes con la respuesta «haz higiene» (D18 A)
- [ ] `flow-close/SKILL.md` la enlaza para una petición de higiene

**Closes:** AC5

**Depends on:** T6

**Locations:** `content/skills/flow-close/references/batch-archive.md`, `content/skills/flow-close/SKILL.md`, `content/skills/hive-mode/playbooks/ticket-status.md`

**Execution:** main thread

**Changes:** reutiliza `apply_deltas.py` y `check_change.py` sin cambios; la referencia no entra al ciclo plan→build; el enlace en `flow-close/SKILL.md` cuenta en el presupuesto de T6.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa (tamaño de referencia, enlaces); lectura cruzada con `hive-rule-close-hygiene` y `support-material.md`: un solo dueño del procedimiento del lote.

### T8 — El plan lee los deltas pendientes

- [ ] `flow-plan/SKILL.md` paso 1 lee también los deltas de los change folders que siguen en `<Specs>/changes/` de la base para las capabilities que toca, y escribe su delta contra el spec con ellos aplicados; «Where the plan lives» deja de decir «since the close merges its delta»
- [ ] `flow-build/references/spec-delta.md` lee los mismos deltas pendientes antes de escribir el suyo

**Closes:** AC6

**Depends on:** none

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-build/references/spec-delta.md`

**Execution:** main thread

**Changes:** una frase en el paso 1 y otra en `spec-delta.md`; ningún script. `flow-plan/SKILL.md` está en el ciclo del presupuesto: compensa sus bytes en el mismo archivo.

**Verification:** `go -C <worktree> test ./tests/content/...` pasa, incluido el presupuesto de lectura.

### T9 — Sitio, documentación y changelog

- [ ] `site/src/content/docs/flows/close.md`, `flows/plan.md` y `flows/overview.md`, `README.md` y `llms.txt` describen el pendiente y el lote de «haz higiene»; `project-layout/references/layout.md` (línea 56), `_support/docs/architecture/instruction-resources.md` y `workspace-and-artifacts.md` apuntan a la casa nueva
- [ ] `_support/docs/harness-engineering/rule-enforcers.md`: la fila «Deltas merge into the specs before a change folder is archived» apunta a `batch-archive.md`
- [ ] `CHANGELOG.md` «Unreleased» lo anota en «Changed»

**Closes:** AC7

**Depends on:** T6, T7, T8

**Locations:** `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/overview.md`, `README.md`, `llms.txt`, `content/skills/project-layout/references/layout.md`, `_support/docs/architecture/instruction-resources.md`, `_support/docs/architecture/workspace-and-artifacts.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** main thread

**Changes:** las páginas describen el comportamiento y enlazan la skill; no copian sus reglas.

**Verification:** `go -C <worktree> test ./...` pasa; `rg -n -i 'higiene|hygiene|pending' site/src/content/docs/flows/close.md site/src/content/docs/flows/plan.md` muestra el lote.

### T10 — Este cambio deja su propio archivado pendiente

- [ ] El change folder de este cambio no se archiva en su PR; en el cierre de seguimiento, cuando el usuario fusione, queda registrado en `pending/archives` y lo entrega la siguiente petición de higiene

**Closes:** AC4, AC5

**Depends on:** T9

**Locations:** ninguna en el repositorio; el registro de memoria del proyecto

**Execution:** main thread

**Changes:** ninguno en la rama.

**Verification:** en la parada `ready`, sin verificar: AC4 se observa en el cierre de seguimiento y AC5 en la primera higiene.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go -C <worktree> test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...` and `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` (native) | The review answer |

Sin parada de validación: la entrega es `ready`, que se detiene con el PR abierto para que el usuario lo fusione.

## Checkpoint

- Updated: 2026-11-20
- Stage: plan revised (D14 A) · ready for flow-build
- Last commit: f18f1977
- Open findings: none
- Next: T6
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/archive-in-work-pr · feat/archive-in-work-pr

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-20 — Plan escrito (D9 A, D10 B).
2026-11-20 — Revisión del plan aplicada; D11 A y D12 A.
2026-11-20 — T1 · 4db2f102 · unverified: none
2026-11-20 — T2 · c932024b · unverified: AC2 needs a real session with support material
2026-11-20 — T3 · 0006a46d · unverified: TestRequiredProcedureReadingFootprint over its baseline by 2,599 B until decided
2026-11-20 — T4 · 46e22039 · unverified: footprint test over baseline
2026-11-20 — D13: el presupuesto de lectura del ciclo bloqueó el enfoque; D14 A lo reemplaza por lotes. T1–T4 revertidos en f18f1977; T1–T5 retiradas; T6–T10 nuevas.
2026-11-20 — Segunda revisión del plan aplicada; D15–D18 A.
