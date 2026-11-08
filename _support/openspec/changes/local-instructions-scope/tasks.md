# Tasks

Base `831e6032` en `development`; rama `fix/local-instructions-scope`. Todas las tareas corren en el hilo principal: son 10 archivos de texto ya localizados y dos archivos locales.

### T7 — Criterios y archivo de máquina en «Local instructions»

- [x] `content/guidance/global.md` tiene la sección «Local instructions» con el texto de `design.md`.

**Closes:** AC34, AC36, AC37, AC44, AC45, AC46, AC48, AC51, AC52, AC53, AC54, AC55, AC56

**Depends on:** none

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** reescribe los puntos 1, 2, 4 y 5 de la sección; el 3 (secretos, `unverified`, tope) solo pasa a decir «each file». El commit nombra la sesión Claude Code `de72f693` (2026-11-07) y su fallo, y la aprobación del usuario, como pide «Measurement».

**Verification:** cada frase de AC34, AC36, AC37, AC44-AC46, AC48, AC51-AC56 da al menos 1 con `rg -c -F "<frase>" content/guidance/global.md`; `several or the whole machine` y `Write only in the current project, as` dan 0; `go test ./tests/content/...` pasa y registra el tamaño nuevo (base 40 430 bytes).

### T8 — La skill reparte entre los dos archivos

- [x] `content/skills/local-instructions/SKILL.md` aplica `design.md` «Skill».

**Closes:** AC39, AC40, AC49

**Depends on:** T7

**Locations:** `content/skills/local-instructions/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** la línea de alcance, la descripción y el título, los pasos 3 y 6 y los modos; no repite la regla, la apunta.

**Verification:** `rg -c -F "go to the machine file"`, `rg -c -F "whether a host answered"` y `rg -c -F "Work only in the current project and the machine file"` dan al menos 1 en `SKILL.md`; `rg -c -F "Work only in the current project,"` da 0; `go test ./tests/content/...` pasa.

### T9 — Sitio, arquitectura y changelog

- [x] `site/src/content/docs/getting-started.md`, `_support/docs/architecture/workspace-and-artifacts.md` y `CHANGELOG.md` («Unreleased») describen el archivo de máquina y los criterios.

**Closes:** AC41, AC42

**Depends on:** T7

**Locations:** los tres archivos

**Execution:** main thread

**Test approach:** check

**Changes:** describen el comportamiento y enlazan a la guía; no copian la regla. En `CHANGELOG.md` se corrige la entrada existente de «Added» sobre `HIVE.local.md`, sin añadir otra.

**Verification:** `rg -l "config/hive" site/src _support/docs/architecture CHANGELOG.md` lista los tres; `npm --prefix site run build` pasa.

### T10 — Migración de los archivos locales de esta máquina

- [x] `~/.config/hive/HIVE.local.md` tiene la línea inicial y las entradas de T3 Code, Antigravity y T3 Connect/cloudflared (esta como binario, versión y comando de comprobación, sin «is reachable»); `tricell-hive/HIVE.local.md` conserva su línea inicial sin las entradas de SetStoreUpdateService ni `bb1`.

**Closes:** AC43

**Depends on:** T7

**Locations:** `~/.config/hive/HIVE.local.md`, `HIVE.local.md` (sin seguimiento)

**Execution:** main thread

**Test approach:** check

**Changes:** archivos locales sin seguimiento; no entran en el PR. Quitar las dos entradas lo autoriza la aprobación del plan.

**Verification:** `rg -c "SetStoreUpdateService|bb1" HIVE.local.md` da 0; `rg -c "reachable" ~/.config/hive/HIVE.local.md` da 0; `rg -c "T3 Code|cloudflared|Antigravity" ~/.config/hive/HIVE.local.md` da 3; `git status` no muestra ninguno de los dos.

### T11 — Otros archivos que limitaban la escritura o nombraban un solo archivo

- [x] `content/skills/security-advisory/SKILL.md:8`, `content/skills/harness-audit/SKILL.md:10`, `content/skills/flow-build/references/delegation.md:28` y `llms.txt` aplican `design.md` «Otros punteros».

**Closes:** AC50

**Depends on:** T7

**Locations:** los cuatro archivos, más `flow-research/SKILL.md:8`, `flow-status/SKILL.md:8`, las páginas `skills.md`, `index.mdx`, `flows/build.md`, `flows/verify.md` y el detector `tests/pilot/flows.go` con su test

**Execution:** main thread

**Test approach:** check

**Changes:** una frase por archivo; ninguna repite la regla.

**Verification:** `rg -l "machine file"` lista los cuatro; `go test ./tests/content/...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Whole-behavior review | Final candidate, locally | Search of global guidance, skills, references, roles, `site/`, `_support/docs/architecture/` and `tests/pilot/` for the old single-file wording, listed in the PR | Covered by the implementation |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | D6 A |

Revisión humana (D5 A): leer la sección nueva de `global.md` y el diff.

## Review status and progress

2026-11-07 — Plan escrito y revisado por `hive-review-plan`; hallazgos aplicados (ver `proposal.md`).
2026-11-07 — T7 en `2a36dfd8`: las frases de AC34, AC36, AC37, AC44-AC48, AC51 y AC52 aparecen; `several or the whole machine` y `Write only in the current project, as` dan 0; `global.md` 41 184 bytes (+754).
2026-11-07 — T8 en `ac5a5843`, T11 en `a09214ff`, T9 en `73e1ac05`: frases de AC39, AC40, AC49, AC50, AC41 y AC42 presentes; el sitio construye.
2026-11-07 — T10 local: `~/.config/hive/HIVE.local.md` con 3 entradas (T3 Code, cloudflared sin «reachable», Antigravity); el archivo del repo solo con su línea inicial; `~/.config` no es un repositorio Git; `XDG_CONFIG_HOME` sin definir.

2026-11-07 — `/code-review` sobre `development..8c8f5a77`: 10 hallazgos confirmados al leer el código; ronda de correcciones en el commit siguiente (plan revisado: AC47 → AC53, AC54-AC56).

Checkpoint: stage verify, paso 3. Siguiente: segunda revisión de código limitada a la ronda, revisión del plan revisado y parada de validación.
