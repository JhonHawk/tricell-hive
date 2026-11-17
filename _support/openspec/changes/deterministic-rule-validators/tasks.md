# Tasks

Commit base del repositorio `tricell-hive` al escribir el plan: `f886538f` en la rama `development`; rama de trabajo propuesta: `feat/deterministic-rule-validators`. Las tareas `T1`–`T6` tocan archivos disjuntos con contratos definidos en `design.md` y pueden ejecutarse en el hilo principal o delegarse a roles de construcción (`hive-build-backend`) cuando superen el umbral de delegación; `T7` integra los enlaces e invocaciones en `content/guidance/global.md`, skills, `_support/docs/architecture/` y `site/` una vez que `T1`–`T6` existen.

### T1 — Validar anclas `#fragment` en referencias de instrucciones (`H12`)

- [x] `validateInstructionReferences` en `tooling/management/references.go` valida anclas `#fragment` locales, relativas y `skill:` contra los encabezados Markdown del archivo destino.

**Closes:** AC1

**Depends on:** none

**Locations:** `tooling/management/references.go`, `tooling/management/references_test.go`, `_support/docs/architecture/instruction-resources.md`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Extraer los slugs de encabezados ATX del archivo Markdown destino en `tooling/management/references.go`, validar tanto `#anchor` dentro del mismo archivo como `path.md#anchor` y `skill:owner/path#anchor`, corregir cualquier ancla rota existente en `content/` si el test la detecta, y actualizar `_support/docs/architecture/instruction-resources.md` para reflejar que las anclas `#fragment` ahora se validan en el empaquetado.

**Verification:** `go test -v ./tooling/management -run Test` pasa e incluye casos de prueba para `#fragment` válido, `#fragment` inexistente y colisiones de encabezados repetidos.

### T2 — Validar y gestionar `HIVE.local.md`, `ACCOUNTS.md` y `_support/secrets/` (`H6`)

- [x] `hive doctor --project` inspecciona las invariantes de `HIVE.local.md`, `ACCOUNTS.md` y `_support/secrets/`, y `content/skills/harness-audit/scripts/local_instructions.py` provee subcomandos determinísticos para validar, asegurar `.git/info/exclude`, agregar entradas y marcar `STALE`.

**Closes:** AC2

**Depends on:** none

**Locations:** `tooling/cli/doctor_project.go`, `tooling/cli/doctor_project_test.go`, `content/skills/harness-audit/scripts/local_instructions.py`, `tests/skills/local_instructions_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Añadir en `tooling/cli/doctor_project.go` la comprobación de exclusión en Git, permisos (`0700` en `_support/secrets`, `0600` en archivos secretos salvo `ACCOUNTS.md`), longitud `<= 200` caracteres en líneas de entrada (ignorando encabezados Markdown y párrafo inicial, y descontando `STALE <date> (<cause>):` y la fecha/causa de fallo que gana una entrada `unverified`), fecha `YYYY-MM-DD`, tope de 50 líneas (bloqueante solo si quedan `STALE` sin podar, advertencia de curación cuando todas son vigentes/`unverified`) y ausencia de secretos en `HIVE.local.md` y `ACCOUNTS.md`; crear `content/skills/harness-audit/scripts/local_instructions.py` reutilizando `secret_like`, `SECRET_FLAG`, `ENTROPY` y `MIXED` de `extract_candidates.py` para `check`, `ensure-ignore`, `add` y `mark-stale`.

**Verification:** `go test -v ./tooling/cli -run TestDoctorProject` y `python3 -m unittest -v tests/skills/local_instructions_test.py` pasan.

### T3 — Implementar verificador determinístico de antigüedad de paquetes (≥ 7 días) y advisories (`H7`)

- [x] `content/skills/flow-build/scripts/check_package.py` verifica la fecha de publicación (`>= 7` días) y la ausencia de advisories activos en OSV para paquetes `npm`, `pypi`, `go` y `cargo`.

**Closes:** AC3

**Depends on:** none

**Locations:** `content/skills/flow-build/scripts/check_package.py`, `tests/skills/check_package_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Crear `content/skills/flow-build/scripts/check_package.py` usando únicamente la biblioteca estándar de Python (`urllib.request`, `datetime`, `json`, `argparse`) con soporte para inyección de respuestas en pruebas unitarias, cálculo exacto de antigüedad en días UTC contra el umbral de 7 días y consulta a `api.osv.dev/v1/query`.

**Verification:** `python3 -m unittest -v tests/skills/check_package_test.py` pasa cubriendo paquetes recientes (< 7 días), paquetes válidos (>= 7 días), paquetes con advisory activo y errores de red/versión inexistente, y `python3 content/skills/flow-build/scripts/check_package.py --ecosystem npm typescript` devuelve 0 en vivo.

### T4 — Añadir subcomando determinístico `commit` con preflight de secretos e índice en `git_state.py` (`H8`)

- [x] `content/skills/git-workflow/scripts/git_state.py commit` ejecuta commits atómicos por rutas explícitas preservando el índice previo, escaneando secretos y validando el trailer `Unreviewed:`.

**Closes:** AC4

**Depends on:** none

**Locations:** `content/skills/git-workflow/scripts/git_state.py`, `tests/skills/git_state_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Extender `content/skills/git-workflow/scripts/git_state.py` manteniendo la invocación sin subcomando de solo lectura y añadiendo el subcomando `commit` que rechaza comodines/flags globales o de salto de hooks (`-a`, `-A`, `.`, `--no-verify`, `-n`) y rutas que sean directorios existentes, registra las rutas previamente preparadas en el índice, marca temporalmente con `git add -N` archivos `untracked` incluidos en `<paths>` (revirtiéndolos del índice si el preflight o commit falla), corre `gitleaks` (si está disponible) además del detector de patrones/entropía exclusivamente sobre el diff/contenido de `<paths>`, añade o exige el trailer `Unreviewed: <reason>` cuando se pasa `--unreviewed-fix`, ejecuta `git commit --only -- <paths>` y verifica que las rutas preparadas ajenas sigan intactas en el índice.

**Verification:** `python3 -m unittest -v tests/skills/git_state_test.py` pasa incluyendo pruebas en repositorios Git temporales con rutas preparadas previas, archivos `untracked`, rechazo de directorios/`--no-verify`, bloqueo por secreto detectado y commit con `--unreviewed-fix`.

### T5 — Crear planificador y ejecutor determinístico `cleanup_branches.py` en `flow-close` (`H9`)

- [x] `content/skills/flow-close/scripts/cleanup_branches.py` planifica en JSON (dry-run por defecto) y ejecuta con `--apply` la limpieza segura de ramas y worktrees al cierre.

**Closes:** AC5

**Depends on:** none

**Locations:** `content/skills/flow-close/scripts/cleanup_branches.py`, `tests/skills/cleanup_branches_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Crear `content/skills/flow-close/scripts/cleanup_branches.py` que inspeccione `git worktree list --porcelain`, cambios sin confirmar o sin seguimiento (`git status --ignored --porcelain`), entradas ignoradas permitidas vs. comparación `cmp`/`diff` contra el checkout principal, procesos con `cwd` en el worktree (`lsof -d cwd -Fn`, reteniendo el worktree si `lsof` no está disponible o falla), ramas protegidas o en curso (rama base, `Environments`, base de PR abierto, otro checkout o ramas nombradas en checkpoints/planes con cierre pendiente `< 24h`) y correspondencia de `headRefOid` para ramas fusionadas, emitiendo el plan JSON y ejecutando las eliminaciones solo bajo `--apply`.

**Verification:** `python3 -m unittest -v tests/skills/cleanup_branches_test.py` pasa probando worktrees limpios, worktrees con ignorados idénticos vs. distintos del checkout principal, procesos activos o fallo de `lsof`, ramas protegidas/frescas y ramas con commits posteriores al merge.

### T6 — Añadir validación y splicer de `## Business rules` (`H10`) y validadores de *ready list* para PRD y tickets (`H11`)

- [ ] `check_change.py` y `business_rules.py` validan y actualizan entre marcadores `## Business rules`, y `check_prd.py` / `check_issue.py` validan borradores de PRD y tickets en inglés y español.

**Closes:** AC6, AC7

**Depends on:** none

**Locations:** `content/skills/flow-plan/scripts/check_change.py`, `content/skills/flow-plan/scripts/business_rules.py`, `content/skills/create-prd/scripts/check_prd.py`, `content/skills/create-issue/scripts/check_issue.py`, `tests/skills/change_check_test.py`, `tests/skills/business_rules_test.py`, `tests/skills/check_prd_test.py`, `tests/skills/check_issue_test.py`

**Execution:** main thread

**Test approach:** tdd

**Changes:** Extender `check_change.py` (sin `--archive`) para validar las líneas de `## Business rules` en `proposal.md` (`None:`, `ticket section: <key> · <heading>`, `ticket write: <date> · <moment> · <key>`, `rule: <rule> · screen <name, or none> · <status> · found by <stage>` sin estado `stated:`, `card stale: <screen> · <rule>`, `frame stale: <node> · <rule>`); crear `business_rules.py` con soporte del encabezado localizado (`ticket section:` o `--heading`), inserción al final en la primera escritura, reemplazo entre `<heading>` y `<heading> · end`, rechazo ante marcador huérfano o duplicado y verificación byte a byte del texto exterior; crear `check_prd.py` bilingüe (inglés/español) para validar las 9 secciones, reglas `### R<n> · P<0-2>[(proposed|propuesta)] · <state>`, `Source:`/`Fuente:` y `Verdict:`/`Veredicto:`; y crear `check_issue.py` bilingüe para validar secciones por tipo de ticket, tope de `<= 7` criterios (con `--finding-group`) y formato no vinculante del solution hint (`<= 5` ubicaciones y 1 oración).

**Verification:** `python3 -m unittest -v tests/skills/change_check_test.py tests/skills/business_rules_test.py tests/skills/check_prd_test.py tests/skills/check_issue_test.py` pasa.

### T7 — Actualizar `content/guidance/global.md`, skills propietarios, documentación y pruebas de especificaciones (`D3 A`)

- [ ] `content/guidance/global.md` y los skills propietarios invocan los validadores y scripts de `T1`–`T6` en lugar de recetas manuales de shell, reduciendo el tamaño de `global.md` y manteniendo la paridad con `_support/openspec/`, `_support/docs/architecture/` y `site/`.

**Closes:** AC8

**Depends on:** T1, T2, T3, T4, T5, T6

**Locations:** `content/guidance/global.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-plan/SKILL.md`, `content/skills/flow-plan/references/business-rules.md`, `content/skills/flow-close/SKILL.md`, `content/skills/flow-close/references/branch-cleanup.md`, `content/skills/git-workflow/SKILL.md`, `content/skills/harness-audit/SKILL.md`, `content/skills/harness-audit/references/local-instructions.md`, `content/skills/create-prd/SKILL.md`, `content/skills/create-prd/references/prd-format.md`, `content/skills/create-issue/SKILL.md`, `content/skills/create-issue/references/ticket-types.md`, `_support/docs/architecture/instruction-resources.md`, `_support/docs/architecture/deployment-manager.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/flows/plan.md`, `site/src/content/docs/flows/build.md`, `site/src/content/docs/flows/close.md`, `site/src/content/docs/flows/create-prd.md`, `site/src/content/docs/flows/create-issue.md`, `_support/openspec/changes/deterministic-rule-validators/analysis-and-baseline.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Sustituir las recetas manuales de comandos shell en `content/guidance/global.md` y en los skills propietarios por la invocación a `local_instructions.py`, `check_package.py`, `git_state.py commit`, `cleanup_branches.py`, `business_rules.py`, `check_prd.py` y `check_issue.py` conservando todos los invariantes; actualizar las páginas de `site/` que describen esos flujos; y registrar los números exactos de **Después** (bytes, caracteres, líneas, huella de `TestGuidanceReadingFootprint` y conteos de validadores) en `analysis-and-baseline.md`.

**Verification:** `wc -c content/guidance/global.md` muestra menos de 52,508 bytes, `go test -v ./tests/content -run Footprint` imprime los nuevos valores de lectura, `analysis-and-baseline.md` tiene completas todas las celdas de Después, `python3 content/skills/flow-plan/scripts/check_change.py _support/openspec/changes/deterministic-rule-validators` sale con 0, `go test ./...` y `python3 -m unittest discover -s tests/skills -p "*_test.py"` pasan sin errores.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./... && go vet ./... && python3 -m unittest discover -s tests/skills -p "*_test.py"` | Covered by the implementation |
| Code review | Final candidate, before the push or the validation stop | `Agent: hive-review-code` | `D5 A` |

Parada de validación humana (`D4 A` — `interactive`): antes de `git push`, presentar el diff y tamaños resultantes de `content/guidance/global.md` y los 7 skills junto con la salida de `go test ./...`, `python3 -m unittest discover -s tests/skills -p "*_test.py"` y `hive-review-code` para que el usuario valide en su checkout local.

## Checkpoint

- Updated: 2026-11-16
- Stage: plan · 6
- Last commit: f886538f
- Open findings: none
- Next: /flow-build _support/openspec/changes/deterministic-rule-validators

### Evidence

- `git rev-parse --short HEAD` → `f886538f` (2026-11-16)
- check_change.py: exit 0 (f886538f)
- apply_deltas.py dry-run: 9 MODIFIED requirements valid (f886538f)

## Review status and progress

- 2026-11-16 — Plan redactado para `T1`–`T7` cubriendo `AC1`–`AC8`.
2026-11-16 — T1 · c50d4220 · unverified: none
2026-11-16 — T3 · 6bfb80bf · unverified: none
2026-11-16 — T4 · e3888838 · unverified: none
2026-11-17 — T5 · 6771a316 · unverified: none
2026-11-17 — T2 · 828f2a3e · unverified: none
