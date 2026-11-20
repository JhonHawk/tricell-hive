# Tasks

Base: `development` en `8a4443a2`, rama `feat/base-sync-admin-grant` en `.claude/worktrees/base-sync`. T3 edita más de 3 archivos del auditor (loaders de las tres hosts, el bloque nuevo, pruebas y fixtures) y va a un hijo `hive-build-backend`; su interfaz queda fijada en `design.md` («Bloque `main_thread_cost` del auditor»). T1 y T2 editan como máximo 3 archivos cada una y quedan en el hilo principal mientras el hijo trabaja, porque sus archivos no se cruzan. T4 va después de T1–T3 porque documenta lo que dejan; T5 va al final porque usa el bloque de T3.

Los commits nombran la observación que originó el cambio: Grok Build, `grok-4.7-build-fast` high (en el floor como modelo de las sesiones del usuario; los commits dicen por error «outside the floor»), sesión `01a1f5fc-60ea-7971-8d04-ca82a187fa9e` en fleetsystem, 2026-11-19: base de specs 20 commits atrás tras el cierre, `--admin` pedido dos veces y un hilo principal de 599 llamadas.

### T1 — Base local al día en `cleanup_branches.py`

- [x] Con `--apply`, el script hace `fetch` de la base y la avanza con `merge --ff-only` en el checkout principal cuando es seguro, y reporta `base_update` en el plan.

**Closes:** AC1, AC2

**Depends on:** ninguna

**Locations:** `content/skills/flow-close/scripts/cleanup_branches.py`, `tests/skills/cleanup_branches_test.py`

**Execution:** hilo principal

**Test approach:** tdd

**Changes:** Las reglas 1–7 de `design.md` («Base al día»), en ese orden, con el fetch y el avance antes de planear la limpieza y `status`, `diff` y `merge` en el checkout principal. Sin `git pull` ni `rebase`; no cambia la rama de ningún checkout. La salida sigue siendo 0 sin errores y 1 con alguno.

**Verification:** `python3 -m unittest discover -s /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/base-sync/tests/skills -p 'cleanup_branches_test.py'` pasa, con cada caso nuevo visto fallar antes. Casos:
- base detrás y limpia: avanza, `fast_forward` con `commits` y `from`/`to`;
- base al día: `current`;
- lo que llega toca un archivo modificado: `keep` con ese archivo en `paths`, la base no se mueve;
- lo que llega toca un archivo sin seguimiento con el mismo nombre: `keep`;
- lo que llega empieza a seguir un archivo ignorado localmente (como `.env`): `keep` y el archivo sigue igual;
- un archivo modificado que lo que llega no toca: avanza y el cambio local sigue;
- base local con un commit que no está en `origin`: `keep`;
- checkout principal en otra rama: `keep`, la rama no cambia;
- `--repo` apunta a un worktree enlazado: avanza el checkout principal, no el worktree;
- una rama mezclada solo en `origin`: con `--apply` se borra, porque la base avanzó antes de planear;
- sin `--apply` (vía `main(argv)`): no hace fetch ni mueve nada.

### T2 — Permiso de admin-merge por sesión en `hive-rule-merge-gate`

- [x] La hoja deja que un permiso explícito de admin-merge valga el resto de la sesión dentro de los límites de D2 y D4, y el global ya no lo contradice.

**Closes:** AC7, AC4

**Depends on:** ninguna

**Locations:** `content/skills/hive-rule-merge-gate/SKILL.md`, `content/guidance/global.md`, `tests/content/leaf_skill_test.go`

**Execution:** hilo principal

**Test approach:** check

**Changes:** El texto de `design.md` («Permiso de admin-merge por sesión»), el final del punto del revisor humano, la excepción de tamaño con su comentario y las tres frases del global. No reescribe el resto de la hoja. Se quedan las frases de `single_home_test.go`. No cambia `hive-rule-promotion-path`. El pull request da el tamaño de `global.md` antes y después y los lugares revisados (la regla solo vive en la hoja; búsqueda en `content/`, `site/`, `_support/docs/`, `tests/pilot/`, `README.md`, `llms.txt`).

**Verification:** `go -C /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/base-sync test ./tests/content/...` pasa (tamaño de hojas, frases de una sola casa, huella de lectura). Lectura de la hoja y del global contra AC7: las cinco exclusiones, los PR sin checks, la rama de producción definida, el revisor vuelto a pedir y la aprobación vieja están, y «does not cover later ones» ya no. AC4 queda sin verificar hasta una sesión real tras el despliegue.

### T3 — Bloque `main_thread_cost` en `--audit-session`

- [x] `--audit-session` da el desglose del hilo principal para Claude, Codex y Grok.

**Closes:** AC8

**Depends on:** ninguna

**Locations:** `tests/pilot/session_audit.go`, un archivo nuevo `tests/pilot/session_audit_cost.go`, `tests/pilot/session_audit_test.go`, fixtures nuevas en `tests/fixtures/sessions/main_thread_cost/`

**Execution:** `hive-build-backend`

**Test approach:** tdd

**Changes:** La forma de `design.md` («Bloque `main_thread_cost` del auditor»), con la resta de hijos de Grok, los `token_count` repetidos de Codex y las lecturas solo de skills instaladas. El uso se guarda en `auditSession`, no en `traceEvent`; ningún criterio cambia de veredicto. Las fixtures son sintéticas y pequeñas: por host, dos llamadas con uso, una compactación, una lectura de una skill instalada, una de `content/skills/` (que no cuenta), un script de Hive y una lectura de otro archivo; Claude con un mensaje repetido en dos bloques para probar que se cuenta una vez; Codex con un `token_count` repetido; Grok con un hijo cuyo `usage.json` se resta y otro caso donde falta y queda `per_turn_with_children`.

**Verification:** `go -C /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/base-sync test ./tests/pilot/...` y `go -C … vet ./tests/pilot/...` pasan, con los casos nuevos vistos fallar antes. Sobre sesiones reales: Grok `01a1f5fc` da `model_calls` 410 e `input_tokens` 55115725 con `usage_source` `per_turn`; una de Claude y una de Codex dan llamadas y tokens procesados que coinciden con `measure-delegation.py`, salvo los `token_count` repetidos de Codex, contados aparte; los demás criterios de esas tres sesiones no cambian frente a la base.

### T4 — Documentación

- [x] La referencia de limpieza, la rúbrica, `rule-enforcers.md`, el sitio y el changelog dicen lo que cambió.

**Closes:** AC1, AC7

**Depends on:** T1, T2, T3

**Locations:** `content/skills/flow-close/references/branch-cleanup.md`, `site/src/content/docs/flows/close.md`, `_support/docs/harness-engineering/session-review-rubric.md`, `_support/docs/harness-engineering/rule-enforcers.md`, `CHANGELOG.md`

**Execution:** hilo principal

**Changes:** La referencia pide reportar `base_update`, su línea 3 nombra el avance de la base como efecto que autoriza, y deja «return its checkout to the base» solo como volver a la base; el sitio dice que el cierre deja la base al día cuando es seguro y cuándo no, y que un permiso de `--admin` dura la sesión con sus límites (donde describe la retención de un revisor); la rúbrica gana una fila para `main_thread_cost` (sección 5) y el permiso por sesión (Authorization); `rule-enforcers.md` anota el avance de la base como hecho por script. Sin nombres de sesiones privadas en el sitio.

**Verification:** `go -C /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/base-sync test ./tests/content/...` pasa; `rg -n 'grants no other effect|return its checkout' content/skills/flow-close` muestra el texto nuevo.

### T5 — Medición de 3 sesiones

- [x] El reporte final da el desglose de 3 sesiones reales con `main_thread_cost`.

**Closes:** AC6

**Depends on:** T3

**Locations:** ninguna (salida en el reporte)

**Execution:** hilo principal

**Test approach:** check

**Changes:** Corre `--audit-session` sobre fleetsystem Grok `01a1f5fc`, cam Codex `01a1f5d4` y Claude `bf678880`, y da por sesión llamadas, tokens de entrada, contexto por llamada, compactaciones y la parte de lecturas y scripts de Hive. No propone ni cambia reglas; las conclusiones quedan como candidatas para que el usuario decida.

**Verification:** las tres salidas JSON leídas y la tabla en el reporte.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go -C <worktree> test ./...`, `go -C <worktree> vet ./...` and `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |
| Security review | Final candidate, before the push | One `hive-review-security` child on T1 and T2 | Seguridad A |

Sin parada de validación (`automatic`).

## Checkpoint

- Updated: 2026-11-20
- Stage: verify · integrate: push, PR, merge
- Last commit: 6cdd97f3
- Open findings: none
- Next: merge PR, deploy, flow-close
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/base-sync · feat/base-sync-admin-grant

### Evidence

- before captures: none (no screen)
- go -C <wt> test ./... · all ok except tests/pilot TestE2ESetupDeadlinesOwnDescendantsAndPipes (flaky: alone 0/5 on branch, 2/5 on base 8a4443a2); tests/pilot rerun ok · 8b98762c
- go -C <wt> vet ./... · ok · 8b98762c
- python3 -m unittest discover -s <wt>/tests/skills -p '*_test.py' · 309 OK · 8b98762c
- python3 -m unittest discover -s <wt>/tests/skills -p '*_test.py' · 314 OK · 6cdd97f3
- go -C <wt> test ./... and vet ./... · ok · 078c694d

## Review status and progress

2026-11-19 — T1 · 77fea812 · unverified: none
2026-11-19 — T3 · d15a21e0 · unverified: AC8: the Grok session kept running after the plan; 413 calls and 55,650,173 tokens today (602 and 71,338,192 minus the 12 children's 189 and 15,688,019), not the 410 and 55,115,725 measured at planning
2026-11-19 — T4 · 10229e74 · unverified: none
2026-11-19 — T5 · uncommitted · unverified: none
2026-11-19 — T2 · 8b98762c · unverified: AC4: needs a real session after the deploy; AC7 checked on the source, the installed copy is the old text until the deploy

2026-11-20 — Revisiones: /code-review (4 importantes corregidos en f7626f7a, dde37c16, b209989f), hive-review-security (P1 tag origin/<base>, P2 borrado de qa, P2 check sin reportar; corregidos en 734e7c09, 078c694d), segunda /code-review sobre 8b98762c..078c694d (ls-remote por sufijo y fetch tapado por tag corregidos en 6cdd97f3, Unreviewed:). Hallazgos P3 y las decisiones F7 y F9 quedan para el cierre.
