# Tareas

Base del cambio: `83c9d3b5` (`origin/development`, 2026-10-30). Worktree `.claude/worktrees/gh-300-context-and-ui-gate`, rama `feat/gh-300-context-and-ui-gate`.

### T1 — Regla de contexto por fase

- [x] `hive-core/references/context.md` existe con la regla de S2, y los turnos de límite de fase apuntan a ella.

**Closes:** AC1, AC2.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/hive-core/references/context.md` (nuevo), `content/skills/hive-core/SKILL.md`, `content/skills/flow-plan/references/plan-handoff.md`, `content/skills/flow-verify/SKILL.md` («Validation by the user»), `content/skills/flow-close/SKILL.md` (opción de continuar con un flujo en curso, ~:55), `_support/docs/architecture/instruction-resources.md` (lista de `hive-core`), `tests/content/reading_footprint_test.go` (escenario `cycle` y mapa `moved`, ~:85-106, porque la base `d5164f71` no tiene el archivo), `site/src/content/docs/flows/plan.md`, `verify.md`, `close.md`.

**Execution:** hilo principal: redacción de una regla nueva y sus punteros, acoplada al texto que no debe contradecir.

**Test approach:** check: `go test ./tests/content/` y las búsquedas de Verification.

**Changes:** S1–S3 de `design.md`. En `site/`, una oración por página que describa la línea, sin copiar la regla. Revisar el comportamiento completo como pide `AGENTS.md`: `global.md:68,73,83`, `flow-close/SKILL.md:47,55`, `plan-handoff.md:9`, `reporting.md:98`, `work-item-boundary.md`, `loops.md:27` y `compaction-survival.md`.

**Verification:**
- `rg -n "references/context.md" content/skills` devuelve `hive-core/SKILL.md`, `plan-handoff.md`, `flow-verify/SKILL.md` y `flow-close/SKILL.md`.
- `context.md` declara cada propiedad de AC1 en una frase propia, y el verificador de C1 las comprueba leyendo el archivo:
  - una sola línea;
  - el estado en disco primero;
  - no es pregunta y no reemplaza la línea de pausa;
  - no parte el flujo ni la sesión;
  - no escribe memoria;
  - los tres puntos como lista cerrada, con la ruta de research excluida.
- `rg -n -i "not a question|state.*on disk|persistent memory|research" content/skills/hive-core/references/context.md` devuelve esas frases.
- `rg -n -i "is never split|another session" content/skills/flow-close/SKILL.md content/skills/flow-plan/references/plan-handoff.md` sigue devolviendo las dos reglas sin cambios de sentido.
- `wc -c content/skills/flow-close/SKILL.md content/skills/hive-core/references/context.md` da 16 000 bytes o menos cada uno.
- `go test ./tests/content/` pasa.

### T2 — UI-gate: declarar antes de preguntar

- [x] La decisión de verificación de UI exige la línea de no aplica, y el alcance de la gate tiene el ejemplo del texto de un aviso.

**Closes:** AC3.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-plan/references/delivery-decisions.md:40`, `content/skills/flow-verify/references/ui-gate.md:7` (Scope), `content/skills/flow-build/SKILL.md:27` (puntero incondicional para el trabajo sin entrega Git), `site/src/content/docs/flows/verify.md` (valores de verificación de UI).

**Execution:** hilo principal: una frase en cada uno de tres archivos y una oración en `site/`.

**Test approach:** check: `go test ./tests/content/` y las búsquedas de Verification.

**Changes:** S4 de `design.md`. Revisar los textos que repiten el disparador: `flow-plan/SKILL.md:54`, `flow-build/SKILL.md:44`, `delivery-decisions.md:40`, `instruction-resources.md:31` y `site/.../build.md`. Solo se corrige una copia que quede en conflicto.

**Verification:**
- `rg -n -i "notification|message text" content/skills/flow-verify/references/ui-gate.md` devuelve el ejemplo en Scope.
- `rg -n -i "does not apply" content/skills/flow-plan/references/delivery-decisions.md` devuelve la línea de la ronda.
- El verificador de C1 confirma que la línea se lee sin pasar por la condición «rendered UI effect».
- `wc -c content/skills/flow-verify/references/ui-gate.md` da 16 000 bytes o menos.
- `go test ./tests/content/` pasa.

### T3 — Límite de roles en Cursor

- [x] `agent-delivery.md` registra lo que carga Cursor CLI 2026.10.22 y el fallback.

Ronda 1 (`[!]`): AC4 no cumplido. La frase de `README.md:211` omitía `.claude/agents` y `.grok/agents` del workspace. Se corrigió en `50e55033` y se re-verifica AC4.

**Closes:** AC4.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `_support/docs/architecture/agent-delivery.md` (~:114, ~:118, ~:167, ~:197), `README.md:211`.

**Execution:** hilo principal: documentación interna breve.

**Test approach:** check: `rg -n "2026.10.22" _support/docs/architecture/agent-delivery.md` y `go test ./tests/content/`.

**Changes:** S5 de `design.md`, con etiquetas de evidencia (**obs (bin)** para el bundle, **doc** para la página de Cursor).

**Verification:**
- `rg -n "computeAgentsDirs|2026.10.22" _support/docs/architecture/agent-delivery.md` devuelve las filas actualizadas.
- `rg -n "not yet observed" _support/docs/architecture/agent-delivery.md` ya no devuelve la fila de Cursor.
- `rg -n "2026.10.22.*\.cursor/agents.*\.claude/agents.*\.grok/agents" README.md` devuelve la fila de Cursor (añadido en la ronda 1).

## Verification and human review

- C1 — T1, T2, T3 — Son cambios de texto independientes y pequeños; un solo verificador al final cubre los cuatro criterios.
- Pruebas de contenido: `go test ./tests/content/`. Suite completa una vez antes de integrar: `go test ./...`.
- Sin UI renderizada: la UI-gate no aplica. Sin in-vivo local: el cambio es solo de guía y no hay comportamiento ejecutable que correr.
- Pilotos con modelos: pausados (`AGENTS.md`, Measurement); la conducta queda por observar en sesiones reales.
- Requisitos de `AGENTS.md` para el PR y los commits:
  - El PR reporta que `content/guidance/global.md` no cambia de tamaño (19,169 bytes) y lista los lugares revisados (S1–S5 y las Locations).
  - El commit de T2 nombra las dos sesiones con host, ID y fecha: Grok `01a1855f-cdbe-7200-8434-7323a59ff52b` del 2026-10-28 y Grok `01a18abb-b9c6-7db3-a80a-31bbc5fad8fb` del 2026-10-29.

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Pruebas de contenido | Tras cada tarea, local | `go test ./tests/content/` pasa | Implementación |
| Verificador de C1 | Tras T3 | `hive-verify-task` con AC1–AC4 | Implementación |
| Suite completa | Una vez antes de integrar | `go test ./...` pasa | Implementación |
| Revisión dedicada | — | No corre (D7-C) | Decisión del usuario |

## Review status and progress

Revisión del plan (2026-10-30), dominio de consistencia de guía, con un revisor `hive-review-plan` que corrió sobre Claude Code:
- **Ronda 1** sobre `9fbdb796c3fe` (~124 k tokens, 3.3 min): 4 bloqueantes (B1–B4), una decisión (D1) y 5 menores (N1–N5).
  - Corregidos sin preguntar, por ser propuestas del revisor que no cambian decisiones del usuario: B1–B4 y N1–N4.
  - D1 pasó al usuario y se resolvió con D8-A.
  - Además corregí el conteo de recurrencia de T2: son 2 sesiones, no 4.
- **Re-revisión** sobre `f8d59140e02f` (~138 k tokens, 1 min): 1 bloqueante (R1) y 1 menor (R2). Ambos se aplicaron tal como los propuso el revisor y se conciliaron sin otra ronda.
- **Límites:** no se verificó cómo maneja cada host una tarjeta descartada. Tampoco que la versión desplegada el 28 y 29-oct incluyera `d1cc2a97`.

Build (2026-10-30), en el hilo principal sobre Claude Code (Opus 5.5, observado). Verificadores `hive-verify-task`, con el perfil configurado y sin modelo observado:
- **T1 (`dd39d74e`):** AC1 y AC2 cumplidos en la ronda 1.
- **T2 (`97748873`):** AC3 cumplido en la ronda 1.
- **T3 (`036c962b`, `50e55033`):** AC4 no cumplido en la ronda 1, porque el README omitía `.claude/agents` y `.grok/agents`. Se corrigió y AC4 quedó cumplido en la ronda 2.
- **Checks del candidato final:** `go test ./...` pasa, y también `gofmt` y `go vet` de `tests/content`. El repo no tiene CI.
- **Sin revisión dedicada**, por D7-C.
- **Entrega:** PR #301 integrado como merge `0ade8ca9`. #300 se cerró.

Vigilancias por registrar en Engram al quedar listo el plan (D4-A), con conteo 1 cada una sobre la regla vigente:
- revisor UX de plan fuera de su disparador (`plan-review.md:9`);
- lint omitido antes del push con la regla de #278;
- checkpoints de una sola tarea (`plan-criteria-and-markers.md:13-16`);
- consultas de `hive-design-architecture` sin disparador.
