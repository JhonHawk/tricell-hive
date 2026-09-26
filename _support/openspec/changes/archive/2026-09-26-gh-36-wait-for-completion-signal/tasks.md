# Tareas

### T1 — Regla de espera en `global.md`

- [x] `content/guidance/global.md:82` tiene la regla de [design.md](design.md#regla-nueva-en-globalmd) en lugar de la cola actual del punto.
- [x] `globalGuidanceBudget` en `tests/content/budget_test.go` sube al tamaño medido: de 38 352 a 39 033 bytes (+681).

**Evidencia:** con el presupuesto anterior, `TestGlobalGuidanceStaysWithinRatchetedBudget` falla ("39033 bytes, over the 38352-byte budget"); con el nuevo, `go test -count=1 ./tests/content/...` pasa.

**Depende de:** nada.
**Ubicaciones:** `content/guidance/global.md`, `tests/content/budget_test.go`.
**Ejecución:** hilo principal. Son dos ediciones cortas y el brief para delegarlas sería más largo que el cambio.
**Enfoque de prueba:** `check`: `go test -count=1 ./tests/content/...` pasa con el presupuesto nuevo y falla con el anterior.
**Verificación:** `go test -count=1 ./tests/content/...` en verde; `rg -n "sleep-and-check" content/guidance/global.md` devuelve una línea.

### T2 — Parser `collab_tool_call` y criterio `no_poll_wait_chain`

- [x] Rama `codex` de `parseTrace` según [design.md](design.md#parser-de-trazas-de-codex).
- [x] `noPollWaitChain` según [design.md](design.md#criterio-no_poll_wait_chain), declarado sin conectar a `regressionCriteria`, con su doc comment de límites. El comentario de `regressionCriteria` lo menciona junto a `flowSkillReadBeforeDelivery`, y `TestRegressionCriteriaReturnsAllSix` no cambia.
- [x] Los nueve fixtures de [design.md](design.md#fixtures) con su `provenance.md`, registrados en `TestRegressionFixtures`. En `tests/fixtures/regression/README.md`, una fila en la tabla y un párrafo en "Limits": criterio declarado sin conectar, límites en este `design.md`.
- [x] Tests unitarios:
  - un `collab_tool_call` `wait` produce un evento `collab_wait` con `Input` que conserva `agents_states`;
  - una traza sin esperas da `not_observed`.

**Evidencia (T2):**
- **Delegación:** `test-engineer` nativo, sin commits, en las rutas del brief.
- **RED:** `go vet` falló con `undefined: noPollWaitChain`, antes de implementar.
- **Reversión:** cada una de las 8 ramas de la tabla hace cambiar de estado su fixture.
- **Corrección del hilo principal:** el subagente leyó "con `statusCheckRollup`" como requisito de las cuatro consultas. Así, `gh run view <id>` a secas, el patrón real de las sesiones, no se habría detectado. Ahora el marcador se exige solo en `gh pr view`; los fixtures de CI vuelven a `--json status`, y un test nuevo (`TestGhStatusQueryRequiresMarkerOnlyForPrView`) cubre la distinción. Volver a la lectura literal hace fallar ese test y los fixtures de CI.
- **Resultado final:** `go test -race ./tests/pilot/...` pasa y el escaneo de fixtures no imprime nada.
- **Sin tocar, anterior al cambio:** `gofmt` no es idempotente en un comentario de `regression.go`, y `regression_test.go` tiene una línea en blanco extra al final.

**Depende de:** nada; es independiente de T1.
**Ubicaciones:** `tests/pilot/trace.go`, `tests/pilot/trace_test.go`, `tests/pilot/regression.go`, `tests/pilot/regression_test.go`, `tests/fixtures/regression/no_poll_wait_chain/**`, `tests/fixtures/regression/README.md`.
**Ejecución:** delegado a `test-engineer`. La interfaz está fijada en el diseño, no se superpone con los archivos de T1 y así la lectura de `regression.go` (1870 líneas) no queda en el hilo principal. Sin commits; el hilo principal revisa el diff.
**Enfoque de prueba:** `tdd`: primero el test del parser y los fixtures registrados (rojo por la función que falta), después la implementación.
**Verificación:**
- `go test -race ./tests/pilot/...` en verde, incluidos los fixtures previos sin cambio de estado.
- Reversión por rama, restaurando cada una después. Cada rama quitada hace fallar su fixture:

  | Rama quitada | Fixture que falla |
  | --- | --- |
  | ciclo en un solo comando | `codex-ci-fail` |
  | `sleep` suelto | `codex-ci-bare-sleep-fail` |
  | corte por otra herramienta | `codex-ci-single-pass` |
  | exclusión de `--log-failed` | `codex-ci-log-failed-pass` |
  | cadena de esperas | `codex-wait-fail` |
  | corte por texto | `codex-wait-pass` |
  | exclusión de esperas fallidas | `codex-wait-failed-pass` |
  | excepción por resultado nuevo | `codex-wait-result-pass` |
- El escaneo de fixtures del README no imprime nada.

### T3 — Verificación conjunta, nota en #36 y entrega

- [x] `go vet ./...` y `go test -race -count=1 ./...` en verde, en el árbol de trabajo sin cambios ajenos, sobre `a64fd38` más este cambio.
- [x] [Comentario en #36](https://github.com/JhonHawk/tricell-hive/issues/36#issuecomment-5850900888) (D2-A): por qué la regla 3 queda fuera y por qué las reglas 1 y 2 se unieron en tramos.
- [x] Entrega según [proposal.md](proposal.md#entrega) (E1-A, E3-A):
  - **Commit y push:** `5bf5b19` (20 rutas, gitleaks sin hallazgos), con push a `rebuild/harness-engineering`; la cabeza remota es `5bf5b19`.
  - **Release:** `86b08f50d04a`, desplegada desde un worktree limpio en `5bf5b19` (plan `b68e02a811c3`), con 431 recursos `installed`.
  - **Verificación por host:** la regla está en `~/.claude/CLAUDE.md` (Claude y Grok), `~/.codex/AGENTS.md`, `~/.cursor/AGENTS.md`, `~/.config/opencode/AGENTS.md` y `~/.pi/agent/AGENTS.md`.
  - **Limpieza:** el worktree y el plan se borraron.
  - **Cierre:** en el commit de archivo que sigue.

**Depende de:** T1 y T2.
**Ejecución:** hilo principal (dueño de la integración, el tracker y Git).
**Verificación:** los comandos anteriores en verde; `gh issue view 36 --comments` muestra el comentario.

## Verificación compartida

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Tests de contenido y pilot | Tras T1 y T2, local | `go vet ./...`, `go test -race ./...` en verde | Incluido en la implementación |
| Reversión por rama | T2, local | Cada fixture que falla deja de fallar al quitar su rama | Incluido en la implementación |
| Verificación in vivo | — | No aplica: el cambio no tiene superficie ejecutable fuera de los tests, y los pilotos de modelo están pausados (`AGENTS.md`, Measurement) | — |

## Revisión y progreso

**Revisión del plan, ronda 1** (revisión `f967fe46bd4f`): dos `review-plan` nativos en paralelo, de solo lectura.

- **Redacción de la guía:**
  - Bloqueantes aceptados: B1 (con notificación, terminar el turno) y B2 (la línea va en la misma respuesta que la espera).
  - No bloqueantes aceptados: N1 (reanudar la misma espera o relanzar el mismo comando) y N2 (el aviso lo da el hilo principal).
  - Límite: no verificó el resto de `global.md` fuera de `:1-40` y `:75-95`.
- **Modelo de trazas:**
  - Bloqueantes aceptados:
    - B1: se define qué corta una cadena, y las esperas fallidas no cuentan.
    - B2: se fija el "siguiente shell", se excluye `--log`, y los bucles dentro de un comando son un límite intencional.
  - Aparte, se quitó la excepción de estado terminal, porque no se verificó qué agentes lista `agents_states`.
  - No bloqueantes aceptados: N1 (dos fixtures más y el test de no observado), N2 (`item.started`), N3 (criterio sin conectar) y N4 (límite de formato).

Los cambios de la regla aplican las propuestas del propio revisor.

**Re-revisión del modelo de trazas** (revisión `bb5d4b0338b8`, el mismo revisor reanudado; única ronda):
- **B1 aplicado:** el fixture de `--log-failed` no ejercía su rama, así que pasó a ser un fixture aparte.
- **N1 aplicado:** vuelve la excepción por resultado, ahora solo para agentes que pasan a estado terminal respecto del evento `collab_*` anterior, con el fixture `codex-wait-result-pass`.
- **N2 aplicado:** se excluye `--watch`, `gh run watch` se reconoce con los mismos prefijos, y cualquier otro evento de herramienta corta la cadena de shell.

Las tres son propuestas del propio revisor, aplicadas tal cual, así que se concilian sin otra ronda.

**Límites que quedan:**
- No se ejecutó ningún test durante la revisión.
- Que Codex emita `agent_message` entre llamadas `collab` se toma del esquema, no de una corrida real.
- El revisor de la guía no leyó `global.md:41-74` ni de `:96` en adelante.

**Estado:** listo para implementar. La entrega está decidida (E1-A, E2-A, E3-A); la ejecución espera la confirmación para `flow-build`.
