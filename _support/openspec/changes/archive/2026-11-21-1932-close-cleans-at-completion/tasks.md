# Tareas

Base `2eef1969` de `development`, rama `feat/close-cleans-at-completion`. T1 y T5 editan más de 3 archivos de guía con la forma ya fijada en `design.md`: corren en un hijo `hive-write-docs` cada una, una tras otra (T5 depende de T1). T2 y T3 escriben código en archivos disjuntos sin contrato compartido entre ellas: T2 en el hilo principal, T3 en un hijo `hive-write-tests`; corren a la vez entre sí, después de T5, para que ninguna verificación lea guía a medio escribir. T4 en el hilo principal. El hilo principal corre las verificaciones y hace los commits.

### T1 — El cierre limpia al llegar al merge, en su núcleo

- [x] `global.md`, `close-modes`, `flow-close`, `close-hygiene`, `support-material`, `batch-archive`, `report-endings` y `close-report` dicen lo que fija `design.md` «Forma propuesta».

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/guidance/global.md`, `content/skills/hive-core/references/close-modes.md`, `content/skills/hive-core/references/close-hygiene.md`, `content/skills/hive-core/references/report-endings.md`, `content/skills/hive-core/references/close-report.md`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-close/references/support-material.md`, `content/skills/flow-close/references/batch-archive.md`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** los textos de `design.md`; lo que crece en `close-modes` y `close-report` se compensa con lo que se quita (la fila duplicada, la pregunta del paso 2, las acciones Commit y Archive) para que los miembros del paquete de cierre sigan bajo 22.500 B. No cambia el movimiento de tickets ni los cierres de pausa y `hold`.

**Verification:** `go test ./tests/content/ -count=1` pasa (presupuestos de `global.md`, del paquete de etapa, de lectura, hogar único y enlaces). `rg -n 'Steps 2 and 5|\*\*Support\.\*\*|step 2.s question' content/` da solo `flow-verify/SKILL.md:27`, que es el paso 2 de `flow-verify` (en la base da 5: además `close-modes.md:7`, `:9`, `:17` y `batch-archive.md:12`). AC1 a AC4 se observan en los archivos instalados tras `hive update`, después de integrar.

### T5 — Las hojas que dependían del criterio viejo

- [x] `trackers`, `viewer-close`, `pr-captures`, `layout.md`, `archive.md`, `inventory.md`, `branch-cleanup.md` (incluida «Report»: la forma «kept `<rama>`: <razón>» y `stray_clones`), `unattended-delegation` y `workspace-and-artifacts.md` dejan de contradecir T1.

**Closes:** AC2

**Depends on:** T1

**Locations:** `content/skills/hive-core/references/trackers.md`, `content/skills/hive-core/references/viewer-close.md`, `content/skills/git-workflow/references/pr-captures.md`, `content/skills/project-layout/references/layout.md`, `content/skills/project-layout/references/archive.md`, `content/skills/project-layout/references/inventory.md`, `content/skills/flow-close/references/branch-cleanup.md`, `content/skills/unattended-delegation/SKILL.md`, `_support/docs/architecture/workspace-and-artifacts.md`

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** el ticket como copia que reemplaza una captura o un video (con la lista de adjuntos leída), el fin de «support decision» y de las acciones de soporte que pasaban por `flow-close`. No cambia lo que `project-layout` hace cuando el usuario lo pide.

**Verification:** `go test ./tests/content/ -count=1` pasa. `rg -n 'support decision|support-material decision|only a capture whose URL|never a disposable input|replaced only when|counts as replaced only|only when its upload' content/ _support/docs/architecture/` no da resultados (en la base da 9: `workspace-and-artifacts.md:55`, `layout.md:61`, `support-material.md:10`, que T1 reescribe, `pr-captures.md:35`, `unattended-delegation/SKILL.md:18`, `branch-cleanup.md:9` y `:15`, `viewer-close.md:7`, `workspace-and-artifacts.md:104`).

### T2 — El helper reporta clones no registrados

- [x] `cleanup_branches.py` devuelve `stray_clones` con ruta, rama, `dirty` y `unpushed`, y no los toca con `--apply`.

**Closes:** AC7

**Depends on:** T5

**Locations:** `content/skills/flow-close/scripts/cleanup_branches.py`, `tests/skills/cleanup_branches_test.py`

**Execution:** hilo principal

**Test approach:** tdd: un repositorio temporal con su origen clonado en `.worktrees/x` dentro del repositorio y en `.worktrees/y` de su carpeta padre (este último con un commit sin publicar y `origin` en otra forma de URL del mismo repositorio), más un clon de un origen ajeno; la prueba ve fallar la clave ausente, luego exige los dos primeros con su `unpushed`, el ajeno fuera, y los tres en disco tras `--apply`.

**Changes:** una función que recorre las carpetas de worktrees del checkout principal y su padre. No borra ni mueve nada; `apply_cleanup` no cambia.

**Verification:** `python3 -m unittest tests.skills.cleanup_branches_test -q` pasa. `go test ./tests/content/ -count=1` pasa. `python3 -I content/skills/flow-close/scripts/cleanup_branches.py --repo /Users/me/Development/projects/org-a/project-b/project-b-frontend-monorepo --base development` (sin `--apply`, solo lectura) lista `.worktrees/prl-905-906-frontend` en `stray_clones` (AC7), y sus campos coinciden con los que nombra `branch-cleanup.md` «Report»; si el clon ya no existe, AC7 se observa en la prueba.

### T3 — La auditoría falla lo que hoy marca como revisión

- [x] `close_before_branch_cleanup` y `merged_branch_deleted` cambian como fija `design.md`, con fixtures y pruebas.

**Closes:** AC5, AC6

**Depends on:** T5 (lee la forma «kept `<rama>`» de `branch-cleanup.md` y el subconjunto de frases de `design.md`)

**Locations:** `tests/pilot/session_audit.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/close_before_branch_cleanup/`, `tests/fixtures/sessions/existing_criteria/`, `_support/docs/harness-engineering/session-review-rubric.md` (§0), `_support/docs/harness-engineering/rule-enforcers.md` (fila 50 y el subconjunto de frases)

**Execution:** hijo `hive-write-tests`

**Test approach:** tdd. Los cuatro casos que hoy esperan `review` (`session_audit_test.go:533`, `:537`, `:540`, `:541`) pasan a `fail`, con sus fixtures renombradas a `-fail`. Casos nuevos, cada uno visto fallar antes del cambio: un reporte con «kept `<rama>`» de una rama borrada (review); un reporte que dice «sigue en `master`» sin esa forma (fail); una promoción con `gh pr create --head development` seguida de «es todo» (review); un merge sin cabeza visible seguido de «es todo» (fail); «kept `<rama>`» en el último mensaje antes de «es todo», con la rama del merge visible por `--head` (review); un borrado por `"$SCRIPT" --apply` antes del reporte (pass); «es todo por el momento, gracias!» tras un merge sin borrado (fail, en Claude, Codex y Grok); «¿eso es todo lo que falta?» tras un merge (sin cambio de estado).

**Changes:** `isCleanupApply` reconoce el helper invocado por variable; `closeBeforeCleanupStatus`, la excepción y el comentario D11; la detección del mensaje de fin de sesión con `s.Prompts` y `s.Turn`. No cambia el criterio de piloto `merged_branch_deleted` de `regression.go`.

**Verification:** `go test ./tests/pilot/ -run 'TestAudit' -count=1` pasa. `go run ./tests/pilot --audit-session ~/.codex/sessions/2026/11/20/rollout-2026-11-20T04-09-42-01a1f86e-f093-7ae2-b566-2df3cbaf2d7b.jsonl --host codex` da `merged_branch_deleted` = `fail` (AC5); `go run ./tests/pilot --audit-session <Grok 01a1fb69 updates.jsonl> --host grok` da `close_before_branch_cleanup` = `fail` en `:5816` y no en `:6000` (AC6).

### T4 — Spec, sitio, README y registros

- [x] El delta de `workflow-lifecycle`, la página de cierre, README, `llms.txt` y `CHANGELOG.md` describen el cierre nuevo.

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** T1, T5

**Locations:** `_support/openspec/changes/close-cleans-at-completion/specs/workflow-lifecycle/spec.md`, `site/src/content/docs/flows/close.md`, `README.md`, `llms.txt`, `CHANGELOG.md`

**Execution:** hilo principal

**Changes:** la página describe el comportamiento y enlaza la skill, sin copiar sus reglas; README (`:459-463`) y `llms.txt` (`:41`) dejan de decir que el material y la limpieza esperan la respuesta; una entrada en `[Unreleased]`.

**Verification:** `pnpm --dir site build` sale con 0. `python3 /Users/me/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/close-cleans-at-completion` sale con 0.

## Verificación del cambio completo

`go test ./...` (241 s en la base) y `python3 -m unittest discover -s tests/skills -p '*_test.py'` (91 s) pasan, a la vez. Revisión de código con `/code-review`. Tras integrar, `hive update` y una lectura de las hojas instaladas para AC1 a AC4.

## Progreso

2026-11-21 — T1 · 31ef6f14 · unverified: AC1–AC4 se observan en las hojas instaladas tras hive update
2026-11-21 — T5 · 2852cc57 · unverified: none
2026-11-21 — T2 · 50a4ccb6 · unverified: none
2026-11-21 — T4 · bf451910 · unverified: AC1–AC4 en las hojas instaladas tras hive update
2026-11-21 — T3 · 5f59dea2 · unverified: none
## Checkpoint

- Updated: 2026-11-21
- Stage: verify · 6 integrate
- Last commit: 53f4ba24
- Open findings: none
- Next: push, PR, merge, hive update, flow-close
- Rules: flow-build/stage.md
- Principles: none

### Evidence

- before captures: none (no screen)
- go test ./... · exit 0 · 5f59dea2
- python3 -m unittest discover -s tests/skills -p '*_test.py' · OK 321 tests · 5f59dea2
- pnpm --dir site build · exit 0 · bf451910
- CI: master-source-branch.yml only gates PRs into master; not applicable
- go test ./... · exit 0 · 53f4ba24
- python3 -m unittest discover · OK 321 tests · 4d739708 (python unchanged since)
- audit Codex 01a1f86e-f093 merged_branch_deleted=fail; Grok 01a1fb69 close_before_branch_cleanup=fail · 53f4ba24
- helper on project-b-frontend lists .worktrees/prl-905-906-frontend · 50a4ccb6
- /code-review 2 rounds: 6+4 important fixed (de6bda8e, 4d739708, 53f4ba24 Unreviewed); nits to close findings
