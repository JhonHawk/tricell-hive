# Diseño

## Contexto verificado

Revisado en `82f2dcc` el 2026-09-29.

- **Contrato del id.** `integrations/agents/agents.go` exige que `name` sea igual al nombre del archivo, y `tooling/management/plan.go:238` acepta cualquier carpeta de un nivel bajo `content/agents/`. El id instalado en cada host sale del nombre del archivo, no de la carpeta.
- **Retiro al desplegar.** `tooling/management/ownership.go:179-207` agrega, en una instalación, un recurso `Retire` por cada registro del estado cuyo destino ya no aparece en la versión nueva, para los hosts seleccionados. Un id nuevo cambia a la vez el origen (`content/agents/<carpeta>/<id>.md`) y el destino, así que la comprobación de "shadowed or relocated" (`ownership.go:157`), que solo compara registros con el mismo origen, no se dispara. Esto se lee del código; AC4 lo comprueba con la vista previa real.
- **Consumidores del id.** El id aparece en: los 20 archivos de rol (`name` y encabezado); `content/skills/flow-build/SKILL.md`, `flow-build/references/{verification,browser-automation}.md`, `flow-plan/SKILL.md`, `flow-plan/references/{delivery-decisions,plan-format,plan-review}.md` y `flow-research/SKILL.md`; `_support/docs/architecture/agent-delivery.md` y `_support/docs/harness-engineering/harness-audit-rules.md`; `integrations/agents/agents_test.go` (rutas y tabla de esfuerzo por rol), `integrations/target/catalog_test.go` (ruta sintética), `tests/skills/harness_audit_rules_test.py` (ruta del rol) y `tests/fixtures/flows/cases.json` (casos de humo y de delegación). `content/guidance/global.md` y `integrations/agent-profiles.json` no citan ids.
- **Sin CI en el repositorio.** No existe `.github/workflows/`; las comprobaciones son locales.
- **Investigación previa.** [agent-names.research.md](../../../sessions/2026-09-26-hive-agent-names/agent-names.research.md) documenta las reglas de nombres de cada host: minúsculas y guiones son válidos en todos. Un prefijo evita el choque con el nombre de puesto sin prefijo, pero no lo hace imposible.

## Enfoque

Un solo cambio coordinado. Los productores (archivos de rol) y los consumidores (skills, documentación, pruebas y la instalación en los hosts) están todos bajo el control de este repositorio, y el gestor ya retira los ids ausentes. No hace falta un periodo de alias: ningún consumidor externo depende de los ids viejos, y una sesión abierta con los nombres viejos los conserva hasta reiniciarse, como con cualquier actualización.

Cada rol se renombra con `git mv` dentro de su carpeta actual para conservar la historia del archivo.

## Tabla de nombres

| Actual | Nuevo |
| --- | --- |
| `development/backend-developer` | `hive-build-backend` |
| `development/frontend-developer` | `hive-build-frontend` |
| `development/kotlin-multiplatform-developer` | `hive-build-kmp` |
| `development/database-specialist` | `hive-build-data` |
| `ops/devops-engineer` | `hive-build-infra` |
| `design/solution-architect` | `hive-design-architecture` |
| `design/visual-designer` | `hive-design-ui` |
| `docs/sdd-spec-writer` | `hive-write-spec` |
| `review/sdd-explore` | `hive-research` |
| `quality/test-engineer` | `hive-write-tests` |
| `quality/sdd-verify` | `hive-verify-change` |
| `quality/review-task` | `hive-verify-task` |
| `quality/performance-engineer` | `hive-tune-performance` |
| `quality/state-fetcher` | `hive-read-state` |
| `review/review-plan` | `hive-review-plan` |
| `review/review-code` | `hive-review-code` |
| `review/review-security` | `hive-review-security` |
| `review/review-ux` | `hive-review-ux` |
| `review/review-harness` | `hive-review-harness` |
| `review/review-refuter` | `hive-refute-claim` |

`hive-design-system` se descartó para el arquitecto porque se lee como "sistema de diseño" de interfaz.

## Descripciones

Las descripciones no citan ids de otros agentes, para que un renombrado futuro no las rompa. Los límites con roles vecinos se describen por su trabajo.

- **H1 `hive-verify-change`** (antes `sdd-verify`; confundía el alcance con `review-task`): "Independently verify a completed change's behavior end to end, including in-vivo and browser runs, and report evidence without fixing source. Use after a change is implemented, whatever its number of tasks and with or without a plan, separate from the implementer; not for verifying one task of a retained plan during the build."
- **H2 `hive-build-infra`** (antes `devops-engineer`; omitía la verificación de despliegues del paso 6 de su cuerpo): "Implement operational, CI, and infrastructure changes using existing deployment mechanisms, or verify, without changing any deployment, that it serves the expected build. Use for CI pipelines, deployment configuration, containers, infrastructure-as-code, or read-only post-deployment verification."
- **H3 `hive-design-ui`** (antes `visual-designer`; repetía la primera frase y no decía que puede implementar, paso 4): "Design or refine interfaces using the product's visual language and real user tasks, delivering the design or its implementation with rendered evidence. Use when the layout, hierarchy, or visual consistency of a screen, component, or shared pattern is not yet settled."
- **H4 `hive-research`** (antes `sdd-explore`; repetía "project and delivery state" y se solapaba con el lector de estado): "Investigate a bounded technical question or the current project and delivery state, and return an evidence-backed synthesis. Use for delegated read-only research into code or conflicting sources, instead of a host's built-in explorer; not when the task is only to collect the current state of named tickets, branches, or runs."
- **H5 `hive-design-architecture`** (antes `solution-architect`; "specifications" chocaba con el redactor de especificaciones): sustituir "including specifications for proposed changes" por "including API, event, or persistence contract specifications for proposed changes"; el resto queda igual.
- **H6 `hive-build-kmp`** (antes `kotlin-multiplatform-developer`; le faltaba el límite que tienen backend y frontend): "Implement Kotlin Multiplatform changes across the project's supported targets. Use for shared or platform-specific Kotlin Multiplatform code whose interface is settled."

Los apóstrofos tipográficos de las demás descripciones no se tocan (N1 quedó fuera por D1-A).

## Referencias en el texto

En skills y documentación, cada mención de un id viejo pasa al id nuevo tal cual, sin reescribir la frase. Donde una frase describa el rol por su función, como "the UI reviewer" junto a `review-ux`, se conserva esa descripción.

Excepción: la tabla "Evidence through 2026-09-22" de `_support/docs/architecture/agent-delivery.md` registra llamadas observadas con los ids de entonces. Sus filas por host se conservan y la tabla recibe una nota: "Ids as observed, before the 2026-09-29 rename to `hive-<verb>-<object>`." Cambiarlas afirmaría haber observado ids que nunca se ejecutaron.
