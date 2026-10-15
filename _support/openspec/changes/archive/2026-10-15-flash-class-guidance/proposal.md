# Guía de Hive accesible a Haiku y modelos Flash

| Campo | Valor actual |
| --- | --- |
| Estado | Integrado en `development` el 2026-10-15 por el [PR #52](https://github.com/JhonHawk/tricell-hive/pull/52), merge `1307b1fb`; desplegado en local desde `development` y archivado el 2026-10-15. Estado previo (historia): Entregado en local: 24 commits en `feat/flash-class-guidance` (4 de `research-child-contract`); 27 commits, desplegados en los seis hosts hasta `982f9828`; sin push, PR ni merge por decisión del usuario · AC4 nativo observado · AC8 registrado |
| Tracker · GitHub Issues | Sin issue asociado; no se ha autorizado abrir uno |
| Git | Integrado por merge del PR #52; la rama `feat/flash-class-guidance` y su worktree se eliminaron tras el merge |
| Verificación | Revisión del plan y de implementación con OpenAI Sol · pruebas ordinarias · piloto acotado autorizado para este cambio |
| Siguiente paso | Ninguno; registro cerrado. T6 (entorno de pruebas en bb1) queda diferida: bb1 solo si una prueba necesita Linux o cómputo (decisión del 2026-09-30) |

Fecha inicial: 2026-10-11, hora local CST.

## Pedido

> Exacto, ese es el nivel de pensamiento que tengo, heiku, y los modelos flash como qwen, glm y deepseek los estoy considerando, entonces me es relevante que hive sea útil tambien a ellos

> vamos con el plan

> si puedes cambiar el modelo de los revisores del plan que se con chatgpt sol 6.1 porfa (no con copilot)

> Contemplemos liberar esta rama /versión a development y desde ahi crear una rama nueva

> asi un hook o un cargador mejora el descubrimiento, analicemoslo

> apruebo los pilotos

## Objetivo

Reorganizar las instrucciones para que Hive resulte más accesible también a Haiku y modelos Flash de Qwen, GLM y DeepSeek, conservando las obligaciones actuales. Reducir el esfuerzo de interpretar reglas compuestas y localizar procedimientos, sin confundir menos texto global con menos contexto total ni instalación con cumplimiento.

No se ha demostrado una incapacidad de esas familias ni un umbral óptimo de tokens. El precio, la denominación Flash y los resultados de otros modelos no establecen por sí solos el comportamiento de una versión concreta.

## Decisiones vigentes

- **D1-A — Conservar:** mantener las reglas actuales y simplificar su organización y ejecución. No se autorizó eliminar glosas, preguntas, pausas, categorías de cierre ni otras obligaciones de comunicación.
- **D2 — Secuencia:** contemplar la integración de `feat/guidance-rework` en `development`; crear después una rama nueva desde la base integrada. No construir sobre la rama Pi actual ni heredar sus commits particulares.
- **D3-A — Entrega previa:** cuando se autorice ejecutar, verificar y publicar `feat/guidance-rework`, abrir PR a `development` y fusionarlo tras los controles y la revisión elegida. Cerrar y entregar una vez los registros de sus dos cambios mediante PR de cierre sin revisión dedicada. No incluye `master`, release ni instalación global.
- **D4-A — Nuevo trabajo en hold:** cuando se autorice construir, crear una rama propuesta `feat/flash-class-guidance`, implementar y verificar localmente; detenerse con diff y registro en disco, sin commits, push, PR ni merge de este nuevo trabajo.
- **D5-A — Revisión de implementación:** `hive-review-code` en `openai/gpt-6.1-sol`, con `hive-refute-claim` en ese proveedor para hallazgos bloqueantes, sobre los candidatos previo y nuevo, antes de entregarlos. No es la revisión del plan ni demuestra eficacia de los modelos objetivo.
- **Modelo de revisión del plan:** el usuario autorizó `openai/gpt-6.1-sol`, sin Copilot, como opción por lanzamiento. No modificar perfiles distribuidos ni configuración global.
- **D6-B — Complemento OpenCode:** además de la reorganización común, incluir un complemento opcional de OpenCode V2 que suministre referencias canónicas según roles explícitos, sin clasificador LLM ni enforcement. No introduce una integración obligatoria para los otros hosts.
- **D7-B — Piloto de cuatro familias:** el usuario levantó la pausa para este cambio y eligió 48 ejecuciones raíz, tres variantes, dos casos y dos repeticiones. Modelos y límites en `design.md`; acceso y cuota pendientes de preflight. La pausa sigue vigente fuera de este experimento. Ejecutar después de obtener el candidato bajo una construcción autorizada, no durante este turno de planificación.
- **D8-A — Construcción y verificación por tarea:** implementadores con `openai/gpt-6.1-sol`; `hive-verify-task` con `openai/gpt-6-astra`, modelos distintos, mediante overrides por lanzamiento, sin modificar perfiles globales. La disponibilidad en catálogo no prueba acceso; ante un fallo del verificador se detiene ese gate, sin sustitución automática.
- **D9-A — Fallo de carga:** rechazar lectura y suministro del recurso inválido y comunicar un diagnóstico acotado al modelo, sin afirmar que OpenCode cancela la petición. El núcleo sigue exigiendo informar y pausar el trabajo dependiente; ausencia de carga no cuenta como éxito. Se descarta el requisito no acreditado de aborto native por excepción.
- **D10-A — Construcción autorizada:** iniciar `flow-build` con este plan, incluida la entrega previa D3-A y las tareas de la rama nueva en `hold`, con gates y piloto ya acordados. No incluye despliegue global, `master` ni release.
- **D11-A — Snapshot generado:** volver a planificación para sustituir lectura viva de referencias por datos generados desde el payload validado e incorporados al recurso instalado. Una fuente editorial por referencia, copia solo derivada; cambia al regenerar y recargar el complemento. Se descarta la lectura viva insegura y no se añade helper ejecutable por petición.
- **D12-B — Mantener Engram:** conservar memoria aislada del experimento y continuar investigación de una limpieza soportada vinculada al mismo store, sin editar Engram, usar SQL ni migrar configuración global. No se desactiva memoria por inferencia y el piloto queda bloqueado si esa garantía no existe.
- **D13-A — Investigar aislamiento:** investigar de forma acotada y solo lectura si capacidades existentes pueden impedir la sustitución del store y conservar el contrato de limpieza. No autoriza instalar herramientas, editar Engram/configuración global, borrar datos ni ejecutar modelos; presentar el resultado antes de diseñar o implementar otra protección.

## Alcance y restricciones

Incluye el núcleo global, los procedimientos y referencias que absorban su detalle, coherencia con skills y roles consumidores, validación de punteros del núcleo, un plugin opcional OpenCode suministrado inactivo, su entrega/recuperación como recurso gestionado, documentación y piloto de cuatro familias.

Se conservarán en el núcleo las fronteras esenciales de autorización, preservación y secretos, las reglas básicas de ubicación y los disparadores explícitos de lectura. Las referencias no son permisos, y un hook no sustituye las restricciones del host.

No incluye cambiar los modelos predeterminados de Hive, un preflight de modelos como función nueva del manager, enforcement de mensajes, editar `tests/fixtures/regression/`, reanudar otros pilotos, crear una release, promover a `master` ni instalar en los directorios globales reales. El cambio Pi pertenece a otro registro y no se absorbe. El plugin no clasifica prompts ni gobierna permisos; su primera cobertura son los dos roles de lectura obligatoria definidos en el diseño, no todos los roles ni los hilos principales.

Una obligación solo se mueve con un destino canónico y una condición de lectura que cubra todos sus consumidores. No se añadirá una prohibición nueva sin la evidencia o decisión que exige `AGENTS.md`. No se impondrá un techo arbitrario de tokens como prueba de utilidad.

## Criterios de aceptación

La base de comparación de contenido es `d5164f71`; registrar la base integrada real al construir y confirmar que estos criterios siguen siendo falsos allí. Las restricciones ya ciertas, como preservar datos ajenos y mantener la suite, no se presentan como criterios nuevos.

- **AC1.** El núcleo es menor en bytes que `global.md` en `d5164f71` y reemplaza las taxonomías completas de artefactos y registros por ubicación básica y disparadores enlazados a recursos canónicos. Cada obligación trasladada tiene destino, condición de lectura y consumidores en la matriz de `design.md`. *Falso en la base:* el detalle sigue en el bloque global y no existe ese reparto implementado. La disminución no demuestra por sí sola menor contexto activo.
- **AC2.** Los procedimientos trasladados de preguntas, reportes y delegación/espera presentan condiciones y acciones separadas, conservan las excepciones actuales e incluyen ejemplos resueltos de pregunta con/sin tarjeta y reporte con trabajo pendiente. El cierre deja de localizar su regla por un número de paso externo como único acceso. *Falso en la base:* las viñetas compuestas y los localizadores por paso permanecen, y faltan esos ejemplos.
- **AC3.** El manager rechaza un enlace empaquetado roto del núcleo al preparar una instalación y al aplicar un plan guardado con hashes recalculados, antes de escribir destinos. Enlaces externos y ejemplos cercados mantienen su tratamiento actual. *Falso en la base:* `instructionMarkdown` excluye el núcleo de la validación.
- **AC4.** El plugin OpenCode resuelve el nombre del agente activo e introduce el snapshot completo de la referencia fija de cada rol cubierto en `context` antes del despacho; un rol no cubierto queda intacto. El snapshot se genera desde el mismo payload validado que entrega sus fuentes canónicas; una petición nueva lo suministra de nuevo sin marcarlo para siempre como ya cargado ni reabrir referencias por ruta. *Falso en la base:* no existe el plugin. No afirma que otros plugins posteriores no modifiquen el contexto ni que una edición local posterior regenere automáticamente el snapshot.
- **AC5.** El complemento no realiza lecturas de archivos de referencias o metadata en su loop de petición; solo usa los dos snapshots generados y su procedencia validada por el manager antes de escribir. Datos ausentes o inválidos impiden aceptar la entrega generada; diagnóstico acotado si el snapshot no está disponible, sin contenidos ajenos ni credenciales. No garantiza aborto del despacho ni cambia herramientas, permisos, modelo, sampling, mensajes persistidos o configuración del usuario. *Falso en la base:* no existe esta frontera de suministro por snapshot.
- **AC6.** El manager entrega el plugin como directorio inactivo (`hive/context-loader/` con `index.js` y la metadata, fuera de `plugins/`; D21-A) únicamente a OpenCode usuario, con su propiedad, drift, reinstalación de ausentes, remoción y retorno a una release anterior definidos. La instalación normal no registra ni autoactiva el plugin, ni escribe `opencode.json(c)`. Los otros cinco hosts y los dos scopes de proyecto existentes no reciben el destino del plugin. *Falso en la base:* ese recurso y destino no existen.
- **AC7.** README, `llms.txt`, sitio y arquitectura describen núcleo/referencias, cobertura opcional y activación/desactivación manual, y distinguen disponibilidad, carga y conducta. El informe de tamaño cubre global y la unión de lecturas declaradas para consulta simple, pregunta y ciclo de plan/construcción; especifica costes de metadata no medidos y usa bytes/caracteres sin llamarlos tokens reales. *Falso en la base:* no se describe el nuevo reparto ni existe esa comparación de recorridos.
- **AC8.** El protocolo y el reporte del piloto registran las tres variantes, modelos solicitados/observados, hashes del contenido, lecturas o inyección verificable, resultado por criterio, pausas, uso/coste disponible, fallos de acceso y limpieza de memoria/procesos/credenciales temporales. La ejecución se ajusta a D7-B y no convierte falta de acceso ni ausencia de lectura observada en éxito. *Falso en la base:* no existe este experimento ni sus resultados.
- **AC9.** En el modo explícito de store desechable en contenedor, el runner termina una corrida con exportación de evidencia y estado `deferred_to_volume_destruction`, sin el error fijo de binding, sin llamar a la limpieza inmediata del store y sin reportar `verified_clean` ni el store cotidiano como verificado. Rechaza el modo fuera de un contenedor, con `ENGRAM_CLOUD_*`, `ENGRAM_URL`, `ENGRAM_PORT` o `ENGRAM_DATA_DIR` heredados, o con un data dir que escape de la raíz de la corrida, también por enlace simbólico; las variables que el runner genera para su store propio siguen permitidas. Sin el modo, el comportamiento actual se conserva. *Falso en la base:* `flash_memory.go:200` devuelve siempre "atomic memory store binding unavailable" y `main.go:614-626` solo conoce `verified_clean` o `cleanup_withheld_unverified`.
- **AC10.** El lanzador de bb1 construye una imagen `linux/amd64` desde fuentes fijadas (base Debian por digest, Engram 3.3.2 comprobado con el checksum de la release, OpenCode 2.0.27 con integridad npm fijada) cuyas versiones observadas coinciden; crea el contenedor sin puertos publicados, sin bind mounts, sin `docker.sock`, como usuario no root, con tmpfs `/run/hive-pilot/secrets` y etiquetas de piloto y lote; escribe el manifiesto del lote con un token de propiedad antes de crear y las identidades inspeccionadas después, aborta si un nombre ya existe, destruye solo recursos cuyo nombre, etiquetas y token coinciden con el manifiesto (y su ID cuando está registrado) y recupera un lote interrumpido, incluso con manifiesto parcial, copiando la evidencia con sumas antes de destruir, conservando los recursos si la copia falla y sin reintentar modelos. *Falso en la base:* no existe definición de imagen ni lanzador.
- **AC11.** En el modo contenedor, el directorio de datos nativos de cada hogar ficticio (donde se importa la credencial) queda bajo la raíz tmpfs indicada y nunca bajo la salida de la corrida; el catálogo de modelos se toma de la caché del hogar del contenedor, no de la caché del Mac; y el calendario usa los cuatro modelos `opencode-go/` de D15 con las mismas variantes y 48 ejecuciones raíz. *Falso en la base:* `guidance_variant.go:745` coloca los datos bajo `<out>/shadow-home`, `guidance_variant.go:454-464` copia la caché real y `flash_class.go:20` fija IDs `opencode/`.

Las comprobaciones estructurales no acreditan utilidad general para Haiku/Flash. Los pilotos exploratorios pueden mostrar fallos o diferencias en los casos elegidos, no fiabilidad estadística ni mejora universal. Una escritura fuera de la fixture o una violación de autorización introducida por el candidato bloquea su aceptación; no se oculta con un promedio de rendimiento. Si faltan acceso, aislamiento o cobertura nativa, AC8 queda sin verificar y la construcción debe detenerse para resolverlo.

## Entrega

La construcción está autorizada por D10-A. Entrega previa completada: [PR #46](https://github.com/JhonHawk/tricell-hive/pull/46) integrado en `16028d3575e7286cc028985cd2138c3268257add`; registros archivados por [PR #47](https://github.com/JhonHawk/tricell-hive/pull/47), integrado en `a4d6e254bae95dab3764d251d9cb2240cab60a10`. La rama nueva `feat/flash-class-guidance` y su worktree ya existen desde esa base, y el registro canónico se trasladó allí sin duplicarlo. El nuevo trabajo continúa en `hold`.

Referencia de contenido para planear: `feat/guidance-rework` en `d5164f718bf8e492d52a6427d8e6d7723cf3dc3f`. Base local observada: `development` en `bba2fd4826528ef46b11bff68c158880c7e93d95`. Tras la integración previa habrá que registrar el commit integrado real y reconciliar el plan antes de crear la rama nueva.

La instalación global queda excluida. No existe `_support/workspace/maintainer.local` en este checkout (verificado 2026-10-11); la regla de refresco automático de mantenedor no aplica aquí. No se encontraron workflows registrados en GitHub ni `.github/workflows/` localmente; se revalidarán controles y efectos remotos antes de publicar, sin inferir que no hay ninguna automatización externa.

## Fuentes y estado

Contexto y alternativas: [design.md](design.md). Tareas y revisión: [tasks.md](tasks.md). Los registros anteriores permanecen intactos: `question-flow-checkpoints` y `reduce-confirmation-pauses`; sus estados escritos no reflejan todos los commits actuales y se reconciliarán durante su entrega, no dentro de la investigación nueva.
