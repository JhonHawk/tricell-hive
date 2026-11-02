# Diseño

## Contexto verificado (2026-10-27, base `840807d9`)

- **Bucles existentes** (investigación A2): research (ruta iterativa con límite y 7 paradas), plan (revisión con una re-revisión), build (`hive-verify-task` hasta 3 rondas), verify (revisión de interfaz, una revisión de código completa y una re-revisión, rondas de corrección sin tope), integración (una re-ejecución de CI por infraestructura), deploy (defecto → build → paso 2).
- **Dónde se decide hoy un cambio pedido:** `global.md:21` (temas abiertos: aclara / relacionado / aparte), `questions.md:84` (una respuesta puede devolver el trabajo a un flujo anterior), `plan-revision.md:11-12` (corrección frente a tarea añadida y cambio nuevo, solo con plan guardado), `flow-build/SKILL.md:20,38`, `flow-verify/SKILL.md:55-59`.
- **Huecos:** G1 sin frontera entre tarea añadida y cambio nuevo; G2 trabajo sin plan que crece sin disparador hacia `flow-plan`; G3 pausa de validación sin plan sin distinción corrección/tarea añadida y sin tope; G4 primera vista rechazada sin rondas; G5 cambio después de integrar y antes de cerrar sin ruta; G6 cambios del conjunto de tickets durante deploy sin ruta; G7 el diagrama del README omite los bucles. Además, `interactive` sin efecto visual no tiene forma de mostrar una corrección (las rondas capturan vistas).
- **Evidencia externa:** OpenSpec decide «actualizar o crear cambio nuevo» por intención (`docs/workflows.md:422-460` en el clon de referencia, v1.14.2); Shape Up y Scrum fijan el objetivo y negocian el alcance; Claude Code recomienda reiniciar tras dos correcciones fallidas; Spec Kit, Kiro y OpenSpec no definen cómo retomar.
- **Límites:** `flow-build/SKILL.md` 15 468 bytes y `flow-close/SKILL.md` 15 572 bytes con tope de 16 000; el sitio no renderiza mermaid; ninguna prueba ni detector de `tests/pilot` depende del texto que se cambia (`go test ./tests/content` pasó en la base).

## Enfoque

### Referencia canónica `scope-changes.md` (D3-A)

Vive junto a `plan-revision.md` porque sus salidas son una revisión del plan o un plan nuevo, pero se lee desde cualquier flujo. Contenido previsto (en inglés, como todo el contenido distribuido):

- `## When to read`: cuando el usuario pide, durante trabajo abierto, algo que cambiaría ese trabajo (en cualquier flujo: research, plan, build, verify, después de integrar, close o deploy), o cuando el trabajo descubre que necesita más de lo pedido; también cuando una respuesta a una pregunta añade alcance.
- **Clasificar por intención (D4-A).** La intención es el resultado que nombra la solicitud citada o sus `AC<n>`:
  1. **Corrección**: cambia cómo el resultado hace lo que la solicitud o sus criterios ya piden. Se queda en la etapa actual.
  2. **Tarea añadida**: comportamiento nuevo que sirve al mismo resultado. Con plan guardado, entra como tarea añadida según `plan-revision.md`. Sin plan, sigue en `flow-build` solo mientras siga siendo un cambio pequeño y entendido; pasa a `flow-plan` cuando el cambio añade otro ticket, otro repositorio, un contrato o datos guardados.
  3. **Cambio nuevo**: cambia el resultado que nombra la solicitud (como hoy `plan-revision.md:12`). El usuario elige si el trabajo actual sigue hasta su punto de cierre o se pausa con `flow-close`; lo nuevo se planea aparte.
  4. **Tema aparte**: algo que no hace falta para este resultado, o que pertenece a otro trabajo. Recibe un destino según la regla de temas abiertos (ticket, plan, vigilado o descartado) y el trabajo sigue.
- **Relación con las etiquetas de la guía global, sin solaparse:** «aclara» = corrección; «relacionado» = tarea añadida o cambio nuevo; «aparte» = tema aparte.
- **Cuándo preguntar:** clasificar y decir la clase en una línea antes de actuar; preguntar solo cuando dos clases son razonables y cambian lo que se entrega (típicamente tarea añadida frente a cambio nuevo), con la recomendada primero, siguiendo `questions.md`.
- **Delegación desatendida:** una corrección dentro del objetivo y el alcance la mantiene; una tarea añadida o un cambio nuevo la termina, como fija `unattended-delegation/SKILL.md` (puntero, sin repetir sus listas).
- **Dónde vuelve el trabajo**, una línea por flujo, como puntero al hogar de cada regla y no como paráfrasis: research (la parada de la ruta iterativa de `flow-research`); plan (`plan-revision.md`); build (tarea añadida o paso a plan); verify (ronda de corrección o tarea añadida verificada según `plan-revision.md`); después de integrar (cambio nuevo; un defecto va como PR de seguimiento según `delivery-decisions.md`); close (la respuesta a la pregunta de cierre se clasifica igual); deploy (cambiar el conjunto de tickets termina la corrida en curso en un punto seguro y la siguiente empieza con el conjunto nuevo; esta regla vive aquí porque `flow-deploy` no tiene otra).
- **Tercera ronda (D6-B):** cuando una tercera ronda de corrección toca el mismo aspecto, en la pausa de validación o en la primera vista, proponer como pregunta volver al plan (revisarlo o abrir un cambio nuevo) en vez de otra ronda; sin tope duro, el usuario decide.
- **Ejemplos** para fijar fronteras (estilo OpenSpec): «el botón debe decir Guardar» → corrección; «también exportar a CSV» en una pantalla de exportación → tarea añadida; «convirtamos esto en un sistema de temas» → cambio nuevo; «estaría bien un modo oscuro algún día» → tema aparte.

### Punteros

Cada puntero nombra la regla y cuándo leerla, sin repetir sus pasos (AGENTS.md):

- `global.md:21`: las tres etiquetas pasan a «clarifies (a correction: update the work in progress)», «opens a related one (an added task or a new change)» y «is separate (it waits for its destination)», y una frase: «Before acting on a topic that would change work in progress, classify it as [scope changes](skill:flow-plan/references/scope-changes.md) sets.» Así ninguna acción de la guía global contradice la referencia.
- `plan-revision.md:11-12`: sustituir las definiciones de corrección, tarea añadida y cambio nuevo por un puntero a la referencia; conservar las reglas de formato de la tarea añadida y del cierre o pausa.
- `flow-plan/SKILL.md:22`: junto a «When an answer returns the work…», una cláusula que remite a la referencia para clasificar un cambio pedido.
- `flow-build/SKILL.md:20`: junto a «a small understood change needs no plan» / la reconciliación, una cláusula corta (≤ 200 bytes) que remite a la referencia cuando el trabajo sin plan crece.
- `flow-verify/SKILL.md:55`: clasificar cada pedido antes de su ronda; para un candidato sin efecto visual, la ronda muestra el diff y las pruebas en vez de capturas; la propuesta de la tercera ronda remite a la referencia.
- `flow-walkthrough/SKILL.md` «First rendered view»: tras un rechazo, las correcciones se hacen como rondas dentro del build, que cuentan para la tercera ronda; `:52` pasa a «In each round with a visual effect, capture the affected views».
- `flow-close/SKILL.md`: una cláusula corta (≤ 200 bytes) donde trata un requisito o tema que llega después de integrar.
- `flow-deploy/SKILL.md` «Entry and exit»: una frase sobre cambios del conjunto de tickets durante una corrida.

### `global.md:36` (D8-B)

Queda como invariante mínima, porque debe actuar sin la skill cargada: las cuatro fuentes de activación (delegación explícita, trabajo programado declarado que coincide, silencio tras la pregunta de ruta con las respuestas de entrega y revisión, y el mensaje de que el usuario se va mientras corre trabajo autorizado); su vencimiento (al completarse, o cuando el usuario cambia el objetivo, el alcance o una respuesta en vigor, pide otro trabajo o dice que pare), que es un límite de autoridad (`context-delivery`, «Portable kernel and situational procedures»); que la duración de la tarea, el silencio antes de esas respuestas o un turno en segundo plano sin trabajo declarado nunca la activan; que nunca permite acciones irreversibles o destructivas salvo el gasto autorizado de antemano y no amplía los permisos del host; y el puntero a la skill para qué mensajes la mantienen y qué efectos autoriza. `global.md:20` sigue remitiendo a «its rule below».

**Ajuste de D8-B sobre `global.md:72`:** no se toca. Su frase sobre controles saltados ya es la invariante mínima que debe actuar para cualquier puerta sin `flow-verify` cargado, y AGENTS.md permite esa repetición; `flow-verify/SKILL.md:33` añade lo propio de verify.

### Diagrama y sitio

- README: subgrafo `Rules["General rules"]` cuyas entradas siguen los títulos de secciones de `global.md` (Language, Communication, Scope and authorization, Preservation, Secrets, Evidence, Verification gates, Proportionality, Project settings, Continuity, Specs and change records, Support folder); `flowchart TD` no fija que quede al lado, se acepta su posición. Flechas de vuelta punteadas con los IDs actuales: `Verify -.->|corrections| Build`, `Build -.->|scope grows| Plan`, `Close -.->|close question| Build` y `Close -.-> Plan`, `Deploy -.->|defect| Build`. Nodo `Change["Change requested"]` conectado solo a los puntos de reentrada (Plan y Build). El enlace a `scope-changes.md` va en prosa debajo del diagrama, porque una etiqueta de mermaid no enlaza.
- Renderizado antes del push (decisión menor tomada sin preguntar): en la parada de validación genero en scratch una página HTML que carga una versión fija de mermaid desde un CDN y te la abro; antes, aplico «Check a dependency before installing» de `security-boundaries.md`. Si esa comprobación no lo permite, valido el código fuente antes del push y el diagrama renderizado en la vista del PR, con el merge condicionado a esa revisión.
- Sitio: `site/src/content/docs/flows/overview.md` con título y descripción, una tabla por flujo (dónde se itera, dónde se pide un cambio, a dónde vuelve) en lenguaje para personas y enlace a la referencia en `master`; entrada `{ label: "Changes and loops", slug: "flows/overview" }` justo después de `{ label: "Overview", slug: "skills" }` (`site/astro.config.mjs:74`). Actualizar las páginas que describen temas abiertos, cambios de plan, la corrida de deploy o la pregunta de cierre (`flows/research.md`, `flows/plan.md`, `flows/build.md`, `flows/verify.md`, `flows/close.md`, `flows/deploy.md`) y enlazar el panorama desde «The main flow» de `site/src/content/docs/index.mdx` y «Main flows» de `site/src/content/docs/skills.md`.
- `llms.txt`: añadir «a flows overview» a la línea que enumera el contenido del sitio.

## Contratos

Sin contratos de código ni datos. El contrato afectado es el texto de las reglas distribuidas y el requisito `documentation-site` (delta en `specs/documentation-site/spec.md`).
