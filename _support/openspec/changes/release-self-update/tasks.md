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

- [ ] Aplicación y recuperación completas preservan el límite de commit y bloquean mutaciones competidoras.

**Closes:** AC1, AC5, AC7, AC8, AC9, AC10, AC13, AC14

**Depends on:** T1 (contratos), T2 (paquete retenido y verificado).

**Locations:** tooling/management/apply.go, plan/state/onboarding/locking y nuevos módulos de update; tooling/cli (ejecutor interno y dispatcher apply/recover); pruebas respectivas.

**Execution:** hive-build-backend, tarea secuencial; el hilo principal integra y ejecuta los checks.

**Test approach:** tdd con inyección de fallos en fronteras durables.

**Changes:** Journal padre, asociación previa al journal hijo, pin de retención, orden de locks, recibo del comando y activación CAS. Antes del commit recuperar el core; después completar la activación. Revalidar paquete, ejecutor y plan en apply/recover, incluida la existencia real y digest del TargetExecutable antes de aplicar hooks. Preservar drift, recursos compartidos, voces y overrides con el motor existente. Añadir un modo explícito de planificación y validación Pi para release update: preflight bloquea declaraciones ausentes o necesitadas de reparación, no genera pasos Pi y apply rechaza planes que intenten introducirlos. No invocar proveedores opcionales en release update.

**Verification:** `go test ./tooling/management ./tooling/cli` y `go test -race ./tooling/management ./tooling/cli`; interrupciones antes/después de cada persistencia/rename/commit, pérdida del resultado del hijo tras commit, lock contention, cambio concurrente del enlace, journal hijo que alcanzaría el umbral de pruning, plan alterado/obsoleto, recovery sin red con lanzador viejo y schema destino nuevo. Preview de hooks con paquete final ausente y rechazo de apply si no se materializó exactamente. Pi válido no se ejecuta; declaración Pi ausente/inválida bloquea release update, preserva settings y estado y muestra reparación aparte; install/source conservan sus tests. hive-verify-task confirma cada criterio con evidencia por frontera.

### T4 — Migrar instalaciones al comando estable

- [ ] Instalación online y offline de paquete compatible registran el enlace propio; migración repetida y desinstalación preservan material ajeno.

**Closes:** AC1, AC6

**Depends on:** T3 (recibo, coordinador y activación).

**Locations:** get-hive.sh, install.sh si se necesita, tooling/cli/install.go, tooling/management (install/remove del recibo), tooling/distribution y tooling/package tests.

**Execution:** hive-build-backend, por los módulos compartidos con T3.

**Test approach:** tdd y caracterización de las instalaciones actuales.

**Changes:** Bootstrap usa la nueva instalación gestionada sin editar PATH. Instalación desde paquete completo conserva funcionamiento offline y la selección de hosts del instalador. Enlaces/archivos no administrados son conflictos; no adoptar el binario local de un mantenedor. Source install no activa comando release. Reutilizar coordinador en lugar de una segunda implementación shell de activación.

**Verification:** `go test ./tooling/distribution ./tooling/management ./tooling/cli ./tooling/package`; migración de layout anterior en home sintético, idempotencia, ausencia de PATH changes, colisión, XDG absoluto/relativo, symlink parents, uninstall con enlace alterado y paquete anterior retenido. hive-verify-task ejecuta la migración real con binarios de prueba en homes sintéticos.

### T5 — Conectar CLI y TUI al coordinador

- [ ] Los dos puntos de entrada muestran el destino correcto, confirman una vez y comparten errores y recuperación.

**Closes:** AC1, AC2, AC4, AC5, AC11, AC12

**Depends on:** T3 (contrato tipado del coordinador). La migración T4 se integra y se comprueba junto con T5 en T6; no es necesaria para empezar los inputs y el flujo compartido.

**Locations:** tooling/cli/update.go, tui_update_view.go, main.go, dispatch de apply/recover/doctor y pruebas.

**Execution:** hive-build-backend, con patrones terminales existentes; no trabajo web ni ReUI.

**Test approach:** tdd y caracterización del modo source actual.

**Changes:** Default release para binarios publicados; `--version`, selección source explícita, exclusión de flags incompatibles, dev protegido, confirmación completa, plan padre, no-op real y resultados parciales. TUI Release/Source según design.md. Aviso sobre sesiones y visor cuya revisión activa no cambia.

**Verification:** `go test ./tooling/cli`; CLI no TTY rechaza antes de descargar, dry-run/out/cancel, source con dirty checkout excluye edits, argumentos incompatibles, dev sin red, no-op sin migración redundante. Walk PTY con home sintético a 80x24 y ancho reducido, release/source, spinner, error/reintento, sin hosts y recovery pendiente; capturas de texto del candidato. hive-verify-task valida criterios; la revisión UX independiente evalúa el flujo terminal en la verificación final.

### T6 — Alinear documentación y verificar el candidato completo

- [ ] Documentación, delta y suite reflejan el nuevo contrato y su límite de disponibilidad pública.

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

Build · T1/T2 verificados; T3/T4/T5 en curso sobre contratos fijados. Base real: origin/development e2a633e4; los commits ajenos cb9b3ffb/2f39491b del checkout principal no se incluyen. User aprobó AC1–AC14 al invocar flow-build. D1 A, D2 A, D3 A y D4 A vigentes. Último commit: 2ed68355 (T1). Próximo paso: terminar T2, ejecutar checks y verificar su checkpoint antes de marcarlo. Recursos: worktree propio; ningún proceso de producto ni paquete de prueba todavía.

2026-11-05 — T1 productor: protocolo, bootstrap y envelope implementados; suite del productor distribution 31.7s, management 40.6s, CLI 218.3s, package 3.3s. Verificador independiente: pruebas focalizadas offline PASS (.794s/.363s/2.039s), sin defecto funcional demostrado; pide ampliar apply/recover y validación semántica antes de Freeze. Corrección de cobertura en los dos tests propietarios; T2 inició en archivos nuevos sin solapar. AC2/AC5/AC9/AC13 completos siguen abiertos para T2/T3/T5; verificar T1 no los cierra prematuramente.

2026-11-05 — ✓ T1 verificado · commit 2ed68355. Main suite distribution35.294s, management60.596s, CLI212.929s, package4.782s PASS; cobertura reforzada comprobada por main y verificador independiente offline (distribution.764s, management.132s). T1 cumple su entregable contractual; AC2/AC5/AC9/AC13 end-to-end siguen abiertos hasta las tareas siguientes. Sin cambios de instalación.

2026-11-05 — Ajuste de dependencia de T5: el contrato tipado T3 ya está disponible y permite implementar CLI/TUI sin esperar la migración T4; su integración end-to-end permanece en T6. No cambia criterios ni delta. T2 terminó productor y está en verificación independiente; T3/T5 editan módulos distintos.

2026-11-05 — ✓ T2 verificado (commit de esta tarea). Productor: distribution/package31.220s/3.445s PASS, race TestRelease3.231s PASS, vet PASS. Main TestRelease1.624s PASS; verificador independiente TestRelease1.615s/package2.807s PASS sin defecto demostrado. Se amplió el owner de resolución para exigir deadline activo de máximo dos minutos y se renombró el caso de cancelación para no confundirlo con vencimiento. T2 completa distribución/preparación; los planes completos y hooks finales siguen en T3/T5.
