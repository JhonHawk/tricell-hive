# Diseño · sesión y recorridos del visor

## Contexto e impacto

Base inspeccionada: `development` en `82ccf111`, 2026-10-26. Las capturas de la conversación son la referencia visual aceptada; no se inspeccionó el runtime con un navegador en esta etapa.

La sesión lógica ya es la tupla exacta `[project_key, host, session_id]` (`frontend/src/sessions.ts:13`); su título actual es el del primer recorrido. La selección y los recursos usan claves físicas, que se mantienen. `store.go` persiste archivo/conservación en `meta.json`; `scan.go` los hereda al reemplazar esa copia. La biblioteca local inspeccionada tiene 20 copias y 16 sesiones, 17 copias archivadas, cero conservadas y un grupo de archivo mixto. El conteo se hizo sin bloquear el servidor: repetirlo antes de una transición real.

Se buscaron rutas literales `/api/sessions/`, llamadas `Archive`, `Pin`, `Delete` y lecturas de índice en `tooling`, `content`, `site` y `_support/docs`.

**Impacto:**

- `frontend/src/controller.ts`, `application.tsx`, `sessions.ts`, `shell.tsx`, `api.ts`: cambiar acciones de copia por acciones de sesión y consumir el nombre local; mantener selección física y lectura de documentos.
- `server.go`, `store.go`, `scan.go`, `index.go`: resolver miembros en servidor; persistir estado lógico, importarlo, aplicar retención y borrar el conjunto. Hay estado concurrente protegido por `Store.mu`: requieren pruebas con race detector.
- `tooling/cli/view_command.go`, `viewIndex` y `viewHasSource`: conservar estructura física y fingerprint de confirmación. Su decoder permite campos añadidos; verificarlo con una prueba de índice extendido, sin atribuirle un esquema nuevo de grupos.
- `store.go`, `readMeta`: usa `DisallowUnknownFields`; no introducir el estado de sesión como campos nuevos de `meta.json`. La nueva representación tiene su archivo propio.
- Tests Go y frontend de archivo, eliminación, navegación, exclusión de fuentes y conservación: sustituir las expectativas obsoletas por acciones del conjunto, conservando autenticación, recursos y aislamiento de documentos.
- `tooling/viewer/doc.go`, `frontend/README.md`, `_support/docs/architecture/deployment-manager.md`, `README.md`, `llms.txt`, `site/src/content/docs/flows/build.md`, `CHANGELOG.md`: actualizar lo que describa el comportamiento afectado. No editar skills ni agentes: no cambia el manifiesto generado ni el procedimiento de publicación.

Señales encontradas: varios lectores internos; datos persistidos y concurrencia. No se encontraron otros clientes de producción dentro del alcance buscado. Clientes particulares fuera del repositorio no verificados: los cambios de alcance de las rutas de escritura son incompatibles y requieren nota explícita de migración, no una afirmación de ausencia de consumidores. La transición es coordinada entre el binario y su frontend embebido; no se mantiene un segundo juego de acciones por copia como compatibilidad especulativa.

## Estado lógico y compatibilidad

- Añadir `viewer/session-groups/<digest>.json`, con versión de formato, tupla de identidad exacta, `title_override` opcional, `archived`, `pinned` y marcador de recuperación `delete_pending` opcional. El digest se deriva de una serialización canónica de la tupla; verificar equivalencia con la agrupación TypeScript, conservando `host` ausente como nulo. El digest no reemplaza las claves de recursos.
- El archivo lógico es la autoridad después de importar. El índice publica los flags efectivos en sus copias, y añade `session_title` opcional, sin cambiar la forma de `projects/tickets/sessions`. La etiqueta visible usa el nombre propio o, en su ausencia, el título cronológico actual, acompañado del resumen de tickets; no inventar un resumen con un modelo ni consultar el host.
- La edición local cambia solo `title_override`. Campo de una línea, trim, máximo 200 puntos de código Unicode; vacío elimina el override y recupera el derivado. Rechazar caracteres de control; renderizar siempre como texto. No se usa el nombre como ruta, identidad ni selector.
- Archivo y conservación se aplican a todos los miembros de la tupla, incluyendo copias ocultas o apartadas por clasificación de fuente. No habilitan lecturas de documentos apartados. Nuevas ingestas consultan la misma autoridad antes de publicar y purgar; así heredan ambos flags. El archivo no cambia TTL, fecha de ingesta ni estado del trabajo.
- Escribir un único archivo lógico de forma atómica bajo el lock existente y actualizar la vista desde esa autoridad. Cambios de nombre/archivo/conservación modifican únicamente sus campos para no perder cambios concurrentes. Fallos de persistencia no deben producir una confirmación de éxito; si el archivo ya se escribió pero el índice no, el próximo acceso/reinicio debe recuperar el estado autorizado por ese archivo, no sobreescribirlo con un índice antiguo.
- La pérdida de `index.json` no pierde nombre ni flags: reconstruir copias desde sus metadatos y estado lógico desde los archivos de grupos. La corrupción de un archivo de grupo produce un error visible y bloquea operaciones/retención destructivas para sus miembros hasta recuperarlo; no tratarlo como un grupo nuevo con flags falsos.
- Cuando desaparece el último miembro por TTL, conservar el estado lógico mientras una fuente registrada siga identificando la sesión; si no es posible demostrar que no reaparecerá, conservar ese pequeño registro. No añadir un recolector de estos registros en este cambio. Eliminar explícitamente la sesión sí retira su nombre y flags tras borrar todos los miembros, para que una entrega posterior pueda aparecer como sesión nueva.

## Importación y recuperación

La migración afecta flags de copias locales, no documentos fuente. El volumen conocido es 20 metadatos; no hay servicios remotos ni base compartida que migrar. Otras instalaciones tienen volumen desconocido y ejecutan la misma importación acotada por biblioteca al abrirla, sin usar el conteo mantenedor como límite.

1. Abrir la biblioteca con un solo escritor y agrupar sus copias válidas. Antes de la primera conversión real, respaldar binario y archivos de metadatos/índice/tombstones y estado grupal previo, en ubicación local no versionada, directorio `0700` y archivos de datos `0600`. No copiar `server.json`, tokens ni credenciales: no son necesarios para revertir este formato. No imprimir contenidos del respaldo.
2. Para cada grupo aún no importado: `archived = all(legacy.archived)`, `pinned = any(legacy.pinned)`, sin override. Persistir primero el archivo de grupo y después retirar los flags legacy de sus metadatos físicos. Aceptar los campos legacy solamente como entrada de importación. El escritor nuevo no los vuelve a generar.
3. La presencia de un archivo grupal válido impide recalcularlo a partir de campos legacy después de una interrupción; terminar la retirada pendiente sin modificar decisiones ya almacenadas. No hace falta un framework de migraciones ni copias paralelas operativas.
4. Reconciliar conteos y hashes de recursos antes/después, identidad, archivo y conservación. El grupo mixto queda activo por decisión del usuario; todos los grupos completamente archivados siguen archivados. Reabrir, reconstruir el índice y repetir la importación no cambia el resultado.
5. Ante fallo, detener la transición y reportar el grupo pendiente, sin borrar contenido ni purgar miembros cuya preferencia no se pudo determinar. Probar interrupciones entre cada escritura relevante con fixtures, no contra la biblioteca cotidiana.

Rollback: el binario antiguo puede no interpretar el nuevo estado lógico y los flags retirados; volver al binario anterior exige restaurar el respaldo coherente de sus metadatos, sin sobreescribir ediciones posteriores. Si hubo acciones después del respaldo, capturar ese estado y preparar una conversión inversa de flags antes de retroceder; nombres locales no son consumidos por el binario anterior y se conservan en el respaldo. La ejecución de rollback no forma parte de la planificación.

## API y eliminación

Reutilizar `/api/sessions/<key>` como localizador de un miembro existente, resolviendo la tupla y todos los miembros dentro del servidor. No enviar una lista de claves derivada de la vista filtrada.

- `POST`/`DELETE /api/sessions/<key>/archive`: fijar archivo verdadero/falso para la sesión, 204 al éxito.
- `POST`/`DELETE /api/sessions/<key>/pin`: fijar conservación verdadera/falsa para la sesión, 204 al éxito.
- `PUT /api/sessions/<key>/title`, JSON `{ "title": "..." }`: fijar o limpiar el nombre local, 204 al éxito. Cuerpo limitado a 2048 bytes, rechazo de campos desconocidos, tipo erróneo o contenido JSON posterior al objeto.
- `DELETE /api/sessions/<key>`: eliminar todos los miembros actuales de la sesión y registrar los tombstones por copia de la implementación existente; nunca tocar fuentes. No bloquear por identidad; contenido nuevo/modificado puede reingerirse.
- Las lecturas de manifiestos y archivos siguen por clave física y no cambian su autorización.

Todas las escrituras conservan token y Origin/Host de mismo origen; no cambian CORS ni la política de capacidades. Además, las cuatro operaciones de sesión exigen `X-Hive-Action-Scope: session`, enviado solo después de que el frontend muestre el alcance del conjunto. Las llamadas de una pestaña antigua sin esa marca reciben 409 con código `reload_required`, sin mutar nada; la marca es protección de compatibilidad, no autenticación. El frontend muestra «Actualiza el visor para continuar» y una acción de recarga, sin reintentar automáticamente una escritura. GET y confirmación CLI no requieren la marca. Actualizar tests y nota de migración; no mantener una implementación operativa de acciones por copia.

Respuestas de validación 400, autorización 403, miembro inexistente 404, incompatibilidad de alcance 409, fallo de almacenamiento 500, método no soportado 405. Tras éxito el cliente recarga el índice antes de decidir selección y controles. Repetir archive/pin/title con el mismo valor no cambia el resultado.

No prometer borrado transaccional entre carpetas. El orden recuperable de borrado, bajo el lock y antes de permitir otra ingesta, es:

1. Capturar los miembros actuales y persistir en `index.json` los tombstones de todos sus fingerprints antes de retirar una carpeta. Si esa escritura falla, abortar sin eliminar ninguna copia.
2. Persistir `delete_pending: true` en el registro lógico antes de retirar carpetas. Si falla, abortar sin borrarlas; los tombstones no ocultan las copias aún existentes en el índice.
3. Intentar retirar las carpetas capturadas; reconstruir los miembros realmente supervivientes desde disco, preservando los tombstones preconfirmados, y publicar esa vista. Un fallo al escribir el índice final no autoriza reutilizar el índice anterior como si describiera las carpetas presentes: mantener la vista real con error explícito y reconstruir desde disco al reiniciar. El índice preconfirmado conserva los tombstones para impedir reingesta idéntica.
4. Si no quedan miembros, retirar el registro lógico pendiente. Si falla esa última retirada, no devolver éxito: el arranque siguiente detecta `delete_pending` sin miembros y completa la retirada antes de iniciar escaneo/ingesta. Si hay supervivientes, conservar su estado y reportar eliminación incompleta; no borrarlos automáticamente al arrancar. Reintentar desde UI vuelve a mostrar y confirmar el conjunto actual.

**Corrección final aprobada por D10-B (recuperación y ejecución):** no depender exclusivamente del siguiente arranque. El ciclo periódico de Sync también intenta retirar registros `delete_pending` sin miembros. Además, antes de ingerir contenido nuevo/modificado para esa identidad, intentar completar su retirada; si falla, mantener solo esa ingesta pendiente y publicar un error, sin instalarla con nombre/archivo/conservación anteriores. Una vez retirado el registro, admitir la nueva entrega con preferencias iniciales. No bloquear el resto de la biblioteca ni crear una exclusión permanente. Esta es la corrección acotada del revisor backend tras la única segunda ronda.

Un fallo después de retirar alguna copia devuelve 500 con código estable `partial_delete` y conteos de eliminados/restantes; nunca texto de fuente ni credenciales. El cliente consume ese cuerpo (su cliente actual conserva solo el status), recarga el índice, conserva la identidad lógica, muestra el error y ofrece reintentar con un miembro superviviente. Con cero supervivientes ofrece Reintentar actualización; el arranque completa la limpieza pendiente sin necesitar una clave física ya eliminada. No mostrar éxito ni cerrar el diálogo como si todo se hubiera borrado. Probar interrupciones antes/después de cada persistencia, fallo del índice final y cero miembros con fallo del último registro. Un error pendiente del registro se presenta en la biblioteca hasta completarse la recuperación.

Con la corrección D10-B (recuperación y ejecución), sustituir «Reintentar actualización» por «Comprobar recuperación» y explicar «Copias eliminadas; limpieza pendiente. El visor reintentará automáticamente». Esa acción solo consulta el resultado del ciclo automático, no afirma reparar ni provoca una escritura mediante GET. Si sigue fallando, mantener el error visible; al recuperar almacenamiento, Sync retira el registro pendiente y desaparece el aviso. El nuevo índice no necesita un miembro físico para mostrar o retirar ese error. Es la corrección del bloqueo UX residual: no ofrecer un botón de reparación que solo vuelve a leer.

## Vista y experiencia

Superficie operativa; arquetipo ReUI `Record detail`; densidad compacta; conservar Sidebar y composición actual con separadores, sin contenedores decorativos ni animación de entrada. `components.json` declara `base-nova`, tokens semánticos y Lucide. La guía aplicable es `.agents/skills/reui/SKILL.md`; no se encontró una guía UI dedicada. La diferencia intencional es el encabezado fijo de sesión; no se cambia el tema.

```text
Proyectos       | Encabezado global existente
                | Nombre de sesión [Editar] [acciones]
                | Fecha · tickets · recorridos
                |-----------------------------------
                | Entregas y saltos                  ^
                | Recorrido 1: título y estado       |
                | Recorrido 2: título y estado       | scroll
                | Recorrido 3: título y estado       v
```

- La barra de sesión queda fuera de `viewer-canvas`; no requiere `position: fixed` sobre el viewport. El requisito es que permanezca visible mientras el canvas se desplaza. No duplicar controles por recorrido. Los estados de trabajo siguen en cada recorrido; el encabezado no afirma que terminó el chat.
- Reutilizar `ButtonGroup`, `ActionButton`, `Dialog`, `AlertDialog`, `Input`, tooltips y `Skeleton` instalados. No añadir dependencias para una notificación: usar feedback breve en una región de estado de la barra con Deshacer, sin el banner persistente dentro del contenido y sin implementar un toast propio. No retirar errores reales, desconexión ni advertencias de procedencia.
- Archivar retira el conjunto de la lista activa y selecciona la siguiente sesión visible; si no existe, mostrar vacío con acceso a archivados. Mantener el feedback de Deshacer en el shell aunque la sesión haya dejado de estar seleccionada. Restaurar desde el filtro existente o Deshacer devuelve el conjunto; el filtro sigue incluyendo activas y archivadas, no se convierte en una vista exclusiva.
- Deshacer conserva identidad y estado previo de la operación, dura al menos ocho segundos y no desaparece mientras el foco esté en la acción. Solo revierte archivo; no se ofrece para borrar. Cambiar de selección no pierde el destino del Deshacer. Errores permanecen visibles con Reintentar; durante una escritura deshabilitar acciones conflictivas.
- Renombrar: diálogo con un único campo prellenado con el nombre mostrado; nombre del campo «Nombre de la sesión», Guardar, Cancelar y Restablecer nombre automático. Pedirlo una sola vez por sesión cuando la persona decida editar, nunca durante la ingesta ni por ticket. Fecha, proyecto, host y tickets se derivan, no se vuelven a solicitar. Validación inline y conservar entrada tras error; mientras guarda anunciar progreso. Acotar a15s la espera de PUT de nombre; al vencer, recuperar controles y refrescar índice, conservar entrada y explicar que el resultado no se confirmó y puede haberse guardado. No reenviar automáticamente ni presentar abort del cliente como rollback del servidor; Escape/cancelar no escribe; devolver foco al botón de edición al cerrar.
- La búsqueda incluye el nombre mostrado del conjunto además de títulos y tickets de sus recorridos. Buscar el nombre editado después de recargar encuentra la sesión completa; buscar un ticket sigue permitiendo el salto a su recorrido sin ocultar hermanos.
- Eliminar: confirmación con el nombre, total de recorridos almacenados (incluidos archivados/apartados) y texto que aclara que las fuentes no cambian y que nuevas entregas pueden reaparecer. No pedir escribir de nuevo un dato que el sistema ya conoce. Error parcial muestra restantes y Reintentar; cancelar no escribe.
- Conservar: etiqueta accesible «Conservar sesión» / «Dejar de conservar sesión», explicación de que evita caducidad. No llamarlo favorito.
- Barra móvil: mantener todas las acciones accesibles en un menú de más opciones si no caben; título largo con reflujo por palabras, sin truncar el nombre ni desplazar horizontalmente la página. No ocupar toda la altura útil. Navegar a un recorrido debe mostrar su encabezado sin quedar oculto. Imprimir omite herramientas y conserva el contenido y aislamiento actuales.
- Loading: esqueleto de barra y contenido, sin controles de escritura sobre selección obsoleta. Vacío: ninguna sesión y acceso a Mostrar archivados. Fallo de carga: error inline con Reintentar. Nombre/estado desconocido por lectura fallida no se presenta como ausencia confirmada.

## Verificación local prevista

Servir el bundle compilado desde Go con `--state-dir` aislado y puerto libre, nunca manipular las 20 copias cotidianas. Datos: una sesión de tres tickets, un ticket con dos recorridos, título largo y Unicode, sesión totalmente archivada, grupo histórico mixto, fuentes apartadas, sesión conservada y otra para borrado; crear continuaciones mientras está archivada/conservada.

Preparar candidato inmutable y base inmutable en puertos distintos, con la misma semilla reproducible; anotar hash del candidato y commit base. `agent-browser` está instalado y será el navegador, con sesiones/perfiles temporales y datos separados para cada verificador. Viewports: 1440×900, 900×800 y 390×844, temas claro/oscuro; medir altura estable de barra y ausencia de overflow horizontal. No usar el modo `?preview` como evidencia de API/auth/CSP.

`hive-review-ux` inspecciona layout, estados, teclado y accesibilidad de las vistas tocadas contra la base; `hive-verify-change` verifica acciones, recargas/reinicio, continuaciones y fallos. Un pase por área/viewport/tema; dos fallos o bloqueo detienen ese check y se reporta sin repetir indefinidamente. Los verificadores no comparten datos mutables.

Mostrar la primera vista renderizada al usuario durante el build. Después de checks locales, verificaciones independientes y Codex review, entregar el recorrido humano con el stack disponible y sus URLs concretas antes de push. Capturas/manifest en `_support/workspace/2026-10-26-viewer-session-toolbar/`, no en commits. Los cambios aceptados que alteren la vista requieren renovar solo la evidencia afectada. Suite y checkpoints concretos en `tasks.md`.

## Revisión de nombres sin espacios · AC12

La corrección de reflujo del candidato39187977 evita recortes, pero divide un token válido sin espacios. El ajuste seleccionado conserva el texto íntegro, sin truncar ni partir tokens, en regiones locales de overflow horizontal en h1 y nombre de la confirmación. El contenedor tiene min-width0 para no ensanchar barra/diálogo. Texto con espacios conserva wrap normal. No se añade scroll vertical a la barra; el canvas mantiene ese desplazamiento.

Cada región del nombre es alcanzable por teclado, con foco visible de los tokens semánticos existentes y nombre/descripción accesibles; flechas permiten llegar a ambos extremos y la salida con Tab es natural. Una ayuda de lectura puede anunciar el desplazamiento sin repetir datos ni pedir entrada nueva. Este punto de foco es deliberado y está cubierto por AC12; no hacer la sesión ni el diálogo completo otro contenedor de scroll horizontal. Cancelar sigue como foco inicial de confirmación, y Cancelar/Eliminar permanecen alcanzables dentro del trap.

Verificar200caracteres sin espacios y nombre largo de palabras normales a320/390/900/1440, y1280×720: ninguna palabra dividida, todo carácter accesible mediante scroll local, ningún overflow de página, acciones visibles y salida/foco por teclado. Repetir espaciado de texto pertinente. Usar composición/CSS nativos; no instalar componentes, añadir librerías ni cambiar el tema o componentes instalados. JSDOM solo verifica foco/semántica, nunca geometría de scroll.
