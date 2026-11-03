# viewer-on-demand-publication — Publicación bajo demanda en el visor

| Campo | Valor |
| --- | --- |
| Estado | Completed · implementación integrada en PR #353 (c22ad325) |
| Tracker | Ninguno solicitado; no crear ticket por este plan |
| Git | `tricell-hive`; base `development` en `158638346b1fae8fc58ff1172273538a67ac0375`, verificada contra remoto el 2026-11-03; checkout de planificación `175a0263b0f15e14700e4432a84529ae8de71259` en `feat/gh-311-minimal-research-entry` |
| Destino | `_support/openspec/changes/viewer-on-demand-publication/` |

## Decisiones

2026-11-03:
- Conservar `changes/` y sus Markdown; sustituir la propuesta de watcher por una publicación explícita bajo demanda. Un trabajo puede compartir sesión nativa con otros, conservando identidad y revisiones propias. Decisiones expresadas por el usuario antes de invocar el plan.
- Coordinar la publicación en el CLI usando la API v3 existente, con recuperación local, manteniendo compatibilidad con los comandos de bajo nivel. La alternativa era un endpoint nuevo de publicación por lote. (D1 A)
- Mantener el renderer y las pestañas actuales. Navegación de enlaces relativos, HTML e imágenes incrustadas quedan fuera de este cambio. (D2 A)
- Entrega automática tras autorizar el build: revisar, push, abrir PR e integrar en development; no incluye una release ni promoción a master. (D3 A)
- Ejecutar revisión nativa Codex `/review` sobre el candidato antes del push. (D4 A)

## Solicitud

> igual y que no sea un watcher, si no que sea una petición, con la ruta de los archivos (pwd) y que con esta el binario escanee a demanda

> en la misma sesión ataco el ticket #10 [...] el agente vuelve a cargar al mismo uuid [...] continuamos con el ticket 11 en la misma sesión [...] se solicita una carga nueva y ahi se conservan los dos

El usuario invocó `$flow-plan` tras discutir el contrato y posteriormente autorizó la implementación con `$flow-build` y entrega automática. El ID existente del visor es opaco, actualmente un hash de 64 caracteres; no se convierte a UUID ni se migra la identidad almacenada.

## Objetivo

El agente publica una revisión de los documentos mediante una llamada pequeña al binario. El CLI lee los archivos locales, conserva los bytes de la operación, coordina uploads y snapshot, verifica el resultado y devuelve un resumen. El agente sigue declarando etapas, tareas y decisiones; el CLI gestiona los IDs técnicos, versiones y reintentos.

## Alcance y aceptación

Incluye un grupo aditivo `hive view work`, lectura acotada de una carpeta de cambio, registro local del publicador, selección y listado de trabajos, registro estructurado de tareas y transiciones breves. Reutiliza la API v3 y el visor existentes. La asociación carpeta/trabajo es explícita; `pwd` solo resuelve rutas relativas.

Quedan fuera: watcher o daemon, renovación de presencia/TTL, modificación del renderer, migración de almacenamiento del servidor, inferencia de progreso desde Markdown, inicio/reinicio automático del visor, consultas automáticas a trackers/CI, pilotos de modelos y publicación en producción. Los adjuntos de imágenes y los flujos avanzados de CI/colaboración mantienen la API existente; esta entrega publica documentos Markdown, incluidos extras seleccionados expresamente.

- AC1 — Una llamada de publicación por carpeta entrega los Markdown originales como una revisión confirmada y una respuesta JSON compacta, sin que el agente construya solicitudes HTTP, hashes o versiones. Falso en la base: `isTimelineViewCommand` y el parser no tienen `work`; el agente debe emitir uploads y snapshot por separado.
- AC2 — Dos carpetas y claves de trabajo dentro de una sesión se abren, listan y actualizan por la interfaz de alto nivel sin mezclarse; cambiar `pwd` no cambia la vinculación y una ruta incompatible falla. Falso en la base: la API distingue claves, pero no existe selección/vinculación local por carpeta ni listado de alto nivel.
- AC3 — Republicar un conjunto idéntico en el mismo contexto confirmado devuelve `unchanged` sin mutación remota; cambiar bytes, orden, etiqueta o contexto crea una revisión pertinente y conserva anteriores. Falso en la base: los uploads deduplicados todavía crean recibos/auditoría y no existe reconciliación semántica del CLI.
- AC4 — Una publicación interrumpida se recupera con solicitudes y bytes congelados; una respuesta perdida no duplica el trabajo ni la revisión. Falso en la base: la API permite replay, pero el CLI deja las solicitudes y operation IDs a cargo del productor.
- AC5 — El escaneo valida contención, tipos de archivo, enlaces, estabilidad y límites de todo el conjunto antes de publicar; un fallo no publica una revisión parcial. Falso en la base: existe lectura individual acotada, no preflight seguro de un conjunto bajo `--change`.
- AC6 — Registro y actualización de tareas/etapas por claves semánticas omiten IDs internos, versiones y operation IDs; republicar `tasks.md` no cambia estados. Falso en la base: el cliente de bajo nivel requiere esos campos y no resuelve claves de tareas/segmentos.
- AC7 — Dos procesos publicadores y una escritura concurrente externa no pierden actualizaciones ni sobrescriben una revisión incompatible; reportan `busy`, conflicto o resultado confirmado según corresponda. Falso en la base: no existe coordinación multiproceso del publicador, aunque el servidor protege operaciones individuales.
- AC8 — El cliente nuevo funciona con el servidor v3 actual sin endpoint ni migración nuevos; los comandos y datos actuales siguen funcionando, y el nuevo journal vive fuera del inventario `viewer/`. Falso en la base: el cliente nuevo y su almacenamiento no existen; la compatibilidad se verifica junto con su introducción, no como criterio aislado ya satisfecho.
- AC9 — La guía y el contrato compacto conducen al flujo de alto nivel para las operaciones cubiertas, y una prueba determinista ejecuta el ejemplo de dos tickets con una invocación por revisión, respuestas exitosas de hasta 4096 bytes y sin emitir documentos completos al stdout. Falso en la base: el procedimiento obliga a administrar el protocolo y no existe esa interfaz. No promete un porcentaje de tokens ni demuestra conducta de modelos.

## Prerrequisitos

No se requieren servicios, credenciales ni dependencias nuevas. La entrega y el code review están elegidos. Al iniciar build se reconcilia la base vigente; los checks locales ya existen.

## Contexto verificado

- `tooling/cli/view_timeline.go`: cliente de bajo nivel; requests JSON y uploads individuales; operación exacta a cargo del agente.
- `tooling/viewer/timeline_store.go`: work keys, replay, recibos, snapshots y límite de conflicto; `TimelineCreate` con operación nueva sobre trabajo existente también escribe auditoría.
- `tooling/viewer/server.go` y `tooling/cli/view_command.go`: `/api/index` y `viewIndex` permiten listado sin endpoint nuevo. El índice no incluye work_key; el cliente obtiene detalles cuando hacen falta.
- `tooling/viewer/timeline_contract.go`: 256 KiB por Markdown, 8 pestañas, 64 documentos únicos, 50 snapshots y 16 MiB agregados. Se mantienen los límites, sin truncar documentos ni dividir revisiones silenciosamente.
- `tooling/viewer/state_v3.go`: entradas ajenas bajo `viewer/` invalidan el inventario; el journal será hermano de esa carpeta.
- `tooling/viewer/frontend/src/timeline-markdown.tsx`: renderer existente, sin HTML ni imágenes, sin resolución local entre documentos.
- `content/skills/flow-walkthrough/references/viewer-publication.md`: propietario de la publicación. `flow-plan`, `flow-build`, `flow-verify` y `flow-close` dependen de él.
- Runtime y cliente inspeccionados coinciden con la base local en las áreas de implementación. El checkout contiene cambios de otra sesión en guía/research; no se entregarán con este trabajo por inferencia.
- `CONTRIBUTING.md`: suite local Go/Python; no CI de tests en PRs a development. Pilotos de modelos siguen pausados.

## Entrega

Entrega automática: una vez autorizado el build, ejecutar los checks y la revisión nativa Codex `/review`, resolver hallazgos, hacer push, abrir PR e integrar en `development` sin una parada adicional de validación. El endpoint es la integración y el refresh local que corresponda al marker del mantenedor, siguiendo sus límites; no autoriza una release ni promoción a `master`. La preferencia del repositorio por Claude Code `/code-review` no está disponible como revisión nativa en esta sesión. (D3 A, D4 A)

El build autorizado concluyó con la integración de la PR #353; el usuario solicitó archivar y limpiar el trabajo.

## Historial de revisión

- 2026-11-03: investigación y decisiones de enfoque completas. Plan revisado por `hive-review-plan`; dos hallazgos: salida para intención rechazada definitivamente (H1, P1), checks y preview del sitio (H2, P2). Corregidos en contrato, delta y tareas. El mismo revisor confirmó ambos resueltos y no encontró nuevos problemas materiales en su revisión focalizada; listo para handoff. Revisión por inspección, sin ejecutar tests de implementación. Ninguna tarea de implementación comenzó.

## Resultado final

Implementado y verificado: AC1–AC9 cumplidos. PR #353 integrada en development como c22ad325. Binario local actualizado; contenido pendiente para preservar instrucciones concurrentes instaladas desde56704a0a. Registro del visor cerrado y confirmado; servicio conservado.
