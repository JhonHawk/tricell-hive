# master-promotion-guard — master solo recibe PRs desde development

| Campo | Valor |
| --- | --- |
| Estado | Completed · integrado en #345 (`959e8876`) · ruleset 26428778 activo |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `2653b0b1` · rama propuesta `chore/master-promotion-guard` en `.claude/worktrees/` |

## Decisiones

2026-11-03
- ¿Cómo se hace cumplir en GitHub que a `master` solo se llegue por PR desde `development`? Con un ruleset en `master` y un check obligatorio de GitHub Actions que falla cuando el PR no sale de `development` de este mismo repositorio. (D1 A)
- ¿Alguien puede saltarse el ruleset en un hotfix urgente? Sí, el rol admin, y solo mergeando un PR (modo `pull_request`). El push directo queda bloqueado también para admin. (D2 B)
- ¿El build puede crear el ruleset y abrir PRs de prueba hacia `master`? Sí: un PR desde una rama de prueba y uno desde `development`, los dos cerrados sin mergear, y la rama de prueba se borra. (D3 A)
- ¿Cómo se entrega? En modo `automatic`: push, PR y merge a `development` sin parar, y después las tareas remotas. (D4 A)
- ¿Qué revisión de código corre? Claude Code `/code-review`, antes del push. (D5 A)
- ¿Cómo se cierra el hueco de que otro workflow publique un job falso `source-branch`? Los workflows de PRs desde forks ahora esperan siempre la aprobación del maintainer (`approval_policy` pasó de `first_time_contributors` a `all_external_contributors`). (D6 A)

## Solicitud

> Podemos poner la regla de que a master solo se llega desde pr de development?

Sin ticket.

## Objetivo

Hoy `master` no tiene branch protection ni rulesets (`gh api repos/JhonHawk/tricell-hive/rules/branches/master` devuelve `[]`), así que cualquiera con permiso de escritura puede hacer push directo, force-push o mergear un PR desde cualquier rama. La regla existe solo en la guía (`content/skills/git-workflow/SKILL.md:29`, `CONTRIBUTING.md:67`). Cuando el cambio termine, GitHub rechazará el push directo, el force-push y el borrado de `master`, y bloqueará el merge de cualquier PR cuyo origen no sea `development` de este repositorio. El admin podrá saltarse esa restricción, pero solo mergeando un PR.

## Alcance y aceptación

Dentro del alcance: el workflow del check, el ruleset de `master`, y `CONTRIBUTING.md` y `CHANGELOG.md`. Fuera del alcance: proteger `development`, reintroducir CI de tests, cambiar la guía distribuida en `content/` y tocar las etiquetas de release.

- AC1 — Un PR hacia `master` cuya rama de origen no es `development` de este repositorio recibe el check `source-branch` en `failure` · hoy es falso porque ningún check corre en los PRs hacia `master` (no existe `.github/workflows/`).
- AC2 — Un PR de `development` hacia `master` recibe el check `source-branch` en `success` · hoy es falso porque el check no existe.
- AC3 — Las reglas activas de `master` exigen un PR, exigen el check `source-branch` emitido por GitHub Actions, bloquean el force-push y el borrado, y solo permiten saltarlas al rol admin mediante un PR · hoy es falso porque `rules/branches/master` devuelve `[]`.
- AC4 — `CONTRIBUTING.md` describe cómo GitHub hace cumplir la promoción y ya no dice que el repositorio no tiene CI · hoy es falso porque `CONTRIBUTING.md:15` dice "The repository has no CI" y `CONTRIBUTING.md:67` solo enuncia la regla.

## Prerrequisitos

Ninguno: el build no necesita nada de fuera del repositorio. Usa la sesión autenticada de `gh` del maintainer, que es owner del repositorio (plan Pro, repositorio público).

## Contexto verificado

- 2026-11-03: `master` sin protección (`branches/master/protection` → 404), `rulesets` → `[]`.
- Los 4 PRs históricos hacia `master` salieron de `development` (#101, #184, #309, #343).
- Rama por defecto: `development`. Según la documentación de GitHub Actions, `pull_request_target` toma el workflow y `GITHUB_SHA` de la rama por defecto, así que el check corre desde que llega a `development` y un PR no puede modificarlo.
- Los tipos de regla de un ruleset en la API REST (creation, update, deletion, required_linear_history, pull_request, required_status_checks, non_fast_forward…) no incluyen ninguna que restrinja la rama de origen.
- `a6a4b3fe` (2026-09-30) quitó GitHub Actions porque la CI por PR costaba más de lo que detectaba (Engram `ci/decision-local-only`). Este workflow no corre tests ni hace checkout, y solo se dispara en PRs hacia `master`.
- `content/skills/flow-deploy/references/urgent-fixes.md:7` acepta un PR de hotfix hacia la rama de producción solo donde la política de ramas lo permita; el bypass de admin por PR (D2 B) lo permite.

## Entrega

Modo `automatic`: revisión con Claude Code `/code-review` sobre el candidato, push de `chore/master-promotion-guard`, PR hacia `development` y merge sin parar. Después corren T4a, T3 y T4b en GitHub, que D3 A autoriza. Esto no autoriza mergear nada hacia `master`, ni promover, ni publicar una release, ni cambiar otra configuración del repositorio.

## Estado de revisión y avance

2026-11-03 — plan escrito. `hive-review-plan` no encontró P0 ni P1, y se aplicaron sus hallazgos H1–H7 (ver `tasks.md`). Fuera de este plan quedó un P3 que solo se lista: `content/skills/git-workflow/SKILL.md:25` pide traer la base a la rama antes de mergear, y en un PR de promoción eso significa traer `master` a `development`. Las promociones anteriores no lo hicieron, y con `strict: false` GitHub no lo exige.
