# Tareas de publicación

Base: 6210efb9d27594dfa0a3130a0413162714b82088.

## T1 — Candidato y favicon
- [x] Integrar los cambios del visor aceptados y el favicon.
**Closes:** AC1.
**Locations:** tooling/viewer/frontend/index.html, vite.config.ts, tooling/viewer/server.go, favicon_test.go, static/hive-mark.svg y bundle.
**Execution:** implementador frontend; verificador independiente del delta.
**Test approach:** check.
**Verification:** go test ./tooling/viewer -run TestServerServesHiveFaviconOverHTTP y observación en navegador real del icono/bytes/MIME, sin404 ni errores; tests del routing existentes y CSP conservados.

## T2 — Documentación y versión
- [x] Preparar0.5.0, changelog y README con notas acumuladas.
**Closes:** AC2.
**Depends on:** T1.
**Execution:** hilo principal por coordinación del corte.
**Verification:** lectura de VERSION y sección0.5.0, Unreleased vacío, checks/lint/build del sitio y enlaces locales.

## T3 — Gates y paquetes previos
- [x] Verificar suites y paquetes en tres runtimes antes de publicar.
**Closes:** AC3.
**Depends on:** T1, T2.
**Execution:** verificador de comandos locales y build-infra para instalación/recuperación aisladas.
**Verification:** go vet ./..., go test ./..., go test -race ./..., tests Python, frontend check/lint/test/build/verify y segundo build sin diff; sitio check/lint/build y revisión de páginas afectadas. Paquetes fuente congelada, checksums, instalación y verify en macOS ARM64 y Docker Linux ARM64/AMD64. En cada runtime crear una transacción de instalación interrumpida reproducible en un estado sintético explícito, ejecutar recover desde el paquete y verificar restauración o finalización, ausencia de transacción pendiente y preservación de archivos ajenos. Usar el esquema/fixtures y operaciones de recuperación ya existentes; no considerar recover sin pendiente ni un downgrade como prueba de recuperación. Sin instalación en hosts reales para los tests.

## T4 — Integración y publicación
- [x] Integrar development, promover master, etiquetar, publicar paquetes y sitio.
**Closes:** AC4, AC5.
**Depends on:** T3.
**Execution:** hilo principal Git/tag/release; infraestructura para identidad y salud del sitio.
**Verification:** SHAs PR/ramas/tag, artefactos descargados e iguales a preflight, tres paquetes y checksums en release; Worker con revisión/assets del tag y HTTP público con0.5.0. Los paquetes se construyen nuevamente desde checkout limpio del tag en salida única y se comparan antes de subir.

## T5 — Seguimiento de versión y refresh local
- [x] Devolver VERSION a dev en development y verificar el refresh local.
**Depends on:** T4.
**Execution:** hilo principal.
**Verification:** VERSIONdev en remoto, development contiene cambios publicados, binary/content locales al commit integrado, registros habituales preservados. El archivo del registro y la retirada de worktrees/ramas son higiene posterior del cierre, fuera del resultado de esta tarea.

## Revisión y evidencias

Revisión del plan de entrega completada en revisión2026-10-23-preflight, infraestructura/entrega, rol nativo hive-review-plan (modelo configurado por rol, no observado; duración/tokens no expuestos por host). Único hallazgo P2 resuelto adoptando su propuesta: recuperación de transacción pendiente real en cada runtime, no rollback o recover vacío. Sin nueva ronda: es la corrección sugerida por el revisor y no altera contratos ni criterios. No es revisión dedicada de código. Sin rama publicada ni tag creado todavía. Las fuentes originales del usuario no se descartarán hasta confirmar que están íntegramente preservadas e integradas y no tienen ediciones posteriores.

Verificador independiente T1 (favicon): AC1 met, tests HTTP+CSP/Host+HEAD y navegador Chrome151; MIME SVG y bytes idénticos al asset del sitio. Digest de9archivos d127739b6e78111494986845093617a5f68727f887d625dbc66411bdaff58e30; rol nativo hive-verify-task, modelo configurado Astra (no observado). Frontend 31 tests/check/lint pasan; dos builds aislados idénticos al bundle. Sitio check/lint/build pasa con advertencias Astro; preview funcional pendiente. Go vet y Python 55 tests pasan; Go 20 paquetes/3363 subtests pasan, race en curso. Tres paquetes preflight construidos juntos; instalación/recuperación en3 runtimes en curso.

Corrección incidental de alcance local aceptado: la vista previa detectó /favicon.svg404 en el sitio. Se añadió site/public/favicon.svg con el SVG de marca existente y se actualizó la nota del changelog; no cambia layouts ni código Go/frontend. Repetir sólo build del sitio y comprobación HTTP/navegador del recurso. Evidencia Go/race y frontend conserva validez para fuentes sin cambios.

Gates finales: Go vet 0.679 s, Go 20 paquetes/3363 pass96.747 s, race20 paquetes/3363 pass291.274 s sin carreras; Python 55 tests. Cuatro skips de plataforma/capturas/subprocess documentados. Paquetes0.5.0:3 runtimes pass,105 archivos/manifiesto,91 recursos instalados, interrupción tras AGENTS escrito y recuperación exacta con archivos ajenos preservados. Linux AMD64 emulado. source ID fc68a8ceb1075a7088226288f0bf8c376a6fbaae0a9c6326fae83063871c7cae. Preview sitio pasó navegación/búsqueda/teclado/móvil y ambos temas; sólo favicon404 corregido por asset existente, comprobación específica final en curso. No se repiten suites de código sin cambios.

Publicación verificada por un hijo independiente: AC4 y AC5 met. Artifacts de tag idénticos a preflight, luego descargas GitHub idénticas; Worker 100% enUUID/versiontag correctos y HTTP público igual a dist. Seguimiento PR #102 (seguimiento de versión) fusionado b85f2c4b9688aa72a2070436e71409c58b945f74; VERSIONdev remoto y local. Managercode04f64f6b,562statusinstalled/verified; contenido de esta entrega no alteró content/ frente a development. No CI configurado; CodeRabbit emitió estado success con revisión omitida, acorde a decisión sin revisión dedicada.

Coste medido: la ronda de verificación completa de Go superó5 min por ejecutar las suites normal96.747 s y race291.274 s secuencialmente; cada ejecución fue menor de5 min y no hubo fallos/reintentos. No se infiere defecto de rendimiento ni se añadieron reglas o maquinaria por esta observación. La persona decide cualquier seguimiento de método al cierre.

Higiene posterior: conservar evidencia seleccionada de sesión; detener previewAstro50840 propio, retirar checkouts/ramas de release/post/cierre sólo tras confirmar merge y limpieza. El visor habitual no pertenece a esta tarea y permanece activo. No hay delta de requisitos pendiente: la spec del visor ya se entregó con el código aprobado en PR #100 (visor y preparación).
