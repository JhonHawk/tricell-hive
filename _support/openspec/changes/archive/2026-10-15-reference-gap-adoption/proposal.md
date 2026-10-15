# Adopción de huecos frente a las referencias (2026-10-13)

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): Completado (2026-10-14): T1–T5 `[x]`, `hive-verify-change` AC1–AC6 cumplidos; D8-A: cuatro commits en `feat/flash-class-guidance` (47320509 T1–T3, bed19f4d T4, 843a708c T5, 303623e6 roles) y despliegue local de 303623e6 con `hive update --out` + `hive apply` · pendiente: integración en `development` tras tus pruebas; el registro sigue sin versionar hasta ese cierre |
| Tracker · GitHub Issues | Sin issue: el usuario decidió (D2) que la lista de vigilancia vive en memoria, no en el tracker |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | `go test ./tests/content/...` por tarea · `go test ./...` sobre el candidato final · `hive-verify-change` · sin revisión dedicada |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

> actualiza gentle-ai y las skills de mattpodock, ve si hay algo nuevo que se pueda usar

> entonces sobre las versiones actuales (y actualicemos superpowers) a ver si hay algo aunque no sea super nuevo que falte

Respuestas del usuario a las preguntas de ruta (2026-10-13, tarjetas nativas):

- D1 (ruta): "Corregir H1 y adoptar algunos ya" (flow-plan).
- D2 (issue de vigilancia): "No, solo en memoria".
- D3–D5 (qué huecos adoptar): H2 con la acotación "Sobre H2, es relevante cuando se asigna o se solicita asignar un revisor para el PR (humano)", H5, H9 y H11. No adoptados y en vigilancia con conteo 0: H3, H4, H6, H7, H8, H10, H12 y la lista de menor valor.

## Objetivo

La revisión de las tres referencias externas actuales (clones locales de la bóveda, 2026-10-13) frente al contenido de Hive en la rama de prueba encontró una contradicción verificada y varios huecos de práctica ordinaria. El usuario eligió corregir la contradicción y adoptar cuatro huecos por decisión explícita, como permite la regla de admisión de `AGENTS.md` (Measurement). El resultado es guía distribuida que (1) deja de afirmar algo falso en Codex sobre la herencia de contexto de los hijos, (2) prohíbe en la guía global que un hijo lance hijos, (3) dice qué hacer cuando un revisor humano del PR deja comentarios, (4) obliga a comprobar código compartido antes de llamar "mecánico" a un cambio y (5) hace visibles en el reporte de cierre las decisiones de implementación tomadas sin preguntar.

## Alcance y aceptación

Incluido: cinco ediciones de texto en la guía global, dos referencias de skills, un documento de arquitectura y las páginas de `site/` que describen ese comportamiento. Todo el texto de archivos fuente en inglés; parafraseado, sin copiar redacción de las referencias ni nombrarlas en contenido versionado (regla local de atribución, 2026-10-02).

Excluido, con razón:

- H3, H4, H6, H7, H8, H10, H12 y los ítems de menor valor: sin dos sesiones fallidas observadas; quedan en memoria como vigilancia con conteo 0.
- Cambios en los roles (`hive-review-code`, `hive-verify-task`, `hive-review-plan`): sus líneas "no delega partes" se conservan como restricción propia del rol; la regla global nueva las cubre y no las contradice.
- Un detector nuevo en `tests/pilot/`: `AGENTS.md` lo prohíbe para cambios de guía mientras los pilotos están pausados.
- Despliegue local (`hive update`) y versionado: fuera del plan; los decide el usuario según el flujo de rama de prueba.

Restricciones que deben mantenerse: `go test ./...` sigue en verde (incluye `tests/content/skill_links_test.go` y `voices_test.go`); el cambio de tamaño de `content/guidance/global.md` se declara en el PR (`tests/content/budget_test.go` lo registra sin cuota); las reglas nuevas nombran su alcance explícito, porque los modelos del perfil `execution` aplican las instrucciones literalmente.

Criterios de aceptación (falsos en la base `f8116af5`):

- AC1. Ni la guía global ni `delegation.md` afirman ya que el hijo "no hereda la conversación del padre"; ambas dicen que, cuando la herramienta de lanzamiento ofrece heredar la historia del padre, el hijo se lanza sin ella, que el brief lleva lo que necesita y que la historia heredada no entrega las skills que el padre leyó. `agent-delivery.md` registra el hecho de Codex (`fork_turns`: un rol lanzado por `agent_type` exige `none`; un hijo genérico sin `agent_type` hereda toda la historia por defecto) con su fuente, versión y etiqueta de evidencia. *Falso en la base cuando* `rg -n "not inherit the parent" content/` devuelve `global.md:80` y `delegation.md:35`, y `rg -n fork_turns _support/docs/architecture/agent-delivery.md` no devuelve nada.
- AC2. El contrato de delegación de la guía global (`global.md`, bullet "Delegation grants no additional authority") y el contrato del hijo en `delegation.md` dicen que un hijo no lanza hijos propios y reporta lo que no puede terminar solo. *Falso en la base cuando* `rg -n -i "launch(es)? (no )?children|children of (its|their) own" content/guidance content/skills` no devuelve nada.
- AC3. `git-workflow/SKILL.md` tiene una regla, acotada a un PR con revisor humano asignado o solicitado, para manejar sus comentarios: leer cada uno contra el código antes de cambiar nada; si alguno es ambiguo, preguntar en el hilo antes de implementar cualquiera; aplicarlos en orden bloqueante, simple, complejo, verificando cada uno; discrepar con evidencia cuando el código ya hace lo pedido o lo pedido rompería comportamiento declarado; responder en el hilo de cada comentario. *Falso en la base cuando* `rg -n -i "reviewer.s comment|human reviewer" content/skills/git-workflow/SKILL.md` no devuelve nada.
- AC4. La definición de cambio mecánico en `global.md` exige comprobar, antes de llamar mecánico a un cambio que altera lo que hace un fragmento de código (no su nombre ni su ubicación), si ese código sirve también a comportamiento que la petición no nombró (helper, parser, opción compartida), y en ese caso lo trata como cambio de contrato; un rename o un move siguen siendo mecánicos. *Falso en la base cuando* `rg -n "did not name" content/guidance/global.md` no devuelve nada.
- AC5. El reporte de cierre (`flow-close/references/reporting.md`, condición "a work item completes") exige listar cada decisión de implementación tomada sin preguntar que elige entre comportamientos o contratos que el usuario pudo razonablemente elegir distinto, con su razón y cómo revertirla, en modo atendido y desatendido; `unattended-delegation` remite a esa regla en lugar de repetirla. *Falso en la base cuando* `rg -n "without asking" content/skills/flow-close/references/reporting.md` no devuelve nada y `unattended-delegation/SKILL.md:16` contiene su propia lista.
- AC6. Las páginas `site/src/content/docs/flows/build.md` y `close.md` describen el comportamiento nuevo de AC3 y AC5 en una frase cada una, sin copiar las reglas. *Falso en la base cuando* `rg -n -i "human reviewer|without asking" site/src/content/docs/flows/build.md site/src/content/docs/flows/close.md` no devuelve nada.

## Entrega

Respuestas del usuario (2026-10-14): D6-A `hold`: editar y verificar en el worktree de `feat/flash-class-guidance`, presentar el diff y no hacer commit, push ni PR hasta que el usuario lo pida; el registro de cambio queda sin versionar y se reporta como pendiente. D7-A sin revisión dedicada: inspección propia del diff, `go test ./tests/content/...`, `go test ./...` sobre el candidato final y `hive-verify-change`. La implementación quedó autorizada con D1-B ("adoptar algunos ya") y D3–D5.
