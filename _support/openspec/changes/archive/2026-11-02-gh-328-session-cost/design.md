# Diseño

## Decisiones

- **Dónde vive la regla de las acciones sueltas.** En `global.md` «Questions and routes», en la misma viñeta del bloque de siguiente paso (línea 66), porque esa viñeta ya prohíbe la tarjeta para cambiar de flujo: la excepción tiene que estar al lado de la prohibición. Las skills solo apuntan.
- **Bloque y tarjeta en el mismo mensaje.** Cuando el mensaje termina en un bloque de siguiente paso, la acción suelta es una respuesta más del bloque, como hoy «abrir el visor» (`viewer-publication.md:47`). Sin bloque, se pregunta como decisión de la ronda; es el caso del ticket inexistente, porque el bloque todavía no puede nombrarlo.
- **Escrituras ya autorizadas.** Mover el ticket a en curso al empezar y a terminado al integrarse ya lo autoriza «Scope and authorization»; no se preguntan.
- **Compactar en una parada de ruta.** El usuario compacta; después el agente vuelve a cerrar con el mismo bloque y espera la respuesta. Esa compactación no es «during work» en el sentido de `global.md:99`: el siguiente paso es la respuesta del usuario. Se nombra el comando por capacidad («el comando de compactación del host»), sin tabla por host.
- **Dónde va la compactación después de un cierre.** `flow-close` termina con una tarjeta, no con un bloque. La respuesta de compactar va en el bloque del informe que sigue cuando el usuario elige otro trabajo: el de `flow-status` tras «Find the next ticket», o el de cualquier research o plan que empiece otro trabajo en la misma sesión. La regla vive en `global.md`; `flow-close` y `plan-handoff.md` solo apuntan. Elegir el ticket que reutiliza el contexto sigue igual: la línea de qué conservar nombra los archivos y decisiones que ese ticket reutiliza.
- **Compatibilidad de la salida del CLI.** Pasar de JSON con sangría a JSON compacto cambia bytes, no contenido: quien parsea JSON no nota diferencia. `--operation` devuelve los mismos campos. Se mantiene compatible.
- **Lectura de vuelta.** El acuse de cada cambio (versiones e IDs que devuelve el servidor) confirma esa escritura. La sesión completa se relee antes de decirle al usuario que algo está visible, al cerrar y al retomar tras una interrupción o un conflicto de versión. Esto acota, sin quitar, la garantía de la spec.
- **`reason` y los esquemas.** Varias operaciones exigen `reason` aunque el esquema no la marca como obligatoria, porque los esquemas salen de los DTO de Go y las reglas semánticas viven en los validadores (`timeline_protocol.go:18-19`). El error pasa a nombrar el campo. La segunda revisión de código mostró que el resumen y el contrato completo no coincidían si solo el resumen la marcaba: el esquema de `stage_start` y `stage_update` pasa a requerir `reason`, que es lo que el validador ya exigía.

## Forma propuesta

### Mapa flujo → operación

Declarado en Go junto a las definiciones del contrato (`tooling/viewer/timeline_protocol.go`), con una prueba que falla si una operación no tiene flujo o nombra un flujo desconocido.

| Flujo | Operaciones |
| --- | --- |
| común (se suma a research, plan, build, verify y delivery) | `create_session`, `session_get`, `session_patch`, `stage_start`, `stage_update`, `append_event`, `ticket_create` |
| research | comunes + `document_add`, `document_snapshot`, `document_get`, `ticket_patch` |
| plan | las de research |
| build | comunes + `event_patch`, `event_withdraw`, `task_create`, `task_patch`, `task_transition`, `candidate_create`, `artifact_add` |
| verify | comunes + `event_patch`, `event_withdraw`, `task_transition`, `candidate_create`, `artifact_add`, `document_add`, `document_snapshot` |
| delivery | comunes + `candidate_create`, `delivery_create`, `delivery_patch`, `wait_create`, `wait_attempt_start`, `wait_observe`, `task_transition` |
| close | `stage_update`, `session_patch`, `session_get`, `document_get` |
| collab | `join_request`, `producer_accept`, `foreground_transfer`, `recover_lead`, `session_delete` |

### Comando

```
hive view contract                      # completo, como hoy, sin sangría
hive view contract --operation NAME     # una operación completa, sin sangría
hive view contract --for FLOW           # resumen de las operaciones del flujo
hive view contract --list               # una entrada por operación
hive view contract --section rules      # bloque común
```

Las opciones son excluyentes; combinarlas, o un flujo o una sección desconocidos, devuelve un error de uso que lista los valores válidos.

Resumen de `--for`:

```json
{"protocol_version":3,"flow":"plan","operations":[{"operation":"stage_update","usage":"hive view stage update <session-id> <event-id> --json <file|->","required":["actor","schema_version","operation_id","expected_version","stage_state","expected_session_version"],"optional":["reason","summary","title"],"enums":{"stage_state":["in_progress","in_review","awaiting_validation","completed","paused","failed"]},"example":{…}}]}
```

`enums` recorre las propiedades del esquema de entrada, incluidas las de objetos anidados con la ruta separada por puntos (`source.host`). No incluye `output_schema`.

`--list`: `{"protocol_version":3,"operations":[{"operation":"stage_update","flows":["research","plan","build","verify","delivery","close"],"usage":"…"}]}`.

`--section rules`: `protocol_version`, `storage_version`, `schema_dialect`, `localization`, `capabilities`, `retention`, `ttl_seconds`, `authority`, `error_codes`, `limits`, `invariants`.

Toda la salida de `hive view` que pasa por `viewTimelineOutput` sale sin sangría.

### Error de validación

`invalid timeline input: <campo>: <código>`, tomando `Field` de `TimelineContractError` con `errors.As`; sin ese tipo, solo el código, como hoy. Nunca `err.Error()` de un error de decodificación, que puede repetir la entrada (`timeline_contract.go:64`). El caso `unsupported_protocol` conserva su mensaje.

### Texto de las reglas

`global.md` «Questions and routes», viñeta del bloque de siguiente paso, en sustancia:
- El comando de un flujo se ofrece solo cuando el paso siguiente necesita su procedimiento: investigar, planear, o cambiar archivos del repositorio con su entrega.
- Una acción suelta que no cambia archivos del repositorio y no necesita procedimiento de flujo (crear, mover o cerrar un ticket, borrar una rama integrada, guardar en memoria, abrir el visor) se pregunta como decisión de la ronda; la respuesta autoriza solo esa acción. Si el flujo siguiente necesita un ticket que no existe, se pregunta si se crea y, ya creado, el bloque nombra el ticket real.
- Entre las otras respuestas, el bloque ofrece compactar primero (el comando de compactación del host, con lo que conservar donde el host acepta instrucciones: la carpeta del cambio, los criterios aprobados y las respuestas de entrega, o los archivos y decisiones que reutiliza el trabajo siguiente) cuando lleva de un plan aprobado al build en la misma sesión, o cuando empieza otro trabajo después de un cierre. Después de esa compactación el agente cierra otra vez con el mismo bloque y no empieza el flujo. Nunca según un porcentaje de la ventana, que el modelo no ve.
- La lista de acciones sueltas excluye las escrituras que «Scope and authorization» ya autoriza.

`flow-walkthrough` y `viewer-publication.md`: cada flujo lee el resumen de su etapa (`flow-verify`, que publica verify y delivery, lee ambos; `flow-close` lee close); lectura única de la referencia (una vez por sesión y después de compactar), contrato por flujo, sin `session get` tras cada cambio, y la relectura en los momentos que fija la decisión de lectura de vuelta.

## Impacto

- **Guía distribuida:** `global.md` crece unos cientos de bytes (se reporta el tamaño en el PR); `plan-handoff.md`, `flow-close`, `flow-research` (la frase «until the user chooses a route»), `flow-walkthrough` y su referencia.
- **CLI:** `tooling/cli/view_timeline.go`, `tooling/cli/main.go:385` (ayuda), `tooling/viewer/timeline_protocol.go` y sus pruebas. No cambia el protocolo v3, los esquemas ni el servidor.
- **Documentación:** `README.md:421-422`, `llms.txt:54` (revisión), `_support/docs/architecture/deployment-manager.md:123,128`, `_support/docs/architecture/instruction-resources.md:19` (revisión), `site/src/content/docs/viewer.md` y `site/src/content/docs/flows/{plan,close,status,research}.md`.
- **Spec:** delta de `walkthrough-viewer`.
- **No toca producción:** el visor es local; las sesiones abiertas cargan la guía nueva al reiniciar.

## Recuperación

Revertir el merge restaura la guía y el CLI anteriores. Nada guarda datos nuevos: el contrato es metadata offline.
