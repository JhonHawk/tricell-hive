# Tasks

Base `e0bb3a95` en `development`; rama `feat/local-instructions-file`. Todas las tareas corren en el hilo principal: son unos 13 archivos de texto con sus sitios ya localizados durante el plan, y ninguno pasa de 5 lecturas grandes.

### T1 — Sección «Local instructions» en la guía global

- [ ] `content/guidance/global.md` tiene la sección «Local instructions» después de «Test accounts», con el texto de `design.md`.

**Closes:** AC6, AC7, AC8, AC9, AC10, AC11, AC12, AC13, AC14

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** añade una sección de cinco puntos. No toca «Secrets», «Test accounts», «Persistent memory» ni «Support folder»: la excepción de la raíz y la autorización de escritura quedan dentro de la sección nueva. El commit nombra las sesiones que motivan el cambio (Claude Code `8d09a5eb` y `94440d0a` de carvento, Codex `01a1a844` de cam, Claude Code `b7986711` y `1776e095` de carvento), como pide «Measurement».

**Verification:** cada frase de AC6-AC14 devuelve 1 coincidencia con `rg -c -F "<frase>" content/guidance/global.md`; `go test ./tests/content/...` pasa y registra el tamaño nuevo de `global.md` (base 37 751 bytes) para el PR.

### T2 — Punteros en skills, referencias, auditoría y rol

- [ ] Los siete archivos de AC15 nombran «Local instructions» o `HIVE.local.md` como fija `design.md` «Punteros y documentos».

**Closes:** AC15

**Depends on:** T1

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-build/references/delegation.md`, `content/skills/workspace-conventions/references/artifact-layout.md`, `content/skills/harness-audit/references/project-layout.md`, `content/skills/harness-audit/references/instruction-files.md`, `content/agents/review/hive-review-harness.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una frase por archivo que nombra la regla y cuándo aplica; ninguna repite sus pasos. HA-IF-15 y el rol de auditoría distinguen «no lo propone como hallazgo» de «la regla global lo escribe». HA-PL-05 y `artifact-layout.md` exceptúan `HIVE.local.md` en la raíz. El puntero de delegación pide además el informe del hijo.

**Verification:** `rg -l "Local instructions|HIVE.local.md" content` lista los siete archivos y `global.md`; `go test ./tests/content/...` pasa (tamaño de skills y referencias, enlaces).

### T3 — Sitio y arquitectura

- [ ] `site/src/content/docs/getting-started.md` («Give the session context») explica `HIVE.local.md` para personas: qué guarda, que no se commitea, que lo escribe el agente y que se puede editar a mano; `flows/build.md` y `flows/verify.md` («What it does») dicen que usan y actualizan lo verificado del entorno local; `_support/docs/architecture/workspace-and-artifacts.md` nombra la excepción de la raíz.

**Closes:** AC16

**Depends on:** T1

**Locations:** `site/src/content/docs/getting-started.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `_support/docs/architecture/workspace-and-artifacts.md`

**Execution:** main thread

**Test approach:** check

**Changes:** describen el comportamiento y enlazan a la guía; no copian la regla. README y `llms.txt` no cambian: no mencionan nada de lo que cambia.

**Verification:** `rg -l "HIVE.local.md" site/src _support/docs/architecture` lista los cuatro archivos; `npm --prefix site run build` pasa (usa `site/node_modules`, ya instalado).

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Whole-behavior review | Final candidate, locally | Search of global guidance, skills, references, roles, `site/`, `_support/docs/architecture/` and `tests/pilot/` for conflicting or duplicated instructions, listed in the PR | Covered by the implementation |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | D6 A |

Revisión humana (D5 A): leer la sección nueva de `global.md` y el diff de los punteros antes del push.

## Review status and progress

2026-11-07 — Plan escrito y revisado por `hive-review-plan`; hallazgos aplicados (ver `proposal.md`).
