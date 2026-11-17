# Tasks

Base `development` en `8ac8d77c`; rama `feat/pr-before-after-captures`. Todas las tareas corren en el hilo principal: son ediciones cortas de texto que comparten una sola regla, y repartirlas entre hijos duplicaría la lectura de esa regla.

### T1 — El build captura el «antes»

- [x] `ui-verification.md` tiene la sección de capturas «antes» y `flow-build/SKILL.md` la manda leer antes de la primera edición, excluye esas capturas del visor y guarda sus rutas en el checkpoint

**Closes:** AC8

**Depends on:** none

**Locations:** `content/skills/flow-verify/references/ui-verification.md`; `content/skills/flow-build/SKILL.md` (línea 10 y párrafo del checkpoint)

**Execution:** main thread

**Test approach:** check

**Changes:** Sección nueva en `ui-verification.md`: antes de la primera edición de cualquier cosa que una pantalla existente muestra (componente, estilo, diseño, contrato de datos o datos sembrados), el hilo principal sirve el checkout sin cambios con los datos que tenga, y captura cada región que el plan cambia, o que la petición cambia sin plan, en los anchos, temas y estados que verify revisará, en `workspace/<fecha>-<trabajo>/images/before/`, con un nombre que la empareja con la captura del candidato; si los datos difieren de los que usará verify, la tabla del PR lo dice; lo hace antes de lanzar una tarea hija que edite esa superficie, y el brief del hijo no le pide capturar; una pantalla nueva no tiene «antes» y un «antes» perdido por una edición previa no se reconstruye. En `flow-build/SKILL.md` línea 10: leer UI verification antes de la primera edición, y las capturas «antes» no se publican en el visor (D6 A); en el párrafo del checkpoint: las rutas del «antes», o que no hay. No cambia cómo verify captura el candidato ni el visor. Lugares revisados que no cambian: `delegation.md` (el brief ya no pide capturar), `e2e-coordination.md` (el coordinador es el hilo principal), `hive-build-frontend` y `hive-verify-change`, `unattended-delegation`.

**Verification:** `rg -n "before its first edit" content/skills/flow-build/SKILL.md`; la sección nueva en `ui-verification.md`; `wc -c content/skills/flow-build/SKILL.md` ≤ 16000; `go test ./tests/content/...` pasa.

### T2 — El PR lleva la tabla antes | después

- [x] Referencia nueva `git-workflow/references/pr-captures.md` y un puntero con su condición en el paso 6 de `git-workflow/SKILL.md`, que queda en 16 000 bytes o menos

**Closes:** AC9, AC10, AC11

**Depends on:** none

**Locations:** `content/skills/git-workflow/references/pr-captures.md` (nuevo, primera carpeta `references/` de la skill); `content/skills/git-workflow/SKILL.md` pasos 6 y 9

**Execution:** main thread

**Test approach:** check

**Changes:** La referencia fija: cuándo aplica (el walk de verify produjo capturas); la detección (`gh pr create --help` lista `--attach`; el host del remoto es github.com o GitHub Enterprise Server; `gh repo view --json visibility`); la tabla por región, ancho, tema y estado, con «pantalla nueva» o «sin antes» cuando falta un lado y la nota cuando los datos del «antes» difieren; la subida con `--attach` y rutas absolutas o un `--body-file` junto a las imágenes; que ninguna imagen entra al commit; tras una subida parcial, completar con `gh pr edit --attach`; toda edición posterior del cuerpo, incluido el resultado de una revisión que añade el paso 6, parte de `gh pr view --json body`; reemplazar una fila tras una corrección con `--body-file` y `--attach`; en un repositorio público o en GitHub Enterprise Server, solo datos sembrados y cuentas de prueba, y solo con la autorización única del trabajo (D7 A), preguntada con la entrega o en la ronda de efectos del plan; el respaldo de texto en cualquier otro caso; y la comprobación tras publicar (`gh pr view --json body` muestra una URL de GitHub por imagen subida). En `SKILL.md` paso 6, una frase que remite a la referencia cuando el walk produjo capturas; para caber, se quita del paso 9 la frase «Step 7 already reads reviews and comments through REST.», que solo describe lo que el paso 7 ya dice. Lugares revisados que no cambian: `global.md` «Delivery» y «Scope and authorization» (la subida pública queda como efecto que la referencia pregunta una vez, no como nuevo modo), `flow-verify/SKILL.md` (la parada de validación no cambia), `viewer-publication.md`, los agentes (ninguno compone el cuerpo del PR), `tests/pilot/`.

**Verification:** `wc -c content/skills/git-workflow/SKILL.md` ≤ 16000; comprobar que el paso 7 lee revisiones y comentarios por REST antes de quitar la frase; `go test ./tests/content/...` pasa (tamaños y referencias resueltas).

### T3 — El cierre conserva lo que el PR no guarda

- [x] `artifact-layout.md` («Retain evidence» y el párrafo de imágenes de revisión de UI) y `support-material.md` «Delete» tratan como reemplazadas solo las capturas cuya subida se comprobó en el PR fusionado

**Closes:** AC12

**Depends on:** T2

**Locations:** `content/skills/flow-close/references/artifact-layout.md`; `content/skills/flow-close/references/support-material.md`

**Execution:** main thread

**Test approach:** check

**Changes:** «Retain evidence»: el cuerpo del PR integrado reemplaza la evidencia de texto y las capturas cuya URL subida se comprobó en él; las demás capturas se conservan. Párrafo de imágenes: añade la carpeta `images/before/` y que la limpieza tras la lectura del visor no borra capturas no subidas al PR. «Delete»: una captura es reemplazable solo con esa comprobación. `viewer-publication.md` no cambia: ya remite a `artifact-layout` para qué conservar. No cambia la ronda de decisión del cierre.

**Verification:** lectura del diff contra AC12; `go test ./tests/content/...` pasa.

### T4 — Documentación

- [x] Antes de editar, capturas «antes» de `flows/close.md` del sitio construido a 1280 y 440 px; después, `flows/build.md`, `flows/verify.md`, `flows/close.md`, `workspace-and-artifacts.md` e `instruction-resources.md` describen el cambio

**Closes:** AC6

**Depends on:** T1, T2, T3

**Locations:** `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `site/src/content/docs/flows/close.md`; `_support/docs/architecture/workspace-and-artifacts.md`; `_support/docs/architecture/instruction-resources.md`; capturas en `_support/workspace/2026-11-17-pr-before-after-captures/images/before/`

**Execution:** main thread

**Test approach:** check

**Changes:** Primero, siguiendo la regla de T1, construir y servir el sitio sin cambios y capturar `flows/close.md` a 1280 y 440 px. Después, las páginas describen el comportamiento y enlazan la skill sin copiar sus reglas; `instruction-resources.md` nombra la referencia nueva de `git-workflow`. `README.md` y `llms.txt` no cambian: no mencionan la evidencia del PR (comprobar con `rg`).

**Verification:** `npm --prefix site run check`, `lint` y `build` terminan con código 0; las páginas construidas muestran el texto nuevo; las 2 capturas «antes» existen.

### T5 — Delta de la especificación

- [x] `specs/workflow-lifecycle/spec.md` de esta carpeta refleja lo que T1–T3 dejaron

**Closes:** AC8, AC9, AC10, AC11, AC12

**Depends on:** T1, T2, T3

**Locations:** `_support/openspec/changes/pr-before-after-captures/specs/workflow-lifecycle/spec.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Ajustar el requisito añadido y sus escenarios a la redacción final.

**Verification:** `python3 -I <flow-plan-dir>/scripts/check_change.py <carpeta>` termina con código 0.

### T6 — Candidato verificado y PR con su propia tabla

- [ ] Suite completa y sitio en verde sobre el candidato; el PR de este cambio lleva la tabla antes | después de `flows/close.md` a 1280 y 440 px, subida con `--attach` (D8 A)

**Closes:** AC9

**Depends on:** T1, T2, T3, T4, T5

**Locations:** el candidato; el PR de este cambio

**Execution:** main thread

**Test approach:** check

**Changes:** Capturar `flows/close.md` del sitio construido del candidato a 1280 y 440 px; tras `/code-review`, abrir el PR no borrador con la tabla según la referencia nueva, esperar CI y parar con el PR abierto (D4 C).

**Verification:** `go test ./...` y el build del sitio pasan; `gh pr view --json body` muestra 4 URLs de imágenes de GitHub; `git diff --stat origin/development...HEAD` no lista imágenes; `wc -c` de cada skill y referencia tocada ≤ 16000.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | `go test ./tests/content/...` | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...`; `npm --prefix site run check && npm --prefix site run lint && npm --prefix site run build` | Covered by the implementation |
| UI check | Final candidate, locally | `flows/close.md` construida a 1280 y 440 px, antes y después | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` (D5 A) | D5 A |
| AC9 en su lugar | Tras el push | cuerpo del PR con 4 imágenes de GitHub | D4 C, D8 A |
| AC8, AC10–AC12 en su lugar | Tras el merge del usuario y la actualización local | texto instalado en los hosts | Actualización según `AGENTS.md` |

Sin recorrido humano antes del push: la entrega es PR listo y el usuario revisa el PR abierto, con su tabla, antes de fusionarlo.

## Checkpoint

- Updated: 2026-11-17
- Stage: plan · revisión aplicada, pendiente de aprobación
- Next: el usuario aprueba el plan y empieza el build

## Review status and progress

2026-11-17 — Plan escrito.
2026-11-17 — Revisión del plan aplicada (2 P1 resueltos con D6 A, D7 A y D8 A; P2 y P3 corregidos en tareas y criterios).
2026-11-17 — T1 · 5ada3116 · unverified: AC8 observed only in installed text after merge and local refresh
2026-11-17 — T2 · 7c7b20f6 · unverified: AC9 needs the PR body after push (T6); AC10-AC11 only in installed text after refresh; whether gh rewrites the relative body-file references checked in T6
2026-11-17 — T3 · d7a7b3c2 · unverified: AC12 only in installed text after merge and local refresh
2026-11-17 — T4 · 5e71fc7b · unverified: none
2026-11-17 — T5 · uncommitted · unverified: none
