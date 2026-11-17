# deterministic-rule-validators — Pasar reglas mecánicas de Hive a validadores y scripts determinísticos (Fase 1: H6–H12)

| Campo | Valor |
| --- | --- |
| Estado | Ready for flow-build |
| Tracker | none |
| Git | `tricell-hive`, rama base `development` (`f886538f`), rama propuesta `feat/deterministic-rule-validators` |

## Decisiones

2026-11-16
- **D1 A** — ¿Qué arquitectura usamos para implementar los validadores y helpers determinísticos (H6–H12)? Enfoque híbrido siguiendo las fronteras actuales del repositorio: Go en `tooling/cli/doctor_project.go` y `tooling/management/references.go` para diagnóstico de proyecto y validación de anclas `#fragment` en empaquetado, y scripts en Python 3 estándar sin dependencias externas bajo `content/skills/<skill>/scripts/` para las operaciones determinísticas de cada flujo.
- **D2 A** — ¿Qué alcance incluimos en este plan de la Fase 1? Incluir los 7 puntos (`H6`–`H12`) en un solo plan con tareas independientes por validador y el ajuste de `content/guidance/global.md`.
- **D3 A** — ¿Cómo ajustamos la prosa de `content/guidance/global.md` y de los skills al introducir estos scripts y comandos? Mantener el invariante normativo de cada regla en `content/guidance/global.md` y en los skills propietarios, y sustituir las recetas manuales de comandos shell por la invocación al script o subcomando determinístico.
- **D4 A** — ¿Qué modo de entrega usamos para este cambio en `tricell-hive`? `interactive`: detenerse antes del `git push` para validación humana del resultado y del texto de `content/guidance/global.md`, y tras la aprobación hacer push, abrir pull request hacia `development` e integrar cuando pase CI.
- **D5 A** — ¿Qué revisión de código corre sobre el candidato antes de tu validación o del push? `Agent: hive-review-code` antes de la parada de validación.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> "https://www.youtube.com/watch?v=O_zcc6Xl12g analizar este videos, toma los puntos relevantes de este y contrastalo con hive, qué skills nos hacen falta, cuales cosas de reglas pueden pasar a un método deterministico o de los bullets que generes del video, cómo podemos llegar a ello?"
>
> Elección posterior del usuario: `A` — Planificar la Fase 1 (`H6`–`H12`: pasar reglas actuales de Hive a scripts/CLI determinísticos y adelgazar `global.md`).

Quedan fuera del alcance de este cambio las Fases 2 y 3 del contraste con el video (`H1`–`H5`: generador de *feature map* / arnés de verificación por proyecto, escalera `correct` para código de proyectos cliente, auditoría `no-comments` y verificación formal con TLA+/Lean/Bombadil).

## Objetivo

Hoy varias reglas mecánicas sin excepciones en `content/guidance/global.md` y en `content/skills/` dependen de que el modelo arme a mano pipelines de shell para validar `HIVE.local.md`, `ACCOUNTS.md` y `_support/secrets/`, calcular si un paquete tiene al menos 7 días de publicado, preservar rutas preparadas en el índice al hacer commit con escaneo de secretos, verificar si un worktree o rama fusionada puede borrarse en `flow-close`, validar o reemplazar la sección `## Business rules` entre marcadores, y comprobar las listas de preparación (*ready lists*) de PRDs y tickets, mientras que `tooling/management/references.go` ignora las anclas `#fragment` al validar enlaces internos.

Al terminar este cambio, cada una de esas 7 verificaciones y operaciones cuenta con un validador o script determinístico cubierto por tests automatizados, y `content/guidance/global.md` junto con los skills propietarios invocan directamente esos métodos determinísticos manteniendo intactos sus invariantes y reduciendo el tamaño en bytes de la guía global.

## Alcance y criterios de aceptación

Incluye: `H12` en `tooling/management/references.go`; `H6` en `tooling/cli/doctor_project.go` y `content/skills/harness-audit/scripts/local_instructions.py`; `H7` en `content/skills/flow-build/scripts/check_package.py`; `H8` en `content/skills/git-workflow/scripts/git_state.py`; `H9` en `content/skills/flow-close/scripts/cleanup_branches.py`; `H10` en `content/skills/flow-plan/scripts/check_change.py` y `content/skills/flow-plan/scripts/business_rules.py`; `H11` en `content/skills/create-prd/scripts/check_prd.py` y `content/skills/create-issue/scripts/check_issue.py`; actualización de `content/guidance/global.md`, skills propietarios, documentación en `_support/docs/architecture/` y `site/`, y deltas en `_support/openspec/`.

Excluye: `H1`–`H5` (Fases 2 y 3), cambios a los perfiles de modelos o al servicio del visor (`tooling/viewer/`).

- AC1 — `validateInstructionReferences` en `tooling/management/references.go` verifica que todo enlace Markdown interno con `#fragment` en `content/guidance/global.md`, `content/skills/` y `content/agents/` resuelva a un encabezado real del archivo destino y rechaza un ancla inexistente · false today because `tooling/management/references.go:13,49` salta enlaces que empiezan con `#` y recorta el sufijo `#fragment` sin comprobar encabezados · Observed in: `go test ./tooling/management/...` y `go run ./tooling/cli update --source . --dry-run` sobre el catálogo del repositorio.
- AC2 — `hive doctor --project <dir>` en `tooling/cli/doctor_project.go` y `content/skills/harness-audit/scripts/local_instructions.py` verifican y aseguran que `info/exclude` ignore `HIVE.local.md`, `_support/secrets/` y `_support/scripts/`, que `_support/secrets/` tenga modo `0700` y sus archivos secretos `0600` (salvo `ACCOUNTS.md`), que cada línea de entrada de `HIVE.local.md` (ignorando el párrafo introductorio y encabezados Markdown `#`) tenga fecha `YYYY-MM-DD` y no supere 200 caracteres (descontando `STALE <date> (<cause>):` y la fecha/causa de fallo que gana una entrada `unverified`), que superar 50 líneas solo bloquee cuando queden entradas `STALE` sin podar (emitiendo aviso de curación no bloqueante cuando todas son vigentes/`unverified`), y que `HIVE.local.md` y `ACCOUNTS.md` no contengan secretos según `secret_like`, `SECRET_FLAG`, `ENTROPY` y `MIXED` de `extract_candidates.py` · false today because `tooling/cli/doctor_project.go` no inspecciona `HIVE.local.md` ni `_support/secrets/` y `local_instructions.py` no existe · Observed in: `go test ./tooling/cli/...`, `python3 -m unittest tests/skills/local_instructions_test.py` y `go run ./tooling/cli doctor --project .`.
- AC3 — `content/skills/flow-build/scripts/check_package.py` verifica para `npm`, `pypi`, `go` y `cargo` la fecha de publicación de una versión y la ausencia de advisories activos en OSV, devolviendo código distinto de cero cuando la versión tiene menos de 7 días de publicada o presenta vulnerabilidades reportadas · false today because `content/skills/flow-build/scripts/check_package.py` no existe y la verificación depende de cálculos manuales del agente · Observed in: ejecución real de `python3 content/skills/flow-build/scripts/check_package.py` contra el registro e `python3 -m unittest tests/skills/check_package_test.py`.
- AC4 — `content/skills/git-workflow/scripts/git_state.py commit` rechaza flags de preparación masiva o salto de hooks (`-a`, `-A`, `.`, `--no-verify`, `-n`) y rutas de directorio, conserva intactas las rutas previamente preparadas en el índice fuera de `<paths>`, registra archivos `untracked` incluidos en `<paths>` con `git add -N` (revirtiéndolos del índice si falla el preflight), ejecuta el escáner de secretos exclusivamente sobre el diff/contenido de `<paths>`, exige el trailer `Unreviewed: <reason>` cuando se indica `--unreviewed-fix` y realiza el commit con `git commit --only -- <paths>` manteniendo la invocación sin subcomando de `git_state.py` de solo lectura · false today because `content/skills/git-workflow/scripts/git_state.py` carece de subcomando `commit` · Observed in: `python3 -m unittest tests/skills/git_state_test.py`.
- AC5 — `content/skills/flow-close/scripts/cleanup_branches.py` genera en modo dry-run JSON (y aplica con `--apply`) el plan determinístico de limpieza de ramas y worktrees al cierre, reteniendo cualquier worktree con cambios sin confirmar, archivos sin seguimiento, procesos con `cwd` dentro del worktree (o cuando `lsof` no está disponible o falla), ramas protegidas/activas (base, entornos, base de PR abierto o nombradas en checkpoints/planes con cierre pendiente `< 24h`) o entradas ignoradas no permitidas distintas de las del checkout principal, y comprobando `headRefOid` antes de eliminar ramas fusionadas por squash o rebase · false today because `content/skills/flow-close/scripts/cleanup_branches.py` no existe · Observed in: `python3 -m unittest tests/skills/cleanup_branches_test.py` y corrida dry-run sobre el repositorio.
- AC6 — `content/skills/flow-plan/scripts/check_change.py` (sin `--archive`) valida la gramática canónica de las líneas bajo `## Business rules` (`None:`, `ticket section: <key> · <heading>`, `ticket write: <date> · <moment> · <key>`, `rule: <rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>` sin estado `stated:`, `card stale: <screen> · <rule>`, `frame stale: <node> · <rule>`) en `proposal.md`, y `content/skills/flow-plan/scripts/business_rules.py` inserta al final en la primera escritura o reemplaza entre `<heading>` y `<heading> · end` (rechazando editar si un marcador aparece huérfano o duplicado) verificando que el texto fuera de los marcadores no cambie ni un byte · false today because `check_change.py` ignora `## Business rules` y `business_rules.py` no existe · Observed in: `python3 -m unittest tests/skills/change_check_test.py tests/skills/business_rules_test.py`.
- AC7 — `content/skills/create-prd/scripts/check_prd.py` valida en inglés y español las 9 secciones obligatorias en orden, encabezados `### R<n> · P<0-2>[(proposed)] · <state>` sin tablas ni códigos duplicados y campos `Source:`/`Fuente:` y `Verdict:`/`Veredicto:`, y `content/skills/create-issue/scripts/check_issue.py` valida en inglés y español las secciones por tipo de ticket (`story`, `bug`, `task`, `spike`), el tope de `<= 7` criterios de aceptación (con excepción `--finding-group`) y las restricciones del solution hint (`<= 5` ubicaciones y 1 oración no vinculante) · false today because `check_prd.py` y `check_issue.py` no existen · Observed in: `python3 -m unittest tests/skills/check_prd_test.py tests/skills/check_issue_test.py`.
- AC8 — `content/guidance/global.md` y los skills propietarios reemplazan las recetas manuales de shell por la invocación a los validadores y scripts de `AC1`–`AC7` conservando todos sus invariantes normativos, disminuyen el tamaño en bytes de `content/guidance/global.md` por debajo de los 52,508 bytes actuales, completan las columnas de **Después** en `analysis-and-baseline.md` con las mediciones reales de `TestGuidanceReadingFootprint` y tamaños por archivo/sección, y pasan la suite completa de validación de contenido y paridad · false today because `content/guidance/global.md` mide 52,508 bytes con recetas manuales de shell, los skills no invocan los nuevos scripts y las columnas de Después en `analysis-and-baseline.md` están pendientes · Observed in: `go test -v ./tests/content -run Footprint`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p "*_test.py"`.

## Prerequisites

None: the build needs nothing from outside the repository.

## Contexto verificado

- El análisis completo del video `O_zcc6Xl12g`, el mapeo de hallazgos `H1`–`H12` y las tablas de línea base numérica (**Antes vs. Después**) quedaron guardados en [analysis-and-baseline.md](analysis-and-baseline.md), junto con la transcripción en [assets/video-O_zcc6Xl12g-transcript.txt](assets/video-O_zcc6Xl12g-transcript.txt) (470 líneas) y metadatos en [assets/video-O_zcc6Xl12g-metadata.json](assets/video-O_zcc6Xl12g-metadata.json).
- `content/guidance/global.md` (**52,508 bytes**, **52,385 caracteres**, **149 líneas** en `f886538f`; unión `plan-build-cycle` en `TestGuidanceReadingFootprint`: **133,441 bytes**) contiene 14 secciones; `## Local instructions` mide 9,218 B, `## Scope and authorization` 6,820 B, `## Preservation` 2,818 B y `## Secrets` 1,574 B. Además, `flow-close/SKILL.md` mide **15,993 bytes** (a 7 B del tope de 16,000 B) y `git-workflow/SKILL.md` mide **15,950 bytes** (a 50 B del tope).
- `tooling/management/references.go:13,49` ignora enlaces que inician con `#` y recorta `#fragment` sin validar hoy los **133 enlaces internos con `#fragment`** (de 528 enlaces internos totales en `content/`).
- `tooling/cli/doctor_project.go:108-395` valida `AGENTS.md` `## Hive` y `<Specs>/changes/` (2 superficies), sin revisar `HIVE.local.md`, `ACCOUNTS.md`, `_support/secrets/` ni `_support/scripts/`.
- `content/skills/` incluye hoy **11 scripts** Python estándar en 6 skills probados en `tests/skills/*_test.py` (`check_change.py`, `tasks.py`, `apply_deltas.py`, `archive_sessions.py`, `git_state.py`, `extract_candidates.py`, `init_workspace.py`, `self_check.py`, `verify_geometry.py`, `contrast-check.py`, `render_pdf.py`); con este cambio pasan a **17 scripts** en 8 skills.

## Effects

None: the tasks take no effect on an external system.

## Entrega

- Modo de entrega: `interactive` (`D4 A`), con parada de validación humana antes de `git push`, seguida de apertura de pull request hacia `development` y merge al pasar CI.
- Revisión de código: `Agent: hive-review-code` (`D5 A`) sobre el diff del candidato antes de la parada de validación.
- Esta elección no autoriza despliegues ni promoción de `development` a `master`.

## Estado de revisión y progreso

- 2026-11-16 — Carpeta de cambio inicial creada tras responder `D1 A`, `D2 A` y `D3 A`; enviada a `hive-review-plan`.
- 2026-11-16 — Reconciliados los 6 hallazgos P1 de `hive-review-plan` en `proposal.md`, `design.md`, `tasks.md` y deltas en `specs/`: gramática exacta y marcadores localizados de `## Business rules` + ruta `tests/skills/change_check_test.py` (`H10`), reglas de `HIVE.local.md` cuando supera 50 líneas sin `STALE` y símbolos reales de `extract_candidates.py` (`H6`), soporte de `untracked` y rechazo de directorios/`--no-verify` en `git_state.py commit` (`H8`), retención ante fallo de `lsof` y escenario `Another session's fresh worktree` (`H9`), soporte bilingüe y `--finding-group` en `check_prd.py` / `check_issue.py` (`H11`), y rutas reales de `site/src/content/docs/flows/*.md` + deltas completos en `specs/workflow-lifecycle/spec.md` y `specs/local-instructions/spec.md`.
