# Tareas — gh-75 recorrido como datos y servidor `hive view`

Base: `development` `8ce403c5`, confirmada al iniciar el build el 2026-10-18 en la rama `feat/gh-75-walkthrough-manifest-viewer`. #75 asignado al usuario como marca de en progreso, que es la convención del repositorio.

Orden: T1 → (T2 ∥ T3) → T4 → T6 → T5 → T7 → T8 (añadida por K1-A). La fase 1 (T1, T2) sirve sola aunque la fase 2 se retrase.

## Fase 1 — recorrido como datos

### T1 — Contrato ejecutable, generador, plantilla y módulo

- [x] verified again at the accepted-candidate gate (AC1–AC5, AC20 `met`)

**Closes:** AC1, AC2, AC3, AC4, AC5, AC20.

**Depends on:** ninguna.

**Locations:**
- `content/skills/flow-walkthrough/assets/walkthrough.html`, `walkthrough.css` y `walkthrough.mjs` (nuevos).
- `content/skills/flow-walkthrough/scripts/render_walkthrough.py` (nuevo).
- `tests/fixtures/walkthrough/` (nuevo, ver DS1).
- `tests/skills/walkthrough_render_test.py` (nuevo).
- Referencias de estilo: `content/skills/flow-report/assets/skeleton-walkthrough.html` (tokens, impresión) y `content/skills/flow-report/scripts/self_check.py`.

**Execution:** delegada a `hive-build-frontend`. La interfaz está fijada en DS1 y DS2, y no comparte escrituras.

**Test approach:** tdd.

**Changes:** primero las fixtures de DS1, después el generador y el módulo según DS2.

**Verification:**
- `python3 -m unittest discover -s tests/skills -p 'walkthrough_render_test.py'` pasa. Cubre:
  - `valid/final` genera una página que pasa `self_check.py` (AC1).
  - Cada carpeta de `invalid/` sale con 1, nombra el campo y no deja `walkthrough.html` (AC2).
  - Si existe `content/skills/flow-walkthrough/references/walkthrough-manifest.md`, su bloque JSON de ejemplo pasa el validador.
  - En el HTML estático no hay atributos `on*` fuera del bloque de datos, y el bloque contiene `\u003c/script\u003e` y ningún `</script>` literal antes de su cierre (AC5, estático).
- Prueba renderizada de `valid/final` con `agent-browser`, según [browser automation](../../../../content/skills/flow-build/references/browser-automation.md), a 1440×900 y 390×844. Si `file://` no abre, se sirve con `python3 -m http.server --bind 127.0.0.1`.
  - Clic en la segunda miniatura del flujo 1 → el diálogo muestra "2 / 3" y su `caption`.
  - El botón siguiente → "3 / 3", con siguiente desactivado.
  - `←` → "2 / 3"; anterior dos veces → "1 / 3", con anterior desactivado.
  - `Esc` cierra y el foco vuelve a la miniatura.
  - En el flujo 2, el diálogo muestra "1 / 1" con los dos botones desactivados (AC3).
  - El resumen muestra el `result`, "pass 3", "warn 1", "not_run 1", "2 flujos" y "4 capturas", y el índice enlaza los dos títulos y desplaza a cada flujo (AC4).
  - El texto hostil aparece literal en el resumen y en el paso, `document.querySelectorAll('img[src="x"]').length` es 0 y no se abre ningún diálogo `alert` (AC5, ejecución).
- Prueba renderizada de `valid/viewports` (flujo con 3 capturas `desktop` y 2 `mobile` de 390×1600): con `valid/extreme` (100 capturas) la altura de la tira es la de una fila de miniaturas y `scrollWidth > clientWidth`; el selector cambia a Móvil y el carrusel muestra "1 / 2"; la captura móvil abierta ocupa como máximo 430 px de ancho y el diálogo se desplaza en vertical con la barra visible (AC20).
- Prueba renderizada de `valid/first-view`: sin tabla, pasos ni grabación. Y de `valid/extreme`: sin desplazamiento horizontal a 320 px. Ninguna cierra un criterio; son comprobaciones de estado.
- Las capturas van a la carpeta de trabajo de la primera vista, según [flow-walkthrough](../../../../content/skills/flow-walkthrough/SKILL.md).

### T2 — Skill, referencia del manifiesto, política de capturas y retención

- [x] verified again at the accepted-candidate gate (AC6, AC7, AC8, AC22 `met`)

**Closes:** AC6, AC7, AC8, AC22.

**Depends on:** T1 (rutas y nombre del generador).

**Locations:**
- `content/skills/flow-walkthrough/SKILL.md`, líneas 3, 14, 24, 28, 32-38, 42 y 44-48.
- `content/skills/flow-walkthrough/references/walkthrough-manifest.md` (nuevo).
- `content/skills/flow-report/SKILL.md`, líneas 10, 22 y 26.
- `content/skills/flow-report/assets/skeleton-walkthrough.html` (se borra).
- `content/skills/workspace-conventions/references/artifact-layout.md:23-25` y `content/skills/flow-close/SKILL.md:78`.
- `site/src/content/docs/flows/build.md:75`, `flows/close.md:56` e `index.mdx:87`.
- `CHANGELOG.md` (`[Unreleased]`).
- Repaso sin cambio previsto, para confirmar que no contradicen V1-A: `content/guidance/global.md:78`, `content/agents/quality/hive-verify-change.md:18`, `content/agents/review/hive-review-ux.md:14`, `content/skills/flow-walkthrough/references/ui-gate.md:25-26`, `content/skills/flow-build/SKILL.md:82`, `content/skills/flow-build/references/verification.md:63`, `tests/pilot/`, `tests/skills/`.

**Execution:** hilo principal. Es texto de reglas acoplado a la revisión de toda la conducta que exige AGENTS.md (Guidance structure).

**Test approach:** check: búsquedas, lecturas y pruebas nombradas abajo.

**Changes:** aplicar DS3 y escribir la referencia del manifiesto desde DS1, con un ejemplo completo. Anotar en el progreso los lugares revisados para el PR, y el tamaño de `global.md`, que no cambia.

**Verification:**
- `rg -n "copy \[the walkthrough skeleton\]|skeleton-walkthrough|walkthrough asset" content site/src` → sin resultados (AC6).
- `rg -n "render_walkthrough.py|walkthrough-manifest.md" content/skills/flow-walkthrough/SKILL.md` → ambos presentes (AC6).
- `wc -w content/skills/flow-walkthrough/SKILL.md` → 1.740 o menos tras T2; 1.797 o menos tras T7 y la ronda 4 (hoy 1.776).
- Lectura de "Give the opening command" en `SKILL.md`, y `rg -n "default viewer|#open-the-page" content site/src` sin resultados (AC22, R1-B; añadido en la revisión de la ronda 4).
- Lectura de `SKILL.md`: capturas por defecto y video solo para una interacción o transición, o a petición. Lectura de `flows/build.md` e `index.mdx`: no describen un video por flujo por defecto y enlazan la skill (AC7).
- `rg -n "walkthrough.json" content/skills/workspace-conventions/references/artifact-layout.md content/skills/flow-close/SKILL.md` → una línea en cada uno (AC8).
- `go test ./tooling/management -run RepositoryCatalogue` pasa (enlaces y recursos desplegables, incluido `.mjs`).
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa, incluida la validación del ejemplo de la referencia, y `pnpm -C site build` pasa.

## Fase 2 — servidor `hive view`

### T3 — Núcleo del servidor: validación, registro, almacén, ingesta, índice y purga

- [x] verified again at the accepted-candidate gate (AC11, AC12, AC13, AC16 `met`)

**Closes:** AC11, AC12, AC13, AC16.

**Depends on:** T1 (fixtures).

**Locations:**
- `tooling/viewer/manifest.go`, `registry.go`, `scan.go`, `store.go` e `index.go`, con sus `_test.go` (nuevos).
- `tooling/management/deps_test.go:30` (añadir `"../viewer"`).

**Execution:** delegada a `hive-build-backend`. Es un paquete nuevo con el contrato fijado en DS1 y DS4.

**Test approach:** tdd.

**Changes:**
- DS4: almacén, claves, lápidas, concurrencia, ingesta, agrupación y purga.
- Reloj e intervalo inyectables.
- Repositorios Git creados en `t.TempDir()`.

**Verification:**
- `go test ./tooling/viewer ./tooling/management -run 'NoExternal|Manifest|Registry|Scan|Store|Index|Purge'` pasa. Cubre:
  - Cada fixture inválida rechazada; un enlace simbólico en `images/` rechazado y listado como error, con el escaneo siguiendo (AC11).
  - `valid/final` copiado con sus archivos (AC11).
  - Reescribir solo `images/a.png` cambia `content_hash` y reemplaza la copia (AC11).
  - La agrupación con y sin remoto, con y sin `Project:` (lector inyectado), con y sin `.engram/config.json`, y un worktree en el mismo proyecto (AC12).
  - Con el reloj adelantado 15 días, una copia no conservada se purga y no vuelve en el sondeo siguiente; una conservada sobrevive a un `index.json` borrado y a un cambio de contenido de su fuente (AC13).
  - `deps_test` falla si `tooling/viewer` importa un módulo externo (AC16).
  - Las pruebas de la ronda de seguridad pasan: carpeta cambiada por un enlace entre validar y copiar no copia nada de fuera (H1), errores de ingesta acotados en longitud y número (H4), y la clave de una raíz no Git es su ruta aunque viva dentro de otro repositorio (W1-A).

### T4 — Servidor HTTP, API, seguridad y copias del módulo

- [x] verified again at the accepted-candidate gate (AC14, AC15 `met`)

**Closes:** AC14, AC15.

**Depends on:** T3, T1.

**Locations:**
- `tooling/viewer/server.go` y `server_test.go` (nuevos).
- `tooling/viewer/static/` (nuevo): `walkthrough.mjs` y `walkthrough.css` copiados de T1, y `index.html`, `viewer.mjs` y `viewer.css` como carcasa mínima que T5 completa.
- `tooling/viewer/assets_test.go` (nuevo).

**Execution:** delegada a `hive-build-backend`.

**Test approach:** tdd con `httptest`.

**Changes:** rutas, cabeceras y comprobaciones de DS4 (API y seguridad), con `//go:embed static`.

**Verification:**
- `go test ./tooling/viewer -run 'Server|Assets'` pasa. Cubre:
  - DELETE y POST `pin` sin token, con token ajeno, con `Origin` ajeno o con `Host: evil.example` → 403.
  - DELETE válido → la copia desaparece, sigue ausente tras un sondeo, y el archivo de origen en el repositorio de prueba conserva su hash (AC14).
  - `GET /api/index`, `/api/sessions/<key>` y `/api/health` sin token o con token ajeno → 403; `/files/<key>/<cap>/<path>` sin capacidad, con una errónea o con la de otra sesión → 403; con la correcta → 200 (AC14, S2-A).
  - `/files/<key>/<cap>/../x` → 400 o 404, y `/files` sirve `image/png` y `video/mp4` por extensión.
  - CSP, `nosniff` y `text/javascript` para `.mjs`.
  - `assets_test` compara byte a byte las copias con `content/skills/flow-walkthrough/assets/` (AC15).
- `go test -race ./tooling/viewer` pasa.

### T6 — Subcomando `hive view`

- [x] verified again at the accepted-candidate gate (AC9, AC10, AC11, AC13, AC17 `met`)

**Closes:** AC9, AC10, AC11, AC13, AC17.

**Depends on:** T3, T4.

**Locations:**
- `tooling/cli/main.go:50-95` (caso `view` y `printHelp`).
- `tooling/cli/view_command.go` y `view_command_test.go` (nuevos).
- `tooling/cli/tui_hosts_test.go:62-82` (`TestMain`: rama `HIVE_VIEW_TEST_SERVE`).
- `tooling/cli/main_test.go:90` (aserción de ayuda).
- Reutiliza `parseHiveSection` y `newGitRunner` de `tooling/cli/doctor_project.go`.

**Execution:** delegada a `hive-build-backend`.

**Test approach:** tdd.

**Changes:** DS4 (ciclo de vida, lanzador y abridor inyectables, `register`, prueba del lanzador).

**Verification:**
- `go test ./tooling/cli -run 'View|HelpAndRequiredScope'` pasa. Cubre:
  - `register` dos veces → el `mtime` de `registry.json` no cambia; un DIR dentro de `<raíz no Git>/_support/workspace/` registra `<raíz>`; un DIR fuera de Git y de cualquier `_support/workspace/` → error (AC9, W1-A).
  - `start` → `server.json` con PID, puerto y versión, y `status` running con URL y sin token; segundo `start` → "already running" y un solo PID; un segundo `serve` directo sobre el mismo `--state-dir` sale con error; `stop` → `status` stopped y sin `server.json`; un `server.json` huérfano con el lock libre no bloquea un `start` nuevo (AC10).
  - Con el servidor iniciado sobre un repositorio de prueba registrado (intervalo de 1 s inyectado), `/api/index` lista `valid/final` antes de dos intervalos; una fixture inválida aparece como error y el servidor sigue respondiendo (AC11, de punta a punta).
  - `start --ttl 1h` deja ese TTL en `health` y en la purga (AC13).
  - `open` sin servidor → mensaje sin lanzar nada; con servidor → el abridor inyectado recibe `<url>/#token=…` (AC17).
- Cada prueba de `start` mata el grupo de procesos en `t.Cleanup`.
- `go test -race ./tooling/cli -run View` pasa.

### T5 — Vista del servidor

- [x] verified again at the accepted-candidate gate (AC19 and the AC14 delete journey `met`)

**Closes:** AC19.

**Depends on:** T6 (arranque y registro reales), T4.

**Locations:** `tooling/viewer/static/index.html`, `viewer.mjs` y `viewer.css`.

**Execution:** delegada a `hive-build-frontend`. La composición está fijada en DS5, y la primera vista rendida va al usuario antes de la puerta de UI.

**Test approach:** check: prueba renderizada con `agent-browser`.

**Changes:** DS5. `viewer.mjs` solo usa `textContent` y `setAttribute`, y `index.html` no tiene código en línea.

**Verification:** con `go run ./tooling/cli view start --state-dir <tmp>` y `view register` en dos repositorios de prueba con fixtures, a 1440×900, 1280×720, 390×844 y 320×640:
- La barra lista los dos proyectos con sus tickets y sesiones, y seleccionar una monta el recorrido con el carrusel de T1.
- ↑, ↓ y Enter recorren y abren, `aria-current` marca la sesión abierta y el foco es visible.
- "Conservar" persiste tras `view stop` y `view start`.
- "Borrar" abre la confirmación; al aceptar, la sesión desaparece, el foco va a la siguiente y los archivos del repositorio de prueba conservan su hash (AC14 de punta a punta).
- Con un `--state-dir` vacío se ve la guía de `hive view register`.
- Con un repositorio registrado sin `walkthrough.json`, el estado sin recorridos.
- Al borrar en el almacén la sesión abierta, en el siguiente refresco (10 s) aparece el aviso y se abre la más reciente.
- Con un manifiesto inválido, el contador de errores del pie.
- Con un repositorio movido, la etiqueta `absent`.
- Abriendo `/` sin `#token`: solo la tarjeta con `hive view open`; la barra lateral queda con una nota y sin datos, y no hay ninguna petición a `/api/*` ni `/files` (S2-A).
- Tras `view stop` con la página abierta, el aviso de servidor caído.
- A 320 px, sin desplazamiento horizontal (AC19).

### T7 — Registro desde la skill y documentación del manager

- [x] verified again at the accepted-candidate gate (AC18 `met`)

**Closes:** AC18.

**Depends on:** T2, T6.

**Locations:**
- `content/skills/flow-walkthrough/SKILL.md` (sección Deliver).
- `README.md` (sección "Update, verify, recover, and uninstall", líneas ~232-290).
- `_support/docs/architecture/deployment-manager.md` (`## Commands`, directorio de estado en las líneas 348-361 y lista de la TUI en la línea 123).
- `_support/docs/architecture/repository-and-distribution.md:12-14`.
- `CHANGELOG.md` (`[Unreleased]`).
- `site/src/content/docs/flows/build.md`.
- `llms.txt` solo si se añade un enlace propio.

**Execution:** hilo principal. Mismo texto de reglas que T2, con el presupuesto de palabras compartido.

**Test approach:** check.

**Changes:** DS3 (registro en el servidor) y DS4 (comandos), sin copiar reglas al sitio.

**Verification:**
- `rg -n "hive view register" content/skills/flow-walkthrough/SKILL.md README.md _support/docs/architecture/deployment-manager.md` → presente en los tres.
- Lectura de la sección Deliver: condición `command -v hive`, fallo no bloqueante, mensaje en el idioma de la sesión con el comando `hive view register` y la ruta del binario literales, la prohibición de buscar el binario y "never run `hive view start`" (AC18). La skill sigue en inglés.
- `wc -w content/skills/flow-walkthrough/SKILL.md` → 1.797 o menos.
- `go test ./tooling/management -run RepositoryCatalogue` y `pnpm -C site build` pasan.

### T8 — Registro de bloques

- [x] verified again at the accepted-candidate gate (AC21 `met`)

**Closes:** AC21.

**Depends on:** T1, T5, y los hallazgos de la puerta de UI en curso, que entran en la misma ronda.

**Locations:** `content/skills/flow-walkthrough/assets/walkthrough.mjs` y su copia `tooling/viewer/static/walkthrough.mjs`, `tooling/viewer/static/viewer.mjs`.

**Execution:** delegada a `hive-build-frontend`, con los hallazgos de la puerta de UI.

**Test approach:** characterization.

**Changes:** DS2 (Bloques).

**Verification:**
- Antes del refactor, guardar el `outerHTML` de `.walkthrough` para cada fixture válida renderizada. Después, el mismo `outerHTML` para cada una, salvo los cambios que pidan los hallazgos de la puerta, nombrados uno a uno (AC21).
- Lectura de `walkthrough.mjs`: un registro ordenado, `renderWalkthrough` sin secciones escritas a mano, y el comentario de cabecera que explica cómo añadir un bloque (AC21).
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` y `go test ./tooling/viewer -run 'Server|Assets'` pasan; la copia sigue byte a byte.

## Verificación compartida y revisión humana

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Suite Go | Candidato final, local | `go vet ./...` y `go test ./...` sin fallos | Parte de la implementación |
| Carreras | Candidato final, local | `go test -race ./tooling/viewer ./tooling/cli` sin fallos (servidor, purga y locks) | Parte de la implementación |
| Skills | Candidato final | `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Parte de la implementación |
| Sitio | Candidato final | `pnpm -C site build` | Parte de la implementación |
| Primera vista | Al existir la página de T1 y la vista de T5 | Capturas a 1440 px en la página del recorrido y pregunta de composición, según [flow-walkthrough](../../../../content/skills/flow-walkthrough/SKILL.md) | Parte de la implementación |
| Puerta de UI | Candidato construido, antes de validar | `full`: `hive-review-ux` y `hive-verify-change`, según [ui-gate.md](../../../../content/skills/flow-walkthrough/references/ui-gate.md) y [ui-review-criteria.md](../../../../content/skills/flow-walkthrough/references/ui-review-criteria.md). Base `8ce403c5`; 1440×900, 1280×720, 390×844 y 320×640; tema oscuro; impresión de la página de la fase 1; `valid/viewports`, `valid/extreme` y un almacén con 30 sesiones; `--state-dir` temporal; sesión `agent-browser` propia por hijo | U1-A (2026-10-18) |
| Seguridad | Candidato final | `hive-review-security` sobre `tooling/viewer` y `view_command.go` (entrada no confiable, servidor local, borrado, ejecución de procesos) | Regla de fronteras de seguridad |
| Revisión de código | Candidato final, antes del push | `/code-review` de Claude Code sobre el diff contra `8ce403c5`; hallazgos atendidos o refutados con evidencia | RV1-A (2026-10-18) |

**Recorrido humano** (con `UI verification` resuelta)
- Fase 1: la página generada desde `valid/final` y desde un recorrido real de este repositorio.
- Fase 2: `go build -o <scratch>/hive-75 ./tooling/cli`, luego `<scratch>/hive-75 view register --state-dir <scratch>/state` en este repositorio, `view start` y `view open`. Navegar por la barra, conservar, borrar y comprobar que el repositorio queda intacto. Al terminar, `view stop`.
- El directorio de estado real solo se usa si el usuario lo pide.

## Estado de revisión y progreso

**Revisión del plan**
- Revisión `9c045496ffb9` por tres `hive-review-plan` (backend, UI, reglas). 13 bloqueantes y 15 no bloqueantes en total.
- Costo: backend 133 k tokens y 156 s; UI 133 k y 161 s; reglas 133 k y 145 s.

**Correcciones aceptadas en esta revisión**
- **Lanzador en pruebas:** inyectable, con rama `HIVE_VIEW_TEST_SERVE` y muerte del grupo de procesos.
- **Almacén:** una copia por fuente y sesión, `content_hash` con las imágenes, y lápidas para que no vuelva lo borrado o purgado.
- **Purga:** por `ingested_at`, no por `created_at`.
- **`stop` y `status`:** por `health` autenticado.
- **Concurrencia:** mutex dentro del proceso y `flock` bloqueante entre procesos.
- **Remoto Git:** normalización propia, sin tocar `parseGitHubOrigin`.
- **Módulo:** `.mjs` en lugar de `.js`, que el manager no despliega.
- **AC:** AC15 dividido en AC15 (copias exactas) y AC19 (vista). AC11 y AC13 verificados en T3 y T6.
- **Contrato del manifiesto:** se mueve a `references/walkthrough-manifest.md`.
- **Lugares que faltaban:** `index.mdx:87`, `flow-report/SKILL.md:10` y `SKILL.md:28,42`.
- **Prueba de enlaces:** `RepositoryCatalogue` en lugar de `tests/content`.
- **CHANGELOG:** de la fase 1 en T2.
- **Sitio:** descripción con enlace, sin copiar condiciones.
- **Verificación de T7:** coherente con DS3.
- **AC3, AC4 y AC5:** concretados con valores esperados y prueba en ejecución.
- **Contrato del módulo:** `renderWalkthrough(root, manifest, {fileBase})`, CSS acotado, `id` por índice, sin `hash`.
- **Dependencias:** T5 depende de T6.
- **DS5:** completado (teclado, confirmación, foco tras borrar, estados, refresco).
- **Puerta de UI:** con 1280×720, 320, datos extremos e impresión.

**Corrección por regla de evidencia.** La cláusula "una sola grabación sobre el candidato aceptado" se retiró: venía de la opción V1-B, que el usuario no eligió. El momento de grabación y la marca `stale` siguen como hoy.

**Re-revisión** (revisión `84846c4afb8a`, mismos revisores)
- Costo: backend 146 k tokens y 52 s; UI 150 k y 64 s; reglas 149 k y 48 s.
- Todos los hallazgos de la primera ronda quedaron resueltos.
- Hallazgos nuevos, aplicados como los propusieron los revisores, sin otra ronda:
  - **Un solo servidor por almacén:** `viewer/server.lock` con `flock` exclusivo durante la vida de `serve`, y `health` sin mutex.
  - **Pin:** se arrastra al reemplazar la copia.
  - **`health`:** devuelve el TTL.
  - **AC10:** la prueba lee PID, puerto y versión de `server.json`.
  - **Ejemplo de la referencia:** lo valida la prueba del generador.
  - **Respaldo sin `hive` en el `PATH`:** frase en el idioma de la sesión, con comando y ruta literales.
  - **AC19:** prueba del estado sin recorridos y del refresco.
  - **Rutas:** segmentos limitados a `^[A-Za-z0-9._-]+$` y codificados en la URL.
  - **Página offline:** la carcasa lleva su estilo de `html` y `body`.
  - **Menú de sesión:** se alcanza con Tab.
- **Va al usuario como decisión, porque la re-revisión la hizo necesaria:** el idioma de las etiquetas fijas de la plantilla. El plan propone `lang` (`es` o `en`) con un diccionario mínimo en el módulo (DS2).
- Revisión actual: `36b44f7cc000`.

**Decisiones de entrega (2026-10-18):** L1-A, E1-A, RV1-A y U1-A (ver `proposal.md`, Entrega).

**Build (2026-10-18)**
- T1 delegado a `hive-build-frontend`.
- T2 editado en el hilo principal mientras T1 corre: sus archivos no se solapan con los de T1.
  - Lugares revisados para el PR: `content/guidance/global.md:78`, `content/agents/quality/hive-verify-change.md:18`, `content/agents/review/hive-review-ux.md:14`, `content/skills/flow-walkthrough/references/ui-gate.md:25-26`, `content/skills/flow-build/SKILL.md:82`, `content/skills/flow-build/references/verification.md:63`, `_support/docs/architecture/`, `tests/pilot/` y `tests/skills/`. Ninguno contradice V1-A ni nombra el esqueleto, y no hizo falta cambiarlos.
  - `content/guidance/global.md` sin cambios: 0 bytes de diferencia.
  - `SKILL.md`: 1.724 palabras (base 1.797). Se quitaron dos frases que repetían reglas cuya casa es `artifact-layout.md` y la referencia del manifiesto.
  - El esqueleto está borrado en el índice y recreado en el árbol de trabajo hasta que T1 termine, porque T1 lo usa como referencia de estilo. Se vuelve a borrar antes de verificar T2.

- T1 verificado por `hive-verify-task` (AC1–AC5 `met`, 142 k tokens, 278 s). Modelo del implementador y del verificador: configurado, no observado.
- **Revisión por la ronda de corrección 1 (primera vista, 2026-10-18).** El usuario pidió galería horizontal, selector Escritorio/Móvil dentro del flujo y captura móvil ajustada en el carrusel. Cambia DS1 (`captures[].viewport`), DS2 (galería, selector, carrusel), añade AC20 y reabre T1. La re-revisión del plan que esto hace debida corre con la puerta del candidato aceptado. Se incluyen dos observaciones del verificador: la etiqueta de sesión mostraba el host, y las miniaturas no reservaban tamaño.

- Ronda de corrección 1 aplicada (36 pruebas de skills OK; 220 k tokens, 371 s). El usuario aceptó la composición (C2-A).
- T2: esqueleto borrado; comprobaciones de T2 pasan (búsquedas, 1.724 palabras, `RepositoryCatalogue`, suite de skills, `pnpm build`).

- T2 verificado por `hive-verify-task` (AC6–AC8 `met`, 138 k tokens, 128 s). Sus tres observaciones menores se corrigieron antes del commit: la entrada de gh-67 en CHANGELOG ya no promete un video por flujo ni el esqueleto de `flow-report`, el comando del generador queda solo en la skill, y la referencia indica `caption` obligatorio, `alt` opcional y los topes.

- T1 verificado de nuevo tras la ronda 1 (AC1–AC5 y AC20 `met`, 177 k tokens, 684 s). Defecto menor corregido en el hilo principal: un valor no textual en un campo enumerado lanzaba `TypeError` sin nombrar el campo (`is_one_of`, dos casos nuevos en `test_wrong_types_are_rejected`; el RED no se observó). T1 y T2 siguen sin commit hasta la puerta de UI.

- T3 verificado (AC11 y AC13 en su parte, AC12 y AC16 `met`; 158 k tokens, 199 s). Defecto del plan corregido: el patrón `-run 'Deps|…'` no seleccionaba `TestNoExternalDependencies`; ahora es `NoExternal`, y esa prueba pasa. Dos huecos de prueba quedan para cerrar en T4 y T6: reemplazo por un cambio solo del manifiesto, purga al iniciar distinguible de la periódica, y agrupación comprobada a través de `/api/index`.

- T4 verificado (AC14 y AC15 `met`; 116 k tokens, 123 s). Cerró los dos huecos de T3 con pruebas que detectan sus mutaciones. Nota del verificador: el borrado desde la barra no lo cubre ninguna tarea de punta a punta; se añade a la verificación de T5.

- T6 verificado (AC9, AC10, AC11, AC13, AC17 `met`; 111 k tokens, 125 s). Observaciones menores: la prueba de AC11 espera hasta 6 s con intervalo de 1 s (la ingesta es inmediata), y `Startup` se llama dos veces (idempotente).

- T7 verificado (AC18 `met`; 113 k tokens, 121 s). Aplicada su observación 1: la skill registra con `--dir <work folder>` para no registrar otro checkout desde un worktree (1.790 palabras). La sección del README quedó bajo "Work in your project", no bajo "Update, verify…"; el enlace del sitio apunta a su ancla.

- T5 implementado (207 k tokens, 874 s); todas las viñetas de su verificación pasaron en el navegador del implementador. Primera vista del visor mostrada al usuario: página del recorrido y visor en vivo (`--state-dir` en el scratchpad, este repositorio registrado). Defecto visto en las capturas, para la puerta de UI: en `walkthrough.css`, una etiqueta de entorno muy larga ocupa casi todo el ancho y aplasta el valor.

- **K1-A (2026-10-19):** el usuario pidió preparar las vistas como bloques. Se añaden AC21 y T8 (refactor que conserva el comportamiento). Por cambiar un criterio, la re-revisión del plan corre con la puerta del candidato aceptado.
- La re-revisión del plan por la ronda 1 (160 k tokens, 49 s) no dejó bloqueantes. Sus cinco ajustes se aplicaron tal como los propuso: AC3 por vista activa, medida de la tira con `valid/extreme`, texto de base de AC20, `min(430px, 100%)` y `valid/viewports` en la puerta.
- El usuario aceptó la composición del visor (C3-A).
- Suite completa sobre el candidato previo a T8: `go vet ./...`, `go test ./...` (20 paquetes), `-race` en `tooling/viewer` y `tooling/cli`, 36 pruebas de skills y `pnpm build`, todo sin fallos.
- T5 verificado (AC19 y el recorrido de borrado de AC14 `met`; 178 k tokens, 456 s). Dos observaciones van a la ronda de T8: el aviso de servidor caído promete reconectar aunque `view start` cambie de puerto, y DS5 dice que el foco va al encabezado de un ticket que ya no existe. Se lanzan la verificación de T5, la puerta de UI `full`, la re-revisión del plan por la revisión de la ronda 1 y la suite completa.

- Puerta de UI (`full`) sobre el candidato previo a T8:
  - `hive-verify-change`: J1–J3 pasan, sin defectos (155 k tokens, 612 s).
  - `hive-review-ux`: no pasa (201 k tokens, 548 s). H1 High: el visor a 320 px desborda con una sesión conservada. M1: la sesión "más reciente" se elige por hora de ingesta. M2: el aviso de borrado no es fiable. M3: miniaturas pequeñas, sin pista de desplazamiento y sin recuento. Notas N1–N5.
- Ronda de corrección 2: T8 (bloques), H1, M1–M3, N1–N5, la etiqueta de entorno larga, el texto de reconexión y el foco al cerrar el carrusel. La hace el implementador de T5.

- Ronda 2 aplicada (286 k tokens, 609 s). El `outerHTML` de las cinco fixtures válidas es idéntico antes y después del refactor a bloques. Cada hallazgo se midió antes y después; por ejemplo, H1 pasa de `scrollWidth` 342 a 320. Las 36 pruebas de skills y `Server|Assets` pasan. Los archivos de T1 y T5 cambiaron por los arreglos de la puerta y se vuelven a verificar en la puerta del candidato aceptado.

- T8 verificado (AC21 `met`; 125 k tokens, 188 s). El verificador reconstruyó el módulo anterior a la ronda: el refactor aislado no cambia el DOM, y las dos diferencias del DOM vienen de la puerta (el recuento en "Capturas · N" y el `aria-label` del video). Ajustes menores aplicados: el comentario de `ctx` nombra `indexEntries`, y DS2 dice `applies(ctx)` con `index(ctx)` opcional.

- Re-revisión de UX (241 k tokens, 262 s): pasa. H1, M1, M2 y N1–N5 resueltos; M3 parcial (Medium: en escritorio las miniaturas miden 333 px). Queda como opción para la validación del usuario. Notas menores: el aviso de servidor caído no se puede descartar a 320 px, y la impresión con carga diferida necesita que la página termine de cargar.
- `hive-review-security` (306 k tokens, 1.270 s): apto con cambios.
  - **Incidente contenido:** un comando mal escrito del revisor envió a hosts públicos el token de su propio servidor de prueba. El servidor ya estaba detenido y el token era de un solo arranque y solo de loopback, así que no hay nada que rotar. Se informó al usuario.
  - **Se corrigen ahora:** H1 (carrera con enlace al copiar), H3 (temporal del generador que sigue enlaces), H4 (errores sin tope), I1 (paridad Python/Go) e I2 (texto de AC16).
  - **Decisión del usuario:** S2-A, las lecturas exigen el token. Cambia AC14 y AC19.
  - **Propuestos como issues de seguimiento al cierre:** H5 (hash en cada sondeo), H6 (token en el argumento de `open`) y H7 (`walkthrough.json` versionado ingerido).
- Ronda de seguridad: la hace el implementador de T4. Toca `viewer.mjs`, así que vuelve a correr la puerta de UI sobre ese diff.

- Ronda de seguridad aplicada (188 k tokens, 364 s). H1, S2-A, H3, I1 y H4 tienen pruebas que fallaban antes del arreglo, salvo la paridad en Go, que ya era estricta, y H4 en Go, cuyo RED fue solo de compilación. Fuera de su lista quedó un ajuste necesario: `status` envía el token al contar. Residual declarado de H1: entre listar y abrir una carpeta, un enlace interno al propio workspace podría seguirse, sin salir de él.
- Puerta de UI sobre el diff de seguridad: re-revisión de UX y re-verificación en vivo en curso. Suite completa en curso.

- Puerta sobre el diff de seguridad: `hive-review-ux` encontró un Blocker introducido. `viewer.mjs` no se había actualizado, aunque el informe de la ronda decía lo contrario, y su prueba de cadenas pasaba porque esas cadenas ya existían en el archivo. Las lecturas salían sin token y el visor no mostraba nada. Se detuvo la re-verificación en vivo, porque habría encontrado el mismo bloqueo. La corrección de `viewer.mjs` la hace el implementador de T5.

- `viewer.mjs` corregido (lecturas con token, `fileBase` del índice, tarjeta sin token, 403 frente a error de red). Puerta repetida sobre el diff: `hive-review-ux` pasa (269 k tokens, 138 s) y `hive-verify-change` pasa J1–J3 (183 k tokens, 177 s). Suite completa en verde después de la ronda de seguridad. Nota de redacción en español corregida.
- Recorrido final entregado para validación humana: página `walkthrough.html` (manifiesto `final`) y visor en vivo reconstruido. Sin grabación (V1-A).

- **Ronda de corrección 3 (validación, 2026-10-19):** el usuario pidió que el bloque de comprobaciones sea plegable y esté cerrado por defecto. Cambia DS2 y no cambia ningún criterio. La aplica el implementador de T5. T1 y T8 tocan `walkthrough.mjs`, así que se reverifican en la puerta del candidato aceptado.

- Ronda 3 aplicada (326 k tokens, 95 s): `<details>` cerrado con los recuentos en la cabecera y desplegado al imprimir; 41 pruebas de skills y `Server|Assets` pasan. Página regenerada y visor reconstruido en `http://127.0.0.1:53025`.

- **Ronda 4 (R1-B, 2026-10-19):** el usuario pidió que nada se abra solo. La skill reemplaza "Open the page" por "Give the opening command": una línea con `hive view open` si el registro funcionó, o la ruta de la página si no. También cambian `flow-build/SKILL.md:76`, `site/.../flows/build.md:75` y CHANGELOG. Reemplaza la preferencia del 2026-10-18 (abrir solo el HTML). `SKILL.md` queda en 1.776 palabras tras el `--dir` del respaldo. El cambio es del hilo principal; la re-revisión del plan por este cambio de regla corre con la puerta del candidato aceptado.

- **Validación (V2-A, 2026-10-19):** el usuario validó tras las rondas 3 y 4. Las pruebas afectadas pasan.
- **W1-A (2026-10-19):** al preguntar el usuario por cam y carvento, se comprobó que son workspaces cuya raíz no es Git, con sus recorridos en `<workspace>/_support/workspace/`, y que `register` fallaba ahí. Es un defecto de esta obra. Cambian AC9 y DS4. Lo corrige el implementador de T3, y entra en la puerta del candidato aceptado.

- Re-revisión del plan, dominio backend (160 k tokens, 47 s). B1 y B2 son bloqueantes: a T4 le faltaba verificar las lecturas con token y a T6 la regla de W1-A. Se aplicaron sus propuestas a la verificación de T3, T4 y T6. N1 (el remoto se lee solo cuando la raíz es el toplevel de Git) va al arreglo W1-A. N3: el mensaje de respaldo cita `hive view register --dir <work folder>`. El arreglo de W1-A toca `view_command.go` (T6) y el registro (T3), y su verificación entra en la puerta para ambas tareas.

- Re-revisión del plan, dominio UI (175 k tokens, 47 s): un bloqueante aplicado (DS5 y T5 con la tarjeta sin token en lugar del modo de solo lectura). También: `fileBase` del índice en DS5, y la cláusula "DOM idéntico" de AC21 pasa a ser evidencia de T8.
- Re-revisión del plan, dominio de reglas (166 k tokens, 57 s): R1-B se añade a las decisiones y como AC22, cerrado por T2. DS3 queda alineado. La regla se acota a "para mostrárselos al usuario", para que siga permitida la navegación de capturas. La línea siempre da la ruta de la página y añade `hive view open` si el registro funcionó. El sitio solo enlaza. Con esto termina la única ronda de re-revisión.

- Arreglo W1-A aplicado (254 k tokens, 213 s): `ResolveRoot` toma la raíz de `_support/workspace` aunque no sea Git, y el remoto solo se lee cuando la raíz es el toplevel de Git. Pruebas en registro, índice y CLI, y humo con binario.
- Puerta del candidato aceptado lanzada: UX limitada a la sección plegable, verificación en vivo de lo que cambió, y `hive-verify-task` de T1–T8.

- Puerta del candidato aceptado:
  - T2, T3, T4, T6, T7 y T8 vuelven a verificarse; `hive-verify-change` pasa.
  - `hive-review-ux` encontró un High introducido: con la tabla de comprobaciones abierta, a 390 y 320 px se partían palabras reales. Se corrigió en `walkthrough.css` y su copia, cambiando `overflow-wrap: break-word` en las celdas y dando a la primera columna un mínimo de 12ch. Lo re-revisa el mismo revisor.
  - Correcciones menores de los verificadores: la ayuda de `register` y el README mencionan los workspaces y `--dir`, y la prueba de H4 pone los caracteres de control al principio.

- Re-revisión de UX del High: pasa (300 k tokens, 80 s). Ninguna palabra partida en página ni visor a 1440, 390 y 320, y sin desplazamiento horizontal de página. Nota aplicada: el contenedor de la tabla es una región enfocable con etiqueta (`tabindex`, `role=region`, `aria-label`). Pruebas de skills y `Server|Assets` pasan.

- Puerta del candidato aceptado completa: T1–T8 verificadas otra vez, y UX y verificación en vivo pasan.

- Suite final en verde. Commits: `122353b8` (T1, T8; también lleva el borrado del esqueleto, que ya estaba en el índice), `5fc7cfb4` (T2, T7), `8447adfa` (T3), `731c8168` (T4, T5, T8), `20e756fc` (T6) y `71a9355a` (T7). Varias tareas comparten archivos, así que cada mensaje nombra todas las que lleva.
- `/code-review` (RV1-A; 197 k tokens, 257 s): 10 hallazgos.
  - **Se corrigen nueve:**
    - En el núcleo del visor y el generador: hash en cada sondeo (H5), `git` en cada sondeo, `Z` en Python 3.9 y 3.10, `1e999`, y `Store.Register` sin uso.
    - En el CLI: carrera del sondeo del lock, plazo de `stop` y `Startup` doble. También se reutiliza `readProjectFile`.
  - **Queda como issue:** el token en el argumento de `open` (H6).
  - Los arreglos irán en commits adicionales `(T<n>)`.

- Arreglos de la revisión de código aplicados (CLI: 167 k tokens; núcleo: 303 k tokens). Suite completa en verde, prueba de H1 incluida.
- Re-chequeo de seguridad del nuevo sondeo (378 k tokens, 600 s): apto. H1 sigue cerrado (682.851 intercambios sin escape), el salto por huella no sirve contenido sin validar y no hay carreras en las pruebas de estrés con `-race`.
- Regresión Baja introducida por el arreglo: la primera pasada retiene todas las fuentes cambiadas (472 MiB con 300 manifiestos de 1 MiB). Se corrige con copia por lotes, `ctime` y `dev` en la huella, y poda de huellas.

- Ajuste aplicado: copia por lotes (memoria de 479 a 78 MiB en la prueba de 300 manifiestos), `ctime` y `dev` en la huella, y poda de huellas. Commits de arreglos `316ec13a` (T3), `5507e944` (T1) y `0a472f03` (T6).
- Entregado: [PR #76](https://github.com/JhonHawk/tricell-hive/pull/76), fusionado en `development` como `9dca1e6d` el 2026-10-19 (sin CI; base al día comprobada). #75 se cerró solo.

**Siguiente paso:** ninguno; registro archivado.
