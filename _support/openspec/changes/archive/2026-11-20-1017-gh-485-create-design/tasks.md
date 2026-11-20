# Tasks

Base: `development` en `34e58d40` (2026-11-20). Rama: `feat/create-design`, en un worktree bajo `.claude/worktrees/create-design`.

Hilo principal: T1 (mover archivos, mecánico), T4 (dos archivos de prueba) y T6 (verificación). Hijos `hive-write-docs` por costo, porque cada tarea edita más de tres archivos con la interfaz ya fijada en `design.md`: T2, y luego T3 y T5 en paralelo, porque escriben archivos disjuntos y ambas dependen solo de T2. T4 corre después de T3.

### T1 — Mover y renombrar las carpetas

- [x] `content/skills/create-design/` existe con el contenido de `flow-figma-design` y las cinco referencias de `hive-ui-criteria`; las dos carpetas viejas ya no existen.

**Closes:** AC1

**Depends on:** none

**Locations:** `content/skills/flow-figma-design/` → `content/skills/create-design/`; `content/skills/hive-ui-criteria/SKILL.md` → `content/skills/create-design/references/ui-criteria.md`; `content/skills/hive-ui-criteria/references/{criteria,design-record,refinement,notice}.md` → `content/skills/create-design/references/`

**Execution:** main thread

**Changes:** `git mv` de cada archivo. En los archivos movidos, el texto de los enlaces que nombra las skills viejas (como `[flow-figma-design](../SKILL.md)`) y los enlaces relativos internos que el movimiento rompe (por ejemplo `references/criteria.md` → `criteria.md` dentro de `ui-criteria.md`, y `../../hive-ui-criteria/references/design-record.md` → `design-record.md` en `design-library.md`), el `name:` del frontmatter, y en `notice.md` el nombre de la skill. `ui-criteria.md` pierde el frontmatter de skill y conserva su texto. Sin cambios de reglas.

**Verification:** `go test ./tests/content -run Size` pasa; `rg -n "hive-ui-criteria|flow-figma-design" content/skills/create-design` sin resultados.

### T2 — Entradas, revisión, ancho móvil y entrega de DESIGN.md en create-design

- [x] `create-design/SKILL.md` abre con las cuatro entradas de `design.md`; existe `references/review.md`; el paso 5, `approval.md`, `rule-cards.md` y `sections.md` aplican el ancho de D1; `design-library.md` «Record it» y `special-cases.md` aplican D4.

**Closes:** AC3, AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/create-design/SKILL.md`, `references/review.md` (nuevo), `references/approval.md`, `references/rule-cards.md`, `references/sections.md`, `references/design-library.md`, `references/special-cases.md`, `references/ui-criteria.md` (solo la línea que dice cuándo cargar cada referencia, si nombra la skill vieja), `references/design-record.md` (campos **Design file:** y **Design library:** y la excepción de «Writing»)

**Execution:** child `hive-write-docs`

**Test approach:** check

**Changes:** Descripción del frontmatter con las cuatro entradas, incluida «crear o actualizar `DESIGN.md`», a 1.024 caracteres o menos. Tabla de entradas al inicio. `review.md` según `design.md` «Entrada de revisión». Paso 5: app web al ancho de escritorio de sus frames existentes o 1440 px, más 440 px si se sirve en móvil; app solo móvil una vez a 440 px, sin segundo frame. Las menciones de «its 440 px frame» en `approval.md`, `rule-cards.md` y `sections.md` quedan condicionadas a la app web. `design-library.md` «Record it»: un sí queda `pending` siempre, sin escribir, sin pregunta de entrega y sin git-workflow; se borran los pasos 1–4 de «On a yes». La misma regla en los demás lugares que dicen que la etapa escribe esos campos: `SKILL.md` línea 8 («in the repository it edits only the two `DESIGN.md` fields»), `approval.md` línea 15 («write the file and the library in `DESIGN.md`»), y en `design-record.md` las líneas de los dos campos y la excepción de «Writing», que pasa a decir que el plan los escribe como cualquier otro campo y que `design-library.md` es el hogar de cómo se preguntan. `special-cases.md`: el pedido solo de diseño nombra pendiente el archivo o la biblioteca. Paso 5: la señal de plataforma de las decisiones (nativa o shell de teléfono sin escritorio = solo móvil; breakpoint o meta viewport = web servida en móvil; sin código ni frames, la plataforma del pedido o del ticket; sin ninguna, web con 440 px, dicho en el mensaje de aprobación). No cambia el resto del procedimiento de dibujo ni los criterios. `SKILL.md` queda por debajo de 16.000 bytes, y cada referencia también.

**Verification:** `go test ./tests/content -run Size` pasa; `wc -c content/skills/create-design/SKILL.md` < 16000; `rg -n "writes these two fields|edits only the two|write the file and the library|writes and delivers" content/skills/create-design` sin resultados; lectura de cada texto cambiado contra AC3, AC4 y AC5.

### T3 — Enlaces entrantes en roles, hojas, playbooks y skills

- [x] Cada enlace a `flow-figma-design` o `hive-ui-criteria` en `content/` apunta a la ruta de `create-design` que existe; `hive-mode` envía un pedido de revisar un diseño a la entrada de revisión; `hive-review-ux` cubre una imagen sin nodos; `hive-rule-ui-example` y `hive-rule-ui-plan` agregan la tarea de un `pending` también con sección `## Hive`; `hive-rule-reviewer-brief` tiene el brief de un diseño que dibujó otra persona; revisar una imagen no pasa por el chequeo de conexión.

**Closes:** AC2, AC3, AC5

**Depends on:** T2

**Locations:** `content/agents/review/hive-review-ux.md`, `content/agents/development/hive-build-frontend.md`, `content/agents/design/hive-design-ui.md`, `content/skills/hive-rule-{design-stage,reviewer-brief,ui-example,ui-plan,ui-verification,business-record,business-rules,business-ticket,checkpoint,delegation}/SKILL.md`, `content/skills/hive-mode/SKILL.md`, `content/skills/hive-mode/playbooks/{deliverable,ui-change}.md`, `content/skills/create-prd/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/unattended-delegation/SKILL.md`

**Execution:** child `hive-write-docs`

**Test approach:** check

**Changes:** Rutas y nombres; los roles usan `skill:create-design/references/...`. Además, solo estas reglas: `deliverable.md` y `hive-mode/SKILL.md` agregan la revisión de un diseño existente como entrada de `create-design`, sin exigir conexión cuando es una imagen; `hive-review-ux` cubre una imagen sin nodos; `hive-rule-reviewer-brief` agrega, en su único hogar, el brief de un diseño de otra persona según el supuesto de la propuesta (quedan unos 600 bytes bajo el límite); `hive-rule-ui-example` y `hive-rule-ui-plan` cubren un `pending` con o sin sección `## Hive`. Ninguna hoja supera 4.000 bytes ni su descripción 200 caracteres.

**Verification:** `go test ./tooling/management -run Catalogue` pasa (resuelve los enlaces); `go test ./tests/content -run 'Size|Leaf'` pasa salvo las filas de `leaf_links_test.go` que T4 corrige; `rg -n "flow-figma-design|hive-ui-criteria" content` sin resultados.

### T4 — Pruebas que nombran las skills

- [x] `tests/content/leaf_links_test.go` exige los enlaces en `create-design`; `tests/pilot/session_audit_routes.go` reconoce `create-design` y también `flow-figma-design` para sesiones anteriores al cambio, en `flowCommand`, `flowSkillRead` y la comparación de nombres.

**Closes:** AC2

**Depends on:** T3

**Locations:** `tests/content/leaf_links_test.go`, `tests/pilot/session_audit_routes.go`

**Execution:** main thread

**Test approach:** check

**Changes:** Ruta del dueño en las filas de `leaf_links_test.go`; `flowCommand` (línea 21, que hoy exige el prefijo `flow-`), `flowSkillRead` y la comparación de la línea 71 aceptan los dos nombres. Ningún fixture nuevo.

**Verification:** `go test ./tests/content/... ./tests/pilot/...` pasa.

### T5 — Sitio, README, llms.txt, arquitectura y CHANGELOG

- [x] El sitio tiene `flows/create-design/` con las cuatro entradas, el ancho móvil y la entrega por el plan, y `flows/figma-design/` redirige ahí; ninguna página, `README.md`, `llms.txt` ni `instruction-resources.md` nombra las skills viejas; `CHANGELOG.md` tiene la línea del cambio en `[Unreleased]`.

**Closes:** AC2

**Depends on:** T2

**Locations:** `site/src/content/docs/flows/figma-design.md` → `flows/create-design.md`, `site/src/content/docs/flows/{overview,plan,verify,build,close,create-prd,hive-mode}.md`, `site/src/content/docs/getting-started.md`, `site/src/content/docs/{index.mdx,skills.md}`, `site/astro.config.mjs` (barra lateral y redirección), `README.md`, `llms.txt`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** child `hive-write-docs`

**Test approach:** check

**Changes:** Las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas; `hive-mode.md` describe la nueva ruta de revisión. Los enlaces a `/flows/figma-design/` y el nombre «Figma design» pasan a `create-design`. Las entradas viejas de `CHANGELOG.md` no se tocan.

**Verification:** en el worktree, enlazar `site/node_modules` del checkout principal, correr `npm run build` en `site/` (como registra `HIVE.local.md`) y quitar el enlace; el build termina sin errores y genera `flows/create-design/index.html` y la redirección de `flows/figma-design/`; `rg -n "flow-figma-design|hive-ui-criteria|figma-design|Figma design" site/src site/astro.config.mjs README.md llms.txt _support/docs` solo muestra la redirección.

### T6 — Verificación del conjunto

- [x] Suite completa en verde, sin nombres viejos fuera del histórico, y las deltas de esta carpeta coinciden con el texto final.

**Closes:** AC1, AC2, AC6

**Depends on:** T3, T4, T5

**Locations:** ninguna nueva; si la redacción final de T2 se aparta de los deltas, `specs/` de esta carpeta

**Execution:** main thread

**Test approach:** check

**Changes:** Si la redacción final se aparta de un delta, no se edita en silencio: se revisa el plan como fija `hive-rule-change-folder` «Revise a plan in progress» (registro, nueva revisión de `hive-review-plan`). El PR lista los lugares revisados y el tamaño de `content/guidance/global.md`, y el commit nombra las observaciones de kin (Codex, `gpt-6.2-sol`, sesiones `01a1e3e1` del 2026-11-16 y `01a1f7d9` del 2026-11-20), como pide `AGENTS.md`.

**Verification:** `go vet ./...` y `go test ./...` pasan; `rg -l "flow-figma-design|hive-ui-criteria" --glob '!_support/openspec/changes/**' --glob '!_support/workspace/**' --glob '!_support/sessions/**' --glob '!CHANGELOG.md' --glob '!_support/openspec/specs/**' --glob '!tests/pilot/session_audit_routes.go'` sin resultados; `python3 <flow-plan>/scripts/check_change.py` sale con 0. AC1 se observa tras el merge, en el refresco local (`hive update --source <checkout>`), con `ls ~/.agents/skills`. AC6 se observa en el cierre, que aplica los deltas.

## Verification and human review

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, local | `go vet ./...`, `go test ./...`, build del sitio como en T5 | Cubierta por la implementación |
| UI | No aplica: no cambia ninguna pantalla de un producto | — | — |
| Revisión de código | Candidato final, antes del push o de la parada de validación | Según la respuesta de revisión | La respuesta de revisión |

Recorrido humano: leer `content/skills/create-design/SKILL.md` y `references/review.md`, y abrir `flows/create-design/` en el sitio construido localmente (`pnpm --dir site preview`).

## Checkpoint

- Updated: 2026-11-20
- Stage: flow-verify · second review (fix diff)
- Last commit: e152d372
- Open findings: H-nit: readsFlowSkill counts any create-design load as a flow start (review/DESIGN.md entries) · close findings
- Next: judge second review, push, PR, merge, refresh install, flow-close
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/create-design · feat/create-design

### Evidence

- before captures: none (no product screen)
- go vet ./... && go test ./... · pass · 18690a94
- site npm run build · pass (20 pages, redirect) · d02ae2c8
- rg old names outside history · none · 18690a94
- check_change.py · exit 0 · 18690a94
- CI workflows: only master-source-branch.yml (PR source guard), not applicable
- go vet ./... && go test ./... · pass · e152d372
- site npm run build · pass · e152d372

## Review status and progress

2026-11-20 — T1 · a7d54683 · unverified: AC1 se observa tras el refresco local
2026-11-20 — T2 · 644f2aa2 · unverified: AC3–AC5 se observan en los archivos desplegados tras el refresco
2026-11-20 — T3 · 61493f78 · unverified: AC2 y AC3 en archivos desplegados, tras el refresco
2026-11-20 — T5 · d02ae2c8 · unverified: también corrigió compaction-survival.md:69, fuera de sus Locations
2026-11-20 — T4 · 18690a94 · unverified: none
2026-11-20 — T6 · uncommitted · unverified: AC1 tras el refresco local; AC6 en el cierre
