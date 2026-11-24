# Tasks

Base: `development` en `9b9450c5`, rama `feat/run-check-helper` en un worktree de `.claude/worktrees/`. T3 edita más de 3 archivos del auditor y va a un hijo `hive-build-backend`; su interfaz queda fijada en `design.md` («Criterio del auditor», con el campo nuevo `CommandTemplate`). T1, T2 y T4 editan como máximo 3 archivos cada una y quedan en el hilo principal, a la vez que el hijo, porque sus archivos no se cruzan.

Los commits de este cambio nombran la observación que lo originó: Codex 0.170.0, `gpt-6.2-sol` medium, rollout de project-g `01a1f5d4-3ec7-7631-97d8-ba540210ac98`, 2026-11-19, un log de CI local de 1.53 TB.

### T1 — Script `run_check.py`

- [x] El script corre un comando con su salida a un log con tope, imprime el final fuera del log y devuelve el código del comando.

**Closes:** AC1, AC2

**Depends on:** ninguna

**Locations:** `content/skills/hive-rule-checks/scripts/run_check.py` (nuevo), `tests/skills/run_check_test.py` (nuevo)

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** La interfaz y el comportamiento de `design.md`, compatible con Python 3.9. Sin `shell=True`. La prueba usa `sys.dont_write_bytecode`, como las demás.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'run_check_test.py'` pasa, con cada caso visto fallar antes. Casos:
- éxito con código 0 y la ruta impresa primero;
- fallo con su código;
- el log sin copias de su final;
- stderr mezclado en el log;
- salida sin salto final y bytes que no son UTF-8;
- el final acotado a 16 KiB;
- `--grep`;
- comando inexistente (127);
- comando terminado por señal (128+n);
- el tope con un productor sin fin (125, log cerca del tope);
- un nieto en segundo plano que no sobrevive;
- un hijo que ignora `SIGTERM` (escala a `SIGKILL`);
- `--log` sobre un archivo o un enlace existente (2);
- reenvío de `SIGTERM` al grupo.

También se corre con `/usr/bin/python3` (3.9) y a mano sobre `go -C <worktree> test ./tests/content/...` con un test roto a propósito.

### T2 — La hoja exige el helper

- [x] `hive-rule-checks` manda correr con `run_check.py` un check cuya salida se guarda, su referencia explica el uso y el respaldo, e `instruction-resources.md` lo nombra.

**Closes:** AC3

**Depends on:** T1

**Locations:** `content/skills/hive-rule-checks/SKILL.md`, `content/skills/hive-rule-checks/references/run-check.md` (nueva), `_support/docs/architecture/instruction-resources.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** En la hoja, «or save the whole output to a file in the host's temporary location and filter the file» pasa a una frase de igual o menor tamaño que manda usar el helper y enlaza la referencia; también se ajusta «Read an unclear result from that saved output». `${PIPESTATUS[0]}` sigue para un pipe directo. La referencia lleva la interfaz, cuándo usarla, la prohibición de envoltorios que lean su propio log y el respaldo sin `python3`. El marcador de ruta es `<hive-rule-checks-dir>`.

**Verification:** `rg -n 'filter the file' <worktree>/content <worktree>/site` sin resultados; `go -C <worktree> test ./tests/content/...` pasa (tamaño de la hoja ≤ 4,000 bytes, huella de lectura de plan-build por debajo de su línea base y enlaces).

### T3 — Criterio `self_feeding_log` del auditor

- [x] El auditor marca `fail` un comando cuyo bloque redirigido a un archivo lo lee en un `trap` o dentro del bloque, también en un `exec` de Codex interpolado.

**Closes:** AC4

**Depends on:** ninguna

**Locations:** `tests/pilot/trace.go` (campo `CommandTemplate`), `tests/pilot/session_audit.go` (extracción del `exec` de Codex y registro), `tests/pilot/session_audit_logs.go` (nuevo), `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/self_feeding_log/`

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd

**Changes:** La regla de `design.md`. Los fixtures son sintéticos: sin rutas, nombres de proyecto, URLs con contraseñas ni identificadores de la transcripción real, porque el repositorio es público. Fixtures:
- **`fail`:**
  - Codex interpolado con `{ …; } >"$log"` y `trap 'tail "$log"' EXIT`;
  - Codex con `rg` dentro del bloque;
  - Claude con `exec >F` y `cat F` en el `trap`.
- **`not_observed`:**
  - un comando simple con redirección y `trap`;
  - `{ …; } >F 2>&1 || tail F`;
  - la salida en un comando y `tail` en otro;
  - `run_check.py`;
  - el patrón dentro de un heredoc de Python.

Ningún fixture de otro criterio cambia de resultado.

**Verification:** `go -C <worktree> test ./tests/pilot/...` y `go vet ./tests/pilot/...` pasan, con los fixtures vistos fallar antes. Sobre las transcripciones reales:
- `01a1f5d4` da `fail` con la línea 535 entre su evidencia, y ninguna línea que solo tenga comandos simples con redirección;
- `bf678880` (`~/.claude/projects/-Users-me-Development-projects-tricell-tricell-hive/bf678880-6647-425a-bdd7-3f024aa3193a.jsonl`) da `not_observed`.

La salida de los otros criterios sobre `01a1f5d4` no cambia.

### T4 — Enforcers, rúbrica y changelog

- [x] `rule-enforcers.md`, la rúbrica y `CHANGELOG.md` reflejan el helper y el criterio.

**Closes:** AC3, AC4

**Depends on:** T2, T3

**Locations:** `_support/docs/harness-engineering/rule-enforcers.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Test approach:** check

**Changes:** Una fila de enforcer para `self_feeding_log` (`detects`) y una para `run_check.py` (`prevents`); una fila de rúbrica con la sesión `01a1f5d4`; la entrada bajo `[Unreleased]`. La página `site/…/verify.md` no cambia: describe cómo se juzgan los checks, no cómo se corren.

**Verification:** `go -C <worktree> test ./...` pasa (incluye `rule_enforcers_test.go`).

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Auditor over real sessions | Final candidate, locally | `--audit-session` sobre `01a1f5d4` y `bf678880` | Covered by the implementation |
| Code review | Final candidate, before the push | `/code-review` de Claude Code | Delivery `automatic` |
| Security review | Final candidate, before the push | un hijo `hive-review-security` | Seguridad A |

Sin parada de validación (entrega `automatic`). La adopción del helper en sesiones reales se sigue en `watch/self-feeding-check-log`.

## Checkpoint

- Updated: 2026-11-19
- Stage: verify · push
- Last commit: 1aa389b
- Open findings: none
- Next: push, PR, merge, deploy, close
- Worktree: /Users/me/Development/projects/tricell/tricell-hive/.claude/worktrees/run-check · feat/run-check-helper

### Evidence

- before captures: none (no screen change)
- go test ./... via run_check: all ok except the known flaky TestE2ESetupDeadlinesOwnDescendantsAndPipes; tests/pilot rerun ok @b93a6b5a
- python3 -m unittest tests/skills: 294 ok @b93a6b5a
- go vet ./... ok
- local CI: only master-source-branch (PRs to master); CI-only, not applicable
- audit 01a1f5d4 self_feeding_log=fail 13 lines (535 first); bf678880 not_observed; other criteria unchanged vs base
- audit 01a1f5d4 self_feeding_log=fail 13 lines after the token→word rename @ae8be52
- go test ./... via run_check exit 0; unittest tests/skills 296 OK; go vet ok @379e15c
- audit 01a1f5d4 self_feeding_log fail 13 lines (535 first) @379e15c
- go test ./... via run_check exit 0; unittest tests/skills 297 OK; go vet ok @1aa389b
- audit 01a1f5d4 self_feeding_log fail 13 lines; other criteria unchanged @1aa389b

## Review status and progress

2026-11-19 — T1 · 5c5b1b29 · unverified: none
2026-11-19 — T2 · b5af0fcb · unverified: adopción en sesiones reales (watch/self-feeding-check-log)
2026-11-19 — T3 · ae8be52 · unverified: none
2026-11-19 — T4 · b93a6b5a · unverified: none
