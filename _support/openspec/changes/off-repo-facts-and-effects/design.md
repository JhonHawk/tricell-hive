# Design

## Approaches

Para medir los datos externos se presentaron tres enfoques:

1. **En el plan, con un hijo por sistema externo en paralelo y contraste en el hilo principal (elegido, D1 A).** Mide lo que el diseño elegido necesita y da un segundo control sobre cada dato. Cuesta de uno a tres hijos más en un plan que sale del repositorio.
2. **En research, antes del plan.** Reutiliza los hijos paralelos de `flow-research` paso 4, pero los datos que dependen del diseño se medirían tarde y un plan iniciado directamente se lo salta.
3. **En el plan, sin hijos.** Lo más barato; nadie contrasta y las lecturas quedan en el contexto del hilo principal.

## Decisions

- D1 A, D2 A, D3 A y D6 A, como las registra `proposal.md`.
- Llamadas menores tomadas sin preguntar:
  - Un solo archivo nuevo, `content/skills/flow-plan/references/outside-the-repository.md`, es el hogar de los datos externos y de la ronda de efectos, porque las dos reglas se leen en la misma condición y `flow-plan/SKILL.md` solo gana punteros.
  - No se añade ni se cambia ningún rol: `hive-read-state` cubre Git remoto, CI, tracker y despliegue; `hive-research` cubre lo demás (capacidad de una cuenta o licencia, permisos, límites y la documentación del proveedor). Lo que el contraste necesita de ellos lo pide el brief.
  - `check_change.py` no cambia.
  - La ronda de efectos no aplica a una tarea operativa, que sigue preguntando cada efecto crítico (`global.md` línea 68).
- Supuesto sin verificar en todos los hosts: un hijo con perfil de solo observación puede correr los CLI de lectura del proveedor. `hive-read-state` ya lee así el build servido; donde un host lo impida, el hijo lo reporta y el dato queda como no medido.

## Proposed shape

Las frases entre comillas invertidas en inglés son literales que el texto final contiene y que las comprobaciones de `tasks.md` buscan.

### `references/outside-the-repository.md` (nuevo)

Cabecera: `flow-plan` la lee después de elegir el enfoque (paso 2), cuando el enfoque elegido lee o cambia un sistema externo. Un sistema externo es una cuenta o un servicio de un proveedor al que se llega con una credencial o un CLI propios, como una cuenta en la nube, una organización de GitHub o un servicio de terceros. No cuentan el remoto, el CI y el tracker de este repositorio en lo que la entrega ya usa: la rama de trabajo, su pull request, sus checks y los movimientos del ticket. Sí cuentan cuando el plan depende de otra cosa en ellos, como el contenido de otra rama, las protecciones de una rama o los permisos de un entorno.

**External facts.** Antes de escribir las tareas:

1. Listar cada `external fact` del que depende el enfoque: el estado que el plan asume, como el contenido de una rama remota, la capacidad de una cuenta o licencia, un permiso, un límite, la disponibilidad de un servicio o un valor de configuración de un entorno.
2. Medir cada uno con una lectura de solo lectura. Una lectura que gasta dinero o cuota de pago sigue la regla de confirmación de «Scope and authorization».
3. Con un solo sistema externo, mide el hilo principal. Con dos o más, lanzar en el mismo paso `one read-only child per external system`: `hive-read-state` para Git remoto, CI, tracker o despliegue, y `hive-research` para lo demás. Cada brief nombra los datos y sus destinos y pide, por dato, el comando exacto de solo lectura y un extracto literal de su salida.
4. Contraste: el hilo principal acepta un dato solo con ese comando y ese extracto; un dato de capacidad, límite o permiso se compara además con la documentación vigente del proveedor, leída como fija «Evidence». Un dato en el que lo observado y lo documentado no coinciden, o que ninguna lectura de solo lectura puede mostrar, se escribe como `not measured`, con la tarea que lo encuentra primero y lo que pasa si resulta falso.
5. Registro: una línea por dato en «Verified context» de `proposal.md` (dato · comando o fuente · resultado · fecha), o en el resumen de un plan que vive en la conversación. Los no medidos van además en el resumen del plan.

**Effects round.** En el paso 4 de `flow-plan`, para cada plan:

1. Listar los efectos sobre sistemas externos que las tareas prevén, sin los que la respuesta de entrega o una regla ya autoriza (commits, push, pull request, merge y movimientos del ticket), y agruparlos en clases. Una `class of effect` es un tipo de acción sobre un destino nombrado, como los cambios de políticas de acceso en una cuenta o los merges con privilegio de administrador en un repositorio, con su techo de gasto (cero cuando no gasta) y las tareas que la usan.
2. Preguntar una decisión por clase, con el formato de «Questions and routes»: en la ronda de entrega y revisión cuando esa ronda se pregunta, y en una ronda propia cuando sus respuestas ya están en vigor. Cuando el destino es de producción, la situación de la pregunta lo dice.
3. La respuesta es la instrucción del usuario para cada efecto de esa clase que tomen las tareas nombradas de este trabajo, dentro del techo; no cubre otro trabajo ni otro plan.
4. Ninguna clase cubre, y conserva su pregunta propia: un despliegue a producción (`flow-deploy` paso 4); un efecto que detiene o interrumpe un componente que atiende usuarios; un efecto irreversible o difícil de revertir; destruir o reemplazar infraestructura; cambiar o borrar datos que la tarea no creó en una base no local; y crear cuentas o registros en un entorno compartido («Test accounts»).
5. Antes de cada efecto cubierto, registrar su destino, el cambio y cómo revertirlo en el registro de la tarea; el reporte final los lista. Un efecto que ninguna clase respondida cubre se pregunta cuando llega, como hoy.
6. Registro: una línea por clase en `## Effects` de `proposal.md` (clase · destino · techo · tareas · respuesta), o en el resumen de un plan que vive en la conversación. Un plan sin efectos sobre sistemas externos no añade preguntas. Una tarea operativa no tiene esta ronda.

### Punteros y cambios en los demás archivos

- `flow-plan/SKILL.md`:
  - paso 2: una frase que, tras elegir el enfoque, manda leer la referencia cuando el enfoque lee o cambia un sistema externo y medir cada `external fact`;
  - paso 3: la tarea de promoción que el plan añade usa la respuesta de la clase que cubre su entorno no productivo, cuando la hay, y pregunta producción como fija `flow-deploy`;
  - paso 4: una frase que añade una decisión por `class of effect`, como fija la referencia;
  - «Where the plan lives» y «Change folder»: las listas de lo que llevan el resumen y `proposal.md` nombran los datos externos y las clases de efecto.
- `global.md` línea 14: cada efecto crítico lleva su pregunta propia «unless the user asks for it on a named target, or, for a critical effect that only spends or is reversible, the user answered its class in a plan's `effects round`, as flow-plan sets, or, in unattended work, it is spend within the `unforeseen-spend amount`, as unattended-delegation sets». Los demás efectos críticos conservan su pregunta propia en esa misma frase. Línea 98: `flow-plan` pregunta en su paso 4 entrega y revisión, y la `effects round` de cada plan. Las reglas completas viven solo en la referencia y en `unattended-delegation`.
- `delegation.md`: una tercera razón, después de independencia y costo: «Cross-check: the external facts of a plan», con `one read-only child per external system` cuando el plan tiene dos o más, como fija `flow-plan`; la frase «when neither applies» pasa a «when none applies», y esa recolección entra en la lista de hijos de solo lectura que corren a la vez.
- `flow-deploy/SKILL.md` paso 4: en la tarea de promoción de un plan, un entorno no productivo o un efecto que una clase respondida en la `effects round` de ese plan cubre no se vuelve a preguntar; la respuesta se lee de la sección `## Effects` de su carpeta de cambio o del resumen del plan. Una corrida que no viene de ese plan, o que lleva otros tickets, pregunta como hoy. Producción se pregunta siempre aparte.
- Plantilla `proposal.md`: el comentario de «Verified context» muestra la línea de un `external fact`; sección nueva `## Effects` con la línea de una `class of effect` o «None»; su comentario dice que este encabezado queda en inglés en cualquier idioma, como `Observed in:`, porque `flow-deploy` lo busca.
- `plan-handoff.md`: el punto 6 lleva también cada `external fact` no medido; el punto 9 nombra cada `class of effect` respondida y los efectos que quedan sin autorizar.
- `unattended-delegation/SKILL.md`:
  - paso 2: para toda fuente de activación, el mensaje de activación dice el `unforeseen-spend amount` que el usuario dio, o lo pregunta en ese mensaje. Esa pregunta no es una de las respuestas abiertas del paso 4 y no deja el trabajo en `hold`; sin respuesta, el monto es cero.
  - paso 4: las respuestas en vigor incluyen las clases que el plan respondió; su techo cuenta como gasto autorizado para ese destino. Una clase sin responder no añade autorización y no es una respuesta abierta que deje el trabajo en `hold`: el paso 6 decide cada uno de sus efectos como hoy, de modo que un efecto reversible que no gasta sigue permitido por su primera viñeta y su gasto cae en el `unforeseen spend`.
  - paso 6, viñeta del efecto crítico que gasta: sigue exigiendo destino y techo exactos, y nombra su única excepción, el `unforeseen spend`: un gasto que ninguna respuesta cubre se toma solo cuando hay un monto en vigor, el cambio de recurso o de configuración se puede deshacer con un comando registrado antes de tomarlo, y el total de imprevistos, contado con el precio del proveedor hasta el vencimiento del encargo o con el máximo del efecto cuando lo tiene, no pasa del monto; cuando no se puede leer ni el precio ni un máximo, se omite y se reporta. El dinero ya gastado cuenta contra el monto y no hace irreversible el efecto. Se registra antes y se lista en el reporte final con cómo revertirlo. En cualquier otro caso se omite y se reporta, como hoy. La lista de lo prohibido no cambia.

## Impact

Lugares que enuncian o dependen de las tres conductas, del barrido del 2026-11-09 y de la revisión del plan:

- Enuncian: `content/guidance/global.md` líneas 14, 22, 68 y 98; `flow-plan/SKILL.md` pasos 1 a 4, «Where the plan lives», «Change folder» y «Prerequisites»; `flow-build/SKILL.md` paso 1; `delegation.md`; `flow-deploy/SKILL.md` paso 4; `deployment-checks.md` (efectos del recorrido funcional); `unattended-delegation/SKILL.md` pasos 2, 4 y 6.
- Dependen: plantillas `proposal.md` y `tasks.md`; `plan-handoff.md`; `hive-review-plan.md`; `hive-read-state.md` y `hive-research.md`; `flow-research/SKILL.md` y `flow-status/SKILL.md` (mismo patrón de hijos de recolección); páginas del sitio `flows/plan.md`, `flows/deploy.md`, `flows/build.md`, `flows/overview.md`, `skills.md`, `index.mdx`; la rúbrica de revisión de sesiones (filas del 2026-11-08, que ya miden estas conductas).
- No cambian y siguen coherentes: `global.md` línea 22 (el monto es gasto autorizado de antemano dentro de un techo) y línea 68 (la tarea operativa pregunta cada efecto crítico y no tiene ronda de efectos); `flow-build` paso 1 (los prerrequisitos son lo que alguien provee; los datos externos son lo que el plan mide); `deployment-checks.md`; los dos roles; `README.md` y `llms.txt`, que no nombran estas conductas.
- Pruebas: ninguna afirma estos textos. `tests/content/skill_size_test.go` limita cada `SKILL.md` y cada referencia a 16,000 bytes; `tests/content/skill_links_test.go` exige que los enlaces resuelvan; `tests/content/spec_deltas_test.go` comprueba la delta al archivar.
- Tamaño esperado: `global.md` crece unos 350 bytes; `flow-plan/SKILL.md` unos 900; `unattended-delegation/SKILL.md` unos 1,100; la referencia nueva queda por debajo de 6,000.
- No toca producción ni datos.

## Recovery

Revertir el pull request restaura las reglas anteriores; después se repite `hive update --source` para desplegar el estado revertido. Nada de este cambio es irreversible.
