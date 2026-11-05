# Tasks

Base: `development` en `47362771`, rama `fix/ux-review-loop`. Las tres tareas corren en el hilo principal: son cambios de redacción en 7 archivos que se leen juntos, y un hijo tardaría más en recibir el contexto que en escribirlos. T1 y T2 son independientes; T3 describe el texto final de ambas.

Cada comprobación `rg` se corre dos veces: sobre `git show 47362771:<path>`, donde debe estar ausente, y sobre el archivo después, donde debe estar presente.

### T1 — Rol `hive-review-ux`: estado no cubierto, seguimiento y veredicto

- [x] El paso 2 emite el veredicto solo sobre lo introducido, con la condición de «aprobado» y el caso sin estados revisados, y lleva los defectos previos a una lista aparte que no cambia el veredicto. El paso 3 trata el estado inalcanzable como «no cubierto» y el seguimiento como una revisión acotada de un solo mensaje. El paso 4 reporta el veredicto, los hallazgos introducidos, la lista de previos y la de no cubiertos.

**Closes:** AC2, AC5, AC6

**Depends on:** ninguna

**Locations:** `content/agents/review/hive-review-ux.md`

**Changes:** Pasos 2–4. Se conservan el escepticismo, «keeps its severity», las regiones, la auditoría y el bloque de límites, que fija `tests/content/agent_boundary_test.go`. Desaparece «Report every defect in those regions, whether the change introduced it or it was already there» y en su lugar entra la lista aparte.

**Verification:**
- AC2: `rg -n 'not covered'`.
- AC5: `rg -n 'follow-up'`.
- AC6: `rg -n 'passed when'` y `rg -n 'separate list'`. Además, `rg -n 'starts as not passed'` y `rg -n 'keeps its severity'` siguen presentes, y `rg -n 'whether the change introduced it or it was already there'` queda ausente.
- `go test ./tests/content/...` pasa.

### T2 — Brief del revisor UX, relectura en `flow-verify` y no cubiertos en el despliegue

- [x] El «Reviewer brief» pide los medios para alcanzar cada estado. `flow-verify` paso 3:
  - relee una corrección de un hallazgo UX con un seguimiento al mismo hijo, o con un hijo nuevo, y define la condición del hijo nuevo por capacidad;
  - da al hijo nuevo los hallazgos corregidos;
  - cubre o reporta cada «no cubierto», también cuando la entrega no se detiene.

  `deployment-checks.md` reporta los no cubiertos del recorrido de despliegue como no verificados.

**Closes:** AC1, AC3, AC4

**Depends on:** ninguna

**Locations:** `content/skills/flow-build/references/delegation.md`, `content/skills/flow-verify/SKILL.md`, `content/skills/flow-deploy/references/deployment-checks.md`

**Changes:** Una frase en «Reviewer brief», dos en `flow-verify` pasos 2–3 y una en `deployment-checks.md`. No cambian la revisión de código, su segunda vuelta, ni `ui-verification.md`.

**Verification:**
- AC1, en `delegation.md`: `rg -n 'how to reach each state'`.
- AC3, en `flow-verify/SKILL.md` y `deployment-checks.md`: `rg -n 'not covered'`; en `flow-verify`, `rg -n 'final report'` dentro de la misma frase.
- AC4, en `flow-verify/SKILL.md`: `rg -n 'same child'` y `rg -n 'no tool to message'`.
- `go test ./tests/content/...` pasa, incluido el límite de 16 000 bytes por skill y por referencia.

### T3 — Página del sitio, delta de especificación y CHANGELOG

- [x] `verify.md` describe los tres comportamientos. El delta de `workflow-lifecycle` añade el requisito de la revisión UX en una pasada. `CHANGELOG.md` añade su entrada bajo `[Unreleased]`.

**Closes:** AC7

**Depends on:** T1 y T2 (describe su texto final)

**Locations:**
- `site/src/content/docs/flows/verify.md`
- `_support/openspec/changes/ux-review-loop/specs/workflow-lifecycle/spec.md`
- `CHANGELOG.md`

**Changes:** Las viñetas de «What it does» y «What it asks you». La página enlaza a la skill y no copia sus reglas. La entrada del CHANGELOG es de nivel patch: cambia cómo los agentes llevan a cabo un flujo existente.

**Verification:**
- La página leída contra T1 y T2.
- Desde `site/`: `pnpm check`, `pnpm lint` y `pnpm build` pasan, como fija `CONTRIBUTING.md:114-121`.
- `go test ./tests/content/...`, que incluye `spec_deltas_test.go`, pasa.

### T4 — Reparar `TestSpecDeltasApplied`, que ya falla en la base

- [x] El test vuelve a pasar en `development`. Hoy falla desde `bb61f71f` porque dos archivos del 2026-11-04 (`viewer-session-lifecycle-and-autoarchive` y `viewer-local-actions-cli`) modifican «Local viewer management actor». `sort.Strings` (`spec_deltas_test.go:113`) toma como último el que se cerró antes.

**Closes:** ninguno: es un defecto que bloquea la verificación de este trabajo.

**Depends on:** ninguna

**Locations:** `tests/content/spec_deltas_test.go`

**Execution:** hilo principal

**Test approach:** tdd. Primero un caso sintético con dos archivos del mismo día que modifican el mismo requisito; falla igual que hoy.

**Changes:** Cuando varios archivos de la misma fecha modifican un requisito, la especificación debe coincidir con alguno de ellos, porque el nombre no fija su orden. No se renombran carpetas archivadas y no se usa la historia de git, que el checkout de CI no garantiza.

**Verification:** El caso nuevo falla antes y pasa después. `go test ./tests/content/...` pasa.

## Verification and human review

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de contenido | Al cerrar cada tarea, local | `go test ./tests/content/...` | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go test ./...`; en `site/`, `pnpm check`, `pnpm lint` y `pnpm build` | Cubierto por la implementación |
| Revisión del comportamiento entero | Antes del push | Lista de lugares revisados y tamaño de `content/guidance/global.md` (sin cambio) en el PR, como pide `AGENTS.md` | Cubierto por la implementación |
| Code review | Candidato final, antes del push | Claude Code `/code-review` (D5 A) | D5 A |

La entrega es automática (D4 A): no hay parada de validación humana. El mensaje de commit nombra las sesiones que motivan cada corrección (host, ID, fecha) y la falla observada, como pide `AGENTS.md` «Measurement».

## Review status and progress

2026-11-04 — Plan escrito y revisado.
- T1 hecho en `bd1acc25`: las comprobaciones antes y después pasan.
- T2 hecho en `5b954f61`: las comprobaciones antes y después pasan; tamaños de 8194, 4842 y 5797 B.
- T4 hecho en `cd6e954a`: el caso del mismo día falló antes del arreglo y pasa después; `go test ./tests/content/...` pasa.
- T3 hecho en el commit del sitio y el CHANGELOG: `pnpm check` (0 errores), `pnpm lint` y `pnpm build` (17 páginas) pasan.

## Checkpoint

- Etapa: verify, paso 3.
- Primera revisión `/code-review` (high) sobre `b285fbf0`: 10 hallazgos.
  - Corregidos (8):
    - destino de los defectos previos sin parada de validación;
    - el revisor nuevo de una relectura queda acotado;
    - puntero desactualizado en `deployment-checks.md`;
    - bloques de una misma carpeta tratados como del mismo día (regresión `TestSpecDeltaSameFolderLastBlockWins`, que falla en `cd6e954a`);
    - caso REMOVED del mismo día sin test;
    - candidatos unificados en una sola lista;
    - «not checked» por «not verified» en el delta;
    - casillas de tareas.
  - No se corrigen (2): la tolerancia a deltas del mismo día debilita el chequeo para esos pares, y la causa de fondo es que los nombres de las carpetas archivadas no dan el orden dentro de un día. Era la decisión menor del plan; se reporta en el cierre.
- Rama: `fix/ux-review-loop` en el worktree `.claude/worktrees/ux-review-loop`.
- Visor: work `57e0a35a…`, con la etapa build en curso.
- Siguiente: segunda revisión acotada a las correcciones, push, PR, CI y merge (D4 A) y, después, `flow-close`.
