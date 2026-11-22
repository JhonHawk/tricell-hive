# Tasks

Base: rama `docs/inspiration-credits-index` en `657c1ea0` (sobre `development` `da2a4f27`). T1 la ejecuta un hijo `hive-build-backend`, porque es código Go del manager con sus tests y no depende del contenido; las demás corren en el hilo principal, porque dependen del resultado del script de T2.

### T1 — El manager instala las carpetas que declara el manifiesto

- [x] Una skill con `assets/domains.json` instala y valida las carpetas listadas, solo esas además de las cuatro fijas, y un manifiesto inválido rechaza la release nombrando manifiesto y motivo.

**Closes:** AC1, AC7

**Depends on:** none

**Locations:** `tooling/management/plan.go`, `tooling/management/references.go`, un archivo nuevo `tooling/management/skill_domains.go` si la lectura del manifiesto no cabe en `plan.go`, y sus tests en `tooling/management/`

**Execution:** hijo `hive-build-backend`

**Test approach:** tdd

**Changes:** como `design.md` «Manager» fija: leer `assets/domains.json` al recoger una skill desde un checkout; validarlo en `validateRelease` desde los archivos de la release, también para una release guardada; aceptar en `validSource` una carpeta kebab-case de primer nivel; pasar los dominios a `validateInstructionReferences`. Casos de test: un checkout con dos dominios y una carpeta no listada; una release guardada con un dominio; cada forma de manifiesto inválido, `agents` y un dominio vacío incluidos; un archivo fuera de los dominios en una release guardada; estado registrado con una ruta de dominio aceptado; y el retiro de AC7: instalación previa con `hive-core/references/aa.md` y una skill oculta `hive-principle-x`, actualización a `domains.json` con `git/aa.md`, que retira los archivos viejos, el directorio `references/`, la skill y su `agents/openai.yaml`, instala el dominio, y una segunda actualización que no cambia nada. No cambia el código del retiro, que ya es genérico, ni el formato de `stage.json`.

**Verification:** los casos nuevos fallan antes del cambio y pasan después; `go test ./tooling/management/... -count=1` ok.

### T2 — Las hojas y los principios viven en sus dominios

- [x] `content/skills/hive-core/` tiene las 12 carpetas de `design.md`, declaradas en `assets/domains.json`, sin `references/`, sin carpetas `hive-principle-*`, y cada enlace de `content/` apunta a la ruta nueva.

**Closes:** AC2, AC4

**Depends on:** none

**Locations:** `content/skills/hive-core/**`, `content/skills/hive-principle-*/`, cada archivo de `content/` que enlaza una hoja o un principio, `content/skills/*/assets/stage.json`

**Execution:** main thread, con un script local que lee la tabla de dominios

**Test approach:** characterization

**Changes:** `git mv` de cada hoja, de sus carpetas de referencias propias y de cada principio como `design.md` «Content move» fija; quitar el frontmatter de los principios y borrar sus `assets/hive-managed.json`; escribir `assets/domains.json`; reescribir los enlaces; actualizar la descripción y el cuerpo de `hive-core/SKILL.md` sin enlaces a hojas.

**Verification:** `git diff -M --stat HEAD` muestra cada hoja como renombrada y `git status --short` muestra `domains.json` nuevo; `rg -n 'hive-core/references|hive-principle-' content` sin resultados; `wc -c content/guidance/global.md` < 29617.

### T3 — Los enlaces resuelven tras instalar

- [x] El manager instala el contenido real, valida cada enlace con su ancla y genera los paquetes de cada flujo con encabezados `## hive-core/<dominio>/<hoja>`.

**Closes:** AC3

**Depends on:** T1, T2

**Locations:** correcciones que muestre la instalación en `content/`; `TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink`, para que compruebe que se instalan `hive-core/principles/index.md` y una hoja de dominio; `tooling/cli/doctor_project_test.go:451`, que lee la hoja por segmentos (`"hive-core", "references", "project-settings.md"`) y pasa a `plan/project-settings.md`; `tooling/management/stage_packets_test.go:158` si arma una ruta de hoja real; comentarios de `tooling/cli/doctor_project.go` y `tooling/management/voice.go`. Además de `rg 'hive-core/references'`, la búsqueda usa `rg '"hive-core"' --glob '*.go'` para encontrar rutas escritas por segmentos.

**Execution:** main thread

**Test approach:** check

**Changes:** corregir cada enlace que el validador rechace; extender ese test; pasar las rutas armadas por segmentos a las nuevas; actualizar las rutas citadas en comentarios del código.

**Verification:** `go test ./tooling/management/ -run 'TestRealContentInstallsHiveCoreAndResolvesEveryGlobalLink|Stage' -count=1` ok y `go test ./tooling/cli/ -run Doctor -count=1` ok.

### T4 — Los tests de contenido exigen la forma nueva

- [x] `go test ./tests/content/...` pasa, y sus casos de violación fallan por cada forma prohibida de AC5.

**Closes:** AC5

**Depends on:** T2

**Locations:** `tests/content/*_test.go` que nombran `hive-core/references` o `hive-principle-` (`leaf_skill`, `leaf_links`, `single_home`, `retired_names` y los demás que liste `rg`), más `skill_size_test.go` (recorre `content/skills/*/references`) y `reading_footprint_test.go` (clasifica por la subcadena `/references/`); por D9 A, `stage_manifest_test.go` deja de sumar miembros y `tooling/management/stage_packet_margin_test.go` mide cada paquete generado contra 24.000 B menos 500 B

**Execution:** main thread

**Test approach:** check

**Changes:** las invariantes leen la lista de `hive-core/assets/domains.json` y recorren esos dominios en vez de `references/`; los principios se validan como archivos de `principles/`; los casos de violación cubren cada forma de AC5; `retired_names_test.go` gana el prefijo `hive-principle-` como ya tiene `hive-rule-`, y rechaza el texto `hive-core/references/` en los archivos que ya recorre; `skill_size_test.go` recorre también las carpetas de los dominios listados, con su guarda de «nada encontrado»; `reading_footprint_test.go` clasifica las hojas por la lista de dominios; el índice de principios se registra por `principles/index.md`; las rutas de `single_home_test.go` y `leaf_links_test.go` pasan a las nuevas. La invariante de `hive-typescript` no cambia.

**Verification:** `go test ./tests/content/... -count=1` ok, incluido `TestLeafSkillInvariantsCatchViolations` con los casos nuevos.

### T5 — El auditor de sesiones acepta rutas viejas y nuevas

- [x] El auditor reconoce la lectura de una hoja por `hive-core/<dominio>/<hoja>.md` y por sus dos rutas anteriores, y los miembros de paquete `hive-core/<dominio>/<hoja>`.

**Closes:** AC6

**Depends on:** T2

**Locations:** `tests/pilot/session_audit.go`, `session_audit_packets.go`, `session_audit_endings.go`, `session_audit_cost.go` (`hiveSkillPath`, que cuenta las lecturas de Hive), sus tests y un fixture en `tests/fixtures/sessions/`

**Execution:** main thread

**Test approach:** tdd

**Changes:** los patrones de lectura, el prefijo de miembro de paquete y `hiveSkillPath` aceptan cualquier carpeta kebab-case de `hive-core` además de `references`; un caso de sesión lee una hoja por su ruta nueva, y un caso en `session_audit_packets_test.go` cuenta como lectura de Hive una ruta de dominio. Las sesiones grabadas antes del cambio se auditan igual.

**Verification:** el caso nuevo falla antes del cambio del patrón y pasa después; `go test ./tests/pilot/... -count=1` ok.

### T6 — La documentación describe la forma nueva

- [x] `README.md`, `llms.txt`, `AGENTS.md`, `site/`, `_support/docs/` y `CHANGELOG.md` describen las carpetas de dominio y el manifiesto, y no nombran rutas viejas fuera de registros fechados.

**Closes:** AC3

**Depends on:** T1, T2

**Locations:** `README.md`, `llms.txt`, `AGENTS.md`, `site/src/content/docs/**`, `_support/docs/architecture/{instruction-resources,deployment-manager}.md` y los demás de `_support/docs/` que liste `rg`, `CHANGELOG.md`

**Execution:** main thread

**Test approach:** check

**Changes:** reescribir las rutas con el script de T2 donde son rutas, y a mano donde el texto describe la estructura: `instruction-resources.md:33`, `:36` y `:74`, `deployment-manager.md:514` (catálogo: suma `playbooks/` y las carpetas del manifiesto), `README.md:35` y `:47`, `llms.txt:34-35`, `site/…/skills.md:28` y `:34`, `AGENTS.md:18`, `harness-audit/references/skills-and-agents.md:10`; y en `_support/docs/` los documentos vivos `rule-enforcers.md`, `session-review-rubric.md:35` y `:38`, `agent-delivery.md` y `workspace-and-artifacts.md`. Una entrada en `CHANGELOG.md` `[Unreleased]` que diga qué cambia, que es versión minor, que las sesiones abiertas necesitan reiniciar y que el binario se reconstruye antes del contenido. Las secciones publicadas de `CHANGELOG.md` y los archivados de `_support/openspec/changes/archive/` no se tocan.

**Verification:** el comando de AC3 sin resultados; `go test ./tests/content/ -run Retired -count=1` ok.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go test ./... -count=1` and `python3 -I -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Task check | T1 (risky: the release contract of the manager) | One `hive-verify-task` child | Covered by the implementation |
| UI check | n/a: no screen changes | — | — |
| Code review | Final candidate, before the validation stop | Claude Code `/code-review` | Review A |

AC7 needs an installation refreshed from `development`; under `hold` it stays not verified and the final report says so.

## Checkpoint

- Updated: 2026-11-22
- Stage: close · hold pause close
- Last commit: 657c1ea0
- Open findings: 2nd review: hiveCall takes the first installed match (pre-existing, kept for AC6); hive-core/assets|scripts .md counted as reads; domain dir greps uncounted; .md tail unanchored · close findings
- Next: deliver: commit change + domains.json + fixtures on a branch, PR, merge; then rebuild binary before hive update; archive the change folder
- Rules: flow-verify/stage-entry.md
- Principles: none

### Evidence

- go test ./... -count=1 · fails only TestStageManifests · 657c1ea0 + working tree
- python3 -I -m unittest discover -s tests/skills -p '*_test.py' · 412 OK · 657c1ea0 + working tree
- go test ./... -count=1 · ok (after D9) · 657c1ea0 + working tree
- go test ./tooling/management/... -count=1 · ok after the null-manifest fix · 657c1ea0 + working tree
- go vet ./... clean; gofmt -l clean
- CI workflow master-source-branch.yml is CI-only (pull_request_target on master)
- AC1 go test ./tooling/management/... · ok 102 s + hive-verify-task met
- AC2 tree: 12 domains, no references/, 0 hive-principle-*; 78 renames at -M30%
- AC3 rg over content site README llms AGENTS _support/docs tooling · no hits
- AC4 global.md 29210 B
- AC5 go test ./tests/content/... · ok, violation cases present
- AC6 go test ./tests/pilot/... · ok 60.7 s (child run after T2)
- go test ./tests/pilot/... -count=1 · ok 65.6 s after H1 fix

## Review status and progress

2026-11-22 — plan escrito y corregido tras dos rondas de revisión.
2026-11-22 — T2 · uncommitted · unverified: link resolution after install is checked in T3; git diff -M --stat HEAD detects 76 of 78 moves as renames at the default 50% threshold and all 78 at -M30% (review-findings and ui-example have many rewritten links); text outside links is unchanged in all 78
2026-11-22 — T1 · uncommitted · unverified: AC7 on real hosts waits for a refresh from development; the synthetic retirement test passes
2026-11-22 — T3 · uncommitted · unverified: none
2026-11-22 — T5 · uncommitted · unverified: none
2026-11-22 — T6 · uncommitted · unverified: none
2026-11-22 — T4 · uncommitted · unverified: none
