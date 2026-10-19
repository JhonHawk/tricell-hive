# Procedencia, archivos y capturas del visor

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado en development por PR86 (e9f6cc8b) |
| Tracker · GitHub Issues | • [#79 — procedencia de recorridos versionados](https://github.com/JhonHawk/tricell-hive/issues/79)<br>• [#80 — enlaces duros y dueño de archivos](https://github.com/JhonHawk/tricell-hive/issues/80)<br>• [#81 — tamaño de capturas de escritorio](https://github.com/JhonHawk/tricell-hive/issues/81) |
| Git | Interactiva a `development` al autorizar implementar · pausa antes de push para validación humana |
| Verificación | Pruebas de ingesta/API y reconstrucción · revisión código/seguridad · UX y recorrido funcional locales |
| Siguiente paso | Registro de cierre por PR pequeño; sin promoción a master ni release |

## Solicitud

> research sobre los tres del visor de recorridos

> D1-A

La opción D1-A (plan conjunto) elegida el 2026-10-19 autoriza preparar este plan, no implementar. Issues consultados el mismo día; no contienen una sección formal de criterios de aceptación. Sus propuestas dicen:

- #79 (procedencia): «Ignorar los archivos versionados (`git ls-files`) o marcar su origen en la vista.»
- #80 (archivos): «Rechazar archivos con `nlink > 1` o de otro dueño.»
- #81 (capturas): «Una miniatura por fila en escritorio (o ensanchar la columna de lectura), manteniendo la tira horizontal y el asomo en móvil. Mantener las copias de `tooling/viewer/static/` idénticas.»

## Objetivo y alcance

Separar del recorrido normal las copias cuyo manifiesto está actualmente versionado en Git; impedir nuevas ingestas de esos manifiestos y de archivos con enlaces duros; ampliar la captura de escritorio visible sin rediseñar la página completa. La clasificación informa del estado actual de la fuente: no acredita quién produjo el documento.

Incluye scanner, clasificación Git inyectada desde el CLI, índice/API derivados, interfaz del visor, CSS compartido, pruebas, documentación y deltas. Excluye firmas de agentes, historial de procedencia, nuevos almacenes, cambios de autenticación, decodificación de imágenes, exigencia de dueño único, migración de `meta.json`, promoción a `master`, release y pilotos de modelos.

## Decisiones en vigor

- D2-A (procedencia): omitir nuevas fuentes versionadas y apartar las copias ya importadas de la vista normal, con explicación y sin borrar archivos ni pins. Origen sin determinar se advierte. El apartamiento refleja el índice actual: al dejar de estar versionada, una fuente presente puede volver a la lista como archivo local, sin autoría verificada.
- D3-A (archivos): rechazar `nlink > 1` sobre el descriptor abierto del manifiesto y recursos. No exigir dueño igual al usuario efectivo del servidor.
- Diseño propuesto para capturas: una captura de escritorio por ancho visible en pantalla; conservar tira horizontal, regla estrecha de asomo, miniaturas móviles, carrusel e impresión. No prometer lectura universal de texto en capturas de 2880 píxeles.

## Criterios de aceptación

- AC1. Cuando una consulta Git completa confirma que el manifiesto de una fuente aún no importada está en su índice, la fuente no genera una nueva copia, incluso con `git add -f` bajo una carpeta ignorada. Un fallo de clasificación sigue el estado advertido de AC4 (origen sin determinar), sin garantía de ausencia de tracking. Falso en la base: una fuente confirmada como versionada se acepta como archivo regular válido.
- AC2. Una copia existente cambia su `source_status` de `local` a `tracked`, y de `tracked` a `local` al quitar el tracking conservando su fuente, en el siguiente barrido completo aunque sus bytes no cambien; conserva clave, hash, fecha de ingesta y pin. Falso en la base: la caché ignora cambios sólo del índice Git y no existe apartamiento.
- AC3. Una copia clasificada `tracked` sale del árbol y conteos normales y aparece en un apartado explicativo sin renderizar `action_required`; al volver a `local` se reintegra al árbol normal; la lectura normal de manifiesto/archivos devuelve 409, mientras pin y borrado explícito siguen disponibles. Falso en la base: todas las copias se renderizan y se leen normalmente.
- AC4. Reconstruir o cargar el almacén clasifica las copias antes de publicar su índice: `tracked`, `local` o `unknown`; fuente ausente sin evidencia positiva y errores de consulta producen `unknown` con `source_reason` acotado, nunca una afirmación de autoría. Falso en la base: índice y UI no incluyen clasificación.
- AC5. Manifiesto o recurso con más de un enlace duro se rechaza antes de leer sus bytes; también se rechaza un recurso sustituido por un enlace duro entre validación y hash/copia, sin instalar una copia parcial. Falso en la base: son archivos regulares admitidos.
- AC6. En una tira de pantalla mayor de 520 CSS px, una captura de escritorio ocupa su ancho útil en vez de la mitad, tanto en el HTML generado de nuevo como en el visor compilado. Falso en la base: la regla divide el ancho entre dos.

- AC7. Lista y documento abierto muestran el aviso de origen sin determinar o la etiqueta local sin autoría verificada según el último índice; transiciones `local`/`unknown` con el mismo hash actualizan el aviso sin reingesta. Falso en la base: no existe clasificación ni aviso dinámico.

- AC8. El clasificador de procedencia no ejecuta el callback `core.fsmonitor` configurado en un repositorio aceptado por Git; la consulta conserva el resultado de tracking sin ejecutar ese helper. Falso en el candidato inicial: el probe con runner de producción ejecutó el callback sintético.

Precisiones de AC1/AC4 (exclusión y clasificación): reconocer nombres de índice que sean alias reales del mismo archivo en filesystems sin distinción de mayúsculas; no confundir archivos distintos en filesystems sensibles. La semántica de pathspec no puede alterarse por variables Git heredadas y convertir una ruta versionada en resultado local satisfactorio.

Restricciones de regresión: workspaces sin Git, worktrees y repositorio ancestro siguen soportados; ausencia comprobada de Git no se confunde con fallo; pins y TTL existentes siguen aplicándose también a copias apartadas; no se alteran fuentes, tombstones ni permisos del almacén. La política conserva archivos legibles de otros dueños con un solo enlace. No se debilitan tokens, HMAC, loopback, límites de bytes, `os.Root`, validación de symlinks ni comparación de hashes. No se apilan todas las capturas en pantalla ni se cambia la impresión. Las dos copias del CSS siguen idénticas.

## Entrega y autorización

Sin `Delivery`, `Review` ni `UI verification` declarados en `AGENTS.md`. El usuario resolvió para esta sesión/repositorio el 2026-10-19:

- D4-A (entrega interactiva): al autorizar implementar, rama desde `development`, commits y verificación local; pausa antes de push para validar el resultado. Tras su aceptación, el agente ejecuta push, PR, revisión y merge cuando los gates pasan. Registro por pequeño PR de cierre sin revisión dedicada, integrado por el agente. `master`, release y publicación del sitio quedan fuera.
- D5-A (revisión): `hive-review-code` y `hive-review-security` sobre candidato local, con refutación de hallazgos bloqueantes de código. Su ejecución queda consentida para la futura implementación, no se ejecuta durante planificación.
- D6-A (interfaz): `full`, con `hive-review-ux` y `hive-verify-change` independientes sobre candidato construido; aplicable al resto de la sesión en este repositorio hasta que el usuario lo cambie. No sustituye la validación humana previa al push.

No hay CI declarado en `CONTRIBUTING.md` ni archivos de workflows versionados. Revalidar antes de publicar. Existe `_support/workspace/maintainer.local`: integrar en `development` exige refrescar binario por cambios en `tooling/` y contenido por CSS en `content/`, en ese orden; D4-A (entrega interactiva) incluye ese efecto. No reiniciar el visor cotidiano: reemplazar el binario no cambia un proceso ya activo, por lo que su nueva conducta/asset queda pendiente de su próximo arranque. Verificar y reportar actualización de archivos y activación por separado. No implica release ni publicación del sitio.

Este registro permanece sin commit hasta el cierre por la ruta que se acuerde. Diseño: [design.md](design.md). Tareas, gates y revisión: [tasks.md](tasks.md).
