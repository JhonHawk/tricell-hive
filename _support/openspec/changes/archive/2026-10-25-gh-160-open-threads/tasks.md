# Tareas

Base: `27fab1ff` (merge de #159, 2026-10-25). Rama `feat/gh-160-open-threads`, worktree `.claude/worktrees/gh-160` (`automatic`). `global.md` en la base: 36 990 bytes.

**Convención de las comprobaciones.** Igual que en [gh-157](../gh-157-test-guidance/tasks.md): una frase por cláusula, verificada en 0 antes de escribirla, más lectura del verificador.

## T1 — La regla en la guía global

- [x] `global.md`, sección «Communication», contiene la regla completa (AC1–AC4), con el alcance de [design.md](design.md); `flow-plan/SKILL.md:16` remite a ella (AC5) y `flow-close/SKILL.md` cuenta un tema abierto sin destino como pendiente.
Verificado por `hive-verify-task`: AC1–AC5 cumplidos (`global.md:21`, `flow-plan/SKILL.md:16`, `flow-close/SKILL.md:36`).
`global.md`: 36 990 → 37 775 bytes (+785). El orquestador recortó del borrador de +881 bytes dos cláusulas redundantes: «with no skill loaded», que es implícita en la guía global, y «Pending at close», que ya vive en `flow-close`.

**Closes:** AC1, AC2, AC3, AC4, AC5.
**Depends on:** ninguna.
**Locations:** `content/guidance/global.md` (sección «Communication»); `content/skills/flow-plan/SKILL.md:16`; `content/skills/flow-close/SKILL.md:36` (excepción de la pregunta de cierre); `content/skills/flow-close/references/reporting.md` (categoría Pending) solo si necesita el puntero.
**Execution:** hilo principal: texto acoplado al diseño y al presupuesto de bytes.
**Test approach:** check: conteo de frases por cláusula, lectura y `go test ./tests/content/...`.
**Verification:** «open topics» da 1 en `content/guidance/global.md` (solo ese archivo); las frases de AC2–AC4 dan 1 en `global.md`; «material expansion beyond that outcome» da 0 en `flow-plan/SKILL.md` y el puntero da 1; `flow-close/SKILL.md` ≤ 16 000 bytes (coordinar con gh-159, que libera espacio); `wc -c content/guidance/global.md` antes y después anotado (techo ~700 bytes); `go test ./tests/content/...` pasa.

## T2 — Páginas del sitio

- [x] `research.md`, `plan.md` y `build.md` describen el comportamiento.
Verificado por `hive-verify-task`: AC6 cumplido (`research.md:51`, `plan.md:70`, `build.md:42`). El orquestador cambió «reports» por «findings reports» en dos páginas, una divergencia que señaló el verificador.

**Closes:** AC6.
**Depends on:** T1.
**Locations:** `site/src/content/docs/flows/research.md`, `plan.md`, `build.md`.
**Execution:** delegado a `hive-write-docs`: fuentes y destino fijos.
**Test approach:** check: lectura y conteo.
**Verification:** cada página menciona la lista de temas abiertos (conteo ≥ 1 por página) y enlaza la guía.

## Verificación compartida

| Comprobación | Cuándo | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| `go vet ./...` y `go test ./...` | candidato final | salida limpia | rutina de implementación |
| `/code-review` | candidato antes del merge | hallazgos verificados por la herramienta | `Review: /code-review` (E4-A) |
| Verificación por tarea | tras cada tarea | `hive-verify-task` | rutina de implementación |
| Lugares revisados en el PR | al abrir el PR | lista de AGENTS.md («Guidance structure») y tamaño de `global.md` antes y después | regla del repositorio |

Lugares que el PR lista (revisión del plan): `global.md`, `flow-plan/SKILL.md`, `plan-format.md:13`, `plan-revision.md:5`, `flow-research/SKILL.md:42`, `flow-close/SKILL.md`, `reporting.md`, `flow-build/SKILL.md`, `content/agents/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`.

## Revisión y progreso

Revisión del plan (2026-10-25, un `hive-review-plan`, ~135 s, ~106k tokens): sin bloqueantes. Aplicados S1–S6 (E8-A fijado, alcance explícito, sin preguntas nuevas, traspasos que escribe el agente, tema sin destino como pendiente, lugares revisados en el PR) y O1–O3. El presupuesto sube de ~500 a ~700 bytes (borrador del revisor: ~680), desviación que se avisa al usuario y se declara en el PR.

Siguiente paso: `flow-build` (R1-A, 2026-10-25: AC1–AC6 aprobados) después del merge de gh-159.
