# Flow Pack v2 — Rediseño: de fases de ciclo de vida a cadena de proceso

> **Estado: PROPUESTA — sin aplicar.** Ningún cambio a skills, hooks ni harnesses se ejecuta hasta aprobación explícita de este diseño.
> Supersede (al aplicarse): `flow-pack-design.md` §4 (estructura del pack), §6 (skills por fase) y §10 (matriz de portabilidad).
> Evidencia: `_support/workspace/flow-pack-usage-audit-2026-07-17.html` (canon adversarial, 3 generadores + 2 rondas de refutación).

## 1. Motivación (resumen del canon)

- Post-liberación pública (2026-06-26), las invocaciones explícitas de fase colapsaron: 1 slash en CC (06-27), 1 en Codex (07-10), más flow-report. El trabajo real siguió por canales que el verbo explícito no captura: plan mode nativo («Implement the plan.» ×34/26 sesiones Codex), segundo operador (Ricardo, 35 commits en internal-apps), ejecución headless.
- Las skills gateadas por **fase de proyecto** (una vez por proyecto) no pueden sostener uso diario; los packs de referencia (improve, optional reference project) sostienen uso diario porque gatean en **condiciones recurrentes** (verbo repetible o intención que recurre).
- Criterio del usuario: skills para el flujo de trabajo del día a día; un plan por unidad de trabajo (el pipeline spec → mock → build con planeación por fase era el desgaste); foundation nunca corrió vía skill (3 ejecuciones informales).

## 2. Principios del rediseño

1. **Cadena de proceso, no pipeline de fases.** Convención superpowers adaptada: brainstorming ≈ `/adversarial-research`, writing-plans ≈ `flow-plan`, executing-plans ≈ `flow-build` — pero con la capa de negocio ADELANTE (superpowers salta directo al diseño; aquí la entrada es la iteración de la idea).
2. **UN plan por unidad de trabajo.** La unidad puede ser un mock o un build fullstack tradicional — ambos recorren la cadena completa (idea → spec → plan → ejecución). No hay plan-por-fase.
3. **El plan nativo del harness NUNCA se sobreescribe.** Intención de planear → mecanismo nativo (plan mode en Claude Code, el propio de Codex/opencode). `flow-plan` captura y adopta; jamás corre una ceremonia paralela. (Ya medio construido: hook `flow-plan-capture` + ADOPT de flow-build.)
4. **Sugerencia discreta por intención, con ejecución nativa en los 3 harnesses.** Hook de sesión que inyecta protocolo + estado dinámico; sugiere, no fuerza (sin `EXTREMELY_IMPORTANT` ni hard-gates de optional reference project).
5. **Higiene por juicio, no por checklist.** La convención documental se sugiere; lo que se revisa con rigor es salud real: estado de git, archivos sueltos (¿preservar o no?), punteros rotos.

## 3. Modelo destino

```
                    ┌─ tipo de trabajo: mock ─────────┐
BRAINSTORM ──► SPEC ──► PLAN (nativo, capturado) ──► EJECUCIÓN
  │                 └─ tipo de trabajo: build fullstack ─┘
  └── preguntas disputadas → /adversarial-research

Transversales: wizard greenfield (arranque) · hygiene (salud) · report (rendering) · core (librería)
Promoción/deploy: convenciones de git (git-workflow.md) — sin skill
```

### 3.1 Mapa de las 11 skills actuales → destino

| Actual | Destino | Cambio |
|---|---|---|
| — | **`/flow-brainstorming`** (nueva) | Entrada real de la cadena. Ver §4. |
| `flow-specs` | `flow-specs` | Sin cambios — el diferenciador frente a optional reference project. Uso real vigente. |
| `flow-plan` | `flow-plan` | Se conserva (el split desde flow-build fue correcto: preservar planes). Rol: capturar/adoptar el plan nativo + routing table. Un plan por unidad de trabajo. |
| `flow-build` | `flow-build` | Se conserva como ejecutor de planes capturados (reconciler + verify gate). ADOPT es la alineación con el plan nativo. |
| `flow-intake` | **`/flow-start`** (wizard) | Fusión. Acepta notas crudas/conversación y PRODUCE el doc de intake (hoy exige documento escrito — causa evidenciada de bypass). |
| `flow-kickoff` | `/flow-start` | Fusión. Usada en su única oportunidad (06-24); bookend legítimo pero no amerita skill propia. |
| `flow-foundation` | `/flow-start` | Fusión. Diálogo repo por repo (mocks, specs, back, front): «¿cómo creamos este?» — como si fuera un plan. Con modo retrofit. Nunca corrió vía skill. |
| `flow-mock` | **disuelta como tipo de trabajo** | NO es retiro de capacidad: un mock recorre la cadena completa igual que un build tradicional (tradicional = movimientos fullstack), contextualizado en el momento. La skill de FASE desaparece; la práctica (repos navegables con feedback de cliente) es unidad de trabajo de primera clase. `/flow-start` absorbe la creación del repo de mocks. **Gate previo: confirmar con Ricardo** (corrió `/flow-mock review` 06-30). |
| `flow-deploy` | **retirada** | Revisar git basta para entender la promoción. Los gates (prod confirma, naming) ya viven en `git-workflow.md`/`devops-principles.md`. El QA walk in-vivo post-deploy migra al hook como sugerencia por intención («desplegué → ofrecer walk»). |
| `flow-hygiene` | `flow-hygiene` aligerada | Ver §6. |
| `flow-report` | `flow-report` | Se conserva; reposicionar descripción como skill general de reporting (ya opera así: dispara en proyectos sin ledger). |
| `flow-core` | `flow-core` | Se conserva (librería). Absorbe las plantillas del wizard. |

Resultado: **11 skills → 8** (`brainstorming`, `start`, `specs`, `plan`, `build`, `hygiene`, `report`, `core`).

## 4. `/flow-brainstorming` — contrato (nueva)

- **Trigger:** intención de explorar una feature o idea de negocio («¿y si el sistema hiciera…?», «el cliente quiere…», «¿vale la pena…?»). Sugerida por el hook; invocable directa.
- **Input:** la idea en cualquier forma (frase, notas, correo del cliente). Sin prerequisito documental.
- **Proceso:** iteración conversacional de la idea Y de lo necesario para llevarla a cabo — valor de negocio, alcance mínimo, impacto en contratos/repos, esfuerzo grueso, riesgos. En preguntas disputadas o decisiones consecuentes, invoca `/adversarial-research` como sub-herramienta.
- **Output:** decisión de negocio registrada (proceder / descartar / aplazar, con el porqué) — el input que `flow-specs` formaliza después. NO produce specs ni diseño técnico detallado: esa es la diferencia deliberada con el brainstorming de optional reference project, que salta al diseño.
- **Un artefacto ligero:** nota de decisión en el specs repo (`decisions/` o sesión), no un documento ceremonial.

## 5. Hook de sesión nativo — los 3 harnesses

### 5.1 Contrato común

Un solo bloque de protocolo (fuente única en `global/hooks/flow-session-context/`, replicado por build/deploy) + secciones dinámicas. Inyección al inicio de sesión; discreta (sugiere, no fuerza); silenciosa fuera de contexto útil.

**Bloque estático (borrador literal — inglés, es config Claude-facing):**

```
<flow-process-protocol>
Process chain for this workspace (flow pack v2) — suggest the matching stage when
user intent matches; never force it:
- BRAINSTORM: iterating a feature/business idea → ask ONCE (yes/no):
  "¿Iniciamos modo brainstorming?" — yes: invoke /flow-brainstorming; no: do not
  re-ask this session, manual invocation only. (Question-gate decidido 2026-07-17;
  enforcement: prompt-convention portada por este bloque del hook.) Contested
  questions route to /adversarial-research. Output: a business decision, not a spec.
- SPEC: a decided idea needs formalization → offer /flow-specs.
- PLAN: planning intent ALWAYS uses the harness's native plan mechanism; flow-plan
  captures/adopts the approved plan — never a parallel planning ceremony.
  ONE plan per unit of work (a mock is a work unit like any fullstack build).
- EXECUTE: a captured plan exists → offer /flow-build to execute it.
- After a deploy/promotion (git conventions own the flow): offer the in-vivo QA walk.
- New greenfield project → offer /flow-start. Workspace health concerns → /flow-hygiene.
</flow-process-protocol>
[dynamic] Ledger: current phase · next suggested (if flow workspace)
[dynamic] Git hygiene: merged branches · gone upstreams (if any)
```

**Regla de dosis:** el bloque completo solo en workspaces flow (walk-up a `_support/PROJECT.md`); fuera de ellos, nada (silencio — el pack no persigue repos ajenos al flujo).

### 5.2 Claude Code

- **Evento:** `SessionStart` (matcher `startup|clear`; excluir `resume|compact`).
- **Script:** `flow-session-context.sh` — stdin JSON (`cwd`, `session_id`) → gate de workspace → bloque estático + secciones dinámicas por stdout.
- **Absorbe `session-hygiene-context`** (mismo evento, cero pérdida): su lógica de ramas merged/`[gone]` pasa a ser la sección dinámica de git.
- Deploy: `/deploy-global` (settings-config.json idéntico al patrón actual).

### 5.3 Codex — nativo, sin plugin

Verificado en disco: Codex tiene un motor de hooks casi clon del de Claude Code (mismos eventos, mismo contrato stdin-JSON → stdout como additionalContext; ground truth: el plugin de Engram en `~/.codex/plugins/cache/engram/.../hooks/hooks.json` y sus scripts).

- **Registro:** `~/.codex/hooks.json` (a nivel usuario, hoy existe vacío) con el mismo schema:
  ```json
  { "hooks": { "SessionStart": [ { "matcher": "startup|resume|clear",
    "hooks": [ { "type": "command",
      "command": "\"$HOME/.codex/hooks/flow-session-context.sh\"", "timeout": 10 } ] } ] } }
  ```
- **El mismo script sirve** (mismo contrato stdin/stdout) — se despliega a `~/.codex/hooks/`.
- **Fricción conocida (aceptada):** `[hooks.state]` en `config.toml` guarda `trusted_hash` por hook — cada edición del comando exige re-trust; y el slot user-level `session_start:0:0` está hoy `enabled = false` → hay que habilitarlo una vez. El deploy documenta ambos pasos.

### 5.4 opencode — plugin local (no hay hooks de shell)

Verificado: opencode NO tiene evento SessionStart ni hooks de comando; el único camino de ejecución nativa es un plugin JS/TS. Patrón canónico ya operando en esta máquina: `~/.config/opencode/plugins/engram.ts` y `opencode-rules` (inyección vía `experimental.chat.system.transform` + gate once-per-session).

- **Entrega:** `~/.config/opencode/plugins/flow-session-context.ts` (plugin local: auto-carga, sin npm).
- **Mecánica (patrón elegido — estilo engram.ts):** `experimental.chat.system.transform`, appendeando al último entry de system (no `push` de un entry nuevo — el append preserva la compatibilidad de modelos que engram.ts ya resolvió); gate una-vez-por-sesión con un `Set<sessionID>` MÁS un guard de content-marker (si el bloque ya está presente en el system, no se re-appendea) para blindar contra doble inyección; el `Set` y el gate de workspace viven en un caché a nivel de módulo (no por proceso — un proceso opencode sirve N sesiones). Las secciones dinámicas (git hygiene) se computan en TS (child_process a git, mismo criterio que el script sh). Se evaluó el patrón de optional reference project.js (`experimental.chat.messages.transform`, inyectando en el primer mensaje de usuario) y se descartó por preferencia del usuario a favor del estilo engram.ts ya operando en esta máquina.
- **Limitación aceptada:** dispara en el primer turno del modelo, no en la apertura literal de la sesión — es el equivalente más cercano que la plataforma ofrece.

### 5.5 Fuente única y replicación

`global/hooks/flow-session-context/` contiene: el script sh (CC + Codex), el plugin ts (opencode), el bloque estático como fragmento compartido, y los settings/registro por harness. `harness/build.py` NO interviene (hooks no son skills); `/deploy-global` gana un paso 13c que despliega a los tres destinos y reporta el estado de trust/enabled en Codex.

## 6. Consolidación de hooks: 6 → 4

Sobre el análisis por-hook (evento, gating, dinámico-vs-estático, overlap con reglas always-on):

| Hook actual | Decisión | Fundamento |
|---|---|---|
| `session-hygiene-context` | **Se fusiona** en `flow-session-context` | Mismo evento SessionStart; su output es la sección dinámica de git. Cero pérdida. |
| `flow-phase-context` + `flow-plan-injector` | **Se fusionan entre sí** (un hook UserPromptSubmit) | Comparten evento y walk-up del ledger; el injector queda como sección gateada por `permission_mode == plan`. NO pueden ir a SessionStart: phase-context re-inyecta al cambiar el ledger a media sesión, y SessionStart no ve el plan mode. |
| `flow-plan-capture` | **Queda** (PostToolUse/ExitPlanMode) | Es la mitad construida del principio §2.3 — captura el plan nativo aprobado. |
| `delegation-reminder` | **Queda** | Contador dinámico de 20 llamadas con reset al delegar — backstop determinista que ninguna regla estática puede expresar. Solo su TEXTO restatea agent-routing.md; el mecanismo es irreemplazable. |
| `pre-push-lint-reminder` | **Queda** (recomendación; el usuario propuso colapsarlo) | Su texto restatea `Build & Lint`, pero su valor es el timing (dispara EN el `git push`), la supresión cuando lefthook ya cubre pre-push y la detección de comandos por repo — nada reproducible en SessionStart. Si aun así se decide retirarlo, lo que muere es el aviso just-in-time, no una redundancia. |

Resultado: `flow-session-context` (nuevo, absorbe 1) · `flow-context` (fusión de 2) · `flow-plan-capture` · `delegation-reminder` · `pre-push-lint-reminder` = **5 archivos hook, 4 unidades funcionales** (vs 6 actuales con 3 inyectores dispersos).

## 7. flow-hygiene aligerada

Del mapa de superficie (136 líneas: ~65 de enforcement rígido de convención, ~20 de juicio real):

**Se conserva y se vuelve el centro (los checks de juicio):**
- **Estado de git:** artefactos de sesión sin commitear (commit standing-authorized), ramas/worktree, `git mv` vs `mv`.
- **Archivos sueltos — ¿preservar o no?:** mover/expirar con «ante la duda, archivar»; planes `Status: planned` estancados (>~2 semanas) → adoptar/concluir/expirar; deletions solo tras confirmación tipeada.
- **Punteros rotos:** back-references de sesión, filas stale del ledger, detección de tracker con evidencia.

**Se aligera (de enforcement a sugerencia):**
- El vocabulario de conformance de sesiones (clasificación `move/promote/expire/...`, estructura iniciativa-vs-flat, renames ISO) deja de ser pass/fail: la convención se SUGIERE cuando aporta, no se audita como falta. «Una convención que no siempre aplicará» no genera hallazgos.
- El enforcement de formato de ledger/plantillas y las tablas de ID/manifest se reducen a lo mínimo operativo (un manifest simple de acciones propuestas).
- `migrate` se conserva como camino opt-in de adopción brownfield (su historial: onboardeó los 5 workspaces) pero re-escrito a principios, no a bootstrap estructural rígido.

## 8. Pendientes antes de aplicar (gates del rediseño)

1. **Ricardo:** ¿tiene el pack instalado? ¿qué invoca? (corrió `/flow-mock review` — su respuesta mueve mock/specs/foundation).
2. **Aprobación de este diseño** — incluye el texto literal del bloque §5.1 y los nombres nuevos (`/flow-brainstorming` — decidido 2026-07-17 —, `/flow-start`; este último abierto a mejor nombre).
3. **Plan de migración** (al aprobarse): (a) hook + consolidación de hooks; (b) `/flow-idea` + `/flow-start`; (c) aligerar hygiene; (d) retiros (deploy; mock tras hablar con Ricardo) con archivado en `_support/archived-agents/`-equivalente para skills; (e) `harness/build.py` + wrappers opencode + `openai.yaml` Codex (cambio de gates de invocación = cambio cross-harness); (f) revisión del flow-pack-manual.html y del `flow-pack-design.md` vivo.
4. **Medición post-cambio:** repetir el censo (con las lecciones del canon: ambos formatos Codex, excluir la sesión de investigación, artefactos como prueba, conciencia multi-operador) a ~4 semanas.
