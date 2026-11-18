# pr-before-after-captures — Capturas antes y después en el cuerpo del PR

| Campo | Valor |
| --- | --- |
| Estado | Borrador · revisión del plan aplicada, pendiente de aprobación |
| Tracker | ninguno |
| Git | `tricell-hive`, base `development` en `8ac8d77c`; rama propuesta `feat/pr-before-after-captures` |

## Decisions

2026-11-17
- **D2 A** — ¿De dónde sale la captura «antes»? La toma el build, en el mismo checkout, antes de su primera edición de cualquier cosa que la pantalla muestra; una pantalla nueva no tiene «antes».
- **D3 A** — ¿Qué se hace en un repositorio público? Se adjuntan solo capturas que muestran datos sembrados y cuentas de prueba, sin URLs internas ni correos reales; si el agente no puede asegurarlo, las capturas quedan locales y el PR lleva el texto de hoy.
- **D4 C** — ¿Cómo se entrega este cambio? PR listo: revisión, push y PR no borrador, espera de CI y de sus revisiones, y parada con el PR abierto para que el usuario lo fusione.
- **D5 A** — ¿Qué revisión corre? La nativa de Claude Code, `/code-review`, sobre el candidato antes del push.
- **D6 A** — ¿Las capturas «antes» van al visor? No: quedan en la carpeta del trabajo y van al PR; el visor sigue igual y no guarda copias de más.
- **D7 A** — ¿Qué autoriza subir capturas en un repositorio público o en GitHub Enterprise Server, donde quedan visibles y solo el soporte de GitHub las borra? Se pregunta una vez por trabajo, con la entrega o en la ronda de efectos del plan; sin respuesta, el PR lleva el texto de hoy y las capturas quedan locales.
- **D8 A** — ¿Se suben al PR de este cambio, en tricell-hive (público), 4 capturas de la página `flows/close.md` del sitio (antes y después, a 1280 y 440 px)? Sí.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> «con el fin de evitar que se llene de imagenes o archiven, es posible que estas se coloquen como cuerpo en el pr no? before | after» (usuario, 2026-11-17)

## Objective

Hoy las capturas de un cambio de interfaz solo se toman del candidato (verify), se guardan sin versionar en `_support/workspace/<fecha>-<trabajo>/images/` y se publican en el visor; el cuerpo del PR describe la evidencia solo en texto, así que las imágenes se acumulan en local. Al terminar, el build captura el «antes» de lo que va a cambiar, el PR lleva una tabla antes | después con las imágenes subidas con `gh … --attach` (nunca al repositorio), y el cierre propone borrar las copias locales cuya subida comprobó en el PR fusionado y conserva las demás.

## Scope and acceptance

Dentro: la captura «antes» en el build y su registro en el checkpoint, la tabla del PR y su respaldo, la regla de repositorio público y su autorización, la conservación y el borrado al cerrar, y la documentación que describe estos pasos. Fuera: el visor (no recibe las capturas «antes» y su almacenamiento no cambia), cómo verify captura el candidato, otros forges distintos de github.com y GitHub Enterprise Server, y videos.

- AC1 — **Retired:** la revisión del plan mostró que el disparador, los datos y la delegación no estaban definidos; replaced by AC8. Antes de su primera edición de una pantalla existente, el build captura cada región que el plan cambia, con los anchos, temas y estados que verify va a revisar y los mismos datos sembrados; una pantalla nueva no tiene captura «antes» · false today because `flow-build/SKILL.md` y `flow-verify/references/ui-verification.md` solo capturan el candidato · Observed in: the installed `flow-build` skill and `ui-verification` reference on the seven hosts after the local refresh
- AC2 — **Retired:** la tabla debe cubrir temas y estados y el host es github.com; replaced by AC9. Cuando el walk de verify produjo capturas y el remoto es GitHub con un `gh` cuya ayuda de `pr create` lista `--attach`, el cuerpo del PR lleva una tabla antes | después por región y ancho, con las imágenes subidas con `--attach` y ninguna imagen añadida al commit · false today because `git-workflow` paso 6 solo pide texto · Observed in: the installed `git-workflow` skill and its new reference, and the body of this change's own pull request
- AC3 — **Retired:** GitHub Enterprise Server sigue la regla pública y una subida parcial necesita su manejo; replaced by AC10. Sin `--attach` en el `gh` instalado, o con un remoto que no es GitHub, el cuerpo del PR queda como hoy, nombra las capturas como conservadas en local y no se intenta otra subida · false today because no rule covers attaching images · Observed in: the installed `git-workflow` reference
- AC4 — **Retired:** D7 añade la autorización única por trabajo; replaced by AC11. En un repositorio público solo se adjuntan capturas que muestran datos sembrados y cuentas de prueba, sin URLs internas ni correos reales; si el agente no puede asegurarlo, las conserva en local y sigue AC3 · false today because no rule covers public repositories · Observed in: the installed `git-workflow` reference
- AC5 — **Retired:** hoy el cierre ya trata todo el material como reemplazado por el PR, así que el cambio endurece la conservación; replaced by AC12. Al cerrar una obra integrada, las capturas cuya subida el agente comprobó en el cuerpo del PR fusionado se proponen para borrar, y las no subidas se conservan · false today because `support-material` y `artifact-layout` tratan el cuerpo del PR como reemplazo sin imágenes · Observed in: the installed `flow-close` references `support-material` and `artifact-layout`
- AC6 — Las páginas del sitio de Build, Verify y Close, y la arquitectura de `_support`, describen la captura «antes», la tabla del PR y el borrado al cerrar · false today because they describe text-only evidence · Observed in: the locally built documentation site and `_support/docs/architecture/workspace-and-artifacts.md`
- AC7 — **Retired:** ya es cierto en la base (todos los archivos caben y los tests pasan); el límite de bytes pasa a la verificación de T2 y T6. `git-workflow/SKILL.md`, `flow-build/SKILL.md` y cada referencia tocada quedan en 16 000 bytes o menos, y `go test ./tests/content/...` pasa · Observed in: the local content test run on the final candidate
- AC8 — **Replaces:** AC1. Antes de la primera edición de cualquier cosa que una pantalla existente muestra (su componente, estilo, diseño, contrato de datos o datos sembrados), el hilo principal del build sirve el checkout sin cambios con los datos que tenga y captura cada región que el plan cambia, o que la petición cambia cuando no hay plan, en los anchos, temas y estados que verify revisará; lo hace antes de lanzar cualquier tarea hija que edite esa superficie; no publica esas capturas en el visor; y el checkpoint guarda sus rutas o dice que no hay. Una pantalla nueva no tiene «antes», y un «antes» perdido por una edición previa no se reconstruye · false today because `flow-build/SKILL.md` y `ui-verification.md` solo capturan el candidato · Observed in: the installed `flow-build` skill and `ui-verification` reference after the local refresh that follows the merge
- AC9 — **Replaces:** AC2. Cuando el walk de verify produjo capturas, el remoto es github.com y la ayuda de `gh pr create` lista `--attach`, el cuerpo del PR lleva una tabla antes | después por región, ancho, tema y estado capturados, con las imágenes subidas con `--attach`, ninguna imagen en el commit, y una celda que dice «pantalla nueva» o «sin antes» cuando falta un lado · false today because `git-workflow` paso 6 solo pide texto · Observed in: the body of this change's own pull request after the push (4 GitHub-hosted images of `flows/close.md` and no image in `git diff --stat origin/development...HEAD`), and the installed `git-workflow` reference after the refresh
- AC10 — **Replaces:** AC3. Sin `--attach` en el `gh` instalado, o con un remoto que no es github.com ni GitHub Enterprise Server, el cuerpo del PR queda en texto y nombra las capturas como conservadas en local; si solo algunas subidas fallan, el agente completa las que faltan con `gh pr edit --attach` y toda edición posterior del cuerpo parte de `gh pr view --json body`, nunca del archivo local · false today because no rule covers attaching images · Observed in: the installed `git-workflow` reference after the refresh
- AC11 — **Replaces:** AC4. En un repositorio público, o en GitHub Enterprise Server, el agente adjunta solo capturas con datos sembrados y cuentas de prueba, sin URLs internas ni correos reales, y solo con una respuesta del usuario que autorice la subida pública de ese trabajo, preguntada una vez con la entrega o en la ronda de efectos del plan; sin ella, o en la duda sobre los datos, sigue el respaldo de texto · false today because no rule covers public repositories · Observed in: the installed `git-workflow` reference after the refresh, and D8 in this plan as its first use
- AC12 — **Replaces:** AC5. Al cerrar, una captura es reemplazable por el PR solo si el agente comprobó su URL subida en el cuerpo del PR fusionado; las capturas no subidas, incluidas las de la limpieza del visor, se conservan · false today because `support-material` «Delete» y `artifact-layout` «Retain evidence» tratan el cuerpo de un PR fusionado como reemplazo de todo el material de verificación · Observed in: the installed `flow-close` references `support-material` and `artifact-layout` after the refresh

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/git-workflow/SKILL.md` paso 6 es el hogar canónico del cuerpo del PR; el archivo pesa 15 930 bytes, a 70 del límite de `tests/content/skill_size_test.go`. El paso 9 incluye la frase «Step 7 already reads reviews and comments through REST.» (57 bytes), que solo describe el paso 7.
- `content/skills/flow-build/SKILL.md` (15 613 bytes): la línea 10 publica en el visor toda captura del build; el paso 4 sirve la primera vista solo después de la tarea de UI, a veces con datos de prueba; el párrafo del checkpoint lista lo que se guarda.
- `content/skills/flow-close/references/artifact-layout.md`: «Retain evidence» hace del cuerpo del PR integrado el reemplazo de la evidencia; el párrafo de imágenes de revisión de UI permite borrar las entradas de subida tras la lectura del visor. `support-material.md` «Delete» admite borrar el material que el cuerpo de un PR fusionado reemplaza. `viewer-publication.md` «Close and input hygiene» remite a `artifact-layout` para qué conservar.
- `site/src/content/docs/flows/close.md:14`, `flows/verify.md`, `flows/build.md` y `_support/docs/architecture/workspace-and-artifacts.md:53,82` describen la evidencia como texto; `_support/docs/architecture/instruction-resources.md` lista los hogares canónicos y debe nombrar la referencia nueva.
- `JhonHawk/tricell-hive` es público (`gh repo view --json visibility`, 2026-11-17). El sitio es responsive, así que su captura incluye 440 px.

- `gh pr create`, `gh pr edit` y `gh pr comment` aceptan `--attach` (hasta 50 archivos; reescriben `![alt](./archivo.png)` a la URL subida; las rutas de `--body-file` son relativas a su carpeta; una subida parcial crea el PR y sale con error; `pr edit` sin cuerpo añade la imagen al final) · ayuda de `gh` 2.104.0 instalado y documentación oficial vía Context7 (`/websites/cli_github_manual`) · presente · 2026-11-17
- Quién ve una imagen subida a un PR · documentación oficial de GitHub vía Context7 (`/github/docs`, «Attaching files» y «About anonymized URLs») · en repos públicos, cualquiera; en repos privados e internos de github.com, solo quien tiene acceso al repo; en GitHub Enterprise Server, cualquiera con la URL aunque el repo sea privado; quitar el enlace no borra el archivo, solo el soporte de GitHub lo hace · 2026-11-17
- Versión de `gh` desde la que existe `--attach` · no medido; la regla detecta la opción en la ayuda en lugar de exigir una versión; si la detección falla, aplica el respaldo de texto · 2026-11-17

## Effects

- Subir 4 capturas públicas de `flows/close.md` del sitio al PR de este cambio (quedan públicas; solo el soporte de GitHub las borra) · target: `JhonHawk/tricell-hive` · ceiling: 0 · tasks: T6 · answer: approved, D8 A

## Delivery

**D4 C** — PR listo: revisión, push y PR no borrador hacia `development`, espera de CI y de sus revisiones, y parada con el PR abierto para que el usuario lo fusione; no autoriza merge, auto-merge ni la actualización de la instalación local, que sigue al merge como pide `AGENTS.md`. **D5 A** — revisión nativa de Claude Code `/code-review` sobre el candidato antes del push.

AC8, AC10, AC11 y AC12 se observan en el texto instalado solo tras la actualización local que sigue al merge; AC9 se observa en el cuerpo del PR tras el push. En la parada del PR abierto quedan como no verificados en su lugar, con el paso que los observa. AC8 se observa como texto de la regla, no como un build que la siga, porque los pilotos están pausados.

## Review status and progress

2026-11-17 — Plan escrito tras D2 A y D3 A.
2026-11-17 — Revisión independiente (`hive-review-plan`): 2 P1, 7 P2, 3 P3. Aplicado: el «antes» se toma antes de cualquier edición de lo que la pantalla muestra, por el hilo principal y antes de delegar, con su ruta en el checkpoint (AC8); manejo de subidas parciales y ediciones del cuerpo (AC10); la conservación cubre la limpieza del visor (AC12); la captura de este PR se toma antes de T4 y a 440 px; criterios observables solo tras el push o el merge marcados; AC7 retirado; frase a recortar en `git-workflow` nombrada; `instruction-resources.md` añadido. D6 A, D7 A y D8 A resolvieron los dos P1.
