# Tasks

Base `development` en `32bef8c8` (antes `8aa361cb`), rama `feat/comment-audit`, worktree `.claude/worktrees/comment-audit`. Todas las tareas corren en el hilo principal: comparten el diseño de la plantilla, cuyo contrato (flags, constantes, salidas) leen T2 y T3, y ninguna supera tres archivos de código o seis de texto. La revisión de código y la de seguridad corren en `flow-verify`.

### T1 — Plantilla del lint gradual y su test

- [x] La plantilla y su prueba pasan con `@babel/parser` 8.1.2 y 7.29.14

**Closes:** AC2

**Depends on:** none

**Locations:** `content/skills/comment-audit/assets/ratchet/check-comments.mjs`, `content/skills/comment-audit/assets/ratchet/check-comments.test.mjs`, `tests/content/comment_audit_template_test.go`, `tests/content/testdata/comment-audit/package.json`, `tests/content/testdata/comment-audit/package-lock.json`, `.gitignore`

**Execution:** main thread

**Test approach:** tdd: la prueba se escribe primero y se ve fallar contra un `check-comments.mjs` que solo sale con 0; es un contrato que otros repositorios copiarán.

**Changes:** la plantilla de `design.md` «Plantilla», escrita de nuevo; ningún nombre de proyecto privado. `package.json` con `"babel8": "npm:@babel/parser@8.1.2"` y `"babel7": "npm:@babel/parser@7.29.14"`, instalado con `npm install --before 2026-11-14` en esa carpeta (ambas versiones pasaron `check_package.py` el 2026-11-22); `.gitignore` agrega `tests/content/testdata/comment-audit/node_modules/`. El test Go de `design.md` «Test Go».

**Verification:** `go test ./tests/content -run TestCommentAuditTemplate -count=1 -v` pasa y muestra, para cada versión de Babel, `ℹ fail 0` y un `ℹ pass` igual al número de casos de `design.md` «Plantilla». Mutaciones en una copia temporal, corriendo `node --test` en esa copia: `lintFile` que no reporta nada debe fallar el caso «comentario nuevo»; quitar la comprobación de `EXCLUDED` debe fallar «ruta excluida»; quitar `--find-renames` debe fallar «renombre»; activar `jsx` en `.ts` debe fallar «cast `<string>x`».

### T2 — La skill `comment-audit` y sus referencias

- [x] `SKILL.md`, `references/ratchet.md`, `references/project-skill.md` y `assets/hive-managed.json` escritos

**Closes:** AC1, AC5, AC6

**Depends on:** T1

**Locations:** `content/skills/comment-audit/SKILL.md`, `content/skills/comment-audit/references/ratchet.md`, `content/skills/comment-audit/references/project-skill.md`, `content/skills/comment-audit/assets/hive-managed.json`

**Execution:** main thread

**Test approach:** check

**Changes:** el procedimiento de `design.md` «Proposed shape», en inglés, con los leaves y skills enlazados por ruta relativa (`no-comments`, `suppression-review`, `defect-fix`, `tests`, `delegation`, `create-issue`, `project-layout`, `flow-build/scripts/check_package.py`) en vez de copiar sus reglas. `description` con su disparador y su límite («not for the comments one change adds, which the no-comments review leaf handles»). `ratchet.md` dice qué hacer con Node < 22.18 (Babel 7.29.x) y en otros lenguajes (cada port permite la marca de atajo con su token de comentario de línea e imprime la ruta del registro en su error), que en CI se pasa siempre `--base` con historial suficiente, y que el regex de la marca comprueba solo la forma.

**Verification:** `go test ./tests/content -count=1` pasa (tamaño de skill y referencias, enlaces, marcador gestionado). Un repositorio temporal, dentro de un workspace declarado temporal, construido siguiendo solo `references/project-skill.md`: `python3 -I content/skills/harness-audit/scripts/skill_registry.py check <repo>` y `… check <workspace>` salen con 0; sin la línea de `## Skills` del repositorio, el primero sale con 1.

### T3 — Reglas de comentarios y de atajos en proyectos con el lint

- [x] `no-comments`, `defect-fix`, `suppression-review` y la página de build del sitio cubren el proyecto con el lint

**Closes:** AC4, AC7

**Depends on:** none

**Locations:** `content/skills/hive-core/review/no-comments.md`, `content/skills/hive-core/build/defect-fix.md`, `content/skills/hive-core/review/suppression-review.md`, `site/src/content/docs/flows/build.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en `no-comments`, un párrafo al final de los pasos, sin renumerar el paso 6. Su disparador: el script de lint o el CI del proyecto corre el chequeo de comentarios. El registro es el que nombra su mensaje de error o la skill generada. El párrafo dice que cada Keep y cada Constraint sin check van ahí con la ruta que afectan, y que eso reemplaza el «keep» de los pasos 4 y 5 y exime del «revert» del paso 6 la entrada del registro. Excluye directivas, licencias y la marca de atajo. Sin un registro nombrado, el Keep se reporta abierto y no se mueve. Agrega al informe las entradas movidas. El mismo párrafo es el hogar único de la marca de atajo en ese proyecto: una línea sola con el token de comentario de línea del lenguaje seguido de `workaround: <ticket>`, o de la entrada del registro (`<ruta>#<ancla>`) mientras no hay ticket, que pasa a nombrar el ticket cuando se crea; la revisión confirma que el destino existe y reporta uno inexistente como defecto. `defect-fix` cambia en su paso 4 «the cause's ticket» por «its ticket» y agrega una cláusula que reemplaza la línea que nombra la causa: «In a project whose lint runs the comment check, the line takes the form [no comments](../review/no-comments.md) sets instead.», sin pasar de 4.000 B: en `32bef8c8` el leaf mide 3.960 B, y la cláusula suma 119 B netos, así que se recortan al menos 100 B de palabras del mismo paso 4 sin cambiar su sentido, se comprueba `wc -c` ≤ 4.000 antes de la suite, y nunca se agrega una excepción de tamaño; `suppression-review` y la página de build del sitio enlazan el mismo párrafo sin repetir la forma. `flow-deploy/references/urgent-fixes.md:11` se relee por nombre y no cambia si el enlace de `defect-fix` cubre la mitigación sin ticket.

**Verification:** `go test ./tests/content -count=1` pasa (tamaño de leaves, frases únicas, enlaces); `rg -n 'no-comments|Delete every Kill|cause comment|workaround' content site _support/docs --glob '!_support/docs/history/**'` no muestra ningún texto que contradiga el párrafo nuevo.

### T4 — Catálogo, arquitectura y changelog

- [x] La skill aparece en el catálogo y su hogar está documentado

**Closes:** AC1

**Depends on:** T2, T3

**Locations:** `llms.txt`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** una línea por catálogo con el mismo resumen. En `instruction-resources.md`, la entrada de hogar canónico de `comment-audit` y los cambios de los tres leaves. En `CHANGELOG.md`, «Added» y «Changed» bajo `[Unreleased]`. `README.md` no lista skills y no cambia.

**Verification:** `go test ./... -count=1` y `go vet ./...` pasan. El sitio construye: enlazar `site/node_modules` al checkout principal, correr `npm run build` en `site/` y quitar el enlace después, como dice `HIVE.local.md`; sale con 0.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | Al cerrar cada tarea, en el worktree | El comando de su `Verification:` | Cubierto por la implementación |
| Full suite | Candidato final, en el worktree | `go test ./... -count=1` y `go vet ./...` | Cubierto por la implementación |
| Cláusulas del prompt (AC5) | Candidato final | Tabla de cláusula a paso con línea de `SKILL.md`, en `### Evidence` | Cubierto por la implementación |
| Code review | Candidato final, antes del push | `Native: Claude Code /code-review` | Review A |
| Security review | Candidato final, antes del push | Hijo `hive-review-security` sobre la plantilla y el test Go | Security A |
| AC1, AC4 y AC7 en la instalación | Tras el merge, en esta máquina | `hive update --source <checkout>`, `ls -l ~/.claude/skills/comment-audit`, lectura de los leaves y del packet instalados | `AGENTS.md` «Refreshing a local installation» (maintainer) |

Sin parada de validación (`automatic`).

## Checkpoint

- Updated: 2026-11-22
- Stage: build · T1
- Last commit: 32bef8c8
- Open findings: none
- Next: T1: write the template test first
- Worktree: .claude/worktrees/comment-audit · feat/comment-audit
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-22 — Plan escrito.
2026-11-22 — Revisión 3: la forma de la marca vive solo en `no-comments`, porque `defect-fix` tiene 3.882 B de 4.000; la revisión comprueba su destino; casos de prueba de Babel ausente y de JSX inline.
2026-11-22 — Revisión 2 tras la relectura: plugins de Babel por extensión, parser con `import()`, `--same-code` en JSX, `ALLOWED` ampliado, marca por lenguaje y con destino de registro, Keep abierto sin registro.
2026-11-22 — Revisión 1 tras la revisión del plan: D5 y D6; T1 con Babel en dos versiones, salidas de error, renombres, subcarpetas, JSX, directivas y `--same-code`; T2 con correspondencia de categorías y registro en workspaces; T3 amplía a `defect-fix`, `suppression-review` y la página de build; AC3 retirado, AC6 y AC7 nuevos.
2026-11-22 — T1 · d32b88cb · unverified: mutation of --find-renames is caught only because the rename test sets diff.renames=false; the Go test skips on machines without the pinned parsers (npm ci --prefix tests/content/testdata/comment-audit)
2026-11-22 — T2 · ec3528b8 · unverified: AC1 waits for the installation after the merge; AC5's clause table is written in verify
2026-11-22 — T3 · ec518780 · unverified: AC4 and AC7 wait for the installation after the merge; defect-fix is at 3,989 of 4,000 B (87 B trimmed, not the 100 B the plan aimed for)
2026-11-22 — T4 · f210537d · unverified: AC1 waits for the installation after the merge
