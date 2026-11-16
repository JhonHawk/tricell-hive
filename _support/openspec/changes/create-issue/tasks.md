# Tasks

Base: `JhonHawk/tricell-hive`, `development` en `b8ee7913`; rama propuesta `feat/create-issue`. T1 a T5 corren en el hilo principal: son textos de reglas distribuidas, cada tarea edita hasta tres archivos y pide revisar el comportamiento completo que gobierna. T6 la corre un hijo `hive-write-docs`, porque son doce archivos de documentación con sus fuentes ya fijadas en esta carpeta. T7 corre en el hilo principal porque escribe en el tracker.

Las tareas fijan frases literales en inglés que el contenido debe llevar, para que cada criterio se pueda comprobar con una búsqueda.

### T1 — La skill y sus tipos de ticket

- [ ] Existen `content/skills/create-issue/SKILL.md` y `references/ticket-types.md` con los nueve pasos, los cuatro tipos, la lista de «listo», la pista y la regla de autoría que fija `design.md`

**Closes:** AC1, AC7, AC10, AC13, AC14, AC15, AC21, AC22

**Depends on:** none

**Locations:** `content/skills/create-issue/SKILL.md`, `content/skills/create-issue/references/ticket-types.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Escribe la skill en inglés, con `name: create-issue` y una descripción de una línea que nombra sus tres usos y cuándo no usarla. Cada enlace a una referencia dice la condición que obliga a leerla. El `SKILL.md` lleva: buscar un ticket existente con la frase "search every tracker"; "at most three questions" por ronda, con respuesta recomendada y supuestos escritos en el ticket; la sesión sin repositorio, que pregunta "which tracker and repository" y omite la pista; la confirmación del borrador solo usada sola o al desglosar, y la creación directa cuando la respuesta del usuario ya eligió el ticket dentro de un flujo; la regla de autoría; y el cierre: usada sola termina en el ticket y ofrece el siguiente paso, seguida desde un flujo hace "return to that flow". `ticket-types.md` lleva las secciones por tipo, las ocho pruebas, los límites de tamaño, y la pista "not binding" con lo que un ticket no lleva y que, si exige decidir, el ticket dice que "requires a plan" en su idioma. No copia la regla de cuándo proponer un ticket: la enlaza en la guía global.

**Verification:** `go test ./tests/content/... ./tooling/management/...` pasa. `wc -c` de cada archivo da 16.000 o menos. `rg -F` encuentra en `content/skills/create-issue/SKILL.md` cada una de: "search every tracker", "at most three questions", "which tracker and repository", "return to that flow"; y en `references/ticket-types.md` cada una de: "not binding", "requires a plan". `rg -c "^[0-9]+\. " ` sobre la lista de «listo» da ocho. El archivo instalado lo observa T7.

### T2 — Desglose, trackers y validación de un ticket existente

- [ ] Existen `references/breakdown.md`, `references/trackers.md` y `references/existing-ticket.md`, enlazadas desde el `SKILL.md` con su condición de lectura

**Closes:** AC3, AC5, AC16, AC17, AC18, AC19, AC20

**Depends on:** T1

**Locations:** `content/skills/create-issue/references/breakdown.md`, `content/skills/create-issue/references/trackers.md`, `content/skills/create-issue/references/existing-ticket.md`

**Execution:** main thread

**Test approach:** check

**Changes:** Antes de escribir `trackers.md`, lee la documentación oficial vigente de sub-issues y dependencias de GitHub, de issues padre y relaciones de Linear, y de padre, épica y enlaces de Jira, porque las cifras de la investigación vinieron de lecturas delegadas. `breakdown.md`: las tres señales con la frase "more than 7 acceptance criteria"; qué es un hijo (se integra y acepta solo); el lote mostrado completo con "one approval for the whole set"; todos los tickets en el "primary tracker" y cada hijo nombrando su repositorio; escrituras de una en una con registro y relectura; parada ante un fallo; que varios tickets de hallazgos no son un lote; y que sin supervisión se crea el ticket suelto de un hallazgo sin confirmación y nunca un lote. `trackers.md`: padre e hijos por tracker; las cuatro formas de relacionar un ticket con su origen que hoy fija la guía global (bloqueo cuando uno espera al otro, "sub-issue" para una parte aplazada, "related" en otro caso, y un comentario "by its URL" en cada ticket cuando el origen está en otro tracker); los tres medios de bloqueo con "never create" etiquetas ni estados; comprobar el permiso antes y releer después; y qué hacer con otro tracker. Nombra conceptos, no opciones de CLI. `existing-ticket.md`: comprobar autoría y permiso antes de escribir, y las tres salidas (comentario cuando la cuenta es "not the author", edición, reporte), cada escritura tras confirmación.

**Verification:** `go test ./tests/content/... ./tooling/management/...` pasa. `wc -c` de cada archivo da 16.000 o menos. `rg -F` encuentra en `references/breakdown.md` cada una de: "more than 7 acceptance criteria", "one approval for the whole set", "primary tracker"; en `references/trackers.md` cada una de: "sub-issue", "related", "by its URL", "never create"; y en `references/existing-ticket.md`: "not the author". Los archivos instalados los observa T7.

### T3 — La guía global remite a la skill

- [ ] `content/guidance/global.md` conserva cuándo proponer un ticket y su autorización, y remite a `create-issue` para el cómo

**Closes:** AC8

**Depends on:** T1, T2

**Locations:** `content/guidance/global.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En «Questions and routes», la regla de la acción suelta gana una frase: una vez respondido, crear el ticket sigue `create-issue`. En «Scope changes», la regla de tickets conserva cuándo proponer uno y sustituye los pasos de búsqueda y las cuatro formas de relación con el origen por el puntero a la skill para crear, desglosar o validar; antes de borrarlos, comprueba que T2 los dejó en `trackers.md`. «Project settings» no cambia. Anota el tamaño del archivo antes y después para el pull request.

**Verification:** `rg -F "search every tracker" content/` solo encuentra archivos bajo `content/skills/create-issue/`. `rg -n "create-issue" content/guidance/global.md` encuentra los dos punteros. `go test ./tests/content/... ./tooling/management/...` pasa, incluido el test de enlaces `skill:`. El archivo instalado lo observa T7.

### T4 — Research y su auditoría remiten a la skill

- [ ] `flow-research` y `code-audit.md` apuntan a la skill para comprobar si un ticket ya existe, crearlo y juzgar si está listo

**Closes:** AC8, AC9

**Depends on:** T1

**Locations:** `content/skills/flow-research/SKILL.md`, `content/skills/flow-research/references/code-audit.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En el paso 4 de research, la búsqueda antes de proponer un ticket se conserva como momento y pasa a remitir a la skill para el cómo. En el paso 5, crear un ticket sigue siendo una acción suelta y gana el puntero. «Assess a chosen ticket» conserva impacto, complejidad y dependencias, remite a la lista de «listo» de la skill, y mantiene que un ticket no listo se resuelve ahí, en rondas. En `code-audit.md`, crear los tickets de los hallazgos elegidos sigue la skill y cada uno es un ticket suelto; se conserva lo que el cuerpo añade para un hallazgo, y el bloque de cierre de la auditoría no cambia.

**Verification:** `rg -F "Search each declared tracker" content/skills/flow-research/` no encuentra nada. `rg -F "that names more than one change" content/skills/flow-research/SKILL.md` no encuentra nada. `rg -n "create-issue" content/skills/flow-research/` encuentra los punteros. `go test ./tests/content/... ./tooling/management/...` pasa. Los archivos instalados los observa T7.

### T5 — Close, delegación sin supervisión y plan

- [ ] `flow-close` apunta a la skill sin superar su límite, `unattended-delegation` acota el ticket sin supervisión y `flow-plan` trata la pista como dato por verificar

**Closes:** AC7, AC8, AC18

**Depends on:** T1

**Locations:** `content/skills/flow-close/SKILL.md`, `content/skills/unattended-delegation/SKILL.md`, `content/skills/flow-plan/SKILL.md`

**Execution:** main thread

**Test approach:** check

**Changes:** En «See the findings» de `flow-close`, la búsqueda por hallazgo y la frase sobre duplicados y relación se sustituyen por el puntero, y «A ticket answer opens it» pasa a crear el ticket como fija la skill; la frase que termina de nuevo con el pie del cierre cuando cada hallazgo tiene respuesta no se toca, y el archivo no puede crecer. En `unattended-delegation`, abrir un ticket por un hallazgo queda como un ticket suelto, nunca un padre con hijos. En `flow-plan`, una frase dice que la pista ("hint") de un ticket es un dato por verificar, no una decisión.

**Verification:** `wc -c content/skills/flow-close/SKILL.md` da 16.000 o menos. `rg -F "after searching each declared tracker" content/skills/flow-close/SKILL.md` no encuentra nada, y `rg -F "end again with the close footer" content/skills/flow-close/SKILL.md` sigue encontrando la frase. `rg -n -i "hint" content/skills/flow-plan/SKILL.md` encuentra la frase nueva. `go test ./tests/content/... ./tooling/management/...` pasa. Los archivos instalados los observa T7.

### T6 — Sitio, listas y arquitectura

- [ ] El sitio, `llms.txt`, el registro de cambios y la lista de hogares canónicos describen la skill

**Closes:** AC11

**Depends on:** T1, T2, T3, T4, T5

**Locations:** `site/src/content/docs/flows/create-issue.md` (nuevo), `site/astro.config.mjs`, `site/src/content/docs/flows/overview.md`, `site/src/content/docs/skills.md`, `site/src/content/docs/index.mdx`, `site/src/content/docs/flows/research.md`, `site/src/content/docs/flows/close.md`, `site/src/content/docs/getting-started.md`, `llms.txt`, `CHANGELOG.md`, `CONTRIBUTING.md`, `_support/docs/architecture/instruction-resources.md`

**Execution:** `hive-write-docs`

**Changes:** Página nueva con los encabezados de las páginas de flujo existentes; describe el comportamiento y enlaza la skill, sin copiar sus reglas. Entrada en la barra lateral y en las listas. Las páginas de research y de close dicen que el ticket se crea siguiendo la skill. `CONTRIBUTING.md` solo se toca si su lista de páginas del sitio debe nombrar la nueva, y `README.md` solo si afirma algo que este cambio altera; la tarea comprueba ambos y lo dice.

**Verification:** `pnpm --dir site check`, `pnpm --dir site lint` y `pnpm --dir site build` terminan sin errores. `rg -n "create-issue" llms.txt CHANGELOG.md _support/docs/architecture/instruction-resources.md` encuentra una entrada en cada uno. La vista previa del sitio la recorre el hilo principal en la verificación final.

### T7 — Refresco de la instalación y primer uso real

- [ ] La skill está instalada en cada host registrado y creó en `JhonHawk/tricell-hive` un ticket padre con dos hijos tras una aprobación del lote

**Closes:** AC12

**Depends on:** T1, T2, T3, T4, T5, T6, y el cambio integrado en `development`

**Locations:** la instalación local de Hive; el tracker de `JhonHawk/tricell-hive`

**Execution:** main thread

**Changes:** Tras la integración, nombra los commits aún no desplegados y refresca la instalación como fija el `AGENTS.md` del repositorio (`hive update --source <checkout> --out FILE` y `hive apply --plan FILE`), avisando de que las sesiones abiertas necesitan reiniciarse. Lee, en cada host registrado, los archivos de la skill en su directorio de skills y la guía global en su archivo de instrucciones globales, y comprueba lo que describen AC1, AC3, AC5, AC7 a AC10 y AC13 a AC22. Después este mismo hilo lee el `create-issue/SKILL.md` instalado y lo sigue para proponer un ticket padre («skills por entregable») con dos hijos (`create-prd`, `create-design`); el usuario aprueba el lote; relee los tres. Antes de escribir registra el destino, el cambio y cómo se revierte: cerrar los tres como no planeados. Si el host no permite leer la skill instalada desde esta sesión, la creación de los tickets queda como trabajo pendiente, `Pending: create-issue`, para una sesión nueva, y el cierre lo anota.

**Verification:** Cada archivo existe instalado en cada host registrado con el contenido que su criterio describe. `gh issue view` de cada uno de los tres tickets muestra lo aprobado, y `gh api repos/JhonHawk/tricell-hive/issues/<padre>/sub_issues` lista los dos hijos.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Tests of each task | When the task closes, locally | The task's command passes | Covered by the implementation |
| Full suite | Final candidate, locally | `go vet ./...`, `go test ./...` and `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Covered by the implementation |
| Instruction links | After the commit, before the push | `hive update --source <checkout> --dry-run` reports no broken link | Covered by the implementation |
| Site | Final candidate, locally | `pnpm --dir site check`, `lint` and `build`; the new page read in the local preview for navigation, search, mobile width and both themes | Covered by the implementation |
| Places checked | Final candidate | The list of every place that states or depends on ticket behavior, and the size of `global.md` before and after, for the pull request | Covered by the implementation |
| Code review | Final candidate, before the push | Claude Code `/code-review` on the candidate's diff | The review answer (D26 A) |
| First real use | After integration and the local refresh | The three tickets read back | The effects answer (D27 A), and the batch approval at that moment |

## Checkpoint

## Review status and progress

2026-11-16 — Plan escrito y revisado; las tareas recogen los nueve hallazgos de la revisión del plan.
