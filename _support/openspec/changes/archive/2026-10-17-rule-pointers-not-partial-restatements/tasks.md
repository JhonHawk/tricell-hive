# Tareas

Base del cambio: `c6c1ee16` (`feat/delegation-threshold-global`), confirmada al empezar el build el 2026-10-17.

## T1: el principio en AGENTS.md

- [x] La viñeta del hogar canónico dice qué va en el hogar, qué va en el punto de uso, cuándo repetir en `global.md` y cómo escribir una excepción local.

**Closes:** AC1.

**Depends on:** ninguna.

**Locations:** `AGENTS.md`, sección "Guidance structure", viñeta "Give each rule one canonical home".

**Execution:** hilo principal. Es una viñeta, y un brief para delegarla sería más largo que el cambio.

**Test approach:** check, `rg`.

**Changes:** el texto de [design.md § T1](design.md#t1-agentsmd).

**Verification:**
- `for p in 'which alone states its full steps, their order, and its exceptions' 'add only a pointer that names the rule and when to read it' 'never restate some of its steps' 'a model follows the shorter version' 'only when it must act without its home being read' 'state its minimal invariant whole' 'names every element of the rule it changes' 'Correct an existing copy when a change edits that copy or the rule it restates'; do echo "$p: $(rg -c -F "$p" AGENTS.md)"; done` → 1 en cada línea.
- Las coincidencias están en una sola línea, dentro de `## Guidance structure` (`rg -n '^## ' AGENTS.md`).

## T2: el orden en `global.md:20`

- [x] La regla de preguntas de `global.md` termina el mensaje con la línea de espera y después pregunta con la tarjeta, o pone las opciones en el texto.

**Closes:** AC2.

**Depends on:** ninguna.

**Locations:** `content/guidance/global.md:20`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg`.

**Changes:** el texto de [design.md § T2](design.md#t2-globalmd20).

**Verification:**
- `rg -n -F 'Group independent questions and end the message with one pause line; then ask through a waiting native card when the host has one, and otherwise put complete text options in the message before that line.' content/guidance/global.md` → la línea 20.
- `rg -c -F 'and end with one pause line' content/guidance/global.md` → 0.

## T3: punteros en las skills

- [x] Los once sitios dejan de resumir la regla en parte: diez remiten a `questions.md` y `verification.md:69` remite a la pregunta de la misma pausa.

**Closes:** AC3.

**Depends on:** ninguna.

**Locations:**
- `content/skills/flow-close/SKILL.md:34`.
- `content/skills/flow-close/references/reporting.md:65,110`.
- `content/skills/flow-plan/SKILL.md:18,75`.
- `content/skills/flow-research/SKILL.md:40`.
- `content/skills/flow-plan/references/delivery-decisions.md:16`.
- `content/skills/flow-build/references/verification.md:67,69`.
- `content/skills/harness-audit/SKILL.md:58`.

**Execution:** hilo principal. Son once reemplazos cortos ya escritos en el diseño.

**Test approach:** check, `rg` y la prueba de enlaces `TestRepositoryCatalogueInstructionReferences`.

**Changes:** la tabla de [design.md § T3](design.md#t3-punteros-en-las-skills).

**Verification:**
- `rg -n -o 'global choice rules|global pause line|ending with the pause line|end with its pause line|goes in the pause line|with the pause line and the in-progress item|native question tool when it has one, otherwise in text' content` → sin coincidencias.
- Una comprobación por sitio, con `rg -n -F '<frase>' <archivo>` → la línea esperada:
  - `flow-close/SKILL.md`: 'ask the close question by reading and following [questions](../flow-plan/references/questions.md).' → 34.
  - `reporting.md`: 'An open choice is not a Pending item: ask it by following [questions](../../flow-plan/references/questions.md).' → 65; 'ask it by following [questions](../../flow-plan/references/questions.md); do not put that open choice' → 110.
  - `flow-plan/SKILL.md`: 'ask it by following [questions](references/questions.md), and stop the work' → 18; 'group the others into rounds as [questions](references/questions.md) sets' → 18; 'say so as [questions](references/questions.md) requires for an answer' → 18; 'End with the build route question, asked by following [questions](references/questions.md),' → 75.
  - `flow-research/SKILL.md`: 'end that same message with the route question, asked by following [questions](../flow-plan/references/questions.md):' → 40.
  - `delivery-decisions.md`: 'the one-line summary [questions](questions.md) describes' → 16.
  - `verification.md`: 'with the pause line that [questions](../../flow-plan/references/questions.md) defines' → 67; 'and ask again right away in the same way.' → 69.
  - `harness-audit/SKILL.md`: 'by following [questions](../flow-plan/references/questions.md), and wait for the answers' → 58.
- `rg -c -F 'global rule for answers' content/skills/flow-plan/SKILL.md` → 0.
- `go test ./tooling/management/ -run TestRepositoryCatalogueInstructionReferences` pasa: valida los enlaces relativos de las skills.

## T4: HA-IF-08 marca copias parciales

- [x] HA-IF-08 marca las copias parciales, pregunta al usuario qué versión queda y propone reanclar las demás como punteros; el catálogo lo refleja.

**Closes:** AC4.

**Depends on:** T3, que reescribe `harness-audit/SKILL.md:58` con el texto que AC4 verifica.

**Locations:**
- `content/skills/harness-audit/references/instruction-files.md:28`.
- `content/skills/harness-audit/SKILL.md:58` (la misma línea que T3; T4 depende de T3 para no editarla dos veces).
- `_support/docs/harness-engineering/harness-audit-rules.md:24`.

**Execution:** hilo principal, por la misma razón que T1.

**Test approach:** check, `rg` y la prueba de paridad.

**Changes:** el texto de [design.md § T4](design.md#t4-ha-if-08).

**Verification:**
- `for p in 'including a partial restatement that repeats only some of a rule' 'Do not choose the version to keep: ask the caller which one is the rule to follow' 'one decision per rule, quoting each version with its location' 'propose `delete` for the rest, subject to HA-IF-13, or `re-anchor` each as a pointer that names the rule and when to read it' 'is not a duplicate'; do echo "$p: $(rg -c -F "$p" content/skills/harness-audit/references/instruction-files.md)"; done` → 1 en cada línea.
- `rg -c -F 'user decision D1-B' _support/docs/harness-engineering/harness-audit-rules.md` → 1.
- `rg -n -F 'each HA-IF-08 choice of which version to keep' content/skills/harness-audit/SKILL.md` → la línea 58.
- `rg -n -F 'ask the caller which version to keep, re-anchor the rest as pointers' _support/docs/harness-engineering/harness-audit-rules.md` → la fila de HA-IF-08, con fecha `2026-10-17` y estado `active`.
- `python3 -m unittest tests.skills.harness_audit_rules_test` pasa.

## Verificación y revisión

| Compuerta | Momento y entorno | Mecanismo y evidencia esperada | Autoridad |
| --- | --- | --- | --- |
| Suite local | Sobre el candidato final, en este checkout | `go vet ./...`, `go test ./...` y `python3 -m unittest discover -s tests/skills -p '*_test.py'` sin fallos | Incluida en la implementación |
| Tamaño de `global.md` | Candidato final | `wc -c` frente a 28 865 bytes en la base; la diferencia va en el commit de T2 | Incluida |
| Revisión del alcance | Candidato final | Releer `questions.md:25-64` con los diez punteros y confirmar que cada sitio sigue diciendo cuándo preguntar; la excepción de `verification.md:67,79` conserva su condición | Incluida |
| Revisión de código dedicada | No aplica | Cambios solo de texto; preferencia guardada para Hive | — |
| Despliegue local | Tras los commits | `hive update --out <plan>` + `hive apply --plan <plan>` | C1-A |
| Prueba en sesiones reales | Tras el despliegue | Tú, en sesiones nuevas de Claude Code | Tuya |

## Revisión y progreso

- 2026-10-17: borrador escrito tras R4-A, tu pedido de volverlo canon y D1-B. Después se añadió tu pedido de que `harness-audit` confirme contigo qué regla seguir.
- 2026-10-17: revisión del plan sobre `8bad7626e8af` con un `hive-review-plan` nativo (dominio: guía distribuida, reglas de redacción de Hive y `harness-audit`). Duró 154 s y usó ~128k tokens.
  - **Bloqueantes:**
    - B1 (`verification.md:69` sin enlace y atado a `:67`): aceptado. Se conserva "in the same way" y se quita solo la cola parcial; AC3 admite esa remisión interna.
    - B2 (la verificación por archivo ya pasaba en la base): aceptado. Ahora hay una frase por sitio con su línea esperada.
  - **No bloqueantes, aplicados como los propuso el revisor:**
    - El comando de enlaces pasa a `TestRepositoryCatalogueInstructionReferences`.
    - Se añade "before that line" en `global.md:20`.
    - Se añade "Correct an existing copy when a change touches it" en AGENTS.md.
    - Se añade "subject to HA-IF-13".
    - Se cita la contraevidencia de arXiv 2605.10039.
    - Se corrige lo que el plan decía de `tests/pilot/` y se añade el undécimo sitio (`flow-plan:18`, "global rule for answers").
  - B1 cambia AC3, así que va a una re-revisión, la única ronda.
- 2026-10-17: re-revisión sobre `5aa8f96af63f`, reanudando al mismo revisor. Duró 43 s y usó ~141k tokens.
  - **Bloqueante:** el reemplazo de T2 dejaba una coma y no coincidía con su verificación. Ahora se reemplaza la oración entera.
  - **No bloqueantes:**
    - "edits that copy or the rule it restates".
    - HA-IF-13 junto a `delete`.
    - Justificación de lo que se pierde en `:69`.
    - Se quitan los dos textos viejos que citaban `tests/content`.
    - T4 comprueba "user decision D1-B".
  - Todos se aplicaron como los propuso el revisor, sin otra ronda.
- **Cobertura:** el dominio quedó revisado en `8bad7626e8af` y re-revisado en `5aa8f96af63f`. Los cambios posteriores aplican las propuestas del revisor.
- 2026-10-17, build: T1–T4 se implementaron en el hilo principal (claude-opus-5-5, observado). Cada una pasó a `[x]` con veredicto `met` de `hive-verify-task`, con modelo `opus` y esfuerzo `high` (configurado, no observado):
  - T1, AC1: 40 s.
  - T2, AC2: 42 s.
  - T3, AC3: 133 s.
  - T4, AC4: 58 s, después de T3.
- `harness-audit/SKILL.md:58` entró en el commit de T3 con el texto que T4 necesita. Su verificador lo revisó como resultado conjunto.
- **Compuertas del candidato final:**
  - `go vet ./...` sin hallazgos, `go test ./...` con salida 0 y tests de Python OK, incluida la paridad de `harness-audit`.
  - `TestRepositoryCatalogueInstructionReferences` pasa.
  - `global.md` pasó de 28 865 a 28 933 bytes (+68).
  - Revisión del alcance: el verificador de T3 releyó `questions.md` y los once sitios y no encontró copias parciales restantes.
- **Commits en `feat/delegation-threshold-global`:** `a2437302` (T1), `541475e3` (T2), `13163f80` (T3) y `aa17fbe1` (T4), sin push.
- **Despliegue:** `hive update --out` + `hive apply --plan` desde `aa17fbe1`, transacción `adfe8bf8…`. El texto nuevo está en las instrucciones globales de Claude (que también usa Grok), Codex, OpenCode, Pi y Cursor, y en las skills instaladas.
- **Estado:** desplegado en local para tus pruebas. Siguiente paso: tú lo pruebas en sesiones nuevas de Claude Code y decides el merge a `development`.
- **Límites:** no se verificaron la evidencia externa ni las sesiones de fleetsystem más allá de la auditoría. Las demás reglas con copias quedan para cuando un cambio las edite.
