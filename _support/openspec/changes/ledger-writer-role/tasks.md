# Tasks

Base `development` en `6d1ab6f0`; worktree `.claude/worktrees/ledger-writer-role`, rama `feat/ledger-writer-role`. Las dos tareas corren en el hilo principal: T1 escribe un archivo y T2 edita seis con cambios de una o dos frases.

### T1 — El rol `hive-write-ledger`

- [x] `content/agents/review/hive-write-ledger.md` existe con la descripción, los pasos de `design.md` y el bloque de límites de `implement`, y se renderiza con escritura en una instalación sintética.

**Closes:** AC1, AC2

**Depends on:** none

**Locations:** `content/agents/review/hive-write-ledger.md`

**Execution:** main thread; un hijo `hive-verify-task` lo verifica (cambio de permisos)

**Test approach:** check

**Changes:** Archivo nuevo con frontmatter (`name`, `description`, `model_profile: "execution"`, `access_profile: "implement"`), los cuatro pasos de `design.md` y el bloque de límites.

**Verification:** `go test ./tests/content/ -run TestAgentBoundaryBlock` y `go test ./integrations/agents/...` pasan. Una instalación del candidato en un home sintético (`hive plan install --hosts claude,codex --scope user --home <tmp>/home --state-dir <tmp>/state --source <worktree> --out <tmp>/plan.json`, luego `hive apply --plan <tmp>/plan.json`) deja `<home>/.claude/agents/hive-write-ledger.md` sin `permissionMode` ni `disallowedTools` y `<home>/.codex/agents/hive-write-ledger.toml`.

### T2 — Las skills, la delegación y la documentación usan el rol

- [ ] `audit-comments` y `audit-tests` dan cada carril a `hive-write-ledger`; `delegation.md` deja correr en paralelo a esos hijos; `agent-delivery.md`, `instruction-resources.md` y `CHANGELOG.md` lo describen.

**Closes:** AC3, AC4

**Depends on:** T1

**Locations:** `content/skills/audit-comments/SKILL.md`, `content/skills/audit-tests/SKILL.md`, `content/skills/hive-core/delegation/delegation.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Los cambios de `design.md` «Skills y hojas» y «Documentación». Antes de editar, buscar en `content/`, `site/`, `_support/docs/architecture/` y `tests/pilot/` cada lugar que trate el paralelismo de hijos o los ledgers, y listarlos en el PR.

**Verification:** `rg -n 'hive-research' content/skills/audit-comments/SKILL.md content/skills/audit-tests/SKILL.md` no da coincidencias; `rg -c 'hive-write-ledger' content/skills/audit-comments/SKILL.md content/skills/audit-tests/SKILL.md` da 1 por archivo; `go test ./...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Task verification | T1, before marking it done | `hive-verify-task` child against AC1 and AC2 | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` | Covered by the implementation |
| Code review | Final candidate, before the push | `Native: Claude Code /code-review` | Review in force |
| Security review | Final candidate, before the push | `hive-review-security` child | Seguridad A |

AC1 se observa en el home sintético de T1 y otra vez en la instalación local tras el merge (marcador de mantenedor presente); AC2 a AC4, en los archivos instalados tras el merge.

## Checkpoint

- Updated: 2026-11-23
- Stage: build · T1
- Last commit: 6d1ab6f0
- Open findings: none
- Next: T1: write content/agents/review/hive-write-ledger.md
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/ledger-writer-role · feat/ledger-writer-role
- Rules: flow-build/stage.md · extra: security-boundaries

### Evidence

- before captures: none (no UI)

## Review status and progress

2026-11-23 — T1 · 8af4be87 · unverified: AC1 en la instalación local tras el merge; AC2 en el rol instalado tras el merge
