# Tasks

Base `development` en `d3d2b46c`; rama `test/audit-integrations-agents` en el worktree `.claude/worktrees/test-audit-agents`, con esta change folder copiada dentro (otra sesión usa la checkout principal). Los lanes corren uno tras otro en ese worktree (test-audit paso 6): primero A, dueño de los helpers compartidos, después B. Cada lane lo ejecuta un hijo `hive-write-tests`, porque son más de 3 archivos o decenas de tests con evidencia por mutante; el hilo principal corre los checks y hace commit. T3 agregó el workflow de CI, que D4 quitó después. T6 y T7 arreglan los defectos que la suite mostró en Linux. T4 la hace un hijo `hive-review-code` (revisión), y las mutaciones el hilo principal.

En cada lane, el orden es fijo: agregar o extender los keepers con las aserciones trasladadas y aplicar las reparaciones F, correrlos y verlos pasar, y solo entonces borrar los tests que absorben. Nunca se edita un archivo mientras un `go test` lee el mismo checkout. Ningún archivo mutado se commitea.

### T1 — Lane A: keepers, reparaciones, huecos y borrados en render, perfiles y resolución

- [x] Las filas R, F, C y D del lane A de `ledger.md` están aplicadas, con los tests nuevos de H2, y el paquete pasa

**Closes:** AC1, AC2, AC3

**Depends on:** none

**Locations:** `integrations/agents/agents_test.go`, `integrations/agents/resolve_test.go`, `integrations/agents/render_golden_test.go`, `tooling/management/skill_links_test.go` (solo la lista de hosts de la línea 62)

**Execution:** `hive-write-tests`

**Test approach:** characterization: cada aserción nueva o trasladada fija el comportamiento actual, pasa en el baseline y se ve fallar bajo su mutante de `ledger.md` antes de dar la tarea por hecha.

**Changes:**
- Keepers como tablas: `agents_test.go:276` absorbe la fila Cursor de `:298`; `:312` absorbe la fila string de `:305` y, si se arma la tabla de acceso de Cursor, `:319`; `:423` absorbe la fila sin Antigravity de `:450`; `resolve_test.go:19` absorbe las siete filas verifier de `:196` y conserva una lectura del rol `synthetic-verifier`; `resolve_test.go:171` gana el `ReadProfiles` explícito de `:158`.
- Reparaciones F: `resolve_test.go:103` con `Model` literales por host para el perfil `focused`; `resolve_test.go:127` con el esfuerzo `bogus` solo en Claude.
- D2: `"antigravity"` en la lista de hosts de `tooling/management/skill_links_test.go:62`.
- H2: una tabla de rechazos de `Parse` (dueña: `TestParseRejectsInvalidEffortValue`, `agents_test.go:364`) con una fila por camino de error de AC2 más `effort_claude` inválido y clave duplicada, cada una comprobando su mensaje; un test positivo aparte donde la misma fuente con cabecera CRLF y con LF da valores `Role` iguales (`agents.go:67-69` acepta CRLF); el cuerpo del round trip de `agents_test.go:205` gana un DEL y un carácter de control para que `tomllib` rechace un escape inválido.
- Borrados, después de lo anterior: `agents_test.go:115, :158, :236, :257, :298, :305, :381, :402, :450, :500`; `resolve_test.go:158, :164, :181, :196`.
- Renombrar `agents_test.go:372` a `TestParseRejectsUnknownFrontmatterKey`; mover a `render_golden_test.go` los comentarios de justificación de `agents_test.go:123-125` y `:254-256`.
- No cambia: `agents.go`, `render.golden`, los tests R sin carga.

**Verification:** `go test -C <worktree> -count=1 ./integrations/agents ./tooling/management` pasa; los mutantes G1, G8–G12, P1–P8, K1–K14 y F1 de «Planned mutants» en `ledger.md` hacen fallar cada uno el test y la fila que la tabla nombra.

### T2 — Lane B: keepers, reparaciones y borrados en overrides y variantes de OpenCode

- [x] Las filas R, F y C del lane B de `ledger.md` están aplicadas y el paquete pasa

**Closes:** AC1, AC3

**Depends on:** T1

**Locations:** `integrations/agents/override_test.go`, `integrations/agents/opencode_variants_test.go`

**Execution:** `hive-write-tests`

**Test approach:** characterization, con el mismo requisito de mutante por aserción que T1.

**Changes:**
- Reparaciones F: `override_test.go:107` comprueba `*OverrideError` con `errors.As` en cada fila y absorbe la fila `{"claude", ModelOverride{}}` de `:96`; `opencode_variants_test.go:101` conserva la tabla existente y agrega una entrada sin uso por fila ("empty list", "unknown effort", "repeated effort", "model with variant"), comprobando el mensaje de cada guard, y agrega comprobación de mensaje a las filas genuinas; `opencode_variants_test.go:154` usa los esfuerzos `xhigh` y `low` con variantes exactas y absorbe los casos sin override de `:25` (`medium` → `#high`, `xhigh` → `#max`).
- Borrados, después de lo anterior: `override_test.go:96, :144`; `opencode_variants_test.go:25`.
- No cambia: `agents.go`, los tests R.

**Verification:** `go test -C <worktree> -count=1 ./integrations/agents` pasa; los mutantes G2–G7, K15 y K16 de «Planned mutants» en `ledger.md` hacen fallar cada uno el test y la fila que la tabla nombra.

### T3 — Workflow de CI que corre la suite

- [x] Un workflow nuevo corre la suite en cada pull request y en cada push a `development` **Retired:** D4. El workflow se agregó en 3d07cb81 y se quitó en 8d01b652.

**Closes:** ninguno: su criterio se retiró por D4 y el workflow se borró en esta misma rama

**Depends on:** none

**Locations:** `.github/workflows/tests.yml` (nuevo)

**Execution:** main thread; `hive-verify-task` lo verifica, porque agrega dependencias (acciones) y fija permisos.

**Test approach:** check

**Changes:**
- **Triggers y permisos:** `on: pull_request` y `push` a `development`; `permissions: contents: read`.
- **Job:** uno solo en `ubuntu-latest`. No se llama `source-branch`.
- **Acciones:** `actions/checkout`, `actions/setup-go` (con `go-version-file: go.mod`), `actions/setup-node` (Node 24) y `actions/setup-python` (3.12), fijadas por SHA de un release de al menos 7 días. Cada una se verifica con `gh api repos/actions/<name>/releases/tags/<tag>` (fecha), `gh api repos/actions/<name>/git/ref/tags/<tag>` (SHA) y `gh api repos/actions/<name>/security-advisories` (sin advisories activos), porque `check_package.py` no cubre GitHub Actions.
- **Pasos:** `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`. Cada `run:` que corre tests empieza con `umask 022`, porque cada paso abre su propia shell.
- **Timeout:** `timeout-minutes: 30`. La suite local tarda unos 6 min en total, medidos en el baseline antes del build, y el margen cubre runners más lentos.
- **Fallos solo en Linux fuera del área (D3 A):** si la causa es el entorno del runner o un test que asume macOS, se arregla aquí como commit propio y se nombra en el reporte. Si es un defecto del producto, se para y se trae al usuario como hallazgo.
- **No cambia:** `master-source-branch.yml`, el ruleset.

**Verification:**
- `actionlint` por `go run github.com/rhysd/actionlint/cmd/actionlint@<versión>` sale con 0. La versión se fija tras verificarla con `check_package.py --ecosystem go`.
- `rg -n 'continue-on-error|\|\| true' .github/workflows/tests.yml` sale vacío.
- En el PR, el check del workflow termina en verde. Antes, la primera corrida registra las versiones reales de Node, Python y Go.

### T4 — Revisión de preservación y mutaciones

- [x] Ningún contrato perdió su única prueba, y cada contrato que perdió un test, cada reparación F y cada hueco cerrado tiene su mutación registrada en `ledger.md`

**Closes:** AC1, AC2, AC3

**Depends on:** T1, T2

**Locations:** `_support/openspec/changes/test-audit-integrations-agents/ledger.md`, y los archivos de test de T1 y T2 donde se restaure una brecha confirmada

**Execution:** un hijo `hive-review-code` para los dos lanes (comparten el paquete), test-audit paso 7; las mutaciones, el hilo principal, paso 8.

**Test approach:** check

**Changes:** el hijo compara cada test borrado o recortado en `5c49e24d` con su keeper actual y reporta cada contrato sin prueba y cada aserción que no puede fallar; cada brecha confirmada se restaura en su keeper. Para cada mutación: copiar `agents.go` a la ubicación temporal, cambiar la línea del contrato, correr el keeper vía `-overlay` o en el archivo con copia guardada, verlo fallar en la aserción de ese contrato, restaurar y confirmar con `cmp`. Se registra cada mutación y la aserción que la detectó en `ledger.md`.

**Verification:** la sección «Mutaciones del candidato» de `ledger.md` registra, para cada fila de «Planned mutants», el test y la línea que falló; `git -C <worktree> diff --stat -- integrations/agents/agents.go integrations/agent-profiles.json` vacío.

### T5 — Documentos que el cambio deja desactualizados

- [x] Ningún documento nombra un test borrado, los documentos del CI conservan su texto previo, y el CHANGELOG registra el cambio

**Closes:** AC6

**Depends on:** T2

**Locations:** `CHANGELOG.md`, `_support/docs/architecture/agent-delivery.md`, `README.md`, `_support/docs/architecture/installer.md`; `CONTRIBUTING.md` y `_support/docs/architecture/deployment-manager.md` vuelven al texto de la base

**Execution:** main thread

**Changes:** `agent-delivery.md:27` nombra el keeper que reemplaza a `TestOpenCodeRolesRequestOfferedVariants`. `CONTRIBUTING.md` y `deployment-manager.md` conservan el texto de la base, que dice que el repositorio no corre tests en CI (D4). `README.md` y `installer.md` nombran `gzip` entre los requisitos del instalador (H7). `CHANGELOG.md` gana entradas en `[Unreleased]`: las pruebas nuevas y la auditoría, y los fixes de `get-hive.sh` y del viewer. `llms.txt` no cambia. **Replaces:** the earlier T5 changes, which documented the tests workflow

**Verification:** `rg -n 'TestOpenCodeRolesRequestOfferedVariants|tests\.yml' <worktree>` sin resultados fuera de `_support/openspec/changes/`; `git diff origin/development -- CONTRIBUTING.md _support/docs/architecture/deployment-manager.md` vacío.


### T6 — `get-hive.sh` exige `gzip` (H7)

- [x] El instalador nombra `gzip` cuando falta, antes de cualquier request

**Closes:** AC7

**Depends on:** none

**Locations:** `get-hive.sh`, `tooling/distribution/get_hive_test.go`, `README.md`, `_support/docs/architecture/installer.md`

**Execution:** main thread

**Test approach:** tdd: la fila `missing gzip` se vio fallar en la base (macOS) y pasar con el fix. El salto de H8 en Linux, con el issue #543, va en el mismo archivo de test.

**Changes:** `gzip` entra en el bucle `need_command` y en la lista `needed` del test. README e `installer.md` lo nombran entre los requisitos.

**Verification:** `go test ./tooling/distribution -run 'TestGetHiveRejectsDisallowedRedirectHosts|TestGetHiveNamesAMissingRequirementBeforeAnyRequest'` pasa.

### T7 — El lector de estado del viewer reintenta cuando un reemplazo atómico cae después del open (H9, H10)

- [x] Los registros concurrentes ya no fallan con `invalid_storage` en Linux

**Closes:** AC8

**Depends on:** none

**Locations:** `tooling/viewer/state_v3.go`, `tooling/viewer/registry_test.go`

**Execution:** main thread

**Test approach:** tdd. El test concurrente que ya existía es el dueño. Ahora muestra la salida de cada helper, y se vio fallar en Linux antes del fix (3 de 3 tandas de 300) y pasar después (5 de 5). Un test determinista necesitaría un hook solo para tests entre `Open` y `Stat`, que la regla de tests rechaza.

**Changes:** `readStateFileAttempt` reintenta, dentro de sus 3 intentos, cuando el inodo abierto quedó con `nlink` 0. Los enlaces, los archivos que no son regulares y los demasiado grandes se siguen rechazando.

**Verification:** el binario de test linux/arm64 de `tooling/viewer`, corrido en un contenedor `debian:bookworm-slim` con `-test.run '^TestRegistryConcurrentProcessesLoseNoEntry$' -test.count=300`, pasa en 5 de 5 tandas; `go test ./tooling/viewer` pasa en macOS.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | El comando de verificación de la tarea pasa | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'` con `umask 022` | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer |
| Security review | Final candidate, before the push | `hive-review-security` sobre el workflow (hecha antes de D4) y sobre `get-hive.sh` y `tooling/viewer/state_v3.go` | The security review answer |

## Checkpoint

- Updated: 2026-11-20
- Stage: verify · security review of fixes, then push and merge
- Last commit: 8d01b652
- Open findings: H11 P3 pre-existing: readPrivateWorkFile and readSafeMapFile in tooling/cli reject a file replaced after open, same pattern as #545 · close findings
- Next: wait hive-review-security on get-hive.sh and state_v3.go, commit folder, push, merge
- Worktree: /Users/jmartinez/Development/projects/tricell/tricell-hive/.claude/worktrees/test-audit-agents · test/audit-integrations-agents

### Evidence

- baseline: go test -count=1 ./... pass 4m51s; python unittest pass 1m12s; HEAD 087efdf0 (area identical to d3d2b46c)
- local CI at d973aeaf: go vet ./... ok; go test -count=1 ./... ok 6m7s; python unittest 314 OK; f14303da changes only comments (agents pkg ok, actionlint 0)
- review round: /code-review 3 blocking fixed in 0cd9b573 (fetch-depth, concurrency group, CHANGELOG heading); hive-review-security no security finding; re-review of 0cd9b573 no findings; merged origin/development 3e3102e4, tests/content + agents ok
- full local suite at 8d01b652: go vet ok; go test ./... ok 4m51s; python 314 OK. Earlier evidence (d973aeaf/0cd9b573) covers the earlier candidate only
- fix review of 91f68c56 0978f5f2 bc60e226 21dac832: no findings; Linux skip accepted

## Review status and progress

2026-11-20 — T3 · d80fb0f7 · unverified: AC4: green run on this work's PR (opens at delivery); hive-verify-task passed everything checkable from the file and the API
2026-11-20 — T1 · 71a26acc · unverified: none; K7 caught by TestRenderGolden and agents_test.go:80, not by resolve_test.go:47 as planned (Render-only mutant); T4 re-runs every mutant
2026-11-20 — T2 · 3da9b3fd · unverified: none; G6 low row does not catch offered[0] (xhigh and no-override rows do); K15 fails management at model_override_test.go:149 before :153
2026-11-20 — T5 · d83c615a · unverified: none
2026-11-20 — T4 · d973aeaf · unverified: none
2026-11-20 — T5 · 8d01b652 · unverified: none
2026-11-20 — T6 · 91f68c56 · unverified: none; H8 skip 21dac832 with issue #543
2026-11-20 — T7 · 0978f5f2 · unverified: none; Linux evidence from a local container, not CI (D4)
