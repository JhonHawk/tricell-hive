# Tareas y verificación

## Estado

Plan v4 completado e integrado mediante PR #124 (`59e2e627`), 2026-10-23. Estado actual: las cuatro tareas y controles compartidos pasaron; validación humana aprobada. Falta confirmar en el visor cotidiano la copia final con delivery; se conserva la unidad de evidencia y su worktree. Contexto inicial histórico: Base real del build `3394a43beeb2d193064cafa94c78f1e361826405`; worktree `/Users/jmartinez/.codex/worktrees/fa01/tricell-hive`, rama `feat/viewer-session-segments`. D6 autoriza implementación. Desde la base inspeccionada sólo cambiaron instrucciones de planificación, sin invalidar contratos del visor. Las tareas son secuenciales por sus contratos y archivos compartidos; la revisión por dominios se realizó en paralelo.

## T1 — Contrato de entregas validado por ambos lectores

- [x] Aceptar `delivery` opcional con paridad Python/Go y rechazo seguro de valores inválidos.

D7-A (contrato de hosts) resuelve las dos rondas fallidas anteriores y autoriza una nueva corrección/verificación acotada: host ASCII/punycode o IP; Unicode en etiqueta/ruta.

**Closes:** AC5.
**Depende de:** ninguna.
**Ubicaciones:** `tooling/viewer/manifest.go`, `manifest_test.go`, `scan_test.go`; `content/skills/flow-walkthrough/scripts/render_walkthrough.py`; `references/walkthrough-manifest.md`; `tests/skills/walkthrough_render_test.py`; fixtures de walkthrough existentes, sin tocar regression de pilotos.
**Ejecución:** delegada a implementación backend, porque el contrato está fijado y los cambios/pruebas abarcan varias fuentes sin solaparse con la preparación del entorno local que realiza el hilo principal.
**Enfoque de prueba:** tdd.
**Cambios:** contrato definido en `design.md`, sin persistir entregas en `indexData` ni modificar históricos. Añadir casos válidos PR/commit/sin campo y discrepancias de tipos/etiqueta/URL, incluyendo esquema peligroso, relativos, credenciales, controles, host ausente, barras invertidas, puerto inválido y mayúsculas de esquema para probar paridad de parsers. Tras D7-A, incluir dominio punycode e IPs admitidos, hosts Unicode (incluidos delimitadores de ancho completo) rechazados y etiqueta/ruta Unicode admitidas, con la misma regla explícita del diseño. Añadir `TestScanDeliveryInvalidKeepsLastValidCopy`: ingerir una entrega válida, alterar sólo delivery a un destino inválido y resincronizar; comprobar bytes/hash previos conservados, error acotado y ausencia de copia parcial.
**Verificación:** `go test ./tooling/viewer -run 'TestManifest|TestScanDeliveryInvalidKeepsLastValidCopy'` y `python3 tests/skills/walkthrough_render_test.py`. Los casos nuevos fallan primero por ausencia del contrato; después ambos lectores coinciden, mantienen históricos válidos y la reingestión inválida no sustituye la copia anterior. La defensa de renderizado se verifica separadamente en AC8 (enlaces seguros en viewer y standalone), mediante T3 (composición y renderer).

## T2 — Sesiones derivadas y operaciones por tramo

- [x] Agrupar copias en sesiones visuales y mantener selección, caché, actualización y acciones individuales.

**Closes:** AC3.
**Depende de:** T1 (contrato de entregas), para tipar los documentos sin inventar otro formato.
**Ubicaciones:** `tooling/viewer/frontend/src/api.ts`, `controller.ts`, `application.tsx` y `src/tests/client.test.ts`.
**Ejecución:** delegada a implementación frontend con propiedad exclusiva de estos archivos durante la tarea; contrato ya definido. No empieza T3 (composición y renderer) hasta que termine y se verifique este resultado.
**Enfoque de prueba:** tdd.
**Cambios:** tupla de agrupación, filtro antes del grupo, cargas por tramo con estados parciales, protección ante respuestas tardías, cache por revisión, selección física y fallback dentro de sesión. Excluir tracked de lecturas; conservar destino individual de archive/pin/delete. Buscar tickets sin separar sus sesiones.
**Verificación:** desde `tooling/viewer/frontend`, `node --import tsx --test src/tests/client.test.ts`. Casos con dos tickets/mismo ID, mismo ticket/distinto ID, varios proyectos/hosts, host ausente, fechas empatadas, archivados mixtos y todos archivados; cambios de clasificación sin hash; error parcial/403; pin sin recargas; eliminar con lecturas pendientes. Observar claves solicitadas por cada escritura y metadatos/documentos de hermanos intactos. La tarea deja preparada la selección/agrupación, pero sólo cierra AC3 (operaciones y filtrado por tramo). AC1 (sidebar agrupado) y AC7 (búsqueda y saltos visibles) se cierran exclusivamente en T3 (composición y renderer).

## T3 — Lectura apilada y header de entregas

- [x] Mostrar una entrada por sesión, documentos simultáneos aislados, separadores y enlaces de entrega.

Ronda 1 gate: AC4 (Dialog accesible) y AC7 (saltos/refresco) requieren corrección: overflow320/resetportalled y restauración visual del tramo al reload.

**Closes:** AC1, AC2, AC4, AC7, AC8.
**Depende de:** T2 (modelo, selección y acciones por tramo).
**Ubicaciones:** `shell.tsx`, `document.tsx`, `application.tsx`, `controller.ts` y tests cliente para retirar adaptación temporal donde ya no tenga consumidor; `messages.ts`, `styles.css`, `src/tests/component.test.tsx`, `src/dev/preview.tsx`; renderer canónico `content/skills/flow-walkthrough/assets/walkthrough.mjs` y copia `tooling/viewer/static/walkthrough.mjs`; pruebas renderer en Python y frontend.
**Ejecución:** delegada a implementación frontend con propiedad de estos archivos; coordinación con el hilo principal para la copia canónica y su igualdad. T2 ya terminado evita escritores superpuestos.
**Enfoque de prueba:** tdd.
**Cambios:** composición ReUI de `design.md`, saltos de tramo, menú individual, entregas agregadas y por documento, aislamiento de IDs/consultas/eventos y disposición de listeners; el padre gestiona scroll de sesión. Conservar standalone de un documento y capacidades por copia. Incorporar Dialog del mismo estilo con Base UI ya instalada, verificando API/documentación mediante la skill find-docs/Context7 antes de usarla; no suponer que está en el árbol ni agregar dependencia de runtime.
**Verificación:** desde frontend, `node --import tsx --test src/tests/component.test.tsx`; desde raíz, `python3 tests/skills/walkthrough_render_test.py`. Montar dos tramos con mismos índices de flows, comprobar IDs únicos y referencias ARIA, anclas y pestañas del dueño, carrusel/foco e impresión/desmontaje sin listeners residuales. Probar enlaces HTTPS, etiqueta escapada, atributos de nuevo contexto y ausencia de `href` peligroso invocando directamente el renderer; comprobar también generación standalone válida. Probar búsqueda/salto con hermanos visibles y acciones desde el separador correcto. La observación compilada compartida completa AC1 (sidebar por sesión), AC2 (documentos aislados), AC4 (entregas accesibles), AC7 (saltos y refresco) y AC8 (enlaces seguros y renderizados).

## T4 — Producción JSON, cierre y fallback HTML

- [x] Registrar entregas verificadas del trabajo actual y documentar navegación y compatibilidad.

reopened by D9: condición de salida JSON/HTML decidida por el usuario; reimplementar y verificar AC6/AC9.

**Closes:** AC6, AC9, AC10.
**Depende de:** T1 (contrato validado). La redacción usa el diseño fijado; su concordancia con la composición final se comprueba en los controles compartidos tras T3. Los archivos son distintos de los de T2/T3, por lo que puede avanzar en paralelo sin escritores superpuestos.
**Ubicaciones:** `content/skills/flow-walkthrough/scripts/render_walkthrough.py`, pruebas `walkthrough_render_test.py` y `walkthrough_status_test.py`; `content/skills/flow-walkthrough/SKILL.md`, su referencia del manifiesto y ejemplo; `content/skills/workspace-conventions/references/artifact-layout.md`; `content/skills/flow-close/SKILL.md`; `tests/skills/walkthrough_status_test.py` y pruebas contractuales de contenido que correspondan; `tooling/viewer/frontend/README.md`; `site/src/content/docs/flows/build.md`, `flows/close.md`; README/llms y `_support/docs/architecture/deployment-manager.md` donde sus afirmaciones cambien.
**Ejecución:** backend delegado sobre generador y pruebas; hilo principal sobre procedimientos y documentación. Archivos exclusivos sin solaparse con T3 (frontend/JS).
**Enfoque de prueba:** tdd para condición de disponibilidad/salidas; check para procedimientos. No ejecutar pilotos de modelos.
**Cambios:** aplicar D9 y la sección Publicación del diseño: resolver disponibilidad del ejecutable una vez, sin fallback por daemon/errores; conservar HTML anterior, validación/atomicidad/concurrencia y confirmación. Revisar el comportamiento completo y consumidores antes de editar instrucciones. La referencia del manifiesto es hogar del contrato/semántica de entrega; los pasos de producir/cerrar llevan punteros con disparador. Documentar lector nuevo antes de escritor, necesidad de reinicio explícito de runtime viejo y ausencia de inferencia/backfill. Cierre sólo sobre la unidad actual. Conservar reglas globales y roles; registrar delta de bytes de `content/guidance/global.md` como cero si sigue sin editarse. Listar superficies revisadas en la PR.
**Verificación:** `python3 tests/skills/walkthrough_status_test.py`, `python3 tests/skills/walkthrough_render_test.py`; pruebas deterministas con binario presente/sin binario (sin PATH del host), `build_page` no llamado con binario, HTML previo byte-idéntico, cierre/confirmación con JSON solamente, fallback y rollback/concurrencia conservados. `go test ./tooling/viewer` en control compartido confirma ingestión sin requerir HTML. Inspección de la cláusula canónica de retención y sus consumidores: elegir manifiesto o página conserva dependencias; fuentes de cierre/sitio deben enlazarla o describir esa misma semántica, sin convertir el JSON en archivo individual. Inspección de los pasos y punteros del productor y cierre para verificar que añaden sólo entregas conocidas, conservan candidato y no recorren otros trabajos. Fixture nueva con entrega y otra sin ella: cierre/render conserva campo/candidato y sólo modifica carpeta actual. Comprobación documental de los lectores, mensajes sin entrega y acciones por tramo. Las comprobaciones de enlaces de instrucciones y sitio forman parte de los controles compartidos.

## Controles compartidos del candidato final

Tras los checks enfocados, ejecutar una vez en el candidato final:

```sh
pnpm --dir tooling/viewer/frontend check
pnpm --dir tooling/viewer/frontend lint
pnpm --dir tooling/viewer/frontend test
pnpm --dir tooling/viewer/frontend build
pnpm --dir tooling/viewer/frontend verify:dist
go test ./...
python3 -m unittest discover -s tests/skills -p '*_test.py'
```

El bundle generado se entrega con las fuentes. Repetir build sin cambios y comprobar que no aparece delta, como pide CONTRIBUTING. No race suite salvo cambios de concurrencia Go; las épocas de requests del cliente no implican goroutines. Antes de editar el sitio, verificar sus scripts reales en `site/package.json` y ejecutar los de check/build que declara CONTRIBUTING para las páginas cambiadas, sin inventar comandos. Ejecutar el self-check ReUI de los archivos UI tocados, máximo dos rondas. Si sólo cambiaran Markdown de frontend, también regenerar bundle porque entra en su digest.

### Prueba local y validación humana

Preparar servidor Go compilado del worktree y store/registro de fixtures aislados en scratch; no cambiar ni reiniciar el daemon cotidiano. Fixtures declaradas sin escribir memorias ficticias: dos sesiones del mismo proyecto, tres tramos en una (dos tickets), otra sesión con mismo ticket, archivados mixtos, una fuente excluida, entregas duplicadas y ausentes, más de dos enlaces, títulos largos, fallos de carga y manifiesto heredado. La preparación local es parte del build, no una escritura en datos cotidianos. Los recursos propios se limpian al cerrar; la evidencia de validación sigue su retención.

Mostrar primero la composición renderizada al usuario conforme a `flow-walkthrough` antes de ejecutar la revisión UI completa. Tras aceptar composición, identificar candidato compilado por commit o tree hash y base real del build. `agent-browser` está disponible; usarlo mediante la guía y límites de `browser-automation.md`, con sesiones y stores independientes para cada agente. No usar Computer Use. Desktop 1440×900 y 1280×900; móvil 390×844; tema oscuro actual. Cuentas externas no necesarias.

- `hive-review-ux`: jerarquía sesión/tramo, densidad ReUI, encabezados y enlaces largos, menús, estados parciales, accesibilidad y responsive. Una observación por área/viewport/tema; no duplicar los recorridos funcionales del otro revisor.
- `hive-verify-change`: recorridos de los criterios, en fixture propia: buscar ticket y saltar a tramo, refrescar selección, pin/archive/delete de uno dejando intactos los hermanos; abrir «Ver todas» y comprobar destino seguro sin navegar a proveedores; pestañas/carrusel de cada tramo e impresión aislados; estados vacíos/error/reintento. Registrar observaciones y detener tras dos fallos de herramienta conforme a la guía.
- Verificación por tarea `hive-verify-task` cuando el implementador reporte cada tarea, antes de cambiar su marcador a verificada.
- `codex review` sobre el candidato local antes de validación final; contraste de bloqueantes con `hive-refute-claim`. `hive-review-security` cubre validación/uso de URL externa, aislamiento de raíces y conservación de controles de acceso. Un dictamen técnico no autoriza publicación.
- Completar pruebas locales de las aplicaciones afectadas y revisión de seguridad en el orden de `flow-build`, entregar walkthrough con capturas/recorrido y mantener stack local disponible para la validación final. Después de aceptación, refrescar sólo evidencia afectada por correcciones y aplicar el gate sobre el candidato aceptado según `ui-gate.md`.

Los revisores recibirán la revisión exacta, base, URL local real, fixtures y rutas de guía. Se registra coste/tiempo que el host exponga, sin estimar tokens no reportados. No se han ejecutado pruebas ni revisiones de código/UI durante esta planificación.

## Revisión del plan y progreso

- v1: solicitud, contrato de sesión/entrega y decisiones D1–D5 (planificación, acciones por tramo, entrega interactiva, Codex y UI completa) incorporados. Sólo archivos del plan escritos; sin rama ni worktree todavía.
- Revisión v1 recibida de tres hijos nativos `hive-review-plan`, sin historial heredado: `revision_contratos` (contratos/lectores/datos), `revision_interfaz` (composición frontend) y `revision_produccion` (productores/documentación/entrega). Contrato sólo lectura; la herramienta no expone aislamiento de filesystem por hijo y no se afirma que exista. El host no reportó duración ni uso de tokens individual; no se estimaron.
- Los tres detectaron cierre anticipado de seguridad del renderer dentro de AC5 (validadores). La revisión frontend detectó además AC1/AC7 (sidebar y saltos) indebidamente cerrados por T2 (controlador). La revisión de contratos pidió observación de ingestión válida→inválida, no sólo validación. No encontraron otros bloqueantes, criterios huérfanos ni criterios completos ya verdaderos en la base dentro de sus coberturas.
- v2 aplica las correcciones propuestas por los revisores: AC5 (lectores y conservación de copia) queda en T1 (contrato/ingestión), nuevo AC8 (renderizado seguro) en T3 (UI/renderer), T2 (controlador) sólo cierra AC3 (acciones/filtrado); AC1/AC7 (sidebar/saltos) quedan en T3. Se incorpora la prueba de resincronización en scan y los casos de paridad URL sugeridos. La advertencia sobre Dialog se resuelve con su incorporación usando Base UI existente. Estas correcciones son las propuestas por los revisores, reconciliadas por el hilo principal sin otra ronda conforme a plan-review; no cambian el alcance ni la elección del usuario. El delta ya separaba validación/renderizado en sus cláusulas, por lo que conserva el contrato revisado.
- Sin dictamen de comportamiento vivo, frescura remota o implementación; no se ejecutaron pruebas. Plan listo tras lectura de coherencia y enlaces locales; la ruta de build permanece pendiente de elección explícita. No se crearon rama, worktree, procesos ni temporales; se conserva el registro sin commit y el material previo ajeno.

- Inicio de build: D6 (implementación) autorizado. T1 (contrato y pruebas) se delega por alcance de varias fuentes; cambio de ejecución sin alterar contrato/criterios. Sólo el orquestador cambia marcadores. La base nueva añade trabajo aún no especificado al formato del plan: en este cambio no hay trabajo dentro del alcance que siga sin especificar.

- Entorno base propio: `/private/tmp/hive-viewer-session-segments.7E6THd/hive-baseline` compilado desde snapshot de base; servidor `http://127.0.0.1:54664`, PID 69181 (exec 22420), estado `baseline-state` en el mismo scratch. Fixture no Git propia con dos sesiones/tres tramos. Captura `../../../workspace/2026-10-23-viewer-session-segments/images/baseline-desktop.png` a 1440×900 CSS, DPR2, oscuro; sesión browser baseline-viewer cerrada por el hijo. Dependencias frontend instaladas con lockfile y pnpm 12.8.1, sin cambios de lockfile.
- Referencia técnica consultada para Dialog: Context7 `/mui/base-ui/v1.9.0`, tipos oficiales https://github.com/mui/base-ui/blob/v1.9.0/docs/src/app/(docs)/react/components/dialog/types.md (Root/Portal/Popup, foco y close); reutilizar al implementar T3 (composición y renderer), verificando también tipos de la dependencia instalada.

- T1 (contrato/ingestión) reportada por implementador y pendiente de verificación independiente. Implementador: rol hive-build-backend, modelo configurado gpt-6.1-sol/medium, no observado por el host. Sin procesos dejados por el hijo.

- Baseline capturada e inspeccionada por el orquestador; servidor propio PID 69181 detenido con el stop autenticado del estado aislado al dejar de ser necesario.

- Verificación T1 ronda 1: AC5 no satisfecho por URL `https://example.com|invalid/pull/1` aceptada en Python y rechazada por Go; pruebas previstas sí pasaron pero faltaba ese caso en fixture. Verificador rol hive-verify-task configurado gpt-6-astra/high, no observado por host. Scratch y procesos del verificador limpiados. Se devuelve al implementador para única ronda de corrección.

- T1 corrección reportada: caso de host con pipe añadido a fixture compartida; rechazo léxico Python acotado. Marcador pendiente de reverificación AC5 (paridad de lectores).

- Verificación T1 ronda 2: caso pipe corregido; AC5 (paridad) sigue no satisfecho por host con dos puntos de ancho completo U+FF1A, rechazado por urlsplit Python y aceptado por Go. Pruebas enfocadas verdes (Go sin caché 3.201s; Python22 en3.354s), pero no prueban paridad de esta entrada. Sin procesos/scratch del verificador. Se detienen correcciones de T1 y tareas dependientes para decisión del usuario; no se ejecuta tercera ronda por inferencia.

- v3 / D7-A: decisión explícita de host ASCII/punycode o IP, sin normalización Unicode del host; etiquetas/rutas Unicode admitidas. Actualizados AC5, contrato, delta y casos de T1. La tarea vuelve a en curso por esta decisión, sin reiniciar ni ampliar otras tareas. Revisión acotada del delta contractual en curso; el orquestador reconcilia también el estado de rama/base ya creadas.

- Revisión acotada v3 del delta D7-A (hosts) recibida por los mismos hijos de contratos y frontend, sin hallazgos materiales. Cobertura sólo coherencia de contrato y verificaciones; no certifica código. T3 (enlaces renderizados) debe comprobar host original antes de new URL, que normaliza IDN. Coste por hijo no expuesto por host.

- Implementador reporta T1 bajo D7-A (host ASCII) con fixture compartida ampliada y comandos enfocados verdes. Pendiente verificación independiente del contrato revisado; sin procesos del hijo.

- T1 verificada bajo D7-A: AC5 satisfecho por hijo independiente verificar_hosts_ascii (hive-verify-task, gpt-6-astra/high configurado, no observado); Go sin caché 3.264s y Python22 en4.022s. El orquestador leyó guardas léxicas antes de parser y aserciones de conservación en scan; no procesos restantes.

- T1 (contrato validado) guardada en commit local `769ab587`; nada publicado. T2 (modelo/acciones) comienza delegada. T4 (producción/docs) comienza en paralelo en el hilo principal: se precisa dependencia sólo en T1 porque sus interfaces están fijadas y sus archivos no se escriben por frontend. Cambio de ejecución/dependencia sin alterar contrato ni criterios; concordancia visual/documental se confirma al final.

- T4 (producción/docs) implementada por hilo principal y pendiente de verificación: 13 pruebas status, 22 renderer y go test ./tests/content pasan. Sitio lint/check sin errores; build en curso. Procedimiento canónico de delivery sigue en referencia de manifiesto; entradas flow-walkthrough y flow-close añaden puntero condicionado. Source/global y roles intactos, delta global0bytes. Superficies revisadas: global, productores/referencia, roles UX/verificación, site build/close, arquitectura manager/artefactos/roles y detectores tests/pilot (sin cambiar casos ni ejecutar pilotos). La concordancia de docs con UI final se comprueba tras T3.

- T4 verificada por hijo independiente verificar_productores: AC6 satisfecho, 13 tests status y22 renderer, diffcheck de fuentes asignadas limpio; orquestador cotejó punteros y prueba de preservación. Modelo hive-verify-task gpt-6-astra/high configurado, no observado; sin procesos. Sitio lint/check y build pasan (10páginas); avisos de directiva use astro:head-inject de index.mdx e i18n/404, efecto no investigado. No fallos ni archivos globales/roles cambiados. Concordancia UI final sigue como control compartido. Commit se coordina con frontend/bundle final porque frontendREADME entra en su sourceDigest; no versionar un bundle que no corresponda a las fuentes.

- T2 (modelo/acciones) reportada y pendiente de verificación independiente. Implementador hive-build-frontend gpt-6.1-sol/medium configurado, no observado; sin procesos. Estado auxiliar document/documentState temporal se retira al migrar consumidores en T3 (composición); no queda como implementación paralela al entregar.

- T2 verificada: AC3 satisfecho por verificar_agrupacion, 22 pruebas cliente en86.24ms, diffcheck limpio; hijo gpt-6-astra/high configurado/no observado, sin procesos. Orquestador leyó agrupador y escrituras por clave/caché. T3 comienza; propiedad application/controller y pruebas cliente transferida secuencialmente para migrar consumidores del estado, sin escritores simultáneos. T2/T3 y docs con frontendREADME se guardarán coordinados con bundle fresco para cumplir el contrato fuente/salida del repositorio; nada publicado.

- Dev server propio Vite iniciado para primera composición: http://127.0.0.1:60485/assets/app/?preview (exec31883). No usa registro cotidiano ni autenticación/servicios externos; fixture dev sólo. Se retira cuando el compilado final sirva la entrega. El modo embedded es una decisión local de composición del diseño, preserva el standalone y se valida en primera vista, sin nuevo contrato de manifiesto.

- Primera composición preparada para validación humana en dev60485: capturas first-view-desktop-v2.png y first-view-continuation-v2.png (1440×900DPR2oscuro), página generada en _support/workspace/2026-10-23-viewer-session-segments. Fuente de review registrada en visor cotidiano (sólo registro autorizado por walkthrough). Primer intento de fixture falló por fileBase no válido; corregida y añadida prueba. Snapshots anteriores conservados. Browser del capturador cerrado; sin gate UX aún.

- D8 (primera composición) aprobada por usuario. T3 reportada por implementador con 39pruebas JS, 22Python, tipos/lint/diffcheck y ReUI completados; fuente JS canónica y embebida idéntica. Sin procesos del hijo. Pendiente verificación independiente de T3 y gate completo del compilado; Dialog portalled requiere inspección tipografía/reset. Detección C2 de duration en etiquetas existentes se descarta como falso positivo, no una animación. No push autorizado hasta validación final.

- v4/D9: usuario requiere HTML sólo sin binario. Reabre T4 (publicación y procedimientos) y añade AC9 (salida por presencia de ejecutable). La investigación constató que Go no usa HTML; no cambia la API ni la composición D8 aprobada. Actualizados propuesta/diseño/delta/tarea; revisión acotada de contratos y procedimiento pendiente. T3 permanece pendiente por ajuste de import externo al sourceDigest, sin cambio de composición.

- Revisión acotada v4: contratos sin brechas; productores identificó preservación faltante al seleccionar JSON. Se aplica su corrección propuesta: AC10 (retención de recursos del manifiesto), contrato y T4 actualizados, sin otra ronda porque adopta propuesta exacta. No se altera composición ni otras autorizaciones; revisión de implementación sigue pendiente.

- T4 v4 implementada: publicador detecta binario una vez; backend reporta RED6casos y GREEN19status/22renderer con PATH aislado. Documentación reconciliada incluyendo fallback, entrega del targetdevuelto y retención de manifiesto+recursos. No HTML nuevo generado en esta sesión desde instrucción D9; HTMLv2 previo conservado histórico, no usar como entrega final. Pendiente verificarAC6/9/10 y controles del conjunto. Sin procesos del implementador.

- Candidato compilado para gate: tree a08f3d813d92abdbcc411f90c31d76fb8075b5e2 (sin registro de plan), base3394a43b; frontend source/bundle fresco tras quitar import estático externo. Binario `/private/tmp/hive-viewer-session-segments.7E6THd/hive-candidate`. Servidores propios aislados UX http://127.0.0.1:64542 PID1445(exec12606), funcional http://127.0.0.1:64543 PID1444(exec81094). Cada store tiene4copiasnormales (1archivada) y1excluida retenida; raíces git fixture propias, sin credenciales cotidianas. Auth via stdin usando helper scratch browser-auth.py y server.json, sin valores en argumentos/salida.

- T4 v4 verificada: AC6/9/10 satisfechos por verificar_publicacion_json; status19 y renderer22, sin procesos. Orquestador leyó detección única, rama sin build_page ni HTML y cláusula retención; gpt-6-astra/high configurado/no observado.
- T3: verificador de tarea reportó criterios cumplidos con41JS/22Python+lectura compilada limitada; gates compartidos aportan contraevidencia. UX no aprobado: Dialog352px/x-16 a320 por content-box; navegación278px en246 recortada; Dialog Times/Arial/márgenesUA, todo introducido. Funcional halló reload conserva clave pero headingsegundo a3348.9px y focoBODY, contra diseño. Se reabre corrección T3 sin aceptar esos gates; evidencias previas siguen válidas para casos no afectados, grabación de navegación/dialog queda stale. Sin browser/procesos de hijos. Una ronda de fix y reverificación afectada.

- Preparado baseline vivo3394a43b para UX http://127.0.0.1:49760 PID20145(exec75389), estado baseline-review-state bajo scratch. Fixture clon de UX con títulos30–50% mayores; omite únicamente delivery nuevo no soportado por baseline, limitación explícita para comparación de ese campo. Nuevas pruebas fuente de Dialog/nav/scroll en curso; los servidores candidato siguen sirviendo árbol anterior hasta reconstrucción.

- Fix T3 ronda1 reportado: resetportal/contenciónnav/restauración onReady con pruebaRED0saltos→GREEN1y pollingno-roba-foco;42JS/tipos/lint/ReUIverdes. Candidato reconstruido f49da64348ae5bb5ba146159c9128416ef81c9af. ServidoresUX64542 PID33376(exec74156), funcional64543 PID33377(exec56889); anteriores1444/1445 detenidos. Re-verificar sóloAC4/AC7 afectados y gateUXdelta+prerrequisitosfaltantes.

- T3 verificada en tree f54b12176a08dcfec00efc3b7008d28c1b4f4b17: criterios de tarea mantienen met; UX aprobó 0BlockerHigh, palabrasnormales0divisiones y hash135charscontenido a320, funcional confirmó accionesbajobarra y clicarchive/restorehermanointacto1440/390/320. Video final session-navigation-v2.mp4 H2649.633333s; orquestador lo comprueba conffprobe. Browserspropioscerrados, servidoresdepruebaquedanparaentregahumana. Los fallos previos ahoraresueltos, no se abrenissuespor introducidosya corregidos. Pendientes controlesglobales/Codex/seguridad/validaciónfinal.

- Entrega final local preparada en commit 50a836c, tras integrar origin/development con merge c2a27d53. Código runtime idéntico al tree verificado f54b12176a08dcfec00efc3b7008d28c1b4f4b17. Go vet/test completos, Python64, frontend45, bundle reproducible y sitio verdes (warnings del sitio sin escenario confirmado). Codex review nativo 0.164.0, gpt-6.1-sol/effort none observado: sin defectos accionables; seguridad independiente sin hallazgos P0–P3. Todas las tareas verificadas. walkthrough.json final publicado y registro confirmado running/found/in_progress; HTML histórico conservado sin regenerar. Pendiente D10 validación humana final antes de push/PR/merge acordados. Servidor propio funcional64543 PID48550 queda activo para validar; UX64542, baseline49760 y Vite60485 detenidos. Visor cotidiano sin reinicio.

- D10 validación final sigue pendiente: usuario detectó puntero enorme. D11 investigó y D12 autorizó ajuste exclusivo del overlay de grabación. Reproducción en agent-browser0.39.2: SVG declarado28 pero computed1440, infracción CSP style-src-elem inline enforce. Hoja construida adoptada sólo dentro del shadow root corrige18×18 y halo32, punta alineada0px; no cambia CSP/Hive/global. Grabaciónv3 H26411.266667s comprobada por orquestador; v2 histórica conservada. Overlay retirado, browser propio cerrado y helper/diagnóstico video temporales retirados. JSON final actualizado sin tocar HTML, pendiente revalidación humana D10.

- D10 aprobado tras video v3. PR #124 integrado en development59e2e627; sin CI en repositorio. Recompilado binario instalado y aplicado contenido59e2e627 (transacción6856d042f0fd8a194b6d598f5e42dcf7), sin reinicio cotidiano. Cierre reconcilia escenario vigente de lifecycle con publicación condicional HTML; no cambia implementación. Registro se archiva y entrega por PR de cierre acordado.

- Cierre de walkthrough: publicación JSON closed exit2, copia actual sin confirmar en daemon cotidiano (status running:true, found:false). HTML anterior SHA intacto; no se reinicia runtime y se conserva worktree/unidad. Servidor propio funcional64543 detenido mediante comando nativo autenticado; todos los servidores de prueba y browsers propios quedan detenidos. Diagnóstico mínimo de cursor retenido con evidencia, sin tokens.
