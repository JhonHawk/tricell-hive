# Tasks

Base `development` en `51f1949d`. Son cuatro PRs, en el checkout principal y en orden. Las tareas que tocan más de 3 archivos van a hijos: `hive-build-backend` para scripts, Go y tests, `hive-write-docs` para guía y documentación. Dos tareas de escritura del mismo PR corren a la vez solo si tocan archivos disjuntos y comparten el contrato fijo del `plan.md` («Proposed shape» de `design.md`). Después corre la suite completa.

### T1 — Casos del auditor dentro del test y borrado de fixtures (PR 1)

- [ ] Cada criterio de `--audit-session` y cada adaptador se prueban con transcripciones escritas dentro del test, y `tests/fixtures/sessions/` desaparece

**Closes:** AC1

**Depends on:** none

**Locations:** `tests/pilot/session_audit_test.go`, `tests/pilot/*_test.go` que usen `sessionFixture` (incluidos los casos de `adapters/` y `main_thread_cost/grok-sessions`), `tests/fixtures/sessions/`

**Execution:** hive-build-backend

**Test approach:** characterization

**Changes:**
- Una función del test escribe en `t.TempDir()` una transcripción dada como cadena, en el formato de cada host.
- Cada criterio conserva:
  - al menos un caso que falla, con la conducta de su fixture más representativa;
  - un caso por cada otro estado que emite (pass, review).
- Una sesión vacía común cubre `not_observed`.
- Antes de borrar, `go test -coverprofile` sobre `tests/pilot` mide la cobertura de `session_audit*.go` con las fixtures y sin ellas. Las ramas que quedan sin cubrir se listan en el progreso.
- `tests/fixtures/regression/` queda fuera.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; `git ls-files tests/fixtures/sessions | wc -l` da 0; la lista de ramas sin cubrir está en el progreso

### T2 — Modo `--audit-compare` (PR 1)

- [ ] `go run ./tests/pilot --audit-compare --auditor-ref <ref> <sesiones…>` imprime solo los estados que cambian

**Closes:** AC2

**Depends on:** none

**Locations:** `tests/pilot/main.go` (flags), nuevo `tests/pilot/session_audit_compare.go` y su test

**Execution:** hive-build-backend

**Test approach:** tdd

**Changes:**
- Construye el auditor de `--auditor-ref` con `git archive` en un directorio temporal, y el del árbol de trabajo.
- Deduce el host de cada sesión por su ruta: proyecto de Claude Code, rollout de Codex o `updates.jsonl` de Grok. `--host` lo fuerza.
- Corre ambos y compara los estados por criterio.
- Imprime `sesión · criterio · antes → después · líneas`, nunca texto de la sesión.
- El test inyecta dos auditores falsos; no usa sesiones reales.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; con `--auditor-ref 15ac9aa5` sobre el rollout `01a20c02` muestra cambios en las líneas 408, 421 y 496

### T3 — Tope de prioridad heurística y documentación del auditor (PR 1)

- [ ] «Findings» limita a P2 los hallazgos heurísticos construidos por el revisor, y los documentos del auditor nombran la comparación

**Closes:** AC3

**Depends on:** T1, T2

**Locations:** `content/skills/hive-core/close/close-report.md` «Findings», `content/skills/hive-core/review/review-evidence.md`, root `AGENTS.md:78,115`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/README.md`, `README.md:44`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **«Findings»:** un hallazgo contra código que clasifica texto libre por patrones, cuyo fallo solo se muestra sobre una entrada que construyó el revisor y no sobre un archivo, transcripción o registro que existía antes de la revisión, tiene como máximo P2. La excepción es el código en un límite que lista `hive-core/build/security-boundaries.md`.
- **`review-evidence.md`:** nombra esa distinción al confirmar un hallazgo.
- **`AGENTS.md`:**
  - el auditor se prueba con casos dentro del test y su cambio se valida con `--audit-compare`;
  - la fila de checks por ruta de `tests/fixtures/sessions/**` sale.
- La rúbrica y `rule-enforcers.md` lo reflejan.

**Verification:** la comprobación de texto de root `AGENTS.md` «Checks by path» pasa

### T4 — Guía del `plan.md` local (PR 2)

- [ ] La guía distribuida describe el `plan.md` local y no deja ninguna instrucción de carpeta, delta, archivo ni lote de higiene

**Closes:** AC4, AC10

**Depends on:** PR 1 integrado

**Locations:**
- `content/guidance/global.md:24,81`
- `hive-core/plan/change-folder.md` → `plan-record.md`, y cada enlace a `#revise-a-plan-in-progress`: `proportionality.md:9`, `external-effects.md:9`, `scope-changes.md:12`, `hive-core/build/tests.md:5`
- `flow-plan/SKILL.md`; las plantillas de `flow-plan/assets/change-folder/` → `assets/plan.md`
- `flow-plan/references/plan-handoff.md`, `plan-review.md`, `execution-record.md`
- `flow-build/SKILL.md:16,20`, `flow-build/references/spec-delta.md` (se borra), `flow-build/references/e2e-coordination.md`
- `hive-core/build/verification-floor.md:15`
- `flow-verify/SKILL.md:21`
- `flow-close/SKILL.md`, `flow-close/references/batch-archive.md` (se borra), `branch-cleanup.md:7`, `support-material.md:5`
- `hive-core/close/close-hygiene.md`, `close-modes.md`, `close-report.md` (Next y variantes)
- `hive-core/messages/report-endings.md`
- `hive-core/git/foreign-commits.md:7`, `git-commit.md:5`; `git-workflow/SKILL.md`: sale el commit de la carpeta, y el cuerpo del PR suma objetivo, criterios con evidencia, decisiones y el registro de reglas
- `hive-core/plan/checkpoint.md`
- `hive-core/business/business-rules.md:22`, `business-record.md`
- `hive-core/delegation/child-brief.md:7,19`
- `hive-core/ui/viewer.md:8`, `ui-plan.md:10`
- `create-design/references/approval.md:15`, `rule-cards.md:44`, `design-library.md`
- `hive-mode/SKILL.md`, `playbooks/ticket-status.md`
- `project-layout/references/layout.md` (`workspace/<fecha>-<trabajo>/` como carpeta única del trabajo; `sessions/` sale; los tipos de `:50` pasan a `<fecha>-<trabajo>.<tipo>.<ext>`; «no crear un `evidence/`» pasa a «de primer nivel»), `inventory.md`, `archive.md:5,11,13,15`, `script-lifecycle.md:8`, `migrate.md` (salvo donde nombra homes viejos para limpiarlos), `project-layout/SKILL.md`
- `flow-close/references/support-material.md:5,11`; `audit-harness/references/hierarchy.md`
- `audit-harness/SKILL.md:16`, `audit-issues/references/checks.md:22`
- roles `hive-verify-task`, `hive-verify-change`, `hive-ticket-owner`, `hive-review-plan`, `hive-review-ux` (`:12`), `hive-design-ui` (`:11`)
- `audit-tests/SKILL.md:38,44`, `create-design/references/design-record.md:26,42`
- `swarm/SKILL.md`, `references/owner-stops.md`

**Execution:** hive-write-docs

**Test approach:** check

**Changes:**
- El plan va a `<support>/workspace/<YYYY-MM-DD>-<trabajo>/<YYYY-MM-DD>-<trabajo>.plan.md` en los casos de hoy: falla la prueba de reanudación, el usuario lo pide, la entrega abarca varios PRs o repositorios, o el cambio es riesgoso. Si no, queda en la conversación, y arreglos y documentación no llevan archivo.
- El `plan.md` tiene las secciones de «Proposed shape» y no se commitea.
- El brief de un hijo da su ruta absoluta. Para un sandbox sbx, el plan se copia al workspace ignorado del worktree. Manda la copia del checkout principal, y se renueva tras cada revisión.
- Al cerrar, el plan queda como material de apoyo (D23 A).
- El cuerpo del PR lleva lo público.
- Un cambio de contrato edita su spec en el mismo PR.
- La guía ya no nombra lote de higiene, archivo ni `archive-pending/`.

**Verification:** la suite completa pasa (recursos renombrados y borrados); la comprobación de texto pasa; `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa (`assets/**`)

### T5 — Scripts del plan, criterio del auditor y tests (PR 2)

- [ ] `tasks.py` y `check_change.py` trabajan sobre un `plan.md`; `apply_deltas.py` y el criterio `change_folder_uncommitted` desaparecen

**Closes:** AC5

**Depends on:** PR 1 integrado

**Locations:**
- `content/skills/flow-plan/scripts/tasks.py`, `check_change.py`
- `content/skills/flow-close/scripts/apply_deltas.py` (se borra)
- `tests/skills/tasks_test.py`, `change_check_test.py`, `apply_deltas_test.py` (se borra)
- `tests/content/reading_footprint_test.go:151,188,265`
- `tests/pilot/session_audit_specs.go` (criterio `change_folder_uncommitted`) y su test
- `content/skills/project-layout/scripts/plan_layout.py` (`session_dest`, `:719-724`), `check_layout.py` (`sessions` como home), `archive_sessions.py` (sale o pasa a `workspace/`), y `tests/skills/layout_plan_test.py`, `layout_check_test.py`, `workspace_archive_test.py`

**Execution:** hive-build-backend

**Test approach:** tdd

**Changes:**
- **`tasks.py`** recibe la ruta de un `plan.md`:
  - `done`, `checkpoint` y `execution` se conservan; `versioned` y `status` salen;
  - `plan_revision` hashea el archivo;
  - el registro de ejecución:
    - toma `grant_ref` dentro del plan (`<YYYY-MM-DD>-<trabajo>.plan.md#decisiones`);
    - lee los criterios del plan;
    - exige que `plan_ref` sea la ruta del plan;
    - vive fuera de la carpeta del plan;
    - se enlaza desde `## Checkpoint`;
  - `done` sigue buscando el encabezado de progreso (`PROGRESS_HEADING`);
  - la guarda de carpeta integrada sale.
- **`check_change.py`** pierde `--archive`, cuyo único llamador era `batch-archive.md`, y conserva dentro de un archivo la cobertura `AC`/`Closes:`, `Replaces`/`Retired`, `Observed in`, unidades de entrega y checkpoint.
- **Auditor:** el criterio `change_folder_uncommitted` sale con sus casos.

**Verification:** la suite completa pasa (se borra `apply_deltas.py`)

### T6 — Tooling: visor, `hive doctor`, `## Hive` y specs de contrato (PR 2)

- [ ] El visor publica el `*.plan.md` de la carpeta del trabajo, `hive doctor` ya no revisa carpetas de cambio, y `Specs` es opcional

**Closes:** AC5, AC9

**Depends on:** PR 1 integrado

**Locations:**
- `tooling/cli/view_work_scan.go`, `view_work_contract.go` (ayuda en `:52`), `view_work_publish_test.go:55,216,439,539`
- `tooling/cli/doctor_project_changes.go` (se borra), `doctor_project.go`, `project_write.go`, `tui_project_view_test.go`, `project_command_test.go`
- `content/skills/hive-core/plan/project-settings.md`
- `_support/docs/harness-engineering/audit-harness-rules.md:61`
- los specs de contrato `_support/openspec/specs/walkthrough-viewer/spec.md` (publicación) y `versioned-installation/spec.md:199,285` (`Specs`)

**Execution:** hive-build-backend

**Test approach:** tdd

**Changes:**
- **Visor:** `--change DIR` sobre la carpeta del trabajo publica por defecto el único `*.plan.md` que contiene, y se niega con un mensaje si hay más de uno; `--explicit-documents` sigue igual.
- **`hive doctor`:** sus advertencias de carpetas de cambio salen.
- **`Specs`:** pasa a opcional, y `project_write.go` deja de escribirla por defecto.
- **Specs de contrato:** los dos se editan en el mismo PR.

**Verification:** la suite completa pasa (`tooling/**`); la comprobación de texto pasa (`project-settings.md`)

### T7 — Borrado del registro de cambios y páginas (PR 2)

- [ ] `_support/openspec/changes/` desaparece, este plan sigue como `plan.md` local, y las páginas describen el registro nuevo

**Closes:** AC6, AC10

**Depends on:** T4, T5, T6

**Locations:**
- `_support/openspec/changes/`
- `tests/content/spec_deltas_test.go` (se borra)
- `tests/fixtures/regression/README.md`
- `site/src/content/docs/flows/plan.md`, `build.md`, `close.md`, `hive-mode.md`, `status.md`
- `README.md:87,465-470,638,654`, `llms.txt:77`
- `CONTRIBUTING.md`
- `_support/docs/architecture/workspace-and-artifacts.md`, `instruction-resources.md`, `deployment-manager.md`
- `_support/docs/harness-engineering/rule-enforcers.md`, `session-review-rubric.md` (`:11`, filas 52, 56, 65), `_support/docs/architecture/repository-and-distribution.md:19`, `site/src/content/docs/skills.md:21`

**Execution:** hive-write-docs para las páginas y los documentos; el hilo principal para mover este plan y borrar

**Test approach:** check

**Changes:**
- Este plan pasa a `_support/workspace/2026-11-24-process-slimming/2026-11-24-process-slimming.plan.md` con la forma nueva, y `tasks.py` sigue sobre él.
- Los 15 archivos de `_support/sessions/` se mueven a carpetas `_support/workspace/<fecha>-<trabajo>/` locales y salen del árbol.
- `session-review-rubric.md:11` deja de enlazar `_support/sessions/`; `repository-and-distribution.md:19` deja de listarla; `site/src/content/docs/skills.md:21` cambia su frase.
- Se borra la carpeta completa; las 4 carpetas de los PRs #612 a #615 ya integraron.
- `TestSpecDeltasApplied` y su comprobador salen.
- Las páginas y los documentos describen el `plan.md` local.

**Verification:** la suite completa pasa; `pnpm --dir site check`, `lint`, `build` pasan; `git ls-files _support/openspec/changes | wc -l` da 0

### T8 — Specs solo de contrato (PR 3)

- [ ] Quedan los specs de contrato; los de guía salen

**Closes:** AC7

**Depends on:** PR 2 integrado

**Locations:** `_support/openspec/specs/`, root `AGENTS.md:13,134`, `_support/docs/architecture/deployment-manager.md:316`, `tooling/cli/time_command.go:6`, `tooling/viewer/timeline_protocol_test.go:279`

**Execution:** main thread

**Test approach:** check

**Changes:**
- **Se conservan:** `versioned-installation`, `walkthrough-viewer` (con `canonical-references.md`), `documentation-site` y `branch-promotion`.
- **Se recortan a su contrato:** `context-delivery`, `project-layout`, `model-qualification` y `personal-rules`.
- **Se borran:** `workflow-lifecycle`, `local-instructions`, `comment-audit`, `ui-criteria`, `test-accounts`, `audit-issues` y `audit-ledger`. Sus reglas viven en `content/`, y la conducta de sus scripts la fijan sus tests.

**Verification:** la comprobación de texto pasa; la suite completa pasa (`tooling/**`)

### T9 — Migración del registro viejo en `project-layout` (PR 4)

- [ ] La migración de `project-layout` limpia el registro de cambios viejo de un proyecto con manifiesto confirmado

**Closes:** AC8

**Depends on:** PR 2 integrado

**Locations:** `content/skills/project-layout/SKILL.md`, `references/migrate.md`, `references/inventory.md`, `scripts/plan_layout.py` (incluido `:66`), `scripts/check_layout.py`, `tests/skills/layout_plan_test.py`, `layout_check_test.py`

**Execution:** hive-build-backend

**Test approach:** tdd

**Changes:**
- El planificador detecta los homes viejos: `<Specs>/changes/archive/`, carpetas de cambio activas, `sessions/`, `evidence/`, `reports/`, `archive/`, `archived/`, `backups/`, `plan/`, `spec/`, y archivos sueltos o índices dentro de ellos.
- Asigna cada elemento:
  - a `workspace/<fecha>-<trabajo>/`, por su prefijo de fecha y su nombre (sin fecha: la del primer commit; si no hay, la modificación más antigua, como `session_dest`);
  - a `data/`, si es un dump o un export, que nunca se borra sin respuesta;
  - a borrar, si es reproducible;
  - a conservar, si su dueño es incierto.
- Lo versionado sale del árbol con `git rm` en el commit de la entrega de ese proyecto (D28).
- En un `_support/` fuera de git: mueve o borra sin etiqueta, y deja cada carpeta de cambio como «dueño incierto». Un caso de test lo cubre.
- Propone borrar el archivo y cada carpeta activa cuyo último commit es alcanzable desde `origin/<base>`, la prueba que hoy hace `versioned`. Conserva y nombra la que no lo es.
- Etiqueta el repositorio antes de borrar, como la limpieza actual.
- **Pasos de `migrate.md`, no del script:**
  - proponer, de a uno, quitar los specs que solo reescriben guía;
  - listar los registros `archive-pending/` de la memoria para que el usuario los descarte.

**Verification:** `python3 -m unittest discover -s tests/skills -p '*_test.py'` pasa; la comprobación de texto pasa

## Delivery units

### E1 — PR 1: rondas y auditor · depends on: none · files: `tests/pilot/`, `tests/fixtures/sessions/`, `content/skills/hive-core/close/close-report.md`, `review-evidence.md`, `AGENTS.md`, `_support/docs/harness-engineering/`, `README.md`

- [ ] unit: `go test -count=1 ./tests/pilot/` y la comprobación de texto
- [ ] live: `--audit-compare --auditor-ref 15ac9aa5` sobre `01a20c02` · pass: cambios en 408, 421 y 496

### E2 — PR 2: registro del plan y fin del archivo · depends on: E1 · files: `content/`, `tests/skills/`, `tests/content/`, `tests/pilot/session_audit_specs.go`, `tooling/cli/`, `_support/openspec/`, `site/`, `README.md`, `llms.txt`, `CONTRIBUTING.md`, `_support/docs/`

- [ ] unit: la suite completa
- [ ] live: un `plan.md` de ejemplo con `tasks.py checkpoint`, `done` y `check_change.py` · pass: los tres terminan con 0; `hive doctor` sin `Specs` · pass: sin advertencia

### E3 — PR 3: specs · depends on: E2 · files: `_support/openspec/specs/`, `AGENTS.md`, `_support/docs/`, comentarios en `tooling/`

- [ ] unit: la comprobación de texto y `go test ./tooling/cli/ ./tooling/viewer/...`
- [ ] live: n/a: solo borra y recorta documentos

### E4 — PR 4: migración en `project-layout` · depends on: E2 · files: `content/skills/project-layout/`, `tests/skills/layout_plan_test.py`

- [ ] unit: `python3 -m unittest discover -s tests/skills -p '*_test.py'` y la comprobación de texto
- [ ] live: `plan_layout.py` sobre un repositorio temporal con una base, una carpeta cuyo commit está en la base y otra que no · pass: el manifiesto propone borrar el archivo y la primera, y conserva la segunda

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea | El comando de la tarea pasa | Cubierto por la implementación |
| Full suite | Candidata de cada PR | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Code review | Candidata de cada PR, antes del push | `/code-review` de Claude Code | La respuesta de revisión |
| Guía instalada (AC3, AC4) | Tras el merge del PR 1 y del PR 2, y `hive update` | `rg` sobre `~/.agents/skills` | La entrega `automatic` y la actualización de mantenedor |

## Checkpoint

- Updated: 2026-11-24
- Stage: flow-build · PR1 · T1+T2 en paralelo (hijos)
- Last commit: 51f1949d
- Open findings: none
- Next: T1 y T2 en hijos; luego T3 en el hilo principal; PR1
- Rules: flow-build/stage.md
- Principles: none

### Evidence

- before captures: none (no screen)

## Review status and progress

2026-11-24 — Plan escrito y revisado; ver el progreso de `proposal.md`.
