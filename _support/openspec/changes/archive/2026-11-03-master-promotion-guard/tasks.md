# Tareas

Base: `development` en `2653b0b1`. Rama: `chore/master-promotion-guard` en `.claude/worktrees/`. Todas las tareas corren en el hilo principal: son pocos archivos y lecturas de la API de GitHub, y repartirlas costaría más que hacerlas. T4a, T3 y T4b empiezan cuando T1 y T2 ya están integradas en `development`, porque `pull_request_target` toma el workflow de la rama por defecto. El orden es T1, T2, integración, T4a, T3, T4b.

### T1 — Workflow del check de origen

- [x] `.github/workflows/master-source-branch.yml` existe, tal como lo define `design.md`

**Closes:** AC1, AC2

**Depends on:** ninguna

**Locations:** `.github/workflows/master-source-branch.yml`

**Execution:** hilo principal

**Test approach:** check

**Changes:** agrega el workflow. No agrega ningún otro workflow ni ningún test de CI.

**Verification:** correr la condición del paso en local con tres combinaciones de `HEAD_REPO`/`HEAD_REF`/`BASE_REPO`: este repo con `development` → exit 0; este repo con `feat/x` → exit 1; `fork/tricell-hive` con `development` → exit 1. Parsear el YAML con el Ruby del sistema: `ruby -ryaml -e 'YAML.load_file(ARGV[0])' .github/workflows/master-source-branch.yml` → exit 0. En GitHub, la verificación es la de T4a y T4b.

### T2 — Documentación

- [x] `CONTRIBUTING.md` y `deployment-manager.md` explican el check y el ruleset, y `CHANGELOG.md` tiene su entrada

**Closes:** AC4

**Depends on:** ninguna

**Locations:** `CONTRIBUTING.md` (líneas 15 y 67), `_support/docs/architecture/deployment-manager.md:475`, `CHANGELOG.md` (`[Unreleased]`)

**Execution:** hilo principal

**Test approach:** omitido, es un cambio mecánico de documentación

**Changes:**
- **`CONTRIBUTING.md:15` y `deployment-manager.md:475`:** dicen que la única automatización en GitHub es el check de origen hacia `master` y que la verificación sigue corriendo en local.
- **`CONTRIBUTING.md:67`:** suma que un ruleset y el check `source-branch` hacen cumplir la promoción. Un PR de hotfix hacia `master` falla el check y solo entra con el bypass de admin, cuando el maintainer lo decide para ese PR.

`README.md`, `llms.txt` y `content/` no cambian: no mencionan `.github/` ni la CI, y la guía distribuida ya remite a la política de ramas del proyecto.

**Verification:** `rg -n "has no CI|source-branch" CONTRIBUTING.md _support/docs/architecture/deployment-manager.md` muestra el texto nuevo, y ninguna afirmación de que el repositorio no tiene CI.

### T4a — PRs de prueba y lectura del check

- [x] Hay un PR de prueba abierto desde `test/master-guard-probe` hacia `master` con `source-branch` en `failure`, y un PR borrador abierto desde `development` hacia `master` con `source-branch` en `success`

**Closes:** AC1, AC2

**Depends on:** T1 integrada en `development`

**Locations:** GitHub: la rama `test/master-guard-probe` (un commit vacío sobre `development`) y los dos PRs

**Execution:** hilo principal, con la autorización de escritura remota de este work item

**Test approach:** check

**Changes:** abre una rama y dos PRs temporales. No mergea nada.

**Verification:** `gh pr checks <n>` en ambos PRs. Para cada uno, `gh api repos/JhonHawk/tricell-hive/commits/<head>/check-runs` muestra `name == source-branch` y `head_sha` igual al head del PR. El `app.id` de esa lectura es el `integration_id` que usa T3. Si da otro valor que `15368`, T3 usa el leído.

### T3 — Ruleset de master

- [x] Existe un ruleset activo en `master` igual al de `design.md`, con el `integration_id` leído en T4a, y su id está anotado en este archivo

**Closes:** AC3

**Depends on:** T4a (el `app.id` real)

**Locations:** configuración remota de GitHub, no toca archivos del repo; el JSON se escribe en el directorio temporal del host

**Execution:** hilo principal, con la autorización de escritura remota de este work item

**Test approach:** check

**Changes:** crea un ruleset nuevo. No toca `development`, las etiquetas ni la configuración general del repositorio.

**Verification:**
- `gh api repos/JhonHawk/tricell-hive/rules/branches/master` devuelve `deletion`, `non_fast_forward`, `pull_request` (0 aprobaciones, solo `merge`) y `required_status_checks` (`source-branch`, el `integration_id` de T4a, sin strict).
- `gh api repos/JhonHawk/tricell-hive/rulesets/<id> --jq .current_user_can_bypass` devuelve `pull_requests_only`.

### T4b — Efecto del ruleset y limpieza

- [x] El PR de prueba queda bloqueado, el PR desde `development` queda mergeable, los dos se cierran sin mergear y la rama de prueba se borra

**Closes:** AC1, AC2, AC3

**Depends on:** T3

**Locations:** los dos PRs y la rama de T4a

**Execution:** hilo principal, con la autorización de escritura remota de este work item

**Test approach:** check

**Changes:** marca el PR de `development` como listo para que reporte `CLEAN` (un borrador siempre reporta `DRAFT`), lee los dos estados, cierra los dos PRs y borra la rama. No mergea nada.

**Verification:**
- `gh pr view <prueba> --json mergeStateStatus` → `BLOCKED`.
- `gh pr ready <dev>` y luego `gh pr view <dev> --json mergeStateStatus` → `CLEAN`.
- `gh pr close` en ambos, y `git push origin --delete test/master-guard-probe`.
- `gh pr list --base master --state open` queda vacío.

Si el bypass de admin hace que `mergeStateStatus` se lea distinto para el owner, se informa con la lectura literal y AC3 se apoya en la verificación de T3.

## Verificación y revisión humana

| Check | Cuándo y dónde | Evidencia | Autorización |
| --- | --- | --- | --- |
| Tests de cada tarea | Al cerrar la tarea, en local | La verificación de la tarea pasa | Cubierta por la implementación |
| Suite completa | Candidato final, en local | `go vet ./... && go test ./...` (el cambio no toca Go; confirma que nada se rompió) | Cubierta por la implementación |
| Revisión de código | Candidato final, antes del push | Claude Code `/code-review` | Respuesta de revisión |
| Check y ruleset en GitHub | Después de integrar T1 y T2 | T4a, T3, T4b | Escritura remota autorizada |

No hay parada humana: la entrega es `automatic`.

## Estado de revisión y avance

2026-11-03 — plan escrito y revisado por `hive-review-plan`. Se aplicaron H1–H7: T4 se partió en T4a y T4b, el estado `CLEAN` se lee en el PR de `development`, el bypass se comprueba con `current_user_can_bypass`, se suma `deployment-manager.md:475`, el texto del hotfix pasa a `CONTRIBUTING.md`, el id del ruleset entra con el cierre y el YAML se parsea con Ruby.
2026-11-03 — T1 hecha: `a2a9906a`. Check local: development → 0, feat/x → 1, fork → 1, nombre con comillas → 1; el YAML parsea con Ruby. En GitHub falta verificar (T4a).
2026-11-03 — T2 hecha: `61e3aeb5`. CONTRIBUTING (líneas 15 y 67), deployment-manager.md:475 y CHANGELOG.

2026-11-03 — Revisión `/code-review` sobre `61e3aeb5`: 8 hallazgos (H8–H15). En `cbd21614` se corrigieron H9, H12, H13 y H14. H8 se cerró con D6 A (aprobación de workflows de forks, aplicada y leída de vuelta). H10, H11 y H15 quedaron para el cierre; H11 quedó verificado en T4a.
2026-11-03 — Segunda revisión, limitada a `61e3aeb5..cbd21614`: 8 hallazgos, ninguno bloqueante; van a los hallazgos del cierre. `d4021820` (una frase sobre los forks) lleva `Unreviewed:`.
2026-11-03 — Suite `go vet ./... && go test ./...` en verde sobre `61e3aeb5`; gitleaks: 4 commits, sin fugas. PR #345 mergeado en `959e8876`.
2026-11-03 — T4a hecha: #346 (`test/master-guard-probe`) `source-branch` failure; #347 (borrador desde `development`) success. Las dos corridas tienen `head_sha` igual al head del PR, `app.id` 15368 (github-actions).
2026-11-03 — T3 hecha: ruleset `26428778` activo; `rules/branches/master` = deletion, non_fast_forward, pull_request (0 aprobaciones, solo merge; GitHub agregó por defecto `require_extra_approval_for_unattributed_changes: true`), required_status_checks (`source-branch`/15368, sin strict). `current_user_can_bypass` = `pull_requests_only`.
2026-11-03 — T4b hecha: #346 `BLOCKED`, #347 (marcado listo) `CLEAN`. Los dos se cerraron sin mergear y `test/master-guard-probe` se borró en local y en remoto.
