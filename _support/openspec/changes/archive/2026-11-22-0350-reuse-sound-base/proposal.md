# reuse-sound-base — reutilizar solo una base sana y sanearla primero

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #575, e83104ef |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `f80c6843` (planeado en `2eef1969`) · rama `feat/reuse-sound-base` |
| Playbook | investigate → cambio de reglas (ninguno) |

## Decisiones

2026-11-21
- **Reply A (investigación)** — ¿Se adopta ya que el peldaño 2 de «Reuse before writing» reutilice solo una base sana y la sanee primero, sin esperar a que el fallo se repita? Sí: el usuario aprueba la corrección sin las dos observaciones que pide «Measurement» (`engram:tricell-hive/decision/reuse-sound-base`). Comparación de las tres versiones: pstack no tiene escalera de reutilización y rediseña sin límite; Hive antes de `a5ec06a8` copiaba sin excepción; Hive actual solo excluye el workaround marcado y no da ruta para sanear.
- **Supuesto** — La condición (c) es una bandera booleana o una rama dentro de la base que solo el código nuevo activa, no cualquier parámetro nuevo: generalizar una base con un parámetro que sirve a todos sigue siendo reutilizar. Alternativa descartada: cualquier parámetro que solo use el llamador nuevo (bloquearía extensiones legítimas). Para revertirlo, se amplía la condición (c) en la hoja.
- **Supuesto** — El saneamiento sin preguntar se limita como el arreglo en la causa de `defect-fix`: en cualquier lugar del repositorio si es un cambio pequeño y entendido; si no, enfoque del plan o reporte antes de construir encima. La hoja lo dice como excepción al criterio **a** de `questions` (tocar otro módulo sin pedirlo), y el revisor de código no lo marca como cambio no pedido. Alternativa descartada: solo en los archivos que la tarea ya cambia (dejaría sin sanear el helper de otro archivo que la tarea reutiliza). Para revertirlo, se acota la frase en la hoja.
- **Supuesto** — Solo se sanea la base que la tarea iba a reutilizar, no cada violación de estándares del entorno. Alternativa descartada: sanear todo lo que la tarea lee. Para revertirlo, se amplía la hoja.
- **Supuesto** — La ruta del build (con o sin plan) no se reescribe: apunta a `defect-fix` paso 4 con la base en lugar de la causa, un hijo lo reporta a su padre, y si el usuario acepta construir encima se marca como workaround según ese paso. La elección queda en un solo hogar. Alternativa descartada: copiar la elección en la hoja de reutilización.

- **D1 A** — ¿Cómo se paga el crecimiento de la unión de lectura plan/build (unos 800 B, con 17 B de margen)? Con una constante de asignación nueva en `reading_footprint_test.go`, igual a la medida tras T1 y con su comentario; no se recorta el texto de otras reglas.

- **D2 A** — ¿Qué hace el build con un workaround ya aceptado, cuyo comentario nombra su causa? Lo nombra con su ticket, propone uno solo si no existe y construye encima sin volver a preguntar. Descartadas: preguntar siempre (B) y no construir nunca encima (C). Resuelve el hallazgo P2 del cierre; los dos P3 del spec se corrigen con ella.
- **Supuesto** — «Aceptado» se prueba por el comentario: un workaround cuyo comentario nombra el ticket de su causa (el paso 4 de `defect-fix` ahora lo pide al marcarlo) se mantiene o se llama sin volver a preguntar y nunca se copia; uno sin ticket, como un comentario anterior a Hive, sigue la ruta de cualquier base no sana. Alternativa descartada: tratar todo comentario de causa como aceptado (la revisión mostró que salta la elección del usuario y no permite encontrar el ticket). Para revertirlo, se quita la condición del ticket en la hoja y en `defect-fix`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Pedido

> «En contraste con Poteto, creo que nosotros sí tenemos una regla de reutilización de código. ¿Cuál es el problema? Que cuando ese código está mal hecho o fue una especie de aspirina para el momento y el agente llega a reutilizar eso que está o funcional pero no bien, no con las buenas prácticas, refactorizar en el futuro se vuelve sumamente complejo. Y P-Stack prefiere una solución de raíz antes de atacar la implementación. Valídalo y dame una propuesta.»

Respuesta del usuario a la propuesta: «A» (adoptarla ya y preparar su plan).

## Objetivo

Hoy el peldaño 2 de `hive-core/references/reuse-before-writing.md` reutiliza lo que el código ya tiene «como lo usa el código de alrededor», salvo un workaround marcado por un comentario o un guard de síntoma; lo que funciona pero no sigue las buenas prácticas se copia o se extiende, y la excepción no dice qué hacer en su lugar. Con el cambio, se reutiliza solo una base sana; una que no lo es se sanea primero cuando es un cambio pequeño y entendido, y si no, se ofrece como enfoque del plan o se reporta antes de construir encima. La revisión de código marca el código nuevo que copia o extiende una base no sana.

## Alcance y aceptación

Dentro: la hoja `reuse-before-writing`, el rol `hive-review-code`, la prueba de hogar único y la asignación de lectura de `reading_footprint_test.go`, la página `site/.../flows/build.md`, la entrada de `CHANGELOG.md` y el delta de `workflow-lifecycle`. Fuera: `defect-fix` (su paso 4 se reutiliza tal cual), `flow-plan` paso 2 (ya lee la hoja al pesar enfoques), los playbooks y los roles de build (solo apuntan a la hoja), `global.md` (no cambia).

- AC1 — La hoja instalada `reuse-before-writing.md` nombra las cuatro condiciones de una base no sana (workaround; rompe `CODING_STANDARDS.md`, el lint o una convención documentada; necesita una bandera o rama que solo activa el código nuevo; lógica duplicada que diverge), prohíbe copiarla o extenderla y da la ruta: sanearla primero con el comportamiento de sus otros llamadores fijado, o enfoque del plan, o reporte antes de construir encima según `defect-fix` paso 4 · false today because `rg --count-matches 'CODING_STANDARDS|boolean flag|diverging' content/skills/hive-core/references/reuse-before-writing.md` no da resultados (sale con 1) · Observed in: `~/.agents/skills/hive-core/references/reuse-before-writing.md` tras `hive update`
- AC2 — El rol instalado `hive-review-code` revisa que el código nuevo no copie ni extienda una base que la hoja no cuenta como sana, con la propagación de esa base como escenario de fallo · false today because `rg -n 'copy or extend' content/agents/review/hive-review-code.md` no da resultados (solo dice «does not copy a workaround») · Observed in: el archivo del rol instalado para Claude Code tras `hive update`
- AC3 — La página del sitio del flujo build dice que reutiliza solo código sano y que sanea primero o pregunta · false today because `rg -n 'reuses only code that is sound' site/src/content/docs/flows/build.md` no da resultados · Observed in: `site/src/content/docs/flows/build.md`
- AC4 — `reading_footprint_test.go` declara la asignación `reuseSoundBaseAllowance` que aceptó D1 A, con su comentario, y la suite de contenido pasa con el paquete de `flow-build` en 22.500 B o menos · false today because `rg -n reuseSoundBaseAllowance tests/content/reading_footprint_test.go` no da resultados · Observed in: `go test ./tests/content/ -count=1` en la rama del build

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- `reuse-before-writing.md` mide 1.591 B; la frase del peldaño 2 que se reemplaza mide 270 B; el borrador revisado mide 1.076 B (+806 B) · medido 2026-11-21.
- Paquete de etapa de `flow-build` (`content/skills/flow-build/assets/stage.json`): 21.610 B de 22.500 B (`tests/content/stage_manifest_test.go:17`) · 2026-11-21.
- Unión plan/build sin git-commit ni merge-gate, medida en `820a0300`: 193.884 B frente a 184.489 B + 9.412 B de asignación = 193.901 B; margen 17 B (`tests/content/reading_footprint_test.go:241-256`, salida `-v`). La rama `feat/close-cleans-at-completion` cambia 8 archivos de esa unión (unos +238 B), así que en `development` el margen ronda 255 B; la asignación se mide en la rama del build (T1) · 2026-11-21.
- La especificación `workflow-lifecycle` «Defects are fixed at their cause» contiene la frase de no copiar un workaround, su revisión y el escenario «Copying an existing workaround»; su última frase sitúa la regla en «a reference of `flow-build`», pero hoy vive en la hoja `hive-core/references/defect-fix.md`.
- Dependientes revisados que solo apuntan a la hoja y no cambian: `global.md:80`, `flow-build/SKILL.md:19`, `flow-plan/SKILL.md:13`, `hive-mode/playbooks/feature.md:10`, `refactor.md:11`, `hive-principle-index/SKILL.md:26`, los roles `hive-build-*`. Sin menciones en `_support/docs/` ni en `tests/pilot/`.
- `go test ./tests/content/ -count=1` · ok (2,1 s) en `820a0300` (`development` `2eef1969` más la rama `feat/close-cleans-at-completion`, que no toca los archivos que este plan edita ni los del paquete de `flow-build`, pero sí 8 de la unión de lectura) · 2026-11-21.
- `rg --count-matches 'CODING_STANDARDS|boolean flag|diverging' content/skills/hive-core/references/reuse-before-writing.md` · sin resultados, sale con 1 · 2026-11-21.
- `rg -n 'copy a workaround|copied workaround' content/` · `hive-review-code.md:12` y `:14` · 2026-11-21.
- `rg -n reuseSoundBaseAllowance tests/content/reading_footprint_test.go` · sin resultados · 2026-11-21.
- `rg -n 'copy or extend' content/agents/review/hive-review-code.md` · sin resultados · 2026-11-21.
- `rg -n 'reuses only code that is sound' site/src/content/docs/flows/build.md` · sin resultados · 2026-11-21.
- `rg -n 'does not copy an existing workaround' site/src/content/docs/flows/build.md` · línea 26 · 2026-11-21.
- `pnpm --dir site check` · 0 errores, 0 avisos, 0 sugerencias (1 s), medido en el checkout principal; en el worktree del build, not measured hasta T2, tras su `install --frozen-lockfile --offline` · 2026-11-21.
- Sesión en paralelo (`feat/close-cleans-at-completion`, último commit `53f4ba24`, sin PR abierto): cambia 62 archivos y ninguno de los que edita este plan; su delta de `workflow-lifecycle` toca otros requisitos; su entrada de CHANGELOG está al inicio de `### Changed` de `[Unreleased]`; cambia 8 archivos de la unión de lectura (cubierto por la medida en los dos órdenes) · 2026-11-21.
- `python3 -I /Users/me/.agents/skills/flow-plan/scripts/check_change.py _support/openspec/changes/reuse-sound-base` · sale con 0 · 2026-11-21.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Modo `ready`: revisión, push, PR no borrador contra `development`, espera del CI y respuesta a sus revisiones; se detiene con el PR abierto para que el usuario haga el merge. Revisión de código: nativa, Claude Code `/code-review`. No autoriza merge, auto-merge, cola de merge, `hive update` ni release.

## Estado de revisión y progreso

2026-11-21 — Plan escrito; revisión de plan (`hive-review-plan`, dominio guía): 3 P1, 4 P2 y 4 P3. H1 ya resuelto por D1 A; aplicados H2–H8 y los P3 (exención del revisor, ruta para construir encima, ruta de build con plan e hijo, excepción al criterio a, frase del sitio, medida en los dos órdenes de integración, comandos `--count-matches`, AC4 falsable).
2026-11-21 — Primera relectura: un P2 (orden de integración) y dos P3 (línea de contexto, hijo fuera de su ruta), aplicados.
2026-11-21 — Revisión del conflicto con la sesión en paralelo, a pedido del usuario: sin archivos en común; se añaden la entrada de CHANGELOG (que CONTRIBUTING exige en cada PR y faltaba), `pnpm check` y el worktree.
2026-11-21 — Tercera relectura: un P2 (el worktree no tiene `site/node_modules`); T2 instala los paquetes fijados sin red antes de `pnpm check`. Sin otros huecos.
2026-11-21 — Segunda relectura: un P3 (worktree temporal en `.claude/worktrees/`, tras el commit de T1, y quitado tras medir), aplicado. Sin otros huecos.
