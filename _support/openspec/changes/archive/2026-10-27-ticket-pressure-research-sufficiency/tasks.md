# Tareas — ticket-pressure-research-sufficiency

Base: `82ccf111` (development), confirmada al iniciar el build (2026-10-27). Rama `fix/ticket-pressure-research-sufficiency` en `.claude/worktrees/ticket-pressure`.

## T1 — El ticket en la pregunta de ruta solo cuando el trabajo espera (D20-A)

- [x] `flow-research` ofrece abrir el ticket y parar solo cuando el trabajo no se hará ahora (grande, de otra persona u otro proyecto) o lo pide el usuario; una opción que actúa puede abrir el ticket del proyecto; `global.md:24` y `:78` dejan de imponerlo.

**Closes:** AC1, AC2, AC7.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-research/SKILL.md:42` y `:46`; `content/guidance/global.md:24` (conserva una condición breve para actuar sin el skill: abrir el ticket y parar cuando el trabajo no se hará ahora) y `:78`; `site/src/content/docs/flows/research.md:37` y `:47`.

**Execution:** hilo principal; edita `global.md`, que T2 y T3 también editan.

**Test approach:** check — `rg` de AC1 y AC2 y `go test ./tests/content/...`.

**Changes:** según el enfoque 1 de `design.md`.

**Verification:** `rg -n "one more option opens the ticket and stops" content/skills/flow-research/SKILL.md site/src/content/docs/flows/research.md` y `rg -n "plus opening a ticket and stopping when no ticket|whose options say whether they open a ticket" content/guidance/global.md` no encuentran nada; `rg -n "asks whether you want a ticket" site/src/content/docs/flows/research.md` no encuentra nada; lectura del párrafo nuevo con las tres condiciones (grande, otra persona u otro proyecto, petición del usuario) y la opción que actúa con tracker, y de la condición breve en `global.md:24`; `go test ./tests/content/...` pasa.

## T2 — Hallazgos al cierre: recomendar resolver en este trabajo (D21-A)

- [x] `incidental-findings.md` recomienda resolver en este trabajo lo que cabe en la sesión y el alcance, y el ticket solo fuera del alcance, grande o de otro responsable; `global.md:34` apunta a ese criterio.

**Closes:** AC3, AC7.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-close/references/incidental-findings.md:9`, `:15` (el título «proposed as an issue» y «not proposed as a ticket» pasan a «proposed») y `:22` (un hallazgo de interfaz se propone, como arreglo o ticket, solo después de verlo renderizado); `content/guidance/global.md:34`; `content/skills/flow-build/references/verification.md:51` (las notas de infraestructura de pruebas o CI se proponen bajo la regla global de hallazgos incidentales, en lugar de ir siempre al tracker); `site/src/content/docs/flows/close.md:47` (la frase entera).

**Execution:** hilo principal; comparte `global.md` con T1 y T3.

**Test approach:** check — lectura de `incidental-findings.md:9` y `go test ./tests/content/...`.

**Changes:** según el enfoque 2 de `design.md`.

**Verification:** `rg -n "mark which ones you recommend opening and why" content/skills/flow-close/references/incidental-findings.md` no encuentra nada y `rg -n "recommend resolving it in this work when" content/skills/flow-close/references/incidental-findings.md` encuentra el criterio con sus condiciones de ticket; `rg -n "proposed as an issue|for the project's tracker" content/skills/flow-close/references/incidental-findings.md content/skills/flow-build/references/verification.md` no encuentra nada; `global.md:34` ya no dice solo «propose an issue in the report»; `rg -n "recommends turning into tickets" site/src/content/docs/flows/close.md` no encuentra nada; `go test ./tests/content/...` pasa.

## T3 — Destinos de los temas abiertos

- [x] `global.md:21` lista el ticket después de la tarea en este trabajo y el plan, y acepta que el tema pertenezca a otra persona u otro proyecto como motivo de descarte.

**Closes:** AC4.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/guidance/global.md:21`.

**Execution:** hilo principal.

**Test approach:** check — lectura de `global.md:21` y `go test ./tests/content/...`.

**Changes:** según el enfoque 3 de `design.md`.

**Verification:** la frase de destinos empieza por la tarea en este trabajo y nombra el motivo «belongs to another person or project»; `go test ./tests/content/...` pasa.

## T4 — Suficiencia de la investigación y brief de los hijos (D22-A)

- [x] `flow-research` exige que los hallazgos respondan cada parte de la pregunta y basten para elegir ruta, que el hilo principal compruebe sus inferencias decisivas cuando haya una comprobación barata o recomiende seguir investigando, y que el brief de cada hijo nombre la decisión y los aspectos pedidos con parada por suficiencia; el rol `hive-research` usa ese criterio para `complete`.

**Closes:** AC5, AC6, AC7.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-research/SKILL.md:26` (brief de los hijos), `:34` (se amplía la viñeta de la comprobación acotada, sin crear otra regla) y «Deliver the finding» (la prueba de suficiencia antes de recomendar una ruta); `content/agents/review/hive-research.md:18` (estado `complete`); `site/src/content/docs/flows/research.md:37`.

**Execution:** hilo principal; comparte `flow-research/SKILL.md` con T1.

**Test approach:** check — `rg` de AC5 y AC6, lectura de las páginas del sitio y `go test ./...`.

**Changes:** según los enfoques 4 y 5 de `design.md`.

**Verification:** `rg -n "every part of the question|own deciding" content/skills/flow-research/SKILL.md` encuentra la regla; `rg -n "once the answer can be named and the evidence converges" content/skills/flow-research/SKILL.md` y `rg -n "answer converged" content/agents/review/hive-research.md` no encuentran nada, y `rg -n "enough for the decision" content/skills/flow-research/SKILL.md content/agents/review/hive-research.md` encuentra el criterio en los dos; la página Research describe el comportamiento sin listar las condiciones; `go test ./...` pasa y `flow-research/SKILL.md` sigue bajo 16 000 bytes.

## T6 — Qué es una petición solo de plan, y el build del plan recién hecho (D24-A)

- [x] `flow-plan` define la petición solo de plan (el usuario dice que no se implemente ahora o que se haga en otra sesión o más tarde; pedir un plan, aunque nombre el worktree o la rama, no lo es) y `plan-handoff.md` recomienda parar solo ante un bloqueo concreto que nombra y nunca propone otra sesión para el build del plan recién terminado.

**Closes:** AC8.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-plan/SKILL.md:72`; `content/skills/flow-plan/references/plan-handoff.md:9`; `site/src/content/docs/flows/plan.md:66`.

**Execution:** hilo principal; dos frases y una línea del sitio.

**Test approach:** check — `rg` de AC8 y `go test ./tests/content/...`.

**Changes:** según el enfoque 6 de `design.md`. Caso Codex `01a17ba2` (2026-10-26): recomendó «Dejar plan listo… como pediste» tras «comienza el plan para esto (lo trabajaremos en worktree)», con el contexto al 41 %. La regla de nueva sesión del cierre (`flow-close/SKILL.md`, «never for a plan still in progress») no cambia.

**Verification:** `rg -n "planning-only when" content/skills/flow-plan/SKILL.md` y `rg -n "named blocker" content/skills/flow-plan/references/plan-handoff.md` encuentran el texto nuevo; lectura de que pedir un plan que nombra el worktree no es solo de plan, de que la definición no autoriza implementar sin respuesta, y de que nunca se recomienda otra sesión para el build del plan recién hecho, nombrando como excepciones la ruta de Recommendations de `reporting.md` y el criterio de sesión nueva de `flow-close`; la página Plan lo describe; `go test ./tests/content/...` pasa.

## T5 — CHANGELOG y revisión del comportamiento completo

- [x] Entrada en `CHANGELOG.md` (Unreleased, Changed) y lista de los sitios revisados para el PR.

**Depends on:** T1, T2, T3, T4, T6.

**Locations:** `CHANGELOG.md`; búsqueda en `content/`, `site/`, `_support/docs/architecture/`, `tests/pilot/`.

**Execution:** hilo principal.

**Changes:** entrada que resume los cambios de T1–T4 y T6 con el tamaño de `global.md`; búsqueda de cada lugar que enuncia o depende de estas reglas (AGENTS.md, Guidance structure) para el cuerpo del PR.

**Verification:** `go test ./...` pasa.

## Verificación y revisión

- C1 — T1, T2, T3, T4, T6 — ningún cambio afecta a contratos públicos o persistentes ni es dependencia de dos tareas posteriores; se verifican juntas.

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Comprobaciones por criterio | tras T1–T6, worktree local | los `rg` y lecturas de cada tarea | parte de la implementación |
| Suite completa | candidato final, local | `go test ./...` en verde | parte de la implementación |
| Revisión dedicada | candidato local, antes del push | `hive-review-code`; `hive-refute-claim` para P0/P1 | D17-A |
| Actualización local | tras el merge | `hive update` (`--out` + `hive apply --plan`), texto nuevo presente en las copias instaladas | regla del mantenedor en AGENTS.md |

Sin comprobación en vivo: es texto de guía y los pilotos de comportamiento están en pausa (AGENTS.md, Measurement); el efecto se observa en sesiones reales.

## Estado de la revisión y progreso

- Revisión del plan: un `hive-review-plan` (coherencia de guía y orden de criterios) sobre la revisión `3fc57ca9f354`, base `82ccf111`. Resultado: 4 bloqueantes y 4 no bloqueantes; todos aplicados como los propuso el revisor, sin otra ronda (plan-review: corrección que aplica la propuesta del revisor).
  - B1: la verificación de T2 usa un ancla que no existe en la base («recommend resolving it in this work when») y la ausencia de «mark which ones you recommend opening and why».
  - B2: la verificación de T4 cubre el rol (`answer converged` desaparece; aparece «enough for the decision»).
  - B3: T2 añade `incidental-findings.md:15` y `:22` y la frase entera de `close.md:47`.
  - B4: T2 añade `verification.md:51`; la exclusión anterior era una suposición, no una decisión del usuario.
  - N1: T1 y T2 también cierran AC7; el falsificador de AC7 incluye `research.md:37` y `:47`.
  - N2: la comprobación de inferencias amplía `flow-research/SKILL.md:34` y define «barata» con el criterio de rutina de la guía global.
  - N3: «cabe en el trabajo» se ancla a un cambio pequeño y entendido para `flow-build` en el área del trabajo, sin efecto nuevo ni decisión de contrato.
  - N4: `global.md:24` conserva una condición breve para actuar sin el skill.
- Coste de la revisión: no lo informa el host.
- Re-revisión (única ronda) sobre `cf0aaa77e9e0`: B1–B4 y N1–N4 reflejados; AC8 falsable y cubierto por T6 en orden. R1 (bloqueante): T6 debía nombrar la excepción a la ruta de sesión nueva de `reporting.md:94` y de `flow-close`; aplicado en AC8, el enfoque 6 y la verificación de T6. N5: añadido el enfoque 6, que deja claro que la definición no autoriza implementar. N6: AC8 se limita a lo que añade sobre `plan-handoff.md:9`. N7: T5 y la tabla de gates incluyen T6. Todo aplicado como lo propuso el revisor; no quedan huecos para el usuario.
- Revisión 2 (D24-A, 2026-10-27): añade AC8 y T6 (petición solo de plan; no recomendar otra sesión para el build del plan recién hecho), con la cita del usuario en la solicitud; T5 depende también de T6. No invalida tareas escritas. Cambia un criterio, así que va a una re-revisión del mismo `hive-review-plan` limitada a AC8 y T6.
- Build (2026-10-27): commits `7174cd7b` (T1), `440cc9b1` y `cfb73d67` (T2; el segundo ajusta el texto al ancla de su verificación, que distingue may/min), `c333c521` (T3), `04f127dc` (T4), `9a2af404` (T6), `fa3d15ef` (T5). `go test ./...` en verde tras T4 y T5. `global.md` 40 942 → 41 113 bytes; `flow-research/SKILL.md` 11 803; `flow-plan/SKILL.md` 14 513.
- T5 incluye, como arreglo peque\u00f1o del mismo archivo, la entrada del `CHANGELOG` que faltaba del PR #224 (regla de arreglar defectos que bloquean).
- Modelos: implementador `anthropic/claude-opus-5-5` (observado: entorno de la sesi\u00f3n Pi); verificadores `hive-verify-task` con el perfil `verifier` de Pi (hereda el modelo de la sesi\u00f3n, esfuerzo high; configurado, no observado).
- Verificaci\u00f3n C1, por tandas por archivos compartidos: T1 y T6 \u2192 T2 y T4 \u2192 T3. Tanda 1 en curso.
- Revisión dedicada `hive-review-code` sobre `fa3d15ef`: P0 0, P1 0, P2 3 corregidos en `edf5f4af` (condición observable del ticket en `global.md:24`; el trabajo de otro proyecto se reporta para su sesión, sin ticket; el mensaje de `edf5f4af` registra las sesiones observadas de todo el cambio). P3 4–6 reportados sin arreglar (regla P3 de review-and-gates).
- Integrado por el PR #225 (`dcaca050`, 2026-10-27); `hive update` aplicado a los seis hosts.
