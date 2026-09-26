# Tareas

### T1 — Definición de cambio mecánico, regla de alcance, cierre e issues en `global.md`

- [x] `global.md:33`, `:41`, `:78` y `:80` tienen la redacción de [design.md](design.md#redacción-distribuida-inglés), y la viñeta de alcance está después de `:78`.

**Depende de:** nada.
**Ubicación:** `content/guidance/global.md:33`, `:41`, `:78`, `:80`; `tests/content/budget_test.go:12`.
**Ejecución:** hilo principal; la redacción está acoplada a D1-A, D2 y D3-A.
**Enfoque de test:** `check` con el test de presupuesto.
**Verificación:**
- `rg -n "A mechanical change is a rename|Build only what the task requires|after a mechanical change with no Pending item|concrete failure scenario with evidence|Mention style or wording nits|unless project guidance requires one|Mechanical changes alone" content/guidance/global.md` encuentra las siete frases.
- `rg -n "Simple questions and mechanical edits do not require a formal workflow or a report" content/` no devuelve nada.
- `wc -c content/guidance/global.md` coincide con `globalGuidanceBudget`, y el aumento se reporta.
- `go test -count=1 ./tests/content/...` pasa.
- Releer `:23-33` (reporte y cierre) y `:41` junto con la regla nueva para confirmar que "Leave code the change does not touch" no contradice la corrección de defectos marginales.

### T2 — Calificadores en `flow-build`, `verification.md` y las demás menciones de la regla de UI

- [x] `flow-build/SKILL.md:30`, `:32`, `:40`, `verification.md:31` y `:59`, `flow-plan/SKILL.md:45`, `plan-format.md:92` y `review-ux.md:3` tienen la redacción del diseño.

**Depende de:** T1, porque `:40` remite a la definición global.
**Ubicación:** `content/skills/flow-build/SKILL.md:30`, `:32`, `:40`; `content/skills/flow-build/references/verification.md:31`, `:59`; `content/skills/flow-plan/SKILL.md:45`; `content/skills/flow-plan/references/plan-format.md:92`; `content/agents/review/review-ux.md:3`.
**Ejecución:** hilo principal.
**Enfoque de test:** `check`.
**Verificación:**
- `rg -n "unless it is mechanical|that \[verification\]\(references/verification.md\) requires|as the global guidance defines it" content/skills/flow-build/SKILL.md` encuentra tres líneas.
- `rg -n "changes only text or a style value|longest realistic text and in each theme|when the UI gate below requires them" content/skills/flow-build/references/verification.md` encuentra las tres frases.
- `rg -n "rendered UI effect" content/` muestra que solo `verification.md:59` define la exención y que las demás apariciones remiten a ella sin decir "any" ni "the required".
- `rg -n "a rename, a move, user-facing text" content/` devuelve solo `global.md`.

### T3 — Mandatos acotados en los agentes de desarrollo

- [x] `frontend-developer.md:12` y `backend-developer.md:12` tienen la redacción del diseño.

**Depende de:** nada.
**Ubicación:** `content/agents/development/frontend-developer.md:12`, `content/agents/development/backend-developer.md:12`.
**Ejecución:** hilo principal; son dos frases.
**Enfoque de test:** `check`.
**Verificación:** `rg -n "that the change introduces or alters|where the change affects them" content/agents/development` encuentra dos líneas; `go test -count=1 ./integrations/agents/...` pasa.

### T4 — Verificación conjunta, revisión y parada humana

- [x] El diff pasa las verificaciones, la revisión está resuelta y el usuario leyó la redacción.

**Depende de:** T1–T3.
**Ejecución:** hilo principal; `/code-review` corre en su propio subagente.
**Verificación:**
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- `go vet ./...` y `go test -race -count=1 ./...` corren una vez sobre el candidato final. Una falla en rutas del trabajo ajeno se reporta aparte, con su salida.
- `git status` muestra intactos los cambios ajenos.
- Los hallazgos de `/code-review` se tratan por prioridad: P0 y P1 se corrigen, P2 si es local y P3 solo se reporta, con una sola re-revisión.

## Verificación y revisión humana

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Lecturas y consistencia | T1–T3 | `rg` y relectura cruzada | Incluido en la implementación |
| Tests del repositorio | T4 | Tests de skills, `go vet` y `go test -race` una vez | Incluido en la implementación |
| Revisión de código | T4, antes del commit | `/code-review` de Claude Code sobre estas rutas | Elegido por el usuario el 2026-09-26 |
| Redacción | Antes del commit | El usuario lee el diff de `global.md`, `flow-build`, `verification.md` y los dos agentes | Parada humana |
| Despliegue a hosts | Después del push | Manager de Hive desde un worktree limpio | Fuera de este plan; se ofrece aparte |

No aplica verificación in vivo ni de UI: es guía distribuida y medir el efecto requiere pilotos, que están pausados.

## Estado de revisión y progreso

**Revisión del plan (2026-09-26):**
- **Cobertura:** un `review-plan` (guía distribuida) sobre `460dbb01fbd5`.
- **Aceptados como los propuso el revisor:**
  - B1: la exención de UI vive solo en `verification.md:59`; `flow-build:32`, `verification.md:31`, `flow-plan:45`, `plan-format.md:92` y `review-ux.md:3` remiten a ella.
  - B2: se mantiene que una pregunta simple no necesita reporte.
  - N1: la lista de exclusión suma dinero, concurrencia y dependencias, y la configuración que toca permisos, seguridad o dependencias no es mecánica.
  - N2: sin revisión dedicada solo si el proyecto no la exige.
  - N3: el cierre sin pregunta aplica cuando no hay ítem Pendiente, y los recordatorios siguen en el reporte.
  - N4: el implementador revisa el texto más largo y cada tema; si cambia el layout, aplican los hijos.
  - N5: un solo término en `:80`; "delegation of the edit itself"; "outside the change's scope"; límite de confianza abierto con "or other external data".
- **Sin segunda ronda:** las correcciones aplican lo que propuso el revisor.
- **Fixture congelado:** `tests/fixtures/workspace-conventions/baseline/global.md:23` conserva la redacción antigua a propósito (SHA-256 fijado); no se edita.
**Implementación (2026-09-26):**
- **T1:** hecho.
  - Las siete frases están en `global.md` y la frase vieja ya no aparece en `content/`.
  - `global.md` pasa de 39059 a 40635 bytes (+1576, unos 394 tokens), que se justifican por la regla de alcance y la definición de cambio mecánico. `globalGuidanceBudget` queda en 40635.
  - `go test ./tests/content/...` pasa.
- **T2:** hecho.
  - Las tres frases están en `flow-build/SKILL.md` y las tres en `verification.md`.
  - La lista de cambios mecánicos queda solo en `global.md`.
  - La exención de UI vive solo en `verification.md:59`; las otras cinco apariciones de "rendered UI effect" remiten a ella.
- **T3:** hecho. Una frase en cada agente; `go test ./integrations/agents/...` pasa.
- **T4:** hecho.
  - Pasan los tests de skills (OK) y `go vet ./...`.
  - `go test -race ./...`: 15 paquetes OK. Falló `TestBootstrapRealManagerReachesInteractiveInstallerThenLeavesNoChange`, en `tooling/distribution/bootstrap_shell_test.go`, un archivo sin versionar que la sesión del instalador editó a las 04:27. Se guardó la salida y, al volver a correr solo ese test, pasó. Es trabajo en curso ajeno y no toca las rutas de este cambio.
  - `/code-review` (medium, sobre las 9 rutas): tres hallazgos, traducidos a P0–P3 como P2 (H2) y P3 (H1, H3). Se corrigieron por ser locales, sin re-revisión porque ninguno es P0 ni P1:
    - H1: frase duplicada en `flow-plan/SKILL.md:45`.
    - H2: "a presentation edit" hacía mecánico un cambio de layout y lo sacaba de la regla de UI. Ahora la definición global dice "a text or style change that alters no layout, state, or flow", igual que `verification.md:59`.
    - H3: los problemas sospechados sin escenario de falla se mencionan en el reporte como no confirmados, en vez de perderse.
- **Entrega:** redacción aprobada por el usuario; commit `8a7e5b9` con push a `rebuild/harness-engineering`. Los cambios ajenos del árbol quedaron fuera del commit.
