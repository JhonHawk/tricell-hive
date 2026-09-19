# executor-dispatch-gate — backstop determinista al implementar sin plan: despacho de un ejecutor o edición inline de código

PreToolUse hook sobre la herramienta de subagentes (`Agent`; `Task` es el nombre anterior y
sigue en el matcher) y sobre las herramientas de edición de archivos (Claude `Write`/`Edit`/
`MultiEdit`/`NotebookEdit`; Grok `search_replace`/`write_file`/`create_file`/`edit_file`),
registrado globalmente vía `settings-config.json`. Solo actúa en un **flow workspace** (un
`_support/PROJECT.md` por encima del cwd) y solo ante uno de dos actos: el despacho de un
**ejecutor** (subagente cuyo `tools:` lleva Write o Edit) o la **edición inline de código de
proyecto** desde el hilo principal. Fuera de eso, silencio.

## Qué resuelve

Incidente (2026-09-13, sesión de Claude Code en un workspace de cliente con ledger): el hilo
principal despachó `backend-developer` y `sdd-spec-writer` y movió tickets a In Progress sin
que existiera un plan aprobado. Había leído la respuesta a la pregunta del modo git y una
decisión de diseño como si fueran la aprobación de implementar. Las reglas ya lo prohíben
(`CLAUDE.md > Communication`: un acuerdo a media conversación no autoriza implementar; el
contrato de aprobación de `/flow-plan`) y son prompt-convention: fallaron. Este hook dispara
**en el momento del acto** — el despacho del ejecutor — que es el punto observable donde
"analizar" se convierte en "implementar".

Segundo incidente (2026-09-17, sesión Grok en ark, ~21 h y 3 compactaciones): tras una
conversación de diseño el usuario dijo "necesito validaciones serias en ese input" y el hilo
principal implementó y commiteó el ticket **inline**, sin despachar a nadie — el camino que el
hook no veía. Desde entonces gatea también la edición inline, con el mismo criterio y el
mismo ámbito.

## Comportamiento

1. Lee `subagent_type` (Claude: `tool_input.subagent_type`; Grok: `toolInput.subagentType`).
   Con valor: no ejecutor → exit 0 sin salida. Vacío: si la herramienta es de edición, lee
   `file_path` (o `notebook_path`/`path`/`filePath`); otra herramienta o sin ruta → exit 0.
1b. **Solo código de proyecto (ruta inline).** La ruta debe estar bajo la raíz del workspace
   y no ser un registro: fuera quedan `_support/**`, cualquier `*-specs/`, `docs/`, los
   directorios de harness (`.claude/`, `.codex/`, `.grok/`, `.agents/`, `.engram/`), las
   extensiones de texto (`md`, `markdown`, `mdx`, `txt`, `rst`, `adoc`, `csv`, `tsv`, `log`)
   y `LICENSE*`/`CHANGELOG*`/`NOTICE*`. Así el ledger, los planes, el registro de tareas de la
   ruta directa, la memoria y los README nunca disparan: el hook no puede entrar en bucle
   sobre los archivos donde se anota la decisión que lo desbloquea. Todo lo demás bajo la
   raíz es código (tests incluidos: el RED de TDD ya es implementar).
2. Walk-up desde `cwd` buscando `_support/PROJECT.md`. Sin ledger → exit 0 sin salida (no es
   flow workspace; la ruta directa no se toca).
3. Resuelve `sessions_home` como `flow-context.sh` (`*-specs/sessions` bajo la raíz, si no
   `_support/sessions`) y corre `plan.py inspect` sobre cada `*-plan.md` (maxdepth 2).
   Si alguno devuelve `can_implement: true` → exit 0 sin salida: hay autoridad.
   Un plan malformado o con una autorización que el validador rechaza no autoriza, y el
   mensaje lo nombra con el código del validador (`<ruta> (<code>)`; `.error.code` para el
   documento, `.authorization.error.code` para la autorización) — un plan en ejecución que una escritura dejó inválido no se reporta
   como "no hay plan". Si `plan.py` no está instalado en
   `~/.claude/skills/flow-core/scripts/`, aproxima con `Status: planned|building` y lo dice en
   el mensaje (`HIVE_PLAN_PY` sobreescribe la ruta; lo usa el test).
4. Sin autoridad, lee la fila `| Executor dispatch | <valor> |` de la tabla de cabecera del
   ledger (solo cuenta el token inicial del valor; el resto puede ser comentario):
   - `plan-required` → **deny**: exit 2, la razón en stderr y `{"decision":"deny","reason":…}`
     en stdout (la forma de `bash-policy.sh`). Se repite en cada despacho y en cada edición
     inline de código; las ediciones de registros (1b) siguen pasando, que es lo que permite
     escribir el plan que lo destraba.
   - cualquier otro valor o fila ausente → **advisory** vía
     `hookSpecificOutput.additionalContext`, **una vez por
     sesión y por firma del estado** (marker en `${TMPDIR:-/tmp}` con el cksum de la lista
     de planes + Status + can_implement, como `flow-context.sh`); un plan nuevo o un cambio de
     Status vuelve a disparar. Sin `session_id` dispara siempre.
5. Cualquier error interno → exit 0 sin salida. El hook es backstop: no puede romper un
   despacho por un bug propio. Timeout 15 s.

El mensaje nombra cuántos planes revisó y dónde, qué agente se despacha (o qué archivo se
edita inline), qué autoriza
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

- **Un `Bash` que escribe código (`sed -i`, heredoc) no pasa por aquí.** Ese camino es de
  `bash-policy`; el hook cubre las herramientas de edición, que es donde ocurrieron los dos
  incidentes.
- **El ámbito "código de proyecto" es una lista de exclusiones, no un detector.** Un formato
  de registro que no esté en la lista (1b) se gatea como código; añadirlo es un cambio de
  una línea con su caso de test.
- **`plan-required` deniega cada edición de código sin plan.** Es el efecto buscado en
  workspaces donde TODO pasa por `/flow-plan`; el mensaje nombra la salida (escribir el plan,
  o el usuario edita la declaración). Sigue siendo opt-in por ledger.
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
`sessions_home` en el specs repo; roster vs disco; cada deny nombra su continuación; ruta
inline (casos l): código → advisory/deny según el ledger, y silencio para `.md`/`.txt`,
`_support/`, ledger, specs repo, `.claude/`, `docs/`, fuera de la raíz, sin ledger, herramienta
de lectura, con autoridad; payload Grok `search_replace`; una edición y un despacho en el
mismo estado comparten el marker; el mensaje nombra el archivo.
Mantenerlo en verde; `shellcheck -S style` sobre ambos `.sh` en cero hallazgos.
