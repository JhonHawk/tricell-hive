# Flujo de decisiones y pausas en la comunicación

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado (2026-10-09) · integrado en `development` con el [PR #36](https://github.com/JhonHawk/tricell-hive/pull/36) (merge `3b872726`), que cerró #9 y #10. Incluye los cuatro commits de la segunda premisa (`efaf2269`, `d084e4ae`, `b1ed8315`, `d49fc1d8`) y sus entradas en `CHANGELOG.md` (`fb1c865e`). `hive update` desde `development` dejó la instalación sin cambios: ya coincidía con lo desplegado desde la rama |
| Tracker · GitHub Issues | • [#9 — Claude Code: la pregunta nativa sale sin texto tras llamadas de herramienta](https://github.com/JhonHawk/tricell-hive/issues/9)<br>• [#10 — Decisiones pedidas en texto libre teniendo la herramienta nativa](https://github.com/JhonHawk/tricell-hive/issues/10) |
| Git | interactivo · rama desde `development` · parada antes del push para que el usuario lea la guía nueva |
| Verificación | `go test ./...` · comprobaciones `rg` por criterio · medición sobre sesiones reales después del despliegue |
| Siguiente paso | ninguno en este cambio; quedan ofrecidas la release y la medición posterior |

## Objetivo

Los agentes que usan Hive pierden decisiones en el camino hacia el usuario. En la semana del 1 al 8 de octubre de 2026, 82 de 208 tarjetas de preguntas de Claude Code salieron sin texto que las acompañara, aunque una regla lo prohíbe; en tres sesiones independientes el usuario tuvo que pedir las opciones porque quedaron solo en un ticket de Linear o solo en una tarjeta que rechazó; y el 5,5 % de sus seguimientos son «¿qué queda pendiente?». La causa es estructural: la sección de comunicación de la guía global creció con 53 commits de arreglos puntuales en tres semanas, varios por host, sin una regla central sobre dónde vive una decisión, con reglas que compiten por cómo empezar un mensaje y con decisiones mezcladas en la sección de pendientes.

Este cambio reescribe el flujo de decisiones y pausas alrededor de una invariante: toda decisión queda completa en el texto del mensaje (contexto, opciones con su efecto y recomendación). La tarjeta de preguntas de cada host se usa e incentiva como atajo para responder, redactada para entenderse sola. Cada pausa termina con una última línea fija que deja claro que el agente espera una decisión. La investigación termina sus rondas ofreciendo las recomendaciones en lugar de una línea de ruta, y una decisión que cambia el diseño de un plan se presenta directa con la etiqueta «Pausa».

## Alcance y aceptación

Incluido (decisión S1-A): las reglas de preguntas, tarjetas, pausas, cierre, rutas y pendientes en `content/guidance/global.md` y en las skills `flow-research`, `flow-plan`, `flow-build` y `flow-close`; la tabla de hosts de `_support/docs/architecture/agent-delivery.md`; las páginas del sitio que describen ese comportamiento.

Excluido: las reglas de formato, enlaces, glosas de identificadores, idioma y reportes de backlog (S1-A: no hay evidencia para cambiarlas); la opción experimental de Codex `default_mode_request_user_input` (D4-A); hooks u otro mecanismo de host (descartado: el historial que lee un hook de Claude Code puede no tener aún el último mensaje); los casos de regresión de `tests/fixtures/regression/` (AGENTS.md, Measurement); reanudar pruebas con modelos (pausadas).

Restricciones que deben seguir siendo verdad:

- `content/guidance/global.md` cabe en el presupuesto de `tests/content`, que sube como máximo 512 bytes, a 46.040 (D5-A); `go test ./...` pasa.
- La guía distribuida nombra capacidades, no herramientas de un host, salvo donde el comportamiento difiere por host y se justifica (AGENTS.md, Guidance structure).
- Cada regla tiene un solo hogar; no se agregan copias en otros archivos.
- Las reglas fuera del alcance quedan textualmente iguales.

Criterios de aceptación:

- AC1. La guía global exige que toda elección que se pide al usuario (decisión, ruta, pregunta de cierre, oferta o sí o no) quede completa en el texto del mensaje: contexto, cada opción con su efecto y la recomendación con su razón, aunque también viva en una tarjeta, un ticket o un documento. *Falso en la base cuando* `global.md:21` dice que el mensaje previo a la tarjeta da «the context and your recommendation, not the option list».
- AC2. La guía global permite e incentiva la tarjeta de preguntas en cualquier host que la ofrezca y que espere la respuesta, sin excepción por host; una tarjeta que vuelve antes de la respuesta solo se usa mientras siga trabajo independiente. *Falso en la base cuando* `global.md:22` dice «Grok Build sessions count as having no native question tool».
- AC3. La guía global exige que la tarjeta se entienda sola: su pregunta lleva el contexto mínimo para responder, cada opción lleva su efecto en la descripción, o en la etiqueta cuando la herramienta no tiene campo de descripción, y la recomendada va marcada. *Falso en la base cuando* ninguna regla pide contexto en el texto de la pregunta ni dice dónde va el efecto si falta el campo de descripción.
- AC4. La guía global dice que rechazar o cerrar una tarjeta sin respuesta cuenta como «sin respuesta», nunca como consentimiento ni como permiso para decidir. *Falso en la base cuando* `rg -n -i 'declin|dismiss' content/guidance/global.md` no encuentra esa regla.
- AC5. La guía global define una línea de pausa: una etiqueta fija en el idioma de la sesión con el significado de «esperando tu decisión», seguida de lo que se pregunta por ID con su glosa, siempre como última línea y después de los demás cierres; una elección abierta nunca va dentro de Pendiente. *Falso en la base cuando* `global.md:28` manda listar en Pendiente «An unanswered choice … ask it once with its options» y no existe etiqueta fija.
- AC6. `flow-research` termina cada ronda de hallazgos sobre trabajo posible con la pausa que ofrece mostrar las recomendaciones; las opciones y la ruta (`flow-plan` o `flow-build`) aparecen solo cuando el usuario acepta o pide continuar. *Falso en la base cuando* `flow-research/SKILL.md:36` manda terminar con «one line of plain text … naming the routes».
- AC7. `flow-plan` presenta una decisión que cambia el diseño del plan o de la investigación directamente con la etiqueta de pausa (en español, «Pausa»): explicación, opciones con efecto y recomendación en el mismo mensaje, y la pregunta como última línea. *Falso en la base cuando* `rg -n -i 'pause|pausa' content/skills/flow-plan` no encuentra esa regla.
- AC8. La guía global tiene una sola regla de precedencia para la apertura de un mensaje (trabajo en curso, luego cierre de un trabajo, luego la respuesta), y la regla de decisiones no pide abrir con la decisión. *Falso en la base cuando* `global.md:16`, `:17`, `:21` y `:24` piden cada una abrir con algo distinto sin orden entre ellas.
- AC9. La regla de enlazar el documento retenido excluye explícitamente las alternativas y todo lo que el usuario debe sopesar. *Falso en la base cuando* `global.md:17` dice «Link the retained document for technical detail instead of reproducing it» sin esa excepción.
- AC10. `flow-build/references/verification.md` y `flow-close` aplican la misma invariante: la validación manual y las rondas de corrección llevan el recorrido completo en el mensaje y pueden preguntar con tarjeta; la pregunta de cierre usa la misma pausa. *Falso en la base cuando* `verification.md` exige «a question in text» en el traspaso de UI y en las rondas de corrección, y `global.md:23` lo exige «on every host».
- AC11. La tabla de hosts de `agent-delivery.md` registra, con versiones verificadas el 2026-10-08, el nombre y los campos de la tarjeta de cada host, si espera la respuesta, qué queda en el historial y los issues abiertos que afectan la invariante. *Falso en la base cuando* la tabla dice «None found» para OpenCode y lista Claude Code 2.1.283 y Cursor 2026.09.23.
- AC12. Las páginas `site/src/content/docs/flows/research.md`, `plan.md` y `close.md` describen la pausa que ofrece recomendaciones, la pausa de plan y la última línea fija. *Falso en la base cuando* `research.md:37` dice que el agente «names the route it recommends».

## Entrega

- Repositorio y base: `tricell-hive`, base `development` (decisión E1-A, 2026-10-08).
- Modo interactivo:
  - Rama de trabajo `fix/gh-9-communication-decision-flow` (propuesta, todavía no creada) y commits locales verificados.
  - Parada antes del push para que el usuario lea el diff de `content/` y de las páginas del sitio. Si pide cambios de redacción, el agente los aplica en la misma rama, vuelve a correr las comprobaciones y se detiene de nuevo.
  - Tras la validación: push, PR a `development` que cierra #9 y #10, y merge cuando pasen las comprobaciones locales.
  - Después del merge, `hive update` despliega el contenido en los hosts registrados. El marcador `maintainer.local` está presente; no hace falta recompilar porque no cambia `tooling/` ni `integrations/`.
- Revisión de implementación: ninguna dedicada (E2-A). La cubren las comprobaciones por criterio, la batería de CONTRIBUTING y la lectura del usuario en la parada.
- `CHANGELOG.md`: el PR agrega su entrada en `[Unreleased]`, como exige CONTRIBUTING. Es un arreglo de nivel parche de la guía; al cerrar se propondrá una release si se cumple el criterio de CONTRIBUTING.
- Registro del cambio: se versiona una vez al cerrar, con un PR pequeño de cierre que fusiona el agente, como en cierres anteriores.
- Mensajes de commit (AGENTS.md, Measurement), con las sesiones reales que motivan el cambio:
  - Claude Code `71388ddc-1b5f-48f1-a08b-29fa4b834cc2`, carvento COM-786, 2026-10-08: las alternativas quedaron solo en el ticket.
  - Claude Code `5aa1a929-d0e5-45a4-a5c9-7600b39a83f9`, state-research, 2026-10-03 a 2026-10-05: las opciones quedaron solo en una tarjeta rechazada.
  - Claude Code `35bd0059-a362-404a-8354-b57a80ded0f6`, dened, 2026-10-02 a 2026-10-06: el usuario tuvo que pedir qué convenía.
  - Las tres sesiones del 2026-10-05 que registra #9, más el conteo de 82 de 208 tarjetas sin texto del 1 al 8 de octubre.
- Fuera de la entrega: la release y la medición posterior, que se ofrecen al cerrar.
