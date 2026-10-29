# Tareas

Base del cambio: `77076052` (`origin/development`, 2026-10-29).

### T1 — Dividir la referencia por responsabilidad

- [x] «Runtime and UI gates» y «Human UI handoff» viven en `runtime-gates.md`, y cada enlace apunta a su sección nueva.

**Closes:** AC1.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-verify/references/review-and-gates.md` (incluido su encabezado, línea 3), `content/skills/flow-verify/references/runtime-gates.md` (nuevo), los 6 archivos que enlazan `#runtime-and-ui-gates` (los que lista el contexto verificado de `design.md`), `content/skills/flow-verify/SKILL.md:14` y `:27` (enlaces sin ancla a lo que se mueve), `tests/content/reading_footprint_test.go`, `_support/docs/architecture/instruction-resources.md:88`.

**Execution:** hilo principal: edición mecánica acoplada a T2 en el mismo archivo.

**Test approach:** check: `go test ./tests/content/ ./tooling/management/ -run 'InstructionAnchors|Footprint|Size'`.

**Changes:** ver la decisión S10 de `design.md`. `flow-verify/SKILL.md:14` separa el puntero: la revisión dedicada y el push a `review-and-gates.md`, y las comprobaciones de runtime y la entrega de UI a `runtime-gates.md`; `:27` apunta a `runtime-gates.md`. Añadir `runtime-gates.md` al escenario del ciclo y al mapa `moved` de la prueba de huella, y nombrarlo en `instruction-resources.md:88`.

**Verification:**
- `wc -c content/skills/flow-verify/references/review-and-gates.md` da 13 000 bytes o menos.
- `rg -n "review-and-gates.md#(runtime-and-ui-gates|human-ui-handoff)" content site _support/docs` no devuelve nada.
- `rg -n "the local runtime checks, the human UI handoff" content` no devuelve nada (ni el puntero de `flow-verify/SKILL.md:14` ni el encabezado de `review-and-gates.md`), y `rg -n "the rule in \[review and gates\]" content/skills/flow-verify/SKILL.md` no devuelve nada.
- `go test ./tests/content/ ./tooling/management/ -run 'InstructionAnchors|Footprint|Size'` pasa.

### T2 — Cerrar los huecos de las reglas de revisión

- [x] Las reglas de revisión de código, de seguridad y de promoción cubren S2, S4, S5, S8 y R6.

**Closes:** AC2, AC3, AC4, AC5, AC6.

**Checkpoint:** C1.

**Depends on:** T1.

**Locations:** `content/skills/flow-verify/references/review-and-gates.md` («Dedicated code review», «Security review findings»), `content/skills/flow-plan/references/delivery-decisions.md` (regla de promoción), `content/skills/flow-research/references/pr-review.md:15`, `content/skills/flow-plan/references/review-decisions.md:12`, `site/src/content/docs/flows/build.md:47`, `verify.md:30`, `security.md:29`. `content/skills/flow-security/SKILL.md:18` y `site/src/content/docs/flows/security.md:25` quedan sin cambios: S8 los respeta (ver `design.md`).

**Execution:** hilo principal: redacción acoplada y corta en archivos que comparte con T1.

**Test approach:** check: observación del texto y `go test ./tests/content/ ./tooling/management/ -run 'InstructionAnchors|Size'`.

**Changes:** las decisiones S2, S4 (D6-A), S5, S8 y R6 (D7-A) de `design.md`.

**Verification:**
- AC2: la viñeta «Later edits» excluye «a validation correction», la remite a «Dedicated review timing» y trata la revisión de correcciones como la última; `site/src/content/docs/flows/build.md:47` dice lo mismo.
- AC3: «Later edits» nombra el mensaje del commit. La regla de promoción de `delivery-decisions.md` sigue exigiendo que el registro de cambio o el cuerpo del PR registre un commit como revisado, y añade los commits que un mensaje marca sin revisar.
- AC4: «Severity» pide nombrar el hallazgo y su arreglo, o la edición, que abre la segunda revisión, o que ninguno la abre.
- AC5: «Security review findings» dice que una edición posterior a su segunda revisión que no cruza un límite no abre otra y se reporta sin revisar, y remite a `flow-security` la que cruza uno.
- AC6: `rg -n -i "verifies its (own )?findings|checks its own findings|does not verify its own findings|reproduced or re-read|or verified by a mechanism|after refutation, or verified" content site` no devuelve nada.
- Todos: `go test ./tests/content/ ./tooling/management/ -run 'InstructionAnchors|Size'` pasa.

## Verificación compartida

- Suite completa local `go test -count=1 ./...`, una vez sobre el candidato final (el repo no tiene CI).
- `/code-review` de Claude Code sobre el candidato local antes del push; sus P0/P1 pasan por `hive-refute-claim`.
- In-vivo local: no aplica. El cambio es redacción de guía y los pilotos de comportamiento están pausados; el reporte lo dice en una línea.

Checkpoints:
- C1 — T1, T2 — T1 no cumple el criterio de riesgo (solo T2 depende de ella y no cambia un contrato), así que se agrupa con T2, la última tarea `check`.

## Review status and progress

Revisión del plan, ronda 1 (consistencia de la guía, `hive-review-plan`, revisión `bf4c58170624`, unos 131k tokens y 207 s): 4 hallazgos bloqueantes y 3 menores, todos aceptados.
- B1: S8 contradecía a `flow-security/SKILL.md:18` y `site/.../security.md:25`. S8 se ajustó para respetar la cobertura de los límites que aparecen más tarde (cambia AC5).
- B2: AC3 debilitaba el registro positivo de la regla de promoción. Se corrigió la dirección (cambia AC3).
- B3: las comprobaciones de T1 no cubrían las anclas ni los enlaces sin ancla. Se añadieron `TestRepositoryInstructionAnchorsResolve`, las líneas `flow-verify/SKILL.md:14` y `:27` y un `rg`; se corrigió la lista de archivos.
- B4: el patrón de AC6 coincidía con `plan-execution.md:30`. Se acotó.
- N5: S2 deja claro que la revisión de correcciones es la última.
- N6: se añadió `review-decisions.md:12` a T2.
- N7: los enlaces del sitio apuntan a `master`. El enlace nuevo dará 404 hasta la promoción, que queda fuera de alcance; se reporta.

Re-revisión (única, mismo revisor reanudado, revisión `a60ded2bc978`, unos 141k tokens y 43 s): las siete correcciones se sostienen. Quedó 1 bloqueante: el `rg` de T1 para enlaces sin ancla daba falso positivo. Se aplicó el patrón que propuso el revisor (la redacción del puntero viejo), conciliado sin otra ronda. Recordatorios para T2: la viñeta de seguridad (`review-and-gates.md:45`) nombra la excepción de S8, y `site/.../build.md:47` refleja S2 (añadido a la verificación de AC2).

Build (ruta D9-A, 2026-10-29). Implementador: hilo principal (Opus 5.5, observado). Verificadores: `hive-verify-task` (modelo configurado, no observado).
- T1: commits `24f4d85b` y `26d709b8` (etiqueta del enlace del sitio, que el verificador encontró desactualizada). AC1 cumplido.
- T2: commit `21ca03f7`, con AC2–AC6 cumplidos. Después, dos ajustes de redacción en «Later edits» que propuso el verificador: `d0ffa9d6` (la revisión de correcciones es la última en todo caso; AC2 y AC3 se re-verificaron y se cumplen) y `75b553ae` (la regla de registro también aplica tras esa última revisión), cubierto por `/code-review`.
- Suite completa local: pasa en `21ca03f7`; las ediciones posteriores son solo de texto, cubiertas por los tests de contenido.
- `/code-review` sobre `75b553ae`: 8 hallazgos. Un P1 confirmado por el refutador y cinco P2 arreglados en `2b604f93`; un P3 (enlace del sitio a `master`) solo reportado. Re-revisión de `75b553ae..2b604f93`: 10 hallazgos. R1 refutado; R2 (plausible) y R8 (confirmado) arreglados en `fff75a26` sin otra revisión. Quedan sin arreglar los P2 R5, R6 y R7, propuestos al cierre, y los P3 R3, R4, R9 y R10, solo reportados.
- Integrado por PR #292 (merge `83008381`) y desplegado.
