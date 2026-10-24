# Tareas · skill-size-budget

Base del cambio: `b0f02c90` (confirmada al arrancar el build el 2026-10-23, rama `feat/skill-size-budget`).

## Diseño en vigor

- **Mover, no reescribir.** Cada pieza se traslada con su texto actual; solo se ajustan los enlaces relativos al nuevo directorio y la frase que la introduce. Donde el destino ya dice la regla, el origen deja un puntero y no se copia nada. Lo que no tiene otro hogar se queda.
- **Punteros con condición exacta**, en el formato de las referencias existentes: «When <situación>, read [<nombre>](references/<archivo>.md)».
- **Referencias nuevas**, cada una con un «When to read» al principio, como las demás:
  - `flow-build/references/plan-execution.md` — condición: «while you execute or resume a retained plan, a change folder or a `<topic>.plan.md`, before you start, mark, verify, or commit any of its tasks». Recibe B1 (línea 48), B2 (líneas 50–57), B3 (59), B4 (61, salvo una frase general que se queda: «When the delivery mode makes commits without a retained plan, commit each coherent result once its check passes»), B5 (63) y B6 (de la línea 34, desde «For a plan task, the brief names the change folder path» hasta «the files win.»; la frase siguiente, «Without a retained plan, the brief quotes the user's request…», se queda).
  - `flow-build/references/tdd.md` (D26) — condición en `flow-build`: «before you change code or tests»; en `flow-plan/SKILL.md:36` («the build skill's TDD criterion»): «when you choose a task's test approach». Recibe las líneas 44–46. También pasan a enlazarla `flow-walkthrough/SKILL.md:60` («the build skill's test-first rule») y `flow-build/references/verification.md:3` («The build skill owns the essential test-first and completion procedure»), cuya frase se ajusta para nombrar `tdd.md`.
  - `flow-build/references/tools-and-dependencies.md` — condición: «before you install a missing tool or add a dependency». Recibe B13 (líneas 26 y 38).
  - `flow-plan/references/interface-changes.md` — condición: «when a change alters an interface that other code, services, or data depend on, while planning or building». Recibe P1 (línea 22 de `flow-plan`) y B10 (línea 42 de `flow-build`).
  - `flow-plan/references/plan-revision.md` — condición: «when a retained plan changes after it was written». Recibe P2 («Revise a plan in progress», líneas 56–64). Los 7 enlaces a `#revise-a-plan-in-progress` (6 archivos, incluido el interno de `flow-plan/SKILL.md:18`) pasan a la referencia con su ruta relativa. Su frase «[flow-build] then verifies it again under its per-task procedure» pasa a apuntar a `plan-execution.md`.
- **Destinos existentes y lo que se queda** (ajustado tras la revisión del plan):
  - B7 (línea 36): su parte de UI (qué hijos corren según `UI verification` y cuándo se pregunta) ya está en `flow-walkthrough/references/ui-gate.md`; `flow-build` deja una frase de puntero. Se queda en `flow-build` la parte general: proponer otra verificación independiente cuando el riesgo lo pide, un solo dueño para mutaciones compartidas y «The implementer still verifies its own change».
  - B8 (línea 82): pasa a `flow-walkthrough/SKILL.md`, junto a «Correction rounds and markers».
  - B9 (línea 78): solo pasan a `flow-walkthrough/SKILL.md` las reejecuciones del gate del candidato aceptado («at its accepted-candidate gate, rerun the full suite only when the corrections could affect what it covers, and the security review only when they touch a boundary») y la corrida local que no avanza («A local run that makes no progress stops…»). Se queda en `flow-build` «These steps run before the first validation request», con el puntero a las rondas de corrección; la revisión de código dedicada ya está en `ui-gate.md:32` y el arreglo posterior en `flow-walkthrough/SKILL.md:64`, así que de esas dos frases queda solo el puntero. La frase «with the reruns [flow-build] sets» (`flow-walkthrough/SKILL.md:62`) pasa a nombrar la regla en su nuevo sitio para no quedar en círculo. `flow-walkthrough/SKILL.md` debe quedar en 16.000 bytes o menos.
  - B11 (línea 20): queda una versión corta con lo que `delivery-decisions.md` no dice: el modo elegido carga los gates posteriores, incluida la espera del CI, y cuando la entrega no es Git y no es mecánica se pregunta solo la revisión y `UI verification`. El resto pasa a ser un puntero a `delivery-decisions.md`.
  - B12 (línea 67): pasa a `flow-research/references/delegation.md`, «Check returned work», solo lo que toca a los hijos (registrar los procesos que dejan, comprobar que siguen vivos). Se queda en `flow-build` lo del hilo principal: el silencio no significa que un comando esté colgado, usar los controles de ejecución larga del host y no detener procesos ajenos. `flow-close/SKILL.md:25` («as [flow-build] sets») pasa a enlazar `delegation.md` para los procesos que dejan los hijos.
  - B14 (la parte de la línea 71 sobre recorridos en entornos desplegados): pasa a `flow-build/references/verification.md`.
  - B15: no sale. La condición de change-impact de la línea 16 no está en la línea 22, que solo enlaza seguridad, datos y diagnóstico.
  - `plan-format.md:40` («`flow-build` owns the transitions») pasa a enlazar `plan-execution.md`.
- **Compactación (AC1):** la frase de `global.md:78` sobre la skill activa pasa a: tras una compactación, relee completo el `SKILL.md` de la skill activa, porque un host puede volver a adjuntar solo su principio. No cambia la parte sobre las referencias.
- **Prueba (AC3):** `tests/content/skill_size_test.go` recorre `content/skills/*/SKILL.md` y falla con el nombre y el tamaño de cada uno que supere 16.000 bytes. El rojo se observa sobre la base (los archivos de `b0f02c90`, por ejemplo con `git show b0f02c90:<path>` a un directorio temporal de la tarea) y su salida se guarda en el progreso; el archivo de la prueba entra al árbol de trabajo después de T3, así T1–T3 verifican sin él.
- **`instruction-resources.md:62`** («there is no new size quota») pasa a decir que la cuota de tamaño de un `SKILL.md` la fija HA-SK-02 y la comprueba `tests/content/skill_size_test.go`.
- **Frases distintivas para AC9:**
  - Salen de `flow-build/SKILL.md` y aparecen en su destino: B1 «record the change's base commit»; B2 «What makes the verdict independent is the verifier's context»; B3 «Do not wait idle while a verification runs»; B4 «commit each task yourself when it reaches»; B5 «Other tasks, such as `characterization`»; B6 «For a plan task, the brief names the change folder path»; B10 «When a plan keeps a compatible transition»; B12 «check its liveness through its tool activity»; B13 «Before installing a missing tool» y «Before adding a dependency»; B14 «When a deployed walk runs»; B8 «leave the verified built local stack available»; B9 «rerun the full suite only when the corrections could affect»; TDD «Use red-green-refactor TDD» y «Preserve implementation that already exists before RED».
  - Salen de `flow-plan/SKILL.md` y aparecen en su destino: P1 «identify its actual producers and consumers»; P2 «Keep the files coherent in every direction».
  - Se quedan en `flow-build/SKILL.md`: «The implementer still verifies its own change» (B7), «Without a retained plan, the brief quotes the user's request» (B6), «Silence alone does not mean a command is hung» (B12), «These steps run before the first validation request» (B9), la condición de change-impact de la línea 16 (B15) y la versión corta de B11.

## Tareas

### T1 — Releer la skill completa tras la compactación

- [x] `global.md:78` pide releer completo el `SKILL.md` activo tras una compactación.

**Closes:** AC1.

**Depends on:** none.

**Locations:** `content/guidance/global.md:78`.

**Execution:** main thread; una frase.

**Test approach:** check — la búsqueda de AC1.

**Changes:** ver «Diseño en vigor», compactación.

**Verification:** `rg -c -i 'only (its|the) (start|beginning)' content/guidance/global.md` encuentra la línea; `git diff b0f02c90 -- content/guidance/global.md` muestra que la frase sobre las referencias no cambia; `go test ./tests/content/... -count=1` pasa.

### T2 — Achicar flow-plan

- [x] `flow-plan/SKILL.md` queda en 16.000 bytes o menos, con P1 y P2 en sus referencias y todos los enlaces actualizados.

**Closes:** AC6 (P1 y el enlace de `flow-plan`), AC7.

**Depends on:** none.

**Locations:** `content/skills/flow-plan/SKILL.md` (incluido el enlace interno de la línea 18), `content/skills/flow-plan/references/interface-changes.md` (nuevo), `content/skills/flow-plan/references/plan-revision.md` (nuevo), `content/skills/flow-plan/references/questions.md:68`, `plan-review.md:38`, `plan-format.md:22`, `content/skills/flow-walkthrough/SKILL.md:66`, `content/skills/flow-build/SKILL.md:16,48`.

**Execution:** main thread; mueve texto que otras tareas enlazan.

**Test approach:** check — tamaño, frases de P1 y P2, búsqueda de AC7, punteros con su condición y validación de enlaces.

**Changes:** ver «Diseño en vigor», `interface-changes.md` (solo P1 en esta tarea) y `plan-revision.md`.

**Verification:** `wc -c content/skills/flow-plan/SKILL.md` ≤ 16000; las frases de P1 y P2 aparecen solo en sus referencias; `rg -c '#revise-a-plan-in-progress' content/` no encuentra nada; `flow-plan/SKILL.md` enlaza `interface-changes.md` y `plan-revision.md` con sus condiciones; `go test ./tests/content/... ./integrations/mdlinks/... ./tooling/management/... -count=1` pasa.

### T3 — Achicar flow-build

- [x] `flow-build/SKILL.md` queda en 16.000 bytes o menos, con B1–B14 y el TDD en sus destinos y lo que no tiene otro hogar en su sitio.
  ronda 1 · AC9 · «The dedicated code review stays where the delivery mode sets it.» sigue en `flow-build/SKILL.md:62` y ya está en `ui-gate.md:32`: dos hogares.

**Closes:** AC5, AC6 (B10 y el enlace de `flow-build`), AC8, AC9, AC12.

**Depends on:** T2 (B10 va a `interface-changes.md`, que crea T2).

**Locations:** `content/skills/flow-build/SKILL.md`, `content/skills/flow-build/references/plan-execution.md` (nuevo), `content/skills/flow-build/references/tdd.md` (nuevo), `content/skills/flow-build/references/tools-and-dependencies.md` (nuevo), `content/skills/flow-build/references/verification.md`, `content/skills/flow-plan/references/interface-changes.md`, `content/skills/flow-plan/SKILL.md:36`, `content/skills/flow-plan/references/plan-format.md:40`, `content/skills/flow-plan/references/plan-revision.md` (su puntero al procedimiento por tarea), `content/skills/flow-research/references/delegation.md`, `content/skills/flow-close/SKILL.md:25`, `content/skills/flow-walkthrough/SKILL.md` (incluidas las líneas 60 y 62).

**Execution:** main thread; es la tarea más delicada y su brief sería más largo que el cambio.

**Test approach:** check — tamaños, las frases de AC9, los punteros con su condición y la validación de enlaces.

**Changes:** ver «Diseño en vigor».

**Verification:** `wc -c content/skills/flow-build/SKILL.md content/skills/flow-walkthrough/SKILL.md` ≤ 16000 cada uno; cada frase que sale aparece en su destino y no en `flow-build/SKILL.md`; cada frase que se queda sigue en `flow-build/SKILL.md`; `flow-build/SKILL.md` enlaza `plan-execution.md`, `tdd.md`, `tools-and-dependencies.md` e `interface-changes.md` con las condiciones del diseño, y `flow-plan/SKILL.md:36`, `flow-walkthrough/SKILL.md:60` y `flow-build/references/verification.md:3` enlazan `tdd.md` con sus condiciones; `go test ./tests/content/... ./integrations/mdlinks/... ./tooling/management/... -count=1` pasa.

### T4 — Prueba de tamaño y HA-SK-02

- [x] Una prueba de contenido falla si un `SKILL.md` supera 16.000 bytes, y `HA-SK-02` fija ese límite con su razón.

**Closes:** AC2, AC3, AC4.

**Depends on:** T2, T3 (la prueba entra al árbol después de ambas; su rojo se observa antes, sobre la base).

**Locations:** `tests/content/skill_size_test.go` (nuevo), `content/skills/harness-audit/references/skills-and-agents.md:8`, `_support/docs/architecture/instruction-resources.md:62`.

**Execution:** main thread.

**Test approach:** tdd — rojo sobre los archivos de la base, nombrando `flow-build` y `flow-plan`; verde sobre el árbol tras T2 y T3.

**Changes:** ver «Diseño en vigor», prueba e `instruction-resources.md`; en HA-SK-02, «stays under 500 lines and 16,000 bytes, because a host may keep only the start of an invoked skill after compaction».

**Verification:** la salida del rojo guardada en el progreso nombra los dos archivos; `go test ./tests/content/ -run SkillEntrypointSize -count=1` pasa; `wc -c content/skills/*/SKILL.md` muestra todos ≤ 16000; `rg -c '16,000' content/skills/harness-audit/references/skills-and-agents.md` encuentra la línea.

### T5 — Guía del repositorio, lecturas y changelog

- [x] `AGENTS.md`, la lista de lecturas de la prueba de huella y `CHANGELOG.md` reflejan el cambio.

**Closes:** AC10, AC11.

**Depends on:** T1–T4.

**Locations:** `AGENTS.md` («Guidance structure»), `tests/content/reading_footprint_test.go` (lista `cycle`: añadir `flow-build/references/plan-execution.md` y `flow-build/references/tdd.md`, que ahora lee un ciclo de plan y build), `CHANGELOG.md`.

**Execution:** main thread.

**Test approach:** check — búsquedas de AC10 y AC11 y la prueba de huella.

**Changes:** una frase en `AGENTS.md` que apunta a la prueba; las rutas nuevas en la lista, como mantenimiento sin criterio propio (la prueba de huella solo mide); una entrada en «Changed».

**Verification:** `rg -c '16,000' AGENTS.md` encuentra la línea; la búsqueda de AC11 encuentra la entrada; `go test ./tests/content/... -count=1` pasa.

## Verificación compartida

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Candidato final, checkout local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`, `pnpm run build` en `site/`; todo pasa | Implementación |
| Verificación por tarea | Tras cada tarea | `hive-verify-task` con los criterios de su `Closes:` | Implementación |

Sin efecto de UI renderizada. El PR cita también las páginas de `site/` revisadas sin cambio (`flows/build.md`, `flows/plan.md`), como pide `AGENTS.md`.

## Revisión del plan y progreso

- 2026-10-23: plan escrito.
- Revisión 1: `hive-review-plan`, dominio «estructura de la guía distribuida», nativo; unos 146k tokens y 209 s según el host. Cinco bloqueantes. Cuatro aplicados como los propuso: condición del puntero de B1–B6, punteros y afirmaciones que faltaban en Locations (`flow-close:25`, `flow-walkthrough:62`, `plan-format:40`, el enlace interno de `flow-plan:18`, `instruction-resources.md:62`), el orden de T4 (rojo sobre la base, prueba al árbol tras T3) y verificaciones que cubren AC5–AC9. El quinto (partes de B6, B7, B11, B12 y B15 sin otro hogar) dejaba `flow-build` en unos 16.600 bytes; el usuario decidió D26 (mover el TDD a `tdd.md`). No bloqueantes aplicados: tamaño base 28.911, condición de `interface-changes.md` que cubre ambos casos, `instruction-resources.md:62`, páginas del sitio citadas en el PR. Corrección del revisor descartada: HA-SK-02 sí está en `skills-and-agents.md:8` (`rg -n` lo muestra), no en la 10. La decisión D26 es una corrección nueva, así que va a la ronda de re-revisión.
- Re-revisión (ronda única, mismo revisor, unos 161k tokens acumulados y 57 s): correcciones aplicadas como las propuso: condiciones de `tdd.md` («before you change code or tests» en `flow-build`, «when you choose a task's test approach» en `flow-plan:36`), dos citas de la regla TDD añadidas (`flow-walkthrough/SKILL.md:60`, `verification.md:3`), B9 partido para no duplicar `ui-gate.md:32` ni `flow-walkthrough/SKILL.md:64`, y la lista de la prueba de huella anotada como mantenimiento. El revisor retiró su corrección sobre HA-SK-02 (está en la línea 8). Estimación del revisor: `flow-build` unos 14.250 bytes; `flow-walkthrough` unos 15.300 tras partir B9. Sin huecos que vayan al usuario.
- 2026-10-23, build: T4, rojo observado sobre la base antes de T1 con `go test ./tests/content/ -run SkillEntrypointSize -count=1`: «content/skills/flow-build/SKILL.md: 28911 bytes, over the 16000-byte skill entrypoint limit» y «content/skills/flow-plan/SKILL.md: 17326 bytes, over the 16000-byte skill entrypoint limit», FAIL. El archivo de la prueba queda fuera del árbol, en el scratch de la sesión, hasta después de T3. Nombre real del test: `TestSkillEntrypointSize` (la verificación de T4 se ajustó).
- 2026-10-23, build terminado: T1 `1acafbba`, T2 `6681ad6a`, T3 `5e8651c1` (una ronda de corrección: frase duplicada sobre la revisión de código dedicada), T4 `2d8a9c8a`, T5 `087484fe`; merges de `origin/development` (#124, #125) en `3a2d63ad` y antes del PR, sin conflictos. Implementador: hilo principal, Claude Opus 5.5 (observado: modelo de la sesión); verificadores: `hive-verify-task` nativo, modelo configurado, no observado. Suite final: `go vet ./...`, `go test ./...`, pruebas de skills y build del sitio pasan. Tamaños finales: flow-build 15.179, flow-walkthrough 15.428, flow-close 15.083, flow-plan 13.790 bytes. PR #126 fusionado como `f44468af`.
