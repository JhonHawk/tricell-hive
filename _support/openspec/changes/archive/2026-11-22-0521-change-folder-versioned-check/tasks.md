# Tasks

Base `development` en `867f5463`; rama `feat/change-folder-versioned-check`. Orden: T1, luego T3, luego T2. T1 y T2 en el hilo principal (dos y tres archivos). T3 en un hijo `hive-build-backend`: lee `session_audit.go` (2359 líneas) y escribe más de tres archivos; su interfaz es la de los criterios existentes y no toca los archivos de T1 ni T2.

### T1 — Subcomando `versioned` en `tasks.py`

- [x] `tasks.py <folder> versioned [--work-repo DIR]` sale con 0, 1 o 2 según el estado del folder, con los cuatro casos de distribución, los de error y los avisos probados con Git real

**Closes:** AC7, AC8, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/flow-plan/scripts/tasks.py`, `tests/skills/tasks_test.py`

**Execution:** main thread

**Test approach:** check

**Changes:**
- Añade el subcomando y su línea en el docstring; no cambia `status`, `done`, `checkpoint` ni `execution`. Solo lee: sin fetch, commit ni push, y sin imprimir comandos de entrega.
- Trabaja con la ruta del folder absoluta y resuelta (en macOS los temporales pasan por `/private/var`). Resuelve el repo con `git -C <folder> rev-parse --show-toplevel`; si falla, sale con 1 («not in a Git repository»), no con el 2 del error interno de `git_out`.
- Lee `Base branch` con `base_branch_of` de `workspace_repo.py`, cargado con `importlib` como `git_reader` carga `git_state.py`; si devuelve `—` (sin `## Hive` ni `origin/HEAD`) y el repo es `direct-base`, sale con 1 y lo dice. Lee `Delivery` de las líneas del `## Hive` del `AGENTS.md` de ese repo.
- Archivos: `git ls-files --others --exclude-standard -z -- <folder>` (sin rastrear) y `git diff --name-only -z HEAD -- <folder>` (modificados), no `status --porcelain`, porque `git_out` recorta el espacio inicial de la primera línea.
- Último commit: `git log -1 --format=%H -- <folder>`; vacío (nunca tuvo commit) sale con 1 sin llamar a `branch --contains`. Con `direct-base`, debe estar en `refs/remotes/origin/<base>` (`git merge-base --is-ancestor`); si no, en alguna rama remota (`git branch -r --contains`).
- Repo del trabajo = `--work-repo` resuelto a su raíz Git, o, sin `--work-repo`, el mismo repo del folder. Se comparan por `git rev-parse --git-common-dir` resuelto, no por la raíz, para que un worktree y el checkout principal del mismo repo cuenten como el mismo. Si el folder vive en el repo del trabajo, un archivo sin rastrear o modificado después de su último commit es un aviso («later edits wait for the archive»); si vive en otro repo, es un fallo.
- Un `--work-repo` que no existe o no está en Git es un argumento inválido: sale con 2.
- Con `--work-repo`, toma el `Specs` de su `## Hive` antes de ` · `, lo resuelve contra ese repo y avisa si falta o si el folder no está debajo de esa ruta; el aviso no cambia la salida.
- Las pruebas crean repos y un `origin` desnudo en un directorio temporal, sin mocks.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'tasks_test.py'` pasa con los casos nuevos: los cuatro de AC7 (el de `Specs` en el repo del trabajo con `tasks.md` modificado y un archivo nuevo tras el commit sale 0 con aviso), folder nunca commiteado, archivo sin rastrear y archivo modificado en un repo de specs separado, commit sin push, `direct-base` con el commit solo en una rama de trabajo, `direct-base` sin base resoluble, carpeta sin Git, y `--work-repo` sin `Specs`, con `Specs: <ruta> · <url>` que contiene el folder (sin aviso), con un `Specs` que apunta a otro lugar, y con `--work-repo` en el checkout principal y el folder en un worktree del mismo repo (aviso, no fallo, para un archivo modificado).

### T2 — `flow-close` corre la comprobación; regla, enforcer y sitio

- [x] `flow-close` paso 3 corre `versioned` y vuelve a `git-workflow` paso 4 con salida 1; la fila `prevents` de `rule-enforcers.md` y la página del cierre lo nombran

**Closes:** AC9

**Depends on:** T1, T3

**Locations:** `content/skills/flow-close/SKILL.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `site/src/content/docs/flows/close.md`

**Execution:** main thread

**Test approach:** check

**Changes:**
- Paso 3, después de la frase del commit del folder, solo un puntero: en el punto de cierre fuera de `hold`, correr `python3 -I <flow-plan-dir>/scripts/tasks.py <change-folder> versioned --work-repo <repo del trabajo>`. Con salida 1, hacer el commit y la entrega como fija `git-workflow` paso 4 y correrlo de nuevo; si sigue en 1, el reporte del paso 6 lista el folder como pendiente con su ruta y su pregunta, que se hace después del reporte como fija `close-modes`. Un aviso se nombra en el reporte. No toca el movimiento del ticket ni copia la regla de entrega.
- `rule-enforcers.md`: dos filas con Home `content/skills/git-workflow/SKILL.md` paso 4: `prevents` con Mechanism `content/skills/flow-plan/scripts/tasks.py` (`versioned`), y `detects` con Mechanism `tests/pilot/session_audit_specs.go` (`change_folder_uncommitted`), el archivo de T3 que contiene ese literal.
- `close.md`: la viñeta sobre el folder dice que el cierre comprueba que quedó versionado, en un repo de specs separado o en el del trabajo, y que lo deja pendiente si no.
- El requisito de la spec va en el delta de este folder.

**Verification:** `go test ./tests/content/...` pasa (tamaño de skills, enlaces, `rule_enforcers_test.go`, huella de lectura) y `rg -n 'versioned' content/skills/flow-close/SKILL.md` muestra el paso en el árbol de trabajo. La guía desplegada se comprueba en el refresco de la instalación tras el merge.

### T3 — Criterio `change_folder_uncommitted` en `--audit-session`

- [x] `--audit-session` reporta el criterio en fail, pass, review y not_observed, con fixtures y las dos sesiones reales de project-g

**Closes:** AC10

**Depends on:** T1

**Locations:** `tests/pilot/session_audit_specs.go` (nuevo), `tests/pilot/session_audit.go` (`codexCompletedItem` y el registro en `auditCriteria`), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/change_folder_uncommitted/`, `_support/docs/harness-engineering/session-review-rubric.md`

**Execution:** child `hive-build-backend`

**Test approach:** check

**Changes:**
- Codex en modo código lanza los comandos interpolados con `exec_command`, cuyo resultado llega sin código de salida (llega por `write_stdin`); el mismo rollout trae un evento `item_completed` con el comando resuelto y `exit_code`. `codexCompletedItem` pasa ese código al éxito de los eventos `exec_item`, solo de ese tipo; el contrato de los eventos `shell` («éxito desconocido no es un merge», `TestAuditUnknownSuccessIsNotRecordedAsSuccess`) no cambia.
- El criterio lee merges, commits y rutas del comando resuelto: los eventos `exec_item` exitosos de Codex y los `shell` exitosos de Claude Code. Su propio lector de merges cuenta también `gh pr merge … --delete-branch`, que `auditMerges` excluye.
- Folder = segmento `changes/<id>` fuera de `archive/` de una ruta escrita. Cuenta como escritura solo un parche, una herramienta de escritura o un comando con forma de escritura conocida: `tasks.py <folder> done|checkpoint|execution`, una redirección (`>`, `>>`, `tee`) o un comando que nombra el segmento y contiene `write_text(` u `open(` con modo `'w'` o `'a'` (como `p = Path('<folder>/…'); p.write_text(…)` de project-g-25). Leer (`cat`, `rg`, `check_change.py`) no cuenta.
- Si todas las escrituras de la sesión al folder son `tasks.py done|checkpoint` posteriores al último merge, el folder ya venía versionado de otra sesión y sus ediciones esperan el archivo (`git-workflow` paso 4): `review`, no `fail`.
- Comandos de Codex: la unión de los `exec_item` con su código de salida y los eventos `shell` con éxito conocido (rollouts sin `CommandExecution`, como los fixtures `adapters/codex-rollout`); no se emparejan, porque contar dos veces un merge o un commit no cambia el resultado.
- `pass`: después de la primera escritura al folder, un `git commit` o `git_state.py commit` cuyo comando nombra el segmento, o un `tasks.py <folder> versioned` sobre él con salida 0. `fail`: un texto del asistente sigue al último merge y no hubo nada de eso. `review`: ningún texto del asistente sigue al último merge. `not_observed`: sin escritura al folder o sin merge. La evidencia nombra el folder.
- Fixtures recortados de Codex y Claude Code: fail, pass por commit, pass por `versioned`, review sin texto tras el merge, review por solo checkpoint tras el merge, escritura por `Path(…).write_text` en un heredoc (de project-g-25), un folder ajeno solo leído (`not_observed`) y sin merge.
- Fila del rubric en «Hive behavior checks» · Delivery. La fila `detects` de `rule-enforcers.md` la escribe T2.

**Verification:** `go test ./tests/pilot -run Audit` y `go test ./tests/content/...` pasan; `go run ./tests/pilot --audit-session <rollout project-g-20> --host codex` y `<rollout project-g-25>` muestran `change_folder_uncommitted` en `fail` con el folder.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |
| Security review | Final candidate, before the push | `hive-review-security` | The security answer |

Recorrido de solo lectura sobre folders reales: `versioned` en `project-f-specs/openspec/changes/fac-235-close-internet-exposure` con `--work-repo` del portal `project-f-web-client-reforge`, corrido desde la raíz del workspace `project-f/` (esperado: 1, archivos sin rastrear y aviso de `Specs` sin declarar); y en `_support/openspec/changes/archive/2026-11-22-0350-reuse-sound-base` de este repo (esperado: 0).

## Checkpoint

- Updated: 2026-11-21
- Stage: verify · 6 integrate
- Last commit: 2c68b46b
- Open findings: H1: squash-merged folder edited later on base reads not pushed — close findings; H2: content writes after the last merge not checked by the audit — close findings; H3: git add -u, :/, * and -Av not counted as broad adds — close findings; H4: commitsAll misreads short-flag clusters with attached values — close findings; H5: printable() not applied to the repo header, relpath suggestion and folder error — close findings; H6: versioned run with . from inside the folder is not credited by the audit — close findings
- Next: push, PR to development, CI wait, merge, flow-close, refresh install
- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md
- Principles: none

### Evidence

- before captures: none (no screen)
- python3 -m unittest discover -s tests/skills -p tasks_test.py · OK 41 tests · 73fbfbff
- mutations same_repo/direct_base/specs-holds · each fails VersionedTest · 73fbfbff
- go vet ./... · exit 0 · be6375c0
- go test ./... · exit 0 · be6375c0
- python3 -m unittest discover -s tests/skills -p *_test.py · OK 355 tests · be6375c0
- CI: only master-source-branch.yml (pull_request_target to master), CI-only, not reproduced locally
- go vet ./... · exit 0 · 68eb563f; go vet ./tests/pilot · exit 0 · 2c68b46b
- go test ./... · exit 0 · 68eb563f; go test ./tests/pilot · ok · 2c68b46b
- python3 -m unittest discover -s tests/skills -p *_test.py · OK · 68eb563f; tasks_test.py · OK · 2c68b46b
- review: native /code-review medium on 867f5463..be6375c0 (3 blocking fixed in c2a97948, 68eb563f) and on be6375c0..68eb563f (2 blocking fixed locally in b5ee1400, 2c68b46b, Unreviewed)
- security: hive-review-security on be6375c0, 1 low fixed in 68eb563f
- walk: project-f fac-235 from workspace root exit 1 + Specs warning; tricell-hive archived folder from tooling/ exit 0; audit project-g-20/project-g-25 fail, PRs 572/575/577 pass

## Review status and progress

2026-11-21 — T1 · 73fbfbff · unverified: none
2026-11-21 — T3 · 7b39ec8e · unverified: none
2026-11-21 — T2 · be6375c0 · unverified: AC9: deployed guidance is observed at the refresh after the merge
