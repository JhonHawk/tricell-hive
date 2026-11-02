# Menos presión por tickets e investigaciones suficientes para la decisión

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #225](https://github.com/JhonHawk/tricell-hive/pull/225) (`dcaca050`) |
| Tracker · GitHub Issues | sin issue, por decisión del usuario (D19) |
| Git | automatic · PR a `development` y merge · registro en PR de cierre que integra el agente (D26-A) · sin parada humana |
| Verificación | comprobaciones `rg` por criterio · `go test ./...` · `hive-review-code` + `hive-refute-claim` |
| Siguiente paso | ninguno; registro archivado |

## Solicitud

> Ahi tambien en todas las sesiones del dia de hoy lo que he visto es que se da muuuucha prioridad a crear tickets y los research aunque más rápidos se estan quedando cortos en lo que se neccesitan

> En la sesión de codex sobre tricell-hive volvio a ocurrir que diseña el plan y me ofrece el build para una sigueinte sesión, una vez que acabes las tareas actuales valida si esto es algo que podemos incluir en el plan a resolver

Respuestas en las tarjetas (2026-10-27): D19 «Planificar ahora aquí», sin ticket; D20-A (el ticket en la ruta de investigación solo si el trabajo espera); D21-A (hallazgos al cierre: recomendar resolver en este trabajo); D22-A (prueba de suficiencia completa); D23-A (vigilar el esfuerzo de los hijos de investigación); D24-A (añadir el caso de Codex como AC8 y T6); D25-A (empezar `flow-build` ahora, aprobando AC1–AC8); D26-A (registro en un PR de cierre que integra el agente). Entrega reutilizada de esta sesión para tricell-hive: D16-A (automatic) y D17-A (`hive-review-code` con `hive-refute-claim`).

## Objetivo

En las sesiones del 2026-10-26/27, las reglas empujan a ofrecer y recomendar tickets en muchos momentos que se suman (la pregunta de ruta de la investigación, los hallazgos incidentales al cierre, los temas abiertos), y el usuario los rechazó con frecuencia para arreglar en el trabajo o seguir investigando. Al mismo tiempo, las investigaciones paran en cuanto sus fuentes convergen y recomiendan una ruta apoyada en inferencias sin comprobar, porque ninguna regla define cuándo la investigación basta para la decisión.

El cambio hace del ticket un destino para el trabajo que de verdad espera, recomienda resolver en el trabajo los hallazgos que caben, y añade a `flow-research` una prueba de suficiencia: la investigación responde cada parte de la pregunta, basta para elegir entre las rutas, y sus inferencias decisivas se comprueban cuando hay una forma barata o se recomienda seguir investigando.

## Evidencia (inspeccionada en esta sesión por tres hijos `hive-research`, citas comprobadas por el hilo principal)

- 32 de 166 tarjetas ofrecían abrir un ticket; en 23 era la opción recomendada; el usuario lo rechazó o cambió en 10.
- Hallazgos incidentales al cierre con «Abrir ticket» recomendado: 6 (Claude cam `c00e1e7f` ×3, cam `df8c65a1` H1 y H3, facti `9f73a051` F1); el usuario eligió resolver o investigar en los 6.
- Investigaciones que pararon a los 1–5 minutos con una inferencia decidiendo la recomendación, corregidas después en 5 de 7: Claude `1624fa4f`, `abdd6ea1`, `5fdb6aad`, `c00e1e7f`, `9f73a051`; Codex `01a179e0`; Pi `01a177d3`.
- Reglas que lo provocan: `content/guidance/global.md:21` (destinos de temas abiertos con el ticket primero), `:24` y `:78` (pregunta de ruta con «opening a ticket and stopping» obligatorio, desde `c3363b70`), `:34` («propose an issue in the report»); `content/skills/flow-research/SKILL.md:42` («one more option opens the ticket and stops») y `:46`; `content/skills/flow-close/references/incidental-findings.md:9` («mark which ones you recommend opening and why»); `flow-research/SKILL.md:26` y `content/agents/review/hive-research.md` (parar cuando la evidencia converge; `complete` cuando converge); ninguna regla define suficiencia.

## Alcance y aceptación

Incluido: la definición de petición solo de plan y la recomendación del build en la entrega del plan (`flow-plan/SKILL.md:72`, `plan-handoff.md`, página Plan del sitio); la pregunta de ruta de la investigación (`global.md:24`, `:78`, `flow-research`), los hallazgos incidentales (`global.md:34`, `incidental-findings.md`, `flow-build/references/verification.md:51`), los destinos de temas abiertos (`global.md:21`), la suficiencia de la investigación y el brief de los hijos (`flow-research`, rol `hive-research`), las páginas Research y Close de `site/` y el `CHANGELOG.md`.

Excluido, con su razón:
- El siguiente ticket de la pregunta de cierre (`flow-close/SKILL.md:43`): elige qué trabajo existente sigue; no crea tickets.
- Los puntos sin especificar de un plan al cerrar (`plan-format.md`): ya permiten tarea, ticket o exclusión.
- El esfuerzo del rol `hive-research` (`integrations/agent-profiles.json`): queda en vigilancia por D23-A.
- La línea de contexto «stage · ticket · progress»: el ticket ya se omite cuando no lo hay.

Restricciones: `go test ./...` pasa, incluido el límite de 16 000 bytes por skill y referencia; cada regla conserva un único hogar y los demás sitios apuntan a ella (AGENTS.md, Guidance structure); la regla de arreglar defectos que bloquean (`global.md:33`) no cambia.

Búsqueda de specs que decidió los deltas: términos «route question», «incidental», «ticket», «open topic», «research», «sufficien», «flow-research», «hive-research», «Build page», «Deploy page», «Research page», «Close page» en `_support/openspec/specs/*/spec.md`; leídos además los SHALL NOT de `documentation-site` y `context-delivery`. Ningún requisito describe estos comportamientos: sin deltas.

- AC1. `flow-research/SKILL.md` ya no exige una opción de abrir el ticket y parar: la ofrece, y puede recomendarla, solo cuando el trabajo no se hará ahora porque es grande o pertenece a otra persona u otro proyecto, o cuando el usuario lo pide; una opción que actúa puede abrir el ticket cuando el proyecto lleva su trabajo en un tracker. *Falso en la base cuando* `rg -n "one more option opens the ticket and stops" content/skills/flow-research/SKILL.md` encuentra la frase.
- AC2. `global.md` deja de imponer el ticket en la pregunta de ruta: `:24` no lista «opening a ticket and stopping» como opción obligatoria y `:78` no exige que cada opción diga si abre ticket. *Falso en la base cuando* `rg -n "plus opening a ticket and stopping|whose options say whether they open a ticket" content/guidance/global.md` encuentra alguna.
- AC3. Para los hallazgos incidentales al cierre, la regla recomienda resolverlos en este trabajo cuando el arreglo cabe en la sesión y en el alcance del trabajo, y recomienda el ticket solo cuando queda fuera del alcance, es grande o es de otra persona u otro proyecto. *Falso en la base cuando* `incidental-findings.md` dice «mark which ones you recommend opening and why» sin ese criterio.
- AC4. Los destinos de un tema abierto en `global.md` ponen el ticket después de la tarea en este trabajo y el plan, y aceptan «pertenece a otra persona u otro proyecto» como motivo para descartarlo. *Falso en la base cuando* `global.md:21` dice «a ticket, plan or task» y no menciona ese motivo.
- AC5. `flow-research` exige, antes de recomendar una ruta, que los hallazgos respondan cada parte de la pregunta y basten para elegir entre las opciones, y que el hilo principal compruebe sus propias afirmaciones decisivas que sean inferencia cuando hay una comprobación barata a su alcance, o recomiende seguir investigando. *Falso en la base cuando* `rg -n "every part of the question|own deciding" content/skills/flow-research/SKILL.md` no encuentra nada.
- AC6. El brief de un hijo de investigación nombra la decisión que informa y cada aspecto que el usuario pidió, y su condición de parada es tener lo suficiente para esa decisión, no solo la convergencia; el rol `hive-research` usa ese criterio para `complete`. *Falso en la base cuando* `flow-research/SKILL.md` solo dice «once the answer can be named and the evidence converges» y `hive-research.md` dice «`complete` when the answer converged».
- AC7. Las páginas Research y Close de `site/` describen el comportamiento nuevo sin copiar las reglas. *Falso en la base cuando* `site/src/content/docs/flows/close.md:47` dice «marks which ones it recommends turning into tickets», `research.md:37` dice «asks whether you want a ticket» o `research.md:47` describe el ticket como opción fija.

- AC8. `flow-plan` define que una petición es solo de plan únicamente cuando el usuario dice que no se implemente ahora o que se ejecute en otra sesión o más tarde, y que pedir un plan, aunque nombre dónde se ejecutará, no lo es; y `plan-handoff.md` añade dos cosas a su recomendación actual del build: parar se recomienda solo ante un bloqueo concreto que se nombra, y nunca se recomienda otra sesión para el build del plan recién terminado, como excepción a la ruta de sesión nueva de Recommendations (`reporting.md`) y al criterio de sesión nueva de la pregunta de cierre (`flow-close/SKILL.md`). La definición cambia solo la recomendación de la pregunta de build: no autoriza implementar sin la respuesta del usuario. *Falso en la base cuando* `flow-plan/SKILL.md:72` dice «For a planning-only request, stop with the plan» sin definirla y `rg -n "named blocker|planning-only when" content/skills/flow-plan` no encuentra nada.

**En vigilancia (D23-A):** si después de este cambio las investigaciones siguen quedándose cortas en Codex o Pi, subir el perfil del rol `hive-research`.

## Entrega

- Repositorio `tricell-hive`, base `development`, rama `fix/ticket-pressure-research-sufficiency` en `.claude/worktrees/ticket-pressure`.
- Modo automatic (D16-A, reutilizado): commits, revisión dedicada sobre el candidato local, push, PR a `development` y merge con las comprobaciones locales en verde (el repo no tiene CI). Después, `hive update` por la regla del mantenedor.
- Revisión dedicada: `hive-review-code`, con `hive-refute-claim` para sus hallazgos bloqueantes (D17-A, reutilizado).
- Ruta de cierre de este registro: PR de cierre pequeño a `development`, sin revisión dedicada, integrado por el agente (D26-A).
- Implementación autorizada por D25-A.
