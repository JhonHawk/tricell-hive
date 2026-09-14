# executor-dispatch-gate — backstop determinista al despachar un agente ejecutor sin plan

PreToolUse hook sobre la herramienta de subagentes (`Agent`; `Task` es el nombre anterior y
sigue en el matcher), registrado globalmente vía `settings-config.json`. Solo actúa en un
**flow workspace** (un `_support/PROJECT.md` por encima del cwd) y solo cuando el subagente
despachado es un **ejecutor** (su `tools:` lleva Write o Edit). Fuera de eso, silencio.

## Qué resuelve

Incidente (2026-09-13, sesión de Claude Code en un workspace de cliente con ledger): el hilo
principal despachó `backend-developer` y `sdd-spec-writer` y movió tickets a In Progress sin
que existiera un plan aprobado. Había leído la respuesta a la pregunta del modo git y una
decisión de diseño como si fueran la aprobación de implementar. Las reglas ya lo prohíben
(`CLAUDE.md > Communication`: un acuerdo a media conversación no autoriza implementar; el
contrato de aprobación de `/flow-plan`) y son prompt-convention: fallaron. Este hook dispara
**en el momento del acto** — el despacho del ejecutor — que es el punto observable donde
"analizar" se convierte en "implementar".

## Comportamiento

1. Lee `subagent_type` (Claude: `tool_input.subagent_type`; Grok: `toolInput.subagentType`).
   Vacío o no ejecutor → exit 0 sin salida.
2. Walk-up desde `cwd` buscando `_support/PROJECT.md`. Sin ledger → exit 0 sin salida (no es
   flow workspace; la ruta directa no se toca).
3. Resuelve `sessions_home` como `flow-context.sh` (`*-specs/sessions` bajo la raíz, si no
   `_support/sessions`) y corre `plan.py inspect` sobre cada `*-plan.md` (maxdepth 2).
   Si alguno devuelve `can_implement: true` → exit 0 sin salida: hay autoridad.
   Un plan malformado no autoriza. Si `plan.py` no está instalado en
   `~/.claude/skills/flow-core/scripts/`, aproxima con `Status: planned|building` y lo dice en
   el mensaje (`HIVE_PLAN_PY` sobreescribe la ruta; lo usa el test).
4. Sin autoridad, lee la fila `| Executor dispatch | <valor> |` de la tabla de cabecera del
   ledger (solo cuenta el token inicial del valor; el resto puede ser comentario):
   - `plan-required` → **deny**: exit 2, la razón en stderr y `{"decision":"deny","reason":…}`
     en stdout (la forma de `bash-policy.sh`). Se repite en cada despacho.
   - cualquier otro valor o fila ausente → **advisory** vía
     `hookSpecificOutput.additionalContext` (la forma de `rule-context.sh`), **una vez por
     sesión y por firma del estado** (marker en `${TMPDIR:-/tmp}` con el cksum de la lista
     de planes + Status + can_implement, como `flow-context.sh`); un plan nuevo o un cambio de
     Status vuelve a disparar. Sin `session_id` dispara siempre.
5. Cualquier error interno → exit 0 sin salida. El hook es backstop: no puede romper un
   despacho por un bug propio. Timeout 15 s.

El mensaje nombra cuántos planes revisó y dónde, qué agente se despacha, qué autoriza
(un `/flow-plan` aprobado con grant de implementación, o el verbo explícito del usuario en
ESTA conversación), qué no autoriza (una respuesta de modo git, una decisión de diseño, un
"los atacaremos") y la continuación: citar el verbo en el prompt del despacho, u ofrecer
`/flow-plan` / preguntar "¿lo aplico?".

## Roster de ejecutores

Literal en el script (`readonly EXECUTORS=…`): los agentes de `global/agents/**/*.md` cuyo
`tools:` lista Write o Edit, **menos** la exclusión declarada (`readonly EXCLUDED=…`):
`sdd-spec-writer`, `sdd-design`, `cloud-architect`, `solution-architect`. Esos escriben los
artefactos que el gate del plan consume (specs, ADRs, contratos, diseño de infra) y por diseño
se despachan antes de que exista un plan (spec-writing playbook → business gate →
`/flow-plan`); gatearlos bloquearía el camino que produce el plan. El test (caso h)
re-deriva "roster con Write/Edit menos la exclusión" desde disco y falla ante cualquier
desviación — incluida una exclusión que ya no nombra a un agente con Write/Edit —, así que un
agente nuevo o un cambio de `tools:` se detecta en test, no en producción. Quedan fuera
también los de solo lectura y los que ejecutan para observar: `sdd-explore`, `review-*`,
`sdd-product-critic`, `sdd-spec-reviewer`, `state-fetcher`, `workspace-custodian`,
`sdd-verify`, `review-ux`, y los built-in (`Explore`, `Plan`, `general-purpose`).

## Capa de enforcement

- **Disparo: determinista.** El hook corre en cada despacho de subagente; leer el ledger y
  los planes no depende del modelo.
- **Efecto advisory (default): prompt-convention.** `additionalContext` entra al contexto
  del modelo; obedecerlo sigue siendo su decisión. Es un STOP en el momento correcto, no un
  candado.
- **Efecto deny (`plan-required`): determinista, opt-in por ledger.** Con la fila declarada,
  el despacho no ocurre. Lo activa el usuario editando el ledger, nunca el agente.

## Límites conocidos

- **No ve ediciones inline del hilo principal.** Fuera de un plan, el hilo principal edita
  directo por diseño (medido en sesiones reales); `Write`/`Edit` no pasan por este hook. Cubre
  el despacho de ejecutores, que es donde ocurrió el incidente.
- **`plan-required` empuja al modelo a editar inline si no hay plan.** Cerrar el despacho
  sin cerrar la edición directa desplaza el trabajo al camino que el hook no ve. Por eso el
  deny es opt-in: sirve en workspaces donde el usuario quiere que TODO pase por `/flow-plan`
  y acepta ese desplazamiento como señal visible, no como default.
- **Claude Code y Grok solamente.** Grok merge `~/.claude/settings.json` por compat y envía
  el payload en camelCase (cubierto). Codex, opencode y PI no leen `settings.json`: allí la
  regla sigue siendo prompt-convention.
- **La autoridad es del plan, no de la conversación.** Si el usuario dio el verbo explícito
  sin plan, el advisory igual dispara una vez; la continuación correcta es citar el verbo en
  el prompt del despacho y seguir. El hook no lee el transcript.

## Tests

`bash global/hooks/executor-dispatch-gate/test-executor-dispatch-gate.sh` — fixtures en un
tmpdir (ledger, planes renderizados con `test_plan.render_plan` del validador, `TMPDIR`
aislado para los markers). Casos: sin ledger; ledger sin planes; plan con autoridad; planes
draft/sin grant; `plan-required` deny (repetible, y cede ante un plan con autoridad); valor
default de la plantilla (`free …`) → advisory; agentes no ejecutores y payload sin
`subagent_type`; payload Grok camelCase; marker (misma sesión y estado → silencio; cambio de
estado → re-dispara; sin `session_id` → siempre); fallback sin `plan.py`; plan malformado;
`sessions_home` en el specs repo; roster vs disco; cada deny nombra su continuación.
Mantenerlo en verde; `shellcheck -S style` sobre ambos `.sh` en cero hallazgos.
