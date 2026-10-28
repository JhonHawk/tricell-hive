# Diseño — gh-259-option-scope-size

## Contexto verificado (development `0f64489b`, 2026-10-28)

- `content/skills/hive-core/references/questions.md:45-47` es el hogar único de la condición de tamaño: disparador de otro ticket, un componente, un lenguaje, una dependencia o una diferencia de 10×; pide "the number of tasks and a time range" con base y "size not estimated: not investigated". No dice qué mide el tiempo. El archivo pesa 15,992 B (límite 16,000, `tests/content/skill_size_test.go`).
- Ninguna regla pide ofrecer la opción mínima. `content/guidance/global.md:63` pide "the smallest sufficient investigation, plan, implementation, and verification", pero nada lo lleva a las opciones de una pregunta.
- `content/skills/flow-plan/references/plan-handoff.md` define la pregunta de ruta de build (`flow-build` o detenerse) sin tamaño. Detenerse no tiene tamaño, así que el disparador de 10× no la cubre.
- Consumidores de la condición de tamaño: `content/skills/flow-close/SKILL.md:47` ("which the size condition of [questions] sets"), `site/src/content/docs/flows/plan.md:70` y `site/src/content/docs/flows/close.md:50`. En `tests/` y `tests/pilot/` ningún detector usa "ten times", "estimated size" ni "time range".
- La regla de tamaño viene de #158 (`12442782`, 2026-10-25).

## Enfoque

1. **Referencia nueva `content/skills/hive-core/references/option-scope.md`** ("Option scope and size"). Recibe entera la condición de tamaño de `questions.md` y la regla nueva de opción mínima (D2-A). Su texto, en inglés como toda la guía distribuida:

   ```markdown
   # Option scope and size

   ## When to read

   Read before formulating a question in which two or more options start work on the same request and build different amounts, such as scope or design options that add or remove tasks, or the fixes a research route offers; before formulating a question with an option that starts work on another ticket, adds a component, a language or a dependency, or is, as far as you can tell without estimating, at least ten times the size of another option; and when another rule asks for an option's estimated size, such as the build route of [plan handoff](../../flow-plan/references/plan-handoff.md) or the close question of [flow-close](../../flow-close/SKILL.md). [Questions](questions.md) sets the rest of the question. Reading this reference does not invoke planning or grant any effects.

   ## Offer the smallest sufficient option

   **Condition:** two or more of a question's options start work on the same request and build different amounts. Options that run the same work in different ways build the same amount, such as `flow-build` or direct execution in the build route of [plan handoff](../../flow-plan/references/plan-handoff.md), or continuing now or in a new session in the close question.

   **Actions:** include among them the smallest option that solves the problem the request, its ticket or an approved acceptance criterion states, even when you do not recommend it. When the Recommended option builds more, its reason names what the smaller option leaves unsolved, in the terms of that request, ticket or criterion; robustness or completeness for a case none of them states is not such a reason.

   ## Give the estimated size

   **Condition:** an option of the question starts work on another ticket, adds a component, a language or a dependency, or is, as far as you can tell without estimating, at least ten times the size of another option of the same question; or another rule asks for the estimated size.

   **Actions:** give every option of that question that starts work, not one such as stop, its estimated size: the number of tasks with its basis, such as the ticket's criteria or the files you read, and a time range. The time range is how long you, the agent, will take from start to finish, children, checks and reviews included, not a person's effort. Base it only on evidence already at hand, such as the durations of tasks completed in this session or the commit times of comparable work you already read; never read, search or delegate only to estimate it. Without such evidence, write "time not estimated: no basis at hand" in place of the range. When you did not read enough to count the tasks, write "size not estimated: not investigated" for that option. Never give a number without a basis.
   ```

2. **`questions.md`** cambia las líneas 45-47 por un puntero que conserva la condición y manda a leer:

   ```markdown
   **Condition:** two or more of a question's options start work on the same request and build different amounts, or one of them starts work on another ticket, adds a component, a language or a dependency, or is, as far as you can tell without estimating, at least ten times the size of another option.

   **Actions:** read [option scope and size](option-scope.md), which sets the smallest option to offer and the estimated size to give.
   ```

3. **`plan-handoff.md`**, solo en la opción `flow-build` de la pregunta de ruta (D3-A, excepción de un solo punto de uso): "It also gives the plan's estimated size, as [option scope and size](../../hive-core/references/option-scope.md#give-the-estimated-size) sets." La condición de la referencia ya cubre que otra regla pida el tamaño. No repite las partes del tamaño (tareas, rango), para que la referencia siga siendo el único hogar.
4. **`flow-close/SKILL.md:47`**: "which the size condition of [questions](../hive-core/references/questions.md) sets" pasa a "which [option scope and size](../hive-core/references/option-scope.md#give-the-estimated-size) sets". "that size is its remaining tasks with their time range" pasa a "that size covers its remaining tasks". Así no queda una copia parcial que pida un rango sin base. La regla de dar el tamaño una sola vez por pregunta se queda en flow-close.
5. **Índices:**
   - `hive-core/SKILL.md` agrega la viñeta "Before formulating a question in which two or more options start work on the same request and build different amounts, or one with an option that starts work on another ticket, adds a component, a language or a dependency, or is, as far as you can tell without estimating, at least ten times the size of another, read [option scope and size](references/option-scope.md)." Repite la condición del puntero de `questions.md` y no agrega una tercera redacción. Su `description` suma "option scope and size".
   - `_support/docs/architecture/instruction-resources.md` la agrega a los hogares canónicos, y en la prosa del ciclo plan/build (líneas 86-92) "questions/reporting" pasa a "questions/option scope/reporting". La tabla fechada en `2d89940e` se queda como está, porque ya dice "rerun the test".
   - `tests/content/reading_footprint_test.go` la suma al ciclo plan/build, que la lee por la ruta de build, y al mapa `moved`, porque no existe en la base A. El escenario `question` (una decisión suelta) no la suma: solo se lee cuando la pregunta cumple su condición.
6. **Sitio y changelog:**
   - `site/.../flows/plan.md:70` describe la opción mínima y el tiempo del agente y enlaza la referencia.
   - `plan.md:67` agrega el tamaño a la pregunta de ruta.
   - `site/.../flows/research.md:50` dice que, entre las formas de actuar, va la más pequeña que resuelve el problema, y que una recomendada mayor dice qué deja sin resolver.
   - `close.md:50` aclara que el tamaño es tiempo del agente.
   - `CHANGELOG.md` suma una entrada en `### Changed` de `[Unreleased]`.

## Alternativas descartadas

- Bajar el disparador a 3× (D3-B): no resuelve la ruta de build, porque detenerse no tiene tamaño.
- Quitar el rango de tiempo (D4-B): el usuario prefirió el tiempo del agente con evidencia a la mano (D4).
- Recortar `questions.md` para que quepa el texto nuevo: editaría reglas ajenas a este cambio.
