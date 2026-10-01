# Correcciones por la revisión de sesiones del 2026-09-30

| Campo | Valor actual |
| --- | --- |
| Estado | Cerrado · integrado en `development` por [PR #86](https://github.com/JhonHawk/tricell-hive/pull/86) (`3988d0b`) y desplegado con `hive update` |
| Tracker · GitHub Issues | Sin issue: el trabajo nace de la revisión de sesiones; no se abre ninguno sin pedirlo |
| Git | `automatic` · PR #86 integrado en `development` |
| Verificación | Pruebas Go (`go test ./...`), prueba de presupuesto de la guía global, búsquedas de texto por criterio, `hive update --dry-run` |
| Siguiente paso | Ninguno. El efecto de H1 se vigila en sesiones reales |

## Objetivo

Hoy se revisaron 16 sesiones reales de ark, sample-project y globex, en Grok, OpenCode y Claude Code. Todas corrían con la guía de la versión a5f3d1f más eddcf41. La revisión encontró nueve hallazgos, H1 a H9. Dos pesan más:

- **H1:** en Grok, las preguntas nativas siguen llegando sin el mensaje que debía precederlas, aunque la regla eddcf41 estaba cargada. Hubo 7 casos hoy y 2 en OpenCode con qwen3.8-flash. El usuario ve la tarjeta de la pregunta sin el informe.
- **H6:** los subagentes de Grok no encuentran las referencias que sus roles enlazan como `skill:owner/path`, y las buscan con `find` en todo el directorio personal. Pasó en 9 sesiones entre el 2026-09-22 y el 2026-09-30.

Este cambio ajusta la guía distribuida para los nueve hallazgos (D1-B) y hace que el instalador escriba en los roles y en el bloque global la ruta real de cada referencia (D3-A). El resultado esperado: el usuario recibe el informe antes de cada pregunta, los subagentes leen sus referencias sin buscarlas, y las conductas observadas hoy quedan cubiertas por un texto que el modelo tiene cargado cuando las necesita.

## Alcance

**Incluye.** Para cada hallazgo, su forma corregida tras verificarlo en la sesión de hoy:

- **H1:** la respuesta que llama la pregunta nativa empieza con el texto completo; el razonamiento no se muestra (D2-A).
- **H2:**
  - El recorrido funcional, también cuando el usuario lo pide a mitad de sesión, va a hijos `hive-verify-change`/`hive-review-ux`; el hilo principal no lo programa.
  - Una comprobación de vida o *smoke* que falla deja el despliegue sin verificar.
- **H3:** un umbral concreto para delegar la investigación.
- **H4:** las preguntas y sus opciones no afirman premisas sin verificar.
- **H5:** los enlaces de tickets se toman del tracker; no se componen.
- **H6:**
  - Los hallazgos de los hijos no se pierden al relevarlos.
  - Los enlaces `skill:` se resuelven al instalar.
- **H7:** un valor obligatorio de `## Hive` que falta se pregunta aunque la documentación del repo sugiera uno.
- **H8:** las etiquetas de opción van en el idioma de la sesión.
- **H9:** la opción de continuar nombra un ticket elegido por el agente.

**Excluye:**

- **Sección `## Hive` de globex:** es otro proyecto y la crea la sesión de globex.
- **Elegir qwen3.8-flash como hilo principal:** es una decisión de modelo, no de guía.
- **Pilotos con modelos:** siguen en pausa. El efecto se mide en sesiones reales posteriores.

**Restricciones:**

- **Presupuesto de la guía global:** `content/guidance/global.md` sigue bajo `globalGuidanceBudget` (44336 bytes; 43649 en la base). Se redacta corto y se condensan los bullets editados. Si aun así no cabe, el presupuesto sube en este cambio al tamaño final más ~1 KiB, según la convención de la propia prueba, con la razón en el PR (ver #47).
- **Pruebas:** la suite local sigue en verde.
- **Enlaces de los roles fuente:** los roles fuente conservan `skill:owner/path`, el localizador de autoría que valida la publicación. Solo cambia lo desplegado.

## Criterios de aceptación

- AC1. La regla de preguntas de `global.md` dice que la respuesta que llama la herramienta de pregunta empieza con el texto completo, que el razonamiento no se muestra, y que una respuesta con solo la llamada es el fallo. *Falso en la base cuando* `grep -c "reasoning is not shown" content/guidance/global.md` da 0.
- AC2. `flow-build/SKILL.md` dice que, cuando corre un recorrido funcional desplegado (porque el proyecto lo declara o el usuario lo pide, también a mitad de sesión), este va a los hijos que nombra la referencia de verificación, y que el hilo principal no lo programa. *Falso en la base cuando* `grep -c "does not script" content/skills/flow-build/SKILL.md` da 0 (la única mención actual de `hive-verify-change` en ese archivo, línea 55, trata la verificación por tarea, no el recorrido).
- AC3. `flow-build/SKILL.md` dice que cualquier comprobación fallida o error observado tras desplegar, como un estado de error en el login, deja el despliegue sin verificar y nunca se reporta como sano. *Falso en la base cuando* `grep -c "never as healthy" content/skills/flow-build/SKILL.md` da 0.
- AC4. `flow-research/SKILL.md` dice que, en la sexta búsqueda o lectura sin respuesta (sin contar la guía del proyecto ni las skills), el hilo principal para y delega el resto, y `flow-plan/SKILL.md` remite a ese criterio para explorar. *Falso en la base cuando* `grep -c "sixth search" content/skills/flow-research/SKILL.md` da 0, o `grep -c "flow-research" content/skills/flow-plan/SKILL.md` da 0.
- AC5. `global.md` dice que el texto y las opciones de una pregunta afirman como hecho solo lo verificado en la sesión y marcan lo demás como supuesto. *Falso en la base cuando* no hay regla sobre premisas en opciones de pregunta (`grep -c "verified in this session" content/guidance/global.md` da 0).
- AC6. `global.md` dice que la URL de un ticket es la que devolvió el tracker o un enlace existente verificado, recortada solo a su forma estable; que sin ella se cita el ID sin enlace; y que nunca se compone el espacio de trabajo. *Falso en la base cuando* `grep -c "cite the ID without a link" content/guidance/global.md` da 0.
- AC7. `global.md` dice que el hilo principal releva, o propone como hallazgo incidental, cada defecto que reporta un hijo, en cualquiera de sus escalas de severidad, sin omitirlo del resumen. *Falso en la base cuando* `grep -c "every defect a child reports" content/guidance/global.md` da 0.
- AC8. Al instalar o actualizar, los roles y el bloque global desplegados reemplazan cada enlace `skill:owner/path` por la ruta del recurso en el directorio de skills. En ámbito de usuario es la ruta absoluta `<home>/.agents/skills/<owner>/<path>`; en ámbito de proyecto, la ruta relativa a la raíz del directorio de skills del host (D6-A). Se cumple en todos los hosts registrados, incluidos los roles TOML de Codex. *Falso en la base cuando* `grep -c "](skill:" ~/.grok/agents/hive-verify-change.md` da 2 (verificado en `dc1a749`).
- AC9. `global.md` dice que, ante un valor obligatorio de `## Hive` ausente, un valor hallado en otra documentación solo sirve como recomendación para la pregunta, que igual se hace y se registra. *Falso en la base cuando* `grep -c "recommendation for that question" content/guidance/global.md` da 0.
- AC10. `global.md` dice que las etiquetas de las opciones de una pregunta van en el idioma de la sesión. La marca de recomendado ya lo exige en `:21`. *Falso en la base cuando* `grep -c "option labels" content/guidance/global.md` da 0.
- AC11. La regla de la pregunta de cierre en `global.md` dice que la opción de continuar nombra un ticket que el agente eligió, y que pedir al usuario que lo nombre no cuenta. *Falso en la base cuando* `grep -c "asks the user to name" content/guidance/global.md` da 0.

## Entrega

Decisiones del usuario, 2026-09-30:

- **D1-B:** los nueve hallazgos cambian la guía.
- **D2-A:** H1 se ataca nombrando el fallo. Si recurre con este texto, se pasa a preguntar en texto (D2-B).
- **D3-A:** el instalador reescribe los enlaces `skill:` al desplegar.
- **D6-A:** en ámbito de proyecto, la reescritura usa la ruta relativa a la raíz.
- **D4-A:** entrega `automatic`. Repositorio `tricell-hive`, base `development`, rama de trabajo `fix/session-findings-2026-09-30` en un worktree, porque otra sesión trabaja en paralelo en este checkout. La secuencia es commits, push, PR a `development` y merge cuando la suite local y la prueba de presupuesto pasan y la revisión queda atendida. Este registro se versiona una sola vez, al cierre, con `flow-close`, mediante un PR pequeño de cierre sin revisión dedicada que mergea el agente. Después del merge, el binario se recompila (hay cambio en `integrations/`) y se ejecuta `hive update` (hay cambio en `content/`).
- **D5-A:** la revisión del código Go es el `/code-review` de Claude Code, antes del merge. El texto de la guía va sin revisión dedicada.
