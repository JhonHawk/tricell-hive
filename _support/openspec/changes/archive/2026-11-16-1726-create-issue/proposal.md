# create-issue — Skill que crea, desglosa y valida tickets, y hogar único de ese procedimiento

| Campo | Valor |
| --- | --- |
| Estado | Completado · PR #482, b315ff00 |
| Tracker | ninguno |
| Git | JhonHawk/tricell-hive · base `development` en `b8ee7913` · rama propuesta `feat/create-issue` |

## Decisiones

2026-11-16

- **D13 A** — ¿Qué se construye primero de las tres skills por entregable? `create-issue`, con un plan propio; después `create-prd` y `create-design`. La skill decide el nivel del ticket, hasta un ticket padre con hijos, y es la que se sigue cuando en build surge la necesidad de un ticket.
- **D9 A** — ¿Qué pasa si al crear un ticket queda una regla de negocio sin decidir? Se crea marcado como bloqueado, con la pregunta y a quién espera; no cuenta como listo.
- **D11 C** — ¿Quién opera la skill? Sirve desde el inicio a una sesión con repositorio y a una sin él.
- **D12 A** — ¿Cuándo corre un revisor independiente? En tickets nunca: la lista fija se comprueba en el mismo hilo.
- **D14 A** — ¿La skill valida un ticket que ya existe? Sí; la prueba de «listo» vive en la skill y `flow-research` la enlaza.
- **D16** — ¿Qué hace al validar un ticket del que no se es dueño? Genera un comentario con lo que falta; no lo edita.
- **D17 A** — ¿Qué tipos de ticket soporta? Historia, bug, tarea y spike; la deuda técnica va como tarea con la línea de qué cuesta hoy.
- **D18 A** — ¿Cuánta solución lleva un ticket? Una pista corta, aparte y no vinculante; si escribirla exige decidir, el ticket dice que requiere un plan.
- **D19 A** — ¿Cuándo propone partir en hijos? Con más de un cambio que se integra por separado, más de un repositorio, o más de 7 criterios de aceptación.
- **D20 A** — ¿Cómo se representa una épica? Como un ticket padre con hijos en GitHub, Linear y Jira; no se crean proyectos.
- **D21 A** — ¿El ticket declara que lo escribió un agente? No, salvo que las reglas de contribución del proyecto lo pidan.
- **D22 A** — ¿Cómo llegan los demás flujos a la skill? Como procedimiento con punteros: el cuándo proponer y la autorización siguen en la guía global; el cómo se muda a la skill y los flujos apuntan a ella.
- **D23 A** — ¿Cómo se marca un ticket bloqueado? Con lo que el tracker ya tiene, sin crear etiquetas ni estados.
- **D24 A** — ¿Se valida con un primer uso real? Sí: tras integrar y refrescar la instalación, la skill crea en este repositorio un ticket padre con dos hijos (`create-prd` y `create-design`), con una aprobación del lote.
- **D25 B** — ¿Con qué modo se entrega? Automático: revisión, push, pull request e integración sin detenerse.
- **D26 A** — ¿Qué revisión de código corre antes del push? La nativa de Claude Code, `/code-review`.
- **D27 A** — ¿Se aprueba que T7 cree tres tickets y sus enlaces en `JhonHawk/tricell-hive`? Sí; la skill pide además la aprobación del lote en ese momento.
- **D28 A** — ¿El plan abarca también mover, comentar y agregar evidencias a un ticket? No: la skill crea y valida; mover y comentar siguen en el flujo que los provoca. Agregar evidencias a un ticket no existe hoy y queda para el trabajo de higiene del tracker.
- **D29 A** — ¿Se confirma el borrador cuando la respuesta del usuario ya eligió «ticket» dentro de un flujo? No: ahí se crea y se muestra el ticket creado con su enlace. Usada sola, o al desglosar, la skill muestra el borrador y espera la confirmación.

Decisiones menores tomadas sin preguntar, dichas al usuario el 2026-11-16:

- Sin supervisión se sigue pudiendo abrir un ticket suelto por un hallazgo, como hoy; un desglose en lote no se crea y queda reportado como pendiente.
- El dueño de un ticket es su autor según la cuenta autenticada en el tracker. El autor ve el cambio propuesto y se edita tras su confirmación; quien no puede escribir en el tracker recibe el reporte.
- El ticket se escribe en el idioma de la sesión, salvo que el proyecto indique otro.
- Tope de tres preguntas por ronda.
- La skill tiene página propia en el sitio.
- La regla de que los tickets nuevos se crean solo en el tracker principal se queda en «Project settings» de la guía global; la skill la enlaza.
- `flow-build`, `flow-verify` y `test-audit` no se editan: llegan a la skill a través de la regla de tickets de la guía global, que ya citan.
- Los hijos de un desglose que toca varios repositorios se crean todos en el tracker principal y cada uno nombra su repositorio. Varios tickets de hallazgos elegidos uno a uno no son un lote.

## Business rules

None: the work adopts no business rule beyond what its ticket and documentation state.

## Request

Del usuario, 2026-11-16:

> «tal vez lo que necesitemos es una sola skill por el tipo de entrega y no un flow, pooooor ejemplo - create-spec - create-issue - create-figma-design (Se ajusta). Ya que son interaciones mucho más cortas que un desarrollo, aunque importante, cómo iterar, cómo evitamos que se cree un ticket, o un diseño con gaps de analisis o diseño»

> «consideremos que la skill debe ser capaz de razonar el nivel de issue hasta incluso poder crear una epica con child issues, y talvez sea esta skill la que se lance cuando en build surga la necesidad de crear tickets/issues»

> «Qué pasa si la skill lo ejecuta un dev sin permiso real para editar lo que esta mal» · «Qué pasa con el tema de que el issue esta tan cargado de canibalice parte de lo que hace la skill de plan?»

> «Cuando no se es el dueño del contenido, talvez lo que corresponda es generar el comentario correspondiente»

Sin ticket de origen.

## Objective

Hoy Hive no tiene un procedimiento para crear un ticket. La guía global dice cuándo proponer uno, y que antes hay que buscar duplicados y relacionarlo con su origen; `flow-research`, `flow-close` y la auditoría de código repiten partes. Nada define qué contiene un ticket ni revisa que esté completo al crearlo: la prueba de «listo» solo corre cuando `flow-research` lee un ticket elegido.

Con el cambio, la skill `create-issue` es el único lugar donde vive ese procedimiento. Crea un ticket de uno de cuatro tipos, desglosa un trabajo en un ticket padre con hijos y valida un ticket existente. La guía global conserva cuándo proponer un ticket y que crearlo necesita la respuesta del usuario, y remite a la skill para el cómo.

## Scope and acceptance

Dentro: la skill y sus referencias; los punteros en la guía global, `flow-research` (con su auditoría de código), `flow-close`, `flow-plan` y `unattended-delegation`; el delta de `workflow-lifecycle`; el sitio, `llms.txt`, el registro de cambios y la lista de hogares canónicos; el refresco de la instalación local y un primer uso real.

Fuera: `create-prd` y `create-design`; reorganizar tickets existentes en lote (la higiene del tracker); mover un ticket de estado, el comentario de pendientes, la escritura de reglas de negocio al ticket y agregarle evidencias; crear etiquetas, estados o proyectos en un tracker; un revisor independiente para tickets; pruebas de comportamiento en hosts (los pilotos siguen en pausa).

- AC1 — El catálogo distribuye la skill `create-issue`, cuya descripción nombra sus tres usos: crear un ticket, desglosar un trabajo en un ticket padre con hijos y validar un ticket que ya existe. Usada sola, termina en el ticket y solo ofrece el siguiente paso; seguida desde un flujo, devuelve el control al paso siguiente de ese flujo · false today because `content/skills/create-issue/` no existe · Observed in: el `create-issue/SKILL.md` instalado en cada host registrado tras `hive update`
- AC2 — La skill define cuatro tipos de ticket con sus campos, una lista fija de ocho pruebas de «listo» en el mismo hilo, y hasta tres preguntas por ronda · false today because `content/skills/create-issue/references/ticket-types.md` no existe · Observed in: el `create-issue/references/ticket-types.md` instalado tras `hive update` **Retired:** partido en AC13, AC14 y AC15 por la revisión del plan, que pidió un comportamiento por criterio
- AC3 — Un ticket con una regla de negocio sin decidir se crea marcado como bloqueado, eligiendo el primer medio que exista: la relación nativa de «bloqueado por» cuando lo bloquea otro ticket; la etiqueta o el estado de bloqueo que el proyecto ya tenga; si no hay ninguno, la primera línea del ticket con la pregunta y a quién espera. La skill nunca crea etiquetas ni estados · false today because ninguna regla de `content/` describe crear un ticket bloqueado y `create-issue/references/trackers.md` no existe · Observed in: el `create-issue/references/trackers.md` instalado tras `hive update`
- AC4 — La skill propone desglosar por tres señales, muestra el lote completo, pide una aprobación y escribe de uno en uno · false today because `content/` no contiene las frases "more than 7 acceptance criteria" ni "one approval for the whole set" · Observed in: el `create-issue/references/breakdown.md` instalado tras `hive update` **Retired:** partido en AC16, AC17 y AC18 por la revisión del plan, que pidió un comportamiento por criterio
- AC5 — Un ticket padre con hijos se crea igual en los tres trackers: el tipo épica de Jira, un issue padre con sub-issues en GitHub (con tipo épica solo si la organización lo define) y un issue padre con sub-issues en Linear, sin crear un proyecto. Para otro tracker, la skill usa su CLI o herramienta propia y dice qué no pudo representar · false today because ninguna regla de `content/` describe cómo se representa un ticket padre en cada tracker · Observed in: el `create-issue/references/trackers.md` instalado tras `hive update`
- AC6 — Al validar un ticket que ya existe, la skill comprueba autoría y permiso, y comenta, edita o entrega un reporte según el caso · false today because ninguna regla dice qué hacer con un ticket ajeno · Observed in: el `create-issue/references/existing-ticket.md` instalado tras `hive update` **Retired:** partido en AC19 y AC20 por la revisión del plan, que pidió un comportamiento por criterio
- AC7 — Un ticket lleva como mucho una pista de solución, en una sección aparte marcada como no vinculante, con hasta cinco ubicaciones y una frase de dirección. No lleva lista de pasos de implementación, elección entre enfoques, definición de contratos ni cambios archivo por archivo; cuando escribir la pista exige una de esas decisiones, el ticket dice, en su idioma, que requiere un plan, y no la incluye. `flow-plan` trata la pista de un ticket como un dato por verificar, no como una decisión · false today because `rg -n -i "hint" content/skills/flow-plan/SKILL.md` no encuentra nada y ninguna regla limita la solución que lleva un ticket · Observed in: los `create-issue/references/ticket-types.md` y `flow-plan/SKILL.md` instalados tras `hive update`
- AC8 — El procedimiento de crear un ticket tiene un solo hogar. La guía global conserva cuándo proponer un ticket y que crearlo necesita la respuesta del usuario, y remite a `create-issue` para el cómo. Los pasos de buscar un ticket existente y de relacionar el nuevo con su origen se enuncian solo en la skill; la guía global, `flow-research`, su auditoría de código y `flow-close` apuntan a ella · false today because `content/guidance/global.md` contiene "search every tracker" y "blocked-by or blocking when one waits on the other", `flow-research/SKILL.md` contiene "Search each declared tracker for an existing ticket before proposing one" y `flow-close/SKILL.md` contiene "after searching each declared tracker for an existing ticket" · Observed in: la guía global, `flow-research`, `code-audit.md`, `flow-close` y la skill instalados tras `hive update`; en el repositorio, `rg -F "search every tracker" content/` solo encuentra archivos bajo `content/skills/create-issue/`
- AC9 — La prueba de «listo» vive en la skill: «Assess a chosen ticket» de `flow-research` conserva impacto, complejidad y dependencias, y remite a la lista de la skill para juzgar la preparación · false today because `flow-research/SKILL.md` enuncia la prueba él mismo ("A ticket whose criteria are missing, that names more than one change, or that the specs contradict is not ready") · Observed in: el `flow-research/SKILL.md` instalado tras `hive update`
- AC10 — La skill no añade al ticket una declaración de que lo escribió un agente, salvo que las reglas de contribución del repositorio destino la pidan; cuando no puede saberlo, lo pregunta · false today because ninguna regla de `content/` trata la autoría de un ticket · Observed in: el `create-issue/SKILL.md` instalado tras `hive update`
- AC11 — El sitio tiene una página de `create-issue` enlazada desde la barra lateral y la lista de skills; las páginas de research y de close describen que el ticket se crea siguiendo la skill; `llms.txt` lista la skill y el registro de cambios la anuncia; la lista de hogares canónicos nombra la skill como hogar del procedimiento de tickets · false today because `site/src/content/docs/flows/create-issue.md` no existe y `rg -n "create-issue" llms.txt CHANGELOG.md site/ _support/docs/architecture/` no encuentra nada · Observed in: el sitio construido localmente (`astro build`) y esos archivos
- AC12 — En una sesión real, tras integrar el cambio y refrescar la instalación, la skill instalada crea en `JhonHawk/tricell-hive` un ticket padre con dos hijos (`create-prd` y `create-design`) enlazados como sub-issues, después de una sola aprobación del lote por el usuario; la relectura de los tres coincide con lo aprobado · false today because la búsqueda del 2026-11-16 en ese tracker no encuentra tickets de `create-prd` ni `create-design` · Observed in: el tracker de `JhonHawk/tricell-hive`
- AC13 — La skill define cuatro tipos de ticket (historia, bug, tarea y spike), cada uno con sus secciones; la deuda técnica se crea como tarea con la línea de qué cuesta hoy · false today because ningún archivo de `content/` define secciones por tipo de ticket (`create-issue/references/ticket-types.md` no existe) · Observed in: el `create-issue/references/ticket-types.md` instalado tras `hive update` **Replaces:** AC2
- AC14 — La skill comprueba cada ticket contra una lista fija de ocho pruebas de «listo», en el mismo hilo y sin lanzar un hijo revisor · false today because ningún archivo de `content/` contiene una lista de «listo» que se aplique al crear un ticket · Observed in: los `create-issue/SKILL.md` y `create-issue/references/ticket-types.md` instalados tras `hive update` **Replaces:** AC2
- AC15 — La skill hace como máximo tres preguntas por ronda, cada una con respuesta recomendada, y escribe en el ticket como supuesto lo que no pregunta · false today because `rg -F "at most three questions" content/` no encuentra nada · Observed in: el `create-issue/SKILL.md` instalado tras `hive update` **Replaces:** AC2
- AC16 — La skill propone desglosar cuando el trabajo nombra más de un cambio que se integra por separado, toca más de un repositorio o tiene más de 7 criterios de aceptación; una parte que no se integra y acepta sola no se propone como hijo · false today because `rg -F "more than 7 acceptance criteria" content/` no encuentra nada · Observed in: el `create-issue/references/breakdown.md` instalado tras `hive update` **Replaces:** AC4
- AC17 — Al desglosar, la skill muestra el lote completo (padre, hijos y dependencias) y pide una sola aprobación antes de la primera escritura. Escribe los tickets de uno en uno, todos en el tracker principal y cada hijo nombrando su repositorio, registra cada uno al crearlo, relee cada uno, y ante un fallo se detiene y reporta lo ya creado · false today because `rg -F "one approval for the whole set" content/` no encuentra nada · Observed in: el `create-issue/references/breakdown.md` instalado tras `hive update` **Replaces:** AC4
- AC18 — Bajo delegación sin supervisión, la skill crea el ticket suelto de un hallazgo sin pedir confirmación, y nunca crea un padre con hijos: reporta el lote propuesto como pendiente · false today because `unattended-delegation/SKILL.md` permite "opening a ticket for a finding" sin distinguir un ticket suelto de un lote · Observed in: los `create-issue/references/breakdown.md` y `unattended-delegation/SKILL.md` instalados tras `hive update` **Replaces:** AC4
- AC19 — Al validar un ticket del que la cuenta autenticada no es autora, la skill no lo edita: muestra un comentario con lo que falta y las preguntas con respuesta recomendada, y lo publica tras la confirmación del usuario. Una regla de negocio sin decidir sale siempre como pregunta al dueño · false today because «Assess a chosen ticket» de `flow-research` solo reporta la preparación de un ticket y ninguna regla dice qué hacer con un ticket ajeno · Observed in: el `create-issue/references/existing-ticket.md` instalado tras `hive update` **Replaces:** AC6
- AC20 — Antes de escribir en un ticket que ya existe, la skill comprueba si la cuenta autenticada es su autora y qué permiso tiene. Si es la autora, muestra el cambio y edita tras la confirmación; si no puede escribir en el tracker, entrega el reporte al usuario · false today because ninguna regla de `content/` pide comprobar autoría o permiso antes de escribir en un ticket · Observed in: el `create-issue/references/existing-ticket.md` instalado tras `hive update` **Replaces:** AC6
- AC21 — En una sesión sin repositorio, donde ningún `## Hive` declara un tracker, la skill pregunta en qué tracker y repositorio crear antes de buscar o escribir, omite la pista de solución y lo dice en el ticket · false today because ninguna regla de `content/` trata la creación de un ticket sin repositorio · Observed in: el `create-issue/SKILL.md` instalado tras `hive update`
- AC22 — Usada sola, o al desglosar, la skill muestra el borrador o el lote y escribe tras la confirmación del usuario. Cuando la respuesta del usuario ya eligió crear el ticket dentro de un flujo, lo crea sin otra confirmación y muestra el ticket creado con su enlace · false today because ninguna regla de `content/` distingue esos dos casos y `flow-close` solo dice "A ticket answer opens it" · Observed in: el `create-issue/SKILL.md` instalado tras `hive update`

## Prerequisites

None: the build needs nothing from outside the repository.

## Verified context

Leído el 2026-11-16 sobre `development` en `b8ee7913`:

- `content/guidance/global.md` (52.397 bytes): línea 95, crear un ticket es una acción suelta que nunca pasa por el comando de un flujo; línea 149, la regla de tickets (cuándo proponer, buscar duplicados, y cuatro formas de relacionarlo con su origen: bloqueo, sub-issue para una parte aplazada, relacionado, y un comentario con la URL cuando el origen está en otro tracker); línea 116, los tickets nuevos solo en el tracker principal.
- `content/skills/flow-research/SKILL.md` (10.076 bytes): líneas 8, 20, 34 y 36–38 («Assess a chosen ticket»). `references/code-audit.md` (13.263 bytes): línea 71, el cuerpo de un ticket de hallazgo.
- `content/skills/flow-close/SKILL.md` (15.887 bytes, a 113 del límite de 16.000): líneas 79–85, la tabla de hallazgos, su búsqueda de duplicados, y que tras responder cada hallazgo el cierre termina con su pie.
- `content/skills/unattended-delegation/SKILL.md` línea 14: abrir un ticket por un hallazgo está permitido sin supervisión.
- `content/skills/flow-plan/SKILL.md` (13.995 bytes) no menciona la pista de un ticket.
- `flow-build` no dice nada sobre crear tickets; `flow-verify` línea 16 y `test-audit` línea 47 remiten a la regla de tickets de la guía global.
- El instalador recoge las skills por patrón (`tooling/management/plan.go:229`) y valida los enlaces entre instrucciones (`tooling/management/references.go`); no hay lista que registrar ni test que cuente el catálogo. Los tests de contenido exigen `name` igual a la carpeta, descripción de una línea de hasta 1.024 caracteres, y 16.000 bytes por `SKILL.md` y por referencia.
- `_support/openspec/specs/workflow-lifecycle/spec.md`: requisito «Tickets follow the tracker their key declares», con el escenario «New ticket», que hoy cubre proponer y crear.
- `_support/workspace/maintainer.local` existe: integrar en `development` pide refrescar la instalación local, y las sesiones abiertas cargan la guía nueva solo tras reiniciarse.
- Investigación de estándares del 2026-11-16 (campos por tipo, lista de «listo», señales de desglose, jerarquía por tracker): resumida en `design.md`.

Hechos externos:

- El usuario tiene permiso de administrador en `JhonHawk/tricell-hive`, que es público y tiene los issues activados · `gh repo view JhonHawk/tricell-hive --json owner,isPrivate,hasIssuesEnabled,viewerPermission,visibility` · `{"issues":true,"owner":"JhonHawk","perm":"ADMIN","private":false,"vis":"PUBLIC"}` · 2026-11-16
- El dueño del repositorio es una cuenta de usuario, no una organización, así que no hay tipos de issue: el ticket padre se crea sin tipo · `gh api users/JhonHawk --jq .type` · `User`; la documentación de GitHub sitúa los tipos de issue en la organización · 2026-11-16
- Los sub-issues están disponibles en ese repositorio · `gh api repos/JhonHawk/tricell-hive/issues/475/sub_issues --jq length` · `0` · 2026-11-16
- El repositorio no tiene etiqueta de bloqueo ni de falta de información · `gh label list --repo JhonHawk/tricell-hive --limit 60` · accessibility, bug, documentation, duplicate, enhancement, good first issue, help wanted, invalid, question, wontfix · 2026-11-16
- No existen tickets de `create-prd` ni `create-design` · `gh issue list --repo JhonHawk/tricell-hive --state all --search "create-issue OR create-prd OR create-design …"` · solo el #333, cerrado y de otro tema · 2026-11-16
- El `gh` instalado (2.104.0) crea un issue con padre y con dependencias · `gh issue create --help` · muestra `--parent`, `--blocked-by`, `--blocking` · 2026-11-16

## Effects

- Crear tres issues (un padre y dos hijos) y sus enlaces de sub-issue · target: `JhonHawk/tricell-hive` (público) · ceiling: 0 · tasks: T7 · answer: approved (D27 A)

## Delivery

Modo automático (D25 B): tras la verificación y la revisión, push de `feat/create-issue`, pull request contra `development` e integración, sin detenerse para validación. Revisión de código: la nativa de Claude Code, `/code-review`, sobre el candidato antes del push (D26 A). Tras integrar, el refresco de la instalación local lo autoriza el `AGENTS.md` del repositorio. Esta elección no autoriza un release, ni escribir en el tracker más allá de la clase aprobada en «Effects».

## Review status and progress

2026-11-16 — Plan escrito. Revisión del plan (`hive-review-plan`): nueve hallazgos, uno P1, todos aplicados. Cambió: la regla de cierre de la skill distingue el uso suelto del uso desde un flujo; el delta vuelve a exigir la búsqueda antes de proponer; se listan las cuatro formas de relación; se añadieron los criterios de la sesión sin repositorio y de la confirmación del borrador (D29 A); AC2, AC4 y AC6 se partieron en AC13 a AC20; se ajustaron las comprobaciones de cada tarea.
