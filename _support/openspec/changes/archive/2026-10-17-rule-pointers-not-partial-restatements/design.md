# Diseño

## Contexto verificado

**Sesiones (auditoría del 2026-10-17, Engram #15674 y #15675):**

| Host | Modelo | Finales de respuesta con decisión |
| --- | --- | --- |
| Claude Code, fleetsystem `02891730` | claude-opus-5-5 | 6 tarjetas a mitad del flujo. 3 finales solo en texto (06:20Z, 06:42Z, 07:09Z); los dos últimos, después de leer `questions.md`, `flow-close` y `reporting.md` |
| Claude Code, fleetsystem `8168b9a2` | claude-opus-5-5 | 1 final solo en texto (06:54Z); nunca leyó `questions.md` |
| Grok, 5 sesiones | grok-4.7-build-fast | 7 de 7 con línea de espera y tarjeta |
| OpenCode, 5 sesiones | deepseek-v4.1-flash | 2 de 2 decisiones con tarjeta. Su razonamiento citó `global.md:20` y dudó si la línea de espera reemplaza a la tarjeta |
| Codex, 3 sesiones | gpt-6.1-sol | Sin finales con decisión. En modo por defecto el texto es lo correcto (`default_mode_request_user_input = false`; `agent-delivery.md:208`) |

**Inventario de la regla de preguntas fuera de `questions.md`** (base `c6c1ee16`):

| Sitio | Texto actual | Problema |
| --- | --- | --- |
| `global.md:20` | "use a waiting native card when available and complete text options otherwise, and end with one pause line" | Orden ambiguo |
| `flow-close/SKILL.md:34` | "ask the close question under [questions], ending with the pause line" | Nombra solo la línea de espera |
| `flow-close/references/reporting.md:65` | "An open choice is not a Pending item: it goes in the pause line." | Igual |
| `reporting.md:110` | "name it through the question procedure and end with its pause line" | Igual |
| `flow-plan/SKILL.md:18` | "ask it under the global choice rules, end with the global pause line"; "group the others into rounds under the global choice rules" | Remite al resumen global y nombra solo la línea de espera |
| `flow-plan/SKILL.md:75` | "End with the build route question under the global choice rules" | Remite al resumen global |
| `flow-research/SKILL.md:40` | "end that same message with the route question under the global choice rules" | Igual |
| `flow-plan/references/delivery-decisions.md:16` | "the one-line summary the global choice rules describe" | Igual |
| `flow-build/references/verification.md:67` | "with the global pause line … otherwise also ask through a native question card" | Excepción local legítima (con hijos corriendo, sin tarjeta); remite a la línea de espera global |
| `verification.md:69` | "ask again right away in the same way, with the pause line and the in-progress item" | Nombra solo la línea de espera |
| `harness-audit/SKILL.md:58` | "through the host's native question tool when it has one, otherwise in text, and wait" | Omite la línea de espera |

`verification.md:79` ("ends with the pause line, with a native question card … only when no children are running") nombra los dos elementos y su condición; queda como excepción local completa.

**Evidencia sobre el principio:**
- **Guía de prompts de OpenAI para GPT-5.6:** "Keep the policy in one place and state each rule once"; las reglas en conflicto desestabilizan más que el detalle que falta. La fuente es un resumen web de la página.
- **Documentación de memoria de Claude Code** (consultada el 2026-10-17): ante una contradicción, Claude puede elegir cualquiera de las dos; para forzar algo, un hook.
- **Buenas prácticas de Skills de Anthropic:** divulgación progresiva y términos consistentes; una regla que se olvida se hace más visible, no se copia.
- **Estudios:** IHEval (NAACL 2025), los conflictos bajan el rendimiento; ManyIFEval (2025), más instrucciones lo bajan.
- **Historial de Hive:**
  - Las tres contradicciones de la poda del 2026-10-15 nacieron de repetir una regla con otro alcance (`archive/2026-10-15-guidance-pruning/design.md:5`).
  - El umbral de delegación que no se leía se subió a `global.md` el 2026-10-16 (`b8d2b785`).
  - Las lecturas parciales de Grok se resolvieron con un hook (PR #52).

- **Contraevidencia:** el catálogo de `harness-audit` cita arXiv 2605.10039, según el cual las contradicciones entre archivos no tuvieron un efecto detectable en la adherencia (`harness-audit-rules.md:24`). Aquí el fallo observado no es una contradicción entre archivos sino que el modelo sigue la versión corta, pero ese estudio limita cuánto se puede afirmar.

Ningún estudio compara directamente "un hogar más punteros" con "copias completas" en arneses de agentes. El principio es una inferencia a partir de esa evidencia y lo adopta la decisión del usuario (R4-A).

## Diseño elegido

### T1: AGENTS.md

Reemplaza la viñeta actual "- Give each rule one canonical home. Avoid duplicate routers, fallback chains, generated copies, hooks, or adapters without a measured need." por:

```
- Give each rule one canonical home, which alone states its full steps, their order, and its exceptions. Where the rule applies elsewhere, add only a pointer that names the rule and when to read it, such as "before asking the user, read and follow [questions](…)"; never restate some of its steps there, because a model follows the shorter version and skips the rest. Restate a rule in `content/guidance/global.md` only when it must act without its home being read, and then state its minimal invariant whole. An exception that applies at only one point of use stays there and names every element of the rule it changes. Correct an existing copy when a change edits that copy or the rule it restates. Avoid duplicate routers, fallback chains, generated copies, hooks, or adapters without a measured need.
```

### T2: `global.md:20`

Reemplaza la oración entera "Group independent questions, use a waiting native card when available and complete text options otherwise, and end with one pause line." por:

```
Group independent questions and end the message with one pause line; then ask through a waiting native card when the host has one, and otherwise put complete text options in the message before that line.
```

El resto de la viñeta no cambia. "Before that line" sigue a `questions.md:56`, donde la línea de espera es la última del mensaje (re-revisión N2).

### T3: punteros en las skills

| Sitio | Texto nuevo |
| --- | --- |
| `flow-close/SKILL.md:34` | "ask the close question by reading and following [questions](../flow-plan/references/questions.md)." |
| `reporting.md:65` | "An open choice is not a Pending item: ask it by following [questions](../../flow-plan/references/questions.md)." |
| `reporting.md:110` | "ask it by following [questions](../../flow-plan/references/questions.md);" en lugar de "name it through the question procedure and end with its pause line;" |
| `flow-plan/SKILL.md:18` | "ask it by following [questions](references/questions.md), and stop the work" en lugar de "ask it under the global choice rules, end with the global pause line, and stop the work"; "group the others into rounds as [questions](references/questions.md) sets" |
| `flow-plan/SKILL.md:18` (tercer sitio) | "say so as [questions](references/questions.md) requires for an answer" en lugar de "say so as the global rule for answers requires" |
| `flow-plan/SKILL.md:75` | "End with the build route question, asked by following [questions](references/questions.md)," |
| `flow-research/SKILL.md:40` | "end that same message with the route question, asked by following [questions](../flow-plan/references/questions.md):" |
| `delivery-decisions.md:16` | "the one-line summary [questions](questions.md) describes" |
| `verification.md:67` | "with the pause line that [questions](../../flow-plan/references/questions.md) defines" en lugar de "with the global pause line"; el resto de la excepción no cambia |
| `verification.md:69` | "ask again right away in the same way." Se quita solo ", with the pause line and the in-progress item". "In the same way" remite a la pregunta de la pausa de validación en curso, sea la de `:67` o la de `:79`; las dos nombran la línea de espera, la tarjeta y la condición de los hijos (B1). Se pierde la exigencia de abrir con el elemento en curso en las correcciones que nacen de `:79`. Se acepta: `:67` sí abre con él, y `:79` es la validación final, sin trabajo en curso que anunciar |
| `harness-audit/SKILL.md:58` | "Ask those decisions, including each HA-IF-08 choice of which version to keep and each conflict with a global layer, by following [questions](../flow-plan/references/questions.md), and wait for the answers;" |

Antes de editar, cada ruta relativa se comprueba desde el archivo que la contiene. `TestRepositoryCatalogueInstructionReferences` (`tooling/management/references.go`) valida los enlaces relativos de las skills.

### T4: HA-IF-08

En `instruction-files.md:28`, la primera oración "Flag duplicates and paraphrases across files and levels; propose one owner and `delete` or `re-anchor` the rest." pasa a:

```
Flag duplicates and paraphrases across files and levels, including a partial restatement that repeats only some of a rule's steps where the rule applies. Do not choose the version to keep: ask the caller which one is the rule to follow, one decision per rule, quoting each version with its location; then propose `delete` for the rest, subject to HA-IF-13, or `re-anchor` each as a pointer that names the rule and when to read it. A rule's minimal invariant restated whole in an always-loaded file, because it must act before its home is read, is not a duplicate.
```

El resto de la regla, sobre contradicciones y capas globales, no cambia. Esa parte ya pregunta al usuario. La pregunta nueva aplica lo mismo a los duplicados y a las copias parciales; lo pidió el usuario el 2026-10-17: "nadamas que harness-audit debe confirmar con el usuario qué regla es la que debe seguir".

La fila de `harness-audit-rules.md:24` cambia así:
- **Regla:** "Each rule has one canonical home; flag duplicates and partial restatements across files and levels, ask the caller which version to keep, re-anchor the rest as pointers, and flag contradictory pairs rather than keeping both."
- **Fuentes:** se agrega "OpenAI GPT-5.6 prompt guidance; Hive sessions 2026-10-17; user decision D1-B".
- **Evidencia:** `LEGACY, DOC, OBS, DEC`.
- **Fecha:** 2026-10-17.
- Los límites y el estado no cambian.

## Lugares revisados

- `global.md:18`, sobre glosar IDs en la tarjeta: no es el procedimiento; sin cambio.
- `global.md:23`, `reporting.md:93` y las líneas 10-12 de las skills que ya son punteros con su condición de lectura: sin cambio.
- `site/src/content/docs/flows/research.md:47` y `close.md:63`: describen la tarjeta y la línea de resumen, y son coherentes; sin cambio.
- `hive-review-harness`: no cita HA-IF-08 por texto; la prueba de paridad compara solo IDs.
- `tests/pilot/`: los detectores `question_after_detail` (S1, texto antes de cada pregunta nativa) y `close_question_after_report` (`tests/fixtures/regression/README.md:7,10`; `regression.go:97`) son coherentes con el orden de T2; sin cambio.
- `_support/docs/architecture/agent-delivery.md:189-237` (tarjetas por host): coherente; sin cambio.

## Notas para los commits

Cada commit nombra:
- Claude Code fleetsystem `02891730` y `8168b9a2` (2026-10-17): preguntas al final de la respuesta sin tarjeta.
- Carvento (auditoría del 2026-10-16, Engram #15642).
- La decisión R4-A y, en T4, D1-B.

T2 declara el cambio de tamaño de `global.md`.
