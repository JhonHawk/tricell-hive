# Flow Pack — Contrato de transición entre fases

> **Estado:** diseño propuesto, pendiente de aprobación. Extiende `flow-pack-design.md` (2026-06-11).
> **Fecha:** 2026-06-14.
> **Fuentes:** sesión de diseño (este spec la consolida); `flow-pack-design.md`; bodies de las skills `flow-*` y `flow-core/references/{handoff-protocol,harness-mechanics}.md`.
> **Fundamento externo:** Stage-Gate de Cooper (gates proporcionales); patrón Plan-then-Execute en agentes (CHI 2025, arXiv 2509.08646); Definition of Ready / Definition of Done + artifact-as-handoff (Atlassian, Scrum Alliance, DevOps artifacts). Citas al pie.

---

## 1. Problema

`flow-pack-design.md` resolvió la **orquestación dentro de cada fase** (P1–P5: delegación determinística, ledger, naming, CD-first). Lo que NO resolvió es la **frontera entre fases**. El pack tiene fases (F1–F7) pero no tiene contratos de transición formales entre ellas. Tres síntomas observados, una causa raíz.

| # | Síntoma observado | Causa raíz |
|---|---|---|
| T1 | `flow-mock build` se lanza a editar/construir un repo entero epic por epic sin presentar un plan aprobable. flow-dev y flow-deploy sí tienen gate; mock no. | El gate de plan está calibrado de forma **implícita e inconsistente**, no por un criterio explícito. |
| T2 | Tras cerrar una fase, "¿cómo disparo la siguiente?" no tiene respuesta clara. Ej.: actualizado el mock, flow-dev lee specs pero **el mock no es input formal** suyo — el `handoff-protocol` ejemplifica con specs/contracts, nunca con el mock repo. | El `PROJECT.md` es un **ledger de estado**, no un handoff. Condensa y pierde el detalle que la siguiente fase necesita consumir. |
| T3 | No hay verbo para "revisar un spec ya existente cuyo alcance cambió por feedback del stakeholder" (incluso post-implementación). Hoy se fuerza por `review` o conversacionalmente. | El pack modela **fases lineales hacia adelante**, no re-entradas. |

**Causa raíz común:** las fases existen, pero las *fronteras entre fases* no son objetos de primera clase. No hay un contrato que declare qué produce una fase al cerrar (su salida) ni qué precondiciones verifica la siguiente al abrir (su entrada).

## 2. Fundamento

Los tres síntomas son problemas resueltos en metodología establecida:

1. **Gates proporcionales (Stage-Gate, Cooper).** Fases con puntos de control, donde el gate evalúa contra criterios *proporcionales al riesgo*. La propia literatura reconoce que el full-gate "entail too much management process… some companies favor a lightweight process with fewer formal deliverables" [1][2]. → El gate de plan no debe ser uniforme: debe **escalar al costo/irreversibilidad de la fase**.

2. **Plan-then-Execute (patrón agéntico).** Hoy es un design pattern de primera clase: "humans can approve the plan before execution… routine tasks run autonomously while specific decision categories escalate to humans for decisions that are hard to reverse or carry accountability" [3][4]. → Es lo que flow-dev ya hace y mock-build no. La asimetría es defecto, no diseño.

3. **Artifact-as-handoff + DoR/DoD.** En un pipeline, "artifacts serve as the 'handoff' between stages… proof of work, verifying the integrated code is ready for the next stage" [5]. El par DoR/DoD formaliza la frontera: *Definition of Done* = qué produce la fase que cierra; *Definition of Ready* = qué verifica la fase que abre [6][7]. → El handoff que falta NO es el ledger; es un artefacto de transición que el ledger no modela.

## 3. Decisión central — el contrato de transición

Cada fase del pack adquiere dos cláusulas explícitas en su contrato (definidas en `flow-core`, no repetidas por skill):

- **DoD de salida (Definition of Done):** al CLOSE, la fase produce un **handoff** — qué se construyó (con paths), qué decisiones quedaron, qué cambios disparó hacia atrás, y la **siguiente fase sugerida según el estado del ledger**.
- **DoR de entrada (Definition of Ready):** al OPEN, la fase verifica sus precondiciones y **consume el handoff de la fase anterior** como bounded context, no solo el ledger condensado.

El handoff es el objeto que hoy no existe. Resuelve T2 directamente y habilita T1 y T3.

### 3.1 Dónde vive el handoff — sección en PROJECT.md con poda

**Decisión:** sección estructurada dentro de `PROJECT.md`, no archivo dedicado.

| | Sección en PROJECT.md (elegida) | Archivo dedicado `HANDOFF.md` (descartado) |
|---|---|---|
| Dispersión | ✅ Un archivo; el DoR ya lee PROJECT.md en OPEN | ❌ Más punteros = más superficie de "broken pointer" (lo que flow-hygiene repara) |
| Engram mirror | ✅ El ledger ya se mirror-ea con `topic_key`; viaja gratis | ❌ Requiere su propio mirror o queda fuera del resume |
| Separación responsabilidades | ❌ Mezcla estado durable con handoff transitorio | ✅ Estado y handoff separados |
| Crecimiento / ruido | ❌ Riesgo de inflar el ledger (1 handoff × fase × N epics) | ✅ Diffs limpios |

**Mitigación de la única desventaja real** (meter algo transitorio en un archivo durable): **política de poda**. Solo el *handoff vigente* (última fase cerrada) vive expandido en `## Current handoff`. Cuando la siguiente fase lo consume en su OPEN, lo **colapsa a una línea** en el historial del ledger. El detalle muerto no se acumula; el handoff se trata como "estado vigente que expira", que es lo que es.

```markdown
## Current handoff
**Fase cerrada:** F4 mock · **epic:** FIN-B0x · **fecha:** 2026-06-14
- Construido: CashierPage.tsx, ConciliacionView.tsx (rutas en mocks repo README)
- Spec-changes disparados: adenda STP/Conekta → FIN-B0x revisado
- Decisiones UX: layout de conciliación E4 pendiente (BLOCKED, externo)
- Mock repo (bounded context para F6): acme-next-frontend-mocks
- Siguiente sugerida: /flow-dev FIN-B0x  (specs reviewed ✓, mock reviewed ✓)

## Handoff history
- F3 specs → F4 (consumido 2026-06-14): specs FIN-B0x reviewed
- F2 kickoff → F3 (consumido 2026-06-10): workspace + ledger creados
```

### 3.2 Sub-decisiones derivadas

**D1 — Gate de plan proporcional al riesgo (resuelve T1).**
El criterio explícito, fundado en Cooper [1][2] y Plan-then-Execute [3]: *gate ∝ costo de deshacer la escritura de la fase.*

| Fase | Gate de plan | Razón |
|---|---|---|
| flow-dev, flow-deploy | **Fuerte** (ya existe) | Tocan producción / código real; irreversible |
| **flow-mock build** | **Ligero** (a añadir) | Construye un repo entero; reconstruir cuesta. Plan = epics/pantallas/stack/orden |
| flow-specs epic/review | Ya tiene su review gate | El draft re-pasa por review de todas formas |
| flow-kickoff, flow-specs init | **Sin gate** | Bootstrap determinista; el plan no aporta |

El gate ligero de mock-build es un resumen aprobable de 5 líneas (qué epics, qué pantallas con sus estados empty/loading/error, qué stack, qué orden), no un plan completo estilo flow-dev.

**D2 — `flow-specs epic` se reencuadra como "draft *or revise*" (resuelve T3).**
No se añade verbo nuevo. Razones: (a) flow-mock review *ya invoca* `/flow-specs epic <name>` para spec-changes — el carril ya se usa así; (b) un epic que cambia de alcance debe re-pasar por el review gate que ya existe. Un cambio post-implementación entra por el mismo carril; el handoff de flow-dev marca qué tareas ya implementadas quedan afectadas (requiere que flow-dev lea "tarea cambió después de implementada" del handoff — otra razón para formalizar el artefacto).

**D3 — Paso EXPLORE proporcional en el contrato de flow-core.**
La traducción multi-harness ya existe (`harness-mechanics.md`: "dispatch Explore" → Codex read-only sandbox / opencode read-only permissions). Lo que falta es que el **contrato declare un paso EXPLORE explícito** antes de ROUTE/ORCHESTRATE, *cuando la fase lo necesite* — no repetido en cada skill (sería el anti-patrón de duplicación que AGENTS.md prohíbe). Proporcional: kickoff no explora; specs review y dev sí. Esto asegura que harnesses con Explore más débil que Claude Code activen su equivalente por el mismo contrato.

**D4 — Recomendación de siguiente fase contextual (resuelve la otra mitad de T2).**
Parte del DoD (§3): el handoff incluye "siguiente sugerida" derivada del **estado del ledger** (qué fases están `done` para este epic), no un sucesor F+1 fijo. Desde specs-review la siguiente puede ser mock, *o* otro epic, *o* dev si el mock ya existe.

## 4. Cambios por archivo

| Archivo | Cambio | Sub-decisión |
|---|---|---|
| `flow-core/SKILL.md` | Añadir cláusulas DoD/DoR al contrato (OPEN consume handoff; CLOSE produce handoff); añadir paso EXPLORE proporcional; documentar criterio de gate ∝ riesgo | §3, D1, D3, D4 |
| `flow-core/references/ledger-template.md` | Añadir secciones `## Current handoff` y `## Handoff history` con política de poda | §3.1 |
| `flow-core/references/handoff-protocol.md` | El bounded-context de un dispatch incluye el mock repo cuando la fase lo tenga | §3, T2 |
| `flow-mock/SKILL.md` | `build` abre con plan ligero aprobable antes de construir | D1 |
| `flow-specs/SKILL.md` | `epic` se reencuadra como draft-or-revise; descripción y pasos | D2 |
| `harness/build.py` (run) | Regenerar árboles tras editar skills — el contrato vive en la skill `flow-core` y se propaga a Codex/opencode por aquí | — |

> **Corrección al inspeccionar:** `harness/AGENTS.md` NO requiere edición manual. La regla de "reflejar a mano en el AGENTS condensado" aplica a cambios en `global/rules/`; el contrato de transición vive en la **skill** flow-core, que se regenera a `agents-skills/` vía `build.py`. Restatear los pasos del contrato en el AGENTS condensado sería duplicación (su sección "Flow Phase Boundaries" es guía conversacional de límites de fase, no los pasos OPEN/CLOSE).

## 5. Plan de implementación (secuencia propuesta)

El contrato de transición es la pieza de la que cuelgan las demás → primero.

1. **Contrato DoD/DoR + handoff** en `flow-core` + plantilla de ledger con poda. (§3, §3.1)
2. **Gate de plan proporcional** — añadir a flow-mock build; documentar criterio en flow-core. (D1)
3. **Reencuadrar `flow-specs epic`** como draft-or-revise. (D2)
4. **Paso EXPLORE proporcional** en el contrato. (D3)
5. Propagar a `harness/AGENTS.md` y correr `build.py`; commit del árbol regenerado.

Cada paso es revisable de forma aislada. El #1 desbloquea #2–#4.

## 6. Riesgos y trade-offs

- **Inflado del ledger (T2 mitigado, no eliminado).** La poda depende de que cada fase colapse el handoff anterior en su OPEN. Si una fase olvida podar, el ledger crece. Mitigación: flow-hygiene audita handoffs sin colapsar como un finding.
- **Sobre-ceremonia.** Añadir DoD/DoR a fases ligeras (kickoff) sería fricción sin valor — por eso el gate y el EXPLORE son **proporcionales**, no uniformes. El riesgo es que "proporcional" se interprete laxo; el criterio escrito (gate ∝ costo de deshacer) lo acota.
- **Reencuadre de `epic` (D2).** "draft or revise" en un solo verbo puede confundir si el usuario espera un verbo `update`. Mitigación: la descripción debe ser explícita y el caso revise debe forzar el review gate.
- **Multi-harness drift.** D3/D4 viven en reglas que no pasan por `build.py` → deben replicarse a mano en `harness/AGENTS.md`. Riesgo de desincronización (ya conocido del repo).

## 7. Decisiones abiertas

- ¿El handoff history se conserva indefinidamente o flow-hygiene lo archiva tras N fases? (Propuesta: conservar — son una línea por fase, costo despreciable.)
- ¿El plan ligero de mock-build usa el gate nativo de plan mode o AskUserQuestion? (Propuesta: mismo patrón dual que flow-dev — detecta el modo.)
- ¿`flow-specs revise` merece ser visible como sub-comando alias de `epic` para descubribilidad, aunque internamente sea el mismo carril?

---

## Referencias

1. Stage-Gate International — *The Stage-Gate Model: An Overview*. <https://www.stage-gate.com/blog/the-stage-gate-model-an-overview/>
2. TCGen — *What is the Stage-Gate Process? Pros and Cons*. <https://www.tcgen.com/product-development/stage-gate-process/>
3. *Plan-Then-Execute: An Empirical Study of User Trust and Team Performance When Using LLM Agents* — CHI 2025, ACM. <https://dl.acm.org/doi/10.1145/3706598.3713218>
4. *Architecting Resilient LLM Agents: A Guide to Secure Plan-then-Execute Implementations* — arXiv 2509.08646. <https://arxiv.org/pdf/2509.08646>
5. Graph AI — *Artifacts: Definition, Examples, and Applications (DevOps)*. <https://www.graphapp.ai/engineering-glossary/devops/artifacts>
6. Atlassian — *What is the Definition of Done (DoD) in Agile?*. <https://www.atlassian.com/agile/project-management/definition-of-done>
7. Scrum Alliance — *Definition of Ready vs. Definition of Done*. <https://resources.scrumalliance.org/Article/definition-vs-ready>
