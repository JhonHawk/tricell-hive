# Tareas

Base del cambio: `092ea3bb` (`feat/flash-class-guidance`). La rama avanzó desde `34b4edc9`, la revisión del plan, con un commit de otra sesión que solo toca `delivery-decisions.md` y otra sección de `CHANGELOG.md`; los criterios siguen siendo falsos en `092ea3bb`.

Todas las tareas se ejecutan en el hilo principal: cada una es una edición corta de texto cuyo brief sería más largo que el cambio, y T4 depende del texto final de T1–T3.

### T1 — Contrato de retorno de `hive-research`

- [x] El rol exige la respuesta, estado con motivo de parada, afirmaciones con etiqueta, ubicación y extracto, contraevidencia, lo no revisado y tope de extensión, con el vocabulario global.

**Closes:** AC1, AC3, AC6.

**Depends on:** ninguna.

**Locations:** `content/agents/review/hive-research.md` pasos 2 y 4.

**Execution:** hilo principal.

**Test approach:** check, con los `rg` de la verificación.

**Changes:** texto de `design.md`, «Texto propuesto», pasos 2 y 4.

**Verification:** en `content/agents/review/hive-research.md`, cada uno de estos `rg -n` devuelve al menos una línea: `"answer in one or two sentences"`, `"complete"`, `"partial"`, `"blocked"`, `"locator"`, `"excerpt"`, `"counterevidence"`, `"did not check"`, `"1,500"`, `"inspected fact"`, `"source claim"`; `rg -n "source facts|live observations"` no devuelve nada. El estado `partial` cubre la parte de AC3 que dice cuál parada ocurrió.

### T2 — Brief, verificación acotada y estado en `flow-research`

- [x] La skill pide en el brief lo sabido, lo excluido y la parada; limita la comprobación a abrir ubicaciones de afirmaciones decisivas; usa `hive-refute-claim` para premisas no resueltas; envía el estado largo a `hive-read-state`; usa el vocabulario global.

**Closes:** AC2, AC3, AC5, AC6.

**Depends on:** T1 (la skill remite al contrato del rol).

**Locations:** `content/skills/flow-research/SKILL.md` líneas 16, 20 y 24.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** texto de `design.md`, «Texto propuesto», para las líneas 16, 20 y 24.

**Verification:** en `content/skills/flow-research/SKILL.md`, cada uno de estos `rg -n` devuelve al menos una línea: `"hive-refute-claim"`, `"one verdict per premise"`, `"hive-read-state"`, `"locator"`, `"out of scope"`, `"scope is exhausted"`, `"inspected facts"`; `rg -n "consequential claims against their sources|documented intent" content/skills/flow-research/SKILL.md` no devuelve nada.

### T3 — Lanzamiento conjunto y primer explorador en la delegación

- [x] La referencia de delegación ordena lanzar juntos los hijos independientes y un solo hijo primero cuando no se sabe dónde está la respuesta; etiqueta las premisas con el vocabulario global.

**Closes:** AC4, AC6.

**Depends on:** ninguna.

**Locations:** `content/skills/flow-research/references/delegation.md` líneas 11 y 27.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** texto de `design.md`, «Texto propuesto», para `delegation.md`.

**Verification:** en `content/skills/flow-research/references/delegation.md`, cada uno de estos `rg -n` devuelve al menos una línea: `"same step"`, `"one child to locate"`, `"inspected facts, source claims, or assumptions"`; `rg -n "verified facts, source claims" content/skills/flow-research/references/delegation.md` no devuelve nada.

### T4 — Página del sitio, changelog y barrido de coherencia

- [x] La página de research describe el nuevo comportamiento, `CHANGELOG.md` lo registra y ninguna otra superficie contradice o duplica las reglas de T1–T3.

**Closes:** AC7.

**Depends on:** T1, T2, T3.

**Locations:** `site/src/content/docs/flows/research.md` líneas 35 y 71–72; `CHANGELOG.md` `[Unreleased]`; barrido en `content/guidance/global.md`, `content/skills/**`, `content/agents/**`, `site/`, `_support/docs/architecture/`, `tests/pilot/`.

**Execution:** hilo principal.

**Test approach:** check.

**Changes:** frase del sitio y entrada de changelog de `design.md`; ajustar la respuesta de la pregunta frecuente si contradice el texto nuevo. Barrido con `rg -n -i "consequential claims|counterevidence|source facts|verified facts|hive-read-state|hive-refute-claim|in waves" content site _support/docs/architecture tests/pilot`; anotar en «Estado» de este archivo los lugares revisados para la descripción del PR.

**Verification:** en `site/src/content/docs/flows/research.md`, cada uno de estos `rg -n` devuelve al menos una línea: `"did not check"`, `"cited lines"`, `"status"`; `rg -n "hive-research. returns a bounded reply" CHANGELOG.md` devuelve la entrada; el barrido no muestra instrucciones en conflicto (resultado anotado abajo).

## Verificación compartida

| Control | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Pruebas Go | Al final, sobre el árbol de trabajo de `feat/flash-class-guidance` | `go vet ./...` y `go test ./...` pasan; `go test ./tests/content/ -run TestGuidanceReadingFootprint -v` registra el nuevo tamaño de `flow-research/SKILL.md` y `delegation.md` | Cubierto por la implementación |
| Prueba en sesión real | Tras despliegue local, solo si el usuario lo pide | El usuario corre sesiones de research y decide merge o ajustes | Pendiente del usuario |

No hay verificación de interfaz: el cambio no tiene efecto visual renderizado (la página del sitio cambia solo texto).

## Revisión y estado

- Revisión del plan (2026-10-13), dominio único: guía distribuida de agentes, rol `hive-review-plan` nativo en Claude Code.
  - Ronda 1 sobre la primera versión guardada: 2 bloqueantes (el retorno no pedía la respuesta; faltaba la entrada de `CHANGELOG.md`) y 8 sugerencias (correspondencia estado/parada, espera externa, un veredicto por premisa, afirmaciones de recomendación, AC3 en T1, comprobaciones `rg` por elemento, consulta de tracker para `hive-read-state`, explorador tras la búsqueda propia). Todas aplicadas; la cláusula contra bucles `sleep` se retiró por tener una sola sesión.
  - Re-revisión (única ronda) sobre la versión corregida: sin bloqueantes. Sus dos sugerencias (seguir el procedimiento de espera que nombre el brief; pedir a `hive-read-state` los campos por los que agrupa el informe) se aplicaron tal cual, sin otra ronda. La tercera (registrar la vigilancia de `d4947fae`) quedó en la memoria del mantenedor.
  - Límite: revisión del texto del plan; nada ejecutado ni comportamiento en hosts.
- Modelos para `hive-verify-task`: implementador = hilo principal, Claude Opus 5.5 (observado). Verificador configurado = perfil `verifier` de Claude, `opus` (configurado, no observado): mismo modelo, así que cada verificación se lanza con `model: sonnet` mediante la opción de modelo del lanzador.
- Línea base: `go vet ./...` y `go test ./...` pasan en `092ea3bb` más el árbol de trabajo sin cambios de código.
- T1: verificada (`hive-verify-task`, sonnet): AC1 cumplido; AC3 y AC6 en su parte. Tras el veredicto se aclaró `blocked` («when the answer cannot be given because…») para no chocar con el paso 3 (credenciales no bloquean todo); las comprobaciones `rg` siguen pasando.
- T3: verificada (`hive-verify-task`, sonnet): AC4 cumplido; AC6 en su parte.
- T2: verificada (`hive-verify-task`, sonnet): AC2, AC3, AC5 y AC6 cumplidos (AC3 y AC6 ya completos con T1 y T3).
- Commits: T1 `6502bdf6`, T3 `8a880797`.
- Commit T2: `7b999437`.
- Barrido de T4 (2026-10-13): `rg -n -i "consequential claims|counterevidence|source facts|verified facts|hive-read-state|hive-refute-claim|in waves|live observations" content site _support/docs/architecture tests/pilot`. Lugares revisados y resultado: `content/guidance/global.md`, leído directamente porque el `rg` no da coincidencias ahí (líneas 60, 74, 78–80: vocabulario y puntero a delegación, compatibles); `content/skills/flow-research/SKILL.md` y `references/delegation.md` (texto nuevo); `content/skills/flow-plan/references/delivery-decisions.md:29` y `content/skills/flow-build/references/verification.md:41` (uso de `hive-refute-claim` en revisión de código, sin cambio); `content/agents/review/hive-refute-claim.md` y `content/agents/quality/hive-read-state.md` (contratos sin cambio, compatibles según las verificaciones de T2); `_support/docs/architecture/agent-delivery.md:30` (esfuerzo `low` de `hive-read-state`, sin cambio); `site/.../flows/research.md` líneas 35 y 71–72 (actualizada la 35; la FAQ sigue siendo cierta); `tests/pilot` (ningún detector depende de la redacción cambiada). Sin instrucciones en conflicto ni duplicadas. `README.md` y `llms.txt` no requieren cambio.
- T4: verificada (`hive-verify-task`, sonnet): AC7 cumplido; barrido confirmado.
- Commit T4: `5d15446f`.
- Pruebas compartidas sobre `5d15446f`: `go vet ./...` y `go test ./...` pasan; `TestGuidanceReadingFootprint` registra `flow-research/SKILL.md` 10.131 bytes (antes 9.382) y `delegation.md` 7.612.
- Despliegue local (2026-10-13): binario `hive` recompilado desde la rama; `hive update --source . --out <plan>` y `hive apply --plan <plan>` desde `5d15446f`, transacción `e9252cee1aa13dce6512c88a97ecae28`, seis hosts (claude, codex, cursor, grok, opencode, pi). Comprobado el texto nuevo de `hive-research` en los seis y de la skill y `delegation.md` en `~/.agents/skills`.
- Pausa: carpeta sin archivar ni versionar hasta el merge a `development`.
- Siguiente paso: auditoría del comportamiento en una sesión nueva; después, merge o ajustes.
