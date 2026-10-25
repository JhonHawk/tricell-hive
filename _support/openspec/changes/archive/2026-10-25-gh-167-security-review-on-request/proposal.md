# Revisión de seguridad solo a pedido

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #170 (`c0307e80`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#167 — El auditor de seguridad se lanza solo: debe correr únicamente a pedido del usuario](https://github.com/JhonHawk/tricell-hive/issues/167) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | conteo de frases · lectura · `go test ./tests/content/...` · `go vet ./...` y `go test ./...` sobre el candidato final |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

Supernota del usuario al final de la sesión `b979e8b2` (2026-10-25):

> El auditor de seguridad y la skill solo deben ejecutarse si se solicita, se puede sugerir, pero no autoinvocar

Sesión `5470b05e` (2026-10-25): «Investiga con flow-research las notas de engram:tricell-hive/pending/supernotes-security-and-verify-checkpoints: auditoría de seguridad solo a pedido, y verificación por puntos de control según el riesgo en vez de una por tarea.» Respuestas: D1-A (un ticket por tema y `flow-plan` conjunto), D3-A (el paso 5 de `security-advisory` también pasa a sugerencia), D7-A (se ofrece en la ronda de entrega y revisión y se menciona en el informe sin frenar la integración); D11-A durante el build (la oferta es una línea, sin pregunta que espere ni pausa; se repite solo ante un límite nuevo; el criterio de recomendación vive en un solo lugar).

Criterios del ticket [#167](https://github.com/JhonHawk/tricell-hive/issues/167), leídos el 2026-10-25:

> - AC1: ningún texto de `content/` lanza `hive-review-security` sin un pedido del usuario; flow-build, delivery-decisions y flow-walkthrough lo sugieren cuando el cambio cruza un límite.
> - AC2: un `Review` declarado ya no autoriza la revisión de seguridad.
> - AC3: la descripción del rol indica que se usa solo a pedido del usuario.
> - AC4: el paso 5 de `security-advisory` sugiere la evaluación de alcance en el informe en lugar de lanzarla.
> - AC5: `plan-execution.md`, la página `site/.../flows/build.md` y los demás textos que dependen de la revisión automática quedan coherentes.

## Objetivo

Hoy `hive-review-security` corre sin que el usuario lo pida: `flow-build` lo lanza antes de pedir validación cuando el cambio cruza un límite de seguridad, un `Review` declarado lo autoriza, la puerta del candidato aceptado lo repite, la descripción del rol permite que el host lo delegue por su cuenta y el barrido de `security-advisory` lo lanza para medir alcance. El usuario quiere decidir cuándo corre: el agente lo sugiere con los límites que cruza el cambio, y lo lanza solo cuando el usuario lo pide. Con `Delivery: automatic`, la sugerencia llega en la ronda de entrega y revisión, donde la respuesta cuenta como pedido, y la que aparezca después va al informe sin detener la integración.

## Alcance y aceptación

Incluido: los puntos de lanzamiento automático de `hive-review-security` y los textos que lo describen. Excluido, con su razón:

- La regla global de escalar de inmediato un problema de seguridad hallado (`content/guidance/global.md:33`): no lanza al auditor.
- La lectura de [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md) por el implementador y el planificador, y la verificación de dependencias antes de instalarlas: son prácticas del propio agente, no el auditor.
- El aviso de límites al final del veredicto de `hive-review-code` (`content/agents/review/hive-review-code.md:20`): se mantiene y pasa a alimentar la sugerencia.
- La verificación por puntos de control: [gh-168](../gh-168-verification-checkpoints/proposal.md).

Restricciones: `flow-build/SKILL.md` y `flow-walkthrough/SKILL.md` siguen ≤ 16 000 bytes (`tests/content/skill_size_test.go`); la suite actual pasa.

- AC1. La secuencia previa a pedir validación de `flow-build` lanza `hive-review-security` solo cuando el usuario lo pidió; si el cambio cruza un límite y nadie lo pidió, el mensaje de validación nombra los límites y ofrece la revisión. *Falso en la base cuando* `content/skills/flow-build/SKILL.md:58` lanza `hive-review-security` por cruzar un límite.
- AC2. Un `Review` declarado ya no autoriza `hive-review-security`. Cuando el cambio planeado cruza un límite, el aviso de una línea o la ronda de entrega ofrece la revisión de seguridad en una línea, sin una pregunta que espere, también cuando `Delivery` y `Review` están declarados (D11-A). *Falso en la base cuando* `content/skills/flow-plan/references/delivery-decisions.md:26` dice que un `Review` declarado «also authorizes `hive-review-security`».
- AC3. Una regla única, «Offer the security review» en `security-boundaries.md`, fija la oferta (límites, recomendación, momento, repetición solo ante un límite nuevo) y la ejecución pedida: antes de pedir validación o, sin parada de validación, antes del push o merge; cubre también los límites que aparecen después en el mismo cambio y el pedido después del merge. Los demás archivos remiten a ella. *Falso en la base cuando* `flow-build/SKILL.md:58` dice que la revisión que nombra un revisor de código «stays at that reviewer's point», es decir, corre.
- AC4. La puerta del candidato aceptado de `flow-walkthrough` repite la revisión de seguridad solo si el usuario la pidió y las correcciones tocan un límite; si una corrección cruza un límite nuevo y nadie la pidió, el siguiente mensaje la ofrece. *Falso en la base cuando* `content/skills/flow-walkthrough/SKILL.md:62` la repite con solo tocar un límite.
- AC8. Bajo delegación desatendida, una revisión de seguridad ofrecida no cuenta como pedida mientras el usuario no la elija: el agente no se contesta esa pregunta. *Falso en la base cuando* `content/skills/unattended-delegation/SKILL.md:16` solo exceptúa de decidir solo «an authorization the user was asked for and left unanswered».
- AC5. La descripción de `hive-review-security` dice que se usa solo cuando el usuario lo pide. *Falso en la base cuando* `content/agents/review/hive-review-security.md:3` dice «Use when a change touches a security boundary».
- AC6. El paso 5 de `security-advisory` ofrece la evaluación de alcance en el informe en lugar de lanzarla. *Falso en la base cuando* `content/skills/security-advisory/SKILL.md:27` dice «launch a `hive-review-security` child».
- AC7. La página `site/src/content/docs/flows/build.md` dice que la revisión de seguridad corre solo a pedido y cuándo se ofrece, incluida la oferta de una línea aunque `Delivery` y `Review` estén declarados (D11-A), y `CHANGELOG.md` lo registra con `(#167)`. *Falso en la base cuando* ninguno de los dos menciona `#167` ni la revisión de seguridad a pedido.

Ticket AC5 → AC7 y la comprobación de coherencia de T1; `plan-execution.md:30` y `site/.../build.md:71` hablan de arreglos que pide «the security review», que siguen válidos cuando el usuario la pidió, así que no cambian.

## Entrega

Valores declarados en `AGENTS.md` `## Hive` (base `development`): `Delivery: automatic` y `Review: /code-review`. Sin efecto de UI renderizada (solo texto de la guía y de una página del sitio). Worktree propuesto `.claude/worktrees/gh-167`, rama propuesta `feat/gh-167-security-review-on-request`; el agente integra tras CI y `/code-review`. La promoción a `master` queda fuera. Con [gh-168](../gh-168-verification-checkpoints/proposal.md) comparte `CHANGELOG.md` y `site/.../flows/build.md`: se integra primero y gh-168 se actualiza sobre esta.

Temas abiertos: la verificación por puntos de control tiene destino en gh-168.
