# Tareas · ux-plan-review-and-admission-scope

Base del cambio: `f08f015c` (confirmada al arrancar el build el 2026-10-24, rama `feat/ux-plan-review-admission-scope`).

## Diseño en vigor

- **Regla de admisión (AC1, D4-A, D6-A).** El segundo párrafo de Measurement en `AGENTS.md` (línea 81 en la base) pasa a este texto, que conserva la vigilancia de un caso único y la excepción del texto contradictorio:

  > A correction in distributed guidance, skills, or roles, meaning a new or stricter rule for a situation an existing rule already addresses, written because that rule was not followed or falls short in a real session, needs that failure observed in two independent real sessions, or the user's approval to adopt it at once. Record a single observed case as a watch item, in an issue or the maintainer's memory, with its count and the occurrence that would justify the rule. A new feature or requirement, including one taken from an external reference, and a gap, a situation no rule addresses, need no recurrence, even when a real session revealed the gap: the user's decision adopts them, and a plausible improvement alone does not. Fixing text that contradicts its own intent or the source it describes needs no recurrence.

  El criterio de desempate (corrección si ya existe una regla para la situación; hueco si ninguna la cubría, aunque se haya observado en una sesión) viene de la revisión del plan y aplica la lectura confirmada en D4-A. El tercer párrafo de Measurement (nombrar la sesión real en el commit) no cambia: el commit de T2–T3 nombra la sesión que reveló el hueco: Claude Code, sesión `94440d0a`, 2026-10-24, carvento COM-645, la tarjeta de MaxiPublica volvía a pedir marca y modelo que la unidad ya guarda y ningún revisor del plan lo detectó.

- **Registro de datos en la vista (AC2, D5-A).** En `ui-planning.md`, «Record the view decisions», se añade tras el punto de componentes:

  > - For a view where the end user enters or chooses data, each value it asks for: where it already lives in the system or can be derived from, or why the end user must provide it, and how many times the end user provides it per record or repeated item. Prefer showing or confirming a value the system already holds over asking for it again, and record the reason when the view asks again.

- **Condición de lectura de UI planning (AC5).** Un plan que solo añade un diálogo o un campo de formulario debe leer `ui-planning.md`, porque el área «User experience» usa su registro:
  - `ui-planning.md:3`: «Read when creating or changing a screen, application shell, layout, or reusable UI pattern.» pasa a «Read when creating or changing a screen, application shell, layout, reusable UI pattern, or a dialog or form where the end user enters or chooses data.»
  - `flow-plan/SKILL.md:28`: «For new or changed screens, shells, layouts, or reusable UI patterns, read [UI planning]…» pasa a «For new or changed screens, shells, layouts, reusable UI patterns, or a dialog or form where the end user enters or chooses data, read [UI planning]…». La skill queda muy por debajo de 16.000 bytes (13.790 en la base).
  - `flow-build/SKILL.md:24` y `hive-build-frontend.md:15` no cambian: en build, el plan ya trae el registro.

- **Área «User experience» en la revisión del plan (AC3, D5-A).** En `plan-review.md`:
  - Fila nueva en la tabla, después de Frontend / UI:

    > | User experience | Walk the end user's real task through each planned view, using the plan's record of the values each view asks for; a view without that record is a blocking gap. Check values the system already holds or can derive, against the data model; entries repeated per record that could be asked once; steps and decisions per repeated item; and recovery from errors. When a premise or a past decision of the person who requested the plan makes the end user enter data twice or adds steps, report it as a decision for that person, not as a fix. Read [UI planning](ui-planning.md). |

  - Tras la frase «Skip absent domains: a backend-only plan needs no UI reviewer.» se añade:

    > Assign User experience to every plan that creates or changes a screen, dialog, or form where the end user enters or chooses data, or that changes the steps of a task the end user performs in the product's interface, as its own `hive-review-plan` assignment apart from Frontend / UI, even when one reviewer could cover both: a reviewer checking pattern reuse can read a repeated entry as a pattern to reuse.

- **Sitio (AC4).** En `site/src/content/docs/flows/plan.md`, la frase «independent reviewer agents read it before the agent calls it ready, one per affected area.» se completa con: «A plan with a screen where you enter or choose data also gets a reviewer that walks your task looking for data you would have to enter twice or steps you could skip.»

- **Changelog.** `CHANGELOG.md`, `[Unreleased]`: una entrada en `Added` para el área «User experience» y el registro de datos, y una en `Changed` para el alcance de la regla de admisión (repositorio, no distribuido).

- **Lugares revisados** (para el PR, `AGENTS.md`, «Guidance structure»): `content/guidance/global.md` (sin menciones de la regla de admisión ni de áreas de revisión; tamaño sin cambio, 33.773 bytes), `content/skills/flow-plan/SKILL.md` («omit unaffected areas, including UI when absent»: compatible; línea 28 cambia por AC5), `content/skills/flow-build/SKILL.md:24` y `content/agents/development/hive-build-frontend.md:15` (misma condición de lectura en build, sin cambio), `integrations/opencode/context-loader.js:6` y `tests/fixtures/flash-class-guidance/cases.json:14` (referencian `plan-review.md`, sin cambio), `content/agents/review/hive-review-plan.md` (sin áreas propias), `content/agents/review/hive-review-ux.md` (interfaz renderizada, no planes), `content/skills/flow-walkthrough/references/ui-gate.md` (gates del candidato construido, sin cambio), `site/src/content/docs/flows/plan.md`, `_support/docs/architecture/agent-delivery.md` y `instruction-resources.md` (solo enlazan `plan-review.md`), `tests/pilot/` (usan `plan-review.md` como recurso con contenido sintético), `integrations/opencode/context_snapshot.go` (genera el snapshot de `plan-review.md` al desplegar; sin archivo versionado que regenerar), `CONTRIBUTING.md`, `README.md` y `llms.txt` (no describen la regla de admisión ni las áreas de revisión).

## Tareas

### T1 — Alcance de la regla de admisión

- [x] El segundo párrafo de Measurement en `AGENTS.md` dice el texto del diseño.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `AGENTS.md:81`.

**Execution:** hilo principal: el cambio es un párrafo y el brief sería más largo que el cambio.

**Test approach:** check — `rg` sobre el texto nuevo.

**Changes:** reemplazar el párrafo por el del diseño.

**Verification:** `rg -c "gap, a situation no rule addresses" AGENTS.md` devuelve 1; `rg -c "An idea from an external reference meets the same bar" AGENTS.md` no encuentra nada; `rg -c "Record a single observed case as a watch item" AGENTS.md` devuelve 1.

### T2 — Registro de los datos que pide una vista

- [x] `ui-planning.md` pide el registro de cada dato de una vista donde el usuario final captura o elige datos, y `ui-planning.md` y `flow-plan/SKILL.md` se leen también para diálogos y formularios de captura.

**Closes:** AC2, AC5.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-plan/references/ui-planning.md` (línea 3 y «Record the view decisions»); `content/skills/flow-plan/SKILL.md:28`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check — `rg`.

**Changes:** añadir el punto del diseño después del de componentes y ampliar las dos condiciones de lectura.

**Verification:** `rg -c -i "already lives" content/skills/flow-plan/references/ui-planning.md` devuelve 1; `rg -c -i "dialog or form where the end user enters or chooses data" content/skills/flow-plan/SKILL.md content/skills/flow-plan/references/ui-planning.md` devuelve 1 en cada archivo.

### T3 — Área «User experience» en la revisión del plan

- [x] `plan-review.md` tiene la fila «User experience» y la frase de asignación del diseño.

**Closes:** AC3.

**Depends on:** T2 (la fila usa el registro de datos que define T2).

**Locations:** `content/skills/flow-plan/references/plan-review.md`, «Choose the useful coverage».

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check — `rg`.

**Changes:** añadir la fila después de Frontend / UI y la frase después de «Skip absent domains…».

**Verification:** sobre `content/skills/flow-plan/references/plan-review.md`, `rg -c "^\| User experience"` devuelve 1, `rg -c "Assign User experience"` devuelve 1 y `rg -c "as a decision for that person, not as a fix" content/skills/flow-plan/references/plan-review.md` devuelve 1.

### T4 — Página del sitio y changelog

- [x] La página del flujo Plan menciona el revisor de experiencia de uso y `CHANGELOG.md` registra el cambio.

**Closes:** AC4.

**Depends on:** T1, T3.

**Locations:** `site/src/content/docs/flows/plan.md:34`; `CHANGELOG.md`, `[Unreleased]`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check — `rg` y el build del sitio.

**Changes:** completar la frase del sitio y añadir las dos entradas del changelog.

**Verification:** `rg -c -i "twice" site/src/content/docs/flows/plan.md` devuelve 1; `npm run build` en `site/` termina sin errores.

## Verificación y revisión humana

- Gates compartidos sobre el candidato final: `go vet ./...`, `go test ./...` y `npm run build` en `site/`.
- Verificación por tarea con `hive-verify-task`, con contexto nuevo, antes de marcar cada tarea `[x]`.
- Sin revisión de código dedicada (D8-A). Sin recorrido de UI: no hay interfaz renderizada que cambie de diseño, estado o flujo.
- Parada humana (D7-A): antes del push, el usuario valida el diff de las cinco rutas.

## Revisión del plan y progreso

- 2026-10-24: plan guardado (revisión `6c625d0087e4`); un `hive-review-plan` nativo (área: guía para agentes) lo revisó en ~2 min, 104k tokens. Sin hallazgos bloqueantes en criterios ni verificación.
- Aceptados y aplicados como los propuso el revisor, sin otra ronda: desempate entre corrección y hueco en el párrafo de admisión y commit de T2–T3 con la sesión `94440d0a` (hallazgo 1); «end user» frente a la persona que pidió el plan (hallazgo 2); alcance «in the product's interface» y «own `hive-review-plan` assignment» (hallazgo 3); condición de lectura de `ui-planning.md` y `flow-plan/SKILL.md:28` ampliada a diálogos y formularios de captura (AC5) y registro faltante como hueco bloqueante (sugerencia 1); verificación por tarea solo con `rg` (sugerencia 2); dos lugares revisados más (sugerencia 3).
- Límite: el revisor no ejecutó pruebas ni el build del sitio; quedan para los gates compartidos.
- Build 2026-10-24 en `feat/ux-plan-review-admission-scope`: T1 `15006c20`, T2 `8b3dcee0`, T3 `6e36aa22`, T4 `98d1746e`, cada una con un `hive-verify-task` nuevo y veredicto `met` en todos sus criterios (AC1; AC2 y AC5; AC3; AC4), sin rondas de corrección. Implementador: hilo principal, `claude-opus-5-5` (observado en el entorno de la sesión); verificadores: perfil `verifier` del rol, modelo configurado, no observado.
- Gates compartidos sobre el candidato final `98d1746e`: `go vet ./...` sin salida, `go test ./...` con 20 paquetes `ok` y salida 0; `npm run build` en `site/` completo (10 páginas). Sin revisión de código dedicada (D8-A); sin revisión de seguridad (no cruza ningún límite); sin recorrido de UI.
- 2026-10-24: el usuario validó el texto (D10-A). Push, [PR #128](https://github.com/JhonHawk/tricell-hive/pull/128) sin revisión dedicada, base al día (`merge-base --is-ancestor` correcto), sin CI configurada en el repositorio; merge `84735609`. `hive update` desplegó la release `a1521e3d` y el texto nuevo está en `~/.agents/skills/flow-plan/references/plan-review.md`.
