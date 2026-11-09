# Design

## Approaches

Dónde se declara el archivo de diseño de un proyecto:

1. **En el `DESIGN.md` de la app (elegido, D1 A).** Su sección `## Hive` ya guarda datos propios de Hive, uno por app, y se lee antes de cualquier trabajo de pantalla. El cambio es solo de contenido. A cambio, `hive doctor` y el formulario del proyecto no lo muestran ni lo validan.
2. **En el `## Hive` de `AGENTS.md`, como clave `Design`.** Un solo lugar para los ajustes, validado por el gestor. Exige cambios en `tooling/cli/doctor_project.go`, el formulario, sus pruebas, dos specs, una página del sitio y las referencias de auditoría, y un formato de varios valores para un monorepo.

De dónde sale el permiso de escribir en el archivo de diseño:

1. **De la respuesta del usuario a una oferta (elegido, D5 C).** Research pregunta si el usuario quiere propuestas; su sí es la instrucción explícita que la guía global exige. La guía global no cambia y la pregunta ofrece algo, no pide permiso para una herramienta.
2. **De una frase nueva en la guía global.** Sin pregunta, pero la guía global crece y gana una segunda autorización permanente.
3. **De una pregunta de permiso por trabajo.** La que el usuario quiso evitar.

Dónde viven los ejemplos en el flujo: solo una forma es sensata. Research decide qué se quiere y el plan decide cómo; un ejemplo de pantalla responde al qué.

## Decisions

- **D1 A** — El archivo de diseño se declara en el `DESIGN.md` de la app, como dato.
- **D5 C** — El permiso sale de la respuesta a la oferta.
- **D6 A** — Un pedido de solo diseño se dibuja sin oferta, no se da por hecho sin aprobación y se cierra como tarea operacional.
- **D7 B** — Sin herramienta de diseño, la oferta se hace con bocetos de texto.
- **Sección con nombre, no un paso nuevo.** `flow-research` gana la sección «Examples for a screen», como «Assess a chosen ticket», y los pasos 0, 3 y 5 apuntan a ella. Renumerar los pasos rompería los punteros de `global.md` («flow-research step 5») y de la referencia de delegación («flow-research step 4»).
- **`global.md` no cambia.** La instrucción explícita de «Scope and authorization» es la respuesta del usuario a la oferta, o su pedido cuando el pedido es el diseño. La oferta no sustituye la pregunta de un efecto crítico: una escritura que gaste dinero o cuota de pago sigue pidiendo la suya. Un pedido de solo diseño es la acción que el paso 0 de research ya deja hacer y reportar; la sección solo añade que no se reporta hecho antes de la aprobación.
- **El viewer no cambia.** Research sigue sin publicar nada en él.

## Proposed shape

**Cuándo se ofrece.** El usuario pidió un cambio, no solo un análisis o la evaluación de un ticket, y lo que se va a construir incluye una pantalla nueva, o una que añade, quita o mueve una región, y ni el pedido ni una pantalla existente del proyecto fijan su composición. Un cambio que conserva la distribución, como un botón, un texto, un campo o una columna, no recibe la oferta. Con la delegación desatendida activa tampoco: el plan decide la UI como hoy.

**La oferta.** Una decisión, una vez por trabajo, con estas opciones:

| Opción | Cuándo aparece | Qué pasa |
| --- | --- | --- |
| Propuestas en el archivo de diseño | El `DESIGN.md` de la app declara `**Design file:**` o el usuario nombró el archivo en la conversación, y el host tiene una conexión que escribe en él | Frames en una página con el nombre del trabajo. La etiqueta o la descripción de la opción nombra esa escritura, y elegirla es la instrucción explícita |
| Propuestas como bocetos de texto | Siempre | Bocetos en el formato de «UI decisions» de `flow-plan` |
| Sin propuestas | Siempre | El plan decide la UI como hoy |

**Ejemplo.** Un dibujo de una pantalla en el ancho de escritorio, hecho con los tokens y los componentes del sistema de diseño del proyecto, con etiquetas reales y datos de volumen realista. Dos o tres cuando la composición está abierta; uno cuando está fijada. El ejemplo aprobado en un archivo de diseño añade el ancho de 440 px cuando la pantalla se sirve en móvil.

**Límite de escritura en el archivo.** El trabajo crea frames; cambia o borra solo los nodos que él creó. Cuando un pedido de solo diseño nombra una página o unos frames como destino, dibuja ahí, y esos frames se cambian pero nunca se borran. No cambia variables, estilos, componentes ni páginas existentes. Crear una librería, variables o componentes en el archivo es trabajo que el usuario pide aparte.

**Quién dibuja.** Los frames, un hijo `hive-design-ui` cuando el host da esa conexión a sus hijos, y si no el hilo principal; la referencia de delegación lo nombra entre el trabajo que un hijo hace por costo. Los bocetos de texto, el hilo principal. En el rol cambian los pasos 1, 3 y 4: la página del archivo de diseño que el encargo asigna cuenta entre las rutas que el encargo nombra, y la evidencia de un frame es su imagen y sus nodos, no las comprobaciones de navegador de una interfaz construida. Su párrafo final no se toca: una prueba lo fija palabra por palabra.

**Elección.** Una decisión con una opción por ejemplo; cada opción lleva la ubicación de su frame y un resumen de su composición, o el boceto en su vista previa. Con un solo ejemplo, las opciones son aprobarlo o corregirlo. Una corrección vuelve a dibujar el ejemplo elegido. Cuando el ejemplo elegido es un frame y la pantalla se sirve en móvil, se dibuja también a 440 px y el usuario lo aprueba antes de registrarlo. Con la delegación desatendida activa no se pregunta: se conservan todos los ejemplos, la elección se reporta como pendiente y el plan decide la UI sin ejemplo aprobado.

**Registro del ejemplo aprobado.** En el reporte de research y en una nota `examples.md` de `_support/workspace/YYYY-MM-DD-<work>/`, para que otra sesión lo encuentre. Para un frame: la herramienta, la ubicación del archivo, el nodo de cada frame y la ruta de su imagen local en la carpeta `images/` de ese trabajo. Para un boceto de texto: el boceto mismo. Los ejemplos no elegidos se borran. Un plan con carpeta de cambio copia el registro a su sección `UI`, y un trabajo sin ella a su archivo de checkpoint, para que una etapa que corre en otro checkout lo encuentre.

**Uso después de research.**

| Etapa | Qué hace con el ejemplo aprobado |
| --- | --- |
| `flow-plan` «UI decisions» | Es el boceto «después» de las regiones que muestra; registra la referencia o el boceto y no redibuja ni lanza el hijo de composición. Los pasos 1 a 3 (sistema de diseño, contrato de datos, componentes, estados y textos) se siguen escribiendo |
| `flow-build` primera vista | Se compara contra el ejemplo |
| `flow-verify` recorrido | Compara la captura de cada región con ejemplo y lista las diferencias en el mensaje de validación |
| `hive-review-ux` | Recibe la imagen del ejemplo en su encargo y reporta cada diferencia como hallazgo |

**Diferencia.** Una región, un componente, la posición de una acción o un orden distinto al del ejemplo aprobado. No lo es una variación de render entre la herramienta de diseño y la pantalla construida, ni lo que el plan registra como decidido distinto, como una variante de componente elegida en sus pasos 1 a 3, ni una corrección que el usuario pidió después de aprobar el ejemplo, que el build añade al registro. Un boceto de texto se compara como está registrado; un frame, por su imagen, y si falta se vuelve a leer desde su ubicación y su nodo. Una diferencia es un defecto de UI que el cambio introdujo: sigue la ronda de corrección que `flow-verify` ya fija, y el revisor la reporta con las prioridades que ya usa. Las que queden se listan en el mensaje de validación cuando la entrega se detiene, y en el reporte final cuando no.

La validación del usuario en verify no cambia: se pide cuando la entrega se detiene para validar.

**Pedido de solo diseño.** Es la acción que el usuario pidió: research la dibuja como fija esta sección, sin la oferta. No la reporta hecha antes de que el usuario apruebe el ejemplo. Tras la aprobación reporta el resultado y corre el cierre, como `flow-build` hace con una tarea operacional; el cierre se ocupa de la imagen local y de nombrar la página creada. Cuando el pedido nombra una página o unos frames como destino, se dibuja ahí y solo se cambian los frames que nombra y los nodos que el trabajo crea. Con la delegación desatendida activa, la aprobación queda sin responder y se reporta como pendiente. Cuando el pedido manda dibujar en un archivo y el host no puede escribir en él, es un bloqueo que se reporta con su salida.

## Impact

Lugares que enuncian o dependen del comportamiento; cada uno se edita o se confirma sin cambio:

- `content/skills/flow-research/SKILL.md` — se edita (T1).
- `content/agents/design/hive-design-ui.md` — se editan los pasos 1, 3 y 4 (T1); su párrafo final no cambia, porque `tests/content/agent_boundary_test.go` lo fija.
- `content/skills/unattended-delegation/SKILL.md` — se edita (T7).
- `content/skills/flow-build/references/delegation.md` — se edita: el trabajo que hace un hijo (T1) y «Reviewer brief» (T5).
- `content/skills/hive-ui-criteria/references/design-record.md` — se edita (T2); `content/skills/hive-ui-criteria/SKILL.md` — confirmar que su lista de campos no necesita el nuevo.
- `content/skills/flow-plan/SKILL.md`, `references/plan-handoff.md`, `assets/change-folder/design.md` — se editan (T3).
- `content/skills/flow-build/SKILL.md`, `content/skills/flow-verify/SKILL.md` — se editan (T4); `content/skills/flow-verify/references/ui-verification.md` — confirmar.
- `content/agents/review/hive-review-ux.md` — se edita (T5).
- `site/src/content/docs/flows/research.md`, `plan.md`, `build.md`, `verify.md` y `_support/docs/architecture/instruction-resources.md` — se editan (T6), junto con `flows/overview.md`, cuya fila de research describe sus preguntas; `skills.md` — confirmar.
- `content/guidance/global.md` — confirmar sin cambio, 45,095 bytes: «Scope and authorization», la definición de tarea operacional de «Proportionality» y «Questions and routes».
- `content/skills/flow-close/SKILL.md` — confirmar que el cierre de una tarea operacional cubre la imagen local sin cambio.
- `content/skills/flow-walkthrough/SKILL.md` — confirmar sin cambio.
- `content/skills/harness-audit/` — confirmar que lo que dice de `DESIGN.md` no lista sus campos.
- `README.md`, `llms.txt` — confirmar que no enuncian nada que cambie.
- `tests/pilot/` y `tests/content/` — el único texto fijado que el cambio rozaba es el párrafo final de los roles (`agent_boundary_test.go`), que no se toca; confirmar que ningún otro detector fija el texto cambiado.
- Specs: `ui-criteria` «Design record» (modificado) y `workflow-lifecycle` (requisito nuevo).

No toca producción. Todos los proyectos ven un cambio: research ofrece propuestas para una pantalla nueva o reorganizada con la composición abierta.

## Recovery

Revertir el commit restaura el comportamiento anterior. Los frames que una sesión haya creado en un archivo de diseño no se revierten con el commit: quedan en una página propia del trabajo, que el usuario conserva o borra.
