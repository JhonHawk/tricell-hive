# Tasks

Base `development` en `48824e10`; worktree `.claude/worktrees/audit-issues`, rama `feat/audit-issues`. Todas las tareas corren en el hilo principal. T1 escribe 4 archivos nuevos de una sola skill y queda aquí, contra el punto 2 de `delegation.md` «When a child does the work», por dos razones: su interfaz es el texto mismo que `design.md` fija junto con el razonamiento de esta sesión (las fuentes, los hallazgos H1–H14 y N1–N7), y un hijo tendría que reconstruir ese contexto desde el brief. T3 solo agrega o corrige líneas de catálogo y documentación.

### T1 — La skill `audit-issues` con sus dos referencias

- [ ] `content/skills/audit-issues/` existe con `SKILL.md`, `assets/hive-managed.json`, `references/checks.md` y `references/tracker-writes.md`, y cumple `design.md` y las decisiones de `proposal.md`.

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

- [ ] `hive-mode/SKILL.md` envía a `audit-issues` un pedido de auditar, agrupar, deduplicar o limpiar los tickets abiertos, y la descripción de `create-issue` lo excluye.

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

- [ ] El catálogo, las páginas del sitio que describen esas rutas, `instruction-resources.md` y `CHANGELOG.md` describen `audit-issues`.

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
- Stage: build · T1
- Last commit: 48824e10
- Open findings: none
- Next: T1: write content/skills/audit-issues
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/audit-issues · feat/audit-issues
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no UI)

## Review status and progress
