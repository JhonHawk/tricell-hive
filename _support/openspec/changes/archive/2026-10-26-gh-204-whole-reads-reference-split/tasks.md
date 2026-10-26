# Tareas — gh-204

Commit base: `2e45e292` (`development`, 2026-10-26). Worktree `.claude/worktrees/gh-204`, rama `feat/gh-204-whole-reads-reference-split`. Esta copia del registro es la vigente; la del checkout principal se borra al cerrar. Los `diff` de líneas de abajo comparan contra ese commit, no contra `HEAD`.

## Tareas

### T1 — Prueba de tamaño de referencias

- [x] `TestSkillReferenceSize` falla con la ruta y los bytes de cada referencia de más de 16 000 bytes.

**Closes:** AC3

**Checkpoint:** C1

**Depends on:** none

**Locations:** `tests/content/skill_size_test.go`

**Execution:** hilo principal, porque es pequeña y las tareas de separación dependen de ella.

**Test approach:** tdd. RED esperado en la base: la prueba nombra los cuatro archivos grandes con sus bytes. El GREEN llega con T3–T6.

**Changes:** constante `skillReferenceLimit = 16000` con comentario propio (recorte de lecturas, #204). `filepath.WalkDir` sobre cada `content/skills/*/references`, fallo si no se encuentra ningún archivo, y `t.Errorf` con el mismo formato de `TestSkillEntrypointSize`. Ver design.md, "Límite de tamaño de referencias".

**Verification:** `go test ./tests/content/ -run TestSkillReferenceSize -count=1` falla y nombra exactamente `delivery-decisions.md` (21084), `plan-format.md` (20514), `diagram-grammar.md` (17663) y `verification.md` (16692).
- Ese RED es la evidencia esperada de C1. `go test ./tests/content/` sigue en rojo hasta que terminen T3–T6.
- Las demás pruebas del paquete pasan: `go test ./tests/content/ -count=1 -skip TestSkillReferenceSize`.

### T2 — Regla de lectura completa y punteros

- [x] `global.md` tiene la regla de lectura completa, la regla de scripts la nombra, y `delegation.md` la nombra al pasar las fuentes a un hijo.

**Closes:** AC1, AC2

**Checkpoint:** C2

**Depends on:** none

**Locations:**
- `content/guidance/global.md:81,86`.
- `content/skills/flow-research/references/delegation.md:39`.
- `_support/docs/architecture/instruction-resources.md:13`.
- `_support/docs/architecture/repository-and-distribution.md:61-64` y `_support/docs/architecture/deployment-manager.md` (sección del hook de Grok).
- `AGENTS.md` (Guidance structure).

**Execution:** hilo principal, porque el texto lo fija este plan y conviene una sola mano en `global.md`.

**Test approach:** check, con la búsqueda y las pruebas de contenido.

**Changes:**
- El texto y los punteros de design.md, "Regla de lectura completa".
- Los punteros nombran la regla, sin repetir sus pasos.
- El bullet de AGENTS.md menciona la prueba de referencias.

**Verification:**
- `rg -n "read it whole|cut its lines|output limit below" content/guidance/global.md` muestra la regla en el bullet de recursos.
- `rg -n "read whole, as the resource rule" content/guidance/global.md` muestra el puntero en el bullet de scripts.
- `rg -n "role files among these sources whole" content/skills/flow-research/references/delegation.md` encuentra el puntero.
- `go test ./tests/content/ -count=1 -v -run 'GlobalGuidanceSize|SkillLinks'` pasa y registra el nuevo tamaño de `global.md` para el PR.

### T3 — Separar `delivery-decisions.md`

- [x] `delivery-decisions.md`, `review-decisions.md` y `ui-verification-decision.md` quedan en 16 000 bytes o menos, cada uno con su condición de lectura. Los enlaces y las frases que citan cada sección apuntan al archivo que ahora la contiene.

**Closes:** AC4

**Checkpoint:** C2

**Depends on:** T1

**Locations:**
- Archivos: `content/skills/flow-plan/references/{delivery-decisions,review-decisions,ui-verification-decision}.md`.
- Enlaces:
  - `flow-plan/SKILL.md:22`.
  - `plan-format.md:136`.
  - `flow-verify/SKILL.md:30,43,47,55,63`.
  - `flow-verify/references/review-and-gates.md:9,10,16`.
  - `flow-verify/references/ui-gate.md:19,38`.
- Frases que atribuyen la pregunta de revisión o el valor de UI:
  - `flow-build/SKILL.md:24`.
  - `harness-audit/SKILL.md:65`.
  - `flow-verify/SKILL.md:22`.
  - `unattended-delegation/SKILL.md:16`.
  - La L5 del núcleo.

**Execution:** hilo principal. Va primera entre las tareas que comparten archivos de enlaces (`review-and-gates.md`, `flow-plan/SKILL.md`) con T4 y T5, y escritores solapados no están permitidos.

**Test approach:** check, con la prueba de tamaño, la búsqueda de anclas y el diff de líneas.

**Changes:** la partición de design.md, "Separación por responsabilidad", entre párrafos completos, y las reglas que cruzan archivos como ahí se indica. El encabezado `## Dedicated review timing` conserva su texto.

**Verification:**
- `go test ./tests/content/ -run TestSkillReferenceSize -count=1` ya no nombra `delivery-decisions.md`.
- `rg -n "delivery-decisions\.md#dedicated-review-timing" content site _support/docs AGENTS.md` no devuelve nada.
- Cada `review-decisions.md#dedicated-review-timing` resuelve a un encabezado existente.
- `wc -c content/skills/flow-build/SKILL.md` da 16 000 bytes o menos.
- `diff <(git show 2e45e292:content/skills/flow-plan/references/delivery-decisions.md | sort) <(cat <las tres partes> | sort)` muestra solo las líneas de condición y de enlace añadidas.

### T4 — Separar `plan-format.md`

- [x] `plan-format.md`, `plan-criteria-and-markers.md` y `plan-handoff.md` quedan en 16 000 bytes o menos. `plan-format.md` conserva `# Retained plan format` y `## What the document must carry`, y los enlaces apuntan al archivo nuevo.

**Closes:** AC4

**Checkpoint:** C2

**Depends on:** T1

**Locations:**
- Archivos: `content/skills/flow-plan/references/{plan-format,plan-criteria-and-markers,plan-handoff}.md`.
- Enlaces:
  - `flow-plan/SKILL.md:48,74`.
  - `flow-build/references/plan-execution.md:17,19`.
  - `flow-plan/references/plan-review.md:22`.
  - `flow-plan/references/plan-revision.md:11`.

**Execution:** hilo principal, después de T3. Escribe en `flow-plan/SKILL.md`, igual que T3, así que el orden se fija aquí y no como dependencia.

**Test approach:** check.

**Changes:** la partición de design.md. La plantilla enlaza a `plan-criteria-and-markers.md` para `Closes:`, `Checkpoint:` y "False at the base when". La frase "lives only here" se conserva.

**Verification:**
- La prueba de tamaño ya no nombra `plan-format.md`.
- `rg -n "plan-format\.md#(acceptance-criteria-and-task-markers|hand-off-a-ready-plan)" content site` no devuelve nada.
- `rg -nx '# Retained plan format|## What the document must carry' content/skills/flow-plan/references/plan-format.md` encuentra las dos líneas que busca `tests/pilot/flows.go:357`.
- El diff de líneas contra `2e45e292`, como en T3.

### T5 — Sacar las autorizaciones previstas de `verification.md`

- [x] `verification.md` y `foreseen-authorizations.md` quedan en 16 000 bytes o menos, y los enlaces a la regla apuntan al archivo nuevo.

**Closes:** AC4

**Checkpoint:** C2

**Depends on:** T1

**Locations:**
- Archivos: `content/skills/flow-build/references/{verification,foreseen-authorizations}.md`.
- Enlaces `#gate-selection-and-timing` a la regla movida:
  - `delivery-decisions.md:5`.
  - `effect-authorization.md:5`.
  - `flow-plan/SKILL.md:40`.
  - `review-and-gates.md:22`.
  - `unattended-delegation/SKILL.md:16`.
  - `git-workflow/SKILL.md:20`, como puntero añadido.

**Execution:** hilo principal, después de T4. Escribe en `review-and-gates.md` y `flow-plan/SKILL.md`; el orden se fija aquí.

**Test approach:** check.

**Changes:** mover el tramo de L27 desde "Ask for the authorizations the plan foresees…". Los primeros 616 bytes de L27 se quedan en `verification.md` con un enlace al archivo nuevo. "this paragraph" se reescribe como enlace, como indica design.md.

**Verification:**
- La prueba de tamaño ya no nombra `verification.md`.
- Cada enlace a la regla movida resuelve en `foreseen-authorizations.md`.
- El diff de líneas contra `2e45e292` muestra solo L27 partida y la frase "this paragraph" reescrita, como excepciones declaradas.

### T6 — Separar `diagram-grammar.md` y remapear sus §

- [x] `diagram-grammar.md` y `diagram-svg-construction.md` quedan en 16 000 bytes o menos, conservan la numeración § y la atribución MIT, y los activos y scripts de `flow-report` nombran el archivo y el § correctos.

**Closes:** AC4

**Checkpoint:** C2

**Depends on:** T1

**Locations:**
- `content/skills/flow-report/references/{diagram-grammar,diagram-svg-construction}.md` y `flow-report/SKILL.md:27`.
- `assets/{baseline,skeleton-comparison,skeleton-deck,skeleton-explainer,skeleton-paper,skeleton-review}.html`.
- `scripts/verify_geometry.py:5-6,22` y `scripts/self_check.py:31`.
- `_support/docs/architecture/instruction-resources.md:108`.

**Execution:** delegada a `hive-write-docs`, en paralelo con T3–T5. Sus rutas no se cruzan con las de esas tareas, la partición ya está fijada en design.md, y el brief es más corto que el cambio.

**Test approach:** check.

**Changes:** la partición de design.md. Puntero nuevo en `flow-report/SKILL.md` con la condición "before writing or changing SVG markup". Se corrigen los desfases § preexistentes, como decisión menor registrada.

**Verification:**
- La prueba de tamaño ya no nombra `diagram-grammar.md`.
- `rg -n "diagram-grammar\.md|diagram-svg-construction\.md|§[0-9]" content/skills/flow-report` muestra que cada § citado existe en el archivo nombrado.
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- El diff de líneas contra `2e45e292`.

### T7 — HA-SK-02, huella de lectura, sitio y CHANGELOG

- [x] HA-SK-02 limita también las referencias, la huella de lectura mide las partes nuevas, el sitio enlaza los archivos nuevos y el CHANGELOG lo registra.

**Closes:** AC5

**Checkpoint:** C2

**Depends on:** T3, T4, T5, T6

**Locations:**
- `content/skills/harness-audit/references/skills-and-agents.md:8` y `_support/docs/harness-engineering/harness-audit-rules.md:41`.
- `tests/content/reading_footprint_test.go:60-106`.
- `site/src/content/docs/flows/{build,verify,plan}.md`.
- `_support/docs/architecture/instruction-resources.md:7`.
- `CHANGELOG.md`.

**Execution:** hilo principal, porque integra los nombres finales de T3–T6.

**Test approach:** check.

**Changes:**
- HA-SK-02 y su espejo, como indica design.md.
- En la huella de lectura, las cuatro partes nuevas entran en `cycle` y en `moved`.
- En el sitio cambian 5 enlaces.
- El CHANGELOG lleva una entrada en `[Unreleased]`.

**Verification:**
- `go test ./tests/content/ -run ReadingFootprint -count=1 -v` pasa y lista las cuatro partes nuevas en el escenario del ciclo.
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- `rg -n -i "each reference stays at or under 16,000 bytes" content/skills/harness-audit/references/skills-and-agents.md` encuentra la frase; en `2e45e292` no aparece. El `-i` se agregó en el build porque la frase empieza oración.
- En `site/`: `pnpm check`, `pnpm lint` y `pnpm build` terminan sin errores.

## Verificación compartida y revisión

Checkpoints:
- **C1** — T1. Las cuatro tareas de separación (T3–T6) dependen de ella.
- **C2** — T2–T7. Ninguna cambia un contrato público o persistido, y de cada tarea de separación solo depende T7. El orden de T3, T4 y T5 lo fija su campo Execution, porque solo ordena escritores en el hilo principal.

| Puerta | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Suite completa | candidato final, local | `go vet ./...` y `go test ./...` (unos 2 min) más `python3 -m unittest discover -s tests/skills -p '*_test.py'`; todo pasa | parte de la implementación |
| Anclas | candidato final | `rg -n -o '(delivery-decisions\|review-decisions\|ui-verification-decision\|plan-format\|plan-criteria-and-markers\|plan-handoff\|verification\|foreseen-authorizations\|diagram-grammar\|diagram-svg-construction)\.md#[a-z0-9-]+' content site _support/docs AGENTS.md`; cada ancla tiene su encabezado (ninguna prueba las valida) | parte de la implementación |
| Release y enlaces relativos | candidato final, commiteado en la rama de trabajo | `hive plan install --scope user --home <tmp bajo /private/...> --state-dir <otro tmp> --hosts codex,claude,grok,pi,opencode,cursor --source <checkout de la rama>`: el plan lista los archivos nuevos en cada host, sin error `instruction reference`. Es la única validación automática de enlaces entre archivos reales | parte de la implementación (local, home sintético) |
| Revisión dedicada | candidato local verificado, antes de pedirte la validación | `/code-review` de Claude Code; se corrigen o refutan sus hallazgos | D5-A |
| Validación humana | después de la revisión | revisas el diff y el PR propuesto; luego push, PR y merge | D4-B |

Lo que no corre:
- Una sesión real con la guía nueva, porque los pilotos de comportamiento están pausados por AGENTS.md (Measurement). La eficacia de la regla se observa en las próximas sesiones de Codex y Pi.
- Verificación de UI: el sitio solo cambia el destino de enlaces, sin cambiar layout ni flujo.

## Revisión del plan y avance

**Revisión f6f9fd12d8f8 (2026-10-26).** Dos `hive-review-plan` en paralelo:
- **Consistencia de la guía:** 4 bloqueantes y 3 no bloqueantes. Duró unos 4,6 min y usó unos 142k tokens.
- **Pruebas, distribución y estructura:** 1 bloqueante y 6 no bloqueantes. Duró unos 3,6 min y usó unos 127k tokens.

**Correcciones aplicadas sin preguntar.** Todas son la propuesta del propio revisor, así que no hace falta otra ronda:
- Checkpoints: opción B. Se quitan las dependencias que solo ordenaban escritores, y el orden pasa al campo Execution.
- Frases hermanas de `delivery-decisions.md`: opción (b), con enlaces nuevos en T3. Se descartó (a) porque deja 6 frases falsas en vez de 5.
- Puntero en `global.md:81`.
- Texto de la regla: "When you read…", y la cláusula sobre las partes aplicables.
- Verificación de AC5 por la frase nueva.
- Alcance del puntero de `delegation.md`.
- Frase "this paragraph" de T5.
- Mención del hook de Grok y de sus documentos.
- Puerta de release con `hive plan install`, que reemplaza a la de retiro: no se retira ningún archivo.
- Filtro `GlobalGuidanceSize`.
- Comprobación de los encabezados con `rg -nx`.
- Huella de lectura: sin `baselineOnly`.
- Diff de líneas contra `2e45e292`.
- RED esperado de C1.

**Límites que quedan:**
- El contenido final de los archivos nuevos se revisa en el build, con el diff de líneas y `/code-review`.
- El remapeo § de `flow-report` lo revisa T6, no la revisión del plan.

**Build (2026-10-26).** Commits sobre `2e45e292`, uno o más por tarea:
- T1: `3fc00888`.
- T2: `ff7ddae6`.
- T3: `d1cc2a97` y `1e528f95`.
- T4: `53a67cfb`.
- T5: `4d32e20d`.
- T6: `f0fb1e10`.
- T7: `23e9378f` y `973c3949`.

Las 7 tareas quedaron `[x]` con `hive-verify-task`:
- **C1:** T1, con el RED esperado.
- **C2:** T2–T7. T3 y T7 tuvieron una corrección cada una después de su primer veredicto, con su reverificación.

Implementación: el hilo principal, salvo T6, que hizo `hive-write-docs`. Modelos: el implementador es `claude-opus-5-5` (observado); los verificadores y `hive-write-docs` usan el perfil configurado, no observado.

**Cambios respecto del plan:**
- **T5.** La frase «When a gate this paragraph confirms…» se queda en `verification.md`, porque se refiere a las puertas de ese párrafo. Así no hubo que reescribirla. `git-workflow/SKILL.md:20` y `review-and-gates.md:22` no cambian, porque citan las puertas que se quedan.
- **T4.** La plantilla no lleva enlaces dentro del bloque de código: hay una frase justo antes que remite a `plan-criteria-and-markers.md`.
- **T3, correcciones tras su verificación:** `AGENTS.md:57` y `unattended-delegation/SKILL.md:16` nombran `review-decisions.md`.
- **T7, correcciones tras su verificación:**
  - `plan.md:69` y `build.md:48` del sitio enlazan `effect-authorization.md` para la clasificación de efectos.
  - Dos etiquetas de enlace se renombraron.
  - El comando de AC5 usa `-i`.
- **Detalles que quedan:**
  - `plan-review.md:22` y `plan-revision.md:11` dicen «the plan format's checkpoint rules», con un enlace a la casa correcta en la misma línea.
  - `build.md:48` no enlaza la decisión de UI.
  - HA-SK-02 dice «under», mientras la prueba admite «at or under». La diferencia ya existía para `SKILL.md`.

**Puertas del candidato:**
- `go vet ./...` y `go test ./...` pasan (20 paquetes).
- Las pruebas de Python pasan (64).
- Las 24 anclas resuelven.
- `hive plan install` sobre un home sintético planea los seis hosts sin error de enlaces.
- En el sitio, `pnpm check`, `lint` y `build` terminan sin errores.
- `global.md` pasa de 39 432 a 40 024 bytes (+592).

**Revisión dedicada (D5-A).** `/code-review` en nivel high sobre la rama, frente a `2e45e292`, reportó 9 hallazgos. Todos se corrigieron salvo uno:
- **P2, corregidos:**
  - En el camino sin entrega Git de `flow-build`, la decisión de UI y las autorizaciones previstas ya no se alcanzaban.
  - La regla de que una autorización sin respuesta queda sin autorizar se había movido lejos de las preguntas de puerta.
  - Las condiciones de lectura de `review-decisions.md` y `ui-verification-decision.md` no cubrían el trabajo sin entrega Git.
  - `flow-verify` no nombraba la decisión de UI.
  - El texto sobre el hook de Grok era demasiado amplio.
- **P3, corregidos:**
  - La frase sobre los chequeos costosos en `build.md:48`.
  - Nombres viejos en `plan-review.md` y `plan-revision.md`.
  - HA-SK-02 decía «under» donde la prueba admite «at or under».
- **P3, solo reportado:** los enlaces `blob/master` del sitio darán 404 hasta la promoción. Es la convención que ya usa el sitio.

Commits de corrección: `2ae94771` y `254b0469` (T3), `7417bdac` (T5), `7d9cb9eb` (T4), `4eb89968` (T2) y `48aadc65` (T7). `hive-verify-task` volvió a verificar todas las tareas tocadas: todos los criterios se cumplen. No hubo segunda revisión, porque ninguna corrección cambió lógica ni contrato.

**Cierre (2026-10-26).**
- **Validación:** la diste el 2026-10-26 (D8-A).
- **Integración:** [PR #213](https://github.com/JhonHawk/tricell-hive/pull/213), merge `c4b7feaf` en `development`. El repo no tiene CI; la suite local pasó sobre el candidato.
- **Tracker:** #204 cerrado por el merge.
- **Registro de cambio:** se archiva por PR de cierre (D7-A).
