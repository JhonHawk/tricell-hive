# Tareas

### T1 — Criterio para recomendar revisión dedicada

- [x] `delivery-decisions.md` tiene la viñeta "No dedicated review" y el criterio de recomendación de [design.md](design.md#redacción-distribuida-inglés).

**Depende de:** nada.
**Ubicación:** `content/skills/flow-plan/references/delivery-decisions.md:23-25`.
**Ejecución:** hilo principal; la redacción está acoplada a D1-B y D2-A.
**Enfoque de test:** `check`, con lectura y consistencia.
**Verificación:**
- `rg -n "No dedicated review" content/skills/flow-plan/references/delivery-decisions.md` devuelve la viñeta y la frase de recomendación.
- `rg -n "Recommend a dedicated review when" content/` devuelve una sola línea.
- Releer `:18-27` para confirmar que "for every mode including hold" sigue siendo cierto: se pregunta siempre, y ahora "No dedicated review" es una respuesta posible.

### T2 — Tratamiento de hallazgos, rondas y un reviewer por candidato

- [x] `verification.md` tiene el párrafo nuevo después de `:39` y las filas `:31` y `:32` ajustadas.

**Depende de:** nada.
**Ubicación:** `content/skills/flow-build/references/verification.md:31-32`, `:39-40`.
**Ejecución:** hilo principal.
**Enfoque de test:** `check`.
**Verificación:**
- `rg -n "Fix P0 and P1 findings|Fix a P2 finding|Report P3 findings without fixing them|Re-review at most once|An approving review closes review|within the agreed review authorization|delivery decision already settled it|if one was selected" content/skills/flow-build/references/verification.md` encuentra las ocho frases.
- El enlace `security-boundaries.md` resuelve desde `content/skills/flow-build/references/`.
- Releer contra `:37`, `:57` (UI, Blocker/High), `:86`, `global.md:41` (hallazgos incidentales) y `flow-build/SKILL.md:32` para confirmar que no se contradicen.

### T3 — Filtro y conteo en `review-code`

- [x] `review-code.md` tiene la frase del punto 3 y el cierre con conteo.

**Depende de:** nada.
**Ubicación:** `content/agents/review/review-code.md:14`, `:20`.
**Ejecución:** hilo principal.
**Enfoque de test:** `check`.
**Verificación:** `rg -n "guards against, or tests|number of findings at each priority" content/agents/review/review-code.md` encuentra dos líneas; `go test -count=1 ./integrations/agents/...` pasa.

### T4 — `Review` declarado subordinado al criterio

- [x] `global.md:105` tiene la redacción nueva y el presupuesto coincide.

**Depende de:** T1.
**Ubicación:** `content/guidance/global.md:105`, `tests/content/budget_test.go:12`.
**Ejecución:** hilo principal.
**Enfoque de test:** `check` con el test de presupuesto.
**Verificación:** `rg -n "recommended mechanism when a dedicated review applies" content/guidance/global.md` devuelve una línea; `wc -c content/guidance/global.md` coincide con `globalGuidanceBudget`; `go test -count=1 ./tests/content/...` pasa.

### T5 — Verificación conjunta, revisión y parada humana

- [ ] El diff pasa las verificaciones, la revisión está resuelta y el usuario leyó la redacción.

**Depende de:** T1–T4.
**Ejecución:** hilo principal; `/code-review` corre en su propio subagente.
**Verificación:**
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa.
- `go vet ./...` y `go test -race -count=1 ./...` corren una vez sobre el candidato final. Las fallas en `tooling/**` del trabajo ajeno se comparan con las de `86473b8` (dos tests de `provider_adapter_test.go`) y se reportan aparte.
- `git status` muestra intactos los cambios ajenos.
- Los hallazgos de `/code-review` se tratan con las reglas nuevas: P0 y P1 se corrigen, P2 si es local y P3 solo se reporta, con una sola re-revisión.

## Verificación y revisión humana

| Gate | Momento | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Lecturas y consistencia | T1–T4 | `rg` y relectura cruzada | Incluido en la implementación |
| Tests del repositorio | T5 | Tests de skills, `go vet` y `go test -race` una vez | Incluido en la implementación |
| Revisión de código | T5, antes del commit | `/code-review` de Claude Code sobre estas rutas | Elegido por el usuario el 2026-09-26 |
| Redacción | Antes del commit | El usuario lee el diff de `delivery-decisions.md`, `verification.md`, `review-code.md` y `global.md:105` | Parada humana |
| Despliegue a hosts | Después del push | Manager de Hive desde un worktree limpio | Fuera de este plan; se ofrece aparte |

No aplica verificación in vivo ni de UI: es guía distribuida y medir el efecto requiere pilotos, que están pausados.

## Estado de revisión y progreso

**Revisión del plan (2026-09-26):**
- **Cobertura:** un `review-plan` (guía distribuida) sobre `e01c3758283d`.
- **Aceptados:**
  - B1: las filas `verification.md:31-32` ya no vuelven a proponer ni exigir revisión cuando el usuario eligió "No dedicated review".
  - B2: el filtro de `review-code` no suprime validaciones en límites de confianza.
  - N1: los P2 que no son locales y los hallazgos preexistentes siguen la regla de hallazgos incidentales (`global.md:41`); P3 es una excepción explícita.
  - N2: se reconocen los mecanismos que verifican sus propios hallazgos.
  - N3: un hallazgo nuevo de la re-revisión va al usuario, y la aprobación cierra la revisión hasta que una edición cambie lógica o contrato.
  - N4: `review-security` requiere la autorización acordada.
  - N5: no se recomienda omitir la revisión cuando la guía del proyecto la exige.
- **Sin segunda ronda:** las correcciones aplican lo que propuso el revisor. N1 usa redacción propia con el mismo efecto: alinear con la regla global y declarar la excepción de P3.
- **Límites:** no se verificó si `/code-review` puede restringirse solo a las correcciones.
**Implementación (2026-09-26):**
- **T1:** hecho. "No dedicated review" aparece en la viñeta y en la recomendación; "Recommend a dedicated review when" aparece una sola vez en `content/`.
- **T2:** hecho. Las ocho frases aparecen una vez cada una en `verification.md` y el enlace a `security-boundaries.md` resuelve.
- **T3:** hecho. Las dos frases están en `review-code.md`; `go test -count=1 ./integrations/agents/...` pasa.
- **T4:** hecho. `global.md` pasa de 39049 a 39059 bytes y `globalGuidanceBudget` queda en 39059; `go test -count=1 ./tests/content/...` pasa.
- **T5:** en curso.
  - Pasan los tests de skills en Python y `go vet ./...`.
  - `go test -race -count=1 ./...` pasa completo (16 paquetes). Las dos fallas de `tooling/cli` que había en `86473b8` ya no aparecen.
  - `/code-review` (medium, sobre las 5 rutas): dos hallazgos en `verification.md:41`, tratados con las reglas nuevas como P2 (locales y dentro del alcance). Se corrigieron sin re-revisión porque ninguno es P0 ni P1:
    - El mecanismo nativo no etiqueta P0–P3; ahora sus severidades se mapean a P0–P3 por impacto observable.
    - P3 chocaba con la regla de hallazgos incidentales; ahora P3 no se corrige ni se registra en otro lado.
