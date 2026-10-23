# Tareas — gh-60-plan-revision-and-write-docs

Commit base: `dae2354e89cc` (`development`).

### T1 — Bloque "Revise a plan in progress" y sus siete enlaces

- [x] `flow-plan/SKILL.md` tiene la sección y las siete frases enlazan a ella.

**Closes:** AC1.

**Locations:** `content/skills/flow-plan/SKILL.md` (sección nueva y línea 18); `content/skills/flow-build/SKILL.md:16,48`; `content/skills/flow-plan/references/questions.md:64`; `content/skills/flow-build/references/verification.md:73`; `content/skills/flow-plan/references/plan-review.md:38`; `content/skills/flow-plan/references/plan-format.md:20`.

**Execution:** hilo principal (redacción acoplada entre siete archivos; un brief sería más largo que el cambio).

**Verification:** `rg -n 'revise-a-plan-in-progress' content` devuelve la sección y siete enlaces; `go test ./tests/content/...` (enlaces de skills) en verde.

### T2 — `hive-write-spec` → `hive-write-docs` con contrato acotado

- [x] Rol renombrado y acotado; nombre viejo solo en la nota de `agent-delivery.md` y en los registros archivados.

**Closes:** AC2.

**Locations:** `content/agents/docs/hive-write-docs.md` (nuevo), `content/agents/docs/hive-write-spec.md` (borrado), `_support/docs/architecture/agent-delivery.md:9,20,30`.

**Execution:** hilo principal (renombre mecánico más redacción).

**Verification:** `rg -n 'hive-write-spec' content integrations tooling tests site README.md llms.txt _support/docs` solo devuelve la nota del renombre; `go test ./integrations/... ./tooling/...` en verde (renderizan cada rol).

### T3 — Registrar el costo de la revisión de planes

- [x] `plan-review.md` pide anotar duración y tokens por revisor y re-revisión.

**Closes:** AC3.

**Locations:** `content/skills/flow-plan/references/plan-review.md:36`.

**Execution:** hilo principal.

**Verification:** `rg -n -i 'duration|token' content/skills/flow-plan/references/plan-review.md` devuelve la frase.

### T4 — Sitio, changelog y cierre de #60

- [x] `plan.md` y `build.md` describen la revisión a medio camino; `CHANGELOG.md` registra el renombre y los dos textos (3dc79121, `pnpm build` completo); el comentario de cierre en #60 va con T5.

**Closes:** AC4.

**Locations:** `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/build.md`, `CHANGELOG.md`.

**Execution:** hilo principal.

**Verification:** `pnpm --dir site build` en verde; lectura de las dos páginas.

### T5 — Entrega automática y cierre

- [x] PR fusionada en `development`, #60 cerrado, instalación local refrescada (se completa en el cierre que versiona este registro en la misma PR).

**Depends on:** T1–T4.

**Verification:** merge observado; `gh issue view 60` cerrado; roles instalados con `hive-write-docs` y sin `hive-write-spec`.

## Verificación y revisión humana

| Compuerta | Momento | Mecanismo | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final | `go vet ./...`, `go test ./...` | Implementación |
| Sitio | Candidato final | `pnpm --dir site build` | Implementación |
| Revisión de código | PR abierta | `/code-review` de Claude Code; hallazgos corregidos o refutados | D3-A |

Sin efecto de UI renderizada.

## Estado de revisión y progreso

- 2026-10-16: rama creada desde `dae2354e`; T1–T4 en el hilo principal: 516fc663 (T1, T3), e8d1a9dd (T2), 3dc79121 (T4). `go vet` limpio y `go test ./...` con 18 paquetes en verde; `pnpm build` del sitio completo. PR [#62](https://github.com/JhonHawk/tricell-hive/pull/62).
- 2026-10-16: `/code-review` ronda 1: 10 hallazgos de redacción, todos aplicados en 216d8d5a (sección movida tras el procedimiento de preparación, disparador ampliado a planeación y reanudación, regla general de reapertura en vez del enlace a UI, una re-revisión por revisión que cumpla el criterio y con la compuerta del candidato aceptado durante una pausa, criterio solo en `plan-review.md`, excepción de costo en `plan-format.md`, `hive-write-docs` toma hechos solo de sus fuentes nombradas, decisión del usuario anotada en `agent-delivery.md`, entrada Removed en el changelog). Ronda 2: 6 hallazgos, aplicados en d2593557 (disparador de reapertura por criterio, decisión, contrato o delta y marcador `[/]` o `[?]`; tope de una ronda antes de la preparación y una por revisión posterior que cumpla el criterio; "above" en `verification.md`; cualquier revisión del registro va a la sección; `hive-write-docs` lee formato y ajustes; cuatro cambios, no cinco). Sin tercera ronda: cubierto por `go test ./tests/content/... ./integrations/agents/` en verde y las pruebas de `tooling` tras la ronda 1. Costo de las dos rondas de `/code-review`: 130k y 118k tokens de subagente, 3,7 y 3,6 min (reportados por el host).
