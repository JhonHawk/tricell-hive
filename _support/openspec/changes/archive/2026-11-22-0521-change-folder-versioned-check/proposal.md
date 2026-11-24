# change-folder-versioned-check — el cierre comprueba que el change folder quedó versionado

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #580, merge 01e43b50 |
| Tracker | ninguno · watch item `engram:tricell-hive/watch/change-folder-uncommitted-specs-repo` (#17924) |
| Git | tricell-hive · base `development` en `867f5463` · rama propuesta `feat/change-folder-versioned-check` |
| Playbook | infra-ci |

## Decisiones

2026-11-21
- **D1 A** — ¿Qué enfoque impide cerrar con el change folder sin commit? Un subcomando `versioned` en `tasks.py`, que `flow-close` paso 3 corre antes del reporte final; descarta extender `cleanup_branches.py`, que no sabe qué folder es de este trabajo.
- **D2 A** — ¿Qué hace la verificación si el repo del código no declara `Specs` pero el folder vive en otro repo? Usa el repo que contiene el folder, exige igual el commit y avisa qué repo de código debe declarar `Specs`; no bloquea por la configuración.
- **D3 A** — ¿Se añade el criterio `change_folder_uncommitted` a `--audit-session`? Sí, para detectar la falla en sesiones de cualquier host, también donde no se corrió el subcomando.
- **supuesto** — Sin aviso en `hive doctor` de folders sin commit: no distingue un plan en curso de un trabajo cerrado; descartado avisar por el estado "Completed" de la propuesta, que es texto libre en varios idiomas.
- **supuesto** — Si el repo que contiene el folder no declara `Base branch`, la base es la de `origin/HEAD`; descartado bloquear.
- **supuesto** — Con `Specs` en el repo del trabajo, un archivo rastreado y modificado después de su último commit es un aviso, no un fallo: `git-workflow` paso 4 deja que el archivo del lote de higiene las commitee. Descartado limitar la comprobación a repos de specs separados, que dejaría fuera el caso repo/monorepo que pediste.
- **supuesto** — `versioned` no imprime comandos de entrega: con salida 1 el cierre vuelve a `git-workflow` paso 4, que fija el commit con escaneo de secretos (`git_state.py commit`), el PR o el push a la base y la revisión de commits ajenos. Descartado imprimir comandos por modo de entrega, que copiaría esa regla.
- **supuesto** — La pausa queda fuera: el subcomando cubre el punto de cierre; las reglas de pausa siguen en prosa (`git-workflow` paso 4). Descartado un `--pause` que nadie pidió.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

«A» (2026-11-21), a la opción «Planear el mecanismo en `tasks.py`/`flow-close` con sus pruebas y la fila en `rule-enforcers.md`» del diagnóstico; el objetivo dado en ese diagnóstico: que el cierre no pueda terminar mientras el change folder siga sin commit en el repo de specs.

«pero tambien importante considerar que puede ser repo/monorepo o un workspace desde donde se ejecuta» (2026-11-21).

Origen: tres sesiones de Codex `gpt-6.2-sol` cerraron con el folder sin commit (project-g PRK-20 `01a1f5d4` y PRK-25 `01a1f5f6`, 2026-11-19, en un repo de specs con `Delivery: direct-base`; project-f `01a200f1`, 2026-11-21). Sin ticket.

## Objetivo

Hoy la regla de `git-workflow` paso 4 («commit it there when the work reaches its completion point») vive solo en prosa y el cierre puede reportar «no quedan pendientes» con el folder sin rastrear. Con el cambio, `tasks.py <folder> versioned` dice con su código de salida si el folder está versionado, resolviendo el repo desde el propio folder (repo, monorepo o workspace), `flow-close` lo corre antes del reporte, y `--audit-session` detecta las sesiones que no lo hicieron.

## Alcance y aceptación

Dentro: el subcomando `versioned` de `tasks.py` y sus pruebas; `flow-close` paso 3; el requisito en la spec `workflow-lifecycle`; el criterio de auditoría con su fixture y su fila del rubric; la fila de `rule-enforcers.md`; la página `site/` del cierre.

Fuera: la pausa; `hive doctor`; corregir la configuración de project-f, project-b y project-g; commitear los folders que ya quedaron locales en esos proyectos.

- AC1 — `python3 -I tasks.py <folder> versioned` sale con 0 e imprime el repo que contiene el folder cuando todos sus archivos están en un commit que ya está en una rama remota, en cuatro distribuciones: `Specs` dentro del repo, un monorepo con el folder en su raíz e invocado desde un subpaquete, un repo de specs hermano y un workspace cuya raíz no es repo Git · false today because el subcomando no existe y `argparse` sale con 2 · Observed in: repos Git temporales con un `origin` desnudo que crea `tests/skills/tasks_test.py` · **Retired:** exigía cero archivos modificados también cuando el folder vive en el repo del trabajo, donde `git-workflow` paso 4 deja las ediciones posteriores para el archivo (revisión del plan, 2026-11-21)
- AC2 — Sale con 1 cuando el folder tiene archivos sin rastrear o modificados, o cuando su último commit no está en ninguna rama remota (en `origin/<base>` si el repo declara `Delivery: direct-base`), e imprime los archivos y los comandos `git -C <repo>` que faltan, con el push a la base solo para `direct-base` · false today because el subcomando no existe · Observed in: repos Git temporales de `tests/skills/tasks_test.py` · **Retired:** los comandos impresos no cubrían la entrega por PR ni el escaneo de secretos de `git_state.py commit`; la entrega vuelve a `git-workflow` paso 4 (revisión del plan, 2026-11-21)
- AC7 — **Replaces:** AC1 · `python3 -I tasks.py <folder> versioned` sale con 0 e imprime el repo que contiene el folder cuando su último commit está en una rama remota y, fuera del repo del trabajo, ningún archivo suyo está sin rastrear ni modificado, en cuatro distribuciones: `Specs` dentro del repo del trabajo (donde un archivo sin rastrear o modificado después de ese commit, como la línea del checkpoint, sale como aviso y no cambia el 0), un monorepo con el folder en su raíz e invocado desde un subpaquete, un repo de specs hermano y un workspace cuya raíz no es repo Git · false today because el subcomando no existe y `argparse` sale con 2 · Observed in: repos Git temporales con un `origin` desnudo que crea `tests/skills/tasks_test.py`
- AC8 — **Replaces:** AC2 · Sale con 1 y nombra cada archivo y el estado que falta, sin imprimir comandos de entrega, cuando nunca tuvo commit, cuando su último commit no está en ninguna rama remota (en `origin/<base>` si su repo declara `Delivery: direct-base`), o cuando, viviendo en un repo distinto del trabajo, tiene un archivo sin rastrear o modificado · false today because el subcomando no existe · Observed in: repos Git temporales de `tests/skills/tasks_test.py`
- AC3 — Sale con 1 e indica que el folder no se puede versionar cuando no está dentro de ningún repo Git · false today because el subcomando no existe · Observed in: carpeta temporal sin Git de `tests/skills/tasks_test.py`
- AC4 — Con `--work-repo <dir>` cuyo `## Hive` no declara `Specs`, o lo declara apuntando a otro lugar, imprime un aviso que nombra ese repo y la ruta que debe declarar, sin cambiar el código de salida · false today because el subcomando no existe · Observed in: repos Git temporales de `tests/skills/tasks_test.py`
- AC5 — `flow-close` paso 3, en el punto de cierre fuera de `hold`, corre `versioned` sobre el change folder del trabajo y, con salida distinta de 0, no reporta el trabajo como terminado: ejecuta los comandos que imprime según el `Delivery` de ese repo o lo deja como pendiente con su pregunta · false today because `content/skills/flow-close/SKILL.md` paso 3 no nombra ninguna comprobación · Observed in: la guía desplegada en `~/.agents/skills/flow-close/SKILL.md` tras `hive update` · **Retired:** ejecutar comandos impresos repetía y contradecía `git-workflow` paso 4, y «no reportar como terminado» chocaba con el movimiento del ticket del mismo paso (revisión del plan, 2026-11-21)
- AC6 — `go run ./tests/pilot --audit-session <rollout> --host codex` reporta `change_folder_uncommitted` en `fail` sobre la sesión real de project-g PRK-20 (`01a1f5d4`), en `pass` cuando después del merge hay un commit que nombra la ruta del folder o un `versioned` que salió con 0, y en `not_observed` sin merge · false today because el criterio no existe · Observed in: el rollout real `~/.codex/sessions/2026/11/19/rollout-2026-11-19T16-01-30-01a1f5d4-3ec7-7631-97d8-ba540210ac98.jsonl` y los fixtures de `tests/fixtures/sessions/change_folder_uncommitted/` · **Retired:** el auditor no ve ese merge (éxito desconocido en la llamada interpolada) y un cierre correcto con `Specs` en el repo del trabajo no hace commit después del merge (revisión del plan, 2026-11-21)
- AC9 — **Replaces:** AC5 · `flow-close` paso 3, en el punto de cierre fuera de `hold`, corre `versioned --work-repo <repo del trabajo>`; con salida 1 hace el commit y la entrega como fija `git-workflow` paso 4 y lo corre de nuevo, y si sigue en 1, el reporte del paso 6 lista el folder como pendiente con su ruta y su pregunta, que se hace después del reporte como fija `close-modes`; el movimiento del ticket no cambia · false today because `content/skills/flow-close/SKILL.md` paso 3 no nombra ninguna comprobación · Observed in: la guía desplegada en `~/.agents/skills/flow-close/SKILL.md` en el refresco de la instalación tras el merge; observa el texto del paso, y que un cierre real lo corra lo detecta AC10 en las sesiones posteriores
- AC10 — **Replaces:** AC6 · `go run ./tests/pilot --audit-session <rollout> --host codex` reporta `change_folder_uncommitted` en `fail` sobre las sesiones reales de project-g PRK-20 (`01a1f5d4`) y PRK-25 (`01a1f5f6`), nombrando el folder; en `pass` cuando, después de la primera escritura al folder, un commit nombra su segmento `changes/<id>` o un `versioned` sobre él sale con 0; en `review` si ningún texto del asistente sigue al último merge; y en `not_observed` sin escritura al folder o sin merge · false today because el criterio no existe · Observed in: los rollouts reales `~/.codex/sessions/2026/11/19/rollout-2026-11-19T16-01-30-01a1f5d4-3ec7-7631-97d8-ba540210ac98.jsonl` y `rollout-2026-11-19T16-39-24-01a1f5f6-f1cd-7ea3-9b4f-176fa9ae04e8.jsonl`, y los fixtures de `tests/fixtures/sessions/change_folder_uncommitted/`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- `tasks.py` (855 líneas) ya recibe la ruta del folder y tiene `git_reader`/`git_out` (líneas 407–426), que reutilizan `git_state.py`. `workspace_repo.py` `base_branch_of` (líneas 73–83) ya resuelve la base desde `## Hive` o `origin/HEAD`.
- `flow-close` paso 3 dice «commit its change folder in the specs repository when git-workflow step 4 leaves that commit to the close» sin comprobación. La spec `workflow-lifecycle` no tiene requisito para el commit en el punto de cierre (#311, `3a95ec70`, solo tocó la guía); sí los tiene para la pausa y para `direct-base`.
- `tests/pilot/session_audit.go` (2359 líneas) ya detecta los merges (`auditMerges`) y los borrados de rama; los fixtures viven en `tests/fixtures/sessions/<criterio>/`.
- Distribuciones reales: tricell-hive y project-a-monorepo (`Specs` dentro), project-g y project-b (repo de specs hermano; project-g con `direct-base`), project-f (repo de specs sin `AGENTS.md`, el portal sin `## Hive`), raíz de workspace sin Git (project-g, project-f).
- `python3 -m unittest discover -s tests/skills -p 'tasks_test.py'` · OK, 18 s · 2026-11-21
- `go test ./tests/pilot -run Audit` · ok, 1 s · 2026-11-21
- `go test ./tests/content/...` · ok, 3 s · 2026-11-21
- `go test ./...` · ok, 232 s · 2026-11-21
- `python3 -m unittest discover -s tests/skills -p '*_test.py'` · OK, 347 pruebas, 94 s · 2026-11-21
- `go run ./tests/pilot --audit-session <rollout project-g PRK-20> --host codex` · sale 0 en 1 s; el criterio aún no existe y `merged_branch_deleted` sale `not_observed` porque la llamada del merge vuelve sin código de salida (llega por `write_stdin`) y `auditMerges` descarta un éxito desconocido; el `exec_item` del mismo merge sí trae `exit_code` 0 · 2026-11-21
- `python3 -I check_change.py _support/openspec/changes/change-folder-versioned-check` · sale 0 · 2026-11-21
- `rg -n 'versioned' content/skills/flow-close/SKILL.md` · sin coincidencias en la base (el paso no existe) · 2026-11-21
- `rg -n 'versioned' ~/.agents/skills/flow-close/SKILL.md` · not measured: lo cumple el refresco de la instalación tras el merge
- `go run ./tests/pilot --audit-session <rollout PRK-25> --host codex` · sale 0; el criterio aún no existe · 2026-11-21
- Recorrido: `tasks.py <fac-235> versioned` · not measured: T1 lo cumple primero; `project-f-specs/openspec/changes/fac-235-close-internet-exposure` está sin rastrear hoy y los folders de PRK-20 y del onboarding Amazon V2 ya se archivaron (`382d266`, `14abf21`)

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic` (2026-11-21): revisión, push, PR contra `development` e integración sin parada de validación; después, refresco de la instalación local (marcador de mantenedor). Revisión de código: nativa, Claude Code `/code-review`. Revisión de seguridad: sí, `hive-review-security`. No autoriza promover a `master` ni publicar una versión.

## Estado de revisión y progreso

2026-11-21 — Plan escrito.
2026-11-21 — Revisión del plan: dos hijos `hive-review-plan`, servidor (≈159k tokens, 296 s) y guía (≈121k tokens, 118 s). Hallazgo común: el caso `Specs` en el repo del trabajo fallaba siempre. Revisión: AC1, AC2, AC5 y AC6 retirados y sustituidos por AC7–AC10; `versioned` ya no imprime comandos y el cierre vuelve a `git-workflow` paso 4; el criterio de auditoría lee los eventos `exec_item` con su código de salida; dos filas en `rule-enforcers.md`, la de detección la escribe el hilo principal al volver T3; el recorrido usa folders que existen hoy. Segunda relectura de código: escrituras por `Path(…).write_text`, unión de eventos de Codex sin emparejar, la fila `detects` pasa a T2 (que ahora sigue a T3), y `--work-repo` fuera de Git sale con 2; aceptado sin cambio que, en el repo del trabajo, un archivo nunca subido tras el último commit sea solo aviso, como fija `git-workflow` paso 4 para las ediciones posteriores. Primera relectura de código: solo cuentan escrituras de forma conocida, un folder con solo checkpoint tras el merge da `review`, el repo del trabajo se compara por `--git-common-dir` y Codex cae al evento `shell` sin `exec_item`. Relectura de guía: en el repo del trabajo, un archivo sin rastrear también es aviso (AC7 y AC8 ajustados en su línea, sin cambiar lo que observan fuera de ese caso).
