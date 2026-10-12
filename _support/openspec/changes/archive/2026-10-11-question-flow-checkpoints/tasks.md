# Tareas: preguntas en puntos de decisión

Base del cambio: `bba2fd48` (`development`, 2026-10-11). Rama de trabajo histórica: `feat/guidance-rework`, integrada por [PR #46](https://github.com/JhonHawk/tricell-hive/pull/46) en `16028d3575e7286cc028985cd2138c3268257add`. Tareas verificadas y registro archivado; la instrucción inicial sin commits quedó sustituida por las entregas posteriores documentadas en `proposal.md`.

## Decisiones de construcción

- **P1-A** (usuario, 2026-10-11): el límite de bytes de `global.md` pasa a ser informativo (T7). Ver [design.md](design.md#texto-propuesto-para-globalmd).

## Tareas

### T1 — Reglas globales de preguntas

- [x] `content/guidance/global.md` aplica las reglas 1 a 6 de [design.md](design.md#texto-propuesto-para-globalmd), sin depender de un límite de bytes.

**Closes:** AC2, AC3, AC4, AC5, y la parte de AC1 que está en `global.md`.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:21-24`, `:31`, `:34`.

**Execution:** hilo principal. La redacción y el presupuesto están acoplados con T2 y T3, y el brief sería más largo que el cambio.

**Test approach:** check, con las comprobaciones de abajo.

**Changes:** aplicar las reglas 1 a 6 del diseño. Revisar contra la lista de intenciones del diseño que nada se pierda.

**Verification:**
- **AC2:** `rg -n "even when the options also live" content/guidance/global.md` no encuentra nada. Leer la línea 22: el resumen lleva ID, pregunta y recomendada con su razón y su ID glosado, sin las demás opciones; las opciones van en el texto siempre que ninguna tarjeta que espera lleve la pregunta.
- **AC3:** cada uno de `rg -n -i "decision point"`, `rg -n -i "minor call"` y `rg -n -i "leave to proceed"` sobre `content/guidance/global.md` encuentra la línea 21. Leerla: la decisión menor excluye lo consecuente y los efectos que necesitan autorización.
- **AC4:** `rg -n -i "in one round" content/guidance/global.md` encuentra la regla. Leer la línea 22: reparte preguntas u opciones con una línea de texto antes de cada llamada.
- **AC5:** `rg -n -i "next round"` y `rg -n -i "earlier flow"` sobre `content/guidance/global.md` encuentran la regla 3.
- **AC1 (parte global):** `rg -n -i "offer to show|showing recommendations|show the recommendations" content/guidance/global.md` no encuentra nada.
- **Tamaño:** `go test -v ./tests/content/...` informa el tamaño nuevo.

### T2 — La investigación termina con la pregunta de ruta

- [x] `flow-research` reemplaza la oferta por la pregunta de ruta y pone la pregunta de ticket en la misma ronda.

**Closes:** AC6, y la parte de AC1 que está en `flow-research`.

**Depends on:** T1, porque usa su vocabulario («decision point», «round»).

**Locations:** `content/skills/flow-research/SKILL.md:30-36`.

**Execution:** hilo principal, por el mismo motivo que T1.

**Test approach:** check.

**Changes:** según la sección «Skills» del diseño.

**Verification:**
- `rg -c -i "offer" content/skills/flow-research/SKILL.md` da 0.
- `rg -n -i "keep investigating|route question" content/skills/flow-research/SKILL.md` encuentra la nueva regla.
- Leer las líneas 30 a 36: la pregunta de ticket va en la misma ronda, y una pregunta respondida por completo, sin trabajo posible, termina sin ruta.

### T3 — Alinear plan, entrega, cierre y validación

- [x] `flow-plan`, `delivery-decisions.md`, la pausa de validación de `verification.md` y el paso 7 de `flow-close` quedan consistentes con T1.

**Closes:** ninguno; mantiene la consistencia que exige AGENTS.md (Guidance structure).

**Depends on:** T1.

**Locations:**
- `content/skills/flow-plan/SKILL.md:16`
- `content/skills/flow-plan/references/delivery-decisions.md:14`, `:30`
- `content/skills/flow-build/references/verification.md:65-67`
- `content/skills/flow-close/SKILL.md` paso 7 (solo revisión)

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check.

**Changes:** según la sección «Skills» del diseño.

**Verification:**
- `rg -n -i "options with their effects|together in one exchange" content/skills/flow-plan/SKILL.md` no encuentra nada.
- `rg -n -i "always carries the options" content/skills/flow-plan/references/delivery-decisions.md` no encuentra nada.
- Leer las cuatro ubicaciones: ninguna exige las opciones completas en el texto cuando una tarjeta que espera lleva la pregunta, y ninguna repite la regla de rondas.

### T4 — Registro por host

- [x] La sección «Native question cards» de `agent-delivery.md` queda con las versiones del 2026-10-11, las filas actualizadas, la tabla de D2-A por host, el conflicto de Cursor, la desviación de Codex y los límites no verificados. La línea 145 pasa a historia.

**Closes:** AC7.

**Depends on:** ninguna; corre en paralelo con T1.

**Locations:** `_support/docs/architecture/agent-delivery.md:141-162`.

**Execution:** delegada a `hive-write-spec`. Es documentación con fuentes ya verificadas y no comparte archivos con T1–T3. El brief incluye la sección «Hosts» del diseño, la evidencia de hosts de [design.md](design.md#contexto-verificado) y las etiquetas de origen de la sección (doc, obs (src|bin|local), report, inf).

**Test approach:** check.

**Changes:** según la sección «Hosts» del diseño.

**Verification:**
- `rg -n "Checked on 2026-10-11" _support/docs/architecture/agent-delivery.md` encuentra la sección.
- `rg -n -i "deviat" _support/docs/architecture/agent-delivery.md` encuentra la desviación de Codex.
- Leer la sección: Cursor tiene la entrada del binario y la de la documentación con sus etiquetas; Pi nombra la extensión; la línea 145 dice que fue reemplazada; cada fila conserva su etiqueta de origen.

### T5 — Sitio y CHANGELOG

- [x] Las páginas del sitio de investigación, cierre y plan describen el comportamiento nuevo, y `CHANGELOG.md` tiene la entrada en `Unreleased`.

**Closes:** AC8.

**Depends on:** T1, T2 y T3, porque describe su redacción final.

**Locations:**
- `site/src/content/docs/flows/research.md:37`, `:39`, `:56`
- `site/src/content/docs/flows/close.md:44`, `:58`
- `site/src/content/docs/flows/plan.md:61`
- `CHANGELOG.md`
- `README.md` y `llms.txt`: solo comprobar que nada de lo que dicen quede falso.

**Execution:** delegada a `hive-write-spec` cuando T1–T3 estén verificadas. Las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas (AGENTS.md).

**Test approach:** check.

**Changes:** describir la pregunta de ruta directa, el resumen más la tarjeta y la respuesta que devuelve el trabajo a un flujo anterior. Agregar la entrada `Changed` en el CHANGELOG.

**Verification:**
- `rg -n -i "shortcut|loses nothing|shows its recommendations|offers to show" site/src` no encuentra nada.
- Leer `research.md`, `close.md` y `plan.md`: cada una describe la parte que le toca, y `plan.md` coincide con `flow-plan/SKILL.md:16` después de T3.
- `npm run build` en `site/` termina sin errores, si el sitio tiene ese script.

### T6 — Base de la medición

- [x] Los scripts de medición de [design.md](design.md#medición-después-del-despliegue) quedan en `_support/workspace/2026-10-11-question-flow-checkpoints/` y recalculan la base de M1 a M5 en la ventana de gh-9.

**Closes:** ninguno; prepara la medición posterior.

**Depends on:** ninguna; corre en paralelo.

**Locations:** `_support/workspace/2026-10-11-question-flow-checkpoints/` (ignorado por Git); se reutiliza `_support/workspace/2026-10-08-communication-decision-flow/bare_split.py`.

**Execution:** delegada a `hive-research`. Lee historiales de sesiones extensos; solo devuelve conteos y no imprime secretos.

**Test approach:** check.

**Verification:** cada script imprime solo conteos por host y ventana, con denominadores. Las copias de una misma conversación cuentan como una sesión. La base queda anotada en la sección de progreso.

### T7 — Tamaño informativo

- [x] `tests/content/budget_test.go` informa el tamaño de `global.md` sin fallar, y `AGENTS.md:24` pide indicar el cambio de tamaño en el PR.

**Closes:** AC9.

**Depends on:** ninguna.

**Locations:** `tests/content/budget_test.go`; `AGENTS.md:24`.

**Execution:** hilo principal. Son dos ediciones cortas, y el brief sería más largo que el cambio.

**Test approach:** check. El cambio quita una aserción; no hay comportamiento nuevo que probar primero.

**Changes:** según P1-A en el diseño.

**Verification:**
- `rg -n "Fatalf|globalGuidanceBudget" tests/content/budget_test.go` no encuentra nada.
- `go test -v -run TestGlobalGuidance ./tests/content/` pasa y muestra los bytes y caracteres.
- `rg -n "size budget" AGENTS.md` no encuentra nada.

### T8 — Resumen del plan listo en la sesión

- [x] `flow-plan` resume el plan listo en la conversación y termina con la pregunta de ruta hacia `flow-build`; la página del plan en el sitio y el CHANGELOG lo describen.

**Closes:** AC10.

**Depends on:** T3, porque toca el mismo archivo `flow-plan/SKILL.md`, en otra línea.

**Locations:** `content/skills/flow-plan/SKILL.md:61`; `site/src/content/docs/flows/plan.md:55`; `CHANGELOG.md`.

**Execution:** hilo principal. Es una edición corta y el brief sería más largo que el cambio.

**Test approach:** check.

**Changes:** pedido del usuario del 2026-10-11, que recupera el comportamiento del Hive anterior (`afdedf5c`, 2026-09-14: «/flow-plan presents the gate and the close as summaries in the chat — the file is the detail»). El resumen lleva el objetivo, las decisiones con su ID, una línea por tarea, la verificación, la entrega, lo abierto y la ruta. No se pega el plan entero. Termina con la pregunta de ruta, que la regla aclara que no es pedir permiso para seguir: un plan no autoriza implementar.

**Verification:**
- `rg -n -i "summarize it in the conversation" content/skills/flow-plan/SKILL.md` encuentra la línea 61.
- `rg -n -i "build route question" content/skills/flow-plan/SKILL.md` la encuentra.
- `rg -n -i "summarizes it in the conversation" site/src/content/docs/flows/plan.md` encuentra la respuesta de la página.

## Verificación compartida

| Comprobación | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go test ./...` pasa | Parte de la implementación |
| Sin la oferta en lo distribuido (AC1) | Candidato final | `rg -n -i "offer to show\|showing recommendations\|show the recommendations\|offer below\|make the offer" content` no encuentra nada | Parte de la implementación |
| Nombres de herramientas | Candidato final | `rg -n "AskUserQuestion\|request_user_input\|ask_user_question\|AskQuestion" content/guidance/global.md` no encuentra nada | Parte de la implementación |
| Lectura del usuario | Después de T1–T5 | El usuario revisa el diff del árbol de trabajo; no hay commits | Instrucción del usuario |
| Medición en sesiones reales | Después de un despliegue futuro | Métricas M1–M5 contra la base de T6, con la muestra mínima del diseño | Se ofrece al cerrar; este plan no la autoriza |

No aplica revisión de UI ni verificación en vivo: el cambio es guía en texto.

## Revisión del plan y progreso

- **Revisión** (2026-10-11), sobre design `90d56f923be0`, proposal `5d880a64c561` y tasks `6f7560aa6039`. Dos `hive-review-plan` nativos en paralelo, de solo lectura, en Claude Code 2.1.296 con el modelo heredado (no observado):
  - **Dominio guía y skills.** Bloqueantes: el presupuesto estaba mal medido (ahora P1); faltaba el caso «sin llamar a la tarjeta» de `verification.md:65` (ahora en la regla 2); la línea 24 chocaba con D2-A (ahora en la regla 4); «minor call» no tenía límite (ahora en la regla 1). Menores aceptados: intenciones perdidas, separadores de texto para `questionAfterDetail`, redacción para modelos de ejecución, líneas 30–32 de `flow-research`, comprobaciones débiles, doble hogar en `flow-plan:16`, glosa del ID recomendado, alcance de la oferta de `flow-report`.
  - **Dominio hosts y sitio.** Bloqueantes: a T5 le faltaban líneas del sitio y comprobaciones que pudieran fallar; el conflicto de Cursor estaba oculto; la medición no se podía comparar (ahora en T6 y el diseño). Menores aceptados: límites no verificados y cierre por tiempo de Grok, desviación de Codex, comprobaciones de T4, `plan.md:61`, nivel de release y ventanas de medición. El mensaje de commit no aplica, porque no hay commits; las sesiones quedan listadas en la propuesta.
  - **Conciliación:** el orquestador aplicó cada corrección tal como la propuso su revisor, por lo que no hace falta otra ronda. Límite: los revisores no verificaron versiones instaladas, la documentación de Cursor ni los historiales de sesiones.
- **Progreso (build, 2026-10-11):**
  - Rama `feat/guidance-rework` creada desde `bba2fd48`.
  - T1–T7 completas: cada tarea `tdd`/`check` aprobada por `hive-verify-task`. Ajustes menores de los verificadores aplicados: en `global.md:23`, «dismissed or timed-out»; en `site/.../plan.md`, el bloque «Pausa» y la ronda de entrega.
  - Modelos:
    - T1, T7 (y T2, T3 cuando lleguen): implementa el hilo principal con Opus 5.5 (observado: el host lo informa); verifica `hive-verify-task` con `sonnet`, configurado con la opción de modelo del lanzamiento, no observado.
    - T4: implementa `hive-write-spec` con `sonnet` (configurado, no observado); verifica `hive-verify-task` con `opus` (configurado, no observado).
  - Tamaño de `global.md`: 46.036 → 46.548 bytes (+512).
  - Tamaño final de `global.md`: 46.561 bytes (`go test -v ./tests/content/` lo informa).
  - **Base de la medición (T6):** scripts en `_support/workspace/2026-10-11-question-flow-checkpoints/` (`qfc_measure.py`, `qfc_selftest.py`, `README.md`). Ventana de gh-9: 2026-10-09T00:49Z → 2026-10-11T07:27Z.
    - Se usa la base **sin** la sesión `8fbc961c`. Esa sesión planeó este cambio y cita la regla vieja, lo que contamina M1. Decisión menor del orquestador.
    - Claude (6 sesiones, 37 momentos): M1 1/37; M2 27/35 (77,1 %); M3 2/35 (5,7 %, igual que `bare_split.py`); M4 12/35 (34,3 %).
    - Codex, Cursor, Grok, OpenCode y Pi: muestra insuficiente.
    - M1 encuentra 4 ofertas donde el diseño citaba 5; la diferencia está en las citas que el script excluye.
  - **Verificación compartida:**
    - `go test ./...` pasó en el candidato después de T1–T4 y T7. Después solo cambiaron texto del sitio, `CHANGELOG.md` y una palabra de `global.md:23`, y `go test ./tests/content/` volvió a pasar.
    - `npm run build` del sitio pasó.
    - Las búsquedas de AC1 y de nombres de herramientas no encuentran nada.
  - **Hallazgo incidental, sin confirmar:** un verificador vio una sola vez `could not import unicode/utf8 (open : no such file or directory)` al compilar `tests/content`, mientras otro verificador compilaba en paralelo. Al repetir pasó. Probable carrera en la caché de compilación de Go; no se reprodujo.
- **P1-A agregado después de la revisión:** AC9 y T7 (tamaño informativo). Sin otra ronda de revisión: lo decidió el usuario sobre evidencia propia, y el cambio es una edición acotada a una prueba y una línea de `AGENTS.md` sin otros consumidores (búsqueda de `budget_test|size budget|globalGuidanceBudget`).
- **Validación:** el usuario aprobó el diff de T1–T7 (V1-A, 2026-10-11).
- **T8** (pedido posterior del usuario): verificada por `hive-verify-task` con `sonnet` (AC10 cumplido); `go test ./tests/content/` pasa. El usuario aprobó su diff («lo veo bien», 2026-10-11).
- **Commits** (2026-10-11), en `feat/guidance-rework`, sin push:
  - T1 `14dde957`, T2 `04e2d405`, T3 `53fea4f1`, T4 `5068ba92`, T7 `ee23c200`, T5 `e9e54162`, T8 `c332a814`.
  - T6 no tiene commit, porque sus archivos están en `_support/workspace/`, que Git ignora.
  - Esta carpeta del cambio queda fuera de los commits hasta `flow-close`.
- **Despliegue** (pedido del usuario, «despleguemos»): `hive update` + `hive apply` desde `feat/guidance-rework` en `c332a814`, release `f41da9b2ce0db7b0be5ba50dc7ab0022`, a 2026-10-11T08:14Z, en los seis hosts. La regla nueva aparece en `~/.claude/CLAUDE.md`, `~/.codex/AGENTS.md`, `~/.config/opencode/AGENTS.md`, `~/.pi/agent/AGENTS.md` y `~/.cursor/AGENTS.md`, sin el texto viejo. Esta hora es el corte «después» de la medición (M1–M5).
- **Siguiente paso:** seguir con otros cambios en `feat/guidance-rework`. Push, PR a `development` y cierre esperan a que el usuario los pida.
