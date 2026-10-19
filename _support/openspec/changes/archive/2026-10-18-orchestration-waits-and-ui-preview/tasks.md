# Tareas

**Base:** `8c15f59a` en `feat/gh-67-flow-walkthrough-skill`, fijada al empezar la construcción el 2026-10-18.

Cada tarea es dueña de sus archivos; dos tareas no escriben el mismo archivo a la vez. T1, T2 y T3 corren en paralelo. T4 espera a T2, porque las dos editan `flow-build/SKILL.md` y `ui-gate.md`. T5 espera a T1–T4.

### T1 — M1 y M2 en `reporting.md`, `global.md` y `questions.md`

- [x] La condición del turno de notificación, la sección «Open with the context line» y la frase invariante en `global.md`, como fijan [design.md](design.md) M1 y M2.

**Closes:** AC1, AC2, AC3.

**Depends on:** ninguna.

**Locations:**

- `content/skills/flow-close/references/reporting.md`: «Lead with the result» `:7-9`, «Report a running or completed item» `:43-55` y «Work-item boundary» `:91`.
- `content/guidance/global.md:16`.
- `content/skills/flow-plan/references/questions.md`: paso 4 de «Use a waiting native card», y «Pause and handle the answer» `:58-68`, donde va la excepción en el punto de uso.
- `content/skills/flow-close/references/reporting.md:95-109`, el ejemplo resuelto.
- `content/voices/preamble.md:7`.

**Execution:** hilo principal. Es la casa de dos momentos que T2, T3 y T5 citan por ancla, y el cambio de `global.md` exige informar el tamaño.

**Test approach:** `check`, con las búsquedas de abajo.

**Changes:**

- La condición nueva va antes de la del ítem en curso, y la de este deja de aplicarse solo a la salida 3 de la condición nueva.
- La cláusula de notificaciones de `:91` pasa a puntero.
- La sección nueva conserva intactas las anclas existentes.

**Verification:**

- `rg -n -F "context line" content/skills/flow-close/references/reporting.md content/guidance/global.md` y `rg -n -F "open-with-the-context-line" content/skills/flow-plan/references/questions.md` devuelven una línea al menos en cada archivo (AC1, AC3). La comprobación original buscaba «context line» en `questions.md`, donde el diseño pide solo el puntero con ancla; se corrigió tras la verificación de T1.
- `rg -n -F "changes nothing" content/skills/flow-close/references/reporting.md` devuelve la condición nueva (AC2).
- `rg -n -F "notification that arrives meanwhile in one line and continue" content/skills/flow-close/references/reporting.md` no devuelve nada (AC2).
- `rg -n "^## Report a running or completed item|^## Open with the context line" content/skills/flow-close/references/reporting.md` devuelve dos líneas.
- `rg -n -F "second report" content/skills/flow-plan/references/questions.md` devuelve la excepción nueva, además de la frase existente (AC2).
- `rg -n -F "after any context line" content/voices/preamble.md` devuelve una línea, y `go test ./tests/content/...` pasa (AC1).
- La frase de `global.md` nombra las dos excepciones: `rg -n -F "context line" content/guidance/global.md` muestra la línea, y leerla confirma que menciona el retorno de un subagente y la respuesta corta (AC3).
- `wc -c -w content/guidance/global.md` se registra frente a 30.717 bytes y 4.699 palabras.

### T2 — M1 (hijo detenido) y M4 en `delegation.md`, `flow-build` y `ui-gate.md`

- [x] La condición del hijo detenido, la extensión del brief y los punteros, como fijan [design.md](design.md) M1 (hijo detenido y punteros de `delegation.md` y `flow-build/SKILL.md:67`) y M4.

**Closes:** AC4 (las partes de `delegation.md`), AC5, AC6.

**Depends on:** ninguna. Usa el ancla `#report-a-running-or-completed-item`, que ya existe, y la condición de notificación que escribe T1.

**Locations:**

- `content/skills/flow-research/references/delegation.md`: `:35`, `:61`, `:65`, `:67-71` y el ejemplo resuelto `:81`.
- `content/skills/flow-build/SKILL.md`: `:34` y `:67`.
- `content/skills/flow-walkthrough/references/ui-gate.md:21`.

**Execution:** delegada a `hive-write-docs`. Las fuentes y los puntos de destino están fijados en el diseño, y ningún otro escritor toca esos archivos mientras corre.

**Test approach:** `check`, con las búsquedas de abajo.

**Changes:** las de M1 y M4 en [design.md](design.md), en términos de capacidad, sin nombrar herramientas del host.

**Verification:**

- `rg -n -F "work of its own" content/skills/flow-research/references/delegation.md` devuelve la condición nueva (AC5).
- `rg -n -F "delegation.md#check-returned-work" content/skills/flow-build/SKILL.md` devuelve el puntero de la frase de vida (AC5), y `rg -n -F "delegation.md#deliver-and-resolve-instructions" content/skills/flow-build/SKILL.md content/skills/flow-walkthrough/references/ui-gate.md` devuelve el del brief de tarea y el de revisores (AC6).
- `rg -n -F "conditional rule" content/skills/flow-research/references/delegation.md` devuelve una línea (AC6).
- `rg -n -F "pre-rating" content/skills/flow-walkthrough/references/ui-gate.md` devuelve una línea (AC6).
- `rg -n -F "ends its turn with the in-progress item" content/skills/flow-research/references/delegation.md` no devuelve nada, y la línea por tramo de espera de `delegation.md` cita la línea de contexto (AC4).
- `wc -w content/skills/flow-walkthrough/references/ui-gate.md` da menos de 880.

### T3 — M3 (página) en `flow-walkthrough` y en el esqueleto

- [x] La primera vista en `walkthrough.html`, la regla «Open the page», los ajustes de `:30` y `:44` y el puntero de M1 en `:24`, como fija [design.md](design.md) M3.

**Closes:** AC4 (la parte de `flow-walkthrough`), AC7, AC8.

**Depends on:** ninguna. El puntero de `:24` cita el ancla que ya existe.

**Locations:**

- `content/skills/flow-walkthrough/SKILL.md`: «First rendered view» `:24`, `:30` y «Deliver» `:42-44`.
- `content/skills/flow-report/assets/skeleton-walkthrough.html`, comentario de uso `:8-31`, que recibe la lista del llenado parcial.
- `content/skills/flow-report/SKILL.md:10` y `:22`.

**Execution:** delegada a `hive-write-docs`, por la misma razón que T2.

**Test approach:** `check`, con las búsquedas y la autocomprobación de abajo.

**Changes:** las de M3 «Casa de la página» en [design.md](design.md).

**Verification:**

- `awk '/^## First rendered view/{f=1;next}/^## /{f=0}f' content/skills/flow-walkthrough/SKILL.md | rg -c walkthrough.html` imprime 1 o más (AC7).
- `rg -n -F "Open the page" content/skills/flow-walkthrough/SKILL.md` devuelve una línea al menos (AC8).
- `rg -n -F "the user opens the page" content/skills/flow-walkthrough/SKILL.md` no devuelve nada (AC8).
- `rg -n -F "nothing is repeated for them" content/skills/flow-walkthrough/SKILL.md` no devuelve nada (AC4).
- `rg -n -F "working tree on" content/skills/flow-report/assets/skeleton-walkthrough.html` devuelve la línea del llenado parcial (AC7).
- `awk '/Open the page/{f=1}f' content/skills/flow-walkthrough/SKILL.md | rg -c -F "absolute path"` imprime 1 o más, y leerla confirma que la ruta va siempre y que nombra el caso en que el host no puede abrirla (AC8).
- `python3 -I content/skills/flow-report/scripts/self_check.py content/skills/flow-report/assets/skeleton-walkthrough.html` sale con 0.
- `wc -w content/skills/flow-walkthrough/SKILL.md` da menos de 1.800.

### T4 — M3 (orden y commit) en `flow-build`, `verification.md`, `delivery-decisions.md` y `ui-gate.md`

- [x] La secuencia antes de pedir validación, el commit de tareas con interfaz (D1-A) y los punteros, como fija [design.md](design.md) M3 «Casa del orden», «Commit (D1-A)» y «Punteros a la casa del orden».

**Closes:** AC9, AC10.

**Depends on:** T2, porque las dos editan `flow-build/SKILL.md` y `ui-gate.md`.

**Locations:**

- `content/skills/flow-build/SKILL.md`: `:61`, `:71` y `:73`.
- `content/skills/flow-build/references/verification.md`: `:33` y `:43`.
- `content/skills/flow-plan/references/delivery-decisions.md:9`.
- `content/skills/flow-walkthrough/references/ui-gate.md`: la viñeta de `:25` y la última frase de `:26`. El resto de `:26` no se toca.

**Execution:** delegada a `hive-write-docs`; empieza cuando T2 está verificada.

**Test approach:** `check`, con las búsquedas de abajo.

**Changes:** las de M3 en [design.md](design.md). La revisión de código conserva su momento.

**Verification:**

- `rg -n -F "first rendered view and the UI gate" content/skills/flow-build/SKILL.md` devuelve la regla de commit. Leer ese párrafo confirma que nombra `automatic`, `UI verification` `none`, las tareas que comparten archivos, el commit adicional por tarea reabierta sin amend, y la pausa (AC9).
- `rg -n -F "before asking the user" content/skills/flow-build/SKILL.md content/skills/flow-build/references/verification.md content/skills/flow-plan/references/delivery-decisions.md content/skills/flow-walkthrough/references/ui-gate.md` devuelve la secuencia en `flow-build` y un puntero en cada uno de los otros tres (AC10).
- `rg -n "hive-review-security" content/skills/flow-build/references/verification.md content/skills/flow-walkthrough/references/ui-gate.md content/skills/flow-plan/references/delivery-decisions.md` muestra que, salvo la revisión que nombra el veredicto del revisor de código, ninguno fija la revisión de seguridad tras la validación; leer la secuencia de `flow-build` confirma que conserva la autorización de `verification.md` y que repite la revisión de interfaz cuando un arreglo de seguridad toca la UI (AC10). Leer la regla de commit confirma que el commit adicional cubre también los arreglos de la suite y de seguridad sobre una tarea ya comprometida (AC9).
- `wc -w content/skills/flow-walkthrough/references/ui-gate.md` da menos de 880.

### T5 — Sitio, `CHANGELOG.md` y notas de mantenedor

- [x] La página `flows/build.md` describe los cuatro momentos, `CHANGELOG.md` tiene la entrada, `agent-delivery.md` documenta el aviso de hijo detenido de Claude Code, e `instruction-resources.md:17` queda alineado con la invariante de `global.md`.

**Closes:** AC11.

**Depends on:** T1, T2, T3 y T4.

**Locations:**

- `site/src/content/docs/flows/build.md`: `:36`, `:37`, `:64` y `:73`.
- `site/src/content/docs/flows/close.md:8,63`, `research.md:47` y `plan.md:68`: la línea de contexto donde hoy describen la pausa, la tarjeta o el bloque «Pausa».
- `CHANGELOG.md`, sección sin publicar.
- `_support/docs/architecture/agent-delivery.md`, tabla `:145-158`.
- `_support/docs/architecture/instruction-resources.md:17`.

**Execution:** delegada a `hive-write-docs`. Las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas.

**Test approach:** `check`.

**Verification:**

- `rg -n -F "context line" site/src/content/docs/flows/build.md site/src/content/docs/flows/close.md site/src/content/docs/flows/research.md site/src/content/docs/flows/plan.md CHANGELOG.md` devuelve una línea en cada archivo (AC11).
- `rg -n -i "stopped|opens the page|UI review" site/src/content/docs/flows/build.md` muestra el hijo detenido, la página que se abre sola y el commit tras la revisión de interfaz (AC11).
- `rg -n -F "e2c8d39f" _support/docs/architecture/agent-delivery.md` devuelve la entrada (AC11).
- `pnpm -C site build` sale con 0.

## Verificación y revisión humana

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` con la carpeta del cambio, la tarea y la base | Parte de la construcción |
| Pruebas del repositorio | Candidato final, local | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasan | Parte de la construcción |
| Sitio | Candidato final | `pnpm -C site build` sale con 0 | Parte de la construcción |
| Coherencia del conjunto | Candidato final, antes de pedir tus pruebas | `hive-review-harness` audita el diff de `content/` contra la base, solo lectura. Busca contradicciones, reglas que dicen lo mismo con distinto alcance, momentos con más de una casa, anclas rotas y punteros que repiten pasos. Las carencias materiales se corrigen y se reaudita una vez | Responde a tu pedido de evaluarlo en conjunto. Es una auditoría de guía, no la revisión de código que R1-A declinó |
| Despliegue de prueba | Tras los commits, solo si lo pides | `hive update --dry-run` desde la rama sale con 0, y `hive update` solo cuando lo pidas | A pedido |
| Tus pruebas | Sesiones reales nuevas con la versión desplegada | Tú observas un turno de notificación, la línea de contexto, la página que se abre sola y el commit de una tarea de UI | Tuyo; después decides el merge |

Recorrido local: no aplica, porque el cambio es texto de guía sin superficie ejecutable.

## Estado de revisión y avance

- Plan escrito el 2026-10-18 sobre la base `8c15f59a`.
- Ronda 1 de revisión sobre `51751a7641f2`:
  - coherencia: 6 hallazgos, 138,6 s, 131 k tokens;
  - orden y ejecución: 10 hallazgos y 3 bajos, 200,9 s, 141 k tokens.
  - Todos aceptados; las correcciones están en [design.md](design.md#revisión-del-plan).
- Re-revisión única sobre `fdddc6a08536`:
  - coherencia: 54 s, 148 k tokens. Resolvió H2–H6. Abrió R1 (T1 decía lo contrario de M1) y R3 (alcance de la nota de una línea), aplicadas como el revisor propuso. Dejó R2 abierto: un resultado que no cambia una pregunta abierta.
  - orden y ejecución: 53 s, 160 k tokens. Tres hallazgos (comprobación de seguridad en T4, arreglos sobre una tarea ya comprometida, búsqueda de T3 ya verdadera en la base), aplicados como el revisor propuso.
- R2 va al usuario como decisión D2; no hay otra ronda.
- D2-A (2026-10-18): con una pregunta abierta, un resultado que no la cambia toma la salida 3. B1-A: construir ahora.
- El cambio se hace en el mismo turno en que el usuario pidió la evaluación en conjunto, que reorganizó C1–C6 en los momentos M1–M4 antes de escribir las tareas.
- **Construcción (2026-10-18), commits:** T2 `ebb2eab5`, T1 `999899fb`, T3 `db72fc02`.
- **Implementación:** `hive-write-docs` para T2 y T3 y el hilo principal para T1. Verificación: `hive-verify-task`, una ronda cada una, todo `met`. Modelos: configurados, no observados; el host no reportó el modelo de los hijos.
- **Correcciones del plan durante la construcción:**
  - La comprobación de T1 buscaba «context line» en `questions.md`, que solo lleva el ancla; ahora busca el ancla.
  - Tras la verificación de T1, «with the in-progress item below» pasó a «opening with the in-progress item that the next condition defines», por la ambigüedad que señaló el verificador; no cambia ningún criterio.
- **Tamaño de `global.md`:** de 4.699 a 4.774 palabras (+75) y de 30.717 a 31.133 bytes.
- **Tope de `flow-walkthrough/SKILL.md`:** 1.794 de 1.800 palabras, con 6 de margen.
- **Commits:** T4 `4dd7949c` y T5 `907cbe21`, verificadas en una ronda con todo `met`.
- **Ajustes antes de comprometer T5:** las observaciones del verificador. `instruction-resources.md` alineado con `global.md`; la condición de frontera en `CHANGELOG.md`; el texto literal del aviso de Claude Code, leído del transcript.
- **Puertas sobre el candidato:** las tres pasan.
  - `go vet ./...` y `go test ./...` (salida 0).
  - Las pruebas de skills: 23 OK.
  - `pnpm -C site build` (salida 0).
- **Auditoría de coherencia (`hive-review-harness`, 232 s, 178 k tokens):**
  - Encontró 14 puntos: H1 y H2 graves, H3–H11 medios, H12–H14 bajos.
  - Se corrigieron 12. H14 y la tensión menor con `hive-verify-task.md:14` quedan anotadas sin cambio.
  - Re-auditoría única (40 s, 192 k tokens): confirmó las correcciones y abrió R1 (trabajo del hijo todavía vivo), R2 (repeticiones en la puerta del candidato aceptado) y R3 (decisión nueva sin pregunta previa).
  - Las tres se aplicaron como las propuso el auditor; las correcciones de la auditoría están en el commit `f8c0f8bd`.
- **Despliegue:** `hive update --dry-run --source .` sale con 0 sobre `f8c0f8bd`. Desplegado en local por P1-A: `hive update --out` y `hive apply`; la release registra `f8c0f8bd` el 2026-10-19 a las 02:36:45Z.
- **Tamaños finales:**
  - `global.md`: 4.785 palabras (+86 sobre la base) y 31.201 bytes (+484).
  - `flow-walkthrough/SKILL.md`: 1.797 palabras, tope 1.800.
  - `ui-gate.md`: 875 palabras, tope 880.
