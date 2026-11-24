# codex-default-question-channel — Canal de preguntas en el modo Default de Codex

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | ninguno · watch `watch/codex-decision-format-async` (Engram #17780) |
| Git | tricell-hive · `development` en `15ac9aa5` · rama `fix/codex-default-question-channel` en `.claude/worktrees/codex-question-channel` |
| Playbook | feature |

## Decisiones

2026-11-24
- **D1 A** — ¿Cómo seguimos tras revisar la sesión Codex de project-a `01a20c02`? Registrar la observación en los watch items y planear la corrección.
- **D4 A** — ¿Qué enfoque toma la corrección? Una línea acotada en la regla de preguntas (sin `request_user_input_async` cuando Codex ofrece `request_user_input`; entrega, revisión y validación son rondas, no permisos), con los criterios del auditor primero. Se descartan «nunca async» en `global.md` (B) y solo auditor (C).
- **Entrega A** — `automatic`. **Dónde A** — worktree.
- **Supuesto** — Una frase que recomienda un texto entre backticks que contiene un modo de entrega como palabra entera («Recomiendo `automatic, Codex, sí`») da `fail` cuando no sigue una tarjeta en su turno. Una frase que solo deja pendiente una decisión de sesión («queda por confirmar la entrega») da `review`. Cada frase se juzga por separado. Descartado: `fail` también para la frase que solo deja algo pendiente, porque marcaría como fallo un informe que menciona una entrega pendiente como hecho.
- **Supuesto** — El auditor no exige la respuesta de compactación del bloque de siguiente paso. Descartado: un `review` adicional; se agrega si vuelve a faltar en otra sesión.
- **Supuesto** — El `$flow-x` que el usuario escribe en Codex pasa a contar como carga de esa skill para todos los criterios que leen cargas, también `flow_entered_without_route`, `analysis_reply_block`, `try_it_block_in_report` y la activación desatendida de `prose_choice_without_card` y `text_round_without_card`; `review_rounds_per_candidate` solo juzga sesiones de Claude Code y no cambia. Descartado: verlo solo en `plan_without_build_route`, que dejaría a los demás ciegos ante la misma carga.
- **Supuesto** — `content/guidance/global.md` no cambia (D4 A). Descartado: repetir la línea allí.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

«no se si lo mas facil sea pedir que nunca se use el modo async en una regla global para codex» (2026-11-24), tras la revisión de la sesión Codex `01a20c02-5d3c-76b2-ad76-f221fd88dd04` (project-a, `gpt-6.2-sol` high, modo Default, 2026-11-24 05:24–05:42Z), que mostró antes del build:

- L200: tarjeta D2 con opciones sin letra («Clave firmada (Recommended)»), cerrada por tiempo.
- L236: la ronda en texto sin `Situation:`, «Si la eliges:», «Recomiendo» ni línea de espera.
- L408: entrega, revisión y seguridad reducidas a «Recomiendo **`automatic, Codex, sí`**», sin tarjeta.
- L421: `request_user_input_async` con opciones planas y sin la ronda en el texto, con `request_user_input` ofrecido.
- L496: «**Siguiente paso · Implementar PRJ-933**» en lugar del bloque `▶`.

El auditor marca L200 y L236 y no ve L408, L421 ni L496.

## Objetivo

Hoy la regla de preguntas trata la tarjeta async como «sin tarjeta», pero no prohíbe usarla cuando Codex ofrece la que espera, ni resuelve el choque con la plantilla del modo Default («la entrada obligatoria en una pregunta de texto; nunca opción múltiple en texto»). El auditor no detecta la entrega pendiente en prosa, la ronda async sin texto ni el bloque de siguiente paso ausente. Al terminar, el auditor detecta los tres casos en esa sesión y la regla dice qué hacer en Codex Default.

## Alcance y aceptación

Dentro: criterios del auditor en `tests/pilot` con sus fixtures; la línea del canal en `content/skills/hive-core/messages/questions.md`; la página `site/src/content/docs/host-limits.md`; `_support/docs/architecture/agent-delivery.md` (nota fechada, la fila Codex de «Where the options go» y la línea de 2026-11-17 marcada como superada); las frases que limitan la tarjeta a «las preguntas que Codex permite» en `README.md`, `llms.txt` y `_support/docs/architecture/tool-dependencies.md`; las filas de la rúbrica y de `rule-enforcers.md`; el delta del spec `workflow-lifecycle`. Fuera: `global.md`, `report-endings.md`, `unanswered.md`, la respuesta de compactación, casos nuevos en `tests/fixtures/regression/`, pilotos.

- AC1 — `prose_choice_without_card` da `fail` en la línea 408 de la sesión `01a20c02` · false today because da `not_observed` (2026-11-24) · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout real `~/.codex/sessions/2026/11/23/rollout-2026-11-23T23-23-31-01a20c02-5d3c-76b2-ad76-f221fd88dd04.jsonl`
- AC2 — Un criterio nuevo, `async_card_when_waiting_offered`, da `fail` en la línea 421 de esa sesión · false today because el criterio no existe y nada marca esa línea · Observed in: el mismo comando sobre el mismo rollout
- AC3 — `plan_without_build_route` da `fail` en la línea 496 de esa sesión, que trae `$flow-build` sin el bloque `▶` · false today because da `not_observed`: el `$flow-plan` tecleado no se ve como carga · Observed in: el mismo comando sobre el mismo rollout
- AC4 — La regla de preguntas instalada dice dos cosas: la tarjeta async es el canal solo cuando el modo no ofrece `request_user_input`; y Hive pregunta por la tarjeta la entrega, la revisión, la validación y toda decisión aunque la plantilla de Codex mande la entrada obligatoria al texto, dejando para esa plantilla solo las escaladas de sandbox o aprobación del host · false today because la fila «No waiting card» de `content/skills/hive-core/messages/questions.md` nombra `request_user_input_async` sin condición, y ninguna línea nombra la entrada obligatoria de la plantilla · Observed in: `~/.agents/skills/hive-core/messages/questions.md` y los paquetes de etapa instalados tras `hive update`
- AC5 — La sección Codex de la página «Host limits» del sitio describe que la tarjeta async no se usa mientras se ofrece `request_user_input` · false today because la sección menciona la herramienta asíncrona (`host-limits.md:80`) solo para no duplicar preguntas · Observed in: la página construida por `pnpm build` en `site/`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- La regla actual: `content/skills/hive-core/messages/questions.md` «Channel» (6.465 bytes), filas 60 y 62; spec `workflow-lifecycle` «Decision rounds go through the card alone» (línea 1133), «The session audit reads an imperative confirmation and quoted replies as prose choices» (1877), «The session audit fails a plan that ends without a build route» (1833), «The session audit flags card options out of format» (1761).
- La sesión `01a20c02` leyó las reglas de preguntas antes de cada tarjeta (paquetes de `hive-mode`, L25, y `flow-plan`, L163); Codex ofreció `request_user_input` y `request_user_input_async` (L3); modo Default con la plantilla que reserva la tarjeta para preguntas opcionales (L10).
- Causas en el auditor (investigación del hijo `hive-research`, 2026-11-24): `proseChoices` (`tests/pilot/session_audit_prose.go:45`) no lee «queda por confirmar» ni un token entre backticks con varias palabras; `session_audit.go:345-346` mapea las dos herramientas a `Kind:"question"`, y `cardOptionFormat` salta la async a propósito (`session_audit_assessment.go:247`); `codexCompletedItem` (`session_audit.go:374`) ignora el ítem `UserMessage` con contenido `skill`, y `planWithoutBuildRoute` (`session_audit_routes.go:265`) aprueba con solo el comando.
- `go test -count=1 ./tests/pilot/` · ok, 63,6 s · 2026-11-24
- `go run ./tests/pilot --audit-session <rollout 01a20c02> --host codex` · `card_option_format` fail L200, `decision_round_format` fail L236, `prose_choice_without_card` y `plan_without_build_route` not_observed · 2026-11-24
- La comprobación de texto de `AGENTS.md` «Checks by path» (cuatro comandos) · ok, 92 s · 2026-11-24
- `pnpm check`, `pnpm lint` y `pnpm build` en `site/` · ok, 20 páginas · 2026-11-24

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic`: revisión, push, PR a `development`, CI y merge sin parar; después, la actualización de la instalación local que `AGENTS.md` autoriza en esta copia de mantenedor. Revisión de código: `/code-review` nativo de Claude Code, que `AGENTS.md` prefiere. Sin revisión de seguridad: no toca un límite de seguridad. Esta respuesta no autoriza una release.

## Estado de revisión y progreso

2026-11-24 — Plan escrito.
2026-11-24 — Revisión del plan: dos `hive-review-plan`, uno del auditor (servidor) y otro de guía. Ninguno encontró P0 ni P1. Cambió:
- AC3 da su fail en la línea del mensaje (496).
- AC4 dice la regla sin el argumento de «no son permisos» y deja la plantilla para las escaladas del host.
- AC5 cambia su razón.
- El fail de T1 exige el modo de entrega como palabra entera.
- T2 juzga la oferta de la tarjeta en toda la sesión.
- T3 corre el auditor sobre rollouts reales antes y después. `typedReview` no cambia, porque `review_rounds_per_candidate` solo juzga sesiones de Claude Code.
- T4 suma `README.md`, `llms.txt`, `tool-dependencies.md` y las líneas 283 y 292 de `agent-delivery.md`.
- T5 corrige la fila 37 de la rúbrica y las filas 50 y 51 de `rule-enforcers.md`.
- Se corrigió la lista de paquetes de `design.md`.
- La verificación suma la lectura de la regla instalada.
