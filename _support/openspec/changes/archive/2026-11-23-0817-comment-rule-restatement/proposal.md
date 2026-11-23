# comment-rule-restatement — un comentario que repite una regla que el código ya expresa se borra, y la auditoría deja la regla de comentarios en el AGENTS.md del proyecto

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #609, merge b0f1955f |
| Tracker | ninguno |
| Git | tricell-hive · base `development` (`f550f3e6`) · rama propuesta `feat/comment-rule-restatement` |
| Playbook | feature |

## Decisions

2026-11-23
- **D1 A** — ¿Cómo seguimos tras revisar las sesiones de auditoría de comentarios? Llevar a Hive el criterio de los comentarios con reglas a medias (H3) y la sección de comentarios en el `AGENTS.md` del proyecto (H4), con plan. Engram: `decision/comment-half-rule-criterion`.
- **D2** — Siguiente paso: `/flow-plan`, aceptado por comando.
- **D3 (supuesto)** — El criterio entra una sola vez, en el Kill de `no-comments`; la categoría Tautology de `audit-comments` lo nombra y remite al Kill, sin repetir su texto, y `tests/content/single_home_test.go` fija la frase en su casa. Descartado: copiar la cláusula completa en los dos archivos, que es como hoy se repite el Kill y puede divergir. Revertir: quitar la fila del test y copiar la cláusula.
- **D4 (supuesto)** — Un comentario cae en Kill por esta cláusula solo si la evidencia nombra la línea, el test o el ajuste que expresa la regla; la parte del comentario que el código no expresa se clasifica aparte (se recorta el comentario a esa parte y se clasifica por sí sola). Así se frena la ampliación de Pi (H2), que borraba también reglas de negocio que el código no mostraba. Descartado: borrar ante la duda, como pstack. Revertir: quitar la segunda frase de la cláusula.
- **D5 (supuesto)** — La sección `## Comments` va en el `AGENTS.md` del repositorio cuyo lint corre el ratchet, con una plantilla en `references/project-skill.md` que el paso 9 llena con la ruta del registro de decisiones, el tracker y el comando del lint; si el archivo ya tiene `## Comments`, se actualiza en lugar de agregar otra. El workspace declarado no la recibe. Descartado: una referencia nueva y un paso 10, que suman texto sin otra condición de lectura. Revertir: quitar la viñeta y la plantilla.
- **D6 (supuesto)** — Ningún detalle contradicho entra al Kill nuevo: la parte que el código no expresa, un detalle que el código contradice incluido, queda como su propio comentario, recortado a esa parte, y se clasifica por sí sola como el paso 1 establece (un Keep pasa por el paso 4). Corrige la redacción anterior («or gets a detail wrong»), que permitía borrar el único registro de una regla que el código cumple mal. Descartado: borrar el detalle contradicho, que es la ampliación de Pi (H2). Revertir: devolver la frase al Kill. **Replaces:** las frases «gets a detail wrong» y «el resto se va con el comentario» de D4.
- **D8 A** — En la auditoría, ¿un detalle que el código contradice se sigue borrando como Stale? Sí, como hoy: la categoría Stale de `audit-comments` no cambia y sigue borrando lo que el código o su historia contradicen; la regla nueva no lo borra por sí sola. Queda la incertidumbre de un comentario que sea el único registro de una regla que el código cumple mal.
- **D7 (supuesto)** — La sección abre con la prohibición que pediste unida al lint («Do not add comments: `<lint>` rejects a new or returned comment in `<paths>`»), y después dice a dónde va cada hecho: no promete una prohibición más ancha que el lint, y lo que sigue a la primera frase no lo cubre ningún mecanismo; una auditoría de harness (HA-IF-05) puede proponer quitar esa primera frase, y entonces decides tú. Una restricción sin check va al registro de decisiones, y el check se propone antes de agregarlo, como el paso 5 de `no-comments`. Descartado: una excepción en HA-IF-05 de audit-harness, que agrega un archivo al cambio. Revertir: volver a «Do not add comments to source code» sin rutas.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «Revisa sesiones de cam/carvento con codex y de fleetsystem con pi, añadí una regla que está resolviendo varios comentarios» · respuesta A a D1 · `/flow-plan` (2026-11-23)

Regla que el usuario dio en dos sesiones reales independientes, el 2026-11-23: «comentarios con reglas a medias, se quitan, es parte de la tautología, si el código se puede autoexplicar, no se necesita el comentario» (cam, Codex `gpt-6.2-sol` high, sesión `01a20548`, 07:24 UTC; fleetsystem, Pi, sesión `01a206d8`, 07:21 UTC). En fleetsystem también pidió «añadir la regla a los AGENTS/CLAUDE.md de que ya no están permitidos los comentarios (y cuándo sí)» (05:34 UTC); Pi la redactó e integró en `fleetsystem-backend-monorepo/AGENTS.md` y en el frontend.

## Objective

Hoy un comentario que enuncia una regla que el código ya implementa cae en Keep, y el paso 4 de `no-comments` lo conserva y lo reporta abierto cuando no está claro si guarda un hecho: así se acumularon los pendientes de cam y fleetsystem. Además la auditoría solo registra la skill del proyecto en `## Skills`; nada le dice a quien edita que no agregue comentarios ni a dónde va cada hecho. Al terminar, ese comentario es Kill cuando la evidencia nombra lo que expresa la regla, y el paso 9 deja la sección `## Comments` en el `AGENTS.md` del repositorio auditado.

## Scope and acceptance

Dentro: `hive-core/review/no-comments.md` (Kill y línea de reporte); `audit-comments/SKILL.md` (descripción, Tautology del paso 4, pasos 9 y 10); `audit-comments/references/project-skill.md` (plantilla y registro); `tests/content/single_home_test.go` (una fila); `_support/docs/architecture/instruction-resources.md` (la descripción de `project-skill.md`); `site/src/content/docs/skills.md` (la línea de la auditoría); `CHANGELOG.md`.

Fuera: la categoría Stale de `audit-comments` (D8); el paso 4 de `no-comments` (ambiguo → se conserva y se reporta) queda igual; la decisión `decision/comment-rule-no-ban` sigue: Hive no prohíbe comentarios en general, la sección solo llega a los proyectos que corren la auditoría; los proyectos que ya la corrieron (cam, carvento, fleetsystem) no se tocan; `content/guidance/global.md` no cambia.

- AC1 — El `no-comments.md` instalado clasifica como Kill un comentario que repite una regla que el código, sus pruebas o su configuración en el repositorio ya expresan, aunque la enuncie solo en parte, siempre que la evidencia nombre la línea, el test o el ajuste que la expresa; la parte que el código no expresa, un detalle que el código contradice incluido, queda como su propio comentario, recortado a esa parte, y se clasifica por sí sola como el paso 1 establece; y el reporte de la revisión lista cada Kill de este tipo con lo que nombra · false today because el Kill de `f550f3e6` solo cubre narrar el código, repetir un nombre o un tipo, marcar secciones, código comentado y cambios pasados (`rg -c 'restates a rule' content/skills/hive-core/review/no-comments.md` sin coincidencias) · Observed in: `~/.agents/skills/hive-core/review/no-comments.md` y el paquete `~/.claude/skills/flow-verify/stage-entry.md` instalados tras el merge
- AC2 — En el `audit-comments/SKILL.md` instalado, la categoría Tautology incluye el comentario que repite una regla que el código expresa y remite al Kill del leaf, sin copiar su cláusula; la frase de la cláusula aparece en un solo archivo de `content/` · false today because Tautology dice solo «what the code, a name or a type already says» · Observed in: `~/.agents/skills/audit-comments/SKILL.md` instalado tras el merge, y `TestRulesHaveOneHome` en el candidato
- AC3 — El paso 9 instalado escribe en el `AGENTS.md` del repositorio auditado una sección `## Comments` desde la plantilla de `project-skill.md`, con el comando del lint del proyecto, las rutas que revisa, la ruta del registro de decisiones y el tracker: prohíbe agregar comentarios, dice que el lint rechaza un comentario nuevo o devuelto en esas rutas, a dónde va una decisión, un atajo y una restricción (sin check, al registro; el check se propone antes de agregarlo) y lo único que se queda (directivas, licencias y el marcador de workaround, con «the ticket's key or URL»); actualiza una sección existente en lugar de duplicarla, no la escribe en el workspace y reporta un `CLAUDE.md` de ese nivel que no importe `@AGENTS.md`; el punto 5 de la skill del proyecto remite a la sección para la forma del marcador; la descripción y los pasos 9 y 10 la nombran · false today because `project-skill.md` solo agrega la línea de `## Skills` (`rg -c '## Comments' content/skills/audit-comments/` sin coincidencias) · Observed in: `~/.agents/skills/audit-comments/` instalado tras el merge
- AC4 — La página de skills del sitio dice que la auditoría deja la regla de comentarios en el `AGENTS.md` del proyecto · false today because la línea 46 de `site/src/content/docs/skills.md` no lo menciona · Observed in: `site/dist/skills/index.html` del build del candidato

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Base `development` en `f550f3e6`, checkout limpio · `git status` · 2026-11-23.
- `no-comments.md:8` (Kill) y `audit-comments/SKILL.md:22` (Tautology) son los dos únicos lugares que enuncian el criterio (`git grep -i -E "tautolog|narrates the code|restates a name"` en `content`, `integrations`, `site`, `_support/docs`).
- `no-comments` solo entra al paquete `flow-verify/stage-entry.md`, que mide 17 491 B contra el tope de 23 500 B (`TestRealStagePacketsKeepAMarginUnderTheCap -v`).
- Dependientes revisados: `content/guidance/global.md:122` (solo el puntero), `flow-verify/SKILL.md:8,16` (el nombre del leaf), `defect-fix.md`, `suppression-review.md:7`, `ratchet.md:12` (`DECISIONS_PATH`), `site/src/content/docs/{skills.md:46,index.mdx:106,flows/build.md}`, `_support/docs/architecture/{agent-delivery,instruction-resources}.md`; `tests/pilot/` no nombra estos textos. Solo cambian los de «Dentro».
- La spec `comment-audit` vive en los deltas pendientes de `comment-audit`, `audit-prefix` y `vendored-exclusion`; este delta agrega dos requisitos a esa capacidad.
- Plantilla de origen: `fleetsystem-backend-monorepo/AGENTS.md` `## Comments` (líneas 29–43), escrita por Pi en la sesión `01a206d8`.

- Text check de `AGENTS.md` «Checks by path» (cuatro comandos) · pass, 90 s · 2026-11-23
- `go test -count=1 ./tests/content/...` · pass · 2026-11-23
- `pnpm check`, `pnpm lint` y `pnpm build` en `site/` · pass (0 errores, 20 páginas), 4 s · 2026-11-23
- `go test -count=1 -run TestRulesHaveOneHome ./tests/content/` · pass, 0.1 s · 2026-11-23
- `rg -c 'comment rule' site/dist/skills/index.html` · sin coincidencias en el build de la base (observado por la revisión del plan) · 2026-11-23
- `rg -c 'restates a rule' content/skills/hive-core/review/no-comments.md` y `rg -c '## Comments' content/skills/audit-comments/` · sin coincidencias (exit 1) · 2026-11-23

## Effects

None: the tasks take no effect on an external system.

## Delivery

- **Dónde** — Worktree `.claude/worktrees/comment-rule-restatement`, rama `feat/comment-rule-restatement` desde `development`; la carpeta de cambio se mueve allí al empezar el build.
- **Delivery** — `automatic` (respuesta de la ronda de entrega, 2026-11-23): revisión, push, PR y merge sin parada de validación; después, refresco local por el marcador de mantenedor.
- **Review** — `Native: Claude Code /code-review`, como prefiere `AGENTS.md`. No toca una frontera de seguridad. El PR lista los lugares revisados y dice que `content/guidance/global.md` no cambia.

## Review status and progress

2026-11-23 — Revisión `hive-review-plan` (guidance, 200 s, 148k tokens): 1 P1, 6 P2 y 7 P3. Aplicados: P1-1 (sin «gets a detail wrong»; la parte no expresada queda como comentario propio, escenario mixto en el delta; D6), P2-1 y P2-4 (restricción sin check al registro, rutas del lint; D7), P2-2 (marcador «key or URL», punto 5 como puntero, `instruction-resources.md`), P2-3 (apertura de la plantilla; D7), P2-5 (reporte de HA-HI-03), P2-6 (Kill de regla en el reporte), y los P3 de alcance «in the repository», paso 10, orden de secciones y verificaciones de T2. Sin aplicar: el lugar de respaldo para `hold` (la entrega es `automatic`).

2026-11-23 — Relectura (65 s): resueltos todos los hallazgos anteriores. Nuevos: P2-A (el Stale de la auditoría sigue borrando un detalle contradicho; pasa al usuario como D8) y cuatro P3 aplicados: D4 y D6 alineados, D7 alineado con «Impact», la sección vuelve a abrir con «Do not add comments» unida al lint, y la redacción del delta y de `tasks.md`.

2026-11-23 — `/code-review` (high) sobre `7dfdac23`: 1 importante y 5 menores; sin supresiones agregadas. Corregidos en `b37cd578` (Unreviewed, solo Markdown): el comentario mixto tiene una fila del ledger por parte (importante); el changelog decía «cut out»; el marcador dice «record entry»; la sección nombra dónde va `## Skills` cuando falta. Al cierre: puntero en el workspace a la sección de cada repo hijo (HA-HI-05), y HA-IF-05 frente a la primera frase de la sección.
