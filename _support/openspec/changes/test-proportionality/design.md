# Diseño

## Contexto verificado (2026-09-26)

- **TDD por defecto:** `content/skills/flow-build/SKILL.md:40` es la única regla que exige TDD, y lo hace "by default for new or changed automatically testable behavior". Su única exención es "trivial presentation or passive documentation changes".
- **Suite completa:** `SKILL.md:54` dice "run the existing full suite locally when no downstream coverage exists", sin límite de repeticiones.
- **Iteración ya acotada:** `verification.md`, fila "Build iteration", dice "Run targeted tests and routine checks". La tabla de superficies dice que las evidencias son "candidates, not a checklist to run in full for every change". Ya cubre la iteración; el conflicto está solo en `SKILL.md:54`.
- **Enfoque de test en planes:** `flow-plan/SKILL.md:34` pide elegir "the verification approach" por resultado, pero no un enfoque de test. `references/plan-format.md:70-80` define los campos de tarea sin él.
- **`test-engineer`:** `test-engineer.md:12` pide "Avoid tests that merely mirror implementation", sin un criterio operativo.
- **Ajuste `TDD`:** `global.md:93` lista `TDD` como ajuste opcional. Ningún archivo de `content/`, `tooling/` (fuera de `tooling/legacy`) ni `_support/docs` le da semántica, y ningún `AGENTS.md` bajo `~/Development/projects` lo declara (búsqueda del 2026-09-26).
- **`review-code`:** `review-code.md:12` ya revisa que los tests nuevos o cambiados puedan fallar.
- **Hive legacy** (`/path/to/reference-volume/dev-resources/tricell-hive-master`, `global/rules-situational/testing.md`): tenía la excepción para cambios triviales, `Test approach: tdd | characterization | not-applicable` y alcance de ejecución (subconjunto afectado al iterar, suite completa en el límite de integración, no repetir suites en verde "for confidence").
- **Proyectos de referencia:** optional reference project corre solo el archivo afectado en la fase de aplicación y deja la suite completa para verificar. optional reference project aplica "YAGNI applies to tests too". optional reference project exige la suite completa siempre, que es el patrón que este cambio evita.

## Enfoque

Separar dos cosas: el **check de verificación**, que siempre se exige porque es el oráculo del agente, y el **test retenido**, que se agrega solo cuando protege un comportamiento cuya falla importa. La regla se formula como una lista positiva de disparadores y otra de cambios mecánicos, con desempate explícito, porque los modelos de ejecución aplican literalmente una exención ambigua.

Se descartó D1-B (mantener el valor por defecto y ampliar exenciones) porque cualquier cambio fuera de la lista sigue empujando a TDD. D1-C (TDD solo si el repo lo pide) se descartó porque pierde el test-first en lógica riesgosa.

La regla tiene tres ramas: disparadores (TDD), mecánicos (sin test nuevo) y un residual para el resto de cambios de comportamiento, como un valor por defecto, un mapeo sin ramas o un endpoint que solo delega (check directo o extender los tests existentes del área). El desempate se ancla al comportamiento, no a la naturaleza del código, para que renombrar una función con ramas no dispare TDD (revisión H1 y H2).

Para el enfoque por tarea se usan tres valores en lugar de los cuatro de la investigación: `check` cubre todo lo que queda fuera del criterio de TDD, incluidos los cambios mecánicos, el residual y la lectura o revisión de un cambio de documentación, porque toda tarea ya debe nombrar su verificación. El legacy usaba `not-applicable`; `check` evita que se lea como "no se verifica".

## Redacción distribuida (inglés)

### `flow-build/SKILL.md:40` (reemplaza el párrafo)

> Use red-green-refactor TDD for new or changed behavior whose failure would matter, when the repository's test infrastructure can exercise it: logic with branches, loops, parsing, or calculation; a public or persisted contract; permissions, security, money, or data integrity; or a fixed defect that could recur. Write the behavioral test, observe it fail for the intended reason, implement the change, observe it pass, then refactor while preserving passing checks. A missing test for such behavior is a reason to write one, not to skip TDD. A mechanical change adds no new test: a rename, a move, user-facing text, configuration, formatting, a presentation or documentation edit, or wiring that the compiler, type checker, or existing tests already exercise; verify it with the checks that already cover it. When a change alters behavior of a kind listed for TDD, TDD applies even if the change also looks mechanical; a change that preserves behavior follows the refactor rule. For behavior-preserving refactors, use existing tests or characterization. Verify any other behavior change with a direct check, or by extending the tests that already cover that area; it needs no failing test first. A plan's declared test approach for a task governs that task unless the work turns out to alter behavior listed for TDD; then use TDD and record the reclassification. If no usable test infrastructure exists, choose an available validator or direct check and report the limitation; do not introduce a framework without an authorized need.

### `flow-build/SKILL.md:42` (agrega dos frases)

Después de "do not weaken a valid check to make the task appear complete.":

> Do not edit or remove a test that the user or the plan supplied as acceptance to make it pass; when it is wrong or obsolete, report why and propose the correction. Add no test for a case the change cannot produce.

Y la frase final pasa a ser:

> Remove obsolete code, configuration, or temporary residue created by the change when that cleanup is within scope, and a test the change makes obsolete after comparing its scenarios under the verification reference.

### `flow-build/SKILL.md:54` (reemplaza la segunda frase)

"Close regression coverage through the project's actual CI gate or run the existing full suite locally when no downstream coverage exists." pasa a:

> While iterating, run the targeted tests for the changed code. Close regression coverage through the project's actual CI gate; when none will run it, run the full suite once on the final candidate, and rerun it only when a later edit could affect what it covers. Do not rerun a passing suite for confidence. When the verification the change needs is much larger than the change itself, say so in the progress update with the reason.

### `flow-plan/SKILL.md:34` (agrega una frase al final del párrafo)

> Give each task that changes behavior a test approach under the build skill's TDD criterion: `tdd`, `characterization` for behavior-preserving refactors, or `check` for a change outside the TDD criterion, naming the check that verifies it.

### `flow-plan/references/plan-format.md` (plantilla de tarea)

Después de `**Execution:**`:

```markdown
**Test approach:** <tdd, characterization, or check naming the check; omit when the task changes no behavior>.
```

### `test-engineer.md:12` (punto 2)

> 2. Test meaningful boundaries, negative cases, asynchronous ordering, and transactions where they affect the change. Before writing each test, name the realistic change to the code under test that would make it fail; do not write a test that no such change can fail or that only restates the current structure.

### `global.md:93`

"Optional: `Environments`, `TDD`, `Review`, and `Hive guidance`." pasa a "Optional: `Environments`, `Review`, and `Hive guidance`.". `globalGuidanceBudget` baja al tamaño medido.

## Riesgos

- **Clasificación:** un modelo podría clasificar como "wiring" un cambio que sí altera lógica. El desempate anclado al comportamiento ("TDD applies even if the change also looks mechanical") y la mención explícita de "a fixed defect that could recur" reducen el riesgo; la medición queda pendiente por la pausa de pilotos.
- **Reporte de desproporción:** la frase sobre la verificación mucho mayor que el cambio pide reportar, no recortar, para que un modelo débil no quite tests requeridos.
