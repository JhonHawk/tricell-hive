# Barra fija y acciones de sesión del visor

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en PR #227, commit eced053f; refresh local verificado |
| Tracker · GitHub Issues | Sin ticket asociado a este cambio |
| Git | Interactiva · worktree desde `development` · validación humana antes de push |
| Verificación | Pruebas locales · Codex review · UI completa independiente · validación humana |
| Siguiente paso | Publicar este registro por el PR de cierre autorizado; conservar el recorrido durante la sesión |

## Solicitud

> como que cada sección tiene su apartado para archivar, pero realmente no es por "sección" (hay que ponerle nombre) es por toda la sesión, como si fuera este chat, es una sola sesión sin importar cuantos tickets se resuelvan

> Igual que esta parte, no le veo sentido, no tiene acción alguna

La segunda observación señala el aviso persistente «Copia archivada. Puedes mostrarla desde Proyectos» de la captura compartida en la conversación.

> mejor vayamos por una opción tipo barra de herramientas (fixed, el contenido es el que scrollea) y tambien D1-C

En esta solicitud, D1-C significa continuar el análisis de títulos, selección y alcance de los controles.

> si, me aparece bien, comienza el plan para esto (lo trabajaremos en worktree)

Fecha inicial: 2026-10-26. Este registro comenzó como planificación; D10-B (ejecución del plan) autorizó después la implementación. La entrega sigue condicionada a la validación humana antes de push y no modifica la instalación activa durante el build.

## Objetivo y alcance acordado

Presentar una conversación como una sesión del visor con varios recorridos, aunque estos documenten distintos tickets. Mantener las herramientas de sesión visibles mientras solamente el contenido central se desplaza. Archivar debe afectar al conjunto, sin dejar hermanos visibles por haber actuado sobre una sola copia.

- Una barra fija de sesión con título y acciones del conjunto; conservar la barra lateral y el sistema visual existente.
- Contenido desplazable con entregas y recorridos cronológicos; usar «sesión» para el conjunto y «recorrido» para cada entrega visual, con tickets como referencias.
- Archivo y restauración de la sesión; sustituir el aviso persistente por confirmación breve con Deshacer.
- Acceso explícito a las sesiones archivadas desde la barra lateral, reutilizando el control existente cuando resulte adecuado.
- Nombre propio editable y persistente en el visor, sin modificar el título de los recorridos ni pedir al usuario los tickets o la fecha que el sistema ya conoce.
- Conservar y eliminar afectan al conjunto desde la misma barra; nuevas entregas heredan archivo y conservación. Eliminar solo afecta a las copias actuales y admite posteriores entregas nuevas o modificadas.

## Criterios acordados

- AC1. Al desplazar una sesión larga, su barra de herramientas permanece visible y el contenido central se desplaza sin cubrir sus controles. Se refuta si el título o las acciones salen de la vista con el contenido, o si la barra tapa el recorrido al navegar a él. La captura aportada muestra el encabezado actual dentro del contenido desplazable; confirmar contra la base de ejecución antes de implementar.
- AC2. Archivar una sesión con recorridos de varios tickets la retira completa de la lista activa; restaurarla devuelve el conjunto. Se refuta si un recorrido hermano continúa visible por no haber recibido la acción. En la base inspeccionada cada botón apunta a una sola clave física.
- AC3. El archivo ofrece una confirmación breve con Deshacer, sin insertar el aviso persistente «Copia archivada» encima de los recorridos. Se refuta si el aviso permanece ocupando el área de contenido o si Deshacer no restaura el conjunto. El aviso persistente está presente en la captura aportada.

- AC4. Conservar una sesión protege todos sus recorridos, incluidos los posteriores, frente al vencimiento; dejar de conservar restablece la retención ordinaria. Se refuta si un hermano o una continuación se purga mientras la sesión está conservada. La base solo conserva la copia indicada.
- AC5. Eliminar sesión requiere confirmar el alcance y elimina todas sus copias locales, incluso ocultas; una entrega nueva o modificada puede reaparecer. Se refuta si quedan hermanos eliminables sin explicación de error o si la ingesta de una entrega nueva queda bloqueada por identidad. La base ofrece eliminación por recorrido sin confirmación del conjunto.
- AC6. El nombre editado se usa en barra, navegación y búsqueda, persiste al recargar, reiniciar y reconstruir el índice, y no altera identidad ni títulos de recorridos. Se refuta si buscarlo no encuentra la sesión completa, se pierde o cambia las fuentes. La base carece de operación de renombrado y busca solo títulos/tickets individuales.
- AC7. Una nueva entrega de una sesión archivada hereda el archivo y no la hace reaparecer en la lista activa. Se refuta si esa continuación resulta visible sin restaurar la sesión. La base guarda el archivo solamente por copia.
- AC8. La primera apertura con el nuevo almacenamiento importa las preferencias históricas: una sesión completamente archivada sigue archivada, una mixta queda activa y una con alguna copia conservada queda conservada. Se refuta si la transición vuelve a separar el estado de los hermanos, pierde contenido, o deja la sesión mixta oculta. En la base hay una sesión mixta y no existe normalización por sesión.
- AC9. Escribir o limpiar el nombre local mediante almacenamiento conserva el valor al reabrir y reconstruir el índice, sin modificar los documentos fuente. Se refuta si esas lecturas pierden el override o alteran los títulos originales. La base no tiene un registro ni escritor de nombre lógico. Este criterio permite verificar almacenamiento antes de tener UI; AC6 (nombre visible y buscable) exige el frontend posterior.
- AC10. El borrado de sesión confirma tombstones antes de retirar carpetas y recupera supervivientes/limpieza pendiente tras fallos o reinicio, sin reingerir contenido idéntico eliminado. La corrección propuesta D10 añade reintento durante Sync y antes de una ingesta nueva, sin heredar preferencias obsoletas ni exigir reinicio. Se refuta si un fallo de persistencia previo elimina archivos, si un reinicio resucita contenido idéntico ya retirado o si una continuación recibe preferencias del registro pendiente eliminado. La base borra una copia antes de confirmar el tombstone. La confirmación humana pertenece a AC5 (eliminación desde UI), no a este criterio de almacenamiento/API.
- AC11. Una escritura sin la marca explícita de alcance de sesión recibe `reload_required` sin modificar datos. Se refuta si una pestaña antigua puede eliminar o cambiar el conjunto con su llamada por copia. La base acepta esas llamadas sin marca. Este guard evita ampliar silenciosamente el alcance de clientes abiertos.

- AC12. Un nombre válido sin espacios permanece sin dividir y se puede leer completo mediante desplazamiento horizontal local por teclado, con foco visible, en barra y confirmación; el resto del texto refluye por palabras y la página no se desplaza horizontalmente. Se refuta si un token se divide entre líneas, queda inaccesible, se pierde foco o las acciones quedan fuera. En la base no hay nombre local ni región enfocables de lectura del nombre; el candidato39187977 parte el token. Este criterio cubre el nuevo punto de foco/lectura que necesita el arreglo.

Las observaciones contra base se comprobarán de nuevo al registrar el commit inicial de implementación. Los criterios y sus tareas están en `tasks.md`; ningún check ejecutado durante planificación demuestra estos comportamientos futuros.

## Restricciones

- El archivo opera sobre la biblioteca local del visor; no archiva ni modifica la conversación del host ni sus archivos fuente.
- Conservar la identidad por proyecto, host e identificador de sesión; un título o ticket no define identidad.
- El estado de trabajo de un recorrido no prueba que la conversación terminó. No presentar un agregado «Cerrado» como estado de la conversación.
- Mantener recursos, navegación interna y referencias de accesibilidad aislados por recorrido.
- Respetar ReUI y los componentes instalados; conservar temas, navegación por teclado y acceso móvil.
- Preservar los archivos ajenos no versionados `_support/evidence/` y `_support/sessions/` del checkout principal.

## Contexto verificado

Inspección del checkout principal en `development`, revisión `82ccf111`, el 2026-10-26; no equivale a verificación del runtime de las capturas.

- `tooling/viewer/frontend/src/sessions.ts`, `sessionIdentity`: la agrupación ya utiliza proyecto, host y `session_id`.
- `tooling/viewer/frontend/src/application.tsx`: cada recorrido construye controles y llama a `controller.archive(segment.id)`.
- `tooling/viewer/store.go`, `Archive`: el archivo persiste por copia en `meta.json`; no modifica el manifiesto fuente.
- `_support/openspec/specs/walkthrough-viewer/spec.md`: navegación y archivo actuales exigen acciones por copia. El cambio necesita deltas, no solamente reorganización visual.

## Búsqueda de requisitos

Se inspeccionó la capacidad `walkthrough-viewer`, buscando `session`, `segment`, `title`, `archive`, `pin`, `delete` y las prohibiciones `SHALL NOT`. Requisitos relevantes leídos: `Walkthrough ingestion and grouping`, `Store retention and deletion`, `Current source classification`, `Explicit walkthrough session lifecycle`, `Hierarchical React viewer navigation`, `Reversible copy archiving` y `Explicit walkthrough delivery links`.

Deltas completos en `specs/walkthrough-viewer/spec.md`: modificar retención/eliminación, navegación y archivo; añadir barra fija, nombre/preferencias locales e importación histórica. La ingesta mantiene identidad y recursos físicos; su terminología interna de segmento no cambia el formato de manifiesto.

## Decisiones del usuario · 2026-10-26

- D2-A (alcance de controles): archivo, conservación y eliminación de toda la sesión.
- D3-B (nombre): título propio editable y local al visor.
- D4-A (archivo de continuaciones): heredar el archivo por identidad de sesión.
- D5-A (entrega): interactiva; commits locales, checks, revisión, validación humana antes de push; tras aceptación, push/PR/merge a `development`. El registro se publica una vez al cerrar mediante otro PR sin revisión dedicada, integrado por el agente.
- D6-A (revisión): revisión nativa de Codex del candidato local antes de mostrarlo o publicarlo.
- D7-A (UI): `full`; `hive-review-ux` y `hive-verify-change` con datos aislados. La respuesta rige esta sesión en el repositorio y puede cambiarse.
- D8-A (histórico mixto): si hay algún recorrido activo, dejar activa la sesión completa. Solo archivar al importar cuando todos estaban archivados.
- D9-A (eliminación y futuro): eliminar copias actuales con los tombstones existentes; admitir contenido posterior nuevo o modificado, sin bloqueo permanente por identidad.

D10-B (recuperación final y ejecución), elegido por el usuario el 2026-10-26: adoptar el reintento automático durante Sync y antes de reingesta, feedback que distingue consulta y reparación, aprobar AC1–AC11 e iniciar flow-build en worktree. Esta decisión resuelve los dos bloqueos residuales de la segunda revisión; no requiere una tercera revisión.

La conservación futura sigue el alcance de sesión elegido en D2-A (controles del conjunto); al importar se conserva si cualquier copia estaba conservada, para no reducir su retención.

## Entrega y ejecución

- D10-B (ejecución del plan) autoriza implementar; la entrega y su pausa antes de push siguen D5-A (interactiva).
- Worktree creado: `.claude/worktrees/viewer-session-toolbar`, ruta ignorada; rama `feat/viewer-session-toolbar`, desde `origin/development` en `82ccf111a16ffefce3684752beea26d2e17a5a3b`, igual al development local tras fetch. Registro trasladado a ese worktree como única copia canónica, sin versionar.
- `CONTRIBUTING.md` exige PR hacia `development`, no hay CI; ejecutar la suite local requerida. No se autorizó promoción a `master`, release ni publicación del sitio.
- Marcador mantenedor presente: al integrar runtime/assets, reconstruir el binario y reiniciar únicamente un visor ya activo, con puerto, directorio de estado y TTL previos. Esto sucedería después de la validación y el merge, no durante el plan. Respaldar antes del primer arranque que migre la biblioteca, bajo el procedimiento de `design.md`.
- Si la implementación cambia solamente `tooling/`, no ejecutar `hive update` para contenido no modificado.

## Temas y destino

- Unidad de archivo y aviso sin acción: este cambio.
- Barra fija y desplazamiento del contenido: este cambio.
- Nombres, títulos, selección y alcance de controles: decisiones de este cambio antes de sus tareas dependientes.
- Ejecución aislada: worktree `.claude/worktrees/viewer-session-toolbar`, rama `feat/viewer-session-toolbar`; registro canónico trasladado allí.

## Progreso

Diseño y decisiones incorporados; revisiones inicial y acotada completadas, con corrección residual adoptada por D10-B (recuperación y ejecución). Estado de construcción actualizado: T1 (preferencias) y T2 (API/recuperación) verificados independientemente; T3 (interfaz) y T4 (documentación/bundle) reportados y con C3 pendiente. Commits locales: 2158b833,6a2c29af,60808eb9,deb8e279. No se creó ticket ni se publicó Git; el runtime cotidiano no se reinició. El recorrido propio se registró localmente para mostrar la primera vista, sin cambiar las copias previas. Volumen local verificado en lectura: 20 copias, 16 sesiones lógicas, 17 copias archivadas, cero conservadas, un grupo mixto de archivo; otras instalaciones no son accesibles y su volumen es desconocido. La transición es por biblioteca local y se valida con fixtures grandes sin extrapolar tiempos.

La revisión de este borrador agrega criterios de conservación, nombre, continuación y transición histórica; sustituye el borrador parcial inicial y completa las tareas pendientes. Esta fue una revisión previa al build; los avances y verificaciones posteriores constan en tasks.md.

## Ajuste de presentación durante verificación

La revisión UX encontró que el nombre válido sin espacios necesitaba presentación completa sin corte ni división. Se adopta dentro de la continuación autónoma un área horizontal local con foco para el nombre en barra y confirmación; los nombres con espacios conservan reflujo normal, y solo el documento se desplaza verticalmente. AC12 (lectura accesible del nombre íntegro) cubre el nuevo foco y navegación. Motivo: respetar el límite aprobado de200puntos y conservar intactos identificadores sin ocultar controles. Alternativa descartada: partir tokens con overflow-wrap:anywhere, que resolvía el desbordamiento pero incumplía el criterio de integridad. Es una decisión de composición local, no una excepción de tema ni una autorización de publicación. La validación humana final puede pedir otra presentación. Revisión acotada de este ajuste en frontend/UX pendiente antes de editar sus fuentes.
