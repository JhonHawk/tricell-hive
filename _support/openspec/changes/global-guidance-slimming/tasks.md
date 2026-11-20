# Tasks

Base: tricell-hive `development` @ 27875867; rama `refactor/global-guidance-slimming` en un worktree bajo `.claude/worktrees/`. T1 recolecta con un hijo `hive-research` (solo lectura) y el hilo principal escribe sus informes; T2 a T6 corren en el hilo principal porque todas escriben `content/guidance/global.md` y deben mantener su coherencia; T7 empieza en el hilo principal y espera las sesiones del usuario. Hasta T6, las tareas verifican con `go test ./tests/content/... -skip TestGlobalGuidanceSize`, porque el test de presupuesto falla a propósito hasta que el global baje de 22.000 bytes.

### T1 — Línea base del auditor sobre sesiones previas

- [ ] `_support/workspace/2026-11-20-global-slimming-trial/baseline.md` (ignorado por Git) registra, por sesión real hecha con el global actual en esta máquina, host, modelo, commit de contenido instalado y el estado de cada criterio de `--audit-session`; `trial.md` en esta carpeta lleva solo los totales por host y criterio (sesiones observadas, fallos, `review`, `not_observed`).

**Closes:** AC7

**Depends on:** none

**Locations:** `_support/workspace/2026-11-20-global-slimming-trial/baseline.md`, `_support/openspec/changes/global-guidance-slimming/trial.md`

**Execution:** un hijo `hive-research` corre `go run ./tests/pilot --audit-session` y devuelve las filas; el hilo principal escribe los dos archivos

**Test approach:** check

**Changes:** No toca guía ni tests. Excluye las sesiones de este repositorio que editaban la guía misma y las de modelos fuera del piso de `_support/docs/harness-engineering/model-floor.md`. `trial.md` no nombra rutas ni IDs de otros proyectos.

**Verification:** `trial.md` tiene, para cada host, el número de sesiones auditadas (al menos 5, o cuántas había); `git check-ignore` confirma que la carpeta de `_support/workspace/` está ignorada.

### T2 — Test de presupuesto y regla de destino

- [ ] Una función que recibe la raíz del repositorio comprueba que el global mide 22.000 bytes o menos y, si no, devuelve un mensaje que nombra la regla de destino; `TestGlobalGuidanceSize` la usa y un caso con un global de 22.001 bytes en una raíz temporal falla; el bullet de `AGENTS.md` «Guidance structure» que ya dice «restate a rule in global.md only when it must act without its home being read» se amplía con la regla de destino y el tope.

**Closes:** AC2

**Depends on:** none

**Locations:** `tests/content/budget_test.go`, `AGENTS.md`, `_support/openspec/changes/global-guidance-slimming/specs/context-delivery/spec.md`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** La función sigue el patrón de `leafProblems` en `leaf_skill_test.go` (raíz como parámetro, caso con fixture). No cambia ningún otro test de presupuesto. `AGENTS.md` no gana un bullet nuevo: se edita el existente.

**Verification:** el caso de 22.001 bytes falla con el mensaje esperado; `go test ./tests/content/ -run TestGlobalGuidanceSize` falla contra el global actual (32.148) y pasa al terminar T6.

### T3 — Mover procedimiento a sus hojas

- [ ] Las filas del «Mapa de reglas» de `design.md` con destino en una hoja o referencia (#7 → `hive-rule-checks`, #11 → `hive-rule-trackers`, #12 → `hive-rule-reasons-and-blockers`, #14 → `hive-rule-ticket-state`, #29 → `hive-rule-local-write/references/worktree.md`, #45 → `hive-rule-checkpoint`) tienen su texto allí una sola vez; el global conserva para cada una la línea que el mapa le asigna y el enlace; cada hoja sigue bajo su límite de `leaf_skill_test.go`.

**Closes:** AC3

**Depends on:** T2

**Locations:** `content/guidance/global.md`, `content/skills/hive-rule-checks/`, `hive-rule-trackers/`, `hive-rule-reasons-and-blockers/`, `hive-rule-ticket-state/`, `hive-rule-local-write/` (y su nueva `references/worktree.md`), `hive-rule-checkpoint/`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Mueve texto sin cambiar su significado. Lo que el mapa deja en el global (definiciones transversales, proyecto externo, prohibiciones de limpieza, las dos cláusulas del hook de compactación) solo se compacta. Si una hoja destino no cabe, la tarea se detiene y propone una referencia con su condición de lectura en vez de exceder el límite.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa; por cada fila movida, `rg` de su frase clave devuelve una sola coincidencia bajo `content/`, en su destino.

### T4 — Quitar las copias de reglas que ya viven en su hoja

- [ ] Las filas #37–42 y #44 del mapa quedan como línea mínima o solo entrada del índice; #17 y #23 quedan en una línea que conserva su autorización o prohibición; `single_home_test.go` registra la frase clave de cada una en su hogar.

**Closes:** AC4

**Depends on:** T3

**Locations:** `content/guidance/global.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Solo elimina texto que su hogar ya contiene; antes de borrar cada frase confirma con `rg` que el hogar la tiene completa.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa; `rg` de cada frase clave no la encuentra en el global.

### T5 — Principio de evidencia

- [ ] Existe `content/skills/hive-principle-verify-current-state/SKILL.md` (oculto, ≤ 4.000 bytes, con su aviso de origen si toma material de pstack), `hive-principle-index` lo enlaza con su momento, el global conserva una línea de invariante que lo enlaza y nombra el momento (antes de presentar memoria, un resumen o el comportamiento de una herramienta como actual), y la primera frase de #32 se elimina porque `hive-rule-reuse-before-writing` peldaño 1 ya la dice.

**Closes:** AC5

**Depends on:** T4

**Locations:** `content/skills/hive-principle-verify-current-state/`, `content/skills/hive-principle-index/SKILL.md`, `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** La regla no cambia de significado.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa (hojas ocultas, índice de principios, enlaces).

### T6 — Compactar el índice, reapuntar citas y documentación

- [ ] La sección «Rule leaves» lista las 61 hojas una vez cada una, con líneas que pueden agrupar hojas del mismo momento; el global mide ≤ 22.000 bytes; cada cita a «Scope and authorization», «Proportionality», «Persistent memory» u otra sección del global bajo `content/` apunta a donde ahora está su contenido, con la lista de lugares revisados en el PR; el delta de `context-delivery` está escrito; `llms.txt:32`, `README.md:470–472`, `site/src/content/docs/getting-started.md:88`, `site/src/content/docs/flows/plan.md:31`, `site/src/content/docs/skills.md` y `README.md:36` no afirman nada que el cambio deje falso; `hive update --dry-run` muestra cuántas partes recibe Antigravity.

**Closes:** AC1

**Depends on:** T5

**Locations:** `content/guidance/global.md`, los archivos de `content/` que citan secciones del global, `site/src/content/docs/`, `README.md`, `llms.txt`, `_support/openspec/changes/global-guidance-slimming/specs/context-delivery/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Cada entrada del índice conserva su momento observable. Si el global no baja de 22.000 bytes sin perder intención, la tarea se detiene y pregunta (design.md «Decisions»).

**Verification:** `wc -c content/guidance/global.md` ≤ 22000; `go test ./tests/content/...` pasa, incluidos el test de presupuesto y `reading_footprint_test.go`; `rg` de cada cita revisada.

### T7 — Instalar la rama y evaluar sesiones reales

- [ ] Con la rama instalada en local, `trial.md` compara las sesiones reales del usuario contra la línea base de T1 según AC7 (a) y (b), con cada sesión etiquetada por host, modelo y commit instalado; el usuario decide fusionar, replantear o descartar.

**Closes:** AC7

**Depends on:** T1, T6

**Locations:** `_support/openspec/changes/global-guidance-slimming/trial.md`, `_support/workspace/2026-11-20-global-slimming-trial/`; instalación local de los hosts registrados

**Execution:** hilo principal (instalación y auditoría)

**Test approach:** check

**Changes:** Instala desde el commit de la rama con `hive update --source <worktree> --out FILE` y `hive apply --plan FILE`, nombrando antes los commits que se despliegan; no fusiona. Tras cada maintainer refresh desde `development` durante la prueba, rebasa la rama sobre lo nuevo y la reinstala (D5 A). Al descartar, reinstala desde `development`. Para AC7 (b), lista cada regla movida con su detector del auditor o «revisión manual» y qué buscar.

**Verification:** `trial.md` tiene la tabla de después, la lista de reglas movidas con su resultado, y cada regresión con su sesión en el archivo ignorado.

## Delivery units

### E1 — PR de `refactor/global-guidance-slimming` a `development` · depends on: none · files: `content/`, `tests/content/`, `AGENTS.md`, `site/`, `README.md`, `llms.txt`, esta carpeta

- [ ] unit: `go test ./tests/content/... ./tests/pilot/...`
- [ ] live: sesiones reales con la rama instalada · pass: `trial.md` sin regresiones según AC7, o el usuario acepta las que haya
- [ ] perf: n/a: el cambio no tiene métrica de rendimiento

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | `/code-review` de Claude Code | The review answer |
| Prueba en sesiones reales | Tras instalar la rama, en esta máquina | `trial.md` | D3 A |

La revisión humana es la decisión de fusión tras leer `trial.md`.

## Checkpoint

- Updated: 2026-11-20
- Stage: build · start
- Last commit: 27875867
- Open findings: none
- Next: T1 baseline (child) and T2 budget test
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/global-guidance-slimming · refactor/global-guidance-slimming

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-20 — plan escrito.
2026-11-20 — revisión del plan (hive-review-plan): 8 P1, 6 P2, 3 P3. Aplicadas: test de presupuesto omitido hasta T6; hojas en su límite (destinos cambiados o referencia nueva); 38 citas a reapuntar en T6; AC1 sobre el archivo fuente; filas por sesión en `_support/workspace/` (repo público); AC7 con denominador, estados, modelo, commit y lista de reglas movidas; cláusulas del hook de compactación y definiciones transversales en el global; #17 y #23 con su línea; AC6 retirado; documentación completa; T2 con función y fixture; el hijo de T1 no escribe. D5 A respondida.
