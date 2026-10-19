# Recorrido de UI como datos: manifiesto JSON con plantilla fija, capturas por defecto y servidor `hive view`

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · fusionado en `development` por [#76](https://github.com/JhonHawk/tricell-hive/pull/76) (`9dca1e6d`) el 2026-10-19; registro archivado |
| Tracker · GitHub Issues | • [#75 — Recorrido de UI como datos: manifiesto JSON con plantilla fija, capturas por defecto y visor persistente opcional](https://github.com/JhonHawk/tricell-hive/issues/75) |
| Git | E1-A interactivo · rama `feat/gh-75-walkthrough-manifest-viewer` desde `development` `8ce403c5` · un commit por tarea verificada · parada antes del push para la validación del usuario |
| Verificación | `go vet ./...` · `go test ./...` · `go test -race ./tooling/viewer ./tooling/cli` · pruebas de skills · `pnpm -C site build` · puerta de UI `full` · `hive-review-security` · `/code-review` · recorrido humano |
| Siguiente paso | Ninguno: registro archivado |

## Petición

> En este mismo proyecto, trabajamos la actualización de la pagina de hive, y se uso la skill de presentación del flujo, lo relevante es lo siguiente:
>
> - Cuanto tiempo se invirtio en generarlo
> - Era necesario para esta vista o bastaba con imagenes actualizadas, no siempre es necesario el video
> - Importante considerar el gasto en razonamiento y tokens para editar el html, sobre esto ultimo se me esta ocurriendo que se convierta en una mini-app, muy sencilla, a la que se le cargue un json con las rutas de los valores y la data, y mejoprar el formato de lectura:
>
> > Las imagenes en carrusel [Image #1]
> > Mejorar la presentación y formato de lectura [Image #2]
> > Al ser una app ejecutandose, taaaal vez valga la pena que NO se ejecute N cantidad de veces por cada sesión y que podamos tener una barra lateral por proyecto / sesión [Image #3] y que solo los agentes manden el json que se necesita.
>
> Haz research completo sobre esto

> - Esto talvez implique que esta UI se quede abierta aun después de que la sesión termino, cómo podemos manejar esto? Una ruta global de assets en sistema que se purge cada X tiempo? Que si la captura es valiosa para el agente la conserve como duplicado (considerando que la captura se eliminara eventualmente), esto se agruparia por proyecto (compatible con las reglas de engram tal vez, usar el nombre de ese config), y sesiones, cada sesion con su json, pero qué pasaria con diferentes tickets trabajados en la misma sesión como continuación?
> - Permitimos que en la UI permita tambien borrar estas carpetas de forma anticipada desde la barra izquierda?
> - Que pasa si no tienen config de engram para nombrar?

> Con el tema de mantenerlo en background no me refiero a que el agente levante un background, si no que se mantenga como un server, como lo hace herdr u opencode2 por ejemplo, claro que si habra comandos para iniciar/detener y talvez necesite un sqlite para mantener algun registro

> (Respuesta a D2) Talvez debemos ampliar el cli de hive para que pueda hacerse desde ahi, y describirlo estos comandos en la skill de walktrought

Las imágenes de referencia son la vista de resultado de Cursor Cloud Agents: tabla de validación, reproductor de video, miniaturas que abren un carrusel con pie y contador ("3 of 3") y barra lateral con las sesiones.

El issue #75 no declara criterios de aceptación propios: resume la investigación del 2026-10-18 (leído el 2026-10-18).

**Decisiones del usuario (2026-10-18):**

- **V1-A:** capturas por defecto; video solo cuando lo que se valida es una interacción o una transición, o cuando el usuario lo pide. Reemplaza D2-A de gh-67 (video por defecto en cada validación humana). No cambia cuándo se graba ni la regrabación de lo obsoleto en el candidato aceptado.
- **T1-A:** abrir el issue #75.
- **R2-A:** un plan en dos fases, primero el JSON y la plantilla, después el servidor. Reabre la exclusión DS5 de gh-67 ("Servidor local para ver la página", `changes/archive/2026-10-18-gh-67-flow-walkthrough-skill/design.md:68`).
- **S1-A:** el agente escribe en el repositorio (`_support/workspace/`) y el servidor copia a su propio almacén.
- **D1-A:** el registro del servidor es un índice JSON reconstruible, sin SQLite.
- **D2 (respuesta libre):** el agente registra el repositorio con un comando del CLI de `hive`, y la skill lo documenta.
- **R1-B (2026-10-19):** el agente nunca abre la página, el visor ni un navegador para mostrárselos al usuario; da una línea con la ruta de la página y `hive view open` cuando el registro funcionó. Reemplaza la preferencia del 2026-10-18.
- **W1-A (2026-10-19):** el registro acepta workspaces cuya raíz no es Git (como `comsys/cam` y `comsys/carvento`).
- **S2-A (2026-10-19):** las lecturas del visor también exigen el token.
- **K1-A (2026-10-19):** las vistas se preparan como bloques mediante un registro interno; el contrato JSON no cambia.
- **Aclaración del servidor:** es un servidor de larga vida que el usuario inicia y detiene (como `herdr` u OpenCode V2), nunca un proceso que lance el agente.

## Objetivo

Hoy, en cada recorrido de validación, el orquestador copia un esqueleto HTML de 14,5 KB (6,8 KB de CSS), lo rellena a mano, pasa `self_check.py` y repite en HTML los pasos que ya escribió en el chat (`content/skills/flow-walkthrough/SKILL.md:42-44`). En la sesión del sitio Ion (Codex `01a151c9-adf7-7471-aed1-2ca176efe472`, 2026-10-18) la página se escribió 10 veces: unos 12.700 tokens de salida y unos 150 s. El video se grabó dos veces sin influir en ninguna decisión. La página muere al cerrar la sesión (`artifact-layout.md:25`) y no deja historial.

La fase 1 convierte la página en datos. El agente escribe `walkthrough.json` siguiendo una referencia de la skill, y un script determinista genera `walkthrough.html` con una plantilla fija: carrusel de capturas, resumen arriba y formato de lectura mejorado. La skill pasa a capturas por defecto (V1-A). La fase 2 añade `hive view`, un servidor local de larga vida que el usuario inicia y detiene. Copia los recorridos de los repositorios registrados a su almacén, los muestra con una barra lateral por proyecto, ticket y sesión, los purga por antigüedad salvo los conservados y permite borrarlos desde la barra.

## Alcance y aceptación

**Incluye**
- Fase 1: el contrato del manifiesto con su referencia, el generador, la plantilla con su módulo JS y su CSS, la política de capturas V1-A, la unidad de retención con `walkthrough.json`, el sitio y la entrada de CHANGELOG de la fase.
- Fase 2: el paquete `tooling/viewer` (solo biblioteca estándar), el subcomando `hive view` (`start`, `stop`, `status`, `open`, `register`, `serve`), la documentación del manager, `README.md` y la entrada de CHANGELOG de la fase.

**Excluye**
- Arranque automático al iniciar sesión (launchd o systemd): exige escribir configuración global; queda como mejora posterior.
- Avisos del agente al servidor por localhost: el sandbox por defecto de Codex bloquea la red.
- Entrada en el menú de la TUI: un servidor que se inicia y se detiene no encaja como vista.
- SQLite (D1-A).
- Windows: el manager no compila allí hoy.
- Publicar recorridos: retener no autoriza publicar.
- Cambiar la lista de extensiones que despliega el manager: el JS usa `.mjs`, que ya está admitida (`tooling/management/plan.go:323`).

**Restricciones**
- La suite existente sigue pasando.
- `tooling/management` y `tooling/distribution` siguen sin dependencias externas.
- Ningún comando ni regla nueva hace que el agente lance el servidor.
- El servidor nunca borra ni modifica archivos de un repositorio.
- `content/skills/flow-walkthrough/SKILL.md` no crece respecto de la base (1.797 palabras).
- Identificadores, rutas, claves JSON y comentarios en inglés.
- Las páginas de `site/` describen y enlazan la skill, sin copiar sus condiciones.

**Fase 1**

- AC1. `python3 content/skills/flow-walkthrough/scripts/render_walkthrough.py <folder>` lee `<folder>/walkthrough.json` y escribe `<folder>/walkthrough.html`, que pasa `content/skills/flow-report/scripts/self_check.py` y no tiene referencias remotas. *Falso en la base cuando* el script no existe.
- AC2. Ante un manifiesto inválido, el generador termina con código distinto de cero, nombra el campo y no escribe la página. Son inválidos: un campo desconocido, un campo obligatorio ausente, `schema_version` distinto de 1, un `flows[].id` repetido o fuera de `^[a-z0-9-]{1,64}$`, una ruta absoluta, con `..`, fuera de la carpeta, con un segmento fuera de `^[A-Za-z0-9._-]+$` o que es un enlace simbólico, una captura que no es `.png`, `.jpg`, `.jpeg` o `.webp`, una grabación que no es `.mp4`, o un archivo referenciado que no existe. *Falso en la base cuando* no hay validador.
- AC3. En la página generada, una captura abre un carrusel con las capturas de la vista activa de su flujo: botones anterior y siguiente, teclas ← y →, contador "n / N" y pie con el `caption`. En la primera captura, anterior está desactivado; en la última, siguiente. `Esc` cierra y el foco vuelve a la miniatura. *Falso en la base cuando* el lightbox del esqueleto solo abre y cierra una imagen.
- AC4. La página generada abre con un resumen: el `result` si existe, los recuentos de comprobaciones por estado, el número de flujos y de capturas, y un índice que enlaza cada flujo por su título. *Falso en la base cuando* el esqueleto no tiene resumen ni índice.
- AC5. El manifiesto va embebido como dato inerte. Un `</script><img src=x onerror=…>` en cualquier campo de texto aparece como texto en la página: no se crea ningún `img[src="x"]`, no se ejecuta nada y el HTML estático no tiene atributos `on*`. *Falso en la base cuando* no hay embebido de datos.
- AC20. En la página generada, la galería de cada flujo es una tira con desplazamiento horizontal: la altura de la página no crece con el número de capturas. Un flujo con capturas `desktop` y `mobile` muestra un selector Escritorio/Móvil que cambia la tira, y el carrusel recorre solo la vista activa. Una captura `mobile` se abre en el carrusel ajustada a una columna de teléfono, con desplazamiento vertical dentro del diálogo. *Falso en la base cuando* la galería del esqueleto es una rejilla que crece en vertical con cada captura, sin selector ni presentación móvil (revisión por la ronda de corrección 1, 2026-10-18).
- AC21. Cada sección de la página (resumen, índice, comprobaciones, flujos, acción requerida, entorno, procedencia) es un bloque registrado con la misma interfaz (identificador, condición para aparecer y función de dibujo) en una lista ordenada de `walkthrough.mjs`, y el visor monta la página por esa misma vía. Añadir un bloque es una función y una entrada en la lista. *Falso en la base cuando* no existe ese registro (K1-A, 2026-10-19). Que el DOM generado no cambió con el refactor es evidencia de T8, ya cumplida en su verificación, no parte del criterio.
- AC22. `flow-walkthrough/SKILL.md` prohíbe abrir la página, el visor o un navegador para mostrárselos al usuario y pide una línea con la ruta absoluta de la página, más `hive view open` cuando el registro funcionó. *Falso en la base cuando* `SKILL.md:48` dice "Open the local `walkthrough.html` in the system's default viewer" (R1-B).
- AC6. `flow-walkthrough/SKILL.md` indica escribir `walkthrough.json` según `references/walkthrough-manifest.md` y ejecutar `render_walkthrough.py`, y ya no indica copiar ni rellenar el esqueleto. `skeleton-walkthrough.html`, su fila y sus menciones desaparecen de `flow-report`. *Falso en la base cuando* `SKILL.md:44` dice "copy [the walkthrough skeleton]".
- AC7. `flow-walkthrough/SKILL.md` establece capturas por defecto y video solo para una interacción o transición que las capturas no muestran, o a petición del usuario. `site/src/content/docs/flows/build.md` e `index.mdx` dejan de describir un video por flujo por defecto. *Falso en la base cuando* `SKILL.md:32` dice "Record one clean pass per flow" para todo recorrido e `index.mdx:87` dice "a short video per flow".
- AC8. `artifact-layout.md` y `flow-close/SKILL.md` incluyen `walkthrough.json` en la unidad de retención del recorrido. *Falso en la base cuando* la unidad solo nombra la página, las imágenes y las grabaciones.

**Fase 2**

- AC9. `hive view register [--dir DIR]` añade al registro del servidor la raíz de DIR (por defecto el directorio actual). Si DIR está dentro de `<raíz>/_support/workspace/`, la raíz es `<raíz>`, sea o no un repositorio Git, como los workspaces de varios repositorios (W1-A, 2026-10-19). Si no, es la raíz Git de DIR. Una segunda ejecución no escribe nada: el `mtime` del archivo no cambia. Fuera de un repositorio Git y de un `_support/workspace/` falla con un mensaje. *Falso en la base cuando* `hive view` es un comando desconocido.
- AC10. `hive view start` deja un servidor desacoplado en `127.0.0.1` y guarda su PID, puerto y versión en el almacén. `status` informa que está corriendo, con la URL y sin el token. Un segundo `start` informa que ya corre y no lanza otro proceso. `stop` lo termina y `status` informa que está detenido. *Falso en la base cuando* no existe el subcomando.
- AC11. Con el servidor corriendo, un `walkthrough.json` válido en un repositorio registrado aparece en el almacén dentro de dos intervalos de sondeo, con el manifiesto y los archivos que referencia. Si cambia el manifiesto o un archivo referenciado, la copia de esa sesión se reemplaza. Un manifiesto inválido o una ruta que escapa de la carpeta se rechaza, se informa como error en el índice y no detiene el servidor. *Falso en la base cuando* no hay servidor.
- AC12. El índice (`GET /api/index`) agrupa por proyecto, ticket y sesión. La clave de proyecto es el remoto Git normalizado, o la ruta de la raíz si no hay remoto. El nombre visible sale de `Project:` de `## Hive`; si falta, de `project_name` de `.engram/config.json`; si tampoco, del último segmento del remoto o del nombre de la carpeta. El ticket es `ticket` del manifiesto, o el grupo "sin ticket" si falta. *Falso en la base cuando* no hay índice.
- AC13. Las copias cuya hora de ingesta supera el TTL (14 días por defecto, configurable con `--ttl` en `start` y `serve`) se purgan al iniciar y periódicamente, salvo las conservadas. Una copia purgada o borrada no se vuelve a ingerir mientras su contenido no cambie. Conservar persiste entre reinicios. *Falso en la base cuando* no hay almacén.
- AC14. Borrar desde la barra lateral elimina solo la copia del almacén; los archivos del repositorio quedan intactos. Una petición que escribe (borrar, conservar, dejar de conservar) sin el token del servidor, o con un `Origin` o `Host` ajeno, recibe 403. Leer el índice, un manifiesto o un archivo copiado también exige el token: `/api/*` con `X-Hive-Token`, y `/files` con una capacidad por sesión derivada del token; sin ellos, 403 (S2-A, 2026-10-19). *Falso en la base cuando* no hay servidor.
- AC15. `tooling/viewer/static/walkthrough.mjs` y `walkthrough.css` son copias exactas de los de `content/skills/flow-walkthrough/assets/`, y una prueba Go lo comprueba byte a byte. *Falso en la base cuando* ninguno de los dos archivos existe.
- AC16. `tooling/viewer` no importa ningún módulo externo: solo la biblioteca estándar y paquetes del propio repositorio, la misma regla que `tooling/management/deps_test.go` aplica a `management` y lo comprueba (texto corregido tras la revisión de seguridad, I2). *Falso en la base cuando* el paquete no existe y la prueba no lo revisa.
- AC17. `hive view open` abre la URL del servidor con el abridor del sistema (`open` en macOS, `xdg-open` en Linux). Si el servidor no corre, lo informa sin lanzarlo. *Falso en la base cuando* no existe el subcomando.
- AC18. `flow-walkthrough/SKILL.md` indica:
  - ejecutar `hive view register` después de generar la página cuando `command -v hive` lo encuentra;
  - que un fallo no bloquea la entrega;
  - que, sin `hive` en el `PATH`, el mensaje (en el idioma de la sesión) incluye literalmente el comando `hive view register` y la ruta del binario de la instalación, sin buscarlo;
  - que nunca se ejecuta `hive view start`.

  `README.md` y `_support/docs/architecture/deployment-manager.md` describen `hive view`. *Falso en la base cuando* ninguno de los tres lo menciona.
- AC19. La vista del servidor muestra:
  - la barra lateral proyecto › ticket › sesión, con conservar y borrar (con confirmación) por sesión;
  - el recorrido montado con `walkthrough.mjs`;
  - los estados sin repositorios, sin recorridos, error de ingesta, repositorio ausente, y sin token (aviso para reabrir con `hive view open`, sin mostrar datos; S2-A);
  - teclado en la barra (flechas y Enter), `aria-current` en la sesión activa y foco visible.

  *Falso en la base cuando* no hay vista.

## Entrega

Respuestas del usuario (2026-10-18). `AGENTS.md` no declara `Delivery`, `Review` ni `UI verification`.

- **L1-A:** las etiquetas fijas de la plantilla salen del campo `lang` (`es` o `en`), con un diccionario mínimo en el módulo (DS2).
- **E1-A, interactivo:**
  - Rama `feat/gh-75-walkthrough-manifest-viewer` desde `development` `8ce403c5`, sin otras ramas abiertas (comprobado el 2026-10-18).
  - Un commit local por tarea verificada.
  - Parada antes del push, con el candidato construido disponible para que el usuario valide y pruebe.
  - Tras su validación: push, PR a `development`, `/code-review`, atención de hallazgos y merge cuando las comprobaciones pasen.
  - El registro del cambio se cierra con un PR pequeño sin revisión dedicada, que fusiona el agente.
  - Despliegue local (recompilar el binario y `hive update`) solo a petición, como sigue el usuario en Hive.
  - Sin release ni promoción a `master`.
- **RV1-A:** `/code-review` de Claude Code sobre el diff del candidato antes del push. `hive-review-security` corre aparte por la regla de fronteras.
- **U1-A, `UI verification` `full`:** `hive-review-ux` y `hive-verify-change` sobre el candidato construido, y en el candidato aceptado tras rondas de corrección. Vale para el resto de la sesión en este repositorio.
