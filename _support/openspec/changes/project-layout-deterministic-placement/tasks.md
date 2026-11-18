# Tasks

- **Base:** tricell-hive `development` en `5a64191b`.
- **Rama:** `feat/project-layout-placement`.
- **Ejecución:** todas las tareas corren en el hilo principal. Ninguna edita más de 3 archivos de código con lecturas grandes fuera de lo ya leído, y el diseño está en este hilo.
- **Clones de prueba:** se crean en el scratch de la sesión con `git clone --local --no-checkout` y `git checkout <commit>`, y nunca escriben en los proyectos.

### T1 — El comprobador ignora dependencias de cualquier lenguaje y falla con código bajo `_support/`

- [ ] `list_files` reemplaza a `walk_files`, y LAYOUT-08 reporta como `fail` el código e IaC versionado bajo `_support/`.

**Closes:** AC16, AC17

**Depends on:** none

**Locations:** `content/skills/project-layout/scripts/check_layout.py`, `tests/skills/layout_check_test.py`

**Execution:** hilo principal

**Test approach:** tdd. Primero las pruebas que fallan:
- una raíz Git con `node_modules`, `.venv`, `vendor` (con `composer.json`) y `target` (con `Cargo.toml`), ignorados, que contienen `.md` bajo `_support/docs`;
- una raíz no Git con las mismas carpetas sin ignorar, más un `bin/` sin manifiesto que sí debe contarse;
- un `_support/scripts/x.sh` ignorado y no registrado que sigue dando LAYOUT-01;
- un `.terraform/` dentro de `_support/infrastructure/` que no se lista;
- un `_support/infrastructure/main.tf` versionado que da LAYOUT-08;
- un `_support/scripts/x.sh` versionado que da LAYOUT-02 y no LAYOUT-08;
- un `_support/sessions/<fecha>-x/run.py` versionado que no da LAYOUT-08;
- un `Layout: infra at infrastructure` que se lee.

**Changes:** cambian la enumeración, la regla nueva, la clave `infra` de `parse_layout` y el docstring. No cambian LAYOUT-01 a 07 ni su severidad.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'layout_check_test.py'` pasa. Además, `python3 -I content/skills/project-layout/scripts/check_layout.py --root /Users/jmartinez/Development/projects/comsys/carvento/carvento-specs` (solo lectura) no lista nada de `node_modules`.

### T2 — `plan_layout.py plan`: destinos y reescrituras por reglas fijas

- [ ] El plan calcula las filas, notas, reescrituras, revisiones y líneas `Layout:` que fija `design.md`, en JSON y en Markdown, con `--override`.

**Closes:** AC18, AC19, AC21, AC22, AC23, AC24

**Depends on:** T1

**Locations:** `content/skills/project-layout/scripts/plan_layout.py` (nuevo), `tests/skills/layout_plan_test.py` (nuevo)

**Execution:** hilo principal

**Test approach:** tdd, con una fixture monorepo que reproduce los casos de carvento:
- `apps/backend/package.json` llama con `../../_support/scripts/backend/x.sh`;
- un workflow llama `_support/scripts/ci.sh`;
- `lefthook.yml` llama un hook;
- un runbook cita `ops.sh`;
- `x.test.sh` sigue a su sujeto;
- una prueba sin sujeto sigue a los candidatos que cita;
- `_support/infrastructure/terraform/main.tf` con `TF_ROOT`;
- `_support/aws/role/policy.json` citado por su `bootstrap.sh`, más un `README.md` que se mueve con el árbol;
- un JSON de IaC sin citas que sale como revisión;
- un `find _support/scripts -name '*.test.sh'` en un workflow;
- un script con `BASH_SOURCE` en una línea y `../../apps/x` en otra;
- citas en `openspec/changes/archive/`;
- `_support/contracts/*.md` citado desde un `.ts`;
- `_support/evidence/2026-01-02-abc-1-x/` y `_support/evidence/ABC-2-2026-03-04-y/`, ignorados;
- `_support/reports/` citado por código no comentado, que sale como revisión;
- `_support/backups/`;
- un runbook dentro de un sitio de docs con `package.json`, que cuenta como `doc`;
- `docs-scripts.sha256.sh`, `node20-check.sh` y `fixture-loader.sh`, que no son de una sola vez;
- un archivo no UTF-8 con un token, que sale como revisión sin tocarse.

Se agregan una segunda fixture idéntica con otros nombres de proyecto, dos corridas comparadas byte a byte, y una fixture con `scripts/deployment/` y `infrastructure/` existentes que deben reutilizarse.

**Changes:** es un script nuevo e importa los ayudantes de `check_layout.py`. No escribe en el árbol en modo `plan`.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'layout_plan_test.py'` pasa.

### T3 — `apply`, `verify` y `undo`: mover, reescribir, comprobar y deshacer sin commit

- [ ] `apply --plan FILE` revalida, mueve (con `git mv` o `os.replace`), reescribe y escribe `<plan>.undo.json`. `verify --plan FILE` falla con citas restantes, herramientas fuera del índice o sin índice. `undo --plan FILE` invierte exactamente lo registrado.

**Closes:** AC20, AC25, AC22, AC29

**Depends on:** T2

**Locations:** `content/skills/project-layout/scripts/plan_layout.py`, `tests/skills/layout_plan_test.py`

**Execution:** hilo principal

**Test approach:** tdd sobre la fixture de T2:
- tras `apply`, `git grep` de cada ruta vieja fuera del historial queda vacío, la ruta relativa en `package.json` resuelve al archivo nuevo y las citas archivadas siguen iguales;
- un origen modificado después del plan hace que `apply` se niegue sin mover nada;
- un destino existente también hace que se niegue;
- `git log` no gana commits;
- una carpeta de evidencia ignorada se mueve entera;
- un archivo CRLF conserva sus finales de línea y un script conserva su modo ejecutable;
- `verify` falla mientras falte una línea del índice, o falte el índice, y pasa cuando se agrega;
- con un cambio local ajeno en un archivo reescrito, `apply` seguido de `undo` deja `git diff` y `git status` idénticos a los previos.

**Changes:** cambia la aplicación del plan. No hay commit, push ni `git add` de archivos ajenos.

**Verification:** `python3 -m unittest discover -s tests/skills -p 'layout_plan_test.py'` pasa.

### T4 — Verificación con datos reales de carvento

- [ ] Se corren cuatro verificaciones y sus resultados quedan en el checkpoint:
  - el plan, `apply`, `verify` y `undo` sobre clones en el scratch de carvento-monorepo en `5f06d0c1` y carvento-specs en `965a73d8`;
  - el plan en seco sobre la raíz real del workspace;
  - el plan en seco sobre el árbol real del monorepo;
  - el check sobre el checkout real de carvento-specs.

**Closes:** AC16, AC17, AC18, AC19, AC20, AC21, AC22, AC23, AC29

**Depends on:** T3

**Locations:** solo el scratch de la sesión; ningún archivo del repo

**Execution:** hilo principal

**Test approach:** check

**Changes:** ninguno en tricell-hive ni en carvento. Las carpetas reales de carvento solo se leen.

**Verification:**
- dos corridas de `plan` dan JSON idénticos (`cmp`);
- las filas de AC19 aparecen con sus destinos;
- AC21 aparece como revisiones y la reescritura de `TF_ROOT` como literal;
- tras `apply` en el clon, `verify` sale 0 salvo las revisiones listadas, y las pruebas `*.test.sh` del clon, encontradas con el descubrimiento ajustado a mano según la revisión de AC21, pasan con `bash`, sin servicios externos;
- el plan sobre la raíz del workspace y sobre el árbol real del monorepo lista las carpetas de AC23 (incluidas `_support/contracts` y `_support/product`) con sus destinos, y no cambia nada (`git status --porcelain` y un listado de `find` iguales antes y después);
- en el clon, con un cambio local añadido a `ci.yml`, `apply` seguido de `undo` deja `git status` y `git diff` iguales;
- el check sobre el checkout real de carvento-specs no lista `node_modules`, mientras se cumpla la precondición de AC16.

### T5 — La convención y la skill usan el planificador

- [ ] Se actualizan los textos de la convención, la skill y sus consumidores para que usen el planificador.

**Closes:** AC28

**Depends on:** T2

**Locations:**
- `content/skills/project-layout/references/layout.md`: vocabulario de destinos y casas, clave `infra at`, LAYOUT-08, y que `Layout:` solo nace de un override.
- `content/skills/project-layout/references/script-lifecycle.md`: el planificador fija el nivel de un script encontrado. Su «never by its file name» cambia explícitamente: el nombre decide solo entre una sola vez y revisión para un script sin llamadores.
- `content/skills/project-layout/SKILL.md`: el modo check y los pasos 2 a 6 del modo organizar corren `plan_layout.py`, copian sus filas, pasan los overrides y aplican con `apply`, `verify` y, si hace falta, `undo`. El «never by its name» del paso 3 queda solo para lo que no clasifica el planificador.
- `content/skills/project-layout/references/migrate.md`: su línea 16 deja de proponer `· tools at <path>` a criterio; una línea `Layout:` nace solo de una negativa o un override.
- `content/skills/workspace-repo/SKILL.md`: el paso 4 acepta los fallos LAYOUT-08 que el usuario dejó fuera del trabajo, o los manda a organizar.
- `content/skills/project-layout/references/inventory.md`, si clasifica scripts.
- `_support/docs/harness-engineering/rule-enforcers.md`.
- `_support/docs/architecture/workspace-and-artifacts.md`.
- La página de skills del sitio que describe project-layout.
- `CHANGELOG.md`.

**Execution:** hilo principal

**Test approach:** check

**Changes:** los textos y su tamaño. No hay reglas nuevas fuera de las que fija `design.md`, salvo los dos cambios explícitos a «never by its name» que permite R9. El tamaño de `content/guidance/global.md` no cambia, y si cambia se declara en el PR.

**Verification:**
- `go test ./tests/content/...` pasa, con tamaños de 16 000 B o menos;
- el escaneo de enlaces del cambio anterior no da enlaces rotos;
- `rg -n 'scripts/automation|walk_files' content/` sale vacío;
- `npm --prefix site run check` pasa, si la página cambió.

### T6 — Detector de rondas en texto, criterio de revisión nativa y watch

- [ ] `additional_tools` cuenta como permiso de la tarjeta; el criterio `review_question_native_option` existe; se registran dos watch en Engram.

**Closes:** AC26, AC27

**Depends on:** none

**Locations:** `tests/pilot/session_audit.go`, `tests/pilot/session_audit_test.go`, `tests/fixtures/sessions/text_round_without_card/` y `tests/fixtures/sessions/review_question_native_option/` (fixtures recortadas, sin datos privados)

**Execution:** hilo principal

**Test approach:** tdd. Primero las fixtures:
- un rollout con `additional_tools`, modo `default`, una ronda en texto y ninguna llamada, que debe fallar y hoy da `not_observed`;
- el mismo rollout en un modo que no es Default ni Plan, que da `not_observed`;
- un título de decisión de revisión sin `/review`, que debe fallar;
- la misma pregunta con «Codex review», que debe pasar;
- un informe de cierre que menciona «revisión de código» sin título de decisión, que da `not_observed`.

**Changes:**
- cambian el parser de Codex y el criterio nuevo;
- en Engram se agregan `watch/review-question-native-option`, con conteo 1, `01a1ed72` y `Enforceable: --audit-session review_question_native_option`, y `watch/external-project-question-format`, con conteo 1, `01a1ed72` y `Enforceable: judgment`.

**Verification:**
- `go -C /Users/jmartinez/Development/projects/tricell/tricell-hive test ./tests/pilot/...` pasa;
- `go run ./tests/pilot --audit-session <01a1ed72> --host codex` da `fail` en los dos criterios;
- sobre `01a1ecec` da `pass` en `review_question_native_option`.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` (known flake `TestE2ESetupDeadlinesOwnDescendantsAndPipes` aside), `python3 -m unittest discover -s tests/skills -p '*_test.py'`, site check | Covered by the implementation |
| Real data | Final candidate, scratch clones of carvento | T4 outputs in the checkpoint | Covered by the implementation (read-only on the projects) |
| Code review | Final candidate, before the push | Claude Code `/code-review` | The review answer in force |

La entrega es `ready`: el PR queda abierto para el merge del usuario. La validación real de AC28 es la próxima sesión de project-layout que el usuario corra en carvento.

## Checkpoint

## Review status and progress
