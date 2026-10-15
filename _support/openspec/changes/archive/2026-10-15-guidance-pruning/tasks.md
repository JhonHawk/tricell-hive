# Tareas

Base del build: `0b2e0bac` (`feat/guidance-pruning`, worktree `.claude/worktrees/feat-guidance-pruning`), 2026-10-15.

Ejecución: T2–T4 en el hilo principal, porque cada edición son una o dos frases y un brief sería más largo que el cambio; T5 delegada a `hive-research`. Enfoque `check` con las observaciones `rg` de cada AC y `go test ./tests/content/...`.

### T1 — Inventario con procedencia

- [x] `inventory/global-md-rules.md` (188 unidades) y `inventory/flow-skills-rules.md` (405 unidades) existen con clase de procedencia, commit, sesión o decisión, duplicados, conflictos y comportamiento por defecto, más un encabezado con fecha, base y límites. Correcciones conocidas, anotadas en `design.md`: la limpieza de otras ramas es `observed` (`8e5fba29`), no `unknown`; dos pares marcados "identical" no lo son; las cifras de duplicados son semánticas, no literales.

**Closes:** ninguno (evidencia retenida por D2-A).

**Depends on:** ninguna.

**Locations:** `inventory/`.

**Execution:** dos hijos `hive-research` (2026-10-15); localizadores decisivos verificados por el orquestador.

**Verification:** `ls inventory/` muestra los dos archivos; `rg -c "^\| [0-9]+ \|" inventory/global-md-rules.md` devuelve 188.

### T2 — C1 y C3 en `global.md`

- [x] `global.md:31` conserva solo "A review requested on its own does not authorize fixes."; `global.md:108`, `flow-close/SKILL.md:8,74` y `artifact-layout.md:41` dicen "merged work branches, from this work or earlier"; `flow-build/SKILL.md:18` añade "code review" a la lista que manda leer `verification.md` (DD2, DD3).

**Closes:** AC1, AC2.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:31,108`; `content/skills/flow-close/SKILL.md:8,74`; `content/skills/workspace-conventions/references/artifact-layout.md:41`; `content/skills/flow-build/SKILL.md:18`; comprobación sin cambio en `content/skills/flow-build/references/verification.md:41,43,61,71`.

**Execution:** hilo principal.

**Test approach:** check: observaciones `rg` de AC1 y AC2; `go test ./tests/content/...`.

**Changes:** aplicar DD2 y DD3 con el texto exacto de `design.md`; antes, buscar "fixed within it" y "merged work branches" en `content/agents/`, `site/`, `_support/docs/architecture/` y `tests/pilot/` y anotar los lugares aquí (DD6; la primera ronda de revisión ya buscó en `content/agents/`, `git-workflow/SKILL.md` y `_support/docs/architecture/` sin más menciones del alcance viejo). Registrar aquí la oración retirada de la línea 31 con su casa restante (`verification.md:41,43`).

**Registro (2026-10-15):** lugares revisados antes de editar: `content/agents/`, `site/`, `_support/docs/architecture/`, `tests/`, `README.md`, `llms.txt`: ninguna mención de "fixed within it"; "merged work branches" solo en los tres archivos editados. Oración retirada de `global.md:31`: "findings from a review run as a gate of authorized implementation are fixed within it, as `flow-build` sets"; casa restante: `verification.md:41,43` (severidad y refutación) y `:61,71` (UX), leída por `flow-build:18` (ahora también para una revisión de código). Tras el veredicto, `hive-verify-change` y `hive-verify-task` señalaron que "merged pull-request branches" dejaba fuera las ramas preview sin PR y dentro a una rama de entorno cabeza de un PR de promoción; se cambió a "merged work branches, from this work or earlier" en los cuatro sitios (fix pequeño y local dentro de la tarea, cubierto por las observaciones de AC2).

**Verification:** `rg -n "are fixed within it" content/guidance/global.md` sin resultados; `rg -c "from this work or earlier"` sobre `global.md`, `flow-close/SKILL.md` y `artifact-layout.md` devuelve 1, 2 y 1; `rg -n "code review" content/skills/flow-build/SKILL.md` devuelve la línea 18; `go test ./tests/content/...` en verde.

### T3 — Reducciones sin cambio de comportamiento

- [x] `global.md:93` queda en una oración más el puntero existente; `global.md:108` y `:23` pierden las oraciones de DD4; `flow-research/SKILL.md:40` pierde la oración duplicada con `reporting.md:77`.

**Closes:** AC3, AC4.

**Depends on:** T2 (mismo archivo; secuencial).

**Locations:** `content/guidance/global.md:23,93,108`; `content/skills/flow-research/SKILL.md:40`; sin cambio en `content/skills/flow-plan/references/change-records.md:7`, `content/skills/workspace-conventions/references/artifact-layout.md:41`, `content/skills/flow-close/references/reporting.md:77`.

**Execution:** hilo principal.

**Test approach:** check: observaciones `rg` de AC3 y AC4; `go test ./tests/content/...` (incluye `skill_links_test.go`: el puntero de `global.md:95` se conserva).

**Changes:** aplicar DD4 con el texto exacto de `design.md`. Registrar aquí cada oración retirada con su casa restante. Comprobar que `tests/pilot/` no detecta por esas frases (`rg -F` de cada una en `tests/`).

**Registro (2026-10-15):** retiradas: de `global.md:93`, "Without a specs repository it is `_support/openspec` in the repository; with one, `<specs-repo>/openspec`." y "It is always versioned: never place it in a workspace outside Git or in an ignored path." (casa: `change-records.md:7`, leída por el puntero de `global.md:95`); de `global.md:108`, "Repeating a test may not reproduce the same evidence." (casa: `artifact-layout.md:41`, leída por `flow-close:10` antes de limpiar); de `global.md:23`, "Reporting details and closing categories live in [reporting](...)" (descriptiva; el puntero de `global.md:17` sigue); de `flow-research/SKILL.md:40`, "Start the chosen flow without a new command." (casa: `reporting.md:77`). `rg -F` en `tests/`: solo `tests/fixtures/workspace-conventions/baseline/global.md:38`, copia congelada que `tests/pilot/main.go:526` lee como entrada, no un detector.

**Verification:** `rg -F "never place it in a workspace outside Git" content/guidance/global.md` sin resultados y `rg -c` del mismo texto en `change-records.md` devuelve 1; `rg -F "Repeating a test may not reproduce" content/guidance/global.md` sin resultados; `rg -F "Reporting details and closing categories" content/guidance/global.md` sin resultados; `rg -F "Start the chosen flow without a new command" content/skills/flow-research/SKILL.md` sin resultados y en `reporting.md` una línea; `go test ./tests/content/...` en verde.

### T4 — Sitio, changelog, tamaños y verificación del cambio

- [x] `site/.../build.md:43` describe C1 en una frase (`close.md:31,54` ya describe C3 y no cambia); `CHANGELOG.md` tiene la entrada bajo `[Unreleased]`; los tamaños antes y después están anotados aquí; el candidato pasa el gate compartido y `hive-verify-change` contra AC1–AC7.

**Closes:** AC5, AC6, AC7.

**Depends on:** T2, T3.

**Locations:** `site/src/content/docs/flows/build.md:43`; `CHANGELOG.md`; `_support/docs/architecture/repository-and-distribution.md` solo si describe el reparto global/referencias tocado.

**Execution:** hilo principal; `hive-verify-change` como hijo.

**Test approach:** check: `rg -n "severity" site/src/content/docs/flows/build.md`; `rg -n "pull-request branches" CHANGELOG.md`; `wc -c content/guidance/global.md`.

**Changes:** una frase en `build.md:43` ("by severity, as Build sets"); entrada de changelog bajo "Changed" (hallazgos de revisión por severidad solo bajo Build; borrado de ramas de pull request fusionado del repositorio, con la observación de `8e5fba29`; texto de `global.md` reducido a una casa por regla) y nada bajo "Removed", porque ninguna regla desaparece. Anotar aquí `wc -c content/guidance/global.md` y la huella de lectura (`go test -v -run 'TestGlobalGuidanceSize|TestGuidanceReadingFootprint' ./tests/content/`) antes (27318) y después.

**Verification:** las observaciones de AC6 y AC7; `wc -c content/guidance/global.md` menor que 27318 (AC5); veredicto de `hive-verify-change` con AC1–AC7 `met`; el gate compartido de la tabla.

### T5 — C2: research del punto de "ticket hecho" por proyecto

- [x] Un hijo `hive-research` entrega, en solo lectura, cómo cierran hoy cam, carvento y dened (estados de Linear y Jira, quién mueve el ticket y cuándo), qué fijan `verification.md:88` y `flow-close:14,56,66`, y dos o tres formas de declarar el punto de cierre en `## Hive` (nombre del ajuste, valores, valor por defecto); el orquestador escribe las opciones fundamentadas en `design.md` (DD5, apartado C2b), presenta la tarjeta C2b y registra la respuesta en `proposal.md`.

**Closes:** AC8.

**Depends on:** ninguna (independiente de T2–T4).

**Locations:** `content/skills/flow-plan/references/project-settings.md`; `content/skills/flow-build/references/verification.md:88`; `content/skills/flow-close/SKILL.md:14,56,66`; `AGENTS.md` y memoria de Grok/Engram de `/Users/jmartinez/Development/projects/comsys/{cam,carvento,dened}` (solo lectura, proyectos externos). Fuentes de estado de tickets: Linear por la herramienta `linear` cuando esté autenticada y, si no, los estados citados en los registros archivados y la memoria del mantenedor; Jira por `acli` como en dened; el brief dice qué fuente no estuvo disponible.

**Execution:** delegada a `hive-research` con brief que fija lo ya sabido (C2 del usuario) y la parada (cuando las tres formas estén fundamentadas o una fuente falte).

**Test approach:** check: `rg -n "C2b" design.md proposal.md` devuelve las opciones y la respuesta.

**Changes:** ninguna edición de guía; la tarjeta ofrece "definir el ajuste" (otro cambio), "dejar el texto actual" o "seguir investigando".

**Verification:** `design.md` contiene el apartado "C2b" con las opciones y `proposal.md` contiene "- C2b" con la respuesta del usuario.

## Verificación y revisión humana

| Gate | Momento y entorno | Mecanismo y evidencia esperada | Autoridad / condición pendiente |
| --- | --- | --- | --- |
| Pruebas de contenido | Tras cada tarea, worktree `feat/guidance-pruning` | `go test ./tests/content/...` en verde | Cubierta por la implementación |
| Suite completa y tamaños | Candidato final (gate compartido; no se repite por tarea) | `go test ./...` en verde; `wc -c content/guidance/global.md` y `go test -v -run 'TestGlobalGuidanceSize\|TestGuidanceReadingFootprint' ./tests/content/` anotados antes y después | Cubierta por la implementación |
| Revisión del plan | Antes de declarar el plan listo | `hive-review-plan`, dominio guía y skills; segunda ronda sobre esta versión | Parte de la planificación |
| Verificación independiente | Al terminar T2–T4 | `hive-verify-change` contra AC1–AC6 | Cubierta por la implementación |
| Revisión de código dedicada | — | No aplica: decisión vigente para este repositorio | Resuelta |

Sin efecto de UI renderizada: no aplica `hive-review-ux`.

## Entrega

Ver `proposal.md`, sección Entrega.

## Estado de revisión y progreso

Plan escrito el 2026-10-15 sobre `fd352a02` y reescrito el mismo día sobre `0b2e0bac` tras la primera ronda de `hive-review-plan` (13 hallazgos: 2 bloqueantes, 10 por corregir, 1 menor) y el análisis adversarial (`design.md`). Hallazgos reconciliados: AC1 sin script y "30 pares" (resuelto: conteo propio con clave de doce palabras, lote reducido a cuatro reducciones concretas); DD2(c) mal atribuido (resuelto: C3 amplía global, evidencia `8e5fba29`); DD2(b) reescribía un criterio observado (resuelto: C2 pasa a research, sin edición); DD2(a) debilitaba la R1 (resuelto: C1 retira la cláusula de global y deja la severidad en `verification.md`); DD3 rompía Cursor (resuelto: punteros intactos); DD7 exigía commit (resuelto: sin ventana de medición); DD4/DD6/lote 2 y 3 (resuelto: no ahora); cifras de `proposal.md` (corregidas); T1 verificación (corregida); CHANGELOG y `site/` (T4). Segunda ronda (misma fecha): 2 bloqueantes y 5 por corregir, todos aplicados: B1 alcance de C3 acotado a ramas de pull request fusionado (no entorno) y S1 las tres frases que citan la concesión (`flow-close:8,74`, `artifact-layout:41`) entran en T2; B2/S2 AC4–AC8 reescritos con comandos exactos y condiciones no disyuntivas; S3 `flow-build/SKILL.md:18` gana "code review" para que un build en `hold` con revisión llegue a la severidad; S4 registrado como decisión abierta C1b (si C1 también cubre la regla de hallazgos incidentales en tareas directas; ver `proposal.md`); S5 T5 nombra fuentes y rutas y AC8 exige las opciones en `design.md`; menores (puntero de `flow-close` en la línea 10, `wc -c` en lugar de `TestGlobalGuidanceSize`, `rg` "sin resultados", `go test ./...` solo como gate compartido, `global.md:32` fuera del alcance). C1b-A y D4-A respondidas; build ejecutado el 2026-10-15 sobre `0b2e0bac` en el hilo principal (T2–T4) y con un hijo `hive-research` (T5). Veredictos: `hive-verify-task` T2 AC1 met, AC2 met; T3 AC3 met, AC4 met; `hive-verify-change` AC1–AC7 met, con el hallazgo de alcance de ramas corregido después (ver T2) y las observaciones de AC2 y AC7 repetidas sobre el texto final. Modelos: implementador Mythos 5.1 (observado, hilo principal); verificadores perfil `verifier` = opus (configurado, no observado); difieren. Tamaños: `global.md` 27 318 → 26 955 bytes (−363); huella de lectura del escenario base −42,55 % frente a `d5164f71`. Gates: `go test ./...` en verde sobre el candidato final; `pnpm --dir site check` 0 errores. C2b-A registrado en `proposal.md`. Entrega: `hold`, sin commit por instrucción del usuario; el diff queda en el worktree `feat/guidance-pruning` (7 archivos, +25/−10).
