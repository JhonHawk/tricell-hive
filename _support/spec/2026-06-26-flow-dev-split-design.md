# Flow-dev Split — research/plan/build con estado portable y organización por iniciativa

> **Estado:** diseño propuesto, pendiente de aprobación.
> **Fecha:** 2026-06-26.
> **Fuentes:** sesión de diseño (este spec la consolida); `optional reference project` (`_support/workspace/optional reference project-main`); `flow-pack-design.md` (2026-06-11); `2026-06-14-flow-phase-transition-contract-design.md`; bodies de `flow-dev` y `flow-core/references/{handoff-protocol,harness-mechanics}.md`.
> **Fundamento externo:** Plan-then-Execute agéntico (CHI 2025, arXiv 2509.08646) [1]; separación design→plan→execute con modelos baratos para trabajo mecánico (optional reference project `writing-plans`/`subagent-driven-development`) [2]; reconciliación declarativa *observed-vs-desired* / control loop (Kubernetes controllers) [3]; artifact-as-handoff + DoR/DoD (ya adoptado en el contrato de transición de fases) [4]; Stage-Gate proporcional (Cooper) [5].

---

## 1. Problema

`flow-dev` (F6) concentra hoy en un solo skill de 263 líneas: scope (Phase 0) + plan con routing table + preflight (Phase 1) + loop de ejecución (Phase 2) + verificación (Phase 3) + close (Phase 4). Tres tensiones, una raíz común: **pensar y ejecutar están fusionados en un mismo skill y una misma sesión.**

| # | Síntoma | Causa raíz |
|---|---|---|
| T1 | No se puede correr el plan en un modelo/harness más barato. El plan (Phase 1) asume que el mismo modelo que planeó ejecuta a continuación, con todo el contexto de la sesión en la cabeza. | El handoff plan→ejecución es **interno a la sesión**, no un artefacto durable que otro harness levante en frío. |
| T2 | No hay research técnico explícito. Phase 0 solo *lee* scope. La exploración de implementación —patrones del repo, enfoques, gaps, riesgos—, que las reglas ya mandan (`gap-resolution.md`, `critical-thinking.md`), no tiene fase ni artefacto propio. | El research está **disperso en reglas** pero no materializado en un paso con salida durable. |
| T3 | Un esfuerzo grande multi-sesión se fragmenta: ejecutado por partes en días distintos, cada corrida crea una carpeta `sessions/YYYY-MM-DD-<slug>/` top-level separada, y el plan original queda huérfano (caso observado: i18n — `plan/` con 7 partes el día 1, y luego `proj-428-auth`, `proj-429-navigation`, `proj-430-finance` como carpetas sueltas en días siguientes). | La capa de sesiones tiene **un solo nivel**; no existe un contenedor para agrupar el trabajo que es un mismo contexto. |

## 2. Fundamento

Los tres síntomas son problemas resueltos en práctica establecida:

1. **Plan-then-Execute es un patrón agéntico de primera clase** [1]: el plan se aprueba antes de ejecutar; lo rutinario corre autónomo y solo escalan las decisiones difíciles de revertir. La asimetría actual (flow-dev planea y ejecuta en un mismo aliento) es lo que impide aprobar/portar el plan.
2. **Separar design→plan→execute habilita modelos baratos para lo mecánico** [2]: optional reference project separa `writing-plans` (modelo capaz) de la ejecución porque, *cuando el plan es una receta completa*, lo mecánico lo puede correr el modelo más barato. El juicio se front-loadea al planear.
3. **La ejecución idempotente es un reconciler** [3]: en vez de "ejecutar a ciegas", el ejecutor lee el **estado deseado** (plan) y el **estado observado** (git) y converge hacia el deseado. Es el patrón control-loop de los controllers declarativos. Hace la ejecución reanudable y portable entre harnesses sin un ledger nuevo.
4. **El handoff entre fases es un artefacto, no la memoria de sesión** [4]: el par DoR/DoD ya adoptado en el pack se extiende — el `plan.md` se vuelve el artefacto de transición que cualquier harness consume.
5. **Los gates escalan al costo/irreversibilidad** [5]: research es read-only (sin gate); el plan tiene gate; el in-vivo se declara por-task (no se asume ni se omite).

## 3. Decisiones

| Decisión | Resolución |
|---|---|
| **Skills** | `flow-dev` se reemplaza por `flow-plan` (subcomandos `research` \| `write`) + `flow-build` (default + subcomando `verify`). |
| **Por qué `flow-plan` bundlea research+write** | Comparten perfil (pensar/diseñar; se benefician del modelo más capaz). El punto **natural** para descargar a un harness barato es `build` —pero es recomendación de costo, no candado: los tres corren en cualquier harness (§4.5). |
| **Estado cross-harness** | El `plan.md` **es** el estado: header `Status: planned → building → built → verified` **por parte de plan**; git es la verdad de las tasks. |
| **`flow-build` reconciler** | Idempotente y dirigido por estado: lee `Status` + git y hace lo siguiente pendiente. Resuelve cross-harness, reanudación tras compactación e híbrido per-sesión con un solo mecanismo. |
| **Tasks** | Auto-contenidas, con ID local al plan (`T1`, `T2`…), independientes de Linear (el tracker es proyección aguas abajo, no dependencia). |
| **Gate in-vivo** | Declarado **por-task** en el plan (flag `in-vivo: sí/no`, lo decide el planificador). `flow-build` default pregunta **una sola vez por sesión** (antes de ejecutar) *inline-vs-diferido*, solo si hay tasks con gate; la decisión persiste hasta que el usuario la cambie. `flow-build verify` salta directo al gate. |
| **Capacidad harness-agnóstica** | Ningún skill ni gate es exclusivo de la frontera. El split modelo/harness es una **elección de costo del operador**, no un requisito. |
| **Organización** | Nuevo nivel **iniciativa** para esfuerzos grandes multi-sesión; one-off sigue siendo sesión plana. |
| **Dónde vive la convención** | La regla de iniciativa va a `project-structure.md` (no depende de invocar la skill) **y** se refleja en `harness/AGENTS.md` (equipo en Codex/opencode). |

## 4. Diseño

### 4.1 Estructura de skills

```
flow-plan research   (cualquier harness)  → <slug>-findings.md   [read-only, sin gate]
flow-plan write      (cualquier harness)  → plan(es)             [plan gate; parte si denso]
flow-build           (cualquier harness)  → ejecuta dirigido por estado (+ pregunta in-vivo)
flow-build verify    (cualquier harness)  → salta al gate in-vivo (asume built)
```

`flow-dev` desaparece. La trinidad lógica es **research → plan → build**, con `verify` como atajo manual al gate dentro de `flow-build`. **Los cuatro corren en cualquier harness** (§4.5): usar el modelo más capaz para `research`/`write` y descargar `build` a uno más barato es una **recomendación de costo**, no un candado — el mismo principio que el gate in-vivo. Ningún paso queda bloqueado para quien solo tenga un harness o agote límites en él.

### 4.2 El plan como estado portable (reconciler)

La pregunta operativa —*¿cómo sabe otro harness que ya se ejecutó un plan?*— se resuelve sin ledger nuevo ni memoria de sesión, porque el estado vive en **archivos + git**:

- **Estado de task = git (verdad).** Cada task aterriza con un commit que referencia su ID local: `feat(scope): T3 …`. Las tasks hechas = las que aparecen en `git log`. Consistente con la regla existente "git es autoritativo para estado de implementación".
- **Estado de plan = header `Status` en el `plan.md`.** Bajo churn, una línea: `Status: planned → building → built → verified`. El `plan.md` está versionado en la sesión y **viaja con el trabajo**.
- **`flow-build` es un reconciler** [3]: al abrir lee `Status` + git y converge hacia lo siguiente pendiente:

  | `Status` observado | Acción de `flow-build` |
  |---|---|
  | `planned` | ejecuta tasks desde la primera sin commit |
  | `building` (cortado / compactado / cambio de harness) | reanuda desde la primera task sin commit en git |
  | `built` | corre el gate in-vivo/review de las tasks marcadas `in-vivo: sí` que no hayan pasado |
  | `verified` | nada que hacer; reporta |

  Un solo mecanismo cubre: handoff cross-harness (grok deja `built`; abres otro harness, ve `built` → corre el gate), reanudación tras compactación, e híbrido per-sesión. El reviewer **verifica el header contra git** antes de confiar (mismo patrón ledger-vs-git ya en uso).

### 4.3 Tasks auto-contenidas

Estándar optional reference project "ingeniero con cero contexto" [2]: cada task carga su definición completa —archivos exactos, interfaces consume/produce, pasos atómicos— y su **comando de verificación con output esperado**. Así el ejecutor no necesita *juzgar*, solo *correr y comparar*: el juicio lo front-loadeó el planificador. El ID es **local al plan**; el mapeo a Linear es una proyección opcional aguas abajo — el plan nunca depende del tracker para saber qué se hizo.

### 4.4 Gate in-vivo

- **El plan declara dónde** (flag `in-vivo: sí/no` por task) → no se asume ni se omite. Lo decide el planificador según si la task es UI/integración vs lógica pura.
- **`flow-build` default decide cuándo — una sola vez por sesión.** Al abrir la ejecución (antes de la task 1), pregunta **obligatoria** *inline (tras cada task con gate) vs diferido (hasta `built`)*, **solo si la sesión tiene ≥1 task con `in-vivo: sí`**. La respuesta es una **decisión de sesión que persiste** para toda la corrida —no se repregunta por task, aunque haya 10— hasta que el usuario la cambie explícitamente. Es el patrón *once-per-session* de `git-workflow.md` (branching, commit mode): se pregunta una vez, persiste, el usuario la cambia con una frase explícita. Si no hay tasks in-vivo, la pregunta no se dispara.
- **`flow-build verify`** salta directo a la etapa de gate (asume `built`): corre los gates pendientes. Es la entrada manual a la transición `built → verified`.

### 4.5 Capacidad harness-agnóstica (principio)

> **La capacidad es harness-agnóstica; el split modelo/harness es una elección de costo del operador, nunca un requisito estructural. Ningún skill ni gate es exclusivo de la frontera.**

Razón: los skills se comparten con el equipo, y no todos tienen el plan de mayor capacidad. Un compañero que agote límites en su harness debe poder ejecutar **todo** —`research`, `write`, `build` y el gate in-vivo— en el que le quede. Aplica por igual a `flow-plan` (research/planning) y al gate: ninguno está bloqueado a un harness. Implicaciones de diseño:

- El "híbrido per-sesión" (planear en frontera, construir en harness barato, revisar en frontera) es una **recomendación de costo**, no una regla de capacidad.
- El lenguaje de modelo en los skills es **advisory** ("usa el modelo más capaz que tengas; puedes descargar build a uno más barato"), nunca prescriptivo ("research corre en Claude Code").
- Los skills se escriben **neutrales al harness** vía `flow-core/references/harness-mechanics.md`.
- Esto **refuerza** el reconciler de §4.2: un compañero construye en el harness A, agota límites y reanuda/verifica en el B; el estado está en `plan.md` + git, no en la sesión.

### 4.6 Organización por iniciativa

Extiende la *session-capture layer* de `project-structure.md` con un nivel intermedio para esfuerzos grandes multi-sesión.

```
sessions/
├── 2026-06-26-proj-431-fix-rounding/        ← one-off: sesión plana (SIN cambios)
│   ├── proj-431-fix-rounding-plan.md
│   └── proj-431-fix-rounding-findings.md
│
└── 2026-06-25-i18n-frontend-replica/       ← INICIATIVA (agrupa el esfuerzo multi-día)
    ├── README.md                           ← índice: alcance, partes, ejecuciones, estado agregado
    ├── findings/                           ← inputs · flow-plan research
    │   ├── audit-findings.md
    │   └── foundation-findings.md
    ├── plan/                               ← inputs · flow-plan write (partido si denso)
    │   ├── i18n-frontend-replica-plan.md    ← plan maestro (lista las partes)
    │   ├── 00-current-state.md
    │   ├── 02-foundation.md
    │   └── 03-domain-migration.md
    ├── 2026-06-25-foundation/              ← ejecución · flow-build (dated)
    │   └── reports/                         ← in-vivo, smoke
    ├── 2026-06-26-auth/
    └── 2026-06-26-navigation/
```

**Reglas:**
- **One-off → sesión plana** (sin cambios). **Esfuerzo grande multi-parte → iniciativa.** Disparador: cuando `flow-plan write` detecta alcance enorme y **propone partir el plan**, ese acto crea la iniciativa.
- **Lectura:** `findings/` y `plan/` son los inputs nombrados; las carpetas con fecha son las ejecuciones.
- **Fecha de inicio en la iniciativa** (inmutable, como arranque de proyecto) → preserva el orden cronológico del índice. Las ejecuciones internas llevan su propia fecha; el "cambio de día" ya no fragmenta.
- **El estado escala por parte:** cada archivo de `plan/` lleva su propio `Status`; el plan maestro las lista. Una parte puede estar `verified` mientras otra está `planned`.
- **Promoción tardía:** si algo arranca plano y crece, se promueve a iniciativa (mover bajo carpeta nueva); lo absorbe `flow-hygiene`.

## 5. Impacto — archivos a tocar

| Archivo / área | Cambio |
|---|---|
| `global/skills/flow-dev/` | Eliminar; reemplazar por `flow-plan/` y `flow-build/`. |
| `global/skills/flow-plan/SKILL.md` | Subcomandos `research` \| `write`; estándar de plan auto-contenido; partición → iniciativa; CLOSE sugiere siguientes pasos. |
| `global/skills/flow-build/SKILL.md` | Reconciler dirigido por estado; pregunta in-vivo; subcomando `verify`; lenguaje de modelo advisory. |
| `global/rules/workflow/project-structure.md` | Nueva regla: nivel **iniciativa** en la session-capture layer. |
| `harness/AGENTS.md` | Reflejar la regla de iniciativa (condensada) — el equipo en Codex/opencode no lee `global/rules/`. |
| `global/skills/flow-core/references/` | Contrato de handoff `research → write → build → verify`; formato harness-agnóstico del `plan.md` (Status, tasks con ID local, flag in-vivo); posible plantilla de plan. |
| `global/rules/workflow/agent-routing.md` | Actualizar la tabla/cadena F6 (flow-dev → flow-plan/flow-build). |
| `global/skills/flow-hygiene/SKILL.md` | Auditar iniciativas (estructura, promoción tardía plana→iniciativa). |
| `harness/build.py` + `harness/opencode/commands/` | Regenerar; wrappers de opencode para los skills nuevos. |
| Referencias cruzadas | `flow-pack-design.md`, `2026-06-14-...-contract`, `_support/docs/methodology-bibliography.md`. |

## 6. Preguntas abiertas (para el plan de implementación)

1. **Formato exacto del `plan.md`** auto-contenido y harness-agnóstico (header `Status`, bloque de task con ID local + flag in-vivo + comando de verificación). ¿Plantilla en `flow-core/references/`?
2. **Contrato de handoff** `research → write → build → verify` en `flow-core` — qué consume cada paso, qué escribe en el ledger/iniciativa.
3. **`flow-hygiene`**: alcance del soporte a iniciativas (auditar vs. también promover plana→iniciativa en `apply`).
4. **Índice de iniciativa** (`README.md`): ¿generado por `flow-plan write` y mantenido por `flow-build`, o derivado de los `Status` de las partes?
5. **Migración de `flow-dev`**: ¿deprecación con puntero, o eliminación directa? Grep de referencias en el pack y en `harness/`.

## 7. Referencias

1. Plan-then-Execute en agentes — CHI 2025, arXiv:2509.08646 (citado también en el contrato de transición de fases).
2. optional reference project — `_support/workspace/optional reference project-main/skills/{brainstorming,writing-plans,subagent-driven-development,executing-plans}/SKILL.md`. Modelo: separación design→plan→execute; "ingeniero con cero contexto"; modelos baratos para trabajo mecánico.
3. Control loop / reconciliación declarativa observed-vs-desired — Kubernetes controllers (patrón establecido de convergencia hacia estado deseado).
4. Artifact-as-handoff + DoR/DoD — adoptado en `2026-06-14-flow-phase-transition-contract-design.md`.
5. Stage-Gate proporcional — Cooper (gates escalados al riesgo; ya fundamenta el gate proporcional del pack).
