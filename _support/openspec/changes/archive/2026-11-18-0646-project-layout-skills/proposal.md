# project-layout-skills — Una casa para la convención de layout y una skill que crea el repo del workspace

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #508, merge 88864d84 |
| Tracker | Sin ticket |
| Git | tricell-hive · base `development` en `9f67fead` · rama propuesta `feat/project-layout-skills` |

## Decisions

2026-11-17
- **Dos layouts** — ¿Monorepo o workspace? Ambos son válidos: monorepo cuando pasa el test de viabilidad de cuatro puntos, workspace de varios repos en otro caso (decisión previa, `engram:tricell-hive/decision/monorepo-or-workspace-layout`).
- **Dos skills** — ¿Cómo se reparte el trabajo? Una skill evoluciona docs-cleanup: establece la convención, organiza y migra, y funciona sin convertir el workspace en repo. Otra skill nueva crea el repo coordinador del workspace: primero evalúa qué hace falta y luego decide qué partes de la primera correr. Por proyecto: fleetsystem recibe el repo coordinador; cam y carvento no, solo la primera skill; facti queda sin decidir.
- **D15 A** — ¿Cómo se llama la skill nueva? `workspace-repo`.
- **D16** — ¿Dónde vive el conocimiento vigente de un repo? Se propone migrarlo a `_support/docs`; si el usuario no lo acepta, se conserva la casa que el repo ya usa, una sola por repo.
- **D17 A** — ¿Hay carpeta para los datos de operaciones? Sí: `_support/data/<fecha>-<trabajo>/`, nunca versionada y protegida en la limpieza.
- **D18 A** — ¿Cuánto de harness-audit se desintegra aquí? Solo el layout: sus checks de layout, el manifiesto de migración y la pregunta del hub pasan a la skill nueva. El resto de harness-audit se divide en skills especializadas en un cambio propio, planeado justo después, porque ya corrió sobre cam y carvento sin detectar el problema de fondo.
- **Línea `Layout:` del índice** (2026-11-18, respuesta «A») — ¿Cómo se conoce en runtime la excepción de un proyecto que no acepta la migración de D16? Una línea `Layout: docs at <path> · tools at <path>` en la cabecera del índice declara sus casas propias; manda sobre los destinos por defecto, el check la respeta y el modo migrar la escribe cuando el usuario declina.
- **D19 A** — ¿La skill nueva absorbe la regla de dónde va cada artefacto? Sí: `flow-close/references/artifact-layout.md` pasa a ella y docs-cleanup se renombra `project-layout`, para que escribir, migrar y limpiar lean una sola versión.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> Primero la skill que evoluciona de docs-cleanup, esto migraria, organizaria y estableceria la convención. Nueva skill para crear el workspace-repo, esto lo que haria es 1) ver lo que se necesita para crear este repo 2) validar que partes de docs-cleanup vale la pena ejecutar. Pero la primera puede ejecutarse sin convertirse en un repo de workspace, por ejemplo, no quiero hacer para cam ni carvento (lo de crear el repo), pero si para fleetsystem.

Antes, en la misma sesión (2026-11-17): «no es solo sobre scripts, si no todos los docs que puedan estar dentro/fuera de _support, considerando tambien el tema de workspaces (por ejemplo la mayoria de scripts irian en el backend), monorepos, tambien tomando en cuenta por ejemplo un script que se hace para un ticket o tarea especifico, cómo puede ser promovido a una tool interna […] carpetas a nivel raiz, archivos que funcionen como indices para de una sola leida saber lo que se tiene, que cosas se va a las carpetas de sesión, como se depuran».

## Objective

Hoy la convención de layout vive en tres casas: dónde va un artefacto nuevo (`flow-close/references/artifact-layout.md`), los checks de layout y el manifiesto de migración (`harness-audit/references/project-layout.md`) y la limpieza (`docs-cleanup`). Ninguna dice dónde va un script que necesita cada clon, uno que se corrió una vez contra datos compartidos, los datos de una operación, ni cómo se encuentra de una lectura lo que un proyecto tiene. El inventario del 2026-11-17 lo muestra: cam deja 5 grupos de scripts sin registrar en `_support/scripts/`, carvento versiona 46 scripts de CI y su Terraform bajo `_support/`, facti versiona 86 scripts bajo `_support/scripts/` en 3 repos y su código de aplicación lee y escribe `_support/`, y unos 540 scripts quedan dentro de `workspace/` y `sessions/`.

Al terminar, una skill `project-layout` es la única casa de la convención (raíz, índice, `_support/`, código fuera de `_support`, ciclo de vida de un script, depuración), la verifica con un script sin juicio, y organiza o migra un proyecto con un manifiesto que el usuario confirma. Una skill `workspace-repo` convierte un workspace de varios repos en un repo coordinador que versiona su conocimiento, usando `project-layout` para lo que haya que ordenar.

## Scope and acceptance

Dentro:
- Renombrar `docs-cleanup` a `project-layout` y ampliarla: convención, check, organizar y migrar.
- Mover `flow-close/references/artifact-layout.md` a `project-layout` y actualizar sus 15 consumidores.
- Mover a `project-layout` los checks HA-PL y el manifiesto de migración de harness-audit, que queda con un puntero.
- Skill nueva `workspace-repo`, con un script que planea y aplica la creación del coordinador.
- La excepción de `content/guidance/global.md` a «do not initialize Git» para el coordinador.
- Que `hive doctor` y el viewer sigan funcionando con un coordinador.
- Documentación de arquitectura, `site/`, `README.md`, `llms.txt` y el artifact «Skills de Hive».

Fuera:
- Dividir el resto de harness-audit (D18 A): cambio propio, justo después.
- Aplicar las skills a cam, carvento, fleetsystem o facti: se hace en la sesión de cada proyecto, con las skills instaladas.
- Entregar la migración pendiente de carvento (`monorepo-knowledge-consolidation`), que sigue en pausa por instrucción del usuario.
- Pilotos de comportamiento, pausados.

- AC1 — La regla de dónde va cada artefacto vive solo en `content/skills/project-layout/references/layout.md`: `rg -l 'artifact-layout' content site/src` no devuelve nada y `global.md` y los demás consumidores enlazan a la nueva ruta · false today because 15 archivos enlazan `flow-close/references/artifact-layout.md` · Observed in: el árbol `content/` y `go test ./tests/content/...`
- AC2 — La convención nombra un destino para cada uno de estos casos: un script que necesita cada clon o CI (el código del repo dueño, fuera de `_support`), un script corrido una vez contra datos compartidos (su carpeta de sesión con salida y reversión), los datos de una operación (`_support/data/`), el índice del proyecto, y prohíbe que el código dependa del material de trabajo bajo `_support/` (sí puede leer `_support/docs` y `_support/openspec`) · false today because `artifact-layout.md` no nombra ninguno de esos casos · Observed in: `content/skills/project-layout/references/layout.md`
- AC3 — La convención define el ciclo de vida de un script en niveles (desechable, corrido una vez, candidato, herramienta interna, mecanismo) con el criterio que sube cada nivel, y la sección «Candidates» de flow-close enlaza a ella · false today because no hay ciclo de vida y «Candidates» no enlaza a ninguna convención · Observed in: `layout.md` (o su referencia hermana) y `content/skills/flow-close/SKILL.md`
- AC4 — `check_layout.py` sobre fixtures imprime un hallazgo con su corrección por cada uno de: un script de `_support/scripts/` que ningún `HIVE.local.md` nombra, un archivo versionado bajo `_support/scripts`, `secrets`, `workspace` o `data`, un archivo de código versionado fuera de `_support/` (por extensión, como `design.md` «check_layout.py» define) que en una línea que no es comentario cita material de trabajo bajo `_support/` (no `docs/` ni `openspec/`), una entrada de primer nivel de `_support/` fuera de la convención, un documento de la casa de docs ausente del índice, un enlace roto del índice, y en un coordinador un repo de su tabla ausente, con otro remoto, no ignorado o no reincluido en `.ignore`; sale con código distinto de 0; sobre un fixture limpio, y sobre un `.gitignore`, `.ignore`, `AGENTS.md` o Markdown que citan `_support/`, un script que lee `_support/openspec` y un comentario que cita `_support/scripts`, sale 0 · false today because el script no existe · Observed in: `tests/skills/layout_check_test.py` · **Retired:** la regla 3 pasó a revisión tras `/code-review`; lo reemplaza AC15
- AC15 — `check_layout.py` sobre fixtures imprime un hallazgo con su corrección por cada uno de: un script de `_support/scripts/` que ningún `HIVE.local.md` nombra exactamente, un archivo versionado bajo `_support/scripts`, `secrets`, `workspace` o `data`, un documento de la casa de docs ausente del índice, un enlace roto del índice, y en un coordinador un repo de su tabla ausente, con otro remoto, no ignorado o no reincluido en `.ignore`, y sale con código distinto de 0; reporta para revisión, sin cambiar el código de salida, un archivo de código versionado fuera de `_support/` que no es test y que en una línea que no es comentario cita material de trabajo bajo `_support/`, y una entrada de primer nivel de `_support/` fuera de la convención; sobre un fixture limpio sale 0 sin hallazgos, y no reporta un `.gitignore`, `.ignore`, `AGENTS.md` o Markdown que citan `_support/`, un script que lee `_support/openspec`, un comentario que cita `_support/scripts`, un test, ni enlaces del índice con `<…>`, `%20`, otro esquema o ruta del sitio · false today because el script no existe · Observed in: `tests/skills/layout_check_test.py` · **Replaces:** AC4
- AC5 — Corrido en modo lectura sobre los checkouts locales reales, `check_layout.py` reporta lo que el inventario del 2026-11-17 encontró (lista en «Verified context»): los 5 grupos sin registrar de `cam/_support/scripts/`, los 46 archivos versionados de `carvento-monorepo/_support/scripts/`, los 65 de `facti-2020-new/_support/scripts/` y los 16 de `facti-backend-client/_support/scripts/`, `server/scripts/cancel-payment-complements.js` de facti-2020-new citando `_support/`, y `evidence/` y `migration-data/` de fleetsystem como entradas fuera de la convención, sin hallazgo de scripts no registrados en fleetsystem · false today because el script no existe · Observed in: esos checkouts locales, sin escribir en ellos
- AC6 — harness-audit ya no contiene checks de layout: `content/skills/harness-audit/references/project-layout.md` no existe, su paso 4 apunta a `project-layout`, y las reglas HA-PL viven en `project-layout` · false today because harness-audit las contiene · Observed in: el árbol `content/`
- AC7 — El modo migrar de `project-layout` dice que, cuando `Specs` o el hub de un workspace no está claro, pregunta con cuatro opciones (repo de specs detectado, repo dueño de los contratos, repo coordinador con `workspace-repo`, solo `_support` local), propone mover la casa de docs a `_support/docs` y conserva la existente si el usuario no acepta, y aplica el test de viabilidad antes de recomendar monorepo · false today because HA-PL-02 solo dice que el caller decide · Observed in: `content/skills/project-layout/`; la conducta en una sesión real queda pendiente de observación
- AC8 — Sobre un workspace de prueba con dos repos hijos y un repo de specs, `workspace_repo.py` sin `--apply` no escribe nada e imprime el plan; con `--apply` deja un coordinador donde `git -C <raíz> status --short` no lista contenido de los hijos, `rg <marca>` desde la raíz encuentra un archivo de un hijo, y `check_layout.py` sale 0 · false today because el script no existe y `rg` desde una raíz que ignora a sus hijos no encuentra nada (observado 2026-11-17) · Observed in: `tests/skills/workspace_repo_test.py`
- AC9 — `global.md` permite inicializar Git en la raíz de un workspace declarado solo mediante `workspace-repo` y a pedido del usuario · false today because `global.md:78` dice «do not initialize Git» sin excepción · Observed in: `content/guidance/global.md`
- AC10 — Sobre la salida real de `workspace_repo.py --apply` con un `HIVE.local.md` presente, `hive doctor` en la raíz no reporta ningún hallazgo bloqueante; y la clave de proyecto del viewer (`timelineProjectKey`) de un workspace con capturas ya registradas no cambia al crear el coordinador sin remoto, y la skill avisa antes de agregar `origin`, que la cambia y devuelve `project_identity_changed` · false today because ningún test cubre un coordinador y `workspace-repo` no existe · Observed in: `go test ./tooling/...` y `tests/skills/workspace_repo_test.py`
- AC11 — `README.md`, `llms.txt`, `AGENTS.md`, `site/src/content/docs/skills.md` e `index.mdx` nombran `project-layout` y `workspace-repo` y no `docs-cleanup`, y `rg -l docs-cleanup` fuera de la historia (CHANGELOG salvo su entrada nueva, cambios archivados, sesiones, y la regex de transcripts de `tests/pilot/session_audit.go`) no devuelve nada · false today because nombran `docs-cleanup` · Observed in: esos archivos y `cd site && pnpm run check` · **Retired:** `README.md` solo lista flows y no skills de soporte; lo reemplaza AC14
- AC14 — `llms.txt`, `site/src/content/docs/skills.md` e `index.mdx` nombran `project-layout` y `workspace-repo`, `AGENTS.md` nombra `project-layout`, ninguno nombra `docs-cleanup`, y `rg -l docs-cleanup` fuera de la historia (cambios archivados, este cambio activo, sesiones, `CHANGELOG.md` salvo su entrada nueva, la regex de transcripts de `tests/pilot/session_audit.go`, la nota histórica de `workspace-and-artifacts.md` y la de `tests/fixtures/workspace-conventions/README.md`) no devuelve nada · false today because nombran `docs-cleanup` · Observed in: esos archivos y `cd site && pnpm run check` · **Replaces:** AC11
- AC12 — El artifact «Skills de Hive» muestra `project-layout` y `workspace-repo` y no `docs-cleanup` · false today because muestra `docs-cleanup` · Observed in: https://claude.ai/artifact/77U7EN4pUTcEtiAijziLxE
- AC13 — Sin línea `Layout:`, `check_layout.py` reporta una carpeta `docs/` con Markdown en la raíz como casa de docs no declarada; con `Layout: docs at docs/ · tools at scripts/` en `docs/README.md` no la reporta, sí reporta un documento de `docs/` y un archivo directo de `scripts/` ausentes del índice; `layout.md` dice que esa línea manda sobre los destinos por defecto y `migrate.md` que el modo migrar la escribe cuando el usuario declina · false today because no existe la línea ni el check · Observed in: `tests/skills/layout_check_test.py` y `content/skills/project-layout/references/`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Inventario de soporte de cam, carvento, fleetsystem y facti, 2026-11-17, cuatro hijos `hive-research` de solo lectura: resultados resumidos en `engram:tricell-hive/decision/monorepo-or-workspace-layout`.
- Comportamiento de los hosts con repos anidados, 2026-11-17 (Claude Code 2.1.327, Codex 0.169.0, Grok 1.0.89, Pi 1.5.2, OpenCode 2.3.0, Cursor agent 2026.11.13; Antigravity no instalado): `rg` desde una raíz cuyo `.gitignore` lista a un hijo no encuentra nada; un `.ignore` con `!<hijo>/` lo devuelve; Codex, Grok y Cursor no cargan el `AGENTS.md` del coordinador desde un hijo; el worktree nativo creado desde el coordinador no contiene a los hijos.
- Supuestos de Hive sobre la raíz del workspace, 2026-11-17: `global.md:78` («do not initialize Git»), `tooling/cli/doctor_project.go:61,281-289` (mensaje fuera de Git), `tooling/viewer/index.go:94-118` y `registry.go:82-100` (clave de proyecto por remoto cuando hay `.git`), `content/skills/flow-plan/scripts/tasks.py:355-376` (exige `_support/workspace` ignorado).
- Práctica externa del patrón coordinador: Home Repo (aihero.dev, 2026-11-12), Anyline (2026-03-05), Superblocks (2026-03-10); evidencia de blogs sin métricas.
- `docs-cleanup` tiene 36 referencias en 22 archivos; `artifact-layout` 15 consumidores; `project-layout.md` de harness-audit 3 consumidores (`rg`, 2026-11-17).
- Hallazgos esperados para AC5, del inventario de solo lectura del 2026-11-17: `cam/HIVE.local.md` no nombra ningún script y `cam/_support/scripts/` tiene `embed-host-harness/`, `seed-local-demo.ts`, `com-904-manual-fixture/`, `manual-dummy-seed/` y `meta-app-review/`; `carvento-monorepo/_support/scripts/` tiene 46 archivos versionados y ningún `HIVE.local.md` existe en carvento; `facti-2020-new/_support/scripts/` 65 y `facti-backend-client/_support/scripts/` 16 archivos versionados; `facti-2020-new/server/scripts/cancel-payment-complements.js` lee `_support/workspace` y escribe `_support/reports`; `fleetsystem/HIVE.local.md` nombra sus 7 scripts y `fleetsystem/_support/` tiene `evidence/` y `migration-data/`.
- Historia: `workspace-conventions` se fusionó en flow-close en el cambio `skills-rebuild` (2026-10-30) para reducir ceremonia; este cambio devuelve la convención a una skill propia porque ahora también organiza y migra.

## Effects

- Republicar el artifact privado «Skills de Hive» en claude.ai · target: https://claude.ai/artifact/77U7EN4pUTcEtiAijziLxE · ceiling: 0 · tasks: T10 · answer: approved, pedido explícito del usuario en la sesión del 2026-11-17 («recuerda actualizar el artifact de las skills»)

## Delivery

Respuestas vigentes en la sesión para este repositorio: entrega `ready` (revisión, push, PR no draft, CI y respuesta a sus revisiones; se detiene con el PR abierto para que el usuario lo integre) y revisión nativa `/code-review` de Claude Code. No autoriza aplicar las skills en otros proyectos, crear repos remotos ni escribir en cam, carvento, fleetsystem o facti.

## Review status and progress

2026-11-18 — Revisión tras `/code-review` (10 hallazgos aplicados): la regla 3 pasa a revisión y excluye tests, porque por extensión no distingue código de la aplicación de una herramienta que gestiona `_support/` y daba 22 fallos en Hive; la regla 1 exige el nombre exacto; los enlaces del índice aceptan `<…>`, `%20`, otros esquemas y rutas del sitio; `workspace_repo.py` conserva las claves de `## Hive`, las filas y el encabezado `## Repositorios`, toma la rama base y no la actual, valida rutas y remotos en `setup`, y sugiere `git_state.py`.
2026-11-18 — Revisión en build de AC4: la regla 3 cuenta solo código, no comentarios, que cita material de trabajo (`workspace`, `data`, `scripts`, `secrets`, `sessions`, `evidence`, `reports`); motivo: la corrida real marcó comentarios que citan contratos en carvento y cam y lecturas legítimas de `_support/openspec`. Revisión de esa revisión (1 P2, 4 P3 aplicadas): marcas de comentario por lenguaje para no saltar flags de shell, «línea que no es comentario», `evidence/` y `reports/` nombrados en `layout.md`, y la prohibición de que la aplicación en ejecución lea `_support/` se conserva como regla de juicio.
2026-11-18 — Revisión de AC14 (2 P1, 1 P2, 1 P3 aplicadas): `AGENTS.md` solo nombra `project-layout`, que es la skill que su línea enruta; exclusiones de historia completas; nota de la fixture actualizada.
2026-11-18 — Revisión en build: AC11 se retira y AC14 lo reemplaza sin `README.md`, que solo lista flows; T5 deja de editarlo.
2026-11-18 — Segunda revisión (2 P2, 2 P3 aplicadas): gramática en inglés de `Layout:`, orden de búsqueda del índice, regla 7 de casa de docs no declarada, efecto de `tools at`, verificación de la documentación.
2026-11-18 — Revisión del plan: se agrega AC13 (línea `Layout:` del índice para las excepciones de D16), en T1, T2 y T3; motivo: la excepción no era legible en runtime ni para el check; no invalida tareas hechas.
2026-11-18 — Plan escrito. Revisión `hive-review-plan`: 4 P1, 7 P2, 6 P3, todas aplicadas (tareas T1–T2 unidas, exenciones del check, `## Hive` del coordinador preguntado por la skill, clave del viewer, comando de tests, consumidores faltantes, CHANGELOG, ciclo de vida sin repetir «Candidates», insumos copiados de memoria).
