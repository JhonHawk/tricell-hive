# Comprobar que cada delta archivado quedó aplicado al spec (2026-10-24)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · validado por el usuario (D17-A); T1, T4 y T5 integradas en `development` el 2026-10-24 por el [PR #137](https://github.com/JhonHawk/tricell-hive/pull/137), merge `39cfe446`; la prueba (T2), los deltas aplicados (T3) y este registro archivado entran por el PR de cierre (D11-A), que cierra #117 |
| Tracker · GitHub Issues | • [#117 — Cierre: comprobar que cada delta archivado quedó aplicado al spec](https://github.com/JhonHawk/tricell-hive/issues/117) |
| Git | `interactive` (D4-A, reutilizado en la sesión) · rama desde `development` · parada antes del push para que el usuario valide |
| Verificación | `hive-verify-task` por tarea · `go vet ./...`, `go test ./...` y build del sitio sobre el candidato final · prueba de deltas en verde tras archivar · sin revisión de código dedicada (D5-A, reutilizado) |
| Siguiente paso | Ninguno; registro cerrado |

## Solicitud

Respuesta del usuario a D8 (siguiente paso tras cerrar #134), 2026-10-24, tarjeta nativa: «Seguir con #117 ahora».

Respuestas siguientes, 2026-10-24, tarjetas nativas:

- D9-A: «Comprobación + prueba + base»: flow-close compara cada bloque con su delta antes de marcar Completado; una prueba en `tests/content` compara el delta más reciente de cada requisito con `specs/`; se reconcilian `context-delivery` y `walkthrough-viewer` para que la prueba arranque en verde. Elegir esta ruta es la aprobación del usuario de la corrección, según la regla de admisión de `AGENTS.md`.
- D10-A: comentario en #117 con H2 y H3 → [comentario](https://github.com/JhonHawk/tricell-hive/issues/117#issuecomment-6389088443).
- D11-A: «Prueba con el cierre»: el PR principal lleva la regla de flow-close, el sitio y el changelog; el PR de cierre lleva la prueba, los deltas aplicados y el registro archivado, así que `development` nunca queda en rojo.
- D12-A: incluir el arreglo de `_support/docs/architecture/deployment-manager.md:310` (`Agent.Info.name` frente a la resolución por identificador).

Acción sugerida en #117, citada textualmente (leído el 2026-10-24):

> - En `flow-close` paso 1, tras aplicar cada delta: comprobar que cada bloque ADDED o MODIFIED aparece igual en el spec vigente y que cada REMOVED ya no está, y no pasar el estado a Completado sin esa evidencia.
> - Opcional: una prueba en `tests/` que recorra `changes/archive/*/specs/**` y compare con `specs/` solo el delta más reciente de cada requisito, para detectar deltas sin aplicar.

## Objetivo

`flow-close` aplica los deltas de un cambio al spec vigente, pero nada comprueba después que el spec quedó como dice cada delta. Tres cambios archivados el 2026-10-15 marcaron su delta como aplicado sin aplicarlo. Dos ya se corrigieron (#115 y #116); el tercero, `2026-10-15-flash-class-guidance`, dejó sin crear `specs/context-delivery/spec.md` con sus cinco requisitos. Además, el requisito «Current source classification» de `walkthrough-viewer` se editó en el spec sin delta. Cuando el spec deja de coincidir con sus deltas, describe un comportamiento que ya no existe y obliga a fusionar a mano el siguiente cambio que toque el mismo requisito.

El cambio añade la comprobación al cierre, una prueba del repositorio que detecta cualquier delta sin aplicar aunque el agente se salte el paso, y pone al día la base para que esa prueba arranque en verde.

## Alcance y aceptación

Incluido: `content/skills/flow-close/SKILL.md` («Close the change record», pasos 1 y 2, y la línea «Abandon»); `content/skills/flow-plan/references/change-records.md` (inventario); una prueba nueva en `tests/content/`; deltas de este cambio para `context-delivery` (cinco ADDED) y `walkthrough-viewer` (un MODIFIED); `site/src/content/docs/flows/close.md`; `CHANGELOG.md`; `_support/docs/architecture/deployment-manager.md:310` (D12-A).

Excluido:
- Mapas de producto (`product/<module>/<view>.md`): Hive no usa uno; el paso de `flow-close` cubre ambos formatos en texto, pero la prueba solo recorre `specs/`.
- Cambiar los nombres de los cinco requisitos de `context-delivery`: se conservan los del delta archivado para que el delta más reciente de cada nombre sea el de este cambio; solo cambia su texto.

Restricciones:
- `content/skills/flow-close/SKILL.md` se mantiene en 16 000 bytes o menos.
- `go vet ./...`, `go test ./...` y el build del sitio pasan sobre el candidato final.
- `content/guidance/global.md` no cambia.

- AC1. El paso 1 de «Close the change record» exige, tras aplicar los deltas y revisar el diff, comprobar que cada requisito ADDED o MODIFIED aparece en el spec vigente con el mismo texto (salvo espacios) y que no queda el encabezado de ningún REMOVED, con el caso equivalente del mapa de producto, y corregir el spec cuando no coincide; el paso 2 solo marca Completado tras esa comprobación; un cambio abandonado archiva sus deltas bajo una carpeta con prefijo `unapplied-`, que `change-records.md` registra. *Falso en la base cuando* `content/skills/flow-close/SKILL.md` («Close the change record», paso 1) solo dice «Edit only those blocks, leave the rest of the file untouched, and review the resulting diff».
- AC2. Una prueba de `tests/content` falla cuando el delta archivado más reciente de un requisito no coincide con `_support/openspec/specs/`: requisito ADDED o MODIFIED ausente o con otro texto, o REMOVED todavía presente; ignora los requisitos que toca un cambio activo en `changes/<id>/specs/`. *Falso en la base cuando* no existe ninguna prueba que lea `changes/archive/*/specs/` (`rg -l "changes/archive" tests/` no devuelve nada).
- AC3. Con este cambio archivado y sus deltas aplicados, esa prueba pasa sobre el repositorio: `specs/context-delivery/spec.md` contiene los cinco requisitos reconciliados con el comportamiento actual y `walkthrough-viewer` coincide con su delta más reciente. *Falso en la base cuando* el escaneo de solo lectura del 2026-10-24 informa 6 requisitos que no coinciden (5 MISSING en `context-delivery`, 1 DIFFERS en `walkthrough-viewer`).
- AC4. `site/src/content/docs/flows/close.md` dice que el cierre comprueba que cada cambio de requisito quedó aplicado antes de marcarlo completado. *Falso en la base cuando* la línea 44 de esa página solo dice que integra los cambios de requisitos y archiva la carpeta.
- AC5. `_support/docs/architecture/deployment-manager.md` describe que el loader de OpenCode resuelve el rol por el identificador nativo, no por `Agent.Info.name`. *Falso en la base cuando* la línea 310 dice «`ctx.agent.get({agentID})` and `Agent.Info.name`».

## Entrega

- Repositorio: `tricell-hive`, base `development`. Modo `interactive` (D4-A, elegido en esta sesión para este repositorio): rama de trabajo, commits locales por tarea verificada, gates compartidos, parada antes del push para la validación del usuario; tras ella, push, PR, merge a `development` y `hive update` local por el marcador de maintainer.
- PR principal: T1, T4 y T5. PR de cierre (D11-A), que fusiona el agente: la prueba de T2, los deltas aplicados a `specs/` y la carpeta archivada; la prueba corre en verde sobre su candidato antes de fusionarlo, como evidencia de AC3.
- Revisión de código: ninguna dedicada (D5-A, reutilizado). No cruza ningún límite de seguridad.
- Ticket: #117 se cierra con el merge del PR principal (`Closes #117`).
