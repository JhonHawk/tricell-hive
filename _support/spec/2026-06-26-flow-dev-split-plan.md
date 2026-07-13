# Flow-dev Split — Implementation Plan

> **Spec:** `_support/spec/2026-06-26-flow-dev-split-design.md`. Este plan deriva de ese diseño; si divergen, el diseño manda (actualizarlo primero).
> **Repo:** config-only — sin build/test/runtime. Verificación = read-review contra el spec + `python3 harness/build.py` limpio + grep de referencias colgantes. Aplican las *Rule Exclusions* del `AGENTS.md` (no testing, no supply-chain).
> **Modo:** Isolated (rama `feat/flow-dev-split`); commits autónomos por task verificada; checkpoint cada ~3 tasks que modifican archivos.

**Goal:** Reemplazar `flow-dev` por `flow-plan` (research|write) + `flow-build` (reconciler + verify), con estado portable en `plan.md`+git, capacidad harness-agnóstica, y el nivel **iniciativa** en la session-capture layer — todo reflejado en reglas y en el árbol harness.

**Arquitectura:** Markdown puro (skills, reglas, plantillas de flow-core) + regen del árbol generado vía `harness/build.py`. Foundation primero (los contratos que los skills consumen), luego los skills, luego reglas/routing, luego regen y limpieza.

---

### Task 1: flow-core — formato del `plan.md` y contrato de handoff

**Resuelve:** open questions #1 y #2 del spec. Es foundation: los skills consumen esto.

**Files:**
- Create: `global/skills/flow-core/references/plan-format.md` — plantilla harness-agnóstica del plan.
- Modify: `global/skills/flow-core/references/handoff-protocol.md` — añadir el contrato `research → write → build → verify`.
- Modify: `global/skills/flow-core/SKILL.md` — listar la nueva referencia.

- [ ] **Step 1:** Definir en `plan-format.md`: header `Status: planned → building → built → verified`; bloque de task con ID local (`T1`…), `Files`/`Interfaces`/pasos atómicos, flag `in-vivo: sí/no`, y comando de verificación con output esperado; convención de commit `feat(scope): T<n> …`; nota de "ingeniero con cero contexto" y harness-neutralidad (vía `harness-mechanics.md`).
- [ ] **Step 2:** En `handoff-protocol.md`, añadir qué consume/escribe cada paso del contrato y cómo se refleja en el ledger/iniciativa.
- [ ] **Step 3 (verificación):** read-back contra el spec §4.2–§4.4; confirmar que el formato no asume ningún harness.

---

### Task 2: project-structure.md — regla de iniciativa

**Files:**
- Modify: `global/rules/workflow/project-structure.md` — extender la *Session capture layer* con el nivel iniciativa.

- [ ] **Step 1:** Añadir la regla: one-off → sesión plana (sin cambios); esfuerzo grande multi-parte → `sessions/<fecha-inicio>-<slug>/` con `README.md`, `findings/`, `plan/` (partido) y ejecuciones dated adentro. Fecha de inicio inmutable; estado escala por parte; promoción tardía vía `flow-hygiene`.
- [ ] **Step 2 (verificación):** read-back contra spec §4.6; confirmar coherencia con las reglas existentes de naming/lifecycle de la misma sección (no contradecir "los archivos top-level de sesión llevan slug, no fecha" — la iniciativa es contenedor, no archivo).

---

### Task 3: harness/AGENTS.md — reflejar la regla de iniciativa

**Por qué:** `global/rules/` solo deploya a Claude Code; el equipo en Codex/opencode lee `harness/AGENTS.md` (regla "Rules don't pass through build.py").

**Files:**
- Modify: `harness/AGENTS.md` — versión condensada de la regla de iniciativa.

- [ ] **Step 1:** Añadir el párrafo condensado de iniciativa donde vive la guía de estructura.
- [ ] **Step 2 (verificación):** replicar a targets con el comando del `AGENTS.md` y validar `cmp`/tail/`wc -c < 49152` (solo si vamos a desplegar; si no, dejar anotado para el deploy).

> **Checkpoint 1** (tras Task 3): foundation + reglas listas. Resumen + estado antes de los skills.

---

### Task 4: flow-plan — skill nuevo (research | write)

**Files:**
- Create: `global/skills/flow-plan/SKILL.md`
- Create: `global/skills/flow-plan/references/` (si aplica; mover aquí lo de research/plan que tuviera flow-dev).

- [ ] **Step 1:** Frontmatter (`disable-model-invocation: true`, `argument-hint: "[research | write]"`), lenguaje de modelo **advisory** + **neutralidad de harness** (research/plan corren en cualquier harness; el modelo capaz es recomendación de costo, no candado — §4.5).
- [ ] **Step 2:** `research` → exploración técnica read-only → `<slug>-findings.md`; acotado a implementación (no producto/UX); puede rutear de vuelta a `flow-specs` si halla inviabilidad; sin gate; sugiere `/flow-plan write`.
- [ ] **Step 3:** `write` → plan auto-contenido per `plan-format.md`; si el alcance es enorme, propone partir en varias partes y **crea la iniciativa** (Task 2); plan gate; CLOSE sugiere los siguientes pasos (comando exacto de `flow-build` + dónde, según harness/costo).
- [ ] **Step 4 (verificación):** read-back contra spec §4.1, §4.3; confirmar que cita reglas (no las duplica).

---

### Task 5: flow-build — skill nuevo (reconciler + verify)

**Files:**
- Create: `global/skills/flow-build/SKILL.md`

- [ ] **Step 1:** Frontmatter (`disable-model-invocation: true`, `argument-hint: "[verify]"`), lenguaje de modelo advisory + neutralidad de harness.
- [ ] **Step 2:** Default = reconciler: lee `Status`+git, actúa según la tabla del spec §4.2 (`planned`/`building`/`built`/`verified`).
- [ ] **Step 3:** Pregunta in-vivo **una sola vez por sesión** (antes de la ejecución), *inline vs diferido*, solo si la sesión tiene ≥1 task `in-vivo: sí`; la decisión **persiste** para toda la corrida (no se repregunta por task) hasta que el usuario la cambie — patrón *once-per-session* de `git-workflow.md`. Preserva las mecánicas de verificación que ya tenía flow-dev (two-stage review, smoke integrado).
- [ ] **Step 4:** `verify` salta al gate (asume `built`); corre gates pendientes.
- [ ] **Step 5 (verificación):** read-back contra spec §4.2, §4.4, §4.5; confirmar que nada es "frontera-only".

> **Checkpoint 2** (tras Task 5): skills nuevos listos. Resumen antes de borrar flow-dev y tocar routing.

---

### Task 6: Retirar flow-dev + actualizar routing

**Files:**
- Delete: `global/skills/flow-dev/`
- Modify: `global/rules/workflow/agent-routing.md` — cadena F6 (flow-dev → flow-plan/flow-build).
- Modify: referencias en `flow-core/SKILL.md`, `flow-pack-design.md`, `2026-06-14-...-contract`, `AGENTS.md` (lista de skills).

- [ ] **Step 1:** `grep -rn "flow-dev" global/ harness/ AGENTS.md CLAUDE.md _support/spec/` → inventario de referencias.
- [ ] **Step 2:** Actualizar cada referencia a flow-plan/flow-build; eliminar el skill.
- [ ] **Step 3 (verificación):** `grep -rn "flow-dev" global/ harness/` debe quedar vacío (salvo specs históricos que documentan el cambio, que se dejan con nota).

---

### Task 7: flow-hygiene — consciente de iniciativas

**Files:**
- Modify: `global/skills/flow-hygiene/SKILL.md`

- [ ] **Step 1:** `audit` reconoce la estructura de iniciativa (no la marca como desorden); detecta candidatos a promoción plana→iniciativa.
- [ ] **Step 2 (verificación):** read-back contra spec §4.6.

---

### Task 8: Regen del árbol harness + wrappers opencode

**Files:**
- Run: `python3 harness/build.py`
- Create: `harness/opencode/commands/` wrappers para `flow-plan` y `flow-build`.
- Commit: árboles regenerados (`harness/agents-skills/`, `harness/codex/agents/`, `harness/opencode/agents/`).

- [ ] **Step 1:** Crear los command wrappers de opencode para los skills nuevos.
- [ ] **Step 2:** `python3 harness/build.py` y confirmar árbol limpio.
- [ ] **Step 3 (verificación):** `git status` muestra solo regen esperado; no hay skills nuevos sin wrapper.

---

### Task 9: Bibliografía de metodología

**Files:**
- Modify: `_support/docs/methodology-bibliography.md` — registrar las fuentes externas que fundamentan las convenciones nuevas (reconciler/control-loop, plan-then-execute, optional reference project).

- [ ] **Step 1:** Añadir las entradas con su uso en el pack.

> **Checkpoint 3 / Gate final:** read-review completo contra el spec, `build.py` limpio, grep de `flow-dev` vacío. Presentar para merge/PR de `feat/flow-dev-split`.

---

## Orden y dependencias

```
Task 1 (flow-core contract) ──┬─► Task 4 (flow-plan)
Task 2 (project-structure)  ──┘   Task 5 (flow-build)
Task 3 (harness/AGENTS)            │
                                   ▼
                          Task 6 (retirar flow-dev + routing)
                                   ▼
                          Task 7 (flow-hygiene) ─► Task 8 (regen) ─► Task 9 (biblio)
```

## Verificación global (config repo)

- **Sin tests/build de runtime.** El gate es read-review contra el design spec + `python3 harness/build.py` limpio + grep de referencias colgantes.
- **Deploy es user-initiated** — este plan NO ejecuta `/deploy-global` ni copia a `~/.claude/`. La replicación de `harness/AGENTS.md` (Task 3 Step 2) se anota para el deploy si no se hace en sesión.
