# Reglas contradictorias y fricción en el flujo; portal y documentación al día

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por el [PR #181](https://github.com/JhonHawk/tricell-hive/pull/181) (merge `31d856d0`) |
| Tracker · GitHub Issues | • [#180 — Reglas contradictorias y pasos de más entre plan, build, verify y close; el portal y la documentación no reflejan development](https://github.com/JhonHawk/tricell-hive/issues/180) |
| Git | automatic · PR a `development` y merge tras `/code-review` y la suite local · sin parada humana |
| Verificación | comprobaciones por frase · `go test ./...` · `pnpm build` y `pnpm check` del sitio · `/code-review` sobre el candidato local |
| Siguiente paso | ninguno; la release que publica el portal la pide el usuario aparte |

## Pedido

> La premisa es la siguiente, es probable que tengamos reglas contradictorio o de fricción en el flujo, analizalas, incluyendo la documentación y que el portal este actualizado

> sobre automatic, cuando se escoge esta opción si deben ofercerse alguna forma de code-review

> Otro tema que creo pendiente es el tema de la preguntas, en algunas partes se pone el bloque completo de preguntas, pero me ha pasado que la segunda pregunta asume que se va a escoger la opción recomendad en la primera, entonces, las preguntas que cambian rumos de investigación o el esquema de preguntas diferentes deben presentarse en turnos diferentes

> Abramos un solo ticket para resolver todo

> todo esto es en conjunto, con la premisa inicial, y puedes analizar las sesiones de los host para confirmar lo que menciono

Criterios de aceptación de [#180](https://github.com/JhonHawk/tricell-hive/issues/180), citados textualmente (leído el 2026-10-25):


- AC1 (H3): `delivery-decisions.md` deja de usar un cambio de pantalla como ejemplo genérico para descartar `automatic`. Solo una regla del proyecto que exija validación humana lo descarta, y esa regla se nombra.
- AC2 (H4): `flow-walkthrough` dice qué pasa con la primera vista renderizada bajo `automatic` y bajo delegación desatendida, y ninguno de los dos casos termina en una pregunta al usuario. `flow-build` y `unattended-delegation` apuntan a esa regla.
- AC3 (H5, pedido del usuario): al elegir `automatic`, la ronda de entrega siempre ofrece un mecanismo de revisión de código, salvo que `Review` esté declarado (y entonces lo nombra). `delivery-decisions.md` dice en un solo lugar, para cada modo, en qué momento corre la revisión dedicada: sobre el diff local antes del push o sobre el PR. Las seis menciones apuntan ahí, y una tarea añadida después de la revisión no vuelve a preguntar cuando `Review` o una respuesta previa ya lo cubre.
- AC4 (H6): `flow-build` dice qué hace `interactive` cuando verify se omite, y queda definida la parada de validación para `direct-base` en un repo de código.
- AC5 (H7): el punto de cierre posterior se declara con una clave de `## Hive` definida en `project-settings.md`, o se elimina de `global.md`. `flow-close`, `reporting.md` y `flow-deploy` lo aplican igual, y `flow-close` define el estado del ticket y el reporte para `hold` y para un PR que espera revisión humana.
- AC6 (H8): `direct-base` elegido como modo y `direct-base` declarado se distinguen por nombre o por una frase explícita, y las concesiones de repo de soporte se aplican solo al valor declarado.
- AC7 (H9, H10): un cambio mecánico no recibe la pregunta de revisión, y `flow-verify` no hace ninguna pregunta propia sobre la prueba en vivo.
- AC8 (H11): el criterio de delegación tiene un solo dueño (`delegation.md`); `flow-build` y `flow-plan` apuntan a él, y `flow-research` y `security-advisory` dejan de remitir a un umbral que no aplica con una skill activa.
- AC9 (H12): `flow-plan` apunta a la línea de pausa de `questions.md` sin pedir una etiqueta de apertura distinta.
- AC10 (H13): la activación por silencio funciona cuando `Delivery` y `Review` están declarados, o la regla dice que en ese caso no aplica.
- AC11 (H14, H15): la frase de borrado de worktrees de `global.md` y de `artifact-layout.md` ya no admite la lectura que permite borrar trabajo sin confirmar (`artifact-layout.md` apunta a `flow-close`). El paso 5 de `flow-close` queda limitado a ramas de este trabajo o de trabajos anteriores del usuario. `git-workflow` incluye la excepción de `--force-with-lease` y el caso sin CI.
- AC12 (H16): cada skill que lanza un rol le pide algo que el contrato del rol permite (diagnóstico, consulta al tracker, solo grabar, revisión base), y `hive-review-ux` pide leer browser automation antes de su primer paso con navegador.
- AC13 (H17): `flow-close/SKILL.md` y `flow-plan/SKILL.md` quedan con al menos 500 bytes libres bajo el límite, y `tests/content/skill_size_test.go` pasa.
- AC14 (H2): las páginas de `site/`, el README, `llms.txt`, `CONTRIBUTING.md` (paso `pnpm build`) y `documentation-site/spec.md` describen el comportamiento de `development` sin copiar reglas.
- AC15: cada hallazgo menor de la lista siguiente queda corregido o descartado con su razón en el registro del cambio.
- AC16 (H18): `questions.md` dice que una pregunta cuyas opciones o contexto dependen de la respuesta a otra, como las que cambian el rumbo de una investigación o abren un esquema distinto de preguntas, se hace en un turno posterior, después de esa respuesta. En una misma ronda solo van preguntas cuyas opciones valen con cualquier respuesta a las demás. La regla de conjunto en secuencia (`:29-31`) queda limitada a las autorizaciones de un objetivo ya decidido.

## Objetivo

La auditoría del 2026-10-25, sobre `70d2bb13` y después del rediseño de etapas (#169), encontró que la guía distribuida da instrucciones incompatibles en varios momentos del flujo. Los casos principales son tres. La revisión dedicada corre «en el punto que fija el modo», pero ningún lugar fija ese punto. `automatic` puede detenerse en la primera vista renderizada o pasar a `interactive` por un ejemplo genérico. Y `direct-base` y el «punto de cierre posterior» significan cosas distintas según el archivo. A esto se suman pasos y preguntas de más en el camino normal, una frase de borrado de worktrees que admite una lectura peligrosa, briefs que piden a un rol lo que su contrato no permite, y dos `SKILL.md` en el límite de tamaño.

El usuario agregó dos requisitos. Primero, con `automatic` siempre se ofrece una forma de revisión de código. Segundo, una pregunta que depende de la respuesta a otra va en un turno posterior. El portal publicado sigue en v0.5.0, y varias páginas, el README y `llms.txt` no reflejan `development`.

El resultado esperado: cada momento del flujo tiene una sola regla, con un dueño claro, que un modelo de ejecución puede aplicar al pie de la letra. Las páginas describen el comportamiento actual, listas para la próxima release.

## Alcance y aceptación

Incluye `content/guidance/global.md`, las skills y referencias de `content/skills/`, los roles de `content/agents/`, las páginas de `site/src/content/docs/`, `README.md`, `llms.txt`, `CONTRIBUTING.md`, `_support/docs/architecture/` y la especificación `documentation-site`.

Queda fuera:
- La release que publica el portal: el usuario la pide aparte, cuando AC14 esté integrado.
- Reanudar pilotos de comportamiento: siguen pausados por `AGENTS.md`.
- Una clave nueva de `## Hive` y cualquier cambio en Go (D6-A).
- Renombrar `Delivery: direct-base`: lo declaran `cam-specs`, `carvento-specs` y `cam-frontend-mocks`.
- El PR de cierre del registro por cada cambio: es una decisión de diseño vigente.
- Los casos de `tests/fixtures/regression/`: no se agregan ni se cambian, como pide `AGENTS.md`.

Restricciones que deben seguir valiendo: `go test ./...` pasa; ningún `SKILL.md` supera 16 000 bytes; cada regla conserva un solo hogar (`AGENTS.md`, Guidance structure); el PR declara el cambio de tamaño de `global.md` y los lugares revisados.

Adopción: el usuario eligió el ticket que reúne todas las correcciones, lo que aprueba adoptarlas sin esperar recurrencia (`AGENTS.md`, Measurement). La evidencia de H18 en sesiones reales se agrega a [design](design.md#evidencia-de-sesiones) cuando vuelve la revisión de sesiones.

Criterios del plan (cada uno es falso en la base `70d2bb13`):

- AC1. `delivery-decisions.md` no usa un cambio de pantalla como ejemplo genérico para descartar `automatic`. Solo lo descarta una regla del proyecto que exija validación humana, y esa regla se nombra (D5-A). *Falso en la base cuando* `delivery-decisions.md:10` contiene «such as a visual check of a changed screen».
- AC2. Bajo `automatic` y bajo delegación desatendida, la primera vista renderizada se captura y se reporta sin hacer una pregunta (D5-A). `flow-build` y `unattended-delegation` apuntan a esa regla. *Falso en la base cuando* `flow-walkthrough/SKILL.md:24` pregunta sin condición de modo.
- AC3. Al elegir `automatic` sin `Review` declarado, la ronda de entrega ofrece un mecanismo de revisión de código como opción recomendada, y «Sin revisión dedicada» nunca es la recomendada (D10-A). `delivery-decisions.md` dice en un solo lugar que la revisión dedicada corre sobre el candidato local, después de la suite completa y antes de la parada de validación o del push, en cada modo (D4-A). Las menciones de «the point the delivery mode sets» apuntan ahí, y una tarea añadida después de la revisión no vuelve a preguntar si `Review` o una respuesta previa ya la cubre. *Falso en la base cuando* la tabla de modos no fija ese momento y `plan-revision.md:11` pregunta sin excepción.
- AC4. `flow-build` dice que, con `interactive`, se detiene antes del push aunque verify se omita, y muestra el diff y las comprobaciones que lo cubren (D7-A). Además, la parada de validación de `direct-base` en un repo de código queda definida. *Falso en la base cuando* `flow-build/SKILL.md:64` no menciona esa parada.
- AC5. `project-settings.md` define una sola vez el punto de cierre posterior declarado en la guía del proyecto (D6-A): la obra se cierra al integrar, el ticket queda en su estado intermedio y pasa al estado final cuando se cumple la condición que nombra la guía del proyecto, movimiento que `flow-deploy` aplica cuando corre. `global.md` apunta ahí y autoriza los dos movimientos; `flow-close`, `reporting.md` y `flow-deploy` también apuntan ahí. `flow-close` define el estado del ticket y la redacción del reporte para `hold` y para un PR que espera revisión humana. *Falso en la base cuando* `project-settings.md` no menciona el punto de cierre.
- AC6. `delivery-decisions.md` dice que las concesiones de repo de soporte (sin pregunta, sin esperar CI, salto de la protección de rama) valen solo para `Delivery: direct-base` declarado, no para el modo elegido en un repo de código. *Falso en la base cuando* `delivery-decisions.md:11` no lo distingue.
- AC7. Un cambio mecánico no recibe la pregunta de revisión, y `flow-verify` no hace ninguna pregunta propia sobre la prueba en vivo. *Falso en la base cuando* `flow-verify/SKILL.md:28` dice «offer it» y `flow-build/SKILL.md:24` no exime los cambios mecánicos.
- AC8. El criterio de delegación tiene un solo dueño en `flow-research/references/delegation.md`: lo pequeño y entendido queda directo (D9-A). `flow-build`, `flow-plan`, `flow-research` y `security-advisory` apuntan a él sin remitir a un umbral que no aplica con una skill activa. *Falso en la base cuando* `flow-build/SKILL.md:38` dice «goes to a child by default».
- AC9. `flow-plan` apunta a la línea de pausa de `questions.md` sin pedir una etiqueta de apertura. *Falso en la base cuando* `flow-plan/SKILL.md:22` contiene «such as "Pausa" in Spanish».
- AC10. Con `Delivery` y `Review` declarados, el silencio después de responder la pregunta de ruta activa la delegación desatendida para esa obra (D8-A). *Falso en la base cuando* `unattended-delegation/SKILL.md:8` exige respuestas a las preguntas de entrega y revisión.
- AC11. La frase de borrado de `global.md` y `artifact-layout.md` no admite una lectura que permita borrar un worktree con archivos sin confirmar, y `artifact-layout.md` apunta a `flow-close`. El paso 5 de `flow-close` limpia solo ramas de PRs cuyo autor es el usuario autenticado. `git-workflow` incluye la excepción desatendida de `--force-with-lease` y el caso sin CI. *Falso en la base cuando* `global.md:117` dice «except a squash- or rebase-merged branch that flow-close confirms against its pull request, or a worktree».
- AC12. Cada skill que lanza un rol le pide algo que su contrato permite: el hilo principal corre los experimentos de diagnóstico, `hive-read-state` acepta consultas al tracker, `hive-verify-change` tiene un modo de solo grabar y el hilo principal sirve la revisión base. `hive-review-ux` pide leer browser automation antes de su primer paso con navegador. *Falso en la base cuando* `hive-read-state.md:3` dice solo «explicitly identified targets».
- AC13. `flow-close/SKILL.md` y `flow-plan/SKILL.md` quedan con al menos 500 bytes libres bajo 16 000. *Falso en la base cuando* miden 15 955 y 15 799 bytes.
- AC14. Las páginas de `site/`, el README, `llms.txt`, `CONTRIBUTING.md` (con `pnpm build` antes de `wrangler deploy`) y la especificación `documentation-site` describen el comportamiento de `development` sin copiar reglas. *Falso en la base cuando* `site/src/content/docs/flows/verify.md:25` dice «A full local suite for a change that is not mechanical» y `skills.md` omite `flow-walkthrough`.
- AC15. Cada hallazgo menor que la tabla de [design](design.md#hallazgos-menores-destino) marca «Corregir» está corregido en el candidato, y cada descartado conserva su razón ahí. *Falso en la base cuando* `index.mdx:48` apunta a «Get started» para los hosts y `hive-build-data.md:10` remite al paso 3 antes de leerlo.
- AC16. `questions.md` dice que una pregunta cuyas opciones o contexto dependen de la respuesta a otra se hace en un turno posterior, después de esa respuesta. En una misma ronda solo van preguntas cuyas opciones valen con cualquier respuesta a las demás. La regla de conjunto en secuencia queda limitada a las autorizaciones de un objetivo ya decidido. *Falso en la base cuando* `questions.md:31` contiene «even when a later step depends on an earlier one's result» sin esa limitación.
- AC17. La pregunta de qué hallazgos abrir como tickets ofrece, por hallazgo o en conjunto, resolverlos en esta obra, nombrando la ruta `flow-build` o `flow-plan` (D11-A). La pregunta de qué sigue va en un turno posterior cuando depende de esa respuesta. El hogar es `incidental-findings.md`; `flow-close` y `global.md:33` apuntan ahí. *Falso en la base cuando* `global.md:33` dice «at the close, ask once which to open» sin la opción de resolverlos.

## Decisiones en vigor

- **D4-A:** la revisión dedicada corre sobre el candidato local, en todos los modos.
- **D5-A:** `automatic` sigue siendo `automatic` con UI; la primera vista se captura y se reporta sin pausa.
- **D6-A:** el punto de cierre posterior sigue declarándose en prosa, con un solo comportamiento.
- **D7-A:** `interactive` se detiene antes del push aunque verify se omita.
- **D8-A:** con valores declarados, el silencio después de responder la ruta activa la delegación desatendida.
- **D9-A:** un solo criterio de delegación: lo pequeño, directo.
- **D10-A:** con `automatic`, «Sin revisión dedicada» nunca es la opción recomendada.
- **D11-A:** la pregunta de tickets para hallazgos ofrece resolverlos en esta obra.
- **Requisitos del usuario:** con `automatic` siempre se ofrece una revisión (AC3), y las preguntas dependientes van en turnos posteriores (AC16).

Las decisiones menores tomadas sin preguntar están en [design](design.md#decisiones-menores).

## Entrega

- **Repositorio:** `tricell-hive`, base `development`, `Delivery: automatic`, `Review: /code-review`, ambos declarados en `AGENTS.md`.
- **Secuencia:** worktree `.claude/worktrees/gh-180`, rama propuesta `fix/gh-180-rule-contradictions`, commits por tarea, `go test ./...` y la construcción del sitio, `/code-review` sobre el candidato local antes del push (D4-A aplicado a este trabajo), push, PR a `development` y merge cuando pase. El repo no tiene CI: si la base se mueve antes del merge, se integra la base, se vuelven a correr `go test ./...` y la construcción del sitio, y recién entonces se mergea.
- **Después del merge:** #180 pasa a cerrado. Como existe `_support/workspace/maintainer.local`, el refresco local con `hive update` queda autorizado por `AGENTS.md`.
- **Registro del cambio:** se cierra con un PR pequeño sin revisión dedicada, que se mergea.
- **Fuera:** la release que publica el portal.
