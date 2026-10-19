# Diseño del visor

## Contexto verificado

Inspección del 2026-10-19 en `eee7cccd994a15bd4750b548352377e4f559e950`, coincidente con `development` remoto. No hay implementación de este plan.

- `tooling/viewer/scan.go:424–450,550–613`: descubre carpetas y omite fuentes por huella, sin tracking Git. `copyFile:109–148` usa descriptor abierto y `Stat`, pero no `Nlink`.
- `manifest.go:127–143`: abre el manifiesto después de comprobar la ruta, sin `f.Stat()` antes de leer. Validación de recursos en `551–580` rechaza symlinks, no hardlinks.
- `index.go:176–215,433–442`: proyección agrupada y enriquecimiento derivados. `store.go:204–228,285–335`: `meta.json` estricto y reconstrucción de índice desde copias. Es posible añadir clasificación derivada sin modificar metadata durable de cada copia.
- `registry.go:88–99` admite raíz de workspace sin Git; `view_command.go:638–660` evita heredar remoto para agrupación, una política distinta de averiguar qué repositorio versiona un archivo.
- `walkthrough.css:80–92`: tira horizontal, captura desktop de media anchura, asomo hasta 520 px del contenedor y recorte `cover`. `viewer.css:126` amplía el contenedor montado. Impresión `walkthrough.css:153–155` muestra todas las vistas y envuelve las figuras.
- Investigación previa: `go test ./tooling/viewer` pasó en 6.172 s. Esto verifica la base, no las nuevas conductas. No se ejecutó navegador ni explotación multiusuario. El detalle documental de hardlinks se verificó contra Go 1.27.0 y documentación oficial Linux; no se estableció explotabilidad universal de macOS.

## Clasificación de fuentes

Contrato de servidor, nunca tomado del manifiesto: `Session.source_status` es `tracked`, `local` o `unknown`; `source_reason` contiene un motivo acotado y seguro cuando se necesita explicación. No hay fecha histórica de autoría ni nuevo `schema_version` del manifiesto.

- `tracked`: ruta del manifiesto presente en el índice Git propietario, incluso si el archivo desapareció del árbol de trabajo.
- `local`: fuente presente y consulta Git satisfactoria sin tracking, o ausencia comprobada de repositorio. Etiqueta: «Archivo local; autoría no verificada».
- `unknown`: fuente ausente sin evidencia positiva, consulta fallida, acceso insuficiente o clasificación pendiente. Etiqueta visible: «Origen sin determinar», con explicación. Se conserva la consulta normal con esta advertencia, como permite D2-A (procedencia).

Inyectar por `viewer.Options` un lector de tracking desde el CLI siguiendo las capacidades inyectadas actuales. Resolver el repositorio que realmente contiene la fuente, incluidos ancestros y `.git` archivo de worktree; no copiar la restricción de remoto heredado que sólo sirve para agrupación. Consultar por lote y raíz propietaria con salida NUL y argumentos estructurados, sin shell. Distinguir `no_repository` de error de ejecución/acceso; no tratar cualquier salida no cero como ausencia de Git. Acotar consultas a los prefijos relevantes, sin ejecutar procesos por manifiesto. El lector ausente produce `unknown`, no `local`.

El runner actual (`doctor_project.go:47,83`, `doctor.go:102`) limita stdout a 64 KiB y descarta el exceso sin error. No reutilizarlo para inferir ausencia de tracking sin detectar explícitamente desbordamiento. El nuevo lector debe conservar límites de memoria/tiempo y devolver `unknown` para las fuentes que no puede clasificar ante salida truncada, aunque el corte coincida con un NUL. No ampliar ni cambiar el runner compartido salvo necesidad demostrada por sus consumidores; resolver el contrato en el lector de tracking basta. No inferir lista completa porque la última entrada esté terminada.

Corrección de seguridad v5: ejecutar esta ruta de lectura con `core.fsmonitor` desactivado explícitamente para el proceso, sin modificar configuración local/global. Las variables de entorno que cambian semántica de pathspec no deben alterar los argumentos elegidos; controlarlas en el entorno del proceso hijo, además de las variables locales al repositorio que Git enumera. Comprobar el efecto con Git instalado 2.56.0 y documentación primaria de la semántica elegida.

Enfoque contrastado por implementador: `-c core.fsmonitor=` en los procesos de este lector; valor vacío para evitar que versiones antiguas interpreten `false` como ruta de un hook. Limpiar `GIT_LITERAL_PATHSPECS`, `GIT_GLOB_PATHSPECS`, `GIT_NOGLOB_PATHSPECS` y `GIT_ICASE_PATHSPECS` sólo del entorno hijo. Consultar prefijos con la magia explícita `icase`, y confirmar alias con `os.SameFile` en vez de plegar cadenas indiscriminadamente. Unir raíces que sean alias del mismo repositorio por identidad del filesystem para mantener una consulta por propietario real.

Fuentes primarias consultadas para el valor vacío: [documentación de core.fsmonitor](https://git-scm.com/docs/git-config#Documentation/git-config.txt-corefsmonitor), [Git 2.35.1 config.c](https://github.com/git/git/blob/v2.35.1/config.c#L2364-L2375), [Git 2.56.0 fsmonitor-settings.c](https://github.com/git/git/blob/v2.56.0/fsmonitor-settings.c#L97-L138). No se introduce detección de versión ni una segunda implementación.

Comparar la identidad real de las rutas cuando su escritura difiere: Git conserva el casing del índice y el filesystem puede resolver nombres distintos al mismo archivo. La selección de prefijos debe incluir los alias de mayúsculas de los directorios; el matching no puede depender de `core.ignorecase` del repositorio ni convertir todas las rutas a minúsculas. Un archivo distinto en un filesystem case-sensitive permanece distinto. Fuente ausente sin evidencia suficiente conserva `unknown`, como ya establece la política; no inventar historial de identidad.

Cada `Sync` clasifica candidatos y copias existentes antes de los retornos de caché/TTL; no reutiliza la caché de nombres/remotos de 60 s. Un índice que cambia se refleja en el siguiente barrido completo. Omitir nuevas ingestas `tracked` antes de validación/lectura/copia, sin crear tombstones. Las fuentes `local` y `unknown` continúan por la validación de archivos habitual.

Reclasificar también las copias cuya fuente no figura en el barrido. En `Open`/reconstrucción, completar clasificación antes de publicar la primera vista. No añadir marcas temporales renovadas por poll ni historial. El índice es derivado; `meta.json` no cambia. Reiniciar una versión anterior no pierde metadata ni archivos, pero vuelve a la conducta previa sin la protección nueva.

## Vista y API

Conservar todas las copias en `indexData.Sessions` para las operaciones existentes de retención. `buildView` agrupa `local`/`unknown` en proyectos y tickets; expone las `tracked` separadas como `excluded_sessions`, sin duplicarlas en el árbol normal ni en sus conteos. El índice autenticado devuelve estado y razón. API y frontend embebido se actualizan coordinadamente en el mismo binario.

Reutilizar el patrón de disclosure del pie de barra lateral para «Copias apartadas (n)», con una fila por copia. Seleccionar una abre el panel explicativo existente, no el renderer del manifiesto: título, ruta fuente, pin y motivo de apartamiento. Mantener fijar/desfijar y eliminar mediante las operaciones actuales. No mostrar instrucciones del documento apartado.

`GET` de manifiesto o archivo de una copia `tracked` devuelve 409 sin bytes del documento; no usar 403, reservado a fallos de autenticación que disparan otro estado UI. La respuesta se basa en el índice publicado, sin Git por solicitud. El frontend distingue el 409 de autenticación y actualiza la clasificación, también si la sesión estaba abierta al cambiar de tracking. La próxima actualización del índice desmonta el recorrido y cierra cualquier carrusel/video de una sesión apartada. Las mutaciones pin/delete conservan validación de token y origen.

Una copia `unknown` recibe aviso en fila y documento; una `local`, etiqueta de autoría no verificada. Los textos de fuente no llegan a `innerHTML`. Errores de Git se convierten en motivos propios, acotados; no devolver stderr crudo ni credenciales/remotos.

Actualizar los avisos de la sesión abierta cuando cambien `source_status` o `source_reason`, independientemente de `content_hash`. `applyIndex` hoy sólo remonta por hash: el cambio `local` a `unknown` o inverso debe actualizar el aviso sin reingesta ni depender de un cambio de bytes. La advertencia debe corresponder al último índice recibido.

## Archivos y carreras

Una comprobación pequeña común sobre `os.FileInfo` rechaza `Nlink > 1` mediante `*syscall.Stat_t` en las plataformas Unix soportadas. Si no puede obtenerse esa información, rechazar la ingesta con error explícito; no fingir una comprobación satisfactoria. No añadir una interfaz de validación extensible ni dependencia.

Comprobar sobre `f.Stat()` del manifiesto abierto antes de `ReadAll`; comprobar sobre `src.Stat()` de recursos antes de hash y copia en `copyFile`. Mantener regularidad y límites sobre el descriptor además de validaciones de ruta. El mismo camino cubre la segunda apertura de copia tras el hash. No inspeccionar ni exigir `Uid`. La seguridad se limita a lo observado al abrir: no garantiza inmutabilidad frente a escritores autorizados, procedencia ni aislamiento de procesos del mismo usuario.

Los errores de lectura/validación no instalan una copia parcial y conservan la última copia válida. No inspeccionar secretos ni limpiar retrospectivamente archivos almacenados. Usar fixtures sintéticos para enlaces y permisos.

## Capturas e interfaz

Superficies: visor en loopback y HTML independiente generado; ambos reutilizan `walkthrough.css`. Reusar shell, colores oscuros, pestañas, carrusel y foco actuales. En pantalla y tira ancha, `.wt-item.desktop` ocupa todo el ancho útil; hasta 520 CSS px se conserva el asomo de 48 px. Mantener miniaturas de capturas móviles y reglas de impresión actuales mediante alcance de medio explícito. No ampliar máximos de página ni cambiar `object-fit` en este trabajo.

Actualizar CSS canónico y copia embebida byte por byte. HTML históricos ya generados conservan su CSS incrustado; sólo nuevas generaciones y el binario actualizado presentan el cambio. No reescribir páginas históricas del usuario.

Referencia visual pendiente de ejecución: investigación estática, sin render usado como prueba de legibilidad. Validar tira ancha, estrecha, ausencia de capturas y varias capturas; recorrido local, desconocido y apartado. Usar 1440×900 y 1280×800, 390×844 y 320×800 CSS px; comprobar ambos lados del umbral 520 del contenedor y del cambio de shell a 719 de viewport. Tema oscuro; impresión clara. Captura realista sintética de 2880 px con viewport/DPR conocidos, sin datos del usuario.

## Persistencia, volumen y recuperación

Conteo sólo lectura del estado por defecto de macOS el 2026-10-19: directorio de visor inexistente; cero copias y cero pins. No se consultaron overrides de otros procesos ni instalaciones de otras máquinas; su volumen es desconocido. No migrar metadata de copias: clasificación derivada y proyección, conservando el formato actual. Fixtures incluyen copias antiguas, fijadas, fuentes borradas e índice perdido para probar reconstrucción y política.

Apartar no cambia `IngestedAt`, hash, claves, pins ni tombstones. Las copias no fijadas siguen venciendo por el TTL existente; no prometer conservación eterna. No borrar copias por clasificarlas. Volver al binario anterior conserva datos pero pierde protección de procedencia/hardlinks y presentación nueva; es reversión de código, no recuperación de datos borrados por una operación explícita o TTL.

Refrescar el ejecutable no reemplaza un proceso de visor ya activo ni sus assets embebidos. Al entregar, comprobar el estado del proceso sin exponer credenciales y reportar por separado binario/contenido actualizados y proceso activado. No reiniciar el visor cotidiano en este trabajo: si estaba activo, el nuevo comportamiento entra en su próximo arranque. La reversión de binario tiene el mismo límite: tampoco cambia un servidor activo hasta volver a arrancarlo. La validación funcional usa sólo el servidor aislado construido para el candidato.

## Fuentes técnicas para implementación

Leer `AGENTS.md`, `CONTRIBUTING.md`, `/Users/jmartinez/.agents/skills/flow-build/references/security-boundaries.md`, `data-changes.md`, `verification.md` y `browser-automation.md` según superficie. Para assets distribuidos, leer `_support/docs/architecture/instruction-resources.md`. No modificar reglas distribuidas: cambio previsto de tamaño de `content/guidance/global.md` igual a cero. Revisar README/llms/site y actualizar sólo afirmaciones afectadas.

## Correspondencia de verificación

La revisión v4 separa criterios por resultado: T1 (clasificación) verifica AC1/AC2/AC4 (ingesta y datos del índice), T2 (presentación) verifica AC3/AC7 (apartamiento/API y avisos UI). Es la misma conducta de este diseño, sin requerir que la UI de T2 exista para verificar el backend de T1.
