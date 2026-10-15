# Diseño

## Contexto verificado

Revisado el 2026-10-13 sobre `feat/flash-class-guidance` en `34b4edc9`, que es la guía desplegada en los seis hosts y la rama de prueba activa. `development` (`a4d6e254`) aún tiene la delegación dentro de `global.md`; esta rama la movió a `content/skills/flow-research/references/delegation.md` (commit `d2673600`).

Hechos inspeccionados:

- `content/agents/review/hive-research.md` (1.225 bytes). El paso 4 pide «a concise evidence-backed synthesis, useful alternatives, and consequential uncertainties»; no pide estado, ubicaciones, extractos, contraevidencia, lo no revisado ni tope. El paso 2 usa «source facts, live observations, inference, and unanswered questions».
- `content/skills/flow-research/SKILL.md` línea 20: pide al padre «Ask for source-backed findings, counterevidence, and access limits» y «The main thread checks consequential claims against their sources». No hay condición de parada ni criterio de qué comprobar. Línea 16: el estado del proyecto se reconcilia con backlog, PR, CI y despliegues, sin rol asignado. Línea 24: «current observed behavior, documented intent, inference, and recommendation».
- `content/skills/flow-research/references/delegation.md`: «Split and announce work» (línea 11) no dice lanzar juntos los hijos independientes; «Supply a bounded brief» (línea 27) etiqueta premisas como «verified facts, source claims, or assumptions»; ya pide pasar decisiones de la tarea (línea 35).
- `content/guidance/global.md` línea 60, regla de evidencia: «inspected facts, source claims, inference, and recommendations». Es el vocabulario canónico.
- `rg -n "hive-read-state" content/skills` no devuelve nada.
- `tests/content/reading_footprint_test.go` y `budget_test.go` solo registran tamaños; no hay límites que romper.
- `site/src/content/docs/flows/research.md` líneas 35 y 71–72 describen ayudantes y verificación.

Evidencia de sesiones reales (análisis de transcripciones del 2026-10-13, 48 episodios de research con hijos en Claude Code, 25-09 a 13-10):

- Hijo más lento = 58 % de la mediana del tiempo hasta la respuesta (p90 79 %); hijo: mediana 215 s, 153 k tokens, 33 llamadas.
- El hilo principal repite exactamente el 15 % de sus llamadas que ya hizo un hijo (75 de 493), en 23 de 48 episodios; los peores: Claude `02eece2f` (12 de 31), `db4d02b1` del 2026-10-12 (11 de 14), `66c195b6` del 2026-10-13 (11 de 40). Parte son comprobaciones deliberadas.
- 11 de 34 episodios con varios hijos los lanzaron en tandas escalonadas.
- Un hijo de 3.943 s esperando un trabajo externo con bucles `sleep` (Claude `d4947fae`, 2026-10-06).
- `hive-read-state`: 0 lanzamientos en todos los hosts; 3 preguntas de estado resueltas en el hilo principal.
- En OpenCode, `hive-research` corre ahora con `opencode-go/deepseek-v4.1-flash#medium` (commit `569f10c9`, instalado), un modelo de clase flash: el contrato tiene que ser explícito para él.

Evidencia externa que sostiene el diseño: Anthropic, «How we built our multi-agent research system» (2025-06-13) y «Effective context engineering» (2025-09-29), retorno condensado de 1.000–2.000 tokens y referencias ligeras; guía de prompting de GPT-5 de OpenAI, parar cuando el objetivo es nombrable o los resultados convergen; DeepTRACE (arXiv 2509.04499), precisión de citas del 40–80 % en sistemas públicos; repositorios de referencia con esquema de estado y paso de verificación por clases de fallo.

## Diseño elegido

D-1. **El contrato de retorno vive en el rol, no en cada brief.** El rol lo carga siempre, incluido el respaldo con hijo genérico, que recibe el contrato del rol (`delegation.md`, «Select the role»). La skill deja de repetir qué pedir y remite al contrato.

D-2. **Comprobar = abrir la ubicación citada.** El padre solo comprueba las afirmaciones que deciden la respuesta o la recomendación; vuelve a buscar solo si falta la ubicación, no sostiene la afirmación o choca con otra fuente. Las premisas de alto impacto que esa comprobación no resuelve van juntas a un único `hive-refute-claim`. Sin número fijo de comprobaciones, en línea con `diagnosis.md`.

D-3. **Parada en el brief de research.** Respuesta nombrable con evidencia convergente (`complete`), o alcance asignado agotado sin convergencia (`partial`); una fuente no disponible o un trabajo del que depende aún en curso es `blocked`. Salvo que el brief pida esperar, el hijo devuelve el trabajo en curso como `blocked` nombrando qué corre, en lugar de esperar; la espera pedida sigue las reglas de espera de `delegation.md`. No se añade una regla contra bucles `sleep`: solo hay una sesión (`d4947fae`) y queda como vigilancia.

D-4. **Lanzar juntos y explorar primero.** Regla general de delegación, porque aplica a cualquier reparto, no solo a research.

D-5. **`hive-read-state` para estado largo.** Se nombra en el párrafo de estado del proyecto de la skill. El rol exige objetivos identificados; una consulta de tracker con proyecto y estados cuenta como tal (sus identificadores declarados incluyen el repositorio o proyecto), lo que cubre también las preguntas de estado del backlog.

D-6. **Vocabulario único.** El de `global.md` línea 60; las premisas delegadas conservan «assumptions» porque son supuestos, no inferencias del hijo. El veredicto de `hive-refute-claim` (confirmed, refuted, plausible, unverifiable) es otra cosa y no cambia.

Tope de extensión del hijo: unas 1.500 palabras, sin salida cruda. Decisión menor tomada por el orquestador a partir de la guía de Anthropic; la mediana actual de los informes finales del hilo principal es de unas 550 palabras.

## Texto propuesto

`hive-research.md`, paso 2:

> 2. Label each finding as an inspected fact, a source claim, an inference, or an unknown. State search coverage before claiming something is absent.

`hive-research.md`, paso 4 (sustituye al actual):

> 4. Return a reply the parent can read in full, about 1,500 words at most, with no raw output:
>    - The answer in one or two sentences, or why none can be given yet.
>    - Status: `complete` when the answer converged, `partial` when the assigned scope ran out first, or `blocked` when a source was unavailable or work you depend on is still running. Unless the brief asks you to wait, return still-running work as `blocked`, naming what is running. When the brief asks you to wait, follow the wait procedure it names.
>    - Each claim that decides the answer or a recommendation, with its label, its locator (path:line, or URL with version or date) and a short verbatim excerpt.
>    - The counterevidence you looked for and what you found.
>    - What you did not check, and the uncertainties that matter.
>    - Useful alternatives, when the question asks for them.
>
>    Do not implement changes.

`flow-research/SKILL.md`, línea 16, al final del párrafo de estado del proyecto:

> When collecting the state of named tickets, pull requests, CI runs, or deployments, or a tracker query with its project and states, would produce long output, delegate that collection to the `hive-read-state` role with those targets and the fields the report groups by.

`flow-research/SKILL.md`, línea 20: se sustituye «Ask for source-backed findings, counterevidence, and access limits. The main thread checks consequential claims against their sources and resolves conflicting conclusions; do not treat agreement among subagents as proof.» por:

> In each research child's brief, state what is already known or settled, what is out of scope, and when to stop: once the answer can be named and the evidence converges, or once the assigned scope is exhausted. The role's contract sets what the child returns. Check the claims that decide the answer or the recommendation by opening their cited locators, not by repeating the child's searches; search again only when a locator is missing, does not support its claim, or conflicts with another source. Send high-consequence premises that this check cannot settle to one `hive-refute-claim` child, asking for one verdict per premise. Resolve conflicting conclusions yourself; agreement among subagents is not proof.

`flow-research/SKILL.md`, línea 24: «Distinguish current observed behavior, documented intent, inference, and recommendation.» pasa a:

> Distinguish inspected facts, source claims, inference, and recommendations.

`delegation.md`, «Split and announce work», tras «Delegate independent bounded work; sequence dependencies and avoid overlapping writers.»:

> Launch independent children in the same step, not in waves. When your own bounded lookup has not found where the answer lives, launch one child to locate it first and split the rest by what it finds.

`delegation.md`, línea 27: «verified facts, source claims, or assumptions» pasa a «inspected facts, source claims, or assumptions».

Sitio, línea 35, sustituye la frase de ayudantes:

> For large questions the agent hands reading work to helper agents. Each helper says where every key claim comes from and what it did not check; the agent opens the cited lines of the claims that decide the answer before reporting. Long ticket, CI or deployment status is collected by a helper too.

`CHANGELOG.md`, bajo `[Unreleased]` → `### Changed` (cambio de nivel patch según `CONTRIBUTING.md`):

> - `hive-research` returns a bounded reply: the answer, a status with its stop reason, each deciding claim with a locator and excerpt, counterevidence and what it did not check. `flow-research` checks those claims by opening their locators, gives each research child a stop condition, routes long state collection to `hive-read-state`, and uses the global evidence labels; independent children launch together.

El implementador puede ajustar la redacción si conserva cada elemento y los términos clave que verifican las tareas.
