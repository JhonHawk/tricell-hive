# Tasks

Base `development` @ `98266174`; rama `feat/figma-design-rule-questions` en `.claude/worktrees/`. Todas las tareas corren en el hilo principal: cada una edita uno a cuatro archivos de texto, sin lecturas grandes por delante, bajo el umbral de delegación.

### T1 — Reglas en `flow-figma-design`

- [x] El skill define la regla, sus fuentes y su línea de registro; las pregunta aparte, declara las menores como supuesto y manda a investigación las que cambian el alcance

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `content/skills/flow-figma-design/SKILL.md`

**Execution:** main thread

**Changes:**
- **Paso 2:** junta las fuentes que fijan reglas:
  - la petición y los criterios del ticket;
  - las decisiones de investigación como las tengan la conversación, el ticket o el registro del trabajo;
  - las decisiones registradas de la app.

  El paso 3 suma lo que ya implementan las pantallas hermanas o su código.
- **Paso 5, al dibujar:**
  - Lista cada regla que un frame muestra y ninguna fuente fija: qué exige una acción, quién la toma, qué estado sigue, qué bloquea o conserva, dónde vive un ajuste que aplica más allá de la pantalla.
  - Dibuja la recomendada y la escribe en `examples.md` como una línea `rule: <regla> · screen <pantalla> · proposed | approved <fecha>: <respuesta> · found by drawer | reviewer`, separada de las líneas de estado de pantalla.
  - Una regla cuyas opciones no cambian alcance, contrato, datos guardados, costo, reversibilidad ni quién puede actuar queda como supuesto (`assumption`) y no se pregunta.
  - Una regla que añade una capacidad no pedida, cambia datos o ajustes que usan otras pantallas o módulos, o cambia un permiso, detiene el dibujo de las pantallas que moldea.
- **Paso 5, tras la revisión:** une las reglas sin fuente que reporta el revisor con las suyas como `proposed`, aplicándoles la misma prueba de alcance. Las nombra aparte de los hallazgos que quedan.
- **Paso 6:**
  - Una decisión por regla propuesta en la misma ronda que la aprobación.
  - La opción de aprobar la pantalla nombra esa decisión y vale con cualquier respuesta (`aprobar la composición, con la regla que elijas en D<n>`).
  - Aprobar no aprueba la regla.
  - Una respuesta distinta de la dibujada redibuja esa parte, y la pantalla queda pendiente hasta que el usuario la apruebe redibujada.
  - Los supuestos se listan en el mensaje.
- **Paso 7:** escribe la respuesta en cada línea `rule:`.
- **Paso 8:**
  - Con reglas de alcance, primero se piden las pantallas no afectadas y luego la etapa termina con el bloque hacia `flow-research` que nombra esas reglas.
  - Al volver, la etapa retoma el registro y dibuja solo las pantallas detenidas.
  - La salida a build exige además que ninguna regla siga `proposed`.
- **Pedido de diseño:** una regla de alcance deja el trabajo sin terminar y termina igual.
- **Desatendido:** las reglas quedan propuestas y se reportan sin resolver, y una regla de alcance igual detiene su pantalla.

**No cambia:** el dibujo por script, la revisión ni la decisión por pantalla.

**Verification:** `rg -n "rule:|assumption|flow-research" content/skills/flow-figma-design/SKILL.md` muestra los pasos 2, 5, 6, 7 y 8 y la sección desatendida. `go test ./tests/content/...` pasa (tamaño ≤ 16 000 bytes).

### T2 — El revisor de frames busca reglas sin fuente

- [x] El modo de frames reporta reglas sin fuente aparte de los hallazgos, y el brief lista sus fuentes como requisitos

**Closes:** AC4

**Depends on:** none

**Locations:** `content/agents/review/hive-review-ux.md`, `content/skills/flow-build/references/delegation.md`

**Execution:** main thread

**Changes:**
- **`delegation.md`:**
  - La lista de requisitos (línea 38) incluye las decisiones de investigación.
  - El caso de frames (línea 41) indica que el brief no incluye las reglas propuestas de quien dibujó, para que el revisor las busque por su cuenta.
- **`hive-review-ux.md`:**
  - El modo de frames reporta cada comportamiento que no fijan sus requisitos ni las pantallas hermanas, en una lista «unsettled rules» aparte que no reprueba el veredicto.
  - No cambia el modo de navegador.

**Verification:** `rg -n -i "unsettled rule|research's decisions" content/agents/review/hive-review-ux.md content/skills/flow-build/references/delegation.md` muestra el texto nuevo; `go test ./tests/...` pasa, incluido `agent_boundary_test.go`.

### T3 — El plan toma las reglas del registro

- [x] `flow-plan` escribe las reglas aprobadas como decisiones y pregunta las propuestas

**Closes:** AC5

**Depends on:** T1

**Locations:** `content/skills/flow-plan/SKILL.md`

**Execution:** main thread

**Changes:** en «UI decisions», paso 4, al leer `examples.md`:
- cada línea `rule:` aprobada entra a las decisiones del plan y al `UI` junto al registro del ejemplo;
- cada línea `proposed` se pregunta como decisión antes de aprobar el plan;
- bajo delegación desatendida, una regla propuesta que cambia el alcance queda sin resolver y se reporta.

**Verification:** `rg -n "rule:" content/skills/flow-plan/SKILL.md` muestra el paso 4; `go test ./tests/content/...` pasa.

### T4 — Guía global, especificación, sitio y registro

- [x] Transición nueva en la guía global, delta aplicada, páginas del sitio, changelog e inventario

**Closes:** AC6

**Depends on:** T1, T2, T3

**Locations:**
- `content/guidance/global.md`
- `_support/openspec/changes/figma-design-rule-questions/specs/workflow-lifecycle/spec.md`
- `site/src/content/docs/flows/figma-design.md`
- `site/src/content/docs/flows/plan.md`
- `CHANGELOG.md`
- `_support/docs/architecture/instruction-resources.md`

**Execution:** main thread

**Changes:**
- **`global.md` línea 90:** «Figma design to plan or build» pasa a «Figma design to plan, build or research». El cambio de tamaño se reporta en el PR.
- **Delta:** ya está escrita (ADDED + MODIFIED); se revisa contra el texto final de T1–T3.
- **`figma-design.md`:**
  - «What it asks you» suma las decisiones de regla y los supuestos.
  - «How it ends» suma la vuelta a Research y la condición de build.
- **`plan.md` línea 26:** suma las reglas aprobadas y propuestas.
- **`CHANGELOG.md`:** entrada en `[Unreleased]`.
- **`instruction-resources.md` línea 19:** suma las preguntas de regla a lo que el skill posee.

README y `llms.txt` no cambian: nada de lo que dicen cambia.

**Verification:**
- `python3 -I content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/figma-design-rule-questions` sale con 0.
- `pnpm build` en `site/` pasa.
- `go test ./...` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y `pnpm build` en `site/` | Covered by the implementation |
| Coherence | Final candidate | búsqueda de cada regla tocada en `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/` | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | D5 A |

Sin validación humana: entrega `automatic` (D4 B).

## Checkpoint

- Updated: 2026-11-11
- Stage: verify · 5 · integrate
- Last commit: e4114cb3
- Open findings: none
- Next: push, PR, merge (automatic), then flow-close

### Evidence

- go test ./... · pass · 6f3b8ffe
- pnpm build (site) · pass, 18 pages · 6f3b8ffe
- check_change.py · exit 0 · 6f3b8ffe
- apply_deltas.py dry run · MODIFIED + ADDED, no refusal · 6f3b8ffe
- go test ./tests/... · pass · e4114cb3
- pnpm build (site) · pass · e4114cb3
- check_change.py · exit 0 · e4114cb3
- apply_deltas.py dry run · MODIFIED + ADDED, no refusal · e4114cb3
- /code-review round 1 · 10 findings fixed in d85e7c07
- /code-review round 2 (fix diff) · 10 findings; blocking ones fixed in e4114cb3

## Review status and progress

- 2026-11-11 — Plan escrito.
- 2026-11-11 — Revisión del plan aplicada; ver proposal.md.
2026-11-11 — T1 · 06d6d060 · unverified: none
2026-11-11 — T2 · 32e1fdc6 · unverified: none
2026-11-11 — T3 · ad08c1ae · unverified: none
2026-11-11 — T4 · 6f3b8ffe · unverified: none
