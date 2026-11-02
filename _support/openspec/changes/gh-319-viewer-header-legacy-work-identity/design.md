# Diseño

## Decisiones

- **Identidad de trabajo (D14).** Opciones: un `work_key` obligatorio, uno opcional o ningún campo nuevo. Se eligió el opcional. Con ticket, el agente envía la clave del ticket y un flujo posterior del mismo ticket retoma el trabajo. Sin ticket, el servidor genera la identidad en el momento a partir del `operation_id` de la petición, que el agente crea nuevo en cada operación. Se descartó un UUID aleatorio del servidor: con él, el reintento de una petición idéntica crearía un trabajo duplicado. Con `operation_id`, un reintento devuelve el mismo trabajo y una petición nueva crea uno distinto.
- **Validaciones (D15).** El número de validaciones pasa a la barra de herramientas como un contador más.
- **Apertura (D16).** Se pregunta en cada parada con algo que revisar, hasta que el usuario acepte una vez en la conversación.
- **Guardas (D11 B).** Se borra el código de las tres guardas. El estado viejo no se tolera: sus nombres caen en el inventario estricto de `state_v3.go:91-98` y fallan con el `invalid_storage` genérico antes de escribir. Tolerarlos obligaba a mantener la lista de nombres viejos y debilitaba la detección de estado corrupto.

## Forma propuesta

### `create_session` con `work_key`

- **Petición.** Gana `work_key`, opcional: texto de 1 a 128 caracteres sin espacios ni controles, el mismo patrón que `external_id`. Go: `WorkKey *string \`json:"work_key,omitempty"\``. TypeScript: clave opcional en la validación de la petición.
- **Derivación del ID.** La función `timelineSessionKey` (`timeline_store.go:55`) recibe la clave efectiva:
  - Vacía, solo en registros guardados antes del cambio: `["timeline-api-v3", proyecto, host, nativa]`, que es la derivación actual, así que esos registros cargan con el mismo ID.
  - `work_key` presente: `["timeline-api-v3", proyecto, host, nativa, "work", work_key]`.
  - `work_key` ausente: `["timeline-api-v3", proyecto, host, nativa, "operation", operation_id]`.
- **Registro guardado.** Gana `work_key` (string, vacío en los antiguos) y la carga re-deriva el ID con ese valor (`:218`). Se guarda la clave efectiva (`work:<key>` u `operation:<id>`), para que la re-derivación no dependa de la petición original.
- **Salida.** `TimelineSession` gana `work_key` nullable: la clave que dio la petición, o `null` cuando no dio ninguna, incluidos los registros antiguos. La clave efectiva `operation:<id>` vive solo en el registro guardado. La CLI comprueba que el `work_key` de la respuesta coincida con el de la petición, junto al chequeo de host y sesión nativa (`view_timeline.go:431`). Se valida en Go (`ValidateTimelineSession`) y en TypeScript (`parseTimelineSession`, clave en la lista de claves exactas). El índice no cambia.
- **Alcance de la clave.** La sesión nativa forma parte de la identidad: la misma clave retoma el trabajo solo desde la misma sesión nativa. Otra sesión nativa sigue uniéndose por `join_request`.
- **Retomar.** Un `create_session` sobre un trabajo existente sigue sin reescribir metadatos ni reabrir un trabajo cerrado, como ya fija la spec «Required native session registration».
- **Contrato offline.** `hive view contract` (`timeline_protocol.go:55-61`) incluye `work_key` en el `input_schema` de `create_session` y en el ejemplo.

### Retiro del legacy

- **Se conserva o se mueve**, porque v3 lo sigue usando: `RunOptions` y `Store.Run` (`scan.go:34,45`, bucle de `serve`), `validateSessionTitle` (`session_groups.go:61`), `PartialDeleteError` (`session_delete.go:15`), `newIndexData`, `IngestError`, `resolveProject` y `NormalizeRemote` (`index.go`), `Open`, `repositoryStatuses` y `writeJSONAtomic` (`store.go`), y `isHex`, `contains`, `safeDeliveryURL`, `segmentPattern` y `maxCapture` (`manifest.go`). Un archivo que queda solo con esos símbolos puede renombrarse; no hace falta.
- **Se borra** lo que lista «Contexto verificado» en `proposal.md`, más:
  - `tests/skills/walkthrough_{render,status}_test.py` y `tests/fixtures/walkthrough/`;
  - `content/skills/flow-walkthrough/{references/walkthrough-manifest.md,scripts/render_walkthrough.py,assets/walkthrough.*}`;
  - los tests Go que solo cubren lo borrado: `manifest_test.go`, `store_test.go`, `view_source_tracking_test.go`, `view_status_test.go` en su parte `--dir`, y la copia de fixtures de `view_command_test.go:459`;
  - los casos de corte de `state_v3_test.go` y `view_cutoff_v2_test.go`.
- **Guardas:**
  - `server.go` deja de enrutar `/api/sessions/`, que pasa a responder 404 por la ruta por defecto.
  - `timeline_contract.go` pierde `TimelineLegacyIndexEntry`, `TimelineLegacySession`, la rama reflect de `legacy` y el `case "legacy"`, así que `mode:"legacy"` cae en `unsupported_mode`.
  - TypeScript hace lo mismo en `timeline-contract.ts:306-324,1535`.
  - `state_v3.go` borra el `case` de nombres legacy, la excepción de `registry.json` sin marcador y el corte de `storage_version` 1 y 2; todos caen en `invalid_storage`. El inventario estricto queda igual.
  - `timeline_protocol.go` quita `cutoff_required` de los códigos de error (`:87`) y la invariante del corte (`:105`).
- **Casos fijados.** Los tres casos legacy de `cases.json` y `legacy-index-entry.json` pasan a esperar `unsupported_mode`. Ese cambio de contrato (Go, TypeScript y fixtures) va en T1, junto con `work_key`, para que T2 y T3 no compartan archivos.
- **`--ttl`.** Sale de `start` y `serve` (`view_command.go:112,135,164`) y del texto de ayuda. Los campos de salida `ttl`/`ttl_seconds` quedan, fuera de alcance.

### Guidance

- **`flow-walkthrough/SKILL.md`.** Pierde la línea 12 (fallback). La línea 10 deja de decir que la skill vuelve «sin agregar una pregunta» y pasa a apuntar a la regla de apertura.
- **`viewer-publication.md`:**
  - **Paso 2.** Envía la clave del ticket como `work_key` cuando el trabajo tiene ticket. Sin ticket la omite, y retoma el trabajo por el `viewer_session_id` guardado, nunca con otro `create_session`.
  - **Línea 11.** Pierde la excepción del binario ausente: sin binario, la publicación no está disponible y se informa.
  - **Línea 45.** Se reemplaza por la regla de apertura (abajo).
  - **Línea 51.** Se borra.
- **Menciones a quitar:**
  - `flow-close/SKILL.md:10`: «or a legacy walkthrough» y la cláusula completa sobre el walkthrough legacy («a legacy walkthrough, which holds only `in_progress`…»);
  - `workspace-conventions/references/artifact-layout.md:23-25`;
  - `flow-report`, si menciona el fallback.
- **Regla de apertura**, con hogar canónico en `viewer-publication.md`:
  - Nunca abrir el visor por cuenta propia.
  - En cada parada donde el visor guarda algo para que el usuario lo revise (un plan en su parada de aprobación, el reporte de una investigación, un walkthrough de UI), preguntar «¿Abrir el visor?» hasta que el usuario acepte una vez en esta conversación.
  - Donde la parada ya tiene una pregunta en tarjeta, va como una pregunta más de esa tarjeta. Donde termina con el bloque de siguiente paso, va como una línea más de ese bloque.
  - Al aceptar, abrir la sesión una vez con `hive view open --session-id ID`. Después no volver a preguntar ni a abrir, salvo pedido explícito.
  - Si el usuario rechaza, preguntar de nuevo en la siguiente parada.
  - La primera vista de una UI en `flow-build` (paso 4) cuenta como walkthrough de UI.
- **Punteros** de una línea a esa regla, como permite `AGENTS.md`, donde se escribe la pregunta: `flow-plan/references/plan-handoff.md`, el paso de reporte de `flow-research`, el paso 4 de `flow-build` y el paso 2 de `flow-verify`. `viewer-publication.md` tiene 11.654 bytes; el cambio la deja bajo 16.000.
- **Publicación de tareas (D19).** «Register once» de `viewer-publication.md` gana su momento: al entrar a build, antes de la primera tarea, registrar una vez las tareas del plan con su `task_key` (`T<n>`) y su `plan_order`; al empezar y al terminar cada tarea, cambiar el estado de su segmento. El puntero va en el párrafo de `flow-build` que pide la lista nativa de tareas del host («When the host has a native task list…»), pero como frase aparte, fuera de esa condición, para que valga en todo host, tenga o no lista nativa: al empezar el build se registran las tareas en el visor, y al empezar y al terminar cada tarea se cambia su segmento. Son los mismos momentos de la lista nativa donde exista. Se escribe por capacidad, sin nombrar la herramienta de ningún host. La lista nativa es para el terminal y el visor para el navegador; una no reemplaza a la otra, y el visor no puede leer la lista nativa.
- **Preservación.** Los HTML standalone ya generados y las unidades de artefactos que el usuario eligió no se borran ni se reescriben al quitar el renderer.

## UI

- **Patrón.** El header actual de `TimelineHeader` se reordena con los mismos componentes ReUI y shadcn ya instalados: `Badge` outline del estado, `ButtonGroup` con `Button` `outline` `sm` y `Popover` para tickets, Git y participantes, `IconAction` para archivar y «más».
- **Componentes.** No se agregan componentes nuevos, ni alturas, radios ni movimiento propios.
- **Primera fila.** `h1` del título y el botón de editar (sin cambios en el modo de edición).
- **Segunda fila.** `TimelineSummary` con `lines={2}`, el mismo componente que hoy está en el canvas, a la izquierda (`min-w-0 flex-1`), y la barra de herramientas a la derecha (`shrink-0`, `justify-end`). Por debajo de `lg` la barra pasa abajo del resumen, alineada a la derecha, sin overflow de página.
- **Contador de validaciones.** Va como `Button` outline `sm` con ícono y número, con nombre accesible «Validaciones: N» / «Validations: N». No abre popover.
- **Estados.** La carga usa el skeleton del header que ya existe. El error de guardado del título se mantiene. El resumen es obligatorio en el contrato, así que no hay estado vacío.
- **Canvas.** Pierde el `TimelineSummary` de `timeline-content.tsx:590` y empieza en la barra de Tareas.

Antes:

```
cam (workspace `..`)
COM-907 · actualización de dependencias [✎]
2 nov 2026 · 2 validaciones
           [● Validación][🎫1][⑂0][👥1][▣][⋯]
──────────────────────────────────────────────
Decisión de alcance: separar HeroUI por la
ventana de 7 días…                    Ver más
[Tareas 5/7] [Ir al actual]
```

Después (desde `lg`):

```
COM-907 · actualización de dependencias [✎]
Decisión de alcance:       [● Validación][✓2]
separar HeroUI… Ver más     [🎫1][⑂0][👥1][▣][⋯]
──────────────────────────────────────────────
[Tareas 5/7] [Ir al actual]
```

En pantallas angostas:

```
COM-907 · actualización de dep… [✎]
Decisión de alcance: separar HeroUI
por la ventana de 7 días… Ver más
   [● Validación][✓2][🎫1][⑂0][👥1][▣][⋯]
```

## Textos

- Nombre accesible del contador: es «Validaciones: {n}» · en «Validations: {n}». Reutiliza la clave `validations` que ya existe en la copia del timeline.
- Pregunta de apertura (guidance, en el idioma de la sesión): «¿Abrir el visor?» · «Open the viewer?». Línea del bloque de siguiente paso: «o responde *abre el visor*» · «or reply *open the viewer*».
- Se elimina la copia del estado vacío que menciona «saving walkthrough.json» (`messages.ts:104,223`) y se reemplaza por la del flujo v3 que ya existe en `timeline-copy`.

## Impacto

- **Contrato.** `create_session` gana un campo opcional y la salida de sesión gana `work_key`. El frontend y el servidor viajan en el mismo binario, así que el parser estricto de TypeScript se actualiza a la vez.
- **Guidance vieja.** Una guidance desplegada antes del cambio no envía `work_key` y retoma con `create_session`. Con el binario nuevo, esa guidance crearía un trabajo nuevo en cada publicación con `operation_id` distinto. `hive update` despliega guidance y binario juntos, y no hay instalaciones públicas.
- **Datos guardados.** Los registros existentes cargan con su ID sin migración. El estado real del mantenedor no tiene nombres legacy.
- **Producción.** No hay.

## Recuperación

Revertir el merge restaura todo el código y la guidance. Los registros creados con `work_key` después del cambio no cargarían en un binario anterior, porque su ID re-derivado no coincidiría. Antes de volver a un binario anterior hay que respaldar `timeline-api-v3/` y quitar esos registros.
