# Design

## Decisions

D62 B asociación por servidor; D63 A Sheet; D68 A Markdown versionado común; D69 A trabajo compartido; D70 A estados/motivos; D71 A protocolo nuevo; D72 A recuperación explícita desde el visor; D73 A tareas con expansión breve y Sheet; D74 A tarea global con proyecciones filtradas por etapa y nodo de espera CI en Entrega. Etiquetas cortas y orden ascendente pedidos por el usuario.

Decisiones menores: conservar viewer_session_id como ID opaco del trabajo y source original inmutable; no renombrar comandos session ni recalcular IDs. Una incorporación no fusiona registros. Productor es procedencia declarada, no autenticación criptográfica del proceso; la capability local sigue siendo la frontera. El ID Engram siempre es el runtime propio, nunca el ID del trabajo.

## Proposed shape

### Protocol and storage — clean cutoff

La instrucción posterior del usuario reemplaza lectura compatible/migración v1. Protocolo, schema y storage_version finales son 3 exclusivamente; la ampliación de referencias supersede el objetivo provisional 2 y evita un handshake adicional de subrevisión. Endpoints/CLI siguen nombres session pero anuncian v3 en health/index/contrato offline. Cliente o escritura incompatible falla antes de mutar con unsupported_protocol/diagnóstico de actualización. No replay de escrituras v1/v2, no adaptador de lectura ni conversión lazy.

El runtime v3 usa un directorio de estado nuevo y vacío: sin importar API v1/v2, preferencias/copias legacy o walkthrough.json. El registro de raíces del proyecto se hace de nuevo en el estado v3. Su index y frontend exponen solo trabajos API v3; no hay scanner legacy en esa ruta. Un estado anterior reconocido hace fallar el inicio con cutoff_required antes de crear/cambiar archivos. Un estado corrupto/desconocido falla explícitamente, no se considera vacío ni se limpia.

No hace falta migrar las pocas sesiones del visor anterior. Este plan no borra su directorio: durante verificación se mantiene el runtime habitual y el candidato usa puerto/estado propios. El corte real debe declarar puerto, rutas anterior/nueva, escritores y guía instalada, parar publicaciones antiguas y cambiar servicio/CLI/content juntos bajo la autorización aplicable. Se muestra advertencia de que el nuevo visor inicia vacío; trabajo activo continúa por publicación nueva explícita con tupla real, no importación ni atribución ficticia. No se borra ni reescribe evidencia fuente del proyecto. El fallback HTML standalone cuando falta el binario no es importación legacy y conserva su propia utilidad.

Índice final v3, conservando el envelope de T2/T4: exactamente protocol_version:3, retention:"manual", ttl_seconds:0, timeline_sessions:TimelineAPIIndexEntry[] requerido, repositories:{root,absent}[] requerido y errors:{path,reason,code?}[] requerido. Los grupos se derivan de los trabajos; no projects/tickets/sessions/excluded_sessions/tombstones legacy, ni siquiera placeholders vacíos. Es metadato operativo más trabajos v3, no un adaptador de compatibilidad.

Recibos/digests/identidades/bounds, commits atómicos y outcome_unknown se verifican para v3 únicamente. No se añaden pruebas de recuperación de formato v1; sí de rechazo seguro del directorio previo, parejas incompatibles y reinicio/crash v3. Los tests de funciones retiradas se eliminan o se reescriben con su razón; no se cambian tests de semántica v3 para hacerlos pasar.

### Shared work

Forma que T1 fija en schemas Go/TS antes de UI:
- Session conserva source y viewer_session_id; añade producers[] con producer_id de servidor/source real, foreground_producer_id y pending_join_requests[].
- Crear por tupla original sigue idempotente. Continuar otro trabajo exige ID explícito confirmado por usuario/handoff autorizado, nunca búsqueda automática por ticket.
- join_request lleva source real, operation_id y expected_session_version; no confiere membresía. El conductor confirma con producer_accept. Ambas operaciones idempotentes, con recibo durable. Un productor puede participar en varios trabajos explícitos sin unir sus registros.
- Mutaciones de productores nativos v3 llevan producer_id aceptado. El servidor valida pertenencia, proyecto y autoridad. join_request es la única excepción nativa a membresía, autenticada por capability local y acotada al mismo proyecto registrado.
- Salvo la recuperación local D72, solo conductor acepta miembros, transfiere conducción, modifica estado global, inicia/reanuda/termina etapas, retira etapas activas o cierra. Transferencia exige expected_session_version y miembro aceptado; no cambia contenido ni etapa.
- Colaboradores aceptados publican hallazgos/revisión/documentos/capturas contra etapa explícita sin modificar foreground. Solo editan/retiran su propio contenido; conductor puede moderarlo con auditoría, nunca modificar bytes inmutables.
- Actor nativo se guarda en contenido y en auditoría para cada mutación nativa, incluidos uploads y cambios de texto; recibo privado conserva digest. ActorSource se obtiene de membresía, no de texto libre del evento. No afirmar identidad de proceso autenticada.

Mutex/commit durable único cubre estado, autoridad, evento y recibo. Stale version falla sin overwrite. Exact retry devuelve recibo; payload cambiado bajo mismo ID falla. Límites: 16 productores aceptados y 16 solicitudes pendientes; límites de eventos existentes se amplían explícitamente en T1 para reservar 8 eventos de cierre/transición sin eliminar historial. Si no cabe auditoría, falla también la mutación.

### Local viewer actions

El contrato de actor es una unión estricta: native_producer con producer_id aceptado para publicación/lifecycle; local_viewer para renombrar, archivar/restaurar, fijar/desfijar y eliminar. local_viewer no lleva ni inventa native_session_id/producer_id, no publica aprobaciones ni documentos, no conduce etapas ni modifica estado declarado por sí solo; la recuperación D72 solo cambia membresía/conductor, conserva los estados y tiene operación separada. API autenticada con capability local, Origin y scope actuales; no se afirma que un humano haya sido autenticado ni que el proceso navegador sea imposible de automatizar. Actor describe el canal de acción, no prueba identidad personal.

Cada acción conserva el control condicional actual y emite auditoría local antes de confirmar; eliminar conserva un recibo/tombstone auditable después de borrar el contenido, sin resucitarlo. La confirmación visual existente de eliminación se conserva. CLI nativa no usa local_viewer para eludir permisos. Revisar todos los handlers de mutación del controller además de CLI; entrada con ambas ramas de actor se rechaza.

### Unavailable lead — D72 A

Operación recover_lead de actor local_viewer, separada de producer_accept y foreground_transfer nativos. Exige work ID, expected_session_version, request_id de incorporación pendiente del mismo proyecto, reason y operation_id. El servidor comprueba que la solicitud sigue pendiente y source real coincide con proyecto; acepta al solicitante, le asigna productor si procede, cambia foreground_producer_id y registra incorporación/recuperación/antiguo-nuevo conductor en un solo commit durable. No toma una identidad arbitraria del payload ni fusiona el registro propio de la sesión entrante. Un retry exacto recupera el mismo resultado tras reinicio; request cancelada o stale version falla sin efectos.

No exige demostrar que el proceso previo murió y no infiere abandono por tiempo. La acción del visor es autorización local explícita, no autenticación personal demostrada. Conserva source, etapas, estado declarado, documentos, capturas e IDs. El anterior sigue como colaborador, pero sus posteriores operaciones reservadas al conductor se rechazan. No lo mata ni revoca la capability compartida. Si vuelve, lee el estado y no recupera automáticamente la conducción. No permitir recuperación de trabajo closed/failed ni reapertura implícita: requiere la ruta de reanudación explícita ya existente, con auditoría local separada si corresponde; el alcance de recover_lead es trabajo no terminal.

UI: en Popover Sesiones mostrar solicitudes pendientes. Una acción Button outline/sm “Reasignar” abre Dialog instalado con conductor actual, host/ID nativo completo del solicitante y razón obligatoria. Una acción primaria “Reasignar” y Cancelar. El dato visible se relee antes de confirmar; ante conflicto no reintentar automáticamente, conservar razón y mostrar error con Actualizar. Tras éxito, cerrar, devolver foco y mostrar nuevo conductor. Sin solicitudes, texto y Actualizar, no entrada libre de IDs. No hay alerta ni confirmación automática en cada handoff ordinario; esta acción existe para la recuperación que el usuario eligió.

### Association and chronology

stage_event_id es salida del servidor. target_stage_event_id opcional de entrada debe referir etapa del mismo trabajo no retirada; colaboradores lo requieren. Conductor sin destino usa foreground o última etapa no retirada anterior. Etapas, closure y actividad global no llevan asociación. Auditoría de etapa sí se asocia a su etapa.

Orden por sequence ascendente: eventos previos, grupos por inicio, cierre al final. Reanudar etapa antigua conserva su grupo y agrega eventos dentro, por lo que timestamps globales entre grupos pueden no ser monótonos: la UI no promete orden temporal global de filas anidadas. Asignaciones tarea/segmento son explícitas; no hay fallback de agrupación v1. Eventos v3 anteriores a toda etapa quedan en el grupo inicial; asociaciones explícitas sobreviven retirada de etapa.

### Lifecycle and immutable audit

Etapas: in_progress, in_review, awaiting_validation, paused, completed, failed.
Trabajo: active, in_review, awaiting_validation, paused, closed, failed.
Labels: En curso, Revisión, Validación, Pausado, Finalizado, Fallido. closed con resultado abandoned sigue mostrando motivo de abandono, sin afirmar éxito.

Solo un foreground en in_progress/in_review/awaiting_validation. Esos estados conducidos sincronizan estado de trabajo y etapa atómicamente; in_review/awaiting_validation conservan puntero. Paused/completed/failed limpian puntero; paused deja trabajo paused, completar etapa deja active entre etapas, fallar etapa deja trabajo paused con motivo hasta recuperación o cierre failed. active sin etapa se permite al crear o entre etapas. closed/failed no tienen foreground.

Cada transición distinta exige reason seguro de 1–500 caracteres; prev/next, hora y productor los calcula servidor. target opcional a decisión/documento/evidencia. Cada mutación registra auditoría inmutable; state_transition, producer_joined y foreground_transferred son eventos solo de servidor y no editables/retirables. No-op de estado no genera eventos; retry exacto devuelve el recibo original.

Revisión paralela es evento de revisión con review_status in_progress/completed/failed sobre etapa explícita, sin cambiar estado de trabajo. Revisión de etapa significa que la actividad conductora es revisar. Antes de preguntar, estado Validación con motivo/destino; respuesta comunicada apunta al destino y luego se reanuda explícitamente. No nuevas preguntas ni aprobación por silencio.

completed/failed son terminales; volver a esa etapa crea ocurrencia enlazada. Pausado/Revisión/Validación se reanudan en la misma ocurrencia. Cierre requiere foreground vacío. Estados son declaraciones, no liveness.

### Document chains

Snapshot v3 añade document_key ASCII 1–128, no ruta/URL, y predecessor_revision_id para revisiones posteriores. Clave única por trabajo con propietario stage-kind/root-occurrence. Una etapa admite varias cadenas. Primera revisión fija raíz; continuación en otra ocurrencia exige mismo tipo de etapa y cadena related_event_id anterior del mismo trabajo. Concurrencia exige current predecessor y expected versions; no ramas divergentes.

Toda cadena v3 declara su clave y predecesor; no se derivan claves de documentos históricos ni se mezclan por nombre de pestaña o fecha.

Comparar label/document_id detecta pestañas cambiadas/nuevas/quitadas, no diff de líneas. Primera dice Revisión inicial. Aviso solo por sucesor válido en misma cadena; revisión abierta no cambia sin acción. Conservar tab por label; al desaparecer seleccionar primera y anunciar fallback. Solo planes con estructura OpenSpec agrupan deltas, no etiquetas arbitrarias de otros documentos.

### Work-owned tasks and stage segments — D74 A

La tarea pertenece al trabajo, nunca exclusivamente a una ocurrencia de etapa. task_id de servidor estable, task_key T<n> único por trabajo, title corto 1–120, task_version, plan_revision_id cuando existe y segments[] explícitos. El conductor declara todo el plan, incluidas tareas futuras; el navegador no parsea Markdown ni deduce etapa por título, orden, foreground o checklist.

Cada segmento tiene segment_id estable, planned_stage de research/plan/build/verify/delivery, order positivo único dentro de su etapa de presentación, state, assigned_producer_id aceptado y display_stage_event_id opcional hasta que exista la ocurrencia real. Una tarea normal tiene un segmento; una tarea como T7 de CAM puede declarar verify y delivery sin convertirse en T7a/T7b ni aumentar el total global. La asignación viene del alcance del plan. El tramo delivery de T7 puede incluir la reconciliación de cierre; closure sigue siendo evento terminal, no otra etapa.

Al abrir una etapa real, conductor vincula explícitamente sus segmentos pendientes con la ocurrencia del mismo tipo mediante task_patch condicional. Una etapa muestra únicamente segmentos vinculados a ella, cada fila con task_key/title y el estado de ese tramo. Tareas futuras sin ocurrencia cuentan globalmente y se consultan en el documento del plan; no se muestran dentro de Build ni crean nodos falsos de etapa. Una tarea puede tener proyección en Verify y Delivery si fue asignada a ambas: son referencias al mismo task_id, no entidades duplicadas. Se elimina la propuesta previa de una única fila principal excluyente de las otras etapas; esta es la corrección de presentación posterior a D74.

Editar título/orden/asignación no cambia IDs; cambiar el alcance de un segmento requiere conductor, versiones, revisión de plan/referencia de decisión y motivo auditado. No hay traslado o clonación por cambiar foreground. Cada segmento tiene como máximo una ocurrencia de presentación actual. Reentrada sustantiva de etapa usa ocurrencia enlazada; un tramo completado que se reabre se vincula explícitamente a esa nueva ocurrencia y conserva en historial la ocurrencia/estado anterior, sin borrar evidencia ni contar dos veces la tarea global.

### Segment states, aggregate and authority

Estados de tramo: not_started, in_progress, in_review, blocked, completed. Cada transición se declara, nunca se infiere de commits/tests o posición. completed del tramo Verify solo significa que terminó esa parte, no que acabó toda T7. Estado global de tarea se calcula de sus tramos requeridos vigentes: blocked si alguno bloqueado; en otro caso in_review si alguno en revisión; completed solo si todos completados; not_started si todos no iniciados; in_progress en el resto. Es un agregado de estados explícitos, no porcentaje ni aprobación inferida. Trabajo/etapa conservan lifecycle propio, sin cascadas automáticas.

Contador global X/Y usa task_id únicos vigentes con todos los tramos requeridos completos. Cada etapa cuenta solo sus segmentos vinculados, completados/total. El mismo T7 puede contar 1/1 en Verify y 0/1 en Delivery mientras el trabajo sigue 6/7. No sumar denominadores de etapas para calcular el global. Tareas/tramos retirados requieren motivo y rastro del plan, excluidos del conteo vigente pero conservados en auditoría. Sin tareas/tramos no mostrar 0/0.

Task_create/task_patch/task_transition usan native_producer, operation_id, expected_session_version y expected_task_version al modificar. Registro/asignación/retirada/binding solo conductor. Estado de segmento: conductor o productor asignado a ese tramo; otros miembros pueden publicar hallazgos contra task/segment sin cambiar estado. local_viewer solo lectura de tareas. Traspaso de conducción conserva asignaciones; al crear cada tramo, asignado por defecto al conductor creador. No asumir nueva autoría al transferir.

Matriz por tramo: not_started→in_progress o blocked; in_progress→in_review, blocked o completed; in_review→in_progress, blocked o completed; blocked→in_progress con resolución; completed→in_progress con motivo de reapertura y ocurrencia nueva si la etapa anterior ya era terminal. Un no-op no crea auditoría. Registro inicia not_started. Finalizar una etapa/trabajo no completa sus segmentos por sí solo.

blocked requiere reason 1–500, steps[] 1–10 acciones de hasta 500 caracteres, owner no vacío y resume_when comprobable. Payload incompleto falla sin mutar; resolución conserva instrucciones anteriores en auditoría. Cada transición produce task_state_transition inmutable con task_id, segment_id, etapa real, prev/next, reason, actor y hora servidor; retry exacto devuelve recibo. Una espera humana usa in_review/review_kind human en el tramo y awaiting_validation en etapa/trabajo conductor; revisión de agente paralela no toma foreground. CI usa un nodo de espera propio, no finge revisión humana.

### Task evidence across stages

Evidencia/documentos pueden referir task_id y segment_id: ambos existen y pertenecen al trabajo; el evento lleva la etapa real de publicación, que coincide con el tipo/asignación del segmento y su ocurrencia vinculada. No se compara contra una supuesta etapa única de la tarea. Contenido de Verify y Delivery de T7 tiene distintos orígenes, un mismo task_id y segmentos explícitos. Referencias cruzadas de trabajo, segmento de otra tarea o etapa no asignada fallan. Evidencia compartida usa referencias explícitas, no copias ni reemplazos implícitos.

El Sheet de tarea muestra estado global, tramos y actividad con su etapa de origen; abierto desde un nodo enfoca inicialmente ese tramo. Cada expansión muestra solo contenido pertinente a su proyección, sin duplicarlo como hermano de la etapa. El mismo blob puede referirse sin duplicar bytes; decisiones siguen apuntando a evidence_id/document_revision_id concretos. Adjuntos son opcionales, incluso en review; corregir o completar conserva historia.

Actividad de cada proyección y actividad general sin tarea van por sequence ascendente, tareas por orden declarado dentro de etapa. No afirmar cronología global estricta entre grupos. Abrir Sheet captura task_version/tramos/bloqueo/referencias; solo nueva task_version dispara aviso manual, no session.version. Toda mutación que cambie el snapshot visible de una tarea incrementa su task_version atómicamente con el contenido/auditoría/recibo: estado/copy/asignación de cualquier tramo, vínculo/retirada/edición de evidencia referenciada, nueva revisión documental enlazada y creación/actualización/promoción de intento CI vinculado. Si una referencia compartida afecta a varias tareas, incrementar cada afectada una vez en el mismo commit. Upload de bytes aún no vinculado, cambio de otra tarea y retry exacto no incrementan esa versión. No-op de contenido idéntico tampoco. La mutación falla completa si no puede persistir el cambio y las versiones afectadas; no hay invalidación asíncrona que pueda perderse tras restart. Foco/expansión/scroll se guardan por trabajo, task y segmento; un cambio de estado nunca salta lectura.

Límites: 100 tareas/trabajo, hasta 5 segmentos/tarea, uno por tipo de etapa, task_key 128, title 120 y 10 pasos de bloqueo; todo dentro de payload/content budget y auditoría existentes, con reserva de cierre. Al agotar capacidad falla sin commit parcial. Timeline instalado es stepper por activeStep; no usarlo para inferir estados independientes. Prueba obligatoria de completadas no contiguas y global sin doble conteo.

### Delivery stage and CI wait

Añadir delivery al enum de etapas, tipos Go/TS, protocolo, fixtures, copy y guías. La secuencia habitual es research→plan→build→verify→delivery y closure; se permite reentrada explícita enlazada por correcciones. Entrega inicia tras salir de Verify y conserva conductor único. Distinguir revisión individual de tarea en Build de la etapa Verify, y espera externa de CI de espera al usuario.

Un wait de tipo ci tiene wait_id estable de servidor, wait_version, stage_event_id delivery requerido, task_id/segment_id opcionales y coherentes con tramo delivery, candidato con commit/base/rama, PR/recurso identificado y enlaces HTTPS seguros. Contiene required_checks observados, su fuente/fecha de consulta y run/attempt IDs/resultados acotados. El productor usa CLI/MCP autenticado del proveedor; no consulta anónima de repo privado, no cambia protecciones/CI, no secretos en URLs. Este diseño no exige polling/daemon ni un proveedor nuevo: el agente publica observaciones reales cuando cambian.

Estados normalizados de espera: waiting, succeeded, failed, cancelled, unknown. Nodo waiting usa pausa, succeeded check, failed/cancelled alerta, unknown signo de interrogación; tooltip/nombre accesible cortos. Límites de espera: 50 waits por trabajo, 20 intentos por wait, 50 checks por intento y 500 caracteres por texto de resultado/acción; snapshots/historia cuentan dentro del budget de contenido existente. Falla por capacidad sin purgar historia ni impedir cierre. Estado raw del proveedor queda acotado y como dato en el Sheet, sin ejecutar texto. La etapa Entrega puede seguir En curso mientras el nodo espera; no marcar trabajo Validación cuando no se espera una respuesta humana.

Wait_start/wait_update son mutaciones condicionadas por versiones y actor conductor/productor asignado al tramo. Una espera se identifica por trabajo/etapa/PR/commit y conjunto de checks requerido: actualizar el mismo wait_id sin crear nodos por cada consulta. Los intentos y resultados publicados son snapshots inmutables con timestamps observados/servidor; cambios tienen auditoría y recibo. Rerun del mismo candidato conserva nodo y agrega intento; cambio de commit o conjunto requerido crea una espera nueva enlazada y no reutiliza el resultado anterior. Resultados para commit distinto fallan. No datos del proveedor o sin permisos: unknown con causa, nunca success por ausencia de checks.

Cada check requerido tiene una clave estable de política/proveedor y un applicable_attempt_id explícito. wait_attempt_start registra una nueva generación asignada por servidor con los IDs externos observados y la fuente que confirma que es la ejecución vigente, y cambia atómicamente el mapa de checks afectados a ese intento, reseteando su resultado a waiting/unknown; no arrastra éxito previo a checks reejecutados. La pertenencia del check al conjunto requerido se valida por clave, no solo por texto visible.

Una observación siempre incluye attempt_id/check_key/IDs externos y solo puede modificar el resultado actual si coincide con applicable_attempt_id. Resultados de intentos anteriores pueden agregarse a historia y actualizar la versión visible, pero nunca cambiar el mapa aplicable ni recuperar success. Observaciones de intento desconocido no se autopromueven: requieren inicio/promoción explícitos y confirmación de vigencia; si el productor no puede demostrar qué intento rige, el estado queda unknown. Un intento anterior conocido no se puede promover de nuevo. Una nueva consulta tardía con expected_version vigente tampoco elude esta regla. El estado del wait se calcula únicamente sobre observaciones de intentos aplicables por check, con fuente/tiempo confirmados; el orden de llegada no decide vigencia. El mapa/generaciones y todas las observaciones sobreviven restart.

succeeded exige confirmar todos los checks requeridos del candidato actual conforme a política real del proveedor/base; la consulta incluye autoridad/origen y momento de esa política, no inferir que los checks visibles son todos los obligatorios. Si la política observada confirma que no hay checks requeridos y no hay CI pendiente, no inventar espera ni éxito: registrar “Sin CI requerido” como observación de entrega; si hay checks opcionales en ejecución, pueden observarse pero no se presentan como regla obligatoria. Política no accesible es unknown, no ausencia de requisitos. Resultado fallido/cancelado/pendiente/desconocido no habilita integrar por sí mismo. Incluso succeeded solo declara CI observado: merge exige aceptación/revisiones/autoridad de entrega dadas y evidencia real de integración. El nodo no ejecuta merge por sí mismo ni dispara CI automáticamente.

Errores de CI no convierten automáticamente una tarea en blocked sin pasos: productor publica fallo de espera y luego, si corresponde, transición bloqueada con desbloqueo completo. Sheet de espera muestra candidato/PR/checks/última observación/intentos/enlaces y next_action seguro cuando falla o falta acceso. Durante espera, leer no implica que el proceso siga vivo. Sin cambios al binario/servicio/guía que usa CAM.

## UI

ReUI: superficie operativa; arquetipo detalle; familia Frame existente; densidad default; encabezado → resumen → historial con tareas y Sheet común documental/de tarea; sin animación nueva. Seguir timeline-header.tsx, timeline-content.tsx y timeline-documents.tsx. Installed: Timeline ReUI, Badge default, Button outline/sm o ghost/icon-sm, Popover, Collapsible, Sheet, Tabs, Select, Tooltip y Spinner instalados; iconos Lucide existentes. No dependencias nuevas.

Encabezado: botones de icono + contador, sin etiquetas largas visibles, con aria-label y ayuda completa; Popover de sesiones siguiendo tickets/entregas: participantes de lectura y recuperación D72 solo si hay una solicitud pendiente. Conductor/host/ID corto, ID completo copiable sin credenciales. Reasignar cambia autoridad del trabajo, nunca controla procesos. Badge default: warning-light En curso/Validación, info-light Revisión, outline Pausado, success-light Finalizado, destructive-light Fallido. Motivo separado text-sm; productor text-xs.

```
Antes
[En curso] [1] [0]    0 recorridos
Resumen sin límite

Después
[Validación] [1 ticket] [0 entregas] [Sesiones: 2]
Espera tu respuesta sobre el plan.
Resumen a 2 líneas [Ver más]
```

Historial: toolbar horizontal con Tareas (Popover instalado del registro global) e Ir al actual (Button outline/sm). El total global se muestra en el Popover, sin columna separada. Inicio explícito arriba, derivado de session.created_at/source sin publicar un evento ficticio; polling conserva viewport. Tareas por orden del plan dentro de cada etapa, actividad general por sequence después de ellas. Deep link prevalece. Volver a trabajo conserva scroll por ID. Actual/Revisión/Validación/Pausado y etapas con tareas activas/bloqueadas abiertos; finalizados/fallidos plegados por defecto pero expandibles. Cabecera muestra último evento resumido, no crea copia ficticia. Cambios de estado auditados son filas breves con productor.

```
Antes                    Después
Corrección               Historial       [Ir al actual]
Validación               Solicitud inicial
Build                    Investigación [Finalizado] >
Plan                     Plan [Finalizado] >
Investigación            Build [En curso]
                           Primera vista
                           Cambios pedidos
                           Segunda vista
                         Cierre (cuando exista)
```

```
Antes                    Después
Plan [En curso]          Plan [Validación]
[Ver plan]               Espera tu respuesta
                         [Ver plan]
                         En curso → Revisión · Codex
                         Revisión → Validación · Codex
```

Sheet: ancho móvil completo/desktop 2/3, excepción solo de width porque SheetContent no ofrece size/variant. Un cierre nativo, Escape y retorno de foco. Header/tabs fijos; cuerpo con scroll propio; tablas scroll horizontal local. Selector de documento solo si varias cadenas, selector de revisión, indicador Publicado único; no implica aprobación. General mantiene etiquetas; OpenSpec agrupa Specs.

```
Antes
[Dialog 40rem]
Proposal Design Tasks spec-1 spec-2
Completo                Completo

Después
[Sheet a la derecha                         X]
Investigación / Plan / Build / Verificación / Entrega
Documento: [Resultado v]  Revisión: [02:45 v]
Publicado
Hay una revisión nueva      [Ver la última]
Cambió: Design; nuevo: Riesgos; quitado: Notas
[Proposal] [Design*] [Tasks] [Especificaciones]
Cuerpo Markdown con scroll propio
```

Estados regionales: Skeleton en carga; error inline Reintentar; región sin datos Actualizar; sin docs no hay apertura; snapshot vacío “Sin contenido” con reintento de lectura. Sin foreground Ir al actual va al último evento; sin eventos deshabilitado. Referencias inaccesibles no filtran contenido ajeno. Sin controles de aprobación en navegador. Claro/oscuro y teclado a 1280/390/320.

Recuperación (patrón existente Popover + Dialog):

```
Antes
Sin participantes ni recuperación

Después
Sesiones: 2
Conductor: Claude · cfb6fcae…
Solicitud: Codex · 01a19638…       [Reasignar]
  Dialog: Reasignar conducción
  De: Claude · ID completo
  A: Codex · ID completo
  Motivo: [________________________]
  [Cancelar] [Reasignar]
```

Carga Skeleton; sin solicitudes Actualizar; error inline Reintentar lectura; conflicto Actualizar sin autoenvío. No aprobación de documentos desde este diálogo.

**Tareas — fila compacta y expansión independiente.** Reutilizar TimelineIndicator con estados explícitos y componentes Spinner/Tooltip, Collapsible por tarea. El estado solo vive en el punto; no Badge ni etiqueta repetida en fila. Título es Button de texto que abre Sheet solo cuando hay contenido significativo; si solo hay descripción/metadatos, alterna la lectura en línea. Los accesos de imágenes/documentos/bloqueo/historial/CI son visibles antes de expandir. La flecha es CollapsibleTrigger con nombre Expandir/Contraer T<n> y aria-expanded. Ningún botón contiene otro. Tooltip del punto y nombre accesible de tarea comunican estado sin depender de color. No spinner para tareas pendientes o bloqueadas. Spinner instalado, sin retimar; reduced motion muestra icono estático y texto accesible.

```
Antes
Build
  Primera vista
  Corrección

Después · contraído
Build                         2/5 finalizadas
  ✓ T1 — Contrato del dashboard                 >
  ✓ T2 — Composición de pantalla                >
  ◌ T3 — Consultas y permisos                   >
  Ⅱ T4 — Revisar actividad reciente             >
  ! T5 — Integración con proveedor              >
```

○ = No iniciada, ◌ = loader En curso, Ⅱ = Revisión, ! = Bloqueada, ✓ = Finalizada. Spinner no prueba liveness. No iniciadas muestran ○; el ejemplo de cinco tareas tiene dos finalizadas. Cada tarea empieza contraída al montar su etapa por primera vez; el polling no abre/cierra tareas por cambio de estado. La etapa con tareas pendientes activas no se pliega automáticamente.

```
Expandido
Build                         2/5 finalizadas
  ✓ T1 — Contrato del dashboard                 >
  ✓ T2 — Composición de pantalla                >
  ◌ T3 — Consultas y permisos                   >
  Ⅱ T4 — Revisar actividad reciente             v
    Espera tu revisión de la tabla.
    [Captura escritorio] [Captura móvil]
  ! T5 — Integración con proveedor              v
    Falta acceso al entorno de pruebas.
    [Ver cómo desbloquear]
```

Expansión: resumen máximo dos líneas, evidencia pertinente y última actividad; sin adjuntos no inventar placeholders de captura. Abrir imágenes conserva el visor de evidencia existente. Expansiones independientes, teclado/foco y scroll se preservan por work/task ID, incluso al cerrar Sheet. En móvil título completo refluye, estado permanece solo en indicador, controles accesibles sin overflow de página.

**Sheet de tarea.** Usar el mismo contenedor Sheet, ancho/cierre/foco del panel común; encabezado task_key/title con indicador, sin Badge redundante. Tabs Resumen/Evidencia/Historial (no pestañas documentales fijas del plan). Bloqueo se presenta primero en Resumen antes del resto. Ver cómo desbloquear abre ese Sheet en Resumen y enfoca el título Desbloqueo. Documentos adjuntos abren su cadena en el mismo panel con Volver a la tarea; conserva revisión de tarea y no anida overlays. Imágenes usan el visor existente y retornan al mismo contexto. Al abrir, capturar task_version, estado, motivo, bloqueo y referencias de evidencia de esa versión. Solo una task_version posterior dispara el aviso; cambios ajenos del trabajo no lo hacen. Actualización viva de tarea ofrece Ver la última, sin sustituir el contenido que la persona lee. Historial muestra transiciones sin duplicar capturas.

```
Antes
Sin detalle por tarea

Después
T5 — Integración con proveedor                    X
! Bloqueada
Falta acceso al entorno de pruebas.
[Resumen] [Evidencia] [Historial]
Desbloqueo
1. Solicitar acceso al responsable del entorno.
2. Configurar la credencial en su ubicación segura.
3. Ejecutar prueba de conexión.
Responsable: equipo del entorno
Reanudar cuando: prueba de conexión correcta.
```

Estados regionales de tareas: Skeleton de filas al cargar; lista vacía “Sin tareas publicadas” con Actualizar; error inline Reintentar. Sheet sin evidencia explica “Sin adjuntos” con Actualizar, no exige imágenes. Historial vacío muestra registro inicial; error por tarea/adjunto se mantiene local con retry y no convierte task a blocked. Enlaces inaccesibles no filtran información ajena. Sin mutaciones de estado/aprobación en navegador. Spinner sustituye aria-label Loading por el estado localizado y evita anuncios repetitivos por frame; el estado estable queda en el nombre accesible de fila. No nueva animación salvo Spinner instalado; respetar reduced motion.

**Filtrado por etapa y progreso global — D74.** No mostrar tareas Verify/Delivery dentro de Build. La fila indica estado del tramo; Sheet presenta el global y cada etapa. Si un plan asigna una tarea a dos etapas, cada una muestra solo su proyección, con un mismo ID. El total global deduplica por task_id, no suma totales de etapas. Future tasks solo en plan hasta abrir su etapa real.

```
Antes · tarea pegada a Build
Build                          6/7 finalizadas
  ✓ T1 … T6
  ○ T7 — Verificación y entrega

Después · asignación explícita
Trabajo                        6/7 finalizadas
Build                          6/6 finalizadas
  ✓ T1 … T6
Verificación                   1/1 finalizadas
  ✓ T7 — Verificación y entrega
Entrega                        0/1 finalizadas
  ◌ T7 — Verificación y entrega                  v
    Ⅱ CI — Checks del pull request              >
```

Este boceto representa Verificación ya terminada y Entrega real iniciada. Antes de esas etapas, no se fabrican nodos: T7 permanece consultable en el plan y cuenta en el global. Al completar el tramo de Entrega y el alcance pendiente de cierre de T7, el trabajo pasa a 7/7 tareas, sin afirmar integración antes del merge observado.

Nodo CI usa el mismo Timeline/Tooltip/Collapsible y abre Sheet común con título “CI — Checks del pull request”. Cuando hay task/segment se muestra una sola vez dentro de ese tramo; sin task sigue a nivel de Entrega. Expandido resume cuántos checks faltan y última observación. Sheet: Resumen/Checks/Intentos; PR/candidato fijos arriba, links reales, estado observado por check y motivo/acción al fallar. Un cierre, ancho común, teclado/foco y snapshot manual. Sin botones que hagan rerun/merge ni acciones remotas nuevas. No confundir el control Reintentar lectura con reejecutar CI.

```
Antes
Sin etapa Entrega ni espera CI

Después
Entrega
  Ⅱ CI — Checks del pull request                v
    Faltan 2 checks del candidato actual.
    [Ver checks]

Sheet: CI — Checks del pull request              X
Candidato: commit …   PR: enlace real
[Resumen] [Checks] [Intentos]
  ✓ Contratos
  ◌ Pruebas
  ○ Build
Última observación: …
```

Cargando checks: Skeleton; consulta fallida: error inline Reintentar lectura; sin autoridad/lista requerida confirmada: “Sin confirmar” y explicación/Actualizar. Snapshot nuevo avisa sin cambiar lo abierto. El punto de tarea sigue siendo solo estado de su tramo; el punto del nodo CI muestra el estado de esa espera. Ambos usan nombres accesibles distintos, sin badges repetidos en fila.

## Copy

| es-MX | en |
| --- | --- |
| No iniciada | Not started |
| Bloqueada | Blocked |
| Desbloqueo | Unblocking steps |
| Ver cómo desbloquear | See how to unblock |
| {x}/{y} finalizadas | {x}/{y} completed |
| Resumen / Evidencia / Historial | Summary / Evidence / History |
| Sin tareas publicadas | No tasks published |
| Sin adjuntos | No attachments |
| Entrega | Delivery |
| En espera | Waiting |
| Sin confirmar | Unconfirmed |
| Ver checks | View checks |
| Checks / Intentos | Checks / Attempts |
| En curso | In progress |
| Revisión | Review |
| Validación | Validation |
| Pausado | Paused |
| Finalizado | Finished |
| Fallido | Failed |
| Sesiones: {n} | Sessions: {n} |
| Conductor | Lead |
| Reasignar | Reassign |
| Reasignar conducción | Reassign lead |
| Motivo | Reason |
| La conducción cambió; actualiza antes de continuar | The lead changed; refresh before continuing |
| Ir al actual | Go to current |
| Espera tu respuesta | Awaiting your response |
| Publicado | Published |
| Revisión inicial | Initial revision |
| Hay una revisión nueva | A newer revision is available |
| Ver la última | Show latest |
| La pestaña anterior ya no existe | The previous tab no longer exists |
| Productor no disponible | Producer unavailable |

Conservar Ver investigación/plan/implementación/verificación, Cambió/Nuevo/Quitado, Ver más/menos y plurales.

## Impact

Go/TS contracts, store/server, CLI, frontend, fixtures, documentación y guía canónica flow-walkthrough. Revisar dependencias en global, flows plan/build/verify/close, roles, site, arquitectura y detectores tests/pilot sin ejecutar pilotos. La espera ya es regla: corregir aplicación, no agregar ceremonia. Colaboración/estados son requisitos adoptados D69/D70, no corrección inferida de un único caso. No cambios a CAM ni producción; datos sintéticos de volumen para UI; no importar registros viejos de CAM.

## Recovery

El candidato v3 se verifica con binario/puerto/estado nuevos. La ruta anterior no se usa, transforma ni elimina. Para recuperar el servicio antes de publicar datos v3, restablecer binario/CLI/content y directorio anteriores bajo autorización; no trasladar v3 al almacén viejo. Una vez existan publicaciones v3 únicas, preservar ese directorio y preferir fix-forward: no hay conversión ni rollback de datos prometido. El corte del runtime se prepara como paso separado y concreto, con writers coordinados; no inferir reinicio/cambio de state-dir de la excepción de mantenedor que exige conservar ajustes. No es requisito del build intervenir el servicio habitual.

## Runtime isolation constraint

La pregunta/restricción del usuario se aplica al build y a cualquier integración posterior: compilar con -o a ruta propia, frontend servido desde candidato, puerto libre y estado v3 nuevo. No reemplazar command -v hive, no hive update, no reiniciar ni cambiar la configuración/estado del visor habitual mientras CAM lo usa. Las pruebas no ejecutan procesos con la ruta del binario habitual como destino. El refresh implícito de mantenedor queda suspendido para este trabajo bajo esta restricción; la entrega puede integrar código, pero instalar/cortar requiere instrucción separada sobre blanco concreto.

### Correcciones de revisión visual — 2026-11-01

La revisión del usuario solicita indicadores sin borde exterior duplicado, color semántico según estado conservando forma/tooltip, y señales compactas de evidencia/documentos/CI/historial visibles antes de desplegar una tarea. Sustituye la apertura incondicional de Sheet de D73: una tarea con solo descripción y metadatos de asignación/etapa usa lectura en línea; el Sheet queda para evidencia, documentos, esperas, bloqueo o historial de transiciones significativo. No se eliminan anillos de foco de teclado. El demo integral conserva las imágenes iniciales visibles y la cabecera de icono + contador. D78 sigue pendiente; no implica aprobación del candidato.

### Toolbar de tareas e inicio — revisión del usuario

```text
Antes
4/9 finalizadas
[Ir al actual]

Después
[Tareas ▾] [Ir al actual]
  Popover: 4/9 finalizadas
  ✓ T1 · descripción corta              1/1
  ! T2 · descripción corta              0/1

○ Inicio · fecha/origen del trabajo
│ Actividad e imágenes publicadas
│ Investigación / Plan / Build / Verify / Entrega
```

El Popover usa las entidades únicas de session.tasks; no crea plan.tasks ni build.tasks. La ampliación de referencias canónicas tiene un solo contrato en [canonical-references.md](canonical-references.md); no se introduce un segundo listado vivo dentro del Sheet documental.

### Canonical-reference implementation boundary

[Canonical references](canonical-references.md) is the owning contract for T9–T12. It supersedes inline entity payload assumptions above: task ownership/segment state remains unchanged; tickets and deliveries gain stable-ID current registries, candidates immutable IDs, and server-generated publication context. No extra inline Plan catalog or provider adapter is added. The public format changes to3; old1/provisional2 state is rejected before writes. Existing T1–T6 verification remains valid evidence for its recorded baseline, not verification of the new schema.

### Browser language and reading-space correction
The user requested the localization/schema extension during build, authorizing it now, and clarified that keys primarily prevent repetition; the implementation chooses existing stage identity rather than a duplicate title_key. The owning contract is the Reserved labels and browser language section of canonical-references.md. Existing stage values select standard headings; optional producer title adds context without repeating that heading. Other nodes retain their free titles. No language setting is stored on the work.
Before: [Tareas popup] rows near scroll boundary; final footer flush to canvas edge.
After:  [Tasks/Tareas popup] bounded content with comfortable gaps; final footer followed by visible empty canvas space.
Keep the operational Record-detail/flat Frame-led composition, dense native controls, existing skeleton/empty/retry states and no added motion. No new dependency or global configuration.
