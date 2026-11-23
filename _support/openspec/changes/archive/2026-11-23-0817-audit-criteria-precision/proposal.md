# audit-criteria-precision — Afinar los dos criterios de auditoría nuevos y el comando del commit desplegado

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #608, merge f550f3e6 |
| Tracker | ninguno (hallazgos H10, H11 y H12 del cierre de session-findings-fixes) |
| Git | tricell-hive · base `development` @ `6f7fe6f3` · rama `fix/audit-criteria-precision` |
| Playbook | bugfix |

## Decisions

2026-11-23
- **H10 B** — ¿Qué hacemos con los casos dudosos de `full_suite_after_text_fix`? Corregir ahora.
- **H11 B** — ¿Qué hacemos con la cuenta de `review_rounds_per_candidate`? Corregir ahora.
- **H12 A** — ¿Qué hacemos con el `null` del comando de `hive releases`? Corregir ahora.
- **H13 A** — ¿Qué hacemos con el fixture de Codex? Ignorar.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

Respuesta del 2026-11-23 a la ronda de hallazgos del cierre de `session-findings-fixes` (PR #607): corregir ahora H10, H11 y H12.

## Objective

Hoy `full_suite_after_text_fix` da `review` cuando una escritura conocida, un `printf` del estado o una tubería dentro de `bash -c` deberían decidir, y cuenta como texto un `.txt` de datos de test; `review_rounds_per_candidate` cuenta dos veces un `/code-review` escrito por el usuario y reinicia la cuenta con `git checkout -- archivo`; y el comando del commit desplegado de `AGENTS.md` imprime `null` sin registro de commits. Al terminar, cada caso decide como el requisito dice.

## Scope and acceptance

Dentro: `tests/pilot/session_audit_reruns.go`, sus tests y fixtures, `AGENTS.md` y el delta de `workflow-lifecycle`. Fuera: el fixture de Codex (H13, ignorado) y cualquier otro criterio.

- AC1 — una escritura de un intérprete con destino conocido de código, en la misma llamada que una corrida y después de ella, da `pass` en el par siguiente · false today because da `review` (fixture `claude-project-python-after-run-review`) · Observed in: `go run ./tests/pilot --audit-session` sobre un fixture nuevo
- AC2 — `go test ./...; printf 'exit=%s\n' $?` con salida `exit=0` cuenta como corrida pasada · false today because la marca guarda el formato y nunca coincide · Observed in: `--audit-session` sobre un fixture nuevo
- AC3 — `bash -c 'go test ./... | tail -3'` cuenta como corrida con tubería, no pasada por el éxito del comando · false today because cuenta como corrida sin tubería · Observed in: `--audit-session` sobre un fixture nuevo
- AC4 — un cambio solo a un `.txt` bajo `testdata/` o `tests/fixtures/` entre dos corridas da `pass` · false today because cuenta como texto y da `fail` · Observed in: `--audit-session` sobre un fixture nuevo
- AC5 — un `/code-review` escrito por el usuario y la llamada al skill que lo ejecuta cuentan una ronda · false today because cuentan dos · Observed in: `--audit-session` sobre `typed2.jsonl` del revisor, que da `pass`
- AC6 — `git checkout -- archivo` no reinicia la cuenta de rondas · false today because reinicia como cambio de rama · Observed in: `--audit-session` sobre un fixture nuevo
- AC7 — el comando de `AGENTS.md` no imprime nada sin registro de commits, y `AGENTS.md` dice qué hacer entonces · false today because imprime `null` · Observed in: el comando corrido con una entrada sin `commits`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `go test -count=1 ./tests/pilot/` · pass, 65,7 s en `2099120c` · 2026-11-23
- Sondas del revisor en el scratchpad de la sesión (`fx/typed2.jsonl`) · `review_rounds_per_candidate` = `review` en `6f7fe6f3` · 2026-11-23

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic` en vigor en la sesión: revisión con `/code-review` nativo (la que `AGENTS.md` prefiere), push, PR y merge a `development`, y refresco local si cambia `content/`.

## Review status and progress

2026-11-23 — carpeta escrita sin revisión de plan (cambio pequeño, como fija el delta sin plan).
