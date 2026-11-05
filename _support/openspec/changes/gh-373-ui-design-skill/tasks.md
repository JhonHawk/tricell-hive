# Tareas

Base `development` en `dd729041`, rama `feat/gh-373-ui-skill` en el worktree `.claude/worktrees/gh-373-ui-skill`.

Quién hace cada tarea:
- **T1 y T4** crean o editan más de 3 archivos con su interfaz ya fijada en `design.md`, así que van a un hijo `hive-write-docs`, como manda [delegation](../../../../content/skills/flow-build/references/delegation.md).
- **T2 y T3** editan tres archivos cada una y se quedan en el hilo principal, al mismo tiempo que trabaja el hijo de T1, porque sus archivos no se cruzan.

El hilo principal revisa cada resultado contra `design.md` antes de marcarlo y hace los commits. Los enlaces de T2 y T3 a la skill se resuelven cuando T1 termina, por eso su verificación corre después de T1.

### T1 — La skill `hive-ui-criteria` y sus referencias

- [x] Existen `content/skills/hive-ui-criteria/SKILL.md`, `references/design-record.md`, `references/criteria.md`, `references/refinement.md` y `references/notice.md`, con el contenido que fija la sección «Forma propuesta» de `design.md`.

**Closes:** AC1, AC2, AC3, AC4, AC8

**Depends on:** ninguna

**Locations:** `content/skills/hive-ui-criteria/`

**Execution:** hijo `hive-write-docs`. El brief:
- nombra la carpeta de cambio y le pide leer `proposal.md` (con «Contexto verificado», que ya trae las fuentes de WCAG, APG y la documentación de las librerías), `design.md` y el spec delta antes de escribir;
- le pide seguir las reglas de redacción de `AGENTS.md` «Guidance structure»: inglés, umbrales observables, una casa por regla, nombrar la capacidad y no la herramienta de un host;
- le prohíbe copiar pasajes de Impeccable o de cualquier otra fuente externa, y nombrar los proyectos de referencia que no aprueba D11 B.

**Test approach:** check

**Changes:** solo archivos nuevos bajo `content/skills/hive-ui-criteria/`. No toca otras skills ni roles.

**Verification:**
- `go test ./tests/content/...` pasa: tamaño, frontmatter y enlaces.
- AC2: `rg -n "lockfile|manifest range" content/skills/hive-ui-criteria/references/design-record.md` encuentra la regla de la versión instalada, y `rg -n "HeroUI|PrimeNG" …/design-record.md` encuentra los dos ejemplos.
- AC3: `rg -n "pending decision" content/skills/hive-ui-criteria/SKILL.md` encuentra la regla de precedencia con sus cinco niveles.
- AC4: `rg -n "^## Hive|Sources|Watch|stale" …/design-record.md` encuentra cada parte del formato.
- AC8:
  - `rg -c "\[measured\]" …/criteria.md` da al menos 1 en cada sección de mínimos, colecciones, textos y casos extremos;
  - `rg -n "AAA|APG" …/criteria.md` encuentra las fuentes de D10;
  - `rg -n -i "never use|always avoid|banned|score" content/skills/hive-ui-criteria` no encuentra prohibiciones estéticas universales ni puntuaciones.
- Copia: el hilo principal compara cinco frases al azar de cada referencia contra el clon local de Impeccable (`rg -F`) y no encuentra coincidencias.
- El hilo principal comprueba punto por punto cada elemento de «Forma propuesta» contra los archivos escritos.
- Prueba de escritorio de la detección en `tooling/viewer/frontend`, solo lectura y sin escribir ningún DESIGN.md: siguiendo `design-record.md` se llega a la librería, a su versión en `pnpm-lock.yaml` (no al rango de `package.json`) y a la capa de personalización con sus rutas.

### T2 — `flow-plan` y los roles de diseño y construcción usan la skill

- [x] El paso 1 de «UI decisions» carga el sistema de diseño con `hive-ui-criteria`. Cuando el DESIGN.md falta o está desactualizado, el plan agrega una tarea que lo escribe o refresca; fuera de un plan, solo se escribe a pedido del usuario. `hive-design-ui` (pasos 1 y 2) y `hive-build-frontend` (pasos 1 y 3) leen el DESIGN.md como guía del proyecto, enlazan la skill y apuntan a su precedencia en vez de repetirla.

**Closes:** AC5, AC6

**Depends on:** ninguna

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/agents/design/hive-design-ui.md`, `content/agents/development/hive-build-frontend.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:**
- Solo punteros, más el reemplazo de los adjetivos del paso 2 de `hive-design-ui` y de las frases que repiten la precedencia.
- El resto de «UI decisions» (contrato de datos, regiones, bocetos) no cambia.
- `flow-plan/SKILL.md` sigue bajo 16 000 bytes.

**Verification:**
- Después de T1, `go test ./tests/content/...` pasa.
- AC5:
  - `rg -n "DESIGN.md" content/skills/flow-plan/SKILL.md` encuentra la tarea del plan cuando falta o está desactualizado;
  - en la skill, `rg -n "user asks|on request" content/skills/hive-ui-criteria/SKILL.md` encuentra la regla de fuera del plan.
- AC6:
  - `rg -n "hive-ui-criteria" content/agents/design/hive-design-ui.md content/agents/development/hive-build-frontend.md` da al menos un enlace en cada archivo;
  - `rg -n "improving hierarchy" content/agents/design/hive-design-ui.md` no encuentra nada;
  - `rg -n "change no token" content/agents/development/hive-build-frontend.md` no encuentra nada.

### T3 — La revisión UX juzga los checks medidos

- [x] Cambian tres archivos:
  - `hive-review-ux` (pasos 1 y 2) juzga contra el DESIGN.md, reporta como defecto P0–P3 cada check **[measured]** de `criteria.md` cuando el cambio lo introdujo, y apunta a la precedencia de la skill para los valores por defecto de la librería y las decisiones registradas;
  - `ui-verification.md` suma el ancho de 320 px CSS para el check de reajuste de las regiones cambiadas;
  - `delegation.md` «Reviewer brief» da cómo llegar a los estados que piden los checks medidos, o dice cuáles no se alcanzan.

**Closes:** AC7

**Depends on:** ninguna

**Locations:** `content/agents/review/hive-review-ux.md`, `content/skills/flow-verify/references/ui-verification.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:**
- No cambian la regla de veredicto, el alcance por regiones ni la auditoría automática; el spec «UX review in one pass» queda igual.
- La lista de checks no se copia en el rol.

**Verification:**
- Después de T1, `go test ./tests/content/...` pasa.
- `rg -n "project UI guide" content/agents/review/hive-review-ux.md` no encuentra nada, y `rg -n "hive-ui-criteria" content/agents/review/hive-review-ux.md` encuentra el enlace a los checks medidos.
- `rg -n "320" content/skills/flow-verify/references/ui-verification.md` encuentra el ancho de reajuste.
- `rg -n "no results|1000|detail" content/skills/flow-build/references/delegation.md` encuentra los estados nuevos del brief.

### T4 — Sitio, catálogo, changelog y atribución

- [x] El sitio y los documentos describen la skill y el comportamiento nuevo:
  - `site/src/content/docs/skills.md` e `index.mdx` tienen la entrada de la skill;
  - `flows/plan.md` describe el DESIGN.md en las decisiones de UI;
  - `flows/verify.md` describe los checks medidos;
  - `inspiration.md` acredita Impeccable (metodología, Apache-2.0) y Google DESIGN.md (formato), aprobado en D11 B;
  - `llms.txt` lista la skill;
  - `CHANGELOG.md` «Unreleased» describe el cambio;
  - `_support/docs/architecture/instruction-resources.md` agrega la casa canónica y una entrada en «Vendored upstream material» que dice que el texto es original, qué metodología se tomó y que no requiere copia de la licencia porque no se adaptó ningún pasaje.

**Closes:** AC9, AC10

**Depends on:** T1 (describe el texto de la skill; el comportamiento de T2 y T3 lo toma de `design.md`)

**Locations:** `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/inspiration.md`, `llms.txt`, `CHANGELOG.md`, `_support/docs/architecture/instruction-resources.md`

**Execution:** hijo `hive-write-docs`. El brief nombra la carpeta de cambio y los archivos de T1 como fuentes. Las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas, y no nombran proyectos de referencia fuera de los que aprueba D11 B.

**Test approach:** check

**Changes:** `README.md` no cambia, porque no lista skills ni cambia ningún comando, host ni requisito.

**Verification:**
- `go test ./...` pasa.
- Un `rg -l "hive-ui-criteria"` por archivo encuentra cada uno de: `skills.md`, `index.mdx`, `flows/plan.md`, `flows/verify.md`, `llms.txt`, `CHANGELOG.md`, `instruction-resources.md`.
- `rg -n "Impeccable|DESIGN.md" site/src/content/docs/inspiration.md` encuentra los dos créditos.
- el diff no agrega nombres de proyectos de referencia fuera de los que aprueba D11 B.
- `npm run build` en `site/` pasa.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, local | `go test ./...` y `go vet ./...` | Cubierta por la implementación |
| Catálogo | Candidato final, después del commit | `hive update --dry-run --source <worktree>` lista `hive-ui-criteria` entre lo que instalaría, sin escribir nada | Cubierta por la implementación |
| Revisión de la guía | Candidato final | Se revisan contra `AGENTS.md` «Guidance structure» los lugares de la lista de impacto de `design.md`; el PR lista los lugares revisados y el cambio de tamaño de `global.md` (cero) | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` (D8 A) | La respuesta de revisión |

No hay UI que recorrer: el cambio es guía distribuida y no cambia ninguna pantalla.

## Estado de la revisión y avance

2026-11-05 — plan escrito y revisado por un hijo `hive-review-plan` (13 hallazgos; los 10 sin decisión se aplicaron y los 3 restantes se resolvieron con D9–D11).
2026-11-05 — T1 hecho, `e8e4cf5e`.
  - Los cinco archivos pesan entre 0,4 y 5,1 KB y `go test ./tests/content/...` pasa.
  - Hay 24 marcas `[measured]`, sin prohibiciones ni puntuaciones.
  - La comprobación de copia (18 frases contra Impeccable) solo coincide en dos nombres de concepto que acredita el aviso.
  - La prueba en seco en `tooling/viewer/frontend` llega a shadcn `base-nova` con registro `@reui`, a `@base-ui/react` 1.9.0 en `pnpm-lock.yaml`, a `src/styles.css` y a `src/components/ui|reui`. Sin verificar: el caso de un rango del manifiesto distinto al instalado, porque ese proyecto fija la versión exacta; y los tipos de valor completos del `docs/spec.md` de Google, que no se leyó.
2026-11-05 — T2 hecho, `d66a3f45`. Los enlaces a la skill están en los tres archivos; ya no quedan «improving hierarchy» ni «change no token».
2026-11-05 — T3 hecho, `509ceef2`. Ya no queda «project UI guide» en `content/`; el ancho de 320 px y los estados del brief están agregados; `go test ./tests/content/...` pasa.

2026-11-05 — T4 hecho, `a0e2f069`.
  - Las siete verificaciones por archivo pasan, y los créditos de `inspiration.md` también.
  - Ningún archivo nuevo nombra los proyectos de referencia.
  - Se ajustó la redacción de `verify.md` y de `CHANGELOG.md`: una decisión registrada no es una desviación.
  - `pnpm --dir site install --frozen-lockfile` instaló las dependencias del worktree. El lockfile es del 2026-10-24 y pnpm aplica `minimumReleaseAge` de 7 días.
  - `pnpm --dir site run build` construyó 17 páginas.

**CI local:** la única workflow, `.github/workflows/master-source-branch.yml`, comprueba la rama de origen de los PR a `master`; no corre jobs locales y no aplica.

### Verify

2026-11-05:
- **Catálogo.** `hive update --dry-run --source <worktree>` sobre `202d1fab` instalaría `hive-ui-criteria` con sus cuatro referencias, sin conflictos (AC1).
- **Revisión de la guía.** Ningún detector de `tests/pilot` ni otro texto depende de las frases cambiadas, y `global.md` no cambia (0 bytes). Lugares revisados: `content/` (skills, referencias y roles), `site/`, `_support/docs/architecture/`, `tests/` e `integrations/`.
- **`/code-review` (D8 A).** Encontró 10 hallazgos; todos se confirmaron leyendo el texto del candidato y se corrigieron en una ronda:
  1. **Ámbito de la decisión pendiente.** La regla solo cubría los arreglos; ahora cubre cualquier cambio que toque un token o un componente instalado.
  2. **Requisito declarado.** La regla del requisito de accesibilidad declarado contradecía la de los valores por defecto sin modificar; se separaron con y sin requisito.
  3. **Mínimos y decisiones registradas.** Los mínimos quedan fuera de los niveles 2–5, y un componente personalizado antes y reutilizado se trata como existente.
  4. **Ancho de 320 px.** Ahora solo aplica cuando aplica el viewport móvil o hay un requisito declarado, para no hacer fallar las apps solo de escritorio.
  5. **Escritura del modo.** El modo ya no se escribe en el DESIGN.md por fuera de la regla de escritura.
  6. **Rutas.** Son relativas al directorio del DESIGN.md.
  7. **Primitivas.** La comprobación de desactualización incluye las primitivas de las librerías que copian su código.
  8. **Nombres de referencia.** La carpeta de cambio nombraba proyectos de referencia y repetía una decisión local; se quitaron, y las commits locales sin publicar se rehicieron para que el historial no los lleve.
  9. **Regla de escritura.** Tenía tres casas; ahora la canónica es «Writing» de `design-record.md`.
  10. **Brief del revisor.** Agrega cómo mantener un envío pendiente.

**Checkpoint:** verify, paso 3 terminado. Siguiente: segunda revisión limitada a los arreglos, suite completa sobre el candidato, rebase sobre `origin/development`, push y PR.
