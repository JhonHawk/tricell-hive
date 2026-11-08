# Tasks

Base: `development` @ `71c18043`, rama `feat/scriptable-ops-helpers`. T1 corre en el hilo principal (2 archivos). T2 va a un hijo `hive-build-backend` porque edita 5 archivos, con interfaz fijada en `design.md` y escrituras disjuntas de T1, así que corren en paralelo. T3 corre en el hilo principal después de T1 y T2, porque cita sus interfaces.

### T1 — `git_state.py`: estado de git y del PR en una llamada

- [x] `content/skills/git-workflow/scripts/git_state.py` con la forma de `design.md`
- [x] `tests/skills/git_state_test.py` con repos temporales (un remoto bare) y un `gh` falso en el `PATH` del test

**Closes:** AC1, AC2, AC3, AC4

**Depends on:** none

**Locations:** `content/skills/git-workflow/scripts/git_state.py`, `tests/skills/git_state_test.py`

**Execution:** main thread

**Test approach:** check

**Changes:** script nuevo de solo lectura (salvo `git fetch` con `--fetch`). No toca la limpieza de ramas ni espera CI. El test fija `sys.dont_write_bytecode`, como `change_check_test.py:11`, y corre el script por su ruta con `python3 -I`.

**Verification:** `python3 -m unittest tests/skills/git_state_test.py` pasa, con casos para: adelante 2 / atrás 1, base integrada en local sin push (`base_in_pushed_head: false`) y tras el push, y fetch fallido con comparaciones en `null` (AC1); con el `gh` falso: revisión de bot vieja y del head, comentario de revisión humano viejo, respuesta en hilo y comentarios antes y después de `--since`, comprobando `on_head` contra `head_ref_oid` y `after_since` (AC2); `gh` ausente, `gh` que falla, un remoto HTTP local que responde 401 y un `gh` falso que tarda más que un límite reducido por el test, que termina con `error` dentro del límite (AC3); `git status --porcelain`, `git rev-parse` de las ramas locales y `git config --list` iguales antes y después (AC4).

### T2 — `tasks.py` y la sección `## Checkpoint`

- [x] `content/skills/flow-plan/scripts/tasks.py` con `status`, `done` y `checkpoint`
- [x] `check_change.py` exime `## Checkpoint` como la sección de progreso
- [x] La plantilla `tasks.md` trae `## Checkpoint` antes de `## Review status and progress`

**Closes:** AC5, AC6

**Depends on:** none

**Locations:** `content/skills/flow-plan/scripts/tasks.py`, `tests/skills/tasks_test.py`, `content/skills/flow-plan/scripts/check_change.py`, `tests/skills/change_check_test.py`, `content/skills/flow-plan/assets/change-folder/tasks.md`

**Execution:** `hive-build-backend`

**Test approach:** check

**Changes:** script nuevo y autocontenido (con `-I` no puede importar `check_change.py`); `check_change.py` solo amplía la exención a `## Checkpoint` y su `### Evidence`; la plantilla suma una sección. No cambia el formato de las tareas ni el visor. Los tests fijan `sys.dont_write_bytecode`.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa, con casos para: `done` marca solo esa tarea, escribe una línea con `unverified:` y no la duplica al repetirse, imprime `✓ T2 · 2/4 · …`, y sale con 2 ante una tarea inexistente o sin título de progreso (AC5); `checkpoint` con un `--check` crea la sección, una segunda llamada reemplaza las líneas fijas y conserva la evidencia, y `check_change.py` sale con 0 sobre una carpeta cuyo checkpoint nombra una decisión reemplazada (AC6).

### T3 — Las skills invocan los scripts, y la documentación

- [ ] `git-workflow/SKILL.md` pasos 2, 7, 8 y 9 invocan `git_state.py`, con los comandos manuales como alternativa si falta o sale con 2, y conservan el filtro de revisiones solo para bots y los hilos humanos abiertos
- [ ] `flow-build/SKILL.md` pasos 2 y 5 y el checkpoint, y `flow-verify/SKILL.md` (checkpoint, líneas 21 y 25), invocan `tasks.py` en carpetas de cambio, con la edición manual como alternativa y sin cambiar el checkpoint de un plan guardado en la conversación
- [ ] README, `llms.txt` si nombra requisitos, `site/` y `instruction-resources.md` al día

**Closes:** AC7

**Depends on:** T1, T2

**Locations:** `content/skills/git-workflow/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md`, `README.md`, `llms.txt`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/verify.md`, `_support/docs/architecture/instruction-resources.md`

**Execution:** main thread

**Changes:** punteros con la forma `python3 -I <git-workflow-dir>/scripts/git_state.py` y su link Markdown; cada regla de orden exacto (fetch antes de comparar, no comparar si el fetch falló, base contenida en el head subido antes del merge, una lectura sin bucles) sigue escrita en la skill, el script solo la ejecuta. «Delivery» en `global.md` no cambia. Si `git-workflow/SKILL.md` pasa de 16.000 B, la alternativa manual va a `git-workflow/references/manual-state.md` con su condición de lectura.

**Verification:** `go test ./tests/content/` pasa (tamaños y links); revisión del comportamiento completo según `AGENTS.md` «Guidance structure», con la lista de lugares revisados en el PR, incluido `site/src/content/docs/flows/verify.md:30`.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Prueba real de `git_state.py` | Candidato final, en este repo | `--pr 426` (fusionado) y la rama del cambio sin `--fetch`, comparados con `gh api` de los mismos endpoints | Cubierto por la implementación (solo lectura) |
| Revisión de código | Candidato final, antes del push o de la validación | Native: Claude Code `/code-review` | La respuesta de revisión |

## Checkpoint

- Updated: 2026-11-08
- Stage: build · T3
- Last commit: 9f7cfb16
- Open findings: none
- Next: T3: skill pointers and docs

### Evidence

- python3 -m unittest discover -s tests/skills -p '*_test.py' · exit 0, 81 tests · 9f7cfb16

## Review status and progress

2026-11-08 — Plan escrito.
2026-11-08 — T1 · d912db7e · unverified: real read only on a PR without reviews (#426)
2026-11-08 — T2 · 9f7cfb16 · unverified: none
