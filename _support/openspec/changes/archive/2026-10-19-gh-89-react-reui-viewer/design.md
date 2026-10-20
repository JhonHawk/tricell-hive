# Diseño del shell React

## Contexto histórico verificado

Referencia del producto: `674df790982d7547bae2cdfa3f936e09d3397124`, todavía `development` remoto al planificar. La rama local añade únicamente preparación en `46526ed4`.

- [Servidor](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/tooling/viewer/server.go): `go:embed static`, allowlist MIME explícita y CSP `default-src 'self'`; no Node ni dependencias Go externas del visor.
- [Cliente actual](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/tooling/viewer/static/viewer.mjs): autenticación, selección, polling, generaciones, foco, borrado y recuperación que el nuevo cliente debe conservar.
- [Renderer](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/tooling/viewer/static/walkthrough.mjs): `renderWalkthrough(root, manifest, {fileBase})` independiente del shell.
- [Prueba de assets](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/tooling/viewer/assets_test.go): igualdad byte por byte con los assets del reporte offline.
- [Empaquetador](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/tooling/package/main.go): compila Go desde fuente congelada para Darwin/Linux; no construye frontend.
- [Contribución](https://github.com/JhonHawk/tricell-hive/blob/674df790982d7547bae2cdfa3f936e09d3397124/CONTRIBUTING.md): no hay CI; las comprobaciones requeridas corren localmente.

MCP/skill ReUI `68f0e6fa91` operativos. El acceso Pro está probado, pero no se necesita ni se usará para esta implementación. En esa referencia aún no había frontend React del visor ni referencia renderizada React aceptada. No se afirma mejora visual medida.

## Arquitectura propuesta

Paquete independiente en `tooling/viewer/frontend/`, sin convertir el repositorio en workspace Node ni unirlo al sitio Astro. React 19, TypeScript, Vite, Tailwind v4, shadcn y componentes MIT de ReUI. Usar Node >=24 y pnpm 12.8.1, coherentes con el sitio; fijar versiones exactas compatibles y lockfile al crear el paquete. Selección `base-nova`, corroborada documentalmente abajo; no mezclar APIs Base UI/Radix.

Los archivos fuente introducidos, sus identificadores y comentarios van en inglés; textos de producto se mantienen en inglés/español mediante claves semánticas inglesas. Importar solo componentes que el shell utiliza y sus dependencias necesarias. No importar bloques premium ni reproducir su código. Registrar origen/licencia y conservar avisos en el paquete y en `THIRD_PARTY_NOTICES.md` según los archivos realmente distribuidos.

Corroboración documental del 19 de octubre: [ReUI Registry](https://reui.io/docs/registry) declara `base-nova` como default; [Base UI CSPProvider](https://base-ui.com/react/utils/csp-provider) permite `disableStyleElements` y documenta que asignaciones de estilo desde JavaScript cliente no equivalen a estilos inline del HTML. Usar SPA cliente, `CSPProvider disableStyleElements`, reglas equivalentes de scrollbar y aislamiento del root en CSS externo; no usar un provider de tema que inyecte script inline. Esto respalda el enfoque, no reemplaza el probe CSP real. Metadatos npm consultados: React 19.3.1, Tailwind 4.3.3, Base UI 1.9.0, Vite 8.3.4 y plugin-react 6.1.2 (peer Vite 8); fijar solo versiones cuyo build conjunto pase.

### Construcción y assets

Versionar el bundle generado en `tooling/viewer/static/app/`. Mantener `walkthrough.mjs` y `walkthrough.css` fuera de ese output. `build` limpia únicamente su directorio dedicado y produce JS/CSS locales sin source maps públicos; el HTML de producción no usa scripts inline ni URLs de desarrollo. Usar base de Vite `/assets/app/` y `assetsInlineLimit: 0`, correspondiendo a los archivos bajo `static/app/`; servir el HTML de entrada en `/`. Derivar del build un manifiesto explícito con todos los entrypoints/chunks, incluidos nombres hash; todos deben tener MIME y ruta permitidos, sin exponer mapas/fuentes ni paths arbitrarios. No tratar los nombres hash como previsibles ni exponer el directorio completo por prefijo.

Definir `pnpm build`, `pnpm check`, `pnpm lint`, `pnpm test` y `pnpm verify:dist` al crear el paquete. El último comprueba correspondencia del output con fuentes/configuración/lockfile sin modificarlo; Go también verifica un digest de fuentes y hashes de los archivos declarados sin requerir Node. La generación actualiza el manifiesto únicamente al terminar con éxito. Verificación final: reconstruir con lockfile congelado y exigir ausencia de diff en bundle/manifiesto. Incluir lockfile, configuración, componentes copiados y scripts en la entrada del digest; excluir node_modules, resultados de pruebas y el propio output. Esto detecta obsolescencia, no es una firma ni prueba de seguridad.

`go build ./tooling/cli` y el empaquetado normal consumen output versionado. Node/pnpm se requieren solo al modificar/reconstruir frontend. Esta decisión conserva los requisitos de construcción Go actuales; cambiarla posteriormente requiere revisar las tareas de empaquetado.

### API y estado

Conservar endpoints y payloads existentes, salvo la ampliación explícita de ciclo de sesión descrita abajo. El único cliente nuevo reemplaza al consumidor vanilla en el mismo release. No hay coexistencia ni migración de meta.json. Recuperar el shell revirtiendo React debe conservar el lector compatible con session.status y su dato derivado; no volver íntegramente al lector antiguo tras producir manifiestos nuevos, porque los rechazaría. Una reversión completa anterior a este contrato requiere un procedimiento explícito separado y no se realiza por inferencia.

Encapsular peticiones y estado de acceso sin imprimir credenciales. Consumir y eliminar el fragmento antes de montar efectos; un código se intercambia una sola vez aunque React repita montaje/efectos. Conservar recuperación de token desde sessionStorage y funcionamiento cuando storage falla. Mantener header `X-Hive-Token`, same-origin y fileBase con capability; no guardar secretos en logs, URL ni fixtures versionados.

Índice conserva proyectos/tickets/sesiones, `excluded_sessions`, errores y datos de retención. Polling 10 s, reconexión 5 s y suspensión en pestaña oculta como hoy. Cancelación/generaciones deben impedir que una respuesta anterior sobrescriba acceso, selección o una eliminación reciente. Cambios de hash/clasificación actualizan el recorrido sin resetear foco/scroll innecesariamente.

### Renderer compartido y CSS

React es dueño del shell; un contenedor ref dedicado es dueño del DOM imperativo del renderer. No reconciliar hijos de ese contenedor con JSX. Invocar renderer cuando cambia la sesión/contenido; verificar manejo de desmontaje y listeners existentes antes de inventar una API de limpieza. El renderer sigue siendo el mismo utilizado por reportes offline.

Separar tokens/estilos del shell de los del recorrido y evitar que preflight/estilos globales alteren capturas, títulos, galería o impresión. No prometer recolorear el recorrido: conserva el diseño y tema propios del documento. La compatibilidad de overlays/posicionamiento con CSP se prueba en servidor Go, no se infiere del dev server. Si un componente exige cambiar la política, detener esa parte y presentar alternativas; no añadir unsafe-inline automáticamente.

## Composición propuesta

Superficie operativa; arquetipo detalle de registro con shell de navegación; familia Frame para el panel del shell; densidad compacta en controles que admitan `size="sm"`; ninguna animación de entrada. No envolver el documento offline en tarjetas/Frames anidados.

Escritorio: navegación lateral por proyecto/ticket/sesión; arriba contexto y acciones de pin/borrado; debajo avisos de procedencia y documento. Móvil: navegación en Sheet con retorno de foco, un panel de lectura. No añadir panel IA, KPIs, ordenamiento persistente ni búsqueda como efecto secundario del cambio de framework.

Reutilización propuesta: Sidebar/Sheet, Collapsible, Button, AlertDialog, Skeleton y toast de shadcn; Badge/Alert y Frame de ReUI cuando su rol lo requiera. Confirmar nombres/APIs mediante MCP y archivos instalados; no instalar componentes solo porque figuran aquí. El árbol actual usa listas y botones expandibles: conservar esa semántica o adoptar un Tree únicamente con su patrón completo de teclado, nunca mezclar ambos.

Estados: acceso pendiente/denegado, índice vacío con acción de registro, carga sin layout jump, error con reintento, desconexión/reconexión, copia excluida, origen desconocido, documento borrado y errores de ingesta. El shell conserva idioma del navegador y el contenido el de su manifiesto. Mantener el tema oscuro actual; no añadir selector de temas en esta migración. Documento conserva sus estilos. El primer render se muestra al usuario antes de gate independiente.

## Verificación compartida autorizada al implementar

- Frontend: `pnpm --dir tooling/viewer/frontend check`, `lint`, `test`, `build`, `verify:dist` (scripts por introducir). Pruebas de estado/API con red controlada; no escribir pruebas que solo busquen strings.
- Final Go/Python: `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`. Race solo si cambian goroutines/locks/estado concurrente; el cambio a assets no lo exige por sí mismo.
- Bundle compilado servido por candidato Go local con store temporal aislado, sin registro de repos cotidianos. Fixtures de manifiestos reales válidos y recursos generados; preservar fuentes antes/después de borrar copia.
- D4-A (UI completa): `full`: UX evalúa composición/contraste/foco en 1440×900, 1024×768 y 390×844, tema oscuro del producto, tanto con preferencia del sistema clara como oscura, y etiquetas largas. Funcional verifica apertura, storage indisponible, selección, pin/borrado, cambios de clasificación sin hash, reconexión/respuestas fuera de orden, captura/video/carrusel e impresión. Sesiones y stores separados, `agent-browser` instalado, sin perfiles personales; seguir su referencia de automatización.
- D3-A (revisión independiente), seguridad: cliente de autenticación, CSP real, assets/MIME/allowlist, URLs de recursos, dependencias y licencias; alcance acotado al diff, no auditoría general.
- Capturar baseline vanilla desde el commit de referencia en entorno aislado durante build; no reiniciar el visor cotidiano. Las capturas no prueban una mejora subjetiva general.
- Primer render, gate UI, suite y seguridad ocurren antes del walkthrough final y de la pausa humana de entrega interactiva elegida mediante D2-A (entrega interactiva). Mantener stack y capturas para la validación humana; limpiar al cerrar.

## Ampliación: ciclo explícito del trabajo de una sesión

El usuario pidió badge de estado y aclaró que Cerrada corresponde a flow-close. No es detección del proceso host, ni derivación de informe final, clasificación Git o antigüedad. El manifiesto agrega `session.status` opcional `in_progress|closed`; los productores nuevos escriben in_progress al generar/reabrir el recorrido. Flow-close marca closed para el trabajo actual que cierra (completado, abandonado o pausado), no para todos los registros de una sesión host. Reanudar explícitamente ese trabajo vuelve a in_progress. Ausencia se proyecta como unknown; no reclasificar retrospectivamente por kind.

Go y Python validan el campo y rechazan otros valores/tipos. Summary/sessionFrom y Session agregan `session_status` derivado; load reconstruye las copias existentes desde su manifiesto sin migración de meta.json. La UI mapea exactamente En curso/Cerrada/Sin informar y conserva procedencia en un indicador independiente.

Flow-close actualiza y renderiza solo el manifiesto de este trabajo antes de eliminar worktree o sus capturas. Usar el renderer existente con una operación explícita y atómica para el campo; no modificar HTML manualmente. Ampliar `hive view status --dir <work folder>` para consultar la copia exacta mediante la API autenticada existente y devolver solo found/session_status, sin tokens/capabilities. El cierre espera hasta 15 segundos por confirmación del estado closed de esa fuente/sesión antes de borrar su unidad de evidencia. `found=true` significa coincidencia de fuente canónica, session.id, key y content_hash de la fuente actual, no solo existencia de cualquier copia. Un wrapper de lectura de viewer reutiliza os.Root, validateRoot, sessionKey y hashFiles para calcular identidad/huella sin abrir Store ni duplicar algoritmo. Se calcula antes de la consulta autenticada y se vuelve a comprobar después; cambios de fuente durante la consulta no autorizan limpieza. Cuando no hay coincidencia se devuelve found=false y session_status=unknown, sin capacidades. La huella incluye recursos: una imagen modificada con manifiesto igual invalida una copia antigua. Timeout, offline, ausencia de copia o runtime antiguo conservan la unidad aún necesaria y se reportan, sin iniciar/reiniciar servidor ni abrir Store paralelo. La espera debe identificar la revisión actual del manifiesto además de source/session para no aceptar una copia closed antigua al reabrir.

Coordinación: el visor cotidiano está ejecutándose (observado en status), y el antiguo validador estricto rechaza el campo nuevo. Binario y contenido nuevos se entregan coordinadamente; el refresh de disco no sustituye reiniciar un proceso viejo. Antes de activar productores nuevos en la instalación cotidiana, resolver explícitamente con el usuario el restart del visor concreto; no hacerlo por inferencia. No añadir compatibilidad perpetua ni un segundo formato.

Cobertura nueva: parser Python/Go ausente y dos valores válidos, inválidos; store startup/scan/recarga preserva status y default unknown; status --dir distingue homónimos y revisiones, no filtra credenciales; cierre sin viewer conserva evidencia, confirmación permite cleanup, reapertura descarta copia cerrada anterior. Revisar todos los consumidores/productores/global/skills/site/arquitectura/detectores y anotar delta de tamaño global al publicar, como exige AGENTS. La regla de marcar cierre procede de decisión explícita del usuario; no se ejecutan pilots de modelos.

La ingesta de un manifiesto modificado por status conserva el pin y el comportamiento actual de reinstalación: ingested_at se actualiza como en otros cambios de contenido. No introducir comparación especial de contenido ignorando status ni una política de retención nueva para esa transición. Actualizar scan.go además del resumen/índice para publicar status inmediatamente y al reconstruir. La retención de la unidad solo se extiende mientras siga siendo necesaria para confirmar cierre; se reporta ese pendiente.
