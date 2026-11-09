# Tareas

Base `development` 695ca885, rama `feat/qualification-action-checks`. Todas en el hilo principal: el cambio es chico y las tareas se tocan entre sí (la regla de T9 hace fallar la suite real hasta T10).

### T8 — Chequeo `commit_made` y lista cerrada de `compare_none`

- [x] `commit_made` existe, se valida su ruta y `compare_none` acepta `commit_made` y `answer_must`

**Closes:** AC61

**Depends on:** none

**Locations:** `tests/pilot/qualify_suite.go`, `tests/pilot/qualify_suite_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** `commitMade` y `changedRefs` (compartido con `noNewCommit`); las rutas se comparan ya limpias y no pueden ser directorios; `answer_must_not` sigue fuera de la lista cerrada.

**Verification:** `go test ./tests/pilot/ -run 'TestQualifySuite(CommitMade|ShapeRejections)'` pasa: commit con la ruta cambiada pasa; sin commit, edición sin commit y commit de otro archivo fallan; rutas en `.git` o fuera del fixture se rechazan.

### T9 — Regla «no hace nada» y `no_action_expected`

- [x] el validador rechaza toda tarea que apruebe `doNothingRun`, salvo `no_action_expected` con `git_unchanged`

**Closes:** AC60

**Depends on:** T8

**Locations:** `tests/pilot/qualify_suite.go`, `tests/pilot/qualify_suite_test.go`

**Execution:** hilo principal

**Test approach:** tdd (la suite real rechazada por la regla antes de T10 es el paso en rojo)

**Changes:** `doNothingRun`, `validateActionRequired`, campo `no_action_expected`.

**Verification:** `go test ./tests/pilot/ -run 'TestQualifySuite(ShapeRejections|RealFileModifiedCopiesRejected)'` pasa: quitar la marca o el `commit_made` de la suite real, o aceptar `not_observed` sin otro chequeo, se rechaza con «does nothing».

### T10 — Suite real, README y delta

- [x] las siete tareas verifican la acción pedida y las cuatro de pregunta llevan `no_action_expected`

**Closes:** AC63

**Depends on:** T9

**Locations:** `tests/fixtures/qualification/suite.json`, `tests/fixtures/qualification/README.md`, `_support/openspec/changes/qualification-action-checks/specs/model-qualification/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** chequeos de AC63; el spec vigente no se edita: el delta se fusiona al cerrar.

**Verification:** `go test ./tests/pilot/ ./tests/content/` pasa (`TestQualifySuiteRealFileAccepted` y las pruebas de deltas) y `rg -l -i -P '/Users/|/home/|/Volumes/|/private/|~/|password=|secret=|(?<!EXAMPLE_API_)token=' -g '!README.md' tests/fixtures/qualification` no imprime nada.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Code review | Candidato final, antes del push | Claude Code `/code-review` | Respuesta de revisión de la sesión |

Sin validación humana: la entrega es automática.

## Checkpoint

- Updated: 2026-11-08
- Stage: flow-verify · 3 (segunda review limitada al diff de arreglos)
- Last commit: 495d9cd6
- Open findings: R1 P2: commit_made puede penalizar la pregunta de entrega de Hive en el brazo A (no confirmado; al cierre); R3/R5/R8 P3: answer_must del último mensaje y regex débiles (al cierre)
- Next: leer segunda review; push, PR, merge; flow-close

### Evidence

- go test ./tests/pilot/ ./tests/content/ · ok · 7fc5494d
- go vet ./... && go test ./... && python3 -m unittest discover -s tests/skills · ok (83 py tests) · 7fc5494d
- go test ./tests/pilot/ ./tests/content/ · ok · 21b66950
- go test ./tests/pilot/ ./tests/content/ · ok · 495d9cd6

## Estado de revisión y progreso

2026-11-08 — T8 · 7fc5494d · unverified: none
2026-11-08 — T9 · 7fc5494d · unverified: T8–T10 en un solo commit: la regla de T9 hace fallar la suite real hasta T10
2026-11-08 — T10 · 7fc5494d · unverified: ninguna corrida real: los regex de answer_must no se probaron con respuestas de modelos
2026-11-08 — T10 · 21b66950 · unverified: ninguna corrida real: los regex de answer_must no se probaron con respuestas de modelos
