# Diseño — gh-75 recorrido como datos y servidor `hive view`

## Contexto verificado (base `development` `8ce403c5`, 2026-10-18)

**Recorrido de hoy**
- El orquestador copia [el esqueleto](../../../../content/skills/flow-report/assets/skeleton-walkthrough.html) y lo rellena a mano, ejecuta `self_check.py` y abre la página ([SKILL.md:44-48](../../../../content/skills/flow-walkthrough/SKILL.md)).
- La primera vista rendida reutiliza el archivo con un llenado parcial (comentario del esqueleto, líneas 36-44).
- El lightbox abre una sola imagen, sin navegación (esqueleto, líneas 256-278).

**Video**
- Por defecto en la entrega para validación humana: [SKILL.md:28-38](../../../../content/skills/flow-walkthrough/SKILL.md).
- Lo graba `hive-verify-change` cuando su brief lo pide: [hive-verify-change.md:18](../../../../content/agents/quality/hive-verify-change.md).
- El candidato aceptado regraba lo obsoleto: [ui-gate.md:26](../../../../content/skills/flow-walkthrough/references/ui-gate.md).
- El sitio lo describe en [flows/build.md:75](../../../../site/src/content/docs/flows/build.md) y en [index.mdx:87](../../../../site/src/content/docs/index.mdx).

**Retención**
- Página, imágenes y grabaciones forman una unidad que `flow-close` borra al terminar la sesión: [artifact-layout.md:23-25](../../../../content/skills/workspace-conventions/references/artifact-layout.md), [flow-close/SKILL.md:78](../../../../content/skills/flow-close/SKILL.md), [flows/close.md:56](../../../../site/src/content/docs/flows/close.md).

**Despliegue de recursos**
- El manager recoge todo `assets/`, `scripts/` y `references/` de cada skill (`tooling/management/plan.go:241-246`).
- Solo admite las extensiones de [plan.go:323](../../../../tooling/management/plan.go): `.mjs` y `.css` sí, `.js` no.
- Los enlaces de las skills los comprueba `go test ./tooling/management -run RepositoryCatalogue`.

**Comprobador**
- `self_check.py` solo analiza el HTML estático.
- Admite `<script>` en línea fuera de SVG y un `<script type="application/json">` (`self_check.py:100-108, 212-220`).
- No ve el DOM que construye el JS.

**Manager**
- Despacho en [main.go:50-82](../../../../tooling/cli/main.go), ayuda en `printHelp`, flags con el patrón de `readOnlyOptions` (`main.go:100-122`).
- Directorio de estado resuelto por `management.NormalizeOptions` ([types.go:355](../../../../tooling/management/types.go)).
- `TestMain` de `tooling/cli` (`tui_hosts_test.go:62-82`) solo redirige `management.HiveExecutable`. En pruebas, `os.Executable()` es `cli.test`.
- Hay precedente de grupo de procesos con `Setsid` y muerte del grupo: [agent_browser_runner.go:122-130](../../../../tooling/cli/agent_browser_runner.go).
- `target.Safe` y `target.Canonical` son exportados ([target.go:88-139](../../../../integrations/target/target.go)).
- `parseHiveSection` ([doctor_project.go:245-304](../../../../tooling/cli/doctor_project.go)) y `newGitRunner` (`doctor_project.go:63-96`) viven en `package main`.
- `parseGitHubOrigin` (`project_write.go:310`) alimenta `suggestHiveValues` y no debe generalizarse.
- [deps_test.go:30-39](../../../../tooling/management/deps_test.go) revisa solo `management` y `distribution`.
- Hay precedente de `//go:embed`: [tooling/legacy/scan.go:7-26](../../../../tooling/legacy/scan.go).
- Nada en `management` recorre ni poda la raíz del directorio de estado.

**Instalación**
- `hive` no queda en el `PATH` ([README.md:91-92](../../../../README.md)). Su ruta es `~/.local/share/hive/packages/hive-<version>-<os>-<arch>/bin/hive`.

**Hosts** (documentación oficial, investigación del 2026-10-18)
- Escribir fuera del repositorio pide aprobación en Claude Code (herramientas de archivo), Codex (`workspace-write`) y Cursor; en Grok solo con el sandbox encendido. OpenCode V2 y Pi no preguntan.
- Ningún host documenta que un proceso lanzado por el agente sobreviva a la sesión.
- Solo Claude Code (`CLAUDE_CODE_SESSION_ID`) y Pi (`PI_SESSION_ID`) documentan un ID de sesión legible desde el shell.

**Nombres de proyecto.** Engram sin config cae a `git_remote`, `git_root` y `dir_basename`. Es una observación con engram 3.4.0. `engram projects list` tiene entradas accidentales como `jmartinez`.

**Antecedentes externos.** Herramientas de skills de terceros tienen servidores locales efímeros: uno arma la página desde un payload JSON, con latido y apagado por inactividad; otro es un servidor sin dependencias en loopback, con clave y verificación de `Origin`. Ninguno es persistente para varios proyectos, y no se adapta código de ellos.

## DS1 — Contrato del manifiesto `walkthrough.json` (versión 1)

El agente escribe un archivo por carpeta de trabajo, junto a `images/`. La primera vista y la página final usan el mismo archivo: la final lo reemplaza. Las claves van en inglés y el texto, en el idioma de la sesión. Las rutas son relativas a la carpeta.

La casa canónica para el agente es `content/skills/flow-walkthrough/references/walkthrough-manifest.md`, que crea T2 a partir de este contrato, con un ejemplo completo. Las fixtures son el contrato ejecutable. La prueba de T1 extrae el bloque JSON de esa referencia, cuando existe, y lo valida con el generador, para que el ejemplo no se desvíe.

```json
{
  "schema_version": 1,
  "lang": "es",
  "kind": "final",
  "title": "gh-67 — one-line description",
  "ticket": "gh-67-flow-walkthrough-skill",
  "session": {"id": "string", "host": "claude-code"},
  "candidate": {"branch": "feat/x", "commit": "1a2b3c4d", "base": "development"},
  "created_at": "2026-10-18T22:00:00-06:00",
  "result": "one-sentence verdict",
  "checks": [{"name": "AC1 — ...", "status": "pass", "evidence": "..."}],
  "flows": [{
    "id": "sidebar-navigation",
    "title": "gh-67 - what the person can now do",
    "steps": ["..."],
    "captures": [{"path": "images/a.png", "caption": "...", "alt": "..."}],
    "recording": {"status": "none"}
  }],
  "action_required": ["..."],
  "environment": [{"label": "App", "value": "http://localhost:3000"}],
  "provenance": ["..."]
}
```

**Campos**
- `kind`: `first_view` o `final`. En `first_view`, `checks` puede ir vacío, y `steps` y `recording` se omiten.
- `checks[].status`: `pass`, `fail`, `warn`, `info` o `not_run`.
- `captures[].viewport`: `desktop` o `mobile`, opcional, `desktop` por defecto (ronda de corrección 1). Un flujo con capturas de los dos valores muestra un selector en lugar de duplicar el flujo.
- `flows[].id`: cumple `^[a-z0-9-]{1,64}$` y es único. Es un identificador de datos; el DOM no lo usa como `id`.
- `recording.status`:
  - `none`: no se graba. Es el valor por defecto bajo V1-A.
  - `recorded`: exige `path` (`.mp4`), `duration_seconds` y `candidate`.
  - `not_recorded`: exige `reason`.
  - `stale`: exige `path` y `candidate`.
- Obligatorios: `schema_version`, `lang`, `kind`, `title`, `session`, `candidate`, `created_at`, `flows` (al menos uno, cada uno con `id`, `title` y `captures`).
- Opcionales: `ticket`, `result`, `checks`, `action_required`, `environment`, `provenance`.

**`session.id`**
- Es el ID del host cuando está documentado (`CLAUDE_CODE_SESSION_ID`, `PI_SESSION_ID`).
- Si no lo hay, el agente lo genera como `<host>-<UTC de la primera escritura, AAAAMMDDThhmmssZ>` y lo reutiliza durante toda la sesión.
- Varias sesiones que continúan un ticket comparten `ticket` y difieren en `session.id`.

**Validación.** Es estricta: rechaza campos desconocidos, comprueba tipos y aplica las reglas de rutas de AC2. Cada segmento de ruta cumple `^[A-Za-z0-9._-]+$`. Pone topes: manifiesto de hasta 1 MiB, captura de hasta 20 MiB, grabación de hasta 200 MiB, y como máximo 50 flujos y 100 capturas por flujo.

**Fixtures** (`tests/fixtures/walkthrough/`)
- `valid/final/`: 2 flujos; el primero con 3 capturas y el segundo con 1. Lleva `checks` con 3 `pass`, 1 `warn` y 1 `not_run`, `result`, y texto hostil (`</script><img src=x onerror=alert(1)>`) en `result` y en un paso.
- `valid/first-view/`.
- `valid/extreme/`: títulos de 200 caracteres y 100 capturas en un flujo.
- `invalid/<rule>/`: una carpeta por cada regla de AC2.

Las usan la prueba Python del generador y la prueba Go del servidor, para que los dos validadores no se desvíen.

## DS2 — Generador, plantilla y módulo compartido (fase 1)

**Archivos**
- `content/skills/flow-walkthrough/assets/walkthrough.html`: carcasa con marcadores para el CSS, el módulo y los datos.
- `content/skills/flow-walkthrough/assets/walkthrough.css`.
- `content/skills/flow-walkthrough/assets/walkthrough.mjs`: módulo ES.
- `content/skills/flow-walkthrough/scripts/render_walkthrough.py`: solo biblioteca estándar de Python 3.

**API del módulo**
- `walkthrough.mjs` exporta `renderWalkthrough(root, manifest, {fileBase})` y no tiene efectos al importarse.
- Construye todo bajo `root` con `textContent` y `setAttribute`. Nunca usa `innerHTML` con datos ni `setAttribute('style', …)`.
- `fileBase` antecede a cada ruta del manifiesto: `""` en la página offline y `/files/<id>/` en el servidor.
- El `<dialog>` del carrusel pertenece a `root`.
- Los `id` del DOM son `wt-flow-<índice>`.
- El índice desplaza con `scrollIntoView` y no cambia `location.hash`, para no chocar con el token del servidor.

**CSS**
- `walkthrough.css` está acotado a `.walkthrough` (el `root`), sin reglas sobre `html`, `body` ni selectores globales. Sus tokens, incluido el bloque de impresión, se definen sobre `.walkthrough`.
- La carcasa `walkthrough.html` lleva su propio estilo mínimo para `html` y `body`: fondo oscuro y `color-scheme: dark` en pantalla, claro al imprimir.
- Hereda los tokens oscuros y el bloque de impresión del esqueleto actual.
- Ancho de lectura de unos 760 px y pasos numerados con más interlineado.

**Generador**
1. Valida el manifiesto (DS1).
2. Inserta el CSS en `<style>`.
3. Inserta el módulo en `<script type="module">`, que se ejecuta sobre `file://` al no importar nada, y añade en ese mismo script la llamada `renderWalkthrough(document.querySelector(".walkthrough"), JSON.parse(…textContent), {fileBase: ""})`.
4. Embebe el manifiesto en `<script type="application/json" id="walkthrough-data">`, escapando `<`, `>` y `&` como las secuencias JSON `\u003c`, `\u003e` y `\u0026`.
5. Escribe `walkthrough.html` de forma atómica: archivo temporal y `os.replace`.

La salida de éxito es una línea en stdout con la ruta escrita. Un error va a stderr nombrando el campo, con salida 1.

**Página**
- Secciones, en este orden:
  - Resumen: `result` si existe, recuentos de comprobaciones por estado, número de flujos y capturas.
  - Índice por título de flujo.
  - Tabla de comprobaciones.
  - Flujos, cada uno con pasos, galería y grabación o su marca.
  - Acción requerida.
  - Entorno.
  - Procedencia.
- `first_view` oculta la tabla, los pasos y la grabación.

**Idioma de las etiquetas fijas.** El manifiesto lleva `lang` (`es` o `en`, obligatorio). El módulo trae un diccionario mínimo para los encabezados, las etiquetas y el contador en esos dos idiomas. Otro valor de `lang` muestra las etiquetas en inglés, y queda como límite declarado. El contenido del manifiesto va siempre en el idioma de la sesión.

**Rutas en URL.** El módulo codifica cada segmento de ruta con `encodeURIComponent` antes de anteponer `fileBase`.

**Bloques** (K1-A)
- `walkthrough.mjs` define un registro ordenado de bloques. Cada bloque es `{id, applies(ctx), render(ctx)}`, con un `index(ctx)` opcional; `ctx` lleva el manifiesto, el diccionario de etiquetas, `fileBase` y los ayudantes compartidos de DOM y carrusel.
- Bloques iniciales, en este orden: `header`, `summary`, `index`, `checks`, `flows`, `action`, `environment`, `provenance`. `renderWalkthrough` recorre el registro; no tiene secciones escritas a mano.
- El índice se construye a partir de los bloques que aparecen, no de una lista fija.
- Añadir un bloque: una función de dibujo, una entrada en el registro y, si necesita datos, su campo en el manifiesto con su validación en Python y Go y su línea en la referencia. Un comentario de cabecera en `walkthrough.mjs` lo explica.
- `viewer.mjs` no duplica secciones: monta el recorrido con `renderWalkthrough`, y sus propias regiones (barra lateral, barra superior, avisos) quedan como funciones separadas con una sola responsabilidad.
- El refactor conserva el comportamiento: el DOM generado para cada fixture válida es idéntico antes y después.

**Comprobaciones plegables** (ronda de corrección 3, validación del usuario)
- El bloque `checks` es un `<details>` cerrado al cargar. Su `<summary>` lleva el encabezado y los recuentos por estado. Al imprimir se despliega.

**Galería y selector** (ronda de corrección 1)
- La galería es una tira horizontal con `scroll-snap`, de altura fija para las miniaturas, enfocable (`tabindex=0`, `aria-label` con el número de capturas) y desplazable con las flechas del teclado. Las miniaturas reservan su tamaño (`width` y `height`) para no desplazar la página al cargar.
- Si un flujo tiene capturas `desktop` y `mobile`, encima de la tira aparece un selector Escritorio/Móvil (`role=tablist`, ← y → entre pestañas). Escritorio es la pestaña inicial; las etiquetas salen del diccionario de `lang`.

**Carrusel**
- Recorre las capturas de la vista activa del flujo de la miniatura pulsada.
- Una captura `mobile` se muestra ajustada a una columna de `min(430px, 100%)`, con desplazamiento vertical dentro del diálogo; la barra con el contador, cerrar y las flechas queda fija. Una captura `desktop` se ajusta al viewport como hasta ahora.
- Botones anterior y siguiente, desactivados en los extremos, sin dar la vuelta.
- Teclas ← y →, contador "n / N" y pie con `caption`.
- `Esc` cierra y devuelve el foco a la miniatura.

**Sin dependencias.** No hay referencias remotas, SVG ni atributos `on*`, que es lo que comprueba `self_check.py`. La seguridad del contenido de ejecución la cubre la prueba con navegador de T1.

**Qué se retira de `flow-report`.** `skeleton-walkthrough.html`, su fila (`flow-report/SKILL.md:22`), la mención de la línea 10 y la excepción de la línea 26.

## DS3 — Texto de la skill y política de capturas (fase 1)

**Captura y grabación** (V1-A)
- Capturas siempre.
- Una pasada grabada solo cuando el flujo valida una interacción o una transición que las capturas no muestran (por ejemplo, plegar un menú o arrastrar), o cuando el usuario lo pide.
- El momento de grabación, quién graba, cómo, las precondiciones, `ffprobe`, la marca `stale` y la regrabación en el candidato aceptado (`ui-gate.md:26`) no cambian.
- `SKILL.md:14`, `:28` y `:42` se ajustan para que la grabación deje de darse por hecha.

**Entrega**
- El paso de copiar y rellenar el esqueleto se sustituye por: escribir `walkthrough.json` según [la referencia del manifiesto](../../../../content/skills/flow-walkthrough/references/walkthrough-manifest.md) (nueva), ejecutar `python3 <flow-walkthrough>/scripts/render_walkthrough.py <folder>` y corregir lo que informe. El agente no abre la página ni el visor: da una línea con la ruta absoluta de la página y `hive view open` si el registro funcionó (R1-B).
- La primera vista usa `kind: first_view`.

**Registro en el servidor** (T7)
- Tras generar la página, si `command -v hive` encuentra el binario, ejecutar `hive view register` desde el repositorio. Un fallo o un rechazo del sandbox no bloquea la entrega.
- Si no lo encuentra, el mensaje, en el idioma de la sesión, pide registrar el repositorio una vez e incluye literalmente el comando `hive view register` y la ruta `~/.local/share/hive/packages/hive-<version>-<os>-<arch>/bin/hive` para cuando `hive` no está en el `PATH`. Literales son el comando y la ruta, no la frase. El agente no busca el binario.
- Nunca ejecutar `hive view start`.

**Presupuesto de palabras**
- Tras T2, `SKILL.md` queda en 1.740 palabras o menos.
- Tras T7, en 1.797 o menos.
- El esquema vive en la referencia, no en `SKILL.md`.

**Unidad de retención.** `artifact-layout.md:25` y `flow-close/SKILL.md:78` añaden `walkthrough.json` a la unidad. La copia del servidor queda fuera del alcance de `flow-close`, porque es del usuario y la purga `hive view`.

**Sitio.** `flows/build.md:75`, `flows/close.md:56` e `index.mdx:87` describen en una frase que el recorrido lleva capturas, y video cuando la skill lo indica, con enlace a la skill. T7 añade una frase sobre `hive view`.

**CHANGELOG.** T2 añade la entrada de la fase 1 bajo `[Unreleased]` y T7 la de la fase 2. `VERSION` no cambia (CONTRIBUTING.md:98).

## DS4 — Servidor `hive view` (fase 2)

**Ciclo de vida.** El usuario lo inicia y lo detiene, como `herdr server stop` u `opencode serve`; el agente nunca lo lanza.

| Comando | Efecto |
| --- | --- |
| `hive view start [--port N] [--ttl D]` | Si `viewer/server.lock` está tomado (prueba con `flock` `LOCK_NB`), informa que ya corre, con la URL de `server.json`. Si el lock está libre y queda un `server.json`, es huérfano y lo borra. Después lanza `serve` desacoplado mediante un lanzador inyectable (por defecto `os.Executable()` con `view serve` y `Setsid`, salida a `viewer/server.log`) y espera hasta 5 s a que `health` responda. |
| `hive view serve` | Primer plano. Toma `flock` exclusivo `LOCK_NB` sobre `viewer/server.lock` durante toda su vida, y sale con error si otro servidor lo tiene. Escribe `server.json` (PID, puerto, versión, token, TTL, directorio de estado, inicio) con modo 0600 y lo borra al salir con SIGINT o SIGTERM. |
| `hive view stop` | Si el lock está libre, borra un `server.json` huérfano y termina. Si está tomado, consulta `health` con el token; si el PID y el directorio de estado coinciden, envía SIGTERM y espera hasta 5 s a que el lock se libere. Si el lock está tomado pero `health` no responde, informa el PID de `server.json` sin enviar señales ni borrar nada. |
| `hive view status` | JSON con `running`, `url` (sin token), `pid`, número de repositorios, recorridos y errores de ingesta. |
| `hive view open` | Abre `<url>/#token=<token>` con un abridor inyectable (`open` en macOS, `xdg-open` en Linux). Si no corre, lo informa sin lanzarlo. |
| `hive view register [--dir DIR]` | Añade a `viewer/registry.json` la raíz de DIR: la carpeta que contiene `_support/workspace/` cuando DIR está dentro de ella, sea o no Git (W1-A); si no, la raíz Git. Si ya está, solo lee y sale con 0. No necesita el servidor. |

**Comportamiento común**
- Todos aceptan `--home` y `--state-dir` con el patrón de `readOnlyOptions`.
- `serve` enlaza en `127.0.0.1` con el puerto de `--port`, o uno efímero.
- El intervalo de sondeo (5 s por defecto) y el reloj son inyectables en pruebas.

**Pruebas del lanzador**
- `TestMain` de `tooling/cli` reconoce `HIVE_VIEW_TEST_SERVE=1` y, con ella, ejecuta `run(os.Args[1:])` en vez de la suite.
- Las pruebas de `start` matan el grupo de procesos en `t.Cleanup`, como `agent_browser_runner.go:122-130`.

**Almacén.** Vive en `<state-dir>/viewer/` y no toca `state.json`:

```
viewer/
  lock                 flock between processes (register vs server)
  server.lock          exclusive flock held by the one running server
  registry.json        {"version":1,"repositories":[{"root":"…","added_at":"…"}]}
  server.json          running server; 0600
  server.log
  index.json           rebuildable cache: sessions, pins, tombstones, errors
  sessions/<key>/      walkthrough.json + referenced files, copied
  sessions/<key>/meta.json   {"source":…, "content_hash":…, "ingested_at":…, "pinned":bool}
```

**Claves y contenido**
- `<key>` es el SHA-256 hexadecimal de `source folder + "\x00" + session.id`: una copia por carpeta de origen y sesión.
- `content_hash` es el SHA-256 del manifiesto más, por cada archivo referenciado, su ruta y su SHA-256.
- Si cambia el contenido, la copia de esa clave se reemplaza (temporal más `rename` y borrado de la anterior), arrastrando `pinned` al `meta.json` nuevo. Así una ronda que reescribe una imagen sin tocar el JSON también se refleja (AC11), y una sesión conservada sigue conservada.

**Lápidas**
- Borrar o purgar deja en `index.json` una lápida `{key, content_hash}`.
- La ingesta salta una fuente cuyo `key` y `content_hash` coinciden con una lápida, de modo que lo borrado no vuelve en el siguiente sondeo.
- Si la fuente cambia de contenido, vuelve a entrar.
- La ingesta también salta fuentes cuyo `walkthrough.json` tenga un `mtime` más antiguo que el TTL.
- Las lápidas se purgan cuando su fuente ya no existe.

**Reconstrucción.** `index.json` se reconstruye leyendo `sessions/*/meta.json` si falta o está corrupto. El pin vive en `meta.json` (D1-A). Las lápidas se pierden con la reconstrucción, y lo único que pasa es que una fuente vigente se reingiere una vez.

**Concurrencia**
- Un `sync.Mutex` serializa ingesta, purga, borrado y pin dentro del servidor. `health` no lo toma, para responder aunque una ingesta esté copiando.
- Solo un servidor por almacén: `viewer/server.lock` (ver la tabla de comandos).
- `flock` bloqueante sobre `viewer/lock` coordina con `register` en otro proceso.
- Las carpetas temporales de una copia fallida se borran al iniciar.

**Ingesta**
- Cada intervalo recorre `<root>/_support/workspace/*/walkthrough.json` de cada repositorio registrado.
- Usa `Lstat` sin seguir enlaces, valida (DS1) y copia según las claves.
- Un error queda en `index.json` con la ruta y el motivo, y no detiene el servidor.
- Un repositorio registrado que ya no existe se marca `absent` sin borrar copias.

**Agrupación** (AC12)
- La precedencia vive en `tooling/viewer`.
- `tooling/cli` inyecta dos lectores:
  - `HiveProjectReader func(root string) (string, bool)`, que usa `parseHiveSection`.
  - `GitRemoteReader func(root string) (string, error)`, que usa `newGitRunner`.
- `.engram/config.json` lo lee el propio paquete: solo `project_name`, 64 KiB como máximo.
- La normalización del remoto es una función nueva de `tooling/viewer` (`host/owner/repo`, sin credenciales ni `.git`). No toca `parseGitHubOrigin`.
- Los worktrees comparten remoto y por eso caen en el mismo proyecto.

**Purga.** Al iniciar y cada hora, borra las copias con `ingested_at` más antiguo que el TTL, salvo las conservadas, y deja una lápida. No usa `created_at`, que escribe el agente y no es confiable.

**API HTTP** (solo `127.0.0.1`)

| Ruta | Método | Efecto |
| --- | --- | --- |
| `/` | GET | `index.html` de la aplicación |
| `/assets/{walkthrough.mjs,walkthrough.css,viewer.mjs,viewer.css}` | GET | Archivos embebidos; `.mjs` como `text/javascript` |
| `/api/health` | GET | `{pid, state_dir, version, ttl}`; exige token |
| `/api/index` | GET | Árbol proyecto › ticket › sesión con pin, `absent` y errores |
| `/api/sessions/<key>` | GET | Manifiesto copiado |
| `/files/<key>/<path>` | GET | Archivo de la copia, validado como en AC2; `Content-Type` por extensión |
| `/api/sessions/<key>/pin` | POST, DELETE | Conservar o dejar de conservar |
| `/api/sessions/<key>` | DELETE | Borrar la copia y dejar la lápida |

**Seguridad** (lee [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md))
- **`Host`.** Toda petición valida `127.0.0.1:<port>` o `localhost:<port>`, contra el *DNS rebinding*.
- **Escrituras, lecturas y `health`** (S2-A). Todo `/api/*` exige `X-Hive-Token` con el token de `server.json`, y las escrituras además un `Origin` igual al del servidor. `/files` no puede llevar cabeceras desde `<img>`, así que su ruta es `/files/<key>/<cap>/<path>`, con `<cap>` = HMAC-SHA256(token, key) en hexadecimal, que el índice entrega a la página autenticada como `fileBase`. Si falta o no coincide, 403. `/` y `/assets/*` no llevan datos y no lo exigen.
- **Una sola apertura por carpeta** (H1). La ingesta abre `<root>/_support/workspace` una vez como `os.Root`, después cada carpeta con `OpenRoot`, y usa ese mismo descriptor para validar, calcular el hash y copiar.
- **Token en la página.** La página lo lee de `location.hash`, lo guarda en `sessionStorage` y lo quita de la URL con `history.replaceState`. Sin token muestra solo el aviso para reabrir con `hive view open`. `status` nunca lo imprime.
- **Errores acotados** (H4). El nombre de campo y el motivo de un error de ingesta se recortan, y hay un tope de errores por repositorio.
- **Cabeceras.** `Content-Security-Policy: default-src 'self'; img-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'` y `X-Content-Type-Options: nosniff`.
- **Sin código en línea.** `index.html` no tiene scripts ni atributos `style` en línea, para que la CSP no necesite `'unsafe-inline'`.
- **Datos no confiables.** Los manifiestos son entrada no confiable: validación estricta, `encoding/json` hacia el cliente y solo `textContent` en el cliente.
- **Borrado acotado.** `<key>` debe ser hexadecimal de 64 caracteres, se resuelve dentro de `sessions/` y nunca toca rutas de repositorios.

**Dependencias.** `tooling/viewer` usa solo la biblioteca estándar y entra en la lista de [deps_test.go:30](../../../../tooling/management/deps_test.go). Embebe `static/` con `//go:embed`.

**Copias del módulo.** `tooling/viewer/static/walkthrough.mjs` y `walkthrough.css` son copias exactas de los de `content/skills/flow-walkthrough/assets/`. `go:embed` no alcanza fuera del paquete, así que una prueba Go compara los bytes (AC15).

## DS5 — Vista del servidor (UI)

**Aplicación.** Nueva, `hive view` en el navegador local. No hay guía de UI en el repositorio. La referencia es la vista de Cursor que mostró el usuario, y los tokens oscuros vienen del esqueleto actual. Composición a 1440 px:

```
┌──────────────┬───────────────────────────────────────┐
│ tricell-hive │ gh-75 — título             [Conservar]│
│  ▾ gh-75-…   │ Resultado: …    pass 3 · warn 1 · …   │
│    · 18 oct  │ Índice: Comprobaciones · Flujo 1 · …  │
│      claude  │ ── Comprobaciones ─────────────────── │
│  ▸ gh-67-…   │ AC1 …                          pass   │
│  ▸ sin ticket│ ── Flujo 1 ────────────────────────── │
│ carvento     │ 1. …  2. …                            │
│  ▸ cvt-12-…  │ [img][img][img]  → carrusel 2 / 3     │
│ TTL 14 d · 1⚠│ Acción requerida · Entorno            │
└──────────────┴───────────────────────────────────────┘
```

**Barra lateral**
- Proyectos y tickets son grupos con encabezado y un botón de despliegue (`aria-expanded`). Las sesiones son botones en una lista.
- Teclado: ↑ y ↓ recorren los botones visibles, Enter abre la sesión o despliega el grupo.
- La sesión abierta lleva `aria-current="true"`.
- Cada sesión muestra fecha, host, y un botón de menú con "Conservar" o "Dejar de conservar" y "Borrar". El botón de menú se alcanza con Tab desde la sesión; ↑ y ↓ solo recorren sesiones y encabezados.
- Borrar abre un `<dialog>` de confirmación que nombra el recorrido.
- Tras borrar, el foco va a la siguiente sesión del mismo ticket; si el ticket queda vacío desaparece del índice y el foco va a la parada visible anterior.
- Al pie, el TTL vigente y un contador de errores de ingesta que despliega la lista con ruta y motivo.

**Vista principal**
- Sin sesión seleccionada, abre la más reciente.
- Monta el recorrido con `renderWalkthrough(main, manifest, {fileBase})`, con el `fileBase` que entrega el índice (S2-A).

**Estados**
- Cargando: un texto de estado.
- Sin repositorios registrados: guía de `hive view register`.
- Sin recorridos.
- Error de ingesta, en el pie.
- Repositorio `absent`: etiqueta en el proyecto.
- Sin token: solo una tarjeta para reabrir con `hive view open`; no se pide ni se muestra ningún dato (S2-A).
- Servidor caído: aviso con el comando `hive view start`.
- Nombres largos: truncados con `title` completo.
- El índice se refresca cada 10 s sin perder la selección ni el foco. Si el refresco elimina la sesión abierta (purga o borrado desde otra pestaña), la vista muestra un aviso y abre la más reciente.

**Responsive.** Por debajo de 720 px la barra se vuelve un panel que abre un botón. A 320 px no hay desplazamiento horizontal.

**Accesibilidad.** Foco visible, contraste AA con los tokens existentes, y los diálogos devuelven el foco al cerrar.

## Alternativas descartadas que explican decisiones

- **El agente escribe directo en la ruta global:** exige permisos en tres o cuatro hosts (S1-A).
- **El agente avisa al servidor por localhost:** el sandbox por defecto de Codex bloquea la red (D2).
- **SQLite:** añade una dependencia para un volumen que un índice JSON cubre (D1-A).
- **Una instantánea por cada versión del manifiesto:** la sesión del sitio Ion habría generado 10 entradas por sesión. Una copia por fuente y sesión mantiene la barra legible.
- **Añadir `.js` a las extensiones del manager:** obliga a recompilar antes de desplegar contenido; `.mjs` ya está admitida.

## Supuestos y límites

- No se probó en un agente real que `hive view register` funcione bajo el sandbox de cada host. La skill trata el fallo como no bloqueante, y la página de la fase 1 existe igual.
- El orden de resolución de Engram es una observación. El servidor solo usa su `project_name` como nombre visible de respaldo.
- Solo se probó macOS. Linux usa `xdg-open` y el directorio de estado XDG ya existente.
- No se verificó que `agent-browser` abra `file://` sin restricciones. Si no puede, T1 sirve la carpeta de la fixture con `python3 -m http.server` en `127.0.0.1` para la prueba renderizada.
