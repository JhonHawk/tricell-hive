# viewer-local-actions-cli — Paridad de la API del visor con su UI

| Campo | Valor |
| --- | --- |
| Estado | Completado · integrado en development (PR #363, 1995e9ce) y desplegado |
| Tracker | ninguno |
| Git | tricell-hive · base `development` 192697cb · rama propuesta `feat/viewer-local-actions-cli` |

## Decisiones

2026-11-04
- **D13 A** — ¿Cómo se da paridad a la API del visor con su UI? Con comandos de CLI que actúan como `local_viewer` con origen `cli` registrado en la auditoría; la guía limita a los agentes a un pedido explícito del usuario que nombre el trabajo.
- **D14 A** — ¿La CLI también puede borrar un trabajo? Sí, con `--confirm <work-id>` (el motivo quedó solo en reasignar, D18 A); la guía lo trata como efecto crítico que se pregunta cada vez.
- **D17 A** — ¿Cómo sabe la CLI que el visor admite los comandos? Por una capacidad `local_actions_cli` en `/api/health`, sin subir la revisión del contrato.
- **D18 A** — ¿Qué acciones llevan motivo? Solo reasignar; las demás guardan el origen en su recibo y borrar en una auditoría privada que sobrevive.
- **D19 A** — ¿Cómo se recupera una respuesta incierta? Releyendo el estado, sin diario; la spec exime a estos comandos.
- **D15 B** — ¿Cómo se entrega? Automática: push, PR, merge y despliegue sin detenerse.
- **D16 A** — ¿Qué revisión corre antes del push? Claude Code `/code-review`, porque el cambio mueve una barrera de autorización.

## Solicitud

«Entonces nos faltan funciones en la api, todo lo que se puede hacer en ui, debe poderse en la api».

## Objetivo

Hoy cinco acciones del visor solo existen en el navegador: renombrar, archivar o restaurar, fijar o desfijar, borrar y reasignar la conducción. El servidor las exige con el actor `local_viewer` y la CLI rechaza ese actor. El 2026-11-04 cuatro trabajos quedaron en revisión con su PR fusionado porque la sesión que los conducía terminó y nadie podía cerrarlos sin la UI. Al terminar, cada una tiene un comando con el mismo efecto, y la auditoría distingue si vino del navegador o de la CLI.

## Alcance y aceptación

Dentro: las cinco acciones en la CLI, el origen en la auditoría, el frontend que valida el actor, el contrato de `hive view work`, la regla de la guía, la spec, la arquitectura, el README, `llms.txt`, el sitio y el CHANGELOG. Fuera: cambiar quién puede hacer las acciones de publicación nativa y cerrar los cuatro trabajos pendientes (se hace después con el comando nuevo, a pedido del usuario).

- AC1 — `hive view work rename|archive|restore|pin|unpin --work-id <id>` cambia el título, el archivo o la fijación del trabajo, y `hive view session get` lo muestra · falso hoy porque esos comandos no existen.
- AC2 — `hive view work recover --work-id <id> --request-id <id> --reason <texto>` acepta la solicitud pendiente y la vuelve conductora en un solo commit, y esa sesión puede después cerrar el trabajo · falso hoy porque la CLI rechaza `local_viewer` y no hay comando.
- AC3 — `hive view work delete --work-id <id>` sin `--confirm` igual al ID no envía nada y no cambia nada; con él, borra el trabajo · falso hoy porque no existe.
- AC4 — El origen `cli` queda en el evento de renombrar y reasignar (visible con `hive view session get`), en el recibo administrativo de archivar, restaurar y fijar, y en la auditoría privada de borrar; el navegador sigue funcionando sin enviar origen y la UI abre trabajos con acciones de la CLI · falso hoy porque el actor `local_viewer` no tiene origen y el validador del frontend rechaza claves extra.
- AC7 — Con un visor que no anuncia `local_actions_cli`, cada comando no envía nada y pide actualizar el visor · falso hoy porque la capacidad no existe.
- AC5 — `viewer-publication.md` dice que un agente usa estas acciones solo cuando el usuario lo pide nombrando el trabajo, y que borrar es un efecto crítico con su propia pregunta · falso hoy porque dice que los publicadores nativos nunca usan ese actor.
- AC6 — `hive view work contract`, `README.md`, `llms.txt`, `deployment-manager.md`, `site/…/viewer.md` y el CHANGELOG describen los comandos, y `hive view session update` y `delete` apuntan a ellos · falso hoy porque no existen.

## Prerrequisitos

Ninguno: el build no necesita nada fuera del repositorio.

## Contexto verificado

- `tooling/viewer/timeline_contract_v3.go:717`: `session_rename`, `session_archive`, `session_pin`, `session_delete` y `recover_lead` exigen `local_viewer`; `TimelineActor` solo tiene `kind` y `producer_id`.
- `tooling/cli/view_timeline.go:336-337`: la CLI rechaza un actor `local_viewer` («local_viewer_actor_forbidden_in_native_cli»).
- La CLI y el navegador usan el mismo token local (`X-Hive-Token`) y el alcance `X-Hive-Action-Scope: session`.
- `_support/openspec/specs/walkthrough-viewer/spec.md`: «Local viewer management actor» y «Original lead unavailable» definen estas acciones como del navegador; la reasignación no infiere que una sesión murió.
- `content/skills/flow-walkthrough/references/viewer-publication.md`: «Browser management/recovery uses local_viewer; native publishers never use that actor to bypass publication authority».

## Entrega

Automática (D15 B): push de `feat/viewer-local-actions-cli`, PR hacia `development`, merge sin detenerse y despliegue local (binario, contenido y reinicio del visor conservando puerto, estado y TTL). Revisión: Claude Code `/code-review` antes del push (D16 A).

## Estado de la revisión y avance

2026-11-04 — plan escrito y revisado por un hijo `hive-review-plan`: 14 hallazgos (R1–R14); aplicados todos, con D17, D18 y D19 resueltos por el usuario.
