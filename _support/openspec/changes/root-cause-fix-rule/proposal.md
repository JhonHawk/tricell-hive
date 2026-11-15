# root-cause-fix-rule — Arreglar un defecto en su causa, con el cambio más pequeño ahí

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | sin ticket · relacionado: #475 (margen de tamaño en otras tres skills) |
| Git | JhonHawk/tricell-hive · base `development` en `6746ceee` · rama propuesta `feat/root-cause-fix-rule` |

## Decisions

2026-11-15
- **A (research)** — ¿Adoptamos una regla de «arreglar en la causa» tras el research adversarial? Sí, como requisito nuevo: se mantiene «smallest sufficient» y se añade que lo más pequeño se mide en la causa, no en el síntoma.
- **D1 A** — ¿Dónde vive la regla? En una referencia nueva de `flow-build`, con un puntero desde `SKILL.md`, que no tiene espacio para la regla completa.
- **D2 A** — ¿Qué más sale de `flow-build/SKILL.md` a referencias? Las secciones Stored data y Security boundaries, que solo aplican en casos concretos; Tests se queda porque se usa en casi cada tarea.
- **D3 A** — ¿Ticket para el margen de `flow-figma-design`, `git-workflow` y `flow-close`? Sí: #475, fuera de este cambio.
- **D6 A** — ¿Cómo se recuperan los bytes que faltan? Se quita la sección «Delegation» de `flow-build/SKILL.md`, que solo repetía el puntero del paso 2 y de la guía global; Checkpoint, reparación del entorno y la primera vista de UI se quedan.
- **D4 C** — ¿Cómo se entrega? Listo para revisión: PR abierto con CI en verde, sin fusionar; lo fusiona el usuario.
- **D5 A** — ¿Qué revisión de código? La nativa de Claude Code, `/code-review`, antes del push.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

Petición del usuario (2026-11-15), a partir de un post público de @poteto («quick fixes considered harmful»): investigar si Hive empuja a los agentes hacia parches de síntoma; tras el research adversarial eligió «adoptarlo como requisito nuevo y preparar el cambio a la guía»: arreglar en la causa que muestra la reproducción con el cambio más pequeño ahí, no contar como arreglo un guard en el síntoma, nombrar como parche lo que no pueda arreglarse en la causa dentro del alcance, no copiar un patrón que es un parche y que la revisión de código lo marque. Durante el plan añadió: «hay que validar si no hay más cosas que deben irse a referencias».

## Objective

Hoy, fuera de la investigación de un fallo reportado, ninguna regla de Hive pide nombrar la causa antes de arreglar, ni distingue un arreglo en la causa de un guard en el síntoma, y la regla de reutilización manda copiar los patrones del código circundante aunque sean parches. Al terminar, todo arreglo de un defecto (en el build, en una ronda de arreglos o en un hijo) nombra su causa y la corrige ahí con el cambio más pequeño; un parche inevitable queda marcado y reportado con su causa; la reutilización no copia parches; la revisión de código los marca; y `flow-build/SKILL.md` recupera margen bajo el límite de 16.000 bytes.

## Scope and acceptance

Dentro: una referencia nueva `flow-build/references/defect-fix.md`, sus punteros desde `flow-build`, `flow-verify` y `flow-deploy/references/urgent-fixes.md`, la excepción en la regla de reutilización, el ángulo nuevo de `hive-review-code`, mover Stored data y Security boundaries a referencias con sus enlaces entrantes y la sección «Delegation» (D6 A), las páginas Build y Verify del sitio, el índice de recursos de `_support/docs/architecture/instruction-resources.md` y la entrada de `CHANGELOG.md`.

Fuera: `content/guidance/global.md` (no cambia; la regla no necesita actuar sin leer su casa), la sección Tests, las otras skills cerca del límite (#475), una prueba que fije la redacción, y cualquier piloto de comportamiento (en pausa).

- AC1 — `content/skills/flow-build/references/defect-fix.md` exige nombrar la causa antes de editar, arreglarla ahí con el cambio más pequeño y buscar la misma falla en los llamadores, arreglándola en los módulos que el trabajo ya toca y reportando las demás; no cuenta como arreglo un guard, retry, fallback, valor por defecto, clamp, error tragado o estado descartado donde aflora el síntoma, salvo que ese estado sea legítimo; define que la causa está fuera del alcance cuando vive en un módulo que el trabajo no toca, en otro repositorio o exige cambiar un contrato o datos guardados, y entonces, o mientras un incidente necesita mitigarse, pide el arreglo mínimo dentro del alcance, marcado en el código con una línea que nombre la causa, nombrado como parche en el informe y con la causa propuesta como ticket; y limita a un hijo a las rutas de su brief · false today because `rg -i symptom content/skills` no devuelve nada y el archivo no existe · Observed in: `content/` en el commit candidato, lo que `hive update` despliega
- AC2 — Quien arregla un defecto llega a esa referencia: `flow-build/SKILL.md` la enlaza con su condición de lectura en el paso 2 y en «Reuse before writing» (que enlazan los roles `hive-build-*`), `flow-verify/SKILL.md` la enlaza en su ronda de arreglos como excepción a «Fix only what each finding names», y `flow-deploy/references/urgent-fixes.md` la enlaza para la mitigación de un incidente · false today because ninguno de esos puntos menciona la causa de un defecto · Observed in: `content/` en el commit candidato
- AC3 — El peldaño 2 de «Reuse before writing» no copia un patrón que es un parche (marcado por un comentario o un guard para un estado que no debería existir) y pide nombrarlo en el informe · false today because el peldaño dice solo «use it as the surrounding code does» · Observed in: `content/skills/flow-build/SKILL.md` en el commit candidato
- AC4 — Las secciones Stored data y Security boundaries viven en `flow-build/references/stored-data.md` y `security-boundaries.md`, `SKILL.md` las enlaza con su condición, ningún enlace de `content/` apunta a las anclas viejas `#stored-data` o `#security-boundaries`, y `flow-build/SKILL.md` queda con al menos 1.000 bytes libres bajo 16.000 · false today because quedan 229 bytes libres y 7 enlaces usan esas anclas · Observed in: `content/` y `go test ./tests/content/` en el commit candidato
- AC5 — `hive-review-code` revisa como ángulo un arreglo puesto donde aflora el síntoma en vez de en su causa y el código nuevo que copia un parche · false today because su paso 2 no menciona síntoma ni parche · Observed in: `content/agents/review/hive-review-code.md` en el commit candidato
- AC6 — La página Build del sitio dice que un bug se arregla donde empieza y que un parche inevitable se marca y se reporta, con enlace a la regla · false today because `site/src/content/docs/flows/build.md` no lo menciona · Observed in: `site/` en el commit candidato

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `content/skills/flow-build/SKILL.md` mide 15.771 bytes; `tests/content/skill_size_test.go` fija 16.000 para `SKILL.md` y para cada referencia (medido el 2026-11-15 en `6746ceee`).
- Secciones de `flow-build/SKILL.md`: Reuse 932 B, Tests 5.004 B, Delegation 250 B, Stored data 714 B, Security boundaries 549 B.
- Enlaces entrantes a `#stored-data`: `hive-design-architecture.md:16`, `hive-build-data.md:10`, `hive-build-infra.md:15`, `flow-deploy/SKILL.md:14`, `flow-deploy/references/deployment-checks.md:7`; a `#security-boundaries`: `hive-review-code.md:20`, `hive-review-security.md:11`. El sitio no enlaza esas anclas.
- Los roles `hive-build-backend`, `-frontend`, `-infra`, `-kmp` y `-data` enlazan `#reuse-before-writing`; `flow-verify/SKILL.md` paso 3 hace la ronda de arreglos.
- La referencia `flow-research/references/diagnosis.md`, con el paso «identify temporary measures that still need retirement», se eliminó en `7a37381b` (2026-10-30); este cambio cubre esa intención con el parche marcado y reportado.
- Research del 2026-11-15 (fuentes en `engram:tricell-hive/decision/root-cause-fix-rule`): no hay evidencia controlada de que pedir el cambio mínimo empeore el código; la sobreedición está medida, así que la regla de mínimo se mantiene; las fuentes de referencia convergen en «smallest root-cause fix in scope; otherwise smallest in-scope fix, report the rest open».

## Effects

None: the tasks take no effect on an external system.

## Delivery

`ready` (D4 C): revisión, push de `feat/root-cause-fix-rule`, PR no borrador hacia `development`, espera del CI y respuesta a sus revisiones; se detiene con el PR abierto para que el usuario lo fusione. No autoriza fusionar, auto-merge ni cola de merge, ni refrescar la instalación local antes de que el usuario fusione. Revisión: nativa, Claude Code `/code-review` (D5 A), sobre el candidato antes del push.

## Review status and progress

2026-11-15 — plan escrito. Revisión del plan (`hive-review-plan`): 1 P1, 8 P2, 3 P3, todos verificados. Aplicados: alcance definido y excepción a «Fix only what each finding names» (H3), límite de los hijos (H4), escenario de fallo de la revisión (H5), changelog (H6), mitigación de incidentes en `urgent-fixes.md` (H7), «tarea o enfoque» para el plan (H8), páginas Build y Verify (H9), comandos de verificación reales (H10), orden con Tests «When a check fails» (H11), marca sin ticket y sin cita inexistente (H12), comentario de `voice.go`. H1 y H2 resueltos con D6 A. Estado: listo para flow-build.
