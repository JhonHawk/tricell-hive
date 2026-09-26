# Tareas

### T1 — Criterio de TDD, oráculo fijo y alcance de la suite en `flow-build`

- [x] `SKILL.md:40`, `:42` y `:54` tienen la redacción de [design.md](design.md#redacción-distribuida-inglés).

**Depende de:** nada.
**Ubicación:** `content/skills/flow-build/SKILL.md:40`, `:42`, `:54`.
**Ejecución:** hilo principal; la redacción está acoplada a D1-A y es más corta que un brief.
**Enfoque de test:** `check`. Es guía distribuida sin runtime, así que se verifica con lectura y consistencia.
**Verificación:**
- `rg -n "by default for new or changed automatically testable|full suite locally" content/` no devuelve nada.
- `rg -n "TDD applies even if|needs no failing test first|supplied as acceptance|cannot produce|test the change makes obsolete|Do not rerun a passing suite|much larger than the change" content/skills/flow-build/SKILL.md` encuentra las siete frases (A1–A4).
- Releer `:40-54` junto con la fila "Build iteration" de `verification.md` y con `verification.md:20` (reemplazo de tests) para confirmar que no se contradicen.

### T2 — Enfoque de test por tarea en `flow-plan`

- [x] `flow-plan/SKILL.md:34` y la plantilla de `plan-format.md` incluyen el enfoque de test.

**Depende de:** T1, porque remite a su criterio.
**Ubicación:** `content/skills/flow-plan/SKILL.md:34`, `content/skills/flow-plan/references/plan-format.md:76`.
**Ejecución:** hilo principal; son dos frases.
**Enfoque de test:** `check`.
**Verificación:** `rg -n "test approach" -i content/skills/flow-plan` devuelve las dos ubicaciones; los tres valores coinciden en ambos archivos y con `flow-build/SKILL.md:40` ("A plan's declared test approach").

### T3 — `test-engineer` nombra el cambio que haría fallar cada test

- [x] El punto 2 de `test-engineer.md` tiene la redacción del diseño.

**Depende de:** nada.
**Ubicación:** `content/agents/quality/test-engineer.md:12`.
**Ejecución:** hilo principal; es una frase.
**Enfoque de test:** `check`.
**Verificación:** `rg -n "realistic change" content/agents/quality/test-engineer.md` devuelve una línea; el frontmatter no cambia (`git diff` solo muestra el punto 2); `go test -count=1 ./integrations/agents/...` pasa (`agents_test.go:24` renderiza cada `content/agents/*/*.md` real para todos los hosts).

### T4 — Quitar el ajuste `TDD` de `global.md`

- [x] `global.md:93` no lista `TDD` y el presupuesto coincide con el tamaño medido.

**Depende de:** nada.
**Ubicación:** `content/guidance/global.md:93`, `tests/content/budget_test.go:12`.
**Ejecución:** hilo principal.
**Enfoque de test:** `check` con el test de presupuesto existente.
**Verificación:** `rg -n '`TDD`' content/guidance/global.md` no devuelve nada; `wc -c content/guidance/global.md` coincide con `globalGuidanceBudget`; `go test -count=1 ./tests/content/...` pasa.

### T5 — Verificación conjunta, revisión y parada humana

- [ ] El diff pasa las verificaciones, la revisión está resuelta y el usuario leyó la redacción.

**Depende de:** T1–T4.
**Ejecución:** hilo principal; `/code-review` corre en su propio subagente.
**Verificación:**
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- `go vet ./...` y `go test -race -count=1 ./...` corren una vez sobre el candidato final. Si falla algo en rutas del trabajo ajeno (`tooling/**`), se guarda la salida como evidencia, se confirma que ninguna ruta de este cambio está involucrada y se reporta aparte.
- `git status` muestra intactos los cambios ajenos.
- Los hallazgos de `/code-review` quedan corregidos o refutados con evidencia.

## Verificación y revisión humana

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Lecturas y consistencia | T1–T4 | Comandos `rg` y relectura cruzada de cada tarea | Incluido en la implementación |
| Tests del repositorio | T5 | Tests de skills, `go vet` y `go test -race` una sola vez | Incluido en la implementación |
| Revisión de código | T5, antes del commit | `/code-review` de Claude Code sobre el diff de estas rutas | Elegido por el usuario el 2026-09-26 |
| Redacción | Antes del commit | El usuario lee el diff de `flow-build/SKILL.md`, `flow-plan`, `test-engineer.md` y `global.md:93` | Parada humana |
| Despliegue a hosts | Después del push | Manager de Hive, release nuevo | Fuera de este plan; se ofrece aparte |

No aplica verificación in vivo ni de UI: es guía distribuida y el efecto en sesiones reales requiere pilotos, que están pausados.

## Estado de revisión y progreso

**Revisión del plan (2026-09-26):**
- **Cobertura:** un solo dominio (guía distribuida y sus checks), un `review-plan` sobre la revisión `1ecf8a0d8565`.
- **Aceptados como los propuso el revisor:**
  - H1: rama residual para cambios de comportamiento fuera de las dos listas; `check` redefinido como "outside the TDD criterion".
  - H2: desempate anclado al comportamiento; "copy" pasa a "user-facing text".
  - H3: T3 verifica con `./integrations/agents/...`, no con `tooling/management`, que usa agentes sintéticos y tiene WIP ajeno.
  - H4: la suite se repite solo si una edición posterior puede afectar lo que cubre.
  - H5: la plantilla se alinea con la definición de `check`.
  - H6: `rg` concretos para A3 y A4.
- **Sin segunda ronda:** las correcciones aplican lo que propuso el propio revisor.
- **Límites:** el revisor no leyó el Hive legacy ni validó las citas externas; esas vienen de la investigación de esta sesión.
**Implementación (2026-09-26):**
- **T1:** hecho. `rg` de las frases viejas sin resultados; las siete frases nuevas aparecen una vez cada una en `flow-build/SKILL.md`.
- **T2:** hecho. `Test approach` aparece en `flow-plan/SKILL.md:34` y `plan-format.md:78`, con la misma definición de `check`.
- **T3:** hecho. `git diff --stat` muestra una sola línea cambiada en `test-engineer.md`; `go test -count=1 ./integrations/agents/...` pasa.
- **T4:** hecho. `global.md` pasa de 39056 a 39049 bytes y `globalGuidanceBudget` queda en 39049; `go test -count=1 ./tests/content/...` pasa.
- **T5:** en curso.
  - Pasan los tests de skills en Python (23 OK) y `go vet ./...`.
  - `go test -race -count=1 ./...`, corrido una vez: 15 paquetes OK y `tooling/cli` falla en `TestNativeAdapterWithoutFixtureOffersManualCapabilityAndRunsNoProcess` y `TestInstallPartialOnboardingListsPerStepDetail`.
  - Las dos fallas están en `provider_adapter_test.go`, un archivo sin versionar del trabajo `versioned-installer-onboarding` que otra sesión está editando. Revisan la redacción de la vista previa del instalador, que no sale de las rutas de este cambio.
  - `/code-review` (medium, sobre las 6 rutas): sin defectos altos ni medios. Hallazgos bajos, ambos corregidos:
    - H1: el enfoque de test declarado en el plan ya no anula el criterio de TDD si la tarea resulta tocar comportamiento de la lista; se registra la reclasificación.
    - H2: un test de aceptación dado por el usuario o el plan tampoco se elimina; se reporta y se propone la corrección.
