# Etapas propias para verificar, desplegar y revisar seguridad

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en `development` por PR #177 (`9b8ce042`) el 2026-10-25 |
| Tracker · GitHub Issues | • [#169 — Los límites entre etapas del flujo son difusos: separar verificación y despliegue en etapas propias](https://github.com/JhonHawk/tricell-hive/issues/169) |
| Git | `automatic` · PR a `development` con `/code-review` · sin parada humana |
| Verificación | conteos y lectura por tarea · `go test ./tests/...` afectados · `go vet ./...` y `go test ./...` sobre el candidato final · build del sitio |
| Siguiente paso | Ninguno; la promoción a `master` va aparte |

## Solicitud

Mensaje del usuario en la sesión `5470b05e` (2026-10-25):

> La premisa es la siguiente:
>
> - Actualmente tenemos en hive un flujo de desarrollo, sin embargo los limites de las etapas son difusos
>
> > flow-research -> flow-plan -> flow-build -> flow-close
>
> Aqui lo relevante es que parece que muchas etapas de evaluación comen tiempo, y no siempre es lo deseado (esto sumado a que faltaria esa etapa), y que tenemos skills que se consideran de apoyo, como el de git, walkthrough, etc
>
> El flujo que me imagino es por gates
>
> > flow-research -> flow-plan -> flow-build -> flow-verify -> flow-security -> flow-deploy -> flow-close
>
> - En flow-verify, se sacan las etapas de revisión de build y se combina con walkthrough, y trabajan de manera independiente a build
> - Claro que se pueden omitir etapas o iniciar directamente desde alguna en especifico
> - Seguiran existiendo las skill de apoyo que se consideren necesarias
> - En hive-legacy, se tenia una skill core de apoyo para todos los demas, no sé si se buena idea retomarlo

> COnsidera que el plan debe contemplar el ejecutar o no a verify y a que nivel (recuerda que tenemos varios)

Respuestas: R1-A (ticket #169; #167 y #168 primero, ya integrados), D15-B (todo #169 en un solo plan), D16-A (`flow-verify` es dueño de la etapa y `flow-walkthrough` queda como skill de apoyo sin mover sus rutas), D17-A (sin clave nueva: el nivel sale de los valores actuales), D18-A (`flow-security`), D19-A (`ui-review-criteria.md` se mueve con `ui-gate.md`), D21-A (pedir «sin verify» o bajar el nivel cuenta como aceptar sin verificar cada chequeo omitido, nombrado uno por uno), D22-A (la oferta de seguridad se queda en `security-boundaries.md`; `flow-security` corre la revisión), D23-A (`flow-deploy` es una etapa aparte, después de integrar y por encima de los tickets), D24-A (integra la última etapa que corre: `flow-verify` en el flujo normal y `flow-build` cuando verify se omite; las reglas de integración siguen solo en `git-workflow`).

Pregunta del usuario a D20, que D24 respondió: «Para el flujo normal si, pero qué pasa cuando son tareas acotadas que no pasan por el flujo completo? o se omite la verificación?»

Criterios del ticket [#169](https://github.com/JhonHawk/tricell-hive/issues/169), leídos el 2026-10-25:

> - AC1: existe `flow-verify` con su contrato de entrada (candidato identificado y criterios) y de salida, y `flow-build` termina al tener el candidato, sin la secuencia de verificación.
> - AC2: la puerta de UI y el walkthrough tienen un solo hogar dentro de la etapa de verificación.
> - AC3: existe `flow-deploy` con la promoción, predeploy y postdeploy que hoy están en `verification.md`, `git-workflow` y `delivery-decisions.md`.
> - AC4: la revisión de seguridad a pedido tiene una skill propia que se invoca solo cuando el usuario la pide.
> - AC5: README, `llms.txt`, el sitio y las pruebas de estructura (`tests/content/reading_footprint_test.go`, `tests/pilot/flows.go`, `tests/fixtures/flows/cases.json`) describen las nuevas etapas.

## Objetivo

El flujo de Hive tiene cuatro etapas y la verificación, la revisión de seguridad y el despliegue viven mezclados dentro de `flow-build`, `verification.md`, `flow-walkthrough` y `delivery-decisions.md`. El cambio da a cada una su etapa con contrato de entrada y salida: `flow-build` termina con un candidato construido y verificado por tarea; `flow-verify` lo lleva hasta su aceptación e integración, y decide al entrar si corre y con qué nivel; cuando se omite, integra `flow-build`; `flow-security` corre la revisión de seguridad solo a pedido; `flow-deploy` es una etapa aparte que promueve lo ya integrado, puede abarcar varios tickets, y verifica lo servido. El camino de un ticket es research → plan → build → verify → close. Las etapas se pueden omitir o iniciar directamente cuando se cumple su contrato de entrada. No se agrega una skill core: lo compartido sigue en la skill dueña de cada regla y en `global.md`.

## Alcance y aceptación

Incluido: tres skills nuevas (`flow-verify`, `flow-deploy`, `flow-security`), el movimiento del texto que les pertenece, los contratos de las siete etapas, las pruebas que fijan la estructura y la documentación. Excluido, con su razón:

- Una skill core de apoyo: la reconstrucción la descartó por baja lectura medida de cadenas índice → enrutador → referencia (`_support/docs/harness-engineering/2026-09-20-portable-harness-research.md:52`).
- Mover rutas del visor: `flow-walkthrough` conserva su carpeta, sus scripts y sus assets (D16-A).
- Una clave `Verify` nueva en `## Hive` (D17-A).
- La verificación por tarea durante el build (`plan-execution.md`, `hive-verify-task`): se queda en `flow-build`.
- #176 (detectores de la revisión del plan): ticket aparte.

Restricciones: todo `content/skills/*/SKILL.md` ≤ 16 000 bytes; `content/guidance/global.md` solo cambia en su puntero de integración y despliegue (`:78`) y en la regla de puertas (`:71`) para D21-A, y el total crece lo mínimo; el PR declara su tamaño; la suite actual pasa; ninguna regla queda con dos hogares.

- AC1. Existe `content/skills/flow-verify/SKILL.md` con contrato de entrada (candidato identificado: commits o árbol, tareas verificadas, valores de revisión resueltos) y de salida, y dueño de la secuencia previa a pedir validación y de la integración que el modo de entrega fija, siguiendo `git-workflow`; `flow-build/SKILL.md` ya no contiene esa secuencia, su salida entrega el candidato a `flow-verify` e integra solo cuando verify se omite (D24-A). *Falso en la base cuando* no existe `content/skills/flow-verify/` y `flow-build/SKILL.md:56-59` tiene la secuencia numerada.
- AC2. Al entrar, `flow-verify` dice en una línea si corre o se omite (cambio mecánico, o pedido del usuario) y con qué nivel, según una tabla única que combina `UI verification`, `Review`, la revisión de seguridad pedida, la suite completa y la prueba en vivo; el usuario puede cambiar el nivel por pedido, y omitir o bajar el nivel por pedido del usuario cuenta como aceptar sin verificar cada chequeo omitido, nombrado al entrar y en el informe, sin eximir CI, protección de ramas ni una revisión exigida (D21-A); en un cambio mecánico informa lo que ya lo cubre. *Falso en la base cuando* ninguna skill tiene esa decisión ni esa tabla.
- AC3. `ui-gate.md` y `ui-review-criteria.md` viven en `content/skills/flow-verify/references/` y ningún enlace apunta a sus rutas viejas. *Falso en la base cuando* los dos archivos están en `content/skills/flow-walkthrough/references/`.
- AC10. `flow-walkthrough` conserva presentación, visor y primera vista renderizada, y remite a `flow-verify` para la puerta de UI y las rondas de corrección. *Falso en la base cuando* `flow-walkthrough/SKILL.md:3` dice «own the UI gate» y su sección «Correction rounds and markers» está en `:58-66`.
- AC4. Los párrafos de revisión de código (refutación, severidades, re-revisión), prueba en vivo y entrega para validación humana de `verification.md` viven en una referencia de `flow-verify`; `verification.md` conserva la selección de evidencia compartida y remite a ella. *Falso en la base cuando* `verification.md:42-46,50-56,64-66` contienen esas reglas.
- AC5. Existe `content/skills/flow-deploy/SKILL.md` con contrato de entrada y salida y una referencia con predeploy, postdeploy, recorrido desplegado, tabla de plataformas y arreglos urgentes de producción; `verification.md`, `delivery-decisions.md`, `git-workflow/SKILL.md` y `global.md` remiten a ella; la descripción de `flow-build` ya no nombra promoción ni despliegue. *Falso en la base cuando* no existe `content/skills/flow-deploy/` y `delivery-decisions.md:40-46` tiene los arreglos urgentes.
- AC6. Existe `content/skills/flow-security/SKILL.md`, que corre la revisión de seguridad solo cuando el usuario la pide y es el único hogar de cómo correrla (alcance, brief a `hive-review-security`, informe, alcanzabilidad por puntero a `security-advisory`); la regla de ofrecerla se queda en `security-boundaries.md` y remite a `flow-security` para correrla (D22-A). *Falso en la base cuando* no existe `content/skills/flow-security/` y `security-boundaries.md:18` define cómo correr la revisión pedida.
- AC7. `flow-research`, `flow-plan`, `flow-build`, `flow-verify`, `flow-deploy` y `flow-close` declaran su contrato de entrada y salida, y cada contrato permite entrar directamente cuando se cumple. *Falso en la base cuando* ninguna de esas skills declara un contrato de entrada y salida.
- AC8. `tests/pilot/protected.go`, `TestAllManagedSkillsProtected` (`tests/pilot/flows_test.go`) y `tests/content/reading_footprint_test.go` incluyen las etapas nuevas, y `go test ./tests/...` pasa; los fixtures de pilotos pausados no cambian. *Falso en la base cuando* ninguna menciona `flow-verify`.
- AC9. `README.md`, `llms.txt`, el sitio (barra lateral, páginas de flujos, `skills.md`, `index.mdx`, `flows/build.md`) `CONTRIBUTING.md` y `CHANGELOG.md` describen las etapas nuevas, la seguridad a pedido y `flow-deploy` como etapa aparte. *Falso en la base cuando* `README.md` y `llms.txt` listan cuatro flujos.

Ticket AC1 → AC1, AC2 y AC7; ticket AC2 → AC3, AC4 y AC10; ticket AC3 → AC5; ticket AC4 → AC6; ticket AC5 → AC8 y AC9.

## Entrega

Valores declarados en `AGENTS.md` `## Hive` (base `development`): `Delivery: automatic` y `Review: /code-review`. `UI verification` no está declarado: el cambio agrega páginas y entradas a la barra lateral del sitio; se resuelve en la ronda de entrega. Worktree propuesto `.claude/worktrees/gh-169`, rama propuesta `feat/gh-169-stage-redesign`. La promoción a `master` queda fuera.

Temas abiertos: #176 (detectores de la revisión del plan) tiene ticket propio.
