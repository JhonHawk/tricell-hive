# Tareas

Base del cambio: `b8d2b785` (`feat/delegation-threshold-global`), confirmada al empezar el build el 2026-10-16.

## T1: Git que reescribe el árbol (U1)

- [x] El pie de los 10 roles que escriben y el brief de `delegation.md` prohíben los comandos de Git que reescriben el árbol cuando otros escriben en él.

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:**
- `content/agents/development/hive-build-backend.md`, `hive-build-frontend.md`, `hive-build-data.md` y `hive-build-kmp.md`.
- `content/agents/ops/hive-build-infra.md`.
- `content/agents/design/hive-design-architecture.md` y `hive-design-ui.md`.
- `content/agents/docs/hive-write-docs.md`.
- `content/agents/quality/hive-write-tests.md` y `hive-tune-performance.md`.
- `content/skills/flow-research/references/delegation.md:29`.

**Execution:** hilo principal. Son once inserciones de texto idéntico, y un brief para delegarlas sería más largo que el cambio.

**Test approach:** check, `rg`.

**Changes:** insertar las dos frases de [design.md § U1](design.md#u1-git-que-reescribe-el-árbol-d2-a), tal cual.

**Verification:** una comprobación por cada cláusula de AC1 y AC2.
- `for p in 'A Git command that rewrites working-tree files' 'git stash' 'git reset' 'git checkout' 'git restore' 'git clean' 'git switch' 'git show <base>:<path>' 'separate worktree'; do echo "$p: $(rg -l -F "$p" content/agents | wc -l)"; done` → 10 en cada línea.
- `diff <(rg -l -F 'git stash' content/agents | sort) <(rg -l 'Modify or delete only the paths the brief names' content/agents | sort)` → sin diferencias.
- `for p in 'git stash' 'git switch' 'neither does the main thread while a child works in that tree' 'read-only Git commands or a separate worktree'; do echo "$p: $(rg -c -F "$p" content/skills/flow-research/references/delegation.md)"; done` → 1 o más en cada línea. Las coincidencias quedan en la sección "Supply a bounded brief": su número de línea está entre el de `## Supply a bounded brief` y el de `## Deliver and resolve instructions` (`rg -n '^## ' …`).

## T2: una señal de fin que cubra todos los resultados (U2)

- [x] `delegation.md` exige que la señal de fin termine con el éxito, con el fallo y con un objetivo que aún no existe.

**Closes:** AC3.

**Depends on:** ninguna. Toca el mismo archivo que T1 pero en otra sección, y lo ejecuta el mismo hilo en secuencia.

**Locations:** `content/skills/flow-research/references/delegation.md`, sección "Launch and wait", después de la línea 45.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** insertar el bloque **Condition/Actions** de [design.md § U2](design.md#u2-una-señal-de-fin-que-cubra-todos-los-resultados).

**Verification:** una comprobación por cada cláusula de AC3.
- `for p in 'giving an external process a completion signal' 'success, failure, and a target that does not exist yet' 'exit status' 'including indentation' 'waits for the target to appear' 'no checks reported' 'rerun that same command'; do echo "$p: $(rg -n -F "$p" content/skills/flow-research/references/delegation.md | cut -d: -f1)"; done` → un número de línea por cláusula.
- Todos los números caen entre la línea de `## Launch and wait` y la de `## Check returned work` (`rg -n '^## ' …`).
- La línea de 'giving an external process a completion signal' es menor que la de 'a child is expected to outlast one wait stretch'.

## T3: hooks y chequeos omitidos (U3)

- [x] `git-workflow` prohíbe saltar hooks sin aprobación y exige correr o reportar los chequeos que un hook omite. El sitio describe ese comportamiento.

**Closes:** AC4, AC5, AC6.

**Depends on:** ninguna.

**Locations:**
- `content/skills/git-workflow/SKILL.md`, "Prepare a coherent delivery", después de la línea 26, y la frase de excepción en la línea 38 ("Verify the authorized effect").
- `site/src/content/docs/flows/build.md`, "Who commits and merges?" (líneas 63–64).

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** insertar el párrafo y la frase del sitio de [design.md § U3](design.md#u3-hooks-y-chequeos-omitidos-d1-a-d3-a).

**Verification:** una comprobación por cada cláusula de AC4, AC5 y AC6.
- `for p in 'no-verify' 'git commit -n' 'HUSKY=0' 'LEFTHOOK=0' 'core.hooksPath' "explicit approval for that repository" 'for the rest of the session' 'name it when reporting each commit' 'skipping checks' 'read the hook' 'tools already available' 'propose installing it' 'When any check remains not run, ask before pushing' 'runs the same commands the hook runs' 'a job name alone does not suffice' 'report the push as pending' 'Skipping a Git hook follows the hooks paragraph above instead'; do echo "$p: $(rg -n -F "$p" content/skills/git-workflow/SKILL.md | cut -d: -f1)"; done` → un número de línea por cláusula. Todas caen dentro de `## Prepare a coherent delivery` (`rg -n '^## ' …`), salvo 'Skipping a Git hook follows the hooks paragraph above instead', que queda en `## Verify the authorized effect`, en la misma línea que "Use any other bypass".
- `rg -n -i 'hooks' site/src/content/docs/flows/build.md` → al menos una línea entre `### Who commits and merges?` y el encabezado siguiente (`rg -n '^##' …`). La frase dice que el agente salta un hook solo con tu aprobación y que avisa de los chequeos que no se corrieron.

## Verificación y revisión

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Sobre el candidato final, en este checkout | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos. `tests/content` vigila los enlaces de las skills y la huella de lectura. | Incluida en la implementación |
| Tamaño de `global.md` | Candidato final | `git diff --stat b8d2b785 -- content/guidance/global.md` vacío | Incluida |
| Revisión del alcance | Candidato final | Releer las tres secciones editadas y buscar una contradicción con `global.md:79`, `verification.md:87` y `git-workflow` línea 24 | Incluida |
| Revisión de código dedicada | No aplica | Son cambios solo de texto; preferencia guardada para Hive | — |
| Prueba en sesiones reales | Tras el despliegue local que pidas | Tú la ejecutas | Despliegue pendiente de tu pedido |

No hay verificación local en vivo: ninguna sesión usa la guía sin desplegarla, y desplegar necesita tu pedido.

## Revisión y progreso

- 2026-10-16: borrador escrito tras D1-A, D2-A y D3-A.
- 2026-10-16: revisión del plan sobre la revisión `195b57f2cd36` con un solo `hive-review-plan` nativo (dominio: guía distribuida de los agentes). Duró 180 s y usó ~121k tokens.
  - **Bloqueantes:**
    - B1 (U2 invitaba a sondear): aceptado y aplicado con la redacción del revisor.
    - B2 (la verificación no cubría cada cláusula): aceptado; ahora hay una comprobación por cláusula.
    - B3 (U3 en modo desatendido): pasa a ti como D4.
  - **No bloqueantes:**
    - N1 (alcance de la aprobación): pasa a ti como D5.
    - N2 (`git commit -n`): aplicado y comprobado con `git commit -h`.
    - N3 (no instalar herramientas): aplicado.
    - N4 (U1 también impide `git switch -c` dentro del propio worktree de un hijo): descartado, es restrictivo pero seguro.
    - N5 (`hive-verify-change` y `hive-review-ux` ejecutan git sin la cláusula): queda como punto a vigilar con conteo 0; `delegation.md` los alcanza a través del brief.
    - N6 (inventario del sitio): corregido.
    - N7 (tooling y pruebas): sin efecto.
  - Las correcciones aplican la propuesta del propio revisor o solo cambian la verificación, así que no requieren otra ronda. Si D4 o D5 cambian el contrato de U3, esa parte tendrá una re-revisión.
- 2026-10-16: D4 (respuesta propia: empujar solo si el CI corre esos chequeos; si no, pendiente y propuesta de instalar) y D5-A (la aprobación vale toda la sesión en ese repositorio) cambiaron el contrato de U3.
- 2026-10-16: re-revisión, la única ronda, sobre `9b7d403afa93`, reanudando el mismo revisor. Duró 43 s y usó ~131k tokens.
  - **Bloqueante:** R1 (U3 contradecía la línea 38 de `git-workflow`): aplicado como lo propuso el revisor.
  - **Aplicados tal como se propusieron:** N1 (preguntar solo si queda algún chequeo sin correr), N2 (qué cuenta como evidencia de CI), N4 (frase del sitio) y N5 (`git checkout` en la verificación de T1).
  - **Descartado:** N3, porque D4 contestó la pregunta tal como se formuló, con el modo automático incluido.
  - **Verificado por el revisor:** los patrones coinciden con el texto propuesto y ningún AC es cierto en la base.
- **Cobertura:** el dominio de la guía distribuida quedó revisado en `195b57f2cd36` y re-revisado en `9b7d403afa93`. Los cambios posteriores aplican las propuestas del revisor sin alterar su contenido.
- **Límites:** el revisor no ejecutó comandos ni la suite; no se consultó documentación sobre `HUSKY=0`/`LEFTHOOK=0`.
- 2026-10-16, build: T1, T3 y T2 implementadas en el hilo principal (claude-opus-5-5, observado). Cada una pasó a `[x]` con veredicto `met` de `hive-verify-task`, cuyo modelo es `opus` con esfuerzo `high` según `~/.claude/agents/hive-verify-task.md` (configurado, no observado):
  - T1, AC1 y AC2: 50 s.
  - T3, AC4, AC5 y AC6: 44 s.
  - T2, AC3: 35 s.
- Tras el veredicto de T1 se aplicó la sugerencia de su verificador en `delegation.md:29` ("It may commit" → "The child may commit"). Solo cambia la redacción; las comprobaciones de AC2 siguen pasando.
- La frase de excepción de T3 quedó en la línea 40 de `git-workflow`, no en la 38, porque el párrafo nuevo desplazó las líneas.
- **Compuertas del candidato final:**
  - `go vet ./...` sin hallazgos.
  - `go test ./...`: todos los paquetes `ok`.
  - Tests de Python: 23 OK.
  - `global.md` sin diferencias frente a `b8d2b785`.
  - Revisión del alcance: el texto nuevo es coherente con `global.md:79`, `verification.md:87` (al que enlaza) y `git-workflow` línea 24.
- 2026-10-16, C1-A (commit y despliegue local):
  - **Commits en `feat/delegation-threshold-global`:** `682591fd` (T1), `7e811e46` (T2) y `32e2a709` (T3), sin push.
  - **Despliegue:** `hive update --out` + `hive apply --plan` desde `32e2a709`, transacción `6b18c550…`.
  - **Comprobado:** el texto nuevo está en `~/.claude`, `~/.agents`, `~/.codex`, `~/.config/opencode`, `~/.grok`, `~/.pi/agent` y `~/.cursor`.
- **Estado:** desplegado en local para tus pruebas en sesiones reales. Siguiente paso: tú lo pruebas y decides el merge a `development`.
