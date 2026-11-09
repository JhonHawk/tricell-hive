# Tareas

Base `development` 695ca885, rama `feat/qualification-action-checks`. Todas en el hilo principal: el cambio es chico y las tareas se tocan entre sí (la regla de T9 hace fallar la suite real hasta T10).

### T8 — Chequeo `commit_made` y lista cerrada de `compare_none`

- [ ] `commit_made` existe, se valida su ruta y `compare_none` acepta `commit_made` y `answer_must`

**Closes:** AC61

**Depends on:** none

**Locations:** `tests/pilot/qualify_suite.go`, `tests/pilot/qualify_suite_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** `commitMade` y `changedRefs` (compartido con `noNewCommit`); `answer_must_not` sigue fuera de la lista cerrada.

**Verification:** `go test ./tests/pilot/ -run 'TestQualifySuite(CommitMade|ShapeRejections)'` pasa: commit con la ruta cambiada pasa; sin commit, edición sin commit y commit de otro archivo fallan; rutas en `.git` o fuera del fixture se rechazan.

### T9 — Regla «no hace nada» y `no_action_expected`

- [ ] el validador rechaza toda tarea que apruebe `doNothingRun`, salvo `no_action_expected` con `git_unchanged`

**Closes:** AC60

**Depends on:** T8

**Locations:** `tests/pilot/qualify_suite.go`, `tests/pilot/qualify_suite_test.go`

**Execution:** hilo principal

**Test approach:** tdd (la suite real rechazada por la regla antes de T10 es el paso en rojo)

**Changes:** `doNothingRun`, `validateActionRequired`, campo `no_action_expected`.

**Verification:** `go test ./tests/pilot/ -run 'TestQualifySuite(ShapeRejections|RealFileModifiedCopiesRejected)'` pasa: quitar la marca o el `commit_made` de la suite real, o aceptar `not_observed` sin otro chequeo, se rechaza con «does nothing».

### T10 — Suite real, README y delta

- [ ] las siete tareas verifican la acción pedida y las cuatro de pregunta llevan `no_action_expected`

**Closes:** AC62

**Depends on:** T9

**Locations:** `tests/fixtures/qualification/suite.json`, `tests/fixtures/qualification/README.md`, `_support/openspec/changes/qualification-action-checks/specs/model-qualification/spec.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** chequeos de AC62; el spec vigente no se edita: el delta se fusiona al cerrar.

**Verification:** `go test ./tests/pilot/ ./tests/content/` pasa (`TestQualifySuiteRealFileAccepted` y las pruebas de deltas) y el escaneo `rg` del README de los fixtures no imprime nada.

## Verificación y revisión humana

| Chequeo | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar cada tarea, local | El comando de la tarea pasa | Cubierto por la implementación |
| Suite completa | Candidato final, local | `go vet ./... && go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Cubierto por la implementación |
| Code review | Candidato final, antes del push | Claude Code `/code-review` | Respuesta de revisión de la sesión |

Sin validación humana: la entrega es automática.

## Checkpoint

## Estado de revisión y progreso
