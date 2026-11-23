# Tasks

Base `development` @ `6f7fe6f3`, rama `fix/audit-criteria-precision`. T1 va al hijo `hive-build-backend` que escribió los criterios (más de tres archivos en Go y fixtures); T2 en el hilo principal.

### T1 — Criterios de auditoría según el requisito

- [x] Los seis casos de H10 y H11 deciden como el delta dice

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6

**Depends on:** none

**Locations:** `tests/pilot/session_audit_reruns.go`, `tests/pilot/session_audit_test.go`, fixtures en `tests/fixtures/sessions/full_suite_after_text_fix/` y `tests/fixtures/sessions/review_rounds_per_candidate/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** un caso por criterio de aceptación, visto fallar antes del arreglo.

**Verification:** `go test -count=1 ./tests/pilot/` pasa; `--audit-session` sobre `typed2.jsonl` da `pass`.

### T2 — Comando del commit desplegado sin `null`

- [x] `AGENTS.md` usa `// empty` y dice qué hacer sin salida

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

- Updated: 2026-11-23
- Stage: verify · push
- Last commit: f7d46e14
- Open findings: none
- Next: PR, merge, close
- Rules: flow-build/stage.md, flow-verify/stage-entry.md, flow-verify/stage-exit.md

### Evidence

- go test ./tests/pilot/ ok 65.4 s · f7d46e14
- go test ./tests/content/... ok · f7d46e14
- text check ok · 42f10614

## Review status and progress

2026-11-23 — T2 · 42f10614 · unverified: none
2026-11-23 — T1 · 8c562d61 · unverified: AC5 probe typed2 gives pass; real sessions not re-audited
