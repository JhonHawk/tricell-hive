# Tareas

Base `2eef1969` de `development`, rama `feat/reuse-sound-base`. Dos tareas pequeñas de guía con la forma ya fijada en `design.md`: ambas en el hilo principal (ninguna pasa de 5 lecturas grandes ni de 3 archivos de guía). El checkout principal lo usa otra sesión activa (rama `feat/close-cleans-at-completion`, con commits cada pocos minutos): el build trabaja en un worktree bajo `.claude/worktrees/` desde `origin/development`, y mueve allí esta carpeta antes de la primera edición, para que el árbol de la otra sesión no la contenga. Ningún archivo que edita este plan lo cambia la otra rama.

### T1 — Peldaño 2: reutilizar solo una base sana

- [x] La hoja `reuse-before-writing` y el rol `hive-review-code` dicen lo que fija `design.md` «Forma propuesta», con la prueba de hogar único y la asignación de lectura que decida D1.

**Closes:** AC1, AC2, AC4

**Depends on:** none

**Locations:** `content/skills/hive-core/references/reuse-before-writing.md`, `content/agents/review/hive-review-code.md`, `tests/content/single_home_test.go`, `tests/content/reading_footprint_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** el peldaño 2, y en el revisor la cláusula de copia y la exención del paso 2 y el escenario del paso 3; la asignación se mide en la rama del build, no en la de otra sesión. No cambian los peldaños 1, 3, 4 y 5, la frase «Prefer deleting to adding», `defect-fix` ni `global.md`.

**Verification:** `go test ./tests/content/ -count=1` pasa en la rama del build y en el worktree temporal con `feat/close-cleans-at-completion` fusionada (ver `design.md`) (paquete de etapa ≤ 22.500 B, unión de lectura bajo la línea base más la asignación, hogar único, enlaces); en la base pasa. `rg -n reuseSoundBaseAllowance tests/content/reading_footprint_test.go` da la constante y su uso (en la base, nada). `rg --count-matches 'CODING_STANDARDS|boolean flag|diverging' content/skills/hive-core/references/reuse-before-writing.md` da 3 (en la base, nada y sale con 1). `rg -n 'copy or extend' content/agents/review/hive-review-code.md` da 1 línea (en la base, ninguna). `rg -n 'copy a workaround|copied workaround' content/` no da resultados (en la base da `hive-review-code.md:12` y `:14`). AC1 y AC2 se observan en los archivos instalados tras `hive update`, después de integrar.

### T2 — La página del flujo build y el CHANGELOG

- [x] `site/src/content/docs/flows/build.md` describe la base sana como fija `design.md`, y `CHANGELOG.md` lleva su entrada bajo `[Unreleased]`.

**Closes:** AC3

**Depends on:** T1

**Locations:** `site/src/content/docs/flows/build.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en la línea 26, la frase del workaround se reemplaza por la frase de `design.md`; no lista las condiciones. En `CHANGELOG.md`, una entrada al final de la lista `### Changed` de `[Unreleased]`, no al principio, donde la otra rama añade la suya: así los dos PR no tocan líneas contiguas. No cambia otras páginas.

**Verification:** en el worktree, primero `pnpm --dir <worktree>/site install --frozen-lockfile --offline` (sin tocar el lockfile ni la red; si necesita la red, se para y pasa antes el chequeo de paquetes global), luego `pnpm --dir <worktree>/site check` pasa (en el checkout principal: 0 errores, 0 avisos). `rg -n 'reuses only code that is sound' site/src/content/docs/flows/build.md` da 1 línea y `rg -n 'does not copy an existing workaround' site/src/content/docs/flows/build.md` ninguna (en la base: ninguna y la línea 26). Antes del push, si `feat/close-cleans-at-completion` ya se integró, rebase y otra vez `go test ./tests/content/ -count=1`; si no, el cuerpo del PR lleva el aviso de `design.md`.

## Verificación y revisión humana

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | Al cerrar cada tarea, local | Los comandos de su `Verification:` | Cubierto por la implementación |
| Full suite | Candidato final, local | `go test ./... -count=1` y el CI local que lea el build | Cubierto por la implementación |
| UI check | n/a: no hay pantalla | — | — |
| Code review | Candidato final, antes del push o de la parada de validación | El mecanismo que responda la ronda | La respuesta de revisión |

Revisión humana: leer el peldaño 2 nuevo y el texto del revisor en el diff.

## Checkpoint

- Updated: 2026-11-21
- Stage: verify · 6 push
- Last commit: 9fa7cf19
- Open findings: spec rationale: 'a workaround whose comment names its cause, which a user already accepted' is false for an incident mitigation · close finding; an accepted workaround is re-asked on each later task and may get a duplicate ticket proposal · close finding (user decision); spec scenario title 'too large to fix in passing' implies small diverging duplication is fixed silently · close finding; the second review does not cover what the first missed · noted
- Next: push, PR ready, CI wait, PR reviews, stop
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/reuse-sound-base · feat/reuse-sound-base
- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md
- Principles: none

### Evidence

- before captures: none (no screen)
- go vet ./... · exit 0 · b0b2f452
- go test ./... -count=1 · all ok · b0b2f452
- python3 -m unittest discover -s tests/skills -p '*_test.py' · 318 OK · b0b2f452
- go test ./tests/content/ -count=1 · ok; union 194910 B < 194911 B; flow-build packet 22457 B · b0b2f452
- pnpm --dir site check · 0 errors (install --offline exit 1, direct deps linked) · b0b2f452
- local CI: only master-source-branch.yml (pull_request_target to master), not applicable to a PR to development · b0b2f452
- walk AC1 (candidate file): 3 matches; AC2: reviewer lines present; AC3: 1 line; AC4: constant + test ok · b0b2f452; AC1/AC2 installed-file observation pending hive update after merge
- no-comments: 1 added comment kept (allowance reason); suppressions: none · b0b2f452
- go vet + go test ./... -count=1 all ok; unittest OK · 9fa7cf19
- go test ./tests/content/ ok; union 194963 B < 194964 B; flow-build packet 22498 B · 9fa7cf19
- pnpm --dir site check 0 errors · 9fa7cf19
- /code-review round 1: 3 important + 5 nits fixed in 9c35281a; round 2 (fix diff): 2 blocking fixed locally in 9fa7cf19 (Unreviewed), 4 nits to close · 9fa7cf19

## Estado de revisión y progreso

2026-11-21 — T1 · c5a5a2be · unverified: AC1 y AC2 se observan en los archivos instalados tras hive update, después de integrar
2026-11-21 — T2 · 90e471cc · unverified: none
