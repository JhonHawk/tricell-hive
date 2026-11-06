# Tasks

Base: development e2a633e49f9ae59654bd272c574b460696ad3dae. Rama de implementación: feat/release-self-update. Worktree: .claude/worktrees/release-self-update (respecto al checkout principal). Cada tarea produce un candidato local; ningún hijo entrega Git ni cambia la instalación real. La implementación se inicia sólo tras la respuesta del usuario al handoff.

### T1 — Fijar los contratos de actualización entre versiones

- [x] Contratos tipados, formato padre y marcador de protocolo definidos y ejercitados con dos ejecutores de prueba.

**Closes:** AC2, AC5, AC9, AC13

**Depends on:** None

**Locations:** tooling/distribution, tooling/management, tooling/package, tooling/cli (protocolo interno), pruebas respectivas.

**Execution:** hive-build-backend, secuencial por su superficie compartida con las tareas siguientes.

**Test approach:** tdd.

**Changes:** Descriptor y discriminador del plan, identidad de paquete/ejecutor, versiones soportadas, mensajes prepare/apply/recover acotados y relación durable padre-hijo. Separar PreparationExecutor temporal de TargetExecutable final ligado al manifest, sin override genérico de hooks. Compatibilidad de manifest, esquema y protocolo validada antes de apply. Sin red ni activación todavía.

**Verification:** `go test ./tooling/distribution ./tooling/management ./tooling/cli ./tooling/package`; casos de protocolo ausente/desconocido, respuesta malformada, versión destino distinta y plan padre reconocible sin leer un esquema nuevo del core. Verificación independiente de criterios de esta tarea mediante hive-verify-task.

### T2 — Preparar releases verificadas y planes offline

- [x] Resolución, descarga, preview y retención del paquete/plan funcionan con transporte local simulado.

**Closes:** AC2, AC3, AC4, AC5, AC13

**Depends on:** T1 (descriptor, bootstrap y tipos del plan padre).

**Locations:** tooling/distribution, tooling/package y pruebas.

**Execution:** hive-build-backend; no edición concurrente de los módulos compartidos.

**Test approach:** tdd.

**Changes:** Origen fijo, latest estable y versión explícita; rechazo de downgrade implícito y targets sin protocolo. HTTPS/hosts por salto, límites, SHA, extracción, identidad completa y package store. Preview temporal sin retención; `--out` congela un paquete y ejecutor para aplicación offline. No usar releases reales en tests.

**Verification:** `go test ./tooling/distribution ./tooling/package`; redirecciones, timeout, truncado, checksum/nombre inválidos, symlinks/hardlinks/rutas duplicadas y traversal, límite de extracción, paquete ya existente alterado, cache miss y ausencia de writes persistentes en preview. Preview con almacén vacío, ejecutor temporal y hooks preparados hacia destino definitivo, sin escribirlo. hive-verify-task revisa los criterios cerrados por esta tarea.

### T3 — Coordinar core, activación y recuperación

- [x] Aplicación y recuperación completas preservan el límite de commit y bloquean mutaciones competidoras.

**Closes:** AC1, AC5, AC7, AC8, AC9, AC10, AC13, AC14

**Depends on:** T1 (contratos), T2 (paquete retenido y verificado).

**Locations:** tooling/management/apply.go, plan/state/onboarding/locking y nuevos módulos de update; tooling/cli (ejecutor interno y dispatcher apply/recover); pruebas respectivas.

**Execution:** hive-build-backend, tarea secuencial; el hilo principal integra y ejecuta los checks.

**Test approach:** tdd con inyección de fallos en fronteras durables.

**Changes:** Journal padre, asociación previa al journal hijo, pin de retención, orden de locks, recibo del comando y activación CAS. Antes del commit recuperar el core; después completar la activación. Revalidar paquete, ejecutor y plan en apply/recover, incluida la existencia real y digest del TargetExecutable antes de aplicar hooks. Preservar drift, recursos compartidos, voces y overrides con el motor existente. Añadir un modo explícito de planificación y validación Pi para release update: preflight bloquea declaraciones ausentes o necesitadas de reparación, no genera pasos Pi y apply rechaza planes que intenten introducirlos. No invocar proveedores opcionales en release update.

**Verification:** `go test ./tooling/management ./tooling/cli` y `go test -race ./tooling/management ./tooling/cli`; interrupciones antes/después de cada persistencia/rename/commit, pérdida del resultado del hijo tras commit, lock contention, cambio concurrente del enlace, journal hijo que alcanzaría el umbral de pruning, plan alterado/obsoleto, recovery sin red con lanzador viejo y schema destino nuevo. Preview de hooks con paquete final ausente y rechazo de apply si no se materializó exactamente. Pi válido no se ejecuta; declaración Pi ausente/inválida bloquea release update, preserva settings y estado y muestra reparación aparte; install/source conservan sus tests. hive-verify-task confirma cada criterio con evidencia por frontera.

### T4 — Migrar instalaciones al comando estable

- [x] Instalación online y offline de paquete compatible registran el enlace propio; migración repetida y desinstalación preservan material ajeno.

**Closes:** AC1, AC6

**Depends on:** T3 (recibo, coordinador y activación).

**Locations:** get-hive.sh, install.sh si se necesita, tooling/cli/install.go, tooling/management/release_paths.go y recibo de comando, tooling/distribution/release_package.go y update_protocol.go (importación local/origen de artefacto) y tests propietarios.

**Execution:** hive-build-backend, por los módulos compartidos con T3.

**Test approach:** tdd y caracterización de las instalaciones actuales.

**Changes:** Bootstrap usa la nueva instalación gestionada sin editar PATH. Instalación desde paquete completo conserva funcionamiento offline y la selección de hosts del instalador. Enlaces/archivos no administrados son conflictos; no adoptar el binario local de un mantenedor. Source install no activa comando release. Reutilizar coordinador en lugar de una segunda implementación shell de activación.

**Verification:** `go test ./tooling/distribution ./tooling/management ./tooling/cli ./tooling/package`; migración de layout anterior en home sintético, idempotencia, ausencia de PATH changes, colisión, XDG absoluto/relativo, symlink parents, uninstall con enlace alterado y paquete anterior retenido. hive-verify-task ejecuta la migración real con binarios de prueba en homes sintéticos.

### T5 — Conectar CLI y TUI al coordinador

- [x] Los dos puntos de entrada muestran el destino correcto, confirman una vez y comparten errores y recuperación.

**Closes:** AC1, AC2, AC4, AC5, AC11, AC12

**Depends on:** T3 (contrato tipado del coordinador). La migración T4 se integra y se comprueba junto con T5 en T6; no es necesaria para empezar los inputs y el flujo compartido.

**Locations:** tooling/cli/update.go, tui_update_view.go, main.go, dispatch de apply/recover/doctor y pruebas.

**Execution:** hive-build-backend, con patrones terminales existentes; no trabajo web ni ReUI.

**Test approach:** tdd y caracterización del modo source actual.

**Changes:** Default release para binarios publicados; `--version`, selección source explícita, exclusión de flags incompatibles, dev protegido, confirmación completa, plan padre, no-op real y resultados parciales. TUI Release/Source según design.md. Aviso sobre sesiones y visor cuya revisión activa no cambia.

**Verification:** `go test ./tooling/cli`; CLI no TTY rechaza antes de descargar, dry-run/out/cancel, source con dirty checkout excluye edits, argumentos incompatibles, dev sin red, no-op sin migración redundante. Walk PTY con home sintético a 80x24 y ancho reducido, release/source, spinner, error/reintento, sin hosts y recovery pendiente; capturas de texto del candidato. hive-verify-task valida criterios; la revisión UX independiente evalúa el flujo terminal en la verificación final.

### T6 — Alinear documentación y verificar el candidato completo

- [x] Documentación, delta y suite reflejan el nuevo contrato y su límite de disponibilidad pública.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9, AC10, AC11, AC12, AC13, AC14

**Depends on:** T1, T2, T3, T4, T5.

**Locations:** README.md, llms.txt, CHANGELOG.md, AGENTS.md (ejemplos de mantenimiento), _support/docs/architecture/installer.md y deployment-manager.md, site/src/content/docs/getting-started.md, delta de versioned-installation y pruebas end-to-end de update.

**Execution:** hive-write-docs para los documentos una vez fijado el comportamiento; hilo principal para integración, walkthrough y suite. Cambios de documentación no se ejecutan en paralelo con escritores de sus fuentes.

**Test approach:** check; los tests end-to-end nuevos verifican el contrato completo, no repiten el detalle de unidades.

**Changes:** Ruta de migración inicial, PATH, versión/protocolo mínimo, recuperación exacta, prereleases/downgrade, archivo de paquetes, instalación offline y origen dev explícito. Actualizar referencias de esquema obsoletas en la arquitectura tocada. Registrar global.md delta esperado cero y las superficies comprobadas. No modificar VERSION ni publicar artefactos de release.

**Verification:** `go vet ./...`; `go test ./...`; `go test -race ./...`; `python3 -m unittest discover -s tests/skills -p '*_test.py'`; `git diff --check`; checker del change folder. Validación del sitio con sus scripts instalados, sin instalar dependencias; documentar un prerrequisito local faltante si lo hay. E2E con dos paquetes/binarios locales: install → update → version/hosts/hooks readback, interrupción → recovery offline, y plan guardado → apply sin resolver latest. hive-verify-task y revisión de código conforme a la decisión de entrega.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Checks de tareas | Candidato local al terminar cada tarea | Comandos anteriores y verificador independiente por tarea | Implementación futura |
| Suite completa | Candidato final; umask 022 | Go vet/test/race, Python, diff y sitio | Implementación futura |
| Flujo terminal | PTY local; homes sintéticos; ninguna cuenta | Texto/capturas de CLI y TUI, estados del diseño; hive-review-ux | Implementación futura |
| Revisión de código | Candidato final antes de validación humana | Nativa: Codex `/review` | D3 A |
| Prueba por plataforma | Harness determinista local y compilación cruzada | darwin/arm64 y builds linux/arm64, linux/amd64 | Implementación futura; no equivale a ejecución nativa Linux |

No usar la instalación real, descargar paquetes publicados para ejecutarlos, lanzar sesiones de modelos ni reiniciar servicios para verificar. Los paquetes de prueba se compilan sólo en scratch. Si un check supera 30 minutos o requiere cuotas/cambios remotos, pedir autorización antes de ejecutarlo. No repetir builds costosos sin una razón nueva.

Entrega D2 A: tras checks y revisión, presentar el flujo terminal y esperar validación del usuario antes del push; después push, PR y merge a development. No release ni promoción a master.

## Revisión y progreso

2026-11-05 — Plan inicial. Ninguna tarea implementada, ninguna suite de producto ejecutada. Revisión de plan y entrega pendientes.
2026-11-05 — Revisión 2: H1 añade el binding de preview/destino en T1/T2/T3; H2 y D4 A añaden el preflight Pi en T3 y su documentación en T6. AC13/AC14 añadidos al delta y cobertura. D2 A y D3 A resuelven entrega y revisión. Todas las tareas permanecen sin comenzar.
2026-11-05 — Segunda revisión independiente sin gaps materiales. Checker del change folder sin salida y exit 0. HEAD avanzó a 2f39491b en trabajo ajeno de flow-close; updater sin cambios según diff de nombres. Candidato de implementación aún inexistente; ningún checkbox se marca por la revisión del plan.

## Checkpoint

Verify · ready for human validation; T1–T6 verified. Source commit f6d8cf915590d8f85572a66785bb5c1e696f9e68 on feat/release-self-update; integrated base a67feac4. D1 A, D2 A, D3 A, D4 A in force. Next action: D5 human validation before push/PR/merge. Public release/promotion are not authorized. No child agents or task test processes remain running.

All21 package results in the final go test -race -timeout=18m ./... log are ok; exact package set matches go list ./..., no failure/race markers. CLI605.134s, management272.560s. Original executor exit status was lost in daemon restart, but native kqueue observed Go process18627 exit and all21 package results were validated; no duplicate run was started. Full normal suite passed; post-fix normal CLI190.606s; latest vet, Python57 tests, Astro17pages and both offline Linux crossbuilds passed. See validation.md for evidence and limits.

Native review and its one follow-up completed. H9/H10 fixed; H11 red→green plus independent UX follow-up approved. H12 local forward-compatibility fallback red.327s→green.269s is marked Unreviewed in97c86edc; no third native review was run. Code since full-race source f6d8cf91 is unchanged; only the change record is being finalized.

Viewer: work059ebc7c344c592434939d3a5e3c91bacb647beea3cb541a67006c554319365a, confirmed evidence version52, actual candidate candidate-c438769fb761b51d0913e27e362bd20b (source f6d8cf91). The confirmation/result evidence was superseded once to correct an incorrectly typed commit reference; current reference is verified with git rev-parse HEAD. No pending publisher intent. Plan revision revision-abe5e6e279d81c5febf6d34671ac72ef. Do not restart/open viewer without the applicable user request.

Retained task resources: worktree, workspace candidate binary/images/transcripts/test logs and Linux binaries; synthetic home /private/tmp/hive-update-pty-home.UuVTEu; task-owned site/node_modules symlink to the main checkout's existing dependencies (remove only the symlink at close); private native review logs /private/tmp/hive-release-update-review.dL1772 and /private/tmp/hive-release-update-fixes-review.FfpNue. All root execution sessions are completed, including watcher31397 and crossbuilds19894/5212. Runtime session identity01a1adfa-af47-7760-a76b-d3349bd0e0fa.

2026-11-05 — T1 productor: protocolo, bootstrap y envelope implementados; suite del productor distribution 31.7s, management 40.6s, CLI 218.3s, package 3.3s. Verificador independiente: pruebas focalizadas offline PASS (.794s/.363s/2.039s), sin defecto funcional demostrado; pide ampliar apply/recover y validación semántica antes de Freeze. Corrección de cobertura en los dos tests propietarios; T2 inició en archivos nuevos sin solapar. AC2/AC5/AC9/AC13 completos siguen abiertos para T2/T3/T5; verificar T1 no los cierra prematuramente.

2026-11-05 — ✓ T1 verificado · commit 2ed68355. Main suite distribution35.294s, management60.596s, CLI212.929s, package4.782s PASS; cobertura reforzada comprobada por main y verificador independiente offline (distribution.764s, management.132s). T1 cumple su entregable contractual; AC2/AC5/AC9/AC13 end-to-end siguen abiertos hasta las tareas siguientes. Sin cambios de instalación.

2026-11-05 — Ajuste de dependencia de T5: el contrato tipado T3 ya está disponible y permite implementar CLI/TUI sin esperar la migración T4; su integración end-to-end permanece en T6. No cambia criterios ni delta. T2 terminó productor y está en verificación independiente; T3/T5 editan módulos distintos.

2026-11-05 — ✓ T2 verificado (commit de esta tarea). Productor: distribution/package31.220s/3.445s PASS, race TestRelease3.231s PASS, vet PASS. Main TestRelease1.624s PASS; verificador independiente TestRelease1.615s/package2.807s PASS sin defecto demostrado. Se amplió el owner de resolución para exigir deadline activo de máximo dos minutos y se renombró el caso de cancelación para no confundirlo con vencimiento. T2 completa distribución/preparación; los planes completos y hooks finales siguen en T3/T5.

2026-11-05 — T4 reutiliza la retención de T2 para importación offline mediante PrepareLocalRelease. El descriptor distingue archive/directory para no inventar un checksum de archivo que el instalador local no conserva. Se amplía ownership de T4 a release_package/update_protocol y sus tests; T3/T5 informados y sin solapamiento. Sin cambio de criterios ni delta.

2026-11-05 — H3 confirmado en integración por main: `go test ./tooling/cli -run '^TestReleaseUpdateCLI$' -count=1 -timeout=180s` FAIL3.272s, `offline apply: true target executor refused update: core payload is not bound to the retained package`. Retain copiaba con0600|exec y cambiaba modos originales, parte del Release ID. T4, dueño actual de release_package.go, corrige preservación exacta de permisos y añade regresión propietaria; T5 conserva el E2E. No se relaja validación de identidad.

## Historical build handoff (superseded)

The earlier T1/T2 handoff is superseded by the current Checkpoint above. All implementation and verification children have finished; their findings and results remain in the chronological progress below.

2026-11-05 — H4 confirmado por lectura del root: tui_app.handleKey devolvía tea.Quit para Ctrl-C antes de informar al Update view; appFilter sólo protegía writes, por lo que preparación/confirmación podían abandonar temporales sin cleanup. T5 amplía ownership a tui_app.go/tui_test.go para cubrir Ctrl-C, SIGINT/QuitMsg y confirmación, manteniendo bloqueo durante writes; pruebas de regresión antes del fix. No cambia AC4 ni el alcance del resultado.

2026-11-05 — Helpers T4/importación local y fix H3 en commit853b7c9a. Main E2E TestReleaseUpdateCLI PASS6.977s; H3 cerrado. Main TestLocalRelease PASS.250s, paths/command helper tests PASS.491s, TestInstallPackage PASS.620s. Verificación independiente T4 PASS (CLI1.452s, distribución1.618s, gestión.486s), incluido probe de binario local real sin Go en PATH; instalador pendiente de commit tras T3.

2026-11-05 — Verificador T5: owners PASS29.153s, H4 regresión cancelación confirmada; AC12 pendiente por H5 (SourceID comparado con VCS Revision) y H6 (comandos de recuperación sin shellQuote en rutas macOS). T5 corrige usando identidades compatibles y quoting existente, con regresiones. T3 normal suite PASS99.178s/292.214s; race gestión299.984s PASS pero CLI timeout480s en prueba existente del visor, sin race report. Final suite race pendiente de una ejecución acotada y sin escritores concurrentes.

Recursos adicionales del root: site/node_modules es un symlink propio hacia node_modules ya existente del checkout principal, para ejecutar Astro sin instalar dependencias; preservar el destino y retirar sólo el enlace propio al cierre. Astro check sesión66161 en curso.

2026-11-05 — ✓ T3 verificado (commit de esta tarea). Main matriz de gestión PASS2.281s; verificador independiente race gestión10.976s, routing1.408s, preservación1.144s y CLI4.168s PASS, sin defecto concreto. AC5/7/8/9/10/13/14 del checkpoint cumplen; AC1 integrado se cubre con el owner T6 ampliado (productor PASS8.269s) y se reejecutará en suite final. No se acredita durabilidad ante power-loss físico, sino fsync de implementación y recuperación lógica inyectada.

2026-11-05 — ✓ T4 verificado · helpers853b7c9a y commit del instalador en esta tarea. Verificador T4 probó binario real/migración/idempotencia/retirada de enlace alterado y paquetes conservados; main TestInstallPackage PASS.620s y sh -n PASS. Main E2E H3 cerrado con6.977s PASS. El owner T6 añade migración legacy→paquete1.0.0→update1.2.3 con dos hosts y PATH sin Git/Go.

2026-11-05 — ✓ T5 verificado (commit de esta tarea). Verificador owners29.153s PASS; H5/H6 corregidos red→green y follow-up.247s PASS, AC12 cumplido. Main identidad/quoting.254s PASS; programa PTY real80x24 recorre Source/Release, dev guard y resize40x24, Ctrl-C exit0; capturas workspace/images y transcript retenidos. Falta UX independiente y aceptación humana del candidato; no se atribuye validez a capturas previas a una corrección visible.

2026-11-05 — Documentación T6 escrita y contrastada en siete archivos. Astro check/build PASS ejecutados desde site/ usando dependencias ya instaladas (primer build desde raíz con --root no encontró iconos; se corrigió cwd, sin cambiar código). Build estático17 páginas; warnings existentes de i18n/404 vacíos. content/guidance/global.md sin cambios, delta0 bytes. T6 conserva pendientes full suite/race, UX y revisión nativa.

2026-11-05 — Main suite candidate d535434c PASS: go test ./... (CLI241.901s, gestión77.647s, distribución49.151s, viewer62.345s, restoPASS), go vet ./... PASS; Python57 tests/2.571s PASS. Después sólo copy de menú/source hint cambió en37eb98ef, owners CLI1.380s/gestión.183s PASS; race final cubrirá esas fuentes. Base origin/development a67feac4 integrada en merge d535434c, único conflicto CHANGELOG resuelto preservando ambas entradas.

2026-11-05 — Native Codex review terminó (log /private/tmp/hive-release-update-review.dL1772), dos bloqueos dentro de ronda1 de fixes: H9 P1 wait-floor reconstruía paquete temporal por builder ordinario (red oficial.397s); T4 corrige builder bound y elimina rebuild tardío, green.461s, source/TUI owners1.474s PASS. H10 P2 release omitía HostHookNotes; T5 añade notas antes de apply al resultado compartido y conserva core futuro opaco con aviso condicional de /hooks. Pendiente green H10, revisión nativa limitada de los fixes y full race acotado.

2026-11-05 — T6 owner E2E ampliado pasó8.269s: paquetes reales1.0.0/1.2.3, migración legacy, dos hosts, PATH sin Git/Go, stable command --version y contenido/versions verificados. PTY con dos hosts instalados recorre Source/Release/dev guard/resize40x24/Ctrl-C exit0; home temporal propio /private/tmp/hive-update-pty-home.UuVTEu se conserva para UX/validación. Capturas se reconstruyeron fielmente desde transcript incluyendo CSI delete-character; imágenes actuales sustituyen las anteriores con estado vacío.

2026-11-05 — H9/H10 corregidos en ronda1: main E2E+instalador PASS10.868s, helper/source owners PASS. Native follow-up terminó e identificó H12 P2: mismo esquema de estado no garantiza mismos campos del core, decoder local de notas bloqueaba un payload futuro válido. Root amplió owner existente, red.327s (unknown field future_schema), fallback a recordatorio no autoritativo y green.269s; no tercera revisión, commit llevará Unreviewed por fix local tras el follow-up único. H11 UX Shift+Tab: red.338s, green.279s y follow-up PTY aprobado Source→Mode→Revision→Source / Release Mode↔Version, exit0.

2026-11-05 — T6 verificador independiente confirma AC1–AC14 en owners (CLI10.893s, distribución2.327s, gestión race10.952s, source/TUI race4.488s, preservación1.087s, routing.277s), sin otra brecha. Su lectura no cubre H12 posterior; lo cubre regresión local. Normal CLI de root tras H9/H10 PASS190.606s; H11/H12 se incluyen en full race final pendiente. Fuentes ahora congeladas.

2026-11-05 — ✓ T6 verificado · final full-race21/21 paquetes ok, sin FAIL/race markers; Go18627 terminó observado por kqueue tras reinicio del daemon. CLI605.134s, gestión272.560s, distribución42.911s, viewer126.199s. CrossbuildLinuxARM64/AMD64 exit0, ELF estáticos verificados; no ejecución Linux. Fuentesf6d8cf91 sin cambios. Se cierra registro para validación humana, no publicación.
