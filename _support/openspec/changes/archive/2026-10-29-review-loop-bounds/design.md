# Diseño

## Verified context

Hechos inspeccionados el 2026-10-29 sobre `development` en f3d15886 (copia desplegada idéntica):

- `content/skills/flow-verify/references/review-and-gates.md:25`: un revisor que no verifica manda sus P0/P1 a `hive-refute-claim`; «Native reviews that already verify each finding, such as Claude Code's `/code-review` at medium effort or above, need no second pass».
- Las 94 bifurcaciones de `/code-review` de las sesiones medidas (Claude Code 2.1.312 instalado) cargaron el mismo prompt: `high effort → 8 inline angles → dedup (no verify) → ≤10 findings`, «Target at least 5 findings», como máximo 10. Ejemplo: la primera línea de `~/.claude/projects/-Users-jmartinez-Development-projects-tricell-tricell-hive/7f3329ce-92e3-4d1a-aa47-68048046b4f9/subagents/agent-acd20208437e0dffa.jsonl`. Solo el esfuerzo `low` difiere (≤4).
- Líneas 29 (severidad solo para mecanismos con niveles), 35 (segunda revisión) y 36 (ediciones posteriores), sin cambios de fondo desde 2fe26075 (2026-10-28).
- `ui-gate.md:23` define la identidad de candidato (commit `HEAD` o hash de árbol sin tocar índice ni working tree).
- Dependientes que repiten o apuntan a estas reglas: `review-and-gates.md:44-45` (seguridad: «same limits»), `content/skills/flow-research/references/pr-review.md:15`, `content/skills/flow-build/references/verification.md:40`, `site/src/content/docs/flows/verify.md:30` y `site/src/content/docs/flows/build.md:47`, que repite la regla vieja de ediciones posteriores («reviews what changed since the reviewed commit when an earlier review answer covers it»). No hay requisitos en `_support/openspec/specs/` ni detectores en `tests/pilot/` que fijen esta redacción; `tests/content/reading_footprint_test.go:57,96` lista el archivo por ruta.
- Tamaño actual del archivo: 15 216 bytes de 16 000.

Medición que motiva el cambio (sesiones 33c2742a, 5235e23a, 6495b668 y 7f3329ce): 19 rondas en 11 tickets, 171 hallazgos sin bugs de ejecución; segundas rondas sin P0/P1 registrado en #262, #272 y #277 (dos sesiones); tercera ronda en #262 por la viñeta «Later edits».

## Proposed design

Se reescriben cuatro piezas de «Dedicated code review», sin mover la sección:

1. **Verificación (AC1, F2).** El párrafo inicial define una revisión que verifica como aquella cuya salida declara, hallazgo por hallazgo, que lo reprodujo o lo releyó contra la fuente; el nombre de un comando o su nivel de esfuerzo no lo establece. Se quita el ejemplo de `/code-review`. Así, `/code-review` cae en la regla general y sus P0/P1 van al refutador.
2. **Severidad (AC2, F3).** La viñeta «Severity» mapea cada hallazgo a P0–P3: traduce los niveles propios del mecanismo y asigna uno cuando no los hay; registra el nivel de cada hallazgo en el reporte y nombra allí el arreglo P0/P1 que inicia la segunda revisión.
3. **Segunda revisión (AC3, AC5, D3-A).** Sigue siendo como máximo una, limitada al diff de los arreglos y solo cuando un arreglo P0/P1 cambió lógica o un contrato. El diff se mide contra el commit revisado o, si el candidato revisado no tenía commit (incluido `hold`), contra su identidad de candidato, que se registra al iniciar la revisión. De esa ronda solo se actúa sobre sus P0/P1, refutados o verificados como en el punto 1. Sus P2 sin arreglar siguen la regla compartida de hallazgos incidentales, como ya hace la viñeta P2 con los que no arregla, y sus P3 se reportan. La cláusula posterior queda limitada a los P0/P1 de esa ronda: se arreglan sin otra revisión si el arreglo es local, dentro del alcance y sin contrato; los demás P0/P1 van al usuario. Sin esa limitación, «would be fixed under these rules» volvería a arreglar un P2 introducido, contra D3-A.
4. **Ediciones posteriores (AC4, D4-A).** Nunca inician una revisión completa. Mientras la segunda revisión no ha corrido, una edición cubierta por la respuesta de revisión entra en su diff (y esa revisión del delta es la única segunda revisión); sin respuesta que la cubra, se reporta y se pregunta como hoy. Después de la segunda revisión no hay otra, salvo la revisión de las correcciones de la validación que fija «Dedicated review timing» (línea 51, sin editar): una edición local, dentro del alcance y sin contrato se reporta como no revisada, y el cuerpo del PR la registra así para que la regla de promoción (`delivery-decisions.md:19`) sepa qué commits no se revisaron; cualquier otra se pregunta al usuario.

La sección de seguridad (líneas 44-45) hereda los límites por referencia («under the same limits») y no se edita; se comprueba que siga siendo coherente. La revisión de las correcciones de la validación humana (línea 51) queda como está, por estar fuera de alcance.

Presupuesto de bytes: las cuatro piezas suman unos +500 bytes. Si el archivo pasa de 16 000, se compacta la redacción de esta misma sección antes que dividir el archivo, porque dividirlo afecta `reading_footprint_test.go` e `instruction-resources.md`.

Alternativa descartada: nombrar `/code-review` como no verificador. Ata la regla a un host y a una versión, en contra de AGENTS.md.
