# Tareas

Base: `b7277d82` (`development`, 2026-10-25). Worktree `.claude/worktrees/gh-168`, rama `feat/gh-168-verification-checkpoints`.

**Convención de las comprobaciones.** Como en gh-157 y gh-158: una frase distintiva por cláusula nueva, con conteo 0 en la base antes de escribirla y 1 después, más la lectura del verificador contra [diseño](design.md#enfoque); las frases viejas que deben desaparecer dan 0 después. El build de este cambio se verifica con las reglas vigentes, tarea por tarea.

## T1 — Puntos de control en el formato y la revisión del plan

- [x] `plan-format.md` define el campo, la lista con razones, el criterio de riesgo, el orden de dependencias y la compatibilidad; `plan-review.md` agrega los dos hallazgos bloqueantes.

**Closes:** AC1, AC2.
**Depends on:** ninguna.
**Locations:** `content/skills/flow-plan/references/plan-format.md` (sección «Acceptance criteria and task markers», fila `[?]` de la tabla de marcadores para `waiting:` y fila `[/]` para `reopened by the fix of T<n>` y para `blocked:` antes de reportarse, plantilla de tarea en «Adaptable template»); `content/skills/flow-plan/references/plan-review.md:22` (párrafo de hallazgos de toda asignación).
**Execution:** delegada a `hive-write-docs`: el diseño fija el contenido.
**Test approach:** check.
**Changes:** el campo `**Checkpoint:** C<n>` y las líneas `waiting: T<n>` y `reopened by the fix of T<n>` quedan sin traducir, como `Closes:`; el criterio de riesgo y la compatibilidad («sin el campo, cada tarea es su propio punto») se escriben una sola vez aquí; `plan-review.md` remite al criterio sin repetir sus condiciones.
**Verification:** `Checkpoint` aparece en la sección de criterios y en la plantilla; dan 1 en `plan-format.md` las frases de la lista de puntos con su razón, del criterio de riesgo, de «solo depende de tareas anteriores», de la compatibilidad, de `waiting:`, de `reopened by the fix of` y de `blocked:` bajo `[/]`; las dos frases bloqueantes dan 1 en `plan-review.md`; lectura contra el diseño.

## T2 — Ejecución, arreglos y commits por punto de control

- [x] `plan-execution.md` lanza las verificaciones al alcanzar un punto, aplica D5-A, `waiting:`, la reapertura y el orden de los arreglos, y commitea según D4-A y D8-A.
Verificado por `hive-verify-task` (~96 s): AC3, AC4, AC5 y AC8 cumplidos (`plan-execution.md:19,28-33,39`). El orquestador aplicó después dos observaciones del verificador que no tocan las frases de los criterios: una tarea en espera sigue esperando si la reelaboración de su dependencia no cambia sus archivos, y la dependiente de una tarea retirada va al usuario (paso 6). Sin re-verificación por ser aclaraciones fuera de los criterios. Commit `df0cb56c`.

**Closes:** AC3, AC4, AC5, AC8.
**Depends on:** T1 (nombres del campo, de `waiting:` y del criterio).
**Locations:** `content/skills/flow-build/references/plan-execution.md:17-30` (paso 1, párrafo «Do not wait idle», «Commit verified tasks»).
**Execution:** delegada a `hive-write-docs`, con `design.md` como fuente.
**Test approach:** check.
**Changes:**
- Paso 1: al reportar, `[?]`; lanzar las verificaciones cuando el punto se alcanza, como define el diseño, un verificador por tarea, juntos. Re-verificaciones fuera de los puntos: en el momento que fija su propia regla (la puerta del candidato aceptado para una corrección del walkthrough), sin esperar un punto.
- «Commit verified tasks»: la frase «this is the one exception to one verified task per commit» pasa a nombrar las dos excepciones (arreglo después del commit y commit conjunto).
- Párrafo «Do not wait idle»: reemplazar «a task that depends on it waits for `[x]`», «When several tasks wait for verification and their **Locations** share no file, launch their verifications together» y «If a fix round must change a file that a task started meanwhile also changes, pause that task until the fix lands» por las reglas del diseño: D5-A, `waiting:`, la condición de archivos sobre toda tarea en `[?]`, el orden de los arreglos y la reapertura sin gastar rondas.
- «Commit verified tasks»: orden del plan con dependencias primero, y el commit conjunto `(T2, T3)` para tareas del mismo punto que comparten un archivo.
- El archivo no repite el criterio de riesgo ni la compatibilidad: remite a `plan-format.md`.
**Verification:** dan 0 «this is the one exception to one verified task per commit», «a task that depends on it waits for `[x]`», «and nothing earlier» y «joint commit»; dan 1 las frases de lanzamiento al alcanzar el punto, de verificación en serie de tareas que comparten archivo, de re-verificación en el momento de su propia regla, de D5-A (mismo punto y punto anterior), de `waiting:`, de la regla de archivos compartidos (implementación, arreglo y verificación), de la reapertura sin gastar rondas, de la tercera reapertura al usuario, del commit al reportar, de la publicación para merge con todo en `[x]` y del revert de una tarea retirada; `go test ./tests/content/...` pasa; lectura contra el diseño (D12-A).

## T3 — Textos que dependen de la verificación por tarea

- [x] La descripción de `hive-verify-task`, la fila «Build iteration» de `verification.md`, `plan-revision.md:10-11` y `agent-delivery.md:9` sitúan la verificación en el punto de control; la búsqueda confirma o corrige el resto.
Verificado por `hive-verify-task` (~74 s; de nuevo tras D12-A, ~71 s): AC6 cumplido (`hive-verify-task.md:3`, `verification.md:33`, `plan-revision.md:10-11`, `agent-delivery.md:9`). Commit `a679cb28`. Disposición de la búsqueda: siguen coherentes sin cambio `ui-gate.md:32-34,40`, `flow-walkthrough/SKILL.md:60,66`, `reporting.md:15,119`, `questions.md:89,117`, `delegation.md:89`, `flow-build/SKILL.md:44,48`, `flow-plan/SKILL.md:36,52,70`, `git-workflow/SKILL.md:16`, `agent-delivery.md:20` (historia), `site/.../flows/plan.md:65`, `site/.../flows/build.md:45`; `site/.../flows/build.md:44` queda para T4; sin resultados en `tests/` ni `integrations/`. Una observación del verificador (cuándo se re-verifica una tarea reabierta por una revisión del plan) se aplicó en `plan-execution.md:19` con un commit `(T2)` aparte.

**Closes:** AC6.
**Depends on:** T2.
**Locations:** `content/agents/quality/hive-verify-task.md:3`; `content/skills/flow-build/references/verification.md:33`; `content/skills/flow-plan/references/plan-revision.md:10-11`; `_support/docs/architecture/agent-delivery.md:9`.
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Changes:** `plan-revision.md:10`: una tarea `[x]` reabierta se verifica en el momento que fija su propia regla, sin esperar un punto; `:11`: una tarea `tdd` o `check` agregada lleva su `Checkpoint` cuando el plan los usa.
**Verification:** «after the implementer reports a task» da 0 en el rol; «under the per-task procedure» da 0 en `plan-revision.md`; las frases nuevas dan 1 en cada archivo; `rg -n "after the implementer reports|before its commit|per-task|each task|tasks verified|\[\?\]" content site _support/docs/architecture tests integrations` se revisa entero, y el progreso anota cada hallazgo con su disposición, incluidos `ui-gate.md:32-34,40`, `flow-walkthrough/SKILL.md:66`, `reporting.md:119`, `questions.md:89,117` y `flow-build/SKILL.md:48`, que la revisión del plan dio por coherentes.

## T4 — Sitio y changelog

- [x] `site/.../flows/plan.md` explica cómo el plan elige los puntos de control, remitiendo al criterio sin copiar sus condiciones; `site/.../flows/build.md:44` explica cuándo verifica el verificador; `CHANGELOG.md` lo registra en `[Unreleased]` con `(#168)`.
Verificado por `hive-verify-task` (~50 s): AC7 cumplido (`CHANGELOG.md:92-98`, `site/.../flows/plan.md:34`, `site/.../flows/build.md:44,71`). El orquestador aplicó dos observaciones de redacción (el criterio de riesgo en `plan.md` no es un ranking; una palabra que faltaba en el changelog). Commits `6774839c` y el ajuste del changelog.

**Closes:** AC7.
**Depends on:** T2; el merge de gh-167 (archivos compartidos).
**Locations:** `site/src/content/docs/flows/plan.md`; `site/src/content/docs/flows/build.md:44`; `CHANGELOG.md`.
**Execution:** delegada a `hive-write-docs`.
**Test approach:** check.
**Verification:** `#168` da 1 en `CHANGELOG.md`; la frase nueva da 1 en cada página; las páginas describen y enlazan, no copian la regla.

## Verificación y revisión

**Suite completa sobre `6774839c`:** `go vet ./...` y `go test ./...` pasan (2026-10-25), con `development` (#167) ya integrado en la rama.


| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Por tarea | tras cada tarea, en el worktree | `hive-verify-task` por criterio | incluida en la implementación |
| Suite completa | candidato final | `go vet ./...` y `go test ./...` pasan | incluida |
| Revisión de código | antes del merge | `/code-review` (declarado), con un tope de una re-revisión | `Review` declarado |
| CI | PR a `development` | checks del repositorio en verde sobre la base actual | `Delivery: automatic` |

Sin recorrido en vivo ni de UI: el cambio es texto de guía y de páginas sin cambio de maquetación. Los pilotos de comportamiento siguen pausados (`AGENTS.md`), así que el efecto en tiempo real se observará en las próximas sesiones reales, no aquí.

## Estado de la revisión y progreso

- **Build.** T1 escrita por un `hive-write-docs` (~36 s, ~95 k tokens). Primer veredicto de `hive-verify-task` (~53 s): AC1 y AC2 cumplidos, con dos observaciones que el orquestador aplicó dentro de las Locations de T1: los puntos de control son obligatorios en un plan con criterios numerados (la compatibilidad queda para planes anteriores) y la plantilla de verificación tiene un lugar para la lista de puntos. Re-verificación (~43 s): AC1 y AC2 cumplidos (`plan-format.md:28-35,112,130`, `plan-review.md:22`). Commit en la rama de gh-168.

- **Revisión del plan, ronda 1** (revisión `2312da1aca75` / `6c44b49c470c` / `e023417e675e`), dos `hive-review-plan`:
  - Lógica de ejecución (~87 s, ~97 k tokens): seis bloqueantes, B1–B6 (verificadores y arreglos sobre archivos compartidos, commits de tareas que comparten archivo, `[x]` sobre una predecesora que cambia, punto que no se alcanza, tope de reaperturas, condición de archivos sobre tareas en `[?]`), y cuatro no bloqueantes (espera más estricta que D5-A, re-verificaciones fuera de puntos, orden del plan, decidibilidad del criterio).
  - Consistencia de la guía (~110 s, ~109 k tokens): tres bloqueantes (frases viejas de `:26` y `:30` sin reescribir, AC5 ya cierto en la base, verificaciones que no cubrían AC3 y parte de AC1) y no bloqueantes (compatibilidad en dos hogares, `plan-revision.md:11`, búsqueda estrecha, `questions.md`).
  - Aplicado: B2 lo decidió el usuario (D8-A, commit conjunto). El resto lo aplicó el orquestador como decisiones menores que se siguen de D4-A y D5-A: diseño ampliado (alcanzar un punto, `waiting:`, orden de los arreglos, reapertura sin gastar rondas, re-verificación en el momento, espera solo por las dependencias, una sola casa para la compatibilidad), AC3–AC5 reescritos, AC8 nuevo, verificaciones con conteo 0 y 1 por cláusula y búsqueda ampliada en T3.
  - D9-B: sin `hive-verify-change` final.
- **Re-revisión** (revisión `0c7fb0e4707d` / `0a107ae7b02c` / `d09caf4fd972`, los mismos revisores):
  - Consistencia (~34 s, ~119 k tokens): resueltos los tres bloqueantes; dos nuevos, aplicados tal como los propuso: la frase «the one exception» de `:30` nombra las dos excepciones; la re-verificación de una tarea reabierta corre «en el momento que fija su propia regla» (no entre rondas del walkthrough). También la fila `[/]` con `reopened by the fix of`.
  - Lógica de ejecución (~43 s, ~107 k tokens): resueltos B1, B3–B6 y N1–N3, y B2 por D8-A; dos bloqueantes nuevos y cuatro no bloqueantes, aplicados tal como los propuso: R1, condición de archivos sobre toda tarea sin commit; R2, el commit conjunto se hace cuando todas están en `[x]` y ninguna espera la puerta de UI; R3, una reelaboración o un cambio en las Locations de la tarea en espera la reabre; R4, la reapertura baja por la cadena y la tercera va al usuario; R5, `blocked:` bajo `[/]` antes de reportarse; R6, `waiting:` en AC1.
  - Como son propuestas de los propios revisores, el orquestador las concilió sin otra ronda (`plan-review.md`, «Reconcile and close»).
- Plan listo el 2026-10-25.
- **`/code-review` del PR #171** (~89 s): 10 hallazgos; el bloqueo (dos tareas independientes del mismo punto que comparten archivo nunca arrancan) lo confirmó un `hive-refute-claim` (~39 s). Aplicados todos: excepción de archivos para tareas del mismo punto, verificador propio para la tarea que llega tarde a su punto, espera de las tareas sin punto, re-verificación de las dependientes tras un arreglo en lugar de la prueba por archivos, pausa ante un arreglo fuera de las Locations, re-verificación tras una revisión del plan que solo cambia la verificación, limpieza de lo de una tarea retirada antes del commit, el bloqueante de `plan-review.md` limitado a tareas `tdd` y `check`, y el changelog acotado a esas tareas. T2 vuelve a `[?]` para re-verificar AC3, AC4, AC5 y AC8.
- **Re-verificación de T2 tras la revisión** (~164 s): AC3, AC4 y AC5 cumplidos; AC8 `not met` solo porque `proposal.md` seguía pidiendo la prueba por archivos. Ronda de arreglo 1: AC8 alineado al diseño, y tres ajustes del verificador en `plan-execution.md`: la condición de archivos cubre cualquier tarea iniciada sin commit de otro punto, la re-verificación tras un arreglo incluye las tareas del mismo punto que comparten archivo y excluye las correcciones del walkthrough, y el orden de commits es el del plan entre las tareas listas.
- **Ronda 2 de T2** (~134 s): AC5 y AC8 cumplidos; AC4 `not met`: una dependiente sin commit en `[/]` o `[!]` quedaba fuera de la re-verificación completa. Arreglo de ronda 2 (único arreglo plausible, dentro de las Locations): esa dependiente recibe todos sus criterios en su próxima verificación. Decisión tomada sin preguntar, por un bloqueo que el verificador infirió en el diseño (una tarea de otro punto que comparte archivo con una tarea de UI en `[x]` esperando la puerta de UI nunca arrancaba): esa tarea puede arrancar y se suma a su commit conjunto. Se revierte quitando la excepción de `plan-execution.md:30` y la cláusula del commit conjunto.
- **Ronda 3 de T2** (~133 s): AC4 cumplido. El verificador infirió un ciclo de commits posible (una tarea de UI en espera, su dependiente y una tarea que se sumó a su commit conjunto); el orquestador agregó un desempate: cuando ninguna tarea lista puede commitearse porque cada una espera a otra, las `[x]` en espera van juntas en un commit conjunto. Commits `5a418178` (T2), `fbe71044` (T1), `9e619448` (T4); `go vet ./...` y `go test ./...` pasan sobre `9e619448`.
- **Re-revisión `/code-review` del PR #171** (~108 s, la única permitida): 10 hallazgos, varios bloqueos y mezclas que salían de tener varias tareas sin commit en el mismo árbol con «commit solo tras verificar». Llevado al usuario: D12-A (commit al reportar). Aplicado: reglas de ejecución y commits reescritas en `plan-execution.md` (verificadores en serie para tareas que comparten archivo, base de diff con los commits de la tarea, re-verificación que espera al verificador en curso), `hive-verify-task.md` (revisa los commits de su tarea), `plan-revision.md:11`, `site/.../flows/build.md:45,71` y el changelog. Quedan sin aplicar dos hallazgos menores: detectores en la revisión del plan para dependencias hacia adelante y tareas después del último punto (propuestos como seguimiento), y el costo de tests duplicados por verificador (mitigado al verificar en serie las tareas que comparten archivo). T2, T3 y T4 vuelven a `[?]` para re-verificar AC3, AC4, AC5, AC8, AC6 y AC7.
- **Re-verificación tras D12-A** (T3 ~71 s, T4 ~80 s, T2 ~211 s): AC6 y AC7 cumplidos (T4 con tres arreglos de redacción en el changelog y `build.md`); T2 cumple AC3, AC4 y AC5, y AC8 `not met` porque la regla de archivos compartidos no cubría verificaciones ni rondas de arreglo. Ronda de arreglo 1 de T2: la regla cubre implementación, arreglo y verificación, y las reglas que lanzan verificadores o arreglos la aplican primero; además, con `hold` lo de una tarea retirada se quita antes de presentar el resultado, un revert que no aplica limpio va al usuario, y `ui-gate.md:23` deja de suponer que las tareas de UI quedan sin commit. Verification de T2, «Ejecución» del diseño y AC3 alineados a D12-A.
- **Ronda 1 tras D12-A verificada** (~94 s): AC8 cumplido. Commits `4932ecdd` (T2), `a308c273` (T3), `6ade8468` (T4); `go vet ./...` y `go test ./...` pasan sobre `6ade8468`. Tope de revisión alcanzado: sin otra ronda de `/code-review`.
