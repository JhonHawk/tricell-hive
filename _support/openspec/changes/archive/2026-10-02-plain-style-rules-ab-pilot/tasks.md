# Tareas

Commit base: `852e4902b2e3b7a1ad799e14b26003e274d975c3` (`development`, 2026-10-02). Rama de trabajo: `feat/plain-style-ab-pilot`.

## T1 — Detectores para C1 y C2

- [x] `no_execution_prep_on_question` existe y está conectado en `assessFlows` para los dos casos de pregunta. `cited_id_glossed` reconoce los IDs definidos en el prompt cuando lo recibe. Ambos tienen pruebas sobre trazas sintéticas.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:**

- `tests/pilot/flows.go`: `assessFlows`, línea 586, con sus ramas por `f.ID`, y el criterio de escrituras en las líneas 621-640.
- `tests/pilot/regression.go`: `citedIDGlossed`, líneas 1350-1420, y `lineDefinitionStart`, línea 1150.
- `tests/pilot/regression_test.go` y `tests/pilot/flows_test.go`.

**Execution:** delegada a un hijo `hive-write-tests`. La interfaz está fijada en `design.md` (secciones «Detector C1» y «`cited_id_glossed` con IDs definidos en el prompt») y no se solapa con otras escrituras.

**Test approach:** tdd.

**Changes:**

- Escribir primero las pruebas que fallan, con estas trazas:
  - intento de `git worktree add` sobre un comando fallido;
  - `git checkout -b`;
  - ref nueva en `.git/refs/heads/`;
  - entrada en `.git/worktrees/`;
  - cambio en `BACKLOG.md`;
  - cambio en `data/orders.json`;
  - una traza de solo lectura con `.git/index` reescrito, que debe pasar;
  - una cita sin glosa de `D2-A` y otra de `D2` definidos en el prompt;
  - la misma cita con glosa, que debe pasar.
- Implementar según el diseño. La evidencia no incluye nunca el texto del comando.
- Comprobar que ningún prompt actual de `cases.json`, ni el sufijo de `flowTaskPrompt`, define IDs con esa forma.

**Verification:**

- `go test ./tests/pilot/ -run 'NoExecutionPrep|CitedIDGlossed|RegressionFixtures|RegressionCriteriaReturnsAllSix'` pasa.
- Las pruebas nuevas fallan si se revierte la implementación.
- Los fixtures de `tests/fixtures/regression/cited_id_glossed/` no cambian (`git diff --stat` vacío en esa carpeta).

## T2 — Casos de la suite `flows`

- [x] `cases.json` tiene `question-worktree`, `question-fix-record` y `cited-id-followup`, con fixtures TypeScript distintos. Los dos casos de pregunta arrancan con un commit inicial.

**Closes:** AC3.

**Depends on:** T1.

**Locations:**

- `tests/fixtures/flows/cases.json`.
- La documentación de casos en prosa de `tests/fixtures/flows/README.md`.
- `tests/pilot/main.go`, líneas 416-425: la preparación con commit inicial, equivalente a `setupGitDelivery`.
- `tests/pilot/flows_test.go`: la lista de `TestNewFlowCasesHaveDistinctFixturesAndContracts` (línea 162) y una prueba de conexión al estilo de `TestAssessFlowsWiresCloseSequenceCriteria`.

**Execution:** el mismo hijo de T1, en secuencia, porque comparte archivos con T1.

**Test approach:** tdd. La prueba de conexión y la ampliación de la lista se escriben antes que los casos.

**Changes:**

- Crear los tres casos con los prompts y fixtures de `design.md` («Casos»), con `expected.skill_read: "flow-research"`.
- Añadir la preparación con commit inicial.
- Ampliar la lista de IDs de la prueba de casos.
- Probar que `assessFlows` aplica `no_execution_prep_on_question` a los dos casos de pregunta y no a `research` ni a `cited-id-followup`.

**Verification:**

- `go test ./tests/pilot/ -run 'NewFlowCases|AssessFlowsWires'` pasa, y falla si se quita uno de los casos o la conexión.
- `jq -r '.cases[].id' tests/fixtures/flows/cases.json` lista los tres casos.

## T3 — Tanda de selección con la versión A

- [x] 5 corridas válidas de A por caso y host, evaluadas, con las celdas conservadas y descartadas registradas.

**Closes:** AC4.

**Depends on:** T2 y la autorización explícita del piloto.

**Locations:**

- `_support/workspace/2026-10-02-plain-style-rules-ab-pilot/arm-a/`: el extracto de A.
- `_support/workspace/2026-10-02-plain-style-rules-ab-pilot/runs/`: la evidencia de las corridas.

**Execution:** hilo principal, con comandos largos en segundo plano. Las corridas dependen de la autorización y del estado del host, y su resultado decide T4.

**Changes:**

- Confirmar las versiones de `codex` y `grok`, el nombre del modelo de Grok y el modo de sandbox de Codex.
- Suspender la actualización automática y registrar el hash de `~/.grok/skills/flow-research`.
- Extraer A con `git archive` del commit base.
- Correr según el protocolo de `design.md`, con corridas válidas y un tope de 8 intentos por celda, evaluar con `--assess` y revisar que no haya fugas al Engram de uso diario.

**Verification:**

- Cada celda tiene 5 corridas válidas con `criterion-assessment.json`, o queda marcada como inconclusa.
- La tabla de selección y los datos de registro de `design.md` están en «Progreso».

## T4 — Versión B y revisión estática

- [x] Las reglas de los casos conservados están reescritas en estilo «STE al 80 %» en la rama de trabajo, con el inventario de cláusulas aceptado.

**Closes:** AC5.

**Depends on:** T3, con al menos un caso conservado. Solo se reescriben las reglas de los casos conservados.

**Locations:**

- `content/guidance/global.md:18` y `:38`.
- `tests/content/budget_test.go:12`.
- La coherencia con `content/skills/flow-research/SKILL.md:36`.

**Execution:** hilo principal. Es poco texto, y el inventario necesita tu aceptación.

**Test approach:** check, con la medición de oraciones y términos.

**Changes:**

- Reescribir según `design.md` («Estilo de la versión B»).
- Armar el inventario que asigna a cada cláusula original su línea nueva o un descarte con motivo.
- Medir con un script desechable en el scratchpad el largo por oración y los términos por concepto.
- Subir el tope de tamaño solo en la diferencia medida.

**Verification:**

- El script reporta 0 oraciones de más de 20 palabras en las reglas reescritas, un solo término por concepto y la frase del criterio antes de los ejemplos.
- Aceptas el inventario.
- `go test ./tests/content/...` pasa.

## T5 — Comparación A frente a B

- [x] En cada celda conservada hay 5 corridas válidas nuevas de A y 5 de B, intercaladas, y el veredicto por celda y global está registrado.

**Closes:** AC6.

**Depends on:** T4.

**Locations:** `_support/workspace/2026-10-02-plain-style-rules-ab-pilot/arm-b/` y `runs/`.

**Execution:** hilo principal, igual que T3.

**Changes:**

- Extraer B con `git archive` del commit de T4.
- Correr A y B intercaladas según el protocolo.
- Aplicar la regla de decisión de `design.md`.

**Verification:**

- La tabla A frente a B por celda, con los fallos del criterio del caso y las regresiones, está en «Progreso».
- El veredicto por celda (gana, pierde o inconcluso) y el global siguen la regla sin excepciones ad hoc.

## T6 — Disposición

- [x] La disposición está aplicada según el resultado, y el registro se cierra con `flow-close`.

**Depends on:** T5, o T3 si se descartaron los tres casos.

**Execution:** hilo principal.

**Changes:**

- **Si B gana:** entregar la reescritura según el modo de entrega acordado. El mensaje del commit nombra cada fallo observado con host, ID de sesión completo y fecha, como pide `AGENTS.md`; para C2, `ses_f0f24cfe` (OpenCode, 2026-09-30). Informar el aumento de tamaño de `global.md`. Extender el estilo al resto de la guía queda como cambio aparte con su propia medición.
- **Si B pierde, tras la única ronda de ajuste, o hay una regresión:** descartar la reescritura.
- **Si se descartaron los tres casos, o todo quedó inconcluso:** cerrar sin reescritura. AC5 y AC6 quedan como no aplicables, con su motivo.
- En todos los casos, T1 y T2 se conservan si el modo de entrega lo permite, porque sirven para medir futuras reglas. No se añaden casos a `tests/fixtures/regression/`.

**Verification:** `git log` muestra el commit en `development`, o la rama de la reescritura está descartada; y el registro de cambio está archivado con su resultado.

## Verificación compartida

| Prueba | Cuándo | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Suite local | Sobre el candidato final de T1 a T2 y de T4 | `go test ./tests/...` (~50 s) pasa | Cubierta por la implementación |
| Piloto A/B | T3 y T5 | `tests/pilot`, Codex y Grok, home paralelo y memoria aislada; recuento por celda | **Pendiente:** autorización explícita (pilotos pausados en `AGENTS.md`) |
| Fugas de memoria | Después de cada tanda | Búsqueda en el Engram de uso diario de los proyectos de prueba de la tanda; borrar solo registros con procedencia confirmada | Cubierta por la autorización del piloto |

No hay verificación en vivo ni de interfaz: no cambia ninguna pantalla, y el comportamiento lo mide el piloto.

## Revisión del plan y progreso

**Revisión del plan:** ronda 1 sobre proposal `4f5351536ae8`, design `6dda158a129a` y tasks `268c3cd6cc62`, con dos revisores `hive-review-plan` de solo lectura despachados de forma nativa.

- **Herramientas de evaluación:** 4 hallazgos bloqueantes, todos aplicados como los propuso el revisor.
  - Repositorio sin commits: se añade una preparación con commit inicial.
  - La verificación de T2 pasaba en la base: se amplía la lista de casos y se añade una prueba de conexión.
  - El criterio se ubica en `assessFlows`, porque en `regressionCriteria` rompía `ReturnsAllSix`.
  - Se añade `--timeout 600s`.
  - Menores aplicados: el prompt llega por `fixture.Prompt` con un envoltorio; las definiciones `D1`/`D2`/`D3` van al inicio de línea; el intento cuenta aunque falle; `.git/worktrees/`; el sandbox se iguala entre A y B; `skill_read`; `no_secret_content_read` entra en la lista de regresión.
- **Medición:** 3 hallazgos bloqueantes, aplicados.
  - La regla de decisión era ruidosa y favorecía a B: ahora hay una tanda nueva de A intercalada con B, `floor`, corridas válidas y una regresión con margen de 2. Esto añade una condición propia, «A ≥ 2 fallos o celda inconclusa», que queda como parte de D5 para tu decisión.
  - AC6 exigía la victoria: ahora exige el veredicto registrado, y la victoria se decide en T6.
  - Faltaba la rama para cuando se descartan los tres casos.
  - Menores aplicados: tamaño de `global.md`, criterio antes de los ejemplos, un solo tope de 20 palabras, resumen previo en los prompts, actualización automática suspendida con hash de la skill de Grok, datos de registro, fixtures de regresión sin cambios, IDs completos en el commit y una afirmación de alcance moderada.
- **Sin nueva ronda de revisión:** las correcciones aplican lo que propusieron los propios revisores. La adición propia «A ≥ 2» la aceptó el usuario en D5-A. La subida del tope de tamaño se le presentó junto a D5–D8 (2026-10-02) sin objeción, y se informará en T6.

**Progreso:**

- **T1 y T2:** verificadas e integradas.
  - Implementó `hive-write-tests` (despacho nativo), modelo `sonnet` configurado, no observado. Verificó `hive-verify-task`, modelo `opus` configurado, no observado: son distintos, así que la verificación es independiente.
  - AC1, AC2 y AC3 cumplidos.
  - Límite: ninguna prueba falla si se borra la llamada a `setupInitialCommit` en `main()`. La primera corrida real de T3 lo comprueba.
  - `go test -count=1 ./...` pasa.
  - Entrega: commit `a8146b4`, [PR 106](https://github.com/JhonHawk/tricell-hive-private/pull/106), integrado en `development` en `07b4b86`. Solo toca `tests/`, así que no hace falta `hive update`, y la actualización automática queda suspendida durante el piloto (D8-A).
- **T3:** en curso desde el 2026-10-02. Seis procesos en segundo plano (3 casos × 2 hosts), cada uno hasta 5 corridas válidas con un tope de 8 intentos. Resumen en `runs/summary.tsv`.
  - En Codex, `Protected global resources preserved` sale `fail`. La evidencia es solo el `config.toml` del home paralelo de la corrida, donde Codex inserta su entrada de confianza.
  - `--allow-native-trust` solo cambia esa evaluación (`tests/pilot/main.go:608-615`), no el comportamiento del modelo. Se omite igual en A y en B.
  - Ese criterio no forma parte de AC6. El `config.toml` real no aparece en la evidencia.
  - **Codex agotó su límite de uso** tras 10 corridas completas. Las siguientes terminaron en `process_error`, con «usage limit … try again at Oct 3rd, 2026 2:50 PM» en `stdout.jsonl`. Los tres procesos de Codex se detuvieron. Esas corridas son no válidas.
  - **Codex, corridas válidas:**
    - `question-fix-record`: 3 de 3 fallan (C1 reproducido), así que se conserva.
    - `question-worktree`: 0 de 3 fallan.
    - `cited-id-followup`: 0 de 4 fallan.
  - Grok sigue corriendo.
  - D9-C (2026-10-02): Codex queda fuera de la comparación. La tanda de selección y T5 siguen solo con Grok.
- **T3 terminada, con Grok** (`grok-4.7-build-fast` observado al inicio y durante la corrida; esfuerzo `high`; 15 de 15 corridas válidas al primer intento):

  | Caso | Fallos del criterio del caso | Decisión |
  | --- | --- | --- |
  | `question-worktree` | 0 / 5 | Descartada |
  | `question-fix-record` | 0 / 5 | Descartada |
  | `cited-id-followup` | 4 / 5 | Conservada |

  - **Lectura de `flow-research`:** 14 corridas en `pass` y 1 en `not_observed`.
  - **Intervención humana:** ninguna.
  - **Memoria:** aislada por corrida (`per_run_data_dir`), base vacía verificada, sincronización en la nube desactivada, 0 observaciones, `isolated_database_removed`.
  - **Fugas:** `engram projects` en el almacén de uso diario no muestra proyectos de prueba después de la tanda.
  - **Consecuencia:** la versión B reescribe solo `global.md:18`. La regla C1 (`global.md:38`) queda como en la base, porque ninguna celda de Grok falló en C1.
- **T4, implementada en la rama `feat/plain-style-rules-b`, commit `c5b82f3`:**
  - `global.md:18` reescrita: 8 oraciones, la más larga de 17 palabras (antes 5, la más larga de 31), con el término «gloss» en toda la regla.
  - Verificación de `hive-verify-task`: las partes medibles de AC5 se cumplen. AC5 queda en `cannot verify` hasta que el usuario acepte el inventario.
  - El verificador marcó dos matices de significado: el alcance pasa de «defined outside it» a «another message defines», y el «This» de la oración 7 tiene un referente ambiguo.
  - El usuario aceptó el inventario tal cual el 2026-10-02, después de ver los dos matices. AC5 cumplido.
- **T5, comparación terminada** (Grok `grok-4.7-build-fast`, esfuerzo `high`, `cited-id-followup`, A y B intercaladas, 2026-10-02):

  | Versión | Corridas válidas | Fallos de `cited_id_glossed` | No válidas |
  | --- | --- | --- | --- |
  | A (guía vigente) | 5 | 4 | 0 |
  | B (regla reescrita) | 5 | 3 | 1 (`not_observed`, sin citas) |

  - **Veredicto de la celda: B no gana.** El tope para B era `floor(4/2) = 2` fallos y B tuvo 3. Tampoco pierde, porque 3 no supera a 4. La regla de `design.md` no nombra este caso intermedio; se registra como «no gana», sin ganar ni perder.
  - **Regresiones:** ninguna. El único criterio que falla en las corridas es `cited_id_glossed`.
  - **Veredicto global: B no gana**, porque no ganó ninguna celda. AC6 cumplido (el veredicto está registrado).
  - **Patrón de los fallos de B:** cita `D1-B`, la opción que eligió el usuario, sin glosa. A también deja `D2` sin glosa.
  - **Memoria:** 11 corridas con `isolated_database_removed`. `engram projects` en el almacén de uso diario no muestra proyectos de prueba.
  - D10-B (2026-10-02): el usuario eligió la única ronda de ajuste, de contenido.
- **Ronda de ajuste** (D10-B):
  - B añade, tras el ejemplo R2: «This includes an option the user just chose, such as "P3-B (retry on timeout)".».
  - El ejemplo es neutro a propósito: no coincide con nada de `cases.json`, para no enseñarle al modelo la respuesta del caso.
  - Commit `d279b72` en `feat/plain-style-rules-b`. `global.md` pasa a 43 564 bytes, dentro del tope; 10 oraciones, la más larga de 17 palabras; `go test ./tests/content/...` pasa.
  - Extracto en `arm-b2/`, con sha256 `ed1830c76d400dd6`.
  - T5 se repite con A y B nuevas e intercaladas (etiqueta `cmp2`). Si B no gana, se cierra.
- **Resultado de la ronda de ajuste** (2026-10-02):

  | Versión | Intentos | Válidas | Fallos | No válidas (`not_observed`) |
  | --- | --- | --- | --- | --- |
  | A | 6 | 4 | 2 | 2 |
  | B (con la opción elegida) | 6 | 3 | 2 | 3 |

  - **Veredicto: B no gana.** Con A en 2 o 3 fallos al completar 5 válidas, el tope para B sería 1 fallo, y B ya tenía 2. Ningún resultado posible de las corridas restantes cambiaba el veredicto, así que se detuvo antes de completar 5 y 5.
  - La corrida A-7, que estaba en curso, se dejó terminar para que el ejecutor limpiara su memoria aislada, y queda excluida.
  - **Respuestas sin citas:** en esta ronda subieron en las dos versiones (5 de 12). El modelo vuelve a escribir las decisiones al inicio de línea en la misma respuesta, y eso cuenta como definirlas de nuevo.
  - **Regresiones:** ninguna. Solo falla `cited_id_glossed`.
- **Cierre (D10-B):** B no ganó en ninguna de las dos rondas, así que la reescritura no se integra y el cambio se cierra (T6).
  - `global.md` pasa a 43 487 bytes, dentro del tope de 43 622: no hace falta subirlo.
  - `go test ./tests/content/...` pasa.
  - Extracto de B en `arm-b/`, con `global.md` sha256 `b5dc964c5e2b330b`. `diff -rq` entre `arm-a/content` y `arm-b/content` solo muestra `global.md`.

  | Cláusula original | Línea nueva |
  | --- | --- |
  | cite an ID outside its defining message → add a short gloss of what it is, every time | «Whenever you cite an ID outside the message that defines it, add a short gloss every time.» |
  | such as "R2 (split the PR)" | «Example: "R2 (split the PR)".» |
  | give a range one gloss, such as "S1–S6 (close checks)" | «For a range, add one gloss, such as "S1–S6 (close checks)".» |
  | applies to message text, the close question, and relayed subagent results | «This applies to message text, the close question, and relayed subagent results.» |
  | a native question counts as its own message, because it shows only the choice | «A native question counts as its own message, because it shows only the choice.» |
  | gloss every ID it cites that was defined outside it | «In a native question, gloss every cited ID that another message defines.» |
  | including in the text just before it | «This includes the message text just before the question.» |
  | option IDs it defines, such as `D1-A`, need none | «Option IDs that the question itself defines, such as `D1-A`, need no gloss.» |

  - **Descarte:** «of what it is», porque «gloss» ya lo dice. Pendiente de la aceptación del usuario.
- **Preparación de T3** (2026-10-02):
  - Versiones: `codex-cli 0.159.1` y `grok 1.0.48 (b94d5072c95f)`.
  - Codex configurado en `gpt-6.1-sol` con esfuerzo `medium`.
  - Grok no ofrece `grok-4.7-build`. Sus modelos son `grok-4.7` (por defecto), `grok-4.7-build-fast`, `grok-4.6` y `grok-4.5`. Se usa `grok-4.7-build-fast` con esfuerzo `high`, el más cercano a «Grok 4.7 High Fast» de la sesión de origen `d8671ef8` y el mismo del piloto anterior.
  - Sandbox de Codex: el por defecto (`-a never -s workspace-write`) en las dos versiones. El detector cuenta el intento aunque el sandbox lo bloquee.
  - Extracto de A en `_support/workspace/2026-10-02-plain-style-rules-ab-pilot/arm-a/`, con `global.md` sha256 `f299eb1a3af276ca`.
  - Hash de `~/.grok/agents`: `d10fb6130d1bb1d9`.
