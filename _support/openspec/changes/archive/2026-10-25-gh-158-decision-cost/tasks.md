# Tareas

Base: `4001d2e0` (merge de #157, 2026-10-25) para T2 y T3; T1 y T4 esperan el merge de gh-159, que se integrará en la rama. Worktree `.claude/worktrees/gh-158`. Rama propuesta `feat/gh-158-decision-cost`.

**Convención de las comprobaciones.** Igual que en [gh-157](../gh-157-test-guidance/tasks.md): una frase por cláusula, verificada en 0 antes de escribirla, más lectura del verificador.

## T1 — Tamaño estimado en las opciones

- [x] Las Actions de `questions.md:33` llevan la regla de AC1; la pregunta de cierre de `flow-close/SKILL.md` aplica AC2 remitiendo a ella.
Verificado por `hive-verify-task`: AC1 y AC2 cumplidos (`questions.md:37-39`, `flow-close/SKILL.md:40`). El orquestador ajustó el puntero, siguiendo dos observaciones del verificador: ahora apunta a «the size condition» y aclara que el tamaño va una vez por pregunta. La regla quedó como condición propia junto a la de costo, no dentro de sus Actions.

**Closes:** AC1, AC2.
**Depends on:** merge de gh-159 (espacio en `flow-close/SKILL.md` y mismas líneas de la pregunta de cierre).
**Locations:** `content/skills/flow-plan/references/questions.md:33`; `content/skills/flow-close/SKILL.md`, sección «Close question».
**Execution:** hilo principal: el texto depende de decisiones del plan y un brief sería más largo que el cambio.
**Test approach:** check: conteo por cláusula, lectura, tamaño de `flow-close/SKILL.md` ≤ 16 000 bytes.
**Verification:** «not investigated» y la frase de los disparadores dan 1 en `questions.md`; la frase de tamaño da 1 en `flow-close/SKILL.md`; lectura de cada cláusula; `go test ./tests/content/...` pasa.

## T2 — Revisor de consistencia de la guía

- [x] `plan-review.md` agrega la asignación «Guidance consistency» de AC3 a la tabla de dominios y al párrafo de cobertura.
Verificado por `hive-verify-task`: AC3 cumplido (`plan-review.md:9,18`).

**Closes:** AC3.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/plan-review.md` (párrafo «Choose the useful coverage» y tabla de dominios). El contrato de `hive-review-plan` no enumera dominios y no cambia.
**Execution:** hilo principal.
**Test approach:** check: una frase en la tabla, otra en el párrafo de cobertura y otra para la cláusula de la decisión registrada.
**Verification:** las tres frases dan 1; `go test ./tests/content/...` pasa.

## T3 — Declarar entrega y revisión

- [x] `AGENTS.md` `## Hive` declara los dos valores y `AGENTS.md:57` reconoce el valor declarado como instrucción explícita.
Verificado por `hive-verify-task`: AC4 cumplido (`AGENTS.md:14-15,59`); `hive doctor` da el proyecto como válido.

**Closes:** AC4.
**Depends on:** ninguna.
**Locations:** `AGENTS.md:9-13`; `AGENTS.md:57`.
**Execution:** hilo principal.
**Test approach:** check.
**Verification:** `rg -c "^- (Delivery: automatic|Review: /code-review)$" AGENTS.md` da 2; la frase nueva de la línea 57 da 1; `rg -n "explicit instruction" AGENTS.md content/guidance/global.md` no deja otra frase que contradiga el valor declarado; `hive doctor` no reporta error en `## Hive` (`tooling/cli/doctor_project.go:355`).

## T4 — Páginas del sitio

- [x] `plan.md` y `close.md` describen AC1–AC3 enlazando las skills.
Verificado por `hive-verify-task`: AC5 cumplido (`plan.md:35,68`, `close.md:49`). El orquestador completó la descripción del revisor de consistencia, una omisión que señaló el verificador.

**Closes:** AC5.
**Depends on:** T1, T2.
**Locations:** `site/src/content/docs/flows/plan.md` (sección de preguntas, línea 66; revisión del plan, línea 34); `site/src/content/docs/flows/close.md` (paso 7, línea 49).
**Execution:** delegado a `hive-write-docs`: fuentes y destino fijos.
**Test approach:** check: conteo por página y lectura.
**Verification:** `plan.md` menciona el tamaño estimado y la revisión de consistencia (dos conteos ≥ 1); `close.md` menciona el tamaño del ticket recomendado (≥ 1).

## Verificación compartida

| Comprobación | Cuándo | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| `go vet ./...` y `go test ./...` | candidato final | salida limpia | rutina de implementación |
| `/code-review` | candidato antes del merge | hallazgos verificados por la herramienta | `Review: /code-review` (E4-A) |
| Verificación por tarea | tras cada tarea | `hive-verify-task` | rutina de implementación |
| Lugares revisados y tamaño de `global.md` en el PR | al abrir el PR | lista de `AGENTS.md:24` | regla del repositorio |

## Revisión y progreso

Revisión del plan (2026-10-25, un `hive-review-plan`, ~148 s, ~116k tokens): un bloqueante, B1 (`AGENTS.md:57` contradecía declarar `Delivery`), resuelto en AC4 y T3. Aplicados S2–S7 y O1–O4. S1 aplicado fuera del repositorio: la memoria del host `feedback_test_branch_workflow` quedó marcada como histórica, con un puntero a Engram.

E9-A y R1-A (2026-10-25): AC1–AC5 aprobados; build autorizado después del merge de gh-159.
