# Tasks

Base: tricell-hive `development` @ 27875867; rama `refactor/global-guidance-slimming` en un worktree bajo `.claude/worktrees/`. T1 recolecta con un hijo `hive-research` (solo lectura) y el hilo principal escribe sus informes; T2 a T6 corren en el hilo principal porque todas escriben `content/guidance/global.md` y deben mantener su coherencia; T7 empieza en el hilo principal y espera las sesiones del usuario. Hasta T6, las tareas verifican con `go test ./tests/content/... -skip TestGlobalGuidanceSize`, porque el test de presupuesto falla a propósito hasta T6.

### T1 — Línea base del auditor sobre sesiones previas

- [x] `_support/workspace/2026-11-20-global-slimming-trial/baseline.md` (ignorado por Git) registra, por sesión real hecha con el global actual en esta máquina, host, modelo, commit de contenido instalado y el estado de cada criterio de `--audit-session`; `trial.md` en esta carpeta lleva solo los totales por host y criterio (sesiones observadas, fallos, `review`, `not_observed`).

**Closes:** AC7

**Depends on:** none

**Locations:** `_support/workspace/2026-11-20-global-slimming-trial/baseline.md`, `_support/openspec/changes/global-guidance-slimming/trial.md`

**Execution:** un hijo `hive-research` corre `go run ./tests/pilot --audit-session` y devuelve las filas; el hilo principal escribe los dos archivos

**Test approach:** check

**Changes:** No toca guía ni tests. Excluye las sesiones de este repositorio que editaban la guía misma y las de modelos fuera del piso de `_support/docs/harness-engineering/model-floor.md`. `trial.md` no nombra rutas ni IDs de otros proyectos.

**Verification:** `trial.md` tiene, para cada host, el número de sesiones auditadas (al menos 5, o cuántas había); `git check-ignore` confirma que la carpeta de `_support/workspace/` está ignorada.

### T2 — Test de presupuesto y regla de destino

- [x] Una función que recibe la raíz del repositorio comprueba que el global mide 30.000 bytes o menos (D8; antes 28.000, 27.500 y 22.000) y, si no, devuelve un mensaje que nombra la regla de destino; `TestGlobalGuidanceSize` la usa y un caso con un global de 30.001 bytes en una raíz temporal falla; el bullet de `AGENTS.md` «Guidance structure» que ya dice «restate a rule in global.md only when it must act without its home being read» se amplía con la regla de destino y el tope.

**Closes:** AC13

**Depends on:** none

**Locations:** `tests/content/budget_test.go`, `AGENTS.md`, `_support/openspec/changes/global-guidance-slimming/specs/context-delivery/spec.md`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** La función sigue el patrón de `leafProblems` en `leaf_skill_test.go` (raíz como parámetro, caso con fixture). No cambia ningún otro test de presupuesto. `AGENTS.md` no gana un bullet nuevo: se edita el existente.

**Verification:** el caso de 30.001 bytes falla con el mensaje esperado; `go test ./tests/content/ -run TestGlobalGuidanceSize` falla contra el global actual (32.148) y pasa al terminar T6.

### T3 — Mover procedimiento a sus hojas

- [x] Las filas del «Mapa de reglas» de `design.md` con destino en una hoja o referencia (#7 → `hive-rule-checks`, #11 → `hive-rule-trackers`, #12 → `hive-rule-reasons-and-blockers`, #14 → `hive-rule-ticket-state`, #29 → `hive-rule-local-write/references/worktree.md`, #45 → `hive-rule-checkpoint`) tienen su texto allí una sola vez; el global conserva para cada una la línea que el mapa le asigna y el enlace; cada hoja sigue bajo su límite de `leaf_skill_test.go`.

**Closes:** AC3

**Depends on:** T2

**Locations:** `content/guidance/global.md`, `content/skills/hive-rule-checks/`, `hive-rule-trackers/`, `hive-rule-reasons-and-blockers/`, `hive-rule-ticket-state/`, `hive-rule-local-write/` (y su nueva `references/worktree.md`), `hive-rule-checkpoint/`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Mueve texto sin cambiar su significado. Lo que el mapa deja en el global (definiciones transversales, proyecto externo, prohibiciones de limpieza, las dos cláusulas del hook de compactación) solo se compacta. Si una hoja destino no cabe, la tarea se detiene y propone una referencia con su condición de lectura en vez de exceder el límite.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa; por cada fila movida, `rg` de su frase clave devuelve una sola coincidencia bajo `content/`, en su destino.

### T4 — Quitar las copias de reglas que ya viven en su hoja

- [x] Las filas #37–42 y #44 del mapa quedan como línea mínima o solo entrada del índice; #17 y #23 quedan en una línea que conserva su autorización o prohibición; `single_home_test.go` registra la frase clave de cada una en su hogar.

**Closes:** AC4

**Depends on:** T3

**Locations:** `content/guidance/global.md`, `tests/content/single_home_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Solo elimina texto que su hogar ya contiene; antes de borrar cada frase confirma con `rg` que el hogar la tiene completa.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa; `rg` de cada frase clave no la encuentra en el global.

### T5 — Principio de evidencia

- [x] Existe `content/skills/hive-principle-verify-current-state/SKILL.md` (oculto, ≤ 4.000 bytes, con su aviso de origen si toma material de pstack), `hive-principle-index` lo enlaza con su momento, el global conserva una línea de invariante que lo enlaza y nombra el momento (antes de presentar memoria, un resumen o el comportamiento de una herramienta como actual), y la primera frase de #32 se elimina porque `hive-rule-reuse-before-writing` peldaño 1 ya la dice.

**Closes:** AC5

**Depends on:** T4

**Locations:** `content/skills/hive-principle-verify-current-state/`, `content/skills/hive-principle-index/SKILL.md`, `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** La regla no cambia de significado.

**Verification:** `go test ./tests/content/... -skip TestGlobalGuidanceSize` pasa (hojas ocultas, índice de principios, enlaces).

### T6 — Compactar el índice, reapuntar citas y documentación

- [x] La sección «Rule leaves» lista las 61 hojas una vez cada una, con líneas que pueden agrupar hojas del mismo momento; el global mide ≤ 30.000 bytes (D8); cada cita a «Scope and authorization», «Proportionality», «Persistent memory» u otra sección del global bajo `content/` apunta a donde ahora está su contenido, con la lista de lugares revisados en el PR; el delta de `context-delivery` está escrito; `llms.txt:32`, `README.md:470–472`, `site/src/content/docs/getting-started.md:88`, `site/src/content/docs/flows/plan.md:31`, `site/src/content/docs/skills.md` y `README.md:36` no afirman nada que el cambio deje falso; `hive update --dry-run` muestra cuántas partes recibe Antigravity.

**Closes:** AC12

**Depends on:** T5

**Locations:** `content/guidance/global.md`, los archivos de `content/` que citan secciones del global, `site/src/content/docs/`, `README.md`, `llms.txt`, `_support/openspec/changes/global-guidance-slimming/specs/context-delivery/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Cada entrada del índice conserva su momento observable. La parada por tamaño ya se resolvió: tope final 30.000 (D8).

**Verification:** `wc -c content/guidance/global.md` ≤ 30000; `go test ./tests/content/...` pasa, incluidos el test de presupuesto y `reading_footprint_test.go`; `rg` de cada cita revisada.

### T7 — Instalar la rama y evaluar sesiones reales

- [ ] Con la rama instalada en local, `trial.md` compara las sesiones reales del usuario contra la línea base de T1 según AC7 (a) y (b), con cada sesión etiquetada por host, modelo y commit instalado; el usuario decide fusionar, replantear o descartar.

**Closes:** AC7

**Depends on:** T1, T6

**Locations:** `_support/openspec/changes/global-guidance-slimming/trial.md`, `_support/workspace/2026-11-20-global-slimming-trial/`; instalación local de los hosts registrados

**Execution:** hilo principal (instalación y auditoría)

**Test approach:** check

**Changes:** Instala desde el commit de la rama con `hive update --source <worktree> --out FILE` y `hive apply --plan FILE`, nombrando antes los commits que se despliegan; no fusiona. Tras cada rebase, una regla que `development` haya añadido al global se reescribe según la regla de destino, se corre `go test ./tests/content/...` y solo se reinstala si pasa; si no cabe en el tope, se detiene y pregunta. Tras cada maintainer refresh desde `development` durante la prueba, rebasa la rama sobre lo nuevo y la reinstala (D5 A). Al descartar, reinstala desde `development`. Para AC7 (b), lista cada regla movida con su detector del auditor o «revisión manual» y qué buscar.

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
- Stage: paused · trial audited, waiting on the user's merge decision; PR #554 open
- Last commit: 860b73cf (development 059868cd merged in)
- Open findings: none
- Next: D1 A (2026-11-20) — keep the trial running; after more Claude and Codex sessions in other projects, re-audit #29 (HIVE.local.md read at start) and the auditor criteria, then the user decides merge
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/global-guidance-slimming · refactor/global-guidance-slimming

### Evidence

- before captures: none (no screen)
- go test ./... · exit 0 in 239 s · 2e6e3514
- hive apply transaction 0d5d4d42 · branch 2e6e3514 installed on all 7 hosts
- go test ./... · exit 0 in 251 s · 02a3e20a
- hive apply · branch 02a3e20a installed on all hosts
- 14:25 the other session's refresh deployed development 277a139a over the branch; sessions started 14:25–14:30 ran development, not the branch, and do not count for T7
- go test ./... · exit 0 · rebased branch (global 28,593 B)
- hive apply transaction 2b0b51ad · branch a19000d2 installed on all hosts
- go test ./... · exit 0 · 860b73cf (development with #557 merged in)
- hive apply transaction a8a959b7 · branch 860b73cf installed on all hosts
- trial audit 2026-11-20 17:50: 2 Claude + 9 Grok sessions; AC7 (a) no regression; (b) #29 fails in both Claude sessions

## Review status and progress

2026-11-20 — plan escrito.
2026-11-20 — revisión del plan (hive-review-plan): 8 P1, 6 P2, 3 P3. Aplicadas: test de presupuesto omitido hasta T6; hojas en su límite (destinos cambiados o referencia nueva); 38 citas a reapuntar en T6; AC1 sobre el archivo fuente; filas por sesión en `_support/workspace/` (repo público); AC7 con denominador, estados, modelo, commit y lista de reglas movidas; cláusulas del hook de compactación y definiciones transversales en el global; #17 y #23 con su línea; AC6 retirado; documentación completa; T2 con función y fixture; el hijo de T1 no escribe. D5 A respondida.
2026-11-20 — T2 · df7209b8 · unverified: TestGlobalGuidanceSize fails until T6 by design
2026-11-20 — T1 · f88e5697 · unverified: no Claude baseline (all period sessions ran in tricell-hive); gpt-6.2-astra sessions excluded pending floor confirmation
2026-11-20 — T3 · f2f691ee · unverified: tooling release validation not yet run (runs at local CI); instruction-resources.md does not yet list references/worktree.md (T6)
2026-11-20 — T4 · 0eff79dc · unverified: none
2026-11-20 — T5 · 0eff79dc · unverified: first sentence of #32 kept: reuse rung 1 covers only code, not investigation, plan and verification
2026-11-20 — T6 · 4e03441c · unverified: Antigravity installs in 2 parts (source would need ~21,000 B for one); global at 27,494 of 27,500 B
2026-11-20 — D6 A: tope 27.500; AC1 y AC2 retirados por AC8 y AC9. T6 restauró en el global las líneas mínimas de preguntas que T4 había quitado (formato de decisión, arreglo más pequeño de un bloqueo, cierre de un análisis): 27.494 bytes. Segunda revisión del plan: 2 P1, 3 P2, 3 P3 aplicados (cifras de T2 y T6, parada resuelta, objetivo y título, regla de rebase en T7, fila del registro de enforcers).
2026-11-20 — verify: /code-review halló 3 importantes (efectos críticos, tickets y continuaciones ensanchados al compactar) y 6 menores; corregidos en una ronda. D7 A: tope 28.000; AC8 y AC9 retirados por AC10 y AC11.
2026-11-20 — verify: segunda revisión (/code-review) halló 4 importantes más de la misma causa (defecto, momentos de escrituras en tickets, alcance de la ronda de efectos, permiso del canal de preguntas) y 5 menores; restaurado el texto original y la línea de «Delivery». D8: tope 30.000; AC10 y AC11 retirados por AC12 y AC13. Global: 28.468 bytes.
2026-11-20 — rebase sobre development 277a139a (#555, hoja background wait): la línea nueva del índice se agrupó con [checks]; sin solapamiento con «Long checks». Global: 28.593 bytes (base 32.279, −3.686). Suite completa en verde.
