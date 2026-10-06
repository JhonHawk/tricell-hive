# Tareas

Commit base: `59aade0ae4cc16b1639d75edd3e12751ad8ad472` (rama `feat/gh-17-docs-site`, 2026-10-06).

Orden: T1 → T5 → T2 → T3 → T4 → verificación compartida → tu validación → entrega a `development` → T6 con la próxima versión.

Todo es secuencial porque los subagentes hacen commits en la misma rama y el mismo árbol de trabajo. T1 y T5 los hace el hilo principal antes de lanzar a los subagentes.

### T1 — Plantilla `starlight-docs-site` al día

- [x] Los dos perfiles de la plantilla, su `SKILL.md` y su prueba usan las versiones de [design.md](design.md#contexto-verificado), y la plantilla instala limpia con pnpm 12.8.1.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/skills/starlight-docs-site/assets/{user-manual,spec-site}/package.json` y, si hace falta, `pnpm-workspace.yaml` nuevo en cada perfil; `content/skills/starlight-docs-site/SKILL.md:12`; `tests/skills/starlight_assets_test.py:20-23`.

**Execution:** hilo principal. Son pocos archivos con valores ya decididos, y su encargo sería más largo que el cambio.

**Test approach:** tdd.

**Changes:**
1. Cambiar primero las versiones esperadas de la prueba (`astro` 7.3.5, `@astrojs/starlight` 0.42.4, `typescript` 6.0.3, `packageManager` `pnpm@12.8.1`) y agregar `@astrojs/check` 0.9.10, `@biomejs/biome` 2.5.14 y `sharp` 0.35.5. Comprobar que falla.
2. Actualizar los dos `package.json` y la frase de `SKILL.md`, que nombra Astro, Starlight, TypeScript y pnpm.
3. Instalar una copia limpia del perfil `user-manual` en el scratchpad. Si pnpm pide aprobar scripts de instalación:
   - confirmar primero en la documentación de pnpm 12 (Context7) el nombre de la clave (`allowBuilds` u otra);
   - cada perfil lleva un `pnpm-workspace.yaml` con esa clave para esos paquetes;
   - la prueba exige que exista y que los nombre;
   - `SKILL.md` agrega una línea: el archivo solo aprueba scripts de instalación, y en un proyecto que ya tiene un espacio de trabajo pnpm sus aprobaciones se fusionan con las existentes en vez de copiar el archivo.

**Verification:**
- `python3 -m unittest discover -s tests/skills -p 'starlight_assets_test.py'` falla antes del paso 2 y pasa después.
- En una copia del perfil `user-manual` en el scratchpad, sin aprobaciones a mano:
  - `pnpm --version` imprime `12.8.1`;
  - `pnpm install`, `pnpm check` y `pnpm build` terminan bien y generan `dist/index.html`.
  - Se registra la salida. La copia no se versiona.

### T5 — Regla del README, menciones de `site/` y CHANGELOG

- [x] El `AGENTS.md` raíz tiene la regla de [design.md](design.md#regla-del-readme-d2-a), `CONTRIBUTING.md` la exige, el README, `llms.txt` y el mapa del `AGENTS.md` nombran `site/`, y `CHANGELOG.md` anota el cambio.

**Closes:** AC7.

**Depends on:** ninguna.

**Locations:** `AGENTS.md` (sección nueva después de `## Guidance structure` y una línea en `## Repository map`); `CONTRIBUTING.md`, sección `Branches and commits`; `README.md`, sección `For AI agents` (líneas 313-329); `llms.txt`, párrafo de cabecera; `CHANGELOG.md`, sección `[Unreleased]`.

**Execution:** hilo principal. Es texto corto con decisiones ya tomadas.

**Test approach:** check.
- `rg -n 'README and llms.txt' AGENTS.md CONTRIBUTING.md`.
- `rg -n 'site/' README.md llms.txt AGENTS.md`.

**Changes:**
1. Escribir la regla en inglés.
2. Escribir la obligación en `CONTRIBUTING.md`.
3. Agregar `site/` como material que no se instala en las tres listas de carpetas.
4. En el CHANGELOG, bajo `[Unreleased]`:
   - `### Added`: el sitio de documentación en `site/`, que se publica con la próxima versión.
   - `### Changed`: las versiones de la plantilla `starlight-docs-site`.

**Verification:**
- Los dos comandos de arriba muestran la sección, la obligación y las tres menciones de `site/`.
- Una lectura de `CHANGELOG.md` confirma las dos entradas en sus secciones.

### T2 — Esqueleto de `site/` y protecciones

- [x] `site/` existe desde la plantilla actualizada, con la configuración de [design.md](design.md#estructura-de-site), la protección para Go y las entradas de `.gitignore`.

**Closes:** AC4, AC5.

**Depends on:** T1.

**Locations:** `site/` (nuevo), salvo `site/src/content/docs/flows/`, `site/src/styles/` y `site/src/assets/`; `.gitignore`.

**Execution:** delegada a `hive-build-frontend`. Es dueña de todo `astro.config.mjs` salvo las claves `logo` y `customCss`, que agrega T4.
- El encargo incluye este plan, [design.md](design.md), [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md), la skill [starlight-docs-site](../../../../content/skills/starlight-docs-site/SKILL.md) y `CONTRIBUTING.md`.
- Puede hacer commits locales en `feat/gh-17-docs-site`; no hace push.

**Test approach:** check.
- `go vet ./...` con el archivo de prueba de AC4.
- `git check-ignore` para AC5.

**Changes:**
1. Copiar `assets/user-manual` a `site/`. Cambiar `name` a `hive-docs`, `title` a `Hive` y agregar `site: "https://hive.tricell.tech"`.
2. Configurar la barra lateral: `Start` (portada) y `Flows` (research, plan, build, close), apuntando a las páginas que crea T3.
3. Agregar `wrangler` 4.143.0 a las dependencias de desarrollo y `.node-version` con `24`.
4. Crear `site/go.mod` con un comentario que explique por qué existe y `module hive.invalid/site`.
5. Agregar `site/node_modules/`, `site/dist/`, `site/.astro/` y `site/.wrangler/` a `.gitignore`.
6. Escribir `wrangler.jsonc` según el diseño, después de confirmar sus campos en la documentación de Workers (Context7 `/cloudflare/cloudflare-docs`).
7. Instalar con pnpm 12.8.1:
   - aprobar solo los scripts de build necesarios en `site/pnpm-workspace.yaml`, con su motivo;
   - versionar ese archivo y `pnpm-lock.yaml`;
   - correr `pnpm audit --prod` y reportar su salida.

**Verification:**
- AC4:
  1. `mkdir -p site/node_modules/fake && printf 'package fake\nimport _ "does.not/exist"\n' > site/node_modules/fake/bad.go`.
  2. `go vet ./... && go test ./...` desde la raíz termina bien.
  3. Mover `site/go.mod` un momento y comprobar que `go vet ./...` falla, para confirmar que la protección es la causa.
  4. Devolver `site/go.mod` y borrar `site/node_modules/fake`.
- AC5: `git check-ignore site/node_modules site/dist site/.astro site/.wrangler` imprime las cuatro rutas.
- T2 no compila el sitio: la barra lateral nombra páginas que crea T3, y Starlight rechaza un build con páginas inexistentes en la barra lateral.

### T3 — Portada y páginas de flujo

- [x] La portada y las cuatro páginas de flujo existen con el contenido de [design.md](design.md#contenido).

**Closes:** AC2, AC3.

**Depends on:** T2.

**Locations:**
- Escribe solo en `site/src/content/docs/index.mdx` y `site/src/content/docs/flows/{research,plan,build,close}.md`.
- Fuentes: `content/skills/flow-{research,plan,build,close}/SKILL.md` y sus `references/`, `README.md:252-311` y `content/guidance/global.md`.

**Execution:** delegada a `hive-write-spec`, porque es documentación a partir de fuentes verificadas. No toca `astro.config.mjs`. Mismas condiciones de commits que T2.

**Test approach:** check.
- Build limpio sobre una copia exportada para AC2.
- Encabezados y comparación de secuencias para AC3.

**Changes:**
1. Escribir las páginas en inglés con los encabezados exactos y el enlace al `SKILL.md` en `master`.
2. Escribir la portada con la instalación de una línea y las tarjetas de los cuatro flujos.

**Verification:**
- AC2:
  1. `git archive HEAD site | tar -x -C <scratch>`.
  2. En `<scratch>/site`: `pnpm install --frozen-lockfile && pnpm check && pnpm build`, sin aprobar nada a mano.
  3. `ls dist/index.html dist/404.html dist/flows/{research,plan,build,close}/index.html` lista los seis archivos.
  4. 404 local: desde `<scratch>/site`, `pnpm exec wrangler dev`, y una petición a una ruta inexistente responde `404` con la página de Starlight. Si `wrangler dev` exige iniciar sesión, se anota y la comprobación queda para T6.
- AC3, encabezados: `rg -c '^## (When to use|What it produces|What it does not do|Related flows|FAQ)$' site/src/content/docs/flows/*.md` da 5 por archivo, y `rg -l 'blob/master/content/skills/flow-' site/src/content/docs/flows/` lista los cuatro.
- AC3, sin copias: un script de un solo uso en el scratchpad compara cada página con las fuentes.
  - Fuentes: `content/skills/flow-*/**` y `content/guidance/global.md`.
  - Preparación: quita el formato Markdown (encabezados, viñetas, énfasis, código y destinos de enlaces) y pasa todo a minúsculas.
  - Comparación: lista toda secuencia de 8 palabras seguidas presente en una página y en alguna fuente.
  - Resultado esperado: ninguna. Las frases cortas inevitables, como nombres de flujos y rutas, no forman secuencias de 8 palabras.

### T4 — Marca visual

- [x] El sitio muestra el ícono de Hive en cada tema y usa los colores de la marca con contraste medido.

**Closes:** AC6.

**Depends on:** T3.

**Locations:** `site/src/assets/hive-mark.svg` (copia), `site/src/assets/hive-mark-light.svg` (nuevo), `site/src/styles/theme.css`; las claves `logo` y `customCss` de `site/astro.config.mjs`.

**Execution:** delegada a `hive-design-ui`, porque incluye una variante nueva del ícono y la elección de colores con evidencia renderizada. Escribe solo en esos archivos y claves. Mismas condiciones de commits que T2.

**Test approach:** check.
- `contrast-check.py` para cada par.
- Revisión del HTML compilado.

**Changes:**
1. Copiar el ícono oscuro y crear la variante clara según [design.md](design.md#estructura-de-site).
2. Leer las correspondencias de variables en `props.css` de la versión instalada.
3. Definir `accent-low`, `accent` y `accent-high` por tema según [design.md](design.md#marca-visual).
4. Configurar `logo: { dark, light, alt: "" }` y `customCss`.

**Verification:**
- Cada par de [design.md](design.md#marca-visual), en cada tema, con `uv run --with coloraide python content/skills/starlight-docs-site/scripts/contrast-check.py <primer plano> <fondo>`: 4.5:1 o más. La tabla de resultados va en el reporte.
- `rg -o '<img[^>]*alt=""[^>]*>' site/dist/index.html` encuentra las dos variantes del ícono, y el título `Hive` aparece como texto en la cabecera.
- Captura de la cabecera en tema claro y oscuro.

### T6 — Publicación con la próxima versión (orden aparte)

- [ ] El sitio está en línea en `hive.tricell.tech`, `/install.sh` sigue redirigiendo, y el README y `llms.txt` enlazan el sitio.

**Closes:** AC8, AC9.

**Depends on:** merge de T1–T5 a `development` y la orden de publicar la próxima versión.

**Locations:**
- `README.md`: navegación (líneas 8-14), flujos (281-289) y `Documentation` (331-344).
- `llms.txt`: sección `## Docs`.
- Panel de Cloudflare.

**Execution:**
- Promoción, PR de enlaces y, si hiciera falta, la vuelta de pnpm: hilo principal, con la orden de publicar.
- Cloudflare: lo haces tú en el panel, con los pasos de [design.md](design.md#publicación-t6-d4-a-y-d7-a).

**Test approach:** check.
- `curl` contra el dominio publicado y lectura de la configuración de builds para AC8.
- `rg` sobre `README.md` y `llms.txt` para AC9.

**Changes:** en el orden de [design.md](design.md#publicación-t6-d4-a-y-d7-a):
1. Promoción.
2. Cloudflare.
3. Comprobación en `workers.dev`.
4. Ruta.
5. PR de enlaces.

**Verification:**
- AC8:
  - `curl -sI https://hive.tricell.tech/` y `curl -sI https://hive.tricell.tech/flows/plan/` responden `200` con `content-type: text/html`.
  - `curl -s -o /dev/null -w '%{http_code}' https://hive.tricell.tech/no-such-page` imprime `404`, y su cuerpo es la página 404 de Starlight.
  - `curl -sI https://hive.tricell.tech/install.sh` responde `302` con `location: https://raw.githubusercontent.com/JhonHawk/tricell-hive/master/get-hive.sh`, también después de volver a desplegar.
  - Una captura o lectura de la configuración de builds muestra la rama de producción `master` y las vistas previas apagadas.
- AC9: `rg -n 'hive.tricell.tech/(flows/|\)|$)' README.md llms.txt` encuentra los enlaces nuevos.

## Verificación compartida y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autorización |
| --- | --- | --- | --- |
| Suite local | candidato final | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` y la verificación de AC2 sobre la copia exportada: todo pasa | parte de la implementación |
| Primera captura | después de T4, con `pnpm --dir site dev` | te muestro la portada y una página de flujo en tema claro y oscuro, antes de las revisiones, para corregir la composición a tiempo; las capturas van en `_support/workspace/2026-10-06-gh-17-docs-site/images/` | parte de la implementación |
| Revisión de UX | sitio compilado servido con `pnpm --dir site preview`, en paralelo con el gate siguiente y con sesión de navegador propia | `hive-review-ux` con [UI review criteria](../../../../content/skills/flow-build/references/ui-review-criteria.md), con `agent-browser`:<br>• Base de comparación: Starlight por defecto, porque no hay sitio previo.<br>• Portada, la página de flujo más larga y la 404.<br>• Anchos 1440, 1280, 390 y 320 px, en tema claro y oscuro.<br>• Buscador, menú móvil y cabecera en tema claro.<br>Resultado esperado: sin hallazgos Blocker ni High introducidos | parte de la implementación |
| Verificación en vivo | mismo candidato, sesión de navegador aparte | `hive-verify-change`:<br>• Navegar de la portada a cada flujo y abrir el enlace al `SKILL.md`.<br>• Usar el buscador y cambiar de tema.<br>• Abrir una ruta inexistente (404) y usar el menú móvil. | parte de la implementación |
| Tu validación | después de los gates anteriores | te dejo `pnpm --dir site preview` corriendo, con las capturas seleccionadas | parada de la entrega interactiva |
| `/code-review` | sobre el diff, después de tu validación y antes del merge | revisión nativa de Claude Code (D6-B) | elegida |
| Comprobación en vivo | durante T6 | los `curl` y la lectura de configuración de T6 | con la orden de publicar |

## Estado de la revisión y avance

Revisión del plan:
- **Primera ronda** sobre `dc61677c8e68`, con tres subagentes `hive-review-plan` (interfaz; infraestructura y entrega; guía y plantilla). Encontraron cinco hallazgos bloqueantes. Las correcciones aplicadas:
  - **Escritores sobre `astro.config.mjs`**: tareas secuenciales y dueños por clave.
  - **Pares de contraste reales de Starlight**: se miden las variables que Starlight pinta, en los dos temas.
  - **Publicación sin corte de `/install.sh`**: ruta del Worker sobre el registro actual en vez de Custom Domain, con su recuperación.
  - **Vuelta de pnpm**: queda dentro de la orden de publicar.
  - **Build limpio**: se comprueba sobre una copia exportada, con aprobaciones versionadas.
- **No bloqueantes aceptados**:
  - ícono para tema claro y `alt` vacío;
  - anchos y alcance de la revisión de UX;
  - primera captura antes de las revisiones;
  - 404 en local y en producción;
  - `compatibility_date` y nombre del Worker;
  - orden de los enlaces del README;
  - lectura de la configuración de builds;
  - registro versionado en el PR principal;
  - menciones de `site/` en el README, `llms.txt` y el mapa;
  - comparación de secuencias de 8 palabras;
  - pnpm 12 sin aprobaciones a mano en la plantilla;
  - secciones del CHANGELOG;
  - qué versiones nombra `SKILL.md`;
  - obligación explícita en `CONTRIBUTING.md`.
- **Segunda ronda** sobre `c2df69e319bc`, con los mismos tres revisores. Dieron por resueltas todas las correcciones. Los hallazgos nuevos se aplicaron tal como los propusieron los revisores, sin otra ronda:
  - el 404 local pasa de T2 a T3, porque T2 no compila el sitio;
  - T1 confirma la clave de aprobaciones de pnpm 12 y `SKILL.md` explica cómo fusionarla en un monorepo;
  - el PR de enlaces de T6 pone `workers_dev: false`;
  - AC8 nombra la comprobación después de volver a desplegar.
- **Límites que quedan para T6** (están en [design.md](design.md#contexto-verificado)):
  - que Workers Builds acepte pnpm 12.8.1;
  - que una ruta creada en el panel sobreviva a `wrangler deploy` con `workers_dev: true`;
  - cómo se interpretan las rutas que disparan el build.
  - Los cubren las comprobaciones de T6, y la recuperación es quitar la ruta.

Avance (build iniciado el 2026-10-06 con `/flow-build`):
- **T1**: commit `fce8f6fb`.
  - Al instalar aparecieron dos ajustes de versión, explicados en [design.md](design.md#contexto-verificado): pnpm 12.8.1 por pnpm#16594, y Starlight 0.42.4, Biome 2.5.14 y wrangler 4.143.0 por la regla de 7 días de tu configuración de pnpm.
  - Se agregó `pnpm-workspace.yaml` con `esbuild: false` a cada perfil.
  - Evidencia: la prueba falló con `'7.0.7' != '7.3.5'` y después pasó, y los dos perfiles instalan, pasan `astro check` (0 errores) y compilan con pnpm 12.8.1. Salida en el scratchpad de la sesión (`t1-evidence.txt`).
- **T5**: commit `391eab5b`.
- **T2**: commits `42da7e28` y `fa4c57e5`.
  - `compatibility_date` bajó a 2026-10-03, porque el `workerd` de wrangler 4.143.0 no soporta una fecha más nueva.
  - Se deniega el script de `workerd`.
  - La auditoría quedó en [design.md](design.md#dependencias-del-sitio-auditoría-de-t2).
  - Verificado: AC4 y AC5 cumplidos.
- **T3**: commit `c0700a1d`, verificado con AC2 y AC3 cumplidos.
  - La verificación también revisó la exactitud de las páginas y encontró seis afirmaciones que omitían condiciones de las fuentes, sin contradicciones.
  - Se corrigieron en el hilo principal en `546c566f`. Después se repitió la comparación de secuencias de 8 palabras (0 coincidencias) y el build (6 páginas).
- **T4**: commit `86af4f87`, verificado con AC6 cumplido (diez pares de color en 4.86:1 o más).
  - `15fc079d` deja `pnpm lint` limpio: Biome excluye `dist/` y `.astro/`.
  - En tema oscuro el acento de texto es `#ff6464`, porque `#ff2a2a` da unos 4.0:1 sobre la cabecera.
- **Primera captura** (2026-10-06): pediste que la portada siguiera la estructura de una guía de una sola página; quedó en `1098631f` y la aprobaste ("me gusta"). No elegiste las propuestas C1, C3 ni C4.
- **Suite local** sobre `1098631f`: `go vet ./...` limpio, `go test ./...` con todos los paquetes `ok`, y 23 pruebas de Python en `OK`.
- **`hive-verify-change`** sobre `1098631f`: los siete recorridos pasan.
  - Defecto D1: los bloques de ejemplos cortan líneas.
- **`hive-review-ux`** sobre `1098631f`: sin Blocker ni High.
  - Introducidos y corregidos en `f569ada0`:
    - M1, igual a D1: los pedidos de ejemplo se ajustan al ancho; a 390 px ningún bloque de pedidos se desborda y solo el comando de instalación sigue con scroll propio.
    - L1: la pestaña de la portada ya no dice "Hive | Hive".
    - L2: las preguntas frecuentes pasan a ser `###`.
  - Introducidos y aceptados:
    - M2: la portada no tiene menú lateral, por la estructura de una sola página que elegiste.
    - N2: los enlaces a GitHub no llevan indicador de enlace externo.
    - N3: el detalle interior del ícono oscuro es tenue; es decorativo.
    - N4: el acento salmón en tema oscuro, ya justificado arriba.
  - Heredados de Starlight, solo reportados: L3 (fragmento de búsqueda con "Terminal window"), L4 (404 sin enlaces), L5 (foco tras cerrar la búsqueda) y N1 (anclas de encabezados como paradas de tabulación).
- **Tu validación** (2026-10-06): aceptaste el candidato `f569ada0` ("perfecto").
  - Después corrió una revisión limitada al diff `1098631f..f569ada0` con `hive-verify-task` (T3), `hive-review-ux` y `hive-verify-change`.
  - Push de `feat/gh-17-docs-site` y [PR #18](https://github.com/JhonHawk/tricell-hive/pull/18) a `development`; `/code-review` sobre el PR antes del merge.
- **Revisión limitada al diff validado**: `hive-verify-task` dio AC2 y AC3 cumplidos; `hive-verify-change` y `hive-review-ux` pasaron sin Blocker ni High.
  - Observaciones menores no corregidas, para no abrir otra ronda: separación entre pedidos ajustados al ancho, tamaño de los encabezados del FAQ, título de la portada en la búsqueda, comando de instalación bajo el botón de copiar y `og:title`.
- **`/code-review`** sobre el PR #18: diez hallazgos.
  - Aplicados en `8063d6fe`:
    - plataformas y terminal del instalador, y enlace a la sección de actualizar y desinstalar;
    - botón "Install";
    - lo que borra Close, completo;
    - comentario de `theme.css` en inglés;
    - `CONTRIBUTING.md` solo enlaza la regla;
    - `site/` en la tabla de estructura;
    - `site/.dev.vars*` y `site/.env*` ignorados.
  - Refutados:
    - el registro sin versionar entra al PR antes del merge;
    - los flujos listados dos veces en la portada son la estructura que validaste.
  - Comprobado después: comparación de 8 palabras en 0, build de 6 páginas, lint limpio y render a 390 px sin desbordes.
- **Modelos de verificación**: T2 y T3 los implementaron roles en `sonnet` (configurado, no observado), así que su verificación corrió con `hive-verify-task` en `opus` (configurado). T1 y T5 los implementó el hilo principal (Claude Opus 5.5, observado). `hive-verify-task` tiene `opus` configurado, que es el mismo modelo, así que se lanza con `sonnet` mediante la opción de modelo del lanzamiento (configurado, no observado).
