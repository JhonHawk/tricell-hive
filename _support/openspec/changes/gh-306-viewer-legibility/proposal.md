# #306 — Legibilidad y continuidad del visor

| Field | Value |
| --- | --- |
| Status | Paused · verified local candidate committed on current branch; user requested no publication |
| Tracker | [#306](https://github.com/JhonHawk/tricell-hive/issues/306) · OPEN, read 2026-11-01 |
| Git | Base development 044664a0; checkout refactor/flow-ceremony-2 at 2f0a133b |
| Decisions | D62 B server association; D63 A Sheet; D64 A issue; D65 A interactive; D66 A Claude Code /code-review; D67 C existing branch; D68 A common documents; D69 A shared work; D70 A states with reasons; D71 A new protocol; D72 A local viewer recovery; D73 A expandable task nodes and task Sheet; D74 A work-owned tasks with stage-filtered views; user correction: clean cutoff, no legacy compatibility; final canonical-reference protocol3 supersedes provisional2 |

## Request

Issue leído el 2026-11-01: “Events of a stage (decisions, first views, captures) render as separate top-level nodes above their stage, so the stage in progress sinks down the list.” “The 40rem dialog truncates its tabs and is narrow for long documents with tables.” “Revisions show no indication of what changed between them.”

El issue también pide contadores con texto, resúmenes legibles, pies distintivos y avisos precisos de revisión. La dirección newest-first original queda supersedida por la petición ascendente del usuario; el texto visible de los contadores fue después reemplazado por icono + contador a petición del usuario.

Ampliación solicitada en Codex el 2026-11-01:
> “puede haberlo para contenido dinamico, no solo para el plan”
> “son dos sesiones en paralelo, y el registro en el visor es con el id de la otra sesión”
> “Faltan estados tambien [...] las pausas con el usuario son relevantes”

D68 A limita el contenido a Markdown versionado de cualquier etapa; D69 A comparte el trabajo con productor real; D70 A añade estados con motivo; D71 A usa protocolo nuevo; la instrucción posterior de cutoff supersede su lectura compatible v1. Correcciones:
> “Los estados con formato corto / Validación | En curso | Finalizado”
> “ascendente, lo mas antiguo arriba y va creciendo hacia abajo”
> “podemos hacer un un cutoff, no es necesario preservar el legacy, solo tenemos unas cuantas sesiones con el binario anterior”

> “que se vayan viendo el progreso de las task, cada una con su estado independiente, pero solo en el punto”
> “Bloqueada [...] el sheet debe contener los pasos para desbloquear”

D73 A acepta flecha para expansión breve y título para abrir el Sheet de tarea.

D74 A adopta tarea perteneciente al trabajo, con identidad conservada a través de etapas. El usuario precisa:
> “en build solo se muestran las tareas de esa etapa, y las de verify u otras etapas las que correspon”
> “la espera del CI entiendo que pertenece a entrega, que tenga su nodo de espera tambien”
> “Puedes ir contruyendo y probando sin alterar el binario que esta usando la sesión de cam verdad?”

La última restricción se conserva: build y pruebas aislados; ningún reemplazo del binario instalado, hive update, reinicio habitual ni corte final v3 mientras CAM dependa de ese runtime. Esta consulta no es una instrucción de flow-build.

## Objective

Leer el trabajo en orden, ver pausas y revisiones y continuarlo entre sesiones sin atribuir publicaciones al ID de otra. Un panel común muestra documentos de cualquier etapa y conserva la revisión abierta. Dentro de cada etapa, tareas compactas muestran progreso independiente en el punto, evidencia pertinente y detalle/desbloqueo en Sheet.

## Scope and acceptance

Dentro: corte limpio a contrato/storage v3 sin lectores ni migración v1/v2, etapa Entrega y espera de CI, productores incorporados explícitamente, conductor único, historial de transiciones, agrupación ascendente, Sheet común, tareas independientes expandibles, revisiones, legibilidad y guía de publicación. Fuera: contenido ejecutable/interactivo, autenticar procesos por productor, importar/fusionar sesiones históricas, CAM, pilotos de modelos, release y producción.

AC1–AC10 conservan sus IDs con los textos actuales que reemplazan la versión anterior; AC11–AC27 amplían el alcance.

- AC1 — Los eventos nuevos asociados a una etapa devuelven stage_event_id calculado por servidor; un target_stage_event_id de entrada permite colaboración explícita. Destinos inexistentes, retirados o ajenos fallan; el cliente no puede asignar el campo de salida. Etapas, cierre y actividad global no lo llevan. · Falso en development: no existen esos campos.
- AC2 — El visor v3 inicia con un almacén nuevo vacío, sin importar sesiones API v1/v2 ni manifiestos legacy. Al apuntar a un estado anterior devuelve cutoff_required antes de escribir y no lo convierte, borra ni mezcla. · Falso en development: inicia/reconcilia los registros actuales y el scanner ingiere manifiestos.
- AC3 — La línea de tiempo muestra primero eventos previos a etapas, luego etapas por inicio ascendente, tareas en orden del plan y actividad interna ascendente y cierre al final. Actual/Revisión/Validación/Pausado quedan abiertas; finalizadas/fallidas pueden plegarse. Las retiradas conservan su contenido identificado. · Falso: lista plana invertida.
- AC4 — Documentos de investigación, plan, build, verificación y entrega usan el mismo Sheet: dos tercios en escritorio, ancho completo móvil, pestañas fijas desplazables y un cierre. Solo planes agrupan deltas bajo Especificaciones. · Falso: Dialog de 40rem.
- AC5 — Cada revisión muestra pestañas nuevas/cambiadas/quitadas frente a su predecesora de la misma cadena por label/document_id; la primera dice Revisión inicial. · Falso: no hay comparación.
- AC6 — Solo una revisión posterior de esa cadena dispara el aviso; Ver la última cambia sin cerrar y conserva la pestaña si existe, con fallback anunciado si desapareció. El documento no cambia solo ni por eventos ajenos. · Falso: se compara session.version.
- AC7 — Encabezado compacto con icono + contador para tickets, entregas y sesiones, con nombres accesibles y ayudas completas; validaciones con nombre y ocultas a cero. Supersede por petición posterior del usuario los contadores con texto visible del issue. · Falso: etiquetas largas ocupan la cabecera o se pierde el nombre accesible.
- AC8 — Resumen de trabajo a dos líneas y de etapa/evento a tres con Ver más; listas Markdown como listas. · Falso: párrafos completos.
- AC9 — Captions muestran texto distintivo no vacío y conservan nombre accesible completo. · Falso: prefijos repetidos.
- AC10 — La guía pide títulos del trabajo y resúmenes simples; distingue trabajo y productor real, incorporación y handoff; publica espera antes de la parada al usuario y registra respuesta/reanudación sin inferir aprobación ni añadir confirmaciones. · Falso: no define colaboración/motivos/títulos simples; la espera genérica ya existe y se conserva.
- AC11 — Dos tuplas nativas incorporadas explícitamente publican en un trabajo sin reasignar source/IDs; cada publicación nativa nueva conserva productor y contenido acotado al trabajo. Productor ajeno, pendiente o de otro proyecto falla sin cambios. · Falso: una sola tupla por sesión.
- AC12 — Solo el conductor vigente conduce etapas, estado global y cierre, salvo recuperación local explícita D72; traspaso condicional a miembro aceptado conserva contenido. Transferencias simultáneas/versiones obsoletas dan un ganador y conflicto sin doble efecto. · Falso: no hay conductor/traspaso.
- AC13 — Etapa/cabecera usan En curso, Revisión, Validación, Pausado, Finalizado y Fallido con motivo separado; revisión paralela no interrumpe build y espera al usuario muestra Validación con destino cuando existe. · Falso: cuatro estados de etapa, sin motivo.
- AC14 — Cada transición distinta, incorporación aceptada y traspaso deja auditoría inmutable con motivo, productor, tiempo del servidor y estados anterior/nuevo cuando aplica, reconstruible tras reinicio. · Falso: patch solo cambia estado y updated_at.
- AC15 — CLI, visor y contrato offline usan solo protocolo/storage v3; escrituras v1/v2 y pares incompatibles fallan antes de mutar con diagnóstico de actualización. Las operaciones y recibos v3 sobreviven reinicio; no existe migración ni lector de registros v1/v2 en la ruta v3. · Falso: protocolo/storage 1 y rutas legacy activas.
- AC16 — Contenido nuevo conserva posición de lectura; Ir al actual salta al foreground o último evento visible; deep links y scroll por trabajo se conservan. · Falso: no existe acción ni política para grupos ascendentes.
- AC17 — document_key estable distingue cadenas independientes; revisión posterior puede continuar en etapa enlazada del mismo tipo sin mezclar documentos. Claves inválidas, destinos ajenos o predecesor incorrecto fallan. · Falso: no hay clave ni predecesor explícito.

- AC18 — Renombrar, archivar, fijar y eliminar desde el navegador siguen funcionando con el actor local del visor, sin suplantar al productor conductor; la recuperación explícita de D72 tiene su operación separada; sus cambios quedan auditados y la eliminación exige el control de confirmación existente. · Falso para el contrato propuesto sobre development: el contrato actual no distingue actor local/nativo ni audita estas acciones; preservación y nueva procedencia se verifican juntas.

- AC19 — Si el conductor no responde, una acción local explícita permite aceptar una solicitud del mismo proyecto y reasignar conducción atómicamente, con auditoría y versión condicional. No cambia source, no infiere abandono, no detiene procesos; el conductor anterior conserva su membresía sin autoridad conductora. Dos recuperaciones simultáneas o una versión obsoleta no sobrescriben al ganador. · Falso: no hay solicitudes, recuperación ni autoridad por productor.

- AC20 — Las tareas se declaran para todo el trabajo con ID estable, clave T<n>, descripción corta y asignaciones explícitas a etapas. Cada etapa real muestra solo sus tramos asignados en orden del plan, incluidas no iniciadas; la fila muestra estado local solo en el punto (círculo/loader/pausa/alerta/check), con tooltip y nombre accesible. · Falso en development: no hay registro/asignaciones ni nodos de tarea.
- AC21 — La flecha expande resumen y evidencia sin cerrar otras tareas; las tareas con contenido significativo abren su Sheet desde el título y muestran accesos de contenido antes de desplegar. Las que solo tienen descripción/metadatos se leen en línea sin Sheet vacío. Polling, cambios de estado y volver del Sheet preservan expansión, foco y scroll. Etapa actual abierta y tareas inicialmente contraídas; una etapa con tareas en curso/revisión/bloqueadas permanece visible aunque ya no sea foreground. · Falso: no existe interacción por tarea.
- AC22 — Pasar una tarea a Bloqueada exige motivo, pasos de desbloqueo, responsable y criterio de reanudación, todos legibles en el Sheet sin secretos; datos incompletos fallan sin cambiar estado. Bloquear un tramo no bloquea automáticamente otras tareas, tramos ni el trabajo. · Falso: no hay task state ni payload de desbloqueo.
- AC23 — Adjuntos y resultados opcionales se vinculan al task_id y a evidencia/documentos del mismo trabajo con etapa real de origen y tramo asignado explícitos; la expansión puede mostrar capturas pertinentes y el Sheet conserva detalle e historial; nuevo contenido vinculado incrementa atómicamente task_version y ofrece actualizar el snapshot sin cambiarlo solo, mientras una tarea ajena no dispara aviso. Tareas sin imágenes funcionan igual, y terminar/corregir conserva evidencia anterior sin mezclarla con otra tarea. · Falso: solo se asocia contenido a sesión/evento, no a tarea.
- AC24 — Transiciones condicionales de tarea conservan actor e historial, sobreviven reinicio/retry y actualizan el mismo nodo sin duplicarlo; corregir un tramo finalizado lo reabre explícitamente con motivo. Los conteos de etapa usan tramos asignados; el global cuenta tareas únicas finalizadas en todos sus tramos requeridos, sin inferir porcentaje, aprobación o estado por índice visual. · Falso: no hay estado/historial/contador de tareas.

- AC25 — Una tarea que atraviesa etapas conserva un task_id y tiene estado independiente por tramo: una proyección en cada etapa explícitamente asignada, sin replicar la entidad ni sus adjuntos. El caso CAM muestra Build 6/6, Verificación T7 finalizada en ese tramo y Entrega T7 pendiente/en curso, mientras el total sigue 6/7 hasta terminar T7 completa. Tareas de etapas futuras no aparecen dentro de Build ni crean etapas ficticias. · Falso: no hay tarea global/tramos/conteos.
- AC26 — Entrega es una etapa propia posterior a Verificación; muestra un nodo de espera de CI con punto de pausa y Sheet que identifica candidato/PR, checks requeridos, resultados observados y enlaces reales. No afirma CI correcto sin confirmar el conjunto requerido para el commit actual ni equivale a integración. · Falso: no existe etapa delivery ni nodo de espera.
- AC27 — El nodo CI conserva identidad durante la misma espera, sus intentos/resultados y tiempos observados; solo el intento aplicable de cada check determina el estado actual, nunca una observación atrasada de un intento anterior; fallo/cancelación/desconocido no se presentan como éxito ni habilitan merge. Reejecuciones conservan historia y un candidato nuevo no reutiliza el resultado anterior; retries/versiones/reinicio no duplican nodos ni pierden resultados. · Falso: no hay contrato/almacén de espera CI.

- AC28 — Un catálogo único con siete tareas muestra siete entradas en el Popover, seis referencias en Build y el mismo T7 en Verify/Delivery, ordenado por plan_order sin duplicar entidades. · Falso: el ejemplo agrega T8/T9 y su Markdown declara otra lista.
- AC29 — Un patch al ID de un ticket/entrega cambia sus valores actuales en todas las proyecciones tras una lectura, sin republicar cada evento o vista. · Falso: tickets usan identidad key+URL y Git conserva copias embebidas.
- AC30 — Payloads de proyección usan solo IDs del trabajo, no redeclaran tickets, entregas, candidatos ni snapshots de servidor; IDs ajenos, mezclados, duplicados o campos retirados fallan antes de escribir. · Falso: el contrato admite entidades inline.
- AC31 — Actualizar una entidad actual conserva contexto publicado, bytes/hash de documentos/imágenes y observaciones históricas; el lector abierto mantiene snapshot y ofrece actualización manual. · Falso: no hay escritor de entidad compartida que permita comprobarlo.
- AC32 — Un cambio visible compartido invalida una vez cada tarea/espera afectada; cambios ajenos, no-ops y retries no la invalidan. · Falso: no existe invalidación de entidad compartida.
- AC33 — Mutaciones de catálogos con CAS concurrente tienen un ganador; IDs, referencias, recibos y versiones sobreviven reinicio/fallo sin commit parcial ni perder reserva de cierre. · Falso: no hay catálogos/operaciones correspondientes.
- AC34 — Inicio es un nodo claro derivado de created_at/Source reales, sin evento API ficticio ni cambio de conteos. · Falso: el primer nodo es actividad sin marcador de origen.
- AC35 — Tareas/Ir al actual forman una toolbar horizontal; el total y catálogo canónico viven en un Popover compacto, accesible, con scroll acotado y una fila por tarea. · Falso: el total se muestra como columna separada sin catálogo vivo.

## Prerequisites

None: the build needs nothing from outside the repository. Preparar un directorio v3 nuevo y comprobar el procedimiento de corte local es tarea del build, no acceso externo. El corte real del servicio habitual requiere concretar sus rutas, escritores y aprobación aplicable; no es necesario para construir/verificar el candidato aislado.

## Verified context

2026-11-01: checkout 2f0a133b, development 044664a0; tooling/viewer sin diferencias entre ambos. Leídos contratos/store/server/CLI, timeline-content/documents/copy, spec walkthrough-viewer y componentes instalados Sheet/Badge/Tabs, patrones Popover/Select. Visor habitual pid 49211 puerto 4483, revisión 0b6f2380 modified=true: distinto al checkout.

CAM 01a195f8 tenía research/plan completos y build activo. Claude cfb6fcae publicó cuatro documentos idénticos a los archivos previos, pero etapa in_progress pese al plan listo. No se demostró escritura cruzada: colaboración es requisito adoptado, no causa afirmada. Esta ampliación tiene registro propio b31295ed… con identidad real Codex 01a19638-64aa-7e73-af2a-7559b46529c4. Antes de D68–D70 se verificó awaiting_validation y etapa paused.

## Delivery

D65 A: interactiva, validación antes del push y después PR/merge a development. D66 A: Claude Code /code-review; si el host de build no puede ejecutarlo, no sustituirlo sin decisión. D67 C: rama actual; enumerar y consultar commits ajenos antes de publicar. No autoriza build en este flujo, release ni cambios en otros proyectos. La restricción actual del usuario prevalece sobre el refresh de mantenedor: no reemplazar el binario instalado ni actualizar guía/globales ni reiniciar el visor habitual mientras CAM lo use. Integrar no autoriza ese corte; solicitar la acción concreta separada cuando corresponda.

## Review status and progress

2026-11-01 — La revisión anterior de Claude no valida esta ampliación. Se conservan decisiones previas salvo orden descendente y alcance exclusivo de plan, supersedidos por el usuario. flow-plan paso 5: hive-review-plan completó una revisión independiente. H1 P1 (actor de acciones del navegador) aplicado con AC18. H2 P1 (conductor no disponible) resuelto con D72 A, recuperación local explícita y AC19. Una revisión independiente completa; correcciones reconciliadas por el principal. Tras la revisión, el usuario pidió corte limpio: se retiran adaptación v1 y recuperación de formato anterior. Reconciliación acotada del mismo revisor: H3 P1 por contratos legacy aún vigentes, corregido ahora con delta explícito de los requisitos afectados; ninguna tarea empezada.

2026-11-01 — D73 A amplía el mismo plan con tareas expandibles y Sheet. AC20–AC24 añadidos; AC3 precisa orden de tareas del plan. Revisión acotada hive-review-plan completada sin hallazgos materiales; build sigue sin autorización. AC1–AC24 y T1–T8 listos.

2026-11-01 — D74 A y corrección por etapa incorporadas; espera de CI en nueva etapa Entrega añadida. AC20/AC22–AC24 ajustados y AC25–AC27 añadidos. Revisión acotada completada: H4 P1 invalidación task_version por contenido enlazado y H5 P1 intento CI aplicable corregidos en contrato/tareas/pruebas. AC1–AC27 y T1–T8 listos. Ninguna implementación ni cambio al binario/servicio habitual.

2026-11-01 — Usuario autoriza flow-build. Continúa en checkout existente por instrucción posterior explícita, sin worktree; no habrá cambios de otras sesiones hasta terminar. HEAD de arranque 5e956577, con commit ajeno de viewport móvil preservado.

2026-11-01 — Usuario pide Inicio, toolbar/popover y fuente única para todas las referencias, autorizando ajustar schema ahora. Se añaden AC28–AC35 y T9–T12; contrato propietario canonical-references.md, protocolo/schema/storage final3 con corte limpio1/2. T1–T6 completados verifican el baselinev2; sus recorridos afectados se repiten sobre3. D65/D67 y restricción CAM conservadas; D77mecanismoreview siguependiente. No publicación, instalación ni cambio habitual.

## Additional acceptance criteria — browser language and reading space
- AC36: English is the UI default/fallback; es/es-* browser languages select Spanish consistently in production and preview, including system headings, controls, accessible names, status/date text and document language. Producer content is unchanged.
- AC37: The offline schema/catalog exposes reserved-label metadata; the existing stage enum localizes standard stage headings once, optional title context never redeclares their names, and other nodes retain free titles. Empty stage context is allowed while empty non-stage titles fail unchanged, with Go/TS parity and restart/replay.
- AC38: Seven task rows do not produce a residual scroll when space permits; constrained viewports retain keyboard scrolling. The true timeline end has visible bottom space after its last footer without clipping or fake content.
