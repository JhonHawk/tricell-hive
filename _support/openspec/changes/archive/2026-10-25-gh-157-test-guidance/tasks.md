# Tareas

Base: `e438e14b` (2026-10-25). Rama `feat/gh-157-test-guidance`, worktree `.claude/worktrees/gh-157` (`automatic`, E3-A). Build iniciado el 2026-10-25.

**Convención de las comprobaciones.** Cada cláusula de un AC tiene su frase, contada con `rg -i -F -c "<frase>" <archivo>`, que debe dar 1 después del cambio; además, el verificador lee cada cláusula contra el texto. Frases fijadas, todas con 0 en `content/`, `site/src/` y `_support/docs/` el 2026-10-25: «same inputs» (AC1), «independent oracle» y «property-based» (AC2), «already owns» (AC3), «test-only entry point» (AC4). Las demás cláusulas se fijan al escribirlas con la misma comprobación de ausencia.

## T1 — Elegir el tipo de prueba y filtrar antes de escribirla

- [x] `tdd.md` gana una sección «Choose the kind of test» con propiedades (AC2), el filtro de dueño y patrones (AC3) y la prohibición de puntos de entrada solo para pruebas (AC4); `hive-write-tests` y `hive-review-code` remiten a ella (AC5).
Verificado por `hive-verify-task`: AC2–AC5 cumplidos (`tdd.md:11-17`, roles en la línea 12–13); `go test ./tests/content/...` pasa. Observación sin bloqueo: `hive-review-code.md:12` conserva «duplicates a test which already covers the behavior», redundante con el filtro.

**Closes:** AC2, AC3, AC4, AC5.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-build/references/tdd.md` (nueva sección tras «Choose and run the cycle»; descripción de la línea 5); `content/agents/quality/hive-write-tests.md:13`; `content/agents/review/hive-review-code.md:12`; `content/skills/flow-build/SKILL.md:42`.
**Execution:** hilo principal: el texto depende de las decisiones del plan y un brief sería más largo que el cambio.
**Test approach:** check: conteo de frases por cláusula, lectura y `go test ./tests/content/...`.
**Changes:** ver [design.md](design.md) D-1 y D-2.
**Verification:** cada frase de AC2–AC5 da 1 en su archivo; «name the realistic change» da 0 en `hive-write-tests.md` y 1 en `tdd.md`; `rg -c "tdd.md"` da ≥ 1 en ambos roles; lectura de cada cláusula; `go test ./tests/content/...` pasa.

## T2 — Comparación viejo contra nuevo al reemplazar o migrar

- [x] `verification.md` gana la fila «Replacement or migration» (AC1) y el párrafo de reemplazo de `flow-plan/SKILL.md` remite a ella (AC6).
Verificado por `hive-verify-task`: AC1 y AC6 cumplidos (`verification.md:20`, `flow-plan/SKILL.md:26`); `go test ./tests/content/...` pasa. Commit en la rama.

**Closes:** AC1, AC6.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-build/references/verification.md` tabla «Changed surface» (línea 13); `content/skills/flow-plan/SKILL.md:26`.
**Execution:** hilo principal, mismo motivo que T1.
**Test approach:** check: conteo de frases, `go test ./tests/content/...` (tamaño de `flow-plan/SKILL.md` ≤ 16 000 bytes).
**Changes:** ver [design.md](design.md) D-3.
**Verification:** frases de AC1 y AC6 dan 1; `go test ./tests/content/...` pasa.

## T3 — Página del sitio

- [x] `build.md` describe T1 y T2, y `plan.md` dice que el plan de un reemplazo incluye la comparación, ambas enlazando la skill.
Verificado por `hive-verify-task`: AC7 cumplido (`build.md:36-37`, `plan.md:32`). El orquestador corrigió la gramática de `build.md:36`, una observación del verificador. `astro check` no corrió: necesita una instalación no autorizada.

**Closes:** AC7.
**Depends on:** T1, T2.
**Locations:** `site/src/content/docs/flows/build.md` líneas 33–55; `site/src/content/docs/flows/plan.md:30`.
**Execution:** delegado a `hive-write-docs`: fuentes y destino fijos.
**Test approach:** check: las páginas nombran las prácticas y no copian reglas.
**Verification:** en `build.md`, tres conteos separados ≥ 1: «property», «same inputs» u otra frase de la comparación, y una frase del filtro; en `plan.md`, una frase de la comparación ≥ 1; lectura del verificador.

## Verificación compartida

| Comprobación | Cuándo | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| `go vet ./...` y `go test ./...` | candidato final | salida limpia | rutina de implementación |
| Revisión de código | según E4 | según E4 | según E4 |
| Verificación por tarea | tras cada tarea | `hive-verify-task` | rutina de implementación |

Sin efecto en UI ni en tiempo de ejecución: no aplica recorrido en vivo.

## Revisión y progreso

Revisión del plan (2026-10-25, un `hive-review-plan` del dominio «guía distribuida», ~167 s, ~131k tokens): sin bloqueantes; aplicados S1 (permitir la inyección en el límite del sistema), S2 (un solo hogar para el filtro y la lista de patrones, incluido `hive-review-code`), S3 (una frase por cláusula y conteos separados), S4 (`plan.md` del sitio) y los opcionales O1–O3. Queda abierta una pregunta del revisor: si los briefs de los implementadores `hive-build-*` les hacen leer `tdd.md`. `flow-build/SKILL.md:42` lo exige antes de tocar código o pruebas, y la regla de delegación pide pasar esa referencia en el brief.

Siguiente paso: `flow-build` (R1-A, 2026-10-25: AC1–AC7 aprobados).
