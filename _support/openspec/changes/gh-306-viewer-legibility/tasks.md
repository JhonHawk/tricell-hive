# Tasks

## Current extension checkpoint

T1–T6 verified the recorded provisional-v2 baseline. The authorized canonical-reference extension is specified in [canonical-references.md](canonical-references.md): final protocol/schema/storage3, no older-format reader/migration. T9 establishes the contract; T10 backend/CLI and T11 UI run independently after it. T12 integrates their evidence. T7 documentation waits for the new authoritative behavior; T8 verifies AC1–AC35. Earlier checkbox evidence is retained as baseline history and does not claim the final contract works yet.

Base development 044664a0; checkout refactor/flow-ceremony-2 en 2f0a133b. 2026-11-01. Esta ampliación reemplaza la descomposición anterior; ninguna tarea empezada al reconciliar. T1–T5 conservan IDs con alcance actualizado; T6–T8 son nuevas.

Orden contrato → UI tipada → backend → publicación/corte. T2/T3 comparten copy y son secuenciales. Desde T1 verificado, T2 y T4 van en paralelo; la preferencia UI-antes-backend no es dependencia real. T4 escribe solo Go y T2 posee el cliente v2, evitando solapamiento. La primera vista T3 conserva su validación, pero no detiene T4/T5/T6 independientes; T8 espera ambas ramas. Hijos con contexto limpio, sin commit, tracker, publicación externa ni más hijos. Principal integra/verifica. Antes de build comprobar cambios de la sesión paralela.

### T1 — Contrato v2, colaboración, estados y documentos
- [x] Go/TS/schemas/fixtures fijan operaciones, errores y límites antes de UI. · 72aa12fd · hive-verify-task aprobó forma T1 tras H6–H12; runtime18viewer/2CLI sigue pendiente T4/T5/T6, no PASS.

**Closes:** AC1 (forma), AC11–AC15 (forma), AC17 (forma), AC18 (actor), AC19 (forma), AC20/AC22–AC27 (forma)
**Depends on:** none
**Locations:** tooling/viewer/timeline_contract.go, timeline_protocol.go, frontend/src/timeline-contract.ts, testdata/timeline_contract/, tooling/cli/view_timeline.go y pruebas.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** unión actor native_producer/local_viewer y matriz de permisos incluyendo las cuatro acciones del navegador; matriz exacta de authority por operación, join request/accept/transfer y recover_lead local, producer_id en mutaciones, campos server-only, stage targets, task registry global con segmentos por etapa y agregación, delivery/wait CI tipados y matriz independiente con bloqueo estructurado/actor/asignación/task evidence targets, state/reason/audit, document_key/predecesor y protocolo 2; rechazo explícito de entradas v1 sin adaptador compatible. Precisar límites y reserva de ciclo de vida; task_version se incrementa atómicamente por cualquier cambio de snapshot asociado y CI mantiene mapa applicable_attempt_id por check con generaciones explícitas.
**Verification:** fixtures compartidos Go/TS válidos/invalidos: cross-work/project, pending producer, enum/targets mezclados, claves desconocidas y campos de servidor en entrada, escrituras v1. Contratos/decodificación/paridad Go+TS y compilación de viewer/CLI deben pasar en T1; reconstruir bundle y verify:dist. Ejecutar suite completa para clasificar fallos, sin adelantar implementación backend: los que dependan exclusivamente de inicialización/enforcement v2 aún no construidos se conservan y se registran para T4/T5, no se presentan como PASS ni se eliminan para aprobar T1. Suite runtime completa obligatoria tras T4/T5 y en T8. hive-verify-task independiente comprueba los criterios de forma cerrados por T1, no afirma backend operativo.

### T2 — Historial ascendente, estados y participantes
- [x] UI tipada ascendente, tareas/tramos y participantes · 35cf9d40 · checks frontend/bundle PASS; primera vista integrada pendiente T3. Se retiraron pruebas exclusivas legacy, conservando seguridad/standalone; recuperación bloqueada hasta relectura válida con regresión 5/5.

**Closes:** AC2 (cliente v2), AC3, AC7–AC9, AC13 (UI), AC14 (lectura), AC16, AC18 (UI), AC19 (UI), AC20–AC21, AC24–AC27 (UI)
**Depends on:** T1
**Locations:** frontend/src/timeline-content.tsx, timeline-header.tsx, timeline-copy.ts, api.ts/application.tsx/controller.ts/shell.tsx donde corresponda, agrupación pura, fixtures/tests y static/app. T2 posee también el retiro de ramas legacy/index frontend de AC2 para evitar solapar T4.
**Execution:** hive-build-frontend
**Test approach:** tdd para orden/captions/scroll; characterization para render
**Changes:** UI de design.md, eventos previos primero y cierre al final, grupos ascendentes; Popover Sesiones con lista y recuperación D72. Ir al actual y scroll por trabajo/deep link. Sin salto por polling ni controles falsos de ejecución. Filtrar proyecciones task/segment por etapa, global task_id único y conteos por tramo; etapa Entrega/nodo espera CI con resultados/links. Tareas con punto de estado explícito, Spinner/Tooltip y Collapsible; título abre Sheet. No usar activeStep para inferir completadas ni duplicar evidencia en etapa y tarea; contador real sin porcentaje. Popover/Dialog de recuperación D72: solicitante explícito, motivo, confirmar, carga/error/conflicto y sin autoenvío.
**Verification:** sesión con retiro, reentrada, continuación de etapa vieja, revisión paralela, espera/cierre/sin foreground; captions distintos/iguales/vacíos; A/B scroll y append sin salto; CAM: T1–T6 Build y T7 Verify+Delivery, Build6/6, Verify1/1, Delivery0/1 y global6/7 sin nodo futuro falso; cinco estados de tramo, pendientes visibles, finalizadas no contiguas, dos expansiones independientes, lista de 30 tareas con texto realista y reduced motion; renombrar/archivar/fijar/eliminar conservan actor local y confirmación de borrado. pnpm check/lint/test/build/verify:dist desde frontend; go test ./tooling/viewer -run 'Test.*Bundle'; self-check ReUI.

### T3 — Sheet común y cadenas de revisión
- [x] Sheet común documentos/tareas/CI · b9579c06 · main typecheck/lint339tests+3scripts/verifydist PASS; recorrido fixtures1280/440 claro/oscuro,320/390 sin overflow; aceptación primera vista pendiente D76.

**Closes:** AC4–AC6, AC17 (UI), AC21–AC23, AC25–AC27 (Sheet)
**Depends on:** T2
**Locations:** frontend/src/timeline-documents.tsx, timeline-copy.ts, tests/fixtures y static/app.
**Execution:** hive-build-frontend
**Test approach:** tdd para cadenas/avisos; characterization para render
**Changes:** Sheet del diseño, un cierre; selector documental si varias cadenas, pestañas por label, Publicado único; deltas agrupados solo en plan. Snapshot no cambia solo. Sheet de tarea con tramos/global y Resumen/Evidencia/Historial; Sheet CI con Resumen/Checks/Intentos y lectura sin efectos remotos, bloqueo primero, retorno desde documento adjunto sin overlays anidados; apertura por título o Ver cómo desbloquear, preservando foco y expansiones.
**Verification:** cinco etapas, dos claves con tabs iguales sin mezcla, inicial/cambiada/nueva/quitada, tab desaparecida, tarea con/sin imágenes, bloqueo con pasos/responsable/criterio, apertura de adjunto y regreso sin perder contexto, snapshot task/wait con aviso acotado y actualización manual; añadir captura/revisión/resultado CI a tarea sin cambiar estado dispara aviso, otra tarea no; estado CI success/failure/cancelled/unknown y distinta política/candidato sin reutilizar éxito, carga/error/retry/vacío/inaccesible, aviso solo por sucesor. Comandos frontend/bundle y ReUI de T2. Mostrar primera vista de T2/T3 al usuario con fixtures antes del backend, contra los bocetos.

### T4 — Runtime v2 con corte limpio
- [x] Corte/storage v2 verificados · 71069507 · hive-verify-task aprobó AC2/AC15 en alcance T4, H13/H14 corregidos; lifecycle pendiente T5. Se retiraron siete suites exclusivas scanner/copia/grupos legacy; guardas reutilizadas conservadas.

**Closes:** AC2, AC15 (persistencia/cutoff)
**Depends on:** T1
**Locations:** tooling/viewer/timeline_store.go, registry/index/server/scanner y pruebas Go; tooling/cli/view_command.go y state-dir validation. El cliente/index/controller frontend lo posee T2; T4 no escribe frontend ni static/app.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** storage 2 exclusivamente; detectar estado previo antes de cualquier reconciliación/escritura; index/frontend solo API v2; desactivar ingestión y render legacy de la ruta v2; no importar preferencias/manifiestos. No migración ni parser v1. Preparar directorio candidato nuevo, no tocar habitual. Eliminar tests exclusivos del comportamiento retirado con razón en checkpoint; preservar guardas de archivos, autorización y seguridad reutilizadas.
**Verification:** inicio vacío v2, index sin legacy, registro/upload/publicación v2, rechazo cutoff_required sobre copia de estado previo con manifest de hashes/mtime idéntico antes/después; corrupto/desconocido falla sin limpiar. Restart/crash/receipts/outcome_unknown v2. go test -race ./tooling/viewer ./tooling/cli. hive-verify-task independiente.

### T5 — Servidor colaborativo y lifecycle
- [x] Colaboración/lifecycle/tareas/CI servidor verificados · fa7d4b81 · hive-verify-task aprobó T5 tras H15/H16; main domain-race9.398s PASS, TSparidad+frontend345PASS. CLI/end-to-end pendiente T6.

**Closes:** AC1, AC11–AC14, AC17 (servidor), AC18 (servidor), AC19 (servidor), AC20/AC22–AC27 (servidor)
**Depends on:** T4
**Locations:** tooling/viewer/timeline_store.go, timeline_server.go, timeline_blobs.go, contrato/pruebas.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** join request/accept, conductor único/transfer, permisos, asociación explícita/fallback, estados/reason, auditoría durable y document chains con predecesor. Task registry global estable con segmentos/asignación/order, agregación determinista y binding de ocurrencias; etapa delivery y wait CI versionado con snapshots/intentos/required checks, transiciones y reopen explícito con auditoría; bloqueo exige payload completo, adjuntos y decisiones scoped por tarea, retirada preserva historia. Reservas de capacidad para cierre.
**Verification:** segmentos no cambian foreground/otros tramos; T7 Verifycompleted/Deliverypending no completa task global; crear espera en etapa no delivery falla; CI commit/política incorrecto, requisitos desconocidos y permisos insuficientes nunca success; rerun mismo candidato conserva nodo/historia, nuevo commit crea espera enlazada; rerun waiting/failed seguido por éxito atrasado del intento anterior usando versión vigente no cambia el resultado actual, también tras restart; observación desconocida no se autopromueve; contenido enlazado cambia task_version una vez en el mismo commit y retry exacto/cambio ajeno no lo hace; task transitions no cambian foreground ni otras tareas; registro/order/source key no duplicados, finalizada→en curso conserva evidencia; bloqueo incompleto rechazado, resolución auditada, task limits y cierre a capacidad; task cross-work/stage/actor rechazado; revisión concurrente no roba foreground; pendiente/ajeno/cross-project rechazado; no editar contenido ajeno; dos transfers concurrentes; recuperación con conductor ausente y dos recover_lead simultáneos, fuente/estado/evidencia intactos, old lead sin privilegios y retry tras restart; transiciones espera/respuesta/reanudación/pausa/fallo/cierre/reentrada; retry tras restart y stale/payload cambiado sin efecto. go test -race ./tooling/viewer ./tooling/cli; hive-verify-task independiente.

### T6 — CLI v2 y recorrido real de publicación
- [x] CLI baseline v2 y recorrido aislado verificados · 06c72bfa · hive-verify-task aprobó T6; main raceSharedTasksCIAndRestart14.355sPASS. Nuevas referencias canónicas se revalidan en T9–T12, no cubiertas por este baseline.

**Closes:** AC11–AC15 (end-to-end), AC17 (CLI), AC19 (recorrido), AC20/AC22–AC27 (CLI)
**Depends on:** T5
**Locations:** tooling/cli/view_timeline.go, view_command.go, ayuda, pruebas CLI y contrato offline.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** JSON/usage del contrato, diagnóstico de actualización sin secretos; no arranque/reinicio automático.
**Verification:** crear A, solicitar B, aceptar A, revisión B, transfer A→B, continuar plan/build, preguntar/reanudar, leer documentos de dos cadenas, reiniciar servicio aislado y comprobar hashes/recibos. CLI vieja contra nuevo y nueva contra viejo fallan antes de escribir cuando incompatibles. Recorrer desde navegador renombrar/archivar/fijar/eliminar con actor local y recuperar sesión A ausente hacia B mediante Dialog, con conflicto de versión y auditoría persistente. Registrar plan completo con fases explícitas, vincular segmentos al iniciar etapas, T7 Verify/Delivery global6/7 y cierre7/7; publicar espera CI e iniciar/promover intentos aplicables explícitos por check, publicar éxito atrasado anterior sin recuperar verde actual, con snapshots sintéticos de proveedor (sin llamadas pagadas ni reruns remotos), falta permisos/requiredset desconocido/fallo/éxito verificado para commit correcto; registrar tarea sin fase no adivina título. Progresarlas con productores asignados, bloquear/desbloquear/reabrir y leer auditoría/evidencia tras restart, verificando no duplicados por retry. Sin modelos ni fixtures en Engram; sin importar sesiones previas. hive-verify-task independiente del contrato CLI.

### T7 — Guía y documentación coherentes
- [x] Guía/manual/spec coherentes verificados · 2c0fc28a. Publisher~11KB, links/currentcontractcopy exactos; contenttests/sitecheck/build/lint PASS, manual1280/440 sin overflow, UXmanual aprobado; integraciónfinalT12/T8 aún pendiente.

**Closes:** AC10, AC20/AC24–AC27 (publicación), AC28–AC33 (publicación canónica)
**Depends on:** T6, T10, T11, T13, T14
**Locations:** content/skills/flow-walkthrough/SKILL.md y references/viewer-publication.md, referencia identidad si corresponde; consumidores global/flows/roles; site, arquitectura, README/llms si cambian afirmaciones y CHANGELOG.
**Execution:** main thread
**Test approach:** check
**Changes:** trabajo vs productor nativo vs Engram; incorporación/handoff autorizados; hijos devuelven evidencia; revisión paralela no toma conducción; estado antes de pregunta y reanudación explícita; snapshots solo en paradas humanas. Títulos/resúmenes simples; publicar registro y transiciones de tareas del plan con IDs estables, estado declarado y motivo/desbloqueo estructurado, sin deducir checklist ni crear eventos por cada refresh; evidencia por task_id/segment_id y estados independientes; registro completo con fases explícitas, contexto real por evento, nueva etapa Entrega y publicación honesta de CI/commit/requiredset sin afirmar merge por checks verdes. No despliegue de guía aquí.
**Retention/promotion:** al incorporar el delta, promover el contrato final de canonical-references.md a `_support/openspec/specs/walkthrough-viewer/canonical-references.md` y usar en ambos bloques el mismo enlace `[canonical references][canonical-references]`. La definición de referencia queda fuera de los bloques normativos: al inicio del delta archivado apunta a `../../canonical-references.md` y al inicio del spec vigente a `canonical-references.md`. El bloque normativo permanece idéntico para TestSpecDeltasApplied. El cambio archivado conserva su contrato original congelado y sus enlaces internos; no es otra fuente editable. Comprobar igualdad de contenido al promover y ambos enlaces después del archivo.
**Verification:** revisar global, skills/referencias, roles, site, arquitectura y detectores tests/pilot para conflictos; sin pilotos. go test ./tests/content/...; recursos <=16000 bytes y links válidos. PR declara lugares revisados y delta global.md. Documentar cutoff limpio v3, retiro de compatibilidad/importación y recuperación operativa por directorios separados. Aplicar el delta explícito de ingestión, lifecycle/cleanup API, navegación, archivo, enlaces, toolbar, preferencias y scope; retirar importación/clasificación legacy, sin eliminar guardas reutilizadas ni fallback HTML standalone.

### T8 — Verificación integrada y entrega
- [ ] Candidato real, revisiones y validación humana completos antes de publicar.

**Closes:** AC1–AC35 (integración)
**Depends on:** T3, T7, T12
**Locations:** evidencias/checkpoint, checks existentes y procedimiento de corte.
**Execution:** main thread; hive-review-ux y Claude Code /code-review según D66
**Test approach:** check
**Changes:** reproducir CI local aplicable, matriz v3, rechazo de estado antiguo, dos productores, cinco etapas documentales, tareas globales/tramos y espera CI de Entrega, además de criterios. No producción.
**Verification:** go test ./...; go test -race ./tooling/viewer ./tooling/cli; pnpm check/lint/test/build/verify:dist desde frontend. Recorrido propio AC1–AC35, revisión UX independiente con capturas 1280/390/320 claro/oscuro/teclado, revisión D66 y aceptación D65. Enumerar commits ajenos y consultar inclusión antes de push. Verificar después de archivar que los requisitos AC28–AC35 aparecen en el spec vigente y su enlace al contrato permanente resuelve, y que el cambio archivado conserva su contexto enlazado. La restricción actual suspende refresh/reinicio habitual aunque exista maintainer.local. Compilar solo a binario propio y probar en puerto/estado propios. Tras merge preparar corte concreto separado: directorios anterior/nuevo, puerto y writers coordinados; no cambiar el servicio habitual sin autorización que cubra el nuevo state-dir. No migrar ni borrar estado anterior. Binario reconstruido no prueba runtime fresco.

### T9 — Contrato canónico final y corte a protocolo 3
- [x] Contrato3 verificado en forma · 8029899b · H18–H21 cerrados por verify_t9_contract; Go/TS267casos, frontend620, racefocal, CLI/vet/bundle PASS. Runtime catálogos T10 y proyecciones T11 pendientes.

**Closes:** AC15 (forma final3), AC28–AC33 (forma)
**Depends on:** T6; revisión del contrato canonical-references.md sin hallazgos abiertos
**Locations:** canonical-references.md como contrato propietario; tooling/viewer/timeline_contract*.go, timeline_protocol.go, frontend/src/timeline-contract*.ts, CLI parsing/usage y fixtures/schemas compartidos. Adaptadores mínimos de compilación donde los DTO retirados eran usados; no stubs que afirmen éxito runtime.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** protocolo/schema/storage final3; registries tickets/deliveries/candidates, natural identities y normalización, plan_order, inputs solo IDs y outputs con contexto publicado generado por servidor; wait candidate_id/pull_request_ref y contexto de validación; cinco operaciones nuevas y retirada de session_patch.tickets. Matriz de actores/operación/errores/límites exacta, caller-server fields cerrados. Sin adaptador para v1/provisional2.
**Verification:** paridad Go/TS y catálogo offline con fixtures válidos/inválidos; IDs desconocidos/ajenos/tipo incorrecto, identidades/URLs colisionadas, snapshots de caller, null/unknown/retired fields, órdenes y bounds. Versiones1/2 rechazadas. Compilar Go/TS y cliente; regenerar bundle/verifydist. Inventariar fallos runtime dependientes de T10 sin eliminar cobertura ni afirmar backend final. hive-verify-task independiente del contrato. Las formas del baseline2 de T1–T6 quedan supersedidas, no sus garantías de seguridad/lifecycle/historia.

### T10 — Registro único, persistencia y CLI de referencias
- [x] Registro canónico runtime/CLI verificado · 1f8077a0 · H22 cerrado por verificador; race/probes/CAS/replay/invalidation/bytes/oldstate PASS; recorrido UI realT12 pendiente.

**Closes:** AC2, AC15 (runtime3), AC28–AC33 (servidor/CLI)
**Depends on:** T9
**Locations:** tooling/viewer stores/domain/server/blobs/registry/preflight/capacity y pruebas; tooling/cli view_timeline/view_command/ayuda/recorridos.
**Execution:** hive-build-backend
**Test approach:** tdd
**Changes:** create/patch tipados y bootstrap atómico; natural-identity dedup exacto, entidad existente con valores distintos requiere patch, URLs en conflicto rechazadas. IDs del trabajo y snapshots server-only resueltos bajo lock; auditoría privada/recibos/CAS. Invalidación única de task_version/wait_version para links actuales compartidos, no-op/retry/ajeno no invalida. Captura histórica intacta, candidatos inmutables, CI conserva commit/intentos aplicables. Actualizar preflight/namespace3 y rechazar estados1/2 antes de efectos. CLI transporta nuevo contrato, sin publicación desde navegador ni fetch a proveedores.
**Verification:** cinco creates→unID; bootstrap retry/restart→mismosIDs; mismo key/distinto scope; URL/key corregidos→mismoID; colisión falla inalterada. Referencias cruzadas, lead/collaborator/local, CAS concurrente, límites/capacidad/closure, fallos durables/outcome_unknown/replay. Patch en entidad enlazada por varios caminos invalida cada tarea/espera una vez; otra/noop/retry no. Documento/artifact/snapshot/audit/providerbytes iguales. Estado1/2/corrupto con hashes+mtime idénticos. Repetir recorridos T6 afectados con binario/estado nuevos, fixtures de proveedor sin red/rerun. go test -race viewer/cli apropiado; hive-verify-task independiente.

### T11 — Proyecciones por ID y ejemplo coherente de siete tareas
- [x] Proyecciones porID/ejemplo7 verificados · f24593e9 · main628+3/check/lint/dist/GoBundlePASS, desktop1280dark/mobile440light, Escape foco/inline/contextoPublicadoActual sin overflow; integraciónrealT12 y UXfinalT8 pendientes.

**Closes:** AC28–AC32 (UI), AC34–AC35
**Depends on:** T9
**Locations:** frontend canonical selectors/client/header/content/task-registry/documents/detail-panel/fixtures/tests y static/app. No Go ni DTOs de contrato salvo coordinación por principal.
**Execution:** hive-build-frontend
**Test approach:** tdd para selectores e invalidaciones; characterization/render para composición
**Changes:** catálogos centrales por ID para referencias actuales y contexto publicado separado/etiquetado; retirar Map(URL) y entidades inline locales. Popover por plan_order con siete tareas, Buildseis y T7compartida. Una definición tipada del ejemplo genera registry y snapshot editorial inicial exacto; retirar tareas inventadas T8/T9 y checklist incongruente. No Markdown dinámico que cambie el pasado. Mantener Inicio real, toolbar compacta, todos los assets, colores/iconos, cues de contenido, inline simple, Sheets inmutables/manuales y header icono+contador.
**Verification:** cambiar una tarea/ticket/entrega en el dataset único y verificar todas las proyecciones actuales sin editar copias; snapshot abierto ofrece actualización específica, no reemplazo automático; publicación anterior/candidato/CI viejos conservan contexto. Popover7/Build6/T7unique, órdenesglobales vssegmentos independientes, IDs inválidos no render falso, loading/empty/error/retry. Capturas1280/440 con temas y320/390 relevantes; teclado/scrollfoco/galerías/casos del usuario sin regresión. Frontendcheck/lint/test/build/verifydist y ReUI. Primera vista ajustada al contrato final, mientras T10independiente sigue.

### T12 — Recorrido integrado del origen canónico
- [x] Recorrido integrado final3 verificado por main y verify_t12_integrated; AC28–AC33met. 2productores/7tasks, updatesmanuales reales, conteos7/7→6/7, hashes/auditprefix/CIcontextintactos; cloneindependiente IDs inválidos/CAS/replay/restartPASS. Commit de checkpoint en esta entrega.

**Closes:** AC28–AC33 (end-to-end)
**Depends on:** T10, T11, T13, T14
**Locations:** pruebas CLI/HTTP/frontend, evidencias retenidas y checkpoint; binario/puerto/estado propios.
**Execution:** main thread; hive-verify-task independiente
**Test approach:** check
**Changes:** recorrer el contrato final3 con dos productores y browser real. No nueva funcionalidad ni corte habitual.
**Verification:** registrar7tareas y entidades una vez, referirlas desde varias etapas/vistas/CI; transición y patch únicos→lectura actualiza todas las referencias actuales y versiones una vez. CorregirURL/key y PRopen→merged: links actuales correctos y contexto Publicado/Actual honesto. Capturar lectores, actualizar catálogo, probar aviso/manualupdate; hashes docs/images/audit/provider iguales. Reinicio/CAS/retry/fallo y invalid/foreign/rejected old1/2 sin efectos. Regenerar fixtures coherentes, sin Engram ficticio ni proveedor externo. No afirmar nuevo contrato a partir del T6baseline. Entregar evidencia a T8final.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Task checks | Local candidate | Commands and criterion verdicts | Build instruction |
| Contract/data verification | T1/T4/T5/T6 | hive-verify-task independent evidence | Build instruction |
| Full suite | Final candidate | T8 commands plus applicable CI | Build instruction |
| UI | First view T3, final candidate | Own walk and hive-review-ux | Build instruction |
| Code review | Before final human validation | D66 Claude PR-only cannot review local candidate; D77 alternative pending | No substitution without answer |
| Delivery | After human acceptance | Push, PR, merge development; conditional maintainer refresh | D65 A and AGENTS.md |

Servidor de prueba con puerto/estado propios. Volumen UI sintético, sin importar registros CAM ni editar su checkout, tracker o servicio. Sin pilotos de modelos. Ofrecer revisión de seguridad de membresía; ejecutarla solo si el usuario la pide.

## Review status and progress

2026-11-01 — flow-plan paso 6; revisión independiente terminada: H1 P1 aplicado (actor local y AC18); H2 P1 resuelto por D72 A con recover_lead local y AC19. Ninguna tarea empezada. Último commit 2f0a133b. D68 A/D69 A/D70 A/D71 A/D72 A/D73 A y ascendente adoptados. Checkpoint tras instrucción cutoff: flow-plan paso 6, reconciliación acotada terminada: H3 P1 (contratos legacy aún vigentes) corregido en el delta completo; criterios AC1–AC27 y tareas T1–T8 listos. Próximo: publicar snapshot completo y presentar handoff; build solo tras instrucción del usuario. No implementación, commits, push, PR ni merge.

2026-11-01 — D73 A añade tareas con estado solo en punto, expansión y Sheet. AC20–AC24 y T1/T2/T3/T5/T6/T7/T8 ampliados, sin nueva numeración de tareas. Revisión acotada hive-review-plan terminada sin hallazgos materiales. Verificó independencia frente al stepper instalado, contratos/evidencia/snapshots y cobertura de AC20–AC24. Checkpoint: flow-plan paso 6; listo para presentar/publicar snapshot y recibir instrucción de build. Ninguna implementación ni commit.

2026-11-01 — D74 A + corrección de filtrado por etapa: task global/segmentos independientes, conteos global y por etapa. Nueva etapa Entrega y nodo espera CI solicitados; AC25–AC27 añadidos y AC20/22–24 corregidos. Restricción: no alterar binario instalado, guía ni runtime de CAM, ni siquiera por refresh implícito. Revisión acotada terminada; H4 P1 task_version invalidado por contenido asociado y H5 P1 vigencia de intentos CI corregidos. Checkpoint flow-plan paso 6: AC1–AC27/T1–T8 listos para handoff y snapshot. Sin autorización de build ni cambios al runtime habitual.

2026-11-01 — Build autorizado mediante $flow-build (typo legibilit resuelto al change existente). D75 A fue supersedida por usuario: no worktree y no habrá otras escrituras hasta terminar. Se conserva checkout/rama refactor/flow-ceremony-2, reconciliado HEAD 5e956577. Worktree nativo 55d1 creado por llamada ya en curso, comprobado limpio/detached/mismo commit y retirado sin borrar rama. Tracker GitHub issue OPEN; project status no consultable por falta read:project, sin cambiar credenciales. Checkpoint flow-build paso 2, T1 en curso, resto T2–T8 pendiente. Runtime habitual CAM pid49211/4483 preservado.

Build runtime baseline: _support/workspace/2026-11-01-gh-306-viewer-legibility/runtime-baseline.json; 4 API records and 10 retained legacy records counted read-only in habitual state, zero candidate records (new directory not initialized). Candidate scratch /var/folders/dg/0d963h2158s10jggj2_k2w_00000gn/T/hive-gh306-tlqm2wak; no global service writes. First view additionally includes current policy mobile440x956 alongside plan390/320 and desktop1280. Viewer build event event-c784c2afc934dd47ebf0fa4d67484cf5, session b31295ed253182fca4b6f7550bc906cb2ee86ab8489c1b0e04e441d6c26c7898 version28 at start.

Preparación T4 solo lectura completada por map_t4_cutoff: preflight común antes de viewStart (orphan server.json/launch log), viewServe (mkdir/server lock), viewer.Open (mkdir/load) y RegisterRepository (lock/registro); no reutilizar loader legacy que trata corrupción como ausencia. V2 requiere marker positivo propio, validar inventario completo antes de cleanup, Run sin legacy Sync ni scanner. Conservar guards de token/origin/capability, archivos/hardlinks y receipts/reconcile. Mapa detallado en conversación; ninguna implementación T4 adelantada al gate de UI.

Ajuste de orden de verificación T1: el texto anterior pedía backend runtime v2 completo antes de construirlo (T4/T5). Se conserva el gate completo final y se acota el checkpoint T1 a contratos/paridad/compilación, registrando fallos runtime diferidos con evidencia. No cambia contratos ni aceptación del producto, y no se declara operatividad backend prematura. T2 preparado solo lectura por build_t2_frontend; no comenzará hasta verificación T1.

T1 primera verificación independiente verify_t1_contract: no aprobado, H6–H11 abiertos. H6 autoridad waits CI conductor/asignado; H7 unión audit estricta/referencias Go/TS; H8 paridad Source/review_status; H9 texto/decision_target de task_patch; H10 límites50waits/20attempts/50checks/500chars; H11 dos testsCLI puros aúnv1. build_t1_contract corrige con regresiones compartidas; T2 se mantiene preparado sin escrituras. Main full CLI99.160s terminó:4fallos, primeros2 H11, otros2 recorrido runtime diferidos; prueba aislada model panel PASS0.415s. ContratosGo/TS,compile,veterificaciónbundle focal PASS no bastan para aprobarlos seis hallazgos. No commits. Checkpoint flow-build paso2, T1 fix round en curso; próximo revalidar con el mismo verificador y después commit/T2.

T1 segunda verificación: H6–H11 corregidos y confirmados con sondas independientes, Go/TS/CLI offline/typecheck/lint/vet/bundle PASS. Main reprodujo focusedviewer0.251s y CLI0.235s PASS. H12 adicional confirmado: validadores confundían Source.ProjectRoot con ProjectKey canónico (timeline_store.go432–455 y test102–114 ya contemplan aliases/worktrees). Sigue AC11/AC19 de forma pendiente. build_t1_contract corrige separación de estructura/identidad y pertenencia resuelta por registro servidor, sin ProjectKey de cliente ni Source falseado; conservar dedup canónico host/nativeID dentro del proyecto. T2 aún sin escrituras. No commits. Próximo congelar/revalidar H12 y cerrar T1.

✓ T1 verificada y commit72aa12fdb79b4ada400f6c836ac0a78396568c51; branch refactor/flow-ceremony-2, índice ajeno vacío preservado, escaneo acotado sin coincidencias, sin instalación/refresh. Source real y contexto canónico H12 verificados en ambos lenguajes; helper server-resolved obligatorio en T4/T5 create/join/accept/recover/load. External commits66444ba0 yba2e9591 preservados (espera CI y dependencias), sin solapamiento visor; antes de publicación se consultará su inclusión según Delivery.
Se releyó flow-build desplegado tras ba2e9591: dependencias reales y mantener trabajo independiente. Corregida dependencia artificial T4→T3 por digest de bundle: prueba de bundle se serializa al estar estable, no el desarrollo backend. T4 depende solo T1, frontendlegacy ownership pasaT2; T8 dependeT3+T7. Checkpoint flow-build paso2: T2 y T4 comienzan en paralelo; T3 seguiráT2, T5 seguiráT4, nada depende de aceptación visual salvo ajustesUI. Próximo recibir candidatos/verificar, mostrar primera vista trasT3 mientras backend independiente continúa.

Wire T2/T4 congelado: indexprotocol2+retentionmanual+ttl0+timeline_sessions[]+repositories[{root,absent}]+errors[{path,reason,code?}], sin contenedores legacy. Recuperación local POST /api/timeline/sessions/:id/recover-lead conactorlocal_viewer y Origin/action-scope existente. Los alias de tipo legacy borrados en compilación pueden seguir para standalone, nunca rama de datos ni render v2.

Preparación T8 solo lectura: Claude Code2.1.316 instalado; registro activo code-review@claude-plugins-official1.0.1. /Users/jmartinez/.claude/plugins/cache/claude-plugins-official/code-review/1.0.1/commands/code-review.md describe exclusivamente review de PR (gh pr diff/view), sin modo working-tree documentado; default sin --comment no publica comentarios. D66 exige esa revisión antes de validación/push interactivo, pero aún no existe PR. No se ejecutó modelo/review ni se publicó comentario. Resolver mecanismo con el usuario en la próxima parada visual concreta; no bloquear T2/T4/T5 ni sustituirlo silenciosamente.

✓ T2 commit35cf9d40; main reprodujo regresiones5/5 y diffcheck. T3 Sheet común en curso con build_t2_frontend; T4 congelado bajo verificación independiente verify_t4_cutoff. Checkpoint flow-build paso2, sin refresh ni reinicio habitual.

T4 verificación independiente: H13 P1 viewStop elimina server.json sin preflight de estado legacy/desconocido/corrupto; H14 P1 alias/repo ausente bloquea reopen antes de binding canónico durable. Ambos asignados a build_t4_cutoff con regresiones, T4 aún no aprobada ni commit. Main focused race3.910s/CLI5.774s PASS no sustituyen estos hallazgos. Servicio habitual health idéntico al baseline, archivo instalado cambió hash por efecto ajeno a este build; observación separada runtime-observation.json, sin instalación/reinicio propio.

Publicación habitual suspendida: verificación offline del binario instalado ahora devuelve protocol/storage2 mientras serviciohabitual permaneceprotocol1. No se intentó mutación incompatible, downgrade ni restart. Evidencia/capturas quedan locales y publicación pendiente hasta un servicio compatible autorizado; build aislado continúa.

✓ T4 commit71069507: segunda verificación independiente aprobada, sondas stoplegacy/desconocido/corrupto y alias/repoausente sin cambios; identidadcontradictoria rechazada. Main race focal PASS. Checkpoint flow-build paso2: T3 frontend en curso, comienza T5 backend; no instalación/cutover.

T5 coordinación: contrato público de auditoría no se amplía. design244 pide Historial de transiciones: usa eventos especializados existentes; mutaciones genéricas quedan en auditoría durable privada tipada, con task_version atómico para copy/asignación/binding/evidencia/CI conforme design106. No nuevo mutation_audit público ni requisito adicional de UI.

✓ T3 commitb9579c06. Correcciones prevalidación: resets scopedSheet, lighttokensreales, header320wrap, icono/nombreestado tramo coherentes. Main comprobó selector cadenas separadas, documento/imagen→retorno, bloqueo/CIchecks, axe0violaciones enSheetdocumentoclaro/oscuro; capturas enworkspace/captures. T5 backend sigue independiente. Parada primera vista pendienteD76 y mecanismo reviewD77 (D66ClaudepluginPRonlyantespushincompatible); no aprobación inferida. Publicación habitual suspendida por binario2/servicio1, sin mutaciones.

Primera vista: usuario solicita corrección de alineación nodo/título y ejemplo único amplio con imágenes/assets/estados para revisar junto, más listado de aspectos (no pasos). D76 no aprobada: ronda de corrección T3 en curso por build_t2_frontend, con demo integral; D77 sin respuesta. T5 candidatoGo congelado, verify_t5_domain independiente activo; main alineó TSvalidator con pruebas130PASS/typecheck y REDprevio, sin ampliar contrato público.

Corrección visual T3: título pasó de Button de altura fija a primitive instalada en flujo normal; main midió30filas/5estados a440, delta máximo punto-centroprimera línea0.5px, sin overflow; también desktop0.5px. Ejemplo integral solicitado aún en construcción, no presentado por partes. T5 TScompanion130+153testsPASS/typecheck; revisiónindependienteGo+TS sigue activa.

Demo integral T3 congelada: ?preview=timeline&state=complete&theme=dark, mismo trabajo theme=light. Cinco etapas/nuevetareas/dieztramos; cincoestados, galerías/imagenúnica/vertical-horizontal/captions/alt/evidenciareemplazada, taskattachments/docs/cadenas/revisiones/manualnotice, vacío/inaccesiblehonesto, CIdosintentos. Main fullfrontend345/345+3scripts/typecheck/lint/distPASS, recorrido T2documento+galeríaretorno/T3galería/T4inaccesible/planhistóricoVerlaúltima; screenshotscomplete* retenidas; Sheetclaroaxe0violaciones. Corrección alineación medida30filasmax0.5px desktop/mobile. Próxima parada D78 primera vista integral, D77reviewmechanismpendiente sin sustituir. T5H15/H16Goenfixrace; TSparidadlistacon133contracttests, sin commitT5todavía.

Corrección posterior: usuario aún no veía ejemplos de imágenes por estar plegados en tareas. Demo complete ahora incluye nodo general visual-examples antes de etapas, siempre visible, con galería escritorio/móvil y captura individual reales; conserva los adjuntos originales por tarea/etapa. Main confirmó3miniaturas cargadas(bytes1280x800/390x760/1280x800) y captura complete-visible-images-1280-dark.png; fixturetestactualizadoPASS. D78 sigue pendiente, sin nueva aceptación.

Usuario prefiere nodos versión anterior (captura15:29): correcciónT3 aplicada aTaskIndicator/CI usando composición anteriorcírculo24px/borde1/icono16/fondo, pendientevacío, conectorbg-border y espaciado nativo24restaurado porlast explícito antewrappersCollapsible. Main nodecenter-firstline0.5px desktop/mobile, connector16pxvisible/noúltimocolgante; mobile sinoverflow. Focusedtasks/contenttests/typecheck/lint/ReUIchecks/build/distPASS. Captura restored-nodes-complete-1280-dark.png; D78validaciónintegral siguependiente. BackendH16bytesfixcongelado127racesuitePASSporproductor; misma verificaciónindependienteactiva, noT5commit.

✓ T5 commitfa7d4b81 incluye Go+paridadTS y correccionesT3 solicitadas/democompleta/bundle acoplados por digest. H15identidadruncommit/run_id y H16reservaeventos/recibos/bytespúblicos-privados verificadosindependienteconforeground/paused/failed/replay y textoescapadomáximo. Main focusedGoDomain-race9.398sPASS; frontend345typecheck/lint/build/distPASS. Cabecera restaurada icono+contador sinetiquetas largas, arialabel/title conservados; 320sinoverflow. D78 sigueenrevisión, D77mecanismopendiente. Checkpoint flow-build paso2: T6comienzaaislado, UIusuariorevisiónno bloqueaCLI. No instalación/reinicio habitual.

Corrección T3 en curso por4hallazgosvisuales usuario: nodosdobleborde, colorescontextuales, tarearesumensoloabreSheetvacío y contenidosrevisiónocultoshastadesplegar. build_t2_frontend implementa sin tocarcontratos; spec/design actualizadosconexcepcióninlinepara tarea sincontenido significativo. T6CLIindependientecontinúa; catalogsession_delete browser-only metadata alineada porchildautorizado, nocontratonuevo.

CuatrocorreccionesT3congeladas y verificadas: root348/348frontend+typecheck/lint/distPASS; tasksimpleT1inline sinSheet, conteosT2visiblescontraída, accesosEvidence/Historydirectos, border0px y coloressemánticos, auditoríabloqueoconalerta/Validaciónpausa. Desktop1280/mobile440+320sinoverflow, axeclaro0violaciones. Capturasfour-corrections-1280-dark.png y task-cues-collapsed-440-light.png. Commita40f1c7b despuésfa7d4b81; D78todavíapendiente, usuario nohaaceptado. T6CLIindependienteen curso, noinstalación/reiniciohabitual.

Ampliación solicitada porusuario: nododeinicio explícito, toolbarTareas/Ir alactual conpopoverdelregistroglobal, y una fuente canónica por entidadreferenciada(tareas,Git,tickets,etc.), actualizaciónúnicareflejadaenproyecciones. Usuarioautorizóajustaresquemaahora sihacefalta. CorrecciónUI sigueT3 conbuild_t2_frontend; normalizaciónes tarea añadida, descomposiciónT9+pendiente deldictamen design_canonical_refs. No se reescriben snapshots/auditoríahistórica inmutables por cambiosactuales. T6baselinecongelado(verificaciónindependiente verify_t6_cli activa), esta verificaciónno cubreel esquemanuevoaún. D77mecanismocodereview siguependiente; no publicación/instalación/cutoverautorizada.

✓ T6 commit06c72bfa: verificaciónindependienteprodutores/permisos/recuperación/accioneslocales/tareas/CI/restart/replay aprobada; productorcombinedrace876testssuperioresPASS y mainfocusedrace14.355sPASS. Se adaptaron2workflowsexistentesv1av2 sin retirarsucobertura; statusretiró expectativasexclusivaslegacyconservandocontrolesreutilizados. Nuevoalcancecanónico aúnpendiente, no se afirmaimplementado. No instalaciónnihosthabitual.

Plan ampliado guardado para revisión independiente: T9contrato3/T10backendCLI/T11UI/T12integrado; criteriosAC28–AC35, snapshot publicado vs actual explicitado, sinpunteroglobalde candidato innecesario. Naturalidentity scope/ID dedup y URLmutable; waitrefPRcanónico invalida wait/task una vez. Mainelige versión3 para evitarhandshakeadicional conbinario provisional2 observado; se conservacorte limpioautorizado y restricciónCAM. Próximo hive-review-plan antesT9, sin nueva pregunta de delivery ni aprobación de esquema ya autorizada.

UIInicio/toolbar baselinecommit8f7bbc93: main351tests/typecheck/lint/distPASS, Popovercompactorevisado1280; taskrowscanónicas sincopiasnuevas. Todavíamuestra9delbaseline; T11corrigedefiniciónúnicaasiete+v3refs. canonical-references.md+AC28–35/T9–12guardados; review_canonical_plan activoaantesdeeditarcontrato. SigueD78revisiónUI/nuevaampliación,D77mecanismopendiente.

Revisión independiente de ampliación: H17P1 deltaAC28–AC35 estaba bajoOperationalconstraint(no procesable) y carecía de destino permanente delcontrato. Corregido marcadorADDEDRequirements y retención/promociónT7+verificaciónpostarchivoT8. Sinotros hallazgosfuncionales; misma revisiónreconciliaráH17 antesT9.

H17reconciliaciónfinal: enlace dereferencia idéntico enbloque normativo y definiciones de destino fueradebloque(preamble) distintas porubicación; conservacomparacióntextualdeltas yresoluciónpostarchivo. Sinmodificarparser/check.

H17 cerrado por mismorevisor; nuevo contrato funcional sin hallazgos materiales. T9 comenzó con build_t9_contract(contextolimpio). Prototipo estable5178 sirve copia temporal de8f7bbc93, no worktree; ViteprimarioPID88499 detenido(comienzopropio), reemplazado porproceso propioexec74056 desdeui-review-runtime.json. Habitual4483 no tocado. T9puedeeditarfuentes sinromperrevisiónhumana; actualizarprototipo alfinalT11. Checkpointflow-buildpaso2 T9contrato3activo; T10/T11esperaninterfazverificada.

EvidenciaúnicaT6baselineretenida yhashverificado enworkspace/verification/t6-baseline-v2/{verification.json,recovery-conflict.png,delete-confirmation.png}; metadatossinmarcadorescredenciales, source/binaryhashes conservados. No representaelcontrato3todavíanoimplementado.

T9 primera verificación independiente: H18P1 plan_order podía aceptarse sin referencia y sin persistir; H19P2 schema candidate.reason256 vs500; H20P2 URLs numéricas aceptadasGo/rechazadasTS; H21P3 diagnósticosv2 obsoletos. Fix round con build_t9_contract activo, sin commitT9. T10/T11 esperan congelación y recheck acotado por verify_t9_contract. Formas/catálogosGo+514TS/vet/compilación/bundle pasan salvo esos hallazgos; no son evidencia del runtime T10. No actualización ni reinicio habitual. Checkpoint flow-build paso2, HEAD8f7bbc93.

✓ T9 commit8029899b: recheck independiente H18–H21 aprobado en formaAC15/28–33, con430sondasURL+11sesionesGoTS sin diferencias, overlayoriginalplan_order y sondaraceCAS/noop/replay/restart PASS. Main reprodujo620frontend+check/lint/dist, racecontrato1.847s/plan_order1.399s yCLI0.259s; stagedscan0coincidencias. Próximo T10backend/T11UI independientes, flow-build paso2. Runtime habitual intacto, sin publicación/instalación.

T10 congelado por build_t10_catalog: cinco writers/rutas, referenciascapturadas/invalidación/auditoría/CAS/replay y CLI. Main reproduce raceviewerTimeline27.743s yCLIafectada19.289s PASS; evidencia propia del productor en temp/hive-t10-irdxfbb3/evidence.json. verify_t10_catalog independiente activo; sin commitT10 hasta veredicto. T11 últimos recorridos/fidelidadfixture, no modificado servidor habitual; health sigueigualalbaseline(pid49211/protocol1/rev0b6f2380).

T11 commitf24593e9 y prototipo5178 actualizado desde copia temporal deesecommit, no worktree; exec29769. ViejoVite85884 de root detenido tras verificarcwd. Viteprimario5183exec40374 permanece solo para checksroot hasta corte propio. H22P1 artifact candidate desconocido podía dejarblob o reutilizarrecibo; productor corrigió validaciónantesdedup/write, mismoverificadorT10reconcilia. T10 sincommit; T7/T12 esperanbackendaprobado.

Parada de primera vistaT11: candidato completo final3 en5178 listo paraD79; D77mecanismoreview siguependiente (D66ClaudepluginPRonly no workingtree). T10verificación/fixH22 continúa independiente mientrasusuario revisa. No aprobaciónfinal/push/PR/merge ni cutover. Main reanudaráT7/T12 despuésbackendaprobado y ajustesUI siD79losrequiere.

### T13 — Reserved-label contract and persistence
- [x] Contrato de etiquetas/contexto verificado · efb2ecd3 · verify_t13_labels aprobóAC37forms/runtime, rootraceviewer1.579/CLI2.309s+303TS PASS; frontendREADME/bundle acoplados incluidos. UIAC36/37 sigueT14.
**Closes:** AC37 (forms/runtime)
**Depends on:** T10 verified; independent review of this extension
**Locations:** Go/TS contracts, protocol schemas/catalog, store/domain/CLI and shared tests. No presentation or static bundle.
**Execution:** hive-build-backend
**Test approach:** tdd
**Verification:** Go/TS shared fixtures for known/unknown stage enums and empty stage vs non-stage context, output and schema parity; live stage create/context patch/CAS/retry/restart and rejection without side effects; preserve producer stage context and captured history. hive-verify-task independent.

### T14 — Browser i18n throughout viewer UI
- [x] i18n visor verificado · 48eff9fc · main648+3/check/lint/dist/GoBundle PASS, Chromiumrealen-US/es-MX/fr-FR confirmaen/es/en+documentlang+etapas; productor18combinacionesygalerías/Sheets/foco PASS. IntegraciónrealT12/UXfinalT8 pendientes.
**Closes:** AC36, AC37 (UI), AC38
**Depends on:** T13 for keyed headings; spacing correction may proceed under T11 now
**Locations:** frontend language/copy/shell/timeline/preview/components/fixtures/tests and bundle; no Go or DTO validator changes.
**Execution:** hive-build-frontend
**Test approach:** tdd for language/key routing, rendered characterization
**Verification:** actual browser en-US/es-MX and unsupported locale; all system UI regions, ARIA/date/status/document lang; standard stages with/without context vs custom non-stage titles, snapshots unchanged. Popup seven rows/long list and canvas bottom at1280/440/320 light/dark, keyboard/focus/scroll/galleries, ReUI self-check, frontend suite/check/lint/build/dist. Final UX review includes this extension.

T7/T8/T12 additionally depend on T13/T14; final acceptance is AC1–AC38. User requested UI correction and localization/schema addition; D79 not accepted, D77 unanswered. T10 independent verification closed H22 and approved runtime; root commit pending. No binary/install/runtime authorization changed.

Usuario aclarókeysparaevitarrepetición yautorizóajustarskills/manual. T13 reutilizastageenumexistente, sin title_key duplicado: visor poneetapa traducida; title opcionalcontexto (vacío permitido soloetapa), otrosnodostítulo libre. T7incluyemanualdelvisor y guíaeditorialproductor (sin copiarprocedimientoenvarioshomes). T11spacingcongelado porproductor: popupgap4/py1 ytimelinepb8, rootchecks/commitpendientes. T10commit1f8077a0 trasaprobaciónindependienteH22cerrado yrootraceartifact1.880s.

T13/T14 planreview aprobado sin hallazgos; no title_key duplicado, stageyaexistenteesclave. T13backend/forms en curso; T14UI esperaformasverificadas. Correcciónspacing commit6bd42fa3 root reproducepopup440244/244pxsin overflow, paddinginferior32px yscrollfinalreal; tests12/check/lint/dist PASS. D77/D79 no respondidas. Próximo verificarT13→T14; T7manual/skills despuéscomportamientoverificado. RestricciónCAMintacta.

T13 commit efb2ecd3: mismo verificador apruebaAC37forms/runtime conprobecontextliteral+Source/documentos/snapshots/auditprefix/restart/replay; sinH23. T14 build_t14_i18n activo contexto limpio. T7coreguía+manual redactados desdeT10/T11verificados, sitio buildPASS; cierrei18n/promociónspec/links esperaT14. No modificación global.md (delta0); revisadosconsumidoresflows/roles/site/arch/tests-pilot, sinpilotos. D77/D79abiertas. PróximoT14freeze→rootchecks→T12realintegrado/T7final→T8.

T7 borrador avanzado: guía publicación~11KB con registroúnico/productores/CI/títuloscontexto; walkthrough-manifest yretención distinguenlegacyconfirm vsAPI3. README/llms/CHANGELOG/arquitectura/manualViewer/sidebar coherentes. Global.md sin cambios (delta0). Promoción actualspec33requirements con27accionesdelta (17MODIFIED/8ADDED/2REMOVED), bloquesidénticos ycontractcopyexact; amboslinksactual/deltaresuelven. SpecDelta PASS0.126s/contentPASS0.282s. Archivorealypostarchivetest siguenT8/close. T14 informa648frontend+3scripts/check/lint/build/dist/GoBundlePASS y18combinacionesrealesidioma/viewport/tema; terminaSheet/gallery/focoantesfreeze.

T14 commit48eff9fc con21rutas verificadasporhash/secretmatches0; evidencia55archivos retenidaycomparadaenworkspace/verification/t14-i18n. Root648testsPASS+3scripts, tsc/biome/dist/GoBundle0.208s ybrowsersreales3locales. T7terminaidiomas/manual+frontendREADMErebuild; siguienteT12realysuites/revisionesT8. D77sinresolver; noaceptaciónvisualfinal ni publicación.

T7 listo: skills de publicación+fallback/retención, manualViewerconidiomas ysidebar, README/llms/CHANGELOG/arquitectura yspecvigente actualizados. Sitecheck0errores/build15páginas/lint14archivos/contentPASS; manual real5195 a1280/440 sin overflow, imágenesmanualretenidas. Avisospreexistentescoleccióni18n/404 ausentes noimpidenbuild. RootrebuildfrontendREADME/distPASS. RevisiónUXindependiente completada sin hallazgos enUI/galería/doc/refs/list30/estados/manual; actualizaciónrealdeSheetquedaenT12main. RuntimeT12propioPID45670/63342 seed53→ticketpatch54→deliverypatch55; abiertoT2conservórefvieja yShowlatestactualizó. FullGo/race/vet encurso(0falloshastaahora), D77/D79aúnabiertas.

✓ T12 aprobado main+verify_t12_integrated AC28–33. Rootfixturev58 PID70997/63342; cloneindependiente terminado sin cambios aloriginal(hashestodosiguales). Rootactualen/es browser: Task/Wait conservansnapshot antepatchcanónico→aviso→Showlatest manual; Tasksummary igual;7/7→6/7 desdeunamutación;2PNG1280x800/390x760cargadasenGalería yMarkdown7readonly. DeltasT2+3/T6+1/T7+3/wait+1(dos entidades+summary+stateprobe); otros0. Document/image/candidate/oldEvents/auditprefix/CIpolicyobservationsiguales,2auditsestadosnuevossoloappend. Restart+5exactreplays idénticos. Verificadorclonado confirma12rechazosinalterados, no-op/retry/ajeno sinaviso,1ganadorCAS200/409 ylectoresactualizadossolomanualmente. Evidenciaretendidayhashverificada workspace/verification/t12-{root,independent}.

T8 checkpoint flow-verify pasos1–2: fullGo1712top-levelPASS/20packages, race894PASSviewerCLI, vetPASS; frontend648+3/check/lint/build/dist/GoBundlePASS; sitecheck0errores/build15pages/lintPASS. UXindependiente completo sin hallazgosconfirmados, actualizaciónreal cubiertaenT12. Versiónservida48eff9fcmodifiedtrue conJSexacto almanifestfinal (cambiosposterioresdocssolo); habitualpid49211protocol1rev0b6f2380healthigualbaseline, instaladohash21dexternoprevioconservado. Democongelada5178fuente2c0fc28aexec55018, manual5195. No instalación/restart/CAMeffects. Antesrevisionfinal: D77mecanismopendiente, D80inclusión15commitsajenosporpreguntar; total antes de este checkpoint30. No aceptaciónfinal ni publicación. Próximo preguntarambasrutas; trasrespuesta revisióncódigo→validaciónhumanaD65→entregaautorizada.
