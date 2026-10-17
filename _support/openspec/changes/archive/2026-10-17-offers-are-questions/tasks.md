# Tareas

Base del cambio: `aa17fbe1` (`feat/delegation-threshold-global`), confirmada al empezar el build el 2026-10-17.

## T1: la oferta es una pregunta en `questions.md`

- [x] `questions.md` trata como pregunta toda oferta con la que termina un mensaje, y "When to read" incluye ese caso.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-plan/references/questions.md:5` y la sección "Decide whether a question is needed" (líneas 7-11).

**Execution:** hilo principal. Son dos inserciones cortas y un brief para delegarlas sería más largo que el cambio.

**Test approach:** check, `rg`.

**Changes:** el texto de [design.md § T1](design.md#t1-questionsmd).

**Verification:**
- `for p in "and before ending a message with an offer or request that waits on the user's answer" "a message would end with an offer or request that waits on the user's answer, in any form" 'when the action is already authorized, including the checks the global guidance says not to ask about separately' 'do it instead of offering it' "under this section's first condition" 'or a question another rule requires, such as the route, close or ticket question' 'ask it under this procedure, never as a closing sentence alone' 'leave the offer out'; do echo "$p: $(rg -n -F "$p" content/skills/flow-plan/references/questions.md | cut -d: -f1)"; done` → la primera en la línea 5 y las demás entre `## Decide whether a question is needed` y `## Use a waiting native card` (`rg -n '^## ' …`).
- `rg -c -F 'Never ask only for leave to proceed or to show something you could show' content/skills/flow-plan/references/questions.md` → 0.

## T2: el aviso de `global.md:20`

- [x] El aviso de `global.md` manda leer `questions.md` también antes de terminar un mensaje con una oferta o petición que espera la respuesta del usuario, sin ejemplo.

**Closes:** AC2.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:20`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** el texto de [design.md § T2](design.md#t2-globalmd20).

**Verification:**
- `rg -n -F "Before formulating a decision, ending a message with an offer or request that waits on the user's answer in any form, or requesting user-dependent review, validation or manual action, read [questions and pauses](skill:flow-plan/references/questions.md)." content/guidance/global.md` → la línea 20.

## T3: la oferta de `flow-report`

- [x] `reporting.md:41` ofrece la página como opción de la pregunta de cierre, o la pregunta siguiendo `questions.md`.

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-close/references/reporting.md:41`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg` y la prueba de enlaces.

**Changes:** el texto de [design.md § T3](design.md#t3-reportingmd41).

**Verification:**
- `rg -n -F 'offer a `flow-report` page as an option of the question the message ends with, such as a route or close question, or, when the message ends with none, ask it by following [questions](../../flow-plan/references/questions.md).' content/skills/flow-close/references/reporting.md` → la línea 41.
- `go test ./tooling/management/ -run TestRepositoryCatalogueInstructionReferences` pasa.

## Verificación y revisión

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos | Incluida |
| Tamaño de `global.md` | Candidato final | `wc -c` frente a 28 933 bytes en la base; la diferencia va en el commit de T2 | Incluida |
| Revisión del alcance | Candidato final | Releer `questions.md:5-24`, `global.md:20,66` y `reporting.md:39-87` y confirmar que no hay contradicción | Incluida |
| Revisión de código dedicada | No aplica | Cambios solo de texto | — |
| Despliegue local | Tras los commits | `hive update --out <plan>` + `hive apply --plan <plan>` | C1-A |
| Prueba en sesiones reales | Tras el despliegue | Tú, en una sesión nueva de Claude Code que termine con una oferta | Tuya |

## Revisión y progreso

- 2026-10-17: borrador escrito tras R5-A.
- 2026-10-17: revisión del plan sobre `403c145c2b7a` con un `hive-review-plan` nativo (dominio: procedimiento de preguntas). Duró 94 s y usó ~109k tokens.
  - **Bloqueante:** H1 ("only reads or reports" chocaba con `global.md:66` y `questions.md:11`): aceptado. Se hace todo lo ya autorizado, incluidas las comprobaciones que no se preguntan, y el ejemplo pasa a ser leer el diff.
  - **No bloqueantes:**
    - H2: la oferta se define por su función, con tres formas de ejemplo; `global.md` va sin ejemplo.
    - H3: tercera salida, omitir la oferta.
    - H4: se une la frase de la línea 11.
    - H5: se arregla la redacción de "When to read".
    - H6: "the question the message ends with".
    - H7: se nombra el detector de `tests/pilot`.
  - Los cambios tocan AC1–AC3, así que van a una re-revisión, la única ronda.
- 2026-10-17: re-revisión sobre `efdb1d5e8c7e`, reanudando al mismo revisor. Duró 36 s y usó ~117k tokens.
  - Sin bloqueantes. Las tres comprobaciones obligatorias se cumplen.
  - **No bloqueantes, aplicados como los propuso el revisor, sin otra ronda:**
    - H8: las preguntas que otra regla exige no se omiten.
    - H9: "this section's first condition".
    - H10: "reading or showing".
    - H11: texto viejo corregido.
- **Cobertura:** el dominio quedó revisado en `403c145c2b7a` y re-revisado en `efdb1d5e8c7e`.
- 2026-10-17, build: T1–T3 se implementaron en el hilo principal (claude-opus-5-5, observado). Cada una pasó a `[x]` con veredicto `met` de `hive-verify-task`, con modelo `opus` y esfuerzo `high` (configurado, no observado). Las tres verificaciones corrieron en paralelo:
  - T1, AC1: 36 s.
  - T2, AC2: 21 s.
  - T3, AC3: 66 s.
- **Compuertas del candidato final:**
  - `go vet ./...` sin hallazgos y `go test ./...` sin fallos.
  - Tests de Python OK.
  - `TestRepositoryCatalogueInstructionReferences` pasa.
  - `global.md` pasó de 28 933 a 29 021 bytes (+88).
- **Commits en `feat/delegation-threshold-global`:** `87c60881` (T2), `d54f3a75` (T1) y `063df6ca` (T3), sin push.
- **Despliegue:** `hive update --out` + `hive apply --plan` desde `063df6ca`, transacción `1d1cd3c1…`. El texto nuevo está en las instrucciones globales de los seis hosts y en el `questions.md` instalado.
- **Estado:** desplegado en local para tus pruebas. Siguiente paso: tú lo pruebas y decides el merge a `development`.
