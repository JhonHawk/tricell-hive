# #316 — Validador dedicado para las entradas del índice del visor

| Campo | Valor |
| --- | --- |
| Estado | Completado · integrado por el PR #320 (`e1ebb0c3`) el 2026-11-02 |
| Tracker | [JhonHawk/tricell-hive#316](https://github.com/JhonHawk/tricell-hive/issues/316) · cerrado · sin etiquetas · relacionado con #318 (mergeado) |
| Git | `tricell-hive`, base `development` en `87a394fa`; rama `refactor/gh-316-index-entry-validator` en `.claude/worktrees/gh-316-index-entry-validator` |

## Decisiones

2026-11-02
- ¿Se planea #316 y se cierra #317? Sí: #317 se cerró como «not planned» con la evidencia de la investigación (las aperturas venían de un visor de prueba en el puerto 4499 con una copia del estado del usuario, no del visor).
- ¿Cómo se entrega? En modo automático: revisión, push, pull request y merge sin parada de validación, porque no hay cambio visible que validar. (D1 A)
- ¿Qué revisión de código corre? La nativa de Claude Code, `/code-review`, sobre el candidato antes del push. (D2 A)
- ¿Se limpia la rama ya mergeada `fix/post-0.7.0-corrections`? Sí: el checkout principal pasó a `development` y la rama se borró en local y en `origin`. (D3 A)
- ¿Qué pasa con #316 al ampliar el alcance con el header, el retiro del respaldo heredado y la pregunta de apertura? #316 sigue con este plan; los tres puntos van aparte, en #319. (D4 A, D5 B)
- ¿Qué se hace con las cuatro reglas que ya daban un código distinto en Go y TypeScript antes del cambio? Se alinea lo posible: `validation_count` fuera de rango da `invalid_count` y más de 100 tickets da `array_required_or_limit` en ambos lenguajes. `archived` con tipo inválido sigue distinto, porque Go lo reporta desde su decodificación JSON genérica. (D6 B)

## Petición

«/flow-plan 316 y cerremos #317» (2026-11-02).

El issue #316, leído el 2026-11-02, propone: «Replace the synthetic session with a validator for index entries: fields, enums, the pointer/state relationship (`timelineStateRequiresForeground`, `timelineStateAllowsForeground`) and tickets, in both languages, sharing the cases in `testdata/timeline_contract/cases.json`.» No trae criterios de aceptación propios. Origen: hallazgo de `/code-review` sobre el arreglo de la biblioteca vacía, aplazado por la decisión D4 B.

## Objetivo

Hoy `DecodeTimelineIndexEntry` (Go) y `parseTimelineIndexEntry` (TypeScript) validan una entrada del índice armando una sesión ficticia sin eventos y validándola como sesión completa. Por eso necesitan tres casos especiales (puntero de etapa forzado a nulo, `validation_count` a 0 y el estado reescrito a `active`), y varias reglas de la entrada se aplican solo como efecto lateral, como los tickets duplicados que detecta la validación del catálogo.

Al terminar, cada lenguaje valida la entrada con un validador propio que nombra cada regla, sin sesión ficticia, y acepta y rechaza exactamente lo mismo que hoy, con el mismo código de error.

## Alcance y aceptación

Dentro: los dos decodificadores de entradas del índice, una función compartida por lenguaje para validar la lista de tickets (que hoy vive dentro de la validación del catálogo de la sesión) y los casos `index_entry` de `cases.json`.

Fuera: el formato del índice, lo que emite el servidor, la validación de sesiones completas, cualquier cambio de interfaz.

- AC1 — Ninguno de los dos decodificadores de entradas del índice construye ni valida una sesión: `DecodeTimelineIndexEntry` no llama a `DecodeTimelineSession` y `parseTimelineIndexEntry` no llama a `parseTimelineSession` · falso hoy porque ambos las llaman (`tooling/viewer/timeline_contract.go:1561`, `tooling/viewer/frontend/src/timeline-contract.ts:1552`).
- AC2 — Cada regla de la entrada tiene en `cases.json` al menos un caso `index_entry` que la rompe y fija su código de error, y Go y TypeScript dan ese mismo código, salvo `archived` con tipo inválido, cuyo código sale de la decodificación JSON genérica de Go (D6 B). Las reglas son: campo desconocido o ausente, `viewer_session_id`, `source`, `project_key`, `project_name`, `version`, `title`, `summary`, `state`, `archived`, `pinned`, `created_at`, `updated_at`, `validation_count`, tickets inválidos, ticket duplicado, URL de ticket repetida, más de 100 tickets, y el puntero de etapa (id inválido, ausente cuando el estado lo exige, presente cuando el estado no lo permite) · falso hoy porque `cases.json` tiene 8 casos `index_entry` y ninguno cubre esos campos ni los tickets.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio.

## Contexto verificado

- 2026-11-02, `origin/development` en `87a394fa`: el decodificador de Go (`timeline_contract.go:1509-1580`) arma la sesión ficticia y luego comprueba `validation_count` y el puntero. El de TypeScript (`timeline-contract.ts:1518-1592`) hace lo mismo con `parseTimelineSession`.
- Reglas de sesión que hoy alcanzan a la entrada (`ValidateTimelineSession`, `timeline_contract.go:1261-1302`, y `parseTimelineSession`, `timeline-contract.ts:1190-1248`): id SHA-256, `source`, `project_key` ≤ 2048, `project_name` ≤ 200, `version`, `title` ≤ 200, `summary` ≤ 4000, `state` del enum, fechas, tickets, y en TypeScript `archived`/`pinned` booleanos. Además `validateTimelineCatalog` (`timeline_catalog.go:274-310`, `timeline-catalog.ts:177-202`) valida cada ticket y rechaza identidad duplicada (`duplicate_identity`) y URL repetida (`entity_conflict`), y `ValidateTimelineV3Graph` aplica la regla del puntero (`foreground_required`).
- Las reglas de primer plano ya son funciones compartidas: `timelineStateRequiresForeground`/`timelineStateAllowsForeground` (`timeline_contract_v3.go:811-818`, `timeline-contract-v3.ts:1293-1298`).
- Consumidores: `viewIndex` en la CLI (`tooling/cli/view_command.go:651`) y la carga de la biblioteca (`tooling/viewer/frontend/src/api.ts:210`). Ambos runners de `cases.json` comparan el código de error (`timeline_contract_test.go:89`, `timeline-contract.test.ts:67`).
- La especificación `walkthrough-viewer` no describe cómo se valida una entrada del índice; el cambio no altera comportamiento observable, así que no lleva delta de especificación.

## Entrega

Modo `automatic`, solo para este trabajo: tras la suite completa y `/code-review` sobre el candidato, push de `refactor/gh-316-index-entry-validator`, pull request a `development` y merge, sin parada de validación. Antes del push se compara la rama con su base y se pregunta por cualquier commit ajeno. No autoriza promover a `master`, publicar una release ni tocar #319.

## Estado de revisión y avance

2026-11-02 — Plan escrito. Sin revisión de plan: el cambio conserva el contrato, los datos, la seguridad y las dependencias.
2026-11-02 — Integrado: PR #320 mergeado en `development` (`e1ebb0c3`), con los commits `e6bd135c`, `7a7bd84d`, `977cb85d` y `da9a9812`.
