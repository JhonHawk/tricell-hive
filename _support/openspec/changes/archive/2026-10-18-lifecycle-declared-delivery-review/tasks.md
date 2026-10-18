# Tareas

Base del cambio: `e484b47b` (`feat/lifecycle-declared-delivery-review`, creada desde `development`), confirmada al empezar el build el 2026-10-17.

## T1: entrega y revisión declaradas (R1, R2)

- [x] `delivery-decisions.md`, `project-settings.md`, `global.md:66`, `flow-build/SKILL.md:20`, `verification.md:33`, `flow-plan/SKILL.md:18` y `flow-close/SKILL.md:67` dicen que un `Delivery` o `Review` declarado se enuncia en una línea y no se pregunta.

**Closes:** AC1, AC2, AC3.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:5,18,20,22,26,34`; `content/skills/flow-plan/references/project-settings.md:9,13-21`; `content/guidance/global.md:66`; `content/skills/flow-build/SKILL.md:20`; `content/skills/flow-build/references/verification.md:33,43`; `content/skills/flow-plan/SKILL.md:18`; `content/skills/flow-close/SKILL.md:67`; `content/skills/flow-close/references/reporting.md:79` (añadido en la ronda de corrección).

**Execution:** hilo principal. Son reemplazos ya escritos en el diseño; un brief sería más largo que el cambio.

**Test approach:** check, `rg`.

**Changes:** los textos de [design.md § T1](design.md#t1--textos-de-r1-y-r2).

**Verification:**
- AC1, una frase por cláusula, cada una → 1 en `content/skills/flow-plan/references/delivery-decisions.md`, y la primera en el párrafo que empieza "Before declaring an implementation plan ready": `for p in 'is the mode for work there and authorizes that mode' 'or a narrower instruction in the request, such as `hold` or commit only, overrides the declared value' 'the mode and where the work stops; the review mechanism that will apply; any deployment the destination triggers; and the change record'"'"'s closing route' 'the work then runs under `interactive`, and the line names that rule and the stop' 'Otherwise, when no `Delivery` value is declared, ask once' 'goes through each question whose value it does not declare' 'for those without a declared `Delivery` value and one review question for those without a declared `Review` value' 'or the declared value'"'"'s route, covers versioning that record'; do echo "$p: $(rg -c -F -e "$p" content/skills/flow-plan/references/delivery-decisions.md)"; done`.
- `rg -c -F -e 'a repository default or remembered preference only informs' content/skills/flow-plan/references/delivery-decisions.md` → 0.
- AC2: `for p in 'is the answer and authorizes that mechanism' 'Confirm without launching it, as the first bullet below sets' 'also authorizes `hive-review-security` when the change crosses a boundary'; do echo "$p: $(rg -c -F -e "$p" content/skills/flow-plan/references/delivery-decisions.md)"; done` → 1 cada una; `rg -c -F -e 'is an explicit preference' content/skills/flow-plan/references/delivery-decisions.md` → 0.
- `rg -c -F -e 'declared in the repository' content/guidance/global.md` → 1, en la viñeta de compuertas (`rg -n '^## ' content/guidance/global.md` sitúa "Verification gates").
- `rg -c -F -e 'settles each question it covers, and a declared setting is stated in one line, not asked' content/skills/flow-build/SKILL.md` → 1; `rg -c -F -e 'or a declared `Review` setting already settled it' content/skills/flow-build/SKILL.md` → 1.
- `rg -c -F -e 'or the declared `Review` setting settled' content/skills/flow-build/references/verification.md` → 1.
- `rg -c -F -e 'stated in one line when the repository declares them' content/skills/flow-plan/SKILL.md` → 1.
- `rg -c -F -e 'and none follows from a declared `Delivery` value' content/skills/flow-close/SKILL.md` → 1; `rg -c -F -e 'or on the route a declared `Delivery` value sets, with you as the merger' content/skills/flow-close/SKILL.md` → 1.
- `rg -c -F -e 'or the authorization a declared `Review` value gives' content/skills/flow-build/references/verification.md` → 1 (redacción ajustada en T4 tras la verificación).
- `rg -c -F -e 'is authorization for the mechanism or mode it names' content/guidance/global.md` → 1.
- Ronda de corrección tras la verificación: `rg -c -F -e 'or whose `Review` setting differs' content/skills/flow-plan/references/delivery-decisions.md` → 0; `rg -c -F -e 'or the one a declared `Delivery` value sets: a direct push' content/skills/flow-close/SKILL.md` → 1; `rg -c -F -e 'or takes them from the repository'"'"'s declared settings' content/skills/flow-close/references/reporting.md` → 1; `rg -c -F -e 'recommend the closest available mechanism and say which value the repository declared' content/skills/flow-plan/references/delivery-decisions.md` → 1.
- AC3: `rg -c -F -e '`automatic` or `interactive` as the delivery mode' content/skills/flow-plan/references/project-settings.md` → 1; `rg -c -F -e '- Delivery: automatic' content/skills/flow-plan/references/project-settings.md` → 1.
- `go test ./tooling/management/ -run TestRepositoryCatalogueInstructionReferences` pasa.

## T2: el manager acepta `automatic` e `interactive`

- [x] `hive doctor` y el formulario de proyecto aceptan los tres valores de `Delivery` y rechazan otros listándolos; el documento del manager lo dice.

**Closes:** AC4.

**Depends on:** ninguna.

**Locations:** `tooling/cli/doctor_project.go:49,354-357`; `tooling/cli/tui_project_view.go:272-275,313-322`; `tooling/cli/doctor_project_test.go:271-277`; `tooling/cli/project_write_test.go:129-131,183` (solo si cambia el prefijo del mensaje; hoy comprueban `Delivery: expected`); `tooling/cli/tui_project_view_test.go:353-362,377-379,405-407,422`; `_support/docs/architecture/deployment-manager.md:77,110`.

**Execution:** delegada a `hive-build-backend`: dos archivos Go con sus pruebas, interfaz fija (los tres valores y el mensaje) y sin solapamiento con las tareas de texto.

**Test approach:** tdd. RED: `TestDoctorProjectAcceptsDeliveryModes` falla en la base porque `Delivery: automatic` produce `expected "direct-base", found "automatic"` (`doctor_project.go:356`). Observar el fallo en la salida: `go test -run` con una prueba inexistente sale con 0 y "no tests to run", así que la prueba debe existir antes de observar RED.

**Changes:** [design.md § T2](design.md#t2--manager-go).

**Verification:**
- `go test ./tooling/cli/ -run 'TestDoctorProject|TestHiveSection|TestHiveText|TestProjectWrite|TestProjectViewForm'` pasa; cubre la prueba nueva (`Valid` con `automatic` e `interactive`; mensaje `expected one of "direct-base", "automatic", "interactive", found "pr"` con `pr`), el caso de `checkHiveText` que acepta `automatic`, la carga del formulario con `Delivery: automatic` sin opción duplicada, y el ciclo del selector en ambos sentidos.
- `rg -c -F 'hiveDeliveryWant' tooling` → 0; `rg -c -F 'their one valid value' tooling/cli/tui_project_view.go` → 0.
- `rg -n -F '`direct-base`, `automatic` or `interactive`' _support/docs/architecture/deployment-manager.md` → dos líneas, 77 y 110.

## T3: base actual antes del merge (R3)

- [x] `git-workflow/SKILL.md` exige la base actual antes de mergear, con la cola de merge como alternativa; `delivery-decisions.md:24` remite a ello.

**Closes:** AC5.

**Depends on:** T1, porque edita `delivery-decisions.md` y T1 lo edita antes; así ningún archivo recibe dos ediciones a la vez.

**Locations:** `content/skills/git-workflow/SKILL.md`, sección "Verify the authorized effect"; `content/skills/flow-plan/references/delivery-decisions.md:24`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** [design.md § T3](design.md#t3--texto-de-r3).

**Verification:**
- `gh pr view --json 2>&1 | rg -c mergeStateStatus` → 1 antes de editar; si da 0, ajustar el ejemplo del diseño a un campo existente.
- `for p in 'A green run on a stale base does not satisfy the merge condition.' 'but another status does not prove it current' 'git merge-base --is-ancestor origin/<base> origin/<branch>' 'When that check fails, bring the branch up to date with a merge from the base and push it, which the merge authorization covers' 'repeat once, and after that report the merge as pending' 'the rewrite and its force-push need the explicit authorization' 'add the pull request to the queue and wait for its result'; do echo "$p: $(rg -c -F -e "$p" content/skills/git-workflow/SKILL.md)"; done` → 1 cada una, dentro de "Verify the authorized effect" (`rg -n '^## ' content/skills/git-workflow/SKILL.md`).
- `rg -c -F 'once CI is green on the current base, as [git-workflow](../../git-workflow/SKILL.md) sets' content/skills/flow-plan/references/delivery-decisions.md` → 1.
- `go test ./tooling/management/ -run TestRepositoryCatalogueInstructionReferences` pasa.

## T4: clases de defecto y fronteras en el revisor (R4, R6)

- [x] `hive-review-code` reporta un defecto repetido una vez y nombra las fronteras de seguridad que toca el diff; `verification.md:43` las incluye.

**Closes:** AC6, AC7.

**Depends on:** T1, porque edita `verification.md:43` después de que T1 edite la misma línea.

**Locations:** `content/agents/review/hive-review-code.md:16,20`; `content/skills/flow-build/references/verification.md:43`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** [design.md § T4](design.md#t4--textos-de-r4-y-r6).

**Verification:**
- `rg -c -F 'names the defect class and every location' content/agents/review/hive-review-code.md` → 1, en el paso 4.
- `rg -c -F '[security boundaries](skill:flow-build/references/security-boundaries.md)' content/agents/review/hive-review-code.md` → 1, en la última línea.
- `rg -c -F "including one the code reviewer's verdict names" content/skills/flow-build/references/verification.md` → 1.
- `go test ./tooling/management/` pasa (incluye la validación de locators `skill:` de los agentes).

## T5: postmortem por capa (R5)

- [x] El postmortem de la ruta urgente nombra la capa que debió atrapar el fallo y pone ahí la acción.

**Closes:** AC8.

**Depends on:** T3, porque edita `delivery-decisions.md` después de T3.

**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:44`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** [design.md § T5](design.md#t5--texto-de-r5).

**Verification:**
- `rg -c -F 'which layer should have caught it (a check, a test, a standard, a review criterion, or a deployment check), and a preventive action in that layer with its owner' content/skills/flow-plan/references/delivery-decisions.md` → 1.

## T6: el sitio describe el comportamiento nuevo

- [x] `flows/build.md` dice que los valores declarados se enuncian y no se preguntan, que se declaran una vez en el `## Hive`, y que el merge espera checks sobre la base actual.

**Closes:** AC9.

**Depends on:** T1 y T3, cuyo comportamiento describe.

**Locations:** `site/src/content/docs/flows/build.md:41,64`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** [design.md § T6](design.md#t6--sitio).

**Verification:**
- `rg -c -F 'declares `Delivery` and `Review`, the agent states them in one line' site/src/content/docs/flows/build.md` → 1.
- `rg -c -F 'declare it once in the repository' site/src/content/docs/flows/build.md` → 1.
- `rg -c -F 'the checks ran on the current base branch' site/src/content/docs/flows/build.md` → 1.

## Verificación y revisión

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, este checkout | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos | Incluida en la implementación |
| Tamaño de `global.md` | Candidato final | `wc -c content/guidance/global.md` frente a 29 638 bytes en la base; la diferencia va en el commit de T1 y en este registro | Incluida |
| Revisión del alcance | Candidato final | Releer `delivery-decisions.md:5-36` completo y confirmar: una sola regla por hogar, los punteros de `flow-build:20`, `verification.md:33,43` y `delivery-decisions.md:24` no repiten pasos, y `global.md:66` enuncia solo la invariante de autorización | Incluida |
| Revisión de código dedicada | No aplica | C1-A: texto y un cambio de Go cubierto por `go test` | — |
| Despliegue local | Tras los commits | `go build -o "$(command -v hive)" ./tooling/cli` (T2 cambia el manager), después `hive update --out <plan>` + `hive apply --plan <plan>` | C1-A |
| Prueba en sesiones reales | Tras el despliegue | Tú, en sesiones nuevas; para R1 y R2 hace falta un repo con `Delivery` y `Review` declarados | Tuya |

## Revisión y progreso

- 2026-10-17: borrador escrito tras D1-A, D2-A, R1-A, R2-A, R3–R6 y C1-A. Revisión del plan con `hive-review-plan` en dos dominios sobre la revisión `947f0ffd3563`.
- 2026-10-17: revisor del manager Go (57 s, ~96k tokens). Citas confirmadas en el hilo principal.
  - **Bloqueantes, aceptados:** B1, `hive project --set` y el formulario sí validan (`project_write.go:525`), y `deployment-manager.md:110` faltaba; corregidos el diseño, el alcance y T2. B2, la rama del valor actual desconocido en el selector no tenía condición nueva; ahora `!slices.Contains(f.options[i], cur)`, y el comentario de `:272-275` entra en T2.
  - **No bloqueantes, aplicados como los propuso:** lista del mensaje construida desde el slice; import `slices`; pruebas para `checkHiveText`, carga del formulario con `automatic` y flecha izquierda; el aviso de que `go test -run` sin prueba sale con 0.
  - Ninguna corrección cambia un contrato ni un criterio: reconciliadas sin re-revisión.
  - **Sin hallazgo:** README y llms.txt no mencionan `Delivery`; la compuerta de despliegue coincide con AGENTS.md.
- 2026-10-17: revisor de la guía distribuida (113 s, ~126k tokens). Citas confirmadas en el hilo principal.
  - **Bloqueantes, aceptados:** B1, `:5,18,20` seguían mandando a la pregunta; ahora la pregunta es solo sin valor declarado, también con varios repos. B2, la línea previa omitía el despliegue que dispara el destino y la ruta de cierre del registro; ahora los nombra, con un PR de cierre sin revisión como ruta bajo valor declarado (decisión del orquestador, revertible en la respuesta del usuario a esa línea); `flow-close:67` entra como puntero. B3, "verify it can run" chocaba con `verification.md:26`; ahora "confirm without launching it". B4, el rebase chocaba con `git-workflow:24`; ahora merge desde la base, y un rebase exigido pide la autorización del force-push.
  - **No bloqueantes, aplicados como los propuso:** N1, comprobación nombrada (`gh pr view --json mergeStateStatus`, `git merge-base --is-ancestor`) y un reintento; N2, "authorizes" vive en `delivery-decisions.md` y `global.md:66` conserva solo la invariante, con la respuesta anterior de la sesión por encima del valor declarado; N3, la línea dice dónde se detiene el trabajo; N5, puntero en `flow-plan:18`; `rg -c -F -e` para frases que empiezan con guion; una comprobación por cláusula de AC1.
  - **Decisión para el usuario (N4):** `Review: none` frente a fronteras de seguridad; redactado como N4-A (la revisión de seguridad se mantiene) y pendiente de confirmación.
  - B1, B2 y B4 cambian el contenido de las reglas: una re-revisión, la única ronda, reanudando al mismo revisor.
- 2026-10-17: re-revisión sobre `24b06c7ba5fc` (48 s, ~140k tokens). B3 y B4 quedaron cerrados; sin contradicción con `git-workflow:24,40`. Cuatro bloqueantes nuevos:
  - **Aplicados como los propuso:** 1, un repo con un solo valor declarado pasa por la pregunta cuyo valor no declara (`:5,18,20`). 2, la ruta de cierre por defecto se limita al registro en este repo o en specs sin `direct-base` con el modo declarado vigente, y `flow-close:67` dice "the recorded answer, or the route a declared `Delivery` value sets, with you as the merger". 4, `global.md:66` dice "is authorization for the mechanism or mode it names". N-a, la regla es `git merge-base --is-ancestor` y `BEHIND` solo indicio positivo (no verificado si GitHub lo reporta sin protección "up to date"). N-b, `:22` nombra también la ruta del valor declarado.
  - **Decisión para el usuario:** 3, N4-A es circular con `verification.md:43`; redactado como "a declared `Review` value, including `none`, also authorizes `hive-review-security` when the change crosses a boundary" más "or one a declared `Review` value gives" en `:43`, con N4-B como alternativa. Pasa a la pregunta de ruta.
  - Orden de precedencia confirmado por el revisor: instrucción más estrecha, elección anterior de la sesión, valor declarado, pregunta.
  - **Cobertura:** guía distribuida revisada en `947f0ffd3563` y re-revisada en `24b06c7ba5fc`; manager Go revisado en `947f0ffd3563`. Las correcciones posteriores aplican propuestas de los revisores, salvo N4, que decide el usuario.
- 2026-10-17: N4-A y D3-A confirmadas; build iniciado en `feat/lifecycle-declared-delivery-review` desde `e484b47b`. T2 delegada a `hive-build-backend` (modelo configurado, no observado) en paralelo con T1 en el hilo principal (claude-fable-5-1, observado).
- 2026-10-17, T1: `hive-verify-task` (modelo `verifier` configurado, no observado; 108 s, ~118k tokens): AC1–AC3 `met`. Cuatro defectos de baja severidad y `reporting.md:79` incompleto, corregidos en una ronda y re-verificados por el mismo verificador (32 s): `met`; una ambigüedad nueva en `flow-build:20` ("or" como "basta uno") corregida con su redacción, y el segundo caso de `:34` añadido; verificado por `rg` en el hilo principal. `global.md`: 29 638 → 29 811 bytes (+173). Commit `c932557f`.
- 2026-10-17, T2: implementada por `hive-build-backend` (modelo configurado, no observado; 319 s, ~95k tokens) con RED observado (`expected "direct-base", found "automatic"`) y `go test ./...` verde; también actualizó `TestHiveSectionUnknownKeyAndOptionalValues`. `hive-verify-task` (77 s, ~94k tokens): AC4 `met`; señaló que el patrón `-run` del plan omitía dos pruebas, corregido en T2. Commit `9ec88c4e`.
- 2026-10-17, T3: `hive-verify-task` (54 s, ~89k tokens): AC5 `met`; cuatro observaciones de redacción aplicadas (disparador por fallo del chequeo, push de actualización cubierto por la autorización del merge, `origin/<branch>` en lugar de `<head>`, "without changing the branch"), diseño actualizado. Commit `60391f51`.
- 2026-10-17, T4: `hive-verify-task` (68 s, ~88k tokens): AC6 y AC7 `met`; dos observaciones P3 aplicadas ("the authorization a declared `Review` value gives" en `verification.md:43`; frase final del revisor en dos oraciones con "crosses"). Confirmó que `TestRepositoryCatalogueInstructionReferences` valida los locators `skill:` de los agentes. Commit `841aee43`.
- 2026-10-17, T5 y T6: editadas en el hilo principal; verificaciones lanzadas en paralelo.
- 2026-10-17, T5: `hive-verify-task` (21 s, ~75k tokens): AC8 `met`, sin defectos. Commit `265eda02`.
- 2026-10-17, T6: `hive-verify-task` (27 s, ~80k tokens): AC9 `met`, sin defectos; `astro check`, `astro build` y `biome check` del sitio no se corrieron. Commit `5269a6c0`.
- **Compuertas del candidato final (`5269a6c0`):** `go vet ./...` sin hallazgos; `go test ./...` con salida 0; 23 pruebas de Python OK; `global.md` 29 638 → 29 811 bytes (+173); relectura de `delivery-decisions.md:5-36` sin copias parciales ni contradicciones.
- **Despliegue:** binario reconstruido con `go build` (T2 cambia el manager); `hive update --out` + `hive apply --plan` desde `5269a6c0`, transacción `a00dbaa272a5…`, 203 archivos en Claude, Codex, OpenCode, Cursor, Grok y Pi. Comprobado el texto nuevo en `~/.claude/CLAUDE.md` y en la skill `git-workflow` instalada.
- **Estado:** desplegado en local para tus pruebas. Siguiente paso: tú lo pruebas en sesiones nuevas y decides el merge a `development`.
- **Límites:** no se verificó si GitHub reporta `BEHIND` sin la protección "require branches up to date" (la regla no depende de ello); las compuertas del sitio no se corrieron; R1 y R2 solo se observan en un repo que declare `Delivery` y `Review`.
