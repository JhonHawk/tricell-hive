# CI posterior a la integración: esperar solo cuando cubre el cambio

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `rebuild/harness-engineering` (`71fcebe`) y desplegado como release `112fed74f1db` |
| Tracker · GitHub Issues | • [#43 — Cierre con el CI de la rama base en rojo: flow-build no espera el CI posterior al merge](https://github.com/JhonHawk/tricell-hive/issues/43) |
| Git | `direct-base` a `rebuild/harness-engineering` · sin PR · release a los seis hosts tras el push |
| Verificación | `go vet ./...` · `go test -race ./...` · comprobación de la regla con ARK-730 y ARK-734 · texto en los archivos de cada host |
| Siguiente paso | Ninguno en este cambio. El efecto en sesiones reales no se midió, porque los pilotos están pausados |

## Objetivo

Hoy ninguna regla de Hive dice qué hacer con el CI que un merge o un push dispara en la rama base. Eso falla en las dos direcciones. En sample-project ARK-730 esa corrida era la única que ejecutaba la suite completa del backend, que cubría la migración del cambio. Falló por un timeout de npm y la sesión cerró el ticket igual. En ark ARK-734 el modelo esperó por su cuenta 7 minutos una corrida cuyo único job nuevo probaba el backend, cuando el cambio era una línea de CSS en `customer-portal`. Así, 7 de los 27 minutos de la sesión fueron espera que no verificaba nada del cambio.

El cambio agrega a `verification.md` una regla escrita por capacidad. Se espera la corrida de la rama base solo cuando ejecuta un job que prueba un archivo cambiado y que no corrió en la PR; si no, su enlace va como recordatorio. Mientras esa corrida no pase, el ticket y el change folder esperan. Si falla, un fallo de infraestructura relanza una vez los jobs de test, un fallo que ya existía en la base es un hallazgo incidental, y uno de código deja la integración pendiente. `git-workflow` apunta a esa regla en el momento de verificar el merge o el push.

## Alcance y aceptación

**Incluye:**
- La regla en `content/skills/flow-build/references/verification.md`, sección "Frequency and coverage".
- Una referencia en `content/skills/git-workflow/SKILL.md`, donde el agente verifica el estado tras la acción de Git.
- Un comentario en #43 con las decisiones y el caso de ark, y el cierre del issue.
- Un mensaje de commit que nombra las dos sesiones y sus fallas, como pide `AGENTS.md` ("Measurement").

**Excluye:**
- Un caso de regresión en `tests/fixtures/regression/`: `AGENTS.md` ya no los agrega para cambios de guía mientras los pilotos sigan pausados.
- Tocar `content/guidance/global.md`. La regla retiene el ticket desde `verification.md`, como ya hace `:85` ([design.md](design.md#relación-con-la-regla-global-de-ticket-completado)).
- Arreglar un fallo de código detectado después de la integración. D2-A solo pide diagnosticarlo y reportarlo.
- Medir el efecto en sesiones reales: los pilotos siguen pausados.

**Criterios de aceptación:**
- AC1. `verification.md` dice lo siguiente. *Falso en la base cuando* `rg -n "base branch" content/skills/flow-build/references/verification.md` no devuelve una regla sobre las corridas que dispara la integración.
  - Tras un merge o push propio a la rama base, se espera la corrida solo cuando ejecuta un job que prueba un archivo cambiado y no corrió en la PR; sin PR, cuenta todo job de ese tipo.
  - El job se juzga por lo que prueba, no por el filtro que lo dispara.
  - También se espera cuando el workflow no muestra qué prueba un job, o cuando todavía no aparece una corrida que el push debía disparar.
  - En los demás casos se da el enlace como recordatorio.
  - Hasta que la corrida pase, esperan el movimiento del ticket y el cierre del change folder. Si el turno termina antes, la corrida abierta queda como pendiente.
- AC2. `verification.md` dice lo siguiente sobre una corrida esperada que falla. *Falso en la base cuando* `rg -n "failed jobs" content/skills/flow-build/references/verification.md` no devuelve nada.
  - Un fallo que la corrida anterior de la base ya tenía se trata como hallazgo incidental.
  - Un fallo de infraestructura relanza una vez solo los jobs de test fallidos, con autorización de la propia regla y nunca para jobs de despliegue o publicación.
  - Si ese relanzamiento falla o el fallo es del código del cambio, la integración no está completa, el ticket no pasa a completado y se reportan los jobs fallidos, el diagnóstico y el enlace como pendiente.
- AC3. El párrafo de `git-workflow/SKILL.md` que verifica el estado tras una acción de Git remite, tras un merge o push a la rama base, al CI que disparó, según `verification.md`. *Falso en la base cuando* `rg -n "also check the CI it triggered" content/skills/git-workflow/SKILL.md` no devuelve nada.
- AC4. #43 tiene un comentario con D1-A, D2-A, D5-A y D6-A, el caso de ark ARK-734 como segunda evidencia y el commit que integra la regla. *Falso en la base cuando* el issue no tiene comentarios.

**Restricciones:**
- `go vet ./...` y `go test -race ./...` siguen pasando.
- No se incluye en el commit ningún cambio ajeno. La otra sesión ya hizo commit de su edición de `verification.md` en `f343468`; si aparecen hunks ajenos sin commit en estos archivos, se conservan fuera del commit.

## Decisiones del usuario (2026-09-27)

- **R1:** resolver #43 con una regla acotada, con `flow-plan` en esta sesión.
- **D1-A:** esperar solo cuando la corrida cubre el cambio; ante la duda, esperar; si no, recordatorio con el enlace.
- **D2-A:** distinguir infraestructura de código. Infraestructura: un relanzamiento de los jobs fallidos y, si vuelve a fallar, pendiente para el usuario. Código: el ticket no se completa; se diagnostica y se reporta.
- **D4-A:** sin revisión dedicada del código. Es texto en una referencia, sin contrato, datos ni seguridad, y lo cubren la revisión del plan y la comprobación con los dos casos.
- **D5-A:** la regla aplica también a los push directos a la rama base; sin PR, cuenta todo job que pruebe un archivo cambiado.
- **D6-A:** un fallo que ya tenía la corrida anterior de la base es un hallazgo incidental y no bloquea el ticket.

## Entrega

Respuesta del usuario, **D3-A** (2026-09-27):
- **`direct-base`:** un commit con las rutas de este cambio y esta carpeta `openspec`, con push a `rebuild/harness-engineering`. Después, el commit de cierre con el archivo del cambio, también con push. Sin PR.
- **Release:** nueva release a los seis hosts tras el push, desde un worktree limpio del commit. Se verifica el texto en los archivos de cada host y se borran el worktree y el plan de instalación.
- **Tracker:** el comentario en #43 y el cierre del issue quedan autorizados.

Este repositorio no tiene `.github/workflows`, así que la regla nueva no afecta a su propio push. Antes del push se hace `git pull --ff-only`, porque otra sesión trabaja en paralelo.
