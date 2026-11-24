# process-slimming — Recorte del proceso: simulación, registro del plan, archivo, specs y rondas

| Campo | Valor |
| --- | --- |
| Estado | Draft |
| Tracker | ninguno · decisión `decision/process-slimming-scope` (Engram) |
| Git | tricell-hive · `development` en `51f1949d` · ramas `chore/slim-rounds-auditor`, `chore/slim-plan-record`, `chore/slim-specs`, `chore/slim-layout-migration`, en el checkout principal |
| Playbook | none |

## Decisiones

2026-11-24
- **D7 A** — Los tests sobre el texto no miden comportamiento; los scripts sí se prueban; el material que simula sesiones no se commitea.
- **D10 A** — Planear el recorte con cinco palancas: L1 simulación del auditor, L2 carpeta de cambio, L3 specs, L4 archivo, L5 rondas.
- **D17 A** — El recorte cambia el flujo distribuido en `content/`, para todos los proyectos.
- **D18 A** — El trabajo grande lleva un solo `plan.md` local, sin commitear, con orden fijo. El cuerpo del PR es el registro público. No hay deltas ni archivo, y un spec de contrato se edita en el mismo PR.
- **D19 A** — Entrega en PRs separados en orden: 1) rondas y auditor, 2) registro del plan y fin del archivo, 3) specs.
- **D20 A** — Las fixtures de sesiones se migran a casos escritos dentro de la tabla del test, y se borran.
- **D21 A** — Los proyectos existentes se limpian con un caso nuevo en la migración de `project-layout` (PR 4), sin un skill nuevo.
- **D22 A, reemplazada por D27 A** — **Replaces:** D22. La carpeta única y autocontenida de un trabajo es `<support>/workspace/<YYYY-MM-DD>-<trabajo>/`, sin versionar. Guarda el plan `<YYYY-MM-DD>-<trabajo>.plan.md`, los registros `<YYYY-MM-DD>-<trabajo>.<tipo>.<ext>` y las subcarpetas `images/` (capturas, como ya fijan `before-captures.md` y `layout.md:43`) y `evidence/` (solo dentro de la carpeta del trabajo; el `evidence/` de primer nivel sigue siendo un home viejo). `sessions/` se retira de la convención. Las conclusiones duraderas se promueven a `docs/` cuando el cierre las nombra en **Next**, como hoy hace `support-material.md:11`; datos y secretos siguen en `data/` y `secrets/`; queda en `workspace/` sin fecha un archivo local que no nombra ningún trabajo, como `local-context/`.
- **D28** — La homologación de `project-layout` en los demás proyectos también saca del árbol, con `git rm` y el commit de la entrega de ese proyecto, lo que está versionado en homes viejos: `sessions/`, `evidence/`, `reports/`, `archive/`, `archived/`, `backups/`, `plan/`, `spec/` y `openspec/changes/`. Que siga en el historial no importa; la convención nueva queda clara en el árbol.
- **Ubicación** — El plan está en el checkout principal. El brief de un hijo da su ruta absoluta. Para un sandbox sbx se copia al workspace ignorado del worktree, manda la copia del checkout principal y la copia se renueva tras cada revisión.
- **D23 A** — Al cerrar, el `plan.md` se queda local como material de apoyo y lo poda la retención de `project-layout`. El PR lleva objetivo, criterios con evidencia, decisiones y el registro de reglas de negocio.
- **D24 A** — Cada criterio del auditor lleva al menos un caso que falla y uno por cada otro estado que emite (pass, review). `not_observed` se cubre con una sesión vacía común, y la pérdida de cobertura se mide.
- **Entrega** — `automatic` y `/code-review` de Claude Code, las respuestas vigentes en esta sesión.
- **Supuesto** — Este plan usa todavía la carpeta de cambio, porque la regla nueva no existe hasta el PR 2. T7 lo pasa a un `plan.md` local antes de borrar la carpeta. Descartado: escribirlo ya como `plan.md`, sin `tasks.py` ni revisión.
- **Supuesto** — Sin delta de spec: `workflow-lifecycle` se retira en el PR 3. Descartado: escribir deltas para un spec que se borra.
- **Supuesto** — El registro de ejecución de `tasks.py` (0 usos reales) se conserva porque el modo swarm lo usa; T5 lo ata al `plan.md`. Descartado: borrarlo antes de medir el swarm.
- **Supuesto** — Un hallazgo sobre código que clasifica texto libre por patrones, cuyo fallo solo se muestra sobre una entrada que construyó el revisor, tiene como máximo prioridad P2, salvo que el código esté en un límite que lista `hive-core/build/security-boundaries.md`. La regla vive en close-report «Findings». Descartado: una regla aparte en `review-findings.md`, que sería un segundo lugar.
- **Supuesto** — Un plan va a `plan.md` en los mismos casos en que hoy va a una carpeta: falla la prueba de reanudación, el usuario lo pide, la entrega abarca varios PRs o repositorios, o el cambio es riesgoso. En cualquier otro caso queda en la conversación, y así arreglos y documentación no llevan archivo. Descartado: decidir solo por la prueba de reanudación, que dejaría un plan riesgoso sin `hive-verify-task`.
- **Supuesto** — En `--audit-compare`, el ref del auditor viejo va en `--auditor-ref`, porque `--base` ya existe (`tests/pilot/main.go:382`). El host se deduce de la ruta de la sesión, y `--host` puede forzarlo. Descartado: reusar `--base`.

## Base para relajar reglas (root `AGENTS.md` «Measurement»)

- **L2 (carpeta para cambios chicos) y L4 (archivo).** La razón escrita de la regla actual, «since a hygiene batch merges its delta» (`flow-plan/SKILL.md:33`), deja de valer cuando el PR 2 retira los deltas y el lote. La guía vigente de Claude Code, el proveedor del modelo objetivo, dice: «If you could describe the diff in one sentence, skip the plan».
- **L3 (specs).** La razón de editar specs solo por delta era `TestSpecDeltasApplied`, que sale con el archivo. Los specs de contrato se siguen editando en el PR que cambia el contrato.
- **L5 (prioridad de hallazgos heurísticos).** Es la decisión del usuario (D10 A). Ninguna de las tres bases de «Measurement» la cubre, y queda dicho así. Se apoya en la sesión de Claude Code `17050a02` (2026-11-24): dos rondas de revisión sobre bordes construidos por el revisor generaron 10 fixtures, y dos proyectos de referencia hacen lo mismo. Los cuerpos de PR no nombran proyectos de referencia.
- **L1.** Es un requisito nuevo de este repositorio (D7 A), no la relajación de una regla distribuida.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

D10 A con D17 a D24 (2026-11-24). Viene de comparar seis proyectos de referencia y medir el costo interno:
- la carpeta de cambio es ~38 % de lo añadido en los PRs que la traen;
- el archivo ocupa 78 mil líneas que nadie consulta como historia;
- hay 465 fixtures sintéticas commiteadas;
- los deltas no tienen consumidor una vez retirado el archivo.

## Objetivo

**Hoy:**
- todo trabajo con plan escribe una carpeta de 4 archivos que se commitea y se archiva para siempre;
- los specs reescriben la guía;
- el auditor se prueba con 465 transcripciones sintéticas;
- un borde hipotético de un detector abre otra ronda de arreglos.

**Al terminar:**
- arreglos y documentación que pasan la prueba de reanudación no llevan plan en archivo;
- el trabajo grande lleva un `plan.md` local;
- el PR lleva el registro público;
- no hay archivo ni deltas, y los specs son solo contratos;
- el auditor se prueba con casos dentro del test y con una comparación sobre sesiones reales;
- un borde construido por el revisor tiene como máximo prioridad P2;
- los proyectos existentes se pueden limpiar con `project-layout`.

## Alcance y aceptación

**Dentro:**
- las reglas distribuidas de `content/` (plan, build, verify, cierre, higiene, revisión, negocio, UI, roles, swarm y E2E) con sus scripts y tests;
- `tooling/` (visor, `hive doctor` y escritura de `## Hive`);
- `tests/pilot`;
- root `AGENTS.md`, `CONTRIBUTING.md`, `_support/docs`, `site/`, `README.md` y `llms.txt`;
- los specs de contrato que cambian;
- el borrado de `_support/openspec/changes/` y de los specs de guía de este repositorio;
- la migración de `project-layout`.

**Fuera:** el historial de git y los registros de Engram ya escritos.

- AC1 — `tests/fixtures/sessions/` no existe. Cada criterio del auditor tiene un caso escrito dentro de su test por cada estado que emite, uno que falla incluido cuando emite fail (los cinco criterios anteriores viven en `regression_test.go`), y `not_observed` tiene una sesión vacía común. Los tests de los adaptadores leen transcripciones escritas dentro del test · false today because `git ls-files tests/fixtures/sessions | wc -l` da 465 · Observed in: `git ls-files` y `go test ./tests/pilot/` en la candidata del PR 1
- AC2 — `go run ./tests/pilot --audit-compare --auditor-ref <ref> <sesiones…>` imprime `sesión · criterio · antes → después · líneas` para cada estado que cambia, y nada del texto de la sesión · false today because el modo no existe · Observed in: la corrida con `--auditor-ref 15ac9aa5` sobre el rollout Codex `01a20c02`, que debe mostrar los cambios del PR #614 en las líneas 408, 421 y 496
- AC3 — La regla instalada de «Findings» limita a P2 un hallazgo contra código que clasifica texto libre por patrones cuando su fallo solo se muestra sobre una entrada construida por el revisor · false today because `close-report.md` no tiene ese tope · Observed in: `~/.agents/skills/hive-core/close/close-report.md` tras el merge del PR 1 y `hive update`
- AC4 — La guía instalada no describe carpetas de cambio, deltas, archivo ni lote de higiene · false today because `rg -l 'change folder|change-folder|spec delta|archive-pending|batch-archive|apply_deltas|`proposal\.md`|`design\.md`|`tasks\.md`' ~/.agents/skills --glob '!project-layout/references/migrate.md' --glob '!project-layout/references/inventory.md' --glob '!project-layout/scripts/*'` da archivos · Observed in: esa búsqueda tras el merge del PR 2 y `hive update`, con 0 archivos. Las referencias de la migración de `project-layout` (PR 4) quedan fuera, porque describen el registro viejo para limpiarlo
- AC5 — `tasks.py <plan.md> checkpoint|done|execution` y `check_change.py <plan.md>` funcionan sobre un único `plan.md`, y `apply_deltas.py` ya no existe · false today because ambos reciben una carpeta · Observed in: una corrida sobre un `plan.md` de ejemplo en la candidata del PR 2
- AC6 — `_support/openspec/changes/` no existe en este repositorio · false today because `git ls-files _support/openspec/changes | wc -l` da más de 900 · Observed in: ese conteo en la candidata del PR 2
- AC7 — `_support/openspec/specs/` tiene solo specs de contrato: instalador, visor, sitio, promoción de ramas y la parte de contrato de los mixtos · false today because están los 15 de hoy, con `workflow-lifecycle` (98 requisitos) · Observed in: `ls _support/openspec/specs` en la candidata del PR 3
- AC8 — Sobre un repositorio temporal con homes viejos versionados (`sessions/`, `evidence/`, `reports/`, `changes/archive/`) y una carpeta de cambio sin integrar, el planificador de `project-layout` propone:
  - mover cada elemento a `workspace/<fecha>-<trabajo>/` según su prefijo de fecha y su nombre;
  - sacar del árbol lo versionado con `git rm`;
  - borrar el archivo y las carpetas integradas;
  - conservar la que no integró;
  - etiquetar antes de tocar.

  · false today because el manifiesto no tiene esos casos · Observed in: la corrida de `plan_layout.py` sobre esa copia temporal en la candidata del PR 4
- AC10 — La convención de layout instalada nombra `workspace/<YYYY-MM-DD>-<trabajo>/` como la carpeta del trabajo, con su `<YYYY-MM-DD>-<trabajo>.plan.md`, y ya no tiene `sessions/`. Este repositorio no versiona nada bajo `_support/sessions/` · false today because `layout.md:22` define `sessions/` y `git ls-files _support/sessions | wc -l` da 15 · Observed in: `~/.agents/skills/project-layout/references/layout.md` tras el merge del PR 2 y `hive update`, y ese conteo en 0
- AC9 — `hive doctor` sobre un `## Hive` sin `Specs` no advierte una clave faltante · false today because `doctor_project.go:47` la marca obligatoria · Observed in: `hive doctor` en un proyecto temporal, en la candidata del PR 2

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Inventario por palanca y uso real: dos hijos `hive-research`, 2026-11-24 (registrado en el progreso del plan).
- Revisión del plan: dos `hive-review-plan` (servidor y guía), 2026-11-24.
- Las 5 carpetas activas (`gh-576-swarm-mode`, `gh-576-swarm-related-tickets`, `codex-default-question-channel`, `public-tree-private-identifiers` y esta) corresponden a los PRs #612, #613, #614 y #615, ya integrados, y a este plan.
- La suite completa (`go vet ./...`, `go test -count=1 ./...`, `python3 -m unittest discover -s tests/skills -p '*_test.py'`) · ok · 2026-11-24 sobre `e16a16d9`.
- La comprobación de texto de root `AGENTS.md` «Checks by path» · ok, 95 s · 2026-11-24.
- `go test -count=1 ./tests/pilot/` · ok, 62 s · 2026-11-24.
- `go test -count=1 ./tooling/cli/ ./tooling/viewer/...` · ok, 161 s y 6 s · 2026-11-24.
- `pnpm --dir site check`, `lint`, `build` · ok · 2026-11-24.
- `go test -coverprofile` sobre `tests/pilot` · not measured · lo corre T1 antes de borrar.

## Effects

None: the tasks take no effect on an external system.

## Entrega

`automatic` en cada uno de los cuatro PRs: `/code-review`, push, PR a `development`, merge y actualización de la instalación local que autoriza `AGENTS.md`. Cada PR sale de la base con el anterior integrado.

Cada cuerpo de PR lleva lo que pide root `AGENTS.md` «Guidance structure»: los lugares revisados y el cambio de tamaño de `content/guidance/global.md`. Para revertir hay que ir en orden inverso: revertir el PR 2 después del 3 choca. Sin revisión de seguridad. No autoriza una release.

## Estado de revisión y progreso

2026-11-24 — Plan escrito.
2026-11-24 — Relectura de la revisión en el build. Se aplicó:
- `images/` en lugar de `captures/`.
- El visor publica el único `*.plan.md` de la carpeta.
- Los scripts de `sessions/` (`plan_layout.py` `session_dest`, `check_layout.py` y `archive_sessions.py`, con sus tests) se retiran en el PR 2.
- T9 fecha lo que no tiene fecha como `session_dest` (primer commit, si no la modificación más antigua) y mueve sin etiqueta en un `_support/` fuera de git.
- Se agregaron a T4 y T7 los lugares de guía que faltaban.
2026-11-24 — Revisión en el build: D27 A reemplaza a D22 con el nombre del plan y la carpeta única del trabajo, y retira `sessions/`. D28 hace que la homologación saque del árbol lo versionado. Se agregó AC10 y se amplió AC8; T4, T7 y T9 cambian; T1 a T3 (PR 1, sin commitear) no.
2026-11-24 — Relectura: el servidor resolvió 5 de sus 6 P1 y la guía todos los de su primera ronda. Tras la relectura se aplicó, sin otra ronda:
- AC4 anclado, con `tasks.md` y sin las referencias de la migración.
- La excepción de seguridad del tope P2.
- Las tres condiciones de hoy para escribir el plan.
- La copia de sbx gobernada por la del checkout principal.
- T9 con prueba de integración por alcance desde la base, y los pasos de juicio en `migrate.md`.
- `check_change.py --archive` sale.
- Los roles y skills faltantes en T4.
- T8 con la suite completa.
- La base de L5 declarada como decisión del usuario.
- Los nombres de proyectos de referencia fuera del texto versionado.
2026-11-24 — Revisión del plan: dos `hive-review-plan` (servidor y guía), 12 P1 en total. Se aplicaron:
- La comprobación de T1 pasa a cobertura.
- Mínimo de casos por estado (D24 A).
- `--auditor-ref`, y AC2 se observa sobre un ref con cambios.
- Se retira el criterio `change_folder_uncommitted`.
- T5 lista los enlaces del registro de ejecución y conserva el encabezado de progreso.
- El visor usa una carpeta por plan (D22 A).
- Los specs de contrato se editan en el PR 2.
- T4 completa sus ubicaciones, y el patrón de AC4 se amplía.
- El `plan.md` conserva todas las secciones de hoy.
- El cierre conserva el plan (D23 A).
- El registro de reglas va en el PR.
- El tope P2 queda en «Findings».
- Se suma la base de «Measurement».
- Las verificaciones siguen «Checks by path».
- AC6, AC7 y AC9 quedan separados.
