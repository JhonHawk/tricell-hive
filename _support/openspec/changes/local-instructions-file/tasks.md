# Tasks

Base `e0bb3a95` en `development`; rama `feat/local-instructions-file`. Todas las tareas corren en el hilo principal: son unos 22 archivos de texto con sus sitios ya localizados durante el plan, y ninguno pasa de 5 lecturas grandes.

### T1 — Sección «Local instructions» en la guía global

- [x] `content/guidance/global.md` tiene la sección «Local instructions» después de «Test accounts», con el texto de `design.md`.

**Closes:** AC6, AC7, AC8, AC9, AC10, AC11, AC12, AC13, AC14

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** añade una sección de cinco puntos. No toca «Secrets», «Test accounts», «Persistent memory» ni «Support folder»: la excepción de la raíz y la autorización de escritura quedan dentro de la sección nueva. El commit nombra las sesiones que motivan el cambio (Claude Code `8d09a5eb` y `94440d0a` de carvento, Codex `01a1a844` de cam, Claude Code `b7986711` y `1776e095` de carvento), como pide «Measurement».

**Verification:** cada frase de AC6-AC14 devuelve 1 coincidencia con `rg -c -F "<frase>" content/guidance/global.md`; `go test ./tests/content/...` pasa y registra el tamaño nuevo de `global.md` (base 37 751 bytes) para el PR.

### T2 — Punteros en skills, referencias, auditoría y rol

- [x] Los siete archivos de AC15 nombran «Local instructions» o `HIVE.local.md` como fija `design.md` «Punteros y documentos».

**Closes:** AC15

**Depends on:** T1

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-build/references/delegation.md`, `content/skills/workspace-conventions/references/artifact-layout.md`, `content/skills/harness-audit/references/project-layout.md`, `content/skills/harness-audit/references/instruction-files.md`, `content/agents/review/hive-review-harness.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una frase por archivo que nombra la regla y cuándo aplica; ninguna repite sus pasos. HA-IF-15 y el rol de auditoría distinguen «no lo propone como hallazgo» de «la regla global lo escribe». HA-PL-05 y `artifact-layout.md` exceptúan `HIVE.local.md` en la raíz. El puntero de delegación pide además el informe del hijo.

**Verification:** `rg -l "Local instructions|HIVE.local.md" content` lista los siete archivos y `global.md`; `go test ./tests/content/...` pasa (tamaño de skills y referencias, enlaces).

### T3 — Sitio y arquitectura

- [x] `site/src/content/docs/getting-started.md` («Give the session context») explica `HIVE.local.md` para personas: qué guarda, que no se commitea, que lo escribe el agente y que se puede editar a mano; `flows/build.md` y `flows/verify.md` («What it does») dicen que usan y actualizan lo verificado del entorno local; `_support/docs/architecture/workspace-and-artifacts.md` nombra la excepción de la raíz.

**Closes:** AC16

**Depends on:** T1

**Locations:** `site/src/content/docs/getting-started.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `_support/docs/architecture/workspace-and-artifacts.md`

**Execution:** main thread

**Test approach:** check

**Changes:** describen el comportamiento y enlazan a la guía; no copian la regla. README y `llms.txt` no cambian: no mencionan nada de lo que cambia.

**Verification:** `rg -l "HIVE.local.md" site/src _support/docs/architecture` lista los cuatro archivos; `npm --prefix site run build` pasa (usa `site/node_modules`, ya instalado).

### T4 — Script `extract_candidates.py` y su test

- [x] `content/skills/local-instructions/scripts/extract_candidates.py` y `tests/skills/local_instructions_extract_test.py` implementan el paso 2 de `design.md` «Skill `local-instructions`».

**Closes:** AC25

**Depends on:** none

**Locations:** `content/skills/local-instructions/scripts/extract_candidates.py`, `tests/skills/local_instructions_extract_test.py`

**Execution:** main thread

**Test approach:** tdd (maneja secretos): el test, con transcripciones sintéticas de Claude Code y Codex, se ve fallar antes de escribir el script.

**Changes:** solo librería estándar; lee, nunca escribe; sin red.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa; el test comprueba que ninguno de los cuatro secretos sembrados aparece en la salida, que se cuentan los comandos con 0 y los que faltaron, y que quedan fuera el proyecto vecino con el mismo prefijo, los subagentes y las sesiones viejas.

### T5 — Skill `local-instructions` y su referencia

- [x] `content/skills/local-instructions/SKILL.md` y `references/transcript-sources.md` implementan el resto de `design.md` «Skill `local-instructions`».

**Closes:** AC17, AC24, AC26, AC27, AC28, AC29

**Depends on:** T1, T4

**Locations:** `content/skills/local-instructions/SKILL.md`, `content/skills/local-instructions/references/transcript-sources.md`

**Execution:** main thread

**Test approach:** check

**Changes:** la skill apunta a «Local instructions» sin repetirla y nombra su única excepción (`unverified` con fecha de origen).

**Verification:** `go test ./tests/content/...` pasa (frontmatter, descripción ≤1 024 caracteres, tamaño ≤16 000 bytes, enlaces); `rg -c -F` encuentra en `SKILL.md`: «*Create* when» y «*Curate* when» (AC17), «scripts/extract_candidates.py --root» y «never their contents» (AC24), «`D - Choose by number`» y «not in persistent memory: this departs» (AC26), «never through `npx`» y «takes the date it was last seen and the mark `unverified`» (AC27), «Write as «Local instructions» sets.» y «**every clone** fact as a candidate» (AC28); en la referencia, «-Users-me-work-shop--claude-worktrees-w1», «`payload.cwd`» y «No: not covered» (AC29). El test de T4 cubre worktrees y la raíz del workspace (AC24).

### T6 — Regla de `unverified` y enlaces a la skill

- [x] «Local instructions» dice qué significa `unverified` para un lector y recomienda la skill en la línea del tope; `site/src/content/docs/skills.md`, `index.mdx` («Supporting skills»), `getting-started.md`, `llms.txt` y `CHANGELOG.md` («Unreleased», sección «Added» nueva) la nombran.

**Closes:** AC30, AC31

**Depends on:** T5

**Locations:** `content/guidance/global.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/getting-started.md`, `llms.txt`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una cláusula en la guía global y un enlace o una línea en cada archivo.

**Verification:** `rg -l "local-instructions" content/guidance site/src llms.txt CHANGELOG.md` lista los seis archivos; `rg -c -F` encuentra en `global.md` «which only the [local-instructions](skill:local-instructions/SKILL.md) skill writes» (AC31) y «recommend [local-instructions](skill:local-instructions/SKILL.md) to curate it» (AC30); `go test ./tests/content/...` pasa y registra el tamaño nuevo de `global.md`; `npm --prefix site run build` pasa; se repite la búsqueda de todo el comportamiento.

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
2026-11-07 — T1 hecho en `0ed42b6a`: las nueve frases de AC6-AC14 aparecen una vez; `global.md` pasa de 37 751 a 39 739 bytes (+1 988, más que los ~0,8 KB estimados).
2026-11-07 — T2 hecho en `8c896e02`: los siete archivos y `global.md` nombran la regla; `go test ./tests/content/...` pasa.
2026-11-07 — T3 hecho en `31d8a2ee`: los cuatro archivos describen `HIVE.local.md`; `npm run build` del sitio construye 17 páginas.

2026-11-07 — Verify sobre `240fc93f`: el único workflow (`master-source-branch.yml`) solo valida la rama de origen de PRs a `master`, CI-only. `go vet ./...` (0), `go test ./...` (0), unittest de skills (0, 58 tests). Las nueve frases de AC6-AC14 aparecen una vez; AC15 y AC16 por `rg -l`. La revisión de todo el comportamiento encontró `harness-audit/SKILL.md:57` y `apply.md:5` nombrando solo `CLAUDE.local.md`: corregidos en `240fc93f`.

2026-11-07 — `/code-review` sobre `240fc93f`: 10 hallazgos confirmados leyendo sus fuentes, corregidos en `87fef783` (tests de detectores ampliados). Plan revisado: T4 y T5 añadidos (D8 B).

2026-11-07 — Segunda revisión del plan aplicada y D12 A: T4 script, T5 skill, T6 enlaces. Hay un borrador sin commitear de la skill en `content/skills/local-instructions/` escrito antes de esa revisión; T5 lo reescribe.

2026-11-07 — T4 hecho en `1b2c5683` y `bc19cd24` (tdd: el test falló antes del script y otra vez antes del enmascarado por lista de permitidos); prueba real sobre cam, carvento y fleetsystem sin fugas en un escaneo de patrones.
2026-11-07 — T5 hecho en `cf28726f`: frases de AC17, AC24 y AC26-AC29 presentes; tests de contenido pasan.
2026-11-07 — T6 hecho en el commit siguiente: frases de AC30 y AC31 en `global.md` (40 228 bytes, +2 477 sobre la base); el sitio construye.

Checkpoint: stage verify, segunda revisión de código sobre `240fc93f..HEAD` y segunda parada de validación. Último commit `87fef783`. Pendiente: revisión del plan revisado, T4, T5, re-verify y segunda parada de validación.
