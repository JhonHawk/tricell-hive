# Tasks

Base `development` en `1b87f0a7`, rama `fix/no-post-merge-writes`, worktree `.claude/worktrees/no-post-merge-writes`. Las tres tareas corren en el hilo principal: T1 toca 2 archivos, T2 es texto corto que cita los mensajes de T1 y T3 es un caso del detector.

### T1 — `tasks.py` no escribe en una carpeta integrada y completa; `checkpoint --out`; `versioned`

- [ ] `checkpoint` y `done` salen con 1 sin escribir cuando la carpeta está en `origin/<base>` con todas sus tareas marcadas; `checkpoint --out` escribe en otro archivo; `versioned` indica `git restore` para un `tasks.md` cambiado solo en su checkpoint.

**Closes:** AC2, AC4, AC6

**Depends on:** none

**Locations:** `content/skills/flow-plan/scripts/tasks.py`, `tests/skills/tasks_test.py`

**Execution:** hilo principal

**Test approach:** tdd — el script es un contrato que usan cuatro flujos. Casos con un repositorio temporal y su remoto desnudo, como `VersionedTest`: integrada y completa desde la base, desde un commit sin rama y desde la rama ya integrada (sale 1, nombra `--out`, `tasks.md` igual byte a byte); merge por squash; `done` en el mismo caso; una tarea abierta sigue escribiendo; una entrega posterior en su rama marca su última tarea con `done` y luego escribe `checkpoint` sin commit intermedio, y sigue escribiendo (la negativa lee la copia integrada); carpeta solo en su rama sigue escribiendo; repositorio sin commits, sin remoto y sin base resoluble siguen escribiendo; `--out` crea el archivo con solo el bloque, imprime la ruta, conserva la evidencia en la segunda llamada y no toca `tasks.md`; `--out` dentro del repositorio se rechaza. `versioned`: la prueba dueña `test_specs_in_the_work_repository_pass_with_later_edits_as_warnings` cambia a las advertencias nuevas (solo checkpoint → `git restore`; otro archivo → reporte).

**Changes:** la comprobación de «en `origin/<base>`» se extrae de `cmd_versioned` a una función que ambos llaman, sin copiarla; cualquier error de Git en la comprobación deja escribir. El docstring y el uso del módulo nombran `--out` y la negativa. Mensajes exactos en `design.md`.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'tasks_test.py'` pasa.

### T2 — Reglas del cierre, `git-workflow`, `checkpoint.md` y el sitio

- [ ] `close-hygiene` dice dónde escribe y lee el cierre su checkpoint después del merge y deja de mandar `git restore`; `checkpoint.md`, `git-workflow` paso 4 y `flows/close.md` dejan de decir que las ediciones posteriores al merge esperan al archivo.

**Closes:** AC4, AC5

**Depends on:** T1

**Locations:** `content/skills/hive-core/references/close-hygiene.md`, `content/skills/hive-core/references/checkpoint.md`, `content/skills/git-workflow/SKILL.md`, `site/src/content/docs/flows/close.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check — pruebas de contenido existentes (tamaños, paquetes de etapa, presupuesto de lectura). La frase nueva de `close-hygiene` no pasa de +10 bytes sobre la que reemplaza, y `checkpoint.md` queda neto cero.

**Changes:** textos de `design.md` «Texto». `batch-archive.md` no cambia. AC5 se observa en el cierre de este mismo cambio, con la versión desplegada tras el refresco del mantenedor.

**Verification:** `go test ./tests/content/...` pasa; `grep -c 'git restore' content/skills/hive-core/references/close-hygiene.md` da 0.

### T3 — Detector de sesiones

- [ ] `commandWrites` no cuenta `checkpoint … --out` como escritura en la carpeta, y el comentario de `changeFolderUncommitted` deja de decir que las ediciones esperan al archivo.

**Closes:** AC4

**Depends on:** T1

**Locations:** `tests/pilot/session_audit_specs.go`, `tests/pilot/session_audit_test.go` (un caso), `tests/fixtures/sessions/change_folder_uncommitted/` (un fixture con `--out`)

**Execution:** hilo principal

**Test approach:** check — un caso de `TestAuditNewCriteria` con una llamada `checkpoint … --out` después del merge que no cuenta como escritura.

**Changes:** el veredicto del criterio no cambia.

**Verification:** `go test ./tests/pilot/ -run TestAuditNewCriteria` pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Pruebas de cada tarea | Al cerrar cada tarea, en el worktree | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Candidato final, en el worktree | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Candidato final, antes del push | `/code-review` de Claude Code | The review answer |
| AC5 | Cierre de este cambio, tras el refresco | `git status` del checkout del cierre | Covered by the implementation |

Sin parada de validación humana (`automatic`).

## Checkpoint

- Updated: 2026-11-21
- Stage: build · T1
- Last commit: 1b87f0a7
- Open findings: none
- Next: T1 tdd: refusal tests first
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/no-post-merge-writes · fix/no-post-merge-writes
- Rules: flow-build/stage.md

### Evidence

- before captures: none (no screen)

## Review status and progress
