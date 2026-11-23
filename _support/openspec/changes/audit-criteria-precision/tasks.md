# Tasks

Base `development` @ `6f7fe6f3`, rama `fix/audit-criteria-precision`. T1 va al hijo `hive-build-backend` que escribió los criterios (más de tres archivos en Go y fixtures); T2 en el hilo principal.

### T1 — Criterios de auditoría según el requisito

- [ ] Los seis casos de H10 y H11 deciden como el delta dice

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6

**Depends on:** none

**Locations:** `tests/pilot/session_audit_reruns.go`, `tests/pilot/session_audit_test.go`, fixtures en `tests/fixtures/sessions/full_suite_after_text_fix/` y `tests/fixtures/sessions/review_rounds_per_candidate/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** un caso por criterio de aceptación, visto fallar antes del arreglo.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; `--audit-session` sobre `typed2.jsonl` da `pass`.

### T2 — Comando del commit desplegado sin `null`

- [ ] `AGENTS.md` usa `// empty` y dice qué hacer sin salida

**Closes:** AC7

**Depends on:** none

**Locations:** `AGENTS.md`

**Execution:** main thread

**Test approach:** check

**Changes:** `commits[-1] // empty`; sin salida, nombrar los commits de `content/` desde el último merge desplegado que el usuario confirme, en vez de `null..HEAD`.

**Verification:** `echo '[{"consumers":["x"],"commits":[]}]' | jq -r '[.[]|select(.consumers|length>0)][0].commits[-1] // empty'` no imprime nada; el check de texto pasa.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de la tarea pasa | Covered by the implementation |
| Full suite | Reutilizada: las rutas cambiadas tienen check asignado | check de texto y `go test ./tests/pilot/` | Covered by the implementation |
| Code review | Candidato final, antes del push | `/code-review` nativo | The review answer |

## Checkpoint

## Review status and progress
