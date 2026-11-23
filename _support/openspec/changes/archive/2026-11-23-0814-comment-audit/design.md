# Design

## Approaches

- **A — Plantilla JS/TS + algoritmo (elegida, D3 A, parser por D5 A).** La skill incluye `assets/ratchet/check-comments.mjs` y su prueba, escritas de nuevo a partir del algoritmo que funcionó en un repositorio real: un parser lista los comentarios de cada archivo y el lint compara esa lista con la del mismo archivo en la base de Git. El parser es `@babel/parser` (plugins `typescript` y `jsx`), que devuelve `ast.comments` en 7.x y 8.x; TypeScript 7.0.2 ya no exporta el parser clásico. Para otros lenguajes, `references/ratchet.md` describe el algoritmo y las pruebas mínimas que el agente implementa con el parser del proyecto. Está en el peldaño 4 de «Reuse before writing» (dependencia instalada) en los proyectos que ya tienen Babel, y en el 5 en los demás, con el check de paquete.
- **B — Script Python con su propio lexer.** Descartada: maquinaria nueva que debe acertar con strings, template literals y regex en cada lenguaje.
- **C — Solo el algoritmo.** Descartada: cada proyecto reescribiría el script y el resultado variaría entre ejecuciones.

## Decisions

- **D4 A — `no-comments` en proyectos con el lint.** Comparación (decisión local del 2026-11-19 sobre evaluar reglas): pstack `no-comments` solo conserva comentarios sobre lo que no controlamos y no contempla un lint; Hive antes de `8b89ed0d` no tenía regla; el leaf actual conserva los Keep en el código. Con el lint, conservarlos hace fallar el CI; moverlos conserva el hecho y deja pasar el CI.
- **D5 A — `@babel/parser`.** Un solo camino que sirve con TS 5, 6 y 7 y con JS puro. Cuesta una devDependency en los proyectos que no la tengan.
- **D6 A — Marca de atajo de una línea.** La regla de reutilización (`reuse-before-writing.md:6`, caso a) reconoce un atajo por su comentario; una marca de una línea mantiene esa señal en el código sin prosa. Forma: el token de comentario de línea del lenguaje seguido de `workaround: <destino>`, sola en su línea (`// workaround: COM-926` en JS/TS, `# workaround: COM-926` en Python). El destino es la clave o URL del ticket; mientras el ticket no existe, o si el usuario no quiere uno, es la entrada del registro de decisiones (`<ruta>#<ancla>`). La marca va solo sobre código que es un atajo; una excusa sin atajo debajo, como «TODO: validar», se borra cuando su issue existe. El lint comprueba solo la forma; la revisión comprueba que el ticket existe o que el ancla está en el registro, y reporta un destino inexistente como defecto. Cuando se crea el ticket, la marca pasa a nombrarlo. Hogar único de la forma: el leaf `no-comments`; `defect-fix` (3.882 B de 4.000) solo agrega un enlace a ese hogar, y `suppression-review` y la página de build del sitio enlazan lo mismo. Supuesto: destino de registro mientras no hay ticket. Descartado: exigir el ticket antes de escribir el atajo, que obliga a escribir en el tracker durante un incidente, o dejar el atajo sin marca si el ticket se rechaza, que lo deja expuesto a que lo copien.

## Proposed shape

```
content/skills/comment-audit/
  SKILL.md                     procedimiento (≤ 16.000 B)
  assets/hive-managed.json     {"managed_by": "tricell-hive"}
  assets/ratchet/check-comments.mjs
  assets/ratchet/check-comments.test.mjs
  references/ratchet.md        algoritmo, parser, otros lenguajes, conexión a lint y CI
  references/project-skill.md  estructura, contenido y registro de la skill generada
tests/content/comment_audit_template_test.go
tests/content/testdata/comment-audit/package.json + package-lock.json
```

**SKILL.md**, como `test-audit`, se suma a los flujos: alcance e inventario son investigación y no editan; el plan se suma a `flow-plan`; tandas, lint y skill generada se suman a `flow-build`; luego `flow-verify` y `flow-close`. El argumento del comando nombra la auditoría (`/flow-plan comment-audit <área>`).

1. **Alcance.** Toda la base de código o un área; los lenguajes presentes; las rutas generadas que se excluyen.
2. **Inventario y estimación.** `--inventory` a una línea base fijada en un commit. Antes del ledger, revisar los comentarios que el inventario no marcó como directiva y agregar a `ALLOWED` las directivas propias del proyecto (de herramientas o de build). Cuenta comentarios, líneas ocupadas solo por comentarios (`fullLine`) y líneas mixtas por separado, sin contar directivas (`directive`).
3. **Ledger.** Categorías y su correspondencia con el leaf `no-comments`: Tautología → Kill; Obsoleto → Kill (registra un cambio pasado) o Keep que falla su paso 4; Registro de decisión → Keep; Restricción → Constraint, con su check propuesto; Excusa → nueva: su issue, y la marca de atajo solo si hay código de atajo debajo; Directiva → se queda, salvo una supresión de lint o de tipos, que va a `suppression-review`. Las clases mecánicas (separadores, etiquetas como `// Arrange`) se registran con el patrón exacto que las encontró. Con más de 200 comentarios, un hijo `hive-research` por carril, como en `test-audit`.
4. **Plan.** Lo pone `flow-plan`, más: el registro de decisiones de destino, el tracker de las excusas (un padre con hijos por `create-issue`), las tandas y el lint.
5. **Tandas.** Primero los issues de las excusas y las entradas del registro; después se borran los comentarios. Cada tanda pasa `--same-code --base <commit de la tanda anterior>`. Una edición masiva con script se informa con `git diff --stat`, porque no aparece en el panel de cambios por turno del host.
6. **Lint gradual.** Desde la plantilla o el algoritmo; se conecta al script de lint del proyecto y a su CI, siempre con `--base` explícito (la base del PR, o `before` en un push) y con historial suficiente (`fetch-depth: 0` o equivalente).
7. **Skill del proyecto.** Como fija `references/project-skill.md`.
8. **Informe.** Conteos por categoría, issues, registros de decisión, comentarios pendientes, estimación de líneas, checks y su resultado.

**Plantilla.**
- Parámetros: `--base <rev>` (por defecto la merge-base con `origin/HEAD`; sale con 2 si no resuelve), `--inventory`, `--same-code`. Constantes al inicio: `SOURCE_PATTERNS`; `EXCLUDED` (rutas exactas o prefijos terminados en `/`); `ALLOWED`, regex de directivas: licencia (`@license`, `@preserve`, `/*!`), `/// <reference …>`, `@ts-check`, `@ts-nocheck`, `@ts-expect-error`, `@ts-ignore`, `eslint-disable*`, `eslint-enable`, `biome-ignore`, `prettier-ignore`, `oxlint-disable*`, `istanbul ignore`, `c8 ignore`, `v8 ignore`, `@vitest-environment`, `@jest-environment`, `@jsx`, `@jsxImportSource`, `@jsxRuntime`, `@jsxFrag`, `/*#__PURE__*/`, `/*@__PURE__*/`, `#__NO_SIDE_EFFECTS__`, `webpackChunkName` y demás comentarios mágicos de webpack, `@vite-ignore`, los bloques JSDoc de tipos (`@type`, `@typedef`, `@param {`, `@returns {`, `@template`) en archivos `.js`, `.mjs`, `.cjs` y `.jsx`, y la marca `^// workaround: \S+$`. El lint acepta la sintaxis de una supresión; su justificación la sigue revisando `suppression-review`. Y `DECISIONS_PATH`, que el mensaje de error nombra.
- Parser: `@babel/parser` cargado con `await import()` dentro del manejador de errores principal, para que su ausencia salga con 2. Plugins por extensión: `.ts`, `.mts`, `.cts` → `typescript`; `.tsx` → `typescript` y `jsx`; `.js`, `.jsx`, `.mjs`, `.cjs` → `jsx`; todos con `decorators-legacy`; `sourceType: "unambiguous"`, `errorRecovery: false`.
- Ejecución: `git` con `execFileSync` y un arreglo de argumentos, nunca un shell, `maxBuffer` de 256 MiB, `--end-of-options` antes de cada revisión, `-z` en toda lista de rutas, `--find-renames` para comparar un archivo renombrado con su ruta anterior; rutas resueltas contra `git rev-parse --show-toplevel`; solo archivos regulares (`lstat`).
- Comparación: multiconjunto por archivo de textos normalizados (cada línea sin sangría); un comentario permitido por `ALLOWED` no cuenta. `--same-code` compara los tokens de Babel (`tokens: true`) sin comentarios, base contra ahora, después de quitar en ambos lados los pares `{` `}` que solo contenían un comentario en JSX, unir los tokens de texto JSX contiguos y colapsar su espacio en blanco.
- Salidas: 0 limpio; 1 con `archivo:línea:columna: no-new-comments` en stderr por hallazgo; 2 ante cualquier error (base, git, parser ausente, archivo ilegible), nunca 0 en un error.
- Prueba (`node:test`, repositorios Git temporales con `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_NOSYSTEM=1` y `-c commit.gpgsign=false`): sin comentarios nuevos; comentario nuevo; duplicado; reintroducción tras mover la base; renombre; nombre no ASCII; corrida desde una subcarpeta; strings, template literals y regex; `.tsx` con `<p>// texto</p>` (no cuenta) y `{/* x */}` (cuenta); `.ts` con un cast `<string>x`; clase con decoradores; ruta excluida; directiva permitida (`/// <reference types="node" />`) y marca de atajo; base inválida (2); sin `node_modules` (2), con `check-comments.mjs` copiado a un directorio temporal nuevo sin ningún `node_modules` encima; `<p>a {/* x */} b</p>` borrado con `--same-code` (0); inventario con sus campos; `--same-code` en 0 (incluido borrar `{/* x */}` y una línea de comentario dentro de JSX en un `.tsx`) y en 1.

**Test Go.** Copia `assets/ratchet/` a un directorio temporal y corre `node --test` dos veces, con `node_modules/@babel/parser` enlazado a cada versión instalada en `tests/content/testdata/comment-audit/node_modules` (alias npm `babel8` y `babel7`). Busca esa carpeta en la raíz del módulo y, en un worktree sin ella, en el checkout principal (`git rev-parse --path-format=absolute --git-common-dir`). Se salta cuando falta `node` o la carpeta, con un mensaje que da el comando para instalarla (`npm ci --prefix tests/content/testdata/comment-audit`); falla con la salida de `node` en cualquier otro caso. `node_modules` de esa carpeta se ignora en Git.

**Otros lenguajes** (`references/ratchet.md`): cada port del lint usa el parser del lenguaje, permite la marca de atajo con el token de comentario de línea de ese lenguaje, e imprime en su mensaje de error la ruta del registro de decisiones.

**Skill generada** (`references/project-skill.md`): `.agents/skills/<nombre>/SKILL.md` con `name` y `description`, sin `assets/hive-managed.json`; contenido: el comando de inventario, las categorías con su acción, los patrones mecánicos usados, la ruta del ledger, el registro de decisiones, el tracker y los checks. Enlace relativo `../../.agents/skills/<nombre>` en `.claude/skills/<nombre>`, y en `.grok/skills/<nombre>` si existe `.grok/`. Línea en `## Skills` del `AGENTS.md` del proyecto («- Before auditing or deleting code comments, or when the comment lint fails: `.agents/skills/<nombre>/SKILL.md`»), con la frase de apertura de HA-IF-19 si la sección no existe, y la misma línea con el prefijo del hijo en el `## Skills` del workspace declarado. Comprobación: `skill_registry.py check` sobre el repositorio y sobre el workspace sale con 0.

## Impact

- `no-comments` lo carga el packet de `flow-verify` (generado al instalar) y el índice de `global.md`, cuyo momento no cambia.
- `defect-fix` lo leen `flow-build`, `reuse-before-writing`, `tests`, `flow-verify`, `urgent-fixes` y `hive-review-code`; el cambio agrega una frase para proyectos con el lint y no cambia el resto. `urgent-fixes.md:11` marca una mitigación durante un incidente, cuando no suele haber ticket: la marca con destino de registro de D6 lo cubre sin escribir en el tracker.
- La skill nueva no cambia los flujos existentes; solo se lee cuando la piden.
- `CHANGELOG.md` e `instruction-resources.md` también los edita `core-cleanup-batches`: quien se integre segundo fusiona.

## Recovery

Revertir el commit del merge. La instalación local se restaura con `hive update --source <checkout>` sobre el commit anterior. Los proyectos que ya generaron su skill la conservan; no dependen de la de Hive.
