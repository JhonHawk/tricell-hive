# Tasks

Base `development` en `48824e10`; worktree `.claude/worktrees/audit-issues`, rama `feat/audit-issues`. Todas las tareas corren en el hilo principal. T1 escribe 4 archivos nuevos de una sola skill y queda aquí, contra el punto 2 de `delegation.md` «When a child does the work», por dos razones: su interfaz es el texto mismo que `design.md` fija junto con el razonamiento de esta sesión (las fuentes, los hallazgos H1–H14 y N1–N7), y un hijo tendría que reconstruir ese contexto desde el brief. T3 solo agrega o corrige líneas de catálogo y documentación.

### T1 — La skill `audit-issues` con sus dos referencias

- [x] `content/skills/audit-issues/` existe con `SKILL.md`, `assets/hive-managed.json`, `references/checks.md` y `references/tracker-writes.md`, y cumple `design.md` y las decisiones de `proposal.md`.

**Closes:** AC4, AC5, AC6

**Depends on:** none

**Locations:** `content/skills/audit-issues/SKILL.md`, `content/skills/audit-issues/assets/hive-managed.json`, `content/skills/audit-issues/references/checks.md`, `content/skills/audit-issues/references/tracker-writes.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Texto nuevo y original con la estructura de `design.md`. Enlaza, sin repetirlos, lo que ya tiene dueño:
- de `create-issue`: los pasos 2 y 4 a 6 para un padre nuevo, `breakdown.md` «Writes», `trackers.md` «Parent and children», `ticket-types.md` (lista de listo) y `existing-ticket.md` paso 3 (autoría);
- `hive-core/tracker/trackers.md` «Move» y «Comment and attach»;
- `ticket-status` «Report»;
- las hojas `questions` y `delegation`.

Las excepciones (autoría de D6, grupo por causa raíz) se escriben en `checks.md`, donde se aplican, nombrando lo que cambian. La descripción evita el disparador «backlog» de `hive-mode`. No cambia ninguna otra skill.

**Verification:** `go test ./tests/content/...` pasa (tamaño, frontmatter, enlaces, marcador `hive-managed`). La línea de cada punto de AC4, AC5 y AC6 queda citada en «Evidence» del checkpoint.

### T2 — Enrutamiento hacia `audit-issues`

- [x] `hive-mode/SKILL.md` envía a `audit-issues` un pedido de auditar, agrupar, deduplicar o limpiar los tickets abiertos, y la descripción de `create-issue` lo excluye.

**Closes:** AC3

**Depends on:** T1

**Locations:** `content/skills/hive-mode/SKILL.md`, `content/skills/create-issue/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- En `hive-mode`, una oración en el párrafo que precede al paso 1. Fija el orden: auditar, agrupar o cerrar tickets va a `audit-issues`, que reemplaza los pasos 0 a 5; un informe de estado o de backlog sigue en `ticket-status`.
- En la descripción de `create-issue`, una cláusula «not for».
- Antes de editar, buscar en `content/` (incluido `guidance/global.md`), `site/`, `_support/docs/architecture/` y `tests/pilot/` cada lugar que enrute pedidos de tickets o de backlog, y listarlos en el PR, como pide `AGENTS.md`. `global.md` no cambia.

**Verification:** `go test ./tests/content/...` pasa (`skill_links_test`, `single_home_test`, `skill_frontmatter_test`). `rg -n 'audit-issues' content/skills/hive-mode/SKILL.md content/skills/create-issue/SKILL.md` muestra una línea en cada archivo; en la base no da coincidencias.

### T3 — Catálogo, sitio y documentación

- [x] El catálogo, las páginas del sitio que describen esas rutas, `instruction-resources.md` y `CHANGELOG.md` describen `audit-issues`.

**Closes:** AC1

**Depends on:** T1, T2

**Locations:** `llms.txt`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/flows/hive-mode.md`, `site/src/content/docs/flows/create-issue.md`, `site/src/content/docs/flows/status.md`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- Una entrada de catálogo junto a `audit-comments` y `audit-tests`.
- En `flows/hive-mode.md` y `flows/status.md`, la ruta nueva frente a la de estado o backlog.
- En `flows/create-issue.md`, «What it does not do» y «Related flows».
- En `instruction-resources.md`, una entrada nueva con su dueño, sus referencias y lo que enlaza, más las líneas 23 (lo que enlaza el enrutador) y 24 (quién enlaza `create-issue`).
- En `CHANGELOG.md`, una entrada en «Added».
- Sin crédito a repositorios de referencia.

**Verification:** `npm --prefix site run build` termina con `Complete!`; `go test ./...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y `npm --prefix site run build` | Covered by the implementation |
| Code review | Final candidate, before the push | `Native: Claude Code /code-review` | Review A |

AC1 se observa tras el merge, al refrescar la instalación local (marcador de mantenedor presente): `hive update --source <checkout>` y `ls -l ~/.claude/skills/audit-issues`. AC3 a AC6 se observan en los archivos instalados en ese mismo refresco.

## Checkpoint

- Updated: 2026-11-22
- Stage: verify · 2 review running
- Last commit: 0dbc206d
- Open findings: none
- Next: apply /code-review findings, then push, PR, merge, refresh install, observe AC1/AC3-AC6
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/audit-issues · feat/audit-issues
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- before captures: none (no UI)
- go test ./... · pass 3m02s · 0dbc206d
- npm --prefix site run build · Complete!, 20 pages · 0dbc206d
- go test ./tests/content/... · pass · f96590e4
- CI master-source-branch.yml: only guards PRs into master; not applicable to development
- AC5 candidate: SKILL.md:12 count+pass before code; :28-44 set template with verdict, evidence, level, auto-close; :46 one approval, correct/take out; :48 unattended no writes; :52 one at a time, stop on failure; checks.md:48 Authorship comment only
- AC6 candidate: checks.md:13-25 checks 1-8, :61-68 checks 9-14; :19 firm duplicate evidence; :15-16 not confirmed; :41 Exclusions; :63 root-cause exception to breakdown «The set»
- AC4 candidate: tracker-writes.md:27-29 close table (gh --duplicate-of, --reason not planned/completed, duplicateOf, canceled, Jira link type); :41 related per tracker; :54-55 child auto-close unknown; :61 read back; :31,:48 link Move and Comment and attach

## Review status and progress

2026-11-22 — T1 · 30e2aa79 · unverified: AC4, AC5 y AC6 se observan en la instalación tras el merge
2026-11-22 — T2 · f96590e4 · unverified: AC3 se observa en la instalación tras el merge
2026-11-22 — T3 · 0dbc206d · unverified: AC1 se observa en la instalación tras el merge
