# Diseño en análisis

## Diseño vigente después de D11-A y D12-B

Este bloque sustituye el contrato de lectura viva descrito en el historial inferior. La primera implementación y su regresión se conservan como evidencia, no como camino operativo paralelo.

El manager toma las dos referencias fijas del payload validado, comprueba propietario/presencia/UTF-8 y genera sus textos, IDs y digests como datos del módulo JS instalado. Serializar como datos con `encoding/json`; no concatenar Markdown como sintaxis ejecutable. La fuente editorial JS es una plantilla, las referencias siguen teniendo su único hogar bajo sus skills y las copias son derivadas. El plugin no reabre referencias ni metadata por petición. Cada callback `context` suministra el snapshot completo del rol desde esos datos y deduplica solo sus propias partes.

Contrato cerrado de generación: la plantilla JS contiene exactamente una aparición del identificador de datos `__HIVE_CONTEXT_SNAPSHOT_DATA__` en el inicializador del snapshot. El manager comprueba cardinalidad uno antes de sustituir; cero o varias apariciones rechazan la generación antes de escribir. No volver a buscar el marcador en el texto insertado. Serializar un objeto con `schema: 2` y `references`, lista de las dos identidades de rol conocidas en orden lexicográfico ascendente por `role` (`hive-review-harness`, `hive-review-plan`). Cada elemento contiene `role`, `resource`, `text` y `sha256`; `text` representa exactamente los bytes UTF-8 válidos del recurso en `Release`, sin normalizar saltos, BOM ni Markdown; `sha256` hexadecimal minúsculo cubre esos mismos bytes. `resource` es la identidad lógica dentro del skill, nunca una ruta runtime aportada por el prompt.

El módulo instalado contiene ese objeto como datos JSON en una única sustitución. La metadata acompañante contiene `schema: 2` y la misma lista ordenada de `role`, `resource` y `sha256`, sin texto ni raíz de lectura. Validación de saved-plan reconstruye ambos derivados desde el payload, compara bytes completos y rechaza texto/digest/identidad u orden adulterados incluso con hashes de plan recalculados. Las pruebas contrastan ida/vuelta por ejecución del módulo con dobles y textos que incluyan comillas, barras, saltos, Unicode y el propio nombre del marcador dentro de los datos. La fuente JSON no se edita ni se mantiene como un manifest editorial adicional.

Mantener destinos inactivos actuales bajo `<OpenCodeHome>/hive/`, fuera de autodiscovery, con metadata acompañante de procedencia/digests generada por el manager y no consumida para resolver rutas en runtime. Esquema propuesto 2 para esa metadata: lista cerrada de rol, recurso y digest, sin `skillsDirectory` como capacidad de lectura. Recalcular datos desde `Plan.Config` y `Release` al validar un saved plan antes de escrituras; el JS renderizado y la metadata deben ser la misma decisión derivada. No hay una segunda fuente de edición ni fallback a archivos vivos.

Actualizar/regenerar el complemento y recargarlo cambia su snapshot; editar a mano un Markdown ya instalado no lo cambia. Documentar esa cadencia explícitamente. Conservar compatibilidad de snapshots antiguos que no contienen complemento; no mantener un cargador vivo y otro de snapshot simultáneamente. El layout del par no cambia, y la retirada manual de una entrada native ajena antes de volver a una release sin complemento sigue vigente. La regresión de sustitución de ancestros se reencuadra para demostrar que la petición no abre/reread archivos y no suministra la víctima ficticia; conservar la evidencia RED del diseño anterior antes de reemplazar su cobertura obsoleta.

La confianza de código JS local sigue siendo la del cargador native de plugins; un actor que pueda sustituir el propio plugin ejecutable no es una frontera que Hive prometa aislar. La garantía aquí elimina la lectura de referencias mutable por pathname durante la petición, sin prometer autenticación criptográfica ni sandbox universal.

D12-B mantiene Engram en todas las variantes: no cambiar a un piloto sin memoria para desbloquearlo. Falta acreditar un mecanismo soportado que vincule exportación y borrado exacto al mismo store frente a sustitución concurrente de DB/sidecars. Comprobar la versión instalada y código/API aplicables; un fence antes/después o verificar el directorio por nombre no se presenta como atomicidad. No editar el proyecto Engram ni limpiar memoria humana para investigar. Si no existe capacidad suficiente, conservar registros/store, declarar cleanup no verificado y presentar la incertidumbre al usuario. La fase de planificación está bloqueada por esa dependencia; no moverla a una tarea que finja resolverla después del piloto.

La revisión de este contrato y la investigación de memoria son read-only y no autorizan reanudar T6. La autorización de construcción D10-A y la entrega `hold` permanecen; los cambios dependientes del snapshot esperarán la revisión del diseño actualizado.

## Entrega del complemento como directorio (D21-A)

Sustituye, en el bloque D11-A y en "Distribución y activación", la forma de entrega del par inactivo; el contenido generado, su validación y la metadata 2 no cambian.

**Evidencia (observación en OpenCode 2.0.27, 2026-10-12).** Una entrada de `plugins` que apunta a un archivo `.js` se rechaza con WARN "configured plugin path must be a directory" y no carga. Un directorio con `index.js`, con o sin `package.json`, carga (`loading plugin`, sin WARN/ERROR) y su `Plugin.define({ id, setup })` se ejecuta; `@opencode/plugin` se resuelve sin `node_modules`; `ctx.session.hook` existe. La forma objeto `{ "package": "<directorio>", "options": {...} }` también carga y entrega `options`. Un archivo dentro de `<config>/plugins/` se autodescubre, lo que activaría el complemento sin decisión del usuario (contrario a AC6). La documentación V2 consultada muestra rutas de archivo como válidas; la conducta observada manda. No se verificó aún la inyección del hook `context` antes del despacho: la cubre el brazo C de T6.

**Contrato.** El manager entrega el par inactivo como un directorio fuera del autodescubrimiento: `<OpenCodeHome>/hive/context-loader/index.js` (script generado) y `<OpenCodeHome>/hive/context-loader/context-loader.json` (metadata, 0600). Sin `package.json`: no es necesario para cargar y añadiría un tercer recurso gestionado. La activación manual documentada es añadir `"<OpenCodeHome>/hive/context-loader"` a `plugins`. Propiedad, drift, reinstalación de ausentes, remoción, retorno a una release anterior y poda del directorio usan los mecanismos existentes, que trabajan por recurso y por directorio creado (`apply.go` `missingDirs`/`cleanupDirs`). Ninguna release publicada entregó `hive/context-loader.js` (`v0.3.3` y `development` no lo contienen), así que no hay migración de instalaciones; estados locales de candidatos no publicados no se aceptan con los nombres viejos.

**Piloto.** El observador se instala como `<config>/hive/pilot-observer/index.js` y se referencia con la forma objeto; el brazo C referencia el directorio del cargador. Los hashes de cargador y observador se toman de sus `index.js`. `ModelObserved` de nivel superior se asigna desde la exportación nativa del primario (`PrimaryCompletion.Model`), porque el stdout JSON de OpenCode 2 no trae el modelo; la comparación con el modelo pedido y el observador HTTP siguen exigiéndose.

**Inspección de operaciones tras el lote.** El gate estricto del runner marca opaca toda herramienta distinta de `subagent` (`operation_safety.go:26-28`), así que lecturas, `glob`, llamadas MCP de memoria y `execute` dejan la corrida en `operation_inspection_required`. No se relaja el gate: el `Terminal` automático se conserva y cada resultado registra además `ProtocolTerminal`, el veredicto de las comprobaciones de protocolo antes de ese gate. Tras el lote, el padre inspecciona cada corrida con `ProtocolTerminal=completed` y `Terminal=operation_inspection_required` usando las operaciones registradas (traza, primario e hijos) y el inventario final. Acepta solo si todas las operaciones son: herramientas integradas de lectura (`read`, `glob`, `grep`, `list`) sobre rutas dentro de la raíz de la fixture o del hogar ficticio; `subagent`; herramientas MCP de Engram sobre el store propio de la corrida con enrutamiento verificado; o `shell`/`execute` cuyo texto de comando o código registrado sea solo lectura (sin escrituras, sin `git init/commit/push`, sin red fuera del proveedor) y sin cambios en el inventario. Si el texto de una operación opaca no está registrado, la inspección la rechaza. `unauthorized_write` nunca es aceptable. Cada decisión (aceptada/rechazada, operaciones y motivo) se registra por corrida en `tasks.md`. Una corrida cuenta como válida para AC8 si `Terminal=completed`, o si `ProtocolTerminal=completed` y la inspección la acepta; el reporte muestra ambos valores sin reescribir el automático.

**Texto de comandos y modo código (D22-A).** El runner persiste el texto de una operación `shell` solo cuando reconoce una forma de lectura acotada (un único comando simple de una lista corta de lectura, sin metacaracteres de shell, redirecciones, sustituciones ni asignaciones de entorno); cualquier otro texto y todo código de `execute` se retiene (`command_withheld`) y la inspección rechaza esa operación. Una lista de patrones prohibidos no establece la frontera y no se usa como criterio de persistencia. Para que la memoria sea inspeccionable sin texto, el servidor Engram de cada hogar ficticio se configura con `codemode: false` en las tres variantes (documentado en OpenCode V2 "MCP servers"), de modo que sus herramientas aparecen con nombre propio en lugar de dentro de `execute`; en el preflight, `execute` envolvía `engram.mem_save`/`mem_session_summary`. Es una diferencia respecto al uso cotidiano con modo código, idéntica para A, B y C.

## Piloto en bb1 después de D14–D18

Este bloque sustituye, solo para el piloto, el requisito de D12-B de borrar registros exactos con un binding acreditado. Engram se mantiene en todas las variantes; cambia dónde vive el store y qué significa limpiarlo. Fuente: investigación read-only del 2026-10-12 y código contrastado (`tasks.md`, Progreso).

**Topología.** El runner `tests/pilot` se compila en el Mac para `linux/amd64` desde el candidato y corre dentro de un contenedor en `bb1`, porque lanza `opencode`, `engram` y grupos de procesos con `exec` local; no se maneja desde el Mac con `docker exec`. Por lote, un contenedor `hive-pilot-<lote>` con `--init`, usuario no root, sin puertos publicados, sin bind mounts de homes del Mac ni de bb1, sin `docker.sock`, red bridge y salida HTTPS. Etiqueta `hive.pilot=flash-class-guidance` en todo recurso creado; nunca `docker prune` ni operaciones sobre recursos sin esa etiqueta, porque bb1 aloja volúmenes de otros laboratorios.

```
Mac: runner linux/amd64 + tar del candidato/baseline/fixtures --ssh--> bb1
bb1: imagen hive-pilot:<digest> (conservada)
     contenedor hive-pilot-<lote>
       tmpfs  ~/.local/share/opencode   <- login del usuario (D17)
       volumen hive-pilot-work-<lote>   <- runs/*/engram-data, evidencia
       runner -> opencode run --standalone -> engram mcp (stdio)
```

**Imagen (D18-A, conservada).** Debian slim fijada por digest, con `procps`, `git` y certificados; Engram 3.3.2 desde el artefacto oficial `linux_amd64` comprobado contra el archivo de checksums de la release; OpenCode 2.0.27 desde el paquete npm que usa su instalador V2 (`@opencode/cli-linux-x64@2.0.27`, observado por la investigación en el script de `https://opencode.ai/v2/install`; un revisor no pudo confirmarlo), con la integridad del tarball fijada, no ejecutando el script del instalador sin fijar. T8 confirma primero que el paquete y la versión existen; si no, se detiene y se decide otra fuente. Comprobar `opencode --version` y `engram version` en la imagen y registrar su digest. Se conservan la imagen etiquetada y un volumen de caché `hive-pilot-cache` para reutilizar en pilotos futuros. No se usa la imagen Alpine oficial porque su `ps` puede no aceptar `-A -o pgid=,stat=`, ni la imagen `ghcr.io/gentleman-programming/engram`, que es el servidor cloud.

**Proveedor (D15).** Los cuatro modelos de D7-B se piden por `opencode-go/` con las mismas variantes: `opencode-go/claude-haiku-5-5#high`, `opencode-go/deepseek-v4.1-flash#high`, `opencode-go/qwen3.8-flash#medium`, `opencode-go/glm-5.3-flash#high`. Catálogo consultado 2026-10-12; no prueba auth, cuota ni variantes. La suscripción Go es la cuota compartida del usuario: el 2026-10-12 a las 15:35 marcaba 41 % semanal usado, 44 % mensual y uso extra apagado. Agotar un límite bloquea la cobertura restante de ese modelo; no activar uso extra, no comprar créditos ni sustituir modelo o proveedor.

**Credencial (D17).** No se exporta la credencial del Mac. Al iniciar el lote, el usuario ejecuta `opencode auth login opencode` dentro del contenedor, sobre un tmpfs `/run/hive-pilot/secrets` (dueño UID/GID del usuario no root, modo 0700) que hace de "instalación real" del runner. La importación existente del runner (`auth export` → `auth import --standalone`) escribe hoy en `<out>/shadow-home/.local/share/opencode` (`guidance_variant.go:690,745`), que estaría en el volumen de trabajo; en el modo contenedor, el directorio de datos nativos de cada hogar ficticio se enruta a un subdirectorio propio de ese tmpfs, y el resto de la salida queda en el volumen. Antes de copiar evidencia, comprobar que no hay bases `opencode.db` ni archivos de credencial bajo el volumen ni en la evidencia seleccionada. Al terminar, `opencode auth logout` dentro del contenedor si el comando existe; eso es cierre local, no revocación en el proveedor, que no está verificada. La eliminación del contenedor descarta el tmpfs. Docker documenta que un tmpfs puede llegar a swap del host: "sin disco" significa sin archivos persistentes en volúmenes o capas, no garantía contra swap. Ninguna credencial en variables de entorno, capas de imagen, volúmenes ni evidencia.

**Configuración nativa mínima.** El runner exige una configuración "real" con `mcp` de Engram antes de importar credenciales (`guidance_variant.go:473-526,739`). T8 escribe en la instalación real del contenedor solo un `opencode.json` con el servidor local `["engram","mcp"]`, sin otras entradas.

**Memoria (D16, D18-A).** Cada corrida conserva su `ENGRAM_DATA_DIR` fresco y verificado vacío dentro de `hive-pilot-work-<lote>`, con `ENGRAM_NO_UPDATE_CHECK=1`. Distinguir dos entornos: el **heredado** del contenedor, que no puede contener `ENGRAM_CLOUD_*`, `ENGRAM_URL`, `ENGRAM_PORT` ni `ENGRAM_DATA_DIR` y se valida antes de cualquier filtrado o escritura; y el **generado** por el runner para el store propio de la corrida (`ENGRAM_DATA_DIR`, `ENGRAM_CLOUD_AUTOSYNC=0`, `ENGRAM_PORT`/`ENGRAM_URL` del servidor propio, `memory.go:70,174`), que sí se permite. El data dir se resuelve sin seguir enlaces simbólicos que escapen de la raíz de la corrida. La limpieza es: exportar por corrida para evidencia, copiar la evidencia seleccionada al Mac, comprobar sumas y destruir el contenedor y el volumen de trabajo según "Lote y recuperación". El runner necesita un modo explícito de store desechable en contenedor: sin él, `flash_memory.go:200` devuelve siempre "atomic memory store binding unavailable" (F1), y `main.go:614-626,800-826` traduce un `finish` exitoso en `verified_clean` y llama a `memory.cleanup()` de inmediato. En el modo, `finish` exporta y reporta `deferred_to_volume_destruction`, no llama a la limpieza inmediata y no afirma haber verificado el store cotidiano, porque dentro del contenedor esa comprobación exporta un store vacío (F2). La marca de contenedor no prueba aislamiento por sí sola; lo acredita la inspección del lanzador. Fuera del contenedor, el modo se rechaza.

**Lote y recuperación.** Antes de crear nada, el lanzador genera un ID de lote y un token de propiedad aleatorio, comprueba que no existe ningún contenedor ni volumen con los nombres del lote y escribe en el Mac un manifiesto con el ID, los nombres, el token y las etiquetas `hive.pilot=flash-class-guidance`, `hive.pilot.batch=<lote>` y `hive.pilot.owner=<token>`. Un nombre ya existente aborta. Después de crear cada recurso, añade al manifiesto su identidad inspeccionada: ID del contenedor y, para el volumen, nombre más `CreatedAt` y etiquetas tal como los devuelve `docker volume inspect`. Destruir solo recursos cuyo nombre y etiquetas (incluido el token) coincidan con el manifiesto, y cuyo ID coincida cuando el manifiesto lo tiene; nunca `prune`. Un manifiesto parcial, por interrupción entre la creación y el registro, se recupera por nombre exacto más etiquetas de lote y token. `hive-pilot-cache` es compartido entre lotes y no se destruye; su contenido se limita a binarios descargados y catálogo de modelos, sin credenciales ni stores. Un lote interrumpido (caída del contenedor, del SSH o del Mac) se recupera con un subcomando que copia la evidencia existente con sumas, hace logout si el contenedor sigue vivo y destruye contenedor y volumen de trabajo tras la comprobación de propiedad; no reintenta corridas de modelos. Si el supervisor de procesos no puede verificar una parada en Linux (`?` propio o `D` persistente), el runner deja de programar corridas, conserva evidencia y la destrucción queda pendiente hasta verificarla.

**Comprobación en el Mac (D16).** El runner nombra cada proyecto de memoria `flash-class-<digest>` a partir de la raíz canónica de la fixture (`main.go:466`), así que los nombres no se conocen antes. Antes del lote, el padre lista en solo lectura los proyectos del Engram cotidiano del Mac y exige cero proyectos con prefijo `flash-class-`. Después, recoge del manifiesto de resultados los IDs de proyecto exactos de cada corrida, comprueba que ninguno existe en el Mac y repite el conteo por prefijo. El comando de lectura se fija en T8 partiendo del que el runner ya usa, `engram projects list` (`memory.go:132`), confirmando en Engram 3.3.2 su salida, el emparejamiento por prefijo y la búsqueda exacta de un proyecto; una salida que no se pueda interpretar deja la comprobación sin verificar y el lote no empieza. Un proyecto nuevo marca el lote como contaminado y detiene el resto. Esto prueba ausencia en ese store, no en logs del proveedor ni en otros stores.

**Catálogo (F4).** En el contenedor, el catálogo `models.json` se refresca con `opencode models` después del login, antes de la primera corrida; los cuatro modelos deben aparecer o el lote no empieza. No copiar la caché del Mac.

**Qué deja de probarse y qué no cambia.** No se prueba el borrado exacto por ID ni la ausencia de sustitución del directorio de datos durante la corrida: la destrucción del store completo los vuelve innecesarios. Que el store cotidiano no se tocó queda por construcción más el conteo del Mac. El observador JS, el transporte HTTP observable, las 48 corridas, el paralelismo máximo de dos, el timeout de 3 minutos y la ausencia de reintentos no cambian. El supervisor de procesos corre por primera vez en Linux real: un estado `D` o `?` del grupo propio deja la parada sin verificar.

**Cláusulas sustituidas.** Para el piloto quedan sustituidas: en "Piloto autorizado (D7-B)", los IDs `opencode/` por los `opencode-go/` de este bloque, "sin nuevo login" por el login del usuario dentro del contenedor, y el directorio Engram "aislado" del Mac por el store del volumen; en "Reconciliación de revisión final", el último párrafo sobre limpiar registros ficticios por mecanismo soportado, por la destrucción del volumen; en D12-B, el requisito de binding para borrar. El resto de esos apartados sigue vigente.

**Autorización.** D14-A es solo diseño. Antes de ejecutar se pide aparte autorización para: copiar runner y archivos a bb1, construir la imagen (descargas), crear contenedor y volúmenes, el login del usuario y las corridas de modelos con la cuota Go.

## Contexto verificado al planificar (historial)

- `feat/guidance-rework` contiene ocho commits no presentes en la base local `development`, hasta `d5164f71`. `git diff development...feat/guidance-rework` muestra cambios en guía global, flujos, referencias, sitio y pruebas de tamaño. No existe PR de esa rama según la consulta a GitHub del 2026-10-11.
- El checkout de la sesión está en `feat/pi-subagents-unversioned-install`, que tiene cambios propios de Pi; no representa la referencia elegida para esta guía.
- En la referencia elegida, `tests/content/budget_test.go` informa bytes y caracteres sin techo. El margen de cuatro bytes mencionado en el research correspondía a otra versión y no define el diseño nuevo.
- [instruction-resources.md](../../../../docs/architecture/instruction-resources.md) define recursos canónicos por skill, condiciones de lectura, enlaces `skill:owner/path` y reescritura a destinos instalados. Leer una referencia no invoca todo su skill.
- `tooling/management/plan.go:131–263`: `loadRelease` descubre todos los entrypoints y `collectResources` recoge `references/`, `scripts/` y `assets/`. `validSource` acepta ese esquema; un directorio separado de módulos de guía no está contemplado como fuente genérica.
- `integrations/target/catalog.go:9–43`: `ExpandSkills` lleva entrypoints y recursos a los destinos existentes. `tooling/management/skill_links_test.go:61–111` cubre enlaces renderizados para seis hosts a nivel usuario y Claude/Codex a nivel proyecto. Son pruebas en hogares ficticios, no conducta de modelos.
- [repository-and-distribution.md](../../../../docs/architecture/repository-and-distribution.md#core-references-and-optional-native-supply) identifica ahora como histórica la comparación de hooks con el bloque global. La versión consultada al planificar, en `d5164f71`, comparaba límites, subagentes, orden y desactivación con revisión del 2026-09-24 y nota de plugins OpenCode del 2026-09-30; por eso se contrastaron después las capacidades actuales, sin usar aquella tabla como prohibición vigente.
- La misma fuente afirma que el router y las referencias llegaron al modelo en aproximadamente 45–70 % de las sesiones medidas. Es una afirmación histórica de la fuente, sin población ni criterio suficientes aquí para convertirla en tasa actual. Es contraevidencia relevante frente a asumir lectura perfecta por punteros.
- [agent-delivery.md](../../../../docs/architecture/agent-delivery.md) distingue entrega nativa, catálogo, lectura y resultado; describe controles por host y sus límites. No demuestra que un hook sirva para todos los hijos ni que una lectura ocurra cuando se necesita.

## Distinciones que gobiernan la comparación

1. **Descubrimiento:** el agente sabe qué recurso existe y cuándo necesita usarlo.
2. **Carga:** el texto correcto entra realmente en el contexto antes de la acción que gobierna.
3. **Cumplimiento:** el modelo actúa según ese texto; una herramienta que lo cargó no lo demuestra.
4. **Control de efectos:** el host puede bloquear una acción. No equivale a completar una regla de comunicación.

Comparar también un prompt directo del usuario, una delegación, una continuación tras resumir contexto, y ejecución no interactiva. Una solución que solo cubre el inicio del hilo principal deja superficies sin resolver.

## Alternativas a contrastar

| Alternativa | Beneficio propuesto | Riesgo o coste | Evidencia necesaria |
| --- | --- | --- | --- |
| Núcleo con referencias explícitas | Reutiliza la distribución actual y evita integración nueva | El modelo puede omitir o escoger mal una lectura | Rutas completas, disparadores y procedencia de mediciones históricas |
| Hooks nativos | Pueden recordar o inyectar contenido en un evento definido | Cobertura, truncado, prioridad, permisos y desactivación varían | Documentación actual por host y versión instalada; no extrapolar de un host |
| Cargador o plugin del host | Puede seleccionar e insertar contenido automáticamente | Predicados de selección, caché, coste, mantenimiento y recuperación | Qué control nativo existe, qué puede observar y qué queda fuera |

No se presupone un cargador universal ni una tabla de implementaciones obligatoria por host. Un loader que entrega todo en cada petición no reduce densidad. Uno que invoca un modelo para clasificar tareas cambia coste y arquitectura y no está autorizado dentro de este análisis.

La regla de proporcionalidad del repositorio exige una necesidad concreta antes de introducir mecanismos. Tampoco autoriza descartar un mecanismo útil solo por preferir texto. La elección se hará con beneficios y límites observados o documentados, distinguiendo hipótesis por comprobar.

## Contratos que deben preservarse

- Seguridad, permisos, secretos y preservación no dependen de que el clasificador elija un módulo.
- Cada obligación tiene un hogar y una ruta de acceso explícita; si no se puede leer un recurso requerido, se informa y se detiene solo la acción dependiente.
- La parte básica de ubicación de archivos sigue siempre visible; el detalle de taxonomía y registros puede ser situacional.
- Los protocolos recientes de preguntas y rutas iterativas de `d5164f71` siguen vigentes.
- Los hijos reciben sus recursos explícitamente, sin suponer que heredan una carga hecha por el padre.
- Conservar los archivos nativos y los bytes ajenos a Hive, permisos del host, recuperación y consumidores compartidos. No editar globalmente durante el plan.
- No confundir menor global con menor contexto activo: comparar la unión de recursos necesarios por recorrido, y declarar metadata de skills/roles y otros costes fuera del cálculo si no se miden.

## Resultado de la revisión preliminar

Revisores nativos `hive-review-plan`, con `model: openai/gpt-6.1-sol` solicitado en ambos lanzamientos; no se ha inspeccionado metadata que acredite su modelo efectivo. Candidato revisado: SHA256 agregado de `proposal.md`, `design.md`, `tasks.md`, en ese orden con separadores nulos, `5604059ffe11ec4077c5d8754edf5e2b3b770a5e22bdeffe6bc2d98e4de408c0`. El padre contrastó los hallazgos de validación con código y las afirmaciones decisivas de carga con documentación oficial.

### Capacidades nativas

El revisor `ses_ed21ae2f3ffeGglM5NZ3ZQnkD2` observó versiones locales con `--version` y consultó documentación oficial el 2026-10-11. Las versiones no son una prueba de funcionamiento de hooks; los docs vivos no están fijados al binario local.

| Host · versión informada | Carga relevante documentada | Límite material | Fuente oficial |
| --- | --- | --- | --- |
| Claude Code · 2.1.296 | Hooks de prompt/inicio/hijo; `skills` precarga texto completo en roles | El contexto de `PreToolUse` acompaña el resultado; no gobierna retrospectivamente la operación | [Hooks](https://code.claude.com/docs/en/hooks), [subagentes](https://code.claude.com/docs/en/sub-agents) |
| Codex · 0.162.1 | Hooks de inicio, prompt e hijos; entrada explícita de skills | Límites por mensaje y confianza; `exec` y compactación de hijos no verificados | [Hooks](https://learn.chatgpt.com/docs/hooks), [app-server](https://learn.chatgpt.com/docs/app-server) |
| Grok · 1.0.50 | Contexto asociado a pre/post herramienta y respuesta de parada | Inicio, hijo y compactación pasivos; stdout permisivo de prompt no incorpora contexto | [Hooks](https://github.com/xai-org/grok-build/blob/main/crates/codegen/xai-grok-pager/docs/user-guide/10-hooks.md) |
| Pi · 1.1.0 | Extensión antes de petición y expansión explícita de skills | Herencia de extensiones en pi-subagents no verificada | [Extensiones](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/extensions.md), [skills](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/skills.md) |
| OpenCode V2 · 2.0.26 | Plugin de sesión `prompt`/`context`; puede modificar entrada y contexto de peticiones | Cobertura depende de ubicación y plugins activos; hijos/no-interactivo no observados | [Plugins](https://opencode.ai/v2/docs/build/plugins) |
| Cursor CLI · 2026.10.01-e373342 | Contexto en inicio y postherramienta | Inicio no espera; subagentStart controla permiso, no carga equivalente; cobertura incompleta | [Hooks](https://cursor.com/docs/hooks), [CLI](https://cursor.com/docs/cli/changelog) |

Verificación directa del padre: Context7 CLI 0.5.11, resolución de Claude Code y consulta `/websites/code_claude` sobre momento de entrega. Como no resolvió esa distinción, se contrastó el documento oficial: `PreToolUse.additionalContext` se añade «alongside the tool result». Los docs de subagentes confirman la inyección del contenido completo de los skills enumerados al inicio. Los docs oficiales de OpenCode describen el hook de admisión previo a resolución de archivos/skills y los hooks de contexto de sesión; Grok declara eventos pasivos y el descarte de stdout permisivo de `UserPromptSubmit`. No se ejecutaron esos mecanismos.

El fallo de acceso a `firecrawl-scrape/SKILL.md` impidió usar su procedimiento; las comprobaciones del padre y el revisor recurrieron a HTTP de solo lectura sobre fuentes oficiales. Context7 fue el primer lookup. No se configuró autenticación ni se copiaron respuestas externas completas al plan.

### Hallazgos y reconciliación

- **Distribución H1, alta — validación del núcleo:** confirmado `references.go:18–20` excluye `GlobalSource`; `mdlinks.go:80–100` reescribe sin verificar existencia. `plan.go:224,346,709` aplica el validador tanto al preparar como al validar un plan guardado. Incorporar en la futura descomposición el rechazo de dependencias globales rotas antes de escribir. Existe el patrón `TestApplyRejectsRehashedReleaseWithMissingDependency` en `references_test.go:43–59`; ampliarlo al núcleo, sin crear un sistema de validación aparte. El test `tests/content/skill_links_test.go` no reemplaza la validación de un payload durante instalación.
- **Distribución H2, alta — selección insuficiente:** aceptado. Preguntas, respuestas y rutas pueden suceder sin herramienta/archivo. Clasificar obligaciones por siempre visibles, seleccionables con señales observables, e interpretación semántica del modelo. No delegar autorización a un selector de módulos.
- **Distribución H3, media — lifecycle:** aceptado como condición si se adopta integración activa. Empaquetar un script no registra un hook. Definir propiedad, configuración nativa, drift, desactivación, remoción, downgrade e interrupción; reutilizar transacciones donde sirvan sin afirmar cobertura existente.
- **Capacidades H1, alta — contexto tardío:** confirmado para Claude con documentación oficial; el revisor informa equivalencia en Grok. No usar contexto permisivo de preherramienta como control preventivo de esa acción. Bloqueo y aprendizaje para la siguiente petición son contratos distintos.
- **Capacidades H2, media — documentación histórica:** aceptado. La tabla local de septiembre no basta para descartar las capacidades actuales. Claude documenta hoy un archivo con vista previa al exceder 10.000 caracteres por cadena, no el truncado silencioso que decía la fuente antigua. La cifra histórica 45–70 % no es reproducible con las fuentes disponibles y no se usará como aceptación ni como tasa actual.
- **Capacidades H3, alta — ausencia de selector de intención:** aceptado. Ningún evento consultado selecciona por sí solo el procedimiento correcto ante texto libre. Un cargador invocado por el modelo mejora acceso; una inyección por rol explícito puede mejorar carga, con cobertura acotada. No afirmar que uno sustituye todo descubrimiento.

### Criterios de elección

Un complemento automático solo aporta una ventaja concreta si una señal verificable elige texto canónico necesario y lo introduce antes de la acción relevante. Repetir rutas no elimina la lectura pendiente; inyectar todos los procedimientos recrea la densidad. La selección por rol o skill explícito es más verificable que inferir intención del prompt.

La recomendación preliminar es núcleo y referencias para todos, sin hacer seis integraciones nuevas un requisito. Si el usuario quiere un complemento en este mismo trabajo, una prueba técnica acotada de carga automática en un host puede acreditar selección/texto/momento en esa superficie, no mejor conducta del piso objetivo. Requiere decidir host y cobertura, y no registrar el complemento globalmente durante `hold`.

## Diseño elegido: núcleo portable y complemento opcional

D6-B incluye el complemento OpenCode y D7-B autoriza el piloto acotado después de construir. No se ha implementado ninguno ni impuesto un techo arbitrario de tokens.

### Reparto común

El núcleo mantiene idioma, resultado esperado, alcance y efectos autorizados, preservación, secretos, evidencia, proporcionalidad, ubicación básica, decisiones esenciales de preguntas y las rutas explícitas de lectura. No depende del plugin para proteger una acción.

Destinos propuestos de los procedimientos, todos bajo el mecanismo actual de skills:

| Procedimiento que se mueve o divide | Hogar canónico | Leer antes de |
| --- | --- | --- |
| Detalle de preguntas y tarjeta, con ejemplos | `flow-plan/references/questions.md` | Formular una decisión al usuario, con o sin tarjeta |
| Detalle de reportes y citas, con ejemplos | `flow-close/references/reporting.md` | Emitir un reporte, o cualquier mensaje/pregunta/retransmisión que incluya citas, enlaces o IDs definidos fuera de él |
| Asignación, entrega de referencias y espera | `flow-research/references/delegation.md` | Lanzar un hijo o una espera de trabajo externo |
| Taxonomía, nombres, fechas y retención de soporte | `workspace-conventions/references/artifact-layout.md` | Escribir o limpiar un artefacto de soporte; la ubicación básica queda global |
| Resolución de settings y registros de cambio | `flow-plan/references/project-settings.md` y `change-records.md` existente | Necesitar esos settings o crear/continuar el registro correspondiente |

La matriz completa se termina con el texto candidato en T1: una fila por obligación trasladada, referencia al fragmento en `d5164f71`, destino, disparador y consumidores. El implementador no puede eliminar una regla para ganar tamaño. Mantener condiciones actuales y registrar cualquier necesidad de una decisión nueva antes de editarla.

Invariantes de citas que permanecen globales: preservar identificadores y evidencia literal; explicar brevemente cualquier ID definido fuera del mensaje; usar enlaces con etiqueta descriptiva y URL verificada, enlazar un ticket solo en su primera mención del mensaje y citarlo luego por ID. La referencia de reporting conserva los detalles de glosas, rangos, tarjetas y formatos; su disparador incluye preguntas, respuestas ordinarias y retransmisiones, no solo reportes. La referencia de preguntas enlaza directamente a esos detalles donde los consume, sin precargar todo `flow-close`. Esta decisión conserva el alcance de `global.md:16,18,21–24` de la base y no elimina micro-reglas por conveniencia.

Los skills consumidores enlazan directamente al recurso que necesitan; no exigen abrir todo un skill intermediario para encontrarlo. Se conserva el dialecto `skill:owner/path` en el núcleo/roles y enlaces relativos en skills. Las reglas básicas de autorización y ubicación no pasan a una actividad opcional.

### Matriz de obligaciones trasladadas

Referencia editorial: `d5164f71:content/guidance/global.md`, líneas indicadas abajo. Base de edición: `a4d6e254`; el núcleo tiene los mismos 46.912 bytes que la referencia editorial. Las mejoras integradas de reutilizar revisión sin entrega Git y selección provisional ante exceso de opciones se conservan en `flow-build/SKILL.md` y `flow-plan/references/delivery-decisions.md`. No se modificaron decisiones, estados ni marcadores de tareas.

Destinos, todos relativos a `content/skills/`:

- **Preguntas:** `flow-plan/references/questions.md`.
- **Reportes:** `flow-close/references/reporting.md`.
- **Delegación:** `flow-research/references/delegation.md`.
- **Artefactos:** `workspace-conventions/references/artifact-layout.md`.
- **Settings:** `flow-plan/references/project-settings.md`.
- **Registros:** `flow-plan/references/change-records.md`.
- **Cierre:** `flow-close/SKILL.md`, sección `Close question`, enlazable por nombre y anchor sin depender del paso 7.

Los nombres de destino en cada fila remiten a estos archivos y a sus secciones nombradas. «Todos» incluye hilo principal y roles, con lectura condicional explícita desde el núcleo. Los consumidores de actividad enlazan directamente a los recursos que necesitan; leerlos no invoca el flujo completo. Las condiciones siguientes preceden la acción, no la justifican retrospectivamente. Reporting se lee antes de toda respuesta de chat, también una explicación sin citas: así las reglas anteriores de formato de L19 no quedan sin ruta por carecer de enlaces o IDs.

| Obligación y fragmento de origen | Destino y sección | Condición de lectura | Consumidores |
| --- | --- | --- | --- |
| Abrir con resultado, prioridad de trabajo activo/completado/respuesta; detalle proporcional (L16) | Reportes · Lead with the result; invariante global | Antes de reporte | Todos; flows plan/build/research/close |
| Prosa natural, contexto suficiente, explicar términos sin vocabulario interno (L16) | Reportes · Lead with the result | Antes de reporte | Todos |
| Evidencia material, no cada llamada (L16) | Reportes · Lead with the result; invariante global | Antes de reporte | Todos |
| Enlace Markdown etiquetado, contexto conocido sin consulta solo por título (L16) | Reportes · Citations and links | Antes de cualquier mensaje con enlace | Todos, incluso respuestas ordinarias y tarjetas |
| URL de tracker devuelta/verificada, forma estable corta, no inventar dominio (L16) | Reportes · Citations and links; invariante global | Antes de citar ticket con enlace | Todos |
| Ticket enlazado solo en primera mención del mensaje; sin URL, ID; siguientes por ID (L16) | Reportes · Citations and links; invariante global | Antes de citar ticket | Todos |
| Separar resultado/hallazgos/decisiones/acción con secciones breves (L17) | Reportes · Shape the message by its content | Antes de hallazgos multitema | Todos |
| Áreas de progreso/cierre con etiqueta bold autónoma en idioma de sesión (L17) | Reportes · Shape the message by its content | Antes de progreso/cierre | Todos |
| Si termina turno con trabajo activo, abrir por ejecución/qué decide/continuación (L17) | Reportes · Report a running or completed item | Antes de reporte con proceso activo | Todos; Delegación enlaza destino |
| Enlazar documento retenido sin copiarlo; alternativas y decisiones quedan en mensaje (L17) | Reportes · Shape the message by its content | Antes de reporte con registro | Plan/build/research/close y roles |
| Reportar PR nuevo o pendiente con enlace; runs cuando útiles (L17) | Reportes · Shape the message by its content | Antes de reporte de PR | Build/git-workflow/close |
| Dos listas numeradas requieren prefijos distintos y referencias por ID (L17) | Reportes · Citations and links | Antes de mensaje con listas múltiples | Todos |
| Glosa de todo ID externo, cada vez; una glosa para rango (L18) | Reportes · Citations and links; invariante global | Antes de cualquier mensaje con ID externo | Todos |
| Tarjeta es mensaje propio; glosar incluso ID definido en texto previo; opciones propias exentas (L18) | Reportes · Citations and links | Antes de tarjeta con ID externo | Preguntas enlaza directamente Reportes |
| Tabla solo ≥3 elementos/≥2 atributos/≤4 columnas; lista paralelos, prosa causas (L19) | Reportes · Shape the message by its content | Antes de reporte | Todos |
| Cada ticket o conjunto unido en línea propia con descripción; no ≥3 en párrafo (L19) | Reportes · Shape the message by its content | Antes de reporte con tickets | Todos |
| Formato de skill/template prevalece; estado escrito en idioma de sesión, no emoji ni adorno (L19) | Reportes · Shape the message by its content | Antes de reporte | Todos |
| ASCII ≤70 columnas al explicar ramificación, ≥3 capas, etapas o antes/después; brevedad no lo elimina (L19) | Reportes · Shape the message by its content | Antes de explicación que cumple condición | Todos |
| Secuencia corta lineal numerada; no Mermaid/SVG en chat (L19) | Reportes · Shape the message by its content | Antes de explicación/reporte | Todos |
| Oferta de página flow-report para comparación/explicación/reporte que se comparte o guarda (L19) | Reportes · Shape the message by its content | Antes de entrega compartible | Todos; flow-report |
| Preguntar solo decisión del usuario que cambia acción, buscar hechos, decidir llamada menor y decirlo (L21) | Preguntas · Decide whether a question is needed; invariante global | Antes de formular decisión | Todos |
| No permiso vacío para seguir/mostrar; agrupar preguntas independientes (L21) | Preguntas · Decide whether a question is needed | Antes de formular ronda | Todos |
| IDs pregunta/opción, efecto, Recommended primero con motivo; sí/no sin IDs opción (L21) | Preguntas · Decide whether a question is needed | Antes de formular decisión | Todos |
| Consecuencia: tradeoff e incertidumbre; no opciones solo en documento (L21) | Preguntas · Decide whether a question is needed | Antes de formular decisión | Todos |
| Reutilizar decisión; silencio/preselección/dismiss/timeout no autorizan (L21) | Preguntas · Decide whether a question is needed; invariante global | Antes de formular/reformular decisión | Todos |
| Usar tarjeta que espera; dividir exceso en llamadas consecutivas con texto previo (L22) | Preguntas · Use a waiting native card | Antes de preguntar con tarjeta | Todos; delivery-decisions conserva selección provisional |
| Tarjeta autónoma: contexto, efecto en descripción o etiqueta sin campo, Recommended (L22) | Preguntas · Use a waiting native card | Antes de tarjeta | Todos |
| Updates antes de mensaje; explicación/resumen por pregunta/recomendación glosada, sin otras opciones (L22) | Preguntas · Use a waiting native card | Antes de tarjeta | Todos |
| Sin herramienta intermedia entre mensaje y pregunta; llamada sola falla (L22) | Preguntas · Use a waiting native card | Antes de tarjeta | Todos |
| Sin tarjeta que espera: texto completo; async solo con trabajo independiente; dismiss/timeout una repetición textual (L22) | Preguntas · Use text when no waiting card carries the question | Antes de preguntar sin tarjeta que espera | Todos |
| Opciones una línea por diferencia; común/restricción una vez; sin Other ni ejemplo respuesta (L22) | Preguntas · Use text when no waiting card carries the question | Antes de pregunta textual | Todos |
| Primera respuesta prevalece; hechos de sesión frente a supuestos por confirmar (L22) | Preguntas · Use text… y Decide whether… | Antes de pregunta/respuesta | Todos |
| Una pausa final después de categorías, etiqueta fija traducida, ID glosado; antes de tarjeta (L23) | Preguntas · Pause and handle the answer; invariante global | Antes de mensaje que espera elección | Todos |
| Preguntar una vez; repetir solo dismiss/timeout (L23) | Preguntas · Pause and handle the answer | Antes de reformular elección | Todos |
| Responder antes de actuar con qué resuelve/reabre/invalida, incluso texto libre (L23) | Preguntas · Pause and handle the answer; invariante global | Al recibir respuesta | Todos |
| Reabrir respuesta solo efectos nuevos no autorizados (L23) | Preguntas · Pause and handle the answer | Antes de reabrir decisión | Todos |
| Solicitud autosuficiente: qué/unblocks, destino concreto/comparación, verificado y respuestas/efectos (L24) | Preguntas · Make user-dependent work self-sufficient | Antes de pedir revisión/validación/acción/dato del usuario | Todos; verification enlaza destino |
| Varias validaciones: bloque por flujo/ticket con pasos y resultado (L24) | Preguntas · Make user-dependent work self-sufficient | Antes de validación múltiple | Build/verification |
| Manual validation/corrección/stop antes push completo en mensaje y pausa; no segundo reporte tras pregunta pendiente (L24) | Preguntas · Make user-dependent work self-sufficient y Pause… | Antes de mensaje que espera usuario | Todos |
| Reporte completo antes de tarjeta; definir punto alcanzado/exclusión, PR/cleanup/incidente durable (L25) | Reportes · Report a running or completed item | Antes de cierre completado | Close/build/git-workflow |
| Pendiente impide claim completo; abrir por punto alcanzado y faltante (L25) | Reportes · Report a running or completed item; invariante global | Antes de reporte con pendiente | Todos; ejemplo resuelto |
| Categorías solo aplicables, sin inventar; cada ítem una sola categoría y orden fijo (L28) | Reportes · Closing categories | Antes de hallazgos/progreso/cierre y seguimiento final | Todos |
| Pending: obligación actual, no elección ni otro proyecto; no completo, dueño único (L29) | Reportes · Pending | Antes de reporte con pendiente | Todos |
| Pending: autosuficiente para usuario, agente qué/cuándo; verificar acceso/autoridad y hacerlo si posible (L29) | Reportes · Pending → Preguntas | Antes de asignar pendiente | Todos |
| Pending del usuario solo decisión/autoridad, acceso ausente comprobado o acción fuera máquina (L29) | Reportes · Pending | Antes de asignar pendiente al usuario | Todos |
| Next step: única ruta ya acordada, ejecutor y nueva autorización; no elegir por usuario (L30) | Reportes · Next step | Antes de reporte con continuación | Todos |
| Recommendations: varias rutas razonables, motivo, numeradas, Recommended cuando clara (L31) | Reportes · Recommendations | Antes de recomendación | Todos |
| Pedido qué seguir: elección con flow-plan para complejo/build para pequeño/ambos si aplican, excepción iterativa (L31) | Reportes · Recommendations | Antes de ofrecer siguiente ruta | Todos; flow-research mantiene ruta iterativa |
| Iniciar flujo elegido sin comando nuevo; elección no agrupa Git/deploy/tracker/secret effects; delivery antes de editar salvo grant existente (L31) | Reportes · Recommendations; autorización sigue global | Antes de ofrecer/ejecutar ruta | Todos |
| Investigación usa su pregunta de ruta (L31) | Reportes · Recommendations y Work-item boundary | Antes de cerrar findings research | Flow-research |
| Reminders: control usuario no bloqueante, promoción por lotes no retiene cierre, PR usuario (L32) | Reportes · Reminders | Antes de reminder | Todos |
| Reminder una línea sin consejo, solo repetir ante cambio/cierre relacionado; coste concreto lo convierte en recomendación (L32) | Reportes · Reminders | Antes de reminder | Todos |
| Separar hecho/continuación autorizada/nueva autorización; una pregunta por conjunto, riesgo/target propio aparte (L33) | Reportes · Closing categories | Antes de reporte/decisión de cierre | Todos |
| Work item vs tarea; progreso entre tareas con herramienta; contestar notificación una línea y continuar (L34) | Reportes · Work-item boundary; invariante global | Durante trabajo autorizado | Plan/build/research |
| Cierre termina con pregunta después de reporte/categorías; opciones/excepciones no solo número externo (L34) | Reportes · Work-item boundary → Cierre · Close question | Antes de cerrar work item | Close/build; núcleo enlaza heading por nombre |
| Investigación/follow-up termina por su ruta en lugar del cierre general (L34) | Reportes · Work-item boundary | Antes de cerrar investigación | Flow-research |
| Coordinar proporcional; anunciar split y cambios, no permanencia en principal; pequeño directo (L89) | Delegación · Split and announce work; invariante global | Antes de delegar/cambiar split | Todos los lanzadores |
| Salida larga a hijo resumidor; reassess si lookup se vuelve diagnosis (L89) | Delegación · Split and announce work | Antes de lookup largo | Research/build/plan |
| Independencia acotada, dependencias secuenciadas, sin escritores solapados (L89) | Delegación · Split and announce work; invariante global | Antes de delegar | Todos los lanzadores |
| Trabajo útil mientras procesos externos; cada proceso con señal antes de terminar turno (L89) | Delegación · Launch and wait | Antes de lanzar/esperar proceso | Todos los lanzadores; verification/git-workflow |
| Hijo que excede tramo en background si host señala; foreground solo con aviso qué/cuánto/decide (L89) | Delegación · Launch and wait | Antes de lanzar hijo | Todos los lanzadores |
| Sin trabajo útil esperar señal, no polling; señal turno posterior exige reporte activo; host descarta resultado obliga recopilar antes (L89) | Delegación · Launch and wait → Reportes | Antes de espera/fin turno | Todos los lanzadores |
| Espera bloqueante 5–10min o máximo menor, no slices 30/60; mismo proceso/watch, sin sleep-check (L89) | Delegación · Launch and wait | Antes de espera explícita | Todos los lanzadores |
| Tramo expira: línea estado y qué decide; hijo devuelve final no interim (L89) | Delegación · Launch and wait | Al expirar tramo/retornar hijo | Principal/hijos |
| Rol por entregable/efectos, brief objetivo/root/autoridad/paths/contracts/output/evidence (L90) | Delegación · Select the role y Supply a bounded brief | Antes de dispatch | Principal |
| Brief no grant extra; tree/paths, listar borrados antes y parar fuera propiedad (L90) | Delegación · Supply a bounded brief; autoridad global | Antes de dispatch y edición hija | Principal/hijos |
| Commit hijo solo grant; no push/PR/merge/tracker; padre revisa diff e integra (L90) | Delegación · Supply a bounded brief; autoridad global | Antes de dispatch y efectos hijos | Principal/hijos |
| ID native exacto si listado; título en prompt no selecciona perfil (L91) | Delegación · Select the role | Antes de dispatch | Principal |
| Falta selector/rol listado: leer contrato instalado, generic acotado, disclose; contrato inaccesible reportar no inventar (L91) | Delegación · Select the role | Antes de fallback | Principal |
| Brief paths/decisiones/conventions/recursos; no herencia de conversación/lecturas; mínimo relevante (L92) | Delegación · Deliver and resolve instructions | Antes de dispatch | Principal/hijos |
| Path instalado/root-relative; resolver owner por catálogo/loader/taskpath, relativo a skill no cwd/role (L92) | Delegación · Deliver and resolve instructions; invariante global | Antes de leer recurso requerido | Todos |
| skill locator no comando/auto URI; referencia no invoca workflow entero (L92) | Delegación · Deliver and resolve instructions; invariante global | Antes de leer recurso requerido | Todos |
| Hijo lee antes dependencias y reporta inaccesible; pause solo dependencia, no disponibilidad = lectura (L92; contrato arquitectura) | Delegación · Deliver and resolve instructions; invariante global | Antes de trabajo dependiente | Todos |
| Premisa fact/sourceclaim/assumption con evidencia o check ausente (L93) | Delegación · Supply a bounded brief | Antes de brief | Principal |
| Hijo findings/changes/evidence/limits/accessgaps; padre contrasta brief y fuente/resuelve/integra no assertion (L94) | Delegación · Check returned work | Al retornar hijo | Principal/hijos |
| Relay todos defectos severidad propia o incidental, ninguno omitido (L94) | Delegación · Check returned work | Antes de relay | Principal |
| Resultado credenciales: comprobar lugar durable/shared y comunicar por Secrets (L94) | Delegación · Check returned work; Secrets global intacto | Antes de relay sensible | Principal |
| Registro rol/path/dispatch/model/accesscontrols en record existente; no verificado explícito (L96) | Delegación · Record consequential delegation | Ante delegación consecuente | Principal |
| Frontmatter/config/exclusion/domain no prueba control efectivo/aislamiento/model/lectura; reportar real y override (L96) | Delegación · Record consequential delegation; invariante global | Ante claim de control | Principal |
| Hive por repo con Project/Base branch/Tracker/Specs y URL externa; opcionales definidos (L100) | Settings · Declare settings | Antes de resolver setting requerido | Todos; plan/build/research/close/git/harness/spec-role |
| Workspace crossrepo, no herencia repo; ejemplo completo (L100–110) | Settings · Declare settings | Antes de resolver setting | Todos |
| Gobierna repo del archivo, package igual; workspace crossrepo/sin target; conflicto repo+reporte (L111) | Settings · Resolve the governing section; invariante global | Antes de resolver setting | Todos |
| Faltante cuando necesario no startup/read-only repo; preguntar y registrar allí; otros docs recomendación no setting (L112) | Settings · Resolve a missing value → Preguntas | Antes de usar valor ausente | Todos |
| Specs relativo root declarante; Locate ante checkout ausente/compartido; preguntar clone, no crear repo; sin URL valor ausente (L123) | Registros · Location and lifecycle → Locate the specs | Antes de localizar specs | Plan/build/close/harness/spec-role |
| Ledger/current requirements/product-map/changes/archive taxonomía (L125) | Registros · Location and lifecycle | Antes de crear/continuar/cerrar registro | Plan/build/close/harness/spec-role |
| Plan retenido changefolder, pequeño sin plan; kebab-case/tracker prefix/format (L127) | Registros · Location and lifecycle; invariante global | Antes de crear/continuar plan | Plan/build/harness/spec-role |
| Close archiva integrated/abandoned y aplica deltas primero si integrado (L128) | Registros · Location and lifecycle → flow-close | Antes de cerrar registro | Close |
| Local/sensitive pointers README soporte; ledger/openspec versionado por reglas repo (L129) | Registros · Location and lifecycle | Antes de escribir registro/puntero | Plan/build/close/spec-role |
| Sessions/ledgers históricos permanecen; mismo home al continuar; migración explícita (L130) | Registros · Location and lifecycle; invariante global | Antes de continuar/migrar registro | Todos |
| Solo carpetas necesarias, work compartido change-id acortable descriptivo nunca tracker solo; sin change English topic (L141) | Artefactos · Identify the work | Antes de escribir soporte | Todos los escritores |
| docs living sin fecha actualiza in-place (L143) | Artefactos · Select the destination | Antes de escribir durable knowledge | Workspace-conventions y escritores |
| sessions record sin change; tipo/path completo y mismo home, no root aun pequeño (L144) | Artefactos · Select the destination | Antes de record retenido sin change | Todos los escritores |
| workspace ignored material abrible/versioning undecided o requerido por herramienta/futura sesión (L145) | Artefactos · Select the destination | Antes de working material | Todos los escritores |
| Scratch reproducible en temp host, no root/reusable; única copia no scratch (L146) | Artefactos · Select the destination; invariante global | Antes de scratch/cleanup | Todos |
| evidence subset probado enlazado (L147) | Artefactos · Select the destination | Antes de guardar evidencia | Todos los escritores |
| UI images workspace/images y handoff; hasta fin sesión; excepción pausa, selección usuario; no commit ni publicación implícita (L148) | Artefactos · Select the destination | Antes de UI capture/cleanup | Build/verification/close |
| Naming topic.type.ext English; research/report propósito; Markdown salvo formato pedido; tool names intactos (L152) | Artefactos · Name and date records | Antes de record nuevo | Todos los escritores |
| Fecha inicial trabajo salvo archive fecha cierre, no duplicar en filename; preservar home/date/names no migrate automático (L153) | Artefactos · Name and date records | Antes de record/continuación/archive | Todos los escritores |
| Calendario usuario, clock/timezone si incierto; timestamp evidence intacto+offset relevante (L154) | Artefactos · Name and date records | Antes de fechar record | Todos los escritores |
| Evidencia necesaria/binaries justificados, Git solo subset curated/shareable; no secrets/raw/unneeded reproducible; privado límites (L159) | Artefactos · Retain evidence | Antes de retener/versionar evidencia | Todos |
| Retener no versionar, elegible no commit/push grant (L159) | Artefactos · Retain evidence; invariante global | Antes de retener/versionar | Todos |

Comprobación del autor: contraste semántico de las 101 filas contra la referencia editorial, más búsqueda de consumidores en `content/guidance/`, `content/skills/`, `content/agents/`, `site/src/content/docs/`, `_support/docs/architecture/` y detectores existentes de `tests/pilot/`. Las frases que cambiaron separan condición/acción, resuelven paths relativos o sustituyen localizador numérico por heading; no pretenden retirar obligaciones. Un contraste textual de secciones confirmó intactos `Language`, `Scope and authorization`, `Preservation`, `Secrets`, `Evidence`, `Verification gates` y `Continuity` respecto a `a4d6e254`; los hashes del núcleo en ambas bases de comparación coinciden. Los ejemplos nuevos cubren tarjeta que espera, texto sin tarjeta y reporte con publicación pendiente. No se añadió estado formal, runtime, hook, límite de tokens ni política nueva.

El núcleo conserva idioma, autorización, preservación, Secrets, Evidence, proporcionalidad, ubicación básica, invariantes de citas y disparadores. Se preservan los contratos de reutilizar la revisión ya elegida en la base integrada y selección provisional de delivery antes de opciones completas. Los enlaces globales y de roles usan `skill:owner/path`; los de skills son relativos. `go test ./tests/content -run 'TestGlobalGuidance(SkillLinksResolve|Size)$' -v` pasó con 25.363 bytes/caracteres del núcleo; reducción 21.549 bytes (45,93 %). `git diff --check` pasó. También pasó `go test ./tooling/management -run '^TestRepositoryCatalogueInstructionReferences$' -v`, que valida los nuevos enlaces relativos y de roles del catálogo. Al ejecutarlo ya estaban presentes cambios de otro escritor en `tooling/management/references.go` y `references_test.go`; se preservaron sin editar ni atribuirlos a T1. Son checks estructurales y una inspección del autor, no verificación independiente de AC1/AC2 ni eficacia del modelo.

Lecturas fuera de propiedad que requieren alineación documental en T5: `_support/docs/architecture/instruction-resources.md` (home runtime de delegación), `agent-delivery.md` (contrato global, esperas y preguntas), `repository-and-distribution.md` (núcleo frente a referencias) y navegación del sitio. La semántica pública actual de tarjetas/rutas continúa vigente; no se editaron esas surfaces. `tests/content/skill_links_test.go` no separa anchors al hacer `os.Stat`, por lo que el puntero global a cierre usa el entrypoint y nombre de heading, no un destino con fragmento; los enlaces relativos sí apuntan al heading. No se editaron pruebas ni detectores. El verificador independiente y T5 revisarán cobertura, semántica y coste de la unión de lecturas; menos global no acredita menos contexto total.

### Plugin OpenCode

Fuente editorial propuesta: `integrations/opencode/context-loader.js`. Destino inactivo: `<OpenCodeHome>/hive/context-loader.js`, fuera de `plugins/`, que OpenCode autodescubre. Identidad native: `hive.context-loader`. No paquete publicado ni dependencia nueva del manager, ni clasificador por modelo.

La entrega produce también `<OpenCodeHome>/hive/context-loader.json`, con schema `1` y la raíz absoluta `skillsDirectory` del mismo `target.Config` que resuelve los skills. Es metadata derivada en la planificación de entrega, no otra fuente editorial ni un permiso para leer cualquier directorio. Validar que los bytes de esa configuración corresponden al plan y su scope antes de aplicar; no aceptar una raíz de los argumentos del modelo, del prompt o de una config de proyecto. El JSON no contiene secretos y comparte la propiedad de recurso OpenCode con el script.

La fuente incorpora una lista cerrada:

- `hive-review-plan` → `flow-plan/references/plan-review.md`.
- `hive-review-harness` → `harness-audit/SKILL.md`.

Ambos roles ya obligan a leer esos recursos incondicionalmente. No se precargan sus referencias condicionales ni todo el catálogo. No se cubren preguntas del hilo principal ni otros roles en este primer complemento; la reorganización común sí les aplica.

Contrato native consultado el 2026-10-11: `Plugin.define({id, setup})`; registrar `ctx.session.hook("context", callback)`. `SessionContextHook.agent` es un identificador, no necesariamente el nombre del archivo. Resolverlo mediante `ctx.agent.get({agentID: event.agent})` y comparar `Agent.Info.name` contra la lista cerrada. El OpenAPI V2 público confirma que `Agent.Info` exige `id` y `name`. No depender de suscripciones ni nombres de modelos.

Para un rol cubierto, leer los recursos permitidos, resolver ruta real y comprobar confinamiento dentro de `skillsDirectory`, archivo regular y UTF-8. Insertar un `SystemPart` `{type: "text", text: ...}` con referencia y contenido canónico, sin reinterpretar ni ejecutar Markdown. No solicitar un archivo elegido por el usuario ni un enlace extraído del mensaje. Una lectura inválida de un recurso requerido impide leer/suministrar ese recurso y añade un diagnóstico acotado de referencia no suministrada a la petición; no promete abortar el despacho (D9-A). Un rol fuera de la lista no se modifica.

El diagnóstico lleva únicamente identidad lógica del recurso y código de causa, como `missing_reference`, `invalid_metadata` o `outside_skills_root`, más la indicación de informar la dependencia y no actuar sobre su procedimiento como si estuviera cargado. No incluye errores crudos, metadata recibida, rutas externas, contenidos ni credenciales. La ausencia/corrupción de la metadata derivada sigue esta misma vía para roles cubiertos. La conducta posterior del modelo es una obligación del núcleo, no un control de permisos del plugin.

El hook trabaja sobre la petición ensamblada, no sobre el historial. No registra «ya suministrado» a nivel sesión: cada petición, incluidas continuaciones y posteriores a compactación, recibe la referencia de nuevo si corresponde al rol. Deduplicar únicamente las partes añadidas por este plugin en la petición actual. No tocar el contenido ajeno, herramientas, permisos ni opciones de generación. Desregistrar el hook al descargar el plugin. Otros plugins pueden desactivarlo o modificar la petición después: el núcleo sigue exigiendo lectura, y la evidencia nativa debe distinguir suministro de cumplimiento.

### Distribución y activación

Ampliar el catálogo con esta única fuente JS y los destinos OpenCode `hive/` mediante las operaciones actuales de archivos gestionados. Añadir un kind explícito para el recurso si la discriminación actual lo exige; no llamarlo skill ni introducir un instalador paralelo. El manager sigue sin lanzar modelos.

El ensamblador offline `tooling/package/main.go:112–116` solo copia hoy `content` y el JSON de perfiles. Agregar la fuente exacta `integrations/opencode/context-loader.js` al conjunto copiado y validar su presencia en el paquete construido/extraído. Comprobar que ese paquete prepara una instalación OpenCode con script y metadata, sin activar nada ni publicar una release; probar solo el checkout no cierra esta superficie.

El esquema de los snapshots nuevos incorpora la fuente JS; conservar lectura de releases antiguas que no la contienen. Una release antigua retira solo los nuevos archivos gestionados cuando ya no los entrega. Hashes y validación de planes deben cubrir la metadata derivada y los consumidores. El código actual que vuelve a validar los destinos al aplicar no debe confiar en un plan manipulado ni en fingerprints antiguos.

Activación posterior y aparte: el usuario registra la ruta inactiva en el campo native `plugins` de `opencode.json(c)`; los docs aceptan rutas de archivos explícitas. La entrada y su elección son del usuario, no propiedad del manager. Desactivar por la directiva native `-hive.context-loader` o retirando esa entrada. No crear ni sustituir el documento de config global desde este trabajo. Antes de desinstalar Hive o volver a una release sin plugin, retirar la entrada explícita; de lo contrario OpenCode puede informar un plugin ausente. El manager preserva esa configuración ajena, y la documentación no debe prometer limpieza automática de la entrada.

En los pilotos, registrar la ruta solo en el documento nuevo del hogar ficticio. Observar que el plugin fue descubierto, registró el hook y el rol recibió el texto antes de despacho. Unit tests con callbacks falsos no sustituyen esa observación.

### Seguridad y recuperación

Leer el contrato de `flow-build/references/security-boundaries.md`: el plugin corre con el acceso del proceso native, no crea sandbox. Su entrada pública es un identificador de agente; el catálogo de recursos es fijo y la raíz procede de entrega registrada. Validar JSON y rutas antes de lectura, incluyendo symlink fuera de raíz, raíz inexistente, archivo ausente y manipulación de metadata. No procesar shell, enlaces o texto del prompt como código.

Las pruebas usan hogares ficticios y revisan inexistencia de escrituras fuera de ellos. Logs de diagnóstico: código de fallo, nombre del recurso permitido y IDs/hashes cuando proceda, nunca el texto de mensajes, instrucciones privadas o credenciales. Recuperación usa backup/journal actuales para script y metadata; un fichero ajeno o editado es conflicto. Volver a la referencia anterior no activa ni edita una config native del usuario.

Ejercitar interrupción entre las escrituras del par script/metadata y durante downgrade. La recuperación debe devolver archivos y propiedad a un estado coherente, preservando config ajena y ficheros cambiados después del inicio de la operación. Los tests existentes de recuperación de skills no acreditan este nuevo kind/metadata.

## Piloto autorizado (D7-B)

48 ejecuciones raíz = 4 modelos × 3 variantes × 2 casos × 2 repeticiones. Cada caso inicia una sesión nueva y despacha un único hijo del rol objetivo; contabilizar también las peticiones y sesiones hijas efectivas, no llamar 48 al total de llamadas al proveedor.

Modelos solicitados:

- `opencode/claude-haiku-5-5#high`.
- `opencode/deepseek-v4.1-flash#high`.
- `opencode/qwen3.8-flash#medium`.
- `opencode/glm-5.3-flash#high`.

Catálogo consultado el 2026-10-11; no es prueba de auth o cuota. Variantes distintas no significan igual esfuerzo entre familias. Preflight con la primera ejecución prevista de cada modelo, sin llamadas de prueba adicionales ni sustituir el modelo automáticamente. Los roles de las fixtures fijan el mismo modelo solicitado que su padre, sin editar perfiles del usuario; verificar modelos observados para no comparar ejecuciones que acabaron en otro.

Variantes:

- **A:** contenido actual de referencia `d5164f71`, sin plugin.
- **B:** candidato común reorganizado, sin plugin activo.
- **C:** exactamente el mismo candidato B, con plugin activo en el hogar ficticio.

Comparar A/B para reorganización y B/C para complemento. Igual petición, archivos, permisos y criterio por caso; rotar el orden de variantes de forma predefinida y guardar el orden. No introducir un check adicional solo en C ni copiar las instrucciones esperadas a la petición.

Casos en `tests/fixtures/flash-class-guidance/`, no en `tests/fixtures/regression/`:

1. `plan-reference`: fixture de plan con brechas verificables de criterio/tarea y autoridad; el padre pide un único `hive-review-plan`. Comprobar descubrimiento o suministro de la referencia de plan-review, hallazgos esperados y ausencia de ejecución del plan.
2. `harness-reference`: fixture de instrucciones con un enlace inválido y una frontera de autorización contradictoria; el padre pide un único `hive-review-harness`. Comprobar lectura/suministro de `harness-audit/SKILL.md`, hallazgos esperados y ausencia de fixes no pedidos.

La inclusión de estos defectos está limitada a datos de prueba y no crea reglas distribuidas ni casos nuevos de detector sintético. Estos dos casos no representan todas las tareas de desarrollo ni comportamiento en otros hosts.

Runner existente: `tests/pilot`, usando `run --standalone` de OpenCode. Adaptar solo lo necesario para C, selección/observación del hijo, estado del complemento y trazabilidad. Sus flags actuales admiten A/B y fijan `--agent build`; no afirmar que soporta ya el protocolo completo. Reconciliar la config native actual y la importación de credenciales en hogares nuevos: no escribir config personal ni mostrar secretos, ni reutilizar a ciegas formato V1.

Memoria **aislada**: directorio Engram fresco y verificado vacío por ejecución, cloud sync desactivado, evidencia de routing efectivo y detección de registros ficticios que pudieran alcanzar el store cotidiano. Usar el mecanismo soportado del runner y comprobarlo; el entorno pedido no prueba aislamiento. Credencial local del proveedor por el acceso native existente, una copia privada descartable dentro del entorno aislado y cleanup comprobado, sin nuevo login ni exportar claves privadas. No copiar auth a evidencia.

Límites: hasta dos ejecuciones en paralelo, 3 minutos por ejecución raíz y sin reintentos ni ampliación automáticos. Seguir el tiempo de vida del hijo para que el fin del padre no deje trabajo ejecutándose ni permita declarar concluida una corrida sin resultado del hijo. Un fallo de acceso deja bloqueada la cobertura de ese modelo; no selecciona otro proveedor. Parar el experimento ante una escritura fuera de fixture, contaminación de memoria, exposición durable de secretos o falta de control de procesos. El usuario aceptó que no hay un tope de facturación garantizado; reportar uso y coste solo cuando sean observables, sin deducir facturación de un precio de catálogo.

Registrar por ejecución fuente/hashes, versión CLI, modelo solicitado/observado, configuración relevante, lectura o suministro previo al despacho, criterios y pausas/intervenciones. Para C, capturar solo referencia/hash/rol/fase, sin prompts ni credenciales, y comprobar que el contenido suministrado coincide con el recurso. B y C usan idéntica instrumentación de observación; el único cambio es activar el cargador.

### Observador independiente del cargador

T4 crea un observador no mutante separado del módulo de producto, idéntico en A/B/C y cargado después de él en las fixtures controladas. En `context`, observa las partes ya ensambladas, resuelve el rol y registra sesión, agente, secuencia de petición dentro de la sesión, recurso permitido, hash de los bytes extraídos y número de copias. Compara contra el archivo canónico de ese brazo, no contra un hash que solo haya declarado el loader. No escribe esa prueba en contexto ni da feedback al modelo.

Correlacionar el snapshot con el despacho mediante los eventos native `model.request`/`http.request` correspondientes. La documentación actual da a los eventos HTTP `sessionID`, `agent`, `model` y `kind`; distinguir `primary` de título/compaction/generate. Capturar solo los campos permitidos de texto de la petición real necesarios para la comparación, nunca headers, URL con credenciales ni un dump de la petición. Un protocolo no reconocido se declara no verificable. Orden de plugins y cambios posteriores deben ser observables en el hogar aislado; si no se puede acreditar el snapshot final o su correspondencia con la petición del hijo, AC4 queda sin verificar.

OpenCode documenta transporte WebSocket experimental que evita los hooks HTTP. El protocolo de piloto requiere un transporte HTTP observable para estas rutas, o declarar esa cobertura bloqueada; no añadir parsers de frames/transportes ni conexiones nuevas silenciosamente. Exportar historial tras el run solo prueba historial, no las modificaciones de contexto que el host no persiste.

Antes del piloto, fijar la rúbrica de hallazgos esperados y la codificación de pausas. Cuando una puntuación requiera juicio humano, evaluar con la etiqueta de variante oculta y revelar después; la evaluación por otro modelo, si se usa alguna vez, no es juicio humano ni demostración de calidad.

## Reconciliación de revisión final

Candidato final revisado por los mismos dos roles: SHA256 `d73df7a9bb57c9fd588cc941d51a72071da7f44b8cc164add83cad059135b6a1`.

- **Entrega H1, alta:** confirmada omisión de la fuente JS en el ensamblador offline. Se adopta la corrección propuesta: incluir fuente exacta y prueba de paquete extraído, en T3/AC6.
- **Entrega H2, alta:** confirmado disparador insuficiente para glosas/enlaces. Se conservan los invariantes globales y se extiende la lectura del detalle a todo mensaje que los use, incluidas preguntas y retransmisiones, en T1/AC1–AC2.
- **Entrega H3, media:** se adopta la prueba explícita de interrupción y recovery del par y de downgrade, en T3/AC6.
- **Entrega, sugerencia Cursor:** se corrige la especificación para describir entrega por los canales existentes y sus requisitos de descubrimiento, sin prometer carga universal automática.
- **Plugin/piloto H1, alta:** se adopta el observador independiente propuesto y su correlación previa al despacho, en T4/T6/AC4–AC8. HTTP/WebSocket quedan delimitados; ausencia de esa observación es bloqueo, no éxito inferido.
- **Plugin/piloto H2, media:** garantía de abortar petición por excepción del hook no acreditada. Resuelta por decisión explícita D9-A posterior a la revisión: el complemento no ofrece esa garantía, rechaza la lectura y da diagnóstico sin cambiar permisos. AC5 y delta ajustados; se verifican lectura confinada, ausencia de contenido inválido y diagnóstico, no aborto ni obediencia garantizada del modelo.

Las correcciones anteriores aplican las propuestas acotadas de los reviewers y el padre las reconcilió sin otra revisión completa. D8 y D9 son decisiones explícitas posteriores al candidato revisado: fijan actores y retiran la garantía de aborto no documentada. Se registra esa diferencia, no se atribuye a un veredicto anterior. No quedan brechas de diseño abiertas; acceso/cuota, suministro native y comportamiento se comprobarán en la construcción con sus gates. Las pruebas y el piloto siguen sin ejecutarse.

El reporte separa errores del modelo, de guía, de cargador y de infraestructura. No claim estadístico ni mejora universal con dos repeticiones. Un timeout, rol no despachado, diagnóstico de referencia no suministrada o modelo no observado no cuenta como cumplimiento de carga. Conservar fallos antes de cualquier repetición futura. La ampliación o nueva verificación de modelos tras fixes necesita una decisión, pues las 48 ejecuciones no son una autorización ilimitada.

Scratch reproducible en la carpeta temporal del host. Evidencia seleccionada en `_support/evidence/2026-10-11-flash-class-guidance/`; datos privados descartables en scratch con permisos restrictivos. El reporte y enlaces de resultados viven en `tasks.md`, sin otro research. Al acabar todos los writers, limpiar solo registros ficticios con IDs y proyectos inequívocos por mecanismo soportado, parar procesos propios y comprobar que no quedan copias de credenciales de pruebas. Un cleanup no verificable sigue pendiente, no aprobado.
