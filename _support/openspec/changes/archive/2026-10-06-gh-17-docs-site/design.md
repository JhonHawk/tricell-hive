# Diseño

## Contexto verificado

Consultado el 2026-10-06 en `development` (`59aade0a`).

- **Paquete de versión.** Usa una lista cerrada: `tooling/package/main.go:205` congela solo `go.mod`, `go.sum`, `install.sh`, `LICENSE`, `THIRD_PARTY_NOTICES.md`, `content` e `integrations`, y `release.json` (`tooling/distribution/manifest.go:121`) rechaza archivos de más. `site/` no entra.
- **`hive update`.** Extrae el commit completo con `git archive` (`tooling/cli/update.go:295`), pero solo lee y despliega `content` e `integrations` (`tooling/management/plan.go:141,159,193`). `site/` sin `node_modules` queda muy por debajo de los límites de `tooling/distribution/extract.go:16-18`.
- **Pruebas.** La suite es `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` (`CONTRIBUTING.md:17-21`). Ninguna prueba recorre todo el repositorio. La revisión de enlaces (`tooling/management/references.go:20-35`) solo mira `content/agents` y los `SKILL.md` y `references/` de las skills.
- **Riesgo de Go.** El subagente lo reprodujo en una copia: Go solo ignora carpetas que empiezan con `.` o `_` o que se llaman `testdata`. Con un `node_modules` plano (`npm`, `yarn`) que trae archivos `.go`, `go vet ./...` falla. Con la disposición de pnpm (`node_modules/.pnpm/…`) pasa.
- **`.gitignore`.** Solo tiene `/dist/` anclado a la raíz; `site/node_modules`, `site/dist` y `site/.astro` no están ignorados.
- **Plantilla.** `content/skills/starlight-docs-site/assets/{user-manual,spec-site}/package.json` fijan Astro 7.0.7, Starlight 0.41.3, TypeScript 6.0.3 y `pnpm@11.21.0`; `tests/skills/starlight_assets_test.py:20-23` los exige y `SKILL.md:12` los nombra.
- **Versiones** (registro de npm y nodejs.org, 2026-10-06): Astro 7.3.5, Starlight 0.42.4 (pide `astro ^7.2.10`), `@astrojs/check` 0.9.10 (pide TypeScript `^5 || ^6`, por eso TypeScript sigue en 6.0.3 aunque exista 7.0.2), Biome 2.5.14, sharp 0.35.5, pnpm 12.8.1, wrangler 4.143.0, Node 24.21.0 LTS. Ajuste del build (2026-10-06): las versiones fijadas son las últimas con más de 7 días de publicadas, porque tu configuración global de pnpm exige `minimumReleaseAge: 10080` en modo estricto. Eso deja Starlight en 0.42.4 (no 0.42.5), Biome en 2.5.14 (no 2.5.15) y wrangler en 4.143.0 (no 4.147.0). Además pnpm 12.9.1 no arranca cuando se llega a él desde pnpm 11 a través de `packageManager` ([pnpm#16594](https://github.com/pnpm/pnpm/issues/16594), corregido pero sin versión publicada), así que se fija 12.8.1, la última que funciona. El script de instalación de `esbuild` queda denegado (`allowBuilds: esbuild: false`), porque su binario llega como dependencia opcional y el build funciona sin él. Los cambios incompatibles de Starlight 0.42 (versiones mínimas, navegadores de 2023-2024 en adelante, marcado del botón del menú móvil) no afectan a un sitio nuevo. Los de pnpm 12 tocan instalaciones globales y `pnpm add` de gestores de paquetes, y conserva `node_modules/.pnpm`.
- **Marca.** `.github/assets/hive-mark.svg` y `hive-logo.svg` usan `#ff2a2a` como acento, con `#8c2028`, `#551318`, `#62565a` y `#fff5f2`. Están pensados para fondo oscuro.
- **README.** 362 líneas, en inglés. La sección `A workflow with room for judgment` (líneas 252-294) enlaza cada flujo a su `SKILL.md`; la navegación de cabecera está en las líneas 8-14 y `Documentation` en 331-351.
- **Ramas.** GitHub muestra `development` por defecto. `master` avanza solo al publicar una versión, con push directo (`git push origin <sha>:master`, archivo de gh-11, `tasks.md:95`).
- **Cloudflare** (documentación oficial vía Context7 `/cloudflare/cloudflare-docs` y developers.cloudflare.com, 2026-10-06):
  - Para sitios nuevos recomienda Workers con *static assets* en lugar de Pages (`/workers/best-practices/workers-best-practices/`).
  - Workers Builds permite elegir carpeta raíz, rama de producción y rutas que disparan el build, y apagar los builds de vista previa (`/workers/ci-cd/builds/configuration/`, `/build-branches/`, `/build-watch-paths/`).
  - La imagen de build trae Node 24.18.0 y pnpm 10.11.1, y se cambian con `.node-version` o `NODE_VERSION` y con `PNPM_VERSION` (`/workers/ci-cd/builds/build-image/`).
  - Un Astro estático solo necesita `assets.directory` en `wrangler.jsonc` (`/workers/framework-guides/web-apps/astro/`).
  - Las Single Redirects corren antes que el Worker (`/ruleset-engine/reference/phases-list/`).
  - Un Custom Domain crea sus propios registros DNS y no admite un CNAME previo en el nombre; una ruta (*Route*) se apoya en un registro con proxy que ya exista, como el `AAAA 100::` actual. El plan usa la ruta, así que no cambia el DNS.
  - El plan gratuito da 3.000 minutos de build al mes.

Supuestos por confirmar en la publicación (T6):
- que Workers Builds instale las dependencias desde el lockfile y acepte `PNPM_VERSION=12.8.1` (D8-A tiene la vuelta a 11.28.5);
- que la Single Redirect siga respondiendo antes que la ruta del Worker (se comprueba con `curl` y la ruta se quita si no);
- que un despliegue con `wrangler deploy` conserve la ruta agregada en el panel (se comprueba volviendo a desplegar);
- cómo se interpretan las rutas que disparan el build.

## Diseño propuesto

### Estructura de `site/`

```
site/
  package.json        hive-docs · pnpm@12.8.1 · engines node >=24
  pnpm-lock.yaml
  pnpm-workspace.yaml allowBuilds con los scripts de build aprobados
  .node-version       24
  go.mod              módulo vacío que saca site/ del módulo raíz
  wrangler.jsonc      name · compatibility_date · assets · routes · workers_dev
  astro.config.mjs    title Hive · site https://hive.tricell.tech · logo · customCss · sidebar
  src/assets/hive-mark.svg        variante oscura (copia de .github/assets)
  src/assets/hive-mark-light.svg  variante clara (nueva, T4)
  src/styles/theme.css
  src/content.config.ts
  src/content/docs/index.mdx
  src/content/docs/flows/{research,plan,build,close}.md
```

- **Base.** Se parte de la plantilla `user-manual` ya actualizada (T1), así que el sitio prueba la plantilla. Se conservan sus scripts `check`, `build`, `preview` y `lint`.
- **Protección para Go.** `site/go.mod` contiene solo un comentario y `module hive.invalid/site`. Un `go.mod` anidado saca su árbol de `./...` del módulo raíz, sea cual sea el gestor de paquetes. Es más robusto que exigir pnpm, y AC4 lo prueba con un archivo `.go` roto dentro de `node_modules`, que es lo que deja un `node_modules` plano. Se descartó renombrar la carpeta a `_site`: la regla de Go se cumpliría, pero el nombre engaña.
- **Dependencias.** Las versiones son exactas, sin `^`. Se instala con `--frozen-lockfile`. Desde pnpm 11, la aprobación de scripts de instalación vive en `pnpm-workspace.yaml` bajo `allowBuilds` (así se hizo en la migración a pnpm 11 de agosto de 2026). Se aprueban solo los que el build necesite (por ejemplo `esbuild` o `sharp`), con un comentario que explique el motivo, y el archivo se versiona para que un build limpio no dependa de aprobaciones locales. Se corre `pnpm audit --prod` y su resultado se registra. Un resultado limpio no prueba ausencia de vulnerabilidades, según [security boundaries](../../../../content/skills/flow-build/references/security-boundaries.md).
- **`wrangler`** queda como dependencia de desarrollo fija (4.143.0), y el comando de despliegue es `pnpm exec wrangler deploy`, para no descargar otra versión con `npx`. `wrangler.jsonc` lleva:
  - `name: "hive-docs"`, que debe coincidir con el nombre del Worker que se crea en el panel;
  - `compatibility_date: "2026-10-03"`, la más nueva que soporta el `workerd` de wrangler 4.143.0 (con `2026-10-06`, `wrangler dev` falla);
  - `assets: { directory: "./dist", not_found_handling: "404-page" }`, para servir el `404.html` de Starlight;
  - en la versión 0.2.1, `workers_dev: true` y sin ruta, para comprobar el sitio en `workers.dev` antes de que el dominio reciba tráfico;
  - desde el PR #22 (D10-B), `routes: [{ pattern: "hive.tricell.tech/*", zone_name: "tricell.tech" }]` y `workers_dev: false`, para que cada despliegue mantenga la ruta y no publique una copia en `workers.dev`.

  Como la ruta está en la configuración, cualquier `wrangler deploy` publica en el dominio. Por eso solo se despliega desde la etiqueta de una versión, como dice `CONTRIBUTING.md`, y hace falta la sesión de la cuenta de Cloudflare.

  El implementador confirma estos campos en la documentación de Workers (Context7 `/cloudflare/cloudflare-docs`) antes de usarlos La comprobación local del 404 con `pnpm exec wrangler dev` la hace T3, después del primer build completo. Si `wrangler dev` exige iniciar sesión, pasa a T6.
- **Logo.** Starlight recibe `logo: { dark, light, alt: "" }` con `replacesTitle: false`. El ícono es decorativo y el título visible `Hive` lo nombra; así un lector de pantalla no anuncia "Hive Hive".
  - Variante oscura: copia de `.github/assets/hive-mark.svg`.
  - Variante clara: `hive-mark-light.svg`, nueva, con la geometría de `hive-mark.svg` y la paleta sin resplandor de `.github/assets/logo-light.png`: contornos rojos, relleno casi blanco y sin filtro de brillo. `hive-mark.svg` usa rellenos casi negros y un brillo hecho con filtro que, sobre fondo claro, se vería como hexágonos oscuros con halo rosado.
  - Las dos van en `site/src/assets/`, porque Starlight procesa el logo desde dentro del proyecto. Si cambia la marca en `.github/assets/`, se actualizan.

### Contenido

- **Idioma**: inglés.
- **Portada** (`index.mdx`): por pedido tuyo en la primera captura (2026-10-06), sigue la estructura de una guía de una sola página: listas simples en vez de tarjetas, en este orden:
  1. Qué es Hive.
  2. `Install`: el comando de una línea y cómo actualizar.
  3. `Start with these`: los cuatro flujos.
  4. `Getting started`: instalar, la sección `## Hive` y un pedido con límites.
  5. `The main flow`: research → plan → build → close.
  6. `Supporting skills`: las demás skills, con enlace a su `SKILL.md`.
  7. `Specialist roles`.
  8. `Learn more`.
- **Páginas de flujo** (`flows/<name>.md`), con este orden fijo de secciones: `When to use`, `What it produces`, `What it does not do`, `Related flows`, `FAQ` y un pie con `Source: content/skills/flow-<name>/SKILL.md` enlazado a `https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-<name>/SKILL.md`.
  - Se escriben desde el `SKILL.md`, las secciones del README y la guía global. Explican con palabras propias y ejemplos de pedidos, no reglas.
  - Las preguntas frecuentes responden dudas reales de uso, como "¿investigar autoriza a implementar?" o "¿necesito un plan para un cambio pequeño?". Cada respuesta remite al `SKILL.md` en vez de citar la regla.
- **Barra lateral**: `Start` (portada) y `Flows` (research, plan, build, close, en ese orden).
- **Diagrama**: Starlight no muestra Mermaid sin otra dependencia. La relación entre flujos va en `Related flows` como texto y en la portada como lista ordenada.

### Marca visual

- **Variables de color.** Se sobrescriben en `src/styles/theme.css` las tres variables de acento de Starlight por tema: `--sl-color-accent-low`, `--sl-color-accent` y `--sl-color-accent-high`. Los grises de Starlight no se tocan.
  - Tema oscuro: `accent-high` sale de `#ff2a2a`, porque Starlight pinta con él el texto de acento y los enlaces (`--sl-color-text-accent`). `accent-low` es el rojo oscuro que usa `--sl-color-text-invert` para el texto sobre fondos de acento, como la entrada actual de la barra lateral y el botón principal de la portada.
  - Tema claro: un rojo más oscuro de la misma familia, por ejemplo cerca de `#8c2028`, porque `#ff2a2a` sobre blanco no llega a 4.5:1.
  - Los valores finales salen de la medición. Las correspondencias entre variables del tema claro (qué variable pinta los enlaces y cuál el texto sobre acento) se leen en `node_modules/@astrojs/starlight/style/props.css` de la versión instalada antes de elegir los valores, porque la revisión no las encontró documentadas.
- **Pares que se miden en cada tema**, con `uv run --with coloraide python content/skills/starlight-docs-site/scripts/contrast-check.py <primer plano> <fondo>`:
  - texto de acento sobre `--sl-color-bg`, `--sl-color-bg-nav` y `--sl-color-bg-sidebar`;
  - `--sl-color-text-invert` sobre el fondo de acento de la entrada actual de la barra lateral;
  - el texto del botón principal de la portada sobre su fondo.
- **Sin efectos de brillo.** El brillo del logo vive en el SVG. El resto conserva la tipografía y la disposición de Starlight, sin componentes reemplazados.

| Vista | Ruta | Tarea del lector | Patrón | Estados |
| --- | --- | --- | --- | --- |
| Portada | `/` | entender qué es Hive y elegir un flujo | `splash` de Starlight con tarjetas | 1440, 1280, 390 y 320 px; claro y oscuro |
| Páginas de flujo | `/flows/<name>/` | decidir si usar el flujo y qué esperar | página `doc` estándar con tabla de contenidos | la página más larga, enlaces externos, buscador, menú móvil, 404 |

### Regla del README (D2-A)

La regla va en una sección nueva del `AGENTS.md` raíz, `README and llms.txt`, con tres puntos:
- **Qué contiene el README**: qué es Hive, la instalación, los hosts y cómo actualizar, recuperar o desinstalar, y, una vez publicado el sitio, el enlace a él. El detalle de cada flujo vive en el sitio.
- **Cuándo se actualiza**: el README y `llms.txt` se actualizan en el mismo PR que cambie algo que afirman, como comandos o flags de instalación, hosts, requisitos o comandos del manager visibles al usuario.
- **Páginas del sitio**: se actualizan en el mismo PR que cambie el comportamiento de un flujo que describen.

`CONTRIBUTING.md` agrega en `Branches and commits` una obligación para quien contribuye, no solo un enlace: el README, `llms.txt` y las páginas del sitio se actualizan en el mismo PR que cambie lo que afirman. Enlaza a la sección del `AGENTS.md` para el detalle. Hace falta decirlo ahí porque el mismo archivo aclara que los contribuyentes no tienen que seguir los procesos internos de `AGENTS.md`.

Como el PR principal agrega `site/`, el mismo PR actualiza lo que el README y `llms.txt` ya afirman sobre las carpetas del repositorio: `site/` aparece en `For AI agents` del README, en la cabecera de `llms.txt` y en el mapa del `AGENTS.md` como material que no se instala. Los enlaces al sitio publicado esperan a T6.

### Publicación (T6, D4-A y D7-A)

El orden evita dos riesgos: enlaces muertos en el README que muestra GitHub, y un corte de `/install.sh`.

1. **Promoción.** La versión lleva `site/` a `master` (con su propia orden). Antes de promover, si ya tienen más de 7 días, se actualizan en el lockfile las versiones corregidas de las tres vulnerabilidades que reportó `pnpm audit --prod` en T2 (ver [Dependencias del sitio](#dependencias-del-sitio-auditoría-de-t2)).
2. **Cloudflare, en el panel, lo haces tú:**
   1. En Workers & Pages, crear la aplicación `hive-docs` (el mismo `name` de `wrangler.jsonc`) e importar `JhonHawk/tricell-hive`.
   2. Configurar el build:
      - Carpeta raíz: `site`.
      - Comando de build: `pnpm build`.
      - Comando de despliegue: `pnpm exec wrangler deploy`.
      - Variable: `PNPM_VERSION=12.8.1`.
   3. En la configuración del build:
      - Rama de producción: `master`.
      - Builds de vista previa: apagados.
      - Rutas que disparan el build: `site/*`. La revisión no pudo confirmar si se interpreta desde la raíz del repositorio o desde la carpeta raíz del build; se comprueba en el punto 4.
3. **Comprobación en `workers.dev`.** El primer build termina bien y el sitio responde en su URL `workers.dev`, incluida una ruta inexistente con `404`. Si el build falla por pnpm 12, se aplica la vuelta de D8-A: PR a `development` con `packageManager` y `PNPM_VERSION` en 11.28.5 y el lockfile regenerado, y su promoción.
4. **Ruta.** Agregar en el panel la ruta `hive.tricell.tech/*` del Worker, sobre el `AAAA 100::` con proxy que ya existe. No se toca el DNS ni la Single Redirect. Enseguida se corren los `curl` de AC8. Después se vuelve a desplegar desde el panel y se comprueba que la ruta sigue y que `/install.sh` sigue respondiendo `302`.
5. **Enlaces.** Con el sitio en línea, PR a `development` con los enlaces del README y de `llms.txt` (AC9). El mismo PR pone `workers_dev: false`, que es la receta de Wrangler para rutas gestionadas solo desde el panel y quita la copia duplicada en `workers.dev`. Ese cambio llega a `master` con la siguiente versión, y entonces se repite la comprobación de que la ruta persiste.

**Lo que se hizo (D10-B, 2026-10-06)**: en lugar de los pasos 2 a 4 en el panel, el despliegue se hizo con `wrangler` desde la CLI, desde una copia limpia de la etiqueta. Primero sin ruta, para comprobar `workers.dev`, y después con `--route hive.tricell.tech/*`. Conectar Workers Builds en el panel queda fuera de #17; si se conecta, usa la misma `wrangler.jsonc`.

**Recuperación**: quitar la ruta en el panel corta el tráfico al momento, y `hive.tricell.tech` vuelve a quedar solo con la redirección. Como la ruta está en `wrangler.jsonc`, también hay que quitarla de ahí, en un PR, antes del siguiente despliegue; si no, el despliegue la vuelve a crear. No hay registros DNS que restaurar.

### Dependencias del sitio (auditoría de T2)

`pnpm audit --prod` en T2 (2026-10-06) reporta tres avisos, todos en dependencias indirectas de `@astrojs/starlight` que solo corren al compilar:

| Paquete | Aviso | Instalado | Corregido | Por qué no se explota aquí |
| --- | --- | --- | --- | --- |
| `http-cache-semantics` | GHSA-ch52-4w7c-c8xp, alto | 4.2.0 | 4.3.0, del 2026-10-04 | expone respuestas entre usuarios de una caché compartida; el sitio es HTML estático sin servidor |
| `source-map-js` | GHSA-68fv-2mgg-jv7q, alto | 1.2.1 | 1.2.2, del 2026-09-30 | bloqueo con un source map manipulado; solo procesa los del propio build |
| `postcss-selector-parser` | GHSA-rj75-hqrm-r3gf, moderado | 6.1.4 | 7.1.6 | complejidad cuadrática con selectores manipulados; solo procesa el CSS del repositorio |

No se corrigen ahora:
- Las dos primeras versiones corregidas tienen menos de 7 días, así que la regla `minimumReleaseAge` de pnpm las rechaza.
- La tercera es un salto de versión mayor que `postcss-nested` (`^6`) no acepta.

Se revisan en el paso 1 de la publicación.
