# Tareas · barra de sesión del visor

Estado: completado e integrado mediante PR #227 (barra y acciones de sesión), commit eced053f. La implementación fue autorizada por D10-B (recuperación final y ejecución). Base de ejecución: `82ccf111a16ffefce3684752beea26d2e17a5a3b`, reconciliada con origin/development. Worktree: `.claude/worktrees/viewer-session-toolbar`; rama `feat/viewer-session-toolbar`.

## T1 — Persistir las preferencias de sesión e importar las copias históricas

- [x] Guardar nombre, archivo y conservación por identidad lógica; importar preferencias históricas y aplicarlas a nuevas ingestas y retención.

**Closes:** AC4, AC7, AC8, AC9.

**Checkpoint:** C1.

**Dependencias:** ninguna.

**Ubicaciones:** `tooling/viewer/{store,index,scan}.go`, nuevas unidades de almacenamiento dentro de `tooling/viewer/`, `archive_test.go`, `store_test.go`, `index_test.go`, `scan_test.go`, `lifecycle_test.go`.

**Ejecución:** delegado al responsable backend, dueño de la persistencia y sus pruebas; sin escrituras simultáneas del hilo principal en esos archivos. La entrega al frontend es el estado efectivo y el nombre local por copia, definidos en `design.md`.

**Enfoque de pruebas:** `tdd`; ciclos independientes por importación, persistencia de nombre, archivo heredado y conservación heredada, extendiendo las pruebas existentes que posean ese comportamiento.

**Cambios:** implementar el formato y recuperación de `design.md`; mantener fuentes/recursos e identidad física; importar archivo con todos/conservación con alguno, retirar campos legacy como autoridad; no usar un índice reconstruible como único estado lógico. Antes de transformar datos reales repetir el conteo y preparar respaldo; los tests usan únicamente bibliotecas temporales.

**Verificación:** `go test ./tooling/viewer -run 'TestSessionGroup(State|Migration|Retention|Ingestion|Recovery)' -count=1`. Esos nombres son pruebas nuevas previstas, no existentes: verificar que la selección ejecute casos. Cubrir: dos tickets misma tupla, tuplas distintas, nombre persistente/reconstrucción de índice, todos archivados, mixto activo, cualquier pin, continuación archivada/conservada, conservación frente a TTL, importación interrumpida y repetida, archivo grupal corrupto, error de escritura sin éxito falso. Comparar hashes y conteos de recursos sin cambios. T1 verifica AC9 (nombre persistente en almacenamiento); el criterio completo de nombre visible/buscable AC6 se verifica solamente con T3 (interfaz de sesión).

## T2 — Exponer acciones del conjunto con alcance y fallos explícitos

- [x] Resolver toda la sesión en servidor para archivo, conservación, nombre y eliminación; entregar estados coherentes al cliente.

**Closes:** AC9, AC10, AC11.

**Checkpoint:** C2.

**Dependencias:** T1 (persistencia y preferencias lógicas).

**Ubicaciones:** `tooling/viewer/server.go`, `store.go`, pruebas de servidor/archivo/exclusión; `tooling/cli/view_command_test.go` para compatibilidad del índice extendido.

**Ejecución:** el mismo responsable backend después de T1 (persistencia), porque comparte lock y almacenamiento; no en paralelo con otro escritor de esos archivos.

**Enfoque de pruebas:** `tdd` sobre HTTP real de prueba y store, sin sustituir módulos internos.

**Cambios:** conservar rutas como localizadores físicos y cambiar su alcance de escritura al conjunto, añadir PUT de nombre y `session_title` opcional al índice. Ver `design.md` para cuerpo, límites, respuestas y borrado parcial. No relajar autenticación/origen. Exigir la marca X-Hive-Action-Scope en escrituras para rechazar pestañas antiguas. Confirmar tombstones antes de retirar carpetas y completar recuperación de delete_pending según design.md. Conservar GET/manifiestos/capacidades por copia y el fingerprint que usa CLI.

La corrección propuesta D10 (recuperación final) añade reintento de limpieza en Sync y guard antes de nueva ingesta para un registro pendiente sin miembros; adoptada explícitamente por D10-B (recuperación y ejecución). Añadir el caso de entrega modificada sin reinicio después de fallar la última retirada, luego recuperar almacenamiento y comprobar preferencias nuevas.

**Verificación:** `go test ./tooling/viewer -run 'TestSessionGroup(API|Delete|Title|LegacyClient)|TestArchive|TestExcluded' -count=1` y `go test ./tooling/cli -run 'TestView' -count=1`; comprobar selección no vacía. Casos: archivo/conservación desde miembro visible abarca hermanos ocultos/apartados; operaciones idempotentes y otra tupla intacta; título válido/vacío/Unicode/control/largo/tipo/campo desconocido; token/Origin/Host ausentes o erróneos; método/clave inválidos; borrado total/tombstones y contenido nuevo permitido; fallo parcial con índice y supervivientes recuperables; fallo de commit y reinicio, tombstones antes de borrar, interrupciones entre cada persistencia y cero miembros con limpieza grupal fallida; rechazar escrituras del cliente anterior sin mutar y aceptar las nuevas con marca de alcance; no alterar fuentes ni habilitar lectura de apartados.

## T3 — Entregar la barra fija y los flujos de sesión

- [x] Una barra estable de sesión permite nombrar, archivar/restaurar, conservar y eliminar el conjunto; recorridos desplazables sin controles repetidos.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC12.

**Checkpoint:** C3.

**Dependencias:** T2 (contrato HTTP de sesión).

**Ubicaciones:** `tooling/viewer/frontend/src/{application,shell,sessions,controller,api}.ts*`, `styles.css`, i18n y `src/tests/{client,component}.test.*`; componentes instalados de diálogo/input/confirmación cuando se reutilicen.

**Ejecución:** un único responsable frontend delegado, dueño del encabezado, selección y feedback como resultado coherente; el hilo principal integra y documenta. Esta distribución conserva un solo escritor de la vista y no mezcla sus componentes con backend. Ajuste operativo al comenzar T3, sin cambio de contrato ni criterio.

**Enfoque de pruebas:** `tdd` para estado/acciones de cliente, renombrado y recuperación; `check` renderizado para geometría y desplazamiento, que el DOM de pruebas no puede demostrar.

**Cambios:** seguir la composición, campos, estados, foco, impresión y feedback definidos en `design.md`. Mantener la selección física para saltos; no inferir el fin del chat del estado de trabajo. Conservar advertencias reales y acceso administrativo a apartados. Nombres y código en inglés, contenido UI mediante i18n existente. ReUI: operacional, Record detail, familia actual, densidad compacta, regiones barra/contenido, estados propios, sin entrada animada; anunciar ese plan antes de UI code y ejecutar self-check ReUI con máximo dos rondas. Sin dependencias nuevas.

**Verificación:** `pnpm --dir tooling/viewer/frontend test` (suite focal del paquete, no suite del repositorio) para interacciones, cambio de selección/Deshacer, restauración, nombre, búsqueda del nombre local tras recarga manteniendo hermanos visibles, y errores; reemplazar el caso obsoleto que esperaba archivar solo un hermano conservando los escenarios de selección. Comprobar sesión de tres tickets y varios recorridos del mismo ticket, editar/guardar/cancelar/restablecer, fallo de escritura sin perder texto, confirmación y fallo parcial de borrado, cuerpo de error con conteos, 409 reload_required con acción de recarga sin reintentar la escritura, foco por teclado, etiquetas y acciones móviles. AC12: nombre íntegro sin tokens divididos, región local con foco de scroll horizontal con foco visible y flechas/Tab, tanto barra como confirmación, sin overflow de página; nombre normal refluye por palabras. La inspección real del candidato Go debe medir barra estable en inicio/mitad/final, saltos visibles, sin overflow horizontal a 390 px, ambos temas y print. Usar el stack inmutable y datos aislados de la verificación compartida.

Con D10 (recuperación final) adoptada, cubrir cero supervivientes: mensaje de limpieza pendiente, Comprobar recuperación solo lee, Sync repara al restablecer almacenamiento y la UI deja de mostrar el error; no exigir reinicio ni ofrecer una recarga como si reparara almacenamiento.

## T4 — Reconciliar los contratos documentados y generar el bundle

- [x] Publicar fuentes y bundle coherentes con los controles y almacenamiento nuevos, sin documentación que siga indicando acciones por copia.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC12.

**Checkpoint:** C3.

**Dependencias:** T3 (vista y flujos) y T2 (API de sesión).

**Ubicaciones:** `README.md`, `llms.txt`, `CHANGELOG.md`, `tooling/viewer/{doc.go,frontend/README.md,static/app/}`, `_support/docs/architecture/deployment-manager.md`, `site/src/content/docs/flows/build.md`, `_support/docs/frontend/ui-conventions.md`; deltas en este registro.

**Ejecución:** hilo principal para coherencia del resultado y generación, sin editar instrucciones distribuidas ni agregar herramientas.

**Enfoque de pruebas:** `check`, comprobar contratos y bundle servido; no añadir tests que busquen texto fuente para probar comportamiento.

**Cambios:** sincronizar las descripciones afectadas, añadir nota de migración y alcance nuevo de rutas de escritura, entradas bajo Unreleased; regenerar assets incluidos según CONTRIBUTING. No tocar VERSION ni publicar sitio. Los deltas se aplican a la especificación vigente al cerrar el cambio integrado, no antes.

**Verificación:** `pnpm --dir tooling/viewer/frontend build` y `pnpm --dir tooling/viewer/frontend verify:dist`; repetir build sobre el mismo lockfile y comprobar que no genera diff adicional. Inspeccionar el bundle servido por Go: el mismo candidato debe satisfacer las observaciones AC1–AC8 (barra, acciones, feedback, conservación, eliminación, nombre, continuaciones e importación) y AC12 (lectura íntegra mediante scroll local accesible). Comprobar en el bundle final foco visible, lectura de ambos extremos por flechas y salida por Tab en encabezado y confirmación, sin overflow de página; puede reutilizar evidencia independiente renovada de T3 cuando código y assets coincidan exactamente. Buscar descripciones anteriores de per-copy controls/archive en las fuentes documentales del Impacto; corregir contradicciones, sin tomar coincidencias de texto como prueba de runtime. Verificar vínculos y comandos documentados.

## Checkpoints y verificación compartida

- C1 — T1 (estado persistente): separar su revisión porque cambia datos y sus errores se propagarían al resto.
- C2 — T2 (contrato HTTP): separar por cambio de interfaz y borrado.
- C3 — T3–T4 (vista, documentación y bundle): comprobar el resultado completo que se entregará a la persona.

En cada checkpoint, `hive-verify-task` contrasta los criterios de las tareas contra evidencia y ejecución real; no confundirlo con revisión del plan ni con el verificador final.

Suite local final según `CONTRIBUTING.md`, una vez sobre el candidato final y repitiendo solo lo afectado después de cambios:

```sh
go vet ./...
go test ./...
python3 -m unittest discover -s tests/skills -p '*_test.py'
pnpm --dir tooling/viewer/frontend check
pnpm --dir tooling/viewer/frontend lint
```

El frontend test/build/verify:dist ya ejecutado en T3/T4 no se duplica si sus entradas siguen iguales. Se toca concurrencia: `go test -race ./...` es requerido por CONTRIBUTING. Antes de lanzarlo verificar su duración reciente; si se prevé más de 30 minutos, confirmar alcance/método como gate costoso, no sustituirlo silenciosamente. No hay CI a la cual delegar checks. Una ejecución que se bloquee conserva su salida y se reporta como no verificada.

| Gate | Momento y entorno | Evidencia y autoridad |
| --- | --- | --- |
| Runtime local | Bundle Go con estado y puerto aislados | API real, persistencia/reinicio/reconstrucción y navegador; rutina de implementación, no biblioteca cotidiana |
| Primera vista | Primer layout utilizable | Mostrar barra y contenido mediante walkthrough; la entrega interactiva conserva la validación humana |
| UI completa | Candidato inmutable local antes de validación final | `hive-review-ux` y `hive-verify-change`, autorizados por D7-A (UI completa), con aislamiento y límites de `design.md` |
| Revisión de código | Tras checks/UI, antes de validación final | Codex review del diff contra la base, autorizado por D6-A (mecanismo de revisión); resolver hallazgos |
| Validación humana | Resultado revisado antes de push | Stack disponible, URL y capturas concretas; detener publicación hasta respuesta por D5-A (entrega interactiva) |
| Integración | Tras aceptación | Push/PR/merge a development y PR de cierre del registro, sin revisión dedicada del registro, por D5-A (entrega interactiva) |
| Refresh mantenedor | Después de integrar runtime/assets | Respaldo y transición; reconstruir binario, reiniciar solo visor ya activo conservando puerto/estado/TTL, verificar salud/ingesta/nuevo UI |

El walkthrough cubre sesión de varios tickets, barra durante scroll y saltos, archivo/Deshacer/restauración, nombre y recarga, conservar y eliminar con confirmación. Datos y cuentas son fixtures locales sin credenciales reales. Los verificadores usan `agent-browser`, cada uno con sesión propia y datasets distintos; no comparten mutaciones. Mantener el stack humano disponible hasta la respuesta y limpiar después solo recursos propios. Imágenes/manifest/fallback siguen el ciclo de retención de Hive.

No ejecutar pilotos de modelos. La verificación del refresh de la biblioteca cotidiana es distinta del in-vivo aislado; no realizarla antes de la integración autorizada ni borrar registros reales para probar eliminación.

## Revisión del plan y progreso

Revisión inicial fija 2026-10-27T05:05:43Z: proposal cc8eddc9, design 186f75fe, tasks 96a224b6, delta 493a49da. Tres roles nativos hive-review-plan, sin historia heredada, breves de solo lectura, perfiles efectivos declarados por el host; no se ejecutó código del plan.

- Backend/datos: tres bloqueantes aceptados: protocolo durable de borrado, protección de pestañas antiguas y criterios que no podían verificarse completos en checkpoints previos. Se definió el orden tombstones/marcador/borrado/reconciliación y se separaron AC9 (nombre en almacenamiento), AC10 (recuperación del borrado) y AC11 (guard de alcance), manteniendo criterios completos de interfaz en T3.
- Frontend/UI: sin hallazgos materiales. Quedan verificaciones reales de layout, foco y cliente contra los contratos nuevos.
- Experiencia de usuario: un bloqueante aceptado: buscar el nombre editado. Se amplió AC6 (nombre visible y buscable), el diseño y su prueba en T3.

Segunda y única ronda acotada: proposal 4acae648, design ea0f6aae, tasks fb21ec3a, delta dfe7e2c9.

- Frontend: sin hallazgos materiales en los cambios de contrato/cliente; no repitió layout sin cambios.
- Backend: resueltos protocolo durable, guard de clientes y criterios/checkpoints. Residual: limpiar también durante ejecución antes de que una continuación herede preferencias obsoletas del registro pendiente sin miembros.
- UX: búsqueda resuelta. Residual: la actualización del índice no repara una limpieza que solo ocurría al reiniciar.

La propuesta D10 (recuperación final) adopta las correcciones acotadas de ambos revisores: Sync reintenta y la ingesta espera la limpieza; el feedback distingue comprobar resultado de reparar. Se presenta al usuario como decisión después de la ronda acotada, sin lanzar una tercera revisión ni declarar cobertura automática. D10-B (recuperación y ejecución) aprobó explícitamente la corrección y todos los criterios; plan listo y ejecución iniciada. El resto del diseño sigue cubierto por las revisiones identificadas arriba.

El host no informó duración ni tokens por revisor ni re-revisión: costo no disponible. Al cerrar la revisión del plan aún no había procesos, browser, pruebas, migraciones ni implementación; el build que sigue en este historial comenzó después de D10-B. Comprobación inicial del registro: 11 criterios, todos con Closes, cuatro tareas, nombres MODIFIED existentes y sin whitespace final.

Inicio de build: registro canónico trasladado al worktree; T1 (persistencia/importación) en progreso. Los checkpoints previos se verifican antes de iniciar tareas dependientes.

Preparación local: baseline inmutable compilado desde82ccf111 en `/var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-session-toolbar-base-xdfg6zwe/hive-baseline`; servidor baseline propio PID54369, puerto62833, estado `baseline-state` bajo ese temporal. Tres recorridos sintéticos de una sesión, fuentes y estado aislados. Mantener para comparación UI y detener/retirar al cerrar. Dependencias frontend restauradas offline con pnpm12.8.1 desde lockfile sin cambiarlo:99paquetes reutilizados,0descargados; política supply-chain en caché verificada hace7d, sin scanner de advisories nuevo disponible. No se añadieron dependencias.

T1 (persistencia/importación) reportada por implementador; checkpoint C1 pendiente. Checks focales con race, paquete viewer y vet reportados verdes; verificador independiente debe rerun. Modelo implementador configurado por rol: gpt-6.1-sol medium, no observado en runtime.

C1 verificado: T1 (preferencias e importación), commit2158b833. hive-verify-task independiente: AC4/AC7/AC8/AC9 met,11pruebas focales PASS en0.951s, sin procesos propios. Evidencia session_groups_test.go:29,54,97,138,161,213,273. Perfil verificador configurado gpt-6.1-sol high, no observado; implementador configurado gpt-6.1-sol medium. T2 (API y recuperación) inicia después de esta verificación.

T2 (API/recuperación) reportada y comprometida localmente; C2 pendiente. Contrato real de error: partial_delete con deleted, remaining, cleanup_pending; índice errors incluye code=delete_cleanup_pending y session_group aun sin miembros. Frontend debe consumir esos campos. Cambios de T1 en arranque/publicación/ingesta vuelven a probarse en checkpoint por integración de recuperación.

C2 verificado: T2 (API/recuperación), commit6a2c29af. AC9/AC10/AC11 met;22tests viewer +4subtests,33CLI +38subtests (2omitidos por puerto ocupado/case-insensitive),11regresiones T1 PASS. El verificador no ejecutó la prueba general token/origin en su selección: permanece para suite final; inspeccionó orden auth. T3 inicia con contrato comprobado; su implementación se asigna a un solo responsable frontend para mantener propiedad de la vista mientras el hilo principal integra.

Preparación T4 sin editar fuentes aún: borrador reproducible de actualización documental en `_support/workspace/2026-10-26-viewer-session-toolbar/update-docs.py`, aplicar únicamente después de que T3 reporte su resultado. Captura base `images/baseline-desktop.png` guardada (1440×900, oscuro,0errores de consola/página); navegador propio cerrado.

Continuación autónoma inferida: silencio tras D10-B (implementar ahora) con D5-A (interactiva), D6-A (Codex review) y D7-A (UI full) respondidas. Fuente: regla de unattended-delegation, no duración del trabajo. Objetivo: completar este candidato en el worktree con checks y evidencia; expira al entregar el candidato para validación final o si el usuario cambia/detiene objetivo, alcance o respuestas. Primera vista se muestra sin pregunta intermedia. Se preserva la condición de D5-A: no push antes de aceptación humana final. No amplía efectos a master/release ni cambia la biblioteca cotidiana antes de integración.

T3 (fuentes frontend) reportada y comprometida localmente; C3 pendiente de T4 y evidencia renderizada. Implementador reportó check/lint y53tests frontend verdes, self-check ReUI limpio en2pasadas; un mock viejo de polling dejó un ensayo esperando, proceso detenido y escenario corregido a recarga autoritativa. No atribuir a DOM la geometría real pendiente. T4 inicia dentro de C3 tras reporte/commit de T3.

T4 (documentación/bundle) reportada, C3 pendiente de verificación y runtime. Build frontend PASS, segunda generación byte-idéntica y verify:dist PASS. Site check/lint PASS; site build en proceso propio. Paquetes site restaurados offline sin lock changes:423reutilizados,0descargas, políticas cacheadas2d; scanner nuevo no disponible. Avisos Vite: renderer CSS externo se resuelve runtime y chunkJS503.75kB; no se modificaron límites ni se añadió code splitting sin necesidad observada.

Primer render deb8e279 disponible: candidato propio PID98119, puerto50908, estado candidate-state en temporal base; sesión browser first-view-toolbar-20261026 transferida al hilo principal para su lectura acotada y limpieza. Barra observada estable y=83 de scroll0a1000,1440×900oscuro,0errores y sin overflow. Lectura propia del integrador confirmó título, canvas único y consola vacía. first_view publicado como JSON y registrado en el visor cotidiano; esto añade solo el recorrido propio, sin modificar las20copias previas. Sitio check/lint/build PASS; avisos informativos de coleccióni18n/entrada404 no investigados como defectos, sin fallo de build.

Navegador transferido first-view-toolbar-20261026 cerrado por el hilo principal después de su lectura; servidores propios baseline62833 y candidato50908 siguen disponibles. Corrección documental pendiente de T4: retirar la descripción antigua de punto/estado agregado en frontend/README y flows/build, porque el código actual muestra estados únicamente por recorrido y usa inProgress solo para apertura inicial de proyectos. Aplicar tras terminar el check T3 para no cambiar su digest durante pruebas.

T3 verificada en C3: AC1–AC6 met;53tests frontend y5Go focales PASS. Chromium real propio a1440/900/390, toolbar estable,0overflow/errores; archivo/Undo/keep/nombre/reinicio/búsqueda/borrado/continuaciones/print observados. Errores parciales/red/409 probados por tests, no inducidos en Chromium. Preferencia clara mantiene el tema oscuro existente (no hay tema claro aparte). Verificador cerró browser/servidores/temporales propios. T4 reabierta para corregir dos frases documentales obsoletas sobre el estado agregado; solo READMEfrontend afecta sourceDigest, no JS/CSS de runtime.

Corrección T4 documental comprometida: descripciones del estado por recorrido alineadas; bundle regenerado/determinista y verify:dist PASS. JS/CSS de runtime idénticos a deb8e279; solo cambió sourceDigest del manifest y los dos documentos. T3 conserva evidencia relevante porque su código, estilos y API no cambiaron. C3 pendiente únicamente de verificación T4.

C3 cerrado: T4 met AC1–AC8, build/verify:dist y segunda generación determinista PASS,8Go focales PASS,100destinosdocumentaleslocales válidos y artefactos HTML/JS/CSS servidos idénticos a manifest. Reutilizó explícitamente evidencia independiente T3 sin repetir UX. Current preview propio PID19556 reemplaza PID98119 intencionalmente; mismo puerto50908/estado/TTL, binario58ff5570. No source code changes pendientes, solo registro no versionado. T1–T4 verificadas; entra flow-verify nivel completo con UI full, suite local y Codex review, luego validación humana antes de push.

Gate UX de58ff5570 no pasó: Blocker introducido en confirmación de eliminación con nombre válido largo sin espacios (contenido1385px en popup320px, botones fuera); High introducido por h1 line-clamp-2 que recorta nombre principal (168pxrequeridos/56visibles a320). Fuente application.tsx:787 y shell.tsx:703; imágenes ux-delete-320.png/ux-long-320.png. Se reabren T3 (layout de nombres) y T4 (bundle dependiente) para corrección local de reflujo/intrinsic sizing sin cambiar componentes instalados ni tema. UIreview quedó incompleta para estados/accesibilidad y debe completarse tras fix. Funcional sigue en candidato fijo58ff5570 y su evidencia se reconciliará por diff; las grabaciones de toolbar/session-actions serán stale si afectan los archivos cambiados y se renovarán en el gate correspondiente. No se declara apto ni se publica.

Fix UX T3 comprometido7d500218: nombre de h1 sin clamp y textos con min-w-0/overflow-wrap:anywhere; confirmación reduce min-content. Solo clases de composición, sin cambios de API/lógica/tokens/componentes instalados; no theme deviation.53tests/check/lint/self-check ReUI PASS, geometría real pendiente de repetir. T4 bundle regenerado/determinista en segundo build; se renuevan evidencias de vistas afectadas y C3 sobre nuevo candidato.

Primera pasada funcional58ff5570 PASS en servidorGo/browserreal: archivo/Undo/cambio selección, herencia,keep,nombreUnicode/reinicio/reconstrucción/búsqueda,foco,delete y nuevasversiones,guardheaders/token/origin,redreal;27Go+4subtests PASS parafallosbackend. Sus2MP4H264 válidos16.87s/17.13s y capturas preservados con sufijo-58ff5570 y marcadosstale antefixvisual. Se refrescan solo vistas/journeysafectados en39187977. UX391 ahoraPID54390/53658;preview se reinició con391mismo50908/estado/TTL. RuntimebinarySHA54a8d4e8252e17877eebee9ec68149ec18da3d8193222a204d5f86beaf8db579;treef98a352570951eac31d4c3fa5be03020312aab8d.

Nueva revisión durante verificación: UX391 corrigió Blocker y clamp, pero mantiene High por dividir el token sin espacios. Ajuste autónomo propuesto: región horizontal local accesible con AC12 y pruebas de foco/semántica + geometría real. Se completa revisión acotada frontend/UX antes de editar; no requiere backend porque identidad/nombre persistido/API/límites no cambian. Verificador frescoT3 encontró AC1–AC6 met en391 pero no cubre AC12; se conserva esa evidencia donde el nuevo diff no la invalide. Su tree417758f difiere def98a352 solo por registro no versionado (reconciliar por diff antes de cierre); SHA de binario coincidió. La pasada funcional391 tambiénPASS y renovó MP4; el siguiente ajuste volverá stale las vistas/grabaciones afectadas y exige renovación posterior.

Revisión del ajuste AC12 completada por frontend/UX nativos hive-review-plan sobre12eef2cb/f6204ef9/a860b41e/0cb0b14f. Ambos encontraron solo la misma omisión de AC12 en Verificación T4; aplicada su corrección propuesta literalmente (lectura de ambos extremos/foco/flechas/Tab en bundlefinal con evidencia T3idéntica reutilizable), sin otra ronda conformeplan-review. No otra brecha material del patrón. Layout/desplazamientonativo/discoverability siguen sujetos a runtime/UX. Duración/tokens de estos dos revisores no informados porhost. Ajuste listo para implementación, dentro de autoridad autónoma; no se atribuye elección explícita al usuario.

Reconciliación tree del verificador391: gitdiff f98a352..417758f solo contiene los4archivos del registro no versionado; ningún código/asset cambió. Su SHA del binario coincidió. La evidencia AC1–AC6 previa no se presenta como cobertura del AC12 posterior.

AC12 implementado en114952cd (T3): h1 enfocable, región semántica enfocables de nombre en confirmación, foco visible/ayuda i18n, nowrap para token sin espacios y reflujo normal para nombres con espacios; Cancel inicial explícito. RED focusheader y54tests/check/lint/selfcheckReUI PASS. Bundle/README empaquetados, secondbuild determinista; pendiente runtimeflechas/extremos/Tab/overflow y UXrestante. T3/T4 reportadas[?] hasta cerrar el checkpoint ampliado.

Corrección del gate UX sobre217baaeb: un PUT de nombre sin respuesta deja el diálogo deshabilitado sin progreso. Se concreta el manejo de error ya exigido por AC6/T3: estado de guardado anunciado, espera máxima15s de esa escritura, recuperación de controles y entrada intacta; refrescar autoridad sin reenvío automático ni afirmar cancelación del servidor. T3 vuelve a implementación y T4 espera nuevo bundle. AC12 sigue pendiente por bloqueo de agent-browser después de muchas flechas (EAGAIN); Chrome DevTools MCP ausente e IAB no disponible. Se asigna reproducción independiente a Cua/Edge en pestaña propia1744527450 con estado sintéticoUX53658, preservando pestañas ajenas. No cambia API, datos, criterio ni alcance; detalle de recuperación de error existente, no nueva revisión de plan.

Hallazgo global UI preexistente P3: regiones/índices repetidos en recorridos apilados (axe6nodos391/7base en oscuro; no conteo217). Registrado una vez sin acción en `_support/docs/frontend/ui-conventions.md`, según ui-gate; búsqueda Engram/GitHub landmark sin duplicados. No implica WCAGAA incumplida ni bloqueo observado. Guía conserva baseline y marca toolbar candidato pendiente de aceptación; navegación desde AGENTS fuera del alcance, no editada.

Reproducción independiente AC12 mediante Cua/Edge PASS en217baaeb/53658 a320CSS: nombre200puntos START+192W+END; barra scrollRight3063.89/max3064 y retorno0, confirmación2356.67/max2356 y retorno0, flechas nativas, Tab sale aEditar/Cancelar, outline sólido y0overflow página/diálogo. Nombre natural146chars refluye5líneasbarra/4confirmación sin overflow. Evidencia `images/cua-{name,delete-name}-{left,right}-320.png` y natural. No se usó scroll programático. Nueva pestaña propia1744527477 (anterior expiró), marcada continuidad, viewport/emulación restaurada, nombrefixture restaurado; token sintético solo sessionStorage. No procesos nuevos. Reutilizable solo tras comparar código de lectura de nombre con candidato final; no verifica recuperaciónPUT concurrente.

Corrección de guardado comprometida7fa8fd44(T3), bundle/guía051d45d9(T4). RED tiempoagotado y progreso ausente;56testsfrontend/check/lint/selfcheckReUI PASS reportados. Build repetido byteidéntico yverify:dist PASS propios. Implementador no abrió procesos, inspección runtime se coordina con nuevo candidato. T3/T4 reportadas[?] paracheckpoint; AC12 fuente shell y región deconfirmación sin cambios respecto217, pendiente reconciliación final independiente.

UX051d45d9 noaprobado: progreso/timeout/entrada/realSave correctos, peroHigh resultado deComprobar ocultodetrásmodal yHigh pérdida de focobody aldeshabilitarSave. Se reabreT3 paramostrar nombreactual comprobado/fallolectura dentro diálogo y mantenerfoco visible duranteespera/recuperación. Ambos corrigen mismo contrato deerror/foco existente, sin API/datos/criteriosnuevos; T4esperabundle. No se repiteámbito noafectado niAC12. Queda matrizcontraste/foco ycomparaciónregionesdesplazadas, se completará siniteraciónindefinida.

Funcional051 confirmó timeout por suspensión temporal de su servidorGo aislado, sinmock:1PUT duranteespera,15srecupera yconservaborrador, servidorreanuda yguardarealmente, comprobar soloGET, segundoPUT solo trasGuardarexplícito. Fuente backend yregionesAC12 iguales porgitdiff217..051; evidenciaanteriorreutilizable. Capturasfunctional-final-name-{pending,timeout}.png. ServidorPID75082/49870 detenido/browserpropio cerrado. PestañaCua1744527477 cerrada porpadre. No grabaciones051 porque correcciónUXharía stale.

CheckpointT3 sobre051: hive-verify-task fresco AC1–AC5/AC12met, AC6notmet porresultadoComprobar noexpuestoendiálogo;56frontend+6Gofocal PASS, barra1440y53alto88 estableen0/4154/8308,390sinoverflow/accionesaccesibles, impresión/AC12reconciliados. Pérdidafocono reproducida porverificador, síUX. FilaT3[/] permaneceporfixyaencurso; esterechazo cuenta ronda1 delmismoflujorecuperación. Próxima revisiónT3soloAC6notmet (AC12evidenciaseconserva si fuenteigual); T4todoscriteriostrasdependencia. Perfilhive-verify-task Solhigh configurado/noobservado; sinprocesospropios.

MatrizUX restante051 completada: seisviewports/ambaspreferencias mismooscuroefectivo,0overflow, regionesdesplazadasmismogeometríabase, finalcontenido visible,4acciones28×28/nombresTaborden, íconos18.97:1/secundario7.66:1. DosHighglobalespreexistentes=>P2 agrupadoscontraste: foco botones1.88:1 (basecopiar igual), borde/rellenobuscador1.28:1/1.06:1, SC1.4.11AAsegúnrevisor; sinempeorar. Cambiarlosseríadesviacióntema,noadoptada. Engramcontrastfocus yGitHubcontrast sincoincidencias. Propuestaagrupadaal cierre; guíaUIregistraestadopendientedisposición yvalorestemas. Capturasux-matrix-{focus-320,search-1280}.png. No son cambiosdeestaentrega; faltaUXsolocorrecciónguardado.

FixT3recovery275d8c0e: nombreactualconsultadoendiálogo/estadodeconsulta/errorlectura, borradorintacto/focoguardadoestadoytimeoutcampo. RED→GREEN deambosHigh,56tests/check/lint/selfcheckPASSreportados. Bundlef11e0160 dosbuildbyteidénticosverify:distPASS. Se revalidanAC6yUXrecovery, T4todoscriterios; regionesAC12barra/delete sinmodificar, nuevaregiónlecturasalvado aplicamismopatrónlocal.

T3checkpointrecoveryronda2f11e0160: AC6met porhive-verify-taskfresco,3casosfocalesPASS16.03s; inspeccionaasserçõesvalorGETcambiante/error500/borrador/1PUT/focostatus→campo. SHAcoincide; persistencia/búsqueda/identidad/backend sin cambios. AC1–AC5/AC12 de051met se mantienenporreconciliacióndiff; T3[x]. Primerlanzamientoverificador dejósalidaparcial,repite3casos sinprocesospendientes. AspectovisualdefocoquedaUIgate,noatribuidoatestDOM. T4iniciareverificacióntodoscriterios.

UI/funcionalf11: ambosHighrecoveryresueltos, realGoPUTdemorado porSIGSTOPpropio, focostatusoutline2px→input15s, consultamuestravalorGETendiálogo/borradorintacto,errorlecturavisible,1PUThastaguardarexplícito, consolaerrores0. UX320×844textspacing diálogo733px botonesvisibles, nuevaregiónsalvadonombretieneayuda/foco/flecha40px/Tab; gatependientesoloextremofinalpornativokeyboarddenuevaregión. AsignadoCuaindependientefuncional. Grabacionesf11H264limpias41.733333s/23.366667s, padrependienteffprobe; nuevascapturasfunctional-ready-*. SesiónAGBfunctional-toolbar-ready/namespacepredeterminadotransferida, serverPID216/51707/exe91723 statefunctional-state-ready TTL336h semantieneparalecturapadre mientraschildcompletaregión. No procesosajenosmodificados.

T4f11metAC1–AC8/AC12: verificadorfresco2buildbyteidénticos/verify:dist, Go8focalPASS1.009s, HTTPbytesHTML/JS/CSS/3recursoscompartidos coinciden, SHAcorrecta,102links/15fragmentosválidos. Nitdocubicación«Deletingfromsidebar» corregido mecánicamente abarra/sesión; no fuentefrontend niartifact/runtimealterados, checks/docsrestantesválidos. T1–T4[x]. ÚltimaregiónrecoveryCuaPASSf11:320CSS, flechas2356.67/max2356→0,1línea20px,outline/TabaComprobar,0overflow; capturasfunctional-recovery-reading-{left,right}.png. UXcombinadocierraal resolverúnicacomprobaciónpendiente; globalesP2separados. ParentffprobeambosMP4H264valid41.733/23.367s. Suitefinaliniciada.

Suitefinalea64359d(códigof11idéntico): govetPASS0.47s, gotest./...PASS110.08s; Python64testsPASS7.57s; frontendtipos/lintPASS2.22/0.45s, sitecheck/lint/buildPASS4.15/0.48/1.87s. Raceenproceso. LecturapropiaAGBrealfunctionalhandoff: h1correcto/acciones4/0overflow/consolavacía; browsercerrado/sesionlistvacía, server51707detenido. Previewhumanoreiniciado conbundlef11mismo50908/state/TTL, PID31064/exe63539. Base62833 yUX53658aúnpropios. T4nitdoccommit ea64359d noaltera código/assets; evidenciaf11válida.

Higiene previaavalidación: base62833/UX53658detenidos porpadre intencionalmente (ya no necesarios); navegadorespropioscerrados. Solo preview50908PID31064/exe63539 permaneceparavalidación, datos/binaries/evidenciapreservados. Race4m54s aúnCLIactivoCPU38%, managementyarestoterminados; noindicioatrasco, timeoutnativo10min vigila. ReglasAGENTSactualizadasporusuarioaplicadas sin cambiarobjetivo/entrega.

Racefinal./...PASS304.95s; suitefinalcompletasinfallos. Codexreviewnativoiniciadosobreea64359dcontra82ccf111, logprivadotemporal. No CIenrepo. Revisiónaúnpendienteantesdevalidaciónhumana.

Codexreviewnativoea64359dcontra82ccf111terminóexit0,1P2introducido: cleanupremaining0deAusadoaleliminarBdespuéscancelarA, branchrefreshbloqueaB. Confirmacióndesourceapplication.tsx, T3reabiertoAC5paraaislarrecoveryporoperation/groupId enbranch/label/error yregresión2sesiones. SinP0/P1,nosegundarevisiónporfixP2 segúnreview-gates; se repitefrontend/bundle/Go viewer yUIafectados,nofullGo/racequenocambia. T4reabiertobundledependiente. Fetchread-onlyorigin: development95077cd4avanzósoloenguías/registro/changelog/páginasnoUI; candidatecodebaseseconserva, integraciónactualizarábaseychecksantesmergecomoD5autoriza.

FixP2b9172f99(T3): deleteError filtradooperation/groupIdgobiernaerror/conteo/label/disabled/onClick/autocierre, regresiónAcleanup0→cancel→BeliminableRED→GREEN.57tests/check/lint/selfcheckPASSreportados. BundlecommitnuevoT4dosbuildbyteidénticos/verify:distPASS; pendientecheckpointAC5yT4/UIaffected. Fullsuiteea previaretieneámbitosnocambiados; frontendposterior57tests/probarGo viewer assets. Grabaciónsession-actionsstale,sepreservaf11yrenueva; toolbar-scrollnoconsumelógicadeletedialogpuedemantenerseevidenciaporreconciliación.

AC5reverificado008dc4c9porhive-verify-taskfresco:8frontend+3GoPASS; filtrooperation/groupIdgoverna todoslosconsumidoresdedialogo, regresiónAcleanup0→BDELETEdistinto. SHAcoincide; AC5met,T3[x]conrestocriteriosconservadosporfuenteigual. Go viewerpaqueteposteriorfixPASS13.092spropio. T4esperareverificaciónartifacts/UIafectada.

T4verificador008mettodoscriterios:2build/verify:dist/HTTPbytes/lockfrozen/SHAcorrectos. UX008 aislamientoA/BPASS(0vs2miembros,label/focoCancelarB/320y1280caben), pero observófocobodytraspartialA, atribucióndiffsinresolver. Alserdiálogonuevoenelcambiocompleto,seadoptacorrecciónmínimadelfocodedialogoduranteytrasDELETE, patrónyaestablecidonameflow, sinlayoutnuevo. T3[/]yT4[?]reabiertasámbitofoco; no se amplíaotrasvistas/matrices. Grabación008serástalesisegeneró.

Funcional008scopePASS: A0cleanup→cancel→B1/Eliminar sinerrorajeno, DELETEBrealGo204/índice0. Fallo500/errorpersistenteA simuladosporfetchinterceptorpreviofrontend; eliminacionesfísicasAyBatravésGoreal,noaveríadurableinducida. Capturasfunctional-delete-scope-{pending-a,confirm-b}.png. Video008H26441.1333s preservadocomostaleparaúltimofixfoco; toolbar-scrollf11siguerelevante. Browser/interceptadorretiradosyserver73191detenido.

58c75e02(T3)focowriteDELETEalconteovisible/tabIndex-1/outline, falloCancelar, successfinalFocusprevio; regresiónrequestdiferidoRED→GREEN.57tests/check/lint/selfcheckPASS. BundleT4doblegeneraciónbyteidéntica/verify:distPASS; checkpoint/UIacotadosfocopendiente.

AC5focuscheckpoint3cc3658fmet: verificadorfresco2componenttestsPASS, focoinsidedialogenabledpending/failureCancelarorecovery/successdocument, finalFocusunchanged. SHAcoincide; T3[x]todoresto conservadopordiff. Go viewerfinalPASS12.975s. Previewfinal3cc50908PID94420/exe62386, mismoestadoTTL. T4reverificaciónartifactpending.

UIgate3cc3658fscopedPASS: UX320/1280countfocusoutlinepending→Cancelarerror→bibliotecasuccess, sin nuevosHigh/Blocker; funcionalDELETErealpendientepordemora yDELETEA/Breales204, errores500/cleanupAsimuladosclaramente; focoenabled/recuperaciónAislamientoPASS, consolaerrores0. Video3ccH26425.033333srenovado(parentffprobePASS), toolbar-scrollf11vigentepordiff. Servidor93594/58348detenido/browserscerrados. RevisiónnativaCodexfinalacotadaea..3cciniciadaporúltimoarreglodeUXposteriorreview; nootrarevisiónfull,norepetirglobalchecks.

T4final3cc3658fverificado:2build idénticos a4archivoscomprometidos/lockintacto, verify:dist, HTTP200MIME+SHAHTML/JS/CSS/3recursos correctos, binarioSHAcoincide. ReutilizacriteriosAC1–AC8/AC12porreconciliacióndiff ycheckpoint/UXfuncionalfinalfoco. Un404derutaJSinicialequivocada corregido aprefijodocumentado, sinfallodeproducto. Sinprocesos/cambiospropios. T1–T4[x]; pendienteúnicamenteterminaciónreviewacotadayentregahumana.

## Historial: entrega para validación humana (aceptada)

- Candidato final: `3cc3658f`, rama `feat/viewer-session-toolbar`. T1–T4 verificadas; los 12 criterios están cubiertos.
- UI completa: revisión UX y verificación funcional aprobadas, con evidencia conservada y renovada solamente donde cambió el código.
- Suite general: Go vet, tests y race; 64 pruebas Python; tipos/lint y 57 pruebas frontend; sitio check/lint/build. Las correcciones posteriores repitieron el ámbito afectado y el bundle, sin repetir la suite global sin necesidad.
- Codex: revisión completa de `ea64359d` contra `82ccf111`, un P2 introducido corregido; revisión acotada final `ea64359d..3cc3658f` sin regresiones accionables (25 pruebas de componentes y verificación de bundle aprobadas por el revisor).
- Recorrido humano: `walkthrough.json`, capturas y dos MP4 H.264 de 41.733 y 25.033 segundos. El primero conserva su candidato f11e0160 porque barra y desplazamiento no cambiaron.
- Solo queda activo el preview propio en 50908, PID94420, sesión de proceso62386, `candidate-state`, TTL336h; binario `hive-candidate-delete-focus`, SHA256 ba802830edc7bb9c52572e55dbba595b10eac8aeb3733856c67c655462ce2d28. Base, UX y servidores/navegadores de los hijos están cerrados.
- La autonomía inferida termina al entregar el candidato para aceptación humana. No push antes de esa respuesta. La aceptación habilita la entrega interactiva ya elegida; la base remota avanzó a95077cd4 solo en guías y documentación, por lo que debe actualizarse la rama y ejecutarse los checks de integración antes de merge.
- Temas aparte: contraste global preexistente de foco/buscador, P2 agrupado pendiente de disposición al cierre; nombres repetidos de regiones, P3 registrado sin acción en la guía UI. No se cambiaron tokens ni componentes de tema.

Integración confirmada: PR227 merged developmenteced053f, head28c54fc4 contiene base95077cd4; CIinexistente, CodeRabbitopcionalSKIPPEDnoasumidoreview. Fullvet/go/PythonrepetidossobrebaseactualPASS(Go109.66s/Python5.75s); sitecheck/lint/buildPASS. D11-Aaprobóvisualantespush. Refreshautorizadopormarcadormantenedor: targetbinariohiveexistente yvisor4483PID72416, estadoApplicationSupport/tricell-hive,TTL336h,21copias/6raíces. Revert: binarioanterior ybacku pcoherentemetadata/index/tombstones/grupos antesprimerinicio; preservar cambiosposteriores. Backupprivado diario en temporal daily-viewer-backup-eced053f,modos700/600; server.json/tokensnocopiados. Trasstopsnapshotcoherente yhashesrecursosantes/después. Nuevoobjetivo capturasestados entregaseparadaworktreecapture-only-walkthrough, decisión16565.

Refresh diario verificado: PID34140,4483,TTL336h,21 copias/17 grupos verificados,172recursos sin cambios; registry/flagsnormalizados/timestamps/tombstones preservados; HTML/assets coincidenconeced053f. Un404 inicialporprefijoasset de verificaciónincorrecto se corrigió leyendo las rutas del HTML, sin fallo del producto. Actualización no publicórelease/master/sitio. Cierre aplica7bloques dedelta actuales por nombre y compara texto normalizadoparaigualdad. Capturas-onlyseentrega separado según solicitud posterior; mediospreviospreservados.

## Cierre verificado

El usuario aprobó D11-A (resultado visual). PR #227 (barra fija y acciones de toda la sesión) quedó integrado en development en eced053f; la biblioteca cotidiana fue respaldada y migrada conservando sus 21 copias, 17 grupos, 172 archivos de recursos, registro, tiempos y tombstones. El visor sigue activo en4483 conTTL336h yPID34140. Los siete bloques de requisitos fueron aplicados y comparados con el delta. Este registro se publica una sola vez mediante el PR de cierre autorizado, sin revisión dedicada. El ajuste de capturas de estados continúa en su propio worktree; los hallazgos globales de contraste esperan disposición del usuario al cierre conjunto.
