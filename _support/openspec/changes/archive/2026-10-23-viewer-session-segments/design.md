# Diseño: una sesión, varios tramos

## Contexto verificado

La [referencia del manifiesto](../../../../content/skills/flow-walkthrough/references/walkthrough-manifest.md) exige reutilizar `session.id` en la conversación: puede ser nativo o sintético. No se auditó si todos los históricos cumplieron esa regla. [scan.go](../../../../tooling/viewer/scan.go) genera la clave física con carpeta e ID; [index.go](../../../../tooling/viewer/index.go) mantiene la copia, host, ID, proyecto y ticket. [api.ts](../../../../tooling/viewer/frontend/src/api.ts) asigna la clave física a `id` y aplana el índice de proyecto/ticket/copias. [controller.ts](../../../../tooling/viewer/frontend/src/controller.ts) selecciona y escribe por esa clave.

El manifiesto completo se obtiene por el endpoint autenticado `GET /api/sessions/<key>`. [store.go](../../../../tooling/viewer/store.go) bloquea lectura normal de fuentes excluidas antes de devolver sus bytes. No hace falta añadir `delivery` al índice persistido: se obtiene de los manifiestos de los tramos cargados.

Los validadores [Go](../../../../tooling/viewer/manifest.go) y [Python](../../../../content/skills/flow-walkthrough/scripts/render_walkthrough.py) rechazan campos desconocidos. `candidate.commit` permite `working tree` y describe la evidencia mostrada, no la entrega. El [renderer compartido](../../../../content/skills/flow-walkthrough/assets/walkthrough.mjs) usa IDs `wt-flow-*`, `wt-panel-*` y listeners de impresión globales; [document.tsx](../../../../tooling/viewer/frontend/src/document.tsx) restaura scroll de un contenedor global al montar un documento. Se deben adaptar para varios documentos simultáneos.

### Impacto

Señales: hay lectores en procesos distintos (renderer offline Python y visor Go instalado, posiblemente anterior) y se añaden URLs externas. No se renombra ni retira un endpoint, proceso o almacén.

- Productores `flow-walkthrough` y cierre actual de `flow-close`: registrar sólo entregas conocidas en el manifiesto del trabajo actual. Sus referencias y docs del sitio cambian coordinadamente.
- Validadores Python/Go y renderer canónico/embebido: actualizar juntos; un lector viejo seguirá rechazando el campo nuevo. Publicar lectores antes de usar nuevos metadatos, sin dos formatos de escritura paralelos ni migración histórica.
- Frontend y API de copias: API/rutas/capacidades permanecen; la agrupación es derivada, las operaciones conservan el destino físico. El frontend cambia de un documento a varios montajes aislados.
- Páginas standalone: conservan una sola unidad, siguen mostrando candidato y añaden entregas cuando existan. No se regeneran históricos automáticamente.
- Daemon cotidiano ya iniciado: revisión efectiva no comprobada aquí; no se reinicia en este plan. La prueba final usa servidor compilado y store de fixtures propios. Un daemon viejo que no confirme la nueva evidencia obliga a conservarla según `flow-close`, nunca a forzar un reinicio.
- Fuentes documentales: frontend README, README/llms y arquitectura del manager si sus afirmaciones cambian; páginas `flows/build.md` y `flows/close.md`; spec `walkthrough-viewer`. No cambios a roles ni a la política global de identidad o autorización.

## Identidad, filtrado y orden

La identidad visual es una tupla exacta `(projectKey, host, session_id)`, serializada sin colisiones. No normalizar ni inferir el host de un valor vacío; una ausencia coincide sólo con otra ausencia del mismo proyecto/ID. Nunca unir por ticket, título, proximidad temporal ni prefijo de un ID. Conservar `raw.key` para cada tramo y para selección persistida/enlaces existentes.

Filtrar archivados antes de formar grupos visibles. Las fuentes excluidas permanecen en su sección explicativa actual, con acciones individuales y sin cargar documentos ni entregas. Cada entrada del sidebar cuenta sesiones visibles, muestra título del primer tramo visible y cantidad de tramos; muestra un ticket si es único y cantidad de tickets si hay varios. Una sesión aparece si queda al menos un tramo visible. El proyecto se abre inicialmente si un tramo visible declara `in_progress`; no inventar un estado de conversación cerrada.

Orden de tramos: instante de `created_at`, con `ingested_at` como alternativa para datos heredados; desempate por clave física. Orden de sesiones: actividad más reciente de sus tramos visibles, con desempate por identidad. Selección persistida sigue siendo la clave física; abrir la sesión vuelve al último tramo visible seleccionado, o al primero. Buscar título/ticket de un tramo devuelve su sesión y ofrece el salto correspondiente, manteniendo los hermanos visibles en la lectura. No crear relaciones nuevas en disco.

Al archivar/eliminar el tramo seleccionado, elegir siguiente visible y después anterior de la misma sesión; sólo después pasar a otra sesión. Cada menú recibe la clave de su tramo, nunca el grupo. Fijar no cambia el documento ni recarga hermanos sin cambios.

## Carga, actualización y composición

Al seleccionar una sesión, obtener los manifiestos de sus tramos normales visibles mediante las rutas actuales. Mantener estado y caché por clave física y revisión (`content_hash`, `fileBase`, clasificación); reutilizar objetos sin cambios. Invalidation por identidad de acceso y revisión, con descarte de respuestas tardías, incluida eliminación/archivo y reclasificación. No solicitar documentos de otras sesiones para indexar entregas.

Cada tramo tiene carga, error con Reintentar y contenido propio; un fallo parcial no borra a sus hermanos. Una pérdida de autenticación limpia todos los documentos y enlaces protegidos. Actualizaciones del índice no reinician scroll/foco de documentos sin cambios. La sección Entregas presenta carga/error parcial cuando no pudo leer todos los tramos; no llama «Sin entrega registrada» a un fallo de lectura.

Todos los documentos de la sesión están apilados en la misma región desplazable. El componente de sesión posee selección/saltos y restauración del scroll; cada documento posee sus medios, diálogos y limpieza. Extender `renderWalkthrough` con un prefijo de IDs por montaje y una disposición/limpieza explícita de listeners. El viewer usa una clave física estable como prefijo; el standalone de un solo documento conserva sus anclas normales. Ámbito de consultas por raíz, referencias ARIA consistentes, handlers de impresión retirados al desmontar y sin acumulación de listeners con StrictMode. Los diálogos, pestañas, enlaces de índice y foco del carrusel apuntan sólo al documento dueño.

## Contrato de entregas

Campo raíz opcional `delivery`, lista de objetos con `kind` (`pull_request` o `commit`), `label` (texto no vacío, por ejemplo `tricell-hive #123` o `abc1234`) y `url`. No campos nuevos en `candidate`, ni estado remoto inferido. Un PR enlazado puede estar abierto; la etiqueta no afirma integración. El productor sólo registra un commit con URL verificable; un hash local o `working tree` no se convierte en URL por conjetura.

Cada URL es HTTPS absoluta con host, sin credenciales, controles ni espacios literales; no se permiten esquemas ejecutables, relativos ni URLs con usuario/contraseña. D7-A (hosts) restringe el host a ASCII: nombres con letras ASCII, dígitos, punto, guion o guion bajo (incluido punycode), o literales IPv6 válidos entre corchetes. El puerto opcional conserva su rango 0..65535. Rechazar hosts Unicode sin normalizarlos; etiquetas y rutas admiten Unicode. La misma regla léxica se aplica en Python, Go y renderizado de enlaces para evitar diferencias entre parsers. `label` se trata como texto. Validadores Go y Python coinciden y el renderer vuelve a validar al crear `href`, con `target="_blank"` y `rel="noopener noreferrer"`. No consultas automáticas ni allowlist de un proveedor particular; no incluir secretos en URL o etiquetas.

El header reúne las entregas de tramos visibles cargados, deduplicadas por URL exacta en orden de los tramos; primera etiqueta prevalece y se conserva su asociación a los tramos. Mostrar dos enlaces y un control «Ver todas» con el resto en un Dialog instalado. El tramo mantiene sus enlaces sin deduplicarlos contra sus hermanos. Sin metadatos, «Sin entrega registrada»; una sesión previa sin campo sigue siendo válida. El standalone también presenta los enlaces de su propio manifiesto.

Se mantiene `schema_version: 1` con campo opcional en lectores nuevos. Los lectores viejos lo rechazan por diseño; la distribución coordinada actualiza binario y contenido juntos, binario primero. No se introduce downgrade/dual-write. Documentar que un proceso ya iniciado necesita reinicio explícito para leer el nuevo contrato. No retirar información ni reescribir fuentes/copias almacenadas: cero registros transformados por la agrupación; volumen histórico no auditado porque no hay migración ni backfill. Las unidades nuevas conservan las reglas existentes de confirmación y retención.

Al construir el walkthrough, añadir entregas ya verificadas si las hay; al cerrar, actualizar sólo la unidad actual con PR/commit conocido antes de renderizar/cerrar/confirmar. Si no hay una entrega conocida, omitir `delivery`; conservar enlaces ya registrados que siguen correspondiendo al trabajo. No barrer carpetas de otros tickets de la conversación. Las instrucciones esenciales viven en la referencia del manifiesto y los puntos de producción/cierre sólo la enlazan.

## Decisiones de interfaz

Aplicación: visor React local `/`, shell existente ReUI/shadcn `Sidebar` y `Frame`, arquetipo detalle de registro. Referencia visual: composición actual en `shell.tsx` y capturas de Orca dadas por el usuario; no se ha inspeccionado todavía una captura viva del visor. El build captura la línea base en su entorno aislado antes de editar, sin afirmar una mejora visual por lectura de código.

- Superficie operativa; contenedor Frame existente; densidad compacta con props admitidas.
- Regiones: navegación por proyecto/sesión; header de sesión y Entregas; saltos de tramos; lista apilada de tramos con título/ticket/fecha, acciones y documento.
- En el visor, evitar títulos de página grandes repetidos dentro de cada tramo; un modo embedded del renderer puede compactar su encabezado manteniendo candidato y demás metadatos. El standalone conserva su presentación. La primera vista renderizada valida esa composición con el usuario.
- Primer tramo rotulado «Trabajo inicial»; posteriores, «Continuación» con ticket y título cuando existan. Es una posición en la lectura, no una relación histórica inferida entre tickets. Sin ticket se conserva el título.
- Reutilizar Separator, Button, Tooltip, Skeleton y AlertDialog para borrar. El árbol actual no incluye Dialog: incorporar el componente del mismo estilo instalado para entregas adicionales, usando la dependencia Base UI existente y verificando su API vigente antes de usarla. No cambiar dimensiones/radios/movimiento de componentes instalados ni añadir dependencias de runtime.
- Estados: biblioteca vacía con acción de actualización, carga de sesión y tramo, fallos recuperables por región, fuente excluida, sin entregas, texto largo, varios PRs y estado de escritura individual.
- Teclado y foco: entradas seleccionables, salto a heading del tramo, acciones etiquetadas por tramo, retorno de foco al cerrar Dialog/carrusel y foco razonable al desaparecer un tramo.
- Responsive: 390 px con sidebar Sheet y header que envuelve sin scroll horizontal de página; desktop a 1280 y 1440 px. Conservar tema oscuro actual; no introducir un tema claro que el producto no ofrece.
- Movimiento propio: ninguno de entrada; conservar el de los componentes instalados.

## Publicación del walkthrough: JSON o fallback HTML

D9 (salida de walkthrough) adopta la condición explícita del usuario. Resolver una vez `shutil.which("hive")` al publicar, equivalente al descubrimiento en PATH del host. Con binario disponible: validar JSON y recursos, aplicar el estado explícito atómicamente si cambió, comprobar escritura concurrente y devolver la ruta de `walkthrough.json`; no leer assets para construir página, no crear ni reescribir HTML. Un HTML previo se conserva como evidencia histórica, sin presentarlo como salida actual. Sin binario: conservar la preparación/publicación de JSON y HTML actual y su recuperación de JSON ante error de publicación del HTML; devolver `walkthrough.html`.

La presencia del daemon, su versión, el registro o una confirmación fallida no cambian ese modo. El registro sigue en flow-walkthrough y la confirmación de cierre sigue después de publicar; si falla, salida2 y conservación de la unidad, sin iniciar/reiniciar visor ni generar HTML como alternativa. No flag público que omita la condición decidida. El nombre del script se conserva para sus consumidores.

Impacto adicional verificado: Go valida/copia JSON y capturas/grabaciones; no requiere ni copia walkthrough.html. No se necesita migración ni API nueva. Los consumidores que prometen siempre página se actualizan: flujo walkthrough, cierre, referencia de manifiesto, artifact-layout, README/llms, arquitectura y sitio. Handoff con binario señala JSON registrado y capturas/visor; sin binario señala HTML. La unidad de retención es manifiesto, recursos y página sólo cuando existe. Seleccionar conservar el manifiesto o la página conserva sus recursos referenciados; seleccionar capturas/grabaciones individuales conserva sólo esos archivos según el hogar de retención. No cambia la producción de reportes HTML de flow-report ni se borra evidencia previa.

Las pruebas deben controlar disponibilidad sin depender de PATH cotidiano: directorio temporal vacío para ausencia usando Python absoluto, fake CLI ejecutable para presencia y mocks de shutil.which para llamadas directas. Cubrir binario presente/daemon apagado, HTML previo idéntico, build_page nunca llamado, validación y estado/candidato/delivery preservados, confirmación sin exigir HTML, fallback sin binario y recuperación/concurrencia existentes. No se ejecutan binarios cotidianos desde esas pruebas.
