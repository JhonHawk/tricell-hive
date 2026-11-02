# Tareas

Base `development` @ `452b4e3e`, rama `fix/gh-333-proportional-ceremony`. Todas en el hilo principal: son ediciones de texto cortas y entrelazadas, y el brief de cada una no sería más corto que el cambio.

### T1 — Reglas globales

- [ ] `global.md` declara el criterio de entrega, la tarea operativa, la regla de tickets, la excepción de tracker y la autorización de limpieza local

**Closes:** AC1, AC2, AC4, AC5, AC6

**Depends on:** none

**Locations:** `content/guidance/global.md` («Scope and authorization», «Preservation», «Proportionality», «Questions and routes», «Delivery», «Scope changes»)

**Execution:** hilo principal

**Test approach:** check

**Changes:** criterio de modo en «Delivery»; definición de tarea operativa en «Proportionality» y su exención en «Delivery»; «Scope changes» punto 4 con creación, duplicados y relación de tickets y la recomendación de corregir ahora; la frase de ticket faltante de «Questions and routes» pasa a ruta sin ticket; excepción de tracker en la viñeta del ticket; «Preservation» suma contenedores, imágenes y dumps locales. No cambia las demás reglas de «Delivery» ni la pregunta de revisión.

**Verification:** `rg -n "operational task|surface only the user|blocked-by" content/guidance/global.md` encuentra cada regla una vez; lectura completa de las secciones tocadas sin contradicciones.

### T2 — Plan en el chat y slug

- [ ] `flow-plan` decide entre plan en el chat y carpeta de cambio, y nombra la carpeta con un slug cuando no hay ticket

**Closes:** AC3, AC4

**Depends on:** none

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-plan/assets/change-folder/proposal.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** paso 3 condicionado a las condiciones de retención; plan en el chat con instantánea en el visor; `<change-id>` con la clave del ticket o un slug; handoff con la ruta para plan en el chat; título de la plantilla. No cambia los criterios de aceptación ni las decisiones de UI.

**Verification:** `go test ./tests/content/...` pasa (tamaños); `rg -n "ticket key" content/skills/flow-plan` solo muestra la forma condicional.

### T3 — Build con plan en el chat y tarea operativa

- [ ] `flow-build` ejecuta un plan del chat, salta el movimiento de ticket en una tarea operativa y la lleva al cierre

**Closes:** AC3, AC6, AC7

**Depends on:** T1 (nombres de las reglas a las que apunta)

**Locations:** `content/skills/flow-build/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** entrada con plan del chat; checkpoint en la conversación cuando no hay `tasks.md`; paso 1 apunta a la excepción de tracker; paso 5 suma la tarea operativa (sin verify ni CI, criterios comprobados y `flow-close`). No cambia la delegación ni los datos almacenados.

**Verification:** lectura de la entrada y los pasos 1 y 5 contra T1; `go test ./tests/content/...`.

### T4 — Cierre

- [ ] `flow-close` declara sus disparadores, sigue sin visor y limpia recursos locales y remotos de la tarea

**Closes:** AC5, AC7

**Depends on:** T1

**Locations:** `content/skills/flow-close/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** párrafo de disparadores (punto de finalización, tarea operativa, abandono; `hold` al entregarse) con los pasos que corre cada uno; paso 1 sin visor; paso 4 con contenedores, imágenes, servicios y dumps locales y la lista con pregunta de remotos; la recomendación por hallazgo apunta a «Scope changes». Corrige de paso la numeración del paso 3, que dice «Steps 2 and 3 are independent». No cambia la limpieza de ramas ni la pregunta de cierre.

**Verification:** `go test ./tests/content/...` (tamaño ≤ 16 000 bytes); lectura completa del archivo.

### T5 — Sitio

- [ ] Las páginas Plan, Build y Close describen el comportamiento nuevo

**Closes:** AC8

**Depends on:** T2, T3, T4

**Locations:** `site/src/content/docs/flows/plan.md`, `build.md`, `close.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** plan en el chat o carpeta; tarea operativa; disparadores y limpieza del cierre. Describen y enlazan, no copian reglas.

**Verification:** build del sitio si su CI lo corre localmente (`pnpm --dir site build`), o la comprobación de enlaces que use el repo.

### T6 — Revisión de comportamiento completo

- [ ] Ningún lugar que declara o depende de estas reglas queda en conflicto

**Closes:** AC8

**Depends on:** T1–T5

**Locations:** `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`

**Execution:** hilo principal

**Test approach:** check

**Changes:** búsquedas de «ticket key», «change folder», «leaves nothing to commit», «Separate topic», «completion point», «automatic» en todas las ubicaciones; corrección de cada copia afectada; lista de lugares revisados y cambio de tamaño de `global.md` para el pull request.

**Verification:** `go test ./...` desde la raíz (o los paquetes de `tests/` y `tooling/` que corre CI) pasa; la lista queda en esta tarea.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, local | `go test ./tests/content/...` pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | Jobs locales del CI del repo | Cubierto por la implementación |
| Revisión de código | Candidato final, antes de la validación o del push | <según la respuesta de revisión> | La respuesta de revisión |

<Recorrido humano: leer el diff de `content/guidance/global.md` y de los tres skills.>

## Estado de revisión y progreso

2026-11-02 — Plan escrito.
- T1 hecho · `11afb0a8` · `global.md` 24 059 → 26 101 bytes (+2 042).
- T2 hecho · `9c7ef582`.
- T3 hecho · `7ff66870`.
- T4 hecho · `a5539416` · la frase del paso 3 «Steps 2 and 3 are independent» no es un error de numeración (declara que los pasos 2 y 3 corren juntos): se dejó igual.
- T5 hecho · `2a20cf3e` · `astro build` 17 páginas OK.
- T6 hecho · `c3ab66ce` · la retención suma dinero y concurrencia porque `hive-verify-task` lee la carpeta; checkpoint de `flow-verify` y ruta de `research.md` alineados. Lugares revisados: `content/guidance`, `content/skills/*` (flow-plan, flow-build, flow-close, flow-verify, flow-research, flow-deploy, git-workflow, unattended-delegation, flow-walkthrough), `content/agents`, `site/src/content/docs` (todas las páginas), `README.md`, `llms.txt`, `_support/docs/architecture/`, `tests/pilot/`, `integrations/`. `go test ./tests/... ./tooling/...` OK.

Checkpoint: build · paso 5 hecho · último commit `c3ab66ce` · sin hallazgos abiertos · siguiente: `flow-verify` (revisión `/code-review`, luego parada de validación).
