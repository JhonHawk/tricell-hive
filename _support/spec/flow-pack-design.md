# Flow Pack — Flujo de desarrollo end-to-end como skills + agentes

> **Estado:** diseño aprobado, pendiente de construcción. Piloto: `/flow-dev`.
> **Fecha:** 2026-06-11.
> **Fuentes:** sesión de diseño (este spec la consolida); reportes `_support/workspace/{evaluacion-tooling-calidad,codex-quality-ops-chat-hub}.html`; `~/Documents/obsidian/tricell-mac/Conventions/{Naming Convention,Projects-Workspaces}.md`; specs de Warp (<https://github.com/warpdotdev/warp/tree/master/specs>) como referencia de estructura.

---

## 1. Problemas que este pack resuelve

| # | Problema observado | Causa raíz | Fix estructural |
|---|---|---|---|
| P1 | El hilo principal hace todo el trabajo; los agentes especialistas no se invocan; contexto sucio | El routing por descripción es probabilístico y compite contra el sesgo del hilo de hacerlo él mismo | Skills como puertas determinísticas: la delegación está escrita en el proceso, no dejada al criterio de cada turno |
| P2 | Se cumple el mínimo sin cuestionar si la funcionalidad debe replantearse | La autocrítica del hilo que implementa es estructuralmente débil (momentum de implementación) | Fase de crítica separada, agente separado, contexto fresco, ANTES de implementar (`product-critic` en spec intake) |
| P3 | Archivos generados sin orden determinístico; difícil encontrar material en sesiones posteriores | Sin proceso que escriba el índice; cada sesión improvisa rutas | Ledger `PROJECT.md` + regla de ruteo de archivos (§3) — determinismo por artefacto, no por disciplina |
| P4 | Recursos de infra con nombres inconsistentes | La convención no estaba escrita/ambiental al momento de pedir cada despliegue | Regla global `infra-naming.md` + tabla instanciada por proyecto en el repo de specs (§3.4) |
| P5 | Deploy "a mano" a QA; re-investigación de comandos en cada arreglo; migración tardía a CD (doble trabajo) | El pipeline se construyó después del primer deploy | Regla de oro en F5: el pipeline existe antes que el primer deploy; el primer deploy lo hace el pipeline |

## 2. Principios de diseño

1. **Skill = proceso, agente = criterio, regla = conocimiento.** La skill define fases y orquestación; el agente define cómo juzgar/construir su parte (prompt de sistema, herramientas, contexto aislado); las reglas globales definen convenciones. Cada cosa vive una sola vez — las skills citan reglas, nunca las duplican.
2. **El hilo principal orquesta, no implementa.** Todo trabajo sustantivo se despacha a un agente. La higiene de contexto se logra porque al hilo principal solo regresan resúmenes.
3. **Verificación en contexto fresco.** Quien revisa nunca es quien implementó (regla ya existente en `agent-routing.md`; aquí se vuelve mecánica).
4. **Determinismo por artefacto.** Una regla abstracta evaluada al vuelo produce interpretaciones distintas por sesión; una tabla/plantilla escrita una vez produce el mismo resultado siempre (ledger, naming instanciado, plantillas de flow-core).
5. **Gates explícitos, invocados por el usuario.** Las skills de fase llevan `disable-model-invocation: true` — son puertas que el usuario cruza, no atajos que el modelo toma. El trabajo conversacional diario queda fuera del pack (no se intenta arreglar con más reglas).
6. **Una fuente de verdad por convención.** El hub es canónico; los docs de Obsidian se convierten en punteros.

## 3. Arquitectura de información

### 3.1 Regla de ruteo de archivos (canónica)

Antes de escribir cualquier archivo, toda skill del pack (y cualquier sesión sobre un workspace) responde en orden:

```text
1. ¿Debe tener historial/versionado?
   Sí → <project>-specs

2. ¿Es temporal, sensible, evidencia bruta o scratch?
   Sí → _support

3. ¿Aplica solo a un repo?
   Sí → <repo>/_support o al lugar natural del repo

4. ¿Un reporte temporal ya produjo una decisión?
   Sí → promoverlo/resumirlo dentro de <project>-specs
```

**Semántica resultante:**

- **`<project>/_support/`** (workspace-level) — lo NO versionable: reportes temporales, scratch, evidencia bruta (screenshots, traces, exports), material sensible. Nunca es archivo histórico: lo que produce decisiones se promueve (pregunta 4) y el resto caduca.
- **`<project>-specs/`** — la memoria versionada del proyecto: épicas, contratos, convenciones instanciadas, decisiones, resúmenes de evidencia. Evolución natural de `_support` a partir de F3.
- **`<repo>/_support/`** — cada repo conserva su propia carpeta `_support` para material repo-local (scripts desechables, infrastructure, workspace, evidencia del repo). Sin cambios.

> **Delta pendiente a `global/rules/workflow/project-structure.md`:** hoy declara `_support/docs/` como "documentación durable". Con esta regla, a nivel workspace lo durable migra al repo de specs una vez que existe (F3); `_support` queda como capa no-versionada. Actualizar la regla global y convertir los docs de Obsidian en punteros es parte del build (§9).

### 3.2 PROJECT.md — el ledger operativo

- **Ubicación:** `<project>/_support/PROJECT.md`. Es índice operativo, no archivo: fase actual, punteros a artefactos (con rutas), pendientes, preguntas abiertas.
- **Contrato:** toda skill del pack lo lee al abrir y lo actualiza al cerrar (§5).
- **Promoción:** las decisiones que registra se promueven/resumen al repo de specs cuando existe (pregunta 4 de la regla de ruteo). El ledger apunta; el specs repo preserva.
- **Plantilla:** `flow-core/references/ledger-template.md`.

### 3.3 Repo de specs (estructura estilo Warp)

- Creado por `/flow-specs init` como `<project>/<project>-specs/`, git versionado, vinculado a Linear.
- Convenciones (a extraer del repo real de Warp durante el build, hacia `flow-core/references/specs-structure.md`): specs numeradas, una carpeta por feature/épica, status en el header, división clara Épica → Tasks con criterios Gherkin.
- Contiene además: `conventions/naming.md` (tabla instanciada, §3.4), decisiones promovidas, resúmenes de evidencia.

### 3.4 Naming en dos capas

> Decisión 2026-06-11: la capa intermedia de presets de cliente (`apply-preset`) se eliminó — la skill no se usaba. Lo específico de cliente colapsa a los artefactos del propio proyecto.

| Capa | Artefacto | Contenido | Ciclo de vida |
|---|---|---|---|
| 1. Regla global | `global/rules/workflow/infra-naming.md` (NUEVA) | Regla de oro `<proyecto>-<componente>-<env>` (kebab-case, env completo al final), tabla por tipo de recurso, excepciones estructurales (DNS con prefijo corto, repos/ECR sin env, snake_case en DBs, stage infijo de Serverless), límite 32 chars ALB/TG | Atemporal, todo cliente |
| 2. Tabla instanciada | `<project>-specs/conventions/naming.md` | Nombres concretos de TODOS los recursos del proyecto (servers, buckets, IAM, secrets, subdominios, DBs), derivados de la capa 1; **las excepciones del cliente se documentan in situ con su razón** (p.ej. DBs invariantes de chat-hub), igual que los legados intocables | Por proyecto; output de F5, input de F7 |

Las convenciones de cliente que no son de naming (modelo de branching, cuentas/regiones) viven en el **`AGENTS.md` del workspace** que genera `/flow-kickoff` (ver §6.2), con `CLAUDE.md` importándolo vía `@AGENTS.md` — así toda sesión de todo harness (Claude por import; Codex/opencode nativo) las carga sin duplicación.

La instanciación (capa 2) existe porque derivar nombres al vuelo produce interpretaciones distintas por sesión (¿`bo` o `backoffice`?); la tabla escrita una vez produce el mismo nombre siempre. Todo recurso nuevo se valida contra la tabla antes de crearse; si no está, primero se agrega (con visto bueno del usuario si es excepción), luego se crea.

## 4. Estructura del pack

```text
global/skills/
├── flow-core/                      # NO invocable — biblioteca compartida
│   ├── SKILL.md                    # user-invocable: false; documenta el pack
│   └── references/
│       ├── ledger-template.md      # estructura canónica de PROJECT.md
│       ├── handoff-protocol.md     # qué DEBE incluir todo prompt a un agente
│       ├── naming-template.md      # plantilla de la tabla instanciada
│       └── specs-structure.md      # convenciones estilo Warp del repo de specs
├── flow-intake/                    # F1 — análisis del documento de requerimientos
├── flow-kickoff/                   # F2 — bootstrap del workspace + ledger
├── flow-specs/                     # F3 — init | epic | review
├── flow-mock/                      # F4 — build | review
├── flow-foundation/                # F5 — repos + schema + contratos + CI/CD
├── flow-dev/                       # F6 — sesión de desarrollo con routing table
├── flow-deploy/                    # F7 — qa | prod | verify
└── flow-hygiene/                   # transversal — audit | apply de higiene documental

global/agents/
├── design/requirement-analyst.md   # NUEVO (azul)
├── review/spec-quality-reviewer.md # NUEVO (cyan)
├── review/product-critic.md        # NUEVO (cyan)
├── review/ux-flow-reviewer.md      # NUEVO (cyan)
└── quality/workspace-custodian.md  # NUEVO (amarillo)
# F5–F7 reutilizan los agentes existentes: database-specialist, backend-developer,
# nextjs-architecture-expert, angular-developer, test-engineer, code-reviewer,
# security-reviewer, devops-engineer, system-designer, Explore.
```

Las skills leen los compartidos vía ruta estable post-deploy: `~/.claude/skills/flow-core/references/...`.

**Herramientas externas siempre disponibles:** Linear (MCP), GitHub (`gh`), `aws` CLI, `hcloud` CLI, chrome-devtools/playwright MCP (los subagentes los alcanzan vía ToolSearch).

## 5. Contrato común (flow contract)

Toda skill `flow-*` cumple:

```markdown
1. OPEN — read `<project>/_support/PROJECT.md`. If missing (and this is not
   kickoff): stop, suggest /flow-kickoff. Never improvise workspace structure.
2. ROUTE FILES — before writing any file, apply the file-routing rule (§3.1).
3. ORCHESTRATE — the main thread routes and synthesizes. It does NOT implement,
   review, or verify by itself. Every substantive work unit goes to an agent.
4. HANDOFF — every agent prompt follows handoff-protocol.md: intent (why + who
   consumes the output), spec excerpt or concrete paths (never "read
   everything"), conventions that apply, expected output shape.
5. CLOSE — update PROJECT.md: phase, artifacts written (with paths), decisions
   taken (and whether promoted to the specs repo), open questions. Then report.
```

## 6. Skills — especificación por fase

> **Decisión allowed-tools (2026-06-11, sellada en build):** ninguna flow skill declara
> `allowed-tools`. 6 de 8 necesitan herramientas fuera de cualquier allowlist razonable
> (Linear MCP, ToolSearch, Skill tool, task tools), y una allowlist incompleta falla a
> media sesión — el peor momento. La disciplina de despacho vive en el cuerpo de cada
> skill (nombra exactamente a qué agentes despacha); las dependencias se documentan aquí
> como `Dispatches:` para mantenerlas grepeables (borrar un agente → grep revela qué
> skills lo usaban).

### 6.1 `/flow-intake <doc>` — F1 requirement

- **Dispatches:** `requirement-analyst`
- **Particularidad:** única skill pre-workspace. Si no hay `PROJECT.md`, opera standalone y deja artefactos junto al documento fuente; kickoff los absorbe después.
- **Fases:** (1) `requirement-analyst` evalúa el documento contra rúbrica de completitud (problema, actores, alcance explícito e implícito, reglas de negocio ausentes) → (2) síntesis: documento mejorado + preguntas separadas en **bloqueantes vs nice-to-know** → (3) escribe `requirements-v2.md` + `open-questions.md`.
- **Gate:** las bloqueantes se presentan antes de cerrar — son la agenda para la siguiente conversación con el cliente.

### 6.2 `/flow-kickoff <group> <project>` — F2

- **Dispatches:** ninguno — invoca la skill `engram-init-workspace`
- **Fases:** (1) crea la estructura 3-niveles citando `project-structure.md` → (2) `PROJECT.md` desde plantilla, registrando el **token del proyecto** (semilla del naming), + **`CLAUDE.md` mínimo en la raíz del workspace** (ver abajo) → (3) reubica artefactos de intake → (4) corre `engram-init-workspace` (memoria unificada multi-repo desde el día cero) → (5) registra en el workspace `CLAUDE.md` las convenciones del cliente que no son de naming: modelo de branching (ramas protegidas, flujo development→qa→master si aplica) y cuentas/proveedores, preguntando al usuario lo que no sea inferible.
- **El puente ledger ↔ harness (actualizado 2026-06-11 a par ambiental):** `PROJECT.md` no es convención de harness — solo entra en contexto cuando algo lo lee. El puente usa las convenciones que los harnesses SÍ leen: kickoff crea **`AGENTS.md`** en la raíz del workspace (3-6 líneas: identidad del workspace, puntero al ledger, regla de ruteo §3.1, convenciones de cliente) — Codex y opencode lo cargan nativo — y **`CLAUDE.md` = `@AGENTS.md`** (import oficial de Claude Code; lo Claude-específico iría debajo del import). Con eso, toda sesión de todo harness — conversacional o con skill, en el workspace o dentro de un repo hijo — recibe el puntero automáticamente, desde un solo archivo canónico.

### 6.3 `/flow-specs [init|epic <name>|review <path>]` — F3

- **Dispatches:** `spec-quality-reviewer` + `product-critic` (en paralelo); integra Linear MCP y la skill `flow-report`
- **`init`:** crea `<project>-specs/` con estructura Warp (`specs-structure.md`), git init, vincula proyecto Linear.
- **`epic`:** redacta épica + tasks con Gherkin desde requirements; sincroniza a Linear (Épica → issues). **El gate `review` es parte del subcomando, no opcional**: toda épica pasa por review antes de marcarse lista.
- **`review`:** lanza ambos agentes EN PARALELO (un solo mensaje, contextos frescos):
  - `spec-quality-reviewer` — rúbrica de calidad de spec (claridad del problema, completitud funcional, verificabilidad Gherkin, impacto técnico, ambigüedad), con la ruta de la rúbrica y la intención en el prompt.
  - `product-critic` — desafía necesidad/forma/alcance con contexto completo del workspace (lee otras épicas, contratos, código existente; detecta reglas de negocio implícitas sin escribir).
  - Síntesis: dedup, las contradicciones entre revisores se marcan explícitas (nunca se resuelven en silencio). Clasificación: `blocker | gap | rethink | polish`.
  - Gate: `blocker`/`rethink` van a AskUserQuestion (proceder / corregir spec / discutir). El reporte completo se rendea vía `flow-report` a `_support/workspace/`.

### 6.4 `/flow-mock [build|review]` — F4

- **Dispatches:** `ux-flow-reviewer`; carga skills de diseño según stack (heroui/shadcn + design-taste)
- **`build`:** detecta stack (HeroUI vs shadcn) → carga la skill correspondiente (`heroui-react-pro` + `heroui-pro-design-taste`, o `shadcn` + `design-taste-frontend`) → construye el prototipo épica por épica contra los specs.
- **`review`:** `ux-flow-reviewer` navega el mock vivo (chrome-devtools vía ToolSearch), aplica rúbrica UX (¿el usuario sabe dónde está?, ¿el siguiente paso es obvio?, ¿los errores dicen cómo recuperarse?, estados vacíos, jerarquía, accesibilidad básica), captura screenshots → fricciones por severidad.
- **Regla:** hallazgo que implique cambio de spec NO se arregla solo en el mock — se registra y dispara `/flow-specs epic <name>` para actualizar la fuente de verdad.

### 6.5 `/flow-foundation` — F5

- **Dispatches:** `system-designer`, `database-specialist`, `devops-engineer` (en paralelo)
- **Stack de referencia:** NestJS, Next.js (HeroUI/shadcn), PostgreSQL/MongoDB, Redis; buckets/correo en AWS; QA en Hetzner, producción en AWS.
- **Fases:** (1) **instancia `naming.md`** (capa 1 + capa 2 → tabla concreta) y lo escribe en `<project>-specs/conventions/` → (2) propone matriz de repos (gate con el usuario) → (3) despacho paralelo: `system-designer` (contratos OpenAPI base), `database-specialist` (schema inicial + migraciones), `devops-engineer` (CI verde desde el primer commit, CD **definido con target parametrizado**, y provisioning **escrito** — idempotente, versionado en `<repo>/_support/infrastructure/` — pero NO ejecutado) → (4) cierre **sin entorno vivo**.
- **División foundation/setup (decisión 2026-06-11):** foundation *escribe* la infraestructura; `/flow-deploy setup <env>` la *ejecuta* — incluido el primer QA. Razón: foundation suele arrancar antes de que los recursos/cuentas del cliente estén confirmados; acoplarla al provisioning la bloquearía con lo que el usuario no controla. `/flow-dev` puede arrancar contra servicios locales sin esperar entorno.
- **Regla de oro (vive en setup):** el primer deploy de TODO entorno lo hace el pipeline. Nunca existe el estado "desplegado a mano, luego migramos a CD".

### 6.6 `/flow-dev [epic-id|TRI-xxx ...]` — F6 (piloto)

SKILL.md completo (borrador):

```markdown
---
name: flow-dev
description: >
  Run a development session: pick an epic and tasks from Linear, produce a plan
  with an explicit agent-routing table, dispatch implementation to specialist
  agents, verify in fresh contexts, and close with Linear/git/ledger updates.
  Use when starting a dev session on an existing project with specs.
argument-hint: "[epic-id|TRI-xxx ...]"
disable-model-invocation: true
# Note: no allowed-tools — pack-wide decision, see the note at the top of §6.
---

# Dev Session

Follow the flow contract (flow-core). Read PROJECT.md and naming.md first.

## Phase 0 — Scope
Read the specs repo for the target epic. Query Linear for its open tasks.
Propose session scope: epic + the tasks that realistically fit one session.
$1 overrides: if the user passed an epic or task IDs, scope to those.

## Phase 1 — Plan with routing table
Produce the session plan. EVERY implementation task carries its assignment:

| Task | Agent | Repo | Branch | Spec ref | Depends on |
|------|-------|------|--------|----------|------------|

Routing follows agent-routing.md (schema→database-specialist, API→backend-developer,
UI→nextjs-architecture-expert, CI→devops-engineer). The main thread appears in
NO row — if you find yourself assigning a task to "main", re-route it.
Branches per session scope, named per git-workflow conventions.
Present the plan for approval (native plan mode handles the gate).

## Phase 2 — Execution loop
Dispatch each task to its agent per handoff-protocol.md. Independent tasks in
parallel (single message, multiple Agent calls); dependent tasks sequential,
passing the predecessor's outputs (contract paths, type names, migration ids).
After each agent returns: verify its build/typecheck claim via Bash before
marking the task done. An agent that reports failure → surface it, do not
silently retry with a different agent.

## Phase 3 — Verification (fresh contexts only)
- code-reviewer over the session diff (it did not write the code).
- In-vivo gate: dispatch an agent with chrome-devtools (via ToolSearch) to run
  the epic's Gherkin acceptance criteria against the running app. Evidence
  (screenshots, traces) → _support/evidence/.
- test-engineer ONLY if the session's goal includes a coverage push
  (specialists already test their own code per testing.md).

## Phase 4 — Close
Linear: move task states, comment evidence links. Git: commits per git-workflow
(session branch; push/PR only with explicit user verb). Ledger: update
PROJECT.md (tasks done, decisions, leftovers for next session). Report: epic
progress, what's verified vs pending, suggested next session scope.
```

### 6.7 `/flow-hygiene [audit|apply]` — transversal

- **Dispatches:** `workspace-custodian` (audit; apply ejecuta en el hilo con aprobación)
- **Razón de existir:** el pack no controla el trabajo conversacional diario (principio 5) — y ese trabajo es el que genera el desorden. Las skills mantienen limpio lo que ellas escriben (flow contract); el custodio atrapa el drift de todo lo demás. Es el control compensatorio de la decisión deliberada de no gatear las sesiones conversacionales.
- **`audit`:** lanza `workspace-custodian` (contexto fresco, read-only) sobre el workspace → reporte con acciones propuestas clasificadas: `move` (archivo mal ruteado según §3.1) | `promote` (reporte temporal que produjo decisión y nunca llegó al specs repo) | `expire` (scratch caduco en `_support/workspace/`) | `repair` (ledger con punteros rotos o fase desactualizada) | `conform` (specs repo fuera de las convenciones de estructura).
- **`apply`:** presenta el plan de acciones vía AskUserQuestion (todas las eliminaciones y movimientos requieren aprobación — son operaciones destructivas) y ejecuta solo lo aprobado. Nunca elimina sin confirmación; en duda, propone archivar.
- **Cadencia sugerida:** al cierre de hitos (fin de épica, pre-entrega) o cuando el usuario perciba desorden — no como hook automático (el costo de un falso positivo moviendo archivos supera el beneficio de la automatización).

### 6.8 `/flow-deploy [qa|prod|verify|setup <env>]` — F7

- **Dispatches:** `devops-engineer`; usa `gh`/`aws`/`hcloud` CLIs
- **`qa`:** gates (pipeline verde, migraciones identificadas, **rollback declarado o no hay deploy** — `devops-principles.md`) → deploy vía pipeline (nunca directo) → smoke E2E contra QA → evidencia bruta a `_support/evidence/`, resumen promovido al specs repo → borrador de release notes desde PRs + Linear.
- **`prod`:** mismos gates + confirmación explícita + dry-run de migraciones + runbook AWS. Topología no trivial (DR, multi-región) escala a `cloud-architect` → `devops-engineer` según `agent-routing.md`; el proyecto cliente típico no la necesita.
- **`verify`:** smoke + health standalone, e **incluye audit de naming**: lista recursos vía CLI y diffea contra `naming.md` — el drift se detecta, no se descubre en auditorías.
- **`setup <env>`** (agregado 2026-06-11): ambienta un entorno nuevo (`development|qa|production`) o converge uno con drift. Foundation ambienta QA al arranque; los entornos posteriores (prod cerca de entrega, dev agregado a medio proyecto) entran por aquí. Flujo: plan de recursos derivado de la tabla de naming (filas nuevas ANTES de crear; un solo gate con blast radius) → provisioning vía los scripts idempotentes de `_support/infrastructure/` (despacho a devops-engineer; un paso manual = bug de provisioning a scriptear) → secrets con placeholders en el manager del proyecto → conectar al pipeline como target de CD — la regla de oro de F5 aplica a TODO entorno: su primer deploy lo hace el pipeline → `verify` + registro en ledger y tabla de naming.

## 7. Agentes nuevos

| Agente | Color | Rol | Reglas clave |
|---|---|---|---|
| `requirement-analyst` | azul | Evalúa/mejora documentos de requerimientos | Rúbrica de completitud; separa bloqueantes de nice-to-know; pregunta por alcance implícito y reglas de negocio ausentes |
| `spec-quality-reviewer` | cyan | Calidad de specs/épicas contra rúbrica | Scorecard por dimensión; huecos por severidad; Gherkin corregido; read-only |
| `product-critic` | cyan | Crítica adversarial pre-implementación | Desafía necesidad/forma/alcance, NO completitud (eso es del reviewer); verifica overlap contra el código citando paths; cada objeción incluye la alternativa; **permiso explícito de decir "está bien" y parar** (max 7 hallazgos: `rethink | shrink | question`); read-only |
| `ux-flow-reviewer` | cyan | Fricción UX en mocks/QA navegando en vivo | Rúbrica UX fija; screenshots como evidencia; fricciones por severidad; distingue hallazgo-de-mock vs hallazgo-de-spec; read-only + browser MCP |
| `workspace-custodian` | amarillo | Higiene documental del workspace (audita, no ejecuta) | Aplica la regla de ruteo §3.1 retroactivamente; detecta promociones pendientes, scratch caduco, punteros rotos del ledger, specs fuera de convención; **read-only — propone acciones, nunca las ejecuta** (la ejecución es de `/flow-hygiene apply` con aprobación del usuario); en duda propone archivar, no eliminar |

Las rúbricas viven como markdown plano en `references/` de cada skill (no en el prompt del agente): el mismo archivo lo consume el agente de Claude Code y los adaptadores de otros harnesses. El agente aporta el criterio de juicio; la rúbrica, el checklist del dominio.

## 8. Plantillas de flow-core (decisiones transversales)

- **`ledger-template.md`** — secciones de PROJECT.md: token del proyecto, fase actual, artefactos (tabla ruta + descripción + fecha), decisiones (+ si fueron promovidas al specs repo), pendientes, preguntas abiertas.
- **`handoff-protocol.md`** — todo prompt a agente incluye: (1) intención — el porqué y quién consume el output; (2) contexto acotado — extractos de spec o rutas concretas, nunca "lee todo el repo"; (3) convenciones aplicables (naming, contratos, estructura); (4) forma esperada del output. Es la pieza a calibrar en el piloto.
- **`naming-template.md`** — tabla instanciada: tipo de recurso → plantilla → nombre concreto → estado (existe/planeado/excepción).
- **`specs-structure.md`** — convenciones del repo de specs, extraídas del repo de Warp durante el build.

## 9. Deltas a la configuración global existente

| Cambio | Archivo | Naturaleza |
|---|---|---|
| Nueva regla de naming de infra (capa genérica) | `global/rules/workflow/infra-naming.md` | NUEVO |
| Semántica `_support` vs specs repo (regla de ruteo §3.1) | `global/rules/workflow/project-structure.md` | EDITAR — hoy declara `_support/docs/` durable; lo durable workspace-level migra al specs repo cuando existe |
| Presets → declaración en `CLAUDE.md` del proyecto (protected branches, branching model) | `global/rules/workflow/git-workflow.md` | HECHO 2026-06-11 — `apply-preset` eliminado; el mecanismo "active preset" se reemplazó por la declaración del proyecto |
| Tabla de routing si los 5 agentes nuevos crean ambigüedad | `global/rules/workflow/agent-routing.md` | REVISAR (mandato de Routing Maintenance del hub) |
| Obsidian `Conventions/*.md` → punteros al hub | Obsidian (manual, fuera del repo) | El hub es canónico; absorber deltas menores primero |

## 10. Portabilidad (Codex / opencode) — CONSTRUIDA 2026-06-11

> Verificado contra docs oficiales del 2026-06-11 (opencode v1.17.1; Codex CLI docs
> developers.openai.com/codex). Alcance: pack completo a ambos harnesses.

### Matriz de replicación (qué pasa al actualizar el hub)

| Capa | Contenido | Destino | Al actualizar |
|---|---|---|---|
| 1 · Universal | SKILL.md + rúbricas + plantillas + `agents/openai.yaml` por skill | `~/.claude/skills/` (Claude, copia verbatim) y `~/.agents/skills/` (Codex/opencode — **copia limpiada por `convert-skills.py`**: frontmatter Claude-only removido, `${CLAUDE_SKILL_DIR}` y rutas `~/.claude/skills/` reescritas a `~/.agents/skills/`) | Editar hub + `/deploy-global`. Cero toques a adaptadores |
| 2 · Generada | Agentes per-harness y skills universales limpias — **versionadas en el repo** bajo `harness/{agents-skills,codex/agents,opencode/agents}/` (modelo optional reference project, adoptado 2026-06-11) | Regeneradas por `harness/build.py` tras editar `global/`; el deploy las copia y detecta árboles desactualizados | `build.py` + commit — nunca se editan a mano |
| 2b · Envolturas | Commands opencode (`/flow-*`) en `harness/opencode/commands/` (fuente manual) | `~/.config/opencode/commands/` | Solo si cambia la firma de una skill o se agrega una nueva |
| 3 · Config one-time | Snippets en `harness/{codex,opencode}/` (MCP Linear, gating, env) | Merge manual a `config.toml` / `opencode.json` | Una vez por máquina; READMEs documentan |

### Piezas clave

- **`flow-core/references/harness-mechanics.md`** — la clave de traducción: los mismos SKILL.md sirven a los 3 harnesses; los no-Claude traducen mecánica (Agent tool → spawn/task tool, AskUserQuestion → preguntar inline, ToolSearch → MCP directo, $ARGUMENTS → texto libre en Codex) y respetan las diferencias de enforcement (read-only vía sandbox/permission).
- **Gating de auto-invocación por harness:** Claude `disable-model-invocation` · Codex `agents/openai.yaml` → `policy.allow_implicit_invocation: false` (viaja con la skill) · opencode `permission.skill` `"flow-*": "ask"` (snippet).
- **opencode:** entrada user-facing vía commands (no hay slash de skills); sin variable de entorno para descubrimiento — opencode no duplica el listado (medido 2026-08-21). Campo `tools` de agentes deprecado → el convertidor emite `permission`.
- **Codex:** `$skill-name` nativo, sin envolturas; **tiene subagentes** (TOML, `developer_instructions`, spawn explícito-only — las instrucciones "dispatch X" de las skills cuentan como petición explícita; `max_depth=1` suficiente); revisores read-only vía `sandbox_mode = "read-only"`; sin sustitución de argumentos en skills (texto libre).

### Gaps aceptados

Sin `$ARGUMENTS` formal en skills de Codex; sin auto-routing por descripción de subagentes en Codex; sin allowlist de tools granular en Codex (sandbox + instrucción conductual); AskUserQuestion degrada a pregunta inline fuera de Claude. Documentados en `harness-mechanics.md` — los modelos degradan con gracia.

## 11. Plan de construcción

| Orden | Entregable | Validación |
|---|---|---|
| 1 | `flow-core` (las 4 plantillas — decisiones transversales primero) | Revisión por lectura |
| 2 | **Piloto `/flow-dev`** + verificación de routing con agentes existentes | Sesión real contra una épica de chat-hub; calibrar handoff-protocol con lo aprendido |
| 3 | `/flow-specs` (init + epic + review) + 2 agentes de review + extracción de estructura Warp | Review de una épica real existente |
| 4 | `/flow-hygiene` + `workspace-custodian` | Audit real contra el workspace de chat-hub (ya acumuló el desorden que debe detectar) |
| 5 | `/flow-intake`, `/flow-kickoff` + `requirement-analyst` | Dry-run con un documento de requerimientos real |
| 6 | `/flow-mock` + `ux-flow-reviewer` | Review del mock de chat-hub |
| 7 | `/flow-foundation`, `/flow-deploy` + regla `infra-naming.md` + deltas de §9 | Dry-run de naming contra chat-hub (recursos ya auditados 2026-06-05) |
| 8 | Adaptadores opencode/Codex | Verificación contra docs vigentes de cada harness |

**Riesgos aceptados:** (1) la orquestación tiene costo — agentes arrancan en frío y un handoff mal armado produce trabajo desalineado; se mitiga calibrando `handoff-protocol.md` en el piloto antes de replicar. (2) El pack no arregla el routing implícito del trabajo conversacional — decisión deliberada (principio 5). (3) `disable-model-invocation` en todas las skills de fase al inicio; se relaja con evidencia, no por anticipación.

**Preguntas abiertas:**
- ~~¿El repo de mocks es siempre `<project>-frontend-mocks` (patrón chat-hub) o lo decide F4 por proyecto?~~ **Resuelta (2026-06-12):** siempre `<project>-frontend-mocks`; F4 lo confirma contra la naming table cuando existe (así ya lo describe `flow-mock`).
- ¿Qué clientes reciben preset de naming en la primera ola? (chat-hub ya tiene la convención adoptada y auditada.)
- ~~Formato exacto de release notes para el cliente (¿plantilla en flow-core o por preset?).~~ **Resuelta (2026-06-12):** plantilla canónica en `flow-core/references/release-notes-template.md` (markdown plano, copy-paste al canal real); overrides por proyecto (canal/idioma/tono) en fila opcional `Release notes` del PROJECT.md — nunca presets de estructura por cliente.
