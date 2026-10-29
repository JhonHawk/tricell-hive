# Tareas

Base: f3d15886 (`origin/development`, 2026-10-29). Worktree `.claude/worktrees/review-loop-bounds`, rama `fix/review-loop-bounds`. La copia canónica de este registro es la del worktree; la del checkout principal, creada en esta sesión, se elimina al cierre.

## Tasks

### T1 — Regla de la revisión dedicada acotada

- [x] `review-and-gates.md` aplica las cuatro piezas de `design.md`.

**Closes:** AC1, AC2, AC3, AC4, AC5.

**Checkpoint:** C1.

**Depends on:** none.

**Locations:** `content/skills/flow-verify/references/review-and-gates.md` (sección «Dedicated code review», líneas 25-36).

**Execution:** hilo principal; es una redacción acoplada y más corta que su propio brief.

**Test approach:** check: `go test ./tests/content/...` (tamaño, enlaces y anclas) más la lectura de cada criterio contra el texto final.

**Changes:** reescribir el párrafo inicial y las viñetas Severity, Re-review y Later edits como fija el diseño; no tocar otras secciones.

**Verification:** `go test ./tests/content/...` pasa; `wc -c` del archivo ≤ 16 000; `rg -n '/code-review|Claude Code' content/skills/flow-verify/references/review-and-gates.md` no devuelve nada (AC1); al leer el texto final aparecen el registro de severidad y el arreglo que dispara la segunda revisión (AC2), la limitación de la segunda ronda a P0/P1 (AC3), la ausencia de revisión después de ella (AC4) y la identidad de candidato fuera de `hold` (AC5).

### T2 — Coherencia de los textos dependientes

- [x] Ningún texto de guía, sitio o documentación de arquitectura contradice la regla nueva.

**Closes:** AC1, AC3, AC4, AC5.

**Checkpoint:** C1.

**Depends on:** T1.

**Locations:** `review-and-gates.md:44-46`, `content/skills/flow-research/references/pr-review.md:15`, `content/skills/flow-build/references/verification.md:40`, `site/src/content/docs/flows/verify.md:30`, `site/src/content/docs/flows/build.md:47` (requiere edición: dice que tras la revisión se revisa lo cambiado desde el commit revisado, sin el tope de D4-A ni la identidad de candidato), `_support/docs/architecture/`.

**Execution:** hilo principal; búsqueda acotada sobre la redacción que acaba de escribir.

**Test approach:** check: búsqueda de los términos del cambio en todo el corpus.

**Changes:** editar solo lo que contradiga la regla; registrar en el PR los lugares revisados.

**Verification:** `rg -n -i 'verif(y|ies) (each|its own) finding|second pass|re-review|at most once|later edits|reviewed commit|candidate identity|same limits|what changed since' content site _support/docs/architecture tests/pilot` solo devuelve líneas coherentes con AC1, AC3, AC4 y AC5; `site/src/content/docs/flows/build.md:47` describe el tope de D4-A; las líneas que se dejen sin cambio quedan listadas en el PR con su razón.

## Verification and human review

- C1 — T1, T2 — un único resultado de redacción; T2 depende de T1.

| Puerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Tests de contenido | Durante el build, local | `go test ./tests/content/...` en verde | Implementación |
| Suite completa | Una vez, sobre el candidato final, local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` en verde | Implementación |
| Revisión dedicada | Candidato local, antes del push | `/code-review` (D6-A); con la regla nueva, sus P0/P1 pasan por `hive-refute-claim` | D6-A |

Sin efecto de UI renderizado: el cambio en el sitio, si lo hay, es solo de texto. No aplica in vivo: es guía distribuida, sin comportamiento que ejecutar; los pilotos siguen pausados (AGENTS.md).

## Entrega

- Repositorio: tricell-hive, base `development`; no hay CI (`CONTRIBUTING.md:15`).
- Modo `automatic` (D5-A, 2026-10-29): rama de trabajo, commits, `/code-review` sobre el candidato local (D6-A), push, PR a `development` y merge cuando pasen los chequeos; después `hive update` local, como pide AGENTS.md para el checkout del mantenedor.
- Registro de cambio: se versiona una vez al cierre, con un PR pequeño sin revisión dedicada que mergea el agente.
- El PR indica el tamaño de `content/guidance/global.md` (sin cambio) y los lugares revisados, como pide AGENTS.md.
- El commit nombra las sesiones observadas (Claude Code 33c2742a y 6495b668 para F3; las 94 bifurcaciones de `/code-review` para F2) y la fecha 2026-10-29.

## Review status and progress

Revisión del plan: revisión 312fc0774ad9, dominio de consistencia de guía, un `hive-review-plan` (72 s, ~99 k tokens). Cuatro bloqueantes y dos no bloqueantes, todos aceptados:

- B1 (búsqueda de T1 imposible de pasar por los enlaces `#dedicated-code-review`): se busca `/code-review|Claude Code`.
- B2 (`site/.../build.md:47` repite la regla vieja de ediciones posteriores): añadido a T2; T2 cierra también AC5 y amplía la búsqueda.
- B3 (la cláusula posterior seguía arreglando P2 introducidos, contra D3-A): limitada a los P0/P1 de la segunda ronda. Decisión tomada sin preguntar: los P2 sin arreglar de esa ronda siguen la regla de hallazgos incidentales, como la viñeta P2 hace hoy con los que no arregla; los P3 se reportan. Razón: es la ruta existente para un P2 que no se arregla. Se revierte usando «reportar» a secas.
- B4 (choque con la revisión de correcciones de validación, línea 51): el texto nuevo la exceptúa sin editarla.
- N1 (commits no revisados en una promoción): el cuerpo del PR registra la edición como no revisada.
- N2 (bytes justos): T1 ya mide con `wc -c`; se compacta dentro de la sección si hace falta.

Todas las correcciones aplican la propuesta del revisor, así que no hubo segunda ronda.

Build (D7-A, 2026-10-29): implementador claude-opus-5-5 (observado); `hive-verify-task` con su perfil configurado, no observado.

- T1 en 86f54cce. Verificación 1: AC1–AC5 met. El verificador señaló que «follow the shared incidental-findings rule» dejaba abierta la vía de arreglo pequeño de `incidental-findings.md:9` para un P2 de la segunda ronda, en contra de D3-A. Corregido en c358bb0c: los P2 y P3 de esa ronda quedan sin arreglar, y los P2 se proponen. Reverificación de AC3: met (40 s). T1 queda en `[x]`.
- T2 en 17ed2cbb: `site/src/content/docs/flows/build.md:47`. El resto de lo que encontró la búsqueda queda coherente sin cambios: `review-and-gates.md:21,44-46,51`, `plan-revision.md:11,13`, `pr-review.md:15`, `verification.md:38,40`, `site/.../verify.md:30` y `ui-gate.md:21-26`.
- Verificación de T2 (87 s): AC1, AC3, AC4 y AC5 met. Señaló que la página no ponía la condición «sin contrato»; corregido en 7eb1a0c0. Reverificación de AC4: met (35 s). T2 queda en `[x]`.
- `/code-review` (high, D6-A) sobre 7eb1a0c0 contra f3d15886 (112 s): 10 hallazgos sin verificar y sin severidad. Los mapeé con la regla nueva:
  - R2 (P1, una edición posterior sin segunda revisión debida no tenía regla): el refutador lo dio por plausible (59 s). Corregido en 25a39194.
  - R1, R4, R7 y R9 (P2 introducidos y locales: «refuted» leído como apto, revisión de seguridad bloqueada, orden severidad → refutación, sin PR no queda registro): corregidos en 25a39194.
  - R3 y R5 (ediciones posteriores sin revisar y P2 de la segunda ronda sin arreglar): descartados porque los cubren D4-A y D3-A.
  - R6 (la nueva definición de revisión verificadora se basa en lo que el revisor dice de sí mismo): P2 que no se puede arreglar en local; queda como límite reportado.
  - R8 («small» frente a «local» en el sitio): P3, se reporta.
  - R10 (34 bytes libres): el archivo queda en 15 968 bytes; se reporta.
  - El cuerpo del commit 86f54cce dice que los P2 «follow» la regla, algo que c358bb0c superó; queda anotado en el PR, sin reescribir la historia.
- Segunda revisión (R2 era un arreglo P1 que cambiaba lógica): `/code-review` sobre 7eb1a0c0..25a39194 (114 s), 10 hallazgos. Mapeo:
  - S1 (los arreglos de la primera ronda cuentan como ediciones posteriores y disparan la segunda revisión) y S3 (se perdió «to the same candidate»): P1, en el refutador.
  - S2 (doble revisión de las correcciones de validación antes de la segunda revisión, ya presente en la base), S4 (dónde se registra una edición sin revisar cuando no hay PR ni registro de cambio), S5 (AC2 menos explícito), S8 (el tope de la revisión de seguridad tras ediciones posteriores) y S10 (32 bytes libres; conviene dividir la referencia): P2, sin arreglar y propuestos.
  - S6 (el alcance del diff): ya refutado en la primera ronda.
  - S7 (los P2 del alcance añadido quedan sin arreglar): descartado porque lo cubren D3-A y D4-A.
  - S9 (enlace y redacción del sitio): P3, se reporta.
  - Refutador (112 s): S1 y S3 son plausibles. Corregidos en 6b968d3d sin otra revisión, como fija la regla; se reportan como arreglados sin revisión. Reverificación de AC4 y AC5 (46 s): met. El archivo queda en 15 991 bytes. Reverificación de T1 tras 25a39194 (69 s): AC1–AC5 met. Nota: AC2 ahora dice «qué inicia una segunda revisión» y ya no «el arreglo P0/P1», porque una edición posterior también puede iniciarla.
- Integración: PR [#286](https://github.com/JhonHawk/tricell-hive/pull/286) mergeado en `development` con merge commit 051471ec el 2026-10-29; el repositorio no tiene CI. `go vet ./...` y `go test ./...` en verde sobre 6b968d3d.
- Suite completa: corrió en segundo plano sobre 17ed2cbb; la corrección c358bb0c solo cambia texto dentro del límite de tamaño, ya cubierto por `go test ./tests/content/...` en verde (15 966 bytes).
