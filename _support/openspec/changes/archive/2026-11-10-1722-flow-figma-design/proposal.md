# flow-figma-design — Etapa opcional de diseño en Figma entre research y plan

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #450, merge 7c1bf212 |
| Tracker | ninguno (sin issue; búsqueda en GitHub Issues sin coincidencias, 2026-11-10) |
| Git | tricell-hive, base `development` en `a1c6c9ac`, rama propuesta `feat/flow-figma-design` |

## Decisiones

2026-11-10
- **D3 A** — ¿Con qué enfoque se construye la etapa de diseño? Una skill propia, sin etapa nueva en el visor: la sección «Examples for a screen» sale de `flow-research` y el protocolo del visor no cambia.
- **D4 A** — ¿Cuándo recomienda la investigación pasar por diseño? Solo cuando el trabajo trae una pantalla nueva, o una que agrega, quita o mueve una región, y nada fija su composición; en otro caso no se menciona.
- **D5 A** — ¿Qué hace la etapa cuando hay conexión a Figma pero el proyecto no declara su archivo? Pide el enlace una vez, en el texto, con la salida de ir al plan.
- **D6 A** — ¿Se anota el archivo de diseño en el `DESIGN.md` del proyecto? La etapa lo pregunta una vez, junto con la aprobación del ejemplo, cuando el archivo aún no está declarado.
- **D7 A** — ¿Puede el hijo que dibuja crear componentes y variables? Sí, solo los propios del trabajo, en su página; cambiar lo existente o una librería compartida vuelve al usuario como decisión.
- **D8 A** — ¿Entran ya las cuatro correcciones observadas en cam? Sí, como pasos de la skill; la aprobación del usuario sustituye la segunda observación que pide «Measurement».
- **D9** — ¿Qué produce la etapa sin herramienta de diseño? Quedó resuelta por la decisión de nombre y alcance (D10 A): sin conexión a Figma no hay etapa, y los bocetos en texto quedan solo en el plan.
- **D10 A** — ¿La skill es específica de Figma o neutral? Específica: se llama `flow-figma-design`; el campo del `DESIGN.md` sigue llamándose «Design file».
- **D11 A** — ¿Cómo se entrega el cambio? Automática: se integra en `development` sin parada de validación.
- **D12 A** — ¿Qué revisión de código corre sobre el candidato? La nativa de Claude Code, `/code-review`, antes del push.

## Solicitud

Del usuario, en la conversación del 2026-11-10: «solo en fleetsystem, nunca me ofrecio usar figma para presentarme propuestas»; «no deberia ser un paso intermedio? entre research - plan? research - design ui - plan»; «creo que por ahi teniamos una skill hecha de diseño no? podemos evaluarla si sigue aplicando o debemos considerar algo mas»; «Tambien consideremos qué pasa si el usuario no tiene figma»; «amenos que la skill sea flow-figma-design».

Sin ticket. Fuera de alcance: mostrar las propuestas en el visor, maquetas HTML, una preferencia por proyecto del tipo «vistas previas siempre en Figma», y otras herramientas de diseño.

## Objetivo

Hoy la oferta de propuestas de pantalla vive en una sección de `flow-research` y solo se hace cuando el usuario pidió un cambio; la evaluación de un ticket seguida de `/flow-plan` nunca la alcanza, y el plan no la tiene. El archivo de Figma que el usuario nombra no se anota en `DESIGN.md`, así que la siguiente sesión tampoco lo conoce.

Con el cambio, el diseño es una etapa opcional con comando propio, `flow-figma-design`, entre la investigación y el plan. La investigación la recomienda en su bloque de siguiente paso, también tras evaluar un ticket, cuando hay una pantalla nueva o recompuesta y el host tiene una conexión que escribe en Figma. La etapa resuelve el archivo, dibuja, pide la elección y deja el registro que el plan, la construcción y la verificación ya consumen.

## Alcance y aceptación

Dentro: la skill nueva; el recableado de `flow-research` y de la guía global; los apuntadores de `flow-plan`, `delegation`, `hive-design-ui`, `design-record` y `unattended-delegation`; el sitio, `README.md`, `llms.txt`, `instruction-resources.md` y `CHANGELOG.md`; los deltas de `workflow-lifecycle` y `ui-criteria`.

Fuera: el visor y su protocolo, las pruebas de piloto (en pausa), y el comportamiento de comparación de la construcción y la verificación frente al ejemplo aprobado, que no cambia; en `flow-verify` solo cambian el enlace a la sección retirada y la mención del boceto en texto.

- AC1 — Existe `content/skills/flow-figma-design/SKILL.md`, con `name` igual a su carpeta y a lo sumo 16.000 bytes, y enuncia en este orden: que la respuesta del usuario que inicia la etapa es su instrucción para escribir en el archivo de Figma lo que este trabajo crea, sin sustituir la pregunta propia de un efecto crítico; cargar las instrucciones propias de la conexión de Figma y el sistema de diseño del proyecto; resolver el archivo (declarado en `DESIGN.md`, nombrado por el usuario, o pedido una vez); avisar antes de dibujar los pasos manuales que el usuario tendrá que hacer en Figma; dibujar dos o tres ejemplos, o uno cuando la composición está fijada; pedir la elección o corrección; dibujar y hacer aprobar el frame de 440 px cuando la pantalla se sirve en móvil; registrar el ejemplo y actualizar su estado al aprobarse; y terminar con el bloque de siguiente paso hacia `flow-plan` · falso hoy porque el archivo no existe · Observed in: el archivo en el candidato y `go test ./tests/content/...`
- AC2 — `flow-research` ya no contiene la sección «Examples for a screen» ni una oferta por tarjeta; su paso 5 recomienda `flow-figma-design` en el bloque de siguiente paso cuando se cumplen las condiciones de pantalla y el host tiene una conexión que escribe en Figma, incluida la terminación de un ticket evaluado, y sin esa conexión o sin esas condiciones no lo menciona; su paso 0 envía a esa etapa, con el bloque de siguiente paso, la petición cuyo entregable es el diseño · falso hoy porque `content/skills/flow-research/SKILL.md:35` excluye la evaluación de un ticket y la sección ocupa las líneas 31–51 · Observed in: `content/skills/flow-research/SKILL.md` en el candidato
- AC3 — Ningún archivo de `content/`, `site/`, `_support/docs/architecture/`, `README.md`, `llms.txt` o `CHANGELOG.md` enlaza a `#examples-for-a-screen` ni dice que la investigación dibuja u ofrece los ejemplos; todos nombran a `flow-figma-design`, y la regla de delegación tiene un punto propio para el hijo que dibuja en esa etapa · falso hoy porque `rg -n -i "examples-for-a-screen|research left an approved|in research, .*design file|offer of screen proposals|Screen examples in research" content site _support/docs/architecture README.md llms.txt CHANGELOG.md` devuelve coincidencias en `flow-plan`, `flow-verify`, `delegation`, `hive-design-ui`, `design-record`, `unattended-delegation`, `flow-research`, el sitio y el changelog · Observed in: la salida vacía de ese `rg` en el candidato final
- AC4 — El rol `hive-design-ui` permite crear componentes y variables locales al archivo, creados por este trabajo y anotados por su identificador en el registro del ejemplo, y borrar solo lo que creó, incluidos los que solo usaban los ejemplos no elegidos, y devuelve al padre como decisión pendiente cualquier cambio a variables, estilos o componentes existentes o de una librería compartida · falso hoy porque `content/agents/design/hive-design-ui.md:11` dice «Never touch variables, styles, components or other pages» · Observed in: `content/agents/design/hive-design-ui.md` en el candidato
- AC5 — Cuando el archivo de Figma que se usó no está declarado en el `DESIGN.md` de la app, la etapa pregunta una vez, en la misma ronda que la aprobación del ejemplo, si lo anota; con un sí deja la instrucción en el registro del ejemplo, y el plan de esa pantalla agrega la tarea que escribe el campo «Design file», de modo que la edición del `DESIGN.md` se entrega con el cambio; cuando el entregable era el diseño y no hay plan, el cierre la nombra como pendiente · falso hoy porque ningún paso hace esa pregunta (`design-record.md:32` solo dice «write it only at the user's instruction») · Observed in: `content/skills/flow-figma-design/SKILL.md`, `content/skills/flow-plan/SKILL.md` y `content/skills/hive-ui-criteria/references/design-record.md` en el candidato
- AC6 — La guía global nombra las transiciones investigación → diseño en Figma → plan en su regla del bloque de siguiente paso, y un diseño en Figma entre los procedimientos por los que se ofrece el comando de un flujo; `README.md` (diagrama y lista), `llms.txt` y el sitio (página `flows/figma-design` con las secciones que exige `documentation-site`, entrada en la barra lateral y en las tablas de flujos) describen la etapa; el sitio compila · falso hoy porque `content/guidance/global.md:90` lista solo «research to plan or build» sin esa etapa ni ese procedimiento, y no existe `site/src/content/docs/flows/figma-design.md` · Observed in: esos archivos en el candidato y `npm run build` en `site/`
- AC7 — **Replaces:** AC2 — `flow-research` ya no contiene la sección «Examples for a screen» ni una oferta por tarjeta; su paso 5 recomienda `flow-figma-design` en el bloque de siguiente paso cuando se cumplen las condiciones de pantalla y el host tiene una conexión que escribe en Figma, incluida la terminación de un ticket evaluado, y sin esa conexión o sin esas condiciones no lo menciona; su paso 0 continúa hacia esa etapa en el mismo turno para la petición cuyo entregable es el diseño cuando el host tiene esa conexión, y sin ella reporta el bloqueo · falso hoy porque `content/skills/flow-research/SKILL.md:35` excluye la evaluación de un ticket y la sección ocupa las líneas 31–51 · Observed in: `content/skills/flow-research/SKILL.md` en el candidato
- AC8 — **Replaces:** AC1 — Existe `content/skills/flow-figma-design/SKILL.md`, con `name` igual a su carpeta y a lo sumo 16.000 bytes, y enuncia en este orden: que la respuesta del usuario que inicia la etapa es su instrucción para escribir en el archivo de Figma lo que este trabajo crea, sin sustituir la pregunta propia de un efecto crítico; cargar las instrucciones propias de la conexión de Figma y el sistema de diseño del proyecto; resolver el archivo (declarado en `DESIGN.md`, nombrado por el usuario, o pedido una vez); avisar antes de dibujar los pasos manuales que el usuario tendrá que hacer en Figma; dibujar dos o tres ejemplos, o uno cuando la composición está fijada, y registrarlos como pendientes de aprobación antes de preguntar; pedir la elección o corrección; dibujar y hacer aprobar el frame de 440 px cuando la pantalla se sirve en móvil; actualizar el registro a aprobado en el mismo turno de la aprobación; y terminar con el bloque de siguiente paso hacia `flow-plan`, salvo la petición cuyo entregable es el diseño, que cierra como tarea operativa · falso hoy porque el archivo no existe · Observed in: el archivo en el candidato y `go test ./tests/content/...`

## Prerrequisitos

None: the build needs nothing from outside the repository.

## Contexto verificado

Leído y observado el 2026-11-10, sobre `development` en `a1c6c9ac`:

- La sección actual está en `content/skills/flow-research/SKILL.md:31-51` (4.986 bytes de 12.283) y la apuntan las líneas 8, 10, 13 y 27 del mismo archivo.
- La consumen `flow-plan/SKILL.md:22-23`, `flow-plan/references/plan-handoff.md:12`, `flow-plan/assets/change-folder/design.md:21`, `flow-build/SKILL.md:15,18`, `flow-build/references/delegation.md:11,38`, `flow-verify/SKILL.md:12,15,16`, `agents/review/hive-review-ux.md:10`, `agents/design/hive-design-ui.md:11,15,17`, `hive-ui-criteria/references/design-record.md:32,63` y `unattended-delegation/SKILL.md:14`.
- La describen `_support/openspec/specs/workflow-lifecycle/spec.md:322-362`, `_support/openspec/specs/ui-criteria/spec.md:25,42-45`, `_support/docs/architecture/instruction-resources.md:19`, y en el sitio `flows/research.md:41,44`, `flows/overview.md:32`, `flows/plan.md:26`, `flows/build.md:33` y `flows/verify.md:26`.
- Enumeran los flujos `content/guidance/global.md:90`, `README.md:355-399`, `llms.txt:30-40`, `site/astro.config.mjs:76-88`, `site/src/content/docs/index.mdx:55-90`, `skills.md:14-21` y `getting-started.md:100`.
- Ningún manifiesto ni prueba de contenido enumera las skills: el despliegue las descubre por patrón (`tooling/management/plan.go:229`) y `tests/content/skill_size_test.go` y `skill_frontmatter_test.go` también. La sintaxis de comandos por host es un patrón (`global.md:94`), no una lista por flujo.
- `flow-close/SKILL.md` está a 10 bytes del tope de 16.000; este cambio no lo toca.
- Sesiones que motivan el cambio: Claude Code, `claude-fable-5-2`, sesión `c71dd314` (fleetsystem, 2026-11-10), donde la evaluación de FSR-855/856 seguida de `/flow-plan` no alcanzó la oferta; y Claude Code, `claude-fable-5-2`, sesión `01acdb18` (cam, 2026-11-10), donde el dibujo no cargó el sistema de diseño, usó 390 px en móvil, dejó el registro sin actualizar tras la aprobación y no avisó los pasos manuales. Ambos modelos están fuera del piso de soporte; la adopción descansa en la aprobación del usuario (D8 A) y en que la ruta por tickets es un hueco sin regla.

## Effects

None: the tasks take no effect on an external system.

## Entrega

Entrega automática (D11 A): revisión, push, pull request e integración en `development` sin parada de validación. Revisión de código nativa de Claude Code, `/code-review` (D12 A), sobre el candidato antes del push. No autoriza una publicación de versión ni cambios fuera de este repositorio. Tras integrarse en `development`, el refresco de la instalación local sigue la regla del checkout del mantenedor.

## Estado de revisión y progreso

2026-11-10 — Plan escrito.
2026-11-10 — Revisión del plan (`hive-review-plan`): 12 hallazgos, todos aplicados. Cambió AC1 (quién autoriza las escrituras en Figma), AC2 (ruta de la petición cuyo entregable es el diseño), AC3 (redacción «en la investigación» y un solo comando sobre todos los lugares), AC4 (variables locales al archivo, con sus identificadores), AC5 (la edición de `DESIGN.md` se entrega con el plan de la pantalla) y AC6 (la cláusula que limita qué comandos se ofrecen); T3 gana `flow-verify`, T4 reescribe la entrada del changelog y se agrega el delta de `documentation-site`. No invalidó ninguna tarea hecha.
2026-11-10 — Revisión de código (`/code-review`, 10 hallazgos sobre `d9b53f5d`), todos aplicados. Cambió AC2: la petición cuyo entregable es el diseño continúa hacia la etapa en el mismo turno, como hacía antes la investigación, en lugar de detenerse en un bloque; la guía global nombra esa cuarta continuación. El resto son precisiones de la skill: el registro se escribe como pendiente al dibujar, un solo ancho por encargo (1440 px por defecto en escritorio), comprobar que la conexión puede escribir en el archivo resuelto, confirmar un archivo que nadie nombró, el archivo declarado debe ser de Figma, los casos de trabajo desatendido, y el plan solo toma un registro aprobado. Reabrió T1, T2 y T3 para esas correcciones, hechas en `b40af1bf`.
2026-11-10 — Segunda revisión del plan (`hive-review-plan`, sobre `b40af1bf`): 9 hallazgos, todos aplicados. El criterio de la investigación y el de la skill pasan a AC7 y AC8, que reemplazan a los dos primeros; el delta deja de contradecirse sobre quién inicia la etapa y cómo termina la petición cuyo entregable es el diseño; la página de investigación del sitio describe la continuación.
