# Tasks

Base `development` @ `6d1ab6f0`, rama `fix/session-findings`. T1–T6, T8 y T9 corren en el hilo principal: cada una edita pocas hojas cortas o archivos fuera del repositorio. T7 va a un hijo `hive-build-backend`: edita más de tres archivos en Go con la interfaz fijada por `session_audit.go` (`criterionAssessment`, `auditSession`) y no comparte archivos con las demás tareas salvo la rúbrica y `rule-enforcers.md`, que T3 y T5 también tocan; por eso T3 y T5 editan sus filas después de que T7 termine.

El check del texto, que esta carpeta llama «check de texto», es: `go test -count=1 ./tests/content/... ./integrations/... && go test -count=1 -run 'TestRealStagePacketsKeepAMarginUnderTheCap|TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|TestRepositoryInstructionAnchorsResolve|TestRepositoryCatalogueInstructionReferences|TestAntigravityRulePartsFromRealGuidanceFitAndReproduceIt' ./tooling/management/ && go test -count=1 -run 'TestFrozenInputsIncludeRuntimeAndLocalDependencies' ./tooling/package/ && go test -count=1 -run 'TestVoiceListShowsRepositoryVoices' ./tooling/cli/` (80 s en la base).

### T1 — Checks por ruta, commit desplegado y CHANGELOG

- [x] `AGENTS.md` asigna un comando a cada ruta y da el comando del commit desplegado

**Closes:** AC10, AC13

**Depends on:** none

**Locations:** `AGENTS.md`, `CONTRIBUTING.md`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Añade a `AGENTS.md` una sección `## Checks by path` con una tabla ruta → comando: todo `content/` salvo `content/skills/**/scripts/**`, `_support/`, `AGENTS.md`, `README.md`, `llms.txt`, `CHANGELOG.md` y `CONTRIBUTING.md` → el check de texto; `site/**` → `pnpm check`, `pnpm lint` y `pnpm build` en `site/`; `tooling/viewer/frontend/**` → sus comandos de pnpm de `CONTRIBUTING.md`; scripts Python de `content/skills/**/scripts/` y `tests/skills/**` → `python3 -m unittest discover -s tests/skills -p '*_test.py'`; `tests/pilot/**` y `tests/fixtures/sessions/**` → `go test -count=1 ./tests/pilot/`; cualquier otra ruta → la suite completa de `CONTRIBUTING.md` (`go vet ./...`, `go test ./...` y la de Python). Mantiene la excepción de `_support/docs/architecture/instruction-resources.md`: renombrar o quitar un skill o recurso pide la suite completa. `CONTRIBUTING.md`, junto a la suite, enlaza esa sección en vez de copiarla. En «Refreshing a local installation» sustituye «the source commit the previous deploy printed» por `hive releases | jq -r '[.[]|select(.consumers|length>0)][0].commits[-1]'` y `git log --oneline <ese>..HEAD -- content`; las instrucciones del binario no cambian. `CHANGELOG.md` gana una entrada bajo `[Unreleased]` con lo que cambia este trabajo.

**Verification:** `rg -n '^## Checks by path|hive releases' AGENTS.md` imprime las dos líneas; `rg -n 'Checks by path' CONTRIBUTING.md` imprime el enlace; `time` del check de texto pasa en menos de 120 s; el comando de `hive releases` imprime un commit que `git cat-file -t` reconoce como `commit`.

### T2 — Alcance conocido y reutilización del pase de la suite

- [x] `checks.md` y `flow-plan` paso 3 usan el check asignado a cada ruta

**Closes:** AC14

**Depends on:** none

**Locations:** `content/skills/hive-core/build/checks.md`, `content/skills/flow-plan/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En «Check reuse», el alcance es conocido cuando la guía del proyecto asigna un check a cada ruta cambiada. En el párrafo de la ronda de arreglos, una ronda cuyas rutas están todas asignadas corre solo esos checks y reutiliza el pase de la suite completa del candidato; la suite completa vuelve a correr solo por una ruta sin check asignado. Conserva las frases que fija `tests/content/single_home_test.go` o actualiza el test con la misma intención. En `flow-plan` paso 3, la línea `Verification:` de una tarea nombra el check que la guía del proyecto asigna a sus rutas cuando lo hay, no la suite completa, que pertenece a la verificación del candidato. No cambia «Local CI» ni «Long checks».

**Verification:** el check de texto pasa; `go test -count=1 -run TestRealStagePacketsKeepAMarginUnderTheCap -v ./tooling/management/` muestra `flow-build/stage.md` y `flow-verify/stage-entry.md` bajo 23.500 B.

### T3 — Segunda revisión por trozos del diff y archivos movidos fuera del alcance

- [x] `review-findings.md` y `reviewer-brief.md` acotan la revisión

**Closes:** AC16

**Depends on:** T7 (fila de la rúbrica)

**Locations:** `content/skills/hive-core/review/review-findings.md`, `content/skills/hive-core/delegation/reviewer-brief.md`, `content/skills/flow-verify/SKILL.md` (paso 2, «the mechanism the user chose»), `site/src/content/docs/flows/verify.md`, `_support/docs/harness-engineering/session-review-rubric.md` (fila «Re-review only after a fix that changed logic or a contract»)

**Execution:** main thread

**Test approach:** check

**Changes:** «A second review» corre cuando algún trozo del diff de los arreglos está fuera de un archivo `.md` o de test y cambia algo más que líneas de comentario, o cambia un contrato; un arreglo cuyos trozos son todos de ese tipo no la lanza y lleva `Unreviewed: <reason>` en su commit. El resto del punto queda igual y conserva las frases de `single_home_test.go`. En `reviewer-brief.md`, el artefacto de una revisión de agente excluye los archivos `R100` de `git diff -M --name-status` y los nombra en una línea; para un mecanismo nativo que acepta rutas, se le pasan solo las editadas. El sitio (`verify.md`, «Fixes, in one round») y la fila de la rúbrica describen el disparador nuevo. `flow-verify/SKILL.md` y `verify.md` dicen «el mecanismo de revisión en vigor» en vez de «el que eligió el usuario».

**Verification:** el check de texto pasa; `go test -count=1 -run TestRealStagePacketsKeepAMarginUnderTheCap -v ./tooling/management/` muestra `flow-verify/stage-exit.md` bajo 23.500 B.

### T4 — Relectura del plan con un solo hogar y un tope

- [x] `change-folder.md` fija el disparador y el tope; `plan-review.md` solo el dominio

**Closes:** AC18

**Depends on:** none

**Locations:** `content/skills/hive-core/plan/change-folder.md`, `content/skills/flow-plan/references/plan-review.md`, `site/src/content/docs/flows/plan.md`

**Execution:** main thread

**Test approach:** check

**Changes:** `change-folder.md` paso 4: una revisión se relee solo cuando cambió un contrato, una interfaz, un criterio, una decisión, un supuesto, un hecho externo, una línea de verificación o el delta; un revisor no relee la revisión que aplica tal cual su propia propuesta; durante la planificación hay como máximo una ronda de relectura, y lo que sobreviva o traiga una revisión posterior va al usuario como decisión en el resumen del plan; durante el build, una ronda por revisión como hoy. `plan-review.md` «Reread a revision» dice solo qué revisor relee según su dominio y enlaza `change-folder.md` para el disparador y el tope. El sitio (`plan.md`, «Revising a plan in progress») lo describe.

**Verification:** el check de texto pasa.

### T5 — Revisión anunciada en una línea y `hold` que no mezcla trabajos

- [x] `delivery-ask.md` y su disparador en `global.md` cubren D6 y D7

**Closes:** AC19, AC20

**Depends on:** T7 (fila de la rúbrica)

**Locations:** `content/skills/hive-core/git/delivery-ask.md`, `content/guidance/global.md` (línea del índice de `delivery-ask`), `content/skills/hive-core/messages/questions.md` (puntos de las rondas que siguen preguntándose y del canal), `AGENTS.md` (línea del mecanismo preferido), `site/src/content/docs/flows/plan.md`, `site/src/content/docs/getting-started.md` (líneas 46–50), `_support/docs/harness-engineering/session-review-rubric.md` (fila «Delivery and review asked once before the first edit»)

**Execution:** main thread

**Test approach:** check

**Changes:** «Review»: cuando la guía del proyecto nombra un mecanismo que corre en este host, sin contar una clave retirada `Review:` del `## Hive`, el agente lo anuncia en una línea que dice cómo cambiarlo y no termina en pregunta, y ese mecanismo es la respuesta de revisión en vigor; si no corre aquí o no hay preferencia, la pregunta sigue como hoy. «When»: se pregunta otra vez la entrega, sola o en la misma ronda que la pregunta de dónde trabajar de `git-workflow` paso 3 cuando ambas aplican, cuando la entrega en vigor es `hold` y el árbol tiene cambios sin commit de un trabajo anterior, antes de la primera edición del siguiente. La primera frase de la hoja y la línea del índice de `global.md` cubren ese caso, como exige el test del índice. `questions.md` enlaza esa excepción donde dice que la revisión se pregunta. La línea de `AGENTS.md` sobre `/code-review` dice que fija el mecanismo y no autoriza efectos. Sitio (`plan.md` y `getting-started.md`, que dice que un mecanismo que nombra la guía del proyecto se anuncia y que una clave retirada `Review:` sigue solo informando la recomendación) y fila de la rúbrica lo describen. Conserva las frases de `single_home_test.go`.

**Verification:** el check de texto pasa; `wc -c content/guidance/global.md` queda en 30.000 B o menos; `go test -count=1 -run TestRealStagePacketsKeepAMarginUnderTheCap -v ./tooling/management/` muestra `flow-plan/stage.md`, `hive-mode/stage.md` y `unattended-delegation/stage.md` bajo 23.500 B.

### T6 — `Situation:` no sustituye el texto previo a la tarjeta

- [x] `questions.md` «Channel» lo dice en una frase

**Closes:** AC9

**Depends on:** T5 (mismo archivo y mismos paquetes)

**Locations:** `content/skills/hive-core/messages/questions.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En el punto «Through the card», una frase: la línea `Situation:` de la tarjeta no sustituye el texto visible previo (línea de contexto, informe y línea de espera). No cambia el formato. Su commit nombra las observaciones: Claude Code, Opus 5.5, 89 de 132 tarjetas sin texto previo en 29 sesiones del 20 al 22 de noviembre, con dos ejemplos (`fa40070d` línea 174, `00c7fa30` línea 646).

**Verification:** el check de texto pasa; `go test -count=1 -run TestRealStagePacketsKeepAMarginUnderTheCap -v ./tooling/management/` muestra `flow-plan`, `flow-verify/stage-exit`, `hive-mode` y `unattended-delegation` bajo 23.500 B.

### T7 — Dos criterios de auditoría: suite tras arreglo de texto y rondas de revisión por candidato

- [x] `--audit-session` informa `full_suite_after_text_fix` y `review_rounds_per_candidate`

**Closes:** AC15, AC17

**Depends on:** none

**Locations:** `tests/pilot/session_audit.go` (registro en `auditCriteria`), un archivo nuevo `tests/pilot/session_audit_reruns.go`, `tests/pilot/session_audit_test.go`, fixtures nuevas en `tests/fixtures/sessions/`, `_support/docs/harness-engineering/session-review-rubric.md` (dos filas nuevas), `_support/docs/harness-engineering/rule-enforcers.md`

**Execution:** `hive-build-backend`

**Test approach:** tdd (contrato del informe de auditoría)

**Changes:** Una corrida de suite sin filtro se reconoce con `shellSegments` y sin contar texto dentro de argumentos o heredocs: `go test ./...` y `go -C <dir> test ./...` con `-count` o `-race`, `npm test`, `pnpm test` y `pytest` sin argumentos; `-run` o un patrón de subpaquete la filtran. `full_suite_after_text_fix`, en el mismo directorio de trabajo o `-C` y sin `git push`, `gh pr merge` ni rama o worktree nuevos entre ambas: `fail` cuando una corrida sigue a otra que se muestra pasada (un eco de su propio `$?` en el comando siguiente cuya línea completa muestra 0, o el indicador de éxito de un comando sin tubería ni segundo plano) y todo lo cambiado entre ellas son archivos `.md`, `.mdx` o `.txt`, por herramientas de edición o por escrituras de shell cuyo destino es de esos tipos; `pass` cuando hay otra edición o un rebase, merge, pull, checkout, switch, cherry-pick, am, apply, reset o restore; `review` cuando una escritura de shell tiene un destino que no se puede mostrar de texto; `git add`, `git commit` y `git push` no cuentan como cambios; `review` cuando no se puede mostrar que la primera pasó; `not_observed` sin dos corridas. Sin cambios entre dos corridas cuenta como solo texto, así que da `fail` solo si la primera se muestra pasada. `review_rounds_per_candidate`, solo en sesiones principales de Claude Code: cuenta rondas de revisión de código (skill `code-review`, hijo `hive-review-code`; las de un mismo mensaje son una ronda) y se reinicia con `git push`, `gh pr merge`, una carga nueva de `flow-build` o `flow-plan` o un cambio de rama; `fail` con más de dos rondas, `review` cuando la tercera sigue a un `/code-review` que escribió el usuario, `pass` con dos o menos, `not_observed` sin revisiones, en hijos y en otros hosts. Cada caso tiene su fixture. La rúbrica gana dos filas con fecha y sesión de origen y `rule-enforcers.md` lista los dos enforcers citando `session_audit_reruns.go`. Corre el criterio sobre las transcripciones de Claude Code del 20 al 22 de noviembre de este repositorio, confirma a mano cada `fail` que dé y anota en su informe cuántas sesiones leyó y cuáles fallan. No cambia ningún otro criterio.

**Verification:** `go test -count=1 ./tests/pilot/` pasa con un caso por estado de cada criterio; `go test -count=1 ./tests/content/...` pasa; `go run ./tests/pilot --audit-session ~/.claude/projects/-Users-jmartinez-Development-projects-tricell-tricell-hive/7319321c-5a79-42a9-91a8-17c47fcfbb69.jsonl --host claude` da `review_rounds_per_candidate` en `fail` con la línea 3346 como evidencia; el informe de T7 nombra las sesiones del 20 al 22 de noviembre con `full_suite_after_text_fix` en `fail` confirmado a mano, o dice que ninguna coincidió entre las n leídas.

### T8 — Regla personal «Where to work» en una línea cuando es «in place»

- [x] `rules.md` y sus copias dicen que «In place» se anuncia y sigue

**Closes:** AC11

**Depends on:** none

**Locations:** `~/.config/hive/rules.md` y las copias que despliega `personal_rules.py deploy` (fuera del repositorio)

**Execution:** main thread, por el modo de reglas personales de `audit-harness`

**Test approach:** check

**Changes:** Primer párrafo de «Where to work»: cuando se cumplen todos los criterios de «In place», el agente dice en una línea dónde trabaja y por qué y sigue; recomienda y pregunta solo cuando corresponde worktree o sandbox, en la misma ronda que la repregunta de entrega de T5 cuando ambas aplican, igual que la pregunta de `git-workflow` paso 3. Muestra el texto completo y el diff antes de escribir, como manda el modo. No cambia los criterios ni la sección de sbx.

**Verification:** `python3 -I ~/.agents/skills/audit-harness/scripts/personal_rules.py check` imprime `ok: every copy matches rules.md`; `rg -n 'one line' ~/.claude/CLAUDE.md` muestra la frase nueva en «Where to work».

### T9 — Watch items y revisión de H4/H5

- [ ] Engram tiene los watch items de H8/H9 y el cierre registra la revisión de H4/H5

**Closes:** AC12

**Depends on:** T1–T8 integradas

**Locations:** Engram, proyecto `tricell-hive`

**Execution:** main thread, en el cierre

**Changes:** Un `watch/<slug>` por caso con conteo 1, observación (host, modelo, sesión, fecha), versión de la regla, línea `Enforceable:` y condición de reapertura: investigación que ignora la memoria (`fb00087f`), metas de poda inalcanzables (`a6144721`), merge y suite encadenados (`6fb1c434`), guardado de memoria fuera de los momentos de la regla (`aa27a201`, `e032f89a`). Revisa con `--audit-session` las sesiones posteriores a #599/#601/#600 por si vuelven H4 o H5 y anota el resultado en el progreso.

**Verification:** `mem_search` con `watch/` en `tricell-hive` devuelve las cuatro entradas.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Candidato final, una vez, localmente | `go vet ./... && go test ./... && python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Candidato final, antes del push | `/code-review` nativo de Claude Code (D6) | The review answer |
| Refresco local | Tras el merge a `development` | `hive update --source . --out FILE` y `hive apply --plan FILE`; las hojas instaladas muestran AC14, AC16, AC18–AC20 y AC9 | Regla del mantenedor en `AGENTS.md` |

El PR lista los lugares revisados por cada regla cambiada y el tamaño de `content/guidance/global.md`, como pide `AGENTS.md` «Guidance structure».

## Checkpoint

- Updated: 2026-11-23
- Stage: verify · step 2 code review running
- Last commit: a3b291ac
- Open findings: none
- Next: judge /code-review findings, one round of fixes, then push, PR, merge (automatic)
- Rules: flow-verify/stage-entry.md

### Evidence

- before captures: none (no screen)
- go vet ./... · exit 0 · a346b972
- go test ./... · exit 0 · a346b972 (a3b291ac only removes comments; go vet ./tests/pilot/ exit 0)
- python3 -m unittest discover -s tests/skills · 412 OK · a346b972
- site pnpm check/lint/build · exit 0 · a346b972
- text check · exit 0, 95 s · 3caa7191 tree
- go test ./tests/pilot/ · ok 62.9 s · c107eced
- audit 7319321c review_rounds_per_candidate fail at 3346 · c107eced

## Review status and progress

2026-11-23 — T1 · f3fbef6e · unverified: none
2026-11-23 — T2 · 924b3f52 · unverified: AC14 seen in the installed leaf only after the local refresh
2026-11-23 — T4 · 9a7bd1cb · unverified: AC18 seen in the installed leaf only after the local refresh; site build runs once after T3 and T5
2026-11-23 — T8 · uncommitted · unverified: outside the repository; previous rules.md kept in the session scratchpad for recovery
2026-11-23 — T7 · c107eced · unverified: AC15: no real session of 2026-11-20..22 gave a fail (0 fail, 4 pass, 10 review in 33); proven by fixtures only
2026-11-23 — T3 · 3caa7191 · unverified: AC16 seen in the installed leaf only after the local refresh
2026-11-23 — T5 · 8af977ab · unverified: AC19, AC20 seen in the installed leaf only after the local refresh
2026-11-23 — T6 · a346b972 · unverified: AC9 seen in the installed leaf only after the local refresh
