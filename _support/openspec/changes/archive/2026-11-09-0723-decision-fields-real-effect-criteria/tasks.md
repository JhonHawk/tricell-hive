# Tareas

Base: `development` en `dab27f73`; sin rama mientras la entrega sea solo local. T1, T3, T4, T5 y T7 corren en el hilo principal: cada una edita tres archivos o menos y su redacción sigue las reglas de autoría de `AGENTS.md`. T6 va a un hijo `hive-write-docs` porque edita hasta siete archivos de documentación a partir de textos ya fijados. No hay T2: la revisión del plan la eliminó. Tampoco T8: se añadió en la validación y se retiró tras su revisión de código.

### T1 — Decisión con campos en la guía global

- [x] «Questions and routes» exige `Situation:`, `If you choose it:` e `I recommend <letter>:` en cada decisión, prohíbe citar reglas como razón y pide la cuenta de decisiones restantes en la línea de espera; la tarjeta reparte situación, consecuencias y razón
- [x] «Delivery» apunta a esas líneas en vez de repetir dónde van la razón, lo que toca el cambio y sus notas
- [x] `viewer-publication.md` exceptúa la opción «Open the viewer» de la línea de consecuencia y dice qué se repite al volver a preguntar

**Closes:** AC1, AC2, AC4, AC10, AC13, AC14

**Depends on:** ninguno

**Locations:** `content/guidance/global.md`, `content/skills/flow-walkthrough/references/viewer-publication.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** el punto del formato de decisión, el punto de la tarjeta y las frases de «Delivery», como fija `design.md` «Forma propuesta»; la línea de la opción del visor. No cambian los IDs, la estrella, la línea de contexto, los hallazgos, el bloque de siguiente paso ni la respuesta de compactación.

**Verification:**
- AC1: `rg -n "Situation:" content/guidance/global.md`, `rg -n "If you choose it:" content/guidance/global.md` y `rg -n "I recommend" content/guidance/global.md` encuentran cada uno al menos una línea, y `rg -n "what you verified and what to check" content/guidance/global.md` encuentra una.
- AC2: `rg -n "quotation of a rule" content/guidance/global.md` encuentra una línea.
- AC10: `rg -n "still to come" content/guidance/global.md` encuentra una línea, y esa línea contiene `waiting line`.
- AC4: `rg -n "its description carries the explanation" content/guidance/global.md` no encuentra nada; `rg -n "the question carries the situation" content/guidance/global.md` y `rg -n "ending with the reason" content/guidance/global.md` encuentran una línea cada uno.
- AC13: `rg -n "next to the star|one line on what the change touches|in the context" content/guidance/global.md` no encuentra nada dentro de «Delivery».
- AC14: `rg -n "only three rules" content/skills/flow-walkthrough/references/viewer-publication.md` no encuentra nada; `rg -n "four rules" …` y `rg -n "If you choose it" …` sobre el mismo archivo encuentran una línea cada uno.
- Tamaño: `awk '/^## Questions and routes/{f=1;next} /^## /{f=0} f' content/guidance/global.md | wc -c` (hoy 11,047) y `wc -c content/guidance/global.md` (hoy 44,450). El reporte nombra ambas cifras. Por encima de 11,700 o de 45,150 no se recorta ningún requisito: se dice en la validación y el usuario decide.
- `go test ./tests/content/... ./tooling/...` pasa, incluido el reparto de Antigravity.

### T3 — Criterios que dicen dónde se observan

- [x] `flow-plan` «Acceptance criteria» exige `Observed in:`, define efecto real y sustituto, dice que un sustituto nunca cierra un criterio y qué hace la tarea que no alcanza el lugar; la plantilla lo muestra
- [x] `flow-walkthrough` limita los datos stub a la primera vista del build

**Closes:** AC11

**Depends on:** ninguno

**Locations:** `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/assets/change-folder/proposal.md`, `content/skills/flow-walkthrough/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en «Acceptance criteria», la frase de cobertura reescrita y el párrafo nuevo de `design.md` «Forma propuesta»; en «Revise a plan in progress», `Observed in:` entre los marcadores que quedan en inglés; la viñeta de criterio de la plantilla termina en `· Observed in: <where its real effect shows>`; la frase de `flow-walkthrough/SKILL.md:10`, que cuenta para AC12 y que T7 verifica.

**Verification:** `rg -n "Observed in" content/skills/flow-plan/SKILL.md content/skills/flow-plan/assets/change-folder/proposal.md` encuentra la regla, la frase de marcadores y la plantilla; `rg -n "never closes a criterion" content/skills/flow-plan/SKILL.md`, `rg -n "not verified until" content/skills/flow-plan/SKILL.md`, `rg -n "names the stage that reaches its place" content/skills/flow-plan/SKILL.md` y `rg -n "does not own" content/skills/flow-plan/SKILL.md` encuentran una línea cada uno; `wc -c content/skills/flow-plan/SKILL.md` da 16,000 o menos; `go test ./tests/content/...` pasa.

### T4 — check_change.py comprueba el marcador

- [x] El script imprime un hallazgo por criterio activo sin `Observed in:` y lo omite con `--archive`

**Closes:** AC7

**Depends on:** T3

**Locations:** `content/skills/flow-plan/scripts/check_change.py`, `tests/skills/change_check_test.py`, `tests/skills/tasks_test.py`

**Execution:** hilo principal

**Test approach:** tdd: el formato de la carpeta de cambio es un contrato persistido. Primero el caso «criterio activo sin marcador imprime un hallazgo» en `change_check_test.py`, que falla sobre el script actual. Con él se añaden los casos que fijan lo que no debe dar hallazgo y que el script actual ya cumple: marcador en sus tres formas, marcador en una línea de continuación de la viñeta, criterio retirado sin marcador, y sin marcador con `--archive`.

**Changes:** una comprobación nueva junto a `criteria_findings`, que lee la viñeta del criterio y sus líneas de continuación, y su línea en el docstring. Las comprobaciones existentes y los códigos de salida no cambian. Los criterios de prueba de `change_check_test.py` y de `tasks_test.py:179-198`, que hoy esperan salida 0 sin marcador, ganan `· Observed in: <lugar>`.

**Verification:** el caso nuevo falla antes de cambiar el script y pasa después; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa; `python3 content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/decision-fields-real-effect-criteria` termina en 0; `go test ./tooling/...` pasa.

### T5 — Build y los verificadores

- [x] `flow-build`, `hive-verify-task` y `hive-verify-change` tratan el criterio cuyo lugar no se alcanzó, o cuya evidencia es de un sustituto, como sin verificar

**Closes:** AC12

**Depends on:** T3

**Locations:** `content/skills/flow-build/SKILL.md`, `content/agents/quality/hive-verify-task.md`, `content/agents/quality/hive-verify-change.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** un puntero en cada archivo a «Acceptance criteria» de `flow-plan`, que es el hogar de la regla, con lo que cambia en ese punto: qué se nombra en `--unverified`; y que la evidencia de un sustituto, o un criterio que nombra un entorno desplegado, dan `cannot verify` como resultado esperado. No se repite la regla.

**Verification:** `rg -n "Observed in" content/skills/flow-build/SKILL.md content/agents/quality/hive-verify-task.md content/agents/quality/hive-verify-change.md` encuentra una mención en cada uno; `rg -n "only evidence for a criterion is a check against a substitute" content/agents/quality/hive-verify-task.md` encuentra una línea; `wc -c content/skills/flow-build/SKILL.md` da 16,000 o menos; `go test ./tests/content/...` pasa, incluido `agent_boundary_test.go`.

### T7 — Verify y el recorrido de deploy

- [x] `flow-verify` comprueba cada criterio en el lugar que nombra, y el recorrido funcional de `flow-deploy`, cuando el usuario lo pide, cierra los de un entorno desplegado

**Closes:** AC12

**Depends on:** T3

**Locations:** `content/skills/flow-verify/SKILL.md`, `content/skills/flow-deploy/references/deployment-checks.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** en el paso 1 de `flow-verify`, «Check each `AC<n>` by its observation» nombra el lugar de `Observed in:` y dice que un criterio de un entorno desplegado queda sin verificar en su reporte; en «Functional walk» de `deployment-checks.md`, que el recorrido, cuando corre, cierra los criterios cuyo lugar es ese entorno y que, cuando no corre, siguen sin verificar y el reporte los nombra. Cuándo corre el recorrido y la petición de sus efectos no cambian.

**Verification:** `rg -n "Observed in" content/skills/flow-verify/SKILL.md content/skills/flow-deploy/references/deployment-checks.md` encuentra una mención en cada uno; `rg -n "stub" content/skills/flow-walkthrough/SKILL.md | rg -v "first view"` no imprime nada; `wc -c` de ambos archivos da 16,000 o menos; `go test ./tests/content/...` pasa.

### T6 — Sitio, documentos de arquitectura y registro de cambios

- [x] Las páginas de plan y verify y los dos documentos de arquitectura describen el formato nuevo
- [x] `CHANGELOG.md` lleva la entrada y la nota de migración

**Closes:** AC9, AC15

**Depends on:** T1, T3, T4, T5, T7

**Locations:** `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/deploy.md`, `site/src/content/docs/flows/research.md`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/instruction-resources.md`, `CHANGELOG.md`

**Execution:** `hive-write-docs`

**Changes:** las páginas describen el comportamiento y enlazan la skill, sin copiar sus reglas: qué líneas trae una decisión y que cada criterio dice dónde se observa. `research.md` y `deploy.md` se editan solo si su mención de las tarjetas o del recorrido funcional queda inexacta. `agent-delivery.md` añade el cambio a su historia del formato de preguntas; `instruction-resources.md` actualiza lo que dice que contiene la sección. `CHANGELOG.md` gana en `[Unreleased]` la entrada y la nota de migración que fija `design.md`, con el formato de `CONTRIBUTING.md`. `README.md` y `llms.txt` no cambian: no enuncian el formato.

**Verification:** `rg -n "Observed in|If you choose" site/src/content/docs/flows _support/docs/architecture` encuentra las menciones en plan, verify y `agent-delivery.md` (AC9); `rg -n "Observed in" CHANGELOG.md` encuentra la entrada bajo `[Unreleased]` (AC15); el sitio construye en local con el comando de `site/package.json`, y las páginas editadas se abren.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar la tarea, en local | El comando de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierta por la implementación |
| Tamaños | Candidato final, en local | Bytes de «Questions and routes» y de `global.md` antes y después; cada `SKILL.md` y referencia ≤ 16,000 | Cubierta por la implementación; por encima de 11,700 o 45,150 decide el usuario |
| Lugares del barrido | Candidato final, en local | Cada lugar de `design.md` «Impacto» editado o confirmado sin cambio, listado en el reporte | Cubierta por la implementación |
| Revisión de código | Candidato final, antes de la parada de validación | `/code-review` de Claude Code sobre el diff de los archivos de «Impacto» (D8 A); la rúbrica de revisión queda fuera | La respuesta de revisión |

Recorrido humano: leer en el diff el punto del formato de decisión de `global.md` y el párrafo nuevo de «Acceptance criteria»; son los dos textos que solo el usuario puede juzgar.

## Checkpoint

- Updated: 2026-11-09
- Stage: close · cierre en pausa: entrega solo local, árbol de trabajo verificado y sin commit
- Last commit: uncommitted
- Open findings: H8 (P3) sin corregir: check_change.py calcula dos veces qué criterios están activos; Hueco de flow-close: ticket #437 (D12 A); Las correcciones de la segunda revisión de código no pasaron por otra revisión; tooling/cli falló una vez y pasó en cinco corridas siguientes; no se guardó su salida
- Next: cuando el usuario lo pida: commit con las observaciones de «Contexto verificado», revisión de las correcciones no revisadas, despliegue con hive update y seguimiento con la rúbrica

### Evidence

- go test -count=1 ./tests/content/... ./tooling/... · pass · sin commit, tras T1
- python3 -I -m unittest discover -s tests/skills -p '*_test.py' · pass (85) · sin commit, tras T4
- check_change.py sobre esta carpeta · exit 0 · sin commit, tras T4
- go vet ./... · pass · sin commit, candidato final
- go test -count=1 ./... · pass (22 paquetes) · sin commit, candidato final
- python3 -I -m unittest discover -s tests/skills -p '*_test.py' · pass (85) · sin commit, candidato final
- site: npm run build · 17 páginas · sin commit (hijo T6)
- go vet ./... · pass · sin commit, tras las correcciones
- go test -count=1 ./... · pass (22 paquetes) · sin commit, tras las correcciones
- python3 -I -m unittest discover -s tests/skills -p '*_test.py' · pass (87) · sin commit, tras las correcciones
- site: npm run build · 17 páginas · sin commit, tras las correcciones
- python3 -I -m unittest discover -s tests/skills -p '*_test.py' · pass (89) · sin commit, candidato final
- site: npm run build · completo · sin commit, candidato final

## Estado de la revisión y avance

2026-11-08 — Revisión del plan, ronda 1 aplicada: T2 eliminada, T7 añadida, T4 con `tasks_test.py`.
2026-11-08 — Revisión del plan, ronda 2 aplicada: T5 con `hive-verify-change`, T6 con la página de deploy y `CHANGELOG.md`, verificaciones con una búsqueda por etiqueta.
2026-11-08 — D10 respondida: T7 y T6 no añaden ninguna declaración del recorrido de deploy.
2026-11-09 — T1 · uncommitted · unverified: none
2026-11-09 — T3 · uncommitted · unverified: none
2026-11-09 — T4 · uncommitted · unverified: go test ./tooling/... se corre con la suite completa al final
2026-11-09 — T5 · uncommitted · unverified: none
2026-11-09 — T7 · uncommitted · unverified: none
2026-11-09 — T6 · uncommitted · unverified: las páginas editadas no se abrieron en un navegador; el sitio construye
2026-11-09 — T8 · uncommitted · unverified: none
2026-11-09 — T8 reabierta: su revisión de código encontró 10 problemas (contradice el requisito vigente «Promotions retain nothing», alcanza a trabajo abandonado o sin recorrido, y recortó una cláusula de regla). `flow-close` y la página de Close se devolvieron a su estado de `HEAD`; la decisión vuelve al usuario como D12.
2026-11-09 — D12 A: la tarea añadida en la validación se retira de este cambio y su hueco va a un ticket.
