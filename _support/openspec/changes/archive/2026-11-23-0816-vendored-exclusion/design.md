# vendored-exclusion — diseño

## Enfoque (D32 A)

- **A, elegido:** `--same-code` reporta cualquier cambio en una ruta de `EXCLUDED`; el texto agrega las rutas vendorizadas desde el paso 2, da a cada carril su archivo de inventario y corre `--same-code` al final. Un chequeo en el script no depende de que el agente lo recuerde, y el resto del texto solo pone el orden.
- **B, descartado:** el script no cambia; el paso 7 y la skill del proyecto agregan `git diff --name-only` y `git ls-files --others` limitados a las rutas excluidas, corridos al final.

## Script

`content/skills/audit-comments/assets/ratchet/check-comments.mjs`:
- Una función nueva, `excludedChanges(root, base)`: `git diff --no-ext-diff --no-renames --name-only -z <base> -- <cada entrada de EXCLUDED como :(top,literal)<entrada>>`, sin filtro de `SOURCE_PATTERNS`, así que lista cualquier archivo versionado de cualquier tipo que cambió, se agregó al índice o se borró bajo esas rutas (D35); cada ruta listada se vuelve a filtrar con `excluded()`, porque un pathspec sin `/` final también abarca una carpeta y `excluded()` lo trata como un archivo exacto. Con `EXCLUDED` vacío no corre.
- `sameCode` reporta primero cada archivo de esa lista como `<file>: changed under an excluded path`, sin leerlo ni parsearlo, así que un archivo vendorizado que `@babel/parser` no entiende da 1 y no 2. Después sigue con `changedFiles` como hoy; un archivo ya reportado no se reporta dos veces.
- Los archivos sin versionar bajo rutas excluidas no cuentan (D36): son salidas de build o de tests, y el commit de borrado nombra sus rutas.
- `lint`, `inventory` y `changedFiles` no cambian.

Suite (`check-comments.test.mjs`), con los casos escritos primero y vistos fallar:
- «--same-code rejects a change under an excluded path»: base con `src/a.js` y `dist/a.js`; quitar un comentario de `src/a.js` y editar `dist/a.js`: sale con 1 y nombra `dist/a.js`; revertir `dist/a.js`: sale con 0. El mismo caso con `dist/style.css` editado (otro tipo de archivo) y con `dist/bad.js` que no parsea: sale con 1, no con 2.
- «--same-code ignores untracked files under an excluded path»: un `dist/out.js` sin versionar y no ignorado: sale con 0.
- «--same-code names both sides of a move across an excluded path once»: un movimiento que sale de `dist/` nombra `dist/a.js`; uno que entra nombra `dist/s.js` y `moved from src/s.js`. También un archivo nuevo en el índice bajo `dist/`, y salida 0 al revertir todo.
- «the inventory skips excluded paths»: un comentario en `dist/a.js` no aparece en `--inventory`.

## Texto

- `SKILL.md` paso 1: deja fuera el código vendorizado, como una biblioteca copiada; en una oración propia, cada ruta completa desde la raíz (en un monorepo, `packages/web/vendor/`), porque `EXCLUDED` compara desde la raíz y sus valores por defecto solo cubren carpetas en la raíz.
- Paso 2: antes del primer inventario, agregar al `EXCLUDED` de la copia temporal las rutas del paso 1, conservando los valores por defecto, y escribir con modo `0600` la salida del inventario y todo archivo con texto de comentarios (archivos de carril y ledgers).
- Paso 4: cada carril recibe un archivo con sus líneas de inventario en la carpeta temporal del host (las líneas llevan el texto de cada comentario, que puede tener un secreto, y un brief no lleva secretos); el brief nombra esa ruta y el hijo marca solo esos comentarios.
- Paso 7: el orden queda lint de comentarios (una vez cableado en el paso 8), lint, type check y tests de los archivos tocados, y después `--same-code --base <el commit de registro>` como último chequeo antes del commit de borrado, porque esos comandos pueden regenerar archivos; un hallazgo bajo una ruta excluida se revierte.
- `ratchet.md`:
  - la línea de `--base` dice que el lint salta las rutas de `EXCLUDED`; la de `--inventory`, que lista los archivos fuera de `EXCLUDED`; la de `--same-code` suma a sus hallazgos «a tracked file changed under an `EXCLUDED` path, of any type»;
  - la viñeta `EXCLUDED`: se compara desde la raíz, incluye código vendorizado, y remite a la línea de `--same-code`;
  - la viñeta `ALLOWED` remite a la de `EXCLUDED`;
  - la frase «Until step 8» dice que `EXCLUDED` ya trae las rutas del paso 2;
  - el contrato de un port pide un modo `--same-code` que hace lo que fija la línea de `--same-code` del template, incluido el rechazo bajo rutas excluidas, y suma a la suite «a change under an excluded path fails `--same-code`».
- `project-skill.md` punto 6: «`--same-code --base <the batch's record commit>`, last, after the comment lint, the lint, the type check and the tests of the touched files; stage the removal commit by file name», porque D36 deja fuera de `--same-code` los archivos sin versionar y la skill del proyecto también corre sin Hive.
- Nombre del test de Go: `instruction-resources.md:28` y la descripción de `tests/content/testdata/audit-comments/package.json` dicen `comment_audit_template_test.go`; el archivo es `audit_comments_template_test.go`.
- `CHANGELOG.md`: la entrada de `audit-comments` en «Added» de `[Unreleased]` suma que el código vendorizado queda fuera y que `--same-code` rechaza cambios en rutas excluidas.
