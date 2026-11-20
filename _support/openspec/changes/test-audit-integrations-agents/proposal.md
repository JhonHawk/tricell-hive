# test-audit-integrations-agents — Poda de tests de `integrations/agents`, huecos cerrados y fixes del instalador y del viewer

| Campo | Valor |
| --- | --- |
| Estado | Listo para flow-build |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `d3d2b46c` (el área auditada no cambia desde el baseline `5c49e24d`) · rama propuesta `test/audit-integrations-agents` en el worktree `.claude/worktrees/test-audit-agents`, porque otra sesión usa la checkout principal |
| Playbook | test-audit (skill), sobre `flow-plan` |

## Decisions

2026-11-20
- **D1 B** — ¿Qué área se audita? `integrations/agents` (47 tests, 1,95× líneas de test por línea de producción, sin aserciones sobre llamadas).
- **D2 A** — ¿Dónde queda la prueba de que Render reescribe los links `skill:` en Antigravity? Se agrega `"antigravity"` a la lista de hosts de `tooling/management/skill_links_test.go:62` y se borra `agents_test.go:500`.
- **H1 B** — El CI no corre `go test`. Se incluye en este trabajo un workflow de GitHub Actions que corre la suite en los pull requests.
- **H2 B** — Hay contratos de `Parse` y `tomlQuote` sin ningún test. Se agregan en este trabajo, en el lane A.
- **D3 A** — ¿Qué se hace si en el CI falla en Linux un paquete fuera del área? Si la causa es el entorno del runner o un test que asume macOS, se arregla en este trabajo como commit propio y se nombra en el reporte. Si es un defecto del producto, se para y se trae al usuario como hallazgo.
- **H7 A** — `get-hive.sh` no exigía `gzip`, que GNU tar necesita para leer el paquete. Se arregla en este trabajo.
- **H8 A** — En Linux, `curl` quita el userinfo de `%{url_effective}` y `get-hive.sh` no puede aplicar su regla. Se abre el issue #543 y se salta ese caso del test solo en Linux.
- **H9 A** — `TestRegistryConcurrentProcessesLoseNoEntry` falló en el CI de Linux. Se abre el issue #545 y se re-corre el job. El re-run volvió a fallar.
- **H10 B** — Se investiga y se arregla aquí la carrera del #545: `readStateFileAttempt` reintenta cuando un rename atómico deja sin enlaces el inodo abierto.
- **D4** — El usuario decide quitar el CI del repositorio: basta con correr la suite en local, y aun así se resuelve el test del viewer. Se borra `.github/workflows/tests.yml` y los documentos vuelven a decir que el repositorio no corre tests en CI. **Replaces:** H1 B
- **supuesto** — Se renombra `TestParseRejectsClaudeEffortAsUnknownField` a `TestParseRejectsUnknownFrontmatterKey`; descartado: dejar el nombre, que engaña sobre el contrato.
- **supuesto** — No se quita el export de `Parse`, `ReadProfiles` ni de los tipos `Role`, `Model`, `Host`, `Profiles`; descartado: hacerlo aquí. No son seams solo para tests.
- **supuesto** — Los comentarios de justificación de los tests borrados (parseo de pi-subagents 0.67.0, Antigravity sin `readonly`) pasan a `render_golden_test.go`; descartado: perderlos.
- **supuesto** — No se toca el mensaje de fallo del golden ni la dependencia de `python3` de `agents_test.go:205`; descartado: mejorarlos en esta poda.
- **supuesto** — El lane A va antes que el B, porque es dueño de los helpers compartidos (`editProfiles`, `syntheticProfiles`).
- **supuesto** — El workflow corre los tres comandos de `CONTRIBUTING.md` (`go vet ./...`, `go test ./...`, tests de Python de `tests/skills`) en `pull_request` y en `push` a `development`, con `permissions: contents: read`, acciones fijadas por SHA, Go desde `go.mod` y `umask 022`; descartado: `-race` en cada corrida y `pull_request_target`. **Retired:** D4 quita el workflow.
- **supuesto** — No se cambia el ruleset de GitHub para exigir el check nuevo; descartado: hacerlo en este trabajo, porque es un cambio de configuración fuera del repositorio. **Retired:** D4 quita el workflow.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> "B, y test-audit" (2026-11-20): evaluar si la regla de tests de Hive debe relajarse (respondido en la conversación: no) y correr `test-audit`. Área elegida en D1; hallazgos H1 y H2 incluidos por decisión del usuario.

Sin ticket.

## Objective

Hoy `integrations/agents` tiene 47 tests: 17 repiten contratos que otro test ya prueba, 5 pasan aunque el código esté roto, y varios guards de `ReadProfiles`, `Parse`, `tomlQuote` y `Resolve` se pueden borrar sin que falle ningún test. Al terminar, cada contrato tiene un solo dueño probado por mutación y los huecos tienen test. Durante el trabajo, la suite corrió en Linux y mostró dos defectos que también se arreglan: `get-hive.sh` no exigía `gzip`, y el lector de estado del viewer fallaba con una carrera de reemplazo atómico. El CI que se agregó para correr la suite se quitó (D4): la suite sigue siendo un paso local.

## Scope and acceptance

Dentro: los 5 archivos de test de `integrations/agents`, la fila `antigravity` en `tooling/management/skill_links_test.go`, el fix de H7 (`get-hive.sh`, `tooling/distribution/get_hive_test.go`, `README.md`, `_support/docs/architecture/installer.md`), el salto de H8 en Linux (issue #543), el fix de H10 (`tooling/viewer/state_v3.go`, `tooling/viewer/registry_test.go`), y los documentos que el cambio deja desactualizados: `CHANGELOG.md` (entradas en `[Unreleased]`) y `_support/docs/architecture/agent-delivery.md:27`. `CONTRIBUTING.md` y `deployment-manager.md` conservan su texto (D4). Fuera: `integrations/agents/agents.go` (sin seams que quitar), el CI, el ruleset de GitHub, y las otras instancias del patrón de H10 en `tooling/cli`, que se reportan como hallazgos.

Baseline en `5c49e24d`: 1177 líneas de test y 604 de producción; 47 tests (90 con subtests), todos pasan; lista de fallos del baseline vacía; el repositorio no mide cobertura. Wall time de cada suite que corre tests del área: `integrations/agents` 0,15 s y `tooling/management` 129 s (dueño del keeper de D2).

- AC1 — Cada mutante que hoy pasa todo el paquete hace fallar un test en la aserción de su contrato: el chequeo de esfuerzo inválido de `ReadProfiles` (`agents.go:183`), el error simple en lugar de `*OverrideError` (`:444`), el borrado de los guards de variantes (`:245-247`, `:248-250`, `:253-255`), `offered[0]` en lugar de `nearestOffered` (`:481`), `effort_claude` inválido (`:110`), clave duplicada (`:82`), cabecera CRLF (`:67`), el escape de DEL en `tomlQuote` (`:603`) y `tomlQuote` con `%q` de Go. Valor esperado: para cada mutante G1–G12 falla el test y la fila que nombra la tabla «Planned mutants» de `ledger.md`; fuente: esa tabla, escrita antes del build. · false today because el apéndice de `ledger.md` registra que todos estos mutantes pasan el paquete en `5c49e24d` · Observed in: corrida local de `go test -overlay` sobre el candidato
- AC2 — Los caminos de error de `Parse` sin test hoy (frontmatter ausente, frontmatter sin cerrar, escalar inválido, escalar sin comillas JSON, campo vacío o multilínea, campo obligatorio ausente, nombre distinto del archivo, perfil de modelo desconocido) tienen cada uno una fila en el test dueño de `Parse` que comprueba su mensaje de error, y cada fila falla en esa aserción bajo su mutante P1–P8 de la tabla «Planned mutants» de `ledger.md`. · false today because ningún test del paquete comprueba esos mensajes (`rg` en `integrations/agents/*_test.go`, 2026-11-20) · Observed in: corrida local de `go test -overlay` sobre el candidato
- AC3 — Los 17 tests marcados C o D en `ledger.md` ya no existen, y para cada contrato que cubrían, su mutante K1–K16 o F1 de la tabla «Planned mutants» de `ledger.md` hace fallar el keeper que esa tabla nombra, incluido K10 contra `TestUserScopeRolesAndBlocksLinkInstalledSkillPaths` con el host `antigravity`. · false today because los 17 tests existen y `m_rewrite_after` pasa `tooling/management` · Observed in: corrida local de `go test` y `go test -overlay` sobre el candidato
- AC4 — Un pull request a `development` ejecuta en GitHub Actions un workflow que corre `go vet ./...`, `go test ./...` y los tests de Python de `tests/skills`, sin `continue-on-error` ni `|| true`, y el check aparece en verde en el PR de este trabajo. · false today because `.github/workflows/` solo tiene `master-source-branch.yml`, que no corre tests · Observed in: la corrida del workflow en el pull request de este trabajo en github.com/JhonHawk/tricell-hive (delivery `automatic` abre ese PR) **Retired:** D4 quita el CI del repositorio.
- AC5 — Ningún documento fuera de las change folders dice que el CI no corre tests ni nombra un test borrado; `CONTRIBUTING.md` dice qué comandos corre el CI en cada pull request, y `CHANGELOG.md` tiene la entrada de este cambio en `[Unreleased]`. · false today because `CONTRIBUTING.md:15` y `deployment-manager.md:694` dicen que el repositorio no corre tests en CI, y `agent-delivery.md:27` nombra `TestOpenCodeRolesRequestOfferedVariants` · Observed in: los documentos del candidato **Retired:** D4 deja los documentos con el texto previo sobre el CI; lo reemplaza AC6.
- AC6 — Ningún documento fuera de las change folders nombra un test borrado, `CONTRIBUTING.md` y `deployment-manager.md` siguen diciendo que el repositorio no corre tests en CI, y `CHANGELOG.md` tiene la entrada de este cambio en `[Unreleased]`. **Replaces:** AC5 · false today because `agent-delivery.md:27` nombra `TestOpenCodeRolesRequestOfferedVariants` y `CHANGELOG.md` no tiene la entrada · Observed in: los documentos del candidato
- AC7 — Cuando falta `gzip`, `get-hive.sh` termina antes de cualquier request y su error nombra `gzip`. · false today because el bucle `need_command` de `get-hive.sh:344` no incluye `gzip`, y la fila `missing gzip` de `TestGetHiveNamesAMissingRequirementBeforeAnyRequest` falla en la base · Observed in: `go test ./tooling/distribution` en local
- AC8 — `TestRegistryConcurrentProcessesLoseNoEntry` pasa 5 tandas de 300 corridas seguidas en Linux. · false today because en la base falla en 3 de 3 tandas con `invalid_storage` sobre `registry.json` (mismo inodo, `nlink=0`) · Observed in: un binario de test linux/arm64 corrido en un contenedor local `debian:bookworm-slim`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Ledger de los 47 tests en `ledger.md`, escrito por dos agentes `hive-research` con mutantes vía `go test -overlay` sobre copias; tres afirmaciones revisadas a mano en el hilo principal (`override_test.go:107`, `agents.go:243-256`, `resolve_test.go:127` con `agents.go:183-186`), 2026-11-20.
- Suite completa en local (macOS, `umask 022`, 2026-11-20, HEAD `087efdf0`, mismo código que `5c49e24d` en el área): `go test -count=1 ./...` pasa en 4 min 51 s (los paquetes más lentos: `tooling/cli` 291 s, `tooling/management` 129 s, `tests/pilot` 89 s); los tests de Python pasan en 1 min 12 s.
- `go.mod`: `go 1.27.4`. Repositorio público (`gh repo view`, 2026-11-20), así que GitHub Actions en runners estándar no cobra. **Retired:** solo servía al CI, que D4 quita.
- `CONTRIBUTING.md:15-33`: la suite local es `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'`, con `umask 022`.
- `.github/workflows/master-source-branch.yml`: único workflow; su job `source-branch` es el check que exige el ruleset de `master`, y ningún otro workflow puede usar ese nombre. **Retired:** solo servía al CI, que D4 quita.
- Política de Actions · `gh api repos/JhonHawk/tricell-hive/actions/permissions` · `enabled: true`, `allowed_actions: all`, `sha_pinning_required: false`; permisos por defecto de los workflows `read`; un solo ruleset activo, "master: promotion from development" · 2026-11-20 (leído por la revisión del plan). **Retired:** solo servía al CI, que D4 quita.
- Releases de acciones · `gh api repos/actions/<name>/releases/latest` · `checkout` v7.0.1 (2026-07-20), `setup-node` v7.1.0 (2026-10-08), `setup-python` v7.1.0 (2026-10-22), `setup-go` v7.1.0 (2026-11-17, menos de 7 días: se usa el release anterior) · 2026-11-20. **Retired:** solo servía al CI, que D4 quita.
- Versiones del runner `ubuntu-latest` para Node (los tests de Pi y OpenCode ejecutan `.ts` con Node y necesitan que Node quite los tipos solo, como el Node 24 local), Python (`tomllib` necesita 3.11+) y Go 1.27.4 en `setup-go` · no medido · lo mide T3 en la primera corrida del workflow; el workflow las fija con `setup-node` 24, `setup-python` 3.12 y `go-version-file`. **Retired:** solo servía al CI, que D4 quita.
- La suite completa en Linux · no medido · T3. **Retired:** solo servía al CI, que D4 quita.

## Effects

None: the tasks take no effect on an external system. La reproducción de H10 corrió en un contenedor local `debian:bookworm-slim` que se borra al terminar (`docker run --rm`).

## Delivery

`automatic` (2026-11-20): review, push, PR a `development` y merge sin parada de validación; un approval humano pendiente sigue bloqueando el merge. Code review: Native, Claude Code `/code-review`. Security review: `hive-review-security` sobre el workflow (hecha, sin hallazgos de seguridad) y, tras D4, sobre `get-hive.sh` y `tooling/viewer/state_v3.go`. No autoriza cambiar el ruleset de GitHub ni promover a `master`.

## Review status and progress

2026-11-20 — Plan escrito; revisión `hive-review-plan` lanzada.
2026-11-20 — La revisión encontró 3 P1, 6 P2 y 5 P3. Aplicado:
- tabla «Planned mutants» en `ledger.md` como oráculo de AC1–AC3;
- CRLF como caso positivo;
- `setup-node` 24 y `setup-python` 3.12 en el workflow, y `umask 022` en cada paso;
- `actionlint` fijado vía `go run` y verificación de las acciones con `gh api`;
- docs desactualizados dentro del alcance (T5);
- worktree propio;
- hechos externos con comando y fecha;
- base `d3d2b46c`;
- archivos de T4;
- wall time de `tooling/management`.

Lo que quedaba para el usuario (fallos solo en Linux fuera del área) se decidió en D3 A.
2026-11-20 — Revisión por D4 (el usuario quita el CI): se retiran AC4 y AC5, AC6 reemplaza a AC5, T3 queda revertida y T5 reabierta para devolver los documentos del CI a su texto previo. Los fixes de H7 y H10 quedan dentro del alcance.
