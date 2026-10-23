# Diseño

## Contexto verificado

- **Carvento `1a3ab3ea`** (Claude Code, claude-opus-5-5): la sesión abrió a las 07:42:56Z, después del despliegue de `aa17fbe1` (`~/.claude/CLAUDE.md` del 01:41:44 local). Su transcripción contiene la frase nueva de `global.md:20` y no la vieja. A "revisa si hay ramas viejas sin fusionar en carvento-monorepo y dime qué hacemos con ellas" respondió con una tabla. Cerró con las dos salidas de la PR #306 en texto y "¿La reviso para que decidas?". No llamó a ninguna herramienta de preguntas, no leyó skills ni referencias, y no puso línea de espera.
- **Causa probable (inferencia):**
  - `global.md:20` y `questions.md:5` activan la lectura "before formulating a decision", y el modelo no clasificó la oferta como decisión.
  - `questions.md:11` ya dice "Never ask only for leave to proceed or to show something you could show", pero solo lo lee quien abrió el archivo.
- **Otros hosts:**
  - OpenCode (deepseek-v4.1-flash): sus 3 finales solo en texto eran ofertas.
  - Grok (grok-4.7-build-fast): usó la tarjeta en 7 de 7 decisiones; no tuvo ofertas al final.
  - Auditoría del 2026-10-17, Engram #15675.
- **Una oferta que la guía misma pide:** `reporting.md:41` dice "offer a `flow-report` page in one line". Con la regla nueva sería una pregunta suelta. Por eso pasa a ser una opción de la pregunta de cierre del mensaje, o una pregunta propia si el mensaje no tiene otra.

## Diseño elegido

### T1: `questions.md`

En "When to read" (línea 5), "Read before formulating a decision for the user, with or without a native card, including route, delivery, review, validation and close questions." pasa a:

```
Read before formulating a decision for the user, with or without a native card, including route, delivery, review, validation and close questions, and before ending a message with an offer or request that waits on the user's answer.
```

En la línea 11 se quita la última oración, "Never ask only for leave to proceed or to show something you could show.", porque el bloque nuevo la contiene (H4). Después del bloque de las líneas 9-11 se agrega:

```
**Condition:** a message would end with an offer or request that waits on the user's answer, in any form, such as "should I review it?", "if you want, I'll close it", or "let me know if you want the report".

**Actions:** when the action is already authorized, including the checks the global guidance says not to ask about separately, such as reading or showing a pull request's diff or running the affected tests, do it instead of offering it. When it is a choice that belongs to the user under this section's first condition, or a question another rule requires, such as the route, close or ticket question, ask it under this procedure, never as a closing sentence alone. Otherwise, such as further work the answer does not call for, leave the offer out.
```

### T2: `global.md:20`

"Before formulating a decision or requesting user-dependent review, validation or manual action, read [questions and pauses](skill:flow-plan/references/questions.md)." pasa a:

```
Before formulating a decision, ending a message with an offer or request that waits on the user's answer in any form, or requesting user-dependent review, validation or manual action, read [questions and pauses](skill:flow-plan/references/questions.md).
```

Sin ejemplo, porque un ejemplo solitario se lee como todo el alcance (AGENTS.md:21); los ejemplos viven en `questions.md` (H2).

Es el aviso con su condición de lectura, como pide la norma de AGENTS.md. El procedimiento queda en `questions.md`.

### T3: `reporting.md:41`

"offer a `flow-report` page in one line." pasa a:

```
offer a `flow-report` page as an option of the question the message ends with, such as a route or close question, or, when the message ends with none, ask it by following [questions](../../flow-plan/references/questions.md).
```

Así se conserva la oferta que pide la regla y deja de ser una frase suelta (H6). La herramienta Artifact de Claude Code también pide "offer the page in one line", pero es texto del host que Hive no controla.

## Lugares revisados

- **Reglas coherentes, sin cambio:**
  - `questions.md:11`: su última oración, "Never ask only for leave to proceed or to show something you could show", se quita y queda dentro de la primera salida del bloque nuevo (H4, H10).
  - `flow-plan/SKILL.md:75`: la pregunta de ruta de build es "not a request for leave to proceed". Coherente: es la segunda salida, porque un plan no autoriza implementar.
  - `flow-research/SKILL.md:40`: "A question the findings fully answer, with no work to run, ends without a route question".
  - `reporting.md` "Next step", "Recommendations" y "Reminders" (líneas 67-87).
- **`global.md:66`** ("do not ask separately to perform" las comprobaciones de rutina): coherente con "do it instead of offering it".
- **Sitio:** `flows/research.md:47` y `close.md:63` describen la tarjeta y siguen coherentes.
- **Tres salidas para una oferta (H3):**
  - Si ya está autorizada, se hace.
  - Si es una decisión del usuario, se pregunta.
  - Si no es ninguna de las dos, se omite.

  La tercera sigue a `questions.md:9` (solo se pregunta lo que es del usuario) y a `flow-research/SKILL.md:40` (una respuesta completa no termina en pregunta de ruta). Completa R5-A sin reabrirla: ninguna oferta queda como frase de cierre en texto.
- **`tests/pilot/regression.go:386-424` (`closeQuestionAfterReport`):** da por buena una oferta en texto que termina en "?", así que el fallo de Carvento `1a3ab3ea` contaría como acierto. Con los pilotos pausados no se cambia; se nombra en el pull request (H7).
- **`flow-build/references/verification.md:49` y `content/agents/`:** coherentes.

## Notas para los commits

Cada commit nombra:
- Carvento `1a3ab3ea` (Claude Code, 2026-10-17, con `aa17fbe1` desplegado).
- fleetsystem `02891730` (2026-10-17).
- OpenCode dened `ses_ebcb6cbd7ffe` (2026-10-16).
- La decisión R5-A.

T2 declara el cambio de tamaño de `global.md`.
