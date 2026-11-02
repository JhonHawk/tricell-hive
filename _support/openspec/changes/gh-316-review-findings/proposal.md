# #316 — Hallazgos de la revisión del validador de entradas del índice

| Campo | Valor |
| --- | --- |
| Estado | Listo para flow-build |
| Tracker | [JhonHawk/tricell-hive#316](https://github.com/JhonHawk/tricell-hive/issues/316) · cerrado (PR #320); estos son sus hallazgos de cierre, sin ticket propio |
| Git | `tricell-hive`, base `development` en `93b0ce79`; rama `fix/gh-316-review-findings` en `.claude/worktrees/gh-316-review-findings` |

## Decisiones

2026-11-02
- ¿Se resuelven los hallazgos del cierre de #316? Sí, todos juntos en una sola rama, a pedido del usuario.
- ¿Cómo se entrega? En modo automático: revisión, push, pull request y merge sin parada de validación. (D7 A)
- ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`, sobre el candidato antes del push. (D8 A)

## Petición

«crea una rama para resolver los hallazgos, porfavor, todos juntos en una rama» (2026-11-02). Los hallazgos son los del cierre de #316:

- H1 (P2): los validadores de entrada del índice copian unos 12 chequeos de cabecera de la validación de sesión, en Go y en TypeScript.
- H2 (P3): con varios errores, Go y TypeScript pueden dar distinto primer código, porque el puntero de etapa y `validation_count` se revisan en otro orden.
- H3 (P3): cada ticket se valida dos veces, en ambos lenguajes.
- H4 (P3): en las sesiones y en el bootstrap de TypeScript, más de 100 tickets da `invalid_array`, mientras Go da `array_required_or_limit`.
- H5 (P3): ningún caso fija el orden límite-antes-que-identidades de los tickets.
- H6 (P3): `TestExecApplyPiInterruptKillsDescendants` falla a veces con `strconv.Atoi: parsing "": invalid syntax`.
- H7 (P3): `gofmt -l` marca `tests/pilot/regression.go`.

## Objetivo

Hoy las reglas de cabecera viven dos veces por lenguaje, los tickets se validan dos veces, los dos lenguajes no coinciden en el primer error de una entrada con varios fallos ni en el código del límite de tickets fuera del índice, un test de `tooling/management` es inestable y `gofmt` no queda limpio. Al terminar, cada regla de cabecera vive una vez por lenguaje, los tickets se validan una vez, Go y TypeScript dan el mismo primer código en las entradas con varios fallos y en el límite de tickets, el test es determinista y `gofmt -l` no imprime nada.

## Alcance y aceptación

Dentro: los validadores de sesión y de entrada del índice en Go y TypeScript, el registro de identidades de tickets, el helper `tickets()` de TypeScript para listas completas, `cases.json`, el test `TestExecApplyPiInterruptKillsDescendants` y un comentario de `tests/pilot/regression.go`.

Fuera: `archived` con tipo inválido y `tickets` que no es lista, que siguen distintos porque Go los reporta desde su decodificación JSON genérica; las referencias de tickets de los eventos (`ticket_refs`); cualquier cambio de formato o de interfaz.

- AC1 — En cada lenguaje, la validación de sesión y la de entrada del índice llaman a una misma función para los campos de cabecera (`viewer_session_id`, `source`, `project_key`, `project_name`, `version`, `title`, `summary`, `state`, fechas, y en TypeScript `archived`/`pinned`) · falso hoy porque están copiados (`tooling/viewer/timeline_contract.go:1268-1297` y `1542-1571`; `tooling/viewer/frontend/src/timeline-contract.ts:1229-1245` y `1552-1568`).
- AC2 — Cada ticket se valida una sola vez por sesión o entrada: el registro de identidades ya no vuelve a validar el ticket · falso hoy porque `addTickets` llama a `timelineTicketInput` (`timeline_catalog.go:306`) y a `validateTimelineTicketRecord` (`timeline-catalog.ts:197`) después de que la lista ya pasó esa validación.
- AC3 — Una entrada del índice con dos fallos da el mismo código en Go y en TypeScript, y lo fijan dos casos: puntero con id inválido más `project_key` vacío da `invalid_text`; sesión cerrada con puntero más `validation_count` -1 da `invalid_count` · falso hoy porque TypeScript revisa antes el id del puntero (`invalid_id`) y la regla de primer plano (`foreground_requires_active_session`).
- AC4 — Una sesión con más de 100 tickets, uno de ellos duplicado, da `array_required_or_limit` en Go y en TypeScript, y lo fija un caso; el bootstrap de creación con más de 100 tickets también da `array_required_or_limit` en ambos; y un caso fija lo mismo para la entrada del índice · falso hoy porque Go revisa las identidades antes del límite (`duplicate_identity`) y TypeScript da `invalid_array`.
- AC5 — El falso `pi` del test publica su PID de forma atómica, y `TestExecApplyPiInterruptKillsDescendants` pasa 50 veces seguidas en paralelo con el resto del paquete · falso hoy porque el script escribe con `echo $! > archivo`, que deja ver el archivo vacío; falló una vez bajo carga el 2026-11-02.
- AC6 — `gofmt -l tooling/ tests/` no imprime nada · falso hoy porque imprime `tests/pilot/regression.go`.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- 2026-11-02, `origin/development` en `93b0ce79` (incluye #320 y #321).
- Orden en Go para la entrada del índice: puntero exigido, cabecera, límite y registros de tickets, identidades, `validation_count`, id del puntero, puntero permitido (`timeline_contract.go:1538-1585`). En TypeScript: claves exactas, id del puntero, puntero exigido, cabecera, límite especial, tickets, identidades, puntero permitido, `validation_count` (`timeline-contract.ts:1545-1595`).
- En la sesión, Go llama a `validateTimelineCatalog` (identidades) antes de `timelineTickets` (límite) (`timeline_contract.go:1298-1303`); TypeScript llama a `tickets()` antes de `validateTimelineCatalogs` (`timeline-contract.ts:1247-1248`).
- `tickets()` de TypeScript usa `array(value, "tickets", 100)`, que falla con `invalid_array`; lo usan la sesión, la entrada del índice, el bootstrap (`timeline-contract.ts:925`) y las referencias de eventos (`:751`, `:798`). Go da `array_required_or_limit` en `timelineTickets` y `timelineTicketBootstrap` (`timeline_contract.go:715-740`).
- `fakeHangingPi` escribe `echo $! > "$HIVE_TEST_PID_FILE"` y el test espera solo a que el archivo exista (`tooling/management/pi_package_exec_test.go:23` y `:94-101`), así que `readPID` puede leerlo vacío.
- El comentario de `tests/pilot/regression.go:1126-1129` contiene `` `` `` ``, que `gofmt` reescribe con comillas tipográficas.
- La especificación `walkthrough-viewer` no fija códigos de error; los códigos que cambian solo los emite el parser del cliente (TypeScript) ante datos inválidos del servidor. No lleva delta de especificación.

## Entrega

Modo `automatic`, solo para este trabajo: tras la suite completa y `/code-review` sobre el candidato, push de `fix/gh-316-review-findings`, pull request a `development` y merge, sin parada de validación. Antes del push se compara la rama con su base y se pregunta por cualquier commit ajeno. No autoriza promover a `master`, publicar una release ni tocar #319.

## Estado de revisión y avance

2026-11-02 — Plan escrito.
2026-11-02 — Revisión del plan (`hive-review-plan`): nada bloqueante; confirma que cada criterio es falso hoy y que ningún otro llamador ni consumidor depende de la validación o los códigos que cambian. Ajustes aplicados: la verificación de T1 dice qué falla en cada lenguaje; un quinto caso fija el orden límite-antes-que-identidades en la entrada del índice; Go reutiliza `TimelineMaxCatalogRecords`; T4 corre solo el test de AC5 y registra el riesgo restante del test de timeout; la conservación pasa también el parser TypeScript sobre el índice real.
