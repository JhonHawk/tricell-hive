# design-library-detection — La librería de diseño obliga aunque no esté declarada, y el revisor de UX sube de perfil

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #468, merge ff52a22f |
| Tracker | ninguno |
| Git | tricell-hive, base `development` en `aaa0458e`, rama `feat/design-library-detection` |

## Decisiones

2026-11-13
- **Opción A del análisis** — ¿Qué se hace con los hallazgos de la sesión de Codex en fleetsystem? Preparar un cambio en Hive para tres huecos: quién registra archivo y librería en `DESIGN.md` cuando el trabajo es solo diseño, inventario de componentes de la pantalla hermana antes de dibujar, y una primera pantalla antes de multiplicar el conjunto.
- **Añadido del usuario** — Que la etapa detecte cuando lo que se le comparte funciona como librería y no está declarado («ya van varias veces que se la comparto e incluso una fue para crearla»).
- **Añadido del usuario** — Evaluar los modelos y el esfuerzo de las revisiones, que a su juicio se quedan cortos.
- **D8 A** — ¿Quién escribe el archivo y la librería de Figma en `DESIGN.md`? La etapa de diseño, en cuanto el usuario dice que sí, mostrando el cambio; ya no espera a un plan.
- **D9 A** — ¿Cómo se valida la primera pantalla antes de dibujar el resto? Sin pausa: el revisor la revisa, se corrige y el usuario recibe su ubicación; la aprobación sigue siendo una sola para el conjunto.
- **D10 A** — ¿Qué modelo y esfuerzo recibe el revisor de UX? El perfil de los otros revisores: Opus/alto en Claude y GPT-6.1 Sol/alto en Codex.
- **Ajuste a D8 A, decisión menor** — ¿Cuándo se pregunta la entrega de esa edición? En la misma ronda de aprobación, cuando no hay respuesta de entrega vigente para ese repositorio, y la edición se entrega en ese momento; al responder D8 se dijo «al cerrar el trabajo». Cambia solo el momento de la pregunta: evita una edición sin commit que un plan posterior dejaría fuera de su rama.
- **Supuesto** — La revisión de planes y la verificación por tarea no cambian: las sesiones revisadas no muestran fallos suyos.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Solicitud

> A, y que tambien detecte si le paso justamente algo que funciona como libreria y no esta declarado (ya van varias veces que se la comparto e incluso una fue para crearla)
> - Otra cosa, evaluemos los modelos y esfuerzo para las revisiones, creeeeo que se estan quedando cortos

La opción A a la que responde: «Preparo el cambio en Hive — investigo y te propongo un plan para tres huecos: quién registra archivo y librería en `DESIGN.md` cuando el trabajo es solo diseño, inventario de componentes de la pantalla hermana antes de dibujar, y una primera pantalla mostrada antes de multiplicar el conjunto.»

Sin ticket. Fuera de alcance: las correcciones que salen sin revisar después de la última ronda de revisión, que research pregunte decisiones antes de leer los documentos fuente, y las lecturas recortadas de Codex; quedan como casos en observación.

## Objetivo

Hoy la etapa de Figma solo exige las piezas de la librería cuando el `DESIGN.md` de la app la declara; solo pregunta si registrar el archivo, nunca la librería, y deja la escritura a un plan que un trabajo solo de diseño no tiene. Una sesión nueva dibujó 38 marcos con 3 componentes de la librería, y su revisor, en el modelo más chico de los revisores, no lo reportó.

Con el cambio, la etapa fija una librería antes de dibujar (nombrada, compartida, declarada o detectada), inventaría lo que usan las hermanas, valida la primera pantalla antes del resto, escribe archivo y librería en `DESIGN.md` al sí del usuario, y el revisor de UX corre en el perfil de los demás revisores y compara contra hermanas y librería.

## Alcance y aceptación

Dentro: `flow-figma-design` con una referencia nueva y la de tarjetas de reglas, el registro de diseño de `hive-ui-criteria`, el paso de decisiones de UI de `flow-plan`, el rol `hive-review-ux` y su brief, la documentación de entrega de agentes y de recursos de instrucciones, las páginas del sitio de la etapa y del plan, el changelog, y los deltas de `workflow-lifecycle` y `ui-criteria`.

Fuera: los perfiles de modelos en `integrations/agent-profiles.json` (ningún modelo cambia, solo qué perfil usa un rol), los demás roles, la guía global y cualquier archivo de fleetsystem.

- AC1 — La etapa de Figma fija una librería de diseño antes de dibujar aunque el `DESIGN.md` no la declare: la que el usuario nombra o comparte en la conversación, si no la declarada, si no la detectada (habilitada en el archivo, usada por los marcos hermanos o nombrada en el registro de una hermana); un enlace a otro archivo de Figma cuenta como librería nombrada cuando el usuario lo presenta como librería, kit, sistema de diseño o fuente de los componentes, o cuando la conexión lo lista entre las librerías habilitadas del archivo, y no cuando lo enlaza como la pantalla a replicar · falso hoy porque `content/skills/flow-figma-design/SKILL.md` paso 5 solo aplica la regla «When the app's `DESIGN.md` declares a Figma `**Design library:**`» y ningún paso detecta ni acepta una librería de la conversación · Observed in: `content/skills/flow-figma-design/references/design-library.md` y sus punteros en `content/skills/flow-figma-design/SKILL.md`
- AC2 — Antes del primer dibujo, el registro del trabajo lleva un inventario: cada componente y variable que los marcos hermanos instancian o enlazan, de cualquier librería o del propio archivo, y por cada pieza que las pantallas necesitan y ninguna hermana usa, el término buscado en la librería fijada y su resultado; lo que usan las hermanas se reusa con o sin librería fijada, una pieza local solo se define cuando ni las hermanas ni la librería la tienen, y un registro que ya existe conserva sus líneas · falso hoy porque ningún paso pide el inventario y una línea `candidate:` no exige una búsqueda · Observed in: `content/skills/flow-figma-design/references/design-library.md` y `content/skills/flow-figma-design/SKILL.md`
- AC3 — Cuando el trabajo tiene más de una pantalla, la etapa dibuja primero una, la hace revisar, corrige lo que se sostiene y da al usuario su ubicación sin preguntar nada ni terminar el turno, y solo entonces dibuja las demás y los marcos de 440 px con sus piezas; la revisión del conjunto incluye esa primera pantalla y la aprobación sigue siendo una sola · falso hoy porque el paso 5 dibuja todas las pantallas y después pide una revisión · Observed in: `content/skills/flow-figma-design/SKILL.md`
- AC4 — Cuando el archivo usado o la librería fijada no están declarados, o el `DESIGN.md` declara otro archivo u otra librería de Figma, la etapa pregunta una vez, en la ronda de aprobación, si registrarlos, mostrando cada valor junto al que reemplaza, y con un sí escribe ella misma los campos `**Design file:**` y `**Design library:**` del `DESIGN.md` en ese turno, sin tocar nada más y mostrando el cambio; esa entrega no cambia cómo cierra el trabajo de diseño; pregunta la entrega en esa misma ronda cuando no hay respuesta vigente para ese repositorio y entrega la edición en ese momento, como cambio mecánico; sin `DESIGN.md` un sí no escribe nada y queda como pendiente, y un no queda anotado para no volver a preguntar · falso hoy porque el paso 6 solo pregunta por el archivo y dice «this stage does not edit `DESIGN.md`», y `design-record.md` «Writing» solo permite escribir por una tarea de plan o a petición del usuario · Observed in: `content/skills/flow-figma-design/references/design-library.md`, `content/skills/flow-figma-design/SKILL.md` y `content/skills/hive-ui-criteria/references/design-record.md`
- AC5 — El plan añade la tarea que escribe archivo y librería solo cuando el registro guarda un sí que la etapa no pudo escribir; las convenciones siguen yendo por tarea de plan · falso hoy porque `content/skills/flow-plan/SKILL.md` «UI decisions» paso 4 añade «the task that writes the `**Design file:**` field» ante cualquier sí del registro, y no nombra la librería · Observed in: `content/skills/flow-plan/SKILL.md`
- AC6 — Para marcos en un archivo de diseño, el revisor de UX lee la imagen de cada hermana además de su árbol de nodos, reporta la pieza dibujada fuera de la librería que su brief nombra aunque el `DESIGN.md` no la declare, busca en esa librería cada pieza `candidate:` y reporta la que la librería sí tiene, reporta la pieza que el inventario muestra en las hermanas y no se dibujó como instancia, no aprueba mientras falte una comparación o búsqueda que la conexión le permite hacer, y reporta como no corrida la que la conexión no permite; su brief le da la librería fijada, o que no hay, y el inventario · falso hoy porque `content/agents/review/hive-review-ux.md:10` condiciona el hallazgo a «when the app's `DESIGN.md` declares a design library», no pide la imagen de las hermanas ni buscar las candidatas, y «Reviewer brief» da «the declared design library» · Observed in: `content/agents/review/hive-review-ux.md` y `content/skills/flow-build/references/delegation.md`
- AC7 — El rol `hive-review-ux` se genera con Opus/alto para Claude y GPT-6.1 Sol/alto para Codex · falso hoy porque su `model_profile` es `focused` y los roles instalados dicen `sonnet`/`high` y `gpt-6.1-terra`/`medium` · Observed in: el plan de despliegue que genera el gestor compilado del candidato (`hive update --source . --out FILE`, sin aplicar)
- AC8 — La documentación de entrega de agentes, la página del sitio de la etapa de Figma y las especificaciones describen lo anterior · falso hoy porque `_support/docs/architecture/agent-delivery.md:19` lista `hive-review-ux` en el perfil `focused`, y `site/src/content/docs/flows/figma-design.md:28`, `:43` y `:49` dicen que la etapa solo pregunta por el archivo, que el plan lo escribe y que no edita `DESIGN.md`, y `_support/docs/architecture/instruction-resources.md:19` da esos campos a `design-record.md` sin nombrar la referencia nueva · Observed in: `_support/docs/architecture/agent-delivery.md`, `_support/docs/architecture/instruction-resources.md`, `site/src/content/docs/flows/figma-design.md`, `site/src/content/docs/flows/plan.md` y los deltas de esta carpeta

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

- Reglas leídas en `aaa0458e` (2026-11-13): `content/skills/flow-figma-design/SKILL.md` (15,978 bytes, tope 16,000), `content/skills/hive-ui-criteria/references/design-record.md`, `content/agents/review/hive-review-ux.md`, `content/skills/flow-build/references/delegation.md` «Reviewer brief», `content/skills/flow-plan/SKILL.md` «UI decisions», `integrations/agent-profiles.json`, `_support/docs/architecture/agent-delivery.md`, `_support/docs/harness-engineering/model-floor.md`.
- Perfiles: `hive-review-ux` usa `focused` (Claude `sonnet`/`high`, Codex `gpt-6.1-terra`/`medium`); `hive-review-code` usa `reasoning` (`opus`/`high`, `gpt-6.1-sol`/`high`). `focused` lo siguen usando `hive-read-state` y `hive-write-docs`, así que la tabla del piso de modelos no cambia.
- Observación 1 (sesión real): Codex 0.169.0, `gpt-6.1-sol` esfuerzo alto, sesión `01a1d465`, 2026-11-13, proyecto fleetsystem. Dibujó 38 marcos con 3 componentes de librería (StatCard, Select, Tabs); el `DESIGN.md` de la app no declara archivo ni librería; el usuario corrigió después («recuerda usar los componentes de la libreria»). Su revisor, hijo `01a1d47c` en `gpt-6.1-terra` esfuerzo medio, reportó 7 hallazgos de cifras y estados y ninguno de piezas locales; no llamó a la búsqueda de la librería ni a la lista de librerías.
- Observación 2 (sesión real): Claude Code, sesión `62cf4d39`, 2026-11-13, proyecto fleetsystem. El revisor `hive-review-ux` en `claude-sonnet-5-5` esfuerzo alto reportó que el armazón y las columnas coincidían con la pantalla hermana; el usuario corrigió después el encabezado gris de tablas hechas a mano. El mismo revisor declaró que solo capturó una hermana y no la revisó.
- Recuento de dos hijos de investigación sobre las sesiones del 10 al 13 de noviembre: tras un veredicto del revisor de UX, el usuario corrigió defectos visibles en los marcos 6 veces en Codex (15 revisiones en `gpt-6.1-terra` medio) y 3 o 4 veces en Claude (30 revisiones en Sonnet alto). La verificación por tarea (`gpt-6.1-sol` alto, Opus alto) halló defectos concretos sin corrección posterior del usuario.
- Documentación (2026-11-13): Claude Code acepta en un subagente los modelos `sonnet`, `opus`, `haiku`, `fable` y esfuerzos `low` a `max` (code.claude.com/docs/en/sub-agents); la de Codex recomienda `high` para agentes revisores (learn.chatgpt.com/docs, «Reasoning effort»).
- Lo que no se pudo separar: cuánto de lo que el revisor deja pasar viene del modelo y cuánto de sus instrucciones; las revisiones de UX solo han corrido en Sonnet y Terra. El cambio toca ambas cosas.
- Naturaleza de cada regla, para la sección «Measurement» de `AGENTS.md`: la librería sin declarar, el inventario y la primera pantalla son huecos o requisitos nuevos que la decisión del usuario adopta; la escritura de `DESIGN.md` y el perfil del revisor cambian reglas existentes por decisión del usuario (D8 A, D10 A); la comparación del revisor contra hermanas es una corrección con dos observaciones independientes en modelos del piso y la aprobación del usuario.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Entrega automática y revisión nativa de Claude Code `/code-review`, las respuestas vigentes de la sesión para este repositorio: revisión, push, PR hacia `development` y merge sin detenerse. Después del merge, el refresco de la instalación local que el checkout del mantenedor autoriza. No autoriza un release ni la promoción a `master`.

## Estado de revisión y progreso

2026-11-13 — Plan escrito. Revisión del plan (`hive-review-plan`): sin P0; aplicados un P1 (la edición de `DESIGN.md` no tenía ruta de entrega: ahora se pregunta en la ronda de aprobación y se entrega al momento), seis P2 (presupuesto de bytes del `SKILL.md`, un solo hogar para la escritura, `rule-cards.md`, verificación de AC7 tras el commit, pasadas de revisión y momento del registro en el delta) y cinco P3 (mapa de recursos, dos líneas del sitio, impacto en Pi y en `flow-verify`, observaciones en el commit, redacción de AC3 y AC5).

2026-11-13 — Revisión de código (`/code-review` sobre `67da38bc`): diez hallazgos, todos confirmados contra el texto y corregidos en una ronda. Cambian la redacción de AC1 (un archivo enlazado como pantalla a replicar no es librería), AC2 (lo que usan las hermanas se reusa con o sin librería; el registro existente se conserva), AC4 (un campo con otro valor se reemplaza solo tras mostrar ambos; sin `DESIGN.md` solo un sí queda pendiente; la entrega de la edición no cambia el cierre del trabajo de diseño) y AC6 (el veredicto no pasa mientras falte una comparación que la conexión permite). Además: la revisión de la primera pantalla no cuenta como la primera revisión de las tarjetas de reglas, entrada en `CHANGELOG.md` y una frase del sitio.

2026-11-13 — Segunda revisión de código, limitada a `67da38bc..c59f7b88`, y revisión del plan revisado (`hive-review-plan`): sin P0 ni P1; corregidas en un último commit, que ya no pasa por otra revisión, las contradicciones que quedaban: la librería fijada gana sobre el componente de una hermana y lo de las hermanas cubre lo que le falta; un archivo enlazado como pantalla a replicar no es librería aunque esté habilitado; con más de una librería dada en la conversación se pregunta antes de tomar otra fuente; solo se escriben los campos que la pregunta mostró; la entrega de la edición no dispara el cierre; el veredicto por comparación no corrida no aplica a la relectura; al reentrar con otra librería se repiten las búsquedas; y los deltas, el sitio y el changelog dicen lo mismo.
