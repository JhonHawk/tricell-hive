# vendored-exclusion — audit-comments deja fuera el código vendorizado sin perder el chequeo de lotes

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`0aad427f`) · worktree `.claude/worktrees/vendored-exclusion` · rama `feat/vendored-exclusion` |
| Playbook | feature |

## Decisions

2026-11-23
- **D32 A** — ¿Cómo se protege el código vendorizado cuando la auditoría lo deja fuera? `--same-code` reporta cualquier cambio en una ruta de `EXCLUDED`, con su test en la suite del template; el texto agrega las rutas vendorizadas desde el paso 2, da a cada carril sus líneas de inventario y corre `--same-code` al final. Descartado: solo texto (`git diff --name-only` y `git ls-files --others` en cada lote), que depende de que el agente recuerde dos comandos.
- **D33 (supuesto)** — Las rutas de `EXCLUDED` se escriben completas desde la raíz, como ya las compara `excluded()` (`startsWith`); los valores por defecto (`node_modules/`, `dist/`, `build/`, `coverage/`) se conservan y el paso 2 agrega, no reemplaza, porque la suite del template depende de `dist/`. Descartado: comparar por nombre de carpeta en cualquier nivel, que cambiaría el lint de proyectos que ya lo copiaron. Revertir: cambiar `excluded()`.
- **D34 (supuesto)** — En `--same-code`, un cambio bajo una ruta excluida falla con `<file>: changed under an excluded path`, también si un comando del lote regeneró un archivo generado que está versionado: un lote que solo quita comentarios no debe tocarlo. El lint normal sigue saltando esas rutas. Revertir: quitar el reporte en `sameCode`.
- **D35 (supuesto)** — El chequeo cubre cualquier archivo versionado bajo una ruta excluida, de cualquier tipo, no solo los de `SOURCE_PATTERNS`, porque un lote con un patrón mecánico puede tocar un CSS o un JSON vendorizado. Descartado: limitarlo a los archivos fuente. Revertir: filtrar `excludedChanges` por `SOURCE_PATTERNS`.
- **D36 (supuesto)** — Los archivos sin versionar bajo una ruta excluida no cuentan como cambio, porque los tests y builds del paso 7 los dejan (`coverage/`, `build/`) y el commit de borrado nombra sus rutas. Descartado: reportarlos y pedir que se borren. Revertir: sumar `git ls-files --others` a `excludedChanges`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «resolvamos los hallazgos» · `/flow-plan` sobre H3 (2026-11-23)

Origen: H3 del cierre de `ledger-writer-role`. Un primer intento en el PR #605 puso el código vendorizado en `EXCLUDED` y se revirtió: la segunda revisión mostró que el inventario y los carriles lo seguían leyendo, que `--same-code` dejaba de cubrirlo, que la skill del proyecto no tenía el chequeo y que reemplazar los valores por defecto rompía la suite.

## Objective

Hoy `audit-comments` incluye el código vendorizado, así que los hijos leen comentarios que el proyecto no escribe y que pueden dirigirse a agentes. Al terminar, el paso 1 lo deja fuera por ruta, el inventario y los carriles no lo ven, y `--same-code` rechaza un lote que cambie cualquier archivo excluido, de modo que sacarlo de la auditoría no lo saca del chequeo de lotes.

## Scope and acceptance

Dentro: `check-comments.mjs` y su suite; `audit-comments/SKILL.md` pasos 1, 2, 4 y 7; `references/ratchet.md` (las líneas de los tres modos, las viñetas `EXCLUDED` y `ALLOWED`, la frase de los pasos 2 y 8 y el contrato de un port); `references/project-skill.md` punto 6; el nombre del test de Go en `instruction-resources.md` y en `tests/content/testdata/audit-comments/package.json`; la entrada de `CHANGELOG.md`.

Fuera: cambiar `excluded()` a coincidencias en cualquier nivel (D33); `audit-tests`; los proyectos que ya copiaron el template.

- AC1 — `node check-comments.mjs --same-code --base <commit>` sale con 1 y nombra el archivo cuando un archivo versionado bajo una ruta de `EXCLUDED`, de cualquier tipo, cambió, se agregó, se borró o salió de esa ruta desde `<commit>`, también si no parsea; un archivo sin versionar bajo esas rutas no cuenta, y un cambio que solo quita comentarios fuera de ellas sigue saliendo con 0 · false today because `changedFiles()` descarta las rutas excluidas y `--same-code` sale con 0 · Observed in: el test nuevo de `check-comments.test.mjs`, que corre el script real sobre un repositorio Git temporal, vía `go test ./tests/content/ -run TestAuditCommentsTemplate` en el candidato
- AC2 — El `audit-comments/SKILL.md` instalado: el paso 1 deja fuera el código vendorizado, cada ruta completa desde la raíz; el paso 2 agrega esas rutas al `EXCLUDED` de la copia temporal antes del primer inventario, sin quitar los valores por defecto; el paso 4 da a cada carril un archivo propio con sus líneas de inventario, con modo `0600` en la carpeta temporal, y el hijo marca solo esas; el paso 7 corre `--same-code` como último chequeo, después del lint, el type check y los tests · false today because ninguno de los cuatro pasos lo dice en `0aad427f` · Observed in: `~/.agents/skills/audit-comments/SKILL.md` instalado tras el merge
- AC3 — Las referencias instaladas: `ratchet.md` dice que `EXCLUDED` se compara desde la raíz e incluye el código vendorizado, que el lint y el inventario lo saltan y que `--same-code` rechaza un cambio ahí; su viñeta `ALLOWED` remite a la de `EXCLUDED`; el contrato de un port pide el modo `--same-code` con ese rechazo y su caso de suite; `project-skill.md` punto 6 corre `--same-code --base <the batch's record commit>` al final · false today because ninguna lo dice en `0aad427f` · Observed in: `~/.agents/skills/audit-comments/references/` instalado tras el merge

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Base `development` en `0aad427f` (PR #605) · `git log` · 2026-11-23.
- `check-comments.mjs:55` compara `EXCLUDED` con `startsWith` desde la raíz; `:194-200` `changedFiles()` descarta las rutas excluidas salvo un renombre que entra a una; `:320` `sameCode()` usa esa lista; `:258` el inventario salta las excluidas.
- La suite usa `dist/` como ruta excluida en «excluded paths are skipped» (`check-comments.test.mjs:165`) y en el movimiento a `dist/a.js` (`:357`).
- `tests/content/audit_comments_template_test.go` corre la suite con `@babel/parser` 8 y 7, y la salta sin Node o sin los parsers fijados (`:28`, `:82`).
- La spec `comment-audit` vive todavía solo en los deltas pendientes de `comment-audit` y `audit-prefix`; este delta agrega un requisito a esa capacidad, que un lote de higiene archiva junto.

- `go test -count=1 ./tests/content/ -run TestAuditCommentsTemplate` · pass, 7.4 s (babel8 y babel7) · 2026-11-23
- `go test ./tests/content/...` · pass · 2026-11-23
- En la base, sin coincidencias (exit 1): `rg -c 'changed under an' …/references/ratchet.md`, `rg -c 'vendored' …/SKILL.md`, `rg -c '0600' …/SKILL.md`, `rg -c 'record commit' …/references/project-skill.md` · 2026-11-23
- `rg -c 'comment_audit_template_test' _support/docs/architecture/instruction-resources.md tests/content/testdata/audit-comments/package.json` · 1 en cada uno en la base · 2026-11-23
- `go test ./...` · pass, salvo una corrida de `tests/pilot` que agotó el tiempo de `gh auth status` y pasó sola (sobre `cf1b3ef1`, mismas entradas) · 2026-11-23

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **Dónde** — Worktree `.claude/worktrees/vendored-exclusion`, rama `feat/vendored-exclusion` desde `development`.
- **Delivery** — `automatic`; **Review** — `Native: Claude Code /code-review`; las respuestas vigentes de esta sesión. No toca una frontera de seguridad de la lista, así que no lleva la pregunta de seguridad.
- El PR lista los lugares revisados y dice que `content/guidance/global.md` no cambia.

## Review status and progress

2026-11-23 — Revisión `hive-review-plan` (guidance): 1 P1, 5 P2 y 7 P3, todos aplicados. P1: cada carril recibe un archivo de inventario `0600`, no las líneas en el brief. P2-1 y P2-2: D35 (cualquier tipo de archivo versionado) y D36 (sin versionar no cuenta). P2-3: base del punto 6. P2-4: las tres líneas de modos de `ratchet.md`. P2-5: el reporte va antes de parsear. P3: nombre del test de Go, entrada del changelog, `rg` previos al merge, `-v` en T1, redacción del escenario, modo `--same-code` en el contrato del port, casos de inventario y de salida de una ruta excluida.

2026-11-23 — La relectura dio los 13 hallazgos por resueltos y encontró dos P3: N1, `excludedChanges` refiltra con `excluded()`; N2, el punto 6 pide preparar el commit de borrado por nombre de archivo. También se ordenaron D34–D36 y el caso de salida exige la línea de `dist/a.js`.

2026-11-23 — `/code-review` (high): 0 importantes, 6 menores y 2 previos. Corregidos los 6: un movimiento entre una ruta excluida y otra se nombra una vez por lado; caso de un archivo nuevo en el índice y salida 0 al revertir; `0600` para el inventario y todo archivo con texto de comentarios desde el paso 2; el paso 7 nombra el lint de comentarios; el contrato del port remite a la línea de `--same-code`; `--ignore-submodules=dirty`. Previos, al cierre: archivos no fuente que los tests reescriben fuera de `EXCLUDED`, y archivos generados dispersos que `EXCLUDED` no expresa sin globs.
