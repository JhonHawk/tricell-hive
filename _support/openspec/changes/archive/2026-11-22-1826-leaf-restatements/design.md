# Design

## Approaches

Solo hay una forma razonable: en cada lugar fuera de la casa de una regla, dejar su momento y un enlace con ancla (D19), o un invariante completo y corto cuando debe actuar antes de leer la hoja; y resolver en la casa cada contradicción. Mover reglas de casa no se justifica en ningún caso: la clasificación solo encontró casas mal puestas en S23, S29 y S32, que se resuelven con un puntero en un sentido. La alternativa de numeración tipo constitución y la de sacar el índice de `global.md` se descartaron en D19 y D20.

## Classes

- **P** puntero: momento + enlace. Se queda.
- **I** invariante: debe actuar antes de leer la hoja; se queda completo y corto.
- **E** excepción local o paso propio del flujo. Se queda.
- **D** copia fiel. Pasa a puntero.
- **X** copia parcial o divergente. Pasa a puntero; si contradecía su casa, manda la casa salvo lo que deciden D15–D17, D22.

## Global (`content/guidance/global.md`, sobre `d4f1eedb`)

| ID | Línea | Hoja | Clase | Acción |
| --- | --- | --- | --- | --- |
| G1 | 14 | external-effects, unattended-effects, checks | I | Se queda entero, con «of that work item's plan» y «it is none of the cases that round keeps with their own question»; solo se acortan las etiquetas de los enlaces si hace falta. |
| G3 | 19 | reasons-and-blockers | X | Puntero: un defecto que bloquea trabajo autorizado pertenece a ese trabajo; la hoja fija cuándo arreglarlo y cuándo reportarlo. |
| G10 | 50 («Evidence») | principles/verify-current-state | D | Pasa al índice, bajo el dominio `principles`; se borra la sección «Evidence». |
| G13–G15 | 107–111 («Questions and routes») | questions, reasons-and-blockers, report-endings | X | Queda una línea invariante (D21): preguntar solo lo que es del usuario y cambia lo que sigue; solo su respuesta o un comando que escribe arranca otro flujo, salvo las continuaciones que lista [report endings]; con enlaces a [questions], [reasons and blockers] y [report endings]. Sin ningún formato. Manda `unanswered.md` (D23) y `reasons-and-blockers.md:5` (D24). |
| G16 | 115 | delivery-ask, foreign-commits | X + D | Puntero a las dos hojas. |
| G17 | 116 (entrada) | merge-gate | X | «…en todo modo, la aprobación pendiente de un revisor humano retiene el merge ([merge gate] fija qué lo levanta):» (D15). Los bullets de modos se quedan. |
| G18 | 125 (cola) | checkpoint | X | «…omite ese guardado y, antes de cualquier llamada a memoria, lee y sigue el [checkpoint] como fija para después de una compactación.» |
| G2, G4–G9, G11, G12 | 18, 20, 21, 32, 39, 40, 46, 55, 105 | varias | I o P | Se quedan. |

## Index (D20)

Un subtítulo `###` corto por dominio (no `##`: el test del índice y el divisor de Antigravity cortan en `## `), en el orden de `domains.json` (`messages`, `git`, `plan`, `build`, `delegation`, `ui`, `review`, `business`, `close`, `local`, `tracker`, `principles`), y una hoja por línea con el disparador que ya cita su primera línea. Las líneas que hoy juntan varias hojas se separan. `principles` lista solo `verify-current-state` y el índice de principios. Las líneas añadidas por separar se compensan con lo que libera «Global»; AC7 fija el techo en 28.000 B (estimado ~27.850 B).

## Roles (`content/agents/`)

| ID | Lugar | Hoja | Clase | Acción |
| --- | --- | --- | --- | --- |
| A1 | design/hive-design-architecture.md:16 | build/stored-data | X | Puntero; quitar «and recovery» del tamaño por conteo. |
| A2 | development/hive-build-data.md:10 | build/stored-data | D | Puntero, conservando «before dependent work». |
| A4 | review/hive-review-code.md:12 | build/defect-fix, plan/reuse-before-writing | X | Puntero a las dos hojas; sin lista truncada de guardas. |
| A5 | review/hive-review-code.md:14, quality/hive-verify-task.md:12 | `global.md` «Language» | X | Puntero a «Language». |
| A6–A8 | design/hive-design-ui.md:15, quality/hive-verify-change.md:12, review/hive-review-ux.md:12 | ui/ui-verification#browser-tool, ui/ui-verification/device-tool.md | X / D | Puntero con ancla `ui-verification.md#browser-tool`; el dispositivo enlaza el archivo `ui-verification/device-tool.md`, que no es una sección. |
| A9 | review/hive-review-ux.md:12 | ui/ui-plan | X | Puntero a las decisiones de UI. |
| A10 | review/hive-review-ux.md:14 | ui/a11y-audit | D | Quitar la cola copiada. |
| A11 | las 6 hive-build-* | principles/index | D | «Antes de código nuevo, lee [code principles]»; el retorno del principio queda en la hoja de checkpoint. |
| A3, A12–A14 | ops/hive-build-infra.md:15; development/hive-build-frontend.md:17; review/hive-review-plan.md:11, 15; review/hive-review-prd.md:15; review/hive-review-harness.md:16; review/hive-review-code.md:10, 20; review/hive-review-security.md:10; quality/hive-write-tests.md:13 | varias | E / P | Se quedan. |

## Skills

| ID | Lugar | Hoja | Clase | Acción |
| --- | --- | --- | --- | --- |
| S1 | create-design/SKILL.md:27 | ui/design-stage | X | Puntero a `design-stage.md` nombrando su punto «Resolving the file in the stage» (es un bullet, no tiene ancla). |
| S2 | hive-mode/SKILL.md:34 | ui/design-stage | X | Quitar «that adds a screen or recomposes one». |
| S3 | flow-plan/SKILL.md:27 | plan-handoff, change-folder | X | «…su resumen es el que fija [plan handoff]». |
| S4 | flow-close/SKILL.md:16 | ui/viewer-close | X | Los estados del readback se citan, no se listan. |
| S6 | flow-verify/SKILL.md:24 | build/checks#local-ci | D | Puntero. |
| S7 | hive-mode/SKILL.md:28 | messages/report-endings | X | Quitar la frase repetida. |
| S8, S12, S21 | data-change.md:10, flow-deploy/SKILL.md:24, flow-plan/SKILL.md:14-15 | plan/external-effects | X | Puntero al effects round; flow-deploy conserva su frase de producción y corrige la etiqueta del enlace. |
| S9 | flow-plan/SKILL.md:13 | plan/external-facts | X | Puntero. |
| S10, S22 | harness-audit/references/local-instructions.md:9, 37, 38, 40 | local/local-write, local/local-record | X / D | Punteros. |
| S14 | data-change.md:8 | build/stored-data | X | Puntero. |
| S15 | flow-plan/SKILL.md:15 | git/delivery-ask | X | «…cuando [delivery questions] dice que se preguntan.» |
| S17 | flow-build/SKILL.md:10 | plan/checkpoint | X | «…retomar como fija, después un `git status`.» |
| S18 | flow-build/references/spec-delta.md:5, 7 | plan/change-folder | X / D | Quitar la lista de contenidos; puntero al paso 3 de «Revise a plan in progress». |
| S20 | hive-mode/playbooks/bugfix.md:10 | build/tests | X | Puntero. |
| S23 | flow-deploy/SKILL.md:12 | close/close-report | D | «Sale con las observaciones que entrega el paso 8.» |
| S24 | create-design/references/draw.md:10 | ui/ui-verification | D | Quitar el paréntesis. |
| S25 | create-design/references/rule-cards.md:12 | business/business-rules | X | Puntero a la definición de «stated». |
| S26 | hive-mode/playbooks/investigate.md:14, 16 | messages/report-endings | D | Borrar :14; :16 conserva el puntero a la fila de análisis. |
| S27 | hive-mode/playbooks/ticket-status.md:14 | close/close-hygiene | X | Puntero a su búsqueda del paso 1. |
| S28, S29, S30 | flow-plan/references/plan-handoff.md:14, 17, 21 | report-endings, external-facts, external-effects | D / X | Punteros. El puntero de vuelta en `external-facts.md:13` se quita en T2 (es una hoja). |
| S32 | flow-deploy/SKILL.md:34 | tracker/ticket-state | X | Puntero a la línea de `Completion` (D22). |
| Cercanos | bugfix.md:8, 9, 11, 12, 13, 15; data-change.md:12, 13, 15; investigate.md:16; ticket-status.md:7; pr-review-deploy.md:12 (momento viejo de foreign-commits, D16); flow-plan/SKILL.md:17; hive-mode/SKILL.md:27; git-workflow/SKILL.md:24 | tests, defect-fix, close-report, report-endings, merge-gate | D / X | Punteros; `hive-mode:27` termina como fila de análisis (D17). |
| S5, S11, S13, S16, S19, S31 | flow-close/SKILL.md:20; flow-verify/SKILL.md:14, 21; create-issue/SKILL.md:3; flow-verify/SKILL.md:25; flow-close/SKILL.md:12; git-workflow/SKILL.md:15, 16 | varias | P / E | Se quedan. |

## Leaves

| ID | Lugar | Casa | Clase | Acción |
| --- | --- | --- | --- | --- |
| L1 | git/pr-reviews.md:9 | git/foreign-commits.md | P | Ya es puntero; la casa suma el momento antes del merge o de `ready` (D16). |
| L18 | plan/external-facts.md:13 | flow-plan/references/plan-handoff.md | X | Quitar el puntero de vuelta circular (S29). |
| L5 | ui/ui-example.md:5 | business/business-record.md:22-24 | X | «copia las líneas del registro al plan como fija [rule record]». |
| L7 | close/close-report.md:24 | close/close-modes.md | X | «Los cierres sin fila propia, como los que lista la segunda fila de [close modes], usan la primera.» |
| L9 | ui/walkthrough.md:5, ui/viewer.md:3 | ui/viewer.md | X | walkthrough apunta a viewer para hijos y publicación no disponible; viewer quita la frase de walkthrough. |
| L10 | close/close-hygiene.md:11 | close/close-modes.md | X | «Mantén pendiente cada elemento que necesita al usuario, con su pregunta, como fija [close modes].» |
| L11 | close/close-report.md:11 | build/verification-floor.md | D | Puntero. |
| L12 | tracker/trackers/github.md:11 | tracker/trackers.md | D | Puntero a «Comment and attach». |
| Tiempo | tracker/trackers/github.md:15, linear.md:15 | tracker/trackers/time.md:3 | D | «Time: [time] lo fija.» |
| Etiquetas | ui/walkthrough.md (`[viewer publication]` ×3), review/review-findings.md:5 (`[flow-close]`) | — | — | Etiquetas que nombran su destino real. |
| ticket-state | tracker/ticket-state.md:10 | `global.md:21` | D | Puntero a «Scope and authorization»; la línea de `Completion` dice lo que fija D22. |
| D17 | messages/report-endings.md, fila «Assessment of a ticket» | `hive-mode` paso 5 | X | La fila deja de devolver a hive-mode una evaluación sin trabajo confirmado. |
| L2–L4, L6, L8, L13–L17 | plan/external-effects.md:12; tracker/trackers/time.md:3; principles/index.md:13; close/close-report.md:43, 49; review/review-findings.md:5; local/local-write.md:5; local/local-script.md:5; build/verification-floor.md:20, 21; build/defect-fix.md:3, 8 | varias | P / E | Se quedan. |

## Overlap scan

Para verificar AC3 sin depender de esta sesión: para cada archivo `.md` de `content/` y cada hoja de `hive-core/<dominio>/*.md` que ese archivo enlaza (relativo o `skill:`), contar las secuencias de 6 palabras que comparten, tras pasar a minúsculas, reemplazar cada enlace `](…)` por `]`, cada `` `código` `` por `CODE`, tokenizar con `[a-z0-9']+` y excluir la primera línea «Read …» de la hoja; un par con 3 o más secuencias es un caso. La build reproduce el script en el host's temporary location; no queda en el repositorio (D13).

## Pairs the scan keeps

AC3 se comprueba por par (archivo de origen → hoja), no por línea, porque las líneas se mueven al reescribir. Tras el cambio, el escaneo solo puede devolver estos pares, cada uno P, E o I:

- `create-design/SKILL.md` → `ui/design-stage` (descripción y ruteo, :3, :25)
- `create-design/references/draw.md` → `ui/ui-verification` (:9, excepción local)
- `create-design/references/rule-cards.md` → `business/business-rules` (:3, puntero)
- `create-issue/SKILL.md` → `plan/scope-changes` (S13)
- `flow-build/SKILL.md` → `plan/checkpoint` (:19, única casa del uso de `done`)
- `flow-close/SKILL.md` → `close/close-hygiene`, `close/close-modes` (S5, S19)
- `flow-plan/references/plan-handoff.md` → `messages/report-endings` (:7, :23, excepción local)
- `flow-verify/SKILL.md` → `ui/walkthrough`, `git/merge-gate` (S11, S16)
- `git-workflow/SKILL.md` → `git/foreign-commits` (S31)
- `harness-audit/references/local-instructions.md` → `local/local-write` (:30, :74, paso propio del modo)
- `hive-mode/SKILL.md` → `messages/report-endings` (:16, otra situación)
- `hive-mode/playbooks/data-change.md` → `build/stored-data` (:5)
- `hive-mode/playbooks/investigate.md` → `messages/report-endings` (:18)
- `hive-core/close/close-modes.md` → `close/close-report` (:8, :9, ejes distintos de la misma tabla)
- `hive-core/ui/ui-example.md` → `business/business-record` (:3, puntero)
- `hive-core/close/close-report.md` ↔ `review/review-findings.md` (L6, L8)
- `hive-core/build/verification-floor.md` → `build/tests`, `close/close-report` (L15, L16)
- `hive-core/build/defect-fix.md` → `build/tests` (L17)
- `hive-core/principles/index.md` → `principles/verify-current-state` (L4)
- `hive-core/plan/external-effects.md` → `plan/external-facts` (L2)
- `hive-core/tracker/trackers/time.md` → `tracker/trackers` (L3)
- `hive-core/local/local-write.md` ↔ `local/local-script.md` (L13, L14)
- `content/guidance/global.md` → las hojas de los invariantes que se quedan (G2, G4–G9, G11, G12, G1) y el índice
- los roles A3, A12–A14

## Impact

~45 archivos de `content/`, `AGENTS.md` («Guidance structure»: la regla de citar con ancla), `README.md:476, 481`, `llms.txt:32`, `_support/docs/architecture/{instruction-resources.md:29, 33, tool-dependencies.md:52, agent-delivery.md:258}`, `_support/docs/harness-engineering/{rule-enforcers.md:38, 42, 56, session-review-rubric.md:51}`, `content/skills/hive-core/close/close-report.md:30` («the tickets waiting on a check»), `CHANGELOG.md` y los tests de contenido que registran frases movidas (`single_home_test.go`, `index_trigger_test.go`, `leaf_links_test.go`). Los paquetes de etapa se regeneran solos. Las sesiones abiertas necesitan reiniciar tras el despliegue.

## Open items outside this plan

- «candidate» nombra dos cosas: una entrada de `HIVE.local.md` para la guía versionada (`local-write.md:5`, `local-script.md:5`) y una repetición en `close-report.md:53`.
- `external-effects.md:7` y `external-facts.md:3` listan distinto lo que la entrega ya autoriza.
- La lectura obligatoria de `HIVE.local.md` al inicio se cumple poco (`research/mandatory-read-compliance`).

## Recovery

Revertir el commit restaura cada frase; el siguiente `hive update` despliega lo revertido.
