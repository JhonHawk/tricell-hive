# Revisión del plan a medio camino, rol de documentación y costo de la revisión (#60)

| Campo | Valor actual |
| --- | --- |
| Estado | Completado · integrado por la [PR #62](https://github.com/JhonHawk/tricell-hive/pull/62) en `development` |
| Tracker · GitHub Issues | • [#60 — Comparación con OpenSpec: completitud al cerrar, revisión de plan con dueño, hive-write-spec huérfano y vigilancias](https://github.com/JhonHawk/tricell-hive/issues/60) |
| Git | automática · rama `feat/gh-60-plan-revision-and-write-docs` → `development` · sin parada humana |
| Verificación | `go vet ./...` y `go test ./...` · `pnpm --dir site build` · búsqueda del nombre viejo · `/code-review` de Claude Code en la PR |
| Siguiente paso | Ninguno: cierre del registro con la PR #62 |

## Petición

> Primero validemos si se pueden cerrar algunos que no apliquen

> D1-A — Continuar con #60 R5 y R4 (Recommended)

> hagamos research, normalmente siempre selecciono el modelo frontier o mas capaz para escribirlo, y supongo que todo ese contexto le es util, pero no lo sé, confirmemos con investigación

> Conservar con contrato acotado y cambio de nombre, aunque me interesa esa parte de superpowers sobre el tema de que delegar la revisión duplica este tiempo, analiza si este tiempo mejora la calidad que justifique la inversión y volvamos a replantear qué hacer, pero ahora con la revisión

> D1-A — Conservar y medir el costo (Recommended)

(2026-10-16.) Del issue #60, leído el 2026-10-16:

> **R4 — Revisar un plan a medio camino con un solo dueño.** La conducta está repartida en `flow-build/SKILL.md:16,48`, `flow-plan/SKILL.md:18`, `questions.md:64`, `verification.md:73`, `plan-review.md:38` y `plan-format.md:20`. No se dice si una edición del plan durante la construcción vuelve a `hive-review-plan` ni cómo se mantienen alineados los `AC<n>` y los deltas. OpenSpec `update-change`: borrador en la conversación, coherencia en cualquier dirección, escribir un artefacto a la vez tras confirmar, cambio nuevo si cambia la intención. Propuesta: un bloque "Revise the plan" en `flow-plan` o `flow-build` al que las siete frases enlacen.

> **R5 — `hive-write-spec` es un rol huérfano.** Ninguna skill, integración ni herramienta lo referencia. Opciones: conectarlo desde `flow-plan` para redactar deltas y specs, o retirarlo con su razón en `_support/docs/architecture/agent-delivery.md`.

## Objetivo

Tres resultados de la misma sesión. (1) R4: una sola casa para revisar un plan retenido mientras se construye, en `flow-plan`, a la que enlazan las siete frases que hoy reparten esa conducta; dice quién edita, cómo se mantienen coherentes petición, criterios, diseño, tareas y deltas, cuándo la revisión vuelve a `hive-review-plan` y cuándo el cambio de intención es un cambio nuevo. (2) R5: la investigación mostró que el plan lo escribe siempre el hilo principal, como hacen todas las referencias, y que `hive-write-spec` sí se usa, asignado por tarea en cinco cambios para documentación con fuentes ya verificadas; el rol pasa a llamarse `hive-write-docs` con ese contrato, y deja de prometer registros de cambio. (3) Revisión de planes: 46 de 46 planes revisados cambiaron por hallazgos aceptados, pero ningún registro guarda su costo; `plan-review.md` pide anotarlo para poder reevaluar la compuerta con datos.

## Alcance y aceptación

Incluido: `content/skills/flow-plan/SKILL.md`, `content/skills/flow-build/SKILL.md`, `content/skills/flow-plan/references/{questions,plan-review,plan-format}.md`, `content/skills/flow-build/references/verification.md`, `content/agents/docs/hive-write-spec.md` → `hive-write-docs.md`, `_support/docs/architecture/agent-delivery.md`, `site/src/content/docs/flows/{plan,build}.md`, `CHANGELOG.md`.

Excluido: R2 (completitud antes de archivar, conteo 0, línea base anotada en #60); H7–H9 (vigilancias); cambios en `content/guidance/global.md` (no menciona ninguna de estas conductas; su tamaño no cambia); preferir otro modelo para `hive-review-plan` (D1-B declinado); cualquier recorte del despacho de revisores.

Restricciones: la suite `go test ./...` y el build del sitio siguen en verde; el nombre viejo no queda en `content`, `integrations`, `tooling`, `tests`, `site`, `README.md`, `llms.txt` ni `_support/docs` (los registros archivados conservan la historia); renombrar un rol sube la versión menor en la próxima release (`CONTRIBUTING.md`).

- AC1. `flow-plan/SKILL.md` tiene una sección "Revise a plan in progress" que dice quién edita, la coherencia entre petición, `AC<n>`, diseño, tareas y deltas, el criterio de re-revisión por referencia a `plan-review.md` y el caso de intención nueva, y las siete frases de R4 enlazan a ella. *Falso en la base cuando* `rg 'revise-a-plan-in-progress' content` no devuelve nada.
- AC2. El rol se llama `hive-write-docs`, su descripción y cuerpo se limitan a documentación y páginas desde fuentes que el plan ya verificó, con el registro de cambio reservado al hilo principal, y `hive-write-spec` no aparece en ningún archivo distribuido ni en `agent-delivery.md` salvo la nota del renombre. *Falso en la base cuando* `content/agents/docs/hive-write-spec.md` existe y su descripción dice "requirement specs, change records".
- AC3. `plan-review.md` pide registrar en la sección de progreso del plan la duración y los tokens de cada revisor y re-revisión cuando el host los reporta. *Falso en la base cuando* `rg -i 'duration|tokens' content/skills/flow-plan/references/plan-review.md` no devuelve nada.
- AC4. Las páginas `plan.md` y `build.md` del sitio describen la revisión del plan a medio camino y enlazan la skill; el changelog registra el renombre bajo Unreleased. *Falso en la base cuando* ninguna de las dos páginas menciona una revisión del plan durante la construcción.

## Contexto verificado

- Las siete frases de R4 leídas el 2026-10-16 en `development` dae2354e; `global.md` no menciona la revisión del plan ni el rol (26 955 bytes).
- `hive-write-spec`: sin referencias en `content`, `integrations`, `tooling`, `site`, `tests`, `README.md`, `llms.txt`; `agent-delivery.md:20,30` nombran su excepción de esfuerzo. Asignado por tarea en `2026-09-29-close-hygiene-and-validation-gates` (T2–T4), `2026-10-05-gh-6-public-readiness`, `2026-10-09-gh-9-communication-decision-flow` (T5, T6; T5 dejó celdas "Not verified" que el hilo principal rellenó), `2026-10-11-question-flow-checkpoints` (T4, T5); cero usos en los seis repos de specs de clientes.
- Referencias (clones de referencia, 2026-10-16): gentle-ai `routing.go:28-29`, superpowers `writing-plans/SKILL.md:10,165`, mattpocock `to-spec/SKILL.md:7`, OpenSpec v1.14.2: el plan o spec lo escribe el hilo principal; los hijos leen por referencia.
- Revisión de planes en los 48 registros archivados: 46 revisados, 46 con al menos un hallazgo aceptado que cambió tarea, criterio, contrato o decisión; ~733 hallazgos, 175 bloqueantes, 3 descartados; sin duración ni tokens registrados en ninguno. superpowers v5.0.6: la afirmación "sin mejora medible" no tiene datos en el repo (commit e6221a4 mide tamaño del plan, tareas y "quality scores"); v7.0.0 reintrodujo un "Builder check" con contexto nuevo. Literatura: Song arXiv 2603.12123 (revisor con contexto nuevo F1 28,6 frente a 27,1 autorrevisión y 21,7 segunda pasada) y 2610.01471 (revisor nuevo + otro modelo 56,7 % frente a 42,7 %).
- Renombre anterior (`2026-09-29-hive-agent-verb-object-names`): el gestor retira el archivo viejo e instala el nuevo en cada host (prueba T3 de ese cambio).

## Diseño propuesto

- R4: sección `## Revise a plan in progress` en `flow-plan/SKILL.md`, después de "Make the work verifiable". Contenido: borrador en la conversación y nombre de lo que cambia; solo el orquestador edita, un archivo a la vez; coherencia en toda dirección (criterio ↔ tareas `Closes:` y verificación ↔ delta; tarea `[x]` tocada sigue la regla de marcadores de `verification.md`); cambio de intención o alcance = cambio nuevo; re-revisión solo bajo el criterio de `plan-review.md`, una ronda; registro en la sección de progreso y línea al usuario como manda `questions.md`. Las siete frases conservan su texto y añaden un enlace.
- R5: `git mv` a `hive-write-docs.md`; descripción "Write documentation and site pages from sources the plan already verified…"; paso 1 nuevo: tomar los hechos solo de las fuentes nombradas y marcar "not verified" lo que falte (caso gh-9 T5); se quita la lectura de `change-records.md`. `agent-delivery.md`: nombre nuevo en `:20,30` y nota con la evidencia del renombre en el párrafo del roster.
- D1-A: una frase en `plan-review.md` para registrar duración y tokens; vigilancia en memoria y en #60: reevaluar la compuerta con ~10 planes medidos.

## Entrega

Repositorio `tricell-hive`, base `development`, modo automático (D2-B, 2026-10-16) con `/code-review` de Claude Code (D3-A): commits por tarea, push, PR a `development`, merge cuando la suite y la revisión pasen; el registro se versiona al cierre en la misma rama. Al integrar: cerrar #60 con el resumen, refrescar la instalación local. Sin release.

## Estado de revisión y progreso

Base: `dae2354e` (`development`, 2026-10-16). Sin revisión de plan delegada: cambio de guía con decisiones ya tomadas por el usuario (R4 según el issue, D1 R5, D1-A revisión) y sin contrato de código.
