# Tareas — gh-50-codex-sandbox-browser-gate

Commit base: `64e9251c9baae866e109695bde0647bc9e527d2c` (`development`).

### T1 — El perfil `observe` de Codex deja de declarar `sandbox_mode`

- [x] `observe` de Codex queda vacío, el validador tolera la clave heredada de releases congeladas y el renderizado de Codex nunca la emite.

**Closes:** AC1, AC2.

**Depends on:** none.

**Locations:** `integrations/agent-profiles.json:10`; `integrations/agents/testdata/profiles.json:10`; `tooling/cli/testdata/agent-profiles.json:10`; `tooling/management/testdata/agent-profiles.json:10`; `integrations/agents/agents.go:274-278` (`validateAccess`); `integrations/agents/agents_test.go:160-175` (`TestCodexTOMLIsFlatAndEscapesInstructionBody`) y las pruebas de rechazo en `:260-280`; `integrations/agents/testdata/render.golden:95,179`; `tooling/management/testdata/state_without_overrides.golden` (regenerado con `-update-state-golden`: los TOML de Codex embebidos y sus hashes cambian al perder la clave).

**Execution:** delegada a `hive-build-backend` (cambio Go con pruebas en varios archivos).

**Test approach:** tdd.

**Changes:** quitar la clave de los cuatro JSON; `validateAccess` conserva el caso de Codex como clave heredada (valores `read-only` y `workspace-write`); `Render` salta `sandbox_mode` para Codex; `TestCodexLegacySandboxModeIsReadableAndNotRendered` inyecta la clave en el bloque de Codex, espera que `ReadProfiles` la acepte y que el TOML no la contenga, y rechaza un valor desconocido; `TestCodexTOMLIsFlatAndEscapesInstructionBody` deja de esperar la línea; `render.golden` regenerado solo en los bloques de Codex. (Primera versión: el validador rechazaba la clave; revertido tras la segunda ronda de revisión, ver AC2.)

**Verification:** `go test ./integrations/agents/ ./tooling/cli/ ./tooling/management/` en verde; `rg -n sandbox_mode integrations tooling` solo devuelve el validador, el renderizador y las dos pruebas.

### T2 — `agent-delivery.md` describe lo que Codex hace con el sandbox del rol

- [x] La línea 44 ya no afirma valores de solo lectura para Codex; la sección cita la fuente 0.163.0, la prueba en vivo y el ajuste de sesión para las compuertas de navegador.

**Closes:** AC4.

**Depends on:** none.

**Locations:** `_support/docs/architecture/agent-delivery.md:44`; tabla histórica `:163` se conserva.

**Execution:** hilo principal.

**Changes:** según "Diseño propuesto" en `proposal.md`.

**Verification:** lectura del párrafo; `rg -n 'network_access|0.163.0' _support/docs/architecture/agent-delivery.md` encuentra la nueva frase.

### T3 — La guía distribuida reporta la compuerta bloqueada por el sandbox del host

- [x] `browser-automation.md` y `verification.md` dicen qué reporta el hijo cuando el sandbox del host impide el navegador o la red local.

**Closes:** AC3.

**Depends on:** none.

**Locations:** `content/skills/flow-build/references/browser-automation.md:35-37` ("Escalate before declaring a check unverifiable"); `content/skills/flow-build/references/verification.md:61`.

**Execution:** hilo principal.

**Changes:** según "Diseño propuesto"; texto con la capacidad primero, sin nombrar hosts.

**Verification:** `rg -n -i 'sandbox' content/skills/flow-build/references/browser-automation.md content/skills/flow-build/references/verification.md` muestra las frases nuevas; `go test ./tooling/...` sigue en verde (los validadores de contenido leen las skills).

### T4 — Entrega automática y cierre

- [x] PR fusionada en `development`, #50 cerrado con el resultado, instalación local refrescada (los tres pasos se completan en el cierre que versiona este registro en la misma PR).

**Depends on:** T1, T2, T3.

**Execution:** hilo principal.

**Changes:** `go vet ./...` y `go test ./...` sobre el candidato final; push; PR con la lista de lugares revisados y el tamaño de `global.md` (sin cambio); `/code-review` de Claude Code; merge; `hive update`; cierre de #50; `flow-close` archiva este registro.

**Verification:** merge observado en `development`; `gh issue view 50` en estado cerrado; `hive status` con el commit desplegado.

## Verificación y revisión humana

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go vet ./...` y `go test ./...` en verde | Implementación |
| Verificador por tarea | Tras T1 | `hive-verify-task` sobre AC1 y AC2 | Implementación |
| Revisión de código | PR abierta | `/code-review` de Claude Code; hallazgos bloqueantes corregidos o refutados | D3-A (2026-10-16) |

Sin efecto de UI renderizada: no aplican `hive-review-ux` ni `hive-verify-change`.

## Estado de revisión y progreso

- 2026-10-16: rama creada desde `64e9251c`; T1 delegada; T2 y T3 en el hilo principal.
- 2026-10-16: T2 (9600e0a1) y T3 (3ca6e6ed) verificadas por lectura; T1 (86d2160d) implementada por `hive-build-backend` (modelo Sonnet del perfil `execution`, configurado, no observado) y verificada por `hive-verify-task` (modelo Opus del perfil `verifier`, configurado, no observado): AC1 y AC2 cumplidos. `go vet ./...` limpio y `go test ./...` con 18 paquetes en verde sobre el candidato final. PR [#61](https://github.com/JhonHawk/tricell-hive/pull/61) abierta.
- 2026-10-16: `/code-review` de Claude Code, dos rondas. Ronda 1: 7 hallazgos no bloqueantes, 6 corregidos en 7f83ead8 (redacción neutral por host, el padre distingue sandbox de servidor caído, etiquetas de evidencia, fila 0.155.1 reconciliada, comentario en `validateAccess`, changelog sin anclar a una versión) y 1 incidental fuera de alcance (Cursor `readonly` quizá tampoco se aplica). Ronda 2: 6 hallazgos, uno real de compatibilidad (rechazar la clave rompía `hive models` con un binario nuevo sobre la release instalada) corregido en 476d7d06 tolerando la clave heredada y omitiéndola al renderizar, AC2 reescrito y T1 reverificada por `hive-verify-task` (AC1 y AC2 cumplidos); los cinco de redacción corregidos en el mismo commit. Sin tercera ronda de revisión: las correcciones de la ronda 2 quedan cubiertas por la suite (`go vet` limpio, 18 paquetes en verde) y la reverificación de T1.
