# Tareas

El plan parte de la base `development` en `6670e799`. Se trabaja en la rama `fix/viewer-before-approval-and-close-hygiene`, dentro del worktree `.claude/worktrees/viewer-close-hygiene`. Se usa un worktree porque el checkout principal tiene carpetas sin trackear que no son de este trabajo y porque otra sesión trabaja en el repositorio. Antes de la primera edición, este change folder, que no tiene commit, se mueve del checkout principal a la misma ruta dentro del worktree, como fija el paso 3 de `git-workflow`.

Cada commit nombra en su mensaje las sesiones donde se observó la falla (host, ID de sesión y fecha), como pide «Measurement» de `AGENTS.md`:
- **Visor:** Claude Code 5ff8bac1, 062c567a y 98476eca; Codex 01a1a00a y 01a1a30f.
- **Higiene:** Codex 01a1a8a8 y 01a1aa61, y Claude Code 746d5e32 (facti).

Quién hace cada tarea:
- **T1** edita cuatro archivos con su texto ya fijado en `design.md`. Supera el umbral de «edits more than 3 files» de `delegation.md`, así que va a un hijo `hive-write-docs`.
- **T2 y T3** editan tres archivos cada una. Se quedan en el hilo principal, una tras otra, mientras trabaja el hijo de T1; no comparten archivos con él.
- **T4** edita cinco archivos y lee las cuatro skills que cambian, así que va a un hijo `hive-write-docs` cuando T1, T2 y T3 terminan.

El hilo principal revisa cada resultado contra `design.md` y hace los commits.

### T1 — La opción de abrir el visor

- [x] La oferta del visor es una opción dentro de la pregunta de aprobación. Al elegirla, se abre el visor y se vuelve a hacer la misma pregunta con el mismo ID y sin esa opción.

**Closes:** AC1, AC2

**Depends on:** none

**Locations:**
- `content/skills/flow-walkthrough/references/viewer-publication.md`
- `content/skills/flow-walkthrough/SKILL.md`
- `content/skills/flow-build/SKILL.md`
- `content/skills/flow-verify/SKILL.md`
- El delta `specs/walkthrough-viewer/spec.md` de este folder, solo para revisarlo frente al texto final

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** la regla «Open the viewer», con su excepción nombrada, y los tres punteros, como fija `design.md` T1. No cambian el ancla `#open-the-viewer`, `plan-handoff.md` ni el resto de `viewer-publication.md`.

**Verification:**
- `rg -n "one more question of that round|open-viewer question" content` no devuelve nada.
- `rg -n "same ID|without the open|never first|starred" content/skills/flow-walkthrough/references/viewer-publication.md` muestra la segunda pregunta y la excepción.
- `rg -n "open-viewer option" content/skills` muestra los tres punteros.
- `go test ./tests/content/...` pasa.

### T2 — Limpieza del cierre: referencia, barrido, autorización, pausa y promociones

- [x] La limpieza de ramas vive en una referencia, recorre los repositorios que tocó el trabajo y barre con seguridad las ramas propias ya mergeadas. `global.md` lo autoriza, el cierre de pausa queda definido fuera de `hold`, y una promoción no hecha no retiene nada.

**Closes:** AC3, AC4, AC6, AC7

**Depends on:** none

**Locations:**
- `content/skills/flow-close/SKILL.md`
- `content/skills/flow-close/references/branch-cleanup.md`, que es nuevo
- `content/guidance/global.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** como fija `design.md` T2. La espera al pie del cierre (el párrafo de la línea 21) y las prohibiciones de `global.md:31` no cambian. Se mide el tamaño de `global.md` antes y después, para el pull request.

**Verification:**
- `wc -c content/skills/flow-close/SKILL.md` da como máximo 16000.
- `rg -n "lsof|--ignored|never switch|no commits" content/skills/flow-close/references/branch-cleanup.md` muestra las condiciones de seguridad.
- `rg -n -i "promotion" content/skills/flow-close/SKILL.md` muestra que una promoción no retiene nada.
- `rg -n "earlier work" content/guidance/global.md` muestra la autorización.
- `go test ./tests/content/...` y `go test ./tooling/management/ -run Reference` pasan.

### T3 — Specs en el repo `direct-base`, commit al pausar, ancla y promociones en el plan

- [x] `git-workflow` manda editar los specs `direct-base` sin rama ni worktree, commitear el change folder al pausar y enlazar la referencia nueva. `flow-plan` agrega la promoción como tarea cuando el ticket o la solicitud piden el cambio en un entorno.

**Closes:** AC5, AC6, AC7

**Depends on:** T2, porque usa la referencia nueva y el cierre de pausa de `flow-close`

**Locations:**
- `content/skills/git-workflow/SKILL.md`
- `content/skills/flow-plan/SKILL.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** como fija `design.md` T3. Bajo `hold`, el folder sigue sin commit. Con `Specs` en otro repositorio sin `direct-base`, el folder queda pendiente y se recomienda declarar `direct-base` (D6 A).

**Verification:**
- `rg -n "#branch-cleanup" content site _support/docs` no devuelve nada.
- `rg -n "direct-base" content/skills/git-workflow/SKILL.md` muestra el caso del paso 3 y el commit al pausar.
- `rg -n -i "environment" content/skills/flow-plan/SKILL.md` muestra la tarea de promoción.
- `go test ./tests/content/...` y `go test ./tooling/management/ -run Reference` pasan.

### T4 — README, llms.txt, sitio y CHANGELOG

- [x] El README, `llms.txt`, el sitio y el CHANGELOG describen los cambios, y los deltas coinciden con el texto final.

**Closes:** AC8

**Depends on:** T1, T2, T3

**Locations:**
- `README.md`, líneas 371-376
- `llms.txt`, línea 37
- `site/src/content/docs/flows/close.md`, incluidas las líneas 3 y 56
- `site/src/content/docs/flows/plan.md`
- `CHANGELOG.md`, en Unreleased
- El delta `specs/workflow-lifecycle/spec.md` de este folder, solo para revisarlo frente al texto final

**Execution:** hijo `hive-write-docs`

**Test approach:** check

**Changes:** las páginas describen el comportamiento y enlazan a la skill, sin copiar sus reglas. No se tocan los detectores de `tests/pilot/`.

**Verification:**
- `rg -n "task-owned|only what that work created|stays where it is" README.md llms.txt site/src/content/docs/flows/close.md` no devuelve texto que contradiga las reglas nuevas.
- `go test ./tests/content/...` pasa, incluido `spec_deltas_test.go`.
- `cd site && pnpm build` termina sin errores.

## Verificación y revisión humana

| Comprobación | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, en local | Pasa el comando de la tarea | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go test ./...` | Cubierta por la implementación |
| Sitio | Candidato final, en local | `cd site && pnpm build` | Cubierta por la implementación |
| Revisión de lo que depende de las reglas | Candidato final | Las búsquedas de `rg` de cada tarea y la lista de lugares revisados en el pull request | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` | La respuesta D4 |

No hay interfaz que recorrer. La entrega es `automatic` (D3), así que no hay parada de validación.

## Estado de revisión y avance

2026-11-05 — Plan escrito. La revisión de `hive-review-plan` dejó 13 hallazgos. Se aplicaron H3 a H13 y la parte de H1 que no necesita decisión. D5 A y D6 A resolvieron H1 y H2.
2026-11-05 — Base real del build: `origin/development` en `c011214e`, porque llegó #380 mientras se planeaba; no toca ningún archivo del plan.
2026-11-05 — T2 hecho en `e667cd96`: `flow-close/SKILL.md` queda en 15 626 bytes y `global.md` pasa de 35 640 a 35 810 bytes. Pasan `go test ./tests/content/...` y `go test ./tooling/management/ -run Reference`.
2026-11-05 — T1 hecho (commit siguiente): lo escribió un hijo `hive-write-docs`. El hilo principal corrigió la excepción para que nombre solo las dos reglas que cambian. El patrón de verificación con `rg` se amplió a `never the first|same ID|without the open|never carries`. Pasa `go test ./tests/content/...`.

2026-11-05 — T3 hecho (commit siguiente): `git-workflow` queda en 14 278 bytes y `flow-plan` en 8 423. `rg "#branch-cleanup"` no encuentra nada. Pasan `go test ./tests/content/...` y `go test ./tooling/management/ -run Reference`.

2026-11-05 — T4 hecho (commit siguiente): lo escribió un hijo `hive-write-docs`. El hilo principal corrigió dos frases: la parada del plan ofrece el visor como respuesta, no como opción, y el README dice «your earlier merged branches». El delta de workflow-lifecycle coincide con el texto. Pasan `go test ./tests/content/...` y `pnpm build` (17 páginas). Pasa `go test ./...` completo sobre `7dcc1f35`.

2026-11-05 — Verify: `/code-review` sobre `03349f19` dejó 10 hallazgos, todos confirmados al leer el código (dos reproducidos por el revisor). Se corrigieron en una ronda (commit siguiente):
- una ticket ya no autoriza un despliegue: solo la petición del usuario lo hace;
- el barrido en un repo sin PRs usa la memoria en lugar del reflog;
- `headRefOid` se compara con la copia local y con la remota;
- se conservan el trabajo en curso y los cierres pendientes de menos de 24 h, y un cierre pendiente más antiguo se reanuda en lugar de borrarse;
- un checkout de specs que no está en su base no se cambia;
- se hace stage de las rutas por nombre antes de `--only`;
- las condiciones de conservación aplican también al worktree del propio trabajo;
- la excepción del visor nombra tres reglas y escribe la opción en el texto;
- se corrigieron las FAQ del sitio.

Pasan `go test ./tests/content/...`, la prueba de referencias y `pnpm build`.

2026-11-05 — La segunda revisión, limitada a `03349f19..27c43d5d`, dejó 10 hallazgos locales, corregidos sin otra ronda (commit siguiente, con línea `Unreviewed`):
- el worktree del propio trabajo ya no queda bloqueado por el proceso de la sesión;
- un worktree conservado conserva también sus ramas;
- el reloj de 24 h parte de un registro de cierre pendiente;
- el barrido exige un merge en la base (`gh pr list --base`);
- `flow-deploy` sigue decidiendo cuándo pregunta producción;
- la copia remota se reporta con `git log <headRefOid>..origin/<branch>`;
- se corrigieron una frase partida en `plan.md`, la referencia a «flow-close step 2» y la FAQ de pausa bajo `hold`;
- se actualizaron `design.md` y el delta.

Checkpoint: verify · siguiente paso: suite completa, push, PR y merge.
