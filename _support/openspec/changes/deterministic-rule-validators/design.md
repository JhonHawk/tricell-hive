# Design

## Approaches

Se evaluaron tres enfoques para ubicar los validadores y helpers determinísticos `H6`–`H12`:

1. **Enfoque A (elegido en D1 A) — Híbrido siguiendo las fronteras actuales del repositorio (Go en `doctor_project.go` y `references.go` + scripts Python 3 estándar en `content/skills/<skill>/scripts/`):**
   - Cumple el peldaño 2 de «Reuse before writing» (`flow-build/SKILL.md`): extiende `tooling/management/references.go` (`H12`), `tooling/cli/doctor_project.go` (`H6` diagnóstico), `content/skills/harness-audit/scripts/extract_candidates.py` (`H6` detección de secretos/entropía), `content/skills/git-workflow/scripts/git_state.py` (`H8`) y `content/skills/flow-plan/scripts/check_change.py` (`H10`), y agrega scripts Python sin dependencias externas donde los skills ya agrupan sus herramientas (`check_package.py`, `cleanup_branches.py`, `business_rules.py`, `check_prd.py`, `check_issue.py`).
   - Ventaja: permite actualizar los scripts de los skills con `hive update` de contenido sin exigir recompilar el binario `hive`, mientras mantiene `H12` en el gate de empaquetado/instalación en Go y `H6` en `hive doctor --project`.
   - Costo: mantener pruebas unitarias tanto en Go (`tooling/...`) como en Python (`tests/skills/*_test.py`).
2. **Enfoque B — Todo dentro del binario Go `hive` (`tooling/cli/`):**
   - Ventaja: un único binario sin depender de `python3`.
   - Descarte: acopla las reglas de formato de PRD, tickets, reglas de negocio y commits al ciclo de compilación del binario y duplica `check_change.py`, `git_state.py` y `extract_candidates.py`.
3. **Enfoque C — Todo como scripts Python en `content/skills/*/scripts/`:**
   - Ventaja: cero cambios en Go.
   - Descarte: deja fuera la validación de anclas `#fragment` (`H12`) durante `hive install`/`hive update` y la inspección de `HIVE.local.md`/`_support/secrets/` en `hive doctor --project`.

## Decisions

- **D1 A — Arquitectura híbrida Go + Python stdlib:** Go valida el empaquetado (`references.go`) y el diagnóstico de proyecto (`doctor_project.go`), y Python 3 estándar provee los scripts operativos y validadores dentro de cada skill propietario.
- **D2 A — Alcance completo de la Fase 1 (`H6`–`H12`):** los 7 puntos se construyen en un mismo cambio descompuesto en tareas independientes por módulo para cerrar de una vez las modificaciones de especificación y el adelgazamiento de `content/guidance/global.md`.
- **D3 A — Adelgazamiento de recetas manuales manteniendo invariantes:** `content/guidance/global.md` conserva íntegro cada invariante normativo y límite observable, delegando la secuencia mecánica de comandos al script o subcomando determinístico.

## Proposed shape

1. **`H12` — Validación de anclas `#fragment` (`tooling/management/references.go`):**
   - En `validateInstructionReferences`, conservar el destino completo antes de separar `#fragment`.
   - Extraer los slugs GitHub-flavored Markdown de los encabezados ATX (`#` a `######`) del archivo Markdown destino (normalizando a minúsculas, eliminando puntuación fuera de `-` y `_`, convirtiendo espacios en `-` y numerando colisiones `-1`, `-2` igual que GitHub/Goldmark).
   - Soportar tanto enlaces con ruta y ancla (`../flow-build/SKILL.md#tests`, `skill:flow-build/SKILL.md#tests`) como enlaces locales `#anchor` dentro del mismo archivo.
2. **`H6` — `HIVE.local.md`, `ACCOUNTS.md` y `_support/secrets/` (`tooling/cli/doctor_project.go` + `content/skills/harness-audit/scripts/local_instructions.py`):**
   - En `doctor_project.go`: cuando existen `HIVE.local.md`, `_support/secrets/` o `_support/scripts/` en el repositorio, comprobar que `git check-ignore -q` los ignore, que `_support/secrets` tenga permisos `0700` y sus archivos secretos `0600` (salvo `ACCOUNTS.md`), que en `HIVE.local.md` las líneas de entrada (ignorando el párrafo introductorio y encabezados Markdown `#`) midan `<= 200` caracteres (descontando `STALE <date> (<cause>):` y la fecha y causa de fallo que gana una entrada `unverified`), incluyan una fecha `YYYY-MM-DD` y no contengan valores secretos, y que superar 50 líneas de entradas solo sea un hallazgo bloqueante cuando existan entradas `STALE` sin podar (y una advertencia no bloqueante recomendando curación cuando no haya entradas `STALE`).
   - En `content/skills/harness-audit/scripts/local_instructions.py`: subcomandos `check`, `ensure-ignore`, `add` y `mark-stale` que resuelven el checkout principal desde `git worktree list --porcelain`, aseguran las entradas en `$(git rev-parse --git-common-dir)/info/exclude`, reutilizan `secret_like`, `SECRET_FLAG`, `ENTROPY` y `MIXED` de `extract_candidates.py` para rechazar cualquier valor secreto en `HIVE.local.md` o `ACCOUNTS.md`, aplican el límite de 200 caracteres y podan entradas `STALE` cuando el archivo supera 50 líneas (conservando todas las entradas vigentes/`unverified` si no quedan `STALE`).
3. **`H7` — Antigüedad de paquetes ≥ 7 días y advisories (`content/skills/flow-build/scripts/check_package.py`):**
   - CLI `python3 check_package.py --ecosystem <npm|pypi|go|cargo> <pkg>[@<version>] [--min-days 7]` que consulta los endpoints públicos de metadatos (`registry.npmjs.org`, `pypi.org/pypi/<pkg>/json`, `proxy.golang.org/<mod>/@v/<ver>.info`, `crates.io/api/v1/crates/<pkg>`) y la API de OSV (`api.osv.dev/v1/query`), imprimiendo JSON con `published_at`, `age_days`, `publisher`, `advisories` y `ok`, y saliendo con código `1` si `age_days < 7` o si hay advisories activos.
4. **`H8` — Commit seguro por rutas y preflight de secretos (`content/skills/git-workflow/scripts/git_state.py`):**
   - La invocación sin subcomando de `git_state.py` permanece de solo lectura. Se añade el subcomando `commit`: `python3 git_state.py commit --repo <dir> -m <msg> [--unreviewed-fix <reason>] -- <path>...`.
   - Rechaza rutas vacías, `.`, rutas que apunten a un directorio existente, o flags `-a`/`-A`/`--no-verify`/`-n`; captura `git diff --cached --name-only` previo; registra temporalmente con `git add -N -- <untracked>` los archivos nuevos en `<paths>` (revirtiéndolos con `git reset -q -- <untracked>` si falla el preflight o el commit); verifica con `gitleaks` (si está instalado) y con el escáner integrado de patrones/entropía exclusivamente sobre el diff/contenido de `<paths>` que no entren secretos; añade `Unreviewed: <reason>` cuando se solicita `--unreviewed-fix`; ejecuta `git commit --only -m <msg> -- <paths>`; y comprueba que las rutas previamente preparadas fuera de `<paths>` permanezcan en el índice.
5. **`H9` — Planificador determinístico de limpieza (`content/skills/flow-close/scripts/cleanup_branches.py`):**
   - Inspecciona `git worktree list --porcelain`, `git status --ignored --porcelain`, comparación `cmp` / `diff` de archivos ignorados fuera de la lista permitida contra el checkout principal, procesos activos vía `lsof -d cwd -Fn` (reteniendo el worktree cuando `lsof` no está disponible o falla), ramas protegidas o activas (rama base, ramas de `Environments`, base de cualquier PR abierto, ramas en otro worktree, y ramas nombradas en checkpoints/planes activos o cierres pendientes `< 24h`) y estado de PRs (`headRefOid`) vía `gh`, devolviendo JSON con las acciones seguras (`remove_worktree`, `delete_local_branch`, `delete_remote_branch`, `keep` con su razón) y ejecutándolas solo cuando recibe `--apply`.
6. **`H10` — Gramática y splicer de `## Business rules` (`content/skills/flow-plan/scripts/check_change.py` + `business_rules.py`):**
   - `check_change.py` (cuando no recibe `--archive`) valida que cada línea no vacía bajo `## Business rules` en `proposal.md` respete la gramática canónica de `flow-plan/references/business-rules.md` y `flow-figma-design/references/rule-cards.md`: `None: ...`, `ticket section: <key> · <heading>`, `ticket write: <YYYY-MM-DD> · <moment> · <key>`, `rule: <the rule> · screen <name, or none> · <status> · found by <drawer | reviewer | research | plan | build | verify>` (rechazando el estado `stated:` en `proposal.md`), `card stale: <screen> · <rule>` y `frame stale: <node> · <rule>`.
   - `business_rules.py` ofrece `render-section` y `splice-description` (usando el encabezado localizado de `ticket section: <key> · <heading>` o `--heading`): cuando ningún marcador existe en la descripción, agrega la sección `<heading>` ... `<heading> · end` al final tras una línea en blanco; cuando ambos existen una sola vez, reemplaza únicamente el bloque entre ellos; cuando falta uno o alguno aparece duplicado, rechaza editar con código distinto de cero para que el llamador use un comentario; y verifica que el prefijo y sufijo fuera de los marcadores queden idénticos byte a byte.
7. **`H11` — Validadores estáticos de *ready list* (`content/skills/create-prd/scripts/check_prd.py` y `content/skills/create-issue/scripts/check_issue.py`):**
   - `check_prd.py` valida en inglés y español las 9 secciones en orden (`Problem`/`Problema`, `Goals`/`Objetivos`, `Non-goals`/`Fuera de alcance`, `User stories`/`Historias de usuario`, `Requirements`/`Requisitos`, `Acceptance criteria`/`Criterios de aceptación`, `Success metrics`/`Métricas de éxito`, `Open questions`/`Preguntas abiertas`, `Version history`/`Historial de versiones`), encabezados `### R<n> · P<0-2>[(proposed|propuesta)] · <state>` (con estados y veredictos en inglés o español), unicidad de `R<n>`, ausencia de reglas en tablas Markdown, y presencia de `Source:`/`Fuente:` y `Verdict:`/`Veredicto:`.
   - `check_issue.py` valida en inglés y español el borrador según `--type <story|bug|task|spike>`, secciones requeridas, conteo `<= 7` de criterios de aceptación (con excepción `--finding-group` para un grupo de hallazgos), y que la sección de solution hint esté marcada como no vinculante con 1 oración de dirección y `<= 5` ubicaciones.

## Impact

- Afecta `tooling/management/references.go`, `tooling/cli/doctor_project.go`, `content/guidance/global.md` y 7 skills en `content/skills/`.
- No rompe instalaciones existentes ni requiere migración de datos; los subcomandos y scripts nuevos son aditivos y se validan en CI local antes de cualquier despliegue.

## Recovery

- Revertir el commit o pull request del cambio restaura `content/guidance/global.md`, los skills y los paquetes Go a su estado previo sin pérdida de datos locales ni de configuración de usuario.
