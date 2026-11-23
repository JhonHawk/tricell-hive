# audit-prefix — Las skills de auditoría empiezan con `audit-`

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #602, merge 48824e10 |
| Tracker | ninguno |
| Git | tricell-hive · base `development` en `ba160011` · rama `refactor/audit-prefix` |
| Playbook | refactor |

## Decisiones

2026-11-22
- **D1 A** — ¿Se renombran las tres skills de auditoría con el verbo al inicio? Sí: `comment-audit` → `audit-comments`, `test-audit` → `audit-tests`, `harness-audit` → `audit-harness`, para que queden ordenadas y agrupadas como `create-*` y `flow-*`, con el objeto en plural donde el singular se leería como "auditar uno".
- **Delivery A** — `automatic`; sin revisión de código, por ser un cambio mecánico.
- **Supuesto** — Los nombres viejos se retiran sin alias, como `flow-report` (descartado: mantener un alias; duplicaría las descripciones cargadas en cada sesión). Se revierte devolviendo los nombres en `retiredSkills` y las carpetas.
- **Supuesto** — Los encabezados de requisito de las specs no cambian, solo su texto (descartado: `REMOVED` + `ADDED` para renombrar el encabezado `Personal rules mode of harness-audit`; `apply_deltas.py` no tiene `RENAMED` y el encabezado es un identificador, como `… through the local-instructions skill`).
- **Supuesto** — La capacidad de spec `comment-audit` conserva su nombre: es una capacidad, no la skill.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

> siguiendo el patron de dominio, los skills de audit tienen el concepto al final, qué pasaria si los pasamos al inicio? … para que esten ordenados y agrupados

## Objective

Hoy las tres skills de auditoría llevan el verbo al final y quedan dispersas al ordenar por nombre. Al terminar se llaman `audit-comments`, `audit-tests` y `audit-harness`, y nada que un agente o una persona lea para usar Hive nombra las viejas.

## Scope and acceptance

Dentro: las carpetas de las tres skills, cada enlace y uso literal en `content/`, `site/`, `README.md`, `llms.txt`, `_support/docs/` (salvo `history/`), `integrations/`, `tooling/`, `tests/`, `.gitignore`, el catálogo `audit-harness-rules.md`, el `testdata` de la plantilla, el CHANGELOG y los requisitos de spec que nombran las skills. Fuera: el historial (`_support/openspec/changes/archive/`, las carpetas de cambio de otros trabajos aún sin archivar, `_support/sessions/`, `_support/docs/history/`, las entradas publicadas del CHANGELOG) y la URL y el nombre del upstream de OpenClaw.

- AC1 — `content/skills/` tiene `audit-comments`, `audit-tests` y `audit-harness`, y ninguna carpeta con el nombre viejo · false today because existen `comment-audit`, `test-audit` y `harness-audit` · Observed in: `ls content/skills`
- AC2 — `TestNoRetiredSkillIsNamed` informa cualquier mención de los tres nombres viejos en lo que un agente o una persona lee · false today because no están en `retiredSkills` · Observed in: `go test ./tests/content/`
- AC3 — `hive-review-harness` en OpenCode carga `audit-harness/SKILL.md` · false today because el mapa nombra `harness-audit/SKILL.md` · Observed in: tests de `integrations/opencode`
- AC4 — La suite completa y el CI local pasan · Observed in: `make` o los jobs del workflow de CI

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

- `git show --stat eb23ae75` — precedente `flow-report` → `create-report`: mismo alcance y forma.
- `tooling/management/apply.go` `retireSkillCaches` — una actualización retira entera la carpeta de una skill que ya no se distribuye.
- `content/skills/comment-audit/references/project-skill.md` — la skill que genera en los proyectos se llama `delete-comment`; no depende del nombre.

## Effects

None: the tasks take no effect on an external system.

## Delivery

`automatic`: push, PR y merge a `development` sin parada. Sin revisión de código (cambio mecánico). No autoriza un release.

## Review status and progress

2026-11-22 — carpeta escrita; sin revisión de plan (cambio sin plan).
