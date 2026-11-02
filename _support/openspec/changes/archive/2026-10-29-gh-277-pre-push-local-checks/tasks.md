# Tareas

Base del cambio: `development` 0afaa7c6, confirmada al iniciar el build. Rama `feat/gh-277-pre-push-local-checks` en `.claude/worktrees/gh-277-pre-push-local-checks`.

## T1 — La regla en review-and-gates y sus punteros en flow-verify y git-workflow

- [x] La regla "Local checks before publication" existe, y flow-verify y git-workflow la aplican antes de un push.

**Closes:** AC1, AC2, AC3, AC4, AC5

**Checkpoint:** C1

**Depends on:** none

**Locations:**
- `content/skills/flow-verify/references/review-and-gates.md` (nueva sección, `:10`, `:40`)
- `content/skills/flow-verify/SKILL.md:14,25,40,47`
- `content/skills/git-workflow/SKILL.md`, sección "Prepare a coherent delivery"

**Execution:** main thread. Es una sola regla con tres punteros acoplados.

**Test approach:** check. Se verifica con las búsquedas de abajo y `go test ./tests/content/... ./tooling/management/ -run 'Anchor|Size|Link'`.

**Changes:** aplicar "Dónde vive la regla", "Texto de la regla" y "Otros cambios" de [design.md](design.md).

**Verification:**
- `rg -n "## Local checks before publication" content/skills/flow-verify/references/review-and-gates.md` devuelve una línea.
- `rg -n "depends on one of them" content/skills/flow-verify/references/review-and-gates.md` devuelve una línea (AC1).
- `rg -n "workflow files" content/skills/flow-verify/references/review-and-gates.md` devuelve una línea (AC2).
- `rg -n "CI-only" content/skills/flow-verify/references/review-and-gates.md` devuelve una línea (AC3).
- `rg -n -F "Close regression coverage through the project's actual CI gate" content` no devuelve nada (AC4).
- `rg -n "local-checks-before-publication" content/skills/git-workflow/SKILL.md content/skills/flow-verify/SKILL.md` devuelve al menos dos líneas (AC5).
- `rg -n "did not skip verify" content/skills/flow-verify/references/review-and-gates.md` devuelve una línea: el salto que pide el usuario se respeta.
- `wc -c` de los tres archivos da 16,000 bytes o menos cada uno.
- `go test ./tests/content/... ./tooling/management/` pasa, incluida la prueba de anclas `TestRepositoryInstructionAnchorsResolve`.

## T2 — Sitio y CHANGELOG

- [x] La página de Verify y el CHANGELOG describen los chequeos locales antes del push.

**Closes:** AC6

**Checkpoint:** C1

**Depends on:** T1

**Locations:** `site/src/content/docs/flows/verify.md:24-25,28`, `site/src/content/docs/flows/build.md:52`, `CHANGELOG.md`

**Execution:** main thread.

**Test approach:** check. Se verifica con una búsqueda y el build del sitio.

**Changes:** describir el comportamiento sin copiar la regla: antes de cada push, el agente corre localmente las suites de los paquetes cambiados y de sus dependientes, y los jobs del CI que no necesitan servicios. La integración, el e2e y los despliegues quedan en el CI y el reporte los nombra.

**Verification:**
- `rg -n -i "before (each|a) push" site/src/content/docs/flows/verify.md` devuelve una línea.
- `astro build` en `site/` termina sin errores. Para correrlo se enlaza `node_modules` desde el checkout principal, como en #272.

## Verificación y revisión

C1 — T1 y T2 — es el único punto de control. T1 no tiene dos tareas posteriores que dependan de ella y no cambia un contrato persistido, así que T1 y T2 se verifican juntas.

| Gate | Momento y entorno | Mecanismo y evidencia | Autoridad |
| --- | --- | --- | --- |
| Suite completa | Candidato final, local | `go test ./...` pasa | Cubierto por la implementación |
| Revisión dedicada | Candidato local, antes del push | `/code-review` de Claude Code; cada hallazgo se resuelve o se refuta | Respuesta reutilizada de esta sesión (D7-A) |
| Prueba en vivo | — | No aplica: es contenido de guía. El efecto real se observará en sesiones de cam con la guía nueva | — |

## Estado de la revisión y progreso

**Revisión del plan.** Un revisor `hive-review-plan` (consistencia de la guía) revisó la primera versión: unos 144 s y 128 k tokens. Dio 3 hallazgos bloqueantes y 3 no bloqueantes, todos aplicados con la corrección que él mismo propuso, sin otra ronda:

- **B1.** La regla respeta un salto de verify pedido por el usuario. Es la opción recomendada por el revisor y concuerda con `flow-verify:31` e `iterate-and-integrate:13`.
- **B2.** "Stay in CI" se limita a los jobs con servicios que este push dispara, más allá de lo que verify ya selecciona. Una suite que corre solo después del merge sigue la regla de `verification.md`.
- **B3.** Una suite completa sobre el mismo candidato cuenta, y el paso 2 de flow-verify corre estos chequeos.
- **N4.** Un push que solo trae commits ya integrados queda exento.
- **N5.** Un chequeo que no puede correr aquí no frena el push si el CI lo corre.
- **N6.** T2 corrige también `verify.md:28` y `build.md:52`.

**Progreso.**

- T1 (cf61d055) y T2 (84dfc139) están `[x]`. `hive-verify-task` dio AC1–AC5 cumplidos (unos 98 s, 90 k tokens) y AC6 cumplido (unos 62 s, 85 k tokens). El modelo del verificador está configurado, no observado; el implementador es Opus 5.5, observado.
- El verificador de T2 notó que la viñeta de `verify.md:26` omitía "deterministic" y "secret".
- `/code-review` (high) sobre `development...84dfc139` dio 9 hallazgos, corregidos en 95226071. T1 y T2 vuelven a `[?]` para verificarse de nuevo, y los arreglos pasan por una revisión acotada a ellos. Los cambios:
  - los chequeos se repiten antes de cada push cuando cambiaron sus paquetes;
  - cubre cualquier repositorio con CI, también el que solo corre después del merge;
  - el paso 2 los corre solo cuando sigue un push;
  - cuenta lo que el hook de pre-push ya corre;
  - un chequeo que no puede correr sigue la regla de hooks;
  - la selección se hace contra la base del merge;
  - los punteros ya no repiten parte de la condición;
  - el sitio y el CHANGELOG quedan alineados.
- La nueva verificación sobre 95226071 dio AC1–AC6 cumplidos, y T1 y T2 volvieron a `[x]`.
- La revisión acotada a los arreglos (`84dfc139..95226071`) dio 10 hallazgos. Casi todos venían de tres arreglos que ampliaron la regla más allá del contrato del plan: cualquier repositorio con CI, los jobs posteriores al merge y la regla de hooks para un chequeo que no puede correr. En 9deb3a4c, sin otra revisión:
  - se restauró el contrato de AC1, AC2 y N5;
  - se quitó el crédito al hook de pre-push, que corre después de la revisión y de la validación;
  - quedó un solo orden: después de correcciones, los chequeos se repiten antes del push y antes de la revisión que las cubre;
  - el sitio y el CHANGELOG quedaron alineados.
- Dos hallazgos de la primera ronda quedan sin adoptar, con su razón:
  - El repositorio cuyo CI solo corre después del merge no queda cubierto: ampliar la regla rompía AC1 y obligaba a una pregunta antes del push.
  - El typecheck se repite con el hook de cam: cuesta poco y evita depender del momento en que corre el hook.
- T1 y T2 se verificaron otra vez sobre 9deb3a4c: AC1–AC6 cumplidos. Vuelven a `[x]`.
- `go test ./...` pasa sobre 9deb3a4c. Integrado por el PR #278 (merge abc5e2bf).
