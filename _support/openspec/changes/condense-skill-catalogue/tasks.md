# Tasks

Base `development` @ bedc9b64, rama `refactor/condense-skill-catalogue`. Todas las tareas corren en el hilo principal: son movimientos con `git mv` y ediciones de enlaces que dependen unas de otras en los mismos archivos (`global.md`, `flow-close/SKILL.md`), y un hijo no podría verificarlas por separado.

### T1 — flow-walkthrough pasa a referencia de flow-verify

- [x] La publicación de capturas y el walkthrough se leen desde `flow-verify/references/`, y nada enlaza ya `flow-walkthrough`

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `content/skills/flow-walkthrough/` → `content/skills/flow-verify/references/{walkthrough,viewer-publication,host-session-identity}.md`; `flow-verify/SKILL.md`; enlaces en `flow-build/SKILL.md`, `flow-close/SKILL.md`, `flow-report/SKILL.md`, `unattended-delegation/SKILL.md`, `flow-verify/references/ui-verification.md`, `content/agents/review/hive-review-ux.md`, `workspace-conventions/references/artifact-layout.md`; `AGENTS.md:76`

**Execution:** main thread

**Test approach:** check

**Changes:** mueve el procedimiento sin cambiar sus pasos; conserva los encabezados `#open-the-viewer` y `#close-and-input-hygiene`. La descripción de flow-verify añade el walkthrough o demo pedido por el usuario. No cambia ninguna regla de captura.

**Verification:** `go test ./tests/content/... ./tooling/management/...` pasa y `rg -n 'flow-walkthrough' content AGENTS.md` no devuelve nada

### T2 — flow-status y security-advisory pasan a modos de flow-research

- [x] El estado del proyecto y la exposición a un aviso se leen desde `flow-research/references/`

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `content/skills/flow-status/SKILL.md` → `flow-research/references/status.md`; `content/skills/security-advisory/SKILL.md` → `flow-research/references/security-advisory.md`; `flow-research/SKILL.md` (descripción y condiciones de lectura); `content/guidance/global.md:59,71,92`; `flow-close/SKILL.md:85`; `flow-build/references/delegation.md:21`

**Execution:** main thread

**Test approach:** check

**Changes:** los pasos de ambos procedimientos no cambian; la descripción de research deja de excluir el estado y nombra los dos disparadores. Cada línea de lectura dice que su referencia reemplaza los pasos 0 a 5 de research y usa su propio cierre. En `global.md`, el estado deja de ser un flujo aparte en las líneas 59 y 92 y la 71 apunta al modo. No cambia ningún otro paso de research.

**Verification:** `go test ./tests/content/...` pasa, la descripción de flow-research cabe en 1.024 caracteres, y `rg -n 'flow-status|security-advisory' content` solo encuentra las referencias nuevas por su ruta

### T3 — local-instructions y engram-init-workspace pasan a modos de harness-audit

- [x] La curación de `HIVE.local.md` y la identidad de Engram se leen desde `harness-audit/references/`, la segunda solo a petición explícita

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/local-instructions/` → `harness-audit/references/{local-instructions,transcript-sources}.md`, `harness-audit/scripts/extract_candidates.py`; `content/skills/engram-init-workspace/` → `harness-audit/references/engram-workspace.md`, `harness-audit/scripts/init_workspace.py`; `harness-audit/SKILL.md`; `harness-audit/references/memory.md:32`; `content/guidance/global.md:31,57`; `tests/skills/local_instructions_extract_test.py`, `tests/skills/engram_workspace_test.py`

**Execution:** main thread

**Test approach:** check

**Changes:** los pasos no cambian; la regla global de identidad de Engram nombra el modo y mantiene su exclusividad. `harness-audit/SKILL.md` crece solo en las líneas de lectura, porque OpenCode lo inyecta entero en `hive-review-harness`.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` y `go test ./tests/content/...` pasan; la descripción de harness-audit cabe en 1.024 caracteres

### T4 — workspace-conventions pasa a referencias de flow-close

- [x] El mantenimiento de `_support`, `artifact-layout.md` y `archive.md` se leen desde `flow-close/references/`, y `flow-close/SKILL.md` sigue ≤ 16.000 bytes sin perder ninguna regla

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** T1, T2 (ambas editan `flow-close/SKILL.md`)

**Locations:** `content/skills/workspace-conventions/` → `flow-close/references/{support-maintenance,artifact-layout,archive}.md` (con `<workspace-conventions-dir>` de `archive.md:13` cambiado al directorio nuevo), `flow-close/scripts/archive_sessions.py`; `flow-close/SKILL.md`; enlaces en `content/guidance/global.md:76`, `content/guidance/README.md`, `harness-audit/SKILL.md`, `harness-audit/references/project-layout.md`, `flow-report/SKILL.md`, `flow-figma-design/SKILL.md`, `unattended-delegation/SKILL.md`, `flow-verify/references/ui-verification.md`, `flow-verify/references/walkthrough.md`, `content/agents/docs/hive-write-docs.md`; `AGENTS.md:53`; `tests/skills/workspace_archive_test.py`; `tests/fixtures/workspace-conventions/README.md`

**Execution:** main thread

**Test approach:** check

**Changes:** los pasos no cambian. `flow-close/SKILL.md` gana en su descripción "or to organize, curate, promote or archive support material" y una línea de lectura hacia `references/support-maintenance.md`; para caber hay que recortar unos 120 B de redacción, según el presupuesto de `design.md`, comparado línea a línea con el original en el commit y sin quitar ninguna regla. Si el recorte necesario pasa de 200 B, se para y se vuelve a preguntar D9. No se renombra la carpeta de fixtures `tests/fixtures/workspace-conventions/`, que es entrada histórica de los pilotos.

**Verification:** `go test ./tests/content/... ./tooling/management/...` y los tests de `tests/skills/` pasan; `wc -c content/skills/flow-close/SKILL.md` ≤ 16000

### T5 — starlight-docs-site pasa a modo de flow-report

- [ ] Un sitio de documentación Starlight se arma desde `flow-report/references/docs-site.md` con los mismos assets

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/starlight-docs-site/` → `flow-report/references/docs-site.md`, `flow-report/assets/starlight/`, `flow-report/scripts/contrast-check.py`; `flow-report/SKILL.md`; `flow-report/assets/skeleton-paper.html:14`; `tests/skills/starlight_assets_test.py`

**Execution:** main thread

**Test approach:** check

**Changes:** los pasos y los assets no cambian, salvo sus rutas relativas.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa, y, con la tarea ya en un commit (el dry-run despliega el HEAD confirmado), `hive update --source . --dry-run` valida los assets

### T6 — Lista de skills de los pilotos

- [ ] La lista de skills del pilot `deployed-smoke` nombra solo skills que existen

**Closes:** AC4

**Depends on:** T1, T3, T4, T5

**Locations:** `tests/pilot/flows.go:550`, `tests/pilot/flows_test.go:756`, `tests/fixtures/flows/cases.json:33`, `tests/fixtures/flows/README.md`

**Execution:** main thread

**Test approach:** check

**Changes:** las 4 skills retiradas de esa lista (`flow-walkthrough`, `engram-init-workspace`, `starlight-docs-site`, `workspace-conventions`) se reemplazan por sus destinos existentes (`flow-verify`, `harness-audit`, `flow-report`, `flow-close`), deduplicando, porque una referencia no se anuncia como skill. La suite histórica `workspace-conventions` de `tests/pilot/main.go` y `assess.go` se deja como está, marcada histórica en su README de fixtures: los pilotos están pausados y es entrada de mediciones pasadas. `tooling/`, `integrations/` y `tooling/legacy/catalog.json` no cambian.

**Verification:** `go test ./...` y `go vet ./...` pasan

### T7 — Documentación, specs y CHANGELOG

- [ ] El sitio, el README, `llms.txt`, la arquitectura y el `CHANGELOG` describen el catálogo de 13 y a dónde fue cada procedimiento

**Closes:** AC3, AC6

**Depends on:** T1, T2, T3, T4, T5

**Locations:** `site/src/content/docs/{skills.md,index.mdx,viewer.md,flows/status.md,flows/security.md,flows/research.md}` y otras páginas que `rg` encuentre; `README.md`; `llms.txt`; `_support/docs/architecture/{instruction-resources,repository-and-distribution,agent-delivery,workspace-and-artifacts}.md` según `rg`; `_support/docs/harness-engineering/`; `CHANGELOG.md` (Unreleased: «Removed» y «Changed»); los deltas de `specs/` de este cambio

**Execution:** main thread

**Test approach:** check

**Changes:** las páginas describen comportamiento; cada página de flujo sigue enlazando el `SKILL.md` de su skill (Status enlaza `flow-research/SKILL.md`), y las listas de skills enlazan el `SKILL.md` o la referencia nueva.

**Verification:** `rg -n 'flow-walkthrough|flow-status|security-advisory|local-instructions|engram-init-workspace|workspace-conventions|starlight-docs-site' content AGENTS.md site README.md llms.txt _support/docs _support/openspec/changes/condense-skill-catalogue/specs` solo devuelve rutas nuevas y las excepciones que AC3 permite; `go test ./...` pasa

### T8 — Observar el retiro en la instalación local

- [ ] Tras el merge, la instalación local de esta máquina ya no tiene las 7 skills y sí las referencias nuevas

**Closes:** AC5

**Depends on:** T1, T2, T3, T4, T5, T6, T7 y el merge

**Locations:** binario `hive` y contenido desplegado de esta máquina

**Execution:** main thread, en el refresco del mantenedor que `AGENTS.md` define

**Changes:** despliega el contenido del commit del merge con `hive update --source <checkout> --out <plan>` y `hive apply --plan <plan>`, sin recompilar el binario porque `tooling/` e `integrations/` no cambian; no toca instalaciones a nivel de proyecto.

**Verification:** `ls ~/.agents/skills ~/.claude/skills ~/.gemini/config/skills ~/.gemini/antigravity/skills` no lista ninguna de las 7 skills, y `ls ~/.agents/skills/flow-close/references/artifact-layout.md ~/.agents/skills/flow-verify/references/walkthrough.md` existe

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...`, `go vet ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Release validation | Final candidate, committed, locally | `hive update --source <worktree> --dry-run`, which also validates the relative links of the moved files | Covered by the implementation |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | The review answer |
| Retirement in the local install (AC5) | After the merge, at the maintainer refresh | `ls ~/.agents/skills ~/.claude/skills` | Maintainer refresh rule of `AGENTS.md` |

Validación humana: leer el diff de `flow-close/SKILL.md` (el recorte), las descripciones de los cuatro destinos y la sección «Removed» del `CHANGELOG`.

## Checkpoint

## Review status and progress

2026-11-14 — T1 · d373e77e · unverified: none
2026-11-14 — T2 · 47e58ddf · unverified: none
2026-11-14 — T3 · a3654f4f · unverified: none
2026-11-14 — T4 · b2484d4e · unverified: none
