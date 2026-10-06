# Sitio de documentación por flujo y regla del README

| Campo | Valor actual |
| --- | --- |
| Estado | Esperando publicación · T1–T5 entregados en el [PR #18](https://github.com/JhonHawk/tricell-hive/pull/18) a `development` · T6 pendiente de la próxima versión |
| Tracker · GitHub Issues | • [#17 — Sitio de documentación por flujo y regla de actualización del README](https://github.com/JhonHawk/tricell-hive/issues/17) |
| Git | interactiva · rama → validación tuya del sitio compilado → push, PR a `development`, `/code-review`, merge · publicación (T6) con la próxima versión y orden aparte |
| Verificación | prueba de la plantilla · build y `astro check` del sitio · `go vet`/`go test` con la protección · contraste · `hive-review-ux` y `hive-verify-change` · suite local · comprobación en vivo al publicar |
| Siguiente paso | T6 con la orden de publicar la próxima versión ([design.md](design.md#publicación-t6-d4-a-y-d7-a)) |

## Objetivo

Hive no tiene documentación para personas que explique cada flujo: cuándo usarlo, qué produce, qué no hace y cómo se encadena con los demás. El `README.md` cubre sobre todo la instalación, `llms.txt` está pensado para agentes y los `SKILL.md` están escritos para modelos. Tampoco hay una regla que diga cuándo se actualiza el README.

Este cambio agrega un sitio Starlight en `site/`, con una portada y una página por flujo (research, plan, build, close) que describe el comportamiento y enlaza al `SKILL.md` en vez de copiar sus reglas. El sitio se publica desde `master` en `https://hive.tricell.tech/` con Cloudflare Workers, sin tocar la redirección de `/install.sh`. Además, el `AGENTS.md` raíz gana la regla de qué contiene el README y cuándo se actualiza, y la plantilla `starlight-docs-site` que Hive distribuye pasa a versiones actuales.

## Alcance

Incluido:
- Plantilla `starlight-docs-site` al día (D9-A): versiones en sus dos `package.json`, en `SKILL.md` y en `tests/skills/starlight_assets_test.py`.
- `site/`: esqueleto desde la plantilla `user-manual`, protección para Go, `.gitignore`, configuración de Cloudflare (`wrangler.jsonc`), marca visual y contenido en inglés.
- Regla del README y de `llms.txt` en el `AGENTS.md` raíz y una línea en `CONTRIBUTING.md` (D2-A), más una entrada en `CHANGELOG.md` bajo `[Unreleased]`.
- Delta de requisitos en `specs/documentation-site/spec.md`.
- Publicación (T6, D4-A): enlaces del README y de `llms.txt` al sitio, promoción a `master` con la próxima versión y configuración de Cloudflare. Se ejecuta solo con orden explícita.

Excluido:
- **Páginas de roles, instalación o del manager.** El issue pide empezar por los cuatro flujos; la instalación sigue en el README.
- **Mover la redirección de `/install.sh`.** Sigue como regla de la zona de Cloudflare (D7-A); no se duplica en `_redirects`.
- **Vistas previas por rama.** Solo `master` compila; la validación previa es local.
- **Glosario y registro del porqué de las decisiones.** Quedaron en observación, sin issue.
- **Cambios de DNS.** El Worker se conecta con una ruta `hive.tricell.tech/*` sobre el registro de relleno actual; no se borra ni se crea ningún registro.
- **`.gitattributes` con `export-ignore`.** El paquete ya excluye `site/` por lista cerrada; `hive update` lo extrae pero no lo despliega.

Restricciones que deben seguir cumpliéndose:
- La suite local (`go vet ./...`, `go test ./...`, pruebas Python de `tests/skills`) pasa.
- El paquete de versión y lo que despliega `hive update` no cambian de contenido salvo la plantilla de la skill.
- `https://hive.tricell.tech/install.sh` sigue respondiendo `302` hacia `get-hive.sh` en `master`.
- Las páginas no copian reglas: cada regla sigue teniendo un solo lugar.

## Criterios de aceptación

- AC1. La plantilla `starlight-docs-site` declara Astro 7.3.5, Starlight 0.42.4, `@astrojs/check` 0.9.10, TypeScript 6.0.3, Biome 2.5.14, sharp 0.35.5 y `pnpm@12.8.1` en sus dos perfiles; `SKILL.md` nombra las versiones de Astro, Starlight, TypeScript y pnpm; la prueba exige todas, y una copia del perfil `user-manual` instala con pnpm 12.8.1 y compila sin aprobar nada a mano, con las aprobaciones de scripts versionadas en su `pnpm-workspace.yaml`. *Falso en la base cuando* los `package.json` dicen `7.0.7`, `0.41.3` y `pnpm@11.21.0`.
- AC2. Sobre una copia exportada del commit (`git archive HEAD site`), `pnpm install --frozen-lockfile`, `pnpm check` (0 errores) y `pnpm build` terminan bien sin aprobar nada a mano, y `dist/` contiene `index.html`, `404.html` y `flows/{research,plan,build,close}/index.html`. *Falso en la base cuando* `site/` no existe.
- AC3. Cada página de flujo tiene los encabezados exactos `## When to use`, `## What it produces`, `## What it does not do`, `## Related flows` y `## FAQ`, enlaza a su `SKILL.md` en `master` y no comparte ninguna secuencia de 8 palabras seguidas (sin formato Markdown, en minúsculas) con `content/skills/flow-*/**` o `content/guidance/global.md`. *Falso en la base cuando* las páginas no existen.
- AC4. Con un archivo `site/node_modules/fake/bad.go` que importa un paquete inexistente, `go vet ./...` y `go test ./...` desde la raíz terminan bien. *Falso en la base cuando* ese mismo archivo hace fallar `go vet ./...`.
- AC5. `git check-ignore` reconoce `site/node_modules`, `site/dist`, `site/.astro` y `site/.wrangler`. *Falso en la base cuando* no imprime ninguno.
- AC6. La cabecera muestra el ícono de Hive (variante oscura en tema oscuro y clara en tema claro, decorativo, con `alt=""`) junto al título visible `Hive`, y cada par de color de [design.md](design.md#marca-visual) alcanza al menos 4.5:1 en los dos temas, medido con `contrast-check.py`. *Falso en la base cuando* no hay sitio.
- AC7. El `AGENTS.md` raíz tiene una regla que dice qué contiene el README y que el README y `llms.txt` se actualizan en el mismo PR que cambie algo que afirman; `CONTRIBUTING.md` la exige a quien contribuye, y el README (`For AI agents`), `llms.txt` y el mapa del `AGENTS.md` nombran `site/` como material que no se instala. *Falso en la base cuando* ninguno de esos archivos tiene la regla ni nombra `site/`.
- AC8. Tras publicar, `https://hive.tricell.tech/` y `https://hive.tricell.tech/flows/plan/` responden `200` con el sitio, una ruta inexistente responde `404` con la página de Starlight, `https://hive.tricell.tech/install.sh` responde `302` hacia `https://raw.githubusercontent.com/JhonHawk/tricell-hive/master/get-hive.sh` antes y después de volver a desplegar. *Falso en la base cuando* la raíz no tiene origen.
- AC9. Tras publicar, el README enlaza el sitio en su navegación y en `Documentation`, sus viñetas de flujos apuntan a las páginas del sitio y `llms.txt` lista el sitio. *Falso en la base cuando* ninguno menciona el sitio.

## Decisiones

| ID | Decisión (2026-10-06) |
| --- | --- |
| D1-A | El sitio vive en este repositorio, en una carpeta propia. |
| D2-A | La regla del README es solo de este repositorio: `AGENTS.md` raíz y `CONTRIBUTING.md`. |
| D3-A | Un solo issue (#17) para sitio y regla. |
| D4-A | Se publica desde `master`; sale con la próxima versión, y los enlaces del README al sitio se agregan en ese paso. |
| D5-A | Entrega interactiva. |
| D6-B | Revisión de código con `/code-review` de Claude Code. |
| D7-A | Hosting en Cloudflare Workers con *static assets* y Workers Builds, conectado por ruta y no por Custom Domain (propuesta de la revisión de infraestructura). |
| D8-A | pnpm 12.8.1; si Workers Builds no lo acepta en el primer build, se baja a 11.28.5. |
| D9-A | La plantilla distribuida se actualiza en este mismo cambio. |
| D10 | Publicación (2026-10-06): primer despliegue y ruta con `wrangler` desde la CLI, en la cuenta de Jmartinez@tricell.com.mx; conectar Workers Builds en el panel queda para después, fuera de #17. Por eso AC8 ya no exige leer la configuración de builds. |

Decidido por convención: el sitio está en inglés, porque el README y el repositorio público lo están y el texto para usuarios sigue el idioma del producto.

## Entrega

| Repositorio | Base | Modo | Incluye | Con orden aparte |
| --- | --- | --- | --- | --- |
| `JhonHawk/tricell-hive` | `development` | interactiva (D5-A) | rama `feat/gh-17-docs-site`, commits locales, parada con el sitio compilado en `pnpm --dir site preview` para tu validación; después push, PR que referencia #17, `/code-review` (D6-B) sobre el diff, merge cuando pasen la suite y la revisión, `hive update` local después del merge porque cambia `content/` | T6: PR de enlaces del README, promoción a `master` con la próxima versión, Cloudflare |

Como entre el merge y la próxima versión pueden pasar semanas, el registro de cambio viaja en el PR principal con estado "esperando publicación", sin archivarse. Después de T6 se archiva en un PR pequeño de cierre sin revisión dedicada que mergeo yo, como en gh-11. #17 se cierra cuando se cumplan AC8 y AC9, porque el issue incluye la publicación; el merge a `development` no lo cierra.

La orden de publicar (T6) cubre también la vuelta de D8-A si el primer build falla por pnpm 12: un PR a `development` que baja `packageManager` a 11.28.5 con el lockfile regenerado, y su promoción a `master`.
