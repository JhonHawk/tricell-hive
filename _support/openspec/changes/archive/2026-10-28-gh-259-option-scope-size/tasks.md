# Tareas — gh-259-option-scope-size

Base: `0f64489b` (development), confirmada al iniciar el build el 2026-10-28. Rama `fix/gh-259-option-scope-size` en `.claude/worktrees/gh-259-option-scope-size`.

## T1 — La referencia nueva fija la opción mínima y el tamaño en tiempo del agente

- [x] `content/skills/hive-core/references/option-scope.md` existe con el texto del enfoque 1 de `design.md`. `questions.md` conserva solo el puntero del enfoque 2. `flow-close/SKILL.md:47`, `hive-core/SKILL.md`, `instruction-resources.md` y `reading_footprint_test.go` apuntan a ella.

**Closes:** AC1, AC2, AC3.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:**
- `content/skills/hive-core/references/option-scope.md` (nuevo);
- `content/skills/hive-core/references/questions.md:45-47`;
- `content/skills/hive-core/SKILL.md`;
- `content/skills/flow-close/SKILL.md:47`;
- `_support/docs/architecture/instruction-resources.md`;
- `tests/content/reading_footprint_test.go`.

**Execution:** hilo principal. Son pocas líneas acopladas, y T2 y T3 enlazan el ancla nueva.

**Test approach:** check: los `rg` de abajo y `go test ./tests/content/...`.

**Changes:** según los enfoques 1, 2, 4 y 5 de `design.md`.

**Verification:**
- `rg -n "smallest option" content/skills/hive-core/references/option-scope.md` encuentra la regla, y `rg -n "leaves unsolved" content/skills/hive-core/references/option-scope.md` la razón exigida a una Recomendada mayor (AC1).
- `rg -n "person's effort" content/skills/hive-core/references/option-scope.md` la encuentra (AC2).
- `rg -n "no basis at hand|never read, search or delegate only to estimate" content/skills/hive-core/references/option-scope.md` encuentra las dos frases (AC3).
- `rg -n "ten times|not investigated" content/skills/hive-core/references/questions.md` solo encuentra el puntero, sin las Actions viejas.
- `rg -n "size condition of \[questions\]|remaining tasks with their time range" content site` no encuentra nada.
- `rg -n "^## Give the estimated size" content/skills/hive-core/references/option-scope.md` encuentra el encabezado del ancla `#give-the-estimated-size`. Las pruebas no validan anclas: `tooling/management/references.go` las excluye.
- `go test ./tests/content/...` pasa: tamaños bajo 16,000 B y huella de lectura sin error.

## T2 — La ruta de build del plan siempre da el tamaño del plan

- [x] En `plan-handoff.md`, la opción `flow-build` da el tamaño estimado del plan, con tareas y rango de tiempo, como la referencia nueva lo fija, aunque no se cumpla su disparador.

**Closes:** AC4.

**Checkpoint:** C2.

**Depends on:** T1.

**Locations:** `content/skills/flow-plan/references/plan-handoff.md`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check: el `rg` de abajo y `go test ./tests/content/...`.

**Changes:** según el enfoque 3 de `design.md`.

**Verification:** `rg -n "estimated size" content/skills/flow-plan/references/plan-handoff.md` encuentra la frase, con el enlace a `option-scope.md#give-the-estimated-size`.

## T3 — El sitio y el changelog describen el comportamiento

- [x] Las páginas de plan y close en `site/` y `CHANGELOG.md` describen la opción mínima, el tiempo del agente y la ruta de build con tamaño, sin copiar las reglas.

**Closes:** AC5.

**Checkpoint:** C2.

**Depends on:** T1, T2.

**Locations:**
- `site/src/content/docs/flows/plan.md:67,70`;
- `site/src/content/docs/flows/research.md:50`;
- `site/src/content/docs/flows/close.md:50`;
- `CHANGELOG.md` (`[Unreleased]` · `### Changed`).

**Execution:** hilo principal. Es documentación corta de lo que T1 y T2 dejan fijo.

**Test approach:** check: el `rg` de abajo.

**Changes:** según el enfoque 6 de `design.md`.

**Verification:**
- `rg -n "smallest" site/src/content/docs/flows/plan.md site/src/content/docs/flows/research.md` encuentra la opción mínima en las dos páginas, y `plan.md` la enlaza a `option-scope.md`.
- `rg -n "agent's own|size" site/src/content/docs/flows/plan.md` encuentra el tiempo del agente y el tamaño en la ruta de build (líneas 67 y 70).
- `rg -n "agent's own" site/src/content/docs/flows/close.md` encuentra el cambio de la línea 50.
- `rg -n "#259" CHANGELOG.md` encuentra la entrada.

## Verificación y revisión

- C1 — T1 — dos tareas posteriores dependen de él (T2 y T3 enlazan su ancla).
- C2 — T2, T3 — la última tarea `check` cierra el último checkpoint.

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite completa | candidato final, worktree local | `go test ./...` y `go vet ./...` pasan | implementación autorizada |
| Revisión dedicada | candidato local, antes del push | `/code-review` de Claude Code; hallazgos resueltos o refutados con evidencia | D6-A |
| Tamaño de `global.md` | candidato final | `wc -c content/guidance/global.md` da 17,773 B (sin cambio) | restricción |
| Enlaces de instrucciones | candidato final ya committeado, worktree local | `hive update --dry-run --source <worktree>` pasa `validateInstructionReferences` sin error de enlaces (no valida anclas) | implementación autorizada; solo vista previa, no despliega |

No hay prueba in vivo: el cambio es redacción de guía, y un piloto de comportamiento está pausado (`AGENTS.md`, Measurement).

## Estado de revisión y progreso

- Plan guardado el 2026-10-28. Revisión R1 de consistencia de la guía (`hive-review-plan`, nativo; ~187 s, ~119k tokens), sobre `4357f6af` / `66e75d18` / `5015e704`:
  - Bloqueante 1, alcance que choca con la ruta de build y el cierre: aceptado. La condición se limita a dos o más opciones que empiezan trabajo sobre el mismo pedido y construyen cantidades distintas, y nombra la ruta de build y el cierre como fuera. Es la opción (a) del revisor, reconciliada sin otra ronda.
  - Bloqueante 2, copias parciales en `flow-close/SKILL.md:47` y en el texto nuevo de `plan-handoff.md`: aceptado. Los dos nombran el tamaño "as option-scope sets", sin repetir sus partes. Propuesta del revisor.
  - Bloqueante 3, falta `site/.../research.md:50`: aceptado y agregado a T3.
  - Bloqueante 4, verificación incompleta: aceptado. Se agregan `rg` para "leaves unsolved", para el ancla y para las tres páginas. Se quitó la validación de enlaces atribuida a `go test` y entró el gate `hive update --dry-run`.
  - No bloqueantes aceptados: el texto de la viñeta de `hive-core/SKILL.md`, la prosa de `instruction-resources.md` y la lista de lugares revisados en el PR.
  - No bloqueante descartado: agregar la referencia al escenario `question` de la huella, porque una decisión suelta solo la lee cuando cumple su condición.
  - Límite que queda: ninguna prueba valida anclas.
- Build (2026-10-28):
  - T1 en `ecae4053`, T2 en `74e74a89` y T3 en `05ed30cb`.
  - Verificación con `hive-verify-task`: C1 (T1: AC1–AC3 met) y C2 (T2: AC4 met; T3: AC5 met). Los verificadores son nativos, sobre el modelo configurado del rol; no observé el modelo de cada uno. El implementador fue el hilo principal (Opus 5.5).
  - Gates del candidato `05ed30cb`: `go test ./...` pasa; `go vet ./...` pasa; `global.md` mide 17,773 B, sin cambio; `hive update --dry-run --source .` (binario instalado) arma el plan desde `05ed30cb` con exit 0 y no cambia la instalación.
  - Siguiente paso: flow-verify con `/code-review` (D6-A) antes del push.
- Verify (2026-10-28): `/code-review` medio sobre `05ed30cb` (3 commits; ~61 s, ~95k tokens). Dio 7 hallazgos de redacción, de severidad P2 o menor.
  - Corregidos:
    - (1) La frase de `plan-handoff.md` "even when that reference's trigger does not apply" contradecía la condición de `option-scope.md`.
    - (2) y (3) Los punteros de `questions.md` y `hive-core/SKILL.md` perdían "as far as you can tell without estimating".
    - (4) y (6) La frase de excepción de `option-scope.md` describía mal la ruta de build y el cierre, que tienen dos opciones de trabajo de igual cantidad; se reescribió como aclaración exacta.
    - Commits: `af30a106` (T1) y `6d3110c4` (T2). `design.md` quedó sincronizado.
  - Descartados:
    - (5) `flow-close` da el tamaño una vez por pregunta: es la excepción en su punto de uso que nombra el elemento que cambia, como pide `AGENTS.md`.
    - (7) El enlace del sitio a `blob/master` sigue la convención vigente: el sitio se publica desde `master`, con el mismo caso que `loops.md` en #256.
  - Sin re-revisión: no hubo P0 ni P1. Después de las correcciones, `go test ./tests/content/...` pasa y `hive update --dry-run` desde `6d3110c4` sale con exit 0.
  - T1 y T2 vuelven a verificación focal (AC1, AC3 y AC4). T3 conserva su evidencia, porque las correcciones no tocan archivos del sitio.
  - Re-verificación focal sobre `6d3110c4`: AC1, AC3 y AC4 met; AC2 sigue met. T2 vuelve a `[x]`.
  - Observaciones del verificador corregidas en `655d9130` (T1): la condición de lectura de `option-scope.md` no tenía el calificador, y la aclaración de cantidad igual no era exacta para un ticket nuevo. T1 vuelve a verificación focal de AC1 y AC3.
  - Re-verificación focal de T1 sobre `655d9130`: AC1 y AC3 met. T1 vuelve a `[x]`. Las tres tareas están verificadas.
- Integración (2026-10-28): push de `fix/gh-259-option-scope-size` con la base al día; [PR #260](https://github.com/JhonHawk/tricell-hive/pull/260) integrado con merge commit `48a616b8`. Sin CI (retirada). El issue #259 se cerró al integrarse. `hive update` desplegó `48a616b8` localmente (plan guardado y aplicado con `hive apply`).
