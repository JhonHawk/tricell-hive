# Tareas

Base del build: `e4fae19ec372245aa2a4ca302db8420f9d70436f`.

### T1 — Detectar la declaración en settings

- [x] Una función pura, con `settings.json` de fixture, clasifica ausente, presente, conflicto (`extensions: []`) y fuente de usuario.

**Closes:** AC5.

**Depends on:** none.

**Locations:** paquete nuevo o `integrations/pi` (p. ej. `package_settings.go`); pruebas junto al paquete.

**Execution:** `hive-build-backend` (lógica Go acotada).

**Test approach:** tdd.

**Changes:** parsear `packages` como lista de cadenas u objetos `{source}`; identidad `npm:pi-subagents` con o sin pin; `extensions: []` y filtros no vacíos según [design.md](design.md).

**Verification:** `go test` del paquete: fixtures de ausente, `@0.67.0`, `@0.74.0` sin pin, objeto con `extensions: []` → conflicto, JSON inválido → error; no toca el home real.

### T2 — El plan de usuario con `pi` incluye el paso o la omisión

- [x] `Plan` con host `pi` y home sintético muestra `install` `npm:pi-subagents@0.74.0` si falta, omisión si está, conflicto si `extensions: []`, y no aparece si `pi` no está en el plan.

**Closes:** AC1.

**Depends on:** T1.

**Locations:** `tooling/management/plan.go`, `types.go`; pruebas en `tooling/management`.

**Execution:** `hive-build-backend`.

**Test approach:** tdd.

**Changes:** campo de paso de paquete en `Plan` (no un `Change` de archivo). Home sintético vía `Options` existente.

**Verification:** `go test ./tooling/management` en los casos de AC1; un plan solo `claude` no lleva el paso.

### T3 — Apply ejecuta `pi install` solo cuando el plan lo pide

- [x] Apply con `pi` falso en PATH y `PI_CODING_AGENT_DIR` al home sintético: install añade la fuente Hive; omisión no invoca el binario; otras entradas de `packages` se conservan.

**Closes:** AC2, AC3.

**Depends on:** T2.

**Locations:** `tooling/management/apply.go` (o helper de proceso); `pi` falso en testdata.

**Execution:** `hive-build-backend`.

**Test approach:** tdd.

**Changes:** `LookPath("pi")`, args estructurados, sin shell. Releer settings inmediatamente antes de instalar; si el paquete ya está, error de plan obsoleto sin invocar `pi`. El falso escribe `settings.json` y `npm/node_modules/pi-subagents/package.json` (o la ruta convencional del instalador de Pi) con `name`/`version`. AC3 no lanza Pi ni lista roles: es esa declaración más esos archivos. La receta nativa queda documentada (Pi 1.0 `docs/packages.md`, pin 0.74.0); el falso no la sustituye.

**Verification:** `go test` del apply: argv/env como arriba; omisión: cero invocaciones; entrada ajena intacta; entre plan y apply aparece `pi-subagents` → apply no llama a `pi`.

### T4 — Recover solo deshace lo que esta operación añadió

- [x] Recover llama a `pi remove` con la fuente journalizada solo si Hive la añadió y la cadena actual coincide; presencia previa: cero `remove`.

Tras H1: el deshacer de un apply terminado es `hive remove` del host Pi (campo `State.PiSubagentsSource`), no `Engine.Recover` de una transacción committed. Recover de una interrupción no llama a `pi remove`.

**Closes:** AC4.

**Depends on:** T3.

**Locations:** recover/journal en `tooling/management`.

**Execution:** `hive-build-backend`.

**Test approach:** tdd.

**Changes:** el journal guarda dueño, fuente exacta y si apply terminó. Recover deshace solo un apply terminado dueño Hive. Resultado desconocido (proceso arrancó, settings no confirmada): no remove. Pin cambiado: no remove. Presencia previa: no remove.

**Verification:** `go test`: (1) apply terminado+recover en ausencia → `remove` de `npm:pi-subagents@0.74.0`; (2) presencia previa → cero `pi`; (3) pin cambiado → no remove; (4) apply interrumpido (falso que no escribe settings) → recover no remove.

### T5 — Prueba de conflicto y documentación

- [x] Conflicto `extensions: []` no aplica; nota en `agent-delivery.md` de que el gestor instala el paquete en alcance de usuario.

**Closes:** AC5.

**Depends on:** T2, T3.

**Locations:** pruebas de T2/T3; `_support/docs/architecture/agent-delivery.md` (frase que hoy dice que Hive no instala la extensión).

**Execution:** hilo principal (doc) y `hive-build-backend` (prueba de conflicto si no quedó en T2).

**Test approach:** tdd para el conflicto; check para el doc (grep de la frase nueva).

**Changes:** no aplicar el paso en conflicto; el plan falla o se niega con mensaje que nombra `extensions: []`.

**Verification:** `go test` del caso conflicto; el doc ya no dice que Hive no instala la extensión.

## Verificación compartida

| Compuerta | Cuándo | Evidencia | Autoridad |
| --- | --- | --- | --- |
| `go test` de `integrations/pi` y `tooling/management` | cada tarea T1–T4 | todos pasan; sin red | implementación |
| Sin piloto de modelo | no corre | AC3 es archivos | pausa de pilotos |
| `hive-review-code` | PR | D2-A | entrega |

## Revisión del plan y progreso

Revisado: `hive-review-plan` gestor, 2026-10-01. H1–H5 incorporados (sin otra ronda: son las correcciones del revisor). Recover tras éxito = deshacer Hive; interrupción = no remove. Sin rama `reinstall-user`.

Siguiente: `flow-build` T1, si se autoriza la implementación.
