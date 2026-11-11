# Tasks

Base: `development` @ `ea4195ab`; rama `feat/figma-design-main-thread-review` en un worktree bajo `.claude/worktrees/`. Todas las tareas corren en el hilo principal: son ediciones de texto enlazadas, ninguna supera el umbral de costo de la delegación y comparten la misma regla, así que repartirlas rompería su coherencia. Antes de editar, leer `_support/docs/architecture/instruction-resources.md`, como exige `AGENTS.md` para skills y roles.

### T1 — `flow-figma-design`: un ejemplo, hilo principal por script, revisión y convenciones

- [ ] La skill dibuja un ejemplo por pantalla siguiendo sus hermanas, alternativas solo para una pantalla nombrada, dibujo en el hilo principal por script, registro al dibujar con estado en revisión, revisión de un `hive-review-ux` antes de preguntar, y la pregunta de convención en la ronda de aprobación

**Closes:** AC1, AC2, AC3, AC5

**Depends on:** none

**Locations:** `content/skills/flow-figma-design/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** descripción ("draw a screen", no "proposals"); paso 2 agrega leer las pantallas hermanas (archivo y código) y las decisiones de **Decisions** en `DESIGN.md` y de las secciones `UI` de las carpetas de cambio; paso 5 reemplaza "two or three examples when the composition is open" por un ejemplo por pantalla y alternativas solo para la pantalla que el usuario nombra, reemplaza el párrafo del subagente por el método por script en el hilo principal, escribe el registro al dibujar con `drawn, under review` y agrega la revisión, que apunta a «Reviewer brief», con una ronda de corrección; paso 6 agrega la pregunta única de convención y la revisión del frame de 440 px; paso 7 guarda la respuesta; paso 8 y «A request for the design itself» nombran la convención junto al archivo de diseño. No cambia: archivo, ancho, desatendido.

**Verification:** releer los pasos 2 y 5 a 8 contra AC1, AC2, AC3 y AC5; `rg -n 'composition is open|hive-design-ui' content/skills/flow-figma-design` no devuelve nada; `rg -n 'two or three' content/skills/flow-figma-design` devuelve solo la regla de alternativas a petición; `go test ./tests/content/...` pasa (tamaño ≤ 16 000 bytes).

### T2 — Delegación y roles

- [ ] La regla de delegación deja el dibujo de Figma en el hilo principal y define el brief del revisor de frames; `hive-review-ux` describe el modo de frames; `hive-design-ui` deja de describir páginas de Figma

**Closes:** AC2, AC3

**Depends on:** T1

**Locations:** `content/skills/flow-build/references/delegation.md`, `content/agents/review/hive-review-ux.md`, `content/agents/design/hive-design-ui.md`

**Execution:** main thread

**Test approach:** check

**Changes:** quitar la viñeta de costo de Figma y agregar que el dibujo de esa etapa queda en el hilo principal sin importar su tamaño; en «Reviewer brief», el caso de frames de un archivo de diseño (nodos e imágenes, petición, nodos de hermanas, decisiones registradas, ancho, sin estados que alcanzar). En el revisor, actualizar la descripción y agregar un párrafo que selecciona el brief: qué de los pasos 1 a 3 no aplica, el veredicto para frames, el respaldo por imágenes sin conexión y la relectura solo de lo corregido; nunca cambia los frames. En `hive-design-ui`, quitar las oraciones sobre la página de Figma y la evidencia de frames. No cambia el uso de `hive-design-ui` en `flow-plan` ni el de `hive-review-ux` en `flow-verify`.

**Verification:** `rg -n 'Figma' content/agents/design/hive-design-ui.md` no devuelve nada; `rg -n 'design file' content/agents/review/hive-review-ux.md content/skills/flow-build/references/delegation.md` devuelve el párrafo y el brief nuevos; `go test ./tests/content/...` pasa, incluido `agent_boundary_test.go`.

### T3 — Research, plan y registro de diseño

- [ ] Research recomienda la etapa para toda pantalla nueva con conexión, también cuando iría directo a build; el plan agrega la tarea que escribe una convención con sí; **Decisions** admite convenciones de la app

**Closes:** AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/flow-research/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/hive-ui-criteria/references/design-record.md`

**Execution:** main thread

**Test approach:** check

**Changes:** en el paso 5 de research, una pantalla nueva con conexión (sin delegación desatendida) termina con el bloque que recomienda la etapa aunque el cambio fuera a continuar en build, con `flow-plan` o `flow-build` como otras respuestas; una pantalla existente que agrega, quita o mueve una región sigue con la condición de que nada fije su composición; se mantienen las exclusiones. En «UI decisions» paso 4 del plan, agregar la tarea de cada convención con sí. En el registro de diseño, **Decisions** pasa a decir desviaciones aprobadas y convenciones de la app, con fecha y origen (la carpeta de cambio que copió el registro).

**Verification:** releer contra AC4 y AC5; `go test ./tests/content/...` pasa.

### T4 — Sitio, README, llms.txt, CHANGELOG y arquitectura

- [ ] Las páginas y documentos describen un ejemplo, el dibujo por script, la revisión, las alternativas a petición, la recomendación para pantallas nuevas y la pregunta de convención

**Closes:** AC6

**Depends on:** T1, T2, T3

**Locations:** `site/src/content/docs/flows/figma-design.md`, `site/src/content/docs/flows/research.md`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/overview.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/skills.md`, `site/src/content/docs/getting-started.md` (si su frase cambia), `README.md`, `llms.txt`, `CHANGELOG.md` (`[Unreleased]`), `_support/docs/architecture/instruction-resources.md` (si su resumen cambia)

**Execution:** main thread

**Test approach:** check

**Changes:** describir el comportamiento y enlazar la skill, sin copiar sus reglas, como piden `AGENTS.md` y `CONTRIBUTING.md`.

**Verification:** `rg -n -i 'two or three|composition open|draw proposals' site/src/content/docs README.md llms.txt` devuelve solo la mención de alternativas a petición; `pnpm --dir site check`, `pnpm --dir site lint` y `pnpm --dir site build` terminan sin errores.

### T5 — Especificaciones y verificación final

- [ ] Los deltas de `workflow-lifecycle` y `ui-criteria` reflejan lo construido; la suite pasa

**Closes:** AC6

**Depends on:** T1, T2, T3, T4

**Locations:** `_support/openspec/changes/figma-design-main-thread-review/specs/`

**Execution:** main thread

**Test approach:** check

**Changes:** ajustar los deltas a la redacción final de las skills; el archivo los fusiona al cierre.

**Verification:** `python3 -I ~/.claude/skills/flow-plan/scripts/check_change.py _support/openspec/changes/figma-design-main-thread-review` sale con 0; `go test ./...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...`; `pnpm --dir site check`, `lint` y `build` | Covered by the implementation |
| UI check | No aplica: el cambio no toca una pantalla servida | — | — |
| Code review | Final candidate, before the push | `/code-review` de Claude Code | D5 A |
| Copia instalada | Tras la integración, con el refresco del mantenedor | `hive update` y lectura de la skill instalada | `AGENTS.md` «Refreshing a local installation» |

El PR lista los lugares revisados y declara que el tamaño de `content/guidance/global.md` no cambia. El commit nombra las observaciones: Codex `01a1c819` (`gpt-6-astra`) y `01a1c8e0` (`gpt-6.1-sol`), fleetsystem, 2026-11-10, y la aprobación del usuario para adoptar la corrección.

## Checkpoint

Pendiente.

## Review status and progress

2026-11-10 — Plan escrito.
2026-11-10 — Revisión del plan aplicada; ver la propuesta.
