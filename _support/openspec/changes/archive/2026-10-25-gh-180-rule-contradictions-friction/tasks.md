# Tareas

Base: `70d2bb13` (`development`, 2026-10-25). Worktree `.claude/worktrees/gh-180`, rama `fix/gh-180-rule-contradictions`.

**Cómo se comprueba cada tarea**
- **Frases distintivas:** cada cláusula nueva tiene una frase distintiva, con conteo 0 en la base y 1 después (`rg -c` en el archivo hogar). Una frase quitada da 0 después. El implementador anota las frases elegidas en esta tarea, y el verificador las cuenta y lee el texto contra [diseño](design.md#enfoque).
- **Pruebas en cada tarea:** toda tarea corre `go test ./tests/content/...` (tamaños y enlaces de la guía global) y `go test ./tooling/management -run TestRepositoryCatalogueInstructionReferences` (enlaces de recursos).
- **Anclas:** ninguna prueba las comprueba. Cada encabezado que se mueve o renombra se busca con `rg` en `content`, `site` y `_support/docs`.
- **Cobertura de criterios:** la tarea que cierra un criterio comprueba todas sus partes, aunque otra tarea anterior haya hecho alguna.

## T1 — Entrega y revisión: un solo hogar

- [x] `delivery-decisions.md`:
Correcciones posteriores a la re-revisión (sin otra ronda de revisión, como fija `review-and-gates.md`), commit `24148f9f`. Re-verificación (~75 s): AC1 y AC6 cumplidos. Dos ajustes de redacción en `35fc9316` (pedir un mecanismo local; ancla del puntero de promoción). Re-verificación (~36 s): AC1 y AC6 cumplidos; el puntero contaba mal los párrafos y en el commit siguiente pasa a nombrar la regla, cambio de redacción que comprobó el orquestador con `rg` (la regla sigue en `delivery-decisions.md:5`).
Reabierta por la corrección de la revisión dedicada (`/code-review`), commit `4880b4c1`. Re-verificación (~76 s): AC1 y AC6 cumplidos.
  - Fija el momento de la revisión dedicada por modo (D4-A), con la revisión acotada después de la validación y la regla de promoción «not reviewed before their integration».
  - Con `automatic`, la pregunta de revisión siempre se hace y su recomendada es un mecanismo (D10-A); las opciones valen con cualquier respuesta a la otra pregunta (AC16).
  - Un cambio mecánico no recibe la pregunta de revisión.
  - `automatic` solo se descarta por una regla del proyecto que se nombra (D5-A).
  - Una frase distingue `direct-base` elegido del declarado.
  - Las definiciones de `UI verification` quedan aquí, con el efecto escrito por modo.
  - `:34` exceptúa `automatic` de la recomendación «No dedicated review» y apunta a D10-A.
  - `:12,18,22`: con `hold`, la entrega y el registro sin versionar se reportan como recordatorio, no como pendiente.

  `project-settings.md` define el punto de cierre posterior (D6-A, con la condición de la guía del proyecto) y apunta para `UI verification`. `plan-revision.md:11` apunta a la excepción de `review-and-gates.md`.

Verificado por `hive-verify-task` (~119 s, modelo configurado, no observado): AC1 y AC6 cumplidos (`delivery-decisions.md:10,11,18`). Commit `4886c1a3`. El verificador señaló dos puntos menores, corregidos en un commit adicional que se vuelve a verificar: recomendar un mecanismo también cuando `automatic` es opción de la misma ronda, y conservar la advertencia de `none`. También señaló `ui-gate.md:19`, una copia de las definiciones de `UI verification`, que pasa a T4. Re-verificación (~86 s): AC1 y AC6 siguen cumplidos con el commit `3980deca`.

**Closes:** AC1, AC6.
**Checkpoint:** C1.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:5,9-12,18,22,26,34,36`; `content/skills/flow-plan/references/project-settings.md:9`; `content/skills/flow-plan/references/plan-revision.md:11`.
**Execution:** hilo principal. Estas reglas son el hogar de las que siguen, y el brief sería más largo que el cambio.
**Test approach:** check.
**Changes:** según [diseño](design.md#entrega-y-revisión-delivery-decisionsmd-hogar) y [punto de cierre](design.md#punto-de-cierre-posterior-project-settingsmd-hogar-d6-a).
**Verification:**
- `rg -c "such as a visual check of a changed screen" content` da 0 (AC1).
- La frase que exige nombrar la regla del proyecto da 1 en `delivery-decisions.md` (AC1).
- La frase de `direct-base` elegido frente al declarado da 1 en `delivery-decisions.md` (AC6).
- `rg -c "reported as pending" content/skills/flow-plan/references/delivery-decisions.md` da 0.
- Las pruebas de la convención pasan.

## T2 — Preguntas dependientes y espacio en `flow-plan`

- [x] `questions.md` agrega la condición de preguntas dependientes y limita la regla de conjunto en secuencia (AC16), con la excepción de «silence grants nothing» que pide D8-A. `flow-plan/SKILL.md` apunta a la línea de pausa sin pedir «Pausa» (AC9) y lleva el resumen de entrega del plan a `plan-format.md`, «Hand off a ready plan». También corrige los hallazgos menores de T2 en la tabla de [diseño](design.md#hallazgos-menores-destino): readiness con puntero, texto previo a #169, glosa de los AC de la tarjeta, la base al planear, el criterio del ticket descartado y la revisión de planes no guardados. `plan-format.md:17` registra la prueba en vivo innecesaria como razón en el reporte, no como oferta. El puntero de delegación de `flow-plan:56` va a `delegation.md`, «Split and announce work».

Ronda 1 de `hive-verify-task` (~116 s): AC16 cumplido; AC9 rechazado porque `flow-plan/SKILL.md:22` no enlazaba `questions.md#pause-and-handle-the-answer`. Además, el puntero de readiness perdía la regla de que todo criterio lo cierra alguna tarea. Arreglo de ambos (enlace y `plan-format.md:28`) en el commit `2d208037`; la tarea es el commit `69e666c2`. Ronda 2 (~39 s): AC9 cumplido (`flow-plan/SKILL.md:22`, `questions.md:72`). Frase de AC16: «depend on the answer to another open question».

**Closes:** AC9, AC16.
**Checkpoint:** C2.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/questions.md:23,25-31`; `content/skills/flow-plan/SKILL.md:12,22,40,48,56,68,74`; `content/skills/flow-plan/references/plan-format.md:17,26,136`, sección nueva «Hand off a ready plan»; `content/skills/flow-plan/references/plan-review.md:3`.
**Execution:** hilo principal, acoplado al texto de `flow-plan`.
**Test approach:** check.
**Verification:**
- La frase nueva de preguntas dependientes da 1 en `questions.md`, y «even when a later step depends on an earlier one's result» queda limitada a las autorizaciones de un objetivo ya decidido (lectura del verificador) (AC16).
- `rg -c "Pausa" content/skills/flow-plan/SKILL.md` da 0, y `flow-plan/SKILL.md` enlaza `questions.md#pause-and-handle-the-answer` (AC9).
- `wc -c content/skills/flow-plan/SKILL.md` da 15 500 bytes o menos.
- `rg -n "#hand-off-a-ready-plan" content` resuelve a la sección nueva.
- `rg -c "the offered walk and user response" content` da 0.
- Las pruebas de la convención pasan.

## T3 — `flow-close`: espacio, finales sin merge y borrado

- [x] «Close the current walkthrough» pasa a `flow-close/references/walkthrough-close.md` con un puntero. «When to run it» define `hold` (punto de cierre, con recordatorio), el PR que espera revisión humana y el final después de una promoción. El paso 5 limpia solo ramas de PRs del usuario autenticado, y una rama remota se borra solo con la cabeza remota igual al `headRefOid` o ancestro de la base. `reporting.md:102` apunta al punto de cierre de `project-settings.md`. «Close question» y `incidental-findings.md` ofrecen resolver los hallazgos en esta obra y dejan la pregunta de qué sigue para después cuando dependa (D11-A). Se corrigen los hallazgos menores de T3: borrado de ramas repetido, una o dos escrituras de memoria, y el paso 1 confirma el movimiento del ticket.
Correcciones posteriores a la re-revisión (sin otra ronda de revisión, como fija `review-and-gates.md`), commit `8f110f33`. Re-verificación (~71 s): AC13 cumplido. Ajuste en `5345f483`: el abandono deja el movimiento del ticket como recordatorio, porque la regla global no autoriza esa escritura. Re-verificación (~22 s): AC13 cumplido (15 473 bytes).
Reabierta por la corrección de la revisión dedicada (`/code-review`), commit `52d6bac5`. Re-verificación (~40 s): AC13 cumplido (`flow-close/SKILL.md` 15 420 bytes).

Verificado por `hive-verify-task` (~96 s): AC13 cumplido (`flow-close/SKILL.md` 15 040 bytes, `flow-plan/SKILL.md` 13 956). Commit `577cb758`. Dos defectos menores corregidos en `f7b4e220`, que se vuelve a verificar: el paso 5 había perdido «merged into the base», y faltaba el caso del registro con revisión humana. `flow-close/SKILL.md` queda en 15 223 bytes. Re-verificación (~46 s): AC13 cumplido, sin defectos.

**Closes:** AC13.
**Checkpoint:** C3.
**Depends on:** T1, T2.
**Locations:** `content/skills/flow-close/SKILL.md:14,18-20,26,28,34-56,58-64,79,80,88,92` (`:80` reporta el registro sin versionar de `hold` como recordatorio); `content/skills/flow-close/references/walkthrough-close.md` (nuevo); `content/skills/flow-close/references/reporting.md:74,102`; `content/skills/flow-close/references/incidental-findings.md:9`; enlaces entrantes a `flow-close/SKILL.md#close-the-current-walkthrough` (`rg` en `content`, `site`, `_support/docs`).
**Execution:** hilo principal.
**Test approach:** check.
**Verification:**
- `wc -c content/skills/flow-close/SKILL.md` da 15 500 bytes o menos, y `wc -c content/skills/flow-plan/SKILL.md` sigue en 15 500 o menos (AC13).
- `rg -n "close-the-current-walkthrough" content site _support/docs` no deja enlaces rotos.
- `rg -c "reported as pending, like the code" content/skills/flow-close/SKILL.md` da 0.
- Las pruebas de la convención pasan.

## T4 — Verify, walkthrough, build y delegación

- [x] `flow-verify`, `review-and-gates.md` (incluida la línea 16 con su excepción y «candidato») y `ui-gate.md:38,42` aplican el momento de `delivery-decisions.md`, con la revisión acotada después de la validación. La prueba en vivo considerada innecesaria pasa a ser una línea del reporte (AC7). `flow-walkthrough` (`:12-14,24,30,34,48,51`) captura y reporta la primera vista sin pregunta bajo `automatic` y bajo delegación desatendida, y agrega la parada de `direct-base` en un repo de código (D5-A). `flow-build` aplica D7-A, exime de la pregunta de revisión a un cambio mecánico y apunta a `delegation.md` (D9-A), que queda como único criterio con «Launch and wait» limitado al padre. `ui-gate.md` dice que el hilo principal sirve la revisión base. Se corrigen los hallazgos menores de T4.
Correcciones posteriores a la re-revisión (sin otra ronda de revisión, como fija `review-and-gates.md`), commit `db3fbd17`. Re-verificación (~81 s): AC3, AC4 y AC7 cumplidos.
Reabierta por la corrección de la revisión dedicada (`/code-review`), commit `f17f907d`. Re-verificación (~92 s): AC3, AC4 y AC7 cumplidos.

Verificado por `hive-verify-task` (~209 s): AC3, AC4 y AC7 cumplidos (`delivery-decisions.md:34,42-44`, `review-and-gates.md:16,22`, `flow-build/SKILL.md:24,64`, `flow-walkthrough/SKILL.md:13`). Commit `8910c8f9`. La mención que queda en `flow-verify/SKILL.md:10` («integrated as the delivery mode sets») habla de la integración, no de la revisión. Arreglos en `f2a440dc`, que se vuelve a verificar: la regla desatendida del walkthrough queda limitada a la primera vista, `flow-build` nombra el caso desatendido y la regla del verificador no nombra modelos de un host (`integrations/` no se despliega, así que no se cita `agent-profiles.json`). Re-verificación (~110 s): AC3, AC4 y AC7 siguen cumplidos, sin defectos.

**Closes:** AC3, AC4, AC7.
**Checkpoint:** C4.
**Depends on:** T1.
**Locations:** `content/skills/flow-verify/SKILL.md:22,28,30,46,54,62`; `content/skills/flow-verify/references/review-and-gates.md:9-10,16,22`; `content/skills/flow-verify/references/ui-gate.md:19,22,26,36,38,42` (`:19` copia las definiciones de `UI verification` y pasa a ser un puntero); `content/skills/flow-walkthrough/SKILL.md:12-14,24,30,34,48,51`; `content/skills/flow-build/SKILL.md:12,24,28,38,50,62,64,68`; `content/skills/flow-build/references/plan-execution.md:23`; `content/skills/flow-build/references/browser-automation.md:9,37,39`; `content/skills/flow-research/references/delegation.md:5,11,41-65`.
**Execution:** hilo principal. Los cambios dependen de las frases de T1, y el brief sería más largo que el cambio.
**Test approach:** check.
**Verification:**
- **AC3:**
  - `rg -c "the point the delivery mode sets|as the delivery mode sets" content` da 0, o cada mención enlaza la sección de `delivery-decisions.md`.
  - La frase del momento por modo da 1 en `delivery-decisions.md`.
  - La frase de D10-A (nunca recomendada) da 1 en `delivery-decisions.md`.
  - La excepción de `review-and-gates.md:16` da 1, y `plan-revision.md:11` ya no pregunta sin excepción.
  - `ui-gate.md` nombra la revisión acotada al diff después de la validación.
- **AC4:**
  - La frase de D7-A da 1 en `flow-build/SKILL.md`.
  - La parada de `direct-base` en un repo de código da 1 en `flow-walkthrough/SKILL.md` o en `delivery-decisions.md`.
- **AC7:**
  - `rg -c "when it seems unnecessary, offer it" content` da 0.
  - `review-and-gates.md:22` no ofrece «Skip / Run».
  - La exención de la pregunta de revisión para un cambio mecánico da 1 en `flow-build/SKILL.md` y en `delivery-decisions.md`.
- **Tamaño y pruebas:** `wc -c` de `flow-build`, `flow-verify` y `flow-walkthrough` da menos de 16 000 bytes, y las pruebas de la convención pasan.

## T5 — Guía global, desatendido, git, deploy y umbral de research

- [x] `global.md:20,29,33,34,35,84,117` aplica los punteros y la redacción de [diseño](design.md#guía-global-desatendido-git-y-deploy), incluidos la autorización de los dos movimientos del ticket y la opción de resolver hallazgos. `unattended-delegation` (`:3,8,16,24`) aplica D8-A con la redacción completa y apunta a la regla de la primera vista y a `git-workflow`. `git-workflow` incluye la excepción de `--force-with-lease` y el caso sin CI. `flow-deploy` apunta al punto de cierre y a la regla de revisión de la promoción, y unifica los estados de aceptación. `artifact-layout.md:43` apunta a `flow-close`. `flow-research/SKILL.md:26` y `security-advisory/SKILL.md:19` apuntan a `delegation.md`.
Correcciones posteriores a la re-revisión (sin otra ronda de revisión, como fija `review-and-gates.md`), commit `85624a6b`. Re-verificación (~88 s): AC11 cumplido.
Reabierta por la corrección de la revisión dedicada (`/code-review`), commit `781fdaf7`. Re-verificación (~86 s): los seis criterios cumplidos.

`global.md` pasa de 38 338 a 38 996 bytes (+658). Verificado por `hive-verify-task` (~164 s): AC2, AC5, AC8, AC10, AC11 y AC17 cumplidos. Commit `0552f163`. Arreglos en `16d424a4`, que se vuelven a verificar: la excepción de `--force-with-lease` queda solo en `git-workflow` y `unattended-delegation` apunta a ella; `global.md:34` recupera el «unless» del punto de cierre posterior. Además se alinea «the user's earlier work» en `flow-close/SKILL.md:86`, archivo de T3: queda en 15 239 bytes y AC13 se vuelve a comprobar. Re-verificación (~86 s): AC5, AC11 y AC13 siguen cumplidos, sin defectos.

**Closes:** AC2, AC5, AC8, AC10, AC11, AC17.
**Checkpoint:** C5.
**Depends on:** T1, T2, T3, T4.
**Locations:** `content/guidance/global.md:20,29,33,34,35,84,117`; `content/skills/unattended-delegation/SKILL.md:3,8,16,24`; `content/skills/git-workflow/SKILL.md:24,40`; `content/skills/flow-deploy/SKILL.md:10,14,25`; `content/skills/workspace-conventions/references/artifact-layout.md:43`; `content/skills/flow-research/SKILL.md:26`; `content/skills/security-advisory/SKILL.md:19`.
**Execution:** hilo principal. `global.md` es la capa que siempre se carga, y su redacción necesita el juicio del orquestador.
**Test approach:** check.
**Verification:**
- **AC2:** la regla de la primera vista bajo `automatic` y desatendida da 1 en `flow-walkthrough/SKILL.md`, y `flow-build/SKILL.md` y `unattended-delegation/SKILL.md` la enlazan.
- **AC5:**
  - `rg -c "completion point" content/skills/flow-plan/references/project-settings.md` da al menos 1.
  - `global.md:34`, `reporting.md:102`, `flow-close/SKILL.md` y `flow-deploy/SKILL.md` enlazan `project-settings.md`.
  - La frase que autoriza los dos movimientos del ticket (intermedio y final) da 1 en `global.md`.
  - `flow-close` «When to run it» nombra `hold`, el PR con revisión humana y la promoción.
- **AC8:**
  - `rg -c "goes to a child by default|at the threshold the global delegation rule sets" content` da 0.
  - `flow-build`, `flow-plan`, `flow-research` y `security-advisory` enlazan `delegation.md#split-and-announce-work`.
- **AC10:** la frase de D8-A da 1 en `unattended-delegation/SKILL.md:3` y `:8`, y en `global.md:35`. `global.md:20` nombra la excepción.
- **AC11:**
  - `rg -c "or a worktree with uncommitted or untracked files" content` da 0.
  - `artifact-layout.md` enlaza `flow-close/SKILL.md#clean-up-after-a-merge`.
  - La frase del límite por autor y la de cabeza remota dan 1 en `flow-close/SKILL.md`.
  - `rg -c "lease" content/skills/git-workflow/SKILL.md` da al menos 1, y la frase del caso sin CI da 1.
- **AC17:** la opción de resolver en esta obra da 1 en `incidental-findings.md`, la frase de la pregunta de qué sigue en un turno posterior da 1 en `incidental-findings.md` o en «Close question» de `flow-close`, y `flow-close` (Close question) y `global.md:33` enlazan la regla.
- **Pruebas:** `go test ./tests/content/...` registra el nuevo tamaño de `global.md` para el PR, y las pruebas de la convención pasan.

## T6 — Roles y skills de soporte

- [x] `diagnosis.md` reparte el trabajo: el hijo reúne evidencia de solo lectura y el hilo principal corre los experimentos. `hive-read-state` acepta consultas al tracker. `hive-verify-change` gana el modo de solo grabar. `hive-review-ux` lee browser automation en el paso 1, y las colas de los roles de verificación y observación ganan las dos prohibiciones. `flow-research/SKILL.md:20` coincide con la descripción de `hive-read-state`. Se corrigen los hallazgos menores de T6.

Implementada por `hive-write-docs` (~83 s); el orquestador revisó el diff y ajustó el modo de solo grabar de `hive-verify-change` para que no exija comprobaciones previas. `plan-format.md` no tiene una fase «abandonada»: `workspace-archive` acepta `Completed` y `cancelled`. Verificado por `hive-verify-task` (~209 s): AC12 cumplido. Commit `2aa3d39e`. Dos filas quedaron a medias y se arreglan en `f0d541e5`, que se vuelve a verificar: `hive-refute-claim.md:16` no permitía refutar afirmaciones que no fueran de código, y solo 2 de los 10 roles de verificación y observación tenían las dos prohibiciones. Re-verificación (~92 s): AC12 sigue cumplido, sin defectos. Frases: «the main thread runs the experiments», «tracker query by project and states», «in recording-only mode», «1. Before driving a browser, read», «an `ask` entry for a clone», «before the first `uv run`, check the dependency».

**Closes:** AC12.
**Checkpoint:** C6.
**Depends on:** T4, T5.
**Locations:** `content/skills/flow-research/references/diagnosis.md:7,9`; `content/skills/flow-research/SKILL.md:20`; `content/agents/quality/hive-read-state.md:3`, paso 1; `content/agents/quality/hive-verify-change.md`; `content/agents/review/hive-review-ux.md:14,16,20`; `content/agents/review/hive-refute-claim.md:16`; `content/agents/review/hive-review-plan.md:15`; `content/agents/design/hive-design-architecture.md:12`; `content/agents/ops/hive-build-infra.md:17`; `content/agents/development/hive-build-data.md:10`; `content/skills/{harness-audit,workspace-conventions,workspace-archive,engram-init-workspace,flow-security,flow-report,starlight-docs-site}/SKILL.md`; `content/skills/harness-audit/references/{native-tools,project-layout}.md`; `content/skills/flow-research/references/backlog-report.md:10`.
**Execution:** delegada a `hive-write-docs`. Las unidades tienen interfaz fija, no se superponen con otras escrituras, y el diseño y la tabla de hallazgos son la fuente. Se pasan `AGENTS.md` (Guidance structure) y `_support/docs/architecture/instruction-resources.md`.
**Test approach:** check.
**Verification:**
- `rg -c "tracker query" content/agents/quality/hive-read-state.md` da al menos 1.
- La frase del modo de solo grabar da 1 en `hive-verify-change.md`.
- En `hive-review-ux.md`, la lectura de browser automation aparece antes del primer paso de captura.
- `diagnosis.md` nombra quién corre los experimentos.
- `ui-gate.md` dice que el hilo principal construye y sirve la revisión base (hecho en T4, comprobado aquí para AC12).
- `go test ./tooling/management -run TestRepositoryCatalogueInstructionReferences` pasa.

## T7 — Documentación y portal

- [x] Las páginas de `site/`, el README, `llms.txt`, `CONTRIBUTING.md` y `_support/docs/architecture/` describen el comportamiento final sin copiar reglas, según [diseño](design.md#documentación-y-portal). Antes de editar, se registra si `pnpm --dir site check` y `pnpm --dir site lint` pasan en la base, para separar regresiones de fallos previos. Los pasajes copiados se reemplazan por una descripción con enlace: `flows/build.md:48` (prueba costosa y clases de efectos), `flows/plan.md:69,74` y `flows/research.md:47-55` (mecánica de las preguntas). Se corrigen los hallazgos menores de T7, y se comprueba la tabla completa de hallazgos menores (AC15). AC14 se cumple con el delta de [specs](specs/documentation-site/spec.md) presente; la especificación vigente cambia al archivar.

Base del sitio registrada el 2026-10-25 en `70d2bb13`: `pnpm --dir site check` sale con 0 errores, 0 avisos y 0 sugerencias, y `pnpm --dir site lint` revisa 14 archivos sin cambios; los dos terminan con código 0.

Implementada por `hive-write-docs` (~226 s); el orquestador revisó el diff y quitó una mención repetida de `hive releases` en el README. Commit `39322df2`. En el worktree, `site/node_modules` es un enlace al checkout principal; los comandos del sitio corren con `--config.verify-deps-before-run=false`, que no instala nada. Ronda 1 de `hive-verify-task` (~313 s): AC15 cumplido (45 filas «Corregir» corregidas, 5 descartes con razón). AC14 rechazado porque `flows/build.md:45` todavía decía que se pregunta siempre por una tarea añadida después de la revisión. Arreglo en `a2ee6173`, junto con «UI gate» en `build.md:83` y `direct-base` elegido en los momentos de revisión de `verify.md:28` y `build.md:50`. Ronda 2 (~111 s): AC14 cumplido, sin contradicciones en el sitio, el README ni `llms.txt`.

**Closes:** AC14, AC15.
**Checkpoint:** C7.
**Depends on:** T1, T2, T3, T4, T5, T6.
**Locations:** `site/src/content/docs/index.mdx:48,97,104-139`; `site/src/content/docs/getting-started.md:31`; `site/src/content/docs/skills.md:21-29`; `site/src/content/docs/flows/build.md:46-50,74,83,89-122`; `flows/plan.md:63-74`; `flows/research.md:47-55`; `flows/close.md:17,62,69,72,78`; `flows/deploy.md:6,39`; `flows/verify.md:25,26,28`; `README.md:344-364,447,471`; `llms.txt:12-36`; `CONTRIBUTING.md:136-137`; `_support/docs/architecture/repository-and-distribution.md:3`.
**Execution:** delegada a `hive-write-docs`, con el diseño, el diff de T1–T6, la regla del sitio (`AGENTS.md`, README and llms.txt) y la lista de comprobaciones siguiente en el brief.
**Test approach:** check.
**Verification:**
- **Frases que deben dar 0:**
  - `rg -c "A full local suite for a change that is not mechanical" site`.
  - `rg -c "whether the composition is right" site/src/content/docs/flows/build.md`, o la frase queda condicionada a un modo con parada (AC2).
  - `rg -c "Pausa|\"Pause\"" site/src/content/docs/flows/plan.md` (AC9).
  - `rg -c "Esperando tu decisión" site`.
  - `rg -c "delivered and reviewed" site/src/content/docs/flows/build.md`, o la frase exime los cambios mecánicos (AC7).
- **Comportamientos que cambian (lectura del verificador contra el diseño):**
  - `index.mdx:97` y `skills.md:29` describen D8-A.
  - `flows/plan.md` describe las preguntas dependientes en turnos posteriores (AC16).
  - `flows/close.md:17,62,72`, `flows/deploy.md:6,39` y `README.md:355` describen D6-A.
  - `flows/close.md:69` describe el límite de borrado.
  - `flows/verify.md:26,28` y `flows/build.md:50,74` describen D4-A y D7-A.
- **Presencia:**
  - `rg -c "flow-walkthrough" site/src/content/docs/skills.md` da al menos 1.
  - Tres búsquedas separadas, `rg -c "hive voice"`, `rg -c "hive project set"` y `rg -c "hive releases"`, dan al menos 1 cada una en `README.md` y en `llms.txt`.
  - `rg -n "pnpm (--dir site )?build" CONTRIBUTING.md` aparece en una línea anterior a `wrangler deploy`.
- **Reglas copiadas:** `flows/build.md:48`, `flows/plan.md:69,74` y `flows/research.md:47-55` ya no reproducen el texto de la regla y enlazan la skill (lectura del verificador).
- **AC15:** el verificador recorre la tabla de hallazgos menores y comprueba cada «Corregir».
- **Sitio:** la construcción del sitio pasa (compuerta compartida).

## Verificación y revisión

**Puntos de control**
- **C1** — T1: cambia la semántica de valores persistidos de `## Hive` (`direct-base`, punto de cierre), y T3, T4, T5 y T7 dependen de ella.
- **C2** — T2: de ella dependen T3, T5 y T7.
- **C3** — T3: cambia qué ramas puede borrar `flow-close` (permisos).
- **C4** — T4: de ella dependen T5, T6 y T7.
- **C5** — T5: cambia `global.md`, la capa que siempre se carga, y las reglas de borrado y de activación desatendida.
- **C6** — T6: cambia lo que aceptan los roles distribuidos (contrato).
- **C7** — T7: último grupo, sin dependientes.

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Verificación por punto de control | En cada C<n>, en el worktree | `hive-verify-task` por criterio, con veredicto por AC | Cubierta por la implementación |
| Suite completa | Una vez, sobre el candidato final (no hay CI) | `go test ./...` pasa | Cubierta por la implementación |
| Construcción del sitio | Sobre el candidato final | `pnpm --dir site build`, `pnpm --dir site check` y `pnpm --dir site lint` pasan, o fallan igual que en la base registrada en T7; las páginas `flows/verify/` y `skills/` se generan en `site/dist` | Cubierta por la implementación |
| Comprobación con render | Cambio de texto sin cambio de layout (la puerta de UI lo exime) | El implementador abre `flows/verify/`, `flows/build/` y `skills/` en el `pnpm --dir site preview` local y confirma que se ven | Cubierta por la implementación |
| Barrido de consistencia | Antes del PR | Por cada regla cambiada, `rg` en `content`, `site`, `_support/docs/architecture` y `tests/pilot`, con la lista de lugares en el PR | Cubierta por la implementación |
| Revisión dedicada | Sobre el candidato local, después de la suite y antes del push (D4-A) | `/code-review` de Claude Code; corregir P0–P1 y P2 locales | `Review: /code-review` declarado |
| Base que se movió | Antes del merge, si `development` avanzó | Integrar la base y volver a correr `go test ./...` y la construcción del sitio | Cubierta por `Delivery: automatic` |

## Estado de la revisión y avance

- **Revisión del plan, ronda 1, sobre `43a18aabf906`:** consistencia de la guía (8 bloqueantes, 3 no bloqueantes), entrega y autorización (3 bloqueantes, 3 no bloqueantes), y documentación y portal (1 bloqueante, 2 no bloqueantes). Todos se aceptaron y se aplicaron en esta revisión; el detalle está en [diseño](design.md#decisiones-menores). Duración: unos 4, 3 y 2 minutos; 164 k, 143 k y 123 k tokens.
- **Decisiones nuevas durante la revisión:** D10-A y D11-A, con AC17.
- **Re-revisión (única ronda), revisor de consistencia:** confirmó resueltos los hallazgos anteriores y el chequeo de orden (17 criterios cubiertos, C1–C7 sin huecos). Encontró 4 brechas: `delivery-decisions.md:34` contra D10-A, la redacción de «pendiente» para `hold` en `delivery-decisions.md:12,18,22` y `flow-close/SKILL.md:80`, y dos líneas de verificación de T5 (AC17, turno posterior; AC5, autorización de los dos movimientos). Se aplicaron tal como las propuso, sin otra ronda. Duración: unos 61 s, 184 k tokens.
- **Cobertura:** consistencia de la guía (dos rondas), entrega y autorización (una ronda) y documentación y portal (una ronda). Las correcciones posteriores a la ronda de cada revisor son sus propias propuestas, o decisiones del usuario (D10-A, D11-A) que revisó el de consistencia.
- **Evidencia de sesiones:** registrada en [diseño](design.md#evidencia-de-sesiones).
- **Revisión dedicada (`/code-review` de Claude Code, esfuerzo medio, sobre el candidato local `901b82fe`):** 10 hallazgos. Se corrigen 9 en `4880b4c1` (T1), `52d6bac5` (T3), `f17f907d` (T4) y `781fdaf7` (T5): mecanismos que solo corren sobre el PR, orden respecto de la suite sin CI, registro de lo revisado para la promoción, quién cierra tras el merge con revisión humana, el paso 1 bajo `hold`, delegación de búsquedas, `ui-gate.md:36`, nueva revisión bajo `hold` y la fila `interactive`. Se descarta 1: la cola repetida en los roles es intencional porque un hijo puede no cargar la guía global, y queda como vigilancia (conteo 1).
- **Re-revisión (única ronda, diff `901b82fe..781fdaf7`):** 9 observaciones locales de redacción, corregidas sin otra ronda en `24148f9f` (T1), `8f110f33` (T3), `db3fbd17` (T4) y `85624a6b` (T5). `flow-close/SKILL.md` queda en 15 479 bytes.
- **Integración:** [PR #181](https://github.com/JhonHawk/tricell-hive/pull/181) mergeado en `development` (`31d856d0`) el 2026-10-25, tras `go test ./...` y la construcción del sitio en verde. #180 cerrado.
- **Siguiente paso:** ninguno.
