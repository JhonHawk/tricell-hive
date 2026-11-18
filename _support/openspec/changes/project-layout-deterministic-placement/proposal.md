# project-layout-deterministic-placement — Destinos deterministas para lo que project-layout mueve

| Campo | Valor |
| --- | --- |
| Estado | Draft |
| Tracker | Sin ticket |
| Git | tricell-hive · base `development` en `5a64191b` · rama propuesta `feat/project-layout-placement` |

## Decisions

2026-11-18
- **Resolver todo antes de la nueva corrida** — ¿Qué se hace con los hallazgos H20–H26 de la primera corrida real en carvento? Se resuelven todos aquí; el usuario vuelve a correr project-layout en una sesión nueva de carvento (`engram:tricell-hive/decision/project-layout-determinism`).
- **Multilenguaje (H20)** — La exclusión de dependencias y salidas de build vale para cualquier ecosistema, no solo `node_modules`.
- **Carpetas nuevas sí, pero deterministas (H21)** — Crear carpetas nuevas está bien; el nombre y la agrupación deben salir iguales al repetir la evaluación y en otro proyecto con el mismo problema.
- **Reparar referencias** — La skill repara las referencias a los archivos que mueve, cuando aplica.
- **D20 A** — ¿Con qué enfoque se construye el planificador? Un script nuevo `plan_layout.py` con `plan` en seco por defecto y `apply`; el comprobador sigue de solo lectura.
- **D21 A** — ¿Cómo se agrupan las herramientas? Por propósito, según quién las llama: las de un paquete en `<paquete>/scripts/`, las de CI en `scripts/ci/`, las de hooks en `scripts/hooks/` y las que solo cita un runbook en `scripts/ops/`, reutilizando el grupo equivalente que el proyecto ya tenga.
- **D22 A** — ¿Qué pasa con la evidencia heredada? Cada subcarpeta de `_support/evidence` o `_support/reports` sin versionar pasa a `_support/workspace/<fecha>-<trabajo>/evidence/`, sin borrar nada, y se cura ahí.
- **D23 A** — ¿Cómo se llama la carpeta de infraestructura cuando no existe? `infra/`.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> H20 - tendremos que hacerlo compatible con multiples lenguajes
> H21 - Esta bien que se creen las carpetas nuevas!, lo que no quiero es que se vuelva algo NO deterministico, que si volvemos a ejecutar la misma evaluación el nombre de la carpeta, agrupación etc cambie (o si es un proyecto diferente con el mismo problema)
> H22/23 - Hagamos research y dame una propuesta para esto
> Debemos resolver todos los hallazgos, y volvere a ejecutar en una sesión nueva en ese proyecto, podemos analizar otros proyectos si nos ayuda a darnos una idea de las posibles agrupaciones necesarias

> Considerar que la skill tambien repare las referencias de los archivos que mueve (si aplica)

Hallazgos de la corrida (Codex 0.170.0, `gpt-6.2-sol`, carvento, rollout `01a1ed72`, 2026-11-18):

- H20: `check_layout.py` cuenta 309 Markdown de `node_modules/`.
- H21: 44 scripts a una carpeta inventada `scripts/automation/`, sin reglas.
- H22: Terraform y JSON de IAM siguen versionados bajo `_support/`.
- H23: `contracts/`, `product/`, la evidencia heredada y el script de aprovisionamiento quedaron sin propuesta.
- H24: rondas de decisiones solo en texto, que el detector no marcó.
- H25: la pregunta de revisión no ofreció «Codex /review».
- H26: líneas `Layout:` declaradas por el agente; la pregunta de cambiar Hive se hizo sin formato de decisión.

## Objective

Hoy:
- el agente decide en cada corrida a qué carpeta va cada script y cómo agruparlo;
- el comprobador recorre carpetas de dependencias;
- Terraform versionado bajo `_support/` solo aparece como aviso;
- la evidencia heredada no tiene una acción definida;
- las referencias a lo movido se corrigen a mano.

Al terminar:
- un script calcula, con reglas en orden fijo, el destino de cada archivo de código, infraestructura, conocimiento y material heredado, y las reescrituras de sus referencias;
- dos corridas sobre el mismo árbol dan el mismo plan;
- `apply` mueve, reescribe y verifica sin hacer commit;
- el comprobador ignora dependencias de cualquier lenguaje y falla con código versionado bajo `_support/`.

## Scope and acceptance

Dentro:
- el comprobador (H20 y la regla nueva LAYOUT-08);
- el planificador `plan_layout.py` con plan, override, apply y verify;
- la convención en `layout.md` y `script-lifecycle.md`, y el modo organizar de la skill;
- el detector de rondas en texto (H24) y un criterio nuevo para la opción nativa de revisión (H25);
- los watch de H25 y de la pregunta de proyecto externo;
- `rule-enforcers.md`, la arquitectura de workspace, el CHANGELOG y la página del sitio que describe la skill.

Fuera:
- aplicar el plan en carvento, cam, fleetsystem o facti, que corre en sesiones de esos proyectos;
- corregir el código de facti que escribe en `_support/evidence` y `_support/reports`;
- cambiar la regla de preguntas T10, que pertenece a la prueba `question-card-alone`.

- AC16 — El comprobador no reporta archivos dentro de carpetas de dependencias o build de ningún ecosistema: dentro de un repo con `git ls-files -co --exclude-standard`, y fuera de Git podando una lista fija (`node_modules`, `.venv`, `venv`, `__pycache__`, `.tox`, `.terraform`, `.gradle`, `.dart_tool`, `_build`, `Pods`, `.next`, `.nuxt`, `.turbo`, `.pnpm-store`, `.bundle`, `site-packages`, `.mypy_cache`, `.pytest_cache`) y los nombres condicionados (`target`, `vendor`, `dist`, `build`, `bin`, `obj`, `deps`) solo cuando a su lado hay el manifiesto que los genera, y los repos hijos; y sigue viendo lo ignorado bajo `_support/` (un `_support/scripts/x.sh` ignorado y no registrado da LAYOUT-01) · false today because el check sobre `carvento-specs` reporta 309 Markdown de `docs/manual-usuario/node_modules/` como LAYOUT-05 · Observed in: las pruebas con fixtures de cinco ecosistemas, y el check sobre el checkout real de `carvento-specs` (solo lectura) mientras su árbol conserve el `Index: README.md` y el `Layout: docs at docs/` sin commit de la otra sesión; sin ellos, el check sale sin índice y la observación real no aplica
- AC17 — Un archivo de código o infraestructura versionado bajo `_support/`, fuera de `docs/`, `openspec/` y `sessions/`, y fuera de lo que ya reporta LAYOUT-02, es un fallo LAYOUT-08 que nombra el comando del planificador · false today because `_support/infrastructure/terraform/*.tf` de carvento-monorepo solo produce un aviso LAYOUT-04 y el check sale 0 por esa carpeta · Observed in: el check sobre un clon local en scratch de carvento-monorepo en `5f06d0c1`
- AC18 — `plan_layout.py plan` es determinista: dos corridas sobre el mismo árbol dan una salida idéntica byte a byte, y dos fixtures con la misma estructura y nombres de proyecto distintos dan los mismos destinos relativos · false today because no existe planificador y la corrida de carvento eligió `scripts/automation/` a criterio · Observed in: el plan corrido dos veces sobre el clon en scratch de carvento-monorepo en `5f06d0c1`, y las fixtures
- AC19 — El plan sigue la tabla de reglas sobre carvento-monorepo en `5f06d0c1`. Destinos: `detect-changes.sh` → `scripts/ci/`; `backend/backup-database.sh` y `backend/set-channel-credentials-key.sh` → `apps/backend/scripts/`; `backoffice/check-i18n.mts` → `apps/backoffice/scripts/`; `verify-contracts-push.sh` → `scripts/hooks/`; `rollback-environment.sh` → `scripts/ops/`; `_support/infrastructure/terraform/**` → `infra/terraform/**`; `_support/aws/**` → `infra/aws/**`, incluidos sus `README.md` y `*.tfvars.example`. Cada `*.test.sh` sigue a su sujeto, y las pruebas sin sujeto siguen a los candidatos que citan. `_support/infrastructure/backend/iam/github-actions-terraform-policy.json` sale como fila de revisión porque ningún archivo lo cita · false today because no hay plan; la corrida dejó todo en `scripts/automation/` y la infraestructura en `_support/` · Observed in: la salida del plan sobre el clon en scratch
- AC20 — `apply` mueve y repara referencias: con `apply` sobre el clon en scratch, `git grep -F` de cada ruta vieja no da coincidencias fuera de los registros históricos. Las rutas relativas al archivo que las cita se recalculan (`../../_support/scripts/backend/contracts-check.sh` en `apps/backend/package.json`). También se reescriben las referencias dentro de los propios scripts movidos (`$REPO_ROOT/_support/scripts/verify-amplify-cleanup.sh` en `deploy-amplify.test.sh`). Las pruebas shell del repo pasan tras ajustar a mano solo las filas de revisión de AC21 · false today because no existe `apply`; la corrida de carvento editó a mano 18 archivos · Observed in: el clon en scratch de carvento-monorepo tras `apply`, con sus pruebas `*.test.sh`
- AC21 — Un prefijo o glob que cubre archivos con destinos distintos (`find _support/scripts -name '*.test.sh'` en `ci.yml`) y una ruta interna `../…` de un script con ancla de ubicación propia (`BASH_SOURCE`, `dirname "$0"`, `__dirname`, `import.meta.url`, `__file__`) que resuelve a otro destino desde la carpeta nueva, también en otra línea que la del ancla (`marketplace/update-contracts.mts`), salen como filas de revisión con las rutas nuevas, sin reescribirse. Un prefijo cuyos archivos van todos al mismo destino (`TF_ROOT: _support/infrastructure/terraform`) se reescribe · false today because no hay reescritura · Observed in: el plan sobre el clon en scratch y las fixtures
- AC22 — Las citas en registros históricos no se tocan (`openspec/changes/archive/`, `_support/openspec/changes/archive/`, `sessions/`, `evidence/`, `_support/workspace/`, `CHANGELOG*`), también cuando están en otra raíz pasada con `--root` · false today because no hay reescritura · Observed in: las fixtures con dos raíces, y el plan sobre el clon en scratch junto a un clon de carvento-specs en `965a73d8`
- AC23 — Las carpetas heredadas reciben su categoría y destino:
  - carpeta solo de Markdown e imágenes → `_support/docs/<nombre>/`;
  - subcarpeta sin versionar de `_support/evidence` o `_support/reports` → `_support/workspace/<fecha>-<trabajo>/evidence/`, con la primera fecha `YYYY-MM-DD` de su nombre (`COM-588-2026-09-17-qa`) o, sin fecha, la del archivo más antiguo;
  - `backups` y `migration-data` → `_support/data/<fecha>-<nombre>/`;
  - una de esas carpetas que una línea no comentada de código o configuración cita → fila de revisión;
  - `_support/contracts` y `_support/product` (solo Markdown) → `_support/docs/contracts/` y `_support/docs/product/`.

  false today because LAYOUT-04 solo dice «classify it with project-layout organize mode» · Observed in: las fixtures, y el plan en seco (solo lectura, sin `apply`) sobre la raíz real del workspace de carvento y sobre el árbol real de carvento-monorepo, cuyos `_support/contracts`, `_support/product` y `_support/evidence` existen hoy sin commit o ignorados
- AC24 — `plan --override <origen>=<destino>` marca la fila como elegida por el usuario. `apply` la usa. Cuando el override cambia la casa de una categoría (tools, infra o docs), el plan imprime la línea `Layout:` que hay que agregar al índice · false today because no hay overrides; la corrida escribió dos líneas `Layout:` por su cuenta · Observed in: las fixtures
- AC25 — `apply` se niega, sin mover nada, cuando el plan está viejo (cambió la huella de un origen o de una carpeta, una línea a reescribir o el commit de una raíz) o un destino existe, y nunca hace commit · false today because no existe `apply` · Observed in: las fixtures
- AC29 — Sobre un árbol con cambios locales ajenos en un archivo que `apply` reescribe, `apply` y luego `undo --plan` devuelven el árbol exactamente al estado previo, conservando esos cambios ajenos (`git diff` y `git status` iguales antes y después) · false today because no existe `undo`, y la recuperación propuesta (`git checkout -- .`) borraría el trabajo de la otra sesión · Observed in: las fixtures, y el clon en scratch de carvento-monorepo con un cambio local añadido
- AC26 — `--audit-session` marca `text_round_without_card` como fallo en una sesión de Codex que no llamó la tarjeta, cuando su `additional_tools` lista `request_user_input` · false today because el rollout `01a1ed72` da `not_observed` · Observed in: `go run ./tests/pilot --audit-session` sobre el rollout real `01a1ed72`
- AC27 — `--audit-session` falla un criterio nuevo cuando, en Codex, una pregunta de revisión de código no ofrece la revisión nativa del host, y pasa cuando la ofrece · false today because ningún criterio lo mira y `01a1ed72` (D4 sin «Codex /review») no da hallazgo · Observed in: `--audit-session` sobre los rollouts reales `01a1ed72` (falla) y `01a1ecec` (pasa)
- AC28 — El modo organizar de project-layout corre el planificador para el código, la infraestructura, las carpetas de conocimiento y el material heredado, copia sus filas al manifiesto y aplica con `apply`. `layout.md` declara el vocabulario de destinos que implementa el script. El agente escribe una línea `Layout:` solo a partir de un override del usuario · false today because el modo organizar clasifica todo por contenido a criterio · Observed in: la próxima sesión real de project-layout en carvento (queda sin verificar hasta entonces), además de las pruebas de contenido

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- Corrida real: Codex rollout `~/.codex/sessions/2026/11/18/rollout-2026-11-18T00-57-53-01a1ed72-9d0c-7a00-b585-00a8031dd3fa.jsonl`. Manifiesto en `carvento/_support/workspace/2026-11-18-project-layout/manifest.md`, leído el 2026-11-18.
- Inventario de cam, carvento, fleetsystem y facti (hijo `hive-research`, 2026-11-18):
  - cam usa `scripts/{ci,rebuild,local,testing}/` e `infra/terraform`;
  - fleetsystem usa `scripts/{deployment,database,migrations,infrastructure}/` e `infrastructure/`;
  - carvento tiene `scripts/prepare-install.sh` y `apps/backend/src/scripts/` en HEAD `5f06d0c1`, con su Terraform en `_support/infrastructure/terraform` (`TF_ROOT` en `terraform-{qa,production}.yml`) y `_support/aws` llamado desde `lefthook.yml:209`;
  - 15 referencias de código de carvento-monorepo apuntan a `carvento-specs/contracts|docs`;
  - un cron de producción de facti escribe en `_support/evidence` (`global-emision.js:64`).
- Convenciones externas (hijo `hive-research`, 2026-11-18): plantillas de github/gitignore por ecosistema; `git ls-files --exclude-standard` respeta `.gitignore` anidados, `info/exclude` y `core.excludesFile`, probado con git 2.56.1; CODEOWNERS y Bazel como precedentes de asignación por precedencia fija y frontera más cercana.
- Detector: `tests/pilot/session_audit.go:1063` exige una llamada de pregunta en la sesión, y `codexItem` (línea 250) no lee `additional_tools`.
- Comprobador: `content/skills/project-layout/scripts/check_layout.py:38` solo excluye `__pycache__` y `.DS_Store`.

## Effects

None: the tasks take no effect on an external system.

## Delivery

Respuestas en vigor en esta sesión para tricell-hive:
- entrega `ready`: revisión, push, PR no draft y espera de CI, y queda abierto para que el usuario haga merge;
- revisión de código con `/code-review` nativo de Claude Code.

No autoriza correr el plan sobre los proyectos ni escribir en ellos: los clones de prueba viven en el scratch de la sesión.

## Review status and progress

2026-11-18 — Plan escrito tras el research de la sesión b103d2f3.
2026-11-18 — La revisión `hive-review-plan` encontró 6 P1, 10 P2 y 5 P3; todos se aplicaron:
- la enumeración vuelve a ver lo ignorado bajo `_support/`;
- LAYOUT-08 excluye `sessions/`;
- la regla IaC mueve el árbol completo y manda a revisión los JSON sin citas;
- se reescriben las referencias dentro de los scripts movidos, con una regla para pruebas sin sujeto;
- tokens con límites y UTF-8 estricto;
- las rutas internas se comparan por resolución;
- hay un orden fijo de clases de llamador y de reglas de carpeta;
- la palabra «una sola vez» se reconoce por token;
- `undo` en lugar de un restore general, y AC29;
- verify sin índice;
- el criterio de revisión se limita a títulos de decisión y tarjetas, y el detector comprueba el modo;
- se agregaron `migrate.md` y `workspace-repo/SKILL.md` a T5, la clave `infra at`, la fecha en cualquier parte del nombre y la huella de carpetas.

`contracts/` y `product/` se observan en el árbol real del monorepo (supuesto en `design.md`).
