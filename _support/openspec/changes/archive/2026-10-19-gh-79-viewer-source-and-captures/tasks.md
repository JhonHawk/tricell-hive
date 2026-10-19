# Tareas y verificación

## Base y autorización

Base de implementación: `eee7cccd994a15bd4750b548352377e4f559e950`. Al empezar un build autorizado, revalidar y registrar su base efectiva; volver a comprobar la falsabilidad de AC1–AC7 (clasificación, presentación, archivos y capturas). No empezar implementación por guardar este plan.

Revisión v5 de seguridad añade AC8 (callback fsmonitor no ejecutado) y reabre T1 (clasificación) para AC1/AC4 (alias de filesystem y entorno determinista). RED de tres probes exactos corroborado por el orquestador en overlay sobre95b959b6; tests propios del implementador reproducen además alias de padres/raíz y consulta duplicada de mismo propietario. No cambia UI, formato de datos, política unknown o dueño, ni autoridad de entrega. Revisión acotada backend del diseño v5 concluida sin hallazgos materiales; implementación de las correcciones autorizada dentro del alcance existente. Compatibilidad Git antigua corroborada en fuentes, no ejecutada con un binario antiguo.

## T1 — Clasificación y exclusión de nuevas fuentes versionadas

- [x] Incorporar lector Git por lotes y clasificación de candidatos/copias, antes de caché y durante reconstrucción.

**Closes:** AC1, AC2, AC4, AC8.

**Depends on:** ninguna.

**Locations:** `tooling/viewer/scan.go`, `index.go`, `store.go`, `view_command.go` del CLI y pruebas de esos paquetes.

**Execution:** hijo backend, por unidad de ingesta con contrato fijado en `design.md`; responsable único de esos archivos mientras corre. El orquestador integra.

**Test approach:** `tdd`.

**Verification:** añadir y ejecutar `go test ./tooling/viewer -run 'TestSourceTracking|TestSourceClassification'` y `go test ./tooling/cli -run TestViewerSourceTracking`. Deben fallar en la base y pasar al implementar: manifest versionado forzado no copiado; local aceptado; transición sólo de índice en ambos sentidos sin reingesta; worktree; raíz sin Git; ancestro que versiona; consulta fallida vs no-repo; fuente borrada aún trackeada vs sin evidencia; índice perdido/reinicio; pin/hash/fecha preservados. El test de lector demuestra una consulta compartida por raíz para múltiples fuentes y ningún comando por manifiesto. Añadir salida mayor de 64 KiB con el manifiesto relevante después del límite y truncamiento tanto dentro de una entrada como justo después de NUL: nunca se clasifica como `local` por una lista incompleta. No afirmar estado local cuando el lector falta o falla.

Reabierta por seguridad v5: añadir regresiones con runner real para `core.fsmonitor` sintético (callback no ejecutado y tracking correcto), alias de nombre/padres en filesystem case-insensitive y distinción de archivos realmente diferentes, y variables heredadas `GIT_LITERAL_PATHSPECS`, `GIT_GLOB_PATHSPECS`, `GIT_NOGLOB_PATHSPECS`, `GIT_ICASE_PATHSPECS`. Las pruebas sensibles al filesystem declaran y verifican su condición antes de omitir un caso, sin asumir APFS o Linux por plataforma. No usar mocks para demostrar la ausencia de callback o la semántica real de Git. La prueba histórica por overlay falló en HEAD95b959b6; conservar ese RED como evidencia y reproducirlo en los tests propios.

## T2 — Apartamiento visible y lectura normal denegada

- [x] Proyectar copias apartadas y mostrarlas con ficha explicativa; bloquear su documento/archivos y preservar acciones de gestión.

**Closes:** AC3, AC7.

**Depends on:** T1 (clasificación y exclusión).

**Locations:** `tooling/viewer/index.go`, `server.go`, `store.go` si el acceso normal lo requiere; `static/viewer.mjs`, `viewer.css`, `index.html`; pruebas API/índice y fixtures de interfaz.

**Execution:** hijo backend para proyección/API, luego hijo frontend para UI con el contrato estable. No escribir en paralelo archivos de T1 (clasificación y exclusión); frontend posee sólo `static/viewer.*` e `index.html` durante su tramo.

**Test approach:** `tdd` para proyección/API y estados UI; `check` de presentación mediante gate local.

**Verification:** `go test ./tooling/viewer -run 'TestExcludedSession|TestSourceClassification'`. Probar 409 sin bytes para manifiesto y archivos aun con credenciales válidas, 403 normal sin ellas, pin/delete intactos, copias apartadas excluidas de conteos normales, reconstrucción y TTL/pins. Probar también proyección `tracked`→`local`: desaparece de `excluded_sessions` y reaparece exactamente una vez en árbol y conteos, conservando clave y pin. En el recorrido local del gate compartido, verificar ficha sin `action_required`, aviso `unknown` en lista y documento, cambio de una sesión abierta a apartada sin contenido/carrusel residual, y transiciones `local`/`unknown` en ambos sentidos con sesión abierta, mismo hash y sin reingesta: el aviso debe actualizarse al último índice. Comprobar pin/despin/borrado sólo de fixtures, errores y vacío. No repetir fixtures API completos en navegador.

## T3 — Rechazo de enlaces duros al abrir los archivos

- [x] Rechazar manifiesto/recursos con más de un enlace antes de lectura y conservar la última copia válida ante rechazo.

**Closes:** AC5.

**Depends on:** T1 (clasificación y exclusión), por propiedad compartida de `scan.go`; no hay dependencia semántica.

**Locations:** `tooling/viewer/manifest.go`, `scan.go`, pruebas de manifiesto y scanner.

**Execution:** hijo backend, secuencial al escritor de ingesta para evitar conflictos.

**Test approach:** `tdd`.

**Verification:** `go test ./tooling/viewer -run 'TestManifestHardlink|TestScanHardlink'`. Fixtures sintéticos de manifiesto enlazado, recurso enlazado, recurso sustituido tras validar y entre hash/copia; error sin instalación parcial, copia válida anterior intacta, archivo regular aceptado. La prueba de helper simula metadata sin Stat_t y comprueba rechazo explícito; no requiere cambiar dueños ni privilegios. La política no añade rechazo por Uid; verificar por inspección y prueba del helper que sólo Nlink decide esa condición. No intentar explotación multiusuario.

## T4 — Captura de escritorio a ancho completo en pantalla

- [x] Ampliar la miniatura de escritorio manteniendo impresión, móvil, tira y carrusel.

**Closes:** AC6.

**Depends on:** ninguna para edición; el gate conjunto espera T2 (apartamiento) y T3 (enlaces duros).

**Locations:** `content/skills/flow-walkthrough/assets/walkthrough.css`, copia en `tooling/viewer/static/walkthrough.css`; comentarios y documentación afectados.

**Execution:** hijo frontend con propiedad de ambos CSS, independiente de T1 (clasificación y exclusión) y T3 (enlaces duros).

**Test approach:** `check`, con geometría en navegador y comprobación existente de igualdad de assets; no un test que sólo busque el string CSS nuevo.

**Verification:** `go test ./tooling/viewer -run TestAssetsEmbeddedCopiesMatchTheSkillAssetsByteForByte` y observación local del gate: ancho de la figura desktop coincide con ancho útil de la tira ancha dentro de 2 CSS px; regla estrecha conserva asomo; página sin overflow horizontal; capturas restantes accesibles; carrusel y teclado funcionan; impresión mantiene disposición anterior. Comparar con base eee7cccd (miniaturas a media anchura). HTML de prueba debe generarse de nuevo; no usar uno histórico como prueba del cambio.

## Gates compartidos de candidato

Ejecutados y aprobados durante el build autorizado, con los límites recogidos en el historial. D4-A (entrega interactiva), D5-A (revisión por roles) y D6-A (interfaz completa) fijaron alcance y mecanismo; D7-A autorizó implementar y D10-A aceptó el candidato final.

- `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` una vez sobre candidato final, según `CONTRIBUTING.md`. `go test -race ./...` si cambia estado compartido o sincronización, como prevé la clasificación durante Sync/lectura API.
- Compilar binario aislado con `go build -o <task-scratch>/hive ./tooling/cli`. Fixture temporal con repositorios Git/worktree, workspace sin Git, copias previas/pins, manifiestos válidos y sintéticos. `view serve` en loopback con home/state-dir exclusivos del fixture; no arrancar/reiniciar visor cotidiano. Generar HTML independiente desde fixture usando `content/skills/flow-walkthrough/scripts/render_walkthrough.py`.
- Navegador seleccionado: `agent-browser` instalado en `/opt/homebrew/bin/agent-browser`. Antes de conducirlo, leer su skill y `browser-automation.md`; cada hijo usa sesión distinta sin perfil, viewport y DPR explícitos, wrapper de 45 s y datos aislados. No imprimir token/código ni guardarlos en evidencia; usar flujo de acceso del CLI existente.
- Superficies y viewports de `design.md`: sidebar normal/apartados, ficha explicativa, aviso unknown, miniaturas y carrusel; 1440/1280/390/320 y umbrales 520 contenedor/719 viewport; oscuro e impresión clara. D6-A (interfaz completa) selecciona `full`: `hive-review-ux` juzga composición, accesibilidad y presentación frente a eee7cccd (base previa al cambio), sin repetir journeys; `hive-verify-change` ejecuta los journeys funcionales y mediciones que cierran criterios. Lanzar ambos en paralelo con perfiles/sesiones y fixtures aislados; una observación por área/viewport/tema, reportar fallo tras dos intentos como no verificado. Primer render/correcciones y gate sobre candidato aceptado siguen `flow-walkthrough`; no ejecutar gates entre rondas de corrección.
- D5-A (revisión por roles) autoriza `hive-review-code` sobre diff final más `hive-review-security` sobre lectura de archivos, Git por proceso, proyección/API y exposición de documentos; refutar bloqueantes de código con `hive-refute-claim`. Ejecutar seguridad antes de validación humana y revisión de código sobre candidato aceptado, en el orden de `flow-build`; no confundir con revisión de este plan.
- Handoff humano con `flow-walkthrough`: primer render de la composición de apartados y capturas, correcciones, gate seleccionado y validación antes de publicar si se elige entrega interactiva. Agentes no abren la página del walkthrough para el usuario; entregan su enlace según flujo vigente. Evidencia en `_support/workspace/2026-10-19-gh-79-viewer-source-and-captures/` sólo al implementar; no crearla durante planificación.
- Actualizar documentación en el mismo cambio: comportamiento del visor en `site/src/content/docs/flows/build.md` y referencias de arquitectura/contrato que realmente afirmen ingesta indiscriminada o disposición anterior. Reconciliar README/llms si sus afirmaciones cambian. Añadir entrada en `[Unreleased]` de `CHANGELOG.md`, conservar `VERSION=dev`, sin release. Cada tarea incorpora su parte documental. Delta de capacidad en `specs/walkthrough-viewer/spec.md` de esta carpeta; no modificar requisitos vigentes hasta integrar/cerrar.
- Si se modifica `site/`, ejecutar desde allí `pnpm check`, `pnpm lint` y `pnpm build` con Node >=24 y pnpm 12.8.1 declarados, y comprobar la página afectada bajo `pnpm preview --host 127.0.0.1` antes de cerrar. Usar dependencias existentes; si faltan, `pnpm install --frozen-lockfile` como preparación local de implementación. No publicar sitio. Comprobación documental local acotada dentro del gate UI, sin convertirla en otra revisión estética completa del sitio.
- Tras integración autorizada, reconstruir binario y ejecutar `hive update` para contenido según AGENTS. Mantener separados resultado de refresh y activación de proceso del visor; no reiniciar servidor cotidiano. El gate runtime ya habrá probado el binario nuevo con estado aislado. Si el proceso cotidiano sigue activo, reportar aplicación pendiente hasta su próximo arranque.

## Revisión y progreso

Revisión final del plan: v2 del 2026-10-19. Tres revisores `hive-review-plan` nativos, sin historial heredado, contratos de sólo lectura y dominios backend/datos, frontend/UI y entrega/operación. El control técnico de lectura está declarado por el rol; no se verificó aislamiento técnico adicional. Duración y tokens no expuestos por el host.

Hallazgos del candidato inicial, corroborados y reconciliados por el orquestador aplicando las propuestas de los propios revisores:

- Backend/datos: alta, salida Git truncada por runner de 64 KiB podía dar falso local. Diseño y T1 (clasificación) ahora requieren detectar desbordamiento y prueban cortes incluso en NUL.
- UI: media, aviso local/unknown podía quedar obsoleto con hash constante. Diseño y T2 (apartamiento/avisos) ahora actualizan y verifican ambas transiciones de sesión abierta sin reingesta.
- Entrega: media, reemplazo de binario no activa proceso existente. Diseño/entrega distinguen archivos desplegados y próximo arranque; no se reinicia visor cotidiano.
- Entrega: media, faltaban checks de documentación y changelog. Gates ahora incluyen checks/build/preview condicionados a `site/` y entrada Unreleased sin release.

No se requiere otra ronda para aplicar las correcciones propuestas por esos revisores según `plan-review.md`. Incorporadas D4-A (entrega interactiva), D5-A (revisión por roles) y D6-A (UI completa), que concretan las decisiones pendientes ya previstas sin cambiar el diseño revisado. No quedan hallazgos materiales abiertos en las coberturas inspeccionadas. Revisión estática: no acredita conducta runtime, legibilidad ni pruebas nuevas ejecutadas.

Comprobación final: v3 acota AC1 (exclusión de nuevas fuentes) a clasificación Git completa, coherente con la excepción `unknown` advertida de D2-A (procedencia). No promete exclusión cuando Git no pudo comprobarla. Criterio y delta actualizados; única re-revisión backend completada sin hallazgos materiales. El resto de coberturas mantiene las correcciones reconciliadas de v2. Cambios posteriores a esa revisión: sólo este registro de resultado y el estado de la propuesta; sin cambios de contrato ni criterio. Duración/tokens de la re-revisión no expuestos por el host.

Estado actual: integrado en development por PR86 (clasificación, archivos y capturas), merge e9f6cc8b. D10-A aceptó el candidato; revisión completa de código y seguridad aprobadas. Las revisiones y avances siguientes se conservan como historial, no como pendientes actuales.

Higiene: preservar `_support/sessions/` preexistente. Este plan es trabajo único, no temporal. No hay procesos ni temporales propios activos. No se hace limpieza de ramas en una planificación sin merge.

## Historial de ejecución autorizada

D7-A (iniciar build), 2026-10-19: implementación autorizada. Worktree `_support/workspace/worktrees/gh-79-viewer-source-and-captures`, rama `fix/gh-79-viewer-source-and-captures`. T1 (clasificación) y T4 (capturas) iniciadas con escritores disjuntos. El checkout principal conserva el registro canónico y material previo.

T4 (capturas) pasa a verificación independiente. Implementador configurado `gpt-6.1-sol`/medium y verificador `gpt-6-astra`/high según roles nativos; modelos configurados, no observados por metadata de lanzamiento.

Preparación documental: `corepack pnpm install --frozen-lockfile` con pnpm 12.8.1 en worktree; check/lint/build del sitio pasaron. Build con avisos en portada/i18n/404 no modificados. Preview propio en http://127.0.0.1:4381, PID 67459; comprobar página `/flows/build/` por HTTP pasó. Detener con `corepack pnpm exec astro preview stop` desde site al cerrar el handoff.

T1 (clasificación) espera veredicto independiente. Implementador `gpt-6.1-sol`/medium y verificador `gpt-6-astra`/high configurados en roles; sin metadata que pruebe modelo efectivo. No hay tarea backend dependiente iniciada mientras se verifica.

Revisión v4 durante build: el verificador detectó un ciclo formal de cierre entre T1 (clasificación) y T2 (presentación). AC1/AC2/AC4 ahora describen sólo resultados del backend de T1; AC3/AC7 describen proyección y avisos de T2, con AC7 separado de la redacción anterior de AC4. No se añade comportamiento ni se altera autoridad. Las verificaciones de cada tarea permanecen, ahora sin cerrar criterios de otra superficie. Revisión acotada de esta separación pendiente; T1 continúa en verificación, T2 no comienza hasta resolverla.

Revisión UI v4: ajuste acotado incorporado tal como propuso revisor: T2 (proyección) prueba explícitamente reintegración tracked→local sin duplicados y con pin/clave. No hay cambio de contrato adicional. T1 recibió veredicto independiente met para AC1/AC2/AC4 (ingesta/clasificación); comprobación de plan backend v4 pendiente antes de marcar/commit.

T1 (clasificación) verificada independientemente: AC1/AC2/AC4 met; fuentes y aserciones corroboradas por orquestador. Revisión de plan backend v4 sin hallazgos; corrección UI de reintegración incorporada como propuesta del revisor, sin nueva ronda requerida. T2 (apartamiento) y T3 (enlaces) pueden comenzar con ownership disjunto; no hay cambio de alcance.

T4 (capturas) verificada: AC6 met en HTML nuevo y visor compilado; 678/678px y 722/722px, asomo48px hasta520, sin overflow en 1440/1280/720/719/390/320 DPR2; carrusel/foco/print comprobados. Evidencia `images/capture-desktop-composition.png` y `sample-desktop.png` en workspace de este cambio. Sigue sin commit hasta aceptación del primer render y gate UI. Scratch retenido para fixtures finales `/private/tmp/hive-verify-captures.CaFNqd/`; servidor/browser del verificador cerrados.

D8-A (primera composición de miniaturas) aceptada por usuario. T3 (enlaces duros) implementada y pasa a verificación independiente. Backend T2 (proyección/API) terminado; frontend sigue en curso. Contrato error409 mantiene text/plain existente, frontend decide por estado; no cambio de conducta ni criterio.

Runtime propio preparado: `/private/tmp/hive-gh79-runtime.TnE7Lt/` (fixture.py, home/state/repo sintéticos), servidor de pruebas http://127.0.0.1:4391, PID 78273, sesión exec 95203. Tres copias locales iniciales; una fijada y luego versionada, otra con manifiesto retirado para unknown. Ningún dato cotidiano. Binario debe reconstruirse al terminar frontend; parar sólo este proceso con `hive view stop --home ... --state-dir ...` del fixture.

T3 (enlaces duros) verificada independientemente: AC5 met; descriptor de manifiesto y cada apertura hash/copia comprobados, helper sólo Nlink, reemplazos y conservación de bytes/metadata/temporales validados. Verificación selectiva 0.383s. Fuentes citadas corroboradas por orquestador. Implementador/verificador con los mismos perfiles configurados registrados para T1; modelos efectivos no observados.

D9-A (primera composición de copias apartadas) aceptada. T2 (proyección/API/avisos) implementada y pasa a verificación independiente. Candidato UI final se recompila desde worktree; gate full con sesiones/datos separados, antes de suite compartida y seguridad. No hay cambios desde primeras vistas aceptadas que alteren composición.

Candidato congelado para gate: `c7bb1d4+f3814b2ea8ec`; binario SHA256 `31018d1f8d7aced886c939c8fc0b7e59b5484a4f353c8b5d5377cb709f4ad78b`, construido desde worktree. Fixture humano4391 reiniciado sólo en estado sintético, PID80779, sesión74086; 2 normales (local/unknown) y1apartada fijada. T2 verificador independiente, UX read-only sobre humano4391 y funcional con estado/sesión propios en paralelo. No se ejecuta aún suite global ni seguridad: siguen al gate UI.

Gate UI inicial: verificador T2 reporta AC3/AC7 met, prueba Go6tests0.434s y DOM con índices simulados sin refetch con hash estable; inspección de fuentes de índice/avisos por orquestador pendiente para marcar. UX halló ruta source partida390/320; fix viewer.css nowrap+scroll y región tabindex/label en viewer.mjs, recheck390/320 satisfactorio con teclado y sin overflow. Observaciones de presentación preexistentes: títulos truncados en sidebar y botón de despin dos líneas320; no se amplía scope. Sospecha de foco carrusel pendiente de discriminación, no defecto confirmado. Funcional real reporta transiciones/pin/delete/hardlinks correctos; vídeo falló2intentos y quedará not_recorded, sin tercer intento.

T2 (apartamiento/avisos) verificada AC3/AC7 met por hijo independiente; locators de proyección/reintegración y actualización de avisos corroborados por orquestador. Gate full concluido: UX H1 introducido corregido y re-review aprobada390/320; funcional real AC1–AC7 met en journeys asignados, con límites conservados. SospechaTab no confirma regresión: mismo efecto CSS base, otroTab vuelve Siguiente y recorrido no bloqueado. Sin nuevos cambios funcionales tras re-review. T2/T4 elegibles para commit local; fullsuite/seguridad aún pendientes antes handoff.

Usuario pidió dejar puerto como follow-up y terminar trabajo actual. Creado y verificado abierto [#85 — evaluar puerto4483](https://github.com/JhonHawk/tricell-hive/issues/85); sin implementar cambio de puerto. HTML como fallback fue discutido, sin autorización de cambio. Commits locales adicionales: b357a15 (T2 apartamiento/avisos), 95b959b (T4 capturas). No push/PR de código. Vet y43testsPython pasan; suites Go normal/race en curso, logs scratch preservados.

Fullsuite final95b959b6: go vet pasó; go test20paquetes pasó (CLI102.505s/management32.372s); Python43tests3.871s pasó; go test-race20paquetes pasó sin carreras (CLI329.510s/management200.696s). RaceCLI excedió5min: registrar coste como hallazgo incidental de tests, revisar duplicados/memoria y proponer al cierre sin abrir otro issue sin elección. Sitio check/lint/build/preview ya pasó sobre mismos docs. Seguridad acordada en curso.

Artefacto final construido desde HEAD95b959b6 limpio: `/private/tmp/hive-gh79-runtime.TnE7Lt/hive-final`, proceso25917/session38654 en4391 con mismo estado sintético. La primera consulta health sin token dio403 esperado; la consulta autenticada debe comprobar200. El visor cotidiano sigue intacto. No source diff desde gate UI; sólo metadataVCS de build limpio.

Preparación de cierre local: preview documental4381 detenido al terminar su comprobación; se conserva servidor4391 necesario para validación humana. Temporal reproducible del body de issue85 eliminado tras verificar creación. Logs de pruebas, fixtures y capturas se conservan hasta resolver validación/entrega. Auth health del artefacto final200, SHA256 f7f5bbef8f2b78d9b6c514a91dc2d88306f4f92a9858e195d16cc70232fa7007. Duplicados para coste raceCLI: búsqueda GitHub `race in:title,body` sin resultados y memoria `race CLI lento` sin resultados; propuesta incidental se consulta al cierre, no se abre aún.

Corrección de seguridad v5 implementada en lector Git; T1 (clasificación segura) vuelve a verificación independiente para AC1/AC2/AC4/AC8. Los otros criterios y la composición permanecen sin cambios. Pruebas focalizadas del implementador y overlay de seguridad pasan; suite completa de paquetes CLI/viewer y vet en ejecución.

Re-review única de seguridad v5 aprobada: callback, alias y entorno heredado corregidos; tres probes originales y regresiones nuevas pasan. Sin nuevos defectos materiales. Límite: volumen insensible omite caso de archivos distintos sólo por mayúsculas; rama revisada por fuente y Git antiguo no ejecutado. Se preservan8MiB/10s/unknown. Verificación de tarea e integración compilada aún pendientes.

T1 (clasificación segura) verificada v5: AC1/AC2/AC4/AC8 met por hijo independiente; logs /private/tmp/hive-verify-t1-bgt40_4h/{viewer,cli}.log y aserciones corroborados por orquestador.9 pruebasCLI y4viewer pasan;1 caso sensible omitido por condición real del volumen, revisado por fuente. Vet CLI/viewer pasó y suite completa de ambos paquetes pasó (CLI91.006s;viewer cache). Race selectiva del lector pasó2.581s; suite race global previa sigue válida para estado compartido sin cambios. Implementador gpt-6.1-sol/medium y verificador gpt-6-astra/high configurados, no observados. Integración del binario pendiente.

Candidato local final `da076180f6c34af6480cddea55e178d2c09cd89e`: corrección de T1 integrada en commit adicional sin amend. Verificación compilada serve→scanner→API pasó callback/alias/entorno/trackingforzado/transiciones con pin y403sin token; procesos/fixtures propios del verificador retirados. UI sin cambios desde gate full aprobado. Artefacto humano reconstruido desde HEADlimpio, SHA256 a9b326494e5255eebe37ff7ba8926ae2f14119af2c4a67dceaf932d0e15466eb; servidor4391 PID52090/session73221 con estado exclusivo sintético. Health e índice autenticados200;1copia apartada. Primera consulta usó header incorrecto y dio403 esperado, corregida con X-Hive-Token; no fallo del candidato. Validación humana pendiente antes push; code review queda tras aceptación.

D10-A (validación final) aceptada por usuario: habilita entrega acordada del candidato da076180. Revisión dedicada de código nativa hive-review-code en curso sobre diff completo desdeeee7cccd; modelo gpt-6.1-sol/high configurado, no observado. Seguridad/UX/funcional permanecen aprobados; no fuente posterior.

Publicado [PR86 — clasificación, archivos y capturas](https://github.com/JhonHawk/tricell-hive/pull/86) contra development con HEADda076180. Base remota aún contenida; sinCI configurado ni checks. Issues79/80/81 sin subissues segúnAPI. Merge pendiente de revisión dedicada.

Tras aceptación humana, servidor de fixture4391 detenido mediante su CLI; no se necesita para entrega. Se conserva unidad de capturas/manifest/page hasta fin de sesión. Resumen de regresiones Git copiado al workspace principal antes de futura retirada de worktree.

Revisión completa de código sin hallazgosP0–P3; tests focalizados viewer0.765s/CLI1.938s y diffcheck pasan. PR86 integrado por merge e9f6cc8b511226bc66e1cf1e005ff6fa000f43c5 en development, verificado MERGED; sinCI/runs postmerge. Issues79/80/81 cerrados, sin subissues. Instalación local refrescada: binarioGo desde merge y hiveupdate→apply de contenido comprometido; releasef3b87a5a8b130f650489319273a44ae8,562filasinstalled y CSS instalado idéntico. No reinicio de visor cotidiano ni publicación/master. Rebuild final desde worktree limpio evita marca dirty causada sólo por registros no versionados en checkout principal.

Cierre2026-10-19: cuatro tareas y ocho criterios satisfechos dentro de límites documentados; usuarioD10-A aceptó. Sin decisiones de producto nuevas no consultadas durante entrega. Observaciones preexistentes de truncamiento/wrap permanecen sólo como presentación. Incidental: coste raceCLI329.510s, sin carrera ni regresión demostrada; decisión de ticket pendiente en pregunta de cierre. Seguimiento de puerto85 queda abierto por usuario, sin implementación.
